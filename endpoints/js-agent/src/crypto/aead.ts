/**
 * AEAD encryption using AES-256-GCM
 * Uses @noble/ciphers for cryptographic operations
 * NHP uses AES-256-GCM (not ChaCha20) to match Go implementation
 */

import { gcm } from '@noble/ciphers/aes.js';

/**
 * Encrypt data using AES-256-GCM AEAD
 * @param key - 32-byte encryption key
 * @param nonce - 12-byte nonce
 * @param plaintext - Data to encrypt
 * @param additionalData - Additional authenticated data (AAD)
 * @returns Ciphertext with 16-byte authentication tag appended
 */
export function aesGcmSeal(
  key: Uint8Array,
  nonce: Uint8Array,
  plaintext: Uint8Array,
  additionalData: Uint8Array
): Uint8Array {
  const cipher = gcm(key, nonce, additionalData);
  return cipher.encrypt(plaintext);
}

/**
 * Decrypt data using AES-256-GCM AEAD
 * @param key - 32-byte encryption key
 * @param nonce - 12-byte nonce
 * @param ciphertextWithTag - Ciphertext with authentication tag
 * @param additionalData - Additional authenticated data (AAD)
 * @returns Decrypted plaintext
 * @throws Error if authentication fails
 */
export function aesGcmOpen(
  key: Uint8Array,
  nonce: Uint8Array,
  ciphertextWithTag: Uint8Array,
  additionalData: Uint8Array
): Uint8Array {
  const cipher = gcm(key, nonce, additionalData);
  return cipher.decrypt(ciphertextWithTag);
}

// Legacy aliases for backwards compatibility during refactoring
export { aesGcmSeal as chacha20Seal };
export { aesGcmOpen as chacha20Open };
