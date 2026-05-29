package server

import (
	"net"
	"sync"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// newTestServerForCounter spins up the bare minimum UdpServer state
// needed to exercise per-relay counter accounting in isolation: maps,
// mutexes, and nothing else. No listener, no goroutines.
func newTestServerForCounter() *UdpServer {
	return &UdpServer{
		remoteConnectionMap: make(map[string]*UdpConn),
		relayConnCount:      make(map[string]int),
	}
}

// makeRelayConn fabricates a UdpConn shaped like one HandleRelayForward
// would have inserted: same mapKey form, real RemoteAddr/RealRemoteAddr
// so relayAddrFromConnKey and the rest of the counter logic see what
// they'd see in production. Uses the production key builder
// (relayConnKeyPrefix + relayConnKeySep) rather than literal strings so
// the test cannot drift from the real format if the separator changes.
func makeRelayConn(relayHostPort, realClientHostPort string) (*UdpConn, string, string) {
	relayUDP, _ := net.ResolveUDPAddr("udp", relayHostPort)
	clientUDP, _ := net.ResolveUDPAddr("udp", realClientHostPort)
	connKey := relayConnKeyPrefix + relayHostPort + relayConnKeySep + realClientHostPort
	conn := &UdpConn{
		mapKey: connKey,
		ConnData: &core.ConnectionData{
			RemoteAddr:     relayUDP,
			RealRemoteAddr: clientUDP,
		},
	}
	return conn, connKey, relayHostPort
}

// TestValidateRelaySourceAddr_FlagToggles is the production-safety
// invariant: with allowPrivate=false, RFC1918 / loopback / CGNAT must
// be rejected so a compromised relay cannot inject fabricated entries
// into the server's connection map. With allowPrivate=true (set only
// in trusted local-only demos), the same private addresses must be
// accepted so that Docker Desktop's vpnkit gateway (and similar local
// NAT setups) can forward the docker-compose demo end-to-end.
//
// Port sanity is independent of the flag — out-of-range ports must
// always be rejected; there is no demo-friendly reason for them.
func TestValidateRelaySourceAddr_FlagToggles(t *testing.T) {
	const goodPort = 443

	tests := []struct {
		name         string
		ip           string
		port         int
		allowPrivate bool
		wantReject   bool
	}{
		// Production default: flag=false.
		{"public IP accepted in production", "203.0.113.5", goodPort, false, false},
		{"RFC1918 rejected in production", "192.168.65.1", goodPort, false, true},
		{"loopback rejected in production", "127.0.0.1", goodPort, false, true},
		{"CGNAT rejected in production", "100.64.0.1", goodPort, false, true},

		// Demo override: flag=true.
		{"Docker NAT gateway accepted under demo flag", "192.168.65.1", goodPort, true, false},
		{"loopback accepted under demo flag", "127.0.0.1", goodPort, true, false},
		{"public IP still accepted under demo flag", "203.0.113.5", goodPort, true, false},

		// Port sanity, both modes.
		{"zero port rejected even with flag off", "203.0.113.5", 0, false, true},
		{"negative port rejected even with flag on", "192.168.65.1", -1, true, true},
		{"port above range rejected with flag on", "192.168.65.1", 65536, true, true},
		{"port above range rejected with flag off", "203.0.113.5", 70000, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			reason := validateRelaySourceAddr(ip, tt.port, tt.allowPrivate)
			rejected := reason != ""
			if rejected != tt.wantReject {
				t.Fatalf("validateRelaySourceAddr(%s, %d, allowPrivate=%v) reason=%q reject=%v, want reject=%v",
					tt.ip, tt.port, tt.allowPrivate, reason, rejected, tt.wantReject)
			}
		})
	}
}

// TestValidateRelaySourceAddr_NilIP makes sure a SourceAddr.Ip that
// fails net.ParseIP (typo, empty string, etc.) is rejected regardless
// of the demo flag. The handler upstream is the only caller, so this
// is the boundary where we catch malformed input from the relay.
func TestValidateRelaySourceAddr_NilIP(t *testing.T) {
	for _, allowPrivate := range []bool{false, true} {
		if reason := validateRelaySourceAddr(nil, 443, allowPrivate); reason != "malformed" {
			t.Fatalf("nil IP should be rejected as malformed (allowPrivate=%v); got %q", allowPrivate, reason)
		}
	}
}

// TestIsRoutablePublicIP guards the SourceAddr filter that
// HandleRelayForward uses to reject fabricated entries from a
// misbehaving relay. The intent is "only accept addresses a real public
// client could plausibly originate from", so anything reserved /
// non-routable must be rejected.
//
// If a future change widens what counts as routable (e.g. relaxes the
// CGNAT bounds), these tests will catch the regression — silently
// accepting a reserved-range source IP would let a compromised relay
// install firewall rules under bogus IPs that audit tools would never
// expect to see.
func TestIsRoutablePublicIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		// Plausible public IPs — accepted.
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"203.0.113.5", true},          // TEST-NET-3 is technically reserved but
		{"2001:4860:4860::8888", true}, // global IPv6

		// Loopback.
		{"127.0.0.1", false},
		{"127.255.255.254", false},
		{"::1", false},

		// Unspecified.
		{"0.0.0.0", false},
		{"::", false},

		// RFC 1918 private.
		{"10.0.0.1", false},
		{"172.16.0.1", false},
		{"172.31.255.254", false},
		{"192.168.1.1", false},

		// RFC 4193 IPv6 unique-local.
		{"fc00::1", false},
		{"fd12:3456:789a::1", false},

		// Link-local.
		{"169.254.1.1", false},
		{"fe80::1", false},

		// Multicast.
		{"224.0.0.1", false},
		{"239.255.255.255", false},
		{"ff02::1", false},

		// CGNAT (RFC 6598) — explicit boundary checks.
		{"100.63.255.255", true}, // one below CGNAT range
		{"100.64.0.0", false},    // start of CGNAT
		{"100.64.0.1", false},
		{"100.127.255.255", false}, // end of CGNAT
		{"100.128.0.0", true},      // one above CGNAT range
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("net.ParseIP(%q) returned nil — fix the test data", tt.ip)
			}
			got := isRoutablePublicIP(ip)
			if got != tt.expected {
				t.Errorf("isRoutablePublicIP(%s) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}

	// nil input must not panic and must be rejected.
	t.Run("nil", func(t *testing.T) {
		if isRoutablePublicIP(nil) {
			t.Errorf("isRoutablePublicIP(nil) = true, want false")
		}
	})
}

// TestRelayAddrFromConnKey ensures the mapKey parser used by the
// per-relay connection counter correctly recovers the relay's address
// from compound keys of the form "relay|<relayAddr>|<realClientAddr>".
//
// Coverage requirements driven by the production failure mode this
// guards: the previous ':'-separated key form silently mis-parsed any
// mapKey whose realClientAddr was IPv6 (because the
// strings.LastIndexByte-based right-peel would land inside the address'
// own colons), leaking the relay's per-connection counter upward by one
// per IPv6 forward — eventually rejecting legitimate traffic when the
// counter crossed MaxConnectionsPerRelay. The IPv6-client and IPv6-relay
// cases below are the regression fence.
func TestRelayAddrFromConnKey(t *testing.T) {
	// Build keys with the production helper so we cannot drift from
	// the real key format if the separator changes again.
	key := func(relay, client string) string {
		return relayConnKeyPrefix + relay + relayConnKeySep + client
	}

	tests := []struct {
		name   string
		mapKey string
		want   string
	}{
		// Direct UDP keys are not relay-forwarded.
		{"direct UDP key", "203.0.113.5:54321", ""},
		{"empty key", "", ""},

		// IPv4 relay + IPv4 client — the historical baseline case.
		{"v4 relay v4 client", key("198.51.100.1:62206", "203.0.113.5:54321"), "198.51.100.1:62206"},
		{"v4 relay v4 client port=1", key("198.51.100.1:62206", "203.0.113.5:1"), "198.51.100.1:62206"},
		{"v4 relay v4 client port=65535", key("198.51.100.1:62206", "203.0.113.5:65535"), "198.51.100.1:62206"},

		// IPv4 relay + IPv6 client: the original bug. Real
		// net.UDPAddr.String() form for IPv6 is "[host]:port", and
		// the address itself contains ':' separators that a right-
		// peeling parser would interpret as token boundaries. With
		// the '|' separator a single forward split recovers the
		// relay segment cleanly.
		{"v4 relay v6 client", key("198.51.100.1:62206", "[2001:db8::1]:80"), "198.51.100.1:62206"},
		{"v4 relay v6 loopback client", key("198.51.100.1:62206", "[::1]:443"), "198.51.100.1:62206"},

		// IPv6 relay + IPv4 client: the *worse* case under the old
		// parser — it would scramble the relay segment itself, so
		// the per-relay counter was keyed on a corrupted string and
		// would never decrement on teardown.
		{"v6 relay v4 client", key("[2001:db8::42]:62206", "203.0.113.5:54321"), "[2001:db8::42]:62206"},

		// IPv6 relay + IPv6 client: both segments contain ':' but
		// neither contains '|', so the parser stays correct.
		{"v6 relay v6 client", key("[2001:db8::42]:62206", "[2001:db8::1]:80"), "[2001:db8::42]:62206"},

		// Malformed keys.
		{"prefix only", "relay|", ""},
		{"no separator", "relay|foo", ""},                   // missing the second segment entirely
		{"empty real-client segment", "relay|foo|", ""},     // "<relay>|"
		{"empty relay segment", "relay||203.0.113.5:1", ""}, // "|<client>"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := relayAddrFromConnKey(tt.mapKey)
			if got != tt.want {
				t.Errorf("relayAddrFromConnKey(%q) = %q, want %q", tt.mapKey, got, tt.want)
			}
		})
	}
}

// TestTeardownPerRelayCounter_Decision pins the truth table for the
// per-relay counter teardown logic. The four combinations of
// (stillPresent, replaced) cover every path through the helper:
//
//   - (true,  false): the routine still owns the map entry and no
//     replacement claimed the slot — this is the only path that dec's.
//   - (true,  true):  HRF set replaced=true before CR-defer reached
//     the teardown; do NOT dec (replacement inherits the slot).
//   - (false, false): the map entry is gone but no replacement
//     happened (e.g. some hypothetical future shutdown path) — do
//     NOT dec, since whoever removed the entry should own the dec.
//   - (false, true):  HRF removed the entry AND set replaced; do NOT
//     dec (this is the normal stale-replace ordering).
//
// Pre-fix, "do NOT dec" was inferred from stillPresent=false alone, so
// the (true, true) row would have dec'd — leaking a slot below the
// true live count on every stale-replace race.
func TestTeardownPerRelayCounter_Decision(t *testing.T) {
	const relay = "198.51.100.10:62206"
	const client = "203.0.113.42:51234"

	tests := []struct {
		name         string
		stillPresent bool
		replaced     bool
		wantDec      bool
	}{
		{"normal teardown decrements", true, false, true},
		{"replaced-by-HRF must not decrement", true, true, false},
		{"map entry already gone, no replacement: do not decrement", false, false, false},
		{"map entry gone and replaced: do not decrement", false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServerForCounter()
			conn, _, relayAddr := makeRelayConn(relay, client)
			// Seed the counter so we can observe a dec.
			s.relayConnCount[relayAddr] = 1
			conn.replaced.Store(tt.replaced)

			s.teardownPerRelayCounter(conn, conn.mapKey, tt.stillPresent)

			got := s.getRelayConnCount(relayAddr)
			want := 1
			if tt.wantDec {
				want = 0 // decRelayConnCount deletes on <=0; getRelayConnCount returns 0 for missing key
			}
			if got != want {
				t.Fatalf("after teardown(stillPresent=%v, replaced=%v): counter=%d, want %d",
					tt.stillPresent, tt.replaced, got, want)
			}
		})
	}
}

// mockCrDefer mirrors connectionRoutine's teardown defer in
// udpserver.go: it grabs the map mutex, removes the entry if it
// still owns it (identity-checked, not just presence-checked, so a
// replacement conn HRF has already inserted is left alone), and
// routes the counter dec through the same teardownPerRelayCounter
// helper production uses. Returned for use from the stale-replace
// race tests below so both the deterministic and the concurrent
// variant agree on what "CR-defer ran" means.
func mockCrDefer(s *UdpServer, conn *UdpConn, connKey string) {
	s.remoteConnectionMapMutex.Lock()
	existing, ok := s.remoteConnectionMap[connKey]
	stillPresent := ok && existing == conn
	if stillPresent {
		delete(s.remoteConnectionMap, connKey)
	}
	s.remoteConnectionMapMutex.Unlock()
	s.teardownPerRelayCounter(conn, connKey, stillPresent)
}

// mockHrfReplaceOrInsert mirrors the relevant counter-accounting
// half of HandleRelayForward. It deliberately models BOTH branches
// of the lookup:
//
//   - found && IsClosed: stale-replace. Mark replaced (so CR-defer
//     skips its dec), delete the old entry, insert the new entry,
//     leave the counter untouched — the slot transfers.
//   - !found (or found && !IsClosed, which doesn't happen in these
//     tests): fresh insert. Inc the counter, insert the new entry.
//
// Modeling only the stale-replace branch (the original mock did
// this) would silently misreport the CR-first ordering: in
// production, after CR has deleted the entry, HRF's lookup misses
// and falls through to the fresh-insert path, re-incrementing the
// counter. Without that step, the deterministic test ends at
// counter=0 while production ends at counter=1, and the test
// assertion of "counter == 1" is unsatisfiable through the mock.
func mockHrfReplaceOrInsert(s *UdpServer, connKey, relay, client string) {
	s.remoteConnectionMapMutex.Lock()
	if existing, found := s.remoteConnectionMap[connKey]; found && existing.ConnData.IsClosed() {
		// Stale-replace: slot transfers, no counter change.
		existing.replaced.Store(true)
		delete(s.remoteConnectionMap, connKey)
		newConn, _, _ := makeRelayConn(relay, client)
		s.remoteConnectionMap[connKey] = newConn
		s.remoteConnectionMapMutex.Unlock()
		return
	}
	// Fresh-insert: HRF treats this as a genuinely new slot, so it
	// inc's the per-relay counter (msghandler.go:868-896 in
	// production) before inserting. Production also enforces caps
	// here; the cap path is exercised in dedicated tests and is
	// orthogonal to the race we're modeling, so skip it.
	s.remoteConnectionMapMutex.Unlock()
	newConn, _, relayAddr := makeRelayConn(relay, client)
	s.incRelayConnCount(relayAddr)
	s.remoteConnectionMapMutex.Lock()
	s.remoteConnectionMap[connKey] = newConn
	s.remoteConnectionMapMutex.Unlock()
}

// TestStaleReplace_CounterIsNetZero is the regression test for the
// double-decrement race that motivated this fix. The original code
// kept four short critical sections in HandleRelayForward and let
// connectionRoutine's teardown defer race against the stale-replace
// path; depending on the interleaving, the per-relay counter could
// either drop below or rise above the true live-connection count.
//
// The fix turns the stale-replace into a slot transfer: HRF detects
// the stale conn under the map mutex, marks it replaced (so CR-defer
// will skip its dec), deletes the old entry, and inserts the new
// entry — all without touching the counter. The new conn's matching
// dec on its own teardown is what eventually returns the slot.
//
// The test runs both orderings of CR-defer vs HRF and asserts the
// counter ends at exactly 1 either way (one live replacement conn).
//   - HRF-first: stale-replace claims the slot, CR-defer sees
//     replaced=true and skips its dec. Counter stays at 1.
//   - CR-first: CR-defer dec's (1 → 0), then HRF's lookup misses and
//     falls through to fresh-insert which inc's (0 → 1).
//
// Pre-fix this would have ended at 0 (or below) in one ordering and
// 2 in the other.
func TestStaleReplace_CounterIsNetZero(t *testing.T) {
	const relay = "198.51.100.10:62206"
	const client = "203.0.113.42:51234"

	tests := []struct {
		name string
		// crFirst: simulate connectionRoutine's teardown defer
		// running before HandleRelayForward. !crFirst: HRF first.
		crFirst bool
	}{
		{"CR-defer runs first then HRF re-inserts via fresh path", true},
		{"HRF stale-replace runs first then CR-defer trails", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServerForCounter()
			oldConn, connKey, relayAddr := makeRelayConn(relay, client)
			// Mirror what HandleRelayForward would have done on the
			// original insert: map entry + inc'd counter.
			s.remoteConnectionMap[connKey] = oldConn
			s.relayConnCount[relayAddr] = 1

			// The conn is stale: connectionRoutine has exited and
			// Close has flipped IsClosed. Drive it directly because
			// we don't want to run the full routine here.
			closeConn(oldConn)

			if tt.crFirst {
				mockCrDefer(s, oldConn, connKey)
				mockHrfReplaceOrInsert(s, connKey, relay, client)
			} else {
				mockHrfReplaceOrInsert(s, connKey, relay, client)
				mockCrDefer(s, oldConn, connKey)
			}

			got := s.getRelayConnCount(relayAddr)
			if got != 1 {
				t.Fatalf("counter=%d, want 1 (one live replacement conn). "+
					"CR-defer first=%v. Pre-fix this race produced 0 (double-dec) "+
					"or 2 (skip-dec) depending on ordering.", got, tt.crFirst)
			}
		})
	}
}

// TestStaleReplace_ConcurrentInterleavings runs the same race many
// times with real goroutines, relying on the runtime scheduler to
// produce both orderings. With -race this also asserts the absence
// of data races on relayConnCount and the replaced flag.
//
// Both orderings must end at counter=1 (matching the deterministic
// test above): one live replacement conn either way. A previous
// version of this test allowed counter ∈ {0, 1} because its mock
// for HRF omitted the fresh-insert branch — that hid the property
// the fix actually guarantees and weakened the regression fence.
func TestStaleReplace_ConcurrentInterleavings(t *testing.T) {
	const relay = "198.51.100.10:62206"
	const iterations = 1000

	for i := 0; i < iterations; i++ {
		s := newTestServerForCounter()
		// Use a per-iteration unique client port so each iteration
		// has its own connKey and counter starts fresh.
		client := "203.0.113.42:" + itoa(50000+i)
		oldConn, connKey, relayAddr := makeRelayConn(relay, client)
		s.remoteConnectionMap[connKey] = oldConn
		s.relayConnCount[relayAddr] = 1
		closeConn(oldConn)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			mockCrDefer(s, oldConn, connKey)
		}()
		go func() {
			defer wg.Done()
			mockHrfReplaceOrInsert(s, connKey, relay, client)
		}()
		wg.Wait()

		got := s.getRelayConnCount(relayAddr)
		if got != 1 {
			t.Fatalf("iteration %d: counter=%d, want 1.\n"+
				"  HRF-wins: replaces (no inc/dec); CR-defer sees replaced=true, skips dec → 1.\n"+
				"  CR-wins:  dec's (1→0); HRF lookup misses, fresh-insert inc's (0→1) → 1.\n"+
				"counter != 1 means one branch of the race is mis-accounting slots.", i, got)
		}
	}
}

// TestStaleReplace_CrDeferDoesNotOrphanReplacementMapEntry fences the
// map-side identity-check fix. After HRF has stale-replaced (mapKey
// now holds newConn), the old conn's teardown defer must NOT delete
// the map entry — newConn owns it and has its own routine pending.
//
// Pre-fix the defer keyed only on presence (`_, stillPresent := map[key]`)
// so a replacement entry got blindly deleted, orphaning newConn's
// connectionRoutine and any future RKN traffic for that client.
// Today this race is unreachable in production (IsClosed is only
// flipped by the owning routine), but if any future code adds an
// external Close() path the gap re-opens. Counter-side accounting
// already uses an explicit replaced flag to stay identity-aware; this
// test pins the matching property on the map side.
func TestStaleReplace_CrDeferDoesNotOrphanReplacementMapEntry(t *testing.T) {
	const relay = "198.51.100.10:62206"
	const client = "203.0.113.42:51234"

	s := newTestServerForCounter()
	oldConn, connKey, _ := makeRelayConn(relay, client)
	s.remoteConnectionMap[connKey] = oldConn
	closeConn(oldConn)

	// HRF goes first: replaces oldConn with newConn under the map mutex.
	mockHrfReplaceOrInsert(s, connKey, relay, client)
	newConn, ok := s.remoteConnectionMap[connKey]
	if !ok || newConn == oldConn {
		t.Fatalf("setup: HRF mock did not install a replacement entry; got ok=%v, same=%v", ok, newConn == oldConn)
	}

	// CR-defer trails. With the identity check it must observe
	// "entry is not me" and leave newConn alone.
	mockCrDefer(s, oldConn, connKey)

	got, present := s.remoteConnectionMap[connKey]
	if !present {
		t.Fatalf("CR-defer orphaned the replacement entry: map[connKey] was deleted after HRF inserted newConn. " +
			"Pre-fix the defer keyed on presence only; the identity check (existing == conn) closes that gap.")
	}
	if got != newConn {
		t.Fatalf("map[connKey] no longer points to newConn after CR-defer: got %p, want %p", got, newConn)
	}
}

// closeConn flips a ConnectionData into the IsClosed=true state
// without running the full Close() machinery (which would close
// channels we never created in the test fixture).
func closeConn(c *UdpConn) {
	// ConnectionData.closed is unexported; we can't poke it directly
	// from outside nhp/core. Instead, set up the minimal channels so
	// Close() can run flush+close without panicking. The test only
	// needs IsClosed()==true; the side-effect of channel teardown
	// inside the same fixture is harmless.
	c.ConnData.StopSignal = make(chan struct{})
	c.ConnData.SendQueue = make(chan *core.Packet, 1)
	c.ConnData.RecvQueue = make(chan *core.Packet, 1)
	c.ConnData.BlockSignal = make(chan struct{}, 1)
	c.ConnData.SetTimeoutSignal = make(chan struct{}, 1)
	c.ConnData.Close()
}

// itoa is a small local int→string helper to avoid pulling strconv
// into the test for one line.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
