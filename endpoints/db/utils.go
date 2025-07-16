package db

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	ztdolib "github.com/OpenNHP/opennhp/nhp/core/ztdo"
)

type DataPrivateKeyStore struct {
	DataPrivateKeyBase64    string `json:"dataPrivateKeyBase64"`
	ProviderPublicKeyBase64 string `json:"providerPublicKeyBase64"`
}

// NewDataPrivateKeyStore create a new DataPrivateKeyStore
func NewDataPrivateKeyStore(providerPublicKeyBase64 string) *DataPrivateKeyStore {
	return &DataPrivateKeyStore{
		ProviderPublicKeyBase64: providerPublicKeyBase64,
	}
}

// NewDataPrivateKeyStoreWith create a new DataPrivateKeyStore with doId
func NewDataPrivateKeyStoreWith(doId string) (d *DataPrivateKeyStore, err error) {
	etcDir := "etc/ztdo"
	fileName := "data-key-" + doId + ".json"

	fullPath := filepath.Join(etcDir, fileName)

	// open and read all the content in file
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	fileContentByte, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	d = &DataPrivateKeyStore{}
	d.fromJson(fileContentByte)

	return
}

func (d *DataPrivateKeyStore) Generate(mode ztdolib.DataKeyPairECCMode) (privateKey []byte) {
	ecdh := core.NewECDH(mode.ToEccType())
	d.DataPrivateKeyBase64 = ecdh.PrivateKeyBase64()
	return ecdh.PrivateKey()
}

// Save saves the dataPrivateKeyBase64 to a file, the format of file name is data-<doId>.json
// Notes: this default way to store data private key is not safe. In the wild environment, need to use a secure way to store data private key.
func (d *DataPrivateKeyStore) Save(doId string) error {
	// Make sure the etc directory exists
	etcDir := "etc/ztdo"
	if err := os.MkdirAll(etcDir, 0755); err != nil {
		return fmt.Errorf("failed to create etc directory: %v", err)
	}

	fileName := "data-key-" + doId + ".json"
	fullPath := filepath.Join(etcDir, fileName)
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("%v already exists, please delete it first", fullPath)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	file.Write(d.toJson())

	return nil
}

func (d *DataPrivateKeyStore) Delete(doId string) error {
	etcDir := "etc/ztdo"
	fileName := "data-key-" + doId + ".json"
	fullPath := filepath.Join(etcDir, fileName)

	// delete the file
	err := os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}
	return nil
}

func (d *DataPrivateKeyStore) toJson() []byte {
	dataPrkStoreJson, err := json.Marshal(d)
	if err != nil {
		return []byte("{}")
	} else {
		return dataPrkStoreJson
	}
}

func (d *DataPrivateKeyStore) fromJson(jsonData []byte) error {
	err := json.Unmarshal(jsonData, d)
	if err != nil {
		return fmt.Errorf("json parsing error: %s", err)
	}
	return nil
}

type AppParams struct {
	Mode                    string // the mode of operation: none, encrypt and decrypt
	Source                  string // the path of plaintext data
	DsType                  string // the type of data source: stream, online and offline
	Output                  string // path of output file
	SmartPolicy             string // path of smart policy
	Metadata                string // path of metadata
	ZtdoFilePath            string // path of ztdo file when mode is decrypt
	ZtdoId                  string // identifier of ztdo file
	DataPrivateKeyBase64    string
	AccessUrl               string // path of access url of ztdo
	ProviderPublicKeyBase64 string
}

func (a *AppParams) GetSmartPolicy() (common.SmartPolicy, error) {
	file, err := os.Open(a.SmartPolicy)
	if err != nil {
		return common.SmartPolicy{}, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	fileContentByte, err := io.ReadAll(file)
	if err != nil {
		return common.SmartPolicy{}, fmt.Errorf("error reading file: %v", err)
	}

	var config common.SmartPolicy

	err = json.Unmarshal(fileContentByte, &config)
	if err != nil {
		return common.SmartPolicy{}, fmt.Errorf("json parsing error: %s", err)
	}
	return config, nil
}

func (a *AppParams) GetMetadata() (string, error) {
	if a.Metadata == "" {
		return "", nil
	}

	content, err := os.ReadFile(a.Metadata)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
