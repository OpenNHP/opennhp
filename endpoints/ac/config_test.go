package ac

import (
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/common/clusterconfig"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// TestNormalizeAndExpand_LegacySingleEndpoint covers the pre-cluster
// schema: a single [[Servers]] entry with Ip+Port produces exactly one
// UdpPeer with those fields preserved. Existing deployments must
// continue to load — that's the whole point of keeping the legacy
// auto-upgrade path inside clusterconfig.Normalize.
func TestNormalizeAndExpand_LegacySingleEndpoint(t *testing.T) {
	entries := []*clusterconfig.ClusterConfig{{
		PubKeyBase64: "ABC=",
		Ip:           "192.168.0.1",
		Port:         62206,
		ExpireTime:   1924991999,
	}}
	peers, err := normalizeAndExpand(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(peers) != 1 {
		t.Fatalf("want 1 peer, got %d", len(peers))
	}
	p := peers[0]
	if p.PubKeyBase64 != "ABC=" || p.Ip != "192.168.0.1" || p.Port != 62206 || p.ExpireTime != 1924991999 {
		t.Fatalf("legacy entry not preserved: %+v", p)
	}
}

// TestNormalizeAndExpand_LegacyHostname: Hostname-only entry (DNS path)
// preserves the Hostname field so core.ResolveHost() can resolve later.
func TestNormalizeAndExpand_LegacyHostname(t *testing.T) {
	entries := []*clusterconfig.ClusterConfig{{
		PubKeyBase64: "ABC=",
		Hostname:     "server.example.com",
		Port:         62206,
	}}
	peers, err := normalizeAndExpand(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(peers) != 1 {
		t.Fatalf("want 1 peer, got %d", len(peers))
	}
	if peers[0].Hostname != "server.example.com" {
		t.Fatalf("hostname dropped: %+v", peers[0])
	}
}

// TestNormalizeAndExpand_InstancesFanOut: the structured [[Instances]]
// form expands to N UdpPeer rows sharing a pubkey. This is the property
// that lets an AC register (AOL) with N nhp-server instances behind one
// logical identity.
func TestNormalizeAndExpand_InstancesFanOut(t *testing.T) {
	entries := []*clusterconfig.ClusterConfig{{
		PubKeyBase64: "ABC=",
		ExpireTime:   1924991999,
		Instances: []clusterconfig.InstanceConfig{
			{Ip: "10.0.0.1", Port: 62206},
			{Ip: "10.0.0.2", Port: 62206},
			{Ip: "10.0.0.3", Port: 62206},
		},
	}}
	peers, err := normalizeAndExpand(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(peers) != 3 {
		t.Fatalf("want 3 peers, got %d", len(peers))
	}
	wantIps := map[string]bool{"10.0.0.1": false, "10.0.0.2": false, "10.0.0.3": false}
	for _, p := range peers {
		if p.PubKeyBase64 != "ABC=" {
			t.Fatalf("pubkey not propagated to expanded peer: %+v", p)
		}
		if p.Port != 62206 {
			t.Fatalf("port not propagated: %+v", p)
		}
		if p.ExpireTime != 1924991999 {
			t.Fatalf("expireTime not propagated: %+v", p)
		}
		if _, ok := wantIps[p.Ip]; !ok {
			t.Fatalf("unexpected ip %q", p.Ip)
		}
		wantIps[p.Ip] = true
	}
	for ip, seen := range wantIps {
		if !seen {
			t.Fatalf("missing peer for %s", ip)
		}
	}
}

// TestNormalizeAndExpand_RejectsMixedForms: setting BOTH top-level
// Ip/Port AND [[Servers.Instances]] in the same entry is almost
// certainly an incomplete migration — fail load rather than guessing
// which form wins. This is what blocks a partial upgrade from silently
// dropping one source of truth.
func TestNormalizeAndExpand_RejectsMixedForms(t *testing.T) {
	entries := []*clusterconfig.ClusterConfig{{
		PubKeyBase64: "ABC=",
		Ip:           "1.1.1.1",
		Port:         9999,
		Instances:    []clusterconfig.InstanceConfig{{Ip: "10.0.0.1", Port: 62206}},
	}}
	peers, err := normalizeAndExpand(entries)
	if err == nil {
		t.Fatalf("expected error for mixed legacy + Instances form; got peers=%+v", peers)
	}
	if peers != nil {
		t.Fatalf("on error the peers slice must be nil so callers can fail-close; got %d peers", len(peers))
	}
}

// TestNormalizeAndExpand_RejectsEmpty: a [[Servers]] entry with no
// instances and no legacy fields is structurally useless — fail load
// rather than booting an AC that silently can't reach any server.
func TestNormalizeAndExpand_RejectsEmpty(t *testing.T) {
	entries := []*clusterconfig.ClusterConfig{{PubKeyBase64: "ABC="}}
	peers, err := normalizeAndExpand(entries)
	if err == nil {
		t.Fatalf("expected error for empty entry; got peers=%+v", peers)
	}
	if peers != nil {
		t.Fatalf("on error the peers slice must be nil; got %d peers", len(peers))
	}
}

// TestNormalizeAndExpand_RejectsBadPort: instance with zero port surfaces
// at load. AOL fan-out with port=0 would send to :0 silently.
func TestNormalizeAndExpand_RejectsBadPort(t *testing.T) {
	entries := []*clusterconfig.ClusterConfig{{
		PubKeyBase64: "ABC=",
		Instances:    []clusterconfig.InstanceConfig{{Ip: "10.0.0.1", Port: 0}},
	}}
	_, err := normalizeAndExpand(entries)
	if err == nil {
		t.Fatal("expected error for invalid Port=0")
	}
}

// TestNormalizeAndExpand_RejectsDuplicatePubKey: two clusters with the
// same pubkey race for the same slot in device.peerMap. Catch it at
// load — the runtime symptom would be "one cluster silently
// disappears" which is much harder to diagnose.
func TestNormalizeAndExpand_RejectsDuplicatePubKey(t *testing.T) {
	entries := []*clusterconfig.ClusterConfig{
		{
			PubKeyBase64: "samekey",
			Instances:    []clusterconfig.InstanceConfig{{Ip: "10.0.0.1", Port: 62206}},
		},
		{
			PubKeyBase64: "samekey",
			Instances:    []clusterconfig.InstanceConfig{{Ip: "10.0.0.2", Port: 62206}},
		},
	}
	_, err := normalizeAndExpand(entries)
	if err == nil || !strings.Contains(err.Error(), "samekey") {
		t.Fatalf("normalize must reject duplicate PubKeyBase64, got: %v", err)
	}
}

// TestLegacyTomlStillParses guards that the pre-cluster on-disk format
// still unmarshals + normalizes cleanly. Existing single-server AC
// deployments must keep loading without operator edits.
func TestLegacyTomlStillParses(t *testing.T) {
	legacy := `
[[Servers]]
Hostname = ""
Ip = "192.168.80.35"
Port = 62206
PubKeyBase64 = "WqJxe+Z4+wLen3VRgZx6YnbjvJFmptz99zkONCt/7gc="
ExpireTime = 1924991999
`
	var peers Peers
	if err := toml.Unmarshal([]byte(legacy), &peers); err != nil {
		t.Fatalf("legacy toml failed to unmarshal: %v", err)
	}
	if len(peers.Servers) != 1 {
		t.Fatalf("want 1 entry, got %d", len(peers.Servers))
	}
	expanded, err := normalizeAndExpand(peers.Servers)
	if err != nil {
		t.Fatalf("legacy toml normalize: %v", err)
	}
	if len(expanded) != 1 {
		t.Fatalf("want 1 expanded peer, got %d", len(expanded))
	}
	if expanded[0].Ip != "192.168.80.35" || expanded[0].Port != 62206 {
		t.Fatalf("legacy endpoint not parsed: %+v", expanded[0])
	}
}

// TestClusterTomlParses guards the new structured form end-to-end.
func TestClusterTomlParses(t *testing.T) {
	cfg := `
[[Servers]]
PubKeyBase64 = "ABC="
ExpireTime = 1924991999

  [[Servers.Instances]]
  Ip = "10.0.0.1"
  Port = 62206

  [[Servers.Instances]]
  Ip = "10.0.0.2"
  Port = 62206
`
	var peers Peers
	if err := toml.Unmarshal([]byte(cfg), &peers); err != nil {
		t.Fatalf("new toml failed to unmarshal: %v", err)
	}
	expanded, err := normalizeAndExpand(peers.Servers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expanded) != 2 {
		t.Fatalf("want 2 expanded peers, got %d", len(expanded))
	}
}

// TestEndpointKey_DistinctSamePubKey: the AC-internal map key for a server
// peer must distinguish two endpoints that share a pubkey but differ in
// address. This is what allows updateServerPeers to keep both rows and what
// makes the discovery fan-out launch one routine per (pubkey, addr).
func TestEndpointKey_DistinctSamePubKey(t *testing.T) {
	a := &core.UdpPeer{PubKeyBase64: "ABC=", Ip: "10.0.0.1", Port: 62206}
	b := &core.UdpPeer{PubKeyBase64: "ABC=", Ip: "10.0.0.2", Port: 62206}
	if endpointKey(a) == endpointKey(b) {
		t.Fatalf("same-pubkey different-addr peers must produce distinct keys; got %q and %q",
			endpointKey(a), endpointKey(b))
	}
}

// TestEndpointKey_StableForSamePeer: identical (pubkey, addr) must produce
// the identical key across calls. updateServerPeers reuses this for diff'ing
// the new config against the live map.
func TestEndpointKey_StableForSamePeer(t *testing.T) {
	a := &core.UdpPeer{PubKeyBase64: "ABC=", Ip: "10.0.0.1", Port: 62206}
	b := &core.UdpPeer{PubKeyBase64: "ABC=", Ip: "10.0.0.1", Port: 62206}
	if endpointKey(a) != endpointKey(b) {
		t.Fatalf("equal peers produced different keys: %q vs %q", endpointKey(a), endpointKey(b))
	}
}

// TestEndpointKey_DistinctSamePubKeyDifferentHostnames: two legacy
// hostname-only entries that share a pubkey but resolve via different
// DNS names must produce distinct keys, otherwise the second overwrites
// the first in serverPeerMap and AOL fan-out skips one of the
// instances. Regression: the original implementation keyed on
// "pubkey|ip:port" where Ip is empty for legacy hostname entries, so
// "pk|:port" collided.
func TestEndpointKey_DistinctSamePubKeyDifferentHostnames(t *testing.T) {
	a := &core.UdpPeer{PubKeyBase64: "ABC=", Hostname: "a.example.com", Port: 62206}
	b := &core.UdpPeer{PubKeyBase64: "ABC=", Hostname: "b.example.com", Port: 62206}
	if endpointKey(a) == endpointKey(b) {
		t.Fatalf("same-pubkey different-hostname peers must produce distinct keys; got %q",
			endpointKey(a))
	}
}
