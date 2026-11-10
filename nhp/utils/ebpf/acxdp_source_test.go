package ebpf

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// acProgramSource returns nhp_ebpf_xdp.c, the AC's ingress program.
//
// These are source-shape tests rather than kernel ones on purpose: loading
// the program needs CAP_BPF + CAP_NET_ADMIN and a compiled object, so the
// functional tests next door skip on every machine that has neither. The
// property guarded here is a single `if` that is easy to widen back by
// accident and whose cost is the whole filter, so it is asserted somewhere
// that always runs.
func acProgramSource(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("..", "..", "ebpf", "xdp", "nhp_ebpf_xdp.c"))
	if err != nil {
		t.Fatalf("read nhp_ebpf_xdp.c: %v", err)
	}
	return string(src)
}

// acProgramCode returns the AC's ingress program with its comments removed.
//
// The comments in this program discuss the port rules at length -- including
// the ones that used to be there, spelled out as `sport == 53` and
// `dport == 67 || dport == 68` -- and a test that scans the text has to be
// able to tell a rule from the sentence describing it.
func acProgramCode(t *testing.T) string {
	t.Helper()
	return stripCComments(acProgramSource(t))
}

// stripCComments removes C comments from the source, keeping the newlines so
// that a failure can still name a line.
func stripCComments(src string) string {
	var b strings.Builder
	b.Grow(len(src))

	for i := 0; i < len(src); {
		switch {
		case strings.HasPrefix(src[i:], "/*"):
			end := strings.Index(src[i+2:], "*/")
			if end < 0 {
				b.WriteString(strings.Repeat("\n", strings.Count(src[i:], "\n")))
				return b.String()
			}
			chunk := src[i : i+2+end+2]
			b.WriteString(strings.Repeat("\n", strings.Count(chunk, "\n")))
			i += len(chunk)
		case strings.HasPrefix(src[i:], "//"):
			end := strings.IndexByte(src[i:], '\n')
			if end < 0 {
				return b.String()
			}
			i += end // the newline itself is kept by the next iteration
		default:
			b.WriteByte(src[i])
			i++
		}
	}

	return b.String()
}

// acUnconditionalUDPSpan returns the part of xdp_white_prog() between the end
// of L4 header parsing and the first conn_track work (`bpf_ktime_get_ns`).
//
// Everything in that span runs before conn_track and before all five
// whitelists, which is what made the bug this file guards so expensive: a
// verdict reached here needs no knock. The span deliberately starts after the
// TCP branch, so the separate (and still open) tcp/22 exception inside the
// parse block is not this test's subject.
func acUnconditionalUDPSpan(t *testing.T, src string) string {
	t.Helper()
	const (
		start = "ct_key.dport = udp->dest;"
		end   = "__u64 now = bpf_ktime_get_ns();"
	)
	i := strings.Index(src, start)
	if i < 0 {
		t.Fatalf("cannot locate %q in nhp_ebpf_xdp.c; if the program was restructured, "+
			"re-anchor this test rather than deleting it", start)
	}
	i += len(start)
	// Searched from the start anchor: `bpf_ktime_get_ns` also appears in the
	// conn_track helpers defined above xdp_white_prog().
	rel := strings.Index(src[i:], end)
	if rel < 0 {
		t.Fatalf("cannot locate %q after %q in nhp_ebpf_xdp.c; if the program was restructured, "+
			"re-anchor this test rather than deleting it", end, start)
	}
	return src[i : i+rel]
}

// portCompareRe matches a comparison of a port field against a named
// constant: `ct_key.sport == bpf_htons(DHCP_PORT_SERVER)` and friends.
//
// The struct the field is read from is captured too, so that the TCP rules
// can be told apart from the UDP ones.
var portCompareRe = regexp.MustCompile(
	`([A-Za-z_][A-Za-z0-9_]*)\s*(?:\.|->)\s*(source|dest|sport|dport)\s*(==|!=)\s*bpf_htons\(\s*([A-Za-z_][A-Za-z0-9_]*)\s*\)`)

// udpPortPairRes match a condition that pins BOTH ends of a UDP flow: the
// source port and the destination port, each against a named constant. Both
// orders are matched, because a rule that names the ports the other way
// round is the same rule.
//
// This is the list of comparisons that have a partner. A comparison that is
// not inside one of these spans stands alone, and a port that stands alone is
// the whole bug this file is about.
var udpPortPairRes = []*regexp.Regexp{
	regexp.MustCompile(`(?:ct_key\.sport|udp->source)\s*==\s*bpf_htons\(\s*[A-Za-z_][A-Za-z0-9_]*\s*\)` +
		`\s*&&\s*` +
		`(?:ct_key\.dport|udp->dest)\s*==\s*bpf_htons\(\s*[A-Za-z_][A-Za-z0-9_]*\s*\)`),
	regexp.MustCompile(`(?:ct_key\.dport|udp->dest)\s*==\s*bpf_htons\(\s*[A-Za-z_][A-Za-z0-9_]*\s*\)` +
		`\s*&&\s*` +
		`(?:ct_key\.sport|udp->source)\s*==\s*bpf_htons\(\s*[A-Za-z_][A-Za-z0-9_]*\s*\)`),
}

// portRuleSide says which end of a flow a port field belongs to.
func portRuleSide(field string) string {
	switch field {
	case "sport", "source":
		return "source"
	case "dport", "dest":
		return "dest"
	}
	return ""
}

// withinAny reports whether the byte offset idx falls inside one of the
// matched spans.
func withinAny(spans [][]int, idx int) bool {
	for _, s := range spans {
		if idx >= s[0] && idx < s[1] {
			return true
		}
	}
	return false
}

// pairedPortComparisons returns the source and destination comparisons that
// are one half of a pinned pair, in the order they appear in the program.
func pairedPortComparisons(code string) (src, dst [][]int) {
	for _, re := range udpPortPairRes {
		for _, m := range re.FindAllStringIndex(code, -1) {
			for _, c := range portCompareRe.FindAllStringSubmatchIndex(code[m[0]:m[1]], -1) {
				switch portRuleSide(code[m[0]+c[4] : m[0]+c[5]]) {
				case "source":
					src = append(src, []int{m[0] + c[0], m[0] + c[1]})
				case "dest":
					dst = append(dst, []int{m[0] + c[0], m[0] + c[1]})
				}
			}
		}
	}
	return src, dst
}

// Exactly two packet shapes may be admitted before conn_track and the
// whitelists: the DHCP lease reply, 67 -> 68, and the NTP answer, 123 -> 123.
//
// Asserting on the span rather than on the spelling of one condition is
// deliberate. The hole this replaces was `sport == bpf_htons(DNS_PORT)`, and a
// test that pattern-matched that one expression would have passed happily on
// `udp->source == bpf_htons(53)`, on `bpf_htons(DNS) == ct_key.sport`, or on a
// pre-swapped literal. Counting the exits out of this region catches a new
// unconditional PASS however it is written.
//
// The two pairs are the two the nhp-server's own filter admits, in the same
// four spellings: DHCP 67 -> 68 and NTP 123 -> 123 in IPv4, DHCPv6 547 -> 546
// and NTPv6 123 -> 123 in IPv6. The IPv6 spellings live in handle_ipv6(),
// which has no conn_track to admit a lease or an answer from -- tc_egress.c
// is IPv4 only -- so they cannot be left to the reverse branch.
func TestAcIngressAdmitsOnlyThePortPairsBeforeTheWhitelists(t *testing.T) {
	code := acProgramCode(t)
	span := acUnconditionalUDPSpan(t, code)

	if n := strings.Count(span, "XDP_PASS"); n != 2 {
		t.Errorf("%d unconditional PASSes before conn_track and the whitelists, want exactly 2 "+
			"(the DHCP lease reply and the NTP answer); anything admitted here needs no knock", n)
	}
	for _, drop := range []string{"XDP_DROP", "XDP_ABORTED", "XDP_TX", "XDP_REDIRECT"} {
		if strings.Contains(span, drop) {
			t.Errorf("unexpected %s in the pre-whitelist span; this region is the exception list, not a verdict tree", drop)
		}
	}

	// Each PASS must be guarded by BOTH halves of its pair, in order.
	// Matched loosely (names, not an exact expression) so that reformatting
	// the condition does not fail the build for no reason.
	for _, tc := range []struct {
		re   *regexp.Regexp
		what string
	}{
		{regexp.MustCompile(`(?s)DHCP_PORT_SERVER.*?DHCP_PORT_CLIENT.*?XDP_PASS`), "the DHCP 67 -> 68 lease reply"},
		{regexp.MustCompile(`(?s)NTP_PORT.*?NTP_PORT.*?XDP_PASS`), "the NTP 123 -> 123 answer"},
	} {
		if !tc.re.MatchString(span) {
			t.Errorf("the pre-whitelist span does not admit %s guarded by both halves of its port pair", tc.what)
		}
	}
	if !strings.Contains(span, "ct_key.sport") || !strings.Contains(span, "ct_key.dport") {
		t.Error("the port-pair guards no longer test both the source and the destination port")
	}

	for _, name := range []string{"DHCP_PORT_SERVER 67", "DHCP_PORT_CLIENT 68", "NTP_PORT 123"} {
		if !strings.Contains(code, "#define "+name) {
			t.Errorf("missing #define %s", name)
		}
	}
	// The old spelling tested the destination port alone, which let any
	// source reach udp/67 and udp/68 on the AC and on every host behind it.
	if strings.Contains(acProgramSource(t), "DHCP_PORT_R") || strings.Contains(acProgramSource(t), "DHCP_PORT_O") {
		t.Error("DHCP_PORT_R / DHCP_PORT_O are back; those names belonged to the destination-port-only test this replaced")
	}
}

// No UDP packet may be admitted anywhere in this program on the strength of
// its SOURCE port alone.
//
// It used to be: `sport == bpf_htons(DNS_PORT)` passed any datagram sent from
// udp/53. A source port is picked by whoever sends the packet, so this proved
// nothing about the sender and handed anyone who set `--sport 53` every UDP
// port the AC protects — DNS, QUIC/HTTP3, WireGuard, game servers — with no
// knock at all.
//
// A source port may appear, but only as one half of a pinned pair: the source
// port AND the destination port, both against named constants. That is what
// the DHCP and NTP rules do, and it is what the nhp-server's filter does.
func TestAcIngressNeverAdmitsOnSourcePortAlone(t *testing.T) {
	code := acProgramCode(t)

	if strings.Contains(code, "DNS_PORT") {
		t.Error("nhp_ebpf_xdp.c still mentions DNS_PORT; a UDP source port of 53 must not admit anything")
	}

	paired, _ := pairedPortComparisons(code)

	for _, m := range portCompareRe.FindAllStringSubmatchIndex(code, -1) {
		structName, field, op, constName := code[m[2]:m[3]], code[m[4]:m[5]], code[m[6]:m[7]], code[m[8]:m[9]]

		// The TCP rules are not this test's subject: the ssh exception is a
		// destination port that then has to clear the address whitelist, and
		// the egress program is what admits the answers.
		if structName == "tcp" {
			continue
		}
		if op != "==" {
			// A "!=" test rejects; it admits nothing.
			continue
		}
		if portRuleSide(field) != "source" {
			continue
		}
		if !withinAny(paired, m[0]) {
			t.Errorf("source-port test on %s: the AC must not admit UDP on the sender's own choice of "+
				"source port (a source port may only be compared as one half of a pinned pair)", constName)
		}
	}

	// A pre-byte-swapped literal compared against a source port would slip
	// past the two patterns above, so reject the shape outright.
	if raw := regexp.MustCompile(`(?:\.sport|->source)\s*[!=]=\s*(?:0[xX])?[0-9]`); raw.MatchString(code) {
		t.Error("a source port is compared against a bare literal; write it as bpf_htons(NAME) so this file can tell what is being admitted")
	}
}

// No UDP packet may be admitted anywhere in this program on the strength of
// its DESTINATION port alone either.
//
// That was the other half of the hole: `dport == 67 || dport == 68` reached
// udp/67 and udp/68 on the AC and on every host behind it, from any source,
// with no knock. Matching either end of the pair on its own turns the rule
// into a way in, so both ends have to be pinned together.
func TestAcIngressNeverAdmitsOnDestinationPortAlone(t *testing.T) {
	code := acProgramCode(t)

	_, paired := pairedPortComparisons(code)

	for _, m := range portCompareRe.FindAllStringSubmatchIndex(code, -1) {
		structName, field, op, constName := code[m[2]:m[3]], code[m[4]:m[5]], code[m[6]:m[7]], code[m[8]:m[9]]

		if structName == "tcp" {
			continue
		}
		if op != "==" {
			continue
		}
		if portRuleSide(field) != "dest" {
			continue
		}
		if !withinAny(paired, m[0]) {
			t.Errorf("destination-port test on %s: the AC must not admit UDP by the destination port "+
				"alone (a destination port may only be compared as one half of a pinned pair)", constName)
		}
	}

	if raw := regexp.MustCompile(`(?:\.dport|->dest)\s*[!=]=\s*(?:0[xX])?[0-9]`); raw.MatchString(code) {
		t.Error("a destination port is compared against a bare literal; write it as bpf_htons(NAME) so this file can tell what is being admitted")
	}
}
