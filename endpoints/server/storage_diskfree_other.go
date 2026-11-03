//go:build !(linux || darwin || freebsd)

package server

// diskFreeBytes cannot determine free space on this platform (no Statfs
// with these fields), so the free-space check is skipped; the upload size
// limit still applies.
func diskFreeBytes(string) (int64, bool) { return 0, false }
