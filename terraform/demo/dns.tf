# Cloudflare DNS records for opennhp.org
# proxied = false because NHP uses UDP which Cloudflare proxy doesn't support

resource "cloudflare_record" "auth_plugin" {
  zone_id = var.cloudflare_zone_id
  name    = "auth-plugin"
  content = aws_eip.server.public_ip
  type    = "A"
  proxied = false
  ttl     = 300
  comment = "NHP Server (auth plugin) - managed by Terraform"
}

resource "cloudflare_record" "ac" {
  zone_id = var.cloudflare_zone_id
  name    = "ac"
  content = aws_eip.ac.public_ip
  type    = "A"
  proxied = false
  ttl     = 300
  comment = "NHP AC - managed by Terraform"
}

# Canonical alias for the cluster 1 nhp-server. Points at auth-plugin so the
# two names resolve to the same host; lets the demo refer to the cluster
# as server.opennhp.org.
resource "cloudflare_record" "server" {
  zone_id = var.cloudflare_zone_id
  name    = "server"
  content = "auth-plugin.opennhp.org"
  type    = "CNAME"
  proxied = false
  ttl     = 300
  comment = "Alias for auth-plugin.opennhp.org (cluster 1 nhp-server) - managed by Terraform"
}

# Legacy aliases. Kept as CNAMEs to the new primary names so existing
# agents, bookmarks, and shipped plugin configs that still reference the
# old hostnames continue to work.
#
# The "demologin" CNAME (demologin.opennhp.org -> auth-plugin.opennhp.org)
# was removed deliberately. It aliased the nhp-server's HTTP demo login
# page, which has been retired along with the server's public HTTP/HTTPS
# surface (see security-groups.tf). Removing the record here stops the
# hostname resolving at all. Re-adding it would only produce a dead name
# unless the nginx vhost and the tcp/443 SG rule are restored too.

resource "cloudflare_record" "acdemo" {
  zone_id = var.cloudflare_zone_id
  name    = "acdemo"
  content = "ac.opennhp.org"
  type    = "CNAME"
  proxied = false
  ttl     = 300
  comment = "Legacy alias for ac.opennhp.org - managed by Terraform"
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

resource "cloudflare_record" "reg" {
  zone_id = var.cloudflare_zone_id
  name    = "reg"
  content = aws_eip.relay.public_ip
  type    = "A"
  proxied = false
  ttl     = 300
  comment = "NHP Agent Registration page (hosted on relay) - managed by Terraform"
}

# Integrated demo app (Gin backend + embedded SPA). Shares the relay host;
# nginx proxies demo.opennhp.org → 127.0.0.1:8081. Same cert lineage as
# relay/agent/reg (SAN expanded by bootstrap-tls.sh).
resource "cloudflare_record" "demo" {
  zone_id = var.cloudflare_zone_id
  name    = "demo"
  content = aws_eip.relay.public_ip
  type    = "A"
  proxied = false
  ttl     = 300
  comment = "OpenNHP integrated demo app (hosted on relay) - managed by Terraform"
}
