// Package loadbalance provides the shared load-balancing schemes and a
// generic picker used by both nhp-relay (selecting which nhp-server
// instance to forward an HTTPS-bridged knock to) and nhp-agent
// (selecting which nhp-server instance in a cluster to send a knock to).
//
// The scheme constants are part of the public configuration surface —
// they appear verbatim in operator-edited TOML files — so they live
// here rather than in any single endpoint's package.
package loadbalance

import (
	"fmt"
	"math/rand/v2"
	"sync/atomic"
)

// Scheme names a strategy for selecting an instance within a cluster.
// The string values are stable: they appear in operator-edited config
// files (relay.toml, agent's server.toml).
type Scheme string

const (
	SchemeRandom         Scheme = "random"
	SchemeWeightedRandom Scheme = "weighted-random"
	SchemeRoundRobin     Scheme = "round-robin"
)

// DefaultScheme is what an empty / unset scheme normalises to.
// Weighted-random matches the documented intuition "spread requests
// proportionally to declared instance weights" without surprising
// operators who left the field blank.
const DefaultScheme = SchemeWeightedRandom

// Validate rejects unknown scheme strings at config-load time. An empty
// string is accepted; callers should normalise it to DefaultScheme via
// Normalize before constructing a Picker.
func (s Scheme) Validate() error {
	switch s {
	case "", SchemeRandom, SchemeWeightedRandom, SchemeRoundRobin:
		return nil
	default:
		return fmt.Errorf("unknown load-balance scheme %q (valid: %q, %q, %q)",
			string(s), SchemeRandom, SchemeWeightedRandom, SchemeRoundRobin)
	}
}

// Normalize returns the scheme with the empty string replaced by
// DefaultScheme. Unknown schemes pass through unchanged — call
// Validate first to reject them.
func (s Scheme) Normalize() Scheme {
	if s == "" {
		return DefaultScheme
	}
	return s
}

// Weighted is the contract an instance type must satisfy to be Picked.
// Implementations should return a non-negative integer; zero weight is
// treated as 1 in NormalizeWeights so an instance with weight 0 in the
// config still receives traffic.
type Weighted interface {
	Weight() int
}

// NormalizeWeights returns the sum of weights with zero-weight instances
// counted as 1. Callers should precompute this once when building a
// Picker; the value stays constant for the picker's lifetime (instance
// churn requires a fresh Picker, just like a config reload).
func NormalizeWeights[T Weighted](instances []T) int {
	total := 0
	for _, inst := range instances {
		w := inst.Weight()
		if w <= 0 {
			w = 1
		}
		total += w
	}
	return total
}

// Picker selects one instance from a fixed slice according to a Scheme.
// Pick is safe for concurrent use; the round-robin cursor uses an
// atomic counter so handlers across goroutines don't contend on a
// mutex.
//
// The slice referenced by Picker MUST be treated as immutable after
// construction — instances are picked by index, so reordering or
// resizing while picks are in flight would race with the counter.
// Build a fresh Picker on config reload instead.
type Picker[T Weighted] struct {
	scheme      Scheme
	instances   []T
	totalWeight int
	rrCounter   atomic.Uint64
}

// NewPicker constructs a Picker for the given instances. The scheme is
// normalised; pass loadbalance.Validate() upstream if you need to
// reject typos before reaching here. An unknown scheme is silently
// downgraded to the default (callers that pre-validated will never hit
// this fallback).
func NewPicker[T Weighted](scheme Scheme, instances []T) *Picker[T] {
	p := &Picker[T]{
		scheme:      scheme.Normalize(),
		instances:   instances,
		totalWeight: NormalizeWeights(instances),
	}
	return p
}

// Pick returns one instance and true on success. When the picker has
// zero instances it returns the zero value and false — callers must
// branch on this rather than indexing the result, or empty clusters
// will silently appear to work.
func (p *Picker[T]) Pick() (T, bool) {
	var zero T
	if len(p.instances) == 0 {
		return zero, false
	}
	if len(p.instances) == 1 {
		return p.instances[0], true
	}

	switch p.scheme {
	case SchemeRoundRobin:
		// Add returns the post-increment value; subtract 1 so the
		// first call lands on index 0 (matches operator intuition
		// when reading logs / debugging which instance was first).
		i := (p.rrCounter.Add(1) - 1) % uint64(len(p.instances))
		return p.instances[i], true

	case SchemeRandom:
		return p.instances[rand.IntN(len(p.instances))], true

	case SchemeWeightedRandom:
		if p.totalWeight <= 0 {
			// Defensive: NormalizeWeights treats 0 as 1, so this
			// is only reachable if a caller constructed a Picker
			// with no instances (already handled above) or
			// supplied a custom Weight() that returned negatives.
			return p.instances[rand.IntN(len(p.instances))], true
		}
		r := rand.IntN(p.totalWeight)
		for _, inst := range p.instances {
			w := inst.Weight()
			if w <= 0 {
				w = 1
			}
			r -= w
			if r < 0 {
				return inst, true
			}
		}
		// Numerically unreachable; defensive return so every
		// path produces a value.
		return p.instances[len(p.instances)-1], true

	default:
		// Validate() should have rejected this at load time. If
		// somehow a Picker is constructed with an invalid scheme,
		// prefer "always pick instance 0" over panicking — the
		// service stays up, the failure mode is non-random
		// selection (visible in metrics) rather than 503s.
		return p.instances[0], true
	}
}

// Len reports the number of instances in this picker. Useful for
// metrics ("cluster has N instances") and tests.
func (p *Picker[T]) Len() int {
	return len(p.instances)
}

// Scheme returns the picker's scheme as normalised at construction.
func (p *Picker[T]) Scheme() Scheme {
	return p.scheme
}

// Instances returns the underlying slice. Callers must not mutate it
// — see the Picker doc-comment. Exposed only for read-only
// introspection (logs, /clusters endpoint, etc.).
func (p *Picker[T]) Instances() []T {
	return p.instances
}
