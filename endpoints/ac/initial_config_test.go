package ac

import (
	"errors"
	"strings"
	"testing"

	"github.com/OpenNHP/opennhp/nhp/etcd"
)

// An etcd value with a base config but no [[Servers]] must stop the initial
// load. Start used to throw this error away and boot without server peers.
func TestLoadInitialConfig_EtcdWithoutServersFails(t *testing.T) {
	a := &UdpAC{
		etcdConn: &etcd.EtcdConn{},
		getRemoteValue: func() ([]byte, error) {
			return []byte("[BaseConfig]\nACId = \"ac-1\"\n"), nil
		},
	}
	err := a.loadInitialConfig()
	if err == nil {
		t.Fatal("loadInitialConfig returned nil for an etcd config with no [[Servers]]")
	}
	if !strings.Contains(err.Error(), "initial load") {
		t.Errorf("error = %q, want it to mention the initial load", err)
	}
}

func TestLoadInitialConfig_EtcdReadErrorFails(t *testing.T) {
	boom := errors.New("etcd down")
	a := &UdpAC{
		etcdConn:       &etcd.EtcdConn{},
		getRemoteValue: func() ([]byte, error) { return nil, boom },
	}
	if err := a.loadInitialConfig(); !errors.Is(err, boom) {
		t.Fatalf("error = %v, want it to wrap %v", err, boom)
	}
}
