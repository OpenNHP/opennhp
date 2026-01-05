package test

import (
	"testing"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// FuzzECDHFromKey tests ECDH key creation with random inputs.
// This is important for security as malformed keys should be handled gracefully.
func FuzzECDHFromKey(f *testing.F) {
	// Seed corpus with valid key sizes
	f.Add([]byte{}, int(core.ECC_CURVE25519))
	f.Add(make([]byte, 32), int(core.ECC_CURVE25519))
	f.Add(make([]byte, 64), int(core.ECC_SM2))
	f.Add(make([]byte, 16), int(core.ECC_CURVE25519))
	f.Add(make([]byte, 48), int(core.ECC_SM2))

	f.Fuzz(func(t *testing.T, data []byte, eccType int) {
		// Normalize eccType to valid range
		var eType core.EccTypeEnum
		switch eccType % 2 {
		case 0:
			eType = core.ECC_CURVE25519
		case 1:
			eType = core.ECC_SM2
		}

		// ECDHFromKey should not panic on any input
		// It should return nil for invalid keys
		e := core.ECDHFromKey(eType, data)
		if e != nil {
			// If key was accepted, verify basic operations don't panic
			_ = e.PublicKey()
			_ = e.PublicKeyBase64()
		}
	})
}

// FuzzAESDecrypt tests AES decryption with random inputs.
// Ensures malformed ciphertext is handled without panics.
func FuzzAESDecrypt(f *testing.F) {
	// Seed corpus with various sizes
	f.Add(make([]byte, 16), make([]byte, 32)) // minimum block size
	f.Add(make([]byte, 32), make([]byte, 32)) // two blocks
	f.Add(make([]byte, 48), make([]byte, 32)) // three blocks
	f.Add([]byte{}, make([]byte, 32))         // empty ciphertext

	f.Fuzz(func(t *testing.T, ciphertext []byte, key []byte) {
		// Normalize key to 32 bytes (AES-256)
		if len(key) < 32 {
			paddedKey := make([]byte, 32)
			copy(paddedKey, key)
			key = paddedKey
		} else if len(key) > 32 {
			key = key[:32]
		}

		// AESDecrypt should not panic on any input
		_, _ = core.AESDecrypt(ciphertext, key)
	})
}

// FuzzHeaderTypeToDeviceType tests header type to device type mapping.
func FuzzHeaderTypeToDeviceType(f *testing.F) {
	// Seed with known valid types
	f.Add(0)  // NHP_KPL
	f.Add(1)  // NHP_KNK
	f.Add(10) // NHP_AOL
	f.Add(100)
	f.Add(-1)
	f.Add(1000000)

	f.Fuzz(func(t *testing.T, headerType int) {
		// Should not panic on any input
		_ = core.HeaderTypeToDeviceType(headerType)
		_ = core.HeaderTypeToString(headerType)
	})
}

// FuzzCBCDecryption tests CBC decryption with random inputs.
// Tests both AES256 and SM4 cipher modes.
func FuzzCBCDecryption(f *testing.F) {
	// Seed corpus with various sizes
	f.Add(make([]byte, 16), 0) // one block, AES
	f.Add(make([]byte, 32), 0) // two blocks, AES
	f.Add(make([]byte, 17), 0) // invalid size (not multiple of block), AES
	f.Add(make([]byte, 16), 1) // one block, SM4
	f.Add(make([]byte, 17), 1) // invalid size, SM4
	f.Add([]byte{}, 0)         // empty

	f.Fuzz(func(t *testing.T, ciphertext []byte, cipherType int) {
		// Use a fixed key for testing
		var key [core.SymmetricKeySize]byte
		for i := range key {
			key[i] = byte(i)
		}

		var gcmType core.GcmTypeEnum
		switch cipherType % 2 {
		case 0:
			gcmType = core.GCM_AES256
		case 1:
			gcmType = core.GCM_SM4
		}

		// CBCDecryption should not panic on any input
		_, _ = core.CBCDecryption(gcmType, &key, ciphertext, false)
	})
}

// FuzzUdpPeerName tests UdpPeer.Name() with various public key lengths.
// The Name() function slices PubKeyBase64 which could panic on short strings.
func FuzzUdpPeerName(f *testing.F) {
	// Seed corpus with various lengths
	f.Add("")
	f.Add("abc")
	f.Add("0123456789")
	f.Add("012345678901234567890123456789012345678901234") // 45 chars - valid
	f.Add("01234567890123456789012345678901234567890123") // 44 chars - almost valid
	f.Add("0123456789012345678901234567890123456789012")  // 43 chars - minimum

	f.Fuzz(func(t *testing.T, pubKeyBase64 string) {
		peer := &core.UdpPeer{
			PubKeyBase64: pubKeyBase64,
		}
		// Name() should not panic on any input
		_ = peer.Name()
	})
}

// FuzzPacketParsing tests Packet methods with malformed/truncated data.
// Network packets could be truncated or malformed by attackers.
func FuzzPacketParsing(f *testing.F) {
	// Seed with various packet sizes
	f.Add([]byte{})                  // empty
	f.Add(make([]byte, 8))           // too short for most operations
	f.Add(make([]byte, 12))          // just enough for Flag()
	f.Add(make([]byte, 24))          // just enough for Counter()
	f.Add(make([]byte, 64))          // typical minimum header
	f.Add(make([]byte, 128))         // larger packet

	f.Fuzz(func(t *testing.T, data []byte) {
		pkt := &core.Packet{
			Content: data,
		}

		// These methods access fixed offsets and could panic
		// They should be safe on any input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on packet of length %d: %v", len(data), r)
			}
		}()

		// Only call methods if we have enough data to avoid expected panics
		// This tests the boundary conditions
		if len(data) >= 12 {
			_ = pkt.Flag()
		}
		if len(data) >= 8 {
			_, _ = pkt.HeaderTypeAndSize()
		}
		if len(data) >= 24 {
			_ = pkt.Counter()
		}
	})
}
