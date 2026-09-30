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

// A missing xdp.toml is the un-opted-in case: loadXdpConfig reports it,
// attaches nothing and installs no watch. The filter must never appear on a
// host whose operator never asked for it — it drops every inbound TCP service
// and leaves no way back in but the whitelist.
func TestLoadXdpConfigMissingFileIsNotFatal(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() {
		ExeDirPath = originalExeDir
		xdpConfigWatch = nil
	})

	s := &UdpServer{config: &Config{}}
	if err := s.loadXdpConfig(1); err == nil {
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
	if err := s.loadXdpConfig(1); err == nil {
		t.Error("loadXdpConfig on a truncated file returned nil, want an error")
	}
	if xdpConfigWatch != nil {
		t.Error("a watch was installed for a file that failed to parse")
	}
}

// Nothing is attached when the file itself does not ask for it, and nothing is
// watched either: with no filter running there is no map for a reload to write
// into, and a watch would only suggest otherwise.
//
// The empty-RelayIPs case is the one that matters most. It is what an unset
// RELAY_IPS in the deploy pipeline renders, and attaching with it would close
// tcp/22 for every source on a host with no break-glass path.
func TestStartXdpFilterRefusesUnlessTheFileAsksForIt(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	tests := []struct {
		name string
		conf XdpTomlConfig
	}{
		{"disabled", XdpTomlConfig{Enabled: false, RelayIPs: []string{"1.2.3.4"}}},
		{"no relay ips", XdpTomlConfig{Enabled: true}},
		{"empty relay ips", XdpTomlConfig{Enabled: true, RelayIPs: []string{}}},
		{"floor out of range", XdpTomlConfig{Enabled: true, RelayIPs: []string{"1.2.3.4"}, NhpMinFrameBytes: 70000}},
		// A list that is non-empty as strings but names nothing the kernel can
		// hold is the same lockout as an empty one, and it is what the obvious
		// mistakes render to.
		{"an unset RELAY_IPS", XdpTomlConfig{Enabled: true, RelayIPs: []string{""}}},
		{"a hostname", XdpTomlConfig{Enabled: true, RelayIPs: []string{"relay.opennhp.org"}}},
		{"an inline comment", XdpTomlConfig{Enabled: true, RelayIPs: []string{"10.0.1.4 # relay"}}},
		{"a typo", XdpTomlConfig{Enabled: true, RelayIPs: []string{"10.0.1.300"}}},
		{"IPv6 only", XdpTomlConfig{Enabled: true, RelayIPs: []string{"2001:db8::1"}}},
		// Half a whitelist is refused too: the entry that failed to parse may
		// be the one SSH arrives from, and attaching decides that blind.
		{"one bad entry among good ones", XdpTomlConfig{Enabled: true, RelayIPs: []string{"10.0.1.0/24", "not-an-ip"}}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &UdpServer{config: &Config{ListenPort: 62206}}
			conf := tc.conf
			if s.startXdpFilter(&conf, 1) {
				t.Error("startXdpFilter attached the filter, want a refusal")
			}
		})
	}
}

// The filter drops every inbound TCP flow it was not told about, so attaching
// it in front of one of the daemon's own listeners would silently take that
// service off the air. Loopback binds are not reachable through the filtered
// interface at all, so they are not a conflict.
func TestXdpServiceConflictNamesListenersTheFilterWouldBlackHole(t *testing.T) {
	tests := []struct {
		name         string
		http         *HttpConfig
		metrics      MetricsConfig
		wantConflict bool
	}{
		{name: "no listeners", wantConflict: false},
		{
			name:         "http on all interfaces",
			http:         &HttpConfig{EnableHttp: true, HttpListenIp: "0.0.0.0", HttpListenPort: 443},
			wantConflict: true,
		},
		{
			name:         "http with an unset bind address is a wildcard bind",
			http:         &HttpConfig{EnableHttp: true, HttpListenPort: 443},
			wantConflict: true,
		},
		{
			name:         "http on loopback only",
			http:         &HttpConfig{EnableHttp: true, HttpListenIp: "127.0.0.1", HttpListenPort: 8443},
			wantConflict: false,
		},
		{
			name:         "http disabled",
			http:         &HttpConfig{EnableHttp: false, HttpListenIp: "0.0.0.0", HttpListenPort: 443},
			wantConflict: false,
		},
		{
			name:         "metrics off-host",
			metrics:      MetricsConfig{Enabled: true, ListenIp: "0.0.0.0", ListenPort: 9100},
			wantConflict: true,
		},
		{
			name:         "metrics on loopback (the default)",
			metrics:      MetricsConfig{Enabled: true, ListenIp: "127.0.0.1"},
			wantConflict: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &UdpServer{
				config:     &Config{ListenPort: 62206, Metrics: tc.metrics},
				httpConfig: tc.http,
			}
			got := s.xdpServiceConflict()
			if tc.wantConflict && got == "" {
				t.Error("xdpServiceConflict = \"\", want a conflict")
			}
			if !tc.wantConflict && got != "" {
				t.Errorf("xdpServiceConflict = %q, want no conflict", got)
			}
		})
	}
}

// The reload path has the same decision to make as startup with a worse
// failure mode: applying a list that parses as TOML but not as addresses does
// not merely refuse to attach, it sweeps the working whitelist out of a filter
// that is already enforcing — closing tcp/22 on an operator who is holding the
// only session that could fix it. So a reload only overwrites the live map when
// every entry in the file is a usable prefix.
func TestApplyXdpConfigKeepsTheActiveWhitelistOnAnUnusableList(t *testing.T) {
	originalExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = originalExeDir })

	tests := []struct {
		name  string
		ips   []string
		apply bool
	}{
		{name: "a working list", ips: []string{"10.0.1.0/24", "203.0.113.7"}, apply: true},
		{name: "no relay ips", ips: nil},
		{name: "an unset RELAY_IPS", ips: []string{""}},
		{name: "a hostname", ips: []string{"relay.opennhp.org"}},
		{name: "an inline comment", ips: []string{"10.0.1.4 # relay"}},
		{name: "a typo", ips: []string{"10.0.1.300"}},
		{name: "IPv6 only", ips: []string{"2001:db8::1"}},
		{name: "one bad entry among good ones", ips: []string{"10.0.1.0/24", "not-an-ip"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &UdpServer{config: &Config{ListenPort: 62206}}
			conf := XdpTomlConfig{Enabled: true, RelayIPs: tc.ips}
			if got := s.applyXdpConfig(&conf); got != tc.apply {
				t.Errorf("applyXdpConfig(%v) = %v, want %v", tc.ips, got, tc.apply)
			}
		})
	}
}

// An opted-in file on a host with no compiled object is the fail-open path: the
// load fails, the daemon carries on unfiltered, and — because there is no map
// to mirror the whitelist into — no watch is left behind.
func TestLoadXdpConfigFailsOpenWithoutTheObject(t *testing.T) {
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
	body := "Enabled = true\nNhpMinFrameBytes = 240\nRelayIPs = [\"1.2.3.4\", \"10.0.1.0/24\"]\n"
	if err := os.WriteFile(filepath.Join(etcDir, "xdp.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	s := &UdpServer{config: &Config{ListenPort: 62206}}
	if err := s.loadXdpConfig(1); err != nil {
		t.Fatalf("loadXdpConfig: %v", err)
	}
	if xdpConfigWatch != nil {
		t.Error("a watch was installed although no filter is attached")
	}
}
