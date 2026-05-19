package core

import (
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
