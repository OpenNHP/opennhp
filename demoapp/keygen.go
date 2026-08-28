package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/emmansun/gmsm/sm2"
)

// KeyPair is a freshly generated NHP key pair. The private key is base64-
// encoded in the format expected by the browser-side js-agent; the public
// key likewise. We store the private key encrypted (see sealKey) but pass
// the public key around in cleartext.
type KeyPair struct {
	PrivateKey string // base64-encoded raw private key bytes
	PublicKey  string // base64-encoded raw public key bytes
}

// GenerateKeyPair creates a new key pair for the given cipher scheme. The
// result is ready to be sent to the browser (privateKey) and persisted
// (publicKey, plus the wrapped privateKey).
func GenerateKeyPair(scheme CipherScheme) (*KeyPair, error) {
	switch scheme {
	case CipherSchemeCurve25519:
		return generateX25519KeyPair()
	case CipherSchemeGMSM:
		return generateSM2KeyPair()
	default:
		return nil, fmt.Errorf("unsupported cipher scheme: %s", scheme)
	}
}

// generateX25519KeyPair uses Go's stdlib crypto/ecdh (X25519) and emits
// raw 32-byte keys in base64 — exactly what js-agent's derivePublicKey
// and crypto/ecdh modules expect on the browser side.
func generateX25519KeyPair() (*KeyPair, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("x25519 keygen: %w", err)
	}
	return &KeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(priv.Bytes()),
		PublicKey:  base64.StdEncoding.EncodeToString(priv.PublicKey().Bytes()),
	}, nil
}

// generateSM2KeyPair produces a SM2 key pair using github.com/emmansun/gmsm
// and emits raw 32-byte private scalar + 64-byte (X||Y, no 04 prefix) public
// key in base64 — the format js-agent's sm2.ts expects.
func generateSM2KeyPair() (*KeyPair, error) {
	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("sm2 keygen: %w", err)
	}
	pub := priv.Public()
	ecPub, ok := pub.(*ecdsa.PublicKey)
	if !ok || ecPub == nil {
		return nil, errors.New("sm2: unexpected public key type")
	}
	pubBytes := sm2PublicKeyXY(ecPub.X, ecPub.Y)
	return &KeyPair{
		PrivateKey: base64.StdEncoding.EncodeToString(padLeft32(priv.D.Bytes())),
		PublicKey:  base64.StdEncoding.EncodeToString(pubBytes),
	}, nil
}

// sm2PublicKeyXY packs the affine coordinates X || Y into 64 raw bytes
// without the SEC1 04 prefix byte — the format js-agent's sm2ECDH uses.
// Coordinates are left-padded to 32 bytes each (big-endian, fixed width).
func sm2PublicKeyXY(x, y *big.Int) []byte {
	out := make([]byte, 64)
	xb := x.Bytes()
	yb := y.Bytes()
	copy(out[32-len(xb):32], xb)
	copy(out[64-len(yb):64], yb)
	return out
}

// padLeft32 pads a big-endian byte slice to exactly 32 bytes (left-padded
// with zeros). SM2 scalars are always 32 bytes; big.Int.Bytes() drops the
// leading zeros, so we re-add them to match js-agent's strict length check.
func padLeft32(b []byte) []byte {
	if len(b) == 32 {
		return b
	}
	out := make([]byte, 32)
	if len(b) > 32 {
		// Should not happen for a valid SM2 scalar; truncate defensively.
		copy(out, b[len(b)-32:])
		return out
	}
	copy(out[32-len(b):], b)
	return out
}

// sealKey wraps a base64-encoded private key under the configured AES-256
// master key using AES-GCM (12-byte nonce prepended to ciphertext+tag).
// The result is itself base64-encoded for safe SQLite TEXT storage.
//
// nonce (12) || ciphertext || tag (16) — all wrapped in base64 on the wire.
func sealKey(masterKey []byte, privKeyB64 string) (string, error) {
	if len(masterKey) != 32 {
		return "", errors.New("master key must be 32 bytes for AES-256-GCM")
	}
	privBytes, err := base64.StdEncoding.DecodeString(privKeyB64)
	if err != nil {
		return "", fmt.Errorf("decode private key: %w", err)
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("aes-gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}
	ct := gcm.Seal(nil, nonce, privBytes, nil)
	// nonce || ct (which already includes the GCM tag)
	combined := make([]byte, 0, len(nonce)+len(ct))
	combined = append(combined, nonce...)
	combined = append(combined, ct...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// openKey reverses sealKey and returns the original base64-encoded private
// key string, suitable for handing straight to js-agent.
func openKey(masterKey []byte, sealedB64 string) (string, error) {
	if len(masterKey) != 32 {
		return "", errors.New("master key must be 32 bytes for AES-256-GCM")
	}
	combined, err := base64.StdEncoding.DecodeString(sealedB64)
	if err != nil {
		return "", fmt.Errorf("decode sealed key: %w", err)
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("aes-gcm: %w", err)
	}
	if len(combined) < gcm.NonceSize() {
		return "", errors.New("sealed key too short")
	}
	nonce := combined[:gcm.NonceSize()]
	ct := combined[gcm.NonceSize():]
	privBytes, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt private key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(privBytes), nil
}
