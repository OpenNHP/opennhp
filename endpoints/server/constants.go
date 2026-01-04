package server

import "github.com/OpenNHP/opennhp/nhp/common"

const (
	MaxConcurrentConnection         = 20480
	OverloadConnectionThreshold     = MaxConcurrentConnection * 4 / 5 // 80%
	BlockAddrRefreshRate            = 20                              // 20 seconds
	BlockAddrExpireTime             = 90                              // 90 seconds
	PreCheckThreatCountBeforeBlock  = 5                               // block source address if packet precheck errors exceeds this count
	DefaultAgentConnectionTimeoutMs = common.ClientSideConnectionTimeoutMs
	DefaultACConnectionTimeoutMs    = common.ServerSideConnectionTimeoutMs
	DefaultDBConnectionTimeoutMs    = common.ServerSideConnectionTimeoutMs
	PacketQueueSizePerConnection    = 256
)

// http APIs
const (
	DefaultHttpRequestReadTimeoutMs   = 4500 // millisecond
	DefaultHttpResponseWriteTimeoutMs = 5500 // millisecond
	DefaultHttpServerIdleTimeoutMs    = 6000 // millisecond
)

// knock
const (
	DefaultIpOpenTime         = 120 // second, align with ipset default timeout
	ACOpenCompensationTime    = 5   // second
	TokenStoreRefreshInterval = common.TokenStoreRefreshInterval
)
