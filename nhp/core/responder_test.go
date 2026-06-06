package core

import (
	"bytes"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// TestShouldCheckRecvAttack_AOPNoLongerExempt is the regression
// fence for issue #1123. The previous implementation skipped the
// per-connection LastRemoteSendTime monotonic check for NHP_AOP,
// which made the responder a no-op replay-protector for AC-side
// AOP processing within an open connection. The fix removes that
// exemption so AOP rides the same gate as every other AC-bound
// message; the cross-connection cousin lives in
// endpoints/ac/aop_replay_cache.go.
func TestShouldCheckRecvAttack_AOPNoLongerExempt(t *testing.T) {
	if !shouldCheckRecvAttack(NHP_AC, NHP_SERVER, NHP_AOP) {
		t.Fatal("NHP_AOP on AC must be subject to the LastRemoteSendTime gate (#1123)")
	}
}

// TestShouldCheckRecvAttack_ARTStillExempt pins the remaining
// exemption: NHP_ART (AC → server response) skips the gate because
// the server-side transaction layer already correlates by
// TransactionId and the AC→server hop occasionally exceeds the
// flood-gate threshold (MinimalRecvIntervalMs in constants.go).
// Tracked for follow-up dedupe in #1457.
func TestShouldCheckRecvAttack_ARTStillExempt(t *testing.T) {
	if shouldCheckRecvAttack(NHP_SERVER, NHP_AC, NHP_ART) {
		t.Fatal("NHP_ART on server must remain exempt from the LastRemoteSendTime gate")
	}
}

// TestShouldCheckRecvAttack_DefaultEnforced sanity-checks that an
// unrelated message type still enforces the gate, so a future
// refactor of shouldCheckRecvAttack cannot silently widen the
// exemption set.
func TestShouldCheckRecvAttack_DefaultEnforced(t *testing.T) {
	if !shouldCheckRecvAttack(NHP_SERVER, NHP_AGENT, NHP_KNK) {
		t.Fatal("non-exempt (deviceType, peerType, msgType) must enforce the gate")
	}
}

// TestShouldCheckFlood_AOPExempt fences the round-7 cr finding for
// issue #1123. The flood gate (`MinimalRecvIntervalMs = 20 ms`)
// must NOT apply to NHP_AOP on AC: the server legitimately emits
// AOPs in tight succession during knock bursts, and applying the
// 20 ms floor would false-flood-block a connection past
// `ThreatCountBeforeBlock`. Replay protection is preserved via
// shouldCheckRecvAttack (still enforced on AOP) plus the
// cross-connection cache in endpoints/ac/aop_replay_cache.go.
func TestShouldCheckFlood_AOPExempt(t *testing.T) {
	if shouldCheckFlood(NHP_AC, NHP_SERVER, NHP_AOP) {
		t.Fatal("NHP_AOP on AC must be exempt from the 20 ms flood gate (#1123 round-7)")
	}
}

// TestShouldCheckFlood_ARTExempt mirrors the existing replay-gate
// exemption on the flood gate so the AC→server hop's legitimate
// latency does not flood-block a connection.
func TestShouldCheckFlood_ARTExempt(t *testing.T) {
	if shouldCheckFlood(NHP_SERVER, NHP_AC, NHP_ART) {
		t.Fatal("NHP_ART on server must remain exempt from the 20 ms flood gate")
	}
}

// TestShouldCheckFlood_DefaultEnforced asserts the flood gate is
// otherwise enforced, so a future refactor of shouldCheckFlood
// cannot silently widen the exemption set.
func TestShouldCheckFlood_DefaultEnforced(t *testing.T) {
	if !shouldCheckFlood(NHP_SERVER, NHP_AGENT, NHP_KNK) {
		t.Fatal("non-exempt (deviceType, peerType, msgType) must enforce the flood gate")
	}
}

// TestShouldEscalateStale_AOPExempt fences the #1464 decision that a
// stale NHP_AOP (server→AC) is dropped but NOT escalated to
// threat/block. Mirrors the existing shouldCheckFlood AOP exemption:
// AOP must not be able to self-inflict a connection block (here, on a
// clock-skew false-reject; there, on a legitimate burst).
func TestShouldEscalateStale_AOPExempt(t *testing.T) {
	if shouldEscalateStale(NHP_AC, NHP_SERVER, NHP_AOP) {
		t.Fatal("stale NHP_AOP on AC must NOT escalate to threat/block (#1464) — a benign clock-skew drop must not sever the server→AC connection")
	}
}

// TestShouldEscalateStale_DefaultEnforced asserts every other
// (deviceType, peerType, msgType) — including AC→server ART — still
// escalates a stale drop, so the AOP exemption stays narrowly scoped
// and a refactor cannot silently disable the block path fleet-wide.
func TestShouldEscalateStale_DefaultEnforced(t *testing.T) {
	cases := []struct {
		name                          string
		deviceType, peerType, msgType int
	}{
		{"agent knock", NHP_SERVER, NHP_AGENT, NHP_KNK},
		{"AC→server ART", NHP_SERVER, NHP_AC, NHP_ART},
		// AOP on a non-AC device must NOT pick up the AC-scoped exemption.
		{"AOP wrong device scope", NHP_SERVER, NHP_SERVER, NHP_AOP},
	}
	for _, tc := range cases {
		if !shouldEscalateStale(tc.deviceType, tc.peerType, tc.msgType) {
			t.Errorf("%s: stale drop must escalate to threat/block", tc.name)
		}
	}
}

// TestRecvStalenessFloor_AOPTightened is the regression fence for
// issue #1464. NHP_AOP (server→AC) must use the tighter
// AOPRecvStalenessFloorSeconds floor, not the 600 s default, so the
// cross-restart replay window the AC dedupe cache cannot cover after
// a restart is bounded by the smaller value. A refactor that drops
// the AOP override would silently restore the 600 s window.
func TestRecvStalenessFloor_AOPTightened(t *testing.T) {
	got := recvStalenessFloor(NHP_AC, NHP_SERVER, NHP_AOP)
	want := AOPRecvStalenessFloorSeconds * int64(time.Second)
	if got != want {
		t.Fatalf("NHP_AOP staleness floor = %d ns, want %d ns (tightened for #1464)", got, want)
	}
	// Pin the security invariant the tightening exists to deliver:
	// the AOP floor must be strictly smaller than the default, or the
	// override is doing nothing.
	if got >= DefaultRecvStalenessFloorSeconds*int64(time.Second) {
		t.Fatalf("AOP floor (%d ns) must be strictly tighter than the default (%d ns)", got, DefaultRecvStalenessFloorSeconds*int64(time.Second))
	}
}

// TestRecvStalenessFloor_DefaultPreserved asserts every non-AOP
// (deviceType, peerType, msgType) keeps the historical 600 s floor,
// so #1464's tightening cannot silently narrow the clock-calibration
// tolerance for agent→server knock paths (which may traverse the
// public internet).
func TestRecvStalenessFloor_DefaultPreserved(t *testing.T) {
	want := DefaultRecvStalenessFloorSeconds * int64(time.Second)
	cases := []struct {
		name                          string
		deviceType, peerType, msgType int
	}{
		{"agent knock", NHP_SERVER, NHP_AGENT, NHP_KNK},
		{"AC→server ART", NHP_SERVER, NHP_AC, NHP_ART},
		// AOP on a non-AC device must NOT pick up the AC-scoped
		// override — the tighter floor is keyed on the full triple.
		{"AOP wrong device scope", NHP_SERVER, NHP_SERVER, NHP_AOP},
	}
	for _, tc := range cases {
		if got := recvStalenessFloor(tc.deviceType, tc.peerType, tc.msgType); got != want {
			t.Errorf("%s: staleness floor = %d ns, want default %d ns", tc.name, got, want)
		}
	}
}

type validatePeerFixture struct {
	acPeer        *UdpPeer
	server        *Device
	serverConn    *ConnectionData
	packetContent []byte
	initTime      int64
}

func TestValidatePeerPopulatesRemotePubKey(t *testing.T) {
	fixture := newValidatePeerFixture(t)
	ppd := fixture.parseAndValidate(t)
	defer ppd.Destroy()

	if !bytes.Equal(ppd.RemotePubKey, fixture.acPeer.PublicKey()) {
		t.Fatalf("RemotePubKey mismatch\ngot:  %x\nwant: %x", ppd.RemotePubKey, fixture.acPeer.PublicKey())
	}
}

func TestValidatePeerRemotePubKeySurvivesDestroy(t *testing.T) {
	fixture := newValidatePeerFixture(t)
	ppd := fixture.parseAndValidate(t)
	want := append([]byte(nil), fixture.acPeer.PublicKey()...)

	ppd.Destroy()

	if !bytes.Equal(ppd.RemotePubKey, want) {
		t.Fatalf("RemotePubKey after Destroy mismatch\ngot:  %x\nwant: %x", ppd.RemotePubKey, want)
	}
}

// TestValidatePeer_AOPStalenessFloorWiredAtCallSite is the call-site
// fence requested in cr on #1464. The recvStalenessFloor unit tests
// prove the helper returns the right number, but the behavior that
// ships is its wiring into validatePeer's stale check — a refactor
// could leave the helper correct yet disconnect it from the call site
// and every pure-helper test would stay green. This drives real
// AEAD-authenticated packets through validatePeer and asserts the
// differential the tightening delivers, end to end:
//
//   - an AOP (server→AC) aged past the 120 s AOP floor but within the
//     600 s default is rejected as stale (the tightened floor is wired);
//   - the same AOP just under the floor is accepted (no false-reject);
//   - an ART (AC→server) at the AOP-rejecting age still passes under
//     the unchanged 600 s default (the tightening is AOP-scoped).
//
// The validatePeerFixture helpers already stand up the full
// Device/Peer/ECDH/AEAD chain, so this is the cheap end-to-end fence
// the recvStalenessFloor docstring's "#1468" note assumed was too
// heavy to write before the fixture existed.
func TestValidatePeer_AOPStalenessFloorWiredAtCallSite(t *testing.T) {
	silenceGlobalLogger(t)

	acDevice := NewDevice(NHP_AC, validatePeerPrivateKey(1), nil)
	serverDevice := NewDevice(NHP_SERVER, validatePeerPrivateKey(33), nil)
	if acDevice == nil || serverDevice == nil {
		t.Fatal("failed to create AC/server devices")
	}
	acPeer := &UdpPeer{PubKeyBase64: acDevice.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12345, Type: NHP_AC}
	serverPeer := &UdpPeer{PubKeyBase64: serverDevice.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12346, Type: NHP_SERVER}
	serverDevice.AddPeer(acPeer)
	acDevice.AddPeer(serverPeer)

	// validateAged runs one packet through the receiver's validatePeer
	// on a fresh connection (LastRemoteSendTime == 0, so the replay gate
	// never trips) and returns the error (nil == accepted).
	validateAged := func(receiver *Device, localPort, remotePort int, pkt *Packet, initTime int64) error {
		t.Helper()
		ppd, err := receiver.createPacketParserData(&PacketData{
			BasePacket: pkt,
			ConnData:   validatePeerConnectionData(receiver, localPort, remotePort),
			InitTime:   initTime,
		})
		if err != nil {
			t.Fatalf("createPacketParserData failed: %v", err)
		}
		defer ppd.Destroy()
		return ppd.validatePeer()
	}

	const overAOPFloor = 200 * time.Second // > 120 s AOP floor, < 600 s default
	const underAOPFloor = 60 * time.Second // < 120 s AOP floor

	// AOP (server→AC) aged past the AOP floor → rejected as stale.
	pkt, initTime := buildAgedPacket(t, serverDevice, validatePeerConnectionData(serverDevice, 12346, 12345), acPeer.PublicKey(), NHP_AOP, overAOPFloor)
	if err := validateAged(acDevice, 12345, 12346, pkt, initTime); !errors.Is(err, ErrStalePacketReceived) {
		t.Fatalf("AOP aged %s: got %v, want ErrStalePacketReceived (120 s AOP floor must be wired at the call site)", overAOPFloor, err)
	}

	// AOP (server→AC) just under the AOP floor → accepted.
	pkt, initTime = buildAgedPacket(t, serverDevice, validatePeerConnectionData(serverDevice, 12346, 12345), acPeer.PublicKey(), NHP_AOP, underAOPFloor)
	if err := validateAged(acDevice, 12345, 12346, pkt, initTime); err != nil {
		t.Fatalf("AOP aged %s (under the 120 s floor): got %v, want accepted", underAOPFloor, err)
	}

	// ART (AC→server) at the AOP-rejecting age → still accepted under
	// the unchanged 600 s default floor.
	pkt, initTime = buildAgedPacket(t, acDevice, validatePeerConnectionData(acDevice, 12345, 12346), serverPeer.PublicKey(), NHP_ART, overAOPFloor)
	if err := validateAged(serverDevice, 12346, 12345, pkt, initTime); err != nil {
		t.Fatalf("ART aged %s: got %v, want accepted (default 600 s floor is unchanged for non-AOP)", overAOPFloor, err)
	}
}

// TestValidatePeer_StaleAOPDropsWithoutBlock is the call-site fence
// for the #1464 escalation exemption. It drives a stale AEAD packet
// through validatePeer on an explicit connection and inspects the
// connection's threat/block state afterward:
//
//   - a stale AOP (server→AC) is dropped (ErrStalePacketReceived) but
//     leaves RecvThreatCount at 0 and never fires SendBlockSignal, so a
//     benign clock-skew false-reject can't sever the server→AC link;
//   - a stale ART (AC→server) at the same relative age IS escalated
//     (RecvThreatCount bumped), proving the exemption is AOP-scoped and
//     the block path is otherwise intact.
//
// This is the behavioral counterpart to the shouldEscalateStale unit
// tests — it fences the wiring of the exemption into validatePeer's
// stale branch, which a refactor could disconnect while the predicate
// stays correct.
func TestValidatePeer_StaleAOPDropsWithoutBlock(t *testing.T) {
	silenceGlobalLogger(t)

	acDevice := NewDevice(NHP_AC, validatePeerPrivateKey(1), nil)
	serverDevice := NewDevice(NHP_SERVER, validatePeerPrivateKey(33), nil)
	if acDevice == nil || serverDevice == nil {
		t.Fatal("failed to create AC/server devices")
	}
	acPeer := &UdpPeer{PubKeyBase64: acDevice.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12345, Type: NHP_AC}
	serverPeer := &UdpPeer{PubKeyBase64: serverDevice.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12346, Type: NHP_SERVER}
	serverDevice.AddPeer(acPeer)
	acDevice.AddPeer(serverPeer)

	validateOn := func(receiver *Device, conn *ConnectionData, pkt *Packet, initTime int64) error {
		t.Helper()
		ppd, err := receiver.createPacketParserData(&PacketData{BasePacket: pkt, ConnData: conn, InitTime: initTime})
		if err != nil {
			t.Fatalf("createPacketParserData failed: %v", err)
		}
		defer ppd.Destroy()
		return ppd.validatePeer()
	}

	// Stale AOP (server→AC), aged past the 120 s AOP floor → dropped,
	// NOT escalated. Drive TWO stale AOPs through the SAME connection so
	// the block-suppression is genuinely exercised, not a single-packet
	// no-op: without the exemption the first would bump RecvThreatCount
	// to 1 and the second would cross ThreatCountBeforeBlock (=1) and
	// fire SendBlockSignal. With the exemption RecvThreatCount stays 0
	// and BlockSignal never fires, so BOTH assertions below are
	// load-bearing. (The staleness gate returns before LastRemoteSendTime
	// is updated, so the second packet still reaches the stale branch
	// rather than tripping the replay gate.) RecvThreatCount is read via
	// atomic.LoadInt32 to match how validatePeer writes it.
	acConn := validatePeerConnectionData(acDevice, 12345, 12346)
	for i := 1; i <= 2; i++ {
		pkt, initTime := buildAgedPacket(t, serverDevice, validatePeerConnectionData(serverDevice, 12346, 12345), acPeer.PublicKey(), NHP_AOP, 200*time.Second)
		if err := validateOn(acDevice, acConn, pkt, initTime); !errors.Is(err, ErrStalePacketReceived) {
			t.Fatalf("stale AOP #%d: got %v, want ErrStalePacketReceived", i, err)
		}
	}
	if got := atomic.LoadInt32(&acConn.RecvThreatCount); got != 0 {
		t.Fatalf("two stale AOPs must NOT bump RecvThreatCount (got %d) — AOP is exempt from stale escalation (#1464)", got)
	}
	if len(acConn.BlockSignal) != 0 {
		t.Fatal("two stale AOPs must NOT fire SendBlockSignal — without the exemption the 2nd would cross ThreatCountBeforeBlock and sever the trusted server→AC connection")
	}

	// Stale ART (AC→server), aged past the 600 s default floor → dropped
	// AND escalated (one stale packet bumps RecvThreatCount to 1; a
	// second would cross ThreatCountBeforeBlock and block).
	srvConn := validatePeerConnectionData(serverDevice, 12346, 12345)
	pkt, initTime := buildAgedPacket(t, acDevice, validatePeerConnectionData(acDevice, 12345, 12346), serverPeer.PublicKey(), NHP_ART, 700*time.Second)
	if err := validateOn(serverDevice, srvConn, pkt, initTime); !errors.Is(err, ErrStalePacketReceived) {
		t.Fatalf("stale ART: got %v, want ErrStalePacketReceived", err)
	}
	if got := atomic.LoadInt32(&srvConn.RecvThreatCount); got != 1 {
		t.Fatalf("stale ART must escalate (RecvThreatCount=1); got %d — the AOP exemption must be scoped, not global", got)
	}
}

// BenchmarkValidatePeer uses ART so one encrypted packet can be replayed through
// validatePeer without tripping the server-side replay/flood gates.
func BenchmarkValidatePeer(b *testing.B) {
	fixture := newValidatePeerFixture(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ppd := fixture.parseAndValidate(b)
		ppd.Destroy()
	}
}

func newValidatePeerFixture(tb testing.TB) validatePeerFixture {
	tb.Helper()
	silenceGlobalLogger(tb)

	acPrivKey := validatePeerPrivateKey(1)
	serverPrivKey := validatePeerPrivateKey(33)
	acDevice := NewDevice(NHP_AC, acPrivKey, nil)
	if acDevice == nil {
		tb.Fatal("failed to create AC device")
	}
	serverDevice := NewDevice(NHP_SERVER, serverPrivKey, nil)
	if serverDevice == nil {
		tb.Fatal("failed to create server device")
	}

	acPeer := &UdpPeer{
		PubKeyBase64: acDevice.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         12345,
		Type:         NHP_AC,
	}
	serverPeer := &UdpPeer{
		PubKeyBase64: serverDevice.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         12346,
		Type:         NHP_SERVER,
	}
	serverDevice.AddPeer(acPeer)
	acDevice.AddPeer(serverPeer)

	acConn := validatePeerConnectionData(acDevice, 12345, 12346)
	mad, err := acDevice.MsgToPacket(&MsgData{
		ConnData:      acConn,
		PeerPk:        serverPeer.PublicKey(),
		HeaderType:    NHP_ART,
		TransactionId: 1,
	})
	if err != nil {
		tb.Fatalf("MsgToPacket failed: %v", err)
	}

	return validatePeerFixture{
		acPeer:        acPeer,
		server:        serverDevice,
		serverConn:    validatePeerConnectionData(serverDevice, 12346, 12345),
		packetContent: append([]byte(nil), mad.BasePacket.Content...),
		initTime:      time.Now().UnixNano(),
	}
}

func (f validatePeerFixture) parseAndValidate(tb testing.TB) *PacketParserData {
	tb.Helper()

	pkt := Packet{
		Content:    f.packetContent,
		HeaderType: NHP_ART,
	}
	pd := PacketData{
		BasePacket: &pkt,
		ConnData:   f.serverConn,
		InitTime:   f.initTime,
	}

	ppd, err := f.server.createPacketParserData(&pd)
	if err != nil {
		tb.Fatalf("createPacketParserData failed: %v", err)
	}
	if err := ppd.validatePeer(); err != nil {
		tb.Fatalf("validatePeer failed: %v", err)
	}

	return ppd
}

func validatePeerPrivateKey(start byte) []byte {
	key := make([]byte, PrivateKeySize)
	for i := range key {
		key[i] = start + byte(i)
	}
	return key
}

func validatePeerConnectionData(device *Device, localPort int, remotePort int) *ConnectionData {
	return &ConnectionData{
		Device:           device,
		LocalAddr:        &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: localPort},
		RemoteAddr:       &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: remotePort},
		InitTime:         time.Now().UnixNano(),
		CookieStore:      &CookieStore{},
		SendQueue:        make(chan *Packet, 1),
		RecvQueue:        make(chan *Packet, 1),
		BlockSignal:      make(chan struct{}, 1),
		SetTimeoutSignal: make(chan struct{}, 1),
		StopSignal:       make(chan struct{}),
	}
}

// buildAgedPacket encrypts one packet from sender→peerPk and returns it
// paired with an InitTime aged `age` past the packet's send timestamp,
// so a later validatePeer sees the packet as `age` old. The send
// timestamp is stamped inside MsgToPacket at ~now; reading now right
// after is accurate to microseconds — far finer than the second-scale
// staleness floors — and any slack only makes the packet marginally
// OLDER, never crossing a boundary for the second-scale ages callers
// use. Shared by the staleness-floor call-site fences.
func buildAgedPacket(tb testing.TB, sender *Device, senderConn *ConnectionData, peerPk []byte, hdr int, age time.Duration) (*Packet, int64) {
	tb.Helper()
	mad, err := sender.MsgToPacket(&MsgData{ConnData: senderConn, PeerPk: peerPk, HeaderType: hdr, TransactionId: 1})
	if err != nil {
		tb.Fatalf("MsgToPacket(hdr=%d) failed: %v", hdr, err)
	}
	return &Packet{Content: append([]byte(nil), mad.BasePacket.Content...), HeaderType: hdr}, time.Now().UnixNano() + int64(age)
}

// overloadKnockFixture builds a real knock packet (NHP_KNK or DHP_KNK) from an
// agent plus the server device that parses it, to exercise the NHP-COK
// early-drop cookie path in validatePeer. registerAgent controls whether the
// agent's pubkey is in the server's peer pool — with agent peer validation on
// (the NHP_SERVER default), that decides whether the knock passes the peer-pool
// gate. These tests fence the behavior the #1131 review confirmed must hold: the
// overload cookie is issued for valid peers but never reflected to a pubkey that
// fails the peer-pool gate.
type overloadKnockFixture struct {
	server        *Device
	serverConn    *ConnectionData
	packetContent []byte
	initTime      int64
	headerType    int
}

func newOverloadKnockFixture(tb testing.TB, registerAgent bool, headerType int) overloadKnockFixture {
	tb.Helper()
	silenceGlobalLogger(tb)

	agentPrivKey := validatePeerPrivateKey(70)
	serverPrivKey := validatePeerPrivateKey(33)
	agentDevice := NewDevice(NHP_AGENT, agentPrivKey, nil)
	if agentDevice == nil {
		tb.Fatal("failed to create agent device")
	}
	serverDevice := NewDevice(NHP_SERVER, serverPrivKey, nil)
	if serverDevice == nil {
		tb.Fatal("failed to create server device")
	}

	serverPeer := &UdpPeer{
		PubKeyBase64: serverDevice.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         12346,
		Type:         NHP_SERVER,
	}
	agentDevice.AddPeer(serverPeer)

	if registerAgent {
		serverDevice.AddPeer(&UdpPeer{
			PubKeyBase64: agentDevice.PublicKeyBase64(),
			Ip:           "127.0.0.1",
			Port:         12345,
			Type:         NHP_AGENT,
		})
	}

	agentConn := validatePeerConnectionData(agentDevice, 12345, 12346)
	mad, err := agentDevice.MsgToPacket(&MsgData{
		ConnData:      agentConn,
		PeerPk:        serverPeer.PublicKey(),
		HeaderType:    headerType,
		TransactionId: 1,
	})
	if err != nil {
		tb.Fatalf("MsgToPacket(%s) failed: %v", HeaderTypeToString(headerType), err)
	}

	return overloadKnockFixture{
		server:        serverDevice,
		serverConn:    validatePeerConnectionData(serverDevice, 12346, 12345),
		packetContent: append([]byte(nil), mad.BasePacket.Content...),
		initTime:      time.Now().UnixNano(),
		headerType:    headerType,
	}
}

// parse runs createPacketParserData + validatePeer and returns the resulting
// ppd together with the validatePeer error. Unlike validatePeerFixture's
// parseAndValidate, it never t.Fatal()s on that error — these tests assert on
// it directly.
func (f overloadKnockFixture) parse(tb testing.TB) (*PacketParserData, error) {
	tb.Helper()
	pkt := Packet{
		Content:    f.packetContent,
		HeaderType: f.headerType,
	}
	pd := PacketData{
		BasePacket: &pkt,
		ConnData:   f.serverConn,
		InitTime:   f.initTime,
	}
	ppd, err := f.server.createPacketParserData(&pd)
	if err != nil {
		return ppd, err
	}
	return ppd, ppd.validatePeer()
}

// TestValidatePeerOverloadKnockIssuesCookie pins the spec's NHP-COK early-drop
// (NHP.pdf §NHP-COK): under server overload, a registered agent's knock is
// answered with a cookie challenge (ErrServerRejectWithCookie) and the cookie
// is generated, so the agent can re-knock (NHP-RKN) carrying it. Covers both
// knock header types the overload short-circuit handles (NHP_KNK and DHP_KNK).
func TestValidatePeerOverloadKnockIssuesCookie(t *testing.T) {
	for _, headerType := range []int{NHP_KNK, DHP_KNK} {
		t.Run(HeaderTypeToString(headerType), func(t *testing.T) {
			fixture := newOverloadKnockFixture(t, true, headerType)
			fixture.server.SetOverload(true)

			ppd, err := fixture.parse(t)
			defer ppd.Destroy()

			assertNHPError(t, err, ErrServerRejectWithCookie)
			if IsZero(fixture.serverConn.CookieStore.CurrCookie[:]) {
				t.Fatal("overload knock: cookie was not generated (CookieStore.CurrCookie still zero)")
			}
		})
	}
}

// TestValidatePeerOverloadKnockUnregisteredNoCookie is the reflection-safety
// fence flagged by the #1131 review. With agent peer validation on (the
// NHP_SERVER default), an unregistered pubkey must be dropped at the peer-pool
// gate (ErrPeerNotFound) and must NOT elicit a cookie. A COK reply (~1.23x the
// minimal KNK that triggers it) sent to a forged source would be an
// amplification primitive — exactly the vector #1131 item 1 warns about — so a
// future refactor that moves the cookie emit ahead of this gate must fail here.
func TestValidatePeerOverloadKnockUnregisteredNoCookie(t *testing.T) {
	fixture := newOverloadKnockFixture(t, false, NHP_KNK)
	fixture.server.SetOverload(true)

	ppd, err := fixture.parse(t)
	defer ppd.Destroy()

	assertNHPError(t, err, ErrPeerNotFound)
	if !IsZero(fixture.serverConn.CookieStore.CurrCookie[:]) {
		t.Fatal("unregistered overload KNK must not generate a cookie (reflection guard)")
	}
}

// TestValidatePeerKnockNotOverloadCompletesHandshake confirms the cookie
// short-circuit is overload-gated: with the server NOT overloaded, a registered
// agent's KNK runs validatePeer to completion and no cookie is generated.
func TestValidatePeerKnockNotOverloadCompletesHandshake(t *testing.T) {
	fixture := newOverloadKnockFixture(t, true, NHP_KNK)
	// server overload deliberately left false

	ppd, err := fixture.parse(t)
	defer ppd.Destroy()

	if err != nil {
		t.Fatalf("non-overload KNK: validatePeer returned %v, want nil", err)
	}
	if !IsZero(fixture.serverConn.CookieStore.CurrCookie[:]) {
		t.Fatal("non-overload KNK must not generate a cookie")
	}
}
