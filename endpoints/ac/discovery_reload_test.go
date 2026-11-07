package ac

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/nhp/core"
	"github.com/OpenNHP/opennhp/nhp/utils"
)

func waitForDiscoveryTest(t *testing.T, fn func() bool) {
	t.Helper()
	end := time.Now().Add(8 * time.Second)
	for time.Now().Before(end) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timeout")
}
func startBlockedDiscovery(t *testing.T) (*UdpAC, string, <-chan struct{}, func()) {
	t.Helper()
	d := t.TempDir()
	binary := filepath.Join(d, "iptables")
	script := `#!/bin/sh
cd "$(dirname "$0")" || exit 2
[ "$1" != -w ] || shift
op=$1; chain=$2
[ -e "$chain" ] || : > "$chain"
case "$op" in
-S) cat "$chain";;
-I) if [ "$chain" = INPUT ]; then touch entered; while [ ! -f release ]; do sleep .02; done; fi
    echo "-A $chain -m comment --comment opennhp-ac-fail-open -j ACCEPT" >> "$chain";;
-D) : > "$chain";;
*) exit 2;;
esac
`
	if e := os.WriteFile(binary, []byte(script), 0700); e != nil {
		t.Fatal(e)
	}
	a := &UdpAC{config: &Config{FailurePolicy: FailurePolicyOpen}, iptables: &utils.IPTables{Binary: binary}, serverPeerMap: map[string]*core.UdpPeer{}}
	a.signals.stop = make(chan struct{})
	a.signals.serverMapUpdated = make(chan struct{}, 1)
	a.wg.Add(1)
	go a.maintainServerConnectionRoutine()
	done := make(chan struct{})
	go func() { a.wg.Wait(); close(done) }()
	var once sync.Once
	stop := func() { once.Do(func() { close(a.signals.stop) }) }
	t.Cleanup(func() {
		os.WriteFile(filepath.Join(d, "release"), nil, 0600)
		stop()
		select {
		case <-done:
		case <-time.After(8 * time.Second):
			t.Error("maintenance did not stop")
		}
	})
	waitForDiscoveryTest(t, func() bool { _, e := os.Stat(filepath.Join(d, "entered")); return e == nil })
	return a, d, done, stop
}
func TestDiscoveryReloadWaitsForFirewallUpdate(t *testing.T) {
	a, d, _, _ := startBlockedDiscovery(t)
	a.serverPeerMutex.Lock()
	a.serverPeerMap = map[string]*core.UdpPeer{"new": {Ip: "invalid-ip"}}
	a.serverPeerMutex.Unlock()
	a.notifyServerMapUpdated()
	// A new cycle must not start while an old cycle is changing firewall rules.
	time.Sleep(200 * time.Millisecond)
	if len(a.signals.serverMapUpdated) != 1 {
		t.Fatal("reload consumed while old firewall update is in flight")
	}
	os.WriteFile(filepath.Join(d, "release"), nil, 0600)
	waitForDiscoveryTest(t, func() bool { return len(a.signals.serverMapUpdated) == 0 })
	// The new peer has not failed discovery. The next check must close fallback.
	waitForDiscoveryTest(t, func() bool { b, err := os.ReadFile(filepath.Join(d, "FORWARD")); return err == nil && len(b) == 0 })
}
func TestDiscoveryShutdownDuringFirewallChange(t *testing.T) {
	_, d, done, stop := startBlockedDiscovery(t)
	stop()
	if err := os.WriteFile(filepath.Join(d, "release"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("shutdown hung")
	}
	for _, chain := range []string{"INPUT", "FORWARD"} {
		b, err := os.ReadFile(filepath.Join(d, chain))
		if err != nil {
			t.Fatal(err)
		}
		if len(b) > 0 {
			t.Errorf("shutdown left fallback: %s", b)
		}
	}
}
func TestPeerReloadNotifiesDiscovery(t *testing.T) {
	key := make([]byte, core.PrivateKeySize)
	key[0] = 1
	device := core.NewDevice(core.NHP_AC, key, nil)
	if device == nil {
		t.Fatal("device")
	}
	defer device.Stop()
	a := &UdpAC{device: device, config: &Config{}, serverPeerMap: map[string]*core.UdpPeer{}}
	a.running.Store(true)
	a.signals.serverMapUpdated = make(chan struct{}, 1)
	if e := a.updateServerPeers([]*core.UdpPeer{{PubKeyBase64: device.PublicKeyBase64(), Ip: "192.0.2.2", Port: 62206}}); e != nil {
		t.Fatal(e)
	}
	if len(a.serverPeerMap) != 1 {
		t.Fatal("map not installed")
	}
	select {
	case <-a.signals.serverMapUpdated:
	default:
		t.Error("new peer map installed without notifying active discovery")
	}
}
