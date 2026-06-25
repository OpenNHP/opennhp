package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/OpenNHP/opennhp/nhp/etcd"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/core/verifier"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/plugins"
	"github.com/OpenNHP/opennhp/nhp/utils"

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
	if err := toml.Unmarshal(content, &config); err != nil {
		log.Error("failed to unmarshal base config: %v", err)
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
	if err := toml.Unmarshal(content, &httpConf); err != nil {
		log.Error("failed to unmarshal http config: %v", err)
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
	if err := toml.Unmarshal(contentAC, &acPeers); err != nil {
		log.Error("failed to unmarshal ac peers config: %v", err)
	}

	if err := s.updateACPeers(acPeers.ACs); err != nil {
		// ignore error
		_ = err
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
	if err := toml.Unmarshal(contentAgent, &agentPeers); err != nil {
		log.Error("failed to unmarshal agent peers config: %v", err)
	}
	if err := s.updateAgentPeers(agentPeers.Agents); err != nil {
		// ignore error
		_ = err
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
		if err := toml.Unmarshal(contentDE, &dePeers); err != nil {
			log.Error("failed to unmarshal db peers config: %v", err)
		}
		if err := s.updateDePeers(dePeers.DBs); err != nil {
			// ignore error
			_ = err
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
		if err := toml.Unmarshal(contentRelay, &relayPeers); err != nil {
			log.Error("failed to unmarshal relay peers config: %v", err)
		}
		if err := s.updateRelayPeers(relayPeers.Relays); err != nil {
			_ = err
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

func (s *UdpServer) loadResources() error {
	// resource.toml
	fileName := filepath.Join(ExeDirPath, "etc", "resource.toml")
	content, err := os.ReadFile(fileName)
	if err != nil {
		log.Error("failed to read resource config: %v", err)
	}
	aspMap := make(common.AuthSvcProviderMap)
	// update
	if err := toml.Unmarshal(content, &aspMap); err != nil {
		log.Error("failed to unmarshal resource config: %v", err)
	}
	if err := s.updateResources(aspMap); err != nil {
		// ignore error
		_ = err
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
	if err := toml.Unmarshal(content, &srcIpMap); err != nil {
		log.Error("failed to unmarshal src ip config: %v", err)
	}
	if err := s.updateSourceIps(srcIpMap); err != nil {
		// ignore error
		_ = err
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
			s.device.SetOption(core.DeviceOptions{
				DisableAgentPeerValidation: conf.DisableAgentValidation,
			})
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

	// Cookie signing key / window: only re-apply when the operator
	// actually changed something AND the new key parses. A blanked-out
	// field on reload is treated as "leave the running key alone" rather
	// than silently regenerating a random one (that'd break a cluster on
	// the next reload). Window-only updates are allowed.
	keyChanged := s.config.CookieSigningKeyBase64 != conf.CookieSigningKeyBase64
	windowChanged := s.config.CookieTimeWindowSeconds != conf.CookieTimeWindowSeconds
	if (keyChanged || windowChanged) && s.device != nil {
		newKey, err := decodeCookieSigningKey(conf.CookieSigningKeyBase64)
		if err != nil {
			log.Warning("ignoring CookieSigningKeyBase64 change: %v (keeping running key)", err)
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
		if err == nil && conf.CookieSigningKeyBase64 != "" {
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
	if err := toml.Unmarshal(content, &tees); err != nil {
		log.Error("failed to unmarshal device peer config: %v", err)
	}
	for _, tee := range tees.TEEs {
		teeMap[tee.Measure] = tee
	}

	s.teeMapMutex.Lock()
	defer s.teeMapMutex.Unlock()
	s.teeMap = teeMap
	return err
}

func (s *UdpServer) AppraiseEvidence(evidenceBase64 string) bool {
	var measure string
	var sn string

	attestationVerifier, err := verifier.NewVerifier(evidenceBase64)
	if err != nil {
		log.Error("failed to create attestation verifier: %v", err)
		return false
	}

	if err := attestationVerifier.Verify(); err != nil {
		log.Error("failed to verify attestation: %v", err)
		return false
	}

	measure = attestationVerifier.GetMeasure()
	sn = attestationVerifier.GetSerialNumber()

	s.teeMapMutex.Lock()
	defer s.teeMapMutex.Unlock()

	if _, exist := s.teeMap[measure]; exist {
		s.teeMap[measure].Verified = true
		return s.teeMap[measure].SerialNumber == sn
	}

	return false
}
