package core

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/OpenNHP/opennhp/nhp/log"
)

type ConnectionData struct {
	// atomic data, keep 64bit(8-bytes) alignment for 32-bit system compatibility
	InitTime           int64 // local connection setup time. immutable after created
	LastRemoteSendTime int64
	LastLocalSendTime  int64
	LastLocalRecvTime  int64

	sync.Mutex
	sync.WaitGroup

	// common
	Device           *Device
	LocalAddr        *net.UDPAddr
	RemoteAddr       *net.UDPAddr
	RealRemoteAddr   *net.UDPAddr // set for relay-forwarded connections: the real client address behind the relay
	CookieStore      *CookieStore
	TimeoutMs        int
	SendQueue        chan *Packet
	RecvQueue        chan *Packet
	BlockSignal      chan struct{}
	SetTimeoutSignal chan struct{}
	StopSignal       chan struct{}

	closed atomic.Bool

	// remote transactions
	RemoteTransactionMutex sync.Mutex
	RemoteTransactionMap   map[uint64]*RemoteTransaction

	// stickyRecvMu guards stickyRecvAddr and stickyRecvTime. Recv stickiness
	// used to live on the peer (keyed by pubkey), but multiple connections can
	// share one pubkey (e.g. an AC fans out AOL to N nhp-server instances that
	// share an ECDH keypair). Sticking by pubkey would then bounce between
	// instances and falsely reject packets. Stickiness belongs to the
	// connection: source addr must stay stable within MinimalPeerAddressHoldTime.
	stickyRecvMu   sync.Mutex
	stickyRecvAddr *net.UDPAddr
	stickyRecvTime int64

	// specific
	RecvThreatCount int32
}

func (c *ConnectionData) Equal(other *ConnectionData) bool {
	// use nanosecond timestamp for comparison
	return c.InitTime == other.InitTime
	//return c.RemoteAddr.String() == other.RemoteAddr.String()
}

func (c *ConnectionData) SetTimeout(ms int) {
	c.TimeoutMs = ms
	c.SetTimeoutSignal <- struct{}{}
}

func (c *ConnectionData) Close() {
	if c.IsClosed() {
		return
	}

	// close all running transactions
	close(c.StopSignal)

	c.closed.Store(true)

	// flush connection remaining packet and close connection thread channels
flush:
	for {
		select {
		case pkt := <-c.SendQueue:
			c.Device.ReleasePoolPacket(pkt)
		case pkt := <-c.RecvQueue:
			c.Device.ReleasePoolPacket(pkt)
		case <-c.BlockSignal:
		default:
			break flush
		}
	}

	close(c.SendQueue)
	close(c.RecvQueue)
	close(c.BlockSignal)
	close(c.SetTimeoutSignal)
	c.SendQueue = nil
	c.RecvQueue = nil
	c.BlockSignal = nil
	c.SetTimeoutSignal = nil

	c.Wait()
}

func (c *ConnectionData) IsClosed() bool {
	return c.closed.Load()
}

func (c *ConnectionData) ForwardOutboundPacket(pkt *Packet) {
	if c.IsClosed() {
		log.Warning("connection %s is closed, discard outbound packet", c.RemoteAddr.String())
		c.Device.ReleasePoolPacket(pkt)
		return
	}

	select {
	case c.SendQueue <- pkt:
		// fully encrypted packet will be forwarded to higher level entity for physical sending
		// may block when send queue is full
	case <-c.StopSignal:
		// discard pending packets when connection is closed
		log.Warning("connection %s stopped, discard pending outbound packet", c.RemoteAddr.String())
		c.Device.ReleasePoolPacket(pkt)
	}
}

func (c *ConnectionData) ForwardInboundPacket(pkt *Packet) {
	if c.IsClosed() {
		log.Warning("connection %s is closed, discard inbound packet", c.RemoteAddr.String())
		c.Device.ReleasePoolPacket(pkt)
		return
	}

	select {
	case c.RecvQueue <- pkt:
		// raw packet will be forwarded to connection routine for packet parsing and decrytion
		// may block when recv queue is full
	case <-c.StopSignal:
		// discard pending packets when connection is closed
		log.Warning("connection %s stopped, discard pending inbound packet", c.RemoteAddr.String())
		c.Device.ReleasePoolPacket(pkt)
	}
}

// CheckRecvAddress enforces source-address stickiness on this connection:
// within MinimalPeerAddressHoldTime of the last accepted packet, currAddr
// must match the previously seen source. After the hold time elapses, a new
// address is accepted.
func (c *ConnectionData) CheckRecvAddress(currTime int64, currAddr net.Addr) bool {
	c.stickyRecvMu.Lock()
	defer c.stickyRecvMu.Unlock()

	if c.stickyRecvAddr == nil {
		return true
	}
	if currTime > c.stickyRecvTime+MinimalPeerAddressHoldTime*int64(time.Second) {
		return true
	}
	return c.stickyRecvAddr.String() == currAddr.String()
}

// UpdateRecvAddress records the latest accepted source address for this
// connection. Call only after CheckRecvAddress has approved currAddr.
func (c *ConnectionData) UpdateRecvAddress(currTime int64, currAddr net.Addr) {
	udpAddr, ok := currAddr.(*net.UDPAddr)
	if !ok {
		return
	}
	c.stickyRecvMu.Lock()
	c.stickyRecvAddr = udpAddr
	c.stickyRecvTime = currTime
	c.stickyRecvMu.Unlock()
}

func (c *ConnectionData) SendBlockSignal() {
	if c.IsClosed() {
		log.Warning("connection is closed, discard block signal")
		return
	}

	select {
	case c.BlockSignal <- struct{}{}:
		// trigger connection to close itself immediately and ask higher level entity to record the blocking connection
	default:
		log.Warning("old block signal not processed")
	}
}
