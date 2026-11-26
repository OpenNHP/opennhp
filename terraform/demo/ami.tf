# Keep fresh hosts on the image version admitted by the first-download guard.
# Expected qualification pair: ami-0bdaca0026bed499d (us-east-2) boots
# 6.12.113-138.209.amzn2023.x86_64, owned by the matching kernel6.12 RPM.
# Confirm the actual boot before a fresh-host release; metadata is not proof.
# Change the ID, image name and admitted pair together in one reviewed commit.

data "aws_ami" "amazon_linux_2023" {
  most_recent        = false
  include_deprecated = true
  owners             = ["amazon"]

  # Keep security rollback plans usable after the public image is deprecated.
  filter {
    name   = "image-id"
    values = ["ami-0bdaca0026bed499d"]
  }

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
