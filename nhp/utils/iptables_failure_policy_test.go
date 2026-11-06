package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A fake executable exercises command errors and partial reconciliation without
// requiring root or ever invoking the host firewall. Real enforcement is tested
// separately in the disposable integration lab.
func TestFailurePolicyRules(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "iptables")
	script := `#!/bin/sh
cd "$(dirname "$0")" || exit 2
shift 2
op=$1; chain=$2
shift 2
[ -e "$chain" ] || : > "$chain"
if [ -f fail ] && [ "$chain" = FORWARD ]; then echo injected-error >&2; exit 2; fi
case "$op" in
-S) cat "$chain" ;;
-I) echo "-A $chain $*" >> "$chain" ;;
-D) sed '1d' "$chain" > next; mv next "$chain" ;;
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
	if err := table.AcceptAllInput(); err != nil {
		t.Fatal(err)
	}
	if err := table.AcceptAllInput(); err != nil {
		t.Fatal(err)
	}
	if count("INPUT") != 1 || count("FORWARD") != 1 {
		t.Fatal("duplicate or missing rule")
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
	if restarted.failurePolicyInitialized {
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
	if err := restarted.AcceptAllInput(); err == nil {
		t.Fatal("insert failure ignored")
	}
	if err := os.Remove(filepath.Join(dir, "fail")); err != nil {
		t.Fatal(err)
	}
	if err := restarted.AcceptAllInput(); err != nil {
		t.Fatal(err)
	}
	if count("INPUT") != 1 || count("FORWARD") != 1 {
		t.Fatal("retry duplicated rule")
	}
	if err := restarted.ResetAllInput(); err != nil {
		t.Fatal(err)
	}
	for _, chain := range []string{"INPUT", "FORWARD"} {
		legacy := "-A " + chain + " -j ACCEPT\n"
		if err := os.WriteFile(filepath.Join(dir, chain), []byte(legacy), 0600); err != nil {
			t.Fatal(err)
		}
		if err := restarted.InitializeFailurePolicy(true); err == nil || !strings.Contains(err.Error(), "untagged") {
			t.Fatalf("legacy rule accepted: %v", err)
		}
		b, _ := os.ReadFile(filepath.Join(dir, chain))
		if string(b) != legacy {
			t.Fatal("operator rule deleted")
		}
		if err := restarted.InitializeFailurePolicy(false); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, chain), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
