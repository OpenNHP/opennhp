package benchmark

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/core"
	"github.com/OpenNHP/opennhp/core/scheme/gmsm"
	"github.com/emmansun/gmsm/sm2"
)

var aeadCount uint64 = 0

func TestRSASignAndVerify(t *testing.T) {
	// msg := "Qt for Windows - Building from Source" +
	// 	"This page describes the process of configuring and building Qt for Windows. To download" +
	// 	" and install a pre-built Qt for Windows, follow the instructions on the Getting Started with Qt page."
	msg := "helloworld"
	now := time.Now()

	for i := 0; i < 10; i++ {
		priv, pub := GenerateRSAKeys()
		hashed, signature, err := SignWithRSAPrivateKey(priv, []byte(msg))
		fmt.Println("hashed:" + string(hashed) + ";signature" + string(signature))
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

		aeadc := core.AeadFromKey(core.GCM_AES256, &sscKey)
		aeads := core.AeadFromKey(core.GCM_AES256, &sssKey)

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

		aeadc := core.AeadFromKey(core.GCM_SM4, &sscKey)
		aeads := core.AeadFromKey(core.GCM_SM4, &sssKey)

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

func TestECCEncryption(t *testing.T) {
	now := time.Now()
	d := time.Since(now)

	pubKey := "048356e642a40ebd18d29ba3532fbd9f3bbee8f027c3f6f39a5ba2f870369f9988981f5efe55d1c5cdf6c0ef2b070847a14f7fdf4272a8df09c442f3058af94ba1"
	encodeMsg, err := core.ECCEncryption(pubKey, "hellotzy")
	if err != nil {
		fmt.Printf("加密失败:%v\n", err)
		return
	}
	fmt.Printf("ECCEncryption success with %d microseconds. encodeMsg:%s\n", d.Microseconds(), encodeMsg)

	//解密
	privateKey := "6c5a0a0b2eed3cbec3e4f1252bfe0e28c504a1c6bf1999eebb0af9ef0f8e6c85"
	encodeMessage := encodeMsg
	message, err := core.ECCDecrypt(privateKey, encodeMessage)
	if err != nil {
		fmt.Printf("解密失败:%v\n", err)
		return
	}
	fmt.Printf("ECCDecrypt success with %d microseconds. encodeMsg:%s\n", d.Microseconds(), message)
}

func TestECCDecrypt(t *testing.T) {
	now := time.Now()
	d := time.Since(now)

	// // 生成密钥对
	// priKey, err := sm2.GenerateKey(rand.Reader)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// pubKey := &priKey.PublicKey

	//Private key: +Jnee2lP6Kn47qzSaqwSmWxORsBkkCV6YHsRqXCegVo=
	// Public key:  iOqqqM3bfyFICgN3pbuSXrdp1++qEvo6NEGN3hrBWUat/ysLorrg+NJuIIylpnpFaUInDAxkDmmSThnixftiqA==
	pubKeyString := "iOqqqM3bfyFICgN3pbuSXrdp1++qEvo6NEGN3hrBWUat/ysLorrg+NJuIIylpnpFaUInDAxkDmmSThnixftiqA=="
	pubKey, err := gmsm.Base64DecodeSM2ECDSAPublicKey(pubKeyString)
	if err != nil {
		fmt.Printf("公钥加载失败：%v \n", err)
		return
	}
	// 明文消息
	message := "Hello, world!"
	// 加密
	cipher, err := sm2.EncryptASN1(rand.Reader, pubKey, []byte(message))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("加密后的密文: %s\n", hex.EncodeToString(cipher))
	//
	privateKey := "+Jnee2lP6Kn47qzSaqwSmWxORsBkkCV6YHsRqXCegVo="

	pubBytes, _ := hex.DecodeString(pubKeyString)
	privBytes, _ := hex.DecodeString(privateKey)

	priKey, err := gmsm.Base64DecodeSM2ECDSAPrivateKey(base64.StdEncoding.EncodeToString(pubBytes), base64.StdEncoding.EncodeToString(privBytes))
	// priKey, err := gmsm.Base64DecodeSM2ECDHPrivateKey(privateKey)
	if err != nil {
		fmt.Printf("私钥加载失败：%v \n", err)
		return
	}
	// 解密
	plain, err := sm2.Decrypt(priKey, cipher)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("解密后的明文: %s\n", string(plain))

	fmt.Printf("ECCEncryption success with %d microseconds. encodeMsg:%s", d.Microseconds(), message)
}
