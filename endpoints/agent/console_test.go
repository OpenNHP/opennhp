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
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/common"
)

const testConsoleToken = "test-console-token"

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
	wrongSig := base64.StdEncoding.EncodeToString(ed25519.Sign(otherPriv, digest[:]))

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
		if _, err := os.Stat(filepath.Join(ExeDirPath, "etc", "ta")); !os.IsNotExist(err) {
			t.Errorf("%s: TA directory was created for a rejected upload", tc.name)
		}
	}
}

func TestVerifyTASignature(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	digest := sha256.Sum256([]byte("ta"))
	sig := base64.StdEncoding.EncodeToString(ed25519.Sign(priv, digest[:]))

	if err := verifyTASignature([]ed25519.PublicKey{pub}, digest[:], sig); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	other := sha256.Sum256([]byte("other"))
	if err := verifyTASignature([]ed25519.PublicKey{pub}, other[:], sig); err == nil {
		t.Fatalf("signature accepted for a different file")
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

func signedTAUpload(t *testing.T, priv ed25519.PrivateKey, content []byte) *http.Request {
	t.Helper()
	digest := sha256.Sum256(content)
	return buildTAUpload(t, content, base64.StdEncoding.EncodeToString(ed25519.Sign(priv, digest[:])))
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
