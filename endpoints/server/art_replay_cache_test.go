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
	cache := newARTReplayCacheWithParams(8, time.Minute, "test", false, func() time.Time { return now })
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
	cache := newARTReplayCacheWithParams(2, time.Second, "test", false, func() time.Time { return now })
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

// TestRelayReplayCacheFailsClosedWhenFull pins the policy chosen for
// the relay replay cache: when the cache is at capacity and a genuinely
// new envelope arrives, MarkSeen returns false (reject) instead of
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
//   - the third MarkSeen returning true (eviction accepted the new
//     envelope)
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
	cache := newARTReplayCacheWithParams(2, time.Hour, "NHP_RLY", true, func() time.Time { return now })
	keyA := artTestPubkey(1)
	keyB := artTestPubkey(2)
	keyC := artTestPubkey(3)

	// Fill the cache to capacity with live entries.
	if !cache.MarkSeen(keyA, 1, 1) {
		t.Fatal("first key rejected")
	}
	if !cache.MarkSeen(keyB, 1, 1) {
		t.Fatal("second key rejected")
	}
	if got := len(cache.entries); got != 2 {
		t.Fatalf("cache size after fill = %d, want 2", got)
	}

	// A genuinely new key at full capacity must be REJECTED, not
	// admitted by evicting the oldest live entry.
	if cache.MarkSeen(keyC, 1, 1) {
		t.Fatal("new key accepted under fail-closed-full policy")
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
	if cache.MarkSeen(keyA, 1, 1) {
		t.Fatal("retained keyA no longer detected as duplicate — replay hole re-opened")
	}
	if cache.MarkSeen(keyB, 1, 1) {
		t.Fatal("retained keyB no longer detected as duplicate — replay hole re-opened")
	}

	// After expiry (advance the clock past the TTL), the oldest
	// entries fall off the front of the order list, the cache
	// shrinks back below capacity, and a new envelope is admitted
	// again. This is the property the eviction-based path ALSO
	// had; both policies converge on "capacity pressure eases once
	// the entries expire" — they differ only on what happens
	// while pressure is sustained.
	now = now.Add(2 * time.Hour)
	if !cache.MarkSeen(keyC, 1, 1) {
		t.Fatal("post-expiry new key still rejected — eviction path didn't free slots after entries expired")
	}
}
