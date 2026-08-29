package common

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateDoID(t *testing.T) {
	tests := []struct {
		name  string
		doID  string
		valid bool
	}{
		{name: "alphanumeric", doID: "object123", valid: true},
		{name: "hyphen and underscore", doID: "Object_123-abc", valid: true},
		{name: "UUID", doID: "550e8400-e29b-41d4-a716-446655440000", valid: true},
		{name: "maximum length", doID: strings.Repeat("a", 64), valid: true},
		{name: "empty", doID: "", valid: false},
		{name: "too long", doID: strings.Repeat("a", 65), valid: false},
		{name: "parent traversal", doID: "foo/../../../bar", valid: false},
		{name: "absolute path", doID: "/etc/passwd", valid: false},
		{name: "double dot", doID: "..", valid: false},
		{name: "forward slash", doID: "foo/bar", valid: false},
		{name: "backslash", doID: `foo\bar`, valid: false},
		{name: "null byte", doID: "foo\x00bar", valid: false},
		{name: "unicode separator", doID: "foo\u2044bar", valid: false},
		{name: "whitespace", doID: "foo bar", valid: false},
		{name: "shell metacharacter", doID: "foo;bar", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateDoID(test.doID)
			if test.valid && err != nil {
				t.Fatalf("ValidateDoID(%q) returned %v", test.doID, err)
			}
			if !test.valid && !errors.Is(err, ErrInvalidDoID) {
				t.Fatalf("ValidateDoID(%q) = %v, want ErrInvalidDoID", test.doID, err)
			}
		})
	}
}
