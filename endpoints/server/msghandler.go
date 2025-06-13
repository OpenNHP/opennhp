package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// HandleOTPRequest
// Server will not respond to agent's otp request
func (s *UdpServer) HandleOTPRequest(ppd *core.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderTrxId
	addrStr := ppd.ConnData.RemoteAddr.String()

	otpMsg := &common.AgentOTPMsg{}
	err = json.Unmarshal(ppd.BodyMessage, otpMsg)
	if err != nil {
		log.Error("server-agent(#%d@%s)[HandleOTPRequest] failed to parse %s message: %v", transactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType), err)
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
func (s *UdpServer) HandleRegisterRequest(ppd *core.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderTrxId
	addrStr := ppd.ConnData.RemoteAddr.String()
	regMsg := &common.AgentRegisterMsg{}
	rakMsg := &common.ServerRegisterAckMsg{}

	func() {
		err = json.Unmarshal(ppd.BodyMessage, regMsg)
		if err != nil {
			log.Error("server-agent(#%d@%s)[HandleRegisterRequest] failed to parse %s message: %v", transactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType), err)
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
	rakMd := &core.MsgData{
		HeaderType:     core.NHP_RAK,
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
func (s *UdpServer) HandleListRequest(ppd *core.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderTrxId
	addrStr := ppd.ConnData.RemoteAddr.String()
	lstMsg := &common.AgentListMsg{}
	lrtMsg := &common.ServerListResultMsg{}

	func() {
		err = json.Unmarshal(ppd.BodyMessage, lstMsg)
		if err != nil {
			log.Error("server-agent(#%d@%s)[HandleListRequest] failed to parse %s message: %v", transactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType), err)
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
	ackMd := &core.MsgData{
		HeaderType:     core.NHP_LRT,
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

func (s *UdpServer) HandleACOnline(ppd *core.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderTrxId
	addrStr := ppd.ConnData.RemoteAddr.String()
	aolMsg := &common.ACOnlineMsg{}

	err = json.Unmarshal(ppd.BodyMessage, aolMsg)
	if err != nil {
		log.Error("server-ac(#%d@%s)[HandleACOnline] failed to parse %s message: %v", transactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType), err)
		return err
	}

	acId := aolMsg.ACId
	acPubkeyBase64 := base64.StdEncoding.EncodeToString(ppd.RemotePubKey)
	s.acPeerMapMutex.Lock()
	acPeer := s.acPeerMap[acPubkeyBase64] // ac peer's recvAddr has already been updated by nhp packet parser
	s.acPeerMapMutex.Unlock()

	acConn := &ACConn{
		ConnData:       ppd.ConnData,
		ACPeer:         acPeer,
		ACCipherScheme: ppd.CipherScheme,
		ACId:           acId,
		ServiceId:      aolMsg.AuthServiceId,
		Apps:           aolMsg.ResourceIds,
	}

	s.acConnectionMapMutex.Lock()
	s.acConnectionMap[acId] = acConn
	s.acConnectionMapMutex.Unlock()

	aakMsg := &common.ServerACAckMsg{
		ErrCode: common.ErrSuccess.ErrorCode(),
		ACAddr:  ppd.ConnData.RemoteAddr.String(),
	}
	aakBytes, _ := json.Marshal(aakMsg)

	aakMd := &core.MsgData{
		HeaderType:     core.NHP_AAK,
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

func (s *UdpServer) HandleDBOnline(ppd *core.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderTrxId
	addrStr := ppd.ConnData.RemoteAddr.String()
	dolMsg := &common.DBOnlineMsg{}

	err = json.Unmarshal(ppd.BodyMessage, dolMsg)
	if err != nil {
		log.Error("server-db(#%d@%s)[HandleDBOnline] failed to parse %s message: %v", transactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType), err)
		return err
	}

	dbId := dolMsg.DBId
	dbPubkeyBase64 := base64.StdEncoding.EncodeToString(ppd.RemotePubKey)
	s.dbPeerMapMutex.Lock()
	dbPeer := s.dbPeerMap[dbPubkeyBase64] // ac peer's recvAddr has already been updated by nhp packet parser
	s.dbPeerMapMutex.Unlock()

	dbConn := &DBConn{
		ConnData:       ppd.ConnData,
		DBPeer:         dbPeer,
		DBCipherScheme: ppd.CipherScheme,
		DBId:           dbId,
	}

	s.dbConnectionMapMutex.Lock()
	s.dbConnectionMap[dbId] = dbConn
	s.dbConnectionMapMutex.Unlock()

	aakMsg := &common.ServerDBAckMsg{
		ErrCode: common.ErrSuccess.ErrorCode(),
		DBAddr:  ppd.ConnData.RemoteAddr.String(),
	}
	aakBytes, _ := json.Marshal(aakMsg)

	aakMd := &core.MsgData{
		HeaderType:     core.NHP_DBA,
		TransactionId:  transactionId,
		Compress:       true,
		PrevParserData: ppd,
		Message:        aakBytes,
	}

	// forward to a specific transaction
	transaction := ppd.ConnData.FindRemoteTransaction(transactionId)
	if transaction == nil {
		log.Error("server-db(@%s#%d@%s)[HandleDBOnline] transaction is not available", dbId, transactionId, addrStr)
		err = common.ErrTransactionIdNotFound
		return err
	}

	transaction.NextMsgCh <- aakMd

	return nil
}

func (s *UdpServer) HandleDHPDARMessage(ppd *core.PacketParserData) (err error) {
	fmt.Println("HandleDHPDARMessage received")
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderTrxId
	addrStr := ppd.ConnData.RemoteAddr.String()
	darMsg := &common.DARMsg{}

	err = json.Unmarshal(ppd.BodyMessage, darMsg)
	if err != nil {
		log.Error("server-agent(#%d@%s)[HandleDHPDARMessage] failed to parse %s message: %v", transactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType), err)
		return err
	}

	doId := darMsg.DoId
	config, err := ReadZdtoConfig(doId)
	dagMsg := &common.DAGMsg{}
	if err != nil {
		dagMsg.DoId = doId
		dagMsg.ErrCode = 1
		dagMsg.ErrMsg = err.Error()
	} else {
		dagMsg.DoId = doId

		dwrMsg := &common.DWRMsg{
			DoId: doId,
			TeePublicKey: darMsg.TeePublicKey,
			ConsumerEphemeralPublicKey: darMsg.ConsumerEphemeralPublicKey,
		}

		dbConn, found := s.dbConnectionMap[config.DbId]
		if !found {
			log.Critical("dbConn not found for dbId:%s", config.DbId)
			err = common.ErrInvalidInput
			dagMsg.ErrCode = 1
			dagMsg.ErrMsg = err.Error()
		} else {
			dwaMsg, err := s.ProcessDataPrivateKeyWrapping(dwrMsg, dbConn)
			if err != nil || dwaMsg.ErrCode != 0 {
				dagMsg.ErrCode = dwaMsg.ErrCode
				dagMsg.ErrMsg = dwaMsg.ErrMsg
			}

			dagMsg.Kao = dwaMsg.Kao
			dagMsg.Spo = &config.Spo
			dagMsg.DataSourceType = config.DataSourceType
			dagMsg.AccessUrl = config.AccessUrl
			dagMsg.AccessByNHP = config.AccessByNHP
			dagMsg.DoType = config.DoType
		}
	}

	aakBytes, _ := json.Marshal(dagMsg)
	log.Debug("dagMsg:%s", (string)(aakBytes))
	aakMd := &core.MsgData{
		HeaderType:     core.NHP_DAG,
		TransactionId:  transactionId,
		Compress:       true,
		PrevParserData: ppd,
		Message:        aakBytes,
	}
	// forward to a specific transaction
	transaction := ppd.ConnData.FindRemoteTransaction(transactionId)
	if transaction == nil {
		log.Error("server-agent(@%s#%d@%s)[HandleDHPDARMessage] transaction is not available", doId, transactionId, addrStr)
		err = common.ErrTransactionIdNotFound
		return err
	}

	transaction.NextMsgCh <- aakMd

	return nil
}

// HandleDHPDRGMessage
func (s *UdpServer) HandleDHPDRGMessage(ppd *core.PacketParserData) (err error) {
	fmt.Println("HandleDHPDRGMessage start")
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderTrxId
	addrStr := ppd.ConnData.RemoteAddr.String()
	aolMsg := &common.DRGMsg{}

	err = json.Unmarshal(ppd.BodyMessage, aolMsg)
	if err != nil {
		log.Error("server-Device(#%d@%s)[HandleDHPDRGMessage] failed to parse %s message: %v", transactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType), err)
		return err
	}

	doId := aolMsg.DoId

	err = SaveZdtoConfig(aolMsg)

	errCode := 0 //success
	errMsg := ""

	if err != nil {
		errCode = 1
		errMsg = err.Error()
	}

	aakMsg := &common.DAKMsg{
		DoId:    doId,
		ErrCode: errCode,
		ErrMsg:  errMsg,
	}
	aakBytes, _ := json.Marshal(aakMsg)

	aakMd := &core.MsgData{
		HeaderType:     core.NHP_DAK,
		TransactionId:  transactionId,
		Compress:       true,
		PrevParserData: ppd,
		Message:        aakBytes,
	}

	// forward to a specific transaction

	transaction := ppd.ConnData.FindRemoteTransaction(transactionId)
	if transaction == nil {
		log.Error("server-DB(@%s#%d@%s)[HandleDHPDRGMessage] transaction is not available", doId, transactionId, addrStr)
		err = common.ErrTransactionIdNotFound
		return err
	}
	transaction.NextMsgCh <- aakMd
	return nil
}

func SaveZdtoConfig(drgMsg *common.DRGMsg) error {

	objectId := drgMsg.DoId
	// Make sure the etc directory exists
	etcDir := "etc/ztdo"
	if err := os.MkdirAll(etcDir, 0755); err != nil {
		return fmt.Errorf("failed to create etc directory: %v", err)
	}

	// Check if etc/config.json already exists
	configFileName := "config-" + objectId + ".json"
	configPath := filepath.Join(etcDir, configFileName)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("%v already exists, please delete it first", configFileName)
	}

	// Create etc/config.json file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config.json: %v", err)
	}
	defer file.Close()

	// Write data to config.json
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(drgMsg)
}

// read ztdo-config.json to DRGMsg Object
func ReadZdtoConfig(doId string) (common.DRGMsg, error) {
	etcDir := "etc/ztdo"
	configFilePath := filepath.Join(etcDir, "config-"+doId+".json")
	file, err := os.Open(configFilePath)
	if err != nil {
		return common.DRGMsg{}, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	fileContentByte, err := io.ReadAll(file)
	if err != nil {
		return common.DRGMsg{}, fmt.Errorf("error reading file: %v", err)
	}

	var config common.DRGMsg

	err = json.Unmarshal(fileContentByte, &config)
	if err != nil {
		return common.DRGMsg{}, fmt.Errorf("json parsing error: %s", err)
	}
	return config, nil
}

