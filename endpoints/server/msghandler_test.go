package server

import (
	"net"
	"testing"
)

// TestIsRoutablePublicIP guards the SourceAddr filter that
// HandleRelayForward uses to reject fabricated entries from a
// misbehaving relay. The intent is "only accept addresses a real public
// client could plausibly originate from", so anything reserved /
// non-routable must be rejected.
//
// If a future change widens what counts as routable (e.g. relaxes the
// CGNAT bounds), these tests will catch the regression — silently
// accepting a reserved-range source IP would let a compromised relay
// install firewall rules under bogus IPs that audit tools would never
// expect to see.
func TestIsRoutablePublicIP(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		// Plausible public IPs — accepted.
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"203.0.113.5", true},          // TEST-NET-3 is technically reserved but
		{"2001:4860:4860::8888", true}, // global IPv6

		// Loopback.
		{"127.0.0.1", false},
		{"127.255.255.254", false},
		{"::1", false},

		// Unspecified.
		{"0.0.0.0", false},
		{"::", false},

		// RFC 1918 private.
		{"10.0.0.1", false},
		{"172.16.0.1", false},
		{"172.31.255.254", false},
		{"192.168.1.1", false},

		// RFC 4193 IPv6 unique-local.
		{"fc00::1", false},
		{"fd12:3456:789a::1", false},

		// Link-local.
		{"169.254.1.1", false},
		{"fe80::1", false},

		// Multicast.
		{"224.0.0.1", false},
		{"239.255.255.255", false},
		{"ff02::1", false},

		// CGNAT (RFC 6598) — explicit boundary checks.
		{"100.63.255.255", true}, // one below CGNAT range
		{"100.64.0.0", false},    // start of CGNAT
		{"100.64.0.1", false},
		{"100.127.255.255", false}, // end of CGNAT
		{"100.128.0.0", true},      // one above CGNAT range
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("net.ParseIP(%q) returned nil — fix the test data", tt.ip)
			}
			got := isRoutablePublicIP(ip)
			if got != tt.expected {
				t.Errorf("isRoutablePublicIP(%s) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}

	// nil input must not panic and must be rejected.
	t.Run("nil", func(t *testing.T) {
		if isRoutablePublicIP(nil) {
			t.Errorf("isRoutablePublicIP(nil) = true, want false")
		}
	})
}
