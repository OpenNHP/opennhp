# Keep fresh hosts on the image version admitted by the first-download guard.
# Qualify a new image and kernel/RPM pair together before changing this pin.

data "aws_ami" "amazon_linux_2023" {
  most_recent = false
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-2023.12.20261120.0-kernel-6.12-x86_64"]
  }

  filter {
    name   = "state"
    values = ["available"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}
