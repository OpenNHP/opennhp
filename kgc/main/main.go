package main

import (
	_ "crypto/rand"
	"fmt"
	"log"

	"github.com/OpenNHP/opennhp/kgc" // 替换为你的模块路径
	_ "github.com/emmansun/gmsm/sm2"
)

func main() {
	// 生成 KGC 的主密钥对
	kgcPrivateKey, kgcPublicKey, err := kgc.GenerateMasterKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate master key pair: %v", err)
	}

	// 输入用户邮箱
	userEmail := "user@example.com"
	entlenA := 16 // 根据需要设置 entlenA

	// 生成用户密钥对
	dA, waPublicKey, tA := kgc.GenerateUserKey(entlenA, kgcPublicKey, kgcPrivateKey, userEmail)
	if dA == nil {
		log.Fatalf("Failed to generate user key")
	}

	// 输出生成的密钥信息
	fmt.Printf("User Email: %s\n", userEmail)
	fmt.Printf("Generated dA: %s\n", dA.String())
	fmt.Printf("Generated WA: (%s, %s)\n", waPublicKey.X.String(), waPublicKey.Y.String())
	fmt.Printf("Generated tA: %s\n", tA.String())

	// 验证密钥对的正确性
	isValid := kgc.VerifyKeyPair(dA, waPublicKey, userEmail, entlenA, kgcPublicKey, waPublicKey, tA)
	if isValid {
		fmt.Println("Key pair verification succeeded.")
	} else {
		fmt.Println("Key pair verification failed.")
	}
}
