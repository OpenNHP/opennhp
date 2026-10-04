output "server_public_ip" {
  description = "NHP Server public IP (auth-plugin.opennhp.org; NHP UDP only, no public HTTP/HTTPS)"
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

# Unless var.relay_private_ip pins it, the relay's private address is whatever
# the subnet hands aws_instance.relay, so replacing the instance (an AMI or
# instance-type change, a taint, an AZ move) gives it a new one. That address is
# the source of every SSH session into the nhp-server host, and the server's XDP
# whitelist is what admits it — so a whitelist naming only the old address drops
# SSH from the new relay, and the one path CI has for pushing a corrected
# whitelist is the path that just closed. Recovery would mean detaching the
# server's root volume.
#
# So an unpinned relay makes the whitelist carry this prefix as well as the two
# host addresses (see "Resolve the relay addresses for the nhp-server XDP
# whitelist" in .github/workflows/deploy-demo-v2.yml). A replacement relay comes
# back inside it and is still allowed in. The cost is that the other demo hosts
# share this subnet and are covered too, which is a far smaller problem than a
# server that cannot be reached at all — and which pinning the address removes.
output "subnet_cidr" {
  description = "Public subnet CIDR (an unpinned relay makes the nhp-server XDP whitelist allow SSH from it, so a replaced relay keeps its way in)"
  value       = aws_subnet.public.cidr_block
}

# Read by the deploy pipeline to decide whether the subnet prefix above has to
# be in the whitelist at all. A pinned relay comes back on the same address, so
# the two host entries cover every case and the whitelist can say exactly what
# the policy says: SSH from the relay.
output "relay_private_ip_pinned" {
  description = "Whether var.relay_private_ip pins the relay's private address (if so, the nhp-server XDP whitelist omits the subnet prefix)"
  value       = var.relay_private_ip != ""
}

output "dns_records" {
  description = "DNS records created"
  value = {
    auth_plugin = "auth-plugin.${var.domain} -> ${aws_eip.server.public_ip}"
    server      = "server.${var.domain} -> CNAME auth-plugin.${var.domain}"
    ac          = "ac.${var.domain} -> ${aws_eip.ac.public_ip}"
    acdemo      = "acdemo.${var.domain} -> CNAME ac.${var.domain} (legacy)"
    relay       = "relay.${var.domain} -> ${aws_eip.relay.public_ip}"
    agent       = "agent.${var.domain} -> ${aws_eip.relay.public_ip}"
  }
}

output "ssh_jump_command" {
  description = "SSH to server/ac via relay jump host"
  value = {
    relay   = "ssh ec2-user@${aws_eip.relay.public_ip}"
    server  = "ssh -J ec2-user@${aws_eip.relay.public_ip} ec2-user@${aws_instance.server.private_ip}"
    ac      = "ssh -J ec2-user@${aws_eip.relay.public_ip} ec2-user@${aws_instance.ac.private_ip}"
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
