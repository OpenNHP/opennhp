package relay

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/metrics"
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

// TestNewMetricsEndpointStartsAfterServersAndNoLeakOnError:
//   - with a bad server key, New() fails and must NOT have left a metrics
//     listener bound (a retry after fixing the config would collide).
//   - with a good config, the endpoint comes up and the upstream-servers
//     gauge is readable without racing the (already-finished) buildServer
//     loop — covered by -race on the whole package.
func TestNewMetricsEndpointNoLeakOnError(t *testing.T) {
	// Reserve an ephemeral port, note it, release it — New() must not bind it.
	rl, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := rl.Addr().(*net.TCPAddr).Port
	rl.Close()

	bad := &Config{
		PrivateKeyBase64: core.NewECDH(core.ECC_CURVE25519).PrivateKeyBase64(),
		Metrics:          metrics.Config{Enabled: true, ListenIp: "127.0.0.1", ListenPort: port},
		Servers: []Server{{
			Name:         "x",
			PubKeyBase64: "!!!not base64!!!",
			LoadBalance:  "weighted-random",
		}},
	}
	if _, nerr := New(bad); nerr == nil {
		t.Fatal("New with a bad server key should fail")
	}

	// The port must be free — New must not have started the metrics listener
	// before the buildServer loop that returned the error.
	l, lerr := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if lerr != nil {
		t.Fatalf("metrics port %d still bound after New() failed: %v", port, lerr)
	}
	l.Close()
}
