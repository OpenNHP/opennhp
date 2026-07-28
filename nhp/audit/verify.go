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
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, scanBufLen), maxLineLen)

	var count uint64
	var skipped uint64
	var unchecked uint64
	prevHash := genesisHash
	var prevSeq uint64
	lineNo := uint64(0)

	for sc.Scan() {
		lineNo++
		raw := sc.Bytes()
		if len(raw) == 0 {
			continue
		}

		var e Event
		if err := json.Unmarshal(raw, &e); err != nil {
			// Not a record at all — count it as damage and keep going.
			// If it replaced a committed entry, the NEXT entry's prevHash
			// no longer matches and the chain check below reports the
			// break, so this cannot be used to hide tampering.
			skipped++
			continue
		}

		if e.PrevHash != prevHash {
			return VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked, BadSeq: e.Seq, Err: fmt.Errorf("entry seq=%d: prevHash mismatch (chain broken — an earlier entry was altered, deleted or reordered)", e.Seq)}
		}
		if e.Seq != prevSeq+1 {
			return VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked, BadSeq: e.Seq, Err: fmt.Errorf("entry seq=%d: expected seq=%d (an entry was dropped or inserted)", e.Seq, prevSeq+1)}
		}

		wantHash, err := computeHash(&e)
		if err != nil {
			return VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked, BadSeq: e.Seq, Err: fmt.Errorf("entry seq=%d: %w", e.Seq, err)}
		}
		if wantHash != e.Hash {
			return VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked, BadSeq: e.Seq, Err: fmt.Errorf("entry seq=%d: hash mismatch (this entry was altered)", e.Seq)}
		}

		if len(hmacKey) > 0 {
			wantSig, err := computeSig(&e, hmacKey)
			if err != nil {
				return VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked, BadSeq: e.Seq, Err: fmt.Errorf("entry seq=%d: %w", e.Seq, err)}
			}
			if !hmac.Equal([]byte(wantSig), []byte(e.Sig)) {
				return VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked, BadSeq: e.Seq, Err: fmt.Errorf("entry seq=%d: HMAC signature mismatch (wrong key or forged entry)", e.Seq)}
			}
		} else if e.Sig != "" {
			// The entry was signed but we have no key to check it with.
			// Record it so the caller can say the chain was verified but
			// the signatures were not, instead of reporting a clean pass.
			unchecked++
		}

		prevHash = e.Hash
		prevSeq = e.Seq
		count++
	}
	if err := sc.Err(); err != nil {
		return VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked, Err: fmt.Errorf("read ledger: %w", err)}
	}

	return VerifyResult{Count: count, Skipped: skipped, UncheckedSigs: unchecked}
}
