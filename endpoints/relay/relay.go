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
// persistent UDP connection to the server, sends encrypted NHP_RLY messages
// through the standard Noise pipeline, and uses NHP_KPL keepalive packets
// to maintain the connection — identical to how NHP-AC communicates.
package relay

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	log "github.com/OpenNHP/opennhp/nhp/log"
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
	ConnData  *core.ConnectionData
	netConn   *net.UDPConn
	connected atomic.Bool
}

func (c *UdpConn) Close() {
	if c.netConn != nil {
		c.netConn.Close()
		c.ConnData.Close()
	}
}

// RelayServer is the NHP HTTP Relay service.
type RelayServer struct {
	config     *Config
	httpServer *http.Server
	device     *core.Device

	serverPubKey []byte       // decoded NHP server public key
	serverAddr   *net.UDPAddr // resolved NHP server UDP address

	connMutex sync.Mutex
	conn      *UdpConn // persistent UDP connection to server

	sendMsgCh chan *core.MsgData
	recvMsgCh <-chan *core.PacketParserData

	// pendingRelayRequests maps an inner packet's counter (uint64) to a channel
	// that receives the raw encrypted ACK/COK bytes from the server.  The HTTP
	// handler registers entries before sending the RLY envelope; the connection
	// routine fulfils them when a matching response arrives.
	pendingMu       sync.Mutex
	pendingRequests map[uint64]chan []byte

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

	// Decode NHP server public key.
	serverPubKey, err := base64.StdEncoding.DecodeString(cfg.NHPServerPublicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("relay: invalid nhpServerPublicKeyBase64: %w", err)
	}

	// Register server as a peer.
	serverPeer := &core.UdpPeer{
		PubKeyBase64: cfg.NHPServerPublicKeyBase64,
		Ip:           cfg.NHPServerHost,
		Port:         cfg.NHPServerPort,
		Type:         core.NHP_SERVER,
	}
	device.AddPeer(serverPeer)

	// Resolve server address.
	serverAddrStr := fmt.Sprintf("%s:%d", cfg.NHPServerHost, cfg.NHPServerPort)
	serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr)
	if err != nil {
		return nil, fmt.Errorf("relay: failed to resolve server address %s: %w", serverAddrStr, err)
	}

	rs := &RelayServer{
		config:          cfg,
		device:          device,
		serverPubKey:    serverPubKey,
		serverAddr:      serverAddr,
		sendMsgCh:       make(chan *core.MsgData, PacketQueueSizePerConnection),
		pendingRequests: make(map[uint64]chan []byte),
		stopCh:          make(chan struct{}),
	}
	rs.recvMsgCh = device.DecryptedMsgQueue

	// Set up HTTP server.
	mux := http.NewServeMux()
	mux.HandleFunc("/relay", rs.handleRelay)
	mux.HandleFunc("/health", rs.handleHealth)

	rs.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.ListenIP, cfg.ListenPort),
		Handler:      corsMiddleware(mux),
		ReadTimeout:  time.Duration(cfg.ReadTimeoutMs) * time.Millisecond,
		WriteTimeout: time.Duration(cfg.WriteTimeoutMs) * time.Millisecond,
		IdleTimeout:  time.Duration(cfg.IdleTimeoutMs) * time.Millisecond,
	}

	log.Info("[Relay] initialized, relay pubkey=%s", device.PublicKeyBase64())
	return rs, nil
}

// Start starts the device, UDP connection, keepalive, and HTTP server.
func (rs *RelayServer) Start() error {
	rs.running.Store(true)

	// Start NHP device (encryption/decryption workers).
	rs.device.Start()

	// Start send/recv message routines.
	rs.wg.Add(3)
	go rs.sendMessageRoutine()
	go rs.recvMessageRoutine()
	go rs.keepaliveRoutine()

	// Start HTTP server (blocks).
	addr := rs.httpServer.Addr
	if rs.config.EnableTLS {
		log.Info("[Relay] starting HTTPS relay on %s → NHP Server %s",
			addr, rs.serverAddr)
		tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}
		rs.httpServer.TLSConfig = tlsCfg
		return rs.httpServer.ListenAndServeTLS(rs.config.TLSCertFile, rs.config.TLSKeyFile)
	}
	log.Info("[Relay] starting HTTP relay on %s → NHP Server %s",
		addr, rs.serverAddr)
	return rs.httpServer.ListenAndServe()
}

// Stop gracefully shuts down the relay service.
func (rs *RelayServer) Stop(ctx context.Context) error {
	rs.running.Store(false)
	close(rs.stopCh)

	// Shut down HTTP server.
	err := rs.httpServer.Shutdown(ctx)

	// Close UDP connection.
	rs.connMutex.Lock()
	if rs.conn != nil {
		rs.conn.Close()
	}
	rs.connMutex.Unlock()

	// Stop device.
	rs.device.Stop()

	rs.wg.Wait()
	return err
}

// ──────────────────────────────────────────────────────────────────────────────
// UDP connection management (mirrors AC pattern)
// ──────────────────────────────────────────────────────────────────────────────

func (rs *RelayServer) getOrCreateConnection() *UdpConn {
	rs.connMutex.Lock()
	defer rs.connMutex.Unlock()

	if rs.conn != nil {
		return rs.conn
	}

	conn := &UdpConn{}
	var err error
	conn.netConn, err = net.DialUDP("udp", nil, rs.serverAddr)
	if err != nil {
		log.Error("[Relay] failed to dial server %s: %v", rs.serverAddr, err)
		return nil
	}

	laddr := conn.netConn.LocalAddr()
	localAddr, err := net.ResolveUDPAddr(laddr.Network(), laddr.String())
	if err != nil {
		log.Error("[Relay] resolve local addr error: %v", err)
		conn.netConn.Close()
		return nil
	}

	log.Info("[Relay] new UDP connection %s → %s", localAddr, rs.serverAddr)

	conn.ConnData = &core.ConnectionData{
		Device:               rs.device,
		CookieStore:          &core.CookieStore{},
		RemoteTransactionMap: make(map[uint64]*core.RemoteTransaction),
		LocalAddr:            localAddr,
		RemoteAddr:           rs.serverAddr,
		TimeoutMs:            DefaultConnectionTimeoutMs,
		SendQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
		RecvQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
		BlockSignal:          make(chan struct{}),
		SetTimeoutSignal:     make(chan struct{}),
		StopSignal:           make(chan struct{}),
	}

	// Start recv and connection routines.
	conn.ConnData.Add(1)
	go rs.recvPacketRoutine(conn)

	rs.wg.Add(1)
	go rs.connectionRoutine(conn)

	rs.conn = conn
	return conn
}

// ──────────────────────────────────────────────────────────────────────────────
// Packet send/recv routines (mirrors AC pattern)
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

func (rs *RelayServer) recvPacketRoutine(conn *UdpConn) {
	addrStr := conn.ConnData.RemoteAddr.String()
	defer conn.ConnData.Done()
	defer log.Info("[Relay] recvPacketRoutine for %s stopped", addrStr)

	log.Info("[Relay] recvPacketRoutine for %s started", addrStr)

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

func (rs *RelayServer) connectionRoutine(conn *UdpConn) {
	addrStr := conn.ConnData.RemoteAddr.String()
	defer rs.wg.Done()
	defer log.Info("[Relay] connectionRoutine for %s stopped", addrStr)

	log.Info("[Relay] connectionRoutine for %s started", addrStr)

	defer func() {
		rs.connMutex.Lock()
		if rs.conn == conn {
			rs.conn = nil
		}
		rs.connMutex.Unlock()
		conn.Close()
	}()

	for {
		select {
		case <-rs.stopCh:
			return

		case <-time.After(time.Duration(conn.ConnData.TimeoutMs) * time.Millisecond):
			log.Info("[Relay] connection idle timeout")
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
				log.Info("[Relay] recv [NHP_KPL] from %s", addrStr)
				continue
			}

			// Check if this is a response (ACK/COK) for a pending relay
			// request.  The counter in the response header matches the
			// counter of the original inner packet sent by the agent.
			if pkt.HeaderType == core.NHP_ACK || pkt.HeaderType == core.NHP_COK {
				counter := pkt.Counter()
				rs.pendingMu.Lock()
				ch, found := rs.pendingRequests[counter]
				if found {
					delete(rs.pendingRequests, counter)
				}
				rs.pendingMu.Unlock()

				if found {
					// Copy raw bytes before releasing the pool packet.
					raw := make([]byte, len(pkt.Content))
					copy(raw, pkt.Content)
					rs.device.ReleasePoolPacket(pkt)

					log.Info("[Relay] matched pending request counter=%d, forwarding %d raw bytes",
						counter, len(raw))
					ch <- raw
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
			log.Warning("[Relay] connection blocked %s", addrStr)
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

			conn := rs.getOrCreateConnection()
			if conn == nil {
				log.Error("[Relay] no connection to server, dropping message")
				continue
			}

			md.ConnData = conn.ConnData
			rs.device.SendMsgToPacket(md)
		}
	}
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
// Keepalive routine (mirrors AC pattern)
// ──────────────────────────────────────────────────────────────────────────────

func (rs *RelayServer) keepaliveRoutine() {
	defer rs.wg.Done()
	defer log.Info("[Relay] keepaliveRoutine stopped")

	log.Info("[Relay] keepaliveRoutine started")

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
			rs.connMutex.Lock()
			conn := rs.conn
			rs.connMutex.Unlock()

			if conn == nil {
				// No connection yet; getOrCreateConnection will be called on first request.
				continue
			}

			lastSend := atomic.LoadInt64(&conn.ConnData.LastLocalSendTime)
			if (time.Now().UnixNano() - lastSend) > int64(time.Duration(interval)*time.Second) {
				md := &core.MsgData{
					RemoteAddr:    rs.serverAddr,
					HeaderType:    core.NHP_KPL,
					CipherScheme:  rs.config.CipherScheme,
					TransactionId: rs.device.NextCounterIndex(),
				}
				rs.sendMsgCh <- md
				log.Info("[Relay] sent NHP_KPL keepalive to %s", rs.serverAddr)
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

// handleRelay is the main relay endpoint.
//
// Expected request:
//
//	POST /relay
//	Content-Type: application/octet-stream
//	Body: raw inner NHP packet bytes (KNK / RKN / etc., encrypted by agent)
//
// Response:
//
//	200 OK  — body contains raw NHP ACK / COK packet bytes (encrypted to agent)
//	400 Bad Request  — empty or over-size body
//	504 Gateway Timeout  — NHP Server did not respond in time
//	502 Bad Gateway  — internal error
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

func (rs *RelayServer) handleRelay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read inner NHP packet from request body.
	innerBuf := make([]byte, maxPacketSize+1)
	n, _ := r.Body.Read(innerBuf)
	if n == 0 {
		http.Error(w, "empty packet", http.StatusBadRequest)
		return
	}
	if n > maxPacketSize {
		http.Error(w, "packet too large", http.StatusBadRequest)
		return
	}
	innerPacket := innerBuf[:n]

	// Extract the counter from the inner packet header (bytes [16:24], big-endian uint64).
	// The NHP server echoes this counter in its ACK/COK response, so we use it
	// to match the response back to this HTTP request.
	if n < 24 {
		http.Error(w, "inner packet too short", http.StatusBadRequest)
		return
	}
	innerCounter := binary.BigEndian.Uint64(innerPacket[16:24])

	realAddr := realClientAddr(r)
	log.Info("[Relay] forwarding %d-byte inner packet (counter=%d) from client %s to server %s",
		n, innerCounter, realAddr, rs.serverAddr)

	// Register a pending request keyed on the inner packet counter.
	// The connectionRoutine will fulfil this channel when a matching
	// ACK/COK arrives from the server.
	responseCh := make(chan []byte, 1)
	rs.pendingMu.Lock()
	rs.pendingRequests[innerCounter] = responseCh
	rs.pendingMu.Unlock()

	// Ensure cleanup on timeout / early return.
	defer func() {
		rs.pendingMu.Lock()
		delete(rs.pendingRequests, innerCounter)
		rs.pendingMu.Unlock()
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

	// Send the NHP_RLY envelope to the server.
	trxId := rs.device.NextCounterIndex()
	md := &core.MsgData{
		RemoteAddr:    rs.serverAddr,
		HeaderType:    core.NHP_RLY,
		CipherScheme:  rs.config.CipherScheme,
		TransactionId: trxId,
		Message:       msgBytes,
		PeerPk:        rs.serverPubKey,
	}

	rs.sendMsgCh <- md

	// Wait for the raw encrypted ACK/COK packet from the server.
	udpTimeout := rs.config.UDPTimeoutMs
	if udpTimeout <= 0 {
		udpTimeout = defaultUDPTimeoutMs
	}

	select {
	case rawBytes := <-responseCh:
		log.Info("[Relay] received response for inner counter=%d, %d raw bytes, forwarding to client %s",
			innerCounter, len(rawBytes), realAddr)

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(rawBytes)

	case <-time.After(time.Duration(udpTimeout) * time.Millisecond):
		log.Warning("[Relay] timeout waiting for server response (inner counter=%d, client %s)", innerCounter, realAddr)
		http.Error(w, "NHP Server timeout", http.StatusGatewayTimeout)
	}
}

// realClientAddr returns the originating address of an HTTP request as a
// *net.UDPAddr so it can be encoded in the RelayForwardMsg.
func realClientAddr(r *http.Request) *net.UDPAddr {
	ipStr := ""
	port := 0

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				ipStr = xff[:i]
				break
			}
		}
		if ipStr == "" {
			ipStr = xff
		}
	} else {
		host, portStr, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ipStr = host
			_, _ = fmt.Sscanf(portStr, "%d", &port)
		} else {
			ipStr = r.RemoteAddr
		}
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		ip = net.IPv4zero
	}
	return &net.UDPAddr{IP: ip, Port: port}
}
