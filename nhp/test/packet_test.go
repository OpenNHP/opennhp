package test

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"testing"
	"unsafe"

	core "github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/core/scheme/curve"
	"github.com/OpenNHP/opennhp/nhp/core/scheme/gmsm"
)

func TestHeaderTypeAndSize(t *testing.T) {
	pkt := &core.Packet{
		//Content: []byte{0x18, 0xca, 0xba, 0xa6, 0x18, 0xcb, 0xba, 0xa4},
		Content: []byte{91, 89, 55, 86, 91, 88, 55, 25},
	}

	tp, sz := pkt.HeaderTypeAndSize()

	fmt.Printf("Header type: %d, payload size: %d", tp, sz)
}

func TestHMAC(t *testing.T) {
	buf := []byte{95, 205, 121, 55, 95, 204, 121, 100, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 115, 222, 224, 148, 163, 67, 89, 253, 194, 76, 203, 140, 48, 126, 226, 192, 141, 20, 34, 116, 14, 189, 159, 83, 219, 82, 88, 229, 146, 21, 229, 223, 208, 164, 174, 217, 154, 219, 90, 190, 140, 27, 9, 52, 89, 155, 197, 189, 145, 46, 69, 230, 45, 25, 127, 121, 196, 37, 112, 130, 155, 240, 47, 83, 25, 151, 82, 247, 255, 146, 54, 40, 8, 22, 33, 92, 112, 85, 158, 43, 170, 58, 165, 12, 161, 97, 195, 67, 178, 12, 205, 119, 98, 161, 5, 238, 75, 188, 12, 65, 35, 195, 197, 71, 12, 24, 0, 141, 27, 135, 228, 196, 136, 190, 195, 72, 246, 86, 65, 145, 95, 234, 80, 177, 91, 237, 186, 212, 231, 49, 118, 236, 156, 232, 5, 131, 218, 129, 213, 199, 46, 141, 47, 198, 4, 205, 31, 72, 91, 103, 125, 216, 54, 233, 222, 93, 203, 62, 96, 215, 42, 53, 147, 115, 69, 35, 151, 126, 249, 10, 38, 46, 89, 13, 146, 107, 14, 110, 109, 159, 19, 82, 95, 111, 104, 36, 251, 135, 148, 88, 231, 197, 131, 50, 254, 254, 249, 112, 125, 237, 19, 218, 198, 210, 207, 83, 75, 27, 117, 46, 176, 73, 65, 151, 40, 136, 133, 59, 32, 57, 51, 238, 222, 95, 134, 201, 21, 241, 153, 1, 176, 152, 179, 88, 118, 200, 130, 222, 210, 212, 11, 69, 229, 7, 56, 208, 241, 37, 148, 19, 111, 80, 221, 238, 247, 38, 88, 204, 43, 42, 13, 215, 25, 12, 100, 131, 34, 113, 100, 142, 170, 29, 20, 24, 196, 242, 46, 101, 3, 253, 111, 104, 25}

	var header core.Header
	var ciphers *core.CipherSuite
	var serverEcdh core.Ecdh

	prk, _ := base64.StdEncoding.DecodeString("kgvvQaBGfHNWCbZMkFWS1K07BgRXlnOo7CHTZF1bsmI=")

	flag := binary.BigEndian.Uint16(buf[10:12])
	if flag&core.NHP_FLAG_EXTENDEDLENGTH == 0 {
		header = (*curve.HeaderCurve)(unsafe.Pointer(&buf[0]))
		ciphers = core.NewCipherSuite(core.CIPHER_SCHEME_CURVE)
		serverEcdh = core.ECDHFromKey(ciphers.EccType, prk)
	} else {
		header = (*gmsm.HeaderGmsm)(unsafe.Pointer(&buf[0]))
		ciphers = core.NewCipherSuite(core.CIPHER_SCHEME_GMSM)
		serverEcdh = core.ECDHFromKey(ciphers.EccType, prk)
	}

	hmacHash := core.NewHash(ciphers.HashType)
	hmacHash.Write([]byte(core.InitialHashString))
	hmacHash.Write(serverEcdh.PublicKey())
	hmacHash.Write(buf[0 : header.Size()-core.HashSize])
	calculatedHmac := hmacHash.Sum(nil)

	fmt.Printf("%v\n", calculatedHmac)
	fmt.Printf("%v\n", header.HMACBytes())
}
