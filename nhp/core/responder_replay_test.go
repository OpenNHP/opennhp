package core

import "testing"

func TestARTReplayAndFloodPredicates(t *testing.T) {
	if !shouldCheckRecvAttack(NHP_SERVER, NHP_AC, NHP_ART) {
		t.Fatal("ART must use the per-connection replay gate")
	}
	if shouldCheckFlood(NHP_SERVER, NHP_AC, NHP_ART) {
		t.Fatal("ART must remain exempt from the 20ms flood gate")
	}
	if shouldEscalateReplay(NHP_SERVER, NHP_AC, NHP_ART) {
		t.Fatal("ART replay must be drop-only")
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
		!shouldCheckFlood(NHP_SERVER, NHP_AGENT, NHP_KNK) ||
		!shouldEscalateReplay(NHP_SERVER, NHP_AGENT, NHP_KNK) {
		t.Fatal("default packet gates must remain enabled")
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
