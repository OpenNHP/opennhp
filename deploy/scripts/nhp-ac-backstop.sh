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
#   up            install the backstop (idempotent)
#   down          lift the IPv4 backstop, keep the IPv6 one (idempotent)
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
install_chain() {
	local cmd="$1"

	command -v "$cmd" >/dev/null 2>&1 || return 0

	"$cmd" -n -L "$CHAIN" >/dev/null 2>&1 || "$cmd" -N "$CHAIN"
	local port
	for port in $(port_list); do
		"$cmd" -C "$CHAIN" -p tcp --dport "$port" -j DROP 2>/dev/null ||
			"$cmd" -A "$CHAIN" -p tcp --dport "$port" -j DROP
	done

	if [ "$("$cmd" -S INPUT | sed -n '2p')" != "-A INPUT -j $CHAIN" ]; then
		"$cmd" -I INPUT 1 -j "$CHAIN"
	fi
	# Drop duplicate jumps a previous run may have left further down the chain.
	local guard=0 idx
	while [ "$("$cmd" -S INPUT | grep -c -- "-j $CHAIN")" -gt 1 ] && [ "$guard" -lt 20 ]; do
		idx=$("$cmd" -L INPUT --line-numbers -n | awk -v c="$CHAIN" '$2 == c { print $1 }' | tail -1)
		[ -n "$idx" ] || break
		"$cmd" -D INPUT "$idx"
		guard=$((guard + 1))
	done
}

remove_chain() {
	local cmd="$1"

	command -v "$cmd" >/dev/null 2>&1 || return 0

	local guard=0
	while "$cmd" -C INPUT -j "$CHAIN" 2>/dev/null && [ "$guard" -lt 20 ]; do
		"$cmd" -D INPUT -j "$CHAIN"
		guard=$((guard + 1))
	done
	if "$cmd" -n -L "$CHAIN" >/dev/null 2>&1; then
		"$cmd" -F "$CHAIN"
		"$cmd" -X "$CHAIN"
	fi
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
	install_chain iptables
	install_chain ip6tables
	log "backstop active: tcp/$(port_list | tr ' ' ',') dropped until XDP is attached (IPv6 permanently)"
}

cmd_down() {
	require_root
	# IPv4 only: the XDP program takes over ingress filtering for v4, but it
	# passes all IPv6, so the v6 chain stays.
	remove_chain iptables
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
			cmd_down
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
		command -v "$cmd" >/dev/null 2>&1 || continue

		"$cmd" -P INPUT ACCEPT
		"$cmd" -P FORWARD ACCEPT
		"$cmd" -P OUTPUT ACCEPT

		for chain in INPUT FORWARD OUTPUT; do
			guard=0
			while [ "$guard" -lt 100 ]; do
				# `-S <chain>` prints the policy line first, so the Nth
				# printed rule is rule number N-1. Deleting by number
				# avoids re-quoting rules with --log-prefix "[NHP-...] ".
				idx=$("$cmd" -S "$chain" | grep -nE "$LEGACY_RULE_RE" | head -1 | cut -d: -f1)
				[ -n "$idx" ] || break
				rule_no=$((idx - 1))
				[ "$rule_no" -ge 1 ] || break
				"$cmd" -D "$chain" "$rule_no" || break
				guard=$((guard + 1))
			done
		done

		if "$cmd" -n -L NHP_DENY >/dev/null 2>&1; then
			"$cmd" -F NHP_DENY
			"$cmd" -X NHP_DENY
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
		command -v "$cmd" >/dev/null 2>&1 || continue
		if "$cmd" -C INPUT -j "$CHAIN" 2>/dev/null; then
			echo "$cmd: backstop active"
			"$cmd" -n -L "$CHAIN"
		else
			echo "$cmd: backstop not installed"
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
