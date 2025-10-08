package ebpf

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"

	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/cilium/ebpf"
)

type whitelistKey struct {
	SrcIP    uint32 `ebpf:"src_ip"`
	DstIP    uint32 `ebpf:"dst_ip"`
	DstPort  uint16 `ebpf:"dst_port"`
	Protocol uint8  `ebpf:"protocol"`
}

type srcDestKey struct {
	SrcIP uint32 `ebpf:"src_ip"`
	DstIP uint32 `ebpf:"dst_ip"`
}

type portListKey struct {
	SrcIP        uint32 `ebpf:"src_ip"`
	DstPortStart uint16 `ebpf:"dst_port_start"`
	DstPortEnd   uint16 `ebpf:"dst_port_end"`
}

type srcIPdstPortKey struct {
	SrcIP   uint32 `ebpf:"src_ip"`
	DstPort uint16 `ebpf:"dst_port"`
}

type procoPortKey struct {
	DstPort  uint16 `ebpf:"dst_port"`
	Protocol uint8  `ebpf:"protocol"`
}

type whitelistValue struct {
	Allowed    uint8
	_          [7]byte
	ExpireTime uint64
}

type procoPortValue struct {
	Allowed    uint8
	ExpireTime uint64
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

func (r *whitelistKey) ToWlKey() []byte {
	keyBytes := make([]byte, 11)
	binary.LittleEndian.PutUint32(keyBytes[0:4], r.SrcIP)
	binary.LittleEndian.PutUint32(keyBytes[4:8], r.DstIP)
	binary.LittleEndian.PutUint16(keyBytes[8:10], r.DstPort)
	keyBytes[10] = r.Protocol
	return keyBytes
}

func (r *whitelistKey) ToWlValue(ttlSec uint64) whitelistValue {
	now, _ := getBootTimeNanos()
	return whitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}
}

func (r *srcDestKey) ToSdKey() []byte {
	keyBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(keyBytes[0:4], r.SrcIP)
	binary.LittleEndian.PutUint32(keyBytes[4:8], r.DstIP)
	return keyBytes
}

func (r *srcDestKey) ToSdValue(ttlSec uint64) whitelistValue {
	now, _ := getBootTimeNanos()
	return whitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}
}

func (r *procoPortKey) ToPpKey() []byte {
	keyBytes := make([]byte, 3)
	binary.LittleEndian.PutUint16(keyBytes[0:2], r.DstPort)
	keyBytes[2] = r.Protocol
	return keyBytes
}

func (r *procoPortKey) ToPpValue(ttlSec uint64) procoPortValue {
	now, _ := getBootTimeNanos()
	return procoPortValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}
}

func (r *portListKey) ToPlKey() []byte {
	keyBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(keyBytes[0:4], r.SrcIP)
	binary.LittleEndian.PutUint16(keyBytes[4:6], r.DstPortStart)
	binary.LittleEndian.PutUint16(keyBytes[6:8], r.DstPortEnd)
	return keyBytes
}

func (r *portListKey) ToPlValue(ttlSec uint64) whitelistValue {
	now, _ := getBootTimeNanos()
	return whitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}
}

func (r *srcIPdstPortKey) ToSpKey() []byte {
	keyBytes := make([]byte, 6)
	binary.LittleEndian.PutUint32(keyBytes[0:4], r.SrcIP)
	binary.LittleEndian.PutUint16(keyBytes[4:6], r.DstPort)
	return keyBytes
}

func (r *srcIPdstPortKey) ToSpValue(ttlSec uint64) whitelistValue {
	now, _ := getBootTimeNanos()
	return whitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}
}

// function for update whitelist map
func AddWhitelistRule(whitelistMap *ebpf.Map, rule *whitelistKey, ttlSec uint64) error {
	keyBytes := rule.ToWlKey()
	value := rule.ToWlValue(ttlSec)

	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		log.Error("failed to update whitelist map: %v", err)
		return err
	}
	return nil
}

// function for update sdwhitelist map
func AddSdWhitelistRule(whitelistMap *ebpf.Map, rule *srcDestKey, ttlSec uint64) error {
	keyBytes := rule.ToSdKey()
	value := rule.ToSdValue(ttlSec)

	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		log.Error("failed to update sdwhitelist map: %v", err)
		return err
	}
	return nil
}

// function for update sdportlist map
func AddSdPortlistRule(whitelistMap *ebpf.Map, rule *portListKey, ttlSec uint64) error {
	keyBytes := rule.ToPlKey()
	value := rule.ToPlValue(ttlSec)

	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		log.Error("failed to update src dst portlist map: %v", err)
		return err
	}
	return nil
}

// function for update protocol_port map
func AddPpWhitelistRule(whitelistMap *ebpf.Map, rule *procoPortKey, ttlSec uint64) error {
	keyBytes := rule.ToPpKey()
	value := rule.ToPpValue(ttlSec)

	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		log.Error("failed to update sdwhitelist map: %v", err)
		return err
	}
	return nil
}

// function for update src dst port map
func AddSrcipDestPortRule(whitelistMap *ebpf.Map, rule *srcIPdstPortKey, ttlSec uint64) error {
	keyBytes := rule.ToSpKey()
	value := rule.ToSpValue(ttlSec)

	if err := whitelistMap.Update(keyBytes, &value, ebpf.UpdateAny); err != nil {
		log.Error("failed to update src_port_list map: %v", err)
		return err
	}
	return nil
}

func AddEbpfRuleForSrcDstPortProto(srcIPStr, dstIPStr string, protocol uint8, dstPort uint16, ttlSec uint64) error {
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/spp", nil)
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

	rule := &whitelistKey{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		DstPort:  dstPort,
		Protocol: protocol,
	}

	return AddWhitelistRule(whitelistMap, rule, ttlSec)
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

	rule := &srcDestKey{
		SrcIP: srcIP,
		DstIP: dstIP,
	}

	return AddSdWhitelistRule(whitelistMap, rule, ttlSec)
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

	dstPortu, err := safeIntToUint16(dstPort)

	if err != nil {
		log.Error("failed to safeIntToUint16 in src_port_list map: %v", err)
		return err
	}

	rule := &srcIPdstPortKey{
		SrcIP:   srcIP,
		DstPort: dstPortu,
	}
	return AddSrcipDestPortRule(whitelistMap, rule, ttlSec)
}

// function for update icmpwhitelist map
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

	rule := &srcDestKey{
		SrcIP: srcIP,
		DstIP: dstIP,
	}

	return AddSdWhitelistRule(whitelistMap, rule, ttlSec)
}

// function for update port_list map
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

	rule := &portListKey{
		SrcIP:        srcIP,
		DstPortStart: portStart,
		DstPortEnd:   portEnd,
	}

	return AddSdPortlistRule(portListMap, rule, ttlSec)
}

// function for update protocol dstport map
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

	rule := &procoPortKey{
		DstPort:  dstPortt,
		Protocol: protocol,
	}

	return AddPpWhitelistRule(portListMap, rule, ttlSec)
}

func safeIntToUint16(i int) (uint16, error) {
	if i < 0 || i > 65535 {
		return 0, fmt.Errorf("value %d is out of range for uint16", i)
	}
	return uint16(i), nil
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
