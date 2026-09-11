package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFileToTempCleansPartialDownloadOnError(t *testing.T) {
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer server.Close()

	if _, err := DownloadFileToTemp(server.URL+"/policy.wasm", "wasm-test-"); err == nil {
		t.Fatal("DownloadFileToTemp() error = nil, want non-200 response error")
	}

	entries, err := os.ReadDir(tempRoot)
	if err != nil {
		t.Fatalf("ReadDir(%q): %v", tempRoot, err)
	}
	if len(entries) != 0 {
		t.Fatalf("partial download left %d entries under %q", len(entries), tempRoot)
	}
}

func TestDownloadFileToTempKeepsSuccessfulDownload(t *testing.T) {
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)

	const body = "valid wasm placeholder"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	path, err := DownloadFileToTemp(server.URL+"/policy.wasm", "wasm-test-")
	if err != nil {
		t.Fatalf("DownloadFileToTemp(): %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(path)) })

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	if string(got) != body {
		t.Fatalf("download body = %q, want %q", got, body)
	}
}
