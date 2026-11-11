//go:build linux

package ebpf

// The loading, pinning, attaching and perf-reading that used to live here now
// lives in nhp/utils/ebpf, shared with the nhp-server's own XDP filter. What
// remains is the AC's parameters: which object files it loads, which id it
// stamps into the event log, and the two loggers endpoints/ac closes on
// shutdown. The sequence the shared loader runs for VariantAC is the same one
// this file used to run inline.

import (
	"path/filepath"

	utilsebpf "github.com/OpenNHP/opennhp/nhp/utils/ebpf"

	"github.com/OpenNHP/opennhp/nhp/log"
)

var (
	DenyLogger *log.Logger
	AcLogger   *log.Logger
)

// LoadOptions are the AC's load-time choices; see options.go, which has no
// build tag so that the shape is the same on every platform.
func EbpfEngineLoad(dirPath string, logLevel int, acId string, opts LoadOptions) error {
	// Object paths stay relative to the working directory, which is where
	// `make acd` puts them and where the systemd unit starts the daemon —
	// changing them to dirPath-relative would move the lookup on any host
	// whose WorkingDirectory is not the exe directory.
	const bpfDir = "etc"

	handle, err := utilsebpf.EngineLoad(utilsebpf.EngineLoadParams{
		Variant:       utilsebpf.VariantAC,
		ProgObjPath:   filepath.Join(bpfDir, "nhp_ebpf_xdp.o"),
		TcProgObjPath: filepath.Join(bpfDir, "tc_egress.o"),
		PinDir:        utilsebpf.DefaultPinDir,
		ComponentId:   acId,
		LogDirPath:    dirPath,
		LogLevel:      logLevel,
		// ScanReportInterval and ScanReportTopN are deliberately left at
		// zero: the loader fills the AC's defaults in (acScanReportSettings),
		// so the [NHP-SCAN] summary runs whether or not the caller remembers
		// to ask for it. The AC's port and source tables are drained by that
		// reporter, and a load that started none would leave them to fill up
		// and never be read.
		DisableIPv6Filter:         opts.DisableIPv6Filter,
		AcAuditLogRetentionDays:   opts.AuditLogRetentionDays,
		AcAuditLogRetentionBytes:  opts.AuditLogRetentionBytes,
		AcAuditLogDailyEventBytes: opts.AuditLogDailyEventBytes,
		DhcpSource:                opts.DhcpSource,
		NtpSource:                 opts.NtpSource,
	})
	if err != nil {
		return err
	}

	DenyLogger = handle.DenyLogger
	AcLogger = handle.AcLogger
	return nil
}

// clean eBPF map file
func CleanupBPFFiles() {
	utilsebpf.CleanupBPFFiles(utilsebpf.VariantAC)
}
