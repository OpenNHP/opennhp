# Cloudflare DNS records for opennhp.org
# proxied = false because NHP uses UDP which Cloudflare proxy doesn't support

resource "cloudflare_record" "demologin" {
  zone_id = var.cloudflare_zone_id
  name    = "demologin"
  content = aws_eip.server.public_ip
  type    = "A"
  proxied = false
  ttl     = 300
  comment = "NHP Server demo - managed by Terraform"
}

resource "cloudflare_record" "acdemo" {
  zone_id = var.cloudflare_zone_id
  name    = "acdemo"
  content = aws_eip.ac.public_ip
  type    = "A"
  proxied = false
  ttl     = 300
  comment = "NHP AC demo - managed by Terraform"
}

resource "cloudflare_record" "relay" {
  zone_id = var.cloudflare_zone_id
  name    = "relay"
  content = aws_eip.relay.public_ip
  type    = "A"
  proxied = false
  ttl     = 300
  comment = "NHP Relay - managed by Terraform"
}

resource "cloudflare_record" "agent" {
  zone_id = var.cloudflare_zone_id
  name    = "agent"
  content = aws_eip.relay.public_ip
  type    = "A"
  proxied = false
  ttl     = 300
  comment = "NHP Agent app (hosted on relay) - managed by Terraform"
}
