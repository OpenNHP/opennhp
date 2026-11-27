package server

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/plugins"
)

// TestStartKeepsRelayPeersLoadedFromRelayToml reproduces the "nhp-server
// reads relay.toml and then loses it" report end to end.
//
// s.relayPeerMap is the config-side index of the relay peers, owned by
// updateRelayPeers (config.go): loadPeers() fills it from etc/relay.toml
// and the reload path swaps it wholesale. Start() used to create the map
// a second time, in the runtime-state block that runs AFTER loadPeers()
// had returned, so the peers read from relay.toml were dropped on the
// floor. Every NHP_RLY envelope was then refused by dedupeRecvRLY with
// ErrServerMissingPeerPubkey — the hook fails closed when the sender is
// not in this map — so the relay path stayed broken until the operator
// touched relay.toml and the fsnotify callback re-ran updateRelayPeers.
//
// This test starts a real server over a temporary config tree and checks
// the invariant the report is about: after Start() returns, the relay
// peer map still holds the relay configured in relay.toml.
func TestStartKeepsRelayPeersLoadedFromRelayToml(t *testing.T) {
	dir := t.TempDir()
	etcDir := filepath.Join(dir, "etc")
	if err := os.MkdirAll(etcDir, 0755); err != nil {
		t.Fatalf("create etc dir: %v", err)
	}

	priv := make([]byte, 32)
	rand.Read(priv)
	cookie := make([]byte, 32)
	rand.Read(cookie)
	relayPubKey := make([]byte, 32)
	rand.Read(relayPubKey)

	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(etcDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	write("config.toml", fmt.Sprintf(`
PrivateKeyBase64 = %q
DefaultCipherScheme = 0
ListenIp = "127.0.0.1"
ListenPort = 0
Hostname = "localhost:0"
LogLevel = 1
DatabasePath = %q
CookieSigningKeyBase64 = %q
`, base64.StdEncoding.EncodeToString(priv),
		filepath.Join(dir, "data", "nhp_server.db"),
		base64.StdEncoding.EncodeToString(cookie)))
	write("ac.toml", "")
	write("agent.toml", "")
	write("relay.toml", fmt.Sprintf("[[Relays]]\nPubKeyBase64 = %q\nExpireTime = 1924991999\n",
		base64.StdEncoding.EncodeToString(relayPubKey)))
	write("http.toml", "")
	write("srcip.toml", "")
	write("resource.toml", "")
	write("tee.toml", "")

	// Start() installs the config directory globally; restore it so the
	// rest of the package sees the tree it expects.
	originalExeDir := ExeDirPath
	originalPluginsExeDir := plugins.ExeDirPath
	t.Cleanup(func() {
		ExeDirPath = originalExeDir
		plugins.ExeDirPath = originalPluginsExeDir
	})

	s := &UdpServer{}
	if err := s.Start(dir, 1); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(s.Stop)

	s.relayPeerMapMutex.Lock()
	got := len(s.relayPeerMap)
	_, found := s.relayPeerMap[base64.StdEncoding.EncodeToString(relayPubKey)]
	s.relayPeerMapMutex.Unlock()

	if got != 1 || !found {
		t.Fatalf("after Start() the relay peer map holds %d entries (configured relay present=%v), want 1 — "+
			"the peers read from relay.toml were dropped, so every NHP_RLY envelope is refused with "+
			"ErrServerMissingPeerPubkey until relay.toml is touched", got, found)
	}
}

// TestStartInitializesRelayPeerMapBeforeLoadingPeers pins the startup
// ordering behind the test above, so the map cannot silently move back
// into the runtime-state block: it must be created exactly once in
// Start(), and that creation must come before the call that loads the
// peer configs.
func TestStartInitializesRelayPeerMapBeforeLoadingPeers(t *testing.T) {
	src, err := os.ReadFile("udpserver.go")
	if err != nil {
		t.Fatalf("read udpserver.go: %v", err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "udpserver.go", src, 0)
	if err != nil {
		t.Fatalf("parse udpserver.go: %v", err)
	}

	var start *ast.FuncDecl
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Start" || fn.Recv == nil {
			continue
		}
		start = fn
		break
	}
	if start == nil {
		t.Fatal("no Start method found in udpserver.go")
	}

	var relayMapInits []token.Pos
	var loadPeersCalls []token.Pos
	ast.Inspect(start.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				sel, ok := lhs.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "relayPeerMap" {
					continue
				}
				if _, ok := sel.X.(*ast.Ident); !ok {
					continue
				}
				relayMapInits = append(relayMapInits, node.Pos())
			}
		case *ast.CallExpr:
			sel, ok := node.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "loadPeers" {
				return true
			}
			if _, ok := sel.X.(*ast.Ident); !ok {
				return true
			}
			loadPeersCalls = append(loadPeersCalls, node.Pos())
		}
		return true
	})

	if len(relayMapInits) != 1 {
		t.Fatalf("Start() assigns s.relayPeerMap %d times, want exactly 1 — a second assignment re-creates the map and drops the relay peers loaded from relay.toml", len(relayMapInits))
	}
	if len(loadPeersCalls) != 1 {
		t.Fatalf("Start() calls loadPeers() %d times, want exactly 1", len(loadPeersCalls))
	}
	initLine := fset.Position(relayMapInits[0]).Line
	loadLine := fset.Position(loadPeersCalls[0]).Line
	if initLine >= loadLine {
		t.Fatalf("Start() creates s.relayPeerMap on line %d, after the loadPeers() call on line %d — "+
			"the relay peers read from relay.toml are dropped before the first packet arrives", initLine, loadLine)
	}
}

// TestUpdateRelayPeers_MakesRelayVisibleToDedupe is the functional half
// of the same invariant: the peers installed by updateRelayPeers must be
// visible to the dedupe hook, otherwise every NHP_RLY envelope from a
// configured relay is refused with ErrServerMissingPeerPubkey (the hook
// fails closed on an unknown sender).
func TestUpdateRelayPeers_MakesRelayVisibleToDedupe(t *testing.T) {
	devPriv := make([]byte, 32)
	for i := range devPriv {
		devPriv[i] = byte(i + 1)
	}
	dev := core.NewDevice(core.NHP_SERVER, devPriv, nil)
	if dev == nil {
		t.Fatal("NewDevice returned nil")
	}
	t.Cleanup(dev.Stop)

	s := &UdpServer{device: dev, relayReplay: newRelayReplayWindow()}
	// Startup order: the map is created before the peer configs load.
	s.relayPeerMap = make(map[string]*core.UdpPeer)

	relayPubKey := artTestPubkey(0x21, core.PublicKeySize)
	relayPeer := &core.UdpPeer{
		PubKeyBase64: base64.StdEncoding.EncodeToString(relayPubKey),
		Type:         core.NHP_RELAY,
	}
	if err := s.updateRelayPeers([]*core.UdpPeer{relayPeer}); err != nil {
		t.Fatalf("updateRelayPeers: %v", err)
	}

	packet := &core.PacketParserData{
		HeaderType:     core.NHP_RLY,
		RemotePubKey:   relayPubKey,
		SenderTrxId:    1,
		RemoteSendTime: 1000,
	}
	if err := s.dedupeRecvRLY(packet); err != nil {
		t.Fatalf("NHP_RLY from the configured relay rejected: %v", err)
	}
}
