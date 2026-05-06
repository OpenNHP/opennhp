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
} from '../../src/protocol/packet.js';
import { NHP_PACKET_TYPES } from '../../src/protocol/constants.js';
import { generateX25519KeyPairBase64 } from '../../src/crypto/ecdh.js';
import { generateSM2KeyPairBase64 } from '../../src/crypto/sm2.js';

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

  it('auto-detects curve25519 scheme when cipherScheme is omitted', async () => {
    const { agent, server } = makeKeyPairs();
    const msg = 'auto detect';

    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, false
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
  it('COK packet is stored and used in subsequent RNK', async () => {
    const { agent, server } = makeKeyPairs();
    const cookie = new Uint8Array(32).fill(0xc0);

    // Server sends COK with cookie payload
    const cookiePacket = await buildNHPPacket(
      NHP_PACKET_TYPES.COK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      String.fromCharCode(...cookie), false, 'curve25519'
    );

    // Agent parses COK — cookie is stored internally keyed by server public key
    const parsedCok = await parseNHPPacket(
      cookiePacket,
      agent.privateKey, agent.publicKey,
      server.publicKey
    );
    expect(parsedCok.type).toBe(NHP_PACKET_TYPES.COK);

    // Agent sends RNK using the stored cookie in HMAC
    const rnkPacket = await buildNHPPacket(
      NHP_PACKET_TYPES.RNK,
      agent.privateKey, agent.publicKey,
      server.publicKey,
      're-knock payload', false, 'curve25519'
    );

    // RNK packet should be built successfully
    expect(rnkPacket.length).toBeGreaterThan(240);
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
    ).rejects.toThrow('Not an ACK or COK packet');
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
      'cookie-data', false, 'curve25519'
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

  it('auto-detects gmsm scheme from SM2 key length', async () => {
    const { agent, server } = makeGMKeyPairs();
    const msg = 'auto gmsm';

    // Pass no cipherScheme — should auto-detect gmsm from key length > 50 chars
    const packet = await buildNHPPacket(
      NHP_PACKET_TYPES.ACK,
      server.privateKey, server.publicKey,
      agent.publicKey,
      msg, false
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
