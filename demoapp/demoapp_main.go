// Command demoapp is the OpenNHP integrated demo: it owns user identity
// (username/password + OIDC), generates and stores each user's NHP key
// pair, and serves the SPA that registers the public key with nhp-server
// and knocks on hidden resources via js-agent + relay.
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// webFS holds an embedded copy of web/dist when built with `go build
// -tags webdist`. Default builds (e.g. `go run`) serve from disk so
// developers can iterate on the SPA without rebuilding the Go binary.
// The actual embed declaration lives in demoapp_web.go behind the same
// build tag — without it, webFS is nil and the server falls back to disk.
var webFS fs.FS

// embeddedWebDist returns the embedded web/dist when the webdist build
// tag is set, or nil otherwise. The build-tag-gated file supplies the
// non-nil value via init().
func embeddedWebDist() fs.FS {
	return webFS
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "demoapp:", err)
		os.Exit(1)
	}
}

// run is the cli entry point. Flags are kept minimal — config is loaded
// from the TOML file so operators can change settings without rebuilding.
func run() error {
	cfgPath := flag.String("c", "demoapp/etc/config.toml", "path to TOML config")
	flag.Parse()

	logger := MustLogger()

	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Auto-fill any placeholder secrets / keys on first boot. The values
	// land in a sidecar next to config.toml so the operator's file stays
	// pristine across rebuilds — and so the demo works out of the box in
	// local docker without an editor step.
	// Bound the relay probe loop generously — see relayProbeTotal in
	// autogen.go. A short budget here only lets ~1-2 attempts complete
	// before ctx.Done() fires; the retry loop is sized for ~30s of relay
	// boot + register dance and must be allowed to run to completion.
	probeCtx, cancel := context.WithTimeout(context.Background(), relayProbeTotal)
	defer cancel()
	filled, err := autoFillConfig(cfg, EtcDirOf(*cfgPath), probeCtx)
	if err != nil {
		return fmt.Errorf("auto-fill config: %w", err)
	}
	if filled {
		logger.Printf("auto-generated local secrets; sidecar at %s/%s",
			EtcDirOf(*cfgPath), autogenFile)
	}

	// Now that placeholders are filled, enforce the post-autogen invariants
	// (KeyEnvelopeKey decodes to 32 bytes, etc.).
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	masterKey, err := cfg.KeyEnvelopeKeyBytes()
	if err != nil {
		return err
	}

	store, err := OpenUserStore(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("open userstore: %w", err)
	}
	defer store.Close()

	oidcRP, err := maybeBuildOIDCRP(cfg)
	if err != nil {
		return fmt.Errorf("init OIDC: %w", err)
	}

	gh, err := maybeBuildGitHubProvider(cfg)
	if err != nil {
		return fmt.Errorf("init GitHub OAuth: %w", err)
	}

	app := &App{
		Cfg:          cfg,
		Store:        store,
		MasterKey:    masterKey,
		OIDCVefifier: oidcRP,
		GitHub:       gh,
		WebFS:        webFS,
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	if err := applyTrustedProxies(engine, cfg.TrustProxyHeaders); err != nil {
		return fmt.Errorf("set trusted proxies: %w", err)
	}
	engine.Use(gin.LoggerWithWriter(logger.Writer()))
	engine.Use(gin.Recovery())

	if err := app.Register(engine); err != nil {
		return fmt.Errorf("register routes: %w", err)
	}

	logger.Printf("listening on %s (db=%s, scheme=%s, oidc=%t, github=%t)",
		cfg.ListenAddr, cfg.DBPath, cfg.CipherScheme, oidcRP != nil, gh != nil)
	if err := engine.Run(cfg.ListenAddr); err != nil {
		return fmt.Errorf("http listen: %w", err)
	}
	return nil
}

// maybeBuildOIDCRP initializes the OIDC Relying Party when an enabled
// block is present. It returns (nil, nil) for "not configured" so callers
// can use a simple nil check rather than carrying an extra bool.
func maybeBuildOIDCRP(cfg *Config) (*OIDCRelyingParty, error) {
	oc, ok := cfg.EnabledOIDC()
	if !ok {
		return nil, nil
	}
	if oc.IssuerURL == "" || oc.ClientID == "" || oc.ClientSecret == "" || oc.RedirectURL == "" {
		return nil, errors.New("OIDC block is Enabled but missing required fields")
	}
	return NewOIDCRelyingParty(oc)
}

// maybeBuildGitHubProvider initializes the GitHub OAuth client when an
// enabled [[OAuth]] block is present. Returns (nil, nil) when none is
// enabled so callers use a simple nil check, mirroring maybeBuildOIDCRP.
func maybeBuildGitHubProvider(cfg *Config) (*GitHubProvider, error) {
	oc, ok := cfg.EnabledOAuth()
	if !ok {
		return nil, nil
	}
	return NewGitHubProvider(oc)
}

// GenerateSessionKey returns a random 32-byte key base64-encoded. Operators
// can use it to bootstrap a fresh SessionKey in config.toml without having
// to call openssl themselves. Wired up so we can ship a small CLI helper
// later if needed.
func GenerateSessionKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// trim helpers so we don't pull in strings repeatedly at startup.
func init() {
	// Strip a trailing newline from a pasted config value if present.
	_ = strings.TrimSpace
}
