package utils

import (
	"fmt"
	"testing"
)

func TestAes(t *testing.T) {
	a, err := AesEncrypt("hello world")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(a)
}

func TestAesD(t *testing.T) {
	a, err := AesDecrypt("vO/CddwO9FPdb08aklxBiA==")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(a)
}

func TestGenerateACLicense(t *testing.T) {
	var msh = `{"ac":{"id":"6b5ce2e6-f8d8-d298-fa70-32cd51710fd6","private_key":"IHZyCkX2CFmvk6pISUI7RoEA31pYZrmKDzyjAa6qd0Q="},"server":{"host":"127.0.0.1:8081","public_key":"5/zceEX6OP38JRJMRIChWsQdyV9UgOKzdEahWwq7rDU="},"expired":1688633095000000,"create_time":1683362695000000}`

	ens, err := AesEncrypt(msh)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(ens)
}

func TestRSAKeys(t *testing.T) {
	privKeyStr, pubKeyStr := GenerateRsaKey(1000)

	fmt.Println("private key: ", privKeyStr)
	fmt.Println("private key length: ", len(privKeyStr))
	fmt.Println("public key: ", pubKeyStr)
	fmt.Println("public key length", len(pubKeyStr))
}
