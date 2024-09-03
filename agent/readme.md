# NHP-Agent

## 架构设计

1. Agent只与Server之间进行通信。Agent主动向Server发起短连接。不存在Agent在未建立连接时被动接收Server消息的情况。

2. 发送消息时，向**sendMsgCh**发送创建好的**MsgAssembler**（必须指定**RemoteAddr**）。如果连接没有建立，agent会尝试建立并记录该连接。同时对此连接开启接收线程。**MsgAssembler**经过加密后会从此连接发出

3. 接收到报文时，会将报文进行解密获取明文消息。由**msghandler**分别进行处理。

## NHP-AGENT配置文件

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
