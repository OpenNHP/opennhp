# --- nhp-relay Security Group ---
# Jump host: only instance with public SSH access
resource "aws_security_group" "relay" {
  name_prefix = "opennhp-demo-relay-"
  description = "NHP Relay - jump host with SSH + HTTPS relay"
  vpc_id      = aws_vpc.demo.id

  # SSH from anywhere (jump host).
  #
  # This demo deploys from GitHub Actions runners, whose egress IPs are
  # a large, changing set documented at
  # https://api.github.com/meta. Restricting to that list is operationally
  # expensive and still very broad. Hardening options for non-demo use:
  #   1. Replace the jump-host pattern with AWS Systems Manager Session
  #      Manager (no open SSH port required).
  #   2. Restrict to a known admin CIDR via a `ssh_admin_cidrs` variable.
  # Neither is done here to keep the demo single-command reproducible.
  ingress {
    description = "SSH (demo jump host)"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # HTTPS relay endpoint
  ingress {
    description = "HTTPS relay"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "opennhp-demo-relay-sg" }

  lifecycle {
    create_before_destroy = true
  }
}

# --- nhp-server Security Group ---
# No public SSH, only SSH from relay SG
resource "aws_security_group" "server" {
  name_prefix = "opennhp-demo-server-"
  # NOTE: keep this description string unchanged. aws_security_group.description
  # is ForceNew in the AWS provider, so editing it would replace the whole SG
  # (churning its id and forcing aws_security_group.ac's inline reference to be
  # rewritten before the old SG can be deleted -- a DependencyViolation-prone
  # replace on a live demo SG). Removing the 443 ingress below is an in-place
  # rule change; the retirement of the HTTP/HTTPS surface is documented in the
  # comment block where that rule used to be.
  description = "NHP Server - UDP knocking + HTTPS auth"
  vpc_id      = aws_vpc.demo.id

  # NHP protocol (UDP) from anywhere
  ingress {
    description = "NHP UDP"
    from_port   = var.nhp_listen_port
    to_port     = var.nhp_listen_port
    protocol    = "udp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # NOTE: there is deliberately NO tcp/443 (or tcp/80) ingress here.
  #
  # The nhp-server used to expose an HTTP "demo login" page
  # (/plugins/example?action=login&resid=demo) via an nginx vhost on 443,
  # reachable as auth-plugin.opennhp.org and the legacy alias
  # demologin.opennhp.org. That surface has been retired: the vhost is torn
  # down by the deploy-server job and nhp-serverd runs with EnableHttp=false
  # (deploy/config-templates/server/http.toml). Closing the SG rule here is
  # the authoritative, internet-facing control -- do not re-add it without
  # also re-enabling those two layers.
  #
  # The UDP rule above is the NHP protocol itself and must stay open. The
  # browser knock demo reaches this host through the relay (HTTPS -> UDP),
  # not through any TCP port on this security group.

  # SSH only from relay (jump host)
  ingress {
    description     = "SSH from relay"
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = [aws_security_group.relay.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "opennhp-demo-server-sg" }

  lifecycle {
    create_before_destroy = true
  }
}

# --- nhp-ac Security Group ---
# No public SSH, only SSH from relay SG
resource "aws_security_group" "ac" {
  name_prefix = "opennhp-demo-ac-"
  description = "NHP AC - access controller with protected resources"
  vpc_id      = aws_vpc.demo.id

  # Protected resource HTTPS from anywhere
  ingress {
    description = "HTTPS protected resource"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # NHP AOP from server (UDP)
  ingress {
    description     = "NHP UDP from server"
    from_port       = var.nhp_listen_port
    to_port         = var.nhp_listen_port
    protocol        = "udp"
    security_groups = [aws_security_group.server.id]
  }

  # SSH only from relay (jump host)
  ingress {
    description     = "SSH from relay"
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = [aws_security_group.relay.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "opennhp-demo-ac-sg" }

  lifecycle {
    create_before_destroy = true
  }
}
