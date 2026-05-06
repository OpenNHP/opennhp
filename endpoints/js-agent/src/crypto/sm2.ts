/**
 * SM2 Elliptic Curve Cryptography
 * Chinese cryptographic standard (GB/T 32918-2016)
 *
 * SM2 uses a specific elliptic curve with parameters similar to secp256r1/P-256.
 * Key sizes:
 *   - Private key: 32 bytes
 *   - Public key: 64 bytes (uncompressed X, Y coordinates)
 *   - Shared secret: 32 bytes
 */

import { sm2 } from 'sm-crypto-v2';
import { bytesToBase64, base64ToBytes, bytesToHex, hexToBytes } from './utils.js';

/** SM2 private key size in bytes */
export const SM2_PRIVATE_KEY_SIZE = 32;

/** SM2 public key size in bytes (uncompressed, without 04 prefix) */
export const SM2_PUBLIC_KEY_SIZE = 64;

/** SM2 key pair as raw bytes */
export interface SM2KeyPairRaw {
  privateKey: Uint8Array;
  publicKey: Uint8Array;
}

/** SM2 key pair as Base64 strings */
export interface SM2KeyPairBase64 {
  privateKey: string;
  publicKey: string;
}

/**
 * Generate a new SM2 key pair
 */
export function generateSM2KeyPair(): SM2KeyPairRaw {
  const keyPair = sm2.generateKeyPairHex();

  // Private key is 64 hex chars = 32 bytes
  const privateKey = hexToBytes(keyPair.privateKey);

  // Public key from sm-crypto-v2 is 128 hex chars (04 prefix + 64 bytes)
  // Strip the 04 prefix if present
  let pubKeyHex = keyPair.publicKey;
  if (pubKeyHex.startsWith('04')) {
    pubKeyHex = pubKeyHex.slice(2);
  }
  const publicKey = hexToBytes(pubKeyHex);

  return { privateKey, publicKey };
}

/**
 * Generate a new SM2 key pair and return as Base64 strings
 */
export function generateSM2KeyPairBase64(): SM2KeyPairBase64 {
  const { privateKey, publicKey } = generateSM2KeyPair();
  return {
    privateKey: bytesToBase64(privateKey),
    publicKey: bytesToBase64(publicKey),
  };
}

/**
 * Derive SM2 public key from private key
 */
export function deriveSM2PublicKey(privateKey: Uint8Array): Uint8Array {
  if (privateKey.length !== SM2_PRIVATE_KEY_SIZE) {
    throw new Error(`SM2 private key must be ${SM2_PRIVATE_KEY_SIZE} bytes`);
  }

  const privateKeyHex = bytesToHex(privateKey);
  let publicKeyHex = sm2.getPublicKeyFromPrivateKey(privateKeyHex);

  // Strip the 04 prefix if present
  if (publicKeyHex.startsWith('04')) {
    publicKeyHex = publicKeyHex.slice(2);
  }

  return hexToBytes(publicKeyHex);
}

/**
 * Derive SM2 public key from Base64-encoded private key
 */
export function deriveSM2PublicKeyFromBase64(privateKeyBase64: string): string {
  const privateKey = base64ToBytes(privateKeyBase64);
  const publicKey = deriveSM2PublicKey(privateKey);
  return bytesToBase64(publicKey);
}

/**
 * Perform SM2 ECDH key exchange
 * @param privateKey - 32-byte private key
 * @param publicKey - 64-byte public key (uncompressed, without 04 prefix)
 * @returns 32-byte shared secret
 */
export function sm2ECDH(privateKey: Uint8Array, publicKey: Uint8Array): Uint8Array {
  if (privateKey.length !== SM2_PRIVATE_KEY_SIZE) {
    throw new Error(`SM2 private key must be ${SM2_PRIVATE_KEY_SIZE} bytes`);
  }
  if (publicKey.length !== SM2_PUBLIC_KEY_SIZE) {
    throw new Error(`SM2 public key must be ${SM2_PUBLIC_KEY_SIZE} bytes`);
  }

  const privateKeyHex = bytesToHex(privateKey);
  // Add 04 prefix for uncompressed point format
  const publicKeyHex = '04' + bytesToHex(publicKey);

  // sm2.ecdh returns a compressed EC point (33 bytes: 02/03 prefix + 32-byte X coordinate).
  // The shared secret is the X coordinate only (matching Go's crypto/ecdh convention).
  const sharedSecret = sm2.ecdh(privateKeyHex, publicKeyHex);

  // Skip the 1-byte compression prefix, return the 32-byte X coordinate
  return sharedSecret.slice(1, 33);
}

/**
 * Check if SM2 is available
 */
export function isSM2Available(): boolean {
  return true;
}
