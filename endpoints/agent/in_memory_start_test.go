package agent

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
)

func TestStartWithPrivateKeyDoesNotMaterializeConfig(t *testing.T) {
	dir := t.TempDir()
	privateKey := make([]byte, 32)
	for i := range privateKey {
		privateKey[i] = byte(i + 1)
	}
	privateKey[0] &= 248
	privateKey[31] = (privateKey[31] & 127) | 64
	wantPrivateKey := bytes.Clone(privateKey)

	a := &UdpAgent{}
	if err := a.StartWithPrivateKey(dir, 4, privateKey); err != nil {
		t.Fatalf("StartWithPrivateKey: %v", err)
	}
	t.Cleanup(a.Stop)
	for i := range privateKey {
		privateKey[i] = 0
	}

	if !a.IsRunning() {
		t.Fatal("agent is not running")
	}
	if a.config.LogLevel != 4 {
		t.Fatalf("LogLevel = %d, want 4", a.config.LogLevel)
	}
	if a.config.DefaultCipherScheme != common.CIPHER_SCHEME_CURVE {
		t.Fatalf("DefaultCipherScheme = %d, want Curve25519", a.config.DefaultCipherScheme)
	}
	if got := a.device.GetEcdhByCipherScheme(common.CIPHER_SCHEME_CURVE).PrivateKey(); !bytes.Equal(got, wantPrivateKey) {
		t.Fatal("agent private key changed when the caller buffer was scrubbed")
	}
	for _, name := range []string{"config.toml", "dhp.toml", "server.toml", "resource.toml"} {
		path := filepath.Join(dir, "etc", name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%s exists after in-memory start (err=%v)", path, err)
		}
	}
}

func TestStartWithPrivateKeyIgnoresPersistedConfig(t *testing.T) {
	dir := t.TempDir()
	etcDir := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etcDir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	for _, name := range []string{"config.toml", "dhp.toml", "server.toml", "resource.toml"} {
		if err := os.WriteFile(filepath.Join(etcDir, name), []byte("not valid toml"), 0o600); err != nil {
			t.Fatalf("WriteFile(%s): %v", name, err)
		}
	}

	privateKey := make([]byte, 32)
	for i := range privateKey {
		privateKey[i] = byte(i + 1)
	}
	privateKey[0] &= 248
	privateKey[31] = (privateKey[31] & 127) | 64

	a := &UdpAgent{}
	if err := a.StartWithPrivateKey(dir, 0, privateKey); err != nil {
		t.Fatalf("StartWithPrivateKey read persisted config: %v", err)
	}
	t.Cleanup(a.Stop)
}

func TestStartWithPrivateKeyRejectsWrongSize(t *testing.T) {
	a := &UdpAgent{}
	if err := a.StartWithPrivateKey(t.TempDir(), 0, make([]byte, 31)); err == nil {
		t.Fatal("StartWithPrivateKey accepted a 31-byte key")
	}
}
