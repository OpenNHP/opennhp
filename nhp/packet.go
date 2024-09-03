package nhp

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/OpenNHP/opennhp/utils"
)

type HeaderTypeEnum int

const (
	NHP_KPL = iota // general keepalive packet
	NHP_KNK        // agent sends knock to server
	NHP_ACK        // server replies knock status to agent
	NHP_AOP        // server asks door for operation
	NHP_ART        // door replies server for operation result
	NHP_LST        // agent requests server for listing services and applications
	NHP_LRT        // server replies to agent with services and applications result
	NHP_COK        // server sends cookie to agent
	NHP_RKN        // agent sends reknock to server
	NHP_RLY        // relay sends relayed packet to server
	NHP_AOL        // door sends online status to server
	NHP_AAK        // server sends ack to door after receving door's online status
	NHP_OTP        // agent requests server for one-time-password
	NHP_REG        // agent asks server for registering
	NHP_RAK        // server sends back ack when agent registers correctly
	NHP_ACC        // agent sends to door/resource for actual ip access
	NHP_EXT        // agent requests immediate disconnection
)

var nhpHeaderTypeStrings []string = []string{
	"NHP-KPL", // general keepalive packet
	"NHP-KNK", // agent sends knock to server
	"NHP-ACK", // server replies knock status to agent
	"NHP-AOP", // server asks door for operation
	"NHP-ART", // door replies server for operation result
	"NHP-LST", // agent requests server for listing services and applications
	"NHP-LRT", // server replies to agent with services and applications result
	"NHP-COK", // server sends cookie to agent
	"NHP-RKN", // agent sends reknock to server
	"NHP-RLY", // relay sends relayed packet to server
	"NHP-AOL", // door sends online status to server
	"NHP-AAK", // server sends ack to door after receving door's online status
	"NHP-OTP", // agent requests server for one-time-password
	"NHP-REG", // agent asks server for registering
	"NHP-RAK", // server sends back ack when agent registers correctly
	"NHP-ACC", // agent sends to door/resource for actual ip access
	"NHP_EXT", // agent requests immediate disconnection
}

func HeaderTypeToString(t int) string {
	if t < len(nhpHeaderTypeStrings) {
		return nhpHeaderTypeStrings[t]
	}
	return "UNKNOWN"
}

func HeaderTypeToDeviceType(t int) int {
	switch t {
	case NHP_KNK, NHP_LST, NHP_RKN, NHP_OTP, NHP_REG, NHP_ACC, NHP_EXT:
		return NHP_AGENT

	case NHP_ACK, NHP_AOP, NHP_LRT, NHP_COK, NHP_AAK, NHP_RAK:
		return NHP_SERVER

	case NHP_AOL, NHP_ART:
		return NHP_AC

	case NHP_RLY:
		return NHP_RELAY
	}

	return NHP_NO_DEVICE
}

type HeaderFlagEnum int

const (
	NHP_FLAG_EXTENDEDLENGTH = 1 << iota
	NHP_FLAG_COMPRESS
)

type NHPHeaderCommon struct {
	Preamble uint32
	Type     uint16
	Size     uint16
	VerMajor uint8
	VerMinor uint8
	Flag     uint16
	Reserved uint32
	Counter  uint64
}

type NHPHeader struct {
	HeaderCommon [HeaderCommonSize]byte
	Ephermeral   [PublicKeySize]byte
	Static       [PublicKeySize + GCMTagSize]byte
	Timestamp    [TimestampSize + GCMTagSize]byte
	HMAC         [HashSize]byte
}

type NHPHeaderEx struct {
	HeaderCommon [HeaderCommonSize]byte
	Ephermeral   [PublicKeySizeEx]byte
	Static       [PublicKeySizeEx + GCMTagSize]byte
	Timestamp    [TimestampSize + GCMTagSize]byte
	HMAC         [HashSize]byte
}

type Header interface {
	SetTypeAndPayloadSize(int, int)
	TypeAndPayloadSize() (int, int)
	Size() int
	SetVersion(int, int)
	Version() (int, int)
	SetFlag(uint16)
	Flag() uint16
	SetCounter(uint64)
	Counter() uint64
	Bytes() []byte
	NonceBytes() []byte
	EphermeralBytes() []byte
	StaticBytes() []byte
	TimestampBytes() []byte
	HMACBytes() []byte
}

// NHPHeader implementations
func (h *NHPHeader) TypeAndPayloadSize() (t int, s int) {
	preamble := binary.BigEndian.Uint32(h.HeaderCommon[0:4])
	tns := preamble ^ binary.BigEndian.Uint32(h.HeaderCommon[4:8])
	t = int((tns & 0xFFFF0000) >> 16)
	s = int(tns & 0x0000FFFF)
	return t, s
}

func (h *NHPHeader) SetTypeAndPayloadSize(t int, s int) {
	preamble := utils.GetRandomUint32()
	t32 := uint32((t & 0x0000FFFF) << 16)
	s32 := uint32(s & 0x0000FFFF)
	tns := preamble ^ (s32 | t32)
	binary.BigEndian.PutUint32(h.HeaderCommon[0:4], preamble)
	binary.BigEndian.PutUint32(h.HeaderCommon[4:8], tns)
}

func (h *NHPHeader) Size() int {
	return HeaderSize
}

func (h *NHPHeader) Version() (int, int) {
	major := h.HeaderCommon[8]
	minor := h.HeaderCommon[9]
	return int(major), int(minor)
}

func (h *NHPHeader) SetVersion(major int, minor int) {
	h.HeaderCommon[8] = uint8(major)
	h.HeaderCommon[9] = uint8(minor)
}

func (h *NHPHeader) Flag() uint16 {
	return binary.BigEndian.Uint16(h.HeaderCommon[10:12])
}

func (h *NHPHeader) SetFlag(flag uint16) {
	flag &= ^uint16(NHP_FLAG_EXTENDEDLENGTH)
	binary.BigEndian.PutUint16(h.HeaderCommon[10:12], flag)
}

func (h *NHPHeader) NonceBytes() []byte {
	var nonce [GCMNonceSize]byte
	copy(nonce[4:GCMNonceSize], h.HeaderCommon[16:24])
	return nonce[:]
}

func (h *NHPHeader) SetCounter(counter uint64) {
	binary.BigEndian.PutUint64(h.HeaderCommon[16:24], counter)
}

func (h *NHPHeader) Counter() uint64 {
	return binary.BigEndian.Uint64(h.HeaderCommon[16:24])
}

func (h *NHPHeader) Bytes() []byte {
	pHeader := (*[HeaderSize]byte)(unsafe.Pointer(&h.HeaderCommon))
	return pHeader[:]
}

func (h *NHPHeader) EphermeralBytes() []byte {
	return h.Ephermeral[:]
}

func (h *NHPHeader) StaticBytes() []byte {
	return h.Static[:]
}

func (h *NHPHeader) TimestampBytes() []byte {
	return h.Timestamp[:]
}

func (h *NHPHeader) HMACBytes() []byte {
	return h.HMAC[:]
}

// NHPHeaderEx implementations
func (h *NHPHeaderEx) TypeAndPayloadSize() (t int, s int) {
	preamble := binary.BigEndian.Uint32(h.HeaderCommon[0:4])
	tns := preamble ^ binary.BigEndian.Uint32(h.HeaderCommon[4:8])
	t = int((tns & 0xFFFF0000) >> 16)
	s = int(tns & 0x0000FFFF)
	return t, s
}

func (h *NHPHeaderEx) SetTypeAndPayloadSize(t int, s int) {
	preamble := utils.GetRandomUint32()
	t32 := uint32((t & 0x0000FFFF) << 16)
	s32 := uint32(s & 0x0000FFFF)
	tns := preamble ^ (s32 | t32)
	binary.BigEndian.PutUint32(h.HeaderCommon[0:4], preamble)
	binary.BigEndian.PutUint32(h.HeaderCommon[4:8], tns)
}

func (h *NHPHeaderEx) Size() int {
	return HeaderSizeEx
}

func (h *NHPHeaderEx) Version() (int, int) {
	major := h.HeaderCommon[8]
	minor := h.HeaderCommon[9]
	return int(major), int(minor)
}

func (h *NHPHeaderEx) SetVersion(major int, minor int) {
	h.HeaderCommon[8] = uint8(major)
	h.HeaderCommon[9] = uint8(minor)
}

func (h *NHPHeaderEx) Flag() uint16 {
	return binary.BigEndian.Uint16(h.HeaderCommon[10:12])
}

func (h *NHPHeaderEx) SetFlag(flag uint16) {
	flag |= uint16(NHP_FLAG_EXTENDEDLENGTH)
	binary.BigEndian.PutUint16(h.HeaderCommon[10:12], flag)
}

func (h *NHPHeaderEx) NonceBytes() []byte {
	var nonce [GCMNonceSize]byte
	copy(nonce[4:GCMNonceSize], h.HeaderCommon[16:24])
	return nonce[:]
}

func (h *NHPHeaderEx) SetCounter(counter uint64) {
	binary.BigEndian.PutUint64(h.HeaderCommon[16:24], counter)
}

func (h *NHPHeaderEx) Counter() uint64 {
	return binary.BigEndian.Uint64(h.HeaderCommon[16:24])
}

func (h *NHPHeaderEx) Bytes() []byte {
	pHeader := (*[HeaderSizeEx]byte)(unsafe.Pointer(&h.HeaderCommon))
	return pHeader[:]
}

func (h *NHPHeaderEx) EphermeralBytes() []byte {
	return h.Ephermeral[:]
}

func (h *NHPHeaderEx) StaticBytes() []byte {
	return h.Static[:]
}

func (h *NHPHeaderEx) TimestampBytes() []byte {
	return h.Timestamp[:]
}

func (h *NHPHeaderEx) HMACBytes() []byte {
	return h.HMAC[:]
}

type PacketBuffer = [PacketBufferSize]byte

// packet buffer pool
type PacketBufferPool struct {
	pool *utils.WaitPool
}

func (bp *PacketBufferPool) Init(max uint32) {
	bp.pool = utils.NewWaitPool(max, func() any { return new(PacketBuffer) })
}

// must be called after Init()
func (bp *PacketBufferPool) Get() *PacketBuffer {
	return bp.pool.Get().(*PacketBuffer)
}

// must be called after Init()
func (bp *PacketBufferPool) Put(packet *PacketBuffer) {
	bp.pool.Put(packet)
}

type UdpPacket struct {
	Buf           *PacketBuffer
	HeaderType    int
	KeepAfterSend bool // only applicable for sending
	Packet        []byte
}

func (pkt *UdpPacket) Flag() uint16 {
	return binary.BigEndian.Uint16(pkt.Packet[10:12])
}

func (pkt *UdpPacket) HeaderTypeAndSize() (t int, s int) {
	preamble := binary.BigEndian.Uint32(pkt.Packet[0:4])
	tns := preamble ^ binary.BigEndian.Uint32(pkt.Packet[4:8])
	t = int((tns & 0xFFFF0000) >> 16)
	s = int(tns & 0x0000FFFF)
	pkt.HeaderType = t

	return t, s
}

func (pkt *UdpPacket) Counter() uint64 {
	return binary.BigEndian.Uint64(pkt.Packet[16:24])
}

func (d *Device) RecvPrecheck(pkt *UdpPacket) (int, int, error) {
	var headerSize int
	flag := pkt.Flag()
	if flag&NHP_FLAG_EXTENDEDLENGTH == 0 {
		headerSize = HeaderSize
	} else {
		headerSize = HeaderSizeEx
	}

	// check type and payload size
	t, s := pkt.HeaderTypeAndSize()
	if t == NHP_KPL {
		if s == 0 {
			return t, s, nil
		} else {
			return t, s, fmt.Errorf("keepalive packet size is incorrect")
		}
	}

	if !d.CheckRecvHeaderType(t) {
		return t, s, fmt.Errorf("packet header type does not match device")
	}

	totalLen := len(pkt.Packet)
	if totalLen != headerSize+s {
		return t, s, fmt.Errorf("packet total size is incorrect")
	}

	return t, s, nil
}

func (d *Device) AllocateUdpPacket() *UdpPacket {
	return &UdpPacket{Buf: d.pool.Get()}
}

func (d *Device) ReleaseUdpPacket(pkt *UdpPacket) {
	if pkt != nil && pkt.Buf != nil {
		d.pool.Put(pkt.Buf)
		pkt.Buf = nil
		pkt.Packet = nil
	}
}
