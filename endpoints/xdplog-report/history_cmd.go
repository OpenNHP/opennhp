package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// The `history` subcommand: turn a saved copy of the stats homepage into the
// history.json snapshot the site reads.
//
// The site publishes nothing before 2026-10-16 (defaultSiteSince) and the
// collector's cache does not reach back past the daemon's 14-day retention, so
// the totals for the days before that floor live nowhere. A copy of the
// homepage saved while it was still current is the only carrier of them, and
// this command is what freezes one into the file `xdplog-report site -history`
// reads.
//
// The generator refuses to write a snapshot that does not add up: the rows it
// parsed have to sum to the page's own all-time total, and the page's stated
// counting start has to be its earliest row. A parsing slip that quietly
// changed the baseline would otherwise be published as the site's history,
// forever.

// historyPage is what a saved page (or a published index) yields before it is
// checked against itself.
type historyPage struct {
	Blocked   uint64
	Sources   int
	Countries int
	Since     string
	Rows      []SiteDaySummary
}

var (
	// sinceRe pulls the counting start out of the page's note paragraph.
	sinceRe = regexp.MustCompile(`Since\s+(\d{4}-\d{2}-\d{2})`)
	// totalNumberRe pulls the three all-time figures out of totalnumber.js.
	totalNumberRe = regexp.MustCompile(`const NHP_TOTAL_ATTACK_(BLOCKED|SOURCE_IPS|COUNTRIES) = (\d+);`)
	// countingStartRe pulls the counting start out of totalnumber.js.
	countingStartRe = regexp.MustCompile(`// Counting starts: (\d{4}-\d{2}-\d{2}) \(UTC\)\.`)
	// sourceNameRe is what a source directory name looks like. It is used to
	// accept a source the site has no display label for: there the label and
	// the directory name are the same string.
	sourceNameRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)
)

// sourceDirNames maps a source's display label back to the directory name the
// site indexes it under. The site shows "nhp-server" and stores "server", and
// a snapshot has to name the day the way the index does — otherwise the
// snapshot would freeze a day under a name nothing matches.
var sourceDirNames = buildSourceDirNames()

func buildSourceDirNames() map[string]string {
	m := make(map[string]string, len(sourceLabels))
	for dir, label := range sourceLabels {
		m[label] = dir
	}
	return m
}

// runHistory parses flags and writes a history.json snapshot.
func runHistory(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("xdplog-report history", flag.ContinueOnError)
	htmlPath := fs.String("html", "", "saved stats index page (HTML) to derive the snapshot from")
	indexPath := fs.String("index", "", "published index.json to derive the snapshot from")
	totalsPath := fs.String("totals", "", "published totalnumber.js (required with -index)")
	outPath := fs.String("out", "", "output history.json path (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *outPath == "" {
		return fmt.Errorf("-out is required")
	}
	switch {
	case *htmlPath == "" && *indexPath == "":
		return fmt.Errorf("one of -html or -index is required")
	case *htmlPath != "" && *indexPath != "":
		return fmt.Errorf("-html and -index are mutually exclusive")
	case *indexPath != "" && *totalsPath == "":
		return fmt.Errorf("-index needs -totals (the all-time figures the index alone does not carry)")
	}

	var page *historyPage
	var err error
	if *htmlPath != "" {
		page, err = historyPageFromHTML(*htmlPath)
	} else {
		page, err = historyPageFromIndex(*indexPath, *totalsPath)
	}
	if err != nil {
		return err
	}

	hist, err := buildSiteHistory(page)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(hist, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if err := writeFileAtomic(*outPath, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", *outPath, err)
	}
	fmt.Fprintf(stdout, "history: %d day(s) from %s, %d blocked packet(s), %d source IP(s), %d country/countries, counting from %s\n",
		len(hist.Days), *outPath, hist.Totals.Blocked, hist.Totals.SourceIPs, hist.Totals.Countries, hist.Since)
	return nil
}

// historyPageFromHTML parses a saved copy of the stats homepage.
func historyPageFromHTML(path string) (*historyPage, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	page, err := historyPageFromHTMLBytes(b)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return page, nil
}

// historyPageFromHTMLBytes parses a saved copy of the stats homepage.
func historyPageFromHTMLBytes(b []byte) (*historyPage, error) {
	doc, err := html.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	page := &historyPage{}
	htmlElements(doc, func(n *html.Node) {
		switch id := htmlAttr(n, "id"); id {
		case "total-attacks", "total-sources", "total-countries":
			v, err := heroNumber(n)
			if err != nil {
				return
			}
			switch id {
			case "total-attacks":
				page.Blocked = v
			case "total-sources":
				page.Sources = int(v)
			case "total-countries":
				page.Countries = int(v)
			}
		}
	})

	// The counting start lives in the note paragraph under the hero. A page
	// without one leaves it empty, and the earliest row then names it.
	htmlElements(doc, func(n *html.Node) {
		if page.Since != "" || n.Data != "p" || !htmlHasClass(n, "note") {
			return
		}
		if m := sinceRe.FindStringSubmatch(htmlText(n)); m != nil {
			page.Since = m[1]
		}
	})

	for _, cells := range historyTableRows(doc) {
		row, err := historyRowFromCells(cells)
		if err != nil {
			return nil, err
		}
		page.Rows = append(page.Rows, row)
	}
	return page, nil
}

// historyPageFromIndex parses a published index.json together with the
// totalnumber.js that carries the all-time figures the index does not.
func historyPageFromIndex(indexPath, totalsPath string) (*historyPage, error) {
	b, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", indexPath, err)
	}
	var rows []SiteDaySummary
	if err = json.Unmarshal(b, &rows); err != nil {
		return nil, fmt.Errorf("%s: %w", indexPath, err)
	}
	for _, r := range rows {
		if r.History {
			return nil, fmt.Errorf("%s already lists history rows; a snapshot must be taken from a site published without one", indexPath)
		}
	}

	tb, err := os.ReadFile(totalsPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", totalsPath, err)
	}
	page := &historyPage{Rows: rows}
	seen := map[string]bool{}
	for _, m := range totalNumberRe.FindAllStringSubmatch(string(tb), -1) {
		v, err := strconv.ParseUint(m[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", totalsPath, m[1], err)
		}
		switch m[1] {
		case "BLOCKED":
			page.Blocked = v
		case "SOURCE_IPS":
			page.Sources = int(v)
		case "COUNTRIES":
			page.Countries = int(v)
		}
		seen[m[1]] = true
	}
	for _, name := range []string{"BLOCKED", "SOURCE_IPS", "COUNTRIES"} {
		if !seen[name] {
			return nil, fmt.Errorf("%s: NHP_TOTAL_ATTACK_%s not found", totalsPath, name)
		}
	}
	if m := countingStartRe.FindStringSubmatch(string(tb)); m != nil {
		page.Since = m[1]
	}
	return page, nil
}

// buildSiteHistory checks the parsed page against itself and returns the
// snapshot to write.
//
// The two checks are the whole point of the generator: the rows have to sum to
// the page's own all-time total, and the page's counting start has to be its
// earliest row. Either one failing means the page was not read the way it was
// meant to be read, and a snapshot built from it would be a wrong baseline
// published as the site's history.
func buildSiteHistory(page *historyPage) (*SiteHistory, error) {
	if len(page.Rows) == 0 {
		return nil, fmt.Errorf("no daily rows found")
	}
	var daySum uint64
	earliest := ""
	for i, r := range page.Rows {
		if _, err := time.Parse("2006-01-02", r.Day); err != nil {
			return nil, fmt.Errorf("row %d has an invalid date %q", i, r.Day)
		}
		if r.Source == "" {
			return nil, fmt.Errorf("row %d (%s) has no source", i, r.Day)
		}
		if r.DropShare < 0 || r.DropShare > 1.000001 {
			return nil, fmt.Errorf("row %d (%s) has drop_share %g, want 0..1", i, r.Day, r.DropShare)
		}
		daySum += r.DropPkts
		if earliest == "" || r.Day < earliest {
			earliest = r.Day
		}
	}
	if daySum != page.Blocked {
		return nil, fmt.Errorf("the page's rows add up to %d DROP packets but its all-time total says %d",
			daySum, page.Blocked)
	}
	since := page.Since
	if since == "" {
		since = earliest
	}
	if since != earliest {
		return nil, fmt.Errorf("the page counts from %s but its earliest row is %s", since, earliest)
	}

	rows := append([]SiteDaySummary(nil), page.Rows...)
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Day != rows[j].Day {
			return rows[i].Day < rows[j].Day
		}
		return rows[i].Source < rows[j].Source
	})
	return &SiteHistory{
		Schema: historySchema,
		Since:  since,
		Totals: HistoryTotals{
			Blocked:   page.Blocked,
			SourceIPs: page.Sources,
			Countries: page.Countries,
		},
		Days: rows,
	}, nil
}

// historyRowFromCells turns one table row of the homepage into an index row.
//
// The snapshot has no pass_pkts column and no per-drop events: the page only
// ever showed the day's summary. pass_pkts therefore stays 0, which on a
// history row means "not recorded" rather than "nothing passed" — the row
// carries the `history` flag for exactly that reason.
func historyRowFromCells(cells []*html.Node) (SiteDaySummary, error) {
	if len(cells) < 8 {
		return SiteDaySummary{}, fmt.Errorf("table row has %d cells, want 8", len(cells))
	}
	day := htmlText(cells[0])
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return SiteDaySummary{}, fmt.Errorf("row date %q is not YYYY-MM-DD", day)
	}
	source, err := historySourceName(htmlText(cells[1]))
	if err != nil {
		return SiteDaySummary{}, err
	}
	drop, err := strconv.ParseUint(strings.ReplaceAll(htmlText(cells[2]), ",", ""), 10, 64)
	if err != nil {
		return SiteDaySummary{}, fmt.Errorf("row %s: DROP packet count: %w", day, err)
	}
	share, err := historyShare(htmlText(cells[3]))
	if err != nil {
		return SiteDaySummary{}, fmt.Errorf("row %s: %w", day, err)
	}
	unique, err := strconv.Atoi(strings.ReplaceAll(htmlText(cells[4]), ",", ""))
	if err != nil {
		return SiteDaySummary{}, fmt.Errorf("row %s: unique source count: %w", day, err)
	}

	row := SiteDaySummary{
		Source:      source,
		Day:         day,
		DropPkts:    drop,
		DropShare:   share,
		UniqueIPs:   unique,
		PeakHour:    htmlText(cells[6]),
		TopCategory: htmlText(cells[5]),
	}
	// The page shows a chip per flag and nothing at all for a day without
	// any. A restart chip says a restart happened, not how many: the
	// snapshot records the count as 1.
	for _, flag := range historyFlags(cells[7]) {
		switch flag {
		case "partial":
			row.Partial = true
		case "budget":
			row.BudgetExhausted = true
		case "restart":
			row.RestartsDetected = 1
		case "unknown":
			row.UnknownReasons = true
		default:
			return SiteDaySummary{}, fmt.Errorf("row %s: unknown flag %q", day, flag)
		}
	}
	return row, nil
}

// historySourceName maps a source's display label back to the directory name
// the index uses.
func historySourceName(label string) (string, error) {
	if dir, ok := sourceDirNames[label]; ok {
		return dir, nil
	}
	// A source the site has no label for shows its directory name as it is.
	if sourceNameRe.MatchString(label) {
		return label, nil
	}
	return "", fmt.Errorf("cannot map source %q to a directory name", label)
}

// historyShare turns a rendered percentage into a fraction.
//
// The result is rounded to the page's own precision — two decimals of a
// percentage is four decimals of a fraction — so the snapshot does not carry
// the binary noise of the division and the homepage renders the same
// percentage it rendered before.
func historyShare(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	s = strings.TrimSuffix(s, "%")
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("share %q: %w", s, err)
	}
	return math.Round((v/100)*10000) / 10000, nil
}

// historyFlags returns the flag chips in a flags cell.
func historyFlags(cell *html.Node) []string {
	var out []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" && htmlHasClass(n, "tag") {
			out = append(out, htmlText(n))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(cell)
	return out
}

// historyTableRows returns the cells of every row in every table body.
func historyTableRows(doc *html.Node) [][]*html.Node {
	var out [][]*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tbody" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type != html.ElementNode || c.Data != "tr" {
					continue
				}
				var cells []*html.Node
				for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
					if cc.Type == html.ElementNode && cc.Data == "td" {
						cells = append(cells, cc)
					}
				}
				out = append(out, cells)
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return out
}

// htmlElements calls fn for every element in the tree.
func htmlElements(n *html.Node, fn func(*html.Node)) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			fn(c)
		}
		htmlElements(c, fn)
	}
}

// htmlText returns an element's text with runs of whitespace collapsed.
func htmlText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
			b.WriteString(" ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.Join(strings.Fields(b.String()), " ")
}

// htmlAttr returns an attribute's value, or the empty string.
func htmlAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// htmlHasClass reports whether an element carries a class.
func htmlHasClass(n *html.Node, class string) bool {
	for _, c := range strings.Fields(htmlAttr(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

// heroNumber reads one of the homepage's hero figures. The counter's
// data-total attribute is authoritative — it is the number the page counts
// up to — and the rendered text is the fallback for a page saved with the
// attribute stripped.
func heroNumber(n *html.Node) (uint64, error) {
	if v := strings.TrimSpace(htmlAttr(n, "data-total")); v != "" {
		return strconv.ParseUint(v, 10, 64)
	}
	txt := strings.ReplaceAll(htmlText(n), ",", "")
	txt = strings.TrimSpace(txt)
	if txt == "" {
		return 0, nil
	}
	return strconv.ParseUint(txt, 10, 64)
}
