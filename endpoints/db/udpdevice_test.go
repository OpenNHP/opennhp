package db

import (
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// TestGetOwnEcdhCipherSchemeMapping pins the fix for an inverted mapping:
// GetOwnEcdh used to select SM2 when DefaultCipherScheme == 0, but
// common.CIPHER_SCHEME_CURVE == 0, not SM2 — and the ztdo static key pair
// setup a few lines below GetOwnEcdh's only caller makes the correct
// comparison. With the shipped default DefaultCipherScheme = 0, that mix-up
// meant the symmetric agreement selected CURVE25519 while GetOwnEcdh handed
// it an SM2 static key pair: mismatched curves in the same handshake,
// deriving a wrong shared secret for DHP data-key wrapping.
func TestGetOwnEcdhCipherSchemeMapping(t *testing.T) {
	raw := core.NewECDH(core.ECC_CURVE25519).PrivateKey()

	for _, tc := range []struct {
		name   string
		scheme int
		want   core.EccTypeEnum
	}{
		{"CIPHER_SCHEME_CURVE (0, the shipped default) selects Curve25519", common.CIPHER_SCHEME_CURVE, core.ECC_CURVE25519},
		{"CIPHER_SCHEME_GMSM (1) selects SM2", common.CIPHER_SCHEME_GMSM, core.ECC_SM2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := &UdpDevice{
				config:     &Config{DefaultCipherScheme: tc.scheme},
				privateKey: raw,
			}
			got := a.GetOwnEcdh()
			want := core.ECDHFromKey(tc.want, raw)
			if got == nil || want == nil {
				t.Fatalf("ECDHFromKey returned nil (got=%v want=%v) — key derivation itself failed, not the mapping under test", got, want)
			}
			if got.PublicKeyBase64() != want.PublicKeyBase64() {
				t.Fatalf("GetOwnEcdh derived the wrong curve for DefaultCipherScheme=%d: got pubkey %s, want %s (%v)",
					tc.scheme, got.PublicKeyBase64(), want.PublicKeyBase64(), tc.want)
			}
		})
	}
}
