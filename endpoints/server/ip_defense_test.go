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
	for i := 0; i <= PreCheckThreatCountBeforeBlock; i++ {
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

func TestBlockAddrUsesIPAcrossPorts(t *testing.T) {
	s := &UdpServer{blockAddrMap: make(map[string]*BlockAddr)}
	s.AddBlockAddr(&net.UDPAddr{IP: net.ParseIP("198.51.100.44"), Port: 1000})
	if !s.IsBlockAddr(&net.UDPAddr{IP: net.ParseIP("198.51.100.44"), Port: 65000}) {
		t.Fatal("blocked IP bypassed the block by changing source port")
	}
	if got := len(s.blockAddrMap); got != 1 {
		t.Fatalf("block map has %d entries, want 1 IP-keyed entry", got)
	}
}

func TestKnownRelayPeerIPGate(t *testing.T) {
	s := &UdpServer{relayPeerMap: map[string]*core.UdpPeer{
		"relay": {Ip: "192.0.2.70"},
	}}
	if !s.isKnownRelayPeerIP("192.0.2.70") {
		t.Fatal("configured relay IP was not recognized")
	}
	if s.isKnownRelayPeerIP("192.0.2.71") {
		t.Fatal("unknown source was trusted as a relay")
	}
}

func TestAuthenticatedControlPlaneAddrRequiresRegisteredTuple(t *testing.T) {
	s := &UdpServer{
		acConnectionMap: make(map[string]*ACConn),
		dbConnectionMap: make(map[string]*DBConn),
	}
	acAddr := &net.UDPAddr{IP: net.ParseIP("192.0.2.80"), Port: 20000}
	dbAddr := &net.UDPAddr{IP: net.ParseIP("192.0.2.81"), Port: 20001}
	if s.isAuthenticatedControlPlaneAddr(acAddr) {
		t.Fatal("unregistered tuple was trusted")
	}
	s.acConnectionMap["ac"] = &ACConn{ConnData: &core.ConnectionData{RemoteAddr: acAddr}}
	s.dbConnectionMap["db"] = &DBConn{ConnData: &core.ConnectionData{RemoteAddr: dbAddr}}
	if !s.isAuthenticatedControlPlaneAddr(acAddr) || !s.isAuthenticatedControlPlaneAddr(dbAddr) {
		t.Fatal("authenticated AC/DB tuple was not recognized")
	}
	if s.isAuthenticatedControlPlaneAddr(&net.UDPAddr{IP: acAddr.IP, Port: acAddr.Port + 1}) {
		t.Fatal("different source port inherited control-plane trust")
	}
}

func TestBlockAddressForConnectionUsesRealRelayClient(t *testing.T) {
	relay := &net.UDPAddr{IP: net.ParseIP("192.0.2.90"), Port: 62206}
	realClient := &net.UDPAddr{IP: net.ParseIP("198.51.100.90"), Port: 40000}
	relayed := &UdpConn{ConnData: &core.ConnectionData{RemoteAddr: relay, RealRemoteAddr: realClient}}
	if got := blockAddressForConnection(relayed); got != realClient {
		t.Fatalf("block address = %v, want real client %v", got, realClient)
	}
	direct := &UdpConn{ConnData: &core.ConnectionData{RemoteAddr: realClient}}
	if got := blockAddressForConnection(direct); got != realClient {
		t.Fatalf("direct block address = %v, want remote %v", got, realClient)
	}
}
