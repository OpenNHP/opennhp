// xdplog-report analyses nhp_server_xdp-YYYY-MM-DD.log files written by
// nhp-serverd's XDP ingress filter and renders an attack report. It is fully
// offline: it never resolves or looks up anything over the network.
package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(s string) error { *m = append(*m, s); return nil }

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "xdplog-report:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("xdplog-report", flag.ContinueOnError)
	var in multiFlag
	fs.Var(&in, "in", "log file, glob or directory (repeatable; .gz supported)")
	geoPath := fs.String("geo-csv", "", "optional CIDR->country CSV (cidr,country[,city]) for the region section")
	geoMMDB := fs.String("geo-mmdb", "", "optional GeoLite2-City .mmdb file for the region section")
	top := fs.Int("top", 10, "rows in each TOP table")
	format := fs.String("format", "md", "md | json | csv | html")
	out := fs.String("out", "", "output file (default stdout)")
	since := fs.String("since", "", "first day to include, YYYY-MM-DD")
	until := fs.String("until", "", "last day to include, YYYY-MM-DD")
	date := fs.String("date", "", "day to assume for files whose name has no YYYY-MM-DD")
	mask := fs.Bool("mask-ip", false, "truncate source IPs (a.b.c.x, IPv6 /48) for sharing")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if len(in) == 0 {
		return fmt.Errorf("at least one -in is required")
	}
	if *top < 1 {
		return fmt.Errorf("-top must be >= 1")
	}
	render := map[string]func(io.Writer, *Report) error{
		"md": renderMarkdown, "json": renderJSON, "csv": renderCSV, "html": renderHTML,
	}[*format]
	if render == nil {
		return fmt.Errorf("unknown -format %q", *format)
	}
	if *geoPath != "" && *geoMMDB != "" {
		return fmt.Errorf("-geo-csv and -geo-mmdb are mutually exclusive")
	}

	var geo GeoResolver
	if *geoPath != "" {
		g, err := loadGeoCSV(*geoPath)
		if err != nil {
			return fmt.Errorf("load geo: %w", err)
		}
		geo = g
	}
	if *geoMMDB != "" {
		g, closer, err := loadGeoMMDB(*geoMMDB)
		if err != nil {
			return fmt.Errorf("load geo mmdb: %w", err)
		}
		defer closer.Close()
		geo = g
	}

	files, err := expandInputs(in)
	if err != nil {
		return err
	}
	type dated struct{ path, day string }
	var list []dated
	for _, f := range files {
		day, ok := dateFromFilename(filepath.Base(f))
		if !ok {
			if *date == "" {
				return fmt.Errorf("%s: no YYYY-MM-DD in file name; pass -date", f)
			}
			day = *date
		}
		list = append(list, dated{f, day})
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].day != list[j].day {
			return list[i].day < list[j].day
		}
		return list[i].path < list[j].path
	})

	an := NewAnalyzer(*since, *until)
	for _, f := range list {
		if err := consumeFile(an, f.path, f.day); err != nil {
			return err
		}
	}

	w := stdout
	if *out != "" {
		fh, err := os.Create(*out)
		if err != nil {
			return err
		}
		defer fh.Close()
		w = fh
	}
	return render(w, an.Build(*top, geo, *mask))
}

func consumeFile(an *Analyzer, path, day string) error {
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

// expandInputs resolves files, globs and directories into a de-duplicated list.
func expandInputs(in []string) ([]string, error) {
	seen := map[string]struct{}{}
	var out []string
	add := func(p string) {
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	for _, p := range in {
		matches, err := filepath.Glob(p)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			matches = []string{p}
		}
		for _, m := range matches {
			st, err := os.Stat(m)
			if err != nil {
				return nil, err
			}
			if !st.IsDir() {
				add(m)
				continue
			}
			ents, err := os.ReadDir(m)
			if err != nil {
				return nil, err
			}
			for _, e := range ents {
				n := e.Name()
				if !e.IsDir() && strings.HasPrefix(n, "nhp_server_xdp") && (strings.HasSuffix(n, ".log") || strings.HasSuffix(n, ".log.gz")) {
					add(filepath.Join(m, n))
				}
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no log files found")
	}
	return out, nil
}
