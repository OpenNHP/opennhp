package server

import (
	"encoding/base64"
	"encoding/binary"
	"time"

	"github.com/emmansun/gmsm/sm3"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/log"
)

type ACTokenEntry struct {
	User       *common.AgentUser
	ResourceId string
	ACTokens   map[string]string
	OpenTime   int
	ExpireTime time.Time
}

type TokenToACMap = map[string]*ACTokenEntry // server access token mapped into mutiple AC tokens
type TokenStore = map[string]TokenToACMap    // upper layer of tokens, indexed by first two characters

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

	s.tokenStoreMutex.Lock()
	defer s.tokenStoreMutex.Unlock()

	entry.ExpireTime = time.Now().Add(time.Duration(entry.OpenTime) * time.Second)
	tokenMap, found := s.tokenStore[token[0:1]]
	if found {
		tokenMap[token] = entry
	} else {
		tokenMap := make(TokenToACMap)
		tokenMap[token] = entry
		s.tokenStore[token[0:1]] = tokenMap
	}

	return token
}

func (s *UdpServer) VerifyAccessToken(token string) *ACTokenEntry {
	s.tokenStoreMutex.Lock()
	defer s.tokenStoreMutex.Unlock()

	tokenMap, found := s.tokenStore[token[0:1]]
	if found {
		entry, found := tokenMap[token]
		if found {
			return entry
		}
	}

	return nil
}

func (s *UdpServer) tokenStoreRefreshRoutine() {
	defer s.wg.Done()
	defer log.Info("tokenStoreRefreshRoutine stopped")

	log.Info("tokenStoreRefreshRoutine started")

	for {
		select {
		case <-s.signals.stop:
			return

		case <-time.After(TokenStoreRefreshInterval * time.Second):
			func() {
				s.tokenStoreMutex.Lock()
				defer s.tokenStoreMutex.Unlock()

				now := time.Now()
				for head, tokenMap := range s.tokenStore {
					for token, entry := range tokenMap {
						if now.After(entry.ExpireTime) {
							log.Info("[TokenStore] token %s expired, remove", token)
							delete(tokenMap, token)
						}
					}
					if len(tokenMap) == 0 {
						delete(s.tokenStore, head)
					}
				}
			}()
		}
	}
}
