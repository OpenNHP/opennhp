package audit

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// VerifyResult reports the outcome of walking a ledger's hash chain.
type VerifyResult struct {
	// Count is the number of entries verified before verification stopped
	// (all of them when Err is nil).
	Count uint64
	// Err is nil when the whole chain is intact, otherwise it describes
	// the first break and BadSeq identifies the offending entry.
	Err error
	// BadSeq is the seq of the first entry that failed, valid only when
	// Err is non-nil.
	BadSeq uint64
	// Skipped counts lines that could not be parsed as entries at all.
	// These are reported as damage, not tampering: a torn write leaves a
	// fragment that is not a record, and failing the whole verification
	// over it would make every crash look identical to an attack — which
	// is how operators learn to ignore a FAILED result. Real tampering
	// still fails, because removing or rewriting a committed entry breaks
	// the prevHash linkage of the entry after it, which is checked below.
	Skipped uint64
	// UncheckedSigs counts entries that carry an HMAC signature which was
	// not verified because no key was supplied. The hash chain alone only
	// proves internal consistency: anyone who can rewrite the file can
	// recompute every hash, so a clean result here is far weaker than a
	// signature-checked one. Callers must surface this rather than let a
	// signed ledger verified without its key look fully verified.
	UncheckedSigs uint64
	// SkippedLines lists the 1-based line numbers skipped as damage, capped
	// at maxReportedSkips so a wholly garbled file cannot produce an
	// unbounded slice. Which lines are bad is more actionable than a bare
	// count when an operator goes to inspect the file.
	SkippedLines []uint64
	// AnchoredAtSeq is non-zero when verification did not start from the
	// genesis entry (seq 1) — e.g. VerifyLedger over a segment set whose
	// earliest segments were archived away. The first entry's own prevHash
	// is trusted as the anchor, so a break BEFORE AnchoredAtSeq cannot be
	// seen from these bytes alone; compare against an off-host anchor.
	AnchoredAtSeq uint64
}

// maxReportedSkips bounds the SkippedLines slice; Skipped still counts them
// all.
const maxReportedSkips = 20

// VerifyLedger verifies a ledger that may have been rolled into numbered
// segments by MaxSizeBytes. It walks "<path>.<n>" siblings in ascending
// numeric order and then the live "<path>" as one continuous chain, so a
// rotated ledger verifies exactly as an unrotated one would. With no
// siblings it is just VerifyChain over the single file.
func VerifyLedger(path string, hmacKey []byte) VerifyResult {
	segs, err := segmentFiles(path)
	if err != nil {
		return VerifyResult{Err: fmt.Errorf("audit: list segments of %q: %w", path, err)}
	}

	if len(segs) == 0 {
		// Nothing to read is NOT a clean pass — the ledger was deleted,
		// renamed, or the path is wrong. Returning a zero result here would
		// print "OK: 0 entries" and exit 0.
		return VerifyResult{Err: fmt.Errorf("audit: no ledger file or numbered segment found at %q", path)}
	}

	// If the first available entry is not seq 1, earlier segments were
	// archived away (a documented workflow). Anchor the walk on that entry's
	// own prevHash instead of failing on a "chain broken" against genesis.
	startPrevHash, startPrevSeq, anchoredAt := genesisHash, uint64(0), uint64(0)
	if first, fErr := firstEntry(segs[0]); fErr != nil {
		return VerifyResult{Err: fmt.Errorf("audit: read segment %q: %w", segs[0], fErr)}
	} else if first != nil && first.Seq > 1 {
		startPrevHash, startPrevSeq, anchoredAt = first.PrevHash, first.Seq-1, first.Seq
	}

	var readers []io.Reader
	var openFiles []io.Closer
	defer func() {
		for _, f := range openFiles {
			_ = f.Close()
		}
	}()
	for i, s := range segs {
		f, oErr := os.Open(filepath.Clean(s))
		if oErr != nil {
			return VerifyResult{Err: fmt.Errorf("audit: open segment %q: %w", s, oErr)}
		}
		openFiles = append(openFiles, f)
		if i > 0 {
			// A segment left unterminated by a torn write (sync partial-write
			// only writes the final '\n' best-effort, and repairTornTail runs
			// on the live file only) would otherwise merge its last line onto
			// the next segment's first line. A separator '\n' is free —
			// verifyChainFrom skips empty lines.
			readers = append(readers, bytes.NewReader([]byte{'\n'}))
		}
		readers = append(readers, f)
	}
	return verifyChainFrom(io.MultiReader(readers...), hmacKey, startPrevHash, startPrevSeq, anchoredAt)
}

// firstEntry returns the first parseable Event in the file, or nil if the
// file has no parseable line (empty / all damage).
func firstEntry(path string) (*Event, error) {
	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	br := bufio.NewReaderSize(f, scanBufLen)
	for {
		line, tooLong, rErr := readLine(br)
		if !tooLong && len(line) > 0 {
			var e Event
			if json.Unmarshal(line, &e) == nil && e.Seq > 0 {
				return &e, nil
			}
		}
		if rErr != nil {
			if rErr == io.EOF {
				return nil, nil
			}
			return nil, rErr
		}
	}
}

// segmentFiles returns the ordered list of files that make up a possibly
// rotated ledger: every "<path>.<n>" sibling (n a decimal integer) sorted by
// n ascending, followed by "<path>" itself if it exists.
func segmentFiles(path string) ([]string, error) {
	numbered, err := numberedSegments(path)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(numbered)+1)
	for _, s := range numbered {
		out = append(out, s.name)
	}
	if _, statErr := os.Stat(path); statErr == nil {
		out = append(out, path)
	}
	return out, nil
}

type numberedSegment struct {
	name string
	n    uint64
}

// HasNumberedSegment reports whether at least one "<path>.<n>" segment (n a
// decimal integer) exists next to path. Callers use it to tell "the ledger
// was rotated and the live file archived away" from "the path is simply
// wrong" — a ".corrupt-<ns>" / ".quarantined.jsonl" / ".bak" sibling does
// NOT count. Keeps the "<path>.<n>" naming convention defined in one place.
func HasNumberedSegment(path string) bool {
	segs, err := numberedSegments(path)
	return err == nil && len(segs) > 0
}

// numberedSegments lists the "<path>.<n>" siblings (n a decimal integer),
// ascending by n. It uses os.ReadDir + a literal prefix match rather than
// filepath.Glob so a FilePath containing glob metacharacters ('*', '?', '[')
// does not silently miss segments.
func numberedSegments(path string) ([]numberedSegment, error) {
	dir := filepath.Dir(path)
	prefix := filepath.Base(path) + "."
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var segs []numberedSegment
	for _, e := range ents {
		if e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		n, convErr := strconv.ParseUint(strings.TrimPrefix(e.Name(), prefix), 10, 64)
		if convErr != nil {
			continue // ".corrupt-<ns>" and other non-numeric siblings
		}
		segs = append(segs, numberedSegment{filepath.Join(dir, e.Name()), n})
	}
	sort.Slice(segs, func(i, j int) bool { return segs[i].n < segs[j].n })
	return segs, nil
}

// VerifyChain walks the ledger read from r and confirms every entry's hash
// is correct and links to the previous one. If hmacKey is non-empty each
// entry's Sig is checked too; if it is empty, signed entries are counted in
// UncheckedSigs so callers do not present an unsigned pass as a full one.
// It returns how many entries verified and, on the first break, which entry
// failed and why.
//
// The checks per entry are:
//   - the recomputed hash equals the stored Hash (no field was altered);
//   - PrevHash equals the previous entry's Hash (nothing deleted/reordered);
//   - Seq increments by one (nothing dropped);
//   - Sig matches when a key is supplied (chain bound to the secret).
func VerifyChain(r io.Reader, hmacKey []byte) VerifyResult {
	return verifyChainFrom(r, hmacKey, genesisHash, 0, 0)
}

// verifyChainFrom is VerifyChain with an explicit starting anchor. startPrev*
// are genesisHash/0 for a full chain; VerifyLedger passes the first entry's
// own prevHash / seq-1 when the earliest segments have been archived away, in
// which case anchoredAt records where the walk actually began.
func verifyChainFrom(r io.Reader, hmacKey []byte, startPrevHash string, startPrevSeq, anchoredAt uint64) VerifyResult {
	br := bufio.NewReaderSize(r, scanBufLen)

	var count uint64
	var skipped uint64
	var unchecked uint64
	var skippedLines []uint64
	prevHash := startPrevHash
	prevSeq := startPrevSeq
	lineNo := uint64(0)

	// skip records a damaged line: an unparseable or over-long line is
	// counted, not fatal. If it replaced a committed entry, the NEXT
	// entry's prevHash no longer matches and the chain check below reports
	// the break, so tolerating it cannot hide tampering.
	skip := func() {
		skipped++
		if len(skippedLines) < maxReportedSkips {
			skippedLines = append(skippedLines, lineNo)
		}
	}

	for {
		line, tooLong, readErr := readLine(br)
		lineNo++

		if tooLong {
			skip()
		} else if len(line) > 0 {
			var e Event
			if json.Unmarshal(line, &e) != nil {
				skip()
			} else if res, ok := verifyEntry(&e, line, hmacKey, &prevHash, &prevSeq, &count, &unchecked, skipped, skippedLines); !ok {
				res.AnchoredAtSeq = anchoredAt
				return res
			}
		}

		if readErr != nil {
			res := VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked, SkippedLines: skippedLines, AnchoredAtSeq: anchoredAt}
			if readErr == io.EOF {
				return res
			}
			res.Err = fmt.Errorf("read ledger: %w", readErr)
			return res
		}
	}
}

// verifyEntry checks one parsed entry against the running chain state,
// advancing it on success. ok is false when the entry breaks the chain, in
// which case res carries the failure. rawLine is the exact bytes the entry
// was read from (newline stripped). The pointer arguments are the running
// state threaded through the scan.
func verifyEntry(e *Event, rawLine, hmacKey []byte, prevHash *string, prevSeq, count, unchecked *uint64, skipped uint64, skippedLines []uint64) (res VerifyResult, ok bool) {
	fail := func(err error) VerifyResult {
		return VerifyResult{Count: *count, Skipped: skipped, UncheckedSigs: *unchecked, SkippedLines: skippedLines, BadSeq: e.Seq, Err: err}
	}

	if e.PrevHash != *prevHash {
		return fail(fmt.Errorf("entry seq=%d: prevHash mismatch (chain broken — an earlier entry was altered, deleted or reordered)", e.Seq)), false
	}
	if e.Seq != *prevSeq+1 {
		return fail(fmt.Errorf("entry seq=%d: expected seq=%d (an entry was dropped or inserted)", e.Seq, *prevSeq+1)), false
	}

	wantHash, err := computeHash(e)
	if err != nil {
		return fail(fmt.Errorf("entry seq=%d: %w", e.Seq, err)), false
	}
	if wantHash != e.Hash {
		return fail(fmt.Errorf("entry seq=%d: hash mismatch (this entry was altered)", e.Seq)), false
	}

	// The hash and signature are computed from the parsed struct, so they
	// still verify if the on-disk line carries extra, duplicate or reordered
	// keys that encoding/json silently drops or last-wins. Re-marshal the
	// parsed event and require it to equal the raw line byte for byte: Log
	// writes exactly json.Marshal(&e) and Go's encoder is deterministic
	// (declaration-order fields, sorted map keys), so an untampered entry
	// round-trips exactly. Anything else is a forged or mangled record.
	if canon, cErr := json.Marshal(e); cErr == nil && !bytes.Equal(canon, rawLine) {
		return fail(fmt.Errorf("entry seq=%d: non-canonical JSON encoding (extra, duplicate or reordered keys — the record on disk is not what was signed)", e.Seq)), false
	}

	if len(hmacKey) > 0 {
		wantSig, err := computeSig(e, hmacKey)
		if err != nil {
			return fail(fmt.Errorf("entry seq=%d: %w", e.Seq, err)), false
		}
		if !hmac.Equal([]byte(wantSig), []byte(e.Sig)) {
			return fail(fmt.Errorf("entry seq=%d: HMAC signature mismatch (wrong key or forged entry)", e.Seq)), false
		}
	} else if e.Sig != "" {
		// Signed but we have no key to check it with. Record it so the
		// caller can say the chain was verified but the signatures were
		// not, instead of reporting a clean pass.
		*unchecked++
	}

	*prevHash = e.Hash
	*prevSeq = e.Seq
	*count++
	return VerifyResult{}, true
}
