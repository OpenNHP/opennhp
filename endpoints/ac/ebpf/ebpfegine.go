//go:build linux

package ebpf

import (
	// "log"

	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
}

type tcBpfObjects struct {
	TcEgressProg *ebpf.Program `ebpf:"tc_egress_prog"`
	// The TC egress program records the AC's own outbound connections in
	// conn_track (shared with the XDP program through its pin) so their
	// replies get back in. It no longer writes the knock whitelist `spp`; it
	// only reads the whitelist maps, to tell a peer that holds a knock apart
	// from a host the AC itself connected to. Those maps are not listed here
	// because nothing in user space drives them through this object — the XDP
	// collection above is loaded first and creates every pin.
	//
	// Conntrack is not used after the load, but keeping it assigned means a
	// pin left over from an incompatible build fails here, at startup, with a
	// `conn_track` error naming the map — see "eBPF map pin mismatch" in
	// terraform/demo/RUNBOOK.md.
	Conntrack *ebpf.Map `ebpf:"conn_track"`
	// Runtime knobs for the TC program, filled in by setEbpfConfig below.
	Config *ebpf.Map `ebpf:"nhp_config"`
}

// Config slots in the `nhp_config` map. Keep in sync with the enum in
// nhp/ebpf/xdp/nhp_maps.h.
const cfgEphemeralPortMin uint32 = 0

// defaultEphemeralPortMin mirrors NHP_EPHEMERAL_PORT_MIN, the fallback the eBPF
// program uses when the slot is unset.
const defaultEphemeralPortMin uint32 = 32768

var (
	DenyLogger *log.Logger
	AcLogger   *log.Logger
)

type Event struct {
	Timestamp  uint64 `ebpf:"timestamp"`
	Action     uint8  `ebpf:"action"`
	SrcIP      uint32 `ebpf:"src_ip"`
	DstIP      uint32 `ebpf:"dst_ip"`
	SrcPort    uint16 `ebpf:"src_port"`
	DstPort    uint16 `ebpf:"dst_port"`
	Protocol   uint8  `ebpf:"protocol"`
	PayloadLen uint16 `ebpf:"payload_len"`
}

var xdpLink link.Link
var tcLink link.Link
var bootTime time.Time

func init() {
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err != nil {
		panic("Failed to get the system running time: " + err.Error())
	}

	now := time.Now()
	bootTime = now.Add(-time.Duration(info.Uptime) * time.Second)
	log.Info("​​System boot time: %v", bootTime)
}

func EbpfEngineLoad(dirPath string, logLevel int, acId string) error {
	CleanupBPFFiles()
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Error("Failed to remove memlock limit")
	}

	const ebpfenginename string = "nhp_ebpf_xdp.o"
	const tcObjName string = "tc_egress.o"
	//ebpf nhp_ebpf_xdp.o save to etc/ after clang compile
	bpfDir := "etc"
	specPath := filepath.Join(bpfDir, ebpfenginename)
	tcSpecPath := filepath.Join(bpfDir, tcObjName)

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		log.Error("eBPF object file not found ")
		return err
	}
	if _, err := os.Stat(tcSpecPath); os.IsNotExist(err) {
		log.Error("tc eBPF object file not found ")
		return err
	}

	spec, err := ebpf.LoadCollectionSpec(specPath)
	if err != nil {
		log.Error("failed to load eBPF object")
		return err
	}
	// Load tc eBPF object
	tcSpec, err := ebpf.LoadCollectionSpec(tcSpecPath)
	if err != nil {
		log.Error("failed to load tc eBPF object")
		return err
	}

	var objs bpfObjects
	if loadErr := spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf/", // automatically mounted to
		},
	}); loadErr != nil {
		log.Error("Failed to load and assign eBPF objects")
		return loadErr
	}

	var tcObjs tcBpfObjects
	if tcLoadErr := tcSpec.LoadAndAssign(&tcObjs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf/", // automatically mounted to
		},
	}); tcLoadErr != nil {
		log.Error("Failed to load and assign tc eBPF objects")
		return tcLoadErr
	}

	if cfgErr := setEbpfConfig(&tcObjs); cfgErr != nil {
		log.Error("Failed to configure tc eBPF program")
		return cfgErr
	}

	if pinErr := objs.XdpProg.Pin("/sys/fs/bpf/xdp_white_prog"); pinErr != nil {
		log.Error("failed to pin XDP program xdp_white_prog to /sys/fs/bpf/")
		return pinErr
	}
	if tcPinErr := tcObjs.TcEgressProg.Pin("/sys/fs/bpf/tc_egress_prog"); tcPinErr != nil {
		log.Error("failed to pin TC egress program tc_egress_prog to /sys/fs/bpf/")
		return tcPinErr
	}

	ifaceName, err := getDefaultRouteInterface()
	if err != nil {
		log.Error("failed to get default route interface")
		return err
	}
	log.Info("Default route interface: %s\n", ifaceName)
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Error("failed to find interface %s", ifaceName)
		os.Exit(1)
	}
	//load ebpf nhp_ebpf_xdp.o to net interface which default route exit
	xdpLink, err = link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpProg,
		Interface: iface.Index,
		Flags:     link.XDPGenericMode, // XDPGenericMode and XDPDriverMode
	})
	if err != nil {
		log.Error("failed to attach XDP program to interface: %s", ifaceName)
		return err
	}
	//load tc eBPF tc_egress.o to net interface which default route exit
	tcLink, err = link.AttachTCX(link.TCXOptions{
		Program:   tcObjs.TcEgressProg,
		Interface: iface.Index,
		Attach:    ebpf.AttachTCXEgress,
	})
	if err != nil {
		log.Error("failed to attach TC egress program to interface: %s", ifaceName)
		return err
	}

	// Accessing the Perf Buffer Map named "events" defined in eBPF.
	eventsMap := objs.Events
	if eventsMap == nil {
		log.Error("failed to load 'events' map from eBPF object (nil)")
		return fmt.Errorf("'events' map not found")
	}

	ExeDirPath := dirPath
	//Set up the DENY logger
	DenyLogger = log.NewLoggerDefine(
		"",
		logLevel,
		filepath.Join(ExeDirPath, "logs"),
		"nhp_deny",
	)
	DenyLogger.SetFlags(stdlog.Lmsgprefix)
	// Set up the ACCEPT logger
	AcLogger = log.NewLoggerDefine(
		"",
		logLevel,
		filepath.Join(ExeDirPath, "logs"),
		"nhp_accept",
	)
	AcLogger.SetFlags(stdlog.Lmsgprefix)
	// Start a goroutine to monitor Perf Buffer events
	go func() {
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
				DenyLogger.Info("%s", logMsg)
			} else { // ACCEPT
				AcLogger.Info("%s", logMsg)
			}

		}
	}()

	return nil
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

func uint32ToIPv4(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(ip>>24)&0xff,
		(ip>>16)&0xff,
		(ip>>8)&0xff,
		ip&0xff)
}

func ipUint32ToString(ip uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip&0xFF,
		(ip>>8)&0xFF,
		(ip>>16)&0xFF,
		(ip>>24)&0xFF)
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

// clean eBPF map file
func CleanupBPFFiles() {
	bpfFiles := []string{
		"/sys/fs/bpf/xdp_white_prog",
		"/sys/fs/bpf/conn_track",
		"/sys/fs/bpf/icmpwhitelist",
		"/sys/fs/bpf/port_list",
		"/sys/fs/bpf/protocol_port",
		"/sys/fs/bpf/sdwhitelist",
		"/sys/fs/bpf/src_port",
		"/sys/fs/bpf/spp",
		"/sys/fs/bpf/nhp_config",
		"/sys/fs/bpf/tc_egress_prog",
	}

	for _, file := range bpfFiles {
		if err := os.Remove(file); err != nil {
			if !os.IsNotExist(err) {
				log.Error("Failed to remove BPF file %s: %v", file, err)
			}
		} else {
			log.Info("Successfully removed BPF file: %s", file)
		}
	}
	if xdpLink != nil {
		xdpLink.Close()
		log.Info("XDP link detached and closed")
	}
	if tcLink != nil {
		tcLink.Close()
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
