/**
 * Cryptographic module exports
 */

// X25519 ECDH (CURVE25519 scheme)
export {
  // Raw byte functions (recommended)
  generateX25519KeyPairRaw,
  generateX25519KeyPairBase64,
  ecdhX25519Raw,
  derivePublicKey,
  derivePublicKeyFromBase64,
  // Web Crypto functions (legacy)
  generateX25519KeyPair,
  ecdhX25519,
  x25519PublicKeyToBytes,
  bytesToX25519PublicKey,
  base64ToX25519PublicKey,
  x25519PrivateKeyToBytes,
  bytesToX25519PrivateKey,
  base64ToX25519PrivateKey,
} from './ecdh.js';

// AES-256-GCM AEAD (CURVE25519 scheme)
export { aesGcmSeal, aesGcmOpen, chacha20Seal, chacha20Open } from './aead.js';

// Blake2s Noise protocol (CURVE25519 scheme)
export {
  Blake2sHasher,
  newBlake2sHash,
  updateBlake2s,
  sumBlake2s,
  hmacBlake2s,
  hmacBlake2s2,
  keyGen1,
  keyGen2,
  mixKey,
  mixHash,
  // Legacy aliases
  SHA256Hasher,
  newSHA256Hash,
  updateSHA256,
  sumSHA256,
  hmac1,
  hmac2,
} from './noise.js';

// GM/SM cryptography (GMSM scheme)
export {
  SM2_PRIVATE_KEY_SIZE,
  SM2_PUBLIC_KEY_SIZE,
  generateSM2KeyPair,
  generateSM2KeyPairBase64,
  deriveSM2PublicKey,
  deriveSM2PublicKeyFromBase64,
  sm2ECDH,
  isSM2Available,
} from './sm2.js';
export type { SM2KeyPairRaw, SM2KeyPairBase64 } from './sm2.js';

export {
  SM3_HASH_SIZE,
  sm3,
  hmacSM3,
  hmacSM3_2,
  SM3Hasher,
  newSM3Hash,
  keyGenSM3_1,
  keyGenSM3_2,
  mixKeySM3,
  mixHashSM3,
  isSM3Available,
} from './sm3.js';

export {
  SM4_KEY_SIZE,
  SM4_BLOCK_SIZE,
  SM4_GCM_NONCE_SIZE,
  SM4_GCM_TAG_SIZE,
  sm4GcmSeal,
  sm4GcmOpen,
  isSM4Available,
} from './sm4.js';

// Utility functions
export {
  bytesToBase64,
  base64ToBytes,
  stringToBytes,
  bytesToString,
  equalBytes,
  getUnixNano,
  zlibCompress,
  zlibDecompress,
  randomBytes,
  concatBytes,
  bytesToHex,
  hexToBytes,
} from './utils.js';
