package resource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadResourceReadsOpenedRegularFile(t *testing.T) {
	originalBaseDir := baseDir
	baseDir = t.TempDir()
	t.Cleanup(func() { baseDir = originalBaseDir })

	resourcePath := filepath.Join(baseDir, "tenant", "resource")
	if err := os.MkdirAll(filepath.Dir(resourcePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(resourcePath, []byte("resource payload"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := loadResource("tenant/resource")
	if err != nil {
		t.Fatalf("loadResource returned %v", err)
	}
	if string(got) != "resource payload" {
		t.Fatalf("loadResource = %q", got)
	}
}

func TestLoadResourceRejectsNonRegularFile(t *testing.T) {
	originalBaseDir := baseDir
	baseDir = t.TempDir()
	t.Cleanup(func() { baseDir = originalBaseDir })

	if err := os.Mkdir(filepath.Join(baseDir, "directory"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := loadResource("directory"); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("loadResource error = %v, want non-regular-file error", err)
	}
}

func TestLoadResourceRejectsSiblingPrefix(t *testing.T) {
	root := t.TempDir()
	originalBaseDir := baseDir
	baseDir = filepath.Join(root, "repository")
	t.Cleanup(func() { baseDir = originalBaseDir })

	siblingFile := filepath.Join(root, "repository-attacker", "secret")
	if err := os.MkdirAll(filepath.Dir(siblingFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(siblingFile, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := loadResource("../repository-attacker/secret"); err == nil || !strings.Contains(err.Error(), "path traversal") {
		t.Fatalf("loadResource error = %v, want path-traversal error", err)
	}
}

func TestLoadResourceReportsMissingFile(t *testing.T) {
	originalBaseDir := baseDir
	baseDir = t.TempDir()
	t.Cleanup(func() { baseDir = originalBaseDir })

	if _, err := loadResource("missing"); err == nil || err.Error() != "resource not found" {
		t.Fatalf("loadResource error = %v, want resource not found", err)
	}
}
