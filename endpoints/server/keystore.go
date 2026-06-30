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

// migrate creates tables if they do not exist.
func (s *AgentKeyStore) migrate() error {
	ddl := `
	CREATE TABLE IF NOT EXISTS otp_records (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		usr_id     TEXT NOT NULL,
		dev_id     TEXT NOT NULL,
		otp_code   TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		expires_at INTEGER NOT NULL,
		used       INTEGER DEFAULT 0
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
	_, err := s.db.Exec(ddl)
	return err
}

// ── OTP operations ────────────────────────────────────────────────────────

// OTPParams holds the parameters for generating an OTP.
type OTPParams struct {
	UserId   string
	DeviceId string
	TTL      time.Duration // OTP validity period; defaults to 5 minutes if <= 0
}

// GenerateOTP creates a 6-digit random OTP, stores it in the database, and
// returns the code. Previous unused OTPs for the same user+device are
// invalidated.
func (s *AgentKeyStore) GenerateOTP(p OTPParams) (string, error) {
	if p.TTL <= 0 {
		p.TTL = 5 * time.Minute
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
		`INSERT INTO otp_records (usr_id, dev_id, otp_code, created_at, expires_at) VALUES (?, ?, ?, ?, ?)`,
		p.UserId, p.DeviceId, code, now, expires,
	)
	if err != nil {
		return "", fmt.Errorf("keystore: insert otp: %w", err)
	}

	log.Info("keystore: otp generated for user=%s device=%s", p.UserId, p.DeviceId)
	return code, nil
}

// ValidateOTP checks the OTP for the given user+device. Returns nil on
// success, or a specific error:
//
//	ErrOTPInvalid     — no matching OTP found
//	ErrOTPExpired     — OTP has expired
//	ErrOTPAlreadyUsed — OTP was already used
func (s *AgentKeyStore) ValidateOTP(userId, deviceId, code string) error {
	var id int64
	var expiresAt int64
	var used int
	err := s.db.QueryRow(
		`SELECT id, expires_at, used FROM otp_records
		 WHERE usr_id = ? AND dev_id = ? AND otp_code = ?
		 ORDER BY created_at DESC LIMIT 1`,
		userId, deviceId, code,
	).Scan(&id, &expiresAt, &used)
	if err == sql.ErrNoRows {
		return common.ErrOTPInvalid
	}
	if err != nil {
		return fmt.Errorf("keystore: query otp: %w", err)
	}

	if used != 0 {
		return common.ErrOTPAlreadyUsed
	}

	if time.Now().Unix() > expiresAt {
		return common.ErrOTPExpired
	}

	// Mark as used.
	_, err = s.db.Exec(`UPDATE otp_records SET used = 1 WHERE id = ?`, id)
	if err != nil {
		log.Error("keystore: mark otp used: %v", err)
	}

	log.Info("keystore: otp validated for user=%s device=%s", userId, deviceId)
	return nil
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

// RegisterAgentKey stores an agent's public key. ttlSeconds == 0 stores
// the row with expires_at = NULL (treated as never-expiring by the
// read paths); any positive value sets expires_at = now + ttlSeconds.
// Negative values are clamped to 0.
//
// Returns a specific error:
//
//	ErrPublicKeyAlreadyRegistered — key belongs to a different user
//
// If (userId, deviceId) already exists with the SAME public key, this
// is an idempotent no-op (the existing expires_at is preserved). With
// a DIFFERENT public key, the row is updated and the clock is reset
// to a fresh expires_at.
func (s *AgentKeyStore) RegisterAgentKey(userId, deviceId, pubKey string, ttlSeconds int64) error {
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
	// inserts and key rotations reset the clock.
	_, err = s.db.Exec(
		`INSERT INTO agent_keys (usr_id, dev_id, public_key, cipher, created_at, expires_at, active)
		 VALUES (?, ?, ?, 0, ?, ?, 1)
		 ON CONFLICT(usr_id, dev_id) DO UPDATE SET
		   public_key = excluded.public_key,
		   cipher     = excluded.cipher,
		   created_at = excluded.created_at,
		   expires_at = excluded.expires_at,
		   active     = 1`,
		userId, deviceId, pubKey, now, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("keystore: insert agent key: %w", err)
	}

	log.Info("keystore: agent key registered for user=%s device=%s ttl=%ds", userId, deviceId, ttlSeconds)
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

// ── Helpers ───────────────────────────────────────────────────────────────

// randomDigits generates a string of n random decimal digits using
// crypto/rand.
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
