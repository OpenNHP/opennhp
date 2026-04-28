# SSH keypair management.
#
# The private key is generated outside Terraform (see runbook in
# terraform/demo/RUNBOOK.md) and stored only in AWS Secrets Manager
# under opennhp/demo -> ssh_deploy_private_key. Terraform receives just
# the public key via var.deploy_public_key — nothing sensitive ever
# lands in Terraform state.

resource "aws_key_pair" "deploy" {
  key_name   = var.key_pair_name
  public_key = var.deploy_public_key

  tags = { Name = "opennhp-demo-keypair" }
}

# Tell Terraform that the previous tls_private_key.deploy and
# aws_secretsmanager_secret_version.ssh_key_writeback resources are gone
# from configuration but should NOT be destroyed in AWS. Without this,
# any environment whose state still contains those resources would, on
# the next apply, destroy the live SecretVersion under opennhp/demo —
# rolling AWSCURRENT back to a stale payload and breaking deploys.
removed {
  from = tls_private_key.deploy
  lifecycle {
    destroy = false
  }
}

removed {
  from = aws_secretsmanager_secret_version.ssh_key_writeback
  lifecycle {
    destroy = false
  }
}
