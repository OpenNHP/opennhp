package curve

import (
	"testing"
)

func TestNonceBytesLength(t *testing.T) {
	var header HeaderCurve
	header.SetCounter(12345)
	header.SetNoncePrefix(0xABCDEF01)

	nonce := header.NonceBytes()
	if len(nonce) != GCMNonceSize {
		t.Errorf("Expected nonce length %d, got %d", GCMNonceSize, len(nonce))
	}
}

func TestNonceBytesContent(t *testing.T) {
	var header HeaderCurve
	header.SetCounter(0x0102030405060708)
	header.SetNoncePrefix(0xAABBCCDD)

	nonce := header.NonceBytes()

	// Check first 4 bytes are the nonce prefix
	expectedPrefix := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	for i := 0; i < 4; i++ {
		if nonce[i] != expectedPrefix[i] {
			t.Errorf("Nonce prefix byte %d: expected 0x%02X, got 0x%02X", i, expectedPrefix[i], nonce[i])
		}
	}

	// Check last 8 bytes are the counter
	expectedCounter := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	for i := 0; i < 8; i++ {
		if nonce[4+i] != expectedCounter[i] {
			t.Errorf("Nonce counter byte %d: expected 0x%02X, got 0x%02X", i, expectedCounter[i], nonce[4+i])
		}
	}
}

func TestNonceNoZeroGap(t *testing.T) {
	var header HeaderCurve
	header.SetCounter(0xFFFFFFFFFFFFFFFF)
	header.SetNoncePrefix(0xFFFFFFFF)

	nonce := header.NonceBytes()

	// Verify no zero bytes when both prefix and counter are set to non-zero
	for i, b := range nonce {
		if b == 0 {
			t.Errorf("Unexpected zero byte at position %d", i)
		}
	}
}

func TestNonceUniqueness(t *testing.T) {
	seen := make(map[string]bool)

	for i := uint64(0); i < 100; i++ {
		var header HeaderCurve
		header.SetCounter(i)
		header.SetNoncePrefix(uint32(i * 12345))

		nonce := string(header.NonceBytes())
		if seen[nonce] {
			t.Errorf("Duplicate nonce detected at iteration %d", i)
		}
		seen[nonce] = true
	}
}
