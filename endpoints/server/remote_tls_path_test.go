package server

import (
	"path/filepath"
	"testing"
)

// Relative TLS file paths in remote.toml resolve against the install
// directory, not the process's working directory.
//
// remote.toml is read from ExeDirPath/etc, so "etc/etcd-ca.pem" written
// there means the file next to it. Passed through unchanged it would
// resolve against the working directory instead, and a daemon started by
// systemd without WorkingDirectory=, or in a container, would refuse to
// start on a path that looks correct. The same rule the XDP config and the
// audit ledger already follow.
func TestResolveRemoteConfigPath(t *testing.T) {
	saved := ExeDirPath
	t.Cleanup(func() { ExeDirPath = saved })
	ExeDirPath = filepath.Join("/opt", "nhp-server")

	abs := filepath.Join("/etc", "pki", "etcd-ca.pem")
	for in, want := range map[string]string{
		"etc/etcd-ca.pem": filepath.Join("/opt", "nhp-server", "etc", "etcd-ca.pem"),
		"etcd-ca.pem":     filepath.Join("/opt", "nhp-server", "etcd-ca.pem"),
		abs:               abs, // absolute paths are left alone
		"":                "",  // unset stays unset, so "is TLS configured" is unaffected
	} {
		if got := resolveRemoteConfigPath(in); got != want {
			t.Errorf("resolveRemoteConfigPath(%q) = %q, want %q", in, got, want)
		}
	}
}
