package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/OpenNHP/opennhp/nhp/audit"
	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// defaultAuditLedgerFile is the ledger path used when [Audit] is enabled but
// FilePath is left blank. Resolved against the executable directory. It sits
// in its own audit/ directory rather than logs/ so a logrotate rule pointed
// at logs/ cannot rename or truncate the ledger out from under the open
// append handle (which would drop entries or make `audit verify` cry
// tampering — see the FilePath rotation note).
const defaultAuditLedgerFile = "audit/audit-ledger.jsonl"

// minSigningKeyLen is the smallest accepted HMAC signing key, in bytes. A
// shorter key is rejected rather than silently used, so an operator cannot
// end up with signed=true and a placeholder key.
const minSigningKeyLen = 32

// errAuditConfig marks an [Audit] configuration error (bad base64, short
// signing key) as opposed to an I/O failure. Start treats it as always
// fatal — a config typo must not silently leave the gateway unaudited.
var errAuditConfig = errors.New("audit: invalid configuration")

// Default disk bounds applied when [Audit] MaxSizeBytes / MaxSegments are
// left at 0. ~256 MiB per segment * 20 kept ≈ 5 GiB. A negative config
// value disables the respective cap.
const (
	defaultAuditMaxSizeBytes = 256 * 1024 * 1024
	defaultAuditMaxSegments  = 20
)

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
			return fmt.Errorf("%w: SigningKeyBase64 is not valid base64: %v", errAuditConfig, err)
		}
		// A short key is almost always a placeholder or a fat-fingered
		// value; accepting it would log signed=true while offering trivially
		// forgeable protection. HMAC-SHA256's block size is 64 bytes, but a
		// 32-byte (256-bit) minimum is the accepted floor and matches the
		// key `head -c 32 /dev/urandom | base64` in the docs.
		if len(key) < minSigningKeyLen {
			return fmt.Errorf("%w: SigningKeyBase64 decodes to %d bytes; need at least %d (generate with: head -c 32 /dev/urandom | base64)",
				errAuditConfig, len(key), minSigningKeyLen)
		}
		hmacKey = key
	}

	// Bound disk use by default. NHP_OTP / NHP_REG are audited before the
	// peer is validated, so "keep everything, forever" (0/0) hands the first
	// operator who enables auditing a disk-exhaustion vector. A negative
	// value is the explicit "I archive segments off-box, don't cap me" opt-out.
	maxSize, maxSegs := s.config.Audit.MaxSizeBytes, s.config.Audit.MaxSegments
	if maxSize == 0 {
		maxSize = defaultAuditMaxSizeBytes
	} else if maxSize < 0 {
		maxSize = 0
	}
	if maxSegs == 0 {
		maxSegs = defaultAuditMaxSegments
	} else if maxSegs < 0 {
		maxSegs = 0
	}

	opts := audit.Options{
		HMACKey:      hmacKey,
		Fsync:        s.config.Audit.Fsync,
		Async:        s.config.Audit.Async,
		QueueSize:    s.config.Audit.AsyncQueueSize,
		MaxSizeBytes: maxSize,
		MaxSegments:  maxSegs,
	}
	ledger, err := audit.Open(path, opts)
	if err != nil {
		if errors.Is(err, audit.ErrNotALedger) && !s.config.Audit.FailClosed {
			// The file at FilePath is not one of our ledgers. Two cases:
			//
			//  - It IS our ledger but its first line got corrupted (an
			//    attacker prepending junk to disable auditing, say). Move it
			//    aside to a ".corrupt-<ns>" sibling and start fresh, so the
			//    trail is not silently lost — a chain that restarts at seq 1
			//    next to a .corrupt file is a loud, detectable signal.
			//
			//  - It is a FOREIGN file (a mistyped FilePath pointing at
			//    another log, a config, a shared-volume file). Renaming that
			//    is exactly what ensureLedgerFile exists to prevent — the
			//    server, often privileged, must not move an operator's
			//    unrelated file around. Leave it untouched and continue at a
			//    FIXED ".quarantined.jsonl" sibling — a fixed name so a
			//    flapping service with a typo'd FilePath resumes one
			//    continuous trail instead of piling up seq-1 files.
			if audit.LooksLikeLedger(path) {
				fresh, qErr := quarantineAndReopen(path, opts)
				if qErr != nil {
					return fmt.Errorf("audit ledger at %s is not readable and could not be quarantined: %w", path, qErr)
				}
				s.auditLedger = fresh
				log.Critical("audit ledger %s has a corrupted header (%v); moved it aside and started a fresh chain — investigate the original file", path, err)
				return nil
			}
			sibling := path + ".quarantined.jsonl"
			fresh, qErr := audit.Open(sibling, opts)
			if qErr != nil {
				return fmt.Errorf("audit: %s is not a ledger file and a sibling ledger could not be opened: %w", path, qErr)
			}
			s.auditLedger = fresh
			log.Critical("audit [Audit] FilePath %s is not an audit ledger (%v) and was left untouched — auditing to %s instead; fix FilePath", path, err, sibling)
			return nil
		}
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

// quarantineAndReopen renames a file that could not be opened as a ledger to
// a timestamped ".corrupt-<ns>" sibling and opens a fresh ledger at the
// original path. The rename keeps the original bytes around against accidents
// and casual edits — not against someone with write access to the directory,
// who can delete the sibling just as easily; a timestamp keeps repeated
// failures from colliding.
func quarantineAndReopen(path string, opts audit.Options) (*audit.Ledger, error) {
	aside := fmt.Sprintf("%s.corrupt-%d", path, time.Now().UnixNano())
	if err := os.Rename(path, aside); err != nil {
		return nil, fmt.Errorf("move %q aside: %w", path, err)
	}
	return audit.Open(path, opts)
}

// auditWriteFailLogEvery rate-limits the Critical emitted while ledger
// writes keep failing: the first failure and every Nth after it escalate,
// the rest stay at Error so a sustained outage neither scrolls away behind a
// single line nor floods the log.
const auditWriteFailLogEvery = 100

// auditEvent appends one security event to the ledger. It is safe to call
// when auditing is disabled (nil ledger) — it simply does nothing. A write
// failure is logged but never propagated: an audit-log hiccup must not break
// the request being served. A PERSISTENT failure is different, though — the
// gateway would keep granting access with no trail, exactly what the signed
// ledger is meant to guard against — so a run of consecutive failures is
// escalated to a rate-limited Critical rather than left as one Error.
func (s *UdpServer) auditEvent(evType, severity string, fields map[string]string) {
	if s == nil {
		return
	}
	// Load once: closeAuditLedger no longer nils the field, but a single
	// read still keeps this off any check-then-use hazard.
	ledger := s.auditLedger
	if ledger == nil {
		return
	}
	if err := ledger.Log(evType, severity, fields); err != nil {
		n := s.auditWriteFails.Add(1)
		if n == 1 || n%auditWriteFailLogEvery == 0 {
			log.Critical("audit ledger write failing (%d consecutive): %v — access decisions are no longer being recorded", n, err)
		} else {
			log.Error("audit ledger write failed: %v", err)
		}
		return
	}
	if prev := s.auditWriteFails.Swap(0); prev > 0 {
		log.Warning("audit ledger writes recovered after %d consecutive failure(s)", prev)
	}
}

// closeAuditLedger flushes and closes the ledger on shutdown. The field is
// deliberately NOT nil-ed: a late in-flight handler calling auditEvent after
// this races a check-then-use on a nil write, and Ledger.Log already returns
// an error (not a panic) once the ledger is closed.
func (s *UdpServer) closeAuditLedger() {
	if s.auditLedger != nil {
		_ = s.auditLedger.Close()
	}
}

// decisionGranted reports whether an access/registration decision counts as
// granted for the audit trail. A nil error is the handler's success criterion
// and matches the operational log, but the plugin extension point may legally
// return a failure ErrCode alongside a nil error — a soft denial. Recording
// that as "granted" would let a SIEM rule keyed on result=="denied" miss a
// whole class of rejections, the opposite of what a tamper-evident trail is
// for, so a non-success code denies. An empty or "0" (ErrSuccess) code with no
// error is a grant. The raw code is still kept in its own errCode field either
// way. The bundled plugins always pair a failure code with a non-nil error, so
// this only matters for a misbehaving third-party plugin — which is exactly
// the case the ledger should surface rather than hide.
func decisionGranted(err error, errCode string) bool {
	return err == nil && (errCode == "" || errCode == common.ErrSuccess.ErrorCode())
}

// shortKey returns a compact, log-safe fingerprint of a base64 public key
// for audit fields — enough to correlate, not the whole key.
func shortKey(pubKeyBase64 string) string {
	if len(pubKeyBase64) <= 12 {
		return pubKeyBase64
	}
	return pubKeyBase64[:12]
}
