output "server_public_ip" {
  description = "NHP Server public IP (demologin.opennhp.org)"
  value       = aws_eip.server.public_ip
}

output "server_private_ip" {
  description = "NHP Server private IP (for inter-service communication)"
  value       = aws_instance.server.private_ip
}

output "ac_public_ip" {
  description = "NHP AC public IP (acdemo.opennhp.org)"
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
    demologin = "demologin.${var.domain} -> ${aws_eip.server.public_ip}"
    acdemo    = "acdemo.${var.domain} -> ${aws_eip.ac.public_ip}"
    relay     = "relay.${var.domain} -> ${aws_eip.relay.public_ip}"
    agent     = "agent.${var.domain} -> ${aws_eip.relay.public_ip}"
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
