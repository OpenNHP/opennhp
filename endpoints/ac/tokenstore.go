package ac

import (
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/emmansun/gmsm/sm3"
)

// AccessEntry represents an access token entry with user and access information.
type AccessEntry struct {
	User       *common.AgentUser
	SrcAddrs   []*common.NetAddress
	DstAddrs   []*common.NetAddress
	OpenTime   int
	ExpireTime time.Time
}

// GetExpireTime implements the common.TokenEntry interface.
func (e *AccessEntry) GetExpireTime() time.Time {
	return e.ExpireTime
}

// GenerateAccessToken creates a new access token for the given entry.
// The token is stored with an additional 5-second buffer to handle late requests.
func (a *UdpAC) GenerateAccessToken(entry *AccessEntry) string {
	var tsBytes [8]byte
	currTime := time.Now().UnixNano()

	hash := sm3.New()
	binary.BigEndian.PutUint64(tsBytes[:], uint64(currTime))
	au := entry.User
	hash.Write([]byte(a.config.ACId + au.UserId + au.DeviceId + au.OrganizationId + au.AuthServiceId))
	hash.Write(tsBytes[:])
	token := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	hash.Reset()

	// Keep token for additional 5 seconds in case a request is received late
	entry.ExpireTime = time.Now().Add(time.Duration(entry.OpenTime+5) * time.Second)
	a.tokenStore.Store(token, entry)

	return token
}

// VerifyAccessToken validates a token and extends its expiry time if valid.
// Returns the AccessEntry if found, nil otherwise.
func (a *UdpAC) VerifyAccessToken(token string) *AccessEntry {
	entry, found := a.tokenStore.Load(token)
	if found {
		// Extend expiry time on successful verification
		entry.ExpireTime = entry.ExpireTime.Add(time.Duration(entry.OpenTime) * time.Second)
		return entry
	}
	return nil
}
