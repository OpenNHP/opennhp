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

func TestShouldEscalateReplay(t *testing.T) {
	tests := []struct {
		name                          string
		deviceType, peerType, msgType int
		want                          bool
	}{
		{"AOP is drop-only", NHP_AC, NHP_SERVER, NHP_AOP, false},
		{"ART unchanged", NHP_SERVER, NHP_AC, NHP_ART, true},
		{"AOP wrong scope", NHP_SERVER, NHP_SERVER, NHP_AOP, true},
		{"default", NHP_SERVER, NHP_AGENT, NHP_KNK, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldEscalateReplay(tt.deviceType, tt.peerType, tt.msgType); got != tt.want {
				t.Fatalf("shouldEscalateReplay() = %v, want %v", got, tt.want)
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
		{"default", NHP_SERVER, NHP_AGENT, NHP_KNK, true},
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
	defaultFloor := recvStalenessFloor(NHP_SERVER, NHP_AGENT, NHP_KNK)
	if want := DefaultRecvStalenessFloorSeconds * int64(time.Second); defaultFloor != want {
		t.Fatalf("default floor = %d, want %d", defaultFloor, want)
	}
	if aop >= defaultFloor {
		t.Fatalf("AOP floor (%d) must be tighter than default (%d)", aop, defaultFloor)
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
}
