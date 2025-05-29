package de

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/version"
)

var (
	ExeDirPath string
)

type KnockUser struct {
	UserId         string
	OrganizationId string
	UserData       map[string]any
}

type KnockResource struct {
	AuthServiceId string `json:"aspId"`
	ResourceId    string `json:"resId"`
	ServerAddr    string `json:"serverAddr"`
}

func (res *KnockResource) Id() string {
	return res.AuthServiceId + "/" + res.ResourceId
}

type KnockTarget struct {
	sync.Mutex
	KnockResource
	ServerPeer           *core.UdpPeer
	LastKnockSuccessTime time.Time
}

func (kt *KnockTarget) SetResource(res *KnockResource) {
	kt.Lock()
	defer kt.Unlock()

	kt.KnockResource = *res
}

func (kt *KnockTarget) SetServer(peer *core.UdpPeer) {
	kt.Lock()
	defer kt.Unlock()

	kt.ServerPeer = peer
}

func (kt *KnockTarget) Server() *core.UdpPeer {
	kt.Lock()
	defer kt.Unlock()

	return kt.ServerPeer
}

type UdpDevice struct {
	stats struct {
		totalRecvBytes uint64
		totalSendBytes uint64
	}

	config *Config
	log    *log.Logger

	remoteConnectionMutex sync.Mutex
	remoteConnectionMap   map[string]*UdpConn // indexed by remote UDP address

	serverPeerMutex sync.Mutex
	serverPeerMap   map[string]*core.UdpPeer // indexed by server's public key

	device  *core.Device
	wg      sync.WaitGroup
	running atomic.Bool

	signals struct {
		stop             chan struct{}
		serverMapUpdated chan struct{}
	}

	recvMsgCh <-chan *core.PacketParserData
	sendMsgCh chan *core.MsgData
	// one device should serve only one specific user at a time
	registerUserMutex sync.RWMutex
}

type UdpConn struct {
	ConnData *core.ConnectionData
	netConn  *net.UDPConn
}

func (c *UdpConn) Close() {
	c.netConn.Close()
	c.ConnData.Close()
}

/*
dirPath: the path of app or shared library entry point
logLevel: 0: silent, 1: error, 2: info, 3: debug, 4: verbose
*/
func (a *UdpDevice) Start(dirPath string, logLevel int) (err error) {
	common.ExeDirPath = dirPath
	ExeDirPath = dirPath
	// init logger
	a.log = log.NewLogger("NHP-DE", logLevel, filepath.Join(ExeDirPath, "logs"), "device")
	log.SetGlobalLogger(a.log)

	log.Info("=========================================================")
	log.Info("=== NHP-DE %s started                           ===", version.Version)
	log.Info("=== REVISION %s ===", version.CommitId)
	log.Info("=== RELEASE %s                       ===", version.BuildTime)
	log.Info("=========================================================")

	err = a.loadBaseConfig()
	if err != nil {
		return err
	}

	prk, err := base64.StdEncoding.DecodeString(a.config.PrivateKeyBase64)
	if err != nil {
		log.Error("private key parse error %v\n", err)
		return fmt.Errorf("private key parse error %v", err)
	}

	a.device = core.NewDevice(core.NHP_DE, prk, nil)
	if a.device == nil {
		log.Critical("failed to create device %v\n", err)
		return fmt.Errorf("failed to create device %v", err)
	}

	a.remoteConnectionMap = make(map[string]*UdpConn)
	a.serverPeerMap = make(map[string]*core.UdpPeer)

	// load peers
	a.loadPeers()

	a.signals.stop = make(chan struct{})
	a.signals.serverMapUpdated = make(chan struct{}, 1)
	a.recvMsgCh = a.device.DecryptedMsgQueue
	a.sendMsgCh = make(chan *core.MsgData, core.SendQueueSize)

	// start device routines
	a.device.Start()

	// start device routines
	a.wg.Add(2)

	go a.sendMessageRoutine()
	go a.recvMessageRoutine()
	a.running.Store(true)
	// time.Sleep(1000 * time.Millisecond)
	return nil
}

// export Stop
func (a *UdpDevice) Stop() {
	a.running.Store(false)
	close(a.signals.stop)
	a.device.Stop()
	a.StopConfigWatch()
	a.wg.Wait()
	close(a.sendMsgCh)
	log.Info("=========================")
	log.Info("=== NHP-Device stopped ===")
	log.Info("=========================")
	a.log.Close()
}

func (a *UdpDevice) IsRunning() bool {
	return a.running.Load()
}

func (a *UdpDevice) newConnection(addr *net.UDPAddr) (conn *UdpConn) {
	conn = &UdpConn{}
	var err error
	// unlike tcp, udp dial is fast (just socket bind), so no need to run in a thread
	conn.netConn, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Error("could not connect to remote addr %s", addr.String())
		return nil
	}

	// retrieve local port
	laddr := conn.netConn.LocalAddr()
	localAddr, err := net.ResolveUDPAddr(laddr.Network(), laddr.String())
	if err != nil {
		log.Error("resolve local UDPAddr error %v\n", err)
		return nil
	}

	log.Info("Dial up new UDP connection from %s to %s", localAddr.String(), addr.String())

	conn.ConnData = &core.ConnectionData{
		Device:               a.device,
		CookieStore:          &core.CookieStore{},
		RemoteTransactionMap: make(map[uint64]*core.RemoteTransaction),
		LocalAddr:            localAddr,
		RemoteAddr:           addr,
		TimeoutMs:            DefaultConnectionTimeoutMs,
		SendQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
		RecvQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
		BlockSignal:          make(chan struct{}),
		SetTimeoutSignal:     make(chan struct{}),
		StopSignal:           make(chan struct{}),
	}

	conn.ConnData.Add(1)
	go a.recvPacketRoutine(conn)

	return conn
}

func (a *UdpDevice) sendMessageRoutine() {
	defer a.wg.Done()
	defer log.Info("sendMessageRoutine stopped")

	log.Info("sendMessageRoutine started")

	for {
		select {
		case <-a.signals.stop:
			return

		case md, ok := <-a.sendMsgCh:
			if !ok {
				return
			}
			if md == nil || md.RemoteAddr == nil {
				log.Warning("Invalid initiator session starter")
				continue
			}

			addrStr := md.RemoteAddr.String()

			a.remoteConnectionMutex.Lock()
			conn, found := a.remoteConnectionMap[addrStr]
			a.remoteConnectionMutex.Unlock()

			if found {
				md.ConnData = conn.ConnData
			} else {
				conn = a.newConnection(md.RemoteAddr)
				if conn == nil {
					log.Error("Failed to dial to remote address: %s", addrStr)
					continue
				}

				a.remoteConnectionMutex.Lock()
				a.remoteConnectionMap[addrStr] = conn
				a.remoteConnectionMutex.Unlock()

				md.ConnData = conn.ConnData

				// launch connection routine
				a.wg.Add(1)
				go a.connectionRoutine(conn)
			}

			a.device.SendMsgToPacket(md)
		}
	}

}

func (a *UdpDevice) SendPacket(pkt *core.Packet, conn *UdpConn) (n int, err error) {
	defer func() {
		atomic.AddUint64(&a.stats.totalSendBytes, uint64(n))
		atomic.StoreInt64(&conn.ConnData.LastLocalSendTime, time.Now().UnixNano())

		if !pkt.KeepAfterSend {
			a.device.ReleasePoolPacket(pkt)
		}
	}()

	pktType := core.HeaderTypeToString(pkt.HeaderType)
	//log.Debug("Send [%s] packet (%s -> %s): %+v", pktType, conn.ConnData.LocalAddr.String(), conn.ConnData.RemoteAddr.String(), pkt.Content)
	log.Info("Send [%s] packet (%s -> %s), %d bytes", pktType, conn.ConnData.LocalAddr.String(), conn.ConnData.RemoteAddr.String(), len(pkt.Content))
	log.Evaluate("Send [%s] packet (%s -> %s), %d bytes", pktType, conn.ConnData.LocalAddr.String(), conn.ConnData.RemoteAddr.String(), len(pkt.Content))
	return conn.netConn.Write(pkt.Content)
}

func (a *UdpDevice) recvPacketRoutine(conn *UdpConn) {
	addrStr := conn.ConnData.RemoteAddr.String()

	defer conn.ConnData.Done()
	defer log.Debug("recvPacketRoutine for %s stopped", addrStr)

	log.Debug("recvPacketRoutine for %s started", addrStr)

	for {
		select {
		case <-conn.ConnData.StopSignal:
			return

		default:
		}

		// udp recv, blocking until packet arrives or netConn.Close()
		pkt := a.device.AllocatePoolPacket()
		n, err := conn.netConn.Read(pkt.Buf[:])
		if err != nil {
			a.device.ReleasePoolPacket(pkt)
			if n == 0 {
				// udp connection closed, it is not an error
				return
			}
			log.Error("Failed to receive from remote address %s (%v)", addrStr, err)
			continue
		}

		// add total recv bytes
		atomic.AddUint64(&a.stats.totalRecvBytes, uint64(n))

		// check minimal length
		if n < core.HeaderSize {
			a.device.ReleasePoolPacket(pkt)
			log.Error("Received UDP packet from %s is too short, discard", addrStr)
			continue
		}

		pkt.Content = pkt.Buf[:n]
		//log.Trace("receive udp packet (%s -> %s): %+v", conn.ConnData.RemoteAddr.String(), conn.ConnData.LocalAddr.String(), pkt.Content)

		typ, _, err := a.device.RecvPrecheck(pkt)
		msgType := core.HeaderTypeToString(typ)
		log.Info("Receive [%s] packet (%s -> %s), %d bytes", msgType, addrStr, conn.ConnData.LocalAddr.String(), n)
		log.Evaluate("Receive [%s] packet (%s -> %s), %d bytes", msgType, addrStr, conn.ConnData.LocalAddr.String(), n)
		if err != nil {
			a.device.ReleasePoolPacket(pkt)
			log.Warning("Receive [%s] packet (%s -> %s), precheck error: %v", msgType, addrStr, conn.ConnData.LocalAddr.String(), err)
			log.Evaluate("Receive [%s] packet (%s -> %s) precheck error: %v", msgType, addrStr, conn.ConnData.LocalAddr.String(), err)
			continue
		}

		atomic.StoreInt64(&conn.ConnData.LastLocalRecvTime, time.Now().UnixNano())

		conn.ConnData.ForwardInboundPacket(pkt)
	}
}

func (a *UdpDevice) connectionRoutine(conn *UdpConn) {
	addrStr := conn.ConnData.RemoteAddr.String()

	defer a.wg.Done()
	defer log.Debug("Connection routine: %s stopped", addrStr)

	log.Debug("Connection routine: %s started", addrStr)

	// stop receiving packets and clean up
	defer func() {
		a.remoteConnectionMutex.Lock()
		delete(a.remoteConnectionMap, addrStr)
		a.remoteConnectionMutex.Unlock()

		conn.Close()
	}()

	for {
		select {
		case <-a.signals.stop:
			return

		case <-conn.ConnData.SetTimeoutSignal:
			if conn.ConnData.TimeoutMs <= 0 {
				log.Debug("Connection routine closed immediately")
				return
			}

		case <-time.After(time.Duration(conn.ConnData.TimeoutMs) * time.Millisecond):
			// timeout, quit routine
			log.Debug("Connection routine idle timeout")
			return

		case pkt, ok := <-conn.ConnData.SendQueue:
			if !ok {
				return
			}
			if pkt == nil {
				continue
			}
			a.SendPacket(pkt, conn)

		case pkt, ok := <-conn.ConnData.RecvQueue:
			if !ok {
				return
			}
			if pkt == nil {
				continue
			}
			log.Debug("Received udp packet len [%d] from addr: %s\n", len(pkt.Content), addrStr)
			// process keepalive packet
			if pkt.HeaderType == core.NHP_KPL {
				a.device.ReleasePoolPacket(pkt)
				log.Info("Receive [NHP_KPL] message (%s -> %s)", addrStr, conn.ConnData.LocalAddr.String())
				continue
			}
			if a.device.IsTransactionResponse(pkt.HeaderType) {
				// forward to a specific transaction
				transactionId := pkt.Counter()
				transaction := a.device.FindLocalTransaction(transactionId)
				if transaction != nil {
					transaction.NextPacketCh <- pkt
					continue
				}
			}

			pd := &core.PacketData{
				BasePacket: pkt,
				ConnData:   conn.ConnData,
				InitTime:   atomic.LoadInt64(&conn.ConnData.LastLocalRecvTime),
			}
			// generic receive
			a.device.RecvPacketToMsg(pd)

		case <-conn.ConnData.BlockSignal:
			log.Critical("blocking address %s", addrStr)
			return
		}
	}
}

func (a *UdpDevice) recvMessageRoutine() {
	defer a.wg.Done()
	defer log.Info("recvMessageRoutine stopped")

	log.Info("recvMessageRoutine started")

	for {
		select {
		case <-a.signals.stop:
			return
		case ppd, ok := <-a.recvMsgCh:
			log.Debug("recvMessageRoutine ppd.HeaderType:%d", ppd.HeaderType)
			if !ok {
				return
			}
			if ppd == nil {
				continue
			}
		}
	}
}

func (a *UdpDevice) AddServer(server *core.UdpPeer) {
	if server.DeviceType() == core.NHP_SERVER {
		a.device.AddPeer(server)
		a.serverPeerMutex.Lock()
		a.serverPeerMap[server.PublicKeyBase64()] = server
		a.serverPeerMutex.Unlock()
	}
}

func (a *UdpDevice) RemoveServer(serverKey string) {
	a.serverPeerMutex.Lock()
	delete(a.serverPeerMap, serverKey)
	a.serverPeerMutex.Unlock()
}

// get first server
func (a *UdpDevice) GetServerPeer() (serverPeer *core.UdpPeer) {
	for key, value := range a.serverPeerMap {
		fmt.Println("Key:", key, "Value:", value)
		serverPeer = value
		return serverPeer
	}
	return nil
}
func (a *UdpDevice) SendDHPRegister(doId string, policy common.DHPPolicy, dataKey string) {
	log.Debug("DHP started")
	serverPeer := a.GetServerPeer()
	wrappedKey := dataKey
	kaoContent := common.DHPKao{
		KeyWrapper:    "consumer",
		PolicyBinding: "",
		ConsumerId:    policy.ConsumerId,
		WrappedKey:    wrappedKey,
	}

	jsonkaoContent, err := json.Marshal(kaoContent)
	if err != nil {
		log.Error("json parse error:%v", err)
		return
	}
	log.Debug("jsonkaoContent:%s \n", string(jsonkaoContent))
	msg := common.DRGMsg{
		DoType:      DoType_Default,
		DoId:        doId,
		AccessUrl:   "",
		AccessByNHP: false,
		AspHost:     "",
		KasType:     0,
		KaoContent:  string(jsonkaoContent),
		PasType:     2,
		PaoContent:  "",
	}

	log.Debug("serverPeer:%s \n", serverPeer)
	result := a.SendNHPDRG(serverPeer, msg)
	if result {
		fmt.Println("File Encryption & Registration: Successful")
	} else {
		fmt.Println("File Encryption & Registration: Failed")
	}
}

// send NHP_DRG to NHP-Server
func (a *UdpDevice) SendNHPDRG(server *core.UdpPeer, msg common.DRGMsg) bool {

	result := false
	sendAddr := server.SendAddr()
	if sendAddr == nil {
		log.Critical("device(%s)[SendNHPDRG] register server IP cannot be parsed", a)
	}
	drgMsg := msg
	drgBytes, _ := json.Marshal(drgMsg)
	drgMd := &core.MsgData{
		RemoteAddr:    sendAddr.(*net.UDPAddr),
		HeaderType:    core.NHP_DRG,
		CipherScheme:  a.config.DefaultCipherScheme,
		TransactionId: a.device.NextCounterIndex(),
		Compress:      true,
		Message:       drgBytes,
		PeerPk:        server.PublicKey(),
		ResponseMsgCh: make(chan *core.PacketParserData),
	}
	currTime := time.Now().UnixNano()
	if !a.IsRunning() {
		log.Error("server-deviceMsgData channel closed or being closed, skip sending")
		return result
	}
	// device will create or find existing connection and sends the MsgAssembler via that connection
	a.sendMsgCh <- drgMd
	server.UpdateSend(currTime)
	// block until transaction completes
	serverPpd := <-drgMd.ResponseMsgCh
	close(drgMd.ResponseMsgCh)

	//Awaiting response from NHP-Server and processing it in the `func()` function below
	var err error
	result = func() bool {

		if serverPpd.Error != nil {
			log.Error("DE(%s#%d)[SendNHPDRG] failed to receive response from server %s: %v", drgMsg.DoId, drgMd.TransactionId, server.Ip, serverPpd.Error)
			err = serverPpd.Error
			return false
		}

		if serverPpd.HeaderType != core.NHP_DAK {
			log.Error("DE(%s#%d)[SendNHPDRG] response from server %s has wrong type: %s", drgMsg.DoId, drgMd.TransactionId, server.Ip, core.HeaderTypeToString(serverPpd.HeaderType))
			err = common.ErrTransactionRepliedWithWrongType
			return false
		}

		dakMsg := &common.DAKMsg{}
		//json string to DAKMsg Object
		err = json.Unmarshal(serverPpd.BodyMessage, dakMsg)
		if err != nil {
			log.Error("DE(%s#%d)[HandleDHPDRGMessage] failed to parse %s message: %v", drgMsg.DoId, serverPpd.SenderTrxId, core.HeaderTypeToString(serverPpd.HeaderType), err)
			return false
		}
		dakMsgString, err := json.Marshal(dakMsg)
		if err != nil {
			log.Error("DE(%s#%d)DAKMsg failed to parse %s message: %v", dakMsg.DoId, err)
			return false
		}
		log.Info("SendNHPDRG result：%v", string(dakMsgString))
		if dakMsg.ErrCode != 0 {
			log.Error("SendNHPDRG send failed,error:", dakMsg.ErrMsg)
			fmt.Println("SendNHPDRG send failed,error:" + dakMsg.ErrMsg)
			return false
		}
		return true
	}()
	log.Info("SendNHPDRG sent successfully | Returned result:%v", result)
	return result
}
