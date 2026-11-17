package server

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

func knockPPD(t *testing.T, headerType int, txid uint64, sendTime int64) *core.PacketParserData {
	t.Helper()
	return &core.PacketParserData{
		HeaderType:     headerType,
		RemotePubKey:   bytes.Repeat([]byte{0x11}, core.PublicKeySize),
		SenderTrxId:    txid,
		RemoteSendTime: sendTime,
	}
}

// A knock accepted once is refused when it comes back, which is the whole
// point: the responder's monotonic timestamp check is per connection, so a
// replayer picking a fresh source port had the knock honored again and the
// AC opened for their address.
func TestKnockReplayRefusedAcrossConnections(t *testing.T) {
	s := &UdpServer{knockReplay: newKnockReplayCache()}
	now := time.Now().UnixNano()

	// A distinct txid per type, because the key is (sender, txid,
	// sendTime) and deliberately does not include the header type. Two
	// message types cannot legitimately share a txid anyway —
	// Device.NextCounterIndex() is a single atomic counter per device —
	// and leaving the type out makes the entry slightly stricter: one
	// accepted knock blocks a replay of it under any type, and the type
	// is covered by the packet HMAC regardless.
	for i, ht := range []int{core.NHP_KNK, core.NHP_RKN, core.NHP_EXT, core.DHP_KNK} {
		name := core.HeaderTypeToString(ht)
		ppd := knockPPD(t, ht, uint64(100+i), now)

		if err := s.dedupeRecvKNK(ppd); err != nil {
			t.Fatalf("%s: first observation refused: %v", name, err)
		}
		// Byte-identical replay. The source connection is not part of the
		// key on purpose — that is exactly what the old check keyed on.
		if err := s.dedupeRecvKNK(ppd); err == nil {
			t.Errorf("%s: a replayed knock was accepted", name)
		} else if err != common.ErrServerDuplicateTransaction {
			t.Errorf("%s: err = %v, want ErrServerDuplicateTransaction", name, err)
		}
	}
}

// Honest traffic must not be caught. Every knock is built with a fresh
// TransactionId from Device.NextCounterIndex(), so no two legitimate
// knocks share a key — including two knocks from the same agent in the
// same nanosecond, and the same counter value from two different agents.
func TestKnockReplayAllowsDistinctKnocks(t *testing.T) {
	s := &UdpServer{knockReplay: newKnockReplayCache()}
	now := time.Now().UnixNano()

	for i := uint64(1); i <= 100; i++ {
		if err := s.dedupeRecvKNK(knockPPD(t, core.NHP_KNK, i, now)); err != nil {
			t.Fatalf("knock %d with a fresh txid was refused: %v", i, err)
		}
	}

	// Same txid, different sender.
	other := knockPPD(t, core.NHP_KNK, 1, now)
	other.RemotePubKey = bytes.Repeat([]byte{0x22}, core.PublicKeySize)
	if err := s.dedupeRecvKNK(other); err != nil {
		t.Errorf("a different agent reusing a counter value was refused: %v", err)
	}

	// Same sender and txid, different send time: a counter reset after a
	// restart must not collide with the pre-restart entry.
	reset := knockPPD(t, core.NHP_KNK, 1, now+int64(time.Second))
	if err := s.dedupeRecvKNK(reset); err != nil {
		t.Errorf("a post-restart counter reset was refused: %v", err)
	}
}

// Messages that are not knocks fall through untouched, so the hook can sit
// in the shared dedupe chain without claiming packets the other hooks own.
func TestKnockReplayIgnoresOtherMessages(t *testing.T) {
	s := &UdpServer{knockReplay: newKnockReplayCache()}
	now := time.Now().UnixNano()
	for _, ht := range []int{core.NHP_ART, core.NHP_RLY, core.NHP_AOL, core.NHP_DRG, core.NHP_REG} {
		ppd := knockPPD(t, ht, 7, now)
		if err := s.dedupeRecvKNK(ppd); err != nil {
			t.Errorf("%s was claimed by the knock hook: %v", core.HeaderTypeToString(ht), err)
		}
		if err := s.dedupeRecvKNK(ppd); err != nil {
			t.Errorf("%s was claimed on repeat: %v", core.HeaderTypeToString(ht), err)
		}
	}
}

// A knock with an unusable sender key is refused rather than silently
// skipped — the key is what the entry is scoped to, so without it there is
// nothing to dedupe against.
func TestKnockReplayRefusesInvalidPubkey(t *testing.T) {
	s := &UdpServer{knockReplay: newKnockReplayCache()}
	ppd := knockPPD(t, core.NHP_KNK, 1, time.Now().UnixNano())
	ppd.RemotePubKey = []byte{0x01, 0x02}
	if err := s.dedupeRecvKNK(ppd); err != common.ErrServerMissingPeerPubkey {
		t.Errorf("err = %v, want ErrServerMissingPeerPubkey", err)
	}
}

// The entry has to outlive every knock the staleness floor still accepts,
// or a packet could leave the cache while remaining replayable.
func TestKnockReplayTTLCoversTheStalenessFloor(t *testing.T) {
	floor := time.Duration(core.DefaultRecvStalenessFloorSeconds) * time.Second
	if knockReplayCacheTTL <= floor {
		t.Errorf("TTL %v does not outlive the staleness floor %v: a knock could expire from the cache and still be accepted",
			knockReplayCacheTTL, floor)
	}
}

// The hook has to be in the dedupe chain, not merely defined.
//
// Every other test here calls dedupeRecvKNK directly, so none of them
// notices if the chain entry is deleted — and a dedupe function the core
// never calls protects nothing. Asserted on the source because installing
// the real hook needs a Device, a listener and a peer that can produce an
// authenticated packet.
func TestKnockDedupeIsInstalledInTheChain(t *testing.T) {
	src, err := os.ReadFile("udpserver.go")
	if err != nil {
		t.Fatalf("read udpserver.go: %v", err)
	}
	code := stripGoLineComments(string(src))

	loc := regexp.MustCompile(`SetRecvReplayDedupe\(func\(ppd \*core\.PacketParserData\) error \{`).FindStringIndex(code)
	if loc == nil {
		t.Fatal("SetRecvReplayDedupe call not found; re-anchor this test rather than deleting it")
	}
	// The chain body: from the call to the closing of the func literal.
	end := strings.Index(code[loc[1]:], "\n\t})")
	if end < 0 {
		t.Fatal("cannot find the end of the dedupe chain")
	}
	chain := code[loc[1] : loc[1]+end]

	for _, want := range []string{"dedupeRecvKNK", "dedupeRecvART", "dedupeRecvRLY"} {
		if !strings.Contains(chain, want) {
			t.Errorf("%s is not chained into SetRecvReplayDedupe; it would never run", want)
		}
	}
}

// The cache has to be constructed, or the hook dereferences a nil cache
// and MarkSeen reports every knock as a duplicate.
func TestKnockReplayCacheIsConstructed(t *testing.T) {
	src, err := os.ReadFile("udpserver.go")
	if err != nil {
		t.Fatalf("read udpserver.go: %v", err)
	}
	if !regexp.MustCompile(`s\.knockReplay\s*=\s*newKnockReplayCache\(\)`).MatchString(stripGoLineComments(string(src))) {
		t.Error("s.knockReplay is never assigned; the dedupe would run against a nil cache")
	}
}

func stripGoLineComments(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}
