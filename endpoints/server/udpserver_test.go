package server

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// TestDispatchHandlerRecordsBudgetExhaustion pins that a packet dropped
// because the handler goroutine budget (handlerSem) is exhausted is counted
// under nhp_server_handler_dropped_total, by message type — the same event
// that was previously only a log line.
func TestDispatchHandlerRecordsBudgetExhaustion(t *testing.T) {
	s := newTestServerWithMetrics()
	// A budget of 0 means the very first dispatch already finds the
	// semaphore full: select's non-blocking send always takes the default
	// branch, so this is deterministic and does not race a spawned handler.
	s.handlerSem = make(chan struct{}, 0)

	ppd := &core.PacketParserData{
		HeaderType: core.NHP_KNK,
		ConnData:   &core.ConnectionData{RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}},
	}

	called := false
	s.dispatchHandler(ppd, func(*core.PacketParserData) error { called = true; return nil })

	if called {
		t.Fatal("dispatchHandler must not invoke fn when the handler budget is exhausted")
	}

	var b strings.Builder
	if err := s.metrics.registry.WriteText(&b); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	body := b.String()
	if !strings.Contains(body, `nhp_server_handler_dropped_total{type="NHP-KNK"} 1`) {
		t.Errorf("missing handler-dropped counter\n---\n%s", body)
	}
}

// TestDispatchHandlerRunsWhenBudgetAvailable is the counter-side sanity
// check: with room in the semaphore, fn runs and nothing is counted as
// dropped.
func TestDispatchHandlerRunsWhenBudgetAvailable(t *testing.T) {
	s := newTestServerWithMetrics()
	s.handlerSem = make(chan struct{}, 1)

	ppd := &core.PacketParserData{
		HeaderType: core.NHP_KNK,
		ConnData:   &core.ConnectionData{RemoteAddr: &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}},
	}

	done := make(chan struct{})
	s.dispatchHandler(ppd, func(*core.PacketParserData) error { close(done); return nil })

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("fn was not invoked")
	}

	var b strings.Builder
	if err := s.metrics.registry.WriteText(&b); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	if strings.Contains(b.String(), `nhp_server_handler_dropped_total{type="NHP-KNK"} 1`) {
		t.Error("handler-dropped counter must not increment when the budget is available")
	}
}
