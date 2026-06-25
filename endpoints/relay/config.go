package relay

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/common/loadbalance"
	log "github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/utils"
)

var ExeDirPath string

// ServerInstance describes one physical nhp-server endpoint backing a logical
// Server entry. Sibling instances in the same Server share a keypair and
// differ only in network address; the relay picks one at request time via
// load balancing.
//
// Field names + TOML tags mirror nhp/common/clusterconfig.InstanceConfig so
// nhp-relay, nhp-agent, nhp-ac, and nhp-db share one TOML dialect for
// describing an upstream nhp-server.
type ServerInstance struct {
	Host   string `toml:"Host"`
	Port   int    `toml:"Port"`
	Weight int    `toml:"Weight"`
}

// LoadBalanceScheme is an alias for the shared loadbalance.Scheme; kept
// under the old name so existing relay references compile unchanged.
// The string values are identical to those documented in TOML configs.
type LoadBalanceScheme = loadbalance.Scheme

const (
	LBRandom         = loadbalance.SchemeRandom
	LBWeightedRandom = loadbalance.SchemeWeightedRandom
	LBRoundRobin     = loadbalance.SchemeRoundRobin
)

// Server is one logical nhp-server identity (one pubkey) reachable at 1..N
// physical instances that share that keypair. Relay clients address a
// Server by the fingerprint of its public key (see utils.PubKeyFingerprint)
// over the HTTP path POST /relay/<serverId>; the relay then routes inside
// the Server to a specific instance per request.
//
// The TOML shape and field names match nhp-agent / nhp-ac / nhp-db's
// [[Servers]] schema (see nhp/common/clusterconfig.ClusterConfig) so the
// same block can be copied between the relay's config.toml and a peer
// table's server.toml.
type Server struct {
	// Optional human-readable label, surfaced in logs and the /servers API.
	Name string `toml:"Name"`

	// Base64-encoded public key shared by all instances backing this Server.
	PubKeyBase64 string `toml:"PubKeyBase64"`

	// Load balancing strategy. Defaults to weighted-random.
	LoadBalance LoadBalanceScheme `toml:"LoadBalance"`

	// StickyInstance (default true) keeps the same real-client IP routed
	// to the same instance via hash-based selection. When false, each
	// request is load-balanced independently. Mirrors
	// clusterconfig.ClusterConfig.StickyInstance.
	StickyInstance *bool `toml:"StickyInstance"`

	// Epoch seconds at which this Server's pubkey is considered
	// expired. Mirrors the same-named field on nhp-agent/ac/db so
	// operators can reuse one ExpireTime across all peer tables.
	// Currently advisory: the relay propagates it to the device peer
	// for noise-handshake validation, the same way nhp-ac does.
	ExpireTime int64 `toml:"ExpireTime"`

	// One or more nhp-server addresses backing this Server's pubkey.
	Instances []ServerInstance `toml:"Instances"`
}

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

	// Upstream nhp-server entries. Each entry is one logical upstream
	// identity (one public key) that may be backed by multiple physical
	// instances. Same schema as nhp-agent/nhp-ac/nhp-db's server.toml
	// (see nhp/common/clusterconfig).
	Servers []Server `toml:"Servers"`

	// Legacy single-server fields. When [[Servers]] is empty, LoadConfig
	// promotes these into Servers[0] with a deprecation warning. When
	// [[Servers]] is also present, the legacy fields are ignored and a
	// louder warning is logged so the operator notices the dead values.
	// Slated for removal in a future release.
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

	if err := cfg.normalize(); err != nil {
		return nil, err
	}

	log.Info("[Relay] loaded config from %s with %d server(s)", path, len(cfg.Servers))
	return cfg, nil
}

// normalize validates the configuration and applies legacy-field migration so
// that the rest of the relay only has to look at Config.Servers. It is
// exported as a method (not a function) to make it directly testable on a
// hand-built Config in unit tests without round-tripping through TOML.
func (cfg *Config) normalize() error {
	if cfg.PrivateKeyBase64 == "" {
		return fmt.Errorf("relay: privateKeyBase64 must be set in config")
	}

	hasLegacy := cfg.NHPServerHost != "" ||
		cfg.NHPServerPort != 0 ||
		cfg.NHPServerPublicKeyBase64 != ""

	switch {
	case hasLegacy && len(cfg.Servers) == 0:
		// Auto-migrate: promote the legacy fields into a single server
		// so existing demo configs keep working through phase 1.
		log.Warning("[Relay] nhpServerHost/nhpServerPort/nhpServerPublicKeyBase64 are deprecated; " +
			"migrate to [[Servers]] / [[Servers.Instances]] in config.toml")
		cfg.Servers = []Server{{
			PubKeyBase64: cfg.NHPServerPublicKeyBase64,
			Instances: []ServerInstance{{
				Host: cfg.NHPServerHost,
				Port: cfg.NHPServerPort,
			}},
		}}
	case hasLegacy && len(cfg.Servers) > 0:
		// Both forms present. The [[Servers]] block wins because it's
		// strictly more expressive; but a copy-paste upgrade that left
		// the old fields behind would silently route to whatever the
		// new block declares and drop the legacy values. Log loudly so
		// the operator notices and can remove the dead config.
		log.Warning("[Relay] both legacy nhpServer* fields and [[Servers]] blocks are set; " +
			"the legacy fields are ignored — remove them from config.toml to silence this warning")
	}

	if len(cfg.Servers) == 0 {
		return fmt.Errorf("relay: no upstream configured; add at least one [[Servers]] with one [[Servers.Instances]]")
	}

	seenFP := make(map[string]int, len(cfg.Servers))
	// seenAddr catches a server+instance pair duplicated under the SAME
	// pubkey — the "operator copied a [[Servers]] block and forgot to
	// change the instance" mistake. The dedupe key is (fingerprint, addr),
	// NOT addr alone: resolveTarget routes by PeerPk, so two DISTINCT
	// pubkeys legitimately sharing one host:port (a SNI/header-routed
	// front-end, or port-multiplexed identities) is a valid topology and
	// must not be a hard config-load failure. Only same-pubkey + same-addr
	// is the unambiguous copy-paste error.
	type addrOrigin struct {
		server   int
		instance int
	}
	seenAddr := make(map[string]addrOrigin)
	for i := range cfg.Servers {
		c := &cfg.Servers[i]
		if c.PubKeyBase64 == "" {
			return fmt.Errorf("relay: server #%d missing publicKeyBase64", i)
		}
		fp, err := utils.PubKeyFingerprintFromBase64(c.PubKeyBase64)
		if err != nil {
			return fmt.Errorf("relay: server #%d publicKeyBase64 invalid: %w", i, err)
		}
		if dup, ok := seenFP[fp]; ok {
			return fmt.Errorf("relay: server #%d and #%d share the same publicKeyBase64 (fingerprint %s)", dup, i, fp)
		}
		seenFP[fp] = i

		if len(c.Instances) == 0 {
			return fmt.Errorf("relay: server #%d (fingerprint %s) has no [[Servers.Instances]]", i, fp)
		}
		for j := range c.Instances {
			inst := &c.Instances[j]
			if inst.Host == "" {
				return fmt.Errorf("relay: server #%d instance #%d missing host", i, j)
			}
			if inst.Port <= 0 {
				return fmt.Errorf("relay: server #%d instance #%d missing or invalid port", i, j)
			}
			addr := fmt.Sprintf("%s:%d", inst.Host, inst.Port)
			// Scope to this server's pubkey: same identity reusing an
			// address is the copy-paste error we reject; a sibling
			// identity on the same address is allowed (see seenAddr docs).
			addrKey := fp + "@" + addr
			if dup, ok := seenAddr[addrKey]; ok {
				return fmt.Errorf("relay: server #%d instance #%d address %s already claimed by server #%d instance #%d under the same publicKeyBase64 (fingerprint %s)",
					i, j, addr, dup.server, dup.instance, fp)
			}
			seenAddr[addrKey] = addrOrigin{server: i, instance: j}
			if inst.Weight <= 0 {
				inst.Weight = 1
			}
		}
		switch c.LoadBalance {
		case "":
			c.LoadBalance = LBWeightedRandom
		case LBRandom, LBWeightedRandom, LBRoundRobin:
			// known scheme, keep as-is
		default:
			// Typos like "weighted_random" or "roundrobin" are harmless in
			// phase 1 (the value is unused with a single instance) but
			// would silently degrade phase-2 load balancing to whatever
			// the default policy is. Reject at load time so the operator
			// hears about it now, not after a later upgrade.
			return fmt.Errorf("relay: server #%d (fingerprint %s) loadBalance %q is not one of: %q, %q, %q",
				i, fp, c.LoadBalance, LBRandom, LBWeightedRandom, LBRoundRobin)
		}
	}

	return nil
}
