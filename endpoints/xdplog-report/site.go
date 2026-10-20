package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// SiteOptions are the parsed flags for the `site` subcommand. -mask-ip defaults
// to true here (different from the single-file CLI) because the public web
// surface should not leak full source IPs.
type SiteOptions struct {
	LogsDir string
	OutDir  string
	// StatePath is the incremental-build state file. It defaults to
	// <OutDir>/state.json, but -state lets a caller keep it OUT of the
	// directory nginx serves: the state records the absolute path of every
	// raw cache file, which is host layout an anonymous visitor has no
	// business reading. Point it somewhere private.
	StatePath string
	TopN      int
	GeoMMDB   string
	GeoCSV    string
	MaskIP    bool
	Force     bool
	Today     string // YYYY-MM-DD (UTC), supplied by the caller
}

// SiteDaySummary is one row of the site's index. Field names are part of the
// published index.json contract; do not rename without a schema bump.
type SiteDaySummary struct {
	Source           string  `json:"source"`
	Day              string  `json:"day"`
	Partial          bool    `json:"partial"`
	DropPkts         uint64  `json:"drop_pkts"`
	PassPkts         uint64  `json:"pass_pkts"`
	DropShare        float64 `json:"drop_share"`
	UniqueIPs        int     `json:"unique_attack_ips"`
	PeakHour         string  `json:"peak_hour,omitempty"`
	RestartsDetected int     `json:"restarts_detected"`
	BudgetExhausted  bool    `json:"budget_exhausted"`
	UnknownReasons   bool    `json:"unknown_reasons"`
	TopCategory      string  `json:"top_category,omitempty"`
}

// SiteDayState is the per-day fingerprint kept in state.json so unchanged
// inputs can be skipped on the next invocation.
type SiteDayState struct {
	Size  int64  `json:"size"`
	MTime int64  `json:"mtime_unix"`
	Built bool   `json:"built"`
	Path  string `json:"path"` // absolute path to the gzipped cache file (informational)
}

// SiteState is the on-disk incremental state.
type SiteState struct {
	Schema int                                `json:"schema"`
	Days   map[string]map[string]SiteDayState `json:"days"` // days[source][day]
}

const siteStateSchema = 1

// runSite parses flags, scans the input directory, and rebuilds the static
// site. It is intentionally read-only on inputs and never deletes user data.
func runSite(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("xdplog-report site", flag.ContinueOnError)
	today := time.Now().UTC().Format("2006-01-02")
	if v := os.Getenv("XDPSTATS_TODAY"); v != "" {
		today = v
	}
	opts := &SiteOptions{
		TopN:   20,
		MaskIP: true,
		Today:  today,
	}
	fs.StringVar(&opts.LogsDir, "logs", "", "directory of <source>/nhp_server_xdp-YYYY-MM-DD.log[.gz]")
	fs.StringVar(&opts.OutDir, "out", "", "web root directory (created if missing)")
	fs.StringVar(&opts.StatePath, "state", "", "incremental-build state file (default <out>/state.json; keep it out of the web root)")
	fs.IntVar(&opts.TopN, "top", opts.TopN, "rows in each TOP table")
	fs.StringVar(&opts.GeoCSV, "geo-csv", "", "optional CIDR->country CSV for the region section")
	fs.StringVar(&opts.GeoMMDB, "geo-mmdb", "", "optional GeoLite2-City .mmdb file for the region section")
	fs.BoolVar(&opts.MaskIP, "mask-ip", opts.MaskIP, "truncate source IPs in the public site (default true)")
	fs.BoolVar(&opts.Force, "force", false, "rebuild every day even if inputs are unchanged")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if opts.LogsDir == "" {
		return fmt.Errorf("-logs is required")
	}
	if opts.OutDir == "" {
		return fmt.Errorf("-out is required")
	}
	if opts.TopN < 1 {
		return fmt.Errorf("-top must be >= 1")
	}
	if opts.GeoCSV != "" && opts.GeoMMDB != "" {
		return fmt.Errorf("-geo-csv and -geo-mmdb are mutually exclusive")
	}

	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return fmt.Errorf("create -out: %w", err)
	}
	if opts.StatePath == "" {
		opts.StatePath = filepath.Join(opts.OutDir, "state.json")
	}
	if err := os.MkdirAll(filepath.Dir(opts.StatePath), 0o755); err != nil {
		return fmt.Errorf("create -state dir: %w", err)
	}

	var geo GeoResolver
	if opts.GeoCSV != "" {
		g, err := loadGeoCSV(opts.GeoCSV)
		if err != nil {
			return fmt.Errorf("load geo: %w", err)
		}
		geo = g
	}
	if opts.GeoMMDB != "" {
		g, closer, err := loadGeoMMDB(opts.GeoMMDB)
		if err != nil {
			return fmt.Errorf("load geo mmdb: %w", err)
		}
		defer closer.Close()
		geo = g
	}

	state, err := loadSiteState(opts.StatePath)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	// Snapshot the loaded state so saveSiteState can detect "nothing changed"
	// and skip the rewrite — that is what makes a cron with no new logs a true
	// no-op (no mtime churn on index/state).
	loadedSnapshot := state.snapshot()

	inputs, err := collectSiteInputs(opts.LogsDir)
	if err != nil {
		return err
	}
	if len(inputs) == 0 {
		return fmt.Errorf("no nhp_server_xdp-YYYY-MM-DD.log[.gz] files under %s", opts.LogsDir)
	}

	var summaries []SiteDaySummary
	dirty := false
	for _, in := range inputs {
		summary, wasDirty, err := processSiteDay(opts, geo, in, state)
		if err != nil {
			return fmt.Errorf("%s/%s: %w", in.Source, in.Day, err)
		}
		summaries = append(summaries, summary)
		if wasDirty {
			dirty = true
		}
	}

	if dirty {
		if err := saveSiteState(opts.StatePath, state); err != nil {
			return fmt.Errorf("save state: %w", err)
		}
		if err := renderIndex(opts, summaries); err != nil {
			return fmt.Errorf("render index: %w", err)
		}
	} else {
		// Verify the on-disk state still matches the loaded one. If they
		// differ, the file was edited by hand; rewrite it from our in-memory
		// view so the next run starts clean.
		if !state.equal(loadedSnapshot) {
			if err := saveSiteState(opts.StatePath, state); err != nil {
				return fmt.Errorf("save state: %w", err)
			}
		}
	}

	var n int
	for _, s := range summaries {
		if s.DropPkts > 0 || s.PassPkts > 0 || s.Partial {
			n++
		}
	}
	fmt.Fprintf(stdout, "site: processed %d day(s); %s\n", n, ternary(dirty, "wrote day pages + index", "no input changes, no writes"))
	return nil
}

// ternary returns a when cond is true, b otherwise. Tiny helper for the
// progress message so we don't pull in a dependency.
func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

// siteInput is one (source, day, file) tuple discovered on disk.
type siteInput struct {
	Source string
	Day    string
	Path   string
}

// collectSiteInputs walks <logs>/<source>/ and yields nhp_server_xdp-YYYY-MM-DD.log
// and *.log.gz entries. evaluate/audit files (anything else matching nhp_server_xdp-)
// and files with malformed dates are ignored. The input directory is expected to
// be already filtered to whatever the collector chose to pull.
func collectSiteInputs(logsDir string) ([]siteInput, error) {
	st, err := os.Stat(logsDir)
	if err != nil {
		return nil, fmt.Errorf("-logs: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("-logs %s is not a directory", logsDir)
	}
	sources, err := os.ReadDir(logsDir)
	if err != nil {
		return nil, err
	}
	dateRe := regexp.MustCompile(`^nhp_server_xdp-(\d{4}-\d{2}-\d{2})\.log(?:\.gz)?$`)
	var out []siteInput
	for _, s := range sources {
		if !s.IsDir() {
			continue
		}
		sourceName := s.Name()
		sourceDir := filepath.Join(logsDir, sourceName)
		entries, err := os.ReadDir(sourceDir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			m := dateRe.FindStringSubmatch(name)
			if m == nil {
				continue
			}
			out = append(out, siteInput{
				Source: sourceName,
				Day:    m[1],
				Path:   filepath.Join(sourceDir, name),
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		if out[i].Day != out[j].Day {
			return out[i].Day < out[j].Day
		}
		return out[i].Path < out[j].Path
	})
	return out, nil
}

// processSiteDay analyzes one (source, day) pair, writes its JSON/HTML if the
// input changed (or opts.Force), updates state in place, and returns the
// summary for the index. The bool reports whether this day was rebuilt so
// the caller can decide whether the index and state.json need to be rewritten.
func processSiteDay(opts *SiteOptions, geo GeoResolver, in siteInput, state *SiteState) (SiteDaySummary, bool, error) {
	fi, err := os.Stat(in.Path)
	if err != nil {
		return SiteDaySummary{}, false, err
	}
	prevDays, ok := state.Days[in.Source]
	if !ok {
		prevDays = map[string]SiteDayState{}
		state.Days[in.Source] = prevDays
	}
	prev := prevDays[in.Day]
	unchanged := !opts.Force &&
		prev.Built &&
		prev.Size == fi.Size() &&
		prev.MTime == fi.ModTime().Unix()
	if unchanged {
		return summaryFromStateWithOpts(in, prev, opts.Today, opts.OutDir), false, nil
	}

	an := NewAnalyzer(in.Day, in.Day)
	if err := openAndConsume(an, in.Path, in.Day); err != nil {
		return SiteDaySummary{}, false, err
	}
	rep := an.Build(opts.TopN, geo, opts.MaskIP)
	// Detail pages live at <out>/d/<source>/<day>.html, so the index is two
	// levels up. A single ".." would resolve to <out>/d/index.html, which
	// does not exist.
	rep.BackLink = "../../index.html"
	rep.IsSiteDetail = true

	dayDir := filepath.Join(opts.OutDir, "d", in.Source)
	if err := os.MkdirAll(dayDir, 0o755); err != nil {
		return SiteDaySummary{}, false, err
	}
	if err := writeJSONAtomic(filepath.Join(dayDir, in.Day+".json"), rep); err != nil {
		return SiteDaySummary{}, false, err
	}
	if err := writeHTMLAtomic(filepath.Join(dayDir, in.Day+".html"), rep); err != nil {
		return SiteDaySummary{}, false, err
	}

	prevDays[in.Day] = SiteDayState{
		Size:  fi.Size(),
		MTime: fi.ModTime().Unix(),
		Built: true,
		Path:  in.Path,
	}

	return summaryFromReport(in, rep, opts.Today), true, nil
}

// snapshot returns a deep copy of state for change detection.
func (s *SiteState) snapshot() SiteState {
	out := SiteState{Schema: s.Schema, Days: map[string]map[string]SiteDayState{}}
	for src, days := range s.Days {
		cp := map[string]SiteDayState{}
		for d, st := range days {
			cp[d] = st
		}
		out.Days[src] = cp
	}
	return out
}

// equal reports whether s and other are byte-equal at the JSON level. We
// compare by re-encoding to JSON because the in-memory maps have no
// stable iteration order.
func (s *SiteState) equal(other SiteState) bool {
	a, err := json.Marshal(s)
	if err != nil {
		return false
	}
	b, err := json.Marshal(&other)
	if err != nil {
		return false
	}
	return string(a) == string(b)
}

// summaryFromReport rebuilds a SiteDaySummary from a freshly built Report.
func summaryFromReport(in siteInput, rep *Report, today string) SiteDaySummary {
	return SiteDaySummary{
		Source:           in.Source,
		Day:              in.Day,
		Partial:          in.Day >= today,
		DropPkts:         rep.DropPkts,
		PassPkts:         rep.PassPkts,
		DropShare:        rep.DropShare,
		UniqueIPs:        rep.UniqueIPs,
		PeakHour:         rep.PeakHour,
		RestartsDetected: len(rep.Restarts),
		BudgetExhausted:  len(rep.BudgetExhausted) > 0,
		UnknownReasons:   len(rep.UnknownRsns) > 0,
		TopCategory:      topCategoryOf(rep),
	}
}

// summaryFromStateWithOpts rebuilds the summary used in index.json when a
// day's inputs are unchanged. The numeric totals come from the previously
// written JSON file (which is the source of truth on disk); we re-read it
// because state.json deliberately does not mirror report numbers.
func summaryFromStateWithOpts(in siteInput, st SiteDayState, today, outDir string) SiteDaySummary {
	sum := SiteDaySummary{
		Source:  in.Source,
		Day:     in.Day,
		Partial: in.Day >= today,
	}
	path := filepath.Join(outDir, "d", in.Source, in.Day+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		return sum
	}
	var rep Report
	if err := json.Unmarshal(b, &rep); err != nil {
		return sum
	}
	sum.DropPkts = rep.DropPkts
	sum.PassPkts = rep.PassPkts
	sum.DropShare = rep.DropShare
	sum.UniqueIPs = rep.UniqueIPs
	sum.PeakHour = rep.PeakHour
	sum.RestartsDetected = len(rep.Restarts)
	sum.BudgetExhausted = len(rep.BudgetExhausted) > 0
	sum.UnknownReasons = len(rep.UnknownRsns) > 0
	sum.TopCategory = topCategoryOf(&rep)
	return sum
}

// topCategoryOf returns the leading category from a freshly built Report.
// Categories are sorted by exact DROP packets in analyze.go so the first
// row is the largest by design.
func topCategoryOf(rep *Report) string {
	if len(rep.Categories) == 0 {
		return ""
	}
	return rep.Categories[0].Category
}

// openAndConsume opens the (possibly gzipped) log and feeds it to the
// Analyzer. It centralizes the .gz suffix check that main.go also does.
func openAndConsume(an *Analyzer, path, day string) error {
	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fh.Close()
	var r io.Reader = fh
	if strings.HasSuffix(path, ".gz") {
		gz, err := gzip.NewReader(fh)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		defer gz.Close()
		r = gz
	}
	if err := an.Consume(r, day); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// writeJSONAtomic writes data as JSON via a sibling temp file then rename.
func writeJSONAtomic(path string, rep *Report) error {
	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return writeFileAtomic(path, b, 0o644)
}

// writeHTMLAtomic writes the rendered HTML page via a sibling temp file then rename.
func writeHTMLAtomic(path string, rep *Report) error {
	var sb strings.Builder
	if err := renderHTML(&sb, rep); err != nil {
		return err
	}
	return writeFileAtomic(path, []byte(sb.String()), 0o644)
}

// writeFileAtomic writes b to path via <path>.tmp.<pid> + os.Rename, so a
// concurrent reader never sees a half-written file.
func writeFileAtomic(path string, b []byte, mode fs.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp.*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if tmpName != "" {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	tmpName = ""
	return nil
}

// loadSiteState reads state.json from path. Missing file is not an error:
// the state is initialized empty. A corrupt file is reported and aborts the
// run to avoid silently rewriting away an inconsistent state.
func loadSiteState(path string) (*SiteState, error) {
	st := &SiteState{
		Schema: siteStateSchema,
		Days:   map[string]map[string]SiteDayState{},
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return nil, err
	}
	var loaded SiteState
	if err := json.Unmarshal(b, &loaded); err != nil {
		return nil, fmt.Errorf("%s: corrupt: %w", path, err)
	}
	if loaded.Days == nil {
		loaded.Days = map[string]map[string]SiteDayState{}
	}
	if loaded.Schema != siteStateSchema {
		return nil, fmt.Errorf("%s: state schema %d, expected %d (rerun with -force or delete state.json)", path, loaded.Schema, siteStateSchema)
	}
	return &loaded, nil
}

// saveSiteState writes state.json atomically.
func saveSiteState(path string, st *SiteState) error {
	st.Schema = siteStateSchema
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return writeFileAtomic(path, b, 0o644)
}
