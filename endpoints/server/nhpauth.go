package server

import (
	"encoding/base64"
	"encoding/json"

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
		// Derive the outcome from err alone, matching the "succeed"/"failed"
		// branch logged above, so the trail never disagrees with the
		// operational log. The plugin's raw ErrCode is recorded as its own
		// field regardless, so a soft denial (nil error, non-"0" code) is
		// still visible instead of silently flipping the recorded result.
		granted := err == nil
		severity, result := audit.SeverityWarn, "denied"
		if granted {
			severity, result = audit.SeverityInfo, "granted"
		}
		fields := map[string]string{
			"user":   knkMsg.UserId,
			"device": knkMsg.DeviceId,
			"src":    addrStr,
			"aspId":  knkMsg.AuthServiceId,
			"resId":  knkMsg.ResourceId,
			// op distinguishes an open from a close: NHP_KNK, NHP_RKN and
			// NHP_EXT all dispatch through this handler, and NHP_EXT closes
			// access. Without it, a close reads as a granted knock. HeaderType
			// is the AEAD-authenticated body value validated against the wire
			// type by verifyKnockHeaderType above.
			"op":      core.HeaderTypeToString(knkMsg.HeaderType),
			"via":     "udp",
			"peerKey": shortKey(base64.StdEncoding.EncodeToString(ppd.RemotePubKey)),
			"result":  result,
		}
		if ackMsg != nil {
			fields["errCode"] = ackMsg.ErrCode
		}
		if err != nil {
			fields["reason"] = err.Error()
		}
		s.auditEvent("knock", severity, fields)
	}

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

	transaction.NextMsgCh <- ackMd
	return nil
}
