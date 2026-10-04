package verifier

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// buildEvidence encodes a JSON map as base64(zlib(json)), matching the
// on-the-wire framing that nhp-agent produces and nhp-server expects.
func buildEvidence(t *testing.T, fields map[string]any) string {
	t.Helper()
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var buf bytes.Buffer
	zw := zlib.NewWriter(&buf)
	if _, err := zw.Write(raw); err != nil {
		t.Fatalf("zlib write: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zlib close: %v", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// TestNewVerifier_RejectsTestPurposeInCSVScheme is the regression test
// for the bug where an attacker-supplied "test_purpose" key in evidence
// opted the server into the no-op FallbackVerifier.
func TestNewVerifier_RejectsTestPurposeInCSVScheme(t *testing.T) {
	const measure = "00112233445566778899aabbccddeeff"
	const sn = "deadbeefcafefeed"

	b64 := buildEvidence(t, map[string]any{
		"test_purpose":  "this evidence is for testing purposes only",
		"measure":       measure,
		"serial_number": sn,
	})

	v, err := NewVerifier(b64, SchemeCSV)
	if err == nil {
		t.Fatalf("expected error for test_purpose evidence under csv scheme, got verifier %T", v)
	}
	if !errors.Is(err, ErrTestEvidenceRejected) {
		t.Fatalf("expected ErrTestEvidenceRejected, got %v", err)
	}
	if _, ok := v.(*FallbackVerifier); ok {
		t.Fatalf("verifier must NOT be *FallbackVerifier when scheme is csv")
	}
}

// TestNewVerifier_TestSchemeReturnsFallback verifies that under the test
// scheme, the same self-asserted evidence is accepted as a fallback
// verifier, and that D1 (the swapped GetMeasure / GetSerialNumber
// implementations) is fixed: measure and serial are returned correctly.
func TestNewVerifier_TestSchemeReturnsFallback(t *testing.T) {
	const measure = "abc123abc123abc123abc123abc123ab"
	const sn = "serl-sern-serl-sern-serl-sern"

	b64 := buildEvidence(t, map[string]any{
		"test_purpose":  "this evidence is for testing purposes only",
		"measure":       measure,
		"serial_number": sn,
	})

	v, err := NewVerifier(b64, SchemeTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fv, ok := v.(*FallbackVerifier)
	if !ok {
		t.Fatalf("expected *FallbackVerifier under scheme=test, got %T", v)
	}
	if fv.TestPurpose == "" {
		t.Errorf("TestPurpose should be set")
	}
	if err := v.Verify(); err != nil {
		t.Errorf("Verify() under scheme=test should not return error for non-empty fields, got %v", err)
	}
	if got := v.GetMeasure(); got != measure {
		t.Errorf("GetMeasure() = %q, want %q", got, measure)
	}
	if got := v.GetSerialNumber(); got != sn {
		t.Errorf("GetSerialNumber() = %q, want %q (D1 regression: was swapped)", got, sn)
	}
}

func TestResolveScheme(t *testing.T) {
	cases := []struct {
		in   string
		want Scheme
	}{
		{"", SchemeCSV},         // fail closed
		{"csv", SchemeCSV},      // explicit csv
		{"CSV", SchemeCSV},      // case-insensitive
		{"  CSV  ", SchemeCSV},  // trimmed
		{"bogus", SchemeCSV},    // unknown -> csv
		{"test", SchemeTest},
		{"TEST", SchemeTest},
		{"  test  ", SchemeTest},
		{"Test", SchemeTest},
	}
	for _, c := range cases {
		if got := ResolveScheme(c.in); got != c.want {
			t.Errorf("ResolveScheme(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNewVerifier_BadInputs(t *testing.T) {
	tests := map[string]string{
		"not-base64!@#":  "garbage base64",
		"AAAA":           "valid base64 but not zlib",
		"eNqLyk4tKEhMSdNJys7PSwIA8TgJ+Q==": // random non-JSON zlib blob
			"valid base64+zlib but not JSON",
	}
	for name, b64 := range tests {
		t.Run(name, func(t *testing.T) {
			v, err := NewVerifier(b64, SchemeCSV)
			if err == nil {
				t.Fatalf("expected error, got verifier %T", v)
			}
		})
	}

	// Manually craft a JSON with non-object root to hit the JSON branch
	// without going through zlib (the csv branch will still complain via
	// json.Unmarshal).
	arrJSON := []byte(`[1,2,3]`)
	var arrBuf bytes.Buffer
	zw := zlib.NewWriter(&arrBuf)
	zw.Write(arrJSON)
	zw.Close()
	arrB64 := base64.StdEncoding.EncodeToString(arrBuf.Bytes())
	if _, err := NewVerifier(arrB64, SchemeCSV); err == nil {
		t.Errorf("expected error for JSON array root, got nil")
	}
}

func TestNewVerifier_TestSchemeEmptyFieldsRejected(t *testing.T) {
	b64 := buildEvidence(t, map[string]any{
		"test_purpose":  "x",
		"measure":       "",
		"serial_number": "",
	})
	v, err := NewVerifier(b64, SchemeTest)
	if err != nil {
		t.Fatalf("NewVerifier: %v", err)
	}
	if err := v.Verify(); err == nil {
		t.Errorf("Verify() should reject empty measure / serial_number, got nil")
	} else if !strings.Contains(err.Error(), "non-empty") {
		t.Logf("Verify() returned error (acceptable): %v", err)
	}
}

func TestErrTestEvidenceRejectedMessage(t *testing.T) {
	if ErrTestEvidenceRejected == nil {
		t.Fatal("ErrTestEvidenceRejected must be defined")
	}
	if !strings.Contains(ErrTestEvidenceRejected.Error(), "test_purpose") {
		t.Errorf("ErrTestEvidenceRejected message should mention test_purpose, got %q",
			ErrTestEvidenceRejected.Error())
	}
}
