package server

import (
	"encoding/base64"
	"encoding/json"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/core"
	"github.com/OpenNHP/opennhp/log"
)

// HandleKnockRequest
// Server will respond with success or error with NHP_ACK message
func (s *UdpServer) HandleKnockRequest(ppd *core.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderTrxId
	addrStr := ppd.ConnData.RemoteAddr.String()
	knkMsg := &common.AgentKnockMsg{}
	ackMsg := &common.ServerKnockAckMsg{
		AgentAddr: addrStr, // optional, to tell agent its own outwards ip address
	}

	func() {
		// parse knockMsg
		err = json.Unmarshal(ppd.BodyMessage, knkMsg)
		if err != nil {
			log.Error("server-agent(#%d@%s)[HandleKnockRequest] failed to parse %s message: %v", transactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType), err)
			ackMsg.ErrCode = common.ErrJsonParseFailed.ErrorCode()
			ackMsg.ErrMsg = err.Error()
			return
		}
		// determine knock type
		knkMsg.HeaderType = ppd.HeaderType

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
				Ip:   ppd.ConnData.RemoteAddr.IP.String(),
				Port: ppd.ConnData.RemoteAddr.Port,
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

	// send back knock ack response
	ackBytes, _ := json.Marshal(ackMsg)

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
