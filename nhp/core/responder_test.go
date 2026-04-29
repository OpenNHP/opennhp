package core

import "testing"

// TestShouldCheckRecvAttack_AOPNoLongerExempt is the regression
// fence for issue #1123. The previous implementation skipped the
// per-connection LastRemoteSendTime monotonic check for NHP_AOP,
// which made the responder a no-op replay-protector for AC-side
// AOP processing within an open connection. The fix removes that
// exemption so AOP rides the same gate as every other AC-bound
// message; the cross-connection cousin lives in
// endpoints/ac/aop_replay_cache.go.
func TestShouldCheckRecvAttack_AOPNoLongerExempt(t *testing.T) {
	if !shouldCheckRecvAttack(NHP_AC, NHP_SERVER, NHP_AOP) {
		t.Fatal("NHP_AOP on AC must be subject to the LastRemoteSendTime gate (#1123)")
	}
}

// TestShouldCheckRecvAttack_ARTStillExempt pins the remaining
// exemption: NHP_ART (AC → server response) skips the gate because
// the server-side transaction layer already correlates by
// TransactionId and the AC→server hop occasionally exceeds the
// flood-gate threshold (MinimalRecvIntervalMs in constants.go).
// Tracked for follow-up dedupe in #1457.
func TestShouldCheckRecvAttack_ARTStillExempt(t *testing.T) {
	if shouldCheckRecvAttack(NHP_SERVER, NHP_AC, NHP_ART) {
		t.Fatal("NHP_ART on server must remain exempt from the LastRemoteSendTime gate")
	}
}

// TestShouldCheckRecvAttack_DefaultEnforced sanity-checks that an
// unrelated message type still enforces the gate, so a future
// refactor of shouldCheckRecvAttack cannot silently widen the
// exemption set.
func TestShouldCheckRecvAttack_DefaultEnforced(t *testing.T) {
	if !shouldCheckRecvAttack(NHP_SERVER, NHP_AGENT, NHP_KNK) {
		t.Fatal("non-exempt (deviceType, peerType, msgType) must enforce the gate")
	}
}

// TestShouldCheckFlood_AOPExempt fences the round-7 cr finding for
// issue #1123. The flood gate (`MinimalRecvIntervalMs = 20 ms`)
// must NOT apply to NHP_AOP on AC: the server legitimately emits
// AOPs in tight succession during knock bursts, and applying the
// 20 ms floor would false-flood-block a connection past
// `ThreatCountBeforeBlock`. Replay protection is preserved via
// shouldCheckRecvAttack (still enforced on AOP) plus the
// cross-connection cache in endpoints/ac/aop_replay_cache.go.
func TestShouldCheckFlood_AOPExempt(t *testing.T) {
	if shouldCheckFlood(NHP_AC, NHP_SERVER, NHP_AOP) {
		t.Fatal("NHP_AOP on AC must be exempt from the 20 ms flood gate (#1123 round-7)")
	}
}

// TestShouldCheckFlood_ARTExempt mirrors the existing replay-gate
// exemption on the flood gate so the AC→server hop's legitimate
// latency does not flood-block a connection.
func TestShouldCheckFlood_ARTExempt(t *testing.T) {
	if shouldCheckFlood(NHP_SERVER, NHP_AC, NHP_ART) {
		t.Fatal("NHP_ART on server must remain exempt from the 20 ms flood gate")
	}
}

// TestShouldCheckFlood_DefaultEnforced asserts the flood gate is
// otherwise enforced, so a future refactor of shouldCheckFlood
// cannot silently widen the exemption set.
func TestShouldCheckFlood_DefaultEnforced(t *testing.T) {
	if !shouldCheckFlood(NHP_SERVER, NHP_AGENT, NHP_KNK) {
		t.Fatal("non-exempt (deviceType, peerType, msgType) must enforce the flood gate")
	}
}
