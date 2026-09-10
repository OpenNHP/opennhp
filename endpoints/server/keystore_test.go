package server

import (
	"errors"
	"fmt"
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
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 0, 0); err != nil {
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
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 0, 60); err != nil {
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
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 0, -42); err != nil {
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
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 0, 1); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	// Wait past expiry.
	time.Sleep(2500 * time.Millisecond)

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
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 0, 1); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	time.Sleep(2500 * time.Millisecond)

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
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 0, 60); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	time.Sleep(1100 * time.Millisecond)
	if err := s.RegisterAgentKey("alice", "dev1", pkB, 0, 60); err != nil {
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
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 0, 60); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	_, firstExp, err := s.GetAgentKeyExpiry("alice", "dev1")
	if err != nil || firstExp == nil {
		t.Fatalf("first register: %v firstExp=%v", err, firstExp)
	}
	time.Sleep(1100 * time.Millisecond)

	// Re-register with the SAME key — should be a no-op.
	if regErr := s.RegisterAgentKey("alice", "dev1", pkA, 0, 60); regErr != nil {
		t.Fatalf("RegisterAgentKey (idempotent): %v", regErr)
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
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 0, 60); err != nil {
		t.Fatalf("RegisterAgentKey alice: %v", err)
	}
	err := s.RegisterAgentKey("bob", "dev1", pkA, 0, 60)
	if !errors.Is(err, common.ErrPublicKeyAlreadyRegistered) {
		t.Fatalf("expected ErrPublicKeyAlreadyRegistered, got %v", err)
	}
}

func TestSweepExpiredDeactivates_ExpiresExpiredOnly(t *testing.T) {
	s, pkA, pkB := newTestStore(t)
	// Two keys: one that will expire, one with no expiry (TTL=0).
	if err := s.RegisterAgentKey("alice", "dev1", pkA, 0, 1); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}
	if err := s.RegisterAgentKey("alice", "dev2", pkB, 0, 0); err != nil {
		t.Fatalf("RegisterAgentKey: %v", err)
	}

	time.Sleep(2500 * time.Millisecond)

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

// ── OTP tests ──────────────────────────────────────────────────────────────

func TestGenerateOTP_ValidateOTP_Success(t *testing.T) {
	s := func() *AgentKeyStore {
		s, _, _ := newTestStore(t)
		return s
	}()

	code, err := s.GenerateOTP(OTPParams{UserId: "alice", DeviceId: "dev1", TTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("expected 6-digit OTP, got %q", code)
	}

	if err := s.ValidateOTP("alice", "dev1", code, ""); err != nil {
		t.Fatalf("ValidateOTP correct code: %v", err)
	}
}

func TestValidateOTP_WrongCode(t *testing.T) {
	s, _, _ := newTestStore(t)

	code, err := s.GenerateOTP(OTPParams{UserId: "alice", DeviceId: "dev1", TTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}

	// Wrong code should return ErrOTPInvalid.
	err = s.ValidateOTP("alice", "dev1", "000000", "")
	if !errors.Is(err, common.ErrOTPInvalid) {
		t.Fatalf("expected ErrOTPInvalid, got %v", err)
	}

	// Correct code must still work after a single wrong guess.
	if err := s.ValidateOTP("alice", "dev1", code, ""); err != nil {
		t.Fatalf("ValidateOTP correct code after wrong guess: %v", err)
	}
}

func TestValidateOTP_AlreadyUsed(t *testing.T) {
	s, _, _ := newTestStore(t)

	code, err := s.GenerateOTP(OTPParams{UserId: "alice", DeviceId: "dev1", TTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}

	if valErr := s.ValidateOTP("alice", "dev1", code, ""); valErr != nil {
		t.Fatalf("first ValidateOTP: %v", valErr)
	}

	// Second use must return ErrOTPAlreadyUsed.
	err = s.ValidateOTP("alice", "dev1", code, "")
	if !errors.Is(err, common.ErrOTPAlreadyUsed) {
		t.Fatalf("expected ErrOTPAlreadyUsed, got %v", err)
	}
}

func TestValidateOTP_Expired(t *testing.T) {
	s, _, _ := newTestStore(t)

	code, err := s.GenerateOTP(OTPParams{UserId: "alice", DeviceId: "dev1", TTL: 1 * time.Second})
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}

	time.Sleep(2500 * time.Millisecond)

	err = s.ValidateOTP("alice", "dev1", code, "")
	if !errors.Is(err, common.ErrOTPExpired) {
		t.Fatalf("expected ErrOTPExpired, got %v", err)
	}
}

func TestValidateOTP_RateLimit(t *testing.T) {
	s, _, _ := newTestStore(t)

	code, err := s.GenerateOTP(OTPParams{UserId: "alice", DeviceId: "dev1", TTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}

	wrong := "000000"
	if wrong == code {
		wrong = "999999"
	}

	// Exhaust attempts.
	for i := 0; i < MaxOTPAttempts; i++ {
		err = s.ValidateOTP("alice", "dev1", wrong, "")
		if i < MaxOTPAttempts-1 {
			if !errors.Is(err, common.ErrOTPInvalid) {
				t.Fatalf("attempt %d: expected ErrOTPInvalid, got %v", i+1, err)
			}
		} else {
			if !errors.Is(err, common.ErrOTPRateLimited) {
				t.Fatalf("attempt %d: expected ErrOTPRateLimited, got %v", i+1, err)
			}
		}
	}

	// After rate-limit, even the correct code must fail (OTP invalidated).
	err = s.ValidateOTP("alice", "dev1", code, "")
	if !errors.Is(err, common.ErrOTPInvalid) {
		t.Fatalf("after rate-limit: expected ErrOTPInvalid, got %v", err)
	}
}

func TestGenerateOTP_InvalidatesPrevious(t *testing.T) {
	s, _, _ := newTestStore(t)

	opts := OTPParams{UserId: "alice", DeviceId: "dev1", TTL: 5 * time.Minute, CooldownSeconds: -1}
	code1, err := s.GenerateOTP(opts)
	if err != nil {
		t.Fatalf("GenerateOTP 1: %v", err)
	}

	// Generate a second OTP for the same user+device — invalidates code1.
	code2, err := s.GenerateOTP(opts)
	if err != nil {
		t.Fatalf("GenerateOTP 2: %v", err)
	}
	if code1 == code2 {
		t.Fatal("expected different OTP codes")
	}

	// First OTP should now be used=1 (invalidated by the second GenerateOTP).
	err = s.ValidateOTP("alice", "dev1", code1, "")
	if !errors.Is(err, common.ErrOTPAlreadyUsed) {
		t.Fatalf("expected ErrOTPAlreadyUsed for old OTP, got %v", err)
	}

	// Second OTP should still work.
	if err := s.ValidateOTP("alice", "dev1", code2, ""); err != nil {
		t.Fatalf("ValidateOTP code2: %v", err)
	}
}

func TestValidateOTP_PublicKeyMismatch(t *testing.T) {
	s, _, _ := newTestStore(t)

	const pubKeyA = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	const pubKeyB = "BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB="

	code, err := s.GenerateOTP(OTPParams{UserId: "alice", DeviceId: "dev1", PublicKey: pubKeyA, TTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}

	// Correct code but wrong public key must be rejected.
	err = s.ValidateOTP("alice", "dev1", code, pubKeyB)
	if !errors.Is(err, common.ErrOTPPublicKeyMismatch) {
		t.Fatalf("expected ErrOTPPublicKeyMismatch, got %v", err)
	}

	// Correct code with correct public key must succeed.
	if err := s.ValidateOTP("alice", "dev1", code, pubKeyA); err != nil {
		t.Fatalf("ValidateOTP with matching pubKey: %v", err)
	}
}

// TestGenerateOTP_PerUserRateLimit verifies that generating OTPs for the
// same userId across different deviceIds is capped at MaxOTPPerUserPerWindow
// within the cooldown window. This prevents an attacker from bypassing the
// per-device cooldown by rotating the (unauthenticated) deviceId field.
func TestGenerateOTP_PerUserRateLimit(t *testing.T) {
	s, _, _ := newTestStore(t)

	// Generate one OTP per deviceId — each is a different device, so the
	// per-device cooldown allows them all. The per-userId cap should
	// reject the (MaxOTPPerUserPerWindow+1)-th request.
	for i := 0; i < MaxOTPPerUserPerWindow; i++ {
		devId := fmt.Sprintf("dev%d", i)
		_, err := s.GenerateOTP(OTPParams{UserId: "alice", DeviceId: devId, TTL: 5 * time.Minute})
		if err != nil {
			t.Fatalf("device %s (iteration %d): expected success, got %v", devId, i+1, err)
		}
	}

	// The next request with yet another deviceId must be rejected by the
	// per-userId rate limit.
	_, err := s.GenerateOTP(OTPParams{UserId: "alice", DeviceId: "dev_overflow", TTL: 5 * time.Minute})
	if !errors.Is(err, common.ErrOTPCooldown) {
		t.Fatalf("expected ErrOTPCooldown after %d devices, got %v", MaxOTPPerUserPerWindow, err)
	}

	// A different user must still be able to request OTP (isolation check).
	_, err = s.GenerateOTP(OTPParams{UserId: "bob", DeviceId: "dev0", TTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("different user should not be rate-limited: %v", err)
	}
}

// TestGenerateOTP_CooldownDisabledBypassesPerUserLimit verifies that a
// negative CooldownSeconds disables the per-userId and distinct-device
// checks as well (not just the per-device cooldown). This is the contract
// that existing tests rely on.
func TestGenerateOTP_CooldownDisabledBypassesPerUserLimit(t *testing.T) {
	s, _, _ := newTestStore(t)

	opts := OTPParams{UserId: "alice", DeviceId: "dev0", TTL: 5 * time.Minute, CooldownSeconds: -1}
	for i := 0; i < MaxOTPPerUserPerWindow+2; i++ {
		opts.DeviceId = fmt.Sprintf("dev%d", i)
		_, err := s.GenerateOTP(opts)
		if err != nil {
			t.Fatalf("cooldown disabled: iteration %d (device %s): expected success, got %v",
				i+1, opts.DeviceId, err)
		}
	}
}

// ── OTP sweep tests ──────────────────────────────────────────────────────

func TestSweepStaleOTPs_DeletesUsedAndExpired(t *testing.T) {
	s, _, _ := newTestStore(t)

	// Generate two OTPs and use one.
	code1, err := s.GenerateOTP(OTPParams{UserId: "alice", DeviceId: "dev1", TTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("GenerateOTP 1: %v", err)
	}
	_, err = s.GenerateOTP(OTPParams{UserId: "bob", DeviceId: "dev1", TTL: 1 * time.Second})
	if err != nil {
		t.Fatalf("GenerateOTP 2: %v", err)
	}

	// Use code1 (marks used=1).
	if valErr := s.ValidateOTP("alice", "dev1", code1, ""); valErr != nil {
		t.Fatalf("ValidateOTP code1: %v", valErr)
	}

	// Wait for code2 to expire.
	time.Sleep(2500 * time.Millisecond)

	// Sweep with 0 retention (delete everything that is already used or expired).
	n, err := s.SweepStaleOTPs(0)
	if err != nil {
		t.Fatalf("SweepStaleOTPs: %v", err)
	}
	if n != 2 {
		t.Fatalf("expected 2 stale OTPs deleted (1 used + 1 expired), got %d", n)
	}

	// Second sweep should be a no-op.
	n, err = s.SweepStaleOTPs(0)
	if err != nil {
		t.Fatalf("SweepStaleOTPs (second): %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 rows on second sweep, got %d", n)
	}
}

func TestSweepStaleOTPs_PreservesPending(t *testing.T) {
	s, _, _ := newTestStore(t)

	code, err := s.GenerateOTP(OTPParams{UserId: "alice", DeviceId: "dev1", TTL: 5 * time.Minute})
	if err != nil {
		t.Fatalf("GenerateOTP: %v", err)
	}

	// Sweep with 0 retention must NOT delete a pending (unused, unexpired) OTP.
	n, err := s.SweepStaleOTPs(0)
	if err != nil {
		t.Fatalf("SweepStaleOTPs: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 rows deleted (pending OTP preserved), got %d", n)
	}

	// Pending OTP must still validate.
	if err := s.ValidateOTP("alice", "dev1", code, ""); err != nil {
		t.Fatalf("ValidateOTP after sweep: %v", err)
	}
}
