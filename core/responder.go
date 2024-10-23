package core

import (
	"bytes"
	"compress/zlib"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/core/scheme/curve"
	"github.com/OpenNHP/opennhp/core/scheme/gmsm"
	"github.com/OpenNHP/opennhp/log"
)

type ResponderScheme interface {
	CreatePacketParserData(d *Device, pd *PacketData) (ppd *PacketParserData, err error)
	DerivePacketParserDataFromPrevAssemblerData(mad *MsgAssemblerData, pkt *Packet, initTime int64) (ppd *PacketParserData)
	validatePeer(d *Device, ppd *PacketParserData) (err error)
	decryptBody(d *Device, ppd *PacketParserData) (err error)
}

type CookieStore struct {
	CurrCookie     [CookieSize]byte
	PrevCookie     [CookieSize]byte
	LastCookieTime int64
}

func (cs *CookieStore) Set(cookie []byte) {
	copy(cs.PrevCookie[:], cs.CurrCookie[:])
	copy(cs.CurrCookie[:], cookie)
}

func (cs *CookieStore) Clear() {
	SetZero(cs.CurrCookie[:])
	SetZero(cs.PrevCookie[:])
}

type PacketData struct {
	BasePacket             *Packet
	ConnData               *ConnectionData
	PrevAssemblerData      *MsgAssemblerData
	ConnLastRemoteSendTime *int64
	ConnCookieStore        *CookieStore
	ConnPeerPublicKey      *[PublicKeySizeEx]byte
	InitTime               int64
	DecryptedMsgCh         chan *PacketParserData
}

type PacketParserData struct {
	device       *Device
	basePacket   *Packet
	ConnData     *ConnectionData
	CipherScheme int
	Ciphers      *CipherSuite

	deviceEcdh Ecdh
	header     Header
	hmacHash   hash.Hash
	chainHash  hash.Hash
	bodyAead   cipher.AEAD
	chainKey   [SymmetricKeySize]byte

	LocalInitTime int64
	SenderTrxId   uint64

	noise        NoiseFactory // int
	HeaderType   int
	BodySize     int
	HeaderFlag   uint16
	BodyCompress bool
	Overload     bool

	SenderIdentity         []byte
	SenderMidPublicKey     []byte
	ConnLastRemoteSendTime *int64
	ConnCookieStore        *CookieStore
	ConnPeerPublicKey      *[PublicKeySizeEx]byte
	RemotePubKey           []byte
	BodyMessage            []byte

	decryptedMsgCh chan<- *PacketParserData
	feedbackMsgCh  chan<- *PacketParserData
	Error          error
}

func (d *Device) createPacketParserData(pd *PacketData) (ppd *PacketParserData, err error) {
	if pd.PrevAssemblerData != nil {
		ppd = pd.PrevAssemblerData.derivePacketParserData(pd.BasePacket, pd.InitTime)
	} else {
		ppd = &PacketParserData{}
		ppd.device = d
		ppd.basePacket = pd.BasePacket
		ppd.ConnData = pd.ConnData
		ppd.ConnCookieStore = pd.ConnCookieStore
		ppd.LocalInitTime = pd.InitTime
		ppd.ConnLastRemoteSendTime = pd.ConnLastRemoteSendTime
		ppd.ConnPeerPublicKey = pd.ConnPeerPublicKey
		ppd.decryptedMsgCh = pd.DecryptedMsgCh

		// init header and init device ecdh
		ppd.HeaderFlag = binary.BigEndian.Uint16(ppd.basePacket.Content[10:12])
		if ppd.HeaderFlag&NHP_FLAG_EXTENDEDLENGTH == 0 {
			log.Info("start decryption using CIPHER_SCHEME_CURVE")
			ppd.CipherScheme = CIPHER_SCHEME_CURVE
			ppd.header = (*curve.HeaderCurve)(unsafe.Pointer(&ppd.basePacket.Content[0]))
			ppd.Ciphers = NewCipherSuite(CIPHER_SCHEME_CURVE)
			ppd.deviceEcdh = d.staticEcdhCurve
		} else {
			// check cipher scheme
			switch ppd.HeaderFlag & (0xF << 12) {
			case NHP_FLAG_SCHEME_GMSM:
				fallthrough
			default:
				log.Info("start decryption using CIPHER_SCHEME_GMSM")
				ppd.CipherScheme = CIPHER_SCHEME_GMSM
				ppd.header = (*gmsm.HeaderGmsm)(unsafe.Pointer(&ppd.basePacket.Content[0]))
				ppd.Ciphers = NewCipherSuite(CIPHER_SCHEME_GMSM)
				ppd.deviceEcdh = d.staticEcdhGmsm
			}
		}

		// init chain hash -> ChainHash0
		ppd.chainHash = NewHash(ppd.Ciphers.HashType)
		ppd.chainHash.Write([]byte(InitialHashString))

		// init chain key -> ChainKey0
		ppd.noise.HashType = ppd.Ciphers.HashType
		ppd.noise.MixKey(&ppd.chainKey, ppd.chainHash.Sum(nil), []byte(InitialChainKeyString))
	}

	ppd.HeaderType, ppd.BodySize = ppd.header.TypeAndPayloadSize()

	// init hmac hash -> HmacHash0
	ppd.hmacHash = NewHash(ppd.Ciphers.HashType)
	ppd.hmacHash.Write([]byte(InitialHashString))

	// evolve hmac hash HmacHash0 -> HmacHash1
	ppd.hmacHash.Write(ppd.deviceEcdh.PublicKey())

	// check hmac
	if ppd.device.deviceType == NHP_SERVER {
		// server overload handling
		overload := ppd.device.IsOverload()
		if overload {
			// overload, further discard unwanted packet type
			ppd.Overload = true
			if !ppd.IsAllowedAtOverload() {
				log.Critical("discard packet type %d due to overload", ppd.HeaderType)
				err = ErrServerOverload
				return
			}
		}

		// for RKN, check HMAC with cookie. For remaining allowed key messages, check HMAC without cookie.
		sumCookie := overload && ppd.HeaderType == NHP_RKN
		if !ppd.checkHMAC(sumCookie) {
			log.Error("HMAC validation failed on server side. sumCookie: %v", sumCookie)
			err = ErrServerHMACCheckFailed
			return
		}

	} else {
		if !ppd.checkHMAC(false) {
			log.Error("HMAC validation failed.")
			err = ErrHMACCheckFailed
			return
		}
	}

	// get sender id
	ppd.SenderTrxId = ppd.header.Counter()

	// init body message
	ppd.BodyCompress = ppd.HeaderFlag&NHP_FLAG_COMPRESS != 0
	ppd.BodyMessage = nil

	return ppd, nil
}

func (ppd *PacketParserData) deriveMsgAssemblerData(t int, compress bool, message []byte) (mad *MsgAssemblerData) {
	mad = &MsgAssemblerData{}
	mad.device = ppd.device
	mad.connData = ppd.ConnData
	mad.HeaderType = t
	mad.ciphers = ppd.Ciphers
	mad.CipherScheme = ppd.CipherScheme
	mad.RemotePubKey = ppd.RemotePubKey
	mad.BodyCompress = compress
	mad.bodyMessage = message

	// init packet buffer
	mad.BasePacket = mad.device.AllocatePoolPacket()
	mad.BasePacket.HeaderType = t

	// create header and init device ecdh
	switch mad.CipherScheme {
	case CIPHER_SCHEME_CURVE:
		mad.header = (*curve.HeaderCurve)(unsafe.Pointer(&mad.BasePacket.Buf[0]))
		mad.deviceEcdh = mad.device.staticEcdhCurve
	case CIPHER_SCHEME_GMSM:
		fallthrough
	default:
		mad.header = (*gmsm.HeaderGmsm)(unsafe.Pointer(&mad.BasePacket.Buf[0]))
		mad.deviceEcdh = mad.device.staticEcdhGmsm

	}

	// continue with the sender's counter
	mad.header.SetCounter(ppd.SenderTrxId)

	// init chain hash -> ChainHash0
	mad.chainHash = NewHash(mad.ciphers.HashType)
	mad.chainHash.Write([]byte(InitialHashString))

	// continue with responder's chain key -> ChainKey4
	mad.noise.HashType = ppd.Ciphers.HashType
	copy(mad.chainKey[:], ppd.chainKey[:])

	return mad
}

func shouldCheckRecvAttack(deviceType int, peerType int, msgType int) bool {
	if (deviceType == NHP_SERVER && peerType == NHP_AC && msgType == NHP_ART) ||
		(deviceType == NHP_AC && peerType == NHP_SERVER && msgType == NHP_AOP) {
		return false
	}

	return true
}

func (ppd *PacketParserData) validatePeer() (err error) {
	// evolve chain hash ChainHash0 -> ChainHash1
	ppd.chainHash.Write(ppd.deviceEcdh.PublicKey())
	ppd.chainHash.Write(ppd.header.EphermeralBytes())

	// evolve chain key ChainKey0 -> ChainKey1 (ChainKey4 -> ChainKey5)
	ppd.noise.MixKey(&ppd.chainKey, ppd.chainKey[:], ppd.header.EphermeralBytes())

	// get ephermeral shared key
	ess := ppd.deviceEcdh.SharedSecret(ppd.header.EphermeralBytes())
	if ess == nil {
		log.Error("device ECDH failed with ephermal")
		err = ErrDeviceECDHEphermalFailed
		return err
	}

	// prepare key for aead
	var key [SymmetricKeySize]byte
	var aead cipher.AEAD

	// generate gcm key and decrypt device pubkey ChainKey1 -> ChainKey2 (ChainKey5 -> ChainKey6)
	ppd.noise.KeyGen2(&ppd.chainKey, &key, ppd.chainKey[:], ess[:])
	SetZero(ess[:])

	peerPk := make([]byte, PublicKeySizeEx)
	switch ppd.CipherScheme {
	case CIPHER_SCHEME_CURVE:
		fallthrough
	case CIPHER_SCHEME_GMSM:
		fallthrough
	default:
		aead = AeadFromKey(ppd.Ciphers.GcmType, &key)
		_, err = aead.Open(peerPk[:0], ppd.header.NonceBytes(), ppd.header.StaticBytes(), ppd.chainHash.Sum(nil))
		if err != nil {
			log.Error("failed to decrypt peer pubkey")
			return err
		}
	}

	if ppd.CipherScheme == CIPHER_SCHEME_CURVE {
		peerPk = peerPk[:PublicKeySize]
	}

	//log.Debug("decrypted pubkey: %v, input: %v", peerPk, ppd.header.StaticBytes())

	// validate peer public key if they already exists in peer pool
	// also validate peer address if it has been changed
	// NOTE: to relieve ac from managing arbitrary agent peers,
	// ac does not validate nor store agent public key. Related msgtype: NHP_ACC.
	var peer Peer
	var toValidate bool
	var peerDeviceType int

	ppd.device.optionMutex.Lock()
	option := ppd.device.option
	ppd.device.optionMutex.Unlock()

	peerDeviceType = HeaderTypeToDeviceType(ppd.HeaderType)
	switch peerDeviceType {
	case NHP_AGENT:
		toValidate = !option.DisableAgentPeerValidation

	case NHP_SERVER:
		toValidate = !option.DisableServerPeerValidation

	case NHP_AC:
		toValidate = !option.DisableACPeerValidation

	case NHP_RELAY:
		toValidate = !option.DisableRelayPeerValidation
	}

	if toValidate {
		peer = ppd.device.LookupPeer(peerPk)
		if peer == nil {
			log.Error("peer not found in peer pool")
			err = fmt.Errorf("peer not found in peer pool")
			return err
		}

		if peer.IsExpired() {
			log.Error("peer expired")
			err = fmt.Errorf("peer expired")
			return err
		}

		if !peer.CheckRecvAddress(ppd.LocalInitTime, ppd.ConnData.RemoteAddr) {
			log.Error("peer does not match its previous address")
			err = fmt.Errorf("peer does not match its previous address")
			return err
		}
		peer.UpdateRecv(ppd.LocalInitTime, ppd.ConnData.RemoteAddr)
	}

	ppd.RemotePubKey = peerPk
	if ppd.ConnPeerPublicKey != nil {
		copy((*ppd.ConnPeerPublicKey)[:], peerPk)
	}

	// evolve chainhash ChainHash1 -> ChainHash2
	ppd.chainHash.Write(ppd.header.StaticBytes())

	// init shared key
	ss := ppd.deviceEcdh.SharedSecret(peerPk)
	if ss == nil {
		log.Error("device ECDH failed with obtained peer")
		err = ErrDeviceECDHObtainedPeerFailed
		return err
	}

	// generate gcm key and decrypt timestamp ChainKey2 -> ChainKey3 (ChainKey6 -> ChainKey7)
	ppd.noise.KeyGen2(&ppd.chainKey, &key, ppd.chainKey[:], ss[:])
	SetZero(ss[:])

	var tsBytes [TimestampSize]byte
	switch ppd.CipherScheme {
	case CIPHER_SCHEME_CURVE:
		fallthrough
	case CIPHER_SCHEME_GMSM:
		fallthrough
	default:
		aead = AeadFromKey(ppd.Ciphers.GcmType, &key)
		_, err = aead.Open(tsBytes[:0], ppd.header.NonceBytes(), ppd.header.TimestampBytes(), ppd.chainHash.Sum(nil))
		if err != nil {
			log.Error("failed to decrypt timestamp")
			return err
		}
	}

	remoteSendTime := int64(binary.BigEndian.Uint64(tsBytes[:]))

	if shouldCheckRecvAttack(ppd.device.deviceType, peerDeviceType, ppd.HeaderType) {
		// block remote if threat level is reached
		if remoteSendTime < ppd.ConnData.LastRemoteSendTime {
			// replay packet, drop
			log.Critical("received replay packet from %s, drop packet", ppd.ConnData.RemoteAddr.String())
			// threat plus 1
			threat := atomic.AddInt32(&ppd.ConnData.RecvThreatCount, 1)
			// with high queue number, the device may use ConnData channels when conn is already closed
			if threat > ThreatCountBeforeBlock && !ppd.ConnData.IsClosed() {
				// clamp threat count to avoid overflow
				atomic.StoreInt32(&ppd.ConnData.RecvThreatCount, ThreatCountBeforeBlock)
				// block source address
				ppd.ConnData.SendBlockSignal()
			}
			err = fmt.Errorf("received replay packet")
			return err
		}
		if remoteSendTime < ppd.ConnData.LastRemoteSendTime+MinimalRecvIntervalMs*int64(time.Millisecond) {
			// flood packet, drop
			log.Critical("received flood packet from %s, drop packet", ppd.ConnData.RemoteAddr.String())
			// threat plus 1
			threat := atomic.AddInt32(&ppd.ConnData.RecvThreatCount, 1)
			if threat > ThreatCountBeforeBlock && !ppd.ConnData.IsClosed() {
				// clamp threat count to avoid overflow
				atomic.StoreInt32(&ppd.ConnData.RecvThreatCount, ThreatCountBeforeBlock)
				// block source address
				ppd.ConnData.SendBlockSignal()
			}
			err = fmt.Errorf("received flood packet")
			return err
		}
	}
	if remoteSendTime < (ppd.LocalInitTime - 600*int64(time.Second)) {
		// send remote timestamp is too old than receive local time, drop
		// note there might be time calibration error between remote and local devices
		log.Critical("received stale packet from %s, drop packet", ppd.ConnData.RemoteAddr.String())
		threat := atomic.AddInt32(&ppd.ConnData.RecvThreatCount, 1)
		if threat > ThreatCountBeforeBlock && !ppd.ConnData.IsClosed() {
			// clamp threat count to avoid overflow
			atomic.StoreInt32(&ppd.ConnData.RecvThreatCount, ThreatCountBeforeBlock)
			// block source address
			ppd.ConnData.SendBlockSignal()
		}
		err = fmt.Errorf("received stale packet")
		return err
	}

	// update remote last send time
	atomic.StoreInt64(&ppd.ConnData.LastRemoteSendTime, remoteSendTime)
	// clear threat
	atomic.StoreInt32(&ppd.ConnData.RecvThreatCount, 0)

	// handle knock packet at overload before going into body decryption
	if ppd.device.deviceType == NHP_SERVER && ppd.Overload && ppd.HeaderType == NHP_KNK {
		switch ppd.CipherScheme {
		case CIPHER_SCHEME_CURVE:
			fallthrough
		case CIPHER_SCHEME_GMSM:
			fallthrough
		default:
			ppd.generateCookie()
			ppd.sendCookie()
			err = ErrServerRejectWithCookie
		}

		return err
	}

	// evolve chainhash ChainHash2 -> ChainHash3
	ppd.chainHash.Write(ppd.header.TimestampBytes())

	// generate gcm key for body decryption ChainKey3 -> ChainKey4 (ChainKey7 -> ChainKey8)
	ppd.noise.KeyGen2(&ppd.chainKey, &key, ppd.chainKey[:], ppd.header.TimestampBytes())
	ppd.bodyAead = AeadFromKey(ppd.Ciphers.GcmType, &key)

	return nil
}

func (ppd *PacketParserData) decryptBody() (err error) {
	defer func() {
		// clear secrets
		ppd.chainHash.Reset()
		ppd.chainHash = nil
		SetZero(ppd.chainKey[:])
	}()

	// message body is empty, skip decryption
	if len(ppd.basePacket.Content) == ppd.header.Size() {
		return nil
	}

	// decrypt body and reuse ppd.BasePacket.Content space
	body, err := ppd.bodyAead.Open(ppd.basePacket.Content[ppd.header.Size():ppd.header.Size()], ppd.header.NonceBytes(), ppd.basePacket.Content[ppd.header.Size():], ppd.chainHash.Sum(nil))
	if err != nil {
		log.Critical("decrypt body failed: %v", err)
		ErrAEADDecryptionFailed.SetExtraError(err)
		err = ErrAEADDecryptionFailed
		return err
	}

	//log.Debug("decrypted body: %v, input: %v", body, ppd.basePacket.Content[ppd.header.Size():])

	// Note: ppd.BodyMessage must be a separate []byte slice because ppd.BasePacket.Buf will be released later
	if ppd.BodyCompress {
		// decompress
		var buf bytes.Buffer
		br := bytes.NewReader(body)
		r, _ := zlib.NewReader(br)
		defer r.Close()

		_, err = io.Copy(&buf, r)
		if err != nil {
			log.Critical("message decompression failed: %v", err)
			ErrDataDecompressionFailed.SetExtraError(err)
			err = ErrDataDecompressionFailed
			return err
		}

		ppd.BodyMessage = buf.Bytes() // separately allocated memory
		//log.Debug("message decompressed %v -> %v", body, ppd.BodyMessage)
	} else {
		ppd.BodyMessage = append(ppd.BodyMessage, body...) // deep copy
	}

	return nil
}

func (ppd *PacketParserData) makeCookieStore(cookieStore *CookieStore) *CookieStore {
	if cookieStore != nil {
		var tsBytes [TimestampSize]byte
		currTime := time.Now().UnixNano()
		if (currTime - cookieStore.LastCookieTime) > CookieRegenerateTime*int64(time.Second) {
			copy(cookieStore.PrevCookie[:], cookieStore.CurrCookie[:])
			binary.BigEndian.PutUint64(tsBytes[:], uint64(currTime))
			ppd.noise.KeyGen1(&cookieStore.CurrCookie, ppd.header.EphermeralBytes(), tsBytes[:])
			cookieStore.LastCookieTime = currTime
		}
		return cookieStore
	}

	return nil
}

func (ppd *PacketParserData) generateCookie() {
	var tsBytes [TimestampSize]byte
	currTime := time.Now().UnixNano()

	ppd.ConnData.Lock()
	defer ppd.ConnData.Unlock()

	if (currTime - ppd.ConnData.CookieStore.LastCookieTime) > CookieRegenerateTime*int64(time.Second) {
		copy(ppd.ConnData.CookieStore.PrevCookie[:], ppd.ConnData.CookieStore.CurrCookie[:])
		binary.BigEndian.PutUint64(tsBytes[:], uint64(currTime))
		ppd.noise.KeyGen1(&ppd.ConnData.CookieStore.CurrCookie, ppd.header.EphermeralBytes(), tsBytes[:])
		ppd.ConnData.CookieStore.LastCookieTime = currTime
	}
}

func (ppd *PacketParserData) sendCookie() {
	cokStr := base64.StdEncoding.EncodeToString(ppd.ConnData.CookieStore.CurrCookie[:])
	cokMsg := &common.ServerCookieMsg{
		TransactionId: ppd.SenderTrxId,
		Cookie:        cokStr,
	}
	cokBytes, _ := json.Marshal(cokMsg)

	md := &MsgData{
		HeaderType:    NHP_COK,
		TransactionId: ppd.device.NextCounterIndex(),
		Compress:      true,
		ConnData:      ppd.ConnData,
		PeerPk:        ppd.RemotePubKey,
		Message:       cokBytes,
	}

	log.Debug("Send cookie back to %s: %s ", ppd.ConnData.RemoteAddr, string(md.Message))
	ppd.device.SendMsgToPacket(md)
}

func (ppd *PacketParserData) checkHMAC(sumCookie bool) bool {
	defer func() {
		ppd.hmacHash.Reset()
		ppd.hmacHash = nil
	}()

	len := ppd.header.Size() - HashSize
	ppd.hmacHash.Write(ppd.header.Bytes()[0:len])

	if sumCookie {
		switch ppd.CipherScheme {
		case CIPHER_SCHEME_CURVE:
			fallthrough
		case CIPHER_SCHEME_GMSM:
			fallthrough
		default:
			ppd.ConnData.Lock()
			defer ppd.ConnData.Unlock()

			if ppd.LocalInitTime < ppd.ConnData.CookieStore.LastCookieTime+CookieRoundTripTimeMs*int64(time.Millisecond) {
				// cookie has already or nearly been updated, use previous cookie
				ppd.hmacHash.Write(ppd.ConnData.CookieStore.PrevCookie[:])
				prevCookieHmac := ppd.hmacHash.Sum(nil)
				return bytes.Equal(prevCookieHmac, ppd.header.HMACBytes())
			} else {
				// use current cookie
				ppd.hmacHash.Write(ppd.ConnData.CookieStore.CurrCookie[:])
				cookieHmac := ppd.hmacHash.Sum(nil)
				return bytes.Equal(cookieHmac, ppd.header.HMACBytes())
			}
		}
	} else {
		calculatedHmac := ppd.hmacHash.Sum(nil)
		return bytes.Equal(calculatedHmac, ppd.header.HMACBytes())
	}
}

func (ppd *PacketParserData) Destroy() {
	ppd.device.ReleasePoolPacket(ppd.basePacket)
	if ppd.hmacHash != nil {
		ppd.hmacHash.Reset()
		ppd.hmacHash = nil
	}
	if ppd.chainHash != nil {
		ppd.chainHash.Reset()
		ppd.chainHash = nil
	}
}

func (ppd *PacketParserData) IsAllowedAtOverload() bool {
	switch ppd.HeaderType {
	case NHP_KNK, NHP_RKN, NHP_EXT, NHP_AOL, NHP_ART:
		return true
	default:
		return false
	}
}
