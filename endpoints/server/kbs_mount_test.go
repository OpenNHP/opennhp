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
