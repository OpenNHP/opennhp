#!/usr/bin/env bash
# Populate node_modules/@opennhp/agent with the locally-built js-agent SDK.
# This avoids using npm's file: dependency (which fails on private packages
# with "Cannot read properties of undefined (reading 'extraneous')").
# vite.config.ts aliases @opennhp/agent to this directory.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WEB_DIR="$(dirname "$SCRIPT_DIR")"
REPO_ROOT="$(cd "$WEB_DIR/../.." && pwd)"
JS_AGENT_DIR="$REPO_ROOT/endpoints/js-agent"

if [ ! -d "$JS_AGENT_DIR/dist" ]; then
  echo "Building js-agent first (dist/ not found)…"
  (cd "$JS_AGENT_DIR" && npm ci --no-audit --no-fund && npm run build)
fi

TARGET="$WEB_DIR/node_modules/@opennhp/agent"
mkdir -p "$TARGET"
rm -rf "$TARGET"/*
cp -r "$JS_AGENT_DIR/dist" "$TARGET/dist"
cp "$JS_AGENT_DIR/package.json" "$TARGET/package.json"

echo "Installed @opennhp/agent from $JS_AGENT_DIR into $TARGET"