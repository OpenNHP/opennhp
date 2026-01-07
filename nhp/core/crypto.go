package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"log"
	"os"

	"github.com/emmansun/gmsm/padding"
	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm3"
	"github.com/emmansun/gmsm/sm4"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/chacha20poly1305"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core/scheme/curve"
	"github.com/OpenNHP/opennhp/nhp/core/scheme/gmsm"
	"github.com/OpenNHP/opennhp/nhp/utils"
)

type HashTypeEnum int

const (
	HASH_BLAKE2S HashTypeEnum = iota
	HASH_SM3
	HASH_SHA256
)

type EccTypeEnum int

const (
	ECC_CURVE25519 EccTypeEnum = iota
	ECC_SM2
	ECC_UMI
)

type GcmTypeEnum int

const (
	GCM_AES256 GcmTypeEnum = iota
	GCM_SM4
	GCM_CHACHA20POLY1305
)

type CipherSuite struct {
	Scheme   int
	EccType  EccTypeEnum
	HashType HashTypeEnum
	GcmType  GcmTypeEnum
}

// init cipher suite
func NewCipherSuite(scheme int) (ciphers *CipherSuite) {
	// init cipher suite
	switch scheme {
	case common.CIPHER_SCHEME_GMSM:
		ciphers = &CipherSuite{
			Scheme:   common.CIPHER_SCHEME_GMSM,
			HashType: HASH_SM3,
			EccType:  ECC_SM2,
			GcmType:  GCM_SM4,
		}

	case common.CIPHER_SCHEME_CURVE:
		fallthrough
	default:
		ciphers = &CipherSuite{
			Scheme:   common.CIPHER_SCHEME_CURVE,
			HashType: HASH_BLAKE2S,
			EccType:  ECC_CURVE25519,
			GcmType:  GCM_AES256,
		}
	}
	return
}

func NewHash(t HashTypeEnum) (hash.Hash, error) {
	switch t {
	case HASH_BLAKE2S:
		h, err := blake2s.New256(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create blake2s hash: %w", err)
		}
		return h, nil

	case HASH_SM3:
		return sm3.New(), nil

	case HASH_SHA256:
		return sha256.New(), nil

	default:
		return nil, fmt.Errorf("unsupported hash type: %d", t)
	}
}

type Ecdh interface {
	SetPrivateKey(prk []byte) error
	PrivateKey() []byte
	PublicKey() []byte
	SharedSecret(pbk []byte) []byte
	Name() string
	PrivateKeyBase64() string
	PublicKeyBase64() string
	Identity() []byte
	MidPublicKey() []byte
}

func ECDHFromKey(t EccTypeEnum, prk []byte) (e Ecdh) {
	switch t {
	case ECC_CURVE25519:
		var c curve.Curve25519ECDH
		err := c.SetPrivateKey(prk)
		if err != nil {
			return nil
		}
		e = &c

	case ECC_SM2:
		var s gmsm.SM2ECDH
		err := s.SetPrivateKey(prk)
		if err != nil {
			return nil
		}
		e = &s
	}

	return e
}

func NewECDH(t EccTypeEnum) (e Ecdh) {
	switch t {
	case ECC_CURVE25519:
		e = curve.NewECDH()

	case ECC_SM2:
		e = gmsm.NewECDH()
	}

	return e
}

func AeadFromKey(t GcmTypeEnum, key *[SymmetricKeySize]byte) (cipher.AEAD, error) {
	switch t {
	case GCM_AES256:
		aesBlock, err := aes.NewCipher(key[:])
		if err != nil {
			return nil, fmt.Errorf("failed to create AES cipher: %w", err)
		}
		aead, err := cipher.NewGCM(aesBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to create AES-GCM: %w", err)
		}
		return aead, nil

	case GCM_SM4:
		sm4Block, err := sm4.NewCipher(key[:16])
		if err != nil {
			return nil, fmt.Errorf("failed to create SM4 cipher: %w", err)
		}
		aead, err := cipher.NewGCM(sm4Block)
		if err != nil {
			return nil, fmt.Errorf("failed to create SM4-GCM: %w", err)
		}
		return aead, nil

	case GCM_CHACHA20POLY1305:
		aead, err := chacha20poly1305.New(key[:])
		if err != nil {
			return nil, fmt.Errorf("failed to create ChaCha20-Poly1305: %w", err)
		}
		return aead, nil

	default:
		return nil, fmt.Errorf("unsupported GCM type: %d", t)
	}
}

func CBCEncryption(t GcmTypeEnum, key *[SymmetricKeySize]byte, plaintext []byte, inPlace bool) ([]byte, error) {
	var block cipher.Block
	var iv []byte
	var err error
	switch t {
	case GCM_AES256:
		block, err = aes.NewCipher(key[:])
		if err != nil {
			return nil, fmt.Errorf("failed to create AES cipher for CBC: %w", err)
		}
		iv = key[8:24]

	case GCM_SM4:
		block, err = sm4.NewCipher(key[:16])
		if err != nil {
			return nil, fmt.Errorf("failed to create SM4 cipher for CBC: %w", err)
		}
		iv = key[16:]

	case GCM_CHACHA20POLY1305:
		return nil, ErrNotApplicable

	default:
		return nil, fmt.Errorf("unsupported cipher type for CBC: %d", t)
	}

	var paddedPlainText []byte
	if len(plaintext)%block.BlockSize() == 0 {
		// skip padding
		paddedPlainText = plaintext
	} else {
		pkcs7 := padding.NewPKCS7Padding(uint(block.BlockSize()))
		paddedPlainText = pkcs7.Pad(plaintext)
	}

	var ciphertext []byte
	if inPlace {
		ciphertext = paddedPlainText
	} else {
		ciphertext = make([]byte, 0, len(plaintext))
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(ciphertext, paddedPlainText)

	return ciphertext, nil
}

func CBCDecryption(t GcmTypeEnum, key *[SymmetricKeySize]byte, ciphertext []byte, inPlace bool) ([]byte, error) {
	var block cipher.Block
	var iv []byte
	var err error
	switch t {
	case GCM_AES256:
		block, err = aes.NewCipher(key[:])
		if err != nil {
			return nil, fmt.Errorf("failed to create AES cipher for CBC decryption: %w", err)
		}
		iv = key[8:24]

	case GCM_SM4:
		block, err = sm4.NewCipher(key[:16])
		if err != nil {
			return nil, fmt.Errorf("failed to create SM4 cipher for CBC decryption: %w", err)
		}
		iv = key[16:]

	case GCM_CHACHA20POLY1305:
		return nil, ErrNotApplicable

	default:
		return nil, fmt.Errorf("unsupported cipher type for CBC decryption: %d", t)
	}

	// Validate ciphertext: must be at least one block and a multiple of block size
	if len(ciphertext) < block.BlockSize() {
		return nil, fmt.Errorf("ciphertext too short: need at least %d bytes", block.BlockSize())
	}
	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, fmt.Errorf("ciphertext length %d is not a multiple of block size %d", len(ciphertext), block.BlockSize())
	}

	var plaintext []byte
	if inPlace {
		plaintext = ciphertext
	} else {
		plaintext = make([]byte, len(ciphertext))
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(plaintext, ciphertext)

	if len(plaintext)%block.BlockSize() == 0 {
		// skip unpadding
	} else {
		// Unpad plaintext
		pkcs7 := padding.NewPKCS7Padding(uint(block.BlockSize()))
		plaintext, err = pkcs7.Unpad(plaintext)
		if err != nil {
			return nil, err
		}
	}

	return plaintext, nil
}

func SM2Encrypt(pubKeyBase64 string, message string) (string, error) {
	//ASN.1

	// real public key should be from cert or public key pem file
	sm2PublicKey, err := gmsm.Base64DecodeSM2ECDSAPublicKey(pubKeyBase64)

	secretMessage := []byte(message)
	// crypto/rand.Reader is a good source of entropy for randomizing the
	// encryption function.
	rng := rand.Reader

	ciphertext, err := sm2.EncryptASN1(rng, sm2PublicKey, secretMessage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from encryption: %s\n", err)
		return "", err
	}
	// Since encryption is a randomized function, ciphertext will be
	// different each time.
	fmt.Printf("Ciphertext: %x\n", ciphertext)
	return hex.EncodeToString(ciphertext), err
}

func SM2Decrypt(privateKeyBase64 string, message string) (string, error) {
	//ASN.1
	ciphertext, err := hex.DecodeString(message)
	privKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return "", fmt.Errorf("size incorrect")
	}

	testkey, err := sm2.NewPrivateKey(privKeyBytes)
	if err != nil {
		log.Fatalf("fail to new private key %v", err)
	}

	sourceText, err := testkey.Decrypt(nil, ciphertext, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from decryption: %s\n", err)
		return "", err
	}
	return string(sourceText), err
}

// AESEncryption Function
func AESEncrypt(plainText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	//Padding Plaintext
	plainText = pad(plainText, aes.BlockSize)
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], plainText)
	return cipherText, nil
}

// pad adds PKCS#7 padding to data. Uses shared implementation from utils.
func pad(data []byte, blockSize int) []byte {
	return utils.PKCS7Pad(data, blockSize)
}

func AESDecrypt(cipherText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// Validate ciphertext length:
	// - Must have at least IV (16 bytes) + one encrypted block (16 bytes)
	// - After IV extraction, remaining must be a multiple of block size
	if len(cipherText) < aes.BlockSize*2 {
		return nil, fmt.Errorf("cipherText too short: need at least %d bytes, got %d", aes.BlockSize*2, len(cipherText))
	}
	if (len(cipherText)-aes.BlockSize)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("cipherText length invalid: must be IV + multiple of block size")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	// Decrypt
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(cipherText))
	mode.CryptBlocks(decrypted, cipherText)

	// Remove padding
	decrypted = unpad(decrypted, aes.BlockSize)

	return decrypted, nil
}

// unpad removes PKCS#7 padding from data. Uses shared implementation from utils.
func unpad(padded []byte, blockSize int) []byte {
	return utils.PKCS7Unpad(padded, blockSize)
}
