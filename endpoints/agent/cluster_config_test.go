package agent

import (
	"strings"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common/loadbalance"
)

// recordingDeprecate captures the warning callbacks normalizeClusters
// would otherwise route to log.Warning. Tests assert both that the
// warning fires when expected AND that it omits the deprecation noise
// when operators use the new schema — silent acceptance of legacy
// configs would mask a half-finished migration.
type recordingDeprecate struct {
	calls []string
}

func (r *recordingDeprecate) Warn(format string, args ...any) {
	r.calls = append(r.calls, format)
}

// TestNormalize_LegacyFlatFormUpgrades: a single-server config in the
// pre-cluster schema (Hostname/Ip/Port at the top level, no Instances
// block) must auto-promote to a single-instance cluster with one
// deprecation warning. This is the migration path most existing
// agents will take on first upgrade — breaking it would force every
// operator to edit server.toml before the new binary boots.
func TestNormalize_LegacyFlatFormUpgrades(t *testing.T) {
	rec := &recordingDeprecate{}
	cfgs := []*ClusterConfig{{
		Name:         "c1",
		PubKeyBase64: "k1",
		Hostname:     "server1.example.com",
		Ip:           "10.0.0.1",
		Port:         62206,
	}}
	if err := normalizeClusters(cfgs, rec.Warn); err != nil {
		t.Fatalf("normalize legacy form: %v", err)
	}
	c := cfgs[0]
	if len(c.Instances) != 1 {
		t.Fatalf("legacy upgrade must produce exactly one instance, got %d", len(c.Instances))
	}
	inst := c.Instances[0]
	if inst.Host != "server1.example.com" || inst.Ip != "10.0.0.1" || inst.Port != 62206 {
		t.Fatalf("legacy fields not copied onto Instance: %+v", inst)
	}
	if c.Hostname != "" || c.Ip != "" || c.Port != 0 {
		t.Fatalf("legacy top-level fields must be zeroed post-upgrade to avoid double-counting: %+v", c)
	}
	if len(rec.calls) != 1 {
		t.Fatalf("expected exactly one deprecation warning, got %d: %v", len(rec.calls), rec.calls)
	}
	if !strings.Contains(rec.calls[0], "legacy") {
		t.Fatalf("deprecation warning must mention the legacy form, got: %q", rec.calls[0])
	}
}

// TestNormalize_ClusterFormSilent: a config using the new
// Instances-based form must validate without firing the deprecation
// warning. Otherwise operators who've already migrated will see
// confusing log noise.
func TestNormalize_ClusterFormSilent(t *testing.T) {
	rec := &recordingDeprecate{}
	cfgs := []*ClusterConfig{{
		Name:         "c1",
		PubKeyBase64: "k1",
		Instances: []InstanceConfig{
			{Ip: "10.0.0.1", Port: 62206, Weight: 1},
			{Ip: "10.0.0.2", Port: 62206, Weight: 2},
		},
	}}
	if err := normalizeClusters(cfgs, rec.Warn); err != nil {
		t.Fatalf("normalize cluster form: %v", err)
	}
	if len(rec.calls) != 0 {
		t.Fatalf("cluster form must not fire deprecation warnings, got %v", rec.calls)
	}
}

// TestNormalize_RejectsMixedForms: setting BOTH top-level Ip and
// [[Instances]] in the same entry is almost certainly an incomplete
// migration — pick one. Rejecting at load time forces the operator to
// resolve the ambiguity rather than silently dropping one form.
func TestNormalize_RejectsMixedForms(t *testing.T) {
	cfgs := []*ClusterConfig{{
		Name:         "c1",
		PubKeyBase64: "k1",
		Ip:           "10.0.0.1",
		Port:         62206,
		Instances:    []InstanceConfig{{Ip: "10.0.0.2", Port: 62206}},
	}}
	err := normalizeClusters(cfgs, (&recordingDeprecate{}).Warn)
	if err == nil {
		t.Fatal("normalize must reject mixed legacy + cluster form")
	}
	if !strings.Contains(err.Error(), "both") {
		t.Fatalf("error must mention 'both' forms, got: %v", err)
	}
}

// TestNormalize_RejectsEmpty: a [[Servers]] entry with neither top-
// level address fields nor an Instances block is structurally
// useless — fail load rather than booting an agent that silently
// can't reach any server.
func TestNormalize_RejectsEmpty(t *testing.T) {
	cfgs := []*ClusterConfig{{Name: "c1", PubKeyBase64: "k1"}}
	err := normalizeClusters(cfgs, (&recordingDeprecate{}).Warn)
	if err == nil {
		t.Fatal("normalize must reject entry with no instances and no legacy fields")
	}
}

// TestNormalize_DuplicatePubKeyRejected: two [[Servers]] entries
// sharing the same PubKeyBase64 race for the same slot in
// device.peerMap. Catch it at load time — the runtime symptom would
// be "one of the two clusters silently disappears" which is much
// harder to diagnose.
func TestNormalize_DuplicatePubKeyRejected(t *testing.T) {
	cfgs := []*ClusterConfig{
		{
			Name:         "c1",
			PubKeyBase64: "samekey",
			Instances:    []InstanceConfig{{Ip: "10.0.0.1", Port: 62206}},
		},
		{
			Name:         "c2",
			PubKeyBase64: "samekey",
			Instances:    []InstanceConfig{{Ip: "10.0.0.2", Port: 62206}},
		},
	}
	err := normalizeClusters(cfgs, (&recordingDeprecate{}).Warn)
	if err == nil {
		t.Fatal("normalize must reject duplicate PubKeyBase64")
	}
	if !strings.Contains(err.Error(), "samekey") {
		t.Fatalf("error must mention the duplicated key, got: %v", err)
	}
}

// TestNormalize_DefaultsAndZeroWeight: an unset LoadBalance scheme
// must normalise to the documented default, and zero-weight instances
// must be promoted to weight 1 so they still receive traffic from
// weighted-random.
func TestNormalize_DefaultsAndZeroWeight(t *testing.T) {
	cfgs := []*ClusterConfig{{
		Name:         "c1",
		PubKeyBase64: "k1",
		Instances: []InstanceConfig{
			{Ip: "10.0.0.1", Port: 62206}, // no Weight set
		},
	}}
	if err := normalizeClusters(cfgs, (&recordingDeprecate{}).Warn); err != nil {
		t.Fatalf("normalize: %v", err)
	}
	c := cfgs[0]
	if c.LoadBalance != loadbalance.DefaultScheme {
		t.Fatalf("LoadBalance default: want %q, got %q", loadbalance.DefaultScheme, c.LoadBalance)
	}
	if c.Instances[0].Weight != 1 {
		t.Fatalf("zero weight must be promoted to 1, got %d", c.Instances[0].Weight)
	}
}

// TestNormalize_RejectsBadScheme: a typo like "weighted_random" in a
// fresh config must surface at load time, not as silent fallback to
// the default scheme. Same rationale as relay's validation:
// degraded-but-running is a worse failure mode than refusing to boot.
func TestNormalize_RejectsBadScheme(t *testing.T) {
	cfgs := []*ClusterConfig{{
		Name:         "c1",
		PubKeyBase64: "k1",
		LoadBalance:  "weighted_random", // underscore typo
		Instances:    []InstanceConfig{{Ip: "10.0.0.1", Port: 62206}},
	}}
	err := normalizeClusters(cfgs, (&recordingDeprecate{}).Warn)
	if err == nil {
		t.Fatal("normalize must reject unknown load-balance scheme")
	}
}

// TestStickyOrDefault: Sticky defaults to true when unset and round-
// trips its explicit value otherwise. This is the user-visible knob
// promise — accidentally flipping the default would change every
// agent's cookie-handshake behaviour on upgrade.
func TestStickyOrDefault(t *testing.T) {
	if !(&ClusterConfig{}).stickyOrDefault() {
		t.Fatal("unset Sticky must default to true")
	}
	tr, fa := true, false
	if !(&ClusterConfig{StickyInstance: &tr}).stickyOrDefault() {
		t.Fatal("explicit Sticky=true must round-trip")
	}
	if (&ClusterConfig{StickyInstance: &fa}).stickyOrDefault() {
		t.Fatal("explicit Sticky=false must round-trip")
	}
}

// TestBuildCluster_PickRespectsScheme: end-to-end check that
// buildCluster wires the Picker correctly and round-robin walks
// instances in order. Earlier tests cover the picker in isolation;
// this one catches mistakes in the cluster-construction glue (e.g.
// passing the wrong slice or scheme to NewPicker).
func TestBuildCluster_PickRespectsScheme(t *testing.T) {
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
	if !sc.Sticky {
		t.Fatal("buildCluster must default Sticky to true")
	}
	if sc.representativePeer != sc.instances[0].peer {
		t.Fatal("representativePeer must pin to instances[0] for determinism")
	}
	got := []string{}
	for i := 0; i < 4; i++ {
		got = append(got, sc.Pick().peer.Ip)
	}
	want := []string{"10.0.0.1", "10.0.0.2", "10.0.0.1", "10.0.0.2"}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("round-robin step %d: got %s want %s", i, got[i], want[i])
		}
	}
}

// TestKnockTarget_StickyHonored: when the cluster's Sticky flag is
// true (the default), PickInstance must capture the first pick and
// return the same instance on every subsequent call until reset. This
// is what keeps KNK and the follow-up RKN on the same nhp-server, so
// stateless cookies aren't strictly required for correctness — a
// regression here would silently re-enable the multi-instance cookie
// failure mode in non-stateless clusters.
func TestKnockTarget_StickyHonored(t *testing.T) {
	cfg := &ClusterConfig{
		Name:         "c1",
		PubKeyBase64: "k1",
		LoadBalance:  loadbalance.SchemeRoundRobin,
		Instances: []InstanceConfig{
			{Ip: "10.0.0.1", Port: 62206},
			{Ip: "10.0.0.2", Port: 62206},
			{Ip: "10.0.0.3", Port: 62206},
		},
	}
	if err := normalizeClusters([]*ClusterConfig{cfg}, (&recordingDeprecate{}).Warn); err != nil {
		t.Fatalf("normalize: %v", err)
	}
	sc, err := buildCluster(cfg)
	if err != nil {
		t.Fatalf("buildCluster: %v", err)
	}
	target := &KnockTarget{ServerCluster: sc}
	first := target.PickInstance()
	if first == nil {
		t.Fatal("first PickInstance returned nil")
	}
	for i := 0; i < 10; i++ {
		if got := target.PickInstance(); got != first {
			t.Fatalf("sticky PickInstance returned a different instance on call %d: got %s want %s",
				i, got.HostPort(), first.HostPort())
		}
	}
	// Reset releases the pin so the next pick re-runs the picker —
	// confirms the retry-on-failure path actually rotates.
	target.ResetInstancePin()
	second := target.PickInstance()
	if second == first {
		// Round-robin advances on every Pick, so a reset followed
		// by a fresh pick must land on a different instance.
		t.Fatalf("after ResetInstancePin, PickInstance returned the same instance %s — pin was not cleared",
			first.HostPort())
	}
}

// TestKnockTarget_NonStickyRotates: when Sticky=false every
// PickInstance call must re-run the picker (here round-robin),
// proving the sticky knob actually toggles behaviour. Catches
// "Sticky=false but pin still applied" bugs.
func TestKnockTarget_NonStickyRotates(t *testing.T) {
	fa := false
	cfg := &ClusterConfig{
		Name:           "c1",
		PubKeyBase64:   "k1",
		LoadBalance:    loadbalance.SchemeRoundRobin,
		StickyInstance: &fa,
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
	if sc.Sticky {
		t.Fatal("Sticky=false in config must propagate to ServerCluster.Sticky")
	}
	target := &KnockTarget{ServerCluster: sc}
	seen := map[string]bool{}
	for i := 0; i < 4; i++ {
		seen[target.PickInstance().HostPort()] = true
	}
	if len(seen) < 2 {
		t.Fatalf("non-sticky PickInstance pinned to one instance %v — sticky knob ignored", seen)
	}
}

// TestNormalize_NameRequired: Name is the operator-facing handle used
// from resource.toml (Cluster = "..."). A missing Name turns the
// resource lookup into "you forgot a string", which is a much worse
// failure mode at runtime than refusing to boot.
func TestNormalize_NameRequired(t *testing.T) {
	cfgs := []*ClusterConfig{{
		PubKeyBase64: "k1",
		Instances:    []InstanceConfig{{Ip: "10.0.0.1", Port: 62206}},
	}}
	err := normalizeClusters(cfgs, (&recordingDeprecate{}).Warn)
	if err == nil || !strings.Contains(err.Error(), "Name") {
		t.Fatalf("normalize must require Name, got: %v", err)
	}
}

// TestNormalize_NameCharsetRejected: names appear unquoted in
// resource.toml and in log lines; whitespace or quoting characters
// would force escaping at every callsite. Reject up front.
func TestNormalize_NameCharsetRejected(t *testing.T) {
	for _, bad := range []string{"has space", `with"quote`, "slash/path", "back\\slash"} {
		cfgs := []*ClusterConfig{{
			Name:         bad,
			PubKeyBase64: "k1",
			Instances:    []InstanceConfig{{Ip: "10.0.0.1", Port: 62206}},
		}}
		err := normalizeClusters(cfgs, (&recordingDeprecate{}).Warn)
		if err == nil {
			t.Fatalf("normalize must reject Name %q", bad)
		}
	}
}

// TestNormalize_DuplicateNameRejected: two clusters sharing a Name
// would cause silent routing to whichever entry won the map-insert
// race in updateServerPeers. Catch at load.
func TestNormalize_DuplicateNameRejected(t *testing.T) {
	cfgs := []*ClusterConfig{
		{
			Name:         "samename",
			PubKeyBase64: "k1",
			Instances:    []InstanceConfig{{Ip: "10.0.0.1", Port: 62206}},
		},
		{
			Name:         "samename",
			PubKeyBase64: "k2",
			Instances:    []InstanceConfig{{Ip: "10.0.0.2", Port: 62206}},
		},
	}
	err := normalizeClusters(cfgs, (&recordingDeprecate{}).Warn)
	if err == nil || !strings.Contains(err.Error(), "samename") {
		t.Fatalf("normalize must reject duplicate Name, got: %v", err)
	}
}
