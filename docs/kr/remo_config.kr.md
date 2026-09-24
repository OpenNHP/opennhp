# OpenNHP 원격 설정

## 1 OpenNHP 원격 설정 설명

OpenNHP 설정 정보를 통합 관리하기 위해, OpenNHP는 etcd를 통합 원격 설정 센터로 사용합니다.

OpenNHP 3대 핵심 구성 요소의 설정 파일 지원 방식:

| 구성 요소  | 원격 설정                        | 로컬 설정                        |
| ---------- | ------------------------------- | ------------------------------- |
| nhp server | <font color="green">지원</font> | <font color="green">지원</font> |
| nhp ac     | <font color="green">지원</font> | <font color="green">지원</font> |
| nhp agent  | <font color="red">미지원</font> | <font color="green">지원</font> |

## 2 etcd 환경 배포

### 2.1 etcd 배포

- etcd 다운로드 주소: https://github.com/etcd-io/etcd/releases/
- 서버 환경에 맞는 etcd 버전을 다운로드합니다.
- etcd 서비스 배포 및 시작 파라미터는 설치 패키지 내의 README.md 파일을 참고하세요.

### 2.2 etcd 시각화 설정 도구

- etcd 서비스를 시작한 후에는 etcd의 시각화 도구를 통해 OpenNHP의 설정 정보를 편집할 수 있습니다. 본 문서에서는 etcdkeeper 도구를 예로 들어 etcd 사용자 정의 설정 환경을 구성합니다.

- etcdkeeper를 etcd가 배포된 서버로 다운로드합니다. 다운로드 방법:

  ```sh
  wget https://github.com/evildecay/etcdkeeper/releases/download/v0.7.6/etcdkeeper-v0.7.6-linux_x86_64.zip
  unzip etcdkeeper-v0.7.6-linux_x86_64.zip
  ```

- etcdkeeper 시작 방법:

  - IP는 서버의 실제 IP로 설정합니다.
  - 포트는 실제 사용할 포트로 설정합니다.
  - 시작에 성공하면 브라우저를 통해 접근할 수 있으며, 접근 주소는 예를 들어 http://192.168.32.30:8800 과 같습니다.

  ```sh
  cd etcdkeeper
  chmod +x etcdkeeper
  ./etcdkeeper -h 192.168.32.30 -p 8800 
  ```

## 3 nhp server 원격 설정

### 3.1 원격 설정 접근 구성

etc 디렉터리 아래의 remote.toml은 nhp server 서비스가 원격 설정 센터인 ETCD에 접근하기 위한 설정 정보입니다.

- Endpoints: etcd 접근 주소
- Key: nhp server가 자신의 서버에 해당하는 key를 가져오는 데 사용하는 값

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


### 3.2 원격 설정 활성화

설정 파일 ./nhp-server/etc/remote.toml 파일이 존재하는지 확인하고, Endpoints와 Key가 올바르게 설정되어 있는지 확인하세요. nhp server 서비스가 시작될 때 remote.toml 파일을 로드하여 파일 내 Key에 해당하는 설정 내용을 가져옵니다.



로컬 파일 방식의 설정을 계속 사용하고 싶다면, remote.toml 파일을 제거하거나 remote.toml 내의 Endpoints 또는 Key 내용을 비워두면 서비스가 로컬 설정 파일을 사용하여 시작됩니다.



### 3.3 원격 설정 내용

nhp server 원격 설정 내용은 다음 일곱 부분으로 구성됩니다:

- [BaseConfig] 부분의 내용은 로컬 설정 파일 config.toml에 대응합니다.
- [HttpConfig] 부분의 내용은 로컬 설정 파일 http.toml에 대응합니다.
- [[ACs]] 부분의 내용은 로컬 설정 파일 ac.toml에 대응합니다.
- [[Agents]] 부분의 내용은 로컬 설정 파일 agent.toml에 대응합니다.
- [[DBs]] 부분의 내용은 로컬 설정 파일 db.toml에 대응합니다.
- [[AuthServiceId]] 부분의 내용은 로컬 설정 파일 resource.toml에 대응합니다.
- [[SrcIps]] 부분의 내용은 로컬 설정 파일 srcip.toml에 대응합니다.

원격 설정 정보는 다음과 같습니다:

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



로컬 파일에서 동적 변경을 지원하는 내용은 원격 설정 환경에서도 동일하게 동적 변경을 지원합니다.

## 4 nhp ac 원격 설정

### 4.1 원격 설정 접근 구성

3.1절을 참조하세요.

### 4.2 원격 설정 활성화

3.2절을 참조하세요.



### 4.3 원격 설정 내용

nhp ac 원격 설정 내용은 다음 세 부분으로 구성됩니다:

- [BaseConfig] 부분의 내용은 로컬 설정 파일 config.toml에 대응합니다.
- [HttpConfig] 부분의 내용은 로컬 설정 파일 http.toml에 대응합니다.
- [[Servers]] 부분의 내용은 로컬 설정 파일 server.toml에 대응합니다.

원격 설정 정보는 다음과 같습니다:

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

로컬 파일에서 동적 변경을 지원하는 내용은 원격 설정 환경에서도 동일하게 동적 변경을 지원합니다.
