package ac

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

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

type ACEtcdConfig struct {
	BaseConfig Config
	HttpConfig HttpConfig
	Servers    []ServerPeerEntry
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

// ServerPeerEntry is the TOML/etcd loader shape for a server peer entry. One
// entry represents one logical nhp-server identity (a single ECDH keypair);
// Endpoints allows that identity to be reachable at multiple addresses, which
// is how an AC fans AOL/KPL out to a same-pubkey nhp-server cluster.
//
// Legacy single-endpoint configs (Hostname or Ip+Port at the entry root)
// remain supported and are normalized to Endpoints during load.
type ServerPeerEntry struct {
	PubKeyBase64 string
	Hostname     string
	Ip           string
	Port         int
	Endpoints    []string
	ExpireTime   int64
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

type Peers struct {
	Servers []ServerPeerEntry
}

// expandServerPeers normalizes loader-shape entries into []*core.UdpPeer,
// fanning a single entry with N endpoints into N UdpPeer rows that share a
// pubkey. The new fan-out path uses the Endpoints slice; the legacy path
// (Hostname or Ip+Port at the entry root) collapses to a single UdpPeer.
//
// Endpoints, when non-empty, take precedence over the legacy fields; the
// legacy fields are then ignored with an info-level log so misconfigurations
// are visible.
func expandServerPeers(entries []ServerPeerEntry) []*core.UdpPeer {
	out := make([]*core.UdpPeer, 0, len(entries))
	for _, e := range entries {
		if len(e.Endpoints) > 0 {
			if e.Hostname != "" || e.Ip != "" || e.Port != 0 {
				log.Info("server peer %s: Endpoints set, ignoring legacy Hostname/Ip/Port fields", e.PubKeyBase64)
			}
			for _, ep := range e.Endpoints {
				ip, port, err := splitHostPort(ep)
				if err != nil {
					log.Error("server peer %s: skip endpoint %q: %v", e.PubKeyBase64, ep, err)
					continue
				}
				out = append(out, &core.UdpPeer{
					PubKeyBase64: e.PubKeyBase64,
					Ip:           ip,
					Port:         port,
					ExpireTime:   e.ExpireTime,
				})
			}
			continue
		}
		out = append(out, &core.UdpPeer{
			PubKeyBase64: e.PubKeyBase64,
			Hostname:     e.Hostname,
			Ip:           e.Ip,
			Port:         e.Port,
			ExpireTime:   e.ExpireTime,
		})
	}
	return out
}

// splitHostPort parses a "host:port" endpoint string. Bare IPv6 must be in
// brackets; the function accepts both IPv4 "1.2.3.4:5000" and IPv6
// "[::1]:5000" forms.
func splitHostPort(endpoint string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		return "", 0, fmt.Errorf("invalid endpoint %q: %w", endpoint, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid port in %q: %w", endpoint, err)
	}
	if port <= 0 || port > 65535 {
		return "", 0, fmt.Errorf("port out of range in %q", endpoint)
	}
	return host, port, nil
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
	content, err := os.ReadFile(fileName)
	if err != nil {
		log.Error("failed to read server peer config: %v", err)
	}

	// update
	var peers Peers
	if unmarshalErr := toml.Unmarshal(content, &peers); unmarshalErr != nil {
		log.Error("failed to unmarshal server peer config: %v", unmarshalErr)
	}
	if updateErr := a.updateServerPeers(expandServerPeers(peers.Servers)); updateErr != nil {
		// ignore error
		_ = updateErr
	}

	serverPeerWatch = utils.WatchFile(fileName, func() {
		log.Info("server peer config: %s has been updated", fileName)
		if content, err = a.loadConfigFile(fileName); err == nil {
			if err = toml.Unmarshal(content, &peers); err == nil {
				_ = a.updateServerPeers(expandServerPeers(peers.Servers))
			}
		}
	})

	return nil
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

	serverPeerMap := make(map[string]*core.UdpPeer)
	for _, p := range peers {
		p.Type = core.NHP_SERVER
		a.device.AddPeer(p)
		serverPeerMap[p.PublicKeyBase64()] = p
	}
	a.config.Servers = peers

	// remove old peers from device
	a.serverPeerMutex.Lock()
	defer a.serverPeerMutex.Unlock()
	for pubKey := range a.serverPeerMap {
		if _, found := serverPeerMap[pubKey]; !found {
			a.device.RemovePeer(pubKey)
		}
	}
	a.serverPeerMap = serverPeerMap

	return err
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
	_ = a.updateServerPeers(expandServerPeers(acEtcdConfig.Servers))
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
