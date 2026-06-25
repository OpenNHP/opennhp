/**
 * OpenNHP JavaScript/TypeScript SDK
 *
 * @packageDocumentation
 * @module @opennhp/agent
 *
 * @example
 * ```typescript
 * import { NHPAgent } from '@opennhp/agent';
 *
 * const agent = new NHPAgent({
 *   cipherScheme: 'curve25519',
 *   logLevel: 'info'
 * });
 *
 * await agent.init();
 * agent.setUser('user123', 'opennhp.org');
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
 * }
 *
 * await agent.close();
 * ```
 */

// Main agent class
export { NHPAgent } from './NHPAgent.js';

// Type exports
export type {
  CipherScheme,
  LogLevel,
  TransportType,
  NHPAgentConfig,
  ServerConfig,
  ResourceConfig,
  KnockResult,
  NHPAgentEvent,
  EventHandler,
  KeyPair,
  KeyPairBase64,
  PacketType,
  ProtocolVersion,
  HeaderFlags,
  ParsedPacket,
  ConnectionState,
  TransportEvent,
  WebSocketTransportConfig,
  TransportMessage,
  AgentKnockMsg,
  ServerKnockAckMsg,
  AgentIdentity,
  AgentOTPMsg,
  AgentRegisterMsg,
  ServerRegisterAckMsg,
  OtpResult,
  RegisterResult,
} from './types.js';

// Crypto utilities (for advanced usage)
export {
  generateX25519KeyPair,
  generateX25519KeyPairBase64,
  bytesToBase64,
  base64ToBytes,
  stringToBytes,
  bytesToString,
} from './crypto/index.js';

// Protocol utilities (for advanced usage)
export {
  buildNHPPacket,
  parseNHPPacket,
  NHPHeader,
  NHPHeaderEx,
  NHP_PACKET_TYPES,
  PROTOCOL_VERSION,
} from './protocol/index.js';
export type { INHPHeader, TypeAndPayloadSize } from './protocol/index.js';

// Transport (for advanced usage)
export { WebSocketTransport, UdpTransport, HttpRelayTransport } from './transport/index.js';
export type { UdpTransportConfig, HttpRelayTransportConfig } from './transport/index.js';
