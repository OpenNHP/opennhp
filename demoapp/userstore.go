package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	// Pure-Go SQLite driver so we don't need CGO in the build image.
	_ "modernc.org/sqlite"
)

// UserStatus tracks the lifecycle of a Demo user through the two-phase
// registration protocol. The keys are generated and persisted BEFORE the
// NHP server side is touched, so "active" requires both the local row to be
// marked active AND the agent to have completed the NHP_REG/RAK handshake.
type UserStatus string

const (
	UserStatusPending UserStatus = "pending"
	UserStatusActive  UserStatus = "active"
)

// User is the in-memory representation of a row in the users table. It
// holds both application credentials (username, bcrypt hash) and the NHP
// material (public key, AES-wrapped private key, deviceId).
type User struct {
	ID               int64
	Username         string
	PasswordHash     string // empty for OIDC-only users
	Email            string // also serves as NHP userId
	OIDCSubject      string // unique when set
	NHPPublicKey     string // base64-encoded public key
	NHPPrivateKeyEnc string // base64(nonce||ct||tag) — AES-GCM-wrapped
	NHPDeviceID      string
	Status           UserStatus
	// ServerName + CipherScheme bind the user to the nhp-server and scheme
	// chosen at registration. The private key is scheme-agnostic, so a
	// scheme switch re-derives the public key (and re-runs NHP_REG) without
	// rotating the stored private key.
	ServerName   string
	CipherScheme string
	CreatedAt    time.Time
}

// RegToken associates an in-flight registration with a user, so a refresh /
// browser-clear can call /api/register/retry to recover the credentials and
// re-attempt the NHP_REG handshake.
type RegToken struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
}

// UserStore is the SQLite-backed persistence layer. It owns one *sql.DB
// and a path; methods are safe for concurrent use.
type UserStore struct {
	db   *sql.DB
	path string
}

// OpenUserStore opens (or creates) the SQLite database file and runs the
// idempotent migration. modernc.org/sqlite accepts a "file:" DSN; we use
// the file path directly for simplicity and add the connection options as
// query parameters.
func OpenUserStore(path string) (*UserStore, error) {
	if path == "" {
		return nil, errors.New("userstore path is empty")
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("mkdir for db: %w", err)
		}
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite is single-writer; one conn avoids locking.
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	s := &UserStore{db: db, path: path}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the underlying *sql.DB. Safe to call multiple times.
func (s *UserStore) Close() error {
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

// migrate creates the users + reg_tokens tables if they don't yet exist,
// and idempotently adds columns introduced after the initial schema. Both
// are kept minimal so an operator can inspect them with `sqlite3`.
func (s *UserStore) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
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
		)`,
		`CREATE TABLE IF NOT EXISTS reg_tokens (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at INTEGER NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	}
	for _, q := range stmts {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("migrate: %w (stmt=%s)", err, q)
		}
	}
	// Idempotent column additions for existing databases. SQLite raises
	// "duplicate column name" when the column already exists; treat that as
	// success. Each new column ships with a default so old rows survive.
	addColumns := []struct {
		col, def string
	}{
		{"server_name", "TEXT DEFAULT ''"},
		{"cipher_scheme", "TEXT DEFAULT ''"},
	}
	for _, c := range addColumns {
		stmt := fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s", c.col, c.def)
		if _, err := s.db.Exec(stmt); err != nil {
			if !strings.Contains(err.Error(), "duplicate column name") {
				return fmt.Errorf("migrate: add column %s: %w", c.col, err)
			}
		}
	}
	return nil
}

// CreateUser inserts a new user row. Status defaults to pending; the
// caller flips it to active via ActivateUser once the NHP_REG handshake
// completes. External-IdP (OIDC/GitHub) users are also created pending and
// follow the same handshake via the complete-registration view.
func (s *UserStore) CreateUser(ctx context.Context, u *User) error {
	if u.Status == "" {
		u.Status = UserStatusPending
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now().UTC()
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO users
			(username, password_hash, email, oidc_subject,
			 nhp_public_key, nhp_private_key_enc, nhp_device_id,
			 status, created_at, server_name, cipher_scheme)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.Username, u.PasswordHash, u.Email, nullIfEmpty(u.OIDCSubject),
		nullIfEmpty(u.NHPPublicKey), nullIfEmpty(u.NHPPrivateKeyEnc), nullIfEmpty(u.NHPDeviceID),
		string(u.Status), u.CreatedAt.Unix(), u.ServerName, u.CipherScheme,
	)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	u.ID = id
	return nil
}

// UpdateUserKeys overwrites the NHP material on a user row (used for OIDC
// users who didn't have keys at password-less creation time, and for retry
// paths where the operator regenerates the key pair). It also rebinds the
// server/scheme so the complete-registration view can switch scheme: the
// private key is scheme-agnostic so it is NOT rotated here, only the
// derived public key + the chosen server/scheme binding.
func (s *UserStore) UpdateUserKeys(ctx context.Context, userID int64, pub, privEnc, deviceID, serverName, cipherScheme string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE users
		   SET nhp_public_key = ?, nhp_private_key_enc = ?, nhp_device_id = ?,
		       server_name = ?, cipher_scheme = ?
		 WHERE id = ?`,
		pub, privEnc, deviceID, serverName, cipherScheme, userID)
	if err != nil {
		return fmt.Errorf("update keys: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %d not found", userID)
	}
	return nil
}

// DeleteUser removes a user row by primary key. The reg_tokens rows cascade
// automatically via the FOREIGN KEY ... ON DELETE CASCADE on the reg_tokens
// table, so in-flight registration tokens for the user are cleaned up too.
// Returns ErrUserNotFound when no row matched (the caller already proved
// identity via the session, so a miss here means a concurrent delete).
func (s *UserStore) DeleteUser(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %d not found: %w", id, ErrUserNotFound)
	}
	return nil
}

// ActivateUser marks a user active. It does NOT verify the NHP server
// side — the caller already received a successful RAK from the agent. We
// trust the caller because /api/register/confirm is gated on the session
// token that the agent received the RAK under.
func (s *UserStore) ActivateUser(ctx context.Context, userID int64) error {
	res, err := s.db.ExecContext(ctx, `UPDATE users SET status = ? WHERE id = ?`, string(UserStatusActive), userID)
	if err != nil {
		return fmt.Errorf("activate user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %d not found", userID)
	}
	return nil
}

// GetUserByUsername looks up a user by username (case-insensitive trim).
func (s *UserStore) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return s.queryOne(ctx, `SELECT * FROM users WHERE LOWER(username) = LOWER(?) LIMIT 1`, strings.TrimSpace(username))
}

// GetUserByEmail looks up by email — used by OIDC to detect a password user
// with a matching email so the OIDC subject is linked rather than duplicating.
func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.queryOne(ctx, `SELECT * FROM users WHERE LOWER(email) = LOWER(?) LIMIT 1`, strings.TrimSpace(email))
}

// GetUserByOIDCSubject looks up by the IdP `sub` claim.
func (s *UserStore) GetUserByOIDCSubject(ctx context.Context, subject string) (*User, error) {
	return s.queryOne(ctx, `SELECT * FROM users WHERE oidc_subject = ? LIMIT 1`, subject)
}

// UpdateUserOIDCSubject attaches an OIDC `sub` claim to an existing user
// row. Used by the merge logic when an OIDC user signs in with the same
// email as a password user.
func (s *UserStore) UpdateUserOIDCSubject(ctx context.Context, userID int64, subject string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE users SET oidc_subject = ? WHERE id = ?`, subject, userID)
	if err != nil {
		return fmt.Errorf("update oidc subject: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %d not found", userID)
	}
	return nil
}

// GetUserByID is the primary-key lookup. Used by /api/register/confirm and
// /api/register/retry after they resolve the reg token.
func (s *UserStore) GetUserByID(ctx context.Context, id int64) (*User, error) {
	return s.queryOne(ctx, `SELECT * FROM users WHERE id = ? LIMIT 1`, id)
}

// SaveRegToken records a pending registration token so a browser-clear
// can pick the user back up without re-doing the password flow.
func (s *UserStore) SaveRegToken(ctx context.Context, t *RegToken) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO reg_tokens (token, user_id, expires_at) VALUES (?, ?, ?)`,
		t.Token, t.UserID, t.ExpiresAt.Unix())
	if err != nil {
		return fmt.Errorf("save reg token: %w", err)
	}
	return nil
}

// LookupRegToken fetches the user associated with a registration token.
// Returns ErrRegTokenExpired if the token is past its expiry; callers should
// then delete the row so the next retry can issue a fresh one.
func (s *UserStore) LookupRegToken(ctx context.Context, token string) (*RegToken, *User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT t.token, t.user_id, t.expires_at, u.id
		  FROM reg_tokens t
		  JOIN users u ON u.id = t.user_id
		 WHERE t.token = ?`, token)
	var (
		rt     RegToken
		uid    int64
		tokenS string
	)
	var expiresUnix int64
	if err := row.Scan(&tokenS, &rt.UserID, &expiresUnix, &uid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrRegTokenNotFound
		}
		return nil, nil, fmt.Errorf("lookup reg token: %w", err)
	}
	rt.Token = tokenS
	rt.ExpiresAt = time.Unix(expiresUnix, 0).UTC()
	if time.Now().UTC().After(rt.ExpiresAt) {
		return nil, nil, ErrRegTokenExpired
	}
	user, err := s.GetUserByID(ctx, uid)
	if err != nil {
		return nil, nil, err
	}
	return &rt, user, nil
}

// DeleteRegToken clears a registration token, used after confirm or on expiry.
func (s *UserStore) DeleteRegToken(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM reg_tokens WHERE token = ?`, token)
	if err != nil {
		return fmt.Errorf("delete reg token: %w", err)
	}
	return nil
}

// Errors surfaced by UserStore lookups. Routes translate these to 400/404.
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrRegTokenNotFound = errors.New("registration token not found")
	ErrRegTokenExpired  = errors.New("registration token expired")
)

// queryOne is the shared scanner for single-row lookups.
func (s *UserStore) queryOne(ctx context.Context, q string, args ...any) (*User, error) {
	row := s.db.QueryRowContext(ctx, q, args...)
	u := &User{}
	var (
		pwdHash    sql.NullString
		oidcSub    sql.NullString
		pub        sql.NullString
		privEnc    sql.NullString
		devID      sql.NullString
		statusStr  string
		createdAt  int64
		serverName sql.NullString
		cipherSch  sql.NullString
	)
	if err := row.Scan(
		&u.ID, &u.Username, &pwdHash, &u.Email, &oidcSub,
		&pub, &privEnc, &devID, &statusStr, &createdAt,
		&serverName, &cipherSch,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("query user: %w", err)
	}
	u.PasswordHash = pwdHash.String
	u.OIDCSubject = oidcSub.String
	u.NHPPublicKey = pub.String
	u.NHPPrivateKeyEnc = privEnc.String
	u.NHPDeviceID = devID.String
	u.Status = UserStatus(statusStr)
	u.ServerName = serverName.String
	u.CipherScheme = cipherSch.String
	u.CreatedAt = time.Unix(createdAt, 0).UTC()
	return u, nil
}

// nullIfEmpty returns a typed nil for sql.NullXxx fields, so we don't store
// empty strings for columns we mean to leave NULL.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
