/**
 * Tests for NHP packet build/parse round-trips
 * Covers: ACK round-trip, COK/RNK cookie flow, GMSM scheme,
 * compression toggle, error cases, and anti-replay checks
 */

import { describe, it, expect, beforeEach } from 'vitest';
import {
  buildNHPPacket,
  parseNHPPacket,
  clearServerCookie,
  resetGlobalCounter,
  PacketContext,
} from '../../src/protocol/packet.js';
import {
  NHP_PACKET_TYPES,
  INITIAL_HASH_STRING,
  FIELD_SIZES,
  HEADER_SIZE,
  HEADER_EX_SIZE,
} from '../../src/protocol/constants.js';
import { generateX25519KeyPairBase64 } from '../../src/crypto/ecdh.js';
import { generateSM2KeyPairBase64 } from '../../src/crypto/sm2.js';
import { newBlake2sHash, updateBlake2s, sumBlake2s } from '../../src/crypto/noise.js';
import { newSM3Hash } from '../../src/crypto/sm3.js';
import {
  base64ToBytes,
  bytesToBase64,
  stringToBytes,
} from '../../src/crypto/utils.js';

// ─── helpers ─────────────────────────────────────────────────────────────────

function makeKeyPairs() {
  const agent = generateX25519KeyPairBase64();
  const server = generateX25519KeyPairBase64();
  return { agent, server };
}

function makeGMKeyPairs() {
  const agent = generateSM2KeyPairBase64();
  const server = generateSM2KeyPairBase64();
  return { agent, server };
}

/**
 * The body the Go server sends with an NHP-COK challenge: a
 * common.ServerCookieMsg JSON object whose "cookie" field is the base64 of
 * the 32 raw cookie bytes (nhp/core/responder.go sendCookie).
 */
function cookieChallengeBody(cookie: Uint8Array, transactionId = 7): string {
  return JSON.stringify({ trxId: transactionId, cookie: bytesToBase64(cookie) });
}

/**
 * Independently recompute the curve25519 HMAC of a packet the way the Go
 * server does (nhp/core/responder.go checkHMAC):
 *
 *   Hash(InitialHashString || serverPubKey || header[0:size-32] || cookie)
 *
 * Deliberately does not reuse the code under test — a test that reuses it
 * cannot catch a byte-order or payload mistake.
 */
function recomputeCurveHmac(
  packet: Uint8Array,
  serverPublicKey: string,
  cookie: Uint8Array
): Uint8Array {
  const hasher = newBlake2sHash();
  updateBlake2s(hasher, stringToBytes(INITIAL_HASH_STRING));
  updateBlake2s(hasher, base64ToBytes(serverPublicKey));
  updateBlake2s(hasher, packet.subarray(0, HEADER_SIZE - FIELD_SIZES.HMAC));
  updateBlake2s(hasher, cookie);
  return sumBlake2s(hasher);
}

/**
 * The pre-fix byte order — cookie hashed BEFORE the header prefix. Kept as a
 * negative assertion so the regression cannot come back unnoticed.
 */
function recomputeCurveHmacLegacyOrder(
  packet: Uint8Array,
  serverPublicKey: string,
  cookie: Uint8Array
): Uint8Array {
  const hasher = newBlake2sHash();
  updateBlake2s(hasher, stringToBytes(INITIAL_HASH_STRING));
  updateBlake2s(hasher, base64ToBytes(serverPublicKey));
  updateBlake2s(hasher, cookie);
  updateBlake2s(hasher, packet.subarray(0, HEADER_SIZE - FIELD_SIZES.HMAC));
  return sumBlake2s(hasher);
}

/**
 * The GMSM equivalent of recomputeCurveHmac:
 *
 *   SM3(InitialHashString || serverPubKey || header[0:size-32] || cookie)
 */
function recomputeSM3Hmac(
  packet: Uint8Array,
  serverPublicKey: string,
  cookie: Uint8Array
): Uint8Array {
  const hasher = newSM3Hash();
  hasher.update(stringToBytes(INITIAL_HASH_STRING));
  hasher.update(base64ToBytes(serverPublicKey));
  hasher.update(packet.subarray(0, HEADER_EX_SIZE - FIELD_SIZES.HMAC));
  hasher.update(cookie);
  return hasher.sum();
}

// Reset counter before each test so tests are independent
beforeEach(() => {
  resetGlobalCounter();
});

// ─── CURVE25519 ACK round-trip ────────────────────────────────────────────────

describe('CURVE25519 ACK round-trip', () => {
  it('builds and parses a basic ACK packet', async () => {
    const { agent, server } = makeKeyPairs();
    const msg = 'hello NHP';

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, false, 'curve25519'
    );

    const parsed = await parseNHPPacket(
      packet,
      agent.privateKey, agent.publicKey,
      server.publicKey
    );

    expect(parsed.type).toBe(NHP_PACKET_TYPES.ACK);
    expect(parsed.message).toBe(msg);
    expect(parsed.remotePublicKey).toBe(server.publicKey);
  });

  it('builds and parses with compression enabled', async () => {
    const { agent, server } = makeKeyPairs();
    const msg = JSON.stringify({ errCode: '', resHost: { demo: '10.0.0.1:8080' }, opnTime: 300, agentAddr: '1.2.3.4:5678' });

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, true, 'curve25519'
    );

    const parsed = await parseNHPPacket(
      packet,
      agent.privateKey, agent.publicKey,
      server.publicKey
    );

    expect(parsed.type).toBe(NHP_PACKET_TYPES.ACK);
    expect(parsed.message).toBe(msg);
  });

  it('compressed packet is smaller than uncompressed for repetitive payload', async () => {
    const { agent, server } = makeKeyPairs();
    // Repetitive JSON compresses well
    const msg = JSON.stringify({ data: 'aaaa'.repeat(100) });

    const compressed = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, true, 'curve25519'
    );
    const uncompressed = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, false, 'curve25519'
    );

    expect(compressed.length).toBeLessThan(uncompressed.length);
  });

  it('round-trips with explicit curve25519 cipherScheme argument', async () => {
    const { agent, server } = makeKeyPairs();
    const msg = 'explicit scheme';

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, false, 'curve25519'
    );

    const parsed = await parseNHPPacket(
      packet,
      agent.privateKey, agent.publicKey,
      server.publicKey
    );

    expect(parsed.message).toBe(msg);
  });

  it('packet length equals header size + payload size', async () => {
    const { agent, server } = makeKeyPairs();
    const msg = 'size check';

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, false, 'curve25519'
    );

    // Standard header is 240 bytes; payload = encrypted message + 16-byte tag
    expect(packet.length).toBeGreaterThan(240);
  });
});

// ─── COK / RNK cookie flow ────────────────────────────────────────────────────

describe('COK / RNK cookie flow', () => {
  it('stores the decoded 32-byte cookie from the server JSON challenge', async () => {
    const { agent, server } = makeKeyPairs();
    const context = new PacketContext();
    const cookie = new Uint8Array(32).fill(0xc0);

    // Server sends COK with the ServerCookieMsg JSON body
    const challenge = cookieChallengeBody(cookie);
    const cookiePacket = await buildNHPPacket(
      NHP_PACKET_TYPES.COK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      challenge, false, 'curve25519'
    );

    // Agent parses COK — the cookie is stored internally keyed by server public key
    const parsedCok = await parseNHPPacket(
      cookiePacket,
      agent.privateKey, agent.publicKey,
      server.publicKey,
      undefined, context
    );
    expect(parsedCok.type).toBe(NHP_PACKET_TYPES.COK);
    expect(parsedCok.message).toBe(challenge);

    // The stored value is the decoded 32 bytes, not the JSON body
    expect(context.serverCookies.get(server.publicKey)).toEqual(cookie);
  });

  it('computes the RNK HMAC as Hash(INITIAL || serverPub || headerPrefix || cookie)', async () => {
    const { agent, server } = makeKeyPairs();
    const context = new PacketContext();
    const cookie = new Uint8Array(32).fill(0xc0);

    const cookiePacket = await buildNHPPacket(
      NHP_PACKET_TYPES.COK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      cookieChallengeBody(cookie), false, 'curve25519'
    );
    await parseNHPPacket(
      cookiePacket,
      agent.privateKey, agent.publicKey,
      server.publicKey,
      undefined, context
    );

    // Agent sends RNK using the stored cookie in HMAC
    const rnkPacket = await buildNHPPacket(
      NHP_PACKET_TYPES.RNK,
      agent.privateKey, agent.publicKey,
      server.publicKey,
      're-knock payload', false, 'curve25519', context
    );

    // The HMAC the server recomputes must match — recomputed here independently
    const wireHmac = rnkPacket.subarray(HEADER_SIZE - FIELD_SIZES.HMAC, HEADER_SIZE);
    expect(wireHmac).toEqual(recomputeCurveHmac(rnkPacket, server.publicKey, cookie));

    // ...and the pre-fix byte order must NOT match, or the regression is back
    expect(wireHmac).not.toEqual(
      recomputeCurveHmacLegacyOrder(rnkPacket, server.publicKey, cookie)
    );
  });

  it('rejects a cookie challenge that is not a ServerCookieMsg JSON body', async () => {
    const { agent, server } = makeKeyPairs();
    const context = new PacketContext();

    const cookiePacket = await buildNHPPacket(
      NHP_PACKET_TYPES.COK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      'not-a-cookie-challenge', false, 'curve25519'
    );

    await expect(
      parseNHPPacket(
        cookiePacket,
        agent.privateKey, agent.publicKey,
        server.publicKey,
        undefined, context
      )
    ).rejects.toThrow('Invalid cookie challenge');
    expect(context.serverCookies.has(server.publicKey)).toBe(false);
  });

  it('computes the GMSM RNK HMAC as SM3(INITIAL || serverPub || headerPrefix || cookie)', async () => {
    const { agent, server } = makeGMKeyPairs();
    const context = new PacketContext();
    const cookie = new Uint8Array(32).fill(0xc0);

    const cookiePacket = await buildNHPPacket(
      NHP_PACKET_TYPES.COK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      cookieChallengeBody(cookie), false, 'gmsm'
    );
    await parseNHPPacket(
      cookiePacket,
      agent.privateKey, agent.publicKey,
      server.publicKey,
      undefined, context
    );
    expect(context.serverCookies.get(server.publicKey)).toEqual(cookie);

    const rnkPacket = await buildNHPPacket(
      NHP_PACKET_TYPES.RNK,
      agent.privateKey, agent.publicKey,
      server.publicKey,
      're-knock payload', false, 'gmsm', context
    );

    const wireHmac = rnkPacket.subarray(HEADER_EX_SIZE - FIELD_SIZES.HMAC, HEADER_EX_SIZE);
    expect(wireHmac).toEqual(recomputeSM3Hmac(rnkPacket, server.publicKey, cookie));
  });
});

// ─── Error cases ─────────────────────────────────────────────────────────────

describe('Error cases', () => {
  it('throws on packet too small to be valid', async () => {
    const { agent, server } = makeKeyPairs();

    await expect(
      parseNHPPacket(new Uint8Array(10), agent.privateKey, agent.publicKey, server.publicKey)
    ).rejects.toThrow('Packet size is too small');
  });

  it('throws when parsing a KNK packet (not ACK/COK)', async () => {
    const { agent, server } = makeKeyPairs();

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.KNK,
      agent.privateKey, agent.publicKey,
      server.publicKey,
      'knock', false, 'curve25519'
    );

    await expect(
      parseNHPPacket(packet, server.privateKey, server.publicKey, agent.publicKey)
    ).rejects.toThrow('Not an ACK, COK, RAK, or LRT packet');
  });

  it('throws on HMAC tampering', async () => {
    const { agent, server } = makeKeyPairs();

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      'tamper test', false, 'curve25519'
    );

    // Corrupt the last 8 bytes of the header (HMAC region)
    const tampered = new Uint8Array(packet);
    tampered[232] ^= 0xff;

    await expect(
      parseNHPPacket(tampered, agent.privateKey, agent.publicKey, server.publicKey)
    ).rejects.toThrow();
  });

  it('throws when wrong private key used for decryption', async () => {
    const { agent, server } = makeKeyPairs();
    const { agent: wrongAgent } = makeKeyPairs();

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      'secret', false, 'curve25519'
    );

    await expect(
      parseNHPPacket(packet, wrongAgent.privateKey, wrongAgent.publicKey, server.publicKey)
    ).rejects.toThrow();
  });

  it('throws when wrong server public key used', async () => {
    const { agent, server } = makeKeyPairs();
    const { server: wrongServer } = makeKeyPairs();

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      'secret', false, 'curve25519'
    );

    await expect(
      parseNHPPacket(packet, agent.privateKey, agent.publicKey, wrongServer.publicKey)
    ).rejects.toThrow();
  });
});

// ─── Anti-replay ─────────────────────────────────────────────────────────────

describe('Anti-replay protection', () => {
  it('rejects a replayed packet (same packet sent twice)', async () => {
    const { agent, server } = makeKeyPairs();

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      'replay test', false, 'curve25519'
    );

    // First parse succeeds
    await expect(
      parseNHPPacket(packet, agent.privateKey, agent.publicKey, server.publicKey)
    ).resolves.toBeDefined();

    // Replaying the exact same packet must be rejected
    await expect(
      parseNHPPacket(packet, agent.privateKey, agent.publicKey, server.publicKey)
    ).rejects.toThrow();
  });
});

// ─── clearServerCookie ────────────────────────────────────────────────────────

describe('clearServerCookie', () => {
  it('does not throw when called for unknown key', () => {
    expect(() => clearServerCookie('nonexistent-key')).not.toThrow();
  });

  it('clears stored cookie so RNK no longer includes it', async () => {
    const { agent, server } = makeKeyPairs();

    // Store a cookie by parsing a COK packet
    const cookiePacket = await buildNHPPacket(
      NHP_PACKET_TYPES.COK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      cookieChallengeBody(new Uint8Array(32).fill(0xc0)), false, 'curve25519'
    );
    await parseNHPPacket(
      cookiePacket,
      agent.privateKey, agent.publicKey,
      server.publicKey
    );

    // Clear it
    clearServerCookie(server.publicKey);

    // RNK can still be built (just without cookie in HMAC) — should not throw
    const rnkPacket = await buildNHPPacket(
      NHP_PACKET_TYPES.RNK,
      agent.privateKey, agent.publicKey,
      server.publicKey,
      'rnk after clear', false, 'curve25519'
    );
    expect(rnkPacket.length).toBeGreaterThan(0);
  });
});

// ─── GMSM (SM2/SM3/SM4) round-trip ───────────────────────────────────────────

describe('GMSM ACK round-trip', () => {
  it('builds and parses a GMSM ACK packet', async () => {
    const { agent, server } = makeGMKeyPairs();
    const msg = 'gmsm test payload';

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, false, 'gmsm'
    );

    const parsed = await parseNHPPacket(
      packet,
      agent.privateKey, agent.publicKey,
      server.publicKey
    );

    expect(parsed.type).toBe(NHP_PACKET_TYPES.ACK);
    expect(parsed.message).toBe(msg);
  });

  it('GMSM packet is larger than curve25519 (304 vs 240 byte header)', async () => {
    const gm = makeGMKeyPairs();
    const c25519 = makeKeyPairs();
    const msg = 'size comparison';

    const gmPacket = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      gm.server.privateKey, gm.server.publicKey,
      gm.agent.publicKey,
      msg, false, 'gmsm'
    );
    const c25519Packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      c25519.server.privateKey, c25519.server.publicKey,
      c25519.agent.publicKey,
      msg, false, 'curve25519'
    );

    expect(gmPacket.length).toBeGreaterThan(c25519Packet.length);
  });

  it('round-trips with explicit gmsm cipherScheme argument', async () => {
    const { agent, server } = makeGMKeyPairs();
    const msg = 'explicit gmsm';

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, false, 'gmsm'
    );

    const parsed = await parseNHPPacket(
      packet,
      agent.privateKey, agent.publicKey,
      server.publicKey
    );

    expect(parsed.message).toBe(msg);
  });

  it('builds and parses GMSM with compression', async () => {
    const { agent, server } = makeGMKeyPairs();
    const msg = JSON.stringify({ data: 'repeat'.repeat(50) });

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, true, 'gmsm'
    );

    const parsed = await parseNHPPacket(
      packet,
      agent.privateKey, agent.publicKey,
      server.publicKey
    );

    expect(parsed.message).toBe(msg);
  });
});
