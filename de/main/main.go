package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/OpenNHP/opennhp/core"
)

// The Config structure is used to represent the structure of the configuration file
type Config struct {
	EncryptionKey string `json:"EncryptionKey"`
}

// Generates a 256-bit random key and returns a Base64-encoded string
func GenerateKey() (string, error) {
	key := make([]byte, 32) 
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate key: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

//Decode the Base64 encoded key into a byte array
func DecodeKey(encodedKey string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encodedKey)
}

// WriteConfig writes to etc/config.json file
func WriteConfig(key string) error {
	config := Config{
		EncryptionKey: key,
	}

	// Make sure the etc directory exists
	etcDir := "etc"
	if err := os.MkdirAll(etcDir, 0755); err != nil {
		return fmt.Errorf("failed to create etc directory: %v", err)
	}

	// Check if etc/config.json already exists
	configPath := filepath.Join(etcDir, "config.json")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config.json already exists, please delete it first")
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

func main() {
	// Generate a key
	key, err := GenerateKey()
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		return
	}

	// Write the key to config.json
	if err := WriteConfig(key); err != nil {
		fmt.Printf("Error writing config: %v\n", err)
		return
	}

	fmt.Println("Config file created successfully with generated key.")

	// Decode the key for encryption
	decodedKey, err := DecodeKey(key)
	if err != nil {
		fmt.Printf("Error decoding key: %v\n", err)
		return
	}

	// Reading .txt files
	ztdo, err := core.WriteSourceFile("F:\\git-document\\opennhp\\de\\example.txt", decodedKey) 
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Creating .ztdo Output Files
	outputFile, err := os.Create("output.ztdo")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer outputFile.Close()

	// Write Ztdo data to a .ztdo file
	if err := core.WriteZtdo(outputFile, ztdo); err != nil {
		fmt.Printf("Error writing Ztdo Data: %v\n", err)
	} else {
		fmt.Println("Ztdo Data written successfully to output.ztdo.")
	}
}
