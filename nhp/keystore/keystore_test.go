package keystore

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
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
	pass := []byte("test-passphrase")
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
	blob, err := Seal(key, []byte("right-passphrase"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Open(blob, []byte("wrong-passphrase")); err != ErrBadPassphrase {
		t.Fatalf("wrong passphrase: got %v want ErrBadPassphrase", err)
	}
}

func TestOpenTamperedCiphertext(t *testing.T) {
	key := randKey(t, 32)
	pass := []byte("test-passphrase")
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
	blob, err := Seal(randKey(t, 32), []byte("test-passphrase"))
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
		if _, err := Open(blob, []byte("test-passphrase")); err == nil {
			t.Fatalf("expected error for malformed blob %q", blob)
		}
	}
}

// TestOpenRejectsOversizedKDFParams ensures a hostile blob cannot drive
// argon2 into a huge allocation: out-of-range time/memory are rejected as
// malformed before the KDF runs.
func TestOpenRejectsOversizedKDFParams(t *testing.T) {
	// Build a structurally valid blob but swap in an absurd memory cost.
	blob, err := Seal(randKey(t, 32), []byte("test-passphrase"))
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(blob, "$")
	parts[3] = "999999999" // memory KiB, far above maxArgonMemory
	if _, err := Open(strings.Join(parts, "$"), []byte("test-passphrase")); err != ErrMalformedBlob {
		t.Fatalf("oversized memory: got %v want ErrMalformedBlob", err)
	}

	parts = strings.Split(blob, "$")
	parts[2] = "9999" // time, far above maxArgonTime
	if _, err := Open(strings.Join(parts, "$"), []byte("test-passphrase")); err != ErrMalformedBlob {
		t.Fatalf("oversized time: got %v want ErrMalformedBlob", err)
	}
}

// TestOpenRejectsBadNonceLength ensures a wrong-length nonce is rejected as
// malformed (fast-fail before the KDF).
func TestOpenRejectsBadNonceLength(t *testing.T) {
	blob, err := Seal(randKey(t, 32), []byte("test-passphrase"))
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(blob, "$")
	parts[6] = base64.RawStdEncoding.EncodeToString([]byte("shortnonce")) // != 12 bytes
	if _, err := Open(strings.Join(parts, "$"), []byte("test-passphrase")); err != ErrMalformedBlob {
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
	if _, err := Seal(nil, []byte("test-passphrase")); err == nil {
		t.Fatal("expected error sealing empty key")
	}
	if _, err := Seal(randKey(t, 32), nil); err == nil {
		t.Fatal("expected error sealing with empty passphrase")
	}
}

// TestSealEnforcesPassphraseFloor: the MinPassphraseLen floor lives in the
// library now, so RotateAgentKey / the register re-seal can't slip a short
// passphrase past the CLI check.
func TestSealEnforcesPassphraseFloor(t *testing.T) {
	key := randKey(t, 32)
	if _, err := Seal(key, []byte("short")); err == nil {
		t.Fatal("Seal accepted a 5-byte passphrase")
	}
	if _, err := Seal(key, []byte("12345678")); err != nil { // exactly MinPassphraseLen
		t.Fatalf("Seal rejected an 8-byte passphrase: %v", err)
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
	pass := []byte("test-passphrase")
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

// TestResolvePrivateKeyAuto pins the decoupling guarantee every daemon relies
// on: a PLAIN base64 key never touches the passphrase env vars (so a stray
// exported NHP_KEY_PASSPHRASE_FILE cannot wedge an unsealed daemon), while a
// SEALED value does consult them.
func TestResolvePrivateKeyAuto(t *testing.T) {
	raw := randKey(t, 32)
	plain := base64.StdEncoding.EncodeToString(raw)
	pass := []byte("test-passphrase")
	blob, err := Seal(raw, pass)
	if err != nil {
		t.Fatal(err)
	}

	// A bogus passphrase file is set. The plain path must ignore it entirely.
	t.Setenv(EnvPassphrase, "")
	t.Setenv(EnvPassphraseFile, filepath.Join(t.TempDir(), "does-not-exist"))

	got, sealed, err := ResolvePrivateKeyAuto(plain)
	if err != nil || sealed {
		t.Fatalf("plain: got (sealed=%v, err=%v), want (false, nil) despite the bad passphrase file", sealed, err)
	}
	if !bytes.Equal(got, raw) {
		t.Fatalf("plain resolve mismatch: %x vs %x", got, raw)
	}

	// The sealed path DOES consult the env — with the bad file still set it
	// must fail closed, not fall through.
	if _, sealedBad, badErr := ResolvePrivateKeyAuto(blob); badErr == nil || !sealedBad {
		t.Fatalf("sealed with an unreadable passphrase file: got (sealed=%v, err=%v), want (true, error)", sealedBad, badErr)
	}

	// With a real passphrase the sealed path round-trips.
	fp := filepath.Join(t.TempDir(), "pass")
	if wErr := os.WriteFile(fp, pass, 0o600); wErr != nil {
		t.Fatal(wErr)
	}
	t.Setenv(EnvPassphraseFile, fp)
	got, sealed, err = ResolvePrivateKeyAuto(blob)
	if err != nil || !sealed || !bytes.Equal(got, raw) {
		t.Fatalf("sealed resolve: got (sealed=%v, err=%v, key match=%v)", sealed, err, bytes.Equal(got, raw))
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

// TestOpenRejectsHeaderTampering checks that the version/KDF/cost fields are
// authenticated: editing any of parts[:5] to a still-in-range value makes
// Open fail rather than silently deriving a different key.
func TestOpenRejectsHeaderTampering(t *testing.T) {
	pass := []byte("correct horse battery staple")
	blob, err := Seal(randKey(t, 32), pass)
	if err != nil {
		t.Fatal(err)
	}

	// Sanity: the untouched blob opens.
	if _, err := Open(blob, pass); err != nil {
		t.Fatalf("baseline Open: %v", err)
	}

	// Drop the time cost from 3 to 1 (still within bounds). Without AAD this
	// would just derive a different key and surface as ErrBadPassphrase; with
	// AAD it is an authentication failure either way, but the point is Open
	// must NOT succeed.
	parts := strings.Split(blob, "$")
	if parts[2] != "3" {
		t.Fatalf("unexpected time cost %q", parts[2])
	}
	parts[2] = "1"
	if _, err := Open(strings.Join(parts, "$"), pass); err == nil {
		t.Fatal("Open accepted a blob with an altered time cost")
	}

	// Same for the memory cost.
	parts = strings.Split(blob, "$")
	parts[3] = "32768"
	if _, err := Open(strings.Join(parts, "$"), pass); err == nil {
		t.Fatal("Open accepted a blob with an altered memory cost")
	}
}

// TestOpenRejectsBadSaltLength ensures a wrong-length salt is rejected as
// malformed before the KDF runs, matching the existing nonce-length guard.
func TestOpenRejectsBadSaltLength(t *testing.T) {
	blob, err := Seal(randKey(t, 32), []byte("test-passphrase"))
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(blob, "$")
	parts[5] = base64.RawStdEncoding.EncodeToString([]byte("short")) // != 16 bytes
	if _, err := Open(strings.Join(parts, "$"), []byte("test-passphrase")); err != ErrMalformedBlob {
		t.Fatalf("bad salt length: got %v want ErrMalformedBlob", err)
	}
}

// TestIsSealedRecognizesAnyVersion: a future "v2$..." blob must still be
// recognized as sealed (so it is not base64-decoded or rotated to plaintext),
// while plain keys and near-misses are not.
func TestIsSealedRecognizesAnyVersion(t *testing.T) {
	sealed := []string{"v1$argon2id$3$65536$4$a$b$c", "v2$whatever", "v10$x"}
	for _, s := range sealed {
		if !IsSealed(s) {
			t.Errorf("IsSealed(%q) = false, want true", s)
		}
	}
	plain := []string{
		"eHdyRHKJy/YZJsResCt5XTAZgtcwvLpSXAiZ8DBc0V4=", // real base64 key
		"v$nodigits", "vabc$x", "version1$x", "", "v1", "v1noDelim",
	}
	for _, s := range plain {
		if IsSealed(s) {
			t.Errorf("IsSealed(%q) = true, want false", s)
		}
	}
}

// TestOpenRejectsWrongLengthRecoveredKey: a blob that unseals to something
// other than a 32-byte scalar is rejected at Open, not much later at device
// creation. Seal now enforces the length on the way in, so the blob is
// hand-crafted here to exercise Open's guard directly.
func TestOpenRejectsWrongLengthRecoveredKey(t *testing.T) {
	pass := []byte("correct horse battery staple")

	salt := randKey(t, saltLen)
	dk := argon2.IDKey(pass, salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	aead, err := newAEAD(dk)
	if err != nil {
		t.Fatal(err)
	}
	nonce := randKey(t, gcmNonceSize)
	aad := headerAAD(strconv.Itoa(argonTime), strconv.Itoa(argonMemory), strconv.Itoa(argonThreads))
	ct := aead.Seal(nil, nonce, []byte("only-eight"), aad) // 10-byte plaintext

	enc := base64.RawStdEncoding.EncodeToString
	blob := strings.Join([]string{
		blobVersion, blobKDF,
		strconv.Itoa(argonTime), strconv.Itoa(argonMemory), strconv.Itoa(argonThreads),
		enc(salt), enc(nonce), enc(ct),
	}, "$")

	if _, err := Open(blob, pass); err == nil {
		t.Fatal("Open accepted a blob that unseals to a non-32-byte key")
	}
}

// TestSealRejectsWrongLengthKey: Seal must refuse a key that is not a
// 32-byte device scalar, mirroring Open, so the library cannot mint a blob
// it would later refuse to open.
func TestSealRejectsWrongLengthKey(t *testing.T) {
	if _, err := Seal([]byte("too short"), []byte("test-passphrase")); err == nil {
		t.Fatal("Seal accepted a short key")
	}
	if _, err := Seal(randKey(t, 33), []byte("test-passphrase")); err == nil {
		t.Fatal("Seal accepted an over-length key")
	}
	if _, err := Seal(randKey(t, 32), []byte("test-passphrase")); err != nil {
		t.Fatalf("Seal rejected a valid 32-byte key: %v", err)
	}
}

// TestOpenUnknownVersionIsDistinctError: a well-formed blob of an unknown
// version reports ErrUnsupportedVersion, not ErrMalformedBlob, so the
// operator is told to upgrade rather than to hunt for corruption.
func TestOpenUnknownVersionIsDistinctError(t *testing.T) {
	_, err := Open("v2$argon2id$3$65536$4$AAAA$BBBB$CCCC", []byte("test-passphrase"))
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("unknown version: got %v want ErrUnsupportedVersion", err)
	}
	if errors.Is(err, ErrMalformedBlob) {
		t.Fatal("unknown version should not also be ErrMalformedBlob")
	}
	// The version is reported even with no passphrase — no passphrase can
	// ever open a forward-version blob, so "upgrade" must win over "set a
	// passphrase".
	if _, noPassErr := Open("v2$argon2id$3$65536$4$AAAA$BBBB$CCCC", nil); !errors.Is(noPassErr, ErrUnsupportedVersion) {
		t.Fatalf("unknown version with no passphrase: got %v want ErrUnsupportedVersion", noPassErr)
	}
}
