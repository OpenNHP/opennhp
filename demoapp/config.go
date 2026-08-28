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

// ResourceConfig describes a single protected resource the Demo UI shows
// after a successful listServices. The id must match what the nhp-server
// basic plugin's ListService returns; the rest (title, URL, acHost) is UI
// metadata that the JS Agent does not know about.
type ResourceConfig struct {
	ID     string `toml:"Id"`
	Title  string `toml:"Title"`
	URL    string `toml:"URL"`
	ACHost string `toml:"ACHost"`
}

// Config is the full TOML-loaded configuration for the demoapp daemon.
type Config struct {
	ListenAddr      string           `toml:"ListenAddr"`
	DBPath          string           `toml:"DBPath"`
	SessionKey      string           `toml:"SessionKey"`
	KeyEnvelopeKey  string           `toml:"KeyEnvelopeKey"` // base64, 32 bytes
	CipherScheme    CipherScheme     `toml:"CipherScheme"`
	RelayURL        string           `toml:"RelayUrl"`       // server-side: demoapp → relay
	PublicRelayURL  string           `toml:"PublicRelayUrl"` // browser-side: SPA → relay
	ServerPublicKey string           `toml:"ServerPublicKey"`
	ServiceID       string           `toml:"ServiceId"`
	OrganizationID  string           `toml:"OrganizationId"`
	WebDistDir      string           `toml:"WebDistDir"`
	OIDCs           []OIDCConfig     `toml:"OIDC"`
	Resources       []ResourceConfig `toml:"Resources"`
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
	return cfg, nil
}

// Validate enforces the post-autogen invariants: KeyEnvelopeKey decodes to
// exactly 32 bytes and CipherScheme is one of the supported values.
// Run this AFTER autoFillConfig so placeholders filled in by sidecar /
// random generation pass.
func (c *Config) Validate() error {
	if c.CipherScheme != CipherSchemeCurve25519 && c.CipherScheme != CipherSchemeGMSM {
		return fmt.Errorf("CipherScheme must be %q or %q (got empty or unknown — set it explicitly or let autoFillConfig detect from /servers)", CipherSchemeCurve25519, CipherSchemeGMSM)
	}
	if c.KeyEnvelopeKey == "" {
		return errors.New("KeyEnvelopeKey is required (32 bytes, base64-encoded)")
	}
	if _, err := c.KeyEnvelopeKeyBytes(); err != nil {
		return err
	}
	return nil
}
