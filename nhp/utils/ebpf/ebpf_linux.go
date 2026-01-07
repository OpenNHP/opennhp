//go:build linux

package ebpf

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func getBootTimeNanos() (uint64, error) {
	var ts unix.Timespec
	if err := unix.ClockGettime(unix.CLOCK_BOOTTIME, &ts); err != nil {
		return 0, fmt.Errorf("clock_gettime failed: %v", err)
	}
	return uint64(ts.Sec)*1e9 + uint64(ts.Nsec), nil
}
