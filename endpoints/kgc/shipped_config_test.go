package kgc

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// The config the Makefile ships must not carry a master key.
//
// `make kgc` copies kgc/main/etc/*.toml into release/nhp-kgc/etc/, so a
// value committed here is the same IBC master private key in every
// installation — and in IBC the master private key derives the private key
// of every identity the KGC serves. The `setup` command exists to generate
// it per deployment; it is only mandatory because ReadConfig refuses to
// start while the field is empty, which it cannot do if the file arrives
// pre-filled.
func TestShippedConfigHasNoMasterKey(t *testing.T) {
	path := filepath.Join("main", "etc", "config.toml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	for _, field := range []string{"MasterPrivateKeyBase64", "MasterPublicKeyBase64"} {
		re := regexp.MustCompile(`(?m)^\s*` + field + `\s*=\s*"([^"]*)"`)
		m := re.FindSubmatch(content)
		if m == nil {
			t.Errorf("%s: %s is absent; keep it present and empty so the file documents what setup fills in", path, field)
			continue
		}
		if len(m[1]) != 0 {
			t.Errorf("%s: %s ships a value (%d chars). Run `nhp-kgc setup` per deployment instead; a committed master key is the same key everywhere.",
				path, field, len(m[1]))
		}
	}
}
