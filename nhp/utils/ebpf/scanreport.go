package ebpf

// The [NHP-SCAN] knobs, shared by the nhp-server's ingress filter and the
// nhp-ac's.
//
// The two daemons report the same log shape out of the same kind of
// telemetry, so the window length and the top-N depth are one decision, not
// two: a host whose summary lists 20 ports and another whose summary lists 5
// would make the published site's per-window numbers incomparable. They live
// here, in the package both daemons load their filter through, so the two
// cannot drift apart.
//
// The nhp-server keeps its own exported names as aliases (see
// endpoints/server/config.go), so nothing outside that file changes.

const (
	// DefaultScanReportIntervalSeconds is the [NHP-SCAN] summary window: one
	// line per active class per window.
	DefaultScanReportIntervalSeconds = 300
	// DefaultScanReportTopN is how many ports and sources each window lists.
	DefaultScanReportTopN = 20

	// MinScanReportIntervalSeconds and MaxScanReportIntervalSeconds bound
	// the accepted window. A value outside the range disables the summary
	// rather than being clamped silently: the caller says so in the log and
	// falls back to the default.
	MinScanReportIntervalSeconds = 10
	MaxScanReportIntervalSeconds = 3600

	// MinScanReportTopN and MaxScanReportTopN bound the accepted depth, for
	// the same reason.
	MinScanReportTopN = 1
	MaxScanReportTopN = 200
)
