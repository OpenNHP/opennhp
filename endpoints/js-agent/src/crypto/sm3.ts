/**
 * SM3 Hash Function
 * Chinese cryptographic hash standard (GB/T 32905-2016)
 *
 * SM3 produces a 256-bit (32-byte) hash output.
 */

import { sm3 as sm3Hash } from 'sm-crypto-v2';
import { hexToBytes, concatBytes } from './utils.js';

/** SM3 hash output size in bytes */
export const SM3_HASH_SIZE = 32;

/**
 * SM3 hash state for incremental hashing
 */
export class SM3Hasher {
  private data: Uint8Array[] = [];

  /**
   * Update the hash with additional data
   */
  update(bytes: Uint8Array): void {
    this.data.push(bytes);
  }

  /**
   * Get the final hash
   */
  sum(): Uint8Array {
    const combined = concatBytes(...this.data);
    return sm3(combined);
  }

  /**
   * Create a clone of this hasher
   */
  clone(): SM3Hasher {
    const cloned = new SM3Hasher();
    cloned.data = [...this.data];
    return cloned;
  }
}

/**
 * Create a new SM3 hasher
 */
export function newSM3Hash(): SM3Hasher {
  return new SM3Hasher();
}

/**
 * Compute SM3 hash of data
 * @param data - Input data to hash
 * @returns 32-byte hash output
 */
export function sm3(data: Uint8Array): Uint8Array {
  // sm3Hash returns hex string
  const hashHex = sm3Hash(data);
  return hexToBytes(hashHex);
}

/**
 * Compute HMAC-SM3
 * @param key - HMAC key
 * @param msg - Message to authenticate
 * @returns 32-byte HMAC output
 */
export function hmacSM3(key: Uint8Array, msg: Uint8Array): Uint8Array {
  // sm3Hash with key option computes HMAC-SM3
  const hashHex = sm3Hash(msg, { key, mode: 'hmac' });
  return hexToBytes(hashHex);
}

/**
 * HMAC-SM3 with two messages concatenated (for Noise protocol)
 * @param key - HMAC key
 * @param msg1 - First message
 * @param msg2 - Second message
 * @returns 32-byte HMAC output
 */
export function hmacSM3_2(key: Uint8Array, msg1: Uint8Array, msg2: Uint8Array): Uint8Array {
  const combined = concatBytes(msg1, msg2);
  return hmacSM3(key, combined);
}

/**
 * Key generation function 1 (for Noise protocol GMSM scheme)
 * Returns first 32 bytes of HKDF-like expansion
 */
export function keyGenSM3_1(chainingKey: Uint8Array, inputKeyMaterial: Uint8Array): Uint8Array {
  const tempKey = hmacSM3(chainingKey, inputKeyMaterial);
  const output = hmacSM3(tempKey, new Uint8Array([0x01]));
  return output;
}

/**
 * Key generation function 2 (for Noise protocol GMSM scheme)
 * Returns two 32-byte keys from HKDF-like expansion
 */
export function keyGenSM3_2(
  chainingKey: Uint8Array,
  inputKeyMaterial: Uint8Array
): [Uint8Array, Uint8Array] {
  const tempKey = hmacSM3(chainingKey, inputKeyMaterial);
  const output1 = hmacSM3(tempKey, new Uint8Array([0x01]));
  const output2 = hmacSM3(tempKey, concatBytes(output1, new Uint8Array([0x02])));
  return [output1, output2];
}

/**
 * Mix key operation (for Noise protocol GMSM scheme)
 */
export function mixKeySM3(
  chainingKey: Uint8Array,
  inputKeyMaterial: Uint8Array
): [Uint8Array, Uint8Array] {
  return keyGenSM3_2(chainingKey, inputKeyMaterial);
}

/**
 * Mix hash operation (for Noise protocol GMSM scheme)
 */
export function mixHashSM3(hash: Uint8Array, data: Uint8Array): Uint8Array {
  return sm3(concatBytes(hash, data));
}

/**
 * Check if SM3 is available
 */
export function isSM3Available(): boolean {
  return true;
}
