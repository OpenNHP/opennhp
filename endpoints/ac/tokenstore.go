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
// entry already carried, so each verification bought another full window
// on top of whatever was left and N calls gave N*OpenTime of life. The
// window is now the same whether a token is used once or often.
//
// This bounds how long one verification is worth; it does not bound the
// entry. A client that refreshes more often than every OpenTime keeps its
// grant alive indefinitely, which is what "refresh" means. Requiring a
// fresh knock eventually would need an absolute cap — issue time plus a
// maximum lifetime — which is a separate decision.
//
// The window matches GenerateAccessToken's, late-packet buffer included:
// without it the first refresh would shorten the token's life by the
// seconds that exist so late packets still match while enforcement is
// being torn down.
func (a *UdpAC) VerifyAccessToken(token string) *AccessEntry {
	entry, found := a.tokenStore.Load(token)
	if found {
		entry.ExpireTime = time.Now().Add(time.Duration(entry.OpenTime+accessTokenLatePacketBufferSeconds) * time.Second)
		return entry
	}
	return nil
}
