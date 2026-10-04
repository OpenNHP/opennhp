package ebpf

// The relay whitelist as user space sees it: the parsing and validation that
// decides what the `nhp_relay_ips` LPM trie may hold.
//
// This lives outside engine_linux.go on purpose. The whitelist is the only
// thing that reaches tcp/22 on a filtered host, so the server's config path has
// to be able to validate a rendered xdp.toml *before* it decides to attach
// anything — and that decision must compile and be testable on every platform,
// not only where the loader does.

import (
	"fmt"
	"net"
	"strings"
)

// RelayPrefixKey is the key of the `nhp_relay_ips` LPM trie in
// nhp/ebpf/xdp/nhp_server_xdp.c: struct relay_prefix_key.
//
// PrefixLen is a plain __u32 the kernel reads in host order; Addr is the
// network-order address, which is also the order an LPM trie compares prefixes
// in, so the eBPF side looks iph->saddr up with no byte swapping.
type RelayPrefixKey struct {
	PrefixLen uint32
	Addr      [4]byte
}

func (k RelayPrefixKey) String() string {
	return fmt.Sprintf("%s/%d", net.IP(k.Addr[:]).String(), k.PrefixLen)
}

// ParseRelayPrefix turns one xdp.toml whitelist entry into a trie key. A bare
// address is a /32; "10.0.1.0/24" is the subnet form that lets the whitelist
// survive the relay being replaced with a new private address.
//
// Host bits below the prefix are masked off so the key is canonical: the trie
// ignores them when matching, but a key that kept them would not compare equal
// to the same prefix written differently, and ReplaceRelayIPs' stale-entry
// sweep would then never recognize its own entries.
func ParseRelayPrefix(s string) (RelayPrefixKey, error) {
	s = strings.TrimSpace(s)

	if strings.Contains(s, "/") {
		_, ipNet, err := net.ParseCIDR(s)
		if err != nil {
			return RelayPrefixKey{}, fmt.Errorf("invalid CIDR %q: %w", s, err)
		}
		ip4 := ipNet.IP.To4()
		if ip4 == nil {
			return RelayPrefixKey{}, fmt.Errorf("only IPv4 prefixes are supported, got %q", s)
		}
		ones, _ := ipNet.Mask.Size()
		return RelayPrefixKey{PrefixLen: uint32(ones), Addr: [4]byte(ip4)}, nil
	}

	ip := net.ParseIP(s)
	if ip == nil {
		return RelayPrefixKey{}, fmt.Errorf("invalid IP address %q", s)
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return RelayPrefixKey{}, fmt.Errorf("only IPv4 addresses are supported, got %q", s)
	}
	return RelayPrefixKey{PrefixLen: 32, Addr: [4]byte(ip4)}, nil
}

// ParseRelayPrefixes parses a whole whitelist and fails on the first entry it
// cannot use, returning nothing. Duplicates (including two spellings of the
// same prefix) collapse into one key.
//
// All-or-nothing is the point. Skipping bad entries looks forgiving but is the
// dangerous reading of a typo: a list of one misspelled address parses to no
// prefixes at all, and a list where only the entry SSH really arrives from is
// misspelled parses to a whitelist that excludes the operator. Both are
// indistinguishable from a deliberate list at the map level, and on a host
// whose only way in is the whitelist the mistake cannot be corrected remotely.
// So a caller gets either every prefix the file asked for or an error to refuse
// on — see startXdpFilter and applyXdpConfig in endpoints/server/config.go.
//
// An entry that is empty or blank is an error like any other: that is what an
// unset RELAY_IPS renders to (`RelayIPs = [""]`), and a length check on the raw
// strings cannot tell it apart from a real one.
func ParseRelayPrefixes(ipStrs []string) ([]RelayPrefixKey, error) {
	keys := make([]RelayPrefixKey, 0, len(ipStrs))
	seen := make(map[RelayPrefixKey]struct{}, len(ipStrs))

	for i, s := range ipStrs {
		key, err := ParseRelayPrefix(s)
		if err != nil {
			return nil, fmt.Errorf("relay whitelist entry %d of %d: %w", i+1, len(ipStrs), err)
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	return keys, nil
}
