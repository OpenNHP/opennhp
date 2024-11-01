package ac

import (
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/emmansun/gmsm/sm3"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/log"
)

type AccessEntry struct {
	User       *common.AgentUser
	SrcAddrs   []*common.NetAddress
	DstAddrs   []*common.NetAddress
	OpenTime   int
	ExpireTime time.Time
}

type TokenToAccessMap = map[string]*AccessEntry // access token mapped into user and access information
type TokenStore = map[string]TokenToAccessMap   // upper layer of tokens, indexed by first two characters

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

	a.tokenStoreMutex.Lock()
	defer a.tokenStoreMutex.Unlock()

	entry.ExpireTime = time.Now().Add(time.Duration(entry.OpenTime+5) * time.Second) // keep token for additional 5 seconds in case a request is received late
	tokenMap, found := a.tokenStore[token[0:1]]
	if found {
		tokenMap[token] = entry
	} else {
		tokenMap := make(TokenToAccessMap)
		tokenMap[token] = entry
		a.tokenStore[token[0:1]] = tokenMap
	}

	return token
}

func (a *UdpAC) VerifyAccessToken(token string) *AccessEntry {
	a.tokenStoreMutex.Lock()
	defer a.tokenStoreMutex.Unlock()

	tokenMap, found := a.tokenStore[token[0:1]]
	if found {
		entry, found := tokenMap[token]
		if found {
			entry.ExpireTime = entry.ExpireTime.Add(time.Duration(entry.OpenTime) * time.Second)
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
			func() {
				a.tokenStoreMutex.Lock()
				defer a.tokenStoreMutex.Unlock()

				now := time.Now()
				for head, tokenMap := range a.tokenStore {
					for token, entry := range tokenMap {
						if now.After(entry.ExpireTime) {
							log.Info("[TokenStore] token %s expired, remove", token)
							delete(tokenMap, token)
						}
					}
					if len(tokenMap) == 0 {
						delete(a.tokenStore, head)
					}
				}
			}()
		}
	}
}
