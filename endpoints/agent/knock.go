package agent

import (
	"encoding/base64"
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
	// knockRequest has early-return paths (no cluster bound, no instance
	// available, unparseable send address) that return (nil, err) without
	// constructing an ackMsg. Synthesize one from err so the rest of Knock
	// can read ackMsg.ErrCode without dereferencing nil.
	if ackMsg == nil {
		if err == nil {
			err = common.ErrKnockServerNotFound
		}
		code := common.ErrorToErrorCode(err)
		if code == "" {
			code = common.ErrKnockServerNotFound.ErrorCode()
		}
		ackMsg = &common.ServerKnockAckMsg{
			ErrCode: code,
			ErrMsg:  err.Error(),
		}
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
		if err := a.preAccessRequest(ackMsg); err != nil {
			log.Warning("agent(%s)[KnockRequest] pre-access request failed: %v", a.knockUser.UserId, err)
		}
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

	// Pick the cluster instance for this send. When the cluster is
	// sticky (default), KNK and the follow-up RKN both route to the
	// same instance so cookie verification stays put even if the
	// cluster's instances don't share a cookie key. With
	// stateless-cookie clusters and Sticky=false, retries spread
	// across instances.
	inst := res.PickInstance()
	if inst == nil {
		log.Critical("agent(%s)[KnockRequest] no instance available in cluster %s",
			a.knockUser.UserId, serverPeer.PublicKeyBase64())
		return nil, common.ErrKnockServerNotFound
	}
	sendAddr := inst.SendAddr()
	if sendAddr == nil {
		log.Critical("agent(%s)[KnockRequest] knock server IP cannot be parsed (instance %s)",
			a.knockUser.UserId, inst.HostPort())
		// Drop the sticky pin so the next retry tries a sibling.
		res.ResetInstancePin()
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
		// Carry the cookie the previous KNK round stashed on this
		// target. ExternalCookie short-circuits the addHMAC fallback
		// to ConnData.CookieStore (initiator.go:444), which would
		// otherwise read from whichever conn this RKN is being sent
		// over — wrong conn in the Sticky=false multi-instance case.
		if c := res.ConsumePendingCookie(); c != nil {
			knkMd.ExternalCookie = c
		}
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
		// Pull the cookie bytes out of the COK body and stash them
		// on the target so the follow-up RKN can pass them via
		// MsgData.ExternalCookie. The legacy behavior wrote into
		// ppd.ConnData.CookieStore — see HandleCookieMessage — but
		// that's tied to the UDP conn the COK arrived through, and
		// a non-sticky cluster picks a different instance (=> new
		// conn => empty CookieStore) for the RKN. Stash on the
		// target so the value travels with the knock attempt.
		cokMsg := &common.ServerCookieMsg{}
		if uerr := json.Unmarshal(serverPpd.BodyMessage, cokMsg); uerr == nil {
			if cookieBytes, derr := base64.StdEncoding.DecodeString(cokMsg.Cookie); derr == nil {
				res.StashCookie(cookieBytes)
			} else {
				log.Error("agent(%s#%d)[KnockRequest] cookie base64 decode failed: %v",
					knkMsg.UserId, knkMd.TransactionId, derr)
			}
		} else {
			log.Error("agent(%s#%d)[KnockRequest] failed to parse NHP-COK body: %v",
				knkMsg.UserId, knkMd.TransactionId, uerr)
		}
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

	// Exit goes to the same instance the KNK landed on, so the
	// server cleans up the right session. PickInstance honors the
	// sticky pin captured during knock.
	inst := res.PickInstance()
	if inst == nil {
		log.Critical("agent(%s)[ExitKnockRequest] no instance available in cluster %s",
			a.knockUser.UserId, serverPeer.PublicKeyBase64())
		return nil, common.ErrKnockServerNotFound
	}
	sendAddr := inst.SendAddr()
	if sendAddr == nil {
		log.Critical("agent(%s)[ExitKnockRequest] knock server IP cannot be parsed (instance %s)",
			a.knockUser.UserId, inst.HostPort())
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

	// Same guard as RequestOtp / RegisterPublicKey / ListResource:
	// ExitKnockRequest is reachable directly from the unsynchronized SDK
	// exports (NhpAgentExitResource) and is NOT wg-tracked, so a concurrent
	// Stop() can close(sendMsgCh) after the IsRunning() check above. Select
	// on signals.stop (closed before sendMsgCh in Stop()) so the send either
	// completes while the channel is open or bails out cleanly. The sends in
	// knockRequest and KnockDHP don't need this — their callers (Knock /
	// dhpKnockResourceRoutine) are wg-tracked and complete before close.
	select {
	case a.sendMsgCh <- knkMd:
	case <-a.signals.stop:
		log.Error("agent(%s#%d)[ExitKnockRequest] message routine stopped, skip sending", knkMsg.UserId, knkMd.TransactionId)
		err = common.ErrPacketToMessageRoutineStopped
		ackMsg.ErrCode = common.ErrPacketToMessageRoutineStopped.ErrorCode()
		ackMsg.ErrMsg = err.Error()
		return ackMsg, err
	}

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
	// DHP knock has no KnockResource → no cluster routing input. Use
	// the "first cluster" pick (with the multi-cluster warning baked
	// into GetFirstServerCluster) and apply the cluster's LB policy
	// to pick an instance.
	sc := a.GetFirstServerCluster()
	if sc == nil {
		log.Critical("agent(%s)[KnockDHP] no server cluster configured", a.knockUser.UserId)
		return nil, common.ErrKnockServerNotFound
	}
	serverPeer := sc.representativePeer
	inst := sc.Pick()
	if inst == nil {
		log.Critical("agent(%s)[KnockDHP] no instance available in cluster %s",
			a.knockUser.UserId, serverPeer.PublicKeyBase64())
		return nil, common.ErrKnockServerNotFound
	}
	sendAddr := inst.SendAddr()
	if sendAddr == nil {
		log.Critical("agent(%s)[KnockDHP] knock server IP cannot be parsed (instance %s)",
			a.knockUser.UserId, inst.HostPort())
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
