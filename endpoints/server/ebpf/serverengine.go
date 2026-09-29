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
	"net"
	"path/filepath"
	"sync"
	"time"

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
	go watchHostAddressing(h)
	return nil
}

// How long the host may go without a routable IPv4 address on the filtered
// interface before the filter takes itself off. Four 30s strikes rather than
// one: an address can be replaced in place (a lease rebind, an operator
// reconfiguring the interface) and a single unlucky sample must not detach a
// working filter.
const (
	addressWatchdogInterval = 30 * time.Second
	addressWatchdogStrikes  = 4
)

// watchHostAddressing is the dead man's switch for the ingress filter.
//
// The filter is written entirely in terms of the host's IPv4 configuration,
// and there is no break-glass path into the demo host: if something the host
// needs is dropped and the effect is that it loses that configuration, nobody
// can get in to fix it. That is not hypothetical — it is what the missing DHCP
// exception did (see the DHCP branch in nhp/ebpf/xdp/nhp_server_xdp.c): the
// lease stopped renewing, systemd-networkd dropped the address at expiry, and
// the host went dark on every port about an hour after a deploy that had
// verified fine.
//
// So: if the filtered interface has no routable IPv4 address for two minutes,
// detach. This cannot weaken a state anyone can reach — without an address the
// host answers nothing anyway — and it turns that class of mistake from
// "detach the root volume or rebuild the instance" into "the address comes
// back on the next DHCP exchange, with a Critical line saying why". The filter
// stays off until the daemon is restarted; coming back automatically would
// just re-arm the same trap.
func watchHostAddressing(h *utilsebpf.EngineHandle) {
	if h.IfaceName == "" {
		// Nothing to watch: only the loader knows which interface it attached
		// to, and it always fills this in.
		return
	}

	strikes := 0
	for {
		time.Sleep(addressWatchdogInterval)

		mu.Lock()
		current := handle == h
		mu.Unlock()
		if !current {
			// Unloaded (shutdown) or superseded: not this goroutine's filter
			// to police any more.
			return
		}

		if interfaceHasRoutableIPv4(h.IfaceName) {
			strikes = 0
			continue
		}

		strikes++
		log.Warning("xdp watchdog: interface %s has no routable IPv4 address (%d/%d)",
			h.IfaceName, strikes, addressWatchdogStrikes)
		if strikes < addressWatchdogStrikes {
			continue
		}

		log.Critical("xdp watchdog: %s has had no routable IPv4 address for %v — detaching the XDP ingress filter so the host can recover its addressing (DHCP). The filter stays off until nhp-serverd is restarted; check REASON=UDP_OTHER with SPT=67 in the event log",
			h.IfaceName, time.Duration(addressWatchdogStrikes)*addressWatchdogInterval)
		CleanupBPFFiles()
		return
	}
}

func interfaceHasRoutableIPv4(name string) bool {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		log.Warning("xdp watchdog: cannot look up interface %s: %v", name, err)
		return false
	}
	addrs, err := iface.Addrs()
	if err != nil {
		log.Warning("xdp watchdog: cannot list addresses of %s: %v", name, err)
		return false
	}
	return hasRoutableIPv4(addrs)
}

// hasRoutableIPv4 reports whether the interface holds an IPv4 address the host
// can actually be reached on. A 169.254.0.0/16 autoconfiguration address does
// not count: that is precisely what a host that has lost its DHCP lease ends
// up with, so treating it as an address would defeat the watchdog.
func hasRoutableIPv4(addrs []net.Addr) bool {
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		ip4 := ip.To4()
		if ip4 == nil {
			continue
		}
		if ip4.IsUnspecified() || ip4.IsLinkLocalUnicast() || ip4.IsLoopback() {
			continue
		}
		return true
	}
	return false
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
