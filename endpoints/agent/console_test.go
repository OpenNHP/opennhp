package agent

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// At least minConsoleTokenLen, like any token the console would accept
// from a file.
const testConsoleToken = "test-console-token-0123456789abcdef"

func newTestConsole(t *testing.T, conf *ConsoleConfig) (*UdpAgent, http.Handler) {
	t.Helper()
	if conf == nil {
		conf = &ConsoleConfig{ListenAddr: defaultConsoleListenAddr}
	}
	a := &UdpAgent{consoleConf: conf}
	return a, a.newConsoleRouter(testConsoleToken)
}

// Every console route — static and dynamic (NoRoute) — must reject callers
// without a valid bearer token before any handler runs.
func TestConsoleRejectsUnauthenticatedRequests(t *testing.T) {
	_, router := newTestConsole(t, nil)

	// Register a dynamic TA route to prove NoRoute dispatch is covered too.
	routes.Lock()
	if routes.m[http.MethodPost] == nil {
		routes.m[http.MethodPost] = make(map[string]gin.HandlerFunc)
	}
	dynPath := taApiPrefix + "/test-ta/fn"
	routes.m[http.MethodPost][dynPath] = func(c *gin.Context) {
		t.Errorf("dynamic TA handler reached without authentication")
		c.Status(http.StatusOK)
	}
	routes.Unlock()
	t.Cleanup(func() {
		routes.Lock()
		delete(routes.m[http.MethodPost], dynPath)
		routes.Unlock()
	})

	cases := []struct{ method, path string }{
		{http.MethodPost, serviceApiPrefix + "/config/server"},
		{http.MethodGet, serviceApiPrefix + "/config/server"},
		{http.MethodGet, serviceApiPrefix + "/key/agent"},
		{http.MethodGet, serviceApiPrefix + "/key/tee"},
		{http.MethodPost, serviceApiPrefix + "/key/agent/rotate"},
		{http.MethodPost, serviceApiPrefix + "/key/tee/rotate"},
		{http.MethodPost, serviceApiPrefix + "/agent/restart"},
		{http.MethodGet, serviceApiPrefix + "/agent/restart"},
		{http.MethodGet, serviceApiPrefix + "/status/agent"},
		{http.MethodGet, serviceApiPrefix + "/attestation/tee"},
		{http.MethodPost, taApiPrefix + "/register"},
		{http.MethodPost, dynPath},
		{http.MethodGet, "/does-not-exist"},
	}

	auths := map[string]string{
		"none":     "",
		"wrong":    "Bearer not-the-token",
		"basic":    "Basic " + base64.StdEncoding.EncodeToString([]byte("a:b")),
		"empty":    "Bearer ",
		"prefixed": "Bearer " + testConsoleToken + "x",
	}

	for _, tc := range cases {
		for name, auth := range auths {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if auth != "" {
				req.Header.Set("Authorization", auth)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s %s with %s auth: got %d, want 401", tc.method, tc.path, name, w.Code)
			}
		}
	}
}

func TestConsoleAcceptsValidToken(t *testing.T) {
	a, router := newTestConsole(t, nil)
	a.safeTee.Store(false)

	req := httptest.NewRequest(http.MethodGet, serviceApiPrefix+"/status/agent", nil)
	req.Header.Set("Authorization", "Bearer "+testConsoleToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("authenticated status request: got %d, want 200 (%s)", w.Code, w.Body.String())
	}
}

// Restart is state-changing and must not be reachable via GET (CSRF).
func TestConsoleRestartRequiresPost(t *testing.T) {
	_, router := newTestConsole(t, nil)

	req := httptest.NewRequest(http.MethodGet, serviceApiPrefix+"/agent/restart", nil)
	req.Header.Set("Authorization", "Bearer "+testConsoleToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatalf("GET /agent/restart succeeded; restart must be POST-only")
	}
}

func buildTAUpload(t *testing.T, content []byte, signature string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for k, v := range map[string]string{
		"name": "demo", "description": "demo", "language": "go", "entry": "ta",
	} {
		_ = mw.WriteField(k, v)
	}
	if signature != "" {
		_ = mw.WriteField("signature", signature)
	}
	fw, err := mw.CreateFormFile("file", "../../evil-ta")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write(content)
	_ = mw.Close()

	req := httptest.NewRequest(http.MethodPost, taApiPrefix+"/register", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+testConsoleToken)
	return req
}

// An authenticated caller still cannot get an unsigned or wrongly signed TA
// written to disk, chmod'ed, or executed.
func TestConsoleRejectsUnsignedTA(t *testing.T) {
	oldExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = oldExeDir })

	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	_, otherPriv, _ := ed25519.GenerateKey(rand.Reader)
	content := []byte("#!/bin/sh\necho pwned\n")
	digest := sha256.Sum256(content)
	wrongSig := base64.StdEncoding.EncodeToString(ed25519.Sign(otherPriv, taSignedMessage(digest[:])))

	cases := []struct {
		name string
		keys []ed25519.PublicKey
		sig  string
	}{
		{"no trusted keys configured", nil, wrongSig},
		{"missing signature", []ed25519.PublicKey{pub}, ""},
		{"garbage signature", []ed25519.PublicKey{pub}, "!!!"},
		{"signature by untrusted key", []ed25519.PublicKey{pub}, wrongSig},
	}

	for _, tc := range cases {
		a, router := newTestConsole(t, &ConsoleConfig{trustedTAKeys: tc.keys})
		a.safeTee.Store(true) // posture alone must not authorize execution

		w := httptest.NewRecorder()
		router.ServeHTTP(w, buildTAUpload(t, content, tc.sig))
		if w.Code != http.StatusForbidden {
			t.Errorf("%s: got %d, want 403 (%s)", tc.name, w.Code, w.Body.String())
		}
		if left := taDirEntries(t, filepath.Join(ExeDirPath, "etc", "ta")); len(left) != 0 {
			t.Errorf("%s: rejected upload left %v in the TA directory", tc.name, left)
		}
	}
}

func TestVerifyTASignature(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	digest := sha256.Sum256([]byte("ta"))
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(priv, taSignedMessage(digest[:])))

	if err := verifyTASignature([]ed25519.PublicKey{pub}, digest[:], sig); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	other := sha256.Sum256([]byte("other"))
	if err := verifyTASignature([]ed25519.PublicKey{pub}, other[:], sig); err == nil {
		t.Fatalf("signature accepted for a different file")
	}

	// A signature over the bare digest (no domain-separation context) —
	// e.g. one made with the same key for another purpose — is rejected.
	bare := base64.StdEncoding.EncodeToString(ed25519.Sign(priv, digest[:]))
	if err := verifyTASignature([]ed25519.PublicKey{pub}, digest[:], bare); err == nil {
		t.Fatalf("signature over the bare digest accepted")
	}
}

func TestLoadConsoleConfigDefaults(t *testing.T) {
	dir := t.TempDir()
	conf, err := loadConsoleConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if conf.ListenAddr != defaultConsoleListenAddr {
		t.Fatalf("default ListenAddr = %q, want %q", conf.ListenAddr, defaultConsoleListenAddr)
	}
	if len(conf.trustedTAKeys) != 0 {
		t.Fatalf("default config trusts TA keys")
	}

	if err := os.WriteFile(filepath.Join(dir, consoleConfigFileName), []byte(`TrustedTAPublicKeys = ["not-a-key"]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConsoleConfig(dir); err == nil {
		t.Fatalf("invalid TA key accepted")
	}
}

func TestLoadOrCreateConsoleToken(t *testing.T) {
	dir := t.TempDir()
	tok1, notices, err := loadOrCreateConsoleToken(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(tok1) != 2*consoleTokenBytes {
		t.Fatalf("token length %d, want %d", len(tok1), 2*consoleTokenBytes)
	}
	if len(notices) != 1 || notices[0].warning || !strings.Contains(notices[0].msg, "generated") {
		t.Fatalf("first start should report token generation, got %+v", notices)
	}
	info, err := os.Stat(filepath.Join(dir, consoleTokenFileName))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm&0o077 != 0 {
		t.Fatalf("token file mode %v is group/other accessible", perm)
	}
	tok2, notices2, err := loadOrCreateConsoleToken(dir)
	if err != nil {
		t.Fatal(err)
	}
	if tok1 != tok2 {
		t.Fatalf("token changed across restarts")
	}
	if len(notices2) != 0 {
		t.Fatalf("unexpected notices on reload: %+v", notices2)
	}
}

// An operator-supplied token must not be trivially guessable.
func TestLoadOrCreateConsoleTokenRejectsWeakToken(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, consoleTokenFileName), []byte("admin\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadOrCreateConsoleToken(dir); err == nil {
		t.Fatal("5-character token accepted")
	}

	strong := strings.Repeat("k", minConsoleTokenLen)
	if err := os.WriteFile(filepath.Join(dir, consoleTokenFileName), []byte(strong+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// WriteFile keeps the existing 0600 mode; widen it explicitly.
	if err := os.Chmod(filepath.Join(dir, consoleTokenFileName), 0o644); err != nil {
		t.Fatal(err)
	}
	tok, notices, err := loadOrCreateConsoleToken(dir)
	if err != nil || tok != strong {
		t.Fatalf("minimum-length token rejected: %v", err)
	}
	if runtime.GOOS != "windows" && (len(notices) != 1 || !notices[0].warning) {
		t.Fatalf("world-readable token file should produce a warning notice, got %+v", notices)
	}
}

// A misspelled console.toml key must fail instead of silently falling back
// to a default (e.g. ListenAdr leaving the console on loopback).
func TestLoadConsoleConfigRejectsUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, consoleConfigFileName), []byte(`ListenAdr = "0.0.0.0:443"`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := loadConsoleConfig(dir)
	if err == nil {
		t.Fatal("unknown key ListenAdr accepted")
	}
	if !strings.Contains(err.Error(), "ListenAdr") {
		t.Fatalf("error does not name the unknown key: %v", err)
	}
}

// configServer must reject a peer table the loader would reject, and must
// not touch the existing trust anchor in that case.
func TestConfigServerRejectsInvalidPeers(t *testing.T) {
	oldExeDir := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = oldExeDir })
	if err := os.MkdirAll(filepath.Join(ExeDirPath, "etc"), 0o755); err != nil {
		t.Fatal(err)
	}
	serverToml := filepath.Join(ExeDirPath, "etc", "server.toml")
	const original = "# original\n"
	if err := os.WriteFile(serverToml, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	_, router := newTestConsole(t, nil)
	for _, body := range []string{`{}`, `{"Servers":[{"Name":"x"}]}`} {
		req := httptest.NewRequest(http.MethodPost, serviceApiPrefix+"/config/server", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+testConsoleToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("body %s: got %d, want 400", body, w.Code)
		}
	}

	got, _ := os.ReadFile(serverToml)
	if string(got) != original {
		t.Fatalf("server.toml modified by a rejected request: %q", got)
	}
}

// writeSelfSignedCert writes etc/certs/server.{crt,key} under dir.
func writeSelfSignedCert(t *testing.T, dir string) {
	t.Helper()
	certDir := filepath.Join(dir, "etc", "certs")
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		t.Fatal(err)
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(certDir, "server.crt"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(certDir, "server.key"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600); err != nil {
		t.Fatal(err)
	}
}

// startConsole runs the same Prepare -> Serve sequence as nhp-agentd.
func startConsole(a *UdpAgent) error {
	if err := a.PrepareDHPWebConsole(); err != nil {
		return err
	}
	return a.ServeDHPWebConsole()
}

func setConsoleExeDir(t *testing.T, dir string) {
	t.Helper()
	oldAgent, oldCommon := ExeDirPath, common.ExeDirPath
	ExeDirPath, common.ExeDirPath = dir, dir
	t.Cleanup(func() { ExeDirPath, common.ExeDirPath = oldAgent, oldCommon })
}

func writeConsoleListenAddr(t *testing.T, dir, addr string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "etc"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "etc", consoleConfigFileName), []byte(`ListenAddr = "`+addr+`"`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

// Console startup must report failures instead of leaving the agent
// running without a console (the old code panicked in a goroutine).
func TestDHPWebConsoleStartupFailsFast(t *testing.T) {
	t.Run("missing certificate", func(t *testing.T) {
		dir := t.TempDir()
		setConsoleExeDir(t, dir)
		writeConsoleListenAddr(t, dir, "127.0.0.1:0")
		if err := startConsole(&UdpAgent{}); err == nil {
			t.Fatal("console started without a TLS certificate")
		}
	})

	t.Run("port in use", func(t *testing.T) {
		busy, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer busy.Close()

		dir := t.TempDir()
		setConsoleExeDir(t, dir)
		writeSelfSignedCert(t, dir)
		writeConsoleListenAddr(t, dir, busy.Addr().String())
		if err := startConsole(&UdpAgent{}); err == nil {
			t.Fatal("console reported success on a port that is already in use")
		}
	})
}

// End to end over real TLS: the console binds loopback and rejects a
// request that carries no token.
func TestDHPWebConsoleServesWithAuth(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := probe.Addr().String()
	probe.Close()

	dir := t.TempDir()
	setConsoleExeDir(t, dir)
	writeSelfSignedCert(t, dir)
	writeConsoleListenAddr(t, dir, addr)

	a := &UdpAgent{}
	if startErr := startConsole(a); startErr != nil {
		t.Fatalf("console start: %v", startErr)
	}
	t.Cleanup(a.CloseDHPWebConsole)

	client := &http.Client{
		Timeout:   5 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec // self-signed test cert
	}
	resp, err := client.Get("https://" + addr + serviceApiPrefix + "/status/agent")
	if err != nil {
		t.Fatalf("GET status: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated GET status: got %d, want 401", resp.StatusCode)
	}
}

// A valid peer table is written in place (so the kqueue file watcher keeps
// working) and keeps the existing file mode.
func TestConfigServerWritesValidPeersInPlace(t *testing.T) {
	dir := t.TempDir()
	setConsoleExeDir(t, dir)
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}
	serverToml := filepath.Join(etc, "server.toml")
	if err := os.WriteFile(serverToml, []byte("# original\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.Stat(serverToml)

	_, router := newTestConsole(t, nil)
	body := `{"Servers":[{"Name":"demo","PubKeyBase64":"4/p0mIknwmVIMocRLQKil7xIthgEdZNncv9UagiBaK2kpcH7i4hEtZjpcHox+Bn7xdV+rBKNbKlV9ye6V1VCLA==","Instances":[{"Ip":"127.0.0.1","Port":62206}]}]}`
	req := httptest.NewRequest(http.MethodPost, serviceApiPrefix+"/config/server", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testConsoleToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("valid config: got %d, want 200 (%s)", w.Code, w.Body.String())
	}

	after, err := os.Stat(serverToml)
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(before, after) {
		t.Fatal("server.toml was replaced (rename) instead of written in place")
	}
	if after.Mode().Perm() != 0o600 {
		t.Fatalf("server.toml mode changed to %v", after.Mode().Perm())
	}
	content, _ := os.ReadFile(serverToml)
	var peers Peers
	if err := toml.Unmarshal(content, &peers); err != nil || len(peers.Servers) != 1 {
		t.Fatalf("written server.toml does not round-trip: %v (%q)", err, content)
	}
}

// Overlapping POST /config/server requests must not interleave their
// in-place writes: the file left behind is exactly one request's body.
func TestConfigServerConcurrentWritesDoNotInterleave(t *testing.T) {
	dir := t.TempDir()
	setConsoleExeDir(t, dir)
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}
	_, router := newTestConsole(t, nil)

	// Truncate, pause, then write, with the short body pausing longer: two
	// unserialized requests then always end as "long write, then short
	// write over it", which leaves the short body followed by the long
	// one's tail.
	old := writeServerConfigFile
	writeServerConfigFile = func(name string, data []byte, perm os.FileMode) error {
		f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return err
		}
		if len(data) < 1024 {
			time.Sleep(20 * time.Millisecond)
		} else {
			time.Sleep(5 * time.Millisecond)
		}
		_, err = f.Write(data)
		if closeErr := f.Close(); err == nil {
			err = closeErr
		}
		return err
	}
	t.Cleanup(func() { writeServerConfigFile = old })

	const pub = "4/p0mIknwmVIMocRLQKil7xIthgEdZNncv9UagiBaK2kpcH7i4hEtZjpcHox+Bn7xdV+rBKNbKlV9ye6V1VCLA=="
	// A long body (many instances) and a short one, so an interleaving
	// would leave the short body followed by the long one's tail.
	var long strings.Builder
	for i := 0; i < 200; i++ {
		if i > 0 {
			long.WriteString(",")
		}
		fmt.Fprintf(&long, `{"Ip":"10.0.%d.%d","Port":62206}`, i/250, i%250+1)
	}
	bodies := []string{
		`{"Servers":[{"Name":"long","PubKeyBase64":"` + pub + `","Instances":[` + long.String() + `]}]}`,
		`{"Servers":[{"Name":"short","PubKeyBase64":"` + pub + `","Instances":[{"Ip":"127.0.0.1","Port":62206}]}]}`,
	}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		for _, body := range bodies {
			wg.Add(1)
			go func(body string) {
				defer wg.Done()
				req := httptest.NewRequest(http.MethodPost, serviceApiPrefix+"/config/server", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+testConsoleToken)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				if w.Code != http.StatusOK {
					t.Errorf("config write: got %d (%s)", w.Code, w.Body.String())
				}
			}(body)
		}
	}
	wg.Wait()

	content, err := os.ReadFile(filepath.Join(etc, "server.toml"))
	if err != nil {
		t.Fatal(err)
	}
	var peers Peers
	if err := toml.Unmarshal(content, &peers); err != nil || len(peers.Servers) != 1 {
		t.Fatalf("server.toml is not a single request's body: %v (%q)", err, content)
	}
	want := map[string]int{"long": 200, "short": 1}[peers.Servers[0].Name]
	if got := len(peers.Servers[0].Instances); got != want {
		t.Fatalf("server.toml mixes two writes: %q has %d instances, want %d", peers.Servers[0].Name, got, want)
	}
}

// Concurrent uploads of the same signed TA must store a single copy: the
// dedup lookup and the index write happen as one step.
func TestRegisterTAConcurrentUploadsStoreOneCopy(t *testing.T) {
	dir := t.TempDir()
	setConsoleExeDir(t, dir)
	taDir := filepath.Join(dir, "etc", "ta")

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	a, router := newTestConsole(t, &ConsoleConfig{trustedTAKeys: []ed25519.PublicKey{pub}})
	a.safeTee.Store(true)

	// Not a valid executable: starting it fails fast after it is stored
	// and indexed, so the test never runs anything real.
	content := []byte("not an executable, concurrent\n")
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		req := signedTAUpload(t, priv, content)
		wg.Add(1)
		go func() {
			defer wg.Done()
			router.ServeHTTP(httptest.NewRecorder(), req)
		}()
	}
	wg.Wait()

	entries, err := os.ReadDir(taDir)
	if err != nil {
		t.Fatal(err)
	}
	var copies []string
	for _, e := range entries {
		if e.IsDir() {
			copies = append(copies, e.Name())
		}
	}
	if len(copies) != 1 {
		t.Fatalf("concurrent uploads stored %d copies (%v), want 1", len(copies), copies)
	}
	digest := sha256.Sum256(content)
	if got := readTAIndexUUID(t, filepath.Join(taDir, "sha256-"+hex.EncodeToString(digest[:]))); got != copies[0] {
		t.Fatalf("index points at %q, stored copy is %q", got, copies[0])
	}
}

func signedTAUpload(t *testing.T, priv ed25519.PrivateKey, content []byte) *http.Request {
	t.Helper()
	digest := sha256.Sum256(content)
	return buildTAUpload(t, content, base64.StdEncoding.EncodeToString(ed25519.Sign(priv, taSignedMessage(digest[:]))))
}

// An oversized upload is reported as 413, not as missing parameters.
func TestRegisterTARejectsOversizedUpload(t *testing.T) {
	setConsoleExeDir(t, t.TempDir())
	old := maxTAUploadSize
	maxTAUploadSize = 1 << 10
	t.Cleanup(func() { maxTAUploadSize = old })

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	a, router := newTestConsole(t, &ConsoleConfig{trustedTAKeys: []ed25519.PublicKey{pub}})
	a.safeTee.Store(true)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, signedTAUpload(t, priv, bytes.Repeat([]byte("x"), 4<<10)))
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized upload: got %d, want 413 (%s)", w.Code, w.Body.String())
	}
}

// Deduplication must key on the verified SHA-256. A legacy MD5-named
// index entry (e.g. an unsigned TA registered before signatures were
// required) must never be reused for a signed upload.
func TestRegisterTAIgnoresLegacyMD5Index(t *testing.T) {
	dir := t.TempDir()
	setConsoleExeDir(t, dir)
	taDir := filepath.Join(dir, "etc", "ta")
	if err := os.MkdirAll(filepath.Join(taDir, "legacy-uuid"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Not a valid executable: starting it fails fast, so the test never
	// runs anything real.
	content := []byte("not an executable\n")
	md5sum := fmt.Sprintf("%x", md5.Sum(content)) //nolint:gosec // reproduces the legacy index name
	legacy := `{"uuid":"legacy-uuid","fileName":"evil-ta"}`
	if err := os.WriteFile(filepath.Join(taDir, md5sum), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	a, router := newTestConsole(t, &ConsoleConfig{trustedTAKeys: []ed25519.PublicKey{pub}})
	a.safeTee.Store(true)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, signedTAUpload(t, priv, content))

	digest := sha256.Sum256(content)
	idx, err := os.ReadFile(filepath.Join(taDir, "sha256-"+hex.EncodeToString(digest[:])))
	if err != nil {
		t.Fatalf("sha256 index not written (status %d %s): %v", w.Code, w.Body.String(), err)
	}
	var info map[string]any
	if err := json.Unmarshal(idx, &info); err != nil {
		t.Fatal(err)
	}
	if info["uuid"] == "legacy-uuid" {
		t.Fatal("signed upload was resolved to the legacy MD5-indexed TA")
	}
	if got, _ := os.ReadFile(filepath.Join(taDir, md5sum)); string(got) != legacy {
		t.Fatal("legacy MD5 index entry was modified")
	}
}

func writeTAIndex(t *testing.T, taDir string, content []byte, uuid, fileName string) string {
	t.Helper()
	digest := sha256.Sum256(content)
	idx := filepath.Join(taDir, "sha256-"+hex.EncodeToString(digest[:]))
	body, _ := json.Marshal(map[string]any{"uuid": uuid, "fileName": fileName})
	if err := os.WriteFile(idx, body, 0o644); err != nil {
		t.Fatal(err)
	}
	return idx
}

func readTAIndexUUID(t *testing.T, idx string) string {
	t.Helper()
	b, err := os.ReadFile(idx)
	if err != nil {
		t.Fatal(err)
	}
	var info map[string]any
	if err := json.Unmarshal(b, &info); err != nil {
		t.Fatal(err)
	}
	u, _ := info["uuid"].(string)
	return u
}

// Re-uploading an already registered (signed) TA reuses the stored binary
// under its previous uuid and file name, without writing a new copy.
func TestRegisterTADedupeReusesStoredBinary(t *testing.T) {
	dir := t.TempDir()
	setConsoleExeDir(t, dir)
	taDir := filepath.Join(dir, "etc", "ta")
	content := []byte("already registered\n")
	if err := os.MkdirAll(filepath.Join(taDir, "prev-uuid"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(taDir, "prev-uuid", "prev-ta"), content, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTAIndex(t, taDir, content, "prev-uuid", "prev-ta")

	bufferedTaLock.Lock()
	bufferedTaMap["prev-uuid"] = &TrustedApplication{Id: "prev-uuid"}
	bufferedTaLock.Unlock()
	t.Cleanup(func() {
		bufferedTaLock.Lock()
		delete(bufferedTaMap, "prev-uuid")
		bufferedTaLock.Unlock()
	})

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	a, router := newTestConsole(t, &ConsoleConfig{trustedTAKeys: []ed25519.PublicKey{pub}})
	a.safeTee.Store(true)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, signedTAUpload(t, priv, content))
	if w.Code != http.StatusOK {
		t.Fatalf("re-upload of registered TA: got %d, want 200 (%s)", w.Code, w.Body.String())
	}

	entries, _ := os.ReadDir(taDir)
	for _, e := range entries {
		if e.IsDir() && e.Name() != "prev-uuid" {
			t.Errorf("duplicate upload wrote a new copy under %s", e.Name())
		}
	}
}

// A stale index (stored binary removed) or a stored binary that no longer
// matches the verified digest must not be reused; the upload is handled as
// new and the index is repaired to point at the fresh copy.
func TestRegisterTADedupeIgnoresStaleOrTamperedEntries(t *testing.T) {
	cases := map[string]func(t *testing.T, taDir string){
		"missing binary": func(t *testing.T, taDir string) {},
		"tampered binary": func(t *testing.T, taDir string) {
			if err := os.MkdirAll(filepath.Join(taDir, "prev-uuid"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(taDir, "prev-uuid", "prev-ta"), []byte("swapped\n"), 0o755); err != nil {
				t.Fatal(err)
			}
		},
	}

	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			setConsoleExeDir(t, dir)
			taDir := filepath.Join(dir, "etc", "ta")
			if err := os.MkdirAll(taDir, 0o755); err != nil {
				t.Fatal(err)
			}
			// Not a valid executable: starting it fails fast.
			content := []byte("signed but not executable\n")
			idx := writeTAIndex(t, taDir, content, "prev-uuid", "prev-ta")
			setup(t, taDir)

			pub, priv, _ := ed25519.GenerateKey(rand.Reader)
			a, router := newTestConsole(t, &ConsoleConfig{trustedTAKeys: []ed25519.PublicKey{pub}})
			a.safeTee.Store(true)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, signedTAUpload(t, priv, content))

			newUUID := readTAIndexUUID(t, idx)
			if newUUID == "prev-uuid" {
				t.Fatalf("stale/tampered entry was reused (status %d %s)", w.Code, w.Body.String())
			}
			got, err := os.ReadFile(filepath.Join(taDir, newUUID, "evil-ta"))
			if err != nil || !bytes.Equal(got, content) {
				t.Fatalf("fresh copy not stored under the repaired index: %v", err)
			}
		})
	}
}

// PrepareDHPWebConsole binds the listener without serving; startup failure
// paths can release it with CloseDHPWebConsole.
func TestPrepareAndCloseDHPWebConsole(t *testing.T) {
	dir := t.TempDir()
	setConsoleExeDir(t, dir)
	writeSelfSignedCert(t, dir)
	writeConsoleListenAddr(t, dir, "127.0.0.1:0")

	a := &UdpAgent{}
	if err := a.PrepareDHPWebConsole(); err != nil {
		t.Fatalf("PrepareDHPWebConsole: %v", err)
	}
	addr := a.consoleListener.Addr().String()
	a.CloseDHPWebConsole()

	// The port must be free again after Close.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listener not released by CloseDHPWebConsole: %v", err)
	}
	ln.Close()

	if err := (&UdpAgent{}).ServeDHPWebConsole(); err == nil {
		t.Fatal("ServeDHPWebConsole succeeded without PrepareDHPWebConsole")
	}
}

// The console.toml files shipped in the repo must load under strict
// decoding (unknown keys are rejected).
func TestShippedConsoleConfigsLoad(t *testing.T) {
	for _, dir := range []string{
		filepath.Join("main", "etc"),
		filepath.Join("..", "..", "docker", "nhp-agent", "etc"),
	} {
		conf, err := loadConsoleConfig(dir)
		if err != nil {
			t.Errorf("%s: %v", dir, err)
			continue
		}
		if _, statErr := os.Stat(filepath.Join(dir, consoleConfigFileName)); statErr != nil {
			t.Errorf("%s: shipped console.toml missing: %v", dir, statErr)
		}
		if conf.ListenAddr == "" {
			t.Errorf("%s: empty ListenAddr", dir)
		}
	}
}

// gin's trailing-slash / fixed-path redirects run before middleware; they
// must be off so an unauthenticated caller gets 401, not a redirect that
// confirms the route exists.
func TestConsoleNoRedirectBeforeAuth(t *testing.T) {
	_, router := newTestConsole(t, nil)
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, serviceApiPrefix + "/agent/restart/"},
		{http.MethodGet, serviceApiPrefix + "/status/agent/"},
		{http.MethodGet, "/API/V1/status/agent"},
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: got %d (Location %q), want 401", tc.method, tc.path, w.Code, w.Header().Get("Location"))
		}
	}
}

// With no signing keys configured (the default), registration is refused
// before the body is parsed, so an oversized upload gets 403, not 413.
func TestRegisterTARefusedBeforeParsingWithoutKeys(t *testing.T) {
	setConsoleExeDir(t, t.TempDir())
	old := maxTAUploadSize
	maxTAUploadSize = 1 << 10
	t.Cleanup(func() { maxTAUploadSize = old })

	a, router := newTestConsole(t, &ConsoleConfig{})
	a.safeTee.Store(true)

	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, signedTAUpload(t, priv, bytes.Repeat([]byte("x"), 4<<10)))
	if w.Code != http.StatusForbidden {
		t.Fatalf("got %d, want 403 before parsing (%s)", w.Code, w.Body.String())
	}
}

func TestIsPlainPathElement(t *testing.T) {
	for name, want := range map[string]bool{
		"ta": true, "prev-uuid": true, "a.b": true,
		"": false, ".": false, "..": false, "a/b": false, "../x": false, "/": false,
		`..\\evil`: false, `a\\b`: false, "a\x00b": false, "a\nb": false, "a\x7fb": false, "\x01": false,
		"ta-v1.2_x86": true,
	} {
		if got := isPlainPathElement(name); got != want {
			t.Errorf("isPlainPathElement(%q) = %v, want %v", name, got, want)
		}
	}
}

// An index entry whose uuid is ".." must not resolve outside etc/ta, even
// if the file it would point at matches the verified digest.
func TestLookupRegisteredTARejectsDotDot(t *testing.T) {
	dir := t.TempDir()
	taDir := filepath.Join(dir, "etc", "ta")
	if err := os.MkdirAll(taDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte("outside\n")
	if err := os.WriteFile(filepath.Join(dir, "etc", "x"), content, 0o755); err != nil {
		t.Fatal(err)
	}
	idx := writeTAIndex(t, taDir, content, "..", "x")
	digest := sha256.Sum256(content)
	if _, _, ok := lookupRegisteredTA(taDir, idx, digest[:]); ok {
		t.Fatal(`index entry with uuid ".." was reused`)
	}
}

// A validly signed TA whose digest is listed in RevokedTADigests is
// refused, and the list is validated at load time.
func TestRevokedTADigests(t *testing.T) {
	dir := t.TempDir()
	setConsoleExeDir(t, dir)

	content := []byte("old vulnerable ta\n")
	digest := sha256.Sum256(content)

	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	cfg := fmt.Sprintf("TrustedTAPublicKeys = [%q]\nRevokedTADigests = [%q]\n",
		base64.StdEncoding.EncodeToString(pub), strings.ToUpper(hex.EncodeToString(digest[:])))
	if err := os.WriteFile(filepath.Join(etc, consoleConfigFileName), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	conf, err := loadConsoleConfig(etc)
	if err != nil {
		t.Fatal(err)
	}

	a, router := newTestConsole(t, conf)
	a.safeTee.Store(true)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, signedTAUpload(t, priv, content))
	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "revoked") {
		t.Fatalf("revoked TA: got %d %s, want 403 revoked", w.Code, w.Body.String())
	}
	if left := taDirEntries(t, filepath.Join(etc, "ta")); len(left) != 0 {
		t.Fatalf("revoked TA left %v in the TA directory", left)
	}

	if err := os.WriteFile(filepath.Join(etc, consoleConfigFileName), []byte(`RevokedTADigests = ["abc"]`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadConsoleConfig(etc); err == nil {
		t.Fatal("malformed RevokedTADigests entry accepted")
	}
}

// Once shutdown has begun, console restart / key rotation are refused, and
// an in-flight one holds off the shutdown until it finishes.
func TestLifecycleOpsSerializedWithShutdown(t *testing.T) {
	a, router := newTestConsole(t, nil)

	// An in-flight lifecycle op blocks shutdown from proceeding.
	if !a.beginLifecycleOp(nil) {
		t.Fatal("beginLifecycleOp refused before shutdown")
	}
	stopped := make(chan struct{})
	go func() {
		// Mirrors StopDHP's locking without stopping a real agent.
		a.consoleLifecycleMu.Lock()
		a.consoleStopping = true
		a.consoleLifecycleMu.Unlock()
		close(stopped)
	}()
	select {
	case <-stopped:
		t.Fatal("shutdown proceeded while a lifecycle op was in flight")
	case <-time.After(100 * time.Millisecond):
	}
	a.endLifecycleOp()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown did not proceed after the lifecycle op finished")
	}

	// New ones are refused with 503.
	for _, path := range []string{"/agent/restart", "/key/agent/rotate", "/key/tee/rotate"} {
		req := httptest.NewRequest(http.MethodPost, serviceApiPrefix+path, nil)
		req.Header.Set("Authorization", "Bearer "+testConsoleToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("POST %s during shutdown: got %d, want 503", path, w.Code)
		}
	}
}

// Stop must be a no-op when the agent is not running: after a failed
// Start (e.g. a console restart whose Start rejected the config) the
// channels Stop closes are already closed, and closing them again panics.
func TestStopIsNoOpWhenNotRunning(t *testing.T) {
	a := &UdpAgent{}
	a.Stop() // never started: channels and device are nil
	a.Stop()

	stopped := &UdpAgent{}
	stopped.signals.stop = make(chan struct{})
	close(stopped.signals.stop) // as left behind by a previous Stop
	stopped.Stop()              // must not close it again
}

// An etc/ta left at 0755 by an older version is tightened to 0700 on the
// next registration (MkdirAll does not change an existing directory).
func TestRegisterTATightensExistingTADir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions")
	}
	dir := t.TempDir()
	setConsoleExeDir(t, dir)
	taDir := filepath.Join(dir, "etc", "ta")
	if err := os.MkdirAll(taDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(taDir, 0o755); err != nil {
		t.Fatal(err)
	}

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	a, router := newTestConsole(t, &ConsoleConfig{trustedTAKeys: []ed25519.PublicKey{pub}})
	a.safeTee.Store(true)
	w := httptest.NewRecorder()
	// Not a valid executable: starting it fails fast after the directory
	// handling under test.
	router.ServeHTTP(w, signedTAUpload(t, priv, []byte("not executable\n")))

	info, err := os.Stat(taDir)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		t.Fatalf("etc/ta mode %v after registration, want 0700 (status %d)", perm, w.Code)
	}
}

// A Start that fails after loadBaseConfig installed its watcher (here: an
// unparsable private key) must release the watcher itself, because the
// agent is not running and a later Stop() is a no-op.
func TestFailedStartReleasesConfigWatchers(t *testing.T) {
	dir := t.TempDir()
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(etc, "config.toml"), []byte("PrivateKeyBase64 = \"not-a-valid-key\"\nLogLevel = 5\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	oldAgent, oldCommon := ExeDirPath, common.ExeDirPath
	t.Cleanup(func() {
		ExeDirPath, common.ExeDirPath = oldAgent, oldCommon
		baseConfigWatch, dhpConfigWatch, serverConfigWatch, resourceConfigWatch = nil, nil, nil, nil
	})

	a := &UdpAgent{}
	if err := a.Start(dir, 5); err == nil {
		t.Fatal("Start succeeded with an unparsable private key")
	}

	want := "file watcher for " + filepath.Join(etc, "config.toml") + " closed"
	deadline := time.Now().Add(3 * time.Second)
	for {
		matches, _ := filepath.Glob(filepath.Join(dir, "logs", "*.log"))
		for _, m := range matches {
			if b, _ := os.ReadFile(m); bytes.Contains(b, []byte(want)) {
				a.log.Close()
				return
			}
		}
		if time.Now().After(deadline) {
			a.log.Close()
			t.Fatalf("config.toml watcher was not closed after a failed Start (looked for %q)", want)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// A failed console restart leaves the agent stopped; the 500 must say so.
func TestConsoleFailedRestartReportsStopped(t *testing.T) {
	dir := t.TempDir()
	etc := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etc, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(etc, "config.toml"), []byte("PrivateKeyBase64 = \"not-a-valid-key\"\nLogLevel = 5\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	oldAgent, oldCommon := ExeDirPath, common.ExeDirPath
	ExeDirPath, common.ExeDirPath = dir, dir
	a, router := newTestConsole(t, nil)
	t.Cleanup(func() {
		if a.log != nil {
			a.log.Close()
		}
		ExeDirPath, common.ExeDirPath = oldAgent, oldCommon
		baseConfigWatch, dhpConfigWatch, serverConfigWatch, resourceConfigWatch = nil, nil, nil, nil
	})

	req := httptest.NewRequest(http.MethodPost, serviceApiPrefix+"/agent/restart", nil)
	req.Header.Set("Authorization", "Bearer "+testConsoleToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("restart with a bad config: got %d, want 500 (%s)", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["running"] != false || !strings.Contains(fmt.Sprint(resp["msg"]), "stopped") {
		t.Fatalf("failed restart response does not report the agent as stopped: %s", w.Body.String())
	}
	if a.IsRunning() {
		t.Fatal("agent reports running after a failed restart")
	}

	// a.config is nil now. Handlers that need the loaded config or the
	// device answer 503 instead of panicking; dynamic TA routes too.
	taPath := taApiPrefix + "/test-stopped-fn"
	routes.Lock()
	if routes.m[http.MethodPost] == nil {
		routes.m[http.MethodPost] = make(map[string]gin.HandlerFunc)
	}
	routes.m[http.MethodPost][taPath] = func(c *gin.Context) { t.Error("TA handler ran while the agent is stopped") }
	routes.Unlock()
	t.Cleanup(func() {
		routes.Lock()
		delete(routes.m[http.MethodPost], taPath)
		routes.Unlock()
	})
	a.safeTee.Store(true)
	for _, r := range []struct{ method, path string }{
		{http.MethodGet, serviceApiPrefix + "/key/agent"},
		{http.MethodGet, serviceApiPrefix + "/key/tee"},
		{http.MethodPost, taPath},
	} {
		req := httptest.NewRequest(r.method, r.path, nil)
		req.Header.Set("Authorization", "Bearer "+testConsoleToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("%s %s with the agent stopped: got %d, want 503 (%s)", r.method, r.path, w.Code, w.Body.String())
		}
	}

	// The routes needed to recover stay reachable.
	for _, path := range []string{"/status/agent", "/config/server"} {
		req := httptest.NewRequest(http.MethodGet, serviceApiPrefix+path, nil)
		req.Header.Set("Authorization", "Bearer "+testConsoleToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code == http.StatusServiceUnavailable {
			t.Errorf("GET %s is gated while the agent is stopped; it must stay usable for recovery", path)
		}
	}
}

// A key read must not interleave with a restart: it waits for the
// lifecycle lock and then sees the restart's outcome. Here the "restart"
// fails, so the read gets 503 rather than reading a.config mid-teardown.
func TestKeyReadWaitsForRestart(t *testing.T) {
	a, router := newTestConsole(t, nil)
	a.config = &Config{}
	a.running.Store(true)

	a.consoleLifecycleMu.Lock() // a restart is in progress
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodGet, serviceApiPrefix+"/key/agent", nil)
		req.Header.Set("Authorization", "Bearer "+testConsoleToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		done <- w
	}()

	select {
	case w := <-done:
		a.consoleLifecycleMu.Unlock()
		t.Fatalf("key read completed during a restart: %d %s", w.Code, w.Body.String())
	case <-time.After(100 * time.Millisecond):
	}

	// The restart stops the agent, clears the config and fails to start.
	a.running.Store(false)
	a.config = nil
	a.consoleLifecycleMu.Unlock()

	select {
	case w := <-done:
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("key read after a failed restart: got %d, want 503 (%s)", w.Code, w.Body.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("key read did not complete after the restart finished")
	}
}

// taDirEntries lists what is in the TA directory; a missing directory is
// empty. A rejected upload may leave the (owner-only) directory itself
// behind, since uploads are staged in it, but nothing inside it.
func taDirEntries(t *testing.T, taDir string) []string {
	t.Helper()
	entries, err := os.ReadDir(taDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// The staged file holds exactly the uploaded bytes, its digest is theirs,
// and it is owner-only and not executable until it is renamed into place.
func TestStageUploadedTA(t *testing.T) {
	taDir := t.TempDir()
	content := bytes.Repeat([]byte("staged ta "), 1000)
	req := buildTAUpload(t, content, "sig")
	if err := req.ParseMultipartForm(1 << 10); err != nil { // force a spill to disk
		t.Fatal(err)
	}
	_, fh, err := req.FormFile("file")
	if err != nil {
		t.Fatal(err)
	}

	staged, size, digest, err := stageUploadedTA(fh, taDir)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(staged) != taDir {
		t.Fatalf("staged outside the TA directory: %s", staged)
	}
	got, err := os.ReadFile(staged)
	if err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256(content)
	if !bytes.Equal(got, content) || !bytes.Equal(digest, want[:]) || size != int64(len(content)) {
		t.Fatal("staged bytes or digest do not match the upload")
	}
	if runtime.GOOS != "windows" {
		st, _ := os.Stat(staged)
		if st.Mode().Perm() != 0o600 {
			t.Fatalf("staged file mode %v, want 0600", st.Mode().Perm())
		}
	}
}

// A store that fails after the new <uuid>/ directory is created removes it
// and the staged upload, leaving nothing unindexed in the TA directory.
func TestRegisterTAFailedStoreCleansUp(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on POSIX directory permissions")
	}
	dir := t.TempDir()
	setConsoleExeDir(t, dir)
	taDir := filepath.Join(dir, "etc", "ta")
	content := []byte("store fails\n")
	digest := sha256.Sum256(content)
	// Make the index write fail: a directory where the index file goes.
	if err := os.MkdirAll(filepath.Join(taDir, "sha256-"+hex.EncodeToString(digest[:]), "blocker"), 0o700); err != nil {
		t.Fatal(err)
	}

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	a, router := newTestConsole(t, &ConsoleConfig{trustedTAKeys: []ed25519.PublicKey{pub}})
	a.safeTee.Store(true)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, signedTAUpload(t, priv, content))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("failed store: got %d, want 500 (%s)", w.Code, w.Body.String())
	}
	left := taDirEntries(t, taDir)
	if len(left) != 1 || left[0] != "sha256-"+hex.EncodeToString(digest[:]) {
		t.Fatalf("failed store left %v in the TA directory, want only the blocker", left)
	}
}

// Every console route except TA registration caps its body at
// maxConsoleBodySize and answers 413; TA registration keeps its own,
// larger cap.
func TestConsoleBodySizeLimit(t *testing.T) {
	dir := t.TempDir()
	setConsoleExeDir(t, dir)
	if err := os.MkdirAll(filepath.Join(dir, "etc"), 0o755); err != nil {
		t.Fatal(err)
	}
	old := maxConsoleBodySize
	maxConsoleBodySize = 1 << 10
	t.Cleanup(func() { maxConsoleBodySize = old })

	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	a, router := newTestConsole(t, &ConsoleConfig{trustedTAKeys: []ed25519.PublicKey{pub}})
	a.safeTee.Store(true)

	big := `{"Servers":[{"Name":"` + strings.Repeat("x", 4<<10) + `"}]}`
	req := httptest.NewRequest(http.MethodPost, serviceApiPrefix+"/config/server", strings.NewReader(big))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testConsoleToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized config/server body: got %d, want 413 (%s)", w.Code, w.Body.String())
	}

	// A TA larger than the console-wide cap is still accepted for
	// verification (it fails later only because it is not executable).
	w = httptest.NewRecorder()
	router.ServeHTTP(w, signedTAUpload(t, priv, bytes.Repeat([]byte("t"), 8<<10)))
	if w.Code == http.StatusRequestEntityTooLarge || w.Code == http.StatusForbidden || w.Code == http.StatusBadRequest {
		t.Fatalf("TA upload above the console-wide cap: got %d (%s)", w.Code, w.Body.String())
	}
}

// If the console server stops on its own after it started serving,
// ConsoleFailed reports it; a deliberate CloseDHPWebConsole does not.
func TestConsoleFailedReportsUnexpectedStop(t *testing.T) {
	for _, tc := range []struct {
		name       string
		stop       func(a *UdpAgent)
		wantFailed bool
	}{
		{"listener lost", func(a *UdpAgent) { a.consoleListener.Close() }, true},
		{"closed on shutdown", func(a *UdpAgent) { a.CloseDHPWebConsole() }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			setConsoleExeDir(t, dir)
			writeSelfSignedCert(t, dir)
			writeConsoleListenAddr(t, dir, "127.0.0.1:0")

			a := &UdpAgent{}
			if a.ConsoleFailed() != nil {
				t.Fatal("ConsoleFailed is non-nil before the console serves")
			}
			if err := startConsole(a); err != nil {
				t.Fatalf("console start: %v", err)
			}
			t.Cleanup(a.CloseDHPWebConsole)

			tc.stop(a)
			select {
			case err := <-a.ConsoleFailed():
				if !tc.wantFailed {
					t.Fatalf("deliberate shutdown reported as a failure: %v", err)
				}
			case <-time.After(500 * time.Millisecond):
				if tc.wantFailed {
					t.Fatal("console stopping on its own was not reported")
				}
			}
		})
	}
}
