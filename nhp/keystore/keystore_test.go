package keystore

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func randKey(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return b
}

func TestSealOpenRoundTrip(t *testing.T) {
	key := randKey(t, 32)
	pass := []byte("correct horse battery staple")

	blob, err := Seal(key, pass)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if !IsSealed(blob) {
		t.Fatalf("sealed blob %q not recognized as sealed", blob)
	}

	got, err := Open(blob, pass)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !bytes.Equal(got, key) {
		t.Fatalf("round-trip mismatch: got %x want %x", got, key)
	}
}

func TestSealIsRandomized(t *testing.T) {
	key := randKey(t, 32)
	pass := []byte("pw")
	a, err := Seal(key, pass)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Seal(key, pass)
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("two seals of the same key produced identical blobs (salt/nonce not random)")
	}
}

func TestOpenWrongPassphrase(t *testing.T) {
	key := randKey(t, 32)
	blob, err := Seal(key, []byte("right"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(blob, []byte("wrong")); err != ErrBadPassphrase {
		t.Fatalf("wrong passphrase: got %v want ErrBadPassphrase", err)
	}
}

func TestOpenTamperedCiphertext(t *testing.T) {
	key := randKey(t, 32)
	pass := []byte("pw")
	blob, err := Seal(key, pass)
	if err != nil {
		t.Fatal(err)
	}

	parts := strings.Split(blob, "$")
	ct, err := base64.RawStdEncoding.DecodeString(parts[7])
	if err != nil {
		t.Fatal(err)
	}
	ct[len(ct)-1] ^= 0x01 // flip a bit in the GCM tag/ciphertext
	parts[7] = base64.RawStdEncoding.EncodeToString(ct)
	tampered := strings.Join(parts, "$")

	if _, err := Open(tampered, pass); err != ErrBadPassphrase {
		t.Fatalf("tampered blob: got %v want ErrBadPassphrase", err)
	}
}

func TestOpenNoPassphrase(t *testing.T) {
	blob, err := Seal(randKey(t, 32), []byte("pw"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(blob, nil); err != ErrNoPassphrase {
		t.Fatalf("no passphrase: got %v want ErrNoPassphrase", err)
	}
}

func TestOpenMalformed(t *testing.T) {
	cases := []string{
		"v1$argon2id$3$65536$4$onlyfour",
		"v2$argon2id$3$65536$4$AAAA$BBBB$CCCC",
		"v1$scrypt$3$65536$4$AAAA$BBBB$CCCC",
		"v1$argon2id$x$65536$4$AAAA$BBBB$CCCC",
		"v1$argon2id$3$65536$4$!!!!$BBBB$CCCC",
	}
	for _, blob := range cases {
		if _, err := Open(blob, []byte("pw")); err == nil {
			t.Fatalf("expected error for malformed blob %q", blob)
		}
	}
}

// TestOpenRejectsOversizedKDFParams ensures a hostile blob cannot drive
// argon2 into a huge allocation: out-of-range time/memory are rejected as
// malformed before the KDF runs.
func TestOpenRejectsOversizedKDFParams(t *testing.T) {
	// Build a structurally valid blob but swap in an absurd memory cost.
	blob, err := Seal(randKey(t, 32), []byte("pw"))
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(blob, "$")
	parts[3] = "999999999" // memory KiB, far above maxArgonMemory
	if _, err := Open(strings.Join(parts, "$"), []byte("pw")); err != ErrMalformedBlob {
		t.Fatalf("oversized memory: got %v want ErrMalformedBlob", err)
	}

	parts = strings.Split(blob, "$")
	parts[2] = "9999" // time, far above maxArgonTime
	if _, err := Open(strings.Join(parts, "$"), []byte("pw")); err != ErrMalformedBlob {
		t.Fatalf("oversized time: got %v want ErrMalformedBlob", err)
	}
}

// TestOpenRejectsBadNonceLength ensures a wrong-length nonce is rejected as
// malformed (fast-fail before the KDF).
func TestOpenRejectsBadNonceLength(t *testing.T) {
	blob, err := Seal(randKey(t, 32), []byte("pw"))
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(blob, "$")
	parts[6] = base64.RawStdEncoding.EncodeToString([]byte("shortnonce")) // != 12 bytes
	if _, err := Open(strings.Join(parts, "$"), []byte("pw")); err != ErrMalformedBlob {
		t.Fatalf("bad nonce length: got %v want ErrMalformedBlob", err)
	}
}

// TestPassphraseFileEmptyIsDistinctError ensures pointing the file var at
// an empty file reports that, rather than the misleading "no passphrase
// was provided" the caller would otherwise surface.
func TestPassphraseFileEmptyIsDistinctError(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "empty")
	if writeErr := os.WriteFile(fp, []byte("\n"), 0600); writeErr != nil {
		t.Fatal(writeErr)
	}
	t.Setenv(EnvPassphraseFile, fp)

	got, err := PassphraseFromEnv()
	if err == nil {
		t.Fatalf("expected an error for an empty passphrase file, got %q", got)
	}
	if !strings.Contains(err.Error(), "is empty") {
		t.Errorf("error should say the file is empty, got: %v", err)
	}
}

// TestPassphraseTrimSymmetry ensures the inline and file forms resolve the
// same secret when one carries a trailing newline and the other doesn't.
func TestPassphraseTrimSymmetry(t *testing.T) {
	t.Setenv(EnvPassphraseFile, "")
	t.Setenv(EnvPassphrase, "sameSecret\n")
	inline, err := PassphraseFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if string(inline) != "sameSecret" {
		t.Fatalf("inline trailing newline not trimmed: %q", inline)
	}
	// Interior/leading spaces must survive.
	t.Setenv(EnvPassphrase, "  keep spaces  ")
	sp, err := PassphraseFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if string(sp) != "  keep spaces  " {
		t.Fatalf("spaces not preserved: %q", sp)
	}
}

func TestSealRejectsEmpty(t *testing.T) {
	if _, err := Seal(nil, []byte("pw")); err == nil {
		t.Fatal("expected error sealing empty key")
	}
	if _, err := Seal(randKey(t, 32), nil); err == nil {
		t.Fatal("expected error sealing with empty passphrase")
	}
}

func TestResolvePrivateKeyPlainBackwardCompat(t *testing.T) {
	// A plain base64 value must resolve identically to the old
	// base64.StdEncoding.DecodeString path, with no passphrase involved.
	raw := randKey(t, 32)
	plain := base64.StdEncoding.EncodeToString(raw)

	got, err := ResolvePrivateKey(plain, nil)
	if err != nil {
		t.Fatalf("ResolvePrivateKey(plain): %v", err)
	}
	if !bytes.Equal(got, raw) {
		t.Fatalf("plain resolve mismatch: got %x want %x", got, raw)
	}
}

func TestResolvePrivateKeySealed(t *testing.T) {
	raw := randKey(t, 32)
	pass := []byte("pw")
	blob, err := Seal(raw, pass)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ResolvePrivateKey(blob, pass)
	if err != nil {
		t.Fatalf("ResolvePrivateKey(sealed): %v", err)
	}
	if !bytes.Equal(got, raw) {
		t.Fatalf("sealed resolve mismatch: got %x want %x", got, raw)
	}
}

func TestResolvePrivateKeyInvalidPlain(t *testing.T) {
	if _, err := ResolvePrivateKey("not valid base64!!!", nil); err == nil {
		t.Fatal("expected error for invalid base64 plain key")
	}
}

func TestPassphraseFromEnv(t *testing.T) {
	t.Setenv(EnvPassphrase, "")
	t.Setenv(EnvPassphraseFile, "")
	if got, err := PassphraseFromEnv(); err != nil || got != nil {
		t.Fatalf("empty env: got (%q, %v) want (nil, nil)", got, err)
	}

	t.Setenv(EnvPassphrase, "inline-secret")
	got, err := PassphraseFromEnv()
	if err != nil || string(got) != "inline-secret" {
		t.Fatalf("inline env: got (%q, %v)", got, err)
	}

	// File form takes precedence and trailing newline is trimmed.
	dir := t.TempDir()
	fp := filepath.Join(dir, "pass")
	if writeErr := os.WriteFile(fp, []byte("file-secret\n"), 0600); writeErr != nil {
		t.Fatal(writeErr)
	}
	t.Setenv(EnvPassphraseFile, fp)
	got, err = PassphraseFromEnv()
	if err != nil || string(got) != "file-secret" {
		t.Fatalf("file env: got (%q, %v) want file-secret", got, err)
	}
}
