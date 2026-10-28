#!/usr/bin/env bash
# Summarise what the nhp-server XDP filter has counted, from its event logs.
#
# Read-only: it ssh's through the relay (ssh -J), cats
# ~/nhp-server/logs/nhp_server_xdp-<date>.log and aggregates locally. Nothing is
# written to, or changed on, the host.
#
#   RELAY_PUB=<relay public ip/name> SERVER_PRIV=<server private ip> \
#     scripts/xdp-attack-report.sh [FROM_DATE [TO_DATE]]      # YYYY-MM-DD, default: today
#
# Or feed it log files you already have:
#   scripts/xdp-attack-report.sh --files nhp_server_xdp-*.log
#
# Tables:
#   1. verdict classes, from [NHP-STAT] (exact; LOGGED vs PKTS shows what the
#      per-packet rate limit swallowed)
#   2. top destination ports, from [NHP-SCAN] KIND=PORT
#   3. top sources, from [NHP-SCAN] KIND=SRC, with the UNIQ_SRC trend
# Class totals are always exact; the port/source tables are Top-N per window and
# approximate whenever a window reported LRU=1.
set -euo pipefail

TOPN="${TOPN:-20}"
SSH_USER="${SSH_USER:-ec2-user}"
LOG_DIR="${LOG_DIR:-nhp-server/logs}"

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

if [[ "${1:-}" == "--files" ]]; then
  shift
  cat "$@" > "$tmp"
else
  : "${RELAY_PUB:?set RELAY_PUB (relay public address)}"
  : "${SERVER_PRIV:?set SERVER_PRIV (server private address)}"
  from="${1:-$(date +%F)}"
  to="${2:-$from}"
  d="$from"
  ssh_opts=(-o BatchMode=yes -o ConnectTimeout=15 -J "$SSH_USER@$RELAY_PUB")
  while :; do
    # Fixed, validated path component: the date comes from date(1) below, never
    # from the caller's string.
    ssh "${ssh_opts[@]}" "$SSH_USER@$SERVER_PRIV" \
      "cat '$LOG_DIR/nhp_server_xdp-$d.log' 2>/dev/null || true" >> "$tmp"
    [[ "$d" == "$to" ]] && break
    d="$(date -d "$d + 1 day" +%F)"
    [[ "$d" > "$to" ]] && break
  done
fi

echo "== 1. Verdict classes ([NHP-STAT]) =="
grep -h '\[NHP-STAT\]' "$tmp" | awk '
  { v=""; r=""; p=0; l=0
    for (i=1;i<=NF;i++) {
      if ($i ~ /^VERDICT=/) v=substr($i,9)
      else if ($i ~ /^REASON=/) r=substr($i,8)
      else if ($i ~ /^PKTS=/) p=substr($i,6)
      else if ($i ~ /^LOGGED=/) l=substr($i,8)
    }
    k=v " " r; pk[k]+=p; lg[k]+=l }
  END { printf "%-6s %-20s %12s %12s %12s\n","VERDICT","REASON","PKTS","LOGGED","SUPPRESSED"
        for (k in pk) { split(k,a," "); printf "%-6s %-20s %12d %12d %12d\n",a[1],a[2],pk[k],lg[k],pk[k]-lg[k] } }' |
  { read -r h; echo "$h"; sort -k3,3nr; }

echo
echo "== 2. Top destination ports ([NHP-SCAN] KIND=PORT; Top-N per window, summed) =="
grep -h '\[NHP-SCAN\].*KIND=PORT' "$tmp" | awk '
  { pr=""; d=""; p=0
    for (i=1;i<=NF;i++) {
      if ($i ~ /^PROTO=/) pr=substr($i,7)
      else if ($i ~ /^DPT=/) d=substr($i,5)
      else if ($i ~ /^PKTS=/) p=substr($i,6)
    }
    h[pr "/" d]+=p }
  END { for (k in h) printf "%-14s %12d\n",k,h[k] }' | sort -k2,2nr | head -n "$TOPN"

echo
echo "== 3. Top sources ([NHP-SCAN] KIND=SRC; Top-N per window, summed) =="
grep -h '\[NHP-SCAN\].*KIND=SRC' "$tmp" | awk '
  { s=""; p=0
    for (i=1;i<=NF;i++) {
      if ($i ~ /^SRC=/) s=substr($i,5)
      else if ($i ~ /^PKTS=/) p=substr($i,6)
    }
    h[s]+=p }
  END { for (k in h) printf "%-18s %12d\n",k,h[k] }' | sort -k2,2nr | head -n "$TOPN"

echo
echo "== UNIQ_SRC / windows with LRU=1 ([NHP-SCAN] KIND=TOTAL) =="
grep -h '\[NHP-SCAN\].*KIND=TOTAL' "$tmp" | awk '
  { u=0; l=0
    for (i=1;i<=NF;i++) {
      if ($i ~ /^UNIQ_SRC=/) u=substr($i,10)
      else if ($i ~ /^LRU=/) l=substr($i,5)
    }
    n++; if (u>mx) mx=u; sum+=u; lru+=l }
  END { if (n) printf "windows=%d max_uniq_src=%d avg_uniq_src=%.0f lru_windows=%d\n",n,mx,sum/n,lru
        else print "no [NHP-SCAN] lines (filter older than this build, or no traffic yet)" }'
