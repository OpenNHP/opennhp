package keystorecli

import (
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
