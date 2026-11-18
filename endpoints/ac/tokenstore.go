package ac

import (
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// accessTokenLatePacketBufferSeconds keeps AC tokens valid briefly beyond the
// requested open window for packets that arrive while enforcement state is
// being torn down.
const accessTokenLatePacketBufferSeconds = 5

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

// GenerateAccessToken issues an opaque random access token for the entry.
// The previous SM3 digest was computed from public metadata plus UnixNano,
// which made the bearer credential susceptible to offline guessing when an
// attacker had a timing signal.
func (a *UdpAC) GenerateAccessToken(entry *AccessEntry) string {
	token := common.GenerateOpaqueToken()
	entry.ExpireTime = time.Now().Add(time.Duration(entry.OpenTime+accessTokenLatePacketBufferSeconds) * time.Second)
	a.tokenStore.Store(token, entry)
	return token
}

// IssueACTokenIfSuccess issues a bearer token only for an explicitly
// successful AC operation. Failed operation results are logged by the server;
// including a valid token in those results would expose the credential to log
// readers without providing any legitimate client benefit.
func (a *UdpAC) IssueACTokenIfSuccess(artMsg *common.ACOpsResultMsg, entry *AccessEntry) {
	if artMsg.ErrCode != common.ErrSuccess.ErrorCode() {
		return
	}
	artMsg.ACToken = a.GenerateAccessToken(entry)
}

// VerifyAccessToken validates a token and resets its expiry if valid.
// Returns the AccessEntry if found, nil otherwise.
//
// Resets, not extends. The previous form added OpenTime to the expiry the
// entry already carried, so each successful verification bought another
// full window on top of whatever was left: N calls gave N*OpenTime of
// life, and a client that simply polled could hold an entry open
// indefinitely without knocking again. A token's lifetime is now the same
// whether it is used once or often — OpenTime from the last use.
func (a *UdpAC) VerifyAccessToken(token string) *AccessEntry {
	entry, found := a.tokenStore.Load(token)
	if found {
		entry.ExpireTime = time.Now().Add(time.Duration(entry.OpenTime) * time.Second)
		return entry
	}
	return nil
}
