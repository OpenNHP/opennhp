package agent

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// resetConfigWatches clears the package-level watch handles so a memory-mode
// test can assert that no watcher was armed, independent of any file-mode test
// that ran earlier in the same process.
func resetConfigWatches() {
	baseConfigWatch = nil
	dhpConfigWatch = nil
	serverConfigWatch = nil
	resourceConfigWatch = nil
}

func configWatchesArmed() bool {
	return baseConfigWatch != nil || dhpConfigWatch != nil ||
		serverConfigWatch != nil || resourceConfigWatch != nil
}

// TestMemoryModeLoadersArmNoWatcher verifies that in ConfigSourceMemory the
// four config loaders apply configuration without arming any file watcher, so
// nothing on disk can later override what was set in code.
func TestMemoryModeLoadersArmNoWatcher(t *testing.T) {
	resetConfigWatches()
	a := &UdpAgent{log: log.NewLogger("test", 0, t.TempDir(), "agent")}
	a.SetConfigSource(ConfigSourceMemory)

	for _, step := range []struct {
		name string
		fn   func() error
	}{
		{"loadBaseConfig", a.loadBaseConfig},
		{"loadDHPConfig", a.loadDHPConfig},
		{"loadPeers", a.loadPeers},
		{"loadResources", a.loadResources},
	} {
		if err := step.fn(); err != nil {
			t.Fatalf("%s returned error in memory mode: %v", step.name, err)
		}
	}

	if configWatchesArmed() {
		t.Fatal("memory mode armed a config watcher")
	}
	if a.config == nil || a.config.DHPConfig == nil {
		t.Fatal("memory mode left config or DHPConfig nil")
	}
}

// TestUseInMemoryConfigAppliesIdentity confirms the programmatic base config is
// applied (knock user and cipher scheme) with no file read.
func TestUseInMemoryConfigAppliesIdentity(t *testing.T) {
	resetConfigWatches()
	a := &UdpAgent{log: log.NewLogger("test", 0, t.TempDir(), "agent")}
	a.UseInMemoryConfig(&Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_GMSM,
		KnockUser:           KnockUser{UserId: "from-code", OrganizationId: "org-code"},
	})

	if err := a.loadBaseConfig(); err != nil {
		t.Fatalf("loadBaseConfig: %v", err)
	}
	if a.config.GetCipherScheme() != common.CIPHER_SCHEME_GMSM {
		t.Fatalf("cipher scheme not applied: got %d", a.config.GetCipherScheme())
	}
	a.knockUserMutex.RLock()
	defer a.knockUserMutex.RUnlock()
	if a.knockUser == nil || a.knockUser.UserId != "from-code" || a.knockUser.OrganizationId != "org-code" {
		t.Fatalf("knock identity not applied: %+v", a.knockUser)
	}
}

// TestUseInMemoryConfigDoesNotShareCallerConfig confirms the agent keeps its own
// copy of the supplied data, so later mutation of the caller's struct does not
// reach the agent.
func TestUseInMemoryConfigDoesNotShareCallerConfig(t *testing.T) {
	a := &UdpAgent{}
	caller := &Config{KnockUser: KnockUser{UserId: "original"}, DefaultCipherScheme: common.CIPHER_SCHEME_CURVE}
	a.UseInMemoryConfig(caller)

	caller.UserId = "mutated"
	caller.DefaultCipherScheme = common.CIPHER_SCHEME_GMSM

	if a.memConfig.UserId != "original" || a.memConfig.DefaultCipherScheme != common.CIPHER_SCHEME_CURVE {
		t.Fatalf("agent observed caller mutation: %+v", a.memConfig)
	}
}

// TestMemoryModeRebuildsConfigFromTemplate verifies that applyInMemoryBaseConfig
// builds a fresh Config each call, so a rekey that mutates the live config does
// not alter the stored template and a Stop/Start restart begins from the
// originally supplied configuration.
func TestMemoryModeRebuildsConfigFromTemplate(t *testing.T) {
	a := &UdpAgent{log: log.NewLogger("test", 0, t.TempDir(), "agent")}
	k1 := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	a.UseInMemoryConfig(&Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    base64.StdEncoding.EncodeToString(k1),
	})

	if err := a.loadBaseConfig(); err != nil {
		t.Fatalf("loadBaseConfig: %v", err)
	}
	// Simulate a rekey, which mutates the live config in place.
	k2 := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	a.config.SetPrivateKeyMaterial(base64.StdEncoding.EncodeToString(k2), common.CIPHER_SCHEME_GMSM, k2)

	// The template must be untouched by the rekey.
	if a.memConfig.PrivateKeyBase64 != base64.StdEncoding.EncodeToString(k1) ||
		a.memConfig.DefaultCipherScheme != common.CIPHER_SCHEME_CURVE {
		t.Fatalf("rekey mutated the stored template: %+v", a.memConfig)
	}

	// A restart reloads from the template, not the rekeyed live config.
	if err := a.loadBaseConfig(); err != nil {
		t.Fatalf("loadBaseConfig (restart): %v", err)
	}
	if a.config.PrivateKeyBase64 != base64.StdEncoding.EncodeToString(k1) ||
		a.config.GetCipherScheme() != common.CIPHER_SCHEME_CURVE {
		t.Fatalf("restart did not start from the original config: %+v", a.config)
	}
}

// TestMemoryModeStartIgnoresConfigFile is the security invariant the maintainer
// asked for: an agent started in memory mode uses the in-code identity and key,
// arms no watcher, and ignores a config.toml that exists on disk (one that would
// otherwise set a different identity and cipher scheme in file mode).
func TestMemoryModeStartIgnoresConfigFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "etc"), 0700); err != nil {
		t.Fatal(err)
	}
	// A config.toml that, in file mode, would set a different identity/scheme.
	onDisk := "" +
		"logLevel = 0\n" +
		"defaultCipherScheme = 1\n" +
		"privateKey = \"\"\n" +
		"UserId = \"from-file\"\n"
	if err := os.WriteFile(filepath.Join(dir, "etc", "config.toml"), []byte(onDisk), 0600); err != nil {
		t.Fatal(err)
	}

	key := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	wantPub := core.ECDHFromKey(core.ECC_CURVE25519, key).PublicKeyBase64()

	resetConfigWatches()
	a := &UdpAgent{}
	a.UseInMemoryConfig(&Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    base64.StdEncoding.EncodeToString(key),
		KnockUser:           KnockUser{UserId: "from-code"},
	})
	if err := a.Start(dir, 0); err != nil {
		t.Fatalf("Start in memory mode: %v", err)
	}
	defer a.Stop()

	if configWatchesArmed() {
		t.Fatal("memory-mode Start armed a config watcher")
	}
	if got := a.PublicKeyBase64ByCipherScheme(); got != wantPub {
		t.Fatalf("public key not from in-code key: got %q want %q", got, wantPub)
	}
	if a.config.GetCipherScheme() != common.CIPHER_SCHEME_CURVE {
		t.Fatalf("cipher scheme came from file: got %d", a.config.GetCipherScheme())
	}
	a.knockUserMutex.RLock()
	gotUser := a.knockUser.UserId
	a.knockUserMutex.RUnlock()
	if gotUser != "from-code" {
		t.Fatalf("identity came from file: got %q", gotUser)
	}
}
