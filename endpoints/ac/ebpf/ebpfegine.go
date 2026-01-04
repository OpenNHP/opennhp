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
	"syscall"
	"time"

	stdlog "log"

	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
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
	Whitelist    *ebpf.Map     `ebpf:"spp"`
}

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
	if err := spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf/", // automatically mounted to
		},
	}); err != nil {
		log.Error("Failed to load and assign eBPF objects")
		return err
	}

	var tcObjs tcBpfObjects
	if err := tcSpec.LoadAndAssign(&tcObjs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf/", // automatically mounted to
		},
	}); err != nil {
		log.Error("Failed to load and assign tc eBPF objects")
		return err
	}

	if err := objs.XdpProg.Pin("/sys/fs/bpf/xdp_white_prog"); err != nil {
		log.Error("failed to pin XDP program xdp_white_prog to /sys/fs/bpf/")
		return err
	}
	if err := tcObjs.TcEgressProg.Pin("/sys/fs/bpf/tc_egress_prog"); err != nil {
		log.Error("failed to pin TC egress program tc_egress_prog to /sys/fs/bpf/")
		return err
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
