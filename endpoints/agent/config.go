package agent

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/common"
	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/log"
	"github.com/OpenNHP/opennhp/nhp/utils"
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

// Peers is the top-level shape of server.toml. Each entry is one
// logical nhp-server identity (one pubkey) that may be backed by 1..N
// physical Instances. See ClusterConfig for both the cluster form and
// the auto-upgraded legacy flat form.
type Peers struct {
	Servers []*ClusterConfig
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
		_ = a.updateBaseConfig(fileName)
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
		_ = a.updateDHPConfig(fileName)
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
		_ = a.updateServerPeers(fileName)
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
		_ = a.updateResources(fileName)
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
		return err
	}

	var peers Peers
	if err := toml.Unmarshal(content, &peers); err != nil {
		log.Error("failed to unmarshal server config: %v", err)
		return err
	}
	// Validate + auto-upgrade legacy flat form to cluster form.
	// Deprecation warnings go through log.Warning so operators see
	// them in the running daemon's logs, not just at first load.
	if err := normalizeClusters(peers.Servers, log.Warning); err != nil {
		log.Error("invalid server.toml: %v", err)
		return err
	}

	newMap := make(map[string]*ServerCluster, len(peers.Servers))
	newByName := make(map[string]*ServerCluster, len(peers.Servers))
	for _, cfg := range peers.Servers {
		sc, berr := buildCluster(cfg)
		if berr != nil {
			log.Error("buildCluster for %s failed: %v", cfg.PubKeyBase64, berr)
			return berr
		}
		a.device.AddPeer(sc.representativePeer)
		newMap[sc.PublicKeyBase64] = sc
		newByName[sc.Name] = sc
	}

	// Swap atomically, then drop peers from device that no longer
	// appear in the new config. Removal-after-swap avoids a brief
	// window where an inbound packet could be matched against a
	// freshly-removed peer that's still in serverClusterMap (or
	// vice versa).
	a.serverPeerMutex.Lock()
	old := a.serverClusterMap
	a.serverClusterMap = newMap
	a.serverClusterByName = newByName
	a.serverPeerMutex.Unlock()

	for pubKey := range old {
		if _, found := newMap[pubKey]; !found {
			a.device.RemovePeer(pubKey)
		}
	}

	// Re-bind any in-flight KnockTargets onto the freshly-loaded
	// clusters. Without this, a server.toml-only reload leaves each
	// KnockTarget holding a stale *ServerCluster (no longer reachable
	// from serverClusterMap) and a stale *UdpPeer (already removed
	// from device.peerMap above), and the next handshake fails the
	// device pubkey lookup. PickInstance's "instance survived the
	// reload" guard only catches in-cluster shrinkage; it cannot
	// detect that the entire cluster object was replaced. On a fresh
	// agent startup loadPeers runs before loadResources, so
	// knockTargetMap is still nil — skip in that case.
	a.refreshKnockTargetClusters()

	return err
}

// refreshKnockTargetClusters re-resolves every KnockTarget's cluster
// pointer after a server.toml reload and signals the knock-cycle
// goroutine to restart its sub-routines so they observe the new
// state. Targets whose Cluster / ServerPubKey reference no longer
// resolves are removed from knockTargetMap so the knock loop won't
// keep retrying against vanished servers.
func (a *UdpAgent) refreshKnockTargetClusters() {
	a.knockTargetMapMutex.Lock()
	if a.knockTargetMap == nil {
		a.knockTargetMapMutex.Unlock()
		return
	}
	// Take a snapshot of the targets we need to revisit; release the
	// map mutex before calling FindServerClusterFromResource because
	// that helper takes serverPeerMutex and we want to keep these
	// locks independent.
	type pending struct {
		id     string
		target *KnockTarget
	}
	snapshot := make([]pending, 0, len(a.knockTargetMap))
	for id, tgt := range a.knockTargetMap {
		snapshot = append(snapshot, pending{id: id, target: tgt})
	}
	a.knockTargetMapMutex.Unlock()

	type removal struct{ id string }
	var toRemove []removal
	for _, p := range snapshot {
		res := &p.target.KnockResource
		sc, ferr := a.FindServerClusterFromResource(res)
		if sc == nil {
			log.Error("server.toml reload removed cluster for resource %s: %v", p.id, ferr)
			toRemove = append(toRemove, removal{id: p.id})
			continue
		}
		// SetServerCluster resets the sticky pin when the cluster
		// pointer changes, which is exactly what we want here even
		// when the cluster name is the same — every *ServerCluster
		// and its instances were replaced by the swap above.
		p.target.SetServerCluster(sc)
		p.target.SetServerPeer(sc.representativePeer)
	}

	if len(toRemove) > 0 {
		a.knockTargetMapMutex.Lock()
		for _, r := range toRemove {
			delete(a.knockTargetMap, r.id)
		}
		a.knockTargetMapMutex.Unlock()
	}

	// Wake knockResourceRoutine so it tears down its per-target
	// sub-routines and restarts them with the refreshed map. Without
	// this, sub-routines that captured the *KnockTarget keep using
	// the stale chosenInstance until their next natural cycle, and
	// removed entries continue to be knocked from old sub-routines.
	// The signal channel is buffered (size 1); a non-blocking send
	// preserves the existing coalescing semantics — if a previous
	// update is already queued, this one folds into it.
	select {
	case a.signals.knockTargetMapUpdated <- struct{}{}:
	default:
	}
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
		sc, err := a.FindServerClusterFromResource(res)
		if sc == nil {
			// err is already specific (the function logs the
			// failure mode); skip this entry and move on so a
			// single bad resource doesn't block the rest.
			log.Error("skipping resource %s: %v", res.Id(), err)
			continue
		}
		targetMap[res.Id()] = &KnockTarget{
			KnockResource: *res,
			ServerPeer:    sc.representativePeer,
			ServerCluster: sc,
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

	// renew knock cycle. Non-blocking send (matches
	// refreshKnockTargetClusters): a debounced file-watcher callback can
	// fire after Stop(), and the previous blocking send guarded only by
	// a len()==0 check both raced (two watchers could both observe 0 and
	// the second would block) and risked blocking forever once the
	// consumer routine had exited. The non-blocking select coalesces
	// duplicate updates into the size-1 buffer and never blocks.
	// knockTargetMapUpdated is intentionally never closed (see Stop()),
	// so a late send after shutdown lands in the buffer rather than
	// panicking.
	select {
	case a.signals.knockTargetMapUpdated <- struct{}{}:
	default:
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
