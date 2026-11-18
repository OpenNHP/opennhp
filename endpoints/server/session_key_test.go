package server

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// The session store must not be keyed by a constant.
//
// cookie.NewStore's first argument is the HMAC key for the session cookie.
// It used to be []byte("nhpstore") -- one value, in the repository, shared
// by every deployment, so anyone who read the source could mint a cookie
// that any nhp-server accepted. The session is handed to auth plugins
// through SessionGet/Set/Save/Clear, so forging it forges whatever a
// plugin keeps there.
//
// Asserted on the source because the alternative is asserting that two
// processes disagree about a key, which needs two processes. This is the
// line that must not come back.
func TestSessionStoreKeysAreNotConstants(t *testing.T) {
	src, err := os.ReadFile("httpserver.go")
	if err != nil {
		t.Fatalf("read httpserver.go: %v", err)
	}
	// Comments quote the old call on purpose, so scan code only.
	code := stripLineComments(string(src))

	re := regexp.MustCompile(`cookie\.NewStore\(([^)]*)\)`)
	m := re.FindStringSubmatch(code)
	if m == nil {
		t.Fatal("cookie.NewStore call not found; re-anchor this test rather than deleting it")
	}
	args := m[1]

	// A literal anywhere in the argument list is a baked-in key.
	if regexp.MustCompile(`"|'|\[\]byte\(`).MatchString(args) {
		t.Errorf("cookie.NewStore(%s) is keyed by a literal; generate the keys at startup", args)
	}
	// Both an authentication and an encryption key: with only the first,
	// the cookie is signed but its contents sit in the browser in clear.
	parts := regexp.MustCompile(`\s*,\s*`).Split(strings.TrimSpace(args), -1)
	if len(parts) != 2 {
		t.Errorf("cookie.NewStore(%s) takes %d argument(s); want an authentication key and an encryption key",
			args, len(parts))
	}
}

// The keys are drawn from crypto/rand and a failure refuses to start,
// rather than falling back to something predictable.
func TestSessionKeyGenerationFailsClosed(t *testing.T) {
	src, err := os.ReadFile("httpserver.go")
	if err != nil {
		t.Fatalf("read httpserver.go: %v", err)
	}
	s := string(src)
	for _, want := range []string{"rand.Read(authKey)", "rand.Read(encKey)"} {
		if !regexp.MustCompile(regexp.QuoteMeta(want)).MatchString(s) {
			t.Errorf("missing %s: session keys must come from crypto/rand", want)
		}
	}
	// Each read is followed by a return, not a fallback.
	re := regexp.MustCompile(`(?s)rand\.Read\((?:auth|enc)Key\); err != nil \{.*?\n\t\}`)
	for _, blk := range re.FindAllString(s, -1) {
		if !regexp.MustCompile(`return err`).MatchString(blk) {
			t.Errorf("a failed key generation does not refuse to start:\n%s", blk)
		}
	}
}

// stripLineComments removes // comments so a test that scans for a code
// shape is not satisfied, or tripped, by prose quoting that shape.
func stripLineComments(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}
