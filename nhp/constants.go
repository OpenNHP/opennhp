package nhp

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
	MinimalRecvIntervalMs  = 20  // millisecond
	ThreatCountBeforeBlock = 1   // block at 2nd attempt
	CookieRegenerateTime   = 120 // second
	CookieRoundTripTimeMs  = 20  // millisecond
)

// transaction
const (
	AgentLocalTransactionResponseTimeoutMs  = 5 * 1000                                     // millisecond
	ServerLocalTransactionResponseTimeoutMs = AgentLocalTransactionResponseTimeoutMs - 300 // millisecond
	ACLocalTransactionResponseTimeoutMs     = ServerLocalTransactionResponseTimeoutMs      // millisecond

	RemoteTransactionProcessTimeoutMs = 10 * 1000 // millisecond
)

// peer
const (
	MinimalPeerAddressHoldTime = 5 // second
)

// hostname resolve
const (
	MinimalNSLookupTime = 300 // second
)

// packet
const (
	HeaderCommonSize      = 24
	HeaderSize            = 160
	HeaderSizeEx          = 224
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
