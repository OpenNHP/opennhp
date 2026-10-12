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
	// relayReplayCacheSize is the NHP_RLY replay cache ceiling. Unlike ART
	// (which evicts the oldest entry on overflow — acceptable because ART
	// is server-driven, attacker-rate-bounded), the relay path is exposed
	// to a flood of legitimate-looking traffic from every web-agent behind
	// every registered relay; evicting a live entry would re-inject a
	// previously-authenticated envelope on a fresh per-client connection
	// and bypass the only replay-rejection hook (NHP_RLY is exempt from
	// shouldCheckRecvAttack/shouldCheckFlood). The cache therefore refuses
	// new envelopes when full — see newRelayReplayCacheWithParams.
	//
	// Sizing rule: max sustained envelopes/s must stay below entries/960
	// (the TTL of (DefaultRecvStalenessFloorSeconds + ARTRecvFutureSkewSeconds
	// + 60) seconds). 500k entries covers ~520 envelopes/s sustained; raise
	// RelayReplayCacheEntries if the per-minute Critical "REFUSED" log
	// starts firing.
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

// replayAdmit is the four-state outcome of Admit. The split between
// replayAdmitDuplicate and replayAdmitCacheFull is what lets the caller
// distinguish "this envelope has been seen before" from "we can't accept any
// more envelopes right now" — confusing the two would mislead operators (one
// points at a security event, the other at capacity planning).
type replayAdmit int

const (
	// replayAdmitNew: first observation; the envelope was added to the cache.
	replayAdmitNew replayAdmit = iota
	// replayAdmitDuplicate: the envelope key already exists in the cache.
	replayAdmitDuplicate
	// replayAdmitInvalid: the cache or the input is not usable (nil cache,
	// zero capacity/ttl, or an unsupported pubkey length). The cache is
	// unchanged. In production this is unreachable — construction sets
	// valid capacity/ttl and dedupeRecvRLY validates the pubkey length
	// before calling Admit — so the result exists primarily for tests.
	replayAdmitInvalid
	// replayAdmitCacheFull: rejectWhenFull is true and the cache is at
	// capacity. The cache is unchanged (no eviction, no insertion).
	replayAdmitCacheFull
)

// artReplayCache is a bounded, concurrency-safe TTL set. Entries are appended
// in expiry order, so expiration and capacity eviction are O(1).
//
// rejectWhenFull picks the overflow policy:
//   - false (ART): evict the oldest entry and insert the new one. Acceptable
//     because ART traffic is server-driven and bounded by the server's own
//     out-bound rate, so an attacker cannot push the cache over capacity.
//   - true (RLY): refuse the new envelope and leave live entries intact.
//     Required because RLY is exposed to a flood of legitimate-looking
//     envelopes from every web-agent behind every registered relay, and
//     evicting a live entry would re-inject a previously-authenticated
//     envelope on a fresh per-client connection, bypassing the only
//     replay-rejection hook.
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
	rejectWhenFull     bool
}

func newARTReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(artReplayCacheSize, artReplayCacheTTL, "ART", time.Now)
}

// newRelayReplayCache builds the same cache struct with relay-throughput
// sizing and the refuse-on-overflow policy. Keeping the cache struct generic
// instead of forking it means every future improvement to eviction / expiry
// order lands on both caches automatically.
func newRelayReplayCache() *artReplayCache {
	return newRelayReplayCacheWithParams(relayReplayCacheSize, relayReplayCacheTTL, "NHP_RLY", time.Now)
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

// newRelayReplayCacheWithParams is the relay-side parallel of
// newARTReplayCacheWithParams; it sets rejectWhenFull=true so the cache
// refuses new envelopes instead of evicting live entries when full. The
// signature matches the ART constructor so the only difference between the
// two is the overflow policy.
func newRelayReplayCacheWithParams(maxEntries int, ttl time.Duration, name string, now func() time.Time) *artReplayCache {
	c := newARTReplayCacheWithParams(maxEntries, ttl, name, now)
	c.rejectWhenFull = true
	return c
}

// Admit records an authenticated (peer pubkey, transaction ID, send time)
// tuple and returns the outcome. The decision is in three fixed steps so
// the result is always well-defined even under capacity pressure:
//  1. evictExpired(now)        — freshen the cache first
//  2. key already present?      — replayAdmitDuplicate (replay is rejected)
//  3. at capacity?             — rejectWhenFull decides:
//     false: evict oldest, then insert (replayAdmitNew)
//     true:  count it and return replayAdmitCacheFull
//
// Step 2 deliberately runs before step 3 so a replay of a RETAINED entry
// under load is classified as Duplicate, not CacheFull — the latter is
// reserved for "new envelope that we cannot admit right now".
func (c *artReplayCache) Admit(peerPubkey []byte, txid uint64, sendTime int64) replayAdmit {
	if c == nil || c.maxEntries <= 0 || c.ttl <= 0 {
		return replayAdmitInvalid
	}
	if len(peerPubkey) != core.PublicKeySize && len(peerPubkey) != core.PublicKeySizeEx {
		return replayAdmitInvalid
	}

	var key artReplayKey
	copy(key.peerPubkey[:], peerPubkey)
	key.peerPubkeyLen = uint8(len(peerPubkey))
	key.txid = txid
	key.sendTime = sendTime

	// Logging is deferred until after unlock so the critical section stays
	// tight. Both branches share the same 1-minute throttle via lastWarning
	// so a sustained overflow produces at most one log line per minute.
	var (
		logEvicted  bool   // ART path: Warning "... live entries evicted ..."
		logRejected bool   // RLY path: Critical "... envelopes are being REFUSED ..."
		countSnap   uint64 // count at the time of the log line
	)
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
		if logEvicted {
			// Name is required so an overflow on the relay path
			// doesn't masquerade as ART (and vice versa) in the
			// log — operators reading "ART replay cache capacity
			// exceeded" would otherwise look in the wrong
			// subsystem to find the QPS ceiling.
			log.Warning("%s replay cache capacity %d exceeded; live entries evicted (%d total), replay window shortened", c.name, c.maxEntries, countSnap)
		} else if logRejected {
			log.Critical("%s replay cache full at %d entries; NEW envelopes are being REFUSED (%d total) — raise RelayReplayCacheEntries", c.name, c.maxEntries, countSnap)
		}
	}()

	now := c.now()
	c.evictExpired(now)
	if _, ok := c.entries[key]; ok {
		return replayAdmitDuplicate
	}
	if len(c.entries) >= c.maxEntries {
		if c.rejectWhenFull {
			c.capacityRejections++
			if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
				c.lastWarning = now
				logRejected = true
				countSnap = c.capacityRejections
			}
			return replayAdmitCacheFull
		}
		// ART path: evict the oldest live entry and proceed with insertion.
		c.capacityEvictions++
		if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
			c.lastWarning = now
			logEvicted = true
			countSnap = c.capacityEvictions
		}
		c.removeOldest()
	}
	c.order.PushBack(artReplayEntry{key: key, expiresAt: now.Add(c.ttl)})
	c.entries[key] = struct{}{}
	return replayAdmitNew
}

// MarkSeen is the bool-returning wrapper used by existing call sites and
// tests. It preserves the original semantics: true on first observation,
// false otherwise (whether duplicate, full, or invalid). The richer
// three-way outcome is available via Admit for new call sites that need to
// distinguish them.
func (c *artReplayCache) MarkSeen(peerPubkey []byte, txid uint64, sendTime int64) bool {
	return c.Admit(peerPubkey, txid, sendTime) == replayAdmitNew
}

// snapshotForMetrics returns the three counters the metrics layer exposes,
// taken under c.mu so the values are mutually consistent. Used by the
// nhp_server_art_replay_cache_evictions_total /
// nhp_server_relay_replay_cache_{full_rejections_total,entries} closures.
func (c *artReplayCache) snapshotForMetrics() (evictions, rejections uint64, size int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.capacityEvictions, c.capacityRejections, len(c.entries)
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
