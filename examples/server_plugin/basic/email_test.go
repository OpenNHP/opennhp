package main

import "testing"

// TestHasEnvPlaceholder pins the whole-string-only behavior so free-form
// operator subjects containing a '$' are not mistaken for an un-expanded
// envsubst variable (the false-positive class from the code review).
func TestHasEnvPlaceholder(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		// Genuine un-expanded placeholders → true.
		{"${SMTP_SUBJECT}", true},
		{"$SMTP_SUBJECT", true},
		{"${SMTP_HOST}", true},
		{"$SMTP2", true},
		// Free-form text that merely contains '$' → false.
		{"Code $5.00", false},
		{"ACME $CORP verification", false},
		{"Verify your $USD account", false},
		{"OpenNHP ${SMTP_SUBJECT}", false}, // embedded, not whole-string
		{"Your OpenNHP Verification Code", false},
		{"", false},
		{"$", false},
		{"${}", false},
		{"$lowercase", false}, // envsubst vars are upper/underscore/digit
	}
	for _, c := range cases {
		if got := hasEnvPlaceholder(c.in); got != c.want {
			t.Errorf("hasEnvPlaceholder(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
