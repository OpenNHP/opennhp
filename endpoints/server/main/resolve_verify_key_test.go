package main

import (
	"encoding/base64"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

func verifyKeyContext(t *testing.T, key, keyFile string) *cli.Context {
	t.Helper()
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.String("key", "", "")
	set.String("key-file", "", "")
	ctx := cli.NewContext(nil, set, nil)
	if key != "" {
		if err := ctx.Set("key", key); err != nil {
			t.Fatalf("set --key: %v", err)
		}
	}
	if keyFile != "" {
		if err := ctx.Set("key-file", keyFile); err != nil {
			t.Fatalf("set --key-file: %v", err)
		}
	}
	return ctx
}

// TestResolveVerifyKeyEmptyKeyFileFallsThroughToEnv: a --key-file that
// exists but is empty (or whitespace-only) must not resolve to "no key" —
// it should fall through to NHP_AUDIT_KEY exactly as an unset --key-file
// already does.
func TestResolveVerifyKeyEmptyKeyFileFallsThroughToEnv(t *testing.T) {
	dir := t.TempDir()
	kf := filepath.Join(dir, "key")
	if err := os.WriteFile(kf, []byte("   \n"), 0600); err != nil {
		t.Fatal(err)
	}

	want := make([]byte, 32)
	for i := range want {
		want[i] = byte(i)
	}
	t.Setenv("NHP_AUDIT_KEY", base64.StdEncoding.EncodeToString(want))

	got, err := resolveVerifyKey(verifyKeyContext(t, "", kf))
	if err != nil {
		t.Fatalf("resolveVerifyKey: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("expected the empty --key-file to fall through to NHP_AUDIT_KEY, got %v", got)
	}
}

// TestResolveVerifyKeyRejectsShortKey pins the floor shared with
// initAuditLedger's SigningKeyBase64 check: a key too short to have come
// from the server must fail loudly here, not decode successfully into a
// key that will make every entry report as tampered.
func TestResolveVerifyKeyRejectsShortKey(t *testing.T) {
	short := base64.StdEncoding.EncodeToString([]byte("too-short"))
	_, err := resolveVerifyKey(verifyKeyContext(t, short, ""))
	if err == nil {
		t.Fatal("expected an error for a key under the minimum length")
	}
	if !strings.Contains(err.Error(), "need at least") {
		t.Fatalf("error should explain the floor, got: %v", err)
	}
}

// TestResolveVerifyKeyAcceptsMinimumLength: exactly MinSigningKeyLen bytes
// must be accepted, matching initAuditLedger's own boundary.
func TestResolveVerifyKeyAcceptsMinimumLength(t *testing.T) {
	key := make([]byte, 32)
	got, err := resolveVerifyKey(verifyKeyContext(t, base64.StdEncoding.EncodeToString(key), ""))
	if err != nil {
		t.Fatalf("a 32-byte key should be accepted: %v", err)
	}
	if len(got) != 32 {
		t.Fatalf("got %d bytes, want 32", len(got))
	}
}

// TestResolveVerifyKeyNoSourceIsNilNotError: absent from every source is
// the documented "verify without signature checking" mode, not an error.
func TestResolveVerifyKeyNoSourceIsNilNotError(t *testing.T) {
	// Otherwise this test depends on the ambient environment: it fails on
	// any machine/CI runner where NHP_AUDIT_KEY happens to be exported.
	t.Setenv("NHP_AUDIT_KEY", "")
	got, err := resolveVerifyKey(verifyKeyContext(t, "", ""))
	if err != nil {
		t.Fatalf("no key configured should not error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil key, got %v", got)
	}
}
