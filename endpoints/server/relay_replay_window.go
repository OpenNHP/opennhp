package server

import (
	"sync"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// relayReplayWindowSize bounds the per-relay anti-replay bitmap. Sized
// at 8192 recent txids per relay, which is large enough for any
// realistic concurrent-client fan-out behind one relay (the relay's
// own on-the-wire envelope rate is one SenderTrxId per packet) but
// small enough to stay cheap: 8192 bits = 128 uint64 words = 1 KiB
// per relay, plus ~50 bytes of map-key overhead.
//
// This size is NOT configurable. There is no global capacity to fill
// and no operator-tunable knob that would change the attack or cost
// shape — every relay's state is fixed-size and independent. Changing
// it would require raising the (already large) reordering tolerance,
// which is bounded in practice by the wire clock of a single relay
// process and not by relay-load configuration.
const relayReplayWindowSize = 8192

// relayReplayBitmapWords is relayReplayWindowSize expressed as uint64
// words for the bitmap backing array.
const relayReplayBitmapWords = relayReplayWindowSize / 64 // 128

// rlyWindow is the per-relay anti-replay state. It tracks the most
// recent W SenderTrxIds the relay has produced (highestTxid + a
// sliding bitmap) plus the corresponding RemoteSendTime (so a relay
// restart with a fresh sendTime is recognized as a fresh sender and
// its low-counter envelope is accepted).
//
// INVARIANT: the bitmap tracks [highestTxid - W + 1, highestTxid],
// where bit i represents txid = (highestTxid - i). Bit 0 always
// represents highestTxid itself. The bitmap is shifted "rightward in
// txid space" as highestTxid advances, so it is always the
// "trailing W" of the relay's monotonic counter.
//
// why this invariant is load-bearing:
//   - the Window check (highestTxid - txid < W) is what bounds how
//     long a captured envelope can be re-injected before it's
//     considered stale. validatePeer's staleness floor bounds the
//     envelope's RECV window; this bounds the txid replay window.
//   - anchoring on highestSendTime (separately from highestTxid)
//     lets a relay process restart re-enter the dedupe: the new
//     process mints a fresh RemoteSendTime, and we accept envelopes
//     with sendTime > highestSendTime as legitimate by resetting
//     the per-relay state. The relay's signature covers
//     RemoteSendTime, so this reset is unforgeable: only the relay's
//     own static key can produce an envelope with a newer sendTime.
type rlyWindow struct {
	highestTxid     uint64
	highestSendTime int64
	bitmap          [relayReplayBitmapWords]uint64
}

// relayReplayWindow holds one rlyWindow per relay pubkey. Memory is
// O(#relays), bounded by the number of distinct relays that have
// ever authenticated an NHP_RLY envelope — never by the volume of
// traffic, so an attacker cannot make it grow by sending more
// envelopes. A relay that disconnects from the server leaves its
// entry in place; growth is therefore bounded by the count of
// distinct relays historically seen, which is bounded in practice by
// server configuration (relays don't migrate at high rate).
//
// CONCURRENCY: a single mutex covers the whole map and every state
// inside it. The critical section is a few memory ops + bit fiddling,
// and contention is bounded by the number of relay→server handler
// goroutines — at the NHP wire rate the lock is held for
// sub-microsecond spans, well under the per-packet envelope cost.
type relayReplayWindow struct {
	mu      sync.Mutex
	windows map[[core.PublicKeySizeEx]byte]*rlyWindow
}

func newRelayReplayWindow() *relayReplayWindow {
	return &relayReplayWindow{
		windows: make(map[[core.PublicKeySizeEx]byte]*rlyWindow),
	}
}

// replayOutcome makes MarkSeen's three-way verdict explicit so the
// caller can distinguish "first observation" from "duplicate" —
// dedupeRecvRLY only treats the latter two as a replay, but tests
// sometimes want to verify "yes, this was new" independently of
// "no, this was not a duplicate".
type replayOutcome int

const (
	replayOutcomeAccepted      replayOutcome = iota // first observation
	replayOutcomeDuplicate                          // txid already in the window
	replayOutcomeOutsideWindow                      // txid below the trailing W
	replayOutcomeInvalidKey                         // caller passed an unusable pubkey
)

// MarkSeen records an authenticated (relay pubkey, SenderTrxId,
// RemoteSendTime) triple against the per-relay sliding window and
// returns the outcome. See rlyWindow for the algorithm invariants.
//
// The two return values are kept for ergonomic compatibility with
// the previous TTL-set implementation (which returned bool):
// callers that only care about accept/reject read the bool; tests
// that care about WHY read the enum.
func (w *relayReplayWindow) MarkSeen(pubkey []byte, txid uint64, sendTime int64) (bool, replayOutcome) {
	if w == nil {
		return false, replayOutcomeInvalidKey
	}
	if len(pubkey) != core.PublicKeySize && len(pubkey) != core.PublicKeySizeEx {
		return false, replayOutcomeInvalidKey
	}
	var key [core.PublicKeySizeEx]byte
	copy(key[:], pubkey)

	w.mu.Lock()
	defer w.mu.Unlock()

	state, ok := w.windows[key]
	if !ok {
		// First envelope from this relay. Anchor the window here.
		state = &rlyWindow{
			highestTxid:     txid,
			highestSendTime: sendTime,
		}
		state.bitmap[0] = 1
		w.windows[key] = state
		return true, replayOutcomeAccepted
	}

	// Restart detection. A relay process that comes back with a
	// fresh handshake gets a new (signed) RemoteSendTime; its
	// txid counter is also fresh (Device.NextCounterIndex seeds
	// from a random number on each new Device). The reset below
	// is what lets a low-counter envelope from the restarted
	// relay land in the window — without it, every legitimate
	// post-restart envelope would be OutsideWindow until the
	// relay sent W envelopes to push the anchor past it.
	//
	// The reset is unforgeable in practice (RemoteSendTime is
	// inside the signed payload, so only the relay's static key
	// can produce a higher sendTime), but we still gate it
	// narrowly: the txid must also have gone backwards past the
	// window. Without that back-of-window check, a stale
	// captured envelope whose sendTime happens to exceed the
	// current anchor would clear the bitmap — the signature
	// still binds sendTime so a captured envelope cannot itself
	// bump it, but tying the reset to a txid that is provably
	// unrecoverable from the bitmap keeps the rule local: we
	// only ever clear state we have already decided is
	// unrecoverable, never state we could still catch a
	// duplicate in.
	if sendTime > state.highestSendTime &&
		txid < state.highestTxid &&
		state.highestTxid-txid >= relayReplayWindowSize {
		for i := range state.bitmap {
			state.bitmap[i] = 0
		}
		state.highestTxid = txid
		state.highestSendTime = sendTime
		state.bitmap[0] = 1
		return true, replayOutcomeAccepted
	}

	// Track the latest signed sendTime. The signature binds
	// sendTime, so taking max() here is just keeping the anchor
	// in sync with the latest legitimate envelope; a stale
	// captured envelope has the old sendTime, which is <= the
	// current anchor and does not change anything. This branch
	// never touches the bitmap.
	if sendTime > state.highestSendTime {
		state.highestSendTime = sendTime
	}

	// Apply the sliding txid window. sendTime may have advanced
	// (above) without affecting the bitmap — every state change
	// from here on is txid-driven.
	if txid > state.highestTxid {
		advance := txid - state.highestTxid
		if advance >= relayReplayWindowSize {
			// Whole window slides out — the new txid is so far
			// ahead that everything we knew is now stale. Clear
			// and reseed at this txid.
			for i := range state.bitmap {
				state.bitmap[i] = 0
			}
		} else {
			slideBitmapForward(&state.bitmap, advance)
		}
		state.highestTxid = txid
		state.bitmap[0] |= 1
		return true, replayOutcomeAccepted
	}

	// txid <= highestTxid. Window check.
	if state.highestTxid-txid >= relayReplayWindowSize {
		// Below the trailing W. The relay's monotonic counter
		// has already advanced past this txid's window.
		// validatePeer's staleness floor catches most such
		// envelopes upstream; this branch handles the in-floor
		// but out-of-window case (a captured envelope still
		// fresh enough to pass staleness but too low for the
		// window).
		return false, replayOutcomeOutsideWindow
	}

	pos := state.highestTxid - txid
	word := pos / 64
	bit := uint64(1) << (pos % 64)
	if state.bitmap[word]&bit != 0 {
		return false, replayOutcomeDuplicate
	}
	state.bitmap[word] |= bit
	return true, replayOutcomeAccepted
}

// slideBitmapForward shifts the bitmap rightward in txid space by
// `shift` bits: new bit position i represents the txid it
// previously represented at position (i - shift). Bits that fall
// off the high end (where the oldest record now exceeds W old
// positions from the new anchor) are dropped. After this call the
// caller still needs to set bit 0 (the new anchor's slot).
//
// shift == 0 is a no-op; shift >= relayReplayWindowSize clears the
// bitmap (the OLD anchor's window has no overlap with the new
// anchor's window).
func slideBitmapForward(b *[relayReplayBitmapWords]uint64, shift uint64) {
	if shift == 0 {
		return
	}
	if shift >= relayReplayWindowSize {
		for i := range b {
			b[i] = 0
		}
		return
	}
	sw := uint(shift / 64)
	sb := shift % 64

	var tmp [relayReplayBitmapWords]uint64
	if sb == 0 {
		// Word-aligned copy. tmp[sw:] = b[:len-sw], tmp[:sw] = 0.
		for j := sw; j < relayReplayBitmapWords; j++ {
			tmp[j] = b[j-sw]
		}
	} else {
		// Bit-shift: bit at old linear position p lands at new
		// linear position p + shift. The shift splits into
		// 64*sw (whole words) and sb (intra-word), so old word
		// (j-sw)'s bit t moves to new position 64j + t + sb.
		//
		// In the new word j:
		//   - bits contributed by old word (j-sw) at bit
		//     positions [sb, 64): b[j-sw] << sb
		//   - bits contributed by old word (j-sw-1) at bit
		//     positions [0, sb): b[j-sw-1] >> (64 - sb)
		// The carry is zero when j == sw (no lower word exists).
		for j := sw; j < relayReplayBitmapWords; j++ {
			lo := b[j-sw] << sb
			var hi uint64
			if j > sw {
				hi = b[j-sw-1] >> (64 - sb)
			}
			tmp[j] = lo | hi
		}
	}
	for i := range b {
		b[i] = tmp[i]
	}
}
