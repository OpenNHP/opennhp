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

func fakeFingerprint(seed byte) string {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = seed + byte(i)
	}
	return utils.PubKeyFingerprint(raw)
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

// TestConfig_RejectsMultipleInstancesPerCluster locks in the phase-1 boundary:
// multi-instance configs MUST be rejected at load time until nhp/core can
// support multiple peers per public key (phase 2).
func TestConfig_RejectsMultipleInstancesPerCluster(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		Clusters: []Cluster{{
			PublicKeyBase64: fakeKey(0x20),
			Instances: []ClusterInstance{
				{Host: "10.0.0.1", Port: 62206},
				{Host: "10.0.0.2", Port: 62206},
			},
		}},
	}
	err := cfg.normalize()
	if err == nil {
		t.Fatalf("expected error for multi-instance cluster, got nil")
	}
	if !strings.Contains(err.Error(), "multi-instance per cluster is not yet supported") {
		t.Errorf("error should mention phase-1 limit, got: %v", err)
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

// TestConfig_DefaultClusterIDMustMatch: a configured defaultClusterId that
// doesn't correspond to any cluster's fingerprint is a typo, and silently
// ignoring it would surface as confusing 4xx errors later.
func TestConfig_DefaultClusterIDMustMatch(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		DefaultClusterID: "not-a-real-fp",
		Clusters: []Cluster{{
			PublicKeyBase64: fakeKey(0x20),
			Instances:       []ClusterInstance{{Host: "1.1.1.1", Port: 62206}},
		}},
	}
	if err := cfg.normalize(); err == nil {
		t.Fatalf("expected error for unknown defaultClusterId")
	}
}

// TestConfig_DefaultClusterIDValid: when defaultClusterId matches an existing
// fingerprint, normalize accepts it.
func TestConfig_DefaultClusterIDValid(t *testing.T) {
	fp := fakeFingerprint(0x20)
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		DefaultClusterID: fp,
		Clusters: []Cluster{{
			PublicKeyBase64: fakeKey(0x20),
			Instances:       []ClusterInstance{{Host: "1.1.1.1", Port: 62206}},
		}},
	}
	if err := cfg.normalize(); err != nil {
		t.Fatalf("unexpected error: %v", err)
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
