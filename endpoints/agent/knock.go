package agent

import (
	"encoding/json"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	wasmEngine "github.com/OpenNHP/opennhp/nhp/core/wasm/engine"
	"github.com/OpenNHP/opennhp/nhp/log"
)

func (a *UdpAgent) Knock(res *KnockTarget) (ackMsg *common.ServerKnockAckMsg, err error) {
	defer a.wg.Done()
	a.wg.Add(1)

	errWaitTime := (core.AgentLocalTransactionResponseTimeoutMs - 100) * time.Millisecond
	startTime := time.Now()

	ackMsg, err = a.knockRequest(res, false)
	if err == common.ErrKnockTerminatedByCookie {
		// if cookie is required by server, use cookie to knock
		// Note: don't use recursive calling method, it may get too deep if cookie message is kept sending
		// use flat calling
		ackMsg, err = a.knockRequest(res, true)
	}
	if ackMsg.ErrCode == common.ErrPacketEncryptionFailed.ErrorCode() {
		// local failure, packet not sent
		return ackMsg, err
	}
	if err != nil {
		// if local error happens, wait some time to return
		elapsedTime := time.Since(startTime)
		if elapsedTime < errWaitTime {
			time.Sleep(errWaitTime - elapsedTime)
		}
		return ackMsg, err
	}

	// deal with ac PASS_ACCESS_IP mode
	if len(ackMsg.PreAccessActions) > 0 {
		_ = a.preAccessRequest(ackMsg)
	}
	res.LastKnockSuccessTime = time.Now()

	log.Info("agent(%s)[KnockRequest] knock for %s:%s success, duration %d seconds", a.knockUser.UserId, res.AuthServiceId, res.ResourceId, ackMsg.OpenTime)
	return ackMsg, err
}

func (a *UdpAgent) knockRequest(res *KnockTarget, useCookie bool) (ackMsg *common.ServerKnockAckMsg, err error) {
	serverPeer := res.GetServerPeer()
	if serverPeer == nil {
		log.Critical("agent(%s)[KnockRequest] knock server is not assigned", a.knockUser.UserId)
		return nil, common.ErrKnockServerNotFound
	}

	sendAddr := serverPeer.SendAddr()
	if sendAddr == nil {
		log.Critical("agent(%s)[KnockRequest] knock server IP cannot be parsed", a.knockUser.UserId)
		return nil, common.ErrKnockServerNotFound
	}
	addrStr := sendAddr.String()

	a.knockUserMutex.RLock()
	knkMsg := &common.AgentKnockMsg{
		UserId:         a.knockUser.UserId,
		DeviceId:       a.deviceId,
		OrganizationId: a.knockUser.OrganizationId,
		AuthServiceId:  res.AuthServiceId,
		ResourceId:     res.ResourceId,
		CheckResults:   a.checkResults,
		UserData:       a.knockUser.UserData,
	}
	a.knockUserMutex.RUnlock()

	knkBytes, _ := json.Marshal(knkMsg)
	knkMd := &core.MsgData{
		RemoteAddr:    sendAddr.(*net.UDPAddr),
		HeaderType:    core.NHP_KNK,
		CipherScheme:  a.config.DefaultCipherScheme,
		TransactionId: a.device.NextCounterIndex(),
		Compress:      true,
		Message:       knkBytes,
		PeerPk:        serverPeer.PublicKey(),
		ResponseMsgCh: make(chan *core.PacketParserData),
	}
	if useCookie {
		knkMd.HeaderType = core.NHP_RKN
	}

	ackMsg = &common.ServerKnockAckMsg{}
	if !a.IsRunning() {
		log.Error("agent(%s#%d)[KnockRequest] MsgData channel closed or being closed, skip sending", knkMsg.UserId, knkMd.TransactionId)
		err = common.ErrPacketToMessageRoutineStopped
		ackMsg.ErrCode = common.ErrPacketToMessageRoutineStopped.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	// device will create or find existing connection and sends the MsgAssembler via that connection
	a.sendMsgCh <- knkMd

	// block until transaction completes
	serverPpd := <-knkMd.ResponseMsgCh
	close(knkMd.ResponseMsgCh)

	if serverPpd.Error != nil {
		log.Error("agent(%s#%d)[KnockRequest] failed to receive response from server %s: %v", knkMsg.UserId, knkMd.TransactionId, addrStr, serverPpd.Error)
		err = serverPpd.Error
		ackMsg.ErrCode = common.ErrPacketEncryptionFailed.ErrorCode()
		ackMsg.ErrMsg = serverPpd.Error.Error()
		return ackMsg, err
	}

	if serverPpd.HeaderType == core.NHP_COK {
		log.Error("agent(%s#%d)[KnockRequest] terminated by server's cookie message", knkMsg.UserId, knkMd.TransactionId)
		err = common.ErrKnockTerminatedByCookie
		ackMsg.ErrCode = common.ErrKnockTerminatedByCookie.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	if serverPpd.HeaderType != core.NHP_ACK {
		log.Error("agent(%s#%d)[KnockRequest] response has wrong type: %s", knkMsg.UserId, knkMd.TransactionId, core.HeaderTypeToString(serverPpd.HeaderType))
		err = common.ErrTransactionRepliedWithWrongType
		ackMsg.ErrCode = common.ErrTransactionRepliedWithWrongType.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	err = json.Unmarshal(serverPpd.BodyMessage, ackMsg)
	if err != nil {
		log.Error("agent(%s#%d)[KnockRequest] failed to parse %s message: %v", knkMsg.UserId, knkMd.TransactionId, core.HeaderTypeToString(serverPpd.HeaderType), err)
		ackMsg.ErrCode = common.ErrJsonParseFailed.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	if ackMsg.ErrCode != common.ErrSuccess.ErrorCode() {
		log.Error("agent(%s#%d)[KnockRequest] response error: %s", knkMsg.UserId, knkMd.TransactionId, ackMsg.ErrMsg)
		err = common.ErrorCodeToError(ackMsg.ErrCode)
		return ackMsg, err
	}

	log.Info("agent(%s#%d)[KnockRequest] succeed", knkMsg.UserId, knkMd.TransactionId)
	return ackMsg, nil
}

func (a *UdpAgent) ExitKnockRequest(res *KnockTarget) (ackMsg *common.ServerKnockAckMsg, err error) {
	serverPeer := res.GetServerPeer()
	if serverPeer == nil {
		log.Critical("agent(%s)[ExitKnockRequest] knock server is not assigned", a.knockUser.UserId)
		return nil, common.ErrKnockServerNotFound
	}

	sendAddr := serverPeer.SendAddr()
	if sendAddr == nil {
		log.Critical("agent(%s)[ExitKnockRequest] knock server IP cannot be parsed", a.knockUser.UserId)
		return nil, common.ErrKnockServerNotFound
	}
	addrStr := sendAddr.String()

	a.knockUserMutex.RLock()
	knkMsg := &common.AgentKnockMsg{
		UserId:         a.knockUser.UserId,
		DeviceId:       a.deviceId,
		OrganizationId: a.knockUser.OrganizationId,
		AuthServiceId:  res.AuthServiceId,
		ResourceId:     res.ResourceId,
		CheckResults:   a.checkResults,
		UserData:       a.knockUser.UserData,
	}
	a.knockUserMutex.RUnlock()

	knkBytes, _ := json.Marshal(knkMsg)
	knkMd := &core.MsgData{
		RemoteAddr:    sendAddr.(*net.UDPAddr),
		HeaderType:    core.NHP_EXT,
		CipherScheme:  a.config.DefaultCipherScheme,
		TransactionId: a.device.NextCounterIndex(),
		Compress:      true,
		Message:       knkBytes,
		PeerPk:        serverPeer.PublicKey(),
		ResponseMsgCh: make(chan *core.PacketParserData),
	}

	ackMsg = &common.ServerKnockAckMsg{}
	if !a.IsRunning() {
		log.Error("agent(%s#%d)[ExitKnockRequest] MsgData channel closed or being closed, skip sending", knkMsg.UserId, knkMd.TransactionId)
		err = common.ErrPacketToMessageRoutineStopped
		ackMsg.ErrCode = common.ErrPacketToMessageRoutineStopped.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	// device will create or find existing connection and sends the MsgAssembler via that connection
	a.sendMsgCh <- knkMd

	// block until transaction completes
	serverPpd := <-knkMd.ResponseMsgCh
	close(knkMd.ResponseMsgCh)

	if serverPpd.Error != nil {
		log.Error("agent(%s#%d)[ExitKnockRequest] failed to receive response from server %s: %v", knkMsg.UserId, knkMd.TransactionId, addrStr, serverPpd.Error)
		err = serverPpd.Error
		ackMsg.ErrCode = common.ErrTransactionFailedByTimeout.ErrorCode()
		ackMsg.ErrMsg = serverPpd.Error.Error()
		return ackMsg, err
	}

	if serverPpd.HeaderType == core.NHP_COK {
		log.Error("agent(%s#%d)[ExitKnockRequest] terminated by server's cookie message", knkMsg.UserId, knkMd.TransactionId)
		err = common.ErrKnockTerminatedByCookie
		ackMsg.ErrCode = common.ErrKnockTerminatedByCookie.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	if serverPpd.HeaderType != core.NHP_ACK {
		log.Error("agent(%s#%d)[ExitKnockRequest] response has wrong type: %s", knkMsg.UserId, knkMd.TransactionId, core.HeaderTypeToString(serverPpd.HeaderType))
		err = common.ErrTransactionRepliedWithWrongType
		ackMsg.ErrCode = common.ErrTransactionRepliedWithWrongType.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	err = json.Unmarshal(serverPpd.BodyMessage, ackMsg)
	if err != nil {
		log.Error("agent(%s#%d)[ExitKnockRequest] failed to parse %s message: %v", knkMsg.UserId, knkMd.TransactionId, core.HeaderTypeToString(serverPpd.HeaderType), err)
		ackMsg.ErrCode = common.ErrJsonParseFailed.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	if ackMsg.ErrCode != common.ErrSuccess.ErrorCode() {
		log.Error("agent(%s#%d)[ExitKnockRequest] response error: %s", knkMsg.UserId, knkMd.TransactionId, ackMsg.ErrMsg)
		err = common.ErrorCodeToError(ackMsg.ErrCode)
		return ackMsg, err
	}

	log.Info("agent(%s#%d)[ExitKnockRequest] succeed", knkMsg.UserId, knkMd.TransactionId)
	return ackMsg, nil
}

// agent -> ac, pre-access
func (a *UdpAgent) preAccessRequest(ackMsg *common.ServerKnockAckMsg) (err error) {
	a.knockUserMutex.RLock()
	if len(a.knockUser.UserId) == 0 {
		a.knockUserMutex.RUnlock()
		return common.ErrKnockUserNotSpecified
	}
	a.knockUserMutex.RUnlock()

	var acWg sync.WaitGroup
	for _, action := range ackMsg.PreAccessActions {
		acWg.Add(1)
		go func(info *common.PreAccessInfo) {
			defer acWg.Done()
			if info != nil {
				_ = a.processPreAccessAction(info)
			}
		}(action)
	}
	acWg.Wait()

	return nil
}

func (a *UdpAgent) processPreAccessAction(info *common.PreAccessInfo) error {
	acIp := net.ParseIP(info.AccessIp)
	if acIp == nil {
		return common.ErrInvalidIpAddress
	}

	acPort, _ := strconv.Atoi(info.AccessPort)
	if acPort <= 0 {
		return common.ErrInvalidIpAddress
	}

	udpACAddr := &net.UDPAddr{
		IP:   acIp,
		Port: acPort,
	}
	tcpACAddr := &net.TCPAddr{
		IP:   acIp,
		Port: acPort,
	}
	acPeer := &core.UdpPeer{
		PubKeyBase64: info.ACPubKey,
		Ip:           info.AccessIp,
		Port:         acPort,
	}
	acPk := acPeer.PublicKey()
	a.device.AddPeer(acPeer)

	a.knockUserMutex.RLock()
	accMsg := &common.AgentAccessMsg{
		UserId:         a.knockUser.UserId,
		DeviceId:       a.deviceId,
		OrganizationId: a.knockUser.OrganizationId,
		ACToken:        info.ACToken,
		UserData:       a.knockUser.UserData,
	}
	a.knockUserMutex.RUnlock()
	accBytes, _ := json.Marshal(accMsg)

	accMd := &core.MsgData{
		RemoteAddr:     udpACAddr,
		HeaderType:     core.NHP_ACC,
		CipherScheme:   a.config.DefaultCipherScheme,
		TransactionId:  a.device.NextCounterIndex(),
		Compress:       true,
		Message:        accBytes,
		PeerPk:         acPk,
		EncryptedPktCh: make(chan *core.MsgAssemblerData),
	}

	if !a.IsRunning() {
		log.Error("agent(%s)[PreAccessRequest] MsgData channel closed or being closed, skip sending", accMsg.UserId)
		return common.ErrPacketToMessageRoutineStopped
	}

	// start message encryption
	a.device.SendMsgToPacket(accMd)

	// waiting for message encryption
	accMad := <-accMd.EncryptedPktCh
	close(accMd.EncryptedPktCh)

	if accMad.Error != nil {
		log.Error("agent(%s)[PreAccessRequest] failed to encrypt access message: %v", accMsg.UserId, accMad.Error)
		return accMad.Error
	}

	// copy the packet for GC recycling and release the original packet buffer
	packetBytes := make([]byte, len(accMad.BasePacket.Content))
	copy(packetBytes, accMad.BasePacket.Content)
	accMad.Destroy()

	// open new routine(s) to send access packet to ac's temporary port
	go func(packet []byte, tcpAddr *net.TCPAddr) {
		// dial tcp connection and send accMad packet
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			log.Error("agent(%s)[PreAccessRequest] failed to connect to temporary tcp access port: %v", accMsg.UserId, err)
			return
		}
		defer conn.Close()

		_, err = conn.Write(packet)
		if err != nil {
			log.Error("agent(%s)[PreAccessRequest] failed to send tcp access packet: %v", accMsg.UserId, err)
		} else {
			log.Info("agent(%s)[PreAccessRequest] send tcp access message succeed", accMsg.UserId)
		}

	}(packetBytes, tcpACAddr)

	go func(packet []byte, udpAddr *net.UDPAddr) {
		// dial udp connection and send accMad packet
		conn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			log.Error("agent(%s)[PreAccessRequest] failed to connect to temporary udp access port: %v", accMsg.UserId, err)
			return
		}
		defer conn.Close()

		_, err = conn.Write(packet)
		if err != nil {
			log.Error("agent(%s)[PreAccessRequest] failed to send udp access packet: %v", accMsg.UserId, err)
		} else {
			log.Info("agent(%s)[PreAccessRequest] send udp access message succeed", accMsg.UserId)
		}

	}(packetBytes, udpACAddr)

	return nil
}

func (a *UdpAgent) KnockDHP() (ackMsg *common.ServerDHPKnockAckMsg, err error) {
	serverPeer := a.GetFirstServerPeer()
	if serverPeer == nil {
		log.Critical("agent(%s)[KnockDHP] knock server is not assigned", a.knockUser.UserId)
		return nil, common.ErrKnockServerNotFound
	}

	sendAddr := serverPeer.SendAddr()
	if sendAddr == nil {
		log.Critical("agent(%s)[KnockDHP] knock server IP cannot be parsed", a.knockUser.UserId)
		return nil, common.ErrKnockServerNotFound
	}
	addrStr := sendAddr.String()

	evidence, err := wasmEngine.GetEvidence()
	if err != nil {
		log.Error("agent(%s)[KnockDHP] cannot get evidence: %s", a.knockUser.UserId, err)
		return nil, common.ErrEvidenceGetFailed
	}

	a.knockUserMutex.RLock()
	knkMsg := &common.DHPKnockMsg{
		UserId:         a.knockUser.UserId,
		DeviceId:       a.deviceId,
		OrganizationId: a.knockUser.OrganizationId,
		UserData:       a.knockUser.UserData,
		Evidence:       evidence,
	}
	a.knockUserMutex.RUnlock()

	knkBytes, _ := json.Marshal(knkMsg)
	knkMd := &core.MsgData{
		RemoteAddr:    sendAddr.(*net.UDPAddr),
		HeaderType:    core.DHP_KNK,
		CipherScheme:  a.config.DefaultCipherScheme,
		TransactionId: a.device.NextCounterIndex(),
		Compress:      true,
		Message:       knkBytes,
		PeerPk:        serverPeer.PublicKey(),
		ResponseMsgCh: make(chan *core.PacketParserData),
	}

	ackMsg = &common.ServerDHPKnockAckMsg{}
	if !a.IsRunning() {
		log.Error("agent(%s#%d)[KnockDHP] MsgData channel closed or being closed, skip sending", knkMsg.UserId, knkMd.TransactionId)
		err = common.ErrPacketToMessageRoutineStopped
		ackMsg.ErrCode = common.ErrPacketToMessageRoutineStopped.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	// device will create or find existing connection and sends the MsgAssembler via that connection
	a.sendMsgCh <- knkMd

	// block until transaction completes
	serverPpd := <-knkMd.ResponseMsgCh
	close(knkMd.ResponseMsgCh)

	if serverPpd.Error != nil {
		log.Error("agent(%s#%d)[KnockDHP] failed to receive response from server %s: %v", knkMsg.UserId, knkMd.TransactionId, addrStr, serverPpd.Error)
		a.trustedByNHPServer.Store(false) // in this case, need to check local server configuration and remote agent public key configuration
		err = serverPpd.Error
		ackMsg.ErrCode = common.ErrPacketEncryptionFailed.ErrorCode()
		ackMsg.ErrMsg = serverPpd.Error.Error()
		return ackMsg, err
	} else { // In case that peer validation is enabled, agent public key has been configured correctly in server
		a.trustedByNHPServer.Store(true)
	}

	if serverPpd.HeaderType != core.NHP_ACK {
		log.Error("agent(%s#%d)[KnockDHP] response has wrong type: %s", knkMsg.UserId, knkMd.TransactionId, core.HeaderTypeToString(serverPpd.HeaderType))
		err = common.ErrTransactionRepliedWithWrongType
		ackMsg.ErrCode = common.ErrTransactionRepliedWithWrongType.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	err = json.Unmarshal(serverPpd.BodyMessage, ackMsg)
	if err != nil {
		log.Error("agent(%s#%d)[KnockDHP] failed to parse %s message: %v", knkMsg.UserId, knkMd.TransactionId, core.HeaderTypeToString(serverPpd.HeaderType), err)
		ackMsg.ErrCode = common.ErrJsonParseFailed.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

	if ackMsg.ErrCode != common.ErrSuccess.ErrorCode() {
		log.Error("agent(%s#%d)[KnockDHP] response error: %s", knkMsg.UserId, knkMd.TransactionId, ackMsg.ErrMsg)
		err = common.ErrorCodeToError(ackMsg.ErrCode)
		return ackMsg, err
	}

	log.Info("agent(%s#%d)[KnockDHP] succeed", knkMsg.UserId, knkMd.TransactionId)
	return ackMsg, nil
}
