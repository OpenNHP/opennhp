package main

import (
	"fmt"
	"io"
	"strings"
)

func pc(f float64) string { return fmt.Sprintf("%.2f%%", f*100) }

func renderMarkdown(w io.Writer, r *Report) error {
	var b strings.Builder
	p := func(f string, a ...any) { fmt.Fprintf(&b, f, a...) }

	p("# NHP Server XDP Attack Report\n\n")
	if len(r.BudgetExhausted) > 0 {
		p("> **WARNING:** the per-packet event budget was exhausted (")
		var days []string
		for _, d := range sortedKeys(r.BudgetExhausted) {
			days = append(days, d+" from "+r.BudgetExhausted[d])
		}
		p("%s). Per-packet lines after that point are missing, so source/port/geo rankings are incomplete. Exact totals are unaffected.\n\n", strings.Join(days, ", "))
	}

	p("## Methodology\n\n")
	p("- Attack = `[NHP-DROP]`; `PASS` is allowed traffic, shown only as a baseline.\n")
	p("- **Exact** packet totals come from the per-minute `[NHP-STAT]` lines (PKTS summed, duplicates removed).\n")
	p("- **Sampled** counts come from per-packet lines, which the filter rate-limits per class; bulk classes may have none. Rankings (IP, geo, port) use sampled events and are not comparable in volume across classes.\n")
	p("- The log has no date per line; dates come from file names.\n\n")

	p("## Overview\n\n")
	p("| Item | Value |\n|---|---|\n")
	p("| Server ID | %s |\n", strings.Join(r.Servers, ", "))
	p("| Time range | %s – %s |\n", r.From, r.To)
	p("| Total packets (exact) | %d |\n", r.TotalPkts)
	p("| Attack (DROP) packets (exact) | %d (%s) |\n", r.DropPkts, pc(r.DropShare))
	p("| Allowed (PASS) packets (exact) | %d |\n", r.PassPkts)
	p("| Sampled DROP events | %d |\n", r.DropEvents)
	p("| Sampling rate (events / exact DROP) | %s |\n", pc(r.SamplingRate))
	p("| Unique attack source IPs (sampled) | %d |\n", r.UniqueIPs)
	p("| Private/Reserved sources (sampled) | %d |\n", r.ReservedIPs)
	if r.PeakHour != "" {
		p("| Peak hour | %s |\n", r.PeakHour)
	}
	p("\n## Attack types\n\n")
	p("| Category | Packets (exact) | Share | Sampled events |\n|---|---:|---:|---:|\n")
	for _, c := range r.Categories {
		p("| %s | %d | %s | %d |\n", c.Category, c.Pkts, pc(c.Share), c.Events)
	}
	p("\n| Reason | Category | Packets (exact) | Share | Sampled events | Note |\n|---|---|---:|---:|---:|---|\n")
	for _, x := range r.Reasons {
		note := ""
		if x.NoDetail {
			note = "totals only, no source detail"
		}
		p("| %s | %s | %d | %s | %d | %s |\n", x.Reason, x.Category, x.Pkts, pc(x.PktShare), x.Events, note)
	}

	p("\n## Top %d attack sources (sampled)\n\n", r.TopN)
	p("| # | IP | Events | Share | Categories | First seen | Last seen | Top dst ports | Location |\n|---|---|---:|---:|---:|---|---|---|---|\n")
	for i, x := range r.TopIPs {
		loc := strings.TrimSpace(x.Country + " " + x.City)
		p("| %d | %s | %d | %s | %d | %s | %s | %v | %s |\n", i+1, x.IP, x.Events, pc(x.Share), x.Categories, x.First, x.Last, x.TopPorts, loc)
	}

	p("\n## Top %d regions (sampled)\n\n", r.TopN)
	if !r.GeoEnabled {
		p("Not enabled (no `-geo-csv` given). No lookups are performed online.\n")
	} else {
		p("| # | Country | Events | Unique IPs | Share |\n|---|---|---:|---:|---:|\n")
		for i, x := range r.TopGeo {
			p("| %d | %s | %d | %d | %s |\n", i+1, x.Country, x.Events, x.IPs, pc(x.Share))
		}
	}

	p("\n## Top destination ports (sampled)\n\n| Port | Events | Share |\n|---|---:|---:|\n")
	for _, x := range r.TopPorts {
		p("| %s | %d | %s |\n", x.Key, x.Count, pc(x.Share))
	}
	p("\n## Protocols (sampled)\n\n| Protocol | Events | Share |\n|---|---:|---:|\n")
	for _, x := range r.Protos {
		p("| %s | %d | %s |\n", x.Key, x.Count, pc(x.Share))
	}

	p("\n## Trend (exact DROP packets)\n\n| Day | Packets | Share |\n|---|---:|---:|\n")
	for _, x := range r.Daily {
		p("| %s | %d | %s |\n", x.Key, x.Count, pc(x.Share))
	}
	p("\n| Hour | Packets |\n|---|---:|\n")
	for _, x := range r.Hourly {
		p("| %s | %d |\n", x.Key, x.Count)
	}

	p("\n## Data quality\n\n")
	p("- Unparsable lines: %d\n- Duplicate STAT lines ignored: %d\n", r.BadLines, r.DupStats)
	p("- Scan-detector summary lines (parsed, not part of DROP totals): %d\n", r.ScanLines)
	p("- IPv6 share of sampled DROP events: %s\n", pc(r.IPv6Share))
	if len(r.UnknownRsns) > 0 {
		p("- Unknown reasons (classifier out of date?): %s\n", strings.Join(r.UnknownRsns, ", "))
	}
	if len(r.Restarts) == 0 {
		p("- Restarts detected: none\n")
	} else {
		p("- Restarts detected (STAT TOTAL went backwards): %d\n", len(r.Restarts))
		for _, s := range r.Restarts {
			p("  - %s\n", s)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}
