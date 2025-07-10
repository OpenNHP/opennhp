package ac

import (
	"fmt"
	"github.com/OpenNHP/opennhp/nhp/etcd"
	"io"
	"os"
	"path/filepath"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/utils"
	toml "github.com/pelletier/go-toml/v2"
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
	Servers    []*core.UdpPeer
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

type HttpConfig struct {
	EnableHttp     bool
	EnableTLS      bool
	HttpListenPort int
	TLSCertFile    string
	TLSKeyFile     string
}

type Peers struct {
	Servers []*core.UdpPeer
}

func (a *UdpAC) loadBaseConfig() error {
	// config.toml
	fileName := filepath.Join(ExeDirPath, "etc", "config.toml")
	content, err := os.ReadFile(fileName)
	if err != nil {
		log.Error("failed to read base config: %v", err)
	}

	var conf Config
	if err := toml.Unmarshal(content, &conf); err != nil {
		log.Error("failed to unmarshal base config: %v", err)
	}
	if err := a.updateBaseConfig(conf); err != nil {
		// report base config error
		return err
	}

	baseConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("base config: %s has been updated", fileName)
		if content, err = a.loadConfigFile(fileName); err == nil {
			if err = toml.Unmarshal(content, &conf); err == nil {
				a.updateBaseConfig(conf)
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
	if err := toml.Unmarshal(content, &httpConf); err != nil {
		log.Error("failed to unmarshal http config: %v", err)
	}
	if err := a.updateHttpConfig(httpConf); err != nil {
		// ignore error
		_ = err
	}

	httpConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("http config: %s has been updated", fileName)
		if content, err = a.loadConfigFile(fileName); err == nil {
			if err = toml.Unmarshal(content, &httpConf); err == nil {
				a.updateHttpConfig(httpConf)
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
	if err := toml.Unmarshal(content, &peers); err != nil {
		log.Error("failed to unmarshal server peer config: %v", err)
	}
	if err := a.updateServerPeers(peers.Servers); err != nil {
		// ignore error
		_ = err
	}

	serverPeerWatch = utils.WatchFile(fileName, func() {
		log.Info("server peer config: %s has been updated", fileName)
		if content, err = a.loadConfigFile(fileName); err == nil {
			if err = toml.Unmarshal(content, &peers); err == nil {
				a.updateServerPeers(peers.Servers)
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
func (a *UdpAC) initEtcdClient() error {
	// etcd.toml
	fileName := filepath.Join(ExeDirPath, "etc", "remote.toml")

	_, e := os.Stat(fileName)
	if os.IsNotExist(e) {
		//etcd.toml file not found,use local config
		return nil
	}

	content, err := os.ReadFile(fileName)
	if err != nil {
		log.Error("failed to read remote config: %v", err)
		return err
	}

	var conf etcd.EtcdConfig
	if err = toml.Unmarshal(content, &conf); err != nil {
		log.Error("failed to unmarshal remote config: %v", err)
		return err
	}

	if len(conf.Endpoints) == 0 {
		log.Error("remote config has no endpoints,open nhp ac will startup with local configuration")
		return nil
	}

	if len(conf.Key) == 0 {
		log.Error("remote config has no key,open nhp ac will startup with local configuration")
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
		a.updateEtcdConfig(val, true)
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
		a.updateBaseConfig(acEtcdConfig.BaseConfig)
	}
	a.updateHttpConfig(acEtcdConfig.HttpConfig)
	a.updateServerPeers(acEtcdConfig.Servers)
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
