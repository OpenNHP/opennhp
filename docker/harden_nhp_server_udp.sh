#!/usr/bin/env bash
set -euo pipefail

# Host-network/bare-metal NHP server hardening. Docker bridge deployments use
# the Compose sysctl below for SO_RCVBUF, but their host firewall placement is
# environment-specific; run this script only where UDP 62206 reaches INPUT.
PORT="${NHP_KNOCK_PORT:-62206}"
GLOBAL_RATE="${NHP_KNOCK_GLOBAL_RATE_PPS:-5000}"
GLOBAL_BURST="${NHP_KNOCK_GLOBAL_RATE_BURST:-10000}"
PER_IP_RATE="${NHP_KNOCK_PER_IP_RATE_PPS:-100}"
PER_IP_BURST="${NHP_KNOCK_PER_IP_RATE_BURST:-50}"
RECV_BUFFER="${NHP_UDP_RECV_BUFFER_BYTES:-8388608}"
CHAIN="NHP_KNOCK_GUARD"

for value in "$PORT" "$GLOBAL_RATE" "$GLOBAL_BURST" "$PER_IP_RATE" "$PER_IP_BURST" "$RECV_BUFFER"; do
  [[ "$value" =~ ^[1-9][0-9]*$ ]] || { echo "all hardening values must be positive integers" >&2; exit 2; }
done
(( PORT <= 65535 )) || { echo "NHP_KNOCK_PORT must be <= 65535" >&2; exit 2; }

command -v iptables >/dev/null || { echo "iptables is required" >&2; exit 1; }
command -v sysctl >/dev/null || { echo "sysctl is required" >&2; exit 1; }

sysctl -w "net.core.rmem_max=$RECV_BUFFER"

iptables -w -N "$CHAIN" 2>/dev/null || true
iptables -w -F "$CHAIN"
iptables -w -A "$CHAIN" -m hashlimit \
  --hashlimit-above "$GLOBAL_RATE/sec" --hashlimit-burst "$GLOBAL_BURST" \
  --hashlimit-mode dstip --hashlimit-name nhp_knock_gbl -j DROP
iptables -w -A "$CHAIN" -m hashlimit \
  --hashlimit-above "$PER_IP_RATE/sec" --hashlimit-burst "$PER_IP_BURST" \
  --hashlimit-mode srcip --hashlimit-name nhp_knock_ip -j DROP
iptables -w -A "$CHAIN" -j RETURN

iptables -w -C INPUT -p udp --dport "$PORT" -j "$CHAIN" 2>/dev/null || \
  iptables -w -I INPUT 1 -p udp --dport "$PORT" -j "$CHAIN"

echo "NHP UDP hardening active on port $PORT: global=${GLOBAL_RATE}/s burst=$GLOBAL_BURST, per-IP=${PER_IP_RATE}/s burst=$PER_IP_BURST, rmem_max=$RECV_BUFFER"
