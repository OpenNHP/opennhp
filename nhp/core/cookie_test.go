package core

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/common"
)

// TestCookieRemoteKey_DirectUDPUsesRemoteAddr: when there is no relay
// in play, ConnData.RealRemoteAddr is nil and the cookie key must fall
// back to RemoteAddr. This is the single-server / direct-UDP path that
// existed before relay forwarding was added; it must keep working.
func TestCookieRemoteKey_DirectUDPUsesRemoteAddr(t *testing.T) {
	cd := &ConnectionData{RemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.7"), Port: 9000}}
	if got := cookieRemoteKey(cd); got != "203.0.113.7" {
		t.Fatalf("direct-UDP key: got %q, want %q", got, "203.0.113.7")
	}
}

// TestCookieRemoteKey_RelayForwardUsesRealRemoteAddr: this is the
// regression the relay-cookie fix is about. For relay-forwarded
// packets ConnData.RemoteAddr is the relay's UDP address and
// RealRemoteAddr is the actual client. Keying on RemoteAddr would
// make every client of one relay share a cookie — so the helper MUST
// prefer RealRemoteAddr when set.
func TestCookieRemoteKey_RelayForwardUsesRealRemoteAddr(t *testing.T) {
	cd := &ConnectionData{
		RemoteAddr:     &net.UDPAddr{IP: net.ParseIP("198.51.100.10"), Port: 62206}, // relay
		RealRemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.42"), Port: 51234},  // client
	}
	if got := cookieRemoteKey(cd); got != "203.0.113.42" {
		t.Fatalf("relay-forward key: got %q (relay address would have given %q), want %q",
			got, "198.51.100.10", "203.0.113.42")
	}
}

// TestCookieRemoteKey_DistinctClientsViaSameRelay: two clients
// arriving through the same relay must derive distinct cookie keys.
// Pre-fix they would have collided on the relay's IP — this test
// pins the property going forward.
func TestCookieRemoteKey_DistinctClientsViaSameRelay(t *testing.T) {
	relay := &net.UDPAddr{IP: net.ParseIP("198.51.100.10"), Port: 62206}
	a := &ConnectionData{
		RemoteAddr:     relay,
		RealRemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.42"), Port: 51234},
	}
	b := &ConnectionData{
		RemoteAddr:     relay,
		RealRemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.43"), Port: 51235},
	}
	if cookieRemoteKey(a) == cookieRemoteKey(b) {
		t.Fatalf("two clients via the same relay must produce distinct cookie keys; got %q for both",
			cookieRemoteKey(a))
	}
}

// TestCookieRemoteKey_IPv4MappedIPv6Normalized: a dual-stack listener
// can present the same agent as "::ffff:1.2.3.4" on one socket and
// "1.2.3.4" on another. The helper must normalize both forms so the
// cookie a sibling server minted under one form is still verifiable
// under the other.
func TestCookieRemoteKey_IPv4MappedIPv6Normalized(t *testing.T) {
	plain := &ConnectionData{RemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.7"), Port: 9000}}
	mapped := &ConnectionData{RemoteAddr: &net.UDPAddr{IP: net.ParseIP("::ffff:203.0.113.7"), Port: 9000}}
	if cookieRemoteKey(plain) != cookieRemoteKey(mapped) {
		t.Fatalf("IPv4 and IPv4-mapped-IPv6 must derive the same key; got %q vs %q",
			cookieRemoteKey(plain), cookieRemoteKey(mapped))
	}
}

// TestCookieRemoteKey_NilSafe: defensive callers — nil ConnData, nil
// addresses, nil IP — must return "" rather than panic. sendCookie /
// checkHMAC route the result into an HMAC input where "" is harmless
// (HMAC of zero-length is well-defined), so the right failure mode is
// "compare fails" not "process crashes".
func TestCookieRemoteKey_NilSafe(t *testing.T) {
	if got := cookieRemoteKey(nil); got != "" {
		t.Fatalf("nil ConnData: got %q, want empty", got)
	}
	if got := cookieRemoteKey(&ConnectionData{}); got != "" {
		t.Fatalf("empty ConnData: got %q, want empty", got)
	}
	if got := cookieRemoteKey(&ConnectionData{RemoteAddr: &net.UDPAddr{}}); got != "" {
		t.Fatalf("nil IP: got %q, want empty", got)
	}
}

// TestDeriveStatelessCookie_PeerPkBindsIdentity: this is the property
// the "stateless cookie has no per-handshake binding" fix is about.
// Two distinct agents arriving from the same source IP (same NAT/CGN
// egress) inside the same time window must derive distinct cookies, so
// one can't mint a cookie via its own legitimate KNK and have a
// sibling behind the same NAT replay it in an unrelated RKN.
func TestDeriveStatelessCookie_PeerPkBindsIdentity(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	remote := "203.0.113.7"
	pkA := bytes.Repeat([]byte{0xAA}, PublicKeySize)
	pkB := bytes.Repeat([]byte{0xBB}, PublicKeySize)
	const window int64 = 12345

	cookieA := deriveStatelessCookie(key, remote, pkA, window)
	cookieB := deriveStatelessCookie(key, remote, pkB, window)
	if bytes.Equal(cookieA, cookieB) {
		t.Fatalf("distinct peerPks behind the same NAT must derive distinct cookies; got identical %x", cookieA)
	}
}

// TestDeriveStatelessCookie_SiblingServerReDerivation: any sibling
// nhp-server with the same signing key, observing the same agent
// (same source IP + same agent static pubkey) inside the same time
// window, must produce the same cookie. This is the cluster routing
// invariant — if it ever breaks, the KNK → COK → RKN handshake stops
// surviving load-balancer shuffling between instances.
func TestDeriveStatelessCookie_SiblingServerReDerivation(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	remote := "203.0.113.7"
	pk := bytes.Repeat([]byte{0xAA}, PublicKeySize)
	const window int64 = 12345

	// Two independent derivations (e.g. mint on instance #1, verify
	// on instance #2) must agree byte-for-byte.
	a := deriveStatelessCookie(key, remote, pk, window)
	b := deriveStatelessCookie(key, remote, pk, window)
	if !bytes.Equal(a, b) {
		t.Fatalf("identical inputs must derive identical cookies; got %x vs %x", a, b)
	}
}

// TestDeriveStatelessCookie_WindowSeparation: cookies expire across
// window rollover. Verify changes to the window index alone produce a
// distinct cookie, so an attacker can't replay a 2-window-old cookie
// indefinitely.
func TestDeriveStatelessCookie_WindowSeparation(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	remote := "203.0.113.7"
	pk := bytes.Repeat([]byte{0xAA}, PublicKeySize)

	curr := deriveStatelessCookie(key, remote, pk, 12345)
	next := deriveStatelessCookie(key, remote, pk, 12346)
	if bytes.Equal(curr, next) {
		t.Fatalf("consecutive windows must derive distinct cookies; got identical %x", curr)
	}
}

// TestSendCookie_WireCounterMatchesAgentKnk fences the
// browser→relay→cluster timeout regression under Overload. The HTTP
// relay matches the server's COK back to the pending agent request
// by comparing the wire header counter against the counter inside
// the agent's KNK header (see endpoints/relay/relay.go). Every other
// server→agent path (ACK/AOP/RAK/LRT/…) routes through
// PrevParserData so deriveMsgAssemblerData copies ppd.SenderTrxId
// onto the wire — sendCookie was the lone exception, taking the
// non-prev branch of createMsgAssemblerData and stamping the
// server's own NextCounterIndex() on the wire. The mismatch made
// the relay silently drop the COK and the browser hit
// UDPTimeoutMs → 504.
//
// The test calls sendCookie directly, drains the MsgData it queues,
// then runs that MsgData through createMsgAssemblerData (the same
// step the msgToPacketRoutine worker performs) and asserts the
// resulting wire header counter equals ppd.SenderTrxId — i.e. that
// a relay sitting in front of this server would find a matching
// pending request. Done end-to-end through sendCookie so a
// regression that drops PrevParserData or reintroduces a
// server-side TransactionId fails here.
func TestSendCookie_WireCounterMatchesAgentKnk(t *testing.T) {
	dev := newDeviceForChainKeyTest(t)
	// sendCookie aborts if cookie params are unset (it's only
	// reachable on a server, where startup installs them); install
	// dummies so the function reaches the SendMsgToPacket call.
	dev.SetStatelessCookieParams(bytes.Repeat([]byte{0x42}, 32), 5)

	const agentTrxId uint64 = 0xA11CE0DEADBEEF42
	ppd := &PacketParserData{
		device:       dev,
		CipherScheme: common.CIPHER_SCHEME_CURVE,
		Ciphers:      NewCipherSuite(common.CIPHER_SCHEME_CURVE),
		RemotePubKey: testPeerPk(),
		SenderTrxId:  agentTrxId,
		ConnData: &ConnectionData{
			RemoteAddr: &net.UDPAddr{IP: net.ParseIP("203.0.113.7"), Port: 51234},
		},
	}

	ppd.sendCookie()

	// newDeviceForChainKeyTest does not call dev.Start(), so the
	// msgToPacketRoutine workers are not draining the queue —
	// sendCookie's MsgData is sitting there waiting for us.
	var md *MsgData
	select {
	case md = <-dev.msgToPacketQueue:
	default:
		t.Fatal("sendCookie did not enqueue an MsgData")
	}

	if md.PrevParserData != ppd {
		t.Fatalf("sendCookie must build MsgData with PrevParserData set so deriveMsgAssemblerData stamps the agent's SenderTrxId on the wire; got PrevParserData=%v", md.PrevParserData)
	}
	if md.TransactionId != 0 {
		t.Fatalf("sendCookie must not assign a server-side TransactionId — that overwrites the agent's counter on the wire and breaks relay correlation; got TransactionId=%#x", md.TransactionId)
	}

	// Run the same finalization step the worker would, then check
	// the on-wire counter — this is what the relay actually reads.
	mad, err := dev.createMsgAssemblerData(md)
	if err != nil {
		t.Fatalf("createMsgAssemblerData: %v", err)
	}
	t.Cleanup(mad.Destroy)

	got := mad.header.Counter()
	if got != agentTrxId {
		t.Fatalf("COK wire counter must equal agent's SenderTrxId so the HTTP relay can match the response.\n"+
			"got=%#x want=%#x — sendCookie likely regressed to the no-prev branch and stamped a server-side counter on the wire",
			got, agentTrxId)
	}
}

// TestDeriveStatelessCookie_EmptyPeerPkSafe: pre-fix call sites (and
// pathological inputs) might pass a nil/empty peerPk. The HMAC
// primitive accepts zero-length writes, so the function must not
// panic. The resulting cookie won't match any well-formed RKN's HMAC
// — fail-closed is the right outcome.
func TestDeriveStatelessCookie_EmptyPeerPkSafe(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, 32)
	remote := "203.0.113.7"

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("empty peerPk must not panic, got %v", r)
		}
	}()
	got := deriveStatelessCookie(key, remote, nil, 12345)
	if len(got) != CookieSize {
		t.Fatalf("expected %d-byte cookie even with nil peerPk, got %d", CookieSize, len(got))
	}
}

// TestStatelessCookieParams_CopiesOut verifies the getter returns an
// independent buffer, not an alias of the device's internal key. A caller
// mutating the returned slice — or a concurrent SetStatelessCookieParams
// rebinding the field — must not be observable through a previously
// returned slice.
func TestStatelessCookieParams_CopiesOut(t *testing.T) {
	dev := newDeviceForChainKeyTest(t)
	orig := bytes.Repeat([]byte{0xAB}, 32)
	dev.SetStatelessCookieParams(orig, 5)

	got, win := dev.StatelessCookieParams()
	if win != 5 {
		t.Fatalf("window = %d, want 5", win)
	}
	if !bytes.Equal(got, orig) {
		t.Fatalf("returned key %x, want %x", got, orig)
	}

	// Mutating the returned slice must not corrupt the device's key.
	got[0] ^= 0xFF
	again, _ := dev.StatelessCookieParams()
	if !bytes.Equal(again, orig) {
		t.Fatalf("device key was aliased: mutation leaked, got %x want %x", again, orig)
	}

	// Rebinding via the setter must not be visible through the slice the
	// first call already returned.
	first, _ := dev.StatelessCookieParams()
	newKey := bytes.Repeat([]byte{0xCD}, 32)
	dev.SetStatelessCookieParams(newKey, 7)
	if first[0] == 0xCD {
		t.Fatal("previously returned slice aliased the internal key across a setter rebind")
	}
	after, win2 := dev.StatelessCookieParams()
	if win2 != 7 || !bytes.Equal(after, newKey) {
		t.Fatalf("after rebind got (%x, %d), want (%x, 7)", after, win2, newKey)
	}
}

// TestStatelessCookieParams_NilWhenDisabled: a zero/empty key disables
// cookies and the getter returns nil (not an empty non-nil slice).
func TestStatelessCookieParams_NilWhenDisabled(t *testing.T) {
	dev := newDeviceForChainKeyTest(t)
	dev.SetStatelessCookieParams(nil, 5)
	got, win := dev.StatelessCookieParams()
	if got != nil {
		t.Fatalf("disabled cookies should return nil key, got %x", got)
	}
	if win != 0 {
		t.Fatalf("disabled cookies should return window 0, got %d", win)
	}
}

// newCurveDevice builds a real curve25519 device from a deterministic
// 32-byte private key (filled with `seed`). Unlike newDeviceForChainKeyTest
// the public key here is a *real* point, so device-ECDH actually works —
// required for the cookie-verify e2e tests below, where the server runs
// extractInitiatorStaticPubKey (a real ECDH + AEAD-Open) against the
// agent's IK static field.
func newCurveDevice(t *testing.T, deviceType int, seed byte) *Device {
	t.Helper()
	priv := make([]byte, 32)
	for i := range priv {
		priv[i] = seed
	}
	dev := NewDevice(deviceType, priv, nil)
	if dev == nil {
		t.Fatalf("NewDevice(seed=%d) returned nil", seed)
	}
	t.Cleanup(dev.Stop)
	return dev
}

// curvePubKey returns a device's curve25519 static public key bytes.
func curvePubKey(dev *Device) []byte {
	return dev.GetEcdhByCipherScheme(common.CIPHER_SCHEME_CURVE).PublicKey()
}

// buildAgentRKNWithCookie drives the *agent* assembly path to produce a
// fully-encrypted NHP_RKN whose HMAC was summed over `cookie` (i.e. the
// real addHMAC(sumCookie=true) at initiator.go). MsgToPacket runs
// createMsgAssemblerData → setPeerPublicKey → encryptBody, and encryptBody
// calls addHMAC(HeaderType==NHP_RKN). Returns a copy of the wire bytes the
// agent would put on the socket.
func buildAgentRKNWithCookie(t *testing.T, agentDev *Device, serverPubKey []byte, cookie *[CookieSize]byte, remote *net.UDPAddr) []byte {
	t.Helper()
	md := &MsgData{
		HeaderType:     NHP_RKN,
		CipherScheme:   common.CIPHER_SCHEME_CURVE,
		TransactionId:  0xC00C1E00DEADBEEF,
		PeerPk:         serverPubKey,
		Message:        []byte(`{"hello":"rkn"}`),
		ExternalCookie: cookie,
		ConnData:       &ConnectionData{RemoteAddr: remote},
	}
	mad, err := agentDev.MsgToPacket(md)
	if err != nil {
		t.Fatalf("agent MsgToPacket(NHP_RKN): %v", err)
	}
	// MsgToPacket defers mad.Destroy() (it releases the pool packet on
	// return), so copy the wire bytes out before they're recycled.
	wire := make([]byte, len(mad.BasePacket.Content))
	copy(wire, mad.BasePacket.Content)
	return wire
}

// parseRKNOnServer feeds raw agent wire bytes through the *server* parse
// pipeline. With the device in overload, createPacketParserData routes
// NHP_RKN to checkHMAC(sumCookie=true) — the production verify path,
// including extractInitiatorStaticPubKey. Returns the parse error (nil on
// HMAC success).
func parseRKNOnServer(t *testing.T, serverDev *Device, wire []byte, remote *net.UDPAddr) error {
	t.Helper()
	pkt := serverDev.AllocatePoolPacket()
	if pkt == nil {
		t.Fatal("AllocatePoolPacket returned nil")
	}
	copy(pkt.Buf[:], wire)
	pkt.Content = pkt.Buf[:len(wire)]

	serverDev.SetOverload(true)
	defer serverDev.SetOverload(false)

	ppd, err := serverDev.createPacketParserData(&PacketData{
		BasePacket: pkt,
		ConnData:   &ConnectionData{RemoteAddr: remote},
		InitTime:   time.Now().UnixNano(),
	})
	if ppd != nil {
		t.Cleanup(ppd.Destroy)
	}
	return err
}

// TestCookieVerify_EndToEnd is the regression the reviewer asked for: it
// exercises the full cookie *verify* path that only runs in production —
// checkHMAC(sumCookie=true), including extractInitiatorStaticPubKey — and
// holds it byte-for-byte against the agent's addHMAC. The unit tests above
// cover deriveStatelessCookie / cookieRemoteKey / sendCookie in isolation,
// but none runs the verify side, so a divergence between the agent's HMAC
// input (initiator.go addHMAC: Hash(Init||ServerPub||header[:n]||cookie))
// and the server's by-hand rebuild (responder.go:809-812) would compile,
// pass every other test, and silently break every cross-replica RKN under
// overload — the headline feature of this change.
//
// The agent mints nothing itself: the cluster mints the cookie (as a
// sibling server would) bound to the agent's static pubkey + source IP +
// window, hands it to the agent, and the agent stamps it into a real RKN.
// The server then re-derives and verifies. This is the KNK→COK→RKN loop
// minus the wire hop.
func TestCookieVerify_EndToEnd(t *testing.T) {
	const window = 5 // seconds
	signingKey := bytes.Repeat([]byte{0x42}, 32)

	agentDev := newCurveDevice(t, NHP_AGENT, 0x11)
	serverDev := newCurveDevice(t, NHP_SERVER, 0x22)
	serverDev.SetStatelessCookieParams(signingKey, window)

	remote := &net.UDPAddr{IP: net.ParseIP("203.0.113.7"), Port: 51234}
	serverPub := curvePubKey(serverDev)
	agentPub := curvePubKey(agentDev)

	// Mint the cookie exactly as a server instance would in sendCookie:
	// bound to (source IP, agent static pubkey, current window).
	remoteKey := cookieRemoteKey(&ConnectionData{RemoteAddr: remote})
	curWindow := time.Now().Unix() / int64(window)
	raw := deriveStatelessCookie(signingKey, remoteKey, agentPub, curWindow)
	var cookie [CookieSize]byte
	copy(cookie[:], raw)

	wire := buildAgentRKNWithCookie(t, agentDev, serverPub, &cookie, remote)

	if err := parseRKNOnServer(t, serverDev, wire, remote); err != nil {
		t.Fatalf("cookie verify failed end-to-end: %v\n"+
			"the agent's addHMAC input (initiator.go) and the server's "+
			"rebuild (responder.go ~809) have diverged, or "+
			"extractInitiatorStaticPubKey failed to recover the agent pubkey",
			err)
	}
}

// TestCookieVerify_CrossReplica locks the cluster invariant: a cookie minted
// by one server instance must verify on a *different* instance that shares
// only the signing key + window (no shared connection state). If this breaks,
// load-balancer reshuffling between KNK and RKN starts failing silently —
// which is exactly what stateless cookies exist to prevent.
func TestCookieVerify_CrossReplica(t *testing.T) {
	const window = 5
	signingKey := bytes.Repeat([]byte{0x42}, 32)

	agentDev := newCurveDevice(t, NHP_AGENT, 0x11)
	// Two server instances with DISTINCT static keys (distinct identities)
	// but the SAME cookie signing key — i.e. two replicas behind an LB.
	// The minter mints below by hand (a sibling server's sendCookie); the
	// verifier — a separate identity — must accept that cookie.
	verifier := newCurveDevice(t, NHP_SERVER, 0x33)
	verifier.SetStatelessCookieParams(signingKey, window)

	remote := &net.UDPAddr{IP: net.ParseIP("203.0.113.7"), Port: 51234}
	agentPub := curvePubKey(agentDev)

	// The minter issues the cookie (bound to agent identity + source IP +
	// window). Because the derivation never references the responding
	// instance's identity, the verifier replica must compute the identical
	// cookie from the same signing key — that is the cross-replica
	// invariant under test.
	remoteKey := cookieRemoteKey(&ConnectionData{RemoteAddr: remote})
	curWindow := time.Now().Unix() / int64(window)
	raw := deriveStatelessCookie(signingKey, remoteKey, agentPub, curWindow)
	var cookie [CookieSize]byte
	copy(cookie[:], raw)

	// The IK static field is encrypted to whichever instance the agent's
	// RKN actually reaches, so assemble against the verifier's pubkey: its
	// extractInitiatorStaticPubKey must then recover the agent key and
	// re-derive the minter's cookie.
	wire := buildAgentRKNWithCookie(t, agentDev, curvePubKey(verifier), &cookie, remote)
	if err := parseRKNOnServer(t, verifier, wire, remote); err != nil {
		t.Fatalf("cookie minted by instance A failed to verify on instance B: %v", err)
	}
}

// TestCookieVerify_RejectsWrongCookie confirms the verify path actually
// rejects — a guard that the e2e success test alone can't give, since a
// checkHMAC that returned true unconditionally would pass it. A cookie
// bound to a different source IP must fail.
func TestCookieVerify_RejectsWrongCookie(t *testing.T) {
	const window = 5
	signingKey := bytes.Repeat([]byte{0x42}, 32)

	agentDev := newCurveDevice(t, NHP_AGENT, 0x11)
	serverDev := newCurveDevice(t, NHP_SERVER, 0x22)
	serverDev.SetStatelessCookieParams(signingKey, window)

	remote := &net.UDPAddr{IP: net.ParseIP("203.0.113.7"), Port: 51234}
	agentPub := curvePubKey(agentDev)

	// Cookie bound to a DIFFERENT source IP than the RKN arrives from.
	curWindow := time.Now().Unix() / int64(window)
	raw := deriveStatelessCookie(signingKey, "198.51.100.99", agentPub, curWindow)
	var cookie [CookieSize]byte
	copy(cookie[:], raw)

	wire := buildAgentRKNWithCookie(t, agentDev, curvePubKey(serverDev), &cookie, remote)

	if err := parseRKNOnServer(t, serverDev, wire, remote); err == nil {
		t.Fatal("expected cookie verify to reject an RKN whose cookie is bound to a different source IP, but it passed")
	}
}
