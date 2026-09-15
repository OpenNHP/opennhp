package server

import (
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// ACTokenEntry represents a server access token entry that maps to multiple AC tokens.
type ACTokenEntry struct {
	User       *common.AgentUser
	ResourceId string
	ACTokens   map[string]string
	OpenTime   int
	ExpireTime time.Time
}

// GetExpireTime implements the common.TokenEntry interface.
func (e *ACTokenEntry) GetExpireTime() time.Time {
	return e.ExpireTime
}

// GenerateAccessToken issues an opaque random access token for the entry.
// The previous SM3 digest was computed from public metadata plus UnixNano,
// which made the bearer credential susceptible to offline guessing when an
// attacker had a timing signal.
func (s *UdpServer) GenerateAccessToken(entry *ACTokenEntry) string {
	token := common.GenerateOpaqueToken()
	entry.ExpireTime = time.Now().Add(time.Duration(entry.OpenTime) * time.Second)
	s.tokenStore.Store(token, entry)
	return token
}

// VerifyAccessToken validates a token and returns the entry if found.
// Unlike AC's version, this does not extend the expiry time.
func (s *UdpServer) VerifyAccessToken(token string) *ACTokenEntry {
	entry, found := s.tokenStore.Load(token)
	if found {
		return entry
	}
	return nil
}
