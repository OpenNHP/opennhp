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
#     --server2-private-ip 10.0.1.30 \
#     --ac2-private-ip 10.0.1.40 \
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
SERVER2_PRIVATE_IP=""
AC2_PRIVATE_IP=""
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
    --server2-private-ip) SERVER2_PRIVATE_IP="$2"; shift 2 ;;
    --ac2-private-ip) AC2_PRIVATE_IP="$2"; shift 2 ;;
    --domain)         DOMAIN="$2"; shift 2 ;;
    --regenerate)     REGENERATE=true; shift ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# Validate required args
for var in BINARY_DIR TEMPLATE_DIR OUTPUT_DIR SERVER_PRIVATE_IP AC_PRIVATE_IP SERVER2_PRIVATE_IP AC2_PRIVATE_IP; do
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
echo "  Server2 IP:     $SERVER2_PRIVATE_IP"
echo "  AC2 IP:         $AC2_PRIVATE_IP"
echo "  Domain:         $DOMAIN"
echo "  Regenerate:     $REGENERATE"
echo ""

# --- Create output directories ---
mkdir -p "$OUTPUT_DIR/server"
mkdir -p "$OUTPUT_DIR/ac"
mkdir -p "$OUTPUT_DIR/relay"
mkdir -p "$OUTPUT_DIR/server2"
mkdir -p "$OUTPUT_DIR/ac2"

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
EXISTING_JSAGENT_SM2_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_jsagent_sm2_public_key // empty')
# Cluster 2 js-agent: independent browser-demo identity so cluster 1 and
# cluster 2 do not share an agent key (each nhp-server trusts only its own).
EXISTING_JSAGENT2_PRIV=$(echo "$SECRETS_JSON" | jq -r '.nhp_jsagent2_private_key // empty')
EXISTING_JSAGENT2_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_jsagent2_public_key // empty')
EXISTING_JSAGENT2_SM2_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_jsagent2_sm2_public_key // empty')
# Server cluster 2 (independent key pairs; see CLAUDE.md opennhp/demo schema)
EXISTING_SERVER2_PRIV=$(echo "$SECRETS_JSON" | jq -r '.nhp_server2_private_key // empty')
EXISTING_SERVER2_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_server2_public_key // empty')
EXISTING_AC2_PRIV=$(echo "$SECRETS_JSON" | jq -r '.nhp_ac2_private_key // empty')
EXISTING_AC2_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_ac2_public_key // empty')
# SM2 public keys derived from server private keys (one private key → two public keys).
# These are populated from existing data when reusing keys, or derived after keygen --both.
EXISTING_SERVER_SM2_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_server_sm2_public_key // empty')
EXISTING_SERVER2_SM2_PUB=$(echo "$SECRETS_JSON" | jq -r '.nhp_server2_sm2_public_key // empty')

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

# generate_server_keys: uses --both to derive SM2 and Curve25519 public keys
# from a single SM2 private key. Returns priv|curve25519_pub|sm2_pub.
#
# Legacy secrets (created before dual-cipher support) store the private key
# and one public key but NOT the SM2 public key. The private key is
# scheme-agnostic — the same 32 bytes yield both an SM2 and a Curve25519
# public key — so when only the SM2 public key is missing we DERIVE it from
# the existing private key with `pubkey --both` rather than regenerating.
# Regenerating would mint a brand-new private key and silently rotate the
# server identity (breaking every pinned peer), which must only happen on an
# explicit --regenerate.
generate_server_keys() {
  local binary="$1"
  local name="$2"
  local existing_priv="$3"
  local existing_curve_pub="$4"
  local existing_sm2_pub="$5"

  if [[ "$REGENERATE" == "false" && -n "$existing_priv" && -n "$existing_curve_pub" && -n "$existing_sm2_pub" ]]; then
    echo "  Reusing existing $name keys from AWS SM" >&2
    echo "$existing_priv|$existing_curve_pub|$existing_sm2_pub"
    return
  fi

  # Backfill path: private key present but SM2 (and/or Curve25519) public key
  # missing. Derive the public keys from the STABLE existing private key.
  if [[ "$REGENERATE" == "false" && -n "$existing_priv" ]]; then
    echo "  Backfilling $name public keys from existing private key (no rotation)..." >&2
    local pub_raw pub_json curve_pub sm2_pub
    pub_raw=$("$binary" pubkey --both --json "$existing_priv")
    pub_json=$(echo "$pub_raw" | grep -E '^\{.*"sm2PublicKey".*\}$' | tail -1)
    if [ -n "$pub_json" ]; then
      curve_pub=$(echo "$pub_json" | jq -r '.curve25519PublicKey')
      sm2_pub=$(echo "$pub_json" | jq -r '.sm2PublicKey')
      if [[ -n "$curve_pub" && "$curve_pub" != "null" && -n "$sm2_pub" && "$sm2_pub" != "null" ]]; then
        echo "$existing_priv|$curve_pub|$sm2_pub"
        return
      fi
    fi
    echo "  WARNING: could not derive $name public keys from existing private key; raw output was:" >&2
    echo "$pub_raw" >&2
    echo "  Falling back to key generation (this ROTATES the $name private key)." >&2
  fi

  echo "  Generating new $name keys (--both)..." >&2
  local raw_output keys_json
  raw_output=$("$binary" keygen --both --json)
  keys_json=$(echo "$raw_output" | grep -E '^\{.*"privateKey".*\}$' | tail -1)
  if [ -z "$keys_json" ]; then
    echo "  ERROR: no JSON output from $name keygen --both; raw output was:" >&2
    echo "$raw_output" >&2
    return 1
  fi
  local priv curve_pub sm2_pub
  priv=$(echo "$keys_json" | jq -r '.privateKey')
  curve_pub=$(echo "$keys_json" | jq -r '.curve25519PublicKey')
  sm2_pub=$(echo "$keys_json" | jq -r '.sm2PublicKey')
  echo "$priv|$curve_pub|$sm2_pub"
}

echo "--- Generating/loading keys ---"

# Server keys: one SM2 private key → Curve25519 public key + SM2 public key
SERVER_KEYS=$(generate_server_keys "$BINARY_DIR/nhp-server/nhp-serverd" "server" "$EXISTING_SERVER_PRIV" "$EXISTING_SERVER_PUB" "$EXISTING_SERVER_SM2_PUB")
NHP_SERVER_PRIVATE_KEY=$(echo "$SERVER_KEYS" | cut -d'|' -f1)
NHP_SERVER_PUBLIC_KEY=$(echo "$SERVER_KEYS" | cut -d'|' -f2)
NHP_SERVER_SM2_PUBLIC_KEY=$(echo "$SERVER_KEYS" | cut -d'|' -f3)

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
# Uses --both so both Curve25519 and SM2 public keys are derived from the same private key.
# Both public keys are registered in server/agent.toml so the agent can knock in either cipher scheme.
JSAGENT_KEYS=$(generate_server_keys "$BINARY_DIR/nhp-server/nhp-serverd" "js-agent" "$EXISTING_JSAGENT_PRIV" "$EXISTING_JSAGENT_PUB" "$EXISTING_JSAGENT_SM2_PUB")
NHP_JSAGENT_PRIVATE_KEY=$(echo "$JSAGENT_KEYS" | cut -d'|' -f1)
NHP_JSAGENT_PUBLIC_KEY=$(echo "$JSAGENT_KEYS" | cut -d'|' -f2)
NHP_JSAGENT_SM2_PUBLIC_KEY=$(echo "$JSAGENT_KEYS" | cut -d'|' -f3)

# Cluster 2 js-agent keys (independent browser-demo identity; trusted only by
# server cluster 2, so the cluster 1 and cluster 2 demo agents are isolated)
JSAGENT2_KEYS=$(generate_server_keys "$BINARY_DIR/nhp-server/nhp-serverd" "js-agent2" "$EXISTING_JSAGENT2_PRIV" "$EXISTING_JSAGENT2_PUB" "$EXISTING_JSAGENT2_SM2_PUB")
NHP_JSAGENT2_PRIVATE_KEY=$(echo "$JSAGENT2_KEYS" | cut -d'|' -f1)
NHP_JSAGENT2_PUBLIC_KEY=$(echo "$JSAGENT2_KEYS" | cut -d'|' -f2)
NHP_JSAGENT2_SM2_PUBLIC_KEY=$(echo "$JSAGENT2_KEYS" | cut -d'|' -f3)

# Server cluster 2 keys (independent identity, isolated from cluster 1)
SERVER2_KEYS=$(generate_server_keys "$BINARY_DIR/nhp-server/nhp-serverd" "server2" "$EXISTING_SERVER2_PRIV" "$EXISTING_SERVER2_PUB" "$EXISTING_SERVER2_SM2_PUB")
NHP_SERVER2_PRIVATE_KEY=$(echo "$SERVER2_KEYS" | cut -d'|' -f1)
NHP_SERVER2_PUBLIC_KEY=$(echo "$SERVER2_KEYS" | cut -d'|' -f2)
NHP_SERVER2_SM2_PUBLIC_KEY=$(echo "$SERVER2_KEYS" | cut -d'|' -f3)

AC2_KEYS=$(generate_keys "$BINARY_DIR/nhp-ac/nhp-acd" "ac2" "$EXISTING_AC2_PRIV" "$EXISTING_AC2_PUB")
NHP_AC2_PRIVATE_KEY=$(echo "$AC2_KEYS" | cut -d'|' -f1)
NHP_AC2_PUBLIC_KEY=$(echo "$AC2_KEYS" | cut -d'|' -f2)

echo ""
echo "--- Key summary ---"
echo "  Server public key (Curve25519): ${NHP_SERVER_PUBLIC_KEY:0:20}..."
echo "  Server public key (SM2):        ${NHP_SERVER_SM2_PUBLIC_KEY:0:20}..."
echo "  AC public key:     ${NHP_AC_PUBLIC_KEY:0:20}..."
echo "  Relay public key:  ${NHP_RELAY_PUBLIC_KEY:0:20}..."
echo "  Agent public key:    ${NHP_AGENT_PUBLIC_KEY:0:20}..."
echo "  js-agent public key (Curve25519):  ${NHP_JSAGENT_PUBLIC_KEY:0:20}..."
echo "  js-agent public key (SM2):         ${NHP_JSAGENT_SM2_PUBLIC_KEY:0:20}..."
echo "  js-agent2 public key (Curve25519): ${NHP_JSAGENT2_PUBLIC_KEY:0:20}..."
echo "  js-agent2 public key (SM2):        ${NHP_JSAGENT2_SM2_PUBLIC_KEY:0:20}..."
echo "  Server2 public key (Curve25519): ${NHP_SERVER2_PUBLIC_KEY:0:20}..."
echo "  Server2 public key (SM2):        ${NHP_SERVER2_SM2_PUBLIC_KEY:0:20}..."
echo "  AC2 public key:      ${NHP_AC2_PUBLIC_KEY:0:20}..."
echo ""

# --- Save keys to AWS Secrets Manager ---
echo "--- Saving keys to AWS Secrets Manager ---"

# Merge new keys into existing secrets (preserving cloudflare tokens etc.)
UPDATED_SECRETS=$(echo "$SECRETS_JSON" | jq \
  --arg sk "$NHP_SERVER_PRIVATE_KEY" \
  --arg sp "$NHP_SERVER_PUBLIC_KEY" \
  --arg ssp "$NHP_SERVER_SM2_PUBLIC_KEY" \
  --arg ak "$NHP_AC_PRIVATE_KEY" \
  --arg ap "$NHP_AC_PUBLIC_KEY" \
  --arg rk "$NHP_RELAY_PRIVATE_KEY" \
  --arg rp "$NHP_RELAY_PUBLIC_KEY" \
  --arg agk "$NHP_AGENT_PRIVATE_KEY" \
  --arg agp "$NHP_AGENT_PUBLIC_KEY" \
  --arg jk "$NHP_JSAGENT_PRIVATE_KEY" \
  --arg jp "$NHP_JSAGENT_PUBLIC_KEY" \
  --arg jsp "$NHP_JSAGENT_SM2_PUBLIC_KEY" \
  --arg j2k "$NHP_JSAGENT2_PRIVATE_KEY" \
  --arg j2p "$NHP_JSAGENT2_PUBLIC_KEY" \
  --arg j2sp "$NHP_JSAGENT2_SM2_PUBLIC_KEY" \
  --arg s2k "$NHP_SERVER2_PRIVATE_KEY" \
  --arg s2p "$NHP_SERVER2_PUBLIC_KEY" \
  --arg s2sp "$NHP_SERVER2_SM2_PUBLIC_KEY" \
  --arg a2k "$NHP_AC2_PRIVATE_KEY" \
  --arg a2p "$NHP_AC2_PUBLIC_KEY" \
  '. + {
    nhp_server_private_key: $sk,
    nhp_server_public_key: $sp,
    nhp_server_sm2_public_key: $ssp,
    nhp_ac_private_key: $ak,
    nhp_ac_public_key: $ap,
    nhp_relay_private_key: $rk,
    nhp_relay_public_key: $rp,
    nhp_agent_private_key: $agk,
    nhp_agent_public_key: $agp,
    nhp_jsagent_private_key: $jk,
    nhp_jsagent_public_key: $jp,
    nhp_jsagent_sm2_public_key: $jsp,
    nhp_jsagent2_private_key: $j2k,
    nhp_jsagent2_public_key: $j2p,
    nhp_jsagent2_sm2_public_key: $j2sp,
    nhp_server2_private_key: $s2k,
    nhp_server2_public_key: $s2p,
    nhp_server2_sm2_public_key: $s2sp,
    nhp_ac2_private_key: $a2k,
    nhp_ac2_public_key: $a2p
  }')

aws secretsmanager put-secret-value \
  --secret-id "$AWS_SECRET_ID" \
  --region us-east-2 \
  --secret-string "$UPDATED_SECRETS"

echo "  Keys saved to AWS Secrets Manager"
echo ""

# --- Render config templates ---
echo "--- Rendering config templates ---"

export NHP_SERVER_PRIVATE_KEY NHP_SERVER_PUBLIC_KEY NHP_SERVER_SM2_PUBLIC_KEY
export NHP_AC_PRIVATE_KEY NHP_AC_PUBLIC_KEY
export NHP_RELAY_PRIVATE_KEY NHP_RELAY_PUBLIC_KEY
export NHP_AGENT_PRIVATE_KEY NHP_AGENT_PUBLIC_KEY
export NHP_JSAGENT_PRIVATE_KEY NHP_JSAGENT_PUBLIC_KEY NHP_JSAGENT_SM2_PUBLIC_KEY
export NHP_JSAGENT2_PRIVATE_KEY NHP_JSAGENT2_PUBLIC_KEY NHP_JSAGENT2_SM2_PUBLIC_KEY
export NHP_SERVER2_PRIVATE_KEY NHP_SERVER2_PUBLIC_KEY NHP_SERVER2_SM2_PUBLIC_KEY
export NHP_AC2_PRIVATE_KEY NHP_AC2_PUBLIC_KEY
export SERVER_PRIVATE_IP="$SERVER_PRIVATE_IP"
export AC_PRIVATE_IP="$AC_PRIVATE_IP"
export SERVER2_PRIVATE_IP="$SERVER2_PRIVATE_IP"
export AC2_PRIVATE_IP="$AC2_PRIVATE_IP"
export DOMAIN="$DOMAIN"

# Render all templates. cluster 2 (server2/ac2) reuses the same key/IP
# env vars exported above; the relay config references both clusters.
for component in server ac relay server2 ac2; do
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
