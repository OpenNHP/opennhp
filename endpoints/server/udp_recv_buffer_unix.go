//go:build linux || darwin

package server

import "syscall"

func getsockoptIntRcvBuf(fd uintptr) (int, error) {
	return syscall.GetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_RCVBUF)
}
