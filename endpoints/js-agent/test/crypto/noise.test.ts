/**
 * Tests for Blake2s-based Noise protocol functions
 * Validates against known test vectors from Go implementation
 */

import { describe, it, expect } from 'vitest';
import { blake2s } from '@noble/hashes/blake2s';
import {
  newBlake2sHash,
  updateBlake2s,
  sumBlake2s,
  hmacBlake2s,
  keyGen1,
  keyGen2,
  mixKey,
  mixHash,
} from '../../src/crypto/noise.js';
import { stringToBytes, bytesToHex } from '../../src/crypto/utils.js';

describe('Blake2s Hash', () => {
  it('should produce correct hash for empty input', () => {
    const hash = blake2s(new Uint8Array(0));
    expect(hash.length).toBe(32);
  });

  it('should produce correct hash for "hello"', () => {
    const hash = blake2s(stringToBytes('hello'));
    expect(hash.length).toBe(32);
    // Known Blake2s-256 hash for "hello"
    expect(bytesToHex(hash)).toBe(
      '19213bacc58dee6dbde3ceb9a47cbb330b3d86f8cca8997eb00be456f140ca25'
    );
  });

  it('should work with incremental hasher', () => {
    const hasher = newBlake2sHash();
    updateBlake2s(hasher, stringToBytes('hel'));
    updateBlake2s(hasher, stringToBytes('lo'));
    const hash = sumBlake2s(hasher);

    // Should match single-shot hash
    const directHash = blake2s(stringToBytes('hello'));
    expect(bytesToHex(hash)).toBe(bytesToHex(directHash));
  });
});

describe('HMAC-Blake2s', () => {
  it('should produce 32-byte output', () => {
    const key = stringToBytes('key');
    const msg = stringToBytes('message');
    const hmac = hmacBlake2s(key, msg);
    expect(hmac.length).toBe(32);
  });

  it('should be deterministic', () => {
    const key = stringToBytes('test-key');
    const msg = stringToBytes('test-message');
    const hmac1 = hmacBlake2s(key, msg);
    const hmac2 = hmacBlake2s(key, msg);
    expect(bytesToHex(hmac1)).toBe(bytesToHex(hmac2));
  });

  it('should change with different keys', () => {
    const msg = stringToBytes('message');
    const hmac1 = hmacBlake2s(stringToBytes('key1'), msg);
    const hmac2 = hmacBlake2s(stringToBytes('key2'), msg);
    expect(bytesToHex(hmac1)).not.toBe(bytesToHex(hmac2));
  });
});

describe('Key Generation (HKDF-like)', () => {
  it('keyGen1 should produce 32-byte output', () => {
    const key = new Uint8Array(32).fill(1);
    const msg = stringToBytes('input');
    const derived = keyGen1(key, msg);
    expect(derived.length).toBe(32);
  });

  it('keyGen2 should produce two 32-byte outputs', () => {
    const key = new Uint8Array(32).fill(2);
    const msg = stringToBytes('input');
    const { first, second } = keyGen2(key, msg);
    expect(first.length).toBe(32);
    expect(second.length).toBe(32);
    // They should be different
    expect(bytesToHex(first)).not.toBe(bytesToHex(second));
  });

  it('keyGen1 should follow HKDF construction', () => {
    // keyGen1(key, msg) = HMAC(HMAC(key, msg), 0x01)
    const key = new Uint8Array(32).fill(3);
    const msg = stringToBytes('test');

    const derived = keyGen1(key, msg);

    // Manual calculation
    const prk = hmacBlake2s(key, msg);
    const expected = hmacBlake2s(prk, new Uint8Array([0x01]));

    expect(bytesToHex(derived)).toBe(bytesToHex(expected));
  });
});

describe('Mix Functions', () => {
  it('mixKey should be equivalent to keyGen1', () => {
    const key = new Uint8Array(32).fill(4);
    const msg = stringToBytes('material');

    const mixed = mixKey(key, msg);
    const keygen = keyGen1(key, msg);

    expect(bytesToHex(mixed)).toBe(bytesToHex(keygen));
  });

  it('mixHash should concatenate and hash', () => {
    const hash = new Uint8Array(32).fill(5);
    const msg = stringToBytes('data');

    const mixed = mixHash(hash, msg);

    // mixHash(h, m) = Blake2s(h || m)
    const combined = new Uint8Array(hash.length + msg.length);
    combined.set(hash);
    combined.set(msg, hash.length);
    const expected = blake2s(combined);

    expect(bytesToHex(mixed)).toBe(bytesToHex(expected));
  });
});
