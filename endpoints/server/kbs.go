package server

import (
	"github.com/OpenNHP/opennhp/nhp/log"

	kbsAttest "github.com/OpenNHP/opennhp/endpoints/server/kbs/attest"
	kbsAuth "github.com/OpenNHP/opennhp/endpoints/server/kbs/auth"
	kbsResource "github.com/OpenNHP/opennhp/endpoints/server/kbs/resource"
)

// initKbsRouter mounts /kbs/v0 only when http.toml opts in.
//
// Off by default. These three routes hand out the content key for any file
// in the KBS repository, and nothing in front of them establishes who is
// asking: /auth issues a nonce and a session id that are never stored (so
// they cannot be checked later), /attest accepts a tee-evidence field that
// is never read and mints a token whose cosign_authorized is a constant,
// and the caller supplies the RSA key the content key is wrapped to. Three
// ordinary requests therefore recover any resource whose path is known.
//
// The switch is a containment measure, not the fix: it stops a deployment
// that never wanted KBS from serving it. Making KBS safe to enable means
// storing the nonce and session from /auth and checking them, feeding
// tee-evidence to nhp/core/verifier, and deriving cosign_authorized from
// the result.
func (hs *HttpServer) initKbsRouter() {
	if hs.config == nil || !hs.config.EnableKbs {
		hs.kbsMounted = false
		log.Info("KBS API disabled (EnableKbs is not true)")
		return
	}
	hs.kbsMounted = true

	// The remedy names a restart on purpose: the routes are built once, so
	// editing http.toml alone leaves them serving. See kbsPending.
	log.Warning("KBS API enabled: /kbs/v0/* releases resource content keys without verifying attestation. " +
		"Bind the HTTP listener somewhere only trusted callers reach, or set EnableKbs = false and restart the HTTP server " +
		"(EnableHttp = false, then back to true).")

	g := hs.ginEngine.Group("/kbs/v0")

	g.POST("/auth", kbsAuth.Auth)
	g.POST("/attest", kbsAttest.Attest)
	g.GET("/resource/*path", kbsResource.GetResource)
}
