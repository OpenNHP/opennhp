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

func failurePolicyTestTable(t *testing.T) (*IPTables, string) {
	t.Helper()
	dir := t.TempDir()
	binary := filepath.Join(dir, "iptables")
	script := `#!/bin/sh
cd "$(dirname "$0")" || exit 2
[ "$1" = -w ] || exit 2
shift
op=$1; chain=$2
if [ "$op" = -S ] && [ -z "$chain" ]; then
    for c in INPUT FORWARD NHP-AC-*; do
        [ -f "$c" ] || continue
        case "$c" in NHP-AC-*) echo "-N $c" ;; esac
        cat "$c"
    done
    exit 0
fi
shift 2
case "$chain" in INPUT|FORWARD) [ -e "$chain" ] || : > "$chain" ;; esac
if [ -f fail ] && [ "$chain" = FORWARD ]; then echo injected-error >&2; exit 2; fi
if [ -f fail-before ] && [ "$(cat fail-before)" = "$op" ]; then rm fail-before; exit 2; fi
case "$op" in
-N) [ ! -e "$chain" ] || exit 1; : > "$chain" ;;
-A|-I) case " $* " in *" -m comment "*) if [ -f no-comment ]; then echo comment-unavailable >&2; exit 2; fi ;; esac
    [ -f "$chain" ] || exit 1
    echo "-A $chain $*" >> "$chain" ;;
-F) [ -f "$chain" ] || exit 1; : > "$chain" ;;
-X) [ -f "$chain" ] && [ ! -s "$chain" ] || exit 1
    rm "$chain" ;;
-S) cat "$chain" || exit 1 ;;
-D)
    awk -v rule="-A $chain $*" '$0 == rule && !removed { removed=1; next } { print } END { if (!removed) exit 1 }' "$chain" > next || exit 1
    mv next "$chain" ;;
*) exit 2 ;;
esac
if [ -f fail-after ] && [ "$(cat fail-after)" = "$op" ]; then rm fail-after; exit 2; fi
`
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	return &IPTables{Binary: binary}, dir
}

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
	table, dir := failurePolicyTestTable(t)
	binary := table.Binary
	count := func(chain string) int {
		b, err := os.ReadFile(filepath.Join(dir, chain))
		if err != nil {
			t.Fatal(err)
		}
		return strings.Count(string(b), "-j "+failurePolicyChain)
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
		legacy := "-A " + chain + " -s 192.0.2.1 -j DROP\n" + strings.ReplaceAll(legacyRule, " ", "  ") + "\n"
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
}

// Default fail-open must still work on kernels without the comment match.
func TestFailurePolicyWithoutCommentSupport(t *testing.T) {
	table, dir := failurePolicyTestTable(t)
	if err := os.WriteFile(filepath.Join(dir, "no-comment"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := table.InitializeFailurePolicy(false); err != nil {
		t.Fatal(err)
	}
	if err := table.SetFailurePolicyOpen(true); err != nil {
		t.Fatal(err)
	}
	for _, chain := range []string{"INPUT", "FORWARD"} {
		got, err := os.ReadFile(filepath.Join(dir, chain))
		if err != nil || string(got) != "-A "+chain+" -j "+failurePolicyChain+"\n" {
			t.Fatalf("fallback missing: %q, %v", got, err)
		}
	}
	// Crash while open, then select fail-closed: no manual cleanup required.
	restarted := &IPTables{Binary: table.Binary}
	if err := restarted.InitializeFailurePolicy(true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, failurePolicyChain)); !os.IsNotExist(err) {
		t.Fatalf("fallback chain remains: %v", err)
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

func TestFailurePolicyChainRecovery(t *testing.T) {
	for _, leftover := range []string{"empty", "populated", "input", "both", "duplicates"} {
		for _, closed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/closed=%v", leftover, closed), func(t *testing.T) {
				table, dir := failurePolicyTestTable(t)
				rule := "-A " + failurePolicyChain + " -j ACCEPT\n"
				files := map[string]string{failurePolicyChain: rule}
				if leftover == "empty" {
					files[failurePolicyChain] = ""
				}
				if leftover == "input" || leftover == "both" || leftover == "duplicates" {
					files["INPUT"] = "-A INPUT -j " + failurePolicyChain + "\n"
				}
				if leftover == "both" || leftover == "duplicates" {
					files["FORWARD"] = "-A FORWARD -j " + failurePolicyChain + "\n"
				}
				if leftover == "duplicates" {
					for key, value := range files {
						files[key] = value + value
					}
				}
				for file, content := range files {
					if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0600); err != nil {
						t.Fatal(err)
					}
				}
				if err := table.InitializeFailurePolicy(closed); err != nil {
					t.Fatal(err)
				}
				if _, err := os.Stat(filepath.Join(dir, failurePolicyChain)); !os.IsNotExist(err) {
					t.Fatalf("chain remains: %v", err)
				}
				for _, chain := range []string{"INPUT", "FORWARD"} {
					got, err := os.ReadFile(filepath.Join(dir, chain))
					if err != nil || len(got) != 0 {
						t.Fatalf("fallback remains: %q, %v", got, err)
					}
				}
			})
		}
	}
	for _, phase := range []string{"before", "after"} {
		for _, op := range []string{"-N", "-A", "-I", "-D", "-X"} {
			t.Run(phase+op, func(t *testing.T) {
				table, dir := failurePolicyTestTable(t)
				closing := op == "-D" || op == "-X"
				if closing {
					if err := table.SetFailurePolicyOpen(true); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.WriteFile(filepath.Join(dir, "fail-"+phase), []byte(op), 0600); err != nil {
					t.Fatal(err)
				}
				if err := table.SetFailurePolicyOpen(!closing); err == nil {
					t.Fatal("failure ignored")
				}
				if table.failurePolicyStateKnown {
					t.Fatal("failed update marked complete")
				}
				if err := table.SetFailurePolicyOpen(!closing); err != nil {
					t.Fatalf("retry failed: %v", err)
				}
				for i := 0; i < 3; i++ {
					restarted := &IPTables{Binary: table.Binary}
					if err := restarted.InitializeFailurePolicy(true); err != nil {
						t.Fatal(err)
					}
				}
				if _, err := os.Stat(filepath.Join(dir, failurePolicyChain)); !os.IsNotExist(err) {
					t.Fatalf("chain remains: %v", err)
				}
			})
		}
	}
}

func TestFailurePolicyPreservesUnfamiliarChains(t *testing.T) {
	for _, kind := range []string{"operator-rule", "conditional-reference", "other-chain", "goto", "similar-name"} {
		for _, open := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/open=%v", kind, open), func(t *testing.T) {
				table, dir := failurePolicyTestTable(t)
				name := failurePolicyChain
				rule := "-A " + name + " -j ACCEPT\n"
				input := ""
				switch kind {
				case "operator-rule":
					rule = "-A " + name + " -j DROP\n"
				case "conditional-reference":
					input = "-A INPUT -s 192.0.2.1 -j " + name + "\n"
				case "other-chain":
					input = "-A OUTPUT -j " + name + "\n"
				case "goto":
					input = "-A INPUT -g " + name + "\n"
				case "similar-name":
					name += "-OTHER"
					rule = strings.ReplaceAll(rule, failurePolicyChain, name)
				}
				for file, content := range map[string]string{name: rule, "INPUT": input} {
					if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0600); err != nil {
						t.Fatal(err)
					}
				}
				err := table.SetFailurePolicyOpen(open)
				if (err != nil) != (kind != "similar-name") {
					t.Fatalf("unexpected result: %v", err)
				}
				got, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil || string(got) != rule {
					t.Fatalf("changed %s: %q, %v", name, got, err)
				}
				if kind != "similar-name" {
					got, err = os.ReadFile(filepath.Join(dir, "INPUT"))
					if err != nil || string(got) != input {
						t.Fatalf("changed reference: %q, %v", got, err)
					}
				}
			})
		}
	}
}

// A comment mentioning the reserved chain is not a reference to that chain.
func TestFailurePolicyIgnoresOperatorComment(t *testing.T) {
	table, dir := failurePolicyTestTable(t)
	rule := "-A INPUT -m comment --comment \"do not use -j " + failurePolicyChain + " in operator rules\" -j DROP\n"
	if err := os.WriteFile(filepath.Join(dir, "INPUT"), []byte(rule), 0600); err != nil {
		t.Fatal(err)
	}
	if err := table.InitializeFailurePolicy(true); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "INPUT"))
	if err != nil || string(got) != rule {
		t.Fatalf("operator rule changed: %q, %v", got, err)
	}
}
