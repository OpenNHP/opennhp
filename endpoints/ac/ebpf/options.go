package ebpf

import (
	utilsebpf "github.com/OpenNHP/opennhp/nhp/utils/ebpf"
)

// LoadOptions are the AC's load-time choices that the AC's own configuration
// decides and the shared loader has to be told about.
//
// They live in their own file, without a build tag, because both the Linux
// loader and the stub that stands in for it on every other platform have to
// agree on the shape: the AC's configuration is parsed on every platform, and
// a type that only existed on Linux would make the configuration's meaning
// depend on the host the daemon was built for.
//
// They are all load-time because the filter and the log writers are set up
// once, when the daemon starts: the IPv6 knob is a constant the verifier
// folds into the program, and the audit log's bounds are read by the writer
// and by the sweep. Changing one of them takes a restart.
type LoadOptions struct {
	// DisableIPv6Filter turns the XDP filter's IPv6 tree off: every IPv6
	// packet is passed to the host, the way it was before the tree existed.
	// False (the default) keeps the tree on, which is the safe answer: the
	// tree is what keeps a v6 listener the AC was never meant to gate out
	// of reach.
	DisableIPv6Filter bool

	// AuditLogRetentionDays, AuditLogRetentionBytes and
	// AuditLogDailyEventBytes bound the AC's legacy audit logs --
	// nhp_deny-*.log and nhp_accept-*.log.
	//
	// Zero means unbounded for all three: the files are kept for as long as
	// the host lives and the day's lines are never suppressed, which is what
	// those logs were before the bounds existed and what a deployment that
	// reads them as an audit trail needs.
	AuditLogRetentionDays   int
	AuditLogRetentionBytes  int64
	AuditLogDailyEventBytes int64

	// DhcpSource and NtpSource narrow the two client-protocol pairs the
	// filter admits -- the DHCP answer (67 -> 68) and the NTP answer
	// (123 -> 123) -- to the addresses the host's own servers answer from.
	//
	// A range that is not set means no narrowing: the pair is admitted by
	// port alone, the way it was before the ranges existed.
	DhcpSource utilsebpf.SourceNet
	NtpSource  utilsebpf.SourceNet
}
