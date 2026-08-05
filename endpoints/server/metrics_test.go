package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

func newTestServerWithMetrics() *UdpServer {
	s := &UdpServer{}
	s.remoteConnectionMap = make(map[string]*UdpConn)
	s.startTime = time.Now()
	s.metrics = newServerMetrics(s, s.startTime)
	return s
}

func TestMetricsEndpointExposesInstrumentedValues(t *testing.T) {
	s := newTestServerWithMetrics()
	s.metrics.recordMessageReceived("NHP-KNK")
	s.metrics.recordMessageReceived("NHP-KNK")
	s.metrics.recordKnockAuth(true)
	s.metrics.recordKnockAuth(false)
	s.metrics.recordACOperation(true, 0.02)
	s.metrics.recordACOperation(false, 0.5)
	s.metrics.recordBlockedAddr()

	ms := &metricsServer{us: s}
	rec := httptest.NewRecorder()
	ms.handleMetrics(rec, httptest.NewRequest("GET", "/metrics", nil))

	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("unexpected content type %q", ct)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`nhp_server_messages_received_total{type="NHP-KNK"} 2`,
		`nhp_server_knock_auth_total{result="ok"} 1`,
		`nhp_server_knock_auth_total{result="denied"} 1`,
		`nhp_server_ac_operations_total{result="ok"} 1`,
		`nhp_server_ac_operations_total{result="error"} 1`,
		`nhp_server_blocked_source_addresses_total 1`,
		`nhp_server_active_connections 0`,
		`nhp_server_overloaded 0`,
		`nhp_server_ac_operation_duration_seconds_count 2`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics output missing %q\n---\n%s", want, body)
		}
	}
}

func TestActiveConnectionsGaugeReflectsMap(t *testing.T) {
	s := newTestServerWithMetrics()
	s.remoteConnectionMap["a"] = &UdpConn{}
	s.remoteConnectionMap["b"] = &UdpConn{}

	ms := &metricsServer{us: s}
	rec := httptest.NewRecorder()
	ms.handleMetrics(rec, httptest.NewRequest("GET", "/metrics", nil))

	if !strings.Contains(rec.Body.String(), "nhp_server_active_connections 2") {
		t.Errorf("active_connections should reflect the live map:\n%s", rec.Body.String())
	}
}

func TestHealthzEndpoint(t *testing.T) {
	s := newTestServerWithMetrics()
	ms := &metricsServer{us: s}
	rec := httptest.NewRecorder()
	ms.handleHealthz(rec, httptest.NewRequest("GET", "/healthz", nil))

	if rec.Code != 200 {
		t.Fatalf("healthz status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"status":"ok"`) || !strings.Contains(body, `"uptime_s"`) {
		t.Errorf("unexpected healthz body: %s", body)
	}
	// Version/commit must NOT be exposed on the liveness probe (fingerprinting).
	if strings.Contains(body, `"version"`) || strings.Contains(body, `"commit"`) {
		t.Errorf("healthz must not leak version/commit: %s", body)
	}
}

// TestClosedSetSeriesPreInitializedToZero ensures the enumerable knock/AC
// result series export an explicit 0 before any event, so dashboards see a
// zero rather than a missing series.
func TestClosedSetSeriesPreInitializedToZero(t *testing.T) {
	s := newTestServerWithMetrics() // no record* calls
	ms := &metricsServer{us: s}
	rec := httptest.NewRecorder()
	ms.handleMetrics(rec, httptest.NewRequest("GET", "/metrics", nil))

	body := rec.Body.String()
	for _, want := range []string{
		`nhp_server_knock_auth_total{result="ok"} 0`,
		`nhp_server_knock_auth_total{result="denied"} 0`,
		`nhp_server_ac_operations_total{result="ok"} 0`,
		`nhp_server_ac_operations_total{result="error"} 0`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("expected pre-initialized series %q\n---\n%s", want, body)
		}
	}
}

func TestServerMetricsNilSafe(t *testing.T) {
	var m *serverMetrics // nil
	// none of these should panic
	m.recordMessageReceived("NHP-KNK")
	m.recordKnockAuth(true)
	m.recordACOperation(false, 1.0)
	m.recordBlockedAddr()
}

func TestMetricsConfigParsing(t *testing.T) {
	var c Config
	in := "[Metrics]\nEnabled = true\nListenIp = \"0.0.0.0\"\nListenPort = 9999\n"
	if err := toml.Unmarshal([]byte(in), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !c.Metrics.Enabled {
		t.Error("Metrics.Enabled should be true")
	}
	if c.Metrics.ListenIp != "0.0.0.0" {
		t.Errorf("ListenIp = %q, want 0.0.0.0", c.Metrics.ListenIp)
	}
	if c.Metrics.ListenPort != 9999 {
		t.Errorf("ListenPort = %d, want 9999", c.Metrics.ListenPort)
	}
}

func TestMetricsServerLiveEndpoint(t *testing.T) {
	// Grab a free port, then let the metrics server bind it for real so the
	// full listen/serve/shutdown path is exercised, not just the handlers.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	s := newTestServerWithMetrics()
	s.config = &Config{Metrics: MetricsConfig{Enabled: true, ListenIp: "127.0.0.1", ListenPort: port}}
	s.metrics.recordKnockAuth(true)

	s.metricsServer = newMetricsServer(s)
	if startErr := s.metricsServer.start(); startErr != nil {
		t.Fatalf("start metrics server: %v", startErr)
	}
	defer s.metricsServer.stop()

	base := fmt.Sprintf("http://127.0.0.1:%d", port)
	client := &http.Client{Timeout: 2 * time.Second}

	// /healthz
	resp, err := client.Get(base + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	hBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(hBody), `"status":"ok"`) {
		t.Errorf("/healthz = %d %s", resp.StatusCode, hBody)
	}

	// /metrics
	resp, err = client.Get(base + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	mBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(mBody), `nhp_server_knock_auth_total{result="ok"} 1`) {
		t.Errorf("/metrics missing instrumented value:\n%s", mBody)
	}
}

func TestMetricsDisabledByDefault(t *testing.T) {
	var c Config
	if err := toml.Unmarshal([]byte("ListenPort = 62206\n"), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if c.Metrics.Enabled {
		t.Error("Metrics must be disabled when no [Metrics] section is present")
	}
}
