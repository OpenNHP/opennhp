package main

import (
	"bufio"
	"fmt"
	"io"
	"net/netip"
	"os"
	"sort"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

// Geo is the location of one address.
type Geo struct {
	Country string
	City    string
}

// GeoResolver maps an address to a location. Implementations must be offline.
type GeoResolver interface {
	Lookup(netip.Addr) (Geo, bool)
}

// csvGeo is a longest-prefix-match table loaded from "cidr,country[,city]".
type csvGeo struct {
	lens  []int // distinct prefix lengths, descending
	table map[int]map[netip.Prefix]Geo
}

func loadGeoCSV(path string) (GeoResolver, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	g := &csvGeo{table: map[int]map[netip.Prefix]Geo{}}
	sc := bufio.NewScanner(fh)
	n := 0
	for sc.Scan() {
		n++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			return nil, fmt.Errorf("%s:%d: want cidr,country[,city]", path, n)
		}
		p, err := netip.ParsePrefix(strings.TrimSpace(parts[0]))
		if err != nil {
			if n == 1 { // tolerate a header row
				continue
			}
			return nil, fmt.Errorf("%s:%d: %w", path, n, err)
		}
		p = p.Masked()
		geo := Geo{Country: strings.TrimSpace(parts[1])}
		if len(parts) > 2 {
			geo.City = strings.TrimSpace(parts[2])
		}
		if g.table[p.Bits()] == nil {
			g.table[p.Bits()] = map[netip.Prefix]Geo{}
			g.lens = append(g.lens, p.Bits())
		}
		g.table[p.Bits()][p] = geo
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.IntSlice(g.lens)))
	return g, nil
}

func (g *csvGeo) Lookup(a netip.Addr) (Geo, bool) {
	for _, bits := range g.lens {
		if bits > a.BitLen() {
			continue
		}
		p, err := a.Prefix(bits)
		if err != nil {
			continue
		}
		if geo, ok := g.table[bits][p]; ok {
			return geo, true
		}
	}
	return Geo{}, false
}

// mmdbGeo resolves locations from a MaxMind GeoLite2-City database. Lookups
// are local: the database file is read from disk, never queried online.
type mmdbGeo struct{ r *geoip2.Reader }

// loadGeoMMDB opens a GeoLite2-City .mmdb file. The returned Closer must be
// closed when lookups are done.
func loadGeoMMDB(path string) (GeoResolver, io.Closer, error) {
	r, err := geoip2.Open(path)
	if err != nil {
		return nil, nil, err
	}
	return &mmdbGeo{r: r}, r, nil
}

func (g *mmdbGeo) Lookup(a netip.Addr) (Geo, bool) {
	rec, err := g.r.City(a.AsSlice())
	if err != nil {
		return Geo{}, false
	}
	country := rec.Country.IsoCode
	if country == "" {
		country = rec.RegisteredCountry.IsoCode
	}
	city := rec.City.Names["en"]
	if country == "" && city == "" {
		return Geo{}, false
	}
	return Geo{Country: country, City: city}, true
}

// isReserved reports addresses that never get a geo lookup.
func isReserved(a netip.Addr) bool {
	return a.IsPrivate() || a.IsLoopback() || a.IsLinkLocalUnicast() ||
		a.IsLinkLocalMulticast() || a.IsMulticast() || a.IsUnspecified()
}
