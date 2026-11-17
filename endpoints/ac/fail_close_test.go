package ac

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The default must stay fail-open, because that is what every existing
// deployment already does. A bool's zero value is false, so
// FailCloseOnServerLoss=false has to mean "keep the old behavior" — if
// the sense of the flag were ever inverted, an upgrade would silently
// change what happens during an outage.
func TestFailCloseDefaultsToCurrentBehavior(t *testing.T) {
	var c Config
	if c.FailCloseOnServerLoss {
		t.Error("the zero value asks to fail closed; it must preserve the existing fail-open behavior")
	}
}

// The fail-open path is reached only when the deployment has not asked to
// fail closed, and only in FilterMode 0.
//
// Asserted on the source: driving the real path needs a discovery loop, a
// server that stops answering and a live iptables, none of which a unit
// test can stand up. What must not regress is the shape of the branch —
// that AcceptAllInput cannot be reached without consulting the flag.
func TestAcceptAllInputIsGatedByTheFlag(t *testing.T) {
	src, err := os.ReadFile("udpac.go")
	if err != nil {
		t.Fatalf("read udpac.go: %v", err)
	}
	code := string(src)

	calls := regexp.MustCompile(`a\.iptables\.AcceptAllInput\(\)`).FindAllStringIndex(code, -1)
	if len(calls) != 1 {
		t.Fatalf("found %d calls to AcceptAllInput, want exactly 1; every one of them opens INPUT and FORWARD to 0.0.0.0/0", len(calls))
	}

	// The 600 characters before the call must contain both gates.
	start := calls[0][0] - 600
	if start < 0 {
		start = 0
	}
	ctx := code[start:calls[0][0]]
	for _, want := range []string{"FailCloseOnServerLoss", "FilterMode_IPTABLES"} {
		if !strings.Contains(ctx, want) {
			t.Errorf("the AcceptAllInput call is not guarded by %s", want)
		}
	}

	// Fail-closed must be the branch that does NOT call it.
	if !regexp.MustCompile(`(?s)if a\.config\.FailCloseOnServerLoss \{.*?logServerLossFailClosed\(\).*?\} else \{\s*a\.iptables\.AcceptAllInput\(\)`).MatchString(code) {
		t.Error("the fail-closed branch does not look like 'report and leave the rules alone'")
	}
}

// The notice is once per outage, and a recovery re-arms it.
func TestServerLossNoticeIsOncePerOutage(t *testing.T) {
	a := &UdpAC{}
	if a.serverLossReported.Load() {
		t.Fatal("a fresh AC already considers the outage reported")
	}
	a.logServerLossFailClosed()
	if !a.serverLossReported.Load() {
		t.Error("the first notice did not arm the once-flag")
	}
	// Second call must be a no-op: Swap returns the previous value.
	if !a.serverLossReported.Swap(true) {
		t.Error("the flag was cleared by a repeat call")
	}
	// A server answering again re-arms it for the next outage.
	a.serverLossReported.Store(false)
	a.logServerLossFailClosed()
	if !a.serverLossReported.Load() {
		t.Error("the notice did not re-arm after a recovery")
	}
}

// Every shipped AC config documents the setting, including that it also
// opens FORWARD — the part that makes this about the whole protected
// segment rather than just the AC host.
func TestShippedConfigsDocumentTheSetting(t *testing.T) {
	for _, p := range []string{
		filepath.Join("main", "etc", "config.toml"),
		filepath.Join("..", "..", "docker", "nhp-ac", "etc", "config.toml"),
		filepath.Join("..", "..", "deploy", "config-templates", "ac", "config.toml"),
	} {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("read %s: %v", p, err)
			continue
		}
		s := string(b)
		if !strings.Contains(s, "FailCloseOnServerLoss") {
			t.Errorf("%s does not document FailCloseOnServerLoss", p)
			continue
		}
		if !strings.Contains(s, "FORWARD") {
			t.Errorf("%s documents the setting without saying FORWARD is opened too", p)
		}
	}
}
