package core

import (
	"testing"
	"time"
)

func TestPeerRoleAllowed(t *testing.T) {
	for _, tc := range []struct {
		expected, actual int
		want             bool
	}{
		{NHP_DB, NHP_DB, true},
		{NHP_DB, NHP_AGENT, false},
		{NHP_DB, NHP_AC, false},
		{NHP_AC, NHP_DB, false},
		{NHP_SERVER, NHP_DB, false},
		{NHP_RELAY, NHP_AGENT, false},
		// NHP and DHP agents share the agent peer table.
		{DHP_AGENT, NHP_AGENT, true},
		{NHP_AGENT, DHP_AGENT, true},
		{DHP_AGENT, NHP_DB, false},
		// Untyped peers are not checked.
		{NHP_DB, NHP_NO_DEVICE, true},
	} {
		if got := peerRoleAllowed(tc.expected, tc.actual); got != tc.want {
			t.Errorf("peerRoleAllowed(%s, %s) = %v, want %v",
				DeviceTypeToString(tc.expected), DeviceTypeToString(tc.actual), got, tc.want)
		}
	}
}

// A key registered in one role cannot send messages that belong to
// another: an agent key in the server's pool must not pass validatePeer
// for NHP_DRG, which only an nhp-db may send.
func TestValidatePeerBindsRole(t *testing.T) {
	newFixture := func(t *testing.T, senderType, headerType int, registeredAs int) func() error {
		t.Helper()
		sender := NewDevice(senderType, sentinelPrivateKey(1), nil)
		receiver := NewDevice(NHP_SERVER, sentinelPrivateKey(33), nil)
		if sender == nil || receiver == nil {
			t.Fatal("failed to create devices")
		}
		t.Cleanup(sender.Stop)
		t.Cleanup(receiver.Stop)

		receiverPeer := &UdpPeer{PubKeyBase64: receiver.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12346, Type: NHP_SERVER}
		sender.AddPeer(receiverPeer)
		receiver.AddPeer(&UdpPeer{PubKeyBase64: sender.PublicKeyBase64(), Ip: "127.0.0.1", Port: 12345, Type: registeredAs})

		mad, err := sender.MsgToPacket(&MsgData{
			ConnData:      sentinelConnection(sender, 12345, 12346),
			PeerPk:        receiverPeer.PublicKey(),
			HeaderType:    headerType,
			TransactionId: 1,
		})
		if err != nil {
			t.Fatalf("MsgToPacket(%s) failed: %v", HeaderTypeToString(headerType), err)
		}
		content := append([]byte(nil), mad.BasePacket.Content...)
		receiverConn := sentinelConnection(receiver, 12346, 12345)
		return func() error {
			ppd, err := receiver.createPacketParserData(&PacketData{
				BasePacket: &Packet{Content: append([]byte(nil), content...), HeaderType: headerType},
				ConnData:   receiverConn,
				InitTime:   time.Now().UnixNano(),
			})
			if err != nil {
				t.Fatalf("createPacketParserData failed: %v", err)
			}
			defer ppd.Destroy()
			return ppd.validatePeer()
		}
	}

	for _, tc := range []struct {
		name                   string
		senderType, headerType int
		registeredAs           int
		want                   error
	}{
		{"agent key sends DRG", NHP_DB, NHP_DRG, NHP_AGENT, ErrPeerRoleMismatch},
		{"AC key sends DRG", NHP_DB, NHP_DRG, NHP_AC, ErrPeerRoleMismatch},
		{"db key sends DRG", NHP_DB, NHP_DRG, NHP_DB, nil},
		{"untyped key sends DRG", NHP_DB, NHP_DRG, NHP_NO_DEVICE, nil},
		{"db key sends AOL", NHP_AC, NHP_AOL, NHP_DB, ErrPeerRoleMismatch},
		{"ac key sends AOL", NHP_AC, NHP_AOL, NHP_AC, nil},
		{"agent key sends DHP knock", DHP_AGENT, DHP_KNK, NHP_AGENT, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			validate := newFixture(t, tc.senderType, tc.headerType, tc.registeredAs)
			if got := validate(); got != tc.want {
				t.Fatalf("validatePeer = %v, want %v", got, tc.want)
			}
		})
	}
}

// The pool holds one peer per key: a key added in a second role replaces
// the first (AddPeer warns about it), so it keeps only the role added last.
func TestAddPeerKeyHoldsOneRole(t *testing.T) {
	d := NewDevice(NHP_SERVER, sentinelPrivateKey(33), nil)
	if d == nil {
		t.Fatal("NewDevice failed")
	}
	t.Cleanup(d.Stop)
	key := NewECDH(ECC_CURVE25519).PublicKeyBase64()
	d.AddPeer(&UdpPeer{PubKeyBase64: key, Type: NHP_AGENT})
	d.AddPeer(&UdpPeer{PubKeyBase64: key, Type: NHP_DB})
	pk := (&UdpPeer{PubKeyBase64: key}).PublicKey()
	if got := d.LookupPeer(pk).DeviceType(); got != NHP_DB {
		t.Fatalf("key kept role %s, want the last one added (%s)", DeviceTypeToString(got), DeviceTypeToString(NHP_DB))
	}
	if peerRoleAllowed(NHP_AGENT, d.LookupPeer(pk).DeviceType()) {
		t.Fatal("the replaced agent role is still accepted")
	}
}
