package agent

import (
	"fmt"

	"github.com/OpenNHP/opennhp/nhp/common/loadbalance"
)

// ClusterConfig is the TOML shape of one [[Servers]] entry. Two forms
// are accepted:
//
//  1. **Cluster form (recommended):** set Instances (a list of one or
//     more nhp-server endpoints sharing this pubkey), plus LoadBalance
//     and optional Sticky / ExpireTime at the cluster level.
//
//  2. **Legacy flat form (auto-upgraded):** leave Instances empty and
//     populate Hostname / Ip / Port at the top level. LoadClusters
//     wraps these as a single-instance cluster and logs a deprecation
//     warning. Existing single-server agent configs keep working with
//     no edits.
//
// Mixing forms in the same entry (both Instances and top-level Ip)
// surfaces a validation error rather than being silently merged.
type ClusterConfig struct {
	// Cluster-level fields.
	Name           string             `toml:"Name"`
	PubKeyBase64   string             `toml:"PubKeyBase64"`
	LoadBalance    loadbalance.Scheme `toml:"LoadBalance"`
	StickyInstance *bool              `toml:"StickyInstance"`
	ExpireTime     int64              `toml:"ExpireTime"`
	Instances      []InstanceConfig   `toml:"Instances"`

	// Legacy top-level fields (single-server form). Carry the same
	// names the original schema used so existing server.toml files
	// keep parsing unchanged.
	Hostname string `toml:"Hostname"`
	Ip       string `toml:"Ip"`
	Port     int    `toml:"Port"`
}

// InstanceConfig is one physical nhp-server endpoint inside a
// ClusterConfig.Instances list. At least one of Host or Ip must be
// non-empty; Port is required.
type InstanceConfig struct {
	Host   string `toml:"Host"`
	Ip     string `toml:"Ip"`
	Port   int    `toml:"Port"`
	Weight int    `toml:"Weight"`
}

// stickyOrDefault returns the configured sticky flag, defaulting to
// true. Sticky keeps a KnockTarget bound to the same instance across
// its KNK → COK → RKN handshake; operators with a known-good
// stateless-cookie cluster can flip it off to spread retries across
// instances.
func (c *ClusterConfig) stickyOrDefault() bool {
	if c.StickyInstance == nil {
		return true
	}
	return *c.StickyInstance
}

// hasLegacyFields reports whether this entry was written in the old
// flat single-server form. Used to drive auto-upgrade in
// normalizeClusters.
func (c *ClusterConfig) hasLegacyFields() bool {
	return c.Hostname != "" || c.Ip != "" || c.Port != 0
}

// (Peers is declared in config.go; we extend it below via the
// ClusterConfig type referenced from there.)

// normalizeClusters validates every [[Servers]] entry and applies the
// legacy auto-upgrade. After it returns, every cluster has at least
// one Instances entry and a normalised LoadBalance. Returned errors
// are surfaced to the operator at load time; warnings (deprecation,
// recoverable inconsistencies) go through deprecate.
//
// deprecate is a callback so callers control how the warning is
// surfaced — production code passes log.Warning; tests can capture
// invocations for assertions.
func normalizeClusters(clusters []*ClusterConfig, deprecate func(string, ...any)) error {
	if len(clusters) == 0 {
		return fmt.Errorf("agent: no [[Servers]] configured")
	}
	for i, c := range clusters {
		if c == nil {
			return fmt.Errorf("agent: [[Servers]][%d] is nil", i)
		}
		if c.PubKeyBase64 == "" {
			return fmt.Errorf("agent: [[Servers]][%d] missing PubKeyBase64", i)
		}

		legacy := c.hasLegacyFields()
		hasInstances := len(c.Instances) > 0

		switch {
		case legacy && hasInstances:
			// Both forms in one entry is almost certainly an
			// incomplete migration. Refuse to guess which one the
			// operator meant.
			return fmt.Errorf("agent: [[Servers]][%d] (%s) sets both top-level Ip/Hostname/Port and [[Servers.Instances]]; "+
				"pick one form — top-level fields are deprecated, prefer Instances",
				i, c.PubKeyBase64)
		case legacy && !hasInstances:
			deprecate("agent: [[Servers]][%d] uses legacy single-server form (Hostname/Ip/Port at top level); "+
				"migrate to [[Servers.Instances]] in server.toml — auto-upgrading for now",
				i)
			c.Instances = []InstanceConfig{{
				Host:   c.Hostname,
				Ip:     c.Ip,
				Port:   c.Port,
				Weight: 1,
			}}
			// Zero the legacy fields so downstream code never sees both.
			c.Hostname, c.Ip, c.Port = "", "", 0
		case !legacy && !hasInstances:
			return fmt.Errorf("agent: [[Servers]][%d] (%s) has no instances", i, c.PubKeyBase64)
		}

		if err := c.LoadBalance.Validate(); err != nil {
			return fmt.Errorf("agent: [[Servers]][%d] (%s): %w", i, c.PubKeyBase64, err)
		}
		c.LoadBalance = c.LoadBalance.Normalize()

		for j := range c.Instances {
			inst := &c.Instances[j]
			if inst.Host == "" && inst.Ip == "" {
				return fmt.Errorf("agent: [[Servers]][%d] instance #%d: Host and Ip both empty", i, j)
			}
			if inst.Port <= 0 {
				return fmt.Errorf("agent: [[Servers]][%d] instance #%d: invalid Port %d", i, j, inst.Port)
			}
			if inst.Weight <= 0 {
				inst.Weight = 1
			}
		}
	}

	// Duplicate pubkey detection — two clusters with the same pubkey
	// would race for the same slot in device.peerMap. Catch it here
	// so the error is obvious at load rather than as a mysterious
	// "wrong instance answered" at runtime.
	seen := make(map[string]int, len(clusters))
	for i, c := range clusters {
		if prev, ok := seen[c.PubKeyBase64]; ok {
			return fmt.Errorf("agent: [[Servers]][%d] and [[Servers]][%d] share PubKeyBase64 %s — "+
				"merge them into one cluster with multiple Instances",
				prev, i, c.PubKeyBase64)
		}
		seen[c.PubKeyBase64] = i
	}
	return nil
}
