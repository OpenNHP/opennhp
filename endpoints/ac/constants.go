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

	TokenStoreRefreshInterval = common.TokenStoreRefreshInterval
	TempPortOpenTime          = 30

	IPSET_DEFAULT_NAME      = "defaultset"
	IPSET_DEFAULT_DOWN_NAME = "defaultset_down"
)
