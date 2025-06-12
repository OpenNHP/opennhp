package utils

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"

	"github.com/OpenNHP/opennhp/nhp/log"
	ebpflocal "github.com/OpenNHP/opennhp/nhp/utils/ebpf"
	"github.com/cilium/ebpf"
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

type procoPortKey struct {
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

type procoPortValue struct {
	Allowed    uint8
	ExpireTime uint64
}

type portListKey struct {
	SrcIP        uint32 `ebpf:"src_ip"`
	DstPortStart uint16 `ebpf:"dst_port_start"`
	DstPortEnd   uint16 `ebpf:"dst_port_end"`
}

// Parse the IP address
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

// Parse the port
func parsePort(portStr string) (uint16, error) {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16([]byte{byte(port >> 8), byte(port & 0xFF)}), nil
}

// function for update whitelist map
func WhitelistRule(whitelistMap *ebpf.Map, srcIP, dstIP uint32, dstPort uint16, protocol uint8, ttlSec uint64) error {
	now, err := ebpflocal.GetBootTimeNanos()
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

// function for update sdwhitelist map
func WhitelistRuleNoProtocol(whitelistMap *ebpf.Map, srcIP, dstIP uint32, ttlSec uint64) error {
	now, err := ebpflocal.GetBootTimeNanos()
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
		log.Error("failed to update sdwhitelist map: %v", err)
		return err
	}

	return nil
}

// function for update src_port_list map
func WhitelistPortRule(whitelistMap *ebpf.Map, srcIP uint32, dstPort int, ttlSec uint64) error {
	now, err := ebpflocal.GetBootTimeNanos()
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

func AddEbpfRuleForSrcDstPortProto(srcIPStr, dstIPStr string, protocol uint8, dstPort uint16, ttlSec uint64) error {
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/whitelist", nil)
	if err != nil {
		log.Error("failed to load pinned whitelist map: %v", err)
		return err
	}
	defer whitelistMap.Close()

	srcIP, err := parseIP(srcIPStr)
	if err != nil {
		log.Error("invalid source IP: %v", err)
		return err
	}

	dstIP, err := parseIP(dstIPStr)
	if err != nil {
		log.Error("invalid destination IP: %v", err)
		return err
	}

	err = WhitelistRule(whitelistMap, srcIP, dstIP, dstPort, protocol, ttlSec)
	if err != nil {
		log.Error("failed to update whitelist map: %v", err)
		return err
	}

	return nil
}

func AddEbpfRuleForSrcDst(srcIPStr, dstIPStr string, ttlSec uint64) error {
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/sdwhitelist", nil)
	if err != nil {
		log.Error("failed to load pinned whitelist map: %v", err)
		return err
	}
	defer whitelistMap.Close()

	srcIP, err := parseIP(srcIPStr)
	if err != nil {
		log.Error("invalid source IP: %v", err)
		return err
	}

	dstIP, err := parseIP(dstIPStr)
	if err != nil {
		log.Error("invalid destination IP: %v", err)
		return err
	}

	err = WhitelistRuleNoProtocol(whitelistMap, srcIP, dstIP, ttlSec)
	if err != nil {
		log.Error("failed to update sdwhitelist map: %v", err)
		return err
	}

	return nil
}

func AddEbpfRuleForSrcDestPort(srcIPStr string, dstPort int, ttlSec uint64) error {
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/src_port_list", nil)
	if err != nil {
		log.Error("failed to load pinned whitelist map: %v", err)
		return err
	}
	defer whitelistMap.Close()

	srcIP, err := parseIP(srcIPStr)
	if err != nil {
		log.Error("invalid source IP: %v", err)
		return err
	}

	err = WhitelistPortRule(whitelistMap, srcIP, dstPort, ttlSec)
	if err != nil {
		log.Error("failed to update src_port_list map: %v", err)
		return err
	}

	return nil
}

// function for update icmpwhitelist map
func icmpWhitelistRule(whitelistMap *ebpf.Map, srcIP, dstIP uint32, ttlSec uint64) error {
	now, err := ebpflocal.GetBootTimeNanos()
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
		log.Error("failed to update icmpwhitelist map: %v", err)
		return err
	}
	return nil
}

func AddEbpfIcmpRuleForSrcDst(srcIPStr, dstIPStr string, ttlSec uint64) error {
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/icmpwhitelist", nil)
	if err != nil {
		log.Error("failed to load pinned whitelist map: %v", err)
		return err
	}
	defer whitelistMap.Close()

	srcIP, err := parseIP(srcIPStr)
	if err != nil {
		log.Error("invalid source IP: %v", err)
		return err
	}

	dstIP, err := parseIP(dstIPStr)
	if err != nil {
		log.Error("invalid destination IP: %v", err)
		return err
	}

	err = icmpWhitelistRule(whitelistMap, srcIP, dstIP, ttlSec)
	if err != nil {
		log.Error("failed to update icmpwhitelist map: %v", err)
		return err
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
	MapTypeSrcAndPort    = 4
	MapTypeSrcPortList   = 5
	MapTypeProtocolPort  = 6
)

type EbpfRuleParams struct {
	SrcIP        string
	DstIP        string
	DstPort      int
	DstPortStart int
	DstPortEnd   int
	Protocol     string
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
		//base the map whitelist
		err = AddEbpfRuleForSrcDstPortProto(params.SrcIP, params.DstIP, protocol, uint16(params.DstPort), TtlSec64)
		if err != nil {
			log.Error("failed add ebpf src: %s dst: %s,  error: %v, protocol: %d, dstport :%d", params.SrcIP, params.DstIP, protocol, uint16(params.DstPort))
			return err
		}

	case MapTypeSdWhitelist:
		//base the map sdwhitelist
		err = AddEbpfRuleForSrcDst(params.SrcIP, params.DstIP, TtlSec64)
		if err != nil {
			log.Error("failed add ebpf src: %s dst: %s", params.SrcIP, params.DstIP)
			return err
		}

	case MapTypeIcmpWhitelist:
		//base the map icmpwhitelist
		err = AddEbpfIcmpRuleForSrcDst(params.SrcIP, params.DstIP, TtlSec64)
		if err != nil {
			log.Error("failed add ebpf icmp src: %s dst: %s", params.SrcIP, params.DstIP)
			return err
		}

	case MapTypeSrcAndPort:
		//base the map src_port_list
		err = AddEbpfRuleForSrcDestPort(params.SrcIP, params.DstPort, TtlSec64)
		if err != nil {
			log.Error("failed add ebpf src: %s dst port: %d", params.SrcIP, params.DstPort)
			return err
		}

	case MapTypeSrcPortList:
		//base the map port_list
		err = AddEbpfRuleForSrcDestPortList(params.SrcIP, params.DstPortStart, params.DstPortEnd, TtlSec64)
		if err != nil {
			log.Error("failed add ebpf src: %s dst port start: %d dst port end: %d", params.SrcIP, params.DstPortStart, params.DstPortEnd)
			return err
		}

	case MapTypeProtocolPort:
		//base the map protocol_port
		dstPort := uint16(params.DstPort)
		err = AddEbpfRuleForProtocolPort(protocol, dstPort, TtlSec64)
		if err != nil {
			log.Error("failed add ebpf protocol: %s dst port: %d", params.Protocol, params.DstPort)
			return err
		}

	default:
		return fmt.Errorf("unsupported map type: %d", mapType)
	}

	return nil
}

// function for update port_list map
func WhitelistPortListRangeRule(whitelistMap *ebpf.Map, srcIP uint32, dstPortStart, dstPortEnd int, ttlSec uint64) error {
	now, err := ebpflocal.GetBootTimeNanos()
	if err != nil {
		return err
	}
	portStart, err := safeIntToUint16(dstPortStart)
	if err != nil {
		log.Error("failed to safeIntToUint16 for dstPortStart: %d", dstPortStart)
		return err
	}
	portEnd, err := safeIntToUint16(dstPortEnd)

	if err != nil {
		log.Error("failed to safeIntToUint16 for dstPortEnd: %d", dstPortEnd)
		return err
	}

	key := portListKey{
		SrcIP:        srcIP,
		DstPortStart: portStart,
		DstPortEnd:   portEnd,
	}

	value := WhitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}
	keyBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(keyBytes[0:4], key.SrcIP)
	binary.LittleEndian.PutUint16(keyBytes[4:6], key.DstPortStart)
	binary.LittleEndian.PutUint16(keyBytes[6:8], key.DstPortEnd)

	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		log.Error("failed to update port_list map: %v", err)
		return err
	}

	return nil
}

func AddEbpfRuleForSrcDestPortList(srcIPStr string, dstPortStart, dstPortEnd int, ttlSec uint64) error {
	portListMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/port_list", nil)
	if err != nil {
		log.Error("failed to load pinned port_list map: %v", err)
		return err
	}
	defer portListMap.Close()

	srcIP, err := parseIP(srcIPStr)
	if err != nil {
		log.Error("invalid source IP: %v", err)
		return err
	}

	err = WhitelistPortListRangeRule(portListMap, srcIP, dstPortStart, dstPortEnd, ttlSec)
	if err != nil {
		log.Error("failed to update port_list map: %v", err)
		return err
	}

	return nil
}

// function for update protocol_port map
func WhitelistProtocolPortRule(whitelistMap *ebpf.Map, dstPort uint16, protocol uint8, ttlSec uint64) error {
	now, err := ebpflocal.GetBootTimeNanos()
	if err != nil {
		return err
	}

	key := procoPortKey{
		DstPort:  dstPort,
		Protocol: protocol,
	}
	value := procoPortValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9, // ​​Convert to nanoseconds
	}

	keyBytes := make([]byte, 3)

	binary.LittleEndian.PutUint16(keyBytes[0:2], key.DstPort)
	keyBytes[2] = key.Protocol

	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		log.Error("failed to update protocol_port map: %v", err)
		return err
	}

	return nil
}

func AddEbpfRuleForProtocolPort(protocol uint8, dstPort uint16, ttlSec uint64) error {
	portStr := fmt.Sprintf("%d", dstPort)
	dstPortt, err := parsePort(portStr)
	if err != nil {
		log.Error("failed to parsePort: %v", portStr)
		return err
	}
	portListMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/protocol_port", nil)
	if err != nil {
		log.Error("failed to load pinned protocol_port map: %v", err)
		return err
	}
	defer portListMap.Close()

	err = WhitelistProtocolPortRule(portListMap, dstPortt, protocol, ttlSec)
	if err != nil {
		log.Error("failed to update protocol_port map: %v", err)
		return err
	}

	return nil
}
