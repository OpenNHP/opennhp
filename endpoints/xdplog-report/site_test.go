package main

import (
	"bytes"
	"encoding/json"
	"net/netip"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
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
	for _, d := range []string{"2026-10-01", "2026-10-02"} {
		b, err := os.ReadFile(filepath.Join("testdata", "nhp_server_xdp-"+d+".log"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(src, "nhp_server_xdp-"+d+".log"), b, 0o644); err != nil {
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
// index, state, and every detail page under d/<source>/<day>.{html,json}.
func indexAndPages(out string) []string {
	files := []string{
		filepath.Join(out, "index.html"),
		filepath.Join(out, "index.json"),
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

	for _, day := range []string{"2026-10-01", "2026-10-02"} {
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

func TestSiteIndexOrderDesc(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err := json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	if len(idx) != 2 {
		t.Fatalf("index rows = %d, want 2", len(idx))
	}
	if idx[0].Day != "2026-10-02" || idx[1].Day != "2026-10-01" {
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

	day1 := filepath.Join(out, "d", "server", "2026-10-01.html")
	day2 := filepath.Join(out, "d", "server", "2026-10-02.html")
	st1, _ := os.Stat(day1)
	st2, _ := os.Stat(day2)

	// Grow day1 input so its size+mtime change.
	src := filepath.Join(logs, "server", "nhp_server_xdp-2026-10-01.log")
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
		filepath.Join(out, "d", "server", "2026-10-01.html"),
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
	detail, err := os.ReadFile(filepath.Join(out, "d", "server", "2026-10-01.html"))
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
		"nhp_server_xdp-2026-1-1.log",     // bad date
		"nhp_server_xdp-20261001.log",     // no dashes
		"nhp_server_xdp-2026-10-01.txt",   // wrong extension
		"other_server_xdp-2026-10-03.log", // wrong prefix
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
	if err := json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, s := range idx {
		got[s.Day] = true
	}
	for _, d := range []string{"2026-10-01", "2026-10-02"} {
		if !got[d] {
			t.Errorf("missing valid day %s", d)
		}
	}
	if got["2026-10-03"] {
		t.Error("rogue day 2026-10-03 leaked into index")
	}
	if _, err := os.Stat(filepath.Join(out, "d", "server", "2026-10-01.txt.html")); err == nil {
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
	if err := json.Unmarshal(b, &idx); err != nil {
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
}

// TestSiteGeoCSVPassesThrough exercises the geo path end-to-end through the
// site subcommand, verifying that geoEnabled flows into the per-day JSON.
func TestSiteGeoCSVPassesThrough(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out, "-geo-csv", "testdata/geo.csv")

	b, err := os.ReadFile(filepath.Join(out, "d", "server", "2026-10-01.json"))
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

// netip sanity so we can reference netip in case future tests use it
var _ = netip.MustParseAddr("203.0.113.5")
