package common

import (
	"fmt"
)

type NetAddress struct {
	Ip       string `json:"ip"`              // IP address, mandatory
	Port     int    `json:"port,omitempty"`  // optional
	Protocol string `json:"proto,omitempty"` // tcp/udp/empty for any optional
}

func (na *NetAddress) String() string {
	if na.Port == 0 {
		return na.Ip
	}
	return fmt.Sprintf("%s:%d", na.Ip, na.Port)
}

// agent <-> server
type ServerCookieMsg struct {
	TransactionId uint64 `json:"trxId"`
	Cookie        string `json:"cookie"`
}

type AgentOTPMsg struct {
	UserId         string         `json:"usrId"`
	DeviceId       string         `json:"devId"`
	OrganizationId string         `json:"orgId,omitempty"`
	AuthServiceId  string         `json:"aspId"`
	Passcode       string         `json:"pass,omitempty"`
	UserData       map[string]any `json:"usrData,omitempty"`
}

type AgentRegisterMsg struct {
	UserId         string         `json:"usrId"`
	DeviceId       string         `json:"devId"`
	OrganizationId string         `json:"orgId,omitempty"`
	AuthServiceId  string         `json:"aspId"`
	OTP            string         `json:"otp,omitempty"`
	UserData       map[string]any `json:"usrData,omitempty"`
}

type ServerRegisterAckMsg struct {
	ErrCode       string `json:"errCode"`
	ErrMsg        string `json:"errMsg,omitempty"`
	AuthServiceId string `json:"aspId"`
}

type AgentKnockMsg struct {
	HeaderType     int            `json:"headerType"`
	UserId         string         `json:"usrId"`
	DeviceId       string         `json:"devId"`
	OrganizationId string         `json:"orgId,omitempty"`
	AuthServiceId  string         `json:"aspId"`
	ResourceId     string         `json:"resId"`
	CheckResults   map[string]any `json:"results,omitempty"`
	UserData       map[string]any `json:"usrData,omitempty"`
}

func (knkMsg *AgentKnockMsg) Id() string {
	return knkMsg.AuthServiceId + "/" + knkMsg.ResourceId
}

type PreAccessInfo struct {
	AccessIp       string `json:"acIp"`
	AccessPort     string `json:"acPort"`
	ACPubKey       string `json:"acPubKey"`
	ACToken        string `json:"acToken"`
	ACCipherScheme int    `json:"acCipherScheme"`
}

type ServerKnockAckMsg struct {
	ErrCode           string                    `json:"errCode"`
	ErrMsg            string                    `json:"errMsg,omitempty"`
	ResourceHost      map[string]string         `json:"resHost"`
	OpenTime          uint32                    `json:"opnTime"`
	AuthProviderToken string                    `json:"aspToken,omitempty"` // optional for ac backend validation
	AgentAddr         string                    `json:"agentAddr"`
	ACTokens          map[string]string         `json:"acTokens"`
	PreAccessActions  map[string]*PreAccessInfo `json:"preActions,omitempty"` // optional for pre-access
	RedirectUrl       string                    `json:"redirectUrl,omitempty"`
}

type AgentListMsg struct {
	UserId         string         `json:"usrId"`
	DeviceId       string         `json:"devId"`
	OrganizationId string         `json:"orgId,omitempty"`
	AuthServiceId  string         `json:"aspId"`
	UserData       map[string]any `json:"usrData,omitempty"`
}

type ServerListResultMsg struct {
	ErrCode     string         `json:"errCode"`
	ErrMsg      string         `json:"errMsg,omitempty"`
	ListResults map[string]any `json:"list,omitempty"`
}

// agent <-> ac
type AgentAccessMsg struct {
	UserId         string         `json:"usrId"`
	DeviceId       string         `json:"devId"`
	OrganizationId string         `json:"orgId,omitempty"`
	ACToken        string         `json:"acToken"`
	UserData       map[string]any `json:"usrData,omitempty"`
}

type ACAccessAckMsg struct {
	ErrCode   string `json:"errCode"`
	ErrMsg    string `json:"errMsg,omitempty"`
	AgentAddr string `json:"agentAddr,omitempty"` // optional
}

// ac <-> server
type ServerACOpsMsg struct {
	UserId           string        `json:"usrId"`
	DeviceId         string        `json:"devId"`
	OrganizationId   string        `json:"orgId,omitempty"`
	AuthServiceId    string        `json:"aspId"`
	ResourceId       string        `json:"resId"`
	SourceAddrs      []*NetAddress `json:"srcAddrs"`
	DestinationAddrs []*NetAddress `json:"dstAddrs"`
	OpenTime         uint32        `json:"opnTime"`
}

type ACOpsResultMsg struct {
	ErrCode         string         `json:"errCode"`
	ErrMsg          string         `json:"errMsg,omitempty"`
	OpenTime        uint32         `json:"opnTime"`
	ACToken         string         `json:"token"`
	PreAccessAction *PreAccessInfo `json:"preAct"`
}

type ACOnlineMsg struct {
	AuthServiceId string   `json:"aspId"`
	ResourceIds   []string `json:"resIds"`
	ACId          string   `json:"acId,omitempty"`
}

type ACRefreshMsg struct {
	NhpToken   string      `json:"nhpToken"`
	SourceAddr *NetAddress `json:"srcAddr"`
}

type ServerACAckMsg struct {
	ErrCode string `json:"errCode"`
	ErrMsg  string `json:"errMsg,omitempty"`
	ACAddr  string `json:"acAddr"`
}

type ResourceInfo struct {
	ACId       string
	Hostname   string      `json:"host,omitempty"` // hostname, optional
	Addr       *NetAddress `json:"addr"`           // dst ip + port + protocol
	PortSuffix bool        `json:"portSuffix,omitempty"`
}

func (r *ResourceInfo) DestHost() string {
	if r.Addr == nil {
		return ""
	}

	host := r.Addr.Ip
	if len(r.Hostname) > 0 {
		host = r.Hostname
	}
	if !r.PortSuffix || r.Addr.Port == 0 {
		return host
	}
	return fmt.Sprintf("%s:%d", host, r.Addr.Port)
}

func (r *ResourceInfo) DstIp() string {
	if r.Addr == nil {
		return ""
	}
	return r.Addr.Ip
}

type ResourceGroup struct {
	AuthServiceId     string                   `json:"aspId"`
	ResourceId        string                   `json:"resId"`
	OpenTime          uint32                   `json:"opnTime,omitempty"`
	AuthProviderToken string                   `json:"aspToken,omitempty"`
	Resources         map[string]*ResourceInfo `json:"resInfo"`
}

func (r *ResourceGroup) Id() string {
	return r.AuthServiceId + "/" + r.ResourceId
}

func (r *ResourceGroup) Hosts() map[string]string {
	hostMap := make(map[string]string)
	for name, info := range r.Resources {
		hostMap[name] = info.DestHost()
	}

	return hostMap
}

// // DHP Msg structs
// 7.2.1.NHP_DRG (DHP Register)
type DRGMsg struct {
	DoType      string `json:"doType"`      // Data object format type, default "ZTDO" (ZTDO format details in Chapter 8). Custom formats allowed.
	DoId        string `json:"doId"`        // Globally unique data object identifier (typically UUID)
	AccessUrl   string `json:"accessUrl"`   // Data access URL (empty indicates offline transfer)
	AccessByNHP bool   `json:"accessByNHP"` // Require NHP handshake before accessing URL (optional if accessUrl empty)
	AspHost     string `json:"aspHost"`     // ASP authorization service provider address (KAS/PAS services)
	KasType     int    `json:"kasType"`     // KAS type: 0=KAS on NHP-server (default), 1=KAS on ASP
	KaoContent  string `json:"kaoContent"`  // KAO JSON data when kasType=0 (see 7.3). Empty otherwise.
	PasType     int    `json:"pasType"`     // PAS type: 0=PAS on NHP-server (default), 1=PAS on ASP
	PaoContent  string `json:"paoContent"`  // PAO content (REGO policy) when pasType=0 (see 7.4). Empty otherwise.
}

// 7.2.2.NHP_DAK (DHP Register Ack)
type DAKMsg struct {
	DoId    string `json:"doId"`    // Echoes registration request's DoId
	ErrCode int    `json:"errCode"` // Registration error code (0=success)
	ErrMsg  string `json:"errMsg"`  // Error message (empty if success)
}

// 7.2.3.NHP_DAR (DHP Access Request)
type DARMsg struct {
	DoId string `json:"doId"` // Requested data object identifier
}

// 7.2.4.NHP_DAG (DHP Access Granted)
type DAGMsg struct {
	DoId       string `json:"doId"`       // Echoes request's DoId
	ErrCode    int    `json:"errCode"`    // Authorization error code (0=success)
	ErrMsg     string `json:"errMsg"`     // Error message (empty if success)
	WrappedKey string `json:"wrappedKey"` // Base64-encoded symmetric key encrypted with data consumer's public key (empty on error)
}

// 7.2.5.NHP_DPC (DHP Policy Challenge)
type DPCMsg struct {
	DoId             string `json:"doId"`             // Data object identifier
	ChallengeId      string `json:"challengeId"`      // Challenge ID (must match corresponding NHP_DPC)
	ChallengeContent string `json:"challengeContent"` // Policy challenge content
	TTL              int    `json:"TTL"`              // Evidence validity period in milliseconds
}

// 7.2.6.NHP_DPV (DHP Policy Verification)
type DPVMsg struct {
	DoId        string `json:"doId"`        // Data object identifier
	ChallengeId string `json:"challengeId"` // Matching challenge ID
	Evidence    string `json:"evidence"`    // Policy verification evidence
	TTL         int    `json:"TTL"`         // Evidence validity period in milliseconds
}

// 7.3.KAO (Key Access Object)
type DHPKao struct {
	KeyWrapper    string `json:"keyWrapper"`    // Key wrapping method: "kas"=KAS public key, "consumer"=data consumer's public key
	PolicyBinding string `json:"policyBinding"` // Base64-encoded HMAC(HMAC(pao), key) using payload key
	ConsumerId    string `json:"ConsumerId"`    // Data consumer identifier (email/phone/etc)
	WrappedKey    string `json:"wrappedKey"`    // Base64-encoded payload key encrypted via keyWrapper
}

// DHP Policy
type DHPPolicy struct {
	ConsumerPublicKey string `json:"publicKey"`  // Data consumer's public key
	ConsumerId        string `json:"consumerId"` // Data consumer ID
}
