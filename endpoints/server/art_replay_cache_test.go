package server

import (
	"errors"
	"sync"
	"sync/atomic"
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

// callOk is the test-side spelling of MarkSeen returning nil. Kept as
// a tiny positive form so test bodies read as "cache.MarkSeen(...)
// returned ok" rather than "cache.MarkSeen(...) == nil", which mirrors
// how production code reads at the call site.
func callOk(err error) bool { return err == nil }

func TestARTReplayCacheSupportsCurveAndGMSMKeys(t *testing.T) {
	for _, size := range []int{core.PublicKeySize, core.PublicKeySizeEx} {
		cache := newARTReplayCache()
		key := artTestPubkey(1, size)
		if !callOk(cache.MarkSeen(key, 1, 1)) {
			t.Fatalf("first %d-byte key rejected", size)
		}
		if err := cache.MarkSeen(key, 1, 1); !errors.Is(err, ErrReplayCacheDuplicate) {
			t.Fatalf("duplicate %d-byte key accepted or wrong reason: %v", size, err)
		}
	}
}

func TestARTReplayCacheAcceptsOutOfOrderFreshPackets(t *testing.T) {
	cache := newARTReplayCache()
	key := artTestPubkey(1)
	if !callOk(cache.MarkSeen(key, 2, 200)) {
		t.Fatal("newer packet rejected")
	}
	if !callOk(cache.MarkSeen(key, 1, 100)) {
		t.Fatal("fresh packet with an earlier timestamp must survive burst reordering")
	}
	if err := cache.MarkSeen(key, 1, 100); !errors.Is(err, ErrReplayCacheDuplicate) {
		t.Fatalf("replay of the reordered packet accepted or wrong reason: %v", err)
	}
}

func TestARTReplayCacheScopesAndRejectsDuplicates(t *testing.T) {
	now := time.Unix(100, 0)
	cache := newARTReplayCacheWithParams(8, time.Minute, "test", false, false, func() time.Time { return now })
	keyA, keyB := artTestPubkey(1), artTestPubkey(2)

	if !callOk(cache.MarkSeen(keyA, 7, 1000)) {
		t.Fatal("first packet rejected")
	}
	if err := cache.MarkSeen(keyA, 7, 1000); !errors.Is(err, ErrReplayCacheDuplicate) {
		t.Fatalf("duplicate accepted or wrong reason: %v", err)
	}
	if !callOk(cache.MarkSeen(keyB, 7, 1000)) {
		t.Fatal("different peer key rejected")
	}
	if !callOk(cache.MarkSeen(keyA, 7, 1001)) {
		t.Fatal("fresh send time after counter reset rejected")
	}
	if !callOk(cache.MarkSeen(keyA, 8, 1000)) {
		t.Fatal("different transaction rejected")
	}
}

func TestARTReplayCacheExpiresAndEvicts(t *testing.T) {
	now := time.Unix(100, 0)
	cache := newARTReplayCacheWithParams(2, time.Second, "test", false, false, func() time.Time { return now })
	key := artTestPubkey(1)

	cache.MarkSeen(key, 1, 1)
	cache.MarkSeen(key, 2, 2)
	cache.MarkSeen(key, 3, 3)
	if cache.capacityEvictions != 1 || cache.lastWarning.IsZero() {
		t.Fatal("live eviction was not reported")
	}
	if !callOk(cache.MarkSeen(key, 1, 1)) {
		t.Fatal("capacity eviction did not release oldest key")
	}

	now = now.Add(2 * time.Second)
	if !callOk(cache.MarkSeen(key, 3, 3)) {
		t.Fatal("expired key remained blocked")
	}
}

func TestARTReplayCacheFailsClosedOnInvalidKey(t *testing.T) {
	cache := newARTReplayCache()
	if err := cache.MarkSeen(nil, 1, 1); !errors.Is(err, ErrReplayCacheInvalid) {
		t.Fatalf("nil key accepted or wrong reason: %v", err)
	}
	if err := cache.MarkSeen(make([]byte, core.PublicKeySize-1), 1, 1); !errors.Is(err, ErrReplayCacheInvalid) {
		t.Fatalf("short key accepted or wrong reason: %v", err)
	}
	if err := cache.MarkSeen(make([]byte, core.PublicKeySizeEx+1), 1, 1); !errors.Is(err, ErrReplayCacheInvalid) {
		t.Fatalf("oversize key accepted or wrong reason: %v", err)
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
			if callOk(cache.MarkSeen(key, 1, 1)) {
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

// TestRelayReplayCacheFailsClosedWhenFull pins the policy chosen for
// the relay replay cache: when the cache is at capacity and a genuinely
// new envelope arrives, MarkSeen returns ErrReplayCacheFull instead of
// evicting the oldest live entry.
//
// Why the relay cache has to fail closed: NHP_RLY carries every
// web-agent KNK/RKN/EXT behind every registered relay, and the
// per-connection monotonic / 20ms flood gates are exempted for it (see
// nhp/core.responder.shouldCheckRecvAttack / shouldCheckFlood and
// GHSA-674q-j8fh-972c). Without this hook an on-path attacker can
// replay any captured relay→server datagram indefinitely; the cache is
// the only line of defense. Evicting the oldest live entry under
// sustained pressure re-opens a replay hole for every envelope still
// inside its 600s acceptance window. Failing closed (rejecting the
// new envelope, leaving every retained entry untouched) is the safer
// direction: the new envelope is genuinely new so the rejection rate is
// bounded by the legitimate relay rate, and the captured-replay window
// stays intact.
//
// Pre-fix this test would have observed:
//   - the third MarkSeen returning nil (eviction accepted the new envelope)
//   - capacityEvictions == 1
//   - the originally-second MarkSeen key, when re-presented, accepted
//     (because eviction freed a slot)
//
// All three are wrong for the relay cache. ART keeps the eviction
// policy — its traffic is low-volume and the staleness window is
// forgiving — so we exercise the eviction path under failClosedFull ==
// false in TestARTReplayCacheExpiresAndEvicts above.
func TestRelayReplayCacheFailsClosedWhenFull(t *testing.T) {
	now := time.Unix(100, 0)
	cache := newARTReplayCacheWithParams(2, time.Hour, "NHP_RLY", true, false, func() time.Time { return now })
	keyA := artTestPubkey(1)
	keyB := artTestPubkey(2)
	keyC := artTestPubkey(3)

	// Fill the cache to capacity with live entries.
	if !callOk(cache.MarkSeen(keyA, 1, 1)) {
		t.Fatal("first key rejected")
	}
	if !callOk(cache.MarkSeen(keyB, 1, 1)) {
		t.Fatal("second key rejected")
	}
	if got := len(cache.entries); got != 2 {
		t.Fatalf("cache size after fill = %d, want 2", got)
	}

	// A genuinely new key at full capacity must be REJECTED, not
	// admitted by evicting the oldest live entry.
	if err := cache.MarkSeen(keyC, 1, 1); !errors.Is(err, ErrReplayCacheFull) {
		t.Fatalf("new key not rejected under fail-closed-full policy: %v", err)
	}
	if got := len(cache.entries); got != 2 {
		t.Fatalf("fail-closed path mutated cache size = %d, want 2 (no eviction)", got)
	}
	if cache.capacityEvictions != 0 {
		t.Fatalf("fail-closed path recorded %d evictions, want 0", cache.capacityEvictions)
	}
	if cache.capacityRejects != 1 {
		t.Fatalf("fail-closed path recorded %d rejects, want 1", cache.capacityRejects)
	}

	// The two retained keys must still be present and still be
	// detected as duplicates — the whole point of the policy is
	// that a captured envelope held by the cache stays rejected
	// even after a flood of new envelopes has pushed the cache
	// past capacity.
	if err := cache.MarkSeen(keyA, 1, 1); !errors.Is(err, ErrReplayCacheDuplicate) {
		t.Fatalf("retained keyA no longer detected as duplicate: %v", err)
	}
	if err := cache.MarkSeen(keyB, 1, 1); !errors.Is(err, ErrReplayCacheDuplicate) {
		t.Fatalf("retained keyB no longer detected as duplicate: %v", err)
	}

	// After expiry (advance the clock past the TTL), the oldest
	// entries fall off the front of the order list, the cache
	// shrinks back below capacity, and a new envelope is admitted
	// again. This is the property the eviction-based path ALSO
	// had; both policies converge on "capacity pressure eases once
	// the entries expire" — they differ only on what happens
	// while pressure is sustained.
	now = now.Add(2 * time.Hour)
	if !callOk(cache.MarkSeen(keyC, 1, 1)) {
		t.Fatal("post-expiry new key still rejected — eviction path didn't free slots after entries expired")
	}
}

// TestRelayReplayCachePerPeerQuota pins the per-peerPubkey sub-quota
// on the relay cache. Without it, one busy (or authenticated-but-
// misbehaving) relay could fill the entire cache and starve every
// other relay for the TTL. The cap is dynamic — maxEntries / N
// active peers — so a relay joining late still gets a fair share, and
// a relay whose traffic drops back below the cap picks up room as
// its peers' entries expire.
//
// Layout for this test (capacity 12, 3 active peers):
//   - peerA fills 5 under N_active=1 cap=12
//   - peerB comes online: N_active=2 cap=6, peerB.count=0<6 → admitted
//   - peerC comes online: N_active=3 cap=4, peerC.count=0<4 → admitted
//   - peerA tries a 6th envelope: perPeerCap=4 (still 3 active peers),
//     peerA.count=5>=4 → REJECTED at per-peer branch (NOT global cap)
//
// The cap is recomputed per-insert based on N_active at that moment.
// Since peerA already has 5 entries when peerC arrives, N_active stays
// at 3 (peerA, peerB, peerC all in the count), and perPeerCap drops
// from 12 → 6 → 4 as more peers join.
func TestRelayReplayCachePerPeerQuota(t *testing.T) {
	now := time.Unix(100, 0)
	cache := newARTReplayCacheWithParams(12, time.Hour, "NHP_RLY", true, true, func() time.Time { return now })

	peerA := artTestPubkey(0xA)
	peerB := artTestPubkey(0xB)
	peerC := artTestPubkey(0xC)

	// peerA fills 5 envelopes under N_active=1 cap=12.
	for i := 0; i < 5; i++ {
		if !callOk(cache.MarkSeen(peerA, uint64(i), int64(i))) {
			t.Fatalf("peerA envelope %d rejected under N_active=1 cap=12", i)
		}
	}

	// peerB's first envelope: perPeerCap=12/2=6, admitted.
	if !callOk(cache.MarkSeen(peerB, 1, 1)) {
		t.Fatal("peerB first envelope rejected under N_active=2 cap=6")
	}

	// peerC's first envelope: perPeerCap=12/3=4, admitted.
	if !callOk(cache.MarkSeen(peerC, 1, 1)) {
		t.Fatal("peerC first envelope rejected under N_active=3 cap=4")
	}

	// peerA tries a 6th envelope: perPeerCap=12/3=4 (peerA is still
	// in peerEntryCount, so N_active stays at 3), peerA.count=5>=4
	// → rejected at the per-peer branch — the global cache is at 7
	// entries with 5 headroom, so this IS a per-peer rejection, not
	// a global-cap rejection.
	if err := cache.MarkSeen(peerA, 100, 100); !errors.Is(err, ErrReplayCacheFull) {
		t.Fatalf("peerA over-quota envelope accepted: %v", err)
	}
	if cache.capacityRejects < 1 {
		t.Fatalf("per-peer rejection did not increment capacityRejects: %d", cache.capacityRejects)
	}

	// Sanity check: peerA's retained entries are still detected as
	// duplicates (the per-peer branch doesn't evict).
	if err := cache.MarkSeen(peerA, 1, 1); !errors.Is(err, ErrReplayCacheDuplicate) {
		t.Fatalf("retained peerA envelope no longer detected as duplicate: %v", err)
	}
}

// TestRelayReplayCacheConcurrentRaceRejection drives a fail-closed
// relay cache past capacity from many goroutines at once. Without the
// counter-snapshot under c.mu (the deferred warning previously read
// c.capacityEvictions / c.capacityRejects after c.mu.Unlock()), this
// test fails under `go test -race` because the deferred warning's
// read races the increment. With the snapshot, the read uses a
// local captured under the same lock as the increment and stays
// race-free.
//
// Pre-fix this test would still pass on its first run (the
// counter starts at 0 and the deferred log never fires because the
// lastWarning rate-limit suppresses re-emission), but `go test -race`
// would flag the unsynchronized uint64 read on any single call that did
// emit a log line. This test forces that path by exercising the
// first-emission window across all workers.
func TestRelayReplayCacheConcurrentRaceRejection(t *testing.T) {
	cache := newARTReplayCacheWithParams(8, time.Hour, "NHP_RLY", true, true, func() time.Time {
		return time.Unix(1_000_000, 0)
	})

	const workers = 32
	var wg sync.WaitGroup
	wg.Add(workers)
	var accepted, rejected atomic.Int64
	for i := 0; i < workers; i++ {
		i := i
		go func() {
			defer wg.Done()
			key := artTestPubkey(byte(i + 1))
			err := cache.MarkSeen(key, uint64(i), int64(i))
			if err == nil {
				accepted.Add(1)
			} else if errors.Is(err, ErrReplayCacheFull) {
				rejected.Add(1)
			}
		}()
	}
	wg.Wait()

	// Capacity is 8; perPeerCap with N_active growing is at most 8
	// (when only one peer is active). We expect at most 8 accepts
	// and at least workers-8 rejects. The exact split depends on
	// goroutine scheduling; the property that matters is no panic /
	// no race-flag under `go test -race`.
	if accepted.Load() > 8 {
		t.Fatalf("accepted %d, want ≤ 8", accepted.Load())
	}
	if accepted.Load()+rejected.Load() != workers {
		t.Fatalf("counts don't sum: accepted=%d rejected=%d workers=%d", accepted.Load(), rejected.Load(), workers)
	}
}

// TestRelayReplayCachePerPeerCountReleasedOnEviction verifies the
// per-peer quota's bookkeeping drains correctly when entries expire.
// If peerEntryCount didn't decrement on evict/expire, an active peer
// would hit ErrReplayCacheFull immediately on their NEXT envelope
// even after the TTL elapsed — the cache would think it still held
// those entries.
func TestRelayReplayCachePerPeerCountReleasedOnEviction(t *testing.T) {
	now := time.Unix(100, 0)
	cache := newARTReplayCacheWithParams(4, time.Second, "NHP_RLY", true, true, func() time.Time { return now })
	peerA := artTestPubkey(0xA)

	// Fill peerA's share.
	for i := 0; i < 3; i++ {
		if !callOk(cache.MarkSeen(peerA, uint64(i), int64(i))) {
			t.Fatalf("peerA envelope %d rejected under cap=4/N=1", i)
		}
	}

	// The 4th envelope pushes peerA to perPeerCap=4. Still accepted
	// (cap = maxEntries/1 = 4).
	if !callOk(cache.MarkSeen(peerA, 100, 100)) {
		t.Fatal("peerA 4th envelope rejected under cap=4")
	}

	// Force expiry.
	now = now.Add(2 * time.Second)

	// After expiry the per-peer count must be back to zero — if it
	// wasn't decremented, peerA would still appear "full" and the
	// next envelope would be rejected.
	if !callOk(cache.MarkSeen(peerA, 200, 200)) {
		t.Fatal("peerA post-expiry envelope rejected — per-peer count leaked")
	}
}
