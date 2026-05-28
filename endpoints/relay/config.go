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

// ClusterInstance describes one physical nhp-server instance backing a logical
// cluster. Instances in the same cluster share a keypair and differ only in
// network address; the relay picks one at request time via load balancing.
//
// Phase 1 supports exactly one instance per cluster. The schema accepts a list
// so that operator configs and the API surface are stable when phase 2 lifts
// the restriction in nhp/core (multi-peer-per-pubkey support).
type ClusterInstance struct {
	Host   string `toml:"host"`
	Port   int    `toml:"port"`
	Weight int    `toml:"weight"`
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

// Cluster groups instances that share a single nhp-server identity (public
// key). Browsers and other relay clients address a cluster by its fingerprint
// (see utils.PubKeyFingerprint); the relay then routes inside the cluster.
type Cluster struct {
	// Optional human-readable label, surfaced in logs and the /clusters API.
	Name string `toml:"name"`

	// Base64-encoded public key shared by all instances in this cluster.
	PublicKeyBase64 string `toml:"publicKeyBase64"`

	// Load balancing strategy. Defaults to weighted-random.
	LoadBalance LoadBalanceScheme `toml:"loadBalance"`

	// One or more nhp-server addresses. Phase 1 requires exactly one.
	Instances []ClusterInstance `toml:"instance"`
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

	// NHP-server clusters. Each entry is one logical upstream identity
	// (one public key) that may be backed by multiple physical instances.
	Clusters []Cluster `toml:"cluster"`

	// Legacy single-server fields. When [[cluster]] is empty, LoadConfig
	// promotes these into Clusters[0] with a deprecation warning. When
	// [[cluster]] is also present, the legacy fields are ignored and a
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

	log.Info("[Relay] loaded config from %s with %d cluster(s)", path, len(cfg.Clusters))
	return cfg, nil
}

// normalize validates the configuration and applies legacy-field migration so
// that the rest of the relay only has to look at Config.Clusters. It is
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
	case hasLegacy && len(cfg.Clusters) == 0:
		// Auto-migrate: promote the legacy fields into a single cluster
		// so existing demo configs keep working through phase 1.
		log.Warning("[Relay] nhpServerHost/nhpServerPort/nhpServerPublicKeyBase64 are deprecated; " +
			"migrate to [[cluster]] / [[cluster.instance]] in config.toml")
		cfg.Clusters = []Cluster{{
			PublicKeyBase64: cfg.NHPServerPublicKeyBase64,
			Instances: []ClusterInstance{{
				Host: cfg.NHPServerHost,
				Port: cfg.NHPServerPort,
			}},
		}}
	case hasLegacy && len(cfg.Clusters) > 0:
		// Both forms present. The [[cluster]] block wins because it's
		// strictly more expressive; but a copy-paste upgrade that left
		// the old fields behind would silently route to whatever the
		// new block declares and drop the legacy values. Log loudly so
		// the operator notices and can remove the dead config.
		log.Warning("[Relay] both legacy nhpServer* fields and [[cluster]] blocks are set; " +
			"the legacy fields are ignored — remove them from config.toml to silence this warning")
	}

	if len(cfg.Clusters) == 0 {
		return fmt.Errorf("relay: no upstream configured; add at least one [[cluster]] with one [[cluster.instance]]")
	}

	seenFP := make(map[string]int, len(cfg.Clusters))
	// seenAddr catches two clusters that point at the same host:port.
	// resolveTarget keys by PeerPk, so duplicate addresses don't cause a
	// routing bug on their own — but they almost certainly mean the
	// operator copied a cluster and forgot to change the instance, and
	// silently accepting that would mask a real config mistake.
	type addrOrigin struct {
		cluster  int
		instance int
	}
	seenAddr := make(map[string]addrOrigin)
	for i := range cfg.Clusters {
		c := &cfg.Clusters[i]
		if c.PublicKeyBase64 == "" {
			return fmt.Errorf("relay: cluster #%d missing publicKeyBase64", i)
		}
		fp, err := utils.PubKeyFingerprintFromBase64(c.PublicKeyBase64)
		if err != nil {
			return fmt.Errorf("relay: cluster #%d publicKeyBase64 invalid: %w", i, err)
		}
		if dup, ok := seenFP[fp]; ok {
			return fmt.Errorf("relay: cluster #%d and #%d share the same publicKeyBase64 (fingerprint %s)", dup, i, fp)
		}
		seenFP[fp] = i

		if len(c.Instances) == 0 {
			return fmt.Errorf("relay: cluster #%d (fingerprint %s) has no [[cluster.instance]]", i, fp)
		}
		// Phase 1: at most one peer per cluster. The schema accepts
		// multiple [[cluster.instance]] blocks per cluster as a
		// forward-compat hook, but the runtime peer table is keyed
		// by pubkey only (device.peerMap; see core/device.go
		// LookupPeer), so sibling instances would silently overwrite
		// each other at AddPeer time and any Noise-level source-IP
		// validation against LookupPeer().Ip would reject responses
		// from the surviving-but-mismatched sibling. Reject the
		// config now rather than letting the operator discover this
		// by chasing dropped responses.
		if len(c.Instances) > 1 {
			return fmt.Errorf("relay: cluster #%d (fingerprint %s) declares %d instances; multi-instance clusters are not supported yet (the device peer table is keyed by pubkey only). Split into one cluster per identity, or declare a single instance",
				i, fp, len(c.Instances))
		}
		for j := range c.Instances {
			inst := &c.Instances[j]
			if inst.Host == "" {
				return fmt.Errorf("relay: cluster #%d instance #%d missing host", i, j)
			}
			if inst.Port <= 0 {
				return fmt.Errorf("relay: cluster #%d instance #%d missing or invalid port", i, j)
			}
			addr := fmt.Sprintf("%s:%d", inst.Host, inst.Port)
			if dup, ok := seenAddr[addr]; ok {
				return fmt.Errorf("relay: cluster #%d instance #%d address %s already claimed by cluster #%d instance #%d",
					i, j, addr, dup.cluster, dup.instance)
			}
			seenAddr[addr] = addrOrigin{cluster: i, instance: j}
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
			return fmt.Errorf("relay: cluster #%d (fingerprint %s) loadBalance %q is not one of: %q, %q, %q",
				i, fp, c.LoadBalance, LBRandom, LBWeightedRandom, LBRoundRobin)
		}
	}

	return nil
}
