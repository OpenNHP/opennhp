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
			if keystore.IsSealed(priv) {
				return fmt.Errorf("that value is already a sealed blob — use `%s unseal` to recover the key, not `seal`", bin)
			}
			raw, err := base64.StdEncoding.DecodeString(priv)
			if err != nil {
				return fmt.Errorf("decode private key: %w", err)
			}
			if len(raw) != keystore.DeviceKeyLen {
				return fmt.Errorf("private key decodes to %d bytes, expected %d — is it the right value?", len(raw), keystore.DeviceKeyLen)
			}
			// PassphraseFromEnv itself warns if NHP_KEY_PASSPHRASE_FILE points
			// at a world/group-readable file — covers every consumer, not
			// just this command.
			pass, err := keystore.PassphraseFromEnv()
			if err != nil {
				return err
			}
			if len(pass) == 0 {
				return fmt.Errorf("no passphrase set: export %s or %s before sealing", keystore.EnvPassphrase, keystore.EnvPassphraseFile)
			}
			if len(pass) < keystore.MinPassphraseLen {
				return fmt.Errorf("passphrase is %d bytes; use at least %d — Argon2id does not make a short passphrase safe against a stolen config", len(pass), keystore.MinPassphraseLen)
			}
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
			pass, err := keystore.PassphraseFromEnv()
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
