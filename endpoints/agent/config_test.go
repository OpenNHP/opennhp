package agent

import (
	"path/filepath"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// TestUpdateBaseConfig_MissingConfigFailsByDefault: for run/dhp (the
// default, allowMissingConfig=false) a missing config.toml must surface as
// an error rather than silently yielding an empty config + throwaway key.
func TestUpdateBaseConfig_MissingConfigFailsByDefault(t *testing.T) {
	a := &UdpAgent{log: log.NewLogger("test", 0, t.TempDir(), "agent")}
	missing := filepath.Join(t.TempDir(), "does-not-exist.toml")

	if err := a.updateBaseConfig(missing); err == nil {
		t.Fatal("expected an error for a missing config when allowMissingConfig is false")
	}
}

// TestUpdateBaseConfig_MissingConfigToleratedForRegister: the register
// bootstrap flow opts in via SetAllowMissingConfig, so a missing config is
// accepted and produces an empty config.
func TestUpdateBaseConfig_MissingConfigToleratedForRegister(t *testing.T) {
	a := &UdpAgent{log: log.NewLogger("test", 0, t.TempDir(), "agent")}
	a.SetAllowMissingConfig(true)
	missing := filepath.Join(t.TempDir(), "does-not-exist.toml")

	if err := a.updateBaseConfig(missing); err != nil {
		t.Fatalf("expected missing config to be tolerated for register, got %v", err)
	}
	if a.config == nil || a.config.PrivateKeyBase64 != "" {
		t.Fatalf("expected an empty config with no private key, got %+v", a.config)
	}
}
