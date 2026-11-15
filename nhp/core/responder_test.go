package core

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestShouldCheckRecvAttack(t *testing.T) {
	tests := []struct {
		name                          string
		deviceType, peerType, msgType int
		want                          bool
	}{
		{"AOP permits reordering", NHP_AC, NHP_SERVER, NHP_AOP, false},
		{"ART remains exempt", NHP_SERVER, NHP_AC, NHP_ART, false},
		{"default", NHP_SERVER, NHP_AGENT, NHP_KNK, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCheckRecvAttack(tt.deviceType, tt.peerType, tt.msgType); got != tt.want {
				t.Fatalf("shouldCheckRecvAttack() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldCheckFlood(t *testing.T) {
	tests := []struct {
		name                          string
		deviceType, peerType, msgType int
		want                          bool
	}{
		{"AOP burst exempt", NHP_AC, NHP_SERVER, NHP_AOP, false},
		{"ART remains exempt", NHP_SERVER, NHP_AC, NHP_ART, false},
		{"default", NHP_SERVER, NHP_AGENT, NHP_KNK, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldCheckFlood(tt.deviceType, tt.peerType, tt.msgType); got != tt.want {
				t.Fatalf("shouldCheckFlood() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldEscalateStale(t *testing.T) {
	tests := []struct {
		name                          string
		deviceType, peerType, msgType int
		want                          bool
	}{
		{"AOP is drop-only", NHP_AC, NHP_SERVER, NHP_AOP, false},
		{"ART unchanged", NHP_SERVER, NHP_AC, NHP_ART, true},
		{"AOP wrong scope", NHP_SERVER, NHP_SERVER, NHP_AOP, true},
		// The agent knock family has a 60 s staleness floor, so a
		// clock-skew false-reject must stay a drop: escalating it would
		// block the agent's source address.
		{"KNK is drop-only", NHP_SERVER, NHP_AGENT, NHP_KNK, false},
		{"RKN is drop-only", NHP_SERVER, NHP_AGENT, NHP_RKN, false},
		{"EXT is drop-only", NHP_SERVER, NHP_AGENT, NHP_EXT, false},
		{"DHP KNK is drop-only", NHP_SERVER, DHP_AGENT, DHP_KNK, false},
		{"agent LST unchanged", NHP_SERVER, NHP_AGENT, NHP_LST, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldEscalateStale(tt.deviceType, tt.peerType, tt.msgType); got != tt.want {
				t.Fatalf("shouldEscalateStale() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecvStalenessFloor(t *testing.T) {
	aop := recvStalenessFloor(NHP_AC, NHP_SERVER, NHP_AOP)
	if want := AOPRecvStalenessFloorSeconds * int64(time.Second); aop != want {
		t.Fatalf("AOP floor = %d, want %d", aop, want)
	}
	knk := recvStalenessFloor(NHP_SERVER, NHP_AGENT, NHP_KNK)
	if want := KNKRecvStalenessFloorSeconds * int64(time.Second); knk != want {
		t.Fatalf("KNK floor = %d, want %d", knk, want)
	}
	defaultFloor := recvStalenessFloor(NHP_SERVER, NHP_AC, NHP_AOL)
	if want := DefaultRecvStalenessFloorSeconds * int64(time.Second); defaultFloor != want {
		t.Fatalf("default floor = %d, want %d", defaultFloor, want)
	}
	if aop >= defaultFloor {
		t.Fatalf("AOP floor (%d) must be tighter than default (%d)", aop, defaultFloor)
	}
	if knk >= defaultFloor {
		t.Fatalf("KNK floor (%d) must be tighter than default (%d)", knk, defaultFloor)
	}
}

// TestKNKRecvStalenessFloor covers the whole agent knock family: the
// tighter floor applies to KNK, RKN, EXT and the DHP knock, and to nothing
// else.
func TestKNKRecvStalenessFloor(t *testing.T) {
	knkFloor := KNKRecvStalenessFloorSeconds * int64(time.Second)
	for _, tc := range []struct {
		deviceType, peerType, msgType int
		want                          int64
	}{
		{NHP_SERVER, NHP_AGENT, NHP_KNK, knkFloor},
		{NHP_SERVER, NHP_AGENT, NHP_RKN, knkFloor},
		{NHP_SERVER, NHP_AGENT, NHP_EXT, knkFloor},
		{NHP_SERVER, DHP_AGENT, DHP_KNK, knkFloor},
		// An agent's other messages keep the generous default.
		{NHP_SERVER, NHP_AGENT, NHP_LST, DefaultRecvStalenessFloorSeconds * int64(time.Second)},
		{NHP_SERVER, NHP_AGENT, NHP_REG, DefaultRecvStalenessFloorSeconds * int64(time.Second)},
		// The AC side is untouched.
		{NHP_AC, NHP_SERVER, NHP_AOP, AOPRecvStalenessFloorSeconds * int64(time.Second)},
	} {
		got := recvStalenessFloor(tc.deviceType, tc.peerType, tc.msgType)
		if got != tc.want {
			t.Fatalf("recvStalenessFloor(%d,%d,%d) = %d, want %d", tc.deviceType, tc.peerType, tc.msgType, got, tc.want)
		}
	}
}

func TestAuthenticatedAOPAllowsTimestampReordering(t *testing.T) {
	sender := NewDevice(NHP_SERVER, sentinelPrivateKey(1), nil)
	receiver := NewDevice(NHP_AC, sentinelPrivateKey(33), nil)
	t.Cleanup(sender.Stop)
	t.Cleanup(receiver.Stop)
	receiverPeer := &UdpPeer{PubKeyBase64: receiver.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12346, Type: NHP_AC}
	sender.AddPeer(receiverPeer)
	receiver.AddPeer(&UdpPeer{PubKeyBase64: sender.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12345, Type: NHP_SERVER})
	mad, err := sender.MsgToPacket(&MsgData{ConnData: sentinelConnection(sender, 12345, 12346), PeerPk: receiverPeer.PublicKey(), HeaderType: NHP_AOP, TransactionId: 1})
	if err != nil {
		t.Fatal(err)
	}
	conn := sentinelConnection(receiver, 12346, 12345)
	atomic.StoreInt64(&conn.LastRemoteSendTime, mad.LocalInitTime+1)
	ppd, err := receiver.createPacketParserData(&PacketData{BasePacket: &Packet{Content: append([]byte(nil), mad.BasePacket.Content...), HeaderType: NHP_AOP}, ConnData: conn, InitTime: time.Now().UnixNano()})
	if err != nil {
		t.Fatal(err)
	}
	defer ppd.Destroy()
	if err := ppd.validatePeer(); err != nil {
		t.Fatalf("reordered authenticated AOP rejected: %v", err)
	}
	if ppd.RemoteSendTime != mad.LocalInitTime {
		t.Fatal("authenticated timestamp not exposed")
	}
}

func TestConfiguredAOPStaleness(t *testing.T) {
	sender := NewDevice(NHP_SERVER, sentinelPrivateKey(1), nil)
	receiver := NewDevice(NHP_AC, sentinelPrivateKey(33), nil)
	t.Cleanup(sender.Stop)
	t.Cleanup(receiver.Stop)
	peer := &UdpPeer{PubKeyBase64: receiver.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12346, Type: NHP_AC}
	sender.AddPeer(peer)
	receiver.AddPeer(&UdpPeer{PubKeyBase64: sender.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12345, Type: NHP_SERVER})
	mad, err := sender.MsgToPacket(&MsgData{ConnData: sentinelConnection(sender, 12345, 12346), PeerPk: peer.PublicKey(), HeaderType: NHP_AOP, TransactionId: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, limit := range []int{0, 300} {
		options := receiver.GetOption()
		options.AOPRecvStalenessSeconds = limit
		receiver.SetOption(options)
		ppd, err := receiver.createPacketParserData(&PacketData{BasePacket: &Packet{Content: append([]byte(nil), mad.BasePacket.Content...), HeaderType: NHP_AOP}, ConnData: sentinelConnection(receiver, 12346, 12345), InitTime: mad.LocalInitTime + int64(180*time.Second)})
		if err != nil {
			t.Fatal(err)
		}
		got := ppd.validatePeer()
		ppd.Destroy()
		if limit == 0 && got != ErrStalePacketReceived {
			t.Fatalf("default accepted old AOP: %v", got)
		}
		if limit == 300 && got != nil {
			t.Fatalf("configured age rejected: %v", got)
		}
	}
}

// TestConfiguredKNKStaleness verifies that DeviceOptions.KNKRecvStalenessSeconds
// overrides the protocol default for the agent knock family, and that the
// default (zero) keeps the tighter KNK floor.
func TestConfiguredKNKStaleness(t *testing.T) {
	sender := NewDevice(NHP_AGENT, sentinelPrivateKey(1), nil)
	receiver := NewDevice(NHP_SERVER, sentinelPrivateKey(33), nil)
	t.Cleanup(sender.Stop)
	t.Cleanup(receiver.Stop)
	peer := &UdpPeer{PubKeyBase64: receiver.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12346, Type: NHP_SERVER}
	sender.AddPeer(peer)
	receiver.AddPeer(&UdpPeer{PubKeyBase64: sender.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12345, Type: NHP_AGENT})
	mad, err := sender.MsgToPacket(&MsgData{ConnData: sentinelConnection(sender, 12345, 12346), PeerPk: peer.PublicKey(), HeaderType: NHP_KNK, TransactionId: 1})
	if err != nil {
		t.Fatal(err)
	}
	// The knock arrives 90 seconds late: older than the 60 s default floor
	// but inside a 300 s configured floor.
	for _, limit := range []int{0, 300} {
		options := receiver.GetOption()
		options.KNKRecvStalenessSeconds = limit
		receiver.SetOption(options)
		ppd, err := receiver.createPacketParserData(&PacketData{BasePacket: &Packet{Content: append([]byte(nil), mad.BasePacket.Content...), HeaderType: NHP_KNK}, ConnData: sentinelConnection(receiver, 12346, 12345), InitTime: mad.LocalInitTime + int64(90*time.Second)})
		if err != nil {
			t.Fatal(err)
		}
		got := ppd.validatePeer()
		ppd.Destroy()
		if limit == 0 && got != ErrStalePacketReceived {
			t.Fatalf("default accepted a 90 s old knock: %v", got)
		}
		if limit == 300 && got != nil {
			t.Fatalf("configured age rejected: %v", got)
		}
	}
}

// TestForwardDatedKNKRejected verifies that the agent knock family is
// bounded by the same forward-skew limit as the other cached message
// types. The staleness floor only rejects packets that are too OLD, so
// without this gate a forward-stamped knock would stay acceptable — and
// stay in the KNK replay cache — for as long as the sender's clock is
// ahead.
func TestForwardDatedKNKRejected(t *testing.T) {
	sender := NewDevice(NHP_AGENT, sentinelPrivateKey(1), nil)
	receiver := NewDevice(NHP_SERVER, sentinelPrivateKey(33), nil)
	t.Cleanup(sender.Stop)
	t.Cleanup(receiver.Stop)
	peer := &UdpPeer{PubKeyBase64: receiver.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12346, Type: NHP_SERVER}
	sender.AddPeer(peer)
	receiver.AddPeer(&UdpPeer{PubKeyBase64: sender.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12345, Type: NHP_AGENT})
	mad, err := sender.MsgToPacket(&MsgData{ConnData: sentinelConnection(sender, 12345, 12346), PeerPk: peer.PublicKey(), HeaderType: NHP_KNK, TransactionId: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		skew time.Duration
		want error
	}{
		{(ARTRecvFutureSkewSeconds - 1) * time.Second, nil},
		{(ARTRecvFutureSkewSeconds + 1) * time.Second, ErrStalePacketReceived},
	} {
		ppd, err := receiver.createPacketParserData(&PacketData{BasePacket: &Packet{Content: append([]byte(nil), mad.BasePacket.Content...), HeaderType: NHP_KNK}, ConnData: sentinelConnection(receiver, 12346, 12345), InitTime: mad.LocalInitTime - int64(tc.skew)})
		if err != nil {
			t.Fatal(err)
		}
		got := ppd.validatePeer()
		ppd.Destroy()
		if got != tc.want {
			t.Fatalf("KNK %v ahead: got %v, want %v", tc.skew, got, tc.want)
		}
	}
}

func TestForwardDatedAOPRejected(t *testing.T) {
	sender := NewDevice(NHP_SERVER, sentinelPrivateKey(1), nil)
	receiver := NewDevice(NHP_AC, sentinelPrivateKey(33), nil)
	t.Cleanup(sender.Stop)
	t.Cleanup(receiver.Stop)
	peer := &UdpPeer{PubKeyBase64: receiver.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12346, Type: NHP_AC}
	sender.AddPeer(peer)
	receiver.AddPeer(&UdpPeer{PubKeyBase64: sender.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12345, Type: NHP_SERVER})
	mad, err := sender.MsgToPacket(&MsgData{ConnData: sentinelConnection(sender, 12345, 12346), PeerPk: peer.PublicKey(), HeaderType: NHP_AOP, TransactionId: 1})
	if err != nil {
		t.Fatal(err)
	}
	// The AC receives the packet as though the server's clock ran ahead by
	// skew: just inside the bound passes, just past it is rejected.
	for _, tc := range []struct {
		skew time.Duration
		want error
	}{
		{(ARTRecvFutureSkewSeconds - 1) * time.Second, nil},
		{(ARTRecvFutureSkewSeconds + 1) * time.Second, ErrStalePacketReceived},
	} {
		ppd, err := receiver.createPacketParserData(&PacketData{BasePacket: &Packet{Content: append([]byte(nil), mad.BasePacket.Content...), HeaderType: NHP_AOP}, ConnData: sentinelConnection(receiver, 12346, 12345), InitTime: mad.LocalInitTime - int64(tc.skew)})
		if err != nil {
			t.Fatal(err)
		}
		got := ppd.validatePeer()
		ppd.Destroy()
		if got != tc.want {
			t.Fatalf("AOP %v ahead: got %v, want %v", tc.skew, got, tc.want)
		}
	}
}
