//go:build linux

package ebpf

// Loader shared by the two eBPF users in the tree.
//
// The AC (nhp/ebpf/xdp/nhp_ebpf_xdp.c + tc_egress.c) opens and closes protected
// ports per knock; the server (nhp/ebpf/xdp/nhp_server_xdp.c) applies one fixed
// ingress policy. They have no maps in common, but everything around the
// maps -- memlock, LoadCollectionSpec, pinning under /sys/fs/bpf, resolving the
// default-route interface, attaching XDP, draining the perf ring into a
// log.Logger, and tearing the pins down again -- was identical, so it lives
// here once and is selected with EngineLoadParams.Variant.
//
// The AC path is a straight lift of what endpoints/ac/ebpf/ebpfegine.go used to
// do inline, in the same order, with the same pin names and the same loggers:
// the variant is chosen by the caller's parameters, never re-decided inside a
// step, so the AC's sequence is unchanged. The one deliberate difference is
// that a missing network interface now returns an error instead of calling
// os.Exit(1) -- the AC's caller already handles a returned error, and the
// server's fail-open startup depends on getting one back.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	stdlog "log"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// EngineVariant selects which object file, map set and attachment points a load
// uses. VariantAC is iota so that a zero-value EngineLoadParams keeps the AC
// behavior it had before this package existed.
type EngineVariant int

const (
	VariantAC EngineVariant = iota
	VariantServer
)

// DefaultPinDir is the bpffs directory both variants pin into. Empty
// EngineLoadParams.PinDir falls back to it.
const DefaultPinDir = "/sys/fs/bpf/"

type EngineLoadParams struct {
	Variant EngineVariant

	// IfaceName is the interface to attach to. Empty resolves the host's
	// default-route interface, which is what both daemons want.
	IfaceName string

	// ProgObjPath is the compiled XDP object. TcProgObjPath is the TC egress
	// object and is required for VariantAC only; VariantServer attaches no TC
	// program (the server has no outbound flows to track).
	ProgObjPath   string
	TcProgObjPath string

	PinDir string

	// ComponentId is stamped into every perf-event log line (the AC id, the
	// server hostname) so a collected log says which host produced it.
	ComponentId string

	// LogDirPath is the daemon's exe directory; the perf-event loggers write
	// into LogDirPath/logs.
	LogDirPath string
	LogLevel   int

	// NhpPort is the UDP port the daemon actually listens on, and
	// NhpMinFrameBytes the shortest datagram it will accept there. VariantServer
	// rewrites both into the object's .rodata before it is verified, so the
	// filter's idea of "the knock port" is the daemon's ListenPort rather than
	// a compile-time 62206 — a server configured anywhere else would otherwise
	// have every knock dropped at the driver. Zero leaves the object's own
	// defaults in place. Ignored by VariantAC, which opens ports per knock.
	NhpPort          uint16
	NhpMinFrameBytes uint16

	// RelayIPs is the SSH whitelist the filter starts life enforcing, in the
	// same spelling as etc/xdp.toml (host addresses or CIDR prefixes). Required
	// by VariantServer, ignored by VariantAC.
	//
	// It is a load parameter rather than something the caller writes afterwards
	// because the map must hold the whitelist *before* the program is attached.
	// Attaching first filters the host with an empty trie, which drops tcp/22
	// from every source: harmless for the microseconds until the write lands,
	// unrecoverable if that write then fails — the caller would be left with an
	// attached filter, no SSH and nothing to correct but a map it could not
	// write a moment ago. Passed in here, a failure happens before the attach
	// and unwinds to fail-open like every other pre-attach error.
	RelayIPs []string
}

// Names of the .rodata constants in nhp/ebpf/xdp/nhp_server_xdp.c that
// EngineLoadParams.NhpPort / NhpMinFrameBytes are written into. Keep in sync
// with the declarations there.
const (
	varNhpListenPort = "nhp_listen_port"
	varNhpMinUdpLen  = "nhp_min_udp_len"
)

// EngineHandle is what the caller keeps: the maps it drives from user space and
// the links it must close on shutdown. Objs is the variant's loaded object
// struct, held so the collection is not garbage collected out from under the
// pins; nothing outside this package reads it.
type EngineHandle struct {
	Variant EngineVariant
	Objs    any

	// IfaceName is the interface the XDP program was actually attached to,
	// resolved from the default route when the caller did not name one. The
	// server's watchdog needs it to tell whether the host still has the IPv4
	// configuration the filter is written in terms of (see
	// endpoints/server/ebpf).
	IfaceName string

	EventsMap   *ebpf.Map
	RelayIPsMap *ebpf.Map
	XdpLink     link.Link
	TcLink      link.Link

	// AC perf-event loggers, kept exported because endpoints/ac closes them
	// on shutdown. Nil for VariantServer, which logs both verdicts to one
	// file (ServerLogger) since its actions are finer-grained than
	// accept/deny.
	DenyLogger   *log.Logger
	AcLogger     *log.Logger
	ServerLogger *log.Logger
}

type bpfObjects struct {
	XdpProg       *ebpf.Program `ebpf:"xdp_white_prog"`
	Whitelist     *ebpf.Map     `ebpf:"spp"`
	Icmpwhitelist *ebpf.Map     `ebpf:"icmpwhitelist"`
	Sdwhitelist   *ebpf.Map     `ebpf:"sdwhitelist"`
	Srcportlist   *ebpf.Map     `ebpf:"src_port"`
	Portlist      *ebpf.Map     `ebpf:"port_list"`
	Protocolport  *ebpf.Map     `ebpf:"protocol_port"`
	Conntrack     *ebpf.Map     `ebpf:"conn_track"`
	Events        *ebpf.Map     `ebpf:"events"`
	// Peers a knock has admitted, and to which port. Written by the XDP
	// program, read by the TC egress program's gate 4 — see `knock_peers` in
	// nhp/ebpf/xdp/nhp_maps.h. Nothing in user space drives it; it is
	// assigned here only so that a pin left over from an incompatible build
	// fails the load by name, the same reason Conntrack is assigned below.
	KnockPeers *ebpf.Map `ebpf:"knock_peers"`
}

type tcBpfObjects struct {
	TcEgressProg *ebpf.Program `ebpf:"tc_egress_prog"`
	// The TC egress program records the AC's own outbound connections in
	// conn_track (shared with the XDP program through its pin) so their
	// replies get back in. It no longer writes the knock whitelist `spp`; it
	// only reads the whitelist maps and `knock_peers`, to tell a peer that
	// holds — or once held — a knock apart from a host the AC itself connected
	// to. Those maps are not listed here because nothing in user space drives
	// them through this object — the XDP collection above is loaded first and
	// creates every pin.
	//
	// Conntrack is not used after the load, but keeping it assigned means a
	// pin left over from an incompatible build fails here, at startup, with a
	// `conn_track` error naming the map — see "eBPF map pin mismatch" in
	// terraform/demo/RUNBOOK.md.
	Conntrack *ebpf.Map `ebpf:"conn_track"`
	// Runtime knobs for the TC program, filled in by setEbpfConfig below.
	Config *ebpf.Map `ebpf:"nhp_config"`
}

// serverBpfObjects mirrors nhp/ebpf/xdp/nhp_server_xdp.c. Every map is
// assigned so that a pin left over from an incompatible build fails the load by
// name rather than at first use, the same reason the AC assigns Conntrack.
type serverBpfObjects struct {
	XdpProg   *ebpf.Program `ebpf:"xdp_server_prog"`
	RelayIPs  *ebpf.Map     `ebpf:"nhp_relay_ips"`
	NhpEvents *ebpf.Map     `ebpf:"nhp_events"`
	// Per-action packet counters, the complete accounting the rate-limited
	// event stream is a sample of. See reportServerStats.
	ActionStats *ebpf.Map `ebpf:"nhp_action_stats"`
}

// Config slots in the `nhp_config` map. Keep in sync with the enum in
// nhp/ebpf/xdp/nhp_maps.h.
const cfgEphemeralPortMin uint32 = 0

// defaultEphemeralPortMin mirrors NHP_EPHEMERAL_PORT_MIN, the fallback the eBPF
// program uses when the slot is unset.
const defaultEphemeralPortMin uint32 = 32768

// Pin names each variant owns, relative to the directory it pinned into.
// CleanupBPFFiles removes only its own variant's list, so the two objects could
// coexist on one host without one tearing down the other's maps.
//
// Names, not absolute paths: the pin directory is EngineLoadParams.PinDir, and
// a cleanup that went to a hard-coded /sys/fs/bpf would both leak the pins of a
// load that used another directory — the next LoadAndAssign then fails against
// a stale map, which is the failure this cleanup exists to prevent — and delete
// the live host's pins from under it.
var acPinnedNames = []string{
	"xdp_white_prog",
	"conn_track",
	"icmpwhitelist",
	"port_list",
	"protocol_port",
	"sdwhitelist",
	"src_port",
	"spp",
	"knock_peers",
	"nhp_config",
	"tc_egress_prog",
}

// `nhp_events` is absent on purpose: it is a perf event array read only by
// this process, so nhp_server_xdp.c does not mark it LIBBPF_PIN_BY_NAME and
// there is no pin to remove. Same as the AC's `events`.
var serverPinnedNames = []string{
	"xdp_server_prog",
	"nhp_relay_ips",
}

var (
	acXdpLink     link.Link
	acTcLink      link.Link
	serverXdpLink link.Link
	bootTime      time.Time

	// Where each variant last pinned, so the exported CleanupBPFFiles(variant)
	// — which takes no directory, and is called from daemon shutdown paths and
	// from the server's watchdog — removes the pins this process actually
	// created. DefaultPinDir until a load says otherwise, which is also the
	// right guess for the pre-load sweep of a previous run's leftovers.
	// Unsynchronized, like the links above: callers serialize their own loads
	// (see endpoints/server/ebpf, which holds a mutex over both).
	acPinDir     = DefaultPinDir
	serverPinDir = DefaultPinDir
)

// pinnedFiles is the variant's pin list under dir.
func pinnedFiles(variant EngineVariant, dir string) []string {
	names := acPinnedNames
	if variant == VariantServer {
		names = serverPinnedNames
	}
	paths := make([]string, 0, len(names))
	for _, name := range names {
		paths = append(paths, filepath.Join(dir, name))
	}
	return paths
}

func activePinDir(variant EngineVariant) string {
	if variant == VariantServer {
		return serverPinDir
	}
	return acPinDir
}

func recordPinDir(variant EngineVariant, dir string) {
	if variant == VariantServer {
		serverPinDir = dir
		return
	}
	acPinDir = dir
}

func init() {
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err != nil {
		panic("Failed to get the system running time: " + err.Error())
	}

	now := time.Now()
	bootTime = now.Add(-time.Duration(info.Uptime) * time.Second)
	log.Info("​​System boot time: %v", bootTime)
}

// EngineLoad compiles-in nothing itself: it loads the already-compiled object at
// params.ProgObjPath, pins it, attaches it and starts the perf reader. A
// returned error leaves nothing attached, which is what lets the server treat a
// failure as fail-open.
func EngineLoad(params EngineLoadParams) (*EngineHandle, error) {
	pinDir := params.PinDir
	if pinDir == "" {
		pinDir = DefaultPinDir
	}
	// Recorded before the sweep below, so that both it and every later
	// CleanupBPFFiles(variant) — including the deferred unwind inside the two
	// loaders — act on the directory this load pins into.
	recordPinDir(params.Variant, pinDir)
	CleanupBPFFiles(params.Variant)

	if err := rlimit.RemoveMemlock(); err != nil {
		log.Error("Failed to remove memlock limit")
	}

	switch params.Variant {
	case VariantServer:
		return loadServerEngine(params, pinDir)
	default:
		return loadAcEngine(params, pinDir)
	}
}

func loadAcEngine(params EngineLoadParams, pinDir string) (h *EngineHandle, retErr error) {
	specPath := params.ProgObjPath
	tcSpecPath := params.TcProgObjPath

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		log.Error("eBPF object file not found ")
		return nil, err
	}
	if _, err := os.Stat(tcSpecPath); os.IsNotExist(err) {
		log.Error("tc eBPF object file not found ")
		return nil, err
	}

	spec, err := ebpf.LoadCollectionSpec(specPath)
	if err != nil {
		log.Error("failed to load eBPF object")
		return nil, err
	}
	// Load tc eBPF object
	tcSpec, err := ebpf.LoadCollectionSpec(tcSpecPath)
	if err != nil {
		log.Error("failed to load tc eBPF object")
		return nil, err
	}

	var objs bpfObjects
	if loadErr := spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinDir, // automatically mounted to
		},
	}); loadErr != nil {
		log.Error("Failed to load and assign eBPF objects")
		return nil, loadErr
	}

	var tcObjs tcBpfObjects
	if tcLoadErr := tcSpec.LoadAndAssign(&tcObjs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinDir, // automatically mounted to
		},
	}); tcLoadErr != nil {
		log.Error("Failed to load and assign tc eBPF objects")
		return nil, tcLoadErr
	}
	// Past the loads, a failure leaves state behind: LoadAndAssign has created
	// the map pins under bpffs, and once the attaches below start, a
	// half-attached pair of programs (XDP filtering ingress with no TC egress to
	// record the AC's own flows, which drops the replies to them). The caller
	// only ever sees an error, so nothing else would take that down.
	defer func() {
		if retErr != nil {
			CleanupBPFFiles(VariantAC)
		}
	}()

	if cfgErr := setEbpfConfig(&tcObjs); cfgErr != nil {
		log.Error("Failed to configure tc eBPF program")
		return nil, cfgErr
	}

	if pinErr := objs.XdpProg.Pin(filepath.Join(pinDir, "xdp_white_prog")); pinErr != nil {
		log.Error("failed to pin XDP program xdp_white_prog to %s", pinDir)
		return nil, pinErr
	}
	if tcPinErr := tcObjs.TcEgressProg.Pin(filepath.Join(pinDir, "tc_egress_prog")); tcPinErr != nil {
		log.Error("failed to pin TC egress program tc_egress_prog to %s", pinDir)
		return nil, tcPinErr
	}

	iface, err := resolveInterface(params.IfaceName)
	if err != nil {
		return nil, err
	}

	//load ebpf nhp_ebpf_xdp.o to net interface which default route exit
	acXdpLink, err = link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpProg,
		Interface: iface.Index,
		Flags:     link.XDPGenericMode, // XDPGenericMode and XDPDriverMode
	})
	if err != nil {
		log.Error("failed to attach XDP program to interface: %s", iface.Name)
		return nil, err
	}
	//load tc eBPF tc_egress.o to net interface which default route exit
	acTcLink, err = link.AttachTCX(link.TCXOptions{
		Program:   tcObjs.TcEgressProg,
		Interface: iface.Index,
		Attach:    ebpf.AttachTCXEgress,
	})
	if err != nil {
		log.Error("failed to attach TC egress program to interface: %s", iface.Name)
		return nil, err
	}

	// Accessing the Perf Buffer Map named "events" defined in eBPF.
	eventsMap := objs.Events
	if eventsMap == nil {
		log.Error("failed to load 'events' map from eBPF object (nil)")
		return nil, fmt.Errorf("'events' map not found")
	}

	h = &EngineHandle{
		Variant:   VariantAC,
		Objs:      &objs,
		IfaceName: iface.Name,
		EventsMap: eventsMap,
		XdpLink:   acXdpLink,
		TcLink:    acTcLink,
	}

	//Set up the DENY logger
	h.DenyLogger = log.NewLoggerDefine(
		"",
		params.LogLevel,
		filepath.Join(params.LogDirPath, "logs"),
		"nhp_deny",
	)
	h.DenyLogger.SetFlags(stdlog.Lmsgprefix)
	// Set up the ACCEPT logger
	h.AcLogger = log.NewLoggerDefine(
		"",
		params.LogLevel,
		filepath.Join(params.LogDirPath, "logs"),
		"nhp_accept",
	)
	h.AcLogger.SetFlags(stdlog.Lmsgprefix)

	// Start a goroutine to monitor Perf Buffer events
	go readAcEvents(eventsMap, params.ComponentId, h.DenyLogger, h.AcLogger)

	return h, nil
}

func readAcEvents(eventsMap *ebpf.Map, acId string, denyLogger, acLogger *log.Logger) {
	perfReader, err := perf.NewReader(eventsMap, os.Getpagesize())
	if err != nil {
		log.Error("failed to create perf reader: %v", err)
		return
	}
	defer perfReader.Close()

	log.Info("Start listening for eBPF events (PERF BUFFER)")

	for {
		record, err := perfReader.Read()
		if err != nil {
			log.Error("Error reading eBPF event: %v", err)
			continue
		}
		if len(record.RawSample) < 24 {
			continue
		}
		action := record.RawSample[8]
		var actionStr string
		switch action {
		case 0:
			actionStr = "DENY"
		case 1:
			actionStr = "ACCEPT"
		default:
			actionStr = "UNKNOWN"
		}
		timestamp := binary.LittleEndian.Uint64(record.RawSample[0:8])
		srcIP := binary.BigEndian.Uint32(record.RawSample[9:13])
		dstIP := binary.BigEndian.Uint32(record.RawSample[13:17])
		srcPort := binary.BigEndian.Uint16(record.RawSample[17:19])
		dstPort := binary.BigEndian.Uint16(record.RawSample[19:21])
		protocol := record.RawSample[21]
		payloadLen := binary.BigEndian.Uint16(record.RawSample[22:24])

		srcIPStr := uint32ToIPv4(srcIP)
		dstIPStr := uint32ToIPv4(dstIP)
		eventTime := bootTime.Add(time.Duration(timestamp))
		protoName := protoToString(protocol)

		logMsg := fmt.Sprintf("%s %s [NHP-%s] SRC=%s DST=%s LEN=%d PROTO=%s SPT=%d DPT=%d",
			eventTime.Format("15:04:05"),
			acId,
			actionStr,
			srcIPStr,
			dstIPStr,
			payloadLen,
			protoName,
			srcPort,
			dstPort,
		)

		if action == 0 { // DENY
			denyLogger.Info("%s", logMsg)
		} else { // ACCEPT
			acLogger.Info("%s", logMsg)
		}
	}
}

// setServerConstants writes the daemon's listen port and datagram floor into
// the server object's .rodata, so the filter enforces the configuration the
// daemon is actually running with.
//
// Refusing rather than warning when the constants are absent is deliberate.
// Attaching an object whose knock port is a compile-time 62206 to a server
// listening somewhere else drops every knock at the driver, which looks exactly
// like the server being down and is not visible anywhere in user space — far
// worse than not attaching at all, which is a state the daemon already handles
// (fail-open, logged). An object that predates the constants is a stale
// nhp_server_xdp.o next to a new binary; rebuilding it is the fix.
func setServerConstants(spec *ebpf.CollectionSpec, params EngineLoadParams) error {
	consts := []struct {
		name  string
		value uint16
	}{
		{varNhpListenPort, params.NhpPort},
		{varNhpMinUdpLen, params.NhpMinFrameBytes},
	}

	for _, c := range consts {
		if c.value == 0 {
			// Caller did not care; leave the object's own default.
			continue
		}
		v, ok := spec.Variables[c.name]
		if !ok {
			log.Error("server eBPF object has no '%s' constant — it is older than the loader; rebuild it with `make ebpf-objs`", c.name)
			return fmt.Errorf("server eBPF object has no %q constant", c.name)
		}
		if err := v.Set(c.value); err != nil {
			log.Error("failed to set server eBPF constant '%s' to %d: %v", c.name, c.value, err)
			return err
		}
	}

	if params.NhpPort != 0 || params.NhpMinFrameBytes != 0 {
		log.Info("server XDP filter configured for udp/%d with a %d-byte datagram floor (0 = the object's own default)",
			params.NhpPort, params.NhpMinFrameBytes)
	}
	return nil
}

// loadServerEngine is ordered so that everything that can fail happens *before*
// the program is attached, and so that any failure after it still leaves the
// host unfiltered.
//
// That ordering is the whole contract the caller relies on: nhp-serverd treats
// a returned error as "fail-open, no ingress filter" and logs it as such, and
// it never gets an EngineHandle back — so a filter that stayed attached behind a
// returned error would be enforcing with an empty SSH whitelist that user space
// no longer has a handle to fix. On a host whose only way in is that whitelist
// that is unrecoverable, so the deferred cleanup below detaches and unpins on
// every error path past the load.
//
// Installing params.RelayIPs is part of "everything that can fail": the filter
// is never attached for one moment holding a whitelist other than the one the
// caller asked for, so there is no window in which SSH from the relay is
// dropped and no way for a failed map write to leave a live filter behind.
func loadServerEngine(params EngineLoadParams, pinDir string) (h *EngineHandle, retErr error) {
	specPath := params.ProgObjPath
	if _, err := os.Stat(specPath); err != nil {
		log.Error("server eBPF object file not found: %s", specPath)
		return nil, err
	}

	spec, err := ebpf.LoadCollectionSpec(specPath)
	if err != nil {
		log.Error("failed to load server eBPF object: %v", err)
		return nil, err
	}

	// Rewrite the .rodata constants before LoadAndAssign: after the collection
	// is loaded the verifier has already folded them, so this is the only
	// moment the knock port and the length floor can still be set.
	if constErr := setServerConstants(spec, params); constErr != nil {
		return nil, constErr
	}

	var objs serverBpfObjects
	if loadErr := spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinDir,
		},
	}); loadErr != nil {
		log.Error("Failed to load and assign server eBPF objects: %v", loadErr)
		return nil, loadErr
	}
	// From here on a failure has left state on the host: LoadAndAssign has
	// already created the map pins under bpffs, and past the attach below there
	// is a live filter. Unwind both, or the daemon runs "fail-open" while the
	// driver is still dropping everything — and with no handle, nothing in user
	// space could fix the whitelist it is enforcing.
	defer func() {
		if retErr != nil {
			CleanupBPFFiles(VariantServer)
		}
	}()

	// Checked here rather than after the attach: a missing map means user space
	// cannot drive the whitelist, and attaching first would filter the host with
	// an empty one and no way to correct it.
	if objs.NhpEvents == nil {
		log.Error("failed to load 'nhp_events' map from server eBPF object (nil)")
		return nil, fmt.Errorf("'nhp_events' map not found")
	}
	if objs.RelayIPs == nil {
		log.Error("failed to load 'nhp_relay_ips' map from server eBPF object (nil)")
		return nil, fmt.Errorf("'nhp_relay_ips' map not found")
	}

	if pinErr := objs.XdpProg.Pin(filepath.Join(pinDir, "xdp_server_prog")); pinErr != nil {
		log.Error("failed to pin XDP program xdp_server_prog to %s: %v", pinDir, pinErr)
		return nil, pinErr
	}

	// The whitelist goes in before the program goes on. Everything the filter
	// admits that is not a knock arrives on tcp/22 from one of these prefixes,
	// so a program attached with an empty trie is a host with no way in — and
	// if this write is what failed, user space has just proved it cannot fix
	// the map it would have to fix. Here it is only an error like any other:
	// the deferred cleanup unpins, nothing is attached, and the daemon runs
	// with the exposure it had before the filter existed.
	//
	// ReplaceRelayIPs re-validates (all-or-nothing parsing, and a refusal to
	// leave the map empty) rather than trusting that the caller already did.
	if relayErr := ReplaceRelayIPs(objs.RelayIPs, params.RelayIPs); relayErr != nil {
		log.Error("refusing to attach the server XDP program: the relay whitelist could not be installed: %v", relayErr)
		return nil, relayErr
	}

	iface, err := resolveInterface(params.IfaceName)
	if err != nil {
		return nil, err
	}
	warnOnGlobalIPv6(iface)

	serverXdpLink, err = link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpProg,
		Interface: iface.Index,
		Flags:     link.XDPGenericMode,
	})
	if err != nil {
		log.Error("failed to attach server XDP program to interface %s: %v", iface.Name, err)
		return nil, err
	}

	h = &EngineHandle{
		Variant:     VariantServer,
		Objs:        &objs,
		IfaceName:   iface.Name,
		EventsMap:   objs.NhpEvents,
		RelayIPsMap: objs.RelayIPs,
		XdpLink:     serverXdpLink,
	}

	// One log file, not the AC's accept/deny pair: the server's actions say
	// *why* a packet was passed or dropped (relay SSH, short datagram, TCP on
	// the knock port), and splitting them by verdict would throw that away.
	logDir := filepath.Join(params.LogDirPath, "logs")
	h.ServerLogger = log.NewLoggerDefine(
		"",
		params.LogLevel,
		logDir,
		serverEventLogName,
	)
	h.ServerLogger.SetFlags(stdlog.Lmsgprefix)

	go readServerEvents(objs.NhpEvents, params.ComponentId, h.ServerLogger)
	if objs.ActionStats != nil {
		go reportServerStats(objs.ActionStats, params.ComponentId, h.ServerLogger)
	}
	go pruneServerEventLogs(logDir)

	return h, nil
}

// warnOnGlobalIPv6 says so in the log when the interface about to be filtered
// has a global IPv6 address.
//
// The filter's policy is expressed in IPv4: the relay whitelist is a trie of
// IPv4 prefixes, so handle_ipv6() in nhp_server_xdp.c has no way to recognize
// the relay and admits no inbound v6 connection at all — including one to
// sshd's [::]:22. On a host reached over IPv4 (every demo host) that is the
// point of the branch; on a host whose only SSH path is v6 it would be a
// lockout, so it gets a line naming the address rather than a silent change of
// exposure. Not a refusal: refusing would leave the host with no filter at all,
// which is the hole the v6 branch was written to close.
func warnOnGlobalIPv6(iface *net.Interface) {
	addrs, err := iface.Addrs()
	if err != nil {
		log.Warning("cannot list the addresses of %s to check for IPv6: %v", iface.Name, err)
		return
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.To4() != nil || !ip.IsGlobalUnicast() {
			continue
		}
		log.Warning("%s has the global IPv6 address %s: the XDP ingress filter admits no inbound IPv6 connection, tcp/22 included, because the relay whitelist holds IPv4 prefixes only. Make sure SSH to this host arrives over IPv4.",
			iface.Name, ip)
		return
	}
}

// serverEventLogName is the log.Logger name, i.e. the file is
// <logdir>/nhp_server_xdp-<date>.log. pruneServerEventLogs deletes by the same
// prefix, so the two must not drift apart.
const serverEventLogName = "nhp_server_xdp"

// Action codes reported by nhp/ebpf/xdp/nhp_server_xdp.c. Keep in sync.
const (
	ActDropOther       uint8 = 0
	ActSshRelay        uint8 = 1
	ActNhpRelay        uint8 = 2
	ActNhpDefault      uint8 = 3
	ActTcpEstablished  uint8 = 4
	ActUdpEstablished  uint8 = 5
	ActIcmpFragNeeded  uint8 = 6
	ActDhcpClient      uint8 = 7
	ActNtpClient       uint8 = 8
	ActDropTcpSshOther uint8 = 10
	ActDropTcpNhp      uint8 = 11
	ActDropTcpOther    uint8 = 12
	ActDropUdpOther    uint8 = 13
	ActDropUdpShort    uint8 = 14
	ActDropNonUdp      uint8 = 15
	ActIpv4Fragment    uint8 = 16
	ActV6IcmpControl   uint8 = 17
	ActV6Established   uint8 = 18
	ActV6Fragment      uint8 = 19
	ActDropV6Other     uint8 = 20
	// serverEventByteSize is sizeof(struct nhp_event_t), which is packed: 25
	// bytes of IPv4-shaped event plus the address family and the two 16-byte
	// IPv6 addresses the v6 branch reports in.
	serverEventByteSize = 58
	// serverEventActions mirrors NHP_EVENT_ACTIONS, the size of the
	// per-action counter and rate-limit arrays.
	serverEventActions = 24
)

// Address families in the event's `family` field. Not syscall.AF_*: the C side
// writes 4 or 6 to say which of the two address pairs it filled in.
const (
	eventFamilyV4 uint8 = 4
	eventFamilyV6 uint8 = 6
)

func serverActionName(action uint8) (verdict, reason string) {
	switch action {
	case ActSshRelay:
		return "PASS", "SSH_RELAY"
	case ActNhpRelay:
		return "PASS", "NHP_RELAY"
	case ActNhpDefault:
		return "PASS", "NHP_DEFAULT"
	case ActTcpEstablished:
		return "PASS", "TCP_ESTABLISHED"
	case ActUdpEstablished:
		return "PASS", "UDP_ESTABLISHED"
	case ActIcmpFragNeeded:
		return "PASS", "ICMP_FRAG_NEEDED"
	case ActDhcpClient:
		return "PASS", "DHCP_CLIENT"
	case ActNtpClient:
		return "PASS", "NTP_CLIENT"
	case ActIpv4Fragment:
		return "PASS", "IPV4_FRAGMENT"
	case ActV6IcmpControl:
		return "PASS", "ICMPV6_CONTROL"
	case ActV6Established:
		return "PASS", "V6_ESTABLISHED"
	case ActV6Fragment:
		return "PASS", "IPV6_FRAGMENT"
	case ActDropV6Other:
		return "DROP", "V6_OTHER"
	case ActDropTcpSshOther:
		return "DROP", "TCP_SSH_OTHER"
	case ActDropTcpNhp:
		return "DROP", "TCP_NHP_PORT"
	case ActDropTcpOther:
		return "DROP", "TCP_OTHER"
	case ActDropUdpOther:
		return "DROP", "UDP_OTHER"
	case ActDropUdpShort:
		return "DROP", "UDP_SHORT"
	case ActDropNonUdp:
		return "DROP", "NON_TCP_UDP"
	case ActDropOther:
		return "DROP", "OTHER"
	default:
		return "DROP", fmt.Sprintf("UNKNOWN-%d", action)
	}
}

func readServerEvents(eventsMap *ebpf.Map, serverId string, logger *log.Logger) {
	perfReader, err := perf.NewReader(eventsMap, os.Getpagesize())
	if err != nil {
		log.Error("failed to create server perf reader: %v", err)
		return
	}
	defer perfReader.Close()

	log.Info("Start listening for server eBPF events (PERF BUFFER)")

	for {
		record, err := perfReader.Read()
		if err != nil {
			log.Error("Error reading server eBPF event: %v", err)
			continue
		}
		// struct nhp_event_t in nhp_server_xdp.c, __attribute__((packed)).
		if len(record.RawSample) < serverEventByteSize {
			continue
		}
		timestamp := binary.LittleEndian.Uint64(record.RawSample[0:8])
		action := record.RawSample[8]
		srcPort := binary.BigEndian.Uint16(record.RawSample[17:19])
		dstPort := binary.BigEndian.Uint16(record.RawSample[19:21])
		protocol := record.RawSample[21]
		pktLen := binary.BigEndian.Uint16(record.RawSample[22:24])
		relayHit := record.RawSample[24]
		family := record.RawSample[25]

		// One of the two address pairs is filled in and the other is zero; the
		// family says which. Formatting the wrong one would print 0.0.0.0 for
		// every IPv6 event, which is the address a drop line exists to carry.
		srcStr, dstStr := "", ""
		if family == eventFamilyV6 {
			srcStr = net.IP(record.RawSample[26:42]).String()
			dstStr = net.IP(record.RawSample[42:58]).String()
		} else {
			srcStr = uint32ToIPv4(binary.BigEndian.Uint32(record.RawSample[9:13]))
			dstStr = uint32ToIPv4(binary.BigEndian.Uint32(record.RawSample[13:17]))
		}

		verdict, reason := serverActionName(action)
		eventTime := bootTime.Add(time.Duration(timestamp))

		logger.Info("%s %s [NHP-%s] REASON=%s SRC=%s DST=%s LEN=%d PROTO=%s SPT=%d DPT=%d RELAY=%d",
			eventTime.Format("15:04:05"),
			serverId,
			verdict,
			reason,
			srcStr,
			dstStr,
			pktLen,
			protoToString(protocol),
			srcPort,
			dstPort,
			relayHit,
		)
	}
}

// serverStatsInterval is how often the per-action counters are summarized into
// the event log. A minute is short enough to localize an incident to the right
// window and long enough that a permanently scanned host writes one line per
// class per minute -- a few kB a day -- instead of one per packet.
const serverStatsInterval = time.Minute

// serverActionStat mirrors `struct nhp_action_stat` in nhp_server_xdp.c. The
// map is per-CPU, so a lookup yields one of these per CPU and they are summed.
type serverActionStat struct {
	Packets uint64
	Logged  uint64
}

// reportServerStats turns the counters the XDP program keeps for every packet
// into one summary line per action per window.
//
// It exists because the event stream is deliberately incomplete. Two things
// hold its volume down -- the per-action token bucket, and the bulk classes
// that are counted but never reported at all (see record_packet in the C) --
// and both of them lose information: without this, a host under a 100k pps
// scan and a host with one visitor an hour write the same number of DROP lines
// per second, and the reply traffic that keeps the daemon working
// (TCP_ESTABLISHED, UDP_ESTABLISHED) leaves no trace whatsoever. The counters
// are exact regardless of either, so the summary is what makes the log
// quantitative again, at a cost that does not depend on offered load.
//
// PKTS is the delta over the window, LOGGED how many of those reached the
// event stream (0 for the bulk classes, less than PKTS whenever the bucket
// ran dry), TOTAL the count since the filter was attached. Actions that saw
// nothing print nothing: a quiet host writes a quiet log.
func reportServerStats(statsMap *ebpf.Map, serverId string, logger *log.Logger) {
	prev := make([]serverActionStat, serverEventActions)

	ticker := time.NewTicker(serverStatsInterval)
	defer ticker.Stop()

	for range ticker.C {
		for action := 0; action < serverEventActions; action++ {
			var perCPU []serverActionStat
			key := uint32(action)
			if err := statsMap.Lookup(&key, &perCPU); err != nil {
				// A closed map (shutdown) or a transient failure: the
				// next tick re-reads absolute counters, so nothing is
				// lost by skipping this one.
				continue
			}

			var total serverActionStat
			for _, c := range perCPU {
				total.Packets += c.Packets
				total.Logged += c.Logged
			}

			delta := total.Packets - prev[action].Packets
			if delta == 0 {
				continue
			}
			logged := total.Logged - prev[action].Logged
			prev[action] = total

			verdict, reason := serverActionName(uint8(action))
			logger.Info("%s %s [NHP-STAT] VERDICT=%s REASON=%s WINDOW=%ds PKTS=%d LOGGED=%d TOTAL=%d",
				time.Now().Format("15:04:05"),
				serverId,
				verdict,
				reason,
				int(serverStatsInterval.Seconds()),
				delta,
				logged,
				total.Packets,
			)
		}
	}
}

// Retention for the event log. nhp/log rotates by date and prunes nothing, so
// without this the directory grows for as long as the host lives -- slowly on a
// quiet host, and fastest exactly when the filter matters most, because every
// line is a packet somebody else chose to send. The daemon filling its own root
// volume would be a denial of service delivered through the defense.
//
// Both bounds apply, whichever bites first: age for the ordinary case (a
// fortnight is more history than any incident review here has wanted), size for
// the pathological one (a sustained flood, where a single day can outweigh the
// budget on its own). Oldest files go first, and the current day's file is
// never removed -- it is open and being appended to.
const (
	serverLogRetentionDays  = 14
	serverLogRetentionBytes = 256 << 20 // 256 MiB
	serverLogSweepInterval  = time.Hour
)

func pruneServerEventLogs(logDir string) {
	for {
		sweepServerEventLogs(logDir, time.Now(), serverLogRetentionDays, serverLogRetentionBytes)
		time.Sleep(serverLogSweepInterval)
	}
}

// sweepServerEventLogs is the body of one sweep. The bounds are arguments
// rather than the constants above so a test can exercise the size path without
// writing a quarter of a gigabyte.
func sweepServerEventLogs(logDir string, now time.Time, retainDays int, maxBytes int64) {
	matches, err := filepath.Glob(filepath.Join(logDir, serverEventLogName+"-*.log"))
	if err != nil {
		log.Error("xdp event log retention: cannot list %s: %v", logDir, err)
		return
	}

	type logFile struct {
		path string
		mod  time.Time
		size int64
	}

	current := filepath.Join(logDir, fmt.Sprintf("%s-%s.log", serverEventLogName, now.Format("2006-01-02")))
	files := make([]logFile, 0, len(matches))
	var totalBytes int64
	for _, path := range matches {
		if path == current {
			continue
		}
		info, statErr := os.Stat(path)
		if statErr != nil {
			continue
		}
		files = append(files, logFile{path: path, mod: info.ModTime(), size: info.Size()})
		totalBytes += info.Size()
	}

	sort.Slice(files, func(i, j int) bool { return files[i].mod.Before(files[j].mod) })

	cutoff := now.AddDate(0, 0, -retainDays)
	for _, f := range files {
		overAge := f.mod.Before(cutoff)
		overSize := totalBytes > maxBytes
		if !overAge && !overSize {
			break // sorted oldest first: nothing later can be over either bound
		}
		if err := os.Remove(f.path); err != nil {
			log.Error("xdp event log retention: cannot remove %s: %v", f.path, err)
			continue
		}
		totalBytes -= f.size
		log.Info("xdp event log retention: removed %s (%d bytes)", f.path, f.size)
	}
}

// ReplaceRelayIPs makes the pinned `nhp_relay_ips` map hold exactly ipStrs.
//
// Entries are host addresses ("10.0.1.4") or prefixes ("10.0.1.0/24"); see
// ParseRelayPrefix and the map's comment in nhp/ebpf/xdp/nhp_server_xdp.c for
// why the subnet form exists.
//
// New entries go in before stale ones come out, so an address that is on both
// the old and the new list is never momentarily absent: a delete-then-insert
// would open a window in which the relay's own SSH session and forwarded knocks
// are dropped by the very reload that was meant to keep them working. Callers
// serialize their own calls (see endpoints/server/ebpf).
//
// Two refusals, both made before a single map write, keep a bad list from
// becoming a lockout. An entry that does not parse fails the whole call
// (ParseRelayPrefixes says why all-or-nothing is the safe reading), and a list
// that names no prefix at all is rejected rather than applied: emptying this
// map closes tcp/22 for every source on a host whose only way in is this map.
// Callers are expected to have refused such a list already — this is the last
// place that can still tell, and it does not rely on them.
//
// A failed write is the third way a reload can weaken a working whitelist, and
// it is handled the same way: the call is all-or-nothing against the map, not
// only against the file. The map is read before anything is written so this
// call knows which keys are its own; if an Update fails — an ENOMEM on the
// trie, an EINVAL from a key the kernel will not hold — the keys this call
// added are taken back out and the stale sweep never runs, so the map is left
// holding exactly the list it held before. Deleting stale entries anyway, which
// is what the first version did, is how a reload that merely failed to add the
// relay's new address ends up removing its old one too: a map holding neither,
// i.e. tcp/22 closed to every source, under a caller logging that it kept the
// active whitelist.
func ReplaceRelayIPs(relayMap *ebpf.Map, ipStrs []string) error {
	if relayMap == nil {
		return fmt.Errorf("relay ip map is not loaded")
	}

	keys, err := ParseRelayPrefixes(ipStrs)
	if err != nil {
		return fmt.Errorf("refusing to apply the relay whitelist: %w", err)
	}
	if len(keys) == 0 {
		return fmt.Errorf("refusing to empty the relay whitelist: the new list names no prefix, and a map with no prefix drops SSH from every source")
	}

	want := make(map[RelayPrefixKey]struct{}, len(keys))
	for _, key := range keys {
		want[key] = struct{}{}
	}

	// Read the active map first. Rolling back needs to know which keys this
	// call introduced, and a snapshot taken after the adds cannot tell them
	// from entries that were already there. A map this call cannot read is a
	// map it cannot change safely, so it writes nothing at all: the live
	// whitelist stays exactly as it is, which is the direction that keeps a
	// reachable host reachable.
	present := make(map[RelayPrefixKey]struct{})
	var iterKey RelayPrefixKey
	var iterValue uint8
	iter := relayMap.Iterate()
	for iter.Next(&iterKey, &iterValue) {
		present[iterKey] = struct{}{}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("refusing to apply the relay whitelist: cannot read the active map: %w", err)
	}

	var added []RelayPrefixKey
	allowed := uint8(1)
	for key := range want {
		k := key
		if _, old := present[k]; old {
			continue // already allowed; leave it untouched so it is never momentarily absent
		}
		if err := relayMap.Update(&k, &allowed, ebpf.UpdateAny); err != nil {
			log.Error("relay whitelist: failed to add %s: %v", k, err)
			rollbackRelayIPs(relayMap, added)
			return fmt.Errorf("refusing to apply the relay whitelist: failed to add %s: %w", k, err)
		}
		added = append(added, k)
	}

	var stale []RelayPrefixKey
	for key := range present {
		if _, keep := want[key]; !keep {
			stale = append(stale, key)
		}
	}

	var firstErr error
	for _, k := range stale {
		key := k
		if err := relayMap.Delete(&key); err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
			log.Error("relay whitelist: failed to remove %s: %v", key, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	log.Info("relay whitelist applied: %d prefix(es) active, %d added, %d removed", len(want), len(added), len(stale))
	return firstErr
}

// rollbackRelayIPs removes the keys a failed ReplaceRelayIPs had already added,
// restoring the whitelist the map held when that call started.
//
// A delete that fails here leaves the map holding more prefixes than the old
// list, never fewer — the error path widens the whitelist rather than closing
// tcp/22 — so it is logged and the rest are still attempted.
func rollbackRelayIPs(relayMap *ebpf.Map, added []RelayPrefixKey) {
	for _, k := range added {
		key := k
		if err := relayMap.Delete(&key); err != nil && !errors.Is(err, ebpf.ErrKeyNotExist) {
			log.Error("relay whitelist: failed to roll back %s after a failed update: %v", key, err)
		}
	}
	if len(added) > 0 {
		log.Info("relay whitelist unchanged: rolled back %d prefix(es) added before the failure", len(added))
	}
}

// setEbpfConfig fills the `nhp_config` map the TC egress program reads.
//
// The only knob so far is the low bound of the host's ephemeral port range. The
// TC program uses it to tell a socket the AC opened as a client (which binds a
// port from that range) from a service it listens on, and that decides whether
// the flow gets a conn_track entry letting the replies back in. Hard-coding the
// Linux default would silently drop the replies to the AC's own UDP flows on
// any host that lowers net.ipv4.ip_local_port_range — including nhp-acd's
// channel to the nhp-server, which would take the AC off the air.
func setEbpfConfig(tcObjs *tcBpfObjects) error {
	if tcObjs.Config == nil {
		return fmt.Errorf("'nhp_config' map not found")
	}

	portMin, err := localEphemeralPortMin()
	if err != nil {
		// Not fatal: the eBPF side falls back to the same default when the
		// slot is left at 0, so a host that hides the sysctl still works as
		// long as it has not moved the range.
		log.Error("failed to read net.ipv4.ip_local_port_range, assuming %d: %v", defaultEphemeralPortMin, err)
		portMin = defaultEphemeralPortMin
	}
	log.Info("eBPF egress tracking treats source ports >= %d as the AC's own (net.ipv4.ip_local_port_range), plus the well-known client ports listed in is_wellknown_client_port()", portMin)

	if err := tcObjs.Config.Put(cfgEphemeralPortMin, portMin); err != nil {
		log.Error("failed to write 'nhp_config' map: %v", err)
		return err
	}
	return nil
}

// localEphemeralPortMin reads the low bound of net.ipv4.ip_local_port_range,
// whose format is "<low>\t<high>".
func localEphemeralPortMin() (uint32, error) {
	const sysctlPath = "/proc/sys/net/ipv4/ip_local_port_range"

	content, err := os.ReadFile(sysctlPath)
	if err != nil {
		return 0, err
	}

	fields := strings.Fields(string(content))
	if len(fields) == 0 {
		return 0, fmt.Errorf("%s is empty", sysctlPath)
	}

	portMin, err := strconv.ParseUint(fields[0], 10, 16)
	if err != nil {
		return 0, fmt.Errorf("%s: unparsable low bound %q: %v", sysctlPath, fields[0], err)
	}
	if portMin == 0 {
		return 0, fmt.Errorf("%s: low bound is 0", sysctlPath)
	}
	return uint32(portMin), nil
}

// resolveInterface returns the interface to attach to, defaulting to the one the
// host's default route exits through.
//
// A missing interface is returned as an error. It used to be os.Exit(1), which
// killed the daemon from inside a library call; both callers handle an error
// (the AC refuses to start, the server falls back to no filtering and logs it).
func resolveInterface(name string) (*net.Interface, error) {
	if name == "" {
		var err error
		name, err = getDefaultRouteInterface()
		if err != nil {
			log.Error("failed to get default route interface")
			return nil, err
		}
		log.Info("Default route interface: %s\n", name)
	}
	iface, err := net.InterfaceByName(name)
	if err != nil {
		log.Error("failed to find interface %s", name)
		return nil, err
	}
	return iface, nil
}

func uint32ToIPv4(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(ip>>24)&0xff,
		(ip>>16)&0xff,
		(ip>>8)&0xff,
		ip&0xff)
}

func getDefaultRouteInterface() (string, error) {
	cmd := exec.Command("ip", "route")
	output, err := cmd.Output()
	if err != nil {
		log.Error("failed to get running ip route:")
		return "", err
	}

	re := regexp.MustCompile(`default via (\S+) dev (\S+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 3 {
		log.Error("failed to parse default route")
		return "", fmt.Errorf("failed to parse default route")
	}
	interfaceName := matches[2]
	return interfaceName, nil
}

// CleanupBPFFiles removes the pins the given variant owns and detaches its
// links. Only that variant's list is touched, so the AC and the server can
// never tear each other's maps down, and only the directory this process last
// loaded into (EngineLoadParams.PinDir, DefaultPinDir until then) is swept.
func CleanupBPFFiles(variant EngineVariant) {
	cleanupPins(variant, activePinDir(variant))
}

// cleanupPins is CleanupBPFFiles with the pin directory spelled out, for
// callers that know it without having gone through EngineLoad — the privileged
// tests, which must not reach into /sys/fs/bpf and delete the pins of a filter
// running on the same host.
func cleanupPins(variant EngineVariant, pinDir string) {
	for _, file := range pinnedFiles(variant, pinDir) {
		if err := os.Remove(file); err != nil {
			if !os.IsNotExist(err) {
				log.Error("Failed to remove BPF file %s: %v", file, err)
			}
		} else {
			log.Info("Successfully removed BPF file: %s", file)
		}
	}

	if variant == VariantServer {
		if serverXdpLink != nil {
			serverXdpLink.Close()
			serverXdpLink = nil
			log.Info("server XDP link detached and closed")
		}
		return
	}

	// Cleared as well as closed, the same as the server's link above: this can
	// run twice (an error path inside loadAcEngine, then the caller's shutdown)
	// and closing a link a second time is an error on an already-freed handle.
	if acXdpLink != nil {
		acXdpLink.Close()
		acXdpLink = nil
		log.Info("XDP link detached and closed")
	}
	if acTcLink != nil {
		acTcLink.Close()
		acTcLink = nil
		log.Info("TCX link detached and closed")
	}
}

func protoToString(proto uint8) string {
	switch proto {
	case 6:
		return "TCP"
	case 17:
		return "UDP"
	case 1:
		return "ICMP"
	case 2:
		return "IGMP"
	case 41:
		return "IPv6"
	case 47:
		return "GRE"
	case 50:
		return "ESP"
	case 51:
		return "AH"
	case 58:
		return "ICMPv6"
	case 88:
		return "EIGRP"
	case 89:
		return "OSPF"
	case 112:
		return "VRRP"
	default:
		return fmt.Sprintf("PROTO-%d", proto)
	}
}
