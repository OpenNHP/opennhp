import { describe, it, expect } from 'vitest';
import {
  generateSM2KeyPair,
  generateSM2KeyPairBase64,
  deriveSM2PublicKey,
  sm2ECDH,
  SM2_PRIVATE_KEY_SIZE,
  SM2_PUBLIC_KEY_SIZE,
  isSM2Available,
} from '../../src/crypto/sm2.js';
import {
  sm3,
  hmacSM3,
  SM3Hasher,
  SM3_HASH_SIZE,
  isSM3Available,
} from '../../src/crypto/sm3.js';
import {
  sm4GcmSeal,
  sm4GcmOpen,
  SM4_KEY_SIZE,
  SM4_GCM_NONCE_SIZE,
  SM4_GCM_TAG_SIZE,
  isSM4Available,
} from '../../src/crypto/sm4.js';
import { randomBytes } from '../../src/crypto/utils.js';

describe('SM2 ECDH', () => {
  it('should report SM2 as available', () => {
    expect(isSM2Available()).toBe(true);
  });

  it('should generate key pair with correct sizes', () => {
    const keyPair = generateSM2KeyPair();
    expect(keyPair.privateKey.length).toBe(SM2_PRIVATE_KEY_SIZE);
    expect(keyPair.publicKey.length).toBe(SM2_PUBLIC_KEY_SIZE);
  });

  it('should generate key pair as Base64', () => {
    const keyPair = generateSM2KeyPairBase64();
    expect(typeof keyPair.privateKey).toBe('string');
    expect(typeof keyPair.publicKey).toBe('string');
  });

  it('should derive public key from private key', () => {
    const keyPair = generateSM2KeyPair();
    const derivedPubKey = deriveSM2PublicKey(keyPair.privateKey);
    expect(derivedPubKey).toEqual(keyPair.publicKey);
  });

  it('should perform ECDH key exchange', () => {
    const alice = generateSM2KeyPair();
    const bob = generateSM2KeyPair();

    const aliceShared = sm2ECDH(alice.privateKey, bob.publicKey);
    const bobShared = sm2ECDH(bob.privateKey, alice.publicKey);

    expect(aliceShared.length).toBe(32);
    expect(bobShared.length).toBe(32);
    expect(aliceShared).toEqual(bobShared);
  });

  it('should reject invalid key sizes', () => {
    expect(() => sm2ECDH(new Uint8Array(31), new Uint8Array(64))).toThrow();
    expect(() => sm2ECDH(new Uint8Array(32), new Uint8Array(63))).toThrow();
  });
});

describe('SM3 Hash', () => {
  it('should report SM3 as available', () => {
    expect(isSM3Available()).toBe(true);
  });

  it('should produce 32-byte hash', () => {
    const hash = sm3(new Uint8Array([1, 2, 3]));
    expect(hash.length).toBe(SM3_HASH_SIZE);
  });

  it('should produce consistent hash for same input', () => {
    const data = new Uint8Array([0x61, 0x62, 0x63]); // "abc"
    const hash1 = sm3(data);
    const hash2 = sm3(data);
    expect(hash1).toEqual(hash2);
  });

  it('should work with SM3Hasher incremental API', () => {
    const data = new Uint8Array([1, 2, 3, 4, 5]);

    const directHash = sm3(data);

    const hasher = new SM3Hasher();
    hasher.update(new Uint8Array([1, 2]));
    hasher.update(new Uint8Array([3, 4, 5]));
    const incrementalHash = hasher.sum();

    expect(incrementalHash).toEqual(directHash);
  });

  it('should compute HMAC-SM3', () => {
    const key = randomBytes(32);
    const msg = new Uint8Array([1, 2, 3]);

    const hmac = hmacSM3(key, msg);
    expect(hmac.length).toBe(SM3_HASH_SIZE);
  });
});

describe('SM4-GCM', () => {
  it('should report SM4 as available', () => {
    expect(isSM4Available()).toBe(true);
  });

  it('should encrypt and decrypt with correct sizes', () => {
    const key = randomBytes(SM4_KEY_SIZE);
    const nonce = randomBytes(SM4_GCM_NONCE_SIZE);
    const plaintext = new Uint8Array([1, 2, 3, 4, 5]);
    const aad = new Uint8Array([]);

    const ciphertext = sm4GcmSeal(key, nonce, plaintext, aad);
    expect(ciphertext.length).toBe(plaintext.length + SM4_GCM_TAG_SIZE);

    const decrypted = sm4GcmOpen(key, nonce, ciphertext, aad);
    expect(decrypted).toEqual(plaintext);
  });

  it('should handle AAD correctly', () => {
    const key = randomBytes(SM4_KEY_SIZE);
    const nonce = randomBytes(SM4_GCM_NONCE_SIZE);
    const plaintext = new Uint8Array([1, 2, 3, 4, 5]);
    const aad = new Uint8Array([0xAA, 0xBB, 0xCC]);

    const ciphertext = sm4GcmSeal(key, nonce, plaintext, aad);
    const decrypted = sm4GcmOpen(key, nonce, ciphertext, aad);
    expect(decrypted).toEqual(plaintext);
  });

  it('should fail with wrong key', () => {
    const key = randomBytes(SM4_KEY_SIZE);
    const wrongKey = randomBytes(SM4_KEY_SIZE);
    const nonce = randomBytes(SM4_GCM_NONCE_SIZE);
    const plaintext = new Uint8Array([1, 2, 3, 4, 5]);
    const aad = new Uint8Array([]);

    const ciphertext = sm4GcmSeal(key, nonce, plaintext, aad);
    expect(() => sm4GcmOpen(wrongKey, nonce, ciphertext, aad)).toThrow();
  });

  it('should fail with wrong AAD', () => {
    const key = randomBytes(SM4_KEY_SIZE);
    const nonce = randomBytes(SM4_GCM_NONCE_SIZE);
    const plaintext = new Uint8Array([1, 2, 3, 4, 5]);
    const aad = new Uint8Array([0xAA]);
    const wrongAad = new Uint8Array([0xBB]);

    const ciphertext = sm4GcmSeal(key, nonce, plaintext, aad);
    expect(() => sm4GcmOpen(key, nonce, ciphertext, wrongAad)).toThrow();
  });

  it('should reject invalid key size', () => {
    const badKey = new Uint8Array(15);
    const nonce = randomBytes(SM4_GCM_NONCE_SIZE);
    const plaintext = new Uint8Array([1, 2, 3]);
    const aad = new Uint8Array([]);

    expect(() => sm4GcmSeal(badKey, nonce, plaintext, aad)).toThrow();
  });

  it('should reject invalid nonce size', () => {
    const key = randomBytes(SM4_KEY_SIZE);
    const badNonce = new Uint8Array(11);
    const plaintext = new Uint8Array([1, 2, 3]);
    const aad = new Uint8Array([]);

    expect(() => sm4GcmSeal(key, badNonce, plaintext, aad)).toThrow();
  });
});
