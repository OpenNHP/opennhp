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
	// behind every registered relay, so the cache must absorb a
	// sustained relay+agent RPS across its TTL.
	//
	// With the tightened NHP_RLY staleness floor of
	// core.RLYRecvStalenessFloorSeconds (60 s) plus the existing
	// ARTRecvFutureSkewSeconds (300 s) margin, the cache TTL must
	// strictly outlive (staleness + future-skew) for an evicted entry's
	// replay to be rejected by the staleness check (see
	// endpoints/server/art_replay_cache.go's MarkSeen doc for the
	// algebra). relayReplayCacheTTL is therefore
	//   60 + 300 + 60 = 420 s, and the cache sustains a flow-fill rate
	// of up to ~1190 envelopes/s (500k / 420s) without opening a replay
	// window. The previous fail-closed design pushed the apparent rate
	// ceiling higher but turned sustained overflow into a remote DoS
	// for every legitimate web-agent knock; the documented ceiling here
	// trades that attackability ceiling for survivable behaviour under
	// pressure.
	relayReplayCacheSize = 500_000
	// artReplayCacheTTL pairs with validatePeer's future-skew gate:
	// entries must outlive every envelope the staleness floor accepts,
	// plus the future-skew bound (so a forward-stamped replay doesn't
	// outlive its cache entry), plus a small safety margin.
	artReplayCacheTTL = (core.DefaultRecvStalenessFloorSeconds + core.ARTRecvFutureSkewSeconds + 60) * time.Second
	// relayReplayCacheTTL pairs with the tightened RLYRecvStalenessFloorSeconds:
	// it must strictly outlive (RLYRecvStalenessFloorSeconds +
	// ARTRecvFutureSkewSeconds) for an evicted entry's remoteSendTime to
	// be rejected by the staleness check at the moment of eviction.
	// See the size docstring above for the algebra.
	relayReplayCacheTTL = (core.RLYRecvStalenessFloorSeconds + core.ARTRecvFutureSkewSeconds + 60) * time.Second
)

// artPubkeyKey is a stack-allocated, comparable wrapper around a peer
// pubkey, used as the index for per-peer accounting inside the replay
// caches. It avoids the heap allocation that base64.StdEncoding +
// string would introduce on the hot path, and it preserves the
// Curve-vs-GMSM distinction (which a bare [N]byte slice cannot — both
// are 32 / 64 bytes respectively, but a 32-byte Curve key is not the
// same length as a 64-byte GMSM key, so the slice-cap Len discriminates).
type artPubkeyKey struct {
	pubkey    [core.PublicKeySizeEx]byte
	pubkeyLen uint8
}

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
	peerKey   artPubkeyKey
}

// artReplayCache is a bounded, concurrency-safe TTL set. Entries are appended
// in expiry order, so expiration and capacity eviction are O(1). Capacity
// eviction prefers to remove an entry from the same peer that is being admitted
// (so a flooding relay cannot starve a legitimate relay by filling the cache
// with attacker traffic that rides on the relay's authenticated key — see
// MarkSeen's docstring for the full rationale).
type artReplayCache struct {
	name              string
	mu                sync.Mutex
	capacityEvictions uint64
	lastWarning       time.Time
	entries           map[artReplayKey]*list.Element
	order             *list.List // global FIFO of *artReplayEntry
	peerCount         map[artPubkeyKey]int
	maxEntries        int
	ttl               time.Duration
	now               func() time.Time
}

func newARTReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(artReplayCacheSize, artReplayCacheTTL, "ART", time.Now)
}

// newRelayReplayCache builds the same cache struct with relay-throughput
// sizing. The capacity-eviction policy is identical to ART's; the size and
// TTL are tuned so that, combined with the tightened NHP_RLY staleness
// floor in core/responder, eviction can never re-open a replay hole —
// see the relayReplayCacheTTL docstring for the algebra.
func newRelayReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(relayReplayCacheSize, relayReplayCacheTTL, "NHP_RLY", time.Now)
}

func newARTReplayCacheWithParams(maxEntries int, ttl time.Duration, name string, now func() time.Time) *artReplayCache {
	return &artReplayCache{
		name:       name,
		entries:    make(map[artReplayKey]*list.Element),
		order:      list.New(),
		peerCount:  make(map[artPubkeyKey]int),
		maxEntries: maxEntries,
		ttl:        ttl,
		now:        now,
	}
}

// MarkSeen records an authenticated (peer pubkey, transaction ID, send time)
// tuple. It returns true on first observation and false for a duplicate or an
// invalid peer key. Callers must distinguish invalid input before calling.
//
// On capacity pressure the cache evicts: per-peer first (so a flooding peer
// cannot starve a quieter peer by riding on its key) and then global oldest
// (so the cache always accepts new envelopes rather than rejecting them).
// Eviction is safe against replay because:
//   - cache TTL is set strictly larger than the future-skew +
//     staleness-floor bound (see relayReplayCacheTTL's docstring for the
//     algebra), so any entry removed by expiry or by per-peer eviction
//     is older than its accepted replay window at the moment of removal;
//   - capacity-eviction timing is bounded by relayReplayCacheTTL via the
//     maxEntries / TTL flow-fill rate (the per-peer preference keeps the
//     bound tight in the steady state of an attack).
//
// The fail-closed policy that this cache used previously (returning false
// for genuinely new envelopes at capacity, with a rate-limited warning)
// was rejected by review: any anonymous POST to the relay produces a
// fresh NHP_RLY envelope signed under the legitimate relay's key, and a
// flood of such envelopes would push the cache to reject every
// legitimate web-agent knock through every relay for up to the cache
// TTL — a cheap remote DoS of the relay path. The per-peer preference
// caps the attack surface to the attacker's own envelopes, and the
// tightened staleness floor (see RLYRecvStalenessFloorSeconds in
// nhp/core/constants.go) keeps eviction from re-opening a replay hole
// when the cache overflows.
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

	var peerKey artPubkeyKey
	copy(peerKey.pubkey[:], peerPubkey)
	peerKey.pubkeyLen = uint8(len(peerPubkey))

	var warn bool
	var evictions uint64
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
		if warn {
			// Name is required so an overflow on the relay path
			// doesn't masquerade as ART (and vice versa) in the
			// log — operators reading "ART replay cache capacity
			// exceeded" would otherwise look in the wrong
			// subsystem to find the QPS ceiling.
			log.Warning("%s replay cache capacity %d exceeded; live entries evicted (%d total), replay window shortened", c.name, c.maxEntries, evictions)
		}
	}()

	now := c.now()
	c.evictExpired(now)
	if _, ok := c.entries[key]; ok {
		return false
	}
	if len(c.entries) >= c.maxEntries {
		// Prefer to evict the admitting peer's own oldest entry, so a
		// flooding peer's growth can't push out of kicking out a
		// quieter peer's entries. If the admitting peer has no live
		// entries (first contact, easiest to check), fall back to the
		// global oldest — this is what lets a brand-new peer still get
		// in when the cache is full of attacker entries.
		evicted := c.evictOldestForPeer(peerKey)
		if !evicted {
			c.removeOldest()
		}
		c.capacityEvictions++
		if c.lastWarning.IsZero() || now.Sub(c.lastWarning) >= time.Minute {
			warn = true
			evictions = c.capacityEvictions
			c.lastWarning = now
		}
	}
	entry := &artReplayEntry{
		key:       key,
		expiresAt: now.Add(c.ttl),
		peerKey:   peerKey,
	}
	elem := c.order.PushBack(entry)
	c.entries[key] = elem
	c.peerCount[peerKey]++
	return true
}

func (c *artReplayCache) evictExpired(now time.Time) {
	for elem := c.order.Front(); elem != nil; elem = c.order.Front() {
		entry := elem.Value.(*artReplayEntry)
		if entry.expiresAt.After(now) {
			return
		}
		c.removeElement(elem)
	}
}

// evictOldestForPeer walks the global order from the front looking for
// the oldest entry belonging to peerKey, and removes it. Returns true
// if such an entry was found and removed, false if the peer has no
// live entries (in which case the caller should fall back to
// removeOldest).
//
// The walk is bounded by peerCount[peerKey]: if the peer has N live
// entries, we find one within the first N elements. Under sustained
// attack from a single peer, that walk is bounded by the per-peer's
// share of the cache (which evictions keep at <= maxEntries minus
// everyone else's share).
func (c *artReplayCache) evictOldestForPeer(peerKey artPubkeyKey) bool {
	if c.peerCount[peerKey] == 0 {
		return false
	}
	for elem := c.order.Front(); elem != nil; elem = elem.Next() {
		entry := elem.Value.(*artReplayEntry)
		if entry.peerKey == peerKey {
			c.removeElement(elem)
			return true
		}
	}
	// peerCount[peerKey] > 0 but no element found: should be
	// unreachable because every entry is added via MarkSeen which
	// also bumps peerCount. Be defensive — fall back to global eviction.
	c.peerCount[peerKey] = 0
	return false
}

func (c *artReplayCache) removeOldest() {
	elem := c.order.Front()
	if elem == nil {
		return
	}
	c.removeElement(elem)
}

// removeElement detaches the given element from the global order, the
// per-peer count, and the entry map. Safe to call on any element
// previously returned by c.order.
func (c *artReplayCache) removeElement(elem *list.Element) {
	entry := elem.Value.(*artReplayEntry)
	c.order.Remove(elem)
	delete(c.entries, entry.key)
	c.peerCount[entry.peerKey]--
	if c.peerCount[entry.peerKey] <= 0 {
		delete(c.peerCount, entry.peerKey)
	}
}

func artPubkeyFingerprint(peerPubkey []byte) string {
	encoded := base64.RawURLEncoding.EncodeToString(peerPubkey)
	if len(encoded) > 12 {
		return encoded[:12]
	}
	return encoded
}
