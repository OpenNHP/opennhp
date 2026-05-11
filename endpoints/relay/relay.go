// Package relay implements the NHP HTTP Relay service.
//
// The relay bridges browser-based NHP agents (which cannot send raw UDP) to
// the NHP Server using the standard NHP protocol:
//
//	Browser ──HTTPS POST──▶ NHP Relay ──NHP_RLY (encrypted)──▶ NHP Server
//	                        ◀──NHP response (encrypted)────────
//	        ◀──HTTP 200────
//
// The relay is a standard NHP node: it holds a core.Device, maintains a
// persistent UDP connection to each configured upstream cluster, sends
// encrypted NHP_RLY messages through the standard Noise pipeline, and uses
// NHP_KPL keepalive packets to maintain those connections — identical to how
// NHP-AC communicates.
//
// One relay instance can serve multiple logical nhp-server clusters. Each
// cluster is identified by the fingerprint of its public key (see
// utils.PubKeyFingerprint), and HTTP clients address a cluster via
// `POST /relay/{clusterId}`. Phase 1 supports one upstream instance per
// cluster; the schema is already shaped for the multi-instance future.
package relay

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	log "github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/utils"
)

const (
	// maxPacketSize is the maximum inner NHP packet size accepted from browsers.
	maxPacketSize = 4096

	// defaultUDPTimeoutMs for individual relay requests when waiting for server response.
	defaultUDPTimeoutMs = 5000

	// PacketQueueSizePerConnection mirrors the AC's queue size.
	PacketQueueSizePerConnection = 16

	// DefaultConnectionTimeoutMs is the idle timeout for the persistent connection.
	DefaultConnectionTimeoutMs = 120000
)

// UdpConn wraps a UDP connection with NHP connection data.
type UdpConn struct {
	ConnData *core.ConnectionData
	netConn  *net.UDPConn
}

func (c *UdpConn) Close() {
	if c.netConn != nil {
		c.netConn.Close()
		c.ConnData.Close()
	}
}

// clusterInstance is the runtime state for one physical nhp-server instance.
// Phase 1: every cluster has exactly one of these. Phase 2 will add health
// state and load-balancing bookkeeping here.
type clusterInstance struct {
	host   string
	port   int
	weight int
	addr   *net.UDPAddr

	connMu sync.Mutex
	conn   *UdpConn

	// pendingRequests correlates server responses back to the HTTP handler
	// that issued each forward. It is per-instance because the connection
	// that delivers the ACK/COK is also per-instance — keying it here
	// means dispatchResponse never has to think about which cluster a
	// packet belongs to.
	//
	// The map is keyed by (counter, realClientAddr); the same hijack-
	// prevention rule as before applies: ambiguous waiters all time out.
	pendingMu       sync.Mutex
	pendingRequests map[uint64]map[string]chan []byte
}

// clusterRuntime is the runtime state for one logical nhp-server cluster.
type clusterRuntime struct {
	id           string
	name         string
	pubKey       []byte
	pubKeyBase64 string
	scheme       LoadBalanceScheme
	instances    []*clusterInstance
}

// pickInstance selects an instance to handle a request. Phase 1 always
// returns instances[0]; the signature is shaped so phase 2 can plug in
// random / weighted / round-robin without touching call sites.
func (c *clusterRuntime) pickInstance() *clusterInstance {
	return c.instances[0]
}

// RelayServer is the NHP HTTP Relay service.
type RelayServer struct {
	config     *Config
	httpServer *http.Server
	device     *core.Device

	clusters  map[string]*clusterRuntime // keyed by fingerprint
	defaultID string                     // optional default cluster fingerprint

	sendMsgCh chan *core.MsgData
	recvMsgCh <-chan *core.PacketParserData

	wg      sync.WaitGroup
	running atomic.Bool
	stopCh  chan struct{}

	stats struct {
		totalRecvBytes uint64
		totalSendBytes uint64
	}
}

// New creates a RelayServer from the given configuration.
func New(cfg *Config) (*RelayServer, error) {
	// Decode relay private key.
	prk, err := base64.StdEncoding.DecodeString(cfg.PrivateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("relay: invalid privateKeyBase64: %w", err)
	}

	// Create NHP device with relay identity.
	device := core.NewDevice(core.NHP_RELAY, prk, nil)
	if device == nil {
		return nil, fmt.Errorf("relay: failed to create NHP device")
	}

	rs := &RelayServer{
		config:    cfg,
		device:    device,
		clusters:  make(map[string]*clusterRuntime, len(cfg.Clusters)),
		defaultID: cfg.DefaultClusterID,
		sendMsgCh: make(chan *core.MsgData, PacketQueueSizePerConnection),
		stopCh:    make(chan struct{}),
	}
	rs.recvMsgCh = device.DecryptedMsgQueue

	for i := range cfg.Clusters {
		c := &cfg.Clusters[i]
		cr, err := rs.buildCluster(c)
		if err != nil {
			return nil, err
		}
		rs.clusters[cr.id] = cr
	}

	// Single-default-cluster shortcut: if exactly one cluster is configured
	// and no explicit DefaultClusterID, treat that cluster as default so
	// the legacy `POST /relay` path keeps working.
	if rs.defaultID == "" && len(rs.clusters) == 1 {
		for id := range rs.clusters {
			rs.defaultID = id
		}
	}

	// Set up HTTP server.
	mux := http.NewServeMux()
	// Trailing slash on /relay/ enables {clusterId} routing via path prefix.
	mux.HandleFunc("/relay", rs.handleRelay)
	mux.HandleFunc("/relay/", rs.handleRelay)
	mux.HandleFunc("/clusters", rs.handleClusters)
	mux.HandleFunc("/health", rs.handleHealth)

	rs.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.ListenIP, cfg.ListenPort),
		Handler:      corsMiddleware(mux),
		ReadTimeout:  time.Duration(cfg.ReadTimeoutMs) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.WriteTimeoutMs) * time.Millisecond,
		IdleTimeout:  time.Duration(cfg.IdleTimeoutMs) * time.Millisecond,
	}

	log.Info("[Relay] initialized, relay pubkey=%s, %d cluster(s)",
		device.PublicKeyBase64(), len(rs.clusters))
	for _, cr := range rs.clusters {
		inst := cr.instances[0]
		log.Info("[Relay]   cluster id=%s name=%q upstream=%s:%d",
			cr.id, cr.name, inst.host, inst.port)
	}
	return rs, nil
}

// buildCluster turns a config Cluster into runtime state, registering each
// instance as a peer on the NHP device.
func (rs *RelayServer) buildCluster(c *Cluster) (*clusterRuntime, error) {
	pubKey, err := base64.StdEncoding.DecodeString(c.PublicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("relay: cluster %q invalid publicKeyBase64: %w", c.Name, err)
	}
	id := utils.PubKeyFingerprint(pubKey)

	cr := &clusterRuntime{
		id:           id,
		name:         c.Name,
		pubKey:       pubKey,
		pubKeyBase64: c.PublicKeyBase64,
		scheme:       c.LoadBalance,
		instances:    make([]*clusterInstance, 0, len(c.Instances)),
	}

	for j := range c.Instances {
		ci := &c.Instances[j]
		addrStr := fmt.Sprintf("%s:%d", ci.Host, ci.Port)
		udpAddr, err := net.ResolveUDPAddr("udp", addrStr)
		if err != nil {
			return nil, fmt.Errorf("relay: cluster %s instance #%d resolve %s: %w", id, j, addrStr, err)
		}

		// Register this instance as a peer. Phase 1: at most one peer
		// per cluster, so device.AddPeer keyed-by-pubkey is fine.
		peer := &core.UdpPeer{
			PubKeyBase64: c.PublicKeyBase64,
			Ip:           ci.Host,
			Port:         ci.Port,
			Type:         core.NHP_SERVER,
		}
		rs.device.AddPeer(peer)

		cr.instances = append(cr.instances, &clusterInstance{
			host:            ci.Host,
			port:            ci.Port,
			weight:          ci.Weight,
			addr:            udpAddr,
			pendingRequests: make(map[uint64]map[string]chan []byte),
		})
	}
	return cr, nil
}

// Start starts the device, UDP connections, keepalives, and HTTP server.
func (rs *RelayServer) Start() error {
	rs.running.Store(true)

	// Start NHP device (encryption/decryption workers).
	rs.device.Start()

	// Start send/recv message routines.
	rs.wg.Add(2)
	go rs.sendMessageRoutine()
	go rs.recvMessageRoutine()

	// One keepalive goroutine per instance. Cheap (mostly sleeping) and
	// keeps the per-instance state owned by exactly one routine.
	for _, cr := range rs.clusters {
		for _, inst := range cr.instances {
			rs.wg.Add(1)
			go rs.keepaliveRoutine(cr, inst)
		}
	}

	addr := rs.httpServer.Addr
	if rs.config.EnableTLS {
		log.Info("[Relay] starting HTTPS relay on %s", addr)
		tlsCfg := &tls.Config{MinVersion: tls.VersionTLS13}
		rs.httpServer.TLSConfig = tlsCfg
		return rs.httpServer.ListenAndServeTLS(rs.config.TLSCertFile, rs.config.TLSKeyFile)
	}
	log.Info("[Relay] starting HTTP relay on %s", addr)
	return rs.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the relay service.
func (rs *RelayServer) Stop(ctx context.Context) error {
	rs.running.Store(false)
	close(rs.stopCh)

	// Shut down HTTP server.
	err := rs.httpServer.Shutdown(ctx)

	// Close UDP connections.
	for _, cr := range rs.clusters {
		for _, inst := range cr.instances {
			inst.connMu.Lock()
			if inst.conn != nil {
				inst.conn.Close()
			}
			inst.connMu.Unlock()
		}
	}

	// Stop device.
	rs.device.Stop()

	rs.wg.Wait()
	return err
}

// ──────────────────────────────────────────────────────────────────────────────
// UDP connection management (per upstream instance)
// ──────────────────────────────────────────────────────────────────────────────

func (rs *RelayServer) getOrCreateConnection(cr *clusterRuntime, inst *clusterInstance) *UdpConn {
	inst.connMu.Lock()
	defer inst.connMu.Unlock()

	if inst.conn != nil {
		return inst.conn
	}

	conn := &UdpConn{}
	var err error
	conn.netConn, err = net.DialUDP("udp", nil, inst.addr)
	if err != nil {
		log.Error("[Relay] cluster %s: failed to dial %s: %v", cr.id, inst.addr, err)
		return nil
	}

	laddr := conn.netConn.LocalAddr()
	localAddr, err := net.ResolveUDPAddr(laddr.Network(), laddr.String())
	if err != nil {
		log.Error("[Relay] cluster %s: resolve local addr error: %v", cr.id, err)
		conn.netConn.Close()
		return nil
	}

	log.Info("[Relay] cluster %s: new UDP connection %s → %s", cr.id, localAddr, inst.addr)

	conn.ConnData = &core.ConnectionData{
		Device:               rs.device,
		CookieStore:          &core.CookieStore{},
		RemoteTransactionMap: make(map[uint64]*core.RemoteTransaction),
		LocalAddr:            localAddr,
		RemoteAddr:           inst.addr,
		TimeoutMs:            DefaultConnectionTimeoutMs,
		SendQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
		RecvQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
		BlockSignal:          make(chan struct{}),
		SetTimeoutSignal:     make(chan struct{}),
		StopSignal:           make(chan struct{}),
	}

	// Start recv and connection routines. Both track rs.wg so Stop() fully
	// drains them before returning.
	conn.ConnData.Add(1)
	rs.wg.Add(1)
	go func() {
		defer rs.wg.Done()
		rs.recvPacketRoutine(cr, inst, conn)
	}()

	rs.wg.Add(1)
	go rs.connectionRoutine(cr, inst, conn)

	inst.conn = conn
	return conn
}

// ──────────────────────────────────────────────────────────────────────────────
// Packet send/recv routines
// ──────────────────────────────────────────────────────────────────────────────

func (rs *RelayServer) sendPacket(pkt *core.Packet, conn *UdpConn) (int, error) {
	defer func() {
		atomic.StoreInt64(&conn.ConnData.LastLocalSendTime, time.Now().UnixNano())
		if !pkt.KeepAfterSend {
			rs.device.ReleasePoolPacket(pkt)
		}
	}()

	pktType := core.HeaderTypeToString(pkt.HeaderType)
	log.Info("[Relay] send [%s] packet (%s -> %s), %d bytes",
		pktType, conn.ConnData.LocalAddr, conn.ConnData.RemoteAddr, len(pkt.Content))
	n, err := conn.netConn.Write(pkt.Content)
	atomic.AddUint64(&rs.stats.totalSendBytes, uint64(n))
	return n, err
}

func (rs *RelayServer) recvPacketRoutine(cr *clusterRuntime, inst *clusterInstance, conn *UdpConn) {
	addrStr := conn.ConnData.RemoteAddr.String()
	defer conn.ConnData.Done()
	defer log.Info("[Relay] recvPacketRoutine for %s (cluster %s) stopped", addrStr, cr.id)

	log.Info("[Relay] recvPacketRoutine for %s (cluster %s) started", addrStr, cr.id)

	for {
		select {
		case <-conn.ConnData.StopSignal:
			return
		default:
		}

		pkt := rs.device.AllocatePoolPacket()
		n, err := conn.netConn.Read(pkt.Buf[:])
		if err != nil {
			rs.device.ReleasePoolPacket(pkt)
			if n == 0 {
				return
			}
			log.Error("[Relay] recv error from %s: %v", addrStr, err)
			continue
		}

		atomic.AddUint64(&rs.stats.totalRecvBytes, uint64(n))

		if n < pkt.MinimalLength() {
			rs.device.ReleasePoolPacket(pkt)
			log.Error("[Relay] packet from %s too short, discard", addrStr)
			continue
		}

		pkt.Content = pkt.Buf[:n]

		typ, _, err := rs.device.RecvPrecheck(pkt)
		msgType := core.HeaderTypeToString(typ)
		log.Info("[Relay] recv [%s] packet (%s -> %s), %d bytes",
			msgType, addrStr, conn.ConnData.LocalAddr, n)
		if err != nil {
			rs.device.ReleasePoolPacket(pkt)
			log.Warning("[Relay] recv [%s] precheck error: %v", msgType, err)
			continue
		}

		atomic.StoreInt64(&conn.ConnData.LastLocalRecvTime, time.Now().UnixNano())
		conn.ConnData.ForwardInboundPacket(pkt)
	}
}

// dispatchResponse routes a server ACK/COK payload back to the HTTP handler
// waiting on the given counter for the given instance. See clusterInstance
// docs for the hijack-prevention rule (ambiguous waiters are all dropped).
//
// Returns (delivered, ambiguous). If both are false the counter was unknown
// (e.g. a late response after the handler timed out).
func (rs *RelayServer) dispatchResponse(inst *clusterInstance, counter uint64, raw []byte) (delivered, ambiguous bool) {
	inst.pendingMu.Lock()
	waiters, found := inst.pendingRequests[counter]
	if !found {
		inst.pendingMu.Unlock()
		return false, false
	}
	if len(waiters) != 1 {
		log.Warning("[Relay] ambiguous response for counter=%d (%d waiters); dropping to prevent hijack",
			counter, len(waiters))
		inst.pendingMu.Unlock()
		return false, true
	}
	var ch chan []byte
	for _, c := range waiters {
		ch = c
	}
	delete(inst.pendingRequests, counter)
	inst.pendingMu.Unlock()

	// Handler channels are buffered (size 1) and the handler only registers
	// one request at a time, so this send never blocks.
	ch <- raw
	return true, false
}

func (rs *RelayServer) connectionRoutine(cr *clusterRuntime, inst *clusterInstance, conn *UdpConn) {
	addrStr := conn.ConnData.RemoteAddr.String()
	defer rs.wg.Done()
	defer log.Info("[Relay] connectionRoutine for %s (cluster %s) stopped", addrStr, cr.id)

	log.Info("[Relay] connectionRoutine for %s (cluster %s) started", addrStr, cr.id)

	defer func() {
		inst.connMu.Lock()
		if inst.conn == conn {
			inst.conn = nil
		}
		inst.connMu.Unlock()
		conn.Close()
	}()

	for {
		select {
		case <-rs.stopCh:
			return

		case <-time.After(time.Duration(conn.ConnData.TimeoutMs) * time.Millisecond):
			log.Info("[Relay] connection idle timeout (cluster %s)", cr.id)
			return

		case pkt, ok := <-conn.ConnData.SendQueue:
			if !ok {
				return
			}
			if pkt == nil {
				continue
			}
			_, _ = rs.sendPacket(pkt, conn)

		case pkt, ok := <-conn.ConnData.RecvQueue:
			if !ok {
				return
			}
			if pkt == nil {
				continue
			}

			// Handle keepalive.
			if pkt.HeaderType == core.NHP_KPL {
				rs.device.ReleasePoolPacket(pkt)
				log.Info("[Relay] recv [NHP_KPL] from %s (cluster %s)", addrStr, cr.id)
				continue
			}

			// Check if this is a response (ACK/COK) for a pending relay
			// request on this instance.
			if pkt.HeaderType == core.NHP_ACK || pkt.HeaderType == core.NHP_COK {
				counter := pkt.Counter()
				// Copy raw bytes before releasing the pool packet — dispatch
				// sends them into a handler channel.
				raw := make([]byte, len(pkt.Content))
				copy(raw, pkt.Content)
				delivered, ambiguous := rs.dispatchResponse(inst, counter, raw)
				if delivered {
					log.Info("[Relay] matched pending request counter=%d on cluster %s, forwarding %d raw bytes",
						counter, cr.id, len(raw))
					rs.device.ReleasePoolPacket(pkt)
					continue
				}
				if ambiguous {
					rs.device.ReleasePoolPacket(pkt)
					continue
				}
			}

			// Generic receive → decrypt.
			pd := &core.PacketData{
				BasePacket: pkt,
				ConnData:   conn.ConnData,
				InitTime:   atomic.LoadInt64(&conn.ConnData.LastLocalRecvTime),
			}
			rs.device.RecvPacketToMsg(pd)

		case <-conn.ConnData.BlockSignal:
			log.Warning("[Relay] connection blocked %s (cluster %s)", addrStr, cr.id)
			return
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Message send/recv routines
// ──────────────────────────────────────────────────────────────────────────────

func (rs *RelayServer) sendMessageRoutine() {
	defer rs.wg.Done()
	defer log.Info("[Relay] sendMessageRoutine stopped")

	log.Info("[Relay] sendMessageRoutine started")

	for {
		select {
		case <-rs.stopCh:
			return

		case md, ok := <-rs.sendMsgCh:
			if !ok {
				return
			}
			if md == nil {
				continue
			}
			rs.dispatchSend(md)
		}
	}
}

// dispatchSend resolves the outbound MsgData to a (cluster, instance) by
// matching md.PeerPk / md.RemoteAddr to the registered clusters, then
// attaches the right ConnData and forwards through the device.
//
// We chose to keep core.MsgData unmodified and route here, rather than
// invent a relay-specific Send path: the device handles the cipher pipeline
// uniformly this way.
func (rs *RelayServer) dispatchSend(md *core.MsgData) {
	cr, inst := rs.resolveTarget(md)
	if cr == nil || inst == nil {
		log.Error("[Relay] dropping outbound: cannot match MsgData to a cluster (peer=%x remote=%v)",
			md.PeerPk, md.RemoteAddr)
		return
	}

	conn := rs.getOrCreateConnection(cr, inst)
	if conn == nil {
		log.Error("[Relay] cluster %s: no connection, dropping message", cr.id)
		return
	}

	md.ConnData = conn.ConnData
	rs.device.SendMsgToPacket(md)
}

// resolveTarget figures out which cluster instance an outbound MsgData is
// destined for. We match on RemoteAddr first (exact), falling back to PeerPk
// (cluster pubkey). PeerPk is required for keepalive packets, which carry
// only the address; RemoteAddr is required for relay forwards because the
// handler set it explicitly.
func (rs *RelayServer) resolveTarget(md *core.MsgData) (*clusterRuntime, *clusterInstance) {
	if md.RemoteAddr != nil {
		want := md.RemoteAddr.String()
		for _, cr := range rs.clusters {
			for _, inst := range cr.instances {
				if inst.addr.String() == want {
					return cr, inst
				}
			}
		}
	}
	if len(md.PeerPk) > 0 {
		fp := utils.PubKeyFingerprint(md.PeerPk)
		if cr, ok := rs.clusters[fp]; ok {
			return cr, cr.pickInstance()
		}
	}
	return nil, nil
}

func (rs *RelayServer) recvMessageRoutine() {
	defer rs.wg.Done()
	defer log.Info("[Relay] recvMessageRoutine stopped")

	log.Info("[Relay] recvMessageRoutine started")

	for {
		select {
		case <-rs.stopCh:
			return

		case ppd, ok := <-rs.recvMsgCh:
			if !ok {
				return
			}
			if ppd == nil {
				continue
			}

			log.Info("[Relay] recv decrypted message type [%s]",
				core.HeaderTypeToString(ppd.HeaderType))
			// Relay doesn't expect messages from server in the current design.
			// Future: handle server-initiated commands here.
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Keepalive routine (per instance)
// ──────────────────────────────────────────────────────────────────────────────

func (rs *RelayServer) keepaliveRoutine(cr *clusterRuntime, inst *clusterInstance) {
	defer rs.wg.Done()
	defer log.Info("[Relay] keepaliveRoutine for cluster %s instance %s stopped", cr.id, inst.addr)

	log.Info("[Relay] keepaliveRoutine for cluster %s instance %s started", cr.id, inst.addr)

	interval := rs.config.KeepaliveIntervalS
	if interval <= 0 {
		interval = common.ServerKeepaliveInterval
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-rs.stopCh:
			return

		case <-ticker.C:
			inst.connMu.Lock()
			conn := inst.conn
			inst.connMu.Unlock()

			if conn == nil {
				// No connection yet; getOrCreateConnection will be called on first request.
				continue
			}

			lastSend := atomic.LoadInt64(&conn.ConnData.LastLocalSendTime)
			if (time.Now().UnixNano() - lastSend) > int64(time.Duration(interval)*time.Second) {
				md := &core.MsgData{
					RemoteAddr:    inst.addr,
					HeaderType:    core.NHP_KPL,
					CipherScheme:  rs.config.CipherScheme,
					TransactionId: rs.device.NextCounterIndex(),
					PeerPk:        cr.pubKey,
				}
				// Non-blocking send: if the queue is full, drop this
				// keepalive rather than stall the routine (which would
				// also miss stopCh on shutdown). A missed keepalive is
				// recovered on the next tick.
				select {
				case rs.sendMsgCh <- md:
					log.Info("[Relay] sent NHP_KPL keepalive to %s (cluster %s)", inst.addr, cr.id)
				case <-rs.stopCh:
					return
				default:
					log.Warning("[Relay] send queue full, skipping keepalive (cluster %s)", cr.id)
				}
			}
		}
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// HTTP handlers
// ──────────────────────────────────────────────────────────────────────────────

func (rs *RelayServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// clusterInfo is the JSON shape returned by /clusters. It surfaces only the
// non-secret routing metadata clients need to choose a cluster (pubkey and
// the fingerprint derived from it). Instance addresses are intentionally
// omitted because clients route by cluster ID, not by instance.
type clusterInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	PublicKey   string `json:"publicKeyBase64"`
	IsDefault   bool   `json:"isDefault,omitempty"`
	InstanceCnt int    `json:"instanceCount"`
	LoadBalance string `json:"loadBalance,omitempty"`
}

// handleClusters lists every configured cluster. The endpoint is intentionally
// unauthenticated: a client that wants to knock must already know the target
// cluster's public key (it's required to encrypt the KNK packet), so exposing
// the pubkey + the routing id derived from it leaks nothing a determined
// caller couldn't recompute. The "name" field is an operator-chosen label —
// keep it free of sensitive context. If a deployment ever needs cluster
// topology to be opaque to anonymous browsers, gate this handler behind the
// reverse proxy (e.g. nginx allow/deny) rather than adding auth here.
func (rs *RelayServer) handleClusters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	out := make([]clusterInfo, 0, len(rs.clusters))
	for _, cr := range rs.clusters {
		out = append(out, clusterInfo{
			ID:          cr.id,
			Name:        cr.name,
			PublicKey:   cr.pubKeyBase64,
			IsDefault:   cr.id == rs.defaultID,
			InstanceCnt: len(cr.instances),
			LoadBalance: string(cr.scheme),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// corsMiddleware adds CORS headers so browser-based NHP agents can reach the
// relay from any origin.  The relay is a public transport bridge — there is no
// session state to protect, so a permissive CORS policy is appropriate.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// resolveCluster picks the target cluster for an incoming HTTP request.
// Supports both `POST /relay` (uses DefaultClusterID) and `POST /relay/{id}`.
func (rs *RelayServer) resolveCluster(r *http.Request) (*clusterRuntime, int, string) {
	// Path may be "/relay", "/relay/", or "/relay/<id>". TrimPrefix handles
	// both registered ServeMux routes.
	id := strings.TrimPrefix(r.URL.Path, "/relay")
	id = strings.TrimPrefix(id, "/")
	id = strings.TrimSuffix(id, "/")

	if id == "" {
		if rs.defaultID == "" {
			return nil, http.StatusBadRequest,
				"no cluster ID in path and no defaultClusterId configured; use POST /relay/<clusterId>"
		}
		id = rs.defaultID
	}

	cr, ok := rs.clusters[id]
	if !ok {
		return nil, http.StatusNotFound, fmt.Sprintf("unknown cluster ID %q", id)
	}
	return cr, 0, ""
}

// handleRelay is the main relay endpoint.
//
// Expected request:
//
//	POST /relay/{clusterId}
//	Content-Type: application/octet-stream
//	Body: raw inner NHP packet bytes (KNK / RKN / etc., encrypted by agent)
//
// Legacy compatible:
//
//	POST /relay  — uses Config.DefaultClusterID; rejected if unset and >1 cluster exists.
//
// Response:
//
//	200 OK  — body contains raw NHP ACK / COK packet bytes (encrypted to agent)
//	400 Bad Request  — empty / over-size body, or missing cluster ID
//	404 Not Found    — unknown cluster ID
//	504 Gateway Timeout  — NHP Server did not respond in time
//	502 Bad Gateway  — internal error
func (rs *RelayServer) handleRelay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cr, status, errMsg := rs.resolveCluster(r)
	if cr == nil {
		http.Error(w, errMsg, status)
		return
	}
	inst := cr.pickInstance()

	// Read inner NHP packet from request body. Cap at maxPacketSize+1 so we
	// can reject oversize bodies without pulling an unbounded amount into
	// memory. A single r.Body.Read() is not guaranteed to return the full
	// payload; io.ReadAll drains until EOF.
	innerPacket, err := io.ReadAll(io.LimitReader(r.Body, int64(maxPacketSize)+1))
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if len(innerPacket) == 0 {
		http.Error(w, "empty packet", http.StatusBadRequest)
		return
	}
	if len(innerPacket) > maxPacketSize {
		http.Error(w, "packet too large", http.StatusBadRequest)
		return
	}
	n := len(innerPacket)

	// Extract the counter from the inner packet header (bytes [16:24], big-endian uint64).
	// The NHP server echoes this counter in its ACK/COK response, so we use it
	// to match the response back to this HTTP request.
	if n < 24 {
		http.Error(w, "inner packet too short", http.StatusBadRequest)
		return
	}
	innerCounter := binary.BigEndian.Uint64(innerPacket[16:24])

	realAddr, err := realClientAddr(r)
	if err != nil {
		log.Error("[Relay] %v", err)
		http.Error(w, "relay misconfigured: missing X-Real-IP header from local reverse proxy", http.StatusBadGateway)
		return
	}
	realAddrKey := realAddr.String()
	log.Info("[Relay] forwarding %d-byte inner packet (counter=%d, cluster=%s) from client %s to %s",
		n, innerCounter, cr.id, realAddr, inst.addr)

	// Register a pending request under (counter, realAddr) on the instance.
	// The connection routine dispatches the server's ACK/COK to this channel
	// only if this handler is the sole waiter on this counter — see the
	// ambiguity check in connectionRoutine above.
	responseCh := make(chan []byte, 1)
	inst.pendingMu.Lock()
	waiters, ok := inst.pendingRequests[innerCounter]
	if !ok {
		waiters = make(map[string]chan []byte)
		inst.pendingRequests[innerCounter] = waiters
	}
	if _, dup := waiters[realAddrKey]; dup {
		// Same client reusing the same counter concurrently — reject fast.
		inst.pendingMu.Unlock()
		http.Error(w, "duplicate in-flight counter", http.StatusConflict)
		return
	}
	waiters[realAddrKey] = responseCh
	inst.pendingMu.Unlock()

	// Ensure cleanup on timeout / early return.
	defer func() {
		inst.pendingMu.Lock()
		if waiters, ok := inst.pendingRequests[innerCounter]; ok {
			delete(waiters, realAddrKey)
			if len(waiters) == 0 {
				delete(inst.pendingRequests, innerCounter)
			}
		}
		inst.pendingMu.Unlock()
	}()

	// Construct RelayForwardMsg (standard JSON body).
	rlyMsg := &common.RelayForwardMsg{
		SourceAddr: &common.NetAddress{
			Ip:   realAddr.IP.String(),
			Port: realAddr.Port,
		},
		InnerPacket: base64.StdEncoding.EncodeToString(innerPacket),
	}
	msgBytes, err := json.Marshal(rlyMsg)
	if err != nil {
		log.Error("[Relay] failed to marshal RelayForwardMsg: %v", err)
		http.Error(w, "relay internal error", http.StatusBadGateway)
		return
	}

	// Send the NHP_RLY envelope to the chosen instance.
	trxId := rs.device.NextCounterIndex()
	md := &core.MsgData{
		RemoteAddr:    inst.addr,
		HeaderType:    core.NHP_RLY,
		CipherScheme:  rs.config.CipherScheme,
		TransactionId: trxId,
		Message:       msgBytes,
		PeerPk:        cr.pubKey,
	}

	udpTimeout := rs.config.UDPTimeoutMs
	if udpTimeout <= 0 {
		udpTimeout = defaultUDPTimeoutMs
	}

	// Hand the message to sendMessageRoutine. A naked send would block
	// indefinitely if the channel (capacity PacketQueueSizePerConnection)
	// is full — net/http's WriteTimeout closes the TCP connection but
	// does not unblock a goroutine parked on a channel send, so a slow
	// upstream server would silently leak handler goroutines under load.
	// Bound the wait by the same UDP timeout used for the response.
	select {
	case rs.sendMsgCh <- md:
	case <-r.Context().Done():
		log.Warning("[Relay] client disconnected before send queued (counter=%d, client %s, cluster %s)",
			innerCounter, realAddr, cr.id)
		return
	case <-time.After(time.Duration(udpTimeout) * time.Millisecond):
		log.Error("[Relay] send queue full for %dms, dropping forward (counter=%d, client %s, cluster %s)",
			udpTimeout, innerCounter, realAddr, cr.id)
		http.Error(w, "relay overloaded", http.StatusServiceUnavailable)
		return
	}

	// Wait for the raw encrypted ACK/COK packet from the server.
	select {
	case rawBytes := <-responseCh:
		log.Info("[Relay] received response for inner counter=%d, %d raw bytes, forwarding to client %s (cluster %s)",
			innerCounter, len(rawBytes), realAddr, cr.id)

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(rawBytes)

	case <-time.After(time.Duration(udpTimeout) * time.Millisecond):
		log.Warning("[Relay] timeout waiting for server response (inner counter=%d, client %s, cluster %s)",
			innerCounter, realAddr, cr.id)
		http.Error(w, "NHP Server timeout", http.StatusGatewayTimeout)
	}
}

// realClientAddr returns the originating address of an HTTP request as a
// *net.UDPAddr so it can be encoded in the RelayForwardMsg.
//
// When the direct TCP peer is on the loopback interface — i.e. a local
// reverse proxy (nginx, etc.) forwarded the request — the proxy's view
// of the real client is taken from X-Real-IP, which the proxy is
// expected to overwrite unconditionally (e.g. nginx
// `proxy_set_header X-Real-IP $remote_addr;`).
//
// X-Forwarded-For is intentionally NOT consulted: nginx's standard
// `$proxy_add_x_forwarded_for` *appends* to whatever XFF the client
// sent, so its first entry is attacker-controlled. Trusting XFF would
// let any HTTP client choose the SourceAddr that flows to nhp-server
// and ultimately to the AC ipset rule, defeating the per-source-IP
// authorization model.
//
// If the direct peer is on loopback but X-Real-IP is missing or
// malformed, the function returns an error rather than falling back to
// the loopback address. Falling back would set SourceAddr=127.0.0.1,
// which the server's isRoutablePublicIP check rejects, producing
// silent 504s that are hard to diagnose. A loud error here points
// operators at the misconfigured reverse proxy instead.
func realClientAddr(r *http.Request) (*net.UDPAddr, error) {
	// Parse the direct TCP peer first so we always have a port.
	peerHost, peerPortStr, err := net.SplitHostPort(r.RemoteAddr)
	peerIP := net.IPv4zero
	peerPort := 0
	if err == nil {
		if ip := net.ParseIP(peerHost); ip != nil {
			peerIP = ip
		}
		_, _ = fmt.Sscanf(peerPortStr, "%d", &peerPort)
	}

	if peerIP.IsLoopback() {
		realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		if realIP == "" {
			return nil, fmt.Errorf("loopback peer %s sent no X-Real-IP header; check reverse proxy config", r.RemoteAddr)
		}
		ip := net.ParseIP(realIP)
		if ip == nil {
			return nil, fmt.Errorf("loopback peer %s sent malformed X-Real-IP %q", r.RemoteAddr, realIP)
		}
		// X-Real-IP carries no port; the proxy peer's port is
		// used so connection-tracking keys remain unique.
		return &net.UDPAddr{IP: ip, Port: peerPort}, nil
	}

	return &net.UDPAddr{IP: peerIP, Port: peerPort}, nil
}
