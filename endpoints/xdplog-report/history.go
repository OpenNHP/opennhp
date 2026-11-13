package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// The history snapshot: the site's own numbers for a period it no longer has
// the data for.
//
// stats.opennhp.org starts counting on 2026-10-16 (defaultSiteSince) and the
// collector's cache does not reach back past the daemon's 14-day retention, so
// the totals for the days before that floor live nowhere. A saved copy of the
// homepage is the only thing that can carry them, and `xdplog-report history`
// turns that copy into the file below.
//
// A frozen day is authoritative for its own numbers. The live copy of the same
// day is not analyzed, not published and not counted, so the day appears once
// on the homepage and the totals do not double — the failure this snapshot
// exists to prevent, since the snapshot's days and the collector's days
// overlap exactly.
//
// What a frozen day does not have is a detail page: the snapshot carries the
// day's summary, not its DROP events. So a frozen day is rendered as plain
// text on the homepage and carries a `history` flag, and the site says so in
// its note.

// historySchema versions the on-disk shape of history.json. A file with a
// different schema is refused rather than half-read: the numbers it carries
// are a baseline the whole site rests on, and guessing what an unknown field
// meant is exactly how a wrong total becomes permanent.
const historySchema = 1

// SiteHistory is the on-disk shape of history.json.
type SiteHistory struct {
	// Schema is the format version, always historySchema.
	Schema int `json:"schema"`
	// Since is the day the site's counting starts, YYYY-MM-DD. It can be
	// earlier than the earliest day this snapshot lists, because a snapshot
	// records what the page said, not every day it published.
	Since string `json:"since"`
	// Totals is the baseline the snapshot carries: the all-time figures the
	// saved page showed.
	Totals HistoryTotals `json:"totals"`
	// Days lists the days the snapshot freezes, one row per published day.
	Days []SiteDaySummary `json:"days"`

	// fingerprint is the sha256 of the file's bytes, and covered indexes the
	// (source, day) pairs the snapshot freezes. Neither is serialized: the
	// fingerprint is a property of the file, not of its meaning, and the
	// index is a lookup table for the rows already in Days.
	fingerprint string
	covered     map[string]struct{}
}

// HistoryTotals is the baseline a snapshot carries for the site's three
// all-time figures.
type HistoryTotals struct {
	// Blocked is the all-time DROP packet count the saved page showed. It
	// must equal the sum of the days' drop_pkts, which is what keeps the
	// site's "the hero total is the sum of index.json's drop_pkts"
	// contract true once history rows are in the index.
	Blocked uint64 `json:"blocked"`
	// SourceIPs is the number of distinct source addresses the saved period
	// saw. It is a union within that period.
	//
	// It is NOT added to the live period's union on top of it. Two unions
	// can only be merged when both sets are present; the site's per-day
	// facts (facts.go) are what make that possible, and a day the live site
	// can analyze contributes its own set to the union. What is left of
	// this baseline is the part that covers days whose facts are gone, and
	// only that part is added -- see historyBaseline.
	SourceIPs int `json:"source_ips"`
	// Regions is the number of distinct regions the saved period saw.
	//
	// It is recorded and never applied. The published region total is the
	// union over the published days and nothing else; see historyBaseline.
	// The field stays in the file for schema compatibility and so a reader
	// can tell what the saved page said, but the site only checks that it is
	// not negative.
	//
	// The figure counts the ISO 3166-1 alpha-2 codes the international
	// standard assigns to every country and area it lists -- what the GeoIP
	// database attributes an address to -- and the page calls them regions.
	// The json key stays "countries": the snapshot format is read by every
	// deployment that holds one, and only the Go name and the labels changed.
	//
	// Regions are a closed set of ~250 codes that the snapshot's period and
	// the live one share almost entirely, so no arithmetic can merge the
	// snapshot's count with the live union: adding counts most of them twice
	// (the site published 162 from a 130-region snapshot) and can push the
	// total past the number of regions there are, and taking the larger of
	// the two can only ever be a lower bound that misses a region seen in
	// one period alone. The count is therefore kept out of the arithmetic
	// entirely.
	Regions int `json:"countries"`
}

// loadSiteHistory reads and validates a history snapshot.
//
// Every check here refuses the file rather than degrading: a snapshot that
// cannot be read is a configuration error, and the site must not fall back to
// counting only its own days, because the whole point of the snapshot is the
// totals that are in nobody else's hands.
func loadSiteHistory(path string) (*SiteHistory, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("history file %s: %w", path, err)
	}
	var h SiteHistory
	if err := json.Unmarshal(b, &h); err != nil {
		return nil, fmt.Errorf("history file %s: %w", path, err)
	}
	if h.Schema != historySchema {
		return nil, fmt.Errorf("history file %s: schema %d, expected %d", path, h.Schema, historySchema)
	}
	if h.Since != "" {
		if _, err := time.Parse("2006-01-02", h.Since); err != nil {
			return nil, fmt.Errorf("history file %s: since must be YYYY-MM-DD: %q", path, h.Since)
		}
	}
	if h.Totals.SourceIPs < 0 {
		return nil, fmt.Errorf("history file %s: totals.source_ips must not be negative", path)
	}
	if h.Totals.Regions < 0 {
		return nil, fmt.Errorf("history file %s: totals.countries must not be negative", path)
	}

	h.covered = make(map[string]struct{}, len(h.Days))
	var daySum uint64
	for i, d := range h.Days {
		if d.Source == "" {
			return nil, fmt.Errorf("history file %s: day %d has no source", path, i)
		}
		if _, err := time.Parse("2006-01-02", d.Day); err != nil {
			return nil, fmt.Errorf("history file %s: day %d has an invalid date %q", path, i, d.Day)
		}
		key := d.Source + "/" + d.Day
		if _, dup := h.covered[key]; dup {
			return nil, fmt.Errorf("history file %s: lists %s twice", path, key)
		}
		h.covered[key] = struct{}{}
		// A share outside 0..1 is a parsing accident, not a rate. The
		// tolerance absorbs the rounding a percentage with two decimals
		// goes through on its way back to a fraction.
		if d.DropShare < 0 || d.DropShare > 1.000001 {
			return nil, fmt.Errorf("history file %s: %s has drop_share %g, want 0..1", path, key, d.DropShare)
		}
		daySum += d.DropPkts
	}
	if daySum != h.Totals.Blocked {
		return nil, fmt.Errorf("history file %s: totals.blocked (%d) does not match the sum of its days (%d)",
			path, h.Totals.Blocked, daySum)
	}

	sum := sha256.Sum256(b)
	h.fingerprint = hex.EncodeToString(sum[:])
	return &h, nil
}

// Covers reports whether the snapshot freezes (source, day). A frozen day is
// published from the snapshot and its live copy is ignored, so the day is
// counted once.
func (h *SiteHistory) Covers(source, day string) bool {
	if h == nil {
		return false
	}
	_, ok := h.covered[source+"/"+day]
	return ok
}

// Rows returns the snapshot's days as index rows, each marked as coming from
// the snapshot. The marker is what makes the homepage render the day as plain
// text: a frozen day has no detail page to link to.
//
// The rows are the fallback for the days the live site cannot produce. A day
// the collector still has a raw log for -- or that the site published before
// -- is analyzed live and supersedes the snapshot's row for it: the live
// numbers, a real detail page, and the day's attackers in the all-time union.
func (h *SiteHistory) Rows() []SiteDaySummary {
	if h == nil {
		return nil
	}
	out := make([]SiteDaySummary, 0, len(h.Days))
	for _, d := range h.Days {
		d.History = true
		out = append(out, d)
	}
	return out
}

// historyBaseline is what the history snapshot contributes to the all-time
// source count.
//
// The snapshot carries its own all-time figures for the period it froze. The
// site's per-day facts (facts.go) carry the same information for every day
// the live site can analyze, and those sets are unioned into the published
// figures -- so the baseline must NOT be added on top of them. Two unions can
// only be merged when both sets are present; adding their sizes instead
// counts an address seen in both periods twice, because |A| + |B| >= |A union
// B|. The site used to publish that sum and call it a lower bound, which is
// the opposite of what it is.
//
// What is left of the baseline is the part that covers the snapshot's days
// whose facts are gone -- the days the live site cannot produce at all.
// Nothing else carries those days' attackers, so the remainder is added and
// the published source figure becomes an upper bound. Addresses are an open
// set, so the remainder is mostly new attackers, and adding it is the closest
// figure available.
//
// The region total is never touched by the snapshot. Regions are a closed
// set of ~250 codes that the two periods share almost entirely, so no
// arithmetic can merge the snapshot's count with the live union: adding
// counts most of them twice and can push the total past the number of
// regions there are, while the larger of the two can only ever miss a
// region seen in one period alone. The published region total is therefore
// the union over the published days, and the snapshot's own count is recorded
// and never applied -- see HistoryTotals.Regions.
type historyBaseline struct {
	// Sources is the number of attackers the baseline adds.
	Sources int
	// Applied is true when the baseline contributed anything, i.e. when the
	// snapshot's own source count stood in for days whose facts are gone.
	Applied bool
	// Unavailable is true when the baseline could not be applied at all: the
	// snapshot lists days below the publishing floor, and a snapshot-level
	// total cannot be split by day.
	//
	// The snapshot's own counts are a union over *all* its days. When some of
	// those days are not published, the part of the counts that covers them
	// cannot be told apart from the part that covers the published days, so
	// subtracting the published days' own sets leaves the unpublished days'
	// attackers in the remainder — and adding that remainder to the published
	// totals counts days the site does not publish at all.
	//
	// The baseline is therefore dropped: the published figures are the union
	// over the published days and nothing else, and a published day whose
	// per-day records are gone contributes nothing to them.
	Unavailable bool
	// Missing is true when a published day contributed no source or region
	// data, whether or not the snapshot stood in for it. Neither figure is
	// then an exact count; sourcesBound and regionsBound say which way each
	// one is off.
	Missing bool
}

// figureBound says what kind of number a published all-time figure is.
type figureBound int

const (
	boundExact figureBound = iota
	boundUpper
	boundLower
)

// sourcesBound is the kind of number the published source IP total is.
//
// The snapshot's remainder is added when it stood in, which can count an
// address seen in both periods twice: an upper bound. A day with no facts
// and nothing added for it is simply missing: a lower bound.
func (b historyBaseline) sourcesBound() figureBound {
	switch {
	case b.Sources > 0:
		return boundUpper
	case b.Missing:
		return boundLower
	}
	return boundExact
}

// regionsBound is the kind of number the published region total is.
//
// The snapshot never contributes to it, so the figure is the union over the
// published days and no more. With any day's facts gone, a region seen only
// on that day can be missing: a lower bound.
func (b historyBaseline) regionsBound() figureBound {
	if b.Missing {
		return boundLower
	}
	return boundExact
}

// boundNote is the sentence the homepage and totalnumber.js both print about
// what kind of number each of the two figures is. Generating it in one place
// is what keeps the two from disagreeing. It is empty when both are exact.
func (b historyBaseline) boundNote() string {
	s, c := b.sourcesBound(), b.regionsBound()
	switch {
	case s == boundExact && c == boundExact:
		return ""
	case s == c:
		return "The source IP and region totals are " + boundPlural(s) + "."
	case s == boundExact:
		return "The region total is " + boundSingular(c) + "."
	case c == boundExact:
		return "The source IP total is " + boundSingular(s) + "."
	}
	return "The source IP total is " + boundSingular(s) + " and the region total is " + boundSingular(c) + "."
}

func boundSingular(f figureBound) string {
	if f == boundUpper {
		return "an upper bound"
	}
	return "a lower bound"
}

func boundPlural(f figureBound) string {
	if f == boundUpper {
		return "upper bounds"
	}
	return "lower bounds"
}

// BelowFloor reports whether the snapshot lists any day the site does not
// publish.
//
// The publishing floor applies to the snapshot's days as well: a day below it
// is not published, is not in the index and does not reach the hero total. The
// snapshot's own all-time counts, though, are a union over every day it
// lists, so the part of them that covers the unpublished days cannot be
// separated from the part that covers the published ones.
//
// The caller drops the baseline when this is true; see historyBaseline.
func (h *SiteHistory) BelowFloor(publishes func(day string) bool) bool {
	if h == nil || publishes == nil {
		return false
	}
	for _, d := range h.Days {
		if !publishes(d.Day) {
			return true
		}
	}
	return false
}

// baselineFor computes what the snapshot contributes to the all-time source
// count.
//
// snapUnion is the union over the snapshot's published days whose per-day
// facts are still on disk, and missing counts the snapshot's published days
// whose facts are not. The baseline is used only when there are such days.
//
// Sources: only the part of the snapshot's count that the days with facts do
// not already account for is added. The snapshot's count is the union over
// *all* its days, so subtracting the part the facts cover leaves what the
// remaining days alone can have contributed.
//
// Regions: nothing. The snapshot's region count is recorded and never
// applied; the published region total is the union over the published days.
// See historyBaseline.
//
// The caller must not use this when the snapshot lists days below the
// publishing floor: the "remaining days" would then include days the site
// does not publish, and their attackers would be counted in the published
// totals. See BelowFloor and historyBaseline.Unavailable.
func (h *SiteHistory) baselineFor(snapUnion factsUnion, missing int) historyBaseline {
	var b historyBaseline
	if h == nil || missing == 0 {
		return b
	}
	b.Sources = h.Totals.SourceIPs - snapUnion.Sources
	if b.Sources < 0 {
		b.Sources = 0
	}
	b.Applied = b.Sources > 0
	return b
}

// Fingerprint identifies the snapshot's bytes. It goes into state.json so a
// changed snapshot rewrites the homepage: without it the counters would keep
// the baseline they had when the snapshot was first read.
func (h *SiteHistory) Fingerprint() string {
	if h == nil {
		return ""
	}
	return h.fingerprint
}
