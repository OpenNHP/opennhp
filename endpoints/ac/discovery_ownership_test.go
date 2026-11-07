package ac

import (
	"errors"
	"net"
	"sync"
	"testing"
	"testing/synctest"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestDiscoveryResultOwnsItsConnection(t *testing.T) {
	for _, outcome := range []string{"removed-success", "replacement-success", "replacement-failure"} {
		t.Run(outcome, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				a := progressAC()
				peer := &core.UdpPeer{Ip: "127.0.0.1", Port: 62206}
				addr := "127.0.0.1:62206"
				old := &UdpConn{ConnData: &core.ConnectionData{LocalAddr: &net.UDPAddr{}}}
				replacement := &UdpConn{ConnData: &core.ConnectionData{LocalAddr: &net.UDPAddr{}}}
				a.remoteConnectionMap[addr] = old
				state := &serverDiscoveryState{retries: 2, failed: 1}
				quit := make(chan struct{})
				var wg sync.WaitGroup
				wg.Add(1)
				panicked := make(chan any, 1)
				go func() { defer func() { panicked <- recover() }(); a.serverDiscovery(peer, &wg, state, quit) }()
				md := <-a.sendMsgCh
				md.ConnData = old.ConnData // sender assigned the transaction's actual connection
				a.remoteConnectionMutex.Lock()
				if outcome == "removed-success" {
					delete(a.remoteConnectionMap, addr)
				} else {
					a.remoteConnectionMap[addr] = replacement
				}
				a.remoteConnectionMutex.Unlock()
				response := &core.PacketParserData{HeaderType: core.NHP_AAK, BodyMessage: []byte(`{}`)}
				if outcome == "replacement-failure" {
					response.Error = errors.New("old transaction failed")
				}
				md.ResponseMsgCh <- response
				synctest.Wait()
				close(quit)
				wg.Wait()
				if p := <-panicked; p != nil {
					t.Fatalf("late result panicked: %v", p)
				}
				if replacement.connected.Load() {
					t.Error("old success marked replacement connected")
				}
				if outcome != "removed-success" && a.remoteConnectionMap[addr] != replacement {
					t.Error("old failure removed replacement connection")
				}
				if outcome != "replacement-failure" && state.failed == 0 {
					t.Error("stale success cleared failure without a current connection")
				}
			})
		})
	}
}

func TestRetiredConnectionDoesNotRemoveReplacement(t *testing.T) {
	a := progressAC()
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 62206}
	old := &UdpConn{ConnData: &core.ConnectionData{RemoteAddr: addr, TimeoutMs: 60000}}
	replacement := &UdpConn{ConnData: &core.ConnectionData{RemoteAddr: addr}}
	a.remoteConnectionMap[addr.String()] = replacement
	close(a.signals.stop)
	a.wg.Add(1)
	go a.connectionRoutine(old)
	a.wg.Wait()
	if a.remoteConnectionMap[addr.String()] != replacement {
		t.Fatal("retired connection teardown removed its replacement")
	}
}
