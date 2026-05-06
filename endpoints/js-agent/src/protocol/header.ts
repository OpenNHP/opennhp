/**
 * NHP Protocol Header Structures
 * Handles binary packet header parsing and construction
 */

import type { ProtocolVersion, HeaderFlags } from '../types.js';
import {
  HEADER_SIZE,
  HEADER_EX_SIZE,
  HEADER_OFFSETS,
  HEADER_EX_OFFSETS,
  FIELD_SIZES,
  FLAG_BITS,
} from './constants.js';

/**
 * Type and payload size packed in header
 */
export interface TypeAndPayloadSize {
  type: number;
  size: number;
}

/**
 * Base interface for NHP headers
 */
export interface INHPHeader {
  readonly size: number;
  typeAndPayloadSize: TypeAndPayloadSize;
  version: ProtocolVersion;
  flags: HeaderFlags;
  counter: bigint;
  readonly nonce: Uint8Array;
  ephemeral: Uint8Array;
  identity: Uint8Array;
  static: Uint8Array;
  timestamp: Uint8Array;
  hmac: Uint8Array;
  readonly bytes: Uint8Array;
}

/**
 * Standard NHP Header (240 bytes) for X25519 cipher scheme
 */
export class NHPHeader implements INHPHeader {
  public readonly size = HEADER_SIZE;
  private readonly view: DataView;
  public readonly bytes: Uint8Array;

  constructor(buffer: ArrayBuffer, offset = 0) {
    this.view = new DataView(buffer, offset, HEADER_SIZE);
    this.bytes = new Uint8Array(buffer, offset, HEADER_SIZE);
  }

  get typeAndPayloadSize(): TypeAndPayloadSize {
    const preamble = this.view.getUint32(HEADER_OFFSETS.PREAMBLE);
    const masked = this.view.getUint32(HEADER_OFFSETS.TYPE_AND_SIZE);
    const val = (preamble ^ masked) >>> 0;
    return {
      type: ((val & 0xffff0000) >>> 16) >>> 0,
      size: (val & 0x0000ffff) >>> 0,
    };
  }

  set typeAndPayloadSize({ type, size }: TypeAndPayloadSize) {
    const preamble = new Uint32Array(1);
    crypto.getRandomValues(preamble);
    let tns = (((type & 0x0000ffff) << 16) | (size & 0x0000ffff)) >>> 0;
    tns = (preamble[0] ^ tns) >>> 0;
    this.view.setUint32(HEADER_OFFSETS.PREAMBLE, preamble[0]);
    this.view.setUint32(HEADER_OFFSETS.TYPE_AND_SIZE, tns);
  }

  get version(): ProtocolVersion {
    return {
      major: this.view.getUint8(HEADER_OFFSETS.VERSION),
      minor: this.view.getUint8(HEADER_OFFSETS.VERSION + 1),
    };
  }

  set version({ major, minor }: ProtocolVersion) {
    this.view.setUint8(HEADER_OFFSETS.VERSION, major);
    this.view.setUint8(HEADER_OFFSETS.VERSION + 1, minor);
  }

  get flags(): HeaderFlags {
    const flag = this.view.getUint16(HEADER_OFFSETS.FLAGS);
    return {
      extended: Boolean((flag & FLAG_BITS.EXTENDED) >>> 0),
      compressed: Boolean((flag & FLAG_BITS.COMPRESSED) >>> 0),
    };
  }

  set flags({ extended, compressed }: HeaderFlags) {
    let flag = 0;
    if (extended) flag |= FLAG_BITS.EXTENDED;
    if (compressed) flag |= FLAG_BITS.COMPRESSED;
    this.view.setUint16(HEADER_OFFSETS.FLAGS, flag);
  }

  get counter(): bigint {
    return this.view.getBigUint64(HEADER_OFFSETS.COUNTER);
  }

  set counter(v: bigint) {
    this.view.setBigUint64(HEADER_OFFSETS.COUNTER, v);
  }

  get nonce(): Uint8Array {
    const bytes = new Uint8Array(FIELD_SIZES.NONCE);
    // First 4 bytes are zero, last 8 bytes are counter
    bytes.set(this.bytes.subarray(HEADER_OFFSETS.COUNTER, HEADER_OFFSETS.COUNTER + 8), 4);
    return bytes;
  }

  get ephemeral(): Uint8Array {
    return this.bytes.subarray(
      HEADER_OFFSETS.EPHEMERAL,
      HEADER_OFFSETS.EPHEMERAL + FIELD_SIZES.X25519_KEY
    );
  }

  set ephemeral(bytes: Uint8Array) {
    if (bytes.length === FIELD_SIZES.X25519_KEY) {
      this.bytes.set(bytes, HEADER_OFFSETS.EPHEMERAL);
    }
  }

  get identity(): Uint8Array {
    return this.bytes.subarray(
      HEADER_OFFSETS.IDENTITY,
      HEADER_OFFSETS.IDENTITY + FIELD_SIZES.IDENTITY
    );
  }

  set identity(bytes: Uint8Array) {
    if (bytes.length === FIELD_SIZES.IDENTITY) {
      this.bytes.set(bytes, HEADER_OFFSETS.IDENTITY);
    }
  }

  get static(): Uint8Array {
    return this.bytes.subarray(
      HEADER_OFFSETS.STATIC,
      HEADER_OFFSETS.STATIC + FIELD_SIZES.X25519_STATIC
    );
  }

  set static(bytes: Uint8Array) {
    if (bytes.length === FIELD_SIZES.X25519_STATIC) {
      this.bytes.set(bytes, HEADER_OFFSETS.STATIC);
    }
  }

  get timestamp(): Uint8Array {
    return this.bytes.subarray(
      HEADER_OFFSETS.TIMESTAMP,
      HEADER_OFFSETS.TIMESTAMP + FIELD_SIZES.TIMESTAMP
    );
  }

  set timestamp(bytes: Uint8Array) {
    if (bytes.length === FIELD_SIZES.TIMESTAMP) {
      this.bytes.set(bytes, HEADER_OFFSETS.TIMESTAMP);
    }
  }

  get hmac(): Uint8Array {
    return this.bytes.subarray(HEADER_OFFSETS.HMAC, HEADER_OFFSETS.HMAC + FIELD_SIZES.HMAC);
  }

  set hmac(bytes: Uint8Array) {
    if (bytes.length === FIELD_SIZES.HMAC) {
      this.bytes.set(bytes, HEADER_OFFSETS.HMAC);
    }
  }
}

/**
 * Extended NHP Header (304 bytes) for GM SM2 cipher scheme
 */
export class NHPHeaderEx implements INHPHeader {
  public readonly size = HEADER_EX_SIZE;
  private readonly view: DataView;
  public readonly bytes: Uint8Array;

  constructor(buffer: ArrayBuffer, offset = 0) {
    this.view = new DataView(buffer, offset, HEADER_EX_SIZE);
    this.bytes = new Uint8Array(buffer, offset, HEADER_EX_SIZE);
  }

  get typeAndPayloadSize(): TypeAndPayloadSize {
    const preamble = this.view.getUint32(HEADER_EX_OFFSETS.PREAMBLE);
    const masked = this.view.getUint32(HEADER_EX_OFFSETS.TYPE_AND_SIZE);
    const val = (preamble ^ masked) >>> 0;
    return {
      type: ((val & 0xffff0000) >>> 16) >>> 0,
      size: (val & 0x0000ffff) >>> 0,
    };
  }

  set typeAndPayloadSize({ type, size }: TypeAndPayloadSize) {
    const preamble = new Uint32Array(1);
    crypto.getRandomValues(preamble);
    let tns = (((type & 0x0000ffff) << 16) | (size & 0x0000ffff)) >>> 0;
    tns = (preamble[0] ^ tns) >>> 0;
    this.view.setUint32(HEADER_EX_OFFSETS.PREAMBLE, preamble[0]);
    this.view.setUint32(HEADER_EX_OFFSETS.TYPE_AND_SIZE, tns);
  }

  get version(): ProtocolVersion {
    return {
      major: this.view.getUint8(HEADER_EX_OFFSETS.VERSION),
      minor: this.view.getUint8(HEADER_EX_OFFSETS.VERSION + 1),
    };
  }

  set version({ major, minor }: ProtocolVersion) {
    this.view.setUint8(HEADER_EX_OFFSETS.VERSION, major);
    this.view.setUint8(HEADER_EX_OFFSETS.VERSION + 1, minor);
  }

  get flags(): HeaderFlags {
    const flag = this.view.getUint16(HEADER_EX_OFFSETS.FLAGS);
    return {
      extended: Boolean((flag & FLAG_BITS.EXTENDED) >>> 0),
      compressed: Boolean((flag & FLAG_BITS.COMPRESSED) >>> 0),
    };
  }

  set flags({ extended, compressed }: HeaderFlags) {
    let flag = 0;
    if (extended) flag |= FLAG_BITS.EXTENDED;
    if (compressed) flag |= FLAG_BITS.COMPRESSED;
    this.view.setUint16(HEADER_EX_OFFSETS.FLAGS, flag);
  }

  get counter(): bigint {
    return this.view.getBigUint64(HEADER_EX_OFFSETS.COUNTER);
  }

  set counter(v: bigint) {
    this.view.setBigUint64(HEADER_EX_OFFSETS.COUNTER, v);
  }

  get nonce(): Uint8Array {
    const bytes = new Uint8Array(FIELD_SIZES.NONCE);
    bytes.set(this.bytes.subarray(HEADER_EX_OFFSETS.COUNTER, HEADER_EX_OFFSETS.COUNTER + 8), 4);
    return bytes;
  }

  get ephemeral(): Uint8Array {
    return this.bytes.subarray(
      HEADER_EX_OFFSETS.EPHEMERAL,
      HEADER_EX_OFFSETS.EPHEMERAL + FIELD_SIZES.SM2_KEY
    );
  }

  set ephemeral(bytes: Uint8Array) {
    if (bytes.length === FIELD_SIZES.SM2_KEY) {
      this.bytes.set(bytes, HEADER_EX_OFFSETS.EPHEMERAL);
    }
  }

  get identity(): Uint8Array {
    return this.bytes.subarray(
      HEADER_EX_OFFSETS.IDENTITY,
      HEADER_EX_OFFSETS.IDENTITY + FIELD_SIZES.IDENTITY
    );
  }

  set identity(bytes: Uint8Array) {
    if (bytes.length === FIELD_SIZES.IDENTITY) {
      this.bytes.set(bytes, HEADER_EX_OFFSETS.IDENTITY);
    }
  }

  get static(): Uint8Array {
    return this.bytes.subarray(
      HEADER_EX_OFFSETS.STATIC,
      HEADER_EX_OFFSETS.STATIC + FIELD_SIZES.SM2_STATIC
    );
  }

  set static(bytes: Uint8Array) {
    if (bytes.length === FIELD_SIZES.SM2_STATIC) {
      this.bytes.set(bytes, HEADER_EX_OFFSETS.STATIC);
    }
  }

  get timestamp(): Uint8Array {
    return this.bytes.subarray(
      HEADER_EX_OFFSETS.TIMESTAMP,
      HEADER_EX_OFFSETS.TIMESTAMP + FIELD_SIZES.TIMESTAMP
    );
  }

  set timestamp(bytes: Uint8Array) {
    if (bytes.length === FIELD_SIZES.TIMESTAMP) {
      this.bytes.set(bytes, HEADER_EX_OFFSETS.TIMESTAMP);
    }
  }

  get hmac(): Uint8Array {
    return this.bytes.subarray(HEADER_EX_OFFSETS.HMAC, HEADER_EX_OFFSETS.HMAC + FIELD_SIZES.HMAC);
  }

  set hmac(bytes: Uint8Array) {
    if (bytes.length === FIELD_SIZES.HMAC) {
      this.bytes.set(bytes, HEADER_EX_OFFSETS.HMAC);
    }
  }
}
