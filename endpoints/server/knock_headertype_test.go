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
