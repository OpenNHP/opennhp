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

// DHP相关消息结构
// 7.2.1.NHP_DRG (DHP Register)消息
type DRGMsg struct {
	DoType      string `json:"doType"`      //数据对象的格式类型，默认值为“ZTDO”，ZTDO的文件格式详细定义见第8章。除此之外，DHP协议还支持其他的自定义格式。
	DoId        string `json:"doId"`        //数据对象标识，要求全局唯一，通常为UUID
	AccessUrl   string `json:"accessUrl"`   //数据对象的访问URL，可以为空，表示通过线下方传输数据对象。
	AccessByNHP bool   `json:"accessByNHP"` //访问accessUrl前是否要先通过NHP敲门。如果accessUrl为空，该字段可省略。
	AspHost     string `json:"aspHost"`     //ASP授权服务提供商的地址，ASP提供密钥访问服务KAS以及策略验证服务PAS。
	KasType     int    `json:"kasType"`     //密钥访问服务KAS的类型：	0：密钥访问服务设置在NHP-server上，默认值; 1：密钥访问服务在ASP服务器上;
	KaoContent  string `json:"kaoContent"`  //当kasType为0本字段用来发送密钥访问对象(KAO)的JSON数据内容，KAO格式的详细定义见7.3章节; 	当kasType为其他值，本字段可以为空。
	PasType     int    `json:"pasType"`     //策略验证服务KAS的类型：0：策略验证服务设置在NHP-Server上，默认值；1：策略验证服务PAS在ASP服务器上；
	PaoContent  string `json:"paoContent"`  //当pasType为0本字段用来发送策略证明对象(PAO)的内容，策略类型为REGO，PAO格式的详细定义见7.4章节。当pasType为其他值，本字段可以为空。
}

// 7.2.2.NHP_DAK (DHP Register Ack)消息
type DAKMsg struct {
	DoId    string `json:"doId"`    //数据对象标识，此响应中标识为申请注册报文中数据对象标识
	ErrCode int    `json:"errCode"` //注册结果错误码，0则表示成功。
	ErrMsg  string `json:"errMsg"`  //错误提示信息，如果errCode为0则本字段为空。
}

// 7.2.3.NHP_DAR (DHP Access Request) 消息
type DARMsg struct {
	DoId string `json:"doId"` //请求访问的数据对象标识ID
}

// 7.2.4.NHP_DAG （DHP Access Granted）消息
type DAGMsg struct {
	DoId       string `json:"doId"`       //数据对象标识，此响应中标识为申请注册报文中数据对象标识
	ErrCode    int    `json:"errCode"`    //授权结果错误码，0表示成功
	ErrMsg     string `json:"errMsg"`     //错误提示信息，如果errCode为0则本字段为空。
	WrappedKey string `json:"wrappedKey"` //payload数据加密的对称密钥经数据对象的使用者的公钥加密后的密文再Base64编码数据。如果errCode不是0，则本字段为空。
}

// 7.2.5.NHP_DPC（DHP Policy Challenge）消息
type DPCMsg struct {
	DoId             string `json:"doId"`             //数据对象标识ID
	ChallengeId      string `json:"challengeId"`      //策略验证挑战的标识ID，必须和所回应的NHP_DPC消息中的challengeId一致。
	ChallengeContent string `json:"challengeContent"` //策略验证挑战的内容
	TTL              int    `json:"TTL"`              //策略验证证据的有效时间TTL（Time To Live），单位毫秒。
}

// 7.2.6.NHP_DPV（DHP Policy Verification）消息
type DPVMsg struct {
	DoId        string `json:"doId"`        //数据对象标识ID
	ChallengeId string `json:"challengeId"` //策略验证挑战的标识ID，必须和所回应的NHP_DPC消息中的challengeId一致。
	Evidence    string `json:"evidence"`    //策略验证的证据内容。
	TTL         int    `json:"TTL"`         //策略验证证据的有效时间TTL（Time To Live），单位毫秒。
}

// 7.3.密钥访问对象KAO
type DHPKao struct {
	KeyWrapper    string `json:"keyWrapper"`    //kas: 数据对象密钥采用KAS公钥加密封装生成；consumer: 数据对象密钥采用数据对向使用者的公钥加密封装生成；
	PolicyBinding string `json:"policyBinding"` //HMAC(HMAC(pao),key)的Base64编码，key采用payload数据加密对称密钥，HMAC(pao)是PAO对象的HMAC值
	ConsumerId    string `json:"ConsumerId"`    //数据访问请求者的标识ID，比如电子邮件、电话号码等
	WrappedKey    string `json:"wrappedKey"`    //payload数据加密的对称密钥经keyWrapper公钥加密后的密文再Base64编码数据。
}

// DHP 加密时Policy对象
type DHPPolicy struct {
	ConsumerPublicKey string `json:"publicKey"`  //数据访问者公钥
	ConsumerId        string `json:"consumerId"` //数据访问者用户Id
}
