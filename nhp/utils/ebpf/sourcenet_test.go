package ebpf

import "testing"

// TestParseSourceNet pins the parsing of the two source ranges the AC's
// client-protocol pairs are narrowed to.
//
// The byte order is the part worth pinning: the values end up in the
// filter's .rodata and are compared with the packet's own saddr field, which
// is in wire order, so an address written the wrong way round would narrow
// the pair to nothing at all -- silently, and with the log saying the pair
// was narrowed.
func TestParseSourceNet(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		wantAddr   uint32
		wantMask   uint32
		wantPrefix int
		wantSet    bool
		wantErr    bool
	}{
		{
			name:    "empty means no narrowing",
			in:      "",
			wantSet: false,
		},
		{
			name:       "a bare address is a host route",
			in:         "169.254.169.253",
			wantAddr:   0xA9FEA9FD,
			wantMask:   0xFFFFFFFF,
			wantPrefix: 32,
			wantSet:    true,
		},
		{
			name:       "a prefix keeps its network bits",
			in:         "169.254.169.0/24",
			wantAddr:   0xA9FEA900,
			wantMask:   0xFFFFFF00,
			wantPrefix: 24,
			wantSet:    true,
		},
		{
			name:       "host bits below the prefix are masked off",
			in:         "10.0.0.5/24",
			wantAddr:   0x0A000000,
			wantMask:   0xFFFFFF00,
			wantPrefix: 24,
			wantSet:    true,
		},
		{
			name:       "a /0 covers everything, which is no narrowing at all",
			in:         "0.0.0.0/0",
			wantAddr:   0,
			wantMask:   0,
			wantPrefix: 0,
			wantSet:    false,
		},
		{
			name:    "a prefix longer than an address is refused",
			in:      "10.0.0.1/33",
			wantErr: true,
		},
		{
			name:    "an IPv6 address is refused",
			in:      "::1",
			wantErr: true,
		},
		{
			name:    "an IPv4-mapped IPv6 address is refused too",
			in:      "::ffff:10.0.0.1",
			wantErr: true,
		},
		{
			name:    "a hostname is refused",
			in:      "example.com",
			wantErr: true,
		},
		{
			name:    "rubbish is refused",
			in:      "169.254.169.300",
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseSourceNet(c.in)
			if c.wantErr {
				if err == nil {
					t.Fatalf("ParseSourceNet(%q) = %v, want an error", c.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSourceNet(%q): %v", c.in, err)
			}
			if got.Addr != c.wantAddr {
				t.Errorf("ParseSourceNet(%q) address = 0x%08X, want 0x%08X", c.in, got.Addr, c.wantAddr)
			}
			if got.Mask != c.wantMask {
				t.Errorf("ParseSourceNet(%q) mask = 0x%08X, want 0x%08X", c.in, got.Mask, c.wantMask)
			}
			if got.IsSet() != c.wantSet {
				t.Errorf("ParseSourceNet(%q) IsSet() = %v, want %v", c.in, got.IsSet(), c.wantSet)
			}
			if c.wantSet && prefixLenOf(got.Mask) != c.wantPrefix {
				t.Errorf("ParseSourceNet(%q) mask spells a /%d, want /%d", c.in, prefixLenOf(got.Mask), c.wantPrefix)
			}
		})
	}
}

// TestSourceNetString pins the spelling used in the log line that says what
// the filter was narrowed to. It is the only place the operator sees the
// value the kernel got, so it has to be the same value they configured.
func TestSourceNetString(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"169.254.169.253", "169.254.169.253/32"},
		{"169.254.169.0/24", "169.254.169.0/24"},
		{"10.0.0.5/24", "10.0.0.0/24"},
		{"", "any source"},
	}

	for _, c := range cases {
		got, err := ParseSourceNet(c.in)
		if err != nil {
			t.Fatalf("ParseSourceNet(%q): %v", c.in, err)
		}
		if got.String() != c.want {
			t.Errorf("ParseSourceNet(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
