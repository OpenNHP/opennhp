package server

import "github.com/OpenNHP/opennhp/nhp/common"

const (
	MaxConcurrentConnection         = 20480
	OverloadConnectionThreshold     = MaxConcurrentConnection * 4 / 5      // 80%
	MaxConnectionsPerRelay          = 1024                                 // per-relay-peer cap on forwarded-client fan-out
	BlockAddrRefreshRate            = 20                                   // 20 seconds
	BlockAddrExpireTime             = 90                                   // 90 seconds
	PreCheckThreatCountBeforeBlock  = 5                                    // block source address if packet precheck errors exceeds this count
	DefaultAgentConnectionTimeoutMs = common.ClientSideConnectionTimeoutMs // 30 seconds to delete idle connection
	DefaultACConnectionTimeoutMs    = common.ServerSideConnectionTimeoutMs // 300 seconds to delete idle connection
	DefaultDBConnectionTimeoutMs    = common.ServerSideConnectionTimeoutMs // 300 seconds to delete idle connection
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

// cookie
const (
	// DefaultCookieTimeWindowSeconds is the rolling window used for
	// stateless cookie derivation when Config.CookieTimeWindowSeconds is
	// unset or non-positive. Verification accepts current + previous
	// window, so an agent has [60, 120) seconds to redeem a cookie.
	DefaultCookieTimeWindowSeconds = 60
)
