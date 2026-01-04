package common

// Connection timeout constants
const (
	// ServerSideConnectionTimeoutMs is the default timeout for server-side connections
	// (AC, Server-to-AC, Server-to-DB). These connections are long-lived and need
	// higher timeouts.
	ServerSideConnectionTimeoutMs = 300 * 1000 // 300 seconds

	// ClientSideConnectionTimeoutMs is the default timeout for client-side connections
	// (Agent, DB). These connections are typically short-lived.
	ClientSideConnectionTimeoutMs = 30 * 1000 // 30 seconds
)

// Token store constants
const (
	// TokenStoreRefreshInterval is the interval in seconds between token store
	// cleanup operations. Used by both AC and Server.
	TokenStoreRefreshInterval = 10 // seconds
)

// Server discovery constants (shared by AC and DB)
const (
	// ReportToServerInterval is how often to report status to the server.
	ReportToServerInterval = 60 // seconds

	// MinimalServerDiscoveryInterval is the minimum time between server discovery attempts.
	MinimalServerDiscoveryInterval = 5 // seconds

	// ServerKeepaliveInterval is how often to send keepalive messages.
	ServerKeepaliveInterval = 20 // seconds

	// ServerDiscoveryRetryBeforeFail is the number of retries before declaring failure.
	ServerDiscoveryRetryBeforeFail = 3
)
