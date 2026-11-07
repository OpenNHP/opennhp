package ac

import (
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestDiscoveryRetryProgressSurvivesPeerChanges(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		peer := &core.UdpPeer{Ip: "127.0.0.1", Port: 62206}
		a := &UdpAC{config: &Config{FailurePolicy: FailurePolicyOpen}, device: &core.Device{}, sendMsgCh: make(chan *core.MsgData), serverPeerMap: map[string]*core.UdpPeer{endpointKey(peer): peer}, remoteConnectionMap: make(map[string]*UdpConn)}
		a.running.Store(true)
		a.signals.stop = make(chan struct{})
		a.signals.serverMapUpdated = make(chan struct{}, 1)
		repliesDone := make(chan struct{})
		go func() {
			defer close(repliesDone)
			for {
				select {
				case <-a.signals.stop:
					return
				case request := <-a.sendMsgCh:
					request.ResponseMsgCh <- &core.PacketParserData{Error: errors.New("unreachable")}
				}
			}
		}()
		var opened atomic.Bool
		a.wg.Add(1)
		go a.maintainServerConnections(func(open bool) error {
			if open {
				opened.Store(true)
			}
			return nil
		})
		defer func() { close(a.signals.stop); a.wg.Wait(); <-repliesDone }()
		synctest.Wait()
		for generation := 0; generation < 3; generation++ {
			// Each generation is shorter than the failure threshold. Only the retained
			// first endpoint can accumulate enough failures across these changes.
			time.Sleep((ServerDiscoveryRetryBeforeFail - 2) * MinialServerDiscoveryInterval * time.Second)
			synctest.Wait()
			if generation == 2 {
				break
			}
			a.serverPeerMutex.Lock()
			a.serverPeerMap = map[string]*core.UdpPeer{endpointKey(peer): {Ip: peer.Ip, Port: peer.Port}}
			if generation == 0 {
				other := &core.UdpPeer{Ip: "127.0.0.1", Port: 62207}
				a.serverPeerMap[endpointKey(other)] = other
			}
			a.serverPeerMutex.Unlock()
			a.notifyServerMapUpdated()
			synctest.Wait()
		}
		if !opened.Load() {
			t.Fatal("peer-list changes prevented fallback despite repeated failures of the retained endpoint")
		}
	})
}
