package server

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/audit"
	"github.com/OpenNHP/opennhp/nhp/common"
)

// TestAuditConfigParsesFromTOML guards the field-name match between the
// [Audit] TOML section and the AuditConfig struct — go-toml matches by Go
// field name, so a renamed field silently stops loading without this test.
func TestAuditConfigParsesFromTOML(t *testing.T) {
	const cfg = `
PrivateKeyBase64 = "x"
[Audit]
Enabled = true
FilePath = "logs/x.jsonl"
Fsync = true
SigningKeyBase64 = "AAAA"
`
	var c Config
	if err := toml.Unmarshal([]byte(cfg), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !c.Audit.Enabled {
		t.Error("Audit.Enabled did not parse")
	}
	if c.Audit.FilePath != "logs/x.jsonl" {
		t.Errorf("Audit.FilePath = %q", c.Audit.FilePath)
	}
	if !c.Audit.Fsync {
		t.Error("Audit.Fsync did not parse")
	}
	if c.Audit.SigningKeyBase64 != "AAAA" {
		t.Errorf("Audit.SigningKeyBase64 = %q", c.Audit.SigningKeyBase64)
	}
}

// TestAuditLedgerEmissionAndVerify exercises the server-side plumbing:
// initAuditLedger opens a file, auditEvent writes chained entries, and the
// resulting file verifies as an intact chain.
func TestAuditLedgerEmissionAndVerify(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	// initAuditLedger resolves relative paths against ExeDirPath; use an
	// absolute path so the test is independent of that package var.
	s := &UdpServer{
		config: &Config{
			Audit: AuditConfig{
				Enabled:  true,
				FilePath: path,
				Fsync:    false,
			},
		},
	}
	if err := s.initAuditLedger(); err != nil {
		t.Fatalf("initAuditLedger: %v", err)
	}
	if s.auditLedger == nil {
		t.Fatal("auditLedger is nil after init with Enabled=true")
	}

	s.auditEvent("knock", audit.SeverityInfo, map[string]string{"user": "alice", "result": "granted"})
	s.auditEvent("knock", audit.SeverityWarn, map[string]string{"user": "bob", "result": "denied"})
	s.auditEvent("agent_register", audit.SeverityNotice, map[string]string{"user": "carol", "result": "registered"})
	s.closeAuditLedger()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	res := audit.VerifyChain(bytes.NewReader(data), nil)
	if res.Err != nil {
		t.Fatalf("chain verify failed: %v", res.Err)
	}
	if res.Count != 3 {
		t.Fatalf("verified %d entries, want 3", res.Count)
	}
}

// TestAuditDisabledIsNoOp confirms auditEvent is safe when auditing is off.
func TestAuditDisabledIsNoOp(t *testing.T) {
	s := &UdpServer{config: &Config{Audit: AuditConfig{Enabled: false}}}
	if err := s.initAuditLedger(); err != nil {
		t.Fatalf("initAuditLedger (disabled): %v", err)
	}
	if s.auditLedger != nil {
		t.Fatal("auditLedger should be nil when disabled")
	}
	// Must not panic.
	s.auditEvent("knock", audit.SeverityInfo, map[string]string{"user": "x"})
	s.closeAuditLedger()
}

// A too-short signing key must be rejected at init rather than silently
// enabling signed=true with trivially forgeable protection.
func TestAuditShortSigningKeyRejected(t *testing.T) {
	dir := t.TempDir()
	s := &UdpServer{
		config: &Config{
			Audit: AuditConfig{
				Enabled:  true,
				FilePath: filepath.Join(dir, "audit.jsonl"),
				// "AAAA" decodes to 3 bytes — well under the 32-byte floor.
				SigningKeyBase64: "AAAA",
			},
		},
	}
	if err := s.initAuditLedger(); err == nil {
		t.Fatal("initAuditLedger accepted a 3-byte signing key")
	}
	if s.auditLedger != nil {
		t.Fatal("auditLedger must stay nil when the key is rejected")
	}
}

// The HTTP open-resource path must also feed the ledger. Exercise the
// early resource-not-found return (no AC connection needed): it must emit a
// denied "knock" entry tagged via=http with the resource recorded.
func TestAuditHTTPOpenResourceDenied(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	s := &UdpServer{config: &Config{Audit: AuditConfig{Enabled: true, FilePath: path}}}
	if err := s.initAuditLedger(); err != nil {
		t.Fatal(err)
	}
	hs := &HttpServer{udpServer: s}

	req := &common.HttpKnockRequest{
		UserId:        "alice",
		DeviceId:      "dev-1",
		AuthServiceId: "asp-1",
		SrcIp:         "203.0.113.7",
	}
	res := &common.ResourceData{} // no Resources -> ErrResourceNotFound
	res.ResourceId = "res-1"

	if _, err := hs.handleHttpOpenResource(req, res); err == nil {
		t.Fatal("expected ErrResourceNotFound for a request with no resources")
	}
	s.closeAuditLedger()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if res := audit.VerifyChain(bytes.NewReader(data), nil); res.Err != nil || res.Count != 1 {
		t.Fatalf("ledger not written for HTTP denial: err=%v count=%d", res.Err, res.Count)
	}
	// Confirm the entry carries the HTTP tag and the resource.
	s2 := string(data)
	for _, want := range []string{`"via":"http"`, `"result":"denied"`, `"resId":"res-1"`, `"user":"alice"`} {
		if !strings.Contains(s2, want) {
			t.Errorf("audit entry missing %s in: %s", want, s2)
		}
	}
}

// A corrupt/foreign file at the ledger path must not silently disable
// auditing: it is quarantined and a fresh chain starts, so the recorder
// keeps running (fail-safe, not fail-open).
func TestAuditQuarantinesNonLedgerAndContinues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	junk := []byte("\x00 not a ledger\nmore junk\n")
	if err := os.WriteFile(path, junk, 0600); err != nil {
		t.Fatal(err)
	}

	s := &UdpServer{
		config: &Config{Audit: AuditConfig{Enabled: true, FilePath: path}},
	}
	if err := s.initAuditLedger(); err != nil {
		t.Fatalf("initAuditLedger should recover a non-ledger file, got: %v", err)
	}
	if s.auditLedger == nil {
		t.Fatal("auditLedger is nil — auditing was disabled instead of quarantined")
	}

	// The original bytes must be preserved in a .corrupt-* sibling.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var quarantined string
	for _, e := range entries {
		if strings.Contains(e.Name(), ".corrupt-") {
			quarantined = filepath.Join(dir, e.Name())
		}
	}
	if quarantined == "" {
		t.Fatal("no .corrupt-* quarantine file was created")
	}
	if got, _ := os.ReadFile(quarantined); !bytes.Equal(got, junk) {
		t.Fatal("quarantined file does not hold the original bytes")
	}

	// The fresh chain works and verifies.
	s.auditEvent("knock", audit.SeverityInfo, map[string]string{"user": "alice"})
	s.closeAuditLedger()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if res := audit.VerifyChain(bytes.NewReader(data), nil); res.Err != nil {
		t.Fatalf("fresh chain failed to verify: %v", res.Err)
	}
}

// FailClosed turns an unrecoverable / opted-out open failure into a hard
// error rather than quarantining, so the caller can refuse to boot.
func TestAuditFailClosedReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	if err := os.WriteFile(path, []byte("not a ledger\n"), 0600); err != nil {
		t.Fatal(err)
	}
	s := &UdpServer{
		config: &Config{Audit: AuditConfig{Enabled: true, FilePath: path, FailClosed: true}},
	}
	if err := s.initAuditLedger(); err == nil {
		t.Fatal("FailClosed: initAuditLedger should return an error, not quarantine")
	}
	if s.auditLedger != nil {
		t.Fatal("FailClosed: auditLedger must stay nil")
	}
	// The file must be left untouched for inspection.
	if got, _ := os.ReadFile(path); !bytes.Equal(got, []byte("not a ledger\n")) {
		t.Fatal("FailClosed must not modify the offending file")
	}
}

// A full-length key is accepted and produces a signed, verifiable chain.
func TestAuditValidSigningKeyAccepted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	s := &UdpServer{
		config: &Config{
			Audit: AuditConfig{
				Enabled:          true,
				FilePath:         path,
				SigningKeyBase64: base64.StdEncoding.EncodeToString(key),
			},
		},
	}
	if err := s.initAuditLedger(); err != nil {
		t.Fatalf("initAuditLedger with a 32-byte key: %v", err)
	}
	s.auditEvent("knock", audit.SeverityInfo, map[string]string{"user": "alice"})
	s.closeAuditLedger()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if res := audit.VerifyChain(bytes.NewReader(data), key); res.Err != nil {
		t.Fatalf("signed chain failed to verify: %v", res.Err)
	}
}
