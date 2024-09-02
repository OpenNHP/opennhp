package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/log"
	"github.com/OpenNHP/opennhp/nhp"
	"github.com/OpenNHP/opennhp/plugins"
	"github.com/OpenNHP/opennhp/utils"
	"github.com/OpenNHP/opennhp/version"
)

var (
	ExeDirPath string
)

type UdpServer struct {
	stats struct {
		totalRecvBytes uint64
		totalSendBytes uint64
	}

	config     *Config
	httpConfig *HttpConfig
	log        *log.Logger
	listenAddr *net.UDPAddr
	listenConn *net.UDPConn
	localIp    string
	localMac   string

	device     *nhp.Device
	httpServer *HttpServer
	wg         sync.WaitGroup
	running    atomic.Bool

	// connection and remote transaction management

	remoteConnectionMapMutex sync.Mutex
	remoteConnectionMap      map[string]*UdpConn // indexed by remote UDP address

	agentPeerMapMutex sync.Mutex
	agentPeerMap      map[string]*nhp.UdpPeer // indexed by peer's public key base64 string

	acConnectionMapMutex sync.Mutex
	acConnectionMap      map[string]*ACConn // ac connection is indexed by remote IP address

	acPeerMapMutex sync.Mutex
	acPeerMap      map[string]*nhp.UdpPeer // indexed by peer's public key base64 string

	// block address management
	blockAddrMapMutex sync.Mutex
	blockAddrMap      map[string]*BlockAddr // indexed by remote UDP address, need lock for dynamic change

	// address association map
	srcIpAssociatedAddrMapMutex sync.Mutex
	srcIpAssociatedAddrMap      map[string][]*common.NetAddress // indexed by source ip

	// preset asp-resource-address map
	authServiceMapMutex sync.Mutex
	authServiceMap      common.AuthSvcProviderMap // indexed by asp id and then resource id

	// plugin handlers
	pluginHandlerMapMutex sync.Mutex
	pluginHandlerMap      map[string]plugins.PluginHandler

	// signals
	signals struct {
		stop chan struct{}
	}

	recvMsgCh <-chan *nhp.PacketParserData
	sendMsgCh chan *nhp.MsgData
}

type BlockAddr struct {
	addr       *net.UDPAddr
	expireTime time.Time
}

type UdpConn struct {
	ConnData       *nhp.ConnectionData
	isACConnection bool // Immutable. Don't change it after creation. Conn object is also stored in acConnectionMap which is indexed by ACId
}

type ACConn struct {
	ConnData  *nhp.ConnectionData
	ACPeer    *nhp.UdpPeer
	ACId      string
	ServiceId string
	Apps      []string
}

func (c *UdpConn) Close() {
	c.ConnData.Close()
}

/*
dirPath: the path of app or shared library entry point
logLevel: 0: silent, 1: error, 2: info, 3: debug, 4: verbose
*/
// UDP server never actively sends first packet outwards. It only reacts to received packet then sends response.
func (s *UdpServer) Start(dirPath string, logLevel int) (err error) {
	plugins.ExeDirPath = dirPath
	ExeDirPath = dirPath

	// init logger
	s.log = log.NewLogger("NHP-Server", logLevel, filepath.Join(ExeDirPath, "logs"), "server")
	log.SetGlobalLogger(s.log)

	log.Info("=========================================================")
	log.Info("=== NHP-Server %s started              ===", version.Version)
	log.Info("=== REVISION %s ===", version.CommitId)
	log.Info("=== RELEASE %s                       ===", version.BuildTime)
	log.Info("=========================================================")

	// init config
	err = s.loadBaseConfig()
	if err != nil {
		return err
	}

	var netIP net.IP
	if len(s.config.ListenIp) > 0 {
		netIP = net.ParseIP(s.config.ListenIp)
		if netIP == nil {
			log.Error("udp listen ip address is incorrect!")
			return fmt.Errorf("udp listen ip address is incorrect")
		}
	} else {
		netIP = net.IPv4zero // will both listen on ipv4 0.0.0.0:port and ipv6 [::]:port
	}

	s.listenConn, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   netIP,
		Port: s.config.ListenPort,
	})
	if err != nil {
		log.Error("listen error %v\n", err)
		return fmt.Errorf("listen error %v", err)
	}

	// retrieve local port
	laddr := s.listenConn.LocalAddr()
	s.listenAddr, err = net.ResolveUDPAddr(laddr.Network(), laddr.String())
	if err != nil {
		log.Error("resolve local UDPAddr error %v\n", err)
		return fmt.Errorf("resolve UDPAddr error %v", err)
	}

	prk, err := base64.StdEncoding.DecodeString(s.config.PrivateKeyBase64)
	if err != nil {
		log.Error("private key parse error %v\n", err)
		return fmt.Errorf("private key parse error %v", err)
	}

	option := &nhp.DeviceOptions{
		DisableAgentPeerValidation: s.config.DisableAgentValidation,
	}
	s.device = nhp.NewDevice(nhp.NHP_SERVER, prk, option)
	if s.device == nil {
		log.Critical("failed to create device %v\n", err)
		return fmt.Errorf("failed to create device %v", err)
	}

	// retrieve local ip and mac
	s.localIp = utils.GetLocalOutboundAddress().String()
	s.localMac = utils.GetMacAddress(s.localIp)

	// load peers
	s.loadPeers()

	// load http config and turn on http server if needed
	s.loadHttpConfig()

	// load ip associated addresses
	s.loadSourceIps()

	// load asp resources and plugins
	s.pluginHandlerMap = make(map[string]plugins.PluginHandler)
	s.loadResources()

	s.remoteConnectionMap = make(map[string]*UdpConn)
	s.acConnectionMap = make(map[string]*ACConn)
	s.blockAddrMap = make(map[string]*BlockAddr)
	s.signals.stop = make(chan struct{})

	s.recvMsgCh = s.device.DecryptedMsgQueue
	s.sendMsgCh = make(chan *nhp.MsgData, nhp.SendQueueSize)

	// start device routines
	s.device.Start()

	// start server routines
	s.wg.Add(4)
	go s.BlockAddrRefreshRoutine()
	go s.recvPacketRoutine()
	go s.sendMessageRoutine()
	go s.recvMessageRoutine()

	s.running.Store(true)
	return nil
}

func (s *UdpServer) Stop() {
	if !s.running.Load() {
		// already stopped, do nothing
		return
	}
	s.running.Store(false)
	// stop http server first
	if s.httpServer != nil {
		s.httpServer.Stop()
	}
	close(s.signals.stop)
	s.listenConn.Close()
	s.device.Stop()
	s.StopConfigWatch()
	s.wg.Wait()
	close(s.sendMsgCh)
	s.ClosePlugins()

	log.Info("==========================")
	log.Info("=== NHP-Server stopped ===")
	log.Info("==========================")
	s.log.Close()
}

func (s *UdpServer) IsRunning() bool {
	return s.running.Load()
}

func (s *UdpServer) SendPacket(pkt *nhp.UdpPacket, conn *UdpConn) (n int, err error) {
	defer func() {
		atomic.AddUint64(&s.stats.totalSendBytes, uint64(n))
		atomic.StoreInt64(&conn.ConnData.LastLocalSendTime, time.Now().UnixNano())

		if !pkt.KeepAfterSend {
			s.device.ReleaseUdpPacket(pkt)
		}
	}()

	pktType := nhp.HeaderTypeToString(pkt.HeaderType)
	//log.Debug("Send [%s] packet (%s -> %s)", pktType, s.localAddr.String(), conn.ConnData.RemoteAddr.String(), pkt.Packet)
	log.Info("Send [%s] packet (%s -> %s), %d bytes", pktType, s.listenAddr.String(), conn.ConnData.RemoteAddr.String(), len(pkt.Packet))
	log.Evaluate("Send [%s] packet (%s -> %s), %d bytes", pktType, s.listenAddr.String(), conn.ConnData.RemoteAddr.String(), len(pkt.Packet))
	return s.listenConn.WriteToUDP(pkt.Packet, conn.ConnData.RemoteAddr)
}

func (s *UdpServer) recvPacketRoutine() {
	defer s.wg.Done()
	defer log.Debug("recvPacketRoutine stopped")

	log.Debug("recvPacketRoutine started")

	preCheckThreats := make(map[string]int32)

	for {
		select {
		case <-s.signals.stop:
			return

		default:
		}

		// allocate a new packet buffer for every read
		pkt := s.device.AllocateUdpPacket()

		// udp recv, blocking until packet arrives or conn.Close()
		n, remoteAddr, err := s.listenConn.ReadFromUDP(pkt.Buf[:])
		if err != nil {
			s.device.ReleaseUdpPacket(pkt)
			log.Error("Read from UDP error: %v\n", err)
			if n == 0 {
				// listenConn closed
				return
			}
			continue
		}
		addrStr := remoteAddr.String()

		// add total recv bytes
		atomic.AddUint64(&s.stats.totalRecvBytes, uint64(n))

		// check minimal length
		if n < nhp.HeaderSize {
			s.device.ReleaseUdpPacket(pkt)
			log.Error("Received UDP packet from %s is too short, discard", addrStr)
			continue
		}

		// check if it is from blocked address
		if s.IsBlockAddr(remoteAddr) {
			s.device.ReleaseUdpPacket(pkt)
			log.Critical("Remote address %s is being blocked at the moment, discard.", addrStr)
			continue
		}

		recvTime := time.Now().UnixNano()
		pkt.Packet = pkt.Buf[:n]
		//log.Trace("receive udp packet (%s -> %s): %+v", addrStr, s.listenAddr.String(), pkt.Packet)

		typ, _, err := s.device.RecvPrecheck(pkt) // this check also records packet header type
		msgType := nhp.HeaderTypeToString(typ)
		log.Info("Receive [%s] packet (%s -> %s), %d bytes", msgType, addrStr, s.listenAddr.String(), n)
		log.Evaluate("Receive [%s] packet (%s -> %s), %d bytes", msgType, addrStr, s.listenAddr.String(), n)
		if err != nil {
			// threat plus 1
			preCheckThreats[addrStr]++
			if preCheckThreats[addrStr] > PreCheckThreatCountBeforeBlock {
				s.AddBlockAddr(remoteAddr)
			}
			s.device.ReleaseUdpPacket(pkt)
			log.Warning("Receive [%s] packet (%s -> %s), precheck error: %v", msgType, addrStr, s.listenAddr.String(), err)
			log.Evaluate("Receive [%s] packet (%s -> %s) precheck error: %v", msgType, addrStr, s.listenAddr.String(), err)
			continue
		}
		// clear threat
		delete(preCheckThreats, addrStr)

		s.remoteConnectionMapMutex.Lock()
		conn, found := s.remoteConnectionMap[addrStr]
		s.remoteConnectionMapMutex.Unlock()

		if found {
			// existing connection
			atomic.StoreInt64(&conn.ConnData.LastLocalRecvTime, recvTime)
			conn.ConnData.ForwardInboundPacket(pkt)

		} else {
			// create new connection if there is room
			s.remoteConnectionMapMutex.Lock()
			if len(s.remoteConnectionMap) > OverloadConnectionThreshold {
				s.device.SetOverload(true)
			} else if len(s.remoteConnectionMap) >= MaxConcurrentConnection {
				s.remoteConnectionMapMutex.Unlock()
				log.Critical("Reached maximum concurrent connection. Discard new packet from addr: %s\n", addrStr)
				s.device.ReleaseUdpPacket(pkt)
				continue
			}
			s.remoteConnectionMapMutex.Unlock()

			isACConn := pkt.HeaderType == nhp.NHP_AOL
			conn = &UdpConn{
				isACConnection: isACConn,
			}
			// setup new routine for connection
			conn.ConnData = &nhp.ConnectionData{
				InitTime:             recvTime,
				LastLocalRecvTime:    recvTime, // not in multithreaded yet, directly assign value
				Device:               s.device,
				LocalAddr:            s.listenAddr,
				RemoteAddr:           remoteAddr,
				CookieStore:          &nhp.CookieStore{},
				RemoteTransactionMap: make(map[uint64]*nhp.RemoteTransaction),
				TimeoutMs:            DefaultAgentConnectionTimeoutMs,
				SendQueue:            make(chan *nhp.UdpPacket, PacketQueueSizePerConnection),
				RecvQueue:            make(chan *nhp.UdpPacket, PacketQueueSizePerConnection),
				BlockSignal:          make(chan struct{}),
				SetTimeoutSignal:     make(chan struct{}),
				StopSignal:           make(chan struct{}),
			}

			if conn.isACConnection {
				conn.ConnData.TimeoutMs = DefaultACConnectionTimeoutMs
				log.Debug("Received new ac connection from %s", addrStr)
			}
			s.remoteConnectionMapMutex.Lock()
			s.remoteConnectionMap[addrStr] = conn
			s.remoteConnectionMapMutex.Unlock()

			conn.ConnData.RecvQueue <- pkt

			log.Info("Accept new UDP connection from %s to %s", addrStr, s.listenAddr.String())

			// launch connection routine
			s.wg.Add(1)
			go s.connectionRoutine(conn)
		}
	}
}

func (s *UdpServer) connectionRoutine(conn *UdpConn) {
	addrStr := conn.ConnData.RemoteAddr.String()

	defer s.wg.Done()
	defer log.Debug("Connection routine: %s stopped", addrStr)

	log.Debug("Connection routine: %s started", addrStr)

	// stop receiving packets and clean up
	defer func() {
		// Remove old ac connection record
		// Note on server side, before an old ac connection times out, the very ac can send new connections (due to restart or deemed connection failure)
		// and it may come with the same remote ip but a different remote port
		// so make sure the timeout removal here does not delete newer ac connections
		if conn.isACConnection {
			var acToDelete string
			s.acConnectionMapMutex.Lock()
			for acId, acConn := range s.acConnectionMap {
				if acConn.ConnData.Equal(conn.ConnData) {
					acToDelete = acId
					break
				}
			}
			delete(s.acConnectionMap, acToDelete)
			s.acConnectionMapMutex.Unlock()
		}

		// remove the udp conn from remoteConnectionMap
		s.remoteConnectionMapMutex.Lock()
		delete(s.remoteConnectionMap, addrStr)
		if len(s.remoteConnectionMap) <= OverloadConnectionThreshold {
			s.device.SetOverload(false)
		}
		s.remoteConnectionMapMutex.Unlock()

		conn.Close()
	}()

	for {
		select {
		case <-s.signals.stop:
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

		case <-conn.ConnData.BlockSignal:
			s.AddBlockAddr(conn.ConnData.RemoteAddr)
			return

		case pkt, ok := <-conn.ConnData.RecvQueue:
			if !ok {
				return
			}
			if pkt == nil {
				continue
			}
			log.Debug("Received udp packet len [%d] from addr: %s\n", len(pkt.Packet), addrStr)

			// process keepalive packet
			if pkt.HeaderType == nhp.NHP_KPL {
				s.device.ReleaseUdpPacket(pkt)
				log.Info("Receive [NHP_KPL] message (%s -> %s)", addrStr, s.listenAddr.String())
				continue
			}

			if s.device.IsTransactionResponse(pkt.HeaderType) {
				// forward to a specific transaction
				transactionId := pkt.Counter()
				transaction := s.device.FindLocalTransaction(transactionId)
				if transaction != nil {
					transaction.NextPacketCh <- pkt
					continue
				}
			}

			pd := &nhp.PacketData{
				BasePacket: pkt,
				ConnData:   conn.ConnData,
				InitTime:   atomic.LoadInt64(&conn.ConnData.LastLocalRecvTime),
			}
			// generic receive
			s.device.RecvPacketToMsg(pd)

		case pkt, ok := <-conn.ConnData.SendQueue:
			if !ok {
				return
			}
			if pkt == nil {
				continue
			}
			s.SendPacket(pkt, conn)
		}
	}
}

func (s *UdpServer) BlockAddrRefreshRoutine() {
	defer s.wg.Done()
	defer log.Info("BlockedAddrRoutine stopped")

	log.Info("BlockedAddrRoutine started")

	for {
		select {
		case <-s.signals.stop:
			return

		case <-time.After(BlockAddrRefreshRate * time.Second):
			s.RefreshBlockAddr()
		}
	}
}

func (s *UdpServer) IsBlockAddr(addr *net.UDPAddr) bool {
	s.blockAddrMapMutex.Lock()
	defer s.blockAddrMapMutex.Unlock()

	_, found := s.blockAddrMap[addr.String()]
	return found
}

func (s *UdpServer) AddBlockAddr(addr *net.UDPAddr) {
	s.blockAddrMapMutex.Lock()
	defer s.blockAddrMapMutex.Unlock()

	addrStr := addr.String()
	log.Critical("add blocking address %s", addrStr)

	if len(s.blockAddrMap) < MaxConcurrentConnection {
		s.blockAddrMap[addrStr] = &BlockAddr{addr, time.Now().Add(BlockAddrExpireTime * time.Second)}
	} else {
		log.Warning("block address pool is full")
	}
}

func (s *UdpServer) RefreshBlockAddr() {
	s.blockAddrMapMutex.Lock()
	defer s.blockAddrMapMutex.Unlock()

	now := time.Now()
	for k, v := range s.blockAddrMap {
		if v.expireTime.Before(now) {
			delete(s.blockAddrMap, k)
		}
	}
}

func (s *UdpServer) sendMessageRoutine() {
	defer s.wg.Done()
	defer log.Info("sendMessageRoutine stopped")

	log.Info("sendMessageRoutine started")

	for {
		select {
		case <-s.signals.stop:
			return

		case md, ok := <-s.sendMsgCh:
			if !ok {
				return
			}
			if md.PrevParserData != nil && s.device.IsTransactionResponse(md.HeaderType) {
				// forward to a specific transaction
				transaction := md.ConnData.FindRemoteTransaction(md.PrevParserData.SenderId)
				if transaction != nil {
					transaction.NextMsgCh <- md
					continue
				}
			}

			// generic send
			s.device.SendMsgToPacket(md)
		}
	}
}

func (s *UdpServer) recvMessageRoutine() {
	defer s.wg.Done()
	defer log.Info("recvMessageRoutine stopped")

	log.Info("recvMessageRoutine started")

	for {
		select {
		case <-s.signals.stop:
			return

		case ppd, ok := <-s.recvMsgCh:
			if !ok {
				return
			}
			if ppd == nil {
				// recvMsgCh is closed
				continue
			}

			switch ppd.HeaderType {
			case nhp.NHP_KNK, nhp.NHP_RKN, nhp.NHP_EXT:
				// aynchronously process knock messages with ack response
				go s.HandleKnockRequest(ppd)

			case nhp.NHP_AOL:
				// synchronously block and deal with NHP_DOL to ensure future ac messages will be correctly processed. Don't use go routine
				s.HandleACOnline(ppd)

			case nhp.NHP_OTP:
				go s.HandleOTPRequest(ppd)

			case nhp.NHP_REG:
				go s.HandleRegisterRequest(ppd)

			case nhp.NHP_LST:
				go s.HandleListRequest(ppd)
			}
		}
	}
}

func (s *UdpServer) AddAgentPeer(agent *nhp.UdpPeer) {
	if agent.DeviceType() == nhp.NHP_AGENT {
		s.device.AddPeer(agent)
		s.agentPeerMapMutex.Lock()
		s.agentPeerMap[agent.PublicKeyBase64()] = agent
		s.agentPeerMapMutex.Unlock()
	}
}

func (s *UdpServer) AddACPeer(acPeer *nhp.UdpPeer) {
	if acPeer.DeviceType() == nhp.NHP_AC {
		s.device.AddPeer(acPeer)
		s.acPeerMapMutex.Lock()
		s.acPeerMap[acPeer.PublicKeyBase64()] = acPeer
		s.acPeerMapMutex.Unlock()
	}
}

func (s *UdpServer) AddAddressAssociation(srcIp string, addrs []*common.NetAddress) {
	s.srcIpAssociatedAddrMapMutex.Lock()
	s.srcIpAssociatedAddrMap[srcIp] = addrs
	s.srcIpAssociatedAddrMapMutex.Unlock()
}

func (s *UdpServer) RemoveAddressAssociation(srcIp string) {
	s.srcIpAssociatedAddrMapMutex.Lock()
	delete(s.srcIpAssociatedAddrMap, srcIp)
	s.srcIpAssociatedAddrMapMutex.Unlock()
}

func (s *UdpServer) AddAuthService(aspData *common.AuthServiceProviderData) error {
	if len(aspData.AuthSvcId) == 0 {
		return fmt.Errorf("aspId is empty")
	}

	s.authServiceMapMutex.Lock()
	s.authServiceMap[aspData.AuthSvcId] = aspData
	s.authServiceMapMutex.Unlock()

	if len(aspData.PluginPath) > 0 {
		h := plugins.ReadPluginHandler(aspData.PluginPath)
		if h != nil {
			err := s.LoadPlugin(aspData.AuthSvcId, h)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *UdpServer) AddResource(res *common.ResourceData) error {
	if len(res.AuthServiceId) == 0 || len(res.ResourceId) == 0 {
		return fmt.Errorf("aspId or resId is empty")
	}

	s.authServiceMapMutex.Lock()
	aspData, found := s.authServiceMap[res.AuthServiceId]
	if !found {
		s.authServiceMapMutex.Unlock()
		return fmt.Errorf("aspId not found")
	}
	aspData.ResourceGroups[res.ResourceId] = res
	s.authServiceMapMutex.Unlock()

	return nil
}

func (s *UdpServer) ValidatePlugin(h plugins.PluginHandler) bool {
	// placeholder to validate plugin file
	// err = checkSignature(s.Signature())
	// if err != nil {
	//   return false
	// }

	return true
}

func (s *UdpServer) LoadPlugin(pluginId string, h plugins.PluginHandler) error {
	if !s.ValidatePlugin(h) {
		log.Error("Plugin: %s validation failed", pluginId)
		return fmt.Errorf("plugin validation failed")
	}

	s.pluginHandlerMapMutex.Lock()
	oldHandler, found := s.pluginHandlerMap[pluginId]
	s.pluginHandlerMapMutex.Unlock()
	if found {
		oldHandler.Close()
	}

	pluginDirPath := filepath.Join(ExeDirPath, "plugins", pluginId)
	err := h.Init(&plugins.PluginParamsIn{
		PluginDirPath: &pluginDirPath,
		Log:           s.log.NewSubLogger("Plugin["+pluginId+"]", log.LogLevelDebug),
		Hostname:      &s.config.Hostname,
		LocalIp:       &s.localIp,
		LocalMac:      &s.localMac,
	})
	if err != nil {
		log.Error("plugin: %s initialization failed, %v", pluginId, err)
		return err
	}

	ver := h.Version()
	info := h.ExportedData()
	// use info if necessary
	_ = info

	s.pluginHandlerMapMutex.Lock()
	s.pluginHandlerMap[pluginId] = h
	s.pluginHandlerMapMutex.Unlock()

	log.Info("plugin %s loaded successfully to %s", ver, pluginId)
	return nil
}

func (s *UdpServer) ClosePlugins() {
	s.pluginHandlerMapMutex.Lock()
	defer s.pluginHandlerMapMutex.Unlock()

	for id, handler := range s.pluginHandlerMap {
		log.Info("closing plugin: %s", id)
		handler.Close()
	}
}

func (s *UdpServer) FindAuthSvcProvider(aspId string) *common.AuthServiceProviderData {
	s.authServiceMapMutex.Lock()
	defer s.authServiceMapMutex.Unlock()

	aspData, found := s.authServiceMap[aspId]
	if found {
		return aspData
	}

	return nil
}

func (s *UdpServer) processACOperation(knkMsg *common.AgentKnockMsg, conn *ACConn, srcAddr *common.NetAddress, dstAddrs []*common.NetAddress, openTime uint32) (artMsg *common.ACOpsResultMsg, err error) {
	artMsg = &common.ACOpsResultMsg{}
	if knkMsg == nil || conn == nil || srcAddr == nil || len(dstAddrs) == 0 {
		log.Error("[processACOperation] no address specified")
		err = common.ErrACEmptyPassAddress
		artMsg.ErrCode = common.ErrACEmptyPassAddress.ErrorCode()
		artMsg.ErrMsg = err.Error()
		return
	}

	srcAddrs := []*common.NetAddress{srcAddr}
	// check source ip associated address
	s.srcIpAssociatedAddrMapMutex.Lock()
	asscAddrs, found := s.srcIpAssociatedAddrMap[srcAddr.Ip]
	s.srcIpAssociatedAddrMapMutex.Unlock()
	if found {
		srcAddrs = append(srcAddrs, asscAddrs...)
	}

	acAddrStr := conn.ACPeer.RecvAddr().String()
	if openTime == 0 {
		openTime = DefaultIpOpenTime
	}

	aopMsg := &common.ServerACOpsMsg{
		UserId:           knkMsg.UserId,
		DeviceId:         knkMsg.DeviceId,
		OrganizationId:   knkMsg.OrganizationId,
		AuthServiceId:    knkMsg.AuthServiceId,
		ResourceId:       knkMsg.ResourceId,
		SourceAddrs:      srcAddrs,
		DestinationAddrs: dstAddrs,
		OpenTime:         openTime + ACOpenCompensationTime, // compensate ac open time
	}
	aopBytes, _ := json.Marshal(aopMsg)

	aopMd := &nhp.MsgData{
		ConnData:      conn.ConnData,
		HeaderType:    nhp.NHP_AOP,
		TransactionId: s.device.NextCounterIndex(),
		Compress:      true,
		PeerPk:        conn.ACPeer.PublicKey(),
		Message:       aopBytes,
		ResponseMsgCh: make(chan *nhp.PacketParserData),
	}

	if !s.IsRunning() {
		log.Error("server-agent(%s@%s)-ac(%s#%d@%s)[processACOperation] MsgData channel closed or being closed, skip sending", knkMsg.UserId, srcAddr.String(), conn.ACId, aopMd.TransactionId, acAddrStr)
		err = common.ErrPacketToMessageRoutineStopped
		artMsg.ErrCode = common.ErrPacketToMessageRoutineStopped.ErrorCode()
		artMsg.ErrMsg = err.Error()
		return
	}

	s.sendMsgCh <- aopMd

	// wait for ac sending back operation result
	// block until transaction completes
	acPpd := <-aopMd.ResponseMsgCh
	close(aopMd.ResponseMsgCh)

	if acPpd.Error != nil {
		log.Error("server-agent(%s@%s)-ac(%s#%d@%s)[processACOperation] failed to receive response from ac: %v", knkMsg.UserId, srcAddr.String(), conn.ACId, aopMd.TransactionId, acAddrStr, acPpd.Error)
		err = acPpd.Error
		artMsg.ErrCode = common.ErrServerACOpsFailed.ErrorCode()
		artMsg.ErrMsg = err.Error()
		return
	}

	if acPpd.HeaderType != nhp.NHP_ART {
		log.Error("server-agent(%s@%s)-ac(%s#%d@%s)[processACOperation] response has wrong type: %s", knkMsg.UserId, srcAddr.String(), conn.ACId, aopMd.TransactionId, acAddrStr, nhp.HeaderTypeToString(acPpd.HeaderType))
		err = common.ErrTransactionRepliedWithWrongType
		artMsg.ErrCode = common.ErrTransactionRepliedWithWrongType.ErrorCode()
		artMsg.ErrMsg = err.Error()
		return
	}

	err = json.Unmarshal(acPpd.BodyMessage, artMsg)
	if err != nil {
		log.Error("server-agent(%s@%s)-ac(%s#%d@%s)[processACOperation] failed to parse %s message: %v", knkMsg.UserId, srcAddr.String(), conn.ACId, aopMd.TransactionId, acAddrStr, nhp.HeaderTypeToString(acPpd.HeaderType), err)
		artMsg.ErrCode = common.ErrJsonParseFailed.ErrorCode()
		artMsg.ErrMsg = err.Error()
		return
	}

	if artMsg.ErrCode != common.ErrSuccess.ErrorCode() {
		log.Error("server-agent(%s@%s)-ac(%s#%d@%s)[processACOperation] response error: %+v", knkMsg.UserId, srcAddr.String(), conn.ACId, aopMd.TransactionId, acAddrStr, artMsg)
		err = common.ErrACOperationFailed
		return
	}

	return artMsg, nil
}

func (s *UdpServer) handleNhpOpenResource(req *common.NhpAuthRequest, res *common.ResourceData) (ackMsg *common.ServerKnockAckMsg, err error) {
	knkMsg := req.Msg
	srcAddr := req.SrcAddr
	addrStr := srcAddr.String()
	ackMsg = req.Ack

	acDstIpMap := make(map[string][]*common.NetAddress)
	for _, info := range res.Resources {
		addrs, exist := acDstIpMap[info.ACId]
		if exist {
			addrs = append(addrs, info.Addr)
			acDstIpMap[info.ACId] = addrs
		} else {
			acDstIpMap[info.ACId] = []*common.NetAddress{info.Addr}
		}
	}

	// PART III: request ac operation for each resource and block for response
	var acWg sync.WaitGroup
	var artMsgsMutex sync.Mutex
	artMsgs := make(map[string]*common.ACOpsResultMsg)

	for acId, dstAddrs := range acDstIpMap {
		s.acConnectionMapMutex.Lock()
		acConn, found := s.acConnectionMap[acId]
		s.acConnectionMapMutex.Unlock()
		if !found {
			log.Error("server-agent(%s@%s)-ac(@%s)[handleNhpOpenResource] no ac connection is available", knkMsg.UserId, addrStr, acId)
			err = common.ErrACConnectionNotFound
			ackMsg.ErrCode = common.ErrACConnectionNotFound.ErrorCode()
			ackMsg.ErrMsg = err.Error()
			return
		}

		acWg.Add(1)
		go func(id string, addrs []*common.NetAddress) {
			defer acWg.Done()

			openTime := res.OpenTime
			if knkMsg.HeaderType == nhp.NHP_EXT {
				openTime = 1 // timeout in 1 second
			}
			artMsg, _ := s.processACOperation(knkMsg, acConn, srcAddr, addrs, openTime)
			artMsgsMutex.Lock()
			artMsgs[id] = artMsg
			artMsgsMutex.Unlock()
		}(acId, dstAddrs)
	}
	acWg.Wait()

	var errCount int
	for _, artMsg := range artMsgs {
		if artMsg.ErrCode != common.ErrSuccess.ErrorCode() {
			errCount++
			break
		}
	}

	if errCount > 0 {
		log.Info("server-agent(%s@%s)[handleNhpOpenResource] failed: %+v", knkMsg.UserId, addrStr, artMsgs)
		err = common.ErrServerACOpsFailed
		ackMsg.ErrCode = common.ErrServerACOpsFailed.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return
	}

	ackMsg.PreAccessActions = make([]*common.PreAccessInfo, 0, len(artMsgs))
	for _, artMsg := range artMsgs {
		if artMsg.PreAccessAction != nil {
			ackMsg.PreAccessActions = append(ackMsg.PreAccessActions, artMsg.PreAccessAction)
		}
	}

	ackMsg.ErrCode = common.ErrSuccess.ErrorCode()
	ackMsg.ErrMsg = common.ErrSuccess.Error()
	return ackMsg, nil
}

func (us *UdpServer) NewNhpServerHelper(ppd *nhp.PacketParserData) *plugins.NhpServerPluginHelper {
	h := &plugins.NhpServerPluginHelper{}
	h.StopSignal = ppd.ConnData.StopSignal

	h.AuthWithNhpCallbackFunc = func(req *common.NhpAuthRequest, res *common.ResourceData) (*common.ServerKnockAckMsg, error) {
		return us.handleNhpOpenResource(req, res)
	}

	return h
}

func (us *UdpServer) FindPluginHandler(aspId string) plugins.PluginHandler {
	us.pluginHandlerMapMutex.Lock()
	defer us.pluginHandlerMapMutex.Unlock()

	handler, found := us.pluginHandlerMap[aspId]
	if !found {
		return nil
	}
	return handler
}
