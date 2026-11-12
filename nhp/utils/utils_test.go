package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// httptest serves on loopback, which the default limits refuse. These two
// tests are about the temp-directory bookkeeping, not about the address
// guard, so they opt loopback back in the same way a deployment would.
func loopbackLimits() DownloadLimits {
	lim := DefaultDownloadLimits()
	lim.AllowLoopback = true
	return lim
}

func TestDownloadFileToTempCleansPartialDownloadOnError(t *testing.T) {
	tempRoot := t.TempDir()
	t.Setenv("TMPDIR", tempRoot)
	t.Setenv("TMP", tempRoot)
	t.Setenv("TEMP", tempRoot)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	defer server.Close()

	if _, err := DownloadFileToTempWithLimits(context.Background(), server.URL+"/policy.wasm", "wasm-test-", loopbackLimits()); err == nil {
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
	t.Setenv("TMP", tempRoot)
	t.Setenv("TEMP", tempRoot)

	const body = "valid wasm placeholder"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	path, err := DownloadFileToTempWithLimits(context.Background(), server.URL+"/policy.wasm", "wasm-test-", loopbackLimits())
	if err != nil {
		t.Fatalf("DownloadFileToTempWithLimits(): %v", err)
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
