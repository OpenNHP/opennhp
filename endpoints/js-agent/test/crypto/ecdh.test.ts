/**
 * Tests for X25519 ECDH key exchange
 * Validates key generation and shared secret derivation
 */

import { describe, it, expect } from 'vitest';
import {
  generateX25519KeyPairRaw,
  generateX25519KeyPairBase64,
  ecdhX25519Raw,
  derivePublicKey,
  derivePublicKeyFromBase64,
} from '../../src/crypto/ecdh.js';
import { base64ToBytes, bytesToHex } from '../../src/crypto/utils.js';

describe('X25519 Key Generation', () => {
  it('generateX25519KeyPairRaw should produce valid key pair', () => {
    const { privateKey, publicKey } = generateX25519KeyPairRaw();

    expect(privateKey.length).toBe(32);
    expect(publicKey.length).toBe(32);
  });

  it('generateX25519KeyPairBase64 should produce valid base64 strings', () => {
    const { privateKey, publicKey } = generateX25519KeyPairBase64();

    // Base64 of 32 bytes = 44 characters (with padding)
    expect(privateKey.length).toBe(44);
    expect(publicKey.length).toBe(44);

    // Should decode to 32 bytes
    expect(base64ToBytes(privateKey).length).toBe(32);
    expect(base64ToBytes(publicKey).length).toBe(32);
  });

  it('should generate different key pairs each time', () => {
    const pair1 = generateX25519KeyPairRaw();
    const pair2 = generateX25519KeyPairRaw();

    expect(bytesToHex(pair1.privateKey)).not.toBe(bytesToHex(pair2.privateKey));
    expect(bytesToHex(pair1.publicKey)).not.toBe(bytesToHex(pair2.publicKey));
  });
});

describe('Public Key Derivation', () => {
  it('derivePublicKey should produce correct public key', () => {
    const { privateKey, publicKey } = generateX25519KeyPairRaw();
    const derivedPublic = derivePublicKey(privateKey);

    expect(bytesToHex(derivedPublic)).toBe(bytesToHex(publicKey));
  });

  it('derivePublicKeyFromBase64 should work with Go test vector', () => {
    // Test vector from Go NHP implementation
    const privateKeyB64 = 'kgvvQaBGfHNWCbZMkFWS1K07BgRXlnOo7CHTZF1bsmI=';
    const expectedPublicKeyB64 = 'c0HALYy3433SqJmfN0JpRk1Q6H7xh84MAg89jYtRrQM=';

    const derivedPublicB64 = derivePublicKeyFromBase64(privateKeyB64);

    expect(derivedPublicB64).toBe(expectedPublicKeyB64);
  });
});

describe('ECDH Key Exchange', () => {
  it('should produce same shared secret from both sides', () => {
    const alice = generateX25519KeyPairRaw();
    const bob = generateX25519KeyPairRaw();

    // Alice computes shared secret using her private key and Bob's public key
    const aliceShared = ecdhX25519Raw(alice.privateKey, bob.publicKey);

    // Bob computes shared secret using his private key and Alice's public key
    const bobShared = ecdhX25519Raw(bob.privateKey, alice.publicKey);

    // Both should get the same shared secret
    expect(bytesToHex(aliceShared)).toBe(bytesToHex(bobShared));
    expect(aliceShared.length).toBe(32);
  });

  it('should produce different shared secrets with different key pairs', () => {
    const alice = generateX25519KeyPairRaw();
    const bob = generateX25519KeyPairRaw();
    const charlie = generateX25519KeyPairRaw();

    const aliceBob = ecdhX25519Raw(alice.privateKey, bob.publicKey);
    const aliceCharlie = ecdhX25519Raw(alice.privateKey, charlie.publicKey);

    expect(bytesToHex(aliceBob)).not.toBe(bytesToHex(aliceCharlie));
  });

  it('should work with Go test vector private key', () => {
    // Using the Go test private key
    const privateKeyB64 = 'kgvvQaBGfHNWCbZMkFWS1K07BgRXlnOo7CHTZF1bsmI=';
    const privateKey = base64ToBytes(privateKeyB64);
    const publicKey = derivePublicKey(privateKey);

    // Create a test peer
    const peer = generateX25519KeyPairRaw();

    // ECDH should work
    const shared = ecdhX25519Raw(privateKey, peer.publicKey);
    expect(shared.length).toBe(32);

    // Peer should get same shared secret
    const peerShared = ecdhX25519Raw(peer.privateKey, publicKey);
    expect(bytesToHex(shared)).toBe(bytesToHex(peerShared));
  });
});
