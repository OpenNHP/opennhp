package relay

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/utils"
)

// TestRealClientAddr covers the security-sensitive parts of how the
// relay derives the originating client address: X-Real-IP must only be
// honored when the direct TCP peer is on loopback (a local reverse
// proxy), and a missing/malformed X-Real-IP from such a peer must
// surface as an error rather than a silent fallback to 127.0.0.1.
func TestRealClientAddr(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xRealIP    string
		xff        string // must NEVER be honored; included to verify it's ignored
		wantErr    bool
		wantIP     string
		wantPort   int
	}{
		{
			name:       "loopback peer with valid X-Real-IP returns the header value",
			remoteAddr: "127.0.0.1:54321",
			xRealIP:    "203.0.113.7",
			wantIP:     "203.0.113.7",
			wantPort:   54321,
		},
		{
			name:       "loopback peer with no X-Real-IP errors out (does not fall back to 127.0.0.1)",
			remoteAddr: "127.0.0.1:54321",
			xRealIP:    "",
			wantErr:    true,
		},
		{
			name:       "loopback peer with malformed X-Real-IP errors out",
			remoteAddr: "127.0.0.1:54321",
			xRealIP:    "not-an-ip",
			wantErr:    true,
		},
		{
			name:       "IPv6 loopback peer is treated as loopback (X-Real-IP honored)",
			remoteAddr: "[::1]:54321",
			xRealIP:    "203.0.113.8",
			wantIP:     "203.0.113.8",
			wantPort:   54321,
		},
		{
			name:       "non-loopback peer ignores X-Real-IP and returns the direct peer",
			remoteAddr: "198.51.100.5:1234",
			xRealIP:    "203.0.113.99", // attacker setting it directly — must not be trusted
			wantIP:     "198.51.100.5",
			wantPort:   1234,
		},
		{
			name:       "non-loopback peer ignores X-Forwarded-For",
			remoteAddr: "198.51.100.5:1234",
			xff:        "203.0.113.99", // even from XFF — must not be trusted
			wantIP:     "198.51.100.5",
			wantPort:   1234,
		},
		{
			name:       "loopback peer ignores X-Forwarded-For even when X-Real-IP is also set",
			remoteAddr: "127.0.0.1:54321",
			xRealIP:    "203.0.113.7",
			xff:        "1.2.3.4", // attacker-prepended XFF — must be ignored
			wantIP:     "203.0.113.7",
			wantPort:   54321,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{
				RemoteAddr: tt.remoteAddr,
				Header:     http.Header{},
			}
			if tt.xRealIP != "" {
				r.Header.Set("X-Real-IP", tt.xRealIP)
			}
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}

			addr, err := realClientAddr(r)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got addr=%v", addr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if addr.IP.String() != tt.wantIP {
				t.Errorf("IP = %q, want %q", addr.IP.String(), tt.wantIP)
			}
			if addr.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", addr.Port, tt.wantPort)
			}
		})
	}
}

// newTestInstance constructs a clusterInstance with just enough state for
// the pending-map/dispatch tests. Everything else (net, device, HTTP) is
// left zero so these tests stay hermetic.
func newTestInstance() *clusterInstance {
	return &clusterInstance{
		pendingRequests: make(map[uint64]map[string]chan []byte),
	}
}

// newAddrInstance constructs a clusterInstance with a resolvable UDP addr,
// which the resolveTarget tests need so they can match by md.RemoteAddr.
func newAddrInstance(t *testing.T, host string, port, weight int) *clusterInstance {
	t.Helper()
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		t.Fatalf("resolve %s:%d: %v", host, port, err)
	}
	return &clusterInstance{
		host:            host,
		port:            port,
		weight:          weight,
		addr:            addr,
		pendingRequests: make(map[uint64]map[string]chan []byte),
	}
}

// TestPickInstance_SingleInstance: with one instance, every scheme returns
// that instance (no math, no division by zero). This was the phase-1
// behavior and stays the same after lifting the multi-instance limit.
func TestPickInstance_SingleInstance(t *testing.T) {
	for _, scheme := range []LoadBalanceScheme{LBRandom, LBWeightedRandom, LBRoundRobin, "", "garbage"} {
		cr := &clusterRuntime{
			scheme:      scheme,
			instances:   []*clusterInstance{newAddrInstance(t, "10.0.0.1", 62206, 1)},
			totalWeight: 1,
		}
		for i := 0; i < 10; i++ {
			if cr.pickInstance() != cr.instances[0] {
				t.Fatalf("scheme=%q must always return the sole instance", scheme)
			}
		}
	}
}

// TestPickInstance_RoundRobin: round-robin must cycle through all instances
// in order. Deterministic, so we can assert an exact sequence rather than a
// statistical bound.
func TestPickInstance_RoundRobin(t *testing.T) {
	cr := &clusterRuntime{
		scheme: LBRoundRobin,
		instances: []*clusterInstance{
			newAddrInstance(t, "10.0.0.1", 62206, 1),
			newAddrInstance(t, "10.0.0.2", 62206, 1),
			newAddrInstance(t, "10.0.0.3", 62206, 1),
		},
		totalWeight: 3,
	}
	want := []*clusterInstance{cr.instances[0], cr.instances[1], cr.instances[2]}
	// Two full cycles to confirm the cursor wraps cleanly.
	for cycle := 0; cycle < 2; cycle++ {
		for i, w := range want {
			got := cr.pickInstance()
			if got != w {
				t.Fatalf("round-robin cycle %d step %d: got %s, want %s",
					cycle, i, got.addr, w.addr)
			}
		}
	}
}

// TestPickInstance_Random: random scheme must eventually hit every instance.
// We give it a generous budget (1000 picks across 3 instances) so the test
// is effectively impossible to flake on a working implementation — the
// probability of missing one instance for that many picks is < 10^-176.
func TestPickInstance_Random(t *testing.T) {
	cr := &clusterRuntime{
		scheme: LBRandom,
		instances: []*clusterInstance{
			newAddrInstance(t, "10.0.0.1", 62206, 1),
			newAddrInstance(t, "10.0.0.2", 62206, 1),
			newAddrInstance(t, "10.0.0.3", 62206, 1),
		},
		totalWeight: 3,
	}
	hit := make(map[*clusterInstance]int)
	for i := 0; i < 1000; i++ {
		hit[cr.pickInstance()]++
	}
	for i, inst := range cr.instances {
		if hit[inst] == 0 {
			t.Fatalf("random scheme never picked instances[%d]", i)
		}
	}
}

// TestPickInstance_WeightedRandom: with weights [1, 1, 10], instance 2's
// share of picks should be near 10/12 ≈ 83%. We assert it's at least 60%
// — that bound is so loose that any reasonable RNG passes, but a
// fully-broken implementation (e.g. ignoring weights, or off-by-one in the
// cumulative-weight loop) gets caught. The point of the test is the
// invariant "weights actually influence distribution", not exact ratios.
func TestPickInstance_WeightedRandom(t *testing.T) {
	cr := &clusterRuntime{
		scheme: LBWeightedRandom,
		instances: []*clusterInstance{
			newAddrInstance(t, "10.0.0.1", 62206, 1),
			newAddrInstance(t, "10.0.0.2", 62206, 1),
			newAddrInstance(t, "10.0.0.3", 62206, 10),
		},
		totalWeight: 12,
	}
	const samples = 5000
	hit := make(map[*clusterInstance]int)
	for i := 0; i < samples; i++ {
		hit[cr.pickInstance()]++
	}
	share := float64(hit[cr.instances[2]]) / float64(samples)
	if share < 0.60 {
		t.Fatalf("weighted-random did not bias toward the high-weight instance: share=%.3f want >=0.60 (hits=%v)",
			share, hit)
	}
	// All instances must still receive at least *some* picks; weight=1
	// each on the first two should land them well above zero in 5000
	// samples (expected ~417 each).
	for i := 0; i < 2; i++ {
		if hit[cr.instances[i]] == 0 {
			t.Fatalf("weighted-random starved instances[%d]: %v", i, hit)
		}
	}
}

// TestPickInstance_EmptyCluster: 0 instances returns nil so handlers can
// answer 503 without panicking.
func TestPickInstance_EmptyCluster(t *testing.T) {
	cr := &clusterRuntime{scheme: LBRoundRobin}
	if got := cr.pickInstance(); got != nil {
		t.Fatalf("empty cluster must return nil, got %v", got)
	}
}

// TestResolveTarget_UsesRemoteAddr is the new core invariant after the
// phase-2 picker lift: once handleRelay has chosen an instance and pinned
// it on md.RemoteAddr, resolveTarget must return that same instance — NOT
// re-pick. If this ever regresses, the symptom is silent: the response
// from the server's ACK lands on the wrong instance's pendingRequests map
// (because the handler registered on instance A while dispatchSend sent
// the packet from instance B's connection), and the handler times out.
//
// We construct a cluster of 3 instances using LBRandom (the picker that
// would most obviously mis-route on a re-pick) and assert that whichever
// instance we put in md.RemoteAddr is what comes back, every time.
func TestResolveTarget_UsesRemoteAddr(t *testing.T) {
	pubKey, _ := pubKeyForTest()
	cr := &clusterRuntime{
		id:     "cluster-x",
		pubKey: pubKey,
		scheme: LBRandom,
		instances: []*clusterInstance{
			newAddrInstance(t, "10.0.0.1", 62206, 1),
			newAddrInstance(t, "10.0.0.2", 62206, 1),
			newAddrInstance(t, "10.0.0.3", 62206, 1),
		},
		totalWeight: 3,
	}
	rs := &RelayServer{
		clusters: map[string]*clusterRuntime{cr.id: cr},
	}
	// Override the fingerprint indexing: we don't actually compute it
	// from pubKey here, we just register the cluster under its id and
	// monkey-patch the map key to be the fingerprint resolveTarget will
	// compute. Computing it via utils.PubKeyFingerprint keeps the test
	// honest.
	rs.clusters = map[string]*clusterRuntime{fingerprintForTest(pubKey): cr}

	for _, want := range cr.instances {
		md := &core.MsgData{
			PeerPk:     pubKey,
			RemoteAddr: want.addr,
		}
		gotCR, gotInst := rs.resolveTarget(md)
		if gotCR != cr {
			t.Fatalf("resolveTarget returned wrong cluster: got %v want %v", gotCR, cr)
		}
		if gotInst != want {
			t.Fatalf("resolveTarget returned wrong instance for RemoteAddr=%s: got %s want %s",
				want.addr, gotInst.addr, want.addr)
		}
	}
}

// TestResolveTarget_UnknownAddr: an md.RemoteAddr that doesn't match any
// instance in the cluster must return (cr, nil). dispatchSend then logs the
// drop with the actual addr for diagnosis — better than silently routing
// to instances[0].
func TestResolveTarget_UnknownAddr(t *testing.T) {
	pubKey, _ := pubKeyForTest()
	cr := &clusterRuntime{
		id:     "cluster-x",
		pubKey: pubKey,
		scheme: LBRandom,
		instances: []*clusterInstance{
			newAddrInstance(t, "10.0.0.1", 62206, 1),
			newAddrInstance(t, "10.0.0.2", 62206, 1),
		},
		totalWeight: 2,
	}
	rs := &RelayServer{clusters: map[string]*clusterRuntime{fingerprintForTest(pubKey): cr}}

	stranger, _ := net.ResolveUDPAddr("udp", "10.99.99.99:62206")
	md := &core.MsgData{
		PeerPk:     pubKey,
		RemoteAddr: stranger,
	}
	gotCR, gotInst := rs.resolveTarget(md)
	if gotCR != cr {
		t.Fatalf("expected cluster match by pubkey, got %v", gotCR)
	}
	if gotInst != nil {
		t.Fatalf("expected nil instance for unknown addr, got %s", gotInst.addr)
	}
}

// pubKeyForTest returns a fixed 32-byte slice usable as a Curve25519 pubkey
// for resolveTarget tests. The value isn't valid cryptographically — we
// only need its bytes to be stable so utils.PubKeyFingerprint is stable.
func pubKeyForTest() ([]byte, string) {
	pk := make([]byte, 32)
	for i := range pk {
		pk[i] = byte(i + 1)
	}
	return pk, fingerprintForTest(pk)
}

// fingerprintForTest computes the cluster-map key the same way resolveTarget
// computes it from md.PeerPk. The test registers the cluster under this key
// so resolveTarget's lookup succeeds against the bytes from pubKeyForTest.
func fingerprintForTest(pk []byte) string {
	return utils.PubKeyFingerprint(pk)
}

// register mirrors what handleRelay does when a request arrives: insert a
// buffered channel under (counter, realAddr). Returns the channel and a
// cleanup func so callers can undo their registration.
func register(inst *clusterInstance, counter uint64, realAddr string) (chan []byte, func()) {
	ch := make(chan []byte, 1)
	inst.pendingMu.Lock()
	waiters, ok := inst.pendingRequests[counter]
	if !ok {
		waiters = make(map[string]chan []byte)
		inst.pendingRequests[counter] = waiters
	}
	waiters[realAddr] = ch
	inst.pendingMu.Unlock()
	return ch, func() {
		inst.pendingMu.Lock()
		if w, ok := inst.pendingRequests[counter]; ok {
			delete(w, realAddr)
			if len(w) == 0 {
				delete(inst.pendingRequests, counter)
			}
		}
		inst.pendingMu.Unlock()
	}
}

// recvOrTimeout reads from ch or returns false after 50ms. 50ms is plenty for
// a same-process channel send and short enough that a bug (blocked goroutine
// holding the mutex) fails the test quickly.
func recvOrTimeout(t *testing.T, ch chan []byte) ([]byte, bool) {
	t.Helper()
	select {
	case b := <-ch:
		return b, true
	case <-time.After(50 * time.Millisecond):
		return nil, false
	}
}

// TestDispatch_SoleWaiter: the normal path. One client waiting, response
// delivered, map cleared.
func TestDispatch_SoleWaiter(t *testing.T) {
	inst := newTestInstance()
	ch, _ := register(inst, 42, "1.2.3.4:5555")

	delivered, ambiguous := inst.dispatch(42, []byte("ok"))
	if !delivered || ambiguous {
		t.Fatalf("expected delivered=true ambiguous=false, got %v/%v", delivered, ambiguous)
	}
	got, okRecv := recvOrTimeout(t, ch)
	if !okRecv || string(got) != "ok" {
		t.Fatalf("channel did not receive payload: recv=%v bytes=%q", okRecv, got)
	}
	inst.pendingMu.Lock()
	_, stillThere := inst.pendingRequests[42]
	inst.pendingMu.Unlock()
	if stillThere {
		t.Fatalf("pendingRequests[42] should have been cleared after dispatch")
	}
}

// TestDispatch_TwoWaitersDropped: the Finding 5 case. Two clients pick the
// same counter. Neither gets the response; both must time out. This is the
// invariant that prevents a malicious client from stealing a legitimate ACK.
func TestDispatch_TwoWaitersDropped(t *testing.T) {
	inst := newTestInstance()
	chA, _ := register(inst, 99, "10.0.0.1:1111") // legitimate client
	chB, _ := register(inst, 99, "10.0.0.2:2222") // attacker guessing the counter

	delivered, ambiguous := inst.dispatch(99, []byte("ack"))
	if delivered {
		t.Fatalf("ambiguous counter must not deliver to any waiter")
	}
	if !ambiguous {
		t.Fatalf("expected ambiguous=true when 2 waiters share a counter")
	}
	if _, ok := recvOrTimeout(t, chA); ok {
		t.Fatalf("legitimate client must NOT receive hijackable response")
	}
	if _, ok := recvOrTimeout(t, chB); ok {
		t.Fatalf("attacker must NOT receive the legitimate response")
	}
	// Entry is kept so late-arriving duplicate responses don't suddenly
	// become unambiguous and get mis-delivered.
	inst.pendingMu.Lock()
	waiters := inst.pendingRequests[99]
	n := len(waiters)
	inst.pendingMu.Unlock()
	if n != 2 {
		t.Fatalf("expected both waiters still registered, got %d", n)
	}
}

// TestDispatch_UnknownCounter: stale/late responses must be silently dropped
// rather than panicking or blocking.
func TestDispatch_UnknownCounter(t *testing.T) {
	inst := newTestInstance()
	delivered, ambiguous := inst.dispatch(0xdeadbeef, []byte("whatever"))
	if delivered || ambiguous {
		t.Fatalf("unknown counter must return (false,false), got %v/%v", delivered, ambiguous)
	}
}

// TestDispatch_WaiterCleanupReleasesMap: after a handler cleans up (timeout
// or normal completion), the counter entry must disappear so later responses
// with the same counter hit the "unknown" path — not the "sole waiter" path
// with a stale channel.
func TestDispatch_WaiterCleanupReleasesMap(t *testing.T) {
	inst := newTestInstance()
	_, cancelA := register(inst, 7, "1.1.1.1:1")
	cancelA() // simulate handler timeout cleanup

	inst.pendingMu.Lock()
	_, stillThere := inst.pendingRequests[7]
	inst.pendingMu.Unlock()
	if stillThere {
		t.Fatalf("empty waiter map must be removed from pendingRequests")
	}

	// A late response now finds no waiter and returns (false, false).
	delivered, ambiguous := inst.dispatch(7, []byte("late"))
	if delivered || ambiguous {
		t.Fatalf("late response to cleaned-up counter must be ignored, got %v/%v", delivered, ambiguous)
	}
}

// TestDispatch_AfterOneWaiterLeavesStillDispatches: if two waiters race but
// one cleans up (timeout) before the response arrives, the remaining sole
// waiter should still receive the response — ambiguity is only a *current*
// state, not a poison pill.
func TestDispatch_AfterOneWaiterLeavesStillDispatches(t *testing.T) {
	inst := newTestInstance()
	chA, _ := register(inst, 11, "1.1.1.1:1")
	_, cancelB := register(inst, 11, "2.2.2.2:2")
	cancelB() // attacker's request timed out first

	delivered, ambiguous := inst.dispatch(11, []byte("ok"))
	if !delivered || ambiguous {
		t.Fatalf("expected sole remaining waiter to receive: delivered=%v ambiguous=%v", delivered, ambiguous)
	}
	got, ok := recvOrTimeout(t, chA)
	if !ok || string(got) != "ok" {
		t.Fatalf("remaining waiter did not receive payload: recv=%v bytes=%q", ok, got)
	}
}
