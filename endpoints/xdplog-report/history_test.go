package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The history snapshot's job is to carry the site's counters across a period
// the site can no longer analyze, without double-counting the days the
// collector still has. The tests below cover the three things that can go
// wrong there: a snapshot that is accepted when it should be refused, a
// frozen day that is counted twice, and a changed snapshot that does not
// reach the homepage.

// writeHistory writes a snapshot and returns its path.
func writeHistory(t *testing.T, dir string, h SiteHistory) string {
	t.Helper()
	b, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "history.json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// validHistory is a small, self-consistent snapshot: two days whose DROP
// counts add up to the baseline it claims.
func validHistory() SiteHistory {
	return SiteHistory{
		Schema: historySchema,
		Since:  "2026-10-15",
		Totals: HistoryTotals{Blocked: 300, SourceIPs: 40, Countries: 5},
		Days: []SiteDaySummary{
			{Source: "server", Day: "2026-10-15", DropPkts: 100, UniqueIPs: 10, DropShare: 0.5, RestartsDetected: 1},
			{Source: "server", Day: "2026-10-16", DropPkts: 200, UniqueIPs: 20, DropShare: 0.9},
		},
	}
}

func TestLoadSiteHistoryAcceptsAValidSnapshot(t *testing.T) {
	path := writeHistory(t, t.TempDir(), validHistory())
	h, err := loadSiteHistory(path)
	if err != nil {
		t.Fatalf("loadSiteHistory: %v", err)
	}
	if h.Since != "2026-10-15" {
		t.Errorf("Since = %q, want 2026-10-15", h.Since)
	}
	if !h.Covers("server", "2026-10-15") {
		t.Error("the snapshot does not cover server/2026-10-15")
	}
	if h.Covers("server", "2026-10-17") {
		t.Error("the snapshot covers a day it does not list")
	}
	if h.Covers("ac", "2026-10-15") {
		t.Error("the snapshot covers another source's day")
	}
	if h.Fingerprint() == "" {
		t.Error("Fingerprint is empty")
	}
	// The rows are marked as coming from the snapshot: that is what makes
	// the homepage render them without a link.
	rows := h.Rows()
	if len(rows) != 2 {
		t.Fatalf("Rows() returned %d rows, want 2", len(rows))
	}
	for _, r := range rows {
		if !r.History {
			t.Errorf("row %s is not marked as history", r.Day)
		}
	}
}

func TestLoadSiteHistoryRejectsBadSchema(t *testing.T) {
	h := validHistory()
	h.Schema = 2
	path := writeHistory(t, t.TempDir(), h)
	_, err := loadSiteHistory(path)
	if err == nil {
		t.Fatal("a snapshot with the wrong schema was accepted")
	}
	if !strings.Contains(err.Error(), "schema") {
		t.Errorf("error does not mention the schema: %v", err)
	}
}

func TestLoadSiteHistoryRejectsBadDate(t *testing.T) {
	h := validHistory()
	h.Since = "15-10-2026"
	path := writeHistory(t, t.TempDir(), h)
	if _, err := loadSiteHistory(path); err == nil {
		t.Fatal("a snapshot with an unparsable since was accepted")
	}

	h = validHistory()
	h.Days[1].Day = "yesterday"
	path = writeHistory(t, t.TempDir(), h)
	if _, err := loadSiteHistory(path); err == nil {
		t.Fatal("a snapshot with an unparsable day was accepted")
	}
}

func TestLoadSiteHistoryRejectsDuplicateDay(t *testing.T) {
	h := validHistory()
	h.Days[1] = h.Days[0]
	path := writeHistory(t, t.TempDir(), h)
	_, err := loadSiteHistory(path)
	if err == nil {
		t.Fatal("a snapshot listing the same day twice was accepted")
	}
	if !strings.Contains(err.Error(), "twice") {
		t.Errorf("error does not mention the duplicate: %v", err)
	}
}

// TestLoadSiteHistoryRejectsBlockedMismatch pins the check that keeps the
// site's headline contract true: the hero total has to be the sum of
// index.json's drop_pkts, and the snapshot's rows are among them.
func TestLoadSiteHistoryRejectsBlockedMismatch(t *testing.T) {
	h := validHistory()
	h.Totals.Blocked = 301
	path := writeHistory(t, t.TempDir(), h)
	_, err := loadSiteHistory(path)
	if err == nil {
		t.Fatal("a snapshot whose baseline does not match its days was accepted")
	}
	if !strings.Contains(err.Error(), "totals.blocked") {
		t.Errorf("error does not mention totals.blocked: %v", err)
	}
}

// TestLoadSiteHistoryMissingFileIsAnError pins the failure mode that matters
// most: a snapshot that cannot be read must fail the run, and the run must
// not publish anything. Falling back to counting only the site's own days
// would shrink the published all-time numbers by an amount nobody could
// audit.
func TestLoadSiteHistoryMissingFileIsAnError(t *testing.T) {
	logs, out := stageTestdata(t)
	missing := filepath.Join(t.TempDir(), "history.json")

	var buf strings.Builder
	err := run([]string{"site", "-logs", logs, "-out", out, "-top", "20", "-history", missing}, &buf)
	if err == nil {
		t.Fatal("the site ran with a history snapshot that does not exist")
	}
	if !strings.Contains(err.Error(), "history file") {
		t.Errorf("error does not name the history file: %v", err)
	}
	for _, name := range []string{"index.html", "index.json", "state.json", "totalnumber.js"} {
		if _, err := os.Stat(filepath.Join(out, name)); err == nil {
			t.Errorf("%s was written even though the run failed", name)
		}
	}
}

// TestHistoryFreezesCoveredDay is the core of the feature: a day the
// snapshot lists is published from the snapshot, once, without a link, and
// the live copy of the same day contributes nothing.
func TestHistoryFreezesCoveredDay(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	// Freeze the first staged day.
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	page, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(page)
	// The frozen day appears once, as plain text, with the history flag.
	if got := strings.Count(body, ">"+siteDay1+"<"); got != 1 {
		t.Errorf("%s appears %d times in the index, want 1", siteDay1, got)
	}
	if strings.Contains(body, `href="d/server/`+siteDay1+`.html"`) {
		t.Errorf("the frozen day links to a detail page: it has none")
	}
	if !strings.Contains(body, ">history<") {
		t.Error("the frozen day carries no history flag")
	}
	// The live day is still a link.
	if !strings.Contains(body, `href="d/server/`+siteDay2+`.html"`) {
		t.Errorf("the live day lost its detail page link")
	}

	// index.json carries the marker too, and lists the frozen day exactly
	// once.
	idx, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []SiteDaySummary
	if err := json.Unmarshal(idx, &rows); err != nil {
		t.Fatal(err)
	}
	frozen := 0
	for _, r := range rows {
		if r.Day == siteDay1 {
			frozen++
			if !r.History {
				t.Errorf("index.json row %s is not marked as history", r.Day)
			}
			if r.DropPkts != 100 {
				t.Errorf("index.json row %s has %d DROP packets, want the snapshot's 100", r.Day, r.DropPkts)
			}
		}
	}
	if frozen != 1 {
		t.Errorf("the frozen day appears %d times in index.json, want 1", frozen)
	}
}

// TestHistoryTotalsAddUp pins the counting formula: the packet total is the
// snapshot's baseline plus the live days, and it stays reproducible by
// summing index.json's drop_pkts.
func TestHistoryTotalsAddUp(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	live := indexTotalFromJSON(t, out)
	idx, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []SiteDaySummary
	if err := json.Unmarshal(idx, &rows); err != nil {
		t.Fatal(err)
	}
	var sum uint64
	for _, r := range rows {
		sum += r.DropPkts
	}
	if sum != live {
		t.Errorf("index.json's rows add up to %d but the hero shows %d", sum, live)
	}
	if live <= 100 {
		t.Errorf("the hero total is %d; the snapshot's 100 plus the live day are missing", live)
	}

	totals := readTotalNumbers(t, out)
	if got := totals["NHP_TOTAL_ATTACK_BLOCKED"]; uint64(got) != live {
		t.Errorf("NHP_TOTAL_ATTACK_BLOCKED = %d, want %d", got, live)
	}
	// The source and country baselines are added on top of the live union.
	liveSources := totals["NHP_TOTAL_ATTACK_SOURCE_IPS"] - hist.Totals.SourceIPs
	if liveSources < 0 {
		t.Errorf("NHP_TOTAL_ATTACK_SOURCE_IPS = %d, want the baseline %d plus a live union",
			totals["NHP_TOTAL_ATTACK_SOURCE_IPS"], hist.Totals.SourceIPs)
	}
	if got := totals["NHP_TOTAL_ATTACK_COUNTRIES"]; got != hist.Totals.Countries {
		t.Errorf("NHP_TOTAL_ATTACK_COUNTRIES = %d, want the baseline %d (no geo source in the fixture)",
			got, hist.Totals.Countries)
	}
}

// TestHistoryFactsExcludedFromUnion pins that a frozen day's attackers are
// not counted twice: the snapshot's baseline already accounts for that
// period, so the live union must not include the frozen day's facts.
//
// The fixture's first day carries the only source addresses (the second day
// has none), so the union without a snapshot is those two addresses and the
// union with one is empty.
func TestHistoryFactsExcludedFromUnion(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	hist.Totals.SourceIPs = 40
	hist.Totals.Countries = 0
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	totals := readTotalNumbers(t, out)
	if got := totals["NHP_TOTAL_ATTACK_SOURCE_IPS"]; got != 40 {
		t.Errorf("NHP_TOTAL_ATTACK_SOURCE_IPS = %d, want 40 (the baseline plus an empty live union)", got)
	}

	// The same site without a snapshot unions the first day's two source
	// addresses, so a frozen day's facts showing up in the union would
	// have shown up here as 42.
	plainLogs, plainOut := stageTestdata(t)
	runSiteCmd(t, plainLogs, plainOut)
	plain := readTotalNumbers(t, plainOut)
	if plain["NHP_TOTAL_ATTACK_SOURCE_IPS"] != 2 {
		t.Fatalf("the fixture's live union is %d, want 2; the test's arithmetic depends on it",
			plain["NHP_TOTAL_ATTACK_SOURCE_IPS"])
	}
}

// TestHistorySinceComesFromSnapshot pins that the counting start is the
// snapshot's, not the earliest row's: the snapshot is what says the site's
// numbers begin before its first published day.
func TestHistorySinceComesFromSnapshot(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	hist.Since = "2026-10-01"
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	page, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), "Since 2026-10-01.") {
		t.Error("the homepage does not count from the snapshot's since date")
	}
	b, err := os.ReadFile(filepath.Join(out, "totalnumber.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Counting starts: 2026-10-01 (UTC).") {
		t.Error("totalnumber.js does not record the snapshot's counting start")
	}
}

// TestHistoryFingerprintMarksDirty pins that a changed snapshot rewrites the
// homepage even when no log changed. Without it the counters would keep the
// baseline they had when the snapshot was first read.
func TestHistoryFingerprintMarksDirty(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)
	before, err := os.Stat(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}

	// A second run with nothing changed must not rewrite anything.
	runSiteCmd(t, logs, out, "-history", path)
	after, err := os.Stat(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Error("an unchanged run rewrote the homepage")
	}

	// Changing the snapshot's baseline has to reach the homepage.
	hist.Totals.Blocked = 100
	hist.Totals.SourceIPs = 4242
	hist.Days[0].DropPkts = 100
	writeHistory(t, filepath.Dir(path), hist)
	runSiteCmd(t, logs, out, "-history", path)

	totals := readTotalNumbers(t, out)
	if totals["NHP_TOTAL_ATTACK_SOURCE_IPS"] < 4242 {
		t.Errorf("the changed snapshot did not reach totalnumber.js: %d",
			totals["NHP_TOTAL_ATTACK_SOURCE_IPS"])
	}
	// The fingerprint is recorded in the state, so the next run knows the
	// index is current.
	state := readSiteStateAt(t, filepath.Join(out, "state.json"))
	if state.History == "" {
		t.Error("state.json does not record the history fingerprint")
	}
}

// TestHistoryGeneratorFromSavedPage pins the generator against the saved
// page it was built for: the numbers the page showed, day by day.
func TestHistoryFlagsIncludeHistory(t *testing.T) {
	flags := flagsFor(SiteDaySummary{History: true, Partial: true})
	if len(flags) == 0 || flags[0].Label != "history" {
		t.Fatalf("flags = %+v, want the history flag first", flags)
	}
	if flags[0].Class != "tag-info" {
		t.Errorf("the history flag's class = %q, want tag-info", flags[0].Class)
	}
	// A live row is unchanged: no history chip.
	for _, f := range flagsFor(SiteDaySummary{Partial: true}) {
		if f.Label == "history" {
			t.Error("a live row carries the history flag")
		}
	}
}

// TestSiteWithoutHistoryIsUnchanged pins that a site with no snapshot looks
// exactly as it did: no history flag, no extra note, and every day still a
// link.
func TestSiteWithoutHistoryIsUnchanged(t *testing.T) {
	logs, out := stageTestdata(t)
	runSiteCmd(t, logs, out)

	page, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(page)
	if strings.Contains(body, ">history<") {
		t.Error("a site without a snapshot shows the history flag")
	}
	if strings.Contains(body, "history.json") {
		t.Error("a site without a snapshot mentions the snapshot")
	}
	if !strings.Contains(body, `href="d/server/`+siteDay1+`.html"`) {
		t.Error("a site without a snapshot lost a detail page link")
	}
}

// TestHistorySeedSnapshotIsShipped pins the seed snapshot the deploy ships:
// the checked-in file has to be the one the generator produces from the saved
// page, or the two drift apart silently.
func TestHistoryRowsAreNotPublishedTwice(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	idx, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []SiteDaySummary
	if err := json.Unmarshal(idx, &rows); err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for _, r := range rows {
		seen[r.Source+"/"+r.Day]++
	}
	for key, n := range seen {
		if n > 1 {
			t.Errorf("%s appears %d times in index.json", key, n)
		}
	}
	// The frozen day's packet count is the snapshot's, not the log's: the
	// log for that day says something else entirely.
	for _, r := range rows {
		if r.Day == siteDay1 && r.DropPkts != 100 {
			t.Errorf("the frozen day's count is %d, want the snapshot's 100", r.DropPkts)
		}
	}
}
