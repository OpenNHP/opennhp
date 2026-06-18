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

# Server cluster 2: a second, independent nhp-server + nhp-ac sharing the
# same relay. The cluster 2 server runs pure NHP (no http/https, no domain);
# its AC is reachable at ac2.opennhp.org.
resource "aws_eip" "server2" {
  domain = "vpc"
  tags   = { Name = "opennhp-demo-server2-eip" }
}

resource "aws_eip" "ac2" {
  domain = "vpc"
  tags   = { Name = "opennhp-demo-ac2-eip" }
}

# --- EC2 Instances ---

# nhp-server (auth-plugin.opennhp.org; legacy: demologin.opennhp.org)
resource "aws_instance" "server" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.server.id]
  key_name               = aws_key_pair.deploy.key_name

  user_data = templatefile("${path.module}/userdata/server.sh", {
    deploy_path = "/home/ec2-user/nhp-server"
  })

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
  }

  tags = { Name = "opennhp-demo-server" }
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

  user_data = templatefile("${path.module}/userdata/ac.sh", {
    deploy_path = "/home/ec2-user/nhp-ac"
  })

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
  }

  tags = { Name = "opennhp-demo-ac" }
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

# --- Server cluster 2 instances ---

# nhp-server (cluster 2): pure NHP, no http/https, no domain. Shares the
# server userdata (deploy_path only) but is reached solely over UDP 62206
# by its AC and the relay.
resource "aws_instance" "server2" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.server2.id]
  key_name               = aws_key_pair.deploy.key_name

  user_data = templatefile("${path.module}/userdata/server.sh", {
    deploy_path = "/home/ec2-user/nhp-server"
  })

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
  }

  tags = { Name = "opennhp-demo-server2" }
}

resource "aws_eip_association" "server2" {
  instance_id   = aws_instance.server2.id
  allocation_id = aws_eip.server2.id
}

# nhp-ac (cluster 2; ac2.opennhp.org). Same deploy content as cluster 1 ac.
resource "aws_instance" "ac2" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.ac2.id]
  key_name               = aws_key_pair.deploy.key_name

  user_data = templatefile("${path.module}/userdata/ac.sh", {
    deploy_path = "/home/ec2-user/nhp-ac"
  })

  root_block_device {
    volume_size = 20
    volume_type = "gp3"
  }

  tags = { Name = "opennhp-demo-ac2" }
}

resource "aws_eip_association" "ac2" {
  instance_id   = aws_instance.ac2.id
  allocation_id = aws_eip.ac2.id
}
