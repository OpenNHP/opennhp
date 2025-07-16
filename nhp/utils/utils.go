package utils

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/OpenNHP/opennhp/nhp/log"
)

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
				log.Error(line)
			}
		}
	}
}

func CatchPanicThenRun(catchFun func()) {
	if x := recover(); x != nil {
		for _, line := range append([]string{fmt.Sprint(x)}, strings.Split(string(debug.Stack()), "\n")...) {
			if len(strings.TrimSpace(line)) > 0 {
				log.Error(line)
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

	resp, err := http.Get(fileUrl)
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