//go:build linux

package ebpf

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// The capability numbers /proc/self/status reports them under.
const (
	capNetAdmin = 12
	capBPF      = 39
)

// hasCaps reports whether this process holds every one of the given
// capabilities.
//
// It exists because the tests in this package used to skip whenever a load
// failed, and two very different things look the same from the outside:
// "the kernel refused this program" and "this process is not allowed to load
// programs" are both an EPERM out of bpf(BPF_PROG_LOAD).
//
// A skip is the right answer for the second and the wrong one for the first.
// The first is the one that matters: a program the verifier rejects is an
// outage on the AC -- nhp-acd cannot start and the fail-closed backstop keeps
// tcp/443 shut -- and a skip is exactly how that reaches a deploy unnoticed.
// So the distinction is made here, from the process's own capability set: if
// the process could have done the thing, a failure is a failure.
func hasCaps(caps ...uint) bool {
	effective, ok := effectiveCaps()
	if !ok {
		return false
	}
	for _, c := range caps {
		if effective&(1<<c) == 0 {
			return false
		}
	}
	return true
}

// effectiveCaps reads the process's effective capability set.
func effectiveCaps() (uint64, bool) {
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0, false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "CapEff:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return 0, false
		}
		effective, err := strconv.ParseUint(fields[1], 16, 64)
		if err != nil {
			return 0, false
		}
		return effective, true
	}
	return 0, false
}

// requireCaps turns a failed privileged operation into a test failure when
// this process held the privilege to do it, and into a skip when it did not.
//
// what names the operation, so the message says which one failed; caps are the
// capabilities it needed.
func requireCaps(t *testing.T, err error, what string, caps ...uint) {
	t.Helper()
	if err == nil {
		return
	}
	if hasCaps(caps...) {
		t.Fatalf("%s failed and this process holds the capabilities it needs, so the kernel refused the program itself: %v", what, err)
	}
	t.Skipf("%s (needs %s): %v", what, capNames(caps), err)
}

// capNames spells the capability list out for a skip message.
func capNames(caps []uint) string {
	names := make([]string, 0, len(caps))
	for _, c := range caps {
		switch c {
		case capNetAdmin:
			names = append(names, "CAP_NET_ADMIN")
		case capBPF:
			names = append(names, "CAP_BPF")
		default:
			names = append(names, "cap"+strconv.FormatUint(uint64(c), 10))
		}
	}
	return strings.Join(names, " + ")
}
