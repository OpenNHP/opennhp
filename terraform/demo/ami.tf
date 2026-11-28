# Keep fresh hosts on the image version admitted by the first-download guard.
# Candidate image: owned ami-09424005afd2baed7 (us-east-2).
# Require 6.12.115-140.210.amzn2023.x86_64 and its matching kernel6.12 RPM.
# Encrypted private copy of AWS ami-0bdaca0026bed499d; confirm actual boot
# before admission or a fresh-host release. Copy metadata is not boot proof.
# Change the ID, image name and admitted pair together in one reviewed commit.

data "aws_ami" "amazon_linux_2023" {
  most_recent        = false
  include_deprecated = true
  owners             = ["401696231478"]

  # Retain an exact owned image instead of depending on public-image custody.
  filter {
    name   = "image-id"
    values = ["ami-09424005afd2baed7"]
  }

  filter {
    name   = "name"
    values = ["opennhp-cve-2026-98432-fixed-al2023-6.12-20261128"]
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
