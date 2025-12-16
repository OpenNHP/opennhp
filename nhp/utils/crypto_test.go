package utils

import (
	"fmt"
	"testing"
)

func TestRSAKeys(t *testing.T) {
	privKeyStr, pubKeyStr := GenerateRsaKey(1000)

	fmt.Println("private key: ", privKeyStr)
	fmt.Println("private key length: ", len(privKeyStr))
	fmt.Println("public key: ", pubKeyStr)
	fmt.Println("public key length", len(pubKeyStr))
}
