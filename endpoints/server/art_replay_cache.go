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
// noEvict toggles two policies at the capacity boundary: when false (the
// ART default), a full cache evicts the oldest entry to make room for the
// new one — the answer-result traffic is low-volume enough that an
// occasional live eviction shortens the replay window in a way the
// staleness check still absorbs. When true (the relay cache), a full cache
// rejects the new envelope: under sustained relay+agent RPS the default
// 500k-entry capacity covers only capacity/TTL ≈ ~520 envelopes/s for the
// ~960s relay-replay TTL, so live eviction would silently shrink the
// replay window below the staleness acceptance floor and let a captured
// envelope repopulate a fresh inner connection. Rejecting new envelopes
// is the only policy that keeps the window intact when capacity saturates
// — the operator-facing consequence is a per-minute warning plus a
// counter, and a sustained under-pressure run degrades the cache (the
// right answer) instead of silently weakening protection.
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
	now                func() time.Time
	noEvict            bool
}

func newARTReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(artReplayCacheSize, artReplayCacheTTL, "ART", time.Now)
}

// newRelayReplayCache builds the cache struct with relay-throughput sizing
// and the noEvict policy. Keeping the cache struct generic instead of
// forking it means every future improvement to eviction / expiry order
// lands on both caches automatically; the policy difference lives on the
// cache instance, not a parallel type.
//
// The size parameter is required (no default) so the production caller
// (Start) must pass it explicitly — an earlier version defaulted it
// internally, but the production path in Start was constructing the
// cache through the lower-level factory instead, silently dropping the
// noEvict policy and leaving NHP_RLY replay protection to the
// evicting-then-shrinking path the commit was meant to remove.
func newRelayReplayCache(size int) *artReplayCache {
	return newARTReplayCacheWithParams(size, relayReplayCacheTTL, "NHP_RLY", time.Now).withNoEvict()
}

// withNoEvict toggles the cache to reject (not evict) when full. Returns
// the receiver so the constructor can chain it on. Tests use the same hook
// to construct a small noEvict cache — the test for "retained envelope
// is still rejected" cannot use the production 500k capacity.
func (c *artReplayCache) withNoEvict() *artReplayCache {
	c.noEvict = true
	return c
}

func newARTReplayCacheWithParams(maxEntries int, ttl time.Duration, name string, now func() time.Time) *artReplayCache {
	return &artReplayCache{
		name:       name,
		entries:    make(map[artReplayKey]struct{}),
		order:      list.New(),
		maxEntries: maxEntries,
		ttl:        ttl,
		now:        now,
	}
}

// MarkSeen records an authenticated (peer pubkey, transaction ID, send time)
// tuple. It returns true on first observation and false for a duplicate or an
// invalid peer key. Callers must distinguish invalid input before calling.
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

	var warnEvict, warnReject bool
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
		// Name is required so an overflow on the relay path doesn't
		// masquerade as ART (and vice versa) in the log — operators
		// reading "ART replay cache capacity exceeded" would otherwise
		// look in the wrong subsystem to find the QPS ceiling.
		if warnEvict {
			log.Warning("%s replay cache capacity %d exceeded; live entries evicted (%d total), replay window shortened", c.name, c.maxEntries, c.capacityEvictions)
		} else if warnReject {
			log.Warning("%s replay cache at capacity %d; new envelopes rejected without eviction (%d rejected since start), replay window preserved", c.name, c.maxEntries, c.capacityRejections)
		}
	}()

	now := c.now()
	c.evictExpired(now)
	if _, ok := c.entries[key]; ok {
		return false
	}
	if len(c.entries) >= c.maxEntries {
		if c.noEvict {
			c.capacityRejections++
			if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
				warnReject = true
				c.lastWarning = now
			}
			return false
		}
		c.capacityEvictions++
		if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
			warnEvict = true
			c.lastWarning = now
		}
		c.removeOldest()
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
