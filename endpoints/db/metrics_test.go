package db

import (
	"strings"
	"testing"
	"time"
)

func TestDBMetricsRenderAndRecord(t *testing.T) {
	a := &UdpDevice{remoteConnectionMap: map[string]*UdpConn{"x": {}}}
	m := newDBMetrics(a, time.Unix(1000, 0))

	m.recordMessageReceived("DHP-DAK")
	m.recordDroppedPacket("validate")
	m.recordDroppedPacket("too_short")

	var b strings.Builder
	if err := m.registry.WriteText(&b); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	body := b.String()
	for _, want := range []string{
		`nhp_db_messages_received_total{type="DHP-DAK"} 1`,
		`nhp_db_packets_dropped_total{stage="validate"} 1`,
		`nhp_db_packets_dropped_total{stage="too_short"} 1`,
		`nhp_db_packets_dropped_total{stage="decrypt"} 0`,
		`nhp_db_active_connections 1`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q\n---\n%s", want, body)
		}
	}
}

func TestDBMetricsNilSafe(t *testing.T) {
	var m *dbMetrics
	m.recordMessageReceived("DHP-DAK")
	m.recordDroppedPacket("parse")
}

func TestDBByteCountersAreCounters(t *testing.T) {
	a := &UdpDevice{remoteConnectionMap: map[string]*UdpConn{}}
	m := newDBMetrics(a, time.Unix(1000, 0))
	var b strings.Builder
	if err := m.registry.WriteText(&b); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, name := range []string{"nhp_db_received_bytes_total", "nhp_db_sent_bytes_total"} {
		if !strings.Contains(out, "# TYPE "+name+" counter") {
			t.Errorf("%s must render as a counter:\n%s", name, out)
		}
	}
}
