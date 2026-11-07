package utils

import (
	"bytes"
	"context"
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

// The exact chain name is reserved for AC fallback ownership. It uses only
// built-in iptables features; the comment extension is not required.
const failurePolicyChain = "NHP-AC-FALLBACK"

// Recognize the exact comment-tagged rules used by earlier drafts, so an
// interrupted deployment of those versions can also be cleaned up safely.
func isFailurePolicyRule(line, chain string) bool {
	line = strings.Join(strings.Fields(line), " ")
	for _, comment := range []string{failurePolicyComment, "\"" + failurePolicyComment + "\"", "'" + failurePolicyComment + "'"} {
		if line == "-A "+chain+" -m comment --comment "+comment+" -j ACCEPT" {
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
	return nil
}

// setFailurePolicyRules reconciles owned rules from the kernel, including
// duplicates and partial updates. On error the next discovery tick retries.
// The caller holds failurePolicyMutex; one AC owns each network namespace.
func (table *IPTables) setFailurePolicyRules(open bool) error {
	table.failurePolicyStateKnown = false
	out, err := table.failurePolicyCommand("-S")
	if err != nil {
		return fmt.Errorf("inspect AC fallback: %w: %s", err, out)
	}
	present, accepts := false, 0
	jumps := map[string]int{"INPUT": 0, "FORWARD": 0}
	legacy := map[string]int{"INPUT": 0, "FORWARD": 0}
	// Validate the whole reserved chain and all references before changing it.
	// Never flush unknown contents or claim a conditional/operator reference.
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		rule := strings.Join(fields, " ")
		if rule == "-N "+failurePolicyChain {
			present = true
		}
		if len(fields) < 2 || fields[0] != "-A" {
			continue
		}
		chain := fields[1]
		if chain == failurePolicyChain {
			if rule != "-A "+failurePolicyChain+" -j ACCEPT" {
				return fmt.Errorf("reserved AC chain has unfamiliar rule %q; inspect before restarting", line)
			}
			accepts++
		}
		// -S renders a user-chain jump/goto last (there are no target options).
		// Checking that suffix avoids mistaking text inside a comment for a jump.
		if strings.HasSuffix(rule, " -j "+failurePolicyChain) || strings.HasSuffix(rule, " -g "+failurePolicyChain) {
			if _, owned := jumps[chain]; !owned || rule != "-A "+chain+" -j "+failurePolicyChain {
				return fmt.Errorf("reserved AC chain has unfamiliar reference %q; inspect before restarting", line)
			}
			jumps[chain]++
		}
		if _, owned := legacy[chain]; owned && isFailurePolicyRule(line, chain) {
			legacy[chain]++
		}
	}
	run := func(args ...string) error {
		if out, err := table.failurePolicyCommand(args...); err != nil {
			return fmt.Errorf("AC fallback %s: %w: %s", strings.Join(args, " "), err, out)
		}
		return nil
	}
	if open {
		if !present {
			if err := run("-N", failurePolicyChain); err != nil {
				return err
			}
		}
		if accepts == 0 {
			if err := run("-A", failurePolicyChain, "-j", "ACCEPT"); err != nil {
				return err
			}
			accepts = 1
		}
	}
	wanted := 0
	if open {
		wanted = 1
	}
	for _, chain := range []string{"INPUT", "FORWARD"} {
		for jumps[chain] > wanted {
			if err := run("-D", chain, "-j", failurePolicyChain); err != nil {
				return err
			}
			jumps[chain]--
		}
		if jumps[chain] < wanted {
			if err := run("-I", chain, "-j", failurePolicyChain); err != nil {
				return err
			}
		}
		for i := 0; i < legacy[chain]; i++ {
			if err := run("-D", chain, "-m", "comment", "--comment", failurePolicyComment, "-j", "ACCEPT"); err != nil {
				return err
			}
		}
	}
	for accepts > wanted {
		if err := run("-D", failurePolicyChain, "-j", "ACCEPT"); err != nil {
			return err
		}
		accepts--
	}
	if !open && present {
		if err := run("-X", failurePolicyChain); err != nil {
			return err
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
