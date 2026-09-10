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

// TestEmailNoLongerUnique verifies the schema no longer enforces UNIQUE
// on email (review #2 — relaxed so IdP sign-in can coexist with a
// password-squat row). The application-level pre-check in
// handleRegister still rejects duplicate password registrations on
// the email match, and the rate limiter (5 / 10min / IP) bounds the
// duplicate-insert race in practice.
func TestEmailNoLongerUnique(t *testing.T) {
	s, cleanup := openTestStore(t)
	defer cleanup()
	ctx := context.Background()

	if err := s.CreateUser(ctx, &User{Username: "u1", Email: "dup@x.com", Status: UserStatusActive}); err != nil {
		t.Fatalf("create u1: %v", err)
	}
	// A second row with the same email must now succeed at the schema
	// level. The application pre-check is what gates duplicate password
	// registrations, not the column constraint.
	err := s.CreateUser(ctx, &User{Username: "u2", Email: "dup@x.com", Status: UserStatusActive})
	if err != nil {
		t.Fatalf("duplicate email insert now fails: %v (column UNIQUE on email should have been removed)", err)
	}
	// Both rows are queryable.
	if _, err := s.GetUserByEmail(ctx, "dup@x.com"); err != nil {
		t.Fatalf("GetUserByEmail after duplicate insert: %v", err)
	}
}

// TestDropEmailUniqueMigration verifies dropEmailUnique is idempotent:
// re-opening a fresh DB is a no-op and preserves existing rows
// (review #2).
func TestDropEmailUniqueMigration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.db")

	// Fresh DB: dropEmailUnique should be a no-op.
	s, err := OpenUserStore(path)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	ctx := context.Background()
	if err := s.CreateUser(ctx, &User{Username: "u1", Email: "u1@x.com", Status: UserStatusActive, AuthProvider: "password", PasswordHash: "pwhash"}); err != nil {
		t.Fatalf("create u1: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Re-open: idempotent migration should not blow away u1.
	s, err = OpenUserStore(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s.Close()
	u, err := s.GetUserByUsername(ctx, "u1")
	if err != nil {
		t.Fatalf("u1 lookup after reopen: %v", err)
	}
	if u.Email != "u1@x.com" {
		t.Errorf("u1.Email = %q, want u1@x.com", u.Email)
	}
}

// TestDropEmailUniqueMigrationLegacyDB exercises the actual rebuild
// path: a database created with the legacy schema (email TEXT UNIQUE)
// must have the column UNIQUE removed on re-open, and a subsequent
// duplicate-email insert must succeed (review #6).
//
// The legacy schema is hand-rolled here rather than pulled from a
// fixture so the test reads cleanly without binary files. We replicate
// the original CREATE TABLE byte-for-byte (email UNIQUE) plus the
// reg_tokens table + custom index that OpenUserStore.migrate expects
// to find after the rebuild.
func TestDropEmailUniqueMigrationLegacyDB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo.db")

	// Hand-roll a DB with the legacy column UNIQUE on email. We do
	// NOT call OpenUserStore here because OpenUserStore's CREATE
	// TABLE no longer carries the UNIQUE (review #2). To exercise
	// the migration we have to manufacture the legacy shape directly.
	raw, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(ON)")
	if err != nil {
		t.Fatalf("open legacy raw: %v", err)
	}
	if _, err := raw.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT,
			email TEXT UNIQUE NOT NULL,
			oidc_subject TEXT UNIQUE,
			nhp_public_key TEXT,
			nhp_private_key_enc TEXT,
			nhp_device_id TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at INTEGER NOT NULL
		)`); err != nil {
		t.Fatalf("create legacy users: %v", err)
	}
	if _, err := raw.Exec(`
		CREATE TABLE reg_tokens (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`); err != nil {
		t.Fatalf("create legacy reg_tokens: %v", err)
	}
	if _, err := raw.Exec(`INSERT INTO users (username, email, status, created_at) VALUES ('u1', 'dup@x.com', 'pending', 0)`); err != nil {
		t.Fatalf("insert legacy u1: %v", err)
	}
	if _, err := raw.Exec(`INSERT INTO reg_tokens (token, user_id, expires_at) VALUES ('tok1', 1, 9999999999)`); err != nil {
		t.Fatalf("insert legacy reg_token: %v", err)
	}
	// Confirm the legacy DB really carries the email UNIQUE — this
	// is what the migration is supposed to detect and remove.
	if _, err := raw.Exec(`INSERT INTO users (username, email, status, created_at) VALUES ('u2', 'dup@x.com', 'pending', 0)`); err == nil {
		t.Fatal("legacy DB accepted duplicate email — UNIQUE on email is missing; migration test is invalid")
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("close legacy raw: %v", err)
	}

	// Re-open via OpenUserStore: dropEmailUnique must rebuild the
	// table, remove the email UNIQUE, and preserve u1 + the in-flight
	// reg_token (foreign_keys=ON; the rebuild must disable cascades
	// during DROP TABLE so the token survives).
	s, err := OpenUserStore(path)
	if err != nil {
		t.Fatalf("open via OpenUserStore: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	u, err := s.GetUserByUsername(ctx, "u1")
	if err != nil {
		t.Fatalf("u1 lookup after migration: %v", err)
	}
	if u.Email != "dup@x.com" {
		t.Errorf("u1.Email = %q, want dup@x.com", u.Email)
	}
	var tokCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM reg_tokens`).Scan(&tokCount); err != nil {
		t.Fatalf("count reg_tokens: %v", err)
	}
	if tokCount != 1 {
		t.Errorf("reg_tokens after migration = %d, want 1 (DROP TABLE on users must not cascade)", tokCount)
	}

	// Duplicate email insert must now succeed at the schema level —
	// this is the exact behavior the legacy-DB migration unlocks.
	if err := s.CreateUser(ctx, &User{
		Username:     "u2",
		Email:        "dup@x.com",
		Status:       UserStatusActive,
		AuthProvider: "password",
		PasswordHash: "pwhash",
	}); err != nil {
		t.Fatalf("duplicate email insert after migration failed: %v (UNIQUE on email should be gone)", err)
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

// TestExpiredRegTokenReapsPendingUser covers the second half of review
// #3's fix: when LookupRegToken finds an expired token it must also
// reap the corresponding status=pending user row, so an attacker who
// registers victim@example.com, walks away, and never re-looks up the
// token cannot leave a permanent reservation on the address.
func TestExpiredRegTokenReapsPendingUser(t *testing.T) {
	s, cleanup := openTestStore(t)
	defer cleanup()
	ctx := context.Background()

	if err := s.CreateUser(ctx, &User{
		Username: "squat", Email: "victim@x.com", Status: UserStatusPending,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	expiredTok := "expired-squat-token"
	past := time.Now().UTC().Add(-1 * time.Minute).Unix()
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO reg_tokens (token, user_id, expires_at) VALUES (?, ?, ?)`,
		expiredTok, 1, past); err != nil {
		t.Fatalf("insert expired token: %v", err)
	}

	if _, _, err := s.LookupRegToken(ctx, expiredTok); !errors.Is(err, ErrRegTokenExpired) {
		t.Fatalf("LookupRegToken expired = %v, want ErrRegTokenExpired", err)
	}
	// Both the token AND the corresponding pending user must be gone.
	if _, err := s.GetUserByID(ctx, 1); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("status=pending user row was not reaped with its expired token: get err=%v", err)
	}
}

// TestReapExpiredPendingUsers covers the bulk counterpart (the startup
// sweep): a password-created status=pending row whose reg_token has
// expired is reaped, but an external-IdP status=pending row with no
// reg_token (the normal state for upsertExternalUser) is NOT reaped
// — the IdP user is still mid-handshake and the next call into the
// complete-registration view is the only thing that should flip them
// active. Active rows are never reaped.
func TestReapExpiredPendingUsers(t *testing.T) {
	s, cleanup := openTestStore(t)
	defer cleanup()
	ctx := context.Background()

	now := time.Now().UTC().Unix()
	past := now - 60
	future := now + 3600

	// Password pending user with an EXPIRED reg_token — must be reaped.
	if err := s.CreateUser(ctx, &User{
		Username: "p_expired", Email: "p_expired@x.com", Status: UserStatusPending,
	}); err != nil {
		t.Fatalf("create password expired: %v", err)
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO reg_tokens (token, user_id, expires_at) VALUES (?, ?, ?)`,
		"tok-p-expired", 1, past); err != nil {
		t.Fatalf("insert reg_token: %v", err)
	}

	// Password pending user with a LIVE reg_token — must NOT be reaped.
	if err := s.CreateUser(ctx, &User{
		Username: "p_live", Email: "p_live@x.com", Status: UserStatusPending,
	}); err != nil {
		t.Fatalf("create password live: %v", err)
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO reg_tokens (token, user_id, expires_at) VALUES (?, ?, ?)`,
		"tok-p-live", 2, future); err != nil {
		t.Fatalf("insert reg_token: %v", err)
	}

	// IdP pending user with NO reg_token — must NOT be reaped (this is
	// the row that distinguishes the bulk-reap from a "delete all
	// status=pending rows" shortcut).
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO users (username, email, oidc_subject, status, created_at, auth_provider)
		VALUES (?, ?, ?, 'pending', ?, 'oidc')`,
		"idp_pending", "idp_pending@x.com", "oidc|sub-1", now); err != nil {
		t.Fatalf("insert idp pending: %v", err)
	}

	// Active user — must NOT be reaped under any circumstance.
	if err := s.CreateUser(ctx, &User{
		Username: "active", Email: "active@x.com", Status: UserStatusActive,
	}); err != nil {
		t.Fatalf("create active: %v", err)
	}

	n, err := s.ReapExpiredPendingUsers(ctx)
	if err != nil {
		t.Fatalf("reap: %v", err)
	}
	if n != 1 {
		t.Errorf("reaped %d rows, want 1 (only the password expired one)", n)
	}

	// Active and live-token rows survive; expired-token row is gone.
	if _, err := s.GetUserByID(ctx, 1); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("password expired user still present: err=%v", err)
	}
	if _, err := s.GetUserByID(ctx, 2); err != nil {
		t.Errorf("password live user unexpectedly reaped: err=%v", err)
	}
	if _, err := s.GetUserByID(ctx, 3); err != nil {
		t.Errorf("idp pending user unexpectedly reaped: err=%v", err)
	}
	if _, err := s.GetUserByID(ctx, 4); err != nil {
		t.Errorf("active user unexpectedly reaped: err=%v", err)
	}
}
