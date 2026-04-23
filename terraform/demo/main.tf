terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }

  backend "s3" {
    bucket       = "opennhp-tfstate-401696231478"
    key          = "demo/terraform.tfstate"
    region       = "us-east-2"
    encrypt      = true
    kms_key_id   = "alias/aws/s3"
    use_lockfile = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = var.tags
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

# Read secrets from AWS Secrets Manager
data "aws_secretsmanager_secret_value" "demo" {
  secret_id = "opennhp/demo"
}

locals {
  secrets = jsondecode(data.aws_secretsmanager_secret_value.demo.secret_string)
}
