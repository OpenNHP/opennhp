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

// Exactly one packet shape may be admitted before conn_track and the
// whitelists: the DHCP lease reply, 67 -> 68.
//
// Asserting on the span rather than on the spelling of one condition is
// deliberate. The hole this replaces was `sport == bpf_htons(DNS_PORT)`, and a
// test that pattern-matched that one expression would have passed happily on
// `udp->source == bpf_htons(53)`, on `bpf_htons(DNS) == ct_key.sport`, or on a
// pre-swapped literal. Counting the exits out of this region catches a new
// unconditional PASS however it is written.
func TestAcIngressAdmitsOnlyDhcpBeforeTheWhitelists(t *testing.T) {
	src := acProgramSource(t)
	span := acUnconditionalUDPSpan(t, src)

	if n := strings.Count(span, "XDP_PASS"); n != 1 {
		t.Errorf("%d unconditional PASSes before conn_track and the whitelists, want exactly 1 (the DHCP lease reply); "+
			"anything admitted here needs no knock", n)
	}
	for _, drop := range []string{"XDP_DROP", "XDP_ABORTED", "XDP_TX", "XDP_REDIRECT"} {
		if strings.Contains(span, drop) {
			t.Errorf("unexpected %s in the pre-whitelist span; this region is the exception list, not a verdict tree", drop)
		}
	}

	// The one PASS must be guarded by BOTH halves of the pair, in order.
	// Matched loosely (names, not an exact expression) so that reformatting
	// the condition does not fail the build for no reason.
	pair := regexp.MustCompile(`(?s)DHCP_PORT_SERVER.*?DHCP_PORT_CLIENT.*?XDP_PASS`)
	if !pair.MatchString(span) {
		t.Error("the surviving PASS is not guarded by the DHCP_PORT_SERVER -> DHCP_PORT_CLIENT pair; " +
			"matching either port alone is how this block became a bypass")
	}
	if !strings.Contains(span, "ct_key.sport") || !strings.Contains(span, "ct_key.dport") {
		t.Error("the DHCP guard no longer tests both the source and the destination port")
	}

	for _, name := range []string{"DHCP_PORT_SERVER 67", "DHCP_PORT_CLIENT 68"} {
		if !strings.Contains(src, "#define "+name) {
			t.Errorf("missing #define %s", name)
		}
	}
	// The old spelling tested the destination port alone, which let any
	// source reach udp/67 and udp/68 on the AC and on every host behind it.
	if strings.Contains(src, "DHCP_PORT_R") || strings.Contains(src, "DHCP_PORT_O") {
		t.Error("DHCP_PORT_R / DHCP_PORT_O are back; those names belonged to the destination-port-only test this replaced")
	}
}

// No UDP packet may be admitted anywhere in this program on the strength of
// its SOURCE port, the DHCP pair excepted.
//
// It used to be: `sport == bpf_htons(DNS_PORT)` passed any datagram sent from
// udp/53. A source port is picked by whoever sends the packet, so this proved
// nothing about the sender and handed anyone who set `--sport 53` every UDP
// port the AC protects — DNS, QUIC/HTTP3, WireGuard, game servers — with no
// knock at all.
//
// This is the span test's backstop: it catches a source-port rule added
// anywhere in the program, including after the whitelists.
func TestAcIngressNeverAdmitsOnSourcePortAlone(t *testing.T) {
	src := acProgramSource(t)

	if strings.Contains(src, "DNS_PORT") {
		t.Error("nhp_ebpf_xdp.c still mentions DNS_PORT; a UDP source port of 53 must not admit anything")
	}

	// Both operand orders, both spellings of the field, == and !=.
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`(?:\.sport|->source)\s*[!=]=\s*bpf_htons\(\s*([A-Za-z0-9_]+)\s*\)`),
		regexp.MustCompile(`bpf_htons\(\s*([A-Za-z0-9_]+)\s*\)\s*[!=]=\s*\w+(?:\.sport|->source)`),
	} {
		for _, m := range re.FindAllStringSubmatch(src, -1) {
			if m[1] != "DHCP_PORT_SERVER" {
				t.Errorf("source-port test on %s: the AC must not admit UDP on the sender's own choice of source port (only the DHCP 67 -> 68 pair may)", m[1])
			}
		}
	}

	// A pre-byte-swapped literal compared against a source port would slip
	// past the two patterns above, so reject the shape outright.
	if raw := regexp.MustCompile(`(?:\.sport|->source)\s*[!=]=\s*(?:0[xX])?[0-9]`); raw.MatchString(src) {
		t.Error("a source port is compared against a bare literal; write it as bpf_htons(NAME) so this file can tell what is being admitted")
	}
}
