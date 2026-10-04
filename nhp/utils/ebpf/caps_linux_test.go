//go:build linux

package ebpf

import (
	"os"
	"path/filepath"
	"sync"
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
	if err := dropLoaderPrivileges(true); err != nil {
		t.Fatalf("first drop: %v", err)
	}
	if err := dropLoaderPrivileges(true); err != nil {
		t.Fatalf("second drop: %v", err)
	}
	// The failed-load form, which keeps nothing at all, has to be as safe to
	// call from an unprivileged process as the other one.
	if err := dropLoaderPrivileges(false); err != nil {
		t.Fatalf("drop without the CAP_BPF exemption: %v", err)
	}
}

// withDropRecorder swaps the drop for a recorder and resets the once-guard, so
// the exported wrapper can be exercised without actually disarming the test
// binary and without one test consuming the guard for the next.
func withDropRecorder(t *testing.T) *[]bool {
	t.Helper()

	origFn := dropLoaderPrivilegesFn
	t.Cleanup(func() {
		dropLoaderPrivilegesFn = origFn
		dropOnce = sync.Once{}
	})

	dropOnce = sync.Once{}
	var calls []bool
	dropLoaderPrivilegesFn = func(allowKeepBpf bool) { calls = append(calls, allowKeepBpf) }
	return &calls
}

// The loader and the caller both drop, and on the attach path both of them run.
// The loader goes first and is the one that knows there is a live map behind
// the CAP_BPF exemption, so its answer has to be the one that sticks: a second
// drop taking CAP_BPF away would leave the reload path returning EPERM on every
// distro that ships unprivileged_bpf_disabled non-zero, which is the one way a
// replaced relay's address reaches a live whitelist.
func TestDropLoaderPrivilegesHappensOnceFirstCallWins(t *testing.T) {
	calls := withDropRecorder(t)

	DropLoaderPrivileges(true)  // the loader, after a successful attach
	DropLoaderPrivileges(false) // the caller's unconditional defer

	if len(*calls) != 1 {
		t.Fatalf("dropped %d times, want exactly 1 (calls=%v)", len(*calls), *calls)
	}
	if !(*calls)[0] {
		t.Error("the surviving call was allowKeepBpf=false; the loader's answer must win on the attach path")
	}
}

// On every path that declines to filter, the caller's defer is the only drop
// there is — and it must keep nothing, because without a filter there is no
// whitelist map for the CAP_BPF exemption to protect.
func TestDropLoaderPrivilegesWithoutLoaderKeepsNothing(t *testing.T) {
	calls := withDropRecorder(t)

	DropLoaderPrivileges(false)

	if len(*calls) != 1 {
		t.Fatalf("dropped %d times, want exactly 1 (calls=%v)", len(*calls), *calls)
	}
	if (*calls)[0] {
		t.Error("allowKeepBpf=true on a path that never attached a filter")
	}
}
