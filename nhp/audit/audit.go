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
	"sync"
	"time"
	"unicode/utf8"
)

// Severity levels for an event, ordered from least to most urgent.
const (
	SeverityInfo   = "info"
	SeverityNotice = "notice"
	SeverityWarn   = "warn"
	SeverityAlert  = "alert"
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

// Event is one record in the ledger. Fields are ordered so the JSON
// encoding is deterministic (encoding/json emits struct fields in
// declaration order and map keys sorted), which is what makes the hash
// ErrNotALedger is returned by Open (via ensureLedgerFile) when the target
// path holds a non-empty file whose first line is not an audit Event. It is
// a distinct, inspectable error so a caller can tell "this is the wrong file"
// apart from a real I/O failure and react accordingly (e.g. quarantine it and
// start a fresh chain rather than run with no audit trail).
var ErrNotALedger = errors.New("audit: file does not look like an audit ledger")

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
}

// Options configures a Ledger.
type Options struct {
	// HMACKey, when non-empty, adds an HMAC signature to every entry that
	// binds the chain to this secret.
	HMACKey []byte
	// Fsync flushes each entry to stable storage before returning. Safer
	// against crash/power loss at the cost of throughput; audit logs are
	// low-volume so this is usually worth enabling.
	Fsync bool
}

// NewLedger writes to w with no restart continuity. Mainly for tests and
// callers that manage their own file handle; production servers use
// Open, which resumes an existing chain across restarts.
func NewLedger(w io.Writer, opts Options) *Ledger {
	l := &Ledger{
		w:        w,
		hmacKey:  opts.HMACKey,
		fsync:    opts.Fsync,
		lastHash: genesisHash,
	}
	if c, ok := w.(io.Closer); ok {
		l.closer = c
	}
	return l
}

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
	if f, openErr := os.Open(filepath.Clean(path)); openErr == nil {
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

	f, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("audit: open for append: %w", err)
	}

	l := NewLedger(f, opts)
	l.closer = f
	l.seq = seq
	l.lastHash = last
	l.MalformedOnOpen = malformed
	l.RepairedOnOpen = repaired
	return l, nil
}

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

	line, tooLong, rerr := readLine(bufio.NewReaderSize(f, scanBufLen))
	if rerr != nil && rerr != io.EOF {
		return fmt.Errorf("audit: read %q: %w", path, rerr)
	}
	if len(line) == 0 && !tooLong {
		return nil // empty file
	}
	var e Event
	if tooLong || json.Unmarshal(line, &e) != nil || e.Seq == 0 {
		return fmt.Errorf("%w: %q (its first line is not an event); check the [Audit] FilePath setting", ErrNotALedger, path)
	}
	return nil
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
// If the WHOLE file is one line with no newline, it is NOT truncated to
// zero — that would erase the file. By the time we reach here Open's guard
// has confirmed the first line is a complete Event, so this is a committed
// entry whose terminating newline was lost; the newline is added back rather
// than the entry deleted.
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
		// No newline anywhere: the whole file is one line. The guard has
		// confirmed it is a complete Event, so terminate it rather than
		// erase it (Truncate(0) here would zero the file).
		if _, err := rf.WriteAt([]byte{'\n'}, size); err != nil {
			return false, fmt.Errorf("audit: terminate line in %q: %w", path, err)
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

// truncate cuts s to at most max bytes without splitting a multi-byte rune
// at the cut point, and marks that it was cut.
func truncate(s string, max int) string {
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
			if len(k) > maxFieldKeyLen || len(v) > maxFieldValueLen {
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

	l.seq++
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
	if n, werr := l.w.Write(line); werr != nil {
		// The entry's JSON is everything but the trailing newline.
		jsonLen := len(line) - 1
		// If the whole JSON reached the file and only the newline was lost
		// (e.g. ENOSPC on the last byte), the entry IS durably committed.
		// Terminate it and advance the chain as on success — rolling back
		// seq here would make the NEXT entry reuse this seq, which
		// VerifyChain reports as a chain break: a disk-full would then read
		// as tampering. Report the write error, but keep the chain state
		// consistent with what is actually on disk.
		if n >= jsonLen {
			if _, termErr := l.w.Write([]byte{'\n'}); termErr == nil {
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
		// n is at most len(line) by the io.Writer contract, so the index is safe.
		if n > 0 && line[n-1] != '\n' {
			_, _ = l.w.Write([]byte{'\n'})
		}
		return werr
	}
	if l.fsync {
		if f, ok := l.w.(*os.File); ok {
			_ = f.Sync()
		}
	}

	l.lastHash = e.Hash
	return nil
}

// Close closes the underlying file if the ledger owns one.
func (l *Ledger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}
