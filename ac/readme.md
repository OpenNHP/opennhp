# NHP-AC

## 架构设计

1. AC支持与多台Server互相进行通信。所有连接均为AC主动向Server发起。AC通过发送心跳包和NHP-AOL包维持与Server的连接。

2. AC与Server通信失效后，将尝试重新建立连接，如果一直无法与任何一台Server建立连接，则进入失效状态。

3. AC在启动后即开始与预设的服务器周期性建立连接并保持连接（AC很有可能在内网，所以不能由服务器先发连接）。连接时发送NHP_DOL消息，在收到服务器的回应后确认连接。连接期间视情况进行发送NHP_KPL消息保持连接。由**maintainServerConnectionRoutine**实现。

4. AC处理服务器发送过来的NHP_DOP消息，判断请求方的serviceId, appId是否匹配并进行IPSET操作，完成后返回NHP_DRT消息。

5. 发送消息时，向**sendMsgCh**发送创建好的**MsgAssembler**（必须指定**RemoteAddr**）。如果连接没有建立，AC会尝试建立并记录该连接。同时对此连接开启接收线程。**MsgAssembler**经过加密后会从此连接发出

6. 接收到报文时，会将报文进行解密获取明文消息。由**msghandler**分别进行处理。

## NHP-AC的IP放行模式

IP放行模式分为两种：

1. ipPassMode为0（默认）时为立即放行模式，门禁开门时将以敲门来源IP地址为准。

2. ipPassMode为1时为预访问模式，门禁开门前将先开启对应协议的临时端口并返回server临时端口和临时访问token，在短时间内需由agent携带临时访问token进行临时连接，如果临时连接有效，则开门时放行将以此次临时连接的来源IP为准。

## NHP-AC配置文件

`etc/config.toml`

```toml
[Door]
  # (optional) assign an unique id for this door
  DoorId = "abc_group_door_001"
  # (mandatory) specify the private key in base64 format
  DoorPrivateKey = "+B0RLGbe+nknJBZ0Fjt7kCBWfSTUttbUqkGteLfIp30="
  # 0: default, passing the knock source IP
  # 1: use pre-access procedure to determine the passing source IP
  IpPassMode = 0
  # (optional) ID of authorization service provider this door belongs to
  AuthServiceId = "abc_group" 
  # (optional) ID of resources controlled by this door
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
