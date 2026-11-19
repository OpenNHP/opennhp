package ebpf

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"

	"github.com/cilium/ebpf"

	"github.com/OpenNHP/opennhp/nhp/log"
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

// The port fields of the map keys below are network byte order on the kernel
// side: nhp/ebpf/xdp/nhp_maps.h declares them __be16, and the programs copy
// them straight out of the packet (ct_key.dport = tcp->dest), so the bytes
// the kernel looks up are the bytes on the wire. User space therefore has to
// write the port with binary.BigEndian -- a host-order port would be looked
// up as its byte swap, and the rule would admit bswap(port) instead of port.
//
// SrcIP is the other way round and deliberately so: it is the raw address
// bytes read as a little-endian integer (parseIP) and written back as one
// (binary.LittleEndian.PutUint32), and the two cancel out, so the bytes in
// the map are the address bytes. Do not "fix" that one -- it is already
// right, and the tests in ebpf_test.go pin it.
//
// The `src_port` and `port_list` keys carry the protocol as well, and that is
// the last byte of both: the address, the port(s), then the IP protocol
// number (6 for TCP, 17 for UDP). The programs fill it from the packet
// (iph->protocol), so an entry is only hit by a packet of the protocol it
// was written for. Without it, a resource that names one protocol would open
// the same port for every protocol -- wider than the ipset mode, whose set
// entries carry the protocol themselves ("net,443" is TCP, "net,udp:443" is
// UDP).

type portListKey struct {
	SrcIP        uint32 `ebpf:"src_ip"`
	DstPortStart uint16 `ebpf:"dst_port_start"`
	DstPortEnd   uint16 `ebpf:"dst_port_end"`
	Protocol     uint8  `ebpf:"protocol"`
}

type srcIPdstPortKey struct {
	SrcIP    uint32 `ebpf:"src_ip"`
	DstPort  uint16 `ebpf:"dst_port"`
	Protocol uint8  `ebpf:"protocol"`
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

// The IP protocol numbers, as they appear in the IP header and therefore in
// the protocol field of a map key.
const (
	ProtocolICMP = 1
	ProtocolTCP  = 6
	ProtocolUDP  = 17
)

// portScopedProtocols returns the IP protocol numbers a port-scoped grant
// covers.
//
// A port-scoped grant ("this peer may use this port") is only meaningful for
// the protocols that have ports, and it has to be written for one protocol at
// a time: the key carries the protocol, and the program looks it up with the
// protocol of the packet it is judging.
//
// The names are the ones a resource carries ("tcp", "udp"), and the empty
// name and "any" both mean "every protocol that has a port" -- TCP and UDP.
// That is what the ipset mode does with the same resource: an entry with no
// protocol named is written once for TCP ("net,443") and once for UDP
// ("net,udp:443"), and the two together are the whole of it.
//
// ICMP is not in that set: it has no port, so a port-scoped entry for it
// could never be looked up. ICMP is granted through the icmpwhitelist map,
// which is keyed on the peer and the target alone.
func portScopedProtocols(name string) ([]uint8, error) {
	switch name {
	case "", "any":
		return []uint8{ProtocolTCP, ProtocolUDP}, nil
	case "tcp":
		return []uint8{ProtocolTCP}, nil
	case "udp":
		return []uint8{ProtocolUDP}, nil
	default:
		return nil, fmt.Errorf("unsupported protocol for a port-scoped grant: %s "+
			"(the port-scoped maps are keyed on the protocol and only TCP and UDP have ports; "+
			"ICMP is granted through the icmpwhitelist map)", name)
	}
}

// validateProtocol rejects a protocol name this package does not know.
//
// It runs for every shape, including the two whose key carries no protocol:
// a resource with a protocol name nobody understands is a configuration
// error, and it has to be an error rather than a grant that silently means
// something else.
func validateProtocol(name string) error {
	switch name {
	case "", "any", "tcp", "udp", "icmp":
		return nil
	default:
		return fmt.Errorf("unsupported protocol: %s", name)
	}
}

// singleProtocol returns the IP protocol number a single-protocol grant
// covers.
//
// It is used for the two shapes whose key names exactly one protocol: the
// `spp` key and the `protocol_port` key. The empty name is the "no protocol
// named" case, which those two shapes cannot express -- a key with protocol 0
// is never looked up -- so it is kept as it was rather than silently turned
// into a grant of some other protocol.
func singleProtocol(name string) (uint8, error) {
	switch name {
	case "":
		return 0, nil
	case "tcp":
		return ProtocolTCP, nil
	case "udp":
		return ProtocolUDP, nil
	case "icmp":
		return ProtocolICMP, nil
	default:
		return 0, fmt.Errorf("unsupported protocol: %s", name)
	}
}

// AnyPortRangeStart and AnyPortRangeEnd are the only port range the eBPF
// programs ever look up in the port_list map.
//
// The map's key is {src_ip, min_port, max_port}, but the programs do not look
// up "the range that was installed": they look up the single key
// {src_ip, MIN_PORT, MAX_PORT} -- MIN_PORT/MAX_PORT in nhp_ebpf_xdp.c, and
// the same two numbers written out in tc_egress.c. The range in the key is
// therefore not a range the program evaluates; it is a fixed sentinel.
//
// A range other than this one is an entry the programs can never find, so
// installing it is a silent loss of access rather than a narrower access:
// AddEbpfRuleForSrcDestPortList refuses it instead of writing it.
const (
	AnyPortRangeStart = 0
	AnyPortRangeEnd   = 65535
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
	binary.BigEndian.PutUint16(keyBytes[8:10], r.DstPort)
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
	binary.BigEndian.PutUint16(keyBytes[0:2], r.DstPort)
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
	keyBytes := make([]byte, 9)
	binary.LittleEndian.PutUint32(keyBytes[0:4], r.SrcIP)
	binary.BigEndian.PutUint16(keyBytes[4:6], r.DstPortStart)
	binary.BigEndian.PutUint16(keyBytes[6:8], r.DstPortEnd)
	keyBytes[8] = r.Protocol
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
	keyBytes := make([]byte, 7)
	binary.LittleEndian.PutUint32(keyBytes[0:4], r.SrcIP)
	binary.BigEndian.PutUint16(keyBytes[4:6], r.DstPort)
	keyBytes[6] = r.Protocol
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

// The map's key carries the protocol, so one entry is written per protocol
// the grant covers: the program looks the key up with the protocol of the
// packet it is judging, and an entry written for TCP is not found by a UDP
// packet.
func AddEbpfRuleForSrcDestPort(srcIPStr string, protocols []uint8, dstPort int, ttlSec uint64) error {
	whitelistMap, err := ebpf.LoadPinnedMap("/sys/fs/bpf/src_port", nil)
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

	for _, protocol := range protocols {
		rule := &srcIPdstPortKey{
			SrcIP:    srcIP,
			DstPort:  dstPortu,
			Protocol: protocol,
		}
		if err := AddSrcipDestPortRule(whitelistMap, rule, ttlSec); err != nil {
			return err
		}
	}
	return nil
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
//
// The map's key carries the protocol, so one entry is written per protocol
// the grant covers: the program looks the key up with the protocol of the
// packet it is judging, and an entry written for TCP is not found by a UDP
// packet.
func AddEbpfRuleForSrcDestPortList(srcIPStr string, protocols []uint8, dstPortStart, dstPortEnd int, ttlSec uint64) error {
	// The range is checked before anything is written, and before the map is
	// even opened, so that a refusal is the same answer whether or not the
	// pin exists.
	//
	// The eBPF programs look up exactly one key in this map: {src_ip, 0,
	// 65535} (MIN_PORT/MAX_PORT in nhp_ebpf_xdp.c, the same two numbers in
	// tc_egress.c). They do not evaluate the range in the key -- they look up
	// the sentinel. An entry with any other range is therefore an entry no
	// lookup can ever find, and installing it would look like an access was
	// granted while the peer is in fact still locked out. Refusing it is the
	// honest answer.
	if dstPortStart != AnyPortRangeStart || dstPortEnd != AnyPortRangeEnd {
		err := fmt.Errorf("port_list only supports the wildcard port range %d-%d: "+
			"the eBPF programs look up that one key (MIN_PORT/MAX_PORT) and no other, "+
			"so an entry with the range %d-%d could never be matched and would grant no access at all",
			AnyPortRangeStart, AnyPortRangeEnd, dstPortStart, dstPortEnd)
		log.Error("refusing to add port_list entry: %v", err)
		return err
	}

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

	for _, protocol := range protocols {
		rule := &portListKey{
			SrcIP:        srcIP,
			DstPortStart: portStart,
			DstPortEnd:   portEnd,
			Protocol:     protocol,
		}
		if err := AddSdPortlistRule(portListMap, rule, ttlSec); err != nil {
			return err
		}
	}

	return nil
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
//
// The protocol of the resource decides which key the grant is written under.
// The two port-scoped shapes (src_port and port_list) carry the protocol in
// their key, so the program only finds an entry with a packet of the protocol
// it was written for: a knock for tcp/443 does not open udp/443. A resource
// that names no protocol covers TCP and UDP, which is what the ipset mode
// does with the same resource.
func EbpfRuleAdd(mapType int, params EbpfRuleParams, TtlSec int) error {
	var err error
	TtlSec64 := uint64(TtlSec)

	if err := validateProtocol(params.Protocol); err != nil {
		return err
	}

	switch mapType {
	case MapTypeWhitelist:
		//base the map whitelist
		protocol, protoErr := singleProtocol(params.Protocol)
		if protoErr != nil {
			return protoErr
		}
		err = AddEbpfRuleForSrcDstPortProto(params.SrcIP, params.DstIP, protocol, uint16(params.DstPort), TtlSec64)
		if err != nil {
			log.Error("failed add ebpf src: %s dst: %s, error: %v, protocol: %d, dstport: %d", params.SrcIP, params.DstIP, err, protocol, uint16(params.DstPort))
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
		protocols, protoErr := portScopedProtocols(params.Protocol)
		if protoErr != nil {
			return protoErr
		}
		err = AddEbpfRuleForSrcDestPort(params.SrcIP, protocols, params.DstPort, TtlSec64)
		if err != nil {
			log.Error("failed add ebpf src: %s dst port: %d, protocol: %v", params.SrcIP, params.DstPort, params.Protocol)
			return err
		}

	case MapTypeSrcPortList:
		//base the map port_list
		protocols, protoErr := portScopedProtocols(params.Protocol)
		if protoErr != nil {
			return protoErr
		}
		err = AddEbpfRuleForSrcDestPortList(params.SrcIP, protocols, params.DstPortStart, params.DstPortEnd, TtlSec64)
		if err != nil {
			log.Error("failed add ebpf src: %s dst port start: %d dst port end: %d, protocol: %v", params.SrcIP, params.DstPortStart, params.DstPortEnd, params.Protocol)
			return err
		}

	case MapTypeProtocolPort:
		//base the map protocol_port
		protocol, protoErr := singleProtocol(params.Protocol)
		if protoErr != nil {
			return protoErr
		}
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
//
// The port is returned as the host-order port number, the same form the other
// key encoders take: the map key is written with binary.BigEndian, so the
// bytes that reach the map are the network-order bytes the kernel looks up.
func parsePort(portStr string) (uint16, error) {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(port), nil
}
