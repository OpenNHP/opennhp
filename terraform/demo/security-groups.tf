# --- nhp-relay Security Group ---
# Jump host: only instance with public SSH access
resource "aws_security_group" "relay" {
  name_prefix = "opennhp-demo-relay-"
  description = "NHP Relay - jump host with SSH + HTTPS relay"
  vpc_id      = aws_vpc.demo.id

  # SSH from anywhere (jump host)
  ingress {
    description = "SSH"
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

  # HTTPS auth endpoint from anywhere
  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
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
