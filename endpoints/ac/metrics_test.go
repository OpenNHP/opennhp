package ac

import (
	"strings"
	"testing"
	"time"
)

func renderAC(t *testing.T, m *acMetrics) string {
	t.Helper()
	var b strings.Builder
	if err := m.registry.WriteText(&b); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	return b.String()
}

func TestACMetricsRenderAndRecord(t *testing.T) {
	a := &UdpAC{remoteConnectionMap: map[string]*UdpConn{"x": {}}}
	m := newACMetrics(a, time.Unix(1000, 0))

	// Closed-set series export 0 before any event.
	if body := renderAC(t, m); !strings.Contains(body, `nhp_ac_operations_total{result="ok"} 0`) ||
		!strings.Contains(body, `nhp_ac_packets_dropped_total{stage="decrypt"} 0`) {
		t.Fatalf("missing pre-initialized series:\n%s", body)
	}

	m.recordMessageReceived("NHP-AOP")
	m.recordACOperation(true, 0.01)
	m.recordACOperation(false, 0.2)
	m.recordDroppedPacket("decrypt")

	body := renderAC(t, m)
	for _, want := range []string{
		`nhp_ac_messages_received_total{type="NHP-AOP"} 1`,
		`nhp_ac_operations_total{result="ok"} 1`,
		`nhp_ac_operations_total{result="error"} 1`,
		`nhp_ac_packets_dropped_total{stage="decrypt"} 1`,
		`nhp_ac_active_connections 1`,
		`nhp_ac_operation_duration_seconds_count 2`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q\n---\n%s", want, body)
		}
	}
}

func TestACMetricsNilSafe(t *testing.T) {
	var m *acMetrics
	m.recordMessageReceived("NHP-AOP")
	m.recordACOperation(true, 1)
	m.recordDroppedPacket("parse")
}
