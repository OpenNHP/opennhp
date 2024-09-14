package utils

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/OpenNHP/opennhp/log"
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
	Binary          string
	InputPolicy     int
	ForwardPolicy   int
	OutputPolicy    int
	AcceptInputMode bool
}

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

func (table *IPTables) ResetAllInput() {
	if !table.AcceptInputMode {
		return
	}
	log.Info("reset iptables input")
	inputPolicy := table.InputPolicy
	table.changePolicy(&inputPolicy, nil, nil)
	table.AcceptInputMode = false
}

func (table *IPTables) AcceptAllInput() {
	if table.AcceptInputMode {
		return
	}
	log.Info("accept iptables input ")
	inputPolicy := POLICY_ACCEPT
	table.changePolicy(&inputPolicy, nil, nil)
	table.AcceptInputMode = true
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
		if ipType == IPV4 {
			name = WhitelistSet
		} else if ipType == IPV6 {
			name = WhitelistSetV6
		}
		name = WhitelistSet
	case 3: // Blacklist
		if ipType == IPV4 {
			name = BlacklistSet
		} else if ipType == IPV6 {
			name = BlacklistSetV6
		}
		name = BlacklistSet
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
