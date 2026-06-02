package server

import (
	"fmt"
	"net"
	"testing"
)

func addr(ip string) *net.UDPAddr {
	return &net.UDPAddr{IP: net.ParseIP(ip), Port: 1234}
}

const sec = int64(1_000_000_000)

// TestRknRateLimiter_BurstThenThrottle: a fresh IP gets exactly `burst`
// immediate allows (bucket starts full), then is throttled until tokens
// accrue.
func TestRknRateLimiter_BurstThenThrottle(t *testing.T) {
	rl := newRknRateLimiter(10 /*rate*/, 5 /*burst*/, 1024, 300*sec)
	now := int64(0)
	a := addr("203.0.113.1")

	allowed := 0
	for i := 0; i < 5; i++ {
		if rl.allow(a, now) {
			allowed++
		}
	}
	if allowed != 5 {
		t.Fatalf("fresh bucket should allow the full burst of 5; got %d", allowed)
	}
	// 6th at the same instant must be throttled (bucket empty).
	if rl.allow(a, now) {
		t.Fatal("6th packet at t=0 should be throttled")
	}
	// rate=10/s -> one token every 100ms. Advance 100ms: exactly one more.
	now += sec / 10
	if !rl.allow(a, now) {
		t.Fatal("after 100ms one token should have accrued")
	}
	if rl.allow(a, now) {
		t.Fatal("only one token should have accrued in 100ms")
	}
}

// TestRknRateLimiter_PerIPIsolation: throttling one IP must not affect
// another.
func TestRknRateLimiter_PerIPIsolation(t *testing.T) {
	rl := newRknRateLimiter(10, 2, 1024, 300*sec)
	now := int64(0)
	attacker := addr("198.51.100.7")
	victim := addr("203.0.113.9")

	// Drain the attacker's bucket.
	for i := 0; i < 10; i++ {
		rl.allow(attacker, now)
	}
	if rl.allow(attacker, now) {
		t.Fatal("attacker should be throttled after draining its bucket")
	}
	// Victim, never seen before, still gets its full burst.
	if !rl.allow(victim, now) || !rl.allow(victim, now) {
		t.Fatal("victim's bucket must be independent of the attacker's")
	}
}

// TestRknRateLimiter_PortIgnored: the same IP on different ports shares one
// bucket — an attacker rotating source ports must not multiply its budget.
func TestRknRateLimiter_PortIgnored(t *testing.T) {
	rl := newRknRateLimiter(10, 3, 1024, 300*sec)
	now := int64(0)
	ip := net.ParseIP("198.51.100.42")

	allowed := 0
	for port := 1000; port < 1100; port++ {
		if rl.allow(&net.UDPAddr{IP: ip, Port: port}, now) {
			allowed++
		}
	}
	if allowed != 3 {
		t.Fatalf("rotating ports on one IP must share one burst of 3; got %d", allowed)
	}
}

// TestRknRateLimiter_NilAddrFailsClosed: an unresolvable source is dropped,
// not admitted.
func TestRknRateLimiter_NilAddrFailsClosed(t *testing.T) {
	rl := newRknRateLimiter(10, 5, 1024, 300*sec)
	if rl.allow(nil, 0) {
		t.Fatal("nil addr must fail closed (drop)")
	}
	if rl.allow(&net.UDPAddr{IP: nil, Port: 80}, 0) {
		t.Fatal("addr with nil IP must fail closed (drop)")
	}
}

// TestRknRateLimiter_BudgetCappedAtBurst: a long idle period must not let a
// bucket accrue more than `burst` tokens (no credit hoarding).
func TestRknRateLimiter_BudgetCappedAtBurst(t *testing.T) {
	rl := newRknRateLimiter(10, 4, 1024, 3600*sec)
	a := addr("203.0.113.50")

	// First touch at t=0 (bucket full = 4), drain it.
	now := int64(0)
	for i := 0; i < 4; i++ {
		rl.allow(a, now)
	}
	// Idle for an hour, then hammer: must get at most `burst` immediate
	// allows, not 36000.
	now += 3600 * sec
	allowed := 0
	for i := 0; i < 100; i++ {
		if rl.allow(a, now) {
			allowed++
		}
	}
	if allowed != 4 {
		t.Fatalf("accrued budget must cap at burst=4 after long idle; got %d", allowed)
	}
}

// TestRknRateLimiter_EvictionBoundsMap: the bucket map never exceeds
// maxEntries even under a flood of distinct source IPs.
func TestRknRateLimiter_EvictionBoundsMap(t *testing.T) {
	const maxEntries = 100
	rl := newRknRateLimiter(10, 5, maxEntries, 300*sec)
	now := int64(0)

	// 10x maxEntries distinct IPs, all "active" (same instant, so the
	// idle-sweep frees nothing and oldest-eviction must kick in).
	for i := 0; i < maxEntries*10; i++ {
		rl.allow(addr(fmt.Sprintf("10.%d.%d.%d", (i>>16)&0xff, (i>>8)&0xff, i&0xff)), now)
		if len(rl.buckets) > maxEntries {
			t.Fatalf("bucket map exceeded maxEntries=%d: got %d at i=%d", maxEntries, len(rl.buckets), i)
		}
	}
}

// TestRknRateLimiter_IdleSweepReclaims: idle buckets are reclaimed by the
// sweep before oldest-eviction is needed.
func TestRknRateLimiter_IdleSweepReclaims(t *testing.T) {
	const maxEntries = 10
	rl := newRknRateLimiter(10, 5, maxEntries, 100*sec)
	now := int64(0)

	// Fill with maxEntries IPs at t=0.
	for i := 0; i < maxEntries; i++ {
		rl.allow(addr(fmt.Sprintf("172.16.0.%d", i)), now)
	}
	if len(rl.buckets) != maxEntries {
		t.Fatalf("expected %d buckets, got %d", maxEntries, len(rl.buckets))
	}
	// Advance past idleNanos so every existing bucket is sweepable, then a
	// new IP triggers eviction: the sweep should clear the stale ones.
	now += 101 * sec
	rl.allow(addr("172.16.0.200"), now)
	if len(rl.buckets) != 1 {
		t.Fatalf("idle sweep should have reclaimed all stale buckets, leaving only the new one; got %d", len(rl.buckets))
	}
}
