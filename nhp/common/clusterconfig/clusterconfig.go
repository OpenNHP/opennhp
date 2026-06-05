// Package clusterconfig defines the shared TOML shape every NHP UDP
// client (nhp-agent, nhp-ac, nhp-db) uses to describe an upstream
// nhp-server peer.
//
// Before this package existed each endpoint carried its own server-peer
// struct with subtly different field names: nhp-ac shipped a single
// "Endpoints []string" form, nhp-db only ever exposed a single flat
// Hostname/Ip/Port row, and nhp-agent grew the structured
// [[Servers.Instances]] form first. Operators had to learn a new
// dialect for each daemon and the documentation lived in three places.
//
// One [[Servers]] entry is one logical nhp-server identity (one pubkey)
// reachable at 1..N physical instances that share that keypair. The
// schema accepts two forms:
//
//  1. Cluster form (recommended): set Instances and optionally
//     LoadBalance / StickyInstance / ExpireTime at the cluster level.
//
//  2. Legacy flat form (auto-upgraded with a deprecation warning):
//     leave Instances empty and populate Hostname / Ip / Port at the
//     entry root. Normalize wraps these as a single-instance cluster
//     so existing single-server configs keep working with no edits.
//
// Mixing forms in the same entry (both Instances and top-level Ip)
// surfaces a validation error rather than being silently merged.
//
// Consumers that don't need every field (nhp-db ignores LoadBalance /
// StickyInstance because it only ever picks one instance; nhp-ac AOL-
// broadcasts and therefore also ignores LoadBalance) still load through
// this package so the operator-facing TOML stays identical across
// daemons.
package clusterconfig

import (
	"fmt"
	"regexp"

	"github.com/OpenNHP/opennhp/nhp/common/loadbalance"
)

// clusterNameRegex constrains cluster names to a TOML- and shell-safe
// subset. Names appear unquoted in resource.toml (Cluster = "...") and
// in log lines, so we forbid whitespace, '/', and quoting characters
// up front rather than discovering the encoding bugs at runtime.
var clusterNameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// NameMaxLen caps the cluster name length. Exposed so callers can
// surface it in their own error messages if they reject names earlier.
const NameMaxLen = 64

// ClusterConfig is the TOML shape of one [[Servers]] entry. See the
// package doc for the two accepted forms.
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

// StickyOrDefault returns the configured sticky flag, defaulting to
// true. Sticky keeps a knock bound to the same instance across its
// KNK → COK → RKN handshake; operators with a known-good
// stateless-cookie cluster can flip it off to spread retries across
// instances.
//
// Consumers that don't run a multi-step handshake (nhp-ac, nhp-db)
// should just ignore the field — Normalize still loads it so the same
// TOML works across daemons.
func (c *ClusterConfig) StickyOrDefault() bool {
	if c.StickyInstance == nil {
		return true
	}
	return *c.StickyInstance
}

// hasLegacyFields reports whether this entry was written in the old
// flat single-server form. Drives auto-upgrade in Normalize.
func (c *ClusterConfig) hasLegacyFields() bool {
	return c.Hostname != "" || c.Ip != "" || c.Port != 0
}

// Options tunes per-consumer validation. Different daemons disagree on
// which fields are required:
//
//   - nhp-agent's resource.toml references clusters by Name, so Name is
//     required and must match the safe charset.
//   - nhp-ac and nhp-db just need a pubkey + at least one instance;
//     Name is optional (log label only).
type Options struct {
	// ConsumerLabel is the daemon name used as an error prefix
	// ("agent", "ac", "db"). Used only for error messages.
	ConsumerLabel string

	// RequireName enforces that every entry has a non-empty Name in
	// the safe charset. Used by nhp-agent where resource.toml's
	// Cluster = "..." references the field. Other daemons leave this
	// false; an absent Name is then accepted as a log-only label.
	RequireName bool
}

// Normalize validates every [[Servers]] entry and applies the legacy
// auto-upgrade in-place. After it returns, every cluster has at least
// one Instances entry and a normalised LoadBalance. Returned errors are
// surfaced to the operator at load time; warnings (deprecation,
// recoverable inconsistencies) go through deprecate.
//
// deprecate is a callback so callers control how the warning is
// surfaced — production code passes log.Warning; tests can capture
// invocations for assertions. A nil deprecate is treated as no-op.
func Normalize(clusters []*ClusterConfig, opts Options, deprecate func(string, ...any)) error {
	if deprecate == nil {
		deprecate = func(string, ...any) {}
	}
	label := opts.ConsumerLabel
	if label == "" {
		label = "cluster"
	}

	if len(clusters) == 0 {
		return fmt.Errorf("%s: no [[Servers]] configured", label)
	}
	for i, c := range clusters {
		if c == nil {
			return fmt.Errorf("%s: [[Servers]][%d] is nil", label, i)
		}
		if c.PubKeyBase64 == "" {
			return fmt.Errorf("%s: [[Servers]][%d] missing PubKeyBase64", label, i)
		}
		if opts.RequireName {
			if c.Name == "" {
				return fmt.Errorf("%s: [[Servers]][%d] (%s) missing Name — clusters are referenced from resource.toml by Name",
					label, i, c.PubKeyBase64)
			}
			if len(c.Name) > NameMaxLen {
				return fmt.Errorf("%s: [[Servers]][%d] Name %q exceeds %d chars",
					label, i, c.Name, NameMaxLen)
			}
			if !clusterNameRegex.MatchString(c.Name) {
				return fmt.Errorf("%s: [[Servers]][%d] Name %q invalid — allowed chars: [a-zA-Z0-9._-]",
					label, i, c.Name)
			}
		}

		legacy := c.hasLegacyFields()
		hasInstances := len(c.Instances) > 0

		switch {
		case legacy && hasInstances:
			// Both forms in one entry is almost certainly an
			// incomplete migration. Refuse to guess which one the
			// operator meant.
			return fmt.Errorf("%s: [[Servers]][%d] (%s) sets both top-level Ip/Hostname/Port and [[Servers.Instances]]; "+
				"pick one form — top-level fields are deprecated, prefer Instances",
				label, i, c.PubKeyBase64)
		case legacy && !hasInstances:
			deprecate("%s: [[Servers]][%d] uses legacy single-server form (Hostname/Ip/Port at top level); "+
				"migrate to [[Servers.Instances]] in server.toml — auto-upgrading for now",
				label, i)
			c.Instances = []InstanceConfig{{
				Host:   c.Hostname,
				Ip:     c.Ip,
				Port:   c.Port,
				Weight: 1,
			}}
			// Zero the legacy fields so downstream code never sees both.
			c.Hostname, c.Ip, c.Port = "", "", 0
		case !legacy && !hasInstances:
			return fmt.Errorf("%s: [[Servers]][%d] (%s) has no instances", label, i, c.PubKeyBase64)
		}

		if err := c.LoadBalance.Validate(); err != nil {
			return fmt.Errorf("%s: [[Servers]][%d] (%s): %w", label, i, c.PubKeyBase64, err)
		}
		c.LoadBalance = c.LoadBalance.Normalize()

		for j := range c.Instances {
			inst := &c.Instances[j]
			if inst.Host == "" && inst.Ip == "" {
				return fmt.Errorf("%s: [[Servers]][%d] instance #%d: Host and Ip both empty", label, i, j)
			}
			if inst.Port <= 0 {
				return fmt.Errorf("%s: [[Servers]][%d] instance #%d: invalid Port %d", label, i, j, inst.Port)
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
	seenPK := make(map[string]int, len(clusters))
	for i, c := range clusters {
		if prev, ok := seenPK[c.PubKeyBase64]; ok {
			return fmt.Errorf("%s: [[Servers]][%d] and [[Servers]][%d] share PubKeyBase64 %s — "+
				"merge them into one cluster with multiple Instances",
				label, prev, i, c.PubKeyBase64)
		}
		seenPK[c.PubKeyBase64] = i
	}

	// Duplicate Name detection — only meaningful when names are
	// required (and therefore non-empty). For consumers that leave Name
	// optional, blank names are allowed to repeat.
	if opts.RequireName {
		seenName := make(map[string]int, len(clusters))
		for i, c := range clusters {
			if prev, ok := seenName[c.Name]; ok {
				return fmt.Errorf("%s: [[Servers]][%d] and [[Servers]][%d] share Name %q",
					label, prev, i, c.Name)
			}
			seenName[c.Name] = i
		}
	}
	return nil
}
