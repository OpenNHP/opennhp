package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// /kbs/v0 is not served unless http.toml asks for it.
//
// The routes release resource content keys and the attestation meant to
// gate them is not implemented, so a deployment that never configured KBS
// must not be serving it — and the shipped http.toml binds 0.0.0.0:443,
// which made "mounted unconditionally" mean "on the internet".
func TestKbsRoutesAreNotMountedByDefault(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	for name, tc := range map[string]struct {
		config *HttpConfig
		want   int
	}{
		"no config":        {nil, http.StatusNotFound},
		"config unset":     {&HttpConfig{}, http.StatusNotFound},
		"explicitly false": {&HttpConfig{EnableKbs: false}, http.StatusNotFound},
		"opted in":         {&HttpConfig{EnableKbs: true}, http.StatusUnauthorized},
	} {
		t.Run(name, func(t *testing.T) {
			hs := &HttpServer{ginEngine: gin.New(), config: tc.config}
			hs.initKbsRouter()

			// Unauthorized rather than 404 once mounted: the route exists
			// and refuses the request on its own terms.
			w := httptest.NewRecorder()
			hs.ginEngine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/kbs/v0/resource/x", nil))
			if w.Code != tc.want {
				t.Errorf("GET /kbs/v0/resource/x = %d, want %d", w.Code, tc.want)
			}

			w = httptest.NewRecorder()
			hs.ginEngine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/kbs/v0/auth", nil))
			mounted := w.Code != http.StatusNotFound
			if mounted != (tc.config != nil && tc.config.EnableKbs) {
				t.Errorf("POST /kbs/v0/auth = %d; mounted=%v, want mounted=%v",
					w.Code, mounted, tc.config != nil && tc.config.EnableKbs)
			}
		})
	}
}

// Turning EnableKbs off in http.toml does not unmount a running listener,
// and kbsPending is what makes updateHttpConfig say so.
//
// This is the incident-response path for an exposed Key Broker Service: an
// operator edits the file, the reload succeeds, and /kbs/v0 keeps serving
// content keys until the HTTP server restarts. Silent would be the worst
// possible outcome here, so the gap is detected and logged Critical.
func TestKbsPendingReportsAReloadThatHasNotTakenEffect(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	for name, tc := range map[string]struct {
		mounted    bool
		running    bool
		configured bool
		want       bool
	}{
		"running, turned off":    {mounted: true, running: true, configured: false, want: true},
		"running, turned on":     {mounted: false, running: true, configured: true, want: true},
		"running, unchanged on":  {mounted: true, running: true, configured: true, want: false},
		"running, unchanged off": {mounted: false, running: true, configured: false, want: false},
		// Not running: the next Start reads the new config, so there is
		// no gap to report.
		"stopped": {mounted: true, running: false, configured: false, want: false},
	} {
		t.Run(name, func(t *testing.T) {
			hs := &HttpServer{kbsMounted: tc.mounted}
			hs.running.Store(tc.running)
			if got := hs.kbsPending(tc.configured); got != tc.want {
				t.Errorf("kbsPending(%v) = %v, want %v", tc.configured, got, tc.want)
			}
		})
	}

	// A nil receiver is the shape updateHttpConfig can hold before the
	// first Start; it must not panic.
	var nilHS *HttpServer
	if nilHS.kbsPending(true) {
		t.Error("a nil HttpServer reported a pending change")
	}
}

// initKbsRouter records what it actually did, which is what kbsPending
// compares against. A mount that did not update the flag would make every
// reload look unchanged.
func TestInitKbsRouterRecordsWhatItMounted(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	for _, enabled := range []bool{true, false} {
		hs := &HttpServer{ginEngine: gin.New(), config: &HttpConfig{EnableKbs: enabled}}
		hs.initKbsRouter()
		if hs.kbsMounted != enabled {
			t.Errorf("EnableKbs=%v: kbsMounted = %v", enabled, hs.kbsMounted)
		}
	}
}
