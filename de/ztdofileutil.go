package de

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/core"
	"github.com/OpenNHP/opennhp/utils"
)

// The Config structure is used to represent the structure of the configuration file
type ZtdoConfig struct {
	EncryptionKey string `json:"encryptionKey"`
}

// Generates a 256-bit random key and returns a Base64-encoded string
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate key: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// Decode the Base64 encoded key into a byte array
func DecodeKey(encodedKey string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedKey)
}

// WriteConfig writes to etc/config.json file
func WriteConfig(key string, objectId string) error {
	config := ZtdoConfig{
		EncryptionKey: key,
	}

	// Make sure the etc directory exists
	etcDir := "etc/" + utils.GetCurrentDate()
	if err := os.MkdirAll(etcDir, 0755); err != nil {
		return fmt.Errorf("failed to create etc directory: %v", err)
	}

	// Check if etc/config.json already exists
	configFileName := "config-" + objectId + ".json"
	configPath := filepath.Join(etcDir, configFileName)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("%v already exists, please delete it first", configFileName)
	}

	// Create etc/config.json file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config.json: %v", err)
	}
	defer file.Close()

	// Write data to config.json
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

// 文件加密
/**
sourceFilePath:源文件路径
targetZtdoFilePath:加密的zdto文件路径
*/
func EncodeToZtoFile(sourceFilePath string, targetZtdoFilePath string) (zoId string, encryptionKey string) {
	// 生成数据加密码
	key, err := GenerateKey()
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		return "", ""
	}

	// Decode the key for encryption
	decodedKey, err := DecodeKey(key)
	if err != nil {
		fmt.Printf("Error decoding key: %v\n", err)
		return "", ""
	}

	// Reading .txt files
	ztdo, err := core.WriteSourceFile(sourceFilePath, decodedKey)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return "", ""
	}

	//将目标用户的公钥信息、文件创建日期及其他策略信息生成到NHP服务的配置文件中，用于以后文件解密时进行校验及解密内容
	// if err := WriteConfig(key, ztdo.Header.Objectid); err != nil {
	// 	fmt.Printf("Error writing config: %v\n", err)
	// 	return
	// }
	//生数据加密键值
	encryptionKey = key //测试时先明文加Base64加密

	fmt.Println("Config file created successfully with generated key.")

	// Creating .ztdo Output Files
	//获取文件的目录路径
	targetZtdoDirPath := filepath.Dir(targetZtdoFilePath)
	//自动创建目录
	err = os.MkdirAll(targetZtdoDirPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating targetDir:%v\n", err)
		return "", ""
	}
	//创建文件
	outputFile, err := os.Create(targetZtdoFilePath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return "", ""
	}
	defer outputFile.Close()

	// Write Ztdo data to a .ztdo file
	if err := core.WriteZtdo(outputFile, ztdo); err != nil {
		fmt.Printf("Error writing Ztdo Data: %v\n", err)
		return "", ""
	} else {
		fmt.Printf("Ztdo Data written successfully to output.ztdo.%v", ztdo.Header.Objectid)
	}
	return ztdo.Header.Objectid, encryptionKey

}

func DecodeZtoFile(ztofileName string, decodeBase64Key string, newFileSaveDir string) {

	// //读取.ztdo文件内容到ZtoFile对象中
	ztdoFile, err := readZtoFile(ztofileName)
	if err != nil {
		fmt.Printf("Error readZtoFile: %v\n", err)
		return
	}

	// config, err := readConfig("E:\\work\\project\\dhp-hygon-arch\\etc\\" + ztdoFile.CreateDate + "\\config-" + ztdoFile.Objectid + ".json")
	// if err != nil {
	// 	fmt.Printf("Error readConfig file: %v\n", err)
	// 	return
	// }
	//从config.json中获取文件加密的密钥
	decodeKey, err := DecodeKey(decodeBase64Key)
	if err != nil {
		fmt.Printf("Error DecodeKey: %v\n", err)
		return
	}
	//还原后的新文件地址
	newDirPath := newFileSaveDir
	if err := os.MkdirAll(newDirPath, 0755); err != nil {
		fmt.Printf("failed to create %v directory: %v", newDirPath, err)
		return
	}
	// Check if etc/config.json already exists
	newFilePath := filepath.Join(newDirPath, ztdoFile.FileName)
	if _, err := os.Stat(newFilePath); err == nil {
		fmt.Printf("%v already exists, please delete it first", newFilePath)
		return
	}

	file, err := os.Create(newFilePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close() // 确保最后关闭文件

	// 创建bufio的writer
	writer := bufio.NewWriter(file)
	fileContent, err := core.Decrypt(ztdoFile.FileContent, decodeKey)
	// fileContent := ztdoFile.FileContent
	// 写入字节数据
	_, err = writer.Write(fileContent)
	if err != nil {
		fmt.Printf("Error writing to file:%v \n", err)
		return
	}

	// 刷新缓冲区，确保数据被写入文件
	err = writer.Flush()
	if err != nil {
		fmt.Printf("Error flushing writer:%v \n", err)
		return
	}
	fmt.Println("Data written successfully.")
	fmt.Println("decodeZtoFile finished.")
}

// 读取config.json文件
func readConfig(filePath string) (ZtdoConfig, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return ZtdoConfig{}, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	fileContentByte, err := io.ReadAll(file)
	if err != nil {
		return ZtdoConfig{}, fmt.Errorf("error reading file: %v", err)
	}

	var config ZtdoConfig

	// 使用json.Unmarshal将JSON字符串解码到person实例中
	err = json.Unmarshal(fileContentByte, &config)
	if err != nil {
		return ZtdoConfig{}, fmt.Errorf("JSON解析错误: %s", err)
	}
	return config, nil
}

func readZtoFile(filePath string) (core.ZtdoFile, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return core.ZtdoFile{}, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	fileContentByte, err := io.ReadAll(file)
	if err != nil {
		return core.ZtdoFile{}, fmt.Errorf("error reading file: %v", err)
	}

	var ztdoFile core.ZtdoFile
	// 使用json.Unmarshal将JSON字符串解码到person实例中
	err = json.Unmarshal(fileContentByte, &ztdoFile)
	if err != nil {
		return core.ZtdoFile{}, fmt.Errorf("JSON解析错误: %s", err)
	}
	return ztdoFile, nil
}

// DHPPolicy配置文件读取
func ReadPolicyFile(filePath string) (common.DHPPolicy, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return common.DHPPolicy{}, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	fileContentByte, err := io.ReadAll(file)
	if err != nil {
		return common.DHPPolicy{}, fmt.Errorf("error reading file: %v", err)
	}

	var config common.DHPPolicy

	// 使用json.Unmarshal将JSON字符串解码到person实例中
	err = json.Unmarshal(fileContentByte, &config)
	if err != nil {
		return common.DHPPolicy{}, fmt.Errorf("JSON解析错误: %s", err)
	}
	return config, nil
}
