package benchmark

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

func GenerateRSAKeys() (priv *rsa.PrivateKey, pub *rsa.PublicKey) {
	var err error
	priv, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Failed to generate RSA private key: %v", err)
		return
	}

	pub = &priv.PublicKey

	/*
		privBytes := x509.MarshalPKCS1PrivateKey(priv)
		pubBytes := x509.MarshalPKCS1PublicKey(pub)

		privBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privBytes,
		}
		pubBlock := &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: pubBytes,
		}
		privPemBytes := pem.EncodeToMemory(privBlock)
		pubPemBytes := pem.EncodeToMemory(pubBlock)

		fmt.Printf("Generated RSA private key:\n%s\n", string(privPemBytes))
		fmt.Printf("Generated RSA public key:\n%s\n", string(pubPemBytes))
	*/

	return
}

func SignWithRSAPrivateKey(priv *rsa.PrivateKey, msg []byte) (hashed []byte, signature []byte, err error) {
	hash := sha256.New()
	hash.Write(msg)
	hashed = hash.Sum(nil)
	signature, err = rsa.SignPKCS1v15(nil, priv, crypto.SHA256, hashed)

	return
}

func VerifyWithRSAPublicKey(pub *rsa.PublicKey, hashed []byte, signature []byte) error {
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed, signature)
}
