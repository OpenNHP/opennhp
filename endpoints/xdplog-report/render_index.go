package main

import (
	"encoding/json"
	"html/template"
	"sort"
	"strings"
)

// indexPageData is the template input for the site's index.html.
type indexPageData struct {
	Summaries []SiteDaySummary
	Masked    bool
	SiteTitle string
}

// renderIndex writes index.html and index.json under opts.OutDir. The HTML
// embeds no external assets and no scripts; the JSON is the same data in
// machine-friendly form. Rows are sorted by day descending (most recent
// first), then by source name for ties.
func renderIndex(opts *SiteOptions, summaries []SiteDaySummary) error {
	ordered := append([]SiteDaySummary(nil), summaries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Day != ordered[j].Day {
			return ordered[i].Day > ordered[j].Day
		}
		return ordered[i].Source < ordered[j].Source
	})

	if err := writeIndexJSON(opts.OutDir, ordered); err != nil {
		return err
	}
	return writeIndexHTML(opts.OutDir, indexPageData{
		Summaries: ordered,
		Masked:    opts.MaskIP,
		SiteTitle: "NHP XDP attack stats",
	})
}

// writeIndexJSON writes index.json atomically.
func writeIndexJSON(outDir string, summaries []SiteDaySummary) error {
	body, err := json.MarshalIndent(summaries, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return writeFileAtomic(outDir+"/index.json", body, 0o644)
}

// writeIndexHTML executes the index template and writes index.html atomically.
func writeIndexHTML(outDir string, data indexPageData) error {
	var sb strings.Builder
	if err := htmlIndex.Execute(&sb, data); err != nil {
		return err
	}
	return writeFileAtomic(outDir+"/index.html", []byte(sb.String()), 0o644)
}

// siteFlag is one status chip in the index table's Flags column. The label
// and class are plain data; the template emits the markup so html/template
// escapes them in place. No raw template.HTML is ever built here.
type siteFlag struct {
	Label string
	Class string
}

// flagsFor returns the chips for one day's summary, in display order. An empty
// result renders as an em dash placeholder in the template.
func flagsFor(s SiteDaySummary) []siteFlag {
	var out []siteFlag
	if s.Partial {
		out = append(out, siteFlag{Label: "partial", Class: "tag-warn"})
	}
	if s.BudgetExhausted {
		out = append(out, siteFlag{Label: "budget", Class: "tag-warn"})
	}
	if s.RestartsDetected > 0 {
		out = append(out, siteFlag{Label: "restart", Class: "tag-info"})
	}
	if s.UnknownReasons {
		out = append(out, siteFlag{Label: "unknown", Class: "tag-info"})
	}
	return out
}

// htmlIndex is the index page template. There are no scripts and no external
// resources. The visual language mirrors the per-day detail page so visitors
// see the same nav and styling on both.
var htmlIndex = template.Must(template.New("htmlIndex").Funcs(template.FuncMap{
	"pct": pc,
	"dayLink": func(s SiteDaySummary) string {
		return "d/" + s.Source + "/" + s.Day + ".html"
	},
	"flags": flagsFor,
}).Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.SiteTitle}}</title>
<style>
:root{--fg:#1f2328;--muted:#656d76;--line:#d0d7de;--chip:#f6f8fa;--accent:#cf222e;--accent2:#0969da;--warn-bg:#fff8e1;--warn-fg:#7c5e16;--info-bg:#ddf4ff;--info-fg:#0969da}
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
.tag{display:inline-block;font-size:11.5px;padding:1px 6px;border-radius:8px;margin-right:3px;background:var(--chip);color:var(--muted)}
.tag-warn{background:var(--warn-bg);color:var(--warn-fg)}
.tag-info{background:var(--info-bg);color:var(--info-fg)}
a{color:var(--accent2);text-decoration:none}
a:hover{text-decoration:underline}
.note{border-left:3px solid var(--line);background:var(--chip);padding:10px 14px;border-radius:6px;margin:18px 0;color:var(--muted);font-size:13px}
footer{margin-top:34px;padding-top:14px;border-top:1px solid var(--line);color:var(--muted);font-size:12px}
@media print{body{background:#fff;padding:0}main{border:0;box-shadow:none;padding:0}}
</style>
</head>
<body>
<main>
<header>
<h1>{{.SiteTitle}}</h1>
<p class="muted">Daily summaries of <code>[NHP-DROP]</code> events from the XDP ingress filter on nhp-server.</p>
</header>

<p class="note">
Days are UTC. Click a row to open the detail page.
{{if .Masked}}Source IPs on the detail pages are truncated to <code>a.b.c.x</code> (IPv4) and the first 48 bits (IPv6).{{else}}
Source IPs on the detail pages are <strong>not</strong> masked.
{{end}}
tcp/22 scans never reach the host's NIC and are not counted. During deploys or restarts the XDP filter is detached and the gap shows as missing minutes, not as zero attacks.
</p>

<h2>Daily index</h2>
<table>
<thead><tr><th>Day (UTC)</th><th>Source</th><th class="num">DROP packets (exact)</th><th class="num">Share</th><th class="num">Unique sources</th><th>Top category</th><th>Peak hour</th><th>Flags</th></tr></thead>
<tbody>
{{range $s := .Summaries}}<tr>
<td><a href="{{dayLink $s}}">{{$s.Day}}</a></td>
<td>{{$s.Source}}</td>
<td class="num">{{$s.DropPkts}}</td>
<td class="num">{{pct $s.DropShare}}</td>
<td class="num">{{$s.UniqueIPs}}</td>
<td>{{$s.TopCategory}}</td>
<td>{{$s.PeakHour}}</td>
<td>{{if flags $s}}{{range flags $s}}<span class="tag {{.Class}}">{{.Label}}</span>{{end}}{{else}}<span class="muted">&mdash;</span>{{end}}</td>
</tr>{{end}}
</tbody>
</table>

<footer>
Generated offline by <code>endpoints/xdplog-report site</code>. Each row links to a per-day report built from the day's <code>nhp_server_xdp-YYYY-MM-DD.log</code> file pulled from the relay cron.
</footer>
</main>
</body>
</html>
`))
