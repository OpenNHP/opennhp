#!/bin/bash
# cleanup-ac-iptables.sh
#
# Remove any iptables/ipset rules + ipsets left behind by a previous
# iptables-mode NHP-AC install. Idempotent: safe to re-run.
#
# Called from the demo deploy pipeline (deploy-demo-v2.yml) just before
# the AC is started in eBPF mode so XDP DROP and iptables ACCEPT can
# never coexist on the same host (which would silently blackhole
# traffic while iptables looks correct).
#
# We deliberately keep `iptables`/`ipset` binaries installed in case
# other system tooling on the AC host relies on them — only the
# NHP-owned chains/sets are torn down.

set -euo pipefail

log() { printf '[cleanup-ac-iptables] %s\n' "$*"; }

# --- flush + delete NHP_BLOCK / NHP_DENY chains across all iptables tables
for tbl in filter nat mangle raw; do
  if command -v iptables >/dev/null 2>&1; then
    iptables -t "$tbl" -F NHP_BLOCK  2>/dev/null || true
    iptables -t "$tbl" -X NHP_BLOCK  2>/dev/null || true
    iptables -t "$tbl" -F NHP_DENY   2>/dev/null || true
    iptables -t "$tbl" -X NHP_DENY   2>/dev/null || true
  fi
  if command -v ip6tables >/dev/null 2>&1; then
    ip6tables -t "$tbl" -F NHP_BLOCK 2>/dev/null || true
    ip6tables -t "$tbl" -X NHP_BLOCK 2>/dev/null || true
    ip6tables -t "$tbl" -F NHP_DENY  2>/dev/null || true
    ip6tables -t "$tbl" -X NHP_DENY  2>/dev/null || true
  fi
done

# --- destroy NHP-owned ipsets
if command -v ipset >/dev/null 2>&1; then
  for s in defaultset defaultset_down defaultset_v6 defaultset_down_v6 \
           tempset tempset_v6 whitelistset blacklistset; do
    ipset destroy "$s" 2>/dev/null || true
  done
fi

# --- detach any leftover XDP/TC from previous runs (best-effort)
if command -v ip >/dev/null 2>&1; then
  for dev in $(ip -o link show 2>/dev/null | awk -F': ' '{print $2}'); do
    ip link set dev "$dev" xdp off        2>/dev/null || true
    ip link set dev "$dev" xdpgeneric off 2>/dev/null || true
  done
fi

# --- unpin eBPF objects/maps under /sys/fs/bpf (best-effort)
for p in xdp_white_prog tc_egress_prog \
         spp icmpwhitelist sdwhitelist src_port port_list protocol_port \
         conn_track events; do
  rm -f "/sys/fs/bpf/${p}" 2>/dev/null || true
done

log "iptables/ipset state and stale eBPF pins cleaned"