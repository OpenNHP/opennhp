package server

import (
	"encoding/base64"
	"fmt"
	"path/filepath"

	"github.com/OpenNHP/opennhp/nhp/audit"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// defaultAuditLedgerFile is the ledger path used when [Audit] is enabled
// but FilePath is left blank. Resolved against the executable directory.
const defaultAuditLedgerFile = "logs/audit-ledger.jsonl"

// minSigningKeyLen is the smallest accepted HMAC signing key, in bytes. A
// shorter key is rejected rather than silently used, so an operator cannot
// end up with signed=true and a placeholder key.
const minSigningKeyLen = 32

// initAuditLedger opens the audit ledger when enabled in config. It is a
// no-op (leaving s.auditLedger nil) when auditing is off, so the rest of
// the server can call auditEvent unconditionally.
func (s *UdpServer) initAuditLedger() error {
	if s.config == nil || !s.config.Audit.Enabled {
		return nil
	}

	path := s.config.Audit.FilePath
	if path == "" {
		path = defaultAuditLedgerFile
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(ExeDirPath, path)
	}

	var hmacKey []byte
	if s.config.Audit.SigningKeyBase64 != "" {
		key, err := base64.StdEncoding.DecodeString(s.config.Audit.SigningKeyBase64)
		if err != nil {
			return err
		}
		// A short key is almost always a placeholder or a fat-fingered
		// value; accepting it would log signed=true while offering trivially
		// forgeable protection. HMAC-SHA256's block size is 64 bytes, but a
		// 32-byte (256-bit) minimum is the accepted floor and matches the
		// key `head -c 32 /dev/urandom | base64` in the docs.
		if len(key) < minSigningKeyLen {
			return fmt.Errorf("audit SigningKeyBase64 decodes to %d bytes; need at least %d (generate with: head -c 32 /dev/urandom | base64)",
				len(key), minSigningKeyLen)
		}
		hmacKey = key
	}

	ledger, err := audit.Open(path, audit.Options{
		HMACKey: hmacKey,
		Fsync:   s.config.Audit.Fsync,
	})
	if err != nil {
		return err
	}
	s.auditLedger = ledger
	log.Info("audit ledger enabled at %s (signed=%v)", path, len(hmacKey) > 0)
	if ledger.RepairedOnOpen {
		// Routine after an unclean shutdown and already self-healed: the
		// torn tail was dropped/re-terminated and the chain resumed. Say so
		// at Warning, not Critical — a clean `audit verify` afterward is the
		// expected outcome, and crying wolf here is how operators learn to
		// ignore the ledger's real alarms.
		log.Warning("audit ledger %s: repaired a torn trailing write from an unclean shutdown; chain resumed cleanly", path)
	}
	if ledger.MalformedOnOpen > 0 {
		// Damage that PERSISTS mid-file (not the repaired tail) — a stray or
		// edited line. This one an operator should actually investigate with
		// `audit verify`, which will report the exact line.
		log.Critical("audit ledger %s: %d unparseable line(s) remain after resuming; run 'audit verify' on it",
			path, ledger.MalformedOnOpen)
	}
	return nil
}

// auditEvent appends one security event to the ledger. It is safe to call
// when auditing is disabled (nil ledger) — it simply does nothing. A write
// failure is logged but never propagated: an audit-log hiccup must not
// break the request being served.
func (s *UdpServer) auditEvent(evType, severity string, fields map[string]string) {
	if s == nil || s.auditLedger == nil {
		return
	}
	if err := s.auditLedger.Log(evType, severity, fields); err != nil {
		log.Error("audit ledger write failed: %v", err)
	}
}

// closeAuditLedger flushes and closes the ledger on shutdown.
func (s *UdpServer) closeAuditLedger() {
	if s.auditLedger != nil {
		_ = s.auditLedger.Close()
		s.auditLedger = nil
	}
}

// shortKey returns a compact, log-safe fingerprint of a base64 public key
// for audit fields — enough to correlate, not the whole key.
func shortKey(pubKeyBase64 string) string {
	if len(pubKeyBase64) <= 12 {
		return pubKeyBase64
	}
	return pubKeyBase64[:12]
}
