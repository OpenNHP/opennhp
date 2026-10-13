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
// below it. false (used for NHP_RLY) attempts safe eviction first: if the
// oldest entry's sendTime is already past the staleness floor (a replay of
// it would be rejected by validatePeer as stale, so dropping it costs no
// protection), evict that one. Only refuse if the oldest entry is still
// inside the acceptance window — those entries will age past the floor on
// their own and become safely evictable, so the cache self-recovers without
// operator action once a flood stops. The previous always-refuse policy
// kept the cache full until TTL expiry (~16 min after a flood stops); the
// safe-eviction policy bounds the unavailability window to the staleness
// floor (~10 min) and lets any stale entries the attacker left behind be
// reaped before they expire.
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
// throughput under burst), while the NHP_RLY cache first attempts safe
// eviction (drop the oldest entry only if it is past the staleness floor,
// so a replay of it would already be refused by validatePeer) and only
// refuses the new entry when no safe eviction is possible. The relay
// path keeps the cache contents at their original recency until TTL
// expiry; the safe-eviction refinement bounds unavailability after a flood
// to the staleness floor (DefaultRecvStalenessFloorSeconds) by letting any
// stale entries the attacker left behind be reaped as soon as they age
// past the floor.
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
			// NHP_RLY cache. Try safe eviction first: if the
			// oldest entry's sendTime is already past the
			// staleness floor, a replay of it would be
			// rejected as stale by validatePeer, so dropping
			// it costs no protection. This bounds the
			// unavailability window after a flood to the
			// floor (~10 min) instead of the TTL (~16 min),
			// because stale entries get reaped before their
			// TTL expires.
			//
			// Only when the cache is full of entries still
			// inside the acceptance window do we refuse —
			// those entries are still replayable and must
			// not be evicted. The caller (dedupeRecvRLY)
			// maps both duplicates and capacity refusals to
			// ErrServerDuplicateTransaction so the upstream
			// takes the same fail-closed path.
			if !c.evictSafelyEvictableOldest(now) {
				c.capacityRejections++
				if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
					warnReject = true
					evictions = c.capacityRejections
					c.lastWarning = now
				}
				return false
			}
		}
	}
	c.order.PushBack(artReplayEntry{key: key, expiresAt: now.Add(c.ttl)})
	c.entries[key] = struct{}{}
	return true
}

// evictSafelyEvictableOldest drops the oldest-by-insertion entry if a
// replay of it would already be rejected by validatePeer's staleness
// floor. Returns true if an entry was evicted.
//
// Safe eviction is the NHP_RLY cache's policy at capacity (see
// evictOnFull field docstring): an entry is "safely evictable" when
// its sendTime is at least DefaultRecvStalenessFloorSeconds behind
// the current local time — i.e. validatePeer would refuse a replay
// of it as ErrStalePacketReceived
// (remoteSendTime < LocalInitTime - DefaultRecvStalenessFloorSeconds),
// so dropping the entry costs no protection. sendTime is part of
// the key, so this is O(1): peek at the head of the order list.
//
// Forward-stamped entries (sendTime up to ARTRecvFutureSkewSeconds
// ahead of local time at insertion) shorten the safe-eviction window
// by that skew — but the staleness floor is the upper bound on
// replayability regardless of skew, so sendTime + floor is the right
// threshold for either skew direction.
func (c *artReplayCache) evictSafelyEvictableOldest(now time.Time) bool {
	elem := c.order.Front()
	if elem == nil {
		return false
	}
	entry := elem.Value.(artReplayEntry)
	// Replay is rejected when sendTime < now - floor, so the
	// entry is safely evictable when sendTime + floor < now.
	// Using strict "<" so an entry that is exactly at the floor
	// (sendTime + floor == now) stays in the cache: validatePeer
	// would still accept a replay of it.
	if entry.key.sendTime+core.DefaultRecvStalenessFloorSeconds >= now.Unix() {
		return false
	}
	delete(c.entries, entry.key)
	c.order.Remove(elem)
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
