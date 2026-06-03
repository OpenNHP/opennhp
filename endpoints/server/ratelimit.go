package server

import (
	"net"
	"sync"
)

// rknRateLimiter is a per-source-IP token-bucket limiter for RKN packets
// received while the server is in overload.
//
// WHY THIS EXISTS: under overload, an NHP_RKN's HMAC is verified against a
// stateless cookie (responder.checkHMAC(sumCookie=true)). That path must
// first recover the initiator's static public key via a full Noise IK
// floor — one ECDH + one AEAD-Open + two hash chains — BEFORE the cheap
// cookie-HMAC comparison can reject the packet. An unauthenticated
// attacker who knows only the server's (public) static key can therefore
// mint syntactically-valid RKN frames and force that per-packet ECDH at
// will. The cookie was supposed to be the cheap rejector; without a gate
// in front of it, it becomes the most expensive step under attack. This
// limiter is that gate: it runs in recvPacketRoutine, before the packet
// is ever queued to a connection routine, so flood traffic is dropped
// ahead of the ECDH.
//
// SCOPE: deliberately narrow — applied only on the (overload && RKN) path
// (see recvPacketRoutine). Normal KNK/handshake traffic and the
// non-overloaded steady state are untouched.
//
// DISPOSITION: over-limit packets are DROPPED only. They are not counted
// as threats and never trigger AddBlockAddr, so a legitimate agent whose
// retry cadence is far below the bucket rate is never penalised, and many
// agents behind one NAT/CGN egress IP degrade to dropped-RKN (which the
// agent already retries) rather than a hard block on the shared address.
//
// CONCURRENCY: safe for concurrent use via an internal mutex. Two callers
// share one limiter: the single recvPacketRoutine goroutine (direct-UDP
// RKN, keyed on the source IP) and the concurrent HandleRelayForward
// handler goroutines (relay-forwarded inner RKN, keyed on the real client
// IP). The mutex is only contended on the overload+RKN path, so it adds
// no cost to the steady state.
//
// MEMORY: the entry map is capacity-bounded. An attacker rotating spoofed
// source IPs cannot grow it without bound — once it exceeds maxEntries a
// sweep drops idle buckets, and if the table is still full the oldest
// bucket is evicted. So the limiter cannot itself become a memory-DoS
// vector.
type rknRateLimiter struct {
	// nanosPerToken is the steady-state cost of one allowed packet: a
	// bucket accrues one token every nanosPerToken nanoseconds.
	nanosPerToken int64
	// burstNanos is the bucket ceiling expressed as accrued time budget
	// (burst tokens * nanosPerToken). Capping the budget — rather than a
	// token count — keeps the arithmetic integer-only and float-free.
	burstNanos int64
	// idleNanos: a bucket untouched for this long is reclaimable by a
	// sweep. Set well above one full window so an agent that legitimately
	// goes quiet between knock cycles keeps its bucket.
	idleNanos  int64
	maxEntries int

	mu      sync.Mutex
	buckets map[string]*rknBucket
}

type rknBucket struct {
	// allowanceNanos is the accrued time budget; one token == nanosPerToken
	// of budget. Clamped to [0, burstNanos].
	allowanceNanos int64
	lastSeenNanos  int64
}

func newRknRateLimiter(ratePerSec, burst, maxEntries int, idleNanos int64) *rknRateLimiter {
	if ratePerSec <= 0 {
		ratePerSec = 1
	}
	if burst <= 0 {
		burst = 1
	}
	npt := int64(1_000_000_000) / int64(ratePerSec)
	if npt <= 0 {
		npt = 1
	}
	return &rknRateLimiter{
		nanosPerToken: npt,
		burstNanos:    npt * int64(burst),
		idleNanos:     idleNanos,
		maxEntries:    maxEntries,
		buckets:       make(map[string]*rknBucket),
	}
}

// keyForAddr returns the limiter key: the source IP only, never the port.
// An attacker rotating source ports must not get a fresh bucket per port.
func keyForAddr(addr *net.UDPAddr) string {
	if addr == nil || addr.IP == nil {
		return ""
	}
	return addr.IP.String()
}

// allow reports whether an RKN from addr may proceed to cookie
// verification at time nowNanos (the caller's already-computed recvTime).
// It consumes one token on success; on failure the packet should be
// dropped. nowNanos is passed in (not read from the clock here) so the
// limiter is deterministic under test.
func (r *rknRateLimiter) allow(addr *net.UDPAddr, nowNanos int64) bool {
	key := keyForAddr(addr)
	if key == "" {
		// No usable source IP — fail closed (drop). A packet with no
		// resolvable source can't be rate-accounted, and admitting it
		// would let such packets bypass the gate entirely.
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.buckets[key]
	if !ok {
		if len(r.buckets) >= r.maxEntries {
			r.evict(nowNanos)
		}
		// Fresh bucket starts full so a legitimate agent's first RKN in
		// an overload window is never dropped.
		b = &rknBucket{allowanceNanos: r.burstNanos, lastSeenNanos: nowNanos}
		r.buckets[key] = b
	} else {
		// Refill: credit the elapsed time, capped at the burst ceiling.
		elapsed := nowNanos - b.lastSeenNanos
		if elapsed > 0 {
			b.allowanceNanos += elapsed
			if b.allowanceNanos > r.burstNanos {
				b.allowanceNanos = r.burstNanos
			}
		}
		b.lastSeenNanos = nowNanos
	}

	if b.allowanceNanos >= r.nanosPerToken {
		b.allowanceNanos -= r.nanosPerToken
		return true
	}
	return false
}

// evictSampleSize is the number of buckets inspected per eviction. Go
// randomizes map iteration order, so the first N keys are an unbiased
// random sample — large enough that the bucket we drop is very likely
// among the older/idle ones, small enough to stay O(1) per packet.
const evictSampleSize = 8

// evict makes room for one new bucket when the map is at capacity. It
// inspects at most evictSampleSize buckets (a random sample, courtesy of
// Go's randomized map iteration) and deletes one: the first idle bucket it
// sees, or — if none in the sample is idle — the oldest in the sample.
//
// This is deliberately O(evictSampleSize), NOT O(len(buckets)). The
// earlier full-table sweep + full-table oldest-scan ran on every packet
// once the map saturated, so an attacker spraying RKN from >maxEntries
// rotating spoofed source IPs forced an O(maxEntries) scan per packet —
// under r.mu, which also serializes recvPacketRoutine and the relay
// handlers — re-introducing the very CPU/contention cost the limiter
// exists to remove. Random-sample eviction keeps the hard memory bound
// (one deletion guarantees room for one insertion) while making the
// per-packet cost a small constant. Called inline from allow() with r.mu
// already held.
func (r *rknRateLimiter) evict(nowNanos int64) {
	var oldestKey string
	var oldest int64
	seen := 0
	for k, b := range r.buckets {
		if nowNanos-b.lastSeenNanos > r.idleNanos {
			// Idle bucket in the sample: drop it immediately, it's the
			// best possible eviction (reclaimable anyway).
			delete(r.buckets, k)
			return
		}
		if seen == 0 || b.lastSeenNanos < oldest {
			oldest = b.lastSeenNanos
			oldestKey = k
		}
		seen++
		if seen >= evictSampleSize {
			break
		}
	}
	if oldestKey != "" {
		delete(r.buckets, oldestKey)
	}
}
