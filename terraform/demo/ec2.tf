# --- Elastic IPs ---
resource "aws_eip" "server" {
  domain = "vpc"
  tags   = { Name = "opennhp-demo-server-eip" }
}

resource "aws_eip" "ac" {
  domain = "vpc"
  tags   = { Name = "opennhp-demo-ac-eip" }
}

resource "aws_eip" "relay" {
  domain = "vpc"
  tags   = { Name = "opennhp-demo-relay-eip" }
}

# --- EC2 Instances ---

# nhp-server (auth-plugin.opennhp.org). NHP UDP only; the HTTP demo login
# page and its demologin.opennhp.org alias have been retired.
resource "aws_instance" "server" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.server.id]
  key_name               = aws_key_pair.deploy.key_name

  # Same reasoning as the AC below, and for the same reason it is needed here
  # now: the XDP ingress filter's unit settings (AmbientCapabilities=, the
  # bpffs ExecStartPre) were added to userdata/server.sh, and without
  # ignore_changes the AWS provider would apply that edit in place on the next
  # apply - which stops and starts this instance. The reboot would buy nothing
  # (cloud-init runs userdata once per instance, so the new settings would not
  # be applied by it anyway; the deploy-server drop-in is what installs them on
  # a running host) and would cost an unplanned nhp-server outage, taking the
  # whole demo down with it.
  user_data = templatefile("${path.module}/userdata/server.sh", {
    deploy_path     = "/home/ec2-user/nhp-server"
    nhp_listen_port = var.nhp_listen_port
  })

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
  }

  tags = { Name = "opennhp-demo-server" }

  lifecycle {
    # See the user_data comment above. To roll a userdata change onto this
    # host deliberately, replace the instance (taint / -replace) rather than
    # removing this - and be ready to re-associate the EIP, re-deploy, and
    # re-check that the XDP whitelist admits the new host's SSH source before
    # the filter attaches.
    ignore_changes = [user_data]
  }
}

resource "aws_eip_association" "server" {
  instance_id   = aws_instance.server.id
  allocation_id = aws_eip.server.id
}

# nhp-ac (ac.opennhp.org; legacy: acdemo.opennhp.org)
resource "aws_instance" "ac" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.ac.id]
  key_name               = aws_key_pair.deploy.key_name

  # This is a long-lived pet holding an EIP association, a Let's Encrypt
  # account and the deployed AC state, so userdata edits must not disturb it.
  # Two separate settings are needed for that, and only one of them is a
  # default:
  #   - user_data_replace_on_change stays at its default (false), so an edit
  #     never destroys and recreates the instance.
  #   - ignore_changes = [user_data] below is what keeps a *running* host
  #     untouched. Without it the AWS provider still applies the change in
  #     place, and an in-place user_data update stops and starts the instance
  #     (the userdata itself does not re-run - cloud-init executes it once per
  #     instance - so the reboot buys nothing and costs an AC outage).
  # The consequence of both is that edits to userdata/ac.sh only reach *new*
  # instances - anything a running host needs (the >= 6.6 kernel for
  # eBPF/TCX, the unit's capability set) is applied by the deploy-ac job in
  # .github/workflows/deploy-demo-v2.yml instead.
  user_data = templatefile("${path.module}/userdata/ac.sh", {
    deploy_path = "/home/ec2-user/nhp-ac"
  })

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
  }

  tags = { Name = "opennhp-demo-ac" }

  lifecycle {
    # See the user_data comment above. To roll a userdata change onto this
    # host deliberately, replace the instance (taint / -replace) rather than
    # removing this - and be ready to re-associate the EIP and re-deploy.
    ignore_changes = [user_data]
  }
}

resource "aws_eip_association" "ac" {
  instance_id   = aws_instance.ac.id
  allocation_id = aws_eip.ac.id
}

# nhp-relay (relay.opennhp.org + agent.opennhp.org)
resource "aws_instance" "relay" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.relay.id]
  key_name               = aws_key_pair.deploy.key_name

  # Unpinned by default (null = let AWS choose), which is what makes the
  # nhp-server XDP whitelist carry the whole public subnet: a replacement of
  # this instance would otherwise come back on an address the server does not
  # admit, and the deploy that would fix it reaches the server through this
  # host. Set var.relay_private_ip to the address this instance already has and
  # the deploy narrows the whitelist to the two relay addresses instead. See
  # the variable's comment in variables.tf — a *different* address here forces
  # a replacement.
  private_ip = var.relay_private_ip != "" ? var.relay_private_ip : null

  user_data = templatefile("${path.module}/userdata/relay.sh", {
    deploy_path = "/home/ec2-user/nhp-relay"
  })

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
  }

  tags = { Name = "opennhp-demo-relay" }
}

resource "aws_eip_association" "relay" {
  instance_id   = aws_instance.relay.id
  allocation_id = aws_eip.relay.id
}

