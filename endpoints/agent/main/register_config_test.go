package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriteResourceConfig verifies the generated resource.toml binds the
// asp-id/res-id to the named cluster and is written under etc/.
func TestWriteResourceConfig(t *testing.T) {
	dir := t.TempDir()
	if err := writeResourceConfig(dir, "example", "demo", "cluster1"); err != nil {
		t.Fatalf("writeResourceConfig: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "etc", "resource.toml"))
	if err != nil {
		t.Fatalf("read resource.toml: %v", err)
	}
	s := string(data)
	for _, want := range []string{
		`AuthServiceId = "example"`,
		`ResourceId    = "demo"`,
		`Cluster       = "cluster1"`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("resource.toml missing %q in:\n%s", want, s)
		}
	}
}

// TestWriteRegistrationConfig verifies the generated config.toml carries
// the registered identity, private key, and the selected cipher scheme.
func TestWriteRegistrationConfig(t *testing.T) {
	dir := t.TempDir()
	if err := writeRegistrationConfig(dir, "PRIVKEYB64", "alice@example.com", "opennhp.org", 1); err != nil {
		t.Fatalf("writeRegistrationConfig: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "etc", "config.toml"))
	if err != nil {
		t.Fatalf("read config.toml: %v", err)
	}
	s := string(data)
	for _, want := range []string{
		`PrivateKeyBase64 = "PRIVKEYB64"`,
		`DefaultCipherScheme = 1`,
		`UserId = "alice@example.com"`,
		`OrganizationId = "opennhp.org"`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("config.toml missing %q in:\n%s", want, s)
		}
	}
}

// TestWriteRegistrationConfig_CurveScheme pins the curve25519 scheme code (0).
func TestWriteRegistrationConfig_CurveScheme(t *testing.T) {
	dir := t.TempDir()
	if err := writeRegistrationConfig(dir, "K", "u", "", 0); err != nil {
		t.Fatalf("writeRegistrationConfig: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "etc", "config.toml"))
	if !strings.Contains(string(data), "DefaultCipherScheme = 0") {
		t.Fatalf("expected DefaultCipherScheme = 0 for curve25519, got:\n%s", data)
	}
}

// TestWriteConfig_BacksUpExisting verifies an existing config.toml is
// preserved as config.toml.bak before being overwritten, so a re-run of
// `register` never irreversibly destroys prior config.
func TestWriteConfig_BacksUpExisting(t *testing.T) {
	dir := t.TempDir()
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}
	original := "# hand-written config\nUserId = \"pre-existing\"\nLogLevel = 4\n"
	if err := os.WriteFile(filepath.Join(etc, "config.toml"), []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := writeRegistrationConfig(dir, "NEWKEY", "new@user", "", 1); err != nil {
		t.Fatalf("writeRegistrationConfig: %v", err)
	}

	// New content is written...
	newData, _ := os.ReadFile(filepath.Join(etc, "config.toml"))
	if !strings.Contains(string(newData), "NEWKEY") {
		t.Fatalf("config.toml was not overwritten with new content:\n%s", newData)
	}
	// ...and the original is preserved in .bak.
	bak, err := os.ReadFile(filepath.Join(etc, "config.toml.bak"))
	if err != nil {
		t.Fatalf("expected config.toml.bak to exist: %v", err)
	}
	if string(bak) != original {
		t.Fatalf("config.toml.bak does not match the original:\ngot:  %s\nwant: %s", bak, original)
	}
}

// TestBackupIfExists_NoFileNoError: backing up a non-existent file is a
// no-op (no error, no .bak created).
func TestBackupIfExists_NoFileNoError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.toml")
	if err := backupIfExists(path); err != nil {
		t.Fatalf("backupIfExists on missing file should be nil, got %v", err)
	}
	if _, err := os.Stat(path + ".bak"); !os.IsNotExist(err) {
		t.Fatalf("no .bak should be created for a missing source file")
	}
}
