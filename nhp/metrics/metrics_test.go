package metrics

import (
	"strings"
	"sync"
	"testing"
)

func render(t *testing.T, r *Registry) string {
	t.Helper()
	var b strings.Builder
	if err := r.WriteText(&b); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	return b.String()
}

func TestCounterExposition(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounter("nhp_test_total", "A test counter", "result")
	c.With("ok").Inc()
	c.With("ok").Inc()
	c.With("denied").Add(3)

	got := render(t, r)
	for _, want := range []string{
		"# HELP nhp_test_total A test counter\n",
		"# TYPE nhp_test_total counter\n",
		`nhp_test_total{result="ok"} 2` + "\n",
		`nhp_test_total{result="denied"} 3` + "\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\n---\n%s", want, got)
		}
	}
}

func TestCounterAddIgnoresNonPositive(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounter("nhp_c_total", "help")
	c.With().Add(5)
	c.With().Add(-3) // ignored
	c.With().Add(0)  // ignored
	if got := render(t, r); !strings.Contains(got, "nhp_c_total 5\n") {
		t.Errorf("counter should be 5, got:\n%s", got)
	}
}

func TestGaugeAndGaugeFunc(t *testing.T) {
	r := NewRegistry()
	g := r.NewGauge("nhp_g", "a gauge")
	g.Set(7)
	g.Inc() // 8
	g.Dec() // 7
	g.Dec() // 6

	n := 42
	r.NewGaugeFunc("nhp_gf", "a gauge func", func() float64 { return float64(n) })

	got := render(t, r)
	if !strings.Contains(got, "nhp_g 6\n") {
		t.Errorf("gauge wrong:\n%s", got)
	}
	if !strings.Contains(got, "# TYPE nhp_gf gauge\n") || !strings.Contains(got, "nhp_gf 42\n") {
		t.Errorf("gauge func wrong:\n%s", got)
	}

	// GaugeFunc is read at scrape time.
	n = 100
	if got := render(t, r); !strings.Contains(got, "nhp_gf 100\n") {
		t.Errorf("gauge func not re-read at scrape:\n%s", got)
	}
}

func TestHistogramExposition(t *testing.T) {
	r := NewRegistry()
	h := r.NewHistogram("nhp_dur_seconds", "durations", []float64{0.1, 0.5, 1}, "op")
	h.With("acop").Observe(0.05) // <=0.1, <=0.5, <=1
	h.With("acop").Observe(0.3)  // <=0.5, <=1
	h.With("acop").Observe(2)    // only +Inf

	got := render(t, r)
	for _, want := range []string{
		"# TYPE nhp_dur_seconds histogram\n",
		`nhp_dur_seconds_bucket{op="acop",le="0.1"} 1` + "\n",
		`nhp_dur_seconds_bucket{op="acop",le="0.5"} 2` + "\n",
		`nhp_dur_seconds_bucket{op="acop",le="1"} 2` + "\n",
		`nhp_dur_seconds_bucket{op="acop",le="+Inf"} 3` + "\n",
		`nhp_dur_seconds_sum{op="acop"} 2.35` + "\n",
		`nhp_dur_seconds_count{op="acop"} 3` + "\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q\n---\n%s", want, got)
		}
	}
}

func TestLabelValueEscaping(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounter("nhp_esc_total", "help \\ and \n newline", "path")
	c.With(`a"b\c` + "\n").Inc()

	got := render(t, r)
	if !strings.Contains(got, `# HELP nhp_esc_total help \\ and \n newline`+"\n") {
		t.Errorf("help not escaped:\n%s", got)
	}
	if !strings.Contains(got, `nhp_esc_total{path="a\"b\\c\n"} 1`+"\n") {
		t.Errorf("label value not escaped:\n%s", got)
	}
}

func TestDuplicateRegistrationPanics(t *testing.T) {
	r := NewRegistry()
	r.NewCounter("dup_total", "x")
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	r.NewGauge("dup_total", "y")
}

func TestInvalidNamePanics(t *testing.T) {
	r := NewRegistry()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on invalid metric name")
		}
	}()
	r.NewCounter("1bad-name", "x")
}

func TestWrongLabelCountPanics(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounter("nhp_lbl_total", "x", "a", "b")
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on wrong label count")
		}
	}()
	c.With("only-one")
}

func TestConcurrentUse(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounter("nhp_race_total", "x", "k")
	h := r.NewHistogram("nhp_race_seconds", "y", nil, "k")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				c.With("a").Inc()
				h.With("a").Observe(0.01)
			}
		}()
	}
	wg.Wait()

	if got := render(t, r); !strings.Contains(got, "nhp_race_total{k=\"a\"} 10000\n") {
		t.Errorf("expected 10000 increments:\n%s", got)
	}
}
