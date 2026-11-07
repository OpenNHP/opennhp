package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/log"
)

// A fake executable exercises command errors and partial reconciliation without
// requiring root or ever invoking the host firewall. Real enforcement is tested
// separately in the disposable integration lab.
func TestFailurePolicyRules(t *testing.T) {
	var warnings []string
	logger := log.NewLogger("", 1, "", "")
	logger.Warning = func(format string, args ...any) {
		warnings = append(warnings, fmt.Sprintf(format, args...))
	}
	log.SetGlobalLogger(logger)
	t.Cleanup(func() { log.SetGlobalLogger(log.NewLogger("", 3, "", "")) })
	dir := t.TempDir()
	binary := filepath.Join(dir, "iptables")
	script := `#!/bin/sh
cd "$(dirname "$0")" || exit 2
# Emulate older iptables: bare -w only; a numeric wait becomes an invalid operation.
[ "$1" = -w ] || exit 2
shift
op=$1; chain=$2
shift 2
[ -e "$chain" ] || : > "$chain"
if [ -f fail ] && [ "$chain" = FORWARD ]; then echo injected-error >&2; exit 2; fi
case "$op" in
-N) : > "$chain" ;;
-A) if [ -f no-comment ]; then echo comment-unavailable >&2; exit 2; fi
    echo "-A $chain $*" >> "$chain" ;;
-F) if [ -f fail-probe-cleanup ]; then echo cleanup-failed >&2; exit 2; fi
    : > "$chain" ;;
-X) rm "$chain" ;;
-S) cat "$chain" ;;
-I) echo "-A $chain $*" >> "$chain" ;;
-D) awk -v rule="-A $chain $*" '$0 == rule && !removed { removed=1; next } { print }' "$chain" > next; mv next "$chain" ;;
*) exit 2 ;;
esac
`
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	table := &IPTables{Binary: binary}
	count := func(chain string) int {
		b, err := os.ReadFile(filepath.Join(dir, chain))
		if err != nil {
			t.Fatal(err)
		}
		return strings.Count(string(b), failurePolicyComment)
	}
	if err := table.InitializeFailurePolicy(true); err != nil {
		t.Fatal(err)
	}
	if err := table.SetFailurePolicyOpen(true); err != nil {
		t.Fatal(err)
	}
	if err := table.SetFailurePolicyOpen(true); err != nil {
		t.Fatal(err)
	}
	if count("INPUT") != 1 || count("FORWARD") != 1 {
		t.Fatal("duplicate or missing rule")
	}
	// Preserve method-value and interface compatibility for existing callers.
	var legacy interface {
		AcceptAllInput()
		ResetAllInput()
	} = table
	var reset func() = legacy.ResetAllInput
	reset()
	legacy.AcceptAllInput()
	if count("INPUT") != 1 || count("FORWARD") != 1 {
		t.Fatal("legacy wrappers failed")
	}

	// Simulate a crash, duplicates, and one chain failing during startup cleanup.
	f, err := os.OpenFile(filepath.Join(dir, "INPUT"), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("-A INPUT -m comment --comment " + failurePolicyComment + " -j ACCEPT\n")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if err := os.WriteFile(filepath.Join(dir, "fail"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	restarted := &IPTables{Binary: binary}
	if err := restarted.InitializeFailurePolicy(true); err == nil {
		t.Fatal("firewall failure ignored")
	}
	if restarted.failurePolicyStateKnown {
		t.Fatal("partial operation marked complete")
	}
	if count("INPUT") != 0 || count("FORWARD") != 1 {
		t.Fatal("unexpected partial cleanup")
	}
	if err := os.Remove(filepath.Join(dir, "fail")); err != nil {
		t.Fatal(err)
	}
	if err := restarted.InitializeFailurePolicy(true); err != nil {
		t.Fatal(err)
	}
	if count("INPUT") != 0 || count("FORWARD") != 0 {
		t.Fatal("stale rules remain")
	}
	// Insert failure after INPUT succeeds; retry must not duplicate INPUT.
	if err := os.WriteFile(filepath.Join(dir, "fail"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := restarted.SetFailurePolicyOpen(true); err == nil {
		t.Fatal("insert failure ignored")
	}
	if err := os.Remove(filepath.Join(dir, "fail")); err != nil {
		t.Fatal(err)
	}
	if err := restarted.SetFailurePolicyOpen(true); err != nil {
		t.Fatal(err)
	}
	if count("INPUT") != 1 || count("FORWARD") != 1 {
		t.Fatal("retry duplicated rule")
	}
	if err := restarted.SetFailurePolicyOpen(false); err != nil {
		t.Fatal(err)
	}
	for _, chain := range []string{"INPUT", "FORWARD"} {
		legacyRule := "-A " + chain + " -j ACCEPT"
		legacy := "-A " + chain + " -s 192.0.2.1 -j DROP\n" + legacyRule + "\n"
		if err := os.WriteFile(filepath.Join(dir, chain), []byte(legacy), 0600); err != nil {
			t.Fatal(err)
		}
		if err := restarted.InitializeFailurePolicy(true); err == nil || !strings.Contains(err.Error(), legacyRule) {
			t.Fatalf("legacy rule accepted: %v", err)
		}
		b, _ := os.ReadFile(filepath.Join(dir, chain))
		if string(b) != legacy {
			t.Fatal("operator rule deleted")
		}
		if err := restarted.InitializeFailurePolicy(false); err != nil {
			t.Fatal(err)
		}
		if len(warnings) == 0 || !strings.Contains(warnings[len(warnings)-1], legacyRule) {
			t.Errorf("fail-open did not warn about preserved legacy rule %q: %v", legacyRule, warnings)
		}
		b, _ = os.ReadFile(filepath.Join(dir, chain))
		if string(b) != legacy {
			t.Fatal("fail-open changed an unowned rule")
		}
		if err := os.WriteFile(filepath.Join(dir, chain), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	operator := "-A INPUT -s 192.0.2.1 -j DROP\n-A INPUT -m comment --comment operator-baseline -j ACCEPT\n"
	// Owned duplicates need not be at the top of the chain. Removing them must
	// preserve an operator's rules and their order.
	owned := "-A INPUT -m comment --comment " + failurePolicyComment + " -j ACCEPT\n"
	if err := os.WriteFile(filepath.Join(dir, "INPUT"), []byte(operator+owned+owned), 0600); err != nil {
		t.Fatal(err)
	}
	if err := restarted.InitializeFailurePolicy(true); err != nil {
		t.Fatal(err)
	}
	preserved, readErr := os.ReadFile(filepath.Join(dir, "INPUT"))
	if readErr != nil || string(preserved) != operator {
		t.Fatalf("intentional operator baseline changed: %s, %v", preserved, readErr)
	}
	if err := os.WriteFile(filepath.Join(dir, "INPUT"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fail-probe-cleanup"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := restarted.InitializeFailurePolicy(false); err == nil || !strings.Contains(err.Error(), "clean up probe") {
		t.Fatalf("probe cleanup failure ignored: %v", err)
	}
	if restarted.failurePolicyStateKnown {
		t.Fatal("failed probe cleanup marked complete")
	}
	if err := os.Remove(filepath.Join(dir, "fail-probe-cleanup")); err != nil {
		t.Fatal(err)
	}

	// Probe kernel comment support before a fail-open outage, leaving no scratch chains.
	if err := os.WriteFile(filepath.Join(dir, "no-comment"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := restarted.InitializeFailurePolicy(false); err == nil || !strings.Contains(err.Error(), "comment match") {
		t.Fatalf("missing comment support accepted: %v", err)
	}
	if restarted.failurePolicyStateKnown {
		t.Fatal("failed probe marked initialized")
	}
	if err := restarted.InitializeFailurePolicy(true); err != nil {
		t.Fatalf("fail-closed needlessly requires insert support: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, "no-comment")); err != nil {
		t.Fatal(err)
	}
	if err := restarted.InitializeFailurePolicy(false); err != nil {
		t.Fatal(err)
	}
	leftovers, globErr := filepath.Glob(filepath.Join(dir, "NHP-AC-*"))
	if globErr != nil || len(leftovers) != 0 {
		t.Fatalf("probe chains remain: %v, %v", leftovers, globErr)
	}
	if count("INPUT") != 0 || count("FORWARD") != 0 {
		t.Fatal("probe changed traffic chains")
	}

}

func TestFailurePolicyCommandTimeout(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "iptables")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nexec sleep 30\n"), 0700); err != nil {
		t.Fatal(err)
	}
	table := &IPTables{Binary: binary, failurePolicyTimeout: 100 * time.Millisecond}
	if err := table.InitializeFailurePolicy(false); err == nil {
		t.Fatal("hung firewall command did not return an initialization error")
	}
	if table.failurePolicyStateKnown {
		t.Fatal("timed-out cleanup marked complete")
	}
}

func TestFailurePolicyRuleOwnership(t *testing.T) {
	for _, comment := range []string{failurePolicyComment, "\"" + failurePolicyComment + "\"", "'" + failurePolicyComment + "'"} {
		if !isFailurePolicyRule("-A INPUT  -m comment --comment "+comment+" -j ACCEPT", "INPUT") {
			t.Fatal("owned rule not recognized", comment)
		}
	}
	for _, line := range []string{
		"-A FORWARD -m comment --comment opennhp-ac-fail-open -j ACCEPT",
		"-A INPUT -s 192.0.2.1 -m comment --comment opennhp-ac-fail-open -j ACCEPT",
		"-A INPUT -m comment --comment opennhp-ac-fail-open-operator -j ACCEPT",
		"-A INPUT -m comment --comment 'opennhp-ac-fail-open operator' -j ACCEPT",
		"-A INPUT -m comment --comment opennhp-ac-fail-open -j DROP",
	} {
		if isFailurePolicyRule(line, "INPUT") {
			t.Fatalf("operator rule claimed: %s", line)
		}
	}
}
