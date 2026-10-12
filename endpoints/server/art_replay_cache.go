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
	// answer-result traffic is a few entries per agent access, so even
	// a 10k-RPS server stays well under the 100k default.
	artReplayCacheSize = 100_000
	// artReplayCacheTTL pairs with validatePeer's future-skew gate:
	// entries must outlive every envelope the staleness floor accepts,
	// plus the future-skew bound (so a forward-stamped replay doesn't
	// outlive its cache entry), plus a small safety margin. The NHP_RLY
	// (relay) path no longer shares this cache — it uses a per-relay
	// sliding window (see relay_replay_window.go) because the
	// timestamp-keyed set was forced to evict live entries under
	// sustained relay throughput, and a strict noEvict variant
	// traded that for a DoS surface. Keeping the two caches' TTLs
	// separate-by-constant in this file even when they happen to
	// match keeps the ART cache readable without dragging relay
	// concerns into it.
	artReplayCacheTTL = (core.DefaultRecvStalenessFloorSeconds + core.ARTRecvFutureSkewSeconds + 60) * time.Second
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

// artReplayCache is a bounded, concurrency-safe TTL set for the
// ART (AC→server answer-result) replay-protection path. Entries are
// appended in expiry order, so expiration and capacity eviction are
// O(1). On capacity pressure the cache evicts the oldest entry; the
// ART traffic profile (a few entries per agent access) keeps live
// eviction rare, so the window-shortening effect is absorbed by the
// staleness check. The NHP_RLY path uses a different mechanism
// (relay_replay_window.go) that does not have a global capacity to
// fill and does not need eviction at all.
type artReplayCache struct {
	name              string
	mu                sync.Mutex
	capacityEvictions uint64
	lastWarning       time.Time
	entries           map[artReplayKey]struct{}
	order             *list.List
	maxEntries        int
	ttl               time.Duration
	now               func() time.Time
}

func newARTReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(artReplayCacheSize, artReplayCacheTTL, "ART", time.Now)
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

	var warn bool
	var evictions uint64
	c.mu.Lock()
	defer func() {
		// Capture the eviction count under the lock; another
		// goroutine can be running capacityEvictions++ while this
		// closure formats the warning, so reading the field after
		// Unlock would race.
		evictions = c.capacityEvictions
		c.mu.Unlock()
		// "ART replay cache capacity exceeded" — the warning names the
		// subsystem so an operator reading it knows where to look. There
		// is no NHP_RLY variant of this log; the relay path uses a
		// different mechanism entirely and lives in
		// relay_replay_window.go.
		if warn {
			log.Warning("%s replay cache capacity %d exceeded; live entries evicted (%d total), replay window shortened", c.name, c.maxEntries, evictions)
		}
	}()

	now := c.now()
	c.evictExpired(now)
	if _, ok := c.entries[key]; ok {
		return false
	}
	if len(c.entries) >= c.maxEntries {
		c.capacityEvictions++
		if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
			warn = true
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
