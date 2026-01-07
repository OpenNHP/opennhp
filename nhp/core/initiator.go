package core

import (
	"bytes"
	"compress/zlib"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"hash"
	"net"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/log"
)

type InitiatorScheme interface {
	CreateMsgAssemblerData(d *Device, md *MsgData) (mad *MsgAssemblerData, err error)
	DeriveMsgAssemblerDataFromPrevParserData(ppd *PacketParserData, t int, message []byte) (mad *MsgAssemblerData)
	SetPeerPublicKey(d *Device, mad *MsgAssemblerData, peerPk []byte) (err error)
	EncryptBody(d *Device, mad *MsgAssemblerData) (err error)
}

type MsgData struct {
	RemoteAddr     *net.UDPAddr      // used by agent and ac create a new connection or pick an existing connection for msg sending
	ConnData       *ConnectionData   // used by server to pick an existing connection for msg sending
	PrevParserData *PacketParserData // when PrevParserData is set, CipherScheme, RemoteAddr, ConnData, TransactionId and PeerPk will be overridden
	CipherScheme   int               // 0: sm2/sm4/sm3, 1: curve25519/chacha20/blake2s
	TransactionId  uint64
	HeaderType     int
	Compress       bool
	ClPkc          bool // 0: non-CL-PKC extented, 1: CL-PKC extended
	ExternalPacket *Packet
	ExternalCookie *[CookieSize]byte
	Message        []byte
	PeerPk         []byte
	EncryptedPktCh chan *MsgAssemblerData
	ResponseMsgCh  chan *PacketParserData
}

func (d *Device) validateMsgData(md *MsgData) (err error) {
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
	BasePacket *Packet
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
	TransactionId uint64
	noise         NoiseFactory // int
	CipherScheme  int
	HeaderType    int
	BodySize      int
	HeaderFlag    uint16
	BodyCompress  bool
	ClPkc         bool

	ExternalCookie *[CookieSize]byte
	RemotePubKey   []byte
	bodyMessage    []byte

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
		mad.CipherScheme = md.CipherScheme
		mad.HeaderType = md.HeaderType
		mad.RemotePubKey = md.PeerPk
		mad.BodyCompress = md.Compress
		mad.ClPkc = md.ClPkc
		mad.bodyMessage = md.Message
		mad.TransactionId = md.TransactionId
		mad.connData = md.ConnData
		mad.encryptedPktCh = md.EncryptedPktCh

		// init packet buffer
		if md.ExternalPacket != nil {
			mad.BasePacket = md.ExternalPacket
		} else {
			mad.BasePacket = d.AllocatePoolPacket()
		}
		mad.BasePacket.HeaderType = mad.HeaderType

		// init cookie if specified
		if md.ExternalCookie != nil {
			mad.ExternalCookie = md.ExternalCookie
		}

		// create header and init device ecdh
		log.Info("start encryption using CIPHER_SCHEME_%d(0: CURVE; 1: GMSM.)", mad.CipherScheme)
		mad.header = mad.BasePacket.HeaderWithCipherScheme(mad.CipherScheme)
		mad.ciphers = NewCipherSuite(mad.CipherScheme)
		mad.deviceEcdh = d.GetEcdhByCipherScheme(mad.CipherScheme)

		// init version
		mad.header.SetVersion(ProtocolVersionMajor, ProtocolVersionMinor)

		// init header counter
		mad.header.SetCounter(mad.TransactionId)

		// init chain hash -> ChainHash0
		mad.chainHash, err = NewHash(mad.ciphers.HashType)
		if err != nil {
			return nil, fmt.Errorf("failed to create chain hash: %w", err)
		}
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
	mad.hmacHash, err = NewHash(mad.ciphers.HashType)
	if err != nil {
		return nil, fmt.Errorf("failed to create HMAC hash: %w", err)
	}
	mad.hmacHash.Write([]byte(InitialHashString))

	// create ephermeral key
	ephermalEccType := mad.ciphers.EccType
	mad.ephermeralEcdh = NewECDH(ephermalEccType)
	copy(mad.header.EphermeralBytes(), mad.ephermeralEcdh.PublicKey())

	return mad, nil
}

func (mad *MsgAssemblerData) derivePacketParserData(pkt *Packet, initTime int64) (ppd *PacketParserData) {
	ppd = &PacketParserData{}
	ppd.device = mad.device
	ppd.basePacket = pkt
	ppd.CipherScheme = mad.CipherScheme
	ppd.ConnData = mad.connData
	ppd.LocalInitTime = initTime
	ppd.feedbackMsgCh = mad.ResponseMsgCh

	// init header and init device ecdh
	ppd.HeaderFlag = ppd.basePacket.Flag()
	ppd.header = ppd.basePacket.HeaderWithCipherScheme(ppd.CipherScheme)
	ppd.Ciphers = NewCipherSuite(ppd.CipherScheme)
	ppd.deviceEcdh = ppd.device.GetEcdhByCipherScheme(ppd.CipherScheme)

	// init chain hash -> ChainHash0
	var err error
	ppd.chainHash, err = NewHash(ppd.Ciphers.HashType)
	if err != nil {
		// This should never happen with valid CipherSuite parameters
		panic(fmt.Sprintf("failed to create chain hash in derivePacketParserData: %v", err))
	}
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
	mad.TransactionId = md.TransactionId
	mad.connData = md.ConnData

	// init packet buffer
	if md.ExternalPacket != nil {
		mad.BasePacket = md.ExternalPacket
	} else {
		mad.BasePacket = d.AllocatePoolPacket()
	}
	mad.BasePacket.HeaderType = NHP_KPL

	// create header
	mad.header = mad.BasePacket.HeaderWithCipherScheme(common.CIPHER_SCHEME_CURVE)
	mad.BasePacket.Content = mad.BasePacket.Buf[:mad.header.Size()]

	// init version
	mad.header.SetVersion(ProtocolVersionMajor, ProtocolVersionMinor)

	// init header counter
	mad.header.SetCounter(mad.TransactionId)

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
		err = ErrEmptyPeerPublicKey
		return err
	}

	lenMismatch := false
	switch mad.CipherScheme {
	case common.CIPHER_SCHEME_CURVE:
		if len(mad.RemotePubKey) != PublicKeySize {
			lenMismatch = true
		}

	case common.CIPHER_SCHEME_GMSM:
		if len(mad.RemotePubKey) != PublicKeySizeEx {
			lenMismatch = true
		}
	default:
		log.Error("cipher scheme not implemented") // should never get here
		err = ErrDeviceECDHPeerFailed
		return err
	}

	if lenMismatch {
		log.Error("remote peer public key length does not match cipher scheme")
		err = ErrDeviceECDHPeerFailed
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
		err = ErrEphermalECDHPeerFailed
		return err
	}

	// prepare key for aead
	var key [SymmetricKeySize]byte

	// generate gcm key and encrypt device pubkey ChainKey1 -> ChainKey2 (ChainKey5 -> ChainKey6)
	mad.noise.KeyGen2(&mad.chainKey, &key, mad.chainKey[:], ess[:])
	SetZero(ess[:])

	var aead cipher.AEAD
	var static []byte
	// encrypt initiator's public key and evolve chainhash with the ciphertext
	switch mad.CipherScheme {
	case common.CIPHER_SCHEME_CURVE:
		fallthrough
	case common.CIPHER_SCHEME_GMSM:
		fallthrough
	default:
		aead, err = AeadFromKey(mad.ciphers.GcmType, &key)
		if err != nil {
			log.Error("failed to create AEAD for static encryption: %v", err)
			return err
		}
		static = aead.Seal(mad.header.StaticBytes()[:0], mad.header.NonceBytes(), mad.deviceEcdh.PublicKey(), mad.chainHash.Sum(nil))
	}

	//log.Debug("encrypted pubkey: %v, output: %v", mad.deviceEcdh.PublicKey(), static)

	// evolve chainhash ChainHash1 -> ChainHash2
	mad.chainHash.Write(static)

	// init shared key
	ss := mad.deviceEcdh.SharedSecret(mad.RemotePubKey)
	if ss == nil {
		log.Error("device ECDH failed with peer")
		err = ErrDeviceECDHPeerFailed
		return err
	}

	// generate gcm key and encrypt timestamp ChainKey2 -> ChainKey3 (ChainKey6 -> ChainKey7)
	mad.noise.KeyGen2(&mad.chainKey, &key, mad.chainKey[:], ss[:])
	SetZero(ss[:])

	var tsBytes [TimestampSize]byte
	var ts []byte
	binary.BigEndian.PutUint64(tsBytes[:], uint64(mad.LocalInitTime))

	switch mad.CipherScheme {
	case common.CIPHER_SCHEME_CURVE:
		fallthrough
	case common.CIPHER_SCHEME_GMSM:
		fallthrough
	default:
		aead, err = AeadFromKey(mad.ciphers.GcmType, &key)
		if err != nil {
			log.Error("failed to create AEAD for timestamp encryption: %v", err)
			return err
		}
		ts = aead.Seal(mad.header.TimestampBytes()[:0], mad.header.NonceBytes(), tsBytes[:], mad.chainHash.Sum(nil))
	}

	// evolve chainhash ChainHash2 -> ChainHash3
	mad.chainHash.Write(ts)

	// generate gcm key for body encryption ChainKey3 -> ChainKey4 (ChainKey7 -> ChainKey8)
	mad.noise.KeyGen2(&mad.chainKey, &key, mad.chainKey[:], ts[:])
	mad.bodyAead, err = AeadFromKey(mad.ciphers.GcmType, &key)
	if err != nil {
		log.Error("failed to create AEAD for body encryption: %v", err)
		return err
	}

	return err
}

func (mad *MsgAssemblerData) encryptBody() (err error) {
	defer func() {
		// clear secrets
		mad.chainHash.Reset()
		mad.chainHash = nil
		SetZero(mad.chainKey[:])
	}()

	// message body is empty, skip encryption. Set header and calculate HMAC
	if len(mad.bodyMessage) == 0 {
		// set header type and payload size
		mad.header.SetTypeAndPayloadSize(mad.HeaderType, 0)
		// set HMAC
		mad.addHMAC(mad.HeaderType == NHP_RKN)
		mad.BasePacket.Content = mad.BasePacket.Buf[:mad.header.Size()]
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
			ErrDataCompressionFailed.SetExtraError(err)
			err = ErrDataCompressionFailed
			return err
		}
		body = buf.Bytes()
		//log.Debug("message compressed: %v -> %v", mad.bodyMessage, body)
		mad.BodySize = len(body) + GCMTagSize

		// set header flag
		mad.HeaderFlag |= common.NHP_FLAG_COMPRESS

	} else {
		// no compress
		body = mad.bodyMessage
		mad.BodySize = len(mad.bodyMessage) + GCMTagSize
	}

	if mad.ClPkc {
		mad.HeaderFlag |= common.NHP_FLAG_CL_PKC
	}

	if mad.BodySize > PacketBufferSize-mad.header.Size() {
		log.Critical("message too long, send buffer exceeded")
		err = ErrPacketSizeExceedsBuffer
		return err
	}

	// set header flag
	mad.header.SetFlag(mad.HeaderFlag)

	// calculate total data length
	packetLen := mad.header.Size() + mad.BodySize

	// set header type and payload size
	mad.header.SetTypeAndPayloadSize(mad.HeaderType, mad.BodySize)

	// set HMAC
	mad.addHMAC(mad.HeaderType == NHP_RKN)

	// encrypt body and write into mad.BasePacket.Buf space
	ciphertext := mad.bodyAead.Seal(mad.BasePacket.Buf[mad.header.Size():mad.header.Size()], mad.header.NonceBytes(), body, mad.chainHash.Sum(nil))
	_ = ciphertext
	//log.Debug("encrypted body: %v, output: %v", body, ciphertext)

	// set valid packet
	mad.BasePacket.Content = mad.BasePacket.Buf[:packetLen]

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
		// use specified cookie, otherwise use connection's cookie
		if mad.ExternalCookie != nil {
			mad.hmacHash.Write((*mad.ExternalCookie)[:])
		} else {
			mad.connData.Lock()
			mad.hmacHash.Write(mad.connData.CookieStore.CurrCookie[:])
			mad.connData.Unlock()
		}
	}
	mad.hmacHash.Sum(mad.header.HMACBytes()[:0])
}

func (mad *MsgAssemblerData) Destroy() {
	mad.device.ReleasePoolPacket(mad.BasePacket)
	if mad.hmacHash != nil {
		mad.hmacHash.Reset()
		mad.hmacHash = nil
	}
	if mad.chainHash != nil {
		mad.chainHash.Reset()
		mad.chainHash = nil
	}
}
