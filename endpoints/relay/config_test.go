package relay

import (
	"strings"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/utils"
)

// fakeKey returns a deterministic 32-byte base64 string suitable for use as a
// publicKeyBase64 in tests. It is *not* a real Curve25519/SM2 key — the
// normalize() path only base64-decodes and fingerprints, so the bytes don't
// have to be valid curve points for these tests.
func fakeKey(seed byte) string {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = seed + byte(i)
	}
	return utils.Base64(raw)
}

// TestConfig_LegacyFieldsPromoted verifies that the deprecated single-server
// fields are auto-promoted into a single-cluster shape so existing demo
// configs keep working through phase 1.
func TestConfig_LegacyFieldsPromoted(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64:         fakeKey(0x10),
		NHPServerHost:            "10.0.0.1",
		NHPServerPort:            62206,
		NHPServerPublicKeyBase64: fakeKey(0x20),
	}
	if err := cfg.normalize(); err != nil {
		t.Fatalf("legacy fields should normalize without error: %v", err)
	}
	if len(cfg.Clusters) != 1 {
		t.Fatalf("legacy promotion should yield 1 cluster, got %d", len(cfg.Clusters))
	}
	c := cfg.Clusters[0]
	if c.PublicKeyBase64 != fakeKey(0x20) {
		t.Errorf("cluster pubkey not promoted from legacy field")
	}
	if len(c.Instances) != 1 || c.Instances[0].Host != "10.0.0.1" || c.Instances[0].Port != 62206 {
		t.Errorf("cluster instance not promoted from legacy fields, got %+v", c.Instances)
	}
	if c.Instances[0].Weight != 1 {
		t.Errorf("instance weight should default to 1, got %d", c.Instances[0].Weight)
	}
	if c.LoadBalance != LBWeightedRandom {
		t.Errorf("loadBalance should default to weighted-random, got %q", c.LoadBalance)
	}
}

// TestConfig_AcceptsMultipleInstancesPerCluster: phase 2 lifts the
// single-instance restriction. Two instances under one [[cluster]] must
// load cleanly, both addresses must be preserved, and weights must default
// to 1 the same way they would for a single-instance cluster.
func TestConfig_AcceptsMultipleInstancesPerCluster(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		Clusters: []Cluster{{
			PublicKeyBase64: fakeKey(0x20),
			LoadBalance:     LBRoundRobin,
			Instances: []ClusterInstance{
				{Host: "10.0.0.1", Port: 62206},
				{Host: "10.0.0.2", Port: 62206},
			},
		}},
	}
	if err := cfg.normalize(); err != nil {
		t.Fatalf("multi-instance cluster should load: %v", err)
	}
	got := cfg.Clusters[0].Instances
	if len(got) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(got))
	}
	if got[0].Host != "10.0.0.1" || got[1].Host != "10.0.0.2" {
		t.Errorf("instance hosts not preserved in order: %+v", got)
	}
	if got[0].Weight != 1 || got[1].Weight != 1 {
		t.Errorf("instance weights should default to 1, got %d/%d", got[0].Weight, got[1].Weight)
	}
}

// TestConfig_RejectsDuplicateClusterPubkey catches operator mistakes where
// the same pubkey appears in two [[cluster]] blocks — the fingerprint would
// collide and routing would be undefined.
func TestConfig_RejectsDuplicateClusterPubkey(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		Clusters: []Cluster{
			{
				PublicKeyBase64: fakeKey(0x20),
				Instances:       []ClusterInstance{{Host: "1.1.1.1", Port: 62206}},
			},
			{
				PublicKeyBase64: fakeKey(0x20), // same key
				Instances:       []ClusterInstance{{Host: "2.2.2.2", Port: 62206}},
			},
		},
	}
	if err := cfg.normalize(); err == nil {
		t.Fatalf("expected error for duplicate cluster pubkey")
	}
}

// TestConfig_RejectsDuplicateInstanceAddress catches the silent-failure mode
// where two clusters point at the same host:port. resolveTarget keys by
// PeerPk so this wouldn't misroute on its own, but it's almost always a
// copy-paste mistake the operator wants to hear about at load time rather
// than discover later.
func TestConfig_RejectsDuplicateInstanceAddress(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		Clusters: []Cluster{
			{
				PublicKeyBase64: fakeKey(0x20),
				Instances:       []ClusterInstance{{Host: "10.0.0.1", Port: 62206}},
			},
			{
				PublicKeyBase64: fakeKey(0x21), // distinct pubkey
				Instances:       []ClusterInstance{{Host: "10.0.0.1", Port: 62206}},
			},
		},
	}
	err := cfg.normalize()
	if err == nil {
		t.Fatalf("expected error for duplicate (host,port)")
	}
	if !strings.Contains(err.Error(), "already claimed") {
		t.Errorf("error should mention the conflict, got: %v", err)
	}
}

// TestConfig_LegacyAndClusterCoexist: when an operator's config has both
// the old single-server fields and a new [[cluster]] block, normalize
// must keep the [[cluster]] data, NOT promote the legacy fields. The
// promotion would otherwise silently overwrite the operator's explicit
// choice on a copy-paste upgrade.
func TestConfig_LegacyAndClusterCoexist(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		// Legacy values that should be ignored.
		NHPServerHost:            "legacy-host.example",
		NHPServerPort:            99999,
		NHPServerPublicKeyBase64: fakeKey(0x99),
		// Real config.
		Clusters: []Cluster{{
			PublicKeyBase64: fakeKey(0x20),
			Instances:       []ClusterInstance{{Host: "10.0.0.1", Port: 62206}},
		}},
	}
	if err := cfg.normalize(); err != nil {
		t.Fatalf("coexistence should warn, not fail: %v", err)
	}
	if len(cfg.Clusters) != 1 {
		t.Fatalf("expected exactly 1 cluster, got %d (legacy must not append)", len(cfg.Clusters))
	}
	if cfg.Clusters[0].Instances[0].Host != "10.0.0.1" {
		t.Errorf("explicit [[cluster]] block must win over legacy fields, got host=%q",
			cfg.Clusters[0].Instances[0].Host)
	}
}

// TestConfig_RequiresAtLeastOneCluster: a config with neither legacy fields
// nor any [[cluster]] block is unusable.
func TestConfig_RequiresAtLeastOneCluster(t *testing.T) {
	cfg := &Config{PrivateKeyBase64: fakeKey(0x10)}
	if err := cfg.normalize(); err == nil {
		t.Fatalf("expected error when no cluster configured")
	}
}

// TestConfig_RequiresPrivateKey: the relay needs its own identity key.
func TestConfig_RequiresPrivateKey(t *testing.T) {
	cfg := &Config{
		Clusters: []Cluster{{
			PublicKeyBase64: fakeKey(0x20),
			Instances:       []ClusterInstance{{Host: "1.1.1.1", Port: 62206}},
		}},
	}
	if err := cfg.normalize(); err == nil {
		t.Fatalf("expected error for missing privateKeyBase64")
	}
}

// TestConfig_LoadBalanceUnknownRejected: a typo like "weighted_random" or
// "roundrobin" is harmless in phase 1 (the value is unused) but would
// silently fall through to the default policy once phase 2 enables LB.
// Reject at load time so the operator notices now.
func TestConfig_LoadBalanceUnknownRejected(t *testing.T) {
	for _, bad := range []string{"weighted_random", "roundrobin", "random-with-jitter", "rr"} {
		t.Run(bad, func(t *testing.T) {
			cfg := &Config{
				PrivateKeyBase64: fakeKey(0x10),
				Clusters: []Cluster{{
					PublicKeyBase64: fakeKey(0x20),
					LoadBalance:     LoadBalanceScheme(bad),
					Instances:       []ClusterInstance{{Host: "1.1.1.1", Port: 62206}},
				}},
			}
			if err := cfg.normalize(); err == nil {
				t.Fatalf("expected error for loadBalance=%q", bad)
			} else if !strings.Contains(err.Error(), "loadBalance") {
				t.Errorf("error should name the field, got: %v", err)
			}
		})
	}
}

// TestConfig_LoadBalanceKnownAccepted: all three documented schemes load
// without error and are preserved through normalization.
func TestConfig_LoadBalanceKnownAccepted(t *testing.T) {
	for _, ok := range []LoadBalanceScheme{LBRandom, LBWeightedRandom, LBRoundRobin} {
		t.Run(string(ok), func(t *testing.T) {
			cfg := &Config{
				PrivateKeyBase64: fakeKey(0x10),
				Clusters: []Cluster{{
					PublicKeyBase64: fakeKey(0x20),
					LoadBalance:     ok,
					Instances:       []ClusterInstance{{Host: "1.1.1.1", Port: 62206}},
				}},
			}
			if err := cfg.normalize(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Clusters[0].LoadBalance != ok {
				t.Errorf("loadBalance mutated: got %q, want %q", cfg.Clusters[0].LoadBalance, ok)
			}
		})
	}
}

// TestConfig_InstanceWeightDefault: unspecified weight should default to 1
// rather than 0 (which would make weighted-random treat the instance as a
// black hole once phase 2 enables LB).
func TestConfig_InstanceWeightDefault(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		Clusters: []Cluster{{
			PublicKeyBase64: fakeKey(0x20),
			Instances:       []ClusterInstance{{Host: "1.1.1.1", Port: 62206}},
		}},
	}
	if err := cfg.normalize(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w := cfg.Clusters[0].Instances[0].Weight; w != 1 {
		t.Errorf("default weight = %d, want 1", w)
	}
}
