package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"log"
	"os"

	"github.com/emmansun/gmsm/padding"
	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm3"
	"github.com/emmansun/gmsm/sm4"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/chacha20poly1305"

	"github.com/OpenNHP/opennhp/core/scheme/curve"
	"github.com/OpenNHP/opennhp/core/scheme/gmsm"
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

const (
	CIPHER_SCHEME_CURVE int = iota
	CIPHER_SCHEME_GMSM
)

// init cipher suite
func NewCipherSuite(scheme int) (ciphers *CipherSuite) {
	// init cipher suite
	switch scheme {
	case CIPHER_SCHEME_CURVE:
		ciphers = &CipherSuite{
			Scheme:   CIPHER_SCHEME_CURVE,
			HashType: HASH_BLAKE2S,
			EccType:  ECC_CURVE25519,
			GcmType:  GCM_AES256,
		}

	case CIPHER_SCHEME_GMSM:
		fallthrough
	default:
		ciphers = &CipherSuite{
			Scheme:   CIPHER_SCHEME_GMSM,
			HashType: HASH_SM3,
			EccType:  ECC_SM2,
			GcmType:  GCM_SM4,
		}
	}
	return
}

func NewHash(t HashTypeEnum) (h hash.Hash) {
	switch t {
	case HASH_BLAKE2S:
		h, _ = blake2s.New256(nil)

	case HASH_SM3:
		h = sm3.New()

	case HASH_SHA256:
		h = sha256.New()
	}

	return h
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

func AeadFromKey(t GcmTypeEnum, key *[SymmetricKeySize]byte) (aead cipher.AEAD) {
	switch t {
	case GCM_AES256:
		aesBlock, _ := aes.NewCipher(key[:])
		aead, _ = cipher.NewGCM(aesBlock)

	case GCM_SM4:
		sm4Block, _ := sm4.NewCipher(key[:16])
		aead, _ = cipher.NewGCM(sm4Block)

	case GCM_CHACHA20POLY1305:
		aead, _ = chacha20poly1305.New(key[:])
	}

	return aead
}

func CBCEncryption(t GcmTypeEnum, key *[SymmetricKeySize]byte, plaintext []byte, inPlace bool) ([]byte, error) {
	var block cipher.Block
	var iv []byte
	switch t {
	case GCM_AES256:
		block, _ = aes.NewCipher(key[:])
		iv = key[8:24]

	case GCM_SM4:
		block, _ = sm4.NewCipher(key[:16])
		iv = key[16:]

	case GCM_CHACHA20POLY1305:
		return nil, ErrNotApplicable
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
		block, _ = aes.NewCipher(key[:])
		iv = key[8:24]

	case GCM_SM4:
		block, _ = sm4.NewCipher(key[:16])
		iv = key[16:]

	case GCM_CHACHA20POLY1305:
		return nil, ErrNotApplicable
	}

	if len(ciphertext) < block.BlockSize() {
		return nil, fmt.Errorf("ciphertext too short")
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

func ECCEncryption(pubKey string, message string) (string, error) {
	//ASN.1
	// real public key should be from cert or public key pem file
	keypoints, _ := hex.DecodeString(pubKey)
	pubKeyByte, err := sm2.NewPublicKey(keypoints)
	if err != nil {
		fmt.Println("fail to new public key %v", err)
		return "", err
	}
	secretMessage := []byte(message)
	// crypto/rand.Reader is a good source of entropy for randomizing the
	// encryption function.
	rng := rand.Reader

	ciphertext, err := sm2.EncryptASN1(rng, pubKeyByte, secretMessage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from encryption: %s\n", err)
		return "", err
	}
	// Since encryption is a randomized function, ciphertext will be
	// different each time.
	fmt.Printf("Ciphertext: %x\n", ciphertext)
	return hex.EncodeToString(ciphertext), err
}

func ECCDecrypt(privateKey string, message string) (string, error) {
	//ASN.1
	ciphertext, err := hex.DecodeString(message)
	// real private key should be from secret storage
	privKey, _ := hex.DecodeString(privateKey)
	testkey, err := sm2.NewPrivateKey(privKey)
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
