package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	toml "github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"

	"github.com/OpenNHP/opennhp/endpoints/ac"
	common "github.com/OpenNHP/opennhp/nhp/common"
	core "github.com/OpenNHP/opennhp/nhp/core"
)

func TestTomlUnmarshal(t *testing.T) {
	config := `
  a = "abc"
  B = "aBc"
  C = "abC"
  ACId = "test.clouddeep.cn-ac"
  PrivateKey = "+B0RLGbe+nknJBZ0Fjt7kCBWfSTUttbUqkGteLfIp30="
  IpPassMode = 0
  AuthServiceId = "core.clouddeep.cn"
  ResourceIds = ["test.clouddeep.cn"]
  OrganizationId = "5f3e36149fa95c0414408ad4"

  [[Servers]]
  Hostname = ""
  Ip = "192.168.80.35"
  Port = 62206
  Type = 2
  PubKeyBase64 = "WqJxe+Z4+wLen3VRgZx6YnbjvJFmptz99zkONCt/7gc="
  ExpireTime = 1716345064

  [Data]
  "Abc" = "TEST"
  "aBc" = true
  "abC" = 1234
  `

	type embed struct {
		a string
		B string
		C string
	}

	type obj struct {
		embed          `mapstructure:",squash"`
		ACId           string
		PrivateKey     string
		IpPassMode     int
		AuthServiceId  string
		ResourceIds    []string
		OrganizationId string
		Servers        []*core.UdpPeer
		Data           map[string]any
	}
	var Obj obj
	err := toml.Unmarshal([]byte(config), &Obj)
	if err != nil {
		fmt.Printf("toml unmarshal error: %v\n", err)
	}
	fmt.Printf("peer0: %+v\n", Obj.Servers[0])
	fmt.Printf("config: %+v, ACId: %s, B: %s\n", Obj, Obj.ACId, Obj.B)
	fmt.Printf("Data: %+v, %s, %t, %d,\n", Obj.Data, Obj.Data["Abc"].(string), Obj.Data["aBc"].(bool), Obj.Data["abC"].(int64))
	// below will trigger panic
	//fmt.Printf("non-exist data %s, %t, %d\n", Obj.Data["Abc1"].(string), Obj.Data["aBc2"].(bool), Obj.Data["abC3"].(int64))
}

func TestTomlViperHandling(t *testing.T) {
	t.Skip("Skipping interactive test with infinite loop - for manual testing only")
	var config ac.Config
	configDir := "../ac/main/etc"
	viper.SupportedExts = []string{"toml"}
	viper.AddConfigPath(configDir)
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("load toml file failed: %v\n", err)
		return
	}
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Printf("file changed: %s\n", in.Name)
		// content, err := os.ReadFile(in.Name)
		// if err != nil {
		// 	fmt.Printf("Failed to read config file: %v\n", err)
		// 	return
		// }
		// err = toml.Unmarshal(content, &config)
		err = viper.Unmarshal(&config)
		if err != nil {
			fmt.Printf("Failed to unmarshal config file: %v\n", err)
			return
		}
		fmt.Printf("config: %+v\n", config)
	})
	viper.WatchConfig()

	//content, err := os.ReadFile(filepath.Join(configDir, "config.toml"))
	//err = toml.Unmarshal(content, &config)
	err = viper.Unmarshal(&config)
	if err != nil {
		fmt.Printf("Failed to unmarshal config file: %v\n", err)
		return
	}
	fmt.Printf("config: %+v\n", config)

	for {
		time.Sleep(5 * time.Second)
	}
}

func TestWxwebTomlViperHandling(t *testing.T) {
	t.Skip("Skipping interactive test with infinite loop - for manual testing only")
	configFile := "../server/plugins/wxweb/etc/resource.toml"
	viper.SetConfigFile(configFile)
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("load toml file failed: %v\n", err)
		return
	}

	var resources common.ResourceGroupMap
	err = viper.Unmarshal(&resources)
	if err != nil {
		fmt.Printf("failed to unmarshal config file: %v\n", err)
	}

	fmt.Printf("config: %+v\n", resources)
	for k, res := range resources {
		fmt.Printf("resid %s, config: %+v\nlogin info: %+v\n", k, res, res.ExInfo)

		for key, info := range res.Resources {
			fmt.Printf("sub res name %s, info: %+v\n", key, info)
		}
	}

	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Printf("file changed: %s\n", in.Name)
		// content, err := os.ReadFile(in.Name)
		// if err != nil {
		// 	fmt.Printf("Failed to read config file: %v\n", err)
		// 	return
		// }
		// err = toml.Unmarshal(content, &config)
		err = viper.Unmarshal(&resources)
		if err != nil {
			fmt.Printf("Failed to unmarshal config file: %v\n", err)
			return
		}
		for _, res := range resources {
			fmt.Printf("config: %+v\n", res)
		}
	})
	viper.WatchConfig()

	for {
		time.Sleep(5 * time.Second)
	}
}

func TestUdpServerTomlViperHandling(t *testing.T) {
	t.Skip("Skipping interactive test with infinite loop - for manual testing only")
	configFile := "../etc/ac.toml"
	viper.SetConfigFile(configFile)
	viper.SetConfigType("toml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("load toml file failed: %v\n", err)
		return
	}

	type acPeers struct {
		ACs []*core.UdpPeer
	}

	var peers acPeers
	err = viper.Unmarshal(&peers)
	if err != nil {
		fmt.Printf("failed to unmarshal config file: %v\n", err)
	}

	fmt.Printf("config: %+v\n", peers)
	for _, peer := range peers.ACs {
		fmt.Printf("peer: %+v\n", peer)
	}

	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Printf("file changed: %s\n", in.Name)
		// content, err := os.ReadFile(in.Name)
		// if err != nil {
		// 	fmt.Printf("Failed to read config file: %v\n", err)
		// 	return
		// }
		// err = toml.Unmarshal(content, &config)
		err = viper.Unmarshal(&peers)
		if err != nil {
			fmt.Printf("Failed to unmarshal config file: %v\n", err)
			return
		}
		for _, peer := range peers.ACs {
			fmt.Printf("config: %+v\n", peer)
		}
	})
	viper.WatchConfig()

	for {
		time.Sleep(5 * time.Second)
	}
}
