package server

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/core"
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
