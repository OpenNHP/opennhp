package server

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
)

func TestZdtoConfigRejectsInvalidDoID(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	invalidDoID := "foo/../../../bar"
	if err := SaveZdtoConfig(&common.DRGMsg{DoId: invalidDoID}); !errors.Is(err, common.ErrInvalidDoID) {
		t.Fatalf("SaveZdtoConfig error = %v, want ErrInvalidDoID", err)
	} else if err.Error() != common.ErrInvalidDoID.Error() {
		t.Fatalf("SaveZdtoConfig reflected rejected input: %q", err)
	}
	if _, err := ReadZdtoConfig(invalidDoID); !errors.Is(err, common.ErrInvalidDoID) {
		t.Fatalf("ReadZdtoConfig error = %v, want ErrInvalidDoID", err)
	}

	entries, err := os.ReadDir(ExeDirPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("invalid DoId changed the filesystem: %v", entries)
	}
}

func TestZdtoConfigValidDoIDRoundTrip(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	want := &common.DRGMsg{DoId: "550e8400-e29b-41d4-a716-446655440000"}
	if err := SaveZdtoConfig(want); err != nil {
		t.Fatalf("SaveZdtoConfig returned %v", err)
	}
	got, err := ReadZdtoConfig(want.DoId)
	if err != nil {
		t.Fatalf("ReadZdtoConfig returned %v", err)
	}
	if got.DoId != want.DoId {
		t.Fatalf("ReadZdtoConfig DoId = %q, want %q", got.DoId, want.DoId)
	}

	expectedPath := filepath.Join(ExeDirPath, "etc", "ztdo", "data-"+want.DoId+".json")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected config at %q: %v", expectedPath, err)
	}
}

func TestReadZdtoConfigScrubsFilesystemPath(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	_, err := ReadZdtoConfig("valid-but-missing")
	if !errors.Is(err, errReadConfigFailed) {
		t.Fatalf("ReadZdtoConfig error = %v, want errReadConfigFailed", err)
	}
	for _, leaked := range []string{ExeDirPath, "etc/ztdo", "data-valid-but-missing.json", "no such file"} {
		if strings.Contains(err.Error(), leaked) {
			t.Errorf("ReadZdtoConfig error %q leaked %q", err, leaked)
		}
	}
}
