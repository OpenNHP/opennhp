package ac

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/OpenNHP/opennhp/core"
	"github.com/OpenNHP/opennhp/log"
	"github.com/OpenNHP/opennhp/utils"

	toml "github.com/pelletier/go-toml/v2"
)

var (
	baseConfigWatch io.Closer
	httpConfigWatch io.Closer
	serverPeerWatch io.Closer

	errLoadConfig = fmt.Errorf("config load error")
)

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
	if err := a.updateBaseConfig(fileName); err != nil {
		// report base config error
		return err
	}

	baseConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("base config: %s has been updated", fileName)
		a.updateBaseConfig(fileName)
	})
	return nil
}

func (a *UdpAC) loadHttpConfig() error {
	// http.toml
	fileName := filepath.Join(ExeDirPath, "etc", "http.toml")
	if err := a.updateHttpConfig(fileName); err != nil {
		// ignore error
		_ = err
	}

	httpConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("http config: %s has been updated", fileName)
		a.updateHttpConfig(fileName)
	})
	return nil
}

func (a *UdpAC) loadPeers() error {
	// server.toml
	fileName := filepath.Join(ExeDirPath, "etc", "server.toml")
	if err := a.updateServerPeers(fileName); err != nil {
		// ignore error
		_ = err
	}

	serverPeerWatch = utils.WatchFile(fileName, func() {
		log.Info("server peer config: %s has been updated", fileName)
		a.updateServerPeers(fileName)
	})

	return nil
}

func (a *UdpAC) updateBaseConfig(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read base config: %v", err)
	}

	var conf Config
	if err := toml.Unmarshal(content, &conf); err != nil {
		log.Error("failed to unmarshal base config: %v", err)
	}

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

func (a *UdpAC) updateHttpConfig(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read http config: %v", err)
	}

	var httpConf HttpConfig
	if err := toml.Unmarshal(content, &httpConf); err != nil {
		log.Error("failed to unmarshal http config: %v", err)
	}

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

func (a *UdpAC) updateServerPeers(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read server peer config: %v", err)
	}

	// update
	var peers Peers
	serverPeerMap := make(map[string]*core.UdpPeer)
	if err := toml.Unmarshal(content, &peers); err != nil {
		log.Error("failed to unmarshal server peer config: %v", err)
	}
	for _, p := range peers.Servers {
		p.Type = core.NHP_SERVER
		a.device.AddPeer(p)
		serverPeerMap[p.PublicKeyBase64()] = p
	}

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
