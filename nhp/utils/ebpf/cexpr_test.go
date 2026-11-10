package ebpf

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// cDefines extracts the #define lines of an eBPF C file into a map of name
// to value, so a test can evaluate one instead of matching its text.
func cDefines(t *testing.T, path string) map[string]string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	defs := make(map[string]string)
	re := regexp.MustCompile(`(?m)^#define\s+([A-Za-z_][A-Za-z0-9_]*)\s+(.+?)\s*$`)
	for _, m := range re.FindAllStringSubmatch(string(b), -1) {
		defs[m[1]] = m[2]
	}
	return defs
}

// cExprParser evaluates the small integer expressions the eBPF headers use in
// their #define lines: literals, other #defines, parentheses and + - *.
//
// The rate map's slot numbers are defined in terms of each other
// (NHP_LEGACY_GLOBAL_SLOT in terms of NHP_LEGACY_RATE_BASE and
// NHP_EVENT_ACTIONS), so reading the numbers out of the source means
// evaluating them. A substring match would pin the formatting instead of the
// arithmetic, and the arithmetic is what matters: spend_token() treats a slot
// the map cannot hand back as "has budget", so a slot number that drifted
// past the map's size would remove a limit without a single error.
type cExprParser struct {
	t     *testing.T
	toks  []string
	pos   int
	defs  map[string]string
	depth int
}

// evalCIntExpr evaluates expr over defs.
func evalCIntExpr(t *testing.T, defs map[string]string, expr string) int {
	t.Helper()
	p := &cExprParser{t: t, toks: tokenizeCExpr(expr), defs: defs}
	v := p.parseExpr(t)
	if p.pos != len(p.toks) {
		t.Fatalf("trailing tokens in C expression %q", expr)
	}
	return v
}

func tokenizeCExpr(expr string) []string {
	var toks []string
	i := 0
	for i < len(expr) {
		c := expr[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c >= '0' && c <= '9':
			j := i
			for j < len(expr) && expr[j] >= '0' && expr[j] <= '9' {
				j++
			}
			toks = append(toks, expr[i:j])
			i = j
		case c == '_' || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z'):
			j := i
			for j < len(expr) && (expr[j] == '_' ||
				(expr[j] >= 'A' && expr[j] <= 'Z') ||
				(expr[j] >= 'a' && expr[j] <= 'z') ||
				(expr[j] >= '0' && expr[j] <= '9')) {
				j++
			}
			toks = append(toks, expr[i:j])
			i = j
		default:
			toks = append(toks, string(c))
			i++
		}
	}
	return toks
}

func (p *cExprParser) peek() string {
	if p.pos < len(p.toks) {
		return p.toks[p.pos]
	}
	return ""
}

func (p *cExprParser) parseExpr(t *testing.T) int {
	t.Helper()
	v := p.parseTerm(t)
	for {
		switch p.peek() {
		case "+":
			p.pos++
			v += p.parseTerm(t)
		case "-":
			p.pos++
			v -= p.parseTerm(t)
		default:
			return v
		}
	}
}

func (p *cExprParser) parseTerm(t *testing.T) int {
	t.Helper()
	v := p.parseFactor(t)
	for p.peek() == "*" {
		p.pos++
		v *= p.parseFactor(t)
	}
	return v
}

func (p *cExprParser) parseFactor(t *testing.T) int {
	t.Helper()
	tok := p.peek()
	if tok == "" {
		t.Fatal("C expression ended early")
	}
	if tok == "(" {
		p.pos++
		v := p.parseExpr(t)
		if p.peek() != ")" {
			t.Fatal("unbalanced parentheses in C expression")
		}
		p.pos++
		return v
	}
	if tok[0] >= '0' && tok[0] <= '9' {
		p.pos++
		n, err := strconv.Atoi(tok)
		if err != nil {
			t.Fatalf("bad number %q in C expression: %v", tok, err)
		}
		return n
	}

	// An identifier: another #define.
	p.pos++
	if p.depth > 16 {
		t.Fatalf("C expression recursion is too deep at %q", tok)
	}
	val, ok := p.defs[tok]
	if !ok {
		t.Fatalf("C expression refers to the unknown name %q", tok)
	}
	p.depth++
	sub := &cExprParser{t: p.t, toks: tokenizeCExpr(val), defs: p.defs, depth: p.depth}
	v := sub.parseExpr(p.t)
	p.depth--
	return v
}

// TestLegacyAuditDropFloodCannotSilenceAdmits pins the property the accept
// channel's own global cap exists for: a flood of dropped packets must not be
// able to spend the allowance of the log that records what was admitted.
//
// The model below is spend_token() as the C spells it -- a per-slot counter
// that resets every window -- with the slot numbers, the burst sizes and the
// action codes read out of the C sources, so the test cannot drift from the
// program it is about. The last part runs the same flood through the shape
// the code used to have, one shared global cap, and pins that the admit is
// lost there: that is the bug this fix removes, and a test that cannot fail
// on it proves nothing.
func TestLegacyAuditDropFloodCannotSilenceAdmits(t *testing.T) {
	statsPath := filepath.Join("..", "..", "ebpf", "xdp", "nhp_xdp_stats.h")
	progPath := filepath.Join("..", "..", "ebpf", "xdp", "nhp_ebpf_xdp.c")

	src, err := os.ReadFile(statsPath)
	if err != nil {
		t.Fatalf("read %s: %v", statsPath, err)
	}
	defs := cDefines(t, statsPath)
	actions := evalCIntExpr(t, defs, defs["NHP_EVENT_ACTIONS"])
	legacyBase := evalCIntExpr(t, defs, defs["NHP_LEGACY_RATE_BASE"])
	legacyGlobal := evalCIntExpr(t, defs, defs["NHP_LEGACY_GLOBAL_SLOT"])
	legacyAccept := evalCIntExpr(t, defs, defs["NHP_LEGACY_ACCEPT_GLOBAL_SLOT"])
	classBurst := evalCIntExpr(t, defs, defs["NHP_EVENT_BURST"])
	globalBurst := evalCIntExpr(t, defs, defs["NHP_EVENT_GLOBAL_BURST"])
	window := uint64(evalCIntExpr(t, defs, trimCSuffix(defs["NHP_EVENT_WINDOW_NS"])))

	progDefs := cDefines(t, progPath)
	admitClass := evalCIntExpr(t, progDefs, progDefs["ACT_AC_PASS_KNOCK_ADMITTED"])
	dropClassA := evalCIntExpr(t, progDefs, progDefs["ACT_AC_DROP_TCP_OTHER"])
	dropClassB := evalCIntExpr(t, progDefs, progDefs["ACT_AC_DROP_UDP_OTHER"])

	// The slot the accept verdict spends is read out of the C, not assumed
	// here: if the program went back to spending the deny channel's slot for
	// accepts, this test has to fail rather than keep asserting a shape the
	// program does not have.
	acceptGlobal := acceptGlobalSlot(t, string(src), defs)
	if acceptGlobal != legacyAccept {
		t.Errorf("the accept verdict spends slot %d, want the accept channel's own slot %d",
			acceptGlobal, legacyAccept)
	}

	// The model of the bucket: one counter per slot, reset every window.
	type bucket struct {
		windowStart uint64
		count       int
	}
	type slots map[int]*bucket

	spend := func(s slots, slot, burst int, now uint64) bool {
		b := s[slot]
		if b == nil {
			b = &bucket{}
			s[slot] = b
		}
		if now-b.windowStart >= window {
			b.windowStart = now
			b.count = 1
			return true
		}
		if b.count >= burst {
			return false
		}
		b.count++
		return true
	}

	// available models legacy_event_budget_available() as the C spells it.
	available := func(s slots, action, verdict int, now uint64) bool {
		if action >= actions {
			return true
		}
		if !spend(s, legacyBase+action, classBurst, now) {
			return false
		}
		global := legacyGlobal
		if verdict == 1 {
			global = acceptGlobal
		}
		return spend(s, global, globalBurst, now)
	}

	// One window, well past the first, so the counters are live.
	now := 5 * window

	// A scan spreads its packets over the classes, so two of them carry the
	// flood and neither class bucket is what runs out.
	flood := func(s slots) {
		for i := 0; i < 200; i++ {
			class := dropClassA
			if i%2 == 1 {
				class = dropClassB
			}
			available(s, class, 0, now)
		}
	}

	// The shape the program has now: the accept channel has a cap of its
	// own, so the admit survives the flood.
	live := slots{}
	flood(live)
	if available(live, dropClassA, 0, now) {
		t.Error("the deny channel is not capped at all; the flood should have spent its global slot")
	}
	if !available(live, admitClass, 1, now) {
		t.Error("a drop flood silenced the accept channel; the record of what was admitted must survive it")
	}

	// The shape the program had before: one shared global cap, and the same
	// flood loses the admit. This is the bug the fix removes, and it is what
	// the assertion above would not catch if the model had stopped working.
	shared := slots{}
	for i := 0; i < 200; i++ {
		class := dropClassA
		if i%2 == 1 {
			class = dropClassB
		}
		if !spend(shared, legacyBase+class, classBurst, now) {
			continue
		}
		spend(shared, legacyGlobal, globalBurst, now)
	}
	if !spend(shared, legacyBase+admitClass, classBurst, now) {
		t.Fatal("the admit's own class bucket was spent by the flood; the model is not what it claims")
	}
	if spend(shared, legacyGlobal, globalBurst, now) {
		t.Error("the model no longer reproduces the shared-cap behaviour; the test above is not pinning anything")
	}
}

// acceptGlobalSlot reads out of the C which slot the accept verdict spends in
// legacy_event_budget_available(). The two channels are capped separately, and
// the accept channel's cap is what a flood of drops must not be able to spend.
func acceptGlobalSlot(t *testing.T, src string, defs map[string]string) int {
	t.Helper()

	re := regexp.MustCompile(`verdict\s*==\s*1\s*\)\s*\?\s*([A-Za-z_][A-Za-z0-9_]*)\s*:\s*([A-Za-z_][A-Za-z0-9_]*)`)
	m := re.FindStringSubmatch(src)
	if m == nil {
		t.Fatal("the audit ring's global cap is not chosen by the verdict")
	}
	accept, deny := m[1], m[2]
	if accept == deny {
		t.Errorf("the accept channel spends the deny channel's global slot (%s); a drop flood would silence the accepts",
			accept)
	}
	for _, name := range []string{accept, deny} {
		if _, ok := defs[name]; !ok {
			t.Fatalf("the C refers to %s, which is not defined in the header", name)
		}
	}
	return evalCIntExpr(t, defs, defs[accept])
}

// trimCSuffix drops the integer suffix a C literal may carry ("1000000000ULL"
// is a number to C and not to the expression parser).
func trimCSuffix(v string) string {
	v = strings.TrimSpace(v)
	for len(v) > 0 {
		c := v[len(v)-1]
		if c != 'u' && c != 'U' && c != 'l' && c != 'L' {
			break
		}
		v = v[:len(v)-1]
	}
	return v
}
