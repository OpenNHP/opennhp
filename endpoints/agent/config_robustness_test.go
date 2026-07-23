package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "cfg.toml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return p
}

// The config update helpers run from a file-watch callback, so a malformed or
// truncated file written mid-save must not take the agent down — they log and
// keep the previous in-memory config. These tests exercise those error paths.
func TestAgentConfigUpdates_MalformedInput(t *testing.T) {
	a := &UdpAgent{config: &Config{}}
	const malformed = "this is ][ not = valid = toml"

	// none of these may panic on malformed input
	_ = a.updateBaseConfig(writeTempFile(t, malformed))
	_ = a.updateDHPConfig(writeTempFile(t, malformed))
	_ = a.updateResources(writeTempFile(t, malformed))

	if err := a.updateServerPeers(writeTempFile(t, malformed)); err == nil {
		t.Error("updateServerPeers should return an error on malformed TOML")
	}
}

func TestAgentConfigUpdates_MissingFile(t *testing.T) {
	a := &UdpAgent{config: &Config{}}
	missing := filepath.Join(t.TempDir(), "does-not-exist.toml")

	_ = a.updateBaseConfig(missing)
	_ = a.updateDHPConfig(missing)
	_ = a.updateResources(missing)

	if err := a.updateServerPeers(missing); err == nil {
		t.Error("updateServerPeers should return an error when the file is missing")
	}
}

func TestAgentConfigUpdates_ValidMinimal(t *testing.T) {
	a := &UdpAgent{config: &Config{}}
	// A well-formed base config with no log-level change exercises the happy
	// path without needing a live logger or device.
	if err := a.updateBaseConfig(writeTempFile(t, "LogLevel = 0\nDefaultCipherScheme = 0\n")); err != nil {
		t.Errorf("updateBaseConfig(valid) returned %v", err)
	}
	if err := a.updateDHPConfig(writeTempFile(t, "teePrivateKeyBase64 = \"\"\n")); err != nil {
		t.Errorf("updateDHPConfig(valid) returned %v", err)
	}
	if err := a.updateResources(writeTempFile(t, "")); err != nil {
		t.Errorf("updateResources(empty) returned %v", err)
	}
}
