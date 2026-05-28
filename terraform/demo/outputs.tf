output "server_public_ip" {
  description = "NHP Server public IP (auth-plugin.opennhp.org; legacy: demologin.opennhp.org)"
  value       = aws_eip.server.public_ip
}

output "server_private_ip" {
  description = "NHP Server private IP (for inter-service communication)"
  value       = aws_instance.server.private_ip
}

output "ac_public_ip" {
  description = "NHP AC public IP (ac.opennhp.org; legacy: acdemo.opennhp.org)"
  value       = aws_eip.ac.public_ip
}

output "ac_private_ip" {
  description = "NHP AC private IP (for inter-service communication)"
  value       = aws_instance.ac.private_ip
}

output "relay_public_ip" {
  description = "NHP Relay public IP (relay.opennhp.org + agent.opennhp.org)"
  value       = aws_eip.relay.public_ip
}

output "relay_private_ip" {
  description = "NHP Relay private IP"
  value       = aws_instance.relay.private_ip
}

output "dns_records" {
  description = "DNS records created"
  value = {
    auth_plugin = "auth-plugin.${var.domain} -> ${aws_eip.server.public_ip}"
    ac          = "ac.${var.domain} -> ${aws_eip.ac.public_ip}"
    demologin   = "demologin.${var.domain} -> CNAME auth-plugin.${var.domain} (legacy)"
    acdemo      = "acdemo.${var.domain} -> CNAME ac.${var.domain} (legacy)"
    relay       = "relay.${var.domain} -> ${aws_eip.relay.public_ip}"
    agent       = "agent.${var.domain} -> ${aws_eip.relay.public_ip}"
  }
}

output "ssh_jump_command" {
  description = "SSH to server/ac via relay jump host"
  value = {
    relay  = "ssh ec2-user@${aws_eip.relay.public_ip}"
    server = "ssh -J ec2-user@${aws_eip.relay.public_ip} ec2-user@${aws_instance.server.private_ip}"
    ac     = "ssh -J ec2-user@${aws_eip.relay.public_ip} ec2-user@${aws_instance.ac.private_ip}"
  }
}

# demo.nhp certificate (signed by stealth CA)
# These outputs are empty strings if stealth CA is not configured.
output "demo_nhp_cert" {
  description = "demo.nhp server certificate chain (leaf + CA, PEM). Empty if stealth CA not configured."
  # Concatenate leaf cert with CA cert so nginx serves the full chain.
  # Browsers with the stealth CA in their root store will validate the leaf,
  # and clients doing path-building from the server-presented chain will have
  # the intermediate/root available.
  value     = local.stealth_ca_enabled ? "${tls_locally_signed_cert.demo_nhp[0].cert_pem}${local.secrets["stealth_ca_cert"]}" : ""
  sensitive = true
}

output "demo_nhp_key" {
  description = "demo.nhp server private key (PEM). Empty if stealth CA not configured."
  value       = try(tls_private_key.demo_nhp[0].private_key_pem, "")
  sensitive   = true
}

output "stealth_ca_enabled" {
  description = "Whether stealth CA is configured and demo.nhp cert is available"
  # nonsensitive() is safe because this is just a boolean indicating whether
  # the CA secrets exist - it doesn't expose any actual secret values.
  value = nonsensitive(local.stealth_ca_enabled)
}
