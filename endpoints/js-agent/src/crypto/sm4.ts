/**
 * SM4 Block Cipher with GCM mode
 * Chinese cryptographic standard (GB/T 32907-2016)
 *
 * SM4 is a 128-bit block cipher with 128-bit key.
 * For NHP, it's used in GCM mode for authenticated encryption.
 */

import { sm4 } from 'sm-crypto-v2';

/** SM4 key size in bytes */
export const SM4_KEY_SIZE = 16;

/** SM4 block size in bytes */
export const SM4_BLOCK_SIZE = 16;

/** GCM nonce size in bytes */
export const SM4_GCM_NONCE_SIZE = 12;

/** GCM authentication tag size in bytes */
export const SM4_GCM_TAG_SIZE = 16;

/**
 * Encrypt data using SM4-GCM AEAD
 * @param key - 16-byte encryption key
 * @param nonce - 12-byte nonce
 * @param plaintext - Data to encrypt
 * @param additionalData - Additional authenticated data (AAD)
 * @returns Ciphertext with 16-byte authentication tag appended
 */
export function sm4GcmSeal(
  key: Uint8Array,
  nonce: Uint8Array,
  plaintext: Uint8Array,
  additionalData: Uint8Array
): Uint8Array {
  if (key.length !== SM4_KEY_SIZE) {
    throw new Error(`SM4 key must be ${SM4_KEY_SIZE} bytes`);
  }
  if (nonce.length !== SM4_GCM_NONCE_SIZE) {
    throw new Error(`SM4-GCM nonce must be ${SM4_GCM_NONCE_SIZE} bytes`);
  }

  const result = sm4.encrypt(plaintext, key, {
    mode: 'gcm',
    iv: nonce,
    associatedData: additionalData,
    output: 'array',
    outputTag: true,
  }) as { output: Uint8Array; tag: Uint8Array };

  // Concatenate ciphertext and tag
  const ciphertext = result.output;
  const tag = result.tag!;

  const output = new Uint8Array(ciphertext.length + tag.length);
  output.set(ciphertext, 0);
  output.set(tag, ciphertext.length);

  return output;
}

/**
 * Decrypt data using SM4-GCM AEAD
 * @param key - 16-byte encryption key
 * @param nonce - 12-byte nonce
 * @param ciphertextWithTag - Ciphertext with authentication tag
 * @param additionalData - Additional authenticated data (AAD)
 * @returns Decrypted plaintext
 * @throws Error if authentication fails
 */
export function sm4GcmOpen(
  key: Uint8Array,
  nonce: Uint8Array,
  ciphertextWithTag: Uint8Array,
  additionalData: Uint8Array
): Uint8Array {
  if (key.length !== SM4_KEY_SIZE) {
    throw new Error(`SM4 key must be ${SM4_KEY_SIZE} bytes`);
  }
  if (nonce.length !== SM4_GCM_NONCE_SIZE) {
    throw new Error(`SM4-GCM nonce must be ${SM4_GCM_NONCE_SIZE} bytes`);
  }
  if (ciphertextWithTag.length < SM4_GCM_TAG_SIZE) {
    throw new Error('Ciphertext too short');
  }

  // Split ciphertext and tag
  const ciphertext = ciphertextWithTag.slice(0, -SM4_GCM_TAG_SIZE);
  const tag = ciphertextWithTag.slice(-SM4_GCM_TAG_SIZE);

  const plaintext = sm4.decrypt(ciphertext, key, {
    mode: 'gcm',
    iv: nonce,
    associatedData: additionalData,
    tag: tag,
    output: 'array',
  });

  return plaintext;
}

/**
 * Check if SM4 is available
 */
export function isSM4Available(): boolean {
  return true;
}
