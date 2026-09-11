// Package keystore adds encryption-at-rest for NHP device private keys.
//
// Historically every daemon stored its private key as plain base64 in
// config.toml and decoded it at startup with base64.StdEncoding. Anyone
// who could read the config file (a backup, a stray copy, an accidental
// commit) read the key. This package lets the same config field instead
// hold a sealed blob that is decrypted at startup with a passphrase kept
// outside the config file (an env var or a mode-0600 file).
//
// The design goal is drop-in compatibility: ResolvePrivateKey accepts
// either form, so a plain base64 value keeps working unchanged and only
// deployments that opt in ever need a passphrase.
package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Sealed-blob format (fields separated by '$', binary fields base64 RawStd):
//
//	v1$argon2id$<time>$<memory>$<threads>$<salt>$<nonce>$<ciphertext>
//
// The prefix is self-describing so the reader never has to be told which
// form it is looking at, and a plain base64 key can never be mistaken for
// a blob because the standard base64 alphabet contains no '$'.
const (
	blobVersion = "v1"
	blobKDF     = "argon2id"

	// Argon2id parameters. 64 MiB / 3 passes / 4 lanes is the reference
	// "interactive" profile from the argon2 RFC draft — strong enough for
	// a key that is unsealed once at process start, cheap enough not to
	// noticeably slow startup.
	argonTime    = 3
	argonMemory  = 64 * 1024 // KiB => 64 MiB
	argonThreads = 4
	argonKeyLen  = 32 // AES-256 derived-key size

	// DeviceKeyLen is the length of an NHP device private scalar
	// (Curve25519 / SM2). It happens to equal argonKeyLen, but the two are
	// unrelated: this one is what Seal accepts and Open must recover, so a
	// truncated or mistyped key is caught at seal time rather than surfacing
	// much later at device creation. Exported so CLI wrappers don't keep
	// their own copy.
	DeviceKeyLen = 32

	// MinPassphraseLen is the shortest passphrase Seal accepts. A short
	// passphrase makes the Argon2id work factor moot; enforced in the
	// library so every producer (the seal CLI, RotateAgentKey, the register
	// re-seal) agrees instead of only the CLI.
	MinPassphraseLen = 8

	// Upper bounds on the KDF cost parsed out of a blob. The blob comes
	// from operator-controlled config (an attacker who can edit it can
	// swap the key outright), but a hostile or corrupt blob should still
	// not be able to drive argon2 into an enormous allocation / CPU spin
	// on startup. These ceilings are far above any sane real setting.
	maxArgonTime = 16
	// 256 MiB is 4x the default and far above any sensible real setting,
	// while still failing fast on an obviously bogus value rather than
	// letting a corrupt blob ask for a gigabyte at startup.
	maxArgonMemory = 256 * 1024 // KiB => 256 MiB

	saltLen = 16

	// gcmNonceSize is the standard AES-GCM nonce length. Declared here so a
	// blob's nonce can be length-checked BEFORE the (expensive) KDF runs.
	gcmNonceSize = 12

	// EnvPassphrase and EnvPassphraseFile name the two ways an operator
	// supplies the unseal passphrase without putting it in config.toml.
	// The file form wins if both are set so a file reference can override
	// a stale exported value.
	//
	// EnvPassphraseFile is the form to recommend. An environment variable
	// is readable from the process environment, is routinely written into
	// systemd units and orchestrator manifests, is captured in crash
	// dumps, and is inherited by every child process — much the same
	// exposure as the config file this feature exists to protect. A
	// 0600 file is a meaningfully smaller target; EnvPassphrase is there
	// for interactive and test use.
	EnvPassphrase     = "NHP_KEY_PASSPHRASE"
	EnvPassphraseFile = "NHP_KEY_PASSPHRASE_FILE"
)

var (
	// ErrNoPassphrase means the value is sealed but no passphrase was
	// provided to unseal it.
	ErrNoPassphrase = errors.New("keystore: private key is sealed but no passphrase was provided (set " + EnvPassphrase + " or " + EnvPassphraseFile + ")")
	// ErrBadPassphrase means decryption failed authentication — a wrong
	// passphrase or a tampered blob are indistinguishable by design.
	ErrBadPassphrase = errors.New("keystore: cannot unseal private key (wrong passphrase or corrupted blob)")
	// ErrMalformedBlob means the value carried the sealed prefix but did
	// not parse as a valid blob.
	ErrMalformedBlob = errors.New("keystore: malformed sealed key blob")
	// ErrUnsupportedVersion means the value is a well-formed sealed blob of a
	// format version this build does not understand — typically written by a
	// newer daemon. It is distinct from ErrMalformedBlob so the operator is
	// told to upgrade rather than to hunt for corruption.
	ErrUnsupportedVersion = errors.New("keystore: unsupported sealed-key format version")
)

// IsSealed reports whether a config value is a sealed blob rather than a
// plain base64 key. It matches any "v<N>$" prefix, not just the current
// version, so a blob from a newer format is still recognized as sealed (and
// Open then rejects it with a clear error) instead of being fed to a base64
// decoder or — worse — rotated to plaintext by RotateAgentKey.
func IsSealed(value string) bool {
	_, ok := sealedVersion(value)
	return ok
}

// sealedVersion extracts the "v<N>" version token from a sealed blob. ok is
// false when value is not a sealed blob at all (no "v<digits>$" prefix).
func sealedVersion(value string) (token string, ok bool) {
	rest, cut := strings.CutPrefix(value, "v")
	if !cut {
		return "", false
	}
	i := 0
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		i++
	}
	if i == 0 || i >= len(rest) || rest[i] != '$' {
		return "", false
	}
	return "v" + rest[:i], true
}

// Seal encrypts raw private-key bytes into a self-describing blob string
// suitable for storing in config.toml in place of the plain base64 key.
func Seal(privKeyRaw, passphrase []byte) (string, error) {
	if len(privKeyRaw) != DeviceKeyLen {
		// Mirror the length check Open applies to the recovered plaintext, so
		// the library can never mint a blob it would later refuse to open. An
		// empty key is just the most common way to hit this.
		if len(privKeyRaw) == 0 {
			return "", errors.New("keystore: refusing to seal an empty key")
		}
		return "", fmt.Errorf("keystore: refusing to seal a %d-byte key, expected %d", len(privKeyRaw), DeviceKeyLen)
	}
	if len(passphrase) < MinPassphraseLen {
		if len(passphrase) == 0 {
			return "", errors.New("keystore: refusing to seal with an empty passphrase")
		}
		return "", fmt.Errorf("keystore: passphrase is %d bytes; use at least %d", len(passphrase), MinPassphraseLen)
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("keystore: salt generation failed: %w", err)
	}

	key := argon2.IDKey(passphrase, salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	aead, err := newAEAD(key)
	zero(key) // the AES cipher has copied the key; drop our copy
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("keystore: nonce generation failed: %w", err)
	}
	// Bind the header (version, KDF name and cost parameters) into the AEAD as
	// additional authenticated data. Without this the params are parsed but
	// unauthenticated, so an attacker with write access could rewrite them
	// (or the version) and the only symptom would be a derived-key mismatch.
	// With it, any edit to parts[:5] makes Open fail authentication.
	aad := headerAAD(strconv.Itoa(argonTime), strconv.Itoa(argonMemory), strconv.Itoa(argonThreads))
	ct := aead.Seal(nil, nonce, privKeyRaw, aad)

	enc := base64.RawStdEncoding.EncodeToString
	blob := strings.Join([]string{
		blobVersion,
		blobKDF,
		strconv.Itoa(argonTime),
		strconv.Itoa(argonMemory),
		strconv.Itoa(argonThreads),
		enc(salt),
		enc(nonce),
		enc(ct),
	}, "$")
	return blob, nil
}

// headerAAD is the additional-authenticated-data bound into every sealed
// blob: "v1$argon2id$<t>$<m>$<p>". It ties the ciphertext to the version and
// KDF cost so those fields cannot be altered without failing authentication.
func headerAAD(t, m, p string) []byte {
	return []byte(strings.Join([]string{blobVersion, blobKDF, t, m, p}, "$"))
}

// Open decrypts a sealed blob back into raw private-key bytes.
func Open(blob string, passphrase []byte) ([]byte, error) {
	// Check the version BEFORE the passphrase: a well-formed "v<N>$" prefix
	// of a version we don't know is an upgrade problem, and no passphrase can
	// ever help — reporting "set NHP_KEY_PASSPHRASE" first would send the
	// operator down the wrong path. This is the reason IsSealed recognizes
	// forward versions in the first place.
	if token, ok := sealedVersion(blob); ok && token != blobVersion {
		return nil, fmt.Errorf("%w: blob is %s, this build understands %s — upgrade the daemon", ErrUnsupportedVersion, token, blobVersion)
	}

	if len(passphrase) == 0 {
		return nil, ErrNoPassphrase
	}

	parts := strings.Split(blob, "$")
	if len(parts) != 8 || parts[0] != blobVersion || parts[1] != blobKDF {
		return nil, ErrMalformedBlob
	}

	t, err1 := strconv.Atoi(parts[2])
	m, err2 := strconv.Atoi(parts[3])
	p, err3 := strconv.Atoi(parts[4])
	if err1 != nil || err2 != nil || err3 != nil ||
		t <= 0 || t > maxArgonTime ||
		m <= 0 || m > maxArgonMemory ||
		p <= 0 || p > 255 {
		return nil, ErrMalformedBlob
	}

	dec := base64.RawStdEncoding.DecodeString
	salt, err1 := dec(parts[5])
	nonce, err2 := dec(parts[6])
	ct, err3 := dec(parts[7])
	if err1 != nil || err2 != nil || err3 != nil {
		return nil, ErrMalformedBlob
	}
	// Validate the salt and nonce lengths before spending the KDF — a
	// malformed blob should fail fast, not after a 64 MiB argon2 pass.
	if len(salt) != saltLen || len(nonce) != gcmNonceSize {
		return nil, ErrMalformedBlob
	}

	key := argon2.IDKey(passphrase, salt, uint32(t), uint32(m), uint8(p), argonKeyLen)
	aead, err := newAEAD(key)
	zero(key) // the AES cipher has copied the key; drop our copy
	if err != nil {
		return nil, err
	}
	// The AAD must match exactly what Seal bound in; parts[2:5] are the
	// same strings Seal joined, so pass them through verbatim rather than
	// re-formatting the parsed ints.
	plain, err := aead.Open(nil, nonce, ct, headerAAD(parts[2], parts[3], parts[4]))
	if err != nil {
		// A wrong passphrase and a tampered blob both surface here as an
		// authentication failure; keep them indistinguishable.
		return nil, ErrBadPassphrase
	}
	// The AEAD proved integrity, so this is a real recovered key — but a
	// blob sealed from a truncated/mistyped key would only fail much later
	// at device creation. NHP device scalars are DeviceKeyLen bytes; mirror
	// the check Seal applies on the way in.
	if len(plain) != DeviceKeyLen {
		return nil, fmt.Errorf("%w: recovered key is %d bytes, expected %d", ErrMalformedBlob, len(plain), DeviceKeyLen)
	}
	return plain, nil
}

// zero best-effort wipes a byte slice holding key material. Go offers no
// guarantee the compiler won't keep copies elsewhere, so this is defense in
// depth, not a hard erasure — but it removes the obvious lingering copy.
func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// ResolvePrivateKey replaces base64.StdEncoding.DecodeString(cfg.PrivateKeyBase64)
// with a passphrase supplied by the caller. If the value is a sealed blob it
// is unsealed with that passphrase; otherwise it is treated as a plain
// base64 key exactly as before, so existing configs are unaffected.
//
// Every daemon in this repo calls ResolvePrivateKeyAuto instead, which reads
// the passphrase from the environment for you — this function is not the
// recommended entry point for that case. It exists for an embedder that
// sources the passphrase some other way (an OS keychain, an operator
// prompt, ...) and needs to hand it in directly.
func ResolvePrivateKey(cfgValue string, passphrase []byte) ([]byte, error) {
	if IsSealed(cfgValue) {
		return Open(cfgValue, passphrase)
	}
	return base64.StdEncoding.DecodeString(cfgValue)
}

// ResolvePrivateKeyAuto resolves cfgValue, consulting the environment for an
// unseal passphrase ONLY when the value is actually a sealed blob. A plain
// base64 key never triggers passphrase resolution, so a daemon whose key is
// not sealed is unaffected by a stray NHP_KEY_PASSPHRASE / _FILE exported in
// a shared profile or systemd unit. When the key IS sealed, a missing or
// unreadable passphrase is still a hard error (fail closed).
//
// The second return value is true when cfgValue was a sealed blob, so a
// caller can log that encryption-at-rest is actually in effect.
func ResolvePrivateKeyAuto(cfgValue string) (key []byte, sealed bool, err error) {
	if !IsSealed(cfgValue) {
		raw, decErr := base64.StdEncoding.DecodeString(cfgValue)
		return raw, false, decErr
	}
	pass, passErr := PassphraseFromEnv()
	if passErr != nil {
		return nil, true, passErr
	}
	raw, openErr := Open(cfgValue, pass)
	return raw, true, openErr
}

// PassphraseFromEnv resolves the unseal passphrase from the environment.
// NHP_KEY_PASSPHRASE_FILE (a path to a file whose contents are the
// passphrase) takes precedence over the inline NHP_KEY_PASSPHRASE. It
// returns nil with no error when neither is set, so callers on the plain
// path never need a passphrase.
//
// Both forms strip a single trailing newline (\n or \r\n) so the SAME
// secret resolves identically whether it is exported inline or read from a
// file written with `echo`/an editor. A passphrase that legitimately ends
// in a newline is not supportable this way — an unlikely case for a secret.
func PassphraseFromEnv() ([]byte, error) {
	if path := os.Getenv(EnvPassphraseFile); path != "" {
		// The path is an operator-supplied config input by design.
		data, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			// Fail closed: if the operator explicitly set the file var but
			// the file can't be read, abort startup rather than silently
			// falling through to a keyless/plain path. A misconfigured
			// passphrase file should surface loudly, not be ignored.
			return nil, fmt.Errorf("keystore: cannot read %s=%q: %w", EnvPassphraseFile, path, err)
		}
		pass := trimTrailingNewline(string(data))
		if pass == "" {
			// Distinguish "you pointed me at an empty file" from "no
			// passphrase was configured" — the operator clearly intended
			// to supply one, so say what's actually wrong.
			return nil, fmt.Errorf("keystore: %s=%q is empty", EnvPassphraseFile, path)
		}
		return []byte(pass), nil
	}
	if pass := os.Getenv(EnvPassphrase); pass != "" {
		return []byte(trimTrailingNewline(pass)), nil
	}
	return nil, nil
}

// trimTrailingNewline removes one trailing "\n" or "\r\n" and nothing else,
// so interior or trailing spaces in a passphrase are preserved.
func trimTrailingNewline(s string) string {
	s = strings.TrimSuffix(s, "\n")
	s = strings.TrimSuffix(s, "\r")
	return s
}

func newAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("keystore: cipher init failed: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("keystore: GCM init failed: %w", err)
	}
	return aead, nil
}
