package server

import (
	"net"
	"testing"
)

// TestValidateRelaySourceAddr_FlagToggles is the production-safety
// invariant: with allowPrivate=false, RFC1918 / loopback / CGNAT must
// be rejected so a compromised relay cannot inject fabricated entries
// into the server's connection map. With allowPrivate=true (set only
// in trusted local-only demos), the same private addresses must be
// accepted so that Docker Desktop's vpnkit gateway (and similar local
// NAT setups) can forward the docker-compose demo end-to-end.
//
// Port sanity is independent of the flag — out-of-range ports must
// always be rejected; there is no demo-friendly reason for them.
func TestValidateRelaySourceAddr_FlagToggles(t *testing.T) {
	const goodPort = 443

	tests := []struct {
		name         string
		ip           string
		port         int
		allowPrivate bool
		wantReject   bool
	}{
		// Production default: flag=false.
		{"public IP accepted in production", "203.0.113.5", goodPort, false, false},
		{"RFC1918 rejected in production", "192.168.65.1", goodPort, false, true},
		{"loopback rejected in production", "127.0.0.1", goodPort, false, true},
		{"CGNAT rejected in production", "100.64.0.1", goodPort, false, true},

		// Demo override: flag=true.
		{"Docker NAT gateway accepted under demo flag", "192.168.65.1", goodPort, true, false},
		{"loopback accepted under demo flag", "127.0.0.1", goodPort, true, false},
		{"public IP still accepted under demo flag", "203.0.113.5", goodPort, true, false},

		// Port sanity, both modes.
		{"zero port rejected even with flag off", "203.0.113.5", 0, false, true},
		{"negative port rejected even with flag on", "192.168.65.1", -1, true, true},
		{"port above range rejected with flag on", "192.168.65.1", 65536, true, true},
		{"port above range rejected with flag off", "203.0.113.5", 70000, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			reason := validateRelaySourceAddr(ip, tt.port, tt.allowPrivate)
			rejected := reason != ""
			if rejected != tt.wantReject {
				t.Fatalf("validateRelaySourceAddr(%s, %d, allowPrivate=%v) reason=%q reject=%v, want reject=%v",
					tt.ip, tt.port, tt.allowPrivate, reason, rejected, tt.wantReject)
			}
		})
	}
}

// TestValidateRelaySourceAddr_NilIP makes sure a SourceAddr.Ip that
// fails net.ParseIP (typo, empty string, etc.) is rejected regardless
// of the demo flag. The handler upstream is the only caller, so this
// is the boundary where we catch malformed input from the relay.
func TestValidateRelaySourceAddr_NilIP(t *testing.T) {
	for _, allowPrivate := range []bool{false, true} {
		if reason := validateRelaySourceAddr(nil, 443, allowPrivate); reason != "malformed" {
			t.Fatalf("nil IP should be rejected as malformed (allowPrivate=%v); got %q", allowPrivate, reason)
		}
	}
}

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

// TestRelayAddrFromConnKey ensures the mapKey parser used by the
// per-relay connection counter correctly recovers the relay's address
// from compound keys of the form "relay:<relayAddr>:<realClientAddr>".
// Both segments are themselves "host:port" so the parser must strip
// exactly the trailing two colon-separated tokens.
func TestRelayAddrFromConnKey(t *testing.T) {
	tests := []struct {
		mapKey string
		want   string
	}{
		// Direct UDP keys are not relay-forwarded.
		{"203.0.113.5:54321", ""},
		{"", ""},

		// IPv4 relay + IPv4 client.
		{"relay:198.51.100.1:62206:203.0.113.5:54321", "198.51.100.1:62206"},

		// Same relay but realClient port edge values.
		{"relay:198.51.100.1:62206:203.0.113.5:1", "198.51.100.1:62206"},
		{"relay:198.51.100.1:62206:203.0.113.5:65535", "198.51.100.1:62206"},

		// Malformed keys.
		{"relay:", ""},
		{"relay:foo", ""},
		{"relay:foo:bar", ""}, // only one colon after prefix → not enough segments
	}

	for _, tt := range tests {
		t.Run(tt.mapKey, func(t *testing.T) {
			got := relayAddrFromConnKey(tt.mapKey)
			if got != tt.want {
				t.Errorf("relayAddrFromConnKey(%q) = %q, want %q", tt.mapKey, got, tt.want)
			}
		})
	}
}
