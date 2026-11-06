package ac

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"

	"github.com/OpenNHP/opennhp/nhp/log"
)

func TestFailurePolicyConfig(t *testing.T) {
	for _, tc := range []struct {
		input, want string
		invalid     bool
	}{
		{"", FailurePolicyOpen, false},
		{`FailurePolicy = "fail-open"`, FailurePolicyOpen, false},
		{`FailurePolicy = "fail-closed"`, FailurePolicyClosed, false},
		{`FailurePolicy = ""`, FailurePolicyOpen, false},
		{`FailurePolicy = "closed"`, "", true},
		{`FailurePolicy = "FAIL-CLOSED"`, "", true},
		{`FailurePolicy = " fail-closed"`, "", true},
		{"FailurePolicy = \"fail-closed\"\nFilterMode = 1", "", true},
	} {
		t.Run(tc.input, func(t *testing.T) {
			var conf Config
			if err := toml.Unmarshal([]byte(tc.input), &conf); err != nil {
				t.Fatal(err)
			}
			a := &UdpAC{log: log.NewLogger("policy-test", 0, t.TempDir(), "test")}
			defer a.log.Close()
			err := a.updateBaseConfig(conf)
			if (err != nil) != tc.invalid {
				t.Fatalf("error=%v, invalid=%v", err, tc.invalid)
			}
			if tc.invalid {
				if a.config != nil {
					t.Fatal("invalid config installed")
				}
				return
			}
			if a.config.FailurePolicy != tc.want {
				t.Fatalf("got %q", a.config.FailurePolicy)
			}
			original := *a.config
			changed := original
			changed.FailurePolicy = FailurePolicyClosed
			if tc.want == FailurePolicyClosed {
				changed.FailurePolicy = FailurePolicyOpen
			}
			changed.DefaultIp = "192.0.2.9"
			if err := a.updateBaseConfig(changed); err == nil || !strings.Contains(err.Error(), "requires restart") {
				t.Fatalf("reload error: %v", err)
			}
			if a.config.FailurePolicy != original.FailurePolicy || a.config.DefaultIp != original.DefaultIp {
				t.Fatal("restart-only reload changed live config")
			}
			changed.FailurePolicy = "invalid"
			if err := a.updateBaseConfig(changed); err == nil {
				t.Fatal("invalid reload accepted")
			}
			if err := a.updateBaseConfig(original); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestFailurePolicyMalformedConfigRefusesStartup(t *testing.T) {
	old := ExeDirPath
	ExeDirPath = t.TempDir()
	defer func() { ExeDirPath = old }()
	if err := os.Mkdir(filepath.Join(ExeDirPath, "etc"), 0700); err != nil {
		t.Fatal(err)
	}
	for _, content := range []string{`FailurePolicy = 42`, `FailurePolicy = "fail-closed`} {
		if err := os.WriteFile(filepath.Join(ExeDirPath, "etc", "config.toml"), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		a := &UdpAC{}
		if err := a.loadBaseConfig(); err == nil {
			t.Fatal("malformed config accepted")
		}
		if a.config != nil {
			t.Fatal("partial config installed")
		}
	}
}

func TestFailurePolicyEtcdReloadRejectsChange(t *testing.T) {
	a := &UdpAC{log: log.NewLogger("policy-test", 0, t.TempDir(), "test")}
	defer a.log.Close()
	if err := a.updateBaseConfig(Config{FailurePolicy: FailurePolicyClosed}); err != nil {
		t.Fatal(err)
	}
	for _, policy := range []string{FailurePolicyOpen, "typo", ""} {
		content, err := toml.Marshal(ACEtcdConfig{BaseConfig: Config{FailurePolicy: policy, DefaultIp: "192.0.2.1"}})
		if err != nil {
			t.Fatal(err)
		}
		if err := a.updateEtcdConfig(content, true); err == nil {
			t.Fatalf("etcd change %q accepted", policy)
		}
		if a.config.FailurePolicy != FailurePolicyClosed || a.config.DefaultIp != "" {
			t.Fatal("rejected etcd update changed running config")
		}
	}
}
