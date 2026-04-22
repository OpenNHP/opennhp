variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-2"
}

variable "aws_account_id" {
  description = "AWS account ID (do not commit)"
  type        = string
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
  description = "EC2 SSH key pair name (auto-created by Terraform)"
  type        = string
  default     = "opennhp-demo"
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
