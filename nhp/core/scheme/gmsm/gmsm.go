package gmsm

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/emmansun/gmsm/ecdh"
	"github.com/emmansun/gmsm/sm2"
)

const (
	PrivateKeySize = 32
	PublicKeySize  = 64
)

type SM2ECDH struct {
	PrivKey       [PrivateKeySize]byte
	PubKey        [PublicKeySize]byte
	prvK          *ecdh.PrivateKey
	PrivKeyBase64 string
	PubKeyBase64  string
	BriefName     string
}

func (s *SM2ECDH) SetPrivateKey(prk []byte) (err error) {
	copy(s.PrivKey[:], prk[:PrivateKeySize])
	s.prvK, err = ecdh.P256().NewPrivateKey(prk)
	if err != nil {
		return err
	}
	copy(s.PubKey[:], s.prvK.PublicKey().Bytes()[1:1+PublicKeySize])
	s.PrivKeyBase64 = base64.StdEncoding.EncodeToString(s.PrivKey[:])
	s.PubKeyBase64 = base64.StdEncoding.EncodeToString(s.PubKey[:])
	s.BriefName = fmt.Sprintf("%s...%s", s.PubKeyBase64[0:4], s.PubKeyBase64[39:43])

	return nil
}

func (s *SM2ECDH) PrivateKey() []byte {
	return s.PrivKey[:]
}

func (s *SM2ECDH) PrivateKeyBase64() string {
	return s.PrivKeyBase64
}

func (s *SM2ECDH) PublicKey() []byte {
	return s.PubKey[:]
}

func (s *SM2ECDH) PublicKeyBase64() string {
	return s.PubKeyBase64
}

func (s *SM2ECDH) SharedSecret(pbk []byte) []byte {
	if len(pbk) != PublicKeySize {
		return nil
	}

	var pkBytes [1 + PublicKeySize]byte
	pkBytes[0] = 4 // uncompressed header
	copy(pkBytes[1:1+PublicKeySize], pbk[:PublicKeySize])

	pubK, err := ecdh.P256().NewPublicKey(pkBytes[:])
	if err != nil {
		return nil
	}

	sk, err := s.prvK.ECDH(pubK)
	if err != nil {
		return nil
	}

	// 32 bytes key
	return sk[:]
}

func (s *SM2ECDH) Name() string {
	return s.BriefName
}

func (s *SM2ECDH) Identity() []byte {
	return nil
}

func (c *SM2ECDH) MidPublicKey() []byte {
	return nil
}

func NewECDH() *SM2ECDH {
	var err error
	var s SM2ECDH
	s.prvK, err = ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil
	}
	copy(s.PrivKey[:], s.prvK.Bytes())
	copy(s.PubKey[:], s.prvK.PublicKey().Bytes()[1:1+PublicKeySize])
	s.PrivKeyBase64 = base64.StdEncoding.EncodeToString(s.PrivKey[:])
	s.PubKeyBase64 = base64.StdEncoding.EncodeToString(s.PubKey[:])
	s.BriefName = fmt.Sprintf("%s...%s", s.PubKeyBase64[0:4], s.PubKeyBase64[39:43])
	return &s
}

// Generate SM2 public and private key pairs for ECDH.
func GenerateSM2ECDHKeypair() (string, string) {
	var err error
	var pubKey [64]byte
	var privKey [32]byte

	pKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return "", ""
	}
	copy(privKey[:32], pKey.Bytes()[:32])             // Private Key 32 bytes
	copy(pubKey[:64], pKey.PublicKey().Bytes()[1:65]) // Public Key 64 bytes

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

// Generate SM2 public and private key pairs for ECDSA.
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
