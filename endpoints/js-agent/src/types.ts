/**
 * TypeScript interfaces for OpenNHP Agent SDK
 */

/** Cipher scheme for cryptographic operations */
export type CipherScheme = 'curve25519' | 'gmsm';

/** Log level for SDK output */
export type LogLevel = 'silent' | 'error' | 'info' | 'debug';

/**
 * Minimal logger interface used by transports.
 * Transports should not call console.* directly — they receive a logger that
 * routes through the NHPAgent's `logLevel` setting.
 */
export interface Logger {
  error(message: string, ...args: unknown[]): void;
  info(message: string, ...args: unknown[]): void;
  debug(message: string, ...args: unknown[]): void;
}

/** Transport type for NHP communication */
export type TransportType = 'udp' | 'websocket' | 'relay';

/** Configuration for initializing the NHP Agent */
export interface NHPAgentConfig {
  /** Base64-encoded private key. If not provided, one will be generated */
  privateKey?: string;
  /** Cipher scheme to use for cryptographic operations */
  cipherScheme?: CipherScheme;
  /** Logging level */
  logLevel?: LogLevel;
  /** Transport type to use (default: 'udp' for Node.js, 'relay' for browser) */
  transport?: TransportType;
  /**
   * Full URL of the HTTP relay endpoint (required when transport='relay').
   * Example: "https://relay.example.com/relay"
   */
  relayUrl?: string;
}

/** Configuration for an NHP server */
export interface ServerConfig {
  /** Unique identifier for the server (auto-generated from host:port if omitted) */
  id?: string;
  /** Base64-encoded public key of the server */
  publicKey: string;
  /** Server hostname or IP address (required for udp/websocket, optional for relay) */
  host?: string;
  /** Server port number (required for udp/websocket, optional for relay) */
  port?: number;
  /** Optional expiration timestamp (Unix milliseconds) */
  expiresAt?: number;
}

/** Configuration for a resource to knock */
export interface ResourceConfig {
  /** Resource identifier */
  resourceId: string;
  /** Service identifier */
  serviceId: string;
  /** Server hostname for the knock (optional for relay transport) */
  serverHost?: string;
  /** Server port for the knock (optional for relay transport) */
  serverPort?: number;
}

/** Result of a knock operation */
export interface KnockResult {
  /** Whether the knock was successful */
  success: boolean;
  /** Access token received on success (ASP token) */
  accessToken?: string;
  /** Expiration timestamp of the access (Unix milliseconds) */
  expiresAt?: number;
  /** Resource host mapping (service -> host:port) */
  resourceHosts?: Record<string, string>;
  /** Agent's address as seen by server */
  agentAddress?: string;
  /** Pre-access URL (for captive portal etc) */
  preAccessUrl?: string;
  /** Error message if knock failed */
  error?: string;
  /** Error code if knock failed */
  errorCode?: number;
}

/** Events emitted by the NHP Agent */
export type NHPAgentEvent = 'connected' | 'disconnected' | 'error' | 'knock' | 'ack';

/** Event handler function type */
export type EventHandler<T = unknown> = (data: T) => void;

/** X25519 Key pair */
export interface KeyPair {
  privateKey: CryptoKey;
  publicKey: CryptoKey;
}

/** Base64 encoded key pair */
export interface KeyPairBase64 {
  privateKey: string;
  publicKey: string;
}

/** NHP packet type identifiers */
export enum PacketType {
  KNK = 1,  // Knock
  ACK = 2,  // Acknowledge
  AOP = 3,  // Agent Operation
  ART = 4,  // Agent Report
  LST = 5,  // List
  LRT = 6,  // List Report
  COK = 7,  // Cookie
  RNK = 8,  // Re-knock
  RLY = 9,  // Relay
  AOL = 10, // Agent Online
  OTP = 12, // One-Time Password request
  REG = 13, // Register public key
  RAK = 14, // Register acknowledge
}

/** NHP protocol version */
export interface ProtocolVersion {
  major: number;
  minor: number;
}

/** NHP packet header flags */
export interface HeaderFlags {
  /** Whether extended header format is used (for GM crypto) */
  extended: boolean;
  /** Whether payload is compressed */
  compressed: boolean;
}

/** Parsed NHP packet result */
export interface ParsedPacket {
  /** Packet type */
  type: PacketType;
  /** Decoded message payload */
  message: string;
  /** Remote public key (base64) */
  remotePublicKey?: string;
}

/** Connection state for transport */
export type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'reconnecting';

/** Transport events */
export type TransportEvent = 'open' | 'close' | 'error' | 'message';

/** WebSocket transport configuration */
export interface WebSocketTransportConfig {
  /** WebSocket URL */
  url: string;
  /** Reconnect automatically on disconnect */
  autoReconnect?: boolean;
  /** Maximum reconnection attempts */
  maxReconnectAttempts?: number;
  /** Base delay between reconnection attempts (ms) */
  reconnectDelay?: number;
}

/** Transport message */
export interface TransportMessage {
  /** Raw packet data */
  data: Uint8Array;
  /** Source address (if applicable) */
  remoteAddress?: string;
  /** Source port (if applicable) */
  remotePort?: number;
}

/**
 * Agent Knock Message - matches Go AgentKnockMsg
 * Sent by agent to request access to a resource
 */
export interface AgentKnockMsg {
  /** User ID */
  usrId: string;
  /** Device ID */
  devId: string;
  /** Organization ID (optional) */
  orgId?: string;
  /** Auth Service Provider ID */
  aspId: string;
  /** Resource ID */
  resId: string;
  /** Check mode (optional, for validation only) */
  cknMode?: number;
  /**
   * Wire packet type, mirrored into the AEAD-authenticated knock body so the
   * server can require body == wire and reject on-path header-type flips
   * (e.g. NHP_KNK <-> NHP_EXT). Populated per send from the same value as the
   * wire packet type (KNK=1, RNK=8) — see {@link NHPAgent.sendKnock}. Matches
   * Go `AgentKnockMsg.HeaderType` (`json:"headerType"`); a server that verifies
   * header types rejects a knock whose body omits it with error 52010.
   */
  headerType?: number;
}

/**
 * Server Knock Acknowledge Message - matches Go ServerKnockAckMsg
 * Sent by server in response to a knock
 */
export interface ServerKnockAckMsg {
  /** Error code (empty string means success) */
  errCode: string;
  /** Error message (if errCode is not empty) */
  errMsg?: string;
  /** Resource host map (service -> host:port) */
  resHost: Record<string, string>;
  /** Open/access time in seconds */
  opnTime: number;
  /** Auth Service Provider token (if ASP mode) */
  aspToken?: string;
  /** Agent's address as seen by server */
  agentAddr: string;
  /** Pre-access URL (for captive portal etc) */
  preAccUrl?: string;
}

/**
 * Agent configuration for knock requests
 */
export interface AgentIdentity {
  /** User ID */
  userId: string;
  /** Device ID (generated or provided) */
  deviceId: string;
  /** Organization ID (optional) */
  organizationId?: string;
}

/**
 * Agent OTP Request Message - matches Go AgentOTPMsg
 * Sent by agent to request a one-time password for registration.
 */
export interface AgentOTPMsg {
  /** User ID */
  usrId: string;
  /** Device ID */
  devId: string;
  /** Organization ID (optional) */
  orgId?: string;
  /** Auth Service Provider ID */
  aspId: string;
  /** User data map (optional) — e.g. email address under "email" key */
  usrData?: Record<string, unknown>;
}

/**
 * Agent Register Message - matches Go AgentRegisterMsg
 * Sent by agent to register its public key with the server.
 */
export interface AgentRegisterMsg {
  /** User ID */
  usrId: string;
  /** Device ID */
  devId: string;
  /** Organization ID (optional) */
  orgId?: string;
  /** Auth Service Provider ID */
  aspId: string;
  /** One-time password received via email */
  otp: string;
  /** User data map (optional) */
  usrData?: Record<string, unknown>;
}

/**
 * Server Register Acknowledge Message - matches Go ServerRegisterAckMsg
 * Sent by server to confirm registration.
 */
export interface ServerRegisterAckMsg {
  /** Error code ("0" means success) */
  errCode: string;
  /** Error message (if errCode is not "0") */
  errMsg?: string;
  /** Auth Service Provider ID */
  aspId?: string;
}

/**
 * Result of a registration OTP request.
 */
export interface OtpResult {
  /** Whether the OTP request was sent successfully */
  success: boolean;
  /** Error message if request failed */
  error?: string;
}

/**
 * Result of a registration attempt.
 */
export interface RegisterResult {
  /** Whether registration was successful */
  success: boolean;
  /** Error message if registration failed */
  error?: string;
  /** Error code if registration failed */
  errorCode?: string;
}
