package agent

import (
	"fmt"
	"sync"
	"testing"

	wasmEngine "github.com/OpenNHP/opennhp/nhp/core/wasm/engine"
)

// TestStartConfidentialComputingConcurrentCacheAccess fences the DHP
// console's request path against its own refresh path.
//
// The console gives every request its own goroutine, so two confidential
// computing calls overlap by construction. The request path reads the
// per-data-object caches (refresh time, decrypted path, policy id, policy
// engine) without taking dataAccessRefreshMutex — it only takes it when it
// has decided to refresh — while a refresh running for any other object
// writes the same maps. Unsynchronized map read + map write is a Go
// runtime fatal error, not a panic: gin.Recovery() does not see it and the
// agent process exits, so one caller can take the agent down by racing
// another.
//
// The object below is pre-populated as already refreshed and already
// decrypted, which is exactly the state that makes the request path skip
// the refresh lock. The TA is deliberately unregistered: the call is
// expected to fail, the point is the cache access on the way there.
func TestStartConfidentialComputingConcurrentCacheAccess(t *testing.T) {
	a := &UdpAgent{}
	a.dataAccess.reset()

	const doId = "data-object-under-test"
	a.dataAccess.markRefreshed(doId)
	a.dataAccess.setDecrypted(doId, "/nonexistent/plaintext")
	a.dataAccess.setPolicy(doId, "policy-0")

	const readers, writers, iterations = 8, 4, 200

	var wg sync.WaitGroup
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < iterations; n++ {
				// Each goroutine owns its params map: the call injects
				// the plaintext path into it.
				_, err := a.StartConfidentialComputing(doId, "unregistered-ta", "fn", map[string]any{})
				if err == nil {
					t.Error("expected an error for an unregistered trusted application")
					return
				}
			}
		}()
	}
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for n := 0; n < iterations; n++ {
				// What a refresh writes: a freshly issued policy, the
				// engine its WASM loaded into, and the bookkeeping.
				policyId := fmt.Sprintf("policy-%d-%d", id, n)
				a.dataAccess.setPolicy(doId, policyId)
				a.dataAccess.putEngine(policyId, wasmEngine.NewEngine())
				a.dataAccess.markRefreshed(doId)
				a.dataAccess.setDecrypted(doId, fmt.Sprintf("/nonexistent/plaintext-%d-%d", id, n))
			}
		}(i)
	}
	wg.Wait()
}

// TestDataAccessCacheConcurrentAccessors covers the accessors the agent
// does not reach in the test above (preCheck's stale branch, policyFor on
// an unknown object, reset against live traffic) across many objects at
// once, which is the shape a multi-object console workload has.
func TestDataAccessCacheConcurrentAccessors(t *testing.T) {
	var c dataAccessCache
	c.reset()

	const objects, iterations = 16, 200

	var wg sync.WaitGroup
	for i := 0; i < objects; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			doId := fmt.Sprintf("do-%d", id)
			for n := 0; n < iterations; n++ {
				policyId := fmt.Sprintf("policy-%d-%d", id, n)
				c.setPolicy(doId, policyId)
				c.putEngine(policyId, wasmEngine.NewEngine())
				c.markRefreshed(doId)
				c.setDecrypted(doId, fmt.Sprintf("/nonexistent/%d-%d", id, n))

				c.preCheck(doId)
				c.preCheck(fmt.Sprintf("do-%d", (id+1)%objects))
				if got, exist := c.policyFor(doId); exist && got == "" {
					t.Errorf("policyFor(%q) returned an empty policy id", doId)
					return
				}
				if ce, exist := c.acquireEngine(policyId); exist {
					c.releaseEngine(ce)
				}
			}
		}(i)
	}

	// A console-triggered restart resets the cache while requests are in
	// flight; the agent documents those as best-effort, not as a crash.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for n := 0; n < iterations; n++ {
			c.reset()
		}
	}()

	wg.Wait()
}

// TestDataAccessCacheRetiredEngineOutlivesItsCall pins the engine
// lifetime rule: a policy refresh that replaces a data object's policy
// drops the old engine from the cache immediately, but must not close it
// while a trusted application call is still inside it. Calling into a
// closed wazero runtime panics in the engine's own Call wrappers, so
// closing eagerly would turn an unlucky overlap into a lost result.
func TestDataAccessCacheRetiredEngineOutlivesItsCall(t *testing.T) {
	var c dataAccessCache
	c.reset()

	c.setPolicy("do", "policy-old")
	c.putEngine("policy-old", wasmEngine.NewEngine())

	ce, exist := c.acquireEngine("policy-old")
	if !exist {
		t.Fatal("acquireEngine did not return the engine just stored")
	}

	// The refresh: binds a new policy, retires the old engine.
	c.setPolicy("do", "policy-new")

	c.mu.RLock()
	_, stillCached := c.engines["policy-old"]
	retired, inUse := ce.retired, ce.inUse
	c.mu.RUnlock()

	if stillCached {
		t.Error("the retired engine is still reachable from the cache")
	}
	if !retired {
		t.Error("the replaced engine was not marked retired")
	}
	if inUse != 1 {
		t.Errorf("in-use count = %d, want 1 while the call holds the engine", inUse)
	}

	if _, exist := c.acquireEngine("policy-old"); exist {
		t.Error("a retired engine must not be handed to a new call")
	}

	c.releaseEngine(ce)

	c.mu.RLock()
	inUse = ce.inUse
	c.mu.RUnlock()
	if inUse != 0 {
		t.Errorf("in-use count = %d after release, want 0", inUse)
	}
}
