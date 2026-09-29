package server

import (
	"os"
	"path/filepath"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

// The XDP whitelist decides who may reach tcp/22 on the server host, and there
// is no break-glass path if it is rendered wrong — so the parsing and
// path-resolution around it are worth pinning down even though attaching the
// program itself needs a privileged kernel.

func TestXdpConfigFileNameDefaultsToEtc(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = "/opt/nhp-server"
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	s := &UdpServer{config: &Config{}}
	if got, want := s.xdpConfigFileName(), filepath.Join("/opt/nhp-server", "etc", "xdp.toml"); got != want {
		t.Errorf("default path = %q, want %q", got, want)
	}

	s.config.XdpConfigPath = "conf/alt.toml"
	if got, want := s.xdpConfigFileName(), filepath.Join("/opt/nhp-server", "conf", "alt.toml"); got != want {
		t.Errorf("relative override = %q, want %q", got, want)
	}

	s.config.XdpConfigPath = "/etc/nhp/xdp.toml"
	if got, want := s.xdpConfigFileName(), "/etc/nhp/xdp.toml"; got != want {
		t.Errorf("absolute override = %q, want %q", got, want)
	}
}

// The file the deploy pipeline renders must parse into the fields the server
// reads; a silent mismatch here would apply an empty whitelist.
func TestXdpTomlConfigParsesDeployedShape(t *testing.T) {
	content := []byte(`
Enabled = true
NhpMinFrameBytes = 240
RelayIPs = ["1.2.3.4", "5.6.7.8"]
`)

	var conf XdpTomlConfig
	if err := toml.Unmarshal(content, &conf); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !conf.Enabled {
		t.Error("Enabled = false, want true")
	}
	if conf.NhpMinFrameBytes != DefaultNhpMinFrameBytes {
		t.Errorf("NhpMinFrameBytes = %d, want %d", conf.NhpMinFrameBytes, DefaultNhpMinFrameBytes)
	}
	if len(conf.RelayIPs) != 2 || conf.RelayIPs[0] != "1.2.3.4" || conf.RelayIPs[1] != "5.6.7.8" {
		t.Errorf("RelayIPs = %v, want [1.2.3.4 5.6.7.8]", conf.RelayIPs)
	}
}

// A missing xdp.toml is the fail-open case: loadXdpConfig reports it and
// installs no watch, rather than applying an empty whitelist.
func TestLoadXdpConfigMissingFileIsNotFatal(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() {
		ExeDirPath = originalExeDir
		xdpConfigWatch = nil
	})

	s := &UdpServer{config: &Config{}}
	if err := s.loadXdpConfig(); err == nil {
		t.Error("loadXdpConfig on a missing file returned nil, want an error")
	}
	if xdpConfigWatch != nil {
		t.Error("a watch was installed for a file that does not exist")
	}
}

// A truncated or otherwise unparsable file must not reach the kernel: the
// active whitelist is what keeps the operator's SSH session alive.
func TestLoadXdpConfigRejectsUnparsableFile(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() {
		ExeDirPath = originalExeDir
		xdpConfigWatch = nil
	})

	etcDir := filepath.Join(ExeDirPath, "etc")
	if err := os.MkdirAll(etcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(etcDir, "xdp.toml"), []byte("RelayIPs = [\"1.2.3.4\""), 0o644); err != nil {
		t.Fatal(err)
	}

	s := &UdpServer{config: &Config{}}
	if err := s.loadXdpConfig(); err == nil {
		t.Error("loadXdpConfig on a truncated file returned nil, want an error")
	}
	if xdpConfigWatch != nil {
		t.Error("a watch was installed for a file that failed to parse")
	}
}

// The happy path: a well-formed file parses, applies (a no-op with no engine
// attached) and leaves a watch behind for hot reload.
func TestLoadXdpConfigInstallsWatch(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() {
		if xdpConfigWatch != nil {
			xdpConfigWatch.Close()
			xdpConfigWatch = nil
		}
		ExeDirPath = originalExeDir
	})

	etcDir := filepath.Join(ExeDirPath, "etc")
	if err := os.MkdirAll(etcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	body := "Enabled = true\nNhpMinFrameBytes = 240\nRelayIPs = [\"1.2.3.4\"]\n"
	if err := os.WriteFile(filepath.Join(etcDir, "xdp.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	s := &UdpServer{config: &Config{}}
	if err := s.loadXdpConfig(); err != nil {
		t.Fatalf("loadXdpConfig: %v", err)
	}
	if xdpConfigWatch == nil {
		t.Error("no watch installed for a valid xdp.toml")
	}
}
