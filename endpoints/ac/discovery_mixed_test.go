package ac

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestMixedDNSAndPacketFailures(t *testing.T) {
	for _, dnsFailures := range []int{1, 2, 3, 4} {
		t.Run(fmt.Sprint(dnsFailures), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				a := progressAC()
				state := new(serverDiscoveryState)
				quit := make(chan struct{})
				peer := &core.UdpPeer{Ip: "127.0.0.1", Port: 62206}
				addr := "127.0.0.1:62206"
				var attempts int
				resolved := make(chan net.Addr, 1)
				go func() {
					resolved <- a.waitForDiscoveryAddress(func() (net.Addr, error) {
						attempts++
						if attempts <= dnsFailures {
							return nil, errors.New("DNS failure")
						}
						return &net.UDPAddr{IP: net.ParseIP(peer.Ip), Port: peer.Port}, nil
					}, state, quit)
				}()
				if <-resolved == nil {
					t.Fatal("resolution did not recover")
				}
				if got := atomic.LoadInt32(&state.failed); (got != 0) != (dnsFailures >= 3) {
					t.Fatalf("DNS failure state %d after %d failures", got, dnsFailures)
				}
				var wg sync.WaitGroup
				wg.Add(1)
				go a.serverDiscovery(peer, &wg, state, quit)
				defer func() { close(quit); wg.Wait() }()
				disposed := false
				for failure := 1; failure <= 3; failure++ {
					md := <-a.sendMsgCh
					a.remoteConnectionMutex.Lock()
					conn := a.remoteConnectionMap[addr]
					if conn == nil {
						conn = &UdpConn{ConnData: &core.ConnectionData{LocalAddr: &net.UDPAddr{}}}
						a.remoteConnectionMap[addr] = conn
					}
					md.ConnData = conn.ConnData
					a.remoteConnectionMutex.Unlock()
					md.ResponseMsgCh <- &core.PacketParserData{Error: errors.New("packet timeout")}
					synctest.Wait()
					if got := atomic.LoadInt32(&state.failed); (got != 0) != (dnsFailures+failure >= 3) {
						t.Errorf("failure state %d after %d completed mixed failures", got, dnsFailures+failure)
					}
					a.remoteConnectionMutex.Lock()
					disposed = disposed || a.remoteConnectionMap[addr] == nil
					a.remoteConnectionMutex.Unlock()
					time.Sleep(5 * time.Second)
				}
				if !disposed {
					t.Error("failed connection not disposed within three packet failures")
				}
				md := <-a.sendMsgCh
				a.remoteConnectionMutex.Lock()
				conn := &UdpConn{ConnData: &core.ConnectionData{}}
				a.remoteConnectionMap[addr] = conn
				md.ConnData = conn.ConnData
				a.remoteConnectionMutex.Unlock()
				md.ResponseMsgCh <- &core.PacketParserData{HeaderType: core.NHP_AAK, BodyMessage: []byte(`{}`)}
				synctest.Wait()
				if atomic.LoadInt32(&state.failed) != 0 || !conn.connected.Load() {
					t.Error("current successful response did not recover discovery")
				}
				state.mu.Lock()
				if state.retries != 0 {
					t.Error("successful response did not reset retries")
				}
				state.mu.Unlock()
			})
		})
	}
}
