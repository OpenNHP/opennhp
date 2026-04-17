package relay

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/common"
	log "github.com/OpenNHP/opennhp/nhp/log"
)

var ExeDirPath string

// Config holds all relay service configuration.
type Config struct {
	// HTTP listener
	ListenIP   string `toml:"listenIp"`
	ListenPort int    `toml:"listenPort"`

	// TLS (optional — set enableTLS = true for HTTPS)
	EnableTLS   bool   `toml:"enableTLS"`
	TLSCertFile string `toml:"tlsCertFile"`
	TLSKeyFile  string `toml:"tlsKeyFile"`

	// Relay identity: the relay's own private key (base64).
	// Generate with: ./nhp-relayd keygen --curve (or --sm2)
	PrivateKeyBase64 string `toml:"privateKeyBase64"`

	// Cipher scheme: 0 = Curve25519 (default), 1 = SM2/GMSM.
	CipherScheme int `toml:"cipherScheme"`

	// NHP Server upstream (UDP)
	NHPServerHost            string `toml:"nhpServerHost"`
	NHPServerPort            int    `toml:"nhpServerPort"`
	NHPServerPublicKeyBase64 string `toml:"nhpServerPublicKeyBase64"`

	// Keepalive interval in seconds (default: 20, matching AC behavior).
	KeepaliveIntervalS int `toml:"keepaliveIntervalS"`

	// Timeouts (milliseconds)
	ReadTimeoutMs  int `toml:"readTimeoutMs"`
	WriteTimeoutMs int `toml:"writeTimeoutMs"`
	IdleTimeoutMs  int `toml:"idleTimeoutMs"`
	UDPTimeoutMs   int `toml:"udpTimeoutMs"`

	// Logging
	LogLevel int `toml:"logLevel"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		ListenIP:           "0.0.0.0",
		ListenPort:         8080,
		NHPServerHost:      "127.0.0.1",
		NHPServerPort:      62206,
		KeepaliveIntervalS: common.ServerKeepaliveInterval,
		ReadTimeoutMs:      10000,
		WriteTimeoutMs:     10000,
		IdleTimeoutMs:      60000,
		UDPTimeoutMs:       5000,
		LogLevel:           2,
	}
}

// LoadConfig reads a TOML config file.  If path is empty it defaults to
// etc/config.toml relative to ExeDirPath.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		path = filepath.Join(ExeDirPath, "etc", "config.toml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warning("[Relay] config file %s not found, using defaults", path)
			return cfg, nil
		}
		return nil, fmt.Errorf("relay: failed to read config %s: %w", path, err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("relay: failed to parse config %s: %w", path, err)
	}

	if cfg.NHPServerHost == "" {
		return nil, fmt.Errorf("relay: nhpServerHost must be set in config")
	}
	if cfg.PrivateKeyBase64 == "" {
		return nil, fmt.Errorf("relay: privateKeyBase64 must be set in config")
	}
	if cfg.NHPServerPublicKeyBase64 == "" {
		return nil, fmt.Errorf("relay: nhpServerPublicKeyBase64 must be set in config")
	}

	log.Info("[Relay] loaded config from %s", path)
	return cfg, nil
}
