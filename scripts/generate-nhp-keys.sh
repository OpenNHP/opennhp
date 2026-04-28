#!/bin/bash
# generate-nhp-keys.sh
#
# Generates NHP key pairs for all components and renders config templates.
# Reads existing keys from AWS Secrets Manager if available, or generates new ones.
#
# Usage:
#   ./scripts/generate-nhp-keys.sh \
#     --binary-dir ./release \
#     --template-dir ./deploy/config-templates \
#     --output-dir ./deploy/configs \
#     --server-private-ip 10.0.1.10 \
#     --ac-private-ip 10.0.1.20 \
#     --domain opennhp.org \
#     [--regenerate]
#
# Requires: jq, aws cli, built nhp-serverd/nhp-acd/nhp-relayd binaries

set -euo pipefail

# --- Parse arguments ---
BINARY_DIR=""
TEMPLATE_DIR=""
OUTPUT_DIR=""
SERVER_PRIVATE_IP=""
AC_PRIVATE_IP=""
DOMAIN="opennhp.org"
REGENERATE=false
AWS_SECRET_ID="opennhp/demo"

while [[ $# -gt 0 ]]; do
  case $1 in
    --binary-dir)     BINARY_DIR="$2"; shift 2 ;;
    --template-dir)   TEMPLATE_DIR="$2"; shift 2 ;;
    --output-dir)     OUTPUT_DIR="$2"; shift 2 ;;
    --server-private-ip) SERVER_PRIVATE_IP="$2"; shift 2 ;;
    --ac-private-ip)  AC_PRIVATE_IP="$2"; shift 2 ;;
    --domain)         DOMAIN="$2"; shift 2 ;;
    --regenerate)     REGENERATE=true; shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# Validate required args
for var in BINARY_DIR TEMPLATE_DIR OUTPUT_DIR SERVER_PRIVATE_IP AC_PRIVATE_IP; do
  if [[ -z "${!var}" ]]; then
    echo "ERROR: --$(echo $var | tr '[:upper:]' '[:lower:]' | tr '_' '-') is required"
    exit 1
  fi
done

echo "=== NHP Key Generation & Config Rendering ==="
echo "  Binary dir:     $BINARY_DIR"
echo "  Template dir:   $TEMPLATE_DIR"
echo "  Output dir:     $OUTPUT_DIR"
echo "  Server IP:      $SERVER_PRIVATE_IP"
echo "  AC IP:          $AC_PRIVATE_IP"
echo "  Domain:         $DOMAIN"
echo "  Regenerate:     $REGENERATE"
echo ""

# --- Create output directories ---
mkdir -p "$OUTPUT_DIR/server"
mkdir -p "$OUTPUT_DIR/ac"
mkdir -p "$OUTPUT_DIR/relay"

# --- Fetch existing keys from AWS Secrets Manager ---
echo "Fetching secrets from AWS Secrets Manager..."
SECRETS_JSON=$(aws secretsmanager get-secret-value \
  --secret-id "$AWS_SECRET_ID" \
  --region us-east-2 \
  --query 'SecretString' \
  --output text 2>/dev/null || echo "{}")

# Extract existing keys (empty string if not present)
EXISTING_SERVER_PRIV=$(echo "$SECRETS_JSON" | jq -r '.nhp_server_private_key // empty')
EXISTING_SERVER_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_server_public_key // empty')
EXISTING_AC_PRIV=$(echo "$SECRETS_JSON" | jq -r '.nhp_ac_private_key // empty')
EXISTING_AC_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_ac_public_key // empty')
EXISTING_RELAY_PRIV=$(echo "$SECRETS_JSON" | jq -r '.nhp_relay_private_key // empty')
EXISTING_RELAY_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_relay_public_key // empty')
EXISTING_AGENT_PRIV=$(echo "$SECRETS_JSON" | jq -r '.nhp_agent_private_key // empty')
EXISTING_AGENT_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_agent_public_key // empty')
EXISTING_JSAGENT_PRIV=$(echo "$SECRETS_JSON" | jq -r '.nhp_jsagent_private_key // empty')
EXISTING_JSAGENT_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_jsagent_public_key // empty')

# --- Generate or reuse keys ---
generate_keys() {
  local binary="$1"
  local name="$2"
  local existing_priv="$3"
  local existing_pub="$4"

  if [[ "$REGENERATE" == "false" && -n "$existing_priv" && -n "$existing_pub" ]]; then
    echo "  Reusing existing $name keys from AWS SM" >&2
    echo "$existing_priv|$existing_pub"
    return
  fi

  echo "  Generating new $name keys..." >&2
  # Some binaries (e.g. nhp-acd) have init() side-effects that write log
  # lines to stdout before the JSON output. Extract the JSON line only.
  local raw_output keys_json
  raw_output=$("$binary" keygen --curve --json)
  keys_json=$(echo "$raw_output" | grep -E '^\{.*"privateKey".*\}$' | tail -1)
  if [ -z "$keys_json" ]; then
    echo "  ERROR: no JSON output from $name keygen; raw output was:" >&2
    echo "$raw_output" >&2
    return 1
  fi
  local priv pub
  priv=$(echo "$keys_json" | jq -r '.privateKey')
  pub=$(echo "$keys_json" | jq -r '.publicKey')
  echo "$priv|$pub"
}

echo "--- Generating/loading keys ---"

# Server keys
SERVER_KEYS=$(generate_keys "$BINARY_DIR/nhp-server/nhp-serverd" "server" "$EXISTING_SERVER_PRIV" "$EXISTING_SERVER_PUB")
NHP_SERVER_PRIVATE_KEY=$(echo "$SERVER_KEYS" | cut -d'|' -f1)
NHP_SERVER_PUBLIC_KEY=$(echo "$SERVER_KEYS" | cut -d'|' -f2)

# AC keys
AC_KEYS=$(generate_keys "$BINARY_DIR/nhp-ac/nhp-acd" "ac" "$EXISTING_AC_PRIV" "$EXISTING_AC_PUB")
NHP_AC_PRIVATE_KEY=$(echo "$AC_KEYS" | cut -d'|' -f1)
NHP_AC_PUBLIC_KEY=$(echo "$AC_KEYS" | cut -d'|' -f2)

# Relay keys
RELAY_KEYS=$(generate_keys "$BINARY_DIR/nhp-relay/nhp-relayd" "relay" "$EXISTING_RELAY_PRIV" "$EXISTING_RELAY_PUB")
NHP_RELAY_PRIVATE_KEY=$(echo "$RELAY_KEYS" | cut -d'|' -f1)
NHP_RELAY_PUBLIC_KEY=$(echo "$RELAY_KEYS" | cut -d'|' -f2)

# Agent keys (generated using serverd binary since agent binary is not built here)
AGENT_KEYS=$(generate_keys "$BINARY_DIR/nhp-server/nhp-serverd" "agent" "$EXISTING_AGENT_PRIV" "$EXISTING_AGENT_PUB")
NHP_AGENT_PRIVATE_KEY=$(echo "$AGENT_KEYS" | cut -d'|' -f1)
NHP_AGENT_PUBLIC_KEY=$(echo "$AGENT_KEYS" | cut -d'|' -f2)

# js-agent keys (browser-side client; private key consumed from AWS SM by the js-agent repo)
JSAGENT_KEYS=$(generate_keys "$BINARY_DIR/nhp-server/nhp-serverd" "js-agent" "$EXISTING_JSAGENT_PRIV" "$EXISTING_JSAGENT_PUB")
NHP_JSAGENT_PRIVATE_KEY=$(echo "$JSAGENT_KEYS" | cut -d'|' -f1)
NHP_JSAGENT_PUBLIC_KEY=$(echo "$JSAGENT_KEYS" | cut -d'|' -f2)

echo ""
echo "--- Key summary ---"
echo "  Server public key: ${NHP_SERVER_PUBLIC_KEY:0:20}..."
echo "  AC public key:     ${NHP_AC_PUBLIC_KEY:0:20}..."
echo "  Relay public key:  ${NHP_RELAY_PUBLIC_KEY:0:20}..."
echo "  Agent public key:    ${NHP_AGENT_PUBLIC_KEY:0:20}..."
echo "  js-agent public key: ${NHP_JSAGENT_PUBLIC_KEY:0:20}..."
echo ""

# --- Save keys to AWS Secrets Manager ---
echo "--- Saving keys to AWS Secrets Manager ---"

# Merge new keys into existing secrets (preserving cloudflare tokens etc.)
UPDATED_SECRETS=$(echo "$SECRETS_JSON" | jq \
  --arg sk "$NHP_SERVER_PRIVATE_KEY" \
  --arg sp "$NHP_SERVER_PUBLIC_KEY" \
  --arg ak "$NHP_AC_PRIVATE_KEY" \
  --arg ap "$NHP_AC_PUBLIC_KEY" \
  --arg rk "$NHP_RELAY_PRIVATE_KEY" \
  --arg rp "$NHP_RELAY_PUBLIC_KEY" \
  --arg agk "$NHP_AGENT_PRIVATE_KEY" \
  --arg agp "$NHP_AGENT_PUBLIC_KEY" \
  --arg jk "$NHP_JSAGENT_PRIVATE_KEY" \
  --arg jp "$NHP_JSAGENT_PUBLIC_KEY" \
  '. + {
    nhp_server_private_key: $sk,
    nhp_server_public_key: $sp,
    nhp_ac_private_key: $ak,
    nhp_ac_public_key: $ap,
    nhp_relay_private_key: $rk,
    nhp_relay_public_key: $rp,
    nhp_agent_private_key: $agk,
    nhp_agent_public_key: $agp,
    nhp_jsagent_private_key: $jk,
    nhp_jsagent_public_key: $jp
  }')

aws secretsmanager put-secret-value \
  --secret-id "$AWS_SECRET_ID" \
  --region us-east-2 \
  --secret-string "$UPDATED_SECRETS"

echo "  Keys saved to AWS Secrets Manager"
echo ""

# --- Render config templates ---
echo "--- Rendering config templates ---"

export NHP_SERVER_PRIVATE_KEY NHP_SERVER_PUBLIC_KEY
export NHP_AC_PRIVATE_KEY NHP_AC_PUBLIC_KEY
export NHP_RELAY_PRIVATE_KEY NHP_RELAY_PUBLIC_KEY
export NHP_AGENT_PRIVATE_KEY NHP_AGENT_PUBLIC_KEY
export NHP_JSAGENT_PRIVATE_KEY NHP_JSAGENT_PUBLIC_KEY
export SERVER_PRIVATE_IP="$SERVER_PRIVATE_IP"
export AC_PRIVATE_IP="$AC_PRIVATE_IP"
export DOMAIN="$DOMAIN"

# Render all templates
for component in server ac relay; do
  echo "  Rendering $component configs..."
  for template in "$TEMPLATE_DIR/$component"/*.toml; do
    filename=$(basename "$template")
    envsubst < "$template" > "$OUTPUT_DIR/$component/$filename"
    echo "    $filename"
  done
done

echo ""
echo "=== Done ==="
echo "Configs written to: $OUTPUT_DIR"
