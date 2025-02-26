package de

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
	baseConfigWatch     io.Closer
	serverConfigWatch   io.Closer
	resourceConfigWatch io.Closer

	errLoadConfig = fmt.Errorf("config load error")
)

type Config struct {
	LogLevel         int
	PrivateKeyBase64 string
}

type Peers struct {
	Servers []*core.UdpPeer
}

type Resources struct {
	Resources []*KnockResource
}

func (a *UdpDevice) loadBaseConfig() error {
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

func (a *UdpDevice) loadPeers() error {
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

// func (a *UdpDevice) loadResources() error {
// 	// resource.toml
// 	fileName := filepath.Join(ExeDirPath, "etc", "resource.toml")
// 	if err := a.updateResources(fileName); err != nil {
// 		// ignore error
// 		_ = err
// 	}

// 	resourceConfigWatch = utils.WatchFile(fileName, func() {
// 		log.Info("resource config: %s has been updated", fileName)
// 		a.updateResources(fileName)
// 	})

// 	return nil
// }

func (a *UdpDevice) updateBaseConfig(file string) (err error) {
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

	return err
}

func (a *UdpDevice) updateServerPeers(file string) (err error) {
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
		p.Type = core.NHP_DE
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

// func (a *UdpDevice) updateResources(file string) (err error) {
// 	utils.CatchPanicThenRun(func() {
// 		err = errLoadConfig
// 	})

// 	content, err := os.ReadFile(file)
// 	if err != nil {
// 		log.Error("failed to read resource config: %v", err)
// 	}

// 	var resources Resources
// 	targetMap := make(map[string]*KnockTarget)
// 	if err := toml.Unmarshal(content, &resources); err != nil {
// 		log.Error("failed to unmarshal resource config: %v", err)
// 	}
// 	for _, res := range resources.Resources {
// 		peer := a.FindServerPeerFromResource(res)
// 		if peer != nil {
// 			targetMap[res.Id()] = &KnockTarget{
// 				KnockResource: *res,
// 				ServerPeer:    peer,
// 			}
// 		}
// 	}

// 	if a.knockTargetMap == nil {
// 		a.knockTargetMap = targetMap
// 		return err
// 	}

// 	// update
// 	a.knockTargetMapMutex.Lock()
// 	a.knockTargetMap = targetMap
// 	a.knockTargetMapMutex.Unlock()

// 	// renew knock cycle
// 	if len(a.signals.knockTargetMapUpdated) == 0 {
// 		a.signals.knockTargetMapUpdated <- struct{}{}
// 	}

// 	return err
// }

func (a *UdpDevice) StopConfigWatch() {
	if baseConfigWatch != nil {
		baseConfigWatch.Close()
	}
	if serverConfigWatch != nil {
		serverConfigWatch.Close()
	}
	if resourceConfigWatch != nil {
		resourceConfigWatch.Close()
	}
}
