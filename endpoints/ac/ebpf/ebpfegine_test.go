//go:build linux

package ebpf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAcProgObjPath(t *testing.T) {
	dir := t.TempDir()
	newPath := filepath.Join(dir, "nhp_ac_xdp.o")
	legacyPath := filepath.Join(dir, "nhp_ebpf_xdp.o")

	writeEmpty := func(path string) {
		t.Helper()
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", path, err)
		}
	}

	// Both present: the new name wins.
	writeEmpty(newPath)
	writeEmpty(legacyPath)
	if got := acProgObjPath(dir); got != newPath {
		t.Fatalf("both present: got %s, want %s", got, newPath)
	}

	// Only the legacy name present: fall back so the upgraded host still
	// starts, and report the legacy path to the operator.
	if err := os.Remove(newPath); err != nil {
		t.Fatalf("Remove(%s): %v", newPath, err)
	}
	if got := acProgObjPath(dir); got != legacyPath {
		t.Fatalf("legacy only: got %s, want %s", got, legacyPath)
	}

	// Neither present: report the new name, so the loader's "not found"
	// error names the path the daemon actually wants.
	if err := os.Remove(legacyPath); err != nil {
		t.Fatalf("Remove(%s): %v", legacyPath, err)
	}
	if got := acProgObjPath(dir); got != newPath {
		t.Fatalf("neither present: got %s, want %s", got, newPath)
	}
}
