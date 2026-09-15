package core

import (
	"bytes"
	"compress/zlib"
	"strings"
	"testing"
)

func compressBodyForTest(t *testing.T, body []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(body); err != nil {
		t.Fatalf("zlib write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zlib close: %v", err)
	}
	return buf.Bytes()
}

func TestInflateBodyBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		wantError bool
	}{
		{name: "small", size: 1024},
		{name: "at limit", size: MaxDecompressedBodySize},
		{name: "one byte over", size: MaxDecompressedBodySize + 1, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.Repeat([]byte{'A'}, tt.size)
			got, err := inflateBody(compressBodyForTest(t, body))
			if tt.wantError {
				if err == nil || !strings.Contains(err.Error(), "exceeds limit") {
					t.Fatalf("inflateBody() error = %v, want size-limit error", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("inflateBody(): %v", err)
			}
			if !bytes.Equal(got, body) {
				t.Fatal("inflateBody() changed the payload")
			}
		})
	}
}

func TestInflateBodyRejectsInvalidZlib(t *testing.T) {
	if _, err := inflateBody([]byte("not zlib")); err == nil {
		t.Fatal("inflateBody() error = nil, want malformed-zlib error")
	}
}

func TestDecompressWarnAllowedThrottle(t *testing.T) {
	lastDecompressWarnNano.Store(0)
	t.Cleanup(func() { lastDecompressWarnNano.Store(0) })

	now := 2 * decompressWarnInterval
	if !decompressWarnAllowed(now) {
		t.Fatal("first warning should be allowed")
	}
	if decompressWarnAllowed(now + decompressWarnInterval - 1) {
		t.Fatal("warning inside throttle interval should be suppressed")
	}
	if !decompressWarnAllowed(now + decompressWarnInterval) {
		t.Fatal("warning at throttle boundary should be allowed")
	}
}
