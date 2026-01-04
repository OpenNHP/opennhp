package utils

import (
	"fmt"
	"net"
)

// CIDR mask constants for IP address handling
const (
	IPv4SingleHost    = "/32"  // Single IPv4 host
	IPv4AdjacentRange = "/25"  // 128 IPv4 addresses
	IPv6SingleHost    = "/128" // Single IPv6 host
	IPv6AdjacentRange = "/121" // 128 IPv6 addresses (equivalent to IPv4 /25)
)

// DetectIPType parses an IP address string and returns whether it's IPv4 or IPv6.
// Returns an error if the IP address is invalid.
func DetectIPType(ipStr string) (IPTYPE, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ipStr)
	}
	if ip.To4() != nil {
		return IPV4, nil
	}
	return IPV6, nil
}

// IsIPv6 returns true if the string is a valid IPv6 address.
// Note: IPv4-mapped IPv6 addresses (::ffff:x.x.x.x) return true.
func IsIPv6(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil && ip.To4() == nil
}

// IsIPv4 returns true if the string is a valid IPv4 address.
func IsIPv4(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	return ip != nil && ip.To4() != nil
}

// GetCIDRMask returns the appropriate CIDR mask suffix for a given IP type and access mode.
// If rangeMode is true, returns the range mask (128 addresses); otherwise returns single-host mask.
func GetCIDRMask(ipType IPTYPE, rangeMode bool) string {
	if ipType == IPV6 {
		if rangeMode {
			return IPv6AdjacentRange
		}
		return IPv6SingleHost
	}
	// IPv4 or default
	if rangeMode {
		return IPv4AdjacentRange
	}
	return IPv4SingleHost
}
