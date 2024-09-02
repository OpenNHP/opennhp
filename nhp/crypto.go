package nhp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"math/big"

	"github.com/emmansun/gmsm/ecdh"
	"github.com/emmansun/gmsm/sm2"
	"github.com/emmansun/gmsm/sm3"
	"github.com/emmansun/gmsm/sm4"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
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
)

type GcmTypeEnum int

const (
	GCM_AES256 GcmTypeEnum = iota
	GCM_SM4
	GCM_CHACHA20POLY1305
)

type CipherSuite struct {
	EccType  EccTypeEnum
	HashType HashTypeEnum
	GcmType  GcmTypeEnum
}

// init cipher suite
func NewCipherSuite(useGM bool) (ciphers *CipherSuite) {
	// init cipher suite
	if useGM {
		ciphers = &CipherSuite{
			HashType: HASH_SM3,
			EccType:  ECC_SM2,
			GcmType:  GCM_SM4,
		}
	} else {
		ciphers = &CipherSuite{
			HashType: HASH_BLAKE2S,
			EccType:  ECC_CURVE25519,
			GcmType:  GCM_AES256,
		}
	}

	return
}

func (c *CipherSuite) IsUseGm() bool {
	return c.EccType == ECC_SM2
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
	SharedSecret(pbk []byte) *[SymmetricKeySize]byte
	Name() string
	PrivateKeyBase64() string
	PublicKeyBase64() string
}

func ECDHFromKey(t EccTypeEnum, prk []byte) (e Ecdh) {
	switch t {
	case ECC_CURVE25519:
		var c Curve25519ECDH
		err := c.SetPrivateKey(prk)
		if err != nil {
			return nil
		}
		e = &c

	case ECC_SM2:
		var s SM2ECDH
		err := s.SetPrivateKey(prk)
		if err != nil {
			return nil
		}
		e = &s
	}

	return e
}

func NewECDH(t EccTypeEnum) (e Ecdh) {
	var err error
	switch t {
	case ECC_CURVE25519:
		key := make([]byte, PrivateKeySize)
		_, err = rand.Read(key)
		if err != nil {
			return nil
		}
		// clamp
		key[0] &= 248
		key[31] = (key[31] & 127) | 64
		var c Curve25519ECDH
		c.SetPrivateKey(key)
		e = &c

	case ECC_SM2:
		var s SM2ECDH
		s.prvK, err = ecdh.P256().GenerateKey(rand.Reader)
		if err != nil {
			return nil
		}
		copy(s.PrivKey[:], s.prvK.Bytes())
		copy(s.PubKey[:], s.prvK.PublicKey().Bytes()[1:1+PublicKeySizeEx])
		s.PrivKeyBase64 = base64.StdEncoding.EncodeToString(s.PrivKey[:])
		s.PubKeyBase64 = base64.StdEncoding.EncodeToString(s.PubKey[:])
		e = &s
	}

	return e
}

type Curve25519ECDH struct {
	PrivKey       [PrivateKeySize]byte
	PubKey        [PublicKeySize]byte
	PrivKeyBase64 string
	PubKeyBase64  string
}

func (c *Curve25519ECDH) SetPrivateKey(prk []byte) (err error) {
	copy(c.PrivKey[:], prk[:PrivateKeySize])
	pbk, err := curve25519.X25519(c.PrivKey[:], curve25519.Basepoint)
	if err != nil {
		return err
	}
	copy(c.PubKey[:], pbk)
	c.PrivKeyBase64 = base64.StdEncoding.EncodeToString(c.PrivKey[:])
	c.PubKeyBase64 = base64.StdEncoding.EncodeToString(c.PubKey[:])

	return nil
}

func (c *Curve25519ECDH) PrivateKey() []byte {
	return c.PrivKey[:]
}

func (c *Curve25519ECDH) PrivateKeyBase64() string {
	return c.PrivKeyBase64
}

func (c *Curve25519ECDH) PublicKey() []byte {
	return c.PubKey[:]
}

func (c *Curve25519ECDH) PublicKeyBase64() string {
	return c.PubKeyBase64
}

func (c *Curve25519ECDH) SharedSecret(pbk []byte) *[SymmetricKeySize]byte {
	if len(pbk) != PublicKeySize {
		return nil
	}

	key, err := curve25519.X25519(c.PrivKey[:], pbk)
	if err != nil {
		return nil
	}

	var ss [SymmetricKeySize]byte
	copy(ss[:], key)
	return &ss
}

func (c *Curve25519ECDH) Name() string {
	return fmt.Sprintf("%s...%s", c.PubKeyBase64[0:4], c.PubKeyBase64[39:43])
}

type SM2ECDH struct {
	PrivKey       [PrivateKeySize]byte
	PubKey        [PublicKeySizeEx]byte
	prvK          *ecdh.PrivateKey
	PrivKeyBase64 string
	PubKeyBase64  string
}

func (s *SM2ECDH) SetPrivateKey(prk []byte) (err error) {
	copy(s.PrivKey[:], prk[:PrivateKeySize])
	s.prvK, err = ecdh.P256().NewPrivateKey(prk)
	if err != nil {
		return err
	}
	copy(s.PubKey[:], s.prvK.PublicKey().Bytes()[1:1+PublicKeySizeEx])
	s.PrivKeyBase64 = base64.StdEncoding.EncodeToString(s.PrivKey[:])
	s.PubKeyBase64 = base64.StdEncoding.EncodeToString(s.PubKey[:])

	return nil
}

func (s *SM2ECDH) PrivateKey() []byte {
	return s.PrivKey[:]
}

func (c *SM2ECDH) PrivateKeyBase64() string {
	return c.PrivKeyBase64
}

func (s *SM2ECDH) PublicKey() []byte {
	return s.PubKey[:]
}

func (c *SM2ECDH) PublicKeyBase64() string {
	return c.PubKeyBase64
}

func (s *SM2ECDH) SharedSecret(pbk []byte) *[SymmetricKeySize]byte {
	if len(pbk) != PublicKeySizeEx {
		return nil
	}

	var pkBytes [1 + PublicKeySizeEx]byte
	pkBytes[0] = 4 // uncompressed header
	copy(pkBytes[1:1+PublicKeySizeEx], pbk[:PublicKeySizeEx])

	pubK, err := ecdh.P256().NewPublicKey(pkBytes[:])
	if err != nil {
		return nil
	}

	key, err := s.prvK.ECDH(pubK)
	if err != nil {
		return nil
	}

	var ss [SymmetricKeySize]byte
	copy(ss[:], key)
	return &ss
}

func (s *SM2ECDH) Name() string {
	return fmt.Sprintf("%s...%s", s.PubKeyBase64[0:4], s.PubKeyBase64[39:43])
}

func AeadFromKey(t GcmTypeEnum, key *[SymmetricKeySize]byte) (aead cipher.AEAD) {
	switch t {
	case GCM_AES256:
		aesBlock, _ := aes.NewCipher(key[:])
		aead, _ = cipher.NewGCM(aesBlock)

	case GCM_SM4:
		sm4Block, _ := sm4.NewCipher(key[8:24])
		aead, _ = cipher.NewGCM(sm4Block)

	case GCM_CHACHA20POLY1305:
		aead, _ = chacha20poly1305.New(key[:])
	}

	return aead
}

// 生成用于ECDH的SM2公私钥对
func GenerateSM2ECDHKeypair() (string, string) {
	var err error
	var pubKey [64]byte
	var privKey [32]byte

	pKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return "", ""
	}
	copy(privKey[:32], pKey.Bytes()[:32])             // 私钥32 bytes
	copy(pubKey[:64], pKey.PublicKey().Bytes()[1:65]) // 公钥64 bytes

	return base64.StdEncoding.EncodeToString(pubKey[:]),
		base64.StdEncoding.EncodeToString(privKey[:])
}

func Base64DecodeSM2ECDHPrivateKey(privStr string) (*ecdh.PrivateKey, error) {
	privKeyBytes, err := base64.StdEncoding.DecodeString(privStr)
	if err != nil {
		return nil, err
	}
	if len(privKeyBytes) != 32 {
		return nil, fmt.Errorf("size incorrect")
	}
	privKey, err := ecdh.P256().NewPrivateKey(privKeyBytes)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}

func Base64DecodeSM2ECDHPublicKey(pubStr string) (*ecdh.PublicKey, error) {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubStr)
	if err != nil {
		return nil, err
	}
	if len(pubKeyBytes) != 64 {
		return nil, fmt.Errorf("size incorrect")
	}
	buf := make([]byte, 65)
	buf[0] = 4 // public key first byte means uncompressed
	copy(buf[1:], pubKeyBytes[:])
	pubKey, err := ecdh.P256().NewPublicKey(buf[:])
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

// 生成用于ECDSA的SM2公私钥对
func GenerateSM2ECDSAKeypair() (*sm2.PrivateKey, string, string) {
	var err error
	var pubKeyBytes [64]byte
	var privKeyBytes [32]byte
	privKey := new(sm2.PrivateKey)

	for {
		privKey, err = sm2.GenerateKey(rand.Reader)
		if err == nil {
			break
		}
	}
	copy(privKeyBytes[:], privKey.D.Bytes()[:])            // 32 bytes D
	copy(pubKeyBytes[:32], privKey.PublicKey.X.Bytes()[:]) // 32 bytes X
	copy(pubKeyBytes[32:], privKey.PublicKey.Y.Bytes()[:]) // 32 bytes Y

	return privKey,
		base64.StdEncoding.EncodeToString(pubKeyBytes[:]),
		base64.StdEncoding.EncodeToString(privKeyBytes[:])
}

func Base64DecodeSM2ECDSAPrivateKey(pubKeyStr string, privKeyStr string) (*sm2.PrivateKey, error) {
	privKeyBytes, err := base64.StdEncoding.DecodeString(privKeyStr)
	if err != nil {
		return nil, err
	}
	if len(privKeyBytes) != 32 {
		return nil, fmt.Errorf("size incorrect")
	}
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		return nil, err
	}
	if len(pubKeyBytes) != 64 {
		return nil, fmt.Errorf("size incorrect")
	}

	privKey := &sm2.PrivateKey{}
	privKey.D = new(big.Int)
	privKey.D.SetBytes(privKeyBytes[:])

	privKey.PublicKey.Curve = sm2.P256()
	privKey.PublicKey.X = new(big.Int)
	privKey.PublicKey.Y = new(big.Int)
	privKey.PublicKey.X.SetBytes(pubKeyBytes[:32])
	privKey.PublicKey.Y.SetBytes(pubKeyBytes[32:])

	return privKey, nil
}

func Base64DecodeSM2ECDSAPublicKey(pubKeyStr string) (*ecdsa.PublicKey, error) {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		return nil, err
	}
	if len(pubKeyBytes) != 64 {
		return nil, fmt.Errorf("size incorrect")
	}

	pubKey := &ecdsa.PublicKey{
		Curve: sm2.P256(),
		X:     new(big.Int),
		Y:     new(big.Int),
	}

	pubKey.X.SetBytes(pubKeyBytes[:32])
	pubKey.Y.SetBytes(pubKeyBytes[32:])

	return pubKey, nil
}
