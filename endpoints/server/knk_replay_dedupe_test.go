package server

import (
	"errors"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// TestDedupeRecvKNK covers the server-level entry point: the first knock
// is accepted, a byte-identical replay is refused with
// ErrServerDuplicateTransaction, and a knock with a different transaction
// id (what a legitimate retry produces) is accepted.
func TestDedupeRecvKNK(t *testing.T) {
	s := &UdpServer{knkReplay: newKNKReplayCache()}
	key := artTestPubkey(1)

	packet := &core.PacketParserData{
		HeaderType:     core.NHP_KNK,
		SenderTrxId:    42,
		RemotePubKey:   key,
		RemoteSendTime: 1000,
	}
	if err := s.dedupeRecvKNK(packet, "192.0.2.1:1234"); err != nil {
		t.Fatalf("first knock rejected: %v", err)
	}
	if err := s.dedupeRecvKNK(packet, "198.51.100.9:1234"); !errors.Is(err, common.ErrServerDuplicateTransaction) {
		t.Fatalf("replayed knock error = %v, want ErrServerDuplicateTransaction", err)
	}

	// A legitimate retry carries a fresh transaction id and timestamp.
	retry := &core.PacketParserData{
		HeaderType:     core.NHP_KNK,
		SenderTrxId:    43,
		RemotePubKey:   key,
		RemoteSendTime: 2000,
	}
	if err := s.dedupeRecvKNK(retry, "192.0.2.1:1234"); err != nil {
		t.Fatalf("legitimate retry rejected: %v", err)
	}
}

// TestDedupeRecvKNKRejectsInvalidPubKey ensures a packet with an
// unexpected public-key length fails closed and does not consume a cache
// slot.
func TestDedupeRecvKNKRejectsInvalidPubKey(t *testing.T) {
	s := &UdpServer{knkReplay: newKNKReplayCache()}
	packet := &core.PacketParserData{
		HeaderType:   core.NHP_KNK,
		SenderTrxId:  1,
		RemotePubKey: []byte{1, 2, 3},
	}
	if err := s.dedupeRecvKNK(packet, "192.0.2.1:1234"); !errors.Is(err, common.ErrServerMissingPeerPubkey) {
		t.Fatalf("invalid pubkey error = %v, want ErrServerMissingPeerPubkey", err)
	}
	if s.knkReplay.Len() != 0 {
		t.Fatal("invalid pubkey populated the cache")
	}
}

// TestDedupeRecvKNKCoversKnockFamily ensures the replay gate applies to
// every message type the knock handler dispatches.
func TestDedupeRecvKNKCoversKnockFamily(t *testing.T) {
	for _, headerType := range []int{core.NHP_KNK, core.NHP_RKN, core.NHP_EXT, core.DHP_KNK} {
		s := &UdpServer{knkReplay: newKNKReplayCache()}
		key := artTestPubkey(1)
		packet := &core.PacketParserData{
			HeaderType:     headerType,
			SenderTrxId:    5,
			RemotePubKey:   key,
			RemoteSendTime: 500,
		}
		if err := s.dedupeRecvKNK(packet, "192.0.2.1:1234"); err != nil {
			t.Fatalf("%s: first knock rejected: %v", core.HeaderTypeToString(headerType), err)
		}
		if err := s.dedupeRecvKNK(packet, "198.51.100.9:1234"); !errors.Is(err, common.ErrServerDuplicateTransaction) {
			t.Fatalf("%s: replay error = %v, want ErrServerDuplicateTransaction", core.HeaderTypeToString(headerType), err)
		}
	}
}

// TestDedupeRecvKNKRelayForwardedKnockIsBucketedOnce covers the relay
// path: an agent behind a relay has its inner KNK re-injected on a
// per-client connection, and that inner knock is bucketed exactly once.
// The outer NHP_RLY envelope has its own per-relay sliding window, so the
// two never share a key space and a legitimate forward is not rejected.
func TestDedupeRecvKNKRelayForwardedKnockIsBucketedOnce(t *testing.T) {
	s := &UdpServer{knkReplay: newKNKReplayCache()}
	agentKey := artTestPubkey(1)

	// The inner KNK: the agent's own static key, its own counter and its
	// own timestamp, forwarded by the relay.
	inner := &core.PacketParserData{
		HeaderType:     core.NHP_KNK,
		SenderTrxId:    77,
		RemotePubKey:   agentKey,
		RemoteSendTime: 7000,
	}
	if err := s.dedupeRecvKNK(inner, "203.0.113.7:5000"); err != nil {
		t.Fatalf("relay-forwarded knock rejected: %v", err)
	}
	// The same inner packet arriving again (a captured relay forward) is
	// refused.
	if err := s.dedupeRecvKNK(inner, "203.0.113.7:5000"); !errors.Is(err, common.ErrServerDuplicateTransaction) {
		t.Fatalf("replayed relay-forwarded knock error = %v, want ErrServerDuplicateTransaction", err)
	}
	// The relay's own envelope is a different message type with its own
	// counter and its own replay window, so it is unaffected.
	relayEnvelope := &core.PacketParserData{
		HeaderType:     core.NHP_RLY,
		SenderTrxId:    78,
		RemotePubKey:   artTestPubkey(2),
		RemoteSendTime: 7000,
	}
	if err := s.dedupeRecvKNK(relayEnvelope, "203.0.113.7:5000"); err != nil {
		t.Fatalf("relay envelope rejected: %v", err)
	}
}
