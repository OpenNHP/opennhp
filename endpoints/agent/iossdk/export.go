package iossdk

import "C"
import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/OpenNHP/opennhp/endpoints/agent"
	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	_ "golang.org/x/mobile/bind"
)

var gAgentInstance *agent.UdpAgent
var gWorkingDir string
var gLogLevel int

// Initialization of the nhp_agent instance working directory path:
// The configuration files to be read are located under workingdir/etc/,
// and log files will be generated under workingdir/logs/.
//
// Input:
// workingDir: the working directory path for the agent
// logLevel: 0: silent, 1: error, 2: info, 3: debug, 4: verbose
//
// Return:
// Whether agent instance has been initialized successfully.
func NhpAgentInit(workingDir string, logLevel int) bool {
	if gAgentInstance != nil {
		return true
	}

	gAgentInstance = &agent.UdpAgent{}
	err := gAgentInstance.Start(workingDir, logLevel)
	if err != nil {
		return false
	}

	return true
}

// Synchronously stop and release nhp_
func NhpAgentClose() {
	if gAgentInstance == nil {
		return
	}

	gAgentInstance.Stop()
	gAgentInstance = nil
}

// Read the user information, resource information, server information,
// and other configuration files written under workingdir/etc,
// and asynchronously start the loop knocking thread.
//
// Input: None
//
// Return:
// -1: Uninitialized error
// >=0: The number of resources requested to knock by the knocking thread at the time of the call
//
//	(knocking resources will be synchronized with changes in the configuration in workingdir/etc/resource.toml).
//
//export NhpAgentKnockloopStart
func NhpAgentKnockloopStart() int {
	if gAgentInstance == nil {
		return -1
	}

	count := gAgentInstance.StartKnockLoop()
	return count
}

// Synchronously stop the loop, knock-on sub thread.
func NhpAgentKnockloopStop() {
	if gAgentInstance == nil {
		return
	}

	gAgentInstance.StopKnockLoop()
}

// Setting agent's represented user information
//
// Input:
// userId: User identification (optional, but not recommended to be empty)
// devId: Device identification (optional)
// orgId: Organization or company identification (optional)
// userData: Additional fields required to interface with backend services (json format string, optional)
//
// Return:
// Whether the user information is set successfully
func NhpAgentSetKnockUser(userId string, devId string, orgId string, userData string) bool {
	if gAgentInstance == nil {
		return false
	}
	var data map[string]any
	if len(userData) > 0 {
		err := json.Unmarshal([]byte(userData), &data)
		if err != nil {
			return false
		}
	}

	gAgentInstance.SetDeviceId(devId)
	gAgentInstance.SetKnockUser(userId, orgId, data)
	return true
}

// Add an NHP server information to the agent for use in knocking on the door
// (the agent can initiate different knocking requests to multiple NHP servers).
//
// Input:
// pubkey: Public key of the NHP server
// ip: IP address of the NHP server
// host: Domain name of the NHP server (if a domain name is set, the ip item is optional)
// port: Port number for the NHP server to operate (if set to 0, the default port 62206 will be used)
// expire: Expiration time of the NHP server's public key (in epoch seconds, set to 0 for permanent)
//
// Return:
// Whether the server information has been successfully added.
func NhpAgentAddServer(pubkey string, ip string, host string, port int, expire int64) bool {
	if gAgentInstance == nil {
		return false
	}

	if len(pubkey) == 0 || (len(ip) == 0 && len(host) == 0) {
		return false
	}

	serverPort := int(port)
	if serverPort == 0 {
		serverPort = 62206 // use default server listening port
	}

	serverPeer := &core.UdpPeer{
		Type:         core.NHP_SERVER,
		PubKeyBase64: pubkey,
		Ip:           ip,
		Port:         serverPort,
		Hostname:     host,
		ExpireTime:   expire,
	}
	gAgentInstance.AddServer(serverPeer)
	return true
}

// Delete NHP server information from the agent
//
// Input:
// pubkey: NHP server public key
func NhpAgentRemoveServer(pubkey string) {
	if gAgentInstance == nil {
		return
	}
	if len(pubkey) == 0 {
		return
	}

	gAgentInstance.RemoveServer(pubkey)
}

// Please add a resource information for the agent to use for knocking on the door
// (the agent can initiate a knock-on request for different resources)
//
// Input:
// aspId: Authentication Service Provider Identifier
// resId: Resource Identifier
// serverIp: NHP server IP address or domain name (the NHP server managing the resource)
// serverHostname: NHP server domain name (the NHP server managing the resource)
// serverPort: NHP server port (the NHP server managing the resource)
//
// Return:
// Whether the resource information has been added successfully
func NhpAgentAddResource(aspId string, resId string, serverIp string, serverHostname string, serverPort int) bool {
	if gAgentInstance == nil {
		return false
	}

	if len(aspId) == 0 || len(resId) == 0 || (len(serverIp) == 0 && len(serverHostname) == 0) {
		return false
	}

	resource := &agent.KnockResource{
		AuthServiceId:  aspId,
		ResourceId:     resId,
		ServerIp:       serverIp,
		ServerHostname: serverHostname,
		ServerPort:     serverPort,
	}
	err := gAgentInstance.AddResource(resource)
	return err == nil
}

// Delete resource information from the agent
//
// Input:
// aspId: Authentication Service Provider Identifier
// resId: Resource Identifier
func NhpAgentRemoveResource(aspId string, resId string) {
	if gAgentInstance == nil {
		return
	}

	if len(aspId) == 0 || len(resId) == 0 {
		return
	}

	gAgentInstance.RemoveResource(aspId, resId)
}

// The agent initiates a single knock on the door request to the server hosting the resource
//
// Input:
// aspId: Authentication service provider identifier
// resId: Resource identifier
// serverIp: NHP server IP address or domain name (the NHP server managing the resource)
// serverHostname: NHP server domain name (the NHP server managing the resource)
// serverPort: NHP server port (the NHP server managing the resource)
//
// Returns:
// The server's response message (json format string buffer pointer):
// "errCode": Error code (string, "0" indicates success)
// "errMsg": Error message (string)
// "resHost": Resource server address ("resHost": {"Server Name 1":"Server Hostname 1", "Server Name 2":"Server Hostname 2", ...})
// "opnTime": Door opening duration (integer, in seconds)
// "aspToken": Token generated after authentication by the ASP (optional)
// "agentAddr": Agent's IP address from the perspective of the NHP server
// "preActs": Pre-connection information related to the resource (optional)
// "redirectUrl": HTTP redirection link (optional)
//
// It is necessary to call NhpAgentAddServer before calling,
// to add the NHP server's public key, address, and other information to the agent
// The caller is responsible for calling NhpFreeCstring to release the returned char* pointer
func NhpAgentKnockResource(aspId string, resId string, serverIp string, serverHostname string, serverPort int) string {
	ackMsg := &common.ServerKnockAckMsg{}

	func() {
		if gAgentInstance == nil {
			ackMsg.ErrCode = common.ErrNoAgentInstance.ErrorCode()
			ackMsg.ErrMsg = common.ErrNoAgentInstance.Error()
			return
		}

		if len(aspId) == 0 || len(resId) == 0 || (len(serverIp) == 0 && len(serverHostname) == 0) {
			ackMsg.ErrCode = common.ErrInvalidInput.ErrorCode()
			ackMsg.ErrMsg = common.ErrInvalidInput.Error()
			return
		}

		resource := &agent.KnockResource{
			AuthServiceId:  aspId,
			ResourceId:     resId,
			ServerIp:       serverIp,
			ServerHostname: serverHostname,
			ServerPort:     serverPort,
		}

		peer := gAgentInstance.FindServerPeerFromResource(resource)
		if peer == nil {
			ackMsg.ErrCode = common.ErrKnockServerNotFound.ErrorCode()
			ackMsg.ErrMsg = common.ErrKnockServerNotFound.Error()
			return
		}

		target := &agent.KnockTarget{
			KnockResource: *resource,
			ServerPeer:    peer,
		}

		ackMsg, _ = gAgentInstance.Knock(target)
	}()

	bytes, _ := json.Marshal(ackMsg)

	return string(bytes)
}

// The agent explicitly informs the NHP server to exit its access permission to the resource.
//
// Input:
// aspId: Authentication Service Provider Identifier
// resId: Resource Identifier
// serverIp: NHP server IP address or domain name (the NHP server managing the resource)
// serverHostname: NHP server domain name (the NHP server managing the resource)
// serverPort: NHP server port (the NHP server managing the resource)
//
// Return:
// Whether the exit was successful
//
// It is necessary to call NhpAgentAddServer before calling, to add the NHP server's public key, address, and other information to the
func NhpAgentExitResource(aspId string, resId string, serverIp string, serverHostname string, serverPort int) bool {
	var err error
	ackMsg := &common.ServerKnockAckMsg{}

	func() {
		if gAgentInstance == nil {
			ackMsg.ErrCode = common.ErrNoAgentInstance.ErrorCode()
			ackMsg.ErrMsg = common.ErrNoAgentInstance.Error()
			err = common.ErrNoAgentInstance
			return
		}

		if len(aspId) == 0 || len(resId) == 0 || (len(serverIp) == 0 && len(serverHostname) == 0) {
			ackMsg.ErrCode = common.ErrInvalidInput.ErrorCode()
			ackMsg.ErrMsg = common.ErrInvalidInput.Error()
			err = common.ErrInvalidInput
			return
		}

		resource := &agent.KnockResource{
			AuthServiceId:  aspId,
			ResourceId:     resId,
			ServerIp:       serverIp,
			ServerHostname: serverHostname,
			ServerPort:     serverPort,
		}

		peer := gAgentInstance.FindServerPeerFromResource(resource)
		if peer == nil {
			ackMsg.ErrCode = common.ErrKnockServerNotFound.ErrorCode()
			ackMsg.ErrMsg = common.ErrKnockServerNotFound.Error()
			err = common.ErrKnockServerNotFound
			return
		}

		target := &agent.KnockTarget{
			KnockResource: *resource,
			ServerPeer:    peer,
		}

		ackMsg, err = gAgentInstance.ExitKnockRequest(target)
	}()

	return err == nil
}

// cipherType: 0-curve25519; 1-sm2
// result: "privatekey"|"publickey"
// caller is responsible to free the returned char* pointer
//
//export NhpGenerateKeys
func NhpGenerateKeys(cipherType int) string {
	var e core.Ecdh
	switch core.EccTypeEnum(cipherType) {
	case core.ECC_SM2:
		e = core.NewECDH(core.ECC_SM2)
	case core.ECC_CURVE25519:
		fallthrough
	default:
		e = core.NewECDH(core.ECC_CURVE25519)
	}
	pub := e.PublicKeyBase64()
	priv := e.PrivateKeyBase64()

	res := fmt.Sprintf("%s|%s", priv, pub)

	return res
}

// cipherType: 0-curve25519; 1-sm2
// privateBase64: private key in base64 format
// result: "publickey"
// caller is responsible to free the returned char* pointer
//
//export NhpPrivkeyToPubkey
func NhpPrivkeyToPubkey(cipherType int, privateBase64 string) string {
	privKey := privateBase64
	privKeyBytes, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		return ""
	}

	e := core.ECDHFromKey(core.EccTypeEnum(cipherType), privKeyBytes)
	if e == nil {
		return ""
	}
	pub := e.PublicKeyBase64()

	return pub
}
