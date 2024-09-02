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
	AccessIp   string `json:"acIp"`
	AccessPort string `json:"acPort"`
	ACPubKey   string `json:"acPubKey"`
	ACToken    string `json:"acToken"`
}

type ServerKnockAckMsg struct {
	ErrCode           string            `json:"errCode"`
	ErrMsg            string            `json:"errMsg,omitempty"`
	ResourceHost      map[string]string `json:"resHost"`
	OpenTime          uint32            `json:"opnTime"`
	AuthProviderToken string            `json:"aspToken,omitempty"` // optional for ac backend validation
	AgentAddr         string            `json:"agentAddr"`
	PreAccessActions  []*PreAccessInfo  `json:"preActs,omitempty"` // optional for pre-access
	RedirectUrl       string            `json:"redirectUrl,omitempty"`
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
	PreAccessAction *PreAccessInfo `json:"preAct"`
}

type ACOnlineMsg struct {
	AuthServiceId string   `json:"aspId"`
	ResourceIds   []string `json:"resIds"`
	ACId          string   `json:"acId,omitempty"`
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
