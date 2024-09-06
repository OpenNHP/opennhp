# NHP-Server

## 架构设计


1. Server启动时监听特定端口，等待Agent和AC进行连接。并由Agent或AC主动触发向Server的通信。不存在Server向Agent或AC主动建立连接的情况，通常情况下这种连接会跨防火墙或NAT导致不能建立。
   - 特殊情况：Server在收到Agent发起的敲门处理时，鉴权后需要主动向AC发起开门请求，并等待回应。

2. 发送消息时，向**sendMsgCh**发送创建好的**MsgAssembler**（必须从已有连接中指定**ConnData**）。**MsgAssembler**经过加密后会从此连接发出

3. 接收到报文时，会将报文进行解密获取明文消息。由**msghandler**分别进行处理。

## NHP-Server配置文件

`etc/config.json`

```json
{
  // (mandatory) private key in base64 format
  "privateKey": "eHdyRHKJy/YZJsResCt5XTAZgtcwvLpSXAiZ8DBc0V4=",
  // (mandatory) specify the udp listening port
  "listenPort": 62206,
  // whether to validate peer's public key when receiving NHP packet from agent. If true, server must have a pre-recorded public key pool (in "agents" field) of all allowed agents. If false, server skip public key validation, so it reduces secure level.
  "disableAgentValidation": false,
  // list of preset allowed AC peers. only public key and expire time are needed. It has the same effect as AddACPeer()
  "acs": [
    {
      // type: NHP-AC
      "type": 3,
      // public key in base64 format
      "pubKeyBase64": "Fr5jzZDVpNh5m9AcBDMtHGmbCAczHyPegT8IxQ3XAzE=",
      // expire time for the public key (seconds from epoch)
      "expireTime": 1716345064
    }
  ],
  // list of preset allowed agent peers. only public key and expire time are needed. It has the same effect as AddAgentPeer()
  "agents": [
    {
      // type: NHP-Agent
      "type": 1,
      // public key in base64 format
      "pubKeyBase64": "WnJAolo88/q0x2VdLQYdmZNtKjwG2ocBd1Ozj41AKlo=",
      // expire time for the public key (seconds from epoch)
      "expireTime": 1716345064
    }
  ],
  // (optional) placeholder of preset url for possible authorization service provider
  "asps": {
    "abc.com": {
      "aspId": "abc.com",
      "urlAddr": "http://120.92.16.228:30088",
      "urlOTP": "/nhp/api/v1/preAuth",
      "urlReg": "/nhp/api/v1/registerAgent",
      "urlAuth": "/nhp/api/v1/verifyAuth",
      "urlList": "/nhp/api/v1/resourceList"
    }
  },
  // (optional) specify other source IP addresses to be opened by the ac that may come along with certain agent IP address 
  "srcAsscAddrs": {
    "192.168.2.27": [
      {
        "ip": "192.168.2.26",
        "port": 54222
      },
      {
        "ip": "192.168.2.28",
        "port": 54223
      }
    ]
  },
  // preset resources for udp knocking
  "udpRess": {
    // ID of authorization service provider
    "abc_group": {
      // ID of resource group
      "app_resource_group_000": {
        // skip service provider authorization and use this preset resource group
        "skipAuth": true,
        // set the desired open time for this resource group (in second)
        "opnTime": 120,
         "resInfo": {
          // name of resource
          "apiServer": {
            // (optional) hostname overrides addr.ip at knock feedback
            "host": "api.abc.com",
            // (mandatory) request ac to open which layer 4 address and protocol of this resource
            "addr": {
              // (mandatory) request ac to open traffic destinated to the public IP address of this resource
              "ip": "12.34.56.78",
              // (optional) request ac to open traffic destinated to the port number where this resource hosts on. empty or 0 means open all port numbers.
              "port": 443,
              // (optional) protocol, "tcp": request ac to open only tcp traffic, "udp": request ac to open only udp traffic, empty: request ac to open tcp + udp + icmp echo traffic
              "proto": "tcp"
            },
          }
         }
      }
    }
  },
  // preset resources for http knocking
  "httpRess": {
    // ID of authorization service provider
    "abc_group": {
      // ID of the resource group, usually it means AppId
      "app_resource_group_001": {
        // set the desired open time for this resource group (in second)
        "opnTime": 120,
        // contains multiple resources
        "resInfo": {
          // name of resource
          "apiServer": {
            // (optional) hostname overrides addr.ip at knock feedback
            "host": "api.abc.com",
            // (mandatory) request ac to open which layer 4 address and protocol of this resource
            "addr": {
              // (mandatory) request ac to open traffic destinated to the public IP address of this resource
              "ip": "12.34.56.78",
              // (optional) request ac to open traffic destinated to the port number where this resource hosts on. empty or 0 means open all port numbers.
              "port": 443,
              // (optional) protocol, "tcp": request ac to open only tcp traffic, "udp": request ac to open only udp traffic, empty: request ac to open tcp + udp + icmp echo traffic
              "proto": "tcp"
            },
            // (optional) the private layer 4 address of the ac. In some network, server may communicate with ac using private addresses. 
            "acAddr": {
              "ip": "172.16.1.2",
              "port": 443
            },
            // whether to append ":port" at the end of hostname/ip at knock feedback. For example, set this field to false if this resource use https and requesting ac to open port 443.
            "portSuffix": false
          },
          // another resource
           "webServer": {
            "host": "www.abc.com",
            "addr": {
              "ip": "23.45.67.89",
              "port": 8080,
              "proto": ""
            },
            "portSuffix": true
          }
        },
        // (optional) additional key info for server calling further authroization APIs
        "accessKey": "b3458c581ef0efb7b669",
        "secretKey": "f21c2a02c09a641a11cf"
      }
    },
    // another authorization service provider
    "xyz_org": {
      "abcd1234": {
        "opnTime": 120,
        "resInfo": {
          "udpServer": {
            "host": "server.xyz.net",
            "addr": {
              "ip": "1.2.3.4",
              "port": 443,
              "proto": "udp"
            },
            "portSuffix": false
          }
        },
        // (optional) additional key info for server calling further authroization APIs
        "appKey": "demo-l2T0J3U3mQZ3",
        "appSecret": "hVqd8eOqCFg5cc1D2ouACs3q"
      }
    }
  }
}
```
