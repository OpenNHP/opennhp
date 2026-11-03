package agent

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

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

// TestMemoryModeRebuildBuildsFreshConfig verifies applyInMemoryBaseConfig does
// not alias the template into a.config: a rekey that mutates a.config in place
// must not reach back into memConfig. (The template is still updated on rekey,
// but only through ReinitWithKey's explicit updateMemConfig write, not by
// aliasing — see TestMemoryModeRestartKeepsRekey.)
func TestMemoryModeRebuildBuildsFreshConfig(t *testing.T) {
	a := &UdpAgent{log: log.NewLogger("test", 0, t.TempDir(), "agent")}
	k1 := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	a.UseInMemoryConfig(&Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    base64.StdEncoding.EncodeToString(k1),
	})

	if err := a.loadBaseConfig(); err != nil {
		t.Fatalf("loadBaseConfig: %v", err)
	}
	if a.config == a.memConfig {
		t.Fatal("a.config aliases the template; a rekey would mutate memConfig")
	}
	// A direct in-place mutation of a.config (what SetPrivateKeyMaterial does)
	// must not touch the template.
	k2 := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	a.config.SetPrivateKeyMaterial(base64.StdEncoding.EncodeToString(k2), common.CIPHER_SCHEME_GMSM, k2)
	if a.memConfig.PrivateKeyBase64 != base64.StdEncoding.EncodeToString(k1) ||
		a.memConfig.DefaultCipherScheme != common.CIPHER_SCHEME_CURVE {
		t.Fatalf("in-place config mutation leaked into the template: %+v", a.memConfig)
	}
}

// TestMemoryModeRestartKeepsRekey verifies the fix for the reviewer's finding 3:
// a rekey via ReinitWithKey is recorded in the template, so a Stop/Start restart
// (which clears a.config and rebuilds from the template) keeps the rotated key
// and cipher scheme instead of reverting to the originally supplied one.
func TestMemoryModeRestartKeepsRekey(t *testing.T) {
	a := &UdpAgent{log: log.NewLogger("test", 0, t.TempDir(), "agent")}
	k1 := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	a.UseInMemoryConfig(&Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    base64.StdEncoding.EncodeToString(k1),
	})
	if err := a.loadBaseConfig(); err != nil {
		t.Fatalf("loadBaseConfig: %v", err)
	}

	// Rotate the in-memory key the way ReinitWithKey records it.
	k2 := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	k2b64 := base64.StdEncoding.EncodeToString(k2)
	a.updateMemConfig(func(c *Config) {
		c.PrivateKeyBase64 = k2b64
		c.DefaultCipherScheme = common.CIPHER_SCHEME_GMSM
	})

	// Simulate a restart: clear a.config and reload from the template.
	a.config = nil
	if err := a.loadBaseConfig(); err != nil {
		t.Fatalf("loadBaseConfig (restart): %v", err)
	}
	if a.config.PrivateKeyBase64 != k2b64 || a.config.GetCipherScheme() != common.CIPHER_SCHEME_GMSM {
		t.Fatalf("restart reverted the rekey instead of keeping it: %+v", a.config)
	}
}

// TestMemoryModeSetKnockUserSurvivesRestart verifies SetKnockUser records the
// identity in the template (allocating it in the nil-template SDK path), so a
// restart does not reset knockUser to empty.
func TestMemoryModeSetKnockUserSurvivesRestart(t *testing.T) {
	a := &UdpAgent{log: log.NewLogger("test", 0, t.TempDir(), "agent")}
	a.SetConfigSource(ConfigSourceMemory) // nil template, SDK install-after-Start path
	a.knockUser = &KnockUser{}
	a.SetKnockUser("sdk-user", "sdk-org", map[string]any{"k": "v"})

	if a.memConfig == nil {
		t.Fatal("SetKnockUser did not allocate the template in memory mode")
	}

	// Restart: clear config and reload from the template.
	a.config = nil
	if err := a.loadBaseConfig(); err != nil {
		t.Fatalf("loadBaseConfig (restart): %v", err)
	}
	a.knockUserMutex.RLock()
	defer a.knockUserMutex.RUnlock()
	if a.knockUser.UserId != "sdk-user" || a.knockUser.OrganizationId != "sdk-org" {
		t.Fatalf("restart lost the programmatic identity: %+v", a.knockUser)
	}
}

// TestMemoryModeRotationNotSupported verifies the fix for finding 2: the
// file-backed rotation paths refuse in memory mode instead of writing a key to
// disk that the running agent never uses and reporting success.
func TestMemoryModeRotationNotSupported(t *testing.T) {
	a := &UdpAgent{}
	a.SetConfigSource(ConfigSourceMemory)

	if err := a.RotateAgentKey(); err != errMemoryModeNotSupported {
		t.Fatalf("RotateAgentKey: got %v, want errMemoryModeNotSupported", err)
	}
	if err := a.RotateTeeKey(); err != errMemoryModeNotSupported {
		t.Fatalf("RotateTeeKey: got %v, want errMemoryModeNotSupported", err)
	}
}

// TestMemoryModeStartRestartKeepsPeers verifies the fix for finding 3's peer
// loss: a memory-mode restart keeps peers added with AddServer rather than
// wiping serverClusterMap and skipping loadPeers.
func TestMemoryModeStartRestartKeepsPeers(t *testing.T) {
	dir := t.TempDir()
	key := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	resetConfigWatches()

	a := &UdpAgent{}
	a.UseInMemoryConfig(&Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    base64.StdEncoding.EncodeToString(key),
	})
	if err := a.Start(dir, 0); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer a.Stop()

	peerKey := core.NewECDH(core.ECC_CURVE25519).PublicKeyBase64()
	a.AddServer(&core.UdpPeer{
		Type:         int(core.NHP_SERVER),
		PubKeyBase64: peerKey,
		Ip:           "203.0.113.10",
		Port:         62206,
	})
	if _, ok := a.serverClusterMap[peerKey]; !ok {
		t.Fatalf("AddServer did not register the peer")
	}

	// Simulate RestartAgent's Stop/Start without the hardcoded ExeDirPath.
	a.Stop()
	a.config = nil
	if err := a.Start(dir, 0); err != nil {
		t.Fatalf("Start (restart): %v", err)
	}
	if _, ok := a.serverClusterMap[peerKey]; !ok {
		t.Fatalf("restart wiped the programmatic peer")
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

// TestMemoryModeSuppliedConfigRequiresKey verifies that a config handed to
// UseInMemoryConfig with an empty key fails Start instead of knocking with a
// throwaway key, and that SetAllowMissingConfig does not widen memory mode.
func TestMemoryModeSuppliedConfigRequiresKey(t *testing.T) {
	resetConfigWatches()
	a := &UdpAgent{}
	a.SetAllowMissingConfig(true)
	a.UseInMemoryConfig(&Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		KnockUser:           KnockUser{UserId: "from-code"},
	})
	err := a.Start(t.TempDir(), 0)
	a.log.Close()
	if err == nil {
		t.Fatal("Start succeeded with an empty key in a supplied in-memory config")
	}
	if a.device != nil {
		t.Fatal("Start created a device despite refusing the empty key")
	}
}

// TestMemoryModeRestartRequiresInstalledKey verifies the install-after-Start
// contract: the first Start with no config runs on a throwaway key, but a
// restart before ReinitWithKey fails instead of coming back with a different
// random key.
func TestMemoryModeRestartRequiresInstalledKey(t *testing.T) {
	dir := t.TempDir()
	resetConfigWatches()
	a := &UdpAgent{}
	a.SetConfigSource(ConfigSourceMemory)
	if err := a.Start(dir, 0); err != nil {
		t.Fatalf("first Start with no config: %v", err)
	}
	a.SetKnockUser("sdk-user", "sdk-org", nil) // allocates memConfig, still no key

	a.Stop()
	a.config = nil
	err := a.Start(dir, 0)
	a.log.Close()
	if err == nil {
		t.Fatal("restart succeeded with no key installed")
	}
}

// TestMemoryModeRestartAfterReinitWithKey verifies that once ReinitWithKey has
// installed the real key, a restart comes back with that key.
func TestMemoryModeRestartAfterReinitWithKey(t *testing.T) {
	dir := t.TempDir()
	resetConfigWatches()
	a := &UdpAgent{}
	a.SetConfigSource(ConfigSourceMemory)
	if err := a.Start(dir, 0); err != nil {
		t.Fatalf("first Start with no config: %v", err)
	}
	key := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	wantPub := core.ECDHFromKey(core.ECC_CURVE25519, key).PublicKeyBase64()
	if err := a.ReinitWithKey(key, common.CIPHER_SCHEME_CURVE); err != nil {
		t.Fatalf("ReinitWithKey: %v", err)
	}

	a.Stop()
	a.config = nil
	if err := a.Start(dir, 0); err != nil {
		t.Fatalf("restart after ReinitWithKey: %v", err)
	}
	defer a.Stop()
	if got := a.PublicKeyBase64(); got != wantPub {
		t.Fatalf("restart did not keep the installed key: got %q want %q", got, wantPub)
	}
}

// TestRestartHandlerReportsFailureOnly verifies that a failed restart answers
// with the 500 alone, not a 500 followed by the success message.
func TestRestartHandlerReportsFailureOnly(t *testing.T) {
	resetConfigWatches()
	a := &UdpAgent{}
	a.SetConfigSource(ConfigSourceMemory)
	if err := a.Start(t.TempDir(), 0); err != nil {
		t.Fatalf("first Start with no config: %v", err)
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	a.restartAgent(c) // no key installed, so the restart fails
	a.log.Close()

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", w.Code, http.StatusInternalServerError)
	}
	if body := w.Body.String(); strings.Contains(body, "successfully restart agent") {
		t.Fatalf("failed restart also reported success: %s", body)
	}
}

// TestStopAfterFailedRestartIsSafe verifies that a failed restart leaves an
// agent that Stop and another RestartAgent can be called on without panicking.
func TestStopAfterFailedRestartIsSafe(t *testing.T) {
	resetConfigWatches()
	a := &UdpAgent{}
	a.SetConfigSource(ConfigSourceMemory)
	if err := a.Start(t.TempDir(), 0); err != nil {
		t.Fatalf("first Start with no config: %v", err)
	}
	if err := a.RestartAgent(); err == nil {
		t.Fatal("restart succeeded with no key installed")
	}
	if a.IsRunning() {
		t.Fatal("agent reports running after a failed restart")
	}
	if err := a.RestartAgent(); err == nil {
		t.Fatal("second restart succeeded with no key installed")
	}
	a.Stop()
	a.log.Close()
}

// TestGetServerConfigNotSupportedInMemoryMode verifies GET /config/server
// returns 501 in memory mode instead of reporting an etc/server.toml the agent
// does not use, while file mode still serves the file.
func TestGetServerConfigNotSupportedInMemoryMode(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "etc"), 0700); err != nil {
		t.Fatal(err)
	}
	bait := "[[Servers]]\nName = \"from-file\"\nPubKeyBase64 = \"bait-pubkey-from-disk\"\n"
	if err := os.WriteFile(filepath.Join(dir, "etc", "server.toml"), []byte(bait), 0600); err != nil {
		t.Fatal(err)
	}
	prev := ExeDirPath
	ExeDirPath = dir
	defer func() { ExeDirPath = prev }()

	gin.SetMode(gin.TestMode)
	get := func(a *UdpAgent) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		a.getServerConfig(c)
		return w
	}

	mem := &UdpAgent{}
	mem.SetConfigSource(ConfigSourceMemory)
	w := get(mem)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("memory mode status: got %d want %d", w.Code, http.StatusNotImplemented)
	}
	if strings.Contains(w.Body.String(), "bait-pubkey-from-disk") {
		t.Fatalf("memory mode served etc/server.toml: %s", w.Body.String())
	}

	w = get(&UdpAgent{})
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "bait-pubkey-from-disk") {
		t.Fatalf("file mode did not serve etc/server.toml: %d %s", w.Code, w.Body.String())
	}
}

// TestRotateHandlersNotSupportedInMemoryMode verifies both rotate endpoints
// answer the memory-mode refusal with 501, like the /config/server handlers.
func TestRotateHandlersNotSupportedInMemoryMode(t *testing.T) {
	a := &UdpAgent{}
	a.SetConfigSource(ConfigSourceMemory)

	gin.SetMode(gin.TestMode)
	for name, h := range map[string]gin.HandlerFunc{
		"rotateAgentKey": a.rotateAgentKey,
		"rotateTeeKey":   a.rotateTeeKey,
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		h(c)
		if w.Code != http.StatusNotImplemented {
			t.Fatalf("%s status: got %d want %d", name, w.Code, http.StatusNotImplemented)
		}
		if !strings.Contains(w.Body.String(), errMemoryModeNotSupported.Error()) {
			t.Fatalf("%s body: %s", name, w.Body.String())
		}
	}
}
