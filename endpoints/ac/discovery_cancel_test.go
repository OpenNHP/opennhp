package ac

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestDiscoveryCancellation(t *testing.T) {
	for _, phase := range []string{"send", "response", "keepalive"} {
		for _, signal := range []string{"reload", "stop"} {
			t.Run(phase+"/"+signal, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					peer := &core.UdpPeer{Ip: "127.0.0.1", Port: 62206}
					a := &UdpAC{config: &Config{}, device: &core.Device{}, sendMsgCh: make(chan *core.MsgData), remoteConnectionMap: make(map[string]*UdpConn)}
					a.running.Store(true)
					a.signals.stop = make(chan struct{})
					if phase == "keepalive" {
						conn := &UdpConn{ConnData: &core.ConnectionData{LastLocalRecvTime: time.Now().UnixNano(), LastLocalSendTime: time.Now().Add(-2 * ServerKeepaliveInterval * time.Second).UnixNano()}}
						conn.connected.Store(true)
						a.remoteConnectionMap[net.JoinHostPort(peer.Ip, "62206")] = conn
					}
					quit := make(chan struct{})
					var wg sync.WaitGroup
					var failed serverDiscoveryState
					wg.Add(1)
					done := make(chan struct{})
					go func() { a.serverDiscovery(peer, &wg, &failed, quit); close(done) }()
					var request *core.MsgData
					if phase == "response" {
						request = <-a.sendMsgCh
					}
					// Wait until discovery is blocked on the exact send/receive under test.
					synctest.Wait()
					if signal == "reload" {
						close(quit)
					} else {
						close(a.signals.stop)
					}
					synctest.Wait()
					select {
					case <-done:
					default:
						t.Error("discovery ignored cancellation")
						// Release the old implementation so a failing regression leaves no goroutine behind.
						if request == nil {
							request = <-a.sendMsgCh
						}
						if request.ResponseMsgCh != nil {
							request.ResponseMsgCh <- &core.PacketParserData{Error: errors.New("test cleanup")}
						}
						<-done
						return
					}
					if request != nil {
						// The transaction owns one eventual reply, even after discovery exits.
						select {
						case request.ResponseMsgCh <- &core.PacketParserData{Error: errors.New("late reply")}:
						default:
							t.Error("late transaction reply would block its producer")
						}
					}
					wg.Wait()
				})
			})
		}
	}
}

func TestDiscoveryReloadWithPendingResponse(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := &UdpAC{config: &Config{FilterMode: FilterMode_EBPFXDP}, device: &core.Device{}, sendMsgCh: make(chan *core.MsgData), serverPeerMap: map[string]*core.UdpPeer{"old": {Ip: "127.0.0.1", Port: 62206}}}
		a.running.Store(true)
		a.signals.stop = make(chan struct{})
		a.signals.serverMapUpdated = make(chan struct{}, 1)
		a.wg.Add(1)
		go a.maintainServerConnectionRoutine()
		old := <-a.sendMsgCh
		a.serverPeerMutex.Lock()
		a.serverPeerMap = map[string]*core.UdpPeer{"new": {Ip: "127.0.0.1", Port: 62207}}
		a.serverPeerMutex.Unlock()
		a.notifyServerMapUpdated()
		synctest.Wait()
		var next *core.MsgData
		select {
		case next = <-a.sendMsgCh:
		default:
			t.Error("replacement discovery waited for old transaction response")
			old.ResponseMsgCh <- &core.PacketParserData{Error: errors.New("test cleanup")}
			next = <-a.sendMsgCh
		}
		if next.RemoteAddr.Port != 62207 {
			t.Errorf("replacement sent to %v", next.RemoteAddr)
		}
		close(a.signals.stop)
		synctest.Wait()
		// Release a pre-fix response wait without blocking after the fix.
		select {
		case next.ResponseMsgCh <- &core.PacketParserData{Error: errors.New("late reply")}:
		default:
			t.Error("late reply blocked")
		}
		a.wg.Wait()
	})
}

func TestDiscoveryAddressRetry(t *testing.T) {
	for _, outcome := range []string{"recover", "stop", "reload", "sustained"} {
		t.Run(outcome, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				a := &UdpAC{}
				a.signals.stop = make(chan struct{})
				quit := make(chan struct{})
				var state serverDiscoveryState
				var attempts atomic.Int32
				var recoverDNS atomic.Bool
				want := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 62206}
				resolve := func() (net.Addr, error) {
					attempts.Add(1)
					if recoverDNS.Load() {
						return want, nil
					}
					return nil, errors.New("temporary DNS failure")
				}
				done := make(chan net.Addr, 1)
				start := func() { go func() { done <- a.waitForDiscoveryAddress(resolve, &state, quit) }(); synctest.Wait() }
				start()
				if atomic.LoadInt32(&state.failed) != 0 {
					t.Error("first DNS failure crossed retry threshold")
				}
				switch outcome {
				case "recover":
					recoverDNS.Store(true)
					time.Sleep(MinialServerDiscoveryInterval * time.Second)
				case "stop":
					close(a.signals.stop)
				case "reload", "sustained":
					for attempts.Load() < ServerDiscoveryRetryBeforeFail {
						if outcome == "reload" {
							close(quit)
							synctest.Wait()
							if <-done != nil {
								t.Fatal("canceled lookup returned an address")
							}
							quit = make(chan struct{})
							start()
						} else {
							time.Sleep(MinialServerDiscoveryInterval * time.Second)
							synctest.Wait()
						}
						wantFailed := int32(0)
						if attempts.Load() >= ServerDiscoveryRetryBeforeFail {
							wantFailed = 1
						}
						if atomic.LoadInt32(&state.failed) != wantFailed {
							t.Errorf("after %d attempts: failed=%d, want %d", attempts.Load(), atomic.LoadInt32(&state.failed), wantFailed)
						}
					}
					recoverDNS.Store(true)
					time.Sleep(MinialServerDiscoveryInterval * time.Second)
				}
				synctest.Wait()
				got := <-done
				if outcome == "stop" {
					if got != nil || attempts.Load() != 1 {
						t.Fatalf("canceled lookup retried: %v, attempts=%d", got, attempts.Load())
					}
				} else if got != want {
					t.Fatalf("resolution did not recover: %v", got)
				}
				wantFailed := int32(0)
				if outcome == "reload" || outcome == "sustained" {
					wantFailed = 1
				}
				if atomic.LoadInt32(&state.failed) != wantFailed {
					t.Error("DNS recovery changed failure state before a discovery reply")
				}
			})
		})
	}
}
