package ac

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestDiscoveryWaitsForFirewallUpdate(t *testing.T) {
	for _, signal := range []string{"reload", "stop"} {
		t.Run(signal, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				a := &UdpAC{
					config:        &Config{FailurePolicy: FailurePolicyOpen},
					device:        &core.Device{},
					sendMsgCh:     make(chan *core.MsgData),
					serverPeerMap: map[string]*core.UdpPeer{},
				}
				a.running.Store(true)
				a.signals.stop = make(chan struct{})
				a.signals.serverMapUpdated = make(chan struct{}, 1)
				entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
				var changes []bool
				a.wg.Add(1)
				go func() {
					a.maintainServerConnections(func(open bool) error {
						if open {
							select {
							case <-entered:
							default:
								close(entered)
							}
							<-release
						}
						changes = append(changes, open)
						return nil
					})
					close(done)
				}()
				defer func() {
					select {
					case <-release:
					default:
						close(release)
					}
					select {
					case <-a.signals.stop:
					default:
						close(a.signals.stop)
					}
					<-done
				}()
				// Empty peers cause the first policy tick to open the fallback.
				<-entered
				if signal == "reload" {
					a.serverPeerMutex.Lock()
					a.serverPeerMap = map[string]*core.UdpPeer{"new": {Ip: "127.0.0.1", Port: 62207}}
					a.serverPeerMutex.Unlock()
					a.notifyServerMapUpdated()
				} else {
					close(a.signals.stop)
				}
				// All goroutines must be blocked before asserting that the pending
				// change prevents a replacement cycle or final shutdown cleanup.
				synctest.Wait()
				if len(changes) != 0 {
					t.Fatal("firewall operation finished before release")
				}
				if signal == "reload" && len(a.signals.serverMapUpdated) != 1 {
					t.Fatal("reload consumed while old firewall update is in flight")
				}
				select {
				case <-done:
					t.Fatal("maintenance stopped while firewall update is in flight")
				default:
				}
				close(release)
				synctest.Wait()
				if signal == "reload" {
					select {
					case request := <-a.sendMsgCh:
						if request.RemoteAddr.Port != 62207 {
							t.Fatalf("wrong peer: %v", request.RemoteAddr)
						}
					default:
						t.Fatal("replacement discovery did not start")
					}
					time.Sleep(MinialServerDiscoveryInterval * time.Second)
					synctest.Wait()
					if len(changes) < 2 || changes[0] != true || changes[len(changes)-1] != false {
						t.Fatalf("new peer did not close fallback: %v", changes)
					}
					close(a.signals.stop)
				}
				<-done
				if len(changes) < 2 || changes[0] != true || changes[len(changes)-1] != false {
					t.Fatalf("shutdown did not finish with closed fallback: %v", changes)
				}
			})
		})
	}
}

func TestPeerReloadNotifiesDiscovery(t *testing.T) {
	key := make([]byte, core.PrivateKeySize)
	key[0] = 1
	device := core.NewDevice(core.NHP_AC, key, nil)
	if device == nil {
		t.Fatal("device")
	}
	defer device.Stop()
	a := &UdpAC{device: device, config: &Config{}, serverPeerMap: map[string]*core.UdpPeer{}}
	a.running.Store(true)
	a.signals.serverMapUpdated = make(chan struct{}, 1)
	if e := a.updateServerPeers([]*core.UdpPeer{{PubKeyBase64: device.PublicKeyBase64(), Ip: "192.0.2.2", Port: 62206}}); e != nil {
		t.Fatal(e)
	}
	if len(a.serverPeerMap) != 1 {
		t.Fatal("map not installed")
	}
	select {
	case <-a.signals.serverMapUpdated:
	default:
		t.Error("new peer map installed without notifying active discovery")
	}
}
