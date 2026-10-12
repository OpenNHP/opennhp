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
// it): when the cache is full, a fresh key must be REFUSED rather
// than evicting a live entry. The eviction-on-full behavior (the legacy
// ART path) is checked separately in TestARTReplayCacheExpiresAndEvicts
// above.
//
// The scenario mirrors the relay-RPS math: with relayReplayCacheSize =
// 500_000 entries and a sustained ~1,000 envelopes/s, a capacity-evicting
// cache only retains ~500 seconds of traffic, below the 600-second
// staleness window. A captured envelope can therefore pass the outer
// replay check again after its shadow left the cache and reach a fresh
// inner connection. Rejecting new keys at full capacity keeps the live
// entries at their original recency until TTL expiry, and (importantly
// for the test) means a replay of an entry that IS still in the cache
// is still rejected afterwards.
func TestARTReplayCacheRejectsAtCapacityPreservesLive(t *testing.T) {
	const cap = 4
	cache := newARTReplayCacheWithParams(cap, time.Hour, "NHP_RLY-test", false, time.Now)

	// One specific entry we want to verify survives a full-cache
	// pressure round — the first one we insert is the "retained
	// envelope" from the test brief.
	retainedKey := artTestPubkey(1)
	if !cache.MarkSeen(retainedKey, 100, 1000) {
		t.Fatal("retained envelope was rejected on first observation")
	}

	// Fill the rest of the cache with entries the test doesn't follow up
	// on; each must be accepted (cache has spare capacity).
	for i := 0; i < cap-1; i++ {
		key := artTestPubkey(byte(2 + i))
		if !cache.MarkSeen(key, uint64(200+i), int64(2000+i)) {
			t.Fatalf("fill entry %d rejected while cache was under capacity", i)
		}
	}

	cache.mu.Lock()
	if n := len(cache.entries); n != cap {
		cache.mu.Unlock()
		t.Fatalf("pre-fill cache has %d entries, want %d", n, cap)
	}
	cache.mu.Unlock()

	// A brand-new key at capacity must be REFUSED. Pre-fix this would
	// have evicted the oldest live entry (retainedKey here) and reported
	// a capacityEvictions++; both are now false and true respectively.
	if cache.MarkSeen(artTestPubkey(99), 999, 9999) {
		t.Fatal("fresh key accepted at full capacity — live entries would be evicted; this is the regression")
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
		sendTime:      1000,
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
	if cache.MarkSeen(retainedKey, 100, 1000) {
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
	cache := newARTReplayCacheWithParams(cap, time.Hour, "NHP_RLY-test", false, time.Now)
	k1 := artTestPubkey(1)
	k2 := artTestPubkey(2)

	if !cache.MarkSeen(k1, 1, 1) || !cache.MarkSeen(k2, 2, 2) {
		t.Fatal("fill under capacity failed")
	}

	// Duplicate of an entry already in the cache. Must be rejected
	// but must NOT count as a capacity rejection (the cache didn't
	// refuse to make room — the key was already there).
	if cache.MarkSeen(k1, 1, 1) {
		t.Fatal("duplicate accepted")
	}

	cache.mu.Lock()
	if cache.capacityRejections != 0 {
		cache.mu.Unlock()
		t.Fatalf("duplicate marked as capacity rejection: rejections=%d, want 0", cache.capacityRejections)
	}
	cache.mu.Unlock()

	// Now a fresh key at full capacity: this IS a capacity rejection.
	if cache.MarkSeen(artTestPubkey(3), 3, 3) {
		t.Fatal("fresh key accepted at capacity")
	}
	cache.mu.Lock()
	if cache.capacityRejections != 1 {
		cache.mu.Unlock()
		t.Fatalf("capacity rejection counter not bumped: rejections=%d, want 1", cache.capacityRejections)
	}
	cache.mu.Unlock()
}
