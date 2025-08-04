package agent

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/utils"
	toml "github.com/pelletier/go-toml/v2"
)

var (
	baseConfigWatch     io.Closer
	dhpConfigWatch      io.Closer
	serverConfigWatch   io.Closer
	resourceConfigWatch io.Closer

	errLoadConfig = fmt.Errorf("config load error")

	secretCreated = "/var/run/secret.created"
)

type Config struct {
	LogLevel            int    `json:"logLevel"`
	DefaultCipherScheme int    `json:"defaultCipherScheme"`
	PrivateKeyBase64    string `json:"privateKey"`
	KnockUser           `mapstructure:",squash"`
	*DHPConfig
}

type DHPConfig struct {
	TEEPrivateKeyBase64 string `json:"teePrivateKeyBase64"`
}

func (c *Config) GetAgentEcdh() core.Ecdh {
	eccType := core.ECC_SM2
	if c.DefaultCipherScheme == common.CIPHER_SCHEME_CURVE {
		eccType = core.ECC_CURVE25519
	}
	teePrk, _ := base64.StdEncoding.DecodeString(c.PrivateKeyBase64)
	return core.ECDHFromKey(eccType, teePrk)
}

func (c *Config) GetTeeEcdh() core.Ecdh {
	eccType := core.ECC_SM2
	if c.DefaultCipherScheme == common.CIPHER_SCHEME_CURVE {
		eccType = core.ECC_CURVE25519
	}
	teePrk, _ := base64.StdEncoding.DecodeString(c.TEEPrivateKeyBase64)
	return core.ECDHFromKey(eccType, teePrk)
}

func (c *Config) GetEccType() core.EccTypeEnum {
	eccType := core.ECC_SM2
	if c.DefaultCipherScheme == common.CIPHER_SCHEME_CURVE {
		eccType = core.ECC_CURVE25519
	}
	return eccType
}

type Peers struct {
	Servers []*core.UdpPeer
}

type Resources struct {
	Resources []*KnockResource
}

func (a *UdpAgent) loadBaseConfig() error {
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

func (a *UdpAgent) loadDHPConfig() error {
	// dhp.toml
	fileName := filepath.Join(ExeDirPath, "etc", "dhp.toml")
	if err := a.updateDHPConfig(fileName); err != nil {
		// ignore error
		_ = err
	}

	dhpConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("DHP config: %s has been updated", fileName)
		a.updateDHPConfig(fileName)
	})

	return nil
}

func (a *UdpAgent) loadPeers() error {
	// server.toml
	fileName := filepath.Join(ExeDirPath, "etc", "server.toml")
	if err := a.updateServerPeers(fileName); err != nil {
		// ignore error
		_ = err
	}

	serverConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("server peer config: %s has been updated", fileName)
		a.updateServerPeers(fileName)
	})

	return nil
}

func (a *UdpAgent) loadResources() error {
	// resource.toml
	fileName := filepath.Join(ExeDirPath, "etc", "resource.toml")
	if err := a.updateResources(fileName); err != nil {
		// ignore error
		_ = err
	}

	resourceConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("resource config: %s has been updated", fileName)
		a.updateResources(fileName)
	})

	return nil
}

func (a *UdpAgent) updateBaseConfig(file string) (err error) {
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

	a.knockUserMutex.Lock()
	a.knockUser = &KnockUser{
		UserId:         conf.UserId,
		OrganizationId: conf.OrganizationId,
		UserData:       conf.UserData,
	}
	a.knockUserMutex.Unlock()

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

	if a.config.DefaultCipherScheme != conf.DefaultCipherScheme {
		log.Info("set default cipher scheme to %d", conf.DefaultCipherScheme)
		a.config.DefaultCipherScheme = conf.DefaultCipherScheme
	}

	return err
}

func (a *UdpAgent) updateDHPConfig(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read DHP config: %v", err)
	}

	var conf DHPConfig
	if err := toml.Unmarshal(content, &conf); err != nil {
		log.Error("failed to unmarshal DHP config: %v", err)
	}

	if a.config.DHPConfig == nil {
		a.config.DHPConfig = &conf
		return err
	}

	return err
}

func (a *UdpAgent) updateServerPeers(file string) (err error) {
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
		log.Error("failed to unmarshal server config: %v", err)
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

func (a *UdpAgent) updateResources(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read resource config: %v", err)
	}

	var resources Resources
	targetMap := make(map[string]*KnockTarget)
	if err := toml.Unmarshal(content, &resources); err != nil {
		log.Error("failed to unmarshal resource config: %v", err)
	}
	for _, res := range resources.Resources {
		peer := a.FindServerPeerFromResource(res)
		if peer == nil {
			log.Error("failed to find corresponding server peer for resource %s", res.Id())
			continue
		}
		targetMap[res.Id()] = &KnockTarget{
			KnockResource: *res,
			ServerPeer:    peer,
		}
	}

	if a.knockTargetMap == nil {
		a.knockTargetMap = targetMap
		return err
	}

	// update
	a.knockTargetMapMutex.Lock()
	a.knockTargetMap = targetMap
	a.knockTargetMapMutex.Unlock()

	// renew knock cycle
	if len(a.signals.knockTargetMapUpdated) == 0 {
		a.signals.knockTargetMapUpdated <- struct{}{}
	}

	return err
}

func (a *UdpAgent) StopConfigWatch() {
	if baseConfigWatch != nil {
		baseConfigWatch.Close()
	}
	if dhpConfigWatch != nil {
		dhpConfigWatch.Close()
	}
	if serverConfigWatch != nil {
		serverConfigWatch.Close()
	}
	if resourceConfigWatch != nil {
		resourceConfigWatch.Close()
	}
}

func (a *UdpAgent) NewEcdhFromConfigFile() (core.Ecdh, error) {
	fileName := filepath.Join(ExeDirPath, "etc", "config.toml")

	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var conf Config
	if err := toml.Unmarshal(content, &conf); err != nil {
		return nil, err
	}

	return core.NewECDH(conf.GetEccType()), nil
}

func (a *UdpAgent) RotateTeeKey() error {
	fileName := filepath.Join(ExeDirPath, "etc", "dhp.toml")

	ecdh, err := a.NewEcdhFromConfigFile()
	if err != nil {
		return err
	}

	if err := utils.UpdateTomlConfig(fileName, "TEEPrivateKeyBase64", ecdh.PrivateKeyBase64()); err != nil {
		return err
	}

	return nil
}

func (a *UdpAgent) RotateAgentKey() error {
	fileName := filepath.Join(ExeDirPath, "etc", "config.toml")

	ecdh, err := a.NewEcdhFromConfigFile()
	if err != nil {
		return err
	}

	if err := utils.UpdateTomlConfig(fileName, "PrivateKeyBase64", ecdh.PrivateKeyBase64()); err != nil {
		return err
	}

	return nil
}

func (a *UdpAgent) InitializeSecret() error {
	if _, err := os.Stat(secretCreated); os.IsNotExist(err) {
		err := a.RotateAgentKey()
		if err != nil {
			return err
		}
		err = a.RotateTeeKey()
		if err != nil {
			return err
		}

		_, err = os.Create(secretCreated)
		if err != nil {
			return err
		}
	}

	return nil
}
