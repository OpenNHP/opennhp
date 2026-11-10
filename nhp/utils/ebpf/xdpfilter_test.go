package ebpf

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The nhp-ac's ingress filter is the only inbound control on its host once
// the security group is opened to the whole internet, so the rules that let
// traffic in have to be as narrow as the nhp-server's.
//
// The rules that were not: a blanket "any datagram *from* port 53, to any
// port, from any address", which with the group open was a standing bypass
// into every UDP listener on the host (chronyd, the SSM agent, nhp-acd's own
// sockets) because a source port is free to choose; a DHCP rule that admitted
// anything addressed to port 67 or 68 rather than the 67 -> 68 pair; an
// unconditional IPv6 pass; and an ICMP echo reply passed from anywhere.
//
// These assertions are about the shape of the program, not about a running
// kernel: the rules live in a decision tree that only the eBPF verifier and
// a live host can exercise, and what has to be prevented here is a rule
// quietly coming back.

// acIngressSource reads the AC's ingress program.
func acIngressSource(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "ebpf", "xdp", "nhp_ebpf_xdp.c"))
	if err != nil {
		t.Fatalf("read nhp_ebpf_xdp.c: %v", err)
	}
	return string(b)
}

// acEgressSource reads the AC's TC egress program, the other half of the pair
// that lets the host's own traffic back in.
func acEgressSource(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "ebpf", "xdp", "tc_egress.c"))
	if err != nil {
		t.Fatalf("read tc_egress.c: %v", err)
	}
	return string(b)
}

// serverIngressSource reads the nhp-server's own ingress program, the filter
// that admits the knock port and the relay's SSH.
func serverIngressSource(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "ebpf", "xdp", "nhp_server_xdp.c"))
	if err != nil {
		t.Fatalf("read nhp_server_xdp.c: %v", err)
	}
	return string(b)
}

// xdpStatsSource reads the header both ingress filters share: the token
// buckets, the counters and the per-port aggregation.
func xdpStatsSource(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "ebpf", "xdp", "nhp_xdp_stats.h"))
	if err != nil {
		t.Fatalf("read nhp_xdp_stats.h: %v", err)
	}
	return string(b)
}

// TestAcIngressFilterHasNoBlanketUdpExceptions pins that the AC's filter
// does not wave through datagrams by their port number alone.
func TestAcIngressFilterHasNoBlanketUdpExceptions(t *testing.T) {
	text := acIngressSource(t)

	// The DNS rule is gone: an answer to one of the host's own queries comes
	// in through the conn_track (the host's own query) or through the
	// kernel's socket table (the host's own client socket), both of which
	// ask the kernel which socket would receive the packet instead of
	// trusting a port number.
	if strings.Contains(text, "DNS_PORT") {
		t.Error("the AC's filter still has a DNS_PORT rule; a datagram from port 53 " +
			"reached every UDP listener on the host, to any port, from any address")
	}

	// DHCP is the 67 -> 68 pair, and NTP is the 123 -> 123 pair. Both are
	// matched on the source port as well as the destination one.
	if !strings.Contains(text, "ct_key.sport == bpf_htons(DHCP_PORT_SERVER) && ct_key.dport == bpf_htons(DHCP_PORT_CLIENT)") {
		t.Error("the AC's filter does not admit the DHCP 67 -> 68 port pair")
	}
	if !strings.Contains(text, "ct_key.sport == bpf_htons(NTP_PORT) && ct_key.dport == bpf_htons(NTP_PORT)") {
		t.Error("the AC's filter does not admit the NTP 123 -> 123 port pair")
	}
	if strings.Contains(text, "bpf_htons(DHCP_PORT_SERVER) ||") || strings.Contains(text, "bpf_htons(DHCP_PORT_CLIENT) ||") {
		t.Error("the AC's filter still admits a datagram by its destination port alone")
	}
}

// TestAcIngressFilterFiltersIPv6 pins that IPv6 is a decision tree and not an
// unconditional pass. The demo VPC has no IPv6 address today, which is
// exactly why the pass was dangerous: the day a v6 CIDR is added, every v6
// listener -- sshd's [::]:22 above all -- would be reachable unfiltered with
// the deploy's "filter attached" check still green.
func TestAcIngressFilterFiltersIPv6(t *testing.T) {
	text := acIngressSource(t)

	if strings.Contains(text, "case ETH_P_IPV6: return XDP_PASS;") {
		t.Error("the AC's filter still passes IPv6 unconditionally")
	}
	if !strings.Contains(text, "case ETH_P_IPV6: return handle_ipv6(") {
		t.Error("the AC's filter does not send IPv6 through handle_ipv6()")
	}
	if !strings.Contains(text, "ACT_AC_DROP_V6_OTHER") {
		t.Error("the AC's filter has no action code for a dropped IPv6 packet")
	}
}

// acFunctionBody returns the body of the C function whose definition starts
// with the given signature.
//
// The AC's program has two decision trees -- one per address family -- and a
// test that has to say something about one of them cannot assert on the whole
// file: the other tree's rules would satisfy it.
func acFunctionBody(t *testing.T, src, signature string) string {
	t.Helper()
	i := strings.Index(src, signature)
	if i < 0 {
		t.Fatalf("cannot locate %q in the AC's ingress program; if the program was "+
			"restructured, re-anchor this test rather than deleting it", signature)
	}
	rest := src[i:]
	// The functions here are top level, so their closing brace is the first
	// one at the left margin; the nested blocks are all indented.
	end := strings.Index(rest, "\n}\n")
	if end < 0 {
		t.Fatalf("cannot find the end of %q in the AC's ingress program", signature)
	}
	return rest[:end]
}

// TestAcV6TreePassesSshLikeTheV4Tree pins that tcp/22 is passed on both
// address families.
//
// The IPv4 tree passes SSH because the security group is what restricts it:
// tcp/22 is carved out of the all-ports range and left open to the relay's
// group only. The IPv6 tree dropped it, and the asymmetry is a lockout: an
// operator who manages the AC over IPv6 loses the host on upgrade, with no
// way in and a filter that says it is doing its job.
//
// The security group is the layer that restricts SSH on both families, so the
// program passes it on both, and a deployment that opens its group's IPv6
// rules is closing that port by hand -- exactly as it does for IPv4.
func TestAcV6TreePassesSshLikeTheV4Tree(t *testing.T) {
	src := acIngressSource(t)

	trees := []struct {
		name string
		body string
	}{
		{"the IPv4 tree", acFunctionBody(t, src, "static __always_inline int xdp_white_prog(")},
		{"the IPv6 tree", acFunctionBody(t, src, "static __always_inline int handle_ipv6(")},
	}

	for _, tree := range trees {
		pass := regexp.MustCompile(`tcp->dest\s*==\s*bpf_htons\(22\)\s*\)\s*\{[^}]*return XDP_PASS;`)
		if !pass.MatchString(tree.body) {
			t.Errorf("%s does not pass tcp/22; an operator who manages the AC over IPv6 is "+
				"locked out, and the security group -- which is what restricts SSH on both "+
				"families -- never sees the packet", tree.name)
		}
	}
}

// TestAcV6FilteringCanBeTurnedOff pins that the IPv6 tree can be turned off,
// and that turning it off is a decision rather than a default.
//
// The tree drops every IPv6 listener the AC was never meant to gate, and it
// has to: the AC's whitelists are IPv4 tables, so "the peer that knocked may
// reach this port" is a policy the program cannot express over IPv6. A host
// that runs an IPv6 service anyway needs a way to say so, and the way is a
// constant the loader writes -- never a second unconditional pass, which
// would be the hole the tree closed.
func TestAcV6FilteringCanBeTurnedOff(t *testing.T) {
	src := acIngressSource(t)

	// The knob lives in .rodata, so the loader can write it before the
	// verifier folds it in, and it defaults to the safe answer.
	if !strings.Contains(src, "volatile const __u8 nhp_ac_filter_ipv6 = 1;") {
		t.Error("the AC's ingress program has no nhp_ac_filter_ipv6 constant defaulting to 1; " +
			"an operator whose host needs its IPv6 listeners has no way to ask for that")
	}

	// And the tree is what consults it.
	v6 := acFunctionBody(t, src, "static __always_inline int handle_ipv6(")
	if !strings.Contains(v6, "if (!nhp_ac_filter_ipv6)") {
		t.Error("the AC's IPv6 tree does not consult nhp_ac_filter_ipv6")
	}
}

// TestAcIngressFilterChecksTheIcmpWhitelistForBothDirections pins that an
// ICMP echo reply is admitted on the same whitelist entry the echo request
// needs. The reply used to be passed unconditionally, from anywhere.
func TestAcIngressFilterChecksTheIcmpWhitelistForBothDirections(t *testing.T) {
	text := acIngressSource(t)

	// The lookup is not inside the echo-request branch any more: both
	// directions share it.
	if strings.Contains(text, "if (icmp->type == ICMP_ECHO) {") {
		t.Error("the ICMP whitelist lookup is still inside the echo-request branch; " +
			"an echo reply from an address that never knocked would be passed")
	}
	if !strings.Contains(text, "bpf_map_lookup_elem(&icmpwhitelist, &icmpkey)") {
		t.Error("the AC's filter no longer looks the ICMP whitelist up")
	}
}

// TestAcIngressFilterSeparatesFragments pins that a fragment other than the
// first is not judged by bytes that are not a header. The bytes at ihl * 4 of
// a non-first fragment are payload, and reading them as a UDP header is how
// a crafted fragment used to be passed by the port-pair exceptions.
func TestAcIngressFilterSeparatesFragments(t *testing.T) {
	text := acIngressSource(t)

	if !strings.Contains(text, "bpf_ntohs(iph->frag_off) & IP_OFFSET") {
		t.Error("the AC's filter does not separate the non-first IPv4 fragments")
	}
	if !strings.Contains(text, "ACT_AC_PASS_IPV4_FRAGMENT") {
		t.Error("the AC's filter has no action code for a non-first fragment")
	}
}

// TestAcIngressFilterAsksTheSocketTable pins that the AC's filter lets the
// host's own traffic back in the way the server's does. The AC's conn_track
// covers the flows the host opens (tc_egress.c records every one of them);
// the socket table is the backstop for the case it cannot cover, a socket
// bound to a fixed port below the ephemeral range.
func TestAcIngressFilterAsksTheSocketTable(t *testing.T) {
	text := acIngressSource(t)

	if !strings.Contains(text, "bpf_sk_lookup_udp(") {
		t.Error("the AC's filter does not ask the kernel's socket table about the host's own flows")
	}
	if !strings.Contains(text, "ACT_AC_PASS_HOST_FLOW") {
		t.Error("the AC's filter has no action code for the host's own flow")
	}
}

// TestAcFilterAnswersTheHostsOwnPings pins that an ICMP echo reply is
// admitted on the record of the request the host itself sent.
//
// An echo reply used to be passed unconditionally, from anywhere. Closing
// that hole also closed the host's own diagnostics with it: the AC could no
// longer get answers to the pings an operator runs on the host, which is also
// how the AC's own reachability is checked.
//
// The replacement is the shape the rest of the filter uses -- the host's own
// traffic is admitted because the host opened the flow. tc_egress.c records
// the echo request, the ingress program admits the reply that answers it, and
// only that reply: the identifier is part of the key, so a reply to another
// request, or from another peer, stays dropped.
//
// The kernel-level counterpart is acxdp_linux_test.go's
// TestAcFilterAnswersTheHostsOwnPings; it needs the object files and
// CAP_BPF, so these assertions are what pins the shape without either.
func TestAcIngressFilterAnswersTheHostsOwnPings(t *testing.T) {
	ingress := acIngressSource(t)
	egress := acEgressSource(t)

	// The egress program records the request: the peer and the identifier
	// the request went out with.
	if !strings.Contains(egress, "bpf_map_update_elem(&icmp_echo, &key, &val, BPF_ANY)") {
		t.Error("the AC's egress program does not record the echo requests the host sends")
	}
	if !strings.Contains(egress, "NHP_ICMP_ECHO_TTL_NS") {
		t.Error("the recorded echo request has no expiry; the record would outlive the request")
	}

	// The ingress program matches the reply against that record, checks the
	// record has not expired, and only then passes it.
	guarded := regexp.MustCompile(`(?s)icmp->type == ICMP_ECHOREPLY\).*?` +
		`bpf_map_lookup_elem\(&icmp_echo, &echokey\).*?` +
		`expire_time >= now.*?` +
		`ACT_AC_PASS_HOST_ECHO.*?` +
		`return XDP_PASS;`)
	if !guarded.MatchString(ingress) {
		t.Error("the AC's filter does not admit the reply to a ping the host sent itself")
	}

	// Nothing may pass the reply before the record has been consulted: the
	// hole this replaces was exactly an unconditional pass.
	branch := strings.Index(ingress, "icmp->type == ICMP_ECHOREPLY)")
	lookup := strings.Index(ingress, "bpf_map_lookup_elem(&icmp_echo, &echokey)")
	if branch < 0 || lookup < 0 || lookup < branch {
		t.Fatal("the AC's filter no longer looks the echo record up")
	}
	if strings.Contains(ingress[branch:lookup], "XDP_PASS") {
		t.Error("the AC's filter passes an echo reply before the record of the request has been consulted")
	}
}

// TestAcIngressFilterAdmitsTheV6ClientProtocols pins that the AC's IPv6
// tree admits the same two client protocols the IPv4 tree admits.
//
// A DHCPv6 lease and a chronyd answer both arrive on a socket the kernel
// cannot tell apart from a listener, so the socket-table lookup finds
// nothing and the answer is dropped. The IPv4 branch admits them by their
// port pair; the IPv6 branch has to do the same, or the day a v6 CIDR is
// added to the VPC the AC loses its DHCPv6 lease -- and a host that loses
// its address loses every port with it, about an hour after a deploy that
// verified clean.
func TestAcIngressFilterAdmitsTheV6ClientProtocols(t *testing.T) {
	text := acIngressSource(t)

	// The v6 branch keeps its own codes: a v6 lease must not be counted as
	// a v4 one.
	for _, tc := range []struct {
		guard  string
		action string
		reason string
	}{
		{`udp->source == bpf_htons\(DHCP6_PORT_SERVER\)`, "ACT_AC_PASS_V6_DHCP_CLIENT", "DHCPv6"},
		{`udp->source == bpf_htons\(NTP_PORT\)`, "ACT_AC_PASS_V6_NTP_CLIENT", "NTPv6"},
	} {
		re := regexp.MustCompile(`(?s)` + tc.guard + `.*?` + tc.action + `.*?return (XDP_\w+);`)
		m := re.FindStringSubmatch(text)
		if m == nil {
			t.Errorf("the AC's filter does not admit the %s port pair under %s", tc.reason, tc.action)
			continue
		}
		if m[1] != "XDP_PASS" {
			t.Errorf("the %s branch returns %s, want XDP_PASS", tc.reason, m[1])
		}
	}

	// Both the ports of a pair, never one of them: a datagram to a single
	// well-known port is how a rule turns into a way in.
	if strings.Contains(text, "bpf_htons(DHCP6_PORT_CLIENT) ||") ||
		strings.Contains(text, "bpf_htons(DHCP6_PORT_SERVER) ||") {
		t.Error("the AC's filter still admits a datagram by its destination port alone")
	}
}

// TestExpiredKnocksAreNotScans pins that the AC's expired-knock class stays
// out of the per-port and per-source tables behind the [NHP-SCAN] summary.
//
// A peer whose knock window closed is a legitimate peer that came back too
// late. It is dropped -- the verdict is unchanged -- but it is not a scan,
// and counting it would put an authorized user's address into the table the
// summary publishes as the top sources. The published site excludes the class
// from its attacker totals for the same reason.
func TestExpiredKnocksAreNotScans(t *testing.T) {
	text := acIngressSource(t)

	matches := regexp.MustCompile(`record_event\(ctx, ACT_AC_DROP_KNOCK_EXPIRED,[^;]*\);`).FindAllString(text, -1)
	if len(matches) == 0 {
		t.Fatal("the AC's filter records no expired-knock event")
	}
	for _, m := range matches {
		if !strings.HasSuffix(m, ", true, false);") {
			t.Errorf("the expired-knock class is recorded as a scan hit: %s", m)
		}
	}
}

// TestLapsedKnocksKeepTheirGrace pins that a peer whose knock window closed
// is still classified as a legitimate peer for a bounded while.
//
// A knock opens a port for a while, and the peer does not always notice that
// the window closed at the same moment the AC did: the last segment of a
// session is retransmitted, a FIN arrives late, a keepalive goes out after
// the AC stopped listening. Only the *first* such packet used to be
// classified as the expired knock it is -- the entry was deleted on the spot
// -- and the rest of them landed in the scan classes, which put an
// authorized user's address, and the country it is in, into the published
// attacker totals.
//
// The grace is bounded, and it applies only to the entries that are scoped to
// the peer *and* to the port:
//
//   - the source-and-destination whitelist (spp), the source-and-port one
//     (src_port) and the source-and-port-range one (port_list) hold their
//     entry for the grace;
//   - the source-and-destination one (sdwhitelist) does not, because it has
//     no port in it: it covers everything the peer sends to the AC, and the
//     peer can be the nhp-server itself, whose channel to the AC depends on
//     it. Holding it would take the AC off the air for the grace;
//   - the port-and-protocol one (protocol_port) does not, because it has no
//     peer in it: it says "anyone may use this port", so holding it would
//     let an attacker hide in the expired-knock class.
//
// After the grace the peer is an unknown source to a closed port again, and
// the packet is judged exactly like one that was never admitted.
func TestLapsedKnocksKeepTheirGrace(t *testing.T) {
	text := acIngressSource(t)
	defs := cDefines(t, filepath.Join("..", "..", "ebpf", "xdp", "nhp_ebpf_xdp.c"))

	// The grace is a number of seconds, and it is bounded: long enough to
	// cover the TCP retransmission ladder, short enough not to be a second
	// window of its own.
	grace := evalCIntExpr(t, defs, "NHP_KNOCK_EXPIRED_GRACE_SEC")
	if grace <= 0 {
		t.Errorf("the expired-knock grace is %d seconds; it must be a positive, finite window", grace)
	}
	if grace > 300 {
		t.Errorf("the expired-knock grace is %d seconds; it must stay short next to the knock windows themselves", grace)
	}

	// The peer-and-port-scoped entries hold their entry for the grace: the
	// grace is asked before the entry is deleted.
	for _, m := range []string{"spp", "src_port", "port_list"} {
		del := "bpf_map_delete_elem(&" + m + ","
		delAt := strings.Index(text, del)
		if delAt < 0 {
			t.Errorf("the %s entry is never deleted", m)
			continue
		}
		branchAt := strings.LastIndex(text[:delAt], "if (expire_time < now) {")
		if branchAt < 0 {
			t.Errorf("the %s entry has no expiry branch", m)
			continue
		}
		if !strings.Contains(text[branchAt:delAt], "knock_within_grace(") {
			t.Errorf("the %s entry is deleted on the first sight of the expiry; the grace is asked after the delete, so a peer's late packets are scans again", m)
		}
	}

	// The two entries that are not scoped to the peer are not held.
	for _, m := range []string{"sdwhitelist", "protocol_port"} {
		del := "bpf_map_delete_elem(&" + m + ","
		delAt := strings.Index(text, del)
		if delAt < 0 {
			t.Errorf("the %s entry is never deleted", m)
			continue
		}
		branchAt := strings.LastIndex(text[:delAt], "if (expire_time < now) {")
		if branchAt < 0 {
			t.Errorf("the %s entry has no expiry branch", m)
			continue
		}
		if strings.Contains(text[branchAt:delAt], "knock_within_grace(") {
			t.Errorf("the %s entry is held for the grace; it is not scoped to the peer, so holding it holds every packet of that peer (or to that port)", m)
		}
	}

	// After the grace the packet is judged like one that was never admitted:
	// the same class, the same accounting, the same audit line.
	for _, m := range []string{"spp", "src_port", "port_list"} {
		del := "bpf_map_delete_elem(&" + m + ","
		delAt := strings.Index(text, del)
		if delAt < 0 {
			continue
		}
		next := strings.Index(text[delAt:], "}")
		if next < 0 {
			continue
		}
		if !strings.Contains(text[delAt:delAt+next], "reject_unadmitted(") {
			t.Errorf("the %s branch does not judge the packet as an unadmitted one after the grace", m)
		}
	}
}

// TestClientProtocolPairsCanBeNarrowedBySource pins that the two UDP pairs
// both filters admit -- the DHCP answer (67 -> 68) and the NTP answer
// (123 -> 123) -- can be narrowed to the addresses the host's own servers
// answer from.
//
// The pair says what the answer looks like, not who sent it, so on a host
// whose group is open to the whole internet the filter admits anything
// wearing the right ports. Only the operator knows which addresses their own
// dhclient and chronyd answer from, so the range has to be configurable.
//
// The default has to stay "any source", and it has to be the *object's* own
// default rather than something the loader writes: a host that says nothing
// keeps the behaviour it had, and a narrowing that is applied wrongly costs
// the host its DHCP lease -- and with it its address.
func TestClientProtocolPairsCanBeNarrowedBySource(t *testing.T) {
	filters := []struct {
		who      string
		text     string
		prefix   string
		dhcpCond string
		ntpCond  string
		dhcpPass string
		ntpPass  string
	}{
		{
			who:      "the AC",
			text:     acIngressSource(t),
			prefix:   "nhp_ac_",
			dhcpCond: "ct_key.sport == bpf_htons(DHCP_PORT_SERVER)",
			ntpCond:  "ct_key.sport == bpf_htons(NTP_PORT)",
			dhcpPass: "ACT_AC_PASS_DHCP_CLIENT",
			ntpPass:  "ACT_AC_PASS_NTP_CLIENT",
		},
		{
			who:      "the server",
			text:     serverIngressSource(t),
			prefix:   "nhp_server_",
			dhcpCond: "udp->source == bpf_htons(DHCP_PORT_SERVER)",
			ntpCond:  "udp->source == bpf_htons(NTP_PORT)",
			dhcpPass: "ACT_DHCP_CLIENT",
			ntpPass:  "ACT_NTP_CLIENT",
		},
	}

	for _, f := range filters {
		t.Run(f.who, func(t *testing.T) {
			// The four .rodata constants, each defaulting to 0 (= any source).
			for _, name := range []string{
				f.prefix + "dhcp_addr", f.prefix + "dhcp_mask",
				f.prefix + "ntp_addr", f.prefix + "ntp_mask",
			} {
				decl := "volatile const __u32 " + name + " = 0;"
				if !strings.Contains(f.text, decl) {
					t.Errorf("%s's filter has no %q constant defaulting to 0; without it the pair cannot be narrowed, and a default other than 0 would narrow it for every host", f.who, name)
				}
			}

			// Each pair consults its range before it admits the packet: the
			// check sits between the port pair and the event that lets the
			// packet through.
			for _, c := range []struct {
				what   string
				cond   string
				addr   string
				mask   string
				accept string
			}{
				{"DHCP", f.dhcpCond, f.prefix + "dhcp_addr", f.prefix + "dhcp_mask", f.dhcpPass},
				{"NTP", f.ntpCond, f.prefix + "ntp_addr", f.prefix + "ntp_mask", f.ntpPass},
			} {
				condAt := strings.Index(f.text, c.cond)
				if condAt < 0 {
					t.Errorf("the %s pair is not in %s's filter at all", c.what, f.who)
					continue
				}
				check := "src_in_range(iph->saddr, " + c.addr + ", " + c.mask + ")"
				checkAt := strings.Index(f.text[condAt:], check)
				if checkAt < 0 {
					t.Errorf("the %s pair is not narrowed by source in %s's filter; the port pair alone admits it", c.what, f.who)
					continue
				}
				checkAt += condAt
				acceptAt := strings.Index(f.text[checkAt:], c.accept)
				if acceptAt < 0 {
					t.Errorf("the %s pair admits nothing in %s's filter", c.what, f.who)
					continue
				}
				acceptAt += checkAt
				if checkAt > acceptAt {
					t.Errorf("the %s pair is admitted before its source range is checked in %s's filter", c.what, f.who)
				}
			}

			// A range that was not configured narrows nothing: src_in_range()
			// says yes to every source when the mask is zero.
			start := strings.Index(f.text, "static __always_inline bool src_in_range(")
			if start < 0 {
				t.Fatalf("%s's filter has no src_in_range() helper", f.who)
			}
			end := strings.Index(f.text[start:], "\n}")
			if end < 0 {
				t.Fatal("src_in_range() has no body")
			}
			body := f.text[start : start+end]
			if !strings.Contains(body, "if (mask == 0)") || !strings.Contains(body, "return true;") {
				t.Errorf("src_in_range() in %s's filter does not treat an unset range as \"any source\"; a host that configures nothing would lose its DHCP lease and its clock", f.who)
			}
		})
	}
}

// TestAcAuditRingIsBehindABudget pins that the AC's legacy audit ring -- the
// one that writes nhp_deny-*.log -- is no longer fed one line per packet.
//
// The ring is the AC's oldest reporting path and it predates every bound the
// telemetry ring has. That was tolerable while the security group let in only
// tcp/443 and the knock port, because almost nothing reached the filter to be
// logged. It is not tolerable with the group open to the whole internet: every
// packet that arrives is then chosen by whoever sent it, and one sender is
// enough to fill the disk the daemon runs on -- a denial of service delivered
// through the defense.
//
// The ring now sits behind a budget of its own in the kernel, and the files
// behind it carry a daily byte budget and are swept (readAcEvents and
// acEventLogGroups in nhp/utils/ebpf/engine_linux.go).
func TestAcAuditRingIsBehindABudget(t *testing.T) {
	ingress := acIngressSource(t)

	// The ring's producer spends a budget before it emits.
	if !strings.Contains(ingress, "if (!legacy_event_budget_available(class, verdict, now))") {
		t.Error("the AC's audit ring is not behind a budget; every dropped packet still " +
			"reaches nhp_deny-*.log, one line each")
	}

	// The budget is keyed on the packet's class, not on the verdict, so a
	// flood of one kind cannot starve the others: a knock that was admitted
	// and a knock that expired are both lines in the same file.
	if !strings.Contains(ingress, "submit_event(ctx, 1, ACT_AC_PASS_KNOCK_ADMITTED,") {
		t.Error("the audit ring's accept path does not key its budget on the packet's class")
	}
	if !strings.Contains(ingress, "submit_event(ctx, 0, drop_class,") {
		t.Error("the audit ring's drop path does not key its budget on the packet's class")
	}
}

// TestLegacyAuditBudgetHasBothBuckets pins the shape of the audit ring's
// budget: a per-class bucket and a global cap across all of them.
//
// A per-class budget alone is not a bound on the log, and that was the first
// version's mistake on the telemetry side: a sender picks the class, and
// spreading the flows across the classes and the RSS queues fills every
// bucket. The global cap is what makes the bound hold.
func TestLegacyAuditBudgetHasBothBuckets(t *testing.T) {
	stats := xdpStatsSource(t)

	if !strings.Contains(stats, "static __always_inline bool legacy_event_budget_available(__u8 action, __u8 verdict, __u64 now)") {
		t.Error("the audit ring has no budget of its own")
	}
	if !strings.Contains(stats, "spend_token(slot, NHP_EVENT_BURST, now)") {
		t.Error("the audit ring's budget has no per-class bucket")
	}
	if !strings.Contains(stats, "spend_token(global, NHP_EVENT_GLOBAL_BURST, now)") {
		t.Error("the audit ring's budget has no global cap across its classes")
	}

	// The buckets live in slots of their own, so a flood of one channel does
	// not spend the other's budget.
	if !strings.Contains(stats, "#define NHP_LEGACY_RATE_BASE (NHP_EVENT_ACTIONS + 1)") {
		t.Error("the audit ring does not have its own bank of slots in the rate map")
	}
}

// TestLegacyAuditAcceptChannelHasItsOwnGlobalCap pins that the audit ring's
// accept channel has a global cap of its own.
//
// The two channels answer two different questions: what did the filter drop,
// and who did it let in. The second is the one an incident review reads. With
// one shared cap, a scan of a few packets a second spent the window's
// allowance on drops before any knock was admitted, and nhp_accept-*.log
// stayed empty for the length of the attack -- the audit trail of who was
// admitted thinned out by whoever was attacking.
//
// The per-class buckets were always separate, so the accept class was never
// starved; it is the global cap that used to be shared, and that is what is
// split here.
func TestLegacyAuditAcceptChannelHasItsOwnGlobalCap(t *testing.T) {
	stats := xdpStatsSource(t)
	ingress := acIngressSource(t)

	if !strings.Contains(stats, "#define NHP_LEGACY_ACCEPT_GLOBAL_SLOT") {
		t.Error("the audit ring has no global slot of its own for the accept channel")
	}
	if !strings.Contains(stats, "#define NHP_EVENT_RATE_SLOTS (NHP_LEGACY_ACCEPT_GLOBAL_SLOT + 1)") {
		t.Error("the rate map is not sized for the accept channel's global slot")
	}

	// The verdict picks the cap: an accept spends the accept channel's, a
	// deny spends the deny channel's.
	sel := regexp.MustCompile(`verdict\s*==\s*1\s*\)\s*\?\s*NHP_LEGACY_ACCEPT_GLOBAL_SLOT\s*:\s*NHP_LEGACY_GLOBAL_SLOT`)
	if !sel.MatchString(stats) {
		t.Error("the audit ring's global cap is not chosen by the verdict; a drop flood would silence the accepts")
	}

	// And the verdict reaches the budget: submit_event has to pass it.
	if !strings.Contains(ingress, "legacy_event_budget_available(class, verdict, now)") {
		t.Error("submit_event does not pass the verdict to the audit ring's budget")
	}
}

// loaderSource reads the shared loader, the code that writes the programs'
// .rodata and decides whether a filter is attached at all.
func loaderSource(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("engine_linux.go")
	if err != nil {
		t.Fatalf("read engine_linux.go: %v", err)
	}
	return string(b)
}

// goFunctionBody returns the body of the Go function with the given name: the
// text between its opening brace and the brace that closes it.
func goFunctionBody(t *testing.T, src, name string) string {
	t.Helper()
	start := strings.Index(src, "func "+name+"(")
	if start < 0 {
		t.Fatalf("%s is not in the loader", name)
	}
	open := strings.Index(src[start:], "{")
	if open < 0 {
		t.Fatalf("%s has no body", name)
	}
	depth := 0
	for i := start + open; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return src[start+open+1 : i]
			}
		}
	}
	t.Fatalf("%s has no closing brace", name)
	return ""
}

// TestServerSourceNetsNeverFailsTheLoad pins the failure direction of the two
// source narrowings on the nhp-server.
//
// With server_open_all_ports = true the filter is the only inbound control on
// every port but tcp/22, so every new reason to refuse attaching is a way to
// leave all ports open -- the invariant the server's other refusals are
// limited by. The DHCP and NTP narrowings are defence in depth (nothing
// downstream trusts those packets: the lease is matched against the request,
// chrony checks the server it asked), so an unusable value must not cost the
// host its filter. It is logged at Critical, the pair keeps the object's own
// default (any source), and the filter attaches.
//
// The AC's half is the opposite on purpose, and pinning it here is what keeps
// the asymmetry from being tidied away: the AC's failure direction is closed
// -- it stops, and the fail-closed netfilter backstop holds the protected
// resource shut -- so a stale object is a refusal there.
func TestServerSourceNetsNeverFailsTheLoad(t *testing.T) {
	src := loaderSource(t)

	server := goFunctionBody(t, src, "setServerSourceNets")
	for _, fatal := range []string{"return fmt.Errorf", "return err"} {
		if strings.Contains(server, fatal) {
			t.Errorf("setServerSourceNets contains %q; the server's source narrowings are defence in depth and must never fail the load", fatal)
		}
	}
	if !strings.Contains(server, "log.Critical") {
		t.Error("setServerSourceNets does not log at Critical when a narrowing cannot be applied")
	}
	if !strings.Contains(server, "narrowed = false") {
		t.Error("setServerSourceNets does not leave the pair unnarrowed when a constant cannot be written")
	}

	ac := goFunctionBody(t, src, "setAcSourceNets")
	if !strings.Contains(ac, "return fmt.Errorf") {
		t.Error("setAcSourceNets no longer refuses a stale object; the AC's failure direction is closed, so it must keep refusing")
	}

	// And the call site: the narrowings are written last and their result is
	// not propagated, so nothing about them can end the load.
	constants := goFunctionBody(t, src, "setServerConstants")
	if strings.Contains(constants, "return setServerSourceNets") {
		t.Error("setServerConstants propagates the source-net result; the narrowings must not be able to fail the load")
	}
	if !strings.Contains(constants, "setServerSourceNets(spec, params)") {
		t.Error("setServerConstants no longer writes the source ranges into the object's .rodata")
	}
}
