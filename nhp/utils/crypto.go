package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// PKCS7Pad adds PKCS#7 padding to data to make it a multiple of blockSize.
// PKCS#5 is a subset of PKCS#7 with a fixed block size of 8 bytes.
// This implementation works for any block size (typically 8 or 16 bytes).
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// PKCS7Unpad removes PKCS#7 padding from data.
// Returns nil if the padding is invalid.
func PKCS7Unpad(data []byte, blockSize int) []byte {
	length := len(data)
	if length == 0 {
		return nil
	}
	unpadLen := int(data[length-1])
	if unpadLen > blockSize || unpadLen > length || unpadLen == 0 {
		return nil
	}
	// Verify padding bytes are all the same value
	for i := length - unpadLen; i < length; i++ {
		if data[i] != byte(unpadLen) {
			return nil
		}
	}
	return data[:length-unpadLen]
}

// pkcs5Padding is an alias for PKCS7Pad for backward compatibility.
// Deprecated: Use PKCS7Pad instead.
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	return PKCS7Pad(ciphertext, blockSize)
}

// pkcs5UnPadding is an alias for PKCS7Unpad for backward compatibility.
// Deprecated: Use PKCS7Unpad instead.
func pkcs5UnPadding(origData []byte) []byte {
	return PKCS7Unpad(origData, 16) // Default to AES block size
}

// PubKeyFingerprintLen is the length in characters of the short ID produced
// by PubKeyFingerprint. It comes from base64.RawURLEncoding of the first 8
// bytes of SHA-256 (8 bytes → 11 base64url chars, unpadded).
const PubKeyFingerprintLen = 11

// PubKeyFingerprint returns a short, URL-safe identifier derived from a raw
// public key: base64url(SHA-256(rawPubKey)[:8]), 12 characters with no
// padding. The same public key always produces the same fingerprint, so
// both Go and TypeScript implementations can compute it independently.
//
// This is a stable routing identifier, not a security primitive. The
// truncation gives ~2^-32 collision probability per pair of distinct keys,
// which is acceptable for the small number of upstream nhp-server clusters a
// relay will host. Callers MUST NOT use it as an authentication token.
func PubKeyFingerprint(rawPubKey []byte) string {
	sum := sha256.Sum256(rawPubKey)
	return base64.RawURLEncoding.EncodeToString(sum[:8])
}

// PubKeyFingerprintFromBase64 is a convenience wrapper that decodes a
// standard-base64-encoded public key (as stored in TOML configs) before
// hashing. Returns the fingerprint and any decode error.
func PubKeyFingerprintFromBase64(pubKeyBase64 string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(pubKeyBase64)
	if err != nil {
		return "", err
	}
	return PubKeyFingerprint(raw), nil
}

func HMACSha256(key, value string) []byte {
	var secretKey = []byte(key)
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(value))

	hash := h.Sum(nil)

	return hash
}

//nolint:gosec // G401: MD5 used for non-cryptographic checksums, not for security
func MD5(value string) string {
	_16bytes := md5.Sum([]byte(value))
	return hex.EncodeToString(_16bytes[:])
}

func Base64(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}

func GenerateRsaKey(bits int) (string, string) {
	// Generate private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return "", ""
	}
	pivKey := x509.MarshalPKCS1PrivateKey(privateKey)
	pubKey := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)

	return base64.StdEncoding.EncodeToString(pivKey), base64.StdEncoding.EncodeToString(pubKey)
}

// Md5sum computes MD5 checksum for file integrity verification (not cryptographic security)
//
//nolint:gosec // G401: MD5 used for file integrity checksums, not for cryptographic security
func Md5sum(fullFilePath string) (string, error) {
	fileInfo, err := os.Stat(fullFilePath)
	if err != nil {
		return "", fmt.Errorf("file not found: %w", err)
	}

	if !fileInfo.Mode().IsRegular() {
		return "", fmt.Errorf("path is not a regular file")
	}

	file, err := os.Open(fullFilePath) //nolint:gosec // G304: Path validated by os.Stat above
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := md5.New()

	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Convert hash to hex string
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
