package server

const (
	MaxConcurrentConnection         = 20480
	OverloadConnectionThreshold     = MaxConcurrentConnection * 4 / 5 // 80%
	BlockAddrRefreshRate            = 20                              // 20 seconds
	BlockAddrExpireTime             = 90                              // 90 seconds
	PreCheckThreatCountBeforeBlock  = 5                               // block source address if packet precheck errors exceeds this count
	DefaultAgentConnectionTimeoutMs = 30 * 1000                       // 30 seconds to delete idle connection
	DefaultACConnectionTimeoutMs    = 300 * 1000                      // 300 seconds to delete idle connection
	PacketQueueSizePerConnection    = 64
)

// http APIs
const (
	HttpTransactionTimeout = 3 // second
)

// knock
const (
	DefaultIpOpenTime         = 120 // second, align with ipset default timeout
	ACOpenCompensationTime    = 5   // second
	TokenStoreRefreshInterval = 10  // second
)
