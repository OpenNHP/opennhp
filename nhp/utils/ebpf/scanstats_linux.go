//go:build linux

package ebpf

// Port / source aggregation of dropped packets, the user-space half of
// nhp_port_hits and nhp_src_hits in nhp/ebpf/xdp/nhp_server_xdp.c.
//
// nhp_action_stats answers "how many packets per verdict class"; this answers
// "which ports, from which sources". Once the security group is wide open that
// is the question, and a single TCP_SYN_SCAN number cannot say whether it was
// port 23 or 3389, one source or four thousand.
//
// Every window prints at most 2*TopN+1 lines whatever the attack intensity, the
// same design rule as [NHP-STAT], so these lines need no byte budget of their
// own. The tables are LRU: a sweep wider than the table evicts the coldest
// keys, which makes Top-N approximate and is announced as LRU=1 on the TOTAL
// line. Exact per-class totals always come from [NHP-STAT], never from here.

import (
	"encoding/binary"
	"sort"
	"time"

	"github.com/cilium/ebpf"

	"github.com/OpenNHP/opennhp/nhp/log"
)

const (
	// Mirrors max_entries of nhp_port_hits / nhp_src_hits.
	scanMapMaxEntries = 65536
	// A table this close to full has very likely been evicting.
	scanMapLruThreshold = scanMapMaxEntries - scanMapMaxEntries/20
)

// scanPortKey mirrors struct nhp_scan_port_key. Port is host byte order.
type scanPortKey struct {
	Proto uint8
	Pad   uint8
	Port  uint16
}

type scanEntry[K comparable] struct {
	Key  K
	Hits uint64
}

// topN returns the n largest entries, largest first. Ties are broken by the
// order produced by less so that output is deterministic; callers with
// non-ordered keys get map iteration order only among exact ties, which is
// resolved by the stable key comparison below.
func topN[K comparable](m map[K]uint64, n int, less func(a, b K) bool) []scanEntry[K] {
	out := make([]scanEntry[K], 0, len(m))
	for k, v := range m {
		out = append(out, scanEntry[K]{Key: k, Hits: v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Hits != out[j].Hits {
			return out[i].Hits > out[j].Hits
		}
		if less != nil {
			return less(out[i].Key, out[j].Key)
		}
		return false
	})
	if n < 0 {
		n = 0
	}
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func sumPerCPU(v []uint64) uint64 {
	var t uint64
	for _, c := range v {
		t += c
	}
	return t
}

// drainMap empties m and hands every (key, per-CPU sum) to add, returning how
// many entries it held. A counter incremented between
// the read and the delete is lost; the figures here are a ranking, not the
// accounting.
func drainMap[K any](m *ebpf.Map, add func(k K, hits uint64)) (count int) {
	var (
		key  K
		vals []uint64
		keys []K
	)
	it := m.Iterate()
	for it.Next(&key, &vals) {
		add(key, sumPerCPU(vals))
		keys = append(keys, key)
		count++
	}
	// Delete after iterating: deleting under the iterator restarts it.
	for i := range keys {
		_ = m.Delete(&keys[i])
	}
	return count
}

func drainPortHits(m *ebpf.Map) (map[scanPortKey]uint64, bool) {
	out := make(map[scanPortKey]uint64)
	n := drainMap(m, func(k scanPortKey, hits uint64) { out[k] += hits })
	return out, n >= scanMapLruThreshold
}

// drainSrcHits keys by the address in host order, ready for uint32ToIPv4. The
// kernel key is the raw network-order __be32, so it is read as four bytes.
func drainSrcHits(m *ebpf.Map) (map[uint32]uint64, bool) {
	out := make(map[uint32]uint64)
	n := drainMap(m, func(k [4]byte, hits uint64) { out[binary.BigEndian.Uint32(k[:])] += hits })
	return out, n >= scanMapLruThreshold
}

func reportServerScan(portMap, srcMap *ebpf.Map, interval time.Duration, topn int, stop <-chan struct{}, serverId string, logger *log.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	window := int(interval.Seconds())

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
		}

		ports, portLru := drainPortHits(portMap)
		srcs, srcLru := drainSrcHits(srcMap)
		now := time.Now().Format("15:04:05")

		var pkts uint64
		for _, v := range ports {
			pkts += v
		}

		for _, e := range topN(ports, topn, func(a, b scanPortKey) bool {
			if a.Proto != b.Proto {
				return a.Proto < b.Proto
			}
			return a.Port < b.Port
		}) {
			logger.Info("%s %s [NHP-SCAN] WINDOW=%ds KIND=PORT PROTO=%s DPT=%d PKTS=%d",
				now, serverId, window, protoToString(e.Key.Proto), e.Key.Port, e.Hits)
		}
		for _, e := range topN(srcs, topn, func(a, b uint32) bool { return a < b }) {
			logger.Info("%s %s [NHP-SCAN] WINDOW=%ds KIND=SRC SRC=%s PKTS=%d",
				now, serverId, window, uint32ToIPv4(e.Key), e.Hits)
		}

		lru := 0
		if portLru || srcLru {
			lru = 1
		}
		logger.Info("%s %s [NHP-SCAN] WINDOW=%ds KIND=TOTAL UNIQ_SRC=%d UNIQ_PORT=%d PKTS=%d LRU=%d",
			now, serverId, window, len(srcs), len(ports), pkts, lru)
	}
}
