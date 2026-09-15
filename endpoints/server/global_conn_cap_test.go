package server

import (
	"strconv"
	"sync"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func newGlobalCapTestServer(t *testing.T) *UdpServer {
	t.Helper()
	privateKey := make([]byte, 32)
	for i := range privateKey {
		privateKey[i] = byte(i + 1)
	}
	device := core.NewDevice(core.NHP_SERVER, privateKey, nil)
	if device == nil {
		t.Fatal("NewDevice returned nil")
	}
	t.Cleanup(device.Stop)
	return &UdpServer{device: device, remoteConnectionMap: make(map[string]*UdpConn)}
}

func fillGlobalCapMap(s *UdpServer, n int) {
	for i := 0; i < n; i++ {
		s.remoteConnectionMap[strconv.Itoa(i)] = nil
	}
}

func TestGlobalCapAdmitsBoundaries(t *testing.T) {
	tests := []struct {
		name         string
		connections  int
		wantAdmit    bool
		wantOverload bool
	}{
		{name: "empty", connections: 0, wantAdmit: true},
		{name: "at overload threshold", connections: OverloadConnectionThreshold, wantAdmit: true},
		{name: "above overload threshold", connections: OverloadConnectionThreshold + 1, wantAdmit: true, wantOverload: true},
		{name: "below cap", connections: MaxConcurrentConnection - 1, wantAdmit: true, wantOverload: true},
		{name: "at cap", connections: MaxConcurrentConnection, wantAdmit: false, wantOverload: true},
		{name: "above cap", connections: MaxConcurrentConnection + 1, wantAdmit: false, wantOverload: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newGlobalCapTestServer(t)
			fillGlobalCapMap(s, tt.connections)
			if got := s.globalCapAdmits(); got != tt.wantAdmit {
				t.Fatalf("globalCapAdmits() = %v, want %v", got, tt.wantAdmit)
			}
			if got := s.device.IsOverload(); got != tt.wantOverload {
				t.Fatalf("device.IsOverload() = %v, want %v", got, tt.wantOverload)
			}
		})
	}
}

func TestGlobalCapAdmitsConcurrentRejectIsConsistent(t *testing.T) {
	s := newGlobalCapTestServer(t)
	fillGlobalCapMap(s, MaxConcurrentConnection)

	const callers = 32
	var wg sync.WaitGroup
	wg.Add(callers)
	for i := 0; i < callers; i++ {
		go func() {
			defer wg.Done()
			if s.globalCapAdmits() {
				t.Error("globalCapAdmits admitted at the global cap")
			}
		}()
	}
	wg.Wait()
	if !s.device.IsOverload() {
		t.Fatal("device overload state was not updated on rejection path")
	}
}
