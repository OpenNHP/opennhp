package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

// LoadFilesRecursively walks a directory and parses every .html/.tmpl file
// into the gin engine. A missing directory, a well-formed template, and a
// malformed one that fails to parse must all be handled without panicking.
func TestLoadFilesRecursively(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// missing directory: returns early, no error
	LoadFilesRecursively(gin.New(), filepath.Join(t.TempDir(), "nonexistent"))

	// well-formed template: parsed and registered
	okDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(okDir, "ok.html"), []byte("<p>{{.X}}</p>"), 0o600); err != nil {
		t.Fatal(err)
	}
	LoadFilesRecursively(gin.New(), okDir)

	// malformed template: the parse error is logged, not fatal
	badDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(badDir, "bad.html"), []byte("{{ .Unclosed "), 0o600); err != nil {
		t.Fatal(err)
	}
	LoadFilesRecursively(gin.New(), badDir)
}
