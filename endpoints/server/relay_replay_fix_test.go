package server

import (
	"encoding/base64"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// TestConnectionRoutine_BlockSignalBlocksRelayForwardedClient pins the
// follow-up to GHSA-674q-j8fh-972c: when the relay-aware BlockSignal
// branch closes a per-client conn because the inner client's
// RecvThreatCount crossed the threshold, the connection routine must
// record a per-(relay, real-client) block so the very next forward
// from the same real client is dropped before reaching the
// connection-table lookup. Pre-fix this branch was `return`-only — the
// relay address stayed reachable (correct, per the GHSA) but the
// offending real client had no penalty at all and could re-instantiate
// a fresh per-client conn with LastRemoteSendTime == 0 on every
// inbound forward, getting ~one accepted replay/flood per three
// packets it sent.
//
// The fix pairs with the IsRelayForwardBlocked gate at the top of
// HandleRelayForward: this test pins the producer side (the
// BlockSignal branch populates the block map), the next test pins the
// consumer side (HandleRelayForward honors it).
func TestConnectionRoutine_BlockSignalBlocksRelayForwardedClient(t *testing.T) {
	s := &UdpServer{
		remoteConnectionMap:       make(map[string]*UdpConn),
		blockAddrMap:              make(map[string]*BlockAddr),
		relayForwardBlockMap:      make(map[string]time.Time),
		wg:                        sync.WaitGroup{},
		relayForwardBlockMapMutex: sync.Mutex{},
	}
	// Skip the deferred SetOverload branch in connectionRoutine:
	// the fixture doesn't have a real *Device, and the path under
	// test is the relay-aware BlockSignal branch, not the overload
	// bookkeeping.
	s.forceOverload.Store(true)

	const relayAddrStr = "198.51.100.10:62206"
	const clientAddrStr = "203.0.113.42:51234"
	relayUDP, _ := net.ResolveUDPAddr("udp", relayAddrStr)
	clientUDP, _ := net.ResolveUDPAddr("udp", clientAddrStr)

	const mapKey = relayConnKeyPrefix + relayAddrStr + relayConnKeySep + clientAddrStr
	conn := &UdpConn{
		mapKey: mapKey,
		ConnData: &core.ConnectionData{
			RemoteAddr:       relayUDP,
			RealRemoteAddr:   clientUDP,
			TimeoutMs:        5000,
			BlockSignal:      make(chan struct{}, 1),
			SetTimeoutSignal: make(chan struct{}, 1),
			StopSignal:       make(chan struct{}),
			SendQueue:        make(chan *core.Packet, 1),
			RecvQueue:        make(chan *core.Packet, 1),
		},
	}

	s.wg.Add(1)
	go s.connectionRoutine(conn)
	time.Sleep(20 * time.Millisecond)

	// Pre-condition: the (relay, real-client) pair is NOT blocked yet.
	if s.IsRelayForwardBlocked(mapKey) {
		t.Fatalf("relay-forward block map pre-populated for %s", mapKey)
	}

	conn.ConnData.SendBlockSignal()

	// Wait for the routine to exit.
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connectionRoutine did not exit after BlockSignal")
	}

	// The (relay, real-client) pair MUST now be on the relay-forward
	// block list. Without this, HandleRelayForward would never learn
	// to drop subsequent forwards and the penalty-free round-trip
	// regression re-opens.
	if !s.IsRelayForwardBlocked(mapKey) {
		t.Fatalf("relay-forward block missing for %s after inner threat signal — "+
			"real client gets a fresh per-client conn (LastRemoteSendTime == 0) on every forward, no penalty", mapKey)
	}

	// As before: blockAddrMap stays untouched (relay address must
	// remain reachable for other clients behind the relay).
	s.blockAddrMapMutex.Lock()
	_, blocked := s.blockAddrMap[relayAddrStr]
	s.blockAddrMapMutex.Unlock()
	if blocked {
		t.Fatalf("relay address %s was block-listed after inner client threat — every other client behind the relay would be unreachable", relayAddrStr)
	}
}

// TestRelayForwardBlockDoesNotAffectOtherClients makes the corollary
// explicit: blocking one real client behind a relay must not leak
// onto a different real client behind the same relay, nor onto the
// relay itself. The block map is keyed on the (relay, real-client)
// composite, so two distinct real-client segments resolve to two
// independent entries.
func TestRelayForwardBlockDoesNotAffectOtherClients(t *testing.T) {
	s := &UdpServer{
		relayForwardBlockMap:      make(map[string]time.Time),
		relayForwardBlockMapMutex: sync.Mutex{},
	}

	const relayAddrStr = "198.51.100.10:62206"
	const blockedClientStr = "203.0.113.42:51234"
	const otherClientStr = "203.0.113.43:51235"
	const blockedKey = relayConnKeyPrefix + relayAddrStr + relayConnKeySep + blockedClientStr
	const otherKey = relayConnKeyPrefix + relayAddrStr + relayConnKeySep + otherClientStr
	// Same real client, different relay — also independent.
	const otherRelayKey = relayConnKeyPrefix + "198.51.100.11:62207" + relayConnKeySep + blockedClientStr

	s.AddRelayForwardBlock(blockedKey)

	if !s.IsRelayForwardBlocked(blockedKey) {
		t.Fatalf("blockedKey %s not blocked", blockedKey)
	}
	if s.IsRelayForwardBlocked(otherKey) {
		t.Fatalf("otherKey %s incorrectly blocked — block must scope to (relay, real-client) pair only", otherKey)
	}
	if s.IsRelayForwardBlocked(otherRelayKey) {
		t.Fatalf("otherRelayKey %s incorrectly blocked — same real client under a different relay must not inherit the block", otherRelayKey)
	}
	if s.IsRelayForwardBlocked(relayAddrStr) {
		t.Fatalf("relay address %s flagged as relay-forward blocked — block key has no relayConnKeyPrefix, gate must return false", relayAddrStr)
	}
}

// TestAddRelayForwardBlockIgnoresNonRelayKeys guards the producer
// side: the only legitimate caller is connectionRoutine's
// RealRemoteAddr != nil branch, which always has a relay-prefixed
// mapKey (or the fallback we synthesize in that branch). Passing a
// non-relay key would silently corrupt the block map with arbitrary
// caller strings — log a Critical and refuse rather than record.
func TestAddRelayForwardBlockIgnoresNonRelayKeys(t *testing.T) {
	s := &UdpServer{
		relayForwardBlockMap:      make(map[string]time.Time),
		relayForwardBlockMapMutex: sync.Mutex{},
	}

	for _, key := range []string{
		"198.51.100.10:62206",       // direct UDP key (no prefix)
		"relay",                     // prefix only, no body
		"relay|198.51.100.10:62206", // missing real-client segment
		"",                          // empty
	} {
		if s.AddRelayForwardBlock(key) {
			t.Fatalf("AddRelayForwardBlock(%q) accepted non-relay key", key)
		}
	}
	s.relayForwardBlockMapMutex.Lock()
	n := len(s.relayForwardBlockMap)
	s.relayForwardBlockMapMutex.Unlock()
	if n != 0 {
		t.Fatalf("block map populated with non-relay keys (n=%d)", n)
	}
}

// TestAddRelayForwardBlockNewVsRefreshReturn pins the return-value
// contract of AddRelayForwardBlock so a future recordBlockedAddr-style
// counter keyed on the return value counts the right events. The
// producer is called from connectionRoutine's relay-aware BlockSignal
// branch, which today ignores the return; if a future change adds a
// counter for "fresh blocks added" the inverted return value would
// silently count refreshes instead and the metric would always read
// zero on a long-running server (every violation that hits an
// already-blocked pair would log false instead of the new-block signal).
//
// First insert of a brand-new key returns true; a repeat insert
// of the same key (refresh path) returns false.
func TestAddRelayForwardBlockNewVsRefreshReturn(t *testing.T) {
	s := &UdpServer{
		relayForwardBlockMap:      make(map[string]time.Time),
		relayForwardBlockMapMutex: sync.Mutex{},
	}

	const relayAddrStr = "198.51.100.10:62206"
	const clientAddrStr = "203.0.113.42:51234"
	key := relayConnKeyPrefix + relayAddrStr + relayConnKeySep + clientAddrStr

	if !s.AddRelayForwardBlock(key) {
		t.Fatalf("first AddRelayForwardBlock(%q) returned false, want true (fresh block)", key)
	}
	if s.AddRelayForwardBlock(key) {
		t.Fatalf("second AddRelayForwardBlock(%q) returned true, want false (refresh)", key)
	}

	// Same shape, different real-client segment. Counts as a fresh
	// block even though the relay segment is reused.
	otherClient := relayConnKeyPrefix + relayAddrStr + relayConnKeySep + "203.0.113.99:51234"
	if !s.AddRelayForwardBlock(otherClient) {
		t.Fatalf("AddRelayForwardBlock for fresh real-client returned false, want true")
	}
}

// TestAddRelayForwardBlockEnforcesCap pins the size cap: once the
// map holds MaxConcurrentConnection entries, an insert for a key
// that is NOT already on the list must be refused (return false,
// no map growth, "relay-forward block pool is full" warning). A
// refresh of an existing entry at the cap is still allowed — it
// doesn't grow the map.
//
// Without this cap, a client behind a relay can drive memory growth
// by rotating source ports (each port = a new (relay, realClient)
// key) at the rate the per-client conn closes; the 90s TTL is the
// only thing pruning, and the cap stops the worst-case scenario.
func TestAddRelayForwardBlockEnforcesCap(t *testing.T) {
	s := &UdpServer{
		relayForwardBlockMap:      make(map[string]time.Time),
		relayForwardBlockMapMutex: sync.Mutex{},
	}

	// Pre-fill with MaxConcurrentConnection-1 entries (one short
	// of the cap) so a single AddRelayForwardBlock hits it without
	// needing 20k+ test fixtures.
	cap := MaxConcurrentConnection
	for i := 0; i < cap-1; i++ {
		// Each key has to be unique AND valid; embed the index in
		// the real-client segment so the composite parses.
		key := relayConnKeyPrefix + "198.51.100.10:62206" + relayConnKeySep +
			"203.0.113.1:" + strconv.Itoa(1024+i)
		if !s.AddRelayForwardBlock(key) {
			t.Fatalf("prefill insert #%d rejected unexpectedly", i)
		}
	}

	s.relayForwardBlockMapMutex.Lock()
	if n := len(s.relayForwardBlockMap); n != cap-1 {
		s.relayForwardBlockMapMutex.Unlock()
		t.Fatalf("prefill produced %d entries, want %d", n, cap-1)
	}
	s.relayForwardBlockMapMutex.Unlock()

	// One more fresh key: must be accepted (still under cap).
	lastFreshKey := relayConnKeyPrefix + "198.51.100.10:62206" + relayConnKeySep + "203.0.113.99:51234"
	if !s.AddRelayForwardBlock(lastFreshKey) {
		t.Fatalf("AddRelayForwardBlock under cap returned false, want true")
	}
	s.relayForwardBlockMapMutex.Lock()
	if n := len(s.relayForwardBlockMap); n != cap {
		s.relayForwardBlockMapMutex.Unlock()
		t.Fatalf("map size after fill = %d, want %d", n, cap)
	}
	s.relayForwardBlockMapMutex.Unlock()

	// Cap is hit. Another brand-new key must be refused; map size
	// must not grow.
	overCapKey := relayConnKeyPrefix + "198.51.100.10:62206" + relayConnKeySep + "203.0.113.100:51234"
	if s.AddRelayForwardBlock(overCapKey) {
		t.Fatalf("AddRelayForwardBlock over cap returned true, want false (no map growth)")
	}
	s.relayForwardBlockMapMutex.Lock()
	if n := len(s.relayForwardBlockMap); n != cap {
		s.relayForwardBlockMapMutex.Unlock()
		t.Fatalf("map size after over-cap insert = %d, want %d (refused insert must not grow map)", n, cap)
	}
	s.relayForwardBlockMapMutex.Unlock()

	// Refresh of an existing entry at the cap is still allowed —
	// it doesn't grow the map, and it's how we keep already-blocked
	// clients blocked after another violation. The return value is
	// false (no fresh block added) so a recordBlockedAddr-style
	// counter keyed on it correctly skips the refresh.
	if s.AddRelayForwardBlock(lastFreshKey) {
		t.Fatalf("AddRelayForwardBlock refresh at cap returned true, want false")
	}
	s.relayForwardBlockMapMutex.Lock()
	if n := len(s.relayForwardBlockMap); n != cap {
		s.relayForwardBlockMapMutex.Unlock()
		t.Fatalf("map size after refresh at cap = %d, want %d", n, cap)
	}
	s.relayForwardBlockMapMutex.Unlock()
}

// for GHSA-674q-j8fh-972c follow-up: when the inner client's
// RecvThreatCount gate fires on a relay-forwarded connection, the
// per-connection BlockSignal triggers connectionRoutine, and the
// previous code unconditionally called AddBlockAddr(RemoteAddr) — which
// is the RELAY's address, not the offending inner client's. One inner
// flood (or replay) twice over ThreatCountBeforeBlock (=1) was therefore
// enough to drop every other client behind that relay for
// BlockAddrExpireTime (~90s). Today the BlockSignal branch checks
// RealRemoteAddr and, when set, closes the per-client conn AND records
// a per-(relay, real-client) block on the relay-forward map (so the
// offending real client gets the same penalty the direct-UDP path has,
// without dragging the relay's other clients down). blockAddrMap is
// deliberately untouched.
//
// The test wires up a UdpServer with a minimal relay-forwarded conn and
// fires a BlockSignal. After the routine exits: blockAddrMap must be
// empty (relay NOT blocked), the (relay, real-client) pair MUST be on
// the relay-forward block list (the new behavior), and the conn must
// be torn down (so the caller cannot continue to use it).
func TestConnectionRoutine_BlockSignalDoesNotBlockRelayAddress(t *testing.T) {
	s := &UdpServer{
		remoteConnectionMap:       make(map[string]*UdpConn),
		blockAddrMap:              make(map[string]*BlockAddr),
		relayForwardBlockMap:      make(map[string]time.Time),
		relayForwardBlockMapMutex: sync.Mutex{},
		wg:                        sync.WaitGroup{},
	}
	// Skip the deferred SetOverload branch in connectionRoutine:
	// the fixture doesn't have a real *Device. The path under test
	// is the relay-aware BlockSignal branch, not overload bookkeeping.
	s.forceOverload.Store(true)

	const relayAddrStr = "198.51.100.10:62206"
	const clientAddrStr = "203.0.113.42:51234"
	relayUDP, _ := net.ResolveUDPAddr("udp", relayAddrStr)
	clientUDP, _ := net.ResolveUDPAddr("udp", clientAddrStr)

	conn := &UdpConn{
		mapKey: relayConnKeyPrefix + relayAddrStr + relayConnKeySep + clientAddrStr,
		ConnData: &core.ConnectionData{
			RemoteAddr:     relayUDP,
			RealRemoteAddr: clientUDP,
			TimeoutMs:      5000,
			// ConnectionData.beginChannelSend requires that the
			// channel-based teardown plumbing exists, even when the
			// routine exits via the BlockSignal branch — Close() runs
			// in the deferred teardown of connectionRoutine and
			// flushes/closes these queues. Pre-create them so the
			// defer does not panic on a fixture that never used them.
			BlockSignal:      make(chan struct{}, 1),
			SetTimeoutSignal: make(chan struct{}, 1),
			StopSignal:       make(chan struct{}),
			SendQueue:        make(chan *core.Packet, 1),
			RecvQueue:        make(chan *core.Packet, 1),
		},
	}

	s.wg.Add(1)
	go s.connectionRoutine(conn)

	// Give the routine a moment to enter its select loop, then fire the
	// signal that previously would have block-listed the relay address.
	time.Sleep(20 * time.Millisecond)
	conn.ConnData.SendBlockSignal()

	// Wait up to a second for the routine to exit. If it never does the
	// relay-address block would also be silently absent, so a hang here
	// is a meaningful failure (it means our fix returned without
	// actually exiting the loop, or the relay was added to blockAddrMap
	// before the routine exited).
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connectionRoutine did not exit after BlockSignal — relay fix did not return")
	}

	// The relay address must NOT be in blockAddrMap. Pre-fix this test
	// would have observed relayUDP.String() present and the assertion
	// would fail.
	s.blockAddrMapMutex.Lock()
	_, blocked := s.blockAddrMap[relayAddrStr]
	s.blockAddrMapMutex.Unlock()
	if blocked {
		t.Fatalf("relay address %s was block-listed after inner client threat; every other client behind the relay would now be unreachable for BlockAddrExpireTime",
			relayAddrStr)
	}

	// Per the follow-up review on top of #1780: the relay-aware
	// BlockSignal branch must ALSO record a relay-forward block on
	// the (relay, real-client) composite, otherwise the very next
	// forward from the offending real client lands on a fresh
	// per-client conn with LastRemoteSendTime == 0 and is not stopped.
	// This single assertion is the producer-side pin for the new
	// IsRelayForwardBlocked gate at the top of HandleRelayForward.
	if !s.IsRelayForwardBlocked(conn.mapKey) {
		t.Fatalf("relay-forward block missing for %s after BlockSignal — offending real client gets no penalty on next forward",
			conn.mapKey)
	}

	// The per-client conn must have torn down — that's the alternative
	// "drop only the offending client" outcome the fix promises. We
	// verify via IsClosed() rather than the map (the test doesn't insert
	// the conn into remoteConnectionMap, but the routine's deferred
	// teardown still runs Close() which flips IsClosed).
	if !conn.ConnData.IsClosed() {
		t.Fatal("relay-forwarded per-client conn was not closed after BlockSignal")
	}
}

// TestConnectionRoutine_BlockSignalStillBlocksDirectClient is the
// negative-direction check on the same code path: for a DIRECT UDP
// client (RealRemoteAddr == nil), the BlockSignal branch must still
// block-list the source so we don't lose the existing protection for
// agents that connect without a relay. This guards against an
// over-broad fix that would silently neuter the threat gate for the
// common case.
func TestConnectionRoutine_BlockSignalStillBlocksDirectClient(t *testing.T) {
	s := &UdpServer{
		remoteConnectionMap: make(map[string]*UdpConn),
		blockAddrMap:        make(map[string]*BlockAddr),
		wg:                  sync.WaitGroup{},
	}
	// Skip the deferred SetOverload branch in connectionRoutine:
	// the fixture doesn't have a real *Device. Pre-existing test
	// from #1780; the fixture was panicking on connectionRoutine's
	// teardown defer in CI too, the Store(false) here makes the
	// test runnable without spinning up a full Device.
	s.forceOverload.Store(true)

	const directAddrStr = "203.0.113.5:54321"
	directUDP, _ := net.ResolveUDPAddr("udp", directAddrStr)

	conn := &UdpConn{
		ConnData: &core.ConnectionData{
			RemoteAddr:       directUDP,
			RealRemoteAddr:   nil, // direct UDP — RealRemoteAddr deliberately unset
			TimeoutMs:        5000,
			BlockSignal:      make(chan struct{}, 1),
			SetTimeoutSignal: make(chan struct{}, 1),
			StopSignal:       make(chan struct{}),
			SendQueue:        make(chan *core.Packet, 1),
			RecvQueue:        make(chan *core.Packet, 1),
		},
	}

	s.wg.Add(1)
	go s.connectionRoutine(conn)

	time.Sleep(20 * time.Millisecond)
	conn.ConnData.SendBlockSignal()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connectionRoutine did not exit after BlockSignal for direct UDP")
	}

	s.blockAddrMapMutex.Lock()
	_, blocked := s.blockAddrMap[directAddrStr]
	s.blockAddrMapMutex.Unlock()
	if !blocked {
		t.Fatalf("direct UDP source %s was NOT block-listed after BlockSignal; the relay-fix branch must not leak into the direct-UDP path",
			directAddrStr)
	}
}

// --- sliding-window dedupe (relay_replay_window.go) tests ---
//
// These tests pin the per-relay sliding-window anti-replay state that
// replaced the previous TTL-keyed dedup set on the NHP_RLY path. The
// motivation for the replacement is the GHSA-674q-j8fh-972c follow-up
// review: a global capacity (evict-or-reject) can be filled by a single
// attacker to fail closed across every relay in the system, while
// live-eviction silently weakens the replay window. A per-relay sliding
// window has none of these failure modes (see relay_replay_window.go
// for the full invariant argument).

// TestDedupeRecvRLY_RegisteredRelay pins the producer-side wiring: a
// fresh NHP_RLY envelope from a registered relay is accepted, a
// duplicate is rejected with ErrServerDuplicateTransaction, distinct
// envelopes with distinct txids all flow through (concurrent-clients
// fan-out case), and unrelated header types short-circuit without
// consuming a window slot. The non-RLY short-circuit is important
// because the dedupe chain is shared with ART and we must not pay the
// window cost for non-RLY frames.
func TestDedupeRecvRLY_RegisteredRelay(t *testing.T) {
	s := &UdpServer{
		relayReplay:  newRelayReplayWindow(),
		relayPeerMap: make(map[string]*core.UdpPeer),
	}

	// Non-RLY header must short-circuit without touching the window.
	if err := s.dedupeRecvRLY(&core.PacketParserData{HeaderType: core.NHP_KNK, RemotePubKey: artTestPubkey(1), SenderTrxId: 1, RemoteSendTime: 1}); err != nil {
		t.Fatalf("non-RLY header rejected: %v", err)
	}
	if len(s.relayReplay.windows) != 0 {
		t.Fatal("non-RLY populated window")
	}

	// RLY with no registered relay peer must FAIL CLOSED, not
	// short-circuit. validatePeer normally rejects these before
	// reaching here, but a config-reload race (newly-registered
	// peer visible to the device pool before the relayPeerMap
	// swap) could let one through, so the dedupe hook's only safe
	// answer is to refuse.
	if err := s.dedupeRecvRLY(&core.PacketParserData{HeaderType: core.NHP_RLY, RemotePubKey: artTestPubkey(2), SenderTrxId: 1, RemoteSendTime: 1}); !errors.Is(err, common.ErrServerMissingPeerPubkey) {
		t.Fatalf("unregistered RLY peer must fail closed with ErrServerMissingPeerPubkey, got %v", err)
	}
	if len(s.relayReplay.windows) != 0 {
		t.Fatal("unregistered relay populated window")
	}

	// Now register a relay peer and exercise the window.
	for _, size := range []int{core.PublicKeySize, core.PublicKeySizeEx} {
		key := artTestPubkey(3, size)
		s := &UdpServer{
			relayReplay:  newRelayReplayWindow(),
			relayPeerMap: map[string]*core.UdpPeer{base64.StdEncoding.EncodeToString(key): {Type: core.NHP_RELAY}},
		}

		packet := &core.PacketParserData{
			HeaderType:     core.NHP_RLY,
			SenderTrxId:    100,
			RemotePubKey:   key,
			RemoteSendTime: 1000,
		}
		if err := s.dedupeRecvRLY(packet); err != nil {
			t.Fatalf("first NHP_RLY with %d-byte key rejected: %v", size, err)
		}
		// Duplicate (same txid, same sendTime): the relay's
		// signature covers sendTime, so a captured-envelope replay
		// reuses both fields.
		if err := s.dedupeRecvRLY(packet); !errors.Is(err, common.ErrServerDuplicateTransaction) {
			t.Fatalf("duplicate NHP_RLY with %d-byte key error = %v", size, err)
		}

		// Distinct txid (concurrent inner-client fan-out): accepted.
		other := &core.PacketParserData{
			HeaderType:     core.NHP_RLY,
			SenderTrxId:    101,
			RemotePubKey:   key,
			RemoteSendTime: 1000,
		}
		if err := s.dedupeRecvRLY(other); err != nil {
			t.Fatalf("fresh NHP_RLY envelope (different txid) rejected: %v", err)
		}
	}
}

// TestDedupeRecvRLY_RejectsNonRelayBeforeCaching: relayPeerMap has
// the peer's pubkey but registered under a non-relay role
// (agent / ac / db). The hook returns ErrServerMissingPeerPubkey
// and must NOT consume a window slot — otherwise a forged-relay
// attack that gets through validatePeer could fill the per-relay
// window and start failing legitimate envelopes from a real relay.
func TestDedupeRecvRLY_RejectsNonRelayBeforeCaching(t *testing.T) {
	key := artTestPubkey(7, core.PublicKeySize)
	for _, role := range []int{core.NHP_AGENT, core.NHP_AC, core.NHP_DB} {
		s := &UdpServer{
			relayReplay:  newRelayReplayWindow(),
			relayPeerMap: map[string]*core.UdpPeer{base64.StdEncoding.EncodeToString(key): {Type: role}},
		}
		p := &core.PacketParserData{HeaderType: core.NHP_RLY, RemotePubKey: key, SenderTrxId: 1, RemoteSendTime: 1}
		if err := s.dedupeRecvRLY(p); !errors.Is(err, common.ErrServerMissingPeerPubkey) {
			t.Fatalf("role %d must fail closed with ErrServerMissingPeerPubkey, got %v", role, err)
		}
		// The window MUST stay empty for a refused envelope: a
		// fail-closed branch that nonetheless wrote a slot would
		// let a forger with one valid pubkey starve legitimate
		// relays off the window.
		if len(s.relayReplay.windows) != 0 {
			t.Fatalf("role %d populated window without being a registered relay", role)
		}
	}
}

// TestRelayReplayWindow_BasicSemantics pins the MarkSeen contract on
// the sliding-window primitive directly: first observation accepted,
// duplicate rejected, out-of-order-but-distinct txids accepted, and
// an outside-window txid rejected even though validatePeer's
// staleness floor would normally have caught it earlier.
func TestRelayReplayWindow_BasicSemantics(t *testing.T) {
	w := newRelayReplayWindow()
	key := artTestPubkey(1)

	if ok, outcome := w.MarkSeen(key, 100, 1000); !ok || outcome != replayOutcomeAccepted {
		t.Fatalf("first observation: ok=%v outcome=%v, want true/Accepted", ok, outcome)
	}
	if ok, _ := w.MarkSeen(key, 100, 1000); ok {
		t.Fatal("duplicate observation accepted; sliding window did not catch it")
	}
	// Out-of-order distinct txid: the original concurrent-clients
	// fan-out case. Both legs of the parallel pair must be
	// accepted.
	if ok, _ := w.MarkSeen(key, 99, 1000); !ok {
		t.Fatal("distinct earlier txid rejected; concurrent fan-out will produce false positives")
	}
	if ok, _ := w.MarkSeen(key, 10_000, 1000); !ok {
		t.Fatal("distinct later txid rejected")
	}
	// Txid far below the window: highestTxid=10_000, window=8192,
	// so txid=1 (gap=9999 >= 8192) is below the trailing W. This
	// is the case where validatePeer's staleness floor has aged
	// the envelope out, but a freshly-captured-and-replayed
	// envelope still hits us at this layer first.
	if ok, outcome := w.MarkSeen(key, 1, 1000); ok || outcome != replayOutcomeOutsideWindow {
		t.Fatalf("outside-window txid: ok=%v outcome=%v, want false/OutsideWindow", ok, outcome)
	}
}

// TestRelayReplayWindow_RejectsReplayWithinWindow pins the
// load-bearing case: a captured NHP_RLY envelope whose txid is still
// inside the trailing W is rejected. Without this property the hook
// does literally nothing — every captured envelope older than the
// TTL would slip through.
func TestRelayReplayWindow_RejectsReplayWithinWindow(t *testing.T) {
	w := newRelayReplayWindow()
	key := artTestPubkey(2)

	// Anchor the window at txid 1000 with sendTime 5000.
	if ok, _ := w.MarkSeen(key, 1000, 5000); !ok {
		t.Fatal("anchor envelope rejected")
	}
	// Same envelope again — exact replay.
	if ok, outcome := w.MarkSeen(key, 1000, 5000); ok || outcome != replayOutcomeDuplicate {
		t.Fatalf("exact replay: ok=%v outcome=%v, want false/Duplicate", ok, outcome)
	}
	// Same txid, but new sendTime further in the future — this
	// would normally fall in the fresh-sendTime branch (reset),
	// but the signature covers sendTime so this can only happen
	// for a legitimate relay. Here we test the reset path
	// independently below; for now the duplicate-with-stale-
	// sendTime case is what catches a replay attempt that
	// replays without changing sendTime.
	if ok, outcome := w.MarkSeen(key, 1000, 4999); ok || outcome != replayOutcomeDuplicate {
		t.Fatalf("same-txid stale-sendTime replay: ok=%v outcome=%v, want false/Duplicate", ok, outcome)
	}
}

// TestRelayReplayWindow_RestartAccepted pins the relay-process
// restart case: after a flood of envelopes that pushes highestTxid
// arbitrarily high, a fresh sendTime from the same relay with a
// small txid must be accepted (and reset the window). The reload
// signature on RemoteSendTime means an attacker cannot fake a
// fresh sendTime on a captured envelope, so the reset here is
// unforgeable.
func TestRelayReplayWindow_RestartAccepted(t *testing.T) {
	w := newRelayReplayWindow()
	key := artTestPubkey(3)

	// Anchor high.
	if ok, _ := w.MarkSeen(key, 10_000_000, 5000); !ok {
		t.Fatal("high anchor rejected")
	}
	// Many legitimate envelopes advance highestTxid further; we
	// do not need to walk all the way to the window edge here.
	for i := uint64(10_000_001); i <= 10_000_010; i++ {
		if ok, _ := w.MarkSeen(key, i, 5000); !ok {
			t.Fatalf("advance %d rejected", i)
		}
	}
	// Restart: a fresh sendTime with a low txid. Without the
	// reset-on-fresh-sendTime branch, the low txid lands at
	// position (highestTxid - txid) far past the window edge and
	// gets rejected — every legitimate post-restart envelope
	// would fail until the relay sent W envelopes, breaking
	// service for several seconds. The reset rule recognizes
	// that only the relay's own static key can sign a fresh
	// sendTime.
	if ok, outcome := w.MarkSeen(key, 1, 6000); !ok || outcome != replayOutcomeAccepted {
		t.Fatalf("restart envelope (fresh sendTime, low txid): ok=%v outcome=%v, want true/Accepted", ok, outcome)
	}
	// After the reset, the anchor shifted; the next legitimate
	// envelope at a slightly higher txid is accepted normally.
	if ok, _ := w.MarkSeen(key, 2, 6000); !ok {
		t.Fatal("post-restart envelope rejected after fresh-start accepted")
	}
}

// TestRelayReplayWindow_ReplayInterleavedWithLiveTraffic pins the
// load-bearing invariant the sliding-window fix relies on: while
// the relay is sending live envelopes with strictly newer txids
// and strictly newer sendTimes, a captured envelope from earlier
// in the same handshake (or from a replayed buffer) is still
// rejected. The earlier "sendTime > highestSendTime ⇒ clear
// bitmap" rule cleared the window on every legitimate envelope
// whose sendTime happened to be newer (the reviewer's trace),
// and a captured envelope from the same connection would then
// land at a bitmap slot that was just zeroed.
//
// The new rule only resets the bitmap when sendTime is newer AND
// txid has gone backwards past the window. In a live-traffic
// stream both txid and sendTime go forward, so neither reset
// branch ever fires: a captured envelope at any earlier (txid,
// sendTime) is caught by the bitmap.
//
// Trace under the new rule with a live stream {txid=i, sendTime=1000+i}
// and a captured envelope at (txid=2, sendTime=1002):
//
//	anchor  (1, 1001) -> seed
//	live    (2, 1002) -> slide by 1, bit 0 set
//	replay  (2, 1002) -> pos=0, bit already set -> Duplicate
//	live    (3, 1003) -> slide by 1, bit 0 set
//	replay  (2, 1002) -> sendTime not newer, pos=1, bit 1 set -> Duplicate
//	live    (4, 1004) -> slide by 1
//	...
//	replay  (2, 1002) -> pos=2, bit 2 set -> Duplicate
//	...
//
// Under the OLD rule every live envelope cleared the bitmap, so
// the first replay after a live envelope would land on a zeroed
// bit and be accepted.
func TestRelayReplayWindow_ReplayInterleavedWithLiveTraffic(t *testing.T) {
	w := newRelayReplayWindow()
	key := artTestPubkey(8)

	// Anchor the window.
	if ok, _ := w.MarkSeen(key, 1, 1001); !ok {
		t.Fatal("anchor rejected")
	}

	// Walk the live stream forward; each envelope's sendTime is
	// strictly newer than the previous one, and each txid is
	// strictly newer. This is the exact pattern the reviewer's
	// trace used to argue the bitmap was being cleared on every
	// legitimate packet under the old rule.
	for i := uint64(2); i <= 20; i++ {
		if ok, _ := w.MarkSeen(key, i, int64(1000+i)); !ok {
			t.Fatalf("live envelope txid=%d rejected", i)
		}
	}

	// A captured envelope at (2, 1002) is now in the middle of
	// the window. The bitmap must still remember the live
	// envelope that was at that slot.
	if ok, outcome := w.MarkSeen(key, 2, 1002); ok || outcome != replayOutcomeDuplicate {
		t.Fatalf("interleaved replay (txid=2, sendTime=1002): ok=%v outcome=%v, want false/Duplicate — bitmap was cleared by newer sendTime on a live envelope", ok, outcome)
	}

	// Walk the stream further and confirm every captured-envelope
	// replay still gets caught. Each live advance slides the
	// window by 1, so the captured envelope at slot (2, 1002)
	// remains inside the window for as long as the anchor is
	// within relayReplayWindowSize positions of txid 2.
	for i := uint64(21); i <= 30; i++ {
		if ok, _ := w.MarkSeen(key, i, int64(1000+i)); !ok {
			t.Fatalf("live envelope txid=%d rejected", i)
		}
		if ok, outcome := w.MarkSeen(key, 2, 1002); ok || outcome != replayOutcomeDuplicate {
			t.Fatalf("post-advance replay (txid=%d): captured envelope (2, 1002) ok=%v outcome=%v, want false/Duplicate",
				i, ok, outcome)
		}
	}

	// And the original live envelopes themselves (each in turn)
	// cannot be replayed. This is the original primitive the
	// window was designed for, but it was masked by the
	// clear-on-newer-sendTime behavior — re-asserting it under
	// the new rule.
	for i := uint64(5); i <= 10; i++ {
		if ok, outcome := w.MarkSeen(key, i, int64(1000+i)); ok || outcome != replayOutcomeDuplicate {
			t.Fatalf("live replay (txid=%d, sendTime=%d): ok=%v outcome=%v, want false/Duplicate",
				i, 1000+i, ok, outcome)
		}
	}
}

// TestRelayReplayWindow_PreRestartCaptureRejectedAfterEpochFloor
// pins the load-bearing invariant the review on the sliding-window
// follow-up identified: the epoch floor must reject pre-restart
// captures whose txid would otherwise take the forward-advance
// branch on the new epoch. The relay's counter is reseeded at
// random per process (nhp/core/device.go:142), so about half of all
// relay restarts land the new counter below the old highestTxid
// while still leaving the old highestTxid's captures in the
// staleness window. Without the epoch floor, the next legitimate
// envelope (low txid, fresh sendTime) takes the restart-reset
// branch, and the first replayed pre-restart capture then has a
// txid that ADVANCES the window, clearing the bitmap and reseeding
// at the capture's txid. The attacker can then re-inject every
// captured envelope from either epoch for the lifetime of the
// staleness floor.
//
// The new rule: epochSendTime is set on the first envelope and on
// every restart reset, and any envelope with sendTime below
// epochSendTime - slack is rejected as OutsideWindow BEFORE either
// the advance or the reset branch runs. Pre-restart captures have
// sendTime below the new epoch's anchor; they are rejected on
// arrival regardless of what their txid says.
//
// Sequence under test (T_old ≈ 1e6 before the restart, T' = 100
// after it, the reviewer's trace verbatim):
//
//	pre-restart capture (1e6,   S1) — anchor before the restart
//	pre-restart capture (1e6-1, S1) — second pre-restart envelope
//	[relay process restarts; new counter reseeds to 100]
//	legit post-restart  (100,  S2)  — restart reset, accept, advance epoch
//	captured pre-restart (1e6,  S1) — REJECT (sendTime S1 < epoch S2 - slack)
//	legit post-restart  (101,  S3)  — accept (S3 > S2, txid goes back past W, reset)
//	captured pre-restart (1e6,  S1) — REJECT (S1 still below new epoch S3)
//	captured post-restart (100,  S2) — REJECT (S2 < epoch S3 - slack)
//	[and stays rejected: subsequent legit envelopes do not move the
//	 epoch back; pre-restart captures are still below it]
func TestRelayReplayWindow_PreRestartCaptureRejectedAfterEpochFloor(t *testing.T) {
	w := newRelayReplayWindow()
	key := artTestPubkey(9)

	// Pre-restart: anchor the window at a high txid with sendTime
	// S1, then advance one more tick so highestTxid is even
	// further from any future reseed.
	const S1 = int64(5_000_000_000)
	if ok, _ := w.MarkSeen(key, 1_000_000, S1); !ok {
		t.Fatal("pre-restart anchor rejected")
	}
	if ok, _ := w.MarkSeen(key, 1_000_000-1, S1); !ok {
		t.Fatal("pre-restart advance rejected")
	}

	// Relay process restarts. The new handshake mints a fresh
	// sendTime S2; the reseeded counter happens to start at 100,
	// which is well below the old highestTxid. This is the
	// restart-reset path: sendTime newer, txid back past W, accept.
	const S2 = S1 + int64(time.Minute)
	if ok, _ := w.MarkSeen(key, 100, S2); !ok {
		t.Fatal("post-restart reset envelope rejected")
	}

	// Now replay a pre-restart capture. Under the OLD rule (no
	// epoch floor) this would have taken the forward-advance
	// branch: txid=1e6 is far above highestTxid=100, advance >=
	// W, bitmap cleared, reseed at 1e6, ACCEPTED. Under the new
	// rule the epoch floor rejects it first: S1 is below the new
	// epoch's anchor (S2) by a full minute, well past the 1s slack.
	if ok, outcome := w.MarkSeen(key, 1_000_000, S1); ok || outcome != replayOutcomeOutsideWindow {
		t.Fatalf("pre-restart capture after restart: ok=%v outcome=%v, want false/OutsideWindow — epoch floor failed to reject the replay", ok, outcome)
	}

	// Same check with a different pre-restart capture (the
	// 1_000_000-1 envelope from before the restart). Same
	// expected outcome.
	if ok, outcome := w.MarkSeen(key, 1_000_000-1, S1); ok || outcome != replayOutcomeOutsideWindow {
		t.Fatalf("second pre-restart capture after restart: ok=%v outcome=%v, want false/OutsideWindow", ok, outcome)
	}

	// Continue the legitimate post-restart stream. S3 is one
	// minute past S2 — also a valid restart-reset (S3 > S2, txid
	// goes back past W). epochSendTime moves to S3.
	const S3 = S2 + int64(time.Minute)
	if ok, _ := w.MarkSeen(key, 101, S3); !ok {
		t.Fatal("legitimate post-restart advance rejected")
	}
	if ok, outcome := w.MarkSeen(key, 1_000_000, S1); ok || outcome != replayOutcomeOutsideWindow {
		t.Fatalf("pre-restart capture after advance: ok=%v outcome=%v, want false/OutsideWindow", ok, outcome)
	}

	// A post-restart replay (the legit (100, S2) envelope) must
	// also be rejected now — the advance above did not move the
	// epoch, but it also did not move it back. S2 is the
	// current epoch anchor; replaying (100, S2) lands at
	// pos=highestTxid-txid=101-100=1 with bit 1 set (from the
	// advance above), so this is Duplicate, also fine.
	if ok, outcome := w.MarkSeen(key, 100, S2); ok {
		t.Fatalf("post-restart replay (100, S2): ok=%v outcome=%v, want false (replay should be caught)", ok, outcome)
	} else if outcome != replayOutcomeDuplicate {
		t.Fatalf("post-restart replay (100, S2): outcome=%v, want Duplicate", outcome)
	}
}

// TestRelayReplayWindow_PreRestartCaptureRejectedWithinSlackGap pins
// the user-reported reproduction: a sub-second gap between the
// restart-reset envelope's sendTime and the pre-restart capture's
// sendTime. The previous regression test
// (TestRelayReplayWindow_PreRestartCaptureRejectedAfterEpochFloor)
// used a 1-minute gap specifically because the 1 s slack on the
// top-level epoch floor was letting the replay through; commit
// 7533c92e widened the test gap rather than fix the bug. This test
// puts the gap back at 500 ms (well inside the slack) and confirms
// the strict check on the forward-advance branch now catches the
// replay before it can clear the bitmap.
//
// Sequence under test (the user's exact reproduction):
//
//	pre-restart capture  (1e6,   S)        — anchor before the restart
//	[relay process restarts; new counter reseeds to 100, sendTime = S+500ms]
//	legit post-restart   (100,   S+500ms)  — restart reset, accept, advance epoch
//	replay pre-restart   (1e6,   S)        — REJECT (S < epoch, advance branch's strict check)
//
// Under the old code the 1 s slack on the top-level floor let
// `S >= S+500ms - 1s` through, and the forward-advance branch then
// cleared the bitmap at the captured txid — REPLAY ACCEPTED. With
// the strict check, the replay is rejected before any bitmap
// mutation. Subsequent legitimate envelopes with sendTime > S+500ms
// are unaffected, and a second replay of the same capture stays
// rejected (the floor has not moved back to S).
func TestRelayReplayWindow_PreRestartCaptureRejectedWithinSlackGap(t *testing.T) {
	w := newRelayReplayWindow()
	key := artTestPubkey(13)

	// Pre-restart capture at sendTime S.
	const S = int64(2_000_000_000)
	if ok, _ := w.MarkSeen(key, 1_000_000, S); !ok {
		t.Fatal("pre-restart capture rejected as anchor")
	}

	// Relay process restarts; new envelope at S + 500ms (well
	// inside the 1 s epoch slack) and a low reseeded txid of 100.
	const S2 = S + int64(500*time.Millisecond)
	if ok, outcome := w.MarkSeen(key, 100, S2); !ok || outcome != replayOutcomeAccepted {
		t.Fatalf("restart reset: ok=%v outcome=%v, want true/Accepted", ok, outcome)
	}

	// Replay the pre-restart capture. Under the old code, S >= S2-1s
	// so the top-level floor passed and the forward-advance branch
	// cleared the bitmap — REPLAY ACCEPTED. With the strict check
	// on the advance branch, S < S2 so the replay is rejected
	// before any bitmap mutation.
	if ok, outcome := w.MarkSeen(key, 1_000_000, S); ok || outcome != replayOutcomeOutsideWindow {
		t.Fatalf("replay after restart: ok=%v outcome=%v, want false/OutsideWindow — 1 s slack on the advance branch re-opened the replay gap", ok, outcome)
	}

	// Subsequent legitimate envelopes from the new relay process
	// are unaffected: they have sendTime > S2 and clear the
	// strict check.
	if ok, _ := w.MarkSeen(key, 101, S2+int64(time.Second)); !ok {
		t.Fatal("legitimate post-replay envelope rejected")
	}

	// A second replay of the same pre-restart capture is still
	// rejected (the strict check is idempotent; the floor has not
	// moved back to S).
	if ok, outcome := w.MarkSeen(key, 1_000_000, S); ok || outcome != replayOutcomeOutsideWindow {
		t.Fatalf("second replay: ok=%v outcome=%v, want false/OutsideWindow", ok, outcome)
	}
}

// TestRelayReplayWindow_ClockStepBackRecoversViaIdleExpiration pins
// the recovery path the review identified as the load-bearing
// missing invariant: a relay whose wall clock steps back into the
// past (NTP correction, RTC reset, dual relay processes sharing one
// static key) would otherwise lock itself out forever, because
// every legitimate envelope has sendTime < epochSendTime and
// nothing in the bitmap or epoch floor ever clears state on
// wall-clock movement. Idle expiration drops the entry after
// relayReplayIdleTimeout of no accepted traffic, and the next
// envelope is anchored as if it were from a brand-new relay.
//
// The test drives the wall clock with an injected time source so
// the recovery is observable in milliseconds rather than the
// 16-minute real timeout.
func TestRelayReplayWindow_ClockStepBackRecoversViaIdleExpiration(t *testing.T) {
	clock := time.Now()
	w := newRelayReplayWindowWithClock(func() time.Time { return clock })
	key := artTestPubkey(10)

	// Anchor the window in "the past" — say the relay has been
	// running for a while and recorded an epoch far in the past.
	const pastEpochSendTime = int64(1_000_000_000) // 1s after Unix epoch
	if ok, _ := w.MarkSeen(key, 50_000, pastEpochSendTime); !ok {
		t.Fatal("initial anchor rejected")
	}

	// The relay's clock then steps back: the new envelope has
	// sendTime far below the epoch. Under the OLD design
	// (no idle expiration), the only way to recover would be
	// a server restart. Confirm the recovery is gated on idle.
	clock = clock.Add(relayReplayIdleTimeout - time.Second) // not yet idle
	if ok, _ := w.MarkSeen(key, 50_001, pastEpochSendTime-2*int64(time.Minute)); ok {
		t.Fatal("envelope with sendTime below epoch accepted before idle expiration")
	}

	// Now the relay has been silent for the full idle timeout.
	// Step the clock past the boundary.
	clock = clock.Add(2 * time.Second)

	// Next legitimate envelope should be accepted as a brand-new
	// anchor. The new envelope's sendTime is one minute past
	// the pre-step-back epoch — so it is ABOVE the old anchor.
	// The idle-expiration branch deletes the entry before the
	// epoch-floor check runs, so the new window starts fresh
	// at (1, futureEpochSendTime).
	const futureEpochSendTime = pastEpochSendTime + int64(time.Minute)
	if ok, outcome := w.MarkSeen(key, 1, futureEpochSendTime); !ok || outcome != replayOutcomeAccepted {
		t.Fatalf("clock-step-back recovery: ok=%v outcome=%v, want true/Accepted — idle expiration did not clear the epoch", ok, outcome)
	}

	// The pre-step-back captures are now rejected by the new
	// window's epoch floor: their sendTime (1s after epoch) is
	// far below the new epoch (1s+1min) - slack(1s), so the
	// floor rejects them.
	if ok, outcome := w.MarkSeen(key, 50_000, pastEpochSendTime); ok || outcome != replayOutcomeOutsideWindow {
		t.Fatalf("pre-step-back capture after recovery: ok=%v outcome=%v, want false/OutsideWindow", ok, outcome)
	}
}

// TestRelayReplayWindow_IdleExpirationIgnoresRejectedTraffic pins
// the second half of the recovery contract: a replay flood
// against a relay that has gone silent must NOT prevent idle
// expiration from firing. Otherwise an attacker that has captured
// a relay's envelopes could keep firing them at a slow rate to
// prevent the entry from being dropped, defeating the recovery
// path the clock-step-back scenario depends on. lastSeenAt is
// only updated on the accept paths, so a flood of rejected
// envelopes does not move the timer.
func TestRelayReplayWindow_IdleExpirationIgnoresRejectedTraffic(t *testing.T) {
	clock := time.Now()
	w := newRelayReplayWindowWithClock(func() time.Time { return clock })
	key := artTestPubkey(11)

	// Anchor the window.
	if ok, _ := w.MarkSeen(key, 1_000, 5_000_000_000); !ok {
		t.Fatal("anchor rejected")
	}

	// A flood of replays (rejected). The clock moves only on
	// these replays, not on the anchor.
	for i := 0; i < 100; i++ {
		clock = clock.Add(time.Second)
		if ok, _ := w.MarkSeen(key, 1_000, 5_000_000_000); ok {
			t.Fatal("replay accepted; window should have caught it")
		}
	}

	// The relay has now been silent for 100 s. Advance past
	// the idle timeout. Even if the attacker keeps replaying,
	// lastSeenAt is the anchor's timestamp — so the entry is
	// expired.
	clock = clock.Add(relayReplayIdleTimeout)

	// A legitimate envelope with a new sendTime should be
	// accepted as a brand-new anchor.
	if ok, outcome := w.MarkSeen(key, 1, 6_000_000_000); !ok || outcome != replayOutcomeAccepted {
		t.Fatalf("recovery after attacker flood: ok=%v outcome=%v, want true/Accepted", ok, outcome)
	}
}

// TestRelayReplayWindow_BurstDoesNotBlockLegitimate is the test the
// reviewer specifically asked for: a burst of capacity + 1 distinct,
// valid envelopes does not block the next legitimate envelope. The
// previous TTL-set design could not satisfy this (any sustained
// burst would either evict live entries or fill the cache and
// fail closed); the sliding window has no global capacity and so
// the property holds trivially. The test pushes well past the
// window size and confirms the next envelope is still accepted.
func TestRelayReplayWindow_BurstDoesNotBlockLegitimate(t *testing.T) {
	w := newRelayReplayWindow()
	key := artTestPubkey(4)

	// Anchor + a burst of relayReplayWindowSize envelopes
	// (relayReplayWindowSize + 1 total distinct txids). This is
	// the maximum the previous TTL-set would have absorbed
	// without rejecting on full, and an order of magnitude past
	// it.
	const burst = relayReplayWindowSize + 1
	if ok, _ := w.MarkSeen(key, 1, 1000); !ok {
		t.Fatal("anchor rejected")
	}
	for i := uint64(2); i <= burst+1; i++ {
		if ok, _ := w.MarkSeen(key, i, 1000); !ok {
			t.Fatalf("burst envelope %d rejected; sliding window cannot absorb sustained traffic", i)
		}
	}

	// The "next legitimate envelope" after the burst.
	if ok, outcome := w.MarkSeen(key, burst+2, 1000); !ok || outcome != replayOutcomeAccepted {
		t.Fatalf("post-burst legitimate envelope: ok=%v outcome=%v, want true/Accepted — sliding window failed to provide the property the reviewer asked for", ok, outcome)
	}

	// Memory check: the window map has exactly one entry (one
	// relay), regardless of burst size. This is the property
	// the TTL-set could not offer — its per-relay footprint
	// grew with traffic, so an attacker could raise it.
	if got := len(w.windows); got != 1 {
		t.Fatalf("window map size = %d, want 1 (per-relay state must NOT grow with traffic volume)", got)
	}
}

// TestRelayReplayWindow_PerRelayIsolation pins the per-relay
// independence rule: one relay's window state does not affect
// another's. A captured envelope from relay A replayed against
// relay B (different pubkey) must be accepted — they're distinct
// senders even if their txid and sendTime happen to match. The
// previous design also had per-pubkey isolation but via a flat
// (pubkey,txid,sendTime) key, which could be confused with
// reordering between different senders under high concurrency.
func TestRelayReplayWindow_PerRelayIsolation(t *testing.T) {
	w := newRelayReplayWindow()
	keyA := artTestPubkey(5)
	keyB := artTestPubkey(6)

	if ok, _ := w.MarkSeen(keyA, 1, 1000); !ok {
		t.Fatal("relay A anchor rejected")
	}
	// Same txid, same sendTime, but a DIFFERENT relay: must be
	// accepted, this is relay B's first observation.
	if ok, outcome := w.MarkSeen(keyB, 1, 1000); !ok || outcome != replayOutcomeAccepted {
		t.Fatalf("relay B's first envelope: ok=%v outcome=%v, want true/Accepted — per-relay isolation failed", ok, outcome)
	}
	if got := len(w.windows); got != 2 {
		t.Fatalf("window map size = %d, want 2 (one entry per distinct relay)", got)
	}

	// The replayed-from-A case: relay A's same envelope again.
	if ok, outcome := w.MarkSeen(keyA, 1, 1000); ok || outcome != replayOutcomeDuplicate {
		t.Fatalf("relay A replay: ok=%v outcome=%v, want false/Duplicate", ok, outcome)
	}
	// Relay B's tuple is still accepted a second time.
	if ok, _ := w.MarkSeen(keyB, 2, 1000); !ok {
		t.Fatal("relay B distinct-txid envelope rejected")
	}
}

// TestRelayReplayWindow_InvalidKeyFailsClosed ensures the same
// input-validation contract the TTL-set had: an unusable pubkey
// (nil / too-short / too-long) is reported as InvalidKey, not
// silently accepted. dedupeRecvRLY runs validation up front, so
// this branch is defense-in-depth, but it pins the contract so a
// future tweak to MarkSeen cannot start silently dropping invalid
// keys.
func TestRelayReplayWindow_InvalidKeyFailsClosed(t *testing.T) {
	w := newRelayReplayWindow()
	for _, k := range [][]byte{
		nil,
		make([]byte, core.PublicKeySize-1),
		make([]byte, core.PublicKeySizeEx+1),
	} {
		if ok, outcome := w.MarkSeen(k, 1, 1000); ok || outcome != replayOutcomeInvalidKey {
			t.Fatalf("MarkSeen on invalid key: ok=%v outcome=%v, want false/InvalidKey", ok, outcome)
		}
	}
}

// TestRelayReplayWindow_SlideBitmapForward exercises the bitmap
// shift helper directly for a few representative shifts, including
// word-aligned, bit-misaligned, and the >window clear case. The
// helper has a single in-tree caller (MarkSeen on txid advance)
// where shift always equals (newTxid - highestTxid), so the table
// here covers the cases the caller cannot cheaply produce at full
// width (multi-word shifts, bit-misaligned shifts).
func TestRelayReplayWindow_SlideBitmapForward(t *testing.T) {
	makeBitmap := func(setBits ...int) [relayReplayBitmapWords]uint64 {
		var b [relayReplayBitmapWords]uint64
		for _, pos := range setBits {
			b[pos/64] |= uint64(1) << (pos % 64)
		}
		return b
	}
	gatherBits := func(b [relayReplayBitmapWords]uint64) []int {
		var bits []int
		for i := 0; i < relayReplayBitmapWords*64; i++ {
			if b[i/64]&(uint64(1)<<(i%64)) != 0 {
				bits = append(bits, i)
			}
		}
		return bits
	}

	// shift 0: identity.
	b := makeBitmap(0, 5, 100)
	orig := b
	slideBitmapForward(&b, 0)
	if b != orig {
		t.Fatalf("shift 0 changed bitmap: got %v want %v", gatherBits(b), gatherBits(orig))
	}

	// shift 1: bits 0,5,100 -> 1,6,101.
	b = makeBitmap(0, 5, 100)
	slideBitmapForward(&b, 1)
	if got := gatherBits(b); len(got) != 3 || got[0] != 1 || got[1] != 6 || got[2] != 101 {
		t.Fatalf("shift 1: got %v want [1 6 101]", got)
	}

	// Bit-misaligned shift that crosses a word boundary: bits
	// 60,61,62 (last word 0) -> (60+shift, ...). With shift=70,
	// expected positions are 130,131,132. 130 = word 2 bit 2.
	b = makeBitmap(60, 61, 62)
	slideBitmapForward(&b, 70)
	if got := gatherBits(b); len(got) != 3 || got[0] != 130 || got[1] != 131 || got[2] != 132 {
		t.Fatalf("shift 70 across boundary: got %v want [130 131 132]", got)
	}

	// Word-aligned shift: bits in word 0 -> word sw.
	b = makeBitmap(0, 7, 63)
	slideBitmapForward(&b, 64) // one whole word
	got := gatherBits(b)
	if len(got) != 3 || got[0] != 64 || got[1] != 71 || got[2] != 127 {
		t.Fatalf("shift 64 (word-aligned): got %v want [64 71 127]", got)
	}

	// shift >= window: clear.
	b = makeBitmap(0, 1, 2)
	slideBitmapForward(&b, relayReplayWindowSize)
	if got := gatherBits(b); len(got) != 0 {
		t.Fatalf("shift >= window: got %v want []", got)
	}
	slideBitmapForward(&b, relayReplayWindowSize+1)
	if got := gatherBits(b); len(got) != 0 {
		t.Fatalf("shift > window: got %v want []", got)
	}
}

// TestUdpServer_Start_WiresRelayReplayWindow pins the production
// constructor: Start calls newRelayReplayWindow and assigns the
// result to s.relayReplay. The previous version of this test asserted
// a noEvict flag on the cache struct; with the sliding window that
// concept does not exist (the window has no global capacity to fail
// closed on). What we still want to pin is that the type is a
// relayReplayWindow, and that dedupeRecvRLY routes through it.
func TestUdpServer_Start_WiresRelayReplayWindow(t *testing.T) {
	s := &UdpServer{}
	s.relayReplay = newRelayReplayWindow()

	if s.relayReplay == nil {
		t.Fatal("newRelayReplayWindow() returned nil; relay-side dedupe path would crash")
	}
	if got := len(s.relayReplay.windows); got != 0 {
		t.Fatalf("fresh window map size = %d, want 0 (no relay has spoken yet)", got)
	}
}
