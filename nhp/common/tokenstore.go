package common

import (
	"sync"
	"time"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// TokenEntry is an interface for token entries that have an expiration time.
// Both AccessEntry (AC) and ACTokenEntry (Server) implement this interface.
type TokenEntry interface {
	GetExpireTime() time.Time
}

// TokenStore is a generic two-level map for efficient token storage.
// The first level is indexed by the first character of the token for fast lookup.
// This design distributes tokens across ~64 buckets (base64 characters).
type TokenStore[E TokenEntry] struct {
	mu    sync.Mutex
	store map[string]map[string]E
}

// NewTokenStore creates a new TokenStore instance.
func NewTokenStore[E TokenEntry]() *TokenStore[E] {
	return &TokenStore[E]{
		store: make(map[string]map[string]E),
	}
}

// Store adds or updates a token entry in the store.
func (ts *TokenStore[E]) Store(token string, entry E) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	prefix := token[0:1]
	tokenMap, found := ts.store[prefix]
	if found {
		tokenMap[token] = entry
	} else {
		tokenMap = make(map[string]E)
		tokenMap[token] = entry
		ts.store[prefix] = tokenMap
	}
}

// Load retrieves a token entry from the store.
// Returns the entry and true if found, or zero value and false if not found.
func (ts *TokenStore[E]) Load(token string) (E, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	prefix := token[0:1]
	tokenMap, found := ts.store[prefix]
	if found {
		entry, ok := tokenMap[token]
		if ok {
			return entry, true
		}
	}
	var zero E
	return zero, false
}

// Delete removes a token from the store.
func (ts *TokenStore[E]) Delete(token string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	prefix := token[0:1]
	if tokenMap, found := ts.store[prefix]; found {
		delete(tokenMap, token)
		if len(tokenMap) == 0 {
			delete(ts.store, prefix)
		}
	}
}

// CleanExpired removes all expired tokens from the store.
// Returns the number of tokens removed.
func (ts *TokenStore[E]) CleanExpired() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	now := time.Now()
	removed := 0

	for prefix, tokenMap := range ts.store {
		for token, entry := range tokenMap {
			if now.After(entry.GetExpireTime()) {
				log.Info("[TokenStore] token %s expired, remove", token)
				delete(tokenMap, token)
				removed++
			}
		}
		if len(tokenMap) == 0 {
			delete(ts.store, prefix)
		}
	}

	return removed
}

// RunRefreshRoutine starts a background goroutine that periodically cleans
// expired tokens. It stops when the stop channel is closed.
// The wg.Done() is called when the routine exits.
func (ts *TokenStore[E]) RunRefreshRoutine(wg *sync.WaitGroup, stop <-chan struct{}, intervalSeconds int) {
	defer wg.Done()
	defer log.Info("tokenStoreRefreshRoutine stopped")

	log.Info("tokenStoreRefreshRoutine started")

	for {
		select {
		case <-stop:
			return
		case <-time.After(time.Duration(intervalSeconds) * time.Second):
			ts.CleanExpired()
		}
	}
}

// Size returns the total number of tokens in the store.
func (ts *TokenStore[E]) Size() int {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	count := 0
	for _, tokenMap := range ts.store {
		count += len(tokenMap)
	}
	return count
}
