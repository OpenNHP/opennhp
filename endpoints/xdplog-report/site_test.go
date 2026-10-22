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

	// The orphan day cannot be re-analyzed — the raw log is gone — but its
	// HTML is a pure function of the JSON beside it, so it adopts the
	// current template with its own numbers intact. That is what keeps a
	// title change from leaving the oldest pages on markup nobody runs.
	html, err := os.ReadFile(filepath.Join(srcOut, orphanDay+".html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(html), "orphan") {
		t.Error("orphan page HTML still the hand-written fixture; expected a re-render")
	}
	if !strings.Contains(string(html), "Daily Attack Report") {
		t.Error("orphan page HTML was not re-rendered with the current template")
	}
	if !strings.Contains(string(html), "1,234 hostile packets from 7 sources") {
		t.Errorf("orphan page lost its own numbers:\n%s", string(html))
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

// TestSiteKeepsPageOnlyDayWhenRerenderFails covers the page-only error the
// orphan rule must not swallow: the day's JSON is perfectly readable, but its
// HTML cannot be written — an EACCES, an ENOSPC, a failed rename. The record is
// intact, so the day keeps its index row and its share of the hero total and
// the run fails loudly afterwards. Treating this as an orphan is what made the
// published total quietly fall whenever the web root was briefly unwritable,
// and jump back on the next good run with nothing having flagged the gap.
func TestSiteKeepsPageOnlyDayWhenRerenderFails(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)
	wantTotal := indexTotalFromJSON(t, out)

	// Age the day out of the collector cache and roll the state's render
	// version back, so its page is due a re-render.
	if err := os.Remove(filepath.Join(logs, "server", "nhp_server_xdp-"+siteDay2+".log")); err != nil {
		t.Fatalf("prune the raw log: %v", err)
	}
	setStateRender(t, out, 0)

	// Put a directory where the page goes: writeFileAtomic renames a sibling
	// temp file over the path, and that rename never succeeds onto a
	// directory. Preferred over chmod because it still fails when the suite
	// runs as root.
	pagePath := filepath.Join(out, "d", "server", siteDay2+".html")
	if err := os.Remove(pagePath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(pagePath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pagePath, "blocker"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := run([]string{"site", "-logs", logs, "-out", out, "-top", "20"}, &buf)
	if err == nil {
		t.Fatal("expected the failed re-render to fail the run")
	}
	if !strings.Contains(err.Error(), siteDay2) {
		t.Errorf("error should name the day whose page could not be written: %v", err)
	}
	if !indexHasDay(t, out, siteDay2) {
		t.Errorf("page-only day %s dropped out of the index because its HTML could not be re-rendered", siteDay2)
	}
	if got := indexTotalFromJSON(t, out); got != wantTotal {
		t.Errorf("published total = %d, want %d; a write failure must not shrink it", got, wantTotal)
	}
	page := readArtifact(t, out, "index.html")
	if m := heroRe.FindStringSubmatch(string(page)); m == nil {
		t.Error("index.html has no #total-attacks counter")
	} else if m[2] != groupThousands(wantTotal) {
		t.Errorf("hero total = %q, want %q", m[2], groupThousands(wantTotal))
	}
	// The state must not claim a re-render that never happened, or the next
	// run would skip the retry.
	if got := readSiteState(t, out).Days["server"][siteDay2].Render; got == siteRenderVersion {
		t.Error("state recorded a successful re-render that never happened")
	}

	// Clear the obstruction. The next run re-renders the day it kept, and the
	// total does not move.
	if err := os.RemoveAll(pagePath); err != nil {
		t.Fatal(err)
	}
	runSiteCmd(t, logs, out)
	if b := readArtifact(t, filepath.Join(out, "d", "server"), siteDay2+".html"); !strings.Contains(string(b), "Daily Attack Report") {
		t.Errorf("page-only day was not re-rendered once the path was writable:\n%s", b)
	}
	if got := indexTotalFromJSON(t, out); got != wantTotal {
		t.Errorf("published total after recovery = %d, want %d", got, wantTotal)
	}
}

// TestFailedRunLeavesPreFloorPagesPublished is the ordering half of the -since
// floor: a pre-floor page comes down only after every publishable day has
// succeeded, never while the loop is still working. A raw log that cannot be
// analyzed must not cost the site what it is already serving — the index on
// disk still links those pages and still counts them in the hero total, so
// deleting them on the way out would leave dead links and a total that no
// longer matches the files behind it.
func TestFailedRunLeavesPreFloorPagesPublished(t *testing.T) {
	root := t.TempDir()
	logs := filepath.Join(root, "raw")
	out := filepath.Join(root, "web")

	const preFloor = "2026-10-15"
	writeDropLog(t, logs, "server", preFloor, 7)
	writeDropLog(t, logs, "server", siteDay1, 5)
	runSiteCmd(t, logs, out, "-since", "") // the site as it looked before the floor

	preHTML := filepath.Join(out, "d", "server", preFloor+".html")
	preJSON := filepath.Join(out, "d", "server", preFloor+".json")
	for _, p := range []string{preHTML, preJSON} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("fixture failed to publish %s: %v", p, err)
		}
	}
	wantTotal := indexTotalFromJSON(t, out)

	// A .gz that is not gzip: undiscoverable by content, unreadable by the
	// analyzer. Its day is on and after the floor, so the run reaches it.
	badPath := filepath.Join(logs, "server", "nhp_server_xdp-"+siteDay2+".log.gz")
	if err := os.WriteFile(badPath, []byte("this is not a gzip stream\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := run([]string{"site", "-logs", logs, "-out", out, "-top", "20"}, &buf)
	if err == nil {
		t.Fatal("expected the unreadable log to fail the run")
	}
	if !strings.Contains(err.Error(), siteDay2) {
		t.Errorf("error should name the day that failed: %v", err)
	}
	for _, p := range []string{preHTML, preJSON} {
		if _, statErr := os.Stat(p); statErr != nil {
			t.Errorf("failed run deleted the pre-floor page %s: %v", p, statErr)
		}
	}
	if !indexHasDay(t, out, preFloor) {
		t.Error("pre-floor day left the index while its page is still being served")
	}
	if got := indexTotalFromJSON(t, out); got != wantTotal {
		t.Errorf("published total = %d, want %d; a failed run must leave the site as it was", got, wantTotal)
	}
	if !strings.Contains(string(readArtifact(t, out, "index.html")), `href="d/server/`+preFloor+`.html"`) {
		t.Error("index no longer links the pre-floor day whose page is still published")
	}

	// The floor still applies on the next good run: with the unreadable log
	// gone, the pre-floor day comes down and the index stops linking it.
	if err := os.Remove(badPath); err != nil {
		t.Fatal(err)
	}
	runSiteCmd(t, logs, out)
	for _, p := range []string{preHTML, preJSON} {
		if _, statErr := os.Stat(p); !os.IsNotExist(statErr) {
			t.Errorf("%s survived the -since floor (stat err = %v)", p, statErr)
		}
	}
	if indexHasDay(t, out, preFloor) {
		t.Error("pre-floor day is still indexed after a successful run")
	}
}

// heroRe captures the homepage counter's raw data-total and its rendered
// fallback text. The text is what a visitor with JavaScript disabled, an
// older CSP, or prefers-reduced-motion sees, so it has to be the real total,
// not a zero waiting for the animation.
var heroRe = regexp.MustCompile(`id="total-attacks" data-total="(\d+)">([^<]*)</span>`)

// absoluteURLRe matches every absolute URL in a published file.
var absoluteURLRe = regexp.MustCompile(`https?://[^"'[:space:]>]+`)

// scriptSrcRe matches the src of every <script> tag — the one construct the
// CSP allows and the only way an injected script would run. canonical and
// og:* URLs are deliberately not matched: a crawler reads those, a browser
// does not fetch them.
var scriptSrcRe = regexp.MustCompile(`(?i)<script[^>]*\ssrc="([^"]+)"`)

// externalStyleRe matches a stylesheet link. The site styles itself with a
// single inline <style>, and the CSP grants style-src 'unsafe-inline' with no
// host, so any such link would be blocked anyway.
var externalStyleRe = regexp.MustCompile(`(?i)<link[^>]*\srel="?stylesheet"?`)

// readArtifact reads one published file, failing the test if it is missing.
func readArtifact(t *testing.T, out, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(out, name))
	if err != nil {
		t.Fatalf("%s not written: %v", name, err)
	}
	return b
}

// assertSelfContained fails if b references any host other than the site's own
// origin. A page whose only absolute URLs are its own is still offline-buildable
// and still has nothing to go and fetch; a CDN would break both the CSP and the
// promise, and an attacker who could inject one would get a beacon for free.
func assertSelfContained(t *testing.T, name string, b []byte) {
	t.Helper()
	if err := selfContainedError(string(b)); err != nil {
		t.Errorf("%s: %v", name, err)
	}
}

// selfContainedError returns the first absolute URL in s that is not the site's
// own origin, or nil. Split out from the t.Errorf so the rule can be tested on
// both sides of the line — including the lookalike host, which a plain
// "does it contain https://" check would wave through.
func selfContainedError(s string) error {
	for _, u := range absoluteURLRe.FindAllString(s, -1) {
		if !strings.HasPrefix(u, siteURL) {
			return fmt.Errorf("references %s; the site must stay self-contained", u)
		}
	}
	return nil
}

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
	if !strings.Contains(string(page), `src="./totalnumber.js"`) {
		t.Error("index.html does not load the exported total")
	}
	if !strings.Contains(string(page), siteTitle) {
		t.Errorf("index.html is missing the site title %q", siteTitle)
	}
	if _, err := os.Stat(filepath.Join(out, "countup.js")); err != nil {
		t.Errorf("countup.js not written: %v", err)
	}

	// The site's whole point is that nothing is fetched from off-host. Both
	// scripts are ours, not a CDN's, and the canonical/og URLs that do name an
	// origin name ours.
	for _, name := range []string{"index.html", "countup.js"} {
		assertSelfContained(t, name, readArtifact(t, out, name))
	}
	// countup.js runs on every page view and has no metadata reason to know
	// the site's name, so it may not contain an absolute URL at all. Same for
	// totalnumber.js, which other sites embed by URL.
	for _, name := range []string{"countup.js", "totalnumber.js"} {
		if b := readArtifact(t, out, name); absoluteURLRe.Match(b) {
			t.Errorf("%s contains an absolute URL:\n%s", name, b)
		}
	}
	// Everything the browser is told to fetch has to be a relative path, so it
	// resolves against whatever origin serves the page.
	for _, ref := range scriptSrcRe.FindAllStringSubmatch(string(page), -1) {
		if strings.Contains(ref[1], "://") {
			t.Errorf("index.html makes the browser fetch the absolute URL %q", ref[1])
		}
	}
	if externalStyleRe.Match(page) {
		t.Error("index.html links an external stylesheet; the site styles itself inline")
	}
}

// TestAssertSelfContained pins the guard itself. A self-containment check that
// cannot fail is not a check, and this one has to allow our own origin in
// metadata while still refusing a CDN — so exercise both sides of the line
// rather than only the side the real output happens to fall on.
func TestAssertSelfContained(t *testing.T) {
	for _, tc := range []struct {
		name    string
		page    string
		wantErr bool
	}{
		{"own origin in canonical", `<link rel="canonical" href="https://stats.opennhp.org/">`, false},
		{"own origin as og:image", `<meta property="og:image" content="` + siteOGImage + `">`, false},
		{"no absolute URL at all", `<script src="countup.js" defer></script>`, false},
		{"cdn script", `<script src="https://cdn.jsdelivr.net/npm/countup.js"></script>`, true},
		{"third-party analytics", `<img src="http://tracker.example/pixel.gif">`, true},
		{"lookalike host", `<link rel="canonical" href="https://stats.opennhp.org.evil.test/">`, true},
		{"http downgrade", `<link rel="canonical" href="http://stats.opennhp.org/">`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := selfContainedError(tc.page)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("selfContainedError(%s) error = %v, want error: %v", tc.page, err, tc.wantErr)
			}
		})
	}
}

// TestSiteIndexHeadMetadata pins the <head> copy: the <title> search engines
// and chat previews show, the description, and the Open Graph / Twitter card
// tags. It also pins that the <title> is deliberately NOT the <h1>: a tab has
// room to say what the site is about, the heading above the fold does not, and
// collapsing them would lose one of the two.
func TestSiteIndexHeadMetadata(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	page, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	head := string(page)
	if i := strings.Index(head, "</head>"); i >= 0 {
		head = head[:i]
	}

	for _, want := range []string{
		"<title>" + siteDocumentTitle + "</title>",
		`<meta name="description" content="` + siteDescription + `">`,
		`<link rel="canonical" href="` + siteURL + `">`,
		`<meta property="og:title" content="` + siteTitle + `">`,
		`<meta property="og:description" content="` + siteOGDescription + `">`,
		`<meta property="og:url" content="` + siteURL + `">`,
		`<meta property="og:type" content="website">`,
		`<meta property="og:image" content="` + siteOGImage + `">`,
		`<meta name="twitter:card" content="summary_large_image">`,
	} {
		if !strings.Contains(head, want) {
			t.Errorf("index.html <head> is missing:\n  %s", want)
		}
	}
	if strings.Contains(head, "<title>"+siteTitle+"</title>") {
		t.Error("the <title> collapsed back onto the <h1> slogan")
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
	if _, statErr := os.Stat(filepath.Join(out, "d", "server", "2026-10-15.html")); !os.IsNotExist(statErr) {
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
	// Publish both fixture days first, so the refusal below has a populated
	// site to protect.
	runSiteCmd(t, logs, out, "-since", "")
	var buf bytes.Buffer
	err := run([]string{"site", "-logs", logs, "-out", out, "-top", "20", "-since", "2099-01-01"}, &buf)
	if err == nil {
		t.Fatal("expected a -since beyond every input day to be refused")
	}
	if !strings.Contains(err.Error(), "-since") {
		t.Errorf("error should name the offending flag: %v", err)
	}
	// The refusal has to come before the unpublish pass: a floor that
	// excludes everything must not delete every detail page on its way to
	// the same empty index.
	for _, day := range []string{siteDay1, siteDay2} {
		if _, statErr := os.Stat(filepath.Join(out, "d", "server", day+".html")); statErr != nil {
			t.Errorf("%s page removed by a refused run: %v", day, statErr)
		}
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

// writeDropLog stages a minimal single-day log with an exact DROP count. Used
// where the fixture content is irrelevant and only the day's presence and
// totals matter.
func writeDropLog(t *testing.T, logs, source, day string, pkts int) {
	t.Helper()
	dir := filepath.Join(logs, source)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	line := fmt.Sprintf("10:00:00 s [NHP-STAT] VERDICT=DROP REASON=TCP_OTHER WINDOW=60s PKTS=%d LOGGED=1 TOTAL=%d\n", pkts, pkts)
	if err := os.WriteFile(filepath.Join(dir, "nhp_server_xdp-"+day+".log"), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
}

// readSiteState loads <out>/state.json, which is where the tests keep it
// because runSiteCmd passes no -state.
func readSiteState(t *testing.T, out string) SiteState {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(out, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var st SiteState
	if err := json.Unmarshal(b, &st); err != nil {
		t.Fatal(err)
	}
	return st
}

// setStateRender rewrites every day entry's render version. Version 0 stands
// in for a state.json written by a binary that did not record the field: the
// key is simply absent and decodes to zero.
func setStateRender(t *testing.T, out string, v int) {
	t.Helper()
	st := readSiteState(t, out)
	for src, days := range st.Days {
		for day, d := range days {
			d.Render = v
			st.Days[src][day] = d
		}
	}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, '\n')
	if err := os.WriteFile(filepath.Join(out, "state.json"), b, 0o644); err != nil {
		t.Fatal(err)
	}
}

// indexHasDay reports whether the published index still lists day.
func indexHasDay(t *testing.T, out, day string) bool {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err := json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	for _, s := range idx {
		if s.Day == day {
			return true
		}
	}
	return false
}

// TestSiteUnpublishesPreFloorPages covers the migration the floor implies on a
// site that is already serving the day. stats.opennhp.org published days
// before 2026-10-16 before the floor existed, and merely skipping them leaves
// their detail pages reachable at their old URLs — unlinked, but still public
// and still carrying the partial record the floor exists to withhold. The run
// has to take them back down.
func TestSiteUnpublishesPreFloorPages(t *testing.T) {
	root := t.TempDir()
	logs := filepath.Join(root, "raw")
	out := filepath.Join(root, "web")

	const preFloor = "2026-10-15"
	writeDropLog(t, logs, "server", preFloor, 7)
	writeDropLog(t, logs, "server", siteDay1, 5)
	writeDropLog(t, logs, "server", siteDay2, 5)

	// Before the floor existed the day published like any other.
	runSiteCmd(t, logs, out, "-since", "")
	preHTML := filepath.Join(out, "d", "server", preFloor+".html")
	preJSON := filepath.Join(out, "d", "server", preFloor+".json")
	for _, p := range []string{preHTML, preJSON} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("fixture failed to publish %s: %v", p, err)
		}
	}

	// The deploy that introduces the floor runs with the default.
	runSiteCmd(t, logs, out)

	for _, p := range []string{preHTML, preJSON} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s survived the -since floor (stat err = %v)", p, err)
		}
	}
	if indexHasDay(t, out, preFloor) {
		t.Errorf("pre-floor day %s is still indexed", preFloor)
	}
	// 2 days x 5 packets. The unpublished day would have made it 17.
	if got := indexTotalFromJSON(t, out); got != 10 {
		t.Errorf("total = %d, want 10 (pre-floor day must not be counted)", got)
	}
	if _, ok := readSiteState(t, out).Days["server"][preFloor]; ok {
		t.Errorf("state.json still carries a fingerprint for the unpublished day %s", preFloor)
	}

	// Every later run must stay a no-op: the page is already gone, and an
	// absent file is not an error.
	runSiteCmd(t, logs, out)
	if _, err := os.Stat(preHTML); !os.IsNotExist(err) {
		t.Errorf("%s came back on the second run (stat err = %v)", preHTML, err)
	}
}

// TestSiteRenderVersionRebuildsUnchangedDay covers the deploy-time half of the
// same story. A markup-only change moves siteRenderVersion, and the input
// fingerprint does not move with it — so without the version in state.json a
// title change would reach only the days whose logs happen to change that
// week, and every other published page would keep the old title forever.
func TestSiteRenderVersionRebuildsUnchangedDay(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	day1 := filepath.Join(out, "d", "server", siteDay1+".html")
	day2 := filepath.Join(out, "d", "server", siteDay2+".html")
	before1, err := os.Stat(day1)
	if err != nil {
		t.Fatal(err)
	}
	before2, err := os.Stat(day2)
	if err != nil {
		t.Fatal(err)
	}

	// No input changed. This is the state an earlier binary leaves behind:
	// matching fingerprints, no render version recorded.
	setStateRender(t, out, 0)
	runSiteCmd(t, logs, out)

	after1, err := os.Stat(day1)
	if err != nil {
		t.Fatal(err)
	}
	after2, err := os.Stat(day2)
	if err != nil {
		t.Fatal(err)
	}
	if after1.ModTime().UnixNano() == before1.ModTime().UnixNano() {
		t.Error("day 1 page was not rebuilt after the template version moved")
	}
	if after2.ModTime().UnixNano() == before2.ModTime().UnixNano() {
		t.Error("day 2 page was not rebuilt after the template version moved")
	}

	b, err := os.ReadFile(day1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Daily Attack Report") {
		t.Errorf("rebuilt day page is not on the current template:\n%s", string(b))
	}
	if got := readSiteState(t, out).Days["server"][siteDay1].Render; got != siteRenderVersion {
		t.Errorf("state render version = %d, want %d", got, siteRenderVersion)
	}
}

// TestSitePageOnlyDayRerendersOnce covers the days whose raw log the collector
// already pruned. They cannot be re-analyzed, but their HTML is a pure
// function of the published JSON beside them, so the template still reaches
// them — once, and then the state keeps every later run quiet.
func TestSitePageOnlyDayRerendersOnce(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	dayHTML := filepath.Join(out, "d", "server", siteDay2+".html")
	if err := os.Remove(filepath.Join(logs, "server", "nhp_server_xdp-"+siteDay2+".log")); err != nil {
		t.Fatalf("prune the raw log: %v", err)
	}
	setStateRender(t, out, 0)

	runSiteCmd(t, logs, out)

	first, err := os.Stat(dayHTML)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(dayHTML)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Daily Attack Report") {
		t.Errorf("page-only day was not re-rendered with the current template:\n%s", string(b))
	}
	// Its numbers still come from the JSON, not from a rebuilt-from-nothing
	// zero row.
	if !indexHasDay(t, out, siteDay2) {
		t.Error("page-only day dropped out of the index")
	}
	if idx := readSiteState(t, out); idx.Days["server"][siteDay2].Render != siteRenderVersion {
		t.Error("state did not record the render version for the page-only day")
	}

	runSiteCmd(t, logs, out)
	second, err := os.Stat(dayHTML)
	if err != nil {
		t.Fatal(err)
	}
	if second.ModTime().UnixNano() != first.ModTime().UnixNano() {
		t.Error("page-only day was re-rendered again on an unchanged run")
	}
}
