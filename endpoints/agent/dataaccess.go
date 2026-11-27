package agent

import (
	"sync"
	"time"

	wasmEngine "github.com/OpenNHP/opennhp/nhp/core/wasm/engine"
)

// cachedEngine is a smart-policy WASM engine plus the bookkeeping that
// keeps it alive while a TA call is still inside it. A retired engine is
// already unreachable from dataAccessCache.engines; it is closed by
// whoever drops the last reference, so an engine replaced mid-call is
// never closed out from under the call.
type cachedEngine struct {
	engine *wasmEngine.Engine

	// inUse and retired are guarded by the owning dataAccessCache's mu.
	inUse   int
	retired bool
}

// retire marks ce unreachable and reports whether the caller must close
// it (nobody is inside it). The caller holds the cache's write lock, and
// must Close outside it: wazero's Runtime.Close tears down the module.
func (ce *cachedEngine) retire() bool {
	ce.retired = true
	return ce.inUse == 0
}

// dataAccessCache holds the per-data-object DHP state that a confidential
// computing call reads and a smart-data-policy refresh writes: when each
// object's policy was last refreshed, where its decrypted plaintext was
// written, which policy it is bound to, and the WASM engine each policy
// loaded.
//
// Every one of those used to be a bare map on UdpAgent, read by
// StartConfidentialComputing outside any lock and written by
// RefreshDataAccess / SendDARMsgToServer under dataAccessRefreshMutex.
// The DHP console hands each request to its own goroutine, so two
// overlapping TA calls on different data objects would read and write the
// same map concurrently. In Go that is not a torn value but a runtime
// "concurrent map read and map write" fatal error: it is not a panic,
// gin.Recovery() cannot catch it, and the agent process exits. (The
// double-checked "check, lock, check again" in StartConfidentialComputing
// does not help — the first check is itself the unsynchronized read.)
//
// So the maps live here and are only reachable through the accessors
// below, which take mu. dataAccessRefreshMutex keeps its separate job:
// serializing the refresh itself so one object is not fetched and
// decrypted twice at once. Guarding the maps is this lock's job alone.
type dataAccessCache struct {
	mu sync.RWMutex

	refreshTime map[string]int64         // data object id -> last smart data policy refresh, UnixNano
	decrypted   map[string]string        // data object id -> decrypted plaintext path
	policyId    map[string]string        // data object id -> smart data policy id
	engines     map[string]*cachedEngine // smart data policy id -> loaded policy engine
}

// reset empties the cache. Start calls it on every (re)start, so a
// RestartAgent does not inherit the stopped instance's decrypted paths —
// and, unlike the "assign four fresh maps" it replaces, does not leak the
// wazero runtimes those maps were holding.
func (c *dataAccessCache) reset() {
	c.mu.Lock()
	var retired []*cachedEngine
	for policyId, ce := range c.engines {
		delete(c.engines, policyId)
		if ce.retire() {
			retired = append(retired, ce)
		}
	}
	c.refreshTime = make(map[string]int64)
	c.decrypted = make(map[string]string)
	c.policyId = make(map[string]string)
	c.engines = make(map[string]*cachedEngine)
	c.mu.Unlock()

	for _, ce := range retired {
		ce.engine.Close()
	}
}

// preCheck reports whether this data object's smart data policy is stale
// (or was never fetched) and whether its plaintext is already on disk.
// An object that has never been decrypted always needs a refresh,
// whatever its policy timestamp says.
func (c *dataAccessCache) preCheck(ztdoId string) (output string, refreshSdp bool, decrypted bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if sdpRefreshTime, exist := c.refreshTime[ztdoId]; exist {
		if time.Now().UnixNano()-sdpRefreshTime > SmartDataPolicyRefreshTime {
			refreshSdp = true
		}
	} else {
		refreshSdp = true
	}

	if plaintextPath, exist := c.decrypted[ztdoId]; exist {
		output = plaintextPath
		decrypted = true
	} else {
		decrypted = false
		refreshSdp = true
	}

	return output, refreshSdp, decrypted
}

// markRefreshed records that this object's smart data policy was just
// fetched from the server.
func (c *dataAccessCache) markRefreshed(ztdoId string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.refreshTime == nil {
		c.refreshTime = make(map[string]int64)
	}
	c.refreshTime[ztdoId] = time.Now().UnixNano()
}

// setDecrypted records where this object's plaintext was written.
func (c *dataAccessCache) setDecrypted(ztdoId string, plaintextPath string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.decrypted == nil {
		c.decrypted = make(map[string]string)
	}
	c.decrypted[ztdoId] = plaintextPath
}

// policyFor returns the smart data policy id bound to a data object.
func (c *dataAccessCache) policyFor(ztdoId string) (policyId string, exist bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	policyId, exist = c.policyId[ztdoId]
	return policyId, exist
}

// setPolicy binds a data object to the policy the server just issued for
// it, dropping the engine its previous policy had loaded.
func (c *dataAccessCache) setPolicy(doId string, policyId string) {
	c.mu.Lock()
	if c.policyId == nil {
		c.policyId = make(map[string]string)
	}
	var retired *cachedEngine
	if oldPolicyId, exist := c.policyId[doId]; exist {
		if ce, exist := c.engines[oldPolicyId]; exist {
			delete(c.engines, oldPolicyId)
			if ce.retire() {
				retired = ce
			}
		}
		delete(c.policyId, doId)
	}
	c.policyId[doId] = policyId
	c.mu.Unlock()

	if retired != nil {
		retired.engine.Close()
	}
}

// putEngine stores the engine a policy's WASM was just loaded into,
// retiring any engine the same policy id already had.
func (c *dataAccessCache) putEngine(policyId string, engine *wasmEngine.Engine) {
	c.mu.Lock()
	if c.engines == nil {
		c.engines = make(map[string]*cachedEngine)
	}
	var retired *cachedEngine
	if old, exist := c.engines[policyId]; exist {
		if old.retire() {
			retired = old
		}
	}
	c.engines[policyId] = &cachedEngine{engine: engine}
	c.mu.Unlock()

	if retired != nil {
		retired.engine.Close()
	}
}

// acquireEngine returns a policy's engine and pins it until the matching
// releaseEngine: a concurrent refresh may retire it meanwhile, but will
// not close it. Calling into a closed wazero runtime panics inside the
// engine's own Call wrappers, which on the console's TA path costs the
// caller its result and leaves a recovered 500 in the log.
func (c *dataAccessCache) acquireEngine(policyId string) (*cachedEngine, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ce, exist := c.engines[policyId]
	// Retiring and removing from the map happen together under mu, so a
	// retired engine should never be found here; check anyway rather than
	// hand out an engine that is about to be closed.
	if !exist || ce.retired {
		return nil, false
	}
	ce.inUse++
	return ce, true
}

// releaseEngine drops a reference taken by acquireEngine, closing the
// engine if it was retired while in use.
func (c *dataAccessCache) releaseEngine(ce *cachedEngine) {
	c.mu.Lock()
	ce.inUse--
	closeNow := ce.retired && ce.inUse == 0
	c.mu.Unlock()

	if closeNow {
		ce.engine.Close()
	}
}
