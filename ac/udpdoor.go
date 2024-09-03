package ac

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"net"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/log"
	"github.com/OpenNHP/opennhp/nhp"
	"github.com/OpenNHP/opennhp/utils"
	"github.com/OpenNHP/opennhp/version"
)

var (
	ExeDirPath string
)

type AgentUser struct {
	UserId         string
	DeviceId       string
	OrganizationId string
	hash           hash.Hash
}

func (au *AgentUser) Hash() string {
	au.hash = nhp.NewHash(nhp.HASH_SM3)
	au.hash.Write([]byte(au.UserId))
	au.hash.Write([]byte(au.DeviceId))
	au.hash.Write([]byte(au.OrganizationId))
	// do not include Agent's PublicKey in calculating hash, because it may vary between Curve25519 and SM2
	sum := au.hash.Sum(nil)
	return string(sum)
}

type AgentUserCodeMap = map[string]*map[string]string // agent hash string first letter > agent hash string > token

type UdpDoor struct {
	config   *Config
	iptables *utils.IPTables
	ipset    *utils.IPSet

	stats struct {
		totalRecvBytes uint64
		totalSendBytes uint64
	}

	log *log.Logger

	remoteConnectionMutex sync.Mutex
	remoteConnectionMap   map[string]*UdpConn // indexed by remote UDP address

	serverPeerMutex sync.Mutex
	serverPeerMap   map[string]*nhp.UdpPeer // indexed by server's public key

	AgentUserTokenMutex sync.Mutex
	agentUserCodeMap    AgentUserCodeMap

	device  *nhp.Device
	wg      sync.WaitGroup
	running atomic.Bool

	signals struct {
		stop             chan struct{}
		serverMapUpdated chan struct{}
	}

	recvMsgCh <-chan *nhp.PacketParserData
	sendMsgCh chan *nhp.MsgData
}

type UdpConn struct {
	ConnData     *nhp.ConnectionData
	netConn      *net.UDPConn
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
func (d *UdpDoor) Start(dirPath string, logLevel int) (err error) {
	common.ExeDirPath = dirPath
	ExeDirPath = dirPath
	// init logger
	d.log = log.NewLogger("NHP-AC", logLevel, filepath.Join(ExeDirPath, "logs"), "ac")
	log.SetGlobalLogger(d.log)

	log.Info("=========================================================")
	log.Info("=== NHP-AC %s started                              ===", version.Version)
	log.Info("=== REVISION %s ===", version.CommitId)
	log.Info("=== RELEASE %s                       ===", version.BuildTime)
	log.Info("=========================================================")

	// init config
	err = d.loadBaseConfig()
	if err != nil {
		return err
	}

	d.iptables, err = utils.NewIPTables()
	if err != nil {
		log.Error("iptables command not found")
		return
	}

	d.ipset, err = utils.NewIPSet(false)
	if err != nil {
		log.Error("ipset command not found")
		return
	}

	prk, err := base64.StdEncoding.DecodeString(d.config.PrivateKeyBase64)
	if err != nil {
		log.Error("private key parse error %v\n", err)
		return fmt.Errorf("private key parse error %v", err)
	}

	d.device = nhp.NewDevice(nhp.NHP_AC, prk, nil)
	if d.device == nil {
		log.Critical("failed to create device %v\n", err)
		return fmt.Errorf("failed to create device %v", err)
	}

	d.remoteConnectionMap = make(map[string]*UdpConn)
	d.serverPeerMap = make(map[string]*nhp.UdpPeer)
	d.agentUserCodeMap = make(AgentUserCodeMap)

	// load peers
	d.loadPeers()

	d.signals.stop = make(chan struct{})
	d.signals.serverMapUpdated = make(chan struct{}, 1)

	d.recvMsgCh = d.device.DecryptedMsgQueue
	d.sendMsgCh = make(chan *nhp.MsgData, nhp.SendQueueSize)

	// start device routines
	d.device.Start()

	// start door routines
	d.wg.Add(3)
	go d.sendMessageRoutine()
	go d.recvMessageRoutine()
	go d.maintainServerConnectionRoutine()

	d.running.Store(true)
	return nil
}

func (d *UdpDoor) Stop() {
	d.running.Store(false)
	close(d.signals.stop)

	d.device.Stop()
	d.StopConfigWatch()
	d.wg.Wait()
	close(d.sendMsgCh)
	close(d.signals.serverMapUpdated)

	log.Info("==========================")
	log.Info("=== NHP-AC stopped ===")
	log.Info("==========================")
	d.log.Close()
}

func (d *UdpDoor) IsRunning() bool {
	return d.running.Load()
}

func (d *UdpDoor) newConnection(addr *net.UDPAddr) (conn *UdpConn) {
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

	conn.ConnData = &nhp.ConnectionData{
		Device:               d.device,
		CookieStore:          &nhp.CookieStore{},
		RemoteTransactionMap: make(map[uint64]*nhp.RemoteTransaction),
		LocalAddr:            localAddr,
		RemoteAddr:           addr,
		TimeoutMs:            DefaultConnectionTimeoutMs,
		SendQueue:            make(chan *nhp.UdpPacket, PacketQueueSizePerConnection),
		RecvQueue:            make(chan *nhp.UdpPacket, PacketQueueSizePerConnection),
		BlockSignal:          make(chan struct{}),
		SetTimeoutSignal:     make(chan struct{}),
		StopSignal:           make(chan struct{}),
	}

	// start connection receive routine
	conn.ConnData.Add(1)
	go d.recvPacketRoutine(conn)

	return conn
}

func (d *UdpDoor) sendMessageRoutine() {
	defer d.wg.Done()
	defer log.Info("sendMessageRoutine stopped")

	log.Info("sendMessageRoutine started")

	for {
		select {
		case <-d.signals.stop:
			return

		case md, ok := <-d.sendMsgCh:
			if !ok {
				return
			}
			if md == nil || md.RemoteAddr == nil {
				log.Warning("Invalid initiator session starter")
				continue
			}

			addrStr := md.RemoteAddr.String()

			d.remoteConnectionMutex.Lock()
			conn, found := d.remoteConnectionMap[addrStr]
			d.remoteConnectionMutex.Unlock()

			if found {
				md.ConnData = conn.ConnData
			} else {
				conn = d.newConnection(md.RemoteAddr)
				if conn == nil {
					log.Error("Failed to dial to remote address: %s", addrStr)
					continue
				}

				d.remoteConnectionMutex.Lock()
				d.remoteConnectionMap[addrStr] = conn
				d.remoteConnectionMutex.Unlock()

				md.ConnData = conn.ConnData

				// launch connection routine
				d.wg.Add(1)
				go d.connectionRoutine(conn)
			}

			d.device.SendMsgToPacket(md)
		}
	}
}

func (d *UdpDoor) SendPacket(pkt *nhp.UdpPacket, conn *UdpConn) (n int, err error) {
	defer func() {
		atomic.AddUint64(&d.stats.totalSendBytes, uint64(n))
		atomic.StoreInt64(&conn.ConnData.LastLocalSendTime, time.Now().UnixNano())

		if !pkt.KeepAfterSend {
			d.device.ReleaseUdpPacket(pkt)
		}
	}()

	pktType := nhp.HeaderTypeToString(pkt.HeaderType)
	//log.Debug("Send [%s] packet (%s -> %s): %+v", pktType, conn.ConnData.LocalAddr.String(), conn.ConnData.RemoteAddr.String(), pkt.Packet)
	log.Info("Send [%s] packet (%s -> %s), %d bytes", pktType, conn.ConnData.LocalAddr.String(), conn.ConnData.RemoteAddr.String(), len(pkt.Packet))
	log.Evaluate("Send [%s] packet (%s -> %s, %d bytes)", pktType, conn.ConnData.LocalAddr.String(), conn.ConnData.RemoteAddr.String(), len(pkt.Packet))
	return conn.netConn.Write(pkt.Packet)
}

func (d *UdpDoor) recvPacketRoutine(conn *UdpConn) {
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
		pkt := d.device.AllocateUdpPacket()
		n, err := conn.netConn.Read(pkt.Buf[:])
		if err != nil {
			d.device.ReleaseUdpPacket(pkt)
			if n == 0 {
				// udp connection closed, it is not an error
				return
			}
			log.Error("Failed to receive from remote address %s (%v)", addrStr, err)
			continue
		}

		// add total recv bytes
		atomic.AddUint64(&d.stats.totalRecvBytes, uint64(n))

		// check minimal length
		if n < nhp.HeaderSize {
			d.device.ReleaseUdpPacket(pkt)
			log.Error("Received UDP packet from %s is too short, discard", addrStr)
			continue
		}

		pkt.Packet = pkt.Buf[:n]
		//log.Trace("receive udp packet (%s -> %s): %+v", conn.ConnData.RemoteAddr.String(), conn.ConnData.LocalAddr.String(), pkt.Packet)

		typ, _, err := d.device.RecvPrecheck(pkt)
		msgType := nhp.HeaderTypeToString(typ)
		log.Info("Receive [%s] packet (%s -> %s), %d bytes", msgType, addrStr, conn.ConnData.LocalAddr.String(), n)
		log.Evaluate("Receive [%s] packet (%s -> %s), %d bytes", msgType, addrStr, conn.ConnData.LocalAddr.String(), n)
		if err != nil {
			d.device.ReleaseUdpPacket(pkt)
			log.Warning("Receive [%s] packet (%s -> %s), precheck error: %v", msgType, addrStr, conn.ConnData.LocalAddr.String(), err)
			log.Evaluate("Receive [%s] packet (%s -> %s) precheck error: %v", msgType, addrStr, conn.ConnData.LocalAddr.String(), err)
			continue
		}

		atomic.StoreInt64(&conn.ConnData.LastLocalRecvTime, time.Now().UnixNano())

		conn.ConnData.ForwardInboundPacket(pkt)
	}
}

func (d *UdpDoor) connectionRoutine(conn *UdpConn) {
	addrStr := conn.ConnData.RemoteAddr.String()

	defer d.wg.Done()
	defer log.Debug("Connection routine: %s stopped", addrStr)

	log.Debug("Connection routine: %s started", addrStr)

	// stop receiving packets and clean up
	defer func() {
		d.remoteConnectionMutex.Lock()
		delete(d.remoteConnectionMap, addrStr)
		d.remoteConnectionMutex.Unlock()

		conn.Close()
	}()

	for {
		select {
		case <-d.signals.stop:
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
			d.SendPacket(pkt, conn)

		case pkt, ok := <-conn.ConnData.RecvQueue:
			if !ok {
				return
			}
			if pkt == nil {
				continue
			}
			log.Debug("Received udp packet len [%d] from addr: %s\n", len(pkt.Packet), addrStr)

			if pkt.HeaderType == nhp.NHP_KPL {
				d.device.ReleaseUdpPacket(pkt)
				log.Info("Receive [NHP_KPL] message (%s -> %s)", addrStr, conn.ConnData.LocalAddr.String())
				continue
			}

			if d.device.IsTransactionResponse(pkt.HeaderType) {
				// forward to a specific transaction
				transactionId := pkt.Counter()
				transaction := d.device.FindLocalTransaction(transactionId)
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
			d.device.RecvPacketToMsg(pd)

		case <-conn.ConnData.BlockSignal:
			log.Critical("blocking address %s", addrStr)
			return
		}
	}
}

func (d *UdpDoor) recvMessageRoutine() {
	defer d.wg.Done()
	defer log.Info("recvMessageRoutine stopped")

	log.Info("recvMessageRoutine started")

	for {
		select {
		case <-d.signals.stop:
			return

		case ppd, ok := <-d.recvMsgCh:
			if !ok {
				return
			}
			if ppd == nil {
				continue
			}

			switch ppd.HeaderType {
			case nhp.NHP_AOP:
				// deal with NHP_AOP message
				go d.HandleACOperations(ppd)
			}
		}
	}
}

// keep interaction between ac and server in certain time interval to keep outwards ip path active
func (d *UdpDoor) maintainServerConnectionRoutine() {
	defer d.wg.Done()
	defer log.Info("maintainServerConnectionRoutine stopped")

	log.Info("maintainServerConnectionRoutine started")

	// reset iptables before exiting
	defer d.iptables.ResetAllInput()

	var discoveryRoutineWg sync.WaitGroup
	defer discoveryRoutineWg.Wait()

	for {
		// make a local copy of servers then iterate because next operations are time consuming (too long to use locked iteration)
		d.serverPeerMutex.Lock()
		var serverCount int32 = int32(len(d.serverPeerMap))
		discoveryQuitArr := make([]chan struct{}, 0, serverCount)
		discoveryFailStatusArr := make([]*int32, 0, serverCount)

		for _, server := range d.serverPeerMap {
			// launch discovery routine for each server
			fail := new(int32)
			discoveryFailStatusArr = append(discoveryFailStatusArr, fail)
			quit := make(chan struct{})
			discoveryQuitArr = append(discoveryQuitArr, quit)

			discoveryRoutineWg.Add(1)
			go d.serverDiscovery(server, &discoveryRoutineWg, fail, quit)
		}
		d.serverPeerMutex.Unlock()

		// check whether all server discovery failed.
		// If so, open all blocked input
		quitCheck := make(chan struct{})
		discoveryQuitArr = append(discoveryQuitArr, quitCheck)
		discoveryRoutineWg.Add(1)
		go func() {
			defer discoveryRoutineWg.Done()

			for {
				select {
				case <-d.signals.stop:
					return
				case <-quitCheck:
					return
				case <-time.After(MinialServerDiscoveryInterval * time.Second):
					var totalFail int32
					for _, status := range discoveryFailStatusArr {
						totalFail += atomic.LoadInt32(status)
					}

					if totalFail < int32(len(discoveryFailStatusArr)) {
						d.iptables.ResetAllInput()
					} else {
						d.iptables.AcceptAllInput()
					}
				}
			}
		}()

		select {
		case <-d.signals.stop:
			return
		case _, ok := <-d.signals.serverMapUpdated:
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

func (d *UdpDoor) serverDiscovery(server *nhp.UdpPeer, discoveryRoutineWg *sync.WaitGroup, serverFailCount *int32, quit <-chan struct{}) {
	defer discoveryRoutineWg.Done()

	acId := d.config.ACId
	serverAddr := server.HostOrAddr()
	server, sendAddr := d.ResolvePeer(server)
	if sendAddr == nil {
		log.Error("Cannot connect to nil server address")
		return
	}

	addrStr := sendAddr.String()

	defer log.Info("server discovery sub-routine at %s stopped", serverAddr)
	log.Info("server discovery sub-routine at %s started", serverAddr)

	var failCount int

	for {
		var lastSendTime int64
		var lastRecvTime int64
		var connected bool

		// find whether connection is already connected
		d.remoteConnectionMutex.Lock()
		conn, found := d.remoteConnectionMap[addrStr]
		d.remoteConnectionMutex.Unlock()

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
				AuthServiceId: d.config.AuthServiceId,
				ResourceIds:   d.config.ResourceIds,
			}
			aolBytes, _ := json.Marshal(aolMsg)

			aolMd := &nhp.MsgData{
				RemoteAddr:    sendAddr.(*net.UDPAddr),
				HeaderType:    nhp.NHP_AOL,
				TransactionId: d.device.NextCounterIndex(),
				Compress:      true,
				PeerPk:        peerPbk,
				Message:       aolBytes,
				ResponseMsgCh: make(chan *nhp.PacketParserData),
			}

			if !d.IsRunning() {
				log.Error("ac(%s#%d)[ACOnline] MsgData channel closed or being closed, skip sending", acId, aolMd.TransactionId)
				return
			}

			d.sendMsgCh <- aolMd // create new connection
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
							d.remoteConnectionMutex.Lock()
							conn = d.remoteConnectionMap[addrStr]
							if conn != nil {
								log.Info("server discovery failed, close local connection: %s", conn.ConnData.LocalAddr.String())
								delete(d.remoteConnectionMap, addrStr)
							}
							d.remoteConnectionMutex.Unlock()
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

				if ppd.HeaderType != nhp.NHP_AAK {
					log.Error("ac(%s#%d)[ACOnline] response from server %s has wrong type: %s", acId, aolMd.TransactionId, addrStr, nhp.HeaderTypeToString(ppd.HeaderType))
					err = common.ErrTransactionRepliedWithWrongType
					return
				}

				aakMsg := &common.ServerACAckMsg{}
				err = json.Unmarshal(ppd.BodyMessage, aakMsg)
				if err != nil {
					log.Error("ac(%s#%d)[HandleACAck] failed to parse %s message: %v", acId, ppd.SenderId, nhp.HeaderTypeToString(ppd.HeaderType), err)
					return
				}

				// server discovery succeeded
				failCount = 0
				atomic.StoreInt32(serverFailCount, 0)
				d.remoteConnectionMutex.Lock()
				conn = d.remoteConnectionMap[addrStr] // conn must be available at this point
				conn.connected.Store(true)
				conn.externalAddr = aakMsg.ACAddr
				d.remoteConnectionMutex.Unlock()
				log.Info("ac(%s#%d)[ACOnline] succeed. ac external address is %s, replied by server %s", acId, aolMd.TransactionId, aakMsg.ACAddr, addrStr)
			}()

		} else if connected {
			if (currTime - lastSendTime) > int64(ServerKeepaliveInterval*time.Second) {
				// send NHP_KPL to server if no send happens within ServerKeepaliveInterval
				md := &nhp.MsgData{
					RemoteAddr: sendAddr.(*net.UDPAddr),
					HeaderType: nhp.NHP_KPL,
					//PeerPk:        peerPbk, // pubkey not needed
					TransactionId: d.device.NextCounterIndex(),
				}

				d.sendMsgCh <- md // send NHP_KPL to server via existing connection
				server.UpdateSend(currTime)
			}
		}

		select {
		case <-d.signals.stop:
			return
		case <-quit:
			return
		case <-time.After(MinialServerDiscoveryInterval * time.Second):
			// wait for ServerConnectionDiscoveryInterval
		}
	}
}

func (d *UdpDoor) AddServerPeer(server *nhp.UdpPeer) {
	if server.DeviceType() == nhp.NHP_SERVER {
		d.device.AddPeer(server)

		d.serverPeerMutex.Lock()
		d.serverPeerMap[server.PublicKeyBase64()] = server
		d.serverPeerMutex.Unlock()

		// renew server connection cycle
		if len(d.signals.serverMapUpdated) == 0 {
			d.signals.serverMapUpdated <- struct{}{}
		}
	}
}

func (d *UdpDoor) RemoveServerPeer(serverKey string) {
	d.serverPeerMutex.Lock()
	beforeSize := len(d.serverPeerMap)
	delete(d.serverPeerMap, serverKey)
	afterSize := len(d.serverPeerMap)
	d.serverPeerMutex.Unlock()

	if beforeSize != afterSize {
		// renew server connection cycle
		if len(d.signals.serverMapUpdated) == 0 {
			d.signals.serverMapUpdated <- struct{}{}
		}
	}
}

func (d *UdpDoor) GenerateAccessToken(au *AgentUser) string {
	hashStr := au.Hash()
	timeStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	au.hash.Write([]byte(timeStr))
	token := base64.StdEncoding.EncodeToString(au.hash.Sum(nil))

	d.AgentUserTokenMutex.Lock()
	defer d.AgentUserTokenMutex.Unlock()

	tokenMap, found := d.agentUserCodeMap[hashStr[0:1]]
	if found {
		(*tokenMap)[hashStr] = token
	} else {
		tokenMap = &map[string]string{hashStr: token}
		d.agentUserCodeMap[hashStr[0:1]] = tokenMap
	}

	// log.Debug("user %+v, hash: %s", au, hashStr)
	// log.Debug("agentUserCodeMap: %+v", d.agentUserCodeMap)
	// log.Debug("tokenMap: %+v", d.agentUserCodeMap[hashStr[0:1]])
	return token
}

func (d *UdpDoor) VerifyAccessToken(au *AgentUser, token string) bool {
	hashStr := au.Hash()

	d.AgentUserTokenMutex.Lock()
	defer d.AgentUserTokenMutex.Unlock()

	// log.Debug("verify access token: %s", token)
	// log.Debug("user %+v, hash: %s", au, hashStr)
	// log.Debug("agentUserCodeMap: %+v", d.agentUserCodeMap)
	// log.Debug("tokenMap: %+v", d.agentUserCodeMap[hashStr[0:1]])

	tokenMap, found := d.agentUserCodeMap[hashStr[0:1]]
	if found {
		foundToken, found := (*tokenMap)[hashStr]
		if found {
			return token == foundToken
		}
	}

	return false
}

func (d *UdpDoor) DeleteAccessToken(au *AgentUser) {
	hashStr := au.Hash()

	d.AgentUserTokenMutex.Lock()
	defer d.AgentUserTokenMutex.Unlock()

	tokenMap, found := d.agentUserCodeMap[hashStr[0:1]]
	if found {
		delete(*tokenMap, hashStr)
		if len(*tokenMap) == 0 {
			delete(d.agentUserCodeMap, hashStr[0:1])
		}
	}
}

// if the server uses hostname as destination, find the correct peer with the actual IP address
func (d *UdpDoor) ResolvePeer(peer *nhp.UdpPeer) (*nhp.UdpPeer, net.Addr) {
	addr := peer.SendAddr()
	if addr == nil {
		return peer, nil
	}

	if len(peer.Hostname) == 0 {
		// peer with fixed ip, no change
		return peer, addr
	}

	actualIp := peer.ResolvedIp()
	if peer.Ip == actualIp {
		// peer with the correct resolved address, no change
		return peer, addr
	}

	d.serverPeerMutex.Lock()
	defer d.serverPeerMutex.Unlock()
	for _, p := range d.serverPeerMap {
		if p.Ip == actualIp {
			p.CopyResolveStatus(peer)
			return p, addr
		}
	}

	return peer, addr
}
