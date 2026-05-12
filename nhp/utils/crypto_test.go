package utils

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestRSAKeys(t *testing.T) {
	privKeyStr, pubKeyStr := GenerateRsaKey(1000)

	fmt.Println("private key: ", privKeyStr)
	fmt.Println("private key length: ", len(privKeyStr))
	fmt.Println("public key: ", pubKeyStr)
	fmt.Println("public key length", len(pubKeyStr))
}

func TestPubKeyFingerprintDeterministic(t *testing.T) {
	key := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
	}
	got1 := PubKeyFingerprint(key)
	got2 := PubKeyFingerprint(key)
	if got1 != got2 {
		t.Fatalf("fingerprint not deterministic: %s vs %s", got1, got2)
	}
	if len(got1) != PubKeyFingerprintLen {
		t.Fatalf("fingerprint length = %d, want %d", len(got1), PubKeyFingerprintLen)
	}
}

func TestPubKeyFingerprintDistinct(t *testing.T) {
	a := []byte("key-a-key-a-key-a-key-a-key-a-32")
	b := []byte("key-b-key-b-key-b-key-b-key-b-32")
	fa := PubKeyFingerprint(a)
	fb := PubKeyFingerprint(b)
	if fa == fb {
		t.Fatalf("distinct keys produced same fingerprint: %s", fa)
	}
}

// TestPubKeyFingerprintCrossLanguageVectors locks in the exact strings the
// TypeScript side asserts in endpoints/js-agent/test/crypto/fingerprint.test.ts.
// If either side changes algorithm — hash, prefix length, or base64 variant —
// these vectors break before a divergent build can ship.
func TestPubKeyFingerprintCrossLanguageVectors(t *testing.T) {
	filled := make([]byte, 32)
	for i := range filled {
		filled[i] = 0x42
	}
	if got := PubKeyFingerprint(filled); got != "Ql7U5KNrMOo" {
		t.Fatalf("fill(0x42) fingerprint = %q, want %q", got, "Ql7U5KNrMOo")
	}

	seq := make([]byte, 32)
	for i := range seq {
		seq[i] = byte(i + 1)
	}
	if got := PubKeyFingerprint(seq); got != "riFsLvUkejc" {
		t.Fatalf("[1..32] fingerprint = %q, want %q", got, "riFsLvUkejc")
	}
}

func TestPubKeyFingerprintFromBase64(t *testing.T) {
	raw := []byte("the-quick-brown-fox-jumps-over-32")
	want := PubKeyFingerprint(raw)
	got, err := PubKeyFingerprintFromBase64(base64.StdEncoding.EncodeToString(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("FromBase64 mismatch: got %s want %s", got, want)
	}

	if _, err := PubKeyFingerprintFromBase64("!!!not-base64!!!"); err == nil {
		t.Fatalf("expected error for invalid base64 input")
	}
}
