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

  # NHP knock (UDP/443) from anywhere.
  #
  # udp/443 is the NHP protocol knock port -- it MUST stay open. The value is
  # driven by var.nhp_listen_port (the single source of truth for the knock
  # port across Terraform, the deploy-demo-v2 templates, and the runtime
  # configuration). Native nhp-agent clients reach this host directly on this
  # UDP port; the browser demo reaches it through the relay (HTTPS -> UDP).
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
  # Do not confuse the two 443s:
  #   - udp/443 (above) -- NHP knock, MUST stay open
  #   - tcp/443         -- retired demo login page, MUST stay closed
  # The browser knock demo reaches this host through the relay (HTTPS -> UDP),
  # not through any TCP port on this security group.

  # SSH only from relay (jump host)
  ingress {
    description     = "SSH from relay"
    from_port       = 22
    to_port         = 22
    protocol        = "tcp"
    security_groups = [aws_security_group.relay.id]
  }

  # Everything else, from anywhere -- deliberately, EXCEPT tcp/22. The
  # nhp-server's XDP filter (nhp/ebpf/xdp/nhp_server_xdp.c) is the ONLY inbound
  # control on the rest of this host in this mode; the point is to let scans
  # reach the NIC so the filter can count them ([NHP-STAT] / [NHP-SCAN]).
  #
  # The filter is fail-open and its link is not pinned, so it is detached on
  # every deploy, crash, restart, reboot and failed load. sshd is the
  # highest-value surface on the host, so its reachability must never depend on
  # the filter: tcp/22 stays restricted to the relay SG at this layer and is
  # deliberately carved out of the open range below. Consequence: scans of
  # tcp/22 no longer reach the NIC and are not counted.
  #
  # An SG rule cannot say "all except 22", so the range is spelled out per
  # protocol: tcp 0-21 and 23-65535, all udp, all icmp. Other IP protocols
  # (GRE, ESP, ...) are no longer reachable, which is narrower than before.
  # Whatever else is listening (chronyd, SSM/cloud-init agents) is still
  # exposed while the filter is detached; see terraform/demo/RUNBOOK.md
  # "All-ports exposure mode". The two rules above are a subset of this one
  # (the knock port) or independent of it (ssh), so flipping the variable to
  # false returns exactly to the narrow posture. Order: deploy-demo-v2 (confirm
  # "Verify the XDP ingress filter attached" is green) BEFORE infra-demo apply
  # -- the opposite of the knock-port rule.
  dynamic "ingress" {
    for_each = var.server_open_all_ports ? [1] : []
    content {
      description = "ALL tcp except 22 (0-21)"
      from_port   = 0
      to_port     = 21
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  dynamic "ingress" {
    for_each = var.server_open_all_ports ? [1] : []
    content {
      description = "ALL tcp except 22 (23-65535)"
      from_port   = 23
      to_port     = 65535
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  dynamic "ingress" {
    for_each = var.server_open_all_ports ? [1] : []
    content {
      description = "ALL udp (XDP is the only control; see CLAUDE.md)"
      from_port   = 0
      to_port     = 65535
      protocol    = "udp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  dynamic "ingress" {
    for_each = var.server_open_all_ports ? [1] : []
    content {
      description = "ALL icmp (XDP is the only control; see CLAUDE.md)"
      from_port   = -1
      to_port     = -1
      protocol    = "icmp"
      cidr_blocks = ["0.0.0.0/0"]
    }
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

  # Historical: nhp-acd has no fixed UDP listen port. It dials server with
  # net.DialUDP from an ephemeral source port (endpoints/ac/msghandler.go:525)
  # and opens a short-lived UDP socket per authorization (the actual listen
  # on AC's side). This rule is kept for documentation and is effectively a
  # no-op on top of AWS SG's stateful return-traffic allowance. Do not
  # remove it as part of the knock-port change -- deleting SG rules is an
  # independent cleanup task and must not be bundled here.
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

  # Everything else, from anywhere -- deliberately, EXCEPT tcp/22. The AC's
  # XDP filter (nhp/ebpf/xdp/nhp_ebpf_xdp.c, attached by nhp-acd at startup in
  # FilterMode = 1) is the ONLY inbound control on the rest of this host in
  # this mode; the point is to let scans reach the NIC so the filter can count
  # them ([NHP-STAT] per verdict class, [NHP-SCAN] per port and per source).
  #
  # Why the AC needs this at all: a packet the security group drops never
  # reaches the NIC, so the filter never sees it and it is never counted. In
  # the narrow posture only tcp/443 and the knock port were reachable, so the
  # AC's statistics covered a fraction of what was actually hitting the host.
  # The server's own all-ports mode exists for the same reason.
  #
  # The filter is fail-open and its link is not pinned, so it is detached on
  # every deploy, crash, restart, reboot and failed load. sshd is the
  # highest-value surface on the host, so its reachability must never depend on
  # the filter: tcp/22 stays restricted to the relay SG at this layer and is
  # deliberately carved out of the open range below. Consequence: scans of
  # tcp/22 do not reach the NIC and are not counted.
  #
  # An SG rule cannot say "all except 22", so the range is spelled out per
  # protocol: tcp 0-21 and 23-65535, all udp, all icmp. Other IP protocols
  # (GRE, ESP, ...) are not reachable, which is narrower than before.
  # Whatever else is listening (chronyd, SSM/cloud-init agents) is still
  # exposed while the filter is detached; see terraform/demo/RUNBOOK.md
  # "All-ports exposure mode".
  #
  # Flipping the variable to false returns exactly to the narrow posture: the
  # tcp/443 rule above is a subset of the open range, the ssh rule is
  # independent of it, and the knock-port rule above widens from "from the
  # server SG" to "from anywhere" in this mode -- the AC has no fixed UDP
  # listener, so that rule was always a no-op (see the comment on it).
  # Order: deploy-demo-v2 (confirm the AC's XDP filter is attached) BEFORE
  # infra-demo apply -- the reverse of the knock-port rule.
  dynamic "ingress" {
    for_each = var.ac_open_all_ports ? [1] : []
    content {
      description = "ALL tcp except 22 (0-21)"
      from_port   = 0
      to_port     = 21
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  dynamic "ingress" {
    for_each = var.ac_open_all_ports ? [1] : []
    content {
      description = "ALL tcp except 22 (23-65535)"
      from_port   = 23
      to_port     = 65535
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  dynamic "ingress" {
    for_each = var.ac_open_all_ports ? [1] : []
    content {
      description = "ALL udp (XDP is the only control; see CLAUDE.md)"
      from_port   = 0
      to_port     = 65535
      protocol    = "udp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  dynamic "ingress" {
    for_each = var.ac_open_all_ports ? [1] : []
    content {
      description = "ALL icmp (XDP is the only control; see CLAUDE.md)"
      from_port   = -1
      to_port     = -1
      protocol    = "icmp"
      cidr_blocks = ["0.0.0.0/0"]
    }
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
