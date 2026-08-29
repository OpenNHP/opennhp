package server

import (
	"encoding/base64"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
)

func TestGenerateAccessTokenUsesOpaqueRandomValues(t *testing.T) {
	s := &UdpServer{tokenStore: common.NewTokenStore[*ACTokenEntry]()}
	user := &common.AgentUser{
		UserId:         "user-1",
		DeviceId:       "device-1",
		OrganizationId: "org-1",
		AuthServiceId:  "service-1",
	}

	const count = 10_000
	seen := make(map[string]struct{}, count)
	for i := 0; i < count; i++ {
		entry := &ACTokenEntry{User: user, ResourceId: "resource-1", OpenTime: 60}
		token := s.GenerateAccessToken(entry)
		if _, duplicate := seen[token]; duplicate {
			t.Fatalf("duplicate token after %d generations", i)
		}
		seen[token] = struct{}{}

		raw, err := base64.StdEncoding.DecodeString(token)
		if err != nil || len(raw) != 32 {
			t.Fatalf("token %d has invalid wire shape: len=%d err=%v", i, len(raw), err)
		}
		if got := s.VerifyAccessToken(token); got != entry {
			t.Fatalf("token %d did not round-trip through the token store", i)
		}
	}
}
