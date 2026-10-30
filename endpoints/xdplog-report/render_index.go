package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
)

// siteTitle is the site's headline: the index page's <h1> and its og:title.
// The browser tab gets siteDocumentTitle instead — see there.
const siteTitle = "Invisible Server, Visible Attacks"

// siteDocumentTitle is the <title>: what a search result, a bookmark or a chat
// preview shows. It is deliberately not siteTitle, because the two answer
// different questions — a tab and a search snippet have room to say what the
// site is *about*, the heading above the fold does not.
const siteDocumentTitle = "Zero Trust in Action: Live Attacks on OpenNHP Servers"

// siteURL is the site's public origin, used only by the canonical, og:url and
// og:image tags. Those name the page to a crawler; the browser never fetches
// them, so an absolute URL in the head does not make the page reach off-host.
const siteURL = "https://stats.opennhp.org/"

// The description tag and its Open Graph twin. The on-page intro paragraph
// says "server" and this copy says "servers": one page reports one host
// today, and search and social copy is written for the whole deployment.
const (
	siteDescription   = "Real attack data from servers hidden by OpenNHP. Daily stats on port scans, SYN floods and probes silently dropped before they could connect."
	siteOGDescription = "Every scan and probe that hit our OpenNHP-protected servers, silently dropped before it could connect. Updated daily."
)

// siteOGImage is the preview image social platforms show when the homepage is
// shared. It lives in the web root like every other static asset, and this
// constant is the one place that says so — note the file itself is not
// generated here and has to be deployed alongside the HTML.
const siteOGImage = siteURL + "og.png"

// indexPageData is the template input for the site's index.html.
type indexPageData struct {
	Summaries []SiteDaySummary
	Masked    bool
	SiteTitle string
	// DocTitle is the <title>, which is not SiteTitle; Descr and OGDescr are
	// the meta description and its Open Graph twin, SiteURL their canonical
	// origin and OGImage the social preview. All are static copy.
	DocTitle string
	Descr    string
	OGDescr  string
	SiteURL  string
	OGImage  string
	// TotalRaw is the un-formatted all-time DROP packet count, used as the
	// counter's data-total attribute; TotalDropPkts is the same number with
	// thousands separators, used as the counter's static fallback text.
	TotalRaw      uint64
	TotalDropPkts string
	// SinceDay is the earliest published day, already in YYYY-MM-DD. Empty
	// when there is nothing to publish, which suppresses the hero block.
	SinceDay string
	// TotalSources and TotalCountries are the all-time distinct source IPs
	// and countries — the same unions totalnumber.js exports as
	// NHP_TOTAL_ATTACK_SOURCE_IPS and NHP_TOTAL_ATTACK_COUNTRIES. Unlike the
	// packet total they cannot be reproduced from index.json (the per-day sets
	// behind them stay private, see facts.go), so the page's note says they
	// are unions. Zero hides the figure: no geo source means no countries, and
	// showing "0 countries" would read as "no attacks from anywhere".
	//
	// With a history snapshot they are that snapshot's baseline plus the live
	// period's union: the two periods' sets were never kept, so they are
	// added rather than merged, and the figures are lower bounds.
	TotalSources   int
	TotalCountries int
	// HasHistory is true when the index carries rows frozen by a history
	// snapshot. Those rows have no detail page, and the page says so: a
	// reader who cannot click a date deserves to know why.
	HasHistory bool
}

// renderIndex writes index.html, index.json, countup.js and totalnumber.js
// under opts.OutDir. The HTML pulls in no third-party resources — the two
// scripts it references are same-origin files this function writes, and the
// absolute URLs in its head are metadata a crawler reads, not subresources a
// browser fetches — and the JSON is the same data in machine-friendly form.
// Rows are sorted by day descending (most recent first), then by source name
// for ties.
//
// union carries the all-time source and country counts, computed by the
// caller from the per-day facts. They are passed in rather than recomputed
// here so the run can report the days that could not contribute.
//
// hist is the history snapshot the caller loaded, if any. Its rows are
// already among the rows passed in; it is passed separately for the two
// things a row cannot carry: the counting start it names, and the baseline
// its period contributes to the source and country totals.
func renderIndex(opts *SiteOptions, rows []SiteDaySummary, union factsUnion, hist *SiteHistory) error {
	ordered := append([]SiteDaySummary(nil), rows...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Day != ordered[j].Day {
			return ordered[i].Day > ordered[j].Day
		}
		return ordered[i].Source < ordered[j].Source
	})

	if err := writeIndexJSON(opts.OutDir, ordered); err != nil {
		return err
	}
	if err := writeCountUpJS(opts.OutDir); err != nil {
		return err
	}
	// The all-time total is the sum of every indexed row's exact DROP packet
	// count, so days whose raw log has already aged out of the collector's
	// cache keep counting as long as their detail page is still published.
	// Days below the -since floor never reach here, so they cannot inflate
	// the number either. A frozen day's row carries the snapshot's own
	// count, which is what keeps the total equal to the sum of index.json's
	// drop_pkts even with the snapshot's days in it.
	var total uint64
	for _, s := range ordered {
		total += s.DropPkts
	}
	// The counting start comes from the snapshot when there is one: the
	// snapshot is what says the site's numbers begin before its first
	// published day.
	var since string
	if hist != nil {
		since = hist.Since
	}
	if since == "" && len(ordered) > 0 {
		// Sorted newest-first, so the last row holds the earliest day. It is
		// already YYYY-MM-DD, which is the format the whole site uses.
		since = ordered[len(ordered)-1].Day
	}
	// The source and country totals are the live period's union plus the
	// snapshot period's baseline. The two periods' address sets were never
	// kept side by side, so they are added rather than merged: a source
	// active in both periods is counted once in each, and the figures are
	// lower bounds on the true distinct totals.
	totalSources, totalCountries := union.Sources, union.Countries
	if hist != nil {
		totalSources += hist.Totals.SourceIPs
		totalCountries += hist.Totals.Countries
	}
	if err := writeTotalNumberJS(opts.OutDir, total, totalSources, totalCountries, since, hist); err != nil {
		return err
	}
	return writeIndexHTML(opts.OutDir, indexPageData{
		Summaries:     ordered,
		Masked:        opts.MaskIP,
		SiteTitle:     siteTitle,
		DocTitle:      siteDocumentTitle,
		Descr:         siteDescription,
		OGDescr:       siteOGDescription,
		SiteURL:       siteURL,
		OGImage:       siteOGImage,
		TotalRaw:      total,
		TotalDropPkts: groupThousands(total),
		SinceDay:      since,

		TotalSources:   totalSources,
		TotalCountries: totalCountries,
		HasHistory:     hist != nil,
	})
}

// groupThousands renders n with thousands separators ("1234567" becomes
// "1,234,567"). Hand-rolled rather than pulling in a formatting dependency for
// the two numbers a visitor sees on the site.
func groupThousands(n uint64) string {
	s := strconv.FormatUint(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s) + len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

// group renders a number with thousands separators from a template. Counts
// reach the templates as several Go types (uint64 packet totals, int IP
// counts), so the value is adapted here instead of at every call site.
func group(v any) string {
	switch n := v.(type) {
	case uint64:
		return groupThousands(n)
	case int:
		if n < 0 {
			return strconv.Itoa(n)
		}
		return groupThousands(uint64(n))
	case int64:
		if n < 0 {
			return strconv.FormatInt(n, 10)
		}
		return groupThousands(uint64(n))
	case uint:
		return groupThousands(uint64(n))
	default:
		return fmt.Sprint(v)
	}
}

// sourceLabels maps a source directory name to what the site calls it in
// prose. "server" alone reads as a generic noun next to a table that is
// otherwise all dates and numbers — and every host in the fleet is a server,
// so the column answered no question. The JSON keeps the directory name:
// `source` is part of the published index.json contract and changing it would
// break anyone already parsing it.
var sourceLabels = map[string]string{
	"server": "nhp-server",
}

// sourceLabel renders a source name for display, falling back to the raw
// name for any source added later (the AC, when it gets a parser).
func sourceLabel(name string) string {
	if label, ok := sourceLabels[name]; ok {
		return label
	}
	return name
}

// writeTotalNumberJS writes totalnumber.js: the all-time DROP packet total
// plus the all-time distinct source-IP and country counts, as plain
// JavaScript constants, so another site can show the same numbers with a
// <script src> tag and no JSON parsing.
//
// It is a sibling of index.json rather than a replacement for it. index.json
// stays the authoritative machine-readable artifact (one row per day, the
// packet total reproducible by summing drop_pkts); this file is the
// headline-figure answer for a consumer that wants the numbers and nothing
// else. Only the packet total is reproducible from index.json — the other two
// need per-day sets that cannot be published, see facts.go.
//
// Contents are a pure function of the summaries and the facts, so re-running
// the site build rewrites a byte-identical file and the daily cron stays a
// no-op.
func writeTotalNumberJS(outDir string, total uint64, sources, countries int, since string, hist *SiteHistory) error {
	body := totalNumberJS(total, sources, countries, since, hist)
	return writeFileAtomic(outDir+"/totalnumber.js", []byte(body), 0o644)
}

// totalNumberJS renders the export. since is empty when nothing is published;
// the constants are still emitted (as 0) so a consumer's <script src> never
// 404s and leave a dangling reference.
//
// hist is the history snapshot, when there is one. It does not change the
// numbers — the caller has already added its baseline to them — but it does
// change what they mean, and a consumer reading this file has to be told:
// the packet total spans two periods, and the source and country counts are
// two unions added together rather than one union.
func totalNumberJS(total uint64, sources, countries int, since string, hist *SiteHistory) string {
	var b strings.Builder
	b.WriteString("// Generated by endpoints/xdplog-report site. Do not edit by hand.\n")
	b.WriteString("// Attack (DROP) packets silently dropped by the nhp-server XDP ingress\n")
	b.WriteString("// filter, summed over every day this site publishes. Same number the\n")
	b.WriteString("// homepage hero shows; index.json is the row-level source of truth.\n")
	b.WriteString("//\n")
	if hist != nil {
		b.WriteString("// The total spans two periods: the days a saved history snapshot\n")
		b.WriteString("// froze contribute the snapshot's own counts, and the days published\n")
		b.WriteString("// since add theirs on top.\n")
		b.WriteString("//\n")
	}
	b.WriteString("// The next two are unions, not sums: an address or country seen on\n")
	b.WriteString("// ten days is counted once. Private/reserved sources are excluded,\n")
	b.WriteString("// and a source the GeoLite2 database cannot place belongs to no\n")
	b.WriteString("// country, so it raises the source count only.\n")
	if hist != nil {
		b.WriteString("// With a history snapshot they are the union of each period's sets\n")
		b.WriteString("// added together: an address or country seen in the snapshot period\n")
		b.WriteString("// and again in the live period is counted once in each, so both\n")
		b.WriteString("// figures are lower bounds on the true distinct totals.\n")
	}
	if since != "" {
		fmt.Fprintf(&b, "// Counting starts: %s (UTC).\n", since)
	}
	fmt.Fprintf(&b, "const NHP_TOTAL_ATTACK_BLOCKED = %d;\n", total)
	fmt.Fprintf(&b, "const NHP_TOTAL_ATTACK_SOURCE_IPS = %d;\n", sources)
	fmt.Fprintf(&b, "const NHP_TOTAL_ATTACK_COUNTRIES = %d;\n", countries)
	return b.String()
}

// writeCountUpJS writes the homepage counter's script next to index.html. Its
// contents are a compile-time constant, so re-running the site build rewrites
// a byte-identical file.
func writeCountUpJS(outDir string) error {
	return writeFileAtomic(outDir+"/countup.js", []byte(countUpJS), 0o644)
}

// countUpJS is the homepage counter animation, served same-origin from the
// site root and allowed by the nginx vhost's `script-src 'self'`. It animates
// every element carrying data-total: the packet total, source IPs and
// countries.
//
// Progressive enhancement: the hero span in index.html already contains the
// real formatted total, so with JavaScript disabled, blocked by an older CSP,
// or under prefers-reduced-motion the number simply reads as static text.
// The script only rewinds it to 0 and counts up.
const countUpJS = `(function () {
  "use strict";
  var reduce = window.matchMedia &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  if (reduce || typeof requestAnimationFrame !== "function") return;

  var DURATION = 2500;
  function easeOutExpo(t) {
    return t === 1 ? 1 : 1 - Math.pow(2, -10 * t);
  }

  function countUp(el) {
    var target = parseInt(el.getAttribute("data-total"), 10);
    if (!isFinite(target) || target <= 0) return;
    var final = el.textContent;

    function run() {
      var start = null;
      function frame(now) {
        if (start === null) start = now;
        var p = Math.min((now - start) / DURATION, 1);
        el.textContent = Math.round(target * easeOutExpo(p))
          .toLocaleString("en-US");
        if (p < 1) {
          requestAnimationFrame(frame);
        } else {
          // Land on the server-rendered text, never on a rounded float.
          el.textContent = final;
        }
      }
      requestAnimationFrame(frame);
    }

    // The hero sits above the fold, but observe anyway: on a short viewport
    // the animation should start when the counter is seen, not at page load.
    if (typeof IntersectionObserver !== "function") {
      run();
      return;
    }
    var io = new IntersectionObserver(function (entries) {
      for (var i = 0; i < entries.length; i++) {
        if (!entries[i].isIntersecting) continue;
        io.disconnect();
        run();
        return;
      }
    }, { threshold: 0.4 });
    io.observe(el);
  }

  var els = document.querySelectorAll("[data-total]");
  for (var n = 0; n < els.length; n++) countUp(els[n]);
})();
`

// writeIndexJSON writes index.json atomically.
func writeIndexJSON(outDir string, summaries []SiteDaySummary) error {
	// Marshal an empty slice rather than a nil one: json.MarshalIndent writes
	// nil as `null`, and the contract is a JSON array. A consumer doing
	// `for (const row of data)` gets a TypeError on null, which on an empty
	// site would look like a broken deploy rather than "no days yet".
	rows := summaries
	if rows == nil {
		rows = []SiteDaySummary{}
	}
	body, err := json.MarshalIndent(rows, "", "  ")
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
//
// The history chip comes first: it is the one that changes how the row is
// read, because a frozen day has no detail page to click.
func flagsFor(s SiteDaySummary) []siteFlag {
	var out []siteFlag
	if s.History {
		out = append(out, siteFlag{Label: "history", Class: "tag-info"})
	}
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

// htmlIndex is the index page template. It loads no third-party resource: the
// only subresources are the two same-origin scripts this generator writes —
// countup.js, the hero counter, and totalnumber.js, whose NHP_TOTAL_ATTACK_BLOCKED
// binding the homepage loads for itself, so a reader (or another script on the
// page) can quote the same total the hero shows — and the nginx vhost allows
// them via `script-src 'self'`. The absolute URLs in <head> are canonical/og
// metadata naming the page's own origin; a crawler reads them, the browser
// never fetches them. The visual language mirrors the per-day detail page so
// visitors see the same nav and styling on both.
var htmlIndex = template.Must(template.New("htmlIndex").Funcs(template.FuncMap{
	"pct": pc,
	"dayLink": func(s SiteDaySummary) string {
		return "d/" + s.Source + "/" + s.Day + ".html"
	},
	"sourceLabel": sourceLabel,
	"flags":       flagsFor,
	"group":       group,
}).Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.DocTitle}}</title>
<meta name="description" content="{{.Descr}}">
<link rel="canonical" href="{{.SiteURL}}">
<meta property="og:title" content="{{.SiteTitle}}">
<meta property="og:description" content="{{.OGDescr}}">
<meta property="og:url" content="{{.SiteURL}}">
<meta property="og:type" content="website">
<meta property="og:image" content="{{.OGImage}}">
<meta name="twitter:card" content="summary_large_image">
<script src="countup.js" defer></script>
<script src="./totalnumber.js" defer></script>
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
.tablewrap{overflow-x:auto}
table{width:100%;border-collapse:collapse;margin:6px 0 22px;font-variant-numeric:tabular-nums}
th,td{padding:6px 8px;border-bottom:1px solid var(--line);text-align:left;vertical-align:top}
th{font-size:11.5px;text-transform:uppercase;letter-spacing:.05em;color:var(--muted);background:var(--chip);white-space:nowrap}
td.num,th.num{text-align:right}
/* Day, top category and peak hour are single tokens a reader scans down the
   column; letting them break mid-value (2026-10-<wbr>19) made every row two
   lines tall. nowrap pushes the width onto the table instead, and the wrapper
   scrolls when the sum exceeds the viewport. */
td.nowrap{white-space:nowrap}
.muted{color:var(--muted)}
.tag{display:inline-block;font-size:11.5px;padding:1px 6px;border-radius:8px;margin-right:3px;background:var(--chip);color:var(--muted)}
.tag-warn{background:var(--warn-bg);color:var(--warn-fg)}
.tag-info{background:var(--info-bg);color:var(--info-fg)}
a{color:var(--accent2);text-decoration:none}
a:hover{text-decoration:underline}
.note{border-left:3px solid var(--line);background:var(--chip);padding:10px 14px;border-radius:6px;margin:18px 0;color:var(--muted);font-size:13px}
/* Hero counter. The label says "packets" because that is what DropPkts
   counts — the daily table calls the same number "DROP packets (exact)" and
   the detail pages "Attack (DROP) packets"; calling it attacks would
   overstate what a dropped packet is. The three figures (packets, source IPs,
   countries) sit side by side with the label under each. They are green
   rather than the red accent so they read as attacks stopped, not as a
   current alert, and deliberately do not use tabular-nums: the digits change
   width while counting and a fixed advance makes them jitter. */
.hero{margin:26px 0 8px;text-align:center}
.hero-label{font-size:11.5px;text-transform:uppercase;letter-spacing:.08em;color:var(--muted)}
.hero-num{font-size:48px;font-weight:700;line-height:1.15;letter-spacing:-.02em;color:#1a7f37;animation:hero-in .9s cubic-bezier(.16,1,.3,1) both}
.hero-stats{display:flex;flex-wrap:wrap;justify-content:center;align-items:flex-start;gap:12px 56px}
.hero-stat{text-align:center}
.hero-stat .hero-label{margin-top:4px}
@keyframes hero-in{from{opacity:0;transform:translateY(12px) scale(.97)}to{opacity:1;transform:none}}
@media (prefers-reduced-motion:reduce){.hero-num{animation:none}}
footer{margin-top:34px;padding-top:14px;border-top:1px solid var(--line);color:var(--muted);font-size:12px}
@media print{body{background:#fff;padding:0}main{border:0;box-shadow:none;padding:0}}
</style>
</head>
<body>
<main>
<header>
<h1>{{.SiteTitle}}</h1>
<p class="muted">A live log of every scan and probe that hit our OpenNHP-protected server and was silently dropped before it could connect. Updated daily.</p>
</header>

{{if .SinceDay}}
<section class="hero" aria-label="All-time attack totals">
<div class="hero-stats">
<div class="hero-stat">
<div><span class="hero-num" id="total-attacks" data-total="{{.TotalRaw}}">{{.TotalDropPkts}}</span></div>
<div class="hero-label">Attack packets blocked</div>
</div>
{{if .TotalSources}}<div class="hero-stat">
<div><span class="hero-num" id="total-sources" data-total="{{.TotalSources}}">{{group .TotalSources}}</span></div>
<div class="hero-label">Source IPs</div>
</div>{{end}}
{{if .TotalCountries}}<div class="hero-stat">
<div><span class="hero-num" id="total-countries" data-total="{{.TotalCountries}}">{{group .TotalCountries}}</span></div>
<div class="hero-label">Countries</div>
</div>{{end}}
</div>
</section>
{{end}}

<p class="note">
{{if .SinceDay}}Since {{.SinceDay}}. {{end}}Days are UTC. Click a row to open the detail page.
{{if .Masked}}Source IPs on the detail pages are truncated to <code>a.b.c.x</code> (IPv4) and the first 48 bits (IPv6).{{else}}
Source IPs on the detail pages are <strong>not</strong> masked.
{{end}}
tcp/22 scans never reach the host's NIC and are not counted. During deploys or restarts the XDP filter is detached and the gap shows as missing minutes, not as zero attacks.
{{if .HasHistory}}Rows tagged <span class="tag tag-info">history</span> come from a saved snapshot of this page and have no detail page. Source IP and country totals are the union of each period's sets added together, so a source active in two periods is counted once in each. {{end}}
</p>

<h2>Daily index</h2>
<div class="tablewrap">
<table>
<thead><tr><th>Day (UTC)</th><th>Source</th><th class="num">DROP packets (exact)</th><th class="num">Share</th><th class="num">Unique sources</th><th>Top category</th><th>Peak hour</th><th>Flags</th></tr></thead>
<tbody>
{{range $s := .Summaries}}<tr>
<td class="nowrap">{{if $s.History}}{{$s.Day}}{{else}}<a href="{{dayLink $s}}">{{$s.Day}}</a>{{end}}</td>
<td>{{sourceLabel $s.Source}}</td>
<td class="num">{{$s.DropPkts}}</td>
<td class="num">{{pct $s.DropShare}}</td>
<td class="num">{{$s.UniqueIPs}}</td>
<td class="nowrap">{{$s.TopCategory}}</td>
<td class="nowrap">{{$s.PeakHour}}</td>
<td>{{if flags $s}}{{range flags $s}}<span class="tag {{.Class}}">{{.Label}}</span>{{end}}{{else}}<span class="muted">&mdash;</span>{{end}}</td>
</tr>{{end}}
</tbody>
</table>
</div>

<footer>
Generated offline by <code>endpoints/xdplog-report site</code>. Each row links to a per-day report built from the day's <code>nhp_server_xdp-YYYY-MM-DD.log</code> file pulled from the relay cron.
{{if .HasHistory}}Rows marked <span class="tag tag-info">history</span> come from a <code>history.json</code> snapshot generated by <code>xdplog-report history</code>; they have no detail page. {{end}}
</footer>
</main>
</body>
</html>
`))
