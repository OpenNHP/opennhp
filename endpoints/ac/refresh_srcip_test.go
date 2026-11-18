package ac

import (
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// The refresh source address comes from the connection, never from the
// request.
//
// It used to be ctx.Query("srcip"). Combined with the enrollment below,
// one access token opened the protected destination to any address its
// holder named. Asserted on the source because the binding happens in a
// route closure built by initRouter, which needs a live listener to
// exercise; what must not regress is which function supplies the value.
func TestRefreshSourceComesFromTheConnection(t *testing.T) {
	src, err := os.ReadFile("httpac.go")
	if err != nil {
		t.Fatalf("read httpac.go: %v", err)
	}
	code := stripACLineComments(string(src))

	m := regexp.MustCompile(`(?s)common\.HttpRefreshRequest\{.*?\}`).FindString(code)
	if m == "" {
		t.Fatal("HttpRefreshRequest literal not found; re-anchor this test rather than deleting it")
	}
	if strings.Contains(m, `Query(`) {
		t.Errorf("the refresh source address is read from the request:\n%s", m)
	}
	if !strings.Contains(m, "RemoteIP()") {
		t.Errorf("the refresh source address does not come from RemoteIP():\n%s", m)
	}
	// ClientIP() would be just as wrong: the engine is gin.New() with no
	// SetTrustedProxies, so gin trusts every proxy and ClientIP() reads
	// X-Forwarded-For, which the caller writes.
	if strings.Contains(m, "ClientIP()") {
		t.Errorf("the refresh source address comes from ClientIP(), which honors caller-supplied forwarding headers:\n%s", m)
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
	if remaining > 35*time.Second {
		t.Errorf("after 5 verifications the entry has %v left; OpenTime is 30s, so the expiry is accumulating", remaining)
	}
	if remaining < 25*time.Second {
		t.Errorf("after verification the entry has only %v left; a successful use should restore the full window", remaining)
	}
}

func stripACLineComments(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}
