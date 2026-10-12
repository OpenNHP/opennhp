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

// TestRelayReplayCacheEvictsFromSamePeer pins the per-peer eviction
// preference. When the cache is full and a new envelope from peer X
// arrives, the cache should evict peer X's own oldest entry first —
// not some other peer's entry — so a flooding peer's growth cannot
// push a quieter peer out of the cache. This is the property that
// stops the relay cache from becoming a per-peer DoS amplifier once
// the fail-closed path is removed (see MarkSeen's docstring for the
// full threat model).
//
// "Same peer" is defined by the per-call pubkey: the cache keeps a
// peerCount map[artPubkeyKey]int so a flooding client that only changes
// the (txid, sendTime) tuple (i.e. flowing under the same authenticated
// relay key) is treated as the same peer as the captured envelopes
// they're trying to evict.
func TestRelayReplayCacheEvictsFromSamePeer(t *testing.T) {
	now := time.Unix(100, 0)
	cache := newARTReplayCacheWithParams(3, time.Hour, "NHP_RLY", func() time.Time { return now })
	keyA := artTestPubkey(1)
	keyB := artTestPubkey(2)
	keyC := artTestPubkey(3)

	// Fill the cache with three of A's envelopes.
	if !cache.MarkSeen(keyA, 1, 1) {
		t.Fatal("A's first envelope rejected")
	}
	if !cache.MarkSeen(keyA, 2, 2) {
		t.Fatal("A's second envelope rejected")
	}
	if !cache.MarkSeen(keyA, 3, 3) {
		t.Fatal("A's third envelope rejected")
	}
	if got := len(cache.entries); got != 3 {
		t.Fatalf("cache size after fill = %d, want 3", got)
	}
	if got := cache.peerCount[artPubkeyKeyFromBytes(keyA)]; got != 3 {
		t.Fatalf("peerCount[A] after fill = %d, want 3", got)
	}

	// B's first envelope: A has no peer B entries to evict, so the
	// cache falls back to evicting A's global oldest. A's (txid=1,
	// sendTime=1) entry goes, B's entry goes in. A now has 2 live
	// entries, B has 1.
	if !cache.MarkSeen(keyB, 1, 1) {
		t.Fatal("B's first envelope rejected")
	}
	if got := len(cache.entries); got != 3 {
		t.Fatalf("cache size after B's first envelope = %d, want 3 (eviction expected)", got)
	}
	if got := cache.peerCount[artPubkeyKeyFromBytes(keyA)]; got != 2 {
		t.Fatalf("peerCount[A] = %d after B's first envelope, want 2", got)
	}
	if got := cache.peerCount[artPubkeyKeyFromBytes(keyB)]; got != 1 {
		t.Fatalf("peerCount[B] = %d after B's first envelope, want 1", got)
	}
	if _, stillPresent := cache.entries[artReplayKeyFor(keyA, 1, 1)]; stillPresent {
		t.Fatal("A's oldest entry (txid=1) should have been evicted to make room for B")
	}

	// B's second envelope: B has 1 live entry to evict, so the cache
	// evicts B's own (txid=1) entry, not one of A's. A's count must
	// stay at 2; B's count stays at 1.
	if !cache.MarkSeen(keyB, 2, 2) {
		t.Fatal("B's second envelope rejected")
	}
	if got := cache.peerCount[artPubkeyKeyFromBytes(keyA)]; got != 2 {
		t.Fatalf("peerCount[A] = %d after B's second envelope, want 2 (B's growth must NOT eat A's entries)", got)
	}
	if got := cache.peerCount[artPubkeyKeyFromBytes(keyB)]; got != 1 {
		t.Fatalf("peerCount[B] = %d after B's second envelope, want 1 (B's own eviction expected)", got)
	}
	if _, stillPresent := cache.entries[artReplayKeyFor(keyB, 1, 1)]; stillPresent {
		// This is the property the test pins: B's own oldest must
		// have been evicted (per-peer eviction prefers the admitting
		// peer's own entries). If an A entry was evicted instead,
		// B's first envelope would still be present and A's count
		// would have dropped.
		t.Fatal("B's own oldest (txid=1) should have been evicted, not one of A's")
	}

	// Now drive a flood from C — a brand-new peer with no live
	// entries. C's first envelope admits via the global-oldest
	// fallback (which evicts one of A's entries — that's the cost of
	// being the third peer to fill an already-full cache). C's
	// subsequent envelopes must evict C's own entries only — NOT
	// touch A or B any more. This is the load-bearing property for
	// the reviewer's "one busy peer does not starve a second peer"
	// claim: across C's growth, A and B's counts stabilize after C's
	// first envelope paid the global-fallback cost. C's count is
	// bounded by maxEntries minus (A's count + B's count).
	for i := 1; i <= 5; i++ {
		if !cache.MarkSeen(keyC, uint64(i), int64(i)) {
			t.Fatalf("C's %d-th envelope rejected", i)
		}
	}
	if got := cache.peerCount[artPubkeyKeyFromBytes(keyA)]; got != 1 {
		t.Fatalf("peerCount[A] = %d after C's five envelopes, want 1 (A's count must stabilize after C's first envelope paid the global-fallback cost)", got)
	}
	if got := cache.peerCount[artPubkeyKeyFromBytes(keyB)]; got != 1 {
		t.Fatalf("peerCount[B] = %d after C's five envelopes, want 1 (B is untouched by C's per-peer-eviction growth)", got)
	}
	if got := cache.peerCount[artPubkeyKeyFromBytes(keyC)]; got != 1 {
		t.Fatalf("peerCount[C] = %d after C's five envelopes, want 1 (C's count is bounded by maxEntries minus A+B's counts)", got)
	}

	// Sanity: a duplicate of A's retained entry is still detected as a
	// duplicate — the cache scopes by (pubkey, txid, sendTime), so
	// A3's (txid=3, sendTime=3) is matched by an identical re-submit
	// regardless of the eviction that just happened.
	if cache.MarkSeen(keyA, 3, 3) {
		t.Fatal("A's retained (txid=3, sendTime=3) no longer detected as duplicate — replay hole re-opened for the surviving entries")
	}
}

// TestRelayReplayCacheAdmitsSecondPeerWhenFull pins the second
// invariant the review asked for: when the global cache is filled with
// one peer's entries and a different peer sends its first envelope, the
// second peer is admitted (one of the first peer's entries is evicted
// to make room). Pre-per-peer-eviction the second peer's first envelope
// would still be admitted via the global oldest eviction path — the
// per-peer preference just makes that eviction kinder (preferring the
// first peer's entries when both the growth and the existing entries
// share a peer).
func TestRelayReplayCacheAdmitsSecondPeerWhenFull(t *testing.T) {
	now := time.Unix(100, 0)
	cache := newARTReplayCacheWithParams(4, time.Hour, "NHP_RLY", func() time.Time { return now })
	keyA := artTestPubkey(1)
	keyB := artTestPubkey(2)

	for i := 1; i <= 4; i++ {
		if !cache.MarkSeen(keyA, uint64(i), int64(i)) {
			t.Fatalf("A's %d-th envelope rejected", i)
		}
	}
	if got := len(cache.entries); got != 4 {
		t.Fatalf("cache size after A fills = %d, want 4", got)
	}

	// B sends its first envelope — must be admitted.
	if !cache.MarkSeen(keyB, 1, 1) {
		t.Fatal("B's first envelope rejected — second peer must still get in when cache is full of A's")
	}
	if got := len(cache.entries); got != 4 {
		t.Fatalf("cache size after B's first envelope = %d, want 4", got)
	}
	if got := cache.peerCount[artPubkeyKeyFromBytes(keyB)]; got != 1 {
		t.Fatalf("peerCount[B] = %d, want 1", got)
	}
	// A's count dropped by 1 — the global-oldest eviction (B has no
	// entries to evict itself, so we fell back to global oldest which
	// was A's first entry).
	if got := cache.peerCount[artPubkeyKeyFromBytes(keyA)]; got != 3 {
		t.Fatalf("peerCount[A] = %d, want 3 (A's first entry went via global-oldest fallback)", got)
	}
}

// artPubkeyKeyFromBytes and artReplayKeyFor mirror the cache's
// internal key shapes — the per-peer test above reaches into the cache
// to assert on its peerCount / entries invariants, so it has to build the
// same keys the cache would. Keeping them next to the test rather than
// exported keeps the cache's key shape private.
func artPubkeyKeyFromBytes(pubkey []byte) artPubkeyKey {
	var k artPubkeyKey
	copy(k.pubkey[:], pubkey)
	k.pubkeyLen = uint8(len(pubkey))
	return k
}

func artReplayKeyFor(pubkey []byte, txid uint64, sendTime int64) artReplayKey {
	var k artReplayKey
	copy(k.peerPubkey[:], pubkey)
	k.peerPubkeyLen = uint8(len(pubkey))
	k.txid = txid
	k.sendTime = sendTime
	return k
}
