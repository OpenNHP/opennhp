package ztdo

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/emmansun/gmsm/sm4"

	"github.com/OpenNHP/opennhp/nhp/core"
)

const (
	InitialDHPKeyWrappingString = "DHP Data Private Key Wrapping"
)

// Data Key Pair generation interface
// - Support locally stored key generation by DB
// - Support KMS (Key Management Service) integration for secure key generation and management
// - Add TPM (Trusted Platform Module) based key derivation for hardware-backed security
// These extensions can be implemented by creating new types that satisfy the DataKeyPairGenerator interface.
type DataKeyPairGenerator interface {
	Generate(mode DataKeyPairECCMode) (privateKey []byte)
}

// Symmetric cipher mode provides symmetric encryption and decryption and supports Chinese standards and International standards.
type SymmetricCipherMode uint8

const (
	AES256GCM64Tag  SymmetricCipherMode = iota // 0x00
	AES256GCM96Tag                             // 0x01
	AES256GCM104Tag                            // 0x02
	AES256GCM112Tag                            // 0x03
	AES256GCM120Tag                            // 0x04
	AES256GCM128Tag                            // 0x05
	SM4GCM64Tag                                // 0x06
	SM4GCM128Tag                               // 0x07
)

func (m SymmetricCipherMode) String() string {
	switch m {
	case AES256GCM64Tag:
		return "AES-256-GCM-64"
	case AES256GCM96Tag:
		return "AES-256-GCM-96"
	case AES256GCM104Tag:
		return "AES-256-GCM-104"
	case AES256GCM112Tag:
		return "AES-256-GCM-112"
	case AES256GCM120Tag:
		return "AES-256-GCM-120"
	case AES256GCM128Tag:
		return "AES-256-GCM-128"
	case SM4GCM64Tag:
		return "SM4-GCM-64"
	case SM4GCM128Tag:
		return "SM4-GCM-128"
	default:
		return "Unknown"
	}
}

func (m SymmetricCipherMode) TagSize() int {
	switch m {
	case AES256GCM64Tag, SM4GCM64Tag:
		return 8
	case AES256GCM96Tag:
		return 12
	case AES256GCM104Tag:
		return 13
	case AES256GCM112Tag:
		return 14
	case AES256GCM120Tag:
		return 15
	case AES256GCM128Tag, SM4GCM128Tag:
		return 16
	default:
		return 0
	}
}

func NewSymmetricCipherMode(mode string) (SymmetricCipherMode, error) {
	switch mode {
	case "AES-256-GCM-64":
		return AES256GCM64Tag, nil
	case "AES-256-GCM-96":
		return AES256GCM96Tag, nil
	case "AES-256-GCM-104":
		return AES256GCM104Tag, nil
	case "AES-256-GCM-112":
		return AES256GCM112Tag, nil
	case "AES-256-GCM-120":
		return AES256GCM120Tag, nil
	case "AES-256-GCM-128":
		return AES256GCM128Tag, nil
	case "SM4-GCM-64":
		return SM4GCM64Tag, nil
	case "SM4-GCM-128":
		return SM4GCM128Tag, nil
	default:
		return 0, fmt.Errorf("unknown symmetric mode name: %s", mode)
	}
}
func (mode SymmetricCipherMode) newCipherBlock(key []byte) (cipher.Block, error) {
	switch mode {
	case AES256GCM64Tag, AES256GCM96Tag, AES256GCM104Tag,
		AES256GCM112Tag, AES256GCM120Tag, AES256GCM128Tag:
		if len(key) != 32 {
			return nil, fmt.Errorf("invalid key length for AES-256-GCM")
		}
		return aes.NewCipher(key)
	case SM4GCM64Tag, SM4GCM128Tag:
		if len(key) < 16 {
			return nil, fmt.Errorf("invalid key length for SM4-GCM")
		} else {
			key = key[:16]
		}
		return sm4.NewCipher(key)
	default:
		return nil, fmt.Errorf("unsupported mode: %v", mode)
	}
}

func (mode SymmetricCipherMode) Encrypt(key, nonce, plaintext, ad []byte) ([]byte, error) {
	tagSize := mode.TagSize()

	cipherBlock, err := mode.newCipherBlock(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCMWithTagSize(cipherBlock, tagSize)
	if err != nil {
		return nil, err
	}

	ciphertext := aead.Seal(plaintext[:0], nonce, plaintext, ad)

	return ciphertext, nil
}

func (mode SymmetricCipherMode) Decrypt(key, nonce, ciphertext, ad []byte) ([]byte, error) {
	tagSize := mode.TagSize()

	cipherBlock, err := mode.newCipherBlock(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCMWithTagSize(cipherBlock, tagSize)
	if err != nil {
		return nil, err
	}

	plaintext, err := aead.Open(ciphertext[:0], nonce, ciphertext, ad)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// DataKeyPairECCMode is a adapter for ECC key pair generation
type DataKeyPairECCMode uint8

const (
	CURVE25519 DataKeyPairECCMode = iota
	SM2
	UNKNOWN
)

func (d DataKeyPairECCMode) String() string {
	switch d {
	case CURVE25519:
		return "CURVE25519"
	case SM2:
		return "SM2"
	default:
		return "UNKNOWN"
	}
}

func (d DataKeyPairECCMode) ToEccType() core.EccTypeEnum {
	switch d {
	case CURVE25519:
		return core.ECC_CURVE25519
	case SM2:
		return core.ECC_SM2
	default:
		return core.ECC_UMI
	}
}

func (d DataKeyPairECCMode) ToHashType() core.HashTypeEnum {
	switch d {
	case CURVE25519:
		return core.HASH_SHA256
	case SM2:
		return core.HASH_SM3
	default:
		return core.HASH_SHA256
	}
}

func NewDataKeyPairECCModeWithName(mode string) (DataKeyPairECCMode, error) {
	switch mode {
	case "CURVE25519":
		return CURVE25519, nil
	case "SM2":
		return SM2, nil
	default:
		return 0, fmt.Errorf("unknown mode: %s", mode)
	}
}

func NewDataKeyPairECCMode(eccMode core.EccTypeEnum) (DataKeyPairECCMode, error) {
	switch eccMode {
	case core.ECC_CURVE25519:
		return CURVE25519, nil
	case core.ECC_SM2:
		return SM2, nil
	default:
		return 0, fmt.Errorf("unknown mode: %d", eccMode)
	}
}

func (d DataKeyPairECCMode) ECDHFromKey(prk []byte) core.Ecdh {
	return core.ECDHFromKey(d.ToEccType(), prk)
}

func (d DataKeyPairECCMode) PublicKeyFromKey(prk []byte) []byte {
	return core.ECDHFromKey(d.ToEccType(), prk).PublicKey()
}

// MessagePattern defines a set of tokens which are used during symmetric key agreement
type MessagePattern int

const (
	MessagePatternS MessagePattern = iota
	MessagePatternE
	MessagePatternRS
	MessagePatternRE
	MessagePatternDHEE
	MessagePatternDHES
	MessagePatternDHSE
	MessagePatternDHSS
)

type SymmetricAgreement struct {
	ss              core.NoiseFactory
	s               core.Ecdh          // local static key pair
	e               core.Ecdh          // local ephemeral key pair
	rs              []byte             // remote static key
	re              []byte             // remote ephemeral key
	messagePatterns [][]MessagePattern // 1st row is for provider, 2nd row for consumer
	psk             []byte             // pre-shared key
	isPskUsed       bool
	provider        bool               // provider or consumer
	eccMode         DataKeyPairECCMode // need to keep the same with NHP agent and NHP-DB
}

func NewSymmetricAgreement(eccMode DataKeyPairECCMode, provider bool) *SymmetricAgreement {
	sa := &SymmetricAgreement{
		psk:       []byte(""),
		isPskUsed: false,
		provider:  provider,
		eccMode:   eccMode,
	}

	sa.ss.HashType = eccMode.ToHashType()

	return sa
}

func (sa *SymmetricAgreement) SetPsk(psk []byte) {
	sa.psk = psk
	sa.isPskUsed = true
}

func (sa *SymmetricAgreement) SetStaticKeyPair(s core.Ecdh) {
	sa.s = s
}

func (sa *SymmetricAgreement) SetEphemeralKeyPair(e core.Ecdh) {
	sa.e = e
}

func (sa *SymmetricAgreement) SetRemoteStaticPublicKey(rs []byte) {
	sa.rs = rs
}

func (sa *SymmetricAgreement) SetRemoteEphemeralPublicKey(re []byte) {
	sa.re = re
}

func (sa *SymmetricAgreement) SetMessagePatterns(msgPatterns [][]MessagePattern) {
	sa.messagePatterns = msgPatterns
}

func (sa *SymmetricAgreement) AgreeSymmetricKey() (gcmKey [core.SymmetricKeySize]byte, ad []byte) {
	// ck is chaining key that hashes all previous DH outputs
	ck := [core.SymmetricKeySize]byte{}

	// adHash hashes all the involved public key, the final value will be used as associated data for AEAD authentication
	adHash, err := core.NewHash(sa.eccMode.ToHashType())
	if err != nil {
		panic("failed to create hash for symmetric agreement: " + err.Error())
	}

	if sa.isPskUsed {
		adHash.Write(sa.psk)
		sa.ss.KeyGen1(&ck, adHash.Sum(nil), sa.psk)
	}

	var msgPatterns []MessagePattern

	if sa.provider {
		msgPatterns = sa.messagePatterns[0]
	} else {
		msgPatterns = sa.messagePatterns[1]
	}

	for idx, pattern := range msgPatterns {
		switch pattern {
		case MessagePatternS:
			adHash.Write(sa.s.PublicKey())
		case MessagePatternE:
			adHash.Write(sa.e.PublicKey())
		case MessagePatternRS:
			adHash.Write(sa.rs)
		case MessagePatternRE:
			adHash.Write(sa.re)
		case MessagePatternDHSS:
			ss := sa.s.SharedSecret(sa.rs)
			if sa.provider {
				adHash.Write(sa.rs)
				sa.ss.MixKey(&ck, ck[:], sa.rs)
			} else {
				adHash.Write(sa.s.PublicKey())
				sa.ss.MixKey(&ck, ck[:], sa.s.PublicKey())
			}
			if idx == len(msgPatterns)-1 {
				sa.ss.KeyGen2(&ck, &gcmKey, ck[:], ss[:])
			} else {
				sa.ss.MixKey(&ck, ck[:], ss[:])
			}
		case MessagePatternDHSE:
			se := sa.s.SharedSecret(sa.re)
			if sa.provider {
				adHash.Write(sa.re)
				sa.ss.MixKey(&ck, ck[:], sa.re)
			} else {
				adHash.Write(sa.s.PublicKey())
				sa.ss.MixKey(&ck, ck[:], sa.s.PublicKey())
			}
			if idx == len(msgPatterns)-1 {
				sa.ss.KeyGen2(&ck, &gcmKey, ck[:], se[:])
			} else {
				sa.ss.MixKey(&ck, ck[:], se[:])
			}
		case MessagePatternDHEE:
			ee := sa.e.SharedSecret(sa.re)
			if sa.provider {
				adHash.Write(sa.re)
				sa.ss.MixKey(&ck, ck[:], sa.re)
			} else {
				adHash.Write(sa.e.PublicKey())
				sa.ss.MixKey(&ck, ck[:], sa.e.PublicKey())
			}
			if idx == len(msgPatterns)-1 {
				sa.ss.KeyGen2(&ck, &gcmKey, ck[:], ee[:])
			} else {
				sa.ss.MixKey(&ck, ck[:], ee[:])
			}
		case MessagePatternDHES:
			es := sa.e.SharedSecret(sa.rs)
			if sa.provider {
				adHash.Write(sa.rs)
				sa.ss.MixKey(&ck, ck[:], sa.rs)
			} else {
				adHash.Write(sa.e.PublicKey())
				sa.ss.MixKey(&ck, ck[:], sa.e.PublicKey())
			}
			if idx == len(msgPatterns)-1 {
				sa.ss.KeyGen2(&ck, &gcmKey, ck[:], es[:])
			} else {
				sa.ss.MixKey(&ck, ck[:], es[:])
			}
		}
	}

	ad = adHash.Sum(nil)

	return
}

// Message patterns that are used for agreeing symmetric key to be used for data private key encryption and decryption.
var DataPrivateKeyWrappingPatterns = [][]MessagePattern{
	{MessagePatternDHSS, MessagePatternS, MessagePatternDHSE},
	{MessagePatternDHSS, MessagePatternRS, MessagePatternDHES},
}

type DataPrivateKeyWrapping struct {
	ProviderPublicKeyBase64 string `json:"providerPublicKeyBase64"`
	IvBase64                string `json:"ivBase64"`
	PrkWrapping             string `json:"prkWrapping"`
}

func NewDataPrivateKeyWrapping(providerPublicKeyBase64 string, dataPrivateKeyBase64 string, key, ad []byte) *DataPrivateKeyWrapping {
	symmetricCipherMode := AES256GCM128Tag
	var Iv [core.GCMNonceSize]byte
	if _, err := rand.Read(Iv[:]); err != nil {
		panic("crypto/rand.Read failed: " + err.Error())
	}

	cipherText, _ := symmetricCipherMode.Encrypt(key, Iv[:], []byte(dataPrivateKeyBase64), ad) //nolint:gosec // G104: Encrypt with valid key/IV doesn't fail

	return &DataPrivateKeyWrapping{
		ProviderPublicKeyBase64: providerPublicKeyBase64,
		IvBase64:                base64.StdEncoding.EncodeToString(Iv[:]),
		PrkWrapping:             base64.StdEncoding.EncodeToString(cipherText),
	}
}

func (d *DataPrivateKeyWrapping) Unwrap(key, ad []byte) (dataPrivateKeyBase64 string, err error) {
	symmetricCipherMode := AES256GCM128Tag

	Iv, err := base64.StdEncoding.DecodeString(d.IvBase64)
	if err != nil {
		return
	}

	cipherText, err := base64.StdEncoding.DecodeString(d.PrkWrapping)
	if err != nil {
		return
	}

	dataPrk, err := symmetricCipherMode.Decrypt(key, Iv, cipherText, ad)
	if err != nil {
		return
	}

	dataPrivateKeyBase64 = string(dataPrk)

	return
}
