# OpenNHP远端配置

## 1 OpenNHP远端配置说明

为方便统一管理OpenNHP配置信息，OpenNHP使用etcd作为统一远端配置中心。

OpenNHP三大核心组件对配置文件的支持方式：

| 组件       | 远端配置                        | 本地配置                        |
| ---------- | ------------------------------- | ------------------------------- |
| nhp server | <font color="green">支持</font> | <font color="green">支持</font> |
| nhp ac     | <font color="green">支持</font> | <font color="green">支持</font> |
| nhp agent  | <font color="red">不支持</font> | <font color="green">支持</font> |

## 2 etcd环境部署

### 2.1 etcd部署

- etcd下载地址：https://github.com/etcd-io/etcd/releases/
- 下载对应服务器环境的etcd版本
- etcd服务部署与启动参数安装包中的README.md文件

### 2.2 etcd可视化配置工具

- 启动etcd服务后，可以通过etcd的可视化工具来进行OpenNHP的配置信息的编辑，本文以工具etcdkeeper为例来配置etcd客户化配置环境

- 将etcdkeeper下载到etcd部署服务器上，下载方法：

  ```sh
  wget https://github.com/evildecay/etcdkeeper/releases/download/v0.7.6/etcdkeeper-v0.7.6-linux_x86_64.zip
  unzip etcdkeeper-v0.7.6-linux_x86_64.zip
  ```

- etcdkeeper启动方法：

  - IP设置为服务器实际IP
  - 端口设置为实际方法端口
  - 启动成功后可通过浏览器访问，访问地址如：http://192.168.32.30:8800

  ```sh
  cd etcdkeeper
  chmod +x etcdkeeper
  ./etcdkeeper -h 192.168.32.30 -p 8800 
  ```

## 3 nhp server远端配置

### 3.1 远端配置访问配置

etc目录下的remote.toml为nhp server服务访问远端配置中心ETCD的配置信息

- Endpoints：etcd访问地址
- Key：nhp server获取本服务器的key

```toml
# NHP-Server remote config
# field with (-) does not support dynamic update
# If the file remote.toml exists, NHP-Server will obtain remote configuration information through the etcd client

# Endpoints: ETCD service access address.
# Key: NHP-Server obtain the configuration information through this key.
# Username: The account of the NHP-Server accessing ETCD.
# Password: The password for NHP-Server to access ETCD.

Endpoints = ["172.16.3.53:2379"]
Key = "openserver-1"
```


### 3.2 启用远端配置

确保配置文件./nhp-server/etc/remote.toml文件存在，并确保Endpoints和Key正确配置，在nhp server服务启动时会加载remote.toml文件并去获取文件中Key对应的配置内容。



如想继续使用本地文件的配置方式，移除remote.toml文件或将remote.toml中的Endpoints或Key的内容置为空，服务将会使用本地配置文件来启动。



### 3.3 远端配置内容

nhp server远端配置内容分为如下七部分：

- [BaseConfig]部分内容对应本地配置文件config.toml
- [HttpConfig]部分内容对应本地配置文件http.toml
- [[ACs]]部分内容对应本地配置文件ac.toml
- [[Agents]]部分内容对应本地配置文件agent.toml
- [[DBs]]部分内容对应本地配置文件db.toml
- [[AuthServiceId]]部分内容对应本地配置文件resource.toml
- [[SrcIps]]部分内容对应本地配置文件srcip.toml

远端配置信息如下：

```toml
# NHP-Server base config
# field with (-) does not support dynamic update

# PrivateKeyBase64 (-): server private key in base64 format.
# DefaultCipherScheme: 0: gmsm, 1: curve25519.
# ListenIp (-): udp listening address.
# ListenPort (-): udp listening port.
# Hostname (-): server domain name.
# LogLevel: 0: silent, 1: error, 2: info, 3: audit, 4: debug, 5: trace.
# DisableAgentValidation: whether for the server to skip the agent's public key validation.
[BaseConfig]
PrivateKeyBase64 = "SFhGcTlhYlU4dTJMemNsaWM5TFZ2NzZDQjJNd2VGZ2Q="
DefaultCipherScheme = 0
ListenIp = ""    # empty for ipv4 + ipv6, "0.0.0.0" for ipv4 only
ListenPort = 62206
Hostname = "localhost" # the hostname of NHP-Server
LogLevel = 4
DisableAgentValidation = false


# http server config

# EnableHttp: true: turn on http server, false: shutdown http server.
# EnableTLS: whether to use TLS certificates for hosting https server.
# TLSCertFile: certificate file path.
# TLSKeyFile: key file path.
# to update http changes, you need to restart the http server by changing "EnableHttp" to "false" and then switch it back to "true".
[HttpConfig]
EnableHttp = true
EnableTLS = true
HttpListenIp = "0.0.0.0"    # empty for ipv4 + ipv6, "0.0.0.0" for ipv4 only, "127.0.0.1" for local ipv4 access only
TLSCertFile = "cert/cert.pem"
TLSKeyFile = "cert/cert.key"

# list the AC peers for the server under [[ACs]] table

# PubKeyBase64: public key for the AC in base64 format.
# ExpireTime (epoch timestamp in seconds): peer key validation will fail when it expires.
[[ACs]]
PubKeyBase64 = "3wDnLkZ3ccK3Ezi3pdG003rFbX4riMIOKfvFlu4t5yKhijSdIkAx8C6mVMFxygfZ0ijt8IDAS2RdTnfZpUCbZA=="
ExpireTime = 1924991999

# list the agent peers for the server under [[Agents]] table

# PubKeyBase64: public key for the agent in base64 format.
# ExpireTime (epoch timestamp in seconds): peer key validation will fail when it expires.
[[Agents]]
PubKeyBase64 = "WnJAolo88/q0x2VdLQYdmZNtKjwG2ocBd1Ozj41AKlo="
ExpireTime = 1924991999

# list the device peers for the server under [[Devices]] table

# PubKeyBase64: public key for the device in base64 format.
# ExpireTime (epoch timestamp in seconds): peer key validation will fail when it expires.
[[DBs]]
PubKeyBase64 = "CtxNuy7lJ1mJgjqWplcwN8dZhXhSNPhECja1A0OWKa+2wtI7xuB3jPcamogGZGBBfQ4SqnoPGLA7zRQaAotoxg=="
ExpireTime = 1924991999



# List resources and their sub-fields here

# syntax ["{AuthServiceId}"]
# AuthSvcId: id of the authentication and authorization service provider.
# PluginPath: path of plugin to implement auth logic.
[[AuthServiceId]]
AuthSvcId="default"
PluginPath = "default"

[[AuthServiceId]]
AuthSvcId="product-sdp"
PluginPath = "product-sdp1"


# list additional source addresses to be passed along with the agent address

# syntax [["{SrcIps}"]]
# SrcIp: specify the agent source ip. Each source ip can have multiple side source ips.
# Ip: specify a side source ip address to be also passed after successful knock.
[[SrcIps]]
SrcIp = "192.168.2.27"
Ip = ["192.168.2.26","192.168.2.28"]

[[SrcIps]]
SrcIp = "192.168.3.27"
Ip = ["192.168.3.28"]
```



本地文件支持动态变更的内容在远端配置环境下同样支持动态变更。

## 4 nhp ac远端配置

### 4.1 远端配置访问配置

参照章节3.1

### 4.2 启用远端配置

参照章节3.2



### 4.3 远端配置内容

nhp ac远端配置内容分为如下三部分：

- [BaseConfig]部分内容对应本地配置文件config.toml
- [HttpConfig]部分内容对应本地配置文件http.toml
- [[Servers]]部分内容对应本地配置文件server.toml

远端配置信息如下：

```toml
# NHP-AC base config
# field with (-) does not support dynamic update

# ACId (-): specify the id of this AC.
# PrivateKeyBase64 (-): AC private key in base64 format.
# DefaultCipherScheme: 0: gmsm, 1: curve25519.
# IpPassMode:
#  0: (default) immediately pass traffic with the agent source ip,
#  2: process pre-access to determine actual agent source ip then pass.
# FilterMode: 
#  0: iptables (default)
#  1: ebpf xdp (requires Linux kernel >= 5.6 and XDP-capable network interface)
# LogLevel: 0: silent, 1: error, 2: info, 3: audit, 4: debug, 5: trace.
# AuthServiceId (-): id for authentication and authorization service provider this AC belongs to.
# ResourceIds (-): resource group ids that this AC protects.
[BaseConfig]
ACId = "testAC-346"
DefaultIp = "172.16.3.52"
PrivateKeyBase64 = "N1o4c1BsSHZXQ1hsUFQyUzQ2QkJ2YlhQSGxYbDVmcU0="
DefaultCipherScheme = 0
IpPassMode = 0
LogLevel = 4
AuthServiceId = "example"
ResourceIds = ["demo"]
FilterMode = 0

# http server config

# EnableHttp: true: turn on http server, false: shutdown http server.
# EnableTLS: whether to use TLS certificates for hosting https server.
# TLSCertFile: certificate file path.
# TLSKeyFile: key file path.
# to update http changes, you need to restart the http server by changing "EnableHttp" to "false" and then switch it back to "true".
[HttpConfig]
EnableHttp = true
EnableTLS = true
HttpListenPort = 62206
TLSCertFile = "cert/cert.pem"
TLSKeyFile = "cert/cert.key"


# list the server peers for the AC under [[Servers]] table

# Hostname: the domain of the server peer. If specified, it overrides the "Ip" field with its first resolved address.
# Ip: specify the ip address of the server peer
# Port: specify the port number of this server peer is listening
# PubKeyBase64: public key of the server peer in base64 format
# ExpireTime (epoch timestamp in seconds): peer key validation will fail when it expires.
[[Servers]]
Hostname = ""
Ip = "172.16.2.15"
Port = 62206
PubKeyBase64 = "vfAyhQfS1Z+gE7aKSqMCw8GJlZOnw7G7OEG6dHxowtPORn9vqCPp3RqKuyBDZeVqWAMFaCjBUlfu9TpQeN1/uA=="
ExpireTime = 1924991999
```

本地文件支持动态变更的内容在远端配置环境下同样支持动态变更。
