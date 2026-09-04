package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// randMasterKey returns a fresh 32-byte AES-256 key for seal/open tests.
func randMasterKey(t *testing.T) []byte {
	t.Helper()
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return k
}

// TestSealOpenKeyRoundTrip verifies that openKey inverts sealKey for a
// representative private key, and that the sealed form is not just the
// plaintext (GCM must actually encrypt + authenticate).
func TestSealOpenKeyRoundTrip(t *testing.T) {
	master := randMasterKey(t)
	priv, err := GenerateSchemeAgnosticPrivateKey()
	if err != nil {
		t.Fatalf("GenerateSchemeAgnosticPrivateKey: %v", err)
	}

	sealed, err := sealKey(master, priv)
	if err != nil {
		t.Fatalf("sealKey: %v", err)
	}
	// The sealed blob must not leak the plaintext private key.
	if strings.Contains(sealed, priv) {
		t.Fatal("sealed form contains the plaintext private key")
	}
	// Two seals of the same key differ (random nonce).
	sealed2, _ := sealKey(master, priv)
	if sealed == sealed2 {
		t.Fatal("sealKey is deterministic; nonce not random")
	}

	got, err := openKey(master, sealed)
	if err != nil {
		t.Fatalf("openKey: %v", err)
	}
	if got != priv {
		t.Fatal("openKey did not invert sealKey")
	}
}

// TestOpenKeyRejectsWrongMasterKey ensures a GCM auth tag mismatch is
// surfaced as an error (not silent wrong plaintext).
func TestOpenKeyRejectsWrongMasterKey(t *testing.T) {
	priv, err := GenerateSchemeAgnosticPrivateKey()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	sealed, err := sealKey(randMasterKey(t), priv)
	if err != nil {
		t.Fatalf("sealKey: %v", err)
	}
	wrong := randMasterKey(t)
	if _, err := openKey(wrong, sealed); err == nil {
		t.Fatal("openKey succeeded with a wrong master key; GCM tag not enforced")
	}
}

// TestOpenKeyRejectsTampered verifies the GCM tag catches a single-bit
// flip in the ciphertext.
func TestOpenKeyRejectsTampered(t *testing.T) {
	master := randMasterKey(t)
	priv, _ := GenerateSchemeAgnosticPrivateKey()
	sealed, _ := sealKey(master, priv)

	combined, err := base64.StdEncoding.DecodeString(sealed)
	if err != nil {
		t.Fatalf("decode sealed: %v", err)
	}
	if len(combined) < 2 {
		t.Fatal("sealed blob too short to tamper")
	}
	combined[1] ^= 0x01 // flip a bit after the nonce
	tampered := base64.StdEncoding.EncodeToString(combined)

	if _, err := openKey(master, tampered); err == nil {
		t.Fatal("openKey accepted a tampered sealed blob; GCM tag not enforced")
	}
}

// TestDerivePublicKeyCrossScheme checks that a single scheme-agnostic
// private key derives a valid, distinct public key under BOTH schemes,
// and that derivation is deterministic (js-agent and the server must
// agree byte-for-byte from the same private key).
func TestDerivePublicKeyCrossScheme(t *testing.T) {
	priv, err := GenerateSchemeAgnosticPrivateKey()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	pubCurve1, err := DerivePublicKey(priv, CipherSchemeCurve25519)
	if err != nil {
		t.Fatalf("DerivePublicKey curve25519: %v", err)
	}
	pubGMSM1, err := DerivePublicKey(priv, CipherSchemeGMSM)
	if err != nil {
		t.Fatalf("DerivePublicKey gmsm: %v", err)
	}
	if pubCurve1 == "" || pubGMSM1 == "" {
		t.Fatal("derived public key is empty")
	}
	if pubCurve1 == pubGMSM1 {
		t.Fatal("curve25519 and gmsm derived the same public key from one private key")
	}

	// Determinism: same input -> same output (the server + js-agent
	// rely on byte-identical agreement).
	pubCurve2, _ := DerivePublicKey(priv, CipherSchemeCurve25519)
	pubGMSM2, _ := DerivePublicKey(priv, CipherSchemeGMSM)
	if pubCurve1 != pubCurve2 || pubGMSM1 != pubGMSM2 {
		t.Fatal("DerivePublicKey is not deterministic")
	}
}

// TestDerivePublicKeyMatchesCore confirms demoapp's DerivePublicKey
// bridge produces the same bytes as calling nhp/core directly — i.e.
// the eccTypeFor mapping + ECDHFromKey path agrees with the reference
// primitive the server uses, so a key registered via the demoapp lands
// byte-identical on the server side.
func TestDerivePublicKeyMatchesCore(t *testing.T) {
	priv, err := GenerateSchemeAgnosticPrivateKey()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	privBytes, err := base64.StdEncoding.DecodeString(priv)
	if err != nil {
		t.Fatalf("decode priv: %v", err)
	}

	for _, tc := range []struct {
		name   string
		scheme CipherScheme
		eccT   core.EccTypeEnum
	}{
		{"curve25519", CipherSchemeCurve25519, core.ECC_CURVE25519},
		{"gmsm", CipherSchemeGMSM, core.ECC_SM2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DerivePublicKey(priv, tc.scheme)
			if err != nil {
				t.Fatalf("DerivePublicKey: %v", err)
			}
			ref := core.ECDHFromKey(tc.eccT, privBytes)
			if ref == nil {
				t.Fatalf("core.ECDHFromKey(%s) returned nil", tc.name)
			}
			want := ref.PublicKeyBase64()
			if got != want {
				t.Fatalf("DerivePublicKey(%s) = %q, core reference = %q", tc.name, got, want)
			}
		})
	}
}

// TestGenerateSchemeAgnosticPrivateKeyFormat ensures the private key is
// base64-decodable to 32 bytes (valid under both schemes).
func TestGenerateSchemeAgnosticPrivateKeyFormat(t *testing.T) {
	priv, err := GenerateSchemeAgnosticPrivateKey()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	b, err := base64.StdEncoding.DecodeString(priv)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(b) != 32 {
		t.Fatalf("private key is %d bytes, want 32", len(b))
	}
	// Regenerating must produce a fresh key (no static IV / no global RNG state).
	other, _ := GenerateSchemeAgnosticPrivateKey()
	if bytes.Equal(b, mustDecode(t, other)) {
		t.Fatal("GenerateSchemeAgnosticPrivateKey returned the same key twice")
	}
}

func mustDecode(t *testing.T, s string) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return b
}
