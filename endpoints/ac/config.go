package ac

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/common/clusterconfig"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/etcd"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/utils"
)

var (
	baseConfigWatch io.Closer
	httpConfigWatch io.Closer
	serverPeerWatch io.Closer

	errLoadConfig = fmt.Errorf("config load error")
)

const (
	FilterMode_IPTABLES = iota // 0
	FilterMode_EBPFXDP         // 1
)

// ACEtcdConfig is the remote-config (etcd) shape. Servers carries the
// shared cluster schema so the etcd value is identical to the on-disk
// server.toml. The previous Endpoints-string form lived only on this
// branch and is removed in the AC → ClusterConfig migration; redeploy
// any etcd values written under that schema.
type ACEtcdConfig struct {
	BaseConfig Config
	HttpConfig HttpConfig
	Servers    []*clusterconfig.ClusterConfig
}

type Config struct {
	PrivateKeyBase64    string          `json:"privateKey"`
	ACId                string          `json:"acId"`
	DefaultIp           string          `json:"defaultIp"`
	AuthServiceId       string          `json:"aspId"`
	ResourceIds         []string        `json:"resIds"`
	Servers             []*core.UdpPeer `json:"servers"`
	IpPassMode          int             `json:"ipPassMode"` // 0: pass the knock source IP, 1: use pre-access mode and release the access source IP
	LogLevel            int             `json:"logLevel"`
	DefaultCipherScheme int             `json:"defaultCipherScheme"`
	FilterMode          int             `json:"filterMode"`
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
	HttpListenPort int
	TLSCertFile    string
	TLSKeyFile     string
}

// Peers is the top-level shape of server.toml. Each entry is one
// logical nhp-server identity (one pubkey) reachable at 1..N instances
// — same schema as nhp-agent's server.toml (see
// nhp/common/clusterconfig). nhp-ac doesn't run a multi-step handshake
// against any single instance, so LoadBalance / StickyInstance are
// loaded but ignored: AOL fan-out broadcasts to every instance for
// failover, not load-balancing.
type Peers struct {
	Servers []*clusterconfig.ClusterConfig
}

// expandServerPeers turns a parsed + normalized cluster list into the
// flat []*core.UdpPeer the rest of nhp-ac consumes. One cluster with N
// Instances becomes N UdpPeer rows sharing a pubkey — that fan-out is
// what lets the AC register (NHP_AOL) with every instance in a
// same-pubkey nhp-server cluster.
//
// The caller is responsible for running clusterconfig.Normalize on the
// input first; this helper trusts that every entry has at least one
// Instances row with a non-zero Port.
func expandServerPeers(clusters []*clusterconfig.ClusterConfig) []*core.UdpPeer {
	out := make([]*core.UdpPeer, 0, len(clusters))
	for _, c := range clusters {
		for _, inst := range c.Instances {
			out = append(out, &core.UdpPeer{
				PubKeyBase64: c.PubKeyBase64,
				Hostname:     inst.Host,
				Ip:           inst.Ip,
				Port:         inst.Port,
				ExpireTime:   c.ExpireTime,
			})
		}
	}
	return out
}

// normalizeAndExpand parses, validates, and flattens the [[Servers]]
// list in one step. Returns (nil, err) on any validation failure so
// callers can fail-close — leaving the running peerMap untouched on
// reload, or refusing to start on initial load.
func normalizeAndExpand(clusters []*clusterconfig.ClusterConfig) ([]*core.UdpPeer, error) {
	if err := clusterconfig.Normalize(clusters, clusterconfig.Options{
		ConsumerLabel: "ac",
		RequireName:   false,
	}, log.Warning); err != nil {
		return nil, err
	}
	return expandServerPeers(clusters), nil
}

func (a *UdpAC) loadBaseConfig() error {
	// config.toml
	fileName := filepath.Join(ExeDirPath, "etc", "config.toml")
	content, err := os.ReadFile(fileName)
	if err != nil {
		log.Error("failed to read base config: %v", err)
	}

	var conf Config
	if unmarshalErr := toml.Unmarshal(content, &conf); unmarshalErr != nil {
		log.Error("failed to unmarshal base config: %v", unmarshalErr)
	}
	if updateErr := a.updateBaseConfig(conf); updateErr != nil {
		// report base config error
		return updateErr
	}

	baseConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("base config: %s has been updated", fileName)
		if content, err = a.loadConfigFile(fileName); err == nil {
			if err = toml.Unmarshal(content, &conf); err == nil {
				_ = a.updateBaseConfig(conf)
			}

		}
	})
	return nil
}

func (a *UdpAC) loadHttpConfig() error {
	// http.toml
	fileName := filepath.Join(ExeDirPath, "etc", "http.toml")
	content, err := os.ReadFile(fileName)
	if err != nil {
		log.Error("failed to read http config: %v", err)
	}

	var httpConf HttpConfig
	if unmarshalErr := toml.Unmarshal(content, &httpConf); unmarshalErr != nil {
		log.Error("failed to unmarshal http config: %v", unmarshalErr)
	}
	if updateErr := a.updateHttpConfig(httpConf); updateErr != nil {
		// ignore error
		_ = updateErr
	}

	httpConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("http config: %s has been updated", fileName)
		if content, err = a.loadConfigFile(fileName); err == nil {
			if err = toml.Unmarshal(content, &httpConf); err == nil {
				_ = a.updateHttpConfig(httpConf)
			}
		}
	})
	return nil
}

func (a *UdpAC) loadPeers() error {
	// server.toml
	fileName := filepath.Join(ExeDirPath, "etc", "server.toml")
	content, readErr := os.ReadFile(fileName)
	if readErr != nil {
		log.Error("failed to read server peer config: %v", readErr)
	}

	// update
	var peers Peers
	var unmarshalErr error
	if unmarshalErr = toml.Unmarshal(content, &peers); unmarshalErr != nil {
		log.Error("failed to unmarshal server peer config: %v", unmarshalErr)
	}
	expanded, expandErr := normalizeAndExpand(peers.Servers)
	var initialErr error
	if expandErr != nil {
		// Initial load with a malformed file: refuse to start with an
		// empty peerMap. Previously this only logged Critical and fell
		// through; the daemon would come up with zero servers, AOL/AOP
		// fan-out would silently no-op, and the operator would have to
		// chase dropped packets back to a single log line. Propagating
		// the error to the caller turns the failure into a refusal-to-
		// start that's much easier to diagnose. The reload branch
		// below keeps its existing keep-previous-table discipline —
		// that's the right behaviour mid-flight when a running
		// process already has a good peer table.
		log.Critical("loadPeers: %v; refusing to start with empty peer table", expandErr)
		initialErr = expandErr
	} else if updateErr := a.updateServerPeers(expanded); updateErr != nil {
		// ignore error
		_ = updateErr
	}

	serverPeerWatch = utils.WatchFile(fileName, func() {
		log.Info("server peer config: %s has been updated", fileName)
		reloadContent, reloadErr := a.loadConfigFile(fileName)
		if reloadErr != nil {
			return
		}
		var reloadPeers Peers
		if uerr := toml.Unmarshal(reloadContent, &reloadPeers); uerr != nil {
			log.Error("failed to unmarshal server peer config on reload: %v", uerr)
			return
		}
		expanded, expandErr := normalizeAndExpand(reloadPeers.Servers)
		if expandErr != nil {
			// Reload: do NOT call updateServerPeers — that
			// would rewrite the live peerMap. Keep the running
			// peer table and let the operator fix the file.
			log.Critical("loadPeers reload: %v; keeping previous peer table", expandErr)
			return
		}
		_ = a.updateServerPeers(expanded)
	})

	return initialErr
}

func (a *UdpAC) updateBaseConfig(conf Config) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	if a.config == nil {
		a.config = &conf
		a.log.SetLogLevel(conf.LogLevel)
		return err
	}

	// update
	if a.config.LogLevel != conf.LogLevel {
		log.Info("set base log level to %d", conf.LogLevel)
		a.log.SetLogLevel(conf.LogLevel)
		a.config.LogLevel = conf.LogLevel
	}

	if a.config.DefaultIp != conf.DefaultIp {
		log.Info("set default ip mode to %s", conf.DefaultIp)
		a.config.DefaultIp = conf.DefaultIp
	}

	if a.config.IpPassMode != conf.IpPassMode {
		log.Info("set ip pass mode to %d", conf.IpPassMode)
		a.config.IpPassMode = conf.IpPassMode
	}

	if a.config.DefaultCipherScheme != conf.DefaultCipherScheme {
		log.Info("set default cipher scheme to %d", conf.DefaultCipherScheme)
		a.config.DefaultCipherScheme = conf.DefaultCipherScheme
	}

	return err
}

func (a *UdpAC) updateHttpConfig(httpConf HttpConfig) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	// update
	if httpConf.EnableHttp {
		// start http server
		if a.httpServer == nil || !a.httpServer.IsRunning() {
			if a.httpServer != nil {
				// stop old http server
				go a.httpServer.Stop()
			}
			hs := &HttpAC{}
			a.httpServer = hs
			err = hs.Start(a, &httpConf)
			if err != nil {
				return err
			}
		}
	} else {
		// stop http server
		if a.httpServer != nil && a.httpServer.IsRunning() {
			go a.httpServer.Stop()
			a.httpServer = nil
		}
	}

	a.httpConfig = &httpConf
	return err
}

func (a *UdpAC) updateServerPeers(peers []*core.UdpPeer) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	// One [[Servers]] entry may have N endpoints sharing a pubkey, so the
	// AC-side map is keyed by (pubkey, addr) to keep each endpoint distinct.
	// The device's peerMap is still keyed by pubkey alone (which is fine — its
	// only job is gating "is this pubkey in the whitelist?"; per-connection
	// state lives on ConnectionData since commit A). We collapse to a set of
	// pubkeys for the device add/remove diff.
	serverPeerMap := make(map[string]*core.UdpPeer, len(peers))
	pubKeyPresent := make(map[string]struct{})
	for _, p := range peers {
		p.Type = core.NHP_SERVER
		a.device.AddPeer(p)
		serverPeerMap[endpointKey(p)] = p
		pubKeyPresent[p.PublicKeyBase64()] = struct{}{}
	}
	a.config.Servers = peers

	a.serverPeerMutex.Lock()
	defer a.serverPeerMutex.Unlock()

	// Remove from device any pubkey that no longer appears in the new config.
	// Iterate the old map to find pubkeys that have fully disappeared.
	oldPubKeys := make(map[string]struct{})
	for _, oldPeer := range a.serverPeerMap {
		oldPubKeys[oldPeer.PublicKeyBase64()] = struct{}{}
	}
	for pubKey := range oldPubKeys {
		if _, stillPresent := pubKeyPresent[pubKey]; !stillPresent {
			a.device.RemovePeer(pubKey)
		}
	}
	a.serverPeerMap = serverPeerMap

	return err
}

// endpointKey is the AC-internal map key for a server peer. It must keep
// same-pubkey peers at different addresses distinct, so the discovery
// fan-out launches one routine per (pubkey, addr).
//
// Hostname is part of the key because legacy entries that use only
// Hostname (no Ip) leave p.Ip empty, so two same-pubkey entries with
// different hostnames would otherwise collide on "pk|:port" and one
// would silently overwrite the other in serverPeerMap — meaning AOL
// fan-out skips one of the instances. Including Hostname keeps each
// instance distinct until the hostname is resolved.
func endpointKey(p *core.UdpPeer) string {
	return "pk=" + p.PublicKeyBase64() +
		"|host=" + p.Hostname +
		"|ip=" + p.Ip +
		":" + strconv.Itoa(p.Port)
}
func (a *UdpAC) loadConfigFile(file string) (content []byte, err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})
	content, err = os.ReadFile(file)
	if err != nil {
		log.Error("failed to read base config: %v", err)
	}
	return
}
func (a *UdpAC) initRemoteConn() error {
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

		a.etcdConn = &etcd.EtcdConn{
			Endpoints: conf.Endpoints,
			Username:  conf.Username,
			Password:  conf.Password,
			Key:       conf.Key,
		}

		err = a.etcdConn.InitClient()
		return err
	} else {
		return errors.New("unknown remote provider")
	}

}

func (a *UdpAC) loadRemoteConfig() error {
	value, err := a.etcdConn.GetValue()
	if err != nil {
		return err
	}
	//base config has been loaded and no secondary loading is required
	if err = a.updateEtcdConfig(value, false); err != nil {
		return err
	}

	go a.etcdConn.WatchValue(func(val []byte) {
		a.remoteConfigUpdateMutex.Lock()
		defer a.remoteConfigUpdateMutex.Unlock()
		_ = a.updateEtcdConfig(val, true)
	})

	return nil
}

func (a *UdpAC) loadRemoteBaseConfig() error {
	var acEtcdConfig ACEtcdConfig
	value, err := a.etcdConn.GetValue()
	if err != nil {
		return err
	}
	if err = toml.Unmarshal(value, &acEtcdConfig); err != nil {
		log.Error("failed to unmarshal remote config: %v", err)
		return err
	}

	err = a.updateBaseConfig(acEtcdConfig.BaseConfig)
	return err
}

func (a *UdpAC) updateEtcdConfig(content []byte, baseLoad bool) (err error) {
	var acEtcdConfig ACEtcdConfig
	if err = toml.Unmarshal(content, &acEtcdConfig); err != nil {
		log.Error("failed to unmarshal remote config: %v", err)
		return err
	}

	if baseLoad {
		_ = a.updateBaseConfig(acEtcdConfig.BaseConfig)
	}

	_ = a.updateHttpConfig(acEtcdConfig.HttpConfig)
	expanded, expandErr := normalizeAndExpand(acEtcdConfig.Servers)
	if expandErr != nil {
		// etcd reload: same fail-close discipline as the file watcher —
		// keep the running peer table rather than swap in a config that
		// would silently drop a cluster from peerMap.
		log.Critical("updateEtcdConfig: %v; keeping previous peer table", expandErr)
		return
	}
	_ = a.updateServerPeers(expanded)
	return
}

func (a *UdpAC) IpPassMode() int {
	return a.config.IpPassMode
}

func (a *UdpAC) StopConfigWatch() {
	if baseConfigWatch != nil {
		baseConfigWatch.Close()
	}
	if httpConfigWatch != nil {
		httpConfigWatch.Close()
	}
	if serverPeerWatch != nil {
		serverPeerWatch.Close()
	}
}
