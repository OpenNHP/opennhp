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
	ACId       string      `json:"acId"`
	Hostname   string      `json:"host,omitempty"` // hostname, optional
	Addr       *NetAddress `json:"addr"`           // dst ip + port + protocol
	PortSuffix bool        `json:"portSuffix,omitempty"`
	MaskHost   bool        `json:"maskHost,omitempty"` // do not reveal resource host in ack info
}

func (r *ResourceInfo) DestHost() string {
	if r.MaskHost || r.Addr == nil {
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

type DRGMsg struct {
	DoType         string      `json:"doType"`         // Data object format type, default "ZTDO" (ZTDO format details in Chapter 8). Custom formats allowed.
	DoId           string      `json:"doId"`           // Globally unique data object identifier (typically UUID)
	DbId           string      `json:"dbId"`           // Data broker identifier
	DataSourceType string      `json:"dataSourceType"` // Data source type, default "file", Supported Values: file, stream
	AccessUrl      string      `json:"accessUrl"`      // Data access URL (empty indicates offline transfer)
	AccessByNHP    bool        `json:"accessByNHP"`    // Require NHP handshake before accessing URL (optional if accessUrl empty)
	Spo            SmartPolicy `json:"spo"`
}

type DAKMsg struct {
	DoId    string `json:"doId"`    // Echoes registration request's DoId
	ErrCode int    `json:"errCode"` // Registration error code (0=success)
	ErrMsg  string `json:"errMsg"`  // Error message (empty if success)
}

type DARMsg struct {
	DoId                       string `json:"doId"`                       // Requested data object identifier
	UserId                     string `json:"userId"`                     // User identifier
	TeePublicKey               string `json:"teePublicKey"`               // Base64-encoded TEE (Trusted Execution Environment) public key
	ConsumerEphemeralPublicKey string `json:"consumerEphemeralPublicKey"` // Base64-encoded consumer ephemeral public key
}

type DAGMsg struct {
	DoId           string          `json:"doId"`                     // Echoes request's DoId
	DoType         string          `json:"doType,omitempty"`         // Echoes request's DoType
	DataSourceType string          `json:"dataSourceType,omitempty"` // Data source type, the default value is online, and supported values are online, offline and stream.
	AccessUrl      string          `json:"accessUrl,omitempty"`      // Data access URL
	AccessByNHP    bool            `json:"accessByNHP,omitempty"`    // Indicates whether to grant access to the data through NHP
	Kao            *KeyAccessObject `json:"kao,omitempty"`            // Key access object
	Spo            *SmartPolicy     `json:"spo,omitempty"`            // Smart policy Object
	ErrCode        int             `json:"errCode"`                  // Registration error code (0=success)
	ErrMsg         string          `json:"errMsg"`                   // Error message (empty if success)
}

type DWRMsg struct {
	DoId                       string `json:"doId"`                       // Data object identifier
	TeePublicKey               string `json:"teePublicKey"`               // Based64 encoded TEE public key
	ConsumerEphemeralPublicKey string `json:"consumerEphemeralPublicKey"` // Based64 encoded consumer ephemeral public key
}

type DWAMsg struct {
	DoId    string          `json:"doId"`          // Data object identifier
	Kao     *KeyAccessObject `json:"kao,omitempty"` // Key access object
	ErrCode int             `json:"errCode"`       // Registration error code (0=success)
	ErrMsg  string          `json:"errMsg"`        // Error message (empty if success)
}

type DSAMsg struct {
	DoId  string      `json:"doId"`          // Data object identifier
	SpoId string      `json:"spoId"`         // Smart Policy Object identifier
	Spo   *SmartPolicy `json:"spo,omitempty"` // Smart policy Object
	TTL   int         `json:"TTL"`           // Evidence validity period in milliseconds
}

type DAVMsg struct {
	DoId     string `json:"doId"`     // Data object identifier
	SpoId    string `json:"spoId"`    // Smart Policy Object identifier
	Evidence string `json:"evidence"` // Policy verification evidence
}

type KeyAccessObject struct {
	WrappedDataKey string `json:"wrappedDataKey"`          // Wrapped data private key
	SpoId          string `json:"spoId,omitempty"`         // SPO identifier
	PolicyBinding  string `json:"policyBinding,omitempty"` // Base64-encoded HMAC(HMAC(pao), key) using payload key
}

type SmartPolicy struct {
	PolicyId string `json:"policyId"` // Policy identifier
	Policy   string `json:"policy"`   // Base64-encoded wasm policy
}

type DBOnlineMsg struct {
	DBId string `json:"dbId,omitempty"`
}

type ServerDBAckMsg struct {
	ErrCode string `json:"errCode"`
	ErrMsg  string `json:"errMsg,omitempty"`
	DBAddr  string `json:"dbAddr"`
}
