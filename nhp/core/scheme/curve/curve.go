package curve

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

const (
	PrivateKeySize = 32
	PublicKeySize  = 32
)

type Curve25519ECDH struct {
	PrivKey       [PrivateKeySize]byte
	PubKey        [PublicKeySize]byte
	PrivKeyBase64 string
	PubKeyBase64  string
	BriefName     string
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
	c.BriefName = fmt.Sprintf("%s...%s", c.PubKeyBase64[0:4], c.PubKeyBase64[39:43])

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

func (c *Curve25519ECDH) SharedSecret(pbk []byte) []byte {
	if len(pbk) != PublicKeySize {
		return nil
	}

	sk, err := curve25519.X25519(c.PrivKey[:], pbk)
	if err != nil {
		return nil
	}

	// 32 bytes key
	return sk[:]
}

func (c *Curve25519ECDH) Name() string {
	return c.BriefName
}

func (c *Curve25519ECDH) Identity() []byte {
	return nil
}

func (c *Curve25519ECDH) MidPublicKey() []byte {
	return nil
}

func NewECDH() *Curve25519ECDH {
	key := make([]byte, PrivateKeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil
	}
	// clamp
	key[0] &= 248
	key[31] = (key[31] & 127) | 64
	var c Curve25519ECDH
	c.SetPrivateKey(key)

	return &c
}
