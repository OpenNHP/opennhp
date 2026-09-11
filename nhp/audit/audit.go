// Package audit provides a tamper-evident security audit ledger for NHP.
//
// The daemons already emit a free-text "[Audit]" log stream, but plain
// text on disk offers no integrity: anyone who can write the file can edit
// or delete lines without leaving a trace. This package records security
// events as append-only JSON lines linked into a hash chain — each entry
// carries the hash of the entry before it, so deleting, editing or
// reordering any line breaks the chain and is detectable after the fact.
//
// An optional HMAC key binds the chain to a secret the log file itself
// does not contain, so an attacker who can rewrite the whole file still
// cannot forge a chain that verifies.
//
// The ledger is opt-in and off by default; enabling it is a config choice
// on the server. When enabled it is the server's structured audit output —
// the nhp/log "[Audit]" stream exists as an API but has no callers, so this
// is not a redundant second copy of an existing trail.
package audit

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// Severity levels for an event, ordered from least to most urgent. These are
// the levels actually emitted today; add a higher one only alongside a call
// site that uses it, so the set never drifts ahead of the code.
const (
	SeverityInfo   = "info"
	SeverityNotice = "notice"
	SeverityWarn   = "warn"
)

// Bounds on the free-text parts of an entry.
//
// A field carrying an unbounded string — an error message, a peer-supplied
// reason — could otherwise produce a line longer than the reader can scan
// back. One such line would count as damage on the next Open and disable
// the resume, a bad outcome for a log whose job is to still be there after
// something goes wrong. So keys and values are truncated and the field
// count is capped.
//
// The bound must hold on the *marshaled* line, not the raw inputs: JSON
// escapes bytes < 0x20 and <, >, & as \u00XX, up to a 6x expansion. The
// values below are sized so the worst case still fits under maxLineLen:
//
//	maxFields * (6*maxFieldKeyLen + 6*maxFieldValueLen + perFieldPunct)
//	  = 64 * (6*128 + 6*2048 + ~8)  ≈ 836 KB
//
// plus a fixed envelope (seq, time, two 64-hex hashes, field labels; the
// 6x-escaped Type/Severity add ~1.5 KB) — well under the 1 MiB cap with
// room to spare. Truncating loses detail; dropping the ledger loses
// everything.
const (
	// maxLineLen is the per-line cap used when reading the ledger back.
	// The write-side bounds above are sized against it.
	maxLineLen = 1024 * 1024
	// scanBufLen is the initial (growable) reader buffer.
	scanBufLen = 64 * 1024

	maxFieldValueLen = 2048
	maxFieldKeyLen   = 128
	maxFields        = 64
	// truncMarker is appended to a value that was cut, so a reader can
	// tell a truncated value from one that happened to be that long.
	truncMarker = "…[truncated]"
	// droppedFieldsKey records how many fields boundFields removed. It is
	// reserved: a caller-supplied field of the same name is treated as a
	// collision (see boundFields) rather than silently overwritten.
	droppedFieldsKey = "_droppedFields"
)

// ErrNotALedger is returned by Open (via ensureLedgerFile) when the target
// path holds a non-empty file whose first line is not an audit Event. It is
// a distinct, inspectable error so a caller can tell "this is the wrong file"
// apart from a real I/O failure and react accordingly (e.g. quarantine it and
// start a fresh chain rather than run with no audit trail).
var ErrNotALedger = errors.New("audit: file does not look like an audit ledger")

// errAsyncQueueFull is the internal signal that a synthetic "audit_gap"
// marker could not be enqueued; recordGapLocked retries it on the next Log.
var errAsyncQueueFull = errors.New("audit: async queue full (gap marker deferred)")

// eventPrefix is the leading bytes of every marshaled Event: Seq is the
// first struct field with no omitempty, so json.Marshal always starts with
// `{"seq":`. A torn append is a prefix of that, which is how ensureLedgerFile
// tells a torn first write from a foreign single-line file.
var eventPrefix = []byte(`{"seq":`)

// Event is one record in the ledger. Fields are ordered so the JSON
// encoding is deterministic (encoding/json emits struct fields in
// declaration order and map keys sorted), which is what makes the hash
// reproducible during verification.
type Event struct {
	Seq      uint64            `json:"seq"`
	Time     string            `json:"time"`
	Type     string            `json:"type"`
	Severity string            `json:"severity"`
	Fields   map[string]string `json:"fields,omitempty"`
	PrevHash string            `json:"prevHash"`
	Hash     string            `json:"hash"`
	Sig      string            `json:"sig,omitempty"`
}

// chainInput is the canonical byte sequence the Hash is computed over: the
// event without its own Hash/Sig. Keeping it a distinct type (rather than
// blanking fields on Event) guarantees the marshaled shape used for
// hashing never accidentally drifts from what verification recomputes.
type chainInput struct {
	Seq      uint64            `json:"seq"`
	Time     string            `json:"time"`
	Type     string            `json:"type"`
	Severity string            `json:"severity"`
	Fields   map[string]string `json:"fields,omitempty"`
	PrevHash string            `json:"prevHash"`
}

func (e *Event) chainBytes() ([]byte, error) {
	return json.Marshal(chainInput{
		Seq:      e.Seq,
		Time:     e.Time,
		Type:     e.Type,
		Severity: e.Severity,
		Fields:   e.Fields,
		PrevHash: e.PrevHash,
	})
}

// computeHash returns the hex SHA-256 of the event's canonical bytes. This
// is the chain link and is always present, key or no key.
func computeHash(e *Event) (string, error) {
	b, err := e.chainBytes()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// computeSig returns the hex HMAC-SHA256 over "chainBytes || hash" using
// key. Empty string when no key is configured.
func computeSig(e *Event, key []byte) (string, error) {
	if len(key) == 0 {
		return "", nil
	}
	b, err := e.chainBytes()
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	mac.Write(b)
	mac.Write([]byte(e.Hash))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// genesisHash is the PrevHash of the very first entry — a fixed, well-known
// value so an empty ledger has a defined starting link.
const genesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

// Ledger is a concurrency-safe, append-only, hash-chained event writer.
type Ledger struct {
	// MalformedOnOpen is how many unparseable lines REMAIN in the file after
	// Open resumed the chain (the torn trailing fragment a crash leaves is
	// repaired first and is NOT counted here — see RepairedOnOpen). Non-zero
	// means damage that persists mid-file and deserves an `audit verify` and
	// an operator's attention; it is not fatal and the chain simply
	// continues from the last good entry. Read-only after Open.
	MalformedOnOpen int

	// RepairedOnOpen is true when Open fixed a torn trailing write (dropped
	// an unterminated fragment, or re-terminated a last line whose newline
	// was lost). Routine after an unclean shutdown and self-healing, so it
	// warrants an informational note, not an alarm. Read-only after Open.
	RepairedOnOpen bool

	mu       sync.Mutex
	w        io.Writer
	closer   io.Closer
	hmacKey  []byte
	fsync    bool
	seq      uint64
	lastHash string

	// Segment rotation (only when opened via Open with MaxSizeBytes > 0).
	// path is the live segment; when curSize would cross maxSize the live
	// file is renamed to "<path>.<lastSegSeq>" and a fresh one is opened.
	// Every entry already carries prevHash + seq, so the chain stays
	// continuous across the split with no special boundary record.
	path        string
	maxSize     int64
	maxSegments int // retention: keep at most this many "<path>.<n>" files (<=0 = keep all)
	curSize     int64
	lastSegSeq  uint64

	// onPrune, when set, is called with the paths of rotated segments that
	// pruneSegments deleted. Runs on whichever goroutine rotated (the
	// caller's under mu in sync mode, the drain goroutine in async mode), so
	// it must not block or re-enter the Ledger. segmentsPruned is the
	// running total for a Close-time summary.
	onPrune        func(removed []string)
	segmentsPruned atomic.Uint64

	// Async mode: Log computes the entry (seq/hash/sig) under mu and hands
	// the marshaled line to a single background writer via queue, so the
	// request path never blocks on the disk write or fsync. queue is nil in
	// the default synchronous mode.
	async        bool
	drainStarted bool // startAsync spawned the drain goroutine (set once, before any concurrency)
	queue        chan queuedLine
	drainWG      sync.WaitGroup
	dropped      atomic.Uint64
	gapReported  uint64      // dropped count already recorded by an in-chain "audit_gap" marker (guarded by mu)
	closing      atomic.Bool // set by Close before it closes the queue
	// asyncErr holds the most recent background write/rotation failure, or
	// nil once a subsequent write succeeds. A pointer wrapper (not a bare
	// error in an atomic.Value) is required: the drain goroutine stores
	// several distinct concrete error types (*fs.PathError, *fmt.wrapError,
	// …) and atomic.Value panics on a type change.
	asyncErr atomic.Pointer[asyncFailure]
}

// asyncFailure wraps the last background failure so it can be stored/cleared
// atomically regardless of the concrete error type.
type asyncFailure struct{ err error }

// queuedLine is one prepared entry waiting for the async writer. seq is
// carried alongside so the writer can name a rotated segment without racing
// on l.seq.
type queuedLine struct {
	line []byte
	seq  uint64
}

// loadAsyncErr returns the last unrecovered background failure, or nil.
func (l *Ledger) loadAsyncErr() error {
	if f := l.asyncErr.Load(); f != nil {
		return f.err
	}
	return nil
}

// Options configures a Ledger.
type Options struct {
	// HMACKey, when non-empty, adds an HMAC signature to every entry that
	// binds the chain to this secret.
	HMACKey []byte
	// Fsync flushes each entry to stable storage before returning, holding
	// the ledger mutex across the flush. Safer against crash/power loss, but
	// on a gateway "audit volume" tracks knock volume and each access
	// decision then costs a synchronous disk flush that serializes the other
	// audit writers — see the server-side config docs. Weigh it for the
	// workload rather than assuming the log is low-volume.
	Fsync bool
	// Async, when true, moves the actual disk write (and fsync) off the
	// caller's goroutine: Log prepares the entry under the lock — so seq and
	// the hash chain stay strictly ordered — then enqueues it for a single
	// background writer. Keeps Fsync usable on the knock path. If the queue
	// is full the entry is DROPPED (and counted, see Dropped) rather than
	// blocking the request; a dropped entry is discarded whole, so the chain
	// stays contiguous with no gap. Close drains the queue before returning.
	Async bool
	// QueueSize bounds the async queue; <= 0 uses defaultAsyncQueueSize.
	// Ignored unless Async.
	QueueSize int
	// MaxSizeBytes, when > 0, rolls the ledger to a numbered segment
	// ("<path>.<seq>") once it would exceed this size, and continues in a
	// fresh file. The hash chain spans the segments (verify with
	// VerifyLedger / `audit verify`, which pick up the siblings). Only
	// honored when the ledger was opened via Open (a path is required to
	// rename); ignored for a bare io.Writer.
	MaxSizeBytes int64
	// MaxSegments, when > 0, is a retention bound: after a rotation the
	// oldest "<path>.<n>" files are deleted so at most this many remain.
	// With MaxSizeBytes this caps disk use at ~MaxSizeBytes*(MaxSegments+1).
	// A value <= 0 keeps EVERY segment — deleting audit evidence is opt-in.
	// Deleting a segment makes the chain no longer verifiable from seq 1 —
	// `audit verify` then anchors on the first surviving entry and says so —
	// so every prune fires OnPrune.
	MaxSegments int

	// OnPrune, when set, is called after pruneSegments deletes rotated
	// segment files, with the paths removed. It runs on whichever goroutine
	// performed the rotation (the caller's in sync mode, the background
	// writer in async mode), so it must not block or re-enter the Ledger.
	// Deleting audit records is security-relevant; the server wires this to
	// a Critical log line.
	OnPrune func(removed []string)
}

// defaultAsyncQueueSize is the async write queue depth when QueueSize is 0.
const defaultAsyncQueueSize = 4096

// NewLedger writes to w with no restart continuity. Mainly for tests and
// callers that manage their own file handle; production servers use
// Open, which resumes an existing chain across restarts.
func NewLedger(w io.Writer, opts Options) *Ledger {
	l := newLedger(w, opts)
	l.startAsync()
	return l
}

// newLedger builds a Ledger WITHOUT starting the async drain goroutine, so a
// caller (Open) can finish populating the fields the drain reads —
// path/maxSize/curSize/lastSegSeq — before the goroutine that reads them
// exists. Any caller that uses newLedger directly must call startAsync once
// those fields are set.
func newLedger(w io.Writer, opts Options) *Ledger {
	l := &Ledger{
		w:        w,
		hmacKey:  opts.HMACKey,
		fsync:    opts.Fsync,
		lastHash: genesisHash,
		onPrune:  opts.OnPrune,
	}
	if c, ok := w.(io.Closer); ok {
		l.closer = c
	}
	if opts.Async {
		size := opts.QueueSize
		if size <= 0 {
			size = defaultAsyncQueueSize
		}
		l.async = true
		l.queue = make(chan queuedLine, size)
	}
	return l
}

// startAsync launches the background writer when the ledger is in async mode.
// It must be called exactly once, after every field drain reads is set. A
// no-op in synchronous mode or if already started.
func (l *Ledger) startAsync() {
	if !l.async || l.queue == nil || l.drainStarted {
		return
	}
	l.drainStarted = true
	l.drainWG.Add(1)
	// Pass the channel in rather than letting drain read l.queue, which
	// Close nils out under the lock — drain must not touch that field.
	go l.drain(l.queue)
}

// drain is the single background writer for async mode. It is the only
// goroutine that touches l.w / the segment fields while async, so no lock is
// needed around the write or the rotation. A panic here (unlike an ordinary
// write error) would kill the process, so recover and record it.
func (l *Ledger) drain(q <-chan queuedLine) {
	defer l.drainWG.Done()
	defer func() {
		if r := recover(); r != nil {
			l.asyncErr.Store(&asyncFailure{err: fmt.Errorf("audit: async writer panicked: %v", r)})
		}
	}()

	// pending holds bytes a previous Write did not accept. A TRANSIENT
	// failure (ENOSPC, then the operator frees space) is retried without
	// losing committed entries or gapping the chain. Once pending reaches
	// maxPendingBytes the drain STOPS reading the queue and just retries in
	// place: the queue then backpressures Log, whose queue-full path rolls
	// l.seq/l.lastHash back so the on-disk chain stays contiguous. Memory is
	// bounded at maxPendingBytes + the queue. Nothing already chained is
	// ever discarded here.
	var pending []byte // unflushed bytes, always a suffix of the byte stream
	var pendingLastSeq uint64
	retry := time.Millisecond

	// attemptFlush makes one Write of pending and folds the result back into
	// pending/l.curSize/l.asyncErr, resetting retry and stamping
	// l.lastSegSeq on full success. Shared by the per-item retry loop below
	// and the idle retryTimer, so both apply the exact same bookkeeping.
	attemptFlush := func() (drained bool) {
		n, err := l.w.Write(pending)
		if n > 0 {
			// Bytes that landed must never be re-sent (a partial write on
			// ENOSPC returns n>0 with an error); drop them from pending.
			l.curSize += int64(n)
			pending = pending[n:]
		}
		if err != nil {
			l.asyncErr.Store(&asyncFailure{err: err})
			return false
		}
		l.lastSegSeq = pendingLastSeq
		retry = time.Millisecond
		l.asyncErr.Store(nil) // recovered; auditEvent can log "writes recovered"
		if l.fsync {
			if f, ok := l.w.(*os.File); ok {
				_ = f.Sync()
			}
		}
		return true
	}

	// retryTimer re-attempts a flush of `pending` on a timeout even when no
	// new item has arrived to trigger one. Without it, a transient failure
	// on an otherwise quiet gateway (few knocks) left the failed bytes
	// sitting in this goroutine's memory — durable in the caller's mind,
	// since Log had already returned nil — until the next Log call or
	// Close, however long that took. Only armed while pending > 0.
	var retryTimer *time.Timer
	defer func() {
		if retryTimer != nil {
			retryTimer.Stop()
		}
	}()

	for {
		var it queuedLine
		var ok bool
		if len(pending) == 0 {
			it, ok = <-q
		} else {
			if retryTimer == nil {
				retryTimer = time.NewTimer(retry)
			}
			select {
			case it, ok = <-q:
				// Both modules declare go 1.26.0, so Go 1.23+'s newer timer
				// semantics apply: Stop's documentation now guarantees the
				// channel carries no stale value after Stop returns, whether
				// it returns true or false, and says a program must NOT then
				// receive from the channel to "drain" it — unlike the
				// pre-1.23 idiom, that receive is no longer guaranteed to
				// ever complete. Just stop; there is nothing to drain.
				retryTimer.Stop()
				retryTimer = nil
			case <-retryTimer.C:
				retryTimer = nil
				attemptFlush()
				if retry < time.Second {
					retry *= 2
				}
				continue
			}
		}
		if !ok {
			break // queue closed: fall through to the final flush below
		}

		pending = append(pending, it.line...)
		pendingLastSeq = it.seq
		if l.shouldRotate(len(pending)) {
			if rErr := l.rollSegment(); rErr != nil {
				l.asyncErr.Store(&asyncFailure{err: rErr})
				// keep going on the current segment
			}
		}

		for len(pending) > 0 {
			if attemptFlush() {
				break
			}
			if len(pending) < maxPendingBytes {
				break // accept more from the queue; retryTimer covers the rest
			}
			if l.closing.Load() {
				// Shutting down onto a dead writer: the pending (already
				// chained) entries are lost, and so is everything still
				// sitting in q. Close sets closing THEN closes q, so seeing
				// closing here means q is already closed — safe to drain
				// to completion without blocking. Count both: closeAuditLedger's
				// shutdown summary is documented as including abandoned
				// queued entries, not just the ones already merged into
				// pending.
				lost := bytes.Count(pending, []byte{'\n'})
				for range q {
					lost++
				}
				l.dropped.Add(uint64(lost))
				return
			}
			time.Sleep(retry)
			if retry < time.Second {
				retry *= 2
			}
		}
	}
	// Final flush on Close. If a fragment cannot be written, terminate what
	// did land with a newline so it does not merge with a future entry after
	// a restart (Open's torn-tail repair also handles this; a stray blank
	// line from an exact-boundary stop is harmless — readers skip it).
	if len(pending) > 0 && !attemptFlush() && len(pending) > 0 {
		_, _ = l.w.Write([]byte{'\n'})
	}
}

// maxPendingBytes is the point at which the async drain stops reading its
// queue and just retries the pending write in place. Memory is then bounded
// at maxPendingBytes + the queue, and the queue backpressures Log (whose
// queue-full path rolls the chain state back, so nothing is written with a
// gap). No already-chained entry is ever discarded.
const maxPendingBytes = 8 * 1024 * 1024

// shouldRotate reports whether appending nextLen bytes would push the live
// segment past maxSize. curSize > 0 keeps a single over-large entry from
// rotating an empty file forever.
func (l *Ledger) shouldRotate(nextLen int) bool {
	return l.maxSize > 0 && l.curSize > 0 && l.curSize+int64(nextLen) > l.maxSize
}

// Dropped reports how many entries the async writer discarded because the
// queue was full. Always 0 in synchronous mode.
func (l *Ledger) Dropped() uint64 { return l.dropped.Load() }

// Open opens (creating parent dirs as needed) the ledger file at path for
// append. If the file already exists its chain is scanned so new entries
// continue the existing sequence and hash chain — a server restart does
// not start a fresh, disconnected chain.
func Open(path string, opts Options) (*Ledger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, fmt.Errorf("audit: create dir: %w", err)
	}

	// Refuse to touch a non-empty file that is not one of our ledgers. A
	// mistyped FilePath pointing at another existing file would otherwise be
	// appended to, or truncated by the torn-tail repair below. A ledger's
	// first line is always a complete Event, so that is the cheapest
	// reliable signal.
	if err := ensureLedgerFile(path); err != nil {
		return nil, err
	}

	// Repair a torn trailing write (a fragment left by a crash mid-append)
	// BEFORE counting, so the malformed count below reflects only damage
	// that actually remains in the file — a fragment that was repaired must
	// not later drive a Critical "run audit verify" that comes back clean.
	// Done before the append handle is opened: on Windows, truncating a file
	// opened with O_APPEND is refused.
	repaired, err := repairTornTail(path)
	if err != nil {
		return nil, err
	}

	seq, last := uint64(0), genesisHash
	malformed := 0

	// Resume from the tail without reading the whole file: on a long-lived
	// gateway the ledger can be gigabytes, and all a restart needs is the
	// last committed entry's seq + hash. resumeFromTail seeks backwards over
	// the last few lines; only if none of them parse (damage deeper than the
	// window) does it hand back ok=false and we do the O(file) forward scan.
	// One consequence of not reading the whole file: mid-file damage is no
	// longer counted at Open. `audit verify` is the full-integrity pass and
	// reports it precisely.
	if rSeq, rHash, rSkip, rOK, rErr := resumeFromTail(path); rErr != nil {
		return nil, fmt.Errorf("audit: existing ledger %q is unreadable: %w", path, rErr)
	} else if rOK {
		malformed = rSkip
		if rHash != "" {
			seq, last = rSeq, rHash
		}
	} else if f, openErr := os.Open(filepath.Clean(path)); openErr == nil {
		lastSeq, lastHash, bad, scanErr := scanTail(f)
		f.Close()
		if scanErr != nil {
			// Only a real I/O failure gets here; unparseable content is
			// tolerated by scanTail (see below).
			return nil, fmt.Errorf("audit: existing ledger %q is unreadable: %w", path, scanErr)
		}
		malformed = bad
		if lastHash != "" {
			seq, last = lastSeq, lastHash
		}
	} else if !errors.Is(openErr, os.ErrNotExist) {
		return nil, fmt.Errorf("audit: open %q: %w", path, openErr)
	}

	// If the live file yielded no entry (missing, empty, or a crash between
	// rollSegment's rename and the first write to the fresh file), resume
	// from the highest-numbered "<path>.<n>" segment instead. Without this
	// the chain would restart at seq 1 next to a full segment — a spurious
	// "chain broken" for VerifyLedger — and lastSegSeq would be 0, so the
	// next rotation would create "<path>.0" and break segment ordering.
	if last == genesisHash {
		if segSeq, segHash, sErr := lastSegmentTip(path); sErr != nil {
			return nil, fmt.Errorf("audit: read rotated segment of %q: %w", path, sErr)
		} else if segHash != "" {
			seq, last = segSeq, segHash
		}
	}

	f, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("audit: open for append: %w", err)
	}

	var curSize int64
	if info, statErr := f.Stat(); statErr == nil {
		curSize = info.Size()
	}

	// newLedger, not NewLedger: the async drain must not start until every
	// field it reads (path/maxSize/curSize/lastSegSeq) is populated below,
	// otherwise `go test -race` flags a write-after-goroutine-start the
	// moment Log is called from a goroutine not ordered against this return.
	l := newLedger(f, opts) // newLedger already sets l.closer from the io.Closer
	l.seq = seq
	l.lastHash = last
	l.MalformedOnOpen = malformed
	l.RepairedOnOpen = repaired
	l.path = path
	l.maxSize = opts.MaxSizeBytes
	l.maxSegments = opts.MaxSegments
	l.curSize = curSize
	l.lastSegSeq = seq
	l.startAsync()
	return l, nil
}

// lastSegmentTip returns the seq and hash of the last committed entry in the
// highest-numbered "<path>.<n>" segment, or ("", 0, nil) when there are no
// numbered segments. Used by Open to resume when the live file has no entry.
func lastSegmentTip(path string) (seq uint64, hash string, err error) {
	segs, err := numberedSegments(path)
	if err != nil || len(segs) == 0 {
		return 0, "", err
	}
	last := segs[len(segs)-1]
	s, h, _, ok, rErr := resumeFromTail(last.name)
	if rErr != nil {
		return 0, "", rErr
	}
	if !ok {
		if f, oErr := os.Open(filepath.Clean(last.name)); oErr == nil {
			s, h, _, rErr = scanTail(f)
			f.Close()
			if rErr != nil {
				return 0, "", rErr
			}
		}
	}
	return s, h, nil
}

// rollSegment closes the live file, renames it to "<path>.<lastSegSeq>" and
// opens a fresh live file. The caller must own the writer (hold l.mu in sync
// mode, or be the drain goroutine in async mode). On any error the ledger is
// left with a WRITABLE handle to l.path (reopened if the close already
// happened) and rotation is skipped — the file keeps growing but stays
// valid. Only if even that reopen fails does l.w stay a closed file, and
// then the returned error names it.
func (l *Ledger) rollSegment() (err error) {
	if l.path == "" || l.closer == nil {
		return nil // bare io.Writer: nothing to roll
	}
	seg := l.path + "." + strconv.FormatUint(l.lastSegSeq, 10)
	if _, statErr := os.Stat(seg); statErr == nil {
		return fmt.Errorf("audit: segment %q already exists; not rotating", seg)
	}

	// After we close the live handle, every early return must leave l.w
	// pointing at a writable handle to l.path or say it could not. rotated
	// is true once the old data is safely in seg (l.path is then a fresh
	// file and curSize resets); false means we are still on the old file.
	reopenLive := func(cause error, rotated bool) error {
		f, rErr := os.OpenFile(filepath.Clean(l.path), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if rErr != nil {
			return fmt.Errorf("audit: rotation failed AND the live ledger could not be reopened (auditing is now dead until restart): %w; reopen: %v", cause, rErr)
		}
		l.w, l.closer = f, f
		if rotated {
			l.curSize = 0
			l.pruneSegments()
		}
		return fmt.Errorf("audit: rotation degraded, still writable: %w", cause)
	}

	closeErr := l.closer.Close() // the fd is gone whether or not this errors

	if renameErr := os.Rename(l.path, seg); renameErr != nil {
		return reopenLive(renameErr, false)
	}
	// Rename succeeded — l.path no longer exists, so any failure from here
	// re-creates it (O_CREATE) as the fresh segment.
	if closeErr != nil {
		return reopenLive(fmt.Errorf("close live segment: %w", closeErr), true)
	}
	f, err := os.OpenFile(filepath.Clean(l.path), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return reopenLive(fmt.Errorf("open fresh segment: %w", err), true)
	}
	l.w, l.closer = f, f
	l.curSize = 0
	l.pruneSegments()
	return nil
}

// pruneSegments enforces maxSegments: after a rotation, delete the
// lowest-numbered "<path>.<n>" files so at most maxSegments of them remain.
// With MaxSizeBytes this caps total disk use at roughly
// MaxSizeBytes*(maxSegments+1). maxSegments <= 0 keeps everything (the
// default — deleting audit evidence is opt-in). Best-effort: a delete
// failure is not fatal. Every actual deletion is counted and reported
// through onPrune, because a silent drop of audit records is exactly the
// failure mode this feature must not have.
func (l *Ledger) pruneSegments() {
	if l.maxSegments <= 0 {
		return
	}
	segs, err := numberedSegments(l.path)
	if err != nil || len(segs) <= l.maxSegments {
		return
	}
	var removed []string
	for _, s := range segs[:len(segs)-l.maxSegments] {
		if rmErr := os.Remove(s.name); rmErr == nil {
			removed = append(removed, s.name)
		}
	}
	if len(removed) == 0 {
		return
	}
	l.segmentsPruned.Add(uint64(len(removed)))
	if l.onPrune != nil {
		l.onPrune(removed)
	}
}

// SegmentsPruned reports how many rotated segment files retention has
// deleted over this ledger's lifetime. Always 0 unless MaxSegments > 0.
func (l *Ledger) SegmentsPruned() uint64 { return l.segmentsPruned.Load() }

// ensureLedgerFile refuses to modify a non-empty file whose first line is
// not a complete audit Event, so a mistyped FilePath pointing at some other
// file is reported loudly instead of being appended to or (via the torn-tail
// repair) truncated. A missing or empty file is fine — that is a fresh
// ledger.
func ensureLedgerFile(path string) error {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("audit: open %q: %w", path, err)
	}
	defer f.Close()

	// "Empty" must be decided from the file's actual size, not from its
	// first LINE being blank: a file whose first byte is '\n' (not unusual
	// in a hand-edited file an operator mistyped into FilePath) reads as an
	// empty first line via readLine below, even though the file has real
	// content after it. Falling through to "empty file" for that case used
	// to skip this guard entirely — the rest of Open would then proceed to
	// repairTornTail (which can Truncate the file) and start appending
	// audit JSON into the middle of an unrelated file.
	if fi, statErr := f.Stat(); statErr == nil && fi.Size() == 0 {
		return nil
	}

	br := bufio.NewReaderSize(f, scanBufLen)
	var line []byte
	var tooLong bool
	var rerr error
	// Skip leading blank lines — the guard cares about the first REAL
	// content, not literally the first byte — until one is found or the
	// file turns out to hold nothing but blank lines (rerr reaches EOF with
	// line still empty; the check below then correctly refuses it).
	for {
		line, tooLong, rerr = readLine(br)
		if rerr != nil && rerr != io.EOF {
			return fmt.Errorf("audit: read %q: %w", path, rerr)
		}
		if len(line) > 0 || tooLong || rerr == io.EOF {
			break
		}
	}
	var e Event
	if tooLong || json.Unmarshal(line, &e) != nil || e.Seq == 0 {
		// A torn FIRST append (a crash during the first write to a new file
		// or a fresh segment after rollSegment) is a *prefix* of a marshaled
		// Event, which always starts with `{"seq":` (Seq is the first field,
		// no omitempty). Only that exact shape — unterminated AND starting
		// with the event prefix — is let through for repairTornTail to reset;
		// anything else is a foreign file and is refused. A single-line JSON
		// blob with no "seq" (Seq == 0) does NOT match and stays protected.
		if rerr == io.EOF && !tooLong && bytes.HasPrefix(line, eventPrefix) {
			return nil
		}
		return fmt.Errorf("%w: %q (its first non-blank line is not an event); check the [Audit] FilePath setting", ErrNotALedger, path)
	}
	return nil
}

// LooksLikeLedger reports whether any of the first maxScanForEntry lines of
// the file at path parses as an audit Event. It lets a caller tell "this is
// our ledger, its header just got corrupted" (worth quarantining) from "this
// is some unrelated file the operator mistyped into FilePath" (leave it
// alone). A missing or empty file returns false.
func LooksLikeLedger(path string) bool {
	const maxScanForEntry = 64
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return false
	}
	defer f.Close()
	br := bufio.NewReaderSize(f, scanBufLen)
	for i := 0; i < maxScanForEntry; i++ {
		line, tooLong, rerr := readLine(br)
		if !tooLong && len(line) > 0 {
			var e Event
			if json.Unmarshal(line, &e) == nil && e.Seq > 0 && e.Hash != "" {
				return true
			}
		}
		if rerr != nil {
			return false
		}
	}
	return false
}

// repairTornTail fixes a trailing write cut off mid-append (a fragment with
// no terminating newline, left by a crash or power loss). It reports whether
// it changed the file.
//
// A fragment sitting after an earlier complete line is dropped: Log writes
// an entry and its newline in one append, so a fragment with no newline
// after it was never a fully committed record, and dropping it keeps it from
// being concatenated onto the next entry.
//
// If the WHOLE file is one line with no newline, there are two cases: a
// complete Event whose terminating newline was lost (add it back), or a
// fragment that does not parse — a torn FIRST append that ensureLedgerFile
// let through. The latter is reset to an empty file (a fresh chain) rather
// than left as an unrepairable fragment that would forever fail Open.
func repairTornTail(path string) (bool, error) {
	rf, err := os.OpenFile(filepath.Clean(path), os.O_RDWR, 0600)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("audit: open %q for repair: %w", path, err)
	}
	defer rf.Close()

	info, err := rf.Stat()
	if err != nil {
		return false, fmt.Errorf("audit: stat %q: %w", path, err)
	}
	size := info.Size()
	if size == 0 {
		return false, nil
	}

	var lastByte [1]byte
	if _, err := rf.ReadAt(lastByte[:], size-1); err != nil {
		return false, fmt.Errorf("audit: read tail of %q: %w", path, err)
	}
	if lastByte[0] == '\n' {
		return false, nil
	}

	// Walk backwards to the newline that ends the last complete line.
	const chunk = 4096
	buf := make([]byte, chunk)
	keep := int64(-1) // -1 => no newline found anywhere yet
	pos := size
	for pos > 0 {
		n := int64(chunk)
		if pos < n {
			n = pos
		}
		start := pos - n
		if _, err := rf.ReadAt(buf[:n], start); err != nil {
			return false, fmt.Errorf("audit: scan tail of %q: %w", path, err)
		}
		if i := bytes.LastIndexByte(buf[:n], '\n'); i >= 0 {
			keep = start + int64(i) + 1
			break
		}
		pos = start
	}

	if keep < 0 {
		// No newline anywhere: the whole file is one line. If it parses as a
		// complete Event its terminator was just lost — add it back. If it is
		// an unterminated PREFIX of an Event (starts with `{"seq":`) it is a
		// torn first append; reset to a fresh chain (segments, if any, are
		// reconnected by Open via lastSegmentTip). Anything else reached here
		// only because ensureLedgerFile's prefix check let it through, so this
		// is defensive: refuse rather than truncate.
		whole := make([]byte, size)
		if _, err := rf.ReadAt(whole, 0); err != nil && err != io.EOF {
			return false, fmt.Errorf("audit: read %q for repair: %w", path, err)
		}
		var e Event
		if json.Unmarshal(bytes.TrimRight(whole, "\n"), &e) == nil && e.Seq > 0 {
			if _, err := rf.WriteAt([]byte{'\n'}, size); err != nil {
				return false, fmt.Errorf("audit: terminate line in %q: %w", path, err)
			}
			return true, nil
		}
		if !bytes.HasPrefix(whole, eventPrefix) {
			return false, fmt.Errorf("%w: %q (single unterminated line, not an event prefix); check the [Audit] FilePath setting", ErrNotALedger, path)
		}
		if err := rf.Truncate(0); err != nil {
			return false, fmt.Errorf("audit: reset torn first append in %q: %w", path, err)
		}
		return true, nil
	}

	if err := rf.Truncate(keep); err != nil {
		return false, fmt.Errorf("audit: truncate partial line in %q: %w", path, err)
	}
	return true, nil
}

// readLine reads one newline-terminated line from br, returning it without
// the trailing '\n'. A line longer than maxLineLen is not buffered: its
// bytes up to and including the next newline are discarded and tooLong is
// true, so a caller can count it as damage and keep going. This is why the
// ledger readers do not use bufio.Scanner — Scanner returns bufio.ErrTooLong
// and *stops*, turning one oversized line into a fatal read, which is
// exactly the failure the write-side bounds and this reader together avoid.
//
// err is io.EOF once the input is exhausted (with any final unterminated
// bytes returned alongside it), or a real I/O error.
func readLine(br *bufio.Reader) (line []byte, tooLong bool, err error) {
	for {
		frag, e := br.ReadSlice('\n')
		// The terminating '\n' is still part of frag here, so the effective
		// content budget is maxLineLen-1. Unreachable given the ~847 KB write
		// ceiling, but worth stating so the cap isn't read as maxLineLen bytes
		// of content.
		if !tooLong && len(line)+len(frag) > maxLineLen {
			tooLong = true
			line = nil // drop what we had; we will not return an oversized line
		}
		if !tooLong {
			line = append(line, frag...) // frag aliases br's buffer, so copy
		}
		if e == bufio.ErrBufferFull {
			continue // same line continues past the reader's buffer
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		return line, tooLong, e
	}
}

// scanTail reads every line and returns the last parseable entry's seq and
// hash (so Open can continue the chain) plus a count of lines it could not
// parse. It parses only the fields it needs.
//
// Unparseable content is deliberately NOT an error. A crash or power loss
// mid-append leaves a torn trailing line, a log-rotation tool can drop a
// stray line in, and a corrupt/oversized line can appear; refusing to open
// the ledger in those cases would take the whole daemon down over a
// cosmetic log problem. Instead the chain resumes from the last good entry
// and the caller is told how many lines were skipped so it can log loudly.
// Detecting real tampering remains the job of VerifyChain / `audit verify`.
// Only a genuine I/O failure returns an error here.
func scanTail(r io.Reader) (uint64, string, int, error) {
	br := bufio.NewReaderSize(r, scanBufLen)
	var seq uint64
	var hash string
	malformed := 0
	for {
		line, tooLong, err := readLine(br)
		if tooLong {
			malformed++
		} else if len(line) > 0 {
			var e Event
			if json.Unmarshal(line, &e) != nil {
				malformed++
			} else {
				seq, hash = e.Seq, e.Hash
			}
		}
		if err != nil {
			if err == io.EOF {
				return seq, hash, malformed, nil
			}
			return 0, "", malformed, err
		}
	}
}

// maxTailResumeLines is how many trailing lines resumeFromTail will inspect
// before giving up and letting Open fall back to a full forward scan. A
// crash leaves at most one torn line; anything past a handful of unparseable
// trailing lines is damage deep enough that the O(file) scan (and then
// `audit verify`) is the right tool.
const maxTailResumeLines = 64

// resumeFromTail recovers the last committed entry's seq and hash by reading
// only the end of the file, so Open does not cost O(file) on every restart.
//
// It walks backwards collecting up to maxTailResumeLines complete
// (newline-terminated) lines, then parses them newest-first and returns the
// first that is a valid Event, with skipped counting the unparseable trailing
// lines above it. ok is false when none of the inspected lines parse AND the
// window did not reach the start of the file — that is damage deeper than the
// tail, and the caller should fall back to scanTail. A missing or empty file
// returns (0, "", 0, true, nil): a fresh chain.
func resumeFromTail(path string) (seq uint64, hash string, skipped int, ok bool, err error) {
	f, oErr := os.OpenFile(filepath.Clean(path), os.O_RDONLY, 0)
	if oErr != nil {
		if errors.Is(oErr, os.ErrNotExist) {
			return 0, "", 0, true, nil
		}
		return 0, "", 0, false, fmt.Errorf("audit: open %q: %w", path, oErr)
	}
	defer f.Close()

	info, sErr := f.Stat()
	if sErr != nil {
		return 0, "", 0, false, fmt.Errorf("audit: stat %q: %w", path, sErr)
	}
	size := info.Size()
	if size == 0 {
		return 0, "", 0, true, nil
	}

	const chunk = 8192
	// Cap the backward walk by bytes as well as by newline count, so a tail
	// with a long newline-free stretch (corruption, or an attacker with
	// write access appending garbage) cannot make startup read the whole
	// file backwards into memory. A typical committed entry is a few hundred
	// bytes; budget a generous 64 KiB per inspected line rather than the
	// 1 MiB read-back ceiling, keeping the scan (and the bytes.Join that
	// follows) to a few MiB. A ledger whose genuine tail entries are larger
	// than this simply falls through to the O(file) forward scan below —
	// slower on that one restart, never wrong.
	const maxTailBytesPerLine = 64 * 1024
	maxScan := int64(maxTailResumeLines+2) * maxTailBytesPerLine
	// Collect chunks oldest-first, then join once — appending each new chunk
	// in front of a growing buffer would re-copy the whole accumulation
	// every 8 KiB (O(n^2)).
	var chunks [][]byte
	pos := size
	newlines := 0         // '\n' seen so far while scanning back
	reachedStart := false // read all the way to offset 0
	truncated := false    // hit the byte cap before enough newlines
	// We need one more newline than lines we want, to bound the oldest line.
	for newlines <= maxTailResumeLines+1 {
		if pos == 0 {
			reachedStart = true
			break
		}
		if size-pos >= maxScan {
			truncated = true
			break
		}
		n := int64(chunk)
		if pos < n {
			n = pos
		}
		start := pos - n
		c := make([]byte, n)
		if _, rErr := f.ReadAt(c, start); rErr != nil {
			return 0, "", 0, false, fmt.Errorf("audit: read tail of %q: %w", path, rErr)
		}
		chunks = append(chunks, c)
		newlines += bytes.Count(c, []byte{'\n'})
		pos = start
	}
	if truncated {
		// The tail is not usably structured within the byte budget; let the
		// caller do the O(file) forward scan (or fall back to a segment).
		return 0, "", 0, false, nil
	}
	// chunks were appended newest-first; reverse into read order before join.
	for i, j := 0, len(chunks)-1; i < j; i, j = i+1, j-1 {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	}
	buf := bytes.Join(chunks, nil)

	// Split the collected bytes into complete lines. The file ends in '\n'
	// (repairTornTail guaranteed it), so the trailing split element is "".
	parts := bytes.Split(buf, []byte{'\n'})
	if len(parts) > 0 && len(parts[len(parts)-1]) == 0 {
		parts = parts[:len(parts)-1]
	}
	// If we didn't reach the start, the oldest element may be a partial line
	// from the middle of the chunk window — drop it.
	if !reachedStart && len(parts) > 0 {
		parts = parts[1:]
	}

	skip := 0
	for i := len(parts) - 1; i >= 0 && skip <= maxTailResumeLines; i-- {
		if len(parts[i]) == 0 {
			// A blank line is not damage: drain's final flush can emit a bare
			// '\n', and scanTail / verifyChainFrom both ignore empty lines.
			// Counting it here would make initAuditLedger cry "run audit
			// verify" over something `audit verify` then reports as clean.
			continue
		}
		var e Event
		if json.Unmarshal(parts[i], &e) == nil && e.Seq > 0 && e.Hash != "" {
			return e.Seq, e.Hash, skip, true, nil
		}
		skip++
	}
	if reachedStart {
		// The whole file fit in the window and nothing parsed. ensureLedgerFile
		// already vouched for line 1, so this should be unreachable, but treat
		// it as a fresh chain rather than erroring.
		return 0, "", skip, true, nil
	}
	return 0, "", 0, false, nil
}

// sanitizeUTF8 replaces any invalid UTF-8 byte sequence in s with U+FFFD so
// the string is valid UTF-8 going in to json.Marshal. Without this,
// encoding/json's own write-time coercion of an invalid byte to the escape
// sequence "�" is not a fixed point: unmarshaling that escape produces
// the literal rune U+FFFD (valid UTF-8), and re-marshaling it on verify
// emits the raw 3-byte rune rather than the original escape. computeHash is
// called on both sides of that round-trip (once on the in-memory event at
// write time, once on the JSON-decoded event at verify time), so a field
// with invalid UTF-8 hashes differently each time and VerifyChain reports a
// permanent, false "this entry was altered" on it and every entry after.
// Sanitizing before either hash runs makes the two sides agree. Fields can
// carry attacker- or plugin-supplied bytes that are not valid UTF-8 (an
// HTTP User-Agent header, for one — net/http does not validate it), so this
// is reachable, not theoretical.
func sanitizeUTF8(s string) string {
	return strings.ToValidUTF8(s, "�")
}

// truncate cuts s to at most max bytes without splitting a multi-byte rune
// at the cut point, and marks that it was cut. Always sanitizes first (see
// sanitizeUTF8) so a value short enough to skip truncation is still safe to
// hash and verify.
func truncate(s string, max int) string {
	s = sanitizeUTF8(s)
	if len(s) <= max {
		return s
	}
	cut := max
	// If the cut lands inside a rune, step back to that rune's start so a
	// valid multi-byte rune is not split. A well-formed rune has at most
	// three continuation bytes, so cap the step-back at three: past that the
	// bytes are malformed anyway (a run of continuation bytes with no lead
	// byte), and walking further would march down to zero and discard the
	// whole value — the very bug the plain utf8.ValidString(s[:cut]) form
	// had. Keeping the cut where it is leaves invalid bytes that json.Marshal
	// renders as U+FFFD, which is fine; losing the content is not.
	for steps := 0; steps < 3 && cut > 0 && !utf8.RuneStart(s[cut]); steps++ {
		cut--
	}
	return s[:cut] + truncMarker
}

// boundFields caps field count, key length and value length so a single
// entry can never exceed the line size the reader can scan back. It
// returns the input untouched when everything already fits, so the common
// path allocates nothing and the caller's map is never mutated.
func boundFields(fields map[string]string) map[string]string {
	if len(fields) == 0 {
		return fields
	}
	needsWork := len(fields) > maxFields
	if !needsWork {
		for k, v := range fields {
			// A key/value within the length limits still needs the slow
			// (sanitizing) path below if it is not valid UTF-8 — see
			// sanitizeUTF8. Checking validity here, not just length, is
			// what makes truncate's sanitization actually reachable: it
			// otherwise only runs when something is already being rebuilt
			// for an unrelated reason (too many fields, or one too long).
			if len(k) > maxFieldKeyLen || len(v) > maxFieldValueLen ||
				!utf8.ValidString(k) || !utf8.ValidString(v) {
				needsWork = true
				break
			}
		}
	}
	if !needsWork {
		return fields
	}

	// Deterministic selection when there are too many fields: sort the
	// keys and keep the first maxFields. An arbitrary map-order pick would
	// make the entry (and so its hash) depend on Go's map iteration.
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	dropped := 0
	if len(keys) > maxFields {
		dropped = len(keys) - maxFields
		keys = keys[:maxFields]
	}

	out := make(map[string]string, len(keys)+1)
	for _, k := range keys {
		tk := truncate(k, maxFieldKeyLen)
		// Two distinct keys can collapse to the same truncated key, or a
		// caller can pass the reserved marker key itself. Either way, count
		// the loser as dropped rather than silently overwriting — the entry
		// stays honest about how many fields it is not showing.
		if _, taken := out[tk]; taken || tk == droppedFieldsKey {
			dropped++
			continue
		}
		out[tk] = truncate(fields[k], maxFieldValueLen)
	}
	if dropped > 0 {
		out[droppedFieldsKey] = strconv.Itoa(dropped)
	}
	return out
}

// Log appends one event of the given type/severity with optional key/value
// fields. It is safe for concurrent use. The written entry links to the
// previous one via its hash.
//
// Oversized fields are truncated rather than rejected: an entry that
// records slightly less detail is far better than one that cannot be read
// back, which is what an unbounded free-text value would produce.
func (l *Ledger) Log(evType, severity string, fields map[string]string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.async && l.queue == nil {
		return errors.New("audit: ledger is closed")
	}

	// If the async writer dropped entries since we last managed to record it,
	// chain an "audit_gap" marker FIRST so the loss is visible to
	// `audit verify` and a SIEM, not only to the runtime log.
	l.recordGapLocked()

	return l.logLocked(evType, severity, fields, false)
}

// recordGapLocked appends a chained "audit_gap" marker when the async queue
// has dropped entries not yet reflected in the chain, so a verified copy of
// the ledger still shows that N records were lost. Best effort: if the
// marker itself cannot be written (queue still full, disk still bad),
// gapReported is left behind and the next Log retries. l.mu must be held.
func (l *Ledger) recordGapLocked() {
	d := l.dropped.Load()
	if d <= l.gapReported {
		return
	}
	before := l.seq
	_ = l.logLocked("audit_gap", SeverityWarn, map[string]string{
		"dropped": strconv.FormatUint(d, 10),
	}, true)
	if l.seq > before { // the marker made it into the chain
		l.gapReported = d
	}
}

// logLocked builds, chains and writes (or enqueues) one event. l.mu must be
// held. gapMarker is true only for the synthetic "audit_gap" entry: when the
// async queue is full it is rolled back WITHOUT bumping the dropped counter
// (it carries no audit data of its own and recordGapLocked retries it).
func (l *Ledger) logLocked(evType, severity string, fields map[string]string, gapMarker bool) error {
	l.seq++
	prevHash := l.lastHash
	e := Event{
		Seq:      l.seq,
		Time:     time.Now().UTC().Format(time.RFC3339Nano),
		Type:     truncate(evType, maxFieldKeyLen),
		Severity: truncate(severity, maxFieldKeyLen),
		Fields:   boundFields(fields),
		PrevHash: l.lastHash,
	}

	hash, err := computeHash(&e)
	if err != nil {
		l.seq-- // roll back so a failed write does not skip a sequence number
		return err
	}
	e.Hash = hash
	sig, err := computeSig(&e, l.hmacKey)
	if err != nil {
		l.seq--
		return err
	}
	e.Sig = sig

	line, err := json.Marshal(&e)
	if err != nil {
		l.seq--
		return err
	}
	line = append(line, '\n')

	if l.async {
		select {
		case l.queue <- queuedLine{line: line, seq: e.Seq}:
			l.lastHash = e.Hash
			// Still surface the most recent background failure (if any) so
			// auditEvent keeps escalating a genuine, ongoing I/O problem —
			// but the new entry HAS been enqueued, so a transient error that
			// the drain goroutine has since cleared does not permanently
			// stop the trail.
			return l.loadAsyncErr()
		default:
			// Queue full: discard this entry whole and roll the chain state
			// back to before it, so there is no seq gap — the entry simply
			// never happened. The Dropped counter and the returned error are
			// the signal.
			l.seq--
			l.lastHash = prevHash
			if gapMarker {
				// Synthetic marker; not a real audit event lost. Leave the
				// counter alone so recordGapLocked retries next Log.
				return errAsyncQueueFull
			}
			n := l.dropped.Add(1)
			return fmt.Errorf("audit: async queue full — dropped entry (%d dropped in total)", n)
		}
	}

	var rotErr error
	if l.shouldRotate(len(line)) {
		if rErr := l.rollSegment(); rErr != nil {
			// A rotation hiccup must not drop the entry: rollSegment leaves
			// l.w writable, so fall through and append to the current
			// segment. Remember it so it is surfaced through Log's return
			// (auditEvent escalates a persistent failure) instead of the
			// file just silently growing past maxSize.
			rotErr = fmt.Errorf("audit: segment rotation failed (entry still written): %w", rErr)
		}
	}

	if n, werr := l.w.Write(line); werr != nil {
		// The entry's JSON is everything but the trailing newline.
		jsonLen := len(line) - 1
		// If the whole line (JSON + newline) actually landed, the entry is
		// durably committed and correctly terminated — just advance the
		// chain and report the write error. Do NOT write another newline
		// here: that would leave a blank line in the file.
		if n >= len(line) {
			l.curSize += int64(len(line))
			l.lastSegSeq = e.Seq
			l.lastHash = e.Hash
			return werr
		}
		// If the JSON reached the file but only the newline was lost
		// (e.g. ENOSPC on the last byte), the entry IS durably committed.
		// Terminate it and advance the chain as on success — rolling back
		// seq here would make the NEXT entry reuse this seq, which
		// VerifyChain reports as a chain break: a disk-full would then read
		// as tampering. Report the write error, but keep the chain state
		// consistent with what is actually on disk.
		if n == jsonLen {
			if _, termErr := l.w.Write([]byte{'\n'}); termErr == nil {
				l.curSize += int64(len(line))
				l.lastSegSeq = e.Seq
				l.lastHash = e.Hash
				return werr
			}
		}
		// The entry itself is incomplete (or we could not even terminate a
		// complete one): roll back so the seq is not skipped, and close off
		// the fragment so the next entry does not merge onto it and turn one
		// damaged line into two. Open's repair handles it on the next
		// restart if this terminating write also fails.
		l.seq--
		// Whatever DID land is on disk — count it so rotation accounting
		// doesn't drift. n is at most len(line) by the io.Writer contract.
		if n > 0 {
			l.curSize += int64(n)
			if line[n-1] != '\n' {
				if tn, _ := l.w.Write([]byte{'\n'}); tn > 0 {
					l.curSize += int64(tn)
				}
			}
		}
		return werr
	}
	if l.fsync {
		if f, ok := l.w.(*os.File); ok {
			_ = f.Sync()
		}
	}

	l.curSize += int64(len(line))
	l.lastSegSeq = e.Seq
	l.lastHash = e.Hash
	return rotErr // nil unless rotation failed above (the entry was still written)
}

// Close flushes any queued async writes and closes the underlying file if
// the ledger owns one. It is safe to call once; a Log after Close returns an
// error rather than panicking on the closed queue.
func (l *Ledger) Close() error {
	if l.async {
		l.closing.Store(true) // let a drain blocked on a dead writer bail out
		l.mu.Lock()
		q := l.queue
		l.queue = nil // further Log calls see "closed" instead of sending
		l.mu.Unlock()
		if q != nil {
			close(q)
			l.drainWG.Wait()
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}
