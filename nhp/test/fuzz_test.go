package test

import (
	"testing"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/core/scheme/gmsm"
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

// FuzzCBCDecryption tests CBC mode decryption with PKCS7 unpadding.
// Critical for security - malformed padding could cause panics or oracle attacks.
func FuzzCBCDecryption(f *testing.F) {
	// Seed with various block-aligned sizes
	f.Add(make([]byte, 16), int(core.GCM_AES256))
	f.Add(make([]byte, 32), int(core.GCM_AES256))
	f.Add(make([]byte, 16), int(core.GCM_SM4))
	f.Add(make([]byte, 48), int(core.GCM_SM4))
	f.Add([]byte{}, int(core.GCM_AES256))

	f.Fuzz(func(t *testing.T, ciphertext []byte, gcmType int) {
		// Create a valid key
		var key [core.SymmetricKeySize]byte
		copy(key[:], "01234567890123456789012345678901")

		// Normalize gcmType
		var gType core.GcmTypeEnum
		switch gcmType % 2 {
		case 0:
			gType = core.GCM_AES256
		case 1:
			gType = core.GCM_SM4
		}

		// CBCDecryption should not panic on any input
		_, _ = core.CBCDecryption(gType, &key, ciphertext, false)
	})
}

// FuzzBase64DecodeSM2ECDHPublicKey tests base64 decoding of SM2 ECDH public keys.
// Handles untrusted key material from network messages.
func FuzzBase64DecodeSM2ECDHPublicKey(f *testing.F) {
	// Seed with valid and invalid base64 strings
	f.Add("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA") // 64 bytes when decoded
	f.Add("invalidbase64!!!")
	f.Add("")
	f.Add("AAAA") // too short
	f.Add("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA") // too long

	f.Fuzz(func(t *testing.T, pubKeyStr string) {
		// Should not panic on any input
		_, _ = gmsm.Base64DecodeSM2ECDHPublicKey(pubKeyStr)
	})
}

// FuzzBase64DecodeSM2ECDHPrivateKey tests base64 decoding of SM2 ECDH private keys.
func FuzzBase64DecodeSM2ECDHPrivateKey(f *testing.F) {
	f.Add("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=") // 32 bytes when decoded
	f.Add("invalidbase64!!!")
	f.Add("")
	f.Add("AAAA")

	f.Fuzz(func(t *testing.T, privKeyStr string) {
		_, _ = gmsm.Base64DecodeSM2ECDHPrivateKey(privKeyStr)
	})
}

// FuzzBase64DecodeSM2ECDSAPublicKey tests base64 decoding of SM2 ECDSA public keys.
// Used for signature verification.
func FuzzBase64DecodeSM2ECDSAPublicKey(f *testing.F) {
	f.Add("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
	f.Add("invalidbase64!!!")
	f.Add("")

	f.Fuzz(func(t *testing.T, pubKeyStr string) {
		_, _ = gmsm.Base64DecodeSM2ECDSAPublicKey(pubKeyStr)
	})
}

// FuzzSharedSecret tests ECDH shared secret computation with malformed public keys.
func FuzzSharedSecret(f *testing.F) {
	// Seed with various key sizes
	f.Add(make([]byte, 32), int(core.ECC_CURVE25519))
	f.Add(make([]byte, 64), int(core.ECC_SM2))
	f.Add(make([]byte, 16), int(core.ECC_CURVE25519))
	f.Add([]byte{}, int(core.ECC_SM2))

	f.Fuzz(func(t *testing.T, remotePubKey []byte, eccType int) {
		// Create a valid local key pair
		var eType core.EccTypeEnum
		switch eccType % 2 {
		case 0:
			eType = core.ECC_CURVE25519
		case 1:
			eType = core.ECC_SM2
		}

		localKey := core.NewECDH(eType)
		if localKey == nil {
			return
		}

		// SharedSecret should not panic on malformed remote public keys
		_ = localKey.SharedSecret(remotePubKey)
	})
}
