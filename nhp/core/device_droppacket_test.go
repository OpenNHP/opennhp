package core

import (
	"sync/atomic"
	"testing"
)

// TestNotifyPacketDropped covers the observation hook that feeds the
// dropped-packet metric: it fires with the caller's stage string, tolerates
// no hook being set, and a panic inside it cannot escape onto the packet
// routine.
func TestNotifyPacketDropped(t *testing.T) {
	priv := make([]byte, 32)
	priv[0] = 1
	d := NewDevice(NHP_SERVER, priv, nil)
	if d == nil {
		t.Fatal("NewDevice returned nil")
	}

	// No hook configured: must be a no-op, not a nil deref.
	d.notifyPacketDropped("decrypt")

	var got atomic.Value
	d.SetOption(DeviceOptions{
		OnPacketDropped: func(stage string) { got.Store(stage) },
	})
	for _, stage := range []string{"parse", "validate", "decrypt", "queue_full"} {
		got.Store("")
		d.notifyPacketDropped(stage)
		if got.Load().(string) != stage {
			t.Fatalf("hook got %q, want %q", got.Load(), stage)
		}
	}

	// A panicking hook is recovered — notifyPacketDropped must return
	// normally so the packet routine keeps running.
	d.SetOption(DeviceOptions{
		OnPacketDropped: func(string) { panic("boom") },
	})
	d.notifyPacketDropped("decrypt") // must not panic
}
