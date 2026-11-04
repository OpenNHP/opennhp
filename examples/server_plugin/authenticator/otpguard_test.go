package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"

	"github.com/OpenNHP/opennhp/nhp/common"
	nhplog "github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/plugins"
)

const testOTPSecret = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestGuard() (*otpGuard, *fakeClock) {
	clk := &fakeClock{t: time.Unix(1_700_000_000, 0)}
	g := newOTPGuard()
	g.now = clk.now
	return g, clk
}

// guess makes one wrong attempt from ip and reports whether it was checked.
func guess(g *otpGuard, ip string) bool {
	_, ok := g.reserve(ip)
	return ok
}

// After otpMaxIPFailures failures an address is locked out for
// otpIPLockout; other addresses are not affected.
func TestOTPGuardLocksOutAnAddress(t *testing.T) {
	g, clk := newTestGuard()
	for i := 0; i < otpMaxIPFailures; i++ {
		if !guess(g, "192.0.2.1") {
			t.Fatalf("attempt %d refused before the limit", i+1)
		}
	}
	if guess(g, "192.0.2.1") {
		t.Fatal("address not locked out after the failure limit")
	}
	if !guess(g, "192.0.2.2") {
		t.Fatal("another address was locked out")
	}
	clk.advance(otpIPLockout + time.Second)
	if !guess(g, "192.0.2.1") {
		t.Fatal("lockout did not expire")
	}
}

// Failures spread over time are forgotten, and a success clears them.
func TestOTPGuardFailureCountResets(t *testing.T) {
	g, clk := newTestGuard()
	for i := 0; i < otpMaxIPFailures-1; i++ {
		guess(g, "192.0.2.1")
	}
	tk, _ := g.reserve("192.0.2.1")
	g.succeed(tk)
	for i := 0; i < otpMaxIPFailures-1; i++ {
		guess(g, "192.0.2.1")
	}
	if !guess(g, "192.0.2.1") {
		t.Fatal("a success did not clear the failure count")
	}

	for i := 0; i < otpMaxIPFailures-1; i++ {
		guess(g, "192.0.2.3")
	}
	clk.advance(otpIPFailureTTL + time.Second)
	for i := 0; i < otpMaxIPFailures-1; i++ {
		guess(g, "192.0.2.3")
	}
	if !guess(g, "192.0.2.3") {
		t.Fatal("stale failures still counted")
	}
}

// A correct but already-used code gives its attempt back, to the address
// and to the global budget, without clearing earlier failures.
func TestOTPGuardRelease(t *testing.T) {
	g, _ := newTestGuard()
	for i := 0; i < otpMaxIPFailures-1; i++ {
		guess(g, "192.0.2.4")
	}
	for i := 0; i < 10; i++ {
		tk, ok := g.reserve("192.0.2.4")
		if !ok {
			t.Fatalf("release %d did not give the attempt back", i)
		}
		g.release(tk)
	}
	if len(g.global) != otpMaxIPFailures-1 {
		t.Fatalf("global failures %d, want %d", len(g.global), otpMaxIPFailures-1)
	}
	if !guess(g, "192.0.2.4") || guess(g, "192.0.2.4") {
		t.Fatal("earlier failures were not kept")
	}
}

// Parallel attempts cannot all pass the check before any is counted.
func TestOTPGuardConcurrentReserve(t *testing.T) {
	g, _ := newTestGuard()
	var wg sync.WaitGroup
	var mu sync.Mutex
	admitted := 0
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if guess(g, "192.0.2.5") {
				mu.Lock()
				admitted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if admitted != otpMaxIPFailures {
		t.Fatalf("admitted %d parallel guesses, want %d", admitted, otpMaxIPFailures)
	}
}

// Guesses spread over many addresses still hit the global limit.
func TestOTPGuardGlobalLimit(t *testing.T) {
	g, clk := newTestGuard()
	for i := 0; i < otpGlobalMaxFailures; i++ {
		if !guess(g, "198.51.100."+strconv.Itoa(i)) {
			t.Fatalf("attempt %d refused before the global limit", i+1)
		}
	}
	if guess(g, "203.0.113.9") {
		t.Fatal("a fresh address was allowed past the global failure limit")
	}
	clk.advance(otpGlobalWindow)
	if !guess(g, "203.0.113.9") {
		t.Fatal("the global window did not reset")
	}
}

// The global limit is a sliding window: failures straddling a minute
// boundary cannot add up to twice the limit.
func TestOTPGuardGlobalSlidingWindow(t *testing.T) {
	g, clk := newTestGuard()
	clk.advance(otpGlobalWindow - time.Second)
	for i := 0; i < otpGlobalMaxFailures; i++ {
		guess(g, "198.51.100."+strconv.Itoa(i))
	}
	clk.advance(2 * time.Second) // a fixed window would have reset here
	if guess(g, "203.0.113.1") {
		t.Fatal("the global limit reset at a window boundary")
	}
	clk.advance(otpGlobalWindow)
	if !guess(g, "203.0.113.1") {
		t.Fatal("old failures never aged out")
	}
}

// A code can be used once while it is valid, unless reuse is allowed, and
// expired codes are pruned.
func TestOTPGuardConsume(t *testing.T) {
	g, clk := newTestGuard()
	life := otpCodeLifetime(30)
	if !g.consume("123456", false, life) {
		t.Fatal("fresh code refused")
	}
	if g.consume("123456", false, life) {
		t.Fatal("code accepted twice")
	}
	if !g.consume("123456", true, life) {
		t.Fatal("reuse refused with OTPAllowReuse")
	}
	clk.advance(life)
	if !g.consume("123456", false, life) || len(g.used) != 1 {
		t.Fatalf("code still refused or not pruned after it can no longer validate (%d left)", len(g.used))
	}
	if otpCodeLifetime(60) != 3*time.Minute {
		t.Fatalf("lifetime ignores the period: %v", otpCodeLifetime(60))
	}
}

// The address table stays bounded and refuses rather than tracking nothing.
func TestOTPGuardBounded(t *testing.T) {
	g, clk := newTestGuard()
	for i := 0; i < otpMaxTrackedIPs+10; i++ {
		g.global = nil // isolate from the global limit
		guess(g, "ip-"+strconv.Itoa(i))
	}
	if len(g.ips) > otpMaxTrackedIPs {
		t.Fatalf("tracked %d addresses, limit %d", len(g.ips), otpMaxTrackedIPs)
	}
	g.global = nil
	if guess(g, "one-more") {
		t.Fatal("checked an address it could not track")
	}
	clk.advance(otpIPFailureTTL + time.Second)
	if !guess(g, "one-more") {
		t.Fatal("stale addresses were not evicted")
	}
}

func TestOTPClientIP(t *testing.T) {
	trusted, err := parseTrustedProxies([]string{"127.0.0.1", "10.0.0.0/8"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parseTrustedProxies([]string{"127.0.0.1", "nope"}); err == nil {
		t.Fatal("an invalid entry was accepted")
	}
	cases := []struct {
		peer, xff string
		trusted   bool
		want      string
	}{
		{"192.0.2.1", "198.51.100.1", true, "192.0.2.1"},                       // untrusted peer: header ignored
		{"127.0.0.1", "198.51.100.1", false, "127.0.0.1"},                      // no trusted list: header ignored
		{"127.0.0.1", "198.51.100.1", true, "198.51.100.1"},                    // behind the proxy
		{"127.0.0.1", "6.6.6.6, 198.51.100.1, 10.1.2.3", true, "198.51.100.1"}, // forged left part skipped
		{"127.0.0.1", "garbage", true, "127.0.0.1"},                            // garbage: the proxy's limit
		{"127.0.0.1", "", true, "127.0.0.1"},
		{"2001:db8:1:2:aaaa::1", "", false, "2001:db8:1:2::/64"}, // IPv6 limited per /64
		{"127.0.0.1", "2001:db8:1:2:bbbb::9", true, "2001:db8:1:2::/64"},
	}
	gin.SetMode(gin.TestMode)
	for _, c := range cases {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		ctx.Request.RemoteAddr = net.JoinHostPort(c.peer, "5000")
		if c.xff != "" {
			ctx.Request.Header.Set("X-Forwarded-For", c.xff)
		}
		var tp []netip.Prefix
		if c.trusted {
			tp = trusted
		}
		if got := otpClientIP(ctx, tp); got != c.want {
			t.Errorf("peer %s xff %q: got %s, want %s", c.peer, c.xff, got, c.want)
		}
	}
}

// otpRequest drives the real handler as the login page does.
func otpRequest(t *testing.T, code, remoteAddr, forwardedFor string, knocked *int) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/plugins/authenticator?resid=demo&action=otpvalid&otpCode="+code, nil)
	ctx.Request.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		ctx.Request.Header.Set("X-Forwarded-For", forwardedFor)
	}
	helper := &plugins.HttpServerPluginHelper{
		AuthWithHttpCallbackFunc: func(*common.HttpKnockRequest, *common.ResourceData) (*common.ServerKnockAckMsg, error) {
			*knocked++
			return &common.ServerKnockAckMsg{ErrCode: common.ErrSuccess.ErrorCode()}, nil
		},
	}
	_, _ = authOTPCode(ctx, &common.HttpKnockRequest{}, &common.ResourceData{}, helper)
	return w
}

func useTestOTPConfig(t *testing.T, allowReuse bool) {
	t.Helper()
	oldConf, oldLimiter, oldLog := baseConf, otpLimiter, log
	baseConf = &config{OTPSecretKey: testOTPSecret, OTPPeriod: 30, OTPDigits: 6, OTPAllowReuse: allowReuse}
	otpLimiter = newOTPGuard()
	log = nhplog.NewLogger("", 0, "", "") // the plugin logger is normally set by Init
	t.Cleanup(func() { baseConf, otpLimiter, log = oldConf, oldLimiter, oldLog })
}

// Wrong codes from one address lock it out, even if it rotates
// X-Forwarded-For, and a correct code is then refused too.
func TestAuthOTPCodeLocksOutGuessing(t *testing.T) {
	useTestOTPConfig(t, false)
	knocked := 0
	for i := 0; i < otpMaxIPFailures; i++ {
		w := otpRequest(t, "000000", "192.0.2.10:5000", "10.0.0."+strconv.Itoa(i), &knocked)
		if w.Code != http.StatusOK {
			t.Fatalf("attempt %d: got %d before the limit", i+1, w.Code)
		}
	}
	good, _ := totp.GenerateCode(testOTPSecret, time.Now())
	if w := otpRequest(t, good, "192.0.2.10:5000", "10.9.9.9", &knocked); w.Code != http.StatusTooManyRequests {
		t.Fatalf("locked-out address: got %d, want 429", w.Code)
	}
	if knocked != 0 {
		t.Fatalf("knocked %d times for a locked-out address", knocked)
	}
	if w := otpRequest(t, good, "192.0.2.11:5000", "", &knocked); w.Code != http.StatusOK || knocked != 1 {
		t.Fatalf("other address with a valid code: got %d, knocked %d", w.Code, knocked)
	}
}

// A valid code is accepted once; a second use is refused unless
// OTPAllowReuse is set.
func TestAuthOTPCodeReplay(t *testing.T) {
	for _, allowReuse := range []bool{false, true} {
		useTestOTPConfig(t, allowReuse)
		knocked := 0
		code, _ := totp.GenerateCode(testOTPSecret, time.Now())
		otpRequest(t, code, "192.0.2.20:5000", "", &knocked)
		w := otpRequest(t, code, "192.0.2.21:5000", "", &knocked)
		var resp map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if allowReuse && knocked != 2 {
			t.Errorf("OTPAllowReuse: knocked %d times, want 2", knocked)
		}
		if !allowReuse && (knocked != 1 || resp["success"] != false) {
			t.Errorf("replayed code: knocked %d times, response %v", knocked, resp)
		}
	}
}

// A correct code refused only because it was already used is not counted
// as a guess: users sharing one seed must not lock each other out.
func TestAuthOTPCodeUsedCodeIsNotAFailure(t *testing.T) {
	useTestOTPConfig(t, false)
	knocked := 0
	code, _ := totp.GenerateCode(testOTPSecret, time.Now())
	otpRequest(t, code, "192.0.2.30:5000", "", &knocked)
	for i := 0; i < otpMaxIPFailures+2; i++ {
		otpRequest(t, code, "192.0.2.31:5000", "", &knocked)
	}
	if len(otpLimiter.global) != 0 || len(otpLimiter.ips) != 0 {
		t.Fatalf("used code counted as failure: global %d, ips %d", len(otpLimiter.global), len(otpLimiter.ips))
	}
}

// The configured OTPPeriod is what validation uses.
func TestValidateConfiguredOTPUsesPeriod(t *testing.T) {
	svc := GetQRAuthService()
	now := time.Now()
	code60, err := totp.GenerateCodeCustom(testOTPSecret, now, totp.ValidateOpts{Period: 60, Digits: 6})
	if err != nil {
		t.Fatal(err)
	}
	if !svc.ValidateConfiguredOTP(testOTPSecret, code60, 60) {
		t.Fatal("a 60s code was refused with OTPPeriod=60")
	}
}

func testResource(resID string) *common.ResourceData {
	res := &common.ResourceData{}
	res.ResourceId = resID
	return res
}

const qrBrowserAddr = "192.0.2.40:5000" // where qrGenerate's browser connects from

func qrvalidRequest(t *testing.T, sessionID, resID string, knocked *int) *httptest.ResponseRecorder {
	t.Helper()
	return qrvalidRequestFrom(t, sessionID, resID, qrBrowserAddr, knocked)
}

func qrvalidRequestFrom(t *testing.T, sessionID, resID, remoteAddr string, knocked *int) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/plugins/authenticator?action=qrvalid&sessionId="+sessionID, nil)
	ctx.Request.RemoteAddr = remoteAddr
	helper := &plugins.HttpServerPluginHelper{
		AuthWithHttpCallbackFunc: func(*common.HttpKnockRequest, *common.ResourceData) (*common.ServerKnockAckMsg, error) {
			*knocked++
			return &common.ServerKnockAckMsg{ErrCode: common.ErrSuccess.ErrorCode()}, nil
		},
	}
	_, _ = authQRCode(ctx, &common.HttpKnockRequest{}, testResource(resID), helper)
	return w
}

func qrHandlerRequest(t *testing.T, target, remoteAddr string, h func(*gin.Context)) map[string]any {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
	ctx.Request.RemoteAddr = remoteAddr
	h(ctx)
	resp := map[string]any{"httpStatus": float64(w.Code)}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp
}

// generate hands the browser a session, as the login page does.
func qrGenerate(t *testing.T) (sessionID, qrData string, resp map[string]any) {
	t.Helper()
	resp = qrHandlerRequest(t, "/?action=generate", qrBrowserAddr, func(ctx *gin.Context) { HandleQRGenerate(ctx, "demo") })
	sessionID, _ = resp["sessionId"].(string)
	qrData, _ = resp["qrData"].(string)
	if sessionID == "" || qrData == "" {
		t.Fatalf("generate failed: %v", resp)
	}
	return sessionID, qrData, resp
}

func qrVerify(t *testing.T, qrData, code, phoneAddr string) map[string]any {
	t.Helper()
	q := url.Values{"action": {"verify"}, "encryptedData": {qrData}, "otpCode": {code}}
	return qrHandlerRequest(t, "/?"+q.Encode(), phoneAddr, HandleQRVerify)
}

// The QR flow cannot be completed with what generate hands out: there is
// no per-session secret, and confirming a session takes a code from the
// configured OTPSecretKey. A confirmed session knocks once, for its own
// resource.
func TestQRFlowRequiresConfiguredOTP(t *testing.T) {
	useTestOTPConfig(t, false)
	knocked := 0

	qrvalidRequest(t, "made-up", "demo", &knocked)
	sessionID, qrData, resp := qrGenerate(t)
	for _, leaked := range []string{"otpSecret", "otpUri"} {
		if _, ok := resp[leaked]; ok {
			t.Fatalf("generate returned %s", leaked)
		}
	}
	qrvalidRequest(t, sessionID, "demo", &knocked)

	// A code from any secret the browser could have made up does not confirm.
	other, _ := totp.Generate(totp.GenerateOpts{Issuer: "x", AccountName: "y"})
	forged, _ := totp.GenerateCode(other.Secret(), time.Now())
	if r := qrVerify(t, qrData, forged, "192.0.2.41:5000"); r["success"] != false {
		t.Fatalf("verify accepted a code from another secret: %v", r)
	}
	qrvalidRequest(t, sessionID, "demo", &knocked)
	if knocked != 0 {
		t.Fatalf("knocked %d times without a confirmed session", knocked)
	}

	good, _ := totp.GenerateCode(testOTPSecret, time.Now())
	if r := qrVerify(t, qrData, good, "192.0.2.41:5000"); r["success"] != true {
		t.Fatalf("verify refused the configured code: %v", r)
	}
	qrvalidRequest(t, sessionID, "other", &knocked)
	if knocked != 0 {
		t.Fatal("knocked for a different resource")
	}
	// Someone else who saw the session ID in the QR code cannot take the knock.
	qrvalidRequestFrom(t, sessionID, "demo", "198.51.100.77:5000", &knocked)
	if knocked != 0 {
		t.Fatal("knocked for a client other than the one that generated the session")
	}
	qrvalidRequest(t, sessionID, "demo", &knocked)
	qrvalidRequest(t, sessionID, "demo", &knocked)
	if knocked != 1 {
		t.Fatalf("confirmed session knocked %d times, want 1", knocked)
	}
}

// verify shares otpvalid's limiter: guesses there lock the address out of
// both paths, and a fresh session per guess does not help.
func TestQRVerifyIsRateLimited(t *testing.T) {
	useTestOTPConfig(t, false)
	for i := 0; i < otpMaxIPFailures; i++ {
		_, qrData, _ := qrGenerate(t)
		qrVerify(t, qrData, "000000", "192.0.2.60:5000")
	}
	_, qrData, _ := qrGenerate(t)
	good, _ := totp.GenerateCode(testOTPSecret, time.Now())
	if r := qrVerify(t, qrData, good, "192.0.2.60:5000"); r["httpStatus"] != float64(http.StatusTooManyRequests) {
		t.Fatalf("locked-out phone address: got %v, want 429", r)
	}
	knocked := 0
	if w := otpRequest(t, good, "192.0.2.60:5000", "", &knocked); w.Code != http.StatusTooManyRequests || knocked != 0 {
		t.Fatalf("otpvalid not locked out by verify guesses: %d, knocked %d", w.Code, knocked)
	}
}

// A QR session dies after qrMaxOTPFailures wrong codes, even for the right one.
func TestVerifySessionCapsFailures(t *testing.T) {
	useTestOTPConfig(t, false)
	svc := GetQRAuthService()
	_, qr, err := svc.GenerateSession("authenticator", "demo", "192.0.2.50", "https://example")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < qrMaxOTPFailures; i++ {
		_ = svc.VerifySession(qr.SessionID, qr.Token, "", "", func() error { return errOTPInvalid })
	}
	called := false
	if err := svc.VerifySession(qr.SessionID, qr.Token, "", "", func() error { called = true; return nil }); err == nil || called {
		t.Fatalf("session checked a code after the failure limit (err %v, checked %v)", err, called)
	}
}

// The shared seed is not published on the login page unless ShowOTPSecret is on.
func TestLoginPageHidesSecret(t *testing.T) {
	for _, show := range []bool{false, true} {
		useTestOTPConfig(t, false)
		baseConf.ShowOTPSecret = show
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		ctx, engine := gin.CreateTestContext(w)
		page, err := os.ReadFile("templates/authenticator_login.html")
		if err != nil {
			t.Fatal(err)
		}
		engine.SetHTMLTemplate(template.Must(template.New("authenticator/authenticator_login.html").Parse(string(page))))
		ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		_, _ = showLoginPage(ctx, &common.HttpKnockRequest{}, testResource("demo"), nil)
		if got := strings.Contains(w.Body.String(), testOTPSecret); got != show {
			t.Errorf("ShowOTPSecret=%v: secret in page = %v", show, got)
		}
	}
}

// Status polling runs concurrently with verify and qrvalid; run with -race.
func TestQRStatusConcurrentWithVerify(t *testing.T) {
	useTestOTPConfig(t, true)
	sessionID, qrData, _ := qrGenerate(t)
	good, _ := totp.GenerateCode(testOTPSecret, time.Now())
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			qrHandlerRequest(t, "/?action=status&sessionId="+sessionID, qrBrowserAddr, HandleQRStatus)
		}()
		go func() {
			defer wg.Done()
			qrVerify(t, qrData, good, "192.0.2.41:5000")
			knocked := 0
			qrvalidRequest(t, sessionID, "demo", &knocked)
		}()
	}
	wg.Wait()
	if r := qrHandlerRequest(t, "/?action=status&sessionId="+sessionID, qrBrowserAddr, HandleQRStatus); r["status"] != float64(QRStatusConfirmed) {
		t.Fatalf("status after verify: %v", r)
	}
}

// The missing-TrustedProxies warning fires for a private peer sending
// X-Forwarded-For, and not for a public one or one without the header.
func TestWarnUntrustedForwarding(t *testing.T) {
	useTestOTPConfig(t, false)
	gin.SetMode(gin.TestMode)
	try := func(peer, xff string) bool {
		otpProxyWarned.Store(false)
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		ctx.Request.RemoteAddr = net.JoinHostPort(peer, "5000")
		if xff != "" {
			ctx.Request.Header.Set("X-Forwarded-For", xff)
		}
		if got := otpClientAddr(ctx, nil); got != peer {
			t.Fatalf("header was honored without TrustedProxies: %s", got)
		}
		return otpProxyWarned.Load()
	}
	t.Cleanup(func() { otpProxyWarned.Store(false) })
	if !try("127.0.0.1", "198.51.100.1") || !try("10.0.0.5", "198.51.100.1") {
		t.Fatal("no warning for a proxy-looking peer")
	}
	if try("198.51.100.9", "198.51.100.1") || try("127.0.0.1", "") {
		t.Fatal("warned without a proxy-looking peer and header")
	}
}

// Rotating /64s inside one IPv6 /48 hits the /48 cap instead of draining
// the global budget; other /48s are unaffected.
func TestOTPGuardIPv6PrefixCap(t *testing.T) {
	g, _ := newTestGuard()
	admitted := 0
	for i := 0; i < 100; i++ {
		key := otpLimitKey(fmt.Sprintf("2001:db8:1:%x::1", i))
		if guess(g, key) {
			admitted++
		}
	}
	if admitted != otpMaxPrefixFailures {
		t.Fatalf("one /48 got %d guesses, want %d", admitted, otpMaxPrefixFailures)
	}
	if len(g.global) >= otpGlobalMaxFailures {
		t.Fatal("one /48 exhausted the global budget")
	}
	if !guess(g, otpLimitKey("2001:db8:2::1")) {
		t.Fatal("another /48 was locked out")
	}
	// A success gives the /48 its attempt back without clearing the others.
	g2, _ := newTestGuard()
	for i := 0; i < otpMaxPrefixFailures-1; i++ {
		guess(g2, otpLimitKey(fmt.Sprintf("2001:db8:1:%x::1", i)))
	}
	tk, ok := g2.reserve(otpLimitKey("2001:db8:1:ff::1"))
	if !ok {
		t.Fatal("refused before the /48 cap")
	}
	g2.succeed(tk)
	if !guess(g2, otpLimitKey("2001:db8:1:fe::1")) || guess(g2, otpLimitKey("2001:db8:1:fd::1")) {
		t.Fatal("success did not refund exactly one /48 attempt")
	}
}

// Access is opened for the address that authenticated, never for a forged
// X-Forwarded-For.
func TestKnockUsesAuthenticatedAddress(t *testing.T) {
	useTestOTPConfig(t, true)
	var opened []string
	helper := &plugins.HttpServerPluginHelper{
		AuthWithHttpCallbackFunc: func(req *common.HttpKnockRequest, _ *common.ResourceData) (*common.ServerKnockAckMsg, error) {
			opened = append(opened, req.SrcIp)
			return &common.ServerKnockAckMsg{ErrCode: common.ErrSuccess.ErrorCode()}, nil
		},
	}
	gin.SetMode(gin.TestMode)
	do := func(target string, h func(*gin.Context, *common.HttpKnockRequest, *common.ResourceData, *plugins.HttpServerPluginHelper) (*common.ServerKnockAckMsg, error)) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
		ctx.Request.RemoteAddr = qrBrowserAddr
		ctx.Request.Header.Set("X-Forwarded-For", "203.0.113.66")
		// What nhp-server puts in req.SrcIp: gin's ClientIP, i.e. the header.
		_, _ = h(ctx, &common.HttpKnockRequest{SrcIp: ctx.ClientIP()}, testResource("demo"), helper)
	}
	good, _ := totp.GenerateCode(testOTPSecret, time.Now())
	do("/?action=otpvalid&otpCode="+good, authOTPCode)

	sessionID, qrData, _ := qrGenerate(t)
	qrVerify(t, qrData, good, "192.0.2.41:5000")
	do("/?action=qrvalid&sessionId="+sessionID, authQRCode)

	want := []string{"192.0.2.40", "192.0.2.40"}
	if len(opened) != 2 || opened[0] != want[0] || opened[1] != want[1] {
		t.Fatalf("opened access for %v, want %v", opened, want)
	}
}

// Parallel verifies of one session cannot check more codes than its cap,
// and the code is checked without holding the service lock.
func TestVerifySessionParallelCap(t *testing.T) {
	useTestOTPConfig(t, false)
	svc := GetQRAuthService()
	_, qr, err := svc.GenerateSession("authenticator", "demo", "192.0.2.70", "https://example")
	if err != nil {
		t.Fatal(err)
	}
	release := make(chan struct{})
	var mu sync.Mutex
	checked := 0
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = svc.VerifySession(qr.SessionID, qr.Token, "", "", func() error {
				mu.Lock()
				checked++
				mu.Unlock()
				<-release
				return errOTPInvalid
			})
		}()
	}
	// While codes are being checked, status polling is not blocked.
	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		n := checked
		mu.Unlock()
		if n == qrMaxOTPFailures || time.Now().After(deadline) {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if svc.GetSession(qr.SessionID) == nil {
		t.Fatal("session lookup failed while codes were being checked")
	}
	close(release)
	wg.Wait()
	if checked != qrMaxOTPFailures {
		t.Fatalf("checked %d codes for one session, cap %d", checked, qrMaxOTPFailures)
	}
}

// An unauthenticated scan notification never moves a session backwards.
func TestScanDoesNotRegressSession(t *testing.T) {
	useTestOTPConfig(t, false)
	sessionID, qrData, _ := qrGenerate(t)
	good, _ := totp.GenerateCode(testOTPSecret, time.Now())
	qrVerify(t, qrData, good, "192.0.2.41:5000")
	qrHandlerRequest(t, "/?action=scan&sessionId="+sessionID, "198.51.100.5:5000", HandleQRScan)
	if r := qrHandlerRequest(t, "/?action=status&sessionId="+sessionID, qrBrowserAddr, HandleQRStatus); r["status"] != float64(QRStatusConfirmed) {
		t.Fatalf("scan moved a confirmed session: %v", r)
	}
	knocked := 0
	qrvalidRequest(t, sessionID, "demo", &knocked)
	if knocked != 1 {
		t.Fatal("confirmed session could not knock after a scan notification")
	}
}

// IPv4-mapped TrustedProxies prefixes match the plain IPv4 peers they
// describe; ones wider than the mapped range are refused.
func TestTrustedProxiesIPv4Mapped(t *testing.T) {
	trusted, err := parseTrustedProxies([]string{"::ffff:10.0.0.0/104", "::ffff:192.0.2.7"})
	if err != nil {
		t.Fatal(err)
	}
	for _, peer := range []string{"10.1.2.3", "192.0.2.7", "::ffff:10.9.9.9"} {
		if !isTrustedProxy(peer, trusted) {
			t.Errorf("%s not matched by an IPv4-mapped entry", peer)
		}
	}
	if isTrustedProxy("11.0.0.1", trusted) {
		t.Error("matched outside the mapped prefix")
	}
	if _, err := parseTrustedProxies([]string{"::ffff:0:0/80"}); err == nil {
		t.Error("an IPv4-mapped prefix shorter than /96 was accepted")
	}
}

// With no parsable peer address, nothing is limited under "" and nothing
// is knocked for an empty address: every entry point refuses.
func TestEmptyClientAddressIsRefused(t *testing.T) {
	useTestOTPConfig(t, true)
	const badPeer = "not-an-address" // what a unix-socket listener can leave in RemoteAddr
	knocked := 0
	good, _ := totp.GenerateCode(testOTPSecret, time.Now())
	if w := otpRequest(t, good, badPeer, "", &knocked); w.Code != http.StatusBadRequest || knocked != 0 {
		t.Fatalf("otpvalid: got %d, knocked %d", w.Code, knocked)
	}
	if r := qrHandlerRequest(t, "/?action=generate", badPeer, func(ctx *gin.Context) { HandleQRGenerate(ctx, "demo") }); r["httpStatus"] != float64(http.StatusBadRequest) {
		t.Fatalf("generate: %v", r)
	}
	sessionID, qrData, _ := qrGenerate(t)
	if r := qrVerify(t, qrData, good, badPeer); r["httpStatus"] != float64(http.StatusBadRequest) {
		t.Fatalf("verify: %v", r)
	}
	qrVerify(t, qrData, good, "192.0.2.41:5000")
	if w := qrvalidRequestFrom(t, sessionID, "demo", badPeer, &knocked); w.Code != http.StatusBadRequest || knocked != 0 {
		t.Fatalf("qrvalid: got %d, knocked %d", w.Code, knocked)
	}
	if _, ok := otpLimiter.reserve(""); ok {
		t.Fatal("reserve admitted an empty key")
	}
	if len(otpLimiter.ips) != 0 {
		t.Fatalf("tracked %d keys for refused requests", len(otpLimiter.ips))
	}
}

// Reaching the global limit is logged at most once per window.
func TestGlobalLimitWarningOncePerWindow(t *testing.T) {
	useTestOTPConfig(t, false)
	g, clk := newTestGuard()
	for i := 0; i < otpGlobalMaxFailures; i++ {
		guess(g, "198.51.100."+strconv.Itoa(i))
	}
	if !g.globalWarned.IsZero() {
		t.Fatal("warned before the limit was reached")
	}
	guess(g, "203.0.113.1")
	first := g.globalWarned
	if first.IsZero() {
		t.Fatal("no warning when the limit was reached")
	}
	clk.advance(time.Second)
	guess(g, "203.0.113.2")
	if !g.globalWarned.Equal(first) {
		t.Fatal("warned twice within one window")
	}
}
