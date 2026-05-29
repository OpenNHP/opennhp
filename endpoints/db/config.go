package db

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/common/clusterconfig"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/utils"
)

var (
	baseConfigWatch     io.Closer
	serverConfigWatch   io.Closer
	teesConfigWatch     io.Closer
	resourceConfigWatch io.Closer

	errLoadConfig = fmt.Errorf("config load error")
)

type Config struct {
	LogLevel            int
	PrivateKeyBase64    string
	DefaultCipherScheme int    `json:"defaultCipherScheme"`
	SymmetricCipherMode string `json:"symmetricCipherMode"`
	DbId                string `json:"dbId"`
}

// Peers is the top-level shape of server.toml. Each entry is one
// logical nhp-server identity (one pubkey) reachable at 1..N instances
// — same schema as nhp-agent's and nhp-ac's server.toml (see
// nhp/common/clusterconfig). nhp-db only ever talks to a single
// instance per cluster today, so LoadBalance / StickyInstance are
// loaded but ignored; the schema is shared so operators don't need a
// per-daemon dialect.
type Peers struct {
	Servers []*clusterconfig.ClusterConfig
}

type Resources struct {
	Resources []*KnockResource
}

type TEE struct {
	TEEPublicKeyBase64 string `json:"teePublicKeyBase64"`
	ExpireTime         int64  `json:"expireTime"`
}

type TEEs struct {
	TEEs []*TEE
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
		_ = a.updateBaseConfig(fileName)
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
		_ = a.updateServerPeers(fileName)
	})

	return nil
}

func (a *UdpDevice) loadTEEs() error {
	// consumer.toml
	fileName := filepath.Join(ExeDirPath, "etc", "tee.toml")
	if err := a.updateTEEConfig(fileName); err != nil {
		// ignore error
		_ = err
	}

	teesConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("tee peer config: %s has been updated", fileName)
		_ = a.updateTEEConfig(fileName)
	})

	return nil
}

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

	if a.config.DefaultCipherScheme != conf.DefaultCipherScheme {
		log.Info("set default cipher scheme to %d", conf.DefaultCipherScheme)
		a.config.DefaultCipherScheme = conf.DefaultCipherScheme
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
	if err := toml.Unmarshal(content, &peers); err != nil {
		log.Error("failed to unmarshal server config: %v", err)
	}
	// Normalize first so legacy single-server entries (Ip/Port at the
	// top level) auto-upgrade to a single-instance cluster. nhp-db
	// only ever picks one instance per cluster, so LoadBalance and
	// StickyInstance are loaded but unused — the shared schema lets
	// nhp-db, nhp-ac, and nhp-agent take the same TOML.
	if nerr := clusterconfig.Normalize(peers.Servers, clusterconfig.Options{
		ConsumerLabel: "db",
		RequireName:   false,
	}, log.Warning); nerr != nil {
		log.Error("invalid server.toml: %v", nerr)
		return nerr
	}

	serverPeerMap := make(map[string]*core.UdpPeer)
	for _, c := range peers.Servers {
		// One UdpPeer per (pubkey, address). nhp-db's serverPeerMap is
		// keyed by pubkey alone — if a cluster declares multiple
		// instances, the later one wins. That's fine: nhp-db only
		// needs one reachable address per identity, and the others
		// stay as failover candidates the operator can swap in.
		for _, inst := range c.Instances {
			p := &core.UdpPeer{
				PubKeyBase64: c.PubKeyBase64,
				Hostname:     inst.Host,
				Ip:           inst.Ip,
				Port:         inst.Port,
				ExpireTime:   c.ExpireTime,
				Type:         core.NHP_DB,
			}
			a.device.AddPeer(p)
			serverPeerMap[p.PublicKeyBase64()] = p
		}
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

func (a *UdpDevice) updateTEEConfig(file string) (err error) {
	utils.CatchPanicThenRun(func() {
		err = errLoadConfig
	})

	content, err := os.ReadFile(file)
	if err != nil {
		log.Error("failed to read TEE config: %v", err)
	}

	var tees TEEs
	if err := toml.Unmarshal(content, &tees); err != nil {
		log.Error("failed to unmarshal TEE config: %v", err)
	}

	teeMap := make(map[string]*TEE)
	for _, tee := range tees.TEEs {
		teeMap[tee.TEEPublicKeyBase64] = tee
	}

	a.teeMutex.Lock()
	defer a.teeMutex.Unlock()
	a.teeMap = teeMap

	return nil
}

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
	if teesConfigWatch != nil {
		teesConfigWatch.Close()
	}
}
