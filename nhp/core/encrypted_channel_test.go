package core

import (
	"net"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// TestMsgToPacketRoutine_WithPrevDeliversEncryptedPacketExactlyOnce fences
// response-side delivery through the caller-provided async channel. Before
// this contract, createMsgAssemblerData copied EncryptedPktCh only for new
// initiator packets; responses derived from PrevParserData silently fell
// through to the socket SendQueue.
func TestMsgToPacketRoutine_WithPrevDeliversEncryptedPacketExactlyOnce(t *testing.T) {
	key := make([]byte, PrivateKeySize)
	key[0] = 1
	dev := NewDevice(NHP_SERVER, key, nil)
	if dev == nil {
		t.Fatal("NewDevice returned nil")
	}

	conn := &ConnectionData{
		Device:               dev,
		LocalAddr:            &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 62206},
		RemoteAddr:           &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 40000},
		CookieStore:          &CookieStore{},
		SendQueue:            make(chan *Packet, 1),
		RecvQueue:            make(chan *Packet, 1),
		BlockSignal:          make(chan struct{}, 1),
		SetTimeoutSignal:     make(chan struct{}, 1),
		StopSignal:           make(chan struct{}),
		RemoteTransactionMap: make(map[uint64]*RemoteTransaction),
	}
	t.Cleanup(func() {
		select {
		case packet := <-conn.SendQueue:
			dev.ReleasePoolPacket(packet)
		default:
		}
	})
	dev.Start()
	t.Cleanup(dev.Stop)

	encryptedPktCh := make(chan *MsgAssemblerData, 2)
	dev.SendMsgToPacket(&MsgData{
		HeaderType: NHP_ACK,
		PrevParserData: &PacketParserData{
			device:       dev,
			ConnData:     conn,
			CipherScheme: common.CIPHER_SCHEME_CURVE,
			Ciphers:      NewCipherSuite(common.CIPHER_SCHEME_CURVE),
			RemotePubKey: testPeerPk(),
			SenderTrxId:  1,
		},
		Message:        []byte(`{"errCode":"0"}`),
		EncryptedPktCh: encryptedPktCh,
	})

	var mad *MsgAssemblerData
	select {
	case mad = <-encryptedPktCh:
	case <-time.After(2 * time.Second):
		t.Fatal("response packet fell through instead of reaching EncryptedPktCh")
	}
	if mad == nil || mad.Error != nil || mad.BasePacket == nil || len(mad.BasePacket.Content) == 0 {
		t.Fatalf("encrypted response = %#v, want one complete packet", mad)
	}
	mad.Destroy()

	select {
	case duplicate := <-encryptedPktCh:
		if duplicate != nil && duplicate != mad {
			duplicate.Destroy()
		}
		t.Fatalf("second write on EncryptedPktCh: %#v", duplicate)
	case <-time.After(200 * time.Millisecond):
	}
	select {
	case packet := <-conn.SendQueue:
		dev.ReleasePoolPacket(packet)
		t.Fatal("response packet also fell through to the socket SendQueue")
	default:
	}
}
