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

func newPerIPCapTestConn(ip string, port int) *UdpConn {
	return &UdpConn{
		ConnData: &core.ConnectionData{
			RemoteAddr: &net.UDPAddr{IP: net.ParseIP(ip), Port: port},
		},
		evictSignal:   make(chan struct{}),
		timeoutUpdate: make(chan struct{}, 1),
	}
}

func TestAdmitDirectConnectionCapsOneIPAndPreservesTuples(t *testing.T) {
	s := newPerIPCapTestServer()
	const ip = "198.51.100.20"
	conns := make([]*UdpConn, 0, MaxAgentConnectionsPerIP+4)
	for i := 0; i < MaxAgentConnectionsPerIP+4; i++ {
		conn := newPerIPCapTestConn(ip, 30000+i)
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
			conn := newPerIPCapTestConn(ip, 40000+i)
			s.admitDirectConnection(conn, conn.ConnData.RemoteAddr.String())
		}
		if got := s.connectionsByIP[ip].Len(); got != MaxAgentConnectionsPerIP {
			t.Fatalf("bucket %s size = %d", ip, got)
		}
	}
}

func TestPromoteControlConnectionRemovesAuthenticatedTupleFromAgentBucket(t *testing.T) {
	for _, tt := range []struct {
		name        string
		kind        controlConnectionKind
		wantTimeout int
	}{
		{name: "AC", kind: controlConnectionAC, wantTimeout: DefaultACConnectionTimeoutMs},
		{name: "DB", kind: controlConnectionDB, wantTimeout: DefaultDBConnectionTimeoutMs},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := newPerIPCapTestServer()
			conn := newPerIPCapTestConn("203.0.113.30", 50000)
			s.admitDirectConnection(conn, conn.ConnData.RemoteAddr.String())
			if !s.promoteControlConnection(conn.ConnData, tt.kind) {
				t.Fatal("authenticated tuple was not promoted")
			}
			if len(s.connectionsByIP) != 0 {
				t.Fatal("promoted control connection remained in agent bucket")
			}
			if got := conn.timeout(); got != tt.wantTimeout {
				t.Fatalf("timeout = %d, want %d", got, tt.wantTimeout)
			}
			if tt.kind == controlConnectionAC && !conn.isACConnection.Load() {
				t.Fatal("AC promotion flag was not set")
			}
			if tt.kind == controlConnectionDB && !conn.isDBConnection.Load() {
				t.Fatal("DB promotion flag was not set")
			}
		})
	}
}

func TestPromoteControlConnectionRejectsStaleTuple(t *testing.T) {
	s := newPerIPCapTestServer()
	current := newPerIPCapTestConn("203.0.113.31", 50001)
	s.admitDirectConnection(current, current.ConnData.RemoteAddr.String())
	stale := &core.ConnectionData{
		RemoteAddr:       current.ConnData.RemoteAddr,
		SetTimeoutSignal: make(chan struct{}, 1),
	}
	if s.promoteControlConnection(stale, controlConnectionAC) {
		t.Fatal("stale ConnectionData promoted the current tuple")
	}
	if len(s.connectionsByIP) != 1 {
		t.Fatal("stale promotion removed the current agent bucket entry")
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
			conn := newPerIPCapTestConn("198.51.100.90", 10000+i)
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
