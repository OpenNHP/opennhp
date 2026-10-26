package agent

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/common/loadbalance"
	"github.com/OpenNHP/opennhp/nhp/core"
)

func TestCipherSchemeFor(t *testing.T) {
	tests := []struct {
		name       string
		cipherType int
		want       int
	}{
		{"curve25519 (0)", int(core.ECC_CURVE25519), common.CIPHER_SCHEME_CURVE},
		{"sm2 (1)", int(core.ECC_SM2), common.CIPHER_SCHEME_GMSM},
		{"umi (2) falls back to curve", int(core.ECC_UMI), common.CIPHER_SCHEME_CURVE},
		{"unknown falls back to curve", 99, common.CIPHER_SCHEME_CURVE},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cipherSchemeFor(tt.cipherType); got != tt.want {
				t.Fatalf("cipherSchemeFor(%d) = %d, want %d", tt.cipherType, got, tt.want)
			}
		})
	}
}

func TestRegistrationResultJSON(t *testing.T) {
	// Optional fields must be omitted when empty so a minimal result stays small
	// and, critically, so no private key or stray field leaks into the blob the
	// platform layer may log.
	minimal := RegistrationResult{ErrCode: "0"}.JSON()
	var decoded map[string]any
	if err := json.Unmarshal([]byte(minimal), &decoded); err != nil {
		t.Fatalf("minimal JSON did not parse: %v", err)
	}
	if decoded["errCode"] != "0" {
		t.Fatalf("errCode = %v, want 0", decoded["errCode"])
	}
	for _, absent := range []string{"errMsg", "publicKey", "aspId", "expiresAt"} {
		if _, ok := decoded[absent]; ok {
			t.Errorf("expected %q to be omitted from minimal result, got %v", absent, decoded[absent])
		}
	}

	// A fully populated result round-trips every field, including the expiresAt
	// pointer.
	exp := int64(1893456000)
	full := RegistrationResult{
		ErrCode:       "0",
		ErrMsg:        "",
		PublicKey:     "pub==",
		AuthServiceId: "asp-1",
		ExpiresAt:     &exp,
	}.JSON()
	var r RegistrationResult
	if err := json.Unmarshal([]byte(full), &r); err != nil {
		t.Fatalf("full JSON did not parse: %v", err)
	}
	if r.PublicKey != "pub==" || r.AuthServiceId != "asp-1" {
		t.Fatalf("round-trip lost fields: %+v", r)
	}
	if r.ExpiresAt == nil || *r.ExpiresAt != exp {
		t.Fatalf("expiresAt = %v, want %d", r.ExpiresAt, exp)
	}
}

// codedErr implements the ErrorCode() interface that registrationErrorCode
// probes for, without depending on common.Error internals.
type codedErr struct{ code string }

func (e codedErr) Error() string     { return "coded: " + e.code }
func (e codedErr) ErrorCode() string { return e.code }

func TestRegistrationErrorCode(t *testing.T) {
	if got := registrationErrorCode(common.ErrInvalidInput); got != common.ErrInvalidInput.ErrorCode() {
		t.Errorf("common.Error: got %q, want %q", got, common.ErrInvalidInput.ErrorCode())
	}
	if got := registrationErrorCode(codedErr{code: "51234"}); got != "51234" {
		t.Errorf("coded error: got %q, want 51234", got)
	}
	// An empty code from the coder must fall back to the generic "1", not "".
	if got := registrationErrorCode(codedErr{code: ""}); got != "1" {
		t.Errorf("empty-coded error: got %q, want 1", got)
	}
	// A plain error that does not implement the interface falls back to "1".
	if got := registrationErrorCode(errors.New("boom")); got != "1" {
		t.Errorf("plain error: got %q, want 1", got)
	}
}

func TestTakeGeneratedKey(t *testing.T) {
	// Isolate the package-global hold from any other test ordering.
	stashGeneratedKey("")

	stashGeneratedKey("secret-key")
	if got := TakeGeneratedKey(); got != "secret-key" {
		t.Fatalf("first take = %q, want secret-key", got)
	}
	// The secret must be surrendered exactly once: a second read is empty.
	if got := TakeGeneratedKey(); got != "" {
		t.Fatalf("second take = %q, want empty (key must be cleared on read)", got)
	}
	// Stashing empty clears a previously held key (the guard at the top of
	// RequestRegistrationOtp relies on this to never carry a stale secret).
	stashGeneratedKey("another")
	stashGeneratedKey("")
	if got := TakeGeneratedKey(); got != "" {
		t.Fatalf("after stash(\"\") take = %q, want empty", got)
	}
}

func TestInvalidRegistration(t *testing.T) {
	r := invalidRegistration("bad input")
	if r.ErrCode != common.ErrInvalidInput.ErrorCode() {
		t.Errorf("ErrCode = %q, want %q", r.ErrCode, common.ErrInvalidInput.ErrorCode())
	}
	if r.ErrMsg != "bad input" {
		t.Errorf("ErrMsg = %q, want \"bad input\"", r.ErrMsg)
	}
}

func TestAgentCipherScheme(t *testing.T) {
	if got := agentCipherScheme(nil); got != common.CIPHER_SCHEME_CURVE {
		t.Errorf("nil agent: got %d, want CURVE", got)
	}
	if got := agentCipherScheme(&UdpAgent{}); got != common.CIPHER_SCHEME_CURVE {
		t.Errorf("nil config: got %d, want CURVE", got)
	}
	a := &UdpAgent{config: &Config{DefaultCipherScheme: common.CIPHER_SCHEME_GMSM}}
	if got := agentCipherScheme(a); got != common.CIPHER_SCHEME_GMSM {
		t.Errorf("GMSM config: got %d, want GMSM", got)
	}
}

func TestConfigTomlExists(t *testing.T) {
	dir := t.TempDir()
	if configTomlExists(dir) {
		t.Fatalf("expected false for a dir with no etc/config.toml")
	}
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(etc, "config.toml"), []byte("x = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !configTomlExists(dir) {
		t.Fatalf("expected true once etc/config.toml exists")
	}
}

func TestApplyRegistrationUser(t *testing.T) {
	t.Run("merges email into user data", func(t *testing.T) {
		a := &UdpAgent{knockUser: &KnockUser{}}
		if err := applyRegistrationUser(a, "user@example.com", "dev-1", "org-1", `{"role":"admin"}`); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.deviceId != "dev-1" {
			t.Errorf("deviceId = %q, want dev-1", a.deviceId)
		}
		if a.knockUser.UserId != "user@example.com" {
			t.Errorf("UserId = %q, want user@example.com", a.knockUser.UserId)
		}
		if a.knockUser.OrganizationId != "org-1" {
			t.Errorf("OrganizationId = %q, want org-1", a.knockUser.OrganizationId)
		}
		if a.knockUser.UserData["email"] != "user@example.com" {
			t.Errorf("UserData[email] = %v, want user@example.com", a.knockUser.UserData["email"])
		}
		if a.knockUser.UserData["role"] != "admin" {
			t.Errorf("UserData[role] = %v, want admin", a.knockUser.UserData["role"])
		}
	})

	t.Run("empty user data is allowed", func(t *testing.T) {
		a := &UdpAgent{knockUser: &KnockUser{}}
		if err := applyRegistrationUser(a, "u@e.com", "dev", "org", ""); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.knockUser.UserData["email"] != "u@e.com" {
			t.Errorf("UserData[email] = %v, want u@e.com", a.knockUser.UserData["email"])
		}
	})

	t.Run("invalid user data JSON is rejected", func(t *testing.T) {
		a := &UdpAgent{knockUser: &KnockUser{}}
		err := applyRegistrationUser(a, "u@e.com", "dev", "org", "{not json")
		if err == nil {
			t.Fatal("expected an error for malformed userData")
		}
	})
}

// buildTestAgentWithCluster mirrors the fixture used by the cluster-resource
// tests: one known cluster "good-cluster" with representative pubkey "k1".
func buildTestAgentWithCluster(t *testing.T) *UdpAgent {
	t.Helper()
	cfg := &ClusterConfig{
		Name:         "good-cluster",
		PubKeyBase64: "k1",
		LoadBalance:  loadbalance.SchemeRoundRobin,
		Instances:    []InstanceConfig{{Ip: "10.0.0.1", Port: 62206}},
	}
	if err := normalizeClusters([]*ClusterConfig{cfg}, (&recordingDeprecate{}).Warn); err != nil {
		t.Fatalf("normalize: %v", err)
	}
	sc, err := buildCluster(cfg)
	if err != nil {
		t.Fatalf("buildCluster: %v", err)
	}
	return &UdpAgent{
		serverClusterMap:    map[string]*ServerCluster{"k1": sc},
		serverClusterByName: map[string]*ServerCluster{"good-cluster": sc},
	}
}

func TestRegistrationTarget(t *testing.T) {
	a := buildTestAgentWithCluster(t)

	t.Run("known cluster resolves to a target", func(t *testing.T) {
		target, err := registrationTarget(a, "asp", "res", "good-cluster")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if target == nil {
			t.Fatal("expected a non-nil target")
		}
		if target.ServerCluster == nil || target.ServerPeer == nil {
			t.Fatalf("target missing cluster/peer: %+v", target)
		}
		if target.AuthServiceId != "asp" || target.ResourceId != "res" || target.Cluster != "good-cluster" {
			t.Fatalf("target resource fields wrong: %+v", target.KnockResource)
		}
	})

	t.Run("unknown cluster is an error, nil target", func(t *testing.T) {
		target, err := registrationTarget(a, "asp", "res", "no-such-cluster")
		if err == nil {
			t.Fatal("expected an error for an unknown cluster")
		}
		if target != nil {
			t.Fatalf("expected nil target on error, got %+v", target)
		}
	})
}

func TestInstallRegistrationKey(t *testing.T) {
	t.Run("nil agent is rejected", func(t *testing.T) {
		if err := InstallRegistrationKey(nil, "", int(core.ECC_CURVE25519)); err == nil {
			t.Fatal("expected an error for a nil agent")
		}
	})

	t.Run("invalid base64 is rejected without reinit", func(t *testing.T) {
		a := &UdpAgent{config: &Config{}}
		err := InstallRegistrationKey(a, "@@ not base64 @@", int(core.ECC_CURVE25519))
		if err == nil {
			t.Fatal("expected an error for non-base64 private key")
		}
	})

	t.Run("invalid cipherType is rejected", func(t *testing.T) {
		a := &UdpAgent{config: &Config{}}
		if err := InstallRegistrationKey(a, "AAAA", int(core.ECC_UMI)); err == nil {
			t.Fatal("expected an error for an unsupported cipherType")
		}
		if err := InstallRegistrationKey(a, "AAAA", 99); err == nil {
			t.Fatal("expected an error for a garbage cipherType")
		}
	})

	t.Run("matching key and scheme is a no-op", func(t *testing.T) {
		// Build an agent whose device holds a known key; installing that exact
		// key+scheme must be a no-op (AgentHoldsRegistrationKey matches by public
		// key, so no device reinit is attempted).
		priv := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
		dev := core.NewDevice(core.NHP_AGENT, priv, nil)
		if dev == nil {
			t.Fatal("NewDevice returned nil")
		}
		dev.Start()
		t.Cleanup(dev.Stop)
		a := &UdpAgent{
			config: &Config{DefaultCipherScheme: common.CIPHER_SCHEME_CURVE},
			device: dev,
		}
		if err := InstallRegistrationKey(a, base64.StdEncoding.EncodeToString(priv), int(core.ECC_CURVE25519)); err != nil {
			t.Fatalf("expected no-op, got error: %v", err)
		}
	})
}

func TestPrepareRegistrationInputValidation(t *testing.T) {
	tests := []struct {
		name                         string
		email, aspId, resId, cluster string
	}{
		{"missing email", "", "asp", "res", "cl"},
		{"missing aspId", "u@e.com", "", "res", "cl"},
		{"missing resId", "u@e.com", "asp", "", "cl"},
		{"missing cluster", "u@e.com", "asp", "res", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, instance, target, owned := prepareRegistration(
				nil, "", 0, "priv", int(core.ECC_CURVE25519),
				tt.email, "dev", "org", "", tt.aspId, tt.resId, tt.cluster, false,
			)
			if result.ErrCode != common.ErrInvalidInput.ErrorCode() {
				t.Errorf("ErrCode = %q, want %q", result.ErrCode, common.ErrInvalidInput.ErrorCode())
			}
			if instance != nil || target != nil || owned {
				t.Errorf("expected nil instance/target and owned=false, got instance=%v target=%v owned=%v", instance, target, owned)
			}
		})
	}
}

func TestPrepareRegistrationRunningAgentGuards(t *testing.T) {
	// A running agent must never be silently re-keyed; each of these cases is a
	// security-relevant refusal telling the caller to close the agent first.
	runningWith := func(key string, scheme int) *UdpAgent {
		a := &UdpAgent{config: &Config{PrivateKeyBase64: key, DefaultCipherScheme: scheme}}
		a.running.Store(true)
		return a
	}

	t.Run("generate while running is refused", func(t *testing.T) {
		a := runningWith("stored", common.CIPHER_SCHEME_CURVE)
		result, instance, _, _ := prepareRegistration(
			a, "", 0, "", int(core.ECC_CURVE25519),
			"u@e.com", "dev", "org", "", "asp", "res", "cl", true,
		)
		if result.ErrCode != common.ErrInvalidInput.ErrorCode() || instance != nil {
			t.Fatalf("expected refusal, got code=%q msg=%q instance=%v", result.ErrCode, result.ErrMsg, instance)
		}
	})

	t.Run("empty key with no stored key is refused", func(t *testing.T) {
		a := runningWith("", common.CIPHER_SCHEME_CURVE)
		result, instance, _, _ := prepareRegistration(
			a, "", 0, "", int(core.ECC_CURVE25519),
			"u@e.com", "dev", "org", "", "asp", "res", "cl", false,
		)
		if result.ErrCode != common.ErrInvalidInput.ErrorCode() || instance != nil {
			t.Fatalf("expected refusal, got code=%q msg=%q instance=%v", result.ErrCode, result.ErrMsg, instance)
		}
	})

	t.Run("different key while running is refused", func(t *testing.T) {
		a := runningWith("stored", common.CIPHER_SCHEME_CURVE)
		result, instance, _, _ := prepareRegistration(
			a, "", 0, "a-different-key", int(core.ECC_CURVE25519),
			"u@e.com", "dev", "org", "", "asp", "res", "cl", false,
		)
		if result.ErrCode != common.ErrInvalidInput.ErrorCode() || instance != nil {
			t.Fatalf("expected refusal, got code=%q msg=%q instance=%v", result.ErrCode, result.ErrMsg, instance)
		}
	})

	t.Run("cipher scheme mismatch while running is refused", func(t *testing.T) {
		a := runningWith("stored", common.CIPHER_SCHEME_CURVE)
		// Same key, but request SM2 (GMSM) against a CURVE agent.
		result, instance, _, _ := prepareRegistration(
			a, "", 0, "stored", int(core.ECC_SM2),
			"u@e.com", "dev", "org", "", "asp", "res", "cl", false,
		)
		if result.ErrCode != common.ErrInvalidInput.ErrorCode() || instance != nil {
			t.Fatalf("expected refusal, got code=%q msg=%q instance=%v", result.ErrCode, result.ErrMsg, instance)
		}
	})

	t.Run("not running and no key and no generate is refused", func(t *testing.T) {
		result, instance, _, _ := prepareRegistration(
			nil, "", 0, "", int(core.ECC_CURVE25519),
			"u@e.com", "dev", "org", "", "asp", "res", "cl", false,
		)
		if result.ErrCode != common.ErrInvalidInput.ErrorCode() || instance != nil {
			t.Fatalf("expected refusal, got code=%q msg=%q instance=%v", result.ErrCode, result.ErrMsg, instance)
		}
	})
}

// newRunningReusableAgent builds a started-enough running agent that the
// register flow will reuse: it holds a real device (so
// PublicKeyBase64ByCipherScheme works) and a known knock identity, but its
// cluster maps are empty so registrationTarget fails after
// applyRegistrationUser has already mutated the identity — which is exactly the
// window the save/restore must cover, and it avoids any real network send.
// newRunningReusableAgent returns the agent and the base64 of the key its device
// actually holds, so callers can pass a privateKey that AgentHoldsRegistrationKey
// will match (reaching the reuse path) rather than one that gets rejected.
func newRunningReusableAgent(t *testing.T) (*UdpAgent, string) {
	t.Helper()
	priv := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	keyB64 := base64.StdEncoding.EncodeToString(priv)
	dev := core.NewDevice(core.NHP_AGENT, priv, nil)
	if dev == nil {
		t.Fatal("NewDevice returned nil")
	}
	dev.Start()
	t.Cleanup(dev.Stop)

	a := &UdpAgent{
		config:              &Config{PrivateKeyBase64: keyB64, DefaultCipherScheme: common.CIPHER_SCHEME_CURVE},
		device:              dev,
		deviceId:            "orig-device",
		knockUser:           &KnockUser{UserId: "orig@user", OrganizationId: "orig-org", UserData: map[string]any{"k": "v"}},
		serverClusterMap:    map[string]*ServerCluster{},
		serverClusterByName: map[string]*ServerCluster{},
	}
	a.running.Store(true)
	return a, keyB64
}

func assertOriginalIdentity(t *testing.T, a *UdpAgent) {
	t.Helper()
	a.knockUserMutex.RLock()
	defer a.knockUserMutex.RUnlock()
	if a.deviceId != "orig-device" {
		t.Errorf("deviceId = %q, want orig-device (registration leaked its identity into the live agent)", a.deviceId)
	}
	if a.knockUser.UserId != "orig@user" {
		t.Errorf("UserId = %q, want orig@user", a.knockUser.UserId)
	}
	if a.knockUser.OrganizationId != "orig-org" {
		t.Errorf("OrganizationId = %q, want orig-org", a.knockUser.OrganizationId)
	}
	if len(a.knockUser.UserData) != 1 || a.knockUser.UserData["k"] != "v" {
		t.Errorf("UserData = %v, want map[k:v] (registration data must not persist)", a.knockUser.UserData)
	}
}

func TestRequestRegistrationOtpRestoresRunningAgentIdentity(t *testing.T) {
	a, key := newRunningReusableAgent(t)
	res := RequestRegistrationOtp(
		a, "", 0, key, int(core.ECC_CURVE25519),
		"new@user", "new-device", "new-org", `{"x":1}`, "asp", "res", "missing-cluster",
	)
	// The unknown cluster makes this fail, which is fine — we only care that the
	// live agent's identity was put back after it was reused and mutated.
	if res.ErrCode == common.ErrSuccess.ErrorCode() {
		t.Fatalf("expected a failure result for an unknown cluster, got success: %+v", res)
	}
	if res.ErrMsg == "close the running agent before registering a different key" {
		t.Fatalf("agent was wrongly rejected instead of reused: %+v", res)
	}
	assertOriginalIdentity(t, a)
}

func TestRegisterDeviceRestoresRunningAgentIdentity(t *testing.T) {
	a, key := newRunningReusableAgent(t)
	res := RegisterDevice(
		a, "", 0, key, int(core.ECC_CURVE25519),
		"new@user", "new-device", "new-org", `{"x":1}`, "asp", "res", "missing-cluster", "123456",
	)
	if res.ErrCode == common.ErrSuccess.ErrorCode() {
		t.Fatalf("expected a failure result for an unknown cluster, got success: %+v", res)
	}
	if res.ErrMsg == "close the running agent before registering a different key" {
		t.Fatalf("agent was wrongly rejected instead of reused: %+v", res)
	}
	assertOriginalIdentity(t, a)
}

// TestRegistrationIdentityUpdatesAreRaceFree hammers a running agent with
// registration calls (each reuses it, overwrites its identity, then restores)
// while another goroutine reads the identity the way the knock path does, under
// knockUserMutex. Run under -race it proves SetDeviceId's write is synchronized
// with those reads; before SetDeviceId took the lock this failed the detector.
func TestRegistrationIdentityUpdatesAreRaceFree(t *testing.T) {
	a, key := newRunningReusableAgent(t)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				a.knockUserMutex.RLock()
				_ = a.deviceId
				_ = a.knockUser.UserId
				a.knockUserMutex.RUnlock()
			}
		}
	}()
	for i := 0; i < 50; i++ {
		RequestRegistrationOtp(
			a, "", 0, key, int(core.ECC_CURVE25519),
			"new@user", "new-device", "new-org", "", "asp", "res", "missing-cluster",
		)
	}
	close(stop)
	wg.Wait()
	assertOriginalIdentity(t, a)
}

func TestAgentHoldsRegistrationKey(t *testing.T) {
	priv := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	keyB64 := base64.StdEncoding.EncodeToString(priv)
	dev := core.NewDevice(core.NHP_AGENT, priv, nil)
	if dev == nil {
		t.Fatal("NewDevice returned nil")
	}
	dev.Start()
	t.Cleanup(dev.Stop)
	a := &UdpAgent{
		config: &Config{DefaultCipherScheme: common.CIPHER_SCHEME_CURVE},
		device: dev,
	}

	if !AgentHoldsRegistrationKey(a, keyB64, int(core.ECC_CURVE25519)) {
		t.Error("expected true for the agent's own key and scheme")
	}
	other := base64.StdEncoding.EncodeToString(core.NewECDH(core.ECC_CURVE25519).PrivateKey())
	if AgentHoldsRegistrationKey(a, other, int(core.ECC_CURVE25519)) {
		t.Error("expected false for a different key")
	}
	if AgentHoldsRegistrationKey(a, keyB64, int(core.ECC_SM2)) {
		t.Error("expected false for a mismatched cipher scheme")
	}
	for _, bad := range []string{"", "@@not base64@@"} {
		if AgentHoldsRegistrationKey(a, bad, int(core.ECC_CURVE25519)) {
			t.Errorf("expected false for invalid key %q", bad)
		}
	}
	if AgentHoldsRegistrationKey(nil, keyB64, int(core.ECC_CURVE25519)) {
		t.Error("expected false for a nil agent")
	}
	if AgentHoldsRegistrationKey(&UdpAgent{config: &Config{}}, keyB64, int(core.ECC_CURVE25519)) {
		t.Error("expected false for an agent with no device")
	}

	// Sealed-key case (#4): the config string is the sealed blob, not the
	// plaintext base64 — identity must still match by derived public key.
	sealed := &UdpAgent{
		config: &Config{PrivateKeyBase64: "AQIDBA==sealed-blob-not-the-plaintext", DefaultCipherScheme: common.CIPHER_SCHEME_CURVE},
		device: dev,
	}
	if !AgentHoldsRegistrationKey(sealed, keyB64, int(core.ECC_CURVE25519)) {
		t.Error("expected true comparing by public key even when the config string differs (sealed key)")
	}
}

func TestPrepareRegistrationRejectsInvalidCipherType(t *testing.T) {
	for _, ct := range []int{int(core.ECC_UMI), 99, -1} {
		result, instance, target, owned := prepareRegistration(
			nil, "", 0, "AAAA", ct,
			"u@e.com", "dev", "org", "", "asp", "res", "cl", false,
		)
		if result.ErrCode != common.ErrInvalidInput.ErrorCode() {
			t.Errorf("cipherType %d: ErrCode = %q, want %q", ct, result.ErrCode, common.ErrInvalidInput.ErrorCode())
		}
		if instance != nil || target != nil || owned {
			t.Errorf("cipherType %d: expected nil instance/target and owned=false", ct)
		}
	}
}

// TestPublicKeyByCipherSchemeConcurrentReadsAreRaceFree exercises the two reads
// PublicKeyBase64ByCipherScheme / agentCipherScheme make — the cipher scheme and
// a.device — against the writers they race in production: the config watcher
// (SetCipherScheme, under keyMu) and ReinitWithKey's device swap (under
// deviceMutex). Run under -race it proves both reads are synchronized; before
// the fix both were bare field reads and the detector flagged them.
func TestPublicKeyByCipherSchemeConcurrentReadsAreRaceFree(t *testing.T) {
	mkDev := func() *core.Device {
		d := core.NewDevice(core.NHP_AGENT, core.NewECDH(core.ECC_CURVE25519).PrivateKey(), nil)
		if d == nil {
			t.Fatal("NewDevice returned nil")
		}
		d.Start()
		t.Cleanup(d.Stop)
		return d
	}
	dev1, dev2 := mkDev(), mkDev()
	a := &UdpAgent{
		config: &Config{DefaultCipherScheme: common.CIPHER_SCHEME_CURVE},
		device: dev1,
	}
	someKey := base64.StdEncoding.EncodeToString(core.NewECDH(core.ECC_CURVE25519).PrivateKey())

	stop := make(chan struct{})
	var wg sync.WaitGroup

	// Scheme writer under keyMu — mimics the config.toml watcher.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				a.config.SetCipherScheme(common.CIPHER_SCHEME_GMSM)
				a.config.SetCipherScheme(common.CIPHER_SCHEME_CURVE)
			}
		}
	}()

	// Device writer under deviceMutex — mimics ReinitWithKey's swap window.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				a.deviceMutex.Lock()
				a.device = dev2
				a.deviceMutex.Unlock()
				a.deviceMutex.Lock()
				a.device = dev1
				a.deviceMutex.Unlock()
			}
		}
	}()

	// Readers through the guarded accessors.
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_ = a.PublicKeyBase64ByCipherScheme()
					_ = agentCipherScheme(a)
					_ = AgentHoldsRegistrationKey(a, someKey, int(core.ECC_CURVE25519))
				}
			}
		}()
	}

	time.Sleep(100 * time.Millisecond)
	close(stop)
	wg.Wait()
}

func TestPrepareRegistrationSealedKeyIsReused(t *testing.T) {
	// A running agent started from a sealed key (config string != plaintext)
	// must still be reused when the caller passes the matching plaintext key,
	// not rejected as "a different key".
	a, key := newRunningReusableAgent(t)
	a.config.PrivateKeyBase64 = "AQIDBA==sealed-blob-not-the-plaintext"

	res := RequestRegistrationOtp(
		a, "", 0, key, int(core.ECC_CURVE25519),
		"new@user", "new-device", "new-org", "", "asp", "res", "missing-cluster",
	)
	if res.ErrMsg == "close the running agent before registering a different key" {
		t.Fatalf("sealed-key agent wrongly rejected as a different key: %+v", res)
	}
	assertOriginalIdentity(t, a)
}
