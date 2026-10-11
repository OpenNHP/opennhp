package core

import "testing"

func TestARTReplayAndFloodPredicates(t *testing.T) {
	if shouldCheckRecvAttack(NHP_SERVER, NHP_AC, NHP_ART) {
		t.Fatal("ART must remain exempt from the monotonic timestamp gate so reordered bursts are accepted")
	}
	if shouldCheckFlood(NHP_SERVER, NHP_AC, NHP_ART) {
		t.Fatal("ART must remain exempt from the 20ms flood gate")
	}
}

func TestAOPAndDefaultPredicatesRemainStable(t *testing.T) {
	if shouldCheckRecvAttack(NHP_AC, NHP_SERVER, NHP_AOP) {
		t.Fatal("independent ART port must not change the AOP replay exemption")
	}
	if shouldCheckFlood(NHP_AC, NHP_SERVER, NHP_AOP) {
		t.Fatal("AOP must remain flood-exempt")
	}
	if !shouldCheckRecvAttack(NHP_SERVER, NHP_AGENT, NHP_KNK) ||
		!shouldCheckFlood(NHP_SERVER, NHP_AGENT, NHP_KNK) {
		t.Fatal("default packet gates must remain enabled")
	}
}

// TestRelayEnvelopeExemptFromConnectionGates pins the fix for
// GHSA-674q-j8fh-972c: the shared, authenticated relay→server connection
// multiplexes many clients, so its outer NHP_RLY envelopes must not be
// subject to the per-connection monotonic/flood gates that would otherwise
// block the whole relay path. Per-client anti-replay/flood is enforced on
// the decoded inner packet instead.
func TestRelayEnvelopeExemptFromConnectionGates(t *testing.T) {
	if shouldCheckRecvAttack(NHP_SERVER, NHP_RELAY, NHP_RLY) {
		t.Fatal("NHP_RLY must be exempt from the monotonic replay gate so concurrent relayed clients don't block the shared path")
	}
	if shouldCheckFlood(NHP_SERVER, NHP_RELAY, NHP_RLY) {
		t.Fatal("NHP_RLY must be exempt from the 20ms flood gate")
	}
	// The exemption is narrow: it must not leak to inner relayed packets
	// (which arrive as KNK/RKN keyed by the real client), nor to a KNK that
	// happens to come from a relay-typed peer by any other path.
	if !shouldCheckRecvAttack(NHP_SERVER, NHP_AGENT, NHP_RKN) ||
		!shouldCheckFlood(NHP_SERVER, NHP_AGENT, NHP_RKN) {
		t.Fatal("inner relayed client packets must keep per-client replay/flood protection")
	}
	if !shouldCheckRecvAttack(NHP_SERVER, NHP_RELAY, NHP_KNK) ||
		!shouldCheckFlood(NHP_SERVER, NHP_RELAY, NHP_KNK) {
		t.Fatal("only NHP_RLY is exempt; other message types from a relay peer stay gated")
	}
}

func TestSetRecvReplayDedupe(t *testing.T) {
	d := &Device{}
	called := false
	d.SetRecvReplayDedupe(func(*PacketParserData) error {
		called = true
		return nil
	})
	if d.recvReplayDedupeFn == nil {
		t.Fatal("replay hook was not installed")
	}
	if err := d.recvReplayDedupeFn(&PacketParserData{}); err != nil || !called {
		t.Fatalf("installed replay hook did not run: called=%v err=%v", called, err)
	}
}
