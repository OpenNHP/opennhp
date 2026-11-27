package ac

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
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

// Whether losing every server opens the filter is decided by the config
// alone, and the matrix is checked by running the decision.
//
// The previous version of this test located the AcceptAllInput call and
// scanned a fixed window of characters before it for two guard names,
// which breaks on a rename or a gofmt reflow and never runs the code.
func TestFailOpenOnServerLossMatrix(t *testing.T) {
	for name, tc := range map[string]struct {
		cfg  *Config
		want bool
	}{
		"iptables, default":     {&Config{FilterMode: FilterMode_IPTABLES}, true},
		"iptables, fail closed": {&Config{FilterMode: FilterMode_IPTABLES, FailCloseOnServerLoss: true}, false},
		"xdp, default":          {&Config{FilterMode: FilterMode_EBPFXDP}, false},
		"xdp, fail closed":      {&Config{FilterMode: FilterMode_EBPFXDP, FailCloseOnServerLoss: true}, false},
		"no config":             {nil, false},
	} {
		t.Run(name, func(t *testing.T) {
			if got := failOpenOnServerLoss(tc.cfg); got != tc.want {
				t.Errorf("failOpenOnServerLoss = %v, want %v", got, tc.want)
			}
		})
	}
}

// AcceptAllInput has exactly one call site and it is inside
// onAllServersLost, which consults the predicate above.
//
// This is the one thing the matrix cannot show — that the decision is
// actually wired to the call. Kept deliberately narrow: a call count and a
// containing function, not a regex over a window of source.
func TestAcceptAllInputHasOneGatedCallSite(t *testing.T) {
	src, err := os.ReadFile("udpac.go")
	if err != nil {
		t.Fatalf("read udpac.go: %v", err)
	}
	code := string(src)

	if n := strings.Count(code, "iptables.AcceptAllInput()"); n != 1 {
		t.Fatalf("%d call sites for AcceptAllInput, want exactly 1; each one opens INPUT and FORWARD to 0.0.0.0/0", n)
	}
	fn := regexp.MustCompile(`(?s)func \(a \*UdpAC\) onAllServersLost\(\) \{.*?\n\}`).FindString(code)
	if fn == "" {
		t.Fatal("onAllServersLost not found; re-anchor this test rather than deleting it")
	}
	if !strings.Contains(fn, "iptables.AcceptAllInput()") {
		t.Error("AcceptAllInput is called from outside onAllServersLost")
	}
	if !strings.Contains(fn, "failOpenOnServerLoss(") {
		t.Error("onAllServersLost does not consult failOpenOnServerLoss")
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
	// Call the function again rather than Swap directly: the claim is that
	// the function returns early, and poking the flag by hand would pass
	// even if it did not.
	a.logServerLossFailClosed()
	if !a.serverLossReported.Load() {
		t.Error("a repeat call cleared the once-flag")
	}
	// A server answering again re-arms it for the next outage.
	a.serverLossReported.Store(false)
	a.logServerLossFailClosed()
	if !a.serverLossReported.Load() {
		t.Error("the notice did not re-arm after a recovery")
	}
}

// Every shipped AC config documents the setting, says that fail-open also
// opens FORWARD, and — the part a string check cannot see — places the
// documented line where uncommenting it actually works.
//
// TOML scopes every key after a [table] header into that table. The first
// version of this change appended the block to the end of two files, which
// put it under [Metrics] and [UserData]; an operator who followed the
// comment would have set Metrics.FailCloseOnServerLoss, which
// toml.Unmarshal silently discards. The AC would then fail OPEN during an
// outage, with no error and no warning — the exact outcome the setting
// exists to prevent, reached by doing what the file said.
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

		// Uncomment the documented line and parse the file for real. This
		// is the check that catches a misplaced block: a key under a table
		// header unmarshals into that table and is dropped.
		live := regexp.MustCompile(`(?m)^#\s*FailCloseOnServerLoss\s*=.*$`).
			ReplaceAllString(s, "FailCloseOnServerLoss = true")
		if live == s {
			t.Errorf("%s has no commented FailCloseOnServerLoss line to uncomment", p)
			continue
		}
		var cfg Config
		if err := toml.Unmarshal([]byte(live), &cfg); err != nil {
			t.Errorf("%s does not parse after uncommenting the setting: %v", p, err)
			continue
		}
		if !cfg.FailCloseOnServerLoss {
			t.Errorf("%s: uncommenting the documented line does not set FailCloseOnServerLoss. "+
				"It is almost certainly below a [table] header, so the key lands in that table and is discarded.", p)
		}
	}
}

// An AC with no servers configured is not "every server unreachable".
//
// The discovery check totals the per-server failure flags and compares
// against the list length, so with an empty list 0 < 0 is false and the
// code fell through to the all-servers-lost branch. In fail-open mode that
// opened INPUT and FORWARD on a host that was never given a control plane;
// in fail-closed mode it logged that servers were unreachable, which is
// not what happened. It is now its own case.
func TestEmptyServerListIsNotAllServersLost(t *testing.T) {
	src, err := os.ReadFile("udpac.go")
	if err != nil {
		t.Fatalf("read udpac.go: %v", err)
	}
	code := string(src)

	loc := regexp.MustCompile(`totalFail < int32\(len\(discoveryFailStatusArr\)\)`).FindStringIndex(code)
	if loc == nil {
		t.Fatal("discovery comparison not found; re-anchor this test rather than deleting it")
	}
	// The empty-list case has to be decided before the comparison, or the
	// arithmetic answers it first.
	before := code[:loc[0]]
	guard := regexp.MustCompile(`len\(discoveryFailStatusArr\) == 0`).FindAllStringIndex(before, -1)
	if len(guard) == 0 {
		t.Error("nothing checks for an empty server list before the failure comparison")
	}
}

// Both postures say what they did, once per outage. The default is the
// exposing one, so the operator who never touched the setting is the one
// who most needs the line.
func TestBothPosturesAreReported(t *testing.T) {
	src, err := os.ReadFile("udpac.go")
	if err != nil {
		t.Fatalf("read udpac.go: %v", err)
	}
	fn := regexp.MustCompile(`(?s)func \(a \*UdpAC\) onAllServersLost\(\) \{.*?\n\}`).FindString(string(src))
	if fn == "" {
		t.Fatal("onAllServersLost not found")
	}
	if !strings.Contains(fn, "serverLossReported.Swap(true)") {
		t.Error("the fail-open branch does not report the posture once per outage")
	}
	if !strings.Contains(fn, "logServerLossFailClosed()") {
		t.Error("the fail-closed branch does not report the posture")
	}
}

// Stale 0.0.0.0/0 ACCEPT rules from an unclean exit are cleared at
// startup, not left to ResetAllInput — which returns early because
// AcceptInputMode is in-memory and starts false.
func TestStartupClearsStaleAcceptAllRules(t *testing.T) {
	src, err := os.ReadFile("udpac.go")
	if err != nil {
		t.Fatalf("read udpac.go: %v", err)
	}
	if !strings.Contains(string(src), "ClearStaleAcceptAllInput()") {
		t.Error("startup does not clear stale ACCEPT rules; an AC killed while failed open stays open " +
			"even after the operator sets FailCloseOnServerLoss")
	}

	ipt, err := os.ReadFile(filepath.Join("..", "..", "nhp", "utils", "iptables.go"))
	if err != nil {
		t.Fatalf("read iptables.go: %v", err)
	}
	fn := regexp.MustCompile(`(?s)func \(table \*IPTables\) ClearStaleAcceptAllInput\(\) \{.*?\n\}`).FindString(string(ipt))
	if fn == "" {
		t.Fatal("ClearStaleAcceptAllInput not found")
	}
	// It must not consult the flag; that is the whole point.
	if strings.Contains(fn, "if !table.AcceptInputMode") || strings.Contains(fn, "if table.AcceptInputMode") {
		t.Error("ClearStaleAcceptAllInput gates on AcceptInputMode, which is exactly the flag that is wrong after an unclean exit")
	}
	if !strings.Contains(fn, "DEL_IPTABLES_RULE_OPERATION") {
		t.Error("ClearStaleAcceptAllInput does not delete the rules")
	}
}
