package server

import (
	"encoding/base64"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/common/storageauth"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

const (
	defaultStorageMaxUploadMB          = 20 * 1024 // the limit before this was configurable
	defaultStorageMinFreeMB            = 1024
	defaultStorageMaxConcurrentUploads = 4

	// Per source IP, on the signed /storage routes only: it bounds the
	// signature checks an unauthenticated client can cause. Anonymous
	// downloads are not limited (agents behind one NAT or proxy share an IP).
	storageRatePerSec  = 5
	storageRateBurst   = 20
	storageRateEntries = 10000
	storageRateIdle    = 10 * time.Minute

	// A nonce is remembered for twice the clock skew window, which covers
	// every timestamp the window can still accept.
	storageNonceTTL        = 2 * storageauth.MaxClockSkew
	storageNonceMaxEntries = 100000

	storageCallerKey   = "storageCaller"   // gin context key: the nhp-db public key
	storageBodyHashKey = "storageBodyHash" // gin context key: the signed body SHA-256
)

// storageState is the per-HttpServer state of the /storage API.
type storageState struct {
	maxUploadBytes int64
	minFreeBytes   int64
	uploadSlots    chan struct{}
	limiter        *rknRateLimiter
	nonces         *nonceCache
	now            func() time.Time
}

func newStorageState(hc *HttpConfig) *storageState {
	maxMB, minFreeMB, slots := defaultStorageMaxUploadMB, defaultStorageMinFreeMB, defaultStorageMaxConcurrentUploads
	if hc != nil {
		if hc.StorageMaxUploadMB > 0 {
			maxMB = hc.StorageMaxUploadMB
		}
		if hc.StorageMinFreeMB > 0 {
			minFreeMB = hc.StorageMinFreeMB
		}
		if hc.StorageMaxConcurrentUploads > 0 {
			slots = hc.StorageMaxConcurrentUploads
		}
	}
	return &storageState{
		maxUploadBytes: int64(maxMB) << 20,
		minFreeBytes:   int64(minFreeMB) << 20,
		uploadSlots:    make(chan struct{}, slots),
		limiter:        newRknRateLimiter(storageRatePerSec, storageRateBurst, storageRateEntries, int64(storageRateIdle)),
		nonces:         newNonceCache(storageNonceMaxEntries),
		now:            time.Now,
	}
}

// storageEnabled reports whether the /storage API is mounted: on unless
// http.toml sets EnableStorage = false.
func storageEnabled(hc *HttpConfig) bool {
	return hc == nil || hc.EnableStorage == nil || *hc.EnableStorage
}

// rateLimit limits signed /storage requests per source IP. The IP is the
// TCP peer address (gin's RemoteIP), deliberately not X-Forwarded-For or
// similar headers, which any client can set.
func (st *storageState) rateLimit(c *gin.Context) {
	if !st.limiter.allowIP(net.ParseIP(c.RemoteIP()), st.now().UnixNano()) {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
	}
}

// requireDBSignature authenticates a request as coming from an nhp-db
// listed (and not expired) in db.toml, see package storageauth.
func (hs *HttpServer) requireDBSignature(c *gin.Context) {
	caller, reason := hs.verifyDBSignature(c.Request)
	if reason != "" {
		log.Warning("storage: rejected %s %s from %s: %s", c.Request.Method, c.Request.URL.Path, c.Request.RemoteAddr, reason)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.Set(storageCallerKey, caller)
	c.Set(storageBodyHashKey, c.Request.Header.Get(storageauth.HeaderBodyHash))
}

// verifyDBSignature returns the caller's public key, or a reason (for the
// log only; clients get a bare 401) why the request is not authenticated.
func (hs *HttpServer) verifyDBSignature(r *http.Request) (caller string, reason string) {
	st := hs.storage
	pubB64 := r.Header.Get(storageauth.HeaderPublicKey)
	sig := r.Header.Get(storageauth.HeaderSignature)
	nonce := r.Header.Get(storageauth.HeaderNonce)
	bodyHash := r.Header.Get(storageauth.HeaderBodyHash)
	if pubB64 == "" || sig == "" || nonce == "" || bodyHash == "" {
		return "", "missing signature headers"
	}
	ts, err := strconv.ParseInt(r.Header.Get(storageauth.HeaderTimestamp), 10, 64)
	if err != nil {
		return "", "bad timestamp"
	}
	if skew := st.now().Sub(time.Unix(ts, 0)); skew > storageauth.MaxClockSkew || skew < -storageauth.MaxClockSkew {
		return "", "timestamp outside the allowed window"
	}
	if !storageauth.ValidNonce(nonce) {
		return "", "bad nonce"
	}
	contentLength := max(r.ContentLength, 0)
	// A request without a body must say so; one with a body is checked
	// against the digest by its handler, as the body is read.
	if !storageauth.ValidBodyHash(bodyHash) || (contentLength == 0 && bodyHash != storageauth.EmptyBodySHA256) {
		return "", "bad body digest"
	}

	pub, err := base64.StdEncoding.DecodeString(pubB64)
	if err != nil {
		return "", "bad public key"
	}
	// The key's size tells the scheme the nhp-db signed with.
	var scheme int
	switch len(pub) {
	case core.PublicKeySize:
		scheme = common.CIPHER_SCHEME_CURVE
	case core.PublicKeySizeEx:
		scheme = common.CIPHER_SCHEME_GMSM
	default:
		return "", "bad public key"
	}

	us := hs.udpServer
	us.dbPeerMapMutex.Lock()
	peer := us.dbPeerMap[pubB64]
	us.dbPeerMapMutex.Unlock()
	if peer == nil {
		return "", "public key not in db.toml"
	}
	if peer.IsExpired() {
		return "", "nhp-db peer expired"
	}

	ecdh := us.device.GetEcdhByCipherScheme(scheme)
	if ecdh == nil {
		return "", "no server key for the scheme"
	}
	secret := ecdh.SharedSecret(pub)
	if secret == nil {
		return "", "ECDH failed"
	}
	req := storageauth.Request{
		Method:          r.Method,
		Path:            r.URL.Path,
		ContentLength:   contentLength,
		PublicKeyBase64: pubB64,
		Timestamp:       ts,
		Nonce:           nonce,
		BodySHA256:      bodyHash,
	}
	if !storageauth.Verify(secret, req, sig) {
		return "", "bad signature"
	}
	// Record the nonce only once the signature is valid, so unauthenticated
	// requests cannot fill the cache.
	if !st.nonces.add(pubB64+"/"+nonce, st.now()) {
		return "", "replayed or too many requests in flight"
	}
	return pubB64, ""
}

// nonceCache remembers recently used nonces so a signed request cannot be
// replayed within the timestamp window. Every entry lives storageNonceTTL,
// so entries expire in the order they were added: a FIFO queue lets each
// add drop the expired ones from the front at amortized O(1) cost. When
// full of live entries it refuses new ones rather than forget a live one.
type nonceCache struct {
	mu         sync.Mutex
	seen       map[string]struct{}
	queue      []nonceEntry // oldest first, from head on
	head       int          // queue[:head] are expired and already removed from seen
	maxEntries int
}

type nonceEntry struct {
	nonce  string
	expiry time.Time
}

func newNonceCache(maxEntries int) *nonceCache {
	return &nonceCache{seen: make(map[string]struct{}), maxEntries: maxEntries}
}

// add records nonce and reports whether it was new.
func (nc *nonceCache) add(nonce string, now time.Time) bool {
	nc.mu.Lock()
	defer nc.mu.Unlock()

	for nc.head < len(nc.queue) && !now.Before(nc.queue[nc.head].expiry) {
		delete(nc.seen, nc.queue[nc.head].nonce)
		nc.queue[nc.head] = nonceEntry{}
		nc.head++
	}
	// Compact once the dead prefix is at least half the slice, so each
	// entry is copied a bounded number of times.
	if nc.head > 0 && nc.head*2 >= len(nc.queue) {
		nc.queue = append(nc.queue[:0], nc.queue[nc.head:]...)
		nc.head = 0
	}

	if _, ok := nc.seen[nonce]; ok {
		return false
	}
	if len(nc.seen) >= nc.maxEntries {
		return false
	}
	nc.seen[nonce] = struct{}{}
	nc.queue = append(nc.queue, nonceEntry{nonce: nonce, expiry: now.Add(storageNonceTTL)})
	return true
}
