//go:build linux

package ebpf

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/cilium/ebpf"
)

// The nhp-ac's own traffic has to get back in.
//
// The AC's filter answers a packet either because a knock admitted it or
// because the host itself opened the flow. The second case is what the
// conn_track is for: tc_egress.c records every flow the host opens, and the
// reverse branch of the ingress program serves the reply out of that entry.
// ICMP has no port to key a conn_track entry on and no handshake to bind one
// to, so the echo requests the host sends are recorded in their own map
// instead -- and the reply is admitted on that record.
//
// These tests exercise the two programs against the kernel, with the real
// object files, and skip when the objects are not built or the process has
// no privilege to load them.

// The verdicts the kernel returns, spelled out because x/sys/unix does not
// carry them.
const (
	xdpDrop = 1
	xdpPass = 2
	tcActOK = 0
)

// acObjPath is where `make ebpf-objs` puts the AC's objects.
func acObjPath(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "release", "nhp-ac", "etc", name)
	if _, err := os.Stat(path); err != nil {
		t.Skipf("%s not built; run `make ebpf-objs` to exercise this test", path)
	}
	return path
}

// testPinDir returns a private pin directory on the host's bpffs, so a test
// never touches the pins a running nhp-acd owns.
//
// The pins have to live on a bpf filesystem: the kernel refuses to pin
// anywhere else, and the loader's PinPath is where it looks for the maps an
// earlier object created.
func testPinDir(t *testing.T) string {
	t.Helper()
	dir := filepath.Join("/sys/fs/bpf", fmt.Sprintf("nhp-ac-test-%d", os.Getpid()))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Skipf("cannot create a pin directory on /sys/fs/bpf: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

// ---- packet building ----

// The two programs parse the packet the same way: the Ethernet header first,
// then the IPv4 header, then the L4 header at ihl * 4. A test packet is
// therefore built in that order too.

// testEth builds an Ethernet header for an IPv4 payload.
func testEth(src, dst net.IP) []byte {
	b := make([]byte, 14)
	copy(b[0:6], dst.To4())
	copy(b[6:12], src.To4())
	binary.BigEndian.PutUint16(b[12:14], 0x0800)
	return b
}

// testIPv4 builds an IPv4 header with the given protocol and payload length.
func testIPv4(src, dst net.IP, proto uint8, payloadLen int) []byte {
	b := make([]byte, 20)
	b[0] = 0x45 // version 4, ihl 5
	b[2] = byte((20 + payloadLen) >> 8)
	b[3] = byte(20 + payloadLen)
	b[8] = 64 // ttl
	b[9] = proto
	copy(b[12:16], src.To4())
	copy(b[16:20], dst.To4())
	return b
}

// testEcho builds an ICMP echo request (type 8) or reply (type 0) with the
// given identifier.
func testEcho(typ uint8, id uint16) []byte {
	b := make([]byte, 8)
	b[0] = typ
	b[1] = 0
	binary.BigEndian.PutUint16(b[4:6], id)
	binary.BigEndian.PutUint16(b[6:8], 1) // sequence
	return b
}

// icmpEchoPacket builds a whole Ethernet/IPv4/ICMP frame.
func icmpEchoPacket(src, dst net.IP, typ uint8, id uint16) []byte {
	echo := testEcho(typ, id)
	pkt := append(testEth(src, dst), testIPv4(src, dst, 1, len(echo))...)
	return append(pkt, echo...)
}

// ---- the maps, as Go sees them ----

// wireIP and wirePort put a number into the form the eBPF maps use: the raw
// bytes as they appear on the wire, read as a little-endian integer. The
// production code does the same (parseIP and parsePort in ebpf.go), because
// the C side keeps __be32 and __be16 fields and a Go uint32 marshals
// little-endian.
func wireIP(ip net.IP) uint32 {
	return binary.LittleEndian.Uint32(ip.To4())
}

func wirePort(p uint16) uint16 {
	return binary.LittleEndian.Uint16([]byte{byte(p >> 8), byte(p & 0xFF)})
}

// acEchoKey mirrors struct icmp_echo_key in nhp_maps.h, which is packed:
// four bytes of address and two of identifier, with no padding.
type acEchoKey struct {
	PeerIP uint32
	ID     uint16
}

// acEchoValue mirrors struct icmp_echo_value in nhp_maps.h.
type acEchoValue struct {
	ExpireTime uint64
}

// acCtTuple mirrors struct ipv4_ct_tuple in nhp_maps.h, including the two
// bytes of tail padding the C struct's alignment leaves behind.
type acCtTuple struct {
	Daddr   uint32
	Saddr   uint32
	Dport   uint16
	Sport   uint16
	Nexthdr uint8
	Flags   uint8
	_       [2]byte
}

// acConnValue mirrors struct conn_value in nhp_maps.h.
type acConnValue struct {
	Timestamp uint64
	LastTime  uint64
	TtlNs     uint64
	State     uint8
	Flags     uint8
	RxPackets uint32
	TxPackets uint32
	_         [6]byte
}

// CT_DIR_EGRESS and CT_FLAG_AC_CLIENT, from nhp_maps.h.
const (
	acCtDirEgress    = 1
	acCtFlagAcClient = 32
)

// closeObjects closes every map and program a load assigned, so the pins do
// not outlive the test.
func closeObjects(v any) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < rv.NumField(); i++ {
		switch f := rv.Field(i).Interface().(type) {
		case *ebpf.Map:
			if f != nil {
				f.Close()
			}
		case *ebpf.Program:
			if f != nil {
				f.Close()
			}
		}
	}
}

// loadAcObjects loads the AC's two objects without attaching them, sharing
// the maps through the pin directory. The programs are then driven with
// BPF_PROG_TEST_RUN, which is the kernel's own way of running a program
// against a packet.
func loadAcObjects(t *testing.T, pinDir string) (xdp *ebpf.Program, tc *ebpf.Program, maps *bpfObjects) {
	t.Helper()

	spec, err := ebpf.LoadCollectionSpec(acObjPath(t, "nhp_ebpf_xdp.o"))
	if err != nil {
		t.Skipf("cannot read the AC's XDP object: %v", err)
	}
	tcSpec, err := ebpf.LoadCollectionSpec(acObjPath(t, "tc_egress.o"))
	if err != nil {
		t.Skipf("cannot read the AC's TC object: %v", err)
	}

	maps = &bpfObjects{}
	opts := &ebpf.CollectionOptions{Maps: ebpf.MapOptions{PinPath: pinDir}}
	if err := spec.LoadAndAssign(maps, opts); err != nil {
		t.Skipf("cannot load the AC's XDP object (needs CAP_BPF): %v", err)
	}
	t.Cleanup(func() { closeObjects(maps) })

	tcObjs := &tcBpfObjects{}
	if err := tcSpec.LoadAndAssign(tcObjs, opts); err != nil {
		t.Skipf("cannot load the AC's TC object (needs CAP_BPF): %v", err)
	}
	t.Cleanup(func() { closeObjects(tcObjs) })

	return maps.XdpProg, tcObjs.TcEgressProg, maps
}

// runProgram feeds a packet to a program and returns its verdict.
func runProgram(t *testing.T, prog *ebpf.Program, pkt []byte) uint32 {
	t.Helper()
	ret, _, err := prog.Test(pkt)
	if err != nil {
		t.Skipf("cannot run the program against a packet (needs CAP_BPF): %v", err)
	}
	return ret
}

// The AC must be able to get answers to its own pings.
//
// An echo reply used to be passed unconditionally, from anywhere. That was a
// standing bypass into the host once the security group was open, so the
// reply was narrowed to the peers that knocked for ICMP access -- which also
// stopped the AC from answering its own pings: a diagnostic an operator runs
// on the host, and the way the AC's own reachability is checked.
//
// The fix is the shape the rest of the filter uses: the host's own traffic
// is admitted because the host opened the flow. tc_egress.c records the echo
// request, the ingress program admits the reply that answers it -- and only
// that reply.
func TestAcFilterAnswersTheHostsOwnPings(t *testing.T) {
	pinDir := testPinDir(t)
	xdpProg, tcProg, maps := loadAcObjects(t, pinDir)

	host := net.IPv4(10, 0, 0, 1)
	peer := net.IPv4(10, 0, 0, 2)
	const echoID = 0x1234

	// The request the AC sends. It goes out through the TC egress program,
	// which records it.
	req := icmpEchoPacket(host, peer, 8 /* echo request */, echoID)
	if ret := runProgram(t, tcProg, req); ret != tcActOK {
		t.Fatalf("the echo request was not allowed out (TC_ACT %d)", ret)
	}

	key := acEchoKey{PeerIP: wireIP(peer), ID: wirePort(echoID)}
	var val acEchoValue
	if err := maps.IcmpEcho.Lookup(&key, &val); err != nil {
		t.Fatalf("the egress program did not record the echo request: %v", err)
	}
	if val.ExpireTime == 0 {
		t.Error("the recorded echo request has no expiry")
	}

	// The reply that answers it comes back in and has to be admitted.
	reply := icmpEchoPacket(peer, host, 0 /* echo reply */, echoID)
	if ret := runProgram(t, xdpProg, reply); ret != xdpPass {
		t.Errorf("the reply to the AC's own ping was dropped (XDP %d); the host must be able to answer its own diagnostics", ret)
	}

	// A reply that does not answer the request we sent is not ours: another
	// identifier, or another peer.
	otherID := icmpEchoPacket(peer, host, 0, echoID+1)
	if ret := runProgram(t, xdpProg, otherID); ret != xdpDrop {
		t.Errorf("a reply with another identifier was admitted (XDP %d)", ret)
	}
	otherPeer := icmpEchoPacket(net.IPv4(10, 0, 0, 3), host, 0, echoID)
	if ret := runProgram(t, xdpProg, otherPeer); ret != xdpDrop {
		t.Errorf("a reply from a peer the AC never pinged was admitted (XDP %d)", ret)
	}
}

// The AC's egress program has to see the packet the way the ingress program
// does.
//
// The two build the same 4-tuple for the same flow and look it up in the
// same map: the ingress program writes the peer -> AC direction, and the
// egress program rebuilds it in the other direction to recognise a flow the
// host itself opened. If the two read the packet differently -- one of them
// starting at the Ethernet header and the other at the IP header -- the
// tuples never line up, the host's own flows stop being answerable, and
// nothing says so.
//
// The check is the behaviour the whole design rests on: the host sends a
// datagram, and the egress program records the flow it opened.
func TestAcEgressProgramRecordsTheHostsOwnFlow(t *testing.T) {
	pinDir := testPinDir(t)
	_, tcProg, maps := loadAcObjects(t, pinDir)

	host := net.IPv4(10, 0, 0, 1)
	peer := net.IPv4(10, 0, 0, 2)
	const srcPort = 40000
	const dstPort = 9999

	udp := make([]byte, 8)
	binary.BigEndian.PutUint16(udp[0:2], srcPort)
	binary.BigEndian.PutUint16(udp[2:4], dstPort)
	pkt := append(testEth(host, peer), testIPv4(host, peer, 17, len(udp))...)
	pkt = append(pkt, udp...)

	if ret := runProgram(t, tcProg, pkt); ret != tcActOK {
		t.Fatalf("the datagram was not allowed out (TC_ACT %d)", ret)
	}

	key := acCtTuple{
		Daddr:   wireIP(peer),
		Saddr:   wireIP(host),
		Dport:   wirePort(dstPort),
		Sport:   wirePort(srcPort),
		Nexthdr: 17,
		Flags:   acCtDirEgress,
	}
	var val acConnValue
	if err := maps.Conntrack.Lookup(&key, &val); err != nil {
		t.Fatalf("the egress program did not record the flow the host opened: %v", err)
	}
	if val.Flags&acCtFlagAcClient == 0 {
		t.Errorf("the recorded flow is not claimed as the AC's own (flags %d)", val.Flags)
	}
}

// The two programs must agree on the tuple.
//
// The egress program records the flow the host opened, and the ingress
// program has to find that entry when the reply comes back -- it rebuilds
// the tuple in the other direction to do so. If the two disagreed about the
// tuple (a byte order, a port, the direction flag), the host's own traffic
// would stop being answerable and every verdict would still look right.
//
// The check is the round trip: the host sends a datagram, and the reply that
// comes back for it is admitted.
func TestAcProgramsAgreeOnTheTuple(t *testing.T) {
	pinDir := testPinDir(t)
	xdpProg, tcProg, _ := loadAcObjects(t, pinDir)

	host := net.IPv4(10, 0, 0, 1)
	peer := net.IPv4(10, 0, 0, 2)
	const srcPort = 40000
	const dstPort = 9999

	udp := make([]byte, 8)
	binary.BigEndian.PutUint16(udp[0:2], srcPort)
	binary.BigEndian.PutUint16(udp[2:4], dstPort)
	out := append(testEth(host, peer), testIPv4(host, peer, 17, len(udp))...)
	out = append(out, udp...)
	if ret := runProgram(t, tcProg, out); ret != tcActOK {
		t.Fatalf("the datagram was not allowed out (TC_ACT %d)", ret)
	}

	// The reply the peer sends back. Nothing else admits it: there is no
	// knock, no socket bound to that port on this host, and no whitelist
	// entry -- only the flow the egress program recorded.
	replyUdp := make([]byte, 8)
	binary.BigEndian.PutUint16(replyUdp[0:2], dstPort)
	binary.BigEndian.PutUint16(replyUdp[2:4], srcPort)
	in := append(testEth(peer, host), testIPv4(peer, host, 17, len(replyUdp))...)
	in = append(in, replyUdp...)
	if ret := runProgram(t, xdpProg, in); ret != xdpPass {
		t.Errorf("the reply to the AC's own flow was dropped (XDP %d); the two programs disagree on the tuple", ret)
	}
}

// The AC's filter has to keep answering the peers that knocked for ICMP, and
// keep dropping the ones that did not.
//
// The echo reply is admitted on two grounds: the peer knocked for ICMP
// access, or the AC pinged the peer itself. The first is what lets a
// legitimate peer reach the service it knocked for; the second is the
// host's own traffic. Neither may become a way in for anybody else.
func TestAcFilterKeepsTheKnockedEchoReply(t *testing.T) {
	pinDir := testPinDir(t)
	xdpProg, _, maps := loadAcObjects(t, pinDir)

	host := net.IPv4(10, 0, 0, 1)
	peer := net.IPv4(10, 0, 0, 2)

	// A peer that knocked for ICMP access: an entry in `icmpwhitelist` for
	// the pair, with a window that has not closed. The clock is the one the
	// program compares against: bpf_ktime_get_ns(), i.e. CLOCK_MONOTONIC.
	now, err := getBootTimeNanos()
	if err != nil {
		t.Skipf("cannot read the monotonic clock: %v", err)
	}
	key := srcDestKey{
		SrcIP: wireIP(peer),
		DstIP: wireIP(host),
	}
	if err := maps.Icmpwhitelist.Put(key, whitelistValue{
		Allowed:    1,
		ExpireTime: now + uint64(time.Minute),
	}); err != nil {
		t.Skipf("cannot write the ICMP whitelist: %v", err)
	}

	reply := icmpEchoPacket(peer, host, 0, 0x4321)
	if ret := runProgram(t, xdpProg, reply); ret != xdpPass {
		t.Errorf("the reply to a peer that knocked for ICMP was dropped (XDP %d)", ret)
	}

	// The same reply from a peer that never knocked stays dropped.
	stranger := icmpEchoPacket(net.IPv4(10, 0, 0, 9), host, 0, 0x4321)
	if ret := runProgram(t, xdpProg, stranger); ret != xdpDrop {
		t.Errorf("a reply from a peer that never knocked was admitted (XDP %d)", ret)
	}
}
