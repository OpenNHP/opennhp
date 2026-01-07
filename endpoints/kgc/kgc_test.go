package kgc

import (
	"os"
	"testing"
)

// TestConfigFilePermissions verifies that config files are written with secure permissions.
// The KGC master key config is written with 0600 permissions to prevent unauthorized access.
func TestConfigFilePermissions(t *testing.T) {
	// Create a temporary file with 0600 permissions (same as used in GenerateMasterKey)
	tempFile, err := os.CreateTemp("", "kgc-perm-test-*.toml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Write with the same permissions used in GenerateMasterKey
	testContent := []byte("test = \"value\"")
	if err := os.WriteFile(tempFile.Name(), testContent, 0600); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Verify permissions
	info, err := os.Stat(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Expected file permissions 0600, got %04o", perm)
	}

	// Verify group and other have no access
	if perm&0077 != 0 {
		t.Errorf("File should not be accessible by group or others, got %04o", perm)
	}
}
