package core

import (
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// TestDerivePacketParserData_DoesNotCarryChainKey and its
// responder-side twin fence the chain-key fix: the derive functions
// must never copy the previous transaction's chainKey into the new
// MAD/PPD. The intermediate-chain-key carry-over design was
// abandoned; encryptBody/decryptBody defer-zero the chainKey on the
// way out, so any surviving copy() in derive* picks up a zeroed
// buffer in Go-Go (which both ends symmetrically "agree on") but
// breaks JS-Go interop because a spec-following JS implementation
// does not replicate the zero-carry-over quirk.
//
// Asymmetric assertion: the parent (MAD/PPD) is seeded with a
// non-zero sentinel; the child must come out zero. The newly
// allocated child struct starts zero by default, so this only
// fails if derive* *actively* copies the sentinel — which is
// exactly the regression we want to fence. Do not "fix" the
// asymmetry by also seeding the child; the test would then no
// longer distinguish "derive copied" from "derive did nothing".
func TestDerivePacketParserData_DoesNotCarryChainKey(t *testing.T) {
	dev := newDeviceForChainKeyTest(t)

	mad := &MsgAssemblerData{
		device:       dev,
		CipherScheme: common.CIPHER_SCHEME_CURVE,
		ciphers:      NewCipherSuite(common.CIPHER_SCHEME_CURVE),
	}
	for i := range mad.chainKey {
		mad.chainKey[i] = 0xAB
	}

	pkt := dev.AllocatePoolPacket()
	if pkt == nil {
		t.Fatal("AllocatePoolPacket returned nil")
	}
	t.Cleanup(func() { dev.ReleasePoolPacket(pkt) })

	ppd := mad.derivePacketParserData(pkt, time.Now().UnixNano())

	var zero [SymmetricKeySize]byte
	if ppd.chainKey != zero {
		t.Errorf("derivePacketParserData copied chainKey from parent MAD — chain-key fix regression. ppd.chainKey=%x", ppd.chainKey)
	}
}

func TestDeriveMsgAssemblerData_DoesNotCarryChainKey(t *testing.T) {
	dev := newDeviceForChainKeyTest(t)

	ppd := &PacketParserData{
		device:       dev,
		CipherScheme: common.CIPHER_SCHEME_CURVE,
		Ciphers:      NewCipherSuite(common.CIPHER_SCHEME_CURVE),
	}
	for i := range ppd.chainKey {
		ppd.chainKey[i] = 0xCD
	}

	mad := ppd.deriveMsgAssemblerData(NHP_ACK, false, nil)
	if mad == nil {
		t.Fatal("deriveMsgAssemblerData returned nil")
	}
	t.Cleanup(func() { mad.Destroy() })

	var zero [SymmetricKeySize]byte
	if mad.chainKey != zero {
		t.Errorf("deriveMsgAssemblerData copied chainKey from parent PPD — chain-key fix regression. mad.chainKey=%x", mad.chainKey)
	}
}

// TestCreateMsgAssemblerData_InitsCanonicalChainKey fences the
// other regression mode: deleting (or skipping) the always-run
// chain-key init block in createMsgAssemblerData. Both branches
// (no-prev / with-prev) must produce the canonical ChainKey0 (a
// non-zero, deterministic value), and they must agree — pre-fix
// code had the with-prev branch carrying over a zeroed chainKey
// from the prior transaction. Companion to the derive-level tests
// above (the derive tests catch a reintroduced copy(); this test
// catches a deleted init block).
func TestCreateMsgAssemblerData_InitsCanonicalChainKey(t *testing.T) {
	dev := newDeviceForChainKeyTest(t)

	peerPk := testPeerPk()

	madNoPrev, err := dev.createMsgAssemblerData(&MsgData{
		HeaderType:    NHP_KNK,
		CipherScheme:  common.CIPHER_SCHEME_CURVE,
		TransactionId: 1,
		PeerPk:        peerPk,
	})
	if err != nil {
		t.Fatalf("no-prev createMsgAssemblerData: %v", err)
	}
	t.Cleanup(madNoPrev.Destroy)

	madWithPrev, err := dev.createMsgAssemblerData(&MsgData{
		HeaderType: NHP_ACK,
		PrevParserData: &PacketParserData{
			device:       dev,
			CipherScheme: common.CIPHER_SCHEME_CURVE,
			Ciphers:      NewCipherSuite(common.CIPHER_SCHEME_CURVE),
			RemotePubKey: peerPk,
			SenderTrxId:  1,
		},
	})
	if err != nil {
		t.Fatalf("with-prev createMsgAssemblerData: %v", err)
	}
	t.Cleanup(madWithPrev.Destroy)

	var zero [SymmetricKeySize]byte
	if madNoPrev.chainKey == zero {
		t.Error("no-prev chainKey is zero — canonical ChainKey0 was not initialized")
	}
	if madWithPrev.chainKey == zero {
		t.Error("with-prev chainKey is zero — chain-key fix regression (always-run init block did not run on the carry-over path)")
	}
	if madNoPrev.chainKey != madWithPrev.chainKey {
		t.Errorf("chainKey diverges between branches; both must produce canonical ChainKey0.\nno-prev:   %x\nwith-prev: %x", madNoPrev.chainKey, madWithPrev.chainKey)
	}
}

// TestCreatePacketParserData_InitsCanonicalChainKey is the responder-
// side twin of TestCreateMsgAssemblerData_InitsCanonicalChainKey.
// The packet has no valid HMAC so the function returns
// ErrHMACCheckFailed — but chain-key init runs BEFORE HMAC
// validation (see createPacketParserData), and the function uses
// named returns so ppd is populated even on the HMAC failure path.
func TestCreatePacketParserData_InitsCanonicalChainKey(t *testing.T) {
	// runResponderWithPrevHMACFailure pins the
	// errors.Is(err, ErrHMACCheckFailed) + ppd != nil contract; we
	// inspect chainKey on the returned ppd to confirm the init block
	// ran above the HMAC check.
	ppd := runResponderWithPrevHMACFailure(t, newDeviceForChainKeyTest(t), nil)

	var zero [SymmetricKeySize]byte
	if ppd.chainKey == zero {
		t.Error("ppd.chainKey is zero — chain-key fix regression (always-run init block did not run on the carry-over path)")
	}
}

func newDeviceForChainKeyTest(t *testing.T) *Device {
	t.Helper()
	priv := make([]byte, 32)
	for i := range priv {
		priv[i] = byte(i + 1)
	}
	dev := NewDevice(NHP_AGENT, priv, nil)
	if dev == nil {
		t.Fatal("NewDevice returned nil")
	}
	t.Cleanup(dev.Stop)
	return dev
}
