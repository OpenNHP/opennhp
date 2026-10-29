package core

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// Sentinel causes returned (wrapped) by UdpPeer.SendAddrErr so callers can
// tell a name-resolution failure apart from an unparseable literal IP.
// These are deliberately plain errors rather than entries in the numbered
// core.Error registry: that registry's SetExtraError mutates a shared
// singleton, which cannot safely carry a per-call DNS cause.
var (
	// ErrPeerHostResolve: the peer is configured by Host (DNS name), the
	// lookup failed, and no usable static Ip fallback exists.
	ErrPeerHostResolve = errors.New("peer host resolution failed")
	// ErrPeerInvalidIp: no Host is configured (or none is needed) and the
	// literal Ip cannot be parsed.
	ErrPeerInvalidIp = errors.New("peer ip cannot be parsed")
)

type Peer interface {
	DeviceType() int
	Name() string
	PublicKey() []byte
	PublicKeyBase64() string

	IsExpired() bool

	ResolveHost() string
	ResolvedIps() []string
	Host() string
	SendAddr() net.Addr
	LastSendTime() int64
	UpdateSend(currTime int64)

	// LastRecvTime is the timestamp of the most recent successfully validated
	// packet from this peer (across all connections). Used by AC/DB discovery
	// loops to decide when to re-register. Source-address stickiness is no
	// longer a peer concern; see ConnectionData.CheckRecvAddress.
	LastRecvTime() int64
	UpdateRecv(currTime int64)
}

type UdpPeer struct {
	sync.Mutex

	// immutable fields. Don't change them after creation, so no lock is required
	PubKeyBase64 string `json:"pubKeyBase64"`
	Hostname     string `json:"host,omitempty"`
	Ip           string `json:"ip"` // static ip, it may be different from primaryResolvedIp
	Port         int    `json:"port"`
	Type         int    `json:"type"`
	ExpireTime   int64  `json:"expireTime"`
	name         string
	pubKey       []byte

	// mutable fields
	lastSendTime                     int64
	lastRecvTime                     int64
	lastNSLookupTime                 int64
	lastNSLookupErr                  error // most recent net.LookupHost failure; nil once a lookup succeeds
	resolvedIpArr                    []string
	primaryResolvedIp                string
	teePublicKeyBase64               string
	consumerEphemeralPublicKeyBase64 string
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
		// Safely handle short or empty PubKeyBase64 strings
		if len(p.PubKeyBase64) >= 43 {
			p.name = fmt.Sprintf("%s...%s", p.PubKeyBase64[0:4], p.PubKeyBase64[39:43])
		} else if len(p.PubKeyBase64) > 0 {
			// For short keys, use what we have
			p.name = p.PubKeyBase64
		} else {
			p.name = "unknown"
		}
	}
	return p.name
}

func (p *UdpPeer) ResolveHost() string {
	if len(p.Hostname) == 0 {
		return p.Ip
	}

	p.Lock()
	defer p.Unlock()

	currTime := time.Now().UnixNano()
	if currTime-p.lastNSLookupTime > MinimalNSLookupInterval*int64(time.Second) {
		addrs, err := net.LookupHost(p.Hostname)
		if err == nil && len(addrs) > 0 {
			p.lastNSLookupTime = currTime
			p.lastNSLookupErr = nil
			p.resolvedIpArr = addrs
			p.primaryResolvedIp = addrs[0]
		} else {
			// Remember the cause so SendAddrErr can report a DNS failure
			// instead of a misleading "IP cannot be parsed". The lookup
			// timer is intentionally NOT advanced on failure, preserving
			// the existing retry-on-next-send behavior.
			if err == nil {
				err = fmt.Errorf("lookup %s: no addresses returned", p.Hostname)
			}
			p.lastNSLookupErr = err
		}
	}

	if len(p.primaryResolvedIp) > 0 {
		return p.primaryResolvedIp
	}
	return p.Ip
}

func (p *UdpPeer) Host() string {
	hostAddr := p.Ip
	if len(p.Hostname) > 0 {
		hostAddr = p.Hostname
	}
	if p.Port == 0 {
		return hostAddr
	}
	return fmt.Sprintf("%s:%d", hostAddr, p.Port)
}

// lastResolveErr returns the most recent DNS lookup failure for this
// peer's Hostname, or nil if the last lookup succeeded.
func (p *UdpPeer) lastResolveErr() error {
	p.Lock()
	defer p.Unlock()

	return p.lastNSLookupErr
}

// SendAddrErr is SendAddr with the failure cause. When the peer is
// configured by Host and the name cannot be resolved (with no usable
// static Ip to fall back on), the error wraps ErrPeerHostResolve and
// names the host; an unparseable literal Ip wraps ErrPeerInvalidIp.
// Callers that log or surface the failure should prefer this over
// SendAddr so a DNS problem is never reported as an IP-parsing one.
func (p *UdpPeer) SendAddrErr() (net.Addr, error) {
	resolvedIp := p.ResolveHost() // happens only when MinimalNSLookupInterval has passed
	if ip := net.ParseIP(resolvedIp); ip != nil {
		return &net.UDPAddr{
			IP:   ip,
			Port: p.Port,
		}, nil
	}

	if len(p.Hostname) > 0 {
		if err := p.lastResolveErr(); err != nil {
			return nil, fmt.Errorf("%w: cannot resolve host %q: %v", ErrPeerHostResolve, p.Hostname, err)
		}
	}
	return nil, fmt.Errorf("%w: %q", ErrPeerInvalidIp, resolvedIp)
}

// SendAddr returns the peer's UDP destination, or nil if it cannot be
// determined. See SendAddrErr for the reason.
func (p *UdpPeer) SendAddr() net.Addr {
	addr, _ := p.SendAddrErr()
	return addr
}

func (p *UdpPeer) ResolvedIps() []string {
	p.Lock()
	defer p.Unlock()

	return p.resolvedIpArr
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

func (p *UdpPeer) LastRecvTime() int64 {
	p.Lock()
	defer p.Unlock()

	return p.lastRecvTime
}

func (p *UdpPeer) UpdateRecv(currTime int64) {
	p.Lock()
	defer p.Unlock()

	p.lastRecvTime = currTime
}

func (p *UdpPeer) TeePublicKeyBase64() string {
	p.Lock()
	defer p.Unlock()

	return p.teePublicKeyBase64
}

func (p *UdpPeer) SetTeePublicKeyBase64(teePublicKeyBase64 string) {
	p.Lock()
	defer p.Unlock()

	p.teePublicKeyBase64 = teePublicKeyBase64
}

func (p *UdpPeer) ConsumerEphemeralPublicKeyBase64() string {
	p.Lock()
	defer p.Unlock()

	return p.consumerEphemeralPublicKeyBase64
}

func (p *UdpPeer) SetConsumerEphemeralPublicKeyBase64(consumerEphemeralPublicKeyBase64 string) {
	p.Lock()
	defer p.Unlock()

	p.consumerEphemeralPublicKeyBase64 = consumerEphemeralPublicKeyBase64
}
