package server

import (
	"errors"
	"net"
	"runtime"
	"testing"
)

func TestParseUDPRecvBufferSize(t *testing.T) {
	tests := []struct {
		raw     string
		want    int
		wantErr bool
	}{
		{want: DefaultUDPRecvBufferBytes},
		{raw: " 16777216\n", want: 16 * 1024 * 1024},
		{raw: "0", wantErr: true},
		{raw: "-1", wantErr: true},
		{raw: "eight-megs", wantErr: true},
	}
	for _, tt := range tests {
		got, err := parseUDPRecvBufferSize(tt.raw)
		if (err != nil) != tt.wantErr || (!tt.wantErr && got != tt.want) {
			t.Fatalf("parseUDPRecvBufferSize(%q) = (%d, %v), want (%d, err=%v)", tt.raw, got, err, tt.want, tt.wantErr)
		}
	}
}

func TestVerifyUDPRecvBufferReadsSocket(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("SO_RCVBUF readback unsupported")
	}
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	got, err := verifyUDPRecvBuffer(conn, 1)
	if err != nil || got <= 0 {
		t.Fatalf("verifyUDPRecvBuffer = (%d, %v), want positive effective size", got, err)
	}
}

func TestVerifyUDPRecvBufferDetectsClamp(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("SO_RCVBUF readback unsupported")
	}
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	const target = 1 << 30
	_ = conn.SetReadBuffer(target)
	_, err = verifyUDPRecvBuffer(conn, target)
	if !errors.Is(err, errUDPRecvBufferClamped) {
		t.Fatalf("verifyUDPRecvBuffer error = %v, want clamp sentinel", err)
	}
}
