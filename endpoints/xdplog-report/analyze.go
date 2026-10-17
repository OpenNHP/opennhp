package main

import (
	"bufio"
	"fmt"
	"io"
	"net/netip"
	"sort"
	"strings"
)

// maxPortsPerIP bounds per-source destination-port tracking.
const maxPortsPerIP = 16

type ipAgg struct {
	events    int
	cats      map[Category]struct{}
	first     string // "YYYY-MM-DD HH:MM:SS"
	last      string
	ports     map[int]int
	portsFull bool
}

type reasonAgg struct {
	verdict string
	pkts    uint64 // exact, from STAT
	events  uint64 // sampled, from per-packet lines
}

type statKey struct{ day, minute, verdict, reason string }

// Analyzer streams log lines and accumulates the report state.
type Analyzer struct {
	since, until string // inclusive YYYY-MM-DD, optional

	servers   map[string]struct{}
	firstTS   string
	lastTS    string
	reasons   map[string]*reasonAgg
	ips       map[netip.Addr]*ipAgg
	ports     map[int]int
	protos    map[string]int
	hourly    map[string]uint64 // "YYYY-MM-DD HH" -> exact DROP pkts
	daily     map[string]uint64
	seenStat  map[statKey]struct{}
	lastTotal map[string]uint64 // verdict/reason -> last TOTAL

	budgetDays  map[string]string // day -> clock when exhausted
	badLines    int
	dupStats    int
	restarts    []string
	unknown     map[string]struct{}
	v6Events    int
	totalEvents int
}

func NewAnalyzer(since, until string) *Analyzer {
	return &Analyzer{
		since: since, until: until,
		servers:    map[string]struct{}{},
		reasons:    map[string]*reasonAgg{},
		ips:        map[netip.Addr]*ipAgg{},
		ports:      map[int]int{},
		protos:     map[string]int{},
		hourly:     map[string]uint64{},
		daily:      map[string]uint64{},
		seenStat:   map[statKey]struct{}{},
		lastTotal:  map[string]uint64{},
		budgetDays: map[string]string{},
		unknown:    map[string]struct{}{},
	}
}

// Consume reads one log file; day is the date taken from its name.
// Files must be fed in chronological order for restart detection.
func (a *Analyzer) Consume(r io.Reader, day string) error {
	if (a.since != "" && day < a.since) || (a.until != "" && day > a.until) {
		return nil
	}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		rec, err := parseLine(line)
		if err != nil {
			a.badLines++
			continue
		}
		a.add(rec, day)
	}
	return sc.Err()
}

func (a *Analyzer) add(rec Record, day string) {
	ts := day + " " + rec.Clock
	a.servers[rec.ServerID] = struct{}{}
	if a.firstTS == "" || ts < a.firstTS {
		a.firstTS = ts
	}
	if ts > a.lastTS {
		a.lastTS = ts
	}

	switch rec.Kind {
	case kindBudget:
		if _, ok := a.budgetDays[day]; !ok {
			a.budgetDays[day] = rec.Clock
		}
	case kindStat:
		key := statKey{day, rec.Clock[:5], rec.Verdict, rec.Reason}
		if _, dup := a.seenStat[key]; dup {
			a.dupStats++
			return
		}
		a.seenStat[key] = struct{}{}

		// TOTAL is cumulative since daemon start; a drop means a restart.
		ck := rec.Verdict + "/" + rec.Reason
		if prev, ok := a.lastTotal[ck]; ok && rec.Total < prev {
			a.restarts = append(a.restarts, fmt.Sprintf("%s (%s TOTAL %d -> %d)", ts, ck, prev, rec.Total))
		}
		a.lastTotal[ck] = rec.Total

		ra := a.reason(rec.Reason, rec.Verdict)
		ra.pkts += rec.Pkts
		if rec.Verdict == "DROP" {
			a.noteUnknown(rec.Reason)
			a.hourly[day+" "+rec.Clock[:2]] += rec.Pkts
			a.daily[day] += rec.Pkts
		}
	case kindEvent:
		a.totalEvents++
		a.reason(rec.Reason, rec.Verdict).events++
		if rec.Verdict != "DROP" {
			return
		}
		a.noteUnknown(rec.Reason)
		if rec.Src.Is6() {
			a.v6Events++
		}
		cat, _ := classify(rec.Reason)
		ia := a.ips[rec.Src]
		if ia == nil {
			ia = &ipAgg{cats: map[Category]struct{}{}, ports: map[int]int{}, first: ts}
			a.ips[rec.Src] = ia
		}
		ia.events++
		ia.cats[cat] = struct{}{}
		if ts < ia.first {
			ia.first = ts
		}
		if ts > ia.last {
			ia.last = ts
		}
		if _, ok := ia.ports[rec.DstPort]; ok || len(ia.ports) < maxPortsPerIP {
			ia.ports[rec.DstPort]++
		} else {
			ia.portsFull = true
		}
		a.ports[rec.DstPort]++
		a.protos[rec.Proto]++
	}
}

func (a *Analyzer) reason(name, verdict string) *reasonAgg {
	ra := a.reasons[name]
	if ra == nil {
		ra = &reasonAgg{verdict: verdict}
		a.reasons[name] = ra
	}
	return ra
}

func (a *Analyzer) noteUnknown(reason string) {
	if _, ok := classify(reason); !ok {
		a.unknown[reason] = struct{}{}
	}
}

// ---- report model ----

type Count struct {
	Key   string  `json:"key"`
	Count uint64  `json:"count"`
	Share float64 `json:"share"`
}

type ReasonRow struct {
	Reason   string  `json:"reason"`
	Category string  `json:"category"`
	Pkts     uint64  `json:"pkts"`
	PktShare float64 `json:"pkt_share"`
	Events   uint64  `json:"sampled_events"`
	EvShare  float64 `json:"sampled_share"`
	NoDetail bool    `json:"no_source_detail"`
}

type CategoryRow struct {
	Category string  `json:"category"`
	Pkts     uint64  `json:"pkts"`
	Share    float64 `json:"share"`
	Events   uint64  `json:"sampled_events"`
}

type IPRow struct {
	IP         string  `json:"ip"`
	Events     int     `json:"events"`
	Share      float64 `json:"share"`
	Categories int     `json:"categories"`
	First      string  `json:"first_seen"`
	Last       string  `json:"last_seen"`
	TopPorts   []int   `json:"top_dst_ports"`
	Country    string  `json:"country,omitempty"`
	City       string  `json:"city,omitempty"`
}

type GeoRow struct {
	Country string  `json:"country"`
	Events  int     `json:"events"`
	IPs     int     `json:"unique_ips"`
	Share   float64 `json:"share"`
}

type Report struct {
	Schema  int      `json:"schema"`
	Servers []string `json:"server_ids"`
	From    string   `json:"from"`
	To      string   `json:"to"`
	TopN    int      `json:"top_n"`

	TotalPkts    uint64  `json:"total_pkts"`
	DropPkts     uint64  `json:"drop_pkts"`
	PassPkts     uint64  `json:"pass_pkts"`
	DropShare    float64 `json:"drop_share"`
	DropEvents   uint64  `json:"drop_sampled_events"`
	SamplingRate float64 `json:"drop_sampling_rate"`
	UniqueIPs    int     `json:"unique_attack_ips"`
	ReservedIPs  int     `json:"private_reserved_ips"`

	BudgetExhausted map[string]string `json:"budget_exhausted,omitempty"`

	Reasons    []ReasonRow   `json:"by_reason"`
	Categories []CategoryRow `json:"by_category"`
	TopIPs     []IPRow       `json:"top_ips"`
	GeoEnabled bool          `json:"geo_enabled"`
	TopGeo     []GeoRow      `json:"top_geo"`
	TopPorts   []Count       `json:"top_dst_ports"`
	Protos     []Count       `json:"protocols"`
	Hourly     []Count       `json:"hourly_drop_pkts"`
	Daily      []Count       `json:"daily_drop_pkts"`
	PeakHour   string        `json:"peak_hour,omitempty"`

	BadLines    int      `json:"bad_lines"`
	DupStats    int      `json:"duplicate_stat_lines"`
	Restarts    []string `json:"restarts_detected"`
	UnknownRsns []string `json:"unknown_reasons"`
	IPv6Share   float64  `json:"ipv6_event_share"`
}

func ratio(n, d uint64) float64 {
	if d == 0 {
		return 0
	}
	return float64(n) / float64(d)
}

func sortedKeys[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func topCounts[K comparable](m map[K]int, n int, name func(K) string) []Count {
	var total uint64
	rows := make([]Count, 0, len(m))
	for k, v := range m {
		rows = append(rows, Count{Key: name(k), Count: uint64(v)})
		total += uint64(v)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].Key < rows[j].Key
	})
	if len(rows) > n {
		rows = rows[:n]
	}
	for i := range rows {
		rows[i].Share = ratio(rows[i].Count, total)
	}
	return rows
}

// maskIP truncates an address for sharing: a.b.c.x / first 48 bits of IPv6.
func maskIP(a netip.Addr) string {
	if a.Is4() {
		b := a.As4()
		return fmt.Sprintf("%d.%d.%d.x", b[0], b[1], b[2])
	}
	p, _ := a.Prefix(48)
	return p.String()
}

// Build turns the accumulated state into a Report.
func (a *Analyzer) Build(topN int, geo GeoResolver, mask bool) *Report {
	rep := &Report{
		Schema: 1, TopN: topN, Servers: sortedKeys(a.servers),
		From: a.firstTS, To: a.lastTS,
		BudgetExhausted: a.budgetDays, GeoEnabled: geo != nil,
		BadLines: a.badLines, DupStats: a.dupStats,
		Restarts: a.restarts, UnknownRsns: sortedKeys(a.unknown),
		Reasons: []ReasonRow{}, Categories: []CategoryRow{}, TopIPs: []IPRow{},
		TopGeo: []GeoRow{}, Hourly: []Count{}, Daily: []Count{},
	}
	if rep.Restarts == nil {
		rep.Restarts = []string{}
	}
	if len(a.budgetDays) == 0 {
		rep.BudgetExhausted = nil
	}

	// Type distribution.
	var dropEvents uint64
	for _, ra := range a.reasons {
		rep.TotalPkts += ra.pkts
		if ra.verdict == "DROP" {
			rep.DropPkts += ra.pkts
			dropEvents += ra.events
		} else {
			rep.PassPkts += ra.pkts
		}
	}
	rep.DropShare = ratio(rep.DropPkts, rep.TotalPkts)
	rep.DropEvents = dropEvents
	rep.SamplingRate = ratio(dropEvents, rep.DropPkts)
	catAgg := map[Category]*CategoryRow{}
	for name, ra := range a.reasons {
		if ra.verdict != "DROP" {
			continue
		}
		cat, _ := classify(name)
		rep.Reasons = append(rep.Reasons, ReasonRow{
			Reason: name, Category: string(cat), Pkts: ra.pkts,
			PktShare: ratio(ra.pkts, rep.DropPkts), Events: ra.events,
			EvShare:  ratio(ra.events, dropEvents),
			NoDetail: ra.pkts > 0 && ra.events == 0,
		})
		c := catAgg[cat]
		if c == nil {
			c = &CategoryRow{Category: string(cat)}
			catAgg[cat] = c
		}
		c.Pkts += ra.pkts
		c.Events += ra.events
	}
	sort.Slice(rep.Reasons, func(i, j int) bool {
		x, y := rep.Reasons[i], rep.Reasons[j]
		if x.Pkts != y.Pkts {
			return x.Pkts > y.Pkts
		}
		return x.Reason < y.Reason
	})
	for _, c := range catAgg {
		c.Share = ratio(c.Pkts, rep.DropPkts)
		rep.Categories = append(rep.Categories, *c)
	}
	sort.Slice(rep.Categories, func(i, j int) bool {
		x, y := rep.Categories[i], rep.Categories[j]
		if x.Pkts != y.Pkts {
			return x.Pkts > y.Pkts
		}
		return x.Category < y.Category
	})

	// Sources: ranked by sampled events, ties by address for stable output.
	type ipEntry struct {
		addr netip.Addr
		agg  *ipAgg
	}
	var ranked []ipEntry
	geoAgg := map[string]*GeoRow{}
	geoIPs := map[string]map[netip.Addr]struct{}{}
	var geoEvents int
	for addr, agg := range a.ips {
		if isReserved(addr) {
			rep.ReservedIPs++
		} else {
			rep.UniqueIPs++
		}
		ranked = append(ranked, ipEntry{addr, agg})
		if geo == nil || isReserved(addr) {
			continue
		}
		country := "Unknown"
		if g, ok := geo.Lookup(addr); ok && g.Country != "" {
			country = g.Country
		}
		row := geoAgg[country]
		if row == nil {
			row = &GeoRow{Country: country}
			geoAgg[country] = row
			geoIPs[country] = map[netip.Addr]struct{}{}
		}
		row.Events += agg.events
		geoIPs[country][addr] = struct{}{}
		geoEvents += agg.events
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].agg.events != ranked[j].agg.events {
			return ranked[i].agg.events > ranked[j].agg.events
		}
		return ranked[i].addr.Less(ranked[j].addr)
	})
	for i, e := range ranked {
		if i >= topN {
			break
		}
		row := IPRow{
			IP: e.addr.String(), Events: e.agg.events,
			Share:      ratio(uint64(e.agg.events), dropEvents),
			Categories: len(e.agg.cats), First: e.agg.first, Last: e.agg.last,
		}
		if isReserved(e.addr) {
			row.Country = "Private/Reserved"
		} else if geo != nil {
			if g, ok := geo.Lookup(e.addr); ok {
				row.Country, row.City = g.Country, g.City
			}
		}
		if mask {
			row.IP = maskIP(e.addr)
		}
		row.TopPorts = topPorts(e.agg.ports, 3)
		rep.TopIPs = append(rep.TopIPs, row)
	}

	for c, row := range geoAgg {
		row.IPs = len(geoIPs[c])
		row.Share = ratio(uint64(row.Events), uint64(geoEvents))
		rep.TopGeo = append(rep.TopGeo, *row)
	}
	sort.Slice(rep.TopGeo, func(i, j int) bool {
		x, y := rep.TopGeo[i], rep.TopGeo[j]
		if x.Events != y.Events {
			return x.Events > y.Events
		}
		return x.Country < y.Country
	})
	if len(rep.TopGeo) > topN {
		rep.TopGeo = rep.TopGeo[:topN]
	}

	rep.TopPorts = topCounts(a.ports, topN, func(p int) string { return fmt.Sprint(p) })
	rep.Protos = topCounts(a.protos, len(a.protos), func(s string) string { return s })

	var peak uint64
	for _, h := range sortedKeys(a.hourly) {
		rep.Hourly = append(rep.Hourly, Count{Key: h + ":00", Count: a.hourly[h], Share: ratio(a.hourly[h], rep.DropPkts)})
		if a.hourly[h] > peak {
			peak, rep.PeakHour = a.hourly[h], h+":00"
		}
	}
	for _, d := range sortedKeys(a.daily) {
		rep.Daily = append(rep.Daily, Count{Key: d, Count: a.daily[d], Share: ratio(a.daily[d], rep.DropPkts)})
	}
	rep.IPv6Share = ratio(uint64(a.v6Events), dropEvents)
	return rep
}

func topPorts(m map[int]int, n int) []int {
	type pc struct{ p, c int }
	var l []pc
	for p, c := range m {
		l = append(l, pc{p, c})
	}
	sort.Slice(l, func(i, j int) bool {
		if l[i].c != l[j].c {
			return l[i].c > l[j].c
		}
		return l[i].p < l[j].p
	})
	var out []int
	for i := 0; i < len(l) && i < n; i++ {
		out = append(out, l[i].p)
	}
	return out
}
