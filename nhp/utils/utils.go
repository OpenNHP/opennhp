package utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// GetRandomUint32 returns a uniformly random non-zero uint32 from the system
// CSPRNG. Zero is excluded so callers can use the result directly as an XOR
// mask without a degenerate all-zero preamble.
func GetRandomUint32() uint32 {
	var b [4]byte
	for {
		if _, err := rand.Read(b[:]); err != nil {
			panic(fmt.Sprintf("utils.GetRandomUint32: crypto/rand failed: %v", err))
		}
		if value := binary.BigEndian.Uint32(b[:]); value != 0 {
			return value
		}
	}
}

func CatchPanic() {
	if x := recover(); x != nil {
		for _, line := range append([]string{fmt.Sprint(x)}, strings.Split(string(debug.Stack()), "\n")...) {
			if len(strings.TrimSpace(line)) > 0 {
				log.Error("%s", line)
			}
		}
	}
}

func CatchPanicThenRun(catchFun func()) {
	if x := recover(); x != nil {
		for _, line := range append([]string{fmt.Sprint(x)}, strings.Split(string(debug.Stack()), "\n")...) {
			if len(strings.TrimSpace(line)) > 0 {
				log.Error("%s", line)
			}
		}
		if catchFun != nil {
			catchFun()
		}
	}
}

// Here's how to get the current date string in the format yyyyMMdd (like 20250210) in various programming languages:
func GetCurrentDate() (date string) {
	now := time.Now()
	date = now.Format("20060102")
	return date
}

func DownloadFileToTemp(fileUrl string, pattern string) (string, error) {
	tempDir, err := os.MkdirTemp("", pattern)
	if err != nil {
		return "", err
	}

	fileName := filepath.Base(fileUrl)
	tempFilePath := filepath.Join(tempDir, fileName)

	outFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	resp, err := http.Get(fileUrl) //nolint:gosec // G107: URL comes from trusted configuration
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file (%s): status code %s", fileUrl, resp.Status)
	}

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFilePath, nil
}

func GenerateTempFilePath(pattern string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}

	tempPath := file.Name()

	if err := file.Close(); err != nil {
		return "", err
	}

	return tempPath, nil
}

func SaveStructAsJsonFile(filePath string, data any) error {
	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	err = os.WriteFile(filePath, jsonData, 0644) //nolint:gosec // G306: Generic utility - callers determine sensitivity
	if err != nil {
		return fmt.Errorf("failed to write JSON to file: %w", err)
	}

	return nil
}

func LoadJsonFileAsStruct(filePath string) (any, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]any

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON %s to struct: %w", string(jsonData), err)
	}

	return data, nil
}

func UpdateTomlConfig(filePath string, key string, value any) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var newContent string

	switch value := value.(type) {
	case string:
		re := regexp.MustCompile(`(?m)^\s*` + key + `\s*=\s*".+"\s*$`)
		newContent = re.ReplaceAllString(string(content), fmt.Sprintf("%s = \"%s\"", key, value))
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	err = os.WriteFile(filePath, []byte(newContent), 0644) //nolint:gosec // G306: Config files are typically world-readable
	if err != nil {
		return err
	}

	return nil
}
