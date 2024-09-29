package ac

import (
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/emmansun/gmsm/sm3"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/log"
)

type AgentUser struct {
	UserId         string
	DeviceId       string
	OrganizationId string
	AuthServiceId  string
}

type AccessEntry struct {
	AgentUser  *AgentUser
	SrcAddrs   []*common.NetAddress
	DstAddrs   []*common.NetAddress
	OpenTime   int
	ExpireTime time.Time
}

type TokenAccessMap = map[string]*AccessEntry // access token mapped into user and access information
type TokenStore = map[string]TokenAccessMap   // upper layer of tokens, indexed by first two characters

func (a *UdpAC) GenerateAccessToken(entry *AccessEntry) string {
	var tsBytes [8]byte
	currTime := time.Now().UnixNano()

	hash := sm3.New()
	binary.BigEndian.PutUint64(tsBytes[:], uint64(currTime))
	au := entry.AgentUser
	hash.Write([]byte(a.config.ACId + au.UserId + au.DeviceId + au.OrganizationId + au.AuthServiceId))
	hash.Write(tsBytes[:])
	token := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	hash.Reset()

	a.TokenStoreMutex.Lock()
	defer a.TokenStoreMutex.Unlock()

	entry.ExpireTime = time.Now().Add(time.Duration(entry.OpenTime) * time.Second)
	tokenMap, found := a.tokenStore[token[0:1]]
	if found {
		tokenMap[token] = entry
	} else {
		tokenMap := make(TokenAccessMap)
		tokenMap[token] = entry
		a.tokenStore[token[0:1]] = tokenMap
	}

	return token
}

func (a *UdpAC) VerifyAccessToken(token string) *AccessEntry {
	a.TokenStoreMutex.Lock()
	defer a.TokenStoreMutex.Unlock()

	tokenMap, found := a.tokenStore[token[0:1]]
	if found {
		entry, found := tokenMap[token]
		if found {
			return entry
		}
	}

	return nil
}

func (a *UdpAC) tokenStoreRefreshRoutine() {
	defer a.wg.Done()
	defer log.Info("tokenStoreRefreshRoutine stopped")

	log.Info("tokenStoreRefreshRoutine started")

	for {
		select {
		case <-a.signals.stop:
			return

		case <-time.After(TokenStoreRefreshInterval * time.Second):
			a.TokenStoreMutex.Lock()
			defer a.TokenStoreMutex.Unlock()

			now := time.Now()
			for head, tokenMap := range a.tokenStore {
				for token, entry := range tokenMap {
					if now.After(entry.ExpireTime) {
						log.Info("[TokenStore] token %s expired", token)
						delete(tokenMap, token)
					}
				}
				if len(tokenMap) == 0 {
					delete(a.tokenStore, head)
				}
			}
		}
	}
}
