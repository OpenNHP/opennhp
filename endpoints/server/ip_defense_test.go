package server

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestIPRateLimiterFreshIPStartsAtHalfBurst(t *testing.T) {
	r := newIPRateLimiter(1, 10, 100, int64(time.Minute))
	allowed := 0
	for i := 0; i < 10; i++ {
		if r.allow("198.51.100.10", 0) {
			allowed++
		}
	}
	if allowed != 5 {
		t.Fatalf("allowed %d fresh-IP packets, want half-burst 5", allowed)
	}
}

func TestIPRateLimiterPortRotationSharesBudget(t *testing.T) {
	r := newIPRateLimiter(1, 4, 100, int64(time.Minute))
	ip := net.ParseIP("203.0.113.8").String()
	if !r.allow(ip, 0) || !r.allow(ip, 0) {
		t.Fatal("starter packets should be admitted")
	}
	if r.allow(ip, 0) {
		t.Fatal("same IP should not gain a new budget by rotating ports")
	}
}

func TestIPRateLimiterCapacityIsBounded(t *testing.T) {
	r := newIPRateLimiter(100, 10, 8, int64(time.Hour))
	for i := 0; i < 100; i++ {
		r.allow(fmt.Sprintf("192.0.2.%d", i), int64(i))
		if got := r.len(); got > 8 {
			t.Fatalf("limiter grew to %d entries, want <= 8", got)
		}
	}
}

func TestIPRateLimiterConcurrentAccess(t *testing.T) {
	r := newIPRateLimiter(1000, 100, 32, int64(time.Minute))
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ip := fmt.Sprintf("198.51.100.%d", i)
			for j := 0; j < 100; j++ {
				r.allow(ip, int64(j)*int64(time.Millisecond))
			}
		}(i)
	}
	wg.Wait()
	if got := r.len(); got != 32 {
		t.Fatalf("limiter has %d entries, want 32", got)
	}
}

func TestPreCheckThreatCacheIsIPKeyedAndBounded(t *testing.T) {
	c := newPreCheckThreatCache(4, int64(time.Minute))
	for i := 0; i < 6; i++ {
		if got := c.increment("203.0.113.10", int64(i)); got != int32(i+1) {
			t.Fatalf("increment %d returned %d", i, got)
		}
	}
	for i := 0; i < 20; i++ {
		c.increment(fmt.Sprintf("192.0.2.%d", i), int64(i))
	}
	if got := c.len(); got != 4 {
		t.Fatalf("cache has %d entries, want hard cap 4", got)
	}
}

func TestBlockAddrDoesNotBlockOtherPorts(t *testing.T) {
	s := &UdpServer{blockAddrMap: make(map[string]*BlockAddr)}
	addr := &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 1000}
	s.AddBlockAddr(addr)
	if !s.IsBlockAddr(addr) {
		t.Fatal("tuple not blocked")
	}
	if s.IsBlockAddr(&net.UDPAddr{IP: addr.IP, Port: 1001}) {
		t.Fatal("unrelated tuple blocked")
	}
}

func TestAuthenticatedInfrastructureExemptionUsesExactConnection(t *testing.T) {
	s := &UdpServer{remoteConnectionMap: make(map[string]*UdpConn)}
	addr := &net.UDPAddr{IP: net.ParseIP("192.0.2.1"), Port: 1000}
	conn := &UdpConn{ConnData: &core.ConnectionData{RemoteAddr: addr}}
	s.remoteConnectionMap[addr.String()] = conn
	if s.isAuthenticatedControlPlaneAddr(addr) {
		t.Fatal("unauthenticated exempt")
	}
	s.markRateExempt(&core.ConnectionData{RemoteAddr: addr})
	if s.isAuthenticatedControlPlaneAddr(addr) {
		t.Fatal("stale connection exempt")
	}
	s.markRateExempt(conn.ConnData)
	if !s.isAuthenticatedControlPlaneAddr(addr) {
		t.Fatal("authenticated peer not exempt")
	}
	delete(s.remoteConnectionMap, addr.String())
	if s.isAuthenticatedControlPlaneAddr(addr) {
		t.Fatal("exemption survived teardown")
	}
}

func TestRelayCannotBlockClaimedClient(t *testing.T) {
	relay := &net.UDPAddr{IP: net.ParseIP("192.0.2.90"), Port: 62206}
	client := &net.UDPAddr{IP: net.ParseIP("198.51.100.90"), Port: 40000}
	if blockAddressForConnection(&UdpConn{ConnData: &core.ConnectionData{RemoteAddr: relay, RealRemoteAddr: client}}) != nil {
		t.Fatal("relay can block third-party address")
	}
}

func TestIPRateLimiterRetainsDrainedBudgetAtCapacity(t *testing.T) {
	r := newIPRateLimiter(1, 2, 1, int64(time.Minute))
	if !r.allow("192.0.2.1", 0) {
		t.Fatal("first packet rejected")
	}
	if r.allow("192.0.2.2", 0) {
		t.Fatal("new source evicted drained budget")
	}
	if r.allow("192.0.2.1", 0) {
		t.Fatal("drained source regained budget")
	}
	if !r.allow("192.0.2.2", int64(2*time.Second)) {
		t.Fatal("replenished entry was not reused")
	}
}

func TestIPRateLimiterSkipsDrainedEntryForReplenishedEntry(t *testing.T) {
	r := newIPRateLimiter(1, 4, 2, int64(time.Minute))
	r.allow("192.0.2.1", 0)
	r.allow("192.0.2.1", 0)
	r.allow("192.0.2.2", 0)
	if !r.allow("192.0.2.3", int64(3*time.Second)) {
		t.Fatal("replenished entry behind drained entry was not reused")
	}
	if _, ok := r.buckets["192.0.2.1"]; !ok {
		t.Fatal("drained entry was evicted")
	}
}
