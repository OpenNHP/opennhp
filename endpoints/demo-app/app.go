// Package demoapp — App constructor and route wiring.
//
// The Demo App exposes two surfaces:
//
//   1. Browser-facing routes (/, /login, /register, /dashboard, /api/*).
//      These carry the user session cookie and render HTML templates or
//      serve JSON to the SPA-style dashboard.
//
//   2. Internal API (/internal/nhp/*). These are called by the
//      demo-app-plugin running inside the nhp-server process. Access is
//      gated by a shared X-Plugin-Api-Key header.
package demoapp

import (
	"database/sql"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// App bundles the dependencies every handler needs. One App is created
// per process; handlers are methods on it so they can be wired up
// directly in SetupRoutes.
type App struct {
	Cfg      *Config
	DB       *sql.DB
	Mailer   Mailer
	Engine   *gin.Engine
	Sessions sessions.Store

	// Templates is the parsed *.html set. Pages render through
	// renderTemplate().
	Templates *template.Template
}

// New constructs an App from already-loaded dependencies. The webFS is
// an fs.FS rooted at the `web/` directory of the demo-app (it contains
// `templates/` and `static/` subdirectories). main/main.go embeds it.
func New(cfg *Config, db *sql.DB, mailer Mailer, store sessions.Store, webFS fs.FS) (*App, error) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(sessions.Sessions(SessionName, store))

	tmpl, err := parseTemplates(webFS)
	if err != nil {
		return nil, err
	}
	return &App{
		Cfg:       cfg,
		DB:        db,
		Mailer:    mailer,
		Engine:    engine,
		Sessions:  store,
		Templates: tmpl,
	}, nil
}

// SetupRoutes registers every endpoint the App serves. It is split into
// five groups so the policy of each group is clear at a glance: public
// pages, public APIs, authenticated pages, authenticated APIs, internal
// callbacks.
func (a *App) SetupRoutes() {
	// Static assets — served from the embedded FS so the binary stays
	// self-contained. Vite-emitted js-agent bundle lives at
	// web/static/nhp-agent.esm.js.
	if staticFS != nil {
		a.Engine.StaticFS("/static", http.FS(staticFS))
	}

	// Public pages
	a.Engine.GET("/", a.handleIndex)
	a.Engine.GET("/login", a.handleLoginPage)
	a.Engine.GET("/register", a.handleRegisterPage)

	// Public APIs (registration, login, OIDC)
	a.Engine.POST("/api/users/register", a.handleAPIRegister)
	a.Engine.POST("/api/users/login", a.handleAPILogin)
	a.Engine.POST("/api/users/logout", a.handleAPILogout)
	a.Engine.GET("/auth/oidc/start", a.handleOIDCStart)
	a.Engine.GET("/auth/oidc/callback", a.handleOIDCCallback)
	a.Engine.POST("/api/oidc/nhp-onboard", a.handleOIDCOnboard)

	// Authenticated pages
	authed := a.Engine.Group("/", AuthRequired(a.DB))
	authed.GET("/dashboard", a.handleDashboard)

	// Authenticated APIs
	authedAPI := a.Engine.Group("/api", AuthRequired(a.DB))
	authedAPI.GET("/user/profile", a.handleAPIProfile)
	authedAPI.GET("/nhp/config", a.handleAPINHPConfig)

	// Internal callbacks (plugin -> demo app)
	internal := a.Engine.Group("/internal", InternalAPIKey(a.Cfg.InternalApiKey))
	internal.POST("/nhp/otp-deliver", a.handleOTPDeliver)
	internal.POST("/nhp/mark-registered", a.handleMarkRegistered)
	internal.GET("/nhp/lookup-email", a.handleLookupEmail)
}

// renderTemplate executes the named template from the embedded FS. We
// pass the gin context so handlers can stash arbitrary values under
// c.Keys for the template.
func (a *App) renderTemplate(c *gin.Context, name string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := a.Templates.ExecuteTemplate(c.Writer, name, data); err != nil {
		// Already partially written — best we can do is log via gin's recovery.
		c.Error(err) //nolint:errcheck
	}
}
