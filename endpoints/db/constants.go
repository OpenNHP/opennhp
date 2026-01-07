package db

import "github.com/OpenNHP/opennhp/nhp/common"

const (
	MaxConcurrentConnection      = 256
	DefaultConnectionTimeoutMs   = common.ClientSideConnectionTimeoutMs
	PacketQueueSizePerConnection = 64     // nhp db does not need large transactions
	DoType_Default               = "ZTDO" // The DHP protocol enforces encryption by default, and its core data unit is the Zero Trust Data Object (ZTDO)
	DoType_Other                 = "OTHER"

	ReportToServerInterval         = common.ReportToServerInterval
	MinialServerDiscoveryInterval  = common.MinimalServerDiscoveryInterval
	ServerKeepaliveInterval        = common.ServerKeepaliveInterval
	ServerDiscoveryRetryBeforeFail = common.ServerDiscoveryRetryBeforeFail
)
