package verifier

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/OpenNHP/opennhp/nhp/core/verifier/csv"
)

type Verifier interface {
	// This interface is used to ask verifier to verify the attestation report
	// which is collected from attestor which is NHP Agent in DHP.
	// After receiving the attestation result, NHP Server makes application-specific decisions.
	Verify() error

	// GetSerialNumber returns the serial number from the attestation report
	GetSerialNumber() string

	// GetMeasure returns the measure from the attestation report
	GetMeasure() string
}

type FallbackVerifier struct {
	TestPurpose  string `json:"test_purpose"`
	Measure      string `json:"measure"`
	SerialNumber string `json:"serial_number"`
}

func (f *FallbackVerifier) Verify() error {
	return nil
}

func (f *FallbackVerifier) GetSerialNumber() string {
	return f.Measure
}

func (f *FallbackVerifier) GetMeasure() string {
	return f.SerialNumber
}

func NewFallbackVerifier(evidence []byte) (*FallbackVerifier, error) {
	fallbackVerifier := &FallbackVerifier{}

	err := json.Unmarshal(evidence, fallbackVerifier)
	if err != nil {
		return nil, err
	}

	return fallbackVerifier, nil
}

func NewVerifier(compressedEvienceBase64 string) (Verifier, error) {
	compressedEvidence, err := base64.StdEncoding.DecodeString(compressedEvienceBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode evidence: %v", err)
	}

	r, err := zlib.NewReader(bytes.NewReader(compressedEvidence))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %v", err)
	}
	defer r.Close()
	evidenceBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read evidence: %v", err)
	}

	var evidence map[string]any

	err = json.Unmarshal(evidenceBytes, &evidence)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal evidence: %v", err)
	}

	var verifier Verifier

	if _, ok := evidence["test_purpose"]; ok {
		verifier, err = NewFallbackVerifier(evidenceBytes)
	} else {
		verifier, err = csv.NewAttestation(string(evidenceBytes))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create csv verifier: %v", err)
	} else {
		return verifier, nil
	}
}
