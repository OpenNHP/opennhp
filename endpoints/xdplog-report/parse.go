package main

import (
	"fmt"
	"net/netip"
	"regexp"
	"strconv"
	"strings"
)

// Line kinds produced by parseLine.
const (
	kindEvent = iota
	kindStat
	kindBudget
	kindScan
)

// Record is one parsed log line. Fields not relevant to the kind stay zero.
type Record struct {
	Kind     int
	Clock    string // HH:MM:SS (the log carries no date)
	ServerID string
	Verdict  string // PASS | DROP
	Reason   string

	// event
	Src     netip.Addr
	Proto   string
	DstPort int
	Relay   bool

	// stat
	Pkts, Logged, Total uint64

	// scan (NHP-SCAN, a scan-detector summary kind emitted by newer builds)
	ScanKind string
}

var dateInName = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)

// dateFromFilename extracts YYYY-MM-DD from nhp_server_xdp-YYYY-MM-DD.log.
func dateFromFilename(name string) (string, bool) {
	d := dateInName.FindString(name)
	return d, d != ""
}

func parseClock(s string) bool {
	if len(s) != 8 || s[2] != ':' || s[5] != ':' {
		return false
	}
	for i, c := range []byte(s) {
		if i == 2 || i == 5 {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// parseLine parses one log line: "HH:MM:SS <serverId> [NHP-xxx] k=v ...".
func parseLine(line string) (Record, error) {
	var r Record
	f := strings.Fields(line)
	if len(f) < 3 || !parseClock(f[0]) {
		return r, fmt.Errorf("not an xdp log line")
	}
	r.Clock, r.ServerID = f[0], f[1]
	tag := f[2]
	if !strings.HasPrefix(tag, "[NHP-") || !strings.HasSuffix(tag, "]") {
		return r, fmt.Errorf("missing [NHP-*] tag")
	}
	tag = tag[len("[NHP-") : len(tag)-1]

	kv := make(map[string]string, len(f))
	for _, tok := range f[3:] {
		if i := strings.IndexByte(tok, '='); i > 0 {
			kv[tok[:i]] = tok[i+1:]
		}
	}

	switch tag {
	case "PASS", "DROP":
		r.Kind, r.Verdict, r.Reason = kindEvent, tag, kv["REASON"]
		if r.Reason == "" {
			return r, fmt.Errorf("event without REASON")
		}
		addr, err := netip.ParseAddr(kv["SRC"])
		if err != nil {
			return r, fmt.Errorf("bad SRC: %w", err)
		}
		r.Src = addr.Unmap()
		r.Proto = kv["PROTO"]
		r.DstPort, _ = strconv.Atoi(kv["DPT"])
		r.Relay = kv["RELAY"] == "1"
	case "STAT":
		if len(f) > 3 && f[3] == "EVENT_LOG_BUDGET_EXHAUSTED" {
			r.Kind = kindBudget
			return r, nil
		}
		r.Kind, r.Verdict, r.Reason = kindStat, kv["VERDICT"], kv["REASON"]
		var e1, e2, e3 error
		r.Pkts, e1 = strconv.ParseUint(kv["PKTS"], 10, 64)
		r.Logged, e2 = strconv.ParseUint(kv["LOGGED"], 10, 64)
		r.Total, e3 = strconv.ParseUint(kv["TOTAL"], 10, 64)
		if r.Verdict == "" || r.Reason == "" || e1 != nil || e2 != nil || e3 != nil {
			return r, fmt.Errorf("malformed STAT line")
		}
	case "SCAN":
		// Periodic scan-detector summaries (KIND=TOTAL|SRC|PORT). They are
		// diagnostics, not DROP events, so they never reach the attack totals.
		r.Kind, r.ScanKind = kindScan, kv["KIND"]
		if r.ScanKind == "" {
			return r, fmt.Errorf("malformed SCAN line")
		}
	default:
		return r, fmt.Errorf("unknown tag %q", tag)
	}
	return r, nil
}
