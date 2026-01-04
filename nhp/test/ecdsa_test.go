package test

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"

	core "github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/core/scheme/gmsm"
	"github.com/emmansun/gmsm/sm2"
)

func TestSM2ECDSAKeys(t *testing.T) {
	var err error
	var pubKey [64]byte
	var privKey [32]byte
	sm2PrivKey := new(sm2.PrivateKey)

	for {
		sm2PrivKey, err = sm2.GenerateKey(rand.Reader)
		if err == nil {
			break
		}
	}
	dBuf := sm2PrivKey.D.Bytes()
	fmt.Printf("Private key length: %d\n", len(dBuf))
	copy(privKey[:], dBuf[:])

	xBuf := sm2PrivKey.PublicKey.X.Bytes()
	yBuf := sm2PrivKey.PublicKey.Y.Bytes()
	fmt.Printf("Public key X length: %d\n", len(xBuf))
	fmt.Printf("Public key Y length: %d\n", len(yBuf))
	copy(pubKey[:len(xBuf)], xBuf[:])
	copy(pubKey[len(xBuf):], yBuf[:])

	pubStr := base64.StdEncoding.EncodeToString(pubKey[:])
	privStr := base64.StdEncoding.EncodeToString(privKey[:])

	fmt.Printf("Public key: %s\n", pubStr)
	fmt.Printf("Private key: %s\n", privStr)
}

func TestSM2ECDSAKeysSerializeAndDeserialize(t *testing.T) {
	var err error
	var pubKey [64]byte
	var privKey [32]byte
	sm2PrivKey := new(sm2.PrivateKey)

	for {
		sm2PrivKey, err = sm2.GenerateKey(rand.Reader)
		if err == nil {
			break
		}
	}
	dBuf := sm2PrivKey.D.Bytes()
	copy(privKey[:], dBuf[:])

	xBuf := sm2PrivKey.PublicKey.X.Bytes()
	yBuf := sm2PrivKey.PublicKey.Y.Bytes()
	copy(pubKey[:len(xBuf)], xBuf[:])
	copy(pubKey[len(xBuf):], yBuf[:])

	pubStr := base64.StdEncoding.EncodeToString(pubKey[:])
	privStr := base64.StdEncoding.EncodeToString(privKey[:])

	fmt.Printf("Public key: %s\n", pubStr)
	fmt.Printf("Private key: %s\n", privStr)
	fmt.Printf("Public key: %s\n", hex.EncodeToString(pubKey[:]))
	fmt.Printf("Private key: %s\n", hex.EncodeToString(privKey[:]))

	plainText := "This is a test for sm2 ecdsa keys serialization"
	hash, err := core.NewHash(core.HASH_SHA256)
	if err != nil {
		fmt.Printf("Failed to create hash: %v\n", err)
		return
	}
	hash.Write([]byte(plainText))
	hashedBytes := hash.Sum(nil)
	fmt.Printf("hashed hex: %s, length: %d\n", hex.EncodeToString(hashedBytes), len(hashedBytes))
	fmt.Printf("hashed byte slice: %v\n", hashedBytes)
	hashedBytes = []byte("AABBCCDD")

	signature, err := sm2PrivKey.Sign(rand.Reader, hashedBytes, nil)
	if err != nil {
		fmt.Printf("Sign error: %v\n", err)
		return
	}
	fmt.Printf("signature hex: %s, length: %d\n", hex.EncodeToString(signature), len(signature))

	success := sm2.VerifyASN1(&sm2PrivKey.PublicKey, hashedBytes[:], signature)
	fmt.Printf("Signature verification: %v\n", success)

	encrypted, err := sm2.Encrypt(rand.Reader, &sm2PrivKey.PublicKey, []byte(plainText), nil)
	if err != nil {
		fmt.Printf("encrypt error: %v\n", err)
		return
	}
	fmt.Printf("encrypted hex: %s, length: %d\n", hex.EncodeToString(encrypted), len(encrypted))
	decrypted, err := sm2.Decrypt(sm2PrivKey, encrypted)
	if err != nil {
		fmt.Printf("decrypt error: %v\n", err)
		return
	}
	fmt.Printf("Decypted plaintext: %s\n", decrypted)

	privKey1, err := gmsm.Base64DecodeSM2ECDSAPrivateKey(pubStr, privStr)
	if err != nil {
		fmt.Printf("decode priv key error: %v\n", err)
		return
	}

	decrypted1, err := sm2.Decrypt(privKey1, encrypted)
	if err != nil {
		fmt.Printf("decrypt1 error: %v\n", err)
		return
	}
	fmt.Printf("Decypted plaintext1: %s\n", decrypted1)

	pubKey1, err := gmsm.Base64DecodeSM2ECDSAPublicKey(pubStr)
	if err != nil {
		fmt.Printf("decode pub key error: %v\n", err)
		return
	}
	success1 := sm2.VerifyASN1(pubKey1, hashedBytes[:], signature)
	fmt.Printf("Signature verification1: %v\n", success1)
}

func TestKeyDeserialization(t *testing.T) {
	pubHexStr := "04627652fd978d0c4290d28e3233309f83e35ec2834fc5f9df2bfd32d658200a9bed7906bde12f4504ea04ca5f65eff1a9253cdcef6415a36999642e5395d9080f"
	privHexStr := "1045f665465c1d78663ee7719592667fe8ca65d3b24c62259a45208c7d2c04f4"

	pubBytes, _ := hex.DecodeString(pubHexStr)
	privBytes, _ := hex.DecodeString(privHexStr)

	privKey, err := gmsm.Base64DecodeSM2ECDSAPrivateKey(base64.StdEncoding.EncodeToString(pubBytes), base64.StdEncoding.EncodeToString(privBytes))
	if err != nil {
		fmt.Printf("decode failed %v\n", err)
		return
	}

	encrypted, err := sm2.Encrypt(rand.Reader, &privKey.PublicKey, []byte("PlainTextAABBCCDD"), nil)
	if err != nil {
		fmt.Printf("encrypt error: %v\n", err)
		return
	}
	fmt.Printf("encrypted hex: %s, length: %d\n", hex.EncodeToString(encrypted), len(encrypted))
	decrypted, err := sm2.Decrypt(privKey, encrypted)
	if err != nil {
		fmt.Printf("decrypt error: %v\n", err)
		return
	}
	fmt.Printf("decrypt ok: %s\n", decrypted)
}

func TestECDHForECDSA(t *testing.T) {
	pubStr, privStr := gmsm.GenerateSM2ECDHKeypair()
	fmt.Printf("Public key: %s\n", pubStr)
	fmt.Printf("Private key: %s\n", privStr)

	sm2PrivKey, err := gmsm.Base64DecodeSM2ECDSAPrivateKey(pubStr, privStr)
	if err != nil {
		fmt.Printf("decode error: %v\n", err)
		return
	}

	plainText := "Test ECDSA functions using ECDH keys"
	signBytes := []byte("AAbbCCdd")
	encrypted, err := sm2.Encrypt(rand.Reader, &sm2PrivKey.PublicKey, []byte(plainText), nil)
	if err != nil {
		fmt.Printf("decode error: %v\n", err)
		return
	}
	decrypted, err := sm2.Decrypt(sm2PrivKey, encrypted)
	if err != nil {
		fmt.Printf("decrypt error: %v\n", err)
		return
	}
	fmt.Printf("Decypted plaintext: %s\n", decrypted)

	signature, err := sm2PrivKey.Sign(rand.Reader, signBytes[:], nil)
	if err != nil {
		fmt.Printf("sign error: %v\n", err)
		return
	}
	fmt.Printf("signature hex: %s\n", hex.EncodeToString(signature))

	verified := sm2.VerifyASN1(&sm2PrivKey.PublicKey, signBytes, signature)
	fmt.Printf("signature verification: %v\n", verified)
}
