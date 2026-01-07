package agent

import "github.com/OpenNHP/opennhp/nhp/common"

const (
	MaxConcurrentConnection      = 256
	DefaultConnectionTimeoutMs   = common.ClientSideConnectionTimeoutMs
	PacketQueueSizePerConnection = 64 // nhp agent does not need large transactions
)
