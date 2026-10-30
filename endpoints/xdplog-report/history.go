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
	// saw. It is a union within that period, and it is added to the live
	// period's union rather than merged with it — the two periods' address
	// sets were never kept, so a true union is not computable. The result
	// is a lower bound.
	SourceIPs int `json:"source_ips"`
	// Countries is the same for countries.
	Countries int `json:"countries"`
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
	if h.Totals.Countries < 0 {
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

// Fingerprint identifies the snapshot's bytes. It goes into state.json so a
// changed snapshot rewrites the homepage: without it the counters would keep
// the baseline they had when the snapshot was first read.
func (h *SiteHistory) Fingerprint() string {
	if h == nil {
		return ""
	}
	return h.fingerprint
}
