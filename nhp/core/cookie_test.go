package core

import (
	"bytes"
	"net"
	"testing"
)

// TestCookieRemoteKey_DirectUDPUsesRemoteAddr: when there is no relay
// in play, ConnData.RealRemoteAddr is nil and the cookie key must fall
// back to RemoteAddr. This is the single-server / direct-UDP path that
// existed before relay forwarding was added; it must keep working.
func TestCookieRemoteKey_DirectUDPUsesRemoteAddr(t *testing.T) {
	cd := &ConnectionData{RemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.7"), Port: 9000}}
	if got := cookieRemoteKey(cd); got != "203.0.113.7" {
		t.Fatalf("direct-UDP key: got %q, want %q", got, "203.0.113.7")
	}
}

// TestCookieRemoteKey_RelayForwardUsesRealRemoteAddr: this is the
// regression the relay-cookie fix is about. For relay-forwarded
// packets ConnData.RemoteAddr is the relay's UDP address and
// RealRemoteAddr is the actual client. Keying on RemoteAddr would
// make every client of one relay share a cookie — so the helper MUST
// prefer RealRemoteAddr when set.
func TestCookieRemoteKey_RelayForwardUsesRealRemoteAddr(t *testing.T) {
	cd := &ConnectionData{
		RemoteAddr:     &net.UDPAddr{IP: net.ParseIP("198.51.100.10"), Port: 62206}, // relay
		RealRemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.42"), Port: 51234},  // client
	}
	if got := cookieRemoteKey(cd); got != "203.0.113.42" {
		t.Fatalf("relay-forward key: got %q (relay address would have given %q), want %q",
			got, "198.51.100.10", "203.0.113.42")
	}
}

// TestCookieRemoteKey_DistinctClientsViaSameRelay: two clients
// arriving through the same relay must derive distinct cookie keys.
// Pre-fix they would have collided on the relay's IP — this test
// pins the property going forward.
func TestCookieRemoteKey_DistinctClientsViaSameRelay(t *testing.T) {
	relay := &net.UDPAddr{IP: net.ParseIP("198.51.100.10"), Port: 62206}
	a := &ConnectionData{
		RemoteAddr:     relay,
		RealRemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.42"), Port: 51234},
	}
	b := &ConnectionData{
		RemoteAddr:     relay,
		RealRemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.43"), Port: 51235},
	}
	if cookieRemoteKey(a) == cookieRemoteKey(b) {
		t.Fatalf("two clients via the same relay must produce distinct cookie keys; got %q for both",
			cookieRemoteKey(a))
	}
}

// TestCookieRemoteKey_IPv4MappedIPv6Normalized: a dual-stack listener
// can present the same agent as "::ffff:1.2.3.4" on one socket and
// "1.2.3.4" on another. The helper must normalize both forms so the
// cookie a sibling server minted under one form is still verifiable
// under the other.
func TestCookieRemoteKey_IPv4MappedIPv6Normalized(t *testing.T) {
	plain := &ConnectionData{RemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.7"), Port: 9000}}
	mapped := &ConnectionData{RemoteAddr: &net.UDPAddr{IP: net.ParseIP("::ffff:203.0.113.7"), Port: 9000}}
	if cookieRemoteKey(plain) != cookieRemoteKey(mapped) {
		t.Fatalf("IPv4 and IPv4-mapped-IPv6 must derive the same key; got %q vs %q",
			cookieRemoteKey(plain), cookieRemoteKey(mapped))
	}
}

// TestCookieRemoteKey_NilSafe: defensive callers — nil ConnData, nil
// addresses, nil IP — must return "" rather than panic. sendCookie /
// checkHMAC route the result into an HMAC input where "" is harmless
// (HMAC of zero-length is well-defined), so the right failure mode is
// "compare fails" not "process crashes".
func TestCookieRemoteKey_NilSafe(t *testing.T) {
	if got := cookieRemoteKey(nil); got != "" {
		t.Fatalf("nil ConnData: got %q, want empty", got)
	}
	if got := cookieRemoteKey(&ConnectionData{}); got != "" {
		t.Fatalf("empty ConnData: got %q, want empty", got)
	}
	if got := cookieRemoteKey(&ConnectionData{RemoteAddr: &net.UDPAddr{}}); got != "" {
		t.Fatalf("nil IP: got %q, want empty", got)
	}
}

// TestDeriveStatelessCookie_PeerPkBindsIdentity: this is the property
// the "stateless cookie has no per-handshake binding" fix is about.
// Two distinct agents arriving from the same source IP (same NAT/CGN
// egress) inside the same time window must derive distinct cookies, so
// one can't mint a cookie via its own legitimate KNK and have a
// sibling behind the same NAT replay it in an unrelated RKN.
func TestDeriveStatelessCookie_PeerPkBindsIdentity(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	remote := "203.0.113.7"
	pkA := bytes.Repeat([]byte{0xAA}, PublicKeySize)
	pkB := bytes.Repeat([]byte{0xBB}, PublicKeySize)
	const window int64 = 12345

	cookieA := deriveStatelessCookie(key, remote, pkA, window)
	cookieB := deriveStatelessCookie(key, remote, pkB, window)
	if bytes.Equal(cookieA, cookieB) {
		t.Fatalf("distinct peerPks behind the same NAT must derive distinct cookies; got identical %x", cookieA)
	}
}

// TestDeriveStatelessCookie_SiblingServerReDerivation: any sibling
// nhp-server with the same signing key, observing the same agent
// (same source IP + same agent static pubkey) inside the same time
// window, must produce the same cookie. This is the cluster routing
// invariant — if it ever breaks, the KNK → COK → RKN handshake stops
// surviving load-balancer shuffling between instances.
func TestDeriveStatelessCookie_SiblingServerReDerivation(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	remote := "203.0.113.7"
	pk := bytes.Repeat([]byte{0xAA}, PublicKeySize)
	const window int64 = 12345

	// Two independent derivations (e.g. mint on instance #1, verify
	// on instance #2) must agree byte-for-byte.
	a := deriveStatelessCookie(key, remote, pk, window)
	b := deriveStatelessCookie(key, remote, pk, window)
	if !bytes.Equal(a, b) {
		t.Fatalf("identical inputs must derive identical cookies; got %x vs %x", a, b)
	}
}

// TestDeriveStatelessCookie_WindowSeparation: cookies expire across
// window rollover. Verify changes to the window index alone produce a
// distinct cookie, so an attacker can't replay a 2-window-old cookie
// indefinitely.
func TestDeriveStatelessCookie_WindowSeparation(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	remote := "203.0.113.7"
	pk := bytes.Repeat([]byte{0xAA}, PublicKeySize)

	curr := deriveStatelessCookie(key, remote, pk, 12345)
	next := deriveStatelessCookie(key, remote, pk, 12346)
	if bytes.Equal(curr, next) {
		t.Fatalf("consecutive windows must derive distinct cookies; got identical %x", curr)
	}
}

// TestDeriveStatelessCookie_EmptyPeerPkSafe: pre-fix call sites (and
// pathological inputs) might pass a nil/empty peerPk. The HMAC
// primitive accepts zero-length writes, so the function must not
// panic. The resulting cookie won't match any well-formed RKN's HMAC
// — fail-closed is the right outcome.
func TestDeriveStatelessCookie_EmptyPeerPkSafe(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	remote := "203.0.113.7"

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("empty peerPk must not panic, got %v", r)
		}
	}()
	got := deriveStatelessCookie(key, remote, nil, 12345)
	if len(got) != CookieSize {
		t.Fatalf("expected %d-byte cookie even with nil peerPk, got %d", CookieSize, len(got))
	}
}
