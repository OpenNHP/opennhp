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

// RegisterAgentKey stores an agent's public key. Returns a specific error:
//
//	ErrPublicKeyAlreadyRegistered — key belongs to a different user
//
// If (userId, deviceId) already exists, the public key is updated (key rotation).
func (s *AgentKeyStore) RegisterAgentKey(userId, deviceId, pubKey string) error {
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
		// Same user, same key — idempotent, no-op.
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("keystore: query pubkey conflict: %w", err)
	}

	// Upsert: insert or update on (usr_id, dev_id) conflict.
	_, err = s.db.Exec(
		`INSERT INTO agent_keys (usr_id, dev_id, public_key, cipher, created_at, expires_at, active)
		 VALUES (?, ?, ?, 0, ?, NULL, 1)
		 ON CONFLICT(usr_id, dev_id) DO UPDATE SET
		   public_key = excluded.public_key,
		   cipher     = excluded.cipher,
		   created_at = excluded.created_at,
		   active     = 1`,
		userId, deviceId, pubKey, now,
	)
	if err != nil {
		return fmt.Errorf("keystore: insert agent key: %w", err)
	}

	log.Info("keystore: agent key registered for user=%s device=%s", userId, deviceId)
	return nil
}

// GetAgentKey returns the public key for a given user+device, or nil if not
// found.
func (s *AgentKeyStore) GetAgentKey(userId, deviceId string) (*AgentKeyRecord, error) {
	rec := &AgentKeyRecord{}
	var expiresAt sql.NullInt64
	var active int
	err := s.db.QueryRow(
		`SELECT usr_id, dev_id, public_key, cipher, created_at, expires_at, active
		 FROM agent_keys WHERE usr_id = ? AND dev_id = ? AND active = 1`,
		userId, deviceId,
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
// is registered and active in the keystore. Used by peer validation fallback.
func (s *AgentKeyStore) FindAgentByPublicKey(pubKeyBase64 string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM agent_keys WHERE public_key = ? AND active = 1`,
		pubKeyBase64,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("keystore: find agent by pubkey: %w", err)
	}
	return count > 0, nil
}

// IsAgentRegistered returns true if the user+device pair has a registered key.
func (s *AgentKeyStore) IsAgentRegistered(userId, deviceId string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM agent_keys WHERE usr_id = ? AND dev_id = ? AND active = 1`,
		userId, deviceId,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("keystore: check agent registered: %w", err)
	}
	return count > 0, nil
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
