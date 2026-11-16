package server

import (
	"encoding/base64"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/nhp/audit"
	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// HandleKnockRequest
// Server will respond with success or error with NHP_ACK message
func (s *UdpServer) HandleKnockRequest(ppd *core.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderTrxId
	// For relay-forwarded connections, use the real client address
	// (the browser behind the relay) for auth and logging.
	clientAddr := ppd.ConnData.RemoteAddr
	if ppd.ConnData.RealRemoteAddr != nil {
		clientAddr = ppd.ConnData.RealRemoteAddr
	}
	addrStr := clientAddr.String()
	knkMsg := &common.AgentKnockMsg{}
	dhpKnkMsg := &common.DHPKnockMsg{}
	ackMsg := &common.ServerKnockAckMsg{
		AgentAddr: addrStr, // optional, to tell agent its own outwards ip address
	}
	dhpAckMsg := &common.ServerDHPKnockAckMsg{
		OpenTime: 30, // currently, use fixed value, unit is seconds.
	}

	func() {
		// parse knockMsg
		if ppd.HeaderType == core.DHP_KNK { // dhp knock
			err = json.Unmarshal(ppd.BodyMessage, dhpKnkMsg)
		} else {
			err = json.Unmarshal(ppd.BodyMessage, knkMsg)
		}

		if err != nil {
			log.Error("server-agent(#%d@%s)[HandleKnockRequest] failed to parse %s message: %v", transactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType), err)
			ackMsg.ErrCode = common.ErrJsonParseFailed.ErrorCode()
			ackMsg.ErrMsg = err.Error()
			return
		}

		// dhp knock
		if ppd.HeaderType == core.DHP_KNK {
			// Replay protection applies to the DHP knock as well: it is
			// dispatched through the same handler and its outcome is the
			// same evidence appraisal on the same server.
			if replayErr := s.dedupeRecvKNK(ppd, addrStr); replayErr != nil {
				err = replayErr
				dhpAckMsg.ErrCode = replayErr.ErrorCode()
				dhpAckMsg.ErrMsg = replayErr.Error()
				return
			}
			log.Info("server-agent(%s#%d@%s)[HandleKnockRequest] start to verify evidence for dhp knock", knkMsg.UserId, transactionId, addrStr)
			if s.AppraiseEvidence(dhpKnkMsg.Evidence) {
				dhpAckMsg.ErrCode = common.ErrSuccess.ErrorCode()
			} else {
				dhpAckMsg.ErrCode = common.ErrEvidenceAppraisalFailed.ErrorCode()
			}
			return
		}

		// The wire HeaderType (ppd.HeaderType) is NOT authenticated, so
		// trusting it would let an on-path attacker flip NHP_KNK <-> NHP_EXT
		// and forge the open/close decision. Require it to match the
		// AEAD-authenticated body HeaderType the agent carries; on success
		// knkMsg.HeaderType already holds that (equal) authenticated value.
		// Rejected unconditionally on a mismatch or a legacy unpopulated
		// body — secure by default. See knock_headertype.go.
		if gateErr := verifyKnockHeaderType(knkMsg.HeaderType, ppd.HeaderType, transactionId, addrStr); gateErr != nil {
			err = gateErr
			ackMsg.ErrCode = gateErr.ErrorCode()
			ackMsg.ErrMsg = gateErr.Error()
			return
		}

		// Replay protection. Placed AFTER the header-type gate so that a
		// packet rejected for an on-path type flip does not consume the
		// agent's cache entry, and BEFORE any auth-service lookup, plugin
		// call or AC operation so a replayed knock triggers neither an
		// authentication attempt nor a firewall grant.
		if replayErr := s.dedupeRecvKNK(ppd, addrStr); replayErr != nil {
			err = replayErr
			ackMsg.ErrCode = replayErr.ErrorCode()
			ackMsg.ErrMsg = replayErr.Error()
			return
		}

		// find out auth service provider
		aspData := s.FindAuthSvcProvider(knkMsg.AuthServiceId)
		if aspData == nil {
			err = common.ErrAuthServiceProviderNotFound
			ackMsg.ErrCode = common.ErrAuthServiceProviderNotFound.ErrorCode()
			ackMsg.ErrMsg = err.Error()
			return
		}

		// find out auth plugin handler
		handler := s.FindPluginHandler(knkMsg.AuthServiceId)
		if handler == nil {
			log.Error("server-agent(%s#%d@%s)[HandleKnockRequest-Auth] failed to find service provider with %s", knkMsg.UserId, transactionId, addrStr, knkMsg.AuthServiceId)
			err = common.ErrAuthServiceProviderNotFound
			ackMsg.ErrCode = common.ErrAuthServiceProviderNotFound.ErrorCode()
			ackMsg.ErrMsg = err.Error()
			return
		}

		authReq := &common.NhpAuthRequest{
			Msg:       knkMsg,
			Ack:       ackMsg,
			PublicKey: base64.StdEncoding.EncodeToString(ppd.RemotePubKey),
			SrcAddr: &common.NetAddress{
				Ip:   clientAddr.IP.String(),
				Port: clientAddr.Port,
			},
		}

		// perform knock auth and open ip rule from the agent src address and resource dst address
		ackMsg, err = handler.AuthWithNHP(authReq, s.NewNhpServerHelper(ppd))
		if err != nil {
			log.Info("server-agent(%s#%d@%s)[HandleKnockRequest] failed: %+v", knkMsg.UserId, transactionId, addrStr, err)
			return
		}

		log.Info("server-agent(%s#%d@%s)[HandleKnockRequest] succeed: %+v", knkMsg.UserId, transactionId, addrStr)
	}()

	// Record the access decision in the tamper-evident audit ledger. Only
	// the NHP knock path is a per-resource access decision; the DHP path
	// (evidence appraisal) is audited elsewhere. The ledger nil-guards, so
	// this whole block is skipped cheaply when auditing is off.
	//
	// This deliberately records the DECISION, not its delivery: it is
	// emitted before the ACK is handed to the transaction, so a later send
	// failure still leaves a "granted" entry. That is the intent — the
	// authorization outcome is what an audit trail must capture, and the
	// agent simply retries when an ACK is lost.
	if s.auditLedger != nil && ppd.HeaderType != core.DHP_KNK {
		// Grant means a non-nil ack, a nil error, AND the explicit success
		// code: a plugin may return a failure ErrCode with a nil error (a
		// soft denial), a recovered plugin panic returns a nil ack with a
		// nil error too, and a plugin that sets no code at all did not
		// affirmatively report success — any of those recorded as "granted"
		// would hide it from a SIEM rule keyed on result. The raw code is
		// kept in errCode regardless. See decisionGranted.
		ackCode := ""
		if ackMsg != nil {
			ackCode = ackMsg.ErrCode
		}
		severity, result := audit.SeverityWarn, "denied"
		if decisionGranted(err, ackMsg == nil, ackCode) {
			severity, result = audit.SeverityInfo, "granted"
		}
		fields := map[string]string{
			"user":    knkMsg.UserId,
			"device":  knkMsg.DeviceId,
			"src":     addrStr,
			"aspId":   knkMsg.AuthServiceId,
			"resId":   knkMsg.ResourceId,
			"op":      auditKnockOp(knkMsg.HeaderType),
			"via":     "udp",
			"peerKey": shortKey(base64.StdEncoding.EncodeToString(ppd.RemotePubKey)),
			"result":  result,
		}
		if ackCode != "" {
			fields["errCode"] = ackCode
		}
		if err != nil {
			fields["reason"] = err.Error()
		}
		s.auditEvent("knock", severity, fields)
	}

	// Record the auth outcome. Reads the ack defensively (see knockAuthorized):
	// AuthWithNHP can legitimately return a nil ackMsg, and HandleKnockRequest
	// runs in a handler goroutine with no recover(), so a nil dereference here
	// would take down the whole server.
	s.metrics.recordKnockAuth(knockAuthorized(ppd.HeaderType, ackMsg, dhpAckMsg))

	// send back knock ack response
	ackBytes, _ := json.Marshal(ackMsg)

	// DHP knock
	if ppd.HeaderType == core.DHP_KNK {
		ackBytes, _ = json.Marshal(dhpAckMsg)
	}

	ackMd := &core.MsgData{
		HeaderType:     core.NHP_ACK,
		TransactionId:  transactionId, // transactionId of the original knock request
		Compress:       true,
		PrevParserData: ppd,
		Message:        ackBytes,
	}

	// forward to a specific transaction
	transaction := ppd.ConnData.FindRemoteTransaction(transactionId)
	if transaction == nil {
		log.Error("server-agent(%s#%d@%s)[HandleKnockRequest] transaction is not available", knkMsg.UserId, transactionId, addrStr)
		err = common.ErrTransactionIdNotFound
		return err
	}

	if sendErr := transaction.SendMessage(ackMd); sendErr != nil {
		log.Error("server-agent(%s#%d@%s)[HandleKnockRequest] transaction closed before forward: %v", knkMsg.UserId, transactionId, addrStr, sendErr)
		return sendErr
	}
	return nil
}

// lastKNKReplayWarn throttles the replay warning to once per minute, so a
// controlled replay flood cannot dominate the log. The drop counter below
// still records every refused packet, so the rate stays observable.
var lastKNKReplayWarn atomic.Int64

// dedupeRecvKNK rejects a byte-identical replay of an authenticated knock.
//
// A knock is the one message whose replay has a security consequence: the
// server derives the AC firewall grant from the packet's SOURCE ADDRESS, so
// a captured knock replayed from a different address gets that address
// admitted. The per-connection monotonic gate in nhp/core/responder.go does
// not stop that — a replay from a new source address lands on a fresh
// ConnectionData whose LastRemoteSendTime is 0.
//
// Why this is NOT installed through the SetRecvReplayDedupe hook chain:
// that hook runs after validatePeer but BEFORE decryptBody, and a knock's
// peer is any registered agent whose public key is public. A forged packet
// with a valid header but a garbage body would claim a real agent's
// (pubkey, txid, sendTime) triple there and poison the cache, so that
// agent's next genuine knock would be rejected as a replay — a
// pre-authentication cache-poisoning DoS. Here the body AEAD has already
// been verified, so only a holder of the agent's private key can reach
// this function.
//
// A legitimate retransmission is never caught: the agent takes a fresh
// transaction id and a fresh timestamp for every packet it builds, so a
// retry produces a different triple. Only a byte-identical replay
// reproduces the triple.
//
// The error returned is ErrServerDuplicateTransaction, the same one the ART
// replay path returns.
func (s *UdpServer) dedupeRecvKNK(ppd *core.PacketParserData, addrStr string) *common.Error {
	if l := len(ppd.RemotePubKey); l != core.PublicKeySize && l != core.PublicKeySizeEx {
		log.Critical("server[dedupeRecvKNK] invalid peer pubkey length %d for txid=%d (want %d or %d)", l, ppd.SenderTrxId, core.PublicKeySize, core.PublicKeySizeEx)
		return common.ErrServerMissingPeerPubkey
	}
	if stage := s.knkReplay.MarkSeen(ppd.RemotePubKey, ppd.SenderTrxId, ppd.RemoteSendTime); stage != "" {
		s.metrics.recordDroppedPacket(stage)
		now := time.Now().UnixNano()
		last := lastKNKReplayWarn.Load()
		if now-last >= int64(time.Minute) && lastKNKReplayWarn.CompareAndSwap(last, now) {
			log.Warning("server[dedupeRecvKNK] dropped knock (stage=%s, txid=%d, pubkey=%s, sendTime=%d, src=%s)",
				stage, ppd.SenderTrxId, artPubkeyFingerprint(ppd.RemotePubKey), ppd.RemoteSendTime, addrStr)
		}
		return common.ErrServerDuplicateTransaction
	}
	return nil
}

// auditKnockOp names the knock operation for the audit trail. NHP_KNK,
// NHP_RKN and NHP_EXT all dispatch through HandleKnockRequest, and NHP_EXT
// closes access, so recording the type keeps a close from reading as a
// granted open. On the granted path knkMsg.HeaderType is the
// AEAD-authenticated body value (verifyKnockHeaderType confirmed it matches
// the wire type). On a denial it can be zero — a body that failed to
// unmarshal, or a legacy agent rejected by that same gate — and
// core.HeaderTypeToString(0) is NHP_KPL (keepalive), which would mislabel a
// whole class of rejected knocks as keepalives. An unset type is reported as
// "unknown" so the trail never claims a type the decision did not
// authenticate.
func auditKnockOp(headerType int) string {
	if headerType == core.NHP_KPL {
		return "unknown"
	}
	return core.HeaderTypeToString(headerType)
}

// knockAuthorized reports whether a knock ended in a successful authorization,
// for metrics. It tolerates a nil ack: AuthWithNHP may return (nil, ...) when a
// plugin doesn't implement it or panics and is recovered, and a nil deref in
// the caller's handler goroutine (no recover) would crash the server. A missing
// ack is treated as "not authorized". This covers the whole knock family
// (KNK/RKN/EXT/DHP_KNK), not just first-time KNK auth.
func knockAuthorized(headerType int, ackMsg *common.ServerKnockAckMsg, dhpAckMsg *common.ServerDHPKnockAckMsg) bool {
	switch {
	case headerType == core.DHP_KNK:
		return dhpAckMsg != nil && dhpAckMsg.ErrCode == common.ErrSuccess.ErrorCode()
	case ackMsg != nil:
		return ackMsg.ErrCode == common.ErrSuccess.ErrorCode()
	default:
		return false
	}
}
