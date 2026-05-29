/**
 * Tests for the relay-side server fingerprint helper.
 *
 * The fingerprint is also computed in Go (nhp/utils.PubKeyFingerprint); this
 * test pins a known input → output so any divergence between the two
 * implementations is caught here rather than at deployment time.
 */

import { describe, it, expect } from 'vitest';
import {
  pubKeyFingerprint,
  pubKeyFingerprintFromBase64,
  bytesToBase64,
} from '../../src/crypto/utils.js';

describe('pubKeyFingerprint', () => {
  it('is deterministic for the same input', async () => {
    const key = new Uint8Array(32);
    for (let i = 0; i < key.length; i++) key[i] = i + 1;

    const a = await pubKeyFingerprint(key);
    const b = await pubKeyFingerprint(key);
    expect(a).toBe(b);
  });

  it('produces an 11-char base64url string without padding', async () => {
    const key = new Uint8Array(32).fill(0x42);
    const fp = await pubKeyFingerprint(key);
    expect(fp).toHaveLength(11);
    expect(fp).toMatch(/^[A-Za-z0-9_-]{11}$/);
    expect(fp.includes('=')).toBe(false);
  });

  it('returns distinct fingerprints for distinct inputs', async () => {
    const a = new Uint8Array(32).fill(0x01);
    const b = new Uint8Array(32).fill(0x02);
    expect(await pubKeyFingerprint(a)).not.toBe(await pubKeyFingerprint(b));
  });

  it('FromBase64 matches the raw-bytes path', async () => {
    const raw = new Uint8Array(32);
    for (let i = 0; i < raw.length; i++) raw[i] = (i * 7) & 0xff;
    const direct = await pubKeyFingerprint(raw);
    const indirect = await pubKeyFingerprintFromBase64(bytesToBase64(raw));
    expect(indirect).toBe(direct);
  });

  // Cross-language pin: these two outputs were computed once with
  // nhp/utils.PubKeyFingerprint (Go) and hardcoded here. Either side
  // changing its algorithm — different hash, different prefix length,
  // different encoding — flips one of these strings and fails the test
  // before the divergence ships. Pair-tested via go run + node, see
  // commit feat(relay): support multiple nhp-server entries.
  it('matches the Go implementation for known keys', async () => {
    const filledKey = new Uint8Array(32).fill(0x42);
    expect(await pubKeyFingerprint(filledKey)).toBe('Ql7U5KNrMOo');

    const seqKey = new Uint8Array(32);
    for (let i = 0; i < seqKey.length; i++) seqKey[i] = i + 1;
    expect(await pubKeyFingerprint(seqKey)).toBe('riFsLvUkejc');
  });
});
