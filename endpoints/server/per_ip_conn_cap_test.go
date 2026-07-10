package server

import (
	"container/list"
	"net"
	"sync"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func newPerIPCapTestServer() *UdpServer {
	return &UdpServer{
		remoteConnectionMap: make(map[string]*UdpConn),
		connectionsByIP:     make(map[string]*list.List),
	}
}

func newPerIPCapTestConn(ip string, port int, ac, db bool) *UdpConn {
	return &UdpConn{
		ConnData: &core.ConnectionData{
			RemoteAddr: &net.UDPAddr{IP: net.ParseIP(ip), Port: port},
		},
		isACConnection: ac,
		isDBConnection: db,
		evictSignal:    make(chan struct{}),
	}
}

func TestAdmitDirectConnectionCapsOneIPAndPreservesTuples(t *testing.T) {
	s := newPerIPCapTestServer()
	const ip = "198.51.100.20"
	conns := make([]*UdpConn, 0, MaxAgentConnectionsPerIP+4)
	for i := 0; i < MaxAgentConnectionsPerIP+4; i++ {
		conn := newPerIPCapTestConn(ip, 30000+i, false, false)
		conns = append(conns, conn)
		s.admitDirectConnection(conn, conn.ConnData.RemoteAddr.String())
	}

	bucket := s.connectionsByIP[ip]
	if bucket == nil || bucket.Len() != MaxAgentConnectionsPerIP {
		t.Fatalf("bucket size = %v, want %d", bucket, MaxAgentConnectionsPerIP)
	}
	for i := 0; i < 4; i++ {
		select {
		case <-conns[i].evictSignal:
		default:
			t.Fatalf("oldest connection %d was not evicted", i)
		}
	}
	for i := 4; i < len(conns); i++ {
		key := conns[i].ConnData.RemoteAddr.String()
		if s.remoteConnectionMap[key] != conns[i] {
			t.Fatalf("NAT reply tuple %s was not preserved in global map", key)
		}
	}
}

func TestAdmitDirectConnectionIPsAreIndependent(t *testing.T) {
	s := newPerIPCapTestServer()
	for _, ip := range []string{"192.0.2.10", "192.0.2.11"} {
		for i := 0; i < MaxAgentConnectionsPerIP; i++ {
			conn := newPerIPCapTestConn(ip, 40000+i, false, false)
			s.admitDirectConnection(conn, conn.ConnData.RemoteAddr.String())
		}
		if got := s.connectionsByIP[ip].Len(); got != MaxAgentConnectionsPerIP {
			t.Fatalf("bucket %s size = %d", ip, got)
		}
	}
}

func TestAdmitDirectConnectionTrustedInfrastructureBypassesAgentCap(t *testing.T) {
	for _, kind := range []struct {
		name string
		ac   bool
		db   bool
	}{{name: "AC", ac: true}, {name: "DB", db: true}} {
		t.Run(kind.name, func(t *testing.T) {
			s := newPerIPCapTestServer()
			for i := 0; i < MaxAgentConnectionsPerIP+2; i++ {
				conn := newPerIPCapTestConn("203.0.113.30", 50000+i, kind.ac, kind.db)
				s.admitDirectConnection(conn, conn.ConnData.RemoteAddr.String())
			}
			if len(s.connectionsByIP) != 0 {
				t.Fatal("trusted infrastructure connection entered agent bucket")
			}
		})
	}
}

func TestAdmitDirectConnectionConcurrentCap(t *testing.T) {
	s := newPerIPCapTestServer()
	const attempts = 128
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn := newPerIPCapTestConn("198.51.100.90", 10000+i, false, false)
			s.admitDirectConnection(conn, conn.ConnData.RemoteAddr.String())
		}(i)
	}
	wg.Wait()
	if got := s.connectionsByIP["198.51.100.90"].Len(); got != MaxAgentConnectionsPerIP {
		t.Fatalf("concurrent bucket size = %d, want %d", got, MaxAgentConnectionsPerIP)
	}
	if got := len(s.remoteConnectionMap); got != attempts {
		t.Fatalf("global tuple map size = %d, want %d pending asynchronous cleanup", got, attempts)
	}
}

func TestKnownPeerIPGate(t *testing.T) {
	s := newPerIPCapTestServer()
	s.acPeerMap = map[string]*core.UdpPeer{"ac": {Ip: "192.0.2.50"}}
	s.dbPeerMap = map[string]*core.UdpPeer{"db": {Ip: "192.0.2.51"}}
	if !s.isKnownACPeerIP("192.0.2.50") || s.isKnownACPeerIP("192.0.2.99") {
		t.Fatal("AC source-IP gate mismatch")
	}
	if !s.isKnownDBPeerIP("192.0.2.51") || s.isKnownDBPeerIP("192.0.2.99") {
		t.Fatal("DB source-IP gate mismatch")
	}
}
