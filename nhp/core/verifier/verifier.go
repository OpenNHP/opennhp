package verifier

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/OpenNHP/opennhp/nhp/core/verifier/csv"
)

// Scheme names the attestation evidence format the relying party expects.
// It is chosen by the VERIFIER's own configuration, never by the contents
// of the evidence being appraised — letting the evidence pick its own
// verifier is what let a caller select the no-op FallbackVerifier by
// adding a "test_purpose" key.
type Scheme string

const (
	// SchemeCSV is the Hygon CSV attestation report, verified against the
	// Hygon certificate chain. The only scheme with cryptographic value.
	SchemeCSV Scheme = "csv"
	// SchemeTest accepts self-asserted evidence with NO cryptographic
	// assurance whatsoever. Demo / quick-start only.
	SchemeTest Scheme = "test"
)

// ErrTestEvidenceRejected is returned by NewVerifier when the evidence
// carries a "test_purpose" key but the verifier is configured for a
// scheme that demands real attestation. Without this guard, an attacker
// could opt the relying party into the no-op FallbackVerifier simply by
// adding that key to their evidence.
var ErrTestEvidenceRejected = errors.New(
	"evidence carries test_purpose but this verifier is configured for scheme csv")

// ResolveScheme maps an operator-supplied string onto a Scheme, failing
// closed: anything unrecognised (including "") is SchemeCSV.
func ResolveScheme(s string) Scheme {
	if Scheme(strings.ToLower(strings.TrimSpace(s))) == SchemeTest {
		return SchemeTest
	}
	return SchemeCSV
}

// decompressEvidence reverses the on-the-wire framing used by nhp-agent:
// base64(zlib(json)).
func decompressEvidence(b64 string) ([]byte, error) {
	compressed, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode evidence: %v", err)
	}

	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %v", err)
	}
	defer r.Close()
	evidenceBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read evidence: %v", err)
	}
	return evidenceBytes, nil
}

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
	// The "test" scheme is for non-TEE demos only. It still refuses
	// empty values so that a knock carrying an empty JSON body cannot
	// synthesize a (Measure="", SerialNumber="") that matches a
	// misconfigured tee.toml entry.
	if f.Measure == "" || f.SerialNumber == "" {
		return errors.New("fallback verifier requires non-empty measure and serial_number")
	}
	return nil
}

func (f *FallbackVerifier) GetSerialNumber() string {
	return f.SerialNumber
}

func (f *FallbackVerifier) GetMeasure() string {
	return f.Measure
}

func NewFallbackVerifier(evidence []byte) (*FallbackVerifier, error) {
	fallbackVerifier := &FallbackVerifier{}

	err := json.Unmarshal(evidence, fallbackVerifier)
	if err != nil {
		return nil, err
	}

	return fallbackVerifier, nil
}

// NewVerifier dispatches to the verifier selected by `scheme`. The scheme
// must come from the verifier's own configuration — never from the
// evidence — because allowing the evidence to pick its own verifier is
// what made a "test_purpose" key opt the server into the no-op
// FallbackVerifier.
func NewVerifier(compressedEvidenceBase64 string, scheme Scheme) (Verifier, error) {
	evidenceBytes, err := decompressEvidence(compressedEvidenceBase64)
	if err != nil {
		return nil, err
	}

	switch scheme {
	case SchemeTest:
		v, err := NewFallbackVerifier(evidenceBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to create test verifier: %v", err)
		}
		return v, nil
	default:
		// csv (or anything else) — reject self-asserted evidence outright
		// instead of forwarding it into the CSV path, which would (a)
		// try to parse a non-CSV blob, and (b) hit a real, network-bound
		// verifyCertChain path. Failing closed here also avoids
		// amplifying a single knock into an unbounded HTTPS GET.
		var probe map[string]any
		if jerr := json.Unmarshal(evidenceBytes, &probe); jerr == nil {
			if _, ok := probe["test_purpose"]; ok {
				return nil, ErrTestEvidenceRejected
			}
		}
		v, err := csv.NewAttestation(string(evidenceBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create csv verifier: %v", err)
		}
		return v, nil
	}
}
