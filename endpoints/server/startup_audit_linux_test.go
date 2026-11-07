package server

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/endpoints/internal/startuptest"
	"github.com/OpenNHP/opennhp/nhp/audit"
	"github.com/OpenNHP/opennhp/nhp/log"
)

// Exercise the real Start error path, including an async ledger's writer.
// The later failure occurs after HTTP, UDP, device and config acquisition.
func TestStartupAuditLedgerRollback(t *testing.T) {
	for _, remote := range []bool{false, true} {
		for _, async := range []bool{false, true} {
			for _, late := range []bool{false, true} {
				t.Run(fmt.Sprintf("remote=%v/async=%v/late=%v", remote, async, late), func(t *testing.T) {
					fixture := &startuptest.Remote{}
					remoteConfig := ""
					if remote {
						remoteConfig = startuptest.Config(fixture.Serve(t))
					}
					dir := startuptest.Directory(t, remoteConfig)
					if !remote {
						if err := os.Remove(filepath.Join(dir, "etc", "remote.toml")); err != nil {
							t.Fatal(err)
						}
					}
					ledgerPath := filepath.Join(dir, "audit.jsonl")
					ip, want := "invalid-ip", "udp listen ip address is incorrect"
					if late {
						ip, want = "127.0.0.1", "keystore: migrate:"
					}
					// Generate fixture key material at runtime; it is never a deployment key.
					key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))
					base := fmt.Sprintf("PrivateKeyBase64=%q\nListenIp=%q\nListenPort=0\nDatabasePath=%q\n", key, ip, dir)
					auditConfig := fmt.Sprintf("Enabled=true\nFilePath=%q\nAsync=%v\n", ledgerPath, async)
					httpConfig := "EnableHttp=true\nHttpListenIp='127.0.0.1'\nHttpListenPort=0\n"
					if remote {
						fixture.Value = "[BaseConfig]\n" + base + "[BaseConfig.Audit]\n" + auditConfig + "[HttpConfig]\n" + httpConfig
					} else {
						for name, content := range map[string]string{
							"config.toml": base + "[Audit]\n" + auditConfig,
							"http.toml":   httpConfig,
						} {
							if err := os.WriteFile(filepath.Join(dir, "etc", name), []byte(content), 0600); err != nil {
								t.Fatal(err)
							}
						}
					}
					before := startupAuditResources(t, dir)
					daemon := &UdpServer{}
					// Cleanup after assertions also contains the deliberate negative control.
					defer func() {
						daemon.closeAuditLedger()
						log.SetGlobalLogger(log.NewLogger("", 0, "", ""))
					}()
					startErr := daemon.Start(dir, 3)
					if startErr == nil || !strings.Contains(startErr.Error(), want) {
						t.Fatalf("Start error = %v, want %q", startErr, want)
					}
					if daemon.auditLedger == nil {
						t.Fatal("startup did not acquire the enabled audit ledger")
					}
					if _, err := os.Stat(ledgerPath); err != nil {
						t.Fatalf("ledger file was not created: %v", err)
					}
					if err := daemon.auditLedger.Log("after_failed_start", audit.SeverityInfo, nil); err == nil {
						t.Error("failed startup left audit ledger writable")
					}
					if daemon.IsRunning() {
						t.Error("failed startup marked daemon running")
					}
					if late {
						if daemon.listenConn == nil || daemon.httpServer == nil || daemon.device == nil {
							t.Fatal("late failure did not acquire expected UDP, HTTP and device resources")
						}
						if err := daemon.listenConn.SetReadBuffer(4096); err == nil {
							t.Error("failed startup left UDP open")
						}
						if daemon.httpServer.IsRunning() {
							t.Error("failed startup left HTTP running")
						}
						select {
						case _, open := <-daemon.device.DecryptedMsgQueue:
							if open {
								t.Error("failed startup left device queue open")
							}
						default:
							t.Error("failed startup did not stop device")
						}
					}
					if remote {
						if _, err := daemon.etcdConn.GetValue(); err == nil {
							t.Error("failed startup left etcd client usable")
						}
					}
					// Observe owned descriptors and goroutines before any test cleanup.
					deadline := time.Now().Add(3 * time.Second)
					for {
						after := startupAuditResources(t, dir)
						leaked := false
						for resource, count := range after {
							if count > before[resource] {
								leaked = true
							}
						}
						if !leaked {
							break
						}
						if time.Now().After(deadline) {
							t.Errorf("startup resources remain: before=%v after=%v", before, after)
							break
						}
						time.Sleep(time.Millisecond)
					}
					var stops sync.WaitGroup
					for range 2 {
						stops.Add(1)
						go func() { defer stops.Done(); daemon.Stop() }()
					}
					stops.Wait()
					// Repeated ledger cleanup must not close a newly reused descriptor.
					probe, err := os.Create(filepath.Join(dir, "cleanup-probe"))
					if err != nil {
						t.Fatal(err)
					}
					defer probe.Close()
					daemon.closeAuditLedger()
					daemon.closeAuditLedger()
					if _, err := probe.WriteString("still open"); err != nil {
						t.Fatalf("repeated cleanup closed unrelated descriptor: %v", err)
					}
				})
			}
		}
	}
}

// Linux resource observations supplement the public ledger and daemon assertions.
// Tests run serially in this package; baseline counts exclude fixture-owned work.
func startupAuditResources(t *testing.T, dir string) map[string]int {
	t.Helper()
	counts := map[string]int{}
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join("/proc/self/fd", entry.Name()))
		if err != nil { // ReadDir's own descriptor may already have closed.
			continue
		}
		switch {
		case strings.HasPrefix(target, dir+string(os.PathSeparator)):
			counts["owned files"]++
		case strings.Contains(target, "inotify"):
			counts["file watchers"]++
		case strings.HasPrefix(target, "socket:"):
			counts["sockets"]++
		}
	}
	buf := make([]byte, 2<<20)
	n := runtime.Stack(buf, true)
	if n == len(buf) {
		t.Fatal("goroutine snapshot truncated")
	}
	for _, role := range []string{
		"nhp/audit.(*Ledger).drain(", "nhp/log.(*AsyncLogWriter).",
		"nhp/utils.WatchFile.func", "nhp/core.(*Device).",
	} {
		counts[role] = strings.Count(string(buf[:n]), role)
	}
	return counts
}
