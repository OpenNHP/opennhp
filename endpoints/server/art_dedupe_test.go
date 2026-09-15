package server

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestDedupeRecvART(t *testing.T) {
	s := &UdpServer{artReplay: newARTReplayCache(), acPeerMap: make(map[string]*core.UdpPeer)}

	if err := s.dedupeRecvART(&core.PacketParserData{HeaderType: core.NHP_AOP}); err != nil {
		t.Fatalf("non-ART packet rejected: %v", err)
	}
	if err := s.dedupeRecvART(&core.PacketParserData{HeaderType: core.NHP_ART, SenderTrxId: 1}); !errors.Is(err, common.ErrServerMissingPeerPubkey) {
		t.Fatalf("missing pubkey error = %v", err)
	}

	for _, size := range []int{core.PublicKeySize, core.PublicKeySizeEx} {
		packet := &core.PacketParserData{
			HeaderType:     core.NHP_ART,
			SenderTrxId:    uint64(size),
			RemotePubKey:   artTestPubkey(1, size),
			RemoteSendTime: 123,
		}
		s.acPeerMap[base64.StdEncoding.EncodeToString(packet.RemotePubKey)] = &core.UdpPeer{Type: core.NHP_AC}
		if err := s.dedupeRecvART(packet); err != nil {
			t.Fatalf("first ART with %d-byte key rejected: %v", size, err)
		}
		if err := s.dedupeRecvART(packet); !errors.Is(err, common.ErrServerDuplicateTransaction) {
			t.Fatalf("duplicate ART with %d-byte key error = %v", size, err)
		}
	}
}

func TestARTRejectsNonACBeforeCaching(t *testing.T) {
	key := artTestPubkey(9, core.PublicKeySize)
	for _, role := range []int{core.NHP_AGENT, core.NHP_DB, core.NHP_RELAY} {
		s := &UdpServer{artReplay: newARTReplayCache(), acPeerMap: map[string]*core.UdpPeer{base64.StdEncoding.EncodeToString(key): {Type: role}}}
		p := &core.PacketParserData{HeaderType: core.NHP_ART, RemotePubKey: key, SenderTrxId: 1, RemoteSendTime: 1}
		if err := s.dedupeRecvART(p); err != common.ErrServerMissingPeerPubkey {
			t.Fatalf("role %d accepted: %v", role, err)
		}
		if len(s.artReplay.entries) != 0 {
			t.Fatal("non-AC populated cache")
		}
	}
}
