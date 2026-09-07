package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCurrentConfigKeySealed covers the three cases: no config (bootstrap),
// a sealed key, and an unparseable file (must error, not read as "not
// sealed" — that would let register overwrite a sealed key with plaintext).
func TestCurrentConfigKeySealed(t *testing.T) {
	dir := t.TempDir()
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}

	// No config yet — bootstrap, not an error.
	if sealed, err := currentConfigKeySealed(dir); err != nil || sealed {
		t.Fatalf("missing config: got (%v, %v), want (false, nil)", sealed, err)
	}

	// Sealed key.
	cfg := filepath.Join(etc, "config.toml")
	if err := os.WriteFile(cfg, []byte("PrivateKeyBase64 = \"v1$argon2id$3$65536$4$a$b$c\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if sealed, err := currentConfigKeySealed(dir); err != nil || !sealed {
		t.Fatalf("sealed config: got (%v, %v), want (true, nil)", sealed, err)
	}

	// Unparseable file — must surface an error.
	if err := os.WriteFile(cfg, []byte("PrivateKeyBase64 = \"oops\n[Broken\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if sealed, err := currentConfigKeySealed(dir); err == nil {
		t.Fatalf("unparseable config: got (%v, nil), want an error", sealed)
	}
}

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
	if err := writeRegistrationConfig(dir, "PRIVKEYB64", "alice@example.com", "opennhp.org", 1, false); err != nil {
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
	if strings.Contains(s, "SEALED") {
		t.Fatalf("plain-key config.toml should not mention a passphrase:\n%s", s)
	}
}

// TestWriteRegistrationConfig_SealedNote verifies that when the registered
// key is sealed the generated config.toml tells the operator a passphrase is
// now required at startup.
func TestWriteRegistrationConfig_SealedNote(t *testing.T) {
	dir := t.TempDir()
	if err := writeRegistrationConfig(dir, "v1$argon2id$3$65536$4$a$b$c", "u", "", 1, true); err != nil {
		t.Fatalf("writeRegistrationConfig: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "etc", "config.toml"))
	s := string(data)
	for _, want := range []string{"sealed blob", "NHP_KEY_PASSPHRASE_FILE"} {
		if !strings.Contains(s, want) {
			t.Fatalf("sealed config.toml missing %q in:\n%s", want, s)
		}
	}
}

// TestWriteRegistrationConfig_CurveScheme pins the curve25519 scheme code (0).
func TestWriteRegistrationConfig_CurveScheme(t *testing.T) {
	dir := t.TempDir()
	if err := writeRegistrationConfig(dir, "K", "u", "", 0, false); err != nil {
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

	if err := writeRegistrationConfig(dir, "NEWKEY", "new@user", "", 1, false); err != nil {
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
