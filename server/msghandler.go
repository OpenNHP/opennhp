package server

import (
	"encoding/base64"
	"encoding/json"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/log"
	"github.com/OpenNHP/opennhp/nhp"
)

// HandleOTPRequest
// Server will not respond to agent's otp request
func (s *UdpServer) HandleOTPRequest(ppd *nhp.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderId
	addrStr := ppd.ConnData.RemoteAddr.String()

	otpMsg := &common.AgentOTPMsg{}
	err = json.Unmarshal(ppd.BodyMessage, otpMsg)
	if err != nil {
		log.Error("server-agent(#%d@%s)[HandleOTPRequest] failed to parse %s message: %v", transactionId, addrStr, nhp.HeaderTypeToString(ppd.HeaderType), err)
		return err
	}

	handler := s.FindPluginHandler(otpMsg.AuthServiceId)
	if handler == nil {
		return common.ErrAuthHandlerNotFound
	}

	otpReq := &common.NhpOTPRequest{
		Msg: otpMsg,
		SrcAddr: &common.NetAddress{
			Ip:   ppd.ConnData.RemoteAddr.IP.String(),
			Port: ppd.ConnData.RemoteAddr.Port,
		},
	}

	err = handler.RequestOTP(otpReq, s.NewNhpServerHelper(ppd))
	if err != nil {
		log.Error("server-agent(%s#%d@%s)[HandleOTPRequest] error: %v", otpMsg.UserId, transactionId, addrStr, err)
		return err
	}

	log.Info("server-agent(%s#%d@%s)[HandleOTPRequest] succeeded", otpMsg.UserId, transactionId, addrStr)
	return nil
}

// HandleRegisterRequest
// Server will respond with success or error with NHP_RAK message
func (s *UdpServer) HandleRegisterRequest(ppd *nhp.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderId
	addrStr := ppd.ConnData.RemoteAddr.String()
	regMsg := &common.AgentRegisterMsg{}
	rakMsg := &common.ServerRegisterAckMsg{}

	func() {
		err = json.Unmarshal(ppd.BodyMessage, regMsg)
		if err != nil {
			log.Error("server-agent(#%d@%s)[HandleRegisterRequest] failed to parse %s message: %v", transactionId, addrStr, nhp.HeaderTypeToString(ppd.HeaderType), err)
			rakMsg.ErrCode = common.ErrJsonParseFailed.ErrorCode()
			rakMsg.ErrMsg = err.Error()
			return
		}

		handler := s.FindPluginHandler(regMsg.AuthServiceId)
		if handler == nil {
			err = common.ErrAuthHandlerNotFound
			rakMsg.ErrCode = common.ErrAuthHandlerNotFound.ErrorCode()
			rakMsg.ErrMsg = err.Error()
			return
		}

		agentPubkey := base64.StdEncoding.EncodeToString(ppd.RemotePubKey)

		regReq := &common.NhpRegisterRequest{
			Msg:       regMsg,
			Ack:       rakMsg,
			PublicKey: agentPubkey,
			SrcAddr: &common.NetAddress{
				Ip:   ppd.ConnData.RemoteAddr.IP.String(),
				Port: ppd.ConnData.RemoteAddr.Port,
			},
		}

		rakMsg, err = handler.RegisterAgent(regReq, s.NewNhpServerHelper(ppd))
		if err != nil {
			log.Error("server-agent(%s#%d@%s)[HandleRegisterRequest] error: %v", regMsg.UserId, transactionId, addrStr, err)
			return
		}

		log.Info("server-agent(%s#%d@%s)[HandleRegisterRequest] succeeded", regMsg.UserId, transactionId, addrStr)
	}()

	// send NHP_RAK message
	rakBytes, _ := json.Marshal(rakMsg)
	rakMd := &nhp.MsgData{
		HeaderType:     nhp.NHP_RAK,
		TransactionId:  transactionId,
		Compress:       true,
		PrevParserData: ppd,
		Message:        rakBytes,
	}

	// forward to a specific transaction
	transaction := ppd.ConnData.FindRemoteTransaction(transactionId)
	if transaction == nil {
		log.Error("server-agent(%s#%d@%s)[HandleRegisterRequest] transaction is not available", regMsg.UserId, transactionId, addrStr)
		err = common.ErrTransactionIdNotFound
		return err
	}

	transaction.NextMsgCh <- rakMd

	return err
}

// HandleListRequest
// Server will respond with success or error with NHP_LRT message
func (s *UdpServer) HandleListRequest(ppd *nhp.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderId
	addrStr := ppd.ConnData.RemoteAddr.String()
	lstMsg := &common.AgentListMsg{}
	lrtMsg := &common.ServerListResultMsg{}

	func() {
		err = json.Unmarshal(ppd.BodyMessage, lstMsg)
		if err != nil {
			log.Error("server-agent(#%d@%s)[HandleListRequest] failed to parse %s message: %v", transactionId, addrStr, nhp.HeaderTypeToString(ppd.HeaderType), err)
			lrtMsg.ErrCode = common.ErrJsonParseFailed.ErrorCode()
			lrtMsg.ErrMsg = err.Error()
			return
		}

		handler := s.FindPluginHandler(lstMsg.AuthServiceId)
		if handler == nil {
			err = common.ErrAuthHandlerNotFound
			lrtMsg.ErrCode = common.ErrAuthHandlerNotFound.ErrorCode()
			lrtMsg.ErrMsg = err.Error()
			return
		}

		agentPubkey := base64.StdEncoding.EncodeToString(ppd.RemotePubKey)
		listReq := &common.NhpListRequest{
			Msg:       lstMsg,
			Ack:       lrtMsg,
			PublicKey: agentPubkey,
			SrcAddr: &common.NetAddress{
				Ip:   ppd.ConnData.RemoteAddr.IP.String(),
				Port: ppd.ConnData.RemoteAddr.Port,
			},
		}

		lrtMsg, err = handler.ListService(listReq, s.NewNhpServerHelper(ppd))
		if err != nil {
			log.Error("server-agent(%s#%d@%s)[HandleListRequest] error: %v", lstMsg.UserId, transactionId, addrStr, err)
			return
		}

		log.Info("server-agent(%s#%d@%s)[HandleListRequest] succeeded", lstMsg.UserId, transactionId, addrStr)
	}()

	lrtBytes, _ := json.Marshal(lrtMsg)
	ackMd := &nhp.MsgData{
		HeaderType:     nhp.NHP_LRT,
		TransactionId:  transactionId,
		Compress:       true,
		PrevParserData: ppd,
		Message:        lrtBytes,
	}

	// forward to a specific transaction
	transaction := ppd.ConnData.FindRemoteTransaction(transactionId)
	if transaction == nil {
		log.Error("server-agent(%s#%d@%s)[HandleListRequest] transaction is not available", lstMsg.UserId, transactionId, addrStr)
		err = common.ErrTransactionIdNotFound
		return err
	}

	transaction.NextMsgCh <- ackMd

	return err
}

func (s *UdpServer) HandleACOnline(ppd *nhp.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderId
	addrStr := ppd.ConnData.RemoteAddr.String()
	aolMsg := &common.ACOnlineMsg{}

	err = json.Unmarshal(ppd.BodyMessage, aolMsg)
	if err != nil {
		log.Error("server-ac(#%d@%s)[HandleACOnline] failed to parse %s message: %v", transactionId, addrStr, nhp.HeaderTypeToString(ppd.HeaderType), err)
		return err
	}

	acId := aolMsg.ACId
	acPubkeyBase64 := base64.StdEncoding.EncodeToString(ppd.RemotePubKey)
	s.acPeerMapMutex.Lock()
	acPeer := s.acPeerMap[acPubkeyBase64] // door peer's recvAddr has already been updated by nhp packet parser
	s.acPeerMapMutex.Unlock()

	acConn := &ACConn{
		ConnData:  ppd.ConnData,
		ACPeer:    acPeer,
		ACId:      acId,
		ServiceId: aolMsg.AuthServiceId,
		Apps:      aolMsg.ResourceIds,
	}

	s.acConnectionMapMutex.Lock()
	s.acConnectionMap[acId] = acConn
	s.acConnectionMapMutex.Unlock()

	aakMsg := &common.ServerACAckMsg{
		ErrCode: common.ErrSuccess.ErrorCode(),
		ACAddr:  ppd.ConnData.RemoteAddr.String(),
	}
	aakBytes, _ := json.Marshal(aakMsg)

	aakMd := &nhp.MsgData{
		HeaderType:     nhp.NHP_AAK,
		TransactionId:  transactionId,
		Compress:       true,
		PrevParserData: ppd,
		Message:        aakBytes,
	}

	// forward to a specific transaction
	transaction := ppd.ConnData.FindRemoteTransaction(transactionId)
	if transaction == nil {
		log.Error("server-ac(@%s#%d@%s)[HandleACOnline] transaction is not available", acId, transactionId, addrStr)
		err = common.ErrTransactionIdNotFound
		return err
	}

	transaction.NextMsgCh <- aakMd

	return nil
}
