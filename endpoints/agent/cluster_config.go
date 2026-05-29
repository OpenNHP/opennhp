package agent

import (
	"github.com/OpenNHP/opennhp/nhp/common/clusterconfig"
)

// ClusterConfig and InstanceConfig are now defined in the shared
// nhp/common/clusterconfig package so nhp-ac and nhp-db can load the
// same TOML shape. The aliases below keep the agent-internal call sites
// (cluster.go, config.go, tests) and any downstream SDK consumers
// compiling unchanged.
type (
	ClusterConfig  = clusterconfig.ClusterConfig
	InstanceConfig = clusterconfig.InstanceConfig
)

// normalizeClusters wraps clusterconfig.Normalize with the
// agent-specific options. nhp-agent requires Name because resource.toml
// references clusters by name (Cluster = "...").
func normalizeClusters(clusters []*ClusterConfig, deprecate func(string, ...any)) error {
	return clusterconfig.Normalize(clusters, clusterconfig.Options{
		ConsumerLabel: "agent",
		RequireName:   true,
	}, deprecate)
}
