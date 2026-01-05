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

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	// padding
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	if length < unpadding {
		return nil
	}
	return origData[:(length - unpadding)]
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
