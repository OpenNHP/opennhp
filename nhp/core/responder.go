package core

import (
	"bytes"
	"compress/zlib"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"sync/atomic"
	"time"

	common "github.com/OpenNHP/opennhp/nhp/common"
	log "github.com/OpenNHP/opennhp/nhp/log"
)

// cookieRemoteKey extracts the IP portion of the real client address as a
// stable string suitable for cookie derivation. The port is deliberately
// dropped: an agent that round-robins its KNK and RKN across two
// nhp-server instances opens a separate kernel UDP conn to each, and
// each conn gets its own ephemeral source port — so the port the
// minting instance sees on the KNK is NOT the port the verifying
// instance sees on the RKN, even though the agent process is the same.
// Keying on IP only lets the second instance re-derive the same cookie.
//
// For relay-forwarded packets we key on ConnData.RealRemoteAddr (the
// actual client behind the relay) rather than RemoteAddr (the relay's
// own UDP address). If we keyed on the relay address, every client
// arriving through one public relay would share a cookie within a time
// window — that scales DoS-proof-of-address with relay throughput
// instead of per-client, and lets one authenticated relay user replay
// another's cookie. Falling back to RemoteAddr for direct UDP keeps
// the single-server path unchanged.
//
// IPv4 addresses are normalized via To4() so a dual-stack listener that
// presents the same agent as "::ffff:1.2.3.4" on one socket and
// "1.2.3.4" on another still derives the same cookie key.
//
// Trade-off: agents behind the same NAT share a cookie within a time
// window. The attack surface is narrow — cookie content stays off the
// wire (it's an HMAC input, never plaintext), windows are short (default
// 60s), and an attacker behind the same NAT can already send packets
// from any source port. Worth it to enable multi-instance routing.
func cookieRemoteKey(connData *ConnectionData) string {
	if connData == nil {
		return ""
	}
	addr := connData.RealRemoteAddr
	if addr == nil {
		addr = connData.RemoteAddr
	}
	if addr == nil || addr.IP == nil {
		return ""
	}
	if v4 := addr.IP.To4(); v4 != nil {
		return v4.String()
	}
	return addr.IP.String()
}

// deriveStatelessCookie produces a deterministic 32-byte cookie from the
// cluster-wide signing key, the remote IP (NOT ip:port — see
// cookieRemoteKey for why), the agent's static public key, and a
// time-window bucket. Any nhp-server instance configured with the same
// key can independently re-derive the cookie an agent received from a
// sibling instance, so the KNK → COK → RKN handshake survives
// load-balancer or agent-side LB shuffling between instances.
//
// peerPk binds each cookie to a specific agent identity. Without it,
// two distinct agents sharing one NAT/CGN egress IP would derive
// identical cookies within a window — any one of them could mint a
// cookie via its own legitimate KNK and let a sibling behind the same
// NAT replay that value in an unrelated RKN. Sibling-server
// re-derivation is unaffected because the agent's static pubkey is the
// same regardless of which server instance handles the packet. This
// also blunts the relay-trust concern: a compromised relay can spoof
// SourceAddr but cannot forge another agent's static pubkey, so a
// minted cookie remains scoped to whichever peer the relay claims.
func deriveStatelessCookie(signingKey []byte, remoteKey string, peerPk []byte, windowIndex int64) []byte {
	mac := hmac.New(sha256.New, signingKey)
	mac.Write([]byte(remoteKey))
	mac.Write(peerPk)
	var wb [8]byte
	binary.BigEndian.PutUint64(wb[:], uint64(windowIndex))
	mac.Write(wb[:])
	return mac.Sum(nil) // 32 bytes (HashSize/CookieSize)
}

// extractInitiatorStaticPubKey re-runs the Noise IK static-decryption
// step (device-ecdh-with-ephemeral → KDF → AEAD-Open of header.Static)
// using purely local state, so the caller can learn the initiator's
// static public key BEFORE validatePeer formally walks the Noise
// transcript. It is used only by the overload cookie-verify path,
// where checkHMAC needs to bind the cookie to the peer identity but
// runs before validatePeer has populated ppd.RemotePubKey.
//
// IMPORTANT: this function MUST NOT touch ppd.chainHash, ppd.chainKey,
// ppd.noise, or any other state that validatePeer will later mutate.
// validatePeer reproduces this same ECDH/AEAD computation on its own
// instances; doing it here too costs one extra ECDH per RKN-under-
// overload packet, but keeps the noise-transcript machinery untouched
// (and untouchable from this helper, which only takes a *Device, the
// cipher suite, and the header).
func extractInitiatorStaticPubKey(device *Device, ciphers *CipherSuite, header Header) ([]byte, error) {
	deviceEcdh := device.GetEcdhByCipherScheme(header.CipherScheme())

	ess := deviceEcdh.SharedSecret(header.EphermeralBytes())
	if ess == nil {
		return nil, ErrDeviceECDHEphermalFailed
	}
	defer SetZero(ess[:])

	// Local chain hash: InitialHash || serverPubKey || ephemeral
	// (mirrors validatePeer's ChainHash0 → ChainHash1 evolution, but
	// in a throwaway hash that never leaks back to ppd).
	chainHash, err := NewHash(ciphers.HashType)
	if err != nil {
		return nil, fmt.Errorf("extractInitiatorStaticPubKey: chain hash: %w", err)
	}
	chainHash.Write([]byte(InitialHashString))
	chainHash.Write(deviceEcdh.PublicKey())
	chainHash.Write(header.EphermeralBytes())

	// Local chain key: ChainKey0 = MixKey(InitialHash, InitialChainKey)
	// then ChainKey0 → ChainKey1 via ess.
	var noise NoiseFactory
	noise.HashType = ciphers.HashType
	var chainKey [SymmetricKeySize]byte
	defer SetZero(chainKey[:])
	// ChainKey0
	initHash, err := NewHash(ciphers.HashType)
	if err != nil {
		return nil, fmt.Errorf("extractInitiatorStaticPubKey: init hash: %w", err)
	}
	initHash.Write([]byte(InitialHashString))
	noise.MixKey(&chainKey, initHash.Sum(nil), []byte(InitialChainKeyString))
	// ChainKey0 → ChainKey1
	noise.MixKey(&chainKey, chainKey[:], header.EphermeralBytes())

	// Derive AEAD key for static-field decryption.
	var key [SymmetricKeySize]byte
	defer SetZero(key[:])
	noise.KeyGen2(&chainKey, &key, chainKey[:], ess[:])

	aead, err := AeadFromKey(ciphers.GcmType, &key)
	if err != nil {
		return nil, fmt.Errorf("extractInitiatorStaticPubKey: aead: %w", err)
	}
	peerPk := make([]byte, PublicKeySizeEx)
	if _, err := aead.Open(peerPk[:0], header.NonceBytes(), header.StaticBytes(), chainHash.Sum(nil)); err != nil {
		return nil, fmt.Errorf("extractInitiatorStaticPubKey: open: %w", err)
	}
	if header.CipherScheme() == common.CIPHER_SCHEME_CURVE {
		peerPk = peerPk[:PublicKeySize]
	}
	return peerPk, nil
}

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

	decryptedMsgCh chan<- *PacketParserData //Plaintext payload dispatched (Decryption cycle completed)
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
		ppd.HeaderFlag = ppd.basePacket.Flag()
		ppd.header = ppd.basePacket.Header()
		ppd.CipherScheme = ppd.header.CipherScheme()
		log.Info("start decryption using CIPHER_SCHEME_%d(0: CURVE; 1: GMSM.)", ppd.CipherScheme)
		ppd.Ciphers = NewCipherSuite(ppd.CipherScheme)
		ppd.deviceEcdh = d.GetEcdhByCipherScheme(ppd.CipherScheme)

		// init chain hash -> ChainHash0
		ppd.chainHash, err = NewHash(ppd.Ciphers.HashType)
		if err != nil {
			return nil, fmt.Errorf("failed to create chain hash: %w", err)
		}
		ppd.chainHash.Write([]byte(InitialHashString))

		// init chain key -> ChainKey0
		ppd.noise.HashType = ppd.Ciphers.HashType
		ppd.noise.MixKey(&ppd.chainKey, ppd.chainHash.Sum(nil), []byte(InitialChainKeyString))
	}

	ppd.HeaderType, ppd.BodySize = ppd.header.TypeAndPayloadSize()

	// init hmac hash -> HmacHash0
	ppd.hmacHash, err = NewHash(ppd.Ciphers.HashType)
	if err != nil {
		return nil, fmt.Errorf("failed to create HMAC hash: %w", err)
	}
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
			msgType := HeaderTypeToString(ppd.HeaderType)
			log.Error("HMAC validation failed on server side. msgType=%s, sumCookie=%v, cipherScheme=%d, bodySize=%d, devicePubKey=%s",
				msgType, sumCookie, ppd.CipherScheme, ppd.BodySize,
				base64.StdEncoding.EncodeToString(ppd.deviceEcdh.PublicKey()))
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
	ppd.BodyCompress = ppd.HeaderFlag&common.NHP_FLAG_COMPRESS != 0
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
	mad.header = mad.BasePacket.HeaderWithCipherScheme(mad.CipherScheme)
	mad.deviceEcdh = mad.device.GetEcdhByCipherScheme(mad.CipherScheme)

	// continue with the sender's counter
	mad.header.SetCounter(ppd.SenderTrxId)

	// init chain hash -> ChainHash0
	var err error
	mad.chainHash, err = NewHash(mad.ciphers.HashType)
	if err != nil {
		// This should never happen with valid CipherSuite parameters
		panic(fmt.Sprintf("failed to create chain hash in deriveMsgAssemblerData: %v", err))
	}
	mad.chainHash.Write([]byte(InitialHashString))

	// continue with responder's chain key
	// Note: ppd.chainKey is cleared by decryptBody's defer, so this is
	// effectively all-zeros.  The Go agent's initiator side has the same
	// behavior (encryptBody clears mad.chainKey), so both sides match.
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
	case common.CIPHER_SCHEME_CURVE:
		fallthrough
	case common.CIPHER_SCHEME_GMSM:
		fallthrough
	default:
		aead, err = AeadFromKey(ppd.Ciphers.GcmType, &key)
		if err != nil {
			log.Error("failed to create AEAD for peer pubkey decryption: %v", err)
			return err
		}
		_, err = aead.Open(peerPk[:0], ppd.header.NonceBytes(), ppd.header.StaticBytes(), ppd.chainHash.Sum(nil))
		if err != nil {
			log.Error("failed to decrypt peer pubkey")
			return err
		}
	}

	if ppd.CipherScheme == common.CIPHER_SCHEME_CURVE {
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
	case NHP_DB:
		toValidate = !option.DisableDePeerValidation
	case DHP_AGENT:
		toValidate = !option.DisableAgentPeerValidation
	}

	if toValidate {
		peerPkBase64 := base64.StdEncoding.EncodeToString(peerPk)
		peerDeviceTypeName := DeviceTypeToString(peerDeviceType)
		log.Debug("validatePeer: looking up %s peer pubkey=%s in peer pool", peerDeviceTypeName, peerPkBase64)

		peer = ppd.device.LookupPeer(peerPk)
		if peer == nil {
			log.Error("validatePeer: %s peer not found in peer pool, pubkey=%s",
				peerDeviceTypeName, peerPkBase64)
			err = fmt.Errorf("peer not found in peer pool (type=%s, pubkey=%s)", peerDeviceTypeName, peerPkBase64)
			return err
		}

		if peer.IsExpired() {
			log.Error("validatePeer: %s peer expired, pubkey=%s", peerDeviceTypeName, peerPkBase64)
			err = fmt.Errorf("peer expired (type=%s, pubkey=%s)", peerDeviceTypeName, peerPkBase64)
			return err
		}

		if !ppd.ConnData.CheckRecvAddress(ppd.LocalInitTime, ppd.ConnData.RemoteAddr) {
			log.Error("validatePeer: %s peer address mismatch on connection, pubkey=%s, remoteAddr=%s",
				peerDeviceTypeName, peerPkBase64, ppd.ConnData.RemoteAddr)
			err = fmt.Errorf("peer does not match its previous address on this connection (type=%s, pubkey=%s)", peerDeviceTypeName, peerPkBase64)
			return err
		}
		ppd.ConnData.UpdateRecvAddress(ppd.LocalInitTime, ppd.ConnData.RemoteAddr)
		peer.UpdateRecv(ppd.LocalInitTime)
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
	case common.CIPHER_SCHEME_CURVE:
		fallthrough
	case common.CIPHER_SCHEME_GMSM:
		fallthrough
	default:
		aead, err = AeadFromKey(ppd.Ciphers.GcmType, &key)
		if err != nil {
			log.Error("failed to create AEAD for timestamp decryption: %v", err)
			return err
		}
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

	// handle knock packet at overload before going into body decryption.
	// sendCookie derives the cookie statelessly from the device's signing
	// key and the remote ip:port + time window, so there is nothing to
	// pre-generate or store per connection.
	if ppd.device.deviceType == NHP_SERVER && ppd.Overload && (ppd.HeaderType == NHP_KNK || ppd.HeaderType == DHP_KNK) {
		ppd.sendCookie()
		err = ErrServerRejectWithCookie
		return err
	}

	// evolve chainhash ChainHash2 -> ChainHash3
	ppd.chainHash.Write(ppd.header.TimestampBytes())

	// generate gcm key for body decryption ChainKey3 -> ChainKey4 (ChainKey7 -> ChainKey8)
	ppd.noise.KeyGen2(&ppd.chainKey, &key, ppd.chainKey[:], ppd.header.TimestampBytes())
	ppd.bodyAead, err = AeadFromKey(ppd.Ciphers.GcmType, &key)
	if err != nil {
		log.Error("failed to create AEAD for body decryption: %v", err)
		return err
	}

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
		// decompress with size limit to prevent decompression bombs
		var buf bytes.Buffer
		br := bytes.NewReader(body)
		r, err := zlib.NewReader(br)
		if err != nil {
			log.Critical("invalid compressed data: %v", err)
			ErrDataDecompressionFailed.SetExtraError(err)
			return ErrDataDecompressionFailed
		}
		defer r.Close()

		// Limit decompressed size to 10MB to prevent DoS via decompression bomb
		const maxDecompressedSize = 10 * 1024 * 1024
		limitedReader := io.LimitReader(r, maxDecompressedSize+1) // +1 to detect overflow
		n, err := io.Copy(&buf, limitedReader)
		if err != nil {
			log.Critical("message decompression failed: %v", err)
			ErrDataDecompressionFailed.SetExtraError(err)
			return ErrDataDecompressionFailed
		}
		if n > maxDecompressedSize {
			log.Critical("decompressed data exceeds maximum size limit (%d bytes)", maxDecompressedSize)
			ErrDataDecompressionFailed.SetExtraError(fmt.Errorf("decompressed size %d exceeds limit %d", n, maxDecompressedSize))
			return ErrDataDecompressionFailed
		}

		ppd.BodyMessage = buf.Bytes() // separately allocated memory
		//log.Debug("message decompressed %v -> %v", body, ppd.BodyMessage)
	} else {
		ppd.BodyMessage = append(ppd.BodyMessage, body...) // deep copy
	}

	return nil
}

func (ppd *PacketParserData) sendCookie() {
	// Only NHP_SERVER reaches this path (the call site in validatePeer is
	// gated on deviceType == NHP_SERVER). Server startup guarantees a
	// signing key is installed — either operator-supplied for clusters,
	// or randomly generated at process start for single-instance
	// deployments — so an empty key here is a programmer error.
	key, win := ppd.device.StatelessCookieParams()
	if len(key) == 0 || win <= 0 {
		log.Critical("sendCookie: device has no stateless cookie params; dropping cookie response")
		return
	}
	window := time.Now().Unix() / win
	// sendCookie runs only from validatePeer after the ECDH+AEAD step
	// has populated ppd.RemotePubKey, so binding the cookie to the
	// agent's static pubkey here is safe — no extra crypto, just one
	// more HMAC.Write of a value we already have.
	cookie := deriveStatelessCookie(key, cookieRemoteKey(ppd.ConnData), ppd.RemotePubKey, window)
	cokStr := base64.StdEncoding.EncodeToString(cookie)
	cokMsg := &common.ServerCookieMsg{
		TransactionId: ppd.SenderTrxId,
		Cookie:        cokStr,
	}
	cokBytes, _ := json.Marshal(cokMsg)

	md := &MsgData{
		HeaderType:    NHP_COK,
		CipherScheme:  ppd.CipherScheme,
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

	headerPrefixLen := ppd.header.Size() - HashSize
	ppd.hmacHash.Write(ppd.header.Bytes()[0:headerPrefixLen])

	if sumCookie {
		// Re-derive the cookie this server (or any sibling sharing the
		// same signing key) would have minted for this remote endpoint
		// in the current + previous time window, and accept either. Two
		// windows cover both the slow agent that takes a moment to
		// round-trip and the window-boundary case where a cookie minted
		// just before the rollover gets used just after.
		//
		// Only reachable on NHP_SERVER under Overload (validatePeer
		// gates the sumCookie=true case behind those two conditions);
		// startup guarantees a signing key is installed, so missing
		// params here is a programmer error.
		key, win := ppd.device.StatelessCookieParams()
		if len(key) == 0 || win <= 0 {
			log.Critical("checkHMAC(sumCookie): device has no stateless cookie params")
			return false
		}
		// Cookies are bound to the initiator's static public key so
		// agents behind a shared NAT/CGN can't replay each other's
		// cookies (see deriveStatelessCookie docstring). validatePeer
		// is what normally fills ppd.RemotePubKey, but at this point
		// in the parse flow it hasn't run yet — so re-do the
		// IK static-decryption locally. One extra ECDH per
		// RKN-under-overload packet; harmless under load (cookie
		// path is the cheap rejector, ECDH cost is bounded by
		// agents that can already mint a legitimate KNK).
		peerPk, err := extractInitiatorStaticPubKey(ppd.device, ppd.Ciphers, ppd.header)
		if err != nil {
			log.Debug("checkHMAC(sumCookie): cannot recover peer static pubkey: %v", err)
			return false
		}
		// The agent computes its HMAC as
		//   Hash(InitialHashString || ServerPubKey || header[:len] || cookie)
		// (initiator.go addHMAC, after the Init+PubKey seeding in
		// createMsgAssemblerData / setPeerPublicKey).
		// Reconstruct that exact byte sequence for each candidate cookie
		// — don't try to clone ppd.hmacHash's intermediate state, since
		// hash.Hash isn't guaranteed to support state cloning across all
		// our cipher suites (blake2s/sm3/sha256).
		headerHmac := ppd.header.HMACBytes()
		serverPubKey := ppd.deviceEcdh.PublicKey()
		headerPrefix := ppd.header.Bytes()[0:headerPrefixLen]
		remote := cookieRemoteKey(ppd.ConnData)
		currWindow := time.Now().Unix() / win
		for _, w := range [2]int64{currWindow, currWindow - 1} {
			cookie := deriveStatelessCookie(key, remote, peerPk, w)
			h, err := NewHash(ppd.Ciphers.HashType)
			if err != nil {
				return false
			}
			h.Write([]byte(InitialHashString))
			h.Write(serverPubKey)
			h.Write(headerPrefix)
			h.Write(cookie)
			// Constant-time compare across the two-window candidate
			// loop. The leak surface is narrow (32-byte HMAC, agent
			// already needed authenticated noise-channel access to
			// reach this path) but the cost is one helper call, so
			// take it.
			if subtle.ConstantTimeCompare(h.Sum(nil), headerHmac) == 1 {
				return true
			}
		}
		return false
	} else {
		calculatedHmac := ppd.hmacHash.Sum(nil)
		headerHmac := ppd.header.HMACBytes()
		if !bytes.Equal(calculatedHmac, headerHmac) {
			log.Debug("checkHMAC: mismatch, calculated=%x, header=%x, headerSize=%d, cipherScheme=%d",
				calculatedHmac[:8], headerHmac[:8], ppd.header.Size(), ppd.CipherScheme)
			return false
		}
		return true
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
	case NHP_KNK, DHP_KNK, NHP_RKN, NHP_EXT, NHP_AOL, NHP_ART:
		return true
	default:
		return false
	}
}
