package ac

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/endpoints/ac/ebpf"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// A Start that fails after the XDP program is loaded and the HTTPS knock
// listener is up must undo both: the CLI only cleans up after a good Start.
func TestStart_FailedInitialLoadCleansUpXDPAndHTTPS(t *testing.T) {
	dir := t.TempDir()
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}
	writeSelfSignedCert(t, filepath.Join(etc, "cert.pem"), filepath.Join(etc, "key.pem"))
	port := freeTCPPort(t)

	prk := core.NewECDH(core.ECC_CURVE25519).PrivateKeyBase64()
	writeFile(t, filepath.Join(etc, "config.toml"), fmt.Sprintf("PrivateKeyBase64 = %q\nACId = \"ac-1\"\nFilterMode = %d\n", prk, FilterMode_EBPFXDP))
	writeFile(t, filepath.Join(etc, "http.toml"), fmt.Sprintf("EnableHttp = true\nEnableTLS = true\nHttpListenPort = %d\nTLSCertFile = \"etc/cert.pem\"\nTLSKeyFile = \"etc/key.pem\"\n", port))
	// No [[Servers]]: the initial load refuses to start.
	writeFile(t, filepath.Join(etc, "server.toml"), "")

	var cleaned bool
	origLoad, origCleanup := ebpfEngineLoad, cleanupBPFFiles
	ebpfEngineLoad = func(string, int, string, ebpf.LoadOptions) error { return nil }
	cleanupBPFFiles = func() { cleaned = true }
	t.Cleanup(func() { ebpfEngineLoad, cleanupBPFFiles = origLoad, origCleanup })

	a := &UdpAC{}
	err := a.Start(dir, 0)
	t.Cleanup(a.log.Close)
	if err == nil {
		a.Stop()
		t.Fatal("Start succeeded with an empty server.toml")
	}
	if !cleaned {
		t.Error("failed Start left the XDP program loaded")
	}
	if a.httpServer == nil {
		t.Fatal("HTTPS listener was never started; the test does not cover it")
	}

	// Without a shutdown the listener binds within a few milliseconds.
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if c, dialErr := net.DialTimeout("tcp", addr, 100*time.Millisecond); dialErr == nil {
			c.Close()
			t.Fatalf("failed Start left the HTTPS listener running on %s", addr)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func writeSelfSignedCert(t *testing.T, certPath, keyPath string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, certPath, string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})))
	writeFile(t, keyPath, string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})))
}
