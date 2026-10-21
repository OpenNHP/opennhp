package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"
)

// siteTitle is the heading on the site's index page (both <title> and <h1>).
const siteTitle = "Invisible Server, Visible Attacks"

// indexPageData is the template input for the site's index.html.
type indexPageData struct {
	Summaries []SiteDaySummary
	Masked    bool
	SiteTitle string
	// TotalRaw is the un-formatted all-time DROP packet count, used as the
	// counter's data-total attribute; TotalDropPkts is the same number with
	// thousands separators, used as the counter's static fallback text.
	TotalRaw      uint64
	TotalDropPkts string
	// SinceDay is the earliest day in the index, formatted MM-DD-YYYY. Empty
	// when there is nothing to publish, which suppresses the hero block.
	SinceDay string
}

// renderIndex writes index.html, index.json and countup.js under opts.OutDir.
// The HTML pulls in no third-party resources — the one script it references is
// a same-origin file this function writes — and the JSON is the same data in
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
	if err := writeCountUpJS(opts.OutDir); err != nil {
		return err
	}
	// The all-time total is the sum of every indexed row's exact DROP packet
	// count, so days whose raw log has already aged out of the collector's
	// cache keep counting as long as their detail page is still published.
	var total uint64
	for _, s := range ordered {
		total += s.DropPkts
	}
	var since string
	if n := len(ordered); n > 0 {
		// Sorted newest-first, so the last row holds the earliest day.
		since = formatSince(ordered[n-1].Day)
	}
	return writeIndexHTML(opts.OutDir, indexPageData{
		Summaries:     ordered,
		Masked:        opts.MaskIP,
		SiteTitle:     siteTitle,
		TotalRaw:      total,
		TotalDropPkts: groupThousands(total),
		SinceDay:      since,
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

// formatSince renders a YYYY-MM-DD day as MM-DD-YYYY, the format the homepage
// hero uses for its "Since" line. An unparsable day is passed through
// unchanged rather than dropped: a malformed date in the index is a collector
// bug worth showing, not one worth hiding.
func formatSince(day string) string {
	t, err := time.Parse("2006-01-02", day)
	if err != nil {
		return day
	}
	return t.Format("01-02-2006")
}

// writeCountUpJS writes the homepage counter's script next to index.html. Its
// contents are a compile-time constant, so re-running the site build rewrites
// a byte-identical file.
func writeCountUpJS(outDir string) error {
	return writeFileAtomic(outDir+"/countup.js", []byte(countUpJS), 0o644)
}

// countUpJS is the homepage counter animation, served same-origin from the
// site root and allowed by the nginx vhost's `script-src 'self'`.
//
// Progressive enhancement: the hero span in index.html already contains the
// real formatted total, so with JavaScript disabled, blocked by an older CSP,
// or under prefers-reduced-motion the number simply reads as static text.
// The script only rewinds it to 0 and counts up.
const countUpJS = `(function () {
  "use strict";
  var el = document.getElementById("total-attacks");
  if (!el) return;
  var target = parseInt(el.getAttribute("data-total"), 10);
  if (!isFinite(target) || target <= 0) return;
  var final = el.textContent;
  var reduce = window.matchMedia &&
    window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  if (reduce || typeof requestAnimationFrame !== "function") return;

  var DURATION = 2500;
  function easeOutExpo(t) {
    return t === 1 ? 1 : 1 - Math.pow(2, -10 * t);
  }
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

  // The hero sits above the fold, but observe anyway: on a short viewport the
  // animation should start when the counter is seen, not at page load.
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
})();
`

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

// htmlIndex is the index page template. It references no third-party
// resource; the only script is the self-hosted homepage counter (countup.js),
// which the nginx vhost allows via `script-src 'self'`. The visual language
// mirrors the per-day detail page so visitors see the same nav and styling on
// both.
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
<script src="countup.js" defer></script>
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
   overstate what a dropped packet is. The number stays in ink rather than
   the red accent so it reads as a total, not as a current alert, and it
   deliberately does not use tabular-nums: the digits change width while
   counting and a fixed advance makes them jitter. */
.hero{margin:26px 0 8px;text-align:center}
.hero-label{font-size:11.5px;text-transform:uppercase;letter-spacing:.08em;color:var(--muted)}
.hero-num{font-size:48px;font-weight:600;line-height:1.15;letter-spacing:-.02em;color:var(--fg);animation:hero-in .9s cubic-bezier(.16,1,.3,1) both}
.hero-since{font-size:12px;color:var(--muted)}
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
<section class="hero" aria-label="Total attack packets blocked">
<div class="hero-label">Total attack packets blocked</div>
<div><span class="hero-num" id="total-attacks" data-total="{{.TotalRaw}}">{{.TotalDropPkts}}</span></div>
<div class="hero-since">Since {{.SinceDay}}</div>
</section>
{{end}}

<p class="note">
Days are UTC. Click a row to open the detail page.
{{if .Masked}}Source IPs on the detail pages are truncated to <code>a.b.c.x</code> (IPv4) and the first 48 bits (IPv6).{{else}}
Source IPs on the detail pages are <strong>not</strong> masked.
{{end}}
tcp/22 scans never reach the host's NIC and are not counted. During deploys or restarts the XDP filter is detached and the gap shows as missing minutes, not as zero attacks.
</p>

<h2>Daily index</h2>
<div class="tablewrap">
<table>
<thead><tr><th>Day (UTC)</th><th>Source</th><th class="num">DROP packets (exact)</th><th class="num">Share</th><th class="num">Unique sources</th><th>Top category</th><th>Peak hour</th><th>Flags</th></tr></thead>
<tbody>
{{range $s := .Summaries}}<tr>
<td class="nowrap"><a href="{{dayLink $s}}">{{$s.Day}}</a></td>
<td>{{$s.Source}}</td>
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
</footer>
</main>
</body>
</html>
`))
