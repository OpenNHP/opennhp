/**
 * Noise protocol key derivation and hashing
 * Uses Blake2s (not SHA256) to match Go NHP implementation
 */

import { blake2s } from '@noble/hashes/blake2s';
import { hmac } from '@noble/hashes/hmac';
import { concatBytes } from './utils.js';

/**
 * Blake2s-256 hash state for incremental hashing
 * NHP uses Blake2s for the CURVE25519 cipher scheme
 */
export class Blake2sHasher {
  private data: Uint8Array[] = [];

  /**
   * Update the hash with additional data
   */
  update(bytes: Uint8Array): void {
    this.data.push(bytes);
  }

  /**
   * Get the final hash without consuming the hasher
   * Allows continued updates after getting intermediate hash
   */
  sum(): Uint8Array {
    const combined = concatBytes(...this.data);
    return blake2s(combined);
  }

  /**
   * Create a clone of this hasher for getting intermediate results
   */
  clone(): Blake2sHasher {
    const cloned = new Blake2sHasher();
    cloned.data = [...this.data];
    return cloned;
  }
}

/**
 * Create a new Blake2s-256 hasher
 */
export function newBlake2sHash(): Blake2sHasher {
  return new Blake2sHasher();
}

/**
 * Update Blake2s hasher with data
 */
export function updateBlake2s(hasher: Blake2sHasher, bytes: Uint8Array): void {
  hasher.update(bytes);
}

/**
 * Get the current hash value (non-consuming)
 */
export function sumBlake2s(hasher: Blake2sHasher): Uint8Array {
  return hasher.clone().sum();
}

/**
 * Compute HMAC-Blake2s with a single message
 */
export function hmacBlake2s(key: Uint8Array, msg: Uint8Array): Uint8Array {
  return hmac(blake2s, key, msg);
}

/**
 * Compute HMAC-Blake2s with two concatenated messages
 */
export function hmacBlake2s2(key: Uint8Array, msg0: Uint8Array, msg1: Uint8Array): Uint8Array {
  const combined = concatBytes(msg0, msg1);
  return hmac(blake2s, key, combined);
}

/**
 * Derive a single key from key and message using HKDF-like construction
 * keyGen1(key, msg) = HMAC(HMAC(key, msg), 0x01)
 */
export function keyGen1(key: Uint8Array, msg: Uint8Array): Uint8Array {
  const prk = hmacBlake2s(key, msg);
  const n = new Uint8Array([0x01]);
  return hmacBlake2s(prk, n);
}

/**
 * Derive two keys from key and message using HKDF-like construction
 * Returns { first: T1, second: T2 } where:
 *   prk = HMAC(key, msg)
 *   T1 = HMAC(prk, 0x01)
 *   T2 = HMAC(prk, T1 || 0x02)
 */
export function keyGen2(key: Uint8Array, msg: Uint8Array): { first: Uint8Array; second: Uint8Array } {
  const prk = hmacBlake2s(key, msg);
  const n1 = new Uint8Array([0x01]);
  const n2 = new Uint8Array([0x02]);

  const first = hmacBlake2s(prk, n1);
  const second = hmacBlake2s2(prk, first, n2);

  return { first, second };
}

/**
 * Mix key material into chain key
 * mixKey(key, msg) = keyGen1(key, msg)
 */
export function mixKey(key: Uint8Array, msg: Uint8Array): Uint8Array {
  return keyGen1(key, msg);
}

/**
 * Mix data into chain hash
 * mixHash(hash, msg) = Blake2s(hash || msg)
 */
export function mixHash(hash: Uint8Array, msg: Uint8Array): Uint8Array {
  const combined = concatBytes(hash, msg);
  return blake2s(combined);
}

// Legacy aliases for backwards compatibility during refactoring
export { Blake2sHasher as SHA256Hasher };
export { newBlake2sHash as newSHA256Hash };
export { updateBlake2s as updateSHA256 };
export { sumBlake2s as sumSHA256 };
export { hmacBlake2s as hmac1 };
export { hmacBlake2s2 as hmac2 };
