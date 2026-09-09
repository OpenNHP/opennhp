package server

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/log"

	_ "modernc.org/sqlite"
)

// AgentKeyStore manages Agent public keys and OTP records in SQLite.
type AgentKeyStore struct {
	db *sql.DB
}

// DefaultAgentKeyTTLSeconds is the lifetime of a newly-registered agent
// public key when the operator has not configured agentKeyTTLSeconds.
// 24 hours. Mirrors how OTPTTLSeconds is defaulted at the helper layer.
const DefaultAgentKeyTTLSeconds int64 = 86400

// NewAgentKeyStore opens (or creates) the SQLite database at dbPath.
// The directory is created if it does not exist.
func NewAgentKeyStore(dbPath string) (*AgentKeyStore, error) {
	if dbPath == "" {
		dbPath = filepath.Join("data", "nhp_server.db")
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("keystore: create directory %s: %w", dir, err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("keystore: open database %s: %w", dbPath, err)
	}

	// Connection pool tuning — SQLite is single-writer; one open conn is
	// usually correct. Keep a small idle pool for concurrent read queries.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	store := &AgentKeyStore{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("keystore: migrate: %w", err)
	}

	log.Info("keystore: database opened at %s", dbPath)
	return store, nil
}

// Close closes the database connection.
func (s *AgentKeyStore) Close() error {
	return s.db.Close()
}

// migrate creates tables if they do not exist and applies incremental schema
// changes to existing databases.
func (s *AgentKeyStore) migrate() error {
	ddl := `
	CREATE TABLE IF NOT EXISTS otp_records (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		usr_id     TEXT NOT NULL,
		dev_id     TEXT NOT NULL,
		otp_code   TEXT NOT NULL,
		pub_key    TEXT NOT NULL DEFAULT '',
		created_at INTEGER NOT NULL,
		expires_at INTEGER NOT NULL,
		used       INTEGER DEFAULT 0,
		attempts   INTEGER DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_otp_usr_dev ON otp_records(usr_id, dev_id);

	CREATE TABLE IF NOT EXISTS agent_keys (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		usr_id     TEXT NOT NULL,
		dev_id     TEXT NOT NULL,
		public_key TEXT NOT NULL UNIQUE,
		cipher     INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		expires_at INTEGER,
		active     INTEGER DEFAULT 1,
		UNIQUE(usr_id, dev_id)
	);
	CREATE INDEX IF NOT EXISTS idx_agent_usr ON agent_keys(usr_id);
	CREATE INDEX IF NOT EXISTS idx_agent_pubkey ON agent_keys(public_key);
	CREATE INDEX IF NOT EXISTS idx_agent_expires ON agent_keys(expires_at);
	`
	if _, err := s.db.Exec(ddl); err != nil {
		return err
	}

	// Incremental migrations: add columns that may be absent in older databases.
	migrations := []struct {
		table  string
		column string
		ddl    string
	}{
		{"otp_records", "pub_key", "ALTER TABLE otp_records ADD COLUMN pub_key TEXT NOT NULL DEFAULT ''"},
		{"otp_records", "attempts", "ALTER TABLE otp_records ADD COLUMN attempts INTEGER DEFAULT 0"},
	}
	for _, m := range migrations {
		var exists int
		err := s.db.QueryRow(
			`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`,
			m.table, m.column,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("keystore: check column %s.%s: %w", m.table, m.column, err)
		}
		if exists == 0 {
			if _, err := s.db.Exec(m.ddl); err != nil {
				return fmt.Errorf("keystore: migrate %s.%s: %w", m.table, m.column, err)
			}
		}
	}

	return nil
}

// ── OTP operations ────────────────────────────────────────────────────────

// OTPCooldownSeconds is the minimum interval between successive OTP
// generations for the same user+device. A request that arrives before
// the cooldown elapses is rejected with ErrOTPCooldown.
const OTPCooldownSeconds int64 = 60

// MaxOTPPerUserPerWindow caps the total number of OTP generations for a
// single userId across all deviceIds within the cooldown window. This
// closes the deviceId-rotation bypass: without it, an attacker can vary
// the (unauthenticated, attacker-controlled) deviceId on each request to
// evade the per-device cooldown and email-bomb the victim. A legitimate
// multi-device user is unlikely to exceed this limit in normal use.
const MaxOTPPerUserPerWindow = 5

// MaxDistinctDevicesPerUserPerWindow caps the number of distinct deviceIds
// for a single userId within the cooldown window. This bounds the
// disk-growth DoS dimension of the deviceId-rotation attack: even if the
// attacker stays just below the per-user OTP cap, they can only create so
// many distinct (usr_id, dev_id) rows before the sweep cleans them up.
const MaxDistinctDevicesPerUserPerWindow = 5

// OTPParams holds the parameters for generating an OTP.
type OTPParams struct {
	UserId    string
	DeviceId  string
	PublicKey string        // base64-encoded agent public key; bound to the OTP at issuance
	TTL       time.Duration // OTP validity period; defaults to 5 minutes if <= 0
	// CooldownSeconds is the minimum interval between successive OTP
	// generations for the same user+device. Zero means use the package
	// default (OTPCooldownSeconds). A negative value disables the
	// cooldown check entirely (intended for tests).
	CooldownSeconds int64
}

// GenerateOTP creates a 6-digit random OTP, stores it in the database, and
// returns the code. Previous unused OTPs for the same user+device are
// invalidated.
//
// A per-(userId, deviceId) cooldown (OTPCooldownSeconds) is enforced: if
// any OTP was issued for the same user+device within the cooldown window
// the call returns ErrOTPCooldown.
func (s *AgentKeyStore) GenerateOTP(p OTPParams) (string, error) {
	if p.TTL <= 0 {
		p.TTL = 5 * time.Minute
	}

	// Enforce a per-(userId, deviceId) cooldown: reject if any OTP was
	// issued for this user+device within the cooldown window. A negative
	// CooldownSeconds disables the check; zero falls back to the package
	// default (OTPCooldownSeconds).
	if p.CooldownSeconds >= 0 {
		cooldown := p.CooldownSeconds
		if cooldown == 0 {
			cooldown = OTPCooldownSeconds
		}
		cutoff := time.Now().Unix() - cooldown
		var recentCount int
		if err := s.db.QueryRow(
			`SELECT COUNT(*) FROM otp_records
			 WHERE usr_id = ? AND dev_id = ? AND created_at > ?`,
			p.UserId, p.DeviceId, cutoff,
		).Scan(&recentCount); err == nil && recentCount > 0 {
			return "", common.ErrOTPCooldown
		}

		// Per-userId rate limit: cap total OTPs for a user across ALL
		// deviceIds within the cooldown window. This prevents an
		// attacker from bypassing the per-device cooldown by rotating
		// the (unauthenticated) deviceId on each request.
		var userCount int
		if err := s.db.QueryRow(
			`SELECT COUNT(*) FROM otp_records
			 WHERE usr_id = ? AND created_at > ?`,
			p.UserId, cutoff,
		).Scan(&userCount); err == nil && userCount >= MaxOTPPerUserPerWindow {
			return "", common.ErrOTPCooldown
		}

		// Cap distinct deviceIds per user per cooldown window. This
		// bounds the disk-growth DoS vector: even if the attacker
		// distributes requests across deviceIds to stay under the
		// per-user OTP cap, they can only create so many distinct
		// rows before the sweep cleans them up.
		var distinctDevices int
		if err := s.db.QueryRow(
			`SELECT COUNT(DISTINCT dev_id) FROM otp_records
			 WHERE usr_id = ? AND created_at > ?`,
			p.UserId, cutoff,
		).Scan(&distinctDevices); err == nil && distinctDevices >= MaxDistinctDevicesPerUserPerWindow {
			return "", common.ErrOTPCooldown
		}
	}

	code, err := randomDigits(6)
	if err != nil {
		return "", fmt.Errorf("keystore: generate otp: %w", err)
	}

	now := time.Now().Unix()
	expires := time.Now().Add(p.TTL).Unix()

	// Invalidate previous unused OTPs for this user+device.
	_, _ = s.db.Exec(
		`UPDATE otp_records SET used = 1 WHERE usr_id = ? AND dev_id = ? AND used = 0`,
		p.UserId, p.DeviceId,
	)

	_, err = s.db.Exec(
		`INSERT INTO otp_records (usr_id, dev_id, otp_code, pub_key, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)`,
		p.UserId, p.DeviceId, code, p.PublicKey, now, expires,
	)
	if err != nil {
		return "", fmt.Errorf("keystore: insert otp: %w", err)
	}

	log.Info("keystore: otp generated for user=%s device=%s", p.UserId, p.DeviceId)
	return code, nil
}

// MaxOTPAttempts is the number of consecutive incorrect OTP guesses allowed
// before the OTP is invalidated.
const MaxOTPAttempts = 5

// ValidateOTP checks the OTP for the given user+device. Returns nil on
// success, or a specific error:
//
//	ErrOTPInvalid     — no matching OTP found (or wrong code)
//	ErrOTPExpired     — OTP has expired
//	ErrOTPAlreadyUsed — OTP was already used
//	ErrOTPRateLimited — too many failed attempts; OTP has been invalidated
//	ErrOTPPublicKeyMismatch — OTP was issued for a different public key
//
// Each incorrect guess increments an attempt counter on the pending OTP.
// After MaxOTPAttempts failures the OTP is invalidated and ErrOTPRateLimited
// is returned. A successful validation resets the counter.
func (s *AgentKeyStore) ValidateOTP(userId, deviceId, code, pubKey string) error {
	// Try exact match first — the common success path.
	var id int64
	var expiresAt int64
	var used int
	var attempts int
	var storedPubKey string
	err := s.db.QueryRow(
		`SELECT id, expires_at, used, attempts, pub_key FROM otp_records
		 WHERE usr_id = ? AND dev_id = ? AND otp_code = ?
		 ORDER BY created_at DESC LIMIT 1`,
		userId, deviceId, code,
	).Scan(&id, &expiresAt, &used, &attempts, &storedPubKey)

	if err == nil {
		// Code matched.
		if used != 0 {
			if attempts >= MaxOTPAttempts {
				return common.ErrOTPInvalid // rate-limited: don't leak that the code was correct
			}
			return common.ErrOTPAlreadyUsed
		}
		if time.Now().Unix() > expiresAt {
			return common.ErrOTPExpired
		}
		// Verify the registering public key matches the one bound at OTP issuance.
		if storedPubKey != "" && pubKey != storedPubKey {
			return common.ErrOTPPublicKeyMismatch
		}
		// Mark as used — reset attempts to 0 on success.
		_, err = s.db.Exec(`UPDATE otp_records SET used = 1, attempts = 0 WHERE id = ?`, id)
		if err != nil {
			log.Error("keystore: mark otp used: %v", err)
		}
		log.Info("keystore: otp validated for user=%s device=%s", userId, deviceId)
		return nil
	}

	if err != sql.ErrNoRows {
		return fmt.Errorf("keystore: query otp: %w", err)
	}

	// Code did not match — track the failed attempt on the most recent
	// pending (unused, unexpired) OTP for this user+device.
	err = s.db.QueryRow(
		`SELECT id, expires_at, used, attempts FROM otp_records
		 WHERE usr_id = ? AND dev_id = ? AND used = 0
		 ORDER BY created_at DESC LIMIT 1`,
		userId, deviceId,
	).Scan(&id, &expiresAt, &used, &attempts)
	if err == sql.ErrNoRows {
		return common.ErrOTPInvalid
	}
	if err != nil {
		return fmt.Errorf("keystore: query pending otp: %w", err)
	}

	if time.Now().Unix() > expiresAt {
		return common.ErrOTPExpired
	}

	// Increment failed-attempt counter.
	attempts++
	if attempts >= MaxOTPAttempts {
		// Too many attempts — invalidate the OTP.
		_, _ = s.db.Exec(`UPDATE otp_records SET used = 1, attempts = ? WHERE id = ?`, attempts, id)
		log.Warning("keystore: otp rate-limited for user=%s device=%s after %d attempts", userId, deviceId, attempts)
		return common.ErrOTPRateLimited
	}

	_, err = s.db.Exec(`UPDATE otp_records SET attempts = ? WHERE id = ?`, attempts, id)
	if err != nil {
		log.Error("keystore: update otp attempts: %v", err)
	}
	log.Info("keystore: otp attempt %d/%d failed for user=%s device=%s", attempts, MaxOTPAttempts, userId, deviceId)
	return common.ErrOTPInvalid
}

// ── Agent key operations ──────────────────────────────────────────────────

// AgentKeyRecord represents a registered agent public key row.
type AgentKeyRecord struct {
	UserId    string
	DeviceId  string
	PublicKey string // Base64-encoded
	Cipher    int
	CreatedAt int64
	ExpiresAt *int64
	Active    bool
}

// RegisterAgentKey stores an agent's public key and cipher scheme.
// ttlSeconds == 0 stores the row with expires_at = NULL (treated as
// never-expiring by the read paths); any positive value sets
// expires_at = now + ttlSeconds. Negative values are clamped to 0.
//
// Returns a specific error:
//
//	ErrPublicKeyAlreadyRegistered — key belongs to a different user
//
// If (userId, deviceId) already exists with the SAME public key, this
// is an idempotent no-op (the existing expires_at is preserved). With
// a DIFFERENT public key (e.g. cipher scheme switch), the row is
// updated and the clock is reset to a fresh expires_at.
func (s *AgentKeyStore) RegisterAgentKey(userId, deviceId, pubKey string, cipherScheme int, ttlSeconds int64) error {
	if ttlSeconds < 0 {
		ttlSeconds = 0
	}
	now := time.Now().Unix()

	// Check for public key conflict (same key, different user/device).
	var existingUserId string
	err := s.db.QueryRow(
		`SELECT usr_id FROM agent_keys WHERE public_key = ? AND active = 1`,
		pubKey,
	).Scan(&existingUserId)
	if err == nil {
		if existingUserId != userId {
			return common.ErrPublicKeyAlreadyRegistered
		}
		// Same user, same key — idempotent, no-op. Do NOT reset the
		// expiry clock: a re-register attempt for the same key should
		// not extend an already-issued lifetime.
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("keystore: query pubkey conflict: %w", err)
	}

	// Compute expires_at for this registration.
	var expiresAt sql.NullInt64
	if ttlSeconds > 0 {
		expiresAt = sql.NullInt64{Int64: now + ttlSeconds, Valid: true}
	}

	// Upsert: insert or update on (usr_id, dev_id) conflict. Both fresh
	// inserts and key rotations (including cipher scheme switches) reset
	// the clock.
	_, err = s.db.Exec(
		`INSERT INTO agent_keys (usr_id, dev_id, public_key, cipher, created_at, expires_at, active)
		 VALUES (?, ?, ?, ?, ?, ?, 1)
		 ON CONFLICT(usr_id, dev_id) DO UPDATE SET
		   public_key = excluded.public_key,
		   cipher     = excluded.cipher,
		   created_at = excluded.created_at,
		   expires_at = excluded.expires_at,
		   active     = 1`,
		userId, deviceId, pubKey, cipherScheme, now, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("keystore: insert agent key: %w", err)
	}

	log.Info("keystore: agent key registered for user=%s device=%s cipher=%d ttl=%ds", userId, deviceId, cipherScheme, ttlSeconds)
	return nil
}

// GetAgentKey returns the public key for a given user+device, or nil if
// not found OR if the row is past its expires_at. Expired rows are
// indistinguishable from never-registered ones to all callers.
func (s *AgentKeyStore) GetAgentKey(userId, deviceId string) (*AgentKeyRecord, error) {
	rec := &AgentKeyRecord{}
	var expiresAt sql.NullInt64
	var active int
	err := s.db.QueryRow(
		`SELECT usr_id, dev_id, public_key, cipher, created_at, expires_at, active
		 FROM agent_keys
		 WHERE usr_id = ? AND dev_id = ? AND active = 1
		   AND (expires_at IS NULL OR expires_at > ?)`,
		userId, deviceId, time.Now().Unix(),
	).Scan(&rec.UserId, &rec.DeviceId, &rec.PublicKey, &rec.Cipher, &rec.CreatedAt, &expiresAt, &active)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("keystore: get agent key: %w", err)
	}

	rec.Active = active == 1
	if expiresAt.Valid {
		rec.ExpiresAt = &expiresAt.Int64
	}
	return rec, nil
}

// FindAgentByPublicKey returns true if the given base64-encoded public key
// is registered, active, and not expired. This is the gate consulted by
// the noise-layer peer validation fallback; an expired key behaves as if
// the agent were never registered.
func (s *AgentKeyStore) FindAgentByPublicKey(pubKeyBase64 string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM agent_keys
		 WHERE public_key = ? AND active = 1
		   AND (expires_at IS NULL OR expires_at > ?)`,
		pubKeyBase64, time.Now().Unix(),
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("keystore: find agent by pubkey: %w", err)
	}
	return count > 0, nil
}

// IsAgentRegistered returns true if the user+device pair has an active,
// non-expired registered key.
func (s *AgentKeyStore) IsAgentRegistered(userId, deviceId string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM agent_keys
		 WHERE usr_id = ? AND dev_id = ? AND active = 1
		   AND (expires_at IS NULL OR expires_at > ?)`,
		userId, deviceId, time.Now().Unix(),
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("keystore: check agent registered: %w", err)
	}
	return count > 0, nil
}

// GetAgentKeyExpiry returns the expiry status for the given user+device:
//
//	(true,  &ts, nil) — row exists and is active with expires_at = ts
//	(true,  nil,  nil) — row exists and is active with no expiry (NULL)
//	(false, nil,  nil) — row is missing, deactivated, or already expired
//
// Used by the plugin helper to surface "valid until when?" without
// reaching into the keystore itself. The third return value is reserved
// for future I/O errors; today it is always nil when the lookup ran.
func (s *AgentKeyStore) GetAgentKeyExpiry(userId, deviceId string) (bool, *int64, error) {
	var active int
	var expiresAt sql.NullInt64
	err := s.db.QueryRow(
		`SELECT active, expires_at FROM agent_keys WHERE usr_id = ? AND dev_id = ?`,
		userId, deviceId,
	).Scan(&active, &expiresAt)
	if err == sql.ErrNoRows {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, fmt.Errorf("keystore: get agent key expiry: %w", err)
	}
	if active != 1 {
		return false, nil, nil
	}
	if expiresAt.Valid && expiresAt.Int64 <= time.Now().Unix() {
		return false, nil, nil
	}
	if expiresAt.Valid {
		ts := expiresAt.Int64
		return true, &ts, nil
	}
	return true, nil, nil
}

// SweepExpiredDeactivates flips active=0 for any row whose expires_at has
// elapsed. Returns the number of rows updated. NULL expires_at rows are
// never swept (they are configured to never expire). The result of
// FindAgentByPublicKey / IsAgentRegistered does not depend on this
// sweeper — those functions already filter on expires_at — so this
// method is purely a hygiene / index-utility measure.
func (s *AgentKeyStore) SweepExpiredDeactivates() (int64, error) {
	res, err := s.db.Exec(
		`UPDATE agent_keys
		 SET active = 0
		 WHERE active = 1
		   AND expires_at IS NOT NULL
		   AND expires_at <= ?`,
		time.Now().Unix(),
	)
	if err != nil {
		return 0, fmt.Errorf("keystore: sweep expired: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("keystore: sweep rows affected: %w", err)
	}
	return n, nil
}

// SweepStaleOTPs deletes OTP rows that are already used or expired and
// were created more than retentionSeconds ago. Returns the number of rows
// deleted. Unused, non-expired OTPs are never swept. Retention defaults
// to 86400s (24 hours) when passed a negative value. Pass 0 to delete all
// used or expired OTPs regardless of age.
func (s *AgentKeyStore) SweepStaleOTPs(retentionSeconds int64) (int64, error) {
	if retentionSeconds < 0 {
		retentionSeconds = 86400
	}
	cutoff := time.Now().Unix() - retentionSeconds
	res, err := s.db.Exec(
		`DELETE FROM otp_records
		 WHERE created_at < ?
		   AND (used = 1 OR expires_at <= ?)`,
		cutoff, time.Now().Unix(),
	)
	if err != nil {
		return 0, fmt.Errorf("keystore: sweep otp: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("keystore: sweep otp rows affected: %w", err)
	}
	return n, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────

func randomDigits(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("invalid digit count: %d", n)
	}

	buf := make([]byte, n)
	for i := range buf {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		buf[i] = byte('0') + byte(digit.Int64())
	}
	return string(buf), nil
}
