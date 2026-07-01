/**
 * Tests for NHPAgent high-level SDK class
 *
 * Transport is mocked via vi.spyOn on the private createTransport method
 * so tests run without any network dependency.
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { NHPAgent } from '../src/NHPAgent.js';
import { buildNHPPacket, resetGlobalCounter } from '../src/protocol/packet.js';
import * as packetModule from '../src/protocol/packet.js';
import { NHP_PACKET_TYPES } from '../src/protocol/constants.js';
import { generateX25519KeyPairBase64 } from '../src/crypto/ecdh.js';
import { generateSM2KeyPairBase64 } from '../src/crypto/sm2.js';
import type { EventHandler } from '../src/types.js';

// ─── Mock transport factory ───────────────────────────────────────────────────

/**
 * Creates a mock transport that immediately delivers a pre-built response packet
 * when send() is called.
 */
function makeMockTransport(responsePacket: Uint8Array) {
  const handlers = new Map<string, Set<EventHandler>>();

  return {
    connect: vi.fn().mockResolvedValue(undefined),
    disconnect: vi.fn(),
    isConnected: vi.fn().mockReturnValue(false),
    send: vi.fn().mockImplementation(() => {
      // Deliver the response asynchronously, simulating network round-trip
      const messageHandlers = handlers.get('message');
      if (messageHandlers) {
        setTimeout(() => {
          messageHandlers.forEach(h => h({ data: responsePacket }));
        }, 0);
      }
    }),
    on: vi.fn().mockImplementation((event: string, handler: EventHandler) => {
      if (!handlers.has(event)) handlers.set(event, new Set());
      handlers.get(event)!.add(handler);
    }),
    off: vi.fn().mockImplementation((event: string, handler: EventHandler) => {
      handlers.get(event)?.delete(handler);
    }),
  };
}

/**
 * Injects a mock transport into a NHPAgent instance by spying on createTransport.
 */
function injectMockTransport(agent: NHPAgent, responsePacket: Uint8Array) {
  const mock = makeMockTransport(responsePacket);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  vi.spyOn(agent as any, 'createTransport').mockReturnValue(mock);
  return mock;
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

async function makeAckPacket(
  serverPrivKey: string,
  serverPubKey: string,
  agentPubKey: string,
  payload: object
) {
  return buildNHPPacket(
    NHP_PACKET_TYPES.ACK,
    serverPrivKey, serverPubKey,
    agentPubKey,
    JSON.stringify(payload),
    true,
    'curve25519'
  );
}

async function makeRakPacket(
  serverPrivKey: string,
  serverPubKey: string,
  agentPubKey: string,
  payload: object
) {
  return buildNHPPacket(
    NHP_PACKET_TYPES.RAK,
    serverPrivKey, serverPubKey,
    agentPubKey,
    JSON.stringify(payload),
    true,
    'curve25519'
  );
}

const SUCCESS_ACK = {
  errCode: '',
  resHost: { 'my-service': '10.0.0.1:8080' },
  opnTime: 300,
  agentAddr: '1.2.3.4:5000',
};

beforeEach(() => {
  resetGlobalCounter();
});

afterEach(() => {
  vi.restoreAllMocks();
});

// ─── Static metadata ──────────────────────────────────────────────────────────

describe('NHPAgent.version', () => {
  it('exposes a non-empty semver string', () => {
    expect(NHPAgent.version).toMatch(/^\d+\.\d+\.\d+/);
  });
});

// ─── Initialization ───────────────────────────────────────────────────────────

describe('NHPAgent initialization', () => {
  it('initializes with default config', async () => {
    const agent = new NHPAgent();
    await agent.init();
    expect(agent.getPublicKey()).toBeTruthy();
    expect(agent.getPublicKey().length).toBe(44); // base64 of 32 bytes
    await agent.close();
  });

  it('init is idempotent — calling twice does not throw', async () => {
    const agent = new NHPAgent();
    await agent.init();
    await expect(agent.init()).resolves.toBeUndefined();
    await agent.close();
  });

  it('generates a different key pair each time', async () => {
    const a = new NHPAgent();
    const b = new NHPAgent();
    await a.init();
    await b.init();
    expect(a.getPublicKey()).not.toBe(b.getPublicKey());
    await a.close();
    await b.close();
  });

  it('uses provided private key and derives correct public key', async () => {
    const { privateKey, publicKey } = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ privateKey });
    await agent.init();
    expect(agent.getPublicKey()).toBe(publicKey);
    await agent.close();
  });

  it('getPublicKey throws before init', () => {
    const agent = new NHPAgent();
    expect(() => agent.getPublicKey()).toThrow('Agent not initialized');
  });

  it('initializes with gmsm cipher scheme', async () => {
    const agent = new NHPAgent({ cipherScheme: 'gmsm' });
    await agent.init();
    // SM2 public key is 64 bytes → base64 = 88 chars
    expect(agent.getPublicKey().length).toBe(88);
    await agent.close();
  });

  it('uses provided SM2 private key', async () => {
    const { privateKey, publicKey } = generateSM2KeyPairBase64();
    const agent = new NHPAgent({ cipherScheme: 'gmsm', privateKey });
    await agent.init();
    expect(agent.getPublicKey()).toBe(publicKey);
    await agent.close();
  });
});

// ─── Server management ────────────────────────────────────────────────────────

describe('Server management', () => {
  it('addServer and removeServer work without throwing', async () => {
    const agent = new NHPAgent();
    await agent.init();

    agent.addServer({ publicKey: 'fakepubkey', host: 'example.com', port: 62206 });
    agent.removeServer('example.com:62206');

    await agent.close();
  });

  it('removeServer with unknown id does not throw', async () => {
    const agent = new NHPAgent();
    await agent.init();
    expect(() => agent.removeServer('nonexistent')).not.toThrow();
    await agent.close();
  });

  it('addServer uses explicit id when provided', async () => {
    const agent = new NHPAgent();
    await agent.init();

    agent.addServer({ id: 'primary', publicKey: 'key', host: 'nhp.example.com', port: 62206 });
    // Should be removable by the explicit id
    expect(() => agent.removeServer('primary')).not.toThrow();

    await agent.close();
  });
});

// ─── Identity ─────────────────────────────────────────────────────────────────

describe('Identity', () => {
  it('setIdentity does not throw', async () => {
    const agent = new NHPAgent();
    await agent.init();
    expect(() => agent.setIdentity({
      userId: 'user@example.com',
      deviceId: 'device-001',
      organizationId: 'example.org',
    })).not.toThrow();
    await agent.close();
  });

  it('setUser (legacy) does not throw', async () => {
    const agent = new NHPAgent();
    await agent.init();
    expect(() => agent.setUser('user@example.com', 'example.org')).not.toThrow();
    await agent.close();
  });
});

// ─── knockResource error cases (no transport needed) ─────────────────────────

describe('knockResource — precondition errors', () => {
  it('returns errorCode 1 when agent not initialized', async () => {
    const agent = new NHPAgent();
    const result = await agent.knockResource({
      resourceId: 'res', serviceId: 'svc',
      serverHost: 'example.com', serverPort: 62206,
    });
    expect(result.success).toBe(false);
    expect(result.errorCode).toBe(1);
  });

  it('returns errorCode 2 when identity not set', async () => {
    const agent = new NHPAgent();
    await agent.init();

    const result = await agent.knockResource({
      resourceId: 'res', serviceId: 'svc',
      serverHost: 'example.com', serverPort: 62206,
    });
    expect(result.success).toBe(false);
    expect(result.errorCode).toBe(2);
    await agent.close();
  });

  it('returns errorCode 3 when server not configured', async () => {
    const agent = new NHPAgent();
    await agent.init();
    agent.setIdentity({ userId: 'u', deviceId: 'd' });

    const result = await agent.knockResource({
      resourceId: 'res', serviceId: 'svc',
      serverHost: 'example.com', serverPort: 62206,
    });
    expect(result.success).toBe(false);
    expect(result.errorCode).toBe(3);
    await agent.close();
  });
});

// ─── knockResource success path (mocked transport) ───────────────────────────

describe('knockResource — success path', () => {
  it('returns success with resource hosts on valid ACK', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    const ackPacket = await makeAckPacket(
      server.privateKey, server.publicKey,
      agent.getPublicKey(),
      SUCCESS_ACK
    );

    agent.setIdentity({ userId: 'user@example.com', deviceId: 'device-001' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });
    injectMockTransport(agent, ackPacket);

    const result = await agent.knockResource({
      resourceId: 'resource-1',
      serviceId: 'my-service',
      serverHost: 'nhp.example.com',
      serverPort: 62206,
    });

    expect(result.success).toBe(true);
    expect(result.resourceHosts).toEqual({ 'my-service': '10.0.0.1:8080' });
    expect(result.agentAddress).toBe('1.2.3.4:5000');
    expect(result.expiresAt).toBeGreaterThan(Date.now());
    await agent.close();
  });

  it('expiresAt is approximately now + opnTime seconds', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    const ackPacket = await makeAckPacket(
      server.privateKey, server.publicKey,
      agent.getPublicKey(),
      { ...SUCCESS_ACK, opnTime: 600 }
    );

    agent.setIdentity({ userId: 'u', deviceId: 'd' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });
    injectMockTransport(agent, ackPacket);

    const before = Date.now();
    const result = await agent.knockResource({
      resourceId: 'r', serviceId: 's',
      serverHost: 'nhp.example.com', serverPort: 62206,
    });

    expect(result.success).toBe(true);
    // expiresAt should be within 1 second of now + 600s
    expect(result.expiresAt!).toBeGreaterThanOrEqual(before + 600_000 - 1000);
    expect(result.expiresAt!).toBeLessThanOrEqual(before + 600_000 + 1000);
    await agent.close();
  });

  it('returns accessToken when provided in ACK', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    const ackPacket = await makeAckPacket(
      server.privateKey, server.publicKey,
      agent.getPublicKey(),
      { ...SUCCESS_ACK, aspToken: 'tok-abc123' }
    );

    agent.setIdentity({ userId: 'u', deviceId: 'd' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });
    injectMockTransport(agent, ackPacket);

    const result = await agent.knockResource({
      resourceId: 'r', serviceId: 's',
      serverHost: 'nhp.example.com', serverPort: 62206,
    });

    expect(result.success).toBe(true);
    expect(result.accessToken).toBe('tok-abc123');
    await agent.close();
  });

  it('returns errorCode 6 when server ACK contains errCode', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    const ackPacket = await makeAckPacket(
      server.privateKey, server.publicKey,
      agent.getPublicKey(),
      { errCode: '403', errMsg: 'Access denied', resHost: {}, opnTime: 0, agentAddr: '' }
    );

    agent.setIdentity({ userId: 'u', deviceId: 'd' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });
    injectMockTransport(agent, ackPacket);

    const result = await agent.knockResource({
      resourceId: 'r', serviceId: 's',
      serverHost: 'nhp.example.com', serverPort: 62206,
    });

    expect(result.success).toBe(false);
    expect(result.error).toBe('Access denied');
    expect(result.errorCode).toBe(403);
    await agent.close();
  });

  it('returns errorCode 5 when transport throws', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    agent.setIdentity({ userId: 'u', deviceId: 'd' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });

    // Mock transport that fails to connect
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.spyOn(agent as any, 'createTransport').mockReturnValue({
      connect: vi.fn().mockRejectedValue(new Error('connection refused')),
      disconnect: vi.fn(),
      isConnected: vi.fn().mockReturnValue(false),
      send: vi.fn(),
      on: vi.fn(),
      off: vi.fn(),
    });

    const result = await agent.knockResource({
      resourceId: 'r', serviceId: 's',
      serverHost: 'nhp.example.com', serverPort: 62206,
    });

    expect(result.success).toBe(false);
    expect(result.errorCode).toBe(5);
    expect(result.error).toBe('connection refused');
    await agent.close();
  });
});

// ─── Knock body authenticates the HeaderType (on-path flip protection) ───────

describe('knock body HeaderType', () => {
  it('KNK knock body carries headerType matching the wire packet type', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    const ackPacket = await makeAckPacket(
      server.privateKey, server.publicKey,
      agent.getPublicKey(), SUCCESS_ACK
    );

    agent.setIdentity({ userId: 'u', deviceId: 'd' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });
    injectMockTransport(agent, ackPacket);

    // Spy after the ACK is built so only the outbound knock body is captured.
    const buildSpy = vi.spyOn(packetModule, 'buildNHPPacket');

    const result = await agent.knockResource({
      resourceId: 'r', serviceId: 's',
      serverHost: 'nhp.example.com', serverPort: 62206,
    });
    expect(result.success).toBe(true);

    // The knock body (5th arg to buildNHPPacket) must serialize headerType ==
    // the wire packet type, using the exact JSON key the Go server unmarshals
    // (AgentKnockMsg.HeaderType, `json:"headerType"`). Otherwise the server
    // rejects the knock as 52010 (legacy/missing) — the demo-breaking bug.
    const knockCall = buildSpy.mock.calls.find(c => c[0] === NHP_PACKET_TYPES.KNK);
    expect(knockCall, 'buildNHPPacket was not called with a KNK packet').toBeTruthy();
    const rawBody = knockCall![4] as string;
    expect(rawBody).toContain('"headerType"');
    expect(JSON.parse(rawBody).headerType).toBe(NHP_PACKET_TYPES.KNK);
    expect(JSON.parse(rawBody).headerType).toBe(1); // wire contract with Go core.NHP_KNK

    await agent.close();
  });

  it('sets headerType per send (KNK then RNK) without mutating the shared knockMsg', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();
    agent.setIdentity({ userId: 'u', deviceId: 'd' });

    // Decouple from crypto + networking: capture the body buildNHPPacket
    // receives and resolve the response immediately, so we exercise the
    // per-send injection for both packet types deterministically.
    const buildSpy = vi
      .spyOn(packetModule, 'buildNHPPacket')
      .mockResolvedValue(new Uint8Array([0]));
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.spyOn(agent as any, 'sendAndWaitForResponse').mockResolvedValue({
      type: NHP_PACKET_TYPES.ACK,
      message: JSON.stringify(SUCCESS_ACK),
    });

    // One shared body object, reused for the initial knock and the cookie-resend.
    const knockMsg = { usrId: 'u', devId: 'd', aspId: 's', resId: 'r' };
    const serverCfg = { id: 's1', publicKey: server.publicKey, host: 'h', port: 1 };
    const resource = { resourceId: 'r', serviceId: 's' };

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (agent as any).sendKnock(NHP_PACKET_TYPES.KNK, knockMsg, serverCfg, resource);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    await (agent as any).sendKnock(NHP_PACKET_TYPES.RNK, knockMsg, serverCfg, resource);

    const bodyFor = (pt: number) => {
      const call = buildSpy.mock.calls.find(c => c[0] === pt);
      expect(call, `buildNHPPacket not called with packet type ${pt}`).toBeTruthy();
      return JSON.parse(call![4] as string);
    };
    expect(bodyFor(NHP_PACKET_TYPES.KNK).headerType).toBe(1); // NHP_KNK
    expect(bodyFor(NHP_PACKET_TYPES.RNK).headerType).toBe(8); // NHP_RKN

    // The shared object must never be mutated, or the cookie-resend RNK would
    // carry the KNK type and the server would reject the body/wire mismatch.
    expect(Object.prototype.hasOwnProperty.call(knockMsg, 'headerType')).toBe(false);

    await agent.close();
  });
});

// ─── Event emitter ────────────────────────────────────────────────────────────

describe('Event emitter', () => {
  it('on/off register and unregister handlers without throwing', async () => {
    const agent = new NHPAgent();
    await agent.init();

    const handler = vi.fn();
    agent.on('knock', handler);
    agent.off('knock', handler);

    await agent.close();
  });

  it('emits knock event when knockResource sends a packet', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    const ackPacket = await makeAckPacket(
      server.privateKey, server.publicKey,
      agent.getPublicKey(), SUCCESS_ACK
    );

    agent.setIdentity({ userId: 'u', deviceId: 'd' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });
    injectMockTransport(agent, ackPacket);

    const knockHandler = vi.fn();
    agent.on('knock', knockHandler);

    await agent.knockResource({
      resourceId: 'r', serviceId: 's',
      serverHost: 'nhp.example.com', serverPort: 62206,
    });

    expect(knockHandler).toHaveBeenCalledOnce();
    await agent.close();
  });

  it('emits ack event on successful knock', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    const ackPacket = await makeAckPacket(
      server.privateKey, server.publicKey,
      agent.getPublicKey(), SUCCESS_ACK
    );

    agent.setIdentity({ userId: 'u', deviceId: 'd' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });
    injectMockTransport(agent, ackPacket);

    const ackHandler = vi.fn();
    agent.on('ack', ackHandler);

    await agent.knockResource({
      resourceId: 'r', serviceId: 's',
      serverHost: 'nhp.example.com', serverPort: 62206,
    });

    expect(ackHandler).toHaveBeenCalledOnce();
    await agent.close();
  });

  it('emits error event on transport failure', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    agent.setIdentity({ userId: 'u', deviceId: 'd' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.spyOn(agent as any, 'createTransport').mockReturnValue({
      connect: vi.fn().mockRejectedValue(new Error('timeout')),
      disconnect: vi.fn(),
      isConnected: vi.fn().mockReturnValue(false),
      send: vi.fn(),
      on: vi.fn(),
      off: vi.fn(),
    });

    const errorHandler = vi.fn();
    agent.on('error', errorHandler);

    await agent.knockResource({
      resourceId: 'r', serviceId: 's',
      serverHost: 'nhp.example.com', serverPort: 62206,
    });

    expect(errorHandler).toHaveBeenCalledOnce();
    await agent.close();
  });
});

// ─── exitResource ─────────────────────────────────────────────────────────────

describe('exitResource', () => {
  it('does not throw for known server', async () => {
    const agent = new NHPAgent();
    await agent.init();
    agent.addServer({ publicKey: 'key', host: 'example.com', port: 62206 });

    await expect(agent.exitResource({
      resourceId: 'r', serviceId: 's',
      serverHost: 'example.com', serverPort: 62206,
    })).resolves.toBeUndefined();

    await agent.close();
  });

  it('does not throw for unknown server', async () => {
    const agent = new NHPAgent();
    await agent.init();

    await expect(agent.exitResource({
      resourceId: 'r', serviceId: 's',
      serverHost: 'unknown.example.com', serverPort: 62206,
    })).resolves.toBeUndefined();

    await agent.close();
  });
});

// ─── close ────────────────────────────────────────────────────────────────────

describe('close', () => {
  it('close is idempotent', async () => {
    const agent = new NHPAgent();
    await agent.init();
    await agent.close();
    await expect(agent.close()).resolves.toBeUndefined();
  });

  it('getPublicKey still works after close (key pair is retained)', async () => {
    const agent = new NHPAgent();
    await agent.init();
    const pubKey = agent.getPublicKey();
    await agent.close();
    // close() clears transports and servers but retains the key pair
    expect(agent.getPublicKey()).toBe(pubKey);
  });
});

describe('registerPublicKey — expiresAt passthrough', () => {
  // ServerRegisterAckMsg carries an optional expiresAt (unix-seconds).
  // The SDK must surface it on RegisterResult converted to milliseconds.
  it('returns expiresAt in milliseconds when server provides it', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    const expiresAtSec = 1_900_000_000; // some future-ish unix-seconds
    const rakPacket = await makeRakPacket(
      server.privateKey, server.publicKey,
      agent.getPublicKey(),
      { errCode: '', aspId: 'example', expiresAt: expiresAtSec }
    );

    agent.setIdentity({ userId: 'user@example.com', deviceId: 'device-001' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });
    injectMockTransport(agent, rakPacket);

    const result = await agent.registerPublicKey('example', '123456');

    expect(result.success).toBe(true);
    expect(result.expiresAt).toBe(expiresAtSec * 1000);
    await agent.close();
  });

  it('omits expiresAt when server does not return one', async () => {
    const server = generateX25519KeyPairBase64();
    const agent = new NHPAgent({ transport: 'websocket' });
    await agent.init();

    const rakPacket = await makeRakPacket(
      server.privateKey, server.publicKey,
      agent.getPublicKey(),
      { errCode: '', aspId: 'example' }
    );

    agent.setIdentity({ userId: 'user@example.com', deviceId: 'device-001' });
    agent.addServer({ publicKey: server.publicKey, host: 'nhp.example.com', port: 62206 });
    injectMockTransport(agent, rakPacket);

    const result = await agent.registerPublicKey('example', '123456');

    expect(result.success).toBe(true);
    expect(result.expiresAt).toBeUndefined();
    await agent.close();
  });
});
