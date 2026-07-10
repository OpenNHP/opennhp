package server

import (
	"sync"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func artTestPubkey(fill byte) []byte {
	key := make([]byte, core.PublicKeySize)
	for i := range key {
		key[i] = fill
	}
	return key
}

func TestARTReplayCacheScopesAndRejectsDuplicates(t *testing.T) {
	now := time.Unix(100, 0)
	cache := newARTReplayCacheWithParams(8, time.Minute, func() time.Time { return now })
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
	cache := newARTReplayCacheWithParams(2, time.Second, func() time.Time { return now })
	key := artTestPubkey(1)

	cache.MarkSeen(key, 1, 1)
	cache.MarkSeen(key, 2, 2)
	cache.MarkSeen(key, 3, 3)
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
	if cache.MarkSeen(nil, 1, 1) || cache.MarkSeen(make([]byte, core.PublicKeySize-1), 1, 1) {
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
