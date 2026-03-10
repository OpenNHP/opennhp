package ac

import "github.com/OpenNHP/opennhp/nhp/common"

const (
	MaxConcurrentConnection      = 256
	DefaultConnectionTimeoutMs   = common.ServerSideConnectionTimeoutMs
	PacketQueueSizePerConnection = 256

	ReportToServerInterval         = common.ReportToServerInterval
	MinialServerDiscoveryInterval  = common.MinimalServerDiscoveryInterval
	ServerKeepaliveInterval        = common.ServerKeepaliveInterval
	ServerDiscoveryRetryBeforeFail = common.ServerDiscoveryRetryBeforeFail

	// StartupGracePeriod is the time after AC startup during which it stays
	// in deny-all mode even if all server discoveries fail. This prevents
	// fail-open during container orchestration when the server hasn't started yet.
	StartupGracePeriod = 60 // seconds

	TokenStoreRefreshInterval = common.TokenStoreRefreshInterval
	TempPortOpenTime          = 30

	IPSET_DEFAULT_NAME      = "defaultset"
	IPSET_DEFAULT_DOWN_NAME = "defaultset_down"
)
