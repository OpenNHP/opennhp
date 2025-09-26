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

	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"github.com/cilium/ebpf/rlimit"
)

type bpfObjects struct {
	XdpProg       *ebpf.Program `ebpf:"xdp_white_prog"`
	Whitelist     *ebpf.Map     `ebpf:"whitelist"`
	Icmpwhitelist *ebpf.Map     `ebpf:"icmpwhitelist"`
	Sdwhitelist   *ebpf.Map     `ebpf:"sdwhitelist"`
	Srcportlist   *ebpf.Map     `ebpf:"src_port_list"`
	Portlist      *ebpf.Map     `ebpf:"port_list"`
	Protocolport  *ebpf.Map     `ebpf:"protocol_port"`
	Conntrack     *ebpf.Map     `ebpf:"conn_track"`
	Events        *ebpf.Map     `ebpf:"events"`
}

// 定义与 eBPF 中的 event_t 完全一致的结构体
// 用于接收从 Perf Buffer 传来的事件
type Event struct {
	Timestamp uint64 `ebpf:"timestamp"` // 纳秒级时间戳
	Action    uint8  `ebpf:"action"`    // 0 = DENY, 1 = ACCEPT
	SrcIP     uint32 `ebpf:"src_ip"`    // 源IP（网络字节序）
	DstIP     uint32 `ebpf:"dst_ip"`    // 目的IP（网络字节序）
	SrcPort   uint16 `ebpf:"src_port"`  // 源端口（主机字节序）
	DstPort   uint16 `ebpf:"dst_port"`  // 目的端口（主机字节序）
	Protocol  uint8  `ebpf:"protocol"`  // 协议号，如 6=TCP, 17=UDP
}

var xdpLink link.Link
var bootTime time.Time

func init() {
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err != nil {
		panic("无法获取系统运行时间: " + err.Error())
	}

	now := time.Now()
	bootTime = now.Add(-time.Duration(info.Uptime) * time.Second)
	log.Info("系统启动时间: %v", bootTime)
}

func EbpfEngineLoad(dirPath string, logLevel int) error {
	CleanupBPFFiles()
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Error("Failed to remove memlock limit")
	}

	const ebpfenginename string = "nhp_ebpf_xdp.o"
	//ebpf nhp_ebpf_xdp.o save to etc/ after clang compile
	bpfDir := "etc"
	specPath := filepath.Join(bpfDir, ebpfenginename)

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		log.Error("eBPF object file not found ")
		return err
	}

	spec, err := ebpf.LoadCollectionSpec(specPath)
	if err != nil {
		log.Error("failed to load eBPF object")
		return err
	}

	var objs bpfObjects
	if err := spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf/", // automatically mounted to
		},
	}); err != nil {
		log.Error("Failed to load and assign eBPF objects")
	}

	if err := objs.XdpProg.Pin("/sys/fs/bpf/xdp_white_prog"); err != nil {
		log.Error("failed to pin XDP program xdp_white_prog to /sys/fs/bpf/")
		return err
	}
	// obtain the nic interface which default route exit
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

	log.Info("Successfully attached XDP program to interface: %s", ifaceName)

	ip := uint32(0x320310AC)                                                   // 小端表示的 172.16.3.50
	log.Info("测试函数ipUint32ToString uint32(0x320310AC)" + ipUint32ToString(ip)) // 输出：172.16.3.50 ✅

	// 加载 eBPF 程序，假设已经成功到 var objs bpfObjects

	// 访问 eBPF 中定义的名为 "events" 的 Perf Buffer Map
	eventsMap := objs.Events // ✅ 不是 objs.Maps["events"]
	if eventsMap == nil {
		log.Error("failed to load 'events' map from eBPF object (nil)")
		return fmt.Errorf("'events' map not found")
	}
	// common.ExeDirPath = dirPath
	ExeDirPath := dirPath
	denyLogger := log.NewLogger(
		"NHP-AC",                          // prepend
		logLevel,                          // level
		filepath.Join(ExeDirPath, "logs"), // dir: 日志目录
		"nhp_deny",                        // filename: 前缀
	)

	// 创建 ACCEPT 日志器
	acLogger := log.NewLogger(
		"NHP-AC",
		logLevel,
		filepath.Join(ExeDirPath, "logs"), // dir: 日志目录
		"nhp_accept",
	)

	log.Info("已创建日志器: nhp_deny-* 和 nhp_accept-*")
	// 启动 goroutine 监听 Perf Buffer 事件
	go func() {
		perfReader, err := perf.NewReader(eventsMap, os.Getpagesize())
		if err != nil {
			log.Error("failed to create perf reader:", err)
			return
		}
		defer perfReader.Close()

		log.Info("[*] 开始监听 eBPF 事件（PERF BUFFER）...")

		for {
			record, err := perfReader.Read()
			if err != nil {
				log.Error("读取事件错误:", err)
				continue
			}
			action := record.RawSample[8]
			var actionStr string
			switch action {
			case 0:
				actionStr = "NHP-DENY"
			case 1:
				actionStr = "NHP-ACCEPT"
			default:
				actionStr = "NHP-UNKNOWN"
			}
			timestamp := binary.LittleEndian.Uint64(record.RawSample[0:8])
			srcIP := binary.BigEndian.Uint32(record.RawSample[9:13])
			dstIP := binary.BigEndian.Uint32(record.RawSample[13:17])
			srcPort := binary.BigEndian.Uint16(record.RawSample[17:19])
			dstPort := binary.BigEndian.Uint16(record.RawSample[19:21])
			protocol := record.RawSample[21]

			srcIPStr := uint32ToIPv4(srcIP)
			dstIPStr := uint32ToIPv4(dstIP)
			eventTime := bootTime.Add(time.Duration(timestamp))
			log.Info("%s [%s] SRC=%s DST=%s PROTO=%d SPT=%d DPT=%d\n",
				eventTime.Format("2006-01-02 15:04:05"),
				actionStr,
				srcIPStr,
				dstIPStr,
				protocol,
				srcPort,
				dstPort,
			)

			logMsg := fmt.Sprintf("%s SRC=%s DST=%s PROTO=%d SPT=%d DPT=%d",
				eventTime.Format("2006-01-02 15:04:05"),
				srcIPStr,
				dstIPStr,
				protocol,
				srcPort,
				dstPort,
			)

			if action == 0 { // DENY
				denyLogger.Info("[NHP-DENY] %s", logMsg)
			} else { // ACCEPT
				acLogger.Info("[NHP-ACCEPT] %s", logMsg)
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
		"/sys/fs/bpf/src_port_list",
		"/sys/fs/bpf/whitelist",
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
}
