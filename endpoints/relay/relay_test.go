package relay

import (
	"net/http"
	"testing"
	"time"
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

// TestPickInstance_Deterministic locks the phase-1 invariant that
// pickInstance returns the same *clusterInstance on every call. The whole
// send pipeline depends on this: handleRelay registers a pending waiter
// on the instance it gets here, and sendMessageRoutine -> resolveTarget
// later calls pickInstance again to choose the connection. If those two
// answers ever diverge, the ACK lands on the "other" instance's conn
// and the handler times out silently.
//
// When phase 2 introduces a non-deterministic picker (random / weighted
// / round-robin), this test WILL fail — and that failure is intentional.
// It forces the contributor to either keep determinism per-request
// (sticky session, request-id hash) or to thread the chosen instance
// through *core.MsgData so resolveTarget stops re-picking. See the
// phase-2 trap notice on (*clusterRuntime).pickInstance.
func TestPickInstance_Deterministic(t *testing.T) {
	cr := &clusterRuntime{
		instances: []*clusterInstance{
			newTestInstance(),
			// The second slot is unreachable in phase 1 (normalize()
			// rejects it). It's here only so the test still has a
			// failure mode when phase 2 lifts that restriction —
			// without this, pickInstance would have nothing to be
			// "non-deterministic about" and a broken phase-2 picker
			// could silently pass.
			newTestInstance(),
		},
	}
	first := cr.pickInstance()
	for i := 0; i < 100; i++ {
		if got := cr.pickInstance(); got != first {
			t.Fatalf("pickInstance not deterministic on call %d: handleRelay and resolveTarget would disagree about the target instance; see the phase-2 trap notice on (*clusterRuntime).pickInstance", i)
		}
	}
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
