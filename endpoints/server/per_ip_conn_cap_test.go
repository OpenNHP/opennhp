package server

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func newPerIPCapTestServer() *UdpServer {
	return &UdpServer{
		remoteConnectionMap: make(map[string]*UdpConn),
		connectionsByIP:     make(map[string]int),
	}
}

func newPerIPCapTestConn(ip string, port int) *UdpConn {
	return &UdpConn{
		ConnData: &core.ConnectionData{
			RemoteAddr: &net.UDPAddr{IP: net.ParseIP(ip), Port: port},
		},
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

	if got := s.connectionsByIP[ip]; got != MaxAgentConnectionsPerIP {
		t.Fatalf("count = %d", got)
	}
	if len(s.remoteConnectionMap) != MaxAgentConnectionsPerIP {
		t.Fatal("global table exceeded per-IP cap")
	}
	for i := 0; i < MaxAgentConnectionsPerIP; i++ {
		if s.remoteConnectionMap[conns[i].ConnData.RemoteAddr.String()] != conns[i] {
			t.Fatal("existing tuple displaced")
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
		if got := s.connectionsByIP[ip]; got != MaxAgentConnectionsPerIP {
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
	const attempts = MaxAgentConnectionsPerIP * 2
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
	if got := s.connectionsByIP["198.51.100.90"]; got != MaxAgentConnectionsPerIP {
		t.Fatalf("concurrent bucket size = %d, want %d", got, MaxAgentConnectionsPerIP)
	}
	if got := len(s.remoteConnectionMap); got != MaxAgentConnectionsPerIP {
		t.Fatalf("global tuple map size = %d, want %d", got, MaxAgentConnectionsPerIP)
	}
}

func TestConfiguredPerIPCapAndRelease(t *testing.T) {
	s := newPerIPCapTestServer()
	s.config = &Config{MaxAgentConnectionsPerIP: 1}
	first := newPerIPCapTestConn("192.0.2.1", 1000)
	second := newPerIPCapTestConn("192.0.2.1", 1001)
	if !s.admitDirectConnection(first, first.ConnData.RemoteAddr.String()) {
		t.Fatal("first refused")
	}
	if s.admitDirectConnection(second, second.ConnData.RemoteAddr.String()) {
		t.Fatal("excess admitted")
	}
	s.remoteConnectionMapMutex.Lock()
	s.releasePerIPCount(first)
	s.remoteConnectionMapMutex.Unlock()
	if !s.admitDirectConnection(second, second.ConnData.RemoteAddr.String()) {
		t.Fatal("released slot not reused")
	}
}

func TestOnlineHandlerRejectsWrongPeerRole(t *testing.T) {
	for _, handler := range []string{"AC", "DB"} {
		s := newPerIPCapTestServer()
		conn := newPerIPCapTestConn("192.0.2.1", 1234)
		s.admitDirectConnection(conn, conn.ConnData.RemoteAddr.String())
		ppd := &core.PacketParserData{ConnData: conn.ConnData, BodyMessage: []byte(`{}`), RemotePubKey: make([]byte, core.PublicKeySize)}
		var err error
		if handler == "AC" {
			err = s.HandleACOnline(ppd)
		} else {
			err = s.HandleDBOnline(ppd)
		}
		if err == nil || !conn.perIPCounted {
			t.Fatalf("%s gained a control exemption", handler)
		}
	}
}

func TestDirectConnectionHonorsSetTimeout(t *testing.T) {
	for _, timeout := range []int{0, 20} {
		s := newGlobalCapTestServer(t)
		s.connectionsByIP = make(map[string]int)
		s.signals.stop = make(chan struct{})
		conn := newPerIPCapTestConn("192.0.2.1", 1234)
		conn.ConnData.Device = s.device
		conn.ConnData.TimeoutMs = 30000
		conn.timeoutMs.Store(30000)
		conn.ConnData.SendQueue = make(chan *core.Packet)
		conn.ConnData.RecvQueue = make(chan *core.Packet)
		conn.ConnData.BlockSignal = make(chan struct{})
		conn.ConnData.StopSignal = make(chan struct{})
		conn.ConnData.SetTimeoutSignal = make(chan struct{})
		s.admitDirectConnection(conn, conn.ConnData.RemoteAddr.String())
		done := make(chan struct{})
		s.wg.Add(1)
		go func() { s.connectionRoutine(conn); close(done) }()
		conn.ConnData.SetTimeout(timeout)
		select {
		case <-done:
		case <-time.After(time.Second):
			close(s.signals.stop)
			<-done
			t.Fatalf("SetTimeout(%d) ignored", timeout)
		}
	}
}
