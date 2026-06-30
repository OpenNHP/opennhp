package server

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// newTestStore opens a fresh keystore in t.TempDir() and registers a
// cleanup. Returns the store and a deterministic pair of test public keys.
func newTestStore(t *testing.T) (*AgentKeyStore, string, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := NewAgentKeyStore(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("NewAgentKeyStore: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })

	const pkA = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	const pkB = "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="
	return s, pkA, pkB
}

func TestRegisterAgentKey_TTLZeroStoresNull(t *testing.T) {
	s, pkA, _ := newTestStore(t)
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 0); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}

	active, exp, err := s.GetAgentKeyExpiry("alice", "dev1")
	if err != nil {
		t.Fatalf("GetAgentKeyExpiry: %v", err)
	}
	if !active {
		t.Fatalf("expected active=true for TTL=0 row")
	}
	if exp != nil {
		t.Fatalf("expected expiresAt=nil for TTL=0, got %v", *exp)
	}

	found, err := s.FindAgentByPublicKey(pkA)
	if err != nil || !found {
		t.Fatalf("FindAgentByPublicKey: found=%v err=%v", found, err)
	}
}

func TestRegisterAgentKey_TTLPositiveSetsFutureExpiry(t *testing.T) {
	s, pkA, _ := newTestStore(t)
	before := time.Now().Unix()
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 60); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	after := time.Now().Unix()

	active, exp, err := s.GetAgentKeyExpiry("alice", "dev1")
	if err != nil {
		t.Fatalf("GetAgentKeyExpiry: %v", err)
	}
	if !active || exp == nil {
		t.Fatalf("expected active=true with non-nil expiresAt")
	}
	// Allow ±2s wall-clock drift across the test boundary.
	if *exp < before+58 || *exp > after+62 {
		t.Fatalf("expiresAt out of expected range: got=%d want [%d,%d]", *exp, before+60, after+60)
	}
}

func TestRegisterAgentKey_TTLNegativeTreatedAsZero(t *testing.T) {
	s, pkA, _ := newTestStore(t)
	if err := s.RegisterAgentKey("alice", "dev1", pkA, -42); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	_, exp, err := s.GetAgentKeyExpiry("alice", "dev1")
	if err != nil {
		t.Fatalf("GetAgentKeyExpiry: %v", err)
	}
	if exp != nil {
		t.Fatalf("expected expiresAt=nil for negative TTL, got %v", *exp)
	}
}

// Load-bearing: an expired key MUST be invisible to FindAgentByPublicKey
// so the noise-layer peer validation fallback rejects the knock.
func TestFindAgentByPublicKey_ExpiredKeyHidden(t *testing.T) {
	s, pkA, _ := newTestStore(t)
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 1); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	// Wait past expiry.
	time.Sleep(1500 * time.Millisecond)

	found, err := s.FindAgentByPublicKey(pkA)
	if err != nil {
		t.Fatalf("FindAgentByPublicKey: %v", err)
	}
	if found {
		t.Fatalf("expired key should NOT be visible to FindAgentByPublicKey")
	}

	reg, err := s.IsAgentRegistered("alice", "dev1")
	if err != nil {
		t.Fatalf("IsAgentRegistered: %v", err)
	}
	if reg {
		t.Fatalf("expired key should NOT be visible to IsAgentRegistered")
	}

	rec, err := s.GetAgentKey("alice", "dev1")
	if err != nil {
		t.Fatalf("GetAgentKey: %v", err)
	}
	if rec != nil {
		t.Fatalf("expired key should NOT be returned by GetAgentKey")
	}
}

func TestGetAgentKeyExpiry_ExpiredReturnsFalse(t *testing.T) {
	s, pkA, _ := newTestStore(t)
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 1); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	time.Sleep(1500 * time.Millisecond)

	active, exp, err := s.GetAgentKeyExpiry("alice", "dev1")
	if err != nil {
		t.Fatalf("GetAgentKeyExpiry: %v", err)
	}
	if active || exp != nil {
		t.Fatalf("expected active=false exp=nil for expired row, got active=%v exp=%v", active, exp)
	}
}

// Key rotation (same user+device, different pubkey) MUST reset the clock.
func TestRegisterAgentKey_RotationResetsClock(t *testing.T) {
	s, pkA, pkB := newTestStore(t)
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 60); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	if err := s.RegisterAgentKey("alice", "dev1", pkB, 60); err != nil {
		t.Fatalf("RegisterAgentKey (rotation): %v", err)
	}

	before := time.Now().Unix()
	_, exp, err := s.GetAgentKeyExpiry("alice", "dev1")
	if err != nil {
		t.Fatalf("GetAgentKeyExpiry: %v", err)
	}
	if exp == nil {
		t.Fatalf("expected non-nil expiresAt after rotation")
	}
	// Rotation happened ~1.1s into the original 60s window. The new
	// expiry should be ~now+60, i.e. > before+58. Use a generous lower
	// bound to avoid flakes from slow CI.
	if *exp < before+58 {
		t.Fatalf("rotation did not reset clock: expiresAt=%d, expected >= %d", *exp, before+58)
	}
}

// Same key, same user — re-registering is idempotent and MUST NOT
// reset the expiry clock. Otherwise a network blip could indefinitely
// extend a key's validity.
func TestRegisterAgentKey_SameKeyNoOpDoesNotResetClock(t *testing.T) {
	s, pkA, _ := newTestStore(t)
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 60); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	_, firstExp, err := s.GetAgentKeyExpiry("alice", "dev1")
	if err != nil || firstExp == nil {
		t.Fatalf("first register: %v firstExp=%v", err, firstExp)
	}
	time.Sleep(1100 * time.Millisecond)

	// Re-register with the SAME key — should be a no-op.
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 60); err != nil {
		t.Fatalf("RegisterAgentKey (idempotent): %v", err)
	}
	_, secondExp, err := s.GetAgentKeyExpiry("alice", "dev1")
	if err != nil || secondExp == nil {
		t.Fatalf("second register: %v secondExp=%v", err, secondExp)
	}
	if *firstExp != *secondExp {
		t.Fatalf("idempotent re-register reset the clock: first=%d second=%d", *firstExp, *secondExp)
	}
}

func TestRegisterAgentKey_PublicKeyConflictAcrossUsers(t *testing.T) {
	s, pkA, _ := newTestStore(t)
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 60); err != nil {
		t.Fatalf("RegisterAgentKey alice: %v", err)
	}
	err := s.RegisterAgentKey("bob", "dev1", pkA, 60)
	if !errors.Is(err, common.ErrPublicKeyAlreadyRegistered) {
		t.Fatalf("expected ErrPublicKeyAlreadyRegistered, got %v", err)
	}
}

func TestSweepExpiredDeactivates_ExpiresExpiredOnly(t *testing.T) {
	s, pkA, pkB := newTestStore(t)
	// Two keys: one that will expire, one with no expiry (TTL=0).
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 1); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	if err := s.RegisterAgentKey("alice", "dev2", pkB, 0); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}

	time.Sleep(1500 * time.Millisecond)

	n, err := s.SweepExpiredDeactivates()
	if err != nil {
		t.Fatalf("SweepExpiredDeactivates: %v", err)
	}
	if n != 1 {
		t.Fatalf("expected 1 row swept, got %d", n)
	}

	// Sweep must not affect the NULL-expires row.
	active, exp, err := s.GetAgentKeyExpiry("alice", "dev2")
	if err != nil {
		t.Fatalf("GetAgentKeyExpiry dev2: %v", err)
	}
	if !active || exp != nil {
		t.Fatalf("NULL-expires row should remain active with no expiry: active=%v exp=%v", active, exp)
	}

	// After sweep, the expired row's active is 0, so FindAgentByPublicKey
	// is already false (the SQL filter on expires_at already hid it).
	found, err := s.FindAgentByPublicKey(pkA)
	if err != nil {
		t.Fatalf("FindAgentByPublicKey: %v", err)
	}
	if found {
		t.Fatalf("expired+swept key should not be visible")
	}
}

func TestSweepExpiredDeactivates_NoRowsNoError(t *testing.T) {
	s, _, _ := newTestStore(t)
	n, err := s.SweepExpiredDeactivates()
	if err != nil {
		t.Fatalf("SweepExpiredDeactivates: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 rows swept on empty store, got %d", n)
	}
}
