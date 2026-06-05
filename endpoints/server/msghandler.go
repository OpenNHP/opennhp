package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	wasmEngine "github.com/OpenNHP/opennhp/nhp/core/wasm/engine"
	"github.com/OpenNHP/opennhp/nhp/log"
	utils "github.com/OpenNHP/opennhp/nhp/utils"
)

// relayConnKeyPrefix is what every relay-forwarded connection's mapKey
// begins with (see HandleRelayForward). Used by helpers below to
// distinguish relay-forwarded entries from direct UDP entries.
const relayConnKeyPrefix = "relay|"

// relayConnKeySep separates the relay-addr segment from the
// real-client-addr segment inside a relay-forwarded mapKey. We
// deliberately avoid ':' here because net.UDPAddr.String() formats
// IPv6 as "[2001:db8::1]:80" — colons appear *inside* the address
// itself, so a ':'-separated key cannot be reliably parsed from the
// right with strings.LastIndexByte. '|' never appears in any IP-or-
// host:port serialization Go produces, so a single split is enough.
const relayConnKeySep = "|"

// relayAddrFromConnKey extracts the relay's address (the
// "<relayHost>:<relayPort>" segment) from a relay-forwarded mapKey of
// the form "relay|<relayAddr>|<realClientAddr>". Returns "" if the key
// is not a relay-forwarded one or doesn't carry both segments.
//
// The map key is in-memory only (never persisted, never wire-
// serialized), so the '|' separator is a free invariant — see
// relayConnKeySep for why we picked it over ':'.
func relayAddrFromConnKey(mapKey string) string {
	if !strings.HasPrefix(mapKey, relayConnKeyPrefix) {
		return ""
	}
	rest := mapKey[len(relayConnKeyPrefix):]
	// rest = "<relayAddr>|<realClientAddr>"; each segment is itself
	// "host:port" but contains no '|', so one split recovers both.
	idx := strings.IndexByte(rest, relayConnKeySep[0])
	if idx <= 0 {
		// idx==0 means "relay||..." (empty relay segment); idx<0
		// means the real-client segment is missing entirely.
		return ""
	}
	if idx == len(rest)-1 {
		// "relay|<addr>|" — real-client segment empty.
		return ""
	}
	return rest[:idx]
}

// incRelayConnCount and decRelayConnCount maintain a per-relay counter
// of forwarded-client connections so HandleRelayForward can enforce
// MaxConnectionsPerRelay in O(1) instead of scanning the whole
// connection map under the global mutex.
func (s *UdpServer) incRelayConnCount(relayAddr string) {
	s.relayConnCountMutex.Lock()
	s.relayConnCount[relayAddr]++
	s.relayConnCountMutex.Unlock()
}

func (s *UdpServer) decRelayConnCount(relayAddr string) {
	s.relayConnCountMutex.Lock()
	s.relayConnCount[relayAddr]--
	if s.relayConnCount[relayAddr] <= 0 {
		delete(s.relayConnCount, relayAddr)
	}
	s.relayConnCountMutex.Unlock()
}

func (s *UdpServer) getRelayConnCount(relayAddr string) int {
	s.relayConnCountMutex.Lock()
	defer s.relayConnCountMutex.Unlock()
	return s.relayConnCount[relayAddr]
}

// teardownPerRelayCounter encapsulates the connection-teardown
// decision "should I decrement the per-relay slot counter?". It's
// extracted so the invariant — at most one dec per inc, and zero dec
// when HRF has transferred the slot to a replacement conn — is
// testable without spinning up a full UdpServer.
//
// Two independent signals gate the dec:
//
//   - stillPresent: the caller still owned the map entry when the
//     teardown defer ran. If false, HandleRelayForward's stale-replace
//     path has already removed the entry and (atomically, under the
//     same map mutex) marked this conn as replaced — the slot has
//     transitioned to the replacement.
//
//   - !conn.replaced.Load(): belt-and-suspenders. Even on a future
//     refactor that removes the entry without going through HRF's
//     stale-replace path, replaced=false means no one has claimed the
//     slot, so the original dec is the right cleanup.
//
// Both must hold for the dec to fire. The pair is what eliminates the
// previous race where CR-defer and HRF could both dec the same slot
// (or, after a naïve fix, both skip the dec): the slot transfer is
// signaled explicitly, not inferred from "who deleted the map entry
// first".
func (s *UdpServer) teardownPerRelayCounter(conn *UdpConn, mapKey string, stillPresent bool) {
	if !stillPresent {
		return
	}
	if conn.replaced.Load() {
		return
	}
	relayAddr := relayAddrFromConnKey(mapKey)
	if relayAddr == "" {
		return
	}
	s.decRelayConnCount(relayAddr)
}

// isRoutablePublicIP returns true only if ip is a plausible public client
// address. Relay peers are trusted to forward traffic, but a malicious or
// compromised relay must not be able to spam the connection map with
// fabricated SourceAddr values, so non-routable ranges are rejected.
func isRoutablePublicIP(ip net.IP) bool {
	if ip == nil || ip.IsUnspecified() || ip.IsLoopback() ||
		ip.IsMulticast() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsPrivate() {
		return false
	}
	// Reject CGNAT (100.64.0.0/10) — not exposed by net.IP helpers.
	if ip4 := ip.To4(); ip4 != nil && ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
		return false
	}
	return true
}

// validateRelaySourceAddr decides whether the (ip, port) pair extracted from
// a RelayForwardMsg.SourceAddr should be accepted. Returns "" on accept, or
// a short reason string on reject (used both in the log line and the
// error returned to the relay).
//
// Port sanity is enforced unconditionally; out-of-range ports are never
// legitimate. The IP-routability check is the part that conflicts with the
// docker-compose demo (Docker Desktop's vpnkit gateway is RFC1918), so it
// is gated behind allowPrivate. Keep that flag OFF in production: it is
// the only thing stopping a compromised relay from injecting any private-
// range SourceAddr it likes into the server's connection map and the
// downstream AC ipset whitelist.
func validateRelaySourceAddr(ip net.IP, port int, allowPrivate bool) string {
	if ip == nil || port <= 0 || port > 65535 {
		return "malformed"
	}
	if !allowPrivate && !isRoutablePublicIP(ip) {
		return "non-routable"
	}
	return ""
}

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
	dsaMsg := &common.DSAMsg{}
	if err != nil {
		dsaMsg.DoId = doId
		dsaMsg.ErrCode = 1
		dsaMsg.ErrMsg = err.Error()
	} else {
		dsaMsg.DoId = doId
		dsaMsg.SpoId = config.Spo.PolicyId
		dsaMsg.Spo = &config.Spo
		dsaMsg.TTL = int((30 * time.Minute).Milliseconds())
		s.UpdateTeePublicKeyAndConsumerEphemeralPublicKey(darMsg.TeePublicKey, darMsg.ConsumerEphemeralPublicKey, ppd.RemotePubKey)
	}

	aakBytes, _ := json.Marshal(dsaMsg)
	log.Debug("dagMsg:%s", (string)(aakBytes))
	aakMd := &core.MsgData{
		HeaderType:     core.NHP_DSA,
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

func (s *UdpServer) HandleDHPDAVMessage(ppd *core.PacketParserData) (err error) {
	defer s.wg.Done()
	s.wg.Add(1)

	transactionId := ppd.SenderTrxId
	addrStr := ppd.ConnData.RemoteAddr.String()
	davMsg := &common.DAVMsg{}

	err = json.Unmarshal(ppd.BodyMessage, davMsg)
	if err != nil {
		log.Error("server-agent(#%d@%s)[HandleDHPDAVMessage] failed to parse %s message: %v", transactionId, addrStr, core.HeaderTypeToString(ppd.HeaderType), err)
		return err
	}

	doId := davMsg.DoId
	config, err := ReadZdtoConfig(doId)

	dagMsg := &common.DAGMsg{}
	if err != nil {
		// ReadZdtoConfig failed → config is the zero value, so its
		// Spo.Policy is empty. Running onAttestationVerify on it would
		// vacuously pass (the function returns nil when Policy == ""),
		// effectively bypassing attestation on a misconfigured /
		// missing object. Short-circuit instead: report the
		// config-read failure to the agent and skip attestation
		// entirely. Attestation must never be evaluated against a
		// policy we couldn't load.
		log.Error("server-agent(#%d@%s)[HandleDHPDAVMessage] ReadZdtoConfig(%s) failed: %v", transactionId, addrStr, doId, err)
		dagMsg.DoId = doId
		dagMsg.ErrCode = 1
		dagMsg.ErrMsg = err.Error()
	} else if attErr := s.onAttestationVerify(&config.Spo, davMsg.Evidence); attErr != nil {
		log.Error("server-agent(#%d@%s)[HandleDHPDAVMessage] failed to verify attesation: %s with error: %s", transactionId, addrStr, davMsg.Evidence, attErr.Error())
		return attErr
	} else {
		dagMsg.DoId = doId

		teePublicKey, consumerEphemeralPublicKey := s.GetTeePublicKeyBase64AndConsumerEphemeralPublicKeyBase64(ppd.RemotePubKey)

		dwrMsg := &common.DWRMsg{
			DoId:                       doId,
			TeePublicKey:               teePublicKey,
			ConsumerEphemeralPublicKey: consumerEphemeralPublicKey,
		}

		dbConn, found := s.dbConnectionMap[config.DbId]
		if !found {
			log.Critical("dbConn not found for dbId:%s", config.DbId)
			err = common.ErrDBOffline
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

func (s *UdpServer) onAttestationVerify(spo *common.SmartPolicy, attestation string) error {
	if spo.Policy == "" {
		return nil
	}

	wasmBytes, err := base64.StdEncoding.DecodeString(spo.Policy)
	if err != nil {
		wasmPath, err := utils.DownloadFileToTemp(spo.Policy, "wasm-")
		defer os.Remove(filepath.Dir(wasmPath))
		defer os.Remove(wasmPath)
		if err != nil {
			return err
		}
		wasmBytes, err = os.ReadFile(wasmPath)
		if err != nil {
			return err
		}
	}

	engine := wasmEngine.NewEngine()
	err = engine.LoadWasm(wasmBytes)
	defer engine.Close()
	if err != nil {
		return err
	}

	if engine.OnAttestationVerify(attestation) {
		return nil
	} else {
		return fmt.Errorf("attestation verification failed")
	}
}

func SaveZdtoConfig(drgMsg *common.DRGMsg) error {
	objectId := drgMsg.DoId
	configFileName := "data-" + objectId + ".json"

	etcDir := filepath.Join(ExeDirPath, "etc", "ztdo")
	configPath := filepath.Join(etcDir, configFileName)

	if existingDrgMsg, err := ReadZdtoConfig(objectId); err == nil {
		// alway keep original date source type
		drgMsg.DataSourceType = existingDrgMsg.DataSourceType

		if drgMsg.AccessUrl == "" { // provider update access url
			drgMsg.AccessUrl = existingDrgMsg.AccessUrl
		}

		os.Remove(configPath)
	}

	// Make sure the etc directory exists
	if err := os.MkdirAll(etcDir, 0755); err != nil {
		return fmt.Errorf("failed to create etc directory: %v", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("%v already exists, please delete it first", configFileName)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config.json: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(drgMsg)
}

// read data-<doId>.json to DRGMsg Object
func ReadZdtoConfig(doId string) (common.DRGMsg, error) {
	etcDir := filepath.Join(ExeDirPath, "etc", "ztdo")
	configFilePath := filepath.Join(etcDir, "data-"+doId+".json")
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

// HandleRelayForward processes a decrypted NHP_RLY message from the standard
// Noise pipeline.  The relay's identity has already been validated by
// validatePeer as part of the standard decryption flow.
//
// The message body is a JSON-encoded RelayForwardMsg containing:
//   - SourceAddr:  the real client's IP/port
//   - InnerPacket: base64-encoded inner NHP packet (encrypted by agent)
//
// The inner packet is injected into the standard pipeline as if the agent
// had connected directly.
func (s *UdpServer) HandleRelayForward(ppd *core.PacketParserData) error {
	var rlyMsg common.RelayForwardMsg
	if err := json.Unmarshal(ppd.BodyMessage, &rlyMsg); err != nil {
		log.Error("server-relay[HandleRelayForward] failed to parse RelayForwardMsg: %v", err)
		return err
	}

	if rlyMsg.SourceAddr == nil {
		log.Error("server-relay[HandleRelayForward] missing source address")
		return fmt.Errorf("missing source address")
	}

	innerBytes, err := base64.StdEncoding.DecodeString(rlyMsg.InnerPacket)
	if err != nil {
		log.Error("server-relay[HandleRelayForward] failed to decode inner packet: %v", err)
		return err
	}

	realIP := net.ParseIP(rlyMsg.SourceAddr.Ip)
	if reason := validateRelaySourceAddr(realIP, rlyMsg.SourceAddr.Port, s.allowPrivateRelaySource.Load()); reason != "" {
		log.Warning("server-relay[HandleRelayForward] rejecting %s from relay %s: %s:%d",
			reason, ppd.ConnData.RemoteAddr.String(), rlyMsg.SourceAddr.Ip, rlyMsg.SourceAddr.Port)
		return fmt.Errorf("%s relay source address", reason)
	}
	realAddr := &net.UDPAddr{IP: realIP, Port: rlyMsg.SourceAddr.Port}

	relayAddrStr := ppd.ConnData.RemoteAddr.String()
	log.Info("server-relay[HandleRelayForward] from relay %s, real client %s, inner %d bytes",
		relayAddrStr, realAddr, len(innerBytes))

	// Allocate a pool packet for the inner bytes.
	innerPkt := s.device.AllocatePoolPacket()
	if len(innerBytes) > len(innerPkt.Buf) {
		s.device.ReleasePoolPacket(innerPkt)
		log.Warning("server-relay[HandleRelayForward] inner packet too large (%d bytes)", len(innerBytes))
		return fmt.Errorf("inner packet too large")
	}
	copy(innerPkt.Buf[:], innerBytes)
	innerPkt.Content = innerPkt.Buf[:len(innerBytes)]

	// Run standard RecvPrecheck on the inner packet.
	innerType, _, err := s.device.RecvPrecheck(innerPkt)
	if err != nil {
		s.device.ReleasePoolPacket(innerPkt)
		log.Warning("server-relay[HandleRelayForward] inner RecvPrecheck failed: %v", err)
		return err
	}
	innerPkt.HeaderType = innerType
	log.Info("server-relay[HandleRelayForward] inner [%s] from real client %s via relay %s",
		core.HeaderTypeToString(innerType), realAddr, relayAddrStr)

	// Same RKN-under-overload gate as the direct-UDP path
	// (recvPacketRoutine), but keyed on the REAL client IP rather than the
	// relay's: a relay legitimately fans out many clients, so keying on
	// the relay address would let one busy relay's honest traffic throttle
	// itself while doing nothing to isolate a single abusive client. The
	// inner RKN reaches the same cookie-verify ECDH (via
	// ForwardInboundPacket -> connectionRoutine -> RecvPacketToMsg), so it
	// needs the same pre-ECDH throttle. Dropped-only, no block-listing.
	if innerType == core.NHP_RKN && s.device.IsOverload() {
		if !s.rknLimiter.allow(realAddr, time.Now().UnixNano()) {
			s.device.ReleasePoolPacket(innerPkt)
			log.Warning("server-relay[HandleRelayForward] inner RKN from real client %s (via relay %s) dropped: per-IP rate limit exceeded under overload",
				realAddr, relayAddrStr)
			return fmt.Errorf("rkn rate limit exceeded")
		}
	}

	// Build or reuse a connection keyed on "relay|<relayAddr>|<realClientAddr>".
	// This avoids collisions with the relay's own NHP_RLY connection (which is
	// already keyed on relayAddrStr) and isolates per-client anti-replay state.
	// RemoteAddr must be the relay's UDP address so that response packets
	// (ACK/COK) are sent back to the relay — the relay then forwards them
	// to the browser over HTTP.  The real client address is used only for
	// auth/logging purposes.
	//
	// The '|' separator is required, not cosmetic: an IPv6 address renders
	// as "[2001:db8::1]:80", so a ':'-delimited key could not be reliably
	// split from the right. See relayConnKeySep for the full reasoning.
	relayAddr := ppd.ConnData.RemoteAddr
	connKey := relayConnKeyPrefix + relayAddrStr + relayConnKeySep + realAddr.String()
	recvTime := time.Now().UnixNano()

	// Three steps run under a single map-mutex critical section so that
	// connectionRoutine's teardown defer can't interleave between stale
	// detection and slot transfer:
	//
	//   1. Detect a live or stale conn at connKey.
	//   2. If stale, mark it replaced and delete it (slot transfers).
	//   3. If absent, check the per-relay + global caps and insert a
	//      fresh conn with an inc'd counter.
	//
	// Previously each of these was its own short critical section, which
	// let CR-defer slip in between stale-detect and counter-dec, causing
	// the relay slot to be double-dec'd (and silently re-clamped to zero
	// inside decRelayConnCount). Holding the map mutex for the whole
	// transition keeps the (conn-table, per-relay-counter) pair
	// consistent — the per-relay counter is now incremented exactly when
	// a new map entry is created, and decremented exactly when the
	// routine that owns the entry tears it down.
	transferred := false
	s.remoteConnectionMapMutex.Lock()
	conn, found := s.remoteConnectionMap[connKey]
	if found && conn.ConnData.IsClosed() {
		// Previous relay-forwarded conn for this client timed out.
		// Mark replaced (under the same mutex that deletes the entry
		// and inserts the replacement) so CR-defer sees an explicit
		// "slot has been transferred, don't dec" signal regardless of
		// when it finally acquires the map mutex.
		conn.replaced.Store(true)
		delete(s.remoteConnectionMap, connKey)
		found = false
		transferred = true
	}

	if found {
		s.remoteConnectionMapMutex.Unlock()
		atomic.StoreInt64(&conn.ConnData.LastLocalRecvTime, recvTime)
	} else {
		// Apply the global overload guard while still holding the map
		// mutex — the map size we just observed is authoritative.
		total := len(s.remoteConnectionMap)
		if total >= MaxConcurrentConnection {
			s.remoteConnectionMapMutex.Unlock()
			// If we got here via the stale-replace path we already
			// deleted the OLD entry and marked it replaced, so the
			// OLD conn's teardown will skip its dec (stillPresent=false
			// AND replaced=true). Refusing to take over the slot here
			// without compensating would leak it permanently — the
			// per-relay counter would stay elevated with no live owner,
			// eventually capping the relay below MaxConnectionsPerRelay.
			// Mirror the per-relay branch's fix-up: reclaim the dec
			// ourselves. (decRelayConnCount takes relayConnCountMutex on
			// its own; do it after releasing the map mutex to preserve
			// the map→counter lock order used elsewhere.)
			if transferred {
				s.decRelayConnCount(relayAddrStr)
			}
			s.device.ReleasePoolPacket(innerPkt)
			log.Critical("server-relay[HandleRelayForward] reached MaxConcurrentConnection (%d), dropping forward from relay %s",
				MaxConcurrentConnection, relayAddrStr)
			return fmt.Errorf("server connection table full")
		}
		if total > OverloadConnectionThreshold {
			s.device.SetOverload(true)
		}

		// Per-relay cap: if we're transferring a slot from a stale
		// conn the counter already accounts for it, so just check the
		// cap; otherwise this is a fresh slot and we must inc.
		s.relayConnCountMutex.Lock()
		curr := s.relayConnCount[relayAddrStr]
		// The cap compares against curr (slot transfer) or curr+1
		// (genuinely new slot). Use the post-action value either way
		// so the check is uniform.
		post := curr
		if !transferred {
			post = curr + 1
		}
		if post > MaxConnectionsPerRelay {
			// When transferred is true the slot WAS owned by the OLD
			// conn whose map entry we just removed above; OLD's
			// teardown will run with stillPresent=false (entry gone)
			// and skip its dec. If we also refuse to take over, the
			// slot leaks — counter stays at curr forever despite no
			// live owner. Reverting replaced=false would help only if
			// OLD's teardown could still see the entry, but it can't
			// (we deleted it). So take the dec ourselves here, while
			// still holding relayConnCountMutex, instead of trying to
			// hand ownership back. Leave replaced=true so OLD's
			// teardown's belt-and-suspenders check also refuses to
			// dec, keeping the invariant "at most one dec per inc".
			//
			// In practice the OLD <= MaxConnectionsPerRelay invariant
			// means this branch is unreachable for the current
			// constant cap; it becomes reachable if
			// MaxConnectionsPerRelay is ever hot-reloaded to a lower
			// value mid-flight. The fix-up is cheap so do it
			// unconditionally.
			if transferred && curr > 0 {
				s.relayConnCount[relayAddrStr]--
				if s.relayConnCount[relayAddrStr] == 0 {
					delete(s.relayConnCount, relayAddrStr)
				}
			}
			s.relayConnCountMutex.Unlock()
			s.remoteConnectionMapMutex.Unlock()
			s.device.ReleasePoolPacket(innerPkt)
			log.Critical("server-relay[HandleRelayForward] relay %s exceeded MaxConnectionsPerRelay (%d), dropping forward",
				relayAddrStr, MaxConnectionsPerRelay)
			return fmt.Errorf("relay forward cap exceeded")
		}
		if !transferred {
			s.relayConnCount[relayAddrStr]++
		}
		s.relayConnCountMutex.Unlock()

		conn = &UdpConn{mapKey: connKey}
		conn.ConnData = &core.ConnectionData{
			InitTime:          recvTime,
			LastLocalRecvTime: recvTime,
			Device:            s.device,
			LocalAddr:         s.listenAddr,
			RemoteAddr:        relayAddr,
			RealRemoteAddr:    realAddr,
			// CookieStore omitted: see udpserver.go for rationale.
			RemoteTransactionMap: make(map[uint64]*core.RemoteTransaction),
			TimeoutMs:            DefaultAgentConnectionTimeoutMs,
			SendQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
			RecvQueue:            make(chan *core.Packet, PacketQueueSizePerConnection),
			BlockSignal:          make(chan struct{}),
			SetTimeoutSignal:     make(chan struct{}),
			StopSignal:           make(chan struct{}),
		}
		s.remoteConnectionMap[connKey] = conn
		s.remoteConnectionMapMutex.Unlock()

		s.wg.Add(1)
		go s.connectionRoutine(conn)
		log.Info("server-relay[HandleRelayForward] new relay connection %s (real client %s, key=%s, slotTransferred=%v)",
			relayAddrStr, realAddr, connKey, transferred)
	}

	conn.ConnData.ForwardInboundPacket(innerPkt)
	return nil
}
