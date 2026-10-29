//go:build !linux && !darwin

package server

import "errors"

func getsockoptIntRcvBuf(uintptr) (int, error) {
	return 0, errors.New("SO_RCVBUF readback is unsupported on this platform")
}
