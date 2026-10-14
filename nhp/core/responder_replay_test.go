package core

import (
	"sync/atomic"
	"testing"
	"time"
)

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

// TestRelayStaleEnvelopeDoesNotBlockRelay exercises the registered-relay
// staleness branch added on top of PR 1778 (GHSA-674q-j8fh-972c):
// the relay's UDP conn is exempted from the monotonic / 20ms flood
// gates for NHP_RLY, but two stale envelopes (older than
// DefaultRecvStalenessFloorSeconds) were still enough to drive
// RecvThreatCount past ThreatCountBeforeBlock (=1) and call
// SendBlockSignal — the connection routine then block-lists the
// RELAY's UDP address, dropping every other agent behind the relay.
//
// The fix: when peerDeviceType == NHP_RELAY, the staleness floor
// returns ErrStalePacketReceived WITHOUT touching RecvThreatCount or
// sending BlockSignal. The relay is authenticated, so this is not a
// vector for unauthenticated abuse — silently drop and keep the relay
// reachable. The per-inner-client replay/flood check inside
// HandleRelayForward still fires on the decoded inner packet because
// the relay-forward conn is keyed on the real client, not the relay.
//
// The test uses the same Noise handshake machinery as the rest of the
// responder tests, with a HandsetRequest from a device typed as
// NHP_RELAY and a stale initTime. It must:
//   - return ErrStalePacketReceived (the relay's RLY was indeed stale)
//   - leave RecvThreatCount at zero (no block)
//   - leave BlockSignal unwritten (the connection routine would never
//     observe a block request)
func TestRelayStaleEnvelopeDoesNotBlockRelay(t *testing.T) {
	sender := NewDevice(NHP_RELAY, sentinelPrivateKey(2), nil)
	receiver := NewDevice(NHP_SERVER, sentinelPrivateKey(34), nil)
	if sender == nil || receiver == nil {
		t.Fatal("failed to create relay/server devices")
	}
	t.Cleanup(sender.Stop)
	t.Cleanup(receiver.Stop)

	receiverPeer := &UdpPeer{
		PubKeyBase64: receiver.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         12347,
		Type:         NHP_SERVER,
	}
	sender.AddPeer(receiverPeer)
	senderPeer := &UdpPeer{
		PubKeyBase64: sender.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         12346,
		Type:         NHP_RELAY,
	}
	receiver.AddPeer(senderPeer)

	senderConn := sentinelConnection(sender, 12346, 12347)
	mad, err := sender.MsgToPacket(&MsgData{
		ConnData:      senderConn,
		PeerPk:        receiverPeer.PublicKey(),
		HeaderType:    NHP_RLY,
		TransactionId: 1,
	})
	if err != nil {
		t.Fatalf("MsgToPacket failed: %v", err)
	}

	receiverConn := sentinelConnection(receiver, 12347, 12346)
	// Make the packet stale: initTime is 601 seconds in the future
	// relative to sendTime (the remote timestamp the sender embedded).
	staleInit := mad.LocalInitTime + int64(601*time.Second)
	packet := &Packet{Content: append([]byte(nil), mad.BasePacket.Content...), HeaderType: NHP_RLY}
	ppd, err := receiver.createPacketParserData(&PacketData{
		BasePacket: packet,
		ConnData:   receiverConn,
		InitTime:   staleInit,
	})
	if err != nil {
		t.Fatalf("createPacketParserData failed: %v", err)
	}
	defer ppd.Destroy()

	if err := ppd.validatePeer(); err != ErrStalePacketReceived {
		t.Fatalf("validatePeer error = %v, want ErrStalePacketReceived", err)
	}

	// The block budget on the relay conn must NOT have been touched.
	// Pre-fix this would have incremented RecvThreatCount to 1 (one
	// short of ThreatCountBeforeBlock+1 but the route threshold is 1,
	// meaning one violation can in principle already fire if the
	// comparator is > rather than >=; the implementation here uses >,
	// so a single stale envelope alone does not block — but two would,
	// and the count is what's load-bearing for the regression).
	if got := atomicLoadInt32(&receiverConn.RecvThreatCount); got != 0 {
		t.Fatalf("registered-relay staleness incremented RecvThreatCount = %d, want 0 "+
			"(the relay must remain reachable so subsequent legitimate envelopes are not block-listed)", got)
	}

	// BlockSignal must remain empty (non-blocking select on the channel
	// would return immediately with zero values). If a stale envelope
	// had pushed a token, the connection routine would tear down and
	// AddBlockAddr the relay's RemoteAddr, taking out every client
	// behind the relay.
	select {
	case <-receiverConn.BlockSignal:
		t.Fatal("registered-relay staleness sent BlockSignal; the relay's per-connection block budget must stay clean")
	default:
	}
}

// atomicLoadInt32 is a tiny local helper for the staleness test —
// ConnectionData.RecvThreatCount is sync/atomic.Int32's underlying
// field but is exposed as a plain int32, so the race detector still
// needs an atomic load to inspect it from the test goroutine. Keeping
// it local avoids an otherwise-unused import in the rest of this file.
func atomicLoadInt32(p *int32) int32 {
	return atomic.LoadInt32(p)
}

// TestRLYFutureTimestampBound pins the second half of the ART skew
// check after it was generalized to any (deviceType, peerType, msgType)
// tuple whose replay is cached via dedupeRecvART/dedupeRecvRLY. Without
// this bound, an NHP_RLY envelope stamped ahead of local time by more
// than ARTRecvFutureSkewSeconds can outlive its cache entry: the
// staleness floor only rejects envelopes BEHIND local time, so a
// forward-stamped replay sent after the entry expires lands on a
// fresh conn, misses the cache, and goes through HandleRelayForward
// again. Before this generalization the only line of defense was the
// per-connection monotonic gate — and the prior PR exempted RLY from
// that gate.
//
// The fixture mirrors TestARTFutureTimestampBound: pull a real NHP_RLY
// packet off a NHP_RELAY↔NHP_SERVER Noise handshake, then drive the
// receiver's initTime so the remote timestamp appears to be ahead by
// (a) exactly the limit (accepted), (b) one nanosecond past it
// (rejected), and (c) far past it (rejected).
func TestRLYFutureTimestampBound(t *testing.T) {
	sender := NewDevice(NHP_RELAY, sentinelPrivateKey(3), nil)
	receiver := NewDevice(NHP_SERVER, sentinelPrivateKey(35), nil)
	if sender == nil || receiver == nil {
		t.Fatal("failed to create relay/server devices")
	}
	t.Cleanup(sender.Stop)
	t.Cleanup(receiver.Stop)

	receiverPeer := &UdpPeer{
		PubKeyBase64: receiver.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         12349,
		Type:         NHP_SERVER,
	}
	sender.AddPeer(receiverPeer)
	senderPeer := &UdpPeer{
		PubKeyBase64: sender.PublicKeyBase64(),
		Ip:           "127.0.0.1",
		Port:         12348,
		Type:         NHP_RELAY,
	}
	receiver.AddPeer(senderPeer)

	senderConn := sentinelConnection(sender, 12348, 12349)
	mad, err := sender.MsgToPacket(&MsgData{
		ConnData:      senderConn,
		PeerPk:        receiverPeer.PublicKey(),
		HeaderType:    NHP_RLY,
		TransactionId: 1,
	})
	if err != nil {
		t.Fatalf("MsgToPacket failed: %v", err)
	}

	for _, delta := range []time.Duration{
		ARTRecvFutureSkewSeconds * time.Second,
		ARTRecvFutureSkewSeconds*time.Second + 1,
		10 * time.Minute,
	} {
		// initTime is the receiver's notion of "now". Driving it
		// *behind* the sender's RemoteSendTime by `delta` makes the
		// envelope look forward-stamped by delta.
		receiverConn := sentinelConnection(receiver, 12349, 12348)
		initTimeBehind := mad.LocalInitTime - int64(delta)
		packet := &Packet{Content: append([]byte(nil), mad.BasePacket.Content...), HeaderType: NHP_RLY}
		ppd, err := receiver.createPacketParserData(&PacketData{
			BasePacket: packet,
			ConnData:   receiverConn,
			InitTime:   initTimeBehind,
		})
		if err != nil {
			t.Fatal(err)
		}
		got := ppd.validatePeer()
		ppd.Destroy()
		if delta > ARTRecvFutureSkewSeconds*time.Second && got != ErrStalePacketReceived {
			t.Fatalf("RLY skew %v accepted: got %v, want ErrStalePacketReceived", delta, got)
		}
		if delta == ARTRecvFutureSkewSeconds*time.Second && got != nil {
			t.Fatalf("RLY allowed skew rejected: got %v", got)
		}
	}
}

// TestFutureSkewCachedPredicate pins the predicate that drives the
// generalized skew check. The check exists for every tuple that
// matches a post-validation replay cache (NHP_ART and NHP_RLY today);
// future cache types must extend this predicate or they inherit the
// staleness-only check and the original replay hole.
func TestFutureSkewCachedPredicate(t *testing.T) {
	cases := []struct {
		name              string
		device, peer, msg int
		want              bool
	}{
		{"ART cached", NHP_SERVER, NHP_AC, NHP_ART, true},
		{"RLY cached", NHP_SERVER, NHP_RELAY, NHP_RLY, true},
		{"AOP not cached", NHP_AC, NHP_SERVER, NHP_AOP, false},
		{"KNK not cached", NHP_SERVER, NHP_AGENT, NHP_KNK, false},
		{"ART on wrong receiver", NHP_AC, NHP_SERVER, NHP_AOP, false},
	}
	for _, c := range cases {
		if got := futureSkewCached(c.device, c.peer, c.msg); got != c.want {
			t.Fatalf("%s: futureSkewCached(%d,%d,%d)=%v, want %v", c.name, c.device, c.peer, c.msg, got, c.want)
		}
	}
}
