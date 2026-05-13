package ac

import (
	"testing"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// TestExpandServerPeers_LegacySingleEndpoint covers the pre-2A schema: a
// single [[Servers]] entry with Ip+Port produces exactly one UdpPeer with
// those fields preserved. Existing deployments must continue to load.
func TestExpandServerPeers_LegacySingleEndpoint(t *testing.T) {
	entries := []ServerPeerEntry{{
		PubKeyBase64: "ABC=",
		Hostname:     "",
		Ip:           "192.168.0.1",
		Port:         62206,
		ExpireTime:   1924991999,
	}}
	peers := expandServerPeers(entries)
	if len(peers) != 1 {
		t.Fatalf("want 1 peer, got %d", len(peers))
	}
	p := peers[0]
	if p.PubKeyBase64 != "ABC=" || p.Ip != "192.168.0.1" || p.Port != 62206 || p.ExpireTime != 1924991999 {
		t.Fatalf("legacy entry not preserved: %+v", p)
	}
}

// TestExpandServerPeers_LegacyHostname: Hostname-only entry (DNS path)
// preserves the Hostname field so core.ResolveHost() can resolve later.
func TestExpandServerPeers_LegacyHostname(t *testing.T) {
	entries := []ServerPeerEntry{{
		PubKeyBase64: "ABC=",
		Hostname:     "server.example.com",
		Port:         62206,
	}}
	peers := expandServerPeers(entries)
	if len(peers) != 1 {
		t.Fatalf("want 1 peer, got %d", len(peers))
	}
	if peers[0].Hostname != "server.example.com" {
		t.Fatalf("hostname dropped: %+v", peers[0])
	}
}

// TestExpandServerPeers_EndpointsFanOut: the new Endpoints field expands
// to N UdpPeer rows sharing a pubkey. This is the property that lets an AC
// register with N nhp-server instances behind one logical identity.
func TestExpandServerPeers_EndpointsFanOut(t *testing.T) {
	entries := []ServerPeerEntry{{
		PubKeyBase64: "ABC=",
		Endpoints:    []string{"10.0.0.1:62206", "10.0.0.2:62206", "10.0.0.3:62206"},
		ExpireTime:   1924991999,
	}}
	peers := expandServerPeers(entries)
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

// TestExpandServerPeers_EndpointsOverrideLegacy: when both Endpoints and
// the legacy Ip/Port are set, Endpoints wins. This must hold so a partial
// migration cannot silently mix two sources of truth.
func TestExpandServerPeers_EndpointsOverrideLegacy(t *testing.T) {
	entries := []ServerPeerEntry{{
		PubKeyBase64: "ABC=",
		Ip:           "1.1.1.1",
		Port:         9999,
		Endpoints:    []string{"10.0.0.1:62206"},
	}}
	peers := expandServerPeers(entries)
	if len(peers) != 1 {
		t.Fatalf("want 1 peer, got %d", len(peers))
	}
	if peers[0].Ip != "10.0.0.1" || peers[0].Port != 62206 {
		t.Fatalf("legacy fields leaked through: %+v", peers[0])
	}
}

// TestExpandServerPeers_InvalidEndpointSkipped: malformed endpoint entries
// are skipped (not fatal) so one typo doesn't block the rest from loading.
func TestExpandServerPeers_InvalidEndpointSkipped(t *testing.T) {
	entries := []ServerPeerEntry{{
		PubKeyBase64: "ABC=",
		Endpoints:    []string{"not-a-valid-endpoint", "10.0.0.2:62206"},
	}}
	peers := expandServerPeers(entries)
	if len(peers) != 1 {
		t.Fatalf("want 1 peer (the valid one), got %d", len(peers))
	}
	if peers[0].Ip != "10.0.0.2" {
		t.Fatalf("kept the wrong endpoint: %+v", peers[0])
	}
}

// TestLegacyTomlStillParses guards that a pre-2A server.toml format still
// unmarshals through the new Peers wrapper without operator changes.
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
	e := peers.Servers[0]
	if e.PubKeyBase64 != "WqJxe+Z4+wLen3VRgZx6YnbjvJFmptz99zkONCt/7gc=" {
		t.Fatalf("pubkey not parsed: %+v", e)
	}
	if e.Ip != "192.168.80.35" || e.Port != 62206 {
		t.Fatalf("legacy endpoint not parsed: %+v", e)
	}
	if len(e.Endpoints) != 0 {
		t.Fatalf("legacy toml should produce no Endpoints, got %v", e.Endpoints)
	}
}

// TestNewTomlEndpointsParses guards the new format end-to-end.
func TestNewTomlEndpointsParses(t *testing.T) {
	cfg := `
[[Servers]]
PubKeyBase64 = "ABC="
Endpoints = ["10.0.0.1:62206", "10.0.0.2:62206"]
ExpireTime = 1924991999
`
	var peers Peers
	if err := toml.Unmarshal([]byte(cfg), &peers); err != nil {
		t.Fatalf("new toml failed to unmarshal: %v", err)
	}
	expanded := expandServerPeers(peers.Servers)
	if len(expanded) != 2 {
		t.Fatalf("want 2 expanded peers, got %d", len(expanded))
	}
}

// TestEndpointKey_DistinctSamePubKey: the AC-internal map key for a server
// peer must distinguish two endpoints that share a pubkey but differ in
// address. This is what allows updateServerPeers to keep both rows and what
// makes the discovery launcher fan AOL out to both.
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
