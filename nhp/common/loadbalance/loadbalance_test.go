package loadbalance

import (
	"strings"
	"testing"
)

// fakeInstance is a minimal Weighted to drive the picker tests without
// pulling in any real endpoint type.
type fakeInstance struct {
	id     string
	weight int
}

func (f *fakeInstance) Weight() int { return f.weight }

func newInsts(specs ...string) []*fakeInstance {
	out := make([]*fakeInstance, len(specs))
	for i, s := range specs {
		out[i] = &fakeInstance{id: s, weight: 1}
	}
	return out
}

// TestScheme_Validate: only the three documented schemes plus empty
// string are accepted. Anything else (typos, legacy values) must error
// so operators learn at load time, not at runtime.
func TestScheme_Validate(t *testing.T) {
	ok := []Scheme{"", SchemeRandom, SchemeWeightedRandom, SchemeRoundRobin}
	for _, s := range ok {
		if err := s.Validate(); err != nil {
			t.Errorf("Validate(%q) returned error: %v", s, err)
		}
	}
	bad := []Scheme{"Random", "weighted_random", "rr", "lb-round-robin", "default"}
	for _, s := range bad {
		err := s.Validate()
		if err == nil {
			t.Errorf("Validate(%q) accepted typo", s)
			continue
		}
		if !strings.Contains(err.Error(), string(s)) {
			t.Errorf("Validate(%q) error must mention the offending value, got: %v", s, err)
		}
	}
}

// TestScheme_NormalizeEmpty: empty scheme becomes the documented default
// (weighted-random) so callers can pass an unconfigured value straight
// through without a separate "is it empty?" check.
func TestScheme_NormalizeEmpty(t *testing.T) {
	if Scheme("").Normalize() != DefaultScheme {
		t.Fatalf("empty scheme must normalise to %q, got %q",
			DefaultScheme, Scheme("").Normalize())
	}
	// Known schemes pass through unchanged.
	for _, s := range []Scheme{SchemeRandom, SchemeWeightedRandom, SchemeRoundRobin} {
		if s.Normalize() != s {
			t.Errorf("known scheme %q must round-trip through Normalize", s)
		}
	}
}

// TestPicker_Empty: a picker with zero instances must report (zero,
// false) rather than panicking. Callers branch on the bool to emit
// the right error to the user.
func TestPicker_Empty(t *testing.T) {
	p := NewPicker[*fakeInstance](SchemeRoundRobin, nil)
	got, ok := p.Pick()
	if ok {
		t.Fatalf("empty picker must return ok=false, got inst=%v", got)
	}
	if got != nil {
		t.Fatalf("empty picker must return zero value, got %v", got)
	}
}

// TestPicker_Single: with one instance every scheme always returns
// that instance — no randomness, no math, no division by zero. This
// covers the trivial-cluster case that we expect to be the most common
// production deployment for single-instance nhp-server setups.
func TestPicker_Single(t *testing.T) {
	for _, s := range []Scheme{SchemeRandom, SchemeWeightedRandom, SchemeRoundRobin, "", "garbage"} {
		p := NewPicker(s, newInsts("only"))
		for i := 0; i < 10; i++ {
			got, ok := p.Pick()
			if !ok || got.id != "only" {
				t.Fatalf("scheme=%q must always return sole instance, got (%v, %v)", s, got, ok)
			}
		}
	}
}

// TestPicker_RoundRobin: round-robin must walk indices in order and
// wrap cleanly. Asserts an exact sequence (deterministic), so any
// off-by-one in the cursor math gets caught immediately.
func TestPicker_RoundRobin(t *testing.T) {
	p := NewPicker(SchemeRoundRobin, newInsts("a", "b", "c"))
	want := []string{"a", "b", "c", "a", "b", "c"}
	for i, w := range want {
		got, ok := p.Pick()
		if !ok || got.id != w {
			t.Fatalf("step %d: got (%v, %v), want %q", i, got, ok, w)
		}
	}
}

// TestPicker_RandomReachesAll: random must eventually hit every
// instance. 1000 picks across 3 instances makes a false negative
// astronomically unlikely (probability of missing one is <10^-176),
// so this test catches "scheme ignored, always picks index 0" bugs.
func TestPicker_RandomReachesAll(t *testing.T) {
	p := NewPicker(SchemeRandom, newInsts("a", "b", "c"))
	seen := map[string]int{}
	for i := 0; i < 1000; i++ {
		got, _ := p.Pick()
		seen[got.id]++
	}
	for _, want := range []string{"a", "b", "c"} {
		if seen[want] == 0 {
			t.Fatalf("random never picked %q (seen=%v)", want, seen)
		}
	}
}

// TestPicker_WeightedRandomBias: weight=10 instance must dominate
// against weight=1 peers. The bound (>=60%) is loose so any sensible
// RNG passes, but a fully-broken impl (ignores weights, off-by-one in
// the cumulative sum) is caught.
func TestPicker_WeightedRandomBias(t *testing.T) {
	insts := []*fakeInstance{
		{id: "low1", weight: 1},
		{id: "low2", weight: 1},
		{id: "high", weight: 10},
	}
	p := NewPicker(SchemeWeightedRandom, insts)
	const samples = 5000
	seen := map[string]int{}
	for i := 0; i < samples; i++ {
		got, _ := p.Pick()
		seen[got.id]++
	}
	share := float64(seen["high"]) / float64(samples)
	if share < 0.60 {
		t.Fatalf("weighted-random did not bias toward weight=10 instance: share=%.3f want >=0.60 (seen=%v)",
			share, seen)
	}
	// Low-weight instances must still receive some traffic; expected
	// ~417 each in 5000 samples, so zero is a clear starvation bug.
	for _, low := range []string{"low1", "low2"} {
		if seen[low] == 0 {
			t.Fatalf("weighted-random starved %q: %v", low, seen)
		}
	}
}

// TestPicker_ZeroWeightTreatedAsOne: an instance configured with
// weight=0 must still receive picks (NormalizeWeights treats it as 1).
// Otherwise a typo in operator config silently sidelines an instance.
func TestPicker_ZeroWeightTreatedAsOne(t *testing.T) {
	insts := []*fakeInstance{
		{id: "zero", weight: 0},
		{id: "one", weight: 1},
	}
	if got := NormalizeWeights(insts); got != 2 {
		t.Fatalf("NormalizeWeights with [0,1] must be 2, got %d", got)
	}
	p := NewPicker(SchemeWeightedRandom, insts)
	seen := map[string]int{}
	for i := 0; i < 2000; i++ {
		got, _ := p.Pick()
		seen[got.id]++
	}
	if seen["zero"] == 0 {
		t.Fatalf("weight=0 instance was starved: %v", seen)
	}
}

// TestPicker_UnknownSchemeFallsBackDeterministically: a Picker
// constructed with a not-yet-rejected unknown scheme (i.e. someone
// skipped Validate before NewPicker) must not panic and must not
// silently pick at random — pick the first instance so the bug is
// visible in logs / metrics rather than masquerading as load
// distribution.
func TestPicker_UnknownSchemeFallsBackDeterministically(t *testing.T) {
	p := NewPicker(Scheme("nonsense"), newInsts("a", "b", "c"))
	for i := 0; i < 5; i++ {
		got, ok := p.Pick()
		if !ok || got.id != "a" {
			t.Fatalf("unknown scheme must deterministically pick index 0, got (%v, %v)", got, ok)
		}
	}
}
