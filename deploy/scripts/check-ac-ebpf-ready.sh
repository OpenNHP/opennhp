#!/bin/bash
# check-ac-ebpf-ready.sh
#
# Verify the AC host is prepared for the eBPF/XDP data path before
# nhp-acd is started. Exits non-zero on the first failure so the
# deploy pipeline aborts instead of starting a daemon that will then
# crash in EbpfEngineLoad.
#
# Checks:
#   1. /sys/fs/bpf is mounted and writable
#   2. bpftool is available (for ops/debugging; not strictly required by
#      nhp-acd itself, but its absence usually means the host is missing
#      kernel-tools and we want a loud failure)
#   3. Kernel >= 5.6 (XDP generic minimum); warn on < 5.8 because
#      CAP_BPF only landed in 5.8

set -euo pipefail

log() { printf '[check-ac-ebpf-ready] %s\n' "$*"; }
fail() { printf '[check-ac-ebpf-ready] ERROR: %s\n' "$*" >&2; exit 1; }

# --- 1. bpffs mounted + writable
if ! mountpoint -q /sys/fs/bpf 2>/dev/null; then
  log "/sys/fs/bpf not mounted, attempting to mount"
  if ! mount -t bpf bpf /sys/fs/bpf 2>/dev/null; then
    fail "/sys/fs/bpf cannot be mounted (bpf filesystem unsupported?)"
  fi
fi

if [ ! -w /sys/fs/bpf ]; then
  fail "/sys/fs/bpf not writable by current user"
fi

# --- 2. bpftool present
if ! command -v bpftool >/dev/null 2>&1; then
  log "bpftool not found in PATH; install kernel-tools/bpftool for diagnostics"
  # not a hard fail: nhp-acd doesn't shell out to it
fi

# --- 3. kernel version
KVER="$(uname -r)"
KMAJOR="$(echo "$KVER" | cut -d. -f1)"
KMINOR="$(echo "$KVER" | cut -d. -f2)"
log "Kernel: $KVER"

if [ "${KMAJOR:-0}" -lt 5 ] || { [ "${KMAJOR:-0}" -eq 5 ] && [ "${KMINOR:-0}" -lt 6 ]; }; then
  fail "kernel $KVER < 5.6 — XDP generic mode not supported"
fi

if [ "${KMAJOR:-0}" -lt 5 ] || { [ "${KMAJOR:-0}" -eq 5 ] && [ "${KMINOR:-0}" -lt 8 ]; }; then
  log "WARNING: kernel $KVER < 5.8 — CAP_BPF unavailable, nhp-acd.service must grant CAP_SYS_ADMIN instead"
fi

# --- 4. default route interface readable (so nhp-acd can attach XDP)
if ! ip route show default >/dev/null 2>&1; then
  fail "no default IPv4 route; XDP cannot be attached to a default egress interface"
fi

DEFAULT_IF="$(ip route show default | head -1 | awk '{for (i=1;i<=NF;i++) if ($i=="dev") {print $(i+1); exit}}')"
if [ -z "${DEFAULT_IF:-}" ]; then
  fail "could not parse default route interface"
fi
log "Default route interface: $DEFAULT_IF"

log "eBPF readiness OK"