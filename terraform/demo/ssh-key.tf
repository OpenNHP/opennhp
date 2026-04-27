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
