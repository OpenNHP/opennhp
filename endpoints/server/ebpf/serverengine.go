// Package ebpf is the nhp-server's side of the shared eBPF loader in
// nhp/utils/ebpf.
//
// The server attaches one XDP program (nhp/ebpf/xdp/nhp_server_xdp.c) and no TC
// egress program: its policy is about what may reach the host, and it opens no
// outbound flows that need tracking back in. The only map user space drives is
// the relay whitelist, refreshed from etc/xdp.toml on every reload.
//
// Every entry point here is safe to call when no engine is loaded. That is the
// fail-open contract: if EngineLoad fails, the daemon keeps running with no
// ingress filter, and later config reloads and the shutdown path must not
// panic or error on the missing handle.
package ebpf

import (
	"path/filepath"
	"sync"

	utilsebpf "github.com/OpenNHP/opennhp/nhp/utils/ebpf"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// ObjFileName is the compiled XDP object the server loads, expected in
// <exe dir>/etc alongside the TOML configs.
const ObjFileName = "nhp_server_xdp.o"

var (
	mu     sync.Mutex
	handle *utilsebpf.EngineHandle
)

// EngineLoad attaches the server's XDP filter. dirPath is the daemon's exe
// directory: the object is read from dirPath/etc and the perf-event log is
// written to dirPath/logs.
func EngineLoad(dirPath string, logLevel int, serverId string) error {
	mu.Lock()
	defer mu.Unlock()

	h, err := utilsebpf.EngineLoad(utilsebpf.EngineLoadParams{
		Variant:     utilsebpf.VariantServer,
		ProgObjPath: filepath.Join(dirPath, "etc", ObjFileName),
		PinDir:      utilsebpf.DefaultPinDir,
		ComponentId: serverId,
		LogDirPath:  dirPath,
		LogLevel:    logLevel,
	})
	if err != nil {
		return err
	}
	handle = h
	return nil
}

// Loaded reports whether the XDP filter is attached. The config path uses it to
// tell "no engine, nothing to apply" apart from a genuine map error.
func Loaded() bool {
	mu.Lock()
	defer mu.Unlock()
	return handle != nil
}

// UpdateRelayIPs makes the kernel-side whitelist hold exactly ips.
//
// Serialised under the package mutex so two overlapping xdp.toml reloads cannot
// interleave their writes and leave the map holding a mix of both lists. With
// no engine loaded this is a no-op: the filter is not enforcing anything, so
// there is no whitelist to keep in step, and returning an error here would turn
// every reload on a fail-open host into a spurious failure.
func UpdateRelayIPs(ips []string) error {
	mu.Lock()
	defer mu.Unlock()

	if handle == nil {
		return nil
	}
	return utilsebpf.ReplaceRelayIPs(handle.RelayIPsMap, ips)
}

// CleanupBPFFiles detaches the XDP link and removes the server's pins. Safe to
// call when nothing was ever loaded — the pins may still be on disk from a
// previous run that died without cleaning up, and removing those is the point.
func CleanupBPFFiles() {
	mu.Lock()
	defer mu.Unlock()

	utilsebpf.CleanupBPFFiles(utilsebpf.VariantServer)
	if handle != nil && handle.ServerLogger != nil {
		handle.ServerLogger.Close()
	}
	handle = nil
	log.Info("server eBPF engine unloaded")
}
