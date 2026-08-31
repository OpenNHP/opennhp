// Package main (demoapp) hosts the OpenNHP integrated demo application.
//
// It owns user identity for the demo (username/password + OIDC), generates
// each user's NHP key pair at registration, and serves the bundled SPA that
// uses js-agent + relay to register the public key with nhp-server and knock
// on hidden resources.
//
// Configuration is loaded from a TOML file (default path: ./etc/config.toml,
// override with -c on the command line). Sensitive fields like KeyEnvelopeKey
// are accepted only via file/env; they are never logged.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// CipherScheme mirrors the JS Agent config; only the two schemes supported
// by the server plugin chain (basic / oidc) are accepted.
type CipherScheme string

const (
	CipherSchemeCurve25519 CipherScheme = "curve25519"
	CipherSchemeGMSM       CipherScheme = "gmsm"
)

// OIDCConfig holds the application-side OIDC RP settings. The Demo app acts
// as an OIDC Relying Party in its own right (independent from the nhp-server
// ASP plugin, which is left on the basic plugin). Email returned by the IdP
// becomes the NHP userId so the same identity is used for both auth and NHP.
type OIDCConfig struct {
	Enabled      bool   `toml:"Enabled"`
	IssuerURL    string `toml:"IssuerURL"`
	ClientID     string `toml:"ClientID"`
	ClientSecret string `toml:"ClientSecret"`
	RedirectURL  string `toml:"RedirectURL"`
}

// OAuthConfig holds settings for a plain OAuth 2.0 login provider. This is
// separate from OIDCConfig because some providers (notably GitHub) do not
// implement OIDC — no /.well-known/openid-configuration and no id_token — so
// they cannot go through the go-oidc discovery path. GitHub uses a fixed set
// of endpoints (see auth_github.go) and the /user API to obtain identity.
//
// Provider names which OAuth handler to use; only "github" is supported today.
// As with OIDC, the user's email becomes their NHP userId.
type OAuthConfig struct {
	Enabled      bool   `toml:"Enabled"`
	Provider     string `toml:"Provider"`
	ClientID     string `toml:"ClientID"`
	ClientSecret string `toml:"ClientSecret"`
	RedirectURL  string `toml:"RedirectURL"`
}

// ResourceConfig describes a single protected resource the Demo UI shows
// after a successful listServices. The id must match what the nhp-server
// basic plugin's ListService returns; the rest (title, URL, acHost) is UI
// metadata that the JS Agent does not know about.
//
// ServerName optionally scopes the resource to one configured [[Servers]]
// entry (matched by Name). When empty the resource is shown to every user;
// when set it is filtered to users registered against that server.
type ResourceConfig struct {
	ID         string `toml:"Id"`
	Title      string `toml:"Title"`
	URL        string `toml:"URL"`
	ACHost     string `toml:"ACHost"`
	ServerName string `toml:"ServerName"`
}

// ServerEntry describes one nhp-server the demo can register against. The
// registration page lists these (via ServerList) so the user picks a
// server AND a cipher scheme independently.
//
// A single nhp-server identity derives BOTH public keys from one private
// key (nhp-server `keygen --both`), so both Sm2PublicKey and
// Curve25519PublicKey are declared here. RelayRegisteredScheme names which
// of the two keys the relay registered the server under — that key drives
// the relay routing fingerprint (see nhp.ts addServer relayPublicKey).
type ServerEntry struct {
	Name                  string       `toml:"Name"`
	ServiceID             string       `toml:"ServiceId"`
	OrganizationID        string       `toml:"OrganizationId"`
	Sm2PublicKey          string       `toml:"Sm2PublicKey"`
	Curve25519PublicKey   string       `toml:"Curve25519PublicKey"`
	RelayRegisteredScheme CipherScheme `toml:"RelayRegisteredScheme"`
}

// publicKeyFor returns the server's public key matching scheme, or "" when
// that scheme's key is not configured for this server (cross-scheme knock
// is then unavailable — the caller surfaces this to the user).
func (s *ServerEntry) publicKeyFor(scheme CipherScheme) string {
	if scheme == CipherSchemeGMSM {
		return s.Sm2PublicKey
	}
	return s.Curve25519PublicKey
}

// relayPublicKeyFor returns the key the relay registered this server with,
// regardless of the knock scheme. It is "" when scheme matches the
// relay-registered scheme (the ECDH key itself already produces the right
// fingerprint, so js-agent needs no separate relayPublicKey).
func (s *ServerEntry) relayPublicKeyFor(scheme CipherScheme) string {
	if scheme == s.RelayRegisteredScheme {
		return ""
	}
	return s.publicKeyFor(s.RelayRegisteredScheme)
}

// Config is the full TOML-loaded configuration for the demoapp daemon.
type Config struct {
	ListenAddr     string `toml:"ListenAddr"`
	DBPath         string `toml:"DBPath"`
	SessionKey     string `toml:"SessionKey"`
	KeyEnvelopeKey string `toml:"KeyEnvelopeKey"` // base64, 32 bytes
	// CipherScheme is the DEFAULT scheme used when a path has no interactive
	// scheme choice (e.g. OIDC upsert before the user reaches the
	// complete-registration view). Interactive registration overrides it.
	CipherScheme   CipherScheme     `toml:"CipherScheme"`
	RelayURL       string           `toml:"RelayUrl"`       // server-side: demoapp → relay
	PublicRelayURL string           `toml:"PublicRelayUrl"` // browser-side: SPA → relay
	WebDistDir     string           `toml:"WebDistDir"`
	Servers        []ServerEntry    `toml:"Servers"`
	OIDCs          []OIDCConfig     `toml:"OIDC"`
	OAuths         []OAuthConfig    `toml:"OAuth"`
	Resources      []ResourceConfig `toml:"Resources"`

	// Legacy single-server fields. When [[Servers]] is empty, LoadConfig
	// promotes these into Servers[0] so existing single-cluster configs
	// keep working unchanged. When [[Servers]] is also present they are
	// ignored. Slated for removal.
	ServerPublicKey string `toml:"ServerPublicKey"`
	ServiceID       string `toml:"ServiceId"`
	OrganizationID  string `toml:"OrganizationId"`
}

// FindServer returns the configured server by Name (case-sensitive), or nil.
func (c *Config) FindServer(name string) *ServerEntry {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			return &c.Servers[i]
		}
	}
	return nil
}

// DefaultServer returns the first configured server, or nil when none. Used
// as the fallback for legacy/OIDC users that have no server_name stored.
func (c *Config) DefaultServer() *ServerEntry {
	if len(c.Servers) == 0 {
		return nil
	}
	return &c.Servers[0]
}

// ServerInfo is the public, key-free shape GET /api/servers returns so the
// registration page can render the server + scheme dropdowns without
// leaking any public keys.
type ServerInfo struct {
	Name                  string   `json:"name"`
	ServiceID             string   `json:"serviceId"`
	OrganizationID        string   `json:"organizationId"`
	Schemes               []string `json:"schemes"`
	RelayRegisteredScheme string   `json:"relayRegisteredScheme"`
}

// schemesFor returns the schemes a server can actually serve (those whose
// public key is configured). A server with both keys offers both; one with
// only the SM2 key (the common docker case) offers gmsm only.
func (s *ServerEntry) schemesFor() []string {
	out := make([]string, 0, 2)
	if s.Sm2PublicKey != "" {
		out = append(out, string(CipherSchemeGMSM))
	}
	if s.Curve25519PublicKey != "" {
		out = append(out, string(CipherSchemeCurve25519))
	}
	return out
}

// ServerList returns the key-free server catalog for the registration page.
func (c *Config) ServerList() []ServerInfo {
	out := make([]ServerInfo, 0, len(c.Servers))
	for i := range c.Servers {
		s := &c.Servers[i]
		out = append(out, ServerInfo{
			Name:                  s.Name,
			ServiceID:             s.ServiceID,
			OrganizationID:        s.OrganizationID,
			Schemes:               s.schemesFor(),
			RelayRegisteredScheme: string(s.RelayRegisteredScheme),
		})
	}
	return out
}

// ResourcesFor returns the resource catalog filtered to the given server's
// Name. Resources with an empty ServerName are global (shown to all).
func (c *Config) ResourcesFor(serverName string) []ResourceConfig {
	out := make([]ResourceConfig, 0, len(c.Resources))
	for _, r := range c.Resources {
		if r.ServerName == "" || r.ServerName == serverName {
			out = append(out, r)
		}
	}
	return out
}

// KeyEnvelopeKeyBytes decodes the base64-encoded 32-byte AES-256 master key
// used to wrap each user's NHP private key at rest. It fails fast on bad
// lengths so an operator typo is caught at startup, not at first login.
func (c *Config) KeyEnvelopeKeyBytes() ([]byte, error) {
	if c.KeyEnvelopeKey == "" {
		return nil, errors.New("KeyEnvelopeKey is not configured")
	}
	// Trim whitespace so operators can paste multi-line base64 values.
	trimmed := strings.TrimSpace(c.KeyEnvelopeKey)
	// Accept both std and url alphabets; we only decode length here.
	for _, enc := range []func(string) ([]byte, error){
		base64StdDecode, base64URLExtraDecode,
	} {
		key, err := enc(trimmed)
		if err == nil && len(key) == 32 {
			return key, nil
		}
	}
	// Surface the actual decoded length so an operator who fat-fingered a
	// 26-byte string sees "decoded to N bytes" instead of a generic
	// "must decode to exactly 32 bytes" message.
	for _, enc := range []func(string) ([]byte, error){
		base64StdDecode, base64URLExtraDecode,
	} {
		key, err := enc(trimmed)
		if err == nil {
			return nil, fmt.Errorf("KeyEnvelopeKey decoded to %d bytes (expected 32)", len(key))
		}
	}
	return nil, errors.New("KeyEnvelopeKey is not valid base64")
}

// EnabledOIDC returns the first enabled OIDC config block, if any.
func (c *Config) EnabledOIDC() (*OIDCConfig, bool) {
	for i := range c.OIDCs {
		if c.OIDCs[i].Enabled {
			return &c.OIDCs[i], true
		}
	}
	return nil, false
}

// EnabledOAuth returns the first enabled OAuth config block, if any.
func (c *Config) EnabledOAuth() (*OAuthConfig, bool) {
	for i := range c.OAuths {
		if c.OAuths[i].Enabled {
			return &c.OAuths[i], true
		}
	}
	return nil, false
}

// LoadConfig reads a TOML file from path and parses it into a Config. It
// applies defaults so the Demo can boot with a minimal config. Sensitive
// fields (KeyEnvelopeKey / SessionKey / ServerPublicKey) are NOT validated
// here — call Validate() after autoFillConfig has had a chance to fill in
// placeholders.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg := &Config{
		ListenAddr:     ":8081",
		DBPath:         "data/demo.db",
		RelayURL:       "http://localhost:8080/relay",
		ServiceID:      "example",
		OrganizationID: "opennhp.org",
		WebDistDir:     "web/dist",
	}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	// Default PublicRelayUrl to RelayUrl when unset so a local-dev setup
	// (where the browser talks to the same host:port as the relay) Just
	// Works. In docker compose the operator overrides this with
	// http://localhost:8080/relay because the SPA can't reach nhp-relay
	// by its container hostname.
	if cfg.PublicRelayURL == "" {
		cfg.PublicRelayURL = cfg.RelayURL
	}
	// CipherScheme is intentionally NOT defaulted here. Leaving it empty
	// lets autoFillConfig detect the relay's actual scheme via /servers
	// and pick the matching curve — important for the docker demo where
	// the relay ships as GMSM by default but the operator may not
	// realize their config.toml needs `CipherScheme = "gmsm"`. Validate()
	// at the end of LoadConfig still rejects non-empty invalid values.
	if cfg.CipherScheme != "" &&
		cfg.CipherScheme != CipherSchemeCurve25519 &&
		cfg.CipherScheme != CipherSchemeGMSM {
		return nil, fmt.Errorf("unsupported CipherScheme %q (expected curve25519 or gmsm)", cfg.CipherScheme)
	}

	// Back-compat: when no [[Servers]] block is present but the legacy
	// single-server fields are, promote them into Servers[0] so the rest of
	// the app only has to look at Config.Servers. The legacy CipherScheme
	// becomes both the server's RelayRegisteredScheme and the default
	// scheme. Only the matching public key is populated — the other
	// scheme's key stays empty, so cross-scheme knock is unavailable for
	// promoted (legacy) servers unless the operator migrates to [[Servers]]
	// and fills both keys.
	if len(cfg.Servers) == 0 {
		scheme := cfg.CipherScheme
		if scheme == "" {
			scheme = CipherSchemeCurve25519
		}
		se := ServerEntry{
			Name:                  "default",
			ServiceID:             cfg.ServiceID,
			OrganizationID:        cfg.OrganizationID,
			RelayRegisteredScheme: scheme,
		}
		if scheme == CipherSchemeGMSM {
			se.Sm2PublicKey = cfg.ServerPublicKey
		} else {
			se.Curve25519PublicKey = cfg.ServerPublicKey
		}
		cfg.Servers = []ServerEntry{se}
		// Default scheme for non-interactive paths follows the (legacy)
		// server's registered scheme when the operator left it empty.
		if cfg.CipherScheme == "" {
			cfg.CipherScheme = scheme
		}
	}
	return cfg, nil
}

// Validate enforces the post-autogen invariants: KeyEnvelopeKey decodes to
// exactly 32 bytes and CipherScheme is one of the supported values.
// Run this AFTER autoFillConfig so placeholders filled in by sidecar /
// random generation pass.
func (c *Config) Validate() error {
	// CipherScheme is the DEFAULT for non-interactive paths. It may be left
	// empty in a [[Servers]] config where each server declares its own
	// scheme; default it to the first server's registered scheme (or
	// curve25519) so OIDC upsert etc. always have something to fall back to.
	if c.CipherScheme == "" {
		if s := c.DefaultServer(); s != nil && s.RelayRegisteredScheme != "" {
			c.CipherScheme = s.RelayRegisteredScheme
		} else {
			c.CipherScheme = CipherSchemeCurve25519
		}
	}
	if c.CipherScheme != CipherSchemeCurve25519 && c.CipherScheme != CipherSchemeGMSM {
		return fmt.Errorf("CipherScheme must be %q or %q (got %q)", CipherSchemeCurve25519, CipherSchemeGMSM, c.CipherScheme)
	}
	if len(c.Servers) == 0 {
		return errors.New("no [[Servers]] configured and no legacy ServerPublicKey to promote")
	}
	if c.KeyEnvelopeKey == "" {
		return errors.New("KeyEnvelopeKey is required (32 bytes, base64-encoded)")
	}
	if _, err := c.KeyEnvelopeKeyBytes(); err != nil {
		return err
	}
	return nil
}
