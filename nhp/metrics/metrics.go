// Package metrics is a small, dependency-free metrics registry that renders
// the Prometheus text exposition format (version 0.0.4).
//
// It intentionally covers only what the OpenNHP daemons need — counters,
// gauges (including scrape-time GaugeFunc), and fixed-bucket histograms — so
// the toolkit can expose a /metrics endpoint without pulling in the full
// Prometheus client library and its transitive dependencies.
//
// All label values are supplied by the daemons themselves and are expected to
// be low-cardinality (a fixed set of schemes, results, message types, ...).
// Never label a metric with attacker-controlled input.
package metrics

import (
	"bufio"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
)

// Registry holds a set of collectors and renders them together.
type Registry struct {
	mu         sync.Mutex
	collectors []collector
	names      map[string]struct{}
}

type collector interface {
	name() string
	writeTo(w *bufio.Writer)
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{names: make(map[string]struct{})}
}

func (r *Registry) register(c collector) {
	if !validMetricName(c.name()) {
		panic("metrics: invalid metric name " + c.name())
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.names[c.name()]; dup {
		panic("metrics: duplicate registration of " + c.name())
	}
	r.names[c.name()] = struct{}{}
	r.collectors = append(r.collectors, c)
}

// WriteText renders every registered collector in registration order using the
// Prometheus text exposition format.
func (r *Registry) WriteText(w io.Writer) error {
	r.mu.Lock()
	snapshot := make([]collector, len(r.collectors))
	copy(snapshot, r.collectors)
	r.mu.Unlock()

	bw := bufio.NewWriter(w)
	for _, c := range snapshot {
		c.writeTo(bw)
	}
	return bw.Flush()
}

// ---- Counter ----

// CounterVec is a family of monotonically increasing counters partitioned by a
// fixed set of label names.
type CounterVec struct {
	metricName string
	help       string
	labelNames []string

	mu       sync.RWMutex
	children map[string]*Counter
	order    []*Counter
}

// Counter is a single monotonically increasing value.
type Counter struct {
	labelValues []string
	v           atomic.Int64
}

// NewCounter registers and returns a counter family.
func (r *Registry) NewCounter(name, help string, labelNames ...string) *CounterVec {
	cv := &CounterVec{
		metricName: name,
		help:       help,
		labelNames: labelNames,
		children:   make(map[string]*Counter),
	}
	r.register(cv)
	return cv
}

// With returns the counter for the given label values, creating it on first
// use. The number of values must match the family's label names.
func (cv *CounterVec) With(labelValues ...string) *Counter {
	if len(labelValues) != len(cv.labelNames) {
		panic("metrics: wrong number of label values for " + cv.metricName)
	}
	key := joinLabelValues(labelValues)

	cv.mu.RLock()
	c := cv.children[key]
	cv.mu.RUnlock()
	if c != nil {
		return c
	}

	cv.mu.Lock()
	defer cv.mu.Unlock()
	if c = cv.children[key]; c != nil {
		return c
	}
	c = &Counter{labelValues: append([]string(nil), labelValues...)}
	cv.children[key] = c
	cv.order = append(cv.order, c)
	return c
}

// Inc adds one.
func (c *Counter) Inc() { c.v.Add(1) }

// Add increases the counter by n. Negative values are ignored — counters only
// go up.
func (c *Counter) Add(n int64) {
	if n <= 0 {
		return
	}
	c.v.Add(n)
}

func (cv *CounterVec) name() string { return cv.metricName }

func (cv *CounterVec) writeTo(w *bufio.Writer) {
	writeHeader(w, cv.metricName, cv.help, "counter")
	cv.mu.RLock()
	children := make([]*Counter, len(cv.order))
	copy(children, cv.order)
	cv.mu.RUnlock()
	for _, c := range children {
		writeSample(w, cv.metricName, "", cv.labelNames, c.labelValues, float64(c.v.Load()))
	}
}

// ---- Gauge ----

// Gauge is a single value that can go up or down.
type Gauge struct {
	metricName string
	help       string
	v          atomic.Int64
}

// NewGauge registers and returns a gauge.
func (r *Registry) NewGauge(name, help string) *Gauge {
	g := &Gauge{metricName: name, help: help}
	r.register(g)
	return g
}

// Set replaces the gauge's value.
func (g *Gauge) Set(v int64) { g.v.Store(v) }

// Inc adds one.
func (g *Gauge) Inc() { g.v.Add(1) }

// Dec subtracts one.
func (g *Gauge) Dec() { g.v.Add(-1) }

func (g *Gauge) name() string { return g.metricName }

func (g *Gauge) writeTo(w *bufio.Writer) {
	writeHeader(w, g.metricName, g.help, "gauge")
	writeSample(w, g.metricName, "", nil, nil, float64(g.v.Load()))
}

// GaugeFunc is a gauge whose value is read from a function at scrape time. It
// is ideal for quantities the daemon already tracks (open connections, an
// overload flag) so no separate counter has to be kept in sync.
type GaugeFunc struct {
	metricName string
	help       string
	fn         func() float64
}

// NewGaugeFunc registers a gauge backed by fn, which is evaluated on every
// scrape. fn must be safe to call concurrently.
func (r *Registry) NewGaugeFunc(name, help string, fn func() float64) *GaugeFunc {
	gf := &GaugeFunc{metricName: name, help: help, fn: fn}
	r.register(gf)
	return gf
}

func (gf *GaugeFunc) name() string { return gf.metricName }

func (gf *GaugeFunc) writeTo(w *bufio.Writer) {
	writeHeader(w, gf.metricName, gf.help, "gauge")
	writeSample(w, gf.metricName, "", nil, nil, gf.fn())
}

// ---- shared rendering helpers ----

func writeHeader(w *bufio.Writer, name, help, typ string) {
	w.WriteString("# HELP ")
	w.WriteString(name)
	w.WriteByte(' ')
	w.WriteString(escapeHelp(help))
	w.WriteByte('\n')
	w.WriteString("# TYPE ")
	w.WriteString(name)
	w.WriteByte(' ')
	w.WriteString(typ)
	w.WriteByte('\n')
}

// writeSample writes one "name{labels} value" line. suffix is appended to the
// metric name (used by histograms for _bucket/_sum/_count); extraName/extraVal
// carry an additional synthetic label (the histogram "le") when non-empty.
func writeSample(w *bufio.Writer, name, suffix string, labelNames, labelValues []string, value float64) {
	writeSampleExtra(w, name, suffix, labelNames, labelValues, "", "", value)
}

func writeSampleExtra(w *bufio.Writer, name, suffix string, labelNames, labelValues []string, extraName, extraVal string, value float64) {
	w.WriteString(name)
	w.WriteString(suffix)
	if len(labelNames) > 0 || extraName != "" {
		w.WriteByte('{')
		first := true
		for i, ln := range labelNames {
			if !first {
				w.WriteByte(',')
			}
			first = false
			w.WriteString(ln)
			w.WriteString(`="`)
			w.WriteString(escapeLabelValue(labelValues[i]))
			w.WriteByte('"')
		}
		if extraName != "" {
			if !first {
				w.WriteByte(',')
			}
			w.WriteString(extraName)
			w.WriteString(`="`)
			w.WriteString(escapeLabelValue(extraVal))
			w.WriteByte('"')
		}
		w.WriteByte('}')
	}
	w.WriteByte(' ')
	w.WriteString(formatFloat(value))
	w.WriteByte('\n')
}

func formatFloat(v float64) string {
	// Integers render without a decimal point; everything else keeps full
	// precision. strconv with -1 gives the shortest round-trippable form.
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func joinLabelValues(values []string) string {
	// The zero byte cannot appear in the label values we produce, so it is a
	// safe separator that keeps ("a","bc") distinct from ("ab","c").
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	}
	total := len(values) - 1
	for _, v := range values {
		total += len(v)
	}
	b := make([]byte, 0, total)
	for i, v := range values {
		if i > 0 {
			b = append(b, 0)
		}
		b = append(b, v...)
	}
	return string(b)
}

func escapeHelp(s string) string {
	if !needsEscape(s, false) {
		return s
	}
	return escape(s, false)
}

func escapeLabelValue(s string) string {
	if !needsEscape(s, true) {
		return s
	}
	return escape(s, true)
}

func needsEscape(s string, quote bool) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\', '\n':
			return true
		case '"':
			if quote {
				return true
			}
		}
	}
	return false
}

func escape(s string, quote bool) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\\':
			b = append(b, '\\', '\\')
		case '\n':
			b = append(b, '\\', 'n')
		case '"':
			if quote {
				b = append(b, '\\', '"')
			} else {
				b = append(b, c)
			}
		default:
			b = append(b, c)
		}
	}
	return string(b)
}

func validMetricName(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '_', c == ':':
			// always allowed
		case c >= '0' && c <= '9':
			if i == 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}
