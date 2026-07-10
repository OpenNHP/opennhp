package ac

import (
	"encoding/base64"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/OpenNHP/opennhp/nhp/core"
)

// AOP TransactionId dedupe (#1123).
//
// The per-connection LastRemoteSendTime check in core.responder.go
// rejects timestamp regressions only within a single open
// connection — server restart, AC failover, or NAT-table flush
// drops that state. Within the 600 s staleness floor a captured
// NHP_AOP can resend verbatim against a fresh connection and
// re-open ipset firewall entries. This cache closes that gap by
// remembering recently processed (sender_pubkey, txid, send_time)
// triples.
//
// Coverage limit — the cache is in-process, so an AC restart
// resets it. A captured AOP whose timestamp is still inside the
// staleness floor can therefore replay successfully against a
// freshly-started AC: both the responder gate and this cache are
// empty, so neither layer catches it. #1464 bounds this window by
// tightening the AOP-specific staleness floor in
// nhp/core/responder.go (recvStalenessFloor /
// AOPRecvStalenessFloorSeconds) from the historical 600 s default
// to 120 s — the staleness floor is the only gate that survives a
// cache wipe, so a tighter floor shrinks the replay window ~5×
// across every restart class (process restart, deploy, failover).
// Full elimination would need a persisted dedupe set or a shared
// instance-spanning store; that was evaluated under #1464 and
// judged disproportionate to the bounded threat (a replay only
// re-opens the ORIGINAL already-authorized (src,dst) ipset entry,
// and requires in-VPC packet capture) and to the priority:low
// severity. A signed monotonic-counter promise from the server
// remains a possible future hardening if the threat model
// escalates.
//
// Operational note — every process replacement creates a fresh cache. A
// captured AOP still inside the 120 s AOP staleness floor can therefore replay
// after restart; the staleness check bounds, but does not eliminate, that gap.
//
// CPU-amplification note — MarkSeen runs after AEAD decryption
// (the timestamp must be authenticated before it can be trusted
// as part of the dedupe key), so an attacker capable of replaying
// captured AOP packets can force the AC to pay decryption work
// for packets that are then dropped silently. The responder
// per-connection LastRemoteSendTime + flood gates bound this from
// a single connection; the cross-connection threat model that
// this cache exists to address by definition sidesteps those
// gates. At current human-paced cadence the amplification cost is
// immaterial — any attacker who can authenticate AEAD against the
// fresh-connection chain hash is already an authenticated
// compromise per the budget below.
//
// Key — (peerPubkey, txid, sendTime). The triple keys on the
// AEAD-authenticated send timestamp in addition to (pubkey, txid)
// for two reasons:
//
//   - Pubkey scoping is required because SenderTrxId is a
//     per-device monotonic counter; post-#676/#677 the AC connects
//     to multiple servers concurrently, so two distinct servers
//     can legitimately produce the same txid value in independent
//     counter spaces. Keying on txid alone would false-reject the
//     second server. The pubkey is already AEAD-authenticated by
//     validatePeer in core.responder.
//
//   - SendTime scoping closes a server-restart false-reject the
//     bare (pubkey, txid) shape would have. The server's
//     SenderTrxId is `atomic.AddUint64` over an in-memory counter
//     that resets on process restart. Deployments may also share a
//     keypair among multiple server instances. After a restart the
//     first post-restart AOPs would collide on
//     (pubkey, 1..K) with cached pre-restart entries and silently
//     drop for up to TTL (11 min). Including the per-packet send
//     timestamp distinguishes them: a captured-and-replayed packet
//     has byte-identical timestamp (the timestamp is in the AEAD
//     plaintext, so any tweak fails verification at the responder
//     before it ever reaches MarkSeen), while a fresh post-restart
//     AOP carries a fresh wall-clock timestamp. Replays still
//     collide on the full triple; legitimate post-restart traffic
//     does not. The dedupe semantics tighten from "same (pubkey,
//     txid)" to "same authenticated packet," which is what the
//     threat model actually wants. Per-instance pubkeys (#1467)
//     would be a stronger long-term fix.
//
// Legitimate server-side AOP retries get a fresh txid via
// `s.device.NextCounterIndex()` (see endpoints/server/udpserver.go
// where the AOP is constructed), so the cache will not
// false-reject a retry — only a byte-for-byte replay of the
// original packet collides on the full triple.
//
// The key is an opaque concatenation of the peer key bytes and the decimal
// transaction/timestamp fields. Curve peers use PublicKeySize bytes and GMSM
// peers use PublicKeySizeEx bytes; because lookup always reconstructs the key
// from the same authenticated tuple, no parsing or delimiter ambiguity is
// involved.
//
// TTL must cover the upstream AOP staleness floor in
// nhp/core/responder.go (`remoteSendTime < LocalInitTime -
// AOPRecvStalenessFloorSeconds`), otherwise a captured AOP aged
// between the cache TTL and the staleness floor would slip past
// both gates: the responder accepts it (timestamp not yet stale)
// and the cache no longer remembers it (entry expired). The floor
// was tightened to 120 s for AOP under #1464 (it was 600 s when
// this TTL was first sized), so the 11 min = 660 s TTL now
// over-covers the floor by ~5×; the "must cover" invariant holds
// with a wide margin. The TTL is deliberately left generous rather
// than shrunk to track the floor: the extra retention costs
// little at expected AOP cadence and preserves the post-restart
// counter-collision protection below, whose correctness depends on
// remembering pre-restart triples for the full TTL regardless of
// the floor. Deployments should synchronize clocks to stay within the
// floor's skew budget.
//
// Size (10 000 entries) is sized off the eviction floor that
// matters operationally: at 660 s TTL, the steady-state sustained
// AOP rate before LRU starts dropping live entries is
// 10_000 / 660 ≈ 15 AOP/sec. Deployments expecting sustained rates above
// that should increase the cap. An attacker cannot inflate the
// cache without first presenting a valid HMAC + AEAD-decrypting
// AOP, so the pre-pubkey-validation flood does not stress this
// budget — any flood that does is already an authenticated
// compromise. Eviction and duplicate-drop counters are not
// instrumented today; tracked in #1458.
//
// Lifecycle — expirable.LRU v2.0.7 has no Close()/Stop() hook for
// its TTL-sweep goroutine. In production the cache lives for the
// AC process lifetime so this is a non-issue; any future
// integration test that spins up/tears down UdpAC in one process,
// or a graceful-reload path, would leak one goroutine per restart
// until the library exposes a teardown hook.
const (
	aopReplayCacheSize = 10_000
	aopReplayCacheTTL  = 11 * time.Minute
	// pubkeyFingerprintLen is the truncation budget for the
	// log-line fingerprint emitted on duplicate-drops. 12 base64
	// chars ≈ 9 bytes (~72 bits) of pubkey entropy — enough to
	// distinguish one misbehaving server from a deployment-wide signal
	// in oncall logs without bloating the line or leaking the full
	// key. Birthday-bound collision probability for a typical
	// deployment of a few dozen servers is on the order of
	// 1e-19, so the truncation is collision-safe in operational
	// terms; if a deployment grows past ~10 000 servers the
	// budget should be revisited.
	pubkeyFingerprintLen = 12
)

// aopReplayCache is a bounded, TTL-expiring set of recently
// observed (sender_pubkey, txid, send_time) triples.
//
// recvMessageRoutine dispatches each NHP_AOP to a fresh goroutine
// (`go a.HandleUdpACOperations(ppd)` in udpac.go), so MarkSeen
// must be safe under concurrent invocation. expirable.LRU is
// internally thread-safe but the natural Contains-then-Add idiom
// has a TOCTOU window where two replays of the same packet both
// observe "not seen" and both pass; the mutex closes that.
type aopReplayCache struct {
	mu  sync.Mutex
	lru *expirable.LRU[string, struct{}]
}

func newAOPReplayCache() *aopReplayCache {
	return newAOPReplayCacheWithParams(aopReplayCacheSize, aopReplayCacheTTL)
}

// newAOPReplayCacheWithParams builds a cache with caller-supplied
// size and TTL. Production wires the constants via
// newAOPReplayCache; tests pass short values so TTL-eviction paths
// run without 5-minute waits.
func newAOPReplayCacheWithParams(size int, ttl time.Duration) *aopReplayCache {
	return &aopReplayCache{
		lru: expirable.NewLRU[string, struct{}](size, nil, ttl),
	}
}

// MarkSeen records the (peerPubkey, txid, sendTime) triple and
// returns true on first observation, false if the triple is
// already in the cache.
//
// Caller contract — peerPubkey and sendTime must both be the
// AEAD-authenticated values from PacketParserData
// (ppd.RemotePubKey and ppd.RemoteSendTime). The handler-level
// `len(ppd.RemotePubKey) == 0` guard in HandleUdpACOperations is
// the canonical place to reject upstream-invariant violations
// (validatePeer should always populate RemotePubKey); MarkSeen
// returning false on an empty pubkey is fail-closed insurance, not
// the primary error surface — a future caller that skips the
// upstream guard will still not silently key every empty-pubkey
// AOP under the same `":<txid>:<ts>"` slot.
//
// Curve and GMSM use different authenticated public-key lengths. Any other
// length fails closed.
func (c *aopReplayCache) MarkSeen(peerPubkey []byte, txid uint64, sendTime int64) bool {
	if len(peerPubkey) != core.PublicKeySize && len(peerPubkey) != core.PublicKeySizeEx {
		return false
	}

	key := aopReplayKey(peerPubkey, txid, sendTime)

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.lru.Get(key); exists {
		return false
	}
	c.lru.Add(key, struct{}{})
	return true
}

// Len is the current entry count, exposed for tests. Takes the
// same mutex as MarkSeen so a Len() call never observes the cache
// between MarkSeen's Get and Add — necessary for deterministic
// `Len() == cap` assertions ordered after a MarkSeen from the
// same goroutine.
func (c *aopReplayCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.Len()
}

// aopReplayKey concatenates pubkey bytes with separators, the
// decimal txid, and the decimal sendTime. Tracked in #1460 —
// switching to a fixed-size byte array key is zero-alloc but
// unneeded at current AOP cadence. Per-call cost is ~3 short
// allocations (`string(peerPubkey)` ~32 B, the
// FormatUint/FormatInt buffers ~20 B each, the final concat
// builder ~80 B) — negligible at <100 AOP/sec, becomes meaningful
// at ~10k AOP/sec where the GC pressure starts dominating the
// per-packet cost.
func aopReplayKey(peerPubkey []byte, txid uint64, sendTime int64) string {
	return string(peerPubkey) + ":" + strconv.FormatUint(txid, 10) + ":" + strconv.FormatInt(sendTime, 10)
}

// pubkeyFingerprint returns a short, log-friendly representation of
// a public key for breadcrumb correlation, URL-safe-base64-encoded
// and truncated to pubkeyFingerprintLen chars. URL-safe encoding
// (-, _) avoids `/` collisions in log indexers that treat it as a
// path separator (Loki, CloudWatch Insights). Empty input returns
// the literal "empty" so a missing-pubkey breadcrumb is still
// grep-able.
func pubkeyFingerprint(peerPubkey []byte) string {
	if len(peerPubkey) == 0 {
		return "empty"
	}
	enc := base64.RawURLEncoding.EncodeToString(peerPubkey)
	if len(enc) > pubkeyFingerprintLen {
		return enc[:pubkeyFingerprintLen]
	}
	return enc
}
