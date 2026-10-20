//go:build linux

package ebpf

import "testing"

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
