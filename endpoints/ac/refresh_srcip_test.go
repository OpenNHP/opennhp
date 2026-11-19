package ac

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// pathSafeAccessToken issues a token whose base64 has no '/' or '+'.
//
// Those two don't survive a URL path segment: '/' splits it, and '+'
// becomes a space in the handler's unescape. The character set is the
// only thing being worked around; the token is otherwise exactly what a
// client would present.
func pathSafeAccessToken(t *testing.T, a *UdpAC, entry *AccessEntry) string {
	t.Helper()
	for i := 0; i < 100; i++ {
		token := a.GenerateAccessToken(entry)
		if !strings.ContainsAny(token, "/+") {
			return token
		}
	}
	t.Fatal("no path-safe access token after 100 attempts")
	return ""
}

// A refresh from an address the entry already covers goes through the
// route and succeeds.
//
// This is the test the first version of this change lacked, and its
// absence is why a regression got as far as review: switching the source
// to ctx.RemoteIP() refused every legitimate refresh, because this
// listener binds 127.0.0.1 only (HttpAC.Start) while an entry holds the
// agent's real address from the knock. The old tests called the handler
// directly with SrcIp pre-filled, or matched the source text, so neither
// went through the route and neither noticed.
func TestRefreshFromAuthorizedAddressSucceeds(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	// FilterMode_IPTABLES with no ipset: HandleAccessControl refuses at
	// its first step. That is deliberate — programming a real filter is
	// not something a unit test can do, and this gives the authorized
	// path a deterministic stop just past the authorization decision, so
	// the two outcomes are told apart by which refusal comes back.
	a := &UdpAC{
		tokenStore: common.NewTokenStore[*AccessEntry](),
		config:     &Config{FilterMode: FilterMode_IPTABLES},
	}
	entry := &AccessEntry{
		User:     &common.AgentUser{},
		SrcAddrs: []*common.NetAddress{{Ip: "10.0.0.5", Port: 443, Protocol: "tcp"}},
		DstAddrs: []*common.NetAddress{{Ip: "10.0.0.9", Port: 443, Protocol: "tcp"}},
		OpenTime: 30,
	}
	entry.ExpireTime = time.Now().Add(30 * time.Second)
	token := pathSafeAccessToken(t, a, entry)

	ha := &HttpAC{ua: a, ginEngine: gin.New()}
	ha.initRouter()

	// The connection is loopback, as it always is here; the real client
	// address arrives as the parameter the local front end passes.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		"/refresh/"+token+"?srcip=10.0.0.5", nil)
	req.RemoteAddr = "127.0.0.1:40000"
	ha.ginEngine.ServeHTTP(w, req)

	body := w.Body.String()
	if strings.Contains(body, "not authorized") {
		t.Fatalf("a refresh from an authorized address was refused: %s", body)
	}
	if !strings.Contains(body, common.ErrACIPSetNotFound.Error()) {
		t.Fatalf("response = %q, want the refresh to have reached the filter step", body)
	}
	if len(entry.SrcAddrs) != 1 {
		t.Errorf("the entry grew to %d source addresses", len(entry.SrcAddrs))
	}
}

// And one naming an address the entry does not cover is refused, through
// the same route.
func TestRefreshFromUnauthorizedAddressIsRefusedViaRoute(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	a := &UdpAC{
		tokenStore: common.NewTokenStore[*AccessEntry](),
		config:     &Config{FilterMode: FilterMode_IPTABLES},
	}
	entry := &AccessEntry{
		User:     &common.AgentUser{},
		SrcAddrs: []*common.NetAddress{{Ip: "10.0.0.5", Port: 443, Protocol: "tcp"}},
		DstAddrs: []*common.NetAddress{{Ip: "10.0.0.9", Port: 443, Protocol: "tcp"}},
		OpenTime: 30,
	}
	entry.ExpireTime = time.Now().Add(30 * time.Second)
	token := pathSafeAccessToken(t, a, entry)

	ha := &HttpAC{ua: a, ginEngine: gin.New()}
	ha.initRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet,
		"/refresh/"+token+"?srcip=203.0.113.7", nil)
	req.RemoteAddr = "127.0.0.1:40000"
	ha.ginEngine.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "not authorized") {
		t.Errorf("response = %q, want a refusal", w.Body.String())
	}
	if len(entry.SrcAddrs) != 1 {
		t.Errorf("the entry grew to %d source addresses; a refresh must not enroll one", len(entry.SrcAddrs))
	}
	for _, addr := range entry.SrcAddrs {
		if addr.Ip == "203.0.113.7" {
			t.Error("the unauthorized address was added to the entry")
		}
	}
}

// A refresh from an address the entry does not cover is refused, not
// enrolled. Refresh extends an existing grant; enrollment belongs to the
// knock, which the NHP protocol authenticates.
func TestRefreshDoesNotEnrollNewAddresses(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	a := &UdpAC{tokenStore: common.NewTokenStore[*AccessEntry]()}
	entry := &AccessEntry{
		User:     &common.AgentUser{},
		SrcAddrs: []*common.NetAddress{{Ip: "10.0.0.5", Port: 443, Protocol: "tcp"}},
		DstAddrs: []*common.NetAddress{{Ip: "10.0.0.9", Port: 443, Protocol: "tcp"}},
		OpenTime: 30,
	}
	entry.ExpireTime = time.Now().Add(30 * time.Second)
	token := a.GenerateAccessToken(entry)

	ha := &HttpAC{ua: a}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/refresh/"+token, nil)
	c.Request.RemoteAddr = "203.0.113.7:51234" // not in SrcAddrs

	ha.HandleHttpRefreshOperations(c, &common.HttpRefreshRequest{Token: token, SrcIp: "203.0.113.7"})

	if len(entry.SrcAddrs) != 1 {
		t.Errorf("the entry grew to %d source addresses; a refresh must not enroll one", len(entry.SrcAddrs))
	}
	for _, addr := range entry.SrcAddrs {
		if addr.Ip == "203.0.113.7" {
			t.Error("the unauthorized address was added to the entry")
		}
	}
	if !strings.Contains(w.Body.String(), "not authorized") {
		t.Errorf("response = %q, want a refusal naming the unauthorized source", w.Body.String())
	}
}

// Verification resets the lifetime rather than accumulating it. The old
// form added OpenTime to whatever was left, so polling kept an entry alive
// indefinitely without another knock.
func TestVerifyAccessTokenResetsRatherThanAccumulates(t *testing.T) {
	a := &UdpAC{tokenStore: common.NewTokenStore[*AccessEntry]()}
	entry := &AccessEntry{OpenTime: 30}
	entry.ExpireTime = time.Now().Add(30 * time.Second)
	token := a.GenerateAccessToken(entry)

	for i := 0; i < 5; i++ {
		if got := a.VerifyAccessToken(token); got == nil {
			t.Fatalf("verification %d failed", i)
		}
	}

	// Five verifications must not buy five windows.
	remaining := time.Until(entry.ExpireTime)
	full := time.Duration(entry.OpenTime+accessTokenLatePacketBufferSeconds) * time.Second
	if remaining > full+time.Second {
		t.Errorf("after 5 verifications the entry has %v left; one window is %v, so the expiry is accumulating", remaining, full)
	}
	// The window must match GenerateAccessToken's, buffer included: a
	// refresh that returned only OpenTime would shorten the token.
	if remaining < full-time.Second {
		t.Errorf("after verification the entry has %v left, want about %v (OpenTime plus the late-packet buffer)", remaining, full)
	}
}
