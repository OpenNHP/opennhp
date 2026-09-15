package metrics

import (
	"bufio"
	"sort"
	"strconv"
	"sync"
)

// DefaultBuckets are the Prometheus default latency buckets, in seconds.
var DefaultBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// HistogramVec is a family of fixed-bucket histograms partitioned by label
// names. Buckets are cumulative, matching the Prometheus histogram convention.
type HistogramVec struct {
	metricName string
	help       string
	labelNames []string
	buckets    []float64

	mu       sync.RWMutex
	children map[string]*Histogram
	order    []*Histogram
}

// Histogram accumulates observations into cumulative buckets plus a running sum
// and count.
type Histogram struct {
	labelValues []string
	buckets     []float64

	mu     sync.Mutex
	counts []uint64
	sum    float64
	total  uint64
}

// NewHistogram registers and returns a histogram family. If buckets is nil,
// DefaultBuckets is used. Buckets must be sorted ascending; they are sorted
// defensively here.
func (r *Registry) NewHistogram(name, help string, buckets []float64, labelNames ...string) *HistogramVec {
	if len(buckets) == 0 {
		buckets = DefaultBuckets
	}
	b := append([]float64(nil), buckets...)
	sort.Float64s(b)
	hv := &HistogramVec{
		metricName: name,
		help:       help,
		labelNames: labelNames,
		buckets:    b,
		children:   make(map[string]*Histogram),
	}
	r.register(hv)
	return hv
}

// With returns the histogram for the given label values, creating it on first
// use.
func (hv *HistogramVec) With(labelValues ...string) *Histogram {
	if len(labelValues) != len(hv.labelNames) {
		panic("metrics: wrong number of label values for " + hv.metricName)
	}
	key := joinLabelValues(labelValues)

	hv.mu.RLock()
	h := hv.children[key]
	hv.mu.RUnlock()
	if h != nil {
		return h
	}

	hv.mu.Lock()
	defer hv.mu.Unlock()
	if h = hv.children[key]; h != nil {
		return h
	}
	h = &Histogram{
		labelValues: append([]string(nil), labelValues...),
		buckets:     hv.buckets,
		counts:      make([]uint64, len(hv.buckets)),
	}
	hv.children[key] = h
	hv.order = append(hv.order, h)
	return h
}

// Observe records a single value (for latency, seconds).
func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	h.sum += v
	h.total++
	// Cumulative buckets: increment every bucket whose upper bound is >= v.
	for i, ub := range h.buckets {
		if v <= ub {
			h.counts[i]++
		}
	}
	h.mu.Unlock()
}

func (hv *HistogramVec) name() string { return hv.metricName }

func (hv *HistogramVec) writeTo(w *bufio.Writer) {
	writeHeader(w, hv.metricName, hv.help, "histogram")
	hv.mu.RLock()
	children := make([]*Histogram, len(hv.order))
	copy(children, hv.order)
	hv.mu.RUnlock()

	for _, h := range children {
		h.mu.Lock()
		counts := append([]uint64(nil), h.counts...)
		sum := h.sum
		total := h.total
		h.mu.Unlock()

		// counts[i] is already cumulative: Observe bumps every bucket whose
		// upper bound is >= the value.
		for i, ub := range hv.buckets {
			writeSampleExtra(w, hv.metricName, "_bucket", hv.labelNames, h.labelValues,
				"le", strconv.FormatFloat(ub, 'g', -1, 64), float64(counts[i]))
		}
		writeSampleExtra(w, hv.metricName, "_bucket", hv.labelNames, h.labelValues,
			"le", "+Inf", float64(total))
		writeSample(w, hv.metricName, "_sum", hv.labelNames, h.labelValues, sum)
		writeSample(w, hv.metricName, "_count", hv.labelNames, h.labelValues, float64(total))
	}
}
