#!/bin/bash
# Upload stealth CA certificate and key to AWS Secrets Manager
# Usage: ./upload-stealth-ca.sh [--region REGION]
#
# Prerequisites:
#   - AWS CLI configured with appropriate credentials
#   - Access to opennhp/demo secret in Secrets Manager

set -euo pipefail

# Default values
REGION="${AWS_REGION:-us-east-2}"
SECRET_ID="opennhp/demo"
CA_CERT_FILE="/opt/fengbi/stealth-dns/etc/cert/rootCA.pem"
CA_KEY_FILE="/opt/fengbi/stealth-dns/etc/cert/rootCA-key.pem"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --region)
            REGION="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

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

# Read certificate and key
echo "Reading CA certificate and key..."
CA_CERT=$(cat "$CA_CERT_FILE")
CA_KEY=$(cat "$CA_KEY_FILE")

# Fetch existing secret
echo "Fetching existing secret from AWS Secrets Manager..."
EXISTING_SECRET=$(aws secretsmanager get-secret-value \
    --secret-id "$SECRET_ID" \
    --region "$REGION" \
    --query SecretString \
    --output text)

# Merge new fields into existing secret using Python (handles JSON properly)
echo "Merging stealth CA into secret..."
UPDATED_SECRET=$(python3 << EOF
import json
import sys

existing = json.loads('''$EXISTING_SECRET''', strict=False)

# Add/update stealth CA fields
existing['stealth_ca_cert'] = '''$CA_CERT'''
existing['stealth_ca_key'] = '''$CA_KEY'''

print(json.dumps(existing))
EOF
)

# Update secret in AWS
echo "Updating AWS Secrets Manager..."
aws secretsmanager put-secret-value \
    --secret-id "$SECRET_ID" \
    --region "$REGION" \
    --secret-string "$UPDATED_SECRET"

echo
echo "=== Success ==="
echo "Stealth CA certificate and key have been stored in $SECRET_ID"
echo "Fields added: stealth_ca_cert, stealth_ca_key"
echo
echo "Next steps:"
echo "  1. Run 'terraform init && terraform apply' in terraform/demo/"
echo "  2. Run deploy-demo-v2 workflow on GitHub"
