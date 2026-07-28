package audit

import (
	"bufio"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"io"
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
}

// maxReportedSkips bounds the SkippedLines slice; Skipped still counts them
// all.
const maxReportedSkips = 20

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
	br := bufio.NewReaderSize(r, scanBufLen)

	var count uint64
	var skipped uint64
	var unchecked uint64
	var skippedLines []uint64
	prevHash := genesisHash
	var prevSeq uint64
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
			} else if res, ok := verifyEntry(&e, hmacKey, &prevHash, &prevSeq, &count, &unchecked, skipped, skippedLines); !ok {
				return res
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				return VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked, SkippedLines: skippedLines}
			}
			return VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked, SkippedLines: skippedLines, Err: fmt.Errorf("read ledger: %w", readErr)}
		}
	}
}

// verifyEntry checks one parsed entry against the running chain state,
// advancing it on success. ok is false when the entry breaks the chain, in
// which case res carries the failure. The pointer arguments are the running
// state threaded through the scan.
func verifyEntry(e *Event, hmacKey []byte, prevHash *string, prevSeq, count, unchecked *uint64, skipped uint64, skippedLines []uint64) (res VerifyResult, ok bool) {
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
