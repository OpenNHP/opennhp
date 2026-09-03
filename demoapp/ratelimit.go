package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// trustedLoopbackProxies is the set of network origins the rate limiter
// honors X-Real-IP / X-Forwarded-For from. Loopback only: the deploy
// config binds 127.0.0.1:8081 behind a same-host nginx, and local-dev
// binds 0.0.0.0 (or :8081) with no proxy at all. Anything that can
// connect from outside loopback sets the headers itself, so trusting a
// non-loopback CIDR here would let any client spoof its IP.
var trustedLoopbackProxies = []string{"127.0.0.1", "::1"}

// ipLimiter is a per-IP fixed-window rate limiter. The demoapp runs as a
// single instance on the relay host, so an in-memory map is sufficient
// (no shared store needed). It exists to choke the public self-
// registration + login endpoints against DB/CPU sinks and the
// registration → requestOtp → nhp-server email chain (review #6).
type ipLimiter struct {
	mu     sync.Mutex
	window time.Duration
	max    int
	hits   map[string]*ipBucket
}

type ipBucket struct {
	count  int
	expiry time.Time
}

func newIPLimiter(max int, window time.Duration) *ipLimiter {
	return &ipLimiter{window: window, max: max, hits: make(map[string]*ipBucket)}
}

// Allow reports whether one more request from ip is permitted. The first
// request in a window starts the window; subsequent requests increment
// until max, after which the window rejects until it expires.
func (l *ipLimiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	// Opportunistic GC: drop expired entries whenever the map has grown
	// past a soft cap, BEFORE the new-bucket branch. A flood of distinct
	// IPs takes the new-bucket path exclusively, so putting the sweep
	// only on the increment branch left it unreachable for the very case
	// it exists to handle. Each sweep is bounded by the map size, and it
	// only runs past 4096 entries, so the steady-state cost is zero.
	if len(l.hits) > 4096 {
		for k, bb := range l.hits {
			if now.After(bb.expiry) {
				delete(l.hits, k)
			}
		}
	}
	b, ok := l.hits[ip]
	if !ok || now.After(b.expiry) {
		l.hits[ip] = &ipBucket{count: 1, expiry: now.Add(l.window)}
		return 1 <= l.max
	}
	b.count++
	return b.count <= l.max
}

// clientIP returns the request's originating IP. With TrustProxyHeaders
// enabled and a loopback-only trusted-proxy set (applyTrustedProxies),
// gin's c.ClientIP() honors X-Real-IP / X-Forwarded-For from the
// reverse proxy and falls back to the TCP remote address otherwise.
// Without TrustProxyHeaders, the trusted-proxy set is empty and
// c.ClientIP() returns the TCP remote address unconditionally — so a
// docker / local-dev deployment that publishes :8081 directly cannot be
// defeated by varying X-Real-IP (review follow-up to #5).
func clientIP(c *gin.Context) string {
	return c.ClientIP()
}

// applyTrustedProxies configures gin's proxy-header trust to match
// cfg.TrustProxyHeaders. Call this once after gin.New() and before any
// handler runs. When TrustProxyHeaders is false we pass nil to disable
// proxy-header parsing entirely; gin otherwise defaults to trusting
// every origin (and warns about it).
func applyTrustedProxies(engine *gin.Engine, trust bool) error {
	if trust {
		return engine.SetTrustedProxies(trustedLoopbackProxies)
	}
	return engine.SetTrustedProxies(nil)
}

// rateLimit is a gin middleware that applies an ipLimiter. Over the
// limit → 429 with a Retry-After equal to the window.
func rateLimit(l *ipLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.Allow(clientIP(c)) {
			c.Header("Retry-After", itoa(int(l.window/time.Second)))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many attempts from this address; try again later",
			})
			return
		}
		c.Next()
	}
}

// itoa is a stdlib-light int->string to avoid pulling strconv into the
// hot middleware path; the value is always a small positive seconds count.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Per-endpoint limiters. Public, unauthenticated, work-performing
// endpoints only. register is tighter because each call also kicks off
// keygen + a row insert and (once the SPA drives the flow) an nhp-server
// OTP email — the email-amplification primitive flagged in review #6.
var (
	registerLimiter = newIPLimiter(5, 10*time.Minute)  // 5 registrations / 10 min / IP
	loginLimiter    = newIPLimiter(10, 10*time.Minute) // 10 logins / 10 min / IP
)
