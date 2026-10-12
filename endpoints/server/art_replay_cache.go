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
type artReplayCache struct {
	name              string
	mu                sync.Mutex
	capacityEvictions uint64
	capacityRejects   uint64
	lastWarning       time.Time
	entries           map[artReplayKey]struct{}
	order             *list.List
	maxEntries        int
	ttl               time.Duration
	now               func() time.Time
	// failClosedFull switches the capacity-exhaustion policy. ART traffic
	// is low-volume and benign, so the cache historically evicts the
	// oldest live entry under pressure (a controlled weakening of the
	// replay window, with a warning log). NHP_RLY carries every
	// web-agent knock behind every registered relay, and the only line
	// of defense for captured-and-replayed envelopes is this cache:
	// evicting a live entry under sustained pressure re-opens a replay
	// hole for every envelope still inside its 600s acceptance window,
	// so the relay cache must FAIL CLOSED — return false from
	// MarkSeen when the cache is at capacity and the new key has not
	// been seen before.
	failClosedFull bool
}

func newARTReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(artReplayCacheSize, artReplayCacheTTL, "ART", false, time.Now)
}

// newRelayReplayCache builds the same cache struct with relay-throughput
// sizing AND the fail-closed capacity policy. Keeping the cache struct
// generic instead of forking it means every future improvement to
// eviction / expiry order lands on both caches automatically, while the
// failClosedFull flag lets the two policies diverge on the one decision
// that matters.
func newRelayReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(relayReplayCacheSize, relayReplayCacheTTL, "NHP_RLY", true, time.Now)
}

func newARTReplayCacheWithParams(maxEntries int, ttl time.Duration, name string, failClosedFull bool, now func() time.Time) *artReplayCache {
	return &artReplayCache{
		name:           name,
		entries:        make(map[artReplayKey]struct{}),
		order:          list.New(),
		maxEntries:     maxEntries,
		ttl:            ttl,
		now:            now,
		failClosedFull: failClosedFull,
	}
}

// MarkSeen records an authenticated (peer pubkey, transaction ID, send time)
// tuple. It returns true on first observation and false for a duplicate or an
// invalid peer key. Callers must distinguish invalid input before calling.
//
// When the cache is at capacity and the key has not been seen before,
// behaviour depends on failClosedFull:
//   - false (ART): the oldest live entry is evicted, the warning fires
//     (rate-limited), and the new entry is admitted. Replay coverage is
//     reduced; this is acceptable for ART because ART traffic is
//     low-volume and the staleness window is generous enough that the
//     evicted entry's remaining replay risk is small.
//   - true (NHP_RLY): the duplicate check happens first; if it's a
//     duplicate of a retained entry we still return false (replay
//     detected, correct behaviour). For a genuinely new key at
//     capacity we ALSO return false and DO NOT evict. The warning
//     still fires so operators see the pressure; do not regress.
//     Failing closed here means a flooding relay sees its new
//     envelopes rejected at the cache boundary instead of replaying
//     captured ones from the just-evicted tail of the window — the
//     safer direction, because the per-connection monotonic / flood
//     gates are exempted for relay envelopes (see nhp/core.responder)
//     and the cache is the only line of defense.
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

	var warn string // "evictions" | "rejects" | ""
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
		if warn != "" {
			// Name is required so an overflow on the relay path
			// doesn't masquerade as ART (and vice versa) in the
			// log — operators reading "ART replay cache capacity
			// exceeded" would otherwise look in the wrong
			// subsystem to find the QPS ceiling.
			switch warn {
			case "evictions":
				log.Warning("%s replay cache capacity %d exceeded; live entries evicted (%d total), replay window shortened", c.name, c.maxEntries, c.capacityEvictions)
			case "rejects":
				log.Warning("%s replay cache capacity %d exceeded; new entries rejected (%d total) — fail closed", c.name, c.maxEntries, c.capacityRejects)
			}
		}
	}()

	now := c.now()
	c.evictExpired(now)
	if _, ok := c.entries[key]; ok {
		return false
	}
	if len(c.entries) >= c.maxEntries {
		if c.failClosedFull {
			c.capacityRejects++
			if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
				warn = "rejects"
				c.lastWarning = now
			}
			return false
		}
		c.capacityEvictions++
		if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
			warn = "evictions"
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
