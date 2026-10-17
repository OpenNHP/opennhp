package main

import (
	"fmt"
	"html/template"
	"io"
	"strings"
)

// renderHTML writes a standalone report page: no scripts, no external assets,
// so it can be attached to a ticket or opened offline.
func renderHTML(w io.Writer, r *Report) error {
	return htmlReport.Execute(w, r)
}

var htmlReport = template.Must(template.New("report").Funcs(template.FuncMap{
	"pct": pc,
	"bar": func(f float64) template.CSS {
		return template.CSS(fmt.Sprintf("width:%.1f%%", f*100))
	},
	"loc": func(x IPRow) string { return strings.TrimSpace(x.Country + " " + x.City) },
	"ports": func(p []int) string {
		var s []string
		for _, x := range p {
			s = append(s, fmt.Sprint(x))
		}
		return strings.Join(s, ", ")
	},
	"inc": func(i int) int { return i + 1 },
	"join": func(xs []string, sep string) string {
		return strings.Join(xs, sep)
	},
	"budgetDays": func(m map[string]string) []string {
		var out []string
		for _, d := range sortedKeys(m) {
			out = append(out, d+" from "+m[d])
		}
		return out
	},
}).Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>NHP Server XDP Attack Report</title>
<style>
:root{--fg:#1f2328;--muted:#656d76;--line:#d0d7de;--chip:#f6f8fa;--accent:#cf222e;--accent2:#0969da}
*{box-sizing:border-box}
body{margin:0;padding:32px 20px 64px;font:14px/1.55 -apple-system,BlinkMacSystemFont,"Segoe UI",Helvetica,Arial,sans-serif;color:var(--fg);background:var(--chip)}
main{max-width:1100px;margin:0 auto;background:#fff;border:1px solid var(--line);border-radius:12px;padding:28px 32px;box-shadow:0 1px 3px rgba(0,0,0,.05)}
h1{margin:0 0 4px;font-size:26px;letter-spacing:-.01em}
h2{margin:32px 0 10px;font-size:18px;border-bottom:1px solid var(--line);padding-bottom:6px}
header p{margin:0}
ul{margin:8px 0 16px;padding-left:22px}
code{font:12.5px/1.4 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;background:var(--chip);padding:1px 5px;border-radius:4px}
table{width:100%;border-collapse:collapse;margin:6px 0 22px;font-variant-numeric:tabular-nums}
th,td{padding:6px 8px;border-bottom:1px solid var(--line);text-align:left;vertical-align:top}
th{font-size:11.5px;text-transform:uppercase;letter-spacing:.05em;color:var(--muted);background:var(--chip);white-space:nowrap}
td.num,th.num{text-align:right}
.muted{color:var(--muted)}
.bar{display:inline-block;width:72px;height:6px;background:#eaeef2;border-radius:3px;overflow:hidden;margin-right:7px;vertical-align:middle}
.bar i{display:block;height:100%;background:var(--accent2)}
.bar.red i{background:var(--accent)}
.warn{border-left:4px solid var(--accent);background:#fff5f5;padding:12px 16px;border-radius:6px;margin:18px 0}
.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(170px,1fr));gap:12px;margin:14px 0 6px}
.card{border:1px solid var(--line);border-radius:8px;padding:11px 14px;background:var(--chip)}
.card .k{font-size:11.5px;color:var(--muted);text-transform:uppercase;letter-spacing:.05em}
.card .v{font-size:21px;font-weight:600;margin-top:2px}
footer{margin-top:34px;padding-top:14px;border-top:1px solid var(--line);color:var(--muted);font-size:12px}
@media print{body{background:#fff;padding:0}main{border:0;box-shadow:none;padding:0}}
</style>
</head>
<body>
<main>
<header>
<h1>NHP Server XDP Attack Report</h1>
<p class="muted">{{join .Servers ", "}} &middot; {{.From}} &ndash; {{.To}}</p>
</header>
{{if .BudgetExhausted}}
<div class="warn"><strong>WARNING:</strong> the per-packet event budget was exhausted ({{join (budgetDays .BudgetExhausted) ", "}}). Per-packet lines after that point are missing, so source/port/geo rankings are incomplete. Exact totals are unaffected.</div>
{{end}}

<h2>Methodology</h2>
<ul>
<li>Attack = <code>[NHP-DROP]</code>; <code>PASS</code> is allowed traffic, shown only as a baseline.</li>
<li><strong>Exact</strong> packet totals come from the per-minute <code>[NHP-STAT]</code> lines (PKTS summed, duplicates removed).</li>
<li><strong>Sampled</strong> counts come from per-packet lines, which the filter rate-limits per class; bulk classes may have none. Rankings (IP, geo, port) use sampled events and are not comparable in volume across classes.</li>
<li>The log has no date per line; dates come from file names.</li>
</ul>

<h2>Overview</h2>
<div class="cards">
<div class="card"><div class="k">Total packets (exact)</div><div class="v">{{.TotalPkts}}</div></div>
<div class="card"><div class="k">Attack (DROP) packets</div><div class="v">{{.DropPkts}} <span class="muted" style="font-size:13px;font-weight:400">{{pct .DropShare}}</span></div></div>
<div class="card"><div class="k">Allowed (PASS) packets</div><div class="v">{{.PassPkts}}</div></div>
<div class="card"><div class="k">Sampled DROP events</div><div class="v">{{.DropEvents}}</div></div>
<div class="card"><div class="k">Sampling rate</div><div class="v">{{pct .SamplingRate}}</div></div>
<div class="card"><div class="k">Unique attack source IPs</div><div class="v">{{.UniqueIPs}}</div></div>
<div class="card"><div class="k">Private / reserved sources</div><div class="v">{{.ReservedIPs}}</div></div>
{{if .PeakHour}}<div class="card"><div class="k">Peak hour</div><div class="v">{{.PeakHour}}</div></div>{{end}}
</div>

<h2>Attack types</h2>
<table>
<thead><tr><th>Category</th><th class="num">Packets (exact)</th><th class="num">Share</th><th class="num">Sampled events</th></tr></thead>
<tbody>
{{range .Categories}}<tr><td>{{.Category}}</td><td class="num">{{.Pkts}}</td><td class="num"><span class="bar red"><i style="{{bar .Share}}"></i></span>{{pct .Share}}</td><td class="num">{{.Events}}</td></tr>
{{end}}</tbody>
</table>

<table>
<thead><tr><th>Reason</th><th>Category</th><th class="num">Packets (exact)</th><th class="num">Share</th><th class="num">Sampled events</th><th>Note</th></tr></thead>
<tbody>
{{range .Reasons}}<tr><td><code>{{.Reason}}</code></td><td>{{.Category}}</td><td class="num">{{.Pkts}}</td><td class="num">{{pct .PktShare}}</td><td class="num">{{.Events}}</td><td>{{if .NoDetail}}<span class="muted">totals only, no source detail</span>{{end}}</td></tr>
{{end}}</tbody>
</table>

<h2>Top {{.TopN}} attack sources <span class="muted">(sampled)</span></h2>
<table>
<thead><tr><th class="num">#</th><th>IP</th><th class="num">Events</th><th class="num">Share</th><th class="num">Categories</th><th>First seen</th><th>Last seen</th><th>Top dst ports</th><th>Location</th></tr></thead>
<tbody>
{{range $i, $x := .TopIPs}}<tr><td class="num">{{inc $i}}</td><td><code>{{$x.IP}}</code></td><td class="num">{{$x.Events}}</td><td class="num"><span class="bar"><i style="{{bar $x.Share}}"></i></span>{{pct $x.Share}}</td><td class="num">{{$x.Categories}}</td><td class="muted">{{$x.First}}</td><td class="muted">{{$x.Last}}</td><td>{{ports $x.TopPorts}}</td><td>{{loc $x}}</td></tr>
{{end}}</tbody>
</table>

<h2>Top {{.TopN}} regions <span class="muted">(sampled)</span></h2>
{{if .GeoEnabled}}
<table>
<thead><tr><th class="num">#</th><th>Country</th><th class="num">Events</th><th class="num">Unique IPs</th><th class="num">Share</th></tr></thead>
<tbody>
{{range $i, $x := .TopGeo}}<tr><td class="num">{{inc $i}}</td><td>{{$x.Country}}</td><td class="num">{{$x.Events}}</td><td class="num">{{$x.IPs}}</td><td class="num"><span class="bar"><i style="{{bar $x.Share}}"></i></span>{{pct $x.Share}}</td></tr>
{{end}}</tbody>
</table>
{{else}}
<p class="muted">Not enabled (no geo source given). No lookups are performed online.</p>
{{end}}

<h2>Top destination ports <span class="muted">(sampled)</span></h2>
<table>
<thead><tr><th>Port</th><th class="num">Events</th><th class="num">Share</th></tr></thead>
<tbody>
{{range .TopPorts}}<tr><td><code>{{.Key}}</code></td><td class="num">{{.Count}}</td><td class="num">{{pct .Share}}</td></tr>
{{end}}</tbody>
</table>

<h2>Protocols <span class="muted">(sampled)</span></h2>
<table>
<thead><tr><th>Protocol</th><th class="num">Events</th><th class="num">Share</th></tr></thead>
<tbody>
{{range .Protos}}<tr><td>{{.Key}}</td><td class="num">{{.Count}}</td><td class="num">{{pct .Share}}</td></tr>
{{end}}</tbody>
</table>

<h2>Trend <span class="muted">(exact DROP packets)</span></h2>
<table>
<thead><tr><th>Day</th><th class="num">Packets</th><th class="num">Share</th></tr></thead>
<tbody>
{{range .Daily}}<tr><td>{{.Key}}</td><td class="num">{{.Count}}</td><td class="num">{{pct .Share}}</td></tr>
{{end}}</tbody>
</table>

<table>
<thead><tr><th>Hour</th><th class="num">Packets</th></tr></thead>
<tbody>
{{range .Hourly}}<tr><td>{{.Key}}</td><td class="num">{{.Count}}</td></tr>
{{end}}</tbody>
</table>

<h2>Data quality</h2>
<ul>
<li>Unparsable lines: {{.BadLines}}</li>
<li>Duplicate STAT lines ignored: {{.DupStats}}</li>
<li>Scan-detector summary lines (parsed, not part of DROP totals): {{.ScanLines}}</li>
<li>IPv6 share of sampled DROP events: {{pct .IPv6Share}}</li>
{{if .UnknownRsns}}<li>Unknown reasons (classifier out of date?): {{join .UnknownRsns ", "}}</li>{{end}}
{{if .Restarts}}<li>Restarts detected (STAT TOTAL went backwards): {{len .Restarts}}<ul>{{range .Restarts}}<li>{{.}}</li>{{end}}</ul></li>{{else}}<li>Restarts detected: none</li>{{end}}
</ul>

<footer>Generated offline by <code>endpoints/xdplog-report</code> from the XDP event log. Sampled tables are rate-limited by the filter and may not cover every dropped packet; exact totals come from the STAT lines.</footer>
</main>
</body>
</html>
`))
