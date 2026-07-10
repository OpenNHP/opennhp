package ac

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestRunUDPHandlerRecoversPanicAndReleasesWaitGroup(t *testing.T) {
	a := &UdpAC{config: &Config{ACId: "test-ac"}}
	a.wg.Add(1)

	done := make(chan struct{})
	go func() {
		a.runUDPHandler(core.NHP_AOP, func() { panic("malformed packet") })
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("panicking UDP handler did not return")
	}

	waitDone := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-time.After(time.Second):
		t.Fatal("panicking UDP handler did not release UdpAC wait group")
	}
}

func TestRunUDPHandlerHappyPath(t *testing.T) {
	a := &UdpAC{}
	a.wg.Add(1)
	var called atomic.Bool

	a.runUDPHandler(core.NHP_AOP, func() { called.Store(true) })
	a.wg.Wait()

	if !called.Load() {
		t.Fatal("handler was not called")
	}
}

func TestRecoverUDPHandlerNilConfigIsSafe(t *testing.T) {
	a := &UdpAC{}
	a.wg.Add(1)
	a.runUDPHandler(core.NHP_AOP, func() { panic("nil config") })
	a.wg.Wait()
}
