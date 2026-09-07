package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

// TestUpdateTomlConfigErrorsWhenKeyMissing: a no-match must be reported, not
// swallowed — RotateAgentKey relies on the write actually persisting the
// re-sealed blob.
func TestUpdateTomlConfigErrorsWhenKeyMissing(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "config.toml")
	// Key present but empty, plus a wholly absent key.
	if err := os.WriteFile(fp, []byte("PrivateKeyBase64 = \"\"\nOther = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := UpdateTomlConfig(fp, "PrivateKeyBase64", "x"); err == nil {
		t.Fatal("expected an error updating an empty-valued key")
	}
	if err := UpdateTomlConfig(fp, "NoSuchKey", "x"); err == nil {
		t.Fatal("expected an error updating an absent key")
	}
	// The file must be left untouched on the error path.
	got, _ := os.ReadFile(fp)
	if string(got) != "PrivateKeyBase64 = \"\"\nOther = 1\n" {
		t.Fatalf("file was modified despite the no-match error:\n%s", got)
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
