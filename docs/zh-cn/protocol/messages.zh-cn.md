---
layout: page
title: 消息类型
parent: 协议参考
nav_order: 2
permalink: /zh-cn/protocol/messages/
description: NHP 与 DHP 全部消息类型的参考：方向、载荷字段。
---

# NHP 消息类型
{: .fs-9 }

每个数据包在消息头中携带一个 2 字节的消息类型 ID，接收方据此将其
分发到相应的处理逻辑。目前共定义 28 个 ID：其中 17 个属于 **NHP**
（ID 0–16），涵盖敲门/鉴权/访问流程、AC 生命周期、中继、注册、OTP 与
显式退出；另有 11 个属于 **DHP**（ID 17–27），由数据对象隐身协议使用。
{: .fs-6 .fw-300 }

*对应 CSA《Stealth Mode SDP》白皮书 §NHP Message Types (Table 4) 与 Appendix 2。ID 与字符串名定义见 [`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go)；载荷结构体见 [`nhp/common/nhpmsg.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/nhpmsg.go)。*

## 消息封装

每种消息都共用同一个消息头（参见 [消息头]({{ '/zh-cn/protocol/header/' | relative_url }})），
后跟加密后的载荷。载荷格式为 JSON，可选 zlib 压缩（由 `NHP_FLAG_COMPRESS`
标志控制），再使用解析消息头时派生出的会话密钥做 AEAD 加密 —— 除非
类型说明另作规定（例如 `NHP-KPL` 没有消息体；`NHP-RLY` 承载一个原
始内层数据包）。

## 索引

NHP —— 网络基础设施隐身协议

| ID | 名称 | 方向 | 用途 |
|---:|---|---|---|
| 0 | [NHP-KPL](#nhp-kpl--keepalive) | Agent ↔ Server, AC ↔ Server | 保活包。空消息体。 |
| 1 | [NHP-KNK](#nhp-knk--knock) | Agent → Server | 敲门：请求访问被保护资源。 |
| 2 | [NHP-ACK](#nhp-ack--acknowledge) | Server → Agent | 敲门回应。成功时携带资源地址与会话参数。 |
| 3 | [NHP-AOP](#nhp-aop--ac-operation) | Server → AC | 指示 AC 为某 agent → 资源流打开（或拒绝）访问。 |
| 4 | [NHP-ART](#nhp-art--ac-result) | AC → Server | AC 对 `NHP-AOP` 的响应。 |
| 5 | [NHP-LST](#nhp-lst--list) | Agent → Server | 请求查询当前 agent 有权访问的服务列表。 |
| 6 | [NHP-LRT](#nhp-lrt--list-result) | Server → Agent | 对 `NHP-LST` 的响应。 |
| 7 | [NHP-COK](#nhp-cok--cookie) | Server → Agent | 服务器过载时下发的限流 / 反 DDoS cookie。 |
| 8 | [NHP-RKN](#nhp-rkn--re-knock) | Agent → Server | 二次敲门，HMAC 计算中加入 `NHP-COK` 的 cookie。 |
| 9 | [NHP-RLY](#nhp-rly--relay) | Relay → Server | NHP-Relay 转发的原始 NHP 包，保留源地址信息。 |
| 10 | [NHP-AOL](#nhp-aol--ac-online) | AC → Server | AC 向控制面通告在线状态与所守护的资源。 |
| 11 | [NHP-AAK](#nhp-aak--ac-acknowledge) | Server → AC | 服务器确认 AC 注册，并回显 AC 的公网 IP/端口。 |
| 12 | [NHP-OTP](#nhp-otp--one-time-password) | Agent → Server | 请求带外下发一次性密码用于注册。 |
| 13 | [NHP-REG](#nhp-reg--register) | Agent → Server | Agent 注册其静态公钥，使用 OTP 作为身份凭据。 |
| 14 | [NHP-RAK](#nhp-rak--register-acknowledge) | Server → Agent | `NHP-REG` 的成功或错误回执。 |
| 15 | [NHP-ACC](#nhp-acc--access) | Agent → AC | Agent 向 AC 的临时监听端口出示访问令牌。 |
| 16 | [NHP-EXT](#nhp-ext--exit) | Agent → Server | 请求提前关闭一条活跃会话。空消息体。 |

DHP —— 数据对象隐身协议 *（此处列出是为索引表的完整性；详细 DHP 语义将由专门的 DHP 参考页面展开）*

{: .note }
DHP 条目暂以 Go 常量名展示（如 `NHP_DRG`），尚未配有逐类型子章节。
后续将新增专门的 DHP 参考页面，为每行提供与上表相同的载荷字段表。
**线上名称差异：** [`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go)
中 DHP 字符串标签并不统一 —— ID 17–26 输出为 `NHP_DRG` … `NHP_DBA`
（下划线、NHP 前缀，与上方 NHP 组使用连字符的命名不同），而 ID 27 输
出为 `"DHP-KNK"`（连字符、DHP 前缀，其 Go 常量为 `DHP_KNK`）。使用
`HeaderTypeToString` 输出的实现应精确匹配下表中给出的字符串。

| ID | 常量 | 方向 | 用途 |
|---:|---|---|---|
| 17 | `NHP_DRG` | DB → Server | 将一个数据对象注册到 NHP-Server。 |
| 18 | `NHP_DAK` | Server → DB | 对 `NHP_DRG` 的确认 / 错误码。 |
| 19 | `NHP_DAR` | DHP-Agent → Server | 申请访问某个已注册数据对象。 |
| 20 | `NHP_DAG` | Server → DHP-Agent | 对 `NHP_DAR` 的授权决定。 |
| 21 | `NHP_DSA` | Server → DHP-Agent | 要求提交自证明（TEE / evidence）。 |
| 22 | `NHP_DAV` | DHP-Agent → Server | 向服务器提交远程证明。 |
| 23 | `NHP_DWR` | Server → DB | 向 DB 索取数据对象私钥的包裹（wrap）。 |
| 24 | `NHP_DWA` | DB → Server | DB 返回包裹后的密钥（Key Access Object）。 |
| 25 | `NHP_DOL` | DB → Server | DB 在线 / 状态通告。 |
| 26 | `NHP_DBA` | Server → DB | 对 `NHP_DOL` 的确认。 |
| 27 | `DHP_KNK` | DHP-Agent → Server | DHP 风格的敲门包，携带远程证明证据。 |

---

## NHP-KPL — Keepalive {#nhp-kpl--keepalive}

**ID：** `0` · **方向：** 双向（Agent ↔ Server，AC ↔ Server） · **消息体：** 空（仅消息头）

用于 Agent/AC 与 Server 之间维持 NAT 绑定 / TCP 连接。接收方除了接收
该包之外不做其他处理。Relay 节点 **不** 转发保活包。

## NHP-KNK — Knock {#nhp-knk--knock}

**ID：** `1` · **方向：** Agent → Server · **载荷结构体：** `common.AgentKnockMsg`

| JSON 键 | 字段 | 说明 |
|---|---|---|
| `headerType` | 消息类型 | 回显线上消息类型。 |
| `usrId` | 用户 ID | 区分发起请求的自然人或服务主体。 |
| `devId` | 设备 ID | 支持多设备 / 终端策略验证。 |
| `orgId` | 组织 ID | 可选，租户范围。 |
| `aspId` | ASP ID | 指明由哪个授权服务提供方进行鉴权。 |
| `resId` | 资源 ID | 所请求的被保护资源（域名、服务名或自定义字符串）。 |
| `results` | 校验结果 | 可选 `{checkID: result}` 映射，反映客户端的终端环境检测结果。 |
| `usrData` | 用户数据 | 可选，透传给 ASP 插件使用的自由字段。 |

敲门包是 NHP 交互从冷启动开始时的 **唯一** 触发报文。

## NHP-ACK — Acknowledge {#nhp-ack--acknowledge}

**ID：** `2` · **方向：** Server → Agent · **载荷结构体：** `common.ServerKnockAckMsg`

| JSON 键 | 字段 | 说明 |
|---|---|---|
| `errCode` / `errMsg` | 错误码 / 信息 | `"0"`（或空）表示成功；其他值为失败原因。 |
| `resHost` | 资源主机 | 资源名 → 主机地址映射。 |
| `opnTime` | 放行时长 | 本次访问授权的有效秒数。 |
| `aspToken` | ASP 令牌 | AC 侧后端二次验证使用（可选）。 |
| `agentAddr` | Agent 地址 | 服务器观察到的 Agent 公网元组。 |
| `acTokens` | AC 令牌 | 资源名 → 短期令牌，用于配合 `NHP-ACC` 使用。 |
| `preActions` | 预访问动作 | 可选，每资源一条 `PreAccessInfo`（AC IP、端口、公钥、令牌、密码套件），供需要 `NHP-ACC` 阶段的部署使用。 |
| `redirectUrl` | 重定向 URL | 可选；例如重定向到 HTTPS 登录流程而非直接连接。 |

## NHP-AOP — AC Operation {#nhp-aop--ac-operation}

**ID：** `3` · **方向：** Server → AC · **载荷结构体：** `common.ServerACOpsMsg`

| JSON 键 | 字段 | 说明 |
|---|---|---|
| `usrId` / `devId` / `orgId` | Agent 身份 | 归属字段。 |
| `aspId` / `resId` | ASP + 资源 | 与原始敲门相关联。 |
| `srcAddrs` | 源地址列表 | AC 应放行的 `NetAddress` 数组。 |
| `dstAddrs` | 目的地址列表 | 被保护资源的元组。 |
| `opnTime` | 放行时长 | 秒数；`0` 表示 **拒绝** / 关闭。 |

NAT 提示：在 CGNAT 或共享出口场景下，源 IP 在不同 Agent 之间并不唯一。
部署方应在 AC 与被保护资源之间再叠加一层应用层令牌/cookie（`NHP-ACC`
路径正是如此），或尽量采用端到端 IPv6。

## NHP-ART — AC Result {#nhp-art--ac-result}

**ID：** `4` · **方向：** AC → Server · **载荷结构体：** `common.ACOpsResultMsg`

携带 `errCode` / `errMsg`、实际放行时长 `opnTime`（0 表示拒绝）、AC 分发
的 `token`，以及可选 `preAct`（`PreAccessInfo`）。服务器仅在收到 ART 之
后，才会向 Agent 回复 NHP-ACK。

## NHP-LST — List {#nhp-lst--list}

**ID：** `5` · **方向：** Agent → Server · **载荷结构体：** `common.AgentListMsg`

携带 `usrId`、`devId`、可选 `orgId`、`aspId` 以及自由的 `usrData`。

## NHP-LRT — List Result {#nhp-lrt--list-result}

**ID：** `6` · **方向：** Server → Agent · **载荷结构体：** `common.ServerListResultMsg`

携带 `errCode` / `errMsg` 和一个 `list` 映射，其具体 schema 由 ASP 插件决定。

## NHP-COK — Cookie {#nhp-cok--cookie}

**ID：** `7` · **方向：** Server → Agent · **载荷结构体：** `common.ServerCookieMsg`

携带服务端回显的 `trxId` 与服务端生成的 `cookie`。在服务器过载时下发。
Agent 必须随后发送 NHP-RKN，并在 HMAC 计算中加入此 cookie，以证明一
次完整的往返并避免被提前丢弃。

## NHP-RKN — Re-Knock {#nhp-rkn--re-knock}

**ID：** `8` · **方向：** Agent → Server · **载荷：** 与 NHP-KNK 相同。

与 NHP-KNK 的区别：除普通链式密钥外，HMAC 还会拌入 NHP-COK 的
cookie（参见 [`nhp/core/initiator.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/initiator.go)
中的 `addHMAC(sumCookie: true)`）。

## NHP-RLY — Relay {#nhp-rly--relay}

**ID：** `9` · **方向：** Relay → Server · **载荷：** 来自源端的原始内层 NHP 数据包。

保留穿越 Relay 的源地址；其它消息类型不需要这一点（会被透明转发）。

## NHP-AOL — AC Online {#nhp-aol--ac-online}

**ID：** `10` · **方向：** AC → Server · **载荷结构体：** `common.ACOnlineMsg`

携带 `aspId`、所守护的 `resIds` 列表，以及可选 `acId`。用于向控制面通告 AC 上线。

## NHP-AAK — AC Acknowledge {#nhp-aak--ac-acknowledge}

**ID：** `11` · **方向：** Server → AC · **载荷结构体：** `common.ServerACAckMsg`

携带 `errCode` / `errMsg` 以及 `acAddr`。确认 AC 注册，并回显 AC 的公网地址（对处于 NAT 后端、需要从服务器获知自身外部元组的 AC 尤其有用）。

## NHP-OTP — One-Time Password {#nhp-otp--one-time-password}

**ID：** `12` · **方向：** Agent → Server · **载荷结构体：** `common.AgentOTPMsg`

携带 `usrId`、`devId`，可选 `orgId`、`aspId`、预共享 `pass`，以及自由的
`usrData`。触发 ASP 插件通过自身带外通道下发 OTP（短信、邮件、二维
码）。服务器不会单独回复一个专用类型；ASP 直接从侧通道投递 OTP 即
视为成功。

## NHP-REG — Register {#nhp-reg--register}

**ID：** `13` · **方向：** Agent → Server · **载荷结构体：** `common.AgentRegisterMsg`

携带 `usrId`、`devId`、可选 `orgId`、`aspId`、经带外取得的 `otp`，以及自由的
`usrData`。将 Agent 的静态公钥（存放在加密消息头中）与其身份绑定注册。

## NHP-RAK — Register Acknowledge {#nhp-rak--register-acknowledge}

**ID：** `14` · **方向：** Server → Agent · **载荷结构体：** `common.ServerRegisterAckMsg`

携带 `errCode`（`"0"` 表示成功）、`errMsg`、`aspId`。注册过程中的失败会在此以非零 `errCode` 显式上报，详见 [`HandleRegisterRequest`](https://github.com/OpenNHP/opennhp/blob/main/endpoints/server/msghandler.go)。

## NHP-ACC — Access {#nhp-acc--access}

**ID：** `15` · **方向：** Agent → AC · **载荷结构体：** `common.AgentAccessMsg`

携带 `usrId`、`devId`、可选 `orgId`、来自 NHP-ACK 的 `acToken`（取自
`acTokens`）以及 `usrData`。当部署使用每会话临时端点（`PreAccessInfo`）
而非长期放行清单时，Agent 会直接向 AC 的短期监听器出示本消息。AC 校
验令牌后，为所观察到的源地址安装防火墙放行规则，随即关闭临时端口 ——
参见 [`tcpTempAccessHandler` / `udpTempAccessHandler`](https://github.com/OpenNHP/opennhp/blob/main/endpoints/ac/msghandler.go)。
此路径 **不** 产生任何 NHP 层响应；成功隐式地表现为随后的数据面连接
能够成功到达被保护资源。`common.ACAccessAckMsg` 结构体已定义但当前未
在发送路径上使用，视为预留字段。

## NHP-EXT — Exit {#nhp-ext--exit}

**ID：** `16` · **方向：** Agent → Server · **消息体：** 空。

Agent 主动请求提前关闭一条活跃会话。服务器随即对 AC 发送
`opnTime = 0` 的 NHP-AOP，关闭相应放行规则。

---

## DHP 消息类型 {#dhp}

DHP 复用了 NHP 的线上格式来承载一组独立流程：数据对象注册、访问、
远程证明、密钥包裹等。这里仅列出它们是为了让 ID 表完整；各类型的
载荷字段（`DRGMsg`、`DAKMsg`、`DARMsg`、`DAGMsg`、`DSAMsg`、
`DAVMsg`、`DWRMsg`、`DWAMsg`、`DBOnlineMsg`、`ServerDBAckMsg`、
`DHPKnockMsg`）定义在
[`nhp/common/nhpmsg.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/nhpmsg.go)。
专门的 DHP 参考页面将把这些展开；暂请参见
[DHP 快速开始]({{ '/zh-cn/dhp_quick_start/' | relative_url }})。

---

## 另见

- [消息头]({{ '/zh-cn/protocol/header/' | relative_url }}) —— 所有类型共享的信封
- [加密算法]({{ '/zh-cn/cryptography/' | relative_url }}) —— 载荷如何被加密
- [术语表]({{ '/zh-cn/glossary/' | relative_url }}) —— 本页涉及到的各角色的规范命名
