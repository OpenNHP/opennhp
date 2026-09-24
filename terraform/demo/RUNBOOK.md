# Demo Terraform Runbook

## Stealth CA Security Model

The demo infrastructure includes a "stealth CA" for signing certificates for
internal hostnames (e.g., `demo.nhp`) that are not publicly resolvable. This
section documents the security implications.

### Trust Model

**Important:** The stealth CA private key is stored in Terraform state.

When `tls_locally_signed_cert.demo_nhp` signs a certificate using
`ca_private_key_pem = local.secrets["stealth_ca_key"]`, Terraform persists this
value (and all resource inputs) to the state file. Although the state bucket is
KMS-encrypted, anyone with `s3:GetObject` on the bucket or access to
`terraform show -json` can extract the CA root private key.

**Risk:** This CA can sign certificates for **any hostname**, not just
`demo.nhp`. If an attacker obtains the CA key and a victim trusts the stealth
CA, the attacker can perform MitM attacks against any HTTPS connection from
that victim.

### Mitigations

1. **Limit who trusts the stealth CA.** Only install the CA certificate on
   machines specifically intended for demo testing. Never install it on
   production systems or personal devices used for sensitive work.

2. **Restrict state bucket access.** Ensure only the minimum required IAM
   principals have `s3:GetObject` on the state bucket. Review access regularly.

3. **Use short-lived certificates.** The `demo.nhp` certificate validity is
   set to 2 years and Terraform is configured to rotate it 60 days before
   expiry during `terraform apply`. The 60-day window gives the monthly
   renewal workflow two scheduled chances to renew before expiry. This limits
   blast radius if state ever leaks, but the renewed certificate still must
   be redeployed to the EC2 host.

4. **Rotate the CA if state is compromised.** If you suspect the state bucket
   was accessed by unauthorized parties, generate a new CA keypair and update
   all clients that trust it.

### Alternative: Out-of-band signing

For higher security, consider signing certificates outside of Terraform:

```bash
# One-shot signing that never writes the CA key to state
openssl x509 -req -in demo_nhp.csr \
  -CA /path/to/ca.crt -CAkey /path/to/ca.key \
  -CAcreateserial -out demo_nhp.crt -days 730
```

Then import the signed certificate into Secrets Manager manually rather than
using `tls_locally_signed_cert`. This keeps the CA key entirely out of
Terraform state at the cost of manual certificate management.

### Renewal and deployment path

The `demo.nhp` certificate is not renewed by certbot on the EC2 host. Renewal
only happens when Terraform reevaluates `tls_locally_signed_cert.demo_nhp`
during `terraform apply`, after which the PEM files must be copied to
`/etc/nginx/certs/` on the AC instance.

The repository includes a scheduled GitHub Actions workflow,
`Renew demo.nhp Certificate`, that runs monthly to:

1. Execute a targeted Terraform plan/apply for the `demo.nhp` certificate resources
2. Read the refreshed `demo_nhp_cert` / `demo_nhp_key` outputs when the stealth CA is enabled
3. Deploy them to the AC instance and reload nginx, or remove stale `demo.nhp` nginx/cert files if the stealth CA has been disabled

If that workflow is disabled or fails repeatedly, re-run it manually before the
certificate reaches expiry.

### Stealth CA upload paths

The canonical path for syncing `stealth_ca_cert` and `stealth_ca_key` into AWS
Secrets Manager is the `infra-demo` workflow during `action=apply`.

Use `scripts/upload-stealth-ca.sh` only for the initial bootstrap or an
emergency/manual re-sync when the GitHub workflow path is unavailable.

### opennhp/demo secret schema (stealth CA fields)

| Field | Populated by | Used by |
|-------|--------------|---------|
| `stealth_ca_cert` | `infra-demo` workflow (from GitHub Secrets) | `tls_locally_signed_cert.demo_nhp` |
| `stealth_ca_key` | `infra-demo` workflow (from GitHub Secrets) | `tls_locally_signed_cert.demo_nhp` |

---

## One-time migration: SSH deploy key out of Terraform state

Until this migration runs, the SSH deploy private key is stored in plaintext
inside Terraform state (`tls_private_key.deploy.private_key_openssh`). Anyone
with read access to the state bucket can recover it.

After this migration:
- Terraform only holds the **public** key (passed in via
  `TF_VAR_deploy_public_key`).
- The **private** key lives only in AWS Secrets Manager
  (`opennhp/demo` → `ssh_deploy_private_key`).
- Future `terraform apply` runs never read or write the private key.

The migration **does not rotate** the keypair — it only relocates ownership.
The same public key stays registered with EC2; SSH access is uninterrupted.

If you suspect the existing private key is compromised (it has been in state
for a while; assume any historical state-bucket reader has a copy), rotate
**after** this migration completes — see "Rotation" below.

### Prerequisites

- AWS CLI configured against the demo account, with
  `secretsmanager:GetSecretValue` and `secretsmanager:PutSecretValue` on
  `opennhp/demo`.
- `terraform` ≥ 1.10, `jq`, `ssh-keygen`, `gh` (GitHub CLI) on PATH.
- `cd terraform/demo`
- `terraform init -backend-config="bucket=$TF_STATE_BUCKET" -backend-config="region=us-east-2"`

### Step 1 — Derive the public key from the existing private key

```bash
PRIV=$(mktemp) && chmod 600 "$PRIV"
trap 'shred -u "$PRIV" 2>/dev/null || rm -f "$PRIV"' EXIT

aws secretsmanager get-secret-value \
  --secret-id opennhp/demo --region us-east-2 \
  --query SecretString --output text \
  | jq -r '.ssh_deploy_private_key' > "$PRIV"

PUB=$(ssh-keygen -y -f "$PRIV")
echo "$PUB"
```

Sanity-check: this string should match the `public_key` field on the existing
`opennhp-demo` keypair in EC2:

```bash
aws ec2 describe-key-pairs --key-names opennhp-demo --include-public-key \
  --region us-east-2 --query 'KeyPairs[0].PublicKey' --output text
```

If they don't match, **stop** — Secrets Manager and EC2 are out of sync, and
the migration would lock you out. Investigate before continuing.

### Step 2 — Drop the old resources from state

These are state-only operations; they don't touch AWS.

```bash
terraform state rm tls_private_key.deploy
terraform state rm aws_secretsmanager_secret_version.ssh_key_writeback
terraform state rm aws_key_pair.deploy
```

### Step 3 — Re-import the existing keypair under the new resource definition

```bash
export TF_VAR_deploy_public_key="$PUB"
# Plus the variables Terraform expects for any apply:
export TF_VAR_cloudflare_api_token="$(aws secretsmanager get-secret-value \
  --secret-id opennhp/demo --region us-east-2 --query SecretString --output text \
  | jq -r '.cloudflare_api_token')"
export TF_VAR_cloudflare_zone_id="$(aws secretsmanager get-secret-value \
  --secret-id opennhp/demo --region us-east-2 --query SecretString --output text \
  | jq -r '.cloudflare_zone_id')"

terraform import aws_key_pair.deploy opennhp-demo
```

### Step 4 — Verify drift is zero

```bash
terraform plan
```

Expected: **No changes**. If Terraform wants to recreate `aws_key_pair.deploy`
or any `aws_instance`, **stop** and reconcile — replacing the keypair forces
EC2 instance replacement, which takes the demo offline. Recreating an instance
also blanks any state your demo has accumulated (logs, generated keys on disk,
etc).

### Step 5 — Commit and push

The repo's `infra-demo` workflow now reads the private key from Secrets
Manager, derives the public key with `ssh-keygen -y`, and passes it via
`TF_VAR_deploy_public_key`. After Step 4 shows zero drift, future CI applies
should be no-ops on the keypair.

---

## Rotation (separate, optional, manual)

Run this only when you're sure you want to invalidate the existing private key
(e.g., after the migration above, to clear the historical exposure window).
**Do not automate this** — a mistake locks you out of every demo host.

### Step R1 — Generate a fresh keypair locally

```bash
NEW_PRIV=$(mktemp) && chmod 600 "$NEW_PRIV"
trap 'shred -u "$NEW_PRIV" "${NEW_PRIV}.pub" 2>/dev/null || rm -f "$NEW_PRIV" "${NEW_PRIV}.pub"' EXIT

ssh-keygen -t ed25519 -f "$NEW_PRIV" -N "" -C "opennhp-demo-deploy-$(date -u +%Y%m%d)"
NEW_PUB=$(cat "${NEW_PRIV}.pub")
```

### Step R2 — Append the new public key to authorized_keys on every host

Use the **current** key (still valid) to add the new one. Do **not** remove
the old one yet.

```bash
OLD_PRIV=$(mktemp) && chmod 600 "$OLD_PRIV"
aws secretsmanager get-secret-value \
  --secret-id opennhp/demo --region us-east-2 \
  --query SecretString --output text \
  | jq -r '.ssh_deploy_private_key' > "$OLD_PRIV"

# Repeat for SERVER, AC, RELAY public IPs (use Terraform outputs)
for HOST in $SERVER_IP $AC_IP $RELAY_IP; do
  ssh -i "$OLD_PRIV" ec2-user@"$HOST" \
    "echo '$NEW_PUB' >> ~/.ssh/authorized_keys"
done
```

### Step R3 — Verify the new key works

```bash
for HOST in $SERVER_IP $AC_IP $RELAY_IP; do
  ssh -i "$NEW_PRIV" -o StrictHostKeyChecking=yes ec2-user@"$HOST" \
    'echo OK from $(hostname)'
done
```

If any host fails, **stop**. Re-run Step R2 for that host. Do not proceed
until every host responds with `OK`.

### Step R4 — Swap the keypair in EC2 / Terraform

```bash
# Drop and re-import with the new public key.
terraform state rm aws_key_pair.deploy
aws ec2 delete-key-pair --key-name opennhp-demo --region us-east-2
aws ec2 import-key-pair --key-name opennhp-demo \
  --public-key-material "fileb://${NEW_PRIV}.pub" --region us-east-2

export TF_VAR_deploy_public_key="$NEW_PUB"
terraform import aws_key_pair.deploy opennhp-demo
terraform plan   # expect: no changes
```

### Step R5 — Rotate the secret in Secrets Manager

```bash
SECRETS=$(aws secretsmanager get-secret-value \
  --secret-id opennhp/demo --region us-east-2 \
  --query SecretString --output text)
NEW_PRIV_PEM=$(cat "$NEW_PRIV")
UPDATED=$(echo "$SECRETS" | jq --arg k "$NEW_PRIV_PEM" '. + {ssh_deploy_private_key: $k}')
aws secretsmanager put-secret-value \
  --secret-id opennhp/demo --region us-east-2 \
  --secret-string "$UPDATED"
```

### Step R6 — Remove the old key from authorized_keys on every host

Use the **new** key now.

```bash
OLD_PUB=$(ssh-keygen -y -f "$OLD_PRIV")
for HOST in $SERVER_IP $AC_IP $RELAY_IP; do
  ssh -i "$NEW_PRIV" ec2-user@"$HOST" \
    "grep -vF '$OLD_PUB' ~/.ssh/authorized_keys > ~/.ssh/authorized_keys.new \
     && mv ~/.ssh/authorized_keys.new ~/.ssh/authorized_keys"
done
```

Verify CI still works by running the `deploy-demo-v2` workflow.

## nhp-ac in eBPF/XDP mode

`deploy/config-templates/ac/config.toml` ships `FilterMode = 1`, so the demo AC
enforces its per-knock whitelist with XDP ingress + TC egress instead of
iptables+ipset. Two operational consequences.

### Kernel prerequisite: >= 6.6

`endpoints/ac/ebpf/ebpfegine.go` attaches the egress program with
`link.AttachTCX`, i.e. the kernel TCX / `bpf_mprog` API added in Linux 6.6, and
there is no fallback. AL2023 AMIs still boot 6.1 by default.

Both paths onto a new enough kernel are automated; **no manual step is needed
for the cutover**:

* **Fresh instances** — `terraform/demo/userdata/ac.sh` installs `kernel6.12`,
  points the default boot entry at it with `grubby --set-default` and reboots
  at the end of first boot.
* **The long-lived AC** — userdata runs once per instance, and
  `aws_instance.ac` is pinned twice over so an edit to `userdata/ac.sh` cannot
  disturb the running host: it leaves `user_data_replace_on_change` at its
  default `false` (replacing that host would drop its EIP association and
  deployed state) **and** carries `lifecycle { ignore_changes = [user_data] }`.
  Both are load-bearing — without the `ignore_changes`, the AWS provider
  applies a `user_data` change in place, which stops and starts the instance
  for no benefit (cloud-init still will not re-run the script). With them,
  `terraform apply` neither re-runs userdata nor bounces the AC. The
  `deploy-ac` job therefore does the upgrade itself: it reads `uname -r` before
  touching the host, and if the kernel is older than 6.6 it runs
  `dnf install -y kernel6.12` and pins it with `grubby --set-default` — both
  *before* anything else on the host is touched, so a failed install aborts the
  deploy with the host exactly as it was. The
  reboot follows later, once the fail-closed backstop is installed and
  `tcp/443` is closed; the job then waits up to 10 minutes for the host to come
  back **on a >= 6.6 kernel**, and `nhp-ac-backstop-boot.service` (below)
  re-closes the port before the network comes up on the other side.

Installing 6.12 is not enough on its own — it has to stay the **default boot
entry**, or the host keeps running fine until the next reboot (maintenance, an
EC2 retirement event, a crash) and then comes up on 6.1 with the TCX attach
failing and the backstop holding `tcp/443` shut, with no deploy running to
explain it. Two things keep that from happening:

* Every eBPF-mode deploy checks `grubby --default-kernel`, not just
  `uname -r`, and re-pins 6.12 if something moved it. A default that cannot be
  repaired fails the deploy.
* Once the host is running 6.12, the job removes the 6.1 `kernel` package
  (best-effort), because `/etc/sysconfig/kernel` ships `UPDATEDEFAULT=yes`: a
  routine `dnf upgrade` that pulls a newer 6.1 build would otherwise hand it
  the default again. `dnf` refuses to remove the running kernel, so this can
  only ever take the unused one.

If `kernel6.12` cannot be installed or pinned, or the host does not come back
on a new enough kernel, the job aborts with `tcp/443` closed and nothing cut
over. To recover by hand:

```bash
ssh nhp-ac   # via the relay jump host
sudo dnf install -y kernel6.12
sudo grubby --set-default "/boot/vmlinuz-$(rpm -q --qf '%{version}-%{release}.%{arch}\n' kernel6.12 | sort -V | tail -1)"
sudo grubby --default-kernel   # confirm it points at 6.12 before rebooting
sudo reboot
uname -r     # confirm the running kernel afterwards
```

Then re-run `deploy-demo-v2`. The alternative is to roll back: set
`FilterMode = 0` in `deploy/config-templates/ac/config.toml` and re-run.

### Fail-closed backstop

The XDP and TCX links are held in the daemon's memory only (package variables
in `ebpfegine.go`); only the *programs* are pinned under `/sys/fs/bpf`, which
keeps the objects alive but attaches nothing. Enforcement therefore ends the
moment nhp-acd exits, and `tcp/443` is open to the world in the AC security
group. `deploy/scripts/nhp-ac-backstop.sh` covers that gap with a small
netfilter chain, wired into the unit by the deploy job as
`/etc/systemd/system/nhp-acd.service.d/10-ebpf-backstop.conf`:

| hook | action |
| --- | --- |
| `ExecStartPre` | `up` — close tcp/443 before the daemon starts |
| `ExecStartPre` | `bpffs-prep` — `chgrp ec2-user` + `chmod 0770 /sys/fs/bpf` so the unprivileged daemon can pin (see below) |
| `ExecStartPost` | `wait-attach` — lift it only once the XDP program is attached; fail the unit otherwise |
| `ExecStopPost` | `up` — close it again when the daemon exits (stop, crash, restart backoff) |

So a stopped or crashed nhp-acd means the protected port is *closed*, not open.

The drop-in also sets `StartLimitIntervalSec=600` / `StartLimitBurst=10`. A
`wait-attach` failure fails the unit, and each retry (30s timeout + 5s
`RestartSec`) outlasts systemd's default 10s start-limit window, so without
this the daemon would retry a hopeless attach forever. The burst is
deliberately generous: a unit that hits its start limit stays `failed` — with
443 closed — until someone runs `systemctl reset-failed` or another deploy, so
a tight limit would turn a handful of transient crashes on an internet-facing
daemon into a standing outage. Ten attempts still burns the hot loop's budget
in about six minutes. If you do find the unit stuck there:

```bash
ssh nhp-ac
sudo systemctl reset-failed nhp-acd && sudo systemctl start nhp-acd
```

Netfilter rules do not survive a reboot, though, and the drop-in only runs when
the unit does — between the network coming up and nhp-acd's `ExecStartPre`,
nothing would hold the port. nginx (what actually listens on 443) starts in
parallel with nhp-acd and can win that race, and the kernel upgrade above
reboots the host from inside the deploy. The same job therefore also installs
`/etc/systemd/system/nhp-ac-backstop-boot.service`, a `oneshot` that runs
`nhp-ac-backstop.sh up` with `DefaultDependencies=no` and
`Before=network-pre.target`, i.e. before any interface is configured. It is
enabled in FilterMode 1 and **disabled and removed in FilterMode 0**: iptables
mode never calls `down`, so its unconditional DROP would sit in front of the
per-knock ACCEPTs and keep 443 dark after the next reboot.

```bash
systemctl status nhp-ac-backstop-boot.service   # "active (exited)" is the healthy state
```

Inspect or drive the backstop by hand on the host:

```bash
sudo /usr/local/sbin/nhp-ac-backstop.sh status
sudo /usr/local/sbin/nhp-ac-backstop.sh up               # close now
sudo /usr/local/sbin/nhp-ac-backstop.sh down             # lift the IPv4 chain
sudo /usr/local/sbin/nhp-ac-backstop.sh flush-guard-up   # park the raw-table DROP
sudo /usr/local/sbin/nhp-ac-backstop.sh flush-guard-down # remove it
```

The IPv6 chain installed by `up` is deliberately permanent: the XDP program
returns `XDP_PASS` for every IPv6 frame, so it does not filter IPv6 at all.

`up` re-reads the rules afterwards and exits non-zero if they are not actually
in the kernel (missing `iptables` binary, a lost `/run/xtables.lock` race), so
an `ExecStartPre` failure aborts the start instead of letting the daemon come
up behind a backstop that was never installed. Never read "backstop active"
from anything but a zero exit status.

### Capabilities: no CAP_DAC_OVERRIDE

In eBPF mode the unit runs with `CAP_BPF CAP_NET_ADMIN CAP_PERFMON` and nothing
else — no `CAP_SYS_ADMIN` and no `CAP_DAC_OVERRIDE`. Both are root-equivalent
as *ambient* capabilities on an internet-facing daemon and `NoNewPrivileges=`
does not mitigate either: `CAP_DAC_OVERRIDE` bypasses every file permission
check, so a compromised nhp-acd could write `/etc/cron.d/*`,
`~root/.ssh/authorized_keys` or the unit file itself. Its only job was pinning
into `/sys/fs/bpf` (systemd mounts it `0700 root:root`), which the drop-in's
second `ExecStartPre` — `nhp-ac-backstop.sh bpffs-prep` — now handles by
group-owning that one directory (falling back to a `mount -o remount` if the
kernel refuses `chgrp`/`chmod` on bpffs, and verifying the result either way).
If pins stop appearing under `/sys/fs/bpf` after a manual unit edit, check that
hook first:

```bash
stat -c '%A %U:%G' /sys/fs/bpf                            # want drwxrwx--- root:ec2-user
sudo NHP_BPFFS_GROUP=ec2-user /usr/local/sbin/nhp-ac-backstop.sh bpffs-prep
```

### Where nhp-acd logs

**Not journald.** `endpoints/ac/udpac.go` installs a file logger
(`log.NewLogger("NHP-AC", ..., <exe dir>/logs, "ac")`) as the global logger
before anything interesting happens, and `nhp/log/logger.go` only writes to
stdout when both the directory and the name are empty. So `journalctl -u
nhp-acd` shows systemd messages, one init line and Go panics — nothing else.
The real log is on the host:

```bash
ls -lt /home/ec2-user/nhp-ac/logs/          # ac-<date>.log, plus nhp_accept / nhp_deny
tail -f /home/ec2-user/nhp-ac/logs/ac-$(date -u +%F).log
```

The `deploy-ac` job reads that file (not the journal) to confirm the perf ring
buffer opened, and dumps both it and `journalctl` on every abort.

### Rolling back to iptables

Set `FilterMode = 0` in `deploy/config-templates/ac/config.toml` and re-run
`deploy-demo-v2`. The `deploy-ac` job reads the rendered value and derives
everything from it: it re-applies `iptables_default.sh -f` as the baseline,
restores the `CAP_NET_ADMIN CAP_NET_RAW CAP_DAC_OVERRIDE` capability set and
removes the backstop drop-in. No workflow edit and no manual host cleanup.

The rollback keeps tcp/443 closed across the whole window even though it
removes the drop-in: on an eBPF host the IPv4 INPUT policy is `ACCEPT` and XDP
is the only filter, so `systemctl stop nhp-acd` would otherwise leave the port
open to the world until the netfilter baseline returns several steps later.

That takes two rules, because the backstop chain alone does not survive the
baseline script. `iptables_default.sh -f` opens with `iptables -F; iptables -X`
— filter table — which deletes the `NHP_BACKSTOP` chain and its INPUT jump, and
the policy only goes back to `DROP` at the very end of the script, after the
ipset creates and the whole `NHP_DENY`/INPUT/FORWARD setup. Pre-setting
`-P INPUT DROP` is not an option: the same `-F` removes the `--dport 22` ACCEPT
and would cut the deploy's own SSH session. So the job parks a second DROP in
the **raw** table (`flush-guard-up`) — a table the script never touches, and
`raw/PREROUTING` runs before conntrack, so established flows are covered too —
runs the script, checks the baseline really left `INPUT` at policy `DROP`,
deletes the IPv6 chain by hand (the `-f` flush is IPv4 only) and only then
removes the raw guard (`flush-guard-down`). Any failure aborts with the port
still closed.

An aborted rollback can therefore leave a raw/PREROUTING DROP behind on
purpose. It is invisible to `iptables -S` (filter) and survives `flush-legacy`,
so the eBPF cutover path clears it before starting the daemon; by hand:

```bash
sudo /usr/local/sbin/nhp-ac-backstop.sh status            # reports the raw guard too
sudo /usr/local/sbin/nhp-ac-backstop.sh flush-guard-down
```
