package server

import (
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

const (
	// knockReplayCacheSize bounds the knock dedupe set.
	//
	// Sized against the window rather than against a rate: an entry lives
	// knockReplayCacheTTL, so the cache holds every knock accepted in that
	// span. At 600s + skew that is roughly 1.5k knocks/second before
	// capacity pressure, which is far above what a knock-gated deployment
	// produces — a knock is one packet per agent per access, not per
	// connection.
	//
	// Capacity eviction matters more here than on the ART path, because
	// evicting a live entry is what lets a replay back in. A flood of the
	// *same* packet cannot cause it (same key, no new entry). Reaching it
	// needs a stream of distinct, correctly authenticated knocks, which
	// needs a registered private key; the cache logs when it happens so
	// that an operator sees the window shortening rather than discovering
	// it from a replay that got through.
	knockReplayCacheSize = 1_000_000

	// knockReplayCacheTTL must outlive every knock the staleness floor
	// still accepts, or a packet could leave the cache while remaining
	// replayable. recvStalenessFloor gives knocks
	// DefaultRecvStalenessFloorSeconds; the skew and margin terms mirror
	// artReplayCacheTTL.
	knockReplayCacheTTL = (core.DefaultRecvStalenessFloorSeconds + core.ARTRecvFutureSkewSeconds + 60) * time.Second
)

// newKnockReplayCache builds the cross-connection dedupe set for knock
// messages.
//
// It reuses the bounded TTL set written for the ART path
// (art_replay_cache.go): same key shape — authenticated sender key,
// transaction id, sender timestamp — same O(1) expiry and eviction, same
// capacity-eviction reporting. Only the name, size and TTL differ, which
// is why that type takes them as parameters.
func newKnockReplayCache() *artReplayCache {
	return newARTReplayCacheWithParams(knockReplayCacheSize, knockReplayCacheTTL, "KNK", time.Now)
}

// knockReplayHeaderTypes are the message types that reach
// HandleKnockRequest. All four carry the same replay exposure: the
// monotonic timestamp check in the responder is per connection, and a
// replayer choosing a fresh source port gets a fresh connection with a
// reset check.
func isKnockReplayHeaderType(t int) bool {
	switch t {
	case core.NHP_KNK, core.NHP_RKN, core.NHP_EXT, core.DHP_KNK:
		return true
	}
	return false
}
