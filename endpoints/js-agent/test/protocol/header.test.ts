/**
 * Tests for NHP protocol header structures
 * Validates field read/write for NHPHeader (240 bytes) and NHPHeaderEx (304 bytes)
 */

import { describe, it, expect } from 'vitest';
import { NHPHeader, NHPHeaderEx } from '../../src/protocol/header.js';
import {
  HEADER_SIZE,
  HEADER_EX_SIZE,
  PROTOCOL_VERSION,
} from '../../src/protocol/constants.js';

// ─── NHPHeader (curve25519, 240 bytes) ───────────────────────────────────────

describe('NHPHeader (curve25519, 240 bytes)', () => {
  function makeHeader() {
    const buf = new ArrayBuffer(HEADER_SIZE);
    return new NHPHeader(buf);
  }

  it('size should be 240', () => {
    expect(makeHeader().size).toBe(240);
  });

  it('version round-trips correctly', () => {
    const h = makeHeader();
    h.version = { major: PROTOCOL_VERSION.MAJOR, minor: PROTOCOL_VERSION.MINOR };
    expect(h.version).toEqual({ major: 1, minor: 0 });
  });

  it('flags: extended=false, compressed=false', () => {
    const h = makeHeader();
    h.flags = { extended: false, compressed: false };
    expect(h.flags).toEqual({ extended: false, compressed: false });
  });

  it('flags: compressed=true sets bit correctly', () => {
    const h = makeHeader();
    h.flags = { extended: false, compressed: true };
    expect(h.flags.compressed).toBe(true);
    expect(h.flags.extended).toBe(false);
  });

  it('flags: extended=true sets bit correctly', () => {
    const h = makeHeader();
    h.flags = { extended: true, compressed: false };
    expect(h.flags.extended).toBe(true);
    expect(h.flags.compressed).toBe(false);
  });

  it('counter round-trips correctly', () => {
    const h = makeHeader();
    h.counter = 42n;
    expect(h.counter).toBe(42n);
  });

  it('counter handles large values', () => {
    const h = makeHeader();
    const big = 0xdeadbeefcafebaben;
    h.counter = big;
    expect(h.counter).toBe(big);
  });

  it('nonce is 12 bytes derived from counter', () => {
    const h = makeHeader();
    h.counter = 1n;
    const nonce = h.nonce;
    expect(nonce.length).toBe(12);
    // First 4 bytes should be zero
    expect(nonce[0]).toBe(0);
    expect(nonce[1]).toBe(0);
    expect(nonce[2]).toBe(0);
    expect(nonce[3]).toBe(0);
    // Last 8 bytes encode the counter (big-endian), counter=1 → last byte = 1
    expect(nonce[11]).toBe(1);
  });

  it('ephemeral key round-trips (32 bytes)', () => {
    const h = makeHeader();
    const key = new Uint8Array(32).fill(0xab);
    h.ephemeral = key;
    expect(h.ephemeral).toEqual(key);
  });

  it('ephemeral setter ignores wrong-length input', () => {
    const h = makeHeader();
    const before = new Uint8Array(h.ephemeral);
    h.ephemeral = new Uint8Array(16); // wrong length
    expect(h.ephemeral).toEqual(before);
  });

  it('static field round-trips (48 bytes)', () => {
    const h = makeHeader();
    const val = new Uint8Array(48).fill(0xcd);
    h.static = val;
    expect(h.static).toEqual(val);
  });

  it('timestamp field round-trips (24 bytes)', () => {
    const h = makeHeader();
    const val = new Uint8Array(24).fill(0x12);
    h.timestamp = val;
    expect(h.timestamp).toEqual(val);
  });

  it('hmac field round-trips (32 bytes)', () => {
    const h = makeHeader();
    const val = new Uint8Array(32).fill(0xff);
    h.hmac = val;
    expect(h.hmac).toEqual(val);
  });

  it('typeAndPayloadSize encodes type and size with random preamble', () => {
    const h = makeHeader();
    h.typeAndPayloadSize = { type: 1, size: 100 };
    const { type, size } = h.typeAndPayloadSize;
    expect(type).toBe(1);
    expect(size).toBe(100);
  });

  it('typeAndPayloadSize produces different bytes on each write (random preamble)', () => {
    const buf = new ArrayBuffer(HEADER_SIZE);
    const h = new NHPHeader(buf);

    h.typeAndPayloadSize = { type: 2, size: 50 };
    const bytes1 = new Uint8Array(buf.slice(0, 8));

    h.typeAndPayloadSize = { type: 2, size: 50 };
    const bytes2 = new Uint8Array(buf.slice(0, 8));

    // Preamble is random so raw bytes should differ (with overwhelming probability)
    expect(bytes1).not.toEqual(bytes2);
    // But decoded values must be identical
    expect(h.typeAndPayloadSize).toEqual({ type: 2, size: 50 });
  });

  it('fields do not overlap (write one, others unchanged)', () => {
    const h = makeHeader();
    h.hmac = new Uint8Array(32).fill(0xaa);
    h.ephemeral = new Uint8Array(32).fill(0xbb);
    // hmac should still be 0xaa
    expect(Array.from(h.hmac).every(b => b === 0xaa)).toBe(true);
  });
});

// ─── NHPHeaderEx (gmsm, 304 bytes) ───────────────────────────────────────────

describe('NHPHeaderEx (gmsm, 304 bytes)', () => {
  function makeHeaderEx() {
    const buf = new ArrayBuffer(HEADER_EX_SIZE);
    return new NHPHeaderEx(buf);
  }

  it('size should be 304', () => {
    expect(makeHeaderEx().size).toBe(304);
  });

  it('version round-trips correctly', () => {
    const h = makeHeaderEx();
    h.version = { major: 1, minor: 0 };
    expect(h.version).toEqual({ major: 1, minor: 0 });
  });

  it('flags: extended=true for GM packets', () => {
    const h = makeHeaderEx();
    h.flags = { extended: true, compressed: true };
    expect(h.flags).toEqual({ extended: true, compressed: true });
  });

  it('counter round-trips correctly', () => {
    const h = makeHeaderEx();
    h.counter = 999n;
    expect(h.counter).toBe(999n);
  });

  it('nonce is 12 bytes derived from counter', () => {
    const h = makeHeaderEx();
    h.counter = 2n;
    const nonce = h.nonce;
    expect(nonce.length).toBe(12);
    expect(nonce[11]).toBe(2);
  });

  it('ephemeral key round-trips (64 bytes for SM2)', () => {
    const h = makeHeaderEx();
    const key = new Uint8Array(64).fill(0x77);
    h.ephemeral = key;
    expect(h.ephemeral).toEqual(key);
  });

  it('static field round-trips (80 bytes for SM2)', () => {
    const h = makeHeaderEx();
    const val = new Uint8Array(80).fill(0x33);
    h.static = val;
    expect(h.static).toEqual(val);
  });

  it('hmac field round-trips (32 bytes)', () => {
    const h = makeHeaderEx();
    const val = new Uint8Array(32).fill(0x55);
    h.hmac = val;
    expect(h.hmac).toEqual(val);
  });

  it('typeAndPayloadSize round-trips', () => {
    const h = makeHeaderEx();
    h.typeAndPayloadSize = { type: 7, size: 200 };
    expect(h.typeAndPayloadSize).toEqual({ type: 7, size: 200 });
  });
});
