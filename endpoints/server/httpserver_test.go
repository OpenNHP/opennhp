package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// LoadFilesRecursively walks a directory and parses every .html/.tmpl file
// into the gin engine. A missing directory, a well-formed template, and a
// malformed one that fails to parse must all be handled without panicking.
func TestLoadFilesRecursively(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// missing directory: returns early, no error
	LoadFilesRecursively(gin.New(), filepath.Join(t.TempDir(), "nonexistent"))

	// well-formed template: parsed and registered
	okDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(okDir, "ok.html"), []byte("<p>{{.X}}</p>"), 0o600); err != nil {
		t.Fatal(err)
	}
	LoadFilesRecursively(gin.New(), okDir)

	// malformed template: the parse error is logged, not fatal
	badDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(badDir, "bad.html"), []byte("{{ .Unclosed "), 0o600); err != nil {
		t.Fatal(err)
	}
	LoadFilesRecursively(gin.New(), badDir)
}

// newTestHttpServer returns an HttpServer whose gin engine is wired but whose
// listener never starts, so the proxy-trust plumbing can be exercised without
// binding a port.
func newTestHttpServer(t *testing.T, trustedProxies []string) *HttpServer {
	t.Helper()
	gin.SetMode(gin.TestMode)
	hs := &HttpServer{ginEngine: gin.New()}
	if err := hs.applyTrustedProxies(trustedProxies); err != nil {
		t.Fatalf("applyTrustedProxies(%v): %v", trustedProxies, err)
	}
	return hs
}

// newGinContext builds a gin context for a request arriving from remoteAddr
// carrying the given headers, the way the plugin routes see one. It is bound
// to hs.ginEngine on purpose: gin resolves the trusted-proxy set through the
// engine the context belongs to, so a context from a throwaway engine would
// answer with gin's trust-everyone default no matter what applyTrustedProxies
// installed.
func newGinContext(t *testing.T, hs *HttpServer, remoteAddr string, headers map[string]string) *gin.Context {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/plugins/example?action=login&resid=demo", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), hs.ginEngine)
	ctx.Request = req
	return ctx
}

// A default HttpConfig names no proxy, and in that state the knock source must
// be the TCP peer no matter what the caller puts in the forwarding headers.
// This is the CWE-348 case: before the fix, gin.New()'s all-proxies-trusted
// default let X-Forwarded-For choose the address the AC opens a rule for.
func TestKnockSourceIPIgnoresForwardingHeadersWithoutTrustedProxies(t *testing.T) {
	hs := newTestHttpServer(t, nil)

	for _, tc := range []struct {
		name    string
		headers map[string]string
	}{
		{"no headers", nil},
		{"x-forwarded-for", map[string]string{"X-Forwarded-For": "203.0.113.7"}},
		{"x-real-ip", map[string]string{"X-Real-IP": "203.0.113.7"}},
		{"both", map[string]string{"X-Forwarded-For": "203.0.113.7", "X-Real-IP": "198.51.100.9"}},
		{"chained", map[string]string{"X-Forwarded-For": "203.0.113.7, 198.51.100.9, 192.0.2.5"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newGinContext(t, hs, "192.0.2.10:51000", tc.headers)
			if got := hs.knockSourceIP(ctx); got != "192.0.2.10" {
				t.Fatalf("knockSourceIP = %q, want the TCP peer 192.0.2.10 (headers %v)", got, tc.headers)
			}
		})
	}
}

// An empty or whitespace-only list is the same as no list: gin must be left
// with nil trustedCIDRs so ClientIP() cannot be steered either.
func TestApplyTrustedProxiesTreatsBlankEntriesAsNone(t *testing.T) {
	for _, proxies := range [][]string{{}, {""}, {"  ", "\t"}} {
		hs := newTestHttpServer(t, proxies)
		if len(hs.trustedProxies) != 0 {
			t.Fatalf("trustedProxies = %v for input %v, want empty", hs.trustedProxies, proxies)
		}
		ctx := newGinContext(t, hs, "127.0.0.1:51000", map[string]string{"X-Forwarded-For": "203.0.113.7"})
		if got := hs.knockSourceIP(ctx); got != "127.0.0.1" {
			t.Fatalf("knockSourceIP = %q for input %v, want 127.0.0.1", got, proxies)
		}
	}
}

// With a proxy configured, the header is honored for requests that actually
// come from it — and only from it. A direct caller that binds the listener
// address itself still gets attributed to its own TCP peer.
func TestKnockSourceIPHonorsConfiguredProxyOnly(t *testing.T) {
	hs := newTestHttpServer(t, []string{"127.0.0.1", "::1"})

	fromProxy := newGinContext(t, hs, "127.0.0.1:51000", map[string]string{"X-Forwarded-For": "203.0.113.7"})
	if got := hs.knockSourceIP(fromProxy); got != "203.0.113.7" {
		t.Fatalf("knockSourceIP via the trusted proxy = %q, want 203.0.113.7", got)
	}

	direct := newGinContext(t, hs, "192.0.2.10:51000", map[string]string{"X-Forwarded-For": "203.0.113.7"})
	if got := hs.knockSourceIP(direct); got != "192.0.2.10" {
		t.Fatalf("knockSourceIP from an untrusted peer = %q, want the TCP peer 192.0.2.10", got)
	}
}

// A CIDR entry covers the prefix it names and nothing outside it.
func TestApplyTrustedProxiesAcceptsCIDR(t *testing.T) {
	hs := newTestHttpServer(t, []string{"10.1.0.0/16"})

	inside := newGinContext(t, hs, "10.1.2.3:51000", map[string]string{"X-Forwarded-For": "203.0.113.7"})
	if got := hs.knockSourceIP(inside); got != "203.0.113.7" {
		t.Fatalf("knockSourceIP from inside the trusted prefix = %q, want 203.0.113.7", got)
	}

	outside := newGinContext(t, hs, "10.2.2.3:51000", map[string]string{"X-Forwarded-For": "203.0.113.7"})
	if got := hs.knockSourceIP(outside); got != "10.2.2.3" {
		t.Fatalf("knockSourceIP from outside the trusted prefix = %q, want 10.2.2.3", got)
	}
}

// One unparsable entry fails the whole list rather than being skipped, and the
// engine is left trusting nobody rather than trusting the prefix gin managed
// to parse before it hit the bad entry. A partial trust is the dangerous
// outcome: it is a trust list the operator never wrote.
func TestApplyTrustedProxiesRejectsBadEntryWithoutPartialTrust(t *testing.T) {
	for _, proxies := range [][]string{
		{"not-an-ip"},
		{"relay.opennhp.org"},
		{"10.0.1.300"},
		{"127.0.0.1", "10.0.1.300"},
		{"127.0.0.1", "10.1.0.0/64"},
	} {
		gin.SetMode(gin.TestMode)
		hs := &HttpServer{ginEngine: gin.New()}
		if err := hs.applyTrustedProxies(proxies); err == nil {
			t.Fatalf("applyTrustedProxies(%v) = nil, want an error", proxies)
		}
		if len(hs.trustedProxies) != 0 {
			t.Fatalf("trustedProxies = %v after a refused list %v, want empty", hs.trustedProxies, proxies)
		}
		// The valid prefix of the list must not survive on the engine either.
		ctx := newGinContext(t, hs, "127.0.0.1:51000", map[string]string{"X-Forwarded-For": "203.0.113.7"})
		if got := ctx.ClientIP(); got != "127.0.0.1" {
			t.Fatalf("ClientIP = %q after a refused list %v, want the TCP peer 127.0.0.1", got, proxies)
		}
	}
}

// Start must refuse to bind rather than fall back to gin's trust-everyone
// default when http.toml names a proxy that does not parse.
func TestHttpServerStartRefusesBadTrustedProxies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hs := &HttpServer{}
	us := &UdpServer{listenAddr: &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 62206}}
	err := hs.Start(us, &HttpConfig{
		HttpListenIp:   "127.0.0.1",
		HttpListenPort: 0,
		TrustedProxies: []string{"relay.opennhp.org"},
	})
	if err == nil {
		t.Fatal("Start with an unparsable TrustedProxies entry = nil, want an error")
	}
	if hs.IsRunning() {
		t.Fatal("HttpServer is running after Start refused the config")
	}
	// The refusal happens after signals.stop is allocated but before
	// http.Server is built and before wg.Add, so nothing should have been
	// bound or spawned. Stop() is the safety net for that: it returns early
	// on a non-running server, so this must not hang or nil-panic.
	if hs.httpServer != nil {
		t.Fatal("Start built an http.Server before refusing the config")
	}
	done := make(chan struct{})
	go func() { hs.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Stop blocked on a server that never started; Start leaked a goroutine or a listener")
	}
}

// handleHttpOpenResource must refuse an unparsable source address instead of
// handing it to the AC as the source of an ipset entry. knockSourceIP returns
// "" when RemoteAddr has no host:port, and a plugin can overwrite req.SrcIp
// after the route set it, so the guard covers both.
func TestHandleHttpOpenResourceRejectsUnparsableSourceIP(t *testing.T) {
	for _, srcIp := range []string{"", "   ", "not-an-ip", "203.0.113.7:443", "10.0.1.300"} {
		hs := &HttpServer{}
		req := &common.HttpKnockRequest{
			AuthServiceId: "example",
			ResourceId:    "demo",
			DeviceId:      "test-agent",
			SrcIp:         srcIp,
		}
		ack, err := hs.handleHttpOpenResource(req, &common.ResourceData{})
		if err != common.ErrInvalidIpAddress {
			t.Fatalf("handleHttpOpenResource(SrcIp=%q) err = %v, want ErrInvalidIpAddress", srcIp, err)
		}
		if ack != nil {
			t.Fatalf("handleHttpOpenResource(SrcIp=%q) returned an ack %+v, want nil", srcIp, ack)
		}
	}
}

// A well-formed address gets past the guard; the request then fails for the
// reason it should (no resources configured), which is what proves the guard
// is not just rejecting everything.
func TestHandleHttpOpenResourceAcceptsParsableSourceIP(t *testing.T) {
	for _, srcIp := range []string{"192.0.2.10", "2001:db8::1"} {
		hs := &HttpServer{}
		req := &common.HttpKnockRequest{
			AuthServiceId: "example",
			ResourceId:    "demo",
			DeviceId:      "test-agent",
			SrcIp:         srcIp,
		}
		if _, err := hs.handleHttpOpenResource(req, &common.ResourceData{}); err != common.ErrResourceNotFound {
			t.Fatalf("handleHttpOpenResource(SrcIp=%q) err = %v, want ErrResourceNotFound", srcIp, err)
		}
	}
}

// End-to-end through a real listener: the synthesized contexts above assume
// gin resolves trust through hs.ginEngine, this one proves it on a request
// that actually traveled over TCP. The probe handler stands in for the
// plugin routes, which build HttpKnockRequest.SrcIp the same way.
func TestKnockSourceIPOverRealConnection(t *testing.T) {
	for _, tc := range []struct {
		name           string
		trustedProxies []string
		want           string
	}{
		// The loopback peer sends X-Forwarded-For and is not trusted, so the
		// header is ignored: this is the spoofing attempt being refused.
		{"no trusted proxy", nil, "127.0.0.1"},
		// Same request, but the operator has declared loopback to be a proxy.
		{"loopback trusted", []string{"127.0.0.1", "::1"}, "203.0.113.7"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hs := newTestHttpServer(t, tc.trustedProxies)

			var got string
			hs.ginEngine.GET("/probe", func(ctx *gin.Context) {
				got = hs.knockSourceIP(ctx)
				ctx.Status(http.StatusOK)
			})

			srv := httptest.NewServer(hs.ginEngine)
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL+"/probe", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Forwarded-For", "203.0.113.7")
			req.Header.Set("X-Real-IP", "203.0.113.7")
			resp, err := srv.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()

			if got != tc.want {
				t.Fatalf("knockSourceIP = %q, want %q (TrustedProxies=%v)", got, tc.want, tc.trustedProxies)
			}
		})
	}
}

// A TrustedProxies edit does not reach a listener that is already running:
// updateHttpConfig only starts a server when there isn't one, so the live
// engine keeps the list it was started with. That is what
// trustedProxiesPending detects, so the caller can say so instead of letting
// an operator believe a removed proxy stopped being trusted.
func TestTrustedProxiesPending(t *testing.T) {
	for _, tc := range []struct {
		name       string
		started    []string
		configured []string
		want       bool
	}{
		{"unchanged", []string{"127.0.0.1"}, []string{"127.0.0.1"}, false},
		{"unchanged but reformatted", []string{"127.0.0.1"}, []string{"  127.0.0.1  ", ""}, false},
		// A bare address and its host prefix admit the same peer, and
		// isTrustedProxy tests every entry, so neither spelling nor order
		// changes who is trusted. Warning about these would spend the
		// Critical line's credibility on edits that change nothing.
		{"bare address spelled as a host prefix", []string{"127.0.0.1"}, []string{"127.0.0.1/32"}, false},
		{"v6 bare address spelled as a host prefix", []string{"::1"}, []string{"::1/128"}, false},
		{"equivalent prefix spelling", []string{"10.0.0.0/8"}, []string{"10.0.0.0/8"}, false},
		{"same entries reordered", []string{"127.0.0.1", "::1"}, []string{"::1", "127.0.0.1"}, false},
		{"duplicate entry", []string{"127.0.0.1"}, []string{"127.0.0.1", "127.0.0.1"}, false},
		{"proxy removed", []string{"127.0.0.1"}, nil, true},
		{"proxy added", nil, []string{"127.0.0.1"}, true},
		{"proxy replaced", []string{"127.0.0.1"}, []string{"10.0.0.0/8"}, true},
		{"entry appended", []string{"127.0.0.1"}, []string{"127.0.0.1", "::1"}, true},
		{"prefix widened", []string{"10.0.0.0/16"}, []string{"10.0.0.0/8"}, true},
		// A list gin would refuse is not equal to anything; report it rather
		// than guess.
		{"configured list does not parse", []string{"127.0.0.1"}, []string{"relay.opennhp.org"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hs := newTestHttpServer(t, tc.started)
			hs.running.Store(true)
			t.Cleanup(func() { hs.running.Store(false) })

			if got := hs.trustedProxiesPending(tc.configured); got != tc.want {
				t.Fatalf("trustedProxiesPending(%v) with a listener started on %v = %v, want %v",
					tc.configured, tc.started, got, tc.want)
			}
		})
	}
}

// Nothing is pending when no listener is running: the next Start applies the
// configured list outright, so reporting a difference would be noise.
func TestTrustedProxiesPendingIgnoresAStoppedServer(t *testing.T) {
	hs := newTestHttpServer(t, []string{"127.0.0.1"})
	if hs.trustedProxiesPending(nil) {
		t.Fatal("trustedProxiesPending reported a pending change on a server that is not running")
	}
	var nilServer *HttpServer
	if nilServer.trustedProxiesPending([]string{"127.0.0.1"}) {
		t.Fatal("trustedProxiesPending on a nil HttpServer reported a pending change")
	}
}

// writeSelfSignedTLSPair writes a throwaway cert/key pair under dir and
// returns their paths relative to it, in the shape http.toml uses.
func writeSelfSignedTLSPair(t *testing.T, dir string) (certRel, keyRel string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "nhp-server-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}

	certRel, keyRel = filepath.Join("cert", "cert.pem"), filepath.Join("cert", "cert.key")
	if err := os.MkdirAll(filepath.Join(dir, "cert"), 0o755); err != nil {
		t.Fatalf("mkdir cert: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, certRel),
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, keyRel),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	return certRel, keyRel
}

// A TLS listener must report itself running, exactly as a plain HTTP one does.
//
// Start's TLS branch returns from inside its own `if`, so a running.Store at
// the end of the function is reached only on the plain HTTP path. Everything
// that manages a live listener keys off that flag, and EnableTLS = true is
// what the shipped http.toml files ship: without it, EnableHttp = false never
// stops the listener, every reload binds onto a port already in use, and
// trustedProxiesPending can never report a pending proxy-list change.
func TestHttpServerStartMarksATLSListenerRunning(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	certRel, keyRel := writeSelfSignedTLSPair(t, dir)
	prevExeDir := ExeDirPath
	ExeDirPath = dir
	t.Cleanup(func() { ExeDirPath = prevExeDir })

	us := &UdpServer{
		listenAddr: &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0},
		log:        log.NewLogger("test", 0, dir, "server"),
	}
	hs := &HttpServer{}
	err := hs.Start(us, &HttpConfig{
		EnableHttp:     true,
		EnableTLS:      true,
		HttpListenIp:   "127.0.0.1",
		HttpListenPort: 0,
		TLSCertFile:    certRel,
		TLSKeyFile:     keyRel,
		TrustedProxies: []string{"127.0.0.1"},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(hs.Stop)

	if !hs.IsRunning() {
		t.Fatal("a TLS listener reports IsRunning() == false; Stop, updateHttpConfig and " +
			"trustedProxiesPending all key off this flag")
	}

	// The whole point of the flag for this PR: a proxy-list change on a live
	// TLS listener has to be reported as pending.
	if !hs.trustedProxiesPending(nil) {
		t.Fatal("removing the only trusted proxy was not reported as pending on a TLS listener")
	}
	if hs.trustedProxiesPending([]string{"127.0.0.1"}) {
		t.Fatal("an unchanged list was reported as pending on a TLS listener")
	}

	// Stop has to actually stop it, which it also refuses to do when the flag
	// is unset -- that is what made the documented EnableHttp cycle a no-op.
	hs.Stop()
	if hs.IsRunning() {
		t.Fatal("Stop left the TLS listener running")
	}
}

// Same assertion for the plain HTTP path, so a future edit cannot fix one and
// regress the other.
func TestHttpServerStartMarksAPlainListenerRunning(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	us := &UdpServer{
		listenAddr: &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0},
		log:        log.NewLogger("test", 0, dir, "server"),
	}
	hs := &HttpServer{}
	if err := hs.Start(us, &HttpConfig{
		EnableHttp:     true,
		HttpListenIp:   "127.0.0.1",
		HttpListenPort: 0,
		TrustedProxies: []string{"127.0.0.1"},
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(hs.Stop)

	if !hs.IsRunning() {
		t.Fatal("a plain HTTP listener reports IsRunning() == false")
	}
	if !hs.trustedProxiesPending(nil) {
		t.Fatal("removing the only trusted proxy was not reported as pending")
	}
	hs.Stop()
	if hs.IsRunning() {
		t.Fatal("Stop left the plain HTTP listener running")
	}
}

// A listener whose bind fails asynchronously must not keep claiming to be
// running. updateHttpConfig only builds a replacement when IsRunning is
// false, so a server stuck at true after a failed bind serves nothing and is
// never retried: the daemon stays dark until someone cycles EnableHttp.
// Clearing the flag lets the next http.toml reload retry, so freeing the port
// is enough to recover.
func TestHttpServerClearsRunningWhenTheBindFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Occupy a port, then point the server at it.
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer occupied.Close()
	port := occupied.Addr().(*net.TCPAddr).Port

	dir := t.TempDir()
	us := &UdpServer{
		listenAddr: &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0},
		log:        log.NewLogger("test", 0, dir, "server"),
	}
	hs := &HttpServer{}
	if err := hs.Start(us, &HttpConfig{
		EnableHttp:     true,
		HttpListenIp:   "127.0.0.1",
		HttpListenPort: port,
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(hs.Stop)

	deadline := time.Now().Add(10 * time.Second)
	for hs.IsRunning() {
		if time.Now().After(deadline) {
			t.Fatal("HttpServer still reports IsRunning() after the bind failed; " +
				"updateHttpConfig will never build a replacement")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// A clean Stop must not go through the failure path: ListenAndServe returns
// http.ErrServerClosed then, and treating that as a failure would have the
// listener goroutine race Stop's own bookkeeping.
func TestHttpServerStopIsNotTreatedAsAFailedBind(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	us := &UdpServer{
		listenAddr: &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0},
		log:        log.NewLogger("test", 0, dir, "server"),
	}
	hs := &HttpServer{}
	if err := hs.Start(us, &HttpConfig{
		EnableHttp:     true,
		HttpListenIp:   "127.0.0.1",
		HttpListenPort: 0,
	}); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !hs.IsRunning() {
		t.Fatal("the listener never came up")
	}
	hs.Stop()
	if hs.IsRunning() {
		t.Fatal("Stop left the listener running")
	}
	// Stop drains hs.wg, so the listener goroutine has already returned and
	// any store it was going to make has happened. A second Stop must be a
	// no-op rather than closing signals.stop twice.
	hs.Stop()
}
