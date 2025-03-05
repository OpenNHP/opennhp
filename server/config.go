package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/OpenNHP/opennhp/common"
	"github.com/OpenNHP/opennhp/core"
	"github.com/OpenNHP/opennhp/log"
	"github.com/OpenNHP/opennhp/plugins"
	"github.com/OpenNHP/opennhp/utils"

	toml "github.com/pelletier/go-toml/v2"
)

var (
	baseConfigWatch  io.Closer
	httpConfigWatch  io.Closer
	acConfigWatch    io.Closer
	agentConfigWatch io.Closer
	resConfigWatch   io.Closer
	srcipConfigWatch io.Closer
	deConfigWatch    io.Closer
	errLoadConfig    = fmt.Errorf("config load error")
)

type Config struct {
	PrivateKeyBase64       string `json:"privateKey"`
	ListenIp               string `json:"listenIp"`
	ListenPort             int    `json:"listenPort"`
	LogLevel               int    `json:"logLevel"`
	Hostname               string `json:"hostname"`
	DisableAgentValidation bool   `json:"disableAgentValidation"`
}

type HttpConfig struct {
	EnableHttp   bool
	EnableTLS    bool
	HttpListenIp string
	TLSCertFile  string
	TLSKeyFile   string
}

type Peers struct {
	ACs    []*core.UdpPeer
	Agents []*core.UdpPeer
	DEs    []*core.UdpPeer
}

func (s *UdpServer) loadBaseConfig() error {
	// config.toml
	fileName := filepath.Join(ExeDirPath, "etc", "config.toml")
	if err := s.updateBaseConfig(fileName); err != nil {
		// report base config error
		return err
	}

	baseConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("base config: %s has been updated", fileName)
		s.updateBaseConfig(fileName)
	})
	return nil
}

func (s *UdpServer) loadHttpConfig() error {
	// http.toml
	fileName := filepath.Join(ExeDirPath, "etc", "http.toml")
	if err := s.updateHttpConfig(fileName); err != nil {
		// ignore error
		_ = err
	}

	httpConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("http config: %s has been updated", fileName)
		s.updateHttpConfig(fileName)
	})
	return nil
}

func (s *UdpServer) loadPeers() error {
	// ac.toml
	fileNameAC := filepath.Join(ExeDirPath, "etc", "ac.toml")
	if err := s.updateACPeers(fileNameAC); err != nil {
		// ignore error
		_ = err
	}

	acConfigWatch = utils.WatchFile(fileNameAC, func() {
		log.Info("ac peer config: %s has been updated", fileNameAC)
		s.updateACPeers(fileNameAC)
	})

	// agent.toml
	fileNameAgent := filepath.Join(ExeDirPath, "etc", "agent.toml")
	if err := s.updateAgentPeers(fileNameAgent); err != nil {
		// ignore error
		_ = err
	}

	agentConfigWatch = utils.WatchFile(fileNameAgent, func() {
		log.Info("agent peer config: %s has been updated", fileNameAgent)
		s.updateAgentPeers(fileNameAgent)
	})

	//de.toml
	fileNameDE := filepath.Join(ExeDirPath, "etc", "de.toml")
	if err := s.updateDePeers(fileNameDE); err != nil {
		// ignore error
		_ = err
	}
	deConfigWatch = utils.WatchFile(fileNameDE, func() {
		log.Info("device peer config: %s has been updated", fileNameDE)
		s.updateDePeers(fileNameDE)
	})
	return nil
}

func (s *UdpServer) loadResources() error {
	// resource.toml
	fileName := filepath.Join(ExeDirPath, "etc", "resource.toml")
	if err := s.updateResources(fileName); err != nil {
		// ignore error
		_ = err
	}

	resConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("resource config: %s has been updated", fileName)
		s.updateResources(fileName)
	})
	return nil
}

func (s *UdpServer) loadSourceIps() error {
	// srcip.toml
	fileName := filepath.Join(ExeDirPath, "etc", "srcip.toml")
	if err := s.updateSourceIps(fileName); err != nil {
		// ignore error
		_ = err
	}

	srcipConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("src ip config: %s has been updated", fileName)
		s.updateSourceIps(fileName)
	})
	return nil
}

func (s *UdpServer) updateBaseConfig(file string) (err error) {
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

	return err
}

func (s *UdpServer) updateHttpConfig(file string) (err error) {
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

func (s *UdpServer) updateACPeers(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read ac peer config: %v", err)
	}

	// update
	var peers Peers
	acPeerMap := make(map[string]*core.UdpPeer)
	if err := toml.Unmarshal(content, &peers); err != nil {
		log.Error("failed to unmarshal ac peer config: %v", err)
	}
	for _, p := range peers.ACs {
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

func (s *UdpServer) updateAgentPeers(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read agent peer config: %v", err)
	}

	var peers Peers
	agentPeerMap := make(map[string]*core.UdpPeer)
	if err := toml.Unmarshal(content, &peers); err != nil {
		log.Error("failed to unmarshal agent peer config: %v", err)
	}
	for _, p := range peers.Agents {
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

func (s *UdpServer) updateResources(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read resource config: %v", err)
	}

	// update
	aspMap := make(common.AuthSvcProviderMap)
	if err := toml.Unmarshal(content, &aspMap); err != nil {
		log.Error("failed to unmarshal resource config: %v", err)
	}

	for aspId, aspData := range aspMap {
		aspData.AuthSvcId = aspId
		if len(aspData.PluginPath) > 0 {
			h := plugins.ReadPluginHandler(aspData.PluginPath)
			if h != nil {
				s.LoadPlugin(aspId, h)
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

func (s *UdpServer) updateSourceIps(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read src ip config: %v", err)
	}

	// update
	srcIpMap := make(map[string][]*common.NetAddress)
	if err := toml.Unmarshal(content, &srcIpMap); err != nil {
		log.Error("failed to unmarshal src ip config: %v", err)
	}

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
	//add deConfigWatch
	if deConfigWatch != nil {
		deConfigWatch.Close()
	}

}

// updateDePeers
func (s *UdpServer) updateDePeers(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read device peer config: %v", err)
	}

	var peers Peers
	dePeerMap := make(map[string]*core.UdpPeer)
	if err := toml.Unmarshal(content, &peers); err != nil {
		log.Error("failed to unmarshal device peer config: %v", err)
	}
	for _, p := range peers.DEs {
		p.Type = core.NHP_DE
		s.device.AddPeer(p)
		dePeerMap[p.PublicKeyBase64()] = p
	}

	// remove old peers from device
	s.dePeerMapMutex.Lock()
	defer s.dePeerMapMutex.Unlock()
	for pubKey := range s.dePeerMap {
		if _, found := dePeerMap[pubKey]; !found {
			s.device.RemovePeer(pubKey)
		}
	}
	s.dePeerMap = dePeerMap
	return err
}
