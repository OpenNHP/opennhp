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

### eBPF map pin mismatch at startup

`nhp-acd` loads two objects — `nhp_ebpf_xdp.o` and `tc_egress.o` — that share
pinned maps by name (`conn_track`, `nhp_config`, `knock_peers`, the
whitelists). The loader
validates map type, key size, value size and `max_entries` when the second
object reuses a pin, so a pin left over from a different build stops the AC
from starting with an error naming the map:

```
Failed to load and assign tc eBPF objects
... map "conn_track": ... incompatible with pinned map ...
```

Previously the map most likely to be named here was `spp`, because the TC
program wrote it; now that TC only reads the whitelists and writes
`conn_track`, the same class of failure shows up as a **`conn_track`** (or
`nhp_config`) error instead. It means the same thing and has the same fix —
the pins are stale:

```bash
ssh nhp-ac 'systemctl stop nhp-acd; ls -l /sys/fs/bpf/'
# nhp-acd calls CleanupBPFFiles() on start, which removes every pin it owns,
# so a plain restart normally clears it:
ssh nhp-ac 'sudo systemctl restart nhp-acd'
# if something else still holds them (an orphaned tc filter or xdp link):
ssh nhp-ac 'sudo tc qdisc del dev $(ip route | awk "/default/ {print \$5; exit}") clsact; \
            sudo ip link set dev $(ip route | awk "/default/ {print \$5; exit}") xdpgeneric off; \
            sudo rm -f /sys/fs/bpf/{conn_track,nhp_config,spp,src_port,sdwhitelist,port_list,protocol_port,icmpwhitelist,knock_peers,xdp_white_prog,tc_egress_prog}'
```

### What the knock window covers (and checking that it closes)

The knock's `OpenTime` governs **new flows**. A TCP session that was opened
while the knock was valid keeps running until it goes idle — a large download,
a WebSocket, a long-poll is not cut mid-stream when the window closes. That is
deliberate, and it is what `FilterMode = 0` does as well: the iptables baseline
(`docker/iptables_defaults_*.sh`) accepts `-m state --state ESTABLISHED` ahead
of the per-knock ipset, so expiry there blocks new connections only.

**UDP is not treated that way, and the difference is deliberate** — see the
UDP bullet below.

In eBPF mode the rule is enforced by `conn_track` (`nhp/ebpf/xdp/`):

* An admitted flow gets an entry keyed on the **full 4-tuple, including the
  peer's source port**. A new connection uses a new source port, misses
  `conn_track`, and has to satisfy the whitelist maps — which only user space
  writes, from the knock.
* **A TCP entry tracks the connection, not just the tuple**, and gets a
  refreshed idle window (`NHP_CT_TCP_IDLE_TTL_NS`, 1h) that every packet pushes
  forward. The peer picks its own source port, so a refreshed window alone
  would let a peer that knocked once pin a 4-tuple open forever and keep
  opening connections on it. So: only a **pure SYN** admitted by a whitelist
  creates an entry; a pure SYN never matches an existing entry and always goes
  back through the whitelists (this is what netfilter does — a SYN on a
  closed/TIME_WAIT conntrack becomes NEW, which `--state ESTABLISHED` does not
  match); a **FIN or an inbound RST** cuts the entry down to
  `NHP_CT_CLOSE_TTL_NS` (60s). An inbound RST does *not* delete the entry —
  XDP cannot tell a genuine RST from a spoofed one (no sequence check, and the
  kernel's RFC 5961 check runs after XDP), and deleting would let any off-path
  host that guesses the 4-tuple cut the session.
* **The close cut is not permanent, for the same reason.** An inbound FIN is as
  unverifiable as an inbound RST, so neither may shorten a session's window for
  the rest of its life — otherwise one spoofed packet would drop an SSH,
  WebSocket or long-poll session from 1h (or 180s for an AC-initiated flow) to
  60s and kill it on the first idle gap after that. The next packet of a flow
  that is still running restores its full window
  (`nhp_ct_close_ttl_refresh()`). The cut becomes permanent only once FINs have
  been seen in **both** directions, and the AC's own half is recorded by the TC
  egress program (`CT_FLAG_FIN_LOCAL`), which a peer cannot fake — so a real
  close still dies in 60s rather than lingering for an hour.
* **A UDP (or other non-TCP) entry is capped at the knock's remaining
  lifetime** — `min(NHP_CT_OTHER_IDLE_TTL_NS (120s), time left on the knock)` —
  and is **never refreshed** by an inbound packet. There is no handshake to
  bind such an entry to a single exchange, so a refreshed window would let a
  peer that knocked once keep a 4-tuple admitted indefinitely by sending one
  datagram every 120s, "any"-protocol resources (every port) included. Instead
  the entry ages out and the flow falls back through the whitelist lookups: it
  keeps going while the knock is live and dies with it. A long UDP session is
  therefore cut when `OpenTime` runs out, unlike a TCP one and unlike
  `FilterMode = 0`, where nf_conntrack keeps it up.
* The TC egress program writes `conn_track` only for connections the **AC
  itself** opens (dnf, certbot, NTP, its UDP channel to the nhp-server), and
  four gates keep it off knock flows: a reverse lookup of the ingress
  direction, TCP pure-SYN only, a UDP source port that must be inside the
  host's `net.ipv4.ip_local_port_range` (read at load time into the
  `nhp_config` map, and logged at startup) or one of the well-known client
  ports listed in `is_wellknown_client_port()`, and a UDP check that the peer
  holds — or has ever held — a knock for the port we are sending from. It never
  writes the whitelist maps.
* **`knock_peers` is what makes that last gate hold after expiry.** The gate
  cannot rely on the whitelist maps alone: XDP *deletes* a whitelist entry the
  moment a packet finds it expired, so by the time the AC answers a peer whose
  knock ran out there is nothing left in `spp` / `src_port` / `sdwhitelist` to
  find. The XDP program therefore records every non-TCP packet a whitelist
  admits in `knock_peers`, keyed on **peer address + the port on the AC +
  protocol**, and that record is never deleted (it is an LRU, and `nhp-acd`
  drops the pin at startup). It grants nothing — no packet is ever admitted
  because of it. It also covers the two knock shapes the whitelist lookups
  cannot be asked about (`port_list`, whose key wildcards the port range, and
  `protocol_port`, whose key has no peer address at all).
* **A record is never written for a port the AC owns as a client**, and that
  matters because the record never lapses: one written for, say, the source port
  of the AC's channel to the nhp-server would shut the egress gate on that
  channel for the life of the daemon. The TC program marks each entry it creates
  for a flow the AC opened with **`CT_FLAG_AC_CLIENT`** (`0x20`) when nothing at
  that moment said the peer was knock-gated on the port being sent from — for
  TCP the pure-SYN gate is proof by itself, for UDP it takes all five whitelist
  shapes and `knock_peers` coming up empty. XDP serves such a flow from the
  reverse lookup even while the peer's address holds a live knock, never
  deletes it on a `knock_peers` record, and records nothing when a whitelist
  admits a packet on that 4-tuple. So the two claims on a (peer, port,
  protocol) triple — the knock's and the AC's — are mutually exclusive, and
  whichever side got there first keeps it.
* An **egress** entry (`NHP_CT_EGRESS_IDLE_TTL_NS`, 180s) is refreshed by the
  AC's **own outgoing packets** only. For UDP, XDP deliberately does not
  refresh it from an inbound packet, so a peer cannot slide the window forward
  by itself — tighter than netfilter, which refreshes on either direction. An
  AC-initiated UDP flow therefore lives 180s past the moment the AC last spoke
  on it, which is what keeps the long-lived AC↔nhp-server channel up (see the
  next case). XDP also refuses to serve the reverse (egress) lookup at all to a
  peer that holds a **live `sdwhitelist` entry** or — for non-TCP — a
  **`knock_peers` record for the port**, so an egress entry that the gates
  missed can never become a second way in. In the `knock_peers` case the entry
  is **deleted** rather than just skipped: that record never lapses, so the
  entry could never serve anything again, and the AC's own sending would
  otherwise keep resurrecting it (the TC refresh path does not test expiry) as a
  permanent passenger in the `conn_track` dump. Both of those step aside for an
  entry carrying `CT_FLAG_AC_CLIENT` — that one *does* pre-date the knock by
  construction, so it is the AC's own flow and is served and kept. An expired
  claimed entry admits nothing but is still kept, because it is the only record
  that the port is the AC's; the AC's next outgoing packet revives it.

So "the door is shut" means *a fresh connection is refused*. Test it that way.

#### TCP

Knock, then wait out `OpenTime` (+ the deploy's compensation) and open a
**fresh** connection. `curl` reuses a live connection, so force a new one:

```bash
# during the window — succeeds
curl -sS --max-time 5 https://<resource> -o /dev/null && echo open

sleep $((OpenTime + 10))

# after the window — must fail to connect (not 4xx/5xx: no TCP handshake)
curl -sS --max-time 5 --no-keepalive https://<resource> -o /dev/null \
  || echo "closed (expected)"
```

An existing connection staying alive across the boundary is **not** a
regression — verify with a long transfer started inside the window, which
should complete. To see the difference, keep a connection open (`nc <ac> 443`)
and start a second one after expiry: the first keeps flowing, the second hangs.

#### UDP

A knock-protected UDP service needs its own case, because UDP has no SYN for
the TC egress program to gate on. Unlike TCP, **a UDP flow is cut when the
knock expires** even if it never went idle — its `conn_track` entry is capped
at the time left on the knock — so the test is simply that traffic stops:

```bash
# knock for a UDP resource, then during the window:
nc -u <ac> <udp-port>     # datagrams get through

sleep $((OpenTime + 10))

# after the window, with datagrams still flowing in both directions, the
# peer's datagrams must be dropped. Watch the AC's deny log while sending.
ssh nhp-ac 'tail -f /home/ec2-user/nhp-ac/logs/nhp_deny-*.log'
```

The failure mode this also guards against, and the reason the egress gates
matter: the knock expires, XDP drops the peer's next datagram, the service then
answers anyway (a retransmit, a keepalive, a queued reply) — and if that answer
created an egress entry, the peer's datagrams would be admitted through the
reverse lookup with no live knock at all.

##### The AC-answers-first case (run this against a service inside the ephemeral range)

This is the ordering that has to be exercised explicitly, because **the AC
sends first** and the peer's own packets are what would otherwise be dropped.
Use a UDP service on a port **inside** the ephemeral range — WireGuard 51820,
QUIC, a game server — behind an `any`-protocol resource (`Protocol` empty or
`"any"`, `Port = 0`, which is the default resource shape and the one that hits
every gate at once):

```bash
# 1. knock, and get datagrams flowing both ways — a WireGuard handshake, or
#    just `nc -u <ac> 51820` against a listener on the AC
sleep $((OpenTime + 10))

# 2. the peer goes quiet, but the AC's service keeps sending: a WireGuard
#    PersistentKeepalive (25s), a QUIC PTO, a reply to an already-queued
#    datagram. Do NOT send anything from the peer during this step.
ssh nhp-ac 'sudo tcpdump -ni any -c 5 "udp and port 51820 and src host <ac>"'

# 3. now the peer sends again. It must be dropped.
nc -u <ac> 51820           # no reply; the peer's address lands in the deny log
ssh nhp-ac 'tail -f /home/ec2-user/nhp-ac/logs/nhp_deny-*.log'
```

Step 2 is the regression: if the AC's outgoing keepalive is allowed to create
an egress `conn_track` entry, step 3 succeeds and each further reply refreshes
that entry, so the peer keeps the 4-tuple admitted indefinitely with no live
knock. Confirm no egress entry was created for that flow — the peer's address
must not appear on the `flags: 01` (CT_DIR_EGRESS) side:

```bash
ssh nhp-ac 'sudo bpftool map dump pinned /sys/fs/bpf/conn_track | head -40'
sudo bpftool map dump pinned /sys/fs/bpf/spp   # knock entries only, and expired ones are gone

# and that the peer is on record as having knocked that port — this is the map
# the gate reads once the whitelist entry above has been deleted
sudo bpftool map dump pinned /sys/fs/bpf/knock_peers
```

If `spp` keeps gaining entries while nobody is knocking, something is writing
the whitelist from the data path again — that was the original bug. If
`knock_peers` is *empty* after a UDP knock has passed traffic, the gate has
nothing to read after expiry and this whole case is open again.

> `knock_peers` is per (peer, port, protocol) and is never expired, so if the
> AC also needs to talk to that peer **as a UDP client** from that same port
> number, its replies get no permit. A flow the AC had already opened when the
> knock arrived is exempt — it carries `CT_FLAG_AC_CLIENT` and no record is
> written for its port, see the case below. The collision that remains is the
> other order: the peer knocks first, and the AC then opens a client socket on
> exactly that port number. The port is not claimable while the knock is on
> file, so the socket gets no reply permit until the knock lapses — and if the
> peer's datagrams reach that port in the meantime, the record is written and
> the permit is withheld for good. In practice the AC's client sockets take a
> random ephemeral port, so this needs the peer to have knocked exactly that
> port; restarting `nhp-acd` clears the map.

#### The AC's own channel to the nhp-server (run this one every time)

The case that the two tests above miss, and the one that breaks knocks for
everybody rather than for one peer. `nhp-acd` reaches the server over a single
`net.DialUDP` socket held for the life of the process
(`endpoints/ac/udpac.go`), kept busy by `NHP_KPL` every 20s. Nothing in user
space whitelists the server, so the server's datagrams — `NHP_AOP` among them —
get back in **only** through the egress `conn_track` entry the TC program
writes. If that entry is ever allowed to hard-expire under the AC, every
inbound server packet is dropped until the next keepalive re-creates it, and
knocks fail intermittently for as long as the AC is up.

It only shows up after the AC has been running longer than
`NHP_CT_EGRESS_IDLE_TTL_NS` (180s), so **knock again once the AC has been up
more than 5 minutes** — a fresh deploy plus an immediate knock will not catch
it:

```bash
# how long has nhp-acd been up? needs to be > 5 min for this to prove anything
ssh nhp-ac 'systemctl show nhp-acd -p ActiveEnterTimestamp'

# knock, and expect the resource to open on the first attempt
# (a knock that needs a retry, or an AOP timeout in the server log, is the
#  symptom — the AC drops the AOP and the server gives up)
```

Then confirm the channel's egress entry is alive and being refreshed, rather
than being re-created from scratch each time. `nexthdr: 17`, `flags: 1`
(CT_DIR_EGRESS), the server's address, and a `tx_packets` that keeps climbing
across two dumps a minute apart:

```bash
ssh nhp-ac 'sudo bpftool map dump pinned /sys/fs/bpf/conn_track' \
  | grep -A 14 '"nexthdr": 17'
# the AC logs to a file, not journald — see "Where nhp-acd logs" above
ssh nhp-ac 'grep -E "AOP|AOL" /home/ec2-user/nhp-ac/logs/ac-$(date -u +%F).log | tail'
```

A `tx_packets` that resets to 1 every few minutes means the entry is being
re-created, i.e. it expired — that is the regression. While you are there,
check the range the egress gate is using; it is logged once at load:

```bash
ssh nhp-ac 'grep ip_local_port_range /home/ec2-user/nhp-ac/logs/ac-$(date -u +%F).log'
# -> "eBPF egress tracking treats source ports >= 32768 as the AC's own"
ssh nhp-ac 'cat /proc/sys/net/ipv4/ip_local_port_range'   # must agree
```

If the two disagree, the AC's own UDP flows from below the logged bound get no
reply permit at all. The value is read from the host at load time and written
to the `nhp_config` map, so a mismatch means the sysctl was changed after
`nhp-acd` started — restart it.

##### A knock from the nhp-server's own address (run this after changing any egress gate)

The variant of the case above that a plain knock never reaches, and the one
that used to take the AC off the air **permanently**. It needs an agent whose
source address, as the AC sees it, is the **nhp-server's** — which happens
whenever the two sit behind one NAT, when a small office runs both, or simply
when someone tests `nhp-agentd` on the server host — plus an `any`-protocol
resource (`Protocol` empty or `"any"`, `Port = 0`, the default resource shape),
because that writes `sdwhitelist`, keyed on the address pair alone and so
matching every port including the AC's client socket.

On the demo stack the cheapest way to get that address is to knock from the
server host itself:

```bash
# 1. baseline: the channel's egress entry exists and is claimed.
#    flags: 32 (0x20 CT_FLAG_AC_CLIENT) on a "flags": 1 (CT_DIR_EGRESS) key
ssh nhp-ac 'sudo bpftool map dump pinned /sys/fs/bpf/conn_track' \
  | grep -B 8 -A 14 '"nexthdr": 17'

# 2. knock from the nhp-server host for an any-protocol resource, and note the
#    AC's OpenTime for it (deploy/config-templates/ac/resource.toml). Any agent
#    running there will do — nhp-agentd knocks for whatever is in its
#    etc/resource.toml — the point is only that the AC sees the server's address
#    as the knocking peer.
ssh nhp-server 'cd /home/ec2-user/nhp-agent && ./nhp-agentd run'

# 3. the knock must NOT have given the AC's own channel an ingress entry, and
#    must NOT have put the channel's source port on record. The server's
#    address may appear in knock_peers only for the resource's port(s) — never
#    for the ephemeral port the AC dials from (step 1's "sport" on the
#    "flags": 1 key).
ssh nhp-ac 'sudo bpftool map dump pinned /sys/fs/bpf/knock_peers'
ssh nhp-ac 'sudo bpftool map dump pinned /sys/fs/bpf/conn_track' \
  | grep -B 8 -A 14 '"nexthdr": 17'

# 4. wait the knock out, then knock again — from anywhere — and expect it to
#    open on the first attempt
sleep $((OpenTime + 10))
```

Step 4 is the regression. If the channel was cut, the knock fails with an AOP
timeout on the server and nothing at all on the AC (the AOP never arrives), and
it stays that way for every later knock until `nhp-acd` is restarted:

```bash
ssh nhp-server 'grep -E "AOP|timeout" /home/ec2-user/nhp-server/logs/server-$(date -u +%F).log | tail'
ssh nhp-ac     'grep -E "AOP|AOL"     /home/ec2-user/nhp-ac/logs/ac-$(date -u +%F).log | tail'

# and the fingerprints of the failure, in this order:
#  - knock_peers holds a record whose dst_port is the AC's dial port
#  - the channel's egress entry is gone from conn_track, and the AC's
#    keepalives do not bring it back (tx_packets never reappears)
ssh nhp-ac 'sudo bpftool map dump pinned /sys/fs/bpf/knock_peers'
ssh nhp-ac 'sudo bpftool map dump pinned /sys/fs/bpf/conn_track | grep -c "\"flags\": 1,"'
```

The same test applies to any other UDP host the AC talks to **as a client** —
an NTP server, a resolver — if an agent ever knocks from its address: knock
from it, wait out `OpenTime`, then check the AC still gets replies
(`chronyc sources`, `dig @<resolver>`). A flow the AC had already opened at the
time of the knock is claimed and survives; one it opens *while* the knock is
live is not, and comes back only once the knock lapses — see the note under the
`knock_peers` dump above.

#### Known limitation: fixed low UDP source ports

Gate 3 recognises the AC's own UDP client sockets by their source port being
inside the host's ephemeral range, plus the fixed well-known client ports
listed in `is_wellknown_client_port()` (`nhp/ebpf/xdp/tc_egress.c`): **123**
(NTP, which classic `ntpd` sends from — chrony's default random acquisition
port is covered by the range anyway), **500** and **4500** (IKE / IKE NAT-T).

Any *other* UDP socket the AC binds to a fixed port below the ephemeral range
gets **no reply permit**, and its replies are dropped by XDP unless a knock
whitelists the peer. In practice that means a daemon with an explicit low
`bind()`, or anything put in `net.ipv4.ip_local_reserved_ports`. One-way
senders (syslog to 514, SNMP traps) are unaffected because they expect no
reply.

Symptom: the AC's outbound UDP request is on the wire (`tcpdump -ni any udp`
shows it leaving) but the reply never reaches the process, and the peer's
address shows up in the deny log. Fixes, in order of preference:

1. let the daemon use an ephemeral source port (chrony rather than `ntpd`,
   drop the explicit `bind()`);
2. add the port to `is_wellknown_client_port()` and redeploy the eBPF objects
   (`make ebpf-objs`), if it is a port the AC does not also serve to knocking
   peers;
3. fall back to `FilterMode = 0`, where nf_conntrack tracks it regardless.

```bash
# which local UDP ports does the AC actually hold open, and from where?
ssh nhp-ac 'sudo ss -unap'
ssh nhp-ac 'cat /proc/sys/net/ipv4/ip_local_reserved_ports'
```

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

## nhp-server ingress filter (XDP)

`nhp-serverd` attaches `nhp_server_xdp.o` at startup and drops everything at
the driver except UDP on the knock port and SSH from the addresses in
`etc/xdp.toml`. Unlike the AC's filter there is no fail-closed backstop and no
break-glass SSH path, so the whitelist is the one config on these hosts that
locks you out when it is wrong — including out of the SSH session doing the
change, since XDP has no connection tracking.

The list carries the relay's **private and public** addresses. CI (and any
operator following this runbook) reaches the server with
`ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV`, so the packets stay inside
the VPC and the server sees the relay's *private* address as the source; the
public address covers the hairpin path through the internet gateway. The
`deploy-server` job refuses to upload a whitelist that omits the address the
host reports for its own SSH connection, which is the check that turns this
mistake into a red run instead of a lockout.

Hands-on verification — a netns rehearsal that exercises every branch of the
decision tree, the baseline to capture before the first deploy (including the
check that the whitelisted address is the one the host actually sees SSH
arrive from), the same commands to re-run afterwards, and the recovery paths
if it does go wrong — is in **[VERIFY-server-xdp.zh-cn.md](VERIFY-server-xdp.zh-cn.md)** (Chinese).
