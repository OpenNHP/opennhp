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
	cache := newARTReplayCacheWithParams(8, time.Minute, "test", func() time.Time { return now })
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
	cache := newARTReplayCacheWithParams(2, time.Second, "test", func() time.Time { return now })
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

// TestRelayReplayCacheRejectsWhenFullAndKeepsEntries is the reviewer-named
// core gate: a RLY cache that is full must REFUSE new envelopes (return
// replayAdmitCacheFull, increment capacityRejections) WITHOUT evicting any
// live entry. A replay of a retained entry must still be classified as
// replayAdmitDuplicate, never CacheFull — the order check (step 2 before
// step 3 in Admit) is what makes that true and is what the cache-full
// branch must NOT be allowed to swallow. After TTL the cache auto-frees,
// proving the rejection is not a permanent stuck state.
func TestRelayReplayCacheRejectsWhenFullAndKeepsEntries(t *testing.T) {
	now := time.Unix(1000, 0)
	cache := newRelayReplayCacheWithParams(2, time.Minute, "test", func() time.Time { return now })
	key := artTestPubkey(0xAA, core.PublicKeySizeEx)

	if got := cache.Admit(key, 1, now.Unix()); got != replayAdmitNew {
		t.Fatalf("first envelope: got %v, want New", got)
	}
	now = now.Add(time.Second)
	if got := cache.Admit(key, 2, now.Unix()); got != replayAdmitNew {
		t.Fatalf("second envelope: got %v, want New", got)
	}

	// Third, distinct envelope — cache is full (cap=2), rejectWhenFull=true.
	now = now.Add(time.Second)
	if got := cache.Admit(key, 3, now.Unix()); got != replayAdmitCacheFull {
		t.Fatalf("third envelope: got %v, want CacheFull", got)
	}
	// MarkSeen wrapper reports false on every non-New outcome.
	if cache.MarkSeen(key, 3, now.Unix()) {
		t.Fatal("MarkSeen must be false on CacheFull")
	}
	// Capacity unchanged — no eviction happened.
	cache.mu.Lock()
	size := len(cache.entries)
	rej := cache.capacityRejections
	ev := cache.capacityEvictions
	cache.mu.Unlock()
	if size != 2 {
		t.Fatalf("entries size=%d, want 2 (no eviction under refuse-when-full)", size)
	}
	if rej != 1 {
		t.Fatalf("capacityRejections=%d, want 1", rej)
	}
	if ev != 0 {
		t.Fatalf("capacityEvictions=%d, want 0 (RLY cache must not evict)", ev)
	}

	// Core assertion: replay of a RETAINED entry under load is Duplicate,
	// not CacheFull. This is what guarantees that a flood of NEW envelopes
	// cannot bleed into the replay-rejection path or vice versa.
	if got := cache.Admit(key, 1, now.Unix()); got != replayAdmitDuplicate {
		t.Fatalf("replay of retained envelope: got %v, want Duplicate (must not be CacheFull)", got)
	}

	// Clock advances past the TTL — the cache auto-frees and accepts a
	// new envelope, proving the rejection is not a permanent stuck state.
	now = now.Add(2 * time.Minute)
	if got := cache.Admit(key, 4, now.Unix()); got != replayAdmitNew {
		t.Fatalf("post-TTL envelope: got %v, want New", got)
	}
}

// TestRelayReplayCacheFullStillRejectsDuplicateFirst pins the decision
// order at the cache-full boundary: a duplicate of a retained entry must
// be classified as Duplicate, and the capacity-rejection counter must
// NOT be touched (the new envelope is not what got rejected, the existing
// one is what got matched).
func TestRelayReplayCacheFullStillRejectsDuplicateFirst(t *testing.T) {
	now := time.Unix(2000, 0)
	cache := newRelayReplayCacheWithParams(2, time.Minute, "test", func() time.Time { return now })
	key := artTestPubkey(0xBB, core.PublicKeySizeEx)

	cache.Admit(key, 10, now.Unix())
	now = now.Add(time.Second)
	cache.Admit(key, 11, now.Unix())
	now = now.Add(time.Second)
	// New envelope rejected — counter bumps.
	if got := cache.Admit(key, 12, now.Unix()); got != replayAdmitCacheFull {
		t.Fatalf("third envelope: got %v, want CacheFull", got)
	}

	// Now send a duplicate of a retained entry — must be Duplicate, and
	// the rejection counter must NOT advance further.
	cache.mu.Lock()
	before := cache.capacityRejections
	cache.mu.Unlock()
	if got := cache.Admit(key, 10, now.Unix()); got != replayAdmitDuplicate {
		t.Fatalf("replay of retained: got %v, want Duplicate (Duplicate must precede capacity check)", got)
	}
	cache.mu.Lock()
	after := cache.capacityRejections
	cache.mu.Unlock()
	if after != before {
		t.Fatalf("capacityRejections advanced on duplicate (before=%d, after=%d) — Duplicate must not cost a capacity rejection", before, after)
	}
}

// TestRelayReplayCacheAdmitInvalidKeepsCacheUntouched is a defensive
// test for the "unreachable in production" path: an invalid input must
// neither insert nor bump any counter, and the cache must stay usable
// afterwards.
func TestRelayReplayCacheAdmitInvalidKeepsCacheUntouched(t *testing.T) {
	now := time.Unix(3000, 0)
	cache := newRelayReplayCacheWithParams(4, time.Minute, "test", func() time.Time { return now })

	// nil pubkey
	if got := cache.Admit(nil, 1, now.Unix()); got != replayAdmitInvalid {
		t.Fatalf("nil pubkey: got %v, want Invalid", got)
	}
	// wrong length
	if got := cache.Admit(make([]byte, 16), 1, now.Unix()); got != replayAdmitInvalid {
		t.Fatalf("short pubkey: got %v, want Invalid", got)
	}
	// nil receiver — the wrapper method tolerates a nil receiver.
	var nilCache *artReplayCache
	if got := nilCache.Admit(artTestPubkey(1), 1, 1); got != replayAdmitInvalid {
		t.Fatalf("nil receiver: got %v, want Invalid", got)
	}

	cache.mu.Lock()
	if cache.capacityRejections != 0 || cache.capacityEvictions != 0 || len(cache.entries) != 0 {
		t.Fatalf("invalid inputs must not mutate cache state: entries=%d evictions=%d rejections=%d",
			len(cache.entries), cache.capacityEvictions, cache.capacityRejections)
	}
	cache.mu.Unlock()
}
