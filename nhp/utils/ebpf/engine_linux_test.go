//go:build linux

package ebpf

import (
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/cilium/ebpf"
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
	m, err := ebpf.NewMap(&ebpf.MapSpec{
		Name:       "test_relay_ips",
		Type:       ebpf.LRUHash,
		KeySize:    4,
		ValueSize:  1,
		MaxEntries: 4096,
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
	var key [4]byte
	var value uint8
	iter := m.Iterate()
	for iter.Next(&key, &value) {
		out = append(out, net.IPv4(key[0], key[1], key[2], key[3]).String())
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("iterate: %v", err)
	}
	sort.Strings(out)
	return out
}

func TestReplaceRelayIPsKeysAreWireOrder(t *testing.T) {
	m := newRelayMap(t)

	if err := ReplaceRelayIPs(m, []string{"1.2.3.4"}); err != nil {
		t.Fatalf("ReplaceRelayIPs: %v", err)
	}

	// The eBPF side compares the key against iph->saddr directly, so the key
	// must be the four bytes as they appear on the wire. Looking it up by
	// those exact bytes is the check that matters: a host-order key would
	// still round-trip through our own iterator but would never match a
	// packet.
	var value uint8
	if err := m.Lookup([]byte{1, 2, 3, 4}, &value); err != nil {
		t.Fatalf("lookup by wire-order key failed: %v", err)
	}
	if value != 1 {
		t.Errorf("value = %d, want 1", value)
	}

	// Guard the same property from the other direction: the byte-swapped key
	// must NOT be present.
	var swapped [4]byte
	binary.LittleEndian.PutUint32(swapped[:], binary.BigEndian.Uint32([]byte{1, 2, 3, 4}))
	if err := m.Lookup(swapped[:], &value); err == nil {
		t.Error("byte-swapped key is present; keys are being written in host order")
	}
}

func TestReplaceRelayIPsIsAFullReplacement(t *testing.T) {
	m := newRelayMap(t)

	if err := ReplaceRelayIPs(m, []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}); err != nil {
		t.Fatalf("initial: %v", err)
	}
	if got, want := relayMapContents(t, m), []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}; !equal(got, want) {
		t.Fatalf("after initial = %v, want %v", got, want)
	}

	// Overlapping replacement: .2 stays, .1 and .3 go, .9 arrives.
	if err := ReplaceRelayIPs(m, []string{"10.0.0.2", "10.0.0.9"}); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if got, want := relayMapContents(t, m), []string{"10.0.0.2", "10.0.0.9"}; !equal(got, want) {
		t.Fatalf("after replace = %v, want %v", got, want)
	}

	// An address carried across a reload must never blink out of the map:
	// the relay's own SSH session rides on it.
	var value uint8
	if err := m.Lookup([]byte{10, 0, 0, 2}, &value); err != nil {
		t.Errorf("carried-over address is missing: %v", err)
	}

	if err := ReplaceRelayIPs(m, nil); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if got := relayMapContents(t, m); len(got) != 0 {
		t.Errorf("after clear = %v, want empty", got)
	}
}

// A malformed entry is skipped, not fatal — one typo in xdp.toml should not
// decide the whole whitelist.
func TestReplaceRelayIPsSkipsInvalidEntries(t *testing.T) {
	m := newRelayMap(t)

	if err := ReplaceRelayIPs(m, []string{"10.0.0.1", "", "  ", "not-an-ip", "2001:db8::1", " 10.0.0.2 "}); err != nil {
		t.Fatalf("ReplaceRelayIPs: %v", err)
	}
	if got, want := relayMapContents(t, m), []string{"10.0.0.1", "10.0.0.2"}; !equal(got, want) {
		t.Errorf("contents = %v, want %v", got, want)
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
		Variant:     VariantServer,
		IfaceName:   "lo",
		ProgObjPath: objPath,
		ComponentId: "test",
		LogDirPath:  t.TempDir(),
		LogLevel:    1,
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
	// behaviour change borrowed from the AC.
	if h.TcLink != nil {
		t.Error("TcLink is set; the server must not attach a TC egress program")
	}

	if err := ReplaceRelayIPs(h.RelayIPsMap, []string{"127.0.0.1"}); err != nil {
		t.Fatalf("ReplaceRelayIPs on the live map: %v", err)
	}
	var value uint8
	if err := h.RelayIPsMap.Lookup([]byte{127, 0, 0, 1}, &value); err != nil {
		t.Errorf("lookup on the live map: %v", err)
	}

	// Pins land where CleanupBPFFiles looks for them; a mismatch would leave
	// stale maps behind and fail the next start.
	for _, pin := range serverPinnedFiles {
		if _, err := os.Stat(pin); err != nil {
			t.Errorf("expected pin %s: %v", pin, err)
		}
	}

	CleanupBPFFiles(VariantServer)
	for _, pin := range serverPinnedFiles {
		if _, err := os.Stat(pin); !os.IsNotExist(err) {
			t.Errorf("pin %s survived cleanup (err=%v)", pin, err)
		}
	}
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
