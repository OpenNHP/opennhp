variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-2"
}

variable "domain" {
  description = "Base domain name"
  type        = string
  default     = "opennhp.org"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.small"
}

variable "key_pair_name" {
  description = "EC2 SSH key pair name (registered with the public key in var.deploy_public_key)"
  type        = string
  default     = "opennhp-demo"
}

variable "deploy_public_key" {
  description = "OpenSSH-format public key for the deploy keypair. Generated outside Terraform; the matching private key lives only in AWS Secrets Manager (opennhp/demo -> ssh_deploy_private_key)."
  type        = string
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "subnet_cidr" {
  description = "Public subnet CIDR block"
  type        = string
  default     = "10.0.1.0/24"
}

# Pinning the relay's private address narrows the nhp-server XDP whitelist.
#
# That whitelist is the only thing that reaches tcp/22 on the server host. It
# carries the relay's private address (SSH arrives from there, because CI jumps
# through the relay), the relay's public address, and — because
# aws_instance.relay normally takes whatever the subnet hands it — the whole
# public subnet, so that a replaced relay coming back on a different address is
# still let in. Without that prefix a replacement is a root-volume-detach
# recovery: the live whitelist names the old address, SSH from the new relay is
# dropped, and the deploy job that would push the correction reaches the server
# *through* the relay.
#
# The cost is that the AC and the server share that subnet and are covered too,
# which is wider than the policy anyone would write down. Pinning the address
# here removes the reason for the prefix: a replaced relay comes back on the
# same address, so the deploy renders the two host addresses alone and the XDP
# layer matches the security group (SSH from the relay's SG and nothing else).
#
# Left empty by default, which keeps today's behaviour exactly. Set it to the
# address the relay *already has* — `terraform output relay_private_ip` — and
# apply: Terraform sees no change, because the attribute already holds that
# value. Setting it to any other address replaces the instance, so do not
# invent one.
variable "relay_private_ip" {
  description = "Pin the relay's private IPv4 (must be the address it already has, or the instance is replaced). Empty lets AWS choose, and the deploy then whitelists the whole subnet so a replaced relay keeps SSH to nhp-server."
  type        = string
  default     = ""

  validation {
    condition     = var.relay_private_ip == "" || can(cidrnetmask("${var.relay_private_ip}/32"))
    error_message = "relay_private_ip must be empty or a single IPv4 address."
  }
}

variable "nhp_listen_port" {
  description = "NHP protocol UDP port"
  type        = number
  default     = 62206
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID for opennhp.org (loaded from AWS SM)"
  type        = string
  default     = ""
}

variable "cloudflare_api_token" {
  description = "Cloudflare API token (loaded from AWS SM)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default = {
    Project     = "opennhp"
    Environment = "demo"
    ManagedBy   = "terraform"
  }
}
