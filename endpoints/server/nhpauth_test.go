package server

import (
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestKnockAuthorized(t *testing.T) {
	success := common.ErrSuccess.ErrorCode()
	denied := common.ErrAuthServiceProviderNotFound.ErrorCode()

	cases := []struct {
		name       string
		headerType int
		ack        *common.ServerKnockAckMsg
		dhpAck     *common.ServerDHPKnockAckMsg
		want       bool
	}{
		{"knk success", core.NHP_KNK, &common.ServerKnockAckMsg{ErrCode: success}, nil, true},
		{"knk denied", core.NHP_KNK, &common.ServerKnockAckMsg{ErrCode: denied}, nil, false},
		// Regression: AuthWithNHP can hand back a nil ack (plugin not
		// implemented / recovered panic). Must be treated as denied, not panic.
		{"knk nil ack", core.NHP_KNK, nil, nil, false},
		{"ext nil ack", core.NHP_EXT, nil, nil, false},
		{"dhp success", core.DHP_KNK, nil, &common.ServerDHPKnockAckMsg{ErrCode: success}, true},
		{"dhp denied", core.DHP_KNK, nil, &common.ServerDHPKnockAckMsg{ErrCode: denied}, false},
		{"dhp nil ack", core.DHP_KNK, nil, nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := knockAuthorized(tc.headerType, tc.ack, tc.dhpAck); got != tc.want {
				t.Errorf("knockAuthorized(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}
