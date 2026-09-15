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
(( GLOBAL_RATE <= 10000 && GLOBAL_BURST <= 10000 && GLOBAL_BURST <= GLOBAL_RATE * 60 )) || { echo "rate must be <= 10000 pps and burst <= 10000 and <= 60 seconds of traffic" >&2; exit 2; }
for tool in iptables ip6tables iptables-restore ip6tables-restore sysctl python3; do
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
        if "." not in peer and ":" not in peer:
            raise ValueError("peer must be an explicit IP address or CIDR")
        net = ipaddress.ip_network(peer, strict=True)
        if net.prefixlen < (24 if net.version == 4 else 64):
            raise ValueError("trusted CIDRs must be /24 or narrower for IPv4, /64 or narrower for IPv6")
        print(net.version, net)
except ValueError as error:
    sys.exit(str(error))
PYCODE
)

rules_dir=$(mktemp -d)
trap 'rm -rf "$rules_dir"' EXIT
for ipt in iptables ip6tables; do
  family=4
  [[ "$ipt" == ip6tables ]] && family=6
  {
    echo '*filter'
    echo ":$CHAIN - [0:0]"
    echo "-A $CHAIN -i lo -j RETURN"
    while read -r peer_family peer; do
      if [[ "$peer_family" == "$family" ]]; then
        echo "-A $CHAIN -s $peer -j RETURN"
      fi
    done <<< "$PEERS"
    echo "-A $CHAIN -m limit --limit $GLOBAL_RATE/second --limit-burst $GLOBAL_BURST -j RETURN"
    echo "-A $CHAIN -j DROP"
    if ! "$ipt" -w -C INPUT -p udp --dport "$PORT" -j "$CHAIN" >/dev/null 2>&1; then
      echo "-I INPUT 1 -p udp --dport $PORT -j $CHAIN"
    fi
    echo COMMIT
  } > "$rules_dir/$ipt"
  "$ipt-restore" --test --wait --noflush < "$rules_dir/$ipt"
done
current=$(sysctl -n net.core.rmem_max)
[[ "$current" =~ ^[0-9]{1,10}$ ]] || { echo "invalid current rmem_max" >&2; exit 1; }
if (( current < RECV_BUFFER )); then
  sysctl -w "net.core.rmem_max=$RECV_BUFFER"
fi

# Each restore replaces only this chain atomically within one IP family.
# The two families cannot commit together. Report partial application clearly.
for ipt in iptables ip6tables; do
  if ! "$ipt-restore" --wait --noflush < "$rules_dir/$ipt"; then
    if [[ "$ipt" == ip6tables ]]; then
      echo "IPv4 applied, IPv6 FAILED: host is in a mixed firewall state. Correct the IPv6 error and rerun this helper." >&2
    else
      echo "IPv4 FAILED: neither family updated. Correct the error and rerun this helper." >&2
    fi
    echo "The receive-buffer sysctl may already have been raised." >&2
    exit 1
  fi
done

echo "UDP $PORT protected in IPv4 and IPv6; untrusted aggregate=${GLOBAL_RATE}/s per family, burst=$GLOBAL_BURST."
echo "Restart nhp-server to apply the receive buffer. Persist sysctl and firewall rules with your host configuration manager."
