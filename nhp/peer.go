package nhp

import (
	"encoding/base64"
	"fmt"
	"net"
	"sync"
	"time"
)

type Peer interface {
	DeviceType() int
	Name() string
	PublicKey() []byte
	PublicKeyBase64() string

	IsExpired() bool

	HostOrAddr() string
	SendAddr() net.Addr
	LastSendTime() int64
	UpdateSend(currTime int64)

	RecvAddr() net.Addr
	LastRecvTime() int64
	UpdateRecv(currTime int64, currAddr net.Addr)
	CheckRecvAddress(currTime int64, currAddr net.Addr) bool
}

type UdpPeer struct {
	sync.Mutex

	// immutable fields. Don't change them after creation
	PubKeyBase64 string `json:"pubKeyBase64"`
	Hostname     string `json:"host,omitempty"`
	Ip           string `json:"ip"`
	Port         int    `json:"port"`
	Type         int    `json:"type"`
	ExpireTime   int64  `json:"expireTime"`
	name         string
	pubKey       []byte

	// mutable fields
	lastSendTime     int64
	lastRecvTime     int64
	lastNSLookupTime int64
	resolvedIp       string
	recvAddr         *net.UDPAddr
}

func (p *UdpPeer) DeviceType() DeviceTypeEnum {
	return p.Type
}

func (p *UdpPeer) PublicKey() []byte {
	p.Lock()
	defer p.Unlock()

	if p.pubKey == nil {
		p.pubKey, _ = base64.StdEncoding.DecodeString(p.PubKeyBase64)
	}
	return p.pubKey
}

func (p *UdpPeer) PublicKeyBase64() string {
	return p.PubKeyBase64
}

func (p *UdpPeer) Name() string {
	p.Lock()
	defer p.Unlock()

	if len(p.name) == 0 {
		p.name = fmt.Sprintf("%s...%s", p.PubKeyBase64[0:4], p.PubKeyBase64[39:43])
	}
	return p.name
}

func (p *UdpPeer) HostOrAddr() string {
	if len(p.Hostname) > 0 {
		return p.Hostname
	}
	return p.Ip
}

func (p *UdpPeer) SendAddr() net.Addr {
	if len(p.Ip) == 0 && len(p.Hostname) == 0 {
		return nil
	}

	var ip net.IP
	if len(p.Hostname) > 0 {
		p.Lock()
		defer p.Unlock()

		currTime := time.Now().UnixNano()
		if currTime-p.lastNSLookupTime > MinimalNSLookupTime*int64(time.Second) {
			addrs, err := net.LookupHost(p.Hostname)
			if err != nil {
				return nil
			}

			p.lastNSLookupTime = currTime
			p.resolvedIp = addrs[0]
		}
		ip = net.ParseIP(p.resolvedIp)
	} else {
		ip = net.ParseIP(p.Ip)
	}

	if ip == nil {
		return nil
	}

	return &net.UDPAddr{
		IP:   ip,
		Port: p.Port,
	}
}

func (p *UdpPeer) ResolvedIp() string {
	p.Lock()
	defer p.Unlock()

	return p.resolvedIp
}

func (p *UdpPeer) CopyResolveStatus(other *UdpPeer) {
	other.Lock()
	ip := other.resolvedIp
	lastTs := other.lastNSLookupTime
	other.Unlock()

	p.Lock()
	defer p.Unlock()

	p.resolvedIp = ip
	p.lastNSLookupTime = lastTs
}

func (p *UdpPeer) IsExpired() bool {
	p.Lock()
	defer p.Unlock()

	// p.ExpireTime is in seconds
	return p.ExpireTime > 0 && time.Now().UnixMilli() > p.ExpireTime*1000
}

func (p *UdpPeer) LastSendTime() int64 {
	p.Lock()
	defer p.Unlock()

	return p.lastSendTime
}

func (p *UdpPeer) UpdateSend(currTime int64) {
	p.Lock()
	defer p.Unlock()

	p.lastSendTime = currTime
}

// a peer should not have multiple layer-4 addresses within its hold time
func (p *UdpPeer) CheckRecvAddress(currTime int64, currAddr net.Addr) bool {
	p.Lock()
	defer p.Unlock()

	if currTime > p.lastRecvTime+MinimalPeerAddressHoldTime*int64(time.Second) {
		return true
	}

	if p.recvAddr.String() == currAddr.String() {
		return true
	}

	return false
}

func (p *UdpPeer) RecvAddr() net.Addr {
	p.Lock()
	defer p.Unlock()

	return p.recvAddr
}

func (p *UdpPeer) LastRecvTime() int64 {
	p.Lock()
	defer p.Unlock()

	return p.lastRecvTime
}

func (p *UdpPeer) UpdateRecv(currTime int64, currAddr net.Addr) {
	p.Lock()
	defer p.Unlock()

	p.lastRecvTime = currTime
	p.recvAddr = currAddr.(*net.UDPAddr)
}
