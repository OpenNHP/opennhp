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

// decompression
const (
	// MaxDecompressedBodySize caps plaintext produced from one compressed NHP
	// packet. The wire packet itself is limited to PacketBufferSize, so this is
	// a defense-in-depth ceiling sized for highly compressible legitimate
	// resource lists while keeping a crafted packet well below the previous
	// 10 MiB allocation limit.
	MaxDecompressedBodySize = 256 * PacketBufferSize // 1 MiB

	// MaxDecompressedBodyWarnSize marks an anomalous compression ratio close to
	// the hard ceiling. Warnings are throttled in decryptBody.
	MaxDecompressedBodyWarnSize = MaxDecompressedBodySize * 4 / 5
)

// session
const (
	MinimalRecvIntervalMs  = 20 // millisecond
	ThreatCountBeforeBlock = 1  // block at 2nd attempt
	FailureRetryInterval   = 10 // second
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

// Replay acceptance bounds also size endpoint dedupe retention.
const (
	DefaultRecvStalenessFloorSeconds = 600
	ARTRecvFutureSkewSeconds         = 300
	// RLYRecvStalenessFloorSeconds is the staleness floor applied
	// specifically to NHP_RLY envelopes. It is intentionally tighter
	// than DefaultRecvStalenessFloorSeconds (60 s vs 600 s) so the
	// relay replay cache can rely on eviction as its primary replay
	// defense: with this floor, an entry evicted by cache overflow at
	// LocalInitTime T satisfies
	//   remoteSendTime = T_admit < T - TTL  ≤  T - RLYRecvStalenessFloorSeconds
	// (since TTL > RLYRecvStalenessFloorSeconds in art_replay_cache.go),
	// so the staleness check rejects the corresponding replay at the
	// moment of eviction. Without the tight bound, the cache's
	// capacity-eviction window would overlap the staleness acceptance
	// window and a replay could succeed at the tail of an attacker's
	// flood. The bound also caps the relay→replay-acceptance window
	// itself: legitimate traffic with a slightly skewed relay clock
	// (>60 s behind) is rejected, which is acceptable for NHP_RLY
	// since the relay's authenticated timestamp is verified end-to-end
	// on each forward.
	RLYRecvStalenessFloorSeconds = 60
)
