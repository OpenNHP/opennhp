package agent

import (
	"encoding/json"
	"net"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// note: code in request.go is for nhp-agent to send request to nhp-server.

func (a *UdpAgent) RequestOtp(target *KnockTarget) error {
	a.knockUserMutex.RLock()
	otpMsg := &common.AgentOTPMsg{
		UserId:         a.knockUser.UserId,
		DeviceId:       a.deviceId,
		OrganizationId: a.knockUser.OrganizationId,
		AuthServiceId:  target.AuthServiceId,
		UserData:       a.knockUser.UserData,
	}
	a.knockUserMutex.RUnlock()
	otpBytes, _ := json.Marshal(otpMsg)

	server := target.ServerPeer
	if server == nil {
		log.Critical("agent(%s)[RequestOtp] server is not assigned", otpMsg.UserId)
		return common.ErrKnockServerNotFound
	}

	server, sendAddr := a.ResolvePeer(server)
	if sendAddr == nil {
		log.Critical("agent(%s)[RequestOtp] server IP cannot be parsed", otpMsg.UserId)
		return common.ErrKnockServerNotFound
	}

	otpMd := &core.MsgData{
		RemoteAddr:    sendAddr.(*net.UDPAddr),
		HeaderType:    core.NHP_OTP,
		CipherScheme:  a.config.DefaultCipherScheme,
		TransactionId: a.device.NextCounterIndex(),
		Compress:      true,
		Message:       otpBytes,
		PeerPk:        server.PublicKey(),
	}

	if !a.IsRunning() {
		log.Error("agent(%s#%d)[RequestOtp] MsgData channel closed or being closed, skip sending", otpMsg.UserId, otpMd.TransactionId)
		return common.ErrPacketToMessageRoutineStopped
	}

	// device will create or find existing connection and sends the MsgAssembler via that connection
	a.sendMsgCh <- otpMd

	log.Info("agent(%s#%d)[RequestOtp] sending otp request", otpMsg.UserId, otpMd.TransactionId)
	return nil
}

func (a *UdpAgent) RegisterPublicKey(otp string, target *KnockTarget) (rakMsg *common.ServerRegisterAckMsg, err error) {
	a.knockUserMutex.RLock()
	regMsg := &common.AgentRegisterMsg{
		UserId:         a.knockUser.UserId,
		DeviceId:       a.deviceId,
		OrganizationId: a.knockUser.OrganizationId,
		AuthServiceId:  target.AuthServiceId,
		OTP:            otp,
		UserData:       a.knockUser.UserData,
	}
	a.knockUserMutex.RUnlock()
	regBytes, _ := json.Marshal(regMsg)

	server := target.ServerPeer
	if server == nil {
		log.Critical("agent(%s)[RegisterPublicKey] server is not assigned", regMsg.UserId)
		return nil, common.ErrKnockServerNotFound
	}

	server, sendAddr := a.ResolvePeer(server)
	if sendAddr == nil {
		log.Critical("agent(%s)[RegisterPublicKey] server IP cannot be parsed", regMsg.UserId)
		return nil, common.ErrKnockServerNotFound
	}
	addrStr := sendAddr.String()

	regMd := &core.MsgData{
		RemoteAddr:    sendAddr.(*net.UDPAddr),
		HeaderType:    core.NHP_REG,
		CipherScheme:  a.config.DefaultCipherScheme,
		TransactionId: a.device.NextCounterIndex(),
		Compress:      true,
		Message:       regBytes,
		PeerPk:        server.PublicKey(),
		ResponseMsgCh: make(chan *core.PacketParserData),
	}

	if !a.IsRunning() {
		log.Error("agent(%s#%d)[RegisterPublicKey] MsgData channel closed or being closed, skip sending", regMsg.UserId, regMd.TransactionId)
		return nil, common.ErrPacketToMessageRoutineStopped
	}

	// device will create or find existing connection and sends the MsgAssembler via that connection
	a.sendMsgCh <- regMd

	// block until transaction completes
	serverPpd := <-regMd.ResponseMsgCh
	close(regMd.ResponseMsgCh)

	if serverPpd.Error != nil {
		log.Error("agent(%s#%d)[RegisterPublicKey] failed to receive response from server %s: %v", regMsg.UserId, regMd.TransactionId, addrStr, serverPpd.Error)
		err = serverPpd.Error
		return nil, err
	}

	if serverPpd.HeaderType != core.NHP_RAK {
		log.Error("agent(%s#%d)[RegisterPublicKey] response has wrong type: %s", regMsg.UserId, regMd.TransactionId, core.HeaderTypeToString(serverPpd.HeaderType))
		err = common.ErrTransactionRepliedWithWrongType
		return nil, err
	}

	rakMsg = &common.ServerRegisterAckMsg{}
	err = json.Unmarshal(serverPpd.BodyMessage, rakMsg)
	if err != nil {
		log.Error("agent(%s#%d)[RegisterPublicKey] failed to parse %s message: %v", regMsg.UserId, regMd.TransactionId, core.HeaderTypeToString(serverPpd.HeaderType), err)
		return nil, err
	}

	if rakMsg.ErrCode != common.ErrSuccess.ErrorCode() {
		log.Error("agent(%s#%d)[RegisterPublicKey] response error: %s", regMsg.UserId, regMd.TransactionId, rakMsg.ErrMsg)
		err = common.ErrorCodeToError(rakMsg.ErrCode)
		return rakMsg, err
	}

	log.Info("agent(%s#%d)[RegisterPublicKey] succeed", regMsg.UserId, regMd.TransactionId)
	return rakMsg, nil
}

func (a *UdpAgent) ListResource(target *KnockTarget) (lrtMsg *common.ServerListResultMsg, err error) {
	a.knockUserMutex.RLock()
	lstMsg := &common.AgentListMsg{
		UserId:         a.knockUser.UserId,
		DeviceId:       a.deviceId,
		OrganizationId: a.knockUser.OrganizationId,
		AuthServiceId:  target.AuthServiceId,
		UserData:       a.knockUser.UserData,
	}
	a.knockUserMutex.RUnlock()
	lstBytes, _ := json.Marshal(lstMsg)

	server := target.ServerPeer
	if server == nil {
		log.Critical("agent(%s)[ListResource] server is not assigned", lstMsg.UserId)
		return nil, common.ErrKnockServerNotFound
	}

	server, sendAddr := a.ResolvePeer(server)
	if sendAddr == nil {
		log.Critical("agent(%s)[ListResource] server IP cannot be parsed", lstMsg.UserId)
		return nil, common.ErrKnockServerNotFound
	}
	addrStr := sendAddr.String()

	lstMd := &core.MsgData{
		RemoteAddr:    sendAddr.(*net.UDPAddr),
		HeaderType:    core.NHP_LST,
		CipherScheme:  a.config.DefaultCipherScheme,
		TransactionId: a.device.NextCounterIndex(),
		Compress:      true,
		Message:       lstBytes,
		PeerPk:        server.PublicKey(),
		ResponseMsgCh: make(chan *core.PacketParserData),
	}

	if !a.IsRunning() {
		log.Error("agent(%s#%d)[ListResource] MsgData channel closed or being closed, skip sending", lstMsg.UserId, lstMd.TransactionId)
		return nil, common.ErrPacketToMessageRoutineStopped
	}

	// device will create or find existing connection and sends the MsgAssembler via that connection
	a.sendMsgCh <- lstMd

	// block until transaction completes
	serverPpd := <-lstMd.ResponseMsgCh
	close(lstMd.ResponseMsgCh)

	if serverPpd.Error != nil {
		log.Error("agent(%s#%d)[ListResource] failed to receive response from server %s: %v", lstMsg.UserId, lstMd.TransactionId, addrStr, serverPpd.Error)
		err = serverPpd.Error
		return nil, err
	}

	if serverPpd.HeaderType != core.NHP_LRT {
		log.Error("agent(%s#%d)[ListResource] response has wrong type: %s", lstMsg.UserId, lstMd.TransactionId, core.HeaderTypeToString(serverPpd.HeaderType))
		err = common.ErrTransactionRepliedWithWrongType
		return nil, err
	}

	lrtMsg = &common.ServerListResultMsg{}
	err = json.Unmarshal(serverPpd.BodyMessage, lrtMsg)
	if err != nil {
		log.Error("agent(%s#%d)[ListResource] failed to parse %s message: %v", lstMsg.UserId, lstMd.TransactionId, core.HeaderTypeToString(serverPpd.HeaderType), err)
		return nil, err
	}

	if lrtMsg.ErrCode != common.ErrSuccess.ErrorCode() {
		log.Error("agent(%s#%d)[ListResource] list response error: %s", lstMsg.UserId, lstMd.TransactionId, lrtMsg.ErrMsg)
		err = common.ErrorCodeToError(lrtMsg.ErrCode)
		return lrtMsg, err
	}

	log.Info("agent(%s#%d)[ListResource] succeed", lstMsg.UserId, lstMd.TransactionId)
	return lrtMsg, nil
}
