/**
 * NHP Protocol Constants
 */

/** Maximum packet buffer size */
export const PACKET_BUFFER_SIZE = 4096;

/** Standard header size (X25519) */
export const HEADER_SIZE = 240;

/** Extended header size (GM SM2) */
export const HEADER_EX_SIZE = 304;

/** Initial chain key derivation string */
export const INITIAL_CHAIN_KEY_STRING = 'NHP keygen v.20230421@clouddeep.cn';

/** Initial hash derivation string */
export const INITIAL_HASH_STRING = 'NHP hashgen v.20230421@deepcloudsdp.com';

/** Minimum interval between received packets (milliseconds) */
export const MINIMAL_RECV_INTERVAL_MS = 2;

/** Packet type values */
export const NHP_PACKET_TYPES = {
  KNK: 1,   // Knock
  ACK: 2,   // Acknowledge
  AOP: 3,   // Agent Operation
  ART: 4,   // Agent Report
  LST: 5,   // List
  LRT: 6,   // List Report
  COK: 7,   // Cookie
  RNK: 8,   // Re-knock
  RLY: 9,   // Relay
  AOL: 10,  // Agent Online
} as const;

/** Protocol version */
export const PROTOCOL_VERSION = {
  MAJOR: 1,
  MINOR: 0,
} as const;

/** Header field offsets for standard header */
export const HEADER_OFFSETS = {
  PREAMBLE: 0,
  TYPE_AND_SIZE: 4,
  VERSION: 8,
  FLAGS: 10,
  RESERVED: 12,
  COUNTER: 16,
  EPHEMERAL: 24,
  IDENTITY: 56,
  STATIC: 136,
  TIMESTAMP: 184,
  HMAC: 208,
} as const;

/** Header field offsets for extended header (GM crypto) */
export const HEADER_EX_OFFSETS = {
  PREAMBLE: 0,
  TYPE_AND_SIZE: 4,
  VERSION: 8,
  FLAGS: 10,
  RESERVED: 12,
  COUNTER: 16,
  EPHEMERAL: 24,
  IDENTITY: 88,
  STATIC: 168,
  TIMESTAMP: 248,
  HMAC: 272,
} as const;

/** Field sizes in bytes */
export const FIELD_SIZES = {
  PREAMBLE: 4,
  TYPE_AND_SIZE: 4,
  VERSION: 2,
  FLAGS: 2,
  RESERVED: 4,
  COUNTER: 8,
  NONCE: 12,
  X25519_KEY: 32,
  SM2_KEY: 64,
  IDENTITY: 80,
  X25519_STATIC: 48,
  SM2_STATIC: 80,
  TIMESTAMP: 24,
  HMAC: 32,
  AEAD_TAG: 16,
} as const;

/** Flag bit positions */
export const FLAG_BITS = {
  EXTENDED: 0x1,
  COMPRESSED: 0x2,
} as const;

/** Stale packet threshold (10 minutes in nanoseconds) */
export const STALE_PACKET_THRESHOLD_NS = BigInt(600 * 1000 * 1000 * 1000);

/** Flood packet threshold (20 milliseconds in nanoseconds) */
export const FLOOD_PACKET_THRESHOLD_NS = BigInt(20 * 1000 * 1000);
