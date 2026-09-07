package metrics

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStartEndpointServesMetricsAndHealthz(t *testing.T) {
	reg := NewRegistry()
	reg.NewCounter("nhp_test_events_total", "test events").With().Inc()

	ep, err := StartEndpoint(
		Config{Enabled: true, ListenIp: "127.0.0.1", ListenPort: 59107},
		EndpointOptions{Registry: reg, Uptime: func() time.Duration { return 42 * time.Second }},
	)
	if err != nil {
		t.Fatalf("StartEndpoint: %v", err)
	}
	defer ep.Stop()

	base := "http://127.0.0.1:59107"

	if body := httpGet(t, base+"/metrics"); !strings.Contains(body, "nhp_test_events_total 1") {
		t.Fatalf("/metrics missing counter:\n%s", body)
	}
	if hz := httpGet(t, base+"/healthz"); !strings.Contains(hz, `"status":"ok"`) || !strings.Contains(hz, `"uptime_s":42`) {
		t.Fatalf("/healthz unexpected body: %s", hz)
	}
}

func TestStartEndpointPortInUseReturnsError(t *testing.T) {
	ep, err := StartEndpoint(Config{Enabled: true, ListenPort: 59108}, EndpointOptions{Registry: NewRegistry()})
	if err != nil {
		t.Fatalf("first StartEndpoint: %v", err)
	}
	defer ep.Stop()

	if _, err := StartEndpoint(Config{Enabled: true, ListenPort: 59108}, EndpointOptions{Registry: NewRegistry()}); err == nil {
		t.Fatal("expected an error binding a port already in use")
	}
}

// TestEndpointStopNilSafe: a nil *Endpoint (endpoint failed to start) is safe
// to Stop — the daemons call ep.Stop() unconditionally.
func TestEndpointStopNilSafe(t *testing.T) {
	var ep *Endpoint
	ep.Stop()
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

func TestHealthzReports503WhenNotRunning(t *testing.T) {
	running := true
	ep, err := StartEndpoint(
		Config{Enabled: true, ListenIp: "127.0.0.1", ListenPort: 59109},
		EndpointOptions{Registry: NewRegistry(), IsRunning: func() bool { return running }},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer ep.Stop()

	base := "http://127.0.0.1:59109"
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
