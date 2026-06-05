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
// fields are auto-promoted into a single-server shape so existing demo
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
	if len(cfg.Servers) != 1 {
		t.Fatalf("legacy promotion should yield 1 server, got %d", len(cfg.Servers))
	}
	c := cfg.Servers[0]
	if c.PubKeyBase64 != fakeKey(0x20) {
		t.Errorf("server pubkey not promoted from legacy field")
	}
	if len(c.Instances) != 1 || c.Instances[0].Host != "10.0.0.1" || c.Instances[0].Port != 62206 {
		t.Errorf("server instance not promoted from legacy fields, got %+v", c.Instances)
	}
	if c.Instances[0].Weight != 1 {
		t.Errorf("instance weight should default to 1, got %d", c.Instances[0].Weight)
	}
	if c.LoadBalance != LBWeightedRandom {
		t.Errorf("loadBalance should default to weighted-random, got %q", c.LoadBalance)
	}
}

// TestConfig_AcceptsMultipleInstancesPerServer: phase 2 lifts the
// single-instance restriction. Two instances under one [[server]] must
// load cleanly, both addresses must be preserved, and weights must default
// to 1 the same way they would for a single-instance server.
//
// device.peerMap being keyed by pubkey is fine for this scheme: the
// per-instance UDP address lives on the serverInstance (used by the
// load-balance picker and outbound MsgData.RemoteAddr); validatePeer
// only consults LookupPeer for pubkey-existence and expiry, and uses
// ConnData.RemoteAddr (not Peer.Ip) for the per-connection address
// stickiness check.
func TestConfig_AcceptsMultipleInstancesPerServer(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		Servers: []Server{{
			PubKeyBase64: fakeKey(0x20),
			LoadBalance:  LBRoundRobin,
			Instances: []ServerInstance{
				{Host: "10.0.0.1", Port: 62206},
				{Host: "10.0.0.2", Port: 62206},
			},
		}},
	}
	if err := cfg.normalize(); err != nil {
		t.Fatalf("multi-instance server should load: %v", err)
	}
	got := cfg.Servers[0].Instances
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

// TestConfig_RejectsDuplicateServerPubkey catches operator mistakes where
// the same pubkey appears in two [[server]] blocks — the fingerprint would
// collide and routing would be undefined.
func TestConfig_RejectsDuplicateServerPubkey(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		Servers: []Server{
			{
				PubKeyBase64: fakeKey(0x20),
				Instances:    []ServerInstance{{Host: "1.1.1.1", Port: 62206}},
			},
			{
				PubKeyBase64: fakeKey(0x20), // same key
				Instances:    []ServerInstance{{Host: "2.2.2.2", Port: 62206}},
			},
		},
	}
	if err := cfg.normalize(); err == nil {
		t.Fatalf("expected error for duplicate server pubkey")
	}
}

// TestConfig_RejectsDuplicateInstanceAddressSamePubkey catches the
// unambiguous copy-paste mistake: the SAME identity listing the same
// host:port twice. The dedupe is scoped to (fingerprint, addr), so this is
// the case that still fails at load time.
func TestConfig_RejectsDuplicateInstanceAddressSamePubkey(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		Servers: []Server{
			{
				PubKeyBase64: fakeKey(0x20),
				Instances: []ServerInstance{
					{Host: "10.0.0.1", Port: 62206},
					{Host: "10.0.0.1", Port: 62206}, // same identity, same addr
				},
			},
		},
	}
	err := cfg.normalize()
	if err == nil {
		t.Fatalf("expected error for duplicate (host,port) under same pubkey")
	}
	if !strings.Contains(err.Error(), "already claimed") {
		t.Errorf("error should mention the conflict, got: %v", err)
	}
}

// TestConfig_AcceptsSameAddressDistinctPubkeys: two DISTINCT identities
// sharing one host:port (a SNI/header-routed front-end, or port-multiplexed
// identities) is a valid topology — resolveTarget routes by PeerPk — and
// must not be rejected at load time. Regression guard for the over-strict
// addr-only dedupe.
func TestConfig_AcceptsSameAddressDistinctPubkeys(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		Servers: []Server{
			{
				PubKeyBase64: fakeKey(0x20),
				Instances:    []ServerInstance{{Host: "10.0.0.1", Port: 62206}},
			},
			{
				PubKeyBase64: fakeKey(0x21), // distinct pubkey, same addr
				Instances:    []ServerInstance{{Host: "10.0.0.1", Port: 62206}},
			},
		},
	}
	if err := cfg.normalize(); err != nil {
		t.Fatalf("distinct pubkeys sharing one address should load: %v", err)
	}
}

// TestConfig_LegacyAndServersCoexist: when an operator's config has both
// the old single-server fields and a new [[server]] block, normalize
// must keep the [[server]] data, NOT promote the legacy fields. The
// promotion would otherwise silently overwrite the operator's explicit
// choice on a copy-paste upgrade.
func TestConfig_LegacyAndServersCoexist(t *testing.T) {
	cfg := &Config{
		PrivateKeyBase64: fakeKey(0x10),
		// Legacy values that should be ignored.
		NHPServerHost:            "legacy-host.example",
		NHPServerPort:            99999,
		NHPServerPublicKeyBase64: fakeKey(0x99),
		// Real config.
		Servers: []Server{{
			PubKeyBase64: fakeKey(0x20),
			Instances:    []ServerInstance{{Host: "10.0.0.1", Port: 62206}},
		}},
	}
	if err := cfg.normalize(); err != nil {
		t.Fatalf("coexistence should warn, not fail: %v", err)
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("expected exactly 1 server, got %d (legacy must not append)", len(cfg.Servers))
	}
	if cfg.Servers[0].Instances[0].Host != "10.0.0.1" {
		t.Errorf("explicit [[server]] block must win over legacy fields, got host=%q",
			cfg.Servers[0].Instances[0].Host)
	}
}

// TestConfig_RequiresAtLeastOneServer: a config with neither legacy fields
// nor any [[server]] block is unusable.
func TestConfig_RequiresAtLeastOneServer(t *testing.T) {
	cfg := &Config{PrivateKeyBase64: fakeKey(0x10)}
	if err := cfg.normalize(); err == nil {
		t.Fatalf("expected error when no server configured")
	}
}

// TestConfig_RequiresPrivateKey: the relay needs its own identity key.
func TestConfig_RequiresPrivateKey(t *testing.T) {
	cfg := &Config{
		Servers: []Server{{
			PubKeyBase64: fakeKey(0x20),
			Instances:    []ServerInstance{{Host: "1.1.1.1", Port: 62206}},
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
				Servers: []Server{{
					PubKeyBase64: fakeKey(0x20),
					LoadBalance:  LoadBalanceScheme(bad),
					Instances:    []ServerInstance{{Host: "1.1.1.1", Port: 62206}},
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
				Servers: []Server{{
					PubKeyBase64: fakeKey(0x20),
					LoadBalance:  ok,
					Instances:    []ServerInstance{{Host: "1.1.1.1", Port: 62206}},
				}},
			}
			if err := cfg.normalize(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Servers[0].LoadBalance != ok {
				t.Errorf("loadBalance mutated: got %q, want %q", cfg.Servers[0].LoadBalance, ok)
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
		Servers: []Server{{
			PubKeyBase64: fakeKey(0x20),
			Instances:    []ServerInstance{{Host: "1.1.1.1", Port: 62206}},
		}},
	}
	if err := cfg.normalize(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w := cfg.Servers[0].Instances[0].Weight; w != 1 {
		t.Errorf("default weight = %d, want 1", w)
	}
}
