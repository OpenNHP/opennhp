package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/OpenNHP/opennhp/nhp/log"
)

type IPTYPE int32

const (
	DefaultSet     = "defaultset"
	DefaultSetV6   = "defaultset_v6"
	WhitelistSet   = "whitelistset"
	WhitelistSetV6 = "whitelistset_v6"
	BlacklistSet   = "blacklistset"
	BlacklistSetV6 = "blacklistset_v6"
	TempSet        = "tempset"
	TempSetV6      = "tempset_v6"
)

const (
	IPV6 IPTYPE = 1
	IPV4 IPTYPE = 2
)

const (
	POLICY_ACCEPT = iota
	POLICY_DROP
	POLICY_REJECT
)

type IPTables struct {
	Binary           string
	Binary6          string // ip6tables binary path
	IPv6Available    bool   // true if ip6tables is available
	InputPolicy      int    // IPv4 INPUT chain policy
	ForwardPolicy    int    // IPv4 FORWARD chain policy
	OutputPolicy     int    // IPv4 OUTPUT chain policy
	Input6Policy     int    // IPv6 INPUT chain policy
	Forward6Policy   int    // IPv6 FORWARD chain policy
	Output6Policy    int    // IPv6 OUTPUT chain policy
	AcceptInputMode  bool
	AcceptInput6Mode bool // separate tracking for IPv6

	failurePolicyMutex      sync.Mutex
	failurePolicyStateKnown bool
	failurePolicyTimeout    time.Duration
}

type IPTablesRuleOperation int32

const (
	ADD_IPTABLES_RULE_OPERATION = 0
	DEL_IPTABLES_RULE_OPERATION = 1
)

// Note: need root to run
func NewIPTables() (*IPTables, error) {
	path, err := exec.LookPath("iptables")
	if err != nil {
		log.Error("unable to locate iptables: %v", err)
		return nil, err
	}

	cmd := exec.Command(path, "-L")
	ret, err := cmd.Output()
	if err != nil {
		log.Error("execute command %s, error: %v\nYou need root privilege to run this appplication", cmd.String(), err)
		return nil, err
	}

	t := &IPTables{
		Binary: path,
	}

	// Detect ip6tables (optional - don't fail if not available)
	path6, err6 := exec.LookPath("ip6tables")
	if err6 == nil {
		cmd6 := exec.Command(path6, "-L")
		ret6, err6 := cmd6.Output()
		if err6 == nil {
			t.Binary6 = path6
			t.IPv6Available = true
			log.Info("ip6tables detected and available")

			// Parse IPv6 policies
			out6Strs := strings.Split(string(ret6), "\n")
			for _, line := range out6Strs {
				switch {
				case strings.HasPrefix(line, "Chain INPUT"):
					switch {
					case strings.Contains(line, "ACCEPT"):
						t.Input6Policy = POLICY_ACCEPT
					case strings.Contains(line, "DROP"):
						t.Input6Policy = POLICY_DROP
					case strings.Contains(line, "REJECT"):
						t.Input6Policy = POLICY_REJECT
					}

				case strings.HasPrefix(line, "Chain FORWARD"):
					switch {
					case strings.Contains(line, "ACCEPT"):
						t.Forward6Policy = POLICY_ACCEPT
					case strings.Contains(line, "DROP"):
						t.Forward6Policy = POLICY_DROP
					case strings.Contains(line, "REJECT"):
						t.Forward6Policy = POLICY_REJECT
					}

				case strings.HasPrefix(line, "Chain OUTPUT"):
					switch {
					case strings.Contains(line, "ACCEPT"):
						t.Output6Policy = POLICY_ACCEPT
					case strings.Contains(line, "DROP"):
						t.Output6Policy = POLICY_DROP
					case strings.Contains(line, "REJECT"):
						t.Output6Policy = POLICY_REJECT
					}
				}
			}
		} else {
			log.Warning("ip6tables found but not functional: %v", err6)
		}
	} else {
		log.Warning("ip6tables not found - IPv6 firewall rules will be skipped")
	}

	outStrs := strings.Split(string(ret), "\n")
	for _, line := range outStrs {
		switch {
		case strings.HasPrefix(line, "Chain INPUT"):
			switch {
			case strings.Contains(line, "ACCEPT"):
				t.InputPolicy = POLICY_ACCEPT
			case strings.Contains(line, "DROP"):
				t.InputPolicy = POLICY_DROP
			case strings.Contains(line, "REJECT"):
				t.InputPolicy = POLICY_REJECT
			}

		case strings.HasPrefix(line, "Chain FORWARD"):
			switch {
			case strings.Contains(line, "ACCEPT"):
				t.ForwardPolicy = POLICY_ACCEPT
			case strings.Contains(line, "DROP"):
				t.ForwardPolicy = POLICY_DROP
			case strings.Contains(line, "REJECT"):
				t.ForwardPolicy = POLICY_REJECT
			}

		case strings.HasPrefix(line, "Chain OUTPUT"):
			switch {
			case strings.Contains(line, "ACCEPT"):
				t.OutputPolicy = POLICY_ACCEPT
			case strings.Contains(line, "DROP"):
				t.OutputPolicy = POLICY_DROP
			case strings.Contains(line, "REJECT"):
				t.OutputPolicy = POLICY_REJECT
			}
		}
	}

	return t, nil
}

func (table *IPTables) changePolicy(inputPolicy, forwardPolicy, outputPolicy *int) {
	operStr := "ACCEPT"

	if inputPolicy != nil {
		switch *inputPolicy {
		case POLICY_ACCEPT:
			operStr = "ACCEPT"
		case POLICY_DROP:
			operStr = "DROP"
		case POLICY_REJECT:
			operStr = "REJECT"
		}
		cmd := exec.Command(table.Binary, "-P", "INPUT", operStr)
		err := cmd.Run()
		if err != nil {
			log.Debug("execute command %s, error: %v", cmd.String(), err)
		}
	}

	if forwardPolicy != nil {
		switch *forwardPolicy {
		case POLICY_ACCEPT:
			operStr = "ACCEPT"
		case POLICY_DROP:
			operStr = "DROP"
		case POLICY_REJECT:
			operStr = "REJECT"
		}
		cmd := exec.Command(table.Binary, "-P", "FORWARD", operStr)
		err := cmd.Run()
		if err != nil {
			log.Debug("execute command %s, error: %v", cmd.String(), err)
		}
	}

	if outputPolicy != nil {
		switch *outputPolicy {
		case POLICY_ACCEPT:
			operStr = "ACCEPT"
		case POLICY_DROP:
			operStr = "DROP"
		case POLICY_REJECT:
			operStr = "REJECT"
		}
		cmd := exec.Command(table.Binary, "-P", "OUTPUT", operStr)
		err := cmd.Run()
		if err != nil {
			log.Debug("execute command %s, error: %v", cmd.String(), err)
		}
	}
}

func (table *IPTables) changeIptablesRule(inputPolicy, forwardPolicy, outputPolicy *int, operator IPTablesRuleOperation, dest string) {
	operStr := "ACCEPT"
	operation := "-I"
	if operator == DEL_IPTABLES_RULE_OPERATION {
		operation = "-D"
	}

	if inputPolicy != nil {
		switch *inputPolicy {
		case POLICY_ACCEPT:
			operStr = "ACCEPT"
		case POLICY_DROP:
			operStr = "DROP"
		case POLICY_REJECT:
			operStr = "REJECT"
		}
		cmd := exec.Command(table.Binary, operation, "INPUT", "-d", dest, "-j", operStr)
		err := cmd.Run()
		if err != nil {
			log.Debug("execute command %s, error: %v", cmd.String(), err)
		}
	}

	if forwardPolicy != nil {
		switch *forwardPolicy {
		case POLICY_ACCEPT:
			operStr = "ACCEPT"
		case POLICY_DROP:
			operStr = "DROP"
		case POLICY_REJECT:
			operStr = "REJECT"
		}
		cmd := exec.Command(table.Binary, operation, "FORWARD", "-d", dest, "-j", operStr)
		err := cmd.Run()
		if err != nil {
			log.Debug("execute command %s, error: %v", cmd.String(), err)
		}
	}

	if outputPolicy != nil {
		switch *outputPolicy {
		case POLICY_ACCEPT:
			operStr = "ACCEPT"
		case POLICY_DROP:
			operStr = "DROP"
		case POLICY_REJECT:
			operStr = "REJECT"
		}
		cmd := exec.Command(table.Binary, operation, "OUTPUT", "-d", dest, "-j", operStr)
		err := cmd.Run()
		if err != nil {
			log.Debug("execute command %s, error: %v", cmd.String(), err)
		}
	}
}

// failurePolicyComment identifies only the IPv4 fallback rules owned by the AC.
// Older versions used untagged rules, which cannot safely be claimed as ours.
const failurePolicyComment = "opennhp-ac-fail-open"

// Older iptables supports bare -w but not a numeric wait. Bound the whole
// command so lock contention cannot stall startup or shutdown indefinitely.
func (table *IPTables) failurePolicyCommand(args ...string) ([]byte, error) {
	timeout := table.failurePolicyTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return exec.CommandContext(ctx, table.Binary, append([]string{"-w"}, args...)...).CombinedOutput()
}

// This exact name is reserved for the AC's unreferenced capability probe. Reuse
// bounds crash leftovers to one chain; unfamiliar contents are never flushed.
const failurePolicyProbeChain = "NHP-AC-COMMENT-PROBE"

// probeFailurePolicyComment checks kernel support without changing packet handling.
func (table *IPTables) probeFailurePolicyComment() (err error) {
	chain := failurePolicyProbeChain
	rule := []string{"-m", "comment", "--comment", failurePolicyComment, "-j", "RETURN"}
	present := false
	if out, cmdErr := table.failurePolicyCommand("-N", chain); cmdErr != nil {
		// A previous process may have died after creating or populating the probe.
		// Inspect the whole table to reject references as well as unfamiliar rules.
		rules, inspectErr := table.failurePolicyCommand("-S")
		if inspectErr != nil {
			return fmt.Errorf("create failure-policy probe: %w: %s; inspect: %v: %s", cmdErr, out, inspectErr, rules)
		}
		found := false
		for _, line := range strings.Split(string(rules), "\n") {
			fields := strings.Fields(line)
			if strings.Join(fields, " ") == "-N "+chain {
				found = true
			}
			for i := 0; i+1 < len(fields); i++ {
				if (fields[i] == "-j" || fields[i] == "-g") && fields[i+1] == chain {
					return fmt.Errorf("probe chain %s is referenced; inspect it manually", chain)
				}
			}
			if len(fields) >= 2 && fields[0] == "-A" && fields[1] == chain {
				if present || !isFailurePolicyRuleTarget(line, chain, "RETURN") {
					return fmt.Errorf("probe chain %s has unfamiliar contents: %q; inspect it manually", chain, line)
				}
				present = true
			}
		}
		if !found {
			return fmt.Errorf("create failure-policy probe: %w: %s", cmdErr, out)
		}
	}
	defer func() {
		// Delete only our exact rule. Never flush a chain, even the reserved one.
		var deleteErr error
		if present {
			if out, cleanupErr := table.failurePolicyCommand(append([]string{"-D", chain}, rule...)...); cleanupErr != nil {
				deleteErr = fmt.Errorf("clean up probe %s: %w: %s", chain, cleanupErr, out)
			}
		}
		if out, cleanupErr := table.failurePolicyCommand("-X", chain); cleanupErr != nil {
			err = errors.Join(err, deleteErr, fmt.Errorf("clean up probe %s: %w: %s", chain, cleanupErr, out))
		} else if err == nil {
			err = deleteErr
		}
		// If insertion failed and -X succeeded, the chain is gone: a failed
		// -D for a rule that was never inserted is not a cleanup failure.
	}()
	if present {
		if out, cmdErr := table.failurePolicyCommand(append([]string{"-D", chain}, rule...)...); cmdErr != nil {
			return fmt.Errorf("remove stale probe rule: %w: %s", cmdErr, out)
		}
	}
	// Recheck support even when recovering an old probe. A command can apply
	// its rule before reporting failure or timing out, so always attempt deletion.
	present = true
	if out, cmdErr := table.failurePolicyCommand(append([]string{"-A", chain}, rule...)...); cmdErr != nil {
		return fmt.Errorf("fail-open requires iptables comment match support: %w: %s", cmdErr, out)
	}
	return nil
}

// Match the entire owned rule, allowing only whitespace and comment quoting
// differences. A substring match could claim an operator's restricted rule.
func isFailurePolicyRule(line, chain string) bool {
	return isFailurePolicyRuleTarget(line, chain, "ACCEPT")
}

func isFailurePolicyRuleTarget(line, chain, target string) bool {
	line = strings.Join(strings.Fields(line), " ")
	for _, comment := range []string{failurePolicyComment, "\"" + failurePolicyComment + "\"", "'" + failurePolicyComment + "'"} {
		if line == "-A "+chain+" -m comment --comment "+comment+" -j "+target {
			return true
		}
	}
	return false
}

// InitializeFailurePolicy removes fallback rules left by an interrupted AC.
// Legacy unconditional ACCEPT rules may belong to the operator: fail-closed
// refuses them, while fail-open warns and preserves them. Neither installs a baseline.
func (table *IPTables) InitializeFailurePolicy(failClosed bool) error {
	table.failurePolicyMutex.Lock()
	defer table.failurePolicyMutex.Unlock()
	if err := table.setFailurePolicyRules(false); err != nil {
		return err
	}
	for _, chain := range []string{"INPUT", "FORWARD"} {
		out, err := table.failurePolicyCommand("-S", chain)
		if err != nil {
			return fmt.Errorf("inspect %s: %w: %s", chain, err, out)
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.Join(strings.Fields(line), " ")
			if line == "-A "+chain+" -j ACCEPT" || line == "-A "+chain+" -d 0.0.0.0/0 -j ACCEPT" {
				if failClosed {
					return fmt.Errorf("fail-closed: untagged unconditional ACCEPT in %s: %q; inspect and remove legacy fallback rules before restarting (existing firewall is still in effect)", chain, line)
				}
				log.Warning("untagged unconditional ACCEPT in %s: %q; this may be an old AC fallback or an operator rule; leaving it unchanged, so traffic it permits remains allowed after discovery recovers; inspect before upgrading", chain, line)
			}
		}
	}
	if !failClosed {
		if err := table.probeFailurePolicyComment(); err != nil {
			table.failurePolicyStateKnown = false
			return err
		}
	}
	return nil
}

// setFailurePolicyRules reconciles tagged rules from the kernel, including
// duplicates and partial updates. On error the next discovery tick retries.
// The caller holds failurePolicyMutex; one AC owns each network namespace.
func (table *IPTables) setFailurePolicyRules(open bool) error {
	table.failurePolicyStateKnown = false
	for _, chain := range []string{"INPUT", "FORWARD"} {
		out, err := table.failurePolicyCommand("-S", chain)
		if err != nil {
			return fmt.Errorf("inspect %s: %w: %s", chain, err, out)
		}
		rule := []string{"-m", "comment", "--comment", failurePolicyComment, "-j", "ACCEPT"}
		count := 0
		for _, line := range strings.Split(string(out), "\n") {
			if isFailurePolicyRule(line, chain) {
				count++
			}
		}
		wanted := 0
		if open {
			wanted = 1
		}
		for count > wanted {
			args := append([]string{"-D", chain}, rule...)
			if out, err := table.failurePolicyCommand(args...); err != nil {
				return fmt.Errorf("remove AC fallback in %s: %w: %s", chain, err, out)
			}
			count--
		}
		if count < wanted {
			args := append([]string{"-I", chain}, rule...)
			if out, err := table.failurePolicyCommand(args...); err != nil {
				return fmt.Errorf("insert AC fallback in %s: %w: %s", chain, err, out)
			}
		}
	}
	table.AcceptInputMode = open
	table.failurePolicyStateKnown = true
	return nil
}

// ResetAllInput removes the AC's IPv4 fallback. Kept as func() for API compatibility.
// Call SetFailurePolicyOpen when the caller needs to handle an error.
func (table *IPTables) ResetAllInput() {
	if err := table.SetFailurePolicyOpen(false); err != nil {
		log.Error("reset AC failure policy: %v", err)
	}
}

// AcceptAllInput installs the AC's IPv4 fallback. Kept as func() for API compatibility.
// Call SetFailurePolicyOpen when the caller needs to handle an error.
func (table *IPTables) AcceptAllInput() {
	if err := table.SetFailurePolicyOpen(true); err != nil {
		log.Error("open AC failure policy: %v", err)
	}
}

// SetFailurePolicyOpen reconciles owned IPv4 fallback rules, leaving grants intact.
// A failed operation remains retryable, including through the legacy wrappers.
func (table *IPTables) SetFailurePolicyOpen(open bool) error {
	table.failurePolicyMutex.Lock()
	defer table.failurePolicyMutex.Unlock()
	if table.failurePolicyStateKnown && table.AcceptInputMode == open {
		return nil
	}
	return table.setFailurePolicyRules(open)
}

// ResetAllInput6 resets ip6tables input rules (IPv6)
func (table *IPTables) ResetAllInput6() {
	if !table.IPv6Available || !table.AcceptInput6Mode {
		return
	}
	log.Info("reset ip6tables input")
	inputPolicy := POLICY_ACCEPT
	forwardPolicy := POLICY_ACCEPT
	table.changeIp6tablesRule(&inputPolicy, &forwardPolicy, nil, DEL_IPTABLES_RULE_OPERATION, "::/0")
	table.AcceptInput6Mode = false
}

// AcceptAllInput6 accepts all ip6tables input (IPv6)
func (table *IPTables) AcceptAllInput6() {
	if !table.IPv6Available || table.AcceptInput6Mode {
		return
	}
	log.Info("accept ip6tables input")
	inputPolicy := POLICY_ACCEPT
	forwardPolicy := POLICY_ACCEPT
	table.changeIp6tablesRule(&inputPolicy, &forwardPolicy, nil, ADD_IPTABLES_RULE_OPERATION, "::/0")
	table.AcceptInput6Mode = true
}

// changeIp6tablesRule applies rules to ip6tables (IPv6 version of changeIptablesRule)
func (table *IPTables) changeIp6tablesRule(inputPolicy, forwardPolicy, outputPolicy *int, operator IPTablesRuleOperation, dest string) {
	if !table.IPv6Available {
		return
	}

	operStr := "ACCEPT"
	operation := "-I"
	if operator == DEL_IPTABLES_RULE_OPERATION {
		operation = "-D"
	}

	if inputPolicy != nil {
		switch *inputPolicy {
		case POLICY_ACCEPT:
			operStr = "ACCEPT"
		case POLICY_DROP:
			operStr = "DROP"
		case POLICY_REJECT:
			operStr = "REJECT"
		}
		cmd := exec.Command(table.Binary6, operation, "INPUT", "-d", dest, "-j", operStr)
		err := cmd.Run()
		if err != nil {
			log.Debug("execute command %s, error: %v", cmd.String(), err)
		}
	}

	if forwardPolicy != nil {
		switch *forwardPolicy {
		case POLICY_ACCEPT:
			operStr = "ACCEPT"
		case POLICY_DROP:
			operStr = "DROP"
		case POLICY_REJECT:
			operStr = "REJECT"
		}
		cmd := exec.Command(table.Binary6, operation, "FORWARD", "-d", dest, "-j", operStr)
		err := cmd.Run()
		if err != nil {
			log.Debug("execute command %s, error: %v", cmd.String(), err)
		}
	}

	if outputPolicy != nil {
		switch *outputPolicy {
		case POLICY_ACCEPT:
			operStr = "ACCEPT"
		case POLICY_DROP:
			operStr = "DROP"
		case POLICY_REJECT:
			operStr = "REJECT"
		}
		cmd := exec.Command(table.Binary6, operation, "OUTPUT", "-d", dest, "-j", operStr)
		err := cmd.Run()
		if err != nil {
			log.Debug("execute command %s, error: %v", cmd.String(), err)
		}
	}
}

type IPSet struct {
	Binary string // Binary is Command path
	Wait   bool
}

// NewIPSet new ipset
// @Return: *IPSet, error
// Note: need root to run
func NewIPSet(wait bool) (*IPSet, error) {
	path, err := exec.LookPath("ipset")
	if err != nil {
		log.Error("unable to locate ipset: %v", err)
		return nil, err
	}

	return &IPSet{
		Binary: path,
		Wait:   wait,
	}, nil
}

func (ipset *IPSet) Add(ipType IPTYPE, t int, expire int, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
	defer cancel()

	name := ipset.GetIpsetName(ipType, t)
	params := append([]string{"add", "-exist", name}, args...)
	params = append(params, "timeout", fmt.Sprintf("%d", expire))

	log.Debug("Execute ipset: ipset %s", params)
	return ipset.Run(ctx, params...)
}

func (ipset *IPSet) GetIpsetName(ipType IPTYPE, t int) string {
	var name string
	switch t {
	case 1:
		if ipType == IPV4 {
			return DefaultSet
		} else if ipType == IPV6 {
			return DefaultSetV6
		}
		return DefaultSet
	case 2: // Blacklist and Whitelist
	case 3:
		return ipset.GetDefaultName(ipType, t)
	case 4:
		if ipType == IPV4 {
			return TempSet
		} else if ipType == IPV6 {
			return TempSetV6
		}
		return TempSet
	}
	return name
}

func (ipset *IPSet) GetDefaultName(ipType IPTYPE, t int) string {
	var name string
	switch t {
	case 2: // Whitelist
		name = WhitelistSet
		if ipType == IPV6 {
			name = WhitelistSetV6
		}
	case 3: // Blacklist
		name = BlacklistSet
		if ipType == IPV6 {
			name = BlacklistSetV6
		}
	}
	return name
}

func (ipset *IPSet) Run(ctx context.Context, args ...string) (string, error) {
	c := make(chan string)
	defer close(c)
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, ipset.Binary, args...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	if stderr.String() != "" {
		return "", fmt.Errorf("%s", stderr.String())
	}
	return stdout.String(), nil
}
