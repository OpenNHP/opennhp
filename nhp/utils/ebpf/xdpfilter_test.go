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
	if !strings.Contains(text, "ct_key.sport == bpf_htons(DHCP_PORT_R) && ct_key.dport == bpf_htons(DHCP_PORT_O)") {
		t.Error("the AC's filter does not admit the DHCP 67 -> 68 port pair")
	}
	if !strings.Contains(text, "ct_key.sport == bpf_htons(NTP_PORT) && ct_key.dport == bpf_htons(NTP_PORT)") {
		t.Error("the AC's filter does not admit the NTP 123 -> 123 port pair")
	}
	if strings.Contains(text, "bpf_htons(DHCP_PORT_R) ||") || strings.Contains(text, "bpf_htons(DHCP_PORT_O) ||") {
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
