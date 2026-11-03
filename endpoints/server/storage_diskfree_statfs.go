//go:build linux || darwin || freebsd

package server

import "syscall"

// diskFreeBytes returns the space available to unprivileged users on the
// filesystem holding path. ok is false when it cannot be determined.
func diskFreeBytes(path string) (free int64, ok bool) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, false
	}
	//nolint:gosec // G115: block counts and sizes fit int64 on any real filesystem
	return int64(st.Bavail) * int64(st.Bsize), true
}
