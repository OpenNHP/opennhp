package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newStorageTestServer(t *testing.T) *HttpServer {
	t.Helper()
	gin.SetMode(gin.TestMode)
	httpServer := &HttpServer{ginEngine: gin.New()}
	httpServer.initStorageRouter()
	return httpServer
}

func TestStorageDownloadServesOpenedRegularFile(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	const (
		uuid     = "download-id"
		filename = "report.txt"
		body     = "descriptor-backed response"
	)
	fileDir := filepath.Join(ExeDirPath, uploadDir, uuid)
	if err := os.MkdirAll(fileDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fileDir, filename), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/storage/download/"+uuid+"/"+filename, nil)
	response := httptest.NewRecorder()
	newStorageTestServer(t).ginEngine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if response.Body.String() != body {
		t.Fatalf("body = %q, want %q", response.Body.String(), body)
	}
	if got := response.Header().Get("Content-Disposition"); got != `attachment; filename="report.txt"` {
		t.Fatalf("Content-Disposition = %q", got)
	}
}

func TestStorageDownloadRejectsNonRegularFile(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	directory := filepath.Join(ExeDirPath, uploadDir, "download-id", "not-a-file")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodGet, "/storage/download/download-id/not-a-file", nil)
	response := httptest.NewRecorder()
	newStorageTestServer(t).ginEngine.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "invalid file") {
		t.Fatalf("status=%d body=%q, want 400 invalid file", response.Code, response.Body.String())
	}
}

func TestStorageDownloadReportsMissingFile(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	request := httptest.NewRequest(http.MethodGet, "/storage/download/download-id/missing.txt", nil)
	response := httptest.NewRecorder()
	newStorageTestServer(t).ginEngine.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), "file not exists") {
		t.Fatalf("status=%d body=%q, want 404 file not exists", response.Code, response.Body.String())
	}
}
