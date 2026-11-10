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
	if !strings.Contains(ingress, "if (!legacy_event_budget_available(class, now))") {
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

	if !strings.Contains(stats, "static __always_inline bool legacy_event_budget_available(__u8 action, __u64 now)") {
		t.Error("the audit ring has no budget of its own")
	}
	if !strings.Contains(stats, "spend_token(slot, NHP_EVENT_BURST, now)") {
		t.Error("the audit ring's budget has no per-class bucket")
	}
	if !strings.Contains(stats, "spend_token(NHP_LEGACY_GLOBAL_SLOT, NHP_EVENT_GLOBAL_BURST, now)") {
		t.Error("the audit ring's budget has no global cap across its classes")
	}

	// The buckets live in slots of their own, so a flood of one channel does
	// not spend the other's budget.
	if !strings.Contains(stats, "#define NHP_LEGACY_RATE_BASE (NHP_EVENT_ACTIONS + 1)") {
		t.Error("the audit ring does not have its own bank of slots in the rate map")
	}
}
