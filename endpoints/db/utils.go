package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	ztdolib "github.com/OpenNHP/opennhp/nhp/core/ztdo"
	"github.com/OpenNHP/opennhp/nhp/utils"
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

	fullPath := filepath.Join(common.ExeDirPath, etcDir, fileName)

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
	fullPath := filepath.Join(common.ExeDirPath, etcDir, fileName)
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
	fullPath := filepath.Join(common.ExeDirPath, etcDir, fileName)

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

func (a *AppParams) NewSmartPolicy() (common.SmartPolicy, error) {
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

	spoId, err := utils.GenerateUUIDv4()
	if err != nil {
		return common.SmartPolicy{}, fmt.Errorf("error generating spoId: %v", err)
	}

	config.PolicyId = spoId

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

func (a *AppParams) LoadMetadataAsStruct() (map[string]any, error) {
	var metadata map[string]any

	if a.Metadata == "" {
		return metadata, nil
	}

	content, err := os.ReadFile(a.Metadata)
	if err != nil {
		return metadata, nil
	}

	err = json.Unmarshal(content, &metadata)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}

func (a *UdpDevice) UploadFileToNHPServer(filePath string) (string, error) {
	httpHost := fmt.Sprintf("http://%s/", a.GetServerPeer().Host())
	testReq, err := http.Get(httpHost) //nolint:gosec // G107: URL constructed from configured server peer
	if err != nil {
		return "", err
	}

	if testReq.StatusCode == http.StatusBadRequest {
		httpHost = fmt.Sprintf("https://%s/", a.GetServerPeer().Host())
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("could not get file info: %v", err)
	}

	// create upload progress
	progress := &UploadProgress{
		TotalSize: fileInfo.Size(),
	}

	startTime := time.Now()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", fmt.Errorf("could not create form file: %v", err)
	}

	progressReader := &ProgressReader{
		Reader:   file,
		Progress: progress,
	}

	_, err = io.Copy(part, progressReader)
	if err != nil {
		return "", fmt.Errorf("could not copy file to server: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("could not close writer: %v", err)
	}

	uploadUrl := httpHost + "storage/upload"

	req, err := http.NewRequest("POST", uploadUrl, body)
	if err != nil {
		return "", fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Timeout: 120 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not send https request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf("unexpected status code: %d, content: %s", resp.StatusCode, string(bodyBytes))
	}

	// read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response body: %v", err)
	}

	// parse response body
	var respBody ServerResponse

	err = json.Unmarshal(bodyBytes, &respBody)
	if err != nil {
		return "", fmt.Errorf("could not parse response body: %v", err)
	}

	duration := time.Since(startTime)
	speed := float64(progress.TotalSize) / duration.Seconds() / (1024 * 1024)

	// change the chinese to english
	fmt.Printf("\nUpload %s to %s success! (time: %.2fs, speed: %.2fMB/s)\n",
		filePath, httpHost+respBody.FileURI, duration.Seconds(), speed)

	return httpHost + respBody.FileURI, nil
}

type UploadProgress struct {
	TotalSize   int64
	BytesRead   int64
	UploadID    string
	LastPercent int
}

type ServerResponse struct {
	Message string `json:"message"`
	FileURI string `json:"file_uri"`
	UUID    string `json:"uuid"`
	MD5     string `json:"md5"`
}

type ProgressReader struct {
	Reader   io.Reader
	Progress *UploadProgress
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	if err == nil {
		pr.Progress.BytesRead += int64(n)

		// calculate percent of upload progress
		percent := int(float64(pr.Progress.BytesRead) / float64(pr.Progress.TotalSize) * 100)
		if percent > pr.Progress.LastPercent && percent%5 == 0 {
			pr.Progress.LastPercent = percent
			pr.displayProgress(percent)
		}
	}
	return
}

// displayProgress
func (pr *ProgressReader) displayProgress(percent int) {
	// barLength is the length of progress bar
	const barLength = 50
	completed := int(float64(barLength) * float64(percent) / 100)

	// create progress bar string
	bar := make([]byte, barLength)
	for i := 0; i < barLength; i++ {
		if i < completed {
			bar[i] = '='
		} else if i == completed {
			bar[i] = '>'
		} else {
			bar[i] = ' '
		}
	}

	// calculate uploaded MB and total MB
	uploadedMB := float64(pr.Progress.BytesRead) / (1024 * 1024)
	totalMB := float64(pr.Progress.TotalSize) / (1024 * 1024)

	fmt.Printf("\r[%s] %3d%%  %.2f/%.2f MB",
		string(bar), percent, uploadedMB, totalMB)
}
