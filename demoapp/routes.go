package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// Session keys. Names are short to keep cookie payloads small.
const (
	sessKeyUserID      = "uid"
	sessKeyUsername    = "uname"
	sessKeyOIDCState   = "ostate"
	sessKeyOIDCNonce   = "ononce"
	sessKeyOIDCSubject = "osub"
	sessKeyOAuthState  = "gstate"
	sessKeyRegToken    = "regtok"
)

// SessionName is the cookie name the demo uses (mirrors nhp-server's
// "nhpsessions" so the two daemons can co-exist if reverse-proxied).
const SessionName = "demosessions"

// clearSession clears and persists the session. gin-contrib/sessions
// only writes the cookie on Save() (every success path calls it
// explicitly), so a bare s.Clear() on an error return leaves the old
// state — including a replayable OAuth/OIDC state/nonce — in the
// cookie until expiry. Call this on every error branch of the IdP
// callbacks so the state bit is actually invalidated.
func clearSession(s sessions.Session) {
	s.Clear()
	_ = s.Save()
}

// App bundles everything a route handler needs: config, store, the master
// key bytes (so we don't decode on every request), and an optional OIDC
// verifier (nil if OIDC isn't enabled).
type App struct {
	Cfg          *Config
	Store        *UserStore
	MasterKey    []byte
	OIDCVefifier *OIDCRelyingParty // nil when OIDC is disabled
	GitHub       *GitHubProvider   // nil when GitHub OAuth is disabled
	WebFS        fs.FS             // optional embedded web/dist
}

// Register wires all HTTP routes onto the given gin engine. The order
// matters: public endpoints (register/login/oidc/health) come first,
// then session-gated endpoints, then the static-file fallback.
func (a *App) Register(r *gin.Engine) error {
	// Cookie-based session store keyed by Cfg.SessionKey. 32 bytes is the
	// minimum recommended by gin-contrib/sessions/cookie.
	if len(a.Cfg.SessionKey) < 16 {
		return errors.New("SessionKey must be at least 16 bytes")
	}
	store := cookie.NewStore([]byte(a.Cfg.SessionKey))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7, // 1 week
		HttpOnly: true,
		Secure:   a.Cfg.SecureCookies, // prod (HTTPS): true; local HTTP dev: false
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions(SessionName, store))

	// Security headers — small but important for a demo that ships a
	// private key to the browser. The CSP connect-src must allow the
	// relay origin in addition to 'self', because the SPA (served from
	// :8081) opens an HTTP fetch to the relay (typically :8080) to drive
	// NHP-OTP / NHP-REG. Without it the browser blocks the knock with a
	// CSP violation.
	connectSrc := "'self'"
	if origin := relayOrigin(a.Cfg.PublicRelayURL); origin != "" {
		connectSrc += " " + origin
	}
	r.Use(func(c *gin.Context) {
		c.Header("Content-Security-Policy",
			fmt.Sprintf("default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src %s; object-src 'none'; frame-ancestors 'none'", connectSrc))
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	})

	pub := r.Group("/api")
	{
		pub.GET("/health", a.handleHealth)
		pub.GET("/servers", a.handleServers)
		// Rate-limit the unauthenticated work endpoints (bcrypt, keygen,
		// row insert, and the registration → requestOtp → nhp-server
		// email chain) so a public caller cannot turn them into a CPU/DB
		// sink or an email-amplification primitive (review #6).
		pub.POST("/register", rateLimit(registerLimiter), a.handleRegister)
		pub.POST("/register/confirm", a.handleRegisterConfirm)
		pub.POST("/login", rateLimit(loginLimiter), a.handleLogin)
		pub.POST("/logout", a.handleLogout)
		// OIDC routes are mounted even when disabled — the handlers
		// return 503 so the SPA can show a friendly message instead of
		// 404-ing on the link click.
		pub.GET("/auth/oidc/login", a.handleOIDCLogin)
		pub.GET("/auth/oidc/callback", a.handleOIDCCallback)

		// GitHub OAuth routes follow the same always-mounted pattern so
		// the SPA link does not 404 when the provider is not configured.
		pub.GET("/auth/github/login", a.handleGitHubLogin)
		pub.GET("/auth/github/callback", a.handleGitHubCallback)
	}

	auth := r.Group("/api")
	auth.Use(a.requireSession)
	{
		auth.GET("/config", a.handleGetConfig)
		auth.GET("/credentials", a.handleGetCredentials)
		auth.GET("/me", a.handleMe)
		auth.DELETE("/account", a.handleDeleteAccount)
		// Re-pick nhp-server cluster + cipher scheme before completing
		// NHP_REG (session-gated). Used by the complete-registration
		// view, where GitHub/OIDC users land with a default binding.
		auth.POST("/register/bind", a.handleRebind)
	}

	// Static SPA fallback. Serve from disk if WebDistDir is configured
	// (dev), else from the embedded fs (release build).
	if a.WebFS != nil {
		r.NoRoute(func(c *gin.Context) {
			if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
				a.serveSPA(c, a.WebFS)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		})
		return nil
	}
	// Fallback: serve from disk if WebDistDir exists; otherwise a tiny
	// index page so the server still boots before `npm run build`.
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		dir := a.Cfg.WebDistDir
		if dir == "" {
			c.JSON(http.StatusOK, gin.H{"message": "OpenNHP demoapp (no UI built)"})
			return
		}
		indexPath := dir + "/index.html"
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "UI not built; run `cd web && npm run build`"})
	})
	return nil
}

// requireSession is the middleware used by /api/config and /api/credentials
// — any handler that returns the user's private key MUST run after this.
func (a *App) requireSession(c *gin.Context) {
	s := sessions.Default(c)
	uidRaw := s.Get(sessKeyUserID)
	if uidRaw == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "login required"})
		return
	}
	uid, ok := uidRaw.(int64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "login required"})
		return
	}
	user, err := a.Store.GetUserByID(c.Request.Context(), uid)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	c.Set("user", user)
	c.Next()
}

// handleHealth is a quick readiness probe — no DB hit, no secrets.
func (a *App) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleServers returns the key-free server catalog so the registration
// page can render the server + cipher-scheme dropdowns. No public keys are
// exposed here — those are delivered per-user via /api/credentials and
// /api/config only after authentication.
func (a *App) handleServers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"servers": a.Cfg.ServerList()})
}

// serveSPA streams the bundled SPA. Falls back to index.html for client-
// side routes that don't have a matching static file.
func (a *App) serveSPA(c *gin.Context, fsys fs.FS) {
	path := strings.TrimPrefix(c.Request.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}
	f, err := fsys.Open(path)
	if err != nil {
		// SPA fallback for client-side routes.
		f, err = fsys.Open("index.html")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
			return
		}
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "stat failed"})
		return
	}
	// embed.FS opens a directory successfully (returning a *fs.File that
	// implements Read/Stat/Close/ReadDir only — no Seek). The unchecked
	// type assertion below would panic on those; treat directories as
	// not-found so vite's dist/assets/ doesn't 500 the public demo on
	// GET /assets.
	if stat.IsDir() {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}
	rs, ok := f.(interface {
		Read(p []byte) (int, error)
		Seek(offset int64, whence int) (int64, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "asset not seekable"})
		return
	}
	// http.ServeContent handles MIME sniffing, Range, If-Modified-Since.
	http.ServeContent(c.Writer, c.Request, path, stat.ModTime(), rs)
}

// contextOf is a tiny helper to avoid repeating context.TODO() at every
// call site; it prefers the request context when available.
func contextOf(c *gin.Context) context.Context {
	if c != nil && c.Request != nil && c.Request.Context() != nil {
		return c.Request.Context()
	}
	return context.TODO()
}

// relayOrigin extracts the scheme://host[:port] of the browser-facing relay
// URL so it can be added to the CSP connect-src directive. Returns "" for
// empty or unparseable input (the CSP then falls back to 'self' only).
func relayOrigin(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// MustLogger returns a *log.Logger that's safe to use even before the gin
// engine exists (used in main during startup).
func MustLogger() *log.Logger {
	return log.New(os.Stdout, "[demoapp] ", log.LstdFlags|log.Lmicroseconds)
}
