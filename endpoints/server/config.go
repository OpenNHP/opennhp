package server

import (
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

var (
	baseConfigWatch  io.Closer
	httpConfigWatch  io.Closer
	acConfigWatch    io.Closer
	agentConfigWatch io.Closer
	resConfigWatch   io.Closer
	srcipConfigWatch io.Closer
	dbConfigWatch    io.Closer
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
	PrivateKeyBase64       string       `json:"privateKey"`
	Hostname               string       `json:"hostname"`
	ListenIp               string       `json:"listenIp"`
	ListenPort             int          `json:"listenPort"`
	LogLevel               int          `json:"logLevel"`
	DefaultCipherScheme    int          `json:"defaultCipherScheme"`
	DisableAgentValidation bool         `json:"disableAgentValidation"`
	WebRTC                 WebRTCConfig `toml:"webrtc"`
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
}

func (s *UdpServer) loadBaseConfig() error {
	// config.toml
	fileName := filepath.Join(ExeDirPath, "etc", "config.toml")
	content, err := s.loadConfigFile(fileName)
	if err != nil {
		log.Error("load base config err: %v", err)
		return err
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

	//db.toml
	fileNameDE := filepath.Join(ExeDirPath, "etc", "db.toml")
	contentDE, err := s.loadConfigFile(fileNameDE)
	if err != nil {
		log.Error("load db peer config err: %v", err)
		return err
	}
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
		if conf.WebRTC.Enable {
			s.webrtcServer = NewWebRTCServer(s, &conf.WebRTC)
			if err := s.webrtcServer.Start(); err != nil {
				log.Error("failed to start WebRTC server: %v", err)
			}
		}
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

	if s.config.DefaultCipherScheme != conf.DefaultCipherScheme {
		log.Info("set default cipher scheme to %d", conf.DefaultCipherScheme)
		s.config.DefaultCipherScheme = conf.DefaultCipherScheme
	}

	// handle WebRTC configuration change
	if conf.WebRTC.Enable && s.webrtcServer == nil {
		s.webrtcServer = NewWebRTCServer(s, &conf.WebRTC)
		if err := s.webrtcServer.Start(); err != nil {
			log.Error("failed to start WebRTC server: %v", err)
		}
	}
	if !conf.WebRTC.Enable && s.webrtcServer != nil {
		s.webrtcServer.Stop()
		s.webrtcServer = nil
	}
	s.config.WebRTC = conf.WebRTC

	return err
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
