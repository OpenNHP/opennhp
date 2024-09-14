package core

import (
	"encoding/base64"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/log"
)

type DeviceTypeEnum = int

const (
	NHP_NO_DEVICE = iota
	NHP_AGENT
	NHP_SERVER
	NHP_AC
	NHP_RELAY
)

type DeviceOptions struct {
	DisableAgentPeerValidation  bool
	DisableServerPeerValidation bool
	DisableACPeerValidation     bool
	DisableRelayPeerValidation  bool
}

type NhpError interface {
	Error() string
	ErrorCode() string
	ErrorNumber() int
}

func defaultDeviceOptions(t int) (option DeviceOptions) {
	switch t {
	case NHP_AGENT:
	case NHP_SERVER:
	case NHP_AC:
		// NHP_AC does not validate, nor store any agent peer. Related message: NHP-ACC (agent-ac pre-access)
		option.DisableAgentPeerValidation = true
	case NHP_RELAY:
	}

	return option
}

type Device struct {
	optionMutex sync.Mutex
	option      DeviceOptions

	counterIndex    uint64
	deviceType      int
	staticEcdhCurve Ecdh // for cipherscheme curve
	staticEcdhGmsm  Ecdh // for cipherscheme gmsm

	peerMapMutex sync.Mutex
	peerMap      map[string]Peer

	localTransactionMutex sync.Mutex
	localTransactionMap   map[uint64]*LocalTransaction

	pool     *PacketBufferPool
	Overload atomic.Bool

	wg      sync.WaitGroup
	signals struct {
		stop chan struct{}
	}

	DecryptedMsgQueue chan *PacketParserData
	packetToMsgQueue  chan *PacketData
	msgToPacketQueue  chan *MsgData
}

func NewDevice(t int, prk []byte, option *DeviceOptions) *Device {
	d := &Device{
		deviceType: t,
	}

	if option != nil {
		d.option = *option
	} else {
		d.option = defaultDeviceOptions(t)
	}

	d.staticEcdhCurve = ECDHFromKey(ECC_CURVE25519, prk)
	if d.staticEcdhCurve == nil {
		log.Critical("Failed to set private key")
		return nil
	}
	d.staticEcdhGmsm = ECDHFromKey(ECC_SM2, prk)
	if d.staticEcdhGmsm == nil {
		log.Critical("Failed to set private key ex")
		return nil
	}

	d.pool = &PacketBufferPool{}
	d.pool.Init(PacketBufferPoolSize)

	d.peerMap = make(map[string]Peer)
	d.localTransactionMap = make(map[uint64]*LocalTransaction)

	d.DecryptedMsgQueue = make(chan *PacketParserData, RecvQueueSize)
	d.msgToPacketQueue = make(chan *MsgData, SendQueueSize)
	d.packetToMsgQueue = make(chan *PacketData, RecvQueueSize)
	d.signals.stop = make(chan struct{})

	return d
}

func (d *Device) SetOption(option DeviceOptions) {
	d.optionMutex.Lock()
	defer d.optionMutex.Unlock()

	d.option = option
}

func (d *Device) Start() {
	cpus := runtime.NumCPU()
	d.wg.Add(2 * cpus)
	for i := 0; i < cpus; i++ {
		go d.msgToPacketRoutine(i)
		go d.packetToMsgRoutine(i)
	}
}

func (d *Device) Stop() {
	close(d.signals.stop)
	d.wg.Wait()
	close(d.msgToPacketQueue)
	close(d.packetToMsgQueue)
	close(d.DecryptedMsgQueue)
}

func (d *Device) PublicKeyBase64() string {
	return d.staticEcdhCurve.PublicKeyBase64()
}

func (d *Device) PublicKeyExBase64() string {
	return d.staticEcdhGmsm.PublicKeyBase64()
}

func (d *Device) NextCounterIndex() uint64 {
	return atomic.AddUint64(&d.counterIndex, 1)
}

// Asynchronous multi-channel processing.
func (d *Device) msgToPacketRoutine(id int) {
	defer d.wg.Done()
	defer log.Info("msgToPacketRoutine %d: quit", id)

	log.Info("msgToPacketRoutine %d: start", id)

	for {
		select {
		case <-d.signals.stop:
			return

		case md, ok := <-d.msgToPacketQueue:
			if !ok {
				return
			}
			if md == nil {
				log.Warning("msgToPacketRoutine %d: msgToPacketRoutine gets nil data", id)
				continue
			}

			// message encryption workflow: raw message -> encryption -> raw packet -> connection.SendQueue
			func() {
				msgType := HeaderTypeToString(md.HeaderType)
				var msgStr string
				if md.Message != nil {
					msgStr = string(md.Message)
				}
				log.Debug("msgToPacketRoutine %d: encrypting [%s] raw message: %s", id, msgType, msgStr)
				log.Evaluate("msgToPacketRoutine %d: encrypting [%s] raw message: %s", id, msgType, msgStr)

				var mad *MsgAssemblerData
				var err error

				// error handling
				defer func() {
					if err != nil {
						mad.Error = err
						mad.Destroy()

						// inform preset channel with error
						if mad.ResponseMsgCh != nil {
							mad.ResponseMsgCh <- &PacketParserData{
								Error: err,
							}
						}
						if mad.encryptedPktCh != nil {
							mad.encryptedPktCh <- mad
						}
					}
				}()

				// process keepalive separately
				if md.HeaderType == NHP_KPL {
					mad, _ = d.createKeepalivePacket(md)
					// send out keepalive packet
					mad.connData.ForwardOutboundPacket(mad.BasePacket)
					return
				}

				mad, err = d.createMsgAssemblerData(md)
				// no errors will happen
				if err != nil {
					return
				}

				err = mad.setPeerPublicKey(nil)
				if err != nil {
					log.Error("msgToPacketRoutine %d: [%s] message randomization failed: %v", id, msgType, err)
					log.Evaluate("msgToPacketRoutine %d: [%s] message randomization failed: %v", id, msgType, err)
					return
				}

				err = mad.encryptBody()
				if err != nil {
					log.Error("msgToPacketRoutine %d: [%s] message encryption failed: %v", id, msgType, err)
					log.Evaluate("msgToPacketRoutine %d: [%s] message encryption failed: %v", id, msgType, err)
					return
				}
				log.Debug("msgToPacketRoutine %d: complete encrypting [%s]", id, msgType)
				log.Evaluate("msgToPacketRoutine %d: complete encrypting [%s]", id, msgType)

				// deliver encrypted packet to specific channel, but be sure to release the packet buffer after use
				if mad.encryptedPktCh != nil {
					mad.encryptedPktCh <- mad
					return
				}

				// create local transaction if needed
				if d.IsTransactionRequest(mad.HeaderType) {
					// save initiator transaction
					mad.BasePacket.KeepAfterSend = true // packet is kept after sending and deleted at transaction level
					t := &LocalTransaction{
						transactionId: mad.header.Counter(),
						connData:      mad.connData,
						mad:           mad,
						NextPacketCh:  make(chan *Packet),
						timeout:       d.LocalTransactionTimeout(),
					}
					d.AddLocalTransaction(t)
				}

				// send out fully encrypted packet
				mad.connData.ForwardOutboundPacket(mad.BasePacket)
			}()
		}
	}
}

// Synchronous linear processing.
func (d *Device) MsgToPacket(md *MsgData) (mad *MsgAssemblerData, err error) {
	defer func() {
		if x := recover(); x != nil {
			mad = nil
			err = fmt.Errorf("!!!recovered from panic: %v\n%s", x, string(debug.Stack()))
			ErrRuntimePanic.SetExtraError(err)
			err = ErrRuntimePanic
		}
	}()

	var buf [PacketBufferSize]byte
	md.ExternalPacket = &Packet{
		Buf:        &buf,
		Content:    buf[:],
		HeaderType: md.HeaderType,
	}
	//md.Compress = len(md.Message) > 64 // no gain for compression if size is small
	// use new transaction id if not specified
	if md.TransactionId == 0 {
		md.TransactionId = d.NextCounterIndex()
	}

	// process keepalive separately
	if md.HeaderType == NHP_KPL {
		mad, _ = d.createKeepalivePacket(md)
		return mad, nil
	}

	mad, err = d.createMsgAssemblerData(md)
	defer mad.Destroy()
	if err != nil {
		return nil, err
	}
	err = mad.setPeerPublicKey(nil)
	if err != nil {
		return nil, err
	}
	err = mad.encryptBody()
	if err != nil {
		return nil, err
	}

	return mad, nil
}

// Asynchronous multi-channel processing.
func (d *Device) packetToMsgRoutine(id int) {
	defer d.wg.Done()
	defer log.Info("packetToMsgRoutine %d: quit", id)

	log.Info("packetToMsgRoutine %d: start", id)

	for {
		select {
		case <-d.signals.stop:
			return

		case pd, ok := <-d.packetToMsgQueue:
			if !ok {
				return
			}
			if pd == nil {
				log.Warning("packetToMsgRoutine %d: packetToMsgQueue gets nil data", id)
				continue
			}

			// packet decryption workflow: connection.RecvQueue -> raw packet -> decryption -> raw message
			func() {
				msgType := HeaderTypeToString(pd.BasePacket.HeaderType)
				log.Debug("packetToMsgRoutine %d: decrypting [%s] raw packet", id, msgType)
				log.Evaluate("packetToMsgRoutine %d: decrypting [%s] raw packet", id, msgType)

				var ppd *PacketParserData
				var err error

				// error handling
				defer func() {
					if err != nil {
						ppd.Error = err
						ppd.Destroy()

						// inform preset channel with error
						if ppd.feedbackMsgCh != nil {
							ppd.feedbackMsgCh <- ppd
						}
						if ppd.decryptedMsgCh != nil {
							ppd.decryptedMsgCh <- ppd
						}
					}
				}()

				ppd, err = d.createPacketParserData(pd)
				if err != nil {
					log.Debug("packetToMsgRoutine %d: [%s] packet precheck failed: %v", id, msgType, err)
					log.Evaluate("packetToMsgRoutine %d: [%s] packet precheck failed: %v", id, msgType, err)
					return
				}

				err = ppd.validatePeer()
				if err != nil {
					log.Debug("packetToMsgRoutine %d: [%s] packet validation failed: %v", id, msgType, err)
					log.Evaluate("packetToMsgRoutine %d: [%s] packet validation failed: %v", id, msgType, err)
					return
				}

				err = ppd.decryptBody()
				if err != nil {
					log.Error("packetToMsgRoutine: %d: [%s] packet decryption failed: %v", id, msgType, err)
					log.Evaluate("packetToMsgRoutine: %d: [%s] packet decryption failed: %v", id, msgType, err)
					return
				}

				var msgStr string
				if ppd.BodyMessage != nil {
					msgStr = string(ppd.BodyMessage)
				}
				log.Debug("packetToMsgRoutine: %d: complete decrypting [%s] message: %s", id, msgType, msgStr)
				log.Evaluate("packetToMsgRoutine: %d: complete decrypting [%s] message: %s", id, msgType, msgStr)

				// deliver decrypted message to specific channel
				if ppd.feedbackMsgCh != nil {
					ppd.Destroy()
					ppd.feedbackMsgCh <- ppd
					return
				}
				if ppd.decryptedMsgCh != nil {
					ppd.Destroy()
					ppd.decryptedMsgCh <- ppd
					return
				}

				// start and save responder transaction
				if d.IsTransactionRequest(ppd.HeaderType) {
					t := &RemoteTransaction{
						transactionId: ppd.SenderTrxId,
						connData:      ppd.ConnData,
						parserData:    ppd, // ppd is owned and to be destroyed by transaction
						NextMsgCh:     make(chan *MsgData),
						timeout:       d.RemoteTransactionTimeout(),
					}
					ppd.ConnData.AddRemoteTransaction(t)
				}

				// release packet buffer, but still keep the decrypted message
				ppd.Destroy()
				// deliver decrypted message to generic channel
				select {
				case d.DecryptedMsgQueue <- ppd:

				default:
					// ppd not delivered, set error to destroy the ppd
					log.Critical("packetToMsgRoutine: %d: decryptedMessageCh is full, discarding message", id)
				}
			}()
		}
	}
}

// Synchronous linear processing.
func (d *Device) PacketToMsg(pd *PacketData) (ppd *PacketParserData, err error) {
	defer func() {
		if x := recover(); x != nil {
			ppd = nil
			err = fmt.Errorf("!!!recovered from panic: %v\n%s", x, string(debug.Stack()))
			ErrRuntimePanic.SetExtraError(err)
			err = ErrRuntimePanic
		}
	}()

	var packetType int
	packetType, _, err = d.RecvPrecheck(pd.BasePacket)
	if err != nil {
		return nil, err
	}
	// skip processing keepalive packet
	if packetType == NHP_KPL {
		return &PacketParserData{HeaderType: NHP_KPL}, nil
	}

	pd.InitTime = time.Now().UnixNano()
	ppd, err = d.createPacketParserData(pd)
	defer ppd.Destroy()
	if err != nil {
		return nil, err
	}
	err = ppd.validatePeer()
	if err != nil {
		return nil, err
	}
	err = ppd.decryptBody()
	if err != nil {
		return nil, err
	}

	return ppd, nil
}

func (d *Device) SendMsgToPacket(md *MsgData) {
	select {
	case d.msgToPacketQueue <- md:
		// process encryption and send encrypted packet via connection
	default:
		// discard
		log.Critical("msgToPacketQueue is full, discarding message")
	}
}

func (d *Device) RecvPacketToMsg(pd *PacketData) {
	select {
	case d.packetToMsgQueue <- pd:
		// process decryption and deliver plain text message to DecryptedMessageCh
	default:
		// discard
		log.Critical("packetToMsgQueue is full, discarding packet")
	}
}

func (d *Device) AddPeer(peer Peer) {
	d.peerMapMutex.Lock()
	defer d.peerMapMutex.Unlock()

	d.peerMap[peer.PublicKeyBase64()] = peer
}

func (d *Device) RemovePeer(pubKey string) {
	d.peerMapMutex.Lock()
	defer d.peerMapMutex.Unlock()

	delete(d.peerMap, pubKey)
}

func (d *Device) ResetPeers() {
	d.peerMapMutex.Lock()
	defer d.peerMapMutex.Unlock()

	d.peerMap = make(map[string]Peer)
}

func (d *Device) LookupPeer(pk []byte) Peer {
	pkStr := base64.StdEncoding.EncodeToString(pk)

	d.peerMapMutex.Lock()
	defer d.peerMapMutex.Unlock()

	peer, found := d.peerMap[pkStr]
	if found {
		return peer
	}
	return nil
}

func (d *Device) IsOverload() bool {
	//return true // debug
	return d.Overload.Load()
}

func (d *Device) SetOverload(overloaded bool) {
	d.Overload.Store(overloaded)
}
