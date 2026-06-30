/**
 * NHPAgent - Main OpenNHP Agent SDK Class
 * Provides a high-level API for NHP authentication and connection management
 */

import type {
  NHPAgentConfig,
  ServerConfig,
  ResourceConfig,
  KnockResult,
  NHPAgentEvent,
  EventHandler,
  KeyPairBase64,
  Logger,
  LogLevel,
  AgentKnockMsg,
  ServerKnockAckMsg,
  AgentIdentity,
  AgentOTPMsg,
  AgentRegisterMsg,
  ServerRegisterAckMsg,
  OtpResult,
  RegisterResult,
  ParsedPacket,
} from './types.js';
import { generateX25519KeyPairBase64, derivePublicKeyFromBase64 } from './crypto/ecdh.js';
import { generateSM2KeyPairBase64, deriveSM2PublicKeyFromBase64 } from './crypto/sm2.js';
import { randomBytes, bytesToHex, pubKeyFingerprintFromBase64 } from './crypto/utils.js';
import { buildNHPPacket, parseNHPPacket, clearServerCookie, PacketContext } from './protocol/packet.js';
import { NHP_PACKET_TYPES } from './protocol/constants.js';
import { WebSocketTransport } from './transport/websocket.js';
import { UdpTransport } from './transport/udp.js';
import { HttpRelayTransport } from './transport/relay.js';

/** Common transport interface */
interface Transport {
  connect(): Promise<void>;
  disconnect(): void;
  send(data: Uint8Array): void;
  on(event: string, handler: (data: unknown) => void): void;
  off(event: string, handler: (data: unknown) => void): void;
  isConnected(): boolean;
}

/** Detect if running in browser or Node.js */
const isBrowser = typeof window !== 'undefined' && typeof window.document !== 'undefined';

/** Default agent configuration */
const DEFAULT_CONFIG: Required<Omit<NHPAgentConfig, 'privateKey' | 'relayUrl'>> = {
  cipherScheme: 'curve25519',
  logLevel: 'error',
  transport: isBrowser ? 'relay' : 'udp',
};

/**
 * OpenNHP Agent SDK
 *
 * @example
 * ```typescript
 * const agent = new NHPAgent({
 *   cipherScheme: 'curve25519',
 *   logLevel: 'info'
 * });
 *
 * await agent.init();
 * agent.setIdentity({
 *   userId: 'user123',
 *   deviceId: 'device456',
 *   organizationId: 'opennhp.org'
 * });
 * agent.addServer({
 *   publicKey: 'abc123...',
 *   host: 'nhp.example.com',
 *   port: 62206
 * });
 *
 * const result = await agent.knockResource({
 *   resourceId: 'demo',
 *   serviceId: 'example',
 *   serverHost: 'nhp.example.com',
 *   serverPort: 62206
 * });
 *
 * if (result.success) {
 *   console.log('Access granted until:', result.expiresAt);
 *   console.log('Resource hosts:', result.resourceHosts);
 * }
 *
 * await agent.close();
 * ```
 */
export class NHPAgent {
  /**
   * SDK version, mirrored from `nhp/version/VERSION` at build time so it
   * matches the version stamped into the OpenNHP Go binaries.
   */
  static readonly version: string = __SDK_VERSION__;

  private config: Required<Omit<NHPAgentConfig, 'privateKey' | 'relayUrl'>> & { privateKey?: string; relayUrl?: string };
  private keyPair: KeyPairBase64 | null = null;
  private identity: AgentIdentity | null = null;
  private servers: Map<string, ServerConfig> = new Map();
  private transports: Map<string, Transport> = new Map();
  private eventHandlers: Map<NHPAgentEvent, Set<EventHandler>> = new Map();
  private initialized = false;
  private readonly packetContext = new PacketContext();
  private readonly logger: Logger = {
    error: (msg: string, ...args: unknown[]) => this.logRaw('error', msg, args),
    info: (msg: string, ...args: unknown[]) => this.logRaw('info', msg, args),
    debug: (msg: string, ...args: unknown[]) => this.logRaw('debug', msg, args),
  };

  constructor(config: NHPAgentConfig = {}) {
    this.config = {
      ...DEFAULT_CONFIG,
      ...config,
    };
  }

  /**
   * Initialize the agent
   * Generates key pair if not provided in config
   */
  async init(): Promise<void> {
    if (this.initialized) {
      return;
    }

    const isGMSM = this.config.cipherScheme === 'gmsm';

    if (this.config.privateKey) {
      // Derive public key from provided private key
      const publicKey = isGMSM
        ? deriveSM2PublicKeyFromBase64(this.config.privateKey)
        : derivePublicKeyFromBase64(this.config.privateKey);
      this.keyPair = {
        privateKey: this.config.privateKey,
        publicKey,
      };
      this.log('info', `Using provided private key (${isGMSM ? 'SM2' : 'X25519'})`);
    } else {
      // Generate new key pair based on cipher scheme
      this.keyPair = isGMSM
        ? generateSM2KeyPairBase64()
        : generateX25519KeyPairBase64();
      this.log('info', `Generated new ${isGMSM ? 'SM2' : 'X25519'} key pair`);
    }

    this.initialized = true;
    this.log('info', `NHPAgent initialized with ${isGMSM ? 'GMSM' : 'CURVE25519'} cipher scheme`);
  }

  /**
   * Close the agent and cleanup resources
   */
  async close(): Promise<void> {
    // Disconnect all transports
    for (const transport of this.transports.values()) {
      transport.disconnect();
    }
    this.transports.clear();

    // Clear all per-server packet state (cookies, replay timestamps, chain keys).
    this.packetContext.clear();
    this.servers.clear();

    this.initialized = false;
    this.log('info', 'NHPAgent closed');
  }

  /**
   * Set the agent identity for knock requests
   */
  setIdentity(identity: AgentIdentity): void {
    this.identity = identity;
    this.log('debug', `Identity set: userId=${identity.userId}, deviceId=${identity.deviceId}`);
  }

  /**
   * Set the user identity for knock requests (legacy method)
   * @deprecated Use setIdentity instead
   */
  setUser(userId: string, organizationId?: string): void {
    // Generate a device ID if not using setIdentity
    const deviceId = bytesToHex(randomBytes(16));
    this.setIdentity({
      userId,
      deviceId,
      organizationId,
    });
  }

  /**
   * Add a server configuration
   */
  addServer(config: ServerConfig): void {
    const serverId = config.id ?? (config.host ? `${config.host}:${config.port}` : config.publicKey.substring(0, 16));
    this.servers.set(serverId, { ...config, id: serverId });
    this.log('debug', `Server added: ${serverId}`);
  }

  /**
   * Remove a server configuration
   */
  removeServer(serverId: string): void {
    const server = this.servers.get(serverId);
    if (server) {
      clearServerCookie(server.publicKey, this.packetContext);
      this.servers.delete(serverId);

      const transport = this.transports.get(serverId);
      if (transport) {
        transport.disconnect();
        this.transports.delete(serverId);
      }

      this.log('debug', `Server removed: ${serverId}`);
    }
  }

  /**
   * Get the agent's public key (base64 encoded)
   */
  getPublicKey(): string {
    if (!this.keyPair) {
      throw new Error('Agent not initialized');
    }
    return this.keyPair.publicKey;
  }

  /**
   * Get the agent's private key (base64 encoded).
   * Call only after init(); handle with care — this is secret material.
   */
  getPrivateKey(): string {
    if (!this.keyPair) {
      throw new Error('Agent not initialized');
    }
    return this.keyPair.privateKey;
  }

  /**
   * Knock on a resource to request access
   */
  async knockResource(resource: ResourceConfig): Promise<KnockResult> {
    if (!this.initialized || !this.keyPair) {
      return {
        success: false,
        error: 'Agent not initialized',
        errorCode: 1,
      };
    }

    if (!this.identity) {
      return {
        success: false,
        error: 'Identity not set. Call setIdentity() first.',
        errorCode: 2,
      };
    }

    let server: ServerConfig | undefined;
    let serverId: string;

    if (resource.serverHost && resource.serverPort) {
      // Direct mode: lookup by host:port
      serverId = `${resource.serverHost}:${resource.serverPort}`;
      server = this.servers.get(serverId);
    } else {
      // Relay mode: use the first (and typically only) registered server
      const first = this.servers.entries().next();
      if (!first.done) {
        [serverId, server] = first.value;
      } else {
        serverId = '';
      }
    }

    if (!server) {
      return {
        success: false,
        error: `Server not configured${serverId ? ': ' + serverId : ''}. Call addServer() first.`,
        errorCode: 3,
      };
    }

    try {
      // Build knock message matching Go AgentKnockMsg structure
      const knockMsg: AgentKnockMsg = {
        usrId: this.identity.userId,
        devId: this.identity.deviceId,
        orgId: this.identity.organizationId,
        aspId: resource.serviceId, // ASP ID maps to service ID
        resId: resource.resourceId,
      };

      // First attempt with KNK packet type
      let result = await this.sendKnock(NHP_PACKET_TYPES.KNK, knockMsg, server, resource);

      // If server returned cookie (overloaded), resend with RNK
      if (result.type === NHP_PACKET_TYPES.COK) {
        this.log('info', 'Server overloaded, resending with cookie');
        result = await this.sendKnock(NHP_PACKET_TYPES.RNK, knockMsg, server, resource);
      }

      // Process ACK response
      if (result.type === NHP_PACKET_TYPES.ACK) {
        return this.parseAckResponse(result.message);
      }

      return {
        success: false,
        error: `Unexpected response type: ${result.type}`,
        errorCode: 4,
      };
    } catch (err) {
      const error = err instanceof Error ? err.message : 'Unknown error';
      this.log('error', `Knock failed: ${error}`);
      this.emit('error', err);

      return {
        success: false,
        error,
        errorCode: 5,
      };
    }
  }

  /**
   * Exit/release access to a resource
   */
  async exitResource(resource: ResourceConfig): Promise<void> {
    const serverId = `${resource.serverHost}:${resource.serverPort}`;
    const server = this.servers.get(serverId);

    if (server) {
      clearServerCookie(server.publicKey, this.packetContext);
    }

    const transport = this.transports.get(serverId);
    if (transport) {
      transport.disconnect();
      this.transports.delete(serverId);
    }

    this.log('debug', `Exited resource: ${resource.resourceId}`);
  }

  /**
   * Request a one-time password for agent registration.
   *
   * The OTP is delivered out-of-band (e.g., via email) by the ASP plugin.
   * This method is fire-and-forget — it sends the request but does not wait
   * for a response (the server does not send an NHP-level reply for OTP).
   *
   * @param serviceId - Auth Service Provider ID (maps to the ASP plugin)
   * @param userData - Optional user data (include `email` key for email delivery)
   * @returns Result indicating whether the request was sent successfully
   */
  async requestOtp(serviceId: string, userData?: Record<string, unknown>): Promise<OtpResult> {
    if (!this.initialized || !this.keyPair) {
      return { success: false, error: 'Agent not initialized' };
    }
    if (!this.identity) {
      return { success: false, error: 'Identity not set. Call setIdentity() first.' };
    }

    const server = this.getDefaultServer();
    if (!server) {
      return { success: false, error: 'No server configured. Call addServer() first.' };
    }

    try {
      const otpMsg: AgentOTPMsg = {
        usrId: this.identity.userId,
        devId: this.identity.deviceId,
        orgId: this.identity.organizationId,
        aspId: serviceId,
        usrData: userData,
      };

      const message = JSON.stringify(otpMsg);
      const packet = await buildNHPPacket(
        NHP_PACKET_TYPES.OTP,
        this.keyPair.privateKey,
        this.keyPair.publicKey,
        server.publicKey,
        message,
        true, // compress
        this.config.cipherScheme,
        this.packetContext
      );

      // OTP is fire-and-forget — the server does not send an NHP response.
      // Use a direct fetch to the relay without waiting for a response body
      // (the relay will time out waiting for a non-existent server reply).
      const serverId = await pubKeyFingerprintFromBase64(server.publicKey);
      const url = `${this.config.relayUrl}/${serverId}`;
      // Note: Uint8Array is not a valid fetch body; wrap in Blob.
      fetch(url, { method: 'POST', body: new Blob([packet as BlobPart]) }).catch(() => {
        // Ignore errors — the relay may return 504 after UDP timeout.
      });

      this.log('info', `OTP request sent for user=${this.identity.userId} device=${this.identity.deviceId}`);
      return { success: true };
    } catch (err) {
      const error = err instanceof Error ? err.message : 'Unknown error';
      this.log('error', `OTP request failed: ${error}`);
      return { success: false, error };
    }
  }

  /**
   * Register the agent's public key with the server.
   *
   * The agent's current key pair (generated during init or set via config)
   * is registered. The server validates the OTP and stores the public key.
   *
   * @param serviceId - Auth Service Provider ID
   * @param otp - One-time password received via email
   * @param userData - Optional user data
   * @returns Registration result
   */
  async registerPublicKey(
    serviceId: string,
    otp: string,
    userData?: Record<string, unknown>
  ): Promise<RegisterResult> {
    if (!this.initialized || !this.keyPair) {
      return { success: false, error: 'Agent not initialized', errorCode: '1' };
    }
    if (!this.identity) {
      return { success: false, error: 'Identity not set. Call setIdentity() first.', errorCode: '2' };
    }

    const server = this.getDefaultServer();
    if (!server) {
      return { success: false, error: 'No server configured. Call addServer() first.', errorCode: '3' };
    }

    try {
      const regMsg: AgentRegisterMsg = {
        usrId: this.identity.userId,
        devId: this.identity.deviceId,
        orgId: this.identity.organizationId,
        aspId: serviceId,
        otp,
        usrData: userData,
      };

      const message = JSON.stringify(regMsg);
      const packet = await buildNHPPacket(
        NHP_PACKET_TYPES.REG,
        this.keyPair.privateKey,
        this.keyPair.publicKey,
        server.publicKey,
        message,
        true, // compress
        this.config.cipherScheme,
        this.packetContext
      );

      const transport = await this.getOrCreateTransport(server);
      const zeroChainKey = new Uint8Array(32);
      const response = await this.sendAndWaitForResponse(transport, packet, server.publicKey, zeroChainKey);

      if (response.type !== (NHP_PACKET_TYPES.RAK as number)) {
        return {
          success: false,
          error: `Unexpected response type: ${response.type}`,
          errorCode: '4',
        };
      }

      const ack: ServerRegisterAckMsg = JSON.parse(response.message);

      if (ack.errCode && ack.errCode !== '' && ack.errCode !== '0') {
        return {
          success: false,
          error: ack.errMsg || `Registration failed: ${ack.errCode}`,
          errorCode: ack.errCode,
        };
      }

      this.log('info', `Registration successful for user=${this.identity.userId} device=${this.identity.deviceId} expiresAt=${ack.expiresAt ?? 'never'}`);
      // Convert expiresAt from unix-seconds (Go wire format) to
      // unix-milliseconds to match ServerConfig.expiresAt / KnockResult.expiresAt.
      const expiresAtMs = ack.expiresAt !== undefined ? ack.expiresAt * 1000 : undefined;
      return { success: true, expiresAt: expiresAtMs };
    } catch (err) {
      const error = err instanceof Error ? err.message : 'Unknown error';
      this.log('error', `Registration failed: ${error}`);
      return { success: false, error, errorCode: '5' };
    }
  }

  /**
   * Get the default (first configured) server.
   */
  private getDefaultServer(): ServerConfig | undefined {
    const first = this.servers.entries().next();
    if (first.done) {
      return undefined;
    }
    return first.value[1];
  }

  /**
   * Get or create transport for a server.
   */
  private async getOrCreateTransport(server: ServerConfig): Promise<Transport> {
    const serverId = server.id!;
    let transport = this.transports.get(serverId);
    if (!transport) {
      transport = await this.createTransport(server.host ?? '', server.port ?? 0, server.publicKey);
      this.transports.set(serverId, transport);
    }
    return transport;
  }

  /**
   * Register an event handler
   */
  on(event: NHPAgentEvent, handler: EventHandler): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, new Set());
    }
    this.eventHandlers.get(event)!.add(handler);
  }

  /**
   * Remove an event handler
   */
  off(event: NHPAgentEvent, handler: EventHandler): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.delete(handler);
    }
  }

  private async sendKnock(
    packetType: number,
    knockMsg: AgentKnockMsg,
    server: ServerConfig,
    resource: ResourceConfig
  ): Promise<ParsedPacket> {
    if (!this.keyPair) {
      throw new Error('Agent not initialized');
    }

    // Authenticate the knock type: mirror the wire packet type into the
    // AEAD-protected body so the server can require body == wire and reject
    // on-path header-type flips (NHP_KNK <-> NHP_EXT). Both the wire packet
    // (below) and the body derive from this same `packetType`, so they cannot
    // drift. Spread rather than mutate `knockMsg` so the caller's shared object
    // can be reused for the cookie-resend RNK without the KNK type leaking.
    const knockMessage = JSON.stringify({ ...knockMsg, headerType: packetType });

    // Build knock packet
    const packet = await buildNHPPacket(
      packetType,
      this.keyPair.privateKey,
      this.keyPair.publicKey,
      server.publicKey,
      knockMessage,
      true, // compress
      this.config.cipherScheme,
      this.packetContext
    );

    this.log('debug', `${packetType === NHP_PACKET_TYPES.KNK ? 'KNK' : 'RNK'} packet built: ${packet.length} bytes`);
    this.emit('knock', { resource, packetType, packet });

    // Get or create transport
    const serverId = server.id!;
    let transport = this.transports.get(serverId);
    if (!transport) {
      transport = await this.createTransport(server.host ?? '', server.port ?? 0, server.publicKey);
      this.transports.set(serverId, transport);
    }

    // The Go server's encryptBody/decryptBody both clear chainKey via defer,
    // so both sides of the Noise chain effectively restart from all-zeros.
    // Pass an all-zeros prevChainKey to match this behavior.
    const zeroChainKey = new Uint8Array(32);

    // Send packet and wait for response
    return this.sendAndWaitForResponse(transport, packet, server.publicKey, zeroChainKey);
  }

  private parseAckResponse(message: string): KnockResult {
    try {
      const ackMsg: ServerKnockAckMsg = JSON.parse(message);

      // Check for error — errCode "0" or "" means success
      if (ackMsg.errCode && ackMsg.errCode !== '' && ackMsg.errCode !== '0') {
        return {
          success: false,
          error: ackMsg.errMsg || `Server error: ${ackMsg.errCode}`,
          errorCode: parseInt(ackMsg.errCode) || 6,
        };
      }

      this.log('info', 'Knock successful');
      this.emit('ack', ackMsg);

      return {
        success: true,
        expiresAt: Date.now() + ackMsg.opnTime * 1000,
        accessToken: ackMsg.aspToken,
        resourceHosts: ackMsg.resHost,
        agentAddress: ackMsg.agentAddr,
        preAccessUrl: ackMsg.preAccUrl,
      };
    } catch (err) {
      const detail = err instanceof Error ? err.message : String(err);
      return {
        success: false,
        error: `Failed to parse server response: ${detail}`,
        errorCode: 7,
      };
    }
  }

  private async createTransport(
    host: string,
    port: number,
    serverPublicKey: string
  ): Promise<Transport> {
    const transportType = this.config.transport;

    switch (transportType) {
      case 'udp':
        this.log('debug', `Creating UDP transport to ${host}:${port}`);
        return new UdpTransport({ host, port }) as Transport;

      case 'websocket':
        this.log('debug', `Creating WebSocket transport to ${host}:${port}`);
        // Handle host that may already include protocol prefix
        const wsHost = host.replace(/^https?:\/\//, '');
        return new WebSocketTransport({
          url: `wss://${wsHost}:${port}/nhp`,
          autoReconnect: false,
        }) as Transport;

      case 'relay': {
        const relayUrl = this.config.relayUrl;
        if (!relayUrl) {
          throw new Error(
            '[NHPAgent] transport="relay" requires relayUrl to be set in NHPAgentConfig'
          );
        }
        // The server ID is derived from the target server's public key so
        // the agent can address any server on a multi-server relay
        // without explicit configuration. The same algorithm runs in Go
        // (utils.PubKeyFingerprint), keeping the routing identifier
        // canonical on both sides.
        const serverId = await pubKeyFingerprintFromBase64(serverPublicKey);
        this.log('debug', `Creating HTTP relay transport via ${relayUrl}/${serverId}`);
        return new HttpRelayTransport({
          relayUrl,
          serverId,
          logger: this.logger,
        }) as unknown as Transport;
      }

      default:
        throw new Error(`Unsupported transport type: ${transportType}`);
    }
  }

  private async sendAndWaitForResponse(
    transport: Transport,
    packet: Uint8Array,
    serverPublicKey: string,
    prevChainKey?: Uint8Array
  ): Promise<ParsedPacket> {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(() => {
        transport.off('message', messageHandler);
        reject(new Error('Request timeout'));
      }, 10000); // 10 second timeout

      const messageHandler = async (msg: unknown) => {
        clearTimeout(timeout);
        transport.off('message', messageHandler);

        try {
          if (!this.keyPair) {
            throw new Error('Agent not initialized');
          }

          const message = msg as { data: Uint8Array };

          // Try parsing with the previous chain key first (real server
          // continues the Noise chain from the KNK).  If that fails, fall
          // back to a fresh chain key (self-test / standalone ACK).  If both
          // fail, surface both errors so HMAC/replay/key-mismatch failures
          // aren't masked as a generic "parse error".
          let parsed: ParsedPacket;
          if (prevChainKey) {
            try {
              parsed = await parseNHPPacket(
                message.data,
                this.keyPair.privateKey,
                this.keyPair.publicKey,
                serverPublicKey,
                prevChainKey,
                this.packetContext
              );
            } catch (firstErr) {
              this.log('debug', `parseNHPPacket with prevChainKey failed: ${(firstErr as Error).message}; retrying without`);
              try {
                parsed = await parseNHPPacket(
                  message.data,
                  this.keyPair.privateKey,
                  this.keyPair.publicKey,
                  serverPublicKey,
                  undefined,
                  this.packetContext
                );
              } catch (secondErr) {
                throw new Error(
                  `parseNHPPacket failed (chained: ${(firstErr as Error).message}; fresh: ${(secondErr as Error).message})`
                );
              }
            }
          } else {
            parsed = await parseNHPPacket(
              message.data,
              this.keyPair.privateKey,
              this.keyPair.publicKey,
              serverPublicKey,
              undefined,
              this.packetContext
            );
          }

          resolve(parsed);
        } catch (err) {
          reject(err);
        }
      };

      transport.on('message', messageHandler);

      transport
        .connect()
        .then(() => {
          transport.send(packet);
        })
        .catch((err) => {
          clearTimeout(timeout);
          transport.off('message', messageHandler);
          reject(err);
        });
    });
  }

  private emit(event: NHPAgentEvent, data: unknown): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.forEach((handler) => handler(data));
    }
  }

  private log(level: LogLevel, message: string): void {
    if (level === 'silent') return;
    this.logRaw(level, message, []);
  }

  private logRaw(level: Exclude<LogLevel, 'silent'>, message: string, args: unknown[]): void {
    const levels: Record<LogLevel, number> = {
      silent: 0,
      error: 1,
      info: 2,
      debug: 3,
    };

    if (levels[level] > levels[this.config.logLevel]) {
      return;
    }

    const prefix = `[NHPAgent:${level.toUpperCase()}]`;
    switch (level) {
      case 'error':
        console.error(prefix, message, ...args);
        break;
      case 'info':
        console.info(prefix, message, ...args);
        break;
      case 'debug':
        console.debug(prefix, message, ...args);
        break;
    }
  }
}
