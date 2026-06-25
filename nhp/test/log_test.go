package test

import (
	"testing"
	"time"

	log "github.com/OpenNHP/opennhp/nhp/log"
)

func TestLog(t *testing.T) {
	// Keep test fast for CI: exercise async logging without multi-second sleeps
	for i := 0; i < 3; i++ {
		log.Info("Info log test %d", i)
	}
	// Brief pause to allow async writer flush before Close
	time.Sleep(150 * time.Millisecond)
	log.Close()
}

func TestLoggerCloseIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	l := log.NewLogger("IdempotentTest", log.LogLevelInfo, tmpDir, "idempotent")
	l.Close()
	// Must not panic when Close is called repeatedly
	l.Close()
}
