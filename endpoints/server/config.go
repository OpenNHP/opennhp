package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/OpenNHP/opennhp/nhp/etcd"

	ebpflocal "github.com/OpenNHP/opennhp/endpoints/server/ebpf"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/core/verifier"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/plugins"
	"github.com/OpenNHP/opennhp/nhp/utils"
	utilsebpf "github.com/OpenNHP/opennhp/nhp/utils/ebpf"

	toml "github.com/pelletier/go-toml/v2"
)

// shippedDemoCookieSigningKeyBase64 is the value committed in
// docker/nhp-server/etc/config.toml so that `docker-compose up` works
// out of the box. udpserver.Start compares the configured key to this
// constant and logs a Critical line if they match, so operators who
// copy the demo and forget to rotate the key get a loud warning
// instead of silently running with a public secret.
//
// Keep this in sync with docker/nhp-server/etc/config.toml (and
// docker/nhp-server/etc2/config.toml, which intentionally shares the
// same value to enable the same-key multi-instance demo). If we ever
// rotate the demo key, update this constant in the same commit.
const shippedDemoCookieSigningKeyBase64 = "w62S2G1P5GOG66Y5tIv3WlfBv8CNBdDe2JJDFr9Q+h0="

// decodeCookieSigningKey parses a base64-encoded 32-byte cookie signing
// key. An empty input yields (nil, nil): the caller will fall back to a
// random per-process key, which is fine for single-instance deployments.
func decodeCookieSigningKey(b64 string) ([]byte, error) {
	if b64 == "" {
		return nil, nil
	}
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}
	if len(raw) != 32 {
		return nil, fmt.Errorf("cookie signing key must be exactly 32 bytes after base64 decode, got %d", len(raw))
	}
	return raw, nil
}

var (
	baseConfigWatch  io.Closer
	httpConfigWatch  io.Closer
	acConfigWatch    io.Closer
	agentConfigWatch io.Closer
	resConfigWatch   io.Closer
	srcipConfigWatch io.Closer
	dbConfigWatch    io.Closer
	relayConfigWatch io.Closer
	xdpConfigWatch   io.Closer
	teeWatch         io.Closer
	errLoadConfig    = fmt.Errorf("config load error")
)

type ServerEtcdConfig struct {
	BaseConfig    Config
	HttpConfig    HttpConfig
	ACs           []*core.UdpPeer
	Agents        []*core.UdpPeer
	DBs           []*core.UdpPeer
	AuthServiceId []*common.AuthServiceProviderData
	SrcIps        []*SrcIpMap
}

type SrcIpMap struct {
	SrcIp string
	Ip    []string
}

type Config struct {
	ARTReplayCacheEntries  int    `json:"artReplayCacheEntries"` // Restart-only; zero selects 100000.
	PrivateKeyBase64       string `json:"privateKey"`
	Hostname               string `json:"hostname"`
	ListenIp               string `json:"listenIp"`
	ListenPort             int    `json:"listenPort"`
	LogLevel               int    `json:"logLevel"`
	DefaultCipherScheme    int    `json:"defaultCipherScheme"`
	DisableAgentValidation bool   `json:"disableAgentValidation"`

	// AllowPrivateRelaySource relaxes the SourceAddr public-routability check
	// that HandleRelayForward applies to inner KNK packets arriving via a
	// relay. When false (production default), private / loopback / CGNAT
	// addresses in RelayForwardMsg.SourceAddr are rejected as fabricated;
	// the threat model assumes a relay peer might be compromised and trying
	// to fill the server's connectionMap with synthetic clients.
	//
	// Set true ONLY in environments where the relay legitimately sees
	// non-public client addresses, such as the bundled docker-compose demo:
	// when a host-side browser hits the relay through Docker Desktop's port
	// mapping, the relay container sees the request as coming from the
	// vpnkit gateway (192.168.65.1 on macOS / 172.x.x.x on Linux), which is
	// RFC1918. A production-facing relay should NEVER see such an address;
	// flipping this on there would let a misbehaving relay inject any
	// private-range SourceAddr it wants into the server's connection map
	// and the downstream AC ipset whitelist.
	AllowPrivateRelaySource bool `json:"allowPrivateRelaySource"`

	// CookieSigningKeyBase64 is a base64-encoded 32-byte HMAC key used to
	// derive overload-mode cookies statelessly from (remoteAddr || time
	// window). When multiple nhp-server instances sit behind a load
	// balancer / round-robin DNS, an agent's KNK and its follow-up RKN may
	// land on different instances; with per-instance random cookie state
	// (the legacy CookieStore), the second instance can't validate the
	// cookie issued by the first, and the handshake stalls. Sharing this
	// signing key across every instance in the cluster lets any of them
	// independently mint and verify the same cookie.
	//
	// Format: base64 of exactly 32 bytes. If empty, the server generates a
	// random per-process key at startup — fine for single-instance
	// deployments, broken for multi-instance ones (the failure mode is the
	// same as the legacy CookieStore: cookies don't cross instances).
	// Generate one with:   head -c 32 /dev/urandom | base64
	CookieSigningKeyBase64 string `json:"cookieSigningKey"`

	// CookieTimeWindowSeconds is the rolling time window used in cookie
	// derivation. The current and previous window are both accepted on
	// verify, so an agent has between [window, 2*window] seconds to use a
	// cookie before it expires. Default 60s if unset / non-positive.
	CookieTimeWindowSeconds int `json:"cookieTimeWindowSeconds"`

	// DatabasePath is the filesystem path to the SQLite database used for
	// agent key registration and OTP storage. If empty, defaults to
	// "<exe_dir>/data/nhp_server.db".
	DatabasePath string `json:"databasePath"`

	// ForceOverload pins the device's Overload flag to true at startup,
	// short-circuiting the connection-count-driven trigger. The normal
	// trigger fires when remoteConnectionMap crosses
	// OverloadConnectionThreshold (~16k concurrent connections), which a
	// local demo will never reach — so this flag exists purely to let
	// developers exercise the cookie path (KNK → NHP_COK → NHP_RKN) on a
	// quiet local stack.
	//
	// Default: false. Production deployments must leave this off; pinning
	// Overload on permanently forces every agent through the slower
	// cookie-stamped handshake even when the server is idle, and tells
	// the cookie store that load is constantly elevated.
	//
	// This is a debug/test affordance, NOT a feature flag. Do not key
	// production behavior off it.
	//
	// Hot-reload caveat: flipping this back to false at runtime via a
	// config reload does NOT immediately restore normal behavior. The
	// connection-teardown path in udpserver.go honors ForceOverload to
	// keep Overload pinned across connection churn, so as long as the
	// process keeps observing ForceOverload=true it will never call
	// SetOverload(false) — the flag is sticky for the lifetime of the
	// process. Restart the server to clear it.
	ForceOverload bool `json:"forceOverload"`

	// OTPTTLSeconds is the lifetime of a one-time password for agent
	// registration, in seconds. Default 300 (5 minutes) if unset / zero.
	OTPTTLSeconds int `json:"otpTTLSeconds"`

	// AgentKeyTTLSeconds is the lifetime of a registered agent public
	// key, in seconds. After this elapses, the noise-layer peer
	// validation will reject the key and knocks will fail as if the
	// agent were never registered. Default 86400 (24h) if unset / zero.
	AgentKeyTTLSeconds int `json:"agentKeyTTLSeconds"`

	// Audit configures the tamper-evident security audit ledger. Disabled
	// by default; see AuditConfig.
	Audit AuditConfig `json:"audit"`

	// Metrics controls the optional Prometheus metrics / health endpoint.
	// Disabled by default; when enabled it binds a separate, local-by-default
	// HTTP listener so operational telemetry never rides on the public knock
	// surface.
	Metrics MetricsConfig `json:"metrics"`

	// XdpConfigPath points at the eBPF/XDP ingress policy file, relative to
	// the exe directory (or absolute). Empty uses "etc/xdp.toml". The file is
	// optional: without it the XDP filter keeps whatever whitelist it was
	// loaded with, which on a fresh start is none.
	XdpConfigPath string `json:"xdpConfigPath"`

	// AttestationScheme selects the TEE evidence format this server accepts
	// on the DHP knock path. "csv" (default, and the result of any
	// unrecognised value) requires a real Hygon CSV attestation report and
	// runs the full Hygon certificate-chain check. "test" accepts
	// self-asserted evidence with NO cryptographic assurance — it exists
	// only for the non-TEE container walkthrough in docs/dhp_quick_start.md
	// and must never be set on anything reachable by an untrusted agent.
	//
	// This is deliberately a server-side choice: before it existed, the
	// verifier was picked from a "test_purpose" key inside the evidence
	// itself, so any agent could opt the server into the no-op verifier.
	AttestationScheme string `json:"attestationScheme"`
}

// XdpTomlConfig is etc/xdp.toml, the ingress policy the XDP program enforces.
//
// The presence of this file is what opts a host into being filtered at all —
// see (*UdpServer).loadXdpConfig. Every field reaches the kernel: Enabled and
// RelayIPs decide whether the program is attached, and NhpMinFrameBytes is
// written into its .rodata at load time. Enabled and NhpMinFrameBytes are read
// once, at startup; only RelayIPs is hot-reloadable.
type XdpTomlConfig struct {
	// Enabled attaches the filter. False (or a missing file) leaves the host
	// with the exposure it had before the filter existed.
	Enabled bool
	// RelayIPs are the sources allowed to reach SSH, and allowed to reach the
	// NHP port without meeting the length floor. Entries are host addresses
	// ("10.0.1.4") or prefixes ("10.0.1.0/24"). Empty refuses to attach: it
	// would close tcp/22 for every source.
	RelayIPs []string
	// NhpMinFrameBytes is the UDP length floor on the knock port. Zero uses
	// DefaultNhpMinFrameBytes.
	NhpMinFrameBytes int
}

// DefaultNhpMinFrameBytes mirrors the nhp_min_udp_len default in
// nhp/ebpf/xdp/nhp_server_xdp.c: the 240-byte NHP_KPL header of the curve
// cipher suite, the shortest datagram that can be a knock.
const DefaultNhpMinFrameBytes = 240

// MetricsConfig configures the observability endpoint exposed by nhp-server.
//
// TODO: this duplicates nhp/metrics.Config (same three fields, same
// defaults) and endpoints/server/metricsserver.go duplicates
// nhp/metrics.Endpoint. nhp-ac/relay/db already use the shared types;
// migrating nhp-server to metrics.StartEndpoint would delete ~110 lines and
// leave one implementation. Kept separate here only to bound this PR's churn.
type MetricsConfig struct {
	// Enabled turns the /metrics + /healthz listener on. Off by default.
	Enabled bool `json:"enabled"`
	// ListenIp is the bind address. Empty defaults to 127.0.0.1 so metrics are
	// not exposed off-host unless the operator explicitly opts in.
	ListenIp string `json:"listenIp"`
	// ListenPort is the TCP port for the endpoint. Empty/zero defaults to 9100.
	ListenPort int `json:"listenPort"`
}

// AuditConfig controls the hash-chained security audit ledger. When
// enabled, security-relevant decisions (knock granted/denied over UDP and
// HTTP, agent registered) are appended as JSON lines linked into a hash
// chain that makes after-the-fact tampering detectable. When enabled it is
// the server's structured audit trail — the nhp/log "[Audit]" stream is an
// unused API, so this does not duplicate an existing log.
type AuditConfig struct {
	// Enabled turns the ledger on. Off by default.
	Enabled bool `json:"enabled"`
	// FilePath is where the ledger is written. Relative paths resolve
	// against <exe_dir>. Defaults to "<exe_dir>/audit/audit-ledger.jsonl"
	// (its own directory, not logs/, so a logrotate rule aimed at logs/ can't
	// reach it).
	//
	// Do NOT point an external log-rotation tool at this file while the
	// server is running. The ledger holds one append handle for the process
	// lifetime and does not reopen: a rename+create rotation sends every
	// later entry to the rotated-away inode (auditing silently stops until
	// restart), and a copytruncate resets the file to offset 0 while the
	// in-memory seq/hash keep advancing, so `audit verify` then reports a
	// chain break — the wording for tampering — for a routine cron job.
	// Rotate only while the server is stopped, or archive whole segments
	// out of band and let this file keep growing.
	//
	// Do NOT point two processes at the same FilePath either. Each writer
	// keeps its own in-memory seq/lastHash, so interleaved appends produce
	// duplicate seq values and broken prevHash links that verify reports as
	// tampering; worse, one process's torn-tail repair (an O_RDWR truncate
	// back to the last newline) can delete an entry the other just committed.
	// There is no cross-process lock, so give each instance its own ledger —
	// this matters for the bundled server/server2 two-cluster demo.
	FilePath string `json:"filePath"`
	// Fsync flushes each entry to disk before returning. Note that entries
	// are written synchronously on the request path, so audit volume tracks
	// knock volume: with this on, every access decision costs a disk flush
	// and all audit writes serialize behind one mutex. Worth the durability
	// on a normal gateway, but turn it off if the ledger becomes a
	// bottleneck under load — the hash chain stays intact either way.
	Fsync bool `json:"fsync"`
	// SigningKeyBase64 is an optional base64 HMAC key (at least 32 bytes
	// after decoding; a shorter value is rejected at startup). When set,
	// each entry is additionally signed so the chain is bound to a secret
	// the log file does not contain. Without it, the hash chain alone still
	// detects local edits, deletions and reordering.
	//
	// Be precise about what the signature buys you. The key lives in this
	// config file, so an attacker who has taken over the server process
	// (running as its uid, or root) can read the key, rewrite the ledger,
	// and re-sign it — the signature does NOT survive a full host
	// compromise. What it does defend against is an attacker who can write
	// the log but not read this config: a compromised log-shipping account,
	// or offline tampering with an archived copy. For protection against a
	// host compromise the key has to live off the host (an append-only sink
	// the server can write but not rewrite).
	//
	// One attack is undetectable from the file alone even with a key:
	// truncating entries off the END leaves a shorter chain that still
	// verifies, because a hash chain cannot prove it was not shortened.
	// Detecting rollback needs an external anchor — periodically record the
	// latest seq+hash off-host and compare against it.
	//
	// Setting or rotating this on an EXISTING ledger is fine: Open resumes
	// the same file, so it ends up with an unsigned prefix (entries logged
	// before the key existed) and a signed suffix. `audit verify --key`
	// treats that prefix as UnsignedEntries, not a signature mismatch — it
	// does not read as tampering.
	SigningKeyBase64 string `json:"signingKey"`

	// FailClosed controls what happens when the ledger cannot be opened at
	// startup (a corrupt or foreign file at FilePath, a permission problem).
	//
	// Default (false): fail SAFE — the gateway keeps producing a trail. The
	// unreadable file at FilePath is handled one of two ways:
	//   - It still looks like one of our ledgers (a corrupted first line, an
	//     attacker prepending junk): it is renamed to "<FilePath>.corrupt-
	//     <nanos>" and a fresh chain starts at FilePath — from seq 1, or, if
	//     numbered "<FilePath>.<n>" segments from size rotation are present,
	//     continuing from the highest one. Either way the .corrupt-* sibling
	//     next to a re-created live file is the loud, detectable signal.
	//   - It is a FOREIGN file (a mistyped FilePath pointing at another log,
	//     a config, a shared-volume file): it is LEFT UNTOUCHED — a
	//     privileged server must not move an operator's unrelated file — and
	//     auditing continues in a fixed sibling, "<FilePath>.quarantined.jsonl".
	//     Note `audit verify <FilePath>` then verifies the foreign file, not
	//     the quarantined ledger; point it at the .quarantined.jsonl path.
	// The rename/sibling guards accidents and casual edits, not someone with
	// write access to the directory, who can delete the file too.
	//
	// true: fail CLOSED. Any open failure aborts startup instead. Choose this
	// when a verifiable, uninterrupted trail is a hard requirement and you
	// would rather the gateway not serve at all than serve unaudited. The
	// offending file is left untouched for inspection.
	//
	// Either way, this only governs STARTUP. A write that fails while the
	// server is already running (disk full, the file made unwritable) is
	// logged but never blocks the request, so access decisions keep flowing
	// while the trail is blind. A sustained run of such failures escalates to
	// a rate-limited Critical (see auditEvent) so it cannot pass unnoticed,
	// but there is deliberately no runtime fail-closed: an audit hiccup must
	// not take the gateway down mid-flight.
	FailClosed bool `json:"failClosed"`

	// Async moves the disk write (and fsync) for each entry off the request
	// goroutine onto a single background writer. Log still computes seq and
	// the hash chain under the lock, so ordering and linkage are unchanged;
	// only the write is deferred. This keeps Fsync usable on a busy gateway.
	// If the writer falls far enough behind that the queue fills, entries are
	// DROPPED rather than blocking the knock — discarded whole so the chain
	// stays contiguous, counted in the shutdown summary and a runtime
	// Critical, and once the writer recovers the next Log chains an
	// "audit_gap" marker (fields.dropped=N) so a verified copy of the ledger
	// still shows the loss. Under Async, chain integrity holds but
	// completeness does not. Off by default.
	Async bool `json:"async"`
	// AsyncQueueSize bounds the pending-write queue when Async is set.
	// 0 uses a sensible default; a value over ~1M is rejected at startup
	// (it would allocate gigabytes). Ignored unless Async.
	AsyncQueueSize int `json:"asyncQueueSize"`

	// MaxSizeBytes controls size-based rotation: once the live file would
	// grow past it, the ledger is renamed to a numbered segment
	// ("<FilePath>.<seq>") and a fresh file continues the chain. The chain
	// spans the segments and `audit verify` picks the siblings up
	// automatically. Config-file semantics:
	//   0        - use the built-in default (256 MiB per segment).
	//   negative - never rotate; one file that grows without bound.
	//   positive - rotate at that many bytes.
	// Rotation on its own deletes nothing; see MaxSegments for retention.
	MaxSizeBytes int64 `json:"maxSizeBytes"`

	// MaxSegments is the retention cap on rotated "<FilePath>.<n>" files.
	// DELETING audit records is opt-in — config-file semantics:
	//   0 (default) - keep every segment forever. Nothing is ever deleted.
	//   negative    - same as 0 (keep everything).
	//   positive    - after a rotation, delete the oldest segments beyond
	//                 this count, bounding disk use at ~MaxSizeBytes*(N+1).
	//                 Each deletion is logged Critical, because it drops
	//                 evidence and `audit verify` can then no longer walk
	//                 from seq 1 (it anchors on the first surviving entry).
	// NHP_OTP / NHP_REG are audited before the peer is validated, so a party
	// that knows the server's public key can drive ledger volume; the answer
	// is off-box archival plus a disk-pressure alarm, not silent deletion —
	// hence the conservative default.
	MaxSegments int `json:"maxSegments"`
}

type RemoteConfig struct {
	Provider  string
	Key       string
	Endpoints []string
	Username  string
	Password  string
}

type HttpConfig struct {
	EnableHttp     bool
	EnableTLS      bool
	HttpListenIp   string
	HttpListenPort int
	TLSCertFile    string
	TLSKeyFile     string
	ReadTimeoutMs  int
	WriteTimeoutMs int
	IdleTimeoutMs  int
}

type Peers struct {
	ACs    []*core.UdpPeer
	Agents []*core.UdpPeer
	DBs    []*core.UdpPeer
	Relays []*core.UdpPeer
}

func (s *UdpServer) loadBaseConfig() error {
	// config.toml
	fileName := filepath.Join(ExeDirPath, "etc", "config.toml")
	content, err := s.loadConfigFile(fileName)
	if err != nil {
		log.Error("load base config err: %v", err)
		return err
	}
	// pelletier/go-toml/v2 silently drops unknown sections, so a stale
	// [webrtc] block in an upgraded config.toml would otherwise produce
	// no signal that the transport is gone. Warn loudly once at load.
	if strings.Contains(string(content), "[webrtc]") {
		log.Warning("[loadBaseConfig] [webrtc] section in config.toml is ignored: " +
			"the WebRTC transport was removed; delete the section to silence this warning")
	}

	var config Config
	if unmarshalErr := toml.Unmarshal(content, &config); unmarshalErr != nil {
		log.Error("failed to unmarshal base config: %v", unmarshalErr)
	}
	if err = s.updateBaseConfig(config); err != nil {
		// report base config error
		return err
	}

	baseConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("base config: %s has been updated", fileName)
		if content, err = s.loadConfigFile(fileName); err == nil {
			if err = toml.Unmarshal(content, &config); err == nil {
				_ = s.updateBaseConfig(config)
			}

		}
	})
	return nil
}

func (s *UdpServer) loadHttpConfig() error {
	// http.toml
	fileName := filepath.Join(ExeDirPath, "etc", "http.toml")
	content, err := s.loadConfigFile(fileName)
	if err != nil {
		log.Error("load http config err: %v", err)
		return err
	}
	var httpConf HttpConfig
	if unmarshalErr := toml.Unmarshal(content, &httpConf); unmarshalErr != nil {
		log.Error("failed to unmarshal http config: %v", unmarshalErr)
	}
	if err = s.updateHttpConfig(httpConf); err != nil {
		// ignore error
		_ = err
	}

	httpConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("http config: %s has been updated", fileName)
		if content, err = s.loadConfigFile(fileName); err == nil {
			if err = toml.Unmarshal(content, &httpConf); err == nil {
				_ = s.updateHttpConfig(httpConf)
			}
		}

	})
	return nil
}

func (s *UdpServer) loadPeers() error {
	// ac.toml
	fileNameAC := filepath.Join(ExeDirPath, "etc", "ac.toml")

	contentAC, err := s.loadConfigFile(fileNameAC)
	if err != nil {
		log.Error("load ac peer config err: %v", err)
		return err
	}
	var acPeers Peers
	if unmarshalErr := toml.Unmarshal(contentAC, &acPeers); unmarshalErr != nil {
		log.Error("failed to unmarshal ac peers config: %v", unmarshalErr)
	}

	if updateErr := s.updateACPeers(acPeers.ACs); updateErr != nil {
		// ignore error
		_ = updateErr
	}

	acConfigWatch = utils.WatchFile(fileNameAC, func() {
		log.Info("ac peer config: %s has been updated", fileNameAC)
		if contentAC, err = s.loadConfigFile(fileNameAC); err == nil {
			if err = toml.Unmarshal(contentAC, &acPeers); err == nil {
				_ = s.updateACPeers(acPeers.ACs)
			}
		}
	})

	// agent.toml
	fileNameAgent := filepath.Join(ExeDirPath, "etc", "agent.toml")
	contentAgent, err := s.loadConfigFile(fileNameAgent)
	if err != nil {
		log.Error("load agent peer config err: %v", err)
		return err
	}
	var agentPeers Peers
	if unmarshalErr := toml.Unmarshal(contentAgent, &agentPeers); unmarshalErr != nil {
		log.Error("failed to unmarshal agent peers config: %v", unmarshalErr)
	}
	if updateErr := s.updateAgentPeers(agentPeers.Agents); updateErr != nil {
		// ignore error
		_ = updateErr
	}

	agentConfigWatch = utils.WatchFile(fileNameAgent, func() {
		log.Info("agent peer config: %s has been updated", fileNameAgent)
		if contentAgent, err = s.loadConfigFile(fileNameAgent); err == nil {
			if err = toml.Unmarshal(contentAgent, &agentPeers); err == nil {
				_ = s.updateAgentPeers(agentPeers.Agents)
			}
		}
	})

	//db.toml (optional)
	fileNameDE := filepath.Join(ExeDirPath, "etc", "db.toml")
	contentDE, err := s.loadConfigFile(fileNameDE)
	if err != nil {
		log.Warning("load db peer config err (optional): %v", err)
	} else {
		var dePeers Peers
		if unmarshalErr := toml.Unmarshal(contentDE, &dePeers); unmarshalErr != nil {
			log.Error("failed to unmarshal db peers config: %v", unmarshalErr)
		}
		if updateErr := s.updateDePeers(dePeers.DBs); updateErr != nil {
			// ignore error
			_ = updateErr
		}
		dbConfigWatch = utils.WatchFile(fileNameDE, func() {
			log.Info("device peer config: %s has been updated", fileNameDE)
			if contentDE, err = s.loadConfigFile(fileNameDE); err == nil {
				if err = toml.Unmarshal(contentDE, &dePeers); err == nil {
					_ = s.updateDePeers(dePeers.DBs)
				}
			}
		})
	}

	// relay.toml (optional)
	fileNameRelay := filepath.Join(ExeDirPath, "etc", "relay.toml")
	contentRelay, err := s.loadConfigFile(fileNameRelay)
	if err != nil {
		log.Warning("load relay peer config err (optional): %v", err)
	} else {
		var relayPeers Peers
		if unmarshalErr := toml.Unmarshal(contentRelay, &relayPeers); unmarshalErr != nil {
			log.Error("failed to unmarshal relay peers config: %v", unmarshalErr)
		}
		if updateErr := s.updateRelayPeers(relayPeers.Relays); updateErr != nil {
			_ = updateErr
		}
		relayConfigWatch = utils.WatchFile(fileNameRelay, func() {
			log.Info("relay peer config: %s has been updated", fileNameRelay)
			if contentRelay, err = s.loadConfigFile(fileNameRelay); err == nil {
				var relayPeers Peers
				if err = toml.Unmarshal(contentRelay, &relayPeers); err == nil {
					_ = s.updateRelayPeers(relayPeers.Relays)
				}
			}
		})
	}

	// tee.toml
	fileNameTee := filepath.Join(ExeDirPath, "etc", "tee.toml")
	if err := s.updateTee(fileNameTee); err != nil {
		// ignore error
		_ = err
	}
	teeWatch = utils.WatchFile(fileNameTee, func() {
		log.Info("tee: %s has been updated", fileNameTee)
		_ = s.updateTee(fileNameTee)
	})

	return nil
}

// xdpConfigFileName resolves the configured xdp.toml path against the exe
// directory. An absolute XdpConfigPath is honored as-is.
func (s *UdpServer) xdpConfigFileName() string {
	path := "etc/xdp.toml"
	if s.config != nil && s.config.XdpConfigPath != "" {
		path = s.config.XdpConfigPath
	}
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(ExeDirPath, path)
}

// dropXdpLoaderPrivileges is ebpflocal.DropLoaderPrivileges behind a variable,
// so a test can assert what no comment can: that every return in loadXdpConfig
// reaches it. Missing one is the whole bug this indirection guards — the drop
// used to live only inside the loader, which the refusal paths never call, so a
// host that declined to filter kept CAP_BPF, CAP_NET_ADMIN and CAP_PERFMON for
// the life of the process. Never reassigned outside tests.
var dropXdpLoaderPrivileges = ebpflocal.DropLoaderPrivileges

// loadXdpConfig reads etc/xdp.toml, attaches the XDP ingress filter if the file
// asks for one, and keeps watching the file for whitelist changes.
//
// **The file is what opts a host in.** The filter is not a hardening tweak: it
// drops everything at the driver except UDP on the knock port and SSH from the
// listed sources, which on a host with no other way in is one bad whitelist away
// from being unreachable for good. So a server that has not been configured for
// it — no xdp.toml, an unparsable one, Enabled = false, or an empty RelayIPs —
// filters nothing and keeps the ingress exposure it had before the filter
// existed. Only a file that says all three things (present, Enabled, non-empty
// whitelist) attaches anything. The one thing such a host does not keep is the
// loader's own three capabilities (see the defer below), which it was granted
// for a load that never happened.
//
// If the file is missing or unparsable on a *reload*, nothing is applied and the
// kernel keeps the whitelist it already had. Overwriting a working whitelist
// with the contents of a truncated or half-written file is the failure that
// locks the operator out of SSH, so a reload only ever takes effect once
// toml.Unmarshal has succeeded.
func (s *UdpServer) loadXdpConfig(logLevel int) error {
	fileName := s.xdpConfigFileName()

	// Give the loader's capabilities back on the way out, whatever is decided
	// below. The unit grants CAP_BPF, CAP_NET_ADMIN and CAP_PERFMON as ambient
	// capabilities before this function gets a say, so the daemon is holding
	// them right now on every host — including the ones that are about to
	// decline to filter. Those are the states that hold them longest and use
	// them least: no object loaded, no filter to detach, no whitelist map to
	// reload, and the same untrusted UDP and dlopen'd plugins in the address
	// space for the life of the process.
	//
	// The loader drops on its own way out too, and gets there first on the
	// attach path, which is what decides whether CAP_BPF survives for the
	// reload path (ebpflocal.DropLoaderPrivileges is once-guarded, first call
	// wins). So `attached` only ever decides the question on the paths where
	// the loader never ran — and there the answer is "keep none of the three".
	//
	// Those three and nothing else: the drop subtracts CAP_BPF, CAP_NET_ADMIN
	// and CAP_PERFMON from whatever the process holds rather than resetting it,
	// so a deployment that runs nhp-serverd as root (docker/Dockerfile.server)
	// keeps the unrelated capabilities it was started with. A host with no
	// xdp.toml still runs as it did before this filter existed, except for
	// three capabilities it had no way to use.
	attached := false
	defer func() { dropXdpLoaderPrivileges(attached) }()

	content, err := s.loadConfigFile(fileName)
	if err != nil {
		log.Info("no %s: the XDP ingress filter is not attached (%v)", fileName, err)
		return err
	}

	var xdpConf XdpTomlConfig
	if unmarshalErr := toml.Unmarshal(content, &xdpConf); unmarshalErr != nil {
		log.Error("failed to unmarshal xdp config, the XDP ingress filter is not attached: %v", unmarshalErr)
		return unmarshalErr
	}

	attached = s.startXdpFilter(&xdpConf, logLevel)
	if !attached {
		// Not attached: there is no map to mirror the file into, and no point
		// watching a file whose only effect is on a filter that is not running.
		// A later edit takes effect on the next restart, which is also when the
		// operator gets to see the decision logged again.
		return nil
	}

	// No apply here: startXdpFilter returning true already means the kernel map
	// holds this file's whitelist, because the loader writes it before it
	// attaches the program. Re-applying it would be a second chance to fail at
	// the one moment a failure cannot be acted on — the filter is live, so
	// "keeping the active whitelist" would mean keeping whatever that write
	// left behind, with tcp/22 the thing at stake. From here the map only ever
	// changes through a reload, which has a working whitelist to fall back on.
	xdpConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("xdp config: %s has been updated", fileName)
		content, err := s.loadConfigFile(fileName)
		if err != nil {
			log.Error("failed to reread xdp config, keeping the active whitelist: %v", err)
			return
		}
		var xdpConf XdpTomlConfig
		if err := toml.Unmarshal(content, &xdpConf); err != nil {
			log.Error("failed to unmarshal xdp config, keeping the active whitelist: %v", err)
			return
		}
		s.applyXdpConfig(&xdpConf)
	})

	return nil
}

// startXdpFilter decides whether to attach the ingress filter and, if so,
// attaches it. It reports whether the filter is now running.
//
// Every refusal below leaves the host exactly as it was without the filter,
// which is a state the daemon has always supported. The failure this guards
// against is the opposite one — attaching a filter whose policy does not match
// the daemon it is protecting — because that drops the host's own traffic with
// nothing in user space to say so, and on a host reachable only through the
// whitelist it cannot be undone remotely.
func (s *UdpServer) startXdpFilter(conf *XdpTomlConfig, logLevel int) bool {
	fileName := s.xdpConfigFileName()

	if !conf.Enabled {
		log.Info("xdp config: %s has Enabled = false; the XDP ingress filter is not attached", fileName)
		return false
	}
	// A list that names no usable prefix is never a policy anyone wants: the
	// whitelist is the only thing that reaches tcp/22 on a filtered host, so
	// attaching with one closes SSH for everybody with no break-glass path. It
	// is, on the other hand, exactly what a config rendered with an unset
	// RELAY_IPS produces, and what a truncated or half-written file parses to
	// (TOML with no RelayIPs key is valid TOML).
	//
	// "No usable prefix" is decided by parsing, not by counting strings: a
	// hostname, an IPv6 address or a typo like 10.0.1.300 all leave a list that
	// looks populated here and reaches the kernel empty.
	if _, ok := xdpRelayPrefixes(conf, fileName, "refusing to attach the XDP ingress filter rather than closing tcp/22 for every source"); !ok {
		return false
	}

	minBytes := conf.NhpMinFrameBytes
	if minBytes == 0 {
		minBytes = DefaultNhpMinFrameBytes
	}
	if minBytes < 0 || minBytes > 65535 {
		log.Error("xdp config: NhpMinFrameBytes=%d is out of range; refusing to attach the XDP ingress filter", minBytes)
		return false
	}

	listenPort := s.GetListenPort()
	if listenPort <= 0 || listenPort > 65535 {
		log.Error("xdp config: cannot filter for UDP listen port %d; refusing to attach the XDP ingress filter", listenPort)
		return false
	}

	// The filter admits UDP on the knock port and SSH from the whitelist, and
	// drops every other packet that would start a new flow — including a SYN to
	// a TCP port this daemon is listening on. Attaching in front of a listener
	// would make that service look dead from everywhere, so refuse instead and
	// say which listener is in the way. An operator who wants both needs a
	// filter that knows about the service ports, which this one deliberately
	// does not (see nhp/ebpf/xdp/nhp_server_xdp.c).
	if reason := s.xdpServiceConflict(); reason != "" {
		log.Critical("xdp config: refusing to attach the XDP ingress filter because %s — the filter drops inbound TCP other than SSH from RelayIPs, so that service would be unreachable. Disable it, bind it to 127.0.0.1, or set Enabled = false in %s",
			reason, fileName)
		return false
	}

	// Fail-open past this point: a kernel too old for XDP, a missing CAP_BPF, an
	// unreadable object file or an unmounted bpffs must not take the gateway off
	// the air, so a failure is logged and the daemon runs with no ingress
	// filter — the same exposure it had before this existed.
	var serverId string
	if s.config != nil {
		serverId = s.config.Hostname
	}
	// RelayIPs travels with the load rather than being written afterwards: the
	// loader installs it before it attaches anything, so the filter is never
	// live holding an empty whitelist — not for the window between attaching
	// and the first map write, and not permanently because that write failed.
	// Either the filter is on and enforcing this list, or it is not on.
	if err := ebpflocal.EngineLoad(ebpflocal.LoadParams{
		DirPath:       ExeDirPath,
		LogLevel:      logLevel,
		ServerId:      serverId,
		NhpPort:       uint16(listenPort),
		MinFrameBytes: uint16(minBytes),
		RelayIPs:      conf.RelayIPs,
	}); err != nil {
		log.Warning("server eBPF engine load failed, fail-open (no XDP ingress filter): %v", err)
		return false
	}
	s.xdpActiveMinFrameBytes.Store(int32(minBytes))
	log.Info("server XDP engine loaded: filtering udp/%d with a %d-byte floor, SSH from %v", listenPort, minBytes, conf.RelayIPs)
	return true
}

// xdpServiceConflict names a TCP listener of this daemon that the ingress
// filter would black-hole, or "" when there is none.
//
// Loopback binds are not a conflict: the program is attached to the host's
// default-route interface, so traffic to 127.0.0.1 never reaches it.
func (s *UdpServer) xdpServiceConflict() string {
	if port, enabled := s.GetHttpPort(); enabled {
		if !isLoopbackListenAddr(s.httpConfig.HttpListenIp) {
			return fmt.Sprintf("the HTTP knock listener is enabled on %s:%d",
				listenAddrForLog(s.httpConfig.HttpListenIp), port)
		}
	}
	if s.config != nil && s.config.Metrics.Enabled && !isLoopbackListenAddr(s.config.Metrics.ListenIp) {
		port := s.config.Metrics.ListenPort
		if port == 0 {
			port = defaultMetricsListenPort
		}
		return fmt.Sprintf("the metrics endpoint is enabled on %s:%d",
			listenAddrForLog(s.config.Metrics.ListenIp), port)
	}
	return ""
}

// xdpBlackHolesHttpListener names the listener an http.toml (or etcd) change
// would start behind the attached ingress filter, or "" when there is none.
//
// The counterpart of xdpServiceConflict, for the other direction: that one asks
// "may the filter attach in front of what is already running", this one asks
// "may this listener start in front of a filter that is already attached".
// Both answer the same question — a TCP service on this host that is not SSH
// from RelayIPs cannot be reached — and both have to exist, because either side
// can change without the other: xdpServiceConflict runs once, at startup, and
// http.toml is hot-reloaded.
//
// filterAttached is a parameter rather than an ebpflocal.Loaded() call inside,
// so the decision can be exercised in both states by a test on any host.
func xdpBlackHolesHttpListener(conf *HttpConfig, filterAttached bool) string {
	if conf == nil || !conf.EnableHttp || !filterAttached {
		return ""
	}
	if isLoopbackListenAddr(conf.HttpListenIp) {
		return ""
	}
	return fmt.Sprintf("etc/http.toml asks for the HTTP knock listener on %s:%d",
		listenAddrForLog(conf.HttpListenIp), conf.HttpListenPort)
}

// isLoopbackListenAddr reports whether a bind address reaches only this host.
// An empty address is a wildcard bind (all interfaces), not loopback.
func isLoopbackListenAddr(addr string) bool {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return false
	}
	ip := net.ParseIP(addr)
	return ip != nil && ip.IsLoopback()
}

func listenAddrForLog(addr string) string {
	if strings.TrimSpace(addr) == "" {
		return "0.0.0.0"
	}
	return addr
}

// xdpRelayPrefixes validates a whitelist as the kernel will see it and reports
// whether it may be used at all, logging refusalWhat when it may not.
//
// The check that matters is not "are there entries" but "does every entry
// parse". The two differ exactly where it hurts: RelayIPs = ["relay.opennhp.org"]
// or ["10.0.1.300"] is a non-empty list of strings that names not one address
// the LPM trie can hold, so a count-the-strings guard waves it through and the
// filter attaches — or a reload sweeps a working whitelist away — with a map
// that drops SSH from every source. A partly valid list is refused for the same
// reason: the entry that failed to parse may be the one the operator's own
// session arrives from, and nothing here can tell.
func xdpRelayPrefixes(conf *XdpTomlConfig, fileName string, refusalWhat string) ([]utilsebpf.RelayPrefixKey, bool) {
	prefixes, err := utilsebpf.ParseRelayPrefixes(conf.RelayIPs)
	if err != nil {
		log.Critical("xdp config: %s has an unusable RelayIPs entry (%v) — %s. Entries must be IPv4 addresses or CIDR prefixes; fix RELAY_IPS in the deploy pipeline",
			fileName, err, refusalWhat)
		return nil, false
	}
	if len(prefixes) == 0 {
		log.Critical("xdp config: %s lists no RelayIPs — %s. Fix RELAY_IPS in the deploy pipeline",
			fileName, refusalWhat)
		return nil, false
	}
	return prefixes, true
}

// applyXdpConfig mirrors a parsed xdp.toml's whitelist into the kernel-side map.
// This is the *reload* path only: the whitelist a filter starts with is written
// by the loader before the program is attached (see startXdpFilter), so by the
// time anything here runs there is a live filter enforcing a list an operator
// chose, and refusing is always an option. It reports whether the whitelist in
// conf is now the one the filter is enforcing.
func (s *UdpServer) applyXdpConfig(conf *XdpTomlConfig) bool {
	if !conf.Enabled {
		// Attachment is decided once, at startup; the flag cannot turn a live
		// filter off. Say so rather than leave an operator believing an edit
		// they made took effect.
		log.Warning("xdp config: Enabled=false has no effect on a running filter — the XDP program is attached at startup; set it before a restart to run unfiltered")
	}
	if conf.NhpMinFrameBytes != 0 && conf.NhpMinFrameBytes != s.xdpMinFrameBytes() {
		// The floor is written into the object at load time, so an edit after
		// that is not enforced until the next restart.
		log.Warning("xdp config: NhpMinFrameBytes=%d is not applied — the attached program is enforcing %d bytes; restart nhp-serverd to change it",
			conf.NhpMinFrameBytes, s.xdpMinFrameBytes())
	}

	// See the same guard in startXdpFilter: a whitelist the kernel would end up
	// holding no prefix for closes tcp/22 for every source. Here it is a
	// *reload* refusing to overwrite a working whitelist, which is always the
	// safer of the two outcomes — the filter carries on enforcing the last list
	// an operator actually chose. Validating before the call is what makes that
	// promise true: UpdateRelayIPs removes every entry that is not on the new
	// list, so a file that parses as TOML but not as addresses would otherwise
	// sweep the live whitelist away.
	if _, ok := xdpRelayPrefixes(conf, s.xdpConfigFileName(), "keeping the active whitelist rather than closing tcp/22 for every source"); !ok {
		return false
	}

	if err := ebpflocal.UpdateRelayIPs(conf.RelayIPs); err != nil {
		log.Error("failed to apply xdp relay whitelist, keeping the active whitelist: %v", err)
		return false
	}
	if ebpflocal.Loaded() {
		log.Info("xdp relay whitelist applied: %v", conf.RelayIPs)
	}
	return true
}

// xdpMinFrameBytes is the floor the attached program is enforcing, i.e. the
// value startXdpFilter passed to the loader.
func (s *UdpServer) xdpMinFrameBytes() int {
	if v := s.xdpActiveMinFrameBytes.Load(); v != 0 {
		return int(v)
	}
	return DefaultNhpMinFrameBytes
}

func (s *UdpServer) loadResources() error {
	// resource.toml
	fileName := filepath.Join(ExeDirPath, "etc", "resource.toml")
	content, err := os.ReadFile(fileName)
	if err != nil {
		log.Error("failed to read resource config: %v", err)
	}
	aspMap := make(common.AuthSvcProviderMap)
	// update
	if unmarshalErr := toml.Unmarshal(content, &aspMap); unmarshalErr != nil {
		log.Error("failed to unmarshal resource config: %v", unmarshalErr)
	}
	if updateErr := s.updateResources(aspMap); updateErr != nil {
		// ignore error
		_ = updateErr
	}

	resConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("resource config: %s has been updated", fileName)
		if content, err = s.loadConfigFile(fileName); err == nil {
			aspMap := make(common.AuthSvcProviderMap)
			if err = toml.Unmarshal(content, &aspMap); err == nil {
				_ = s.updateResources(aspMap)
			}
		}
	})
	return nil
}

func (s *UdpServer) loadSourceIps() error {
	// srcip.toml
	fileName := filepath.Join(ExeDirPath, "etc", "srcip.toml")
	content, err := os.ReadFile(fileName)
	if err != nil {
		log.Error("failed to read src ip config: %v", err)
	}

	// update
	srcIpMap := make(map[string][]*common.NetAddress)
	if unmarshalErr := toml.Unmarshal(content, &srcIpMap); unmarshalErr != nil {
		log.Error("failed to unmarshal src ip config: %v", unmarshalErr)
	}
	if updateErr := s.updateSourceIps(srcIpMap); updateErr != nil {
		// ignore error
		_ = updateErr
	}

	srcipConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("src ip config: %s has been updated", fileName)
		if content, err = s.loadConfigFile(fileName); err == nil {
			if err = toml.Unmarshal(content, &srcIpMap); err == nil {
				_ = s.updateSourceIps(srcIpMap)
			}
		}
	})
	return nil
}

func (s *UdpServer) initRemoteConn() error {
	// remote.toml
	fileName := filepath.Join(ExeDirPath, "etc", "remote.toml")

	_, e := os.Stat(fileName)
	if os.IsNotExist(e) {
		//remote.toml file not found,use local config
		return nil
	}

	content, err := os.ReadFile(fileName)
	if err != nil {
		log.Error("failed to read remote config: %v", err)
		return err
	}

	var conf RemoteConfig
	if err = toml.Unmarshal(content, &conf); err != nil {
		log.Error("failed to unmarshal remote config: %v", err)
		return err
	}

	if strings.EqualFold(conf.Provider, "etcd") {
		if len(conf.Endpoints) == 0 {
			log.Error("remote config has no endpoints,open nhp server will startup with local configuration")
			return nil
		}

		if len(conf.Key) == 0 {
			log.Error("remote config has no key,open nhp server will startup with local configuration")
			return nil
		}

		s.etcdConn = &etcd.EtcdConn{
			Endpoints: conf.Endpoints,
			Username:  conf.Username,
			Password:  conf.Password,
			Key:       conf.Key,
		}

		err = s.etcdConn.InitClient()
		return err
	} else {
		return errors.New("unknown remote provider")
	}

}

func (s *UdpServer) loadRemoteBaseConfig() error {
	var serverEtcdConfig ServerEtcdConfig
	value, err := s.etcdConn.GetValue()
	if err != nil {
		return err
	}
	if err = toml.Unmarshal(value, &serverEtcdConfig); err != nil {
		log.Error("failed to unmarshal remote config: %v", err)
		return err
	}

	err = s.updateBaseConfig(serverEtcdConfig.BaseConfig)
	return err
}

func (s *UdpServer) loadRemoteConfig() error {
	value, err := s.etcdConn.GetValue()
	if err != nil {
		return err
	}
	//base config has been loaded and no secondary loading is required
	if err = s.updateEtcdConfig(value, false); err != nil {
		return err
	}

	go s.etcdConn.WatchValue(func(val []byte) {
		s.remoteConfigUpdateMutex.Lock()
		defer s.remoteConfigUpdateMutex.Unlock()
		_ = s.updateEtcdConfig(val, true)
	})

	return nil
}

func (s *UdpServer) updateEtcdConfig(content []byte, baseLoad bool) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	var serverEtcdConfig ServerEtcdConfig
	if err = toml.Unmarshal(content, &serverEtcdConfig); err != nil {
		log.Error("failed to unmarshal remote config: %v", err)
		return err
	}
	if baseLoad {
		_ = s.updateBaseConfig(serverEtcdConfig.BaseConfig)
	}
	_ = s.updateHttpConfig(serverEtcdConfig.HttpConfig)
	_ = s.updateACPeers(serverEtcdConfig.ACs)
	_ = s.updateAgentPeers(serverEtcdConfig.Agents)
	_ = s.updateDePeers(serverEtcdConfig.DBs)

	aspMap := make(common.AuthSvcProviderMap)
	for _, aspData := range serverEtcdConfig.AuthServiceId {
		aspId := aspData.AuthSvcId
		aspMap[aspId] = aspData
	}
	_ = s.updateResources(aspMap)

	srcIpMap := make(map[string][]*common.NetAddress)
	for _, srcIp := range serverEtcdConfig.SrcIps {
		ips := make([]*common.NetAddress, 0)
		for _, ip := range srcIp.Ip {
			addr := &common.NetAddress{
				Ip: ip,
			}
			ips = append(ips, addr)
		}
		srcIpMap[srcIp.SrcIp] = ips
	}
	_ = s.updateSourceIps(srcIpMap)

	return err
}

func (s *UdpServer) loadConfigFile(file string) (content []byte, err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})
	content, err = os.ReadFile(file)
	if err != nil {
		log.Error("failed to read base config: %v", err)
	}
	return
}

func (s *UdpServer) updateBaseConfig(conf Config) (err error) {
	if s.config == nil {
		s.config = &conf
		s.log.SetLogLevel(conf.LogLevel)
		return err
	}

	// update
	if s.config.LogLevel != conf.LogLevel {
		log.Info("set base log level to %d", conf.LogLevel)
		s.log.SetLogLevel(conf.LogLevel)
		s.config.LogLevel = conf.LogLevel
	}

	if s.config.DisableAgentValidation != conf.DisableAgentValidation {
		if s.device != nil {
			// Read-modify-write: SetOption replaces the whole options struct,
			// so building a fresh one here would wipe OnPacketDropped and
			// PeerLookupFallback (both installed in udpserver.Start). Mutate
			// only the field that changed — same pattern as Start uses.
			opt := s.device.GetOption()
			opt.DisableAgentPeerValidation = conf.DisableAgentValidation
			s.device.SetOption(opt)
		}
		s.config.DisableAgentValidation = conf.DisableAgentValidation
	}

	if s.config.AllowPrivateRelaySource != conf.AllowPrivateRelaySource {
		log.Info("AllowPrivateRelaySource set to %v (relay SourceAddr public-IP check is %s)",
			conf.AllowPrivateRelaySource,
			map[bool]string{true: "disabled", false: "enforced"}[conf.AllowPrivateRelaySource])
		s.config.AllowPrivateRelaySource = conf.AllowPrivateRelaySource
		// Mirror onto the atomic-read field that hot paths consult; see
		// the field comment in UdpServer for why the bool isn't read
		// directly from s.config under no lock.
		s.allowPrivateRelaySource.Store(conf.AllowPrivateRelaySource)
	}

	if s.config.DefaultCipherScheme != conf.DefaultCipherScheme {
		log.Info("set default cipher scheme to %d", conf.DefaultCipherScheme)
		s.config.DefaultCipherScheme = conf.DefaultCipherScheme
	}

	// ForceOverload: the in-memory Overload state is sticky for the
	// process lifetime (see the field docstring), so a reload can't
	// actually toggle behavior. But silently dropping the new value
	// leaves s.config disagreeing with the on-disk config.toml — which
	// confuses anyone reading the in-memory view. Adopt the new value
	// for accuracy and surface a warning so the operator knows a
	// restart is required.
	if s.config.ForceOverload != conf.ForceOverload {
		log.Warning("ForceOverload changed in config (%v -> %v) on reload; the in-memory Overload state is sticky for the process lifetime, restart to apply",
			s.config.ForceOverload, conf.ForceOverload)
		s.config.ForceOverload = conf.ForceOverload
		// Mirror onto the atomic-read field; same reasoning as
		// AllowPrivateRelaySource above. The Overload itself remains
		// sticky for the process lifetime even when this transitions
		// false → true → false; only the teardown predicate observes
		// the new value.
		s.forceOverload.Store(conf.ForceOverload)
	}

	if s.config.AttestationScheme != conf.AttestationScheme {
		resolved := verifier.ResolveScheme(conf.AttestationScheme)
		if resolved == verifier.SchemeTest {
			// Critical (not Warning): SchemeTest disables a real attestation
			// safeguard, and Warning is filtered out of default journalctl
			// views — exactly where an operator who flipped this on for a
			// quick demo would forget to flip it back. Match the demo-cookie
			// and ForceOverload Criticals so all three "you are running with
			// a guard off" signals share one severity.
			log.Critical("AttestationScheme=%q: DHP knock path accepts self-asserted evidence with NO cryptographic assurance — for non-TEE demos only, never on a production-facing host", conf.AttestationScheme)
		} else {
			log.Info("AttestationScheme set to %q (csv: Hygon attestation chain verified)", conf.AttestationScheme)
		}
		s.config.AttestationScheme = conf.AttestationScheme
		s.attestationScheme.Store(string(resolved))
	}

	// [Audit]: the ledger handle is opened once at startup (initAuditLedger)
	// and never re-opened on reload — changing FilePath, SigningKeyBase64,
	// Fsync, FailClosed or MaxSizeBytes at runtime does nothing until a
	// restart. Silently dropping the new value would leave s.config
	// disagreeing with config.toml, so adopt it for read-consistency and
	// warn, exactly as ForceOverload above. Editing SigningKeyBase64 and
	// believing it took effect is the bad outcome this warning prevents.
	if s.config.Audit != conf.Audit {
		log.Warning("[Audit] config changed on reload; the ledger is opened once at startup — restart to apply the new settings")
		s.config.Audit = conf.Audit
	}

	// Cookie signing key / window: only re-apply when the operator
	// actually changed something AND the new key parses. A blanked-out
	// field on reload is treated as "leave the running key alone" rather
	// than silently regenerating a random one (that'd break a cluster on
	// the next reload). Window-only updates are allowed.
	keyChanged := s.config.CookieSigningKeyBase64 != conf.CookieSigningKeyBase64
	windowChanged := s.config.CookieTimeWindowSeconds != conf.CookieTimeWindowSeconds
	if (keyChanged || windowChanged) && s.device != nil {
		newKey, decodeErr := decodeCookieSigningKey(conf.CookieSigningKeyBase64)
		if decodeErr != nil {
			log.Warning("ignoring CookieSigningKeyBase64 change: %v (keeping running key)", decodeErr)
		} else {
			// Mirror the startup demo-key guard in udpserver.Start: a
			// hot-reload that swaps in the committed docker-compose demo
			// key is just as dangerous as booting with it, and previously
			// got no warning at all. Check the raw config string (not the
			// decoded bytes) so this fires regardless of the preservation
			// logic below.
			if keyChanged && conf.CookieSigningKeyBase64 == shippedDemoCookieSigningKeyBase64 {
				log.Critical("CookieSigningKeyBase64 reloaded to the docker-compose demo value committed at " +
					"docker/nhp-server/etc/config.toml — this key is PUBLIC. Regenerate before any " +
					"deployment reachable from outside the host.")
			}
			currKey, currWin := s.device.StatelessCookieParams()
			if len(newKey) == 0 {
				// Preserve the running key whenever the config field is
				// empty — NOT only when it changed. The single-instance
				// flow never sets CookieSigningKeyBase64: udpserver.Start
				// mints a random per-process key, so s.config's field
				// stays "" and keyChanged is false on a window-only
				// reload. Gating preservation on keyChanged here would
				// let newKey stay nil and hand SetStatelessCookieParams a
				// nil key, which silently DISABLES stateless cookies and
				// stalls every agent that hits the overload path. Only
				// the operator-cleared-a-configured-key case warrants a
				// warning; the always-empty case is normal.
				if keyChanged {
					log.Warning("CookieSigningKeyBase64 cleared on reload; keeping previous key in memory")
				}
				newKey = currKey
			}
			newWin := conf.CookieTimeWindowSeconds
			if newWin <= 0 {
				newWin = DefaultCookieTimeWindowSeconds
			}
			if !bytesEqualConstantTime(newKey, currKey) || int64(newWin) != currWin {
				s.device.SetStatelessCookieParams(newKey, newWin)
				log.Info("stateless cookie params updated (window=%ds, keyChanged=%v)", newWin, !bytesEqualConstantTime(newKey, currKey))
			}
		}
		// Only persist the new base64 into s.config when we actually
		// applied (or were able to preserve) a usable key — i.e. the
		// operator handed us a non-empty, well-formed value. If we
		// instead write back an empty or malformed string, the next
		// reload will see no delta (s.config == conf), skip the whole
		// validation/preservation block, and silently leave the
		// running device key disagreeing with the in-memory config —
		// so the operator stops seeing the "cleared, keeping previous
		// key" / "ignoring CookieSigningKeyBase64 change" warning even
		// though the divergence persists. Window is always written
		// back since it's plain numeric and re-validated each reload.
		if decodeErr == nil && conf.CookieSigningKeyBase64 != "" {
			s.config.CookieSigningKeyBase64 = conf.CookieSigningKeyBase64
		}
		s.config.CookieTimeWindowSeconds = conf.CookieTimeWindowSeconds
	}

	return err
}

func bytesEqualConstantTime(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

func (s *UdpServer) updateHttpConfig(httpConf HttpConfig) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	// set http default timeout values
	// 4.5s for read timeout, 4s for write timeout, 5s for idle timeout
	if httpConf.ReadTimeoutMs == 0 {
		httpConf.ReadTimeoutMs = DefaultHttpRequestReadTimeoutMs
	}
	if httpConf.WriteTimeoutMs == 0 {
		httpConf.WriteTimeoutMs = DefaultHttpResponseWriteTimeoutMs
	}
	if httpConf.IdleTimeoutMs == 0 {
		httpConf.IdleTimeoutMs = DefaultHttpServerIdleTimeoutMs
	}

	// update
	if httpConf.EnableHttp {
		// The startup refusal in startXdpFilter only sees the http.toml the
		// daemon booted with, and this config is hot — the file watcher and the
		// etcd watcher both land here. Turning EnableHttp on under a live
		// filter would bind a listener the XDP program drops at the driver: the
		// service answers nothing, nothing in user space says why, and that is
		// the exact failure the startup check exists to prevent. Refusing is
		// the lesser outcome of the two, and the only one that leaves a reason
		// behind. Detaching the filter instead is not on the table: it would
		// re-open tcp/22 to the internet on the strength of an http.toml edit.
		if reason := xdpBlackHolesHttpListener(&httpConf, ebpflocal.Loaded()); reason != "" {
			log.Critical("refusing to start the HTTP knock listener: %s. The XDP ingress filter is attached and drops inbound TCP other than SSH from RelayIPs, so the listener would be unreachable with nothing to say so. Bind it to 127.0.0.1, or set Enabled = false in etc/xdp.toml and restart nhp-serverd.",
				reason)
			return fmt.Errorf("http listener conflicts with the attached XDP ingress filter: %s", reason)
		}
		// start http server
		if s.httpServer == nil || !s.httpServer.IsRunning() {
			if s.httpServer != nil {
				// stop old http server
				go s.httpServer.Stop()
			}
			hs := &HttpServer{}
			s.httpServer = hs
			err = hs.Start(s, &httpConf)
			if err != nil {
				return err
			}
		}
	} else {
		// stop http server
		if s.httpServer != nil && s.httpServer.IsRunning() {
			go s.httpServer.Stop()
			s.httpServer = nil
		}
	}

	s.httpConfig = &httpConf
	return err
}

func (s *UdpServer) updateACPeers(peers []*core.UdpPeer) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	acPeerMap := make(map[string]*core.UdpPeer)
	for _, p := range peers {
		p.Type = core.NHP_AC
		s.device.AddPeer(p)
		acPeerMap[p.PublicKeyBase64()] = p
	}

	// remove old peers from device
	s.acPeerMapMutex.Lock()
	defer s.acPeerMapMutex.Unlock()
	for pubKey := range s.acPeerMap {
		if _, found := acPeerMap[pubKey]; !found {
			s.device.RemovePeer(pubKey)
		}
	}
	s.acPeerMap = acPeerMap

	return err
}

func (s *UdpServer) updateAgentPeers(peers []*core.UdpPeer) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})
	agentPeerMap := make(map[string]*core.UdpPeer)
	for _, p := range peers {
		p.Type = core.NHP_AGENT
		s.device.AddPeer(p)
		agentPeerMap[p.PublicKeyBase64()] = p
	}

	// remove old peers from device
	s.agentPeerMapMutex.Lock()
	defer s.agentPeerMapMutex.Unlock()
	for pubKey := range s.agentPeerMap {
		if _, found := agentPeerMap[pubKey]; !found {
			s.device.RemovePeer(pubKey)
		}
	}
	s.agentPeerMap = agentPeerMap

	return err
}

func (s *UdpServer) updateRelayPeers(peers []*core.UdpPeer) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	relayPeerMap := make(map[string]*core.UdpPeer)
	for _, p := range peers {
		p.Type = core.NHP_RELAY
		s.device.AddPeer(p)
		relayPeerMap[p.PublicKeyBase64()] = p
	}

	// remove old peers from device
	s.relayPeerMapMutex.Lock()
	defer s.relayPeerMapMutex.Unlock()
	for pubKey := range s.relayPeerMap {
		if _, found := relayPeerMap[pubKey]; !found {
			s.device.RemovePeer(pubKey)
		}
	}
	s.relayPeerMap = relayPeerMap

	return err
}

func (s *UdpServer) updateResources(aspMap common.AuthSvcProviderMap) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	for aspId, aspData := range aspMap {
		aspData.AuthSvcId = aspId
		if len(aspData.PluginPath) > 0 {
			h := plugins.ReadPluginHandler(aspData.PluginPath)
			if h != nil {
				_ = s.LoadPlugin(aspId, h)
			}
		}

		for resId, res := range aspData.ResourceGroups {
			// Note: res is a pointer, so we can update its value
			res.AuthServiceId = aspId
			res.ResourceId = resId
		}
	}

	s.authServiceMapMutex.Lock()
	defer s.authServiceMapMutex.Unlock()
	s.authServiceMap = aspMap

	return err
}

func (s *UdpServer) updateSourceIps(srcIpMap map[string][]*common.NetAddress) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	s.srcIpAssociatedAddrMapMutex.Lock()
	defer s.srcIpAssociatedAddrMapMutex.Unlock()
	s.srcIpAssociatedAddrMap = srcIpMap

	return err
}

func (s *UdpServer) StopConfigWatch() {
	if baseConfigWatch != nil {
		baseConfigWatch.Close()
	}
	if httpConfigWatch != nil {
		httpConfigWatch.Close()
	}
	if acConfigWatch != nil {
		acConfigWatch.Close()
	}
	if agentConfigWatch != nil {
		agentConfigWatch.Close()
	}
	if resConfigWatch != nil {
		resConfigWatch.Close()
	}
	if srcipConfigWatch != nil {
		srcipConfigWatch.Close()
	}
	//add dbConfigWatch
	if dbConfigWatch != nil {
		dbConfigWatch.Close()
	}
	if relayConfigWatch != nil {
		relayConfigWatch.Close()
	}
	if xdpConfigWatch != nil {
		xdpConfigWatch.Close()
	}
	if teeWatch != nil {
		teeWatch.Close()
	}
}

// updateDePeers
func (s *UdpServer) updateDePeers(peers []*core.UdpPeer) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	dbPeerMap := make(map[string]*core.UdpPeer)
	for _, p := range peers {
		p.Type = core.NHP_DB
		s.device.AddPeer(p)
		dbPeerMap[p.PublicKeyBase64()] = p
	}

	// remove old peers from device
	s.dbPeerMapMutex.Lock()
	defer s.dbPeerMapMutex.Unlock()
	for pubKey := range s.dbPeerMap {
		if _, found := dbPeerMap[pubKey]; !found {
			s.device.RemovePeer(pubKey)
		}
	}
	s.dbPeerMap = dbPeerMap
	return err
}

// update tee
func (s *UdpServer) updateTee(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read tee config: %v", err)
	}

	var tees TeeAttestationReports
	teeMap := make(map[string]*TeeAttestationReport)
	if unmarshalErr := toml.Unmarshal(content, &tees); unmarshalErr != nil {
		log.Error("failed to unmarshal device peer config: %v", unmarshalErr)
	}
	for _, tee := range tees.TEEs {
		// Skip empty Measure entries: a successful Verify() that
		// produced an empty measure would otherwise match this entry
		// and silently approve any knock.
		if tee.Measure == "" {
			log.Warning("tee.toml entry %q skipped: Measure must not be empty", tee.SerialNumber)
			continue
		}
		teeMap[tee.Measure] = tee
	}

	s.teeMapMutex.Lock()
	defer s.teeMapMutex.Unlock()
	s.teeMap = teeMap
	return err
}

func (s *UdpServer) AppraiseEvidence(evidenceBase64 string) bool {
	scheme := s.currentAttestationScheme()

	attestationVerifier, err := verifier.NewVerifier(evidenceBase64, scheme)
	if err != nil {
		log.Error("failed to create attestation verifier (scheme=%s): %v", scheme, err)
		return false
	}

	if err := attestationVerifier.Verify(); err != nil {
		log.Error("failed to verify attestation (scheme=%s): %v", scheme, err)
		return false
	}

	measure := attestationVerifier.GetMeasure()
	sn := attestationVerifier.GetSerialNumber()
	if measure == "" {
		// A successful Verify() that produces an empty measure would
		// otherwise match a misconfigured tee.toml entry with Measure=""
		// and silently approve the knock.
		log.Error("attestation produced an empty measure (scheme=%s)", scheme)
		return false
	}

	s.teeMapMutex.Lock()
	defer s.teeMapMutex.Unlock()

	tee, exist := s.teeMap[measure]
	if !exist || tee.SerialNumber != sn {
		// Only an appraisal that actually matched may mark the entry
		// verified; setting the flag before the serial-number comparison
		// caused failed appraisals to flip Verified=true and corrupt the
		// operator view.
		return false
	}
	tee.Verified = true
	return true
}

// currentAttestationScheme returns the active verifier.Scheme for this
// server. The value is read from an atomic mirror populated at startup
// (and on every config reload) so the per-knock handler does not race
// with the file-watch goroutine writing s.config.
func (s *UdpServer) currentAttestationScheme() verifier.Scheme {
	if v := s.attestationScheme.Load(); v != nil {
		if str, ok := v.(string); ok {
			return verifier.Scheme(str)
		}
	}
	return verifier.SchemeCSV
}
