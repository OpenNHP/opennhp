# =============================================================================
# TLS certificates for demo.nhp (signed by custom CA)
# =============================================================================
# The CA root certificate and private key are stored in AWS Secrets Manager
# (opennhp/demo) under keys: stealth_ca_cert, stealth_ca_key
#
# This generates a server certificate for demo.nhp signed by that CA.
# =============================================================================

# Generate a new private key for demo.nhp server certificate
resource "tls_private_key" "demo_nhp" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

# Create a certificate signing request
resource "tls_cert_request" "demo_nhp" {
  private_key_pem = tls_private_key.demo_nhp.private_key_pem

  subject {
    common_name  = "demo.nhp"
    organization = "OpenNHP"
  }

  dns_names = ["demo.nhp"]
}

# Sign the certificate with our CA
resource "tls_locally_signed_cert" "demo_nhp" {
  cert_request_pem   = tls_cert_request.demo_nhp.cert_request_pem
  ca_private_key_pem = local.secrets["stealth_ca_key"]
  ca_cert_pem        = local.secrets["stealth_ca_cert"]

  validity_period_hours = 87600 # 10 years

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}
