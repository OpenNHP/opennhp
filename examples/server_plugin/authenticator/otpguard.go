package main

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
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
//   - across all sources: at most otpGlobalMaxFailures failures in any
//     otpGlobalWindow (a sliding window, not fixed buckets), so spreading
//     guesses over many addresses does not help. The cost is that a flood of
//     failures blocks OTP logins for everyone; this is an authentication
//     gate, so it fails closed. Holding it exhausted takes about 90 IPv4
//     addresses (or 30 IPv6 /48s) every 15 minutes, given the per-address
//     lockouts: cheap for a botnet, not possible from a single host. That
//     ceiling is also the only bound on a distributed attacker: 30 guesses a
//     minute against 3 valid codes is about a 12% chance of a correct guess
//     per day.
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

	global       []time.Time // when each failure counted in the last otpGlobalWindow was reserved, oldest first
	globalWarned time.Time   // when the global limit was last logged as reached

	used map[string]time.Time // OTP code -> when it may be used again
}

type otpIPState struct {
	failures    int // includes attempts still being checked
	lastFailure time.Time
	lockedUntil time.Time
}

// otpTicket is one reserved attempt; pass it back to release or succeed.
type otpTicket struct {
	keys []string  // the address's limit key, then any enclosing prefix keys
	at   time.Time // when the attempt was counted in global
}

const (
	otpMaxIPFailures = 5
	// otpMaxPrefixFailures caps a whole IPv6 /48 (see otpLimitKeys). One
	// /48, an ordinary customer allocation, holds 65,536 /64s, enough to keep
	// the global budget exhausted on per-/64 limits alone; with this cap it
	// gets 15 guesses per 15 minutes, half of one minute's global budget.
	otpMaxPrefixFailures = 15
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
	// errNoClientAddr: the peer address could not be parsed (for example a
	// unix-socket listener in front of nhp-server). Refused rather than
	// sharing one limit key and opening access for an empty address.
	errNoClientAddr = errors.New("cannot determine the client address")
)

// otpCheck is checkConfiguredOTP; a variable only so tests can make it
// return an error the handlers do not know about. Both otpvalid and verify
// call it.
var otpCheck = checkConfiguredOTP

// checkConfiguredOTP checks code against the configured OTPSecretKey under
// the rate limit for clientIP, and refuses a used code unless
// OTPAllowReuse is set. Every path that accepts the configured OTP goes
// through here, so none of them is weaker than another.
func checkConfiguredOTP(conf *config, clientIP, code string) error {
	if clientIP == "" {
		return errNoClientAddr
	}
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

func (g *otpGuard) pruneGlobal(now time.Time) {
	n := 0
	for n < len(g.global) && now.Sub(g.global[n]) >= otpGlobalWindow {
		n++
	}
	g.global = g.global[n:]
}

// otpLimitKeys returns the keys an attempt from limit key ip counts
// against: the key itself and, for an IPv6 /64, its enclosing /48.
func otpLimitKeys(ip string) []string {
	if p, err := netip.ParsePrefix(ip); err == nil && p.Addr().Is6() && p.Bits() == 64 {
		return []string{ip, netip.PrefixFrom(p.Addr(), 48).Masked().String()}
	}
	return []string{ip}
}

// otpKeyLimit is the failure cap for a key from otpLimitKeys.
func otpKeyLimit(key string) int {
	if p, err := netip.ParsePrefix(key); err == nil && p.Bits() == 48 {
		return otpMaxPrefixFailures
	}
	return otpMaxIPFailures
}

// reserve reports whether an OTP attempt from ip (an otpLimitKey) may be
// checked and, if so, counts it as a failure, against ip and every
// enclosing prefix, until release or succeed says otherwise.
func (g *otpGuard) reserve(ip string) (otpTicket, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()

	if ip == "" {
		return otpTicket{}, false
	}
	g.pruneGlobal(now)
	if len(g.global) >= otpGlobalMaxFailures {
		// Make an everyone-is-locked-out period visible, once per window.
		if now.Sub(g.globalWarned) >= otpGlobalWindow {
			g.globalWarned = now
			if log != nil {
				log.Warning("[OTP] global limit reached: %d failed attempts in the last %v; refusing every OTP attempt until they age out",
					len(g.global), otpGlobalWindow)
			}
		}
		return otpTicket{}, false
	}

	keys := otpLimitKeys(ip)
	// Check every key before changing any, so a refusal leaves no trace.
	untracked := 0
	for _, k := range keys {
		st, ok := g.ips[k]
		switch {
		case !ok:
			untracked++
		case now.Before(st.lockedUntil):
			return otpTicket{}, false
		}
	}
	if len(g.ips)+untracked > otpMaxTrackedIPs {
		g.evictIPs(now)
		if len(g.ips)+untracked > otpMaxTrackedIPs {
			// Cannot track another address; refuse rather than check it
			// without a per-address limit.
			return otpTicket{}, false
		}
	}

	for _, k := range keys {
		st, ok := g.ips[k]
		switch {
		case !ok:
			st = &otpIPState{}
			g.ips[k] = st
		case !st.lockedUntil.IsZero() || now.Sub(st.lastFailure) > otpIPFailureTTL:
			// The lockout ended or the failures went stale: start over.
			*st = otpIPState{}
		}
		st.failures++
		st.lastFailure = now
		if st.failures >= otpKeyLimit(k) {
			st.lockedUntil = now.Add(otpIPLockout)
		}
	}
	g.global = append(g.global, now)
	return otpTicket{keys: keys, at: now}, true
}

// release gives back an attempt that turned out not to be a guess (a
// correct code refused only because it was already used). If a counter was
// reset between reserve and release, this forgives at most one failure.
func (g *otpGuard) release(t otpTicket) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.refundGlobal(t)
	for _, k := range t.keys {
		g.refundKey(k)
	}
}

// succeed gives back the attempt after a correct code and clears the
// address's own failures. Enclosing prefixes only get this attempt back:
// other addresses in them may still be guessing.
//
// Clearing is deliberate, and it also drops failures that parallel requests
// from the same address reserved but have not resolved. That gains nothing
// today: it takes a valid code, and whoever has one need not guess. If the
// guard is ever reused with per-user secrets, subtract only this ticket's
// attempt instead.
func (g *otpGuard) succeed(t otpTicket) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.refundGlobal(t)
	for i, k := range t.keys {
		if i == 0 {
			delete(g.ips, k)
		} else {
			g.refundKey(k)
		}
	}
}

func (g *otpGuard) refundKey(k string) {
	st, ok := g.ips[k]
	if !ok {
		return
	}
	if st.failures > 0 {
		st.failures--
	}
	if st.failures < otpKeyLimit(k) {
		st.lockedUntil = time.Time{}
	}
	if st.failures == 0 {
		delete(g.ips, k)
	}
}

func (g *otpGuard) refundGlobal(t otpTicket) {
	// Remove this attempt's own entry; if it already aged out, nothing.
	// Entries reserved in the same clock tick are equal and interchangeable,
	// so removing another one of them leaves the count just as correct.
	for i := len(g.global) - 1; i >= 0; i-- {
		if g.global[i].Equal(t.at) {
			g.global = append(g.global[:i], g.global[i+1:]...)
			return
		}
	}
}

// consume marks code as used for lifetime and reports whether it was
// unused. With reuse allowed it always succeeds: all users share one TOTP
// seed, so in a multi-user deployment several people legitimately enter
// the same code.
//
// The table is keyed by the code alone, not by TOTP step: two steps inside
// one lifetime can produce the same six digits, and the second is then
// refused as used. That costs one login a wait for the next code with
// probability about 1e-6, which is not worth validating each step here.
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
// prefixes). One bad entry fails the whole list. Peers are compared as
// plain IPv4 (see isTrustedProxy), so IPv4-mapped entries are converted to
// IPv4 too; an IPv4-mapped prefix shorter than /96 covers more than the
// mapped range and is refused rather than silently never matching.
func parseTrustedProxies(entries []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(entries))
	for _, e := range entries {
		e = strings.TrimSpace(e)
		if strings.Contains(e, "/") {
			p, err := netip.ParsePrefix(e)
			if err != nil {
				return nil, fmt.Errorf("invalid TrustedProxies entry %q: %v", e, err)
			}
			if p.Addr().Is4In6() {
				if p.Bits() < 96 {
					return nil, fmt.Errorf("invalid TrustedProxies entry %q: an IPv4-mapped prefix must be /96 or longer", e)
				}
				p = netip.PrefixFrom(p.Addr().Unmap(), p.Bits()-96)
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
	return otpLimitKey(otpClientAddr(ctx, trusted))
}

// otpClientAddr is the client address otpClientIP limits by, unmasked, for
// recording in sessions and logs. It is "" when the peer address cannot be
// parsed; callers must refuse then (errNoClientAddr).
func otpClientAddr(ctx *gin.Context, trusted []netip.Prefix) string {
	// gin's RemoteIP is the host part of RemoteAddr, not checked to be an
	// IP; parse it so only a real address can become a limit key or the
	// knock's source, and so IPv4-mapped peers read as plain IPv4.
	a, err := netip.ParseAddr(ctx.RemoteIP())
	if err != nil {
		return ""
	}
	peer := a.Unmap().String()
	if len(trusted) == 0 {
		warnUntrustedForwarding(ctx, peer)
		return peer
	}
	if !isTrustedProxy(peer, trusted) {
		return peer
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
		if a, err := netip.ParseAddr(hop); err == nil {
			return a.Unmap().String()
		}
		// Garbage where an address should be: fall back to the proxy, which
		// shares one limit across its clients rather than trusting the header.
		return peer
	}
	return peer
}

// otpProxyWarned makes warnUntrustedForwarding log once per process.
var otpProxyWarned atomic.Bool

// warnUntrustedForwarding logs, once, when any peer sends X-Forwarded-For
// while TrustedProxies is empty. That usually means a reverse proxy, which
// may have a public address (a CDN, a cloud load balancer), so the peer's
// address is not used to decide. Limiting by the proxy's address means five
// wrong codes from anyone lock every user out. The header is still ignored.
// A directly connected client sending the header can also trigger it; it
// is logged once per process, so that costs one line.
func warnUntrustedForwarding(ctx *gin.Context, peer string) {
	if otpProxyWarned.Load() || ctx.GetHeader("X-Forwarded-For") == "" {
		return
	}
	if otpProxyWarned.CompareAndSwap(false, true) && log != nil {
		log.Warning("OTP rate limit: request from %s carries X-Forwarded-For but TrustedProxies is empty; "+
			"if nhp-server is behind a reverse proxy, every client shares the proxy's limit, one client can lock out all users, "+
			"and access is opened for the proxy's address instead of the client's. "+
			"Set TrustedProxies in the authenticator config.", peer)
	}
}

// requestTrustedProxies is the configured TrustedProxies, or none before a
// config has loaded.
func requestTrustedProxies() []netip.Prefix {
	if conf := baseConf; conf != nil {
		return conf.trustedProxies
	}
	return nil
}
