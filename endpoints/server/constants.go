package server

import "github.com/OpenNHP/opennhp/nhp/common"

const (
	MaxConcurrentConnection     = 20480
	OverloadConnectionThreshold = MaxConcurrentConnection * 4 / 5 // 80%
	MaxConnectionsPerRelay      = 1024                            // per-relay-peer cap on forwarded-client fan-out

	// MaxConcurrentHandlers bounds in-flight handler goroutines spawned
	// per inbound NHP packet (KNK, RKN, EXT, RLY, OTP, REG, LST, DAR,
	// DRG, DAV). MaxConcurrentConnection caps unique remote-addr entries
	// in the connection table; it does NOT cap how many handshake-class
	// packets a single peer can have in flight concurrently, so without
	// this an attacker controlling one relay/agent can drive the server
	// into OOM well before the connection cap matters. The cookie path
	// is designed to make cheap CPU rejection possible — this cap is
	// what keeps that design from being undermined by an unbounded
	// goroutine spawn. When the semaphore is full, packets are dropped
	// at recvMessageRoutine so the device-receive queue stays drainable.
	MaxConcurrentHandlers           = 4096
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

// overload RKN rate limiting (see rknRateLimiter)
const (
	// OverloadRknRatePerSecondPerIP / OverloadRknBurstPerIP size the
	// per-source-IP token bucket that gates the expensive cookie-verify
	// ECDH while the server is overloaded. A legitimate agent reknocks at
	// most a handful of times per cookie window (default 60s), so 20/s
	// with a burst of 40 is orders of magnitude above any honest cadence
	// while still capping an attacker's forced ECDH rate per address.
	OverloadRknRatePerSecondPerIP = 20
	OverloadRknBurstPerIP         = 40
	// OverloadRknLimiterMaxEntries bounds the limiter's per-IP table so a
	// spoofed-source flood can't grow it without bound. At ~48 bytes per
	// entry this is a few MB worst case, and it is reclaimed by sweeps and
	// oldest-eviction (see rknRateLimiter.evict).
	OverloadRknLimiterMaxEntries = 65536
	// OverloadRknLimiterIdleSeconds: a bucket untouched this long is
	// reclaimable. Set above a couple of cookie windows so an agent that
	// goes briefly quiet keeps its bucket.
	OverloadRknLimiterIdleSeconds = 300
)
