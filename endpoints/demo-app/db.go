// Database access for the Demo App.
//
// We use modernc.org/sqlite (pure-go, same as nhp-server's keystore) to keep
// the build CGo-free. The schema is intentionally minimal: a single users
// table that holds both password credentials and OIDC subjects.
package demoapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// ErrUserNotFound is returned by Find* helpers when no row matches.
var ErrUserNotFound = errors.New("user not found")

// ErrUserAlreadyExists is returned by CreateUser on UNIQUE conflicts.
var ErrUserAlreadyExists = errors.New("user already exists")

// User is the runtime representation of a row in the users table.
// All optional fields use sql.Null* types so empty values are distinguishable
// from "" in the JSON response.
type User struct {
	ID                   int64
	Username             string
	PasswordHash         sql.NullString
	Email                string
	NhpPublicKey         sql.NullString
	EncryptedPrivateKey  sql.NullString
	NhpRegisteredAt      sql.NullInt64
	OIDCSub              sql.NullString
	OIDCProvider         sql.NullString
	CreatedAt            time.Time
}

// Open opens (or creates) the SQLite database at path and applies the schema.
// The parent directory is created if missing.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	// SQLite is single-writer; cap the connection pool to avoid spurious
	// "database is locked" errors under modest concurrent load.
	db.SetMaxOpenConns(1)
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return db, nil
}

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT,
    email TEXT NOT NULL,
    nhp_public_key TEXT,
    encrypted_private_key TEXT,
    nhp_registered_at INTEGER,
    oidc_sub TEXT,
    oidc_provider TEXT,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oidc ON users(oidc_provider, oidc_sub)
    WHERE oidc_sub IS NOT NULL;
`

// CreateUser inserts a new user row in tx. Returns the new user ID on success,
// or ErrUserAlreadyExists if either the username or the (provider, oidc_sub)
// pair is already taken.
func CreateUser(ctx context.Context, tx *sql.Tx, u *User) (int64, error) {
	res, err := tx.ExecContext(ctx, `
        INSERT INTO users (username, password_hash, email, nhp_public_key, encrypted_private_key, oidc_sub, oidc_provider)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `,
		u.Username,
		u.PasswordHash,
		u.Email,
		u.NhpPublicKey,
		u.EncryptedPrivateKey,
		u.OIDCSub,
		u.OIDCProvider,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrUserAlreadyExists
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

// FindByUsername returns the user with the given username, or ErrUserNotFound.
func FindByUsername(ctx context.Context, db *sql.DB, username string) (*User, error) {
	row := db.QueryRowContext(ctx, `
        SELECT id, username, password_hash, email, nhp_public_key, encrypted_private_key,
               nhp_registered_at, oidc_sub, oidc_provider, created_at
        FROM users WHERE username = ?`, username)
	return scanUser(row)
}

// FindByOIDCSub returns the user matching (provider, sub), or ErrUserNotFound.
func FindByOIDCSub(ctx context.Context, db *sql.DB, provider, sub string) (*User, error) {
	row := db.QueryRowContext(ctx, `
        SELECT id, username, password_hash, email, nhp_public_key, encrypted_private_key,
               nhp_registered_at, oidc_sub, oidc_provider, created_at
        FROM users WHERE oidc_provider = ? AND oidc_sub = ?`, provider, sub)
	return scanUser(row)
}

// FindByID returns the user with the given numeric ID, or ErrUserNotFound.
func FindByID(ctx context.Context, db *sql.DB, id int64) (*User, error) {
	row := db.QueryRowContext(ctx, `
        SELECT id, username, password_hash, email, nhp_public_key, encrypted_private_key,
               nhp_registered_at, oidc_sub, oidc_provider, created_at
        FROM users WHERE id = ?`, id)
	return scanUser(row)
}

// MarkNHPRegistered sets nhp_registered_at to "now" for the user with the
// given username, but only if the public key matches the one we recorded
// during /api/users/register. This prevents a stale call from the plugin
// (e.g., a prior registration) from spuriously marking the user complete.
func MarkNHPRegistered(ctx context.Context, db *sql.DB, username, publicKey string) error {
	res, err := db.ExecContext(ctx, `
        UPDATE users
        SET nhp_registered_at = strftime('%s','now')
        WHERE username = ? AND nhp_public_key = ?
    `, username, publicKey)
	if err != nil {
		return fmt.Errorf("mark registered: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("user %q with the given public key not found", username)
	}
	return nil
}

// SetOIDCKeys stores the NHP public key and the password-encrypted private
// key for an OIDC user that just completed the NHP key onboarding step.
func SetOIDCKeys(ctx context.Context, db *sql.DB, username, pubKey, encPriv string) error {
	_, err := db.ExecContext(ctx, `
        UPDATE users SET nhp_public_key = ?, encrypted_private_key = ?
        WHERE username = ?`, pubKey, encPriv, username)
	if err != nil {
		return fmt.Errorf("set oidc keys: %w", err)
	}
	return nil
}

// FindByEmail is used by the plugin's RequestOTP fallback when the email
// isn't supplied in UserData (e.g., for OIDC users re-issuing a registration
// key). Returns ErrUserNotFound when no user owns the email.
func FindByEmail(ctx context.Context, db *sql.DB, email string) (string, error) {
	var username string
	err := db.QueryRowContext(ctx, `SELECT username FROM users WHERE email = ?`, email).Scan(&username)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lookup by email: %w", err)
	}
	return username, nil
}

func scanUser(row *sql.Row) (*User, error) {
	u := &User{}
	var createdAt int64
	err := row.Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Email,
		&u.NhpPublicKey, &u.EncryptedPrivateKey, &u.NhpRegisteredAt,
		&u.OIDCSub, &u.OIDCProvider, &createdAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	u.CreatedAt = time.Unix(createdAt, 0).UTC()
	return u, nil
}

// isUniqueViolation looks for the SQLITE_CONSTRAINT error code that
// modernc.org/sqlite returns on UNIQUE / PRIMARY KEY violations.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "PRIMARY KEY constraint failed")
}
