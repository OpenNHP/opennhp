//go:build linux

package ebpf

import (
	"testing"

	"github.com/cilium/ebpf"
)

// restartIter yields a, b, then restarts (as an LRU eviction does) and yields
// a, b again before finishing.
type restartIter struct {
	seq []uint32
	pos int
}

func (r *restartIter) Next(k, v any) bool {
	if r.pos >= len(r.seq) {
		return false
	}
	*(k.(*uint32)) = r.seq[r.pos]
	*(v.(*[]uint64)) = []uint64{1, 2}
	r.pos++
	return true
}

func (r *restartIter) Err() error { return nil }

func TestDrainIterRestartDoesNotDoubleCount(t *testing.T) {
	out := map[uint32]uint64{}
	it := &restartIter{seq: []uint32{1, 2, 1, 2}}
	seen, err := drainIter[uint32](it, func(k uint32, h uint64) { out[k] += h })
	if err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 || out[1] != 3 || out[2] != 3 {
		t.Fatalf("seen=%d out=%v", len(seen), out)
	}
}

// TestDrainMapBatchPerCPU drives drainMapBatch against a real
// LRU_PERCPU_HASH, the map type the XDP program uses. It needs CAP_BPF, so it
// skips where the map cannot be created.
func TestDrainMapBatchPerCPU(t *testing.T) {
	m, err := ebpf.NewMap(&ebpf.MapSpec{
		Type:       ebpf.LRUCPUHash,
		KeySize:    4,
		ValueSize:  8,
		MaxEntries: 64,
	})
	if err != nil {
		t.Skipf("cannot create per-CPU map (needs CAP_BPF): %v", err)
	}
	defer m.Close()

	nCPU, err := ebpf.PossibleCPU()
	if err != nil {
		t.Fatal(err)
	}
	for k := uint32(1); k <= 3; k++ {
		vals := make([]uint64, nCPU)
		for i := range vals {
			vals[i] = uint64(k) // each key sums to k*nCPU
		}
		if err := m.Put(&k, vals); err != nil {
			t.Fatal(err)
		}
	}

	got := map[uint32]uint64{}
	n, ok := drainMapBatch(m, func(k uint32, hits uint64) { got[k] += hits })
	if !ok || n != 3 {
		t.Fatalf("drainMapBatch = (%d, %v), want (3, true)", n, ok)
	}
	for k := uint32(1); k <= 3; k++ {
		if want := uint64(k) * uint64(nCPU); got[k] != want {
			t.Errorf("key %d: hits = %d, want %d", k, got[k], want)
		}
	}

	// The batch must have emptied the map.
	n, _ = drainMapBatch(m, func(uint32, uint64) {})
	if n != 0 {
		t.Errorf("second drain returned %d entries, want 0", n)
	}
}
