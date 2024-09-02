package nhp

import (
	"bytes"
	"compress/zlib"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"hash"
	"net"
	"time"
	"unsafe"

	"github.com/OpenNHP/opennhp/log"
)

type MsgData struct {
	RemoteAddr     *net.UDPAddr      // used by agent and door create a new connection or pick an existing connection for msg sending
	ConnData       *ConnectionData   // used by server to pick an existing connection for msg sending
	PrevParserData *PacketParserData // when PrevParserData is set, RemoteAddr, ConnData, TransactionId and PeerPk will be overridden
	HeaderType     int
	TransactionId  uint64
	Compress       bool
	Message        []byte
	PeerPk         []byte
	EncryptedPktCh chan *MsgAssemblerData
	ResponseMsgCh  chan *PacketParserData
}

func (d *Device) ValidateMsgData(md *MsgData) (err error) {
	if md.PrevParserData == nil {
		if d.deviceType == NHP_SERVER && md.ConnData == nil {
			err = fmt.Errorf("missing connection data for server")
		} else if d.deviceType != NHP_SERVER && md.RemoteAddr == nil {
			err = fmt.Errorf("missing remote address")
		}

		if md.PeerPk == nil {
			err = fmt.Errorf("missing remote peer public key")
		}
	}

	return err
}

type MsgAssemblerData struct {
	device     *Device
	BasePacket *UdpPacket
	connData   *ConnectionData
	ciphers    *CipherSuite

	deviceEcdh     Ecdh
	ephermeralEcdh Ecdh
	header         Header
	hmacHash       hash.Hash
	chainHash      hash.Hash
	bodyAead       cipher.AEAD
	chainKey       [SymmetricKeySize]byte

	LocalInitTime int64
	noise         NoiseFactory // int
	HeaderType    int
	BodySize      int
	HeaderFlag    uint16
	BodyCompress  bool

	RemotePubKey []byte
	bodyMessage  []byte

	encryptedPktCh chan<- *MsgAssemblerData
	ResponseMsgCh  chan<- *PacketParserData
	Error          error
}

func (d *Device) createMsgAssemblerData(md *MsgData) (mad *MsgAssemblerData, err error) {
	if md.PrevParserData != nil {
		// continue from previous received packet to form one transaction
		mad = md.PrevParserData.deriveMsgAssemblerData(md.HeaderType, md.Compress, md.Message)
	} else {
		mad = &MsgAssemblerData{}
		mad.device = d
		mad.HeaderType = md.HeaderType
		mad.RemotePubKey = md.PeerPk
		mad.BodyCompress = md.Compress
		mad.bodyMessage = []byte(md.Message)
		mad.connData = md.ConnData
		mad.encryptedPktCh = md.EncryptedPktCh

		// init packet buffer
		mad.BasePacket = d.AllocateUdpPacket()
		mad.BasePacket.HeaderType = mad.HeaderType

		// create header and init device ecdh
		useGm := len(mad.RemotePubKey) == PublicKeySizeEx
		mad.ciphers = NewCipherSuite(useGm)
		if useGm {
			mad.HeaderFlag |= NHP_FLAG_EXTENDEDLENGTH
			mad.header = (*NHPHeaderEx)(unsafe.Pointer(&mad.BasePacket.Buf[0]))
			mad.deviceEcdh = d.staticEcdhEx
		} else {
			mad.header = (*NHPHeader)(unsafe.Pointer(&mad.BasePacket.Buf[0]))
			mad.deviceEcdh = d.staticEcdh
		}

		// init version
		mad.header.SetVersion(ProtocolVersionMajor, ProtocolVersionMinor)

		// init header counter
		mad.header.SetCounter(md.TransactionId)

		// init chain hash -> ChainHash0
		mad.chainHash = NewHash(mad.ciphers.HashType)
		mad.chainHash.Write([]byte(InitialHashString))

		// init chain key -> ChainKey0
		mad.noise.HashType = mad.ciphers.HashType
		mad.noise.MixKey(&mad.chainKey, mad.chainHash.Sum(nil), []byte(InitialChainKeyString))
	}

	// init timestamp
	mad.LocalInitTime = time.Now().UnixNano()

	// assign channel
	mad.ResponseMsgCh = md.ResponseMsgCh

	// init hmac hash -> HmacHash0
	mad.hmacHash = NewHash(mad.ciphers.HashType)
	mad.hmacHash.Write([]byte(InitialHashString))

	// create ephermeral key
	mad.ephermeralEcdh = NewECDH(mad.ciphers.EccType)
	copy(mad.header.EphermeralBytes(), mad.ephermeralEcdh.PublicKey())

	return mad, nil
}

func (mad *MsgAssemblerData) derivePacketParserData(pkt *UdpPacket, initTime int64) (ppd *PacketParserData) {
	ppd = &PacketParserData{}
	ppd.device = mad.device
	ppd.basePacket = pkt
	ppd.ConnData = mad.connData
	ppd.LocalInitTime = initTime
	ppd.feedbackMsgCh = mad.ResponseMsgCh

	// init header and init device ecdh
	ppd.HeaderFlag = binary.BigEndian.Uint16(ppd.basePacket.Packet[10:12])
	if ppd.HeaderFlag&NHP_FLAG_EXTENDEDLENGTH == 0 {
		ppd.header = (*NHPHeader)(unsafe.Pointer(&ppd.basePacket.Packet[0]))
		ppd.Ciphers = NewCipherSuite(false)
		ppd.deviceEcdh = ppd.device.staticEcdh
	} else {
		ppd.header = (*NHPHeaderEx)(unsafe.Pointer(&ppd.basePacket.Packet[0]))
		ppd.Ciphers = NewCipherSuite(true)
		ppd.deviceEcdh = ppd.device.staticEcdhEx
	}

	// init chain hash -> ChainHash0
	ppd.chainHash = NewHash(ppd.Ciphers.HashType)
	ppd.chainHash.Write([]byte(InitialHashString))

	// continue with initiator's chain key -> ChainKey4
	ppd.noise.HashType = mad.ciphers.HashType
	copy(ppd.chainKey[:], mad.chainKey[:])

	return ppd
}

func (d *Device) createKeepalivePacket(md *MsgData) (mad *MsgAssemblerData, err error) {
	mad = &MsgAssemblerData{}
	mad.device = d
	mad.HeaderType = NHP_KPL
	mad.connData = md.ConnData

	// init packet buffer
	mad.BasePacket = d.AllocateUdpPacket()
	mad.BasePacket.HeaderType = NHP_KPL
	mad.BasePacket.Packet = mad.BasePacket.Buf[:HeaderSize]

	// create header
	mad.header = (*NHPHeader)(unsafe.Pointer(&mad.BasePacket.Buf[0]))

	// init version
	mad.header.SetVersion(ProtocolVersionMajor, ProtocolVersionMinor)

	// init header counter
	mad.header.SetCounter(md.TransactionId)

	// set header flag
	mad.header.SetFlag(mad.HeaderFlag)

	// set header type and payload size
	mad.header.SetTypeAndPayloadSize(mad.HeaderType, 0)

	return mad, nil
}

func (mad *MsgAssemblerData) setPeerPublicKey(peerPk []byte) (err error) {
	// override peer public key
	if peerPk != nil {
		mad.RemotePubKey = peerPk
	}

	if mad.RemotePubKey == nil {
		log.Error("remote peer public key is not set")
		err = fmt.Errorf("remote peer public key is not set")
		return err
	}

	// evolve hmac hash HmacHash0 -> HmacHash1
	mad.hmacHash.Write(mad.RemotePubKey)

	// evolve chain hash ChainHash0 -> ChainHash1
	mad.chainHash.Write(mad.RemotePubKey)
	mad.chainHash.Write(mad.ephermeralEcdh.PublicKey())

	// evolve chain key ChainKey0 -> ChainKey1 (ChainKey4 -> ChainKey5)
	mad.noise.MixKey(&mad.chainKey, mad.chainKey[:], mad.ephermeralEcdh.PublicKey())

	// init ephermeral shared key
	ess := mad.ephermeralEcdh.SharedSecret(mad.RemotePubKey)
	if ess == nil {
		log.Error("ephermal ECDH failed with peer")
		err = fmt.Errorf("ephermal ECDH failed with peer")
		return err
	}

	// prepare key for aead
	var key [SymmetricKeySize]byte

	// generate gcm key and encrypt device pubkey ChainKey1 -> ChainKey2 (ChainKey5 -> ChainKey6)
	mad.noise.KeyGen2(&mad.chainKey, &key, mad.chainKey[:], ess[:])
	SetZero(ess[:])
	aead := AeadFromKey(mad.ciphers.GcmType, &key)
	static := aead.Seal(mad.header.StaticBytes()[:0], mad.header.NonceBytes(), mad.deviceEcdh.PublicKey(), mad.chainHash.Sum(nil))

	// evolve chainhash ChainHash1 -> ChainHash2
	mad.chainHash.Write(static)

	// init shared key
	ss := mad.deviceEcdh.SharedSecret(mad.RemotePubKey)
	if ss == nil {
		log.Error("device ECDH failed with peer")
		err = fmt.Errorf("device ECDH failed with peer")
		return err
	}

	// generate gcm key and encrypt timestamp ChainKey2 -> ChainKey3 (ChainKey6 -> ChainKey7)
	mad.noise.KeyGen2(&mad.chainKey, &key, mad.chainKey[:], ss[:])
	SetZero(ss[:])
	var tsBytes [TimestampSize]byte
	binary.BigEndian.PutUint64(tsBytes[:], uint64(mad.LocalInitTime))
	aead = AeadFromKey(mad.ciphers.GcmType, &key)
	ts := aead.Seal(mad.header.TimestampBytes()[:0], mad.header.NonceBytes(), tsBytes[:], mad.chainHash.Sum(nil))

	// evolve chainhash ChainHash2 -> ChainHash3
	mad.chainHash.Write(ts)

	// generate gcm key for body encryption ChainKey3 -> ChainKey4 (ChainKey7 -> ChainKey8)
	mad.noise.KeyGen2(&mad.chainKey, &key, mad.chainKey[:], ts[:])
	mad.bodyAead = AeadFromKey(mad.ciphers.GcmType, &key)

	return err
}

func (mad *MsgAssemblerData) encryptBody() (err error) {
	defer func() {
		mad.chainHash.Reset()
		mad.chainHash = nil
	}()

	// message body is empty, skip encryption. Set header and calculate HMAC
	if len(mad.bodyMessage) == 0 {
		// set header type and payload size
		mad.header.SetTypeAndPayloadSize(mad.HeaderType, 0)
		// set HMAC
		mad.addHMAC(mad.HeaderType == NHP_RKN)
		mad.BasePacket.Packet = mad.BasePacket.Buf[:mad.header.Size()]
		return nil
	}

	var body []byte

	if mad.BodyCompress {
		// compress
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)

		_, err = w.Write(mad.bodyMessage)
		w.Close()
		if err != nil {
			log.Critical("message compression failed: %v", err)
			return err
		}
		body = buf.Bytes()
		mad.BodySize = len(body)

		// set header flag
		mad.HeaderFlag |= NHP_FLAG_COMPRESS

	} else {
		// no compress
		body = mad.bodyMessage
		mad.BodySize = len(mad.bodyMessage)
	}

	if mad.BodySize > PacketBufferSize-GCMTagSize-mad.header.Size() {
		log.Critical("message too long, send buffer exceeded")
		err = fmt.Errorf("message longer than send buffer")
		return err
	}

	// set header flag
	mad.header.SetFlag(mad.HeaderFlag)

	// calculate total data length
	packetLen := mad.header.Size() + mad.BodySize + GCMTagSize

	// set header type and payload size
	mad.header.SetTypeAndPayloadSize(mad.HeaderType, mad.BodySize+GCMTagSize)

	// set HMAC
	mad.addHMAC(mad.HeaderType == NHP_RKN)

	// encrypt body and write into mad.BasePacket.Buf space
	mad.bodyAead.Seal(mad.BasePacket.Buf[mad.header.Size():mad.header.Size()], mad.header.NonceBytes(), body, mad.chainHash.Sum(nil))

	// set valid packet
	mad.BasePacket.Packet = mad.BasePacket.Buf[:packetLen]

	return nil
}

// must be called after header is filled
func (mad *MsgAssemblerData) addHMAC(sumCookie bool) {
	defer func() {
		mad.hmacHash.Reset()
		mad.hmacHash = nil
	}()

	len := mad.header.Size() - HashSize
	mad.hmacHash.Write(mad.header.Bytes()[0:len])

	if sumCookie {
		mad.connData.Lock()
		mad.hmacHash.Write(mad.connData.CookieStore.Cookie[:])
		mad.connData.Unlock()
	}
	mad.hmacHash.Sum(mad.header.HMACBytes()[:0])
}

func (mad *MsgAssemblerData) Destroy() {
	mad.device.ReleaseUdpPacket(mad.BasePacket)
	if mad.hmacHash != nil {
		mad.hmacHash.Reset()
		mad.hmacHash = nil
	}
	if mad.chainHash != nil {
		mad.chainHash.Reset()
		mad.chainHash = nil
	}
	SetZero(mad.chainKey[:])
}
