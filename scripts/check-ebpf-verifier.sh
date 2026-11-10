#!/usr/bin/env bash
# Ask the kernel whether it will load an eBPF object.
#
# Compiling an object proves nothing about it: the kernel's verifier decides
# whether a program is loadable at all, and it can refuse a program that
# compiled without a warning (an instruction or complexity limit, a helper the
# program type does not allow). On the nhp-ac a refused object is an outage --
# nhp-acd cannot start, the fail-closed backstop keeps tcp/443 shut -- and on
# the nhp-server it is a silent loss of protection, because nhp-serverd falls
# open.
#
# This script is the manual form of the gate deploy-demo-v2 runs on the demo
# hosts: it loads the object with bpftool, never attaches it, and removes its
# pins again. Nothing on the host is changed, so it can be run on a live host
# without touching the filter that is running there.
#
#   ./scripts/check-ebpf-verifier.sh release/nhp-ac/etc/nhp_ebpf_xdp.o
#
# Exit status:
#   0  the kernel accepted the object
#   1  the kernel refused it (the verifier log is on stderr)
#   2  the script was called with the wrong arguments
#   3  bpftool is not available on this host
set -euo pipefail

obj="${1:-}"
if [ -z "$obj" ] || [ ! -f "$obj" ]; then
  echo "usage: $0 <eBPF object>" >&2
  exit 2
fi

# bpftool is the kernel's own loader and prints the verifier log. Amazon Linux
# 2023 ships it in its core repository, versioned for the kernel stream it
# belongs to (bpftool6.12 for the 6.12 stream the demo runs).
bpftool_bin=""
if command -v bpftool >/dev/null 2>&1; then
  bpftool_bin="bpftool"
else
  for candidate in /usr/sbin/bpftool* /usr/bin/bpftool*; do
    if [ -x "$candidate" ]; then
      bpftool_bin="$candidate"
      break
    fi
  done
fi
if [ -z "$bpftool_bin" ]; then
  echo "bpftool is not installed; on Amazon Linux 2023: sudo dnf install -y bpftool6.12" >&2
  exit 3
fi

# The pins go under a directory of their own, so the pins a running daemon
# owns cannot be touched, and they are removed again on every exit.
pin_dir="/sys/fs/bpf/nhp-verify-$$"
cleanup() {
  sudo rm -rf "$pin_dir" >/dev/null 2>&1 || true
}
trap cleanup EXIT

sudo mkdir -p "$pin_dir"
if sudo "$bpftool_bin" prog loadall "$obj" "$pin_dir"; then
  echo "The kernel accepted $obj"
  exit 0
fi

echo "The kernel refused $obj; the verifier log is above" >&2
exit 1
