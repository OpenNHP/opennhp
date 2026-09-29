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

	// The last argument of each record_packet() call, by action.
	re := regexp.MustCompile(`record_packet\(ctx,\s*(ACT_\w+),[^;]*?,\s*([a-z_]+)\);`)
	matches := re.FindAllStringSubmatch(string(src), -1)
	if len(matches) == 0 {
		t.Fatal("no record_packet() call sites found; has the program been restructured?")
	}

	report := make(map[string]string, len(matches))
	for _, m := range matches {
		report[m[1]] = m[2]
	}

	// Counted only, never a line per packet.
	for _, action := range []string{"ACT_TCP_ESTABLISHED", "ACT_UDP_ESTABLISHED"} {
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

// NHP_EVENT_ACTIONS sizes the per-action arrays on the C side and the summary
// loop on this one. If the C grew a slot and this did not, the extra classes
// would simply never be summarised.
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
