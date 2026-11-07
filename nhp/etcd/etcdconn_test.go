package etcd

import (
	"os"
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
	time.Sleep(3100 * time.Millisecond)
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
