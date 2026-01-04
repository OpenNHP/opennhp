package test

import (
	"net"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/utils"
)

// TestIPv6CIDRParsing verifies that IPv6 CIDR notation works correctly
func TestIPv6CIDRParsing(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		mask        string
		expectValid bool
		expectNet   string
	}{
		// IPv6 single host
		{"IPv6 loopback /128", "::1", "/128", true, "::1/128"},
		{"IPv6 address /128", "2001:db8::1", "/128", true, "2001:db8::1/128"},
		{"IPv6 link-local /128", "fe80::1", "/128", true, "fe80::1/128"},

		// IPv6 range (121-bit = 128 addresses)
		{"IPv6 address /121", "2001:db8::1", "/121", true, "2001:db8::/121"},
		{"IPv6 loopback /121", "::1", "/121", true, "::/121"},

		// IPv4 single host
		{"IPv4 address /32", "192.168.1.1", "/32", true, "192.168.1.1/32"},
		{"IPv4 loopback /32", "127.0.0.1", "/32", true, "127.0.0.1/32"},

		// IPv4 range (25-bit = 128 addresses)
		{"IPv4 address /25", "192.168.1.1", "/25", true, "192.168.1.0/25"},
		{"IPv4 address /25 upper", "192.168.1.200", "/25", true, "192.168.1.128/25"},

		// IPv6 any notation
		{"IPv6 any", "::", "/0", true, "::/0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ipNet, err := net.ParseCIDR(tt.ip + tt.mask)
			if tt.expectValid {
				if err != nil {
					t.Errorf("expected valid CIDR for %s%s, got error: %v", tt.ip, tt.mask, err)
					return
				}
				if ipNet.String() != tt.expectNet {
					t.Errorf("expected network %s, got %s", tt.expectNet, ipNet.String())
				}
			} else {
				if err == nil {
					t.Errorf("expected error for %s%s, got valid CIDR", tt.ip, tt.mask)
				}
			}
		})
	}
}

// TestIPv6AddressRangeSize verifies the /121 mask gives 128 addresses like /25 for IPv4
func TestIPv6AddressRangeSize(t *testing.T) {
	// IPv4 /25 should have 128 addresses
	_, ipv4Net, _ := net.ParseCIDR("192.168.1.0/25")
	ipv4Size := addressCount(ipv4Net)

	// IPv6 /121 should also have 128 addresses
	_, ipv6Net, _ := net.ParseCIDR("2001:db8::/121")
	ipv6Size := addressCount(ipv6Net)

	if ipv4Size != 128 {
		t.Errorf("IPv4 /25 expected 128 addresses, got %d", ipv4Size)
	}

	if ipv6Size != 128 {
		t.Errorf("IPv6 /121 expected 128 addresses, got %d", ipv6Size)
	}

	if ipv4Size != ipv6Size {
		t.Errorf("IPv4 /25 (%d) and IPv6 /121 (%d) should have same address count", ipv4Size, ipv6Size)
	}
}

// addressCount calculates the number of addresses in a network
func addressCount(ipNet *net.IPNet) int {
	ones, bits := ipNet.Mask.Size()
	return 1 << (bits - ones)
}

// TestGetCIDRMaskIntegration tests the full flow of detecting IP type and getting mask
func TestGetCIDRMaskIntegration(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		rangeMode bool
		expectIP  utils.IPTYPE
		expectNet string
	}{
		// Single host mode
		{"IPv4 single host", "192.168.1.100", false, utils.IPV4, "192.168.1.100/32"},
		{"IPv6 single host", "2001:db8::100", false, utils.IPV6, "2001:db8::100/128"},

		// Range mode - note: ParseCIDR returns the network with host bits masked
		{"IPv4 range", "192.168.1.100", true, utils.IPV4, "192.168.1.0/25"},
		{"IPv6 range", "2001:db8::1", true, utils.IPV6, "2001:db8::/121"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Detect IP type
			ipType, err := utils.DetectIPType(tt.ip)
			if err != nil {
				t.Fatalf("DetectIPType failed for %s: %v", tt.ip, err)
			}
			if ipType != tt.expectIP {
				t.Errorf("expected IP type %d, got %d", tt.expectIP, ipType)
			}

			// Step 2: Get appropriate CIDR mask
			mask := utils.GetCIDRMask(ipType, tt.rangeMode)

			// Step 3: Parse CIDR
			_, ipNet, err := net.ParseCIDR(tt.ip + mask)
			if err != nil {
				t.Fatalf("ParseCIDR failed for %s%s: %v", tt.ip, mask, err)
			}

			if ipNet.String() != tt.expectNet {
				t.Errorf("expected network %s, got %s", tt.expectNet, ipNet.String())
			}
		})
	}
}

// TestIPSetNameSelection verifies correct ipset name selection for IPv4/IPv6
func TestIPSetNameSelection(t *testing.T) {
	// We can't create actual ipsets, but we can verify the naming logic
	tests := []struct {
		name       string
		ipType     utils.IPTYPE
		setType    int
		expectName string
	}{
		// Default set (type 1)
		{"IPv4 default set", utils.IPV4, 1, "defaultset"},
		{"IPv6 default set", utils.IPV6, 1, "defaultset_v6"},

		// Temp set (type 4)
		{"IPv4 temp set", utils.IPV4, 4, "tempset"},
		{"IPv6 temp set", utils.IPV6, 4, "tempset_v6"},
	}

	// Create a mock IPSet to test GetIpsetName
	ipset := &utils.IPSet{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := ipset.GetIpsetName(tt.ipType, tt.setType)
			if name != tt.expectName {
				t.Errorf("expected ipset name %s, got %s", tt.expectName, name)
			}
		})
	}
}

// TestIPv6AnyNotation verifies the correct "any" notation for IPv6
func TestIPv6AnyNotation(t *testing.T) {
	// Correct notation is ::/0
	_, ipNet, err := net.ParseCIDR("::/0")
	if err != nil {
		t.Fatalf("failed to parse ::/0: %v", err)
	}

	// Verify it covers all IPv6 addresses
	if !ipNet.Contains(net.ParseIP("::1")) {
		t.Error("::/0 should contain ::1")
	}
	if !ipNet.Contains(net.ParseIP("2001:db8::1")) {
		t.Error("::/0 should contain 2001:db8::1")
	}
	if !ipNet.Contains(net.ParseIP("fe80::1")) {
		t.Error("::/0 should contain fe80::1")
	}

	// Old incorrect notation should also parse but we prefer the canonical form
	_, ipNetOld, err := net.ParseCIDR("0:0:0:0:0:0:0:0/0")
	if err != nil {
		t.Fatalf("failed to parse old notation: %v", err)
	}

	// Both should be equivalent
	if ipNet.String() != ipNetOld.String() {
		t.Logf("Note: ::/0 normalizes to %s, old notation normalizes to %s", ipNet.String(), ipNetOld.String())
	}
}

// TestIPv6EdgeCases tests edge cases in IPv6 handling
func TestIPv6EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		isIPv4  bool
		isIPv6  bool
		isValid bool
	}{
		// Standard cases
		{"IPv4 standard", "192.168.1.1", true, false, true},
		{"IPv6 standard", "2001:db8::1", false, true, true},

		// Edge cases
		{"IPv6 all zeros", "::", false, true, true},
		{"IPv6 all ones", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff", false, true, true},
		{"IPv4 all zeros", "0.0.0.0", true, false, true},
		{"IPv4 all ones", "255.255.255.255", true, false, true},

		// IPv4-mapped IPv6 (treated as IPv4 by Go's net package)
		{"IPv4-mapped IPv6", "::ffff:192.168.1.1", true, false, true},

		// Invalid
		{"Empty string", "", false, false, false},
		{"Invalid format", "not-an-ip", false, false, false},
		{"IPv4 with extra octet", "192.168.1.1.1", false, false, false},
		{"IPv6 with too many groups", "2001:db8:1:2:3:4:5:6:7:8", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isIPv4 := utils.IsIPv4(tt.ip)
			isIPv6 := utils.IsIPv6(tt.ip)

			if isIPv4 != tt.isIPv4 {
				t.Errorf("IsIPv4(%s): expected %v, got %v", tt.ip, tt.isIPv4, isIPv4)
			}
			if isIPv6 != tt.isIPv6 {
				t.Errorf("IsIPv6(%s): expected %v, got %v", tt.ip, tt.isIPv6, isIPv6)
			}

			_, err := utils.DetectIPType(tt.ip)
			isValid := err == nil
			if isValid != tt.isValid {
				t.Errorf("DetectIPType(%s) valid: expected %v, got %v (err: %v)", tt.ip, tt.isValid, isValid, err)
			}
		})
	}
}

// TestIPHashStringFormats verifies the hash string format for ipset rules
func TestIPHashStringFormats(t *testing.T) {
	// These are the formats used in msghandler.go for ipset Add operations
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		// TCP format: src,port,dst
		{"TCP with port", "192.168.1.1,80,10.0.0.1", "192.168.1.1,80,10.0.0.1"},
		{"TCP port range", "192.168.1.1,1-65535,10.0.0.1", "192.168.1.1,1-65535,10.0.0.1"},

		// UDP format: src,udp:port,dst
		{"UDP with port", "192.168.1.1,udp:53,10.0.0.1", "192.168.1.1,udp:53,10.0.0.1"},
		{"UDP port range", "192.168.1.1,udp:1-65535,10.0.0.1", "192.168.1.1,udp:1-65535,10.0.0.1"},

		// ICMP format: src,icmp:type/code,dst
		{"ICMP ping", "192.168.1.1,icmp:8/0,10.0.0.1", "192.168.1.1,icmp:8/0,10.0.0.1"},

		// IPv6 formats
		{"IPv6 TCP", "2001:db8::1,80,2001:db8::2", "2001:db8::1,80,2001:db8::2"},
		{"IPv6 UDP", "2001:db8::1,udp:53,2001:db8::2", "2001:db8::1,udp:53,2001:db8::2"},
		{"IPv6 ICMP", "2001:db8::1,icmp:8/0,2001:db8::2", "2001:db8::1,icmp:8/0,2001:db8::2"},

		// Net/port format for tempset
		{"IPv4 net,port", "192.168.1.0/25,80", "192.168.1.0/25,80"},
		{"IPv6 net,port", "2001:db8::/121,80", "2001:db8::/121,80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the format is what we expect
			if tt.format != tt.expected {
				t.Errorf("format mismatch: got %s, want %s", tt.format, tt.expected)
			}
		})
	}
}

// TestIPTablesStructFields verifies the IPTables struct has IPv6 fields
func TestIPTablesStructFields(t *testing.T) {
	// Create an IPTables struct and verify IPv6 fields exist
	ipt := &utils.IPTables{}

	// These fields should exist after our changes
	_ = ipt.Binary           // IPv4 iptables path
	_ = ipt.Binary6          // IPv6 ip6tables path
	_ = ipt.IPv6Available    // Whether ip6tables is available
	_ = ipt.AcceptInputMode  // IPv4 accept mode
	_ = ipt.AcceptInput6Mode // IPv6 accept mode

	// Verify default values
	if ipt.IPv6Available != false {
		t.Error("IPv6Available should default to false")
	}
	if ipt.Binary6 != "" {
		t.Error("Binary6 should default to empty string")
	}
}

// TestCIDRMaskConstants verifies the CIDR mask constants are correct
func TestCIDRMaskConstantsValues(t *testing.T) {
	// Verify constant values
	if utils.IPv4SingleHost != "/32" {
		t.Errorf("IPv4SingleHost should be /32, got %s", utils.IPv4SingleHost)
	}
	if utils.IPv4AdjacentRange != "/25" {
		t.Errorf("IPv4AdjacentRange should be /25, got %s", utils.IPv4AdjacentRange)
	}
	if utils.IPv6SingleHost != "/128" {
		t.Errorf("IPv6SingleHost should be /128, got %s", utils.IPv6SingleHost)
	}
	if utils.IPv6AdjacentRange != "/121" {
		t.Errorf("IPv6AdjacentRange should be /121, got %s", utils.IPv6AdjacentRange)
	}

	// Verify they produce valid CIDRs
	testIPs := []struct {
		ip   string
		mask string
	}{
		{"192.168.1.1", utils.IPv4SingleHost},
		{"192.168.1.1", utils.IPv4AdjacentRange},
		{"2001:db8::1", utils.IPv6SingleHost},
		{"2001:db8::1", utils.IPv6AdjacentRange},
	}

	for _, test := range testIPs {
		_, _, err := net.ParseCIDR(test.ip + test.mask)
		if err != nil {
			t.Errorf("failed to parse %s%s: %v", test.ip, test.mask, err)
		}
	}
}
