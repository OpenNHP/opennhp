package db

import (
	"errors"
	"os"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/common"
)

func TestDataPrivateKeyStoreRejectsInvalidDoID(t *testing.T) {
	originalExeDir := common.ExeDirPath
	common.ExeDirPath = t.TempDir()
	t.Cleanup(func() { common.ExeDirPath = originalExeDir })

	store := NewDataPrivateKeyStore("provider-key")
	invalidDoID := "foo/../../../bar"

	if _, err := NewDataPrivateKeyStoreWith(invalidDoID); !errors.Is(err, common.ErrInvalidDoID) {
		t.Fatalf("NewDataPrivateKeyStoreWith error = %v, want ErrInvalidDoID", err)
	}
	if err := store.Save(invalidDoID); !errors.Is(err, common.ErrInvalidDoID) {
		t.Fatalf("Save error = %v, want ErrInvalidDoID", err)
	}
	if err := store.Delete(invalidDoID); !errors.Is(err, common.ErrInvalidDoID) {
		t.Fatalf("Delete error = %v, want ErrInvalidDoID", err)
	}

	entries, err := os.ReadDir(common.ExeDirPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("invalid DoId changed the filesystem: %v", entries)
	}
}

func TestDataPrivateKeyStoreRoundTrip(t *testing.T) {
	originalExeDir := common.ExeDirPath
	common.ExeDirPath = t.TempDir()
	t.Cleanup(func() { common.ExeDirPath = originalExeDir })

	const doID = "550e8400-e29b-41d4-a716-446655440000"
	want := &DataPrivateKeyStore{
		DataPrivateKeyBase64:    "private-key",
		ProviderPublicKeyBase64: "provider-key",
	}
	if err := want.Save(doID); err != nil {
		t.Fatalf("Save returned %v", err)
	}

	got, err := NewDataPrivateKeyStoreWith(doID)
	if err != nil {
		t.Fatalf("NewDataPrivateKeyStoreWith returned %v", err)
	}
	if got.DataPrivateKeyBase64 != want.DataPrivateKeyBase64 || got.ProviderPublicKeyBase64 != want.ProviderPublicKeyBase64 {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}

	if err := got.Delete(doID); err != nil {
		t.Fatalf("Delete returned %v", err)
	}
	if _, err := NewDataPrivateKeyStoreWith(doID); !errors.Is(err, common.ErrDataPrivateKeyStore) {
		t.Fatalf("read after delete error = %v, want ErrDataPrivateKeyStore", err)
	}
}
