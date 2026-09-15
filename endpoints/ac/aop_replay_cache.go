package ac

import (
	"encoding/base64"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/golang-lru/v2/simplelru"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// AOP dedupe remembers authenticated (peer key, transaction, timestamp) tuples
// across connections. The monotonic timestamp and flood gates exempt AOP so
// concurrent, reordered requests remain valid. This cache rejects duplicates.
// State is process-local: a restart clears it; the AOP staleness check limits
// the remaining replay window. Entries can also be lost under capacity pressure.
const (
	aopReplayCacheSize = 100_000
	aopReplayCacheTTL  = (core.DefaultRecvStalenessFloorSeconds + 60) * time.Second
	// pubkeyFingerprintLen is the truncation budget for the
	// log-line fingerprint emitted on duplicate-drops. 12 base64
	// chars ≈ 9 bytes (~72 bits) of pubkey entropy — enough to
	// distinguish one misbehaving server from a deployment-wide signal
	// in oncall logs without bloating the line or leaking the full
	// key. Birthday-bound collision probability for a typical
	// deployment of a few dozen servers is on the order of
	// 1e-19, so the truncation is collision-safe in operational
	// terms; if a deployment grows past ~10 000 servers the
	// budget should be revisited.
	pubkeyFingerprintLen = 12
)

// aopReplayCache is a bounded set with lazy expiry. It owns no goroutine.
// The outer mutex makes lookup and insertion one operation.
type aopReplayCache struct {
	mu                sync.Mutex
	lru               *simplelru.LRU[string, time.Time]
	ttl               time.Duration
	capacityEvictions atomic.Uint64
}

func newAOPReplayCache() *aopReplayCache {
	return newAOPReplayCacheWithParams(aopReplayCacheSize, aopReplayCacheTTL)
}

// newAOPReplayCacheWithParams allows short TTLs and small capacities in tests.
func newAOPReplayCacheWithParams(size int, ttl time.Duration) *aopReplayCache {
	c := &aopReplayCache{ttl: ttl}
	var err error
	c.lru, err = simplelru.NewLRU[string, time.Time](size, func(_ string, expires time.Time) {
		if time.Now().Before(expires) {
			c.capacityEvictions.Add(1)
		}
	})
	if err != nil {
		panic(err)
	} // sizes are validated at startup; tests use positive sizes.
	return c
}

// MarkSeen records the (peerPubkey, txid, sendTime) triple and
// returns true on first observation, false if the triple is
// already in the cache.
//
// Caller contract — peerPubkey and sendTime must both be the
// AEAD-authenticated values from PacketParserData
// (ppd.RemotePubKey and ppd.RemoteSendTime). The handler-level
// public-key length guard in HandleUdpACOperations is
// the canonical place to reject upstream-invariant violations
// (validatePeer should always populate RemotePubKey); MarkSeen
// returning false on an empty pubkey is fail-closed insurance, not
// the primary error surface — a future caller that skips the
// upstream guard will still not silently key every empty-pubkey
// AOP under the same `":<txid>:<ts>"` slot.
//
// Curve and GMSM use different authenticated public-key lengths. Any other
// length fails closed.
func (c *aopReplayCache) MarkSeen(peerPubkey []byte, txid uint64, sendTime int64) bool {
	if len(peerPubkey) != core.PublicKeySize && len(peerPubkey) != core.PublicKeySizeEx {
		return false
	}

	key := aopReplayKey(peerPubkey, txid, sendTime)

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if expires, exists := c.lru.Get(key); exists && now.Before(expires) {
		return false
	}
	c.lru.Add(key, now.Add(c.ttl))
	return true
}

// Len is the current entry count, exposed for tests. Takes the
// same mutex as MarkSeen so a Len() call never observes the cache
// between MarkSeen's Get and Add — necessary for deterministic
// `Len() == cap` assertions ordered after a MarkSeen from the
// same goroutine.
func (c *aopReplayCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.Len()
}

// aopReplayKey concatenates pubkey bytes with separators, the
// decimal txid, and the decimal sendTime. Tracked in #1460 —
// switching to a fixed-size byte array key is zero-alloc but
// unneeded at current AOP cadence. Per-call cost is ~3 short
// allocations (`string(peerPubkey)` ~32 B, the
// FormatUint/FormatInt buffers ~20 B each, the final concat
// builder ~80 B) — negligible at <100 AOP/sec, becomes meaningful
// at ~10k AOP/sec where the GC pressure starts dominating the
// per-packet cost.
func aopReplayKey(peerPubkey []byte, txid uint64, sendTime int64) string {
	return string(peerPubkey) + ":" + strconv.FormatUint(txid, 10) + ":" + strconv.FormatInt(sendTime, 10)
}

// pubkeyFingerprint returns a short, log-friendly representation of
// a public key for breadcrumb correlation, URL-safe-base64-encoded
// and truncated to pubkeyFingerprintLen chars. URL-safe encoding
// (-, _) avoids `/` collisions in log indexers that treat it as a
// path separator (Loki, CloudWatch Insights). Empty input returns
// the literal "empty" so a missing-pubkey breadcrumb is still
// grep-able.
func pubkeyFingerprint(peerPubkey []byte) string {
	if len(peerPubkey) == 0 {
		return "empty"
	}
	enc := base64.RawURLEncoding.EncodeToString(peerPubkey)
	if len(enc) > pubkeyFingerprintLen {
		return enc[:pubkeyFingerprintLen]
	}
	return enc
}
