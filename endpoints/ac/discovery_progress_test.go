package ac

import (
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

func progressAC() *UdpAC {
	a := &UdpAC{config: &Config{FailurePolicy: FailurePolicyOpen}, device: &core.Device{}, sendMsgCh: make(chan *core.MsgData), serverPeerMap: make(map[string]*core.UdpPeer), remoteConnectionMap: make(map[string]*UdpConn)}
	a.running.Store(true)
	a.signals.stop = make(chan struct{})
	a.signals.serverMapUpdated = make(chan struct{}, 1)
	return a
}

func progressPeers(a *UdpAC, peers ...*core.UdpPeer) {
	a.serverPeerMutex.Lock()
	a.serverPeerMap = make(map[string]*core.UdpPeer)
	for _, p := range peers {
		a.serverPeerMap[endpointKey(p)] = p
	}
	a.serverPeerMutex.Unlock()
	a.notifyServerMapUpdated()
}

func TestPolicyEvaluationSurvivesFastReload(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := progressAC()
		var open atomic.Bool
		a.wg.Add(1)
		go a.maintainServerConnections(func(v bool) error { open.Store(v); return nil })
		defer func() { close(a.signals.stop); a.wg.Wait() }()
		synctest.Wait()
		time.Sleep(5 * time.Second)
		synctest.Wait()
		if !open.Load() {
			t.Fatal("positive control: fallback did not open")
		}
		peer := &core.UdpPeer{Ip: "127.0.0.1", Port: 62206}
		go func() {
			for {
				select {
				case <-a.signals.stop:
					return
				case md := <-a.sendMsgCh:
					if md.ResponseMsgCh == nil {
						continue
					}
					a.remoteConnectionMutex.Lock()
					a.remoteConnectionMap[md.RemoteAddr.String()] = &UdpConn{ConnData: &core.ConnectionData{LastLocalRecvTime: time.Now().UnixNano()}}
					md.ConnData = a.remoteConnectionMap[md.RemoteAddr.String()].ConnData
					a.remoteConnectionMutex.Unlock()
					md.ResponseMsgCh <- &core.PacketParserData{HeaderType: core.NHP_AAK, BodyMessage: []byte(`{}`)}
				}
			}
		}()
		progressPeers(a, peer)
		synctest.Wait() // successful AAK has been consumed
		for i := 0; i < 30; i++ {
			time.Sleep(time.Second)
			if i%2 == 0 {
				progressPeers(a, peer, &core.UdpPeer{Ip: "127.0.0.1", Port: 62207})
			} else {
				progressPeers(a, peer)
			}
			synctest.Wait()
			if i >= 4 && open.Load() {
				t.Errorf("fallback still open %ds after successful discovery", i+1)
				break
			}
		}
	})
}

func TestUnchangedDiscoveryKeepsPendingRequest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		a := progressAC()
		peer := &core.UdpPeer{Ip: "127.0.0.1", Port: 62206}
		progressPeers(a, peer)
		var requestMu sync.Mutex
		var requests []time.Time
		go func() {
			for {
				select {
				case <-a.signals.stop:
					return
				case md := <-a.sendMsgCh:
					if md.RemoteAddr.Port != peer.Port {
						continue
					}
					requestMu.Lock()
					requests = append(requests, time.Now())
					requestMu.Unlock()
					go func() {
						time.Sleep(time.Duration(core.ACLocalTransactionResponseTimeoutMs) * time.Millisecond)
						md.ResponseMsgCh <- &core.PacketParserData{Error: errors.New("timeout")}
					}()
				}
			}
		}()
		var open atomic.Bool
		a.wg.Add(1)
		go a.maintainServerConnections(func(v bool) error { open.Store(v); return nil })
		defer func() { close(a.signals.stop); a.wg.Wait() }()
		synctest.Wait()
		for i := 0; i < 30; i++ {
			time.Sleep(time.Second)
			if i%2 == 0 {
				progressPeers(a, peer, &core.UdpPeer{Ip: "127.0.0.1", Port: 62207})
			} else {
				progressPeers(a, peer)
			}
			synctest.Wait()
		}
		minimum := time.Duration(core.ACLocalTransactionResponseTimeoutMs)*time.Millisecond + 5*time.Second
		requestMu.Lock()
		for i := 1; i < len(requests); i++ {
			if requests[i].Sub(requests[i-1]) < minimum {
				t.Errorf("unchanged endpoint request restarted after %s; minimum %s", requests[i].Sub(requests[i-1]), minimum)
				break
			}
		}
		requestMu.Unlock()
		// B is now absent. A's completed failures must already be retained.
		time.Sleep(5 * time.Second)
		synctest.Wait()
		if !open.Load() {
			t.Error("completed failures did not survive unrelated endpoint changes")
		}
	})
}

func TestDNSRetrySpacingSurvivesReload(t *testing.T) {
	oldResolver := net.DefaultResolver
	net.DefaultResolver = &net.Resolver{PreferGo: true, Dial: func(context.Context, string, string) (net.Conn, error) { return nil, errors.New("DNS unavailable") }}
	defer func() { net.DefaultResolver = oldResolver }()
	// Initialize net package resolver bookkeeping outside the synthetic-time bubble.
	_, _ = net.DefaultResolver.LookupHost(context.Background(), "warmup.invalid")
	logger := log.NewLogger("", 0, "", "")
	defer logger.Close()
	defer log.SetGlobalLogger(log.NewLogger("", 3, "", ""))
	synctest.Test(t, func(t *testing.T) {
		var attemptsMu sync.Mutex
		var attempts []time.Time
		logger.Error = func(format string, args ...any) {
			if strings.HasPrefix(format, "server discovery address unavailable") {
				attemptsMu.Lock()
				attempts = append(attempts, time.Now())
				attemptsMu.Unlock()
			}
		}
		log.SetGlobalLogger(logger)
		a := progressAC()
		peer := &core.UdpPeer{Hostname: "unavailable.invalid", Port: 62206}
		progressPeers(a, peer)
		a.wg.Add(1)
		go a.maintainServerConnections(func(bool) error { return nil })
		defer func() { close(a.signals.stop); a.wg.Wait() }()
		synctest.Wait()
		for i := 0; i < 12; i++ {
			time.Sleep(time.Second)
			if i%2 == 0 {
				progressPeers(a, peer, &core.UdpPeer{Ip: "127.0.0.1", Port: 62207})
			} else {
				progressPeers(a, peer)
			}
			synctest.Wait()
		}
		attemptsMu.Lock()
		defer attemptsMu.Unlock()
		if len(attempts) < 3 {
			t.Fatalf("DNS retry positive control: only %d attempts", len(attempts))
		}
		for i := 1; i < len(attempts); i++ {
			if attempts[i].Sub(attempts[i-1]) < 5*time.Second {
				t.Fatalf("DNS retried after %s because unrelated endpoint changed", attempts[i].Sub(attempts[i-1]))
			}
		}
	})
}

func TestPolicyLogStateSurvivesReload(t *testing.T) {
	for _, fails := range []bool{false, true} {
		t.Run(map[bool]string{false: "outcome", true: "error"}[fails], func(t *testing.T) {
			logger := log.NewLogger("", 0, "", "")
			defer logger.Close()
			defer log.SetGlobalLogger(log.NewLogger("", 3, "", ""))
			synctest.Test(t, func(t *testing.T) {
				var alerts atomic.Int32
				logger.Info = func(format string, args ...any) {
					if strings.HasPrefix(format, "AC failure policy:") {
						alerts.Add(1)
					}
				}
				logger.Error = func(format string, args ...any) {
					if strings.HasPrefix(format, "AC failure policy:") {
						alerts.Add(1)
					}
				}
				log.SetGlobalLogger(logger)
				a := progressAC()
				peer := &core.UdpPeer{Ip: "127.0.0.1", Port: 62206}
				progressPeers(a, peer) // blocked request stays unknown; outcome stays closed
				a.wg.Add(1)
				go a.maintainServerConnections(func(bool) error {
					if fails {
						return errors.New("same firewall failure")
					}
					return nil
				})
				defer func() { close(a.signals.stop); a.wg.Wait() }()
				synctest.Wait()
				for i := 0; i < 3; i++ {
					time.Sleep(6 * time.Second)
					synctest.Wait()
					if i%2 == 0 {
						progressPeers(a, peer, &core.UdpPeer{Ip: "127.0.0.1", Port: 62207})
					} else {
						progressPeers(a, peer)
					}
					synctest.Wait()
				}
				if alerts.Load() != 1 {
					t.Fatalf("unchanged result logged %d times across reloads", alerts.Load())
				}
			})
		})
	}
}
