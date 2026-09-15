package db

import (
	"os"
	"path/filepath"
	"testing"
)

func dbWriteTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "cfg.toml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return p
}

// The nhp-db config helpers run from a file-watch callback; malformed or
// missing input must be logged and tolerated, not fatal. These cover those
// error paths.
func TestDBConfigUpdates_MalformedAndMissing(t *testing.T) {
	d := &UdpDevice{config: &Config{}}
	const malformed = "][ = not = valid = toml"

	_ = d.updateBaseConfig(dbWriteTemp(t, malformed))
	_ = d.updateServerPeers(dbWriteTemp(t, malformed))

	missing := filepath.Join(t.TempDir(), "does-not-exist.toml")
	_ = d.updateBaseConfig(missing)
	_ = d.updateServerPeers(missing)
}

func TestDBConfigUpdates_ValidMinimal(t *testing.T) {
	d := &UdpDevice{config: &Config{}}
	if err := d.updateBaseConfig(dbWriteTemp(t, "LogLevel = 0\nDefaultCipherScheme = 0\n")); err != nil {
		t.Errorf("updateBaseConfig(valid) returned %v", err)
	}
	// An empty server list is itself invalid for nhp-db (it needs at least one
	// [[Servers]] entry); just exercise the path, the error is expected.
	_ = d.updateServerPeers(dbWriteTemp(t, ""))
}
