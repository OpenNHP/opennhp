package main

import (
	"bytes"
	"net/netip"
	"os"
	"regexp"
	"strings"
	"testing"
)

func runReport(t *testing.T, args ...string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := run(args, &buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func build(t *testing.T, geo GeoResolver) *Report {
	t.Helper()
	an := NewAnalyzer("", "")
	for _, d := range []string{"2026-10-01", "2026-10-02"} {
		f, err := os.Open("testdata/nhp_server_xdp-" + d + ".log")
		if err != nil {
			t.Fatal(err)
		}
		if err := an.Consume(f, d); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	return an.Build(10, geo, false)
}

func TestTotalsDedupRestartBudget(t *testing.T) {
	r := build(t, nil)
	if r.DropPkts != 1000+500+200+50 { // duplicate STAT ignored
		t.Errorf("drop pkts = %d", r.DropPkts)
	}
	if r.PassPkts != 10 || r.TotalPkts != 1760 {
		t.Errorf("pass=%d total=%d", r.PassPkts, r.TotalPkts)
	}
	if r.DupStats != 1 || r.BadLines != 1 {
		t.Errorf("dup=%d bad=%d", r.DupStats, r.BadLines)
	}
	if r.ScanLines != 1 {
		t.Errorf("scan lines = %d", r.ScanLines)
	}
	if len(r.Restarts) != 1 { // TOTAL 1200 -> 50 on day 2
		t.Errorf("restarts = %v", r.Restarts)
	}
	if r.BudgetExhausted["2026-10-01"] != "10:02:00" {
		t.Errorf("budget = %v", r.BudgetExhausted)
	}
	if len(r.UnknownRsns) != 1 || r.UnknownRsns[0] != "UNKNOWN-99" {
		t.Errorf("unknown = %v", r.UnknownRsns)
	}
	for _, c := range r.Reasons {
		if c.Reason == "UDP_SHORT" && !c.NoDetail {
			t.Error("UDP_SHORT should be flagged totals-only")
		}
	}
}

// A reason that was quiet across a restart must not be reported as a second
// restart when its first post-restart STAT line arrives.
func TestRestartResetsQuietCounters(t *testing.T) {
	an := NewAnalyzer("", "")
	for _, tc := range []struct{ day, text string }{
		{"2026-10-01", "10:00:00 s [NHP-STAT] VERDICT=DROP REASON=TCP_OTHER WINDOW=60s PKTS=5 LOGGED=1 TOTAL=5\n" +
			"10:00:00 s [NHP-STAT] VERDICT=DROP REASON=UDP_OTHER WINDOW=60s PKTS=9 LOGGED=1 TOTAL=9\n"},
		{"2026-10-02", "10:00:00 s [NHP-STAT] VERDICT=DROP REASON=UDP_OTHER WINDOW=60s PKTS=1 LOGGED=1 TOTAL=1\n" +
			"10:01:00 s [NHP-STAT] VERDICT=DROP REASON=TCP_OTHER WINDOW=60s PKTS=1 LOGGED=1 TOTAL=1\n"},
	} {
		if err := an.Consume(strings.NewReader(tc.text), tc.day); err != nil {
			t.Fatal(err)
		}
	}
	if r := an.Build(10, nil, false); len(r.Restarts) != 1 {
		t.Errorf("restarts = %v", r.Restarts)
	}
}

func TestRankingGeoAndReserved(t *testing.T) {
	g, err := loadGeoCSV("testdata/geo.csv")
	if err != nil {
		t.Fatal(err)
	}
	r := build(t, g)
	if r.TopIPs[0].IP != "203.0.113.5" || r.TopIPs[0].Events != 3 || r.TopIPs[0].Categories != 2 {
		t.Errorf("top ip = %+v", r.TopIPs[0])
	}
	if r.TopIPs[0].Country != "AU" || r.TopIPs[0].City != "Sydney" {
		t.Errorf("geo = %+v", r.TopIPs[0])
	}
	if r.ReservedIPs != 1 {
		t.Errorf("reserved = %d", r.ReservedIPs)
	}
	for _, x := range r.TopGeo {
		if strings.Contains(x.Country, "Private") {
			t.Error("private source leaked into geo ranking")
		}
	}
	if r.TopGeo[0].Country != "AU" {
		t.Errorf("top geo = %+v", r.TopGeo)
	}
	if _, ok := g.Lookup(netip.MustParseAddr("2001:db8::1")); !ok {
		t.Error("ipv6 lookup failed")
	}
}

func TestParseLineIPv6AndErrors(t *testing.T) {
	rec, err := parseLine("10:00:04 srv1 [NHP-DROP] REASON=V6_OTHER SRC=2001:db8::1 DST=x LEN=1 PROTO=TCP SPT=1 DPT=80 RELAY=0")
	if err != nil || !rec.Src.Is6() || rec.DstPort != 80 {
		t.Fatalf("%+v %v", rec, err)
	}
	for _, bad := range []string{"", "hello world", "10:00:00 s [NHP-DROP] REASON=X SRC=nope", "10:00:00 s [NHP-STAT] VERDICT=DROP REASON=X PKTS=a"} {
		if _, err := parseLine(bad); err == nil {
			t.Errorf("expected error for %q", bad)
		}
	}
}

func TestCLIFormatsAndMask(t *testing.T) {
	args := []string{"-in", "testdata", "-top", "3"}
	md := runReport(t, append(args, "-format", "md", "-mask-ip")...)
	if !strings.Contains(md, "203.0.113.x") || strings.Contains(md, "203.0.113.5") {
		t.Error("mask-ip not applied")
	}
	if !strings.Contains(md, "WARNING") || !strings.Contains(md, "Not enabled") {
		t.Error("missing budget warning or geo-disabled note")
	}
	if js := runReport(t, append(args, "-format", "json")...); !strings.Contains(js, `"schema": 1`) {
		t.Error("json schema missing")
	}
	if csv := runReport(t, append(args, "-format", "csv")...); !strings.HasPrefix(csv, "section,key") {
		t.Error("csv header missing")
	}
	page := runReport(t, append(args, "-format", "html")...)
	if !strings.HasPrefix(page, "<!doctype html>") || !strings.Contains(page, "Daily Attack Report: 2026-10-01 – 2026-10-02") {
		t.Error("html page missing")
	}
	// The headline sentence is built from the same counters as the Overview
	// cards, so a drift between them (or a formatter that stopped grouping)
	// shows up as a mismatch with those cards rather than as a silent prose
	// number. UniqueIPs is the "Unique attack source IPs" card: private and
	// reserved sources are excluded there, so the fixture's 10.0.0.9 does
	// not count.
	if !strings.Contains(page, "1,750 hostile packets from 2 sources") {
		t.Error("detail subtitle does not carry the grouped totals")
	}
	if !strings.Contains(page, "That&rsquo;s 99.43% of all inbound traffic") {
		t.Error("detail subtitle does not carry the DROP share")
	}
	if !strings.Contains(page, "WARNING") || !strings.Contains(page, "Not enabled") {
		t.Error("html warning or geo-disabled note missing")
	}
	if !strings.Contains(page, "203.0.113.5") {
		t.Error("html should not mask by default")
	}
}

// The classifier must know every DROP reason the checked-in XDP side emits.
// It may know more: a deployed build can lead this branch.
func TestClassifierMatchesEngine(t *testing.T) {
	src, err := os.ReadFile("../../nhp/utils/ebpf/engine_linux.go")
	if err != nil {
		t.Skip("engine source not available")
	}
	re := regexp.MustCompile(`return "DROP", "([A-Z0-9_]+)"`)
	m := re.FindAllStringSubmatch(string(src), -1)
	if len(m) == 0 {
		t.Fatal("no DROP reasons found in engine_linux.go")
	}
	for _, x := range m {
		if _, ok := dropCategories[x[1]]; !ok {
			t.Errorf("reason %s missing from dropCategories", x[1])
		}
	}
	for _, r := range []string{"TCP_SYN_SCAN", "TCP_STEALTH_SCAN"} {
		if _, ok := dropCategories[r]; !ok {
			t.Errorf("%s missing from dropCategories", r)
		}
	}
}
