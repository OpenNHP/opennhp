package main

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// openTestStore creates a UserStore backed by a fresh temp SQLite file.
// Returns the store and a cleanup function.
func openTestStore(t *testing.T) (*UserStore, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.db")
	s, err := OpenUserStore(path)
	if err != nil {
		t.Fatalf("OpenUserStore: %v", err)
	}
	return s, func() { _ = s.Close() }
}

// TestExternalSubjectKey confirms the provider prefix is applied so a
// GitHub id and an OIDC sub with the same string differ.
func TestExternalSubjectKey(t *testing.T) {
	got := externalSubjectKey("github", "12345")
	want := "github|12345"
	if got != want {
		t.Fatalf("externalSubjectKey = %q, want %q", got, want)
	}
	if externalSubjectKey("oidc", "12345") == externalSubjectKey("github", "12345") {
		t.Fatal("github and oidc keys collide for the same subject string")
	}
}

// insertLegacySubject writes a row with a RAW (pre-migration) subject +
// auth_provider, bypassing CreateUser so we can exercise the backfill.
func insertLegacySubject(t *testing.T, s *UserStore, subject, provider string) {
	t.Helper()
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO users (username, email, oidc_subject, status, created_at, auth_provider)
		VALUES (?, ?, ?, 'pending', 0, ?)`,
		"u_"+subject, subject+"@example.com", subject, provider)
	if err != nil {
		t.Fatalf("insert legacy: %v", err)
	}
}

// TestOIDCSubjectNamespacingMigration verifies that a database created
// with raw (pre-migration) oidc_subject values is backfilled to the
// provider-namespaced form on the next OpenUserStore, and that the
// backfill is idempotent (a second re-open does not double-prefix).
func TestOIDCSubjectNamespacingMigration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.db")

	// First open: schema + migration run, no legacy rows yet.
	s, err := OpenUserStore(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	insertLegacySubject(t, s, "12345", "github")  // numeric GitHub id
	insertLegacySubject(t, s, "abc-sub", "oidc")  // OIDC sub
	insertLegacySubject(t, s, "noop", "password") // provider not namespaced → skipped
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Re-open: migration backfill should namespace github/oidc rows.
	s, err = OpenUserStore(path)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	defer s.Close()

	ctx := context.Background()
	var ghSub, oidcSub, pwdSub sql.NullString
	if err := s.db.QueryRowContext(ctx, `SELECT oidc_subject FROM users WHERE username = 'u_12345'`).Scan(&ghSub); err != nil {
		t.Fatalf("query gh: %v", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT oidc_subject FROM users WHERE username = 'u_abc-sub'`).Scan(&oidcSub); err != nil {
		t.Fatalf("query oidc: %v", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT oidc_subject FROM users WHERE username = 'u_noop'`).Scan(&pwdSub); err != nil {
		t.Fatalf("query pwd: %v", err)
	}
	if got := ghSub.String; got != "github|12345" {
		t.Errorf("github subject = %q, want %q", got, "github|12345")
	}
	if got := oidcSub.String; got != "oidc|abc-sub" {
		t.Errorf("oidc subject = %q, want %q", got, "oidc|abc-sub")
	}
	// password-provider rows are left raw (no provider to namespace with).
	if got := pwdSub.String; got != "noop" {
		t.Errorf("password subject = %q, want %q (left untouched)", got, "noop")
	}

	// Idempotency: a third open must not double-prefix.
	s2, err := OpenUserStore(path)
	if err != nil {
		t.Fatalf("third open: %v", err)
	}
	s2.Close()
	s3, err := OpenUserStore(path)
	if err != nil {
		t.Fatalf("fourth open: %v", err)
	}
	defer s3.Close()
	var ghSub2 sql.NullString
	if err := s3.db.QueryRowContext(ctx, `SELECT oidc_subject FROM users WHERE username = 'u_12345'`).Scan(&ghSub2); err != nil {
		t.Fatalf("query gh re-open: %v", err)
	}
	if got := ghSub2.String; got != "github|12345" {
		t.Errorf("after re-open github subject = %q, want %q (idempotency broken)", got, "github|12345")
	}
}

// TestGetUserByOIDCSubjectNamespaced verifies the lookup uses the
// namespaced key so an OIDC sub that equals a victim's GitHub numeric
// id cannot resolve the victim's row (cross-provider takeover — the
// core of review #5). The legacy UNIQUE on oidc_subject meant a raw
// "999" could only belong to one provider; the namespaced lookup makes
// a wrong-provider lookup miss.
func TestGetUserByOIDCSubjectNamespaced(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.db")

	// Legacy GitHub row with raw numeric subject "999".
	s, err := OpenUserStore(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	insertLegacySubject(t, s, "999", "github")
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	// Re-open to run the backfill migration (999 -> github|999).
	s, err = OpenUserStore(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	// GitHub lookup with "999" resolves the victim row.
	gh, err := s.GetUserByOIDCSubject(ctx, "github", "999")
	if err != nil {
		t.Fatalf("github lookup: %v", err)
	}
	if gh.Username != "u_999" {
		t.Errorf("github lookup returned %q, want u_999", gh.Username)
	}
	// OIDC lookup with the SAME "999" must NOT resolve the GitHub row —
	// this is the takeover the namespacing prevents.
	if _, err := s.GetUserByOIDCSubject(ctx, "oidc", "999"); err == nil {
		t.Fatal("OIDC lookup matched a GitHub-linked row; cross-provider takeover not prevented")
	} else if !errorIsUserNotFound(err) {
		t.Fatalf("oidc lookup returned unexpected error: %v", err)
	}
}

// errorIsUserNotFound reports whether err wraps ErrUserNotFound.
func errorIsUserNotFound(err error) bool {
	for e := err; e != nil; {
		if e == ErrUserNotFound {
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := e.(unwrapper)
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}

// TestUsernameCaseInsensitiveUnique verifies the unique index on
// LOWER(username) prevents case-variant duplicates: "alice" and "Alice"
// cannot both exist, so GetUserByUsername (which lowercases) never
// compares a victim's password against the wrong row's bcrypt hash
// (review #4).
func TestUsernameCaseInsensitiveUnique(t *testing.T) {
	s, cleanup := openTestStore(t)
	defer cleanup()
	ctx := context.Background()

	if err := s.CreateUser(ctx, &User{Username: "alice", Email: "a1@x.com", Status: UserStatusActive}); err != nil {
		t.Fatalf("create alice: %v", err)
	}
	err := s.CreateUser(ctx, &User{Username: "Alice", Email: "a2@x.com", Status: UserStatusActive})
	if err == nil {
		t.Fatal("Alice insert succeeded; case-variant duplicate not prevented")
	}
	if !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("Alice insert returned %v, want ErrUsernameTaken", err)
	}
}

// TestEmailDuplicateMapsToSentinel verifies an email UNIQUE violation
// surfaces as ErrEmailTaken (so handleRegister maps it to 409, not 500)
// — review #4.
func TestEmailDuplicateMapsToSentinel(t *testing.T) {
	s, cleanup := openTestStore(t)
	defer cleanup()
	ctx := context.Background()

	if err := s.CreateUser(ctx, &User{Username: "u1", Email: "dup@x.com", Status: UserStatusActive}); err != nil {
		t.Fatalf("create u1: %v", err)
	}
	// handleRegister lowercases email before insert, so the column UNIQUE
	// (which is case-sensitive) is effectively case-insensitive in
	// production; mirror that here by using the same case.
	err := s.CreateUser(ctx, &User{Username: "u2", Email: "dup@x.com", Status: UserStatusActive})
	if err == nil {
		t.Fatal("duplicate email insert succeeded")
	}
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("duplicate email returned %v, want ErrEmailTaken", err)
	}
}

// TestExternalUsernameDisambiguation verifies that external-IdP users
// whose email matches a squatter's password username do NOT collide on
// the UNIQUE username column — because external users use the
// provider-namespaced subject as their username, not the email. A
// squatter registering username "victim@example.com" must not block a
// later external user with that email (review #3).
func TestExternalUsernameDisambiguation(t *testing.T) {
	s, cleanup := openTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Squatter registers a password account using the victim's email as
	// their (free-form) username.
	squatter := &User{
		Username:     "victim@example.com",
		Email:        "squatter@evil.com",
		PasswordHash: "pwhash",
		Status:       UserStatusActive,
		AuthProvider: "password",
	}
	if err := s.CreateUser(ctx, squatter); err != nil {
		t.Fatalf("create squatter: %v", err)
	}

	// The real victim signs in with an external IdP; upsert would set
	// Username = externalSubjectKey("oidc", sub), which must NOT collide
	// with the squatter's "victim@example.com" username.
	victim := &User{
		Username:     externalSubjectKey("oidc", "victim-sub"),
		Email:        "victim@example.com",
		OIDCSubject:  externalSubjectKey("oidc", "victim-sub"),
		Status:       UserStatusPending,
		AuthProvider: "oidc",
	}
	if err := s.CreateUser(ctx, victim); err != nil {
		t.Fatalf("create victim external user: %v (should not collide with squatter username)", err)
	}
}

// TestExpiredRegTokenReaped verifies that looking up an expired
// reg_token deletes the row so the table does not accumulate dead tokens
// (review #2).
func TestExpiredRegTokenReaped(t *testing.T) {
	s, cleanup := openTestStore(t)
	defer cleanup()
	ctx := context.Background()

	if err := s.CreateUser(ctx, &User{Username: "u", Email: "u@x.com", Status: UserStatusActive}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	expiredTok := "expired-token-xyz"
	// Insert a token whose expiry is already in the past.
	past := time.Now().UTC().Add(-1 * time.Minute).Unix()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO reg_tokens (token, user_id, expires_at) VALUES (?, ?, ?)`, expiredTok, 1, past); err != nil {
		t.Fatalf("insert expired token: %v", err)
	}

	if _, _, err := s.LookupRegToken(ctx, expiredTok); !errors.Is(err, ErrRegTokenExpired) {
		t.Fatalf("LookupRegToken expired = %v, want ErrRegTokenExpired", err)
	}
	// Row should have been reaped.
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reg_tokens WHERE token = ?`, expiredTok).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Errorf("expired reg_token row was not reaped (count=%d)", n)
	}
}
