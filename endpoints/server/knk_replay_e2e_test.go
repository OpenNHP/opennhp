package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// knkReplayFixture builds a real agent and a real server device, wires them
// as peers, and produces a genuine encrypted NHP_KNK packet — the same bytes
// a capturing attacker would lift off the wire.
type knkReplayFixture struct {
	agent    *core.Device
	server   *core.Device
	agentPk  []byte
	serverPk []byte
}

// newKNKReplayFixture sets up the agent/server pair. The agent's key is
// registered on the server as an agent peer, which is what lets validatePeer
// accept the knock.
func newKNKReplayFixture(t *testing.T) *knkReplayFixture {
	t.Helper()

	agentKey := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	serverKey := core.NewECDH(core.ECC_CURVE25519).PrivateKey()

	agent := core.NewDevice(core.NHP_AGENT, agentKey, nil)
	if agent == nil {
		t.Fatal("failed to create the agent device")
	}
	t.Cleanup(agent.Stop)
	server := core.NewDevice(core.NHP_SERVER, serverKey, nil)
	if server == nil {
		t.Fatal("failed to create the server device")
	}
	t.Cleanup(server.Stop)

	agentPeer := &core.UdpPeer{
		PubKeyBase64: server.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         62206,
		Type:         core.NHP_SERVER,
	}
	agent.AddPeer(agentPeer)
	server.AddPeer(&core.UdpPeer{
		PubKeyBase64: agent.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         62207,
		Type:         core.NHP_AGENT,
	})

	agentPk, err := base64.StdEncoding.DecodeString(agent.PublicKeyBase64())
	if err != nil {
		t.Fatal(err)
	}
	serverPk, err := base64.StdEncoding.DecodeString(server.PublicKeyBase64())
	if err != nil {
		t.Fatal(err)
	}
	return &knkReplayFixture{agent: agent, server: server, agentPk: agentPk, serverPk: serverPk}
}

// buildKnock produces a genuine encrypted NHP_KNK packet from the agent,
// along with the transaction id and the header counter it carries.
func (f *knkReplayFixture) buildKnock(t *testing.T) ([]byte, uint64) {
	t.Helper()

	body, err := json.Marshal(&common.AgentKnockMsg{
		HeaderType:    core.NHP_KNK,
		UserId:        "knk-replay-user",
		DeviceId:      "knk-replay-device",
		AuthServiceId: "test-asp",
		ResourceId:    "test-res",
	})
	if err != nil {
		t.Fatal(err)
	}

	trxId := f.agent.NextCounterIndex()
	mad, err := f.agent.MsgToPacket(&core.MsgData{
		HeaderType:    core.NHP_KNK,
		CipherScheme:  common.CIPHER_SCHEME_CURVE,
		TransactionId: trxId,
		Message:       body,
		PeerPk:        f.serverPk,
	})
	if err != nil {
		t.Fatal(err)
	}
	return append([]byte(nil), mad.BasePacket.Content...), trxId
}

// parseKnock runs the server's receive path over a captured packet and
// returns the parsed data, exactly as the packet routine would. The packet
// arrives on a fresh connection from srcAddr — the replay scenario needs a
// NEW connection, because that is what defeats the per-connection
// monotonic gate the replay cache backs up.
func (f *knkReplayFixture) parseKnock(t *testing.T, raw []byte, srcAddr string) *core.PacketParserData {
	t.Helper()

	buf := make([]byte, len(raw))
	copy(buf, raw)
	pkt := &core.Packet{Content: buf}
	pkt.HeaderType = core.NHP_KNK

	conn := f.newConn(srcAddr)
	ppd, err := f.server.PacketToMsg(&core.PacketData{BasePacket: pkt, ConnData: conn})
	if err != nil {
		t.Fatalf("server failed to parse the knock: %v", err)
	}
	// PacketToMsg destroys the ppd on return; rebuild the fields the test
	// needs from the authenticated header instead of keeping the pointer.
	return &core.PacketParserData{
		HeaderType:     ppd.HeaderType,
		SenderTrxId:    ppd.SenderTrxId,
		RemotePubKey:   ppd.RemotePubKey,
		RemoteSendTime: ppd.RemoteSendTime,
	}
}

// newConn builds the server-side connection a packet from srcAddr arrives
// on. Each call is a distinct connection, which is what an attacker replaying
// from a different address gets.
func (f *knkReplayFixture) newConn(srcAddr string) *core.ConnectionData {
	remote, err := net.ResolveUDPAddr("udp", srcAddr)
	if err != nil {
		remote = &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 5000}
	}
	return &core.ConnectionData{
		Device:           f.server,
		LocalAddr:        &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 62206},
		RemoteAddr:       remote,
		InitTime:         time.Now().UnixNano(),
		CookieStore:      &core.CookieStore{},
		SendQueue:        make(chan *core.Packet, 1),
		RecvQueue:        make(chan *core.Packet, 1),
		BlockSignal:      make(chan struct{}, 1),
		SetTimeoutSignal: make(chan struct{}, 1),
		StopSignal:       make(chan struct{}),
	}
}

// TestKNKReplayEndToEnd walks the whole attack: a legitimate agent knocks,
// the server accepts it, and the very same bytes replayed from another
// source address are refused. It also pins that a legitimate retry — a new
// packet with a new counter and a new timestamp — is still accepted.
func TestKNKReplayEndToEnd(t *testing.T) {
	f := newKNKReplayFixture(t)
	s := &UdpServer{knkReplay: newKNKReplayCache()}

	raw, _ := f.buildKnock(t)

	// The legitimate knock arrives.
	ppd := f.parseKnock(t, raw, "192.0.2.10:5000")
	if ppd.RemotePubKey == nil {
		t.Fatal("the parsed knock carries no peer public key")
	}
	if base64.StdEncoding.EncodeToString(ppd.RemotePubKey) != base64.StdEncoding.EncodeToString(f.agentPk) {
		t.Fatal("the parsed knock carries the wrong peer public key")
	}
	if err := s.dedupeRecvKNK(ppd, "192.0.2.10:5000"); err != nil {
		t.Fatalf("the legitimate knock was rejected: %v", err)
	}

	// The captured bytes are replayed from a DIFFERENT source address.
	replay := f.parseKnock(t, raw, "192.0.2.10:5000")
	if err := s.dedupeRecvKNK(replay, "198.51.100.20:5000"); !errors.Is(err, common.ErrServerDuplicateTransaction) {
		t.Fatalf("the replayed knock was not refused: %v", err)
	}

	// A legitimate retry — a fresh packet with a fresh counter and a fresh
	// timestamp — is accepted.
	retryRaw, _ := f.buildKnock(t)
	retry := f.parseKnock(t, retryRaw, "192.0.2.10:5000")
	if err := s.dedupeRecvKNK(retry, "192.0.2.10:5000"); err != nil {
		t.Fatalf("the legitimate retry was rejected: %v", err)
	}
}

// TestKNKReplayEndToEndIsIndependentOfSourceAddress pins the property the
// whole control rests on: the verdict does not depend on where the packet
// came from. The first arrival of the captured bytes is accepted wherever it
// lands; every repeat of the same bytes is refused, from every address.
func TestKNKReplayEndToEndIsIndependentOfSourceAddress(t *testing.T) {
	f := newKNKReplayFixture(t)
	s := &UdpServer{knkReplay: newKNKReplayCache()}

	raw, _ := f.buildKnock(t)

	// The first arrival, from the agent's own address.
	first := f.parseKnock(t, raw, "192.0.2.10:5000")
	if err := s.dedupeRecvKNK(first, "192.0.2.10:5000"); err != nil {
		t.Fatalf("the first knock was rejected: %v", err)
	}

	// Every repeat of the same bytes is refused, from every address —
	// including the agent's own.
	for _, src := range []string{"192.0.2.10:5000", "198.51.100.20:5000", "203.0.113.30:5000"} {
		replay := f.parseKnock(t, raw, src)
		if err := s.dedupeRecvKNK(replay, src); !errors.Is(err, common.ErrServerDuplicateTransaction) {
			t.Fatalf("the replay from %s was not refused: %v", src, err)
		}
	}
}

// TestKNKReplayEndToEndPerConnectionGateDoesNotHelpTheAttacker pins the
// gap the replay cache fills: the responder's per-connection monotonic gate
// lets the replay through, because a replay from a new source address lands
// on a fresh connection whose watermark is zero. The cache is what stops it.
func TestKNKReplayEndToEndPerConnectionGateDoesNotHelpTheAttacker(t *testing.T) {
	f := newKNKReplayFixture(t)
	s := &UdpServer{knkReplay: newKNKReplayCache()}

	raw, _ := f.buildKnock(t)

	// The legitimate knock on the agent's own connection.
	legit := f.parseKnock(t, raw, "192.0.2.10:5000")
	if err := s.dedupeRecvKNK(legit, "192.0.2.10:5000"); err != nil {
		t.Fatalf("the legitimate knock was rejected: %v", err)
	}

	// The replay lands on a DIFFERENT connection — a different source
	// address means a different ConnectionData, whose LastRemoteSendTime
	// is 0, so the monotonic gate in nhp/core/responder.go accepts it.
	// The replay cache is the only thing that stops it.
	replay := f.parseKnock(t, raw, "198.51.100.20:5000")
	if err := s.dedupeRecvKNK(replay, "198.51.100.20:5000"); !errors.Is(err, common.ErrServerDuplicateTransaction) {
		t.Fatalf("the replay on a fresh connection was not refused: %v", err)
	}
}
