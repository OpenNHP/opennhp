package server

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/audit"
	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// A plugin may legally return a failure ErrCode with a nil error (a soft
// denial). That must record as denied, not granted, or a SIEM rule keyed on
// result misses it. An empty or success ("0") code with no error is a grant.
func TestDecisionGranted(t *testing.T) {
	success := common.ErrSuccess.ErrorCode()
	fail := common.ErrResourceNotFound.ErrorCode()
	cases := []struct {
		name string
		err  error
		code string
		want bool
	}{
		{"nil err, empty code", nil, "", true},
		{"nil err, success code", nil, success, true},
		{"nil err, failure code (soft denial)", nil, fail, false},
		{"error, empty code", common.ErrResourceNotFound, "", false},
		{"error, success code", common.ErrResourceNotFound, success, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := decisionGranted(tc.err, tc.code); got != tc.want {
				t.Fatalf("decisionGranted(%v, %q) = %v, want %v", tc.err, tc.code, got, tc.want)
			}
		})
	}
}

// A rejected knock whose body never populated HeaderType (parse failure, or a
// legacy agent) must not be recorded as a keepalive: HeaderTypeToString(0) is
// NHP_KPL, so an unset type is reported as "unknown" instead.
func TestAuditKnockOp(t *testing.T) {
	cases := []struct {
		name       string
		headerType int
		want       string
	}{
		{"unset/parse-failed", core.NHP_KPL, "unknown"},
		{"open", core.NHP_KNK, core.HeaderTypeToString(core.NHP_KNK)},
		{"close", core.NHP_EXT, core.HeaderTypeToString(core.NHP_EXT)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := auditKnockOp(tc.headerType); got != tc.want {
				t.Fatalf("auditKnockOp(%d) = %q, want %q", tc.headerType, got, tc.want)
			}
		})
	}
}

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
	// Confirm the entry carries the HTTP tag, the resource, and the error
	// code from the denial. errCode must be read off the local ackMsg, not
	// the named return (which is nil on the naked return here); asserting the
	// exact ErrResourceNotFound code guards that fix from regressing.
	s2 := string(data)
	wantCode := `"errCode":"` + common.ErrResourceNotFound.ErrorCode() + `"`
	for _, want := range []string{`"via":"http"`, `"result":"denied"`, `"resId":"res-1"`, `"user":"alice"`, wantCode} {
		if !strings.Contains(s2, want) {
			t.Errorf("audit entry missing %s in: %s", want, s2)
		}
	}
}

// A corrupt/foreign file at the ledger path must not silently disable
// auditing: it is quarantined and a fresh chain starts, so the recorder
// keeps running (fail-safe, not fail-open) — WITHOUT renaming the operator's
// file, which the server (often privileged) must not do to an unrelated path.
func TestAuditForeignFileLeftUntouchedAndContinuesAtSibling(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "some-other.log")
	junk := []byte("\x00 not a ledger\nmore junk\n")
	if err := os.WriteFile(path, junk, 0600); err != nil {
		t.Fatal(err)
	}

	s := &UdpServer{
		config: &Config{Audit: AuditConfig{Enabled: true, FilePath: path}},
	}
	if err := s.initAuditLedger(); err != nil {
		t.Fatalf("initAuditLedger should continue past a foreign file, got: %v", err)
	}
	if s.auditLedger == nil {
		t.Fatal("auditLedger is nil — auditing was disabled instead of continuing at a sibling")
	}
	t.Cleanup(s.closeAuditLedger)

	// The operator's file must be exactly as it was — not renamed, not touched.
	if got, _ := os.ReadFile(path); !bytes.Equal(got, junk) {
		t.Fatal("the foreign file was modified — the server must leave it alone")
	}
	entries, _ := os.ReadDir(dir)
	var sibling string
	for _, e := range entries {
		if strings.Contains(e.Name(), ".corrupt-") {
			t.Fatalf("a .corrupt-* rename happened on a foreign file: %s", e.Name())
		}
		if strings.Contains(e.Name(), ".quarantined.") {
			sibling = filepath.Join(dir, e.Name())
		}
	}
	if sibling == "" {
		t.Fatal("no .quarantined.jsonl sibling ledger was created")
	}

	// The fresh chain at the sibling works and verifies.
	s.auditEvent("knock", audit.SeverityInfo, map[string]string{"user": "alice"})
	s.closeAuditLedger()
	if data, _ := os.ReadFile(sibling); audit.VerifyChain(bytes.NewReader(data), nil).Err != nil {
		t.Fatal("sibling chain failed to verify")
	}
}

// A file that IS our ledger but whose first line got corrupted (an attacker
// prepending junk to disable auditing) is still quarantined to .corrupt-*
// and a fresh chain started — that path is a loud, detectable signal.
func TestAuditCorruptedLedgerHeaderIsQuarantined(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")

	// Build a real 3-entry ledger, then prepend a junk byte to its first line.
	l, err := audit.Open(path, audit.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		_ = l.Log("knock", audit.SeverityInfo, nil)
	}
	l.Close()
	orig, _ := os.ReadFile(path)
	if err := os.WriteFile(path, append([]byte("x"), orig...), 0600); err != nil {
		t.Fatal(err)
	}

	s := &UdpServer{config: &Config{Audit: AuditConfig{Enabled: true, FilePath: path}}}
	if err := s.initAuditLedger(); err != nil {
		t.Fatalf("initAuditLedger: %v", err)
	}
	t.Cleanup(s.closeAuditLedger)

	entries, _ := os.ReadDir(dir)
	var corrupt string
	for _, e := range entries {
		if strings.Contains(e.Name(), ".corrupt-") {
			corrupt = filepath.Join(dir, e.Name())
		}
	}
	if corrupt == "" {
		t.Fatal("a corrupted ledger header should be quarantined to .corrupt-*")
	}
	if got, _ := os.ReadFile(corrupt); !bytes.Equal(got, append([]byte("x"), orig...)) {
		t.Fatal("quarantined file does not hold the original (corrupted) bytes")
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

// TestAuditInitRunsBeforeListeners guards the fix for the FailClosed gap:
// initAuditLedger must be called before loadHttpConfig (which starts the
// HTTP knock listener) and before the UDP socket is opened, or "fail
// closed" would still have served requests in the window between bind and
// the error return. Pinned by scanning Start's source so a future reorder
// trips this test.
func TestAuditInitRunsBeforeListeners(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Skip("cannot locate source file")
	}
	src, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "udpserver.go"))
	if err != nil {
		t.Fatalf("read udpserver.go: %v", err)
	}
	body := string(src)
	start := strings.Index(body, "func (s *UdpServer) Start(")
	if start < 0 {
		t.Fatal("Start not found")
	}
	seg := body[start:]
	iAudit := strings.Index(seg, "s.initAuditLedger()")
	iHTTP := strings.Index(seg, "s.loadHttpConfig()")
	iListen := strings.Index(seg, "net.ListenUDP(")
	if iAudit < 0 || iHTTP < 0 || iListen < 0 {
		t.Fatalf("markers not found (audit=%d http=%d listen=%d)", iAudit, iHTTP, iListen)
	}
	if iAudit > iHTTP {
		t.Fatal("initAuditLedger() must precede loadHttpConfig() in Start()")
	}
	if iAudit > iListen {
		t.Fatal("initAuditLedger() must precede net.ListenUDP() in Start()")
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
