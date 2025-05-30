//go:build linux

package utils

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
)

type WhitelistKey struct {
	SrcIP    uint32 `ebpf:"src_ip"`
	DstIP    uint32 `ebpf:"dst_ip"`
	DstPort  uint16 `ebpf:"dst_port"`
	Protocol uint8  `ebpf:"protocol"`
}

type EbpfMapKey struct {
	SrcIP    uint32 `ebpf:"src_ip"`
	DstIP    uint32 `ebpf:"dst_ip"`
	DstPort  uint16 `ebpf:"dst_port"`
	Protocol uint8  `ebpf:"protocol"`
}

type WhitelistKeyicmp struct {
	SrcIP uint32 `ebpf:"src_ip"`
	DstIP uint32 `ebpf:"dst_ip"`
}

type WhitelistValue struct {
	Allowed    uint8
	_          [7]byte
	ExpireTime uint64
}

func getBootTimeNanos() (uint64, error) {
	var ts unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &ts); err != nil {
		return 0, fmt.Errorf("clock_gettime failed: %v", err)
	}
	return uint64(ts.Sec)*1e9 + uint64(ts.Nsec), nil
}

func xdp_manager() {

	// 直接从 /sys/fs/bpf/ 加载已固定的 map
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/whitelist", nil)
	if err != nil {
		log.Error("Failed to load pinned whitelist map: %v", err)
	}
	defer whitelistMap.Close()

	icmpwhitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/icmpwhitelist", nil)
	if err != nil {
		log.Error("Failed to load pinned icmpwhitelist map: %v", err)
	}
	defer icmpwhitelistMap.Close()

	conntrackMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/conn_track", nil)
	if err != nil {
		log.Error("Failed to load pinned conn_track map: %v", err)
	}
	defer conntrackMap.Close()

	// 提供命令行接口来管理白名单规则
	fmt.Println("XDP White Program Manager is running. Type commands to manage whitelist:")
	fmt.Println("add <src_ip> <dst_ip> <dst_port> <protocol> - Add a whitelist rule")
	fmt.Println("del <src_ip> <dst_ip> <dst_port> <protocol> - Delete a whitelist rule")
	fmt.Println("exit - Exit the program")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		switch cmd {
		case "add":
			if len(parts) != 6 {
				fmt.Println("Usage: add <src_ip> <dst_ip> <dst_port> <protocol> <ttl_seconds> - Add a whitelist rule with TTL")
				continue
			}
			srcIP, err := parseIP(parts[1])
			if err != nil {
				fmt.Printf("Invalid source IP: %v\n", err)
				continue
			}
			dstIP, err := parseIP(parts[2])
			if err != nil {
				fmt.Printf("Invalid destination IP: %v\n", err)
				continue
			}
			dstPort, err := parsePort(parts[3])
			if err != nil {
				fmt.Printf("Invalid destination port: %v\n", err)
				continue
			}
			protocol, err := parseProtocol(parts[4])
			if err != nil {
				fmt.Printf("Invalid protocol: %v\n", err)
				continue
			}

			ttlSec, err := strconv.ParseUint(parts[5], 10, 64)
			if err != nil {
				fmt.Printf("Invalid TTL: %v\n", err)
				continue
			}

			if err := WhitelistRule(whitelistMap, srcIP, dstIP, dstPort, protocol, ttlSec); err != nil {
				log.Error("Failed to add whitelist rule: %v", err)
			} else {
				fmt.Println("Whitelist rule added successfully.")
			}
		case "del":
			if len(parts) != 5 {
				fmt.Println("Usage: del <src_ip> <dst_ip> <dst_port> <protocol>")
				continue
			}
			srcIP, err := parseIP(parts[1])
			if err != nil {
				fmt.Printf("Invalid source IP: %v\n", err)
				continue
			}
			dstIP, err := parseIP(parts[2])
			if err != nil {
				fmt.Printf("Invalid destination IP: %v\n", err)
				continue
			}
			dstPort, err := parsePort(parts[3])
			if err != nil {
				fmt.Printf("Invalid destination port: %v\n", err)
				continue
			}
			protocol, err := parseProtocol(parts[4])
			if err != nil {
				fmt.Printf("Invalid protocol: %v\n", err)
				continue
			}

			if err := delWhitelistRule(whitelistMap, srcIP, dstIP, dstPort, protocol); err != nil {
				log.Error("Failed to delete whitelist rule: %v", err)
			} else {
				fmt.Println("Whitelist rule deleted successfully.")
			}
		case "exit":
			os.Exit(0)
		default:
			fmt.Println("Unknown command. Available commands: add, del, exit")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error("Error reading input: %v", err)
	}
}

// 解析 IP 地址
func parseIP(ipStr string) (uint32, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		log.Error("invalid IP address: %s", ipStr)
		return 0, fmt.Errorf("invalid IP address: %s", ipStr)
	}
	ip = ip.To4()
	if ip == nil {
		log.Error("only IPv4 addresses are supported: %s", ipStr)
		return 0, fmt.Errorf("only IPv4 addresses are supported: %s", ipStr)
	}
	return binary.LittleEndian.Uint32(ip), nil
}

func parsePort(portStr string) (uint16, error) {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16([]byte{byte(port >> 8), byte(port & 0xFF)}), nil
}

func parseProtocol(protoStr string) (uint8, error) {
	proto, err := strconv.ParseUint(protoStr, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(proto), nil
}

func WhitelistRule(whitelistMap *ebpf.Map, srcIP, dstIP uint32, dstPort uint16, protocol uint8, ttlSec uint64) error {
	now, err := getBootTimeNanos()
	if err != nil {
		return err
	}
	key := WhitelistKey{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		DstPort:  dstPort,
		Protocol: protocol,
	}
	value := WhitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9, // ​​Convert to nanoseconds
	}

	keyBytes := make([]byte, 11)
	binary.LittleEndian.PutUint32(keyBytes[0:4], key.SrcIP)
	binary.LittleEndian.PutUint32(keyBytes[4:8], key.DstIP)
	binary.LittleEndian.PutUint16(keyBytes[8:10], key.DstPort)
	keyBytes[10] = key.Protocol

	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		log.Error("failed to update whitelist map: %v", err)
		return fmt.Errorf("failed to update whitelist map: %v", err)
	}

	return nil
}

func WhitelistRuleNoProtocol(whitelistMap *ebpf.Map, srcIP, dstIP uint32, ttlSec uint64) error {
	now, err := getBootTimeNanos()
	if err != nil {
		return err
	}
	key := WhitelistKey{
		SrcIP: srcIP,
		DstIP: dstIP,
	}
	value := WhitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}

	keyBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(keyBytes[0:4], key.SrcIP)
	binary.LittleEndian.PutUint32(keyBytes[4:8], key.DstIP)

	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		log.Error("failed to update whitelist map: %v", err)
		return fmt.Errorf("failed to update whitelist map: %v", err)
	}

	return nil
}

func WhitelistPortRule(whitelistMap *ebpf.Map, srcIP uint32, dstPort int, ttlSec uint64) error {
	now, err := getBootTimeNanos()
	if err != nil {
		return err
	}
	udstPort, err := safeIntToUint16(dstPort)

	if err != nil {
		log.Error("failed to safeIntToUint16 in src_port_list map: %v", err)
		return err
	}

	key := WhitelistKey{
		SrcIP:   srcIP,
		DstPort: udstPort,
	}

	value := WhitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}
	keyBytes := make([]byte, 6)
	binary.LittleEndian.PutUint32(keyBytes[0:4], key.SrcIP)
	binary.LittleEndian.PutUint16(keyBytes[4:6], key.DstPort)

	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		log.Error("failed to update src_port_list map: %v", err)
		return fmt.Errorf("failed to update src_port_list map: %v", err)
	}

	return nil
}

func delWhitelistRule(whitelistMap *ebpf.Map, srcIP, dstIP uint32, dstPort uint16, protocol uint8) error {
	key := WhitelistKey{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		DstPort:  dstPort,
		Protocol: protocol,
	}

	if err := whitelistMap.Delete(&key); err != nil {
		log.Error("failed to delete whitelist rule: %v", err)
		return fmt.Errorf("failed to delete whitelist rule: %v", err)
	}

	return nil
}

func AddEbpfRuleForSrcDstPortProto(srcIPStr, dstIPStr string, protocol uint8, dstPort uint16, ttlSec uint64) error {
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/whitelist", nil)
	if err != nil {
		log.Error("failed to load pinned whitelist map: %v", err)
		return fmt.Errorf("failed to load pinned whitelist map: %v", err)
	}
	defer whitelistMap.Close()

	srcIP, err := parseIP(srcIPStr)
	if err != nil {
		log.Error("invalid source IP: %v", err)
		return fmt.Errorf("invalid source IP: %v", err)
	}

	dstIP, err := parseIP(dstIPStr)
	if err != nil {
		log.Error("invalid destination IP: %v", err)
		return fmt.Errorf("invalid destination IP: %v", err)
	}

	err = WhitelistRule(whitelistMap, srcIP, dstIP, dstPort, protocol, ttlSec)
	if err != nil {
		log.Error("[HandleAccessControl] add ipset %s error: %v")
		return err
	}

	return nil
}

func AddEbpfRuleForSrcDst(srcIPStr, dstIPStr string, ttlSec uint64) error {
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/sdwhitelist", nil)
	if err != nil {
		log.Error("failed to load pinned whitelist map: %v", err)
		return fmt.Errorf("failed to load pinned whitelist map: %v", err)
	}
	defer whitelistMap.Close()

	srcIP, err := parseIP(srcIPStr)
	if err != nil {
		log.Error("invalid source IP: %v", err)
		return fmt.Errorf("invalid source IP: %v", err)
	}

	dstIP, err := parseIP(dstIPStr)
	if err != nil {
		log.Error("invalid destination IP: %v", err)
		return fmt.Errorf("invalid destination IP: %v", err)
	}

	err = WhitelistRuleNoProtocol(whitelistMap, srcIP, dstIP, ttlSec)
	if err != nil {
		log.Error("[HandleAccessControl] add ipset %s error: %v")
		return err
	}

	return nil
}

func AddEbpfRuleForSrcDestPort(srcIPStr string, dstPort int, ttlSec uint64) error {
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/src_port_list", nil)
	if err != nil {
		log.Error("failed to load pinned whitelist map: %v", err)
		return fmt.Errorf("failed to load pinned whitelist map: %v", err)
	}
	defer whitelistMap.Close()

	srcIP, err := parseIP(srcIPStr)
	if err != nil {
		log.Error("invalid source IP: %v", err)
		return fmt.Errorf("invalid source IP: %v", err)
	}

	err = WhitelistPortRule(whitelistMap, srcIP, dstPort, ttlSec)
	if err != nil {
		log.Error("[HandleAccessControl] add ipset %s error: %v")
		return err
	}

	return nil
}

func icmpWhitelistRule(whitelistMap *ebpf.Map, srcIP, dstIP uint32, ttlSec uint64) error {
	now, err := getBootTimeNanos()
	if err != nil {
		return err
	}
	key := WhitelistKeyicmp{
		SrcIP: srcIP,
		DstIP: dstIP,
	}

	value := WhitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}

	keyBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(keyBytes[0:4], key.SrcIP)
	binary.LittleEndian.PutUint32(keyBytes[4:8], key.DstIP)
	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		return fmt.Errorf("failed to update whitelist map: %v", err)
	}

	return nil
}

func AddEbpfIcmpRuleForSrcDst(srcIPStr, dstIPStr string, ttlSec uint64) error {
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/icmpwhitelist", nil)
	if err != nil {
		log.Error("failed to load pinned whitelist map: %v", err)
		return fmt.Errorf("failed to load pinned whitelist map: %v", err)
	}
	defer whitelistMap.Close()

	srcIP, err := parseIP(srcIPStr)
	if err != nil {
		log.Error("invalid source IP: %v", err)
		return fmt.Errorf("invalid source IP: %v", err)
	}

	dstIP, err := parseIP(dstIPStr)
	if err != nil {
		log.Error("invalid destination IP: %v", err)
		return fmt.Errorf("invalid destination IP: %v", err)
	}

	err = icmpWhitelistRule(whitelistMap, srcIP, dstIP, ttlSec)
	if err != nil {
		log.Error("failed to add rule: %v", err)
		return fmt.Errorf("failed to add rule: %v", err)
	}

	return nil
}

func safeIntToUint16(i int) (uint16, error) {
	if i < 0 || i > 65535 {
		return 0, fmt.Errorf("value %d is out of range for uint16", i)
	}
	return uint16(i), nil
}

const (
	MapTypeWhitelist     = 1
	MapTypeSdWhitelist   = 2
	MapTypeIcmpWhitelist = 3
	MapTypeSrcPortList   = 4
)

type EbpfRuleParams struct {
	SrcIP    string
	DstIP    string
	DstPort  int
	Protocol string
}

// A generic entry function that calls the corresponding function to add whitelist entries based on mapTypeandparams.
func EbpfRuleAdd(mapType int, params EbpfRuleParams, TtlSec int) error {
	var err error
	TtlSec64 := uint64(TtlSec)
	var protocol uint8
	if len(params.Protocol) > 0 {
		switch params.Protocol {
		case "tcp":
			protocol = 6
		case "udp":
			protocol = 17
		case "icmp":
			protocol = 1
		default:
			return fmt.Errorf("unsupported protocol: %s", params.Protocol)
		}
	}

	switch mapType {
	case MapTypeWhitelist:
		err = AddEbpfRuleForSrcDstPortProto(params.SrcIP, params.DstIP, protocol, uint16(params.DstPort), TtlSec64)
		if err != nil {
			log.Error("[EbpfRuleAdd] add ebpf src: %s dst: %s,  error: %v, protocol: %d, dstport :%d, %v", params.SrcIP, params.DstIP, protocol, uint16(params.DstPort), err)
			return err
		}

	case MapTypeSdWhitelist:

		err = AddEbpfRuleForSrcDst(params.SrcIP, params.DstIP, TtlSec64)
		if err != nil {
			log.Error("[EbpfRuleAdd] add ebpf src: %s dst: %s,  error: %v", params.SrcIP, params.DstIP, err)
			return err
		}

	case MapTypeIcmpWhitelist:

		err = AddEbpfIcmpRuleForSrcDst(params.SrcIP, params.DstIP, TtlSec64)
		if err != nil {
			log.Error("[EbpfRuleAdd] add ebpf icmp src: %s dst: %s,  error: %v", params.SrcIP, params.DstIP, err)
			return err
		}

	case MapTypeSrcPortList:

		err = AddEbpfRuleForSrcDestPort(params.SrcIP, params.DstPort, TtlSec64)
		if err != nil {
			log.Error("[EbpfRuleAdd] add ebpf dst port src: %s dstport: %s,  error: %v", params.SrcIP, params.DstPort, err)
			return err
		}

	default:
		return fmt.Errorf("unsupported map type: %d", mapType)
	}

	return nil
}
