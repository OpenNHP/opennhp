package server

import (
	"container/list"
	"encoding/base64"
	"sync"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

const (
	// artReplayCacheSize is sized for the AC↔server ART channel —
	// answer-result traffic is a few entries per agent access, so even a
	// 10k-RPS server stays well under the 100k default.
	artReplayCacheSize = 100_000
	// relayReplayCacheSize is sized for relay throughput, not ART
	// throughput. NHP_RLY envelopes carry every web-agent KNK/RKN/EXT
	// behind every registered relay — one envelope per real agent
	// knock — so the cache needs to absorb a sustained relay+agent RPS
	// across its 16-minute TTL without capacity eviction (which would
	// shorten the replay window for every live entry evicted under
	// pressure). 500k entries over a 16min TTL is ~520 envelopes/s
	// sustained; the past failure mode (copying the 100k ART size) was
	// a per-second sustained ceiling roughly five times too low.
	relayReplayCacheSize = 500_000
	// artReplayCacheTTL pairs with validatePeer's future-skew gate:
	// entries must outlive every envelope the staleness floor accepts,
	// plus the future-skew bound (so a forward-stamped replay doesn't
	// outlive its cache entry), plus a small safety margin. Today this
	// is the same value for both caches; kept as separate constants so
	// the two caches can be tuned independently if the relay path ever
	// needs a different reconciliation window.
	artReplayCacheTTL   = (core.DefaultRecvStalenessFloorSeconds + core.ARTRecvFutureSkewSeconds + 60) * time.Second
	relayReplayCacheTTL = artReplayCacheTTL
)

// artReplayKey identifies one authenticated ART packet. The sender timestamp
// distinguishes a fresh packet after an AC counter reset from a byte-identical
// replay of an older packet.
type artReplayKey struct {
	peerPubkey    [core.PublicKeySizeEx]byte
	peerPubkeyLen uint8
	txid          uint64
	sendTime      int64
}

type artReplayEntry struct {
	key       artReplayKey
	expiresAt time.Time
}

// artReplayCache is a bounded, concurrency-safe TTL set. Entries are appended
// in expiry order, so expiration and capacity eviction are O(1).
//
// evictOnFull governs what happens when the cache is at capacity and a fresh
// key arrives. true (the legacy behavior, used for ART) evicts the oldest
// live entry to make room, shortening the replay window for every entry
// below it. false (used for NHP_RLY) refuses the new key instead — at relay
// RPS a capacity-evicting cache keeps only ~5 minutes of live entries, well
// under the 600-second acceptance window, so a captured envelope could be
// re-accepted after eviction and reach a fresh inner connection. Rejecting
// keeps the cache's contents at their original recency until TTL expiry.
type artReplayCache struct {
	name               string
	mu                 sync.Mutex
	capacityEvictions  uint64
	capacityRejections uint64
	lastWarning        time.Time
	entries            map[artReplayKey]struct{}
	order              *list.List
	maxEntries         int
	ttl                time.Duration
	evictOnFull        bool
	now                func() time.Time
}

func newARTReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(artReplayCacheSize, artReplayCacheTTL, "ART", true, time.Now)
}

// newRelayReplayCache builds the same cache struct with relay-throughput
// sizing. Keeping the cache struct generic instead of forking it means every
// future improvement to eviction / expiry order lands on both caches
// automatically. Relay sizing deliberately refuses new keys at capacity
// (see evictOnFull comment on artReplayCache) — evicting live entries at
// relay RPS would shorten the cache's effective replay window to ~5min,
// below the 600-second acceptance window.
func newRelayReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(relayReplayCacheSize, relayReplayCacheTTL, "NHP_RLY", false, time.Now)
}

func newARTReplayCacheWithParams(maxEntries int, ttl time.Duration, name string, evictOnFull bool, now func() time.Time) *artReplayCache {
	return &artReplayCache{
		name:        name,
		entries:     make(map[artReplayKey]struct{}),
		order:       list.New(),
		maxEntries:  maxEntries,
		ttl:         ttl,
		evictOnFull: evictOnFull,
		now:         now,
	}
}

// MarkSeen records an authenticated (peer pubkey, transaction ID, send time)
// tuple. It returns true on first observation and false for a duplicate, a
// capacity rejection, or an invalid peer key. Callers must distinguish
// invalid input before calling.
//
// When the cache is at capacity and the key is fresh, behavior depends on
// evictOnFull: the ART cache evicts the oldest live entry (preserving
// throughput under burst), while the NHP_RLY cache refuses the new entry so
// a captured-and-replayed envelope cannot be re-accepted after the live
// entry it shadowed was evicted under pressure.
func (c *artReplayCache) MarkSeen(peerPubkey []byte, txid uint64, sendTime int64) bool {
	if c == nil || c.maxEntries <= 0 || c.ttl <= 0 {
		return false
	}
	if len(peerPubkey) != core.PublicKeySize && len(peerPubkey) != core.PublicKeySizeEx {
		return false
	}

	var key artReplayKey
	copy(key.peerPubkey[:], peerPubkey)
	key.peerPubkeyLen = uint8(len(peerPubkey))
	key.txid = txid
	key.sendTime = sendTime

	var warnEvict bool
	var warnReject bool
	var evictions uint64
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
		// Name is required so an overflow on the relay path
		// doesn't masquerade as ART (and vice versa) in the
		// log — operators reading "ART replay cache capacity
		// exceeded" would otherwise look in the wrong
		// subsystem to find the QPS ceiling.
		if warnEvict {
			log.Warning("%s replay cache capacity %d exceeded; live entries evicted (%d total), replay window shortened", c.name, c.maxEntries, evictions)
		}
		if warnReject {
			log.Warning("%s replay cache capacity %d reached; new envelope refused to protect live entries (%d total refusals)", c.name, c.maxEntries, evictions)
		}
	}()

	now := c.now()
	c.evictExpired(now)
	if _, ok := c.entries[key]; ok {
		return false
	}
	if len(c.entries) >= c.maxEntries {
		if c.evictOnFull {
			c.capacityEvictions++
			if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
				warnEvict = true
				evictions = c.capacityEvictions
				c.lastWarning = now
			}
			c.removeOldest()
		} else {
			// Capacity-full but configured to reject. Refuse
			// the new key rather than evicting a live entry —
			// evicting would let a captured envelope be
			// re-accepted once its shadow left the cache,
			// reaching a fresh inner connection. The caller
			// (dedupeRecvRLY) maps both duplicates and
			// capacity refusals to ErrServerDuplicateTransaction
			// so the upstream takes the same fail-closed path.
			c.capacityRejections++
			if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
				warnReject = true
				evictions = c.capacityRejections
				c.lastWarning = now
			}
			return false
		}
	}
	c.order.PushBack(artReplayEntry{key: key, expiresAt: now.Add(c.ttl)})
	c.entries[key] = struct{}{}
	return true
}

func (c *artReplayCache) evictExpired(now time.Time) {
	for elem := c.order.Front(); elem != nil; elem = c.order.Front() {
		entry := elem.Value.(artReplayEntry)
		if entry.expiresAt.After(now) {
			return
		}
		delete(c.entries, entry.key)
		c.order.Remove(elem)
	}
}

func (c *artReplayCache) removeOldest() {
	elem := c.order.Front()
	if elem == nil {
		return
	}
	entry := elem.Value.(artReplayEntry)
	delete(c.entries, entry.key)
	c.order.Remove(elem)
}

func artPubkeyFingerprint(peerPubkey []byte) string {
	encoded := base64.RawURLEncoding.EncodeToString(peerPubkey)
	if len(encoded) > 12 {
		return encoded[:12]
	}
	return encoded
}
