package server

import (
	"encoding/base64"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/golang-lru/v2/simplelru"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// KNK replay protection.
//
// A knock (NHP_KNK / NHP_RKN / NHP_EXT / DHP_KNK) is the one message
// whose replay has a SECURITY consequence: the server derives the AC
// firewall grant from the packet's SOURCE ADDRESS, so a captured knock
// replayed from a different address gets that address admitted. The
// per-connection monotonic gate in nhp/core/responder.go does not stop
// that — a replay from a new source address lands on a fresh
// ConnectionData whose LastRemoteSendTime is 0.
//
// This cache is the KNK counterpart of the ART cache (art_replay_cache.go)
// and the AC's AOP cache (endpoints/ac/aop_replay_cache.go): a bounded
// LRU of authenticated (peer pubkey, transaction id, send time) triples
// with lazy expiry.
//
// The key includes the AEAD-authenticated send time, so a legitimate
// retransmission is never mistaken for a replay: the agent takes a fresh
// transaction id (device.NextCounterIndex()) and a fresh timestamp for
// every packet it builds, so a retry produces a DIFFERENT triple. A
// byte-identical replay produces the SAME triple and is rejected.
//
// State is process-local: a restart clears it. The staleness floor
// (nhp/core/responder.go, KNKRecvStalenessFloorSeconds) bounds the
// remaining replay window after a restart. A full cache refuses new knocks and retains live entries.
const (
	// knkReplayCacheSize is the default number of KNK replay entries.
	// A knock is one packet per agent access attempt, so even a busy
	// server stays well under 100k entries at a 660 s TTL.
	knkReplayCacheSize = 100_000
	// knkReplayCacheMaxSize is the largest accepted KNKReplayCacheEntries;
	// docs/deploy.md and the server config template quote it.
	knkReplayCacheMaxSize = 1_000_000
	// knkReplayCacheMarginSeconds is added to the effective staleness
	// window to derive the TTL. The window is measured from the packet's
	// own send time, so an entry must outlive the window by a margin to
	// cover a packet that arrives late in that window.
	knkReplayCacheMarginSeconds = 60
)

// knkReplayCache is a bounded set with lazy expiry. It owns no goroutine.
// The outer mutex makes lookup and insertion one operation.
type knkReplayCache struct {
	mu               sync.Mutex
	lru              *simplelru.LRU[string, time.Time]
	ttl              time.Duration
	capacity         int
	capacityRefusals atomic.Uint64
}

// newKNKReplayCache returns a cache sized for the default entry count and
// the TTL derived from the protocol's default KNK staleness window.
func newKNKReplayCache() *knkReplayCache {
	return newKNKReplayCacheWithParams(knkReplayCacheSize, knkReplayTTL(core.KNKRecvStalenessFloorSeconds))
}

// knkReplayTTL derives the cache TTL from the EFFECTIVE staleness window
// in seconds: the window plus a margin. Shortening the window therefore
// shortens how long an entry is retained, which is the point of making
// the window configurable.
func knkReplayTTL(windowSeconds int) time.Duration {
	if windowSeconds <= 0 {
		windowSeconds = core.KNKRecvStalenessFloorSeconds
	}
	return time.Duration(windowSeconds+knkReplayCacheMarginSeconds) * time.Second
}

// newKNKReplayCacheWithParams allows short TTLs and small capacities in tests.
func newKNKReplayCacheWithParams(size int, ttl time.Duration) *knkReplayCache {
	c := &knkReplayCache{ttl: ttl, capacity: size}
	var err error
	c.lru, err = simplelru.NewLRU[string, time.Time](size, nil)
	if err != nil {
		panic(err)
	} // sizes are validated at startup; tests use positive sizes.
	return c
}

// MarkSeen records the (peerPubkey, txid, sendTime) triple and returns
// true on first observation, false if the triple is already in the cache.
//
// Caller contract — peerPubkey and sendTime must both be the
// AEAD-authenticated values from PacketParserData (ppd.RemotePubKey and
// ppd.RemoteSendTime), and the caller must have completed body
// decryption. MarkSeen is called from HandleKnockRequest, which runs
// after the body AEAD has been verified: that ordering is what keeps a
// forged-but-header-valid packet from poisoning a real agent's triple.
//
// Curve and GMSM use different authenticated public-key lengths. Any other
// length fails closed (returns false, i.e. "already seen"), so a caller
// that skips the upstream length guard still cannot key every empty
// pubkey under one slot.
//
// A nil receiver also fails closed (returns false), the same shape as the
// ART replay cache: a server with no replay cache must not accept knocks
// it cannot protect. Start() always builds the cache before the device
// starts, so this only fires on a mis-wired server.
func (c *knkReplayCache) MarkSeen(peerPubkey []byte, txid uint64, sendTime int64) bool {
	if c == nil {
		return false
	}
	if len(peerPubkey) != core.PublicKeySize && len(peerPubkey) != core.PublicKeySizeEx {
		return false
	}

	key := knkReplayKey(peerPubkey, txid, sendTime)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Round(0) strips the monotonic reading so expiry is compared on the
	// wall clock, the same clock the responder's staleness gate uses.
	now := time.Now().Round(0)
	if expires, exists := c.lru.Peek(key); exists {
		if now.Before(expires) {
			return false
		}
		c.lru.Remove(key)
	}
	if c.lru.Len() == c.capacity {
		_, oldestExpiry, _ := c.lru.GetOldest()
		if now.Before(oldestExpiry) {
			c.capacityRefusals.Add(1)
			return false
		}
		c.lru.RemoveOldest()
	}
	// Anchor expiry to the later of receive time and the authenticated send
	// time. The staleness gate is measured from the send time, so a
	// forward-dated knock (fast sender clock) would otherwise keep passing
	// it after a receive-anchored entry had expired, and replay.
	anchor := now
	if sent := time.Unix(0, sendTime); sent.After(anchor) {
		anchor = sent
	}
	c.lru.Add(key, anchor.Add(c.ttl))
	return true
}

// Len is the current entry count, exposed for tests. Takes the same mutex
// as MarkSeen so a Len() call never observes the cache between MarkSeen's
// Get and Add.
func (c *knkReplayCache) Len() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.Len()
}

// CapacityRefusals reports how many new knocks were refused at capacity.
func (c *knkReplayCache) CapacityRefusals() uint64 {
	if c == nil {
		return 0
	}
	return c.capacityRefusals.Load()
}

// knkReplayKey concatenates the base64 pubkey, the decimal txid and the
// decimal sendTime. The '|' separator cannot make two keys collide: the
// pubkey prefix has a known length and the two decimal fields contain no
// '|'.
func knkReplayKey(peerPubkey []byte, txid uint64, sendTime int64) string {
	return base64.StdEncoding.EncodeToString(peerPubkey) + "|" +
		strconv.FormatUint(txid, 10) + "|" + strconv.FormatInt(sendTime, 10)
}
