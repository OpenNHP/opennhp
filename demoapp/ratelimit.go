package main

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

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
	b, ok := l.hits[ip]
	if !ok || now.After(b.expiry) {
		l.hits[ip] = &ipBucket{count: 1, expiry: now.Add(l.window)}
		return 1 <= l.max
	}
	b.count++
	// Opportunistic GC: if the map has grown past a soft cap, drop
	// expired entries so a flood of distinct IPs does not bloat memory.
	if len(l.hits) > 4096 {
		for k, bb := range l.hits {
			if now.After(bb.expiry) {
				delete(l.hits, k)
			}
		}
	}
	return b.count <= l.max
}

// clientIP returns the request's originating IP. nginx (demoapp.conf
// .template) sets X-Real-IP to $remote_addr with replace, not append,
// so the header is non-spoofable in prod — and the prod config binds
// ListenAddr = 127.0.0.1:8081, so only the local nginx can reach the
// app at all (review #5). The loopback bind is the primary defense;
// this header trust is secondary. We fall back to the TCP remote
// address when the header is absent (local dev without nginx).
func clientIP(c *gin.Context) string {
	if v := c.GetHeader("X-Real-IP"); v != "" {
		return v
	}
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return host
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
