package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/OpenNHP/opennhp/core"
)

// Config 结构体用来表示配置文件的结构
type Config struct {
	EncryptionKey string `json:"EncryptionKey"`
}

// 生成 256 位随机密钥，并返回 Base64 编码的字符串
func GenerateKey() (string, error) {
	key := make([]byte, 32) // 256位密钥
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate key: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// 解码 Base64 编码的密钥为字节数组
func DecodeKey(encodedKey string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedKey)
}

// WriteConfig 写入 config.json 文件
func WriteConfig(key string) error {
	config := Config{
		EncryptionKey: key,
	}

	// 检查 config.json 是否已存在
	if _, err := os.Stat("config.json"); err == nil {
		return fmt.Errorf("config.json already exists, please delete it first")
	}

	// 创建 config.json 文件
	file, err := os.Create("config.json")
	if err != nil {
		return fmt.Errorf("failed to create config.json: %v", err)
	}
	defer file.Close()

	// 写入数据到 config.json
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // 设置缩进格式
	return encoder.Encode(config)
}

func main() {
	// 生成密钥
	key, err := GenerateKey()
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		return
	}

	// 写入密钥到 config.json
	if err := WriteConfig(key); err != nil {
		fmt.Printf("Error writing config: %v\n", err)
		return
	}

	fmt.Println("Config file created successfully with generated key.")

	// 解码密钥以用于加密
	decodedKey, err := DecodeKey(key)
	if err != nil {
		fmt.Printf("Error decoding key: %v\n", err)
		return
	}

	// 读取 .txt 文件
	ztdo, err := core.ParseZtdoFromFile("F:\\git-document\\opennhp\\de\\example.txt", decodedKey) // 输入您的 txt 文件路径
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// 创建 .tdf 输出文件
	outputFile, err := os.Create("output.tdf") // 输出文件名设为 output.tdf
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer outputFile.Close()

	// 写入 Ztdo 数据到 .tdf 文件
	if err := core.WriteZtdo(outputFile, ztdo); err != nil {
		fmt.Printf("Error writing Ztdo Data: %v\n", err)
	} else {
		fmt.Println("Ztdo Data written successfully to output.tdf.")
	}
}
