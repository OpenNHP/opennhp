package agent

import (
	"encoding/base64"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/keystore"
)

// TestGetAgentEcdhWithSealedKey guards the fix for the /publicKey handler:
// when the agent private key is stored as a sealed blob, GetAgentEcdh must
// still report the real device public key, not a key derived from a broken
// base64 decode of the "v1$..." blob.
func TestGetAgentEcdhWithSealedKey(t *testing.T) {
	// A known raw private key and the public key its device would present.
	e := core.NewECDH(core.ECC_CURVE25519)
	raw := e.PrivateKey()
	wantPub := core.ECDHFromKey(core.ECC_CURVE25519, raw).PublicKeyBase64()

	blob, err := keystore.Seal(raw, []byte("pw"))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    blob,
	}

	// Cached path (as populated by Start / ReinitWithKey): the resolved key
	// is used directly.
	cfg.resolvedPrivateKey = raw
	if got := cfg.GetAgentEcdh().PublicKeyBase64(); got != wantPub {
		t.Fatalf("cached path: got %q want %q", got, wantPub)
	}

	// Fallback path (cache empty, sealed blob in config): resolving on
	// demand from the env passphrase must yield the same public key. The
	// old code base64-decoded the blob directly and returned a wrong key.
	cfg.resolvedPrivateKey = nil
	t.Setenv(keystore.EnvPassphrase, "pw")
	if got := cfg.GetAgentEcdh().PublicKeyBase64(); got != wantPub {
		t.Fatalf("fallback path: got %q want %q", got, wantPub)
	}
}

// TestGetAgentEcdhPlainKeyUnchanged confirms the plain-base64 path is
// unaffected by the sealed-key support.
func TestGetAgentEcdhPlainKeyUnchanged(t *testing.T) {
	e := core.NewECDH(core.ECC_CURVE25519)
	raw := e.PrivateKey()
	wantPub := core.ECDHFromKey(core.ECC_CURVE25519, raw).PublicKeyBase64()

	cfg := &Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    base64.StdEncoding.EncodeToString(raw),
	}
	if got := cfg.GetAgentEcdh().PublicKeyBase64(); got != wantPub {
		t.Fatalf("plain path: got %q want %q", got, wantPub)
	}
}
