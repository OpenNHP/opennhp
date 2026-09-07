package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// GetRandomUint32 returns a non-zero random uint32 for packet preamble obfuscation.
// Uses math/rand which is sufficient for this non-cryptographic use case.
//
//nolint:gosec // G404: math/rand is intentional - used for packet obfuscation, not security
func GetRandomUint32() (r uint32) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		r = rng.Uint32()
		if r != 0 {
			break
		}
	}
	return r
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

// UpdateTomlConfig sets a top-level string key in a TOML file.
//
// If a `key = "…"` assignment exists (empty value included) it is replaced
// in place. If the key is ABSENT and the file has no `[table]` headers, the
// assignment is appended so a bootstrap/partial config self-heals rather
// than the update silently doing nothing. If the key is absent but the file
// DOES have table headers, appending a bare key at EOF would land it inside
// the last table, so this returns an error instead — the caller must add the
// root-table line by hand.
func UpdateTomlConfig(filePath string, key string, value any) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var newContent string

	switch value := value.(type) {
	case string:
		replacement := fmt.Sprintf("%s = \"%s\"", key, value)
		// Match the assignment line whatever its current value (including
		// an empty ""), so a re-seal over `key = ""` still lands.
		lineRe := regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(key) + `[ \t]*=.*$`)
		switch {
		case lineRe.MatchString(string(content)):
			// ReplaceAllLiteralString, not ReplaceAllString: the replacement
			// is a verbatim value, and a sealed key blob ("v1$argon2id$...")
			// contains '$' sequences that ReplaceAllString would interpret as
			// capture-group references and mangle.
			newContent = lineRe.ReplaceAllLiteralString(string(content), replacement)
		case !regexp.MustCompile(`(?m)^\s*\[`).MatchString(string(content)):
			// Key absent, no tables: safe to append as a root-table key.
			base := string(content)
			if len(base) > 0 && !strings.HasSuffix(base, "\n") {
				base += "\n"
			}
			newContent = base + replacement + "\n"
		default:
			return fmt.Errorf("key %q not found in %s and the file has [table] sections; add the line under the root table by hand", key, filePath)
		}
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	err = os.WriteFile(filePath, []byte(newContent), 0644) //nolint:gosec // G306: Config files are typically world-readable
	if err != nil {
		return err
	}

	return nil
}
