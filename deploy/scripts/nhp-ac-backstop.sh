#!/bin/bash
# Fail-closed netfilter backstop for the eBPF/XDP access controller.
#
# In FilterMode = 1 nhp-acd enforces its per-knock whitelist with an XDP
# program whose link lives only in process memory: endpoints/ac/ebpf/ebpfegine.go
# attaches with link.AttachXDP / link.AttachTCX and keeps the returned links in
# package variables, so link lifetime == process lifetime. Only the *programs*
# are pinned under /sys/fs/bpf, which keeps the objects alive but attaches
# nothing. Enforcement therefore disappears the moment the daemon goes away:
# a `systemctl stop`, a crash plus Restart=on-failure backoff, or a failed
# deploy would all leave the protected port open to the world (the AC security
# group allows tcp/443 from 0.0.0.0/0). The old iptables mode did not have that
# problem: iptables_default.sh left a default-DROP baseline in the kernel that
# outlived the process, and udpac.go even deferred iptables.ResetAllInput().
#
# This script restores that property from outside the daemon: a small netfilter
# chain that DROPs the protected ports, installed whenever nhp-acd is not known
# to be enforcing and lifted only once the XDP program is confirmed attached.
# It is wired into the unit by
# /etc/systemd/system/nhp-acd.service.d/10-ebpf-backstop.conf, written by the
# deploy-ac job in .github/workflows/deploy-demo-v2.yml:
#
#   ExecStartPre=+  ... up            close before the daemon starts
#   ExecStartPost=+ ... wait-attach   open once XDP is really attached
#   ExecStopPost=+  ... up            close again when the daemon goes away
#
# IPv6 is a special case: the XDP program returns XDP_PASS for every
# ETH_P_IPV6 frame (nhp/ebpf/xdp/nhp_ebpf_xdp.c), i.e. it does not filter IPv6
# at all. So the v6 backstop is *permanent* - `up` installs it and `down`
# deliberately leaves it in place.
#
# Commands:
#   up            install the backstop (idempotent). Verifies the rules are
#                 really in the kernel afterwards and exits non-zero if not,
#                 so ExecStartPre aborts the start rather than letting the
#                 daemon come up behind a backstop that was never installed.
#   down          lift the IPv4 backstop, keep the IPv6 one (idempotent).
#                 Non-zero if the jump is still there afterwards, i.e. the
#                 protected ports are still closed.
#   wait-attach   wait for the XDP program, then `down`. Non-zero exit on
#                 timeout, which makes systemd fail the unit with the backstop
#                 still in place (fail closed, and the deploy notices).
#   flush-legacy  remove the iptables_default.sh baseline (chains, ipset rules,
#                 DROP policies) while preserving the backstop chains
#   status        print the current state
#
# Environment:
#   NHP_BACKSTOP_PORTS    tcp ports to guard, space or comma separated
#                         (default: 443 - the only world-reachable protected
#                         port in the demo AC security group)
#   NHP_BACKSTOP_TIMEOUT  wait-attach timeout in seconds (default: 30)
#
# Must run as root; the systemd unit lines use the "+" prefix for that.

set -uo pipefail

CHAIN="NHP_BACKSTOP"
PORTS="${NHP_BACKSTOP_PORTS:-443}"
TIMEOUT="${NHP_BACKSTOP_TIMEOUT:-30}"
PIN_XDP="/sys/fs/bpf/xdp_white_prog"
PIN_TC="/sys/fs/bpf/tc_egress_prog"

# ipsets created by iptables_default.sh (both families).
LEGACY_SETS="defaultset defaultset_down tempset defaultset_v6 defaultset_down_v6 tempset_v6"
# Rules created by iptables_default.sh: ipset matches, the NHP_DENY chain and
# its references, and the kernel-log rules feeding rsyslog.
LEGACY_RULE_RE='match-set|NHP_DENY|\[NHP-(ACCEPT|DENY|FORWARD)\]'

log() {
	echo "[nhp-ac-backstop] $*"
}

have() {
	command -v "$1" >/dev/null 2>&1
}

# Every netfilter call goes through here, for the -w: without it a concurrent
# iptables user makes the call fail on /run/xtables.lock instead of waiting.
# That collision is not hypothetical - nhp-acd shells out to iptables itself in
# FilterMode 0 (nhp/utils/iptables.go, also without -w) and the deploy runs
# `up` while that daemon is still running - and a failed -N/-A/-I here means
# the protected port is open.
ipt() {
	local cmd="$1"
	shift
	"$cmd" -w 5 "$@"
}

require_root() {
	if [ "$(id -u)" -ne 0 ]; then
		log "must run as root"
		exit 1
	fi
}

port_list() {
	echo "$PORTS" | tr ',' ' '
}

# Install the DROP chain and put its jump first in INPUT. It has to be first:
# an ACCEPT rule left over from an earlier iptables-mode baseline would
# otherwise short-circuit it.
#
# Returns non-zero on any failure. The caller must not report success without
# also calling verify_chain: the whole point of this script is a guarantee
# about the kernel state, and "the commands exited 0" is not that guarantee.
install_chain() {
	local cmd="$1"

	if ! have "$cmd"; then
		# Not a no-op to be shrugged off: with no binary for this family
		# there is no backstop for it either. ExecStartPre at boot has no
		# equivalent of the deploy's `dnf install iptables`.
		log "ERROR: $cmd not found - cannot install the backstop for this family (dnf install -y iptables)"
		return 1
	fi

	ipt "$cmd" -n -L "$CHAIN" >/dev/null 2>&1 || ipt "$cmd" -N "$CHAIN" || return 1
	local port
	for port in $(port_list); do
		ipt "$cmd" -C "$CHAIN" -p tcp --dport "$port" -j DROP 2>/dev/null ||
			ipt "$cmd" -A "$CHAIN" -p tcp --dport "$port" -j DROP || return 1
	done

	if [ "$(ipt "$cmd" -S INPUT | sed -n '2p')" != "-A INPUT -j $CHAIN" ]; then
		ipt "$cmd" -I INPUT 1 -j "$CHAIN" || return 1
	fi
	# Drop duplicate jumps a previous run may have left further down the chain.
	local guard=0 idx
	while [ "$(ipt "$cmd" -S INPUT | grep -c -- "-j $CHAIN")" -gt 1 ] && [ "$guard" -lt 20 ]; do
		idx=$(ipt "$cmd" -L INPUT --line-numbers -n | awk -v c="$CHAIN" '$2 == c { print $1 }' | tail -1)
		[ -n "$idx" ] || break
		ipt "$cmd" -D INPUT "$idx" || break
		guard=$((guard + 1))
	done
	return 0
}

# Re-read the rules install_chain claims to have made. Cheap, and it is the
# only thing that actually establishes the fail-closed property for callers
# (ExecStartPre, ExecStopPost, the deploy's fail_closed()).
verify_chain() {
	local cmd="$1"

	have "$cmd" || return 1

	ipt "$cmd" -C INPUT -j "$CHAIN" >/dev/null 2>&1 || return 1
	local port
	for port in $(port_list); do
		ipt "$cmd" -C "$CHAIN" -p tcp --dport "$port" -j DROP >/dev/null 2>&1 || return 1
	done

	# Position is a warning, not a failure: install_chain puts the jump at
	# INPUT 1, but both iptables_default.sh and nhp-acd insert their own
	# rules there (`-I INPUT ...`), so losing the first slot to them is
	# normal and only matters if what got in front ACCEPTs a guarded port.
	if [ "$(ipt "$cmd" -S INPUT | sed -n '2p')" != "-A INPUT -j $CHAIN" ]; then
		log "WARNING: the $cmd jump to $CHAIN is not the first INPUT rule; an earlier ACCEPT could bypass it"
	fi
	return 0
}

remove_chain() {
	local cmd="$1"

	have "$cmd" || return 0

	local guard=0
	while ipt "$cmd" -C INPUT -j "$CHAIN" 2>/dev/null && [ "$guard" -lt 20 ]; do
		ipt "$cmd" -D INPUT -j "$CHAIN" || return 1
		guard=$((guard + 1))
	done
	if ipt "$cmd" -n -L "$CHAIN" >/dev/null 2>&1; then
		ipt "$cmd" -F "$CHAIN" || return 1
		ipt "$cmd" -X "$CHAIN" || return 1
	fi
	# The jump is what enforces the DROP, so that is what `down` has to have
	# got rid of; a leftover empty chain would be harmless but is a bug.
	! ipt "$cmd" -C INPUT -j "$CHAIN" 2>/dev/null
}

default_iface() {
	# Same "default via <gw> dev <iface>" shape that
	# ebpfegine.go:getDefaultRouteInterface() parses, so both agree on which
	# interface carries the XDP program.
	ip route | sed -n 's/^default via [^ ]* dev \([^ ]*\).*/\1/p' | head -1
}

xdp_attached() {
	local iface="$1"

	[ -n "$iface" ] || return 1
	[ -e "$PIN_XDP" ] || return 1
	[ -e "$PIN_TC" ] || return 1
	ip -details link show dev "$iface" 2>/dev/null | grep -qi 'xdp'
}

cmd_up() {
	require_root

	local rc=0 cmd
	for cmd in iptables ip6tables; do
		if ! install_chain "$cmd" || ! verify_chain "$cmd"; then
			log "ERROR: the $cmd backstop is not in place - tcp/$(port_list | tr ' ' ',') may be OPEN for this family"
			rc=1
		fi
	done

	if [ "$rc" -ne 0 ]; then
		log "ERROR: the fail-closed backstop is NOT in place; do not treat the protected ports as closed"
		return 1
	fi

	log "backstop active: tcp/$(port_list | tr ' ' ',') dropped until XDP is attached (IPv6 permanently)"
}

cmd_down() {
	require_root
	# IPv4 only: the XDP program takes over ingress filtering for v4, but it
	# passes all IPv6, so the v6 chain stays.
	if ! remove_chain iptables; then
		log "ERROR: could not lift the IPv4 backstop - tcp/$(port_list | tr ' ' ',') stays closed"
		return 1
	fi
	log "IPv4 backstop lifted; XDP is now the only IPv4 ingress filter"
}

cmd_wait_attach() {
	require_root

	local iface
	iface=$(default_iface)
	if [ -z "$iface" ]; then
		log "cannot determine the default route interface; keeping the backstop"
		return 1
	fi

	local deadline=$(($(date +%s) + TIMEOUT))
	while [ "$(date +%s)" -lt "$deadline" ]; do
		if xdp_attached "$iface"; then
			log "XDP program attached on $iface"
			# A failed `down` leaves the port closed with XDP up, which
			# is safe but broken; fail the unit so the deploy sees it.
			cmd_down || return 1
			return 0
		fi
		sleep 1
	done

	log "timed out after ${TIMEOUT}s waiting for the XDP program on $iface"
	log "keeping the backstop in place - the protected ports stay closed"
	return 1
}

# Remove the iptables_default.sh baseline without touching the backstop.
# A blanket -F/-X would take the backstop chain with it and reopen the
# protected port for the rest of the deploy.
cmd_flush_legacy() {
	require_root

	local cmd chain guard idx rule_no
	for cmd in iptables ip6tables; do
		have "$cmd" || continue

		ipt "$cmd" -P INPUT ACCEPT
		ipt "$cmd" -P FORWARD ACCEPT
		ipt "$cmd" -P OUTPUT ACCEPT

		for chain in INPUT FORWARD OUTPUT; do
			guard=0
			while [ "$guard" -lt 100 ]; do
				# `-S <chain>` prints the policy line first, so the Nth
				# printed rule is rule number N-1. Deleting by number
				# avoids re-quoting rules with --log-prefix "[NHP-...] ".
				idx=$(ipt "$cmd" -S "$chain" | grep -nE "$LEGACY_RULE_RE" | head -1 | cut -d: -f1)
				[ -n "$idx" ] || break
				rule_no=$((idx - 1))
				[ "$rule_no" -ge 1 ] || break
				ipt "$cmd" -D "$chain" "$rule_no" || break
				guard=$((guard + 1))
			done
		done

		if ipt "$cmd" -n -L NHP_DENY >/dev/null 2>&1; then
			ipt "$cmd" -F NHP_DENY
			ipt "$cmd" -X NHP_DENY
		fi
	done

	if command -v ipset >/dev/null 2>&1; then
		local set_name
		for set_name in $LEGACY_SETS; do
			ipset flush "$set_name" 2>/dev/null || true
			ipset destroy "$set_name" 2>/dev/null || true
		done
	fi

	log "legacy iptables/ipset baseline removed (backstop preserved)"
}

cmd_status() {
	local iface
	iface=$(default_iface)
	echo "default route interface: ${iface:-unknown}"
	if xdp_attached "$iface"; then
		echo "XDP: attached"
	else
		echo "XDP: NOT attached"
	fi
	local cmd
	for cmd in iptables ip6tables; do
		if ! have "$cmd"; then
			echo "$cmd: not installed (no backstop possible for this family)"
			continue
		fi
		if verify_chain "$cmd"; then
			echo "$cmd: backstop active"
			ipt "$cmd" -n -L "$CHAIN"
		else
			echo "$cmd: backstop NOT installed"
		fi
	done
}

case "${1:-}" in
up) cmd_up ;;
down) cmd_down ;;
wait-attach) cmd_wait_attach ;;
flush-legacy) cmd_flush_legacy ;;
status) cmd_status ;;
*)
	echo "usage: $0 {up|down|wait-attach|flush-legacy|status}" >&2
	exit 2
	;;
esac
