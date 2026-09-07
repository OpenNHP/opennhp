// Package keystorecli provides the `seal` / `unseal` CLI subcommands shared
// by every NHP daemon. Sealed private-key blobs are consumed by nhp-server,
// nhp-ac, nhp-db, nhp-relay and nhp-agent alike, so the tooling to produce
// one must be available on each binary — an operator running only nhp-acd
// should not have to install nhp-serverd just to seal the AC key.
package keystorecli

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/OpenNHP/opennhp/nhp/keystore"
)

// rawKeyLen is the fixed length of an NHP device private key (Curve25519 and
// SM2 scalars are both 32 bytes). A mistyped or truncated key is caught here
// rather than much later at device init.
const rawKeyLen = 32

// minPassphraseLen is a floor enforced by the CLI only (not the library, so
// unit tests stay cheap). Argon2id does not rescue a two-character
// passphrase against a stolen config.toml.
const minPassphraseLen = 8

// SealCommand returns the `seal` subcommand: read a base64 private key from
// stdin and print a sealed blob for config.toml. The passphrase comes from
// the environment (NHP_KEY_PASSPHRASE / NHP_KEY_PASSPHRASE_FILE) so it never
// appears on the command line; the key is read from stdin rather than argv
// for the same reason (argv is visible in shell history and /proc).
func SealCommand() *cli.Command {
	return &cli.Command{
		Name:  "seal",
		Usage: "encrypt a base64 private key (read from stdin) into a sealed blob for config.toml",
		Action: func(c *cli.Context) error {
			bin := c.App.Name
			if c.Args().Len() > 0 {
				return fmt.Errorf("seal reads the private key from stdin, not from an argument "+
					"(an argv key leaks into shell history and /proc); try: echo \"$KEY\" | %s seal", bin)
			}
			if fi, statErr := os.Stdin.Stat(); statErr == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
				fmt.Fprintln(os.Stderr, "reading base64 private key from stdin (Ctrl-D to end)…")
			}
			input, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read private key from stdin: %w", err)
			}
			priv := strings.TrimSpace(string(input))
			if priv == "" {
				return fmt.Errorf("no private key on stdin; usage: echo \"$KEY\" | %s seal", bin)
			}
			raw, err := base64.StdEncoding.DecodeString(priv)
			if err != nil {
				return fmt.Errorf("decode private key: %w", err)
			}
			if len(raw) != rawKeyLen {
				return fmt.Errorf("private key decodes to %d bytes, expected %d — is it the right value?", len(raw), rawKeyLen)
			}
			pass, source, err := passphraseFromEnv()
			if err != nil {
				return err
			}
			if len(pass) == 0 {
				return fmt.Errorf("no passphrase set: export %s or %s before sealing", keystore.EnvPassphrase, keystore.EnvPassphraseFile)
			}
			if len(pass) < minPassphraseLen {
				return fmt.Errorf("passphrase is %d bytes; use at least %d — Argon2id does not make a short passphrase safe against a stolen config", len(pass), minPassphraseLen)
			}
			warnIfWorldReadable(source)
			blob, err := keystore.Seal(raw, pass)
			if err != nil {
				return err
			}
			fmt.Println(blob)
			return nil
		},
	}
}

// UnsealCommand returns the `unseal` subcommand: decrypt a sealed blob back
// to a plain base64 private key (for debugging or key rotation). The blob is
// not secret on its own so it is fine as an argument; the recovered key it
// prints is, hence the stderr warning.
func UnsealCommand() *cli.Command {
	return &cli.Command{
		Name:      "unseal",
		Usage:     "decrypt a sealed key blob back to a base64 private key (output is secret)",
		ArgsUsage: "<sealedBlob>",
		Action: func(c *cli.Context) error {
			blob := c.Args().First()
			if blob == "" {
				return fmt.Errorf("usage: %s unseal <sealedBlob>", c.App.Name)
			}
			pass, _, err := passphraseFromEnv()
			if err != nil {
				return err
			}
			raw, err := keystore.Open(blob, pass)
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, "warning: the private key below is plaintext — keep it out of logs and scrollback")
			fmt.Println(base64.StdEncoding.EncodeToString(raw))
			return nil
		},
	}
}

// Commands returns both subcommands, ready to append to app.Commands.
func Commands() []*cli.Command {
	return []*cli.Command{SealCommand(), UnsealCommand()}
}

// passphraseFromEnv wraps keystore.PassphraseFromEnv and also reports which
// source it came from, so seal can warn about a world-readable file.
func passphraseFromEnv() (pass []byte, source string, err error) {
	if path := os.Getenv(keystore.EnvPassphraseFile); path != "" {
		source = path
	} else if os.Getenv(keystore.EnvPassphrase) != "" {
		source = "env:" + keystore.EnvPassphrase
	}
	pass, err = keystore.PassphraseFromEnv()
	return pass, source, err
}

// warnIfWorldReadable prints a warning when the passphrase file is readable
// by group or other. The docs recommend mode 0600; nothing enforced it.
func warnIfWorldReadable(source string) {
	if source == "" || strings.HasPrefix(source, "env:") {
		return
	}
	fi, err := os.Stat(source)
	if err != nil {
		return
	}
	if fi.Mode().Perm()&0o077 != 0 {
		fmt.Fprintf(os.Stderr, "warning: passphrase file %s is mode %o — restrict it to 0600\n", source, fi.Mode().Perm())
	}
}
