#!/usr/bin/env bash
set -euo pipefail

# Host-network/bare-metal only: bridge traffic normally traverses FORWARD.
PORT="${NHP_KNOCK_PORT:-62206}"
GLOBAL_RATE="${NHP_KNOCK_GLOBAL_RATE_PPS:-5000}"
GLOBAL_BURST="${NHP_KNOCK_GLOBAL_RATE_BURST:-10000}"
RECV_BUFFER="${NHP_UDP_RECV_BUFFER_BYTES:-8388608}"
CHAIN="NHP_KNOCK_GUARD"

for value in "$PORT" "$GLOBAL_RATE" "$GLOBAL_BURST" "$RECV_BUFFER"; do
  [[ "$value" =~ ^[1-9][0-9]{0,9}$ ]] || { echo "hardening values must be positive integers of at most 10 digits" >&2; exit 2; }
done
(( PORT <= 65535 )) || { echo "NHP_KNOCK_PORT must be <= 65535" >&2; exit 2; }
(( RECV_BUFFER >= 65536 && RECV_BUFFER <= 1073741823 )) || { echo "receive buffer must be 65536..1073741823" >&2; exit 2; }
for tool in iptables ip6tables sysctl python3; do
  command -v "$tool" >/dev/null || { echo "$tool is required (both IP families must be protected)" >&2; exit 1; }
done
# Validate every peer before changing either firewall. Hostnames are not allowed.
PEERS=$(python3 - <<'PYCODE'
import ipaddress, os, sys
peers = os.environ.get("NHP_TRUSTED_PEERS", "").replace(",", " ").split()
if not peers:
    sys.exit("set NHP_TRUSTED_PEERS to the AC, relay and peer-server IPs/CIDRs before enabling limits")
try:
    for peer in peers:
        net = ipaddress.ip_network(peer, strict=False)
        if net.prefixlen == 0:
            raise ValueError("a default route would exempt all traffic")
        print(net.version, net)
except ValueError as error:
    sys.exit(str(error))
PYCODE
)

sysctl -w "net.core.rmem_max=$RECV_BUFFER"
for ipt in iptables ip6tables; do
  family=4
  [[ "$ipt" == ip6tables ]] && family=6
  "$ipt" -w -N "$CHAIN" 2>/dev/null || "$ipt" -w -S "$CHAIN" >/dev/null
  "$ipt" -w -F "$CHAIN"
  while read -r peer_family peer; do
    if [[ "$peer_family" == "$family" ]]; then
      "$ipt" -w -A "$CHAIN" -s "$peer" -j RETURN
    fi
  done <<< "$PEERS"
  # One fixed-size bucket per family: no spoofable per-source allocation table.
  # RETURN preserves the host's remaining INPUT rules; it does not grant access.
  "$ipt" -w -A "$CHAIN" -m limit --limit "$GLOBAL_RATE/second" --limit-burst "$GLOBAL_BURST" -j RETURN
  "$ipt" -w -A "$CHAIN" -j DROP
  "$ipt" -w -C INPUT -p udp --dport "$PORT" -j "$CHAIN" 2>/dev/null || \
    "$ipt" -w -I INPUT 1 -p udp --dport "$PORT" -j "$CHAIN"
done

echo "UDP $PORT protected in IPv4 and IPv6; untrusted aggregate=${GLOBAL_RATE}/s per family, burst=$GLOBAL_BURST."
echo "Restart nhp-server to apply the receive buffer. Persist sysctl and firewall rules with your host configuration manager."
