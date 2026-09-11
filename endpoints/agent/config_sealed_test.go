package agent

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/keystore"
)

// clearPassphraseFileEnv makes the inline NHP_KEY_PASSPHRASE the only source
// during a test, so a developer's exported NHP_KEY_PASSPHRASE_FILE (which
// wins in PassphraseFromEnv) cannot make these tests non-deterministic.
func clearPassphraseFileEnv(t *testing.T) {
	t.Helper()
	t.Setenv(keystore.EnvPassphraseFile, "")
}

func mustSeal(t *testing.T, raw, pass []byte) string {
	t.Helper()
	blob, err := keystore.Seal(raw, pass)
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	return blob
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// TestGetAgentEcdhWithSealedKey guards the fix for the /publicKey handler:
// when the agent private key is stored as a sealed blob, GetAgentEcdh must
// still report the real device public key, not a key derived from a broken
// base64 decode of the "v1$..." blob.
func TestGetAgentEcdhWithSealedKey(t *testing.T) {
	// A known raw private key and the public key its device would present.
	e := core.NewECDH(core.ECC_CURVE25519)
	raw := e.PrivateKey()
	wantPub := core.ECDHFromKey(core.ECC_CURVE25519, raw).PublicKeyBase64()

	blob, err := keystore.Seal(raw, []byte("test-passphrase"))
	if err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    blob,
	}

	// Cached path (as populated by Start / ReinitWithKey): the resolved key
	// is used directly, NOT a broken base64 decode of the "v1$..." blob.
	cfg.SetResolvedPrivateKey(raw)
	if got := cfg.GetAgentEcdh().PublicKeyBase64(); got != wantPub {
		t.Fatalf("cached path: got %q want %q", got, wantPub)
	}

	// Empty cache: there is deliberately no resolve-on-demand fallback (it
	// would run a 64 MiB Argon2 in an HTTP handler). GetAgentEcdh returns
	// nil and getAgentPublicKey answers 500.
	cfg.SetResolvedPrivateKey(nil)
	if got := cfg.GetAgentEcdh(); got != nil {
		t.Fatalf("empty cache: expected nil Ecdh, got %v", got)
	}
}

// TestRotateAgentKeyPreservesSealed guards against the rotate path silently
// downgrading a sealed key to plaintext-at-rest. After rotation the config
// value must still be a sealed blob that unseals with the same passphrase.
func TestRotateAgentKeyPreservesSealed(t *testing.T) {
	clearPassphraseFileEnv(t)
	t.Setenv(keystore.EnvPassphrase, "correct horse battery staple")

	dir := t.TempDir()
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}

	pass := []byte("correct horse battery staple")
	orig := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	sealed := mustSeal(t, orig, pass)
	cfgBody := "DefaultCipherScheme = 0\nPrivateKeyBase64 = \"" + sealed + "\"\n"
	mustWrite(t, filepath.Join(etc, "config.toml"), cfgBody)

	oldExe := ExeDirPath
	ExeDirPath = dir
	defer func() { ExeDirPath = oldExe }()

	a := &UdpAgent{}
	if rotErr := a.RotateAgentKey(); rotErr != nil {
		t.Fatalf("RotateAgentKey: %v", rotErr)
	}

	raw := mustRead(t, filepath.Join(etc, "config.toml"))
	newVal := tomlStringValue(t, raw, "PrivateKeyBase64")
	if !keystore.IsSealed(newVal) {
		t.Fatalf("rotated key is no longer sealed: %q", newVal)
	}
	if newVal == sealed {
		t.Fatal("rotated key blob is identical to the original — key was not rotated")
	}
	rotatedRaw, openErr := keystore.Open(newVal, pass)
	if openErr != nil {
		t.Fatalf("rotated blob does not unseal with the configured passphrase: %v", openErr)
	}
	if len(rotatedRaw) != 32 {
		t.Fatalf("rotated key is %d bytes, want 32", len(rotatedRaw))
	}
}

// TestRotateAgentKeyPlainStaysPlain confirms a plain-base64 key still
// rotates to a plain-base64 key (no accidental sealing).
func TestRotateAgentKeyPlainStaysPlain(t *testing.T) {
	dir := t.TempDir()
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}
	orig := core.NewECDH(core.ECC_CURVE25519).PrivateKey()
	cfgBody := "DefaultCipherScheme = 0\nPrivateKeyBase64 = \"" + base64.StdEncoding.EncodeToString(orig) + "\"\n"
	mustWrite(t, filepath.Join(etc, "config.toml"), cfgBody)

	oldExe := ExeDirPath
	ExeDirPath = dir
	defer func() { ExeDirPath = oldExe }()

	a := &UdpAgent{}
	if err := a.RotateAgentKey(); err != nil {
		t.Fatalf("RotateAgentKey: %v", err)
	}
	val := tomlStringValue(t, mustRead(t, filepath.Join(etc, "config.toml")), "PrivateKeyBase64")
	if keystore.IsSealed(val) {
		t.Fatalf("plain key was unexpectedly sealed on rotate: %q", val)
	}
	if _, err := base64.StdEncoding.DecodeString(val); err != nil {
		t.Fatalf("rotated plain key is not valid base64: %v", err)
	}
}

// tomlStringValue returns the value of `key = "..."` from a small TOML body.
func tomlStringValue(t *testing.T, body, key string) string {
	t.Helper()
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, key+" ") && !strings.HasPrefix(line, key+"=") {
			continue
		}
		if i := strings.Index(line, "\""); i >= 0 {
			if j := strings.LastIndex(line, "\""); j > i {
				return line[i+1 : j]
			}
		}
	}
	t.Fatalf("key %q not found in TOML body:\n%s", key, body)
	return ""
}

// TestGetAgentEcdhPlainKeyUnchanged confirms the plain-base64 path is
// unaffected by the sealed-key support.
func TestGetAgentEcdhPlainKeyUnchanged(t *testing.T) {
	e := core.NewECDH(core.ECC_CURVE25519)
	raw := e.PrivateKey()
	wantPub := core.ECDHFromKey(core.ECC_CURVE25519, raw).PublicKeyBase64()

	cfg := &Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    base64.StdEncoding.EncodeToString(raw),
	}
	if got := cfg.GetAgentEcdh().PublicKeyBase64(); got != wantPub {
		t.Fatalf("plain path: got %q want %q", got, wantPub)
	}
}

// TestConfigKeyMuCoversCipherSchemeAndPrivateKeyReads: run with -race. The
// config-reload watcher (updateBaseConfig) writes DefaultCipherScheme via
// SetCipherScheme, and ReinitWithKey writes both PrivateKeyBase64 and
// DefaultCipherScheme via SetPrivateKeyMaterial — both under keyMu. A
// concurrent /publicKey request reads them via GetAgentEcdh, and
// UdpAgent.PrivateKeyBase64() reads PrivateKeyBase64 via
// GetPrivateKeyBase64. Every read and write here must go through one of
// those, or this test races.
func TestConfigKeyMuCoversCipherSchemeAndPrivateKeyReads(t *testing.T) {
	e := core.NewECDH(core.ECC_CURVE25519)
	cfg := &Config{
		DefaultCipherScheme: common.CIPHER_SCHEME_CURVE,
		PrivateKeyBase64:    base64.StdEncoding.EncodeToString(e.PrivateKey()),
	}
	cfg.SetResolvedPrivateKey(e.PrivateKey())

	done := make(chan struct{})
	var wg sync.WaitGroup

	// Writers: the config-reload watcher (SetCipherScheme alone) and a
	// rekey (SetPrivateKeyMaterial, all three fields together).
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-done:
				return
			default:
			}
			cfg.SetCipherScheme(i % 2)
		}
	}()
	go func() {
		defer wg.Done()
		raw := e.PrivateKey()
		b64 := base64.StdEncoding.EncodeToString(raw)
		for {
			select {
			case <-done:
				return
			default:
			}
			cfg.SetPrivateKeyMaterial(b64, common.CIPHER_SCHEME_CURVE, raw)
		}
	}()

	// Readers: the two paths the review flagged.
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
			}
			cfg.GetAgentEcdh()
		}
	}()
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
			}
			_ = cfg.GetPrivateKeyBase64()
		}
	}()

	time.Sleep(20 * time.Millisecond)
	close(done)
	wg.Wait()
}
