package ac

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/OpenNHP/opennhp/log"
	"github.com/OpenNHP/opennhp/nhp"
	"github.com/OpenNHP/opennhp/utils"

	toml "github.com/pelletier/go-toml/v2"
)

var (
	baseConfigWatch   io.Closer
	serverConfigWatch io.Closer

	errLoadConfig = fmt.Errorf("config load error")
)

type Config struct {
	PrivateKeyBase64 string         `json:"privateKey"`
	ACId             string         `json:"acId"`
	DefaultIp        string         `json:"defaultIp"`
	AuthServiceId    string         `json:"aspId"`
	ResourceIds      []string       `json:"resIds"`
	Servers          []*nhp.UdpPeer `json:"servers"`
	IpPassMode       int            `json:"ipPassMode"` // 0: pass the knock source IP, 1: use pre-access mode and release the access source IP
	LogLevel         int            `json:"logLevel"`
}

type Peers struct {
	Servers []*nhp.UdpPeer
}

func (d *UdpDoor) loadBaseConfig() error {
	// config.toml
	fileName := filepath.Join(ExeDirPath, "etc", "config.toml")
	if err := d.updateBaseConfig(fileName); err != nil {
		// report base config error
		return err
	}

	baseConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("base config: %s has been updated", fileName)
		d.updateBaseConfig(fileName)
	})
	return nil
}

func (d *UdpDoor) loadPeers() error {
	// server.toml
	fileName := filepath.Join(ExeDirPath, "etc", "server.toml")
	if err := d.updateServerPeers(fileName); err != nil {
		// ignore error
		_ = err
	}

	serverConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("server peer config: %s has been updated", fileName)
		d.updateServerPeers(fileName)
	})

	return nil
}

func (d *UdpDoor) updateBaseConfig(file string) (err error) {
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

	if d.config == nil {
		d.config = &conf
		d.log.SetLogLevel(conf.LogLevel)
		return err
	}

	// update
	if d.config.LogLevel != conf.LogLevel {
		log.Info("set base log level to %d", conf.LogLevel)
		d.log.SetLogLevel(conf.LogLevel)
		d.config.LogLevel = conf.LogLevel
	}

	if d.config.DefaultIp != conf.DefaultIp {
		log.Info("set default ip mode to %s", conf.DefaultIp)
		d.config.DefaultIp = conf.DefaultIp
	}

	if d.config.IpPassMode != conf.IpPassMode {
		log.Info("set ip pass mode to %d", conf.IpPassMode)
		d.config.IpPassMode = conf.IpPassMode
	}

	return err
}

func (d *UdpDoor) updateServerPeers(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read server peer config: %v", err)
	}

	// update
	var peers Peers
	serverPeerMap := make(map[string]*nhp.UdpPeer)
	if err := toml.Unmarshal(content, &peers); err != nil {
		log.Error("failed to unmarshal server peer config: %v", err)
	}
	for _, p := range peers.Servers {
		p.Type = nhp.NHP_SERVER
		d.device.AddPeer(p)
		serverPeerMap[p.PublicKeyBase64()] = p
	}

	// remove old peers from device
	d.serverPeerMutex.Lock()
	defer d.serverPeerMutex.Unlock()
	for pubKey := range d.serverPeerMap {
		if _, found := serverPeerMap[pubKey]; !found {
			d.device.RemovePeer(pubKey)
		}
	}
	d.serverPeerMap = serverPeerMap

	return err
}

func (d *UdpDoor) IpPassMode() int {
	return d.config.IpPassMode
}

func (d *UdpDoor) StopConfigWatch() {
	if baseConfigWatch != nil {
		baseConfigWatch.Close()
	}
	if serverConfigWatch != nil {
		serverConfigWatch.Close()
	}
}
