package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestUpdateTomlConfigPreservesMode: an update must NOT widen a mode-0600
// config to world-readable — RotateAgentKey/RotateTeeKey write key-bearing
// files through this path.
func TestUpdateTomlConfigPreservesMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows does not enforce unix file modes")
	}
	dir := t.TempDir()
	fp := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(fp, []byte("PrivateKeyBase64 = \"old\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateTomlConfig(fp, "PrivateKeyBase64", "v1$new"); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(fp)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("mode widened to %o, want 0600", perm)
	}
}

// TestUpdateTomlConfigWritesValueLiterally guards against the replacement
// being interpreted as a regexp template. A sealed key blob is full of '$'
// ("v1$argon2id$3$65536$4$...") which ReplaceAllString would treat as
// capture-group references and mangle.
func TestUpdateTomlConfigWritesValueLiterally(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(fp, []byte("Name = \"old\"\nPrivateKeyBase64 = \"AAAA\"\nOther = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	blob := "v1$argon2id$3$65536$4$c2FsdHNhbHRzYWx0$bm9uY2Vub25j2Vu$Y2lwaGVydGV4dA"
	if err := UpdateTomlConfig(fp, "PrivateKeyBase64", blob); err != nil {
		t.Fatalf("UpdateTomlConfig: %v", err)
	}

	got, err := os.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	want := "PrivateKeyBase64 = \"" + blob + "\""
	if !strings.Contains(string(got), want) {
		t.Fatalf("blob was not written verbatim.\nwant line: %s\ngot file:\n%s", want, got)
	}
	// Untouched lines stay put.
	if !strings.Contains(string(got), "Name = \"old\"") || !strings.Contains(string(got), "Other = 1") {
		t.Fatalf("unrelated lines changed:\n%s", got)
	}
}

// TestUpdateTomlConfigReplacesEmptyValue: a currently-empty `key = ""` must
// be filled in, not skipped — RotateAgentKey relies on the write persisting
// the re-sealed blob.
func TestUpdateTomlConfigReplacesEmptyValue(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(fp, []byte("PrivateKeyBase64 = \"\"\nOther = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateTomlConfig(fp, "PrivateKeyBase64", "v1$x"); err != nil {
		t.Fatalf("UpdateTomlConfig over an empty value: %v", err)
	}
	got, _ := os.ReadFile(fp)
	if !strings.Contains(string(got), `PrivateKeyBase64 = "v1$x"`) || !strings.Contains(string(got), "Other = 1") {
		t.Fatalf("empty value not replaced in place:\n%s", got)
	}
}

// TestUpdateTomlConfigAppendsWhenAbsentAndNoTables: a bootstrap/partial file
// (no [table] headers) self-heals by appending the missing root key.
func TestUpdateTomlConfigAppendsWhenAbsentAndNoTables(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "dhp.toml")
	if err := os.WriteFile(fp, []byte("# generated\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateTomlConfig(fp, "TEEPrivateKeyBase64", "AAAA"); err != nil {
		t.Fatalf("UpdateTomlConfig append: %v", err)
	}
	got, _ := os.ReadFile(fp)
	if !strings.Contains(string(got), `TEEPrivateKeyBase64 = "AAAA"`) {
		t.Fatalf("missing key not appended:\n%s", got)
	}
}

// TestUpdateTomlConfigErrorsWhenAbsentWithTables: appending a bare key at EOF
// would land it inside the last [table], so that case is refused instead.
func TestUpdateTomlConfigErrorsWhenAbsentWithTables(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "config.toml")
	orig := "Name = \"x\"\n[UserData]\n\"k\" = \"v\"\n"
	if err := os.WriteFile(fp, []byte(orig), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateTomlConfig(fp, "PrivateKeyBase64", "x"); err == nil {
		t.Fatal("expected an error: absent key + [table] sections present")
	}
	got, _ := os.ReadFile(fp)
	if string(got) != orig {
		t.Fatalf("file was modified despite the refusal:\n%s", got)
	}
}

// TestUpdateTomlConfigReplacesCRLFLine: a \r\n-terminated line (checked out
// with core.autocrlf=true, or saved by a Windows editor) must still hit the
// in-place replace branch. Before [ \t\r]*$ this fell through to "absent",
// and combined with the old regex-miss-means-absent heuristic it produced a
// second, duplicate key rather than editing the existing line.
func TestUpdateTomlConfigReplacesCRLFLine(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "dhp.toml")
	// No [table] header — this is exactly the shape of
	// docker/nhp-agent/etc/dhp.toml, the file the CRLF regression bit.
	orig := "# generated\r\nTEEPrivateKeyBase64 = \"old\"\r\n"
	if err := os.WriteFile(fp, []byte(orig), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateTomlConfig(fp, "TEEPrivateKeyBase64", "new"); err != nil {
		t.Fatalf("UpdateTomlConfig over a CRLF file: %v", err)
	}
	got, _ := os.ReadFile(fp)
	if n := strings.Count(string(got), "TEEPrivateKeyBase64"); n != 1 {
		t.Fatalf("expected exactly one TEEPrivateKeyBase64 line, got %d:\n%s", n, got)
	}
	if !strings.Contains(string(got), `TEEPrivateKeyBase64 = "new"`) {
		t.Fatalf("value not replaced in place:\n%s", got)
	}
}

// TestUpdateTomlConfigErrorsRatherThanDuplicatesOnUnrecognizedForm: a value
// form the line regex does not match (here, a trailing comment) must not be
// silently treated as "key absent" and appended — go-toml/v2 rejects the
// resulting duplicate key outright, and the caller that only logs the parse
// error (updateDHPConfig) would carry on with an empty config.
func TestUpdateTomlConfigErrorsRatherThanDuplicatesOnUnrecognizedForm(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "dhp.toml")
	orig := "TEEPrivateKeyBase64 = \"old\" # rotated 2026-01-01\n"
	if err := os.WriteFile(fp, []byte(orig), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateTomlConfig(fp, "TEEPrivateKeyBase64", "new"); err == nil {
		t.Fatal("expected an error: the existing line has a trailing comment the regex does not match")
	}
	got, _ := os.ReadFile(fp)
	if string(got) != orig {
		t.Fatalf("file was modified despite the refusal:\n%s", got)
	}
}

// TestUpdateTomlConfigRoundTripsPlainKey confirms the ordinary path is
// unaffected.
func TestUpdateTomlConfigRoundTripsPlainKey(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(fp, []byte("PrivateKeyBase64 = \"AAAA\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateTomlConfig(fp, "PrivateKeyBase64", "eHdyRHKJy/YZJsResCt5XTAZgtcwvLpSXAiZ8DBc0V4="); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(fp)
	if !strings.Contains(string(got), `PrivateKeyBase64 = "eHdyRHKJy/YZJsResCt5XTAZgtcwvLpSXAiZ8DBc0V4="`) {
		t.Fatalf("plain key not written as expected:\n%s", got)
	}
}
