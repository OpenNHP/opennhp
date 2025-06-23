package db

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	ztdolib "github.com/OpenNHP/opennhp/nhp/core/ztdo"
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

	teeMutex sync.Mutex
	teeMap  map[string]*TEE // indexed by tee's public key

	device  *core.Device
	wg      sync.WaitGroup
	running atomic.Bool

	signals struct {
		stop             chan struct{}
		serverMapUpdated chan struct{}
	}

	recvMsgCh <-chan *core.PacketParserData
	sendMsgCh chan *core.MsgData

	EnableOnlineReport bool
}

type UdpConn struct {
	ConnData *core.ConnectionData
	netConn  *net.UDPConn
	connected    atomic.Bool
	externalAddr string
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
	a.log = log.NewLogger("NHP-DB", logLevel, filepath.Join(ExeDirPath, "logs"), "device")
	log.SetGlobalLogger(a.log)

	log.Info("=========================================================")
	log.Info("=== NHP-DB %s started                           ===", version.Version)
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

	a.device = core.NewDevice(core.NHP_DB, prk, nil)
	if a.device == nil {
		log.Critical("failed to create device %v\n", err)
		return fmt.Errorf("failed to create device %v", err)
	}

	a.remoteConnectionMap = make(map[string]*UdpConn)
	a.serverPeerMap = make(map[string]*core.UdpPeer)

	// load peers
	a.loadPeers()

	// load TEEs
	a.loadTEEs()

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
	if a.EnableOnlineReport{
		a.wg.Add(1)
		go a.maintainServerConnectionRoutine()
	}
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
	close(a.signals.serverMapUpdated)

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
		if n < pkt.MinimalLength() {
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

			switch ppd.HeaderType {
			case core.NHP_DWR:
				// deal with NHP_AOP message
				a.wg.Add(1)
				go a.HandleUdpDataKeyWrappingOperations(ppd)
			}
		}
	}
}

func (a *UdpDevice) maintainServerConnectionRoutine() {
	defer a.wg.Done()
	defer log.Info("maintainServerConnectionRoutine stopped")

	log.Info("maintainServerConnectionRoutine started")


	var discoveryRoutineWg sync.WaitGroup
	defer discoveryRoutineWg.Wait()

	for {
		// make a local copy of servers then iterate because next operations are time consuming (too long to use locked iteration)
		a.serverPeerMutex.Lock()
		var serverCount int32 = int32(len(a.serverPeerMap))
		discoveryQuitArr := make([]chan struct{}, 0, serverCount)

		for _, server := range a.serverPeerMap {
			// launch discovery routine for each server
			fail := new(int32)
			quit := make(chan struct{})
			discoveryQuitArr = append(discoveryQuitArr, quit)

			discoveryRoutineWg.Add(1)
			go a.serverDiscovery(server, &discoveryRoutineWg, fail, quit)
		}
		a.serverPeerMutex.Unlock()

		select {
		case <-a.signals.stop:
			log.Info("maintainServerConnectionRoutine receives stop signal")
			return
		case _, ok := <-a.signals.serverMapUpdated:
			if !ok {
				return
			}
			// stop all current discovery routines
			for _, q := range discoveryQuitArr {
				close(q)
			}
			// continue and restart with new server discovery cycle
		}
	}
}

func (a *UdpDevice) serverDiscovery(server *core.UdpPeer, discoveryRoutineWg *sync.WaitGroup, serverFailCount *int32, quit <-chan struct{}) {
	defer discoveryRoutineWg.Done()

	dbId := a.config.DbId
	sendAddr := server.SendAddr()
	if sendAddr == nil {
		log.Error("Cannot connect to nil server address")
		return
	}

	addrStr := sendAddr.String()

	defer log.Info("server discovery sub-routine at %s stopped", addrStr)
	log.Info("server discovery sub-routine at %s started", addrStr)

	var failCount int

	for {
		var lastSendTime int64
		var lastRecvTime int64
		var connected bool

		// find whether connection is already connected
		a.remoteConnectionMutex.Lock()
		conn, found := a.remoteConnectionMap[addrStr]
		a.remoteConnectionMutex.Unlock()

		if found {
			// connection based timing
			lastSendTime = atomic.LoadInt64(&conn.ConnData.LastLocalSendTime)
			lastRecvTime = atomic.LoadInt64(&conn.ConnData.LastLocalRecvTime)
			connected = conn.connected.Load()
		} else {
			// peer based timing
			conn = nil
			lastSendTime = server.LastSendTime()
			lastRecvTime = server.LastRecvTime()
		}

		currTime := time.Now().UnixNano()
		peerPbk := server.PublicKey()

		// when a server is not connected, try to connect in every ACLocalTransactionResponseTimeoutMs
		// when a server is connected when ServerConnectionInterval is reached since last receive, try resend NHP_AOL for maintaining server connection
		if !connected || (currTime-lastRecvTime) > int64(ReportToServerInterval*time.Second) {
			// send NHP_AOL message to server
			aolMsg := &common.DBOnlineMsg{
				DBId:          dbId,
			}
			aolBytes, _ := json.Marshal(aolMsg)

			aolMd := &core.MsgData{
				RemoteAddr:    sendAddr.(*net.UDPAddr),
				HeaderType:    core.NHP_DOL,
				CipherScheme:  a.config.DefaultCipherScheme,
				TransactionId: a.device.NextCounterIndex(),
				Compress:      true,
				PeerPk:        peerPbk,
				Message:       aolBytes,
				ResponseMsgCh: make(chan *core.PacketParserData),
			}

			if !a.IsRunning() {
				log.Error("db(%s#%d)[DBOnline] MsgData channel closed or being closed, skip sending", dbId, aolMd.TransactionId)
				return
			}

			a.sendMsgCh <- aolMd // create new connection
			server.UpdateSend(currTime)

			// block until transaction completes or timeouts
			ppd := <-aolMd.ResponseMsgCh
			close(aolMd.ResponseMsgCh)

			var err error
			func() {
				defer func() {
					if err != nil {
						if conn != nil {
							conn.connected.Store(false)
						}

						failCount += 1
						if failCount%ServerDiscoveryRetryBeforeFail == 0 {
							atomic.StoreInt32(serverFailCount, 1)
							// remove failed connection
							a.remoteConnectionMutex.Lock()
							conn = a.remoteConnectionMap[addrStr]
							if conn != nil {
								log.Info("server discovery failed, close local connection: %s", conn.ConnData.LocalAddr.String())
								delete(a.remoteConnectionMap, addrStr)
							}
							a.remoteConnectionMutex.Unlock()
							conn.Close()
						}
						log.Error("db(%s#%d)[DBOnline] reporting to server %s failed", dbId, aolMd.TransactionId, addrStr)
					}
				}()

				if ppd.Error != nil {
					log.Error("db(%s#%d)[DBOnline] failed to receive response from server %s: %v", dbId, aolMd.TransactionId, addrStr, ppd.Error)
					err = ppd.Error
					return
				}

				if ppd.HeaderType != core.NHP_DBA {
					log.Error("db(%s#%d)[DBOnline] response from server %s has wrong type: %s", dbId, aolMd.TransactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType))
					err = common.ErrTransactionRepliedWithWrongType
					return
				}

				aakMsg := &common.ServerDBAckMsg{}
				err = json.Unmarshal(ppd.BodyMessage, aakMsg)
				if err != nil {
					log.Error("db(%s#%d)[HandleDBAck] failed to parse %s message: %v", dbId, ppd.SenderTrxId, core.HeaderTypeToString(ppd.HeaderType), err)
					return
				}

				// server discovery succeeded
				failCount = 0
				atomic.StoreInt32(serverFailCount, 0)
				a.remoteConnectionMutex.Lock()
				conn = a.remoteConnectionMap[addrStr] // conn must be available at this point
				conn.connected.Store(true)
				conn.externalAddr = aakMsg.DBAddr
				a.remoteConnectionMutex.Unlock()
				log.Info("db(%s#%d)[DBOnline] succeed. db external address is %s, replied by server %s", dbId, aolMd.TransactionId, aakMsg.DBAddr, addrStr)
			}()

		}  else if connected {
			if (currTime - lastSendTime) > int64(ServerKeepaliveInterval*time.Second) {
				// send NHP_KPL to server if no send happens within ServerKeepaliveInterval
				md := &core.MsgData{
					RemoteAddr:   sendAddr.(*net.UDPAddr),
					HeaderType:   core.NHP_KPL,
					CipherScheme: a.config.DefaultCipherScheme,
					//PeerPk:        peerPbk, // pubkey not needed
					TransactionId: a.device.NextCounterIndex(),
				}

				a.sendMsgCh <- md // send NHP_KPL to server via existing connection
				server.UpdateSend(currTime)
			}
		}

		select {
		case <-a.signals.stop:
			log.Info("server discovery sub-routine at %s receives stop signal", addrStr)
			return
		case <-quit:
			return
		case <-time.After(MinialServerDiscoveryInterval * time.Second):
			// wait for ServerConnectionDiscoveryInterval
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
	for _, value := range a.serverPeerMap {
		serverPeer = value
		return serverPeer
	}
	return nil
}
func (a *UdpDevice) SendDHPRegister(msg common.DRGMsg) {
	log.Debug("DHP started")
	serverPeer := a.GetServerPeer()

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
			log.Error("DB(%s#%d)[SendNHPDRG] failed to receive response from server %s: %v", drgMsg.DoId, drgMd.TransactionId, server.Ip, serverPpd.Error)
			err = serverPpd.Error
			return false
		}

		if serverPpd.HeaderType != core.NHP_DAK {
			log.Error("DB(%s#%d)[SendNHPDRG] response from server %s has wrong type: %s", drgMsg.DoId, drgMd.TransactionId, server.Ip, core.HeaderTypeToString(serverPpd.HeaderType))
			err = common.ErrTransactionRepliedWithWrongType
			return false
		}

		dakMsg := &common.DAKMsg{}
		//json string to DAKMsg Object
		err = json.Unmarshal(serverPpd.BodyMessage, dakMsg)
		if err != nil {
			log.Error("DB(%s#%d)[HandleDHPDRGMessage] failed to parse %s message: %v", drgMsg.DoId, serverPpd.SenderTrxId, core.HeaderTypeToString(serverPpd.HeaderType), err)
			return false
		}
		dakMsgString, err := json.Marshal(dakMsg)
		if err != nil {
			log.Error("DB(%s#%d)DAKMsg failed to parse %s message: %v", dakMsg.DoId, err)
			return false
		}
		log.Info("SendNHPDRG result：%v", string(dakMsgString))
		if dakMsg.ErrCode != 0 {
			log.Error("SendNHPDRG send failed,error:", dakMsg.ErrMsg)
			return false
		}
		return true
	}()
	log.Info("SendNHPDRG sent successfully | Returned result:%v", result)
	return result
}

func (a *UdpDevice) GetCipherSchema() int {
	return a.config.DefaultCipherScheme
}

func (a *UdpDevice) GetSymmetricCipherMode() string {
	return a.config.SymmetricCipherMode
}

func (a *UdpDevice) GetDataBrokerId() string {
	return a.config.DbId
}

func (a *UdpDevice) GetOwnEcdh() core.Ecdh {
	prk, _ := base64.StdEncoding.DecodeString(a.config.PrivateKeyBase64)
	eccMode := core.ECC_CURVE25519
	if a.config.DefaultCipherScheme == 0 {
		eccMode = core.ECC_SM2
	}

	return core.ECDHFromKey(eccMode, prk)
}

func (a *UdpDevice) isTEEAuthorized(teePbkBase64 string) bool {
	a.teeMutex.Lock()
	defer a.teeMutex.Unlock()
	if tee, found := a.teeMap[teePbkBase64]; found {
		return tee.ExpireTime > 0 && time.Now().UnixMilli() < tee.ExpireTime*1000
	}

	return false
}

func (a *UdpDevice) HandleUdpDataKeyWrappingOperations(ppd *core.PacketParserData) (err error) {
	defer a.wg.Done()

	dbId := a.config.DbId

	dwrMsg := &common.DWRMsg{}
	dwaMsg := &common.DWAMsg{}

	transactionId := ppd.SenderTrxId
	err = json.Unmarshal(ppd.BodyMessage, dwrMsg)
	if err == nil {
		if !a.isTEEAuthorized(dwrMsg.TeePublicKey) {
			errCode, _ := strconv.Atoi(common.ErrTEENotAuthorized.ErrorCode())
			dwaMsg.ErrCode = errCode
			dwaMsg.ErrMsg = common.ErrTEENotAuthorized.Error()
		} else {
			dataPrkStore, err := NewDataPrivateKeyStoreWith(dwrMsg.DoId)
			if err != nil {
				errCode, _ := strconv.Atoi(common.ErrDataPrivateKeyStore.ErrorCode())
				dwaMsg.ErrCode = errCode
				dwaMsg.ErrMsg = common.ErrDataPrivateKeyStore.Error()
			} else {
				dataKeyPairEccMode := ztdolib.SM2
				if a.config.DefaultCipherScheme == common.CIPHER_SCHEME_CURVE {
					dataKeyPairEccMode = ztdolib.CURVE25519
				}

				teePbk, _ := base64.StdEncoding.DecodeString(dwrMsg.TeePublicKey)
				consumerEPbk, _ := base64.StdEncoding.DecodeString(dwrMsg.ConsumerEphemeralPublicKey)


				sa := ztdolib.NewSymmetricAgreement(dataKeyPairEccMode, true)
				sa.SetMessagePatterns(ztdolib.DataPrivateKeyWrappingPatterns)
				sa.SetPsk([]byte(ztdolib.InitialDHPKeyWrappingString))
				sa.SetStaticKeyPair(a.GetOwnEcdh())
				sa.SetRemoteStaticPublicKey(teePbk)
				sa.SetRemoteEphemeralPublicKey(consumerEPbk)

				gcmKey, ad := sa.AgreeSymmetricKey()

				dataPrkWrapping := ztdolib.NewDataPrivateKeyWrapping(dataPrkStore.ProviderPublicKeyBase64, dataPrkStore.DataPrivateKeyBase64, gcmKey[:], ad)

				dataPrkWrappingJson, _ := json.Marshal(dataPrkWrapping)

				kao := common.KeyAccessObject{
					WrappedDataKey: string(dataPrkWrappingJson),
				}
				dwaMsg.Kao = &kao
				dwaMsg.DoId = dwrMsg.DoId
			}
		}
	} else {
		log.Error("db(%s#%d)[HandleUdpDataKeyWrappingOperations] failed to parse %s message: %v", dbId, transactionId, core.HeaderTypeToString(ppd.HeaderType), err)
		errCode, _ := strconv.Atoi(common.ErrJsonParseFailed.ErrorCode())
		dwaMsg.ErrCode = errCode
		dwaMsg.ErrMsg = err.Error()
	}

	dwaBytes, _ := json.Marshal(dwaMsg)
	md := &core.MsgData{
		HeaderType:     core.NHP_DWA,
		TransactionId:  transactionId,
		Compress:       true,
		PrevParserData: ppd,
		Message:        dwaBytes,
	}

	// forward to a specific transaction
	transaction := ppd.ConnData.FindRemoteTransaction(transactionId)
	if transaction == nil {
		log.Error("db(%s#%d)[HandleUdpDataKeyWrappingOperations] transaction is not available", dbId, transactionId)
		err = common.ErrTransactionIdNotFound
		return err
	}

	transaction.NextMsgCh <- md

	return err
}