package server

import (
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// TestKNKReplayCacheRejectsDuplicate ensures a byte-identical replay of an
// accepted knock is refused.
func TestKNKReplayCacheRejectsDuplicate(t *testing.T) {
	cache := newKNKReplayCache()
	key := artTestPubkey(1)
	if !cache.MarkSeen(key, 7, 1000) {
		t.Fatal("first knock rejected")
	}
	if cache.MarkSeen(key, 7, 1000) {
		t.Fatal("replayed knock accepted")
	}
	if cache.Len() != 1 {
		t.Fatalf("cache length = %d, want 1", cache.Len())
	}
}

// TestKNKReplayCacheAcceptsNewTransactionID ensures a legitimate retry —
// which the agent issues with a fresh transaction id — is not mistaken for
// a replay.
func TestKNKReplayCacheAcceptsNewTransactionID(t *testing.T) {
	cache := newKNKReplayCache()
	key := artTestPubkey(1)
	if !cache.MarkSeen(key, 7, 1000) {
		t.Fatal("first knock rejected")
	}
	if !cache.MarkSeen(key, 8, 1000) {
		t.Fatal("knock with a new transaction id rejected")
	}
	if cache.Len() != 2 {
		t.Fatalf("cache length = %d, want 2", cache.Len())
	}
}

// TestKNKReplayCacheAcceptsNewSendTime ensures the send time is part of
// the key: an agent whose transaction counter restarted (agent process
// restart) but whose timestamp differs must not be rejected.
func TestKNKReplayCacheAcceptsNewSendTime(t *testing.T) {
	cache := newKNKReplayCache()
	key := artTestPubkey(1)
	if !cache.MarkSeen(key, 7, 1000) {
		t.Fatal("first knock rejected")
	}
	if !cache.MarkSeen(key, 7, 2000) {
		t.Fatal("knock with the same transaction id but a new send time rejected")
	}
}

// TestKNKReplayCacheSeparatesPeers ensures two agents never share a slot.
func TestKNKReplayCacheSeparatesPeers(t *testing.T) {
	cache := newKNKReplayCache()
	keyA, keyB := artTestPubkey(1), artTestPubkey(2)
	if !cache.MarkSeen(keyA, 7, 1000) {
		t.Fatal("first agent's knock rejected")
	}
	if !cache.MarkSeen(keyB, 7, 1000) {
		t.Fatal("second agent's knock rejected")
	}
	if cache.MarkSeen(keyA, 7, 1000) {
		t.Fatal("replayed knock accepted")
	}
}

// TestKNKReplayCacheExpiresAfterTTL ensures an entry stops blocking once
// its TTL has passed, so a knock that arrives after the freshness window
// (and is therefore rejected by the staleness gate) does not pin the slot
// forever.
func TestKNKReplayCacheExpiresAfterTTL(t *testing.T) {
	cache := newKNKReplayCacheWithParams(10, 20*time.Millisecond)
	key := artTestPubkey(1)
	if !cache.MarkSeen(key, 7, 1000) {
		t.Fatal("first knock rejected")
	}
	if cache.MarkSeen(key, 7, 1000) {
		t.Fatal("replayed knock accepted")
	}
	time.Sleep(40 * time.Millisecond)
	if !cache.MarkSeen(key, 7, 1000) {
		t.Fatal("knock rejected after the TTL expired")
	}
}

// TestKNKReplayCacheRejectsInvalidPubKey ensures a caller that skips the
// upstream public-key length guard still cannot key every empty pubkey
// under one slot.
func TestKNKReplayCacheRejectsInvalidPubKey(t *testing.T) {
	cache := newKNKReplayCache()
	if cache.MarkSeen(nil, 1, 1) {
		t.Fatal("empty pubkey accepted")
	}
	if cache.MarkSeen([]byte{1, 2, 3}, 1, 1) {
		t.Fatal("short pubkey accepted")
	}
	if cache.Len() != 0 {
		t.Fatalf("cache length = %d, want 0", cache.Len())
	}
}

// TestKNKReplayCacheSupportsCurveAndGMSMKeys ensures both authenticated
// public-key lengths are accepted.
func TestKNKReplayCacheSupportsCurveAndGMSMKeys(t *testing.T) {
	for _, size := range []int{core.PublicKeySize, core.PublicKeySizeEx} {
		cache := newKNKReplayCache()
		key := artTestPubkey(1, size)
		if !cache.MarkSeen(key, 1, 1) {
			t.Fatalf("first %d-byte key rejected", size)
		}
		if cache.MarkSeen(key, 1, 1) {
			t.Fatalf("replayed %d-byte key accepted", size)
		}
	}
}

func TestKNKReplayCacheRetainsLiveEntriesAtCapacity(t *testing.T) {
	cache := newKNKReplayCacheWithParams(1, time.Minute)
	key := artTestPubkey(1)
	if !cache.MarkSeen(key, 1, 1) {
		t.Fatal("empty cache refused first knock")
	}
	if cache.MarkSeen(artTestPubkey(2), 2, 2) {
		t.Fatal("full cache accepted new knock")
	}
	if cache.MarkSeen(key, 1, 1) || cache.CapacityRefusals() != 1 {
		t.Fatal("full cache did not retain the first entry")
	}
}

// TestKNKReplayTTLFollowsWindow ensures the TTL is derived from the
// effective freshness window, so shortening the window also shortens how
// long an entry is retained.
func TestKNKReplayTTLFollowsWindow(t *testing.T) {
	short := knkReplayTTL(60)
	long := knkReplayTTL(600)
	if short >= long {
		t.Fatalf("TTL for a 60 s window (%s) must be shorter than for 600 s (%s)", short, long)
	}
	if want := 420 * time.Second; short != want {
		t.Fatalf("TTL for a 60 s window = %s, want %s", short, want)
	}
	// A zero (unset) window falls back to the protocol default.
	if knkReplayTTL(0) != short {
		t.Fatalf("TTL for an unset window = %s, want %s", knkReplayTTL(0), short)
	}
}

func TestKNKReplayCachePeerQuotaAndCleanup(t *testing.T) {
	cache := newKNKReplayCacheWithParams(8, 20*time.Millisecond)
	peer := artTestPubkey(1)
	if !cache.MarkSeen(peer, 1, time.Now().Add(time.Minute).UnixNano()) {
		t.Fatal("first entry rejected")
	}
	if cache.MarkSeen(peer, 2, 0) || !cache.MarkSeen(artTestPubkey(2), 2, 0) {
		t.Fatal("a full peer quota must leave room for another peer")
	}
	time.Sleep(40 * time.Millisecond)
	if !cache.MarkSeen(peer, 3, 0) || cache.Len() != 1 {
		t.Fatal("expired entries must release their peer quotas")
	}
}
