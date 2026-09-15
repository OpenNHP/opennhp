package resource

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
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

	got, err := loadResource("/tenant/resource")
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

	if _, err := loadResource("/../repository-attacker/secret"); err == nil || !strings.Contains(err.Error(), "path traversal") {
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

func TestLoadResourceRejectsEscapingSymlink(t *testing.T) {
	old := baseDir
	baseDir = t.TempDir()
	t.Cleanup(func() { baseDir = old })
	outside := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(outside, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(baseDir, "link")); err != nil {
		t.Skip(err)
	}
	if data, err := loadResource("link"); err == nil {
		t.Fatalf("escaped root: %q", data)
	}
}

func TestResourceCatchAllPath(t *testing.T) {
	old := baseDir
	baseDir = t.TempDir()
	t.Cleanup(func() { baseDir = old })
	if err := os.MkdirAll(filepath.Join(baseDir, "tenant"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "tenant", "resource"), []byte("payload"), 0600); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/kbs/v0/resource/*path", func(c *gin.Context) {
		data, err := loadResource(c.Param("path"))
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "application/octet-stream", data)
	})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/kbs/v0/resource/tenant/resource", nil))
	if response.Code != http.StatusOK || response.Body.String() != "payload" {
		t.Fatalf("wire resource read: %d %q", response.Code, response.Body.String())
	}
}
