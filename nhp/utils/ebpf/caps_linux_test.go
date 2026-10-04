//go:build linux

package ebpf

import (
	"os"
	"path/filepath"
	"testing"
)

// The sysctl decides whether the loader may give CAP_BPF back or has to keep
// it so that xdp.toml reloads can still write the relay whitelist. Getting it
// backwards is either a pointless capability held for the life of an
// internet-facing daemon, or a reload path that fails with EPERM on exactly
// the hosts the demo runs on.
func TestBpfSyscallNeedsCapabilityFrom(t *testing.T) {
	cases := []struct {
		name    string
		content string
		// write == false means "the file is not there at all", which is what a
		// kernel without the knob looks like.
		write bool
		want  bool
	}{
		{name: "unprivileged bpf allowed", content: "0\n", write: true, want: false},
		{name: "disabled", content: "1\n", write: true, want: true},
		{name: "disabled without recourse", content: "2\n", write: true, want: true},
		{name: "no trailing newline", content: "0", write: true, want: false},
		{name: "unreadable is assumed privileged", want: true},
		{name: "garbage is assumed privileged", content: "sometimes\n", write: true, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "unprivileged_bpf_disabled")
			if tc.write {
				if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
					t.Fatalf("write: %v", err)
				}
			}
			if got := bpfSyscallNeedsCapabilityFrom(path); got != tc.want {
				t.Fatalf("bpfSyscallNeedsCapabilityFrom(%q) = %v, want %v", tc.content, got, tc.want)
			}
		})
	}
}

// Dropping must be safe to call from a process that holds nothing to drop —
// every developer machine, every CI runner, and any host where the unit's
// AmbientCapabilities= did not take. A failure there would be logged as "the
// daemon keeps CAP_BPF, CAP_NET_ADMIN and CAP_PERFMON", which would be a lie.
func TestDropLoaderPrivilegesIsIdempotent(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: dropping here would disarm the rest of the test binary")
	}
	if err := dropLoaderPrivileges(); err != nil {
		t.Fatalf("first drop: %v", err)
	}
	if err := dropLoaderPrivileges(); err != nil {
		t.Fatalf("second drop: %v", err)
	}
}
