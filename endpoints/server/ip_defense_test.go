package server

import (
	"bytes"
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

func TestIPRateLimiterAdmitsAtCapacityAndReportsBudgetLoss(t *testing.T) {
	r := newIPRateLimiter(1, 2, 1, int64(time.Minute))
	if !r.allow("192.0.2.1", 0) || !r.allow("192.0.2.2", 0) {
		t.Fatal("source admission failed")
	}
	if r.len() != 1 || r.capacityEvictions.Load() != 1 {
		t.Fatal("capacity or eviction count incorrect")
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

func TestRelayRateNamespaceDoesNotDrainDirectSource(t *testing.T) {
	r := newIPRateLimiter(1, 2, 8, int64(time.Minute))
	relay := "rly|198.51.100.1:62206|192.0.2.1"
	if !r.allow(relay, 0) || r.allow(relay, 0) {
		t.Fatal("relay budget incorrect")
	}
	if !r.allow("192.0.2.1", 0) {
		t.Fatal("relay drained direct source")
	}
}

func TestMalformedDatagramsConsumeSourceBudget(t *testing.T) {
	for _, size := range []int{1, 256} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			s := newGlobalCapTestServer(t)
			listener, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
			if err != nil {
				t.Fatal(err)
			}
			s.listenConn = listener
			s.listenAddr = listener.LocalAddr().(*net.UDPAddr)
			s.packetLimiter = newIPRateLimiter(1, 2, 8, int64(time.Hour))
			s.packetLimiter.nanosPerToken = int64(time.Hour)
			s.packetLimiter.burstNanos = int64(2 * time.Hour)
			s.wg.Add(1)
			go s.recvPacketRoutine()
			t.Cleanup(func() { listener.Close(); s.wg.Wait() })
			sender, err := net.DialUDP("udp4", nil, s.listenAddr)
			if err != nil {
				t.Fatal(err)
			}
			defer sender.Close()
			payload := bytes.Repeat([]byte{0xff}, size)
			payload[0] = 0 // XOR header words now encode an unsupported message type.
			for i := 0; i < 2; i++ {
				if _, err := sender.Write(payload); err != nil {
					t.Fatal(err)
				}
			}
			deadline := time.Now().Add(3 * time.Second)
			for s.packetRateLimitDrops.Load() == 0 && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			if s.packetRateLimitDrops.Load() != 1 || s.malformedPacketDrops.Load() != 1 {
				t.Fatalf("rate drops=%d malformed=%d", s.packetRateLimitDrops.Load(), s.malformedPacketDrops.Load())
			}
		})
	}
}
