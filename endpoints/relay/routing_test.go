package relay

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newRoutingTestServer builds a RelayServer just complete enough for
// resolveServer + handleRelay's HTTP-layer checks. It never connects a UDP
// socket: every server is created with an empty instances list, so the
// pickInstance() nil-guard short-circuits the request with 503 before any
// device code is touched. That's exactly the boundary we want to exercise —
// the HTTP routing contract — without dragging the full NHP pipeline into
// the test.
func newRoutingTestServer(serverIDs ...string) *RelayServer {
	servers := make(map[string]*serverRuntime, len(serverIDs))
	for _, id := range serverIDs {
		servers[id] = &serverRuntime{
			id:        id,
			name:      "test-" + id,
			scheme:    LBWeightedRandom,
			instances: nil, // pickInstance() returns nil → handler emits 503
		}
	}
	return &RelayServer{
		servers: servers,
	}
}

// validInnerPacket is the smallest payload that gets past the size and
// counter-offset checks in handleRelay. The contents don't have to parse
// as NHP — these tests stop at the pickInstance() guard, well before any
// crypto. 24 bytes covers the counter offset; we pad to 32 to make the
// fixture obviously "not empty" for human readers.
var validInnerPacket = bytes.Repeat([]byte{0x42}, 32)

// withLoopbackPeer rewrites the request's RemoteAddr so realClientAddr is
// satisfied. The handler treats loopback peers as proxied and requires
// X-Real-IP; tests would otherwise need to mock both the TCP peer and the
// header. Using a routable IP keeps the fixture small.
func newRelayRequest(method, path string, body []byte) *http.Request {
	var rdr *bytes.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	var r *http.Request
	if rdr != nil {
		r = httptest.NewRequest(method, path, rdr)
	} else {
		r = httptest.NewRequest(method, path, nil)
	}
	r.RemoteAddr = "203.0.113.10:54321"
	return r
}

// TestRouting_UnknownServerReturns404 confirms the no-silent-fallback
// invariant: an unrecognized server id must fail loudly, never get
// auto-redirected to some "default" server.
func TestRouting_UnknownServerReturns404(t *testing.T) {
	rs := newRoutingTestServer("good-id")

	w := httptest.NewRecorder()
	rs.handleRelay(w, newRelayRequest(http.MethodPost, "/relay/no-such-id", validInnerPacket))

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d (body: %s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "no-such-id") {
		t.Errorf("error body should mention the unknown id, got: %s", w.Body.String())
	}
}

// TestRouting_MissingServerIdReturns400 covers POST /relay/ (just the
// trailing slash). The legacy POST /relay path was removed, so an empty id
// must surface as a clear 400 telling the caller to use /relay/<id>.
func TestRouting_MissingServerIdReturns400(t *testing.T) {
	rs := newRoutingTestServer("good-id")

	w := httptest.NewRecorder()
	rs.handleRelay(w, newRelayRequest(http.MethodPost, "/relay/", validInnerPacket))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "missing server id") {
		t.Errorf("error body should explain the missing id, got: %s", w.Body.String())
	}
}

// TestRouting_LegacyBarePathRedirects documents the post-legacy behavior
// of bare POST /relay (no id, no trailing slash). The relay only registers
// "/relay/" with ServeMux, so a bare "/relay" gets the standard library's
// 301 redirect to "/relay/" — which then hits handleRelay and 400s on the
// missing id. Either response makes "use the legacy URL" loud, so this is
// the desired contract; the test pins it so a future change can't silently
// re-introduce a fallback to a default server.
func TestRouting_LegacyBarePathRedirects(t *testing.T) {
	rs := newRoutingTestServer("good-id")
	mux := http.NewServeMux()
	mux.HandleFunc("/relay/", rs.handleRelay)
	mux.HandleFunc("/servers", rs.handleServers)

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, newRelayRequest(http.MethodPost, "/relay", validInnerPacket))

	// 301 (ServeMux redirect) or 400 (handler missing-id) are both acceptable.
	// 200 / 2xx would mean some default-server path snuck back in — fail loudly.
	if w.Code == http.StatusMovedPermanently {
		loc := w.Header().Get("Location")
		if loc != "/relay/" {
			t.Errorf("expected redirect to /relay/, got Location=%q", loc)
		}
		return
	}
	if w.Code == http.StatusBadRequest {
		return
	}
	t.Fatalf("legacy POST /relay should 301 or 400, got %d (body: %s)", w.Code, w.Body.String())
}

// TestRouting_NonPostMethodReturns405 makes sure GET/PUT/DELETE on the
// relay path don't accidentally fall through to body parsing.
func TestRouting_NonPostMethodReturns405(t *testing.T) {
	rs := newRoutingTestServer("good-id")

	w := httptest.NewRecorder()
	rs.handleRelay(w, newRelayRequest(http.MethodGet, "/relay/good-id", nil))

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// TestRouting_EmptyBodyReturns400 keeps body validation in the regression
// net. The handler must reject empty / too-large / too-short bodies before
// it commits any state.
func TestRouting_EmptyBodyReturns400(t *testing.T) {
	rs := newRoutingTestServer("good-id")

	w := httptest.NewRecorder()
	rs.handleRelay(w, newRelayRequest(http.MethodPost, "/relay/good-id", []byte{}))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on empty body, got %d", w.Code)
	}
}

// TestRouting_ShortBodyReturns400 covers the < 24 byte rejection (the
// inner-packet counter offset).
func TestRouting_ShortBodyReturns400(t *testing.T) {
	rs := newRoutingTestServer("good-id")

	w := httptest.NewRecorder()
	rs.handleRelay(w, newRelayRequest(http.MethodPost, "/relay/good-id", []byte("short")))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on short body, got %d", w.Code)
	}
}

// TestRouting_OversizeBodyReturns400 covers the maxPacketSize cap. We
// intentionally send maxPacketSize+1 so io.LimitReader pulls exactly the
// boundary the handler checks.
func TestRouting_OversizeBodyReturns400(t *testing.T) {
	rs := newRoutingTestServer("good-id")

	big := bytes.Repeat([]byte{0x55}, maxPacketSize+1)
	w := httptest.NewRecorder()
	rs.handleRelay(w, newRelayRequest(http.MethodPost, "/relay/good-id", big))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on oversize body, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestRouting_NoInstanceReturns503 exercises the phase-2-future safety
// guard: a configured server with zero usable instances must return 503,
// not panic on a nil pickInstance().
func TestRouting_NoInstanceReturns503(t *testing.T) {
	rs := newRoutingTestServer("good-id")

	w := httptest.NewRecorder()
	rs.handleRelay(w, newRelayRequest(http.MethodPost, "/relay/good-id", validInnerPacket))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when server has no instance, got %d (body: %s)",
			w.Code, w.Body.String())
	}
}

// TestRouting_TrailingSlashTolerated: POST /relay/<id>/ should resolve the
// same as POST /relay/<id>. Some HTTP clients and proxies normalize paths
// by appending or stripping the trailing slash; routing must not depend on it.
func TestRouting_TrailingSlashTolerated(t *testing.T) {
	rs := newRoutingTestServer("good-id")

	w := httptest.NewRecorder()
	rs.handleRelay(w, newRelayRequest(http.MethodPost, "/relay/good-id/", validInnerPacket))

	if w.Code != http.StatusServiceUnavailable {
		// 503 = recognized server id but no instance (our test setup); that's
		// the expected "passed routing" outcome here, not 404.
		t.Fatalf("expected 503 (recognized id), got %d (body: %s)",
			w.Code, w.Body.String())
	}
}

// TestServers_SortedAndStable verifies the /servers output is sorted by
// id, so callers that diff or hash the response don't see Go-map-order
// churn between requests.
func TestServers_SortedAndStable(t *testing.T) {
	rs := newRoutingTestServer("zeta", "alpha", "mu")

	w := httptest.NewRecorder()
	rs.handleServers(w, httptest.NewRequest(http.MethodGet, "/servers", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var got []serverInfo
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(got))
	}
	wantOrder := []string{"alpha", "mu", "zeta"}
	for i := range got {
		if got[i].ID != wantOrder[i] {
			t.Errorf("position %d: got id=%q, want %q", i, got[i].ID, wantOrder[i])
		}
	}
}

// TestServers_NonGetReturns405 keeps the method check on the discovery
// endpoint in the regression net.
func TestServers_NonGetReturns405(t *testing.T) {
	rs := newRoutingTestServer("good-id")

	w := httptest.NewRecorder()
	rs.handleServers(w, httptest.NewRequest(http.MethodPost, "/servers", nil))

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
