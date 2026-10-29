package ebpf

import (
	"net"
	"testing"

	utilsebpf "github.com/OpenNHP/opennhp/nhp/utils/ebpf"
)

// The watchdog's whole job is to tell "the host still has an address" apart
// from "the host has lost its DHCP lease", and the second state is not
// "no addresses at all": a client that cannot renew ends up with a
// 169.254.0.0/16 autoconfiguration address, sometimes alongside an IPv6 one.
// Counting either as an address would leave the filter attached on exactly the
// host nobody can reach.
func TestHasRoutableIPv4(t *testing.T) {
	cidr := func(s string) net.Addr {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			t.Fatalf("ParseCIDR(%q): %v", s, err)
		}
		ip, _, _ := net.ParseCIDR(s)
		return &net.IPNet{IP: ip, Mask: n.Mask}
	}

	tests := []struct {
		name  string
		addrs []net.Addr
		want  bool
	}{
		{"vpc address", []net.Addr{cidr("10.0.1.78/24")}, true},
		{"vpc address plus ipv6", []net.Addr{cidr("fe80::1/64"), cidr("10.0.1.78/24")}, true},
		{"no addresses at all", nil, false},
		{"lease lost, link-local only", []net.Addr{cidr("169.254.12.34/16")}, false},
		{"ipv6 only", []net.Addr{cidr("2001:db8::1/64"), cidr("fe80::1/64")}, false},
		{"loopback only", []net.Addr{cidr("127.0.0.1/8")}, false},
		{"unspecified", []net.Addr{cidr("0.0.0.0/32")}, false},
		{"link-local then real", []net.Addr{cidr("169.254.12.34/16"), cidr("10.0.1.78/24")}, true},
	}

	for _, tc := range tests {
		if got := hasRoutableIPv4(tc.addrs); got != tc.want {
			t.Errorf("%s: hasRoutableIPv4 = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// The watchdog must survive being handed a handle with no interface name
// (nothing to watch) rather than spinning or panicking.
func TestWatchHostAddressingWithoutInterfaceReturns(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		watchHostAddressing(&utilsebpf.EngineHandle{})
	}()
	<-done
}
