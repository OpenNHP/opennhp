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
# hosts: it loads the object the way the daemon loads it, never attaches it,
# and removes its pins again. Nothing on the host is changed, so it can be
# run on a live host without touching the filter that is running there.
#
# It used to do this with bpftool, and could not: bpftool's libbpf refuses
# these objects outright ("program 'xdp_white_prog' is static and not
# supported"), because the programs are declared static in the C sources. It
# also pins a LIBBPF_PIN_BY_NAME map into /sys/fs/bpf/<name> -- the very
# entries the running daemon owns -- so a load would bind to the live filter's
# maps instead of to its own. The verifier asks the kernel through the loader
# the daemons use instead, which supports the static programs and pins into a
# directory of its own.
#
# The object has to have been built already (make ebpf-objs); this script
# only asks the kernel about it.
#
#   ./scripts/check-ebpf-verifier.sh --variant ac \
#       --obj release/nhp-ac/etc/nhp_ebpf_xdp.o \
#       --tc-obj release/nhp-ac/etc/tc_egress.o
#   ./scripts/check-ebpf-verifier.sh --variant server \
#       --obj release/nhp-server/etc/nhp_server_xdp.o --port 62206
#
# It needs CAP_BPF, which in practice means root: a load without it is
# refused with EPERM, and that says nothing about the program.
#
# Exit status:
#   0  the kernel accepted the object
#   1  the kernel refused it (the verifier log is on stderr)
#   2  the arguments were wrong
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
usage: check-ebpf-verifier.sh --variant ac|server --obj <file.o> [--tc-obj <file.o>]
                               [--port N] [--min-frame-bytes N]

  --variant           which filter's object to check: "ac" or "server"
  --obj               the XDP object to check
  --tc-obj            the TC egress object to check (the AC only)
  --port              the UDP port the daemon listens on (the server only)
  --min-frame-bytes   the shortest datagram it accepts on that port
EOF
}

if [ "$#" -eq 0 ]; then
  usage
  exit 2
fi

repo_root=$(cd "$(dirname "$0")/.." && pwd)
verify="$repo_root/release/nhp-ebpf-verify/nhp-ebpf-verify"

# The verifier lives in the nhp module, next to the loader it drives. It is
# built here rather than shipped, so that the check is run against the loader
# in this checkout and not against a copy of it.
if [ ! -x "$verify" ]; then
  echo "Building nhp-ebpf-verify..." >&2
  (cd "$repo_root" && make ebpf-verify)
fi

if [ "$(id -u)" -ne 0 ]; then
  exec sudo "$verify" "$@"
fi
exec "$verify" "$@"
