package agent

import (
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common/loadbalance"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// buildTwoInstanceCluster is a small helper mirroring the setup used by
// the sticky/round-robin tests: a normalized, built two-instance cluster.
func buildTwoInstanceCluster(t *testing.T) *ServerCluster {
	t.Helper()
	cfg := &ClusterConfig{
		Name:         "c1",
		PubKeyBase64: "k1",
		LoadBalance:  loadbalance.SchemeRoundRobin,
		Instances: []InstanceConfig{
			{Ip: "10.0.0.1", Port: 62206},
			{Ip: "10.0.0.2", Port: 62206},
		},
	}
	if err := normalizeClusters([]*ClusterConfig{cfg}, (&recordingDeprecate{}).Warn); err != nil {
		t.Fatalf("normalize: %v", err)
	}
	sc, err := buildCluster(cfg)
	if err != nil {
		t.Fatalf("buildCluster: %v", err)
	}
	return sc
}

// TestServerCluster_PeerForCipherScheme documents the current
// (scheme-agnostic) contract: the cluster's single configured peer is
// returned for every cipher scheme, and it is the same object as
// RepresentativePeer(). If per-scheme peers are ever introduced this test
// is the guard that forces the contract to be revisited deliberately.
func TestServerCluster_PeerForCipherScheme(t *testing.T) {
	sc := buildTwoInstanceCluster(t)
	rep := sc.RepresentativePeer()
	if rep == nil {
		t.Fatal("RepresentativePeer must be non-nil after buildCluster")
	}
	for _, scheme := range []int{0, 1, 2, -1} {
		if got := sc.PeerForCipherScheme(scheme); got != rep {
			t.Fatalf("PeerForCipherScheme(%d) = %p, want representativePeer %p", scheme, got, rep)
		}
	}
}

// TestKnockTarget_GetServerPeerForScheme_WithCluster: a target bound to a
// cluster delegates to the cluster's PeerForCipherScheme.
func TestKnockTarget_GetServerPeerForScheme_WithCluster(t *testing.T) {
	sc := buildTwoInstanceCluster(t)
	target := &KnockTarget{ServerCluster: sc}
	if got := target.GetServerPeerForScheme(0); got != sc.RepresentativePeer() {
		t.Fatalf("GetServerPeerForScheme with cluster = %p, want %p", got, sc.RepresentativePeer())
	}
}

// TestKnockTarget_GetServerPeerForScheme_NoCluster: a cluster-less target
// (legacy SDK shape) falls back to its ServerPeer instead of panicking.
func TestKnockTarget_GetServerPeerForScheme_NoCluster(t *testing.T) {
	peer := &core.UdpPeer{}
	target := &KnockTarget{ServerPeer: peer}
	if got := target.GetServerPeerForScheme(1); got != peer {
		t.Fatalf("GetServerPeerForScheme without cluster = %p, want ServerPeer %p", got, peer)
	}

	// And with neither cluster nor peer, it returns nil rather than crashing.
	empty := &KnockTarget{}
	if got := empty.GetServerPeerForScheme(0); got != nil {
		t.Fatalf("GetServerPeerForScheme on empty target = %p, want nil", got)
	}
}
