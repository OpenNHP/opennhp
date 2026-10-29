package resource

import (
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestLoadResourceDoesNotBlockOnFIFO(t *testing.T) {
	old := baseDir
	baseDir = t.TempDir()
	t.Cleanup(func() { baseDir = old })
	if err := syscall.Mkfifo(filepath.Join(baseDir, "pipe"), 0600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := loadResource("pipe"); done <- err }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("accepted FIFO")
		}
	case <-time.After(time.Second):
		t.Fatal("FIFO blocked resource read")
	}
}
