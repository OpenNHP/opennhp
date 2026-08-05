package server

import (
	"container/list"
	"math"
	"net"
	"sync"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// ipRateLimiter is a bounded, IP-keyed token bucket for structurally valid
// packets. It is intentionally separate from rknRateLimiter, whose narrower
// overload-only budget protects the expensive RKN cookie-verification path.
type ipRateLimiter struct {
	mu sync.Mutex

	nanosPerToken int64
	burstNanos    int64
	idleNanos     int64
	maxEntries    int
	buckets       map[string]*ipRateBucket
	lru           list.List
}

type ipRateBucket struct {
	ip             string
	allowanceNanos int64
	lastSeenNanos  int64
	elem           *list.Element
}

func newIPRateLimiter(ratePerSecond, burst, maxEntries int, idleNanos int64) *ipRateLimiter {
	if ratePerSecond <= 0 {
		ratePerSecond = 1
	}
	if burst <= 0 {
		burst = 1
	}
	if maxEntries <= 0 {
		maxEntries = 1
	}
	nanosPerToken := int64(1_000_000_000) / int64(ratePerSecond)
	if nanosPerToken <= 0 {
		nanosPerToken = 1
	}
	return &ipRateLimiter{
		nanosPerToken: nanosPerToken,
		burstNanos:    nanosPerToken * int64(burst),
		idleNanos:     idleNanos,
		maxEntries:    maxEntries,
		buckets:       make(map[string]*ipRateBucket),
	}
}

func (r *ipRateLimiter) allow(ip string, nowNanos int64) bool {
	if ip == "" {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	b := r.buckets[ip]
	if b != nil && r.idleNanos > 0 && nowNanos-b.lastSeenNanos > r.idleNanos {
		r.remove(b)
		b = nil
	}
	if b == nil {
		if len(r.buckets) >= r.maxEntries {
			oldest := r.lru.Back()
			if oldest != nil {
				r.remove(oldest.Value.(*ipRateBucket))
			}
		}
		// A fresh IP starts at half burst. This admits normal knock bursts while
		// halving the amplification available to source-IP rotation.
		b = &ipRateBucket{
			ip:             ip,
			allowanceNanos: r.burstNanos / 2,
			lastSeenNanos:  nowNanos,
		}
		b.elem = r.lru.PushFront(b)
		r.buckets[ip] = b
	} else {
		elapsed := nowNanos - b.lastSeenNanos
		if elapsed > 0 {
			b.allowanceNanos += elapsed
			if b.allowanceNanos > r.burstNanos {
				b.allowanceNanos = r.burstNanos
			}
		}
		b.lastSeenNanos = nowNanos
		r.lru.MoveToFront(b.elem)
	}

	if b.allowanceNanos < r.nanosPerToken {
		return false
	}
	b.allowanceNanos -= r.nanosPerToken
	return true
}

func (r *ipRateLimiter) remove(b *ipRateBucket) {
	delete(r.buckets, b.ip)
	r.lru.Remove(b.elem)
}

func (r *ipRateLimiter) len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.buckets)
}

// preCheckThreatCache bounds malformed-packet counters and collapses source
// port rotation into one IP entry. It uses opportunistic TTL expiration and
// strict LRU eviction, so no background goroutine is required.
type preCheckThreatCache struct {
	mu         sync.Mutex
	maxEntries int
	idleNanos  int64
	entries    map[string]*preCheckThreat
	lru        list.List
}

type preCheckThreat struct {
	ip            string
	count         int32
	lastSeenNanos int64
	elem          *list.Element
}

func newPreCheckThreatCache(maxEntries int, idleNanos int64) *preCheckThreatCache {
	if maxEntries <= 0 {
		maxEntries = 1
	}
	return &preCheckThreatCache{
		maxEntries: maxEntries,
		idleNanos:  idleNanos,
		entries:    make(map[string]*preCheckThreat),
	}
}

func (c *preCheckThreatCache) increment(ip string, nowNanos int64) int32 {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := c.entries[ip]
	if e != nil && c.idleNanos > 0 && nowNanos-e.lastSeenNanos > c.idleNanos {
		c.remove(e)
		e = nil
	}
	if e == nil {
		if len(c.entries) >= c.maxEntries {
			if oldest := c.lru.Back(); oldest != nil {
				c.remove(oldest.Value.(*preCheckThreat))
			}
		}
		e = &preCheckThreat{ip: ip}
		e.elem = c.lru.PushFront(e)
		c.entries[ip] = e
	} else {
		c.lru.MoveToFront(e.elem)
	}
	if e.count < math.MaxInt32 {
		e.count++
	}
	e.lastSeenNanos = nowNanos
	return e.count
}

func (c *preCheckThreatCache) clear(ip string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e := c.entries[ip]; e != nil {
		c.remove(e)
	}
}

func (c *preCheckThreatCache) remove(e *preCheckThreat) {
	delete(c.entries, e.ip)
	c.lru.Remove(e.elem)
}

func (c *preCheckThreatCache) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

func (s *UdpServer) allowPacketFromIP(ip string, nowNanos int64) bool {
	// Start initializes the limiter before either receive path runs. A nil
	// limiter is accepted only for focused handler tests that construct a
	// partial UdpServer without calling Start.
	return s.packetLimiter == nil || s.packetLimiter.allow(ip, nowNanos)
}

// isKnownRelayPeerIP prevents a busy configured relay from sharing one outer
// transport bucket across all of its clients. Unknown sources cannot claim an
// NHP_RLY header to bypass the general limiter; their packets remain charged
// to the source IP until cryptographic relay identity validation.
func (s *UdpServer) isKnownRelayPeerIP(ip string) bool {
	s.relayPeerMapMutex.Lock()
	defer s.relayPeerMapMutex.Unlock()
	for _, peer := range s.relayPeerMap {
		if peer.Ip == ip {
			return true
		}
		for _, resolved := range peer.ResolvedIps() {
			if resolved == ip {
				return true
			}
		}
	}
	return false
}

// isAuthenticatedControlPlaneAddr recognizes an AC/DB transport tuple only
// after HandleACOnline/HandleDBOnline has validated its cryptographic identity
// and registered it. This avoids throttling high-fan-in control traffic without
// trusting attacker-controlled AOL/DOL header bytes or requiring IPs in the
// shipped public-key-only peer configuration.
func (s *UdpServer) isAuthenticatedControlPlaneAddr(addr *net.UDPAddr) bool {
	if addr == nil {
		return false
	}
	key := addr.String()
	s.acConnectionMapMutex.Lock()
	for _, conn := range s.acConnectionMap {
		if conn != nil && conn.ConnData != nil && conn.ConnData.RemoteAddr != nil && conn.ConnData.RemoteAddr.String() == key {
			s.acConnectionMapMutex.Unlock()
			return true
		}
	}
	s.acConnectionMapMutex.Unlock()

	s.dbConnectionMapMutex.Lock()
	defer s.dbConnectionMapMutex.Unlock()
	for _, conn := range s.dbConnectionMap {
		if conn != nil && conn.ConnData != nil && conn.ConnData.RemoteAddr != nil && conn.ConnData.RemoteAddr.String() == key {
			return true
		}
	}
	return false
}

func blockAddressForConnection(conn *UdpConn) *net.UDPAddr {
	if conn != nil && conn.ConnData != nil && conn.ConnData.RealRemoteAddr != nil {
		return conn.ConnData.RealRemoteAddr
	}
	if conn == nil || conn.ConnData == nil {
		return nil
	}
	return conn.ConnData.RemoteAddr
}

func (s *UdpServer) logPacketRateLimitDrop(ip string) {
	drops := s.packetRateLimitDrops.Add(1)
	if drops == 1 || drops%1000 == 0 {
		log.Warning("packet from %s dropped: per-IP rate limit exceeded (total drops: %d)", ip, drops)
	}
}

func blockIPKey(addr *net.UDPAddr) string {
	if addr == nil || addr.IP == nil {
		return ""
	}
	return addr.IP.String()
}
