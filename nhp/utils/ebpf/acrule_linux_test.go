//go:build linux

package ebpf

import (
	"errors"
	"testing"

	"github.com/cilium/ebpf"
)

// The rule writers must never reach into a live filter's maps.
//
// EbpfRuleAdd opens the AC's pinned maps by path, and the path used to be
// hard-coded to /sys/fs/bpf. On a host that runs nhp-acd, that is the live
// filter's own map: a test that called EbpfRuleAdd there installed real
// whitelist entries -- a real grant, to a real peer -- into a running access
// controller, and never removed them. A test that opens a port on a live
// access controller is not a test.
//
// The pin directory is injectable (see pindir.go), so the tests below point
// the writers at a directory of their own, load the AC's objects into it,
// and read the grants back from there. Nothing outside the test's own
// directory is touched.

// useAcPinDir points the AC's rule writers at dir for the duration of the
// test, and restores the previous one afterwards.
//
// The writers are the only thing that changes: the maps themselves are the
// AC's own, loaded from the real object files, so the keys and the values
// are the ones the kernel's programs will look for.
func useAcPinDir(t *testing.T, dir string) {
	t.Helper()
	previous := activePinDir(VariantAC)
	recordPinDir(VariantAC, dir)
	t.Cleanup(func() { recordPinDir(VariantAC, previous) })
}

// TestEbpfRuleAddWritesOneEntryPerProtocol pins the shape of a port-scoped
// grant.
//
// A port-scoped grant for a resource that names no protocol is written once
// for TCP and once for UDP -- the two entries the ipset mode writes for the
// same resource -- and not once for every protocol there is.
//
// The map it is written into is the AC's own, in a pin directory of this
// test's: the grants below are real entries in a real kernel map, but in a
// map no filter is attached to and no peer ever looks at.
func TestEbpfRuleAddWritesOneEntryPerProtocol(t *testing.T) {
	pinDir := testPinDir(t)
	_, _, maps := loadAcObjects(t, pinDir)
	useAcPinDir(t, pinDir)

	const srcIP = "10.0.0.2"
	const ttlSec = 30
	err := EbpfRuleAdd(MapTypeSrcAndPort, EbpfRuleParams{
		SrcIP:   srcIP,
		DstPort: 443,
	}, ttlSec)
	if err != nil {
		t.Fatalf("the grant was refused: %v", err)
	}

	now, err := getBootTimeNanos()
	if err != nil {
		t.Skipf("cannot read the monotonic clock: %v", err)
	}

	ip, err := parseIP(srcIP)
	if err != nil {
		t.Fatalf("parseIP(%q): %v", srcIP, err)
	}
	for _, protocol := range []uint8{ProtocolTCP, ProtocolUDP} {
		key := (&srcIPdstPortKey{SrcIP: ip, DstPort: 443, Protocol: protocol}).ToSpKey()
		var value whitelistValue
		if err := maps.Srcportlist.Lookup(key, &value); err != nil {
			t.Errorf("no src_port entry for protocol %d: %v", protocol, err)
			continue
		}
		// The entry has to be a live grant: the program admits the peer
		// while it holds, and an entry that expired the moment it was
		// written would admit nobody.
		if value.Allowed != 1 {
			t.Errorf("the src_port entry for protocol %d does not allow anything", protocol)
		}
		if value.ExpireTime < now {
			t.Errorf("the src_port entry for protocol %d has already expired", protocol)
		}
	}
	for _, protocol := range []uint8{ProtocolICMP, 47, 255} {
		key := (&srcIPdstPortKey{SrcIP: ip, DstPort: 443, Protocol: protocol}).ToSpKey()
		var value whitelistValue
		if err := maps.Srcportlist.Lookup(key, &value); err == nil {
			t.Errorf("the src_port map holds an entry for protocol %d, which the grant does not cover", protocol)
		} else if !errors.Is(err, ebpf.ErrKeyNotExist) {
			t.Errorf("the src_port map did not answer for protocol %d: %v", protocol, err)
		}
	}
}

// TestEbpfRuleAddWritesNothingForARefusedPortRange pins the other half of
// the refusal: that nothing was written.
//
// The refusal itself is checked without a map at all, in
// TestEbpfRuleAddRejectsANonWildcardPortRange, because it happens before
// the map is opened. Here the map exists, and it has to come out empty: an
// entry with a range the programs never look up is a grant that was reported
// and not given.
//
// Both ranges are checked: the refused one, and the wildcard one the
// programs do look up. The second is what makes the first meaningful --
// if the refusal had written anything at all for this peer, that is the key
// it would have written.
func TestEbpfRuleAddWritesNothingForARefusedPortRange(t *testing.T) {
	pinDir := testPinDir(t)
	_, _, maps := loadAcObjects(t, pinDir)
	useAcPinDir(t, pinDir)

	const srcIP = "10.0.0.2"
	err := EbpfRuleAdd(MapTypeSrcPortList, EbpfRuleParams{
		SrcIP:        srcIP,
		DstPortStart: 100,
		DstPortEnd:   200,
	}, 30)
	if err == nil {
		t.Fatal("a port_list entry with the range 100-200 was accepted; the eBPF programs " +
			"only ever look up the wildcard range, so such an entry grants nothing")
	}

	ip, err := parseIP(srcIP)
	if err != nil {
		t.Fatalf("parseIP(%q): %v", srcIP, err)
	}
	for _, protocol := range []uint8{ProtocolTCP, ProtocolUDP} {
		for _, r := range []struct {
			what       string
			start, end int
		}{
			{"the refused range", 100, 200},
			{"the wildcard range", AnyPortRangeStart, AnyPortRangeEnd},
		} {
			key := (&portListKey{
				SrcIP:        ip,
				DstPortStart: uint16(r.start),
				DstPortEnd:   uint16(r.end),
				Protocol:     protocol,
			}).ToPlKey()
			var value whitelistValue
			if err := maps.Portlist.Lookup(key, &value); err == nil {
				t.Errorf("%s was written to the port_list map for protocol %d", r.what, protocol)
			} else if !errors.Is(err, ebpf.ErrKeyNotExist) {
				t.Errorf("the port_list map did not answer for %s, protocol %d: %v", r.what, protocol, err)
			}
		}
	}
}
