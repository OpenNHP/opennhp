package server

import (
	"sync"
	"time"

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

// relayReplayEpochSlackNanos bounds the allowed backward gap between
// an incoming envelope's sendTime and the per-relay epoch anchor
// (rlyWindow.epochSendTime) without that envelope being rejected.
// sendTime is per-packet (each envelope goes through
// nhp/core/initiator.go's createMsgAssemblerData, which sets
// mad.LocalInitTime = time.Now().UnixNano() on every call), so the
// 1 s slack is for the realistic jitter on a busy host: a relay
// can have two enqueued envelopes at almost-but-not-quite the
// same instant, and an attacker could try to time a replay to
// land in the gap between them. 1 s is generous — typical UDP
// enqueue jitter is microseconds — but a tight bound would turn
// legitimate load spikes into false rejections, and the attack
// surface is bounded by the signature on sendTime anyway.
//
// IMPORTANT: this slack is only applied by the top-level epoch
// floor that gates the in-window path. The bitmap-mutating branches
// (forward-advance, restart-reset) use a STRICT < check against
// state.epochSendTime — no slack — so a pre-restart capture whose
// sendTime is anywhere strictly below the current epoch cannot
// reach the bitmap-clearing code. See MarkSeen for the branch tree.
const relayReplayEpochSlackNanos = int64(time.Second)

// relayReplayIdleTimeout bounds how long a relay's per-relay
// state survives without any accepted envelope. After this
// elapsed wall-clock time, the next MarkSeen call drops the
// entry and starts fresh from the incoming envelope. The figure
// is the same shape as artReplayCacheTTL — every envelope older
// than the staleness floor (DefaultRecvStalenessFloorSeconds) is
// rejected upstream by validatePeer anyway, plus the future-skew
// bound (ARTRecvFutureSkewSeconds) so a fast-clock relay can age
// out instead of holding the entry open, plus a 60 s safety
// margin for the time it takes validatePeer's clocks to converge
// after a step.
//
// This is the recovery path the review identified: a relay whose
// wall clock steps back into the past (NTP correction, RTC
// reset, dual relay processes sharing one key) would otherwise
// push the epoch forward and then get locked out forever, since
// nothing in the bitmap or the epoch floor clears state on
// wall-clock movement. Without this timeout, a clock-step-back
// outage required an nhp-server restart. With it, the outage is
// bounded by the timeout (about 16 minutes worst case) and the
// future-skew gate's 300 s upper bound on epoch-push.
const relayReplayIdleTimeout = (core.DefaultRecvStalenessFloorSeconds + core.ARTRecvFutureSkewSeconds + 60) * time.Second

// rlyWindow is the per-relay anti-replay state. It tracks the most
// recent W SenderTrxIds the relay has produced (highestTxid + a
// sliding bitmap), the latest RemoteSendTime it has recorded (so
// forward-advance and restart-reset have a single monotonic anchor
// to compare against), and the epochSendTime floor that pre-restart
// captures are rejected against.
//
// INVARIANT: the bitmap tracks [highestTxid - W + 1, highestTxid],
// where bit i represents txid = (highestTxid - i). Bit 0 always
// represents highestTxid itself. The bitmap is shifted "rightward in
// txid space" as highestTxid advances, so it is always the
// "trailing W" of the relay's monotonic counter.
//
// EPOCH ANCHOR. epochSendTime is the lower bound on sendTime that
// any accepted envelope must clear. It is set on the first envelope
// from this relay and re-set on every restart reset, so once the
// relay has moved on, no pre-restart capture can land in the window
// no matter what txid it carries. The relay's counter is reseeded
// at random per process (nhp/core/device.go:142), so about half of
// all relay restarts land below the old highestTxid and would
// otherwise take the forward-advance branch on a pre-restart
// capture — clearing the bitmap and accepting the replay. The
// epoch floor rejects those envelopes before they reach the
// advance branch.
//
// The epoch anchor is the load-bearing floor for the bitmap-
// mutating branches (forward-advance, restart-reset), which use a
// STRICT `sendTime < epochSendTime` check. The top-level floor
// above the in-window path uses the same epoch but with the
// per-packet jitter slack (relayReplayEpochSlackNanos) — its job
// is to keep the in-window case tolerant of UDP reordering, not
// to gate the bitmap-mutating branches. Without the strict check
// on the advance/reset branches, a 1 s window of opportunity
// remains: a pre-restart capture at sendTime S can pass the slack-
// bounded top-level floor when a restart envelope at sendTime
// S+Δ (Δ < 1 s) advances the epoch, and then clear the bitmap
// through the advance branch.
//
// why this invariant is load-bearing:
//   - the Window check (highestTxid - txid < W) is what bounds how
//     long a captured envelope can be re-injected before it's
//     considered stale. validatePeer's staleness floor bounds the
//     envelope's RECV window; this bounds the txid replay window.
//   - anchoring on highestSendTime (separately from highestTxid)
//     lets a relay process restart re-enter the dedupe: the new
//     process is sending in a fresh sendTime range (each envelope
//     carries time.Now().UnixNano() at the relay's
//     createMsgAssemblerData call site), and we accept envelopes
//     with sendTime > highestSendTime AND a txid that has gone
//     backwards past the window as legitimate by resetting the
//     per-relay state. The relay's signature covers sendTime, so
//     this reset is unforgeable: only the relay's own static key
//     can produce an envelope with a newer sendTime.
//   - the epochSendTime floor is what makes the txid-driven
//     branches safe across a restart. Without it, a pre-restart
//     capture whose txid is larger than the new epoch's anchor
//     would take the advance branch (advance >= W) and reseed the
//     window at the capture's txid — letting an on-path attacker
//     re-inject every captured envelope at the rate of one
//     legitimate envelope. The floor rejects the capture first
//     because its sendTime is below the new epoch's anchor.
//
// IDLE EXPIRATION. lastSeenAt is the wall-clock time of the most
// recent accepted envelope for this relay. After
// relayReplayIdleTimeout of no accepted traffic, MarkSeen drops
// the entry and starts fresh — the recovery path the epoch floor
// alone does not provide. It is updated only on the accept paths
// (first observation, restart reset, normal advance, in-window
// distinct); rejected envelopes do not reset the timer, so a
// replay flood cannot prevent idle expiration.
type rlyWindow struct {
	highestTxid     uint64
	highestSendTime int64
	epochSendTime   int64
	lastSeenAt      time.Time
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
//
// now is the wall-clock source; tests can inject a fixed clock to
// drive the idle-expiration path deterministically. Production code
// keeps time.Now.
type relayReplayWindow struct {
	mu      sync.Mutex
	windows map[[core.PublicKeySizeEx]byte]*rlyWindow
	now     func() time.Time
}

func newRelayReplayWindow() *relayReplayWindow {
	return &relayReplayWindow{
		windows: make(map[[core.PublicKeySizeEx]byte]*rlyWindow),
		now:     time.Now,
	}
}

// newRelayReplayWindowWithClock is the test-only constructor. The
// production constructor above keeps the time.Now call site
// obvious; tests that drive the idle-expiration path swap in a
// fixed clock so they don't have to wait the full 16-minute
// timeout to see the recovery path fire.
func newRelayReplayWindowWithClock(now func() time.Time) *relayReplayWindow {
	w := newRelayReplayWindow()
	w.now = now
	return w
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

	now := w.now()
	state, ok := w.windows[key]
	if ok && now.Sub(state.lastSeenAt) >= relayReplayIdleTimeout {
		// Idle expiration. The relay has been silent long enough
		// that no envelope older than the staleness floor could
		// still be in flight, so it is safe to drop the entry and
		// start fresh from the incoming envelope. This is the
		// recovery path the epoch floor alone does not provide:
		// a relay whose wall clock has stepped back into the past
		// (NTP correction, RTC reset, dual relay processes
		// sharing one key) would otherwise push the epoch forward
		// on every legitimate envelope and then get locked out
		// forever, since nothing in the bitmap clears state on
		// wall-clock movement. Once the relay is silent for
		// relayReplayIdleTimeout, the next envelope is anchored
		// here as if it were a brand-new relay.
		delete(w.windows, key)
		ok = false
	}
	if !ok {
		// First envelope from this relay (or first envelope after
		// idle expiration above). Anchor the window here.
		state = &rlyWindow{
			highestTxid:     txid,
			highestSendTime: sendTime,
			epochSendTime:   sendTime,
			lastSeenAt:      now,
		}
		state.bitmap[0] = 1
		w.windows[key] = state
		return true, replayOutcomeAccepted
	}

	// Epoch floor (top-level). The relay's counter is reseeded
	// at random per process, so a pre-restart capture can carry
	// a txid that lands above the new epoch's anchor and would
	// otherwise take the forward-advance branch below (advancing
	// the bitmap past the new epoch and accepting the replay).
	//
	// This top-level floor covers the in-window path only. The
	// 1 s slack (relayReplayEpochSlackNanos) is for per-packet
	// UDP reordering: a relay can have two enqueued envelopes at
	// almost-but-not-quite the same instant, and an attacker
	// could try to time a replay to land in the gap. The slack
	// does NOT belong on the bitmap-mutating branches below —
	// a pre-restart capture at sendTime S can pass this floor
	// when a restart envelope at sendTime S+Δ (Δ < 1 s) advances
	// the epoch, so the forward-advance and restart-reset
	// branches run their own STRICT check (no slack) below. The
	// signature on sendTime means an attacker cannot push a
	// captured envelope above the floor by re-signing it, and
	// they cannot forge a fresh sendTime on a different capture
	// without the relay's static key.
	if sendTime < state.epochSendTime-relayReplayEpochSlackNanos {
		return false, replayOutcomeOutsideWindow
	}

	// Restart detection. A relay process restart produces
	// envelopes in a fresh sendTime range (the new process
	// stamps time.Now().UnixNano() at each call site, see
	// nhp/core/initiator.go:149, so its envelopes are typically
	// newer than the old process's last recorded sendTime) AND
	// a reseeded txid counter (Device.NextCounterIndex seeds at
	// random per new Device, nhp/core/device.go:142). The reset
	// below is what lets a low-counter envelope from the
	// restarted relay land in the window — without it, every
	// legitimate post-restart envelope would be OutsideWindow
	// until the relay sent W envelopes to push the anchor past
	// it.
	//
	// The reset is unforgeable in practice (sendTime is inside
	// the signed payload, so only the relay's static key can
	// produce a higher sendTime), but we still gate it narrowly:
	// the txid must also have gone backwards past the window.
	// Without that back-of-window check, a stale captured
	// envelope whose sendTime happens to exceed the current
	// anchor would clear the bitmap — the signature still binds
	// sendTime so a captured envelope cannot itself bump it, but
	// tying the reset to a txid that is provably unrecoverable
	// from the bitmap keeps the rule local: we only ever clear
	// state we have already decided is unrecoverable, never
	// state we could still catch a duplicate in.
	//
	// STRICT epoch check. The header above (`sendTime > state.highestSendTime`)
	// already prevents a pre-restart capture (whose sendTime is <= the
	// original first-envelope sendTime = epoch) from entering this branch
	// when no other envelopes have advanced `highestSendTime` past `epoch`.
	// But `highestSendTime` IS moved forward by every forward-advance
	// envelope, while `epoch` is NOT. So after K post-restart envelopes,
	// `highestSendTime = S_restart + K·Δ` while `epoch = S_restart`. A
	// subsequent legitimate restart envelope can still pass the
	// `sendTime > highestSendTime` gate (because `S_restart2 > S_restart + K·Δ`),
	// and this strict check makes the invariant "this envelope is at
	// least as new as the first one we ever saw from this relay"
	// explicit and locally enforced. No slack here: a pre-restart
	// capture that is strictly below the current epoch must not be
	// allowed to clear the bitmap, regardless of how close it is to
	// the anchor.
	//
	// The reset also advances epochSendTime to the new envelope's
	// sendTime, so pre-restart captures can no longer reach the
	// bitmap-advance path through any later envelope — the
	// strict epoch check now rejects them at the top of this branch.
	if sendTime > state.highestSendTime &&
		txid < state.highestTxid &&
		state.highestTxid-txid >= relayReplayWindowSize {
		if sendTime < state.epochSendTime {
			return false, replayOutcomeOutsideWindow
		}
		for i := range state.bitmap {
			state.bitmap[i] = 0
		}
		state.highestTxid = txid
		state.highestSendTime = sendTime
		state.epochSendTime = sendTime
		state.lastSeenAt = now
		state.bitmap[0] = 1
		return true, replayOutcomeAccepted
	}

	// Track the latest signed sendTime. The signature binds
	// sendTime, so taking max() here is just keeping the anchor
	// in sync with the latest legitimate envelope; a stale
	// captured envelope has the old sendTime, which is <= the
	// current anchor and does not change anything. This branch
	// never touches the bitmap. epochSendTime stays put here —
	// the only paths that move the epoch forward are the first
	// envelope from this relay and the explicit restart reset
	// above.
	if sendTime > state.highestSendTime {
		state.highestSendTime = sendTime
	}

	// Apply the sliding txid window. sendTime may have advanced
	// (above) without affecting the bitmap — every state change
	// from here on is txid-driven.
	if txid > state.highestTxid {
		// STRICT epoch check. A pre-restart capture whose txid
		// happens to advance the new epoch's anchor (the relay's
		// counter is reseeded at random, so ~half of all restarts
		// land the new counter below the old highestTxid while a
		// captured pre-restart envelope has a HIGHER txid) must
		// not be allowed to clear the bitmap. The top-level epoch
		// floor above has a 1 s slack for in-window reordering; a
		// restart envelope at sendTime S+Δ (Δ < 1 s) advances the
		// epoch past the pre-restart capture at sendTime S, so the
		// top-level floor accepts the replay — and without this
		// strict check the advance branch would clear the bitmap.
		// No slack here: a pre-restart capture that is strictly
		// below the current epoch must not advance the window.
		if sendTime < state.epochSendTime {
			return false, replayOutcomeOutsideWindow
		}
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
		state.lastSeenAt = now
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
	state.lastSeenAt = now
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
