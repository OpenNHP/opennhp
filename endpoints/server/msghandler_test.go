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
// they'd see in production.
func makeRelayConn(relayHostPort, realClientHostPort string) (*UdpConn, string, string) {
	relayUDP, _ := net.ResolveUDPAddr("udp", relayHostPort)
	clientUDP, _ := net.ResolveUDPAddr("udp", realClientHostPort)
	connKey := "relay:" + relayHostPort + ":" + realClientHostPort
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
// from compound keys of the form "relay:<relayAddr>:<realClientAddr>".
// Both segments are themselves "host:port" so the parser must strip
// exactly the trailing two colon-separated tokens.
func TestRelayAddrFromConnKey(t *testing.T) {
	tests := []struct {
		mapKey string
		want   string
	}{
		// Direct UDP keys are not relay-forwarded.
		{"203.0.113.5:54321", ""},
		{"", ""},

		// IPv4 relay + IPv4 client.
		{"relay:198.51.100.1:62206:203.0.113.5:54321", "198.51.100.1:62206"},

		// Same relay but realClient port edge values.
		{"relay:198.51.100.1:62206:203.0.113.5:1", "198.51.100.1:62206"},
		{"relay:198.51.100.1:62206:203.0.113.5:65535", "198.51.100.1:62206"},

		// Malformed keys.
		{"relay:", ""},
		{"relay:foo", ""},
		{"relay:foo:bar", ""}, // only one colon after prefix → not enough segments
	}

	for _, tt := range tests {
		t.Run(tt.mapKey, func(t *testing.T) {
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
// The test runs both orderings of CR-defer vs HRF stale-replace and
// asserts the counter ends at exactly 1 either way (one live
// replacement conn). Pre-fix this would have ended at 0 (or below)
// in one ordering and 2 in the other.
func TestStaleReplace_CounterIsNetZero(t *testing.T) {
	const relay = "198.51.100.10:62206"
	const client = "203.0.113.42:51234"

	tests := []struct {
		name string
		// crFirst: simulate connectionRoutine's teardown defer
		// running before HandleRelayForward's stale-replace path.
		// !crFirst: HRF runs first; CR-defer trails.
		crFirst bool
	}{
		{"CR-defer runs first then HRF takes over the slot", true},
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

			crDefer := func() {
				// Mirror udpserver.go connectionRoutine's defer.
				s.remoteConnectionMapMutex.Lock()
				_, stillPresent := s.remoteConnectionMap[connKey]
				if stillPresent {
					delete(s.remoteConnectionMap, connKey)
				}
				s.remoteConnectionMapMutex.Unlock()
				s.teardownPerRelayCounter(oldConn, connKey, stillPresent)
			}

			hrfStaleReplace := func() {
				// Mirror the slot-transfer half of
				// HandleRelayForward. Skips the full
				// packet-parsing/insert path because we only care
				// about the counter-accounting half of the race.
				s.remoteConnectionMapMutex.Lock()
				if existing, found := s.remoteConnectionMap[connKey]; found && existing.ConnData.IsClosed() {
					existing.replaced.Store(true)
					delete(s.remoteConnectionMap, connKey)
					// Slot transfers; no inc, no dec. The
					// replacement insert below would be the
					// "new conn inherits the slot" half.
					newConn, _, _ := makeRelayConn(relay, client)
					s.remoteConnectionMap[connKey] = newConn
				}
				s.remoteConnectionMapMutex.Unlock()
			}

			if tt.crFirst {
				crDefer()
				hrfStaleReplace()
			} else {
				hrfStaleReplace()
				crDefer()
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
// produce both orderings. With -race this also asserts the absence of
// data races on relayConnCount and the replaced flag. The invariant
// is the same as the deterministic test above: after the race
// completes, the counter must equal the live-replacement count (1).
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
			s.remoteConnectionMapMutex.Lock()
			_, stillPresent := s.remoteConnectionMap[connKey]
			if stillPresent {
				delete(s.remoteConnectionMap, connKey)
			}
			s.remoteConnectionMapMutex.Unlock()
			s.teardownPerRelayCounter(oldConn, connKey, stillPresent)
		}()
		go func() {
			defer wg.Done()
			s.remoteConnectionMapMutex.Lock()
			if existing, found := s.remoteConnectionMap[connKey]; found && existing.ConnData.IsClosed() {
				existing.replaced.Store(true)
				delete(s.remoteConnectionMap, connKey)
				newConn, _, _ := makeRelayConn(relay, client)
				s.remoteConnectionMap[connKey] = newConn
			}
			s.remoteConnectionMapMutex.Unlock()
		}()
		wg.Wait()

		got := s.getRelayConnCount(relayAddr)
		// Acceptable outcomes:
		//   - CR wins: HRF sees no entry → no replace → CR-defer dec's. counter=0.
		//   - HRF wins: HRF replaces (no inc, no dec) → CR-defer sees
		//     replaced=true OR stillPresent=false → skip dec. counter=1.
		// Either is consistent with "counter == live conn count": 0
		// live (CR finished cleanup, HRF never claimed slot) or 1
		// live (HRF inserted replacement).
		if got != 0 && got != 1 {
			t.Fatalf("iteration %d: counter=%d, want 0 (no replacement) or 1 (replacement inserted). "+
				"A value of 2 means CR-defer skipped dec while HRF didn't actually claim the slot. "+
				"A value of -clamped-to-deleted means both dec'd the same slot.", i, got)
		}
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
