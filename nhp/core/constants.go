package core

// protocol
const ProtocolVersionMajor = 1
const ProtocolVersionMinor = 0

// device
const (
	MaxMemoryUsage         = 1 * 1024 * 1024 * 1024 // 1GB
	PacketBufferSize       = 4096
	PacketBufferPoolSize   = MaxMemoryUsage / PacketBufferSize
	AllocateTimeToOverload = 2 // 2 seconds
	SendQueueSize          = 10240
	RecvQueueSize          = 10240
)

// session
const (
	MinimalRecvIntervalMs  = 20 // millisecond
	ThreatCountBeforeBlock = 1  // block at 2nd attempt
	FailureRetryInterval   = 10 // second
)

// staleness floor (#1464) — the maximum age, in seconds, that a
// received packet's AEAD-authenticated send timestamp may lag the
// local receive time before the packet is rejected as stale in
// nhp/core/responder.go. See recvStalenessFloor for the dispatch.
const (
	// DefaultRecvStalenessFloorSeconds is the historical, generous
	// floor applied to every (deviceType, peerType, msgType) the
	// AOP-specific override below does not match. Kept at 600 s so
	// agent→server knocks (which may traverse the public internet
	// with looser clock-calibration assumptions) are byte-for-byte
	// unchanged by #1464.
	DefaultRecvStalenessFloorSeconds = 600

	// AOPRecvStalenessFloorSeconds is the tighter floor for NHP_AOP
	// (server→AC). recvStalenessFloor documents WHY AOP gets one (the
	// cross-restart replay window); this constant is the canonical home
	// for the value. AOP is normally exchanged between infrastructure
	// components with synchronized clocks, so 120 s provides margin for
	// wall-clock skew while still shrinking the replay window documented in
	// endpoints/ac/aop_replay_cache.go — comfortable margin against
	// clock-step-at-boot and receive-queue-delay false rejects, which for
	// AOP drop the packet but do NOT block the connection (see
	// shouldEscalateStale) — while still shrinking the replay window 5×
	// versus the 600 s default.
	AOPRecvStalenessFloorSeconds = 120
)

// transaction
const (
	AgentLocalTransactionResponseTimeoutMs  = 5 * 1000                                     // millisecond
	ServerLocalTransactionResponseTimeoutMs = AgentLocalTransactionResponseTimeoutMs - 300 // millisecond
	ACLocalTransactionResponseTimeoutMs     = ServerLocalTransactionResponseTimeoutMs      // millisecond

	RemoteTransactionProcessTimeoutMs   = 10 * 1000 // millisecond
	DELocalTransactionResponseTimeoutMs = 5 * 1000
)

// peer
const (
	MinimalPeerAddressHoldTime = 5 // second
)

// hostname resolve
const (
	MinimalNSLookupInterval = 300 // second
)

// packet
const (
	HeaderCommonSize      = 24
	SymmetricKeySize      = 32
	PrivateKeySize        = 32
	PublicKeySize         = 32
	PublicKeySizeEx       = 64
	HashSize              = 32
	CookieSize            = 32
	TimestampSize         = 8
	GCMNonceSize          = 12
	GCMTagSize            = 16
	PublicKeyBase64Size   = 44
	PublicKeyBase64SizeEx = 88
)

// noise
const (
	InitialChainKeyString = "NHP keygen v.20230421@clouddeep.cn"
	InitialHashString     = "NHP hashgen v.20230421@deepcloudsdp.com"
)
