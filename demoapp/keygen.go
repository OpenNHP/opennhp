package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// KeyPair is a freshly generated NHP key pair. The private key is base64-
// encoded in the format expected by the browser-side js-agent; the public
// key likewise. We store the private key encrypted (see sealKey) but pass
// the public key around in cleartext.
type KeyPair struct {
	PrivateKey string // base64-encoded raw private key bytes
	PublicKey  string // base64-encoded raw public key bytes
}

// eccTypeFor maps a demoapp CipherScheme to the nhp/core EccTypeEnum used by
// ECDHFromKey. This is the single place the two enums are bridged.
func eccTypeFor(scheme CipherScheme) core.EccTypeEnum {
	if scheme == CipherSchemeGMSM {
		return core.ECC_SM2
	}
	return core.ECC_CURVE25519
}

// GenerateSchemeAgnosticPrivateKey generates a single 32-byte NHP private
// key that is valid under BOTH cipher schemes. It mirrors nhp-server's
// `keygen --both` path (endpoints/server/main/main.go:64-83): an SM2
// scalar is produced first (it must lie in [1, N-1] for SM2), and the same
// 32 bytes are a legal X25519 private key (X25519 clamps internally).
//
// The scheme is NOT chosen here — the private key is scheme-agnostic. The
// matching public key is derived later via DerivePublicKey once the user
// picks a scheme at registration. js-agent derives the same public key in
// the browser from this private key + cipherScheme, so both sides agree.
func GenerateSchemeAgnosticPrivateKey() (string, error) {
	e := core.NewECDH(core.ECC_SM2)
	if e == nil {
		return "", errors.New("core.NewECDH(ECC_SM2) returned nil")
	}
	return e.PrivateKeyBase64(), nil
}

// DerivePublicKey computes the public key for the given scheme from a
// scheme-agnostic base64 private key. Uses core.ECDHFromKey — the same
// primitive nhp-server's `pubkey --both` and `keygen --both` use, so the
// derived public key is byte-identical to what the server and js-agent
// produce from the same private key.
func DerivePublicKey(privKeyB64 string, scheme CipherScheme) (string, error) {
	privBytes, err := base64.StdEncoding.DecodeString(privKeyB64)
	if err != nil {
		return "", fmt.Errorf("decode private key: %w", err)
	}
	e := core.ECDHFromKey(eccTypeFor(scheme), privBytes)
	if e == nil {
		return "", fmt.Errorf("ECDHFromKey returned nil for scheme %s", scheme)
	}
	return e.PublicKeyBase64(), nil
}

// GenerateKeyPair creates a new key pair for the given cipher scheme. It is
// a compatibility wrapper over GenerateSchemeAgnosticPrivateKey +
// DerivePublicKey, retained for callers (e.g. the OIDC default path) that
// still want a ready-made (priv, pub) pair in one call.
func GenerateKeyPair(scheme CipherScheme) (*KeyPair, error) {
	priv, err := GenerateSchemeAgnosticPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("keygen: %w", err)
	}
	pub, err := DerivePublicKey(priv, scheme)
	if err != nil {
		return nil, fmt.Errorf("derive public key: %w", err)
	}
	return &KeyPair{PrivateKey: priv, PublicKey: pub}, nil
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
