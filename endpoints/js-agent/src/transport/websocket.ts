/**
 * WebSocket Transport for Browser Environments
 * Provides connection management and message handling for NHP over WebSocket
 */

import type {
  ConnectionState,
  WebSocketTransportConfig,
  TransportMessage,
  TransportEvent,
  EventHandler,
} from '../types.js';

/** Default configuration values */
const DEFAULT_CONFIG = {
  autoReconnect: true,
  maxReconnectAttempts: 5,
  reconnectDelay: 1000,
};

/**
 * WebSocket transport for NHP communication in browser environments
 */
export class WebSocketTransport {
  private ws: WebSocket | null = null;
  private config: Required<WebSocketTransportConfig>;
  private state: ConnectionState = 'disconnected';
  private reconnectAttempts = 0;
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
  private eventHandlers: Map<TransportEvent, Set<EventHandler>> = new Map();

  constructor(config: WebSocketTransportConfig) {
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
   * Connect to the WebSocket server
   */
  async connect(): Promise<void> {
    if (this.state === 'connected' || this.state === 'connecting') {
      return;
    }

    this.state = 'connecting';

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.config.url);
        this.ws.binaryType = 'arraybuffer';

        this.ws.onopen = () => {
          this.state = 'connected';
          this.reconnectAttempts = 0;
          this.emit('open', undefined);
          resolve();
        };

        this.ws.onclose = (event) => {
          this.handleClose(event);
        };

        this.ws.onerror = (event) => {
          this.emit('error', event);
          if (this.state === 'connecting') {
            reject(new Error('WebSocket connection failed'));
          }
        };

        this.ws.onmessage = (event) => {
          const data = new Uint8Array(event.data as ArrayBuffer);
          const message: TransportMessage = { data };
          this.emit('message', message);
        };
      } catch (err) {
        this.state = 'disconnected';
        reject(err);
      }
    });
  }

  /**
   * Disconnect from the WebSocket server
   */
  disconnect(): void {
    this.clearReconnectTimeout();
    this.config.autoReconnect = false;

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.state = 'disconnected';
  }

  /**
   * Send data over the WebSocket
   */
  send(data: Uint8Array): void {
    if (!this.ws || this.state !== 'connected') {
      throw new Error('WebSocket is not connected');
    }

    this.ws.send(data);
  }

  /**
   * Register an event handler
   */
  on(event: TransportEvent, handler: EventHandler): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    this.eventHandlers.get(event)!.add(handler);
  }

  /**
   * Remove an event handler
   */
  off(event: TransportEvent, handler: EventHandler): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.delete(handler);
    }
  }

  /**
   * Check if the transport is connected
   */
  isConnected(): boolean {
    return this.state === 'connected';
  }

  private emit(event: TransportEvent, data: unknown): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.forEach((handler) => handler(data));
    }
  }

  private handleClose(event: CloseEvent): void {
    const wasConnected = this.state === 'connected';
    this.ws = null;

    this.emit('close', event);

    if (this.config.autoReconnect && wasConnected) {
      this.scheduleReconnect();
    } else {
      this.state = 'disconnected';
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      this.state = 'disconnected';
      this.emit('error', new Error('Max reconnection attempts reached'));
      return;
    }

    this.state = 'reconnecting';
    this.reconnectAttempts++;

    const delay = this.config.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    this.reconnectTimeout = setTimeout(() => {
      this.connect().catch(() => {
        this.scheduleReconnect();
      });
    }, delay);
  }

  private clearReconnectTimeout(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
  }
}
