package agent

import (
	"encoding/base64"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/core"
	"github.com/OpenNHP/opennhp/log"
	"github.com/OpenNHP/opennhp/version"
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

type UdpAgent struct {
	stats struct {
		totalRecvBytes uint64
		totalSendBytes uint64
	}

	config *Config
	log    *log.Logger

	remoteConnectionMutex sync.Mutex
	remoteConnectionMap   map[string]*UdpConn // indexed by remote UDP address

	knockTargetMapMutex sync.Mutex
	knockTargetMap      map[string]*KnockTarget // indexed by aspId + resId

	serverPeerMutex sync.Mutex
	serverPeerMap   map[string]*core.UdpPeer // indexed by server's public key

	device  *core.Device
	wg      sync.WaitGroup
	running atomic.Bool

	signals struct {
		stop                  chan struct{}
		knockTargetStop       chan struct{}
		knockTargetMapUpdated chan struct{}
	}

	recvMsgCh <-chan *core.PacketParserData
	sendMsgCh chan *core.MsgData

	// one agent should serve only one specific user at a time
	knockUserMutex sync.RWMutex
	knockUser      *KnockUser
	deviceId       string
	checkResults   map[string]any
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
func (a *UdpAgent) Start(dirPath string, logLevel int) (err error) {
	common.ExeDirPath = dirPath
	ExeDirPath = dirPath
	// init logger
	a.log = log.NewLogger("NHP-Agent", logLevel, filepath.Join(ExeDirPath, "logs"), "agent")
	log.SetGlobalLogger(a.log)

	log.Info("=========================================================")
	log.Info("=== NHP-Agent %s started                           ===", version.Version)
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

	a.device = core.NewDevice(core.NHP_AGENT, prk, nil)
	if a.device == nil {
		log.Critical("failed to create device %v\n", err)
		return fmt.Errorf("failed to create device %v", err)
	}

	// start device routines
	a.device.Start()

	// load peers
	a.loadPeers()

	a.remoteConnectionMap = make(map[string]*UdpConn)

	a.signals.stop = make(chan struct{})
	a.signals.knockTargetStop = make(chan struct{})
	a.signals.knockTargetMapUpdated = make(chan struct{}, 1)

	// load knock resources
	a.loadResources()

	a.recvMsgCh = a.device.DecryptedMsgQueue
	a.sendMsgCh = make(chan *core.MsgData, core.SendQueueSize)

	// start agent routines
	a.wg.Add(2)
	go a.sendMessageRoutine()
	go a.recvMessageRoutine()

	a.running.Store(true)

	time.Sleep(1000 * time.Millisecond)

	return nil
}

func (a *UdpAgent) StartKnockLoop() int {
	a.knockTargetMapMutex.Lock()
	size := len(a.knockTargetMap)
	a.knockTargetMapMutex.Unlock()
	// start knock preset resources
	a.wg.Add(1)
	go a.knockResourceRoutine()

	return size
}

func (a *UdpAgent) StopKnockLoop() {
	close(a.signals.knockTargetStop)
}

func (a *UdpAgent) SetKnockUser(usrId string, orgId string, userData map[string]any) {
	a.knockUserMutex.Lock()
	a.knockUser.UserId = usrId
	a.knockUser.OrganizationId = orgId
	a.knockUser.UserData = userData
	a.knockUserMutex.Unlock()
}

func (a *UdpAgent) SetDeviceId(devId string) {
	a.deviceId = devId
}

func (a *UdpAgent) SetCheckResults(results map[string]any) {
	a.checkResults = results
}

// export Stop
func (a *UdpAgent) Stop() {
	a.running.Store(false)
	close(a.signals.knockTargetStop)
	close(a.signals.stop)
	a.device.Stop()
	a.StopConfigWatch()
	a.wg.Wait()
	close(a.sendMsgCh)
	close(a.signals.knockTargetMapUpdated)

	log.Info("=========================")
	log.Info("=== NHP-Agent stopped ===")
	log.Info("=========================")
	a.log.Close()
}

func (a *UdpAgent) IsRunning() bool {
	return a.running.Load()
}

func (a *UdpAgent) newConnection(addr *net.UDPAddr) (conn *UdpConn) {
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

func (a *UdpAgent) sendMessageRoutine() {
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

func (a *UdpAgent) SendPacket(pkt *core.Packet, conn *UdpConn) (n int, err error) {
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

func (a *UdpAgent) recvPacketRoutine(conn *UdpConn) {
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

func (a *UdpAgent) connectionRoutine(conn *UdpConn) {
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

func (a *UdpAgent) recvMessageRoutine() {
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
			case core.NHP_COK:
				// synchronously block and deal with cookie message to ensure future messages will be correctly processed. note cookie is not handled as a transaction, so it arrives in here
				a.HandleCookieMessage(ppd)

			}
		}
	}
}

func (a *UdpAgent) knockResourceRoutine() {
	defer a.wg.Done()
	defer log.Info("knockResourceRoutine stopped")

	log.Info("knockResourceRoutine started")

	var knockRoutineWg sync.WaitGroup
	defer knockRoutineWg.Wait()

	for {
		a.knockTargetMapMutex.Lock()
		targetSize := len(a.knockTargetMap)
		targetQuitArr := make([]chan struct{}, 0, targetSize)

		for k, r := range a.knockTargetMap {
			// launch knock routine for each knock target
			q := make(chan struct{})
			targetQuitArr = append(targetQuitArr, q)

			knockRoutineWg.Add(1)
			go func(knockStr string, res *KnockTarget, quit <-chan struct{}) {
				defer knockRoutineWg.Done()
				defer log.Info("knock %s sub-routine stopped", knockStr)
				defer func() {
					a.ExitKnockRequest(res)
				}()

				log.Info("knock %s sub-routine started", knockStr)

				for {
					select {
					case <-a.signals.knockTargetStop:
						return
					case <-quit:
						return
					default:
					}

					ackMsg, err := a.Knock(res) // timeout in AgentLocalTransactionTimeoutMs
					if err != nil {
						// if error happens wait some time (total AgentLocalTransactionResponseTimeoutMs) to retry
						log.Error("failed to knock %s, error: %v", knockStr, err)
						continue // retry knock
					}

					log.Info("knock %s succeeded, next knock in %d seconds", knockStr, ackMsg.OpenTime)
					select {
					case <-a.signals.knockTargetStop:
						return
					case <-quit:
						return
					case <-time.After(time.Second * time.Duration(ackMsg.OpenTime)):
						// continue knock
					}
				}
			}(k, r, q)
		}
		a.knockTargetMapMutex.Unlock()

		// block until knockTargetMap is updated
		select {
		case <-a.signals.knockTargetStop:
			return
		case <-a.signals.knockTargetMapUpdated:
			// stop all current knock routines
			for _, q := range targetQuitArr {
				close(q)
			}
			log.Info("restart knock cycle with updated targets")
			// continue and restart with new knock targets
		}
	}
}

func (a *UdpAgent) AddServer(server *core.UdpPeer) {
	if server.DeviceType() == core.NHP_SERVER {
		a.device.AddPeer(server)
		a.serverPeerMutex.Lock()
		a.serverPeerMap[server.PublicKeyBase64()] = server
		a.serverPeerMutex.Unlock()
	}
}

func (a *UdpAgent) RemoveServer(serverKey string) {
	a.serverPeerMutex.Lock()
	delete(a.serverPeerMap, serverKey)
	a.serverPeerMutex.Unlock()
}

func (a *UdpAgent) AddResource(res *KnockResource) error {
	peer := a.FindServerPeerFromResource(res)
	if peer == nil {
		return common.ErrKnockServerNotFound
	}

	updated := false
	a.knockTargetMapMutex.Lock()
	target, found := a.knockTargetMap[res.Id()]
	if found {
		target.SetResource(res)
		target.SetServer(peer)
	} else {
		a.knockTargetMap[res.Id()] = &KnockTarget{
			KnockResource: *res,
			ServerPeer:    peer,
		}
		updated = true
	}
	a.knockTargetMapMutex.Unlock()

	if updated {
		// renew knock cycle
		if len(a.signals.knockTargetMapUpdated) == 0 {
			a.signals.knockTargetMapUpdated <- struct{}{}
		}
	}

	return nil
}

func (a *UdpAgent) RemoveResource(aspId string, resId string) {
	res := &KnockResource{
		AuthServiceId: aspId,
		ResourceId:    resId,
	}

	a.knockTargetMapMutex.Lock()
	beforeSize := len(a.knockTargetMap)
	delete(a.knockTargetMap, res.Id())
	afterSize := len(a.knockTargetMap)
	a.knockTargetMapMutex.Unlock()

	if beforeSize != afterSize {
		// renew knock cycle
		if len(a.signals.knockTargetMapUpdated) == 0 {
			a.signals.knockTargetMapUpdated <- struct{}{}
		}
	}
}

func (a *UdpAgent) FindServerPeerFromResource(res *KnockResource) *core.UdpPeer {
	a.serverPeerMutex.Lock()
	defer a.serverPeerMutex.Unlock()
	for _, peer := range a.serverPeerMap {
		if peer.HostOrAddr() == res.ServerAddr {
			return peer
		}
	}

	return nil
}

// if the server uses hostname as destination, find the correct peer with the actual IP address
func (a *UdpAgent) ResolvePeer(peer *core.UdpPeer) (*core.UdpPeer, net.Addr) {
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

	a.serverPeerMutex.Lock()
	defer a.serverPeerMutex.Unlock()
	for _, p := range a.serverPeerMap {
		if p.Ip == actualIp {
			p.CopyResolveStatus(peer)
			return p, addr
		}
	}

	return peer, addr
}
