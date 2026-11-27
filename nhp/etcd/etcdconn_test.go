package etcd

import (
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

// Run against a disposable etcd service; no production endpoint is assumed.
func TestEtcdLifecycle(t *testing.T) {
	endpoint := os.Getenv("OPENNHP_TEST_ETCD")
	if endpoint == "" {
		t.Skip("set OPENNHP_TEST_ETCD to a disposable etcd endpoint")
	}
	c := &EtcdConn{Endpoints: []string{endpoint}, Key: "opennhp-lifecycle-test"}
	if err := c.InitClient(); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.SetValue("initial"); err != nil {
		t.Fatal(err)
	}
	if value, err := c.GetValue(); err != nil || string(value) != "initial" {
		t.Fatalf("read after initialization: %q, %v", value, err)
	}
	// Client lifetime must outlast the startup status request's timeout.
	time.Sleep(requestTimeout + 100*time.Millisecond)
	if _, err := c.GetValue(); err != nil {
		t.Fatal(err)
	}
	received := make(chan string, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		c.WatchValue(func(value []byte) {
			select {
			case received <- string(value):
			default:
			}
		})
	}()
	deadline := time.After(5 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
waiting:
	for {
		select {
		case <-ticker.C:
			if err := c.SetValue("updated"); err != nil {
				t.Fatal(err)
			}
		case value := <-received:
			if value != "updated" {
				t.Fatalf("unexpected watch value %q", value)
			}
			break waiting
		case <-deadline:
			t.Fatal("watch did not deliver update")
		}
	}
	c.Close()
	c.Close() // repeated cleanup must not panic
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("watch did not stop")
	}
	if _, err := c.GetValue(); err == nil {
		t.Fatal("read after close succeeded")
	}
	// A watcher scheduled after Close must also exit.
	lateDone := make(chan struct{})
	go func() { c.WatchValue(func([]byte) { t.Error("callback after close") }); close(lateDone) }()
	select {
	case <-lateDone:
	case <-time.After(3 * time.Second):
		t.Fatal("watch after close did not exit")
	}
}

func TestEtcdEmptyEndpoints(t *testing.T) {
	c := &EtcdConn{}
	if err := c.InitClient(); err == nil {
		t.Fatal("empty endpoints accepted")
	}
	c.Close()
}

func TestEtcdUninitializedOperations(t *testing.T) {
	for _, setup := range []string{"uninitialized", "empty-endpoints"} {
		for _, method := range []string{"get", "set", "watch"} {
			t.Run(setup+"/"+method, func(t *testing.T) {
				c := &EtcdConn{}
				defer c.Close()
				if setup != "uninitialized" && c.InitClient() == nil {
					t.Fatal("invalid endpoints accepted")
				}
				defer func() {
					if p := recover(); p != nil {
						t.Errorf("operation after failed/missing initialization panicked: %v", p)
					}
				}()
				switch method {
				case "get":
					if _, err := c.GetValue(); err == nil {
						t.Error("uninitialized read succeeded")
					}
				case "set":
					if err := c.SetValue("value"); err == nil {
						t.Error("uninitialized write succeeded")
					}
				case "watch":
					c.WatchValue(func([]byte) { t.Error("uninitialized watch delivered a value") })
				}
			})
		}
	}
}

func TestEtcdOperationsAfterFailedStatus(t *testing.T) {
	// Keep the port reserved without serving etcd, so Status times out rather
	// than accidentally connecting to a service on a guessed unused port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	c := &EtcdConn{Endpoints: []string{"http://" + listener.Addr().String()}, Key: "failed-status"}
	defer c.Close()
	if err := c.InitClient(); err == nil {
		t.Fatal("Status unexpectedly succeeded")
	}
	if c.client == nil || c.ctx == nil || c.ctx.Err() == nil {
		t.Fatal("failed Status did not close initialized client")
	}
	select {
	case <-c.client.Ctx().Done():
	default:
		t.Error("failed Status left the underlying client open")
	}
	if _, err := c.GetValue(); err == nil {
		t.Error("read after failed Status succeeded")
	}
	if err := c.SetValue("value"); err == nil {
		t.Error("write after failed Status succeeded")
	}
	done := make(chan struct{})
	go func() { c.WatchValue(func([]byte) { t.Error("callback after failed Status") }); close(done) }()
	select {
	case <-done:
	case <-time.After(requestTimeout):
		t.Fatal("watch after failed Status did not return")
	}
	c.Close() // deferred Close repeats cleanup
}

// Shutdown must not return while a config callback can still mutate the daemon.
func TestEtcdCloseWaitsForCallback(t *testing.T) {
	endpoint := os.Getenv("OPENNHP_TEST_ETCD")
	if endpoint == "" {
		t.Skip("set OPENNHP_TEST_ETCD to a disposable etcd endpoint")
	}
	c := &EtcdConn{Endpoints: []string{endpoint}, Key: "opennhp-close-callback-test"}
	if err := c.InitClient(); err != nil {
		t.Fatal(err)
	}
	entered, release, watchDone := make(chan struct{}, 1), make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	go func() {
		defer close(watchDone)
		c.WatchValue(func([]byte) {
			calls.Add(1)
			select {
			case entered <- struct{}{}:
			default:
			}
			<-release
		})
	}()
	// Release the callback before cleanup, even when an assertion fails.
	defer c.Close()
	defer close(release)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(5 * time.Second)
waiting:
	for {
		select {
		case <-entered:
			break waiting
		case <-ticker.C:
			if err := c.SetValue("update"); err != nil {
				t.Fatal(err)
			}
		case <-deadline:
			t.Fatal("watch did not deliver update")
		}
	}
	// Queue another event while the first callback is held.
	if err := c.SetValue("queued"); err != nil {
		t.Fatal(err)
	}
	closed := make(chan struct{}, 2)
	for range 2 {
		go func() { c.Close(); closed <- struct{}{} }()
	}
	<-c.ctx.Done()
	select {
	case <-closed:
		t.Fatal("Close returned while a callback was still running")
	case <-time.After(100 * time.Millisecond):
	}
	// The deferred release must happen only once.
	release <- struct{}{}
	for range 2 {
		select {
		case <-closed:
		case <-time.After(3 * time.Second):
			t.Fatal("Close did not finish after callback returned")
		}
	}
	select {
	case <-watchDone:
	case <-time.After(3 * time.Second):
		t.Fatal("watch did not stop")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("queued callback ran during shutdown: %d calls", got)
	}
}
