package core

import (
	"errors"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// testPeerPk returns the deterministic byte-pattern public key the
// chain-key tests use as a stand-in. Not a valid curve25519 key —
// callers are expected to be on paths where peer-key validation
// doesn't fire (e.g., setPeerPublicKey accepts arbitrary
// PublicKeySize-byte inputs; signing/AEAD steps that would reject
// are gated behind earlier checks the tests don't reach).
func testPeerPk() []byte {
	peerPk := make([]byte, PublicKeySize)
	for i := range peerPk {
		peerPk[i] = byte(i + 1)
	}
	return peerPk
}

// runResponderWithPrevHMACFailure builds a prev MAD and a junk
// packet, calls createPacketParserData with PrevAssemblerData set
// plus any extra config the caller applies via `configure`, asserts
// the HMAC-failure / bare-return contract (err is ErrHMACCheckFailed,
// ppd is non-nil), and returns the populated ppd for further
// inspection. Shared by the responder-side chain-key tests in
// derive_chainkey_test.go.
func runResponderWithPrevHMACFailure(t *testing.T, dev *Device, configure func(*PacketData)) *PacketParserData {
	t.Helper()

	prevMad, err := dev.createMsgAssemblerData(&MsgData{
		HeaderType:    NHP_KNK,
		CipherScheme:  common.CIPHER_SCHEME_CURVE,
		TransactionId: 1,
		PeerPk:        testPeerPk(),
	})
	if err != nil {
		t.Fatalf("build prev MAD: %v", err)
	}
	t.Cleanup(prevMad.Destroy)

	pkt := dev.AllocatePoolPacket()
	if pkt == nil {
		t.Fatal("AllocatePoolPacket returned nil")
	}
	pkt.Content = pkt.Buf[:prevMad.header.Size()]

	pd := &PacketData{
		BasePacket:        pkt,
		PrevAssemblerData: prevMad,
		InitTime:          time.Now().UnixNano(),
	}
	if configure != nil {
		configure(pd)
	}

	ppd, err := dev.createPacketParserData(pd)
	if !errors.Is(err, ErrHMACCheckFailed) {
		t.Fatalf("expected ErrHMACCheckFailed, got %v", err)
	}
	if ppd == nil {
		t.Fatal("ppd nil — createPacketParserData did not populate named return on err path")
	}
	t.Cleanup(ppd.Destroy)

	return ppd
}
