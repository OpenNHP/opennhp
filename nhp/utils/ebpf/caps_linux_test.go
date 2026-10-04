//go:build linux

package ebpf

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"kernel.org/pub/linux/libs/security/libcap/cap"
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

// The drop subtracts the loader's three capabilities; it must never assign a
// fixed set. nhp-serverd also runs as root in docker/Dockerfile.server, where
// the process holds the full root set for reasons that have nothing to do with
// this loader — CAP_DAC_OVERRIDE for volume-mounted config owned by another
// uid, CAP_NET_BIND_SERVICE for an http.toml reload that rebinds 80/443 — and
// loadXdpConfig reaches the drop on every path, including a host with no
// etc/xdp.toml that never asked to filter anything. Resetting to empty there
// would be a silent, irreversible privilege change outside the opt-in.
func TestReduceLoaderCapsOnlyTakesTheLoadersThree(t *testing.T) {
	// A stand-in for the root container: the loader's three plus two the
	// loader knows nothing about.
	held := cap.NewSet()
	for _, vec := range []cap.Flag{cap.Effective, cap.Permitted, cap.Inheritable} {
		if err := held.SetFlag(vec, true, cap.BPF, cap.NET_ADMIN, cap.PERFMON, cap.DAC_OVERRIDE, cap.NET_BIND_SERVICE); err != nil {
			t.Fatalf("build held set: %v", err)
		}
	}

	for _, tc := range []struct {
		name        string
		keepBpf     bool
		wantKeptBpf bool
		gone        []cap.Value
		kept        []cap.Value
	}{
		{
			name: "bpf given back too",
			gone: []cap.Value{cap.BPF, cap.NET_ADMIN, cap.PERFMON},
			kept: []cap.Value{cap.DAC_OVERRIDE, cap.NET_BIND_SERVICE},
		},
		{
			name:        "bpf kept for the reload path",
			keepBpf:     true,
			wantKeptBpf: true,
			gone:        []cap.Value{cap.NET_ADMIN, cap.PERFMON},
			kept:        []cap.Value{cap.BPF, cap.DAC_OVERRIDE, cap.NET_BIND_SERVICE},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, keptBpf, err := reduceLoaderCaps(held, tc.keepBpf)
			if err != nil {
				t.Fatalf("reduceLoaderCaps: %v", err)
			}
			if keptBpf != tc.wantKeptBpf {
				t.Errorf("keptBpf = %v, want %v", keptBpf, tc.wantKeptBpf)
			}
			for _, vec := range []cap.Flag{cap.Effective, cap.Permitted, cap.Inheritable} {
				for _, val := range tc.gone {
					if has, err := want.GetFlag(vec, val); err != nil || has {
						t.Errorf("%v still holds %v (err=%v)", vec, val, err)
					}
				}
				for _, val := range tc.kept {
					if has, err := want.GetFlag(vec, val); err != nil || !has {
						t.Errorf("%v lost %v, which the loader never asked for (err=%v)", vec, val, err)
					}
				}
			}
		})
	}
}

// On the demo unit the grant *is* the loader's three (AmbientCapabilities= on
// an otherwise unprivileged user), so subtracting them has to leave a process
// holding nothing — that is what deploy-server's "Verify the XDP loader gave
// its capabilities back" reads out of /proc/<pid>/task/*/status.
func TestReduceLoaderCapsLeavesTheDemoUnitWithNothing(t *testing.T) {
	held := cap.NewSet()
	for _, vec := range []cap.Flag{cap.Effective, cap.Permitted, cap.Inheritable} {
		if err := held.SetFlag(vec, true, cap.BPF, cap.NET_ADMIN, cap.PERFMON); err != nil {
			t.Fatalf("build held set: %v", err)
		}
	}

	want, _, err := reduceLoaderCaps(held, false)
	if err != nil {
		t.Fatalf("reduceLoaderCaps: %v", err)
	}
	if got := want.String(); got != cap.NewSet().String() {
		t.Errorf("reduced set is %q, want the empty set %q", got, cap.NewSet())
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
