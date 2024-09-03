package common

import "net/url"

// authsvcprovider and resource
type LoginPageContext struct {
	Title              string `json:"title,omitempty"`
	ClientId           string `json:"clientId,omitempty"`
	AppKey             string `json:"appKey,omitempty"`
	AppSecret          string `json:"appSecret,omitempty"`
	RedirectUrl        string `json:"redirectUrl,omitempty"`
	RedirectWithParams bool   `json:"redirectWithParams,omitempty"`
}

type ResourceData struct {
	ResourceGroup `mapstructure:",squash"`
	// optional extension data
	AppKey             string         `json:"appKey,omitempty"`
	AppSecret          string         `json:"appSecret,omitempty"`
	AccessKey          string         `json:"accessKey,omitempty"`
	SecretKey          string         `json:"secretKey,omitempty"`
	ExInfo             map[string]any `json:"exinfo,omitempty"`
	RedirectUrl        string         `json:"redirectUrl,omitempty"`
	RedirectWithParams bool           `json:"redirectWithParams,omitempty"`
	SkipAuth           bool           `json:"skipAuth,omitempty"`
}

type ResourceGroupMap map[string]*ResourceData
type AuthServiceProviderData struct {
	ResourceGroups ResourceGroupMap `json:"ress"`
	AuthSvcId      string           `json:"aspId"`
	PluginPath     string           `json:"pluginPath,omitempty"`
	PluginHash     string           `json:"pluginHash,omitempty"`
}
type AuthSvcProviderMap map[string]*AuthServiceProviderData

// requests
type NhpOTPRequest struct {
	Msg     *AgentOTPMsg `json:"msg"`
	SrcAddr *NetAddress  `json:"srcAddr"`
}

type NhpRegisterRequest struct {
	Msg       *AgentRegisterMsg     `json:"msg"`
	Ack       *ServerRegisterAckMsg `json:"ack"`
	PublicKey string                `json:"pubKey"`
	SrcAddr   *NetAddress           `json:"srcAddr"`
}

type NhpAuthRequest struct {
	Msg       *AgentKnockMsg     `json:"msg"`
	Ack       *ServerKnockAckMsg `json:"ack"`
	PublicKey string             `json:"pubKey"`
	SrcAddr   *NetAddress        `json:"srcAddr"`
}

type NhpListRequest struct {
	Msg       *AgentListMsg        `json:"msg"`
	Ack       *ServerListResultMsg `json:"ack"`
	PublicKey string               `json:"pubKey"`
	SrcAddr   *NetAddress          `json:"srcAddr"`
}

type HttpKnockRequest struct {
	UserId         string   `json:"usrId"`
	DeviceId       string   `json:"devId"`
	OrganizationId string   `json:"orgId,omitempty"`
	AuthServiceId  string   `json:"aspId"`
	ResourceId     string   `json:"resId"`
	Token          string   `json:"token"`
	Code           string   `json:"code"`
	DstUrl         string   `json:"dstUrl"`
	Url            *url.URL `json:"-"`
	UserAgent      string   `json:"-"`
	SrcIp          string   `json:"-"`
	SrcPort        int      `json:"-"`
}
