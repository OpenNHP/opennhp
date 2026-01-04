package test

import (
	"testing"

	"github.com/OpenNHP/opennhp/nhp/utils"
)

// FuzzDecompression tests gzip decompression with malformed data.
// Compression/decompression of untrusted network data could cause panics.
func FuzzDecompression(f *testing.F) {
	// Seed with valid gzip data (base64 encoded compressed "hello")
	f.Add("H4sIAAAAAAAAA8tIzcnJBwQAAP//SqC0rQUAAAA=")
	// Invalid base64
	f.Add("not-valid-base64!!!")
	// Empty
	f.Add("")
	// Valid base64 but invalid gzip
	f.Add("SGVsbG8gV29ybGQ=")
	// Truncated gzip header
	f.Add("H4sIAAAA")

	f.Fuzz(func(t *testing.T, data string) {
		// Decompression should not panic on any input
		_, _ = utils.Decompression(data)
	})
}

// FuzzCompression tests gzip compression with various inputs.
func FuzzCompression(f *testing.F) {
	f.Add("hello world")
	f.Add("")
	f.Add(string(make([]byte, 1024))) // larger input
	f.Add("\x00\xFF\x00\xFF")

	f.Fuzz(func(t *testing.T, data string) {
		// Compression should not panic on any input
		compressed, err := utils.Compression(data)
		if err != nil {
			return
		}

		// If compression succeeded, verify decompression works
		_, _ = utils.Decompression(string(compressed))
	})
}
