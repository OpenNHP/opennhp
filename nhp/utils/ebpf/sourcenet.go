package ebpf

// The source range the AC's client-protocol pairs are narrowed to.
//
// The DHCP pair (67 -> 68) and the NTP pair (123 -> 123) are admitted by port
// pair alone: the pair says what the answer looks like, not who sent it. On a
// host whose security group is open to the whole internet that leaves the
// filter admitting anything wearing the right ports, and the operator is the
// only one who knows which addresses their own dhclient and chronyd answer
// from.
//
// This is that operator's input, as user space sees it: the parsing and the
// validation that decide what the filter's .rodata holds.
//
// It lives outside engine_linux.go for the same reason relayips.go does: the
// value has to be parseable and testable on every platform, not only where
// the loader does, because the AC's configuration is parsed everywhere and a
// type that only existed on Linux would make the configuration's meaning
// depend on the host the daemon was built for.

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"strings"
)

// SourceNet is an IPv4 address range: the address, and the mask of the bits
// that are the network's.
//
// Both are the big-endian value of the corresponding bytes, because the eBPF
// side compares them with the packet's own saddr field -- which is in wire
// order -- and turns them back with bpf_htonl() before it does.
//
// A zero mask means "no range": every source is inside it. That is the
// default, and it is the behaviour of the filter before the ranges existed.
type SourceNet struct {
	Addr uint32
	Mask uint32
}

// IsSet reports whether the range narrows anything. A range of 0.0.0.0/0 is
// not one: it covers the whole address space, which is the same answer the
// filter gives when no range was configured at all.
func (n SourceNet) IsSet() bool {
	return n.Mask != 0
}

// String renders the range the way the configuration spells it, for the log
// line that says what the filter was narrowed to.
func (n SourceNet) String() string {
	if !n.IsSet() {
		return "any source"
	}
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], n.Addr)
	return fmt.Sprintf("%s/%d", netip.AddrFrom4(b).String(), prefixLenOf(n.Mask))
}

// ParseSourceNet turns one configured value into the pair the filter holds.
//
// A bare address is a /32; "169.254.169.0/24" is the subnet form, which lets
// one entry cover a host's own servers and the addresses they are given
// elsewhere. An empty value means "not configured" -- the pair is not narrowed
// at all -- and is not an error.
//
// Host bits below the prefix are masked off, so "10.0.0.5/24" is the
// 10.0.0.0/24 network rather than a range that starts at an address nobody
// serves from.
//
// Parsing goes through net/netip rather than net.ParseCIDR/net.ParseIP
// because only netip tells a real IPv4 address from an IPv4-mapped IPv6 one:
// net's To4() is non-nil for "::ffff:10.0.1.0", so an "IPv4 only" check built
// on it would pass a value the filter cannot compare against an IPv4 field.
func ParseSourceNet(s string) (SourceNet, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return SourceNet{}, nil
	}

	var (
		addr netip.Addr
		bits int
	)
	if strings.Contains(s, "/") {
		prefix, err := netip.ParsePrefix(s)
		if err != nil {
			return SourceNet{}, fmt.Errorf("invalid CIDR %q: %w", s, err)
		}
		if !prefix.Addr().Is4() {
			return SourceNet{}, fmt.Errorf("only IPv4 ranges are supported, got %q", s)
		}
		masked := prefix.Masked()
		addr, bits = masked.Addr(), masked.Bits()
	} else {
		parsed, err := netip.ParseAddr(s)
		if err != nil {
			return SourceNet{}, fmt.Errorf("invalid IP address %q: %w", s, err)
		}
		if !parsed.Is4() {
			return SourceNet{}, fmt.Errorf("only IPv4 addresses are supported, got %q", s)
		}
		addr, bits = parsed, 32
	}

	a4 := addr.As4()
	return SourceNet{
		Addr: binary.BigEndian.Uint32(a4[:]),
		Mask: binary.BigEndian.Uint32(prefixMaskBytes(bits)),
	}, nil
}

// prefixMaskBytes returns the mask of a prefix length, as the four bytes of
// the mask in wire order.
func prefixMaskBytes(bits int) []byte {
	b := make([]byte, 4)
	for i := 0; i < bits; i++ {
		b[i/8] |= 1 << (7 - uint(i%8))
	}
	return b
}

// prefixLenOf returns the prefix length a mask spells out.
func prefixLenOf(mask uint32) int {
	n := 0
	for i := 31; i >= 0; i-- {
		if mask&(1<<uint(i)) == 0 {
			break
		}
		n++
	}
	return n
}
