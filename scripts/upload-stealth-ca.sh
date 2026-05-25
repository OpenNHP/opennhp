#!/bin/bash
# Upload stealth CA certificate and key to AWS Secrets Manager
#
# Usage:
#   ./upload-stealth-ca.sh --ca-cert /path/to/ca.pem --ca-key /path/to/ca-key.pem
#   ./upload-stealth-ca.sh [--region REGION] [--ca-cert FILE] [--ca-key FILE]
#
# Options:
#   --region      AWS region (default: us-east-2 or $AWS_REGION)
#   --ca-cert     Path to CA certificate PEM file (required)
#   --ca-key      Path to CA private key PEM file (required)
#
# Prerequisites:
#   - AWS CLI configured with appropriate credentials
#   - jq installed locally
#   - Access to opennhp/demo secret in Secrets Manager
#
# This script is intended for the initial bootstrap or emergency/manual sync.
# The normal path is the infra-demo workflow, which keeps GitHub Secrets and
# AWS Secrets Manager aligned during Terraform apply.

set -euo pipefail

usage() {
    echo "Usage: $0 --ca-cert FILE --ca-key FILE [--region REGION]"
    echo ""
    echo "Options:"
    echo "  --ca-cert FILE   Path to CA certificate PEM file (required)"
    echo "  --ca-key FILE    Path to CA private key PEM file (required)"
    echo "  --region REGION  AWS region (default: us-east-2)"
    echo ""
    echo "Example:"
    echo "  $0 --ca-cert ./rootCA.pem --ca-key ./rootCA-key.pem"
    exit 1
}

# Default values
REGION="${AWS_REGION:-us-east-2}"
SECRET_ID="opennhp/demo"
CA_CERT_FILE=""
CA_KEY_FILE=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --region)
            REGION="$2"
            shift 2
            ;;
        --ca-cert)
            CA_CERT_FILE="$2"
            shift 2
            ;;
        --ca-key)
            CA_KEY_FILE="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate required arguments
if [[ -z "$CA_CERT_FILE" ]] || [[ -z "$CA_KEY_FILE" ]]; then
    echo "ERROR: --ca-cert and --ca-key are required"
    echo ""
    usage
fi

echo "=== Upload Stealth CA to AWS Secrets Manager ==="
echo "Region: $REGION"
echo "Secret: $SECRET_ID"
echo "CA Cert: $CA_CERT_FILE"
echo "CA Key:  $CA_KEY_FILE"
echo

# Check files exist
if [[ ! -f "$CA_CERT_FILE" ]]; then
    echo "ERROR: CA certificate not found: $CA_CERT_FILE"
    exit 1
fi
if [[ ! -f "$CA_KEY_FILE" ]]; then
    echo "ERROR: CA key not found: $CA_KEY_FILE"
    exit 1
fi

# Fetch existing secret
echo "Fetching existing secret from AWS Secrets Manager..."
EXISTING_SECRET=$(aws secretsmanager get-secret-value \
    --secret-id "$SECRET_ID" \
    --region "$REGION" \
    --query SecretString \
    --output text)

# Merge in the PEM files byte-for-byte, preserving trailing newlines.
echo "Merging stealth CA into secret..."
UPDATED_SECRET=$(jq \
    --rawfile cert "$CA_CERT_FILE" \
    --rawfile key "$CA_KEY_FILE" \
    '. + {stealth_ca_cert: $cert, stealth_ca_key: $key}' \
    <<< "$EXISTING_SECRET")

# Update secret in AWS
echo "Updating AWS Secrets Manager..."
SECRET_FILE=$(mktemp)
trap 'rm -f "$SECRET_FILE"' EXIT
printf '%s' "$UPDATED_SECRET" > "$SECRET_FILE"

aws secretsmanager put-secret-value \
    --secret-id "$SECRET_ID" \
    --region "$REGION" \
    --secret-string "file://$SECRET_FILE"

echo
echo "=== Success ==="
echo "Stealth CA certificate and key have been stored in $SECRET_ID"
echo "Fields added: stealth_ca_cert, stealth_ca_key"
echo
echo "Next steps:"
echo "  1. Use this script only for bootstrap or emergency/manual sync."
echo "  2. Run 'terraform init && terraform apply' in terraform/demo/"
echo "  3. Run deploy-demo-v2 workflow on GitHub"
