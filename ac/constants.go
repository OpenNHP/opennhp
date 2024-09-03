package ac

const (
	MaxConcurrentConnection      = 256
	DefaultConnectionTimeoutMs   = 300 * 1000 // 300 seconds to delete idle connection, align with server
	PacketQueueSizePerConnection = 32

	ReportToServerInterval         = 60 // seconds
	MinialServerDiscoveryInterval  = 5  // seconds
	ServerKeepaliveInterval        = 20 // seconds
	ServerDiscoveryRetryBeforeFail = 3

	TempPortOpenTime = 30 //

	IPSET_DEFAULT_NAME      = "defaultset"
	IPSET_DEFAULT_DOWN_NAME = "defaultset_down"
)
