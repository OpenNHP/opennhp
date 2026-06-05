package test

import (
	"net/url"
	"testing"
)

func TestUrlEncoding(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"中文", "%E4%B8%AD%E6%96%87"},
		{"", ""},
		{"hello world", "hello+world"},
		{"a+b&c=d", "a%2Bb%26c%3Dd"},
	}

	for _, c := range cases {
		encoded := url.QueryEscape(c.input)
		if encoded != c.expected {
			t.Errorf("QueryEscape(%q) = %q, want %q", c.input, encoded, c.expected)
		}

		decoded, err := url.QueryUnescape(encoded)
		if err != nil {
			t.Errorf("QueryUnescape(%q) unexpected error: %v", encoded, err)
			continue
		}
		if decoded != c.input {
			t.Errorf("QueryUnescape(%q) = %q, want %q", encoded, decoded, c.input)
		}
	}
}
