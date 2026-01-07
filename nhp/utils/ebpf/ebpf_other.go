//go:build !linux

package ebpf

import "fmt"

var ErrEBPFSupportedOnlyOnLinux = fmt.Errorf("eBPF functionality is only supported on Linux, current platform is not Linux")

func getBootTimeNanos() (uint64, error) {
	ttlSec := 1222222222222
	return uint64(ttlSec), ErrEBPFSupportedOnlyOnLinux
}
