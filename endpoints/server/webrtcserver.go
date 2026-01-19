package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// WebRTCConfig holds settings for the optional WebRTC transport.
type WebRTCConfig struct {
	Enable      bool     `toml:"enable" json:"enable"`
	OfferFile   string   `toml:"offerFile" json:"offerFile"`
	AnswerFile  string   `toml:"answerFile" json:"answerFile"`
	StunServers []string `toml:"stunServers" json:"stunServers"`
	TurnServers []string `toml:"turnServers" json:"turnServers"`
	// WebSocket signaling configuration
	EnableWebSocket  bool   `toml:"enableWebSocket" json:"enableWebSocket"`
	WebSocketPath    string `toml:"webSocketPath" json:"webSocketPath"`       // default: "/webrtc/signaling"
	MaxConnections   int    `toml:"maxConnections" json:"maxConnections"`     // maximum concurrent WebSocket connections
	PingIntervalSecs int    `toml:"pingIntervalSecs" json:"pingIntervalSecs"` // WebSocket ping interval in seconds
}

// SignalingMessageType defines the type of signaling message
type SignalingMessageType string

const (
	SignalingTypeOffer     SignalingMessageType = "offer"
	SignalingTypeAnswer    SignalingMessageType = "answer"
	SignalingTypeCandidate SignalingMessageType = "candidate"
	SignalingTypePing      SignalingMessageType = "ping"
	SignalingTypePong      SignalingMessageType = "pong"
	SignalingTypeError     SignalingMessageType = "error"
	SignalingTypeReady     SignalingMessageType = "ready"
)

// SignalingMessage represents a WebSocket signaling message
type SignalingMessage struct {
	Type      SignalingMessageType       `json:"type"`
	SessionID string                     `json:"sessionId,omitempty"`
	SDP       *webrtc.SessionDescription `json:"sdp,omitempty"`
	Candidate *webrtc.ICECandidateInit   `json:"candidate,omitempty"`
	Error     string                     `json:"error,omitempty"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	id        string
	conn      *websocket.Conn
	pc        *webrtc.PeerConnection
	dc        *webrtc.DataChannel
	sendCh    chan []byte
	closeCh   chan struct{}
	closeOnce sync.Once
}

// WebRTCServer bridges a WebRTC DataChannel with the UDP server message pipeline.
type WebRTCServer struct {
	us          *UdpServer
	conf        *WebRTCConfig
	pc          *webrtc.PeerConnection
	upgrader    websocket.Upgrader
	clientsMu   sync.RWMutex
	clients     map[string]*WebSocketClient
	clientCount atomic.Int32
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

func NewWebRTCServer(us *UdpServer, conf *WebRTCConfig) *WebRTCServer {
	// Set default values
	if conf.WebSocketPath == "" {
		conf.WebSocketPath = "/webrtc/signaling"
	}
	if conf.MaxConnections <= 0 {
		conf.MaxConnections = 100
	}
	if conf.PingIntervalSecs <= 0 {
		conf.PingIntervalSecs = 30
	}

	return &WebRTCServer{
		us:      us,
		conf:    conf,
		clients: make(map[string]*WebSocketClient),
		stopCh:  make(chan struct{}),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now; can be restricted in production
			},
		},
	}
}

func (w *WebRTCServer) Start() error {
	if w.conf == nil || !w.conf.Enable {
		return nil
	}

	cfg := webrtc.Configuration{}
	for _, u := range w.conf.StunServers {
		cfg.ICEServers = append(cfg.ICEServers, webrtc.ICEServer{URLs: []string{u}})
	}
	for _, u := range w.conf.TurnServers {
		cfg.ICEServers = append(cfg.ICEServers, webrtc.ICEServer{URLs: []string{u}})
	}

	var err error
	w.pc, err = webrtc.NewPeerConnection(cfg)
	if err != nil {
		return err
	}

	w.pc.OnDataChannel(w.setupDataChannel)

	w.pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateConnected ||
			state == webrtc.ICEConnectionStateCompleted {
			updatePeerRemoteAddr(w.pc)
		}
	})

	// if an offer file is provided, perform one-shot signaling using files
	if w.conf.OfferFile != "" {
		if offerBytes, err := os.ReadFile(w.conf.OfferFile); err == nil {
			var offer webrtc.SessionDescription
			if err := json.Unmarshal(offerBytes, &offer); err == nil {
				if err := w.pc.SetRemoteDescription(offer); err == nil {
					answer, err := w.pc.CreateAnswer(nil)
					if err == nil {
						if err = w.pc.SetLocalDescription(answer); err == nil {
							if w.conf.AnswerFile != "" {
								if data, err := json.Marshal(answer); err == nil {
									_ = os.WriteFile(w.conf.AnswerFile, data, 0644) //nolint:gosec // G306: SDP signaling data is not sensitive
								}
							}
						}
					}
				}
			}
		}
	}

	if w.conf.EnableWebSocket {
		log.Info("WebRTC WebSocket signaling enabled on path: %s", w.conf.WebSocketPath)
	}

	return nil
}

// GetWebSocketPath returns the WebSocket signaling path
func (w *WebRTCServer) GetWebSocketPath() string {
	if w.conf == nil {
		return ""
	}
	return w.conf.WebSocketPath
}

// IsWebSocketEnabled returns whether WebSocket signaling is enabled
func (w *WebRTCServer) IsWebSocketEnabled() bool {
	return w.conf != nil && w.conf.Enable && w.conf.EnableWebSocket
}

// HandleWebSocket handles incoming WebSocket connections for WebRTC signaling
func (w *WebRTCServer) HandleWebSocket(rw http.ResponseWriter, r *http.Request) {
	log.Debug("HandleWebSocket request received")
	if !w.IsWebSocketEnabled() {
		http.Error(rw, "WebSocket signaling not enabled", http.StatusServiceUnavailable)
		return
	}

	// Check connection limit
	if int(w.clientCount.Load()) >= w.conf.MaxConnections {
		http.Error(rw, "Maximum WebSocket connections reached", http.StatusServiceUnavailable)
		return
	}

	// Upgrade HTTP connection to WebSocket
	wsConn, err := w.upgrader.Upgrade(rw, r, nil)
	if err != nil {
		log.Error("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	log.Debug("HandleWebSocket update ok")

	// Generate client ID
	clientID := generateClientID()
	client := &WebSocketClient{
		id:      clientID,
		conn:    wsConn,
		sendCh:  make(chan []byte, 256),
		closeCh: make(chan struct{}),
	}

	w.clientsMu.Lock()
	w.clients[clientID] = client
	w.clientsMu.Unlock()
	w.clientCount.Add(1)

	log.Info("WebRTC WebSocket client connected: %s from %s", clientID, r.RemoteAddr)

	// Start client goroutines
	w.wg.Add(2)
	go w.clientReadPump(client)
	go w.clientWritePump(client)

	// Send ready message to client
	readyMsg := SignalingMessage{
		Type:      SignalingTypeReady,
		SessionID: clientID,
	}
	w.sendToClient(client, readyMsg)
}

// clientReadPump reads messages from the WebSocket connection
func (w *WebRTCServer) clientReadPump(client *WebSocketClient) {
	defer w.wg.Done()
	defer w.removeClient(client)

	client.conn.SetReadLimit(64 * 1024) // 64KB max message size
	_ = client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		_ = client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-w.stopCh:
			return
		case <-client.closeCh:
			return
		default:
		}

		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warning("WebSocket read error for client %s: %v", client.id, err)
			}
			return
		}

		var msg SignalingMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Warning("Failed to parse signaling message from client %s: %v", client.id, err)
			w.sendError(client, "Invalid message format")
			continue
		}

		w.handleSignalingMessage(client, &msg)
	}
}

// clientWritePump writes messages to the WebSocket connection
func (w *WebRTCServer) clientWritePump(client *WebSocketClient) {
	defer w.wg.Done()
	defer client.conn.Close()

	pingInterval := time.Duration(w.conf.PingIntervalSecs) * time.Second
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-client.closeCh:
			return
		case message, ok := <-client.sendCh:
			if !ok {
				_ = client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			_ = client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Warning("WebSocket write error for client %s: %v", client.id, err)
				return
			}

		case <-ticker.C:
			_ = client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSignalingMessage processes incoming signaling messages
func (w *WebRTCServer) handleSignalingMessage(client *WebSocketClient, msg *SignalingMessage) {
	log.Debug("handleSignalingMessage")
	switch msg.Type {
	case SignalingTypeOffer:
		w.handleOffer(client, msg)
	case SignalingTypeCandidate:
		w.handleICECandidate(client, msg)
	case SignalingTypePing:
		w.sendToClient(client, SignalingMessage{Type: SignalingTypePong, SessionID: client.id})
	default:
		log.Warning("Unknown signaling message type from client %s: %s", client.id, msg.Type)
	}
}

// handleOffer processes an SDP offer from a client
func (w *WebRTCServer) handleOffer(client *WebSocketClient, msg *SignalingMessage) {
	log.Debug("handleOffer")
	if msg.SDP == nil {
		w.sendError(client, "Offer message missing SDP")
		return
	}

	// Create WebRTC configuration
	cfg := webrtc.Configuration{}
	for _, u := range w.conf.StunServers {
		cfg.ICEServers = append(cfg.ICEServers, webrtc.ICEServer{URLs: []string{u}})
	}
	for _, u := range w.conf.TurnServers {
		cfg.ICEServers = append(cfg.ICEServers, webrtc.ICEServer{URLs: []string{u}})
	}

	// Create new peer connection for this client
	pc, err := webrtc.NewPeerConnection(cfg)
	if err != nil {
		log.Error("Failed to create peer connection for client %s: %v", client.id, err)
		w.sendError(client, "Failed to create peer connection")
		return
	}

	client.pc = pc

	// Set up ICE candidate handling
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		candidateInit := candidate.ToJSON()
		candidateMsg := SignalingMessage{
			Type:      SignalingTypeCandidate,
			SessionID: client.id,
			Candidate: &candidateInit,
		}
		w.sendToClient(client, candidateMsg)
	})

	// Set up data channel handling
	pc.OnDataChannel(func(dc *webrtc.DataChannel) {
		client.dc = dc
		w.setupClientDataChannel(client, dc)
	})

	// Set up connection state handling
	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Info("WebRTC connection state changed for client %s: %s", client.id, state.String())
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			w.removeClient(client)
		}
	})

	// Set remote description (the offer)
	if err := pc.SetRemoteDescription(*msg.SDP); err != nil {
		log.Error("Failed to set remote description for client %s: %v", client.id, err)
		w.sendError(client, "Failed to set remote description")
		return
	}

	// Create answer
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		log.Error("Failed to create answer for client %s: %v", client.id, err)
		w.sendError(client, "Failed to create answer")
		return
	}

	// Set local description
	if err := pc.SetLocalDescription(answer); err != nil {
		log.Error("Failed to set local description for client %s: %v", client.id, err)
		w.sendError(client, "Failed to set local description")
		return
	}

	// Send answer back to client
	answerMsg := SignalingMessage{
		Type:      SignalingTypeAnswer,
		SessionID: client.id,
		SDP:       pc.LocalDescription(),
	}
	w.sendToClient(client, answerMsg)

	log.Info("WebRTC signaling completed for client %s", client.id)
}

// handleICECandidate processes an ICE candidate from a client
func (w *WebRTCServer) handleICECandidate(client *WebSocketClient, msg *SignalingMessage) {
	log.Debug("handleICECandidate")
	if client.pc == nil {
		w.sendError(client, "Peer connection not established")
		return
	}

	if msg.Candidate == nil {
		return // Empty candidate signals end of candidates
	}

	if err := client.pc.AddICECandidate(*msg.Candidate); err != nil {
		log.Warning("Failed to add ICE candidate for client %s: %v", client.id, err)
	}
}

// setupClientDataChannel configures a data channel for a specific client
func (w *WebRTCServer) setupClientDataChannel(client *WebSocketClient, dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		log.Info("WebRTC data channel %d open for client %s", dc.ID(), client.id)
		recvTime := time.Now().UnixNano()
		addr := &net.UDPAddr{IP: net.IPv4zero, Port: int(*dc.ID())}
		conn := &UdpConn{isWebRTC: true, dc: dc}
		conn.ConnData = &core.ConnectionData{
			InitTime:             recvTime,
			LastLocalRecvTime:    recvTime,
			Device:               w.us.device,
			LocalAddr:            w.us.listenAddr,
			RemoteAddr:           addr,
			CookieStore:          &core.CookieStore{},
			RemoteTransactionMap: make(map[uint64]*core.RemoteTransaction),
			TimeoutMs:            DefaultAgentConnectionTimeoutMs,
			SendQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
			RecvQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
			BlockSignal:          make(chan struct{}),
			SetTimeoutSignal:     make(chan struct{}),
			StopSignal:           make(chan struct{}),
		}

		key := addr.String()
		w.us.remoteConnectionMapMutex.Lock()
		w.us.remoteConnectionMap[key] = conn
		w.us.remoteConnectionMapMutex.Unlock()

		w.us.wg.Add(1)
		go w.us.connectionRoutine(conn)
	})

	dc.OnMessage(func(m webrtc.DataChannelMessage) {
		if m.IsString {
			return
		}
		addr := &net.UDPAddr{IP: net.IPv4zero, Port: int(*dc.ID())}
		key := addr.String()
		w.us.remoteConnectionMapMutex.Lock()
		conn, ok := w.us.remoteConnectionMap[key]
		w.us.remoteConnectionMapMutex.Unlock()
		if !ok {
			return
		}
		pkt := w.us.device.AllocatePoolPacket()
		copy(pkt.Buf[:], m.Data)
		pkt.Content = pkt.Buf[:len(m.Data)]
		if len(pkt.Content) < pkt.MinimalLength() {
			w.us.device.ReleasePoolPacket(pkt)
			return
		}
		_, _, err := w.us.device.RecvPrecheck(pkt) // this check also records packet header type
		if err != nil {
			// discard if precheck failed
			return
		}
		atomic.AddUint64(&w.us.stats.totalRecvBytes, uint64(len(m.Data)))
		conn.ConnData.ForwardInboundPacket(pkt)
	})

	dc.OnClose(func() {
		log.Info("WebRTC data channel %d closed for client %s", dc.ID(), client.id)
		addr := &net.UDPAddr{IP: net.IPv4zero, Port: int(*dc.ID())}
		key := addr.String()
		w.us.remoteConnectionMapMutex.Lock()
		conn, ok := w.us.remoteConnectionMap[key]
		if ok {
			delete(w.us.remoteConnectionMap, key)
		}
		w.us.remoteConnectionMapMutex.Unlock()
		if ok {
			conn.Close()
		}
	})
}

// SendToDataChannel sends raw data directly through a DataChannel
// This is a low-level method for sending arbitrary binary data
func (w *WebRTCServer) SendToDataChannel(dc *webrtc.DataChannel, data []byte) error {
	if dc == nil {
		return fmt.Errorf("data channel is nil")
	}
	if dc.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("data channel is not open, current state: %s", dc.ReadyState().String())
	}
	return dc.Send(data)
}

// SendToDataChannelString sends a string message through a DataChannel
func (w *WebRTCServer) SendToDataChannelString(dc *webrtc.DataChannel, message string) error {
	if dc == nil {
		return fmt.Errorf("data channel is nil")
	}
	if dc.ReadyState() != webrtc.DataChannelStateOpen {
		return fmt.Errorf("data channel is not open, current state: %s", dc.ReadyState().String())
	}
	return dc.SendText(message)
}

// SendToClient sends binary data to a specific WebSocket client's DataChannel
func (w *WebRTCServer) SendToClientDataChannel(clientID string, data []byte) error {
	w.clientsMu.RLock()
	client, ok := w.clients[clientID]
	w.clientsMu.RUnlock()

	if !ok {
		return fmt.Errorf("client %s not found", clientID)
	}

	if client.dc == nil {
		return fmt.Errorf("client %s has no data channel", clientID)
	}

	return w.SendToDataChannel(client.dc, data)
}

// BroadcastToAllDataChannels sends binary data to all connected DataChannels
func (w *WebRTCServer) BroadcastToAllDataChannels(data []byte) {
	w.clientsMu.RLock()
	defer w.clientsMu.RUnlock()

	for clientID, client := range w.clients {
		if client.dc != nil && client.dc.ReadyState() == webrtc.DataChannelStateOpen {
			if err := client.dc.Send(data); err != nil {
				log.Warning("Failed to send data to client %s: %v", clientID, err)
			}
		}
	}
}

// GetClientDataChannel returns the DataChannel for a specific client
func (w *WebRTCServer) GetClientDataChannel(clientID string) *webrtc.DataChannel {
	w.clientsMu.RLock()
	defer w.clientsMu.RUnlock()

	if client, ok := w.clients[clientID]; ok {
		return client.dc
	}
	return nil
}

// GetConnectedClientIDs returns a list of all connected client IDs
func (w *WebRTCServer) GetConnectedClientIDs() []string {
	w.clientsMu.RLock()
	defer w.clientsMu.RUnlock()

	ids := make([]string, 0, len(w.clients))
	for id := range w.clients {
		ids = append(ids, id)
	}
	return ids
}

func (w *WebRTCServer) setupDataChannel(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		log.Info("WebRTC data channel %d open", dc.ID())
		recvTime := time.Now().UnixNano()
		addr := &net.UDPAddr{IP: net.IPv4zero, Port: int(*dc.ID())}
		conn := &UdpConn{isWebRTC: true, dc: dc}
		conn.ConnData = &core.ConnectionData{
			InitTime:             recvTime,
			LastLocalRecvTime:    recvTime,
			Device:               w.us.device,
			LocalAddr:            w.us.listenAddr,
			RemoteAddr:           addr,
			CookieStore:          &core.CookieStore{},
			RemoteTransactionMap: make(map[uint64]*core.RemoteTransaction),
			TimeoutMs:            DefaultAgentConnectionTimeoutMs,
			SendQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
			RecvQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
			BlockSignal:          make(chan struct{}),
			SetTimeoutSignal:     make(chan struct{}),
			StopSignal:           make(chan struct{}),
		}

		key := addr.String()
		w.us.remoteConnectionMapMutex.Lock()
		w.us.remoteConnectionMap[key] = conn
		w.us.remoteConnectionMapMutex.Unlock()

		w.us.wg.Add(1)
		go w.us.connectionRoutine(conn)
	})

	dc.OnMessage(func(m webrtc.DataChannelMessage) {
		if m.IsString {
			return
		}
		addr := &net.UDPAddr{IP: net.IPv4zero, Port: int(*dc.ID())}
		key := addr.String()
		w.us.remoteConnectionMapMutex.Lock()
		conn, ok := w.us.remoteConnectionMap[key]
		w.us.remoteConnectionMapMutex.Unlock()
		if !ok {
			return
		}
		pkt := w.us.device.AllocatePoolPacket()
		copy(pkt.Buf[:], m.Data)
		pkt.Content = pkt.Buf[:len(m.Data)]
		if len(pkt.Content) < pkt.MinimalLength() {
			w.us.device.ReleasePoolPacket(pkt)
			return
		}
		atomic.AddUint64(&w.us.stats.totalRecvBytes, uint64(len(m.Data)))
		conn.ConnData.ForwardInboundPacket(pkt)
	})

	dc.OnClose(func() {
		addr := &net.UDPAddr{IP: net.IPv4zero, Port: int(*dc.ID())}
		key := addr.String()
		w.us.remoteConnectionMapMutex.Lock()
		conn, ok := w.us.remoteConnectionMap[key]
		if ok {
			delete(w.us.remoteConnectionMap, key)
		}
		w.us.remoteConnectionMapMutex.Unlock()
		if ok {
			conn.Close()
		}
	})
}

// sendToClient sends a signaling message to a client
func (w *WebRTCServer) sendToClient(client *WebSocketClient, msg SignalingMessage) {
	log.Debug("sendToClient")
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error("Failed to marshal signaling message: %v", err)
		return
	}

	select {
	case client.sendCh <- data:
	default:
		log.Warning("Send channel full for client %s, dropping message", client.id)
	}
}

// sendError sends an error message to a client
func (w *WebRTCServer) sendError(client *WebSocketClient, errMsg string) {
	w.sendToClient(client, SignalingMessage{
		Type:      SignalingTypeError,
		SessionID: client.id,
		Error:     errMsg,
	})
}

// removeClient removes a client and cleans up resources
func (w *WebRTCServer) removeClient(client *WebSocketClient) {
	client.closeOnce.Do(func() {
		close(client.closeCh)

		w.clientsMu.Lock()
		delete(w.clients, client.id)
		w.clientsMu.Unlock()
		w.clientCount.Add(-1)

		if client.pc != nil {
			_ = client.pc.Close()
		}

		log.Info("WebRTC WebSocket client disconnected: %s", client.id)
	})
}

func (w *WebRTCServer) Stop() {
	close(w.stopCh)

	// Close all client connections
	w.clientsMu.Lock()
	for _, client := range w.clients {
		client.closeOnce.Do(func() {
			close(client.closeCh)
			if client.pc != nil {
				_ = client.pc.Close()
			}
		})
	}
	w.clients = make(map[string]*WebSocketClient)
	w.clientsMu.Unlock()

	w.wg.Wait()

	if w.pc != nil {
		_ = w.pc.Close()
	}

	log.Info("WebRTC server stopped")
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of the specified length
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond) // ensure different values
	}
	return string(b)
}

func updatePeerRemoteAddr(pc *webrtc.PeerConnection) *net.UDPAddr {
	stats := pc.GetStats()

	var selectedPair *webrtc.ICECandidatePairStats
	candidates := map[string]webrtc.ICECandidateStats{}

	for _, s := range stats {
		switch v := s.(type) {
		case webrtc.ICECandidatePairStats:
			if v.State == "succeeded" && v.Nominated {
				selectedPair = &v
			}
		case webrtc.ICECandidateStats:
			candidates[v.ID] = v
		}
	}

	if selectedPair == nil {
		return nil
	}
	remote, found := candidates[selectedPair.RemoteCandidateID]
	if !found {
		return nil
	}
	remoteIp := net.ParseIP(remote.IP)
	if remoteIp == nil {
		return nil
	}

	return &net.UDPAddr{
		IP:   remoteIp,
		Port: int(remote.Port),
	}
}
