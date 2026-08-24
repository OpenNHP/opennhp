// Package demoapp implements the standalone OpenNHP Demo application.
//
// The Demo App hosts user accounts (username/password or OIDC) and drives
// the NHP key registration flow through js-agent in the browser. Private
// keys never leave the browser — they are encrypted with a PBKDF2/AES-GCM
// blob keyed on the user's password and stored alongside the public key.
package demoapp

import (
	"fmt"
	"os"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

// Config is the root TOML configuration for the Demo App.
type Config struct {
	ListenAddr     string         `toml:"listenAddr"`
	DbPath         string         `toml:"dbPath"`
	SessionSecret  string         `toml:"sessionSecret"`
	InternalApiKey string         `toml:"internalApiKey"`
	OtpTTLMinutes  int            `toml:"otpTTLMinutes"`
	LogLevel       int            `toml:"logLevel"`
	SMTP           SMTPConfig     `toml:"smtp"`
	OIDC           OIDCConfig     `toml:"oidc"`
	NHP            NHPPublicConfig `toml:"nhp"`
}

// SMTPConfig — when Mode == "console" (default for dev), OTP codes are
// printed to stdout and email is not sent.
type SMTPConfig struct {
	Mode       string `toml:"mode"` // "console" | "smtp"
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	Username   string `toml:"username"`
	Password   string `toml:"password"`
	From       string `toml:"from"`
	Subject    string `toml:"subject"`
	CodeInSubject *bool `toml:"codeInSubject"`
}

// OIDCConfig — populated only when OIDC login is enabled (Enabled == true).
type OIDCConfig struct {
	Enabled      bool   `toml:"enabled"`
	IssuerURL    string `toml:"issuerUrl"`
	ClientID     string `toml:"clientId"`
	ClientSecret string `toml:"clientSecret"`
	RedirectURL  string `toml:"redirectUrl"`
	Scopes       string `toml:"scopes"` // space-separated; defaults to "openid email profile"
}

// NHPPublicConfig is the configuration returned to the browser for js-agent.
// ServerPublicKey is the Curve25519 base64 public key of the nhp-server.
type NHPPublicConfig struct {
	RelayUrl        string `toml:"relayUrl"`
	ServerPublicKey string `toml:"serverPublicKey"`
	AuthServiceId   string `toml:"authServiceId"`
	CipherScheme    string `toml:"cipherScheme"` // "curve25519" or "gmsm"
}

// LoadConfig parses the TOML file at path. Missing fields fall back to safe
// defaults so a fresh checkout can run locally with `demo-app --config x.toml`.
func LoadConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := toml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.ListenAddr == "" {
		c.ListenAddr = ":8088"
	}
	if c.DbPath == "" {
		c.DbPath = "data/demo.db"
	}
	if c.SessionSecret == "" {
		// dev fallback — operators must override in production
		c.SessionSecret = "dev-only-secret-change-me-32bytes!"
	}
	if c.OtpTTLMinutes == 0 {
		c.OtpTTLMinutes = 5
	}
	if c.LogLevel == 0 {
		c.LogLevel = 5
	}
	if c.SMTP.Mode == "" {
		c.SMTP.Mode = "console"
	}
	if c.SMTP.Subject == "" {
		c.SMTP.Subject = "Your OpenNHP Demo verification code"
	}
	if c.SMTP.CodeInSubject == nil {
		v := true
		c.SMTP.CodeInSubject = &v
	}
	if c.OIDC.Scopes == "" {
		c.OIDC.Scopes = "openid email profile"
	}
	if c.NHP.AuthServiceId == "" {
		c.NHP.AuthServiceId = "demo-app"
	}
	if c.NHP.CipherScheme == "" {
		c.NHP.CipherScheme = "curve25519"
	}
	if c.NHP.CipherScheme != "curve25519" && c.NHP.CipherScheme != "gmsm" {
		c.NHP.CipherScheme = "curve25519"
	}
}

func (c *Config) validate() error {
	if c.InternalApiKey == "" || strings.Contains(c.InternalApiKey, "change-me") {
		// In production the InternalApiKey MUST be set; we only warn at startup.
	}
	if c.NHP.ServerPublicKey == "" {
		return fmt.Errorf("nhp.serverPublicKey is required")
	}
	if c.NHP.RelayUrl == "" {
		return fmt.Errorf("nhp.relayUrl is required")
	}
	if c.OIDC.Enabled {
		if c.OIDC.IssuerURL == "" || c.OIDC.ClientID == "" || c.OIDC.RedirectURL == "" {
			return fmt.Errorf("oidc.enabled=true requires issuerUrl, clientId, redirectUrl")
		}
	}
	return nil
}

// OtpTTL exposes the configured OTP TTL as a Go duration.
func (c *Config) OtpTTL() time.Duration {
	return time.Duration(c.OtpTTLMinutes) * time.Minute
}
