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

  # Backend config is partial: bucket + region are supplied at init time via
  # -backend-config (see workflows). Keeping them out of source avoids leaking
  # the AWS account ID embedded in the bucket name.
  backend "s3" {
    key          = "demo/terraform.tfstate"
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
data "aws_secretsmanager_secret" "demo" {
  name = "opennhp/demo"
}

data "aws_secretsmanager_secret_version" "demo" {
  secret_id = data.aws_secretsmanager_secret.demo.id
}

locals {
  secrets = jsondecode(data.aws_secretsmanager_secret_version.demo.secret_string)
}
