package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
)

// Only an nhp-db listed, and not expired, in db.toml may send NHP_DRG.
func TestAuthorizeDRGSender(t *testing.T) {
	dbKey := []byte("0123456789abcdef0123456789abcdef")
	expiredKey := []byte("fedcba9876543210fedcba9876543210")
	agentKey := []byte("agentagentagentagentagentagentag")
	b64 := base64.StdEncoding.EncodeToString
	s := &UdpServer{dbPeerMap: map[string]*core.UdpPeer{
		b64(dbKey):      {PubKeyBase64: b64(dbKey), Type: core.NHP_DB},
		b64(expiredKey): {PubKeyBase64: b64(expiredKey), Type: core.NHP_DB, ExpireTime: time.Now().Add(-time.Hour).Unix()},
	}}

	if owner, err := s.authorizeDRGSender(dbKey); err != nil || owner != b64(dbKey) {
		t.Fatalf("listed nhp-db: owner %q, err %v", owner, err)
	}
	for name, key := range map[string][]byte{"agent key": agentKey, "expired nhp-db": expiredKey, "no key": nil} {
		if _, err := s.authorizeDRGSender(key); !errors.Is(err, errNotDBPeer) {
			t.Errorf("%s: got %v, want errNotDBPeer", name, err)
		}
	}
}

// An object is bound to the nhp-db that registered it: another nhp-db can
// neither rewrite its policy nor take it over, and its dbId is fixed.
func TestZdtoConfigOwnership(t *testing.T) {
	old := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = old })
	const id = "owned-object"

	first := testDRG(id)
	first.OwnerPublicKey = "forged-by-sender" // never trusted
	if err := SaveZdtoConfig(first, "db-a"); err != nil {
		t.Fatal(err)
	}
	if got, _ := ReadZdtoConfig(id); got.OwnerPublicKey != "db-a" {
		t.Fatalf("owner = %q, want the authenticated sender", got.OwnerPublicKey)
	}

	takeover := testDRG(id)
	takeover.Spo.Policy = "b3RoZXI=" // another nhp-db's policy
	if err := SaveZdtoConfig(takeover, "db-b"); !errors.Is(err, errNotObjectOwner) {
		t.Fatalf("update by another nhp-db: got %v, want errNotObjectOwner", err)
	}

	renamed := testDRG(id)
	renamed.DbId = "db-2"
	if err := SaveZdtoConfig(renamed, "db-a"); !errors.Is(err, errDbIdChange) {
		t.Fatalf("dbId change: got %v, want errDbIdChange", err)
	}

	update := testDRG(id)
	update.Spo.Policy = "bmV3IHBvbGljeQ=="
	if err := SaveZdtoConfig(update, "db-a"); err != nil {
		t.Fatalf("update by the owner: %v", err)
	}
	if got, _ := ReadZdtoConfig(id); got.Spo.Policy != update.Spo.Policy || got.OwnerPublicKey != "db-a" {
		t.Fatalf("owner update not stored: %+v", got)
	}
}

// An object stored before owners were recorded is claimed by the first
// update whose dbId matches; a different dbId cannot claim it.
func TestZdtoConfigLegacyObjectClaim(t *testing.T) {
	old := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = old })
	const id = "legacy-object"

	legacy := testDRG(id)
	if err := SaveZdtoConfig(legacy, "db-a"); err != nil {
		t.Fatal(err)
	}
	// Simulate a config written before this change: no owner.
	stored, _ := ReadZdtoConfig(id)
	stored.OwnerPublicKey = ""
	writeRawZdto(t, &stored)

	other := testDRG(id)
	other.DbId = "someone-else"
	if err := SaveZdtoConfig(other, "db-b"); !errors.Is(err, errDbIdChange) {
		t.Fatalf("legacy object claimed under another dbId: got %v", err)
	}
	if err := SaveZdtoConfig(testDRG(id), "db-a"); err != nil {
		t.Fatalf("legacy object update with its dbId: %v", err)
	}
	if got, _ := ReadZdtoConfig(id); got.OwnerPublicKey != "db-a" {
		t.Fatalf("legacy object not claimed: owner %q", got.OwnerPublicKey)
	}
}

func TestZdtoConfigRequiresPolicyAndOwner(t *testing.T) {
	old := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = old })

	noPolicy := testDRG("no-policy")
	noPolicy.Spo.Policy = ""
	if err := SaveZdtoConfig(noPolicy, "db-a"); !errors.Is(err, errEmptyPolicy) {
		t.Fatalf("empty policy: got %v, want errEmptyPolicy", err)
	}
	if err := SaveZdtoConfig(testDRG("no-owner"), ""); !errors.Is(err, errNotDBPeer) {
		t.Fatalf("no owner: got %v, want errNotDBPeer", err)
	}
}

// Attestation fails closed for an object without a policy.
func TestAttestationEmptyPolicyFailsClosed(t *testing.T) {
	s := &UdpServer{}
	if err := s.onAttestationVerify(&common.SmartPolicy{}, "evidence"); !errors.Is(err, errEmptyPolicy) {
		t.Fatalf("empty policy: got %v, want errEmptyPolicy", err)
	}
}

// writeRawZdto stores msg as is, bypassing SaveZdtoConfig's checks.
func writeRawZdto(t *testing.T, msg *common.DRGMsg) {
	t.Helper()
	b, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(ExeDirPath, "etc", "ztdo", "data-"+msg.DoId+".json")
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

// A stored object that exists but cannot be read is not treated as a new
// object: otherwise any nhp-db could take it over by registering the same
// DoId while the file is corrupt or temporarily unreadable.
func TestZdtoConfigUnreadableObjectFailsClosed(t *testing.T) {
	old := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = old })
	const id = "corrupt-object"

	if err := SaveZdtoConfig(testDRG(id), "db-a"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(ExeDirPath, "etc", "ztdo", "data-"+id+".json")
	if err := os.WriteFile(path, []byte(`{"doId": "corrupt-obj`), 0o600); err != nil {
		t.Fatal(err)
	}

	takeover := testDRG(id)
	takeover.Spo.Policy = "b3RoZXI="
	if err := SaveZdtoConfig(takeover, "db-b"); !errors.Is(err, errSaveConfigFailed) {
		t.Fatalf("takeover of a corrupt object: got %v, want errSaveConfigFailed", err)
	}
	if b, _ := os.ReadFile(path); string(b) != `{"doId": "corrupt-obj` {
		t.Fatal("the corrupt object was overwritten")
	}

	// A truly new object is still accepted, and a missing one reads as
	// not found.
	if _, err := ReadZdtoConfig("never-registered"); !errors.Is(err, errZdtoNotFound) || !errors.Is(err, errReadConfigFailed) {
		t.Fatalf("missing object: got %v, want errZdtoNotFound wrapping errReadConfigFailed", err)
	}
	if err := SaveZdtoConfig(testDRG("never-registered"), "db-b"); err != nil {
		t.Fatalf("new object: %v", err)
	}
}

// Agents get one message for every ZTDO read failure, so they cannot tell
// a missing object from an unreadable one; other errors pass through.
func TestZdtoReadErrorForAgent(t *testing.T) {
	for name, err := range map[string]error{
		"missing":    errZdtoNotFound,
		"unreadable": errReadConfigFailed,
	} {
		if got := zdtoReadErrorForAgent(err); got != errReadConfigFailed.Error() {
			t.Errorf("%s: got %q, want %q", name, got, errReadConfigFailed.Error())
		}
	}
	if got := zdtoReadErrorForAgent(common.ErrInvalidDoID); got != common.ErrInvalidDoID.Error() {
		t.Errorf("invalid DoId: got %q", got)
	}
}

// DAK answers do not tell one nhp-db whether a DoId exists or who owns it.
func TestDRGErrorForSender(t *testing.T) {
	for name, err := range map[string]error{"owned by another": errNotObjectOwner, "dbId change": errDbIdChange} {
		if got := drgErrorForSender(err); got != errUnauthorized.Error() {
			t.Errorf("%s: got %q, want %q", name, got, errUnauthorized.Error())
		}
	}
	if got := drgErrorForSender(errEmptyPolicy); got != errEmptyPolicy.Error() {
		t.Errorf("the sender's own mistake should pass through, got %q", got)
	}
}
