package core

import (
	"net"
	"sync/atomic"
	"testing"
	"time"
)

type validatePeerSentinelFixture struct {
	receiver     *Device
	senderPeer   *UdpPeer
	receiverConn *ConnectionData
	packet       []byte
	sendTime     int64
	initTime     int64
}

func newValidatePeerSentinelFixture(t *testing.T) *validatePeerSentinelFixture {
	t.Helper()

	sender := NewDevice(NHP_AC, sentinelPrivateKey(1), nil)
	receiver := NewDevice(NHP_SERVER, sentinelPrivateKey(33), nil)
	if sender == nil || receiver == nil {
		t.Fatal("failed to create AC/server devices")
	}
	t.Cleanup(sender.Stop)
	t.Cleanup(receiver.Stop)

	receiverPeer := &UdpPeer{
		PubKeyBase64: receiver.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         12346,
		Type:         NHP_SERVER,
	}
	sender.AddPeer(receiverPeer)
	senderPeer := &UdpPeer{
		PubKeyBase64: sender.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         12345,
		Type:         NHP_AC,
	}

	senderConn := sentinelConnection(sender, 12345, 12346)
	mad, err := sender.MsgToPacket(&MsgData{
		ConnData:      senderConn,
		PeerPk:        receiverPeer.PublicKey(),
		HeaderType:    NHP_AOL,
		TransactionId: 1,
	})
	if err != nil {
		t.Fatalf("MsgToPacket failed: %v", err)
	}

	return &validatePeerSentinelFixture{
		receiver:     receiver,
		senderPeer:   senderPeer,
		receiverConn: sentinelConnection(receiver, 12346, 12345),
		packet:       append([]byte(nil), mad.BasePacket.Content...),
		sendTime:     mad.LocalInitTime,
		initTime:     time.Now().UnixNano(),
	}
}

func (f *validatePeerSentinelFixture) validate(t *testing.T) error {
	t.Helper()
	packet := &Packet{Content: append([]byte(nil), f.packet...), HeaderType: NHP_AOL}
	ppd, err := f.receiver.createPacketParserData(&PacketData{
		BasePacket: packet,
		ConnData:   f.receiverConn,
		InitTime:   f.initTime,
	})
	if err != nil {
		t.Fatalf("createPacketParserData failed: %v", err)
	}
	defer ppd.Destroy()
	return ppd.validatePeer()
}

func TestValidatePeerReturnsSentinelErrors(t *testing.T) {
	t.Run("peer not found", func(t *testing.T) {
		fixture := newValidatePeerSentinelFixture(t)
		if got := fixture.validate(t); got != ErrPeerNotFound {
			t.Fatalf("validatePeer error = %v, want ErrPeerNotFound", got)
		}
	})

	t.Run("peer expired", func(t *testing.T) {
		fixture := newValidatePeerSentinelFixture(t)
		fixture.senderPeer.ExpireTime = time.Now().Add(-time.Minute).Unix()
		fixture.receiver.AddPeer(fixture.senderPeer)
		if got := fixture.validate(t); got != ErrPeerExpired {
			t.Fatalf("validatePeer error = %v, want ErrPeerExpired", got)
		}
	})

	t.Run("peer address mismatch", func(t *testing.T) {
		fixture := newValidatePeerSentinelFixture(t)
		fixture.receiver.AddPeer(fixture.senderPeer)
		fixture.receiverConn.UpdateRecvAddress(fixture.initTime, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 54321})
		if got := fixture.validate(t); got != ErrPeerAddressMismatch {
			t.Fatalf("validatePeer error = %v, want ErrPeerAddressMismatch", got)
		}
	})

	t.Run("replay packet", func(t *testing.T) {
		fixture := newValidatePeerSentinelFixture(t)
		fixture.receiver.AddPeer(fixture.senderPeer)
		atomic.StoreInt64(&fixture.receiverConn.LastRemoteSendTime, fixture.sendTime+1)
		if got := fixture.validate(t); got != ErrReplayPacketReceived {
			t.Fatalf("validatePeer error = %v, want ErrReplayPacketReceived", got)
		}
	})

	t.Run("flood packet", func(t *testing.T) {
		fixture := newValidatePeerSentinelFixture(t)
		fixture.receiver.AddPeer(fixture.senderPeer)
		atomic.StoreInt64(&fixture.receiverConn.LastRemoteSendTime, fixture.sendTime-int64(time.Millisecond))
		if got := fixture.validate(t); got != ErrFloodPacketReceived {
			t.Fatalf("validatePeer error = %v, want ErrFloodPacketReceived", got)
		}
	})

	t.Run("stale packet", func(t *testing.T) {
		fixture := newValidatePeerSentinelFixture(t)
		fixture.receiver.AddPeer(fixture.senderPeer)
		fixture.initTime = fixture.sendTime + int64(601*time.Second)
		if got := fixture.validate(t); got != ErrStalePacketReceived {
			t.Fatalf("validatePeer error = %v, want ErrStalePacketReceived", got)
		}
	})
}

func sentinelPrivateKey(start byte) []byte {
	key := make([]byte, PrivateKeySize)
	for i := range key {
		key[i] = start + byte(i)
	}
	return key
}

func sentinelConnection(device *Device, localPort, remotePort int) *ConnectionData {
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
