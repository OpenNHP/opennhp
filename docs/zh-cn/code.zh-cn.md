---
layout: page
title: 源代码解读
parent: 中文版
nav_order: 6
permalink: /zh-cn/code/
---

# OpeNHP代码解读
{: .fs-9 }

---

## 1. 层级架构

1. 上层逻辑组件层负责UDP的连接建立、维护与断开
2. Device层负责：1.将上层的消息明文转为NHP报文并发送到连接；2.将从连接收到的NHP报文转化为消息明文并提供上层处理
3. 上层逻辑组件提供
![avatar](/images/provide.png)

## 2. 连接管理

1. 上层逻辑组件可以建立并维护多个连接UdpConn，根据实际需求创建所需对象成员。每一个UdpConn起一个线程进行收发包操作。
2. 每一个UdpConn需要建立一个Device层的ConnData，并向Device ConnData传递实际连接中的远端地址，报文收发通道，cookie等。
3. 每一个UdpConn允许进行多次双向的transaction或单向发包。（agent除外，原则上agent每次请求都创建一个新的连接）
4. 每一个transaction都建立一个自身的线程和通道用于维持交互操作，超时后自行销毁。Local transaction（本地创建的交互）由device统一管理，Remote transaction（远端创建的交互）由远端连接管理，transaction的回应在收发包时需要找出相应的transaction线程进行后续操作。

## 3. 对象命名

1. 上层逻辑组件在收发方向上可能具有多重身份，Device层中使用initiator和responder表示发起方和接收方。

## 4. 报文缓冲区的创建与销毁（回收）

1. 为了提高吞吐率，报文缓冲区不采用自动垃圾回收机制而采用waitpool分配回收机制。
2. 接收：device创建报文缓冲区接收网络数据，根据NHP包头对报文进行解析与校验。解析结果存储在ResponderSessionParams结构中（名称不好理解，可能会改变）。明文消息仍然会使用报文缓冲区。缓冲区的销毁分两种情况，单向通信的结构体在上层应用获取明文消息后销毁。transaction接收缓冲区在transaction结束后销毁。
3. 发送：device创建报文缓冲区，填充包头并对消息进行加密后存储在InitiatorSessionParams结构中并发送。transaction发送在未收到对端回应时会重试发送。缓冲区的销毁分两种情况，单向通信的结构体在发送后销毁。transaction发送的缓冲区在transaction结束后销毁。

**消息的加密与解密：**
连接中接收到的UDP原始数据会被device解析并放入device的MsgToPacketQueue队列中，等待后端处理。
发送消息到连接时，需构建initiatorsessionstarter结构传入消息信息与连接信息，放入device的MsgToPacketQueue队列中，device会将消息进行加密发出。

## 5. NHP-Device 架构设计

1. Device负责NHP报文与消息的转换。Device初始化时需要指定类型和私钥。Device视自身类型只对相应的包进行处理。

2. 用于承载发送和接收报文的buffer比较大，所以由Device的内存Pool统一发放并回收（如果依赖于Go后台垃圾回收，高并发时会造成大量内存开销）。所以在开发时一定要注意buffer的分配**Device.AllocatePoolPacket\(\)** 和回收**Device.ReleasePoolPacket\(\)**。

   - 报文buffer回收点位于
     - 发送报文被发送后（本地transaction除外）
     - 接收报文解析完毕时（远程transaction除外）
     - 本地或远程transaction线程停止时

3. 上层逻辑调用接口**SendMsgToPacket**将消息转换成加密报文并发送到连接。

4. 上层逻辑调用接口**RecvPacketToMsg**将加密报文解析成消息后放入**DecryptedMsgQueue**队列并等待处理（通常情况）。

   - 特殊情况：如果请求发起方已指定接收通道，解析后的消息会被送到请求方指定的消息通道**ResponseMsgCh**，而不放进常规消息队列进行排队。

5. 交互（**transaction**）：一次请求需要等待一次回复的操作称为交互。一次由Device发起的交互请求为本地交互（**LocalTransaction**），一次由Device接收到的交互请求为远程交互（**RemoteTransaction**）。由于回应报文需要继承请求报文生成的**ChainKey**，所以所有的交互分发由Device进行管理。

6. 连接上下文（**ConnectionData**）：由上层逻辑传入的与连接相关的所有信息，Device在加密消息后将报文发送到连接。一个连接可以进行多个**transaction**。

7. 在建立发送请求时，需要创建**MsgAssembler**结构体。

   - Agent和AC必须填写消息类型**HeaderType**、对端**RemoteAddr**、对端公钥**PeerPk**和消息明文**Message**（如无特殊情况都采用消息压缩）。将填写好的**MsgAssembler**发给各自的**sendMessageRoutine\(\)** 即可进行新连接的建立或寻找已存在连接并进行转换后报文的发送。

   - Server必须填写消息类型**HeaderType**、连接上下文**ConnData**、对端公钥**PeerPk**和消息明文**Message**（如无特殊情况都采用消息压缩）。将填写好的**MsgAssembler**发给**Device.SendMsgToPacket\(\)** 即可进行转换后报文的发送。

   - 如果存在交互，可以直接使用上一条获得的 **\*PacketParserData**填入**MsgAssembler**结构体的**PrevParserData**字段，从而可以省略填写**RemoteAddr**、**ConnData**、**PeerPk**。

   - 如果请求期待回复数据，需要创建一个接收**PacketParserData**的通道，并对**MsgAssembler**结构体的**ResponseMsgCh**字段赋值。

## 6. NHP-Server

### NHP-Server 架构设计


1. Server启动时监听特定端口，等待Agent和AC进行连接。并由Agent或AC主动触发向Server的通信。不存在Server向Agent或AC主动建立连接的情况，通常情况下这种连接会跨防火墙或NAT导致不能建立。
   - 特殊情况：Server在收到Agent发起的敲门处理时，鉴权后需要主动向AC发起开门请求，并等待回应。

2. 发送消息时，向**sendMsgCh**发送创建好的**MsgAssembler**（必须从已有连接中指定**ConnData**）。**MsgAssembler**经过加密后会从此连接发出

3. 接收到报文时，会将报文进行解密获取明文消息。由**msghandler**分别进行处理。

### NHP-Server 配置文件

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

## 7. NHP-AC

### NHP-AC 架构设计

1. AC支持与多台Server互相进行通信。所有连接均为AC主动向Server发起。AC通过发送心跳包和NHP-AOL包维持与Server的连接。

2. AC与Server通信失效后，将尝试重新建立连接，如果一直无法与任何一台Server建立连接，则进入失效状态。

3. AC在启动后即开始与预设的服务器周期性建立连接并保持连接（AC很有可能在内网，所以不能由服务器先发连接）。连接时发送NHP_DOL消息，在收到服务器的回应后确认连接。连接期间视情况进行发送NHP_KPL消息保持连接。由**maintainServerConnectionRoutine**实现。

4. AC处理服务器发送过来的NHP_DOP消息，判断请求方的serviceId, appId是否匹配并进行IPSET操作，完成后返回NHP_DRT消息。

5. 发送消息时，向**sendMsgCh**发送创建好的**MsgAssembler**（必须指定**RemoteAddr**）。如果连接没有建立，AC会尝试建立并记录该连接。同时对此连接开启接收线程。**MsgAssembler**经过加密后会从此连接发出

6. 接收到报文时，会将报文进行解密获取明文消息。由**msghandler**分别进行处理。

### NHP-AC IP放行模式

IP放行模式分为两种：

1. ipPassMode为0（默认）时为立即放行模式，门禁开门时将以敲门来源IP地址为准。

2. ipPassMode为1时为预访问模式，门禁开门前将先开启对应协议的临时端口并返回server临时端口和临时访问token，在短时间内需由agent携带临时访问token进行临时连接，如果临时连接有效，则开门时放行将以此次临时连接的来源IP为准。

### NHP-AC 配置文件

`etc/config.toml`

```toml
[AC]
  # (optional) assign an unique id for this ac
  ACId = "abc_group_ac_001"
  # (mandatory) specify the private key in base64 format
  ACPrivateKey = "+B0RLGbe+nknJBZ0Fjt7kCBWfSTUttbUqkGteLfIp30="
  # 0: default, passing the knock source IP
  # 1: use pre-access procedure to determine the passing source IP
  IpPassMode = 0
  # (optional) ID of authorization service provider this ac belongs to
  AuthServiceId = "abc_group" 
  # (optional) ID of resources controlled by this ac
  ResourceIds = ["abc_group_web_server", "abc_group_api_server"]
  # (optional) ID of organization
  OrganizationId = "5f3e36149fa95c0414408ad4"

# server peers list
[[Servers]]
  # (optional) the server's hostname. Its resolved address overrides the "Ip" field
  Host = ""
  # IP address of the server peer
  Ip = "192.168.80.35"
  # listening port for the server peer
  Port = 62206
  # type: NHP-Server
  Type = 2
  # specify the server peer's public key in base64 format
  PublicKey = "WqJxe+Z4+wLen3VRgZx6YnbjvJFmptz99zkONCt/7gc="
  # expire timestamp of the public key (seconds from epoch)
  ExpireTime = 1716345064

# another server
#[[Servers]]
#  Ip = "192.168.135.1"
#  Port = 7776
#  Type = 2
#  PublicKey = "dstv1KlD2oVXiwgOxWtgZd+YmrOhU46W3emTGrHRADk="
#  ExpireTime = 1716345064

```

## NHP-Agent

### NHP-Agent 架构设计

1. Agent只与Server之间进行通信。Agent主动向Server发起短连接。不存在Agent在未建立连接时被动接收Server消息的情况。

2. 发送消息时，向**sendMsgCh**发送创建好的**MsgAssembler**（必须指定**RemoteAddr**）。如果连接没有建立，agent会尝试建立并记录该连接。同时对此连接开启接收线程。**MsgAssembler**经过加密后会从此连接发出

3. 接收到报文时，会将报文进行解密获取明文消息。由**msghandler**分别进行处理。

### NHP-Agent 配置文件

`etc/config.json`

```json
{
  // (mandatory) specify the private key in base64 format
  "privateKey": "+Jnee2lP6Kn47qzSaqwSmWxORsBkkCV6YHsRqXCegVo=",
  // (optional) ID of authorization service provider this agent belongs to
  "aspId": "abc_group",
  // (mandatory) an user object is necessary to carry out knock requests
  "user": {
    "userId": "zengl",
    "devId": "0123456789abcdef",
    "orgId": "abc.com.cn"
  },
  // preset resources to begin knock after start
  "knockRess": [
    {
      "aspId": "abc_group",
      "resId": "app_resource_group_001",
      "serverKey": "WqJxe+Z4+wLen3VRgZx6YnbjvJFmptz99zkONCt/7gc="
    }
  ],
  // list of preset allowed server peers to send knock request. It has the same effect as AddServer()
  "servers": [
    {
      // (optional) the server's hostname. Its resolved address overrides the "Ip" field
      "host": "",
      // IP address of the server peer
      "ip": "192.168.80.35",
      // listening port for the server peer
      "port": 62206,
      // type: NHP-Server
      "type": 2,
      // public key in base64 format
      "pubKeyBase64": "WqJxe+Z4+wLen3VRgZx6YnbjvJFmptz99zkONCt/7gc=",
      /// expire time for the public key (seconds from epoch)
      "expireTime": 1716345064
    }
  ]
}
```

## 9. Log 设计

日志log设计为异步写入，相比同步日志写入，在调用时不会立即进行写入日志文件的I/O操作而影响正常业务逻辑，在高并发时可以聚合多条日志并合并为一次文件写入，大幅减少文件I/O操作次数。

Logger对象可以单独创建使用（**NewLogger()**），也可以在应用程序启动时指定Package全局变量**glbLogger**，供整个工程使用。

注意：应用程序结束前，需调用Logger.Close()，确保最后缓存的日志能够写入文件。
