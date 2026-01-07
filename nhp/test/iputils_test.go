package test

import (
	"testing"

	"github.com/OpenNHP/opennhp/nhp/utils"
)

func TestDetectIPType(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected utils.IPTYPE
		hasError bool
	}{
		// Valid IPv4 addresses
		{"standard IPv4", "192.168.1.1", utils.IPV4, false},
		{"IPv4 loopback", "127.0.0.1", utils.IPV4, false},
		{"IPv4 broadcast", "255.255.255.255", utils.IPV4, false},
		{"IPv4 zero", "0.0.0.0", utils.IPV4, false},
		{"IPv4 private class A", "10.0.0.1", utils.IPV4, false},
		{"IPv4 private class B", "172.16.0.1", utils.IPV4, false},
		{"IPv4 private class C", "192.168.0.1", utils.IPV4, false},

		// Valid IPv6 addresses
		{"IPv6 loopback", "::1", utils.IPV6, false},
		{"IPv6 full", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", utils.IPV6, false},
		{"IPv6 compressed", "2001:db8:85a3::8a2e:370:7334", utils.IPV6, false},
		{"IPv6 link-local", "fe80::1", utils.IPV6, false},
		{"IPv6 all zeros", "::", utils.IPV6, false},
		{"IPv6 documentation", "2001:db8::1", utils.IPV6, false},

		// IPv4-mapped IPv6 addresses are treated as IPv4 by Go's net package
		// because they represent IPv4 addresses embedded in IPv6 notation.
		// This is the correct behavior for firewall rules.
		{"IPv4-mapped IPv6", "::ffff:192.168.1.1", utils.IPV4, false},
		{"IPv4-mapped IPv6 zeros", "::ffff:0.0.0.0", utils.IPV4, false},

		// Invalid addresses
		{"empty string", "", 0, true},
		{"random text", "not-an-ip", 0, true},
		{"incomplete IPv4", "192.168.1", 0, true},
		{"IPv4 with port", "192.168.1.1:8080", 0, true},
		{"IPv6 with brackets", "[::1]", 0, true},
		{"overflow IPv4", "256.256.256.256", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.DetectIPType(tt.ip)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error for %s, got nil", tt.ip)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for %s: %v", tt.ip, err)
				}
				if result != tt.expected {
					t.Errorf("for %s: expected %d, got %d", tt.ip, tt.expected, result)
				}
			}
		})
	}
}

func TestIsIPv4(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"valid IPv4", "192.168.1.1", true},
		{"IPv4 loopback", "127.0.0.1", true},
		{"IPv6 loopback", "::1", false},
		{"IPv6 full", "2001:db8::1", false},
		{"IPv4-mapped IPv6", "::ffff:192.168.1.1", true}, // Treated as IPv4
		{"invalid", "not-an-ip", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsIPv4(tt.ip)
			if result != tt.expected {
				t.Errorf("IsIPv4(%s): expected %v, got %v", tt.ip, tt.expected, result)
			}
		})
	}
}

func TestIsIPv6(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"IPv6 loopback", "::1", true},
		{"IPv6 full", "2001:db8::1", true},
		{"IPv6 link-local", "fe80::1", true},
		{"IPv4-mapped IPv6", "::ffff:192.168.1.1", false}, // Treated as IPv4
		{"valid IPv4", "192.168.1.1", false},
		{"IPv4 loopback", "127.0.0.1", false},
		{"invalid", "not-an-ip", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsIPv6(tt.ip)
			if result != tt.expected {
				t.Errorf("IsIPv6(%s): expected %v, got %v", tt.ip, tt.expected, result)
			}
		})
	}
}

func TestGetCIDRMask(t *testing.T) {
	tests := []struct {
		name      string
		ipType    utils.IPTYPE
		rangeMode bool
		expected  string
	}{
		{"IPv4 single host", utils.IPV4, false, "/32"},
		{"IPv4 range", utils.IPV4, true, "/25"},
		{"IPv6 single host", utils.IPV6, false, "/128"},
		{"IPv6 range", utils.IPV6, true, "/121"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.GetCIDRMask(tt.ipType, tt.rangeMode)
			if result != tt.expected {
				t.Errorf("GetCIDRMask(%d, %v): expected %s, got %s", tt.ipType, tt.rangeMode, tt.expected, result)
			}
		})
	}
}

func TestCIDRMaskConstants(t *testing.T) {
	if utils.IPv4SingleHost != "/32" {
		t.Errorf("IPv4SingleHost: expected /32, got %s", utils.IPv4SingleHost)
	}
	if utils.IPv4AdjacentRange != "/25" {
		t.Errorf("IPv4AdjacentRange: expected /25, got %s", utils.IPv4AdjacentRange)
	}
	if utils.IPv6SingleHost != "/128" {
		t.Errorf("IPv6SingleHost: expected /128, got %s", utils.IPv6SingleHost)
	}
	if utils.IPv6AdjacentRange != "/121" {
		t.Errorf("IPv6AdjacentRange: expected /121, got %s", utils.IPv6AdjacentRange)
	}
}
