package server

import (
	"sync"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func artTestPubkey(fill byte, sizes ...int) []byte {
	size := core.PublicKeySize
	if len(sizes) > 0 {
		size = sizes[0]
	}
	key := make([]byte, size)
	for i := range key {
		key[i] = fill
	}
	return key
}

func TestARTReplayCacheSupportsCurveAndGMSMKeys(t *testing.T) {
	for _, size := range []int{core.PublicKeySize, core.PublicKeySizeEx} {
		cache := newARTReplayCache()
		key := artTestPubkey(1, size)
		if !cache.MarkSeen(key, 1, 1) {
			t.Fatalf("first %d-byte key rejected", size)
		}
		if cache.MarkSeen(key, 1, 1) {
			t.Fatalf("duplicate %d-byte key accepted", size)
		}
	}
}

func TestARTReplayCacheAcceptsOutOfOrderFreshPackets(t *testing.T) {
	cache := newARTReplayCache()
	key := artTestPubkey(1)
	if !cache.MarkSeen(key, 2, 200) {
		t.Fatal("newer packet rejected")
	}
	if !cache.MarkSeen(key, 1, 100) {
		t.Fatal("fresh packet with an earlier timestamp must survive burst reordering")
	}
	if cache.MarkSeen(key, 1, 100) {
		t.Fatal("replay of the reordered packet accepted")
	}
}

func TestARTReplayCacheScopesAndRejectsDuplicates(t *testing.T) {
	now := time.Unix(100, 0)
	cache := newARTReplayCacheWithParams(8, time.Minute, "test", true, func() time.Time { return now })
	keyA, keyB := artTestPubkey(1), artTestPubkey(2)

	if !cache.MarkSeen(keyA, 7, 1000) {
		t.Fatal("first packet rejected")
	}
	if cache.MarkSeen(keyA, 7, 1000) {
		t.Fatal("duplicate accepted")
	}
	if !cache.MarkSeen(keyB, 7, 1000) {
		t.Fatal("different peer key rejected")
	}
	if !cache.MarkSeen(keyA, 7, 1001) {
		t.Fatal("fresh send time after counter reset rejected")
	}
	if !cache.MarkSeen(keyA, 8, 1000) {
		t.Fatal("different transaction rejected")
	}
}

func TestARTReplayCacheExpiresAndEvicts(t *testing.T) {
	now := time.Unix(100, 0)
	cache := newARTReplayCacheWithParams(2, time.Second, "test", true, func() time.Time { return now })
	key := artTestPubkey(1)

	cache.MarkSeen(key, 1, 1)
	cache.MarkSeen(key, 2, 2)
	cache.MarkSeen(key, 3, 3)
	if cache.capacityEvictions != 1 || cache.lastWarning.IsZero() {
		t.Fatal("live eviction was not reported")
	}
	if !cache.MarkSeen(key, 1, 1) {
		t.Fatal("capacity eviction did not release oldest key")
	}

	now = now.Add(2 * time.Second)
	if !cache.MarkSeen(key, 3, 3) {
		t.Fatal("expired key remained blocked")
	}
}

func TestARTReplayCacheFailsClosedOnInvalidKey(t *testing.T) {
	cache := newARTReplayCache()
	if cache.MarkSeen(nil, 1, 1) ||
		cache.MarkSeen(make([]byte, core.PublicKeySize-1), 1, 1) ||
		cache.MarkSeen(make([]byte, core.PublicKeySizeEx+1), 1, 1) {
		t.Fatal("invalid peer key accepted")
	}
}

func TestARTReplayCacheConcurrentDuplicate(t *testing.T) {
	cache := newARTReplayCache()
	key := artTestPubkey(1)
	const workers = 64
	var wg sync.WaitGroup
	wg.Add(workers)
	accepted := make(chan struct{}, workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			if cache.MarkSeen(key, 1, 1) {
				accepted <- struct{}{}
			}
		}()
	}
	wg.Wait()
	close(accepted)
	if got := len(accepted); got != 1 {
		t.Fatalf("accepted %d concurrent copies, want 1", got)
	}
}

// TestARTReplayCacheRejectsAtCapacityPreservesLive pins the
// relay-cache behavior (and the cache-construction knob that drives
// it): when the cache is full AND every live entry is still inside
// the staleness floor, a fresh key must be REFUSED rather than
// evicting a live entry. The eviction-on-full behavior (the legacy ART
// path) is checked separately in TestARTReplayCacheExpiresAndEvicts
// above; the safe-eviction behavior (a full cache of stale entries
// accepts a fresh key) is checked in
// TestARTReplayCacheSafeEvictsStaleAtCapacity below.
//
// The scenario mirrors the relay-RPS math: with relayReplayCacheSize =
// 500_000 entries and a sustained ~1,000 envelopes/s, a capacity-evicting
// cache only retains ~500 seconds of traffic, below the 600-second
// staleness window. A captured envelope can therefore pass the outer
// replay check again after its shadow left the cache and reach a fresh
// inner connection. Safe eviction refuses when every entry is still
// replayable, and reaps stale entries once they age past the floor.
func TestARTReplayCacheRejectsAtCapacityPreservesLive(t *testing.T) {
	const cap = 4
	fakeNow := time.Unix(1_000_000, 0)
	clock := func() time.Time { return fakeNow }
	cache := newARTReplayCacheWithParams(cap, time.Hour, "NHP_RLY-test", false, clock)

	// All entries sit INSIDE the staleness floor (sendTime is only
	// 100s behind the fake clock, well under the 600s floor). None
	// is safely evictable, so the only correct response at full
	// capacity is to refuse the new envelope.
	recentSendTime := fakeNow.Add(-100 * time.Second).Unix()

	// One specific entry we want to verify survives a full-cache
	// pressure round — the first one we insert is the "retained
	// envelope" from the test brief.
	retainedKey := artTestPubkey(1)
	if !cache.MarkSeen(retainedKey, 100, recentSendTime) {
		t.Fatal("retained envelope was rejected on first observation")
	}

	// Fill the rest of the cache with entries the test doesn't follow up
	// on; each must be accepted (cache has spare capacity).
	for i := 0; i < cap-1; i++ {
		key := artTestPubkey(byte(2 + i))
		if !cache.MarkSeen(key, uint64(200+i), recentSendTime) {
			t.Fatalf("fill entry %d rejected while cache was under capacity", i)
		}
	}

	cache.mu.Lock()
	if n := len(cache.entries); n != cap {
		cache.mu.Unlock()
		t.Fatalf("pre-fill cache has %d entries, want %d", n, cap)
	}
	cache.mu.Unlock()

	// A brand-new key at capacity must be REFUSED — every live entry
	// is still inside the staleness floor and would lose its
	// replay protection if evicted.
	if cache.MarkSeen(artTestPubkey(99), 999, recentSendTime) {
		t.Fatal("fresh key accepted at full capacity of live entries — live entries would be evicted; this is the regression")
	}

	cache.mu.Lock()
	if cache.capacityEvictions != 0 {
		cache.mu.Unlock()
		t.Fatalf("capacityEvictions = %d, want 0 (rejection path must not bump the eviction counter)", cache.capacityEvictions)
	}
	if cache.capacityRejections != 1 {
		cache.mu.Unlock()
		t.Fatalf("capacityRejections = %d, want 1", cache.capacityRejections)
	}
	if n := len(cache.entries); n != cap {
		cache.mu.Unlock()
		t.Fatalf("cache grew past capacity on refused insert: len(entries)=%d, want %d", n, cap)
	}
	cache.mu.Unlock()

	// The retained envelope must STILL be in the cache — capacity
	// rejection must not have evicted it.
	cache.mu.Lock()
	_, retainedPresent := cache.entries[artReplayKey{
		peerPubkey:    func() [core.PublicKeySizeEx]byte { var k [core.PublicKeySizeEx]byte; copy(k[:], retainedKey); return k }(),
		peerPubkeyLen: uint8(len(retainedKey)),
		txid:          100,
		sendTime:      recentSendTime,
	}]
	cache.mu.Unlock()
	if !retainedPresent {
		t.Fatal("retained envelope was evicted from the cache — capacity rejection path must not touch live entries")
	}

	// The brief's load-bearing assertion: a replay of the retained
	// envelope must STILL be rejected. Pre-fix, the retained envelope
	// had been evicted at step (capacity-refusal) — well, it had been
	// the oldest entry, so the pre-fix eviction would have removed it
	// and a follow-up MarkSeen of the same key would return TRUE (as a
	// fresh acceptance). That's the regression: a captured NHP_RLY
	// envelope reaches the inner conn a second time.
	if cache.MarkSeen(retainedKey, 100, recentSendTime) {
		t.Fatal("replay of retained envelope accepted — captured envelope could reach a fresh inner connection (the regression)")
	}
}

// TestARTReplayCacheRejectsAtCapacityIsDistinctFromDuplicate pins the
// counter distinction so dashboards can attribute the source of
// ErrServerDuplicateTransaction events correctly. A duplicate
// MarkSeen at full capacity stays rejected AND bumps neither counter
// — it short-circuits before the capacity check.
func TestARTReplayCacheRejectsAtCapacityIsDistinctFromDuplicate(t *testing.T) {
	const cap = 2
	fakeNow := time.Unix(1_000_000, 0)
	clock := func() time.Time { return fakeNow }
	cache := newARTReplayCacheWithParams(cap, time.Hour, "NHP_RLY-test", false, clock)

	// Stay inside the staleness floor (well under 600s behind the
	// fake clock) so the safe-eviction path can't drop entries we
	// intend to keep.
	recentSendTime := fakeNow.Add(-100 * time.Second).Unix()

	if !cache.MarkSeen(artTestPubkey(1), 1, recentSendTime) ||
		!cache.MarkSeen(artTestPubkey(2), 2, recentSendTime) {
		t.Fatal("fill under capacity failed")
	}

	// Duplicate of an entry already in the cache. Must be rejected
	// but must NOT count as a capacity rejection (the cache didn't
	// refuse to make room — the key was already there).
	if cache.MarkSeen(artTestPubkey(1), 1, recentSendTime) {
		t.Fatal("duplicate accepted")
	}

	cache.mu.Lock()
	if cache.capacityRejections != 0 {
		cache.mu.Unlock()
		t.Fatalf("duplicate marked as capacity rejection: rejections=%d, want 0", cache.capacityRejections)
	}
	cache.mu.Unlock()

	// Now a fresh key at full capacity: this IS a capacity rejection.
	if cache.MarkSeen(artTestPubkey(3), 3, recentSendTime) {
		t.Fatal("fresh key accepted at capacity")
	}
	cache.mu.Lock()
	if cache.capacityRejections != 1 {
		cache.mu.Unlock()
		t.Fatalf("capacity rejection counter not bumped: rejections=%d, want 1", cache.capacityRejections)
	}
	cache.mu.Unlock()
}

// TestARTReplayCacheSafeEvictsStaleAtCapacity pins the safe-eviction
// path added to the NHP_RLY cache: when the cache is at capacity
// but every entry is already past the staleness floor, a fresh key
// must be accepted by dropping a stale entry rather than refused.
//
// This is the brief's "full cache of only stale-by-floor entries still
// accepts a fresh key" scenario — the regression the previous PR
// introduced (always-refuse at capacity) was that an attacker who
// filled the cache with stale-by-clock envelopes could keep any
// legitimate fresh envelope out for as long as the TTL allowed.
// ValidatePeer would already refuse a replay of any of those stale
// envelopes, so evicting them costs no protection; the cache should
// make room and accept the legitimate envelope.
func TestARTReplayCacheSafeEvictsStaleAtCapacity(t *testing.T) {
	const cap = 4
	fakeNow := time.Unix(1_000_000, 0)
	clock := func() time.Time { return fakeNow }
	cache := newARTReplayCacheWithParams(cap, time.Hour, "NHP_RLY-test", false, clock)

	// Every cache entry is past the staleness floor: sendTime is
	// 1000s behind the fake clock, 400s past the 600s floor.
	staleSendTime := fakeNow.Add(-1000 * time.Second).Unix()

	staleKeys := make([][]byte, cap)
	for i := 0; i < cap; i++ {
		staleKeys[i] = artTestPubkey(byte(1 + i))
		if !cache.MarkSeen(staleKeys[i], uint64(100+i), staleSendTime) {
			t.Fatalf("stale entry %d rejected on insertion", i)
		}
	}

	cache.mu.Lock()
	if n := len(cache.entries); n != cap {
		cache.mu.Unlock()
		t.Fatalf("pre-fill cache has %d entries, want %d", n, cap)
	}
	cache.mu.Unlock()

	// A fresh key arrives at capacity. The head entry is
	// staleKeys[0] with sendTime 1000s behind the clock — well
	// past the 600s staleness floor — so safe eviction must drop
	// it and accept the fresh envelope. This is the key property
	// the brief asks the test to assert.
	freshKey := artTestPubkey(99)
	freshSendTime := fakeNow.Unix()
	if !cache.MarkSeen(freshKey, 999, freshSendTime) {
		t.Fatal("fresh key refused at capacity of stale-by-floor entries — safe eviction must make room")
	}

	// Safe eviction is not "eviction" (the legacy counter) and not
	// "rejection" (the refuse-at-capacity counter); it is silent.
	cache.mu.Lock()
	if cache.capacityEvictions != 0 {
		cache.mu.Unlock()
		t.Fatalf("capacityEvictions = %d, want 0 (safe eviction must not bump the eviction counter)", cache.capacityEvictions)
	}
	if cache.capacityRejections != 0 {
		cache.mu.Unlock()
		t.Fatalf("capacityRejections = %d, want 0 (safe eviction must not refuse the envelope)", cache.capacityRejections)
	}
	if n := len(cache.entries); n != cap {
		cache.mu.Unlock()
		t.Fatalf("cache size after safe eviction = %d, want %d", n, cap)
	}
	cache.mu.Unlock()

	// The head entry (staleKeys[0]) is gone; the rest of the
	// stale entries plus the fresh envelope remain.
	for i, k := range staleKeys {
		key := artReplayKey{
			peerPubkey:    func() [core.PublicKeySizeEx]byte { var b [core.PublicKeySizeEx]byte; copy(b[:], k); return b }(),
			peerPubkeyLen: uint8(len(k)),
			txid:          uint64(100 + i),
			sendTime:      staleSendTime,
		}
		cache.mu.Lock()
		_, present := cache.entries[key]
		cache.mu.Unlock()
		wantPresent := i > 0
		if present != wantPresent {
			t.Fatalf("staleKeys[%d] presence = %v, want %v (safe eviction should drop only the head)", i, present, wantPresent)
		}
	}
	freshKeyEntry := artReplayKey{
		peerPubkey:    func() [core.PublicKeySizeEx]byte { var b [core.PublicKeySizeEx]byte; copy(b[:], freshKey); return b }(),
		peerPubkeyLen: uint8(len(freshKey)),
		txid:          999,
		sendTime:      freshSendTime,
	}
	cache.mu.Lock()
	_, freshPresent := cache.entries[freshKeyEntry]
	cache.mu.Unlock()
	if !freshPresent {
		t.Fatal("fresh envelope was not inserted into the cache after safe eviction")
	}
}

// TestARTReplayCacheSafeEvictsThenRefusesWhenOnlyFreshRemain pins the
// transition: once the cache has reaped its stale entries down to
// "every entry still inside the staleness floor", a subsequent fresh
// key MUST be refused rather than dropped. The safe-eviction policy
// never drops a still-replayable entry, so the only path back to
// capacity for a sustained fresh flood is the refusal path — which
// is bounded by the staleness floor (those entries will age past
// floor and become safely evictable on their own).
func TestARTReplayCacheSafeEvictsThenRefusesWhenOnlyFreshRemain(t *testing.T) {
	const cap = 2
	fakeNow := time.Unix(1_000_000, 0)
	clock := func() time.Time { return fakeNow }
	cache := newARTReplayCacheWithParams(cap, time.Hour, "NHP_RLY-test", false, clock)

	staleSendTime := fakeNow.Add(-1000 * time.Second).Unix()
	recentSendTime := fakeNow.Add(-100 * time.Second).Unix()

	if !cache.MarkSeen(artTestPubkey(1), 1, staleSendTime) ||
		!cache.MarkSeen(artTestPubkey(2), 2, staleSendTime) {
		t.Fatal("fill with stale entries failed")
	}

	// The first fresh arrival should be accepted via safe eviction
	// of artTestPubkey(1) — the head entry, which is stale.
	if !cache.MarkSeen(artTestPubkey(3), 3, recentSendTime) {
		t.Fatal("first fresh key refused despite stale entries being evictable")
	}

	// Now the cache contains artTestPubkey(2) (still stale) and
	// artTestPubkey(3) (recent — inside the staleness floor).
	// A second fresh arrival: the head is still stale, so safe
	// eviction still has work to do.
	if !cache.MarkSeen(artTestPubkey(4), 4, recentSendTime) {
		t.Fatal("second fresh key refused despite the remaining head being stale")
	}

	// Now the cache contains artTestPubkey(3) and artTestPubkey(4),
	// BOTH recent (inside the floor). The next fresh arrival has no
	// safely-evictable head and must be refused.
	if cache.MarkSeen(artTestPubkey(5), 5, recentSendTime) {
		t.Fatal("third fresh key accepted despite the cache being full of fresh entries — must refuse")
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.capacityRejections != 1 {
		t.Fatalf("capacityRejections = %d, want 1 (one refusal once the cache holds only fresh entries)", cache.capacityRejections)
	}
	if cache.capacityEvictions != 0 {
		t.Fatalf("capacityEvictions = %d, want 0 (safe-eviction path is silent on the eviction counter)", cache.capacityEvictions)
	}
	if n := len(cache.entries); n != cap {
		t.Fatalf("cache size after mixed safe+refuse = %d, want %d", n, cap)
	}
}

// TestARTReplayCacheRelayRestartRecoversByTTL pins the recovery path
// for an NHP_RLY relay whose process restarted: Device.counterIndex is
// a random seed in every process, so about half of relay restarts boot
// with a counter far below the highestTxid the server previously saw
// from this peer. The cache must not lock the relay out until the
// operator restarts nhp-server — entries must self-age out via TTL and
// the fresh counter/sendTime must be accepted without a manual reset.
//
// Two scenarios in one test, both anchored to the relay cache's TTL
// (~16 min in production, artReplayCacheTTL). The fixture uses a
// 1000-second TTL so we can advance fakeNow past it deterministically:
//
//  1. "counter below old highestTxid, sendTime ≤ highestSendTime":
//     the realistic worst case from the review — relay's clock is
//     behind the server's recorded highestSendTime after NTP step or
//     RTC reset, so the new envelope's sendTime is older than what we
//     recorded last. The cache must still accept it after TTL expiry.
//
//  2. "counter below old highestTxid, sendTime > highestSendTime":
//     relay clock recovered between the two sends; the new envelope
//     has both a lower txid AND a higher sendTime. The cache must
//     accept it (different key, the prior key already expired).
//
// Pre-fix the planned implementation used a sliding txid window with a
// sendTime-based "reset" branch (highestSendTime + skew). That branch
// would have rejected both scenarios after a relay restart with a
// regressed clock — the relay would have stayed blocked until
// nhp-server restarted. The TTL cache that the code actually uses
// ages entries out without depending on the relay's wall clock moving
// forward, so neither scenario needs a server restart.
func TestARTReplayCacheRelayRestartRecoversByTTL(t *testing.T) {
	const relayTTL = 1000 * time.Second
	fakeNow := time.Unix(2_000_000, 0)
	clock := func() time.Time { return fakeNow }
	cache := newARTReplayCacheWithParams(1_000, relayTTL, "NHP_RLY-test", false, clock)

	// Relay identity stays the same across the "restart" — only the
	// per-process counter seed and the wall clock move. The cache keys
	// on (pubkey, txid, sendTime); same pubkey exercises the restart
	// path because the prior (pubkey, txid, sendTime) tuples still
	// exist in the cache as recorded-history entries that must not
	// shadow the new ones.
	relayKey := artTestPubkey(0xAB, core.PublicKeySize)

	// Pre-restart: relay stamps its envelopes with high txid + sendTime.
	const preRestartTxid uint64 = 1 << 40 // realistic "well-aged" counter
	preRestartSendTime := fakeNow.Add(-50 * time.Second).Unix()
	if !cache.MarkSeen(relayKey, preRestartTxid, preRestartSendTime) {
		t.Fatalf("pre-restart envelope rejected")
	}

	// Advance past the relay TTL. Every pre-restart entry has aged out.
	// The relay's recorded (highestTxid, highestSendTime) effectively
	// become "unknown" from the cache's perspective — which is what
	// the planned sliding-window implementation would NOT have
	// recovered from without operator intervention.
	fakeNow = fakeNow.Add(relayTTL + time.Second)

	// Scenario 1: counter below pre-restart's, sendTime still behind
	// (relay clock has not caught up). This is the case the planned
	// "reset branch" would have refused — it requires
	// sendTime > highestSendTime, but here sendTime < highestSendTime
	// after NTP stepped the relay's clock back.
	postRestartTxid1 := uint64(1) // ~2^40 below preRestartTxid
	postRestartSendTime1 := fakeNow.Add(-10 * time.Second).Unix()
	if !cache.MarkSeen(relayKey, postRestartTxid1, postRestartSendTime1) {
		t.Fatal("post-restart envelope (sendTime ≤ old highestSendTime) rejected — relay would stay blocked until server restart")
	}

	// Advance again to simulate another restart cycle.
	fakeNow = fakeNow.Add(relayTTL + time.Second)

	// Scenario 2: counter below pre-restart's, sendTime ahead. Same
	// shape as a captured-and-replayed envelope would be on the wire,
	// but it's a fresh envelope — the prior key already aged out and
	// the new (txid, sendTime) tuple is brand new.
	postRestartTxid2 := uint64(2) // ~2^40 below preRestartTxid
	postRestartSendTime2 := fakeNow.Add(10 * time.Second).Unix()
	if !cache.MarkSeen(relayKey, postRestartTxid2, postRestartSendTime2) {
		t.Fatal("post-restart envelope (sendTime > old highestSendTime) rejected — relay would stay blocked until server restart")
	}

	// Sanity: within the TTL window, a true replay of either
	// post-restart envelope is still rejected. The cache is not "open
	// season" after the restart — it just doesn't depend on the
	// relay's clock advancing to clear state.
	if cache.MarkSeen(relayKey, postRestartTxid1, postRestartSendTime1) {
		t.Fatal("replay of post-restart envelope #1 accepted")
	}
	if cache.MarkSeen(relayKey, postRestartTxid2, postRestartSendTime2) {
		t.Fatal("replay of post-restart envelope #2 accepted")
	}
}
