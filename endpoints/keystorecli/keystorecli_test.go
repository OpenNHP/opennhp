package keystorecli

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"

	"github.com/OpenNHP/opennhp/nhp/keystore"
)

func TestCommandsShape(t *testing.T) {
	cmds := Commands()
	if len(cmds) != 2 {
		t.Fatalf("Commands() returned %d, want 2", len(cmds))
	}
	if cmds[0].Name != "seal" || cmds[1].Name != "unseal" {
		t.Fatalf("unexpected command names: %s, %s", cmds[0].Name, cmds[1].Name)
	}
}

// TestSealRejectsArgvKey ensures the key cannot be passed as an argument
// (which would leak it into shell history / /proc).
func TestSealRejectsArgvKey(t *testing.T) {
	t.Setenv(keystore.EnvPassphraseFile, "")
	t.Setenv(keystore.EnvPassphrase, "correct horse battery staple")

	app := &cli.App{Name: "nhp-test", Commands: Commands()}
	err := app.Run([]string{"nhp-test", "seal", "somekey"})
	if err == nil {
		t.Fatal("seal accepted an argv key")
	}
}

// TestUnsealRoundTrip drives seal's library path and unseal together.
func TestUnsealRoundTrip(t *testing.T) {
	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i + 1)
	}
	blob, err := keystore.Seal(raw, []byte("correct horse battery staple"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(keystore.EnvPassphraseFile, "")
	t.Setenv(keystore.EnvPassphrase, "correct horse battery staple")

	app := &cli.App{Name: "nhp-test", Commands: Commands()}
	if err := app.Run([]string{"nhp-test", "unseal", blob}); err != nil {
		t.Fatalf("unseal: %v", err)
	}
}

// TestSealDoesNotWarnOnStdout: a permissive passphrase file must not put any
// warning text on stdout, which this command uses for the sealed blob
// itself — a mixed-in warning line would corrupt `BLOB=$(... seal)` or
// `... seal > config-fragment`. Regression test for the warning having been
// routed through nhp/log's package-default logger, which writes to stdout
// until a daemon's Start replaces it (never, for this command).
func TestSealDoesNotWarnOnStdout(t *testing.T) {
	dir := t.TempDir()
	passFile := filepath.Join(dir, "pass")
	if err := os.WriteFile(passFile, []byte("correct horse battery staple"), 0o644); err != nil { // permissive on purpose
		t.Fatal(err)
	}
	t.Setenv(keystore.EnvPassphraseFile, passFile)
	t.Setenv(keystore.EnvPassphrase, "")

	raw := make([]byte, 32)
	for i := range raw {
		raw[i] = byte(i + 1)
	}
	key := base64.StdEncoding.EncodeToString(raw)

	oldStdin, oldStdout := os.Stdin, os.Stdout
	inR, inW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin, os.Stdout = inR, outW
	defer func() { os.Stdin, os.Stdout = oldStdin, oldStdout }()

	go func() {
		_, _ = inW.WriteString(key)
		_ = inW.Close()
	}()

	app := &cli.App{Name: "nhp-test", Commands: Commands()}
	runErr := app.Run([]string{"nhp-test", "seal"})
	_ = outW.Close()
	out, readErr := io.ReadAll(outR)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if runErr != nil {
		t.Fatalf("seal: %v (stdout: %q)", runErr, out)
	}

	got := strings.TrimSpace(string(out))
	if strings.Contains(got, "\n") || !strings.HasPrefix(got, "v1$") {
		t.Fatalf("stdout must be exactly the sealed blob, nothing else; got %q", got)
	}
	recovered, uErr := keystore.Open(got, []byte("correct horse battery staple"))
	if uErr != nil {
		t.Fatalf("stdout did not unseal to the original key: %v", uErr)
	}
	if !bytes.Equal(recovered, raw) {
		t.Fatalf("recovered key does not match the original: got %x want %x", recovered, raw)
	}
}
