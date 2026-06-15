package server

import (
	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// Knock HeaderType verification (on-path header-flip protection).
//
// The server decides whether an incoming knock is an open (NHP_KNK /
// NHP_RKN) or a close (NHP_EXT) from the packet's WIRE HeaderType. That
// byte is not covered by the noise chain-hash / AEAD — only the message
// body is — and the trailing header digest is an unkeyed hash over
// public values. An on-path attacker can therefore flip a knock's wire
// HeaderType from NHP_KNK to NHP_EXT, recompute that digest, and make
// the server treat a legitimate open as a close: HandleKnockRequest ->
// processACOperation runs with openTime=1 (udpserver.go), revoking the
// victim's access. No key material is needed and the cryptography is
// never broken.
//
// Defense: the agent mirrors the wire HeaderType inside the
// AEAD-authenticated knock body (AgentKnockMsg.HeaderType). Because the
// body sits inside the chain-hash-authenticated payload, an attacker
// cannot flip it without the initiator's static private key. The server
// requires the body and wire HeaderType to agree and then uses the
// authenticated body value for the open/close decision.
//
// Verification is unconditional — secure by default, no compatibility
// switch. A knock whose body HeaderType is missing (agent predates this
// field) or disagrees with the wire is rejected.

// verifyKnockHeaderType compares the AEAD-authenticated body HeaderType
// against the unauthenticated wire HeaderType. It returns nil when they
// agree (the caller then trusts the body value, which equals the wire
// value), or the reject error otherwise — logging the rejection:
//
//   - bodyType is the zero value (NHP_KPL): the agent does not populate
//     the field and must be upgraded -> ErrKnockHeaderTypeLegacy.
//   - bodyType is populated but disagrees with wireType: the wire header
//     was tampered with on path -> ErrKnockHeaderTypeMismatch.
//
// This zero-value test relies on NHP_KPL being the iota-zero sentinel
// (nhp/core/packet.go) that never appears as a routed knock type on either
// side: the device dispatcher handles keepalives elsewhere, so no routed
// knock carries a NHP_KPL wire type, and an upgraded agent always sets a
// non-zero body type (NHP_KNK/NHP_RKN/NHP_EXT). A zero body therefore
// unambiguously means "field absent" (legacy agent), not a legitimately
// sent type whose value happens to be 0.
func verifyKnockHeaderType(bodyType, wireType int, transactionId uint64, addrStr string) *common.Error {
	if bodyType == core.NHP_KPL {
		log.Warning("server-agent(#%d@%s)[HeaderType] knock rejected: body HeaderType missing, upgrade agent; wire=%s",
			transactionId, addrStr, core.HeaderTypeToString(wireType))
		return common.ErrKnockHeaderTypeLegacy
	}
	if bodyType != wireType {
		log.Warning("server-agent(#%d@%s)[HeaderType] knock rejected: body/wire HeaderType mismatch, possible on-path tampering; body=%s wire=%s",
			transactionId, addrStr, core.HeaderTypeToString(bodyType), core.HeaderTypeToString(wireType))
		return common.ErrKnockHeaderTypeMismatch
	}
	return nil
}
