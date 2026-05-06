/**
 * ECDH key exchange using X25519
 * Uses @noble/curves for raw byte operations (NHP protocol)
 * Also supports Web Crypto API for backwards compatibility
 */

import { x25519 } from '@noble/curves/ed25519.js';
import type { KeyPair, KeyPairBase64 } from '../types.js';
import { bytesToBase64, base64ToBytes } from './utils.js';

/**
 * Generate a new X25519 key pair as raw bytes
 */
export function generateX25519KeyPairRaw(): { privateKey: Uint8Array; publicKey: Uint8Array } {
  const privateKey = x25519.utils.randomSecretKey();
  const publicKey = x25519.getPublicKey(privateKey);
  return { privateKey, publicKey };
}

/**
 * Generate a new X25519 key pair and return as Base64 strings
 */
export function generateX25519KeyPairBase64(): KeyPairBase64 {
  const { privateKey, publicKey } = generateX25519KeyPairRaw();
  return {
    privateKey: bytesToBase64(privateKey),
    publicKey: bytesToBase64(publicKey),
  };
}

/**
 * Perform ECDH key exchange using X25519
 * Returns the raw 32-byte shared secret
 */
export function ecdhX25519Raw(privateKey: Uint8Array, publicKey: Uint8Array): Uint8Array {
  return x25519.getSharedSecret(privateKey, publicKey);
}

/**
 * Derive public key from private key
 */
export function derivePublicKey(privateKey: Uint8Array): Uint8Array {
  return x25519.getPublicKey(privateKey);
}

/**
 * Derive public key from base64-encoded private key
 */
export function derivePublicKeyFromBase64(privateKeyBase64: string): string {
  const privateKey = base64ToBytes(privateKeyBase64);
  const publicKey = derivePublicKey(privateKey);
  return bytesToBase64(publicKey);
}

// ============================================================================
// Web Crypto API functions (for backwards compatibility)
// These are kept for existing code that uses CryptoKey objects
// ============================================================================

/**
 * Generate a new X25519 key pair using Web Crypto
 */
export async function generateX25519KeyPair(): Promise<KeyPair> {
  const keyPair = await crypto.subtle.generateKey(
    { name: 'X25519' },
    true,
    ['deriveBits', 'deriveKey']
  ) as CryptoKeyPair;
  return keyPair;
}

/**
 * Perform ECDH key exchange using Web Crypto
 * Returns a CryptoKey (for legacy code)
 */
export async function ecdhX25519(
  privateKey: CryptoKey,
  remotePublicKey: CryptoKey
): Promise<CryptoKey> {
  return await crypto.subtle.deriveKey(
    {
      name: 'X25519',
      public: remotePublicKey,
    },
    privateKey,
    {
      name: 'AES-GCM',
      length: 256,
    },
    true,
    ['encrypt', 'decrypt']
  );
}

/**
 * Export X25519 public key to raw bytes
 */
export async function x25519PublicKeyToBytes(key: CryptoKey): Promise<Uint8Array> {
  const keyBuffer = await crypto.subtle.exportKey('raw', key);
  return new Uint8Array(keyBuffer);
}

/**
 * Import raw bytes as X25519 public key
 */
export async function bytesToX25519PublicKey(bytes: Uint8Array): Promise<CryptoKey> {
  // Create a clean ArrayBuffer copy to avoid SharedArrayBuffer issues
  const buffer = new ArrayBuffer(bytes.length);
  new Uint8Array(buffer).set(bytes);
  return await crypto.subtle.importKey(
    'raw',
    buffer,
    { name: 'X25519' },
    true,
    []
  );
}

/**
 * Import Base64 string as X25519 public key
 */
export async function base64ToX25519PublicKey(b64: string): Promise<CryptoKey> {
  const bytes = base64ToBytes(b64);
  return await bytesToX25519PublicKey(bytes);
}

/**
 * Export X25519 private key to raw bytes
 */
export async function x25519PrivateKeyToBytes(key: CryptoKey): Promise<Uint8Array> {
  const pkcs8Key = await crypto.subtle.exportKey('pkcs8', key);
  return extractX25519FromPKCS8(pkcs8Key);
}

/**
 * Import raw bytes as X25519 private key
 */
export async function bytesToX25519PrivateKey(bytes: Uint8Array): Promise<CryptoKey> {
  if (bytes.length !== 32) {
    throw new Error('X25519 private key must be exactly 32 bytes');
  }

  const pkcs8 = wrapX25519InPKCS8(bytes);

  // Create a clean ArrayBuffer copy to avoid SharedArrayBuffer issues
  const buffer = new ArrayBuffer(pkcs8.length);
  new Uint8Array(buffer).set(pkcs8);

  return await crypto.subtle.importKey(
    'pkcs8',
    buffer,
    { name: 'X25519' },
    true,
    ['deriveBits', 'deriveKey']
  );
}

/**
 * Import Base64 string as X25519 private key
 */
export async function base64ToX25519PrivateKey(b64: string): Promise<CryptoKey> {
  const bytes = base64ToBytes(b64);
  return await bytesToX25519PrivateKey(bytes);
}

// ============================================================================
// ASN.1/DER helper functions for PKCS#8 format
// ============================================================================

function decodeASN1Length(bytes: Uint8Array, offset: number): { length: number; nextOffset: number } {
  let len = bytes[offset++];
  if (len < 0x80) {
    return { length: len, nextOffset: offset };
  }

  const numBytes = len & 0x7f;
  len = 0;
  for (let i = 0; i < numBytes; i++) {
    len = (len << 8) | bytes[offset++];
  }
  return { length: len, nextOffset: offset };
}

function readASN1(bytes: Uint8Array, offset: number): { tag: number; offset: number; length: number; end: number } {
  const tag = bytes[offset++];
  const { length, nextOffset } = decodeASN1Length(bytes, offset);
  const end = nextOffset + length;
  return { tag, offset: nextOffset, length, end };
}

/**
 * Extract raw 32-byte X25519 private key from PKCS#8 format
 */
function extractX25519FromPKCS8(pkcs8: ArrayBuffer): Uint8Array {
  const bytes = new Uint8Array(pkcs8);

  // Root SEQUENCE
  const root = readASN1(bytes, 0);
  let p = root.offset;

  // INTEGER (version)
  const ver = readASN1(bytes, p);
  p = ver.end;

  // AlgorithmIdentifier SEQUENCE
  const alg = readASN1(bytes, p);
  p = alg.end;

  // privateKey OCTET STRING
  const pkOuter = readASN1(bytes, p);

  // parse inner OCTET STRING inside privateKey
  const inner = readASN1(bytes, pkOuter.offset);

  // raw private key = inner OCTET STRING content
  return bytes.slice(inner.offset, inner.offset + inner.length);
}

function encodeLength(len: number): number[] {
  if (len < 0x80) return [len];
  const bytes: number[] = [];
  let remaining = len;
  while (remaining > 0) {
    bytes.unshift(remaining & 0xff);
    remaining >>= 8;
  }
  return [0x80 | bytes.length, ...bytes];
}

function derEncodeInteger(n: number): number[] {
  return [0x02, 0x01, n];
}

function derEncodeOID(oid: string): number[] {
  const parts = oid.split('.').map(Number);
  const first = 40 * parts[0] + parts[1];
  const rest = parts.slice(2).flatMap(n => {
    const bytes: number[] = [];
    let val = n;
    do {
      bytes.unshift(val & 0x7f);
      val >>= 7;
    } while (val > 0);
    for (let i = 0; i < bytes.length - 1; i++) {
      bytes[i] |= 0x80;
    }
    return bytes;
  });
  return [0x06, rest.length + 1, first, ...rest];
}

function derEncodeOctetString(bytes: Uint8Array | number[]): number[] {
  const arr = bytes instanceof Uint8Array ? Array.from(bytes) : bytes;
  return [0x04, ...encodeLength(arr.length), ...arr];
}

function derEncodeSequence(bytes: number[]): number[] {
  return [0x30, ...encodeLength(bytes.length), ...bytes];
}

/**
 * Wrap raw 32-byte X25519 private key in PKCS#8 format
 */
function wrapX25519InPKCS8(rawKey: Uint8Array): Uint8Array {
  // AlgorithmIdentifier SEQUENCE with X25519 OID (1.3.101.110)
  const algOid = derEncodeOID('1.3.101.110');
  const algIdSeq = derEncodeSequence(algOid);

  // privateKey OCTET STRING (wrapped)
  const privOctet = derEncodeOctetString(rawKey);
  const privAttrOctet = derEncodeOctetString(privOctet);

  // version INTEGER (0)
  const version = derEncodeInteger(0);

  // Full PrivateKeyInfo SEQUENCE
  const seq = derEncodeSequence([...version, ...algIdSeq, ...privAttrOctet]);

  return new Uint8Array(seq);
}
