package audit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

func writeN(t *testing.T, l *Ledger, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if err := l.Log("knock_denied", SeverityWarn, map[string]string{
			"srcIp":  "1.2.3.4",
			"reason": "peer_not_found",
		}); err != nil {
			t.Fatalf("Log #%d: %v", i, err)
		}
	}
}

func TestChainVerifiesIntact(t *testing.T) {
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{})
	writeN(t, l, 5)

	res := VerifyChain(bytes.NewReader(buf.Bytes()), nil)
	if res.Err != nil {
		t.Fatalf("expected intact chain, got err: %v", res.Err)
	}
	if res.Count != 5 {
		t.Fatalf("verified %d entries, want 5", res.Count)
	}
}

func TestDetectsAlteredField(t *testing.T) {
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{})
	writeN(t, l, 5)

	lines := splitLines(buf.Bytes())
	// Tamper with the "reason" field inside entry seq=3 while leaving its
	// stored hash untouched.
	var e Event
	if err := json.Unmarshal(lines[2], &e); err != nil {
		t.Fatal(err)
	}
	e.Fields["reason"] = "totally_fine"
	lines[2] = mustMarshal(t, &e)

	res := VerifyChain(bytes.NewReader(join(lines)), nil)
	if res.Err == nil {
		t.Fatal("expected verification failure for altered field")
	}
	if res.BadSeq != 3 {
		t.Fatalf("BadSeq=%d, want 3", res.BadSeq)
	}
}

func TestDetectsDeletedEntry(t *testing.T) {
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{})
	writeN(t, l, 5)

	lines := splitLines(buf.Bytes())
	// Remove entry seq=3; seq=4's prevHash now points at a hash that is
	// no longer the previous line.
	lines = append(lines[:2], lines[3:]...)

	res := VerifyChain(bytes.NewReader(join(lines)), nil)
	if res.Err == nil {
		t.Fatal("expected verification failure for deleted entry")
	}
	if res.BadSeq != 4 {
		t.Fatalf("BadSeq=%d, want 4", res.BadSeq)
	}
}

func TestDetectsTruncatedTail(t *testing.T) {
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{})
	writeN(t, l, 5)

	lines := splitLines(buf.Bytes())
	lines = lines[:4] // drop the last entry

	// Truncation alone still verifies as a valid (shorter) chain — that is
	// expected, hash chains detect edits/reorders, not a clean tail cut.
	// The count is what reveals the truncation to a caller who knows how
	// many entries there should be.
	res := VerifyChain(bytes.NewReader(join(lines)), nil)
	if res.Err != nil {
		t.Fatalf("truncated-but-consistent chain should verify, got: %v", res.Err)
	}
	if res.Count != 4 {
		t.Fatalf("Count=%d, want 4", res.Count)
	}
}

func TestHMACDetectsForgery(t *testing.T) {
	key := []byte("audit-signing-key")
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{HMACKey: key})
	writeN(t, l, 4)

	// A verifier with the right key accepts it.
	if res := VerifyChain(bytes.NewReader(buf.Bytes()), key); res.Err != nil {
		t.Fatalf("valid signed chain rejected: %v", res.Err)
	}

	// An attacker who rewrites an entry AND recomputes its plain hash (but
	// lacks the key) still fails the HMAC check.
	lines := splitLines(buf.Bytes())
	var e Event
	if err := json.Unmarshal(lines[1], &e); err != nil {
		t.Fatal(err)
	}
	e.Fields["srcIp"] = "9.9.9.9"
	h, err := computeHash(&e)
	if err != nil {
		t.Fatal(err)
	}
	e.Hash = h // recompute the public hash to keep the chain internally consistent
	// but Sig is left stale (attacker cannot recompute without the key)
	lines[1] = mustMarshal(t, &e)

	res := VerifyChain(bytes.NewReader(join(lines)), key)
	if res.Err == nil {
		t.Fatal("expected HMAC verification failure for forged entry")
	}
	if res.BadSeq != 2 {
		t.Fatalf("BadSeq=%d, want 2", res.BadSeq)
	}
}

func TestWrongHMACKeyRejected(t *testing.T) {
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{HMACKey: []byte("right")})
	writeN(t, l, 3)

	res := VerifyChain(bytes.NewReader(buf.Bytes()), []byte("wrong"))
	if res.Err == nil {
		t.Fatal("expected failure verifying with the wrong key")
	}
}

// A signed ledger verified without the key must not look fully verified.
// The hash chain alone is forgeable by anyone who can rewrite the file, so
// the result has to say the signatures went unchecked.
func TestSignedLedgerVerifiedWithoutKeyIsReported(t *testing.T) {
	key := []byte("audit-signing-key")
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{HMACKey: key})
	writeN(t, l, 3)

	res := VerifyChain(bytes.NewReader(buf.Bytes()), nil)
	if res.Err != nil {
		t.Fatalf("chain should still verify without the key: %v", res.Err)
	}
	if res.UncheckedSigs != 3 {
		t.Fatalf("UncheckedSigs=%d, want 3", res.UncheckedSigs)
	}

	// With the key, nothing is left unchecked.
	res = VerifyChain(bytes.NewReader(buf.Bytes()), key)
	if res.Err != nil {
		t.Fatalf("signed chain rejected with the right key: %v", res.Err)
	}
	if res.UncheckedSigs != 0 {
		t.Fatalf("UncheckedSigs=%d, want 0 when the key is supplied", res.UncheckedSigs)
	}
}

// An unsigned ledger has nothing to leave unchecked, so a keyless verify of
// it is a genuinely clean result and must not warn.
func TestUnsignedLedgerReportsNoUncheckedSigs(t *testing.T) {
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{})
	writeN(t, l, 3)

	res := VerifyChain(bytes.NewReader(buf.Bytes()), nil)
	if res.Err != nil {
		t.Fatalf("unsigned chain rejected: %v", res.Err)
	}
	if res.UncheckedSigs != 0 {
		t.Fatalf("UncheckedSigs=%d, want 0 for an unsigned ledger", res.UncheckedSigs)
	}
}

// An unbounded free-text field must not be able to produce a line that
// cannot be read back. Without the write-side bounds, a multi-megabyte
// value would exceed the scanner's per-line cap and the entry would be
// unreadable — which would silently disable auditing on the next restart.
func TestOversizedFieldStaysReadable(t *testing.T) {
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{})

	huge := strings.Repeat("A", 4*maxLineLen)
	if err := l.Log("knock", SeverityWarn, map[string]string{"reason": huge}); err != nil {
		t.Fatal(err)
	}
	if err := l.Log("knock", SeverityInfo, map[string]string{"reason": "ok"}); err != nil {
		t.Fatal(err)
	}

	for i, line := range splitLines(buf.Bytes()) {
		if len(line) > maxLineLen {
			t.Fatalf("line %d is %d bytes, over the %d cap", i, len(line), maxLineLen)
		}
	}

	// Both entries must verify — the chain is intact and nothing was skipped.
	res := VerifyChain(bytes.NewReader(buf.Bytes()), nil)
	if res.Err != nil {
		t.Fatalf("chain with a truncated field failed: %v", res.Err)
	}
	if res.Count != 2 || res.Skipped != 0 {
		t.Fatalf("Count=%d Skipped=%d, want 2/0", res.Count, res.Skipped)
	}

	// The value was kept, just cut, and is marked as such.
	var e Event
	if err := json.Unmarshal(splitLines(buf.Bytes())[0], &e); err != nil {
		t.Fatal(err)
	}
	got := e.Fields["reason"]
	if !strings.HasSuffix(got, truncMarker) {
		t.Fatalf("truncated value not marked: %q", got[max(0, len(got)-40):])
	}
	if !strings.HasPrefix(got, "AAAA") {
		t.Fatal("truncated value lost its content")
	}
}

// Too many fields must also be bounded, and the surviving set has to be
// deterministic — picking by map order would make the entry's hash depend
// on Go's randomized map iteration.
func TestTooManyFieldsBoundedDeterministically(t *testing.T) {
	fields := make(map[string]string)
	for i := 0; i < maxFields*3; i++ {
		fields[fmt.Sprintf("k%03d", i)] = "v"
	}

	render := func() map[string]string {
		var buf bytes.Buffer
		l := NewLedger(&buf, Options{})
		if err := l.Log("knock", SeverityInfo, fields); err != nil {
			t.Fatal(err)
		}
		var e Event
		if err := json.Unmarshal(splitLines(buf.Bytes())[0], &e); err != nil {
			t.Fatal(err)
		}
		return e.Fields
	}

	first := render()
	// maxFields kept plus the _droppedFields marker.
	if len(first) != maxFields+1 {
		t.Fatalf("kept %d fields, want %d", len(first), maxFields+1)
	}
	if first["_droppedFields"] != strconv.Itoa(maxFields*3-maxFields) {
		t.Fatalf("_droppedFields=%q", first["_droppedFields"])
	}

	// Same input, same surviving fields, every time.
	for i := 0; i < 5; i++ {
		if !reflect.DeepEqual(first, render()) {
			t.Fatal("field selection is not deterministic across runs")
		}
	}
}

// Truncation must not split a multi-byte rune, or the entry would carry
// mojibake and json.Marshal would silently substitute U+FFFD.
func TestTruncationRespectsRuneBoundaries(t *testing.T) {
	// "€" is 3 bytes, so a cut at maxFieldValueLen lands mid-rune for some
	// repeat counts — walk a few offsets to hit that case.
	for pad := 0; pad < 4; pad++ {
		v := strings.Repeat("x", pad) + strings.Repeat("€", maxFieldValueLen)
		got := truncate(v, maxFieldValueLen)
		body := strings.TrimSuffix(got, truncMarker)
		if !utf8.ValidString(body) {
			t.Fatalf("pad=%d: truncation produced invalid UTF-8", pad)
		}
	}
}

func TestOpenResumesChainAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "audit.log")

	l1, err := Open(path, Options{})
	if err != nil {
		t.Fatalf("Open #1: %v", err)
	}
	writeN(t, l1, 3)
	if closeErr := l1.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	// Reopen and append more — the chain must continue, not restart.
	l2, err := Open(path, Options{})
	if err != nil {
		t.Fatalf("Open #2: %v", err)
	}
	writeN(t, l2, 2)
	if closeErr := l2.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	res := VerifyChain(bytes.NewReader(data), nil)
	if res.Err != nil {
		t.Fatalf("resumed chain failed to verify: %v", res.Err)
	}
	if res.Count != 5 {
		t.Fatalf("Count=%d, want 5 (3 before restart + 2 after)", res.Count)
	}

	// Confirm seq numbering is continuous 1..5.
	lines := splitLines(data)
	for i, ln := range lines {
		var e Event
		if err := json.Unmarshal(ln, &e); err != nil {
			t.Fatal(err)
		}
		if e.Seq != uint64(i+1) {
			t.Fatalf("entry %d has seq=%d, want %d", i, e.Seq, i+1)
		}
	}
}

// TestOpenToleratesTornTrailingLine covers the crash/power-loss case: a
// partial trailing line must not stop the ledger from opening. The chain
// resumes from the last good entry and the damage is reported, not fatal.
func TestOpenToleratesTornTrailingLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	l1, err := Open(path, Options{})
	if err != nil {
		t.Fatalf("Open #1: %v", err)
	}
	writeN(t, l1, 3)
	if closeErr := l1.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	// Simulate a torn append: a truncated JSON fragment with no newline.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, werr := f.WriteString(`{"seq":4,"time":"2026-01-0`); werr != nil {
		t.Fatal(werr)
	}
	f.Close()

	l2, err := Open(path, Options{})
	if err != nil {
		t.Fatalf("Open must tolerate a torn trailing line, got: %v", err)
	}
	if l2.MalformedOnOpen != 1 {
		t.Errorf("MalformedOnOpen = %d, want 1", l2.MalformedOnOpen)
	}
	// The fragment must be gone: leaving it in place would make
	// `audit verify` report FAILED forever for a benign crash.
	afterOpen, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(afterOpen, []byte(`"time":"2026-01-0`)) {
		t.Error("torn fragment should have been truncated on Open")
	}
	// The chain must continue from entry 3, not restart at 1.
	if logErr := l2.Log("knock", SeverityInfo, nil); logErr != nil {
		t.Fatal(logErr)
	}
	if closeErr := l2.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(data)
	var lastEvt Event
	if uerr := json.Unmarshal(lines[len(lines)-1], &lastEvt); uerr != nil {
		t.Fatalf("last line should be a valid entry: %v", uerr)
	}
	if lastEvt.Seq != 4 {
		t.Errorf("resumed seq = %d, want 4 (continuing after the 3 good entries)", lastEvt.Seq)
	}
}

// TestTornLedgerStillVerifiesAfterReopen is the end-to-end version of the
// guarantee: a crash-damaged ledger, once reopened by the server, must
// verify clean rather than reporting FAILED forever. Operators who get a
// FAILED after every crash stop trusting FAILED at all.
func TestTornLedgerStillVerifiesAfterReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	l1, err := Open(path, Options{})
	if err != nil {
		t.Fatal(err)
	}
	writeN(t, l1, 3)
	if closeErr := l1.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, werr := f.WriteString(`{"seq":4,"ti`); werr != nil {
		t.Fatal(werr)
	}
	f.Close()

	l2, err := Open(path, Options{})
	if err != nil {
		t.Fatal(err)
	}
	writeN(t, l2, 1)
	if closeErr := l2.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	res := VerifyChain(bytes.NewReader(data), nil)
	if res.Err != nil {
		t.Fatalf("a reopened torn ledger must verify clean, got: %v", res.Err)
	}
	if res.Skipped != 0 {
		t.Errorf("Skipped = %d, want 0 (the fragment was truncated on reopen)", res.Skipped)
	}
	if res.Count != 4 {
		t.Errorf("Count = %d, want 4", res.Count)
	}
}

// TestVerifyReportsUnparseableAsDamageNotTampering covers a ledger that
// still contains an unparseable line (e.g. one a rotation tool inserted,
// which Open would not truncate because it is not the trailing fragment):
// it is counted, not treated as a chain break, as long as the committed
// entries still link up.
func TestVerifyReportsUnparseableAsDamageNotTampering(t *testing.T) {
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{})
	writeN(t, l, 3)

	lines := splitLines(buf.Bytes())
	// Insert a junk line between two good entries without touching them.
	withJunk := [][]byte{lines[0], []byte("not json at all"), lines[1], lines[2]}

	res := VerifyChain(bytes.NewReader(join(withJunk)), nil)
	if res.Err != nil {
		t.Fatalf("junk line should not fail the chain, got: %v", res.Err)
	}
	if res.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1", res.Skipped)
	}
	if res.Count != 3 {
		t.Errorf("Count = %d, want 3", res.Count)
	}
}

// TestVerifyStillCatchesReplacedEntry is the security counterpart: tolerating
// unparseable lines must NOT let an attacker hide a removed/rewritten entry.
// Replacing a committed entry with junk breaks the next entry's prevHash.
func TestVerifyStillCatchesReplacedEntry(t *testing.T) {
	var buf bytes.Buffer
	l := NewLedger(&buf, Options{})
	writeN(t, l, 4)

	lines := splitLines(buf.Bytes())
	lines[1] = []byte("garbage that replaced a real entry")

	res := VerifyChain(bytes.NewReader(join(lines)), nil)
	if res.Err == nil {
		t.Fatal("replacing a committed entry with junk must still FAIL verification")
	}
	if res.BadSeq != 3 {
		t.Errorf("BadSeq = %d, want 3 (the entry whose prevHash no longer matches)", res.BadSeq)
	}
}

// shortWriter writes only the first `limit` bytes of the next Write and
// then reports an error, simulating a partial in-process write.
type shortWriter struct {
	buf    bytes.Buffer
	limit  int
	failed bool
}

func (w *shortWriter) Write(p []byte) (int, error) {
	if !w.failed && w.limit > 0 && len(p) > w.limit {
		w.failed = true
		n, _ := w.buf.Write(p[:w.limit])
		return n, io.ErrShortWrite
	}
	return w.buf.Write(p)
}

// TestShortWriteKeepsDamageOnOneLine covers the in-process torn write: a
// failed partial write must not swallow the entry that follows it. The
// fragment is closed off so verification reports damage (Skipped) rather
// than a chain break, which is what an operator needs to tell a disk
// hiccup apart from tampering.
func TestShortWriteKeepsDamageOnOneLine(t *testing.T) {
	w := &shortWriter{limit: 40}
	l := NewLedger(w, Options{})

	// First write is cut short and must return an error.
	if err := l.Log("knock", SeverityInfo, map[string]string{"user": "alice"}); err == nil {
		t.Fatal("expected the short write to report an error")
	}
	// Subsequent writes succeed and must land on their own lines.
	if err := l.Log("knock", SeverityWarn, map[string]string{"user": "bob"}); err != nil {
		t.Fatal(err)
	}
	if err := l.Log("knock", SeverityInfo, map[string]string{"user": "carol"}); err != nil {
		t.Fatal(err)
	}

	res := VerifyChain(bytes.NewReader(w.buf.Bytes()), nil)
	if res.Err != nil {
		t.Fatalf("a short write should degrade to skipped damage, not a chain break: %v", res.Err)
	}
	if res.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1 (the truncated fragment)", res.Skipped)
	}
	if res.Count != 2 {
		t.Errorf("Count = %d, want 2 (both entries written after the failure)", res.Count)
	}
}

func TestEmptyLedgerVerifies(t *testing.T) {
	res := VerifyChain(strings.NewReader(""), nil)
	if res.Err != nil || res.Count != 0 {
		t.Fatalf("empty ledger: got (count=%d, err=%v), want (0, nil)", res.Count, res.Err)
	}
}

// helpers

func splitLines(b []byte) [][]byte {
	var out [][]byte
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		cp := make([]byte, len(sc.Bytes()))
		copy(cp, sc.Bytes())
		out = append(out, cp)
	}
	return out
}

func join(lines [][]byte) []byte {
	var buf bytes.Buffer
	for _, l := range lines {
		buf.Write(l)
		buf.WriteByte('\n')
	}
	return buf.Bytes()
}

func mustMarshal(t *testing.T, e *Event) []byte {
	t.Helper()
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
