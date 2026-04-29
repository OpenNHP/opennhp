package ac

import (
	"bytes"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// All cache-key tests pin sendTime explicitly so the (pubkey, txid,
// sendTime) triple is exercised directly. Tests that need to
// distinguish "same packet" from "different packet" vary sendTime;
// tests that fence the (pubkey, txid)-only behavior keep it
// constant. The constant is an arbitrary monotonic-nanos value so
// it is readable in failure messages.
const testSendTime int64 = 1_700_000_000_000_000_000

// pubkeyN returns a 32-byte (PublicKeySize) test pubkey filled with
// the given byte. Inline at call sites to avoid a helper file —
// MarkSeen runtime-rejects anything that is not exactly
// PublicKeySize, so every test fixture must hit that length.
func pubkeyN(b byte) []byte {
	return bytes.Repeat([]byte{b}, core.PublicKeySize)
}

// TestAOPReplayCache_FirstSeenAccepted ensures a fresh
// (pubkey, txid, sendTime) triple is recorded and reported as
// first-seen.
//
// Regression fence for issue #1123: the AC AOP path previously had
// no TransactionId dedupe, allowing replays of a captured AOP to
// re-open ipset entries on a fresh connection.
func TestAOPReplayCache_FirstSeenAccepted(t *testing.T) {
	c := newAOPReplayCache()
	pub := pubkeyN('A')

	if !c.MarkSeen(pub, 100, testSendTime) {
		t.Fatal("first observation must be reported as first-seen")
	}
}

// TestAOPReplayCache_DuplicateRejected verifies the second
// observation of the same (pubkey, txid, sendTime) triple is
// rejected as a replay. Both calls use identical sendTime —
// captured-and-replayed packets carry byte-identical
// AEAD-authenticated timestamps.
//
// Regression fence for issue #1123.
func TestAOPReplayCache_DuplicateRejected(t *testing.T) {
	c := newAOPReplayCache()
	pub := pubkeyN('A')

	if !c.MarkSeen(pub, 100, testSendTime) {
		t.Fatal("first observation must be reported as first-seen")
	}
	if c.MarkSeen(pub, 100, testSendTime) {
		t.Fatal("duplicate observation must be reported as already-seen")
	}
}

// TestAOPReplayCache_DifferentPubkeysSameTxid asserts the cache is
// pubkey-scoped: post-#676/#677 the AC connects to multiple servers
// with independent monotonic counters, so the same txid value from
// two distinct pubkeys must both be accepted.
func TestAOPReplayCache_DifferentPubkeysSameTxid(t *testing.T) {
	c := newAOPReplayCache()
	pubA := pubkeyN('A')
	pubB := pubkeyN('B')

	if !c.MarkSeen(pubA, 100, testSendTime) {
		t.Fatal("server-A txid=100 must be accepted")
	}
	if !c.MarkSeen(pubB, 100, testSendTime) {
		t.Fatal("server-B txid=100 must be accepted (different pubkey scope)")
	}
}

// TestAOPReplayCache_PostRestartCounterCollision asserts the
// sendTime component closes the false-reject window the bare
// (pubkey, txid) shape would have. Production servers share one
// keypair across the fleet (terraform/modules/compute/main.tf:
// aws_secretsmanager_secret "server" is one secret per cell), and
// SenderTrxId is `atomic.AddUint64` over an in-memory counter that
// resets on process restart. Without the sendTime in the key, the
// fleet's first ~K post-restart AOPs would silently drop for up to
// TTL (11 min). With the sendTime in the key, a fresh wall-clock
// timestamp distinguishes the legitimate post-restart packet from
// the cached pre-restart entry.
func TestAOPReplayCache_PostRestartCounterCollision(t *testing.T) {
	c := newAOPReplayCache()
	pub := pubkeyN('A')

	const txid uint64 = 1
	const preRestartTime = testSendTime
	// `int64(...)` is required: testSendTime is a typed int64 (a
	// nanos-since-epoch counter, not a Duration), and Go does not
	// implicitly convert time.Duration to int64 for arithmetic.
	const postRestartTime = testSendTime + int64(30*time.Second)

	if !c.MarkSeen(pub, txid, preRestartTime) {
		t.Fatal("pre-restart AOP must be accepted as first-seen")
	}
	if !c.MarkSeen(pub, txid, postRestartTime) {
		t.Fatal("post-restart AOP with same (pubkey, txid) but fresh sendTime must be accepted — counter-collision is not a replay")
	}

	// Sanity check: a true byte-for-byte replay of the pre-restart
	// packet (same triple) must still be rejected.
	if c.MarkSeen(pub, txid, preRestartTime) {
		t.Fatal("byte-identical replay of pre-restart packet must still be rejected")
	}
}

// TestAOPReplayCache_TTLExpiry verifies entries age out so the
// cache cannot grow unbounded under steady traffic. Uses a polling
// loop with a generous deadline rather than a fixed sleep — heavily-
// loaded CI runners can stall a goroutine well past the TTL, so a
// fixed 3× margin is a known flake source (see
// endpoints/server/precheck_threat_cache_test.go for the same
// pattern).
func TestAOPReplayCache_TTLExpiry(t *testing.T) {
	c := newAOPReplayCacheWithParams(aopReplayCacheSize, 50*time.Millisecond)
	pub := pubkeyN('A')

	if !c.MarkSeen(pub, 100, testSendTime) {
		t.Fatal("first observation must be accepted")
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if c.MarkSeen(pub, 100, testSendTime) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("entry never aged out within the polling deadline")
}

// TestAOPReplayCache_ConcurrentMarkSeen pins the atomicity of
// MarkSeen: 64 goroutines racing to mark the same key must produce
// exactly one "first-seen" result. A naive Contains-then-Add on the
// underlying LRU has a TOCTOU window where two callers both observe
// "not seen" and both pass; the test asserts that window is closed.
func TestAOPReplayCache_ConcurrentMarkSeen(t *testing.T) {
	c := newAOPReplayCache()
	pub := pubkeyN('A')

	const goroutines = 64
	const txid uint64 = 100

	var firstSeen atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if c.MarkSeen(pub, txid, testSendTime) {
				firstSeen.Add(1)
			}
		}()
	}

	close(start)
	wg.Wait()

	if got := firstSeen.Load(); got != 1 {
		t.Fatalf("exactly one MarkSeen must report first-seen; got %d", got)
	}
}

// TestAOPReplayCache_EmptyPubkeyRejected fails closed when the
// authenticated-pubkey upstream invariant is violated: a zero-length
// pubkey is rejected by the runtime guard so distinct AOPs cannot
// all key under the same `":<txid>:<ts>"` slot.
func TestAOPReplayCache_EmptyPubkeyRejected(t *testing.T) {
	c := newAOPReplayCache()

	if c.MarkSeen(nil, 100, testSendTime) {
		t.Fatal("nil pubkey must be rejected (fail-closed on upstream invariant violation)")
	}
	if c.MarkSeen([]byte{}, 100, testSendTime) {
		t.Fatal("empty pubkey must be rejected")
	}
	if got := c.Len(); got != 0 {
		t.Fatalf("rejected pubkeys must not enter the cache; Len=%d", got)
	}
}

// TestAOPReplayCache_WrongLengthPubkeyRejected pins the runtime
// enforcement of the fixed-PublicKeySize invariant. The
// key-construction docstring relies on every key bytes-block being
// the same length so the `:` separators delimit unambiguously; a
// future cipher scheme that produced 64-byte pubkeys without
// revisiting this file would silently violate that argument. The
// length guard fails closed instead.
func TestAOPReplayCache_WrongLengthPubkeyRejected(t *testing.T) {
	c := newAOPReplayCache()

	short := bytes.Repeat([]byte{'A'}, core.PublicKeySize-1)
	long := bytes.Repeat([]byte{'A'}, core.PublicKeySize+1)
	doublelong := bytes.Repeat([]byte{'A'}, core.PublicKeySize*2)

	for label, pub := range map[string][]byte{
		"short":      short,
		"long":       long,
		"doublelong": doublelong,
	} {
		if c.MarkSeen(pub, 100, testSendTime) {
			t.Errorf("pubkey of len=%d (%s) must be rejected (PublicKeySize=%d)", len(pub), label, core.PublicKeySize)
		}
	}
	if got := c.Len(); got != 0 {
		t.Fatalf("rejected pubkeys must not enter the cache; Len=%d", got)
	}
}

// TestAOPReplayCache_CapacityEviction asserts that the size cap is
// honored: pushing more than aopReplayCacheSize distinct entries
// must not let the cache grow beyond the cap, the most-recently
// inserted entry must still be a duplicate (fence the LRU eviction
// policy — a future change to FIFO/random would let the latest
// entry get evicted instead and would silently slip past a
// less-strict assertion), and the LRU-oldest entry must be
// re-acceptable as first-seen. Holds sendTime constant so eviction
// is keyed off txid drift only.
func TestAOPReplayCache_CapacityEviction(t *testing.T) {
	const cap = 8
	c := newAOPReplayCacheWithParams(cap, time.Hour)
	pub := pubkeyN('A')

	for i := 0; i < cap; i++ {
		if !c.MarkSeen(pub, uint64(i), testSendTime) {
			t.Fatalf("entry %d should be first-seen", i)
		}
	}
	if got := c.Len(); got != cap {
		t.Fatalf("Len=%d, want %d", got, cap)
	}

	// One more entry forces eviction of the LRU-oldest (txid=0).
	if !c.MarkSeen(pub, uint64(cap), testSendTime) {
		t.Fatalf("entry %d should be first-seen", cap)
	}
	if got := c.Len(); got != cap {
		t.Fatalf("after one more insert, Len=%d, want %d (cap holds)", got, cap)
	}
	// Pin the eviction policy: the most-recently-inserted entry
	// (txid=cap) MUST still be present as a duplicate. A change to
	// FIFO/random eviction would let txid=cap fall out instead of
	// txid=0; this assertion catches that policy drift.
	if c.MarkSeen(pub, uint64(cap), testSendTime) {
		t.Fatal("most-recently-inserted entry must still be a duplicate (LRU eviction policy fence)")
	}
	if !c.MarkSeen(pub, 0, testSendTime) {
		t.Fatal("LRU-evicted entry (txid=0) must be re-acceptable as first-seen")
	}
}

// TestPubkeyFingerprint_StableAndTruncated pins the breadcrumb
// helper used in the duplicate-drop log line. Empty input returns a
// grep-able sentinel; a real key returns base64 truncated to
// pubkeyFingerprintLen. URL-safe encoding is used so log indexers
// that treat `/` as a path separator (Loki, CloudWatch Insights)
// don't fragment the fingerprint.
func TestPubkeyFingerprint_StableAndTruncated(t *testing.T) {
	if got := pubkeyFingerprint(nil); got != "empty" {
		t.Errorf("pubkeyFingerprint(nil) = %q, want %q", got, "empty")
	}
	if got := pubkeyFingerprint([]byte{}); got != "empty" {
		t.Errorf("pubkeyFingerprint(empty) = %q, want %q", got, "empty")
	}

	// 32-byte pubkey → base64 RawURL is 43 chars → truncated to
	// pubkeyFingerprintLen.
	pub := make([]byte, core.PublicKeySize)
	for i := range pub {
		pub[i] = byte(i)
	}
	got := pubkeyFingerprint(pub)
	if len(got) != pubkeyFingerprintLen {
		t.Fatalf("pubkeyFingerprint(32-byte) length = %d, want %d", len(got), pubkeyFingerprintLen)
	}
	// URL-safe alphabet only — must not contain `+` or `/`.
	for _, c := range got {
		if c == '+' || c == '/' {
			t.Errorf("pubkeyFingerprint must use URL-safe base64; got %q (contains %q)", got, c)
		}
	}
	// Stable across calls.
	if again := pubkeyFingerprint(pub); again != got {
		t.Errorf("pubkeyFingerprint not stable: %q vs %q", got, again)
	}
}
