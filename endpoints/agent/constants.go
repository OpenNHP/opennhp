package agent

const (
	MaxConcurrentConnection      = 256
	DefaultConnectionTimeoutMs   = 30 * 1000 // 30 seconds to delete idle connection
	PacketQueueSizePerConnection = 8         // nhp agent does not need large transactions
)
