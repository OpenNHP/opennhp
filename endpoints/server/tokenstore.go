package server

import (
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/emmansun/gmsm/sm3"
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

// GenerateAccessToken creates a new access token for the given entry.
func (s *UdpServer) GenerateAccessToken(entry *ACTokenEntry) string {
	var tsBytes [8]byte
	currTime := time.Now().UnixNano()

	hash := sm3.New()
	binary.BigEndian.PutUint64(tsBytes[:], uint64(currTime))
	au := entry.User
	hash.Write([]byte(s.config.Hostname + au.UserId + au.DeviceId + au.OrganizationId + au.AuthServiceId))
	hash.Write(tsBytes[:])
	token := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	hash.Reset()

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
