package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"io/fs"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
)

// Per-day fact files: what makes the all-time source and country counts on
// the site exact instead of a sum of daily ones.
//
// The published record for a day is d/<source>/<day>.json, which carries
// unique_attack_ips as a count and only the top N addresses. A count cannot be
// added up: an address that scans the host every day is one attacker, and
// summing the days counts it once per day. The country total has the same
// shape — the day's countries are a set, and two sets only union once they are
// both present.
//
// Those sets cannot go in the published JSON. The site masks source IPs
// because it is public, and a 64-bit hash of an IPv4 address is reversible by
// brute force over the whole 2^32 space, so publishing hashes would hand back
// exactly what the masking hides. So the sets live in a private directory
// beside state.json — outside the web root on the relay deploy, and refused by
// the nginx vhost as a second line of defense, like /state.json.
//
// What is stored per address is the hash, never the address: the file is not a
// second copy of the unmasked raw logs, and dedup and union are all the site
// needs it for.

// factsExt and factsMagic version the on-disk layout. A file that does not
// start with the magic is refused rather than misread — a future layout must
// not be decoded as this one and produce a plausible wrong total.
const (
	factsExt   = ".facts"
	factsMagic = "XDF1"
)

// factsHeaderSize is magic (4) + source count (4) + country count (4).
const factsHeaderSize = 12

// dayFactsPath is one day's fact file. It hangs off the -state file rather
// than -out, so it follows the state into the private directory by default
// instead of needing a flag of its own to stay unpublished.
func dayFactsPath(statePath, source, day string) string {
	return filepath.Join(filepath.Dir(statePath), "facts", source, day+factsExt)
}

// sourceHash is FNV-1a over the address's 16-byte form, so an IPv4 and its
// IPv4-in-IPv6 spelling hash alike. It is a fixed-width key that is stable
// across runs and hosts, not a secret — see the note above.
func sourceHash(a netip.Addr) uint64 {
	b := a.As16()
	h := fnv.New64a()
	_, _ = h.Write(b[:])
	return h.Sum64()
}

// encodeDayFacts renders one day's sets. Sources go in as sorted uint64s so
// the bytes are a pure function of the day's attackers and a rebuild that
// finds the same ones rewrites an identical file; countries are sorted ASCII,
// which makes the file legible to `cat` and `sort` during an incident.
func encodeDayFacts(srcs []uint64, countries []string) []byte {
	b := make([]byte, factsHeaderSize, factsHeaderSize+8*len(srcs)+8*len(countries))
	copy(b, factsMagic)
	binary.LittleEndian.PutUint32(b[4:], uint32(len(srcs)))
	binary.LittleEndian.PutUint32(b[8:], uint32(len(countries)))
	for _, h := range srcs {
		b = binary.LittleEndian.AppendUint64(b, h)
	}
	for _, c := range countries {
		b = append(b, byte(len(c)))
		b = append(b, c...)
	}
	return b
}

// decodeDayFacts is the inverse of encodeDayFacts.
//
// A truncated tail is ignored rather than fatal: half a fact file must not
// take the site down, and the day it came from becomes a missing contributor
// to a total instead of a wrong one.
func decodeDayFacts(b []byte) ([]uint64, []string, error) {
	if len(b) < factsHeaderSize || string(b[:4]) != factsMagic {
		return nil, nil, fmt.Errorf("not a %s fact file", factsMagic)
	}
	nSrc := int(binary.LittleEndian.Uint32(b[4:]))
	nCountry := int(binary.LittleEndian.Uint32(b[8:]))
	rest := b[factsHeaderSize:]

	srcs := make([]uint64, 0, nSrc)
	for i := 0; i < nSrc && len(rest) >= 8; i++ {
		srcs = append(srcs, binary.LittleEndian.Uint64(rest[:8]))
		rest = rest[8:]
	}
	countries := make([]string, 0, nCountry)
	for i := 0; i < nCountry && len(rest) > 0; i++ {
		n := int(rest[0])
		rest = rest[1:]
		if len(rest) < n {
			break
		}
		countries = append(countries, string(rest[:n]))
		rest = rest[n:]
	}
	return srcs, countries, nil
}

// writeDayFacts writes one day's sets atomically, creating the private
// directory on the way. 0600 on a 0700 directory: the file is derived from the
// unmasked source addresses, so it gets the raw cache's treatment rather than
// the published site's.
func writeDayFacts(path string, srcs []uint64, countries []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return writeFileAtomic(path, encodeDayFacts(srcs, countries), 0o600)
}

// readDayFacts loads one day's sets. A missing file is not an error: it means
// the day was published before this feature existed, or its facts were
// removed, and the caller reports it as a day that contributes nothing.
func readDayFacts(path string) ([]uint64, []string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	return decodeDayFacts(b)
}

// removeDayFacts deletes one day's sets. A file that is already gone is not
// an error, exactly as for the detail page in unpublishDay.
func removeDayFacts(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

// factsUnion is the all-time picture the sets add up to, plus the days that
// could not contribute to it.
type factsUnion struct {
	// Sources is the number of distinct attack addresses seen on at least one
	// published day — one attacker seen on ten days is one attacker.
	Sources int
	// Countries is the number of distinct country codes those sources
	// resolved to. Sources the database does not resolve belong to no
	// country and are counted in Sources only.
	Countries int
	// Missing names the days whose facts could not be read, "<source>/<day>
	// (reason)". A non-empty list means the totals above are a lower bound,
	// which is why runSite prints it rather than swallowing it.
	Missing []string
}

// unionDayFacts merges every published day's sets into the all-time counts.
func unionDayFacts(statePath string, summaries []SiteDaySummary) factsUnion {
	var u factsUnion
	if statePath == "" {
		return u
	}
	seenCountry := map[string]struct{}{}
	var all []uint64
	for _, s := range summaries {
		key := s.Source + "/" + s.Day
		srcs, countries, err := readDayFacts(dayFactsPath(statePath, s.Source, s.Day))
		if err != nil {
			u.Missing = append(u.Missing, fmt.Sprintf("%s (%v)", key, err))
			continue
		}
		all = append(all, srcs...)
		for _, c := range countries {
			seenCountry[c] = struct{}{}
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i] < all[j] })
	for i, h := range all {
		if i == 0 || h != all[i-1] {
			u.Sources++
		}
	}
	u.Countries = len(seenCountry)
	return u
}
