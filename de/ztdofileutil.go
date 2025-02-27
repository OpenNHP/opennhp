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

// File
/**
sourceFilePath:
targetZtdoFilePath:加密的zdto文件路径
*/
func EncodeToZtoFile(sourceFilePath string, targetZtdoFilePath string) (zoId string, encryptionKey string) {
	// Generate Key
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

	encryptionKey = key

	fmt.Println("Config file created successfully with generated key.")

	// Creating .ztdo Output Files
	//directory path of the file
	targetZtdoDirPath := filepath.Dir(targetZtdoFilePath)
	//Auto-create directory
	err = os.MkdirAll(targetZtdoDirPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating targetDir:%v\n", err)
		return "", ""
	}
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

func DecodeZtoFile(ztdofileName string, decodeBase64Key string, newFileSaveDir string) {

	// //read ztdo file
	ztdoFile, err := readZtoFile(ztdofileName)
	if err != nil {
		fmt.Printf("Error readZtoFile: %v\n", err)
		return
	}

	decodeKey, err := DecodeKey(decodeBase64Key)
	if err != nil {
		fmt.Printf("Error DecodeKey: %v\n", err)
		return
	}
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
	defer file.Close()

	//
	writer := bufio.NewWriter(file)
	fileContent, err := core.Decrypt(ztdoFile.FileContent, decodeKey)
	// write bytes
	_, err = writer.Write(fileContent)
	if err != nil {
		fmt.Printf("error writing to file:%v \n", err)
		return
	}
	err = writer.Flush()
	if err != nil {
		fmt.Printf("error flushing writer:%v \n", err)
		return
	}
	fmt.Println("data written successfully.")
	fmt.Println("decodeZtoFile finished.")
}

// read config.json
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

	err = json.Unmarshal(fileContentByte, &config)
	if err != nil {
		return ZtdoConfig{}, fmt.Errorf("json parsing error: %s", err)
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

	err = json.Unmarshal(fileContentByte, &ztdoFile)
	if err != nil {
		return core.ZtdoFile{}, fmt.Errorf("json parsing error: %s", err)
	}
	return ztdoFile, nil
}

// read Polic file
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

	err = json.Unmarshal(fileContentByte, &config)
	if err != nil {
		return common.DHPPolicy{}, fmt.Errorf("json parsing error: %s", err)
	}
	return config, nil
}
