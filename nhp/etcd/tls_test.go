package etcd

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeCA writes a self-signed CA and a leaf keypair signed by it, and
// returns their paths.
func writeCA(t *testing.T) (caPath, certPath, keyPath string) {
	t.Helper()
	dir := t.TempDir()

	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caPath = filepath.Join(dir, "ca.pem")
	writePEM(t, caPath, "CERTIFICATE", caDER)

	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	leafTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "test-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTmpl, caCert, &leafKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	certPath = filepath.Join(dir, "client.pem")
	writePEM(t, certPath, "CERTIFICATE", leafDER)

	keyDER, err := x509.MarshalECPrivateKey(leafKey)
	if err != nil {
		t.Fatal(err)
	}
	keyPath = filepath.Join(dir, "client-key.pem")
	writePEM(t, keyPath, "EC PRIVATE KEY", keyDER)

	return caPath, certPath, keyPath
}

func writePEM(t *testing.T, path, blockType string, der []byte) {
	t.Helper()
	b := pem.EncodeToMemory(&pem.Block{Type: blockType, Bytes: der})
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TLS is optional. A deployment that configures nothing keeps the
// plaintext connection it has always had — this is the property that makes
// the setting safe to add to a released product, and it is easy to lose by
// making the zero value mean "verify".
func TestTLSConfigIsOptional(t *testing.T) {
	conn := &EtcdConn{Endpoints: []string{"127.0.0.1:2379"}}
	conf, err := conn.tlsConfig()
	if err != nil {
		t.Fatalf("an unconfigured connection was refused: %v", err)
	}
	if conf != nil {
		t.Error("an unconfigured connection produced a TLS config")
	}
}

// A half-configured connection is refused rather than quietly downgraded.
// The failure mode of guessing here is a connection that looks encrypted
// and is not, or one that accepts any server.
func TestTLSConfigRefusesHalfConfigured(t *testing.T) {
	ca, cert, key := writeCA(t)

	for name, tc := range map[string]struct {
		conn *EtcdConn
		want string
	}{
		"cert without key": {&EtcdConn{CAFile: ca, CertFile: cert}, "together"},
		"key without cert": {&EtcdConn{CAFile: ca, KeyFile: key}, "together"},
		"client cert without CA": {
			&EtcdConn{CertFile: cert, KeyFile: key}, "CAFile is required",
		},
		"unreadable CA": {&EtcdConn{CAFile: filepath.Join(t.TempDir(), "absent.pem")}, "cannot read"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := tc.conn.tlsConfig()
			if err == nil {
				t.Fatal("accepted a half-configured TLS setup")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to mention %q", err, tc.want)
			}
		})
	}

	// A file that exists but holds no certificate must not produce an
	// empty pool that trusts nothing and says nothing.
	junk := filepath.Join(t.TempDir(), "junk.pem")
	if err := os.WriteFile(junk, []byte("not a certificate\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := (&EtcdConn{CAFile: junk}).tlsConfig(); err == nil {
		t.Error("a CAFile with no certificate was accepted")
	}
}

func TestTLSConfigServerAndClientAuth(t *testing.T) {
	ca, cert, key := writeCA(t)

	// CAFile alone: server authentication and confidentiality, no client
	// certificate.
	conf, err := (&EtcdConn{CAFile: ca}).tlsConfig()
	if err != nil {
		t.Fatalf("CAFile alone was refused: %v", err)
	}
	if conf == nil || conf.RootCAs == nil {
		t.Fatal("no root CA pool was built")
	}
	if len(conf.Certificates) != 0 {
		t.Error("a client certificate was presented without one being configured")
	}
	if conf.InsecureSkipVerify {
		t.Error("certificate verification is disabled")
	}
	if conf.MinVersion < 0x0303 { // TLS 1.2
		t.Errorf("MinVersion = %#x, want TLS 1.2 or later", conf.MinVersion)
	}

	// With a client keypair as well.
	conf, err = (&EtcdConn{CAFile: ca, CertFile: cert, KeyFile: key}).tlsConfig()
	if err != nil {
		t.Fatalf("a full client TLS setup was refused: %v", err)
	}
	if len(conf.Certificates) != 1 {
		t.Fatalf("client certificates = %d, want 1", len(conf.Certificates))
	}
}
