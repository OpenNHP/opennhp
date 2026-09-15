package ac

import (
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
)

func TestGenerateAccessTokenUsesOpaqueRandomValues(t *testing.T) {
	a := &UdpAC{tokenStore: common.NewTokenStore[*AccessEntry]()}
	user := &common.AgentUser{
		UserId:         "user-1",
		DeviceId:       "device-1",
		OrganizationId: "org-1",
		AuthServiceId:  "service-1",
	}

	const count = 10_000
	seen := make(map[string]struct{}, count)
	for i := 0; i < count; i++ {
		entry := &AccessEntry{User: user, OpenTime: 60}
		token := a.GenerateAccessToken(entry)
		if _, duplicate := seen[token]; duplicate {
			t.Fatalf("duplicate token after %d generations", i)
		}
		seen[token] = struct{}{}
		if got := a.VerifyAccessToken(token); got != entry {
			t.Fatalf("token %d did not round-trip through the token store", i)
		}
	}
}

func TestIssueACTokenIfSuccess(t *testing.T) {
	tests := []struct {
		name       string
		errCode    string
		wantIssued bool
	}{
		{name: "explicit success", errCode: common.ErrSuccess.ErrorCode(), wantIssued: true},
		{name: "empty code", errCode: "", wantIssued: false},
		{name: "explicit failure", errCode: common.ErrACOperationFailed.ErrorCode(), wantIssued: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := &UdpAC{tokenStore: common.NewTokenStore[*AccessEntry]()}
			result := &common.ACOpsResultMsg{ErrCode: test.errCode}
			entry := &AccessEntry{User: &common.AgentUser{UserId: "user-1"}, OpenTime: 60}

			a.IssueACTokenIfSuccess(result, entry)

			if test.wantIssued {
				if result.ACToken == "" || a.VerifyAccessToken(result.ACToken) != entry {
					t.Fatal("successful operation did not issue a usable token")
				}
				return
			}
			if result.ACToken != "" || a.tokenStore.Size() != 0 {
				t.Fatal("failed operation issued or stored a token")
			}
		})
	}
}

func TestGenerateAccessTokenIncludesLatePacketBuffer(t *testing.T) {
	a := &UdpAC{tokenStore: common.NewTokenStore[*AccessEntry]()}
	entry := &AccessEntry{User: &common.AgentUser{UserId: "user-1"}, OpenTime: 60}

	before := time.Now()
	a.GenerateAccessToken(entry)
	after := time.Now()

	wantMin := before.Add(time.Duration(entry.OpenTime+accessTokenLatePacketBufferSeconds) * time.Second)
	wantMax := after.Add(time.Duration(entry.OpenTime+accessTokenLatePacketBufferSeconds) * time.Second)
	if entry.ExpireTime.Before(wantMin) || entry.ExpireTime.After(wantMax) {
		t.Fatalf("ExpireTime %v outside expected range [%v, %v]", entry.ExpireTime, wantMin, wantMax)
	}
}
