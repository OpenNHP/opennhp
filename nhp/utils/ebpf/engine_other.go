//go:build !linux

package ebpf

// Non-Linux stubs for the shared eBPF loader. ErrEBPFSupportedOnlyOnLinux is
// declared in ebpf_other.go, which this file sits next to under the same build
// tag.

import (
	"time"

	"github.com/cilium/ebpf"

	"github.com/OpenNHP/opennhp/nhp/log"
)

type EngineVariant int

const (
	VariantAC EngineVariant = iota
	VariantServer
)

const DefaultPinDir = "/sys/fs/bpf/"

type EngineLoadParams struct {
	Variant          EngineVariant
	IfaceName        string
	ProgObjPath      string
	TcProgObjPath    string
	PinDir           string
	ComponentId      string
	LogDirPath       string
	LogLevel         int
	NhpPort          uint16
	NhpMinFrameBytes uint16
	RelayIPs         []string

	ScanReportInterval time.Duration
	ScanReportTopN     int

	DisableEventLog bool

	// DisableIPv6Filter, AcAuditLogRetentionDays, AcAuditLogRetentionBytes
	// and AcAuditLogDailyEventBytes are inert off Linux — nothing is loaded,
	// so nothing reads them. They are here because the shape of the
	// parameters has to be the same on every platform: the tool that asks
	// the kernel whether it accepts an object (see VerifyLoad) is built on
	// every platform, and a field that only existed on Linux would make that
	// tool's meaning depend on the host it was built for.
	DisableIPv6Filter bool

	AcAuditLogRetentionDays   int
	AcAuditLogRetentionBytes  int64
	AcAuditLogDailyEventBytes int64

	// DhcpSource and NtpSource narrow the two client-protocol pairs the
	// filters admit. Off Linux nothing is loaded, so nothing reads them; they
	// are here because the server's loader is portable and sets them.
	DhcpSource SourceNet
	NtpSource  SourceNet
}

type EngineHandle struct {
	Variant     EngineVariant
	Objs        any
	IfaceName   string
	EventsMap   *ebpf.Map
	RelayIPsMap *ebpf.Map

	DenyLogger   *log.Logger
	AcLogger     *log.Logger
	ServerLogger *log.Logger
}

func EngineLoad(params EngineLoadParams) (*EngineHandle, error) {
	log.Info("eBPF function must be compiled on Linux OS")
	return nil, ErrEBPFSupportedOnlyOnLinux
}

// ReplaceRelayIPs is a no-op off Linux: there is no map to hold the whitelist,
// and the caller (the server's xdp.toml reload path) treats a nil error as
// "nothing to do" rather than surfacing a platform complaint on every reload.
func ReplaceRelayIPs(relayMap *ebpf.Map, ipStrs []string) error {
	return nil
}

func CleanupBPFFiles(variant EngineVariant) {
	log.Info("ebpf func must be compile based linux os")
}

// DropLoaderPrivileges is a no-op off Linux: there is no ambient capability set
// to clear, and nothing here ever held CAP_BPF to give back. Declared so the
// caller can drop unconditionally without a build tag of its own.
func DropLoaderPrivileges(allowKeepBpf bool) {}
