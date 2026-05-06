/**
 * UDP Transport for Node.js Environments
 * Provides connection management and message handling for NHP over UDP
 *
 * Note: This transport only works in Node.js environments.
 * For browser environments, use WebRTCTransport.
 */

import type {
  ConnectionState,
  TransportMessage,
  TransportEvent,
  EventHandler,
} from '../types.js';

/** UDP transport configuration */
export interface UdpTransportConfig {
  /** Server hostname or IP address */
  host: string;
  /** Server port number (default: 62206) */
  port?: number;
  /** Connection timeout in milliseconds (default: 10000) */
  timeout?: number;
}

/** Default configuration values */
const DEFAULT_CONFIG = {
  port: 62206,
  timeout: 10000,
};

/**
 * UDP transport for NHP communication in Node.js environments
 */
export class UdpTransport {
  private socket: any = null; // dgram.Socket
  private config: Required<UdpTransportConfig>;
  private state: ConnectionState = 'disconnected';
  private eventHandlers: Map<TransportEvent, Set<EventHandler>> = new Map();
  private dgram: any = null;

  constructor(config: UdpTransportConfig) {
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
   * Connect to the UDP server
   */
  async connect(): Promise<void> {
    if (this.state === 'connected' || this.state === 'connecting') {
      return;
    }

    this.state = 'connecting';

    // Dynamic import for Node.js dgram module
    try {
      this.dgram = await import('dgram');
    } catch {
      this.state = 'disconnected';
      throw new Error('UDP transport is only available in Node.js environments');
    }

    return new Promise((resolve, reject) => {
      try {
        this.socket = this.dgram.createSocket('udp4');

        this.socket.on('error', (err: Error) => {
          this.emit('error', err);
          if (this.state === 'connecting') {
            this.state = 'disconnected';
            reject(err);
          }
        });

        this.socket.on('message', (msg: Buffer, rinfo: { address: string; port: number }) => {
          const message: TransportMessage = {
            data: new Uint8Array(msg),
            remoteAddress: rinfo.address,
            remotePort: rinfo.port,
          };
          this.emit('message', message);
        });

        this.socket.on('close', () => {
          this.state = 'disconnected';
          this.emit('close', undefined);
        });

        // Bind to any available port
        this.socket.bind(() => {
          this.state = 'connected';
          this.emit('open', undefined);
          resolve();
        });
      } catch (err) {
        this.state = 'disconnected';
        reject(err);
      }
    });
  }

  /**
   * Disconnect from the UDP server
   */
  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
    this.state = 'disconnected';
  }

  /**
   * Send data over UDP to the configured server
   */
  send(data: Uint8Array): void {
    if (!this.socket || this.state !== 'connected') {
      throw new Error('UDP socket is not connected');
    }

    const buffer = Buffer.from(data);
    this.socket.send(buffer, 0, buffer.length, this.config.port, this.config.host, (err: Error | null) => {
      if (err) {
        this.emit('error', err);
      }
    });
  }

  /**
   * Send data and wait for a response with timeout
   */
  async sendAndReceive(data: Uint8Array, timeout?: number): Promise<TransportMessage> {
    const timeoutMs = timeout ?? this.config.timeout;

    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        this.off('message', messageHandler as EventHandler);
        reject(new Error('UDP request timeout'));
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
}
