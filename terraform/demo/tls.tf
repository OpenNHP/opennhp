# =============================================================================
# TLS certificates for demo.nhp (signed by custom CA)
# =============================================================================
# The CA root certificate and private key are stored in AWS Secrets Manager
# (opennhp/demo) under keys: stealth_ca_cert, stealth_ca_key
#
# This generates a server certificate for demo.nhp signed by that CA.
#
# PREREQUISITE: STEALTH_CA_CERT and STEALTH_CA_KEY must be configured in GitHub
# Secrets and synced to AWS Secrets Manager via the infra-demo workflow before
# these resources can be created. If the secrets are not present, these
# resources are skipped (count = 0) to allow terraform plan/apply to succeed.
# =============================================================================

locals {
  # Check if stealth CA is configured. If either cert or key is missing/empty,
  # skip creating the demo.nhp certificate resources.
  stealth_ca_enabled = (
    lookup(local.secrets, "stealth_ca_cert", "") != "" &&
    lookup(local.secrets, "stealth_ca_key", "") != ""
  )
}

# Generate a new private key for demo.nhp server certificate
resource "tls_private_key" "demo_nhp" {
  # Intentionally tie the server keypair lifecycle to stealth_ca_enabled:
  # removing and later restoring the CA secrets regenerates demo.nhp.
  count = local.stealth_ca_enabled ? 1 : 0

  algorithm = "RSA"
  rsa_bits  = 2048
}

# Create a certificate signing request
resource "tls_cert_request" "demo_nhp" {
  count = local.stealth_ca_enabled ? 1 : 0

  private_key_pem = tls_private_key.demo_nhp[0].private_key_pem

  subject {
    common_name  = "demo.nhp"
    organization = "OpenNHP"
  }

  dns_names = ["demo.nhp"]
}

# Sign the certificate with our CA
resource "tls_locally_signed_cert" "demo_nhp" {
  count = local.stealth_ca_enabled ? 1 : 0

  cert_request_pem   = tls_cert_request.demo_nhp[0].cert_request_pem
  ca_private_key_pem = local.secrets["stealth_ca_key"]
  ca_cert_pem        = local.secrets["stealth_ca_cert"]

  # 2 years validity with automatic renewal 30 days before expiry.
  # Shorter validity limits blast radius if state leaks and ensures
  # regular key rotation. Terraform will recreate the cert when
  # early_renewal_hours threshold is reached.
  validity_period_hours = 17520 # 2 years (365 * 2 * 24)
  early_renewal_hours   = 720   # 30 days (30 * 24)

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}
