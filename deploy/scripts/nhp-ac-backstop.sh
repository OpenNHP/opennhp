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
# at all. Nothing takes over v6 filtering after the cutover, so this script
# treats v6 as "stays closed by default":
#   * the v6 backstop chain is *permanent* - `up` installs it and `down`
#     deliberately leaves it in place;
#   * `flush-legacy` clears the v6 ipset/NHP_DENY rules but keeps the
#     ip6tables INPUT/FORWARD DROP policy that iptables_default.sh left
#     behind, after making sure the lo / tcp22 / ESTABLISHED ACCEPTs that
#     keep SSH working are present.
# The v6 backstop is best-effort, though: a host booted with ipv6.disable=1
# has no ip6tables at all, and turning that into a permanent nhp-acd outage
# (ExecStartPre fails -> the daemon never starts) is a far worse trade than
# running with IPv4-only enforcement, which is all the demo VPC has anyway.
# So `up` hard-fails on IPv4 and only warns on IPv6.
#
# Commands:
#   up            install the backstop (idempotent). Verifies the rules are
#                 really in the kernel afterwards and exits non-zero if the
#                 IPv4 one is not there, so ExecStartPre aborts the start
#                 rather than letting the daemon come up behind a backstop
#                 that was never installed. IPv6 failures only warn.
#   down          lift the IPv4 backstop, keep the IPv6 one (idempotent).
#                 Non-zero if the jump is still there afterwards, i.e. the
#                 protected ports are still closed.
#   flush-guard-up   park a DROP for the protected ports in the *raw* table.
#                 The backstop chain above lives in the filter table, so
#                 anything that runs `iptables -F` (release/nhp-ac/iptables_default.sh
#                 opens with exactly that) takes it down together with the
#                 INPUT jump. The raw table is untouched by a filter-table
#                 flush, so a DROP parked there keeps the ports shut across
#                 the seconds iptables_default.sh spends between its flush and
#                 its closing `iptables -P INPUT DROP`. Verifies and exits
#                 non-zero if the rule is not in the kernel.
#   flush-guard-down remove the raw-table guard (idempotent). Non-zero if any
#                 rule survived: a forgotten raw DROP takes the protected
#                 ports down for good, which is fail-closed but broken.
#   wait-attach   wait for the XDP program, then `down`. Non-zero exit on
#                 timeout, which makes systemd fail the unit with the backstop
#                 still in place (fail closed, and the deploy notices).
#   bpffs-prep    make /sys/fs/bpf writable by the service group, so the
#                 unprivileged daemon can pin without ambient
#                 CAP_DAC_OVERRIDE. Verifies and exits non-zero otherwise.
#                 Lives in this script because the unit already runs it as
#                 root from ExecStartPre; see the block above cmd_bpffs_prep.
#   flush-legacy  remove the iptables_default.sh baseline (chains, ipset rules,
#                 IPv4 DROP policies) while preserving the backstop chains and
#                 the IPv6 default-deny. Re-reads the kernel afterwards and
#                 exits non-zero if anything survived: the caller lifts the
#                 backstop right after this, so a half-done flush means
#                 netfilter silently dropping traffic XDP is passing.
#   status        print the current state
#
# Environment:
#   NHP_BACKSTOP_PORTS    tcp ports to guard, space or comma separated
#                         (default: 443 - the only world-reachable protected
#                         port in the demo AC security group)
#   NHP_BACKSTOP_TIMEOUT  wait-attach timeout in seconds (default: 30)
#   NHP_BPFFS_GROUP       group that owns /sys/fs/bpf after bpffs-prep
#                         (default: ec2-user, the demo service user)
#
# Must run as root; the systemd unit lines use the "+" prefix for that.

set -uo pipefail

CHAIN="NHP_BACKSTOP"
PORTS="${NHP_BACKSTOP_PORTS:-443}"
TIMEOUT="${NHP_BACKSTOP_TIMEOUT:-30}"
BPFFS="/sys/fs/bpf"
BPFFS_GROUP="${NHP_BPFFS_GROUP:-ec2-user}"
PIN_XDP="$BPFFS/xdp_white_prog"
PIN_TC="$BPFFS/tc_egress_prog"

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

# True when ip6tables can actually be used. Three separate ways it cannot:
# the binary is missing, the kernel was booted with ipv6.disable=1 (no
# /proc/net/if_inet6, and every ip6tables call then fails with "can't
# initialize ip6tables table filter: Address family not supported by
# protocol"), or ip6_tables is simply not loadable. The -L probe covers the
# last two; checking it up front is what lets callers tell "no IPv6 on this
# host" apart from "the IPv6 backstop failed to install".
ipv6_usable() {
	have ip6tables || return 1
	[ -e /proc/net/if_inet6 ] || return 1
	ipt ip6tables -n -L INPUT >/dev/null 2>&1
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

# --- flush-proof raw-table guard -------------------------------------------
#
# Same DROP, different table. install_chain/verify_chain/remove_chain above
# work in the filter table, which is where iptables_default.sh's opening
# `iptables -F; iptables -X` hits: it deletes the INPUT jump and the
# NHP_BACKSTOP chain, and the default-DROP policy only comes back ~170 lines
# later, after six ipset creates and the whole NHP_DENY/INPUT/FORWARD setup.
# On an eBPF host mid-rollback that gap is an ACCEPT policy, an empty filter
# table and no XDP - i.e. tcp/443 open to the internet. These three functions
# park the DROP in the raw table for the duration; PREROUTING in raw runs
# before conntrack, so it also covers already-established flows.
#
# IPv4 only, on purpose: the filter-table IPv6 chain survives
# `iptables -F` (that flush is IPv4) and stays up until the caller removes it.
install_raw_guard() {
	if ! have iptables; then
		log "ERROR: iptables not found - cannot park the flush-proof guard (dnf install -y iptables)"
		return 1
	fi

	local port
	for port in $(port_list); do
		ipt iptables -t raw -C PREROUTING -p tcp --dport "$port" -j DROP 2>/dev/null ||
			ipt iptables -t raw -I PREROUTING 1 -p tcp --dport "$port" -j DROP || return 1
	done
	return 0
}

verify_raw_guard() {
	have iptables || return 1

	local port
	for port in $(port_list); do
		ipt iptables -t raw -C PREROUTING -p tcp --dport "$port" -j DROP >/dev/null 2>&1 || return 1
	done
	return 0
}

remove_raw_guard() {
	have iptables || return 0

	local port guard
	for port in $(port_list); do
		guard=0
		while ipt iptables -t raw -C PREROUTING -p tcp --dport "$port" -j DROP 2>/dev/null &&
			[ "$guard" -lt 20 ]; do
			ipt iptables -t raw -D PREROUTING -p tcp --dport "$port" -j DROP || return 1
			guard=$((guard + 1))
		done
		# Re-read: a rule still matching here means the port stays dark.
		if ipt iptables -t raw -C PREROUTING -p tcp --dport "$port" -j DROP 2>/dev/null; then
			return 1
		fi
	done
	return 0
}

raw_guard_present() {
	have iptables || return 1

	local port
	for port in $(port_list); do
		ipt iptables -t raw -C PREROUTING -p tcp --dport "$port" -j DROP >/dev/null 2>&1 && return 0
	done
	return 1
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

	# IPv4 is the family the cutover actually hands over to XDP, and the
	# family the security group exposes to the world, so a missing v4
	# backstop is fatal: ExecStartPre aborts the start and the port stays
	# shut behind whatever was enforcing before.
	if ! install_chain iptables || ! verify_chain iptables; then
		log "ERROR: the IPv4 backstop is not in place - tcp/$(port_list | tr ' ' ',') may be OPEN"
		log "ERROR: the fail-closed backstop is NOT in place; do not treat the protected ports as closed"
		return 1
	fi

	# IPv6 is best-effort on purpose - see the ipv6_usable comment. Losing
	# it costs a defense-in-depth layer on a VPC that has no IPv6 at all
	# (terraform/demo/security-groups.tf is cidr_blocks only); making it
	# fatal costs the whole access controller.
	if ! ipv6_usable; then
		log "IPv6 netfilter is unavailable on this host (no ip6tables, or IPv6 disabled in the kernel); skipping the IPv6 backstop"
	elif ! install_chain ip6tables || ! verify_chain ip6tables; then
		log "WARNING: the IPv6 backstop is not in place - tcp/$(port_list | tr ' ' ',') may be reachable over IPv6"
		log "WARNING: continuing anyway; the IPv4 backstop is up and XDP does not filter IPv6 either way"
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

cmd_flush_guard_up() {
	require_root

	if ! install_raw_guard || ! verify_raw_guard; then
		log "ERROR: the raw-table guard is not in place - tcp/$(port_list | tr ' ' ',') would be OPEN while the filter table is flushed"
		return 1
	fi
	log "raw-table guard parked: tcp/$(port_list | tr ' ' ',') dropped in raw/PREROUTING (survives 'iptables -F')"
}

cmd_flush_guard_down() {
	require_root

	if ! remove_raw_guard; then
		log "ERROR: could not remove the raw-table guard - tcp/$(port_list | tr ' ' ',') stays dropped before conntrack"
		return 1
	fi
	log "raw-table guard removed"
}

# ebpfegine.go pins its programs and maps straight into /sys/fs/bpf, which
# systemd's sys-fs-bpf.mount leaves 0700 root:root, and nhp-acd runs as an
# unprivileged user. The obvious way in - ambient CAP_DAC_OVERRIDE - is
# root-equivalent by the same argument that keeps CAP_SYS_ADMIN off the unit:
# it bypasses every DAC file permission check, so a compromised daemon could
# write /etc/cron.d/*, ~root/.ssh/authorized_keys or the unit file itself, and
# NoNewPrivileges= does not mitigate that. Group-owning this one directory is
# a far narrower grant, so the unit does that from a root ("+") ExecStartPre
# instead - ordered after `up`, so a failure here leaves the protected ports
# closed and the daemon simply never starts.
cmd_bpffs_prep() {
	require_root

	if ! grep -q " $BPFFS bpf " /proc/self/mounts; then
		log "ERROR: $BPFFS is not a mounted bpffs - nhp-acd cannot pin (try: systemctl start sys-fs-bpf.mount)"
		return 1
	fi

	if ! getent group "$BPFFS_GROUP" >/dev/null 2>&1; then
		log "ERROR: group $BPFFS_GROUP does not exist (set NHP_BPFFS_GROUP)"
		return 1
	fi

	if ! chgrp "$BPFFS_GROUP" "$BPFFS" 2>/dev/null || ! chmod 0770 "$BPFFS" 2>/dev/null; then
		# bpffs inodes normally accept setattr, but if this kernel refuses
		# it the mount options do the same job. Both paths are checked
		# against the kernel below, so a silent failure here is caught.
		local gid
		gid=$(getent group "$BPFFS_GROUP" | cut -d: -f3)
		log "chgrp/chmod on $BPFFS failed; retrying as a remount (gid=$gid)"
		mount -o "remount,mode=0770,gid=$gid" "$BPFFS" 2>/dev/null || true
	fi

	local mode group
	mode=$(stat -c '%a' "$BPFFS" 2>/dev/null)
	group=$(stat -c '%G' "$BPFFS" 2>/dev/null)
	if [ -z "$mode" ] || [ "$group" != "$BPFFS_GROUP" ] || [ $((0$mode & 070)) -ne $((070)) ]; then
		log "ERROR: $BPFFS is ${mode:-?} ${group:-?}, need group $BPFFS_GROUP with rwx - the daemon cannot pin there without CAP_DAC_OVERRIDE"
		return 1
	fi

	log "$BPFFS prepared: mode $mode, group $group"
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

# Delete every LEGACY_RULE_RE rule from one chain of one family. Non-zero on
# the first failed deletion: a rule we meant to remove but did not is a rule
# that will drop traffic XDP is passing, so the caller must not keep going as
# if the chain were clean.
strip_legacy_rules() {
	local cmd="$1" chain="$2"
	local guard=0 idx rule_no

	while [ "$guard" -lt 100 ]; do
		# `-S <chain>` prints the policy line first, so the Nth printed
		# line is rule number N-1. Deleting by number avoids re-quoting
		# rules with --log-prefix "[NHP-...] ".
		idx=$(ipt "$cmd" -S "$chain" | grep -nE "$LEGACY_RULE_RE" | head -1 | cut -d: -f1)
		[ -n "$idx" ] || return 0
		rule_no=$((idx - 1))
		if [ "$rule_no" -lt 1 ]; then
			# Only reachable if the policy line itself matched, which
			# none of these patterns can do. Bail rather than pass 0
			# to -D.
			log "ERROR: $cmd $chain policy line matched the legacy rule pattern; refusing to guess"
			return 1
		fi
		if ! ipt "$cmd" -D "$chain" "$rule_no"; then
			log "ERROR: could not delete $cmd $chain rule $rule_no (legacy baseline)"
			return 1
		fi
		guard=$((guard + 1))
	done

	log "ERROR: $cmd $chain still has legacy rules after $guard deletions; giving up"
	return 1
}

# Re-read the kernel and confirm flush-legacy actually achieved its contract
# for this family. Same reasoning as verify_chain: "the commands exited 0" is
# not the guarantee the caller needs.
verify_flushed() {
	local cmd="$1" rc=0 chain residual

	for chain in INPUT FORWARD OUTPUT; do
		residual=$(ipt "$cmd" -S "$chain" | grep -E "$LEGACY_RULE_RE" || true)
		if [ -n "$residual" ]; then
			log "ERROR: legacy rules still present in $cmd $chain:"
			echo "$residual"
			rc=1
		fi
	done

	if ipt "$cmd" -n -L NHP_DENY >/dev/null 2>&1; then
		log "ERROR: the $cmd NHP_DENY chain still exists"
		rc=1
	fi

	return $rc
}

# The v6 INPUT policy stays at DROP after the flush (XDP passes every IPv6
# frame, so nothing else would filter v6), which means the ACCEPTs that keep
# this host administrable have to be there. iptables_default.sh leaves them
# behind - the strip only removes ipset/NHP_DENY rules - but do not rely on
# that: locking ourselves out over a rule that was never installed is exactly
# the failure this whole script exists to avoid.
# Appended, not inserted: after the strip the only DROP left in v6 INPUT is the
# backstop jump, which matches the guarded tcp ports only, so position does not
# matter - and -A keeps that jump in INPUT 1 where install_chain wants it.
ensure_v6_admin_rules() {
	local rc=0

	ipt ip6tables -C INPUT -i lo -j ACCEPT >/dev/null 2>&1 ||
		ipt ip6tables -A INPUT -i lo -j ACCEPT || rc=1
	ipt ip6tables -C INPUT -p tcp --dport 22 -j ACCEPT >/dev/null 2>&1 ||
		ipt ip6tables -A INPUT -p tcp --dport 22 -j ACCEPT || rc=1
	ipt ip6tables -C INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT >/dev/null 2>&1 ||
		ipt ip6tables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT || rc=1

	return $rc
}

# Remove the iptables_default.sh baseline without touching the backstop.
# A blanket -F/-X would take the backstop chain with it and reopen the
# protected port for the rest of the deploy.
#
# Non-zero exit means the host is NOT ready for the cutover. The caller lifts
# the IPv4 backstop immediately after this (wait-attach, once XDP attaches), so
# a residual `-A INPUT -j NHP_DENY` or a leftover `-P INPUT DROP` would then
# drop exactly the traffic XDP is busy passing - tcp/443 hard down, with the
# deploy reporting success.
cmd_flush_legacy() {
	require_root

	local rc=0 chain

	# IPv4: XDP takes over, so netfilter has to get out of the way
	# completely - ACCEPT policies and no legacy rules.
	for chain in INPUT FORWARD OUTPUT; do
		strip_legacy_rules iptables "$chain" || rc=1
	done
	if ipt iptables -n -L NHP_DENY >/dev/null 2>&1; then
		ipt iptables -F NHP_DENY || rc=1
		ipt iptables -X NHP_DENY || rc=1
	fi
	# Policies last, and only if the rules really went: an ACCEPT policy in
	# front of rules we failed to remove is the worst of both worlds, while
	# a surviving DROP policy at least stays fail-closed for the abort the
	# non-zero return below triggers.
	if [ "$rc" -eq 0 ]; then
		ipt iptables -P INPUT ACCEPT || rc=1
		ipt iptables -P FORWARD ACCEPT || rc=1
		ipt iptables -P OUTPUT ACCEPT || rc=1
	fi

	# IPv6: same rule cleanup, but the DROP policy deliberately stays. XDP
	# does not look at IPv6 at all (nhp_ebpf_xdp.c XDP_PASSes every
	# ETH_P_IPV6 frame) and the permanent v6 backstop only covers the
	# guarded tcp ports, so dropping the policy here would leave sshd and
	# the NHP UDP listener on :: with no ingress filtering whatsoever.
	# OUTPUT is set ACCEPT because iptables_default.sh sets it ACCEPT too.
	if ! ipv6_usable; then
		log "IPv6 netfilter is unavailable on this host; nothing to flush for IPv6"
	else
		ipt ip6tables -P OUTPUT ACCEPT || rc=1
		for chain in INPUT FORWARD OUTPUT; do
			strip_legacy_rules ip6tables "$chain" || rc=1
		done
		if ipt ip6tables -n -L NHP_DENY >/dev/null 2>&1; then
			ipt ip6tables -F NHP_DENY || rc=1
			ipt ip6tables -X NHP_DENY || rc=1
		fi
		if [ "$(ipt ip6tables -S INPUT | head -1)" = "-P INPUT DROP" ]; then
			ensure_v6_admin_rules || rc=1
		fi
	fi

	if command -v ipset >/dev/null 2>&1; then
		local set_name
		for set_name in $LEGACY_SETS; do
			ipset list "$set_name" >/dev/null 2>&1 || continue
			ipset flush "$set_name" 2>/dev/null || true
			# Best-effort: a set that survives is harmless once nothing
			# references it, and a set that survives *because* something
			# still references it is caught by verify_flushed below.
			if ! ipset destroy "$set_name" 2>/dev/null; then
				log "WARNING: could not destroy ipset $set_name (still referenced?)"
			fi
		done
	fi

	# Nothing above is trusted on its own: re-read the kernel.
	verify_flushed iptables || rc=1
	if [ "$(ipt iptables -S INPUT | head -1)" != "-P INPUT ACCEPT" ]; then
		log "ERROR: the IPv4 INPUT policy is still $(ipt iptables -S INPUT | head -1) - XDP-passed traffic would be dropped by it"
		rc=1
	fi
	if [ "$(ipt iptables -S FORWARD | head -1)" != "-P FORWARD ACCEPT" ]; then
		log "ERROR: the IPv4 FORWARD policy is still $(ipt iptables -S FORWARD | head -1)"
		rc=1
	fi
	if ipv6_usable; then
		verify_flushed ip6tables || rc=1
	fi

	# The flush just set the IPv4 INPUT policy to ACCEPT, so the backstop
	# chain is now the only thing keeping the guarded ports shut until XDP
	# attaches. If the strip took it with it, say so loudly.
	if ! verify_chain iptables; then
		log "ERROR: the IPv4 backstop did not survive the flush - tcp/$(port_list | tr ' ' ',') is OPEN with an ACCEPT policy"
		rc=1
	fi

	if [ "$rc" -ne 0 ]; then
		log "ERROR: the legacy baseline was NOT fully removed; do not cut over to eBPF mode on this host"
		return 1
	fi

	log "legacy iptables/ipset baseline removed (backstop preserved, IPv6 default-deny kept)"
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
		if [ "$cmd" = ip6tables ] && ! ipv6_usable; then
			echo "ip6tables: unusable (missing, or IPv6 disabled in the kernel) - no IPv6 backstop"
			continue
		fi
		if ! have "$cmd"; then
			echo "$cmd: not installed (no backstop possible for this family)"
			continue
		fi
		echo "$cmd INPUT policy: $(ipt "$cmd" -S INPUT | head -1)"
		if verify_chain "$cmd"; then
			echo "$cmd: backstop active"
			ipt "$cmd" -n -L "$CHAIN"
		else
			echo "$cmd: backstop NOT installed"
		fi
	done
	if [ -d "$BPFFS" ]; then
		echo "$BPFFS: $(stat -c '%A %U:%G' "$BPFFS" 2>/dev/null || echo unknown) (want group $BPFFS_GROUP with rwx)"
	fi
	if raw_guard_present; then
		echo "raw-table guard: parked (tcp/$(port_list | tr ' ' ',') dropped in raw/PREROUTING)"
		ipt iptables -t raw -S PREROUTING
	else
		echo "raw-table guard: not parked"
	fi
}

case "${1:-}" in
up) cmd_up ;;
down) cmd_down ;;
wait-attach) cmd_wait_attach ;;
flush-legacy) cmd_flush_legacy ;;
flush-guard-up) cmd_flush_guard_up ;;
flush-guard-down) cmd_flush_guard_down ;;
bpffs-prep) cmd_bpffs_prep ;;
status) cmd_status ;;
*)
	echo "usage: $0 {up|down|wait-attach|flush-legacy|flush-guard-up|flush-guard-down|bpffs-prep|status}" >&2
	exit 2
	;;
esac
