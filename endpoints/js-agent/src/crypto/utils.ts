/**
 * Utility functions for cryptographic operations
 * Works in both browser and Node.js environments
 */

// Type declarations for Node.js compatibility.  Node's Buffer is a Uint8Array
// subclass, so we model just the bits we need without claiming full coverage.
interface NodeBuffer extends Uint8Array {
  toString(encoding?: string): string;
}
declare const Buffer: {
  from(data: Uint8Array | ArrayBuffer): NodeBuffer;
  from(data: string, encoding?: string): NodeBuffer;
};
declare const process: { versions?: { node?: string } };

// Detect Node.js environment
const isNode = typeof process !== 'undefined' && process.versions?.node;

/**
 * Convert Uint8Array to Base64 string
 * Works in both browser and Node.js
 */
export function bytesToBase64(bytes: Uint8Array): string {
  if (isNode && typeof Buffer !== 'undefined') {
    // Node.js path — Buffer.from(typedarray) shares the underlying memory and
    // gives us the native, fast base64 encoder.
    return Buffer.from(bytes).toString('base64');
  }
  // Browser path
  let binary = '';
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

/**
 * Convert Base64 string to Uint8Array
 * Works in both browser and Node.js
 */
export function base64ToBytes(base64: string): Uint8Array {
  if (isNode && typeof Buffer !== 'undefined') {
    // Node.js path — copy out of the Buffer into a fresh Uint8Array so callers
    // see a plain typed array and don't accidentally share Buffer memory.
    const buf = Buffer.from(base64, 'base64');
    return new Uint8Array(buf);
  }
  // Browser path
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

/**
 * Convert string to UTF-8 bytes
 */
export function stringToBytes(str: string): Uint8Array {
  const encoder = new TextEncoder();
  return encoder.encode(str);
}

/**
 * Convert UTF-8 bytes to string
 */
export function bytesToString(bytes: Uint8Array): string {
  const decoder = new TextDecoder();
  return decoder.decode(bytes);
}

/**
 * Constant-time byte-array equality.
 * Used for HMAC verification on attacker-controlled bytes, so the loop must
 * always touch every byte and accumulate differences without short-circuiting.
 */
export function equalBytes(a: Uint8Array, b: Uint8Array): boolean {
  if (a.length !== b.length) return false;
  let diff = 0;
  for (let i = 0; i < a.length; i++) {
    diff |= a[i] ^ b[i];
  }
  return diff === 0;
}

/**
 * Get current Unix timestamp in nanoseconds
 */
export function getUnixNano(): bigint {
  const ms = Date.now();
  // Use performance.now() for sub-millisecond precision if available
  const subMs = typeof performance !== 'undefined' ? performance.now() % 1 : 0;
  return BigInt(ms) * 1_000_000n + BigInt(Math.floor(subMs * 1_000_000));
}

/**
 * Compress data using zlib deflate
 * Works in both browser (CompressionStream) and Node.js (zlib)
 */
export async function zlibCompress(data: Uint8Array): Promise<Uint8Array> {
  if (typeof CompressionStream !== 'undefined') {
    // Browser path using Compression Streams API
    const cs = new CompressionStream('deflate');
    // Start reading immediately (consumer)
    const readPromise = new Response(cs.readable).arrayBuffer();
    const writer = cs.writable.getWriter();
    // Copy to a new ArrayBuffer to avoid SharedArrayBuffer issues
    const buffer = data.slice(); // copies only byteLength bytes
    await writer.write(buffer);
    await writer.close();

    const compressedBuffer = await readPromise;
    return new Uint8Array(compressedBuffer);
  }

  // Node.js path using dynamic import
  if (isNode) {
    const { deflateSync } = await import('zlib');
    return new Uint8Array(deflateSync(data));
  }

  throw new Error('Compression not supported in this environment');
}

/**
 * Decompress data using zlib inflate
 * Works in both browser (DecompressionStream) and Node.js (zlib)
 */
export async function zlibDecompress(compressedData: Uint8Array): Promise<Uint8Array> {
  if (typeof DecompressionStream !== 'undefined') {
    // Browser path using Compression Streams API
    const ds = new DecompressionStream('deflate');
    // Start reading immediately (consumer)
    const readPromise = new Response(ds.readable).arrayBuffer();
    const writer = ds.writable.getWriter();
    // Copy to a new ArrayBuffer to avoid SharedArrayBuffer issues
    const buffer = compressedData.slice();
    await writer.write(buffer);
    await writer.close();

    const arrayBuffer = await readPromise;
    return new Uint8Array(arrayBuffer);
  }

  // Node.js path using dynamic import
  if (isNode) {
    const { inflateSync } = await import('zlib');
    return new Uint8Array(inflateSync(compressedData));
  }

  throw new Error('Decompression not supported in this environment');
}

/**
 * Generate cryptographically secure random bytes
 */
export function randomBytes(length: number): Uint8Array {
  const bytes = new Uint8Array(length);
  if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
    crypto.getRandomValues(bytes);
  } else if (isNode) {
    // Fallback for older Node.js versions
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const { randomFillSync } = require('crypto');
    randomFillSync(bytes);
  } else {
    throw new Error('No secure random source available');
  }
  return bytes;
}

/**
 * Concatenate multiple Uint8Arrays
 */
export function concatBytes(...arrays: Uint8Array[]): Uint8Array {
  const totalLength = arrays.reduce((sum, arr) => sum + arr.length, 0);
  const result = new Uint8Array(totalLength);
  let offset = 0;
  for (const arr of arrays) {
    result.set(arr, offset);
    offset += arr.length;
  }
  return result;
}

/**
 * Convert Uint8Array to hex string
 */
export function bytesToHex(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('');
}

/**
 * Convert hex string to Uint8Array
 */
export function hexToBytes(hex: string): Uint8Array {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(hex.substr(i * 2, 2), 16);
  }
  return bytes;
}

/**
 * Encode bytes using base64url without padding — the same encoding the relay
 * uses for cluster fingerprints. Kept inline rather than depending on a
 * library because the agent already vendors only what it needs.
 */
function bytesToBase64Url(bytes: Uint8Array): string {
  return bytesToBase64(bytes)
    .replaceAll('+', '-')
    .replaceAll('/', '_')
    .replaceAll('=', '');
}

/**
 * Compute the relay cluster fingerprint for a raw public key.
 *
 * Mirrors nhp/utils.PubKeyFingerprint exactly:
 *   base64url(SHA-256(rawPubKey)[:8])
 *
 * The result is 11 characters with no padding. Use this to derive the
 * `/relay/<clusterId>` path segment from a server's public key without
 * having to ask the relay for its cluster list.
 */
export async function pubKeyFingerprint(rawPubKey: Uint8Array): Promise<string> {
  // Copy the bytes into a fresh ArrayBuffer so the typed array is guaranteed
  // to be a plain Uint8Array<ArrayBuffer>. Without this, callers passing
  // Buffer or a SharedArrayBuffer view trip the BufferSource type guard.
  const input = new Uint8Array(rawPubKey.byteLength);
  input.set(rawPubKey);
  const digest = await crypto.subtle.digest('SHA-256', input);
  return bytesToBase64Url(new Uint8Array(digest, 0, 8));
}

/**
 * Convenience wrapper that decodes a standard-base64-encoded public key
 * (as stored in agent ServerConfig) before fingerprinting.
 */
export async function pubKeyFingerprintFromBase64(pubKeyBase64: string): Promise<string> {
  return pubKeyFingerprint(base64ToBytes(pubKeyBase64));
}
