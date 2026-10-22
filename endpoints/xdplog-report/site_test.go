package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The two days the site fixtures are staged under. They sit on and after
// defaultSiteSince so the -since floor keeps them published; the checked-in
// testdata files keep their original names because report_test.go reads them
// by date for the single-file CLI, which has no floor. Dates come from file
// names, so staging the same bytes under a different day is faithful.
const (
	siteDay1 = "2026-10-16"
	siteDay2 = "2026-10-17"
)

// stageTestdata copies the checked-in testdata log files into a fresh
// <tmp>/raw/<source>/ directory so each test gets an isolated input set.
// Returns (logsDir, outDir).
func stageTestdata(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	logs := filepath.Join(root, "raw")
	out := filepath.Join(root, "web")
	src := filepath.Join(logs, "server")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	// fixture day -> testdata file whose bytes it stages.
	for _, pair := range []struct{ day, fixture string }{
		{siteDay1, "2026-10-01"},
		{siteDay2, "2026-10-02"},
	} {
		b, err := os.ReadFile(filepath.Join("testdata", "nhp_server_xdp-"+pair.fixture+".log"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(src, "nhp_server_xdp-"+pair.day+".log"), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return logs, out
}

// runSite invokes the site subcommand. extraArgs is appended after the default
// -logs / -out / -top flags.
func runSiteCmd(t *testing.T, logs, out string, extraArgs ...string) {
	t.Helper()
	args := []string{"site", "-logs", logs, "-out", out, "-top", "20"}
	args = append(args, extraArgs...)
	var buf bytes.Buffer
	if err := run(args, &buf); err != nil {
		t.Fatalf("runSiteCmd: %v\n%s", err, buf.String())
	}
}

// indexAndPages enumerates the files that the site subcommand owns: the
// index, the scripts it writes, state, and every detail page under
// d/<source>/<day>.{html,json}.
func indexAndPages(out string) []string {
	files := []string{
		filepath.Join(out, "index.html"),
		filepath.Join(out, "index.json"),
		filepath.Join(out, "countup.js"),
		filepath.Join(out, "totalnumber.js"),
		filepath.Join(out, "state.json"),
	}
	_ = filepath.Walk(out, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasPrefix(p, filepath.Join(out, "d", "server")) {
			files = append(files, p)
		}
		return nil
	})
	return files
}

func TestSiteGeneratesDays(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	for _, day := range []string{siteDay1, siteDay2} {
		htmlPath := filepath.Join(out, "d", "server", day+".html")
		jsonPath := filepath.Join(out, "d", "server", day+".json")
		if _, err := os.Stat(htmlPath); err != nil {
			t.Errorf("%s missing: %v", htmlPath, err)
		}
		if _, err := os.Stat(jsonPath); err != nil {
			t.Errorf("%s missing: %v", jsonPath, err)
		}
	}
	for _, f := range []string{"index.html", "index.json", "state.json"} {
		if _, err := os.Stat(filepath.Join(out, f)); err != nil {
			t.Errorf("%s missing: %v", f, err)
		}
	}
}

// TestSiteDetailTitleAndSubtitle covers the per-day page's headline: the
// "Daily Attack Report: <day>" title and the sentence under it. The numbers in
// that sentence are asserted against the day's own JSON, so prose and the
// Overview cards below it can never disagree.
func TestSiteDetailTitleAndSubtitle(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	const day = siteDay1
	jb, err := os.ReadFile(filepath.Join(out, "d", "server", day+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var rep Report
	if err = json.Unmarshal(jb, &rep); err != nil {
		t.Fatal(err)
	}
	pb, err := os.ReadFile(filepath.Join(out, "d", "server", day+".html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(pb)

	// The site publishes one day per page, so the title names that one day —
	// Report.From/To carry timestamps and must not leak into it.
	if !strings.Contains(page, "<title>Daily Attack Report: "+day+"</title>") {
		t.Errorf("detail page title is not the day-scoped headline:\n%s", firstLines(page, 8))
	}
	if !strings.Contains(page, "<h1>Daily Attack Report: "+day+"</h1>") {
		t.Errorf("detail page h1 is not the day-scoped headline:\n%s", firstLines(page, 12))
	}
	want := fmt.Sprintf("%s hostile packets from %s sources were silently dropped",
		groupThousands(rep.DropPkts), group(rep.UniqueIPs))
	if !strings.Contains(page, want) {
		t.Errorf("detail subtitle missing %q:\n%s", want, firstLines(page, 12))
	}
	if want := "That&rsquo;s " + pc(rep.DropShare) + " of all inbound traffic"; !strings.Contains(page, want) {
		t.Errorf("detail subtitle missing %q", want)
	}
	// The old third line still carries the full timestamp range, which the
	// title deliberately drops.
	if !strings.Contains(page, rep.From+" &ndash; "+rep.To) {
		t.Error("detail page lost its servers/timescale line")
	}
}

// firstLines returns the first n lines of s, for test failure messages.
func firstLines(s string, n int) string {
	lines := strings.SplitN(s, "\n", n+1)
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

func TestSiteIndexOrderDesc(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err = json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	if len(idx) != 2 {
		t.Fatalf("index rows = %d, want 2", len(idx))
	}
	if idx[0].Day != siteDay2 || idx[1].Day != siteDay1 {
		t.Errorf("index order = %s, %s", idx[0].Day, idx[1].Day)
	}
}

// TestSiteIdempotentNoRewrite asserts the second run is a no-op: every file
// the first run created must keep its mtime. Combined with state.json this
// means a daily cron hitting an unchanged input set does zero work.
func TestSiteIdempotentNoRewrite(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	before := map[string]int64{}
	for _, p := range indexAndPages(out) {
		st, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		before[p] = st.ModTime().UnixNano()
	}

	runSiteCmd(t, logs, out)

	for p, mt := range before {
		st, err := os.Stat(p)
		if err != nil {
			t.Fatalf("missing %s: %v", p, err)
		}
		if st.ModTime().UnixNano() != mt {
			t.Errorf("%s was rewritten: %d -> %d", p, mt, st.ModTime().UnixNano())
		}
	}
}

// TestSiteIncrementalReRender exercises the state.json fingerprint: only the
// day whose input changed must be rewritten.
func TestSiteIncrementalReRender(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	day1 := filepath.Join(out, "d", "server", siteDay1+".html")
	day2 := filepath.Join(out, "d", "server", siteDay2+".html")
	st1, _ := os.Stat(day1)
	st2, _ := os.Stat(day2)

	// Grow day1 input so its size+mtime change.
	src := filepath.Join(logs, "server", "nhp_server_xdp-"+siteDay1+".log")
	f, err := os.OpenFile(src, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("12:34:56 srv1 [NHP-DROP] REASON=TCP_OTHER SRC=203.0.113.99 DST=x LEN=1 PROTO=TCP SPT=1 DPT=80 RELAY=0\n"); err != nil {
		t.Fatal(err)
	}
	f.Close()

	runSiteCmd(t, logs, out)

	st1After, _ := os.Stat(day1)
	st2After, _ := os.Stat(day2)
	if st1After.ModTime().UnixNano() == st1.ModTime().UnixNano() {
		t.Errorf("day1 page should have been rewritten")
	}
	if st2After.ModTime().UnixNano() != st2.ModTime().UnixNano() {
		t.Errorf("day2 page should NOT have been rewritten")
	}
}

// TestSiteMaskIPDefault asserts the public-facing site never carries a full
// IPv4 anywhere. Masking is also exercised in the existing single-file test
// (TestCLIFormatsAndMask), so here we focus on the index, the detail page,
// and the JSON files the site publishes.
func TestSiteMaskIPDefault(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	fullV4 := regexp.MustCompile(`\b(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})\b`)
	maskedV4 := regexp.MustCompile(`\b(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.x\b`)

	htmlPaths := []string{
		filepath.Join(out, "index.html"),
		filepath.Join(out, "d", "server", siteDay1+".html"),
	}
	for _, p := range htmlPaths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range fullV4.FindAllString(string(b), -1) {
			if !strings.HasSuffix(m, ".x") {
				t.Errorf("%s contains full IPv4 %q", p, m)
			}
		}
	}

	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range fullV4.FindAllString(string(b), -1) {
		if !strings.HasSuffix(m, ".x") {
			t.Errorf("index.json contains full IPv4 %q", m)
		}
	}

	// Sanity: at least one masked form should appear in the detail page so a
	// regression that drops masking entirely fails loudly.
	detail, err := os.ReadFile(filepath.Join(out, "d", "server", siteDay1+".html"))
	if err != nil {
		t.Fatal(err)
	}
	if !maskedV4.Match(detail) {
		t.Error("detail page contains no masked IPv4; masking may be silently disabled")
	}
}

// TestSiteIgnoreInvalidFilename asserts evaluate/audit files and malformed
// dates are silently dropped from the index.
func TestSiteIgnoreInvalidFilename(t *testing.T) {
	logs, out := stageTestdata(t)
	src := filepath.Join(logs, "server")

	skip := []string{
		"nhp_server_xdp-evaluate-2026-10-01.log",
		"nhp_server_xdp-audit-2026-10-01.log",
		"nhp_server_xdp-2026-1-1.log",         // bad date
		"nhp_server_xdp-20261001.log",         // no dashes
		"nhp_server_xdp-" + siteDay1 + ".txt", // wrong extension
		"other_server_xdp-2026-10-03.log",     // wrong prefix
	}
	for _, name := range skip {
		if err := os.WriteFile(filepath.Join(src, name), []byte("garbage\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runSiteCmd(t, logs, out)

	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err = json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, s := range idx {
		got[s.Day] = true
	}
	for _, d := range []string{siteDay1, siteDay2} {
		if !got[d] {
			t.Errorf("missing valid day %s", d)
		}
	}
	if got["2026-10-03"] {
		t.Error("rogue day 2026-10-03 leaked into index")
	}
	if _, err := os.Stat(filepath.Join(out, "d", "server", siteDay1+".txt.html")); err == nil {
		t.Error(".txt file was rendered as a page")
	}
}

// TestSitePartialFlag asserts today's row carries partial=true and historical
// days do not. We use the XDPSTATS_TODAY env hook to pin "today" without
// depending on wall clock.
func TestSitePartialFlag(t *testing.T) {
	const today = "2099-01-15"
	logs, out := stageTestdata(t)

	src := filepath.Join(logs, "server")
	if err := os.WriteFile(filepath.Join(src, "nhp_server_xdp-"+today+".log"),
		[]byte("10:00:00 s [NHP-STAT] VERDICT=DROP REASON=TCP_OTHER WINDOW=60s PKTS=5 LOGGED=1 TOTAL=5\n"),
		0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDPSTATS_TODAY", today)
	runSiteCmd(t, logs, out)

	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err = json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	sort.Slice(idx, func(i, j int) bool { return idx[i].Day < idx[j].Day })

	var got *SiteDaySummary
	for i := range idx {
		if idx[i].Day == today {
			got = &idx[i]
		}
	}
	if got == nil {
		t.Fatalf("today's row %s missing", today)
	}
	if !got.Partial {
		t.Errorf("today row should be partial")
	}
	for _, s := range idx {
		if s.Day != today && s.Partial {
			t.Errorf("historical day %s marked partial", s.Day)
		}
	}

	// The index page renders the chip itself; the template emits the markup so
	// nothing needs an unsafe template.HTML conversion.
	page, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), `<span class="tag tag-warn">partial</span>`) {
		t.Errorf("index.html is missing the partial chip for %s", today)
	}
}

// TestFlagsFor pins the Flags column: chip order, the classes that drive the
// colors, and the empty case that renders as an em dash placeholder.
func TestFlagsFor(t *testing.T) {
	tests := []struct {
		name string
		sum  SiteDaySummary
		want []siteFlag
	}{
		{
			name: "none",
			sum:  SiteDaySummary{},
			want: nil,
		},
		{
			name: "all in display order",
			sum: SiteDaySummary{
				UnknownReasons:   true,
				RestartsDetected: 2,
				BudgetExhausted:  true,
				Partial:          true,
			},
			want: []siteFlag{
				{Label: "partial", Class: "tag-warn"},
				{Label: "budget", Class: "tag-warn"},
				{Label: "restart", Class: "tag-info"},
				{Label: "unknown", Class: "tag-info"},
			},
		},
		{
			name: "warnings only",
			sum:  SiteDaySummary{Partial: true},
			want: []siteFlag{{Label: "partial", Class: "tag-warn"}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := flagsFor(tc.sum)
			if len(got) != len(tc.want) {
				t.Fatalf("flagsFor = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("flag[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// TestSiteGeoCSVPassesThrough exercises the geo path end-to-end through the
// site subcommand, verifying that geoEnabled flows into the per-day JSON.
func TestSiteGeoCSVPassesThrough(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out, "-geo-csv", "testdata/geo.csv")

	b, err := os.ReadFile(filepath.Join(out, "d", "server", siteDay1+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var rep Report
	if err := json.Unmarshal(b, &rep); err != nil {
		t.Fatal(err)
	}
	if !rep.GeoEnabled {
		t.Error("geo_enabled should be true with -geo-csv")
	}
	if len(rep.TopGeo) == 0 {
		t.Error("expected top_geo rows")
	}
}

// TestSiteBackLinkResolves asserts the "back to index" href on a detail page
// actually lands on the generated index.html. Detail pages are written to
// <out>/d/<source>/<day>.html, so the correct relative path is two levels up;
// "../../index.html" resolving to an existing file is the whole assertion.
// A plain link-presence check would not have caught the earlier single-".."
// value, which resolves to <out>/d/index.html — a 404 in production.
func TestSiteBackLinkResolves(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	page := filepath.Join(out, "d", "server", siteDay1+".html")
	b, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}
	m := backlinkRe.FindStringSubmatch(string(b))
	if m == nil {
		t.Fatalf("%s has no back-to-index link", page)
	}
	target := filepath.Clean(filepath.Join(filepath.Dir(page), m[1]))
	want := filepath.Join(out, "index.html")
	if target != want {
		t.Errorf("back link resolves to %s, want %s", target, want)
	}
	if _, err := os.Stat(target); err != nil {
		t.Errorf("back link target does not exist: %v", err)
	}
}

var backlinkRe = regexp.MustCompile(`href="(\.\./[^"]*index\.html)"`)

// TestSiteStateOutsideWebRoot asserts -state keeps the incremental-build state
// out of the directory nginx publishes. state.json records the absolute path
// of every raw cache file, which is host layout an anonymous visitor has no
// business reading, so the relay deploy points -state at a private directory
// and the default only stays inside -out for local/CI use.
func TestSiteStateOutsideWebRoot(t *testing.T) {
	logs, out := stageTestdata(t)
	stateDir := filepath.Join(filepath.Dir(out), "state")
	statePath := filepath.Join(stateDir, "state.json")
	runSiteCmd(t, logs, out, "-state", statePath)

	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("-state file not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "state.json")); !os.IsNotExist(err) {
		t.Errorf("state.json leaked into the web root (err=%v)", err)
	}
	// The pages themselves must still be complete: the state is what makes
	// the second run a no-op.
	for _, day := range []string{siteDay1, siteDay2} {
		if _, err := os.Stat(filepath.Join(out, "d", "server", day+".html")); err != nil {
			t.Errorf("%s missing: %v", day, err)
		}
	}
	before, err := os.Stat(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	runSiteCmd(t, logs, out, "-state", statePath)
	after, err := os.Stat(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if before.ModTime().UnixNano() != after.ModTime().UnixNano() {
		t.Error("second run with -state should be a no-op; state.json outside -out still works")
	}
}

// netip sanity so we can reference netip in case future tests use it
var _ = netip.MustParseAddr("203.0.113.5")

// TestSiteRebuildsWhenPageMissingButStateSaysBuilt covers the public-site
// failure mode where state.json recorded a day as "built" but the detail
// page on disk has gone missing — a cleared web root, a -out change, or a
// half-finished first run. Without the dirty-on-missing-page branch, the
// index published 0 DROP packets with a link to a 404, and "no attacks" was
// the wrong answer on a public "attack stats" site. The day must be
// rebuilt from the still-present raw input.
func TestSiteRebuildsWhenPageMissingButStateSaysBuilt(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	day := siteDay1
	htmlPath := filepath.Join(out, "d", "server", day+".html")
	jsonPath := filepath.Join(out, "d", "server", day+".json")
	if err := os.Remove(htmlPath); err != nil {
		t.Fatalf("remove html: %v", err)
	}
	if err := os.Remove(jsonPath); err != nil {
		t.Fatalf("remove json: %v", err)
	}

	runSiteCmd(t, logs, out)

	if _, err := os.Stat(htmlPath); err != nil {
		t.Errorf("day page was not rebuilt: %v", err)
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Errorf("day json was not rebuilt: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err = json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	for _, s := range idx {
		if s.Day != day {
			continue
		}
		if s.DropPkts == 0 {
			t.Errorf("index row for %s published 0 DROP packets after rebuild; expected a non-zero count", day)
		}
		return
	}
	t.Fatalf("day %s missing from index", day)
}

// TestSiteKeepsOrphanDetailPage covers the second failure mode: the
// collector prunes a raw cache older than RETENTION_DAYS but the detail
// page under d/<src>/<day>.{html,json} is still published. The next
// incremental rebuild must not drop the day from the index — it must
// fall back to the on-disk page so the public history keeps growing.
//
// We simulate the prune by leaving only the day-2 raw input and
// pre-creating a day-1 detail page. The index must list both.
func TestSiteKeepsOrphanDetailPage(t *testing.T) {
	root := t.TempDir()
	logs := filepath.Join(root, "raw")
	out := filepath.Join(root, "web")
	srcLogs := filepath.Join(logs, "server")
	srcOut := filepath.Join(out, "d", "server")
	if err := os.MkdirAll(srcLogs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(srcOut, 0o755); err != nil {
		t.Fatal(err)
	}

	// day-1 raw log is GONE (pruned by the collector), but a detail page
	// already lives on disk.
	const orphanDay = siteDay1
	const freshDay = siteDay2
	if err := os.WriteFile(filepath.Join(srcOut, orphanDay+".html"),
		[]byte("<!doctype html><html><body>orphan</body></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	orphanRep := Report{
		Schema:    1,
		DropPkts:  1234,
		PassPkts:  42,
		UniqueIPs: 7,
	}
	ob, err := json.Marshal(&orphanRep)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(srcOut, orphanDay+".json"), ob, 0o644); err != nil {
		t.Fatal(err)
	}

	// day-2 is a normal raw input that gets rebuilt into a page.
	if err = os.WriteFile(filepath.Join(srcLogs, "nhp_server_xdp-"+freshDay+".log"),
		[]byte("10:00:00 s [NHP-STAT] VERDICT=DROP REASON=TCP_OTHER WINDOW=60s PKTS=5 LOGGED=1 TOTAL=5\n"),
		0o644); err != nil {
		t.Fatal(err)
	}

	runSiteCmd(t, logs, out)

	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err = json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	var gotOrphan *SiteDaySummary
	var gotFresh *SiteDaySummary
	for i := range idx {
		switch idx[i].Day {
		case orphanDay:
			gotOrphan = &idx[i]
		case freshDay:
			gotFresh = &idx[i]
		}
	}
	if gotOrphan == nil {
		t.Fatalf("orphan day %s missing from index after raw prune", orphanDay)
	}
	if gotOrphan.DropPkts != 1234 || gotOrphan.PassPkts != 42 || gotOrphan.UniqueIPs != 7 {
		t.Errorf("orphan day summary = %+v; want DropPkts=1234 PassPkts=42 UniqueIPs=7", gotOrphan)
	}
	if gotFresh == nil {
		t.Fatalf("fresh day %s missing from index", freshDay)
	}

	// The orphan detail page's pre-built HTML must be left alone — the
	// raw log is gone, so there is nothing to rebuild from.
	html, err := os.ReadFile(filepath.Join(srcOut, orphanDay+".html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), "orphan") {
		t.Error("orphan page HTML was rewritten; expected the page to be preserved verbatim")
	}

	// The homepage counter is the sum over the index, so the aged-out day
	// still counts: the visitor's lifetime total is 1234 + 5, and the
	// earliest day — which only the orphan page knows — is the "Since" date.
	page, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	m := heroRe.FindStringSubmatch(string(page))
	if m == nil {
		t.Fatalf("index.html has no #total-attacks counter")
	}
	if m[1] != "1239" || m[2] != "1,239" {
		t.Errorf("hero total = %s / %q, want 1239 / \"1,239\" (orphan day must still count)", m[1], m[2])
	}
	if !strings.Contains(string(page), "Since "+orphanDay) {
		t.Error("Since line does not come from the oldest day, which here is the orphan")
	}
}

// TestSiteDropsUnreadableOrphanPage guards the last-resort behavior: if
// a day has neither a raw log under -logs nor a parseable detail page
// under d/<src>/, the index must skip it rather than emit a zeroed row
// that would read as "no attacks".
func TestSiteDropsUnreadableOrphanPage(t *testing.T) {
	root := t.TempDir()
	logs := filepath.Join(root, "raw")
	out := filepath.Join(root, "web")
	srcLogs := filepath.Join(logs, "server")
	srcOut := filepath.Join(out, "d", "server")
	if err := os.MkdirAll(srcLogs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(srcOut, 0o755); err != nil {
		t.Fatal(err)
	}

	// Unparseable orphan page (e.g. a half-written first run).
	const badDay = siteDay1
	if err := os.WriteFile(filepath.Join(srcOut, badDay+".html"), []byte("garbage"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcOut, badDay+".json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A healthy raw input so the bail-out at "no raw inputs" still kicks in.
	const goodDay = siteDay2
	if err := os.WriteFile(filepath.Join(srcLogs, "nhp_server_xdp-"+goodDay+".log"),
		[]byte("10:00:00 s [NHP-STAT] VERDICT=DROP REASON=TCP_OTHER WINDOW=60s PKTS=5 LOGGED=1 TOTAL=5\n"),
		0o644); err != nil {
		t.Fatal(err)
	}

	runSiteCmd(t, logs, out)

	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err = json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	for _, s := range idx {
		if s.Day == badDay {
			t.Errorf("unparseable orphan %s leaked into index: %+v", badDay, s)
		}
	}
	found := false
	for _, s := range idx {
		if s.Day == goodDay {
			found = true
		}
	}
	if !found {
		t.Errorf("healthy day %s missing from index", goodDay)
	}
}

// heroRe captures the homepage counter's raw data-total and its rendered
// fallback text. The text is what a visitor with JavaScript disabled, an
// older CSP, or prefers-reduced-motion sees, so it has to be the real total,
// not a zero waiting for the animation.
var heroRe = regexp.MustCompile(`id="total-attacks" data-total="(\d+)">([^<]*)</span>`)

// indexTotalFromJSON sums the published index contract. index.html's hero is
// asserted against this so a visitor's number stays reproducible from the
// machine-readable file.
func indexTotalFromJSON(t *testing.T, out string) uint64 {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err := json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	var total uint64
	for _, s := range idx {
		total += s.DropPkts
	}
	return total
}

// TestSiteIndexTotalAndSince covers the homepage hero: the all-time DROP
// packet total, the "Since" line, and the self-hosted counter that animates
// it.
func TestSiteIndexTotalAndSince(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	page, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	want := indexTotalFromJSON(t, out)
	if want != 1750 {
		t.Fatalf("fixture total changed: index.json sums to %d, want 1750", want)
	}
	if got, formatted := groupThousands(want), "1,750"; got != formatted {
		t.Fatalf("groupThousands(%d) = %q, want %q", want, got, formatted)
	}

	m := heroRe.FindStringSubmatch(string(page))
	if m == nil {
		t.Fatalf("index.html has no #total-attacks counter:\n%s", page)
	}
	if m[1] != "1750" {
		t.Errorf("data-total = %s, want 1750", m[1])
	}
	if m[2] != "1,750" {
		t.Errorf("static fallback text = %q, want %q", m[2], "1,750")
	}
	if !strings.Contains(string(page), "Since "+siteDay1) {
		t.Error("index.html is missing the earliest-day \"Since\" line")
	}
	if !strings.Contains(string(page), `src="countup.js"`) {
		t.Error("index.html does not load the self-hosted counter")
	}
	if !strings.Contains(string(page), siteTitle) {
		t.Errorf("index.html is missing the site title %q", siteTitle)
	}
	if _, err := os.Stat(filepath.Join(out, "countup.js")); err != nil {
		t.Errorf("countup.js not written: %v", err)
	}

	// The site's whole point is that nothing is fetched from off-host. The
	// counter is the one script, and it must be ours, not a CDN's.
	for _, name := range []string{"index.html", "countup.js"} {
		b, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		for _, scheme := range []string{"http://", "https://"} {
			if strings.Contains(string(b), scheme) {
				t.Errorf("%s references %s; the site must stay self-contained", name, scheme)
			}
		}
	}
}

// TestSiteIndexNoHeroWhenEmpty asserts the hero is suppressed rather than
// rendered with a zero total and a bogus "Since" date. renderIndex is called
// directly because the CLI refuses to publish a site with no inputs at all,
// so this state is only reachable from a caller with no rows to hand.
func TestSiteIndexNoHeroWhenEmpty(t *testing.T) {
	out := filepath.Join(t.TempDir(), "web")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := renderIndex(&SiteOptions{OutDir: out}, nil); err != nil {
		t.Fatal(err)
	}
	page, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if heroRe.Match(page) {
		t.Error("index.html rendered a hero counter for an empty site")
	}
}

// TestGroupThousands pins the display formatter. Zero and sub-thousand values
// carry no separator, and every other value groups from the right.
func TestGroupThousands(t *testing.T) {
	for _, tc := range []struct {
		in   uint64
		want string
	}{
		{0, "0"},
		{7, "7"},
		{999, "999"},
		{1000, "1,000"},
		{12345, "12,345"},
		{1234567, "1,234,567"},
		{1000000000, "1,000,000,000"},
	} {
		if got := groupThousands(tc.in); got != tc.want {
			t.Errorf("groupThousands(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
	// The template funcs see int and uint64 depending on the Report field.
	if got := group(uint64(1234567)); got != "1,234,567" {
		t.Errorf("group(uint64) = %q", got)
	}
	if got := group(3); got != "3" {
		t.Errorf("group(int) = %q", got)
	}
	if got := group(int64(1234567)); got != "1,234,567" {
		t.Errorf("group(int64) = %q", got)
	}
	if got := group(1234567); got != "1,234,567" {
		t.Errorf("group(int) large = %q", got)
	}
}

// TestSourceLabel pins the display name for a source directory. "server" is
// the directory name the collector stages logs under; the column says
// nhp-server so it names the product rather than repeating a generic noun.
func TestSourceLabel(t *testing.T) {
	if got := sourceLabel("server"); got != "nhp-server" {
		t.Errorf("sourceLabel(server) = %q, want %q", got, "nhp-server")
	}
	// An unmapped source (the AC, once it has a parser) renders as-is rather
	// than blanking the cell.
	if got := sourceLabel("ac"); got != "ac" {
		t.Errorf("sourceLabel(ac) = %q, want %q", got, "ac")
	}
}

// TestSourceColumnRendersLabel asserts the index table shows the label while
// index.json keeps the directory name. The JSON field is a published contract
// — renaming it in the HTML alone would leave the two disagreeing.
func TestSourceColumnRendersLabel(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	page, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "<td>nhp-server</td>") {
		t.Errorf("index.html Source column does not render the nhp-server label")
	}
	idx, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(idx), `"source": "server"`) {
		t.Error("index.json must keep the raw directory name in `source`")
	}
}

// TestTotalNumberJS pins the machine-readable export another site embeds with
// a <script src>: one constant, the same total the homepage hero shows.
func TestTotalNumberJS(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	b, err := os.ReadFile(filepath.Join(out, "totalnumber.js"))
	if err != nil {
		t.Fatalf("totalnumber.js not written: %v", err)
	}
	body := string(b)
	want := indexTotalFromJSON(t, out)
	if !strings.Contains(body, fmt.Sprintf("const NHP_TOTAL_ATTACK_BLOCKED = %d;", want)) {
		t.Errorf("totalnumber.js does not export the index total (%d):\n%s", want, body)
	}
	if !strings.Contains(body, "Counting starts: "+siteDay1) {
		t.Error("totalnumber.js does not record the counting start date")
	}
	// It has to stay executable as a plain classic script.
	for _, scheme := range []string{"http://", "https://"} {
		if strings.Contains(body, scheme) {
			t.Errorf("totalnumber.js references %s; it must stay self-contained", scheme)
		}
	}
}

// TestTotalNumberJSEmptySite asserts the export is still written when nothing
// is published: a consumer's <script src> must not 404 and leave a dangling
// reference, and 0 with no counting start says exactly that.
func TestTotalNumberJSEmptySite(t *testing.T) {
	out := filepath.Join(t.TempDir(), "web")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := renderIndex(&SiteOptions{OutDir: out}, nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(out, "totalnumber.js"))
	if err != nil {
		t.Fatalf("totalnumber.js not written on an empty site: %v", err)
	}
	if !strings.Contains(string(b), "const NHP_TOTAL_ATTACK_BLOCKED = 0;") {
		t.Errorf("empty site total is not 0:\n%s", b)
	}
	if strings.Contains(string(b), "Counting starts:") {
		t.Error("empty site must not claim a counting start date")
	}
}

// TestSinceFloor pins the -since publishing floor. Days older than the floor
// are incomplete (the XDP filter was attached after they were recorded), so
// they are neither rendered nor counted — counting them would silently
// understate the all-time total.
func TestSinceFloor(t *testing.T) {
	root := t.TempDir()
	logs := filepath.Join(root, "raw")
	out := filepath.Join(root, "web")
	src := filepath.Join(logs, "server")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	line := "10:00:00 s [NHP-STAT] VERDICT=DROP REASON=TCP_OTHER WINDOW=60s PKTS=5 LOGGED=1 TOTAL=5\n"
	for _, d := range []string{"2026-10-15", "2026-10-16", "2026-10-17"} {
		if err := os.WriteFile(filepath.Join(src, "nhp_server_xdp-"+d+".log"), []byte(line), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runSiteCmd(t, logs, out) // default -since

	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err = json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	if len(idx) != 2 {
		t.Fatalf("index rows = %d, want 2 (the pre-floor day must be absent)", len(idx))
	}
	for _, s := range idx {
		if s.Day == "2026-10-15" {
			t.Error("day before the -since floor leaked into the index")
		}
	}
	// No page for it either — the floor is a publishing decision, not a
	// rendering filter over an otherwise published day.
	if _, err := os.Stat(filepath.Join(out, "d", "server", "2026-10-15.html")); !os.IsNotExist(err) {
		t.Error("pre-floor day produced a detail page")
	}
	// 2 days x 5 packets. The excluded day would have made it 15.
	if got := indexTotalFromJSON(t, out); got != 10 {
		t.Errorf("total = %d, want 10 (pre-floor day must not be counted)", got)
	}
	page, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "Since 2026-10-16") {
		t.Error("Since line must name the floor, not the oldest input on disk")
	}
}

// TestSinceFloorEmpty asserts a -since that excludes every discovered day is
// refused rather than published: blanking a populated index is the one
// outcome a misconfigured floor must never produce.
func TestSinceFloorEmpty(t *testing.T) {
	logs, out := stageTestdata(t)
	var buf bytes.Buffer
	err := run([]string{"site", "-logs", logs, "-out", out, "-top", "20", "-since", "2099-01-01"}, &buf)
	if err == nil {
		t.Fatal("expected a -since beyond every input day to be refused")
	}
	if !strings.Contains(err.Error(), "-since") {
		t.Errorf("error should name the offending flag: %v", err)
	}
}

// TestSinceFloorOverride asserts -since can move the floor, including
// clearing it to publish everything. This is what a local debug run wants.
func TestSinceFloorOverride(t *testing.T) {
	logs, out := stageTestdata(t)
	// The fixtures are already on/after the floor; clearing it must still
	// publish both days rather than error on an empty filter.
	runSiteCmd(t, logs, out, "-since", "")

	var idx []SiteDaySummary
	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	if len(idx) != 2 {
		t.Errorf("index rows = %d, want 2 with -since cleared", len(idx))
	}

	// A bad -since is a flag error, not a silent no-op.
	var buf bytes.Buffer
	err = run([]string{"site", "-logs", logs, "-out", out, "-since", "16-10-2026"}, &buf)
	if err == nil {
		t.Fatal("expected -since to reject a non-ISO date")
	}
	if !strings.Contains(err.Error(), "-since must be YYYY-MM-DD") {
		t.Errorf("unexpected -since validation error: %v", err)
	}
}

// TestPublishesDay pins the floor comparison itself, including the
// malformed-day escape hatch.
func TestPublishesDay(t *testing.T) {
	o := &SiteOptions{Since: defaultSiteSince}
	for _, tc := range []struct {
		day  string
		want bool
	}{
		{"2026-10-15", false}, // one day before the floor
		{"2026-10-16", true},  // exactly the floor is published
		{"2026-10-17", true},
		{"2026-11-01", true},
		{"2025-01-01", false},
		// A day the parser rejects is kept: a collector bug is worth showing
		// in the index rather than hiding.
		{"not-a-date", true},
	} {
		if got := o.publishesDay(tc.day); got != tc.want {
			t.Errorf("publishesDay(%q) = %v, want %v", tc.day, got, tc.want)
		}
	}
	// An empty floor publishes everything.
	all := &SiteOptions{}
	if !all.publishesDay("2020-01-01") {
		t.Error("an empty -since must publish every day")
	}
}

// TestIndexJSONIsArrayNotNull guards the published contract: index.json is an
// array even when the site is empty. json.Marshal writes a nil slice as
// `null`, and a consumer doing `for (const row of data)` would throw on that
// and read it as a broken deploy rather than "no days yet".
func TestIndexJSONIsArrayNotNull(t *testing.T) {
	out := filepath.Join(t.TempDir(), "web")
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := renderIndex(&SiteOptions{OutDir: out}, nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(b)); got != "[]" {
		t.Errorf("empty index.json = %q, want %q", got, "[]")
	}
}
