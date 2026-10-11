package server

import (
	"encoding/base64"
	"errors"
	"net"
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
// consumer side (HandleRelayForward honours it).
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
// mapKey (or the fallback we synthesise in that branch). Passing a
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

// TestConnectionRoutine_BlockSignalDoesNotBlockRelayAddress pins the fix
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
// the relay-forward block list (the new behaviour), and the conn must
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

// TestDedupeRecvRLY_RegisteredRelay pins the third leg of the
// GHSA-674q-j8fh-972c follow-up: dedupe captured-and-replayed NHP_RLY
// envelopes on (relay pubkey, SenderTrxId, RemoteSendTime). NHP_RLY is
// exempted from the per-connection monotonic/flood gates, and the
// per-inner-client conn the decoded packet is re-injected on idles out
// after DefaultAgentConnectionTimeoutMs — without this hook, an
// on-path observer could replay a captured relay→server UDP datagram
// indefinitely and the server would re-authenticate it as fresh.
//
// The hook short-circuits on non-RLY headers so the chain can continue
// to ART's own dedupe — but for any RLY envelope whose decrypted peer
// pubkey is not in s.relayPeerMap (whether the peer is unregistered at
// all or registered under a non-relay role), the hook now FAILS CLOSED
// with ErrServerMissingPeerPubkey. The earlier pass-through was a
// replay hole during config reloads (a newly-registered relay was
// visible to the device peer pool that validatePeer uses before the
// relayPeerMap swap landed), and the fail-closed behaviour matches
// dedupeRecvART's "peer not registered as an AC" path.
func TestDedupeRecvRLY_RegisteredRelay(t *testing.T) {
	s := &UdpServer{
		relayReplay:  newRelayReplayCache(),
		relayPeerMap: make(map[string]*core.UdpPeer),
	}

	// Non-RLY header must not consume a slot.
	if err := s.dedupeRecvRLY(&core.PacketParserData{HeaderType: core.NHP_KNK, RemotePubKey: artTestPubkey(1), SenderTrxId: 1, RemoteSendTime: 1}); err != nil {
		t.Fatalf("non-RLY header rejected: %v", err)
	}
	if len(s.relayReplay.entries) != 0 {
		t.Fatal("non-RLY populated cache")
	}

	// RLY with no registered relay peer must FAIL CLOSED, not
	// short-circuit. The previous behaviour returned nil and left the
	// envelope un-replay-protected; under the new policy the only
	// safe answer to "I don't see this peer as a relay" is to refuse
	// the envelope. validatePeer would normally have rejected it
	// before reaching the hook, but a config-reload race could
	// produce a freshly-validated envelope for a peer the hook has
	// not yet seen — that race is why the policy lives here.
	if err := s.dedupeRecvRLY(&core.PacketParserData{HeaderType: core.NHP_RLY, RemotePubKey: artTestPubkey(2), SenderTrxId: 1, RemoteSendTime: 1}); !errors.Is(err, common.ErrServerMissingPeerPubkey) {
		t.Fatalf("unregistered RLY peer must fail closed with ErrServerMissingPeerPubkey, got %v", err)
	}
	if len(s.relayReplay.entries) != 0 {
		t.Fatal("unregistered relay populated cache")
	}

	// Now register a relay peer and exercise the cache.
	for _, size := range []int{core.PublicKeySize, core.PublicKeySizeEx} {
		key := artTestPubkey(3, size)
		s.relayPeerMap[base64.StdEncoding.EncodeToString(key)] = &core.UdpPeer{Type: core.NHP_RELAY}

		packet := &core.PacketParserData{
			HeaderType:     core.NHP_RLY,
			SenderTrxId:    uint64(size),
			RemotePubKey:   key,
			RemoteSendTime: 42,
		}
		if err := s.dedupeRecvRLY(packet); err != nil {
			t.Fatalf("first NHP_RLY with %d-byte key rejected: %v", size, err)
		}
		if err := s.dedupeRecvRLY(packet); !errors.Is(err, common.ErrServerDuplicateTransaction) {
			t.Fatalf("duplicate NHP_RLY with %d-byte key error = %v", size, err)
		}

		// A different send time / txid / peer key must not be rejected —
		// out-of-order envelopes from concurrent web agents are valid, and
		// the cache must scope by (pubkey, txid, sendTime) like artReplay.
		other := &core.PacketParserData{
			HeaderType:     core.NHP_RLY,
			SenderTrxId:    uint64(size) + 1,
			RemotePubKey:   key,
			RemoteSendTime: 43,
		}
		if err := s.dedupeRecvRLY(other); err != nil {
			t.Fatalf("fresh NHP_RLY envelope rejected: %v", err)
		}
	}
}

// TestDedupeRecvRLY_RejectsNonRelayBeforeCaching mirrors the
// ART-dedupe sanity check: a packet that walks through the cache
// because its msgType matches but whose decrypted peer is not the
// expected role must not consume a slot. Under the new fail-closed
// policy, "registered in relayPeerMap with a non-relay role" is just
// another way of saying "I have no record of this pubkey as a relay",
// and the hook returns ErrServerMissingPeerPubkey instead of nil —
// which is also the property that prevents a single forged peer key
// from filling the capacity and starving legitimate relays.
func TestDedupeRecvRLY_RejectsNonRelayBeforeCaching(t *testing.T) {
	key := artTestPubkey(7, core.PublicKeySize)
	for _, role := range []int{core.NHP_AGENT, core.NHP_AC, core.NHP_DB} {
		s := &UdpServer{
			relayReplay:  newRelayReplayCache(),
			relayPeerMap: map[string]*core.UdpPeer{base64.StdEncoding.EncodeToString(key): {Type: role}},
		}
		p := &core.PacketParserData{HeaderType: core.NHP_RLY, RemotePubKey: key, SenderTrxId: 1, RemoteSendTime: 1}
		if err := s.dedupeRecvRLY(p); !errors.Is(err, common.ErrServerMissingPeerPubkey) {
			t.Fatalf("role %d must fail closed with ErrServerMissingPeerPubkey, got %v", role, err)
		}
		// The cache MUST stay empty for an envelope the hook rejected:
		// a fail-closed branch that nonetheless wrote a (pubkey,txid,
		// sendTime) entry would let a forger with one valid pubkey
		// starve every legitimate relay off the tail of the cache.
		if len(s.relayReplay.entries) != 0 {
			t.Fatalf("role %d populated cache without being a registered relay", role)
		}
	}
}
