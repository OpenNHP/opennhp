package agent

import (
	"fmt"
	"net"

	"github.com/OpenNHP/opennhp/nhp/common/loadbalance"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// ServerInstance is one physical nhp-server endpoint backing a logical
// cluster identity. It embeds a *core.UdpPeer for noise handshake key
// material lookups (PublicKey, SendAddr, ExpireTime, LastSend/Recv
// bookkeeping), plus a Weight used by the load-balance picker.
//
// All instances of the same cluster share PubKeyBase64. They differ
// only by network address (Hostname + Ip + Port). The device peer map
// is keyed by pubkey, so we register exactly *one* representative peer
// per cluster on the device; per-instance addresses are looked up at
// send time via the cluster's Picker.
type ServerInstance struct {
	peer     *core.UdpPeer
	weight   int
	hostPort string // cached "host:port" for logging / sticky lookup
}

// Peer returns the noise/peer view of this instance — the same *UdpPeer
// every send path has used historically. Callers that need PublicKey,
// SendAddr, IsExpired, LastSendTime, UpdateSend/Recv use this.
func (si *ServerInstance) Peer() *core.UdpPeer { return si.peer }

// SendAddr resolves this instance's UDP destination. We can't simply
// delegate to si.peer.SendAddr() in the multi-instance world because
// every instance in a cluster shares one peer pubkey but has its own
// Hostname/Ip/Port — and the peer's resolution cache would otherwise
// flap between instance addresses. Each ServerInstance is its own
// *UdpPeer (same pubkey, distinct address), so si.peer.SendAddr() is
// safe and returns this instance's address.
func (si *ServerInstance) SendAddr() net.Addr { return si.peer.SendAddr() }

// HostPort returns "host:port" (or "ip:port") suitable for logs and
// sticky-instance lookup. Stable for the lifetime of the instance.
func (si *ServerInstance) HostPort() string { return si.hostPort }

// Weight satisfies loadbalance.Weighted.
func (si *ServerInstance) Weight() int { return si.weight }

// ServerCluster is one logical nhp-server identity (one pubkey)
// addressable as a set of physical instances with a load-balance
// policy.
//
// Cluster lifecycle is managed entirely by updateServerPeers in
// config.go; nothing in the cluster is mutable after construction
// except the Picker's internal round-robin counter (which is itself
// concurrency-safe).
type ServerCluster struct {
	// PublicKeyBase64 is the common pubkey of every instance in the
	// cluster — the value the device's peer map keys on.
	PublicKeyBase64 string

	// Name is an optional human-readable label echoed in logs.
	Name string

	// Sticky controls whether a KnockTarget that already picked an
	// instance reuses that instance on retries (KNK → COK → RKN).
	// True (the default) avoids re-running the cookie handshake on a
	// different instance; false makes every send re-pick, which is
	// only safe because the cluster's nhp-servers share a stateless
	// cookie signing key.
	Sticky bool

	instances []*ServerInstance
	picker    *loadbalance.Picker[*ServerInstance]

	// representativePeer is the *UdpPeer registered with the core
	// device for this cluster. The device's peerMap is keyed by
	// pubkey, so only one peer per cluster ever lives there; the
	// other instances are addressed by SendAddr at send time.
	representativePeer *core.UdpPeer
}

// Instances returns the cluster's instance list; mutation is not
// permitted (the picker holds a snapshot).
func (sc *ServerCluster) Instances() []*ServerInstance { return sc.instances }

// RepresentativePeer returns the *UdpPeer registered with core.Device
// for this cluster — its public key drives noise handshake lookups.
// Callers that need only a pubkey/identity can use this; callers that
// need an address must Pick an instance.
func (sc *ServerCluster) RepresentativePeer() *core.UdpPeer { return sc.representativePeer }

// Pick selects an instance according to the cluster's load-balance
// scheme. Returns nil on an empty cluster (caller logs + errors).
func (sc *ServerCluster) Pick() *ServerInstance {
	if sc == nil || sc.picker == nil {
		return nil
	}
	inst, ok := sc.picker.Pick()
	if !ok {
		return nil
	}
	return inst
}

// FindInstanceByAddr looks up an instance whose HostPort matches addr.
// Used to validate that a sticky-pinned instance is still in the
// cluster after a config reload (instance lists may shrink). Returns
// nil if not found.
func (sc *ServerCluster) FindInstanceByAddr(addr string) *ServerInstance {
	if sc == nil {
		return nil
	}
	for _, inst := range sc.instances {
		if inst.hostPort == addr {
			return inst
		}
	}
	return nil
}

// buildCluster turns a parsed ClusterConfig into a runtime cluster.
// The returned cluster's representativePeer is NOT yet registered on a
// device — callers (updateServerPeers) are responsible for that, so
// they can also handle peer removal on reload.
func buildCluster(cfg *ClusterConfig) (*ServerCluster, error) {
	if cfg.PubKeyBase64 == "" {
		return nil, fmt.Errorf("cluster %q: missing publicKeyBase64", cfg.Name)
	}
	if len(cfg.Instances) == 0 {
		return nil, fmt.Errorf("cluster %q (%s): no instances configured",
			cfg.Name, cfg.PubKeyBase64)
	}
	if err := cfg.LoadBalance.Validate(); err != nil {
		return nil, fmt.Errorf("cluster %q (%s): %w",
			cfg.Name, cfg.PubKeyBase64, err)
	}

	sc := &ServerCluster{
		PublicKeyBase64: cfg.PubKeyBase64,
		Name:            cfg.Name,
		Sticky:          cfg.stickyOrDefault(),
		instances:       make([]*ServerInstance, 0, len(cfg.Instances)),
	}

	for i, ic := range cfg.Instances {
		host := ic.Host
		ip := ic.Ip
		if host == "" && ip == "" {
			return nil, fmt.Errorf("cluster %q instance #%d: must set either Host or Ip",
				cfg.Name, i)
		}
		if ic.Port <= 0 {
			return nil, fmt.Errorf("cluster %q instance #%d: invalid port %d",
				cfg.Name, i, ic.Port)
		}
		peer := &core.UdpPeer{
			PubKeyBase64: cfg.PubKeyBase64,
			Hostname:     host,
			Ip:           ip,
			Port:         ic.Port,
			Type:         core.NHP_SERVER,
			ExpireTime:   cfg.ExpireTime,
		}
		displayHost := host
		if displayHost == "" {
			displayHost = ip
		}
		sc.instances = append(sc.instances, &ServerInstance{
			peer:     peer,
			weight:   ic.Weight,
			hostPort: fmt.Sprintf("%s:%d", displayHost, ic.Port),
		})
	}

	// The representative peer is the one we register with the device's
	// peer map for noise handshake key material. Any instance works —
	// they all share the same pubkey — but we pin to instances[0] for
	// determinism (so logs / debugging show a stable choice).
	sc.representativePeer = sc.instances[0].peer
	sc.picker = loadbalance.NewPicker(cfg.LoadBalance, sc.instances)
	return sc, nil
}
