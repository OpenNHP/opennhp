# Auto-generate SSH keypair on first apply, register with EC2, and persist the
# private key to AWS Secrets Manager for CI consumption.
#
# The keypair is managed by Terraform and survives apply cycles. The private
# key itself lives only in Terraform state (encrypted S3 backend) and in AWS SM.

resource "tls_private_key" "deploy" {
  algorithm = "ED25519"
}

resource "aws_key_pair" "deploy" {
  key_name   = var.key_pair_name
  public_key = tls_private_key.deploy.public_key_openssh

  tags = { Name = "opennhp-demo-keypair" }
}

# Merge generated private key back into opennhp/demo secret without clobbering
# other fields (cloudflare tokens, nhp keys, host keys, etc.).
resource "aws_secretsmanager_secret_version" "ssh_key_writeback" {
  secret_id = "opennhp/demo"
  secret_string = jsonencode(merge(
    local.secrets,
    { ssh_deploy_private_key = tls_private_key.deploy.private_key_openssh }
  ))

  # Depend on the keypair so we only write back once the key is actually usable.
  depends_on = [aws_key_pair.deploy]

  lifecycle {
    # Ignore drift - other CI jobs (key generation, host keys) update this
    # secret too, so don't let Terraform revert those edits.
    ignore_changes = [secret_string]
  }
}
