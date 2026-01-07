// Package engine provides host functions that can be called from within a WebAssembly (WASM) virtual machine.
// These functions enable interaction between the WASM runtime and the host environment,
// such as logging operations or other system-level interactions.

package engine

import (
	"bytes"
	"compress/zlib"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tetratelabs/wazero/api"
)

var (
	confidentialContainerEvidenceUrl = "http://127.0.0.1:8006/aa/evidence?runtime_data=dhp"
)

func logString(_ context.Context, m api.Module, offset, byteCount uint32) {
	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		log.Panicf("Memory.Read(%d, %d) out of range", offset, byteCount)
	}
	fmt.Println(string(buf))
}

func GetEvidenceWithCCUrl() ([]byte, error) {
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(confidentialContainerEvidenceUrl)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err = w.Write(body)
	w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to compress response body: %w", err)
	}

	compressedBody := buf.Bytes()

	return compressedBody, nil
}

func GetEvidenceWithAgentUuid() ([]byte, error) {
	agentUniqueId, err := CalculateAgentUniqueId()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent unique id: %v", err)
	}

	evidence := map[string]any{
		"test_purpose":  "this evidence is for testing purposes only",
		"measure":       agentUniqueId,
		"serial_number": agentUniqueId,
	}

	evidenceBytes, err := json.Marshal(evidence)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal evidence: %v", err)
	}

	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	_, err = w.Write(evidenceBytes)
	w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to compress response body: %w", err)
	}

	return buf.Bytes(), nil
}

func GetEvidence() (string, error) {
	evidence, err := GetEvidenceWithCCUrl()
	if err != nil {
		evidence, err = GetEvidenceWithAgentUuid()
		if err != nil {
			return "", fmt.Errorf("failed to get evidence from CC or agent uuid")
		}
	}

	return base64.StdEncoding.EncodeToString(evidence), nil
}

func CalculateAgentUniqueId() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	cgroup, _ := os.ReadFile("/proc/self/cgroup")
	combined := hostname + string(cgroup)
	sum := sha256.Sum256([]byte(combined))

	return hex.EncodeToString(sum[:]), nil
}
