package server

import (
	"errors"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestDedupeRecvART(t *testing.T) {
	s := &UdpServer{artReplay: newARTReplayCache()}
	key := artTestPubkey(1)

	if err := s.dedupeRecvART(&core.PacketParserData{HeaderType: core.NHP_AOP}); err != nil {
		t.Fatalf("non-ART packet rejected: %v", err)
	}
	if err := s.dedupeRecvART(&core.PacketParserData{HeaderType: core.NHP_ART, SenderTrxId: 1}); !errors.Is(err, common.ErrServerMissingPeerPubkey) {
		t.Fatalf("missing pubkey error = %v", err)
	}

	packet := &core.PacketParserData{
		HeaderType:     core.NHP_ART,
		SenderTrxId:    7,
		RemotePubKey:   key,
		RemoteSendTime: 123,
	}
	if err := s.dedupeRecvART(packet); err != nil {
		t.Fatalf("first ART rejected: %v", err)
	}
	if err := s.dedupeRecvART(packet); !errors.Is(err, common.ErrServerDuplicateTransaction) {
		t.Fatalf("duplicate ART error = %v", err)
	}
}
