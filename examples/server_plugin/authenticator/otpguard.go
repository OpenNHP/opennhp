package main

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// otpGuard limits guessing of the configured OTP code and, unless
// OTPAllowReuse is set, refuses a code that has already been used.
//
// A 6-digit TOTP checked with one step of skew accepts 3 codes at a time,
// so an unthrottled endpoint falls to brute force in about 3*10^5 guesses.
// Two limits apply, and neither depends on headers a client can forge:
//   - per source address (see otpClientIP; IPv6 per /64): at most otpMaxIPFailures
//     failures, after which the address is locked out for otpIPLockout;
//   - across all sources: at most otpGlobalMaxFailures failures per
//     otpGlobalWindow, so spreading guesses over many addresses does not
//     help. The cost is that a flood of failures blocks OTP logins for
//     everyone until the window ends; this is an authentication gate, so it
//     fails closed.
//
// The check and the count are one step: reserve counts the attempt as a
// failure before the code is checked, and a correct code gives it back.
// Otherwise a burst of parallel requests would all pass the check before
// any of their failures were recorded.
//
// Every table is bounded and its entries expire. The global limit also
// bounds the address table in practice: every tracked address holds at
// least one failure, and at most otpGlobalMaxFailures are admitted per
// window.
type otpGuard struct {
	mu  sync.Mutex
	now func() time.Time

	ips map[string]*otpIPState

	globalWindowStart time.Time
	globalFailures    int

	used map[string]time.Time // OTP code -> when it may be used again
}

type otpIPState struct {
	failures    int // includes attempts still being checked
	lastFailure time.Time
	lockedUntil time.Time
}

// otpTicket is one reserved attempt; pass it back to release or succeed.
type otpTicket struct {
	ip     string
	window time.Time // the global window the attempt was counted in
}

const (
	otpMaxIPFailures     = 5
	otpIPLockout         = 15 * time.Minute
	otpIPFailureTTL      = 15 * time.Minute // failures older than this are forgotten
	otpGlobalMaxFailures = 30
	otpGlobalWindow      = time.Minute
	otpMaxTrackedIPs     = 10000

	// otpValidationSkew is the number of time steps accepted on either
	// side of the current one (clock drift between server and phone).
	otpValidationSkew = 1
)

// otpCodeLifetime is how long a code can still validate after it was first
// accepted: the current step plus the skew on either side.
func otpCodeLifetime(periodSeconds int) time.Duration {
	return time.Duration(2*otpValidationSkew+1) * time.Duration(periodSeconds) * time.Second
}

func newOTPGuard() *otpGuard {
	return &otpGuard{
		now:  time.Now,
		ips:  make(map[string]*otpIPState),
		used: make(map[string]time.Time),
	}
}

var otpLimiter = newOTPGuard()

var (
	errOTPRateLimited = errors.New("too many failed OTP attempts")
	errOTPInvalid     = errors.New("invalid OTP code")
	errOTPUsed        = errors.New("OTP code already used")
)

// checkConfiguredOTP checks code against the configured OTPSecretKey under
// the rate limit for clientIP, and refuses a used code unless
// OTPAllowReuse is set. Every path that accepts the configured OTP goes
// through here, so none of them is weaker than another.
func checkConfiguredOTP(conf *config, clientIP, code string) error {
	ticket, ok := otpLimiter.reserve(clientIP)
	if !ok {
		return errOTPRateLimited
	}
	// A wrong code keeps the failure reserve already counted.
	if !GetQRAuthService().ValidateConfiguredOTP(conf.OTPSecretKey, code, conf.OTPPeriod) {
		return errOTPInvalid
	}
	if !otpLimiter.consume(code, conf.OTPAllowReuse, otpCodeLifetime(conf.OTPPeriod)) {
		// The code is correct, so this is not a guess: give the attempt back.
		otpLimiter.release(ticket)
		return errOTPUsed
	}
	otpLimiter.succeed(ticket)
	return nil
}

func (g *otpGuard) rollGlobalWindow(now time.Time) {
	if now.Sub(g.globalWindowStart) >= otpGlobalWindow {
		g.globalWindowStart, g.globalFailures = now, 0
	}
}

// reserve reports whether an OTP attempt from ip may be checked and, if so,
// counts it as a failure until release or succeed says otherwise.
func (g *otpGuard) reserve(ip string) (otpTicket, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()

	g.rollGlobalWindow(now)
	if g.globalFailures >= otpGlobalMaxFailures {
		return otpTicket{}, false
	}

	st, ok := g.ips[ip]
	switch {
	case ok && now.Before(st.lockedUntil):
		return otpTicket{}, false
	case ok && (!st.lockedUntil.IsZero() || now.Sub(st.lastFailure) > otpIPFailureTTL):
		// The lockout ended or the failures went stale: start over.
		*st = otpIPState{}
	case !ok:
		if len(g.ips) >= otpMaxTrackedIPs {
			g.evictIPs(now)
			if len(g.ips) >= otpMaxTrackedIPs {
				// Cannot track another address; refuse rather than check it
				// without a per-address limit.
				return otpTicket{}, false
			}
		}
		st = &otpIPState{}
		g.ips[ip] = st
	}

	g.globalFailures++
	st.failures++
	st.lastFailure = now
	if st.failures >= otpMaxIPFailures {
		st.lockedUntil = now.Add(otpIPLockout)
	}
	return otpTicket{ip: ip, window: g.globalWindowStart}, true
}

// release gives back an attempt that turned out not to be a guess (a
// correct code refused only because it was already used).
func (g *otpGuard) release(t otpTicket) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.refundGlobal(t)
	st, ok := g.ips[t.ip]
	if !ok {
		return
	}
	if st.failures > 0 {
		st.failures--
	}
	if st.failures < otpMaxIPFailures {
		st.lockedUntil = time.Time{}
	}
	if st.failures == 0 {
		delete(g.ips, t.ip)
	}
}

// succeed gives back the attempt and clears the address's failures after a
// correct code.
func (g *otpGuard) succeed(t otpTicket) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.refundGlobal(t)
	delete(g.ips, t.ip)
}

func (g *otpGuard) refundGlobal(t otpTicket) {
	// An attempt counted in an earlier window is not refunded from this one.
	if g.globalWindowStart.Equal(t.window) && g.globalFailures > 0 {
		g.globalFailures--
	}
}

// consume marks code as used for lifetime and reports whether it was
// unused. With reuse allowed it always succeeds: all users share one TOTP
// seed, so in a multi-user deployment several people legitimately enter
// the same code.
//
// Only codes that validated reach this table, and at most 2*skew+1 codes
// validate at any time, so pruning on every call keeps it a few entries.
func (g *otpGuard) consume(code string, allowReuse bool, lifetime time.Duration) bool {
	if allowReuse {
		return true
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()

	for c, until := range g.used {
		if !now.Before(until) {
			delete(g.used, c)
		}
	}
	if _, ok := g.used[code]; ok {
		return false
	}
	g.used[code] = now.Add(lifetime)
	return true
}

func (g *otpGuard) evictIPs(now time.Time) {
	for ip, st := range g.ips {
		if !now.Before(st.lockedUntil) && now.Sub(st.lastFailure) > otpIPFailureTTL {
			delete(g.ips, ip)
		}
	}
}

// parseTrustedProxies parses TrustedProxies entries (addresses or CIDR
// prefixes). One bad entry fails the whole list.
func parseTrustedProxies(entries []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(entries))
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if strings.Contains(e, "/") {
			p, err := netip.ParsePrefix(e)
			if err != nil {
				return nil, fmt.Errorf("invalid TrustedProxies entry %q: %v", e, err)
			}
			prefixes = append(prefixes, p.Masked())
			continue
		}
		a, err := netip.ParseAddr(e)
		if err != nil {
			return nil, fmt.Errorf("invalid TrustedProxies entry %q: %v", e, err)
		}
		a = a.Unmap()
		prefixes = append(prefixes, netip.PrefixFrom(a, a.BitLen()))
	}
	return prefixes, nil
}

func isTrustedProxy(s string, trusted []netip.Prefix) bool {
	a, err := netip.ParseAddr(strings.TrimSpace(s))
	if err != nil {
		return false
	}
	a = a.Unmap()
	for _, p := range trusted {
		if p.Contains(a) {
			return true
		}
	}
	return false
}

// otpLimitKey is the key an address is limited under: IPv4 addresses as
// they are, IPv6 addresses by their /64, because a single host commonly
// holds a whole /64 and could otherwise rotate through fresh keys.
func otpLimitKey(addr string) string {
	a, err := netip.ParseAddr(addr)
	if err != nil {
		return addr
	}
	a = a.Unmap()
	if a.Is4() {
		return a.String()
	}
	return netip.PrefixFrom(a, 64).Masked().String()
}

// otpClientIP is the key OTP attempts are limited by (see otpLimitKey) for
// the client's address. That address is the TCP peer
// unless that peer is one of the configured trusted proxies, in which case
// it is the right-most X-Forwarded-For entry that is not itself a trusted
// proxy: everything to the left of it was written by the client and can be
// forged. Gin's ClientIP is not used because nhp-server's engine trusts
// X-Forwarded-For from any peer.
func otpClientIP(ctx *gin.Context, trusted []netip.Prefix) string {
	peer := ctx.RemoteIP()
	if len(trusted) == 0 || !isTrustedProxy(peer, trusted) {
		return otpLimitKey(peer)
	}
	var hops []string
	for _, h := range ctx.Request.Header.Values("X-Forwarded-For") {
		hops = append(hops, strings.Split(h, ",")...)
	}
	for i := len(hops) - 1; i >= 0; i-- {
		hop := strings.TrimSpace(hops[i])
		if isTrustedProxy(hop, trusted) {
			continue
		}
		if _, err := netip.ParseAddr(hop); err == nil {
			return otpLimitKey(hop)
		}
		// Garbage where an address should be: fall back to the proxy, which
		// shares one limit across its clients rather than trusting the header.
		return otpLimitKey(peer)
	}
	return otpLimitKey(peer)
}
