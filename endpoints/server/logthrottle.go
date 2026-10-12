package server

import (
	"sync/atomic"
	"time"
)

// logThrottle caps a per-event log line to at most one output per
// intervalNanos, while still counting the suppressed occurrences so the
// next emitted line can surface them as "(N suppressed since last line)".
// Designed for the per-packet Critical log that would otherwise be
// attacker-rate-bound (HandleRelayForward's per-(relay, real-client) block
// gate is the canonical example).
//
// The clock is injected via nowNanos so tests are deterministic. Production
// callers pass time.Now().UnixNano() — the existing rknRateLimiter.allow
// (ratelimit.go) uses the same convention.
//
// All operations are lock-free atomics so the hot path on
// HandleRelayForward's blocked branch does not contend. Under contention
// CompareAndSwap guarantees exactly one goroutine wins a window; the
// losers fall into the suppressed counter.
//
// Zero value: intervalNanos==0 → allow() always returns false. Callers
// that want emissions must call setInterval during construction (the
// UdpServer field is initialized in Start with a 1-minute window). The
// type is safe for concurrent use by multiple goroutines.
type logThrottle struct {
	intervalNanos int64
	lastNanos     atomic.Int64
	suppressed    atomic.Uint64
}

// allow returns (emit, suppressedCount) where emit=true means the caller
// just won the window and should emit a log line. The first call after
// construction (or after intervalNanos has elapsed) emits; subsequent
// calls inside the window are counted in the suppressed counter but
// suppressed from the log.
//
// When emit is true, suppressedCount is the number of suppressed events
// since the previous emit (zero on the very first call). When emit is
// false, suppressedCount is the running total of suppressed events
// including this one.
func (t *logThrottle) allow(nowNanos int64) (bool, uint64) {
	last := t.lastNanos.Load()
	if last == 0 || nowNanos-last >= t.intervalNanos {
		if t.lastNanos.CompareAndSwap(last, nowNanos) {
			suppressed := t.suppressed.Swap(0)
			return true, suppressed
		}
	}
	n := t.suppressed.Add(1)
	return false, n
}

// setInterval configures the throttle window. Intended for one-shot use
// during construction (e.g. from (*UdpServer).Start); not concurrency-safe
// — must not race with allow().
func (t *logThrottle) setInterval(d time.Duration) {
	t.intervalNanos = int64(d)
}
