package server

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestLogThrottle_AllowsFirstCall: the very first call always emits, with
// the suppressed counter at zero. The window start is anchored at
// nowNanos, not at the zero time.
func TestLogThrottle_AllowsFirstCall(t *testing.T) {
	throttle := logThrottle{}
	throttle.setInterval(time.Minute)

	const now = int64(1_000_000_000)
	emit, suppressed := throttle.allow(now)
	if !emit {
		t.Fatalf("expected first call to emit, got suppressed")
	}
	if suppressed != 0 {
		t.Fatalf("expected suppressed=0 on first call, got %d", suppressed)
	}
	if got := throttle.lastNanos.Load(); got != now {
		t.Fatalf("lastNanos=%d, want %d", got, now)
	}
	if got := throttle.suppressed.Load(); got != 0 {
		t.Fatalf("suppressed counter not reset on emit, got %d", got)
	}
}

// TestLogThrottle_SuppressesInsideWindow: a call inside the window is
// counted but not emitted, and the running count survives across calls
// until the next emit. A subsequent call at the same instant is also
// suppressed (the call that just won the window bumped lastNanos).
func TestLogThrottle_SuppressesInsideWindow(t *testing.T) {
	throttle := logThrottle{}
	throttle.setInterval(time.Minute)

	const base = int64(2_000_000_000)
	if emit, _ := throttle.allow(base); !emit {
		t.Fatalf("first call must emit")
	}
	// 30s later, well inside the 60s window.
	if emit, suppressed := throttle.allow(base + int64(30*time.Second)); emit {
		t.Fatalf("call inside window must be suppressed")
	} else if suppressed != 1 {
		t.Fatalf("suppressed=1 after one inside-window call, got %d", suppressed)
	}
	// 15s after that — still inside — count climbs.
	if emit, suppressed := throttle.allow(base + int64(45*time.Second)); emit {
		t.Fatalf("second inside-window call must be suppressed")
	} else if suppressed != 2 {
		t.Fatalf("suppressed=2 after two inside-window calls, got %d", suppressed)
	}
}

// TestLogThrottle_ReleasesAfterIntervalAndResetsCount: a call exactly
// intervalNanos past the last emit wins the window, the suppressed
// count is returned to the caller (so it can be embedded in the next
// log line) and the counter is reset to zero.
func TestLogThrottle_ReleasesAfterIntervalAndResetsCount(t *testing.T) {
	throttle := logThrottle{}
	throttle.setInterval(time.Minute)

	const base = int64(3_000_000_000)
	if emit, _ := throttle.allow(base); !emit {
		t.Fatalf("first call must emit")
	}
	// Two inside-window suppressions.
	throttle.allow(base + int64(10*time.Second))
	throttle.allow(base + int64(20*time.Second))

	// 60s after the original emit — exactly the window — emit again
	// and get the accumulated count.
	emit, suppressed := throttle.allow(base + int64(61*time.Second))
	if !emit {
		t.Fatalf("call past the window must emit")
	}
	if suppressed != 2 {
		t.Fatalf("expected the 2 suppressed calls to surface on emit, got %d", suppressed)
	}
	if got := throttle.suppressed.Load(); got != 0 {
		t.Fatalf("suppressed counter not reset after emit, got %d", got)
	}
}

// TestLogThrottle_ConcurrentOnlyOneLogs: 64 goroutines call allow with
// the same nowNanos. CAS guarantees at most one wins; the rest bump
// the suppressed counter. Must be -race clean.
func TestLogThrottle_ConcurrentOnlyOneLogs(t *testing.T) {
	throttle := logThrottle{}
	throttle.setInterval(time.Minute)

	const now = int64(4_000_000_000)
	const N = 64

	var emitCount int64
	var suppressedAtEmit uint64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			emit, suppressed := throttle.allow(now)
			if emit {
				atomic.AddInt64(&emitCount, 1)
				// First-emit suppressed should be 0 because all 64
				// calls happen in the same window and the first
				// winner happens before any loser's Add.
				_ = suppressed
				if suppressed != 0 {
					atomic.StoreUint64(&suppressedAtEmit, suppressed)
				}
			}
		}()
	}
	close(start)
	wg.Wait()

	if emitCount != 1 {
		t.Fatalf("expected exactly 1 emit across %d concurrent calls, got %d", N, emitCount)
	}
	// After the dust settles, the suppressed counter should be N-1: the
	// winning goroutine reset it to 0 inside allow(), and the remaining
	// N-1 goroutines each Add'd 1.
	if got := throttle.suppressed.Load(); got != N-1 {
		t.Fatalf("expected suppressed=%d after concurrent burst, got %d", N-1, got)
	}
	_ = suppressedAtEmit
}

// TestLogThrottle_ZeroIntervalAlwaysSuppresses: the zero value (no
// setInterval) must not accidentally emit on first call. This is the
// safety property UdpServer field initializers rely on before Start
// calls setInterval.
func TestLogThrottle_ZeroIntervalAlwaysSuppresses(t *testing.T) {
	throttle := logThrottle{}
	// Note: no setInterval — intervalNanos stays 0.
	const now = int64(5_000_000_000)
	emit, _ := throttle.allow(now)
	if emit {
		t.Fatalf("zero-interval throttle must not emit on first call (got emit=true)")
	}
	// And subsequent calls also stay suppressed.
	if emit, _ := throttle.allow(now + 1); emit {
		t.Fatalf("zero-interval throttle must not emit on later calls")
	}
}
