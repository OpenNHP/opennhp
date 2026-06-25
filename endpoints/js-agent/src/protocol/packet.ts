/**
 * NHP Packet Building and Parsing
 * Implements the core NHP protocol packet operations
 * Supports both CURVE25519 (Blake2s + AES-256-GCM) and GMSM (SM3 + SM4-GCM) schemes
 */

import type { PacketType, ParsedPacket, CipherScheme } from '../types.js';
import { NHPHeader, NHPHeaderEx } from './header.js';
import {
  PACKET_BUFFER_SIZE,
  HEADER_SIZE,
  INITIAL_CHAIN_KEY_STRING,
  INITIAL_HASH_STRING,
  NHP_PACKET_TYPES,
  PROTOCOL_VERSION,
  FIELD_SIZES,
  STALE_PACKET_THRESHOLD_NS,
  FLOOD_PACKET_THRESHOLD_NS,
} from './constants.js';
import {
  generateX25519KeyPairRaw,
  ecdhX25519Raw,
} from '../crypto/ecdh.js';
import { aesGcmSeal, aesGcmOpen } from '../crypto/aead.js';
import {
  newBlake2sHash,
  updateBlake2s,
  sumBlake2s,
  keyGen2,
  mixKey,
} from '../crypto/noise.js';
import {
  generateSM2KeyPair,
  sm2ECDH,
} from '../crypto/sm2.js';
import {
  sm3,
  hmacSM3,
  newSM3Hash,
  keyGenSM3_2,
  mixKeySM3,
  mixHashSM3,
} from '../crypto/sm3.js';
import { sm4GcmSeal, sm4GcmOpen } from '../crypto/sm4.js';
import {
  base64ToBytes,
  stringToBytes,
  bytesToString,
  equalBytes,
  getUnixNano,
  zlibCompress,
  zlibDecompress,
  concatBytes,
} from '../crypto/utils.js';

/**
 * Per-agent packet state.
 *
 * Holds the monotonic counter, server cookie cache, anti-replay timestamps,
 * and saved chain keys for one logical NHP agent.  Two NHPAgent instances
 * sharing process scope must NOT share a context — otherwise their counters
 * interleave, their replay-timestamp maps poison each other, and the maps
 * grow unbounded.
 */
export class PacketContext {
  // Seed counter from current time so page reloads don't trigger anti-replay on the server.
  counter: bigint = BigInt(Date.now()) * 1_000_000n;
  readonly serverCookies = new Map<string, Uint8Array>();
  readonly lastSendTimes = new Map<string, bigint>();
  readonly lastRemoteSendTimes = new Map<string, bigint>();

  // Stores the chain key produced by buildNHPPacket, keyed by remotePublicKey.
  // The NHP Noise protocol is a continuous chain: the server's ACK is encrypted
  // starting from the chain key left over after decrypting the agent's KNK.
  // parseNHPPacket uses this saved key as the starting point instead of
  // re-deriving from scratch.
  readonly lastBuildChainKeys = new Map<string, Uint8Array>();

  /** Drop all per-remote state. */
  clear(): void {
    this.serverCookies.clear();
    this.lastSendTimes.clear();
    this.lastRemoteSendTimes.clear();
    this.lastBuildChainKeys.clear();
  }

  /** Drop all state for a specific remote peer. */
  clearRemote(remotePublicKey: string): void {
    this.serverCookies.delete(remotePublicKey);
    this.lastSendTimes.delete(remotePublicKey);
    this.lastRemoteSendTimes.delete(remotePublicKey);
    this.lastBuildChainKeys.delete(remotePublicKey);
  }
}

// Default context for callers that don't supply one (test code, simple SDK
// users with one agent).  Real agents own their own PacketContext.
const defaultContext = new PacketContext();

/**
 * Retrieve and consume the chain key saved by the most recent buildNHPPacket
 * call for the given remotePublicKey.  Returns undefined if no key was saved.
 * The entry is deleted after retrieval so it cannot be re-used accidentally.
 */
export function consumeLastBuildChainKey(
  remotePublicKey: string,
  context: PacketContext = defaultContext
): Uint8Array | undefined {
  const ck = context.lastBuildChainKeys.get(remotePublicKey);
  if (ck) {
    context.lastBuildChainKeys.delete(remotePublicKey);
  }
  return ck;
}

/**
 * Build an NHP packet for transmission
 * @param type - Packet type (KNK, ACK, etc.)
 * @param privateKey - Base64-encoded local private key
 * @param publicKey - Base64-encoded local public key
 * @param remotePublicKey - Base64-encoded remote public key
 * @param message - Message payload to encrypt
 * @param compress - Whether to compress the payload
 * @param cipherScheme - Cipher scheme to use (must match the keys; required)
 * @param context - Per-agent state container (defaults to a process-wide instance)
 * @returns Encrypted packet bytes
 */
export async function buildNHPPacket(
  type: number,
  privateKey: string,
  publicKey: string,
  remotePublicKey: string,
  message: string,
  compress: boolean,
  cipherScheme: CipherScheme,
  context: PacketContext = defaultContext
): Promise<Uint8Array> {
  if (cipherScheme === 'gmsm') {
    return buildNHPPacketGMSM(type, privateKey, publicKey, remotePublicKey, message, compress, context);
  }

  const packet = new Uint8Array(PACKET_BUFFER_SIZE);
  const header = new NHPHeader(packet.buffer);

  // Convert keys from base64 to raw bytes
  const localPrivKeyBytes = base64ToBytes(privateKey);
  const localPubKeyBytes = base64ToBytes(publicKey);
  const remotePubKeyBytes = base64ToBytes(remotePublicKey);
  const msgBytes = stringToBytes(message);

  // Set header fields
  header.version = { major: PROTOCOL_VERSION.MAJOR, minor: PROTOCOL_VERSION.MINOR };
  header.flags = { extended: false, compressed: compress };
  context.counter++;
  header.counter = context.counter;
  const nonce = header.nonce;

  // Initialize chain key and hash
  const chainKey = new Uint8Array(32);
  const chainHash = new Uint8Array(32);
  const hmacHasher = newBlake2sHash();
  const chainHasher = newBlake2sHash();

  // Initialize with protocol strings
  updateBlake2s(hmacHasher, stringToBytes(INITIAL_HASH_STRING));
  updateBlake2s(chainHasher, stringToBytes(INITIAL_HASH_STRING));
  chainHash.set(sumBlake2s(chainHasher));
  chainKey.set(mixKey(chainHash, stringToBytes(INITIAL_CHAIN_KEY_STRING)));

  // Mix in remote public key
  updateBlake2s(hmacHasher, remotePubKeyBytes);
  updateBlake2s(chainHasher, remotePubKeyBytes);

  // Generate ephemeral key pair and perform ECDH
  const ephemeralKeys = generateX25519KeyPairRaw();
  const ephemeralPublicKeyBytes = ephemeralKeys.publicKey;
  header.ephemeral = ephemeralPublicKeyBytes;

  updateBlake2s(chainHasher, ephemeralPublicKeyBytes);
  chainHash.set(sumBlake2s(chainHasher));
  chainKey.set(mixKey(chainKey, ephemeralPublicKeyBytes));

  // ECDH: ephemeral private * remote public
  const ess = ecdhX25519Raw(ephemeralKeys.privateKey, remotePubKeyBytes);

  // Encrypt local public key
  const derivedKeys0 = keyGen2(chainKey, ess);
  chainKey.set(derivedKeys0.first);

  const keyStatic = aesGcmSeal(derivedKeys0.second, nonce, localPubKeyBytes, chainHash);
  header.static = keyStatic;

  updateBlake2s(chainHasher, keyStatic);
  chainHash.set(sumBlake2s(chainHasher));

  // ECDH: local private * remote public
  const ss = ecdhX25519Raw(localPrivKeyBytes, remotePubKeyBytes);

  // Encrypt timestamp
  const derivedKeys1 = keyGen2(chainKey, ss);
  chainKey.set(derivedKeys1.first);

  const tsBuf = new ArrayBuffer(8);
  const tsView = new DataView(tsBuf);
  const timestamp = getUnixNano();
  tsView.setBigUint64(0, timestamp);
  const ts = new Uint8Array(tsBuf);
  context.lastSendTimes.set(remotePublicKey, timestamp);

  const tsStatic = aesGcmSeal(derivedKeys1.second, nonce, ts, chainHash);
  header.timestamp = tsStatic;

  // Encrypt message payload
  const derivedKeys2 = keyGen2(chainKey, tsStatic);
  chainKey.set(derivedKeys2.first);
  updateBlake2s(chainHasher, tsStatic);
  chainHash.set(sumBlake2s(chainHasher));

  let payload = msgBytes;
  if (compress) {
    payload = await zlibCompress(msgBytes);
  }

  const msgStatic = aesGcmSeal(derivedKeys2.second, nonce, payload, chainHash);
  packet.set(msgStatic, header.size);

  const payloadSize = payload.byteLength + FIELD_SIZES.AEAD_TAG;
  header.typeAndPayloadSize = { type, size: payloadSize };

  // Compute HMAC
  if (type === NHP_PACKET_TYPES.RNK) {
    const cookie = context.serverCookies.get(remotePublicKey);
    if (cookie) {
      updateBlake2s(hmacHasher, cookie);
    }
  }
  updateBlake2s(hmacHasher, packet.subarray(0, header.size - FIELD_SIZES.HMAC));
  header.hmac = sumBlake2s(hmacHasher);

  // Save the final chain key so parseNHPPacket can continue the Noise chain
  // when decrypting the server's ACK response.
  context.lastBuildChainKeys.set(remotePublicKey, new Uint8Array(chainKey));

  return packet.subarray(0, header.size + payloadSize);
}

/**
 * Parse an incoming NHP packet
 * @param packet - Raw packet bytes
 * @param privateKey - Base64-encoded local private key
 * @param publicKey - Base64-encoded local public key
 * @param remotePublicKey - Base64-encoded expected remote public key
 * @returns Parsed packet with type and decrypted message
 */
export async function parseNHPPacket(
  packet: Uint8Array,
  privateKey: string,
  publicKey: string,
  remotePublicKey: string,
  prevChainKey?: Uint8Array,
  context: PacketContext = defaultContext
): Promise<ParsedPacket> {
  if (packet.length < HEADER_SIZE) {
    throw new Error('Packet size is too small');
  }

  // Check if this is an extended (GMSM) packet by reading the flags field directly.
  // The flags field is at a fixed offset (10-11) in both standard and extended headers,
  // so we can safely peek at it before deciding which parser to use.
  // Do NOT rely on packet.length >= HEADER_EX_SIZE: a GMSM packet with a small payload
  // can be shorter than 304 bytes, causing misrouting.
  // Read flags as big-endian uint16 to match DataView.getUint16 used in NHPHeader/NHPHeaderEx.
  // FLAGS field is at offset 10-11. Big-endian: high byte at [10], low byte at [11].
  const flagByte = (packet[10] << 8) | packet[11];
  if (flagByte & 0x1) {
    return parseNHPPacketGMSM(packet, privateKey, publicKey, remotePublicKey, prevChainKey, context);
  }

  // Create a clean ArrayBuffer copy to avoid SharedArrayBuffer issues
  const packetBuffer = new ArrayBuffer(packet.length);
  new Uint8Array(packetBuffer).set(packet);
  const header = new NHPHeader(packetBuffer);
  const { compressed } = header.flags;

  const { type, size } = header.typeAndPayloadSize;

  if (type !== NHP_PACKET_TYPES.ACK && type !== NHP_PACKET_TYPES.COK && type !== NHP_PACKET_TYPES.RAK) {
    throw new Error('Not an ACK, COK, or RAK packet');
  }

  if (packet.length !== header.size + size) {
    throw new Error('Wrong packet size');
  }

  const recvTime = getUnixNano();

  // Convert keys from base64 to raw bytes
  const localPrivKeyBytes = base64ToBytes(privateKey);
  const localPubKeyBytes = base64ToBytes(publicKey);
  const remotePubKeyBytes = base64ToBytes(remotePublicKey);

  // Verify HMAC
  const hmacHasher = newBlake2sHash();
  updateBlake2s(hmacHasher, stringToBytes(INITIAL_HASH_STRING));
  updateBlake2s(hmacHasher, localPubKeyBytes);
  updateBlake2s(hmacHasher, packet.subarray(0, header.size - FIELD_SIZES.HMAC));
  const checkSum = sumBlake2s(hmacHasher);

  if (!equalBytes(checkSum, header.hmac)) {
    throw new Error('HMAC check failed');
  }

  const ephemeralPublicKeyBytes = header.ephemeral;
  const nonce = header.nonce;
  const keyStatic = header.static;
  const tsStatic = header.timestamp;
  const msgStatic = packet.subarray(header.size);

  // Initialize chain key and hash.
  // The NHP Noise protocol is a continuous chain: the server encrypts its ACK
  // starting from the chain key left over after decrypting the agent's KNK.
  // When prevChainKey is provided, resume from it; otherwise start fresh.
  const chainKey = new Uint8Array(32);
  const chainHash = new Uint8Array(32);
  const chainHasher = newBlake2sHash();

  updateBlake2s(chainHasher, stringToBytes(INITIAL_HASH_STRING));
  chainHash.set(sumBlake2s(chainHasher));

  if (prevChainKey) {
    chainKey.set(prevChainKey);
  } else {
    chainKey.set(mixKey(chainHash, stringToBytes(INITIAL_CHAIN_KEY_STRING)));
  }

  updateBlake2s(chainHasher, localPubKeyBytes);
  updateBlake2s(chainHasher, ephemeralPublicKeyBytes);
  chainHash.set(sumBlake2s(chainHasher));
  chainKey.set(mixKey(chainKey, ephemeralPublicKeyBytes));

  // ECDH: local private * ephemeral public
  const ess = ecdhX25519Raw(localPrivKeyBytes, ephemeralPublicKeyBytes);

  // Decrypt remote public key
  const derivedKeys0 = keyGen2(chainKey, ess);
  chainKey.set(derivedKeys0.first);
  const decryptedPubKeyBytes = aesGcmOpen(derivedKeys0.second, nonce, keyStatic, chainHash);

  if (!equalBytes(remotePubKeyBytes, decryptedPubKeyBytes)) {
    throw new Error('Remote public key check failed');
  }

  updateBlake2s(chainHasher, keyStatic);
  chainHash.set(sumBlake2s(chainHasher));

  // ECDH: local private * remote public
  const ss = ecdhX25519Raw(localPrivKeyBytes, remotePubKeyBytes);

  // Decrypt timestamp
  const derivedKeys1 = keyGen2(chainKey, ss);
  chainKey.set(derivedKeys1.first);

  const decryptedTs = aesGcmOpen(derivedKeys1.second, nonce, tsStatic, chainHash);
  // Create a new ArrayBuffer to avoid SharedArrayBuffer issues
  const tsBuf = new ArrayBuffer(decryptedTs.length);
  new Uint8Array(tsBuf).set(decryptedTs);
  const tsView = new DataView(tsBuf);
  const remoteSendTime = tsView.getBigUint64(0);

  // Anti-replay checks — only mutate the recorded timestamp after every check
  // passes, so a rejected packet (replay/flood/stale) cannot poison the
  // accepted-window for subsequent legitimate packets.
  const lastRemoteSendTime = context.lastRemoteSendTimes.get(remotePublicKey);

  if (lastRemoteSendTime !== undefined) {
    if (remoteSendTime < lastRemoteSendTime) {
      throw new Error('Received replay packet');
    }
    if (remoteSendTime < lastRemoteSendTime + FLOOD_PACKET_THRESHOLD_NS) {
      throw new Error('Received flood packet');
    }
  }

  if (remoteSendTime < recvTime - STALE_PACKET_THRESHOLD_NS) {
    throw new Error('Received stale packet');
  }

  context.lastRemoteSendTimes.set(remotePublicKey, remoteSendTime);

  // Decrypt message
  const derivedKeys2 = keyGen2(chainKey, header.timestamp);
  chainKey.set(derivedKeys2.first);
  updateBlake2s(chainHasher, tsStatic);
  chainHash.set(sumBlake2s(chainHasher));

  let msg = aesGcmOpen(derivedKeys2.second, nonce, msgStatic, chainHash);

  if (compressed) {
    msg = await zlibDecompress(msg);
  }

  // Handle cookie packets
  if (type === NHP_PACKET_TYPES.COK) {
    context.serverCookies.set(remotePublicKey, msg);
  }

  return {
    type: type as PacketType,
    message: bytesToString(msg),
    remotePublicKey,
  };
}

/**
 * Clear stored cookies for a server (and any other per-remote state).
 */
export function clearServerCookie(
  remotePublicKey: string,
  context: PacketContext = defaultContext
): void {
  context.clearRemote(remotePublicKey);
}

/**
 * Reset the global packet counter (for testing).
 */
export function resetGlobalCounter(context: PacketContext = defaultContext): void {
  context.counter = 0n;
  context.clear();
}

/**
 * Build an NHP packet using GMSM cipher scheme (SM2/SM3/SM4)
 */
async function buildNHPPacketGMSM(
  type: number,
  privateKey: string,
  publicKey: string,
  remotePublicKey: string,
  message: string,
  compress: boolean,
  context: PacketContext
): Promise<Uint8Array> {
  const packet = new Uint8Array(PACKET_BUFFER_SIZE);
  const header = new NHPHeaderEx(packet.buffer);

  // Convert keys from base64 to raw bytes
  const localPrivKeyBytes = base64ToBytes(privateKey);
  const localPubKeyBytes = base64ToBytes(publicKey);
  const remotePubKeyBytes = base64ToBytes(remotePublicKey);
  const msgBytes = stringToBytes(message);

  // Set header fields
  header.version = { major: PROTOCOL_VERSION.MAJOR, minor: PROTOCOL_VERSION.MINOR };
  header.flags = { extended: true, compressed: compress };
  context.counter++;
  header.counter = context.counter;
  const nonce = header.nonce;

  // Initialize chain key and chain hash using SM3 (streaming, matching Go's hash.Hash pattern)
  const chainHasher = newSM3Hash();
  chainHasher.update(stringToBytes(INITIAL_HASH_STRING));  // ChainHash0 state
  const chainHash0 = chainHasher.sum();
  let chainKey = mixKeySM3(chainHash0, stringToBytes(INITIAL_CHAIN_KEY_STRING))[0];

  // HMAC data accumulator (plain SM3, not HMAC-SM3)
  let hmacData = concatBytes(stringToBytes(INITIAL_HASH_STRING), remotePubKeyBytes);

  // Mix in remote public key → ChainHash0 state + remotePubKey
  chainHasher.update(remotePubKeyBytes);

  // Generate ephemeral SM2 key pair and perform ECDH
  const ephemeralKeys = generateSM2KeyPair();
  const ephemeralPublicKeyBytes = ephemeralKeys.publicKey;
  header.ephemeral = ephemeralPublicKeyBytes;

  // ChainHash state += ephemeralPubKey
  chainHasher.update(ephemeralPublicKeyBytes);
  chainKey = mixKeySM3(chainKey, ephemeralPublicKeyBytes)[0];

  // SM2 ECDH: ephemeral private * remote public
  const ess = sm2ECDH(ephemeralKeys.privateKey, remotePubKeyBytes);

  // Encrypt local public key using SM4-GCM (AD = current chainHash snapshot)
  const derivedKeys0 = keyGenSM3_2(chainKey, ess);
  chainKey = derivedKeys0[0];

  const chainHashSnap = chainHasher.sum();

  const keyStatic = sm4GcmSeal(derivedKeys0[1].slice(0, 16), nonce, localPubKeyBytes, chainHashSnap);
  header.static = keyStatic;

  // Evolve chainHash with static ciphertext
  chainHasher.update(keyStatic);

  // SM2 ECDH: local private * remote public
  const ss = sm2ECDH(localPrivKeyBytes, remotePubKeyBytes);

  // Encrypt timestamp
  const derivedKeys1 = keyGenSM3_2(chainKey, ss);
  chainKey = derivedKeys1[0];

  const tsBuf = new ArrayBuffer(8);
  const tsView = new DataView(tsBuf);
  const timestamp = getUnixNano();
  tsView.setBigUint64(0, timestamp);
  const ts = new Uint8Array(tsBuf);
  context.lastSendTimes.set(remotePublicKey, timestamp);

  const tsStatic = sm4GcmSeal(derivedKeys1[1].slice(0, 16), nonce, ts, chainHasher.sum());
  header.timestamp = tsStatic;

  // Encrypt message payload
  const derivedKeys2 = keyGenSM3_2(chainKey, tsStatic);
  chainKey = derivedKeys2[0];
  chainHasher.update(tsStatic);

  let payload = msgBytes;
  if (compress) {
    payload = await zlibCompress(msgBytes);
  }

  const msgStatic = sm4GcmSeal(derivedKeys2[1].slice(0, 16), nonce, payload, chainHasher.sum());
  packet.set(msgStatic, header.size);

  const payloadSize = payload.byteLength + FIELD_SIZES.AEAD_TAG;
  header.typeAndPayloadSize = { type, size: payloadSize };

  // Compute HMAC using plain SM3 hash (matching Go server's hmacHash pattern):
  //   SM3(InitialHashString || serverPubKey || [cookie] || header[0:size-32])
  if (type === NHP_PACKET_TYPES.RNK) {
    const cookie = context.serverCookies.get(remotePublicKey);
    if (cookie) {
      hmacData = concatBytes(hmacData, cookie);
    }
  }
  hmacData = concatBytes(hmacData, packet.subarray(0, header.size - FIELD_SIZES.HMAC));
  header.hmac = sm3(hmacData);

  // Save the final chain key so parseNHPPacketGMSM can continue the Noise chain
  // when decrypting the server's ACK response.
  context.lastBuildChainKeys.set(remotePublicKey, new Uint8Array(chainKey));

  return packet.subarray(0, header.size + payloadSize);
}

/**
 * Parse an NHP packet using GMSM cipher scheme (SM2/SM3/SM4)
 */
async function parseNHPPacketGMSM(
  packet: Uint8Array,
  privateKey: string,
  publicKey: string,
  remotePublicKey: string,
  prevChainKey: Uint8Array | undefined,
  context: PacketContext
): Promise<ParsedPacket> {
  // Create a clean ArrayBuffer copy
  const packetBuffer = new ArrayBuffer(packet.length);
  new Uint8Array(packetBuffer).set(packet);
  const header = new NHPHeaderEx(packetBuffer);
  const { compressed } = header.flags;

  const { type, size } = header.typeAndPayloadSize;

  if (type !== NHP_PACKET_TYPES.ACK && type !== NHP_PACKET_TYPES.COK && type !== NHP_PACKET_TYPES.RAK) {
    throw new Error('Not an ACK, COK, or RAK packet');
  }

  if (packet.length !== header.size + size) {
    throw new Error('Wrong packet size');
  }

  const recvTime = getUnixNano();

  // Convert keys from base64 to raw bytes
  const localPrivKeyBytes = base64ToBytes(privateKey);
  const localPubKeyBytes = base64ToBytes(publicKey);
  const remotePubKeyBytes = base64ToBytes(remotePublicKey);

  // Verify HMAC using plain SM3 hash (matching Go server's hmacHash pattern):
  //   SM3(InitialHashString || localPubKey || header[0:size-32])
  const hmacData = concatBytes(
    stringToBytes(INITIAL_HASH_STRING),
    localPubKeyBytes,
    packet.subarray(0, header.size - FIELD_SIZES.HMAC)
  );
  const checkSum = sm3(hmacData);

  if (!equalBytes(checkSum, header.hmac)) {
    throw new Error('HMAC check failed');
  }

  const ephemeralPublicKeyBytes = header.ephemeral;
  const nonce = header.nonce;
  const keyStatic = header.static;
  const tsStatic = header.timestamp;
  const msgStatic = packet.subarray(header.size);

  // Initialize chain key and chain hash using SM3 (streaming, matching Go's hash.Hash pattern).
  // The NHP Noise protocol is a continuous chain: the server encrypts its ACK
  // starting from the chain key left over after decrypting the agent's KNK.
  // When prevChainKey is provided, resume from it; otherwise start fresh.
  const chainHasher = newSM3Hash();
  chainHasher.update(stringToBytes(INITIAL_HASH_STRING));  // ChainHash0 state
  const chainHash0 = chainHasher.sum();

  let chainKey = prevChainKey
    ? new Uint8Array(prevChainKey)
    : mixKeySM3(chainHash0, stringToBytes(INITIAL_CHAIN_KEY_STRING))[0];

  // Note: for parse (responder), we use localPubKey (= device public key on responder side)
  chainHasher.update(localPubKeyBytes);
  chainHasher.update(ephemeralPublicKeyBytes);
  chainKey = mixKeySM3(chainKey, ephemeralPublicKeyBytes)[0];

  // SM2 ECDH: local private * ephemeral public
  const ess = sm2ECDH(localPrivKeyBytes, ephemeralPublicKeyBytes);

  // Decrypt remote public key (AD = current chainHash snapshot)
  const derivedKeys0 = keyGenSM3_2(chainKey, ess);
  chainKey = derivedKeys0[0];
  const decryptedPubKeyBytes = sm4GcmOpen(derivedKeys0[1].slice(0, 16), nonce, keyStatic, chainHasher.sum());

  if (!equalBytes(remotePubKeyBytes, decryptedPubKeyBytes)) {
    throw new Error('Remote public key check failed');
  }

  chainHasher.update(keyStatic);

  // SM2 ECDH: local private * remote public
  const ss = sm2ECDH(localPrivKeyBytes, remotePubKeyBytes);

  // Decrypt timestamp
  const derivedKeys1 = keyGenSM3_2(chainKey, ss);
  chainKey = derivedKeys1[0];

  const decryptedTs = sm4GcmOpen(derivedKeys1[1].slice(0, 16), nonce, tsStatic, chainHasher.sum());
  const tsBuf = new ArrayBuffer(decryptedTs.length);
  new Uint8Array(tsBuf).set(decryptedTs);
  const tsView = new DataView(tsBuf);
  const remoteSendTime = tsView.getBigUint64(0);

  // Anti-replay checks — only mutate the recorded timestamp after every check
  // passes (see parseNHPPacket above for rationale).
  const lastRemoteSendTime = context.lastRemoteSendTimes.get(remotePublicKey);

  if (lastRemoteSendTime !== undefined) {
    if (remoteSendTime < lastRemoteSendTime) {
      throw new Error('Received replay packet');
    }
    if (remoteSendTime < lastRemoteSendTime + FLOOD_PACKET_THRESHOLD_NS) {
      throw new Error('Received flood packet');
    }
  }

  if (remoteSendTime < recvTime - STALE_PACKET_THRESHOLD_NS) {
    throw new Error('Received stale packet');
  }

  context.lastRemoteSendTimes.set(remotePublicKey, remoteSendTime);

  // Decrypt message
  const derivedKeys2 = keyGenSM3_2(chainKey, header.timestamp);
  chainKey = derivedKeys2[0];
  chainHasher.update(tsStatic);

  let msg = sm4GcmOpen(derivedKeys2[1].slice(0, 16), nonce, msgStatic, chainHasher.sum());

  if (compressed) {
    msg = await zlibDecompress(msg);
  }

  // Handle cookie packets
  if (type === NHP_PACKET_TYPES.COK) {
    context.serverCookies.set(remotePublicKey, msg);
  }

  return {
    type: type as PacketType,
    message: bytesToString(msg),
    remotePublicKey,
  };
}
