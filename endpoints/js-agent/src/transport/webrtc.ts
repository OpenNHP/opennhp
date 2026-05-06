/**
 * WebRTC Transport for Browser Environments
 * Provides connection management and message handling for NHP over WebRTC DataChannel
 *
 * The Go NHP server supports WebRTC DataChannel for browser clients.
 * This transport handles the WebRTC signaling and data channel setup.
 * 
 * Supports two signaling modes:
 * 1. HTTP POST signaling (legacy) - single request/response for offer/answer
 * 2. WebSocket signaling (preferred) - supports trickle ICE for faster connection
 */

import type {
  ConnectionState,
  TransportMessage,
  TransportEvent,
  EventHandler,
} from '../types.js';

/** WebRTC transport configuration */
export interface WebRTCTransportConfig {
  /** STUN server URLs for ICE connectivity */
  stunServers?: string[];
  /** TURN server URLs for ICE connectivity (with credentials) */
  turnServers?: RTCIceServer[];
  /** HTTP signaling server URL for offer/answer exchange (legacy) */
  signalingUrl?: string;
  /** WebSocket signaling server URL (preferred, supports trickle ICE) */
  wsSignalingUrl?: string;
  /** Pre-configured remote SDP answer (for file-based signaling) */
  remoteAnswer?: RTCSessionDescriptionInit;
  /** Connection timeout in milliseconds (default: 30000) */
  timeout?: number;
  /** Data channel label (default: 'nhp') */
  channelLabel?: string;
}

/** Signaling message format for WebSocket communication */
interface SignalingMessage {
  type: 'offer' | 'answer' | 'candidate' | 'error';
  sdp?: RTCSessionDescriptionInit;
  candidate?: RTCIceCandidateInit;
  error?: string;
}

/** Default configuration values */
const DEFAULT_CONFIG = {
  stunServers: ['stun:stun.l.google.com:19302'],
  timeout: 30000,
  channelLabel: 'nhp',
};

/** WebRTC-specific transport events (extends base TransportEvent with 'offer') */
export type WebRTCTransportEvent = TransportEvent | 'offer';

/**
 * WebRTC transport for NHP communication in browser environments
 */
export class WebRTCTransport {
  private pc: RTCPeerConnection | null = null;
  private dc: RTCDataChannel | null = null;
  private ws: WebSocket | null = null;
  private config: WebRTCTransportConfig & typeof DEFAULT_CONFIG;
  private state: ConnectionState = 'disconnected';
  private eventHandlers: Map<WebRTCTransportEvent, Set<EventHandler>> = new Map();
  private pendingCandidates: RTCIceCandidateInit[] = [];

  constructor(config: WebRTCTransportConfig = {}) {
    this.config = {
      ...DEFAULT_CONFIG,
      ...config,
    };
  }

  /**
   * Get current connection state
   */
  getState(): ConnectionState {
    return this.state;
  }

  /**
   * Get the local SDP offer for signaling
   * Call this to get the offer to send to the server
   */
  async createOffer(): Promise<RTCSessionDescriptionInit> {
    if (!this.pc) {
      await this.initPeerConnection();
    }

    // Create data channel before creating offer
    this.dc = this.pc!.createDataChannel(this.config.channelLabel, {
      ordered: true,
    });
    this.setupDataChannel();

    const offer = await this.pc!.createOffer();
    await this.pc!.setLocalDescription(offer);

    // Wait for ICE gathering to complete (for non-trickle ICE)
    await this.waitForIceGathering();

    const localDesc = this.pc!.localDescription;
    if (!localDesc) {
      throw new Error('Failed to create local description');
    }

    this.emit('offer', localDesc);
    return localDesc;
  }

  /**
   * Set the remote SDP answer from the server
   * Call this after receiving the answer from signaling
   */
  async setAnswer(answer: RTCSessionDescriptionInit): Promise<void> {
    if (!this.pc) {
      throw new Error('Peer connection not initialized. Call createOffer() first.');
    }

    await this.pc.setRemoteDescription(answer);

    // Add any pending ICE candidates
    for (const candidate of this.pendingCandidates) {
      await this.pc.addIceCandidate(candidate);
    }
    this.pendingCandidates = [];
  }

  /**
   * Add an ICE candidate from the remote peer
   */
  async addIceCandidate(candidate: RTCIceCandidateInit): Promise<void> {
    if (!this.pc) {
      // Queue candidate for later
      this.pendingCandidates.push(candidate);
      return;
    }

    if (this.pc.remoteDescription) {
      await this.pc.addIceCandidate(candidate);
    } else {
      // Queue candidate until remote description is set
      this.pendingCandidates.push(candidate);
    }
  }

  /**
   * Connect using a signaling server
   * This handles the full offer/answer exchange automatically
   */
  async connect(): Promise<void> {
    if (this.state === 'connected' || this.state === 'connecting') {
      return;
    }

    this.state = 'connecting';

    try {
      // If we have a pre-configured answer, use it
      if (this.config.remoteAnswer) {
        const offer = await this.createOffer();
        await this.setAnswer(this.config.remoteAnswer);
      }
      // Prefer WebSocket signaling if available
      else if (this.config.wsSignalingUrl) {
        await this.connectViaWebSocket();
      }
      // Fall back to HTTP signaling
      else if (this.config.signalingUrl) {
        const offer = await this.createOffer();
        const answer = await this.exchangeViaSignaling(offer);
        await this.setAnswer(answer);
      } else {
        // Manual signaling - emit offer and wait for setAnswer to be called
        const offer = await this.createOffer();
        this.emit('offer', offer);
        // The caller must call setAnswer() with the server's response
      }

      // Wait for data channel to open
      await this.waitForDataChannel();

      this.state = 'connected';
      this.emit('open', undefined);
    } catch (err) {
      this.state = 'disconnected';
      this.emit('error', err);
      throw err;
    }
  }

  /**
   * Connect using WebSocket signaling with trickle ICE support
   */
  private async connectViaWebSocket(): Promise<void> {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('WebSocket signaling timeout'));
      }, this.config.timeout);

      const wsUrl = this.config.wsSignalingUrl!;
      console.log(`[WebRTC] Connecting to signaling server: ${wsUrl}`);
      
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = async () => {
        console.log('[WebRTC] WebSocket connected, initializing peer connection');
        try {
          // Initialize peer connection with trickle ICE support
          await this.initPeerConnectionWithTrickleICE();

          // Create data channel
          this.dc = this.pc!.createDataChannel(this.config.channelLabel, {
            ordered: true,
          });
          this.setupDataChannel();

          // Create and send offer
          const offer = await this.pc!.createOffer();
          await this.pc!.setLocalDescription(offer);

          console.log('[WebRTC] Sending SDP offer');
          this.sendSignalingMessage({
            type: 'offer',
            sdp: offer,
          });
        } catch (err) {
          clearTimeout(timeout);
          reject(err);
        }
      };

      this.ws.onmessage = async (event) => {
        try {
          const msg: SignalingMessage = JSON.parse(event.data);
          console.log(`[WebRTC] Received signaling message: ${msg.type}`);

          switch (msg.type) {
            case 'answer':
              if (msg.sdp) {
                await this.setAnswer(msg.sdp);
                console.log('[WebRTC] Remote description set');
                // Don't resolve yet - wait for data channel to open
              }
              break;

            case 'candidate':
              if (msg.candidate) {
                await this.addIceCandidate(msg.candidate);
                console.log('[WebRTC] Added ICE candidate');
              }
              break;

            case 'error':
              console.error('[WebRTC] Signaling error:', msg.error);
              clearTimeout(timeout);
              reject(new Error(msg.error || 'Signaling error'));
              break;
          }
        } catch (err) {
          console.error('[WebRTC] Error processing signaling message:', err);
        }
      };

      this.ws.onerror = (event) => {
        console.error('[WebRTC] WebSocket error:', event);
        clearTimeout(timeout);
        reject(new Error('WebSocket signaling error'));
      };

      this.ws.onclose = () => {
        console.log('[WebRTC] WebSocket closed');
        // Only reject if we haven't connected yet
        if (this.state === 'connecting') {
          clearTimeout(timeout);
          reject(new Error('WebSocket closed before connection established'));
        }
      };

      // Resolve when data channel opens
      const checkDataChannel = setInterval(() => {
        if (this.dc?.readyState === 'open') {
          clearInterval(checkDataChannel);
          clearTimeout(timeout);
          console.log('[WebRTC] Data channel open, connection complete');
          resolve();
        }
      }, 100);
    });
  }

  /**
   * Initialize peer connection with trickle ICE (candidates sent as they're discovered)
   */
  private async initPeerConnectionWithTrickleICE(): Promise<void> {
    const iceServers: RTCIceServer[] = [];

    // Add STUN servers
    if (this.config.stunServers) {
      for (const url of this.config.stunServers) {
        iceServers.push({ urls: url });
      }
    }

    // Add TURN servers
    if (this.config.turnServers) {
      iceServers.push(...this.config.turnServers);
    }

    this.pc = new RTCPeerConnection({
      iceServers,
    });

    // Send ICE candidates to the signaling server as they're discovered (trickle ICE)
    this.pc.onicecandidate = (event) => {
      if (event.candidate && this.ws?.readyState === WebSocket.OPEN) {
        console.log('[WebRTC] Sending ICE candidate');
        this.sendSignalingMessage({
          type: 'candidate',
          candidate: event.candidate.toJSON(),
        });
      }
    };

    this.pc.oniceconnectionstatechange = () => {
      console.log(`[WebRTC] ICE connection state: ${this.pc?.iceConnectionState}`);
      if (this.pc?.iceConnectionState === 'failed') {
        this.emit('error', new Error('ICE connection failed'));
      } else if (this.pc?.iceConnectionState === 'disconnected') {
        this.state = 'disconnected';
        this.emit('close', undefined);
      }
    };

    this.pc.onconnectionstatechange = () => {
      console.log(`[WebRTC] Connection state: ${this.pc?.connectionState}`);
      if (this.pc?.connectionState === 'failed') {
        this.emit('error', new Error('Peer connection failed'));
        this.state = 'disconnected';
      }
    };
  }

  /**
   * Send a signaling message via WebSocket
   */
  private sendSignalingMessage(msg: SignalingMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  /**
   * Connect with manual signaling
   * Returns the offer - caller must exchange with server and call setAnswer()
   */
  async connectManual(): Promise<RTCSessionDescriptionInit> {
    if (this.state === 'connected' || this.state === 'connecting') {
      throw new Error('Already connected or connecting');
    }

    this.state = 'connecting';

    try {
      const offer = await this.createOffer();
      return offer;
    } catch (err) {
      this.state = 'disconnected';
      throw err;
    }
  }

  /**
   * Complete manual connection after setting the answer
   */
  async completeConnection(): Promise<void> {
    await this.waitForDataChannel();
    this.state = 'connected';
    this.emit('open', undefined);
  }

  /**
   * Disconnect from the WebRTC peer
   */
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    if (this.dc) {
      this.dc.close();
      this.dc = null;
    }
    if (this.pc) {
      this.pc.close();
      this.pc = null;
    }
    this.pendingCandidates = [];
    this.state = 'disconnected';
    this.emit('close', undefined);
  }

  /**
   * Send data over the WebRTC data channel
   */
  send(data: Uint8Array): void {
    if (!this.dc || this.dc.readyState !== 'open') {
      throw new Error('WebRTC data channel is not open');
    }

    // Create a clean ArrayBuffer copy to avoid SharedArrayBuffer issues
    const buffer = new ArrayBuffer(data.length);
    new Uint8Array(buffer).set(data);
    this.dc.send(buffer);
  }

  /**
   * Send data and wait for a response with timeout
   */
  async sendAndReceive(data: Uint8Array, timeout?: number): Promise<TransportMessage> {
    const timeoutMs = timeout ?? this.config.timeout;

    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        this.off('message', messageHandler as EventHandler);
        reject(new Error('WebRTC request timeout'));
      }, timeoutMs);

      const messageHandler = (msg: unknown) => {
        clearTimeout(timer);
        this.off('message', messageHandler as EventHandler);
        resolve(msg as TransportMessage);
      };

      this.on('message', messageHandler as EventHandler);
      this.send(data);
    });
  }

  /**
   * Register an event handler
   */
  on(event: WebRTCTransportEvent, handler: EventHandler): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    this.eventHandlers.get(event)!.add(handler);
  }

  /**
   * Remove an event handler
   */
  off(event: WebRTCTransportEvent, handler: EventHandler): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.delete(handler);
    }
  }

  /**
   * Check if the transport is connected
   */
  isConnected(): boolean {
    return this.state === 'connected' && this.dc?.readyState === 'open';
  }

  /**
   * Get the local SDP for debugging/manual signaling
   */
  getLocalDescription(): RTCSessionDescription | null {
    return this.pc?.localDescription ?? null;
  }

  private async initPeerConnection(): Promise<void> {
    const iceServers: RTCIceServer[] = [];

    // Add STUN servers
    if (this.config.stunServers) {
      for (const url of this.config.stunServers) {
        iceServers.push({ urls: url });
      }
    }

    // Add TURN servers
    if (this.config.turnServers) {
      iceServers.push(...this.config.turnServers);
    }

    this.pc = new RTCPeerConnection({
      iceServers,
    });

    this.pc.oniceconnectionstatechange = () => {
      if (this.pc?.iceConnectionState === 'failed') {
        this.emit('error', new Error('ICE connection failed'));
      } else if (this.pc?.iceConnectionState === 'disconnected') {
        this.state = 'disconnected';
        this.emit('close', undefined);
      }
    };

    this.pc.onconnectionstatechange = () => {
      if (this.pc?.connectionState === 'failed') {
        this.emit('error', new Error('Peer connection failed'));
        this.state = 'disconnected';
      }
    };
  }

  private setupDataChannel(): void {
    if (!this.dc) return;

    this.dc.binaryType = 'arraybuffer';

    this.dc.onopen = () => {
      // Data channel opened - connection complete
    };

    this.dc.onclose = () => {
      this.state = 'disconnected';
      this.emit('close', undefined);
    };

    this.dc.onerror = (event) => {
      this.emit('error', event);
    };

    this.dc.onmessage = (event) => {
      const data = new Uint8Array(event.data as ArrayBuffer);
      const message: TransportMessage = { data };
      this.emit('message', message);
    };
  }

  private async waitForIceGathering(): Promise<void> {
    if (this.pc?.iceGatheringState === 'complete') {
      return;
    }

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('ICE gathering timeout'));
      }, this.config.timeout);

      const checkState = () => {
        if (this.pc?.iceGatheringState === 'complete') {
          clearTimeout(timeout);
          this.pc.removeEventListener('icegatheringstatechange', checkState);
          resolve();
        }
      };

      this.pc?.addEventListener('icegatheringstatechange', checkState);

      // Also resolve on null candidate (gathering complete signal)
      const onCandidate = (event: RTCPeerConnectionIceEvent) => {
        if (event.candidate === null) {
          clearTimeout(timeout);
          this.pc?.removeEventListener('icecandidate', onCandidate);
          this.pc?.removeEventListener('icegatheringstatechange', checkState);
          resolve();
        }
      };
      this.pc?.addEventListener('icecandidate', onCandidate);
    });
  }

  private async waitForDataChannel(): Promise<void> {
    if (this.dc?.readyState === 'open') {
      return;
    }

    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error('Data channel open timeout'));
      }, this.config.timeout);

      if (!this.dc) {
        clearTimeout(timeout);
        reject(new Error('Data channel not created'));
        return;
      }

      const onOpen = () => {
        clearTimeout(timeout);
        this.dc?.removeEventListener('open', onOpen);
        resolve();
      };

      const onError = (event: Event) => {
        clearTimeout(timeout);
        this.dc?.removeEventListener('error', onError);
        reject(event);
      };

      this.dc.addEventListener('open', onOpen);
      this.dc.addEventListener('error', onError);
    });
  }

  private async exchangeViaSignaling(offer: RTCSessionDescriptionInit): Promise<RTCSessionDescriptionInit> {
    const response = await fetch(this.config.signalingUrl!, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(offer),
    });

    if (!response.ok) {
      throw new Error(`Signaling request failed: ${response.status}`);
    }

    return await response.json();
  }

  private emit(event: WebRTCTransportEvent, data: unknown): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.forEach((handler) => handler(data));
    }
  }
}
