package server

import (
	"container/list"
	"encoding/base64"
	"errors"
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
	//
	// Operational note (post fail-closed policy): when this cache is
	// at capacity, NEW envelopes from any registered relay are
	// rejected rather than admitted at the expense of older entries.
	// Sustained aggregate relay throughput above ~520 envelopes/s
	// therefore starts dropping legitimate knocks — it does not just
	// shorten the replay window. Operators seeing the
	// "NHP_RLY replay cache capacity ... new entries rejected"
	// warning should raise RelayReplayCacheEntries in config.toml
	// before users notice. The
	// nhp_server_relay_replay_cache_rejects_total metric exposes the
	// same signal in a way dashboards/alerts can graph.
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

// Sentinel errors returned by MarkSeen. Callers (dedupeRecvART /
// dedupeRecvRLY) use errors.Is to distinguish the three rejection
// reasons and pick the right drop stage / log message / metric.
//
// "duplicate" is the load-bearing one for replay detection; "full"
// only happens on caches built with failClosedFull = true (NHP_RLY
// today) and on entries that exceed perPeerQuota; "invalid" covers
// malformed peer keys (the dedupe hook catches those before this is
// normally called).
var (
	ErrReplayCacheInvalid   = errors.New("replay cache: invalid peer key")
	ErrReplayCacheDuplicate = errors.New("replay cache: duplicate envelope")
	ErrReplayCacheFull      = errors.New("replay cache: full")
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
	name string
	mu   sync.Mutex
	// capacityEvictions and capacityRejects are counters incremented
	// while c.mu is held; the deferred warning reads them via a
	// snapshot taken under the same lock (see MarkSeen) so the read
	// never races with the increment.
	capacityEvictions uint64
	capacityRejects   uint64
	lastWarning       time.Time
	entries           map[artReplayKey]struct{}
	order             *list.List
	maxEntries        int
	ttl               time.Duration
	now               func() time.Time
	// failClosedFull switches the capacity-exhaustion policy. ART
	// traffic is low-volume and benign, so the cache historically
	// evicts the oldest live entry under pressure (a controlled
	// weakening of the replay window, with a warning log). NHP_RLY
	// carries every web-agent knock behind every registered relay,
	// and the only line of defense for captured-and-replayed
	// envelopes is this cache: evicting a live entry under sustained
	// pressure re-opens a replay hole for every envelope still
	// inside its 600s acceptance window, so the relay cache must
	// FAIL CLOSED — return ErrReplayCacheFull from MarkSeen when
	// the cache is at capacity and the new key has not been seen
	// before.
	failClosedFull bool
	// perPeerEnabled enables a per-peerPubkey sub-quota so a single
	// busy (or misbehaving-but-authenticated) relay cannot exhaust
	// the global cache and starve every other relay for the TTL.
	// The cap is computed dynamically as maxEntries /
	// max(len(peerEntryCount), 1) — i.e. each active peer gets an
	// even share of the global pool, and a peer's first envelope is
	// counted toward that share. Only enabled for the relay cache
	// today; ART traffic is AC-keyed and not subject to the
	// "one peer fills the pool" attack shape.
	perPeerEnabled bool
	peerEntryCount map[[core.PublicKeySizeEx]byte]uint32
}

func newARTReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(artReplayCacheSize, artReplayCacheTTL, "ART", false, false, time.Now)
}

// newRelayReplayCache builds the same cache struct with relay-throughput
// sizing, the fail-closed capacity policy, AND the per-peerPubkey
// sub-quota. Keeping the cache struct generic instead of forking it
// means every future improvement to eviction / expiry order lands on
// both caches automatically, while the two flags let the policies
// diverge on the decisions that matter.
func newRelayReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(relayReplayCacheSize, relayReplayCacheTTL, "NHP_RLY", true, true, time.Now)
}

func newARTReplayCacheWithParams(maxEntries int, ttl time.Duration, name string, failClosedFull, perPeerEnabled bool, now func() time.Time) *artReplayCache {
	return &artReplayCache{
		name:           name,
		entries:        make(map[artReplayKey]struct{}),
		order:          list.New(),
		maxEntries:     maxEntries,
		ttl:            ttl,
		now:            now,
		failClosedFull: failClosedFull,
		perPeerEnabled: perPeerEnabled,
		peerEntryCount: make(map[[core.PublicKeySizeEx]byte]uint32),
	}
}

// MarkSeen records an authenticated (peer pubkey, transaction ID, send time)
// tuple. Returns nil on first observation, ErrReplayCacheDuplicate on a
// duplicate of an existing entry, ErrReplayCacheFull when the cache (or
// this peer's per-peer quota, if enabled) is at capacity and the new
// key has not been seen before, or ErrReplayCacheInvalid on a
// malformed peer key.
//
// Capacity-exhaustion policy:
//   - failClosedFull=false (ART): the oldest live entry is evicted,
//     the warning fires (rate-limited), and the new entry is admitted.
//     Replay coverage is reduced; this is acceptable for ART because
//     ART traffic is low-volume and the staleness window is generous
//     enough that the evicted entry's remaining replay risk is small.
//   - failClosedFull=true (NHP_RLY): a NEW key at capacity returns
//     ErrReplayCacheFull without evicting any live entry. The cache
//     is the only line of defense for relay envelopes (the per-
//     connection monotonic / flood gates are exempted; see
//     nhp/core.responder.shouldCheckRecvAttack / shouldCheckFlood
//     and GHSA-674q-j8fh-972c), so failing closed is the safer
//     direction. Trade-off: when legitimate relay throughput exceeds
//     the cache's steady-state capacity, NEW envelopes are rejected
//     at the cache boundary — operators see this in the
//     "NHP_RLY replay cache capacity ... new entries rejected"
//     warning and in the
//     nhp_server_relay_replay_cache_rejects_total metric.
//
// Per-peer quota (relay cache only): on top of the global cap, each
// active peerPubkey gets a fair share of maxEntries. A single relay
// that fills its share is rejected at ErrReplayCacheFull even when
// other peers have headroom — without it, that one peer could
// otherwise exhaust the entire cache and starve every other relay.
func (c *artReplayCache) MarkSeen(peerPubkey []byte, txid uint64, sendTime int64) error {
	if c == nil || c.maxEntries <= 0 || c.ttl <= 0 {
		return ErrReplayCacheInvalid
	}
	if len(peerPubkey) != core.PublicKeySize && len(peerPubkey) != core.PublicKeySizeEx {
		return ErrReplayCacheInvalid
	}

	var key artReplayKey
	copy(key.peerPubkey[:], peerPubkey)
	key.peerPubkeyLen = uint8(len(peerPubkey))
	key.txid = txid
	key.sendTime = sendTime

	// Snapshot state under the lock for the deferred warning log so
	// the log does not read fields concurrently with another
	// goroutine's increment. Without this snapshot, go test -race
	// flags the read of c.capacityEvictions / c.capacityRejects in
	// the deferred func as an unsynchronized uint64 access (the
	// counters are written under c.mu on the increment path).
	var warn string // "evictions" | "rejects" | ""
	var snapshot uint64

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
				log.Warning("%s replay cache capacity %d exceeded; live entries evicted (%d total), replay window shortened", c.name, c.maxEntries, snapshot)
			case "rejects":
				log.Warning("%s replay cache capacity %d exceeded; new entries rejected (%d total) — fail closed", c.name, c.maxEntries, snapshot)
			}
		}
	}()

	now := c.now()
	c.evictExpiredLocked(now)
	if _, ok := c.entries[key]; ok {
		return ErrReplayCacheDuplicate
	}
	if c.perPeerEnabled {
		// Per-peer quota: maxEntries / N_active, where N_active is
		// the number of peers currently holding at least one
		// entry. The new peer is included in the divisor if this
		// is its first entry, so a 5th relay joining a cache
		// already serving 4 doesn't immediately over-commit any of
		// the existing 4.
		activePeers := len(c.peerEntryCount)
		if _, exists := c.peerEntryCount[key.peerPubkey]; !exists {
			activePeers++
		}
		perPeerCap := c.maxEntries / maxInt(activePeers, 1)
		if c.peerEntryCount[key.peerPubkey] >= uint32(perPeerCap) {
			c.capacityRejects++
			if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
				warn = "rejects"
				snapshot = c.capacityRejects
				c.lastWarning = now
			}
			return ErrReplayCacheFull
		}
	}
	if len(c.entries) >= c.maxEntries {
		if c.failClosedFull {
			c.capacityRejects++
			if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
				warn = "rejects"
				snapshot = c.capacityRejects
				c.lastWarning = now
			}
			return ErrReplayCacheFull
		}
		c.capacityEvictions++
		if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
			warn = "evictions"
			snapshot = c.capacityEvictions
			c.lastWarning = now
		}
		c.removeOldestLocked()
	}
	c.order.PushBack(artReplayEntry{key: key, expiresAt: now.Add(c.ttl)})
	c.entries[key] = struct{}{}
	if c.perPeerEnabled {
		c.peerEntryCount[key.peerPubkey]++
	}
	return nil
}

func (c *artReplayCache) evictExpiredLocked(now time.Time) {
	for elem := c.order.Front(); elem != nil; elem = c.order.Front() {
		entry := elem.Value.(artReplayEntry)
		if entry.expiresAt.After(now) {
			return
		}
		delete(c.entries, entry.key)
		if c.perPeerEnabled {
			c.decrPeerLocked(entry.key.peerPubkey)
		}
		c.order.Remove(elem)
	}
}

func (c *artReplayCache) removeOldestLocked() {
	elem := c.order.Front()
	if elem == nil {
		return
	}
	entry := elem.Value.(artReplayEntry)
	delete(c.entries, entry.key)
	if c.perPeerEnabled {
		c.decrPeerLocked(entry.key.peerPubkey)
	}
	c.order.Remove(elem)
}

// decrPeerLocked decrements the per-peer entry count for pubkey,
// removing the map entry when it hits zero so len(peerEntryCount)
// reflects only peers currently holding live entries. Called with
// c.mu held.
func (c *artReplayCache) decrPeerLocked(pubkey [core.PublicKeySizeEx]byte) {
	cnt := c.peerEntryCount[pubkey]
	if cnt == 0 {
		return
	}
	if cnt == 1 {
		delete(c.peerEntryCount, pubkey)
		return
	}
	c.peerEntryCount[pubkey] = cnt - 1
}

// maxInt is a tiny helper. We avoid the built-in max to keep this
// file readable for callers who haven't memorised Go 1.21+ ordering
// of generic built-ins yet.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func artPubkeyFingerprint(peerPubkey []byte) string {
	encoded := base64.RawURLEncoding.EncodeToString(peerPubkey)
	if len(encoded) > 12 {
		return encoded[:12]
	}
	return encoded
}
