//go:build linux

package ebpf

import (
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cilium/ebpf"
	"golang.org/x/sys/unix"
)

// serverObjPath is where `make ebpf-objs` puts the server's XDP object.
func serverObjPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "release", "nhp-server", "etc", "nhp_server_xdp.o")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("%s not built; run `make ebpf-objs` to exercise this test", path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

// newRelayMap creates a standalone map with the same shape as `nhp_relay_ips`
// in nhp/ebpf/xdp/nhp_server_xdp.c, so ReplaceRelayIPs can be exercised
// without loading the program or needing an interface to attach to. Creating
// any BPF map still needs privilege, so this skips when unprivileged.
func newRelayMap(t *testing.T) *ebpf.Map {
	t.Helper()
	return newRelayMapSized(t, 4096)
}

// newRelayMapSized is newRelayMap with a chosen max_entries, so a test can make
// the kernel refuse an insert (E2BIG) on demand and exercise what ReplaceRelayIPs
// does to a live whitelist when a map write fails.
func newRelayMapSized(t *testing.T, maxEntries uint32) *ebpf.Map {
	t.Helper()
	m, err := ebpf.NewMap(&ebpf.MapSpec{
		Name:       "test_relay_ips",
		Type:       ebpf.LPMTrie,
		KeySize:    8, // __u32 prefixlen + __be32 addr
		ValueSize:  1,
		MaxEntries: maxEntries,
		Flags:      unix.BPF_F_NO_PREALLOC,
	})
	if err != nil {
		t.Skipf("cannot create BPF map (needs CAP_BPF): %v", err)
	}
	t.Cleanup(func() { m.Close() })
	return m
}

func relayMapContents(t *testing.T, m *ebpf.Map) []string {
	t.Helper()
	var out []string
	var key RelayPrefixKey
	var value uint8
	iter := m.Iterate()
	for iter.Next(&key, &value) {
		out = append(out, key.String())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterate: %v", err)
	}
	sort.Strings(out)
	return out
}

// hostKey is the key the eBPF side looks an incoming packet up with: a /32 of
// the source address. Matching one against a stored /24 is the whole point of
// the trie.
func hostKey(a, b, c, d byte) RelayPrefixKey {
	return RelayPrefixKey{PrefixLen: 32, Addr: [4]byte{a, b, c, d}}
}

func TestReplaceRelayIPsKeysAreWireOrder(t *testing.T) {
	m := newRelayMap(t)

	if err := ReplaceRelayIPs(m, []string{"1.2.3.4"}); err != nil {
		t.Fatalf("ReplaceRelayIPs: %v", err)
	}

	// The eBPF side builds the key straight from iph->saddr, so the address
	// half must be the four bytes as they appear on the wire. Looking it up by
	// those exact bytes is the check that matters: a host-order key would
	// still round-trip through our own iterator but would never match a
	// packet.
	var value uint8
	key := hostKey(1, 2, 3, 4)
	if err := m.Lookup(&key, &value); err != nil {
		t.Fatalf("lookup by wire-order key failed: %v", err)
	}
	if value != 1 {
		t.Errorf("value = %d, want 1", value)
	}

	// Guard the same property from the other direction: the byte-swapped
	// address must NOT match.
	var swapped [4]byte
	binary.LittleEndian.PutUint32(swapped[:], binary.BigEndian.Uint32([]byte{1, 2, 3, 4}))
	swappedKey := RelayPrefixKey{PrefixLen: 32, Addr: swapped}
	if err := m.Lookup(&swappedKey, &value); err == nil {
		t.Error("byte-swapped key is present; keys are being written in host order")
	}
}

func TestReplaceRelayIPsIsAFullReplacement(t *testing.T) {
	m := newRelayMap(t)

	if err := ReplaceRelayIPs(m, []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}); err != nil {
		t.Fatalf("initial: %v", err)
	}
	if got, want := relayMapContents(t, m), []string{"10.0.0.1/32", "10.0.0.2/32", "10.0.0.3/32"}; !equal(got, want) {
		t.Fatalf("after initial = %v, want %v", got, want)
	}

	// Overlapping replacement: .2 stays, .1 and .3 go, .9 arrives.
	if err := ReplaceRelayIPs(m, []string{"10.0.0.2", "10.0.0.9"}); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if got, want := relayMapContents(t, m), []string{"10.0.0.2/32", "10.0.0.9/32"}; !equal(got, want) {
		t.Fatalf("after replace = %v, want %v", got, want)
	}

	// An address carried across a reload must never blink out of the map:
	// the relay's own SSH session rides on it.
	var value uint8
	key := hostKey(10, 0, 0, 2)
	if err := m.Lookup(&key, &value); err != nil {
		t.Errorf("carried-over address is missing: %v", err)
	}

	// Replacing with nothing is the one replacement that is refused: an empty
	// map drops SSH from every source, and the host has no other way in.
	if err := ReplaceRelayIPs(m, nil); err == nil {
		t.Error("ReplaceRelayIPs(m, nil) returned nil, want a refusal to empty the map")
	}
	if got, want := relayMapContents(t, m), []string{"10.0.0.2/32", "10.0.0.9/32"}; !equal(got, want) {
		t.Errorf("after the refused clear = %v, want the previous list %v", got, want)
	}
}

// A whitelist that reaches the kernel empty is the lockout this whole path
// exists to prevent, and a list of strings is not the same thing as a list of
// prefixes: ["relay.opennhp.org"], [""] or ["10.0.1.300"] all count as
// non-empty to a caller that only measures len(). The map keeps what it had.
func TestReplaceRelayIPsRefusesToEmptyTheMap(t *testing.T) {
	for _, tc := range []struct {
		name string
		ips  []string
	}{
		{"nil", nil},
		{"empty slice", []string{}},
		{"one empty string", []string{""}},
		{"whitespace", []string{"   "}},
		{"a hostname", []string{"relay.opennhp.org"}},
		{"a typo", []string{"10.0.1.300"}},
		{"IPv6 only", []string{"2001:db8::1", "2001:db8::/32"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := newRelayMap(t)
			if err := ReplaceRelayIPs(m, []string{"10.0.1.0/24"}); err != nil {
				t.Fatalf("seed: %v", err)
			}

			if err := ReplaceRelayIPs(m, tc.ips); err == nil {
				t.Errorf("ReplaceRelayIPs(%v) returned nil, want an error", tc.ips)
			}
			if got, want := relayMapContents(t, m), []string{"10.0.1.0/24"}; !equal(got, want) {
				t.Errorf("contents = %v, want the seeded %v", got, want)
			}
		})
	}
}

// A prefix entry covers every host in it, which is what keeps the server
// reachable when the relay is replaced and comes back with a different private
// address out of the same subnet. Without it the live whitelist names an
// address nothing has any more, SSH from the new relay is dropped, and the only
// path CI has for pushing a corrected whitelist is the one that just closed.
func TestReplaceRelayIPsAcceptsPrefixes(t *testing.T) {
	m := newRelayMap(t)

	if err := ReplaceRelayIPs(m, []string{"10.0.1.0/24", "203.0.113.7"}); err != nil {
		t.Fatalf("ReplaceRelayIPs: %v", err)
	}

	var value uint8
	for _, host := range [][4]byte{{10, 0, 1, 4}, {10, 0, 1, 250}} {
		key := RelayPrefixKey{PrefixLen: 32, Addr: host}
		if err := m.Lookup(&key, &value); err != nil {
			t.Errorf("%v is not matched by 10.0.1.0/24: %v", net.IP(host[:]), err)
		}
	}

	// The prefix must not reach beyond itself.
	outside := hostKey(10, 0, 2, 4)
	if err := m.Lookup(&outside, &value); err == nil {
		t.Error("10.0.2.4 matched 10.0.1.0/24; the prefix length is not being honored")
	}

	// Host bits below the prefix are masked off, so a prefix written with them
	// set is the same entry — otherwise the stale sweep would not recognize its
	// own keys and a reload would leave both behind.
	if err := ReplaceRelayIPs(m, []string{"10.0.1.99/24", "203.0.113.7"}); err != nil {
		t.Fatalf("re-apply: %v", err)
	}
	if got, want := relayMapContents(t, m), []string{"10.0.1.0/24", "203.0.113.7/32"}; !equal(got, want) {
		t.Errorf("contents = %v, want %v", got, want)
	}
}

// A malformed entry fails the whole call and writes nothing. Applying the rest
// would be the more forgiving reading and the wrong one: the entry that did not
// parse may be the only one the operator's SSH arrives from, and the map cannot
// be corrected from anywhere the map itself does not admit.
func TestReplaceRelayIPsRejectsInvalidEntries(t *testing.T) {
	m := newRelayMap(t)

	if err := ReplaceRelayIPs(m, []string{"10.0.1.0/24"}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	for _, ips := range [][]string{
		{"10.0.0.1", "not-an-ip"},
		{"10.0.0.1", ""},
		{"10.0.0.1", "  "},
		{"10.0.0.1", "2001:db8::1"},
		{"10.0.0.1", "10.0.0.0/33"},
		{"10.0.0.1", "2001:db8::/32"},
		{"10.0.0.1", "10.0.1.4 # relay"},
	} {
		if err := ReplaceRelayIPs(m, ips); err == nil {
			t.Errorf("ReplaceRelayIPs(%v) returned nil, want an error", ips)
		}
		if got, want := relayMapContents(t, m), []string{"10.0.1.0/24"}; !equal(got, want) {
			t.Fatalf("after %v the map = %v, want the seeded %v", ips, got, want)
		}
	}

	// Surrounding whitespace is not a typo, and duplicates are not either.
	if err := ReplaceRelayIPs(m, []string{" 10.0.0.1 ", "10.0.0.2", "10.0.0.1"}); err != nil {
		t.Fatalf("ReplaceRelayIPs: %v", err)
	}
	if got, want := relayMapContents(t, m), []string{"10.0.0.1/32", "10.0.0.2/32"}; !equal(got, want) {
		t.Errorf("contents = %v, want %v", got, want)
	}
}

// A map write that fails must leave the previous whitelist whole. This is the
// lockout the parse-time gates cannot see: the list is valid, the file is
// right, and the kernel simply will not take an entry (E2BIG here, ENOMEM on a
// no-prealloc trie in the field). Sweeping stale entries anyway turns "failed
// to add the relay's new address" into "removed its old one too" — a map
// holding neither, i.e. tcp/22 closed to every source, while applyXdpConfig
// logs that it kept the active whitelist.
func TestReplaceRelayIPsKeepsTheOldListWhenAnAddFails(t *testing.T) {
	// Room for the two seeded entries plus exactly one of the two new ones,
	// so whichever is attempted second is refused by the kernel.
	m := newRelayMapSized(t, 3)

	seeded := []string{"10.0.0.1/32", "10.0.0.2/32"}
	if err := ReplaceRelayIPs(m, []string{"10.0.0.1", "10.0.0.2"}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := ReplaceRelayIPs(m, []string{"10.0.0.3", "10.0.0.4"}); err == nil {
		t.Error("ReplaceRelayIPs returned nil, want the failed insert reported")
	}
	if got := relayMapContents(t, m); !equal(got, seeded) {
		t.Errorf("after the failed replace the map = %v, want the previous list %v", got, seeded)
	}

	// The addresses the old list named are still matched, which is the
	// property that actually keeps SSH working.
	var value uint8
	for _, host := range [][4]byte{{10, 0, 0, 1}, {10, 0, 0, 2}} {
		key := RelayPrefixKey{PrefixLen: 32, Addr: host}
		if err := m.Lookup(&key, &value); err != nil {
			t.Errorf("%v was dropped from the whitelist by a failed reload: %v", net.IP(host[:]), err)
		}
	}

	// A list the map has no room for at all fails the same way, with nothing
	// added to roll back.
	full := newRelayMapSized(t, 1)
	if err := ReplaceRelayIPs(full, []string{"10.0.0.1"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := ReplaceRelayIPs(full, []string{"10.0.0.2"}); err == nil {
		t.Error("ReplaceRelayIPs returned nil, want the failed insert reported")
	}
	if got, want := relayMapContents(t, full), []string{"10.0.0.1/32"}; !equal(got, want) {
		t.Errorf("after the failed replace the map = %v, want the previous list %v", got, want)
	}
}

func TestReplaceRelayIPsWithoutMap(t *testing.T) {
	if err := ReplaceRelayIPs(nil, []string{"10.0.0.1"}); err == nil {
		t.Error("ReplaceRelayIPs(nil, ...) returned nil, want an error")
	}
}

// End-to-end: the compiled object passes the kernel verifier, attaches, and
// exposes the maps the Go side expects by name. Needs CAP_BPF + CAP_NET_ADMIN,
// so it skips when unprivileged. Attaches to loopback to avoid disturbing the
// host's real traffic.
func TestServerEngineLoadAttachesToLoopback(t *testing.T) {
	objPath := serverObjPath(t)

	h, err := EngineLoad(EngineLoadParams{
		Variant:          VariantServer,
		IfaceName:        "lo",
		ProgObjPath:      objPath,
		ComponentId:      "test",
		LogDirPath:       t.TempDir(),
		LogLevel:         1,
		NhpPort:          62206,
		NhpMinFrameBytes: 240,
		// The loader installs the whitelist before it attaches anything, so a
		// load with no usable prefix fails outright (see
		// TestServerEngineLoadRefusesAnUnusableWhitelist).
		RelayIPs: []string{"127.0.0.1"},
	})
	if err != nil {
		t.Skipf("cannot load/attach the server XDP program (needs CAP_BPF + CAP_NET_ADMIN): %v", err)
	}
	t.Cleanup(func() { CleanupBPFFiles(VariantServer) })

	if h.RelayIPsMap == nil {
		t.Error("RelayIPsMap is nil")
	}
	if h.EventsMap == nil {
		t.Error("EventsMap is nil")
	}
	if h.XdpLink == nil {
		t.Error("XdpLink is nil")
	}
	// The server attaches no TC program: an egress hook here would be a
	// behavior change borrowed from the AC.
	if h.TcLink != nil {
		t.Error("TcLink is set; the server must not attach a TC egress program")
	}

	if err := ReplaceRelayIPs(h.RelayIPsMap, []string{"127.0.0.1"}); err != nil {
		t.Fatalf("ReplaceRelayIPs on the live map: %v", err)
	}
	var value uint8
	liveKey := hostKey(127, 0, 0, 1)
	if err := h.RelayIPsMap.Lookup(&liveKey, &value); err != nil {
		t.Errorf("lookup on the live map: %v", err)
	}

	// Pins land where CleanupBPFFiles looks for them; a mismatch would leave
	// stale maps behind and fail the next start.
	for _, pin := range pinnedFiles(VariantServer, DefaultPinDir) {
		if _, err := os.Stat(pin); err != nil {
			t.Errorf("expected pin %s: %v", pin, err)
		}
	}

	CleanupBPFFiles(VariantServer)
	for _, pin := range pinnedFiles(VariantServer, DefaultPinDir) {
		if _, err := os.Stat(pin); !os.IsNotExist(err) {
			t.Errorf("pin %s survived cleanup (err=%v)", pin, err)
		}
	}
}

// Functional proof of the two port-pair exceptions, and of the property they
// must not break.
//
// The regex test above guards the shape of the C; this one asks the kernel.
// The receiving socket in each case is *bound and not connected* — the exact
// shape has_local_flow() refuses, and the shape every DHCP client and chronyd
// use — so a datagram only arrives if the program's own branch passed it. The
// last case is the control: the same unconnected-receiver setup on an ordinary
// port must still be dropped, or the exceptions would have been widened into
// "anything addressed to a bound port gets in", which is the whole filter
// given away.
//
// Runs on loopback so it cannot disturb the host's real traffic, and skips
// without CAP_BPF + CAP_NET_ADMIN (binding udp/68 and udp/123 needs privilege
// too).
func TestServerFilterPassesClientRepliesAndDropsUnsolicited(t *testing.T) {
	objPath := serverObjPath(t)

	_, err := EngineLoad(EngineLoadParams{
		Variant:          VariantServer,
		IfaceName:        "lo",
		ProgObjPath:      objPath,
		ComponentId:      "test",
		LogDirPath:       t.TempDir(),
		LogLevel:         1,
		NhpPort:          62206,
		NhpMinFrameBytes: 240,
		RelayIPs:         []string{"127.0.0.1"},
	})
	if err != nil {
		t.Skipf("cannot load/attach the server XDP program (needs CAP_BPF + CAP_NET_ADMIN): %v", err)
	}
	t.Cleanup(func() { CleanupBPFFiles(VariantServer) })

	tests := []struct {
		name       string
		serverPort int // the source port of the "reply"
		clientPort int // the port the host's client has bound
		wantPass   bool
	}{
		{"dhcp lease renewal", 67, 68, true},
		{"ntp poll answer", 123, 123, true},
		{"unsolicited datagram to a bound port", 40000, 40001, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lo := net.IPv4(127, 0, 0, 1)

			// Bound, never connected: the kernel's socket table cannot tell
			// this from a listener, which is why has_local_flow() says no.
			rx, err := net.ListenUDP("udp4", &net.UDPAddr{IP: lo, Port: tc.clientPort})
			if err != nil {
				t.Skipf("cannot bind udp/%d: %v", tc.clientPort, err)
			}
			defer rx.Close()

			// NTP answers arrive on the very socket the poll went out of
			// (chronyd binds udp/123 and both sends and receives on it), so
			// that case sends from `rx` itself rather than opening a second
			// socket on a port already taken.
			if tc.serverPort == tc.clientPort {
				if _, err := rx.WriteToUDP([]byte("reply"), &net.UDPAddr{IP: lo, Port: tc.clientPort}); err != nil {
					t.Fatalf("send: %v", err)
				}
			} else {
				tx, err := net.DialUDP("udp4", &net.UDPAddr{IP: lo, Port: tc.serverPort}, &net.UDPAddr{IP: lo, Port: tc.clientPort})
				if err != nil {
					t.Skipf("cannot send from udp/%d: %v", tc.serverPort, err)
				}
				defer tx.Close()

				if _, err := tx.Write([]byte("reply")); err != nil {
					t.Fatalf("send: %v", err)
				}
			}

			buf := make([]byte, 64)
			_ = rx.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, readErr := rx.Read(buf)

			if tc.wantPass {
				if readErr != nil {
					t.Errorf("%d -> %d was dropped (%v); the host's client traffic must get in",
						tc.serverPort, tc.clientPort, readErr)
				}
				return
			}
			if readErr == nil {
				t.Errorf("%d -> %d arrived (%d bytes); unsolicited datagrams must stay dropped",
					tc.serverPort, tc.clientPort, n)
			}
		})
	}
}

// IPv6 is filtered, not waved through.
//
// The first version of the program answered `case ETH_P_IPV6: return XDP_PASS`,
// on the grounds that the demo hosts have no v6 service. sshd binds [::]:22 by
// default, though, so on any host with a routable v6 address — or on the demo
// the day a v6 CIDR is added to the VPC — that one line made tcp/22 and every
// other port reachable with no whitelist at all, while the deploy's "filter
// attached" check stayed green. Both halves of the replacement are asserted
// here: an unsolicited v6 datagram to a bound-but-unconnected socket must be
// dropped (the hole), and the reply to a connected one must still arrive (the
// v6 spelling of the DNS outage this filter caused over v4).
func TestServerFilterFiltersIPv6(t *testing.T) {
	objPath := serverObjPath(t)
	logDirPath := t.TempDir()

	_, err := EngineLoad(EngineLoadParams{
		Variant:     VariantServer,
		IfaceName:   "lo",
		ProgObjPath: objPath,
		ComponentId: "test",
		LogDirPath:  logDirPath,
		// LogLevelInfo, unlike the other tests here: the event lines are
		// Info, and this one reads them back.
		LogLevel:         2,
		NhpPort:          62206,
		NhpMinFrameBytes: 240,
		RelayIPs:         []string{"127.0.0.1"},
	})
	if err != nil {
		t.Skipf("cannot load/attach the server XDP program (needs CAP_BPF + CAP_NET_ADMIN): %v", err)
	}
	t.Cleanup(func() { CleanupBPFFiles(VariantServer) })

	lo6 := &net.UDPAddr{IP: net.IPv6loopback}

	t.Run("unsolicited datagram to a bound port", func(t *testing.T) {
		rx, err := net.ListenUDP("udp6", lo6)
		if err != nil {
			t.Skipf("cannot bind an IPv6 socket on ::1: %v", err)
		}
		defer rx.Close()

		tx, err := net.DialUDP("udp6", nil, rx.LocalAddr().(*net.UDPAddr))
		if err != nil {
			t.Skipf("cannot dial ::1: %v", err)
		}
		defer tx.Close()

		if _, err := tx.Write([]byte("scan")); err != nil {
			t.Fatalf("send: %v", err)
		}

		_ = rx.SetReadDeadline(time.Now().Add(2 * time.Second))
		if n, err := rx.Read(make([]byte, 64)); err == nil {
			t.Errorf("an unsolicited IPv6 datagram arrived (%d bytes); IPv6 is bypassing the filter", n)
		}
	})

	t.Run("reply to a connected socket", func(t *testing.T) {
		peer, err := net.ListenUDP("udp6", lo6)
		if err != nil {
			t.Skipf("cannot bind an IPv6 socket on ::1: %v", err)
		}
		defer peer.Close()

		// Connected, i.e. what a resolver or an SMTP dialer holds: the kernel's
		// socket table can vouch for this one, and has_local_flow6() is what
		// asks it. Only the inbound direction is asserted — `peer` stands in
		// for the remote server, and on loopback its own socket is subject to
		// the same filter, which drops datagrams to an unconnected socket by
		// design (the case above).
		client, err := net.DialUDP("udp6", lo6, peer.LocalAddr().(*net.UDPAddr))
		if err != nil {
			t.Skipf("cannot dial ::1: %v", err)
		}
		defer client.Close()

		if _, err := peer.WriteToUDP([]byte("answer"), client.LocalAddr().(*net.UDPAddr)); err != nil {
			t.Fatalf("reply: %v", err)
		}
		_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
		if _, err := client.Read(make([]byte, 64)); err != nil {
			t.Errorf("the reply to the host's own IPv6 flow was dropped (%v); v6 client traffic must get in", err)
		}
	})

	// The event record is a packed struct read by offset (serverEventByteSize),
	// and carrying v6 addresses is what made it grow past the IPv4-shaped 25
	// bytes. A mismatch between the two sides would not fail any of the
	// verdict tests above — it would just print the wrong bytes in the log
	// nobody reads until an incident — so a dropped v6 datagram is read back
	// out of the event log here.
	t.Run("the drop is logged with its IPv6 source", func(t *testing.T) {
		rx, err := net.ListenUDP("udp6", lo6)
		if err != nil {
			t.Skipf("cannot bind an IPv6 socket on ::1: %v", err)
		}
		defer rx.Close()

		tx, err := net.DialUDP("udp6", nil, rx.LocalAddr().(*net.UDPAddr))
		if err != nil {
			t.Skipf("cannot dial ::1: %v", err)
		}
		defer tx.Close()

		// Re-sent on every pass rather than relying on the drop from the first
		// subtest: the perf reader is started in a goroutine by EngineLoad, and
		// bpf_perf_event_output() has nowhere to put an event until it is up.
		var last string
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if _, err := tx.Write([]byte("scan")); err != nil {
				t.Fatalf("send: %v", err)
			}
			time.Sleep(100 * time.Millisecond)

			matches, _ := filepath.Glob(filepath.Join(logDirPath, "logs", serverEventLogName+"-*.log"))
			for _, path := range matches {
				body, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				last = string(body)
				if strings.Contains(last, "REASON=V6_OTHER") && strings.Contains(last, "SRC=::1") {
					return
				}
			}
		}
		t.Fatalf("no 'REASON=V6_OTHER ... SRC=::1' line in the event log; the event record and its parser disagree. Log was:\n%s", last)
	})
}

// A non-first fragment has no L4 header, and the program must not read one out
// of it.
//
// It used to: the TCP and UDP branches took the header at ihl*4 unconditionally,
// so a later fragment yielded garbage ports. The visible effect was that any
// datagram over the path MTU lost its tail (the first fragment passed, the rest
// were dropped as UDP_OTHER, reassembly timed out, the knock vanished); the
// invisible one was that a crafted fragment whose payload read 67 -> 68 took the
// DHCP exception. The source is asserted rather than the packets because
// fragmenting on loopback means an MTU change on a shared interface.
func TestNonFirstFragmentsAreSeparatedFromTheL4Tree(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "ebpf", "xdp", "nhp_server_xdp.c"))
	if err != nil {
		t.Fatalf("read nhp_server_xdp.c: %v", err)
	}

	fragIdx := strings.Index(string(src), "IP_OFFSET)")
	if fragIdx < 0 {
		t.Fatal("no IP_OFFSET test; fragments are being parsed as if they carried an L4 header")
	}
	// The guard has to come before the branches that dereference the L4
	// header, or it decides nothing.
	for _, marker := range []string{"struct tcphdr *tcp = (void *)iph", "struct udphdr *udp = (void *)iph"} {
		l4Idx := strings.Index(string(src), marker)
		if l4Idx < 0 {
			t.Fatalf("%q not found; has the program been restructured?", marker)
		}
		if fragIdx > l4Idx {
			t.Errorf("the IP_OFFSET test comes after %q, so fragments still reach it", marker)
		}
	}

	// And IPv6 must not be an unconditional pass any more.
	if regexp.MustCompile(`case ETH_P_IPV6:\s*return XDP_PASS`).Match(src) {
		t.Error("IPv6 is passed unconditionally; every port on a v6-addressed host is unfiltered")
	}
}

// The pin directory is EngineLoadParams.PinDir, and the cleanup has to use the
// same one.
//
// It used to delete a hard-coded list of /sys/fs/bpf paths whatever the load
// had used, which is wrong in both directions: a load with another PinDir
// leaked its pins (and its next LoadAndAssign then fails against a stale map,
// the exact failure this cleanup exists to prevent), while the sweep reached
// into /sys/fs/bpf and removed the pins of whatever filter was running on the
// host — including from a test, or from an operator's netns rehearsal on a live
// server.
func TestCleanupRemovesThePinsOfTheDirectoryItWasGiven(t *testing.T) {
	for _, variant := range []EngineVariant{VariantServer, VariantAC} {
		mine := t.TempDir()
		theirs := t.TempDir()

		// Stand-ins for the pins: cleanup only ever os.Remove()s these paths,
		// so plain files exercise it without bpffs or privilege.
		for _, dir := range []string{mine, theirs} {
			for _, path := range pinnedFiles(variant, dir) {
				if err := os.WriteFile(path, nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
		}
		keep := filepath.Join(mine, "not_ours")
		if err := os.WriteFile(keep, nil, 0600); err != nil {
			t.Fatal(err)
		}

		// The exported entry point takes no directory: it must use the one the
		// load recorded, not the default.
		restore := activePinDir(variant)
		t.Cleanup(func() { recordPinDir(variant, restore) })
		recordPinDir(variant, mine)
		CleanupBPFFiles(variant)

		for _, path := range pinnedFiles(variant, mine) {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("variant %d: %s survived cleanup (err=%v)", variant, path, err)
			}
		}
		for _, path := range pinnedFiles(variant, theirs) {
			if _, err := os.Stat(path); err != nil {
				t.Errorf("variant %d: cleanup reached outside its own pin directory and removed %s", variant, path)
			}
		}
		if _, err := os.Stat(keep); err != nil {
			t.Errorf("variant %d: cleanup removed a file it does not own: %v", variant, err)
		}

		// And the same thing with the directory spelled out, which is how a
		// caller that pinned somewhere without going through EngineLoad cleans
		// up after itself.
		cleanupPins(variant, theirs)
		for _, path := range pinnedFiles(variant, theirs) {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("variant %d: %s survived cleanupPins (err=%v)", variant, path, err)
			}
		}
	}
}

// The knock port is not a property of the object file.
//
// It used to be: nhp_server_xdp.c hard-coded 62206, so attaching the filter on
// a server whose ListenPort is anything else dropped every knock at the driver,
// invisibly — the daemon sees no packets and says nothing. The port is now a
// .rodata constant the loader rewrites from the daemon's own listen address, and
// this is the test of that. The receiving socket is bound and never connected,
// exactly the shape has_local_flow() refuses, so the datagram can only arrive
// through the knock-port branch.
func TestServerFilterKnocksOnTheConfiguredPort(t *testing.T) {
	objPath := serverObjPath(t)

	const (
		knockPort = 40206 // deliberately not the object's compiled-in default
		otherPort = 40207
		floor     = 240
	)

	_, err := EngineLoad(EngineLoadParams{
		Variant:          VariantServer,
		IfaceName:        "lo",
		ProgObjPath:      objPath,
		ComponentId:      "test",
		LogDirPath:       t.TempDir(),
		LogLevel:         1,
		NhpPort:          knockPort,
		NhpMinFrameBytes: floor,
		RelayIPs:         []string{"127.0.0.1"},
	})
	if err != nil {
		t.Skipf("cannot load/attach the server XDP program (needs CAP_BPF + CAP_NET_ADMIN): %v", err)
	}
	t.Cleanup(func() { CleanupBPFFiles(VariantServer) })

	tests := []struct {
		name     string
		dstPort  int
		payload  int
		wantPass bool
	}{
		// A full-length knock on the configured port: this is the case that
		// fails outright if the rewrite did not happen.
		{"knock on the configured port", knockPort, 300, true},
		// Same datagram, ordinary port: still dropped, so the pass above is the
		// knock branch and not something that admits any bound port.
		{"knock-sized datagram on another port", otherPort, 300, false},
		// The floor still applies on the configured port.
		{"short datagram on the configured port", knockPort, 8, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lo := net.IPv4(127, 0, 0, 1)

			rx, err := net.ListenUDP("udp4", &net.UDPAddr{IP: lo, Port: tc.dstPort})
			if err != nil {
				t.Skipf("cannot bind udp/%d: %v", tc.dstPort, err)
			}
			defer rx.Close()

			tx, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: lo, Port: tc.dstPort})
			if err != nil {
				t.Skipf("cannot dial udp/%d: %v", tc.dstPort, err)
			}
			defer tx.Close()

			if _, err := tx.Write(make([]byte, tc.payload)); err != nil {
				t.Fatalf("send: %v", err)
			}

			buf := make([]byte, 1024)
			_ = rx.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, readErr := rx.Read(buf)

			if tc.wantPass {
				if readErr != nil {
					t.Errorf("a %d-byte datagram to udp/%d was dropped (%v); the filter is not using the configured knock port",
						tc.payload, tc.dstPort, readErr)
				}
				return
			}
			if readErr == nil {
				t.Errorf("a %d-byte datagram to udp/%d arrived (%d bytes), want it dropped",
					tc.payload, tc.dstPort, n)
			}
		})
	}
}

// A failed load must leave nothing behind.
//
// The caller treats an error as "fail-open, no ingress filter" and logs it that
// way, and it never receives an EngineHandle — so any pin or link surviving the
// error is state nothing owns. For the server that is the worst case there is:
// a filter enforcing an empty SSH whitelist on a host whose only way in is that
// whitelist, while the log says the filter is off. This drives the failure
// through resolveInterface, which runs after the maps and the program have
// already been pinned.
func TestServerEngineLoadLeavesNothingBehindOnFailure(t *testing.T) {
	objPath := serverObjPath(t)

	// Probe first: without CAP_BPF the load fails before it has pinned
	// anything, so the test would pass without proving anything.
	h, err := EngineLoad(EngineLoadParams{
		Variant:          VariantServer,
		IfaceName:        "lo",
		ProgObjPath:      objPath,
		ComponentId:      "test",
		LogDirPath:       t.TempDir(),
		LogLevel:         1,
		NhpPort:          62206,
		NhpMinFrameBytes: 240,
		RelayIPs:         []string{"127.0.0.1"},
	})
	if err != nil {
		t.Skipf("cannot load/attach the server XDP program (needs CAP_BPF + CAP_NET_ADMIN): %v", err)
	}
	_ = h
	CleanupBPFFiles(VariantServer)

	t.Cleanup(func() { CleanupBPFFiles(VariantServer) })
	if _, err := EngineLoad(EngineLoadParams{
		Variant:          VariantServer,
		IfaceName:        "nhp-no-such-iface",
		ProgObjPath:      objPath,
		ComponentId:      "test",
		LogDirPath:       t.TempDir(),
		LogLevel:         1,
		NhpPort:          62206,
		NhpMinFrameBytes: 240,
		RelayIPs:         []string{"127.0.0.1"},
	}); err == nil {
		t.Fatal("EngineLoad on a nonexistent interface returned nil, want an error")
	}

	for _, pin := range pinnedFiles(VariantServer, DefaultPinDir) {
		if _, err := os.Stat(pin); !os.IsNotExist(err) {
			t.Errorf("pin %s survived a failed load (err=%v); the daemon believes it is fail-open", pin, err)
		}
	}
	if serverXdpLink != nil {
		t.Error("serverXdpLink is set after a failed load; the filter is attached with no handle to drive it")
	}
}

// The whitelist is installed before the program is attached, and a whitelist
// that cannot be installed fails the load.
//
// This is the ordering the whole filter rests on. The alternative — attach,
// then write the map — has two bad outcomes on a host whose only way in is that
// whitelist: a window after every restart in which SSH from the relay is
// dropped, and, if the write fails, a filter left attached enforcing an empty
// trie, i.e. tcp/22 closed to every source for good. Asserted the only way user
// space can: a load whose RelayIPs names no prefix the trie can hold must come
// back as an error with nothing pinned and nothing attached, exactly like the
// pre-attach failures above.
func TestServerEngineLoadRefusesAnUnusableWhitelist(t *testing.T) {
	objPath := serverObjPath(t)

	// Probe first, for the same reason as the test above: an unprivileged load
	// fails before it ever reaches the whitelist, which would prove nothing.
	if _, err := EngineLoad(EngineLoadParams{
		Variant:          VariantServer,
		IfaceName:        "lo",
		ProgObjPath:      objPath,
		ComponentId:      "test",
		LogDirPath:       t.TempDir(),
		LogLevel:         1,
		NhpPort:          62206,
		NhpMinFrameBytes: 240,
		RelayIPs:         []string{"127.0.0.1"},
	}); err != nil {
		t.Skipf("cannot load/attach the server XDP program (needs CAP_BPF + CAP_NET_ADMIN): %v", err)
	}
	CleanupBPFFiles(VariantServer)

	t.Cleanup(func() { CleanupBPFFiles(VariantServer) })

	for _, tc := range []struct {
		name     string
		relayIPs []string
	}{
		{"no whitelist at all", nil},
		{"an unset RELAY_IPS", []string{""}},
		{"a hostname", []string{"relay.opennhp.org"}},
		{"one bad entry among good ones", []string{"127.0.0.1", "not-an-ip"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := EngineLoad(EngineLoadParams{
				Variant:          VariantServer,
				IfaceName:        "lo",
				ProgObjPath:      objPath,
				ComponentId:      "test",
				LogDirPath:       t.TempDir(),
				LogLevel:         1,
				NhpPort:          62206,
				NhpMinFrameBytes: 240,
				RelayIPs:         tc.relayIPs,
			}); err == nil {
				CleanupBPFFiles(VariantServer)
				t.Fatalf("EngineLoad with RelayIPs=%v returned nil, want a refusal", tc.relayIPs)
			}

			for _, pin := range pinnedFiles(VariantServer, DefaultPinDir) {
				if _, err := os.Stat(pin); !os.IsNotExist(err) {
					t.Errorf("pin %s survived the refusal (err=%v); the daemon believes it is fail-open", pin, err)
				}
			}
			if serverXdpLink != nil {
				t.Error("serverXdpLink is set; the filter is attached with an empty SSH whitelist")
			}
		})
	}
}

// The loader refuses rather than attaching an object it cannot configure: a
// filter whose knock port is not the daemon's drops every knock at the driver
// with nothing in user space to say so, which is strictly worse than the
// fail-open state the daemon already handles.
func TestSetServerConstantsRefusesAnObjectWithoutThem(t *testing.T) {
	spec := &ebpf.CollectionSpec{}

	err := setServerConstants(spec, EngineLoadParams{NhpPort: 62206, NhpMinFrameBytes: 240})
	if err == nil {
		t.Fatal("setServerConstants on an object with no constants returned nil, want an error")
	}
	if !strings.Contains(err.Error(), varNhpListenPort) {
		t.Errorf("error %q does not name the missing constant %q", err, varNhpListenPort)
	}

	// Nothing asked for, nothing to rewrite: the object's own defaults stand.
	if err := setServerConstants(spec, EngineLoadParams{}); err != nil {
		t.Errorf("setServerConstants with no overrides: %v", err)
	}
}

// The compiled object must actually expose the constants the loader rewrites;
// a rename on either side would otherwise only show up as a refused load on a
// real host.
func TestServerObjectExposesTheRewritableConstants(t *testing.T) {
	spec, err := ebpf.LoadCollectionSpec(serverObjPath(t))
	if err != nil {
		t.Fatalf("load collection spec: %v", err)
	}

	for _, name := range []string{varNhpListenPort, varNhpMinUdpLen} {
		v, ok := spec.Variables[name]
		if !ok {
			t.Errorf("object has no %q variable", name)
			continue
		}
		if !v.Constant() {
			t.Errorf("%q is not in .rodata; the verifier will not fold it", name)
		}
		if v.Size() != 2 {
			t.Errorf("%q is %d bytes, want 2 (__u16)", name, v.Size())
		}
	}
}

// The action codes are a wire format between nhp_server_xdp.c and
// serverActionName(): the C side puts a bare byte on the perf ring and this
// side is the only thing that gives it a meaning. Parse the constants out of
// the source rather than restating them, so a new branch in the program that
// nobody mirrored here shows up as a failing test instead of as
// "REASON=UNKNOWN-7" in an event log months later.
func TestServerActionNamesCoverTheProgramsActionCodes(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "ebpf", "xdp", "nhp_server_xdp.c"))
	if err != nil {
		t.Fatalf("read nhp_server_xdp.c: %v", err)
	}

	re := regexp.MustCompile(`(?m)^#define\s+(ACT_\w+)\s+(\d+)`)
	matches := re.FindAllStringSubmatch(string(src), -1)
	if len(matches) == 0 {
		t.Fatal("no ACT_* constants found; has the program been renamed?")
	}

	seen := make(map[uint8]string, len(matches))
	for _, m := range matches {
		code, err := strconv.ParseUint(m[2], 10, 8)
		if err != nil {
			t.Fatalf("%s = %q: %v", m[1], m[2], err)
		}
		if other, dup := seen[uint8(code)]; dup {
			t.Errorf("%s and %s share action code %d", other, m[1], code)
		}
		seen[uint8(code)] = m[1]

		verdict, reason := serverActionName(uint8(code))
		if strings.HasPrefix(reason, "UNKNOWN-") {
			t.Errorf("%s (%d) has no name in serverActionName", m[1], code)
		}
		// A DROP_* constant that logs as PASS (or the reverse) would make the
		// event log say the opposite of what the kernel did.
		wantDrop := strings.HasPrefix(m[1], "ACT_DROP_")
		if got := verdict == "DROP"; got != wantDrop {
			t.Errorf("%s (%d) logs verdict %q", m[1], code, verdict)
		}
	}
}

// The bulk classes must stay unreported.
//
// TCP_ESTABLISHED, UDP_ESTABLISHED and post-SYN SSH_RELAY are the verdicts
// whose rate is the host's *throughput*, not its event rate: one release scp
// or one `dnf update` is tens of thousands of them, all saying the same thing
// about the same already-decided flow. They were logged per packet once, and
// the fix was to count them instead (record_packet's `report` argument). A
// well-meaning edit that hands one of them `true` again would look harmless in
// review and quietly turn the event log back into a disk-filler, so the shape
// of those call sites is asserted here rather than left to memory.
func TestBulkPassClassesAreCountedNotLogged(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "ebpf", "xdp", "nhp_server_xdp.c"))
	if err != nil {
		t.Fatalf("read nhp_server_xdp.c: %v", err)
	}

	// The last argument of each record_packet() / record_v6_packet() call, by
	// action. Both recorders are matched: the v6 branch reports through its own
	// entry point, and a drop class that goes unlogged is a blind spot whichever
	// family it is in.
	re := regexp.MustCompile(`record(?:_v6)?_packet\(ctx,\s*(ACT_\w+),[^;]*?,\s*([a-z_]+)\);`)
	matches := re.FindAllStringSubmatch(string(src), -1)
	if len(matches) == 0 {
		t.Fatal("no record_packet() call sites found; has the program been restructured?")
	}

	report := make(map[string]string, len(matches))
	for _, m := range matches {
		report[m[1]] = m[2]
	}

	// Counted only, never a line per packet. ACT_V6_ICMP_CONTROL is on the
	// list for a second reason: neighbor discovery and MLD are constant
	// background chatter on a v6 subnet, and drowning the ring in it is what
	// the old unconditional IPv6 pass was avoiding.
	for _, action := range []string{
		"ACT_TCP_ESTABLISHED", "ACT_UDP_ESTABLISHED",
		"ACT_V6_ESTABLISHED", "ACT_V6_ICMP_CONTROL",
	} {
		got, ok := report[action]
		if !ok {
			t.Errorf("%s is no longer recorded at all", action)
			continue
		}
		if got != "false" {
			t.Errorf("%s is reported with report=%s; bulk flow traffic must be counted only", action, got)
		}
	}

	// One line per SSH session, not per segment.
	if got := report["ACT_SSH_RELAY"]; got != "is_syn" {
		t.Errorf("ACT_SSH_RELAY is reported with report=%s; want the SYN-only guard is_syn", got)
	}

	// A lease renewal is a couple of datagrams an hour and the host's address
	// depends on it, so it stays on the ring: after the outage it caused, the
	// DHCP_CLIENT line is the one an operator greps for to see the exception
	// working.
	if got, ok := report["ACT_DHCP_CLIENT"]; !ok {
		t.Error("ACT_DHCP_CLIENT is not recorded; DHCP replies must be admitted and visible")
	} else if got != "true" {
		t.Errorf("ACT_DHCP_CLIENT is reported with report=%s; want true", got)
	}
	// chrony polls every 32-64s, so this one belongs in the summary, not the
	// per-packet log.
	if got, ok := report["ACT_NTP_CLIENT"]; !ok {
		t.Error("ACT_NTP_CLIENT is no longer recorded at all")
	} else if got != "false" {
		t.Errorf("ACT_NTP_CLIENT is reported with report=%s; a periodic poll should be counted only", got)
	}

	// Every drop stays on the ring (the token bucket, not the call site, is
	// what bounds those) -- a silent drop class is a blind spot.
	for action, got := range report {
		if strings.HasPrefix(action, "ACT_DROP_") && got != "true" {
			t.Errorf("%s is reported with report=%s; drops must stay visible", action, got)
		}
	}

	// The old unconditional emitter must be gone, or half the call sites
	// would bypass the counters entirely.
	if strings.Contains(string(src), "submit_event(") {
		t.Error("submit_event() still present; all call sites should go through record_packet()")
	}
}

// The two client protocols the socket table cannot vouch for must be PASSed,
// by port pair.
//
// This is the regression with the highest blast radius in the whole program:
// the DHCP reply (67 -> 68) arrives on a raw or merely-bound socket, so
// has_local_flow() cannot admit it, and dropping it costs the host its IPv4
// address one lease later — every port dark, no SSH, no knock, no way in. The
// NTP pair (123 -> 123) is the same shape one layer up and costs the clock.
// Both are three lines of C that look removable in a cleanup, so the verdict
// is asserted here.
func TestClientProtocolsWithoutAConnectedSocketArePassed(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "ebpf", "xdp", "nhp_server_xdp.c"))
	if err != nil {
		t.Fatalf("read nhp_server_xdp.c: %v", err)
	}

	for _, tc := range []struct {
		action string
		guard  string
	}{
		{"ACT_DHCP_CLIENT", `udp->source == bpf_htons\(DHCP_PORT_SERVER\)`},
		{"ACT_NTP_CLIENT", `udp->source == bpf_htons\(NTP_PORT\)`},
	} {
		re := regexp.MustCompile(`(?s)` + tc.guard + `.*?record_packet\(ctx, ` + tc.action + `.*?return (XDP_\w+);`)
		m := re.FindStringSubmatch(string(src))
		if m == nil {
			t.Errorf("no branch guarded by the %s port pair reporting %s; the host's own client traffic would be dropped", tc.action, tc.action)
			continue
		}
		if m[1] != "XDP_PASS" {
			t.Errorf("the %s branch returns %s, want XDP_PASS", tc.action, m[1])
		}
	}
}

// NHP_EVENT_ACTIONS sizes the per-action arrays on the C side and the summary
// loop on this one. If the C grew a slot and this did not, the extra classes
// would simply never be summarized.
func TestServerEventActionsMatchesTheProgram(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("..", "..", "ebpf", "xdp", "nhp_server_xdp.c"))
	if err != nil {
		t.Fatalf("read nhp_server_xdp.c: %v", err)
	}

	m := regexp.MustCompile(`(?m)^#define\s+NHP_EVENT_ACTIONS\s+(\d+)`).FindStringSubmatch(string(src))
	if m == nil {
		t.Fatal("NHP_EVENT_ACTIONS not found in nhp_server_xdp.c")
	}
	want, err := strconv.Atoi(m[1])
	if err != nil {
		t.Fatalf("NHP_EVENT_ACTIONS = %q: %v", m[1], err)
	}
	if serverEventActions != want {
		t.Errorf("serverEventActions = %d, NHP_EVENT_ACTIONS = %d", serverEventActions, want)
	}
}

// Retention is the other half of keeping the log bounded: the rate limiter
// caps how fast it grows, this caps how much of it survives. nhp/log rotates
// by date and prunes nothing, so without a sweep a long-lived host accumulates
// every day it has ever seen.
func TestSweepServerEventLogsKeepsTodayAndForeignFiles(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)

	today := writeLogFile(t, dir, "nhp_server_xdp-2026-09-29.log", now, 10)
	recent := writeLogFile(t, dir, "nhp_server_xdp-2026-09-27.log", now.AddDate(0, 0, -2), 10)
	stale := writeLogFile(t, dir, "nhp_server_xdp-2026-09-01.log", now.AddDate(0, 0, -28), 10)
	// The daemon's own log lives in the same directory and is not ours to
	// delete.
	foreign := writeLogFile(t, dir, "server-2026-09-01.log", now.AddDate(0, 0, -28), 10)

	sweepServerEventLogs(dir, now, 14, 1<<30)

	// Today's file is open and being appended to, so it is exempt; `recent`
	// is inside the age bound; `foreign` is another logger's.
	for _, path := range []string{today, recent, foreign} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s should have survived: %v", filepath.Base(path), err)
		}
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("%s should have been pruned (err=%v)", filepath.Base(stale), err)
	}
}

func TestSweepServerEventLogsEnforcesTheByteBudget(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)

	// Three days well inside the age bound, together over a 250-byte budget:
	// a flood can blow the budget in a day, which is exactly when the age
	// bound is no help.
	oldest := writeLogFile(t, dir, "nhp_server_xdp-2026-09-26.log", now.AddDate(0, 0, -3), 100)
	middle := writeLogFile(t, dir, "nhp_server_xdp-2026-09-27.log", now.AddDate(0, 0, -2), 100)
	newest := writeLogFile(t, dir, "nhp_server_xdp-2026-09-28.log", now.AddDate(0, 0, -1), 100)

	sweepServerEventLogs(dir, now, 14, 250)

	if _, err := os.Stat(oldest); !os.IsNotExist(err) {
		t.Errorf("the oldest file should go first (err=%v)", err)
	}
	for _, path := range []string{middle, newest} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s is inside the budget once the oldest is gone: %v", filepath.Base(path), err)
		}
	}
}

func writeLogFile(t *testing.T, dir, name string, mod time.Time, size int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, make([]byte, size), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, mod, mod); err != nil {
		t.Fatal(err)
	}
	return path
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
