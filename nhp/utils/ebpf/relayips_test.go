package ebpf

import (
	"testing"
)

func TestParseRelayPrefix(t *testing.T) {
	tests := []struct {
		in      string
		want    RelayPrefixKey
		wantErr bool
	}{
		{in: "1.2.3.4", want: RelayPrefixKey{PrefixLen: 32, Addr: [4]byte{1, 2, 3, 4}}},
		{in: " 10.0.1.7 ", want: RelayPrefixKey{PrefixLen: 32, Addr: [4]byte{10, 0, 1, 7}}},
		{in: "10.0.1.0/24", want: RelayPrefixKey{PrefixLen: 24, Addr: [4]byte{10, 0, 1, 0}}},
		{in: "10.0.1.200/24", want: RelayPrefixKey{PrefixLen: 24, Addr: [4]byte{10, 0, 1, 0}}},
		{in: "0.0.0.0/0", want: RelayPrefixKey{PrefixLen: 0, Addr: [4]byte{0, 0, 0, 0}}},
		{in: "not-an-ip", wantErr: true},
		{in: "relay.opennhp.org", wantErr: true},
		{in: "10.0.1.300", wantErr: true},
		{in: "10.0.1.4 # relay", wantErr: true},
		{in: "2001:db8::1", wantErr: true},
		{in: "2001:db8::/32", wantErr: true},
		{in: "10.0.0.0/33", wantErr: true},
		{in: "", wantErr: true},
		{in: "   ", wantErr: true},
	}

	for _, tc := range tests {
		got, err := ParseRelayPrefix(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseRelayPrefix(%q) = %v, want an error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseRelayPrefix(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseRelayPrefix(%q) = %+v, want %+v", tc.in, got, tc.want)
		}
	}
}

// The whole list is parsed before any of it counts. A caller that measures
// len(conf.RelayIPs) cannot tell ["relay.opennhp.org"] — which names no address
// the trie can hold — from a real whitelist, and attaching or reloading on the
// strength of that count is what closes tcp/22 for every source.
func TestParseRelayPrefixesIsAllOrNothing(t *testing.T) {
	tests := []struct {
		name    string
		in      []string
		want    []string
		wantErr bool
	}{
		{name: "nil", in: nil},
		{name: "empty", in: []string{}},
		{
			name: "addresses and prefixes",
			in:   []string{"10.0.1.4", "203.0.113.7", "10.0.1.0/24"},
			want: []string{"10.0.1.4/32", "203.0.113.7/32", "10.0.1.0/24"},
		},
		{
			name: "duplicates collapse, including two spellings of one prefix",
			in:   []string{"10.0.1.0/24", " 10.0.1.0/24 ", "10.0.1.99/24"},
			want: []string{"10.0.1.0/24"},
		},
		{name: "a hostname", in: []string{"relay.opennhp.org"}, wantErr: true},
		{name: "an inline comment", in: []string{"10.0.1.4 # relay"}, wantErr: true},
		{name: "a typo", in: []string{"10.0.1.300"}, wantErr: true},
		{name: "IPv6 only", in: []string{"2001:db8::1"}, wantErr: true},
		{name: "an unset RELAY_IPS", in: []string{""}, wantErr: true},
		// The dangerous half-valid case: everything parses but the one entry
		// SSH actually arrives from.
		{name: "one bad entry among good ones", in: []string{"10.0.1.0/24", "203.0.113.300"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRelayPrefixes(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseRelayPrefixes(%v) = %v, want an error", tc.in, got)
				}
				if got != nil {
					t.Errorf("ParseRelayPrefixes(%v) returned %v alongside its error, want nothing", tc.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRelayPrefixes(%v): %v", tc.in, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("ParseRelayPrefixes(%v) = %v, want %v", tc.in, got, tc.want)
			}
			for i, key := range got {
				if key.String() != tc.want[i] {
					t.Errorf("prefix %d = %s, want %s", i, key, tc.want[i])
				}
			}
		})
	}
}
