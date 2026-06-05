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

	serverPeer := target.GetServerPeer()
	if serverPeer == nil {
		log.Critical("agent(%s)[RequestOtp] server is not assigned", otpMsg.UserId)
		return common.ErrKnockServerNotFound
	}
	inst := target.PickInstance()
	if inst == nil {
		log.Critical("agent(%s)[RequestOtp] no instance available", otpMsg.UserId)
		return common.ErrKnockServerNotFound
	}
	sendAddr := inst.SendAddr()
	if sendAddr == nil {
		log.Critical("agent(%s)[RequestOtp] server IP cannot be parsed (instance %s)", otpMsg.UserId, inst.HostPort())
		return common.ErrKnockServerNotFound
	}

	otpMd := &core.MsgData{
		RemoteAddr:    sendAddr.(*net.UDPAddr),
		HeaderType:    core.NHP_OTP,
		CipherScheme:  a.config.DefaultCipherScheme,
		TransactionId: a.device.NextCounterIndex(),
		Compress:      true,
		Message:       otpBytes,
		PeerPk:        serverPeer.PublicKey(),
	}

	if !a.IsRunning() {
		log.Error("agent(%s#%d)[RequestOtp] MsgData channel closed or being closed, skip sending", otpMsg.UserId, otpMd.TransactionId)
		return common.ErrPacketToMessageRoutineStopped
	}

	// IsRunning() above is only a cheap fast-path reject; it cannot make
	// the bare send safe, because Stop() can win the race between the
	// check and the send and close(sendMsgCh) underneath us (these request
	// methods do not register with a.wg, so wg.Wait() in Stop() does not
	// wait for them — unlike Knock). signals.stop is closed BEFORE
	// sendMsgCh in Stop(), so selecting on it means we either send while
	// the channel is still open or bail out cleanly — never send on a
	// closed channel. device will create or find existing connection and
	// sends the MsgAssembler via that connection.
	select {
	case a.sendMsgCh <- otpMd:
	case <-a.signals.stop:
		log.Error("agent(%s#%d)[RequestOtp] message routine stopped, skip sending", otpMsg.UserId, otpMd.TransactionId)
		return common.ErrPacketToMessageRoutineStopped
	}

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

	serverPeer := target.GetServerPeer()
	if serverPeer == nil {
		log.Critical("agent(%s)[RegisterPublicKey] server is not assigned", regMsg.UserId)
		return nil, common.ErrKnockServerNotFound
	}
	inst := target.PickInstance()
	if inst == nil {
		log.Critical("agent(%s)[RegisterPublicKey] no instance available", regMsg.UserId)
		return nil, common.ErrKnockServerNotFound
	}
	sendAddr := inst.SendAddr()
	if sendAddr == nil {
		log.Critical("agent(%s)[RegisterPublicKey] server IP cannot be parsed (instance %s)", regMsg.UserId, inst.HostPort())
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
		PeerPk:        serverPeer.PublicKey(),
		ResponseMsgCh: make(chan *core.PacketParserData),
	}

	if !a.IsRunning() {
		log.Error("agent(%s#%d)[RegisterPublicKey] MsgData channel closed or being closed, skip sending", regMsg.UserId, regMd.TransactionId)
		return nil, common.ErrPacketToMessageRoutineStopped
	}

	// See RequestOtp: select on signals.stop so a concurrent Stop() that
	// closes sendMsgCh after the IsRunning() check can't panic the send.
	select {
	case a.sendMsgCh <- regMd:
	case <-a.signals.stop:
		log.Error("agent(%s#%d)[RegisterPublicKey] message routine stopped, skip sending", regMsg.UserId, regMd.TransactionId)
		return nil, common.ErrPacketToMessageRoutineStopped
	}

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

	serverPeer := target.GetServerPeer()
	if serverPeer == nil {
		log.Critical("agent(%s)[ListResource] server is not assigned", lstMsg.UserId)
		return nil, common.ErrKnockServerNotFound
	}
	inst := target.PickInstance()
	if inst == nil {
		log.Critical("agent(%s)[ListResource] no instance available", lstMsg.UserId)
		return nil, common.ErrKnockServerNotFound
	}
	sendAddr := inst.SendAddr()
	if sendAddr == nil {
		log.Critical("agent(%s)[ListResource] server IP cannot be parsed (instance %s)", lstMsg.UserId, inst.HostPort())
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
		PeerPk:        serverPeer.PublicKey(),
		ResponseMsgCh: make(chan *core.PacketParserData),
	}

	if !a.IsRunning() {
		log.Error("agent(%s#%d)[ListResource] MsgData channel closed or being closed, skip sending", lstMsg.UserId, lstMd.TransactionId)
		return nil, common.ErrPacketToMessageRoutineStopped
	}

	// See RequestOtp: select on signals.stop so a concurrent Stop() that
	// closes sendMsgCh after the IsRunning() check can't panic the send.
	select {
	case a.sendMsgCh <- lstMd:
	case <-a.signals.stop:
		log.Error("agent(%s#%d)[ListResource] message routine stopped, skip sending", lstMsg.UserId, lstMd.TransactionId)
		return nil, common.ErrPacketToMessageRoutineStopped
	}

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
