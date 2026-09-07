package relay

import (
	"strings"
	"testing"
	"time"
)

func TestRelayMetricsRenderAndRecord(t *testing.T) {
	rs := &RelayServer{servers: map[string]*serverRuntime{"a": {}, "b": {}}}
	m := newRelayMetrics(rs, time.Unix(1000, 0))

	m.recordMessageReceived("NHP-RKN")
	m.recordDroppedPacket("decrypt")

	var b strings.Builder
	if err := m.registry.WriteText(&b); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	body := b.String()
	for _, want := range []string{
		`nhp_relay_messages_received_total{type="NHP-RKN"} 1`,
		`nhp_relay_packets_dropped_total{stage="decrypt"} 1`,
		`nhp_relay_packets_dropped_total{stage="parse"} 0`,
		`nhp_relay_upstream_servers 2`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q\n---\n%s", want, body)
		}
	}
}

func TestRelayMetricsNilSafe(t *testing.T) {
	var m *relayMetrics
	m.recordMessageReceived("NHP-RKN")
	m.recordDroppedPacket("parse")
}

func TestRelayByteCountersAreCounters(t *testing.T) {
	rs := &RelayServer{servers: map[string]*serverRuntime{}}
	m := newRelayMetrics(rs, time.Unix(1000, 0))
	var b strings.Builder
	if err := m.registry.WriteText(&b); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, name := range []string{"nhp_relay_received_bytes_total", "nhp_relay_sent_bytes_total"} {
		if !strings.Contains(out, "# TYPE "+name+" counter") {
			t.Errorf("%s must render as a counter (it is a monotonic total):\n%s", name, out)
		}
	}
}
