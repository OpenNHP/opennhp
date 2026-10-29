//go:build linux

package ebpf

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// getBootTimeNanos returns the clock reading that the eBPF programs compare
// expiry timestamps against.
//
// It must be CLOCK_MONOTONIC, not CLOCK_BOOTTIME: the eBPF side stamps packets
// with bpf_ktime_get_ns(), which is CLOCK_MONOTONIC (bpf_ktime_get_boot_ns() is
// the BOOTTIME one). The two clocks diverge by however long the host has spent
// suspended, and reading BOOTTIME here made the computed expire_time land too
// far in the future — i.e. the door stayed open past OpenTime.
func getBootTimeNanos() (uint64, error) {
	var ts unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_MONOTONIC, &ts); err != nil {
		return 0, fmt.Errorf("clock_gettime failed: %v", err)
	}
	return uint64(ts.Sec)*1e9 + uint64(ts.Nsec), nil
}
