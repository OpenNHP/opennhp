package metrics

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestStartEndpointServesMetricsAndHealthz(t *testing.T) {
	reg := NewRegistry()
	reg.NewCounter("nhp_test_events_total", "test events").With().Inc()

	ep, err := StartEndpoint(
		Config{Enabled: true, ListenIp: "127.0.0.1", ListenPort: 0}, // ephemeral
		EndpointOptions{Registry: reg, Uptime: func() time.Duration { return 42 * time.Second }},
	)
	if err != nil {
		t.Fatalf("StartEndpoint: %v", err)
	}
	defer ep.Stop()

	base := "http://" + ep.Addr()
	if body := httpGet(t, base+"/metrics"); !strings.Contains(body, "nhp_test_events_total 1") {
		t.Fatalf("/metrics missing counter:\n%s", body)
	}
	if hz := httpGet(t, base+"/healthz"); !strings.Contains(hz, `"status":"ok"`) || !strings.Contains(hz, `"uptime_s":42`) {
		t.Fatalf("/healthz unexpected body: %s", hz)
	}
}

func TestStartEndpointDisabledIsNoOp(t *testing.T) {
	ep, err := StartEndpoint(Config{Enabled: false}, EndpointOptions{Registry: NewRegistry()})
	if err != nil {
		t.Fatalf("disabled StartEndpoint should not error: %v", err)
	}
	if ep != nil {
		t.Fatal("disabled StartEndpoint should return a nil Endpoint")
	}
	ep.Stop() // must be nil-safe
}

func TestStartEndpointPortInUseReturnsError(t *testing.T) {
	ep, err := StartEndpoint(Config{Enabled: true, ListenIp: "127.0.0.1", ListenPort: 0}, EndpointOptions{Registry: NewRegistry()})
	if err != nil {
		t.Fatalf("first StartEndpoint: %v", err)
	}
	defer ep.Stop()

	_, portStr, _ := strings.Cut(ep.Addr(), ":")
	port, _ := strconv.Atoi(portStr)
	if _, err := StartEndpoint(Config{Enabled: true, ListenIp: "127.0.0.1", ListenPort: port}, EndpointOptions{Registry: NewRegistry()}); err == nil {
		t.Fatal("expected an error binding a port already in use")
	}
}

// TestHealthzReports503WhenNotRunning: the IsRunning gate flips /healthz to
// 503 so the endpoint works as a container liveness probe.
func TestHealthzReports503WhenNotRunning(t *testing.T) {
	running := true
	ep, err := StartEndpoint(
		Config{Enabled: true, ListenIp: "127.0.0.1", ListenPort: 0},
		EndpointOptions{Registry: NewRegistry(), IsRunning: func() bool { return running }},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer ep.Stop()

	base := "http://" + ep.Addr()
	if hz := httpGet(t, base+"/healthz"); !strings.Contains(hz, `"status":"ok"`) {
		t.Fatalf("running: %s", hz)
	}
	running = false
	resp, err := http.Get(base + "/healthz") //nolint:gosec // test-only localhost URL
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable || !strings.Contains(string(body), `"status":"stopping"`) {
		t.Fatalf("stopping: got %d %s", resp.StatusCode, body)
	}
}

// TestEndpointRejectsNonGET: POST /metrics is refused.
func TestEndpointRejectsNonGET(t *testing.T) {
	ep, err := StartEndpoint(Config{Enabled: true, ListenIp: "127.0.0.1", ListenPort: 0}, EndpointOptions{Registry: NewRegistry()})
	if err != nil {
		t.Fatal(err)
	}
	defer ep.Stop()

	resp, err := http.Post("http://"+ep.Addr()+"/metrics", "text/plain", nil) //nolint:gosec // test-only localhost URL
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST /metrics = %d, want 405", resp.StatusCode)
	}
}

func httpGet(t *testing.T, url string) string {
	t.Helper()
	var lastErr error
	for i := 0; i < 50; i++ {
		resp, err := http.Get(url) //nolint:gosec // test-only localhost URL
		if err != nil {
			lastErr = err
			time.Sleep(10 * time.Millisecond)
			continue
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return string(b)
	}
	t.Fatalf("GET %s failed: %v", url, lastErr)
	return ""
}
