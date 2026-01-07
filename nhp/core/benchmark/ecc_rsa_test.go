package benchmark

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
)

var aeadCount uint64 = 0

func TestRSASignAndVerify(t *testing.T) {
	msg := "Qt for Windows - Building from Source" +
		"This page describes the process of configuring and building Qt for Windows. To download" +
		" and install a pre-built Qt for Windows, follow the instructions on the Getting Started with Qt page."

	now := time.Now()

	for i := 0; i < 10; i++ {
		priv, pub := GenerateRSAKeys()
		hashed, signature, err := SignWithRSAPrivateKey(priv, []byte(msg))

		if err != nil {
			fmt.Printf("RSA sign error: %v", err)
			return
		}

		err = VerifyWithRSAPublicKey(pub, hashed, signature)
		if err != nil {
			fmt.Printf("RSA verify error: %v", err)
			return
		}
	}

	d := time.Since(now)
	fmt.Printf("RSA verify success with %d microseconds.\n", d.Microseconds())
}

func TestECCSharedKey(t *testing.T) {
	now := time.Now()

	msg := "Qt for Windows - Building from Source" +
		"This page describes the process of configuring and building Qt for Windows. To download" +
		" and install a pre-built Qt for Windows, follow the instructions on the Getting Started with Qt page."

	for i := 0; i < 10; i++ {
		ecdh := core.NewECDH(core.ECC_CURVE25519)
		ecdhr := core.NewECDH(core.ECC_CURVE25519)

		ssc := ecdh.SharedSecret(ecdhr.PublicKey())
		sss := ecdhr.SharedSecret(ecdh.PublicKey())

		//if !bytes.Equal(ssc[:], sss[:]) {
		//	fmt.Printf("shared key is not identical, quit")
		//	return
		//}

		var sscKey, sssKey [core.SymmetricKeySize]byte
		copy(sscKey[:], ssc[:])
		copy(sssKey[:], sss[:])

		hashc := sha256.New()
		hashc.Write(ssc[:])
		hashedc := hashc.Sum(nil)

		hashs := sha256.New()
		hashs.Write(ssc[:])
		hasheds := hashs.Sum(nil)

		aeadc, err := core.AeadFromKey(core.GCM_AES256, &sscKey)
		if err != nil {
			fmt.Printf("aead creation error: %v", err)
			return
		}
		aeads, err := core.AeadFromKey(core.GCM_AES256, &sssKey)
		if err != nil {
			fmt.Printf("aead creation error: %v", err)
			return
		}

		var nonceBytes [12]byte
		aeadCount++
		binary.BigEndian.PutUint64(nonceBytes[:], aeadCount)

		encrypted := aeadc.Seal(nil, nonceBytes[:], []byte(msg), hashedc)
		decrypted, err := aeads.Open(nil, nonceBytes[:], encrypted, hasheds)
		_ = decrypted
		if err != nil {
			fmt.Printf("aead decrypt error: %v", err)
			return
		}
	}

	d := time.Since(now)
	//fmt.Printf("Decrypted message:\n%s\n", string(decrypted))
	fmt.Printf("ECC verify success with %d microseconds.\n", d.Microseconds())
}

func TestGMSharedKey(t *testing.T) {
	now := time.Now()

	msg := "Qt for Windows - Building from Source" +
		"This page describes the process of configuring and building Qt for Windows. To download" +
		" and install a pre-built Qt for Windows, follow the instructions on the Getting Started with Qt page."

	for i := 0; i < 10; i++ {
		ecdh := core.NewECDH(core.ECC_SM2)
		ecdhr := core.NewECDH(core.ECC_SM2)

		ssc := ecdh.SharedSecret(ecdhr.PublicKey())
		sss := ecdhr.SharedSecret(ecdh.PublicKey())

		//if !bytes.Equal(ssc[:], sss[:]) {
		//	fmt.Printf("shared key is not identical, quit")
		//	return
		//}

		var sscKey, sssKey [core.SymmetricKeySize]byte
		copy(sscKey[:], ssc[:])
		copy(sssKey[:], sss[:])

		hashc := sha256.New()
		hashc.Write(ssc[:])
		hashedc := hashc.Sum(nil)

		hashs := sha256.New()
		hashs.Write(ssc[:])
		hasheds := hashs.Sum(nil)

		aeadc, err := core.AeadFromKey(core.GCM_SM4, &sscKey)
		if err != nil {
			fmt.Printf("aead creation error: %v", err)
			return
		}
		aeads, err := core.AeadFromKey(core.GCM_SM4, &sssKey)
		if err != nil {
			fmt.Printf("aead creation error: %v", err)
			return
		}

		var nonceBytes [12]byte
		aeadCount++
		binary.BigEndian.PutUint64(nonceBytes[:], aeadCount)

		encrypted := aeadc.Seal(nil, nonceBytes[:], []byte(msg), hashedc)
		decrypted, err := aeads.Open(nil, nonceBytes[:], encrypted, hasheds)
		_ = decrypted
		if err != nil {
			fmt.Printf("aead decrypt error: %v", err)
			return
		}
	}

	d := time.Since(now)
	//fmt.Printf("Decrypted message:\n%s\n", string(decrypted))
	fmt.Printf("ECC verify success with %d microseconds.\n", d.Microseconds())
}
