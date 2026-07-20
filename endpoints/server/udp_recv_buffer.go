package server

import (
	"errors"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"

	"github.com/OpenNHP/opennhp/nhp/log"
)

const (
	DefaultUDPRecvBufferBytes = 8 * 1024 * 1024
	UDPRecvBufferEnvVar       = "NHP_UDP_RECV_BUFFER_BYTES"
)

var errUDPRecvBufferClamped = errors.New("kernel clamped SO_RCVBUF below target")

func parseUDPRecvBufferSize(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultUDPRecvBufferBytes, nil
	}
	size, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer %q: %w", UDPRecvBufferEnvVar, raw, err)
	}
	if size <= 0 {
		return 0, fmt.Errorf("%s: must be positive, got %d", UDPRecvBufferEnvVar, size)
	}
	return size, nil
}

// tuneUDPRecvBuffer raises the UDP listen socket's queue and verifies the
// effective value because kernels silently clamp SetReadBuffer to rmem_max.
// Tuning/readback failures are non-fatal; an invalid explicit override is
// rejected earlier by parseUDPRecvBufferSize.
func tuneUDPRecvBuffer(conn *net.UDPConn, target int) {
	if err := conn.SetReadBuffer(target); err != nil {
		log.Warning("[Server] SetReadBuffer(%d) failed: %v (continuing with kernel default)", target, err)
		return
	}
	effective, err := verifyUDPRecvBuffer(conn, target)
	switch {
	case err == nil:
		log.Info("[Server] SO_RCVBUF configured: requested %d bytes, effective %d bytes", target, effective)
	case errors.Is(err, errUDPRecvBufferClamped):
		log.Warning("[Server] SO_RCVBUF clamped: requested %d bytes, effective %d bytes; raise net.core.rmem_max", target, effective)
	default:
		log.Warning("[Server] SO_RCVBUF readback failed: %v (continuing unverified)", err)
	}
}

func verifyUDPRecvBuffer(conn *net.UDPConn, target int) (effective int, err error) {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return 0, fmt.Errorf("syscall conn: %w", err)
	}
	var socketErr error
	if err := rawConn.Control(func(fd uintptr) {
		effective, socketErr = getsockoptIntRcvBuf(fd)
	}); err != nil {
		return 0, fmt.Errorf("control fd: %w", err)
	}
	if socketErr != nil {
		return 0, fmt.Errorf("getsockopt SO_RCVBUF: %w", socketErr)
	}
	threshold := target
	if runtime.GOOS == "linux" {
		threshold *= 2 // Linux reports its internal doubled accounting value.
	}
	if effective < threshold {
		return effective, errUDPRecvBufferClamped
	}
	return effective, nil
}
