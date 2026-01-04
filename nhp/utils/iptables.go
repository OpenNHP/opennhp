package utils

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
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
	InputPolicy      int
	ForwardPolicy    int
	OutputPolicy     int
	AcceptInputMode  bool
	AcceptInput6Mode bool // separate tracking for IPv6
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
		if err6 = cmd6.Run(); err6 == nil {
			t.Binary6 = path6
			t.IPv6Available = true
			log.Info("ip6tables detected and available")
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

func (table *IPTables) ResetAllInput() {
	if !table.AcceptInputMode {
		return
	}
	log.Info("reset iptables input v2")
	inputPolicy := POLICY_ACCEPT
	forwardPolicy := POLICY_ACCEPT
	//table.changePolicy(&inputPolicy, nil, nil)
	table.changeIptablesRule(&inputPolicy, &forwardPolicy, nil, DEL_IPTABLES_RULE_OPERATION, "0.0.0.0/0")
	table.AcceptInputMode = false
}

func (table *IPTables) AcceptAllInput() {
	if table.AcceptInputMode {
		return
	}
	log.Info("accept iptables input v2")
	inputPolicy := POLICY_ACCEPT
	forwardPolicy := POLICY_ACCEPT
	//table.changePolicy(&inputPolicy, nil, nil)
	table.changeIptablesRule(&inputPolicy, &forwardPolicy, nil, ADD_IPTABLES_RULE_OPERATION, "0.0.0.0/0")
	table.AcceptInputMode = true
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
		return "", fmt.Errorf(stderr.String())
	}
	return stdout.String(), nil
}
