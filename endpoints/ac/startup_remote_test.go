package ac

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/OpenNHP/opennhp/endpoints/internal/startuptest"
	"github.com/OpenNHP/opennhp/nhp/log"
)

func TestStartupPreservesRemoteInitializationError(t *testing.T) {
	for _, mode := range []string{"provider", "status"} {
		t.Run(mode, func(t *testing.T) {
			remote := "Provider = 'unsupported'\n"
			want := "unknown remote provider"
			if mode == "status" {
				fixture := &startuptest.Remote{FailStatus: true}
				remote = startuptest.Config(fixture.Serve(t))
				want = "startup status denied"
			}
			dir := startuptest.Directory(t, remote)
			daemon := &UdpAC{}
			defer log.SetGlobalLogger(log.NewLogger("", 3, "", ""))
			defer func() {
				if daemon.etcdConn != nil {
					daemon.etcdConn.Close()
				}
				if daemon.log != nil {
					daemon.log.Close()
				}
				if p := recover(); p != nil {
					t.Errorf("Start panicked instead of returning initialization error: %v", p)
				}
			}()
			err := daemon.Start(dir, 0)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("Start error = %v, want original %q", err, want)
			}
		})
	}
}

func TestStartupRejectsRemoteLoadAndRollsBack(t *testing.T) {
	startuptest.FakeFirewall(t)
	for _, mode := range []string{"read", "payload", "partial", "tls-partial"} {
		t.Run(mode, func(t *testing.T) {
			fixture := &startuptest.Remote{}
			dir := startuptest.Directory(t, startuptest.Config(fixture.Serve(t)))
			fixture.Value = fmt.Sprintf("[BaseConfig]\nPrivateKeyBase64='AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE='\nListenIp='127.0.0.1'\nListenPort=0\nDatabasePath=%q\n", dir+"/keys.db")
			if mode == "read" {
				fixture.FailSecond = true
			}
			if mode == "payload" {
				fixture.SecondValue = "invalid = ["
			}
			if mode == "partial" || mode == "tls-partial" {
				fixture.SecondValue = fixture.Value + "[HttpConfig]\nEnableHttp=true\nHttpListenIp='127.0.0.1'\nHttpListenPort=0\n[[Servers]]\nPubKeyBase64='invalid'\nPort=-1\n"
			}
			if mode == "tls-partial" {
				ts := httptest.NewTLSServer(http.NotFoundHandler())
				cert := ts.TLS.Certificates[0]
				ts.Close()
				key, e := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(dir, "cert.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}), 0600); e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(dir, "key.pem"), pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}), 0600); e != nil {
					t.Fatal(e)
				}
				fixture.SecondValue = strings.Replace(fixture.SecondValue, "EnableHttp=true", "EnableHttp=true\nEnableTLS=true\nTLSCertFile='cert.pem'\nTLSKeyFile='key.pem'", 1)
			}

			daemon := &UdpAC{}
			defer log.SetGlobalLogger(log.NewLogger("", 3, "", ""))
			defer func() {
				if daemon.IsRunning() {
					daemon.Stop()
					return
				}
				if daemon.etcdConn != nil {
					daemon.etcdConn.Close()
				}
				if daemon.httpServer != nil {
					_ = daemon.httpServer.httpServer.Close()
					daemon.httpServer.wg.Wait()
				}

				if daemon.log != nil {
					daemon.log.Close()
				}
			}()
			err := daemon.Start(dir, 0)
			if err == nil {
				t.Error("Start accepted failed initial remote configuration")
			}
			if daemon.IsRunning() {
				t.Error("failed startup marked running")
			}
			if daemon.etcdConn != nil {
				if _, e := daemon.etcdConn.GetValue(); e == nil {
					t.Error("failed startup left etcd client usable")
				}
			}
			if daemon.httpServer != nil && daemon.httpServer.IsRunning() {
				t.Error("failed startup left HTTP running")
			}
			if daemon.httpServer != nil {
				select {
				case <-daemon.httpServer.signals.stop:
				default:
					t.Error("HTTP shutdown was not performed during rollback")
				}
			}

		})
	}
}

func TestDaemonStopDuringRemoteUpdates(t *testing.T) {
	startuptest.FakeFirewall(t)
	fixture := &startuptest.Remote{Updates: make(chan []byte, 16), WatchStarted: make(chan struct{})}
	dir := startuptest.Directory(t, startuptest.Config(fixture.Serve(t)))
	fixture.Value = fmt.Sprintf("[BaseConfig]\nPrivateKeyBase64='AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE='\nListenIp='127.0.0.1'\nListenPort=0\nDatabasePath=%q\n", dir+"/keys.db")
	fixture.Value += "\n[[Servers]]\nPubKeyBase64='AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE='\nIp='127.0.0.1'\nPort=62208\n"
	daemon := &UdpAC{}
	defer log.SetGlobalLogger(log.NewLogger("", 3, "", ""))
	daemon.Stop() // An early no-op must not consume the later shutdown.
	if err := daemon.Start(dir, 0); err != nil {
		t.Fatal(err)
	}
	defer daemon.Stop()
	select {
	case <-fixture.WatchStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("watch did not start")
	}
	value := []byte(fixture.Value + "[HttpConfig]\nEnableHttp=true\nHttpListenIp='127.0.0.1'\nHttpListenPort=0\n")
	fixture.Updates <- value
	// Prove the callback actually created a listener before racing shutdown.
	deadline := time.Now().Add(3 * time.Second)
	for {
		daemon.remoteConfigUpdateMutex.Lock()
		applied := daemon.httpServer != nil && daemon.httpServer.IsRunning()
		daemon.remoteConfigUpdateMutex.Unlock()
		if applied {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("remote HTTP update was not applied")
		}
		time.Sleep(time.Millisecond)
	}
	done := make(chan struct{})
	var sender sync.WaitGroup
	sender.Add(1)
	go func() {
		defer sender.Done()
		for i := 0; ; i++ {
			update := []byte(strings.Replace(string(value), "Port=62208", fmt.Sprintf("Port=%d", 62209+i%2), 1))
			select {
			case <-done:
				return
			case fixture.Updates <- update:
			}
		}
	}()
	// Exercise repeated/concurrent Stop with an actual registered watch and
	// valid HTTP updates in flight, not just an EtcdConn callback in isolation.
	stopped := make(chan any, 2)
	for range 2 {
		go func() { defer func() { stopped <- recover() }(); daemon.Stop() }()
	}
	for range 2 {
		select {
		case p := <-stopped:
			if p != nil {
				t.Errorf("concurrent Stop panicked: %v", p)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Stop did not finish")
		}
	}
	close(done)
	sender.Wait()
	if daemon.IsRunning() {
		t.Error("daemon remains running")
	}
	if daemon.httpServer != nil && daemon.httpServer.IsRunning() {
		t.Error("remote update left HTTP running after Stop")
	}
	if _, err := daemon.etcdConn.GetValue(); err == nil {
		t.Error("etcd client usable after Stop")
	}
}
