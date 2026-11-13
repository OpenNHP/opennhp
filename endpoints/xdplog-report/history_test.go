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
		Totals: HistoryTotals{Blocked: 300, SourceIPs: 40, Regions: 5},
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

// TestHistoryLiveDaySupersedesTheSnapshot is the rule that keeps the
// snapshot from double-counting: a day the live site can produce is analyzed
// from the log, and the snapshot's row for it is dropped. The published
// numbers are then the live ones -- with a real detail page to click, and the
// day's attackers in the all-time union.
func TestHistoryLiveDaySupersedesTheSnapshot(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	body := string(readArtifact(t, out, "index.html"))
	// The day appears once, and it is a link: the live site produced it.
	if got := strings.Count(body, ">"+siteDay1+"<"); got != 1 {
		t.Errorf("%s appears %d times in the index, want 1", siteDay1, got)
	}
	if !strings.Contains(body, `href="d/server/`+siteDay1+`.html"`) {
		t.Error("the day the live site produced lost its detail page link")
	}
	if strings.Contains(body, ">history<") {
		t.Error("a day the live site produced carries the history flag")
	}
	// The live day is still a link.
	if !strings.Contains(body, `href="d/server/`+siteDay2+`.html"`) {
		t.Errorf("the live day lost its detail page link")
	}
}

// TestHistoryFreezesADayTheLiveSiteCannotProduce pins the case the snapshot
// exists for: the collector no longer has the day's log, so the snapshot's
// row is the only record of it. It is published from the snapshot -- as plain
// text, with no detail page to click, and with the snapshot's own numbers.
func TestHistoryFreezesADayTheLiveSiteCannotProduce(t *testing.T) {
	logs, out := stageTestdata(t)
	const frozenDay = "2026-10-18"
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = frozenDay
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	body := string(readArtifact(t, out, "index.html"))
	// The frozen day appears once, as plain text, with the history flag.
	if got := strings.Count(body, ">"+frozenDay+"<"); got != 1 {
		t.Errorf("%s appears %d times in the index, want 1", frozenDay, got)
	}
	if strings.Contains(body, `href="d/server/`+frozenDay+`.html"`) {
		t.Error("the frozen day links to a detail page: it has none")
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
	idx := readArtifact(t, out, "index.json")
	var rows []SiteDaySummary
	if err := json.Unmarshal(idx, &rows); err != nil {
		t.Fatal(err)
	}
	frozen := 0
	for _, r := range rows {
		if r.Day == frozenDay {
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
	// The source count is the union over the published days -- not the
	// snapshot's baseline added on top of it. Two unions can only be merged
	// when both sets are present; adding their sizes counts an address seen
	// in both periods twice.
	plainLogs, plainOut := stageTestdata(t)
	runSiteCmd(t, plainLogs, plainOut)
	plain := readTotalNumbers(t, plainOut)
	if got := totals["NHP_TOTAL_ATTACK_SOURCE_IPS"]; got != plain["NHP_TOTAL_ATTACK_SOURCE_IPS"] {
		t.Errorf("NHP_TOTAL_ATTACK_SOURCE_IPS = %d, want %d (the union of the published days; the snapshot's baseline is not added on top)",
			got, plain["NHP_TOTAL_ATTACK_SOURCE_IPS"])
	}
	// The region baseline is not added either: the published figure is the
	// union over the published days, and this fixture has no geo source, so
	// it is zero.
	if got := totals["NHP_TOTAL_ATTACK_REGIONS"]; got != 0 {
		t.Errorf("NHP_TOTAL_ATTACK_REGIONS = %d, want 0 (the union of the published days; no geo source in the fixture)",
			got)
	}
}

// TestHistoryRegionTotalUnionsTheSnapshotDays pins the region total's
// source of truth: the union over every published day, the snapshot's days
// included.
//
// The snapshot's own region count is not added on top of it. A region seen
// in both periods would then be counted twice, and the total would exceed
// the number of regions there are -- and the snapshot's days' regions
// are already in the union, because those days are analyzed from the logs
// when the collector still has them.
func TestHistoryRegionTotalUnionsTheSnapshotDays(t *testing.T) {
	root := t.TempDir()
	logs := filepath.Join(root, "raw")
	out := filepath.Join(root, "web")
	statePath := filepath.Join(root, "state", "state.json")

	writeDropLogWithSources(t, logs, "server", siteDay1, 100,
		"203.0.113.5", "198.51.100.7", "10.0.0.9")
	writeDropLogWithSources(t, logs, "server", siteDay2, 200,
		"203.0.113.5", "192.0.2.9")

	// A snapshot with a region count of its own. Its day is one the live
	// site also has, so the day is analyzed from the log and its regions
	// join the union.
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	hist.Totals.SourceIPs = 40
	hist.Totals.Regions = 130
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-state", statePath, "-geo-csv", "testdata/geo.csv", "-history", path)

	totals := readTotalNumbers(t, out)
	// The source count is the union of the two days' addresses:
	// 203.0.113.5, 198.51.100.7 and 192.0.2.9 (10.0.0.9 is a private
	// address and is not counted). The snapshot's own 40 is not added on top
	// of it.
	if got := totals["NHP_TOTAL_ATTACK_SOURCE_IPS"]; got != 3 {
		t.Errorf("NHP_TOTAL_ATTACK_SOURCE_IPS = %d, want 3 (the union of the published days; the snapshot's baseline 40 is not added on top)", got)
	}
	// The region count is the union over the same days: 203.0.113.0/24
	// resolves to AU in testdata/geo.csv and the other addresses resolve to
	// nothing. The snapshot's own 130 is not added on top of it.
	if got := totals["NHP_TOTAL_ATTACK_REGIONS"]; got != 1 {
		t.Errorf("NHP_TOTAL_ATTACK_REGIONS = %d, want 1 (the union of the published days, not the baseline 130 plus 1)", got)
	}
}

// TestHistoryFactsJoinTheUnion pins that a snapshot's day contributes its
// attackers to the all-time union exactly once.
//
// The old behavior added the snapshot's baseline on top of the union, which
// counted an address seen in both periods twice and published the sum as a
// "lower bound" -- the opposite of what the sum is.
func TestHistoryFactsJoinTheUnion(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	hist.Totals.SourceIPs = 40
	hist.Totals.Regions = 0
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	totals := readTotalNumbers(t, out)
	if got := totals["NHP_TOTAL_ATTACK_SOURCE_IPS"]; got != 2 {
		t.Errorf("NHP_TOTAL_ATTACK_SOURCE_IPS = %d, want 2 (the union of the published days; the snapshot's baseline 40 is not added on top)", got)
	}

	// The same site without a snapshot unions the first day's two source
	// addresses, and the snapshot's day is analyzed from the log rather than
	// skipped, so the union is the same.
	plainLogs, plainOut := stageTestdata(t)
	runSiteCmd(t, plainLogs, plainOut)
	plain := readTotalNumbers(t, plainOut)
	if plain["NHP_TOTAL_ATTACK_SOURCE_IPS"] != 2 {
		t.Fatalf("the fixture's live union is %d, want 2; the test's arithmetic depends on it",
			plain["NHP_TOTAL_ATTACK_SOURCE_IPS"])
	}
}

// TestHistoryBaselineStandsInForDaysWithoutFacts pins the one case where the
// snapshot's own source count is added to the union: a day the live site
// cannot produce and whose per-day records are gone. Then the snapshot is the
// only carrier of that day's attackers, and the published source figure is
// an upper bound rather than an exact count.
func TestHistoryBaselineStandsInForDaysWithoutFacts(t *testing.T) {
	logs, out := stageTestdata(t)
	const frozenDay = "2026-10-18"
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = frozenDay
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	hist.Totals.SourceIPs = 40
	hist.Totals.Regions = 0
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	totals := readTotalNumbers(t, out)
	plainLogs, plainOut := stageTestdata(t)
	runSiteCmd(t, plainLogs, plainOut)
	plain := readTotalNumbers(t, plainOut)
	want := plain["NHP_TOTAL_ATTACK_SOURCE_IPS"] + 40
	if got := totals["NHP_TOTAL_ATTACK_SOURCE_IPS"]; got != want {
		t.Errorf("NHP_TOTAL_ATTACK_SOURCE_IPS = %d, want %d (the union of the published days plus the snapshot's baseline, which stands in for the day whose records are gone)",
			got, want)
	}
	// The homepage and totalnumber.js say so, figure by figure: the source
	// total is an upper bound, and the region total -- which the snapshot
	// only ever floors, never adds to -- is a lower bound.
	const note = "The source IP total is an upper bound and the region total is a lower bound."
	for _, name := range []string{"index.html", "totalnumber.js"} {
		body := string(readArtifact(t, out, name))
		if !strings.Contains(body, note) {
			t.Errorf("%s does not say %q", name, note)
		}
		if strings.Contains(body, "upper bounds") {
			t.Errorf("%s calls both figures upper bounds; the region figure is a lower bound", name)
		}
	}
}

// TestHistorySnapshotRegionCountIsIgnored pins the invariant the region
// total rests on: the snapshot's own region count cannot reach the published
// figure, whatever it says.
//
// Regions are a small closed set that the snapshot's period and the live one
// share almost entirely, so no arithmetic can merge the snapshot's count with
// the live union. Adding it counted nearly every region twice -- the
// published site reached 162 from a 130-region snapshot, and grew past the
// number of regions there are. The published region total is therefore the
// union over the published days, and the snapshot's count is recorded and
// ignored.
func TestHistorySnapshotRegionCountIsIgnored(t *testing.T) {
	plainLogs, plainOut := stageTestdata(t)
	runSiteCmd(t, plainLogs, plainOut, "-geo-csv", "testdata/geo.csv")
	live := readTotalNumbers(t, plainOut)["NHP_TOTAL_ATTACK_REGIONS"]
	if live == 0 {
		t.Fatal("the fixture's live region union is empty; the test's arithmetic depends on it")
	}

	for _, tc := range []struct {
		name     string
		snapshot int
	}{
		// The snapshot saw more regions than the live days.
		{"snapshot above the union", 130},
		// The snapshot saw the same number of regions.
		{"snapshot equal to the union", live},
		// The snapshot saw none.
		{"snapshot below the union", 0},
		// A number no geo source could have produced.
		{"snapshot beyond the number of regions", 999},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logs, out := stageTestdata(t)
			hist := validHistory()
			hist.Days = hist.Days[:1]
			hist.Days[0].Day = "2026-10-18" // a frozen day with no facts
			hist.Days[0].DropPkts = 100
			hist.Totals.Blocked = 100
			hist.Totals.SourceIPs = 0
			hist.Totals.Regions = tc.snapshot
			path := writeHistory(t, t.TempDir(), hist)

			runSiteCmd(t, logs, out, "-geo-csv", "testdata/geo.csv", "-history", path)

			if got := readTotalNumbers(t, out)["NHP_TOTAL_ATTACK_REGIONS"]; got != live {
				t.Errorf("NHP_TOTAL_ATTACK_REGIONS = %d, want %d (the union over the published days; the snapshot's %d is ignored)",
					got, live, tc.snapshot)
			}
			// Both figures miss the frozen day's own data, and the snapshot
			// adds no sources, so both are lower bounds -- never upper.
			const note = "The source IP and region totals are lower bounds."
			for _, name := range []string{"index.html", "totalnumber.js"} {
				body := string(readArtifact(t, out, name))
				if !strings.Contains(body, note) {
					t.Errorf("%s does not say %q", name, note)
				}
				if strings.Contains(body, "upper bound") {
					t.Errorf("%s calls a figure an upper bound; the snapshot added no sources", name)
				}
			}
		})
	}
}

// TestHistoryBaselineForFigures pins the arithmetic of baselineFor: the
// source remainder over the snapshot's own days with facts is added. The
// snapshot's region count is never part of it.
func TestHistoryBaselineForFigures(t *testing.T) {
	h := &SiteHistory{Totals: HistoryTotals{SourceIPs: 100, Regions: 130}}

	if b := h.baselineFor(factsUnion{Sources: 10, Regions: 5}, 0); b != (historyBaseline{}) {
		t.Errorf("no snapshot day without facts: baseline = %+v, want none", b)
	}

	b := h.baselineFor(factsUnion{Sources: 10, Regions: 5}, 1)
	if b.Sources != 90 {
		t.Errorf("Sources = %d, want 90 (the snapshot's 100 less the 10 its days with facts cover)", b.Sources)
	}
	if !b.Applied {
		t.Error("Applied = false, want true")
	}

	// A snapshot whose own count is already covered by the days with facts
	// adds nothing.
	b = h.baselineFor(factsUnion{Sources: 100, Regions: 5}, 1)
	if b.Sources != 0 {
		t.Errorf("Sources = %d, want 0 (the snapshot's 100 is already covered by the days with facts)", b.Sources)
	}
	if b.Applied {
		t.Error("Applied = true, want false (the baseline adds nothing)")
	}
}

// TestHistoryBoundNote pins the sentence both the homepage and totalnumber.js
// print about each figure, so the two can never disagree and the region
// figure is never called an upper bound.
func TestHistoryBoundNote(t *testing.T) {
	for _, tc := range []struct {
		name string
		b    historyBaseline
		want string
	}{
		{"exact", historyBaseline{}, ""},
		{"nothing stands in", historyBaseline{Missing: true},
			"The source IP and region totals are lower bounds."},
		{"snapshot below the floor", historyBaseline{Unavailable: true, Missing: true},
			"The source IP and region totals are lower bounds."},
		{"snapshot stands in", historyBaseline{Sources: 5, Applied: true, Missing: true},
			"The source IP total is an upper bound and the region total is a lower bound."},
	} {
		if got := tc.b.boundNote(); got != tc.want {
			t.Errorf("%s: boundNote() = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestHistoryBaselineRespectsThePublishingFloor pins that a snapshot listing
// days below the -since floor does not add its own counts to the published
// totals.
//
// The snapshot's own source and region counts are a union over *all* the days
// it lists. When some of those days are below the floor they are not
// published, and the part of the counts that covers them cannot be told apart
// from the part that covers the published days -- a snapshot-level total
// cannot be split by day. Subtracting the published days' facts from the
// snapshot's counts leaves the unpublished days' attackers in the remainder,
// and adding that remainder to the union counts days the site does not
// publish at all.
//
// The baseline is dropped instead: the published figures are the union over
// the published days, and a published day whose per-day records are gone
// contributes nothing to them -- so they are lower bounds.
func TestHistoryBaselineRespectsThePublishingFloor(t *testing.T) {
	logs, out := stageTestdata(t)
	const (
		belowFloor = "2026-10-10"
		frozenDay  = "2026-10-18"
	)
	hist := validHistory()
	hist.Days = []SiteDaySummary{
		{Source: "server", Day: belowFloor, DropPkts: 70, UniqueIPs: 7, DropShare: 0.4},
		{Source: "server", Day: frozenDay, DropPkts: 100, UniqueIPs: 10, DropShare: 0.5},
	}
	hist.Totals.Blocked = 170
	hist.Totals.SourceIPs = 40
	hist.Totals.Regions = 3
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-geo-csv", "testdata/geo.csv", "-history", path)

	// The live union of the staged days, which is what the published
	// figures have to be.
	plainLogs, plainOut := stageTestdata(t)
	runSiteCmd(t, plainLogs, plainOut, "-geo-csv", "testdata/geo.csv")
	plain := readTotalNumbers(t, plainOut)
	live := plain["NHP_TOTAL_ATTACK_SOURCE_IPS"]
	if live == 0 {
		t.Fatal("the fixture's live union is empty; the test's arithmetic depends on it")
	}

	totals := readTotalNumbers(t, out)
	if got := totals["NHP_TOTAL_ATTACK_SOURCE_IPS"]; got != live {
		t.Errorf("NHP_TOTAL_ATTACK_SOURCE_IPS = %d, want %d (the union of the published days; the snapshot's own count covers a day below the floor and cannot be applied)",
			got, live)
	}
	// The regions of the unpublished day must not be in the total either.
	if got, want := totals["NHP_TOTAL_ATTACK_REGIONS"], plain["NHP_TOTAL_ATTACK_REGIONS"]; got != want {
		t.Errorf("NHP_TOTAL_ATTACK_REGIONS = %d, want %d (the union of the published days; the snapshot's own count of 3 covers a day below the floor and cannot be applied)",
			got, want)
	}

	// The packets of the day below the floor are not published either: the
	// floor applies to the snapshot's days as well.
	wantPkts := plain["NHP_TOTAL_ATTACK_BLOCKED"] + 100
	if got := totals["NHP_TOTAL_ATTACK_BLOCKED"]; got != wantPkts {
		t.Errorf("NHP_TOTAL_ATTACK_BLOCKED = %d, want %d (the published days plus the frozen day's 100; the day below the floor is not published)",
			got, wantPkts)
	}

	// The day below the floor is not in the index either.
	b, err := os.ReadFile(filepath.Join(out, "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var idx []SiteDaySummary
	if err := json.Unmarshal(b, &idx); err != nil {
		t.Fatal(err)
	}
	for _, r := range idx {
		if r.Day == belowFloor {
			t.Errorf("index.json lists %s, which is below the publishing floor", belowFloor)
		}
	}

	// The homepage says the figures are lower bounds: the frozen day's
	// records are gone and nothing stands in for them.
	page := string(readArtifact(t, out, "index.html"))
	if !strings.Contains(page, "lower bounds") {
		t.Error("the homepage does not say that the totals are lower bounds")
	}
	if strings.Contains(page, "upper bounds") {
		t.Error("the homepage says the totals are upper bounds; the snapshot's own count is not applied at all")
	}
}

// TestHistorySinceComesFromSnapshot pins that the counting start is the
// snapshot's, not the earliest row's: the snapshot is what says the site's
// numbers begin before its first published day.
func TestHistorySinceComesFromSnapshot(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	// The snapshot counts from a day the site does not publish. The floor
	// is lowered to match, so the snapshot's day is not below it.
	hist.Since = "2026-10-15"
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-since", "2026-10-15", "-history", path)

	page := string(readArtifact(t, out, "index.html"))
	if !strings.Contains(page, "Since 2026-10-15.") {
		t.Error("the homepage does not count from the snapshot's since date")
	}
	b := string(readArtifact(t, out, "totalnumber.js"))
	if !strings.Contains(b, "Counting starts: 2026-10-15 (UTC).") {
		t.Error("totalnumber.js does not record the snapshot's counting start")
	}
}

// TestHistorySinceBelowTheFloorIsIgnored pins that the snapshot's own
// counting start cannot put the site below its publishing floor: the days
// before the floor are not published, so the site does not count from them.
func TestHistorySinceBelowTheFloorIsIgnored(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	hist.Since = "2026-10-01"
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = siteDay1
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	page := string(readArtifact(t, out, "index.html"))
	if strings.Contains(page, "Since 2026-10-01.") {
		t.Error("the homepage counts from a day below the publishing floor")
	}
	if !strings.Contains(page, "Since 2026-10-16.") {
		t.Error("the homepage does not count from the earliest published day")
	}
	b := string(readArtifact(t, out, "totalnumber.js"))
	if strings.Contains(b, "Counting starts: 2026-10-01") {
		t.Error("totalnumber.js counts from a day below the publishing floor")
	}
}

// TestHistoryFrozenRowsRespectTheSinceFloor pins that the publishing floor
// applies to the snapshot's days as well: a day below it is not in the index
// and does not reach the hero total.
func TestHistoryFrozenRowsRespectTheSinceFloor(t *testing.T) {
	logs, out := stageTestdata(t)
	hist := validHistory()
	hist.Since = "2026-10-16"
	hist.Days = []SiteDaySummary{
		{Source: "server", Day: "2026-10-14", DropPkts: 111},
		{Source: "server", Day: "2026-10-15", DropPkts: 222},
		{Source: "server", Day: "2026-10-18", DropPkts: 333},
	}
	hist.Totals.Blocked = 111 + 222 + 333
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	page := string(readArtifact(t, out, "index.html"))
	for _, d := range []string{"2026-10-14", "2026-10-15"} {
		if strings.Contains(page, ">"+d+"<") {
			t.Errorf("%s is in the index although it is below the publishing floor", d)
		}
	}
	if !strings.Contains(page, ">2026-10-18<") {
		t.Error("the day on or after the floor is missing from the index")
	}

	// The hero total counts the published days only.
	idx := readArtifact(t, out, "index.json")
	var rows []SiteDaySummary
	if err := json.Unmarshal(idx, &rows); err != nil {
		t.Fatal(err)
	}
	var sum uint64
	for _, r := range rows {
		if r.Day == "2026-10-14" || r.Day == "2026-10-15" {
			t.Errorf("index.json carries the pre-floor row %s", r.Day)
		}
		sum += r.DropPkts
	}
	if sum != indexTotalFromJSON(t, out) {
		t.Errorf("the rows add up to %d but the hero shows %d", sum, indexTotalFromJSON(t, out))
	}
	if sum <= 111+222 {
		t.Errorf("the hero total is %d; the pre-floor days are still counted in it", sum)
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
	firstFP := readSiteStateAt(t, filepath.Join(out, "state.json")).History
	if firstFP == "" {
		t.Fatal("state.json does not record the history fingerprint")
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

	// Changing the snapshot has to reach the state, and the state is what
	// makes the next run rewrite the index.
	hist.Totals.Blocked = 100
	hist.Totals.SourceIPs = 4242
	hist.Days[0].DropPkts = 100
	writeHistory(t, filepath.Dir(path), hist)
	runSiteCmd(t, logs, out, "-history", path)

	state := readSiteStateAt(t, filepath.Join(out, "state.json"))
	if state.History == firstFP {
		t.Error("the changed snapshot's fingerprint did not reach state.json")
	}
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
	const frozenDay = "2026-10-18"
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = frozenDay
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-history", path)

	idx := readArtifact(t, out, "index.json")
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
	// The frozen day's packet count is the snapshot's: there is no log for
	// it to be counted from.
	for _, r := range rows {
		if r.Day == frozenDay && r.DropPkts != 100 {
			t.Errorf("the frozen day's count is %d, want the snapshot's 100", r.DropPkts)
		}
	}
}
func TestHistoryGeneratorFromSavedPage(t *testing.T) {
	out := filepath.Join(t.TempDir(), "history.json")
	var buf strings.Builder
	err := run([]string{"history", "-html", filepath.Join("testdata", "history-page-2026-10-26.html"), "-out", out}, &buf)
	if err != nil {
		t.Fatalf("history: %v\n%s", err, buf.String())
	}
	h, err := loadSiteHistory(out)
	if err != nil {
		t.Fatalf("the generated snapshot does not load: %v", err)
	}
	if h.Since != "2026-10-16" {
		t.Errorf("since = %q, want 2026-10-16", h.Since)
	}
	if h.Totals.Blocked != 3501020 {
		t.Errorf("totals.blocked = %d, want 3501020", h.Totals.Blocked)
	}
	if h.Totals.SourceIPs != 25136 {
		t.Errorf("totals.source_ips = %d, want 25136", h.Totals.SourceIPs)
	}
	if h.Totals.Regions != 0 {
		t.Errorf("totals.regions = %d, want 0 (the snapshot records no region count; the site's region total is the union over the published days)",
			h.Totals.Regions)
	}
	if len(h.Days) != 11 {
		t.Fatalf("the snapshot lists %d days, want 11", len(h.Days))
	}
	want := map[string]uint64{
		"2026-10-26": 8162,
		"2026-10-25": 554239,
		"2026-10-24": 876945,
		"2026-10-23": 382917,
		"2026-10-22": 574967,
		"2026-10-21": 358779,
		"2026-10-20": 217022,
		"2026-10-19": 385531,
		"2026-10-18": 28456,
		"2026-10-17": 24092,
		"2026-10-16": 89910,
	}
	for _, d := range h.Days {
		w, ok := want[d.Day]
		if !ok {
			t.Errorf("the snapshot lists an unexpected day %s", d.Day)
			continue
		}
		if d.DropPkts != w {
			t.Errorf("%s: drop_pkts = %d, want %d", d.Day, d.DropPkts, w)
		}
		if d.Source != "server" {
			t.Errorf("%s: source = %q, want server", d.Day, d.Source)
		}
	}
	// The flags survive the round trip: the page's chips are the only
	// record of them the snapshot has.
	byDay := map[string]SiteDaySummary{}
	for _, d := range h.Days {
		byDay[d.Day] = d
	}
	if !byDay["2026-10-26"].Partial {
		t.Error("2026-10-26 lost its partial flag")
	}
	if byDay["2026-10-26"].RestartsDetected != 1 {
		t.Errorf("2026-10-26: restarts_detected = %d, want 1", byDay["2026-10-26"].RestartsDetected)
	}
	if byDay["2026-10-25"].RestartsDetected != 0 {
		t.Errorf("2026-10-25: restarts_detected = %d, want 0", byDay["2026-10-25"].RestartsDetected)
	}
	if byDay["2026-10-26"].UniqueIPs != 3125 {
		t.Errorf("2026-10-26: unique_attack_ips = %d, want 3125", byDay["2026-10-26"].UniqueIPs)
	}
	if byDay["2026-10-26"].PeakHour != "2026-10-26 00:00" {
		t.Errorf("2026-10-26: peak_hour = %q", byDay["2026-10-26"].PeakHour)
	}
	if byDay["2026-10-26"].TopCategory != "TCP SYN scan" {
		t.Errorf("2026-10-26: top_category = %q", byDay["2026-10-26"].TopCategory)
	}
}

// TestHistoryPageParsesBothRegionHeroIds pins the backward compatibility the
// homepage's third hero figure rests on.
//
// The figure was labeled "Countries" until it was renamed to "Regions" (the
// codes it counts are ISO 3166-1 alpha-2, which cover a country, a dependency
// and a special political area alike), and its element id went with it. A copy
// of the homepage saved before the rename is the only carrier of the days
// below the publishing floor, so the parser has to read the old id as well as
// the new one.
func TestHistoryPageParsesBothRegionHeroIds(t *testing.T) {
	for _, id := range []string{"total-regions", "total-countries"} {
		t.Run(id, func(t *testing.T) {
			page := `<html><body><div class="hero-stats">` +
				`<div class="hero-stat"><div><span class="hero-num" id="total-attacks" data-total="3501020">3,501,020</span></div><div class="hero-label">Attack packets blocked</div></div>` +
				`<div class="hero-stat"><div><span class="hero-num" id="total-sources" data-total="25136">25,136</span></div><div class="hero-label">Source IPs</div></div>` +
				`<div class="hero-stat"><div><span class="hero-num" id="` + id + `" data-total="130">130</span></div><div class="hero-label">Regions</div></div>` +
				`</div></body></html>`

			got, err := historyPageFromHTMLBytes([]byte(page))
			if err != nil {
				t.Fatalf("historyPageFromHTMLBytes: %v", err)
			}
			if got.Regions != 130 {
				t.Errorf("regions = %d, want 130", got.Regions)
			}
		})
	}
}

// TestHistoryGeneratorRefusesInconsistentPage pins the generator's own
// check: a page whose rows do not add up to its hero total is refused, so a
// parsing slip cannot become a permanent baseline.
func TestHistoryGeneratorRefusesInconsistentPage(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "history-page-2026-10-26.html"))
	if err != nil {
		t.Fatal(err)
	}
	broken := strings.Replace(string(b), `data-total="3501020"`, `data-total="3501021"`, 1)
	if broken == string(b) {
		t.Fatal("the fixture no longer carries the hero total")
	}
	dir := t.TempDir()
	page := filepath.Join(dir, "page.html")
	if err = os.WriteFile(page, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	err = run([]string{"history", "-html", page, "-out", filepath.Join(dir, "history.json")}, &buf)
	if err == nil {
		t.Fatal("the generator accepted a page whose rows do not add up")
	}
	if !strings.Contains(err.Error(), "DROP packets") {
		t.Errorf("error does not explain the mismatch: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "history.json")); err == nil {
		t.Error("a snapshot was written for a page that does not add up")
	}
}

// TestHistoryGeneratorRefusesACountItCannotRead pins the bound on the counts
// the page shows. A figure past the range of an int is not a count of source
// addresses or of countries: converting it into an int wraps it into a number
// the page never said, and freezing that number would publish it as the
// site's own baseline.
func TestHistoryGeneratorRefusesACountItCannotRead(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "history-page-2026-10-26.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		total string
	}{
		// 2^63 fits in a uint64 but not in an int, so converting it wraps
		// it into a negative count.
		{name: "past the range of an int", total: "9223372036854775808"},
		{name: "negative", total: "-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			broken := strings.Replace(string(b), `data-total="25136"`, `data-total="`+tc.total+`"`, 1)
			if broken == string(b) {
				t.Fatal("the fixture no longer carries the source count")
			}
			dir := t.TempDir()
			page := filepath.Join(dir, "page.html")
			if err := os.WriteFile(page, []byte(broken), 0o644); err != nil {
				t.Fatal(err)
			}

			var buf strings.Builder
			err := run([]string{"history", "-html", page, "-out", filepath.Join(dir, "history.json")}, &buf)
			if err == nil {
				t.Fatal("the generator accepted a source count it cannot read")
			}
			if !strings.Contains(err.Error(), "total-sources") {
				t.Errorf("error does not name the figure: %v", err)
			}
			if _, err := os.Stat(filepath.Join(dir, "history.json")); err == nil {
				t.Error("a snapshot was written for a page whose count cannot be read")
			}
		})
	}
}

// TestHistoryGeneratorFromIndexJSON pins the second input path: a published
// index.json plus the totalnumber.js beside it produce the same snapshot the
// saved page does.
func TestHistoryGeneratorFromIndexJSON(t *testing.T) {
	dir := t.TempDir()
	fromPage := filepath.Join(dir, "from-page.json")
	var buf strings.Builder
	if err := run([]string{"history", "-html", filepath.Join("testdata", "history-page-2026-10-26.html"), "-out", fromPage}, &buf); err != nil {
		t.Fatalf("history -html: %v\n%s", err, buf.String())
	}

	// The index.json path needs a site to have published it, so the fixture
	// page's own rows are rebuilt into one — the same numbers, in the
	// machine-readable shape the site publishes.
	page, err := os.ReadFile(filepath.Join("testdata", "history-page-2026-10-26.html"))
	if err != nil {
		t.Fatal(err)
	}
	hist, err := historyPageFromHTMLBytes(page)
	if err != nil {
		t.Fatal(err)
	}
	rows := hist.Rows
	idxBytes, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	idxPath := filepath.Join(dir, "index.json")
	if err = os.WriteFile(idxPath, idxBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	totalsPath := filepath.Join(dir, "totalnumber.js")
	totalsBody := totalNumberJS(hist.Blocked, hist.Sources, hist.Regions, hist.Since, historyBaseline{}, nil)
	if err = os.WriteFile(totalsPath, []byte(totalsBody), 0o644); err != nil {
		t.Fatal(err)
	}

	fromIndex := filepath.Join(dir, "from-index.json")
	if err = run([]string{"history", "-index", idxPath, "-totals", totalsPath, "-out", fromIndex}, &buf); err != nil {
		t.Fatalf("history -index: %v\n%s", err, buf.String())
	}

	a, err := os.ReadFile(fromPage)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(fromIndex)
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Errorf("the two input paths disagree:\n--- page ---\n%s\n--- index ---\n%s", a, b)
	}
}

// TestHistoryGeneratorRefusesAnIndexWithHistoryRows pins that a snapshot
// cannot be taken from a site that already carries one: the frozen days
// would be counted twice.
func TestHistoryGeneratorRefusesAnIndexWithHistoryRows(t *testing.T) {
	dir := t.TempDir()
	idx := filepath.Join(dir, "index.json")
	rows := []SiteDaySummary{{Source: "server", Day: "2026-10-16", DropPkts: 10, History: true}}
	b, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(idx, b, 0o644); err != nil {
		t.Fatal(err)
	}
	totals := filepath.Join(dir, "totalnumber.js")
	if err = os.WriteFile(totals, []byte(totalNumberJS(10, 1, 1, "2026-10-16", historyBaseline{}, nil)), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf strings.Builder
	err = run([]string{"history", "-index", idx, "-totals", totals, "-out", filepath.Join(dir, "history.json")}, &buf)
	if err == nil {
		t.Fatal("the generator accepted an index that already lists history rows")
	}
	if !strings.Contains(err.Error(), "history rows") {
		t.Errorf("error does not explain the problem: %v", err)
	}
}

// TestHistoryFlagsIncludeHistory pins that the frozen rows are marked on the
// homepage: a reader who cannot click a date deserves to know why.
func TestHistorySeedSnapshotIsShipped(t *testing.T) {
	seed := filepath.Join("..", "..", "deploy", "config-templates", "xdpstats", "history.json")
	want := filepath.Join(t.TempDir(), "history.json")
	var buf strings.Builder
	if err := run([]string{"history", "-html", filepath.Join("testdata", "history-page-2026-10-26.html"), "-out", want}, &buf); err != nil {
		t.Fatalf("history: %v\n%s", err, buf.String())
	}
	a, err := os.ReadFile(seed)
	if err != nil {
		t.Fatalf("the shipped snapshot is missing: %v", err)
	}
	b, err := os.ReadFile(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Errorf("the shipped snapshot differs from the generated one:\n--- shipped ---\n%s\n--- generated ---\n%s", a, b)
	}
}

// TestHistorySeedSnapshotCarriesNoRegionCount pins the shipped seed: it
// records no region count at all.
//
// The site never applies the snapshot's region count, but a snapshot that
// carries one invites the next reader to use it -- and a hand-edited file is
// exactly how a region count got back into this one before.
func TestHistorySeedSnapshotCarriesNoRegionCount(t *testing.T) {
	seed := filepath.Join("..", "..", "deploy", "config-templates", "xdpstats", "history.json")
	b, err := os.ReadFile(seed)
	if err != nil {
		t.Fatalf("the shipped snapshot is missing: %v", err)
	}
	var h SiteHistory
	if err := json.Unmarshal(b, &h); err != nil {
		t.Fatalf("the shipped snapshot does not parse: %v", err)
	}
	if h.Totals.Regions != 0 {
		t.Errorf("the shipped snapshot records %d regions, want 0 (the site's region total is the union over the published days)",
			h.Totals.Regions)
	}
}

// TestHistorySnapshotReadsTheLegacyRegionKey pins that a snapshot written
// before the rename still loads.
//
// The key was renamed with the figure: "countries" was the wrong word for the
// ISO 3166-1 codes the standard assigns to every country and area it lists.
// The field's meaning did not change with the spelling -- it is recorded and
// never applied -- so a snapshot that carries the old key is read in full
// rather than refused.
func TestHistorySnapshotReadsTheLegacyRegionKey(t *testing.T) {
	legacy := []byte(`{
  "schema": 1,
  "since": "2026-10-15",
  "totals": { "blocked": 300, "source_ips": 40, "countries": 5 },
  "days": [
    { "source": "server", "day": "2026-10-15", "drop_pkts": 100, "unique_attack_ips": 10, "drop_share": 0.5 },
    { "source": "server", "day": "2026-10-16", "drop_pkts": 200, "unique_attack_ips": 20, "drop_share": 0.9 }
  ]
}`)
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, legacy, 0o644); err != nil {
		t.Fatal(err)
	}
	h, err := loadSiteHistory(path)
	if err != nil {
		t.Fatalf("a snapshot written before the rename was refused: %v", err)
	}
	if h.Totals.Regions != 5 {
		t.Errorf("regions = %d, want 5 (the legacy key is read)", h.Totals.Regions)
	}
	if h.Totals.Blocked != 300 || h.Totals.SourceIPs != 40 {
		t.Errorf("the other totals were not read: %+v", h.Totals)
	}
}

// TestHistoryRegionTotalIsBoundedByTheGeoSource pins where the published
// region total comes from: the distinct region labels the geo source
// produced over the published days, and nothing else.
//
// The site's figure therefore cannot exceed the number of regions there
// are, whatever the snapshot says -- the property the whole region rule
// exists to protect.
func TestHistoryRegionTotalIsBoundedByTheGeoSource(t *testing.T) {
	root := t.TempDir()
	logs := filepath.Join(root, "raw")
	out := filepath.Join(root, "web")
	statePath := filepath.Join(root, "state", "state.json")

	// Three region labels, one per source range.
	geoPath := filepath.Join(root, "geo.csv")
	geo := "cidr,region,city\n" +
		"198.51.100.0/24,US,New York\n" +
		"203.0.113.0/24,AU,Sydney\n" +
		"192.0.2.0/24,DE,Berlin\n"
	if err := os.WriteFile(geoPath, []byte(geo), 0o644); err != nil {
		t.Fatal(err)
	}

	writeDropLogWithSources(t, logs, "server", siteDay1, 100,
		"198.51.100.7", "203.0.113.5", "192.0.2.9")
	writeDropLogWithSources(t, logs, "server", siteDay2, 200,
		"198.51.100.7")

	// A snapshot that claims a region count of its own, frozen on a day the
	// live site cannot produce -- so the site's baseline path runs, and the
	// published region total still has to be the union.
	hist := validHistory()
	hist.Days = hist.Days[:1]
	hist.Days[0].Day = "2026-10-18"
	hist.Days[0].DropPkts = 100
	hist.Totals.Blocked = 100
	hist.Totals.SourceIPs = 40
	hist.Totals.Regions = 130
	path := writeHistory(t, t.TempDir(), hist)

	runSiteCmd(t, logs, out, "-state", statePath, "-geo-csv", geoPath, "-history", path)

	const want = 3
	if got := readTotalNumbers(t, out)["NHP_TOTAL_ATTACK_REGIONS"]; got != want {
		t.Errorf("NHP_TOTAL_ATTACK_REGIONS = %d, want %d (the union over the published days; the snapshot's 130 is ignored)",
			got, want)
	}
}

// TestHistoryRowsAreNotPublishedTwice pins the failure the whole feature
// exists to prevent: the snapshot's days and the collector's days overlap
// exactly, and counting both would double the site's headline total.
