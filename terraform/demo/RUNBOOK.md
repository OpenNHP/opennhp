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
