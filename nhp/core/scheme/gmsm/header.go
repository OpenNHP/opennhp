package gmsm

import (
	"encoding/binary"
	"unsafe"

	"github.com/OpenNHP/opennhp/nhp/common"
	utils "github.com/OpenNHP/opennhp/nhp/utils"
)

const (
	HeaderCommonSize    = 24
	HashSize            = 32
	GCMNonceSize        = 12
	GCMTagSize          = 16
	TimestampSize       = 8
	MaximumIdentitySize = 64
	HeaderSize          = HeaderCommonSize + PublicKeySize + MaximumIdentitySize + GCMTagSize + PublicKeySize + GCMTagSize + TimestampSize + GCMTagSize + HashSize
)

type HeaderGmsm struct {
	HeaderCommon [HeaderCommonSize]byte
	Ephermeral   [PublicKeySize]byte
	Identity     [MaximumIdentitySize + GCMTagSize]byte
	Static       [PublicKeySize + GCMTagSize]byte
	Timestamp    [TimestampSize + GCMTagSize]byte
	HMAC         [HashSize]byte
}

// gmsm header implementations
func (h *HeaderGmsm) TypeAndPayloadSize() (t int, s int) {
	preamble := binary.BigEndian.Uint32(h.HeaderCommon[0:4])
	tns := preamble ^ binary.BigEndian.Uint32(h.HeaderCommon[4:8])
	t = int((tns & 0xFFFF0000) >> 16)
	s = int(tns & 0x0000FFFF)
	return t, s
}

func (h *HeaderGmsm) SetTypeAndPayloadSize(t int, s int) {
	preamble := utils.GetRandomUint32()
	t32 := uint32((t & 0x0000FFFF) << 16)
	s32 := uint32(s & 0x0000FFFF)
	tns := preamble ^ (s32 | t32)
	binary.BigEndian.PutUint32(h.HeaderCommon[0:4], preamble)
	binary.BigEndian.PutUint32(h.HeaderCommon[4:8], tns)
}

func (h *HeaderGmsm) Size() int {
	return HeaderSize
}

func (h *HeaderGmsm) Version() (int, int) {
	major := h.HeaderCommon[8]
	minor := h.HeaderCommon[9]
	return int(major), int(minor)
}

func (h *HeaderGmsm) SetVersion(major int, minor int) {
	h.HeaderCommon[8] = uint8(major)
	h.HeaderCommon[9] = uint8(minor)
}

func (h *HeaderGmsm) Flag() uint16 {
	return binary.BigEndian.Uint16(h.HeaderCommon[10:12])
}

func (h *HeaderGmsm) SetFlag(flag uint16) {
	flag |= uint16(common.NHP_FLAG_EXTENDEDLENGTH)
	flag &= 0x0FFF
	flag |= common.NHP_FLAG_SCHEME_GMSM << 12
	binary.BigEndian.PutUint16(h.HeaderCommon[10:12], flag)
}

func (h *HeaderGmsm) NonceBytes() []byte {
	var nonce [GCMNonceSize]byte
	copy(nonce[4:GCMNonceSize], h.HeaderCommon[16:24])
	return nonce[:]
}

func (h *HeaderGmsm) SetCounter(counter uint64) {
	binary.BigEndian.PutUint64(h.HeaderCommon[16:24], counter)
}

func (h *HeaderGmsm) Counter() uint64 {
	return binary.BigEndian.Uint64(h.HeaderCommon[16:24])
}

func (h *HeaderGmsm) Bytes() []byte {
	pHeader := (*[HeaderSize]byte)(unsafe.Pointer(&h.HeaderCommon))
	return pHeader[:]
}

func (h *HeaderGmsm) EphermeralBytes() []byte {
	return h.Ephermeral[:]
}

func (h *HeaderGmsm) StaticBytes() []byte {
	return h.Static[:]
}

func (h *HeaderGmsm) TimestampBytes() []byte {
	return h.Timestamp[:]
}

func (h *HeaderGmsm) IdentityBytes() []byte {
	return h.Identity[:]
}

func (h *HeaderGmsm) HMACBytes() []byte {
	return h.HMAC[:]
}

func (h *HeaderGmsm) CipherScheme() int {
	return common.CIPHER_SCHEME_GMSM
}