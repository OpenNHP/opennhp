package ac

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	ebpflocal "github.com/OpenNHP/opennhp/endpoints/ac/ebpf"
	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/utils"
	"github.com/OpenNHP/opennhp/nhp/utils/ebpf"
	"github.com/OpenNHP/opennhp/nhp/version"
)

var (
	ExeDirPath string
)

type UdpAC struct {
	config     *Config
	httpConfig *HttpConfig
	iptables   *utils.IPTables
	ipset      *utils.IPSet

	stats struct {
		totalRecvBytes uint64
		totalSendBytes uint64
	}

	log *log.Logger

	remoteConnectionMutex sync.Mutex
	remoteConnectionMap   map[string]*UdpConn // indexed by remote UDP address

	serverPeerMutex sync.Mutex
	serverPeerMap   map[string]*core.UdpPeer // indexed by server's public key

	tokenStoreMutex sync.Mutex
	tokenStore      TokenStore

	device     *core.Device
	httpServer *HttpAC
	wg         sync.WaitGroup
	running    atomic.Bool

	signals struct {
		stop             chan struct{}
		serverMapUpdated chan struct{}
	}

	recvMsgCh <-chan *core.PacketParserData
	sendMsgCh chan *core.MsgData
}

type UdpConn struct {
	ConnData     *core.ConnectionData
	netConn      *net.UDPConn
	connected    atomic.Bool
	externalAddr string
}

func (c *UdpConn) Close() {
	if c.netConn != nil {
		c.netConn.Close()
		c.ConnData.Close()
	}
}

/*
dirPath: the path of app or shared library entry point
logLevel: 0: silent, 1: error, 2: info, 3: debug, 4: verbose
*/
func (a *UdpAC) Start(dirPath string, logLevel int) (err error) {
	common.ExeDirPath = dirPath
	ExeDirPath = dirPath
	// init logger
	a.log = log.NewLogger("NHP-AC", logLevel, filepath.Join(ExeDirPath, "logs"), "ac")
	log.SetGlobalLogger(a.log)

	log.Info("=========================================================")
	log.Info("=== NHP-AC %s started                              ===", version.Version)
	log.Info("=== REVISION %s ===", version.CommitId)
	log.Info("=== RELEASE %s                       ===", version.BuildTime)
	log.Info("=========================================================")

	// init config
	err = a.loadBaseConfig()
	if err != nil {
		return err
	}

	// load http config and turn on http server if needed
	a.loadHttpConfig()
	switch a.config.FilterMode {
	case FilterMode_IPTABLES:
		a.iptables, err = utils.NewIPTables()
		if err != nil {
			log.Error("iptables command not found")
			return
		}

		a.ipset, err = utils.NewIPSet(false)
		if err != nil {
			log.Error("ipset command not found")
			return
		}
	case FilterMode_EBPFXDP:
		err = ebpflocal.EbpfEngineLoad()
		if err != nil {
			return err
		}
	default:
		log.Error("[HandleAccessControl] unsupported FilterMode: %d (expected 0=IPTABLES or 1=EBPFXDP)", a.config.FilterMode)
		return
	}

	prk, err := base64.StdEncoding.DecodeString(a.config.PrivateKeyBase64)
	if err != nil {
		log.Error("private key parse error %v\n", err)
		return fmt.Errorf("private key parse error %v", err)
	}

	a.device = core.NewDevice(core.NHP_AC, prk, nil)
	if a.device == nil {
		log.Critical("failed to create device %v\n", err)
		return fmt.Errorf("failed to create device %v", err)
	}

	a.remoteConnectionMap = make(map[string]*UdpConn)
	a.serverPeerMap = make(map[string]*core.UdpPeer)
	a.tokenStore = make(TokenStore)

	// load peers
	a.loadPeers()

	if a.config.FilterMode == FilterMode_EBPFXDP {
		for _, server := range a.config.Servers {
			ebpfHashStr := ebpf.EbpfRuleParams{
				SrcIP: server.Ip,
				DstIP: a.config.DefaultIp,
			}
			log.Info("server ip is %s", server.Ip)
			err = ebpf.EbpfRuleAdd(2, ebpfHashStr, 31536000)
			if err != nil {
				log.Error("[EbpfRuleAdd] add ebpf src: %s dst: %s,  error: %v, protocol: %d, dstport :%d, %v", ebpfHashStr.SrcIP, ebpfHashStr.DstIP, err)
				continue
			}
		}
	}

	a.signals.stop = make(chan struct{})
	a.signals.serverMapUpdated = make(chan struct{}, 1)

	a.recvMsgCh = a.device.DecryptedMsgQueue
	a.sendMsgCh = make(chan *core.MsgData, core.SendQueueSize)

	// start device routines
	a.device.Start()

	// start ac routines
	a.wg.Add(4)
	go a.tokenStoreRefreshRoutine()
	go a.sendMessageRoutine()
	go a.recvMessageRoutine()
	go a.maintainServerConnectionRoutine()

	a.running.Store(true)
	return nil
}

func (ac *UdpAC) Stop() {
	ac.running.Store(false)
	close(ac.signals.stop)

	ac.device.Stop()
	ac.StopConfigWatch()
	ac.wg.Wait()
	close(ac.sendMsgCh)
	close(ac.signals.serverMapUpdated)

	log.Info("==========================")
	log.Info("=== NHP-AC stopped ===")
	log.Info("==========================")
	ac.log.Close()
}

func (a *UdpAC) IsRunning() bool {
	return a.running.Load()
}

func (a *UdpAC) newConnection(addr *net.UDPAddr) (conn *UdpConn) {
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

	// start connection receive routine
	conn.ConnData.Add(1)
	go a.recvPacketRoutine(conn)

	return conn
}

func (a *UdpAC) sendMessageRoutine() {
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

func (a *UdpAC) SendPacket(pkt *core.Packet, conn *UdpConn) (n int, err error) {
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
	log.Evaluate("Send [%s] packet (%s -> %s, %d bytes)", pktType, conn.ConnData.LocalAddr.String(), conn.ConnData.RemoteAddr.String(), len(pkt.Content))
	return conn.netConn.Write(pkt.Content)
}

func (a *UdpAC) recvPacketRoutine(conn *UdpConn) {
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

func (a *UdpAC) connectionRoutine(conn *UdpConn) {
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

func (a *UdpAC) recvMessageRoutine() {
	defer a.wg.Done()
	defer log.Info("recvMessageRoutine stopped")

	log.Info("recvMessageRoutine started")

	for {
		select {
		case <-a.signals.stop:
			return

		case ppd, ok := <-a.recvMsgCh:
			if !ok {
				return
			}
			if ppd == nil {
				continue
			}

			switch ppd.HeaderType {
			case core.NHP_AOP:
				// deal with NHP_AOP message
				a.wg.Add(1)
				go a.HandleUdpACOperations(ppd)
			}
		}
	}
}

// keep interaction between ac and server in certain time interval to keep outwards ip path active
func (a *UdpAC) maintainServerConnectionRoutine() {
	defer a.wg.Done()
	defer log.Info("maintainServerConnectionRoutine stopped")

	log.Info("maintainServerConnectionRoutine started")

	// reset iptables before exiting
	if a.config.FilterMode == FilterMode_IPTABLES {
		defer a.iptables.ResetAllInput()
	}

	var discoveryRoutineWg sync.WaitGroup
	defer discoveryRoutineWg.Wait()

	for {
		// make a local copy of servers then iterate because next operations are time consuming (too long to use locked iteration)
		a.serverPeerMutex.Lock()
		var serverCount int32 = int32(len(a.serverPeerMap))
		discoveryQuitArr := make([]chan struct{}, 0, serverCount)
		discoveryFailStatusArr := make([]*int32, 0, serverCount)

		for _, server := range a.serverPeerMap {
			// launch discovery routine for each server
			fail := new(int32)
			discoveryFailStatusArr = append(discoveryFailStatusArr, fail)
			quit := make(chan struct{})
			discoveryQuitArr = append(discoveryQuitArr, quit)

			discoveryRoutineWg.Add(1)
			go a.serverDiscovery(server, &discoveryRoutineWg, fail, quit)
		}
		a.serverPeerMutex.Unlock()

		// check whether all server discovery failed.
		// If so, open all blocked input
		quitCheck := make(chan struct{})
		discoveryQuitArr = append(discoveryQuitArr, quitCheck)
		discoveryRoutineWg.Add(1)
		go func() {
			defer discoveryRoutineWg.Done()

			for {
				select {
				case <-a.signals.stop:
					return
				case <-quitCheck:
					return
				case <-time.After(MinialServerDiscoveryInterval * time.Second):
					var totalFail int32
					for _, status := range discoveryFailStatusArr {
						totalFail += atomic.LoadInt32(status)
					}

					if totalFail < int32(len(discoveryFailStatusArr)) {
						if a.config.FilterMode == FilterMode_IPTABLES {
							a.iptables.ResetAllInput()
						}
					} else {
						if a.config.FilterMode == FilterMode_IPTABLES {
							a.iptables.AcceptAllInput()
						}
					}
				}
			}
		}()

		select {
		case <-a.signals.stop:
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

func (a *UdpAC) serverDiscovery(server *core.UdpPeer, discoveryRoutineWg *sync.WaitGroup, serverFailCount *int32, quit <-chan struct{}) {
	defer discoveryRoutineWg.Done()

	acId := a.config.ACId
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
			aolMsg := &common.ACOnlineMsg{
				ACId:          acId,
				AuthServiceId: a.config.AuthServiceId,
				ResourceIds:   a.config.ResourceIds,
			}
			aolBytes, _ := json.Marshal(aolMsg)

			aolMd := &core.MsgData{
				RemoteAddr:    sendAddr.(*net.UDPAddr),
				HeaderType:    core.NHP_AOL,
				CipherScheme:  a.config.DefaultCipherScheme,
				TransactionId: a.device.NextCounterIndex(),
				Compress:      true,
				PeerPk:        peerPbk,
				Message:       aolBytes,
				ResponseMsgCh: make(chan *core.PacketParserData),
			}

			if !a.IsRunning() {
				log.Error("ac(%s#%d)[ACOnline] MsgData channel closed or being closed, skip sending", acId, aolMd.TransactionId)
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
						log.Error("ac(%s#%d)[ACOnline] reporting to server %s failed", acId, aolMd.TransactionId, addrStr)
					}

				}()

				if ppd.Error != nil {
					log.Error("ac(%s#%d)[ACOnline] failed to receive response from server %s: %v", acId, aolMd.TransactionId, addrStr, ppd.Error)
					err = ppd.Error
					return
				}

				if ppd.HeaderType != core.NHP_AAK {
					log.Error("ac(%s#%d)[ACOnline] response from server %s has wrong type: %s", acId, aolMd.TransactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType))
					err = common.ErrTransactionRepliedWithWrongType
					return
				}

				aakMsg := &common.ServerACAckMsg{}
				err = json.Unmarshal(ppd.BodyMessage, aakMsg)
				if err != nil {
					log.Error("ac(%s#%d)[HandleACAck] failed to parse %s message: %v", acId, ppd.SenderTrxId, core.HeaderTypeToString(ppd.HeaderType), err)
					return
				}

				// server discovery succeeded
				failCount = 0
				atomic.StoreInt32(serverFailCount, 0)
				a.remoteConnectionMutex.Lock()
				conn = a.remoteConnectionMap[addrStr] // conn must be available at this point
				conn.connected.Store(true)
				conn.externalAddr = aakMsg.ACAddr
				a.remoteConnectionMutex.Unlock()
				log.Info("ac(%s#%d)[ACOnline] succeed. ac external address is %s, replied by server %s", acId, aolMd.TransactionId, aakMsg.ACAddr, addrStr)
			}()

		} else if connected {
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
			return
		case <-quit:
			return
		case <-time.After(MinialServerDiscoveryInterval * time.Second):
			// wait for ServerConnectionDiscoveryInterval
		}
	}
}

func (a *UdpAC) AddServerPeer(server *core.UdpPeer) {
	if server.DeviceType() == core.NHP_SERVER {
		a.device.AddPeer(server)

		a.serverPeerMutex.Lock()
		a.serverPeerMap[server.PublicKeyBase64()] = server
		a.serverPeerMutex.Unlock()

		// renew server connection cycle
		if len(a.signals.serverMapUpdated) == 0 {
			a.signals.serverMapUpdated <- struct{}{}
		}
	}
}

func (a *UdpAC) RemoveServerPeer(serverKey string) {
	a.serverPeerMutex.Lock()
	beforeSize := len(a.serverPeerMap)
	delete(a.serverPeerMap, serverKey)
	afterSize := len(a.serverPeerMap)
	a.serverPeerMutex.Unlock()

	if beforeSize != afterSize {
		// renew server connection cycle
		if len(a.signals.serverMapUpdated) == 0 {
			a.signals.serverMapUpdated <- struct{}{}
		}
	}
}

func (a *UdpAC) GetConfig() *Config {
	return a.config // return  config
}
