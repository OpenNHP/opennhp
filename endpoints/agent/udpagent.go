package agent

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
	wasmEngine "github.com/OpenNHP/opennhp/nhp/core/wasm/engine"
	ztdolib "github.com/OpenNHP/opennhp/nhp/core/ztdo"
	"github.com/OpenNHP/opennhp/nhp/log"
	utils "github.com/OpenNHP/opennhp/nhp/utils"
	"github.com/OpenNHP/opennhp/nhp/version"
)

var (
	ExeDirPath                 string
	SmartDataPolicyRefreshTime = 15 * int64(time.Second)
)

type KnockUser struct {
	UserId         string
	OrganizationId string
	UserData       map[string]any
}

type KnockResource struct {
	AuthServiceId  string `json:"aspId"`
	ResourceId     string `json:"resId"`
	ServerHostname string `json:"serverHostname"`
	ServerIp       string `json:"serverIp"`
	ServerPort     int    `json:"serverPort"`
}

func (res *KnockResource) Id() string {
	return res.AuthServiceId + "/" + res.ResourceId
}

func (res *KnockResource) ServerHost() string {
	hostAddr := res.ServerIp
	if len(res.ServerHostname) > 0 {
		hostAddr = res.ServerHostname
	}
	if res.ServerPort == 0 {
		return hostAddr
	}
	return fmt.Sprintf("%s:%d", hostAddr, res.ServerPort)
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

func (kt *KnockTarget) SetServerPeer(peer *core.UdpPeer) {
	kt.Lock()
	defer kt.Unlock()

	kt.ServerPeer = peer
}

func (kt *KnockTarget) GetServerPeer() *core.UdpPeer {
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

	// dhp
	smartPolicyEngine          map[string]*wasmEngine.Engine // index by smart data policy identifier
	decryptedZtdoRecord        map[string]string             // index by data object id
	smartPolicyIdentifier      map[string]string             // index by data object id
	smartDataPolicyRefreshTime map[string]int64              // indexed by data object id, use to record the refresh time of smart data policy, the unit of time is UnixNano
	dataAccessRefreshMutex     sync.Mutex

	safeTee            atomic.Bool
	trustedByNHPServer atomic.Bool
	trustedByNHPDB     atomic.Bool
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
	err = a.loadDHPConfig()
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
	_ = a.loadPeers()

	a.remoteConnectionMap = make(map[string]*UdpConn)

	a.signals.stop = make(chan struct{})
	a.signals.knockTargetStop = make(chan struct{})
	a.signals.knockTargetMapUpdated = make(chan struct{}, 1)

	// load knock resources
	_ = a.loadResources()

	a.recvMsgCh = a.device.DecryptedMsgQueue
	a.sendMsgCh = make(chan *core.MsgData, core.SendQueueSize)

	// initialize dhp related stuff
	a.smartPolicyEngine = make(map[string]*wasmEngine.Engine)
	a.decryptedZtdoRecord = make(map[string]string)
	a.smartDataPolicyRefreshTime = make(map[string]int64)
	a.smartPolicyIdentifier = make(map[string]string)
	a.trustedByNHPServer.Store(false)
	a.trustedByNHPDB.Store(false)

	// start agent routines
	a.wg.Add(2)
	go a.sendMessageRoutine()
	go a.recvMessageRoutine()

	a.running.Store(true)
	a.safeTee.Store(false)

	time.Sleep(1000 * time.Millisecond)

	return nil
}

func (a *UdpAgent) RestartAgent() error {
	a.Stop()
	a.config = nil // re-load config
	err := a.Start(common.ExeDirPath, 4)
	if err != nil {
		return err
	}

	a.StartDHPKnockLoop()
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

func (a *UdpAgent) StartDHPKnockLoop() {
	a.wg.Add(1)
	go a.dhpKnockResourceRoutine()
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
			_, _ = a.SendPacket(pkt, conn)

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
					_, _ = a.ExitKnockRequest(res)
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

func (a *UdpAgent) dhpKnockResourceRoutine() {
	defer a.wg.Done()
	defer log.Info("dhpKnockResourceRoutine stopped")

	log.Info("dhpKnockResourceRoutine started")

	for {
		select {
		case <-a.signals.stop:
			return
		default: // don't block for knock failure
		}
		ackMsg, err := a.KnockDHP()

		if err != nil {
			a.safeTee.Store(false)

			// if error happens wait some time (total AgentLocalTransactionResponseTimeoutMs) to retry
			log.Error("failed to knock, error: %v", err)
			// avoid flood attack from server side
			time.Sleep(core.FailureRetryInterval * time.Second)
			continue // retry knock
		}

		log.Info("knock succeeded, next knock in %d seconds", ackMsg.OpenTime)
		a.safeTee.Store(true)

		select {
		case <-a.signals.stop:
			return
		case <-time.After(time.Second * time.Duration(ackMsg.OpenTime)):
			// continue knock
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
		log.Error("failed to find corresponding server peer for resource %s", res.Id())
		return common.ErrKnockServerNotFound
	}

	updated := false
	a.knockTargetMapMutex.Lock()
	target, found := a.knockTargetMap[res.Id()]
	if found {
		target.SetResource(res)
		target.SetServerPeer(peer)
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
		if peer.Host() == res.ServerHost() {
			return peer
		}
	}

	return nil
}

func (a *UdpAgent) StartConfidentialComputing(ztdoId string, taId string, function string, params map[string]any) (any, error) {
	var err error
	var policyId string

	output, refreshSdp, decrypted := a.PreCheckDataAccess(ztdoId)

	if refreshSdp {
		a.dataAccessRefreshMutex.Lock()
		defer a.dataAccessRefreshMutex.Unlock()

		// secondly check again
		output, refreshSdp, decrypted = a.PreCheckDataAccess(ztdoId)

		if refreshSdp {
			output, err = a.RefreshDataAccess(ztdoId, decrypted, output)
			if err != nil {
				return nil, fmt.Errorf("Failed to refresh SDP: %s", err.Error())
			}
		}
	}

	// inject data path to params
	params["path"] = output

	var exist bool
	if policyId, exist = a.smartPolicyIdentifier[ztdoId]; !exist {
		return nil, fmt.Errorf("Error: fail to find policyId for ztdoId %s.\n", ztdoId)
	}

	taRes, err := a.CallTrustedApplication(taId, function, params, policyId)
	if err != nil {
		return nil, fmt.Errorf("fail to call trusted application with error: %s\n", err.Error())
	} else {
		var structResult map[string]any

		err := json.Unmarshal([]byte(taRes), &structResult)
		if err != nil {
			return nil, fmt.Errorf("fail to unmarshal confidential computing result: %s\n", err.Error())
		}

		return structResult, nil
	}
}

func (a *UdpAgent) PreCheckDataAccess(ztdoId string) (output string, refreshSdp bool, decrypted bool) {
	output = ""

	// Check whether the smart data policy needs to be refreshed
	if sdpRefreshTime, exist := a.smartDataPolicyRefreshTime[ztdoId]; exist {
		if time.Now().UnixNano()-sdpRefreshTime > SmartDataPolicyRefreshTime {
			refreshSdp = true
		}
	} else {
		refreshSdp = true
	}

	// Check whether the ZTDO has been decrypted
	if plaintextPath, exist := a.decryptedZtdoRecord[ztdoId]; exist {
		output = plaintextPath
		decrypted = true
	} else {
		decrypted = false
		refreshSdp = true
	}

	return output, refreshSdp, decrypted
}

func (a *UdpAgent) RefreshDataAccess(ztdoId string, decrypted bool, decryptedOutput string) (output string, err error) {
	ztdo := ztdolib.NewZtdo()

	consumerEphemeralEcdh := core.NewECDH(a.config.GetEccType())
	teeEcdh := a.config.GetTeeEcdh()

	darMsg := common.DARMsg{
		DoId:                       ztdoId,
		UserId:                     a.config.UserId,
		TeePublicKey:               teeEcdh.PublicKeyBase64(),
		ConsumerEphemeralPublicKey: consumerEphemeralEcdh.PublicKeyBase64(),
	}
	serverPeer := a.GetFirstServerPeer()
	result, dagMsg := a.SendDARMsgToServer(serverPeer, darMsg)
	if result {
		a.trustedByNHPDB.Store(true) // agent has been trusted by NHP DB

		// update smart data policy refresh time
		a.smartDataPolicyRefreshTime[ztdoId] = time.Now().UnixNano()

		log.Info("[StartConfidentialComputing] Refresh smart data policy for data object which id is %s", ztdoId)

		if !decrypted {
			output, err = utils.GenerateTempFilePath("plaintext-*")
			if err != nil {
				return "", fmt.Errorf("Error: fail to generating temporary file path: %w", err)
			}

			dataPrkWrapping := ztdolib.DataPrivateKeyWrapping{}

			if err := json.Unmarshal([]byte(dagMsg.Kao.WrappedDataKey), &dataPrkWrapping); err != nil {
				log.Error("failed to unmarshal data private key wrapping: %v\n", err)
				return "", fmt.Errorf("failed to unmarshal data private key wrapping: %v", err)
			}

			providerPbk, _ := base64.StdEncoding.DecodeString(dataPrkWrapping.ProviderPublicKeyBase64)

			if dagMsg.AccessUrl == "" {
				log.Error("access url is empty, please check with data provider")
				return "", fmt.Errorf("access url is empty, please check with data provider")
			}

			var err error
			ztdoPath, err := utils.DownloadFileToTemp(dagMsg.AccessUrl, "ztdo-")
			if err != nil {
				log.Error("failed to download ztdo: %v\n", err)
				return "", fmt.Errorf("failed to download ztdo: %v", err)
			}

			if err := ztdo.ParseHeader(ztdoPath); err != nil {
				fmt.Printf("Error: failed to parse ztdo header:%s\n", err)
				return "", fmt.Errorf("failed to parse ztdo header:%s", err)
			}

			if ztdoId != ztdo.GetObjectID() {
				fmt.Printf("Error: ztdo id mismatch, please check with data provider\n")
				return "", fmt.Errorf("ztdo id mismatch, please check with data provider")
			}

			// decrypt data private key
			saDataPrk := ztdolib.NewSymmetricAgreement(ztdo.GetECCMode(), false)
			saDataPrk.SetMessagePatterns(ztdolib.DataPrivateKeyWrappingPatterns)
			saDataPrk.SetPsk([]byte(ztdolib.InitialDHPKeyWrappingString))
			saDataPrk.SetStaticKeyPair(teeEcdh)
			saDataPrk.SetEphemeralKeyPair(consumerEphemeralEcdh)
			saDataPrk.SetRemoteStaticPublicKey(providerPbk)

			gcmKey, ad := saDataPrk.AgreeSymmetricKey()

			dataPrkBase64, err := dataPrkWrapping.Unwrap(gcmKey[:], ad)
			if err != nil {
				return "", fmt.Errorf("failed to unwrap data private key: %s", err)
			}

			if ztdoPath == "" || output == "" {
				return "", fmt.Errorf("ztdo path or output is empty")
			}

			// decrypt data
			dataKeyPairEccMode := ztdo.GetECCMode()

			dataMsgPattern := [][]ztdolib.MessagePattern{
				{ztdolib.MessagePatternS, ztdolib.MessagePatternDHSS},
				{ztdolib.MessagePatternRS, ztdolib.MessagePatternDHSS},
			}

			dataPrk, _ := base64.StdEncoding.DecodeString(dataPrkBase64)
			saData := ztdolib.NewSymmetricAgreement(dataKeyPairEccMode, false)
			saData.SetMessagePatterns(dataMsgPattern)
			saData.SetStaticKeyPair(core.ECDHFromKey(dataKeyPairEccMode.ToEccType(), dataPrk))

			providerPublicKey, _ := base64.StdEncoding.DecodeString(dataPrkWrapping.ProviderPublicKeyBase64)
			saData.SetRemoteStaticPublicKey(providerPublicKey)

			gcmKey, ad = saData.AgreeSymmetricKey()

			if err := ztdo.DecryptZtdoFile(ztdoPath, output, gcmKey[:], ad); err != nil {
				return "", fmt.Errorf("Failed to decrypt ztdo file: %v", err)
			} else {
				a.decryptedZtdoRecord[ztdoId] = output
			}
		} else {
			output = decryptedOutput
		}
	} else {
		teeNotAuthorizedCode, _ := strconv.Atoi(common.ErrTEENotAuthorized.ErrorCode())
		if dagMsg.ErrCode == teeNotAuthorizedCode {
			a.trustedByNHPDB.Store(false)
		}

		return "", fmt.Errorf("Error: fail to request ztdo with error: %s.", dagMsg.ErrMsg)
	}
	return output, nil
}

func (a *UdpAgent) GetFirstServerPeer() (serverPeer *core.UdpPeer) {
	for _, value := range a.serverPeerMap {
		serverPeer = value
		return serverPeer
	}
	return nil
}

func (a *UdpAgent) SendDARMsgToServer(server *core.UdpPeer, msg common.DARMsg) (bool, *common.DAGMsg) {
	result := false
	sendAddr := server.SendAddr()
	if sendAddr == nil {
		log.Critical("device(%v)[SendDARMsgToServer] register server IP cannot be parsed", a)
	}
	drgMsg := msg
	drgBytes, _ := json.Marshal(drgMsg)
	drgMd := &core.MsgData{
		RemoteAddr:    sendAddr.(*net.UDPAddr),
		HeaderType:    core.NHP_DAR,
		TransactionId: a.device.NextCounterIndex(),
		Compress:      true,
		Message:       drgBytes,
		PeerPk:        server.PublicKey(),
		ResponseMsgCh: make(chan *core.PacketParserData),
	}

	currTime := time.Now().UnixNano()
	if !a.IsRunning() {
		log.Error("server-agentMsgData channel closed or being closed, skip sending")
		return result, nil
	}
	// device will create or find existing connection and sends the MsgAssembler via that connection
	a.sendMsgCh <- drgMd
	server.UpdateSend(currTime)
	// block until transaction completes
	serverPpd := <-drgMd.ResponseMsgCh
	close(drgMd.ResponseMsgCh)

	//Wait for NHP-Server response and implement reception and processing within the func() function below.
	var err error
	result, dsaMsg := func() (bool, *common.DSAMsg) {
		dsaMsg := &common.DSAMsg{}
		if serverPpd.Error != nil {
			log.Error("Agent(%s#%d)[SendDARMsgToServer] failed to receive response from server %s: %v", drgMsg.DoId, drgMd.TransactionId, server.Ip, serverPpd.Error)
			err = serverPpd.Error
			return false, dsaMsg
		}

		if serverPpd.HeaderType != core.NHP_DSA {
			log.Error("DB(%s#%d)[SendDARMsgToServer] response from server %s has wrong type: %s", drgMsg.DoId, drgMd.TransactionId, server.Ip, core.HeaderTypeToString(serverPpd.HeaderType))
			err = common.ErrTransactionRepliedWithWrongType
			return false, dsaMsg
		}
		//message []byte to DSAMSg Object
		err = json.Unmarshal(serverPpd.BodyMessage, dsaMsg)
		if err != nil {
			log.Error("Agent(%s#%d)[HandleDHPDAGMessage] failed to parse %s message: %v", drgMsg.DoId, serverPpd.SenderTrxId, core.HeaderTypeToString(serverPpd.HeaderType), err)
			return false, dsaMsg
		}
		dsaMsgString, err := json.Marshal(dsaMsg)
		if err != nil {
			log.Error("Agent(%s) DSAMsg failed to parse message: %v", dsaMsg.DoId, err)
			return false, dsaMsg
		}
		log.Info("SendDARMsgToServer response result: %v", dsaMsgString)
		if dsaMsg.ErrCode != 0 {
			log.Error("SendDARMsgToServer send failed, error: %s", dsaMsg.ErrMsg)
			return false, dsaMsg
		}
		return true, dsaMsg
	}()

	if result {
		// clear related resources when load new smart data policy
		if spoId, exist := a.smartPolicyIdentifier[dsaMsg.DoId]; exist {
			if _, exist := a.smartPolicyEngine[spoId]; exist {
				a.smartPolicyEngine[spoId].Close()
				delete(a.smartPolicyEngine, spoId)
			}
			delete(a.smartPolicyIdentifier, dsaMsg.DoId)
		}
		a.smartPolicyIdentifier[dsaMsg.DoId] = dsaMsg.Spo.PolicyId

		// Collect attestation proofs with smart policy
		evidence, err := a.onAttestationCollect(dsaMsg.Spo)
		if err != nil {
			dagMsg := &common.DAGMsg{}
			dagMsg.DoId = dsaMsg.DoId
			dagMsg.ErrCode = 1
			dagMsg.ErrMsg = err.Error()

			return false, dagMsg
		}

		// avoid flood attack from server side
		time.Sleep(core.MinimalRecvIntervalMs * time.Millisecond)

		davMsg := common.DAVMsg{
			DoId:     msg.DoId,
			SpoId:    dsaMsg.SpoId,
			Evidence: evidence,
		}

		return a.SendDAVMsgToServer(server, davMsg)
	} else {
		dagMsg := &common.DAGMsg{}
		dagMsg.DoId = dsaMsg.DoId
		dagMsg.ErrCode = dsaMsg.ErrCode
		dagMsg.ErrMsg = dsaMsg.ErrMsg

		return result, dagMsg
	}
}

func (a *UdpAgent) SendDAVMsgToServer(server *core.UdpPeer, msg common.DAVMsg) (bool, *common.DAGMsg) {
	result := false
	sendAddr := server.SendAddr()
	if sendAddr == nil {
		log.Critical("device(%v)[SendDAVMsgToServer] register server IP cannot be parsed", a)
	}
	davMsg := msg
	davBytes, _ := json.Marshal(davMsg)
	davMd := &core.MsgData{
		RemoteAddr:    sendAddr.(*net.UDPAddr),
		HeaderType:    core.NHP_DAV,
		TransactionId: a.device.NextCounterIndex(),
		Compress:      true,
		Message:       davBytes,
		PeerPk:        server.PublicKey(),
		ResponseMsgCh: make(chan *core.PacketParserData),
	}

	currTime := time.Now().UnixNano()
	if !a.IsRunning() {
		log.Error("server-agentMsgData channel closed or being closed, skip sending")
		return result, nil
	}
	// device will create or find existing connection and sends the MsgAssembler via that connection
	a.sendMsgCh <- davMd
	server.UpdateSend(currTime)
	// block until transaction completes
	serverPpd := <-davMd.ResponseMsgCh
	close(davMd.ResponseMsgCh)

	//Wait for NHP-Server response and implement reception and processing within the func() function below.
	var err error
	result, dagMsg := func() (bool, *common.DAGMsg) {
		dagMsg := &common.DAGMsg{}
		if serverPpd.Error != nil {
			log.Error("Agent(%s#%d)[SendDAVMsgToServer] failed to receive response from server %s: %v", davMsg.DoId, davMd.TransactionId, server.Ip, serverPpd.Error)
			err = serverPpd.Error
			return false, dagMsg
		}

		if serverPpd.HeaderType != core.NHP_DAG {
			log.Error("DB(%s#%d)[SendDAVMsgToServer] response from server %s has wrong type: %s", davMsg.DoId, davMd.TransactionId, server.Ip, core.HeaderTypeToString(serverPpd.HeaderType))
			err = common.ErrTransactionRepliedWithWrongType
			return false, dagMsg
		}
		//message []byte to DAGMSg Object
		err = json.Unmarshal(serverPpd.BodyMessage, dagMsg)
		if err != nil {
			log.Error("Agent(%s#%d)[HandleDHPDAVMessage] failed to parse %s message: %v", davMsg.DoId, serverPpd.SenderTrxId, core.HeaderTypeToString(serverPpd.HeaderType), err)
			return false, dagMsg
		}
		dagMsgString, err := json.Marshal(dagMsg)
		if err != nil {
			log.Error("Agent(%s) DAKMsg failed to parse message: %v", dagMsg.DoId, err)
			return false, dagMsg
		}
		log.Info("SendDAVMsgToServer response result: %v", dagMsgString)
		if dagMsg.ErrCode != 0 {
			log.Error("SendDAVMsgToServer send failed, error: %s", dagMsg.ErrMsg)
			return false, dagMsg
		}
		return true, dagMsg
	}()
	return result, dagMsg
}

func (s *UdpAgent) onAttestationCollect(spo *common.SmartPolicy) (string, error) {
	if spo.Policy == "" {
		return "", nil
	}

	wasmBytes, err := spo.GetPolicy()
	if err != nil {
		return "", err
	}

	engine := wasmEngine.NewEngine()
	err = engine.LoadWasm(wasmBytes)
	if err != nil {
		return "", err
	}

	s.smartPolicyEngine[spo.PolicyId] = engine

	attestation := engine.OnAttestationCollect()

	return attestation, nil
}

func (a *UdpAgent) CallTrustedApplication(taId string, function string, params map[string]any, spoId string) (string, error) {
	ta, err := GetTrustedApplication(taId)
	if err != nil {
		return "", err
	}

	taRes, err := ta.CallFunction(function, params)
	if err != nil {
		return "", err
	} else {
		if spEngine, exist := a.smartPolicyEngine[spoId]; exist {
			resultWithPostProcess := spEngine.OnDataPostprocess(taRes)
			return resultWithPostProcess, nil
		} else {
			return taRes, nil
		}
	}
}
