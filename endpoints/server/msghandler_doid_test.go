package server

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
)

const testOwner = "owner-db-public-key"

// testDRG is a registration that SaveZdtoConfig accepts: it carries a
// smart policy.
func testDRG(doId string) *common.DRGMsg {
	return &common.DRGMsg{DoId: doId, DbId: "db-1", Spo: common.SmartPolicy{Policy: "cG9saWN5"}}
}

func TestZdtoConfigRejectsInvalidDoID(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	invalidDoID := "foo/../../../bar"
	if err := SaveZdtoConfig(testDRG(invalidDoID), testOwner); !errors.Is(err, common.ErrInvalidDoID) {
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

	want := testDRG("550e8400-e29b-41d4-a716-446655440000")
	if err := SaveZdtoConfig(want, testOwner); err != nil {
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

func TestZdtoConfigConcurrentUpdatesRemainReadable(t *testing.T) {
	old := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = old })
	const id = "concurrent-update"
	first := testDRG(id)
	first.AccessUrl = "original"
	if err := SaveZdtoConfig(first, testOwner); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			for range 10 {
				if err := SaveZdtoConfig(testDRG(id), testOwner); err != nil {
					t.Error(err)
					return
				}
				got, err := ReadZdtoConfig(id)
				if err != nil || got.AccessUrl != "original" {
					t.Errorf("read during update = %+v, %v", got, err)
					return
				}
			}
		})
	}
	wg.Wait()
}

func TestZdtoConfigRenameFailureRemovesTemporaryFile(t *testing.T) {
	old := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = old })
	dir := filepath.Join(ExeDirPath, "etc", "ztdo")
	if err := os.MkdirAll(filepath.Join(dir, "data-object.json"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := SaveZdtoConfig(testDRG("object"), testOwner); !errors.Is(err, errSaveConfigFailed) {
		t.Fatalf("save error = %v", err)
	}
	files, err := filepath.Glob(filepath.Join(dir, ".data-*.json"))
	if err != nil || len(files) != 0 {
		t.Fatalf("temporary files = %v, error = %v", files, err)
	}
}
