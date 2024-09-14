package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

var (
	pwd = []byte{55, 74, 102, 112, 49, 46, 35, 41, 33, 105, 106, 73, 50, 49, 50, 51} // 7Jfp1.#)!ijI2123
	iv  = []byte{56, 76, 46, 40, 33, 106, 40, 106, 64, 106, 73, 46, 64, 35, 106, 46} // 8L.(!j(j@jI.@#j.
)

func AesEncrypt(str string) (string, error) {
	strBytes := []byte(str)

	block, err := aes.NewCipher(pwd)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	strBytes = pkcs5Padding(strBytes, blockSize)

	blockMode := cipher.NewCBCEncrypter(block, iv)
	crypted := make([]byte, len(strBytes))
	blockMode.CryptBlocks(crypted, strBytes)

	return base64.StdEncoding.EncodeToString(crypted), nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	// padding
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}

// AesDecrypt
func AesDecrypt(str string) (string, error) {
	// Decrypt the base64 first.
	strBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(pwd)
	if err != nil {
		return "", err
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(strBytes))

	blockMode.CryptBlocks(decrypted, strBytes)
	decrypted = pkcs5UnPadding(decrypted)
	if decrypted == nil {
		return "", fmt.Errorf("slice bounds out of range")
	}
	return string(decrypted), nil
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
