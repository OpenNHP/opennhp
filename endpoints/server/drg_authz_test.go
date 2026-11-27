package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
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
	if err := SaveZdtoConfig(first, "db-a", ""); err != nil {
		t.Fatal(err)
	}
	if got, _ := ReadZdtoConfig(id); got.OwnerPublicKey != "db-a" {
		t.Fatalf("owner = %q, want the authenticated sender", got.OwnerPublicKey)
	}

	takeover := testDRG(id)
	takeover.Spo.Policy = "b3RoZXI=" // another nhp-db's policy
	if err := SaveZdtoConfig(takeover, "db-b", ""); !errors.Is(err, errNotObjectOwner) {
		t.Fatalf("update by another nhp-db: got %v, want errNotObjectOwner", err)
	}

	renamed := testDRG(id)
	renamed.DbId = "db-2"
	if err := SaveZdtoConfig(renamed, "db-a", ""); !errors.Is(err, errDbIdChange) {
		t.Fatalf("dbId change: got %v, want errDbIdChange", err)
	}

	update := testDRG(id)
	update.Spo.Policy = "bmV3IHBvbGljeQ=="
	if err := SaveZdtoConfig(update, "db-a", ""); err != nil {
		t.Fatalf("update by the owner: %v", err)
	}
	if got, _ := ReadZdtoConfig(id); got.Spo.Policy != update.Spo.Policy || got.OwnerPublicKey != "db-a" {
		t.Fatalf("owner update not stored: %+v", got)
	}
}

// An object stored before owners were recorded is claimed by the first
// update whose dbId matches, sent by the nhp-db online with that dbId; a
// different dbId, or an nhp-db not holding the dbId, cannot claim it.
func TestZdtoConfigLegacyObjectClaim(t *testing.T) {
	old := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = old })
	const id = "legacy-object"

	legacy := testDRG(id)
	if err := SaveZdtoConfig(legacy, "db-a", ""); err != nil {
		t.Fatal(err)
	}
	// Simulate a config written before this change: no owner.
	stored, _ := ReadZdtoConfig(id)
	stored.OwnerPublicKey = ""
	writeRawZdto(t, &stored)

	other := testDRG(id)
	other.DbId = "someone-else"
	if err := SaveZdtoConfig(other, "db-b", ""); !errors.Is(err, errDbIdChange) {
		t.Fatalf("legacy object claimed under another dbId: got %v", err)
	}
	// The dbId in a DRG is only a claim: an nhp-db that is not online with
	// the object's dbId cannot claim it, even naming the right dbId.
	for name, holder := range map[string]string{"nobody online": "", "another nhp-db online": "db-b"} {
		if err := SaveZdtoConfig(testDRG(id), "db-a", holder); !errors.Is(err, errNotObjectOwner) {
			t.Fatalf("legacy claim, %s: got %v, want errNotObjectOwner", name, err)
		}
	}
	if got, _ := ReadZdtoConfig(id); got.OwnerPublicKey != "" {
		t.Fatalf("refused claim stored an owner: %q", got.OwnerPublicKey)
	}
	if err := SaveZdtoConfig(testDRG(id), "db-a", "db-a"); err != nil {
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
	if err := SaveZdtoConfig(noPolicy, "db-a", ""); !errors.Is(err, errEmptyPolicy) {
		t.Fatalf("empty policy: got %v, want errEmptyPolicy", err)
	}
	if err := SaveZdtoConfig(testDRG("no-owner"), "", ""); !errors.Is(err, errNotDBPeer) {
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

// HandleDHPDRGMessage passes the dbId's online holder to SaveZdtoConfig:
// a legacy object is claimed only by the nhp-db online with its dbId, not
// by another configured nhp-db naming the same dbId, and not by anyone
// while nobody is online with it (a one-shot "nhp-db encrypt" run).
func TestHandleDHPDRGMessageLegacyClaimNeedsOnlineHolder(t *testing.T) {
	old := ExeDirPath
	ExeDirPath = t.TempDir()
	t.Cleanup(func() { ExeDirPath = old })
	const id = "legacy-via-handler"

	s, peers := dbOnlineTestServer(t, "key-a", "key-b")
	keyA := peers["key-a"].PubKeyBase64

	if err := SaveZdtoConfig(testDRG(id), keyA, ""); err != nil {
		t.Fatal(err)
	}
	stored, _ := ReadZdtoConfig(id)
	stored.OwnerPublicKey = ""
	writeRawZdto(t, &stored)

	drg := func(rawKey string) {
		t.Helper()
		body, err := json.Marshal(testDRG(id))
		if err != nil {
			t.Fatal(err)
		}
		// No transaction to answer on: the handler saves (or refuses)
		// first, then ends at ErrTransactionIdNotFound.
		err = s.HandleDHPDRGMessage(&core.PacketParserData{
			ConnData:     &core.ConnectionData{RemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.50"), Port: 42000}},
			BodyMessage:  body,
			RemotePubKey: []byte(rawKey),
			SenderTrxId:  1,
		})
		if !errors.Is(err, common.ErrTransactionIdNotFound) {
			t.Fatalf("HandleDHPDRGMessage = %v, want ErrTransactionIdNotFound (no transaction)", err)
		}
	}
	owner := func() string {
		t.Helper()
		got, err := ReadZdtoConfig(id)
		if err != nil {
			t.Fatal(err)
		}
		return got.OwnerPublicKey
	}

	drg("key-a")
	if got := owner(); got != "" {
		t.Fatalf("claimed while nobody is online with the dbId: owner %q", got)
	}

	if _, err := dbOnline(t, s, "key-a", "db-1", 42001); !errors.Is(err, common.ErrTransactionIdNotFound) {
		t.Fatalf("key-a online: %v", err)
	}
	drg("key-b")
	if got := owner(); got != "" {
		t.Fatalf("claimed by an nhp-db not online with the dbId: owner %q", got)
	}

	drg("key-a")
	if got := owner(); got != keyA {
		t.Fatalf("holder's claim: owner %q, want %q", got, keyA)
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

	if err := SaveZdtoConfig(testDRG(id), "db-a", ""); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(ExeDirPath, "etc", "ztdo", "data-"+id+".json")
	if err := os.WriteFile(path, []byte(`{"doId": "corrupt-obj`), 0o600); err != nil {
		t.Fatal(err)
	}

	takeover := testDRG(id)
	takeover.Spo.Policy = "b3RoZXI="
	if err := SaveZdtoConfig(takeover, "db-b", ""); !errors.Is(err, errSaveConfigFailed) {
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
	if err := SaveZdtoConfig(testDRG("never-registered"), "db-b", ""); err != nil {
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
