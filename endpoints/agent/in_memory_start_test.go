package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStartWithPrivateKeyDoesNotMaterializeConfig(t *testing.T) {
	dir := t.TempDir()
	privateKey := make([]byte, 32)
	for i := range privateKey {
		privateKey[i] = byte(i + 1)
	}
	privateKey[0] &= 248
	privateKey[31] = (privateKey[31] & 127) | 64

	a := &UdpAgent{}
	if err := a.StartWithPrivateKey(dir, 0, privateKey); err != nil {
		t.Fatalf("StartWithPrivateKey: %v", err)
	}
	t.Cleanup(a.Stop)

	if !a.IsRunning() {
		t.Fatal("agent is not running")
	}
	for _, name := range []string{"config.toml", "dhp.toml", "server.toml", "resource.toml"} {
		path := filepath.Join(dir, "etc", name)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("%s exists after in-memory start (err=%v)", path, err)
		}
	}
}

func TestStartWithPrivateKeyRejectsWrongSize(t *testing.T) {
	a := &UdpAgent{}
	if err := a.StartWithPrivateKey(t.TempDir(), 0, make([]byte, 31)); err == nil {
		t.Fatal("StartWithPrivateKey accepted a 31-byte key")
	}
}
