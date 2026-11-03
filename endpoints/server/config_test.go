package server

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/core/verifier"
)

// newServerForReloadTest builds the minimum UdpServer state needed to
// exercise updateBaseConfig's cookie-signing-key reload path:
//   - a real *core.Device so SetStatelessCookieParams / StatelessCookieParams
//     round-trip the actual bytes (the function under test relies on the
//     read-back equality check, so a stubbed device would mask the bug)
//   - a pre-populated s.config so updateBaseConfig takes the reload branch
//     rather than the initial-population branch (which has different semantics
//     and short-circuits past the code we're testing)
//   - matched LogLevel across reloads so s.log.SetLogLevel never fires
//     (s.log is left nil; this test deliberately avoids the log path)
func newServerForReloadTest(t *testing.T, initialKey []byte) (*UdpServer, *core.Device) {
	t.Helper()

	devPriv := make([]byte, 32)
	for i := range devPriv {
		devPriv[i] = byte(i + 1)
	}
	dev := core.NewDevice(core.NHP_SERVER, devPriv, nil)
	if dev == nil {
		t.Fatal("NewDevice returned nil")
	}
	t.Cleanup(dev.Stop)

	dev.SetStatelessCookieParams(initialKey, DefaultCookieTimeWindowSeconds)

	s := &UdpServer{
		device: dev,
		config: &Config{
			CookieSigningKeyBase64:  base64.StdEncoding.EncodeToString(initialKey),
			CookieTimeWindowSeconds: DefaultCookieTimeWindowSeconds,
		},
	}
	return s, dev
}

// TestUpdateBaseConfig_ClearedKeyDoesNotPoisonInMemoryConfig fences the
// "cleared key on reload silently desyncs s.config from the running device"
// regression. Before the fix, updateBaseConfig unconditionally wrote the
// new (possibly empty) base64 into s.config.CookieSigningKeyBase64 — even
// when the "len(newKey) == 0 && keyChanged" branch had just preserved the
// running device key. The next reload then saw s.config == conf == "",
// computed keyChanged=false, and skipped the whole validation/preservation
// block. The operator stopped seeing the "cleared, keeping previous key"
// warning even though the divergence between s.config (empty) and the
// device (still holding the original key) persisted.
//
// Invariant under test: after a cleared-key reload, s.config still names
// the key that is actually live on the device, so a subsequent reload
// (even of the same empty value, or of an unrelated field) re-enters the
// block and re-emits the operator-facing signal.
func TestUpdateBaseConfig_ClearedKeyDoesNotPoisonInMemoryConfig(t *testing.T) {
	k1 := bytes.Repeat([]byte{0x11}, 32)
	k1b64 := base64.StdEncoding.EncodeToString(k1)
	s, dev := newServerForReloadTest(t, k1)

	// Reload 1: operator clears CookieSigningKeyBase64 on disk. The
	// preservation branch must keep the running device key (k1), and —
	// per the fix — must NOT write the empty string back into s.config.
	if err := s.updateBaseConfig(Config{
		CookieSigningKeyBase64:  "",
		CookieTimeWindowSeconds: DefaultCookieTimeWindowSeconds,
	}); err != nil {
		t.Fatalf("clear-reload returned error: %v", err)
	}

	devKey, _ := dev.StatelessCookieParams()
	if !bytes.Equal(devKey, k1) {
		t.Fatalf("device key must be preserved across a cleared reload; got %x, want %x", devKey, k1)
	}
	if s.config.CookieSigningKeyBase64 != k1b64 {
		t.Fatalf("s.config.CookieSigningKeyBase64 must still name the running key after a cleared reload — otherwise the next reload sees no delta and the warning vanishes while the desync persists.\ngot  = %q\nwant = %q (the original)",
			s.config.CookieSigningKeyBase64, k1b64)
	}
}

// TestUpdateBaseConfig_RefillAfterClearReachesDevice is the
// operator-recovery story: after a cleared-key reload, the operator
// fixes their config and reloads with a real (different) key. That
// must reach the device — not get masked by stale s.config state.
//
// Pre-fix this path already worked (keyChanged was true because
// s.config had been written to "" and the new value differs), so a
// no-fix run also passes this test. We keep it as documentation of
// the end-to-end recovery flow rather than as a sharp regression
// fence; the sharp fence is the cleared-config-write-back test above.
func TestUpdateBaseConfig_RefillAfterClearReachesDevice(t *testing.T) {
	k1 := bytes.Repeat([]byte{0x11}, 32)
	k2 := bytes.Repeat([]byte{0x22}, 32)
	k2b64 := base64.StdEncoding.EncodeToString(k2)
	s, dev := newServerForReloadTest(t, k1)

	// Clear.
	if err := s.updateBaseConfig(Config{
		CookieSigningKeyBase64:  "",
		CookieTimeWindowSeconds: DefaultCookieTimeWindowSeconds,
	}); err != nil {
		t.Fatalf("clear-reload returned error: %v", err)
	}

	// Refill with a different key — must reach the device.
	if err := s.updateBaseConfig(Config{
		CookieSigningKeyBase64:  k2b64,
		CookieTimeWindowSeconds: DefaultCookieTimeWindowSeconds,
	}); err != nil {
		t.Fatalf("refill-reload returned error: %v", err)
	}

	devKey, _ := dev.StatelessCookieParams()
	if !bytes.Equal(devKey, k2) {
		t.Fatalf("device key must adopt the refilled key; got %x, want %x", devKey, k2)
	}
	if s.config.CookieSigningKeyBase64 != k2b64 {
		t.Fatalf("s.config must adopt the refilled key; got %q, want %q",
			s.config.CookieSigningKeyBase64, k2b64)
	}
}

// TestUpdateBaseConfig_MalformedKeyDoesNotPoisonInMemoryConfig is the
// same invariant as the cleared-key test, applied to the decode-error
// branch. Pre-fix, a malformed base64 still got written back into
// s.config (line 619), so the next reload of the same bad value
// computed keyChanged=false and skipped the warning. With the fix,
// only well-formed non-empty values are persisted, so a malformed
// reload re-emits the "ignoring CookieSigningKeyBase64 change"
// warning on every subsequent attempt until the operator fixes it.
func TestUpdateBaseConfig_MalformedKeyDoesNotPoisonInMemoryConfig(t *testing.T) {
	k1 := bytes.Repeat([]byte{0x11}, 32)
	k1b64 := base64.StdEncoding.EncodeToString(k1)
	s, dev := newServerForReloadTest(t, k1)

	if err := s.updateBaseConfig(Config{
		CookieSigningKeyBase64:  "not-valid-base64!!!",
		CookieTimeWindowSeconds: DefaultCookieTimeWindowSeconds,
	}); err != nil {
		t.Fatalf("malformed-reload returned error: %v", err)
	}

	devKey, _ := dev.StatelessCookieParams()
	if !bytes.Equal(devKey, k1) {
		t.Fatalf("device key must be preserved across a malformed-key reload; got %x, want %x", devKey, k1)
	}
	if s.config.CookieSigningKeyBase64 != k1b64 {
		t.Fatalf("s.config must not adopt a malformed base64 string — that hides the validation warning on subsequent reloads.\ngot  = %q\nwant = %q",
			s.config.CookieSigningKeyBase64, k1b64)
	}
}

// TestUpdateBaseConfig_WindowOnlyChangePersists guards that the
// fail-close on the key write-back does not block window updates.
// Window is plain numeric, re-validated each reload, and writing it
// back is the only way an operator can durably tune the cookie
// lifetime. So even on the cleared-key path, the window must still
// land in s.config.
func TestUpdateBaseConfig_WindowOnlyChangePersists(t *testing.T) {
	k1 := bytes.Repeat([]byte{0x11}, 32)
	s, _ := newServerForReloadTest(t, k1)

	const newWin = DefaultCookieTimeWindowSeconds + 30
	if err := s.updateBaseConfig(Config{
		CookieSigningKeyBase64:  s.config.CookieSigningKeyBase64,
		CookieTimeWindowSeconds: newWin,
	}); err != nil {
		t.Fatalf("window-only reload returned error: %v", err)
	}
	if s.config.CookieTimeWindowSeconds != newWin {
		t.Fatalf("window update did not persist; got %d, want %d", s.config.CookieTimeWindowSeconds, newWin)
	}
}

// TestUpdateBaseConfig_WindowOnlyReloadPreservesRandomKey fences the
// single-instance random-key regression. In production, an operator who
// leaves CookieSigningKeyBase64 unset gets a random per-process key
// minted by udpserver.Start; s.config.CookieSigningKeyBase64 stays "".
//
// A later window-only reload (CookieSigningKeyBase64 still "") computes
// keyChanged=false. Pre-fix, the preservation branch was gated on
// "len(newKey)==0 && keyChanged", so it was skipped, newKey stayed nil,
// and SetStatelessCookieParams(nil, win) DISABLED stateless cookies —
// silently stalling every agent that hit the overload path. The fix
// preserves the running key whenever the config field is empty,
// regardless of keyChanged.
//
// Note this test sets up the state directly rather than via
// newServerForReloadTest, which always populates a non-empty base64
// (and so could never reproduce the always-empty production flow).
func TestUpdateBaseConfig_WindowOnlyReloadPreservesRandomKey(t *testing.T) {
	devPriv := make([]byte, 32)
	for i := range devPriv {
		devPriv[i] = byte(i + 1)
	}
	dev := core.NewDevice(core.NHP_SERVER, devPriv, nil)
	if dev == nil {
		t.Fatal("NewDevice returned nil")
	}
	t.Cleanup(dev.Stop)

	// Mimic udpserver.Start's single-instance path: a random key on the
	// device, but an EMPTY CookieSigningKeyBase64 in s.config.
	randomKey := bytes.Repeat([]byte{0xAB}, 32)
	dev.SetStatelessCookieParams(randomKey, DefaultCookieTimeWindowSeconds)
	s := &UdpServer{
		device: dev,
		config: &Config{
			CookieSigningKeyBase64:  "",
			CookieTimeWindowSeconds: DefaultCookieTimeWindowSeconds,
		},
	}

	const newWin = DefaultCookieTimeWindowSeconds + 30
	if err := s.updateBaseConfig(Config{
		CookieSigningKeyBase64:  "", // still unset, as on disk
		CookieTimeWindowSeconds: newWin,
	}); err != nil {
		t.Fatalf("window-only reload returned error: %v", err)
	}

	devKey, devWin := dev.StatelessCookieParams()
	if len(devKey) == 0 {
		t.Fatal("window-only reload disabled stateless cookies (device key nil'd); the random per-process key must be preserved")
	}
	if !bytes.Equal(devKey, randomKey) {
		t.Fatalf("device key must be the preserved random key; got %x, want %x", devKey, randomKey)
	}
	if devWin != int64(newWin) {
		t.Fatalf("window update must still reach the device; got %d, want %d", devWin, newWin)
	}
}

// TestShippedDemoCookieSigningKey_MatchesCommittedConfig pins the
// constant we compare against at startup to the value actually
// committed under docker/nhp-server/etc/config.toml. The startup
// warning is only useful if those two stay aligned — if the demo
// config rotates to a new key and we forget to update the constant,
// the warning silently stops firing and we ship the new "demo"
// secret in real deployments without anyone noticing.
//
// docker/nhp-server/etc2/config.toml is intentionally the same key
// (the multi-instance demo relies on both replicas sharing the
// cookie signing key); cross-check both.
func TestShippedDemoCookieSigningKey_MatchesCommittedConfig(t *testing.T) {
	// Walk up from the test's working dir to the repo root, then read
	// the demo configs. The package lives at endpoints/server, so the
	// repo root is two levels up — but use filepath.Abs + a directory
	// climb so this works regardless of how `go test` was invoked.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	root := cwd
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(root, "docker", "nhp-server", "etc", "config.toml")); err == nil {
			break
		}
		root = filepath.Dir(root)
	}

	for _, rel := range []string{
		"docker/nhp-server/etc/config.toml",
		"docker/nhp-server/etc2/config.toml",
	} {
		t.Run(rel, func(t *testing.T) {
			path := filepath.Join(root, rel)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Skipf("cannot read %s: %v", path, err)
			}
			// Look for an uncommented `CookieSigningKeyBase64 = "..."`
			// line. We don't pull in a TOML parser here on purpose —
			// the demo files are stable, this check is purely "did
			// the literal value drift", and reaching for a parser
			// would invite parsing surprises the regression itself
			// doesn't need.
			var got string
			for _, line := range strings.Split(string(data), "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "#") {
					continue
				}
				if !strings.HasPrefix(trimmed, "CookieSigningKeyBase64") {
					continue
				}
				// "CookieSigningKeyBase64 = \"...\"" — pull out the
				// value between the first pair of double quotes.
				first := strings.IndexByte(trimmed, '"')
				last := strings.LastIndexByte(trimmed, '"')
				if first < 0 || last <= first {
					t.Fatalf("%s: malformed line %q", rel, trimmed)
				}
				got = trimmed[first+1 : last]
				break
			}
			if got == "" {
				t.Fatalf("%s: no CookieSigningKeyBase64 line found", rel)
			}
			if got != shippedDemoCookieSigningKeyBase64 {
				t.Fatalf("%s ships CookieSigningKeyBase64=%q but the demo-key constant is %q — "+
					"update shippedDemoCookieSigningKeyBase64 in config.go (or revert the demo config), "+
					"otherwise the startup warning will silently stop catching copy-paste deployments.",
					rel, got, shippedDemoCookieSigningKeyBase64)
			}
		})
	}
}

// TestUpdateBaseConfig_ToggleValidationKeepsDeviceHooks fences the
// hot-reload clobber: toggling DisableAgentValidation must not wipe the
// OnPacketDropped hook (or PeerLookupFallback) that Start installed on the
// device. SetOption replaces the whole struct, so updateBaseConfig has to
// read-modify-write.
func TestUpdateBaseConfig_ToggleValidationKeepsDeviceHooks(t *testing.T) {
	devPriv := make([]byte, 32)
	for i := range devPriv {
		devPriv[i] = byte(i + 1)
	}
	dev := core.NewDevice(core.NHP_SERVER, devPriv, nil)
	if dev == nil {
		t.Fatal("NewDevice returned nil")
	}
	t.Cleanup(dev.Stop)

	dropped := 0
	opt := dev.GetOption()
	opt.OnPacketDropped = func(string) { dropped++ }
	dev.SetOption(opt)

	s := &UdpServer{
		device: dev,
		config: &Config{DisableAgentValidation: false},
	}

	if err := s.updateBaseConfig(Config{DisableAgentValidation: true}); err != nil {
		t.Fatalf("updateBaseConfig: %v", err)
	}

	got := dev.GetOption()
	if !got.DisableAgentPeerValidation {
		t.Fatal("DisableAgentPeerValidation was not applied on reload")
	}
	if got.OnPacketDropped == nil {
		t.Fatal("OnPacketDropped hook was wiped by the reload")
	}
	got.OnPacketDropped("decrypt")
	if dropped != 1 {
		t.Fatalf("surviving hook did not fire: dropped=%d", dropped)
	}
}

// TestUpdateBaseConfig_AttestationSchemeReloadResetsToCSV fences the
// "operator removes AttestationScheme = 'test' from config.toml and the
// server keeps accepting self-asserted evidence" regression.
//
// Pre-fix, loadBaseConfig's file-watch branch re-unmarshalled into the
// same outer `config` variable. go-toml v2 does not zero fields that
// are absent from the new file, so the removed `AttestationScheme`
// key stayed "test" in memory, `s.config.AttestationScheme != conf.AttestationScheme`
// was false on every subsequent reload, and the resolved scheme stayed
// SchemeTest until a restart. The fix unmarshals each reload into a fresh
// Config struct, so a removed key arrives in updateBaseConfig as "" —
// which equals the s.config slot only when the value already matches, and
// in this scenario it does not.
//
// Invariant under test: a server that booted with AttestationScheme =
// "test" must adopt SchemeCSV when the next config.toml drops the key.
// This is the only behavior change from the fresh-struct fix that the
// existing tests do not already pin.
func TestUpdateBaseConfig_AttestationSchemeReloadResetsToCSV(t *testing.T) {
	s := &UdpServer{
		config: &Config{AttestationScheme: "test"},
	}
	// Mirror what udpserver.Start stores at boot so we can observe the
	// resolved-scheme transition end-to-end (not just the raw field).
	s.attestationScheme.Store(string(verifier.SchemeTest))

	// The reload callback now builds `newConf` itself, so simulate that
	// path: conf.AttestationScheme is absent (empty string) because the
	// operator deleted the line from config.toml.
	if err := s.updateBaseConfig(Config{AttestationScheme: ""}); err != nil {
		t.Fatalf("reload after AttestationScheme removed: %v", err)
	}

	if s.config.AttestationScheme != "" {
		t.Fatalf("removed AttestationScheme key must clear s.config.AttestationScheme; got %q, want \"\"",
			s.config.AttestationScheme)
	}
	got, _ := s.attestationScheme.Load().(string)
	if got != string(verifier.SchemeCSV) {
		t.Fatalf("removed AttestationScheme key must resolve to SchemeCSV; got %q, want %q",
			got, verifier.SchemeCSV)
	}
}

// TestUpdateBaseConfig_AttestationSchemeReloadSticksToCSV guards the
// follow-up reload — after the first transition off "test", a second
// reload of the same empty file must NOT flip the resolved scheme back
// to "test". This is what makes the removal stick: the previous fix's
// failure mode was "reload to empty state key, then any further reload
// still says AttestationScheme != conf is true", which would have
// produced a steady stream of misleading "AttestationScheme set to"
// Info logs after the operator removed the demo switch. The fresh-struct
// reload path lands the empty AttestationScheme in both s.config and
// conf, so the second reload computes `s.config.AttestationScheme ==
// conf.AttestationScheme` and does not re-enter the log/store branch.
func TestUpdateBaseConfig_AttestationSchemeReloadSticksToCSV(t *testing.T) {
	s := &UdpServer{
		config: &Config{AttestationScheme: "test"},
	}
	s.attestationScheme.Store(string(verifier.SchemeTest))

	if err := s.updateBaseConfig(Config{AttestationScheme: ""}); err != nil {
		t.Fatalf("first reload: %v", err)
	}
	// Second reload — same empty AttestationScheme — must be a true no-op.
	if err := s.updateBaseConfig(Config{AttestationScheme: ""}); err != nil {
		t.Fatalf("second reload: %v", err)
	}

	if s.config.AttestationScheme != "" {
		t.Fatalf("AttestationScheme must stay cleared after a follow-up reload; got %q", s.config.AttestationScheme)
	}
	got, _ := s.attestationScheme.Load().(string)
	if got != string(verifier.SchemeCSV) {
		t.Fatalf("resolved scheme must stay at SchemeCSV after a follow-up reload; got %q, want %q",
			got, verifier.SchemeCSV)
	}
}

// Every shipped http.toml must parse, and none of them may name a trusted
// proxy. The shipped files bind 0.0.0.0, so a non-empty list there would make
// the spoofable-source-IP case (CWE-348) reachable by connecting to the
// listener directly, past whatever proxy the list describes. Operators opt in
// per deployment; the repo never ships the opt-in.
func TestShippedHttpConfigs_TrustNoProxy(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	root := cwd
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(root, "docker", "nhp-server", "etc", "http.toml")); err == nil {
			break
		}
		root = filepath.Dir(root)
	}

	for _, rel := range []string{
		"endpoints/server/main/etc/http.toml",
		"docker/nhp-server/etc/http.toml",
		"docker/nhp-server/etc2/http.toml",
		"docker/multicluster/nhp-server-c2/etc/http.toml",
		"deploy/config-templates/server/http.toml",
	} {
		t.Run(rel, func(t *testing.T) {
			path := filepath.Join(root, rel)
			data, err := os.ReadFile(path)
			if err != nil {
				// Fatal, not Skip: these paths are fixed by the repo layout,
				// so an unreadable one means a file moved and this regression
				// check stopped covering it rather than that it passed.
				t.Fatalf("cannot read %s: %v", path, err)
			}
			var conf HttpConfig
			if err := toml.Unmarshal(data, &conf); err != nil {
				t.Fatalf("unmarshal %s: %v", rel, err)
			}
			if len(conf.TrustedProxies) != 0 {
				t.Fatalf("%s sets TrustedProxies = %v; shipped configs must trust no proxy", rel, conf.TrustedProxies)
			}
		})
	}
}

// The TOML key has to reach the field: a typo in the struct tag or the shipped
// comment block would leave TrustedProxies silently empty, which fails safe
// but would also make the opt-in impossible to actually use.
func TestHttpConfig_TrustedProxiesRoundTripsThroughTOML(t *testing.T) {
	var conf HttpConfig
	if err := toml.Unmarshal([]byte("TrustedProxies = [\"127.0.0.1\", \"10.1.0.0/16\"]\n"), &conf); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	want := []string{"127.0.0.1", "10.1.0.0/16"}
	if len(conf.TrustedProxies) != len(want) {
		t.Fatalf("TrustedProxies = %v, want %v", conf.TrustedProxies, want)
	}
	for i := range want {
		if conf.TrustedProxies[i] != want[i] {
			t.Fatalf("TrustedProxies = %v, want %v", conf.TrustedProxies, want)
		}
	}

	// A file with no TrustedProxies key at all must leave the list empty,
	// not inherit anything.
	var absent HttpConfig
	if err := toml.Unmarshal([]byte("EnableHttp = true\n"), &absent); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(absent.TrustedProxies) != 0 {
		t.Fatalf("TrustedProxies = %v with the key absent, want empty", absent.TrustedProxies)
	}
}

// writeHttpToml writes etc/http.toml under a temporary ExeDirPath and returns
// the file path. The caller is responsible for restoring ExeDirPath.
func writeHttpToml(t *testing.T, dir, body string) string {
	t.Helper()
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", etc, err)
	}
	path := filepath.Join(etc, "http.toml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// Deleting TrustedProxies from http.toml must stop trusting the proxy.
//
// go-toml v2, like encoding/json, leaves struct fields the document does not
// mention untouched, so the reload watcher has to decode into a fresh
// HttpConfig. Decoding into one struct reused across reloads would let the
// list from an earlier file survive an edit that removes the key -- an
// operator who deletes the line to stop trusting a proxy would keep trusting
// it, which is precisely the fail-open the field exists to close.
//
// EnableHttp stays false throughout so no listener is bound; the assertion is
// on the config the reload leaves behind.
func TestLoadHttpConfig_ReloadWithoutTrustedProxiesClearsTheList(t *testing.T) {
	dir := t.TempDir()
	prevExeDir := ExeDirPath
	ExeDirPath = dir
	t.Cleanup(func() {
		ExeDirPath = prevExeDir
		if httpConfigWatch != nil {
			_ = httpConfigWatch.Close()
			httpConfigWatch = nil
		}
	})

	path := writeHttpToml(t, dir, "EnableHttp = false\nTrustedProxies = [\"10.0.0.0/8\"]\n")

	s := &UdpServer{}
	if err := s.loadHttpConfig(); err != nil {
		t.Fatalf("loadHttpConfig: %v", err)
	}
	if got := s.httpConfig.TrustedProxies; len(got) != 1 || got[0] != "10.0.0.0/8" {
		t.Fatalf("after the initial load TrustedProxies = %v, want [10.0.0.0/8]", got)
	}

	// Same file, key removed. This is the documented way to stop trusting a
	// proxy, so the reload must land on an empty list.
	if err := os.WriteFile(path, []byte("EnableHttp = false\n"), 0o600); err != nil {
		t.Fatalf("rewrite %s: %v", path, err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for {
		if len(s.httpConfig.TrustedProxies) == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("after a reload that removed the key TrustedProxies = %v, want empty; "+
				"the watcher must decode into a fresh HttpConfig", s.httpConfig.TrustedProxies)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// The same reload path must also pick up a list that changed rather than one
// that was deleted, so the fresh-struct decode is not just zeroing everything.
func TestLoadHttpConfig_ReloadReplacesTrustedProxies(t *testing.T) {
	dir := t.TempDir()
	prevExeDir := ExeDirPath
	ExeDirPath = dir
	t.Cleanup(func() {
		ExeDirPath = prevExeDir
		if httpConfigWatch != nil {
			_ = httpConfigWatch.Close()
			httpConfigWatch = nil
		}
	})

	path := writeHttpToml(t, dir, "EnableHttp = false\nTrustedProxies = [\"10.0.0.0/8\"]\n")

	s := &UdpServer{}
	if err := s.loadHttpConfig(); err != nil {
		t.Fatalf("loadHttpConfig: %v", err)
	}

	if err := os.WriteFile(path, []byte("EnableHttp = false\nTrustedProxies = [\"127.0.0.1\"]\n"), 0o600); err != nil {
		t.Fatalf("rewrite %s: %v", path, err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for {
		got := s.httpConfig.TrustedProxies
		if len(got) == 1 && got[0] == "127.0.0.1" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("after a reload that replaced the key TrustedProxies = %v, want [127.0.0.1]", got)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// A malformed http.toml must not leave a half-decoded config applied: go-toml
// v2 populates the fields ahead of the error and stops, so applying the result
// would start a listener from settings nobody wrote.
func TestLoadHttpConfig_MalformedFileIsNotApplied(t *testing.T) {
	dir := t.TempDir()
	prevExeDir := ExeDirPath
	ExeDirPath = dir
	t.Cleanup(func() {
		ExeDirPath = prevExeDir
		if httpConfigWatch != nil {
			_ = httpConfigWatch.Close()
			httpConfigWatch = nil
		}
	})

	writeHttpToml(t, dir, "TrustedProxies = [\"10.0.0.0/8\"]\nEnableHttp = true\nthis is not toml\n")

	s := &UdpServer{}
	if err := s.loadHttpConfig(); err != nil {
		t.Fatalf("loadHttpConfig: %v", err)
	}
	if s.httpConfig != nil {
		t.Fatalf("a malformed http.toml was applied as %+v, want no config", s.httpConfig)
	}
	if s.httpServer != nil {
		t.Fatal("a malformed http.toml started an HTTP server")
	}
}

// The etcd config path decodes a fresh ServerEtcdConfig per call and bails on
// a decode error, so it does not carry a stale TrustedProxies across reloads
// the way loadHttpConfig's watcher used to. Confirm both halves, since the
// field is now a security property and the next person to touch this function
// should find out from a test rather than from a mis-attributed AC rule.
func TestUpdateEtcdConfig_HttpTrustedProxiesDoesNotPersistAcrossReloads(t *testing.T) {
	s := &UdpServer{}

	withList := "[HttpConfig]\nEnableHttp = false\nTrustedProxies = [\"10.0.0.0/8\"]\n"
	if err := s.updateEtcdConfig([]byte(withList), false); err != nil {
		t.Fatalf("updateEtcdConfig: %v", err)
	}
	if got := s.httpConfig.TrustedProxies; len(got) != 1 || got[0] != "10.0.0.0/8" {
		t.Fatalf("TrustedProxies = %v after the first load, want [10.0.0.0/8]", got)
	}

	withoutList := "[HttpConfig]\nEnableHttp = false\n"
	if err := s.updateEtcdConfig([]byte(withoutList), false); err != nil {
		t.Fatalf("updateEtcdConfig: %v", err)
	}
	if got := s.httpConfig.TrustedProxies; len(got) != 0 {
		t.Fatalf("TrustedProxies = %v after a reload that removed the key, want empty", got)
	}

	// A document that does not decode must leave the running config alone
	// rather than apply a half-decoded one.
	if err := s.updateEtcdConfig([]byte(withList), false); err != nil {
		t.Fatalf("updateEtcdConfig: %v", err)
	}
	if err := s.updateEtcdConfig([]byte("[HttpConfig]\nthis is not toml\n"), false); err == nil {
		t.Fatal("updateEtcdConfig on a malformed document = nil, want an error")
	}
	if got := s.httpConfig.TrustedProxies; len(got) != 1 || got[0] != "10.0.0.0/8" {
		t.Fatalf("TrustedProxies = %v after a refused reload, want the running [10.0.0.0/8]", got)
	}
}
