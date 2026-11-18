package ebpf

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cilium/ebpf"
)

// The ports in the map keys are network byte order.
//
// The eBPF programs look up a key whose port field is the one from the
// packet: ct_key.dport = tcp->dest, copied into a __be16 field, and the map
// is keyed on the bytes. So user space has to write the port the same way.
//
// It used to write two of them the other way round, and the effect was not a
// rule that matched nothing: it was a rule that matched the *swapped* port.
// A knock for tcp/443 installed a src_port entry keyed by 47873, which
// admitted the peer to a port nobody asked for and closed the port that was
// asked for. That is GHSA-j2qp-fw3h-rpfh.
//
// These tests pin the byte order of every port in every key, and pin it
// against the C sources rather than against a copy of the C sources: the
// contract test below reads nhp_maps.h and nhp_ebpf_xdp.c, so a change on
// either side of the boundary has to be made on both sides.

// The port the tests below use, and the bytes it has on the wire.
const (
	testPort     = 443
	testPortSwap = 47873 // bswap(443): the port the old encoder admitted
)

// networkPortBytes is the wire form of a port: big endian.
func networkPortBytes(p uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, p)
	return b
}

// TestSrcPortKeyIsNetworkByteOrder pins the src_port key's port field.
//
// This is the one that failed open: the key was written host order, so the
// kernel looked up a key that no knock had installed and admitted the peer
// to the byte-swapped port instead.
func TestSrcPortKeyIsNetworkByteOrder(t *testing.T) {
	key := &srcIPdstPortKey{
		SrcIP:   0x0200000a, // 10.0.0.2, as parseIP hands it over
		DstPort: testPort,
	}
	got := key.ToSpKey()

	if len(got) != 6 {
		t.Fatalf("the src_port key is %d bytes, want 6", len(got))
	}
	want := networkPortBytes(testPort)
	if got[4] != want[0] || got[5] != want[1] {
		t.Errorf("the src_port key carries the port as %02x %02x, want %02x %02x (network order)",
			got[4], got[5], want[0], want[1])
	}
	if got[4] == 0xBB && got[5] == 0x01 {
		t.Error("the src_port key carries the port in host order; the kernel would look up bswap(port)")
	}
}

// TestPortListKeyIsNetworkByteOrder pins the port_list key's port fields,
// and the range the key is allowed to carry.
func TestPortListKeyIsNetworkByteOrder(t *testing.T) {
	key := &portListKey{
		SrcIP:        0x0200000a,
		DstPortStart: AnyPortRangeStart,
		DstPortEnd:   AnyPortRangeEnd,
	}
	got := key.ToPlKey()

	if len(got) != 8 {
		t.Fatalf("the port_list key is %d bytes, want 8", len(got))
	}
	if got[4] != 0x00 || got[5] != 0x00 {
		t.Errorf("the port_list key's min port is %02x %02x, want 00 00", got[4], got[5])
	}
	if got[6] != 0xFF || got[7] != 0xFF {
		t.Errorf("the port_list key's max port is %02x %02x, want ff ff", got[6], got[7])
	}
	if AnyPortRangeStart != 0 || AnyPortRangeEnd != 65535 {
		t.Errorf("the wildcard range is %d-%d, want 0-65535", AnyPortRangeStart, AnyPortRangeEnd)
	}
}

// TestWhitelistAndProtocolPortKeys pins the two keys that were already
// right, so that the fix for the other two cannot quietly change them.
//
// The protocol_port key is pinned byte for byte: it used to be written with
// a byte swap in parsePort and a little-endian write in the encoder, and the
// two cancelled out. The bytes on the wire were correct and have to stay
// exactly the same -- an already-authorised peer must not be locked out by
// a change that is supposed to be a no-op.
func TestWhitelistAndProtocolPortKeys(t *testing.T) {
	wl := &whitelistKey{
		SrcIP:    0x0200000a,
		DstIP:    0x0100000a,
		DstPort:  testPort,
		Protocol: 6,
	}
	got := wl.ToWlKey()
	want := networkPortBytes(testPort)
	if got[8] != want[0] || got[9] != want[1] {
		t.Errorf("the spp key carries the port as %02x %02x, want %02x %02x (network order)",
			got[8], got[9], want[0], want[1])
	}

	pp := &procoPortKey{
		DstPort:  testPort,
		Protocol: 6,
	}
	gotPp := pp.ToPpKey()
	wantPp := []byte{0x01, 0xBB, 0x06}
	for i := range wantPp {
		if gotPp[i] != wantPp[i] {
			t.Fatalf("the protocol_port key is % x, want % x", gotPp, wantPp)
		}
	}

	// The port that reaches the encoder is the one parsePort hands over.
	// It used to hand over a byte-swapped port and the encoder wrote it
	// little endian, and the two cancelled out; now the encoder writes
	// network order and the port arrives in the host order it is written in.
	// The bytes are the same either way, and they have to stay the same.
	parsed, err := parsePort("443")
	if err != nil {
		t.Fatalf("parsePort: %v", err)
	}
	if parsed != testPort {
		t.Errorf("parsePort(\"443\") is %d, want %d: the port is written in network order by the encoder, "+
			"so it has to arrive in host order", parsed, testPort)
	}
	gotParsed := (&procoPortKey{DstPort: parsed, Protocol: 6}).ToPpKey()
	for i := range wantPp {
		if gotParsed[i] != wantPp[i] {
			t.Fatalf("the protocol_port key for a parsed port is % x, want % x", gotParsed, wantPp)
		}
	}
}

// TestEbpfRuleAddRejectsANonWildcardPortRange pins the refusal.
//
// The port_list map's key carries a port range, but the programs do not
// evaluate it: they look up one fixed key, {src_ip, 0, 65535}. An entry
// with any other range is an entry no lookup can find, so writing it would
// report an access that was never granted. The refusal has to happen
// instead, and it has to happen before the entry is written.
func TestEbpfRuleAddRejectsANonWildcardPortRange(t *testing.T) {
	const srcIP = "10.0.0.2"
	err := EbpfRuleAdd(MapTypeSrcPortList, EbpfRuleParams{
		SrcIP:        srcIP,
		DstPortStart: 100,
		DstPortEnd:   200,
	}, 30)
	if err == nil {
		t.Fatal("a port_list entry with the range 100-200 was accepted; the eBPF programs " +
			"only ever look up the wildcard range, so such an entry grants nothing")
	}
	if !strings.Contains(err.Error(), "wildcard port range") {
		t.Errorf("the refusal does not say what is wrong: %v", err)
	}

	// Nothing may have been written. The pin only exists on a host that runs
	// the AC, so this half is skipped where there is none.
	m, err := ebpf.LoadPinnedMap("/sys/fs/bpf/port_list", nil)
	if err != nil {
		t.Skipf("the port_list pin is not available here: %v", err)
	}
	defer m.Close()

	key := (&portListKey{SrcIP: 0x0200000a, DstPortStart: 100, DstPortEnd: 200}).ToPlKey()
	var value whitelistValue
	err = m.Lookup(key, &value)
	if err == nil {
		t.Error("the refused range was written to the port_list map")
	} else if !errors.Is(err, ebpf.ErrKeyNotExist) {
		t.Errorf("the port_list map did not answer for the refused key: %v", err)
	}
}

// ---- the contract with the C side ----

// The contract test.
//
// It reads the two C files the boundary runs through -- the map key layouts
// and the program that looks them up -- and checks three things: the port
// fields of the keys are __be16, the wildcard range the programs probe is
// the one user space writes, and the bytes user space produces are the bytes
// the program looks for.
//
// It deliberately does not parse the control flow. What it pins is the
// contract, and a change to the C side has to be made on the Go side as
// well -- which is the point: the defect this guards against was two sides
// disagreeing about a byte order, with nothing to say so.

// cStructBody returns the body of a C struct definition.
func cStructBody(t *testing.T, src, name string) string {
	t.Helper()
	re := regexp.MustCompile(`(?s)struct\s+` + regexp.QuoteMeta(name) + `\s*\{(.*?)\};`)
	m := re.FindStringSubmatch(src)
	if m == nil {
		t.Fatalf("the C source has no struct %s", name)
	}
	return m[1]
}

// cFieldType returns the declared type of a field in a struct body.
func cFieldType(t *testing.T, body, field string) string {
	t.Helper()
	re := regexp.MustCompile(`(?m)^\s*([A-Za-z_][A-Za-z0-9_]*)\s+` + regexp.QuoteMeta(field) + `\s*;`)
	m := re.FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("the struct has no field %s", field)
	}
	return m[1]
}

// TestMapKeyPortsAreNetworkByteOrderInC pins the C side of the contract.
//
// The port fields of the five map keys are __be16. If one of them ever
// becomes a plain __u16, the program will compare it against a byte-swapped
// port and the user-space encoder has to change with it -- so this test has
// to fail when it happens.
func TestMapKeyPortsAreNetworkByteOrderInC(t *testing.T) {
	mapsPath := filepath.Join("..", "..", "ebpf", "xdp", "nhp_maps.h")
	raw, err := os.ReadFile(mapsPath)
	if err != nil {
		t.Fatalf("read %s: %v", mapsPath, err)
	}
	src := string(raw)

	for _, c := range []struct {
		structName string
		field      string
	}{
		{"whitelist_key", "dst_port"},
		{"src_port_list_key", "dst_port"},
		{"port_list_key", "min_port"},
		{"port_list_key", "max_port"},
		{"protocol_port_key", "dst_port"},
	} {
		body := cStructBody(t, src, c.structName)
		got := cFieldType(t, body, c.field)
		if got != "__be16" {
			t.Errorf("%s.%s is declared %s in %s, want __be16: the programs fill it "+
				"from the packet, so user space has to write the port in network byte order",
				c.structName, c.field, got, mapsPath)
		}
	}
}

// cInitBody returns the body of a C struct initializer.
func cInitBody(t *testing.T, src, name string) string {
	t.Helper()
	re := regexp.MustCompile(`(?s)struct\s+` + regexp.QuoteMeta(name) + `\s+\w+\s*=\s*\{(.*?)\};`)
	m := re.FindStringSubmatch(src)
	if m == nil {
		t.Fatalf("the C source has no %s initializer", name)
	}
	return m[1]
}

// TestWildcardRangeMatchesThePrograms pins the one range the port_list map
// is read with.
//
// The programs look up {src_ip, MIN_PORT, MAX_PORT}; user space writes
// {src_ip, AnyPortRangeStart, AnyPortRangeEnd}. The two have to be the same
// numbers, or the entry is written and never found.
func TestWildcardRangeMatchesThePrograms(t *testing.T) {
	xdpPath := filepath.Join("..", "..", "ebpf", "xdp", "nhp_ebpf_xdp.c")
	defs := cDefines(t, xdpPath)
	minPort := evalCIntExpr(t, defs, defs["MIN_PORT"])
	maxPort := evalCIntExpr(t, defs, defs["MAX_PORT"])
	if minPort != AnyPortRangeStart || maxPort != AnyPortRangeEnd {
		t.Errorf("the programs look up the port_list key with the range %d-%d, user space writes %d-%d",
			minPort, maxPort, AnyPortRangeStart, AnyPortRangeEnd)
	}

	// The egress program writes the same two numbers out, so it is checked
	// here as well -- it is the other half of the contract.
	tcPath := filepath.Join("..", "..", "ebpf", "xdp", "tc_egress.c")
	tcRaw, err := os.ReadFile(tcPath)
	if err != nil {
		t.Fatalf("read %s: %v", tcPath, err)
	}
	block := cInitBody(t, string(tcRaw), "port_list_key")
	for _, c := range []struct {
		field string
		want  string
	}{
		{"min_port", "0"},
		{"max_port", "65535"},
	} {
		re := regexp.MustCompile(`(?m)^\s*\.` + c.field + `\s*=\s*([0-9]+)\s*,`)
		m := re.FindStringSubmatch(block)
		if m == nil {
			t.Fatalf("the egress program does not set .%s of its port_list key", c.field)
		}
		if m[1] != c.want {
			t.Errorf("the egress program looks up the port_list key with .%s = %s, want %s",
				c.field, m[1], c.want)
		}
	}
}

// TestGoKeysMatchTheWireBytes pins the bytes user space writes against the
// bytes the program looks for.
//
// The program copies the destination port out of the packet (tcp->dest,
// udp->dest) and the source address out of the IP header, and uses both as
// the key. So the key user space writes has to be the address bytes and the
// port bytes of the packet, in that order.
func TestGoKeysMatchTheWireBytes(t *testing.T) {
	// 10.0.0.2 and 10.0.0.1, as they appear in an IP header.
	srcIP := []byte{10, 0, 0, 2}
	dstIP := []byte{10, 0, 0, 1}

	src := &srcIPdstPortKey{
		SrcIP:   binary.LittleEndian.Uint32(srcIP),
		DstPort: testPort,
	}
	want := append(append([]byte{}, srcIP...), 0x01, 0xBB)
	got := src.ToSpKey()
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("the src_port key is % x, want % x", got, want)
		}
	}

	wl := &whitelistKey{
		SrcIP:    binary.LittleEndian.Uint32(srcIP),
		DstIP:    binary.LittleEndian.Uint32(dstIP),
		DstPort:  testPort,
		Protocol: 6,
	}
	wantWl := append(append(append([]byte{}, srcIP...), dstIP...), 0x01, 0xBB, 0x06)
	gotWl := wl.ToWlKey()
	for i := range wantWl {
		if gotWl[i] != wantWl[i] {
			t.Fatalf("the spp key is % x, want % x", gotWl, wantWl)
		}
	}

	pl := &portListKey{
		SrcIP:        binary.LittleEndian.Uint32(srcIP),
		DstPortStart: AnyPortRangeStart,
		DstPortEnd:   AnyPortRangeEnd,
	}
	wantPl := append(append([]byte{}, srcIP...), 0x00, 0x00, 0xFF, 0xFF)
	gotPl := pl.ToPlKey()
	for i := range wantPl {
		if gotPl[i] != wantPl[i] {
			t.Fatalf("the port_list key is % x, want % x", gotPl, wantPl)
		}
	}
}

// TestTheSwappedPortIsNotTheOneAdmitted pins the shape of the defect itself.
//
// The old encoder wrote the port host order, so the entry it installed was
// the byte-swapped port -- the one nobody asked for. The key the kernel
// looks up is the port on the wire, so the swapped port is the one that was
// admitted.
func TestTheSwappedPortIsNotTheOneAdmitted(t *testing.T) {
	key := (&srcIPdstPortKey{SrcIP: 0x0200000a, DstPort: testPort}).ToSpKey()
	swapped := (&srcIPdstPortKey{SrcIP: 0x0200000a, DstPort: testPortSwap}).ToSpKey()
	if string(key) == string(swapped) {
		t.Error("the src_port key for 443 is the key for 47873; the rule admits the wrong port")
	}
}
