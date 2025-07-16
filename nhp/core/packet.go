package core

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core/scheme/curve"
	"github.com/OpenNHP/opennhp/nhp/core/scheme/gmsm"
	log "github.com/OpenNHP/opennhp/nhp/log"
	utils "github.com/OpenNHP/opennhp/nhp/utils"
)

const (
	NHP_KPL = iota // general keepalive packet
	NHP_KNK        // agent sends knock to server
	NHP_ACK        // server replies knock status to agent
	NHP_AOP        // server asks ac for operation
	NHP_ART        // ac replies server for operation result
	NHP_LST        // agent requests server for listing services and applications
	NHP_LRT        // server replies to agent with services and applications result
	NHP_COK        // server sends cookie to agent
	NHP_RKN        // agent sends reknock to server
	NHP_RLY        // relay sends relayed packet to server
	NHP_AOL        // ac sends online status to server
	NHP_AAK        // server sends ack to ac after receving ac's online status
	NHP_OTP        // agent requests server for one-time-password
	NHP_REG        // agent asks server for registering
	NHP_RAK        // server sends back ack when agent registers correctly
	NHP_ACC        // agent sends to ac/resource for actual ip access
	NHP_EXT        // agent requests immediate disconnection
	//DHP
	NHP_DRG //DB sends a message to register a data object file to the NHP Server
	NHP_DAK //NHP-Server sends a result of the NHP_DRG registration request to the DB.
	NHP_DAR //NHP Agent sends messages to get access to the file and then work with it.
	NHP_DAG //The NHP Server sends  the authorization status of the data object to NHP Agent.
	NHP_DSA //The NHP Server sends a self attestation requiestr to the NHP Agent
	NHP_DAV //The NHP Agent sends the attestation proof to the NHP Server.
	NHP_DWR // The NHP Server sends a request to the NHP DB to get the wrapping of the data private key
	NHP_DWA //The NHP DB sends the data private key to the NHP Server
	NHP_DOL //DB sends online status to server
	NHP_DBA //server send ack to db after receiving db's online status
)

var nhpHeaderTypeStrings []string = []string{
	"NHP-KPL", // general keepalive packet
	"NHP-KNK", // agent sends knock to server
	"NHP-ACK", // server replies knock status to agent
	"NHP-AOP", // server asks ac for operation
	"NHP-ART", // ac replies server for operation result
	"NHP-LST", // agent requests server for listing services and applications
	"NHP-LRT", // server replies to agent with services and applications result
	"NHP-COK", // server sends cookie to agent
	"NHP-RKN", // agent sends reknock to server
	"NHP-RLY", // relay sends relayed packet to server
	"NHP-AOL", // ac sends online status to server
	"NHP-AAK", // server sends ack to ac after receving ac's online status
	"NHP-OTP", // agent requests server for one-time-password
	"NHP-REG", // agent asks server for registering
	"NHP-RAK", // server sends back ack when agent registers correctly
	"NHP-ACC", // agent sends to ac/resource for actual ip access
	"NHP-EXT", // agent requests immediate disconnection
	"NHP_DRG", //DB sends a message to register a data object file to the NHP Server
	"NHP_DAK", //NHP-Server sends a result of the NHP_DRG registration request to the DB.
	"NHP_DAR", //NHP Agent sends messages to get access to the file and then work with it.
	"NHP_DAG", //The NHP Server sends  the authorization status of the data object to NHP Agent.
	"NHP_DSA", //The NHP Server sends a self attestation request to the NHP Agent
	"NHP_DAV", //The NHP Agent sends the attestation proof to the NHP Server.
	"NHP_DWR", //The NHP Server sends a request to the NHP DB to get the wrapping of the data private key
	"NHP_DWA", //The NHP DB sends the data private key to the NHP Server
	"NHP_DOL", //DB sends online status to server
	"NHP_DBA", //server send ack to db after receiving db's online status
}

func HeaderTypeToString(t int) string {
	if t < len(nhpHeaderTypeStrings) {
		return nhpHeaderTypeStrings[t]
	}
	return "UNKNOWN"
}

func HeaderTypeToDeviceType(t int) int {
	switch t {
	case NHP_KNK, NHP_LST, NHP_RKN, NHP_OTP, NHP_REG, NHP_ACC, NHP_EXT, NHP_DAR, NHP_DAV:
		return NHP_AGENT
	case NHP_ACK, NHP_AOP, NHP_LRT, NHP_COK, NHP_AAK, NHP_RAK, NHP_DAK, NHP_DAG, NHP_DBA, NHP_DWR, NHP_DSA:
		return NHP_SERVER

	case NHP_AOL, NHP_ART:
		return NHP_AC

	case NHP_RLY:
		return NHP_RELAY
	case NHP_DRG, NHP_DOL, NHP_DWA:
		return NHP_DB
	}

	return NHP_NO_DEVICE
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

type Packet struct {
	Buf           *PacketBuffer
	HeaderType    int
	PoolAllocated bool
	KeepAfterSend bool // only applicable for sending
	Content       []byte
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
	IdentityBytes() []byte
	HMACBytes() []byte
	CipherScheme() int
}

func (pkt *Packet) Flag() uint16 {
	return binary.BigEndian.Uint16(pkt.Content[10:12])
}

func (pkt *Packet) Header() Header {
	if pkt.Flag() & common.NHP_FLAG_EXTENDEDLENGTH == 0 {
		return (*curve.HeaderCurve)(unsafe.Pointer(&pkt.Content[0]))
	} else {
		switch pkt.Flag() & (0xF << 12) {
		case common.NHP_FLAG_SCHEME_GMSM:
			fallthrough
		default:
			return (*gmsm.HeaderGmsm)(unsafe.Pointer(&pkt.Content[0]))
		}
	}
}

func (pkt *Packet) HeaderWithCipherScheme(cipherScheme int) Header {
	switch cipherScheme {
	case common.CIPHER_SCHEME_CURVE:
		return (*curve.HeaderCurve)(unsafe.Pointer(&pkt.Content[0]))
	case common.CIPHER_SCHEME_GMSM:
		fallthrough
	default:
		return (*gmsm.HeaderGmsm)(unsafe.Pointer(&pkt.Content[0]))
	}
}

func (pkt *Packet) HeaderTypeAndSize() (t int, s int) {
	preamble := binary.BigEndian.Uint32(pkt.Content[0:4])
	tns := preamble ^ binary.BigEndian.Uint32(pkt.Content[4:8])
	t = int((tns & 0xFFFF0000) >> 16)
	s = int(tns & 0x0000FFFF)
	pkt.HeaderType = t

	return t, s
}

func (pkt *Packet) Counter() uint64 {
	return binary.BigEndian.Uint64(pkt.Content[16:24])
}

func (pkt *Packet) MinimalLength() int {
	return pkt.HeaderWithCipherScheme(common.CIPHER_SCHEME_CURVE).Size()
}

// Data Receiver  allowed message types
func (d *Device) CheckRecvHeaderType(t int) bool {
	// NHP_KPL is handled elsewhere
	switch d.deviceType {
	case NHP_AGENT:
		switch t {
		case NHP_ACK, NHP_LRT, NHP_COK, NHP_RAK, NHP_DAG, NHP_DSA:
			return true
		}
	case NHP_SERVER:
		switch t {
		case NHP_REG, NHP_KNK, NHP_LST, NHP_RKN, NHP_EXT, NHP_ART, NHP_RLY, NHP_AOL, NHP_OTP, NHP_DRG, NHP_DAR, NHP_DAV, NHP_DOL, NHP_DWA:
			return true
		}
	case NHP_AC:
		switch t {
		case NHP_AOP, NHP_LRT, NHP_AAK:
			return true
		}
	case NHP_RELAY:
		switch t {
		case NHP_REG, NHP_KNK, NHP_ACK, NHP_LST, NHP_LRT, NHP_COK, NHP_RKN, NHP_EXT:
			return true
		}

	case NHP_DB:
		switch t {
		case NHP_DRG, NHP_DAG, NHP_DAK, NHP_DBA, NHP_DWR:
			return true
		}
	}
	log.Info("Device type: %d, recv header type %d not allowed", d.deviceType, t)
	return false
}

func (d *Device) RecvPrecheck(pkt *Packet) (int, int, error) {
	headerSize := pkt.Header().Size()

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

	totalLen := len(pkt.Content)
	if totalLen != headerSize+s {
		return t, s, fmt.Errorf("packet total size is incorrect")
	}

	return t, s, nil
}

func (d *Device) AllocatePoolPacket() *Packet {
	buf := d.pool.Get()
	return &Packet{Buf: buf, Content: buf[:], PoolAllocated: true}
}

func (d *Device) ReleasePoolPacket(pkt *Packet) {
	if pkt != nil && pkt.Buf != nil && pkt.PoolAllocated {
		d.pool.Put(pkt.Buf)
		pkt.Buf = nil
		pkt.Content = nil
	}
}
