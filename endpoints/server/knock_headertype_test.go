package server

import (
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// TestVerifyKnockHeaderType pins the secure-by-default behavior: a
// populated body that matches the wire is accepted (nil), an unpopulated
// (legacy) body is rejected as legacy, and a populated body that
// disagrees with the wire is rejected as a mismatch.
func TestVerifyKnockHeaderType(t *testing.T) {
	cases := []struct {
		name    string
		body    int
		wire    int
		wantErr *common.Error
	}{
		{"ok_knk", core.NHP_KNK, core.NHP_KNK, nil},
		{"ok_ext", core.NHP_EXT, core.NHP_EXT, nil},
		{"ok_rkn", core.NHP_RKN, core.NHP_RKN, nil},

		// Legacy agent: body never populated -> zero value (NHP_KPL).
		{"legacy_body_zero_vs_knk", core.NHP_KPL, core.NHP_KNK, common.ErrKnockHeaderTypeLegacy},
		{"legacy_body_zero_vs_ext", core.NHP_KPL, core.NHP_EXT, common.ErrKnockHeaderTypeLegacy},

		// On-path flip: authenticated body disagrees with the wire header.
		{"mismatch_knk_flipped_to_ext", core.NHP_KNK, core.NHP_EXT, common.ErrKnockHeaderTypeMismatch},
		{"mismatch_ext_flipped_to_knk", core.NHP_EXT, core.NHP_KNK, common.ErrKnockHeaderTypeMismatch},
		{"mismatch_knk_flipped_to_rkn", core.NHP_KNK, core.NHP_RKN, common.ErrKnockHeaderTypeMismatch},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := verifyKnockHeaderType(tc.body, tc.wire, 1, "test-addr")
			if got != tc.wantErr {
				t.Errorf("verifyKnockHeaderType(body=%d, wire=%d) = %v, want %v", tc.body, tc.wire, got, tc.wantErr)
			}
		})
	}
}

// TestEnforceKnockHeaderType_Gate covers the DisableKnockHeaderTypeValidation
// gate around verifyKnockHeaderType: with enforcement on (flag false) a legacy
// or mismatched body is rejected; with enforcement disabled (flag true, the
// transition escape hatch) the same body is accepted. A matching body is
// accepted either way.
func TestEnforceKnockHeaderType_Gate(t *testing.T) {
	cases := []struct {
		name     string
		disabled bool
		body     int
		wire     int
		wantErr  *common.Error
	}{
		{"enforced_ok", false, core.NHP_KNK, core.NHP_KNK, nil},
		{"enforced_legacy_rejected", false, core.NHP_KPL, core.NHP_KNK, common.ErrKnockHeaderTypeLegacy},
		{"enforced_mismatch_rejected", false, core.NHP_KNK, core.NHP_EXT, common.ErrKnockHeaderTypeMismatch},

		{"disabled_ok_still_ok", true, core.NHP_KNK, core.NHP_KNK, nil},
		{"disabled_legacy_accepted", true, core.NHP_KPL, core.NHP_KNK, nil},
		{"disabled_mismatch_accepted", true, core.NHP_KNK, core.NHP_EXT, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &UdpServer{}
			s.disableKnockHeaderTypeValidation.Store(tc.disabled)
			got := s.enforceKnockHeaderType(tc.body, tc.wire, 1, "test-addr")
			if got != tc.wantErr {
				t.Errorf("enforceKnockHeaderType(disabled=%v, body=%d, wire=%d) = %v, want %v",
					tc.disabled, tc.body, tc.wire, got, tc.wantErr)
			}
		})
	}
}
