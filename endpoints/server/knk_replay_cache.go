package server

import (
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
	// Size for the peak knock rate and configured freshness window.
	knkReplayCacheSize = 100_000
	// knkReplayCacheMaxSize is the largest accepted KNKReplayCacheEntries;
	// docs/deploy.md and the server config template quote it.
	knkReplayCacheMaxSize        = 1_000_000
	knkReplayCacheMaxPeerEntries = 512
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
	lru              *simplelru.LRU[knkReplayIdentity, time.Time]
	ttl              time.Duration
	capacity         int
	peerCapacity     int
	peerEntries      map[string]int
	capacityRefusals atomic.Uint64
}

type knkReplayIdentity struct {
	peer     string
	txid     uint64
	sendTime int64
}

// newKNKReplayCache uses the default capacity and freshness window.
func newKNKReplayCache() *knkReplayCache {
	return newKNKReplayCacheWithParams(knkReplayCacheSize, knkReplayTTL(core.KNKRecvStalenessFloorSeconds))
}

// knkReplayTTL covers the configured packet age, maximum clock lead, and a margin.
func knkReplayTTL(windowSeconds int) time.Duration {
	if windowSeconds <= 0 {
		windowSeconds = core.KNKRecvStalenessFloorSeconds
	}
	return time.Duration(windowSeconds+core.ARTRecvFutureSkewSeconds+knkReplayCacheMarginSeconds) * time.Second
}

// newKNKReplayCacheWithParams allows short TTLs and small capacities in tests.
func newKNKReplayCacheWithParams(size int, ttl time.Duration) *knkReplayCache {
	c := &knkReplayCache{
		ttl: ttl, capacity: size, peerCapacity: min(knkReplayCacheMaxPeerEntries, max(1, size/8)),
		peerEntries: make(map[string]int),
	}
	var err error
	c.lru, err = simplelru.NewLRU[knkReplayIdentity, time.Time](size, nil)
	if err != nil {
		panic(err)
	} // sizes are validated at startup; tests use positive sizes.
	return c
}

// MarkSeen records the (peerPubkey, txid, sendTime) triple and returns
// an empty drop stage on success, or the reason for refusing it.
//
// Caller contract — peerPubkey and sendTime must both be the
// AEAD-authenticated values from PacketParserData (ppd.RemotePubKey and
// ppd.RemoteSendTime), and the caller must have completed body
// decryption. MarkSeen is called from HandleKnockRequest, which runs
// after the body AEAD has been verified: that ordering is what keeps a
// forged-but-header-valid packet from poisoning a real agent's triple.
//
// Curve and GMSM use different authenticated public-key lengths. Any other
// length fails closed, so a caller
// that skips the upstream length guard still cannot key every empty
// pubkey under one slot.
//
// A nil receiver also fails closed, the same as the
// ART replay cache: a server with no replay cache must not accept knocks
// it cannot protect. Start() always builds the cache before the device
// starts, so this only fires on a mis-wired server.
func (c *knkReplayCache) MarkSeen(peerPubkey []byte, txid uint64, sendTime int64) string {
	if c == nil {
		return "knk_replay"
	}
	if len(peerPubkey) != core.PublicKeySize && len(peerPubkey) != core.PublicKeySizeEx {
		return "knk_replay"
	}

	peer := string(peerPubkey)
	key := knkReplayIdentity{peer: peer, txid: txid, sendTime: sendTime}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Equal TTLs keep monotonic expiries in insertion order. Keep the wall-clock
	// deadline too because the receive freshness check uses the wall clock.
	now := time.Now()
	for {
		oldest, expires, exists := c.lru.GetOldest()
		if !exists || now.Before(expires) || now.UnixNano() < expires.UnixNano() {
			break
		}
		c.lru.RemoveOldest()
		c.peerEntries[oldest.peer]--
		if c.peerEntries[oldest.peer] == 0 {
			delete(c.peerEntries, oldest.peer)
		}
	}
	if _, exists := c.lru.Peek(key); exists {
		return "knk_replay"
	}
	if c.lru.Len() == c.capacity || c.peerEntries[peer] >= c.peerCapacity {
		c.capacityRefusals.Add(1)
		return "knk_replay_capacity"
	}
	c.lru.Add(key, now.Add(c.ttl))
	c.peerEntries[peer]++
	return ""
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

// CapacityRefusals reports how many new knocks reached the server or peer capacity limit.
func (c *knkReplayCache) CapacityRefusals() uint64 {
	if c == nil {
		return 0
	}
	return c.capacityRefusals.Load()
}
