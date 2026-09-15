package main

import "encoding/base64"

// base64StdDecode is a tiny wrapper so config.go can iterate encoders without
// pulling encoding/base64 into the public surface twice.
func base64StdDecode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// base64URLExtraDecode accepts URL-safe base64 (no padding). Some operators
// paste `-`/`_` characters from base64url tools.
func base64URLExtraDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
