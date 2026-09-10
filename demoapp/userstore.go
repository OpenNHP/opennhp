package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

// usersColumns is the canonical column list for the users table, in the
// order both INSERT and SELECT bind to. Centralized here so a schema
// change forces every caller to update — `SELECT *` would silently
// mis-bind Scan targets if a future column is added out of order
// (password_hash landing in email, etc.). The order MUST match the
// migrate() CREATE TABLE statement plus the ALTER TABLE ADD COLUMN
// statements in the same order they run.
const usersColumns = "id, username, password_hash, email, oidc_subject, nhp_public_key, nhp_private_key_enc, nhp_device_id, status, created_at, server_name, cipher_scheme, auth_provider"

// usersColumnsMinusID is the same list without `id` (AUTOINCREMENT, so
// the INSERT omits the placeholder too). Kept in lock-step with
// usersColumns so any future column addition must update both — the
// alternative would be `usersColumns` minus its first token, which would
// silently miscount if `id` is renamed.
const usersColumnsMinusID = "username, password_hash, email, oidc_subject, nhp_public_key, nhp_private_key_enc, nhp_device_id, status, created_at, server_name, cipher_scheme, auth_provider"

// usersRebuildSchema is the post-migration users table definition used by
// dropEmailUnique to rebuild the table without the email column UNIQUE.
// Mirrors the CREATE TABLE in migrate() minus the email UNIQUE; new
// columns added by ALTER TABLE in migrate() are listed here too so the
// copy step captures them. Keep in lock-step with the CREATE TABLE in
// migrate() — if you add a column there, add it here as well.
const usersRebuildSchema = `
CREATE TABLE users_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    password_hash TEXT,
    email TEXT NOT NULL,
    oidc_subject TEXT UNIQUE,
    nhp_public_key TEXT,
    nhp_private_key_enc TEXT,
    nhp_device_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at INTEGER NOT NULL,
    server_name TEXT DEFAULT '',
    cipher_scheme TEXT DEFAULT '',
    auth_provider TEXT DEFAULT ''
)`

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
	// AuthProvider records how the account was created: "password" for
	// local username/password sign-up, "github" / "oidc" for an external
	// IdP. Shown on the post-login screen so the user knows their origin.
	AuthProvider string
	CreatedAt    time.Time
}

// RegToken associates an in-flight registration with a user. The reg_token
// is a one-shot bearer credential consumed by /api/register/confirm when
// the browser session does not yet exist (the resume path uses an
// authenticated session instead).
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
	// Best-effort startup reap of stale status=pending rows whose
	// reg_token has expired (or which were never paired with one). The
	// in-lookup path in LookupRegToken handles per-row cleanup when a
	// caller happens to ask, but a registration that nobody looks up
	// again would otherwise pin the email indefinitely (review #3).
	// Errors are intentionally swallowed: a failed reap is a soft
	// "rows will be cleaned on the next lookups" fallback.
	if _, err := s.ReapExpiredPendingUsers(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "[demoapp] startup reap of expired pending users: %v\n", err)
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
		// email is intentionally NOT UNIQUE: an external-IdP sign-in
		// must be able to create a separate row for a victim whose
		// address was pre-registered by a password-squatting attacker
		// (review #2). The pre-check in handleRegister still rejects
		// duplicate password registrations on the email match, and
		// the rate limiter (5 / 10min / IP) bounds the duplicate-
		// insert race. Existing databases that DID carry the column
		// UNIQUE are migrated by dropEmailUnique below.
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT,
			email TEXT NOT NULL,
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
		{"auth_provider", "TEXT DEFAULT ''"},
	}
	for _, c := range addColumns {
		stmt := fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s", c.col, c.def)
		if _, err := s.db.Exec(stmt); err != nil {
			if !strings.Contains(err.Error(), "duplicate column name") {
				return fmt.Errorf("migrate: add column %s: %w", c.col, err)
			}
		}
	}
	// Case-insensitive username uniqueness. The username column is a plain
	// TEXT UNIQUE, which SQLite collates case-sensitively on write, while
	// GetUserByUsername compares with LOWER() (case-insensitive read).
	// Without enforcement on the write side, "alice" and "Alice" can both
	// insert; afterwards LOWER() returns whichever row scans first, so one
	// account can never log in — its bcrypt hash is never the one compared.
	// A unique index on the LOWER(username) expression makes the constraint
	// match the lookup, closing the read/write mismatch. CREATE ... IF NOT
	// EXISTS keeps it idempotent across re-opens. If a legacy database
	// already holds case-variant duplicates, this CREATE fails loudly at
	// startup so the operator cleans them up rather than silently keeping
	// the race open.
	if _, err := s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_ci ON users(LOWER(username))`); err != nil {
		return fmt.Errorf("migrate: username ci index: %w", err)
	}
	// Backfill oidc_subject to the provider-namespaced form
	// ("github|<id>" / "oidc|<sub>"). Rows created before this migration
	// stored the raw subject; without a provider prefix a GitHub numeric
	// id could collide with a numeric OIDC sub in the same UNIQUE column
	// (cross-provider account takeover — see review #5). The guard
	// (`NOT LIKE 'github|%' AND NOT LIKE 'oidc|%'`) makes it idempotent
	// so re-runs and already-migrated rows are left untouched. auth_provider
	// records which IdP originally created the row, so we can prefix
	// unambiguously; rows with an unknown/empty auth_provider are skipped
	// rather than guessed.
	if _, err := s.db.Exec(`
		UPDATE users
		   SET oidc_subject = auth_provider || '|' || oidc_subject
		 WHERE oidc_subject IS NOT NULL
		   AND oidc_subject != ''
		   AND auth_provider IN ('github', 'oidc')
		   AND oidc_subject NOT LIKE 'github|%'
		   AND oidc_subject NOT LIKE 'oidc|%'`); err != nil {
		return fmt.Errorf("migrate: namespace oidc_subject: %w", err)
	}
	// Drop the column-level UNIQUE on email (added in fresh CREATE TABLE
	// only; legacy databases that pre-date review #2 still have it).
	// See dropEmailUnique for the rationale and the rebuild mechanics.
	if err := s.dropEmailUnique(); err != nil {
		return fmt.Errorf("migrate: drop email unique: %w", err)
	}
	return nil
}

// dropEmailUnique rebuilds the users table without the column-level
// UNIQUE on email. Idempotent: detects the auto-generated UNIQUE index
// via PRAGMA index_list / index_info and rebuilds only when present.
// Required so an external-IdP sign-in can create a separate IdP-only
// row for a user whose email was pre-registered (squat) by a password
// account — otherwise upsertExternalUser's "skip password-holder,
// create separate IdP-only row" branch (review #2) would collide on
// INSERT with the squat row's email.
//
// Detection has to come from the pragmas: sqlite_schema.sql is NULL
// for every implicitly-created index (the very property that
// distinguishes sqlite_autoindex_* rows from user-created ones in
// sqlite_master), so scanning sqlite_master always reports "no email
// UNIQUE index" on a fresh legacy DB and the migration never fires.
// PRAGMA index_list returns origin='u' for indexes that back a
// column UNIQUE, and PRAGMA index_info returns the columns each one
// covers — that is the supported way to enumerate them.
//
// The rebuild runs inside a single transaction; on failure the
// original table is left intact (the DROP happens after the copy
// succeeds). PRAGMA foreign_keys=OFF is set OUTSIDE the transaction
// because the pragma is a no-op inside a tx — with foreign_keys=ON,
// DROP TABLE users would fire the ON DELETE CASCADE on reg_tokens and
// silently wipe every in-flight registration token. The original
// value is restored after the rebuild commits.
func (s *UserStore) dropEmailUnique() error {
	has, err := s.hasEmailUniqueIndex()
	if err != nil {
		return fmt.Errorf("detect email unique: %w", err)
	}
	if !has {
		// No email UNIQUE index — fresh DB or already migrated.
		return nil
	}

	// Pin a single *sql.Conn from the pool for the entire rebuild. The
	// pragma + transaction MUST run on the same connection: PRAGMA
	// foreign_keys is connection-scoped (a no-op inside a tx) and
	// s.db is configured with SetMaxOpenConns(1), so the *db.Conn pin
	// is a defense-in-depth guarantee against the ErrBadConn reopen
	// path — a stale connection can be silently replaced with a fresh
	// one built from the DSN, which carries _pragma=foreign_keys(ON).
	// If that fresh conn lands on the Begin, DROP TABLE users fires
	// the ON DELETE CASCADE on reg_tokens and silently wipes every
	// in-flight registration token (review #4 of the demoapp review).
	ctx := context.Background()
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("pin conn: %w", err)
	}
	defer conn.Close()

	// Save the prior value of foreign_keys so the rebuild can disable
	// cascades for the DROP TABLE on users and restore the operator's
	// setting afterwards. Both the read and the restore run on the
	// pinned conn, so a pool rotation cannot leave foreign_keys=OFF
	// dangling after the migration completes.
	prevFK, err := currentForeignKeysOnConn(ctx, conn)
	if err != nil {
		return fmt.Errorf("read foreign_keys: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		return fmt.Errorf("disable foreign_keys: %w", err)
	}
	defer func() {
		if _, err := conn.ExecContext(ctx, fmt.Sprintf(`PRAGMA foreign_keys = %d`, prevFK)); err != nil {
			fmt.Fprintf(os.Stderr, "[demoapp] restore foreign_keys=%d failed: %v\n", prevFK, err)
		}
	}()

	// Rebuild under a transaction so a failed COPY / DROP leaves the
	// original table intact. BeginTx on a *sql.Conn uses the same
	// underlying connection (the pin survives), so the pragma set
	// above is still in effect for the DROP TABLE.
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		// On any error below, roll back so the original users table
		// is preserved. A deferred rollback after a successful commit
		// is a no-op.
		_ = tx.Rollback()
	}()
	if _, err := tx.ExecContext(ctx, usersRebuildSchema); err != nil {
		return fmt.Errorf("create users_new: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users_new
		    (id, username, password_hash, email, oidc_subject,
		     nhp_public_key, nhp_private_key_enc, nhp_device_id,
		     status, created_at, server_name, cipher_scheme, auth_provider)
		SELECT id, username, password_hash, email, oidc_subject,
		       nhp_public_key, nhp_private_key_enc, nhp_device_id,
		       status, created_at, server_name, cipher_scheme, auth_provider
		  FROM users`); err != nil {
		return fmt.Errorf("copy users to users_new: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DROP TABLE users`); err != nil {
		return fmt.Errorf("drop old users: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE users_new RENAME TO users`); err != nil {
		return fmt.Errorf("rename users_new: %w", err)
	}
	// Reattach the custom case-insensitive username index; the column
	// UNIQUE indexes re-attach automatically via usersRebuildSchema.
	if _, err := tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_ci ON users(LOWER(username))`); err != nil {
		return fmt.Errorf("recreate username ci index: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// hasEmailUniqueIndex reports whether a UNIQUE-constraint-backed index
// covers the email column. PRAGMA index_list returns origin='u' for
// indexes that back a column UNIQUE; we then look up the index_info
// for any matching index and check whether email is among the columns.
//
// PRAGMA index_info is the supported way to enumerate the columns of
// an index — sqlite_schema.sql is NULL for every implicitly-created
// index (a property, not a bug — that is what distinguishes
// sqlite_autoindex_* rows from user-created ones), so a sqlite_master
// scan always reports "no email UNIQUE index" on a fresh legacy DB
// and the migration would never fire.
func (s *UserStore) hasEmailUniqueIndex() (bool, error) {
	rows, err := s.db.Query(`SELECT name, origin FROM pragma_index_list('users')`)
	if err != nil {
		return false, fmt.Errorf("pragma_index_list: %w", err)
	}
	defer rows.Close()
	type idx struct {
		name   string
		origin string
	}
	var uniqueIdxs []idx
	for rows.Next() {
		var i idx
		if err := rows.Scan(&i.name, &i.origin); err != nil {
			return false, fmt.Errorf("scan index list: %w", err)
		}
		// origin 'u' = UNIQUE-constraint-backed, 'pk' = PRIMARY KEY,
		// 'c' = user-created index, 'f' = FOREIGN KEY. We only care
		// about UNIQUE-backed ones — user-created indexes like
		// idx_users_username_ci have origin 'c' and can be ignored.
		if i.origin == "u" {
			uniqueIdxs = append(uniqueIdxs, i)
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate index list: %w", err)
	}
	for _, i := range uniqueIdxs {
		infoRows, err := s.db.Query(fmt.Sprintf(`SELECT name FROM pragma_index_info(%q)`, i.name))
		if err != nil {
			return false, fmt.Errorf("pragma_index_info(%s): %w", i.name, err)
		}
		var cols []string
		for infoRows.Next() {
			var c string
			if err := infoRows.Scan(&c); err != nil {
				infoRows.Close()
				return false, fmt.Errorf("scan index info: %w", err)
			}
			cols = append(cols, c)
		}
		infoRows.Close()
		if err := infoRows.Err(); err != nil {
			return false, fmt.Errorf("iterate index info: %w", err)
		}
		if slices.Contains(cols, "email") {
			return true, nil
		}
	}
	return false, nil
}

// currentForeignKeysOnConn reads the current PRAGMA foreign_keys value on
// the given pinned connection. The pragma returns 0/1 in
// modernc.org/sqlite. Taking an explicit *sql.Conn (rather than going
// through *sql.DB) is required by dropEmailUnique — the pragma is
// connection-scoped, so reading it from a different connection than the
// one we're about to set OFF on would silently disagree.
func currentForeignKeysOnConn(ctx context.Context, conn *sql.Conn) (int, error) {
	var v int
	if err := conn.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&v); err != nil {
		return 0, err
	}
	return v, nil
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
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (`+usersColumnsMinusID+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.Username, u.PasswordHash, u.Email, nullIfEmpty(u.OIDCSubject),
		nullIfEmpty(u.NHPPublicKey), nullIfEmpty(u.NHPPrivateKeyEnc), nullIfEmpty(u.NHPDeviceID),
		string(u.Status), u.CreatedAt.Unix(), u.ServerName, u.CipherScheme, u.AuthProvider,
	)
	if err != nil {
		// Translate SQLite UNIQUE violations into typed sentinel errors
		// so callers can map them to 409 Conflict instead of 500. The
		// check-then-insert in handleRegister is a fast-path UX hint; this
		// insert is the authoritative check, so a race that slips past the
		// pre-check surfaces as a clean 409 rather than an opaque 500
		// (review #4).
		//
		// Email is no longer UNIQUE on the column (review #2: relaxed
		// so IdP sign-in can coexist with a password-squat row), but the
		// match is kept for any database still carrying the constraint
		// during a transitional period.
		if isUniqueViolation(err) {
			msg := err.Error()
			// A violation on the LOWER(username) expression index is
			// reported against the index name, not the column, so match
			// both forms. Email / oidc_subject fire against the column.
			if strings.Contains(msg, "users.username") || strings.Contains(msg, "idx_users_username_ci") {
				return ErrUsernameTaken
			}
			if strings.Contains(msg, "users.email") {
				return ErrEmailTaken
			}
			if strings.Contains(msg, "users.oidc_subject") {
				return ErrSubjectTaken
			}
		}
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
	return s.queryOne(ctx, `SELECT `+usersColumns+` FROM users WHERE LOWER(username) = LOWER(?) LIMIT 1`, strings.TrimSpace(username))
}

// GetUserByEmail looks up by email — used by OIDC to detect a password user
// with a matching email so the OIDC subject is linked rather than duplicating.
//
// Multiple rows may share an email since the column UNIQUE was dropped
// (a password-squat row plus one or more IdP-only rows for the same
// verified address). The lookup is therefore deterministic: an IdP-only
// row (PasswordHash IS NULL OR ”) wins over a password-holder row so
// upsertExternalUser's merge branch always reaches the IdP row; ties
// break by id ASC so a fresh database returns the same row across
// reopens. Password users sign in via /api/login (Username lookup), not
// this one, so the preference never affects them.
func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.queryOne(ctx,
		`SELECT `+usersColumns+` FROM users WHERE LOWER(email) = LOWER(?) `+
			`ORDER BY (password_hash IS NULL OR password_hash = '') DESC, id ASC LIMIT 1`,
		strings.TrimSpace(email))
}

// GetActiveUserByEmail is the status=active counterpart of
// GetUserByEmail — used by handleRegister's pre-check to gate a fresh
// password registration on an existing active holder, without rejecting
// the real owner whose address was previously squatted by an attacker
// who never reached /api/register/confirm (review #3 of the demoapp
// review). The pending row(s) get reaped via LookupRegToken /
// ReapExpiredPendingUsers.
func (s *UserStore) GetActiveUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.queryOne(ctx,
		`SELECT `+usersColumns+` FROM users WHERE LOWER(email) = LOWER(?) AND status = ? `+
			`ORDER BY id ASC LIMIT 1`,
		strings.TrimSpace(email), string(UserStatusActive))
}

// externalSubjectKey namespaces an IdP subject by provider so that a
// GitHub numeric id and an OIDC sub sharing the same string (e.g. a
// numeric sub colliding with a GitHub account id) cannot land on the
// same oidc_subject row. The provider prefix is stored verbatim in the
// unique oidc_subject column, so the existing UNIQUE constraint now
// effectively enforces uniqueness on (provider, subject) without a
// schema rebuild. "|" is chosen as the separator because GitHub ids are
// numeric and OIDC subs are alphanumeric/hyphen — neither contains "|"
// in practice.
func externalSubjectKey(provider, subject string) string {
	return provider + "|" + subject
}

// GetUserByOIDCSubject looks up by the namespaced (provider, subject) key.
func (s *UserStore) GetUserByOIDCSubject(ctx context.Context, provider, subject string) (*User, error) {
	return s.queryOne(ctx, `SELECT `+usersColumns+` FROM users WHERE oidc_subject = ? LIMIT 1`, externalSubjectKey(provider, subject))
}

// UpdateUserOIDCSubject attaches an external subject to an existing user
// row, namespaced by provider. Used by the merge logic when an external
// IdP user signs in with the same email as an existing IdP-only account.
func (s *UserStore) UpdateUserOIDCSubject(ctx context.Context, userID int64, provider, subject string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE users SET oidc_subject = ? WHERE id = ?`, externalSubjectKey(provider, subject), userID)
	if err != nil {
		return fmt.Errorf("update oidc subject: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %d not found", userID)
	}
	return nil
}

// GetUserByID is the primary-key lookup. Used by /api/register/confirm
// (and by requireSession) after they resolve the reg token or session uid.
func (s *UserStore) GetUserByID(ctx context.Context, id int64) (*User, error) {
	return s.queryOne(ctx, `SELECT `+usersColumns+` FROM users WHERE id = ? LIMIT 1`, id)
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
		// Reap the expired row so the reg_tokens table does not accumulate
		// dead tokens (review #2). A failed delete is non-fatal — the row
		// just gets reaped on the next lookup or never matched again.
		_ = s.DeleteRegToken(ctx, rt.Token)
		// Opportunistically reap the corresponding status=pending user row
		// too, so a squatter who abandons a registration cannot leave a
		// permanent reservation on someone else's email (review #3). Best-
		// effort: a failed delete is non-fatal — the user row gets reaped
		// on the next LookupRegToken / ReapExpiredPendingUsers call.
		_, _ = s.db.ExecContext(ctx,
			`DELETE FROM users WHERE id = ? AND status = ?`, rt.UserID, string(UserStatusPending))
		return nil, nil, ErrRegTokenExpired
	}
	user, err := s.GetUserByID(ctx, uid)
	if err != nil {
		return nil, nil, err
	}
	return &rt, user, nil
}

// ReapExpiredPendingUsers deletes every status=pending user that once
// held a reg_token (so they were created via the password /api/register
// path, which always pairs a CreateUser with a SaveRegToken) and whose
// every reg_token has now expired. This is the bulk counterpart to the
// opportunistic single-row reap in LookupRegToken — the opportunistic
// path only fires for the specific user the caller happened to look up,
// so an attacker who abandons a registration never to be looked up again
// would otherwise pin the address until the row is manually cleaned.
//
// Crucially, the SQL only matches rows that have a reg_token AT ALL.
// External-IdP pending users (OIDC/GitHub — created via
// upsertExternalUser, no reg_token) are NOT reaped here: their reg-token
// absence is a normal state, not a sign of an abandoned registration,
// and the complete-registration view (rather than a reg-token lookup)
// drives their activation. Reaping them would silently drop pending
// IdP users before they finish the NHP_REG handshake.
//
// Best-effort: errors are returned for logging only.
func (s *UserStore) ReapExpiredPendingUsers(ctx context.Context) (int64, error) {
	now := time.Now().UTC().Unix()
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM users
		 WHERE status = ?
		   AND EXISTS (
		       SELECT 1 FROM reg_tokens t
		        WHERE t.user_id = users.id
		   )
		   AND NOT EXISTS (
		       SELECT 1 FROM reg_tokens t
		        WHERE t.user_id = users.id
		          AND t.expires_at > ?
		   )`, string(UserStatusPending), now)
	if err != nil {
		return 0, fmt.Errorf("reap expired pending users: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// DeleteRegToken clears a registration token, used after confirm or on expiry.
func (s *UserStore) DeleteRegToken(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM reg_tokens WHERE token = ?`, token)
	if err != nil {
		return fmt.Errorf("delete reg token: %w", err)
	}
	return nil
}

// Errors surfaced by UserStore lookups. Routes translate these to 400/404/409.
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrRegTokenNotFound = errors.New("registration token not found")
	ErrRegTokenExpired  = errors.New("registration token expired")
	// ErrUsernameTaken / ErrEmailTaken / ErrSubjectTaken are returned by
	// CreateUser when the corresponding UNIQUE constraint fires. Callers
	// map them to 409 Conflict (review #4).
	ErrUsernameTaken = errors.New("username already exists")
	ErrEmailTaken    = errors.New("email already exists")
	ErrSubjectTaken  = errors.New("external subject already linked to an account")
)

// isUniqueViolation reports whether err is a SQLite UNIQUE constraint
// violation. modernc.org/sqlite surfaces these as errors whose message
// contains "UNIQUE constraint failed" (extended result code 2067). Used by
// CreateUser to translate constraint fires into typed sentinels.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// queryOne is the shared scanner for single-row lookups.
func (s *UserStore) queryOne(ctx context.Context, q string, args ...any) (*User, error) {
	row := s.db.QueryRowContext(ctx, q, args...)
	u := &User{}
	var (
		pwdHash     sql.NullString
		oidcSub     sql.NullString
		pub         sql.NullString
		privEnc     sql.NullString
		devID       sql.NullString
		statusStr   string
		createdAt   int64
		serverName  sql.NullString
		cipherSch   sql.NullString
		authProvide sql.NullString
	)
	if err := row.Scan(
		&u.ID, &u.Username, &pwdHash, &u.Email, &oidcSub,
		&pub, &privEnc, &devID, &statusStr, &createdAt,
		&serverName, &cipherSch, &authProvide,
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
	u.AuthProvider = authProvide.String
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
