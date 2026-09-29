//go:build !linux

package ebpf

// Non-Linux stubs for the shared eBPF loader. ErrEBPFSupportedOnlyOnLinux is
// declared in ebpf_other.go, which this file sits next to under the same build
// tag.

import (
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
	Variant       EngineVariant
	IfaceName     string
	ProgObjPath   string
	TcProgObjPath string
	PinDir        string
	ComponentId   string
	LogDirPath    string
	LogLevel      int
}

type EngineHandle struct {
	Variant     EngineVariant
	Objs        any
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
