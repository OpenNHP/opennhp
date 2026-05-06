/**
 * Tests for AES-256-GCM AEAD encryption
 * Validates encryption/decryption roundtrip
 */

import { describe, it, expect } from 'vitest';
import { aesGcmSeal, aesGcmOpen } from '../../src/crypto/aead.js';
import { stringToBytes, bytesToString, bytesToHex } from '../../src/crypto/utils.js';

describe('AES-256-GCM AEAD', () => {
  it('should encrypt and decrypt successfully', () => {
    const key = new Uint8Array(32).fill(0x42);
    const nonce = new Uint8Array(12).fill(0x24);
    const plaintext = stringToBytes('Hello, NHP!');
    const aad = stringToBytes('additional data');

    const ciphertext = aesGcmSeal(key, nonce, plaintext, aad);

    // Ciphertext should be plaintext length + 16 bytes tag
    expect(ciphertext.length).toBe(plaintext.length + 16);

    // Decrypt
    const decrypted = aesGcmOpen(key, nonce, ciphertext, aad);
    expect(bytesToString(decrypted)).toBe('Hello, NHP!');
  });

  it('should produce different ciphertext with different nonces', () => {
    const key = new Uint8Array(32).fill(0x42);
    const nonce1 = new Uint8Array(12).fill(0x01);
    const nonce2 = new Uint8Array(12).fill(0x02);
    const plaintext = stringToBytes('Same message');
    const aad = new Uint8Array(0);

    const ciphertext1 = aesGcmSeal(key, nonce1, plaintext, aad);
    const ciphertext2 = aesGcmSeal(key, nonce2, plaintext, aad);

    expect(bytesToHex(ciphertext1)).not.toBe(bytesToHex(ciphertext2));
  });

  it('should produce different ciphertext with different keys', () => {
    const key1 = new Uint8Array(32).fill(0x01);
    const key2 = new Uint8Array(32).fill(0x02);
    const nonce = new Uint8Array(12).fill(0x42);
    const plaintext = stringToBytes('Same message');
    const aad = new Uint8Array(0);

    const ciphertext1 = aesGcmSeal(key1, nonce, plaintext, aad);
    const ciphertext2 = aesGcmSeal(key2, nonce, plaintext, aad);

    expect(bytesToHex(ciphertext1)).not.toBe(bytesToHex(ciphertext2));
  });

  it('should fail decryption with wrong key', () => {
    const key = new Uint8Array(32).fill(0x42);
    const wrongKey = new Uint8Array(32).fill(0x43);
    const nonce = new Uint8Array(12).fill(0x24);
    const plaintext = stringToBytes('Secret message');
    const aad = new Uint8Array(0);

    const ciphertext = aesGcmSeal(key, nonce, plaintext, aad);

    expect(() => {
      aesGcmOpen(wrongKey, nonce, ciphertext, aad);
    }).toThrow();
  });

  it('should fail decryption with wrong AAD', () => {
    const key = new Uint8Array(32).fill(0x42);
    const nonce = new Uint8Array(12).fill(0x24);
    const plaintext = stringToBytes('Secret message');
    const aad = stringToBytes('correct aad');
    const wrongAad = stringToBytes('wrong aad');

    const ciphertext = aesGcmSeal(key, nonce, plaintext, aad);

    expect(() => {
      aesGcmOpen(key, nonce, ciphertext, wrongAad);
    }).toThrow();
  });

  it('should fail decryption with tampered ciphertext', () => {
    const key = new Uint8Array(32).fill(0x42);
    const nonce = new Uint8Array(12).fill(0x24);
    const plaintext = stringToBytes('Secret message');
    const aad = new Uint8Array(0);

    const ciphertext = aesGcmSeal(key, nonce, plaintext, aad);

    // Tamper with the ciphertext
    const tampered = new Uint8Array(ciphertext);
    tampered[0] ^= 0xff;

    expect(() => {
      aesGcmOpen(key, nonce, tampered, aad);
    }).toThrow();
  });

  it('should handle empty plaintext', () => {
    const key = new Uint8Array(32).fill(0x42);
    const nonce = new Uint8Array(12).fill(0x24);
    const plaintext = new Uint8Array(0);
    const aad = stringToBytes('some aad');

    const ciphertext = aesGcmSeal(key, nonce, plaintext, aad);

    // Should be just the 16-byte tag
    expect(ciphertext.length).toBe(16);

    const decrypted = aesGcmOpen(key, nonce, ciphertext, aad);
    expect(decrypted.length).toBe(0);
  });

  it('should handle empty AAD', () => {
    const key = new Uint8Array(32).fill(0x42);
    const nonce = new Uint8Array(12).fill(0x24);
    const plaintext = stringToBytes('Message without AAD');
    const aad = new Uint8Array(0);

    const ciphertext = aesGcmSeal(key, nonce, plaintext, aad);
    const decrypted = aesGcmOpen(key, nonce, ciphertext, aad);

    expect(bytesToString(decrypted)).toBe('Message without AAD');
  });
});
