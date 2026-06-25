# Agent Register — 现状分析与实现方案

> 分支：`feature/agent-register` | 日期：2026-06-24 | 基于 main 分支

## 1. 需求概述

依据 NHP 协议规范（T/CCF 0002—2024），完善 NHP 协议的 Agent Register 功能，涉及以下消息类型：

| 消息 | 类型码 | 方向 | 用途 |
| --- | --- | --- | --- |
| **NHP-OTP** | 12 | Agent → Server | 申请一次性验证码（可选前置步骤） |
| **NHP-REG** | 13 | Agent → Server | Agent 向 Server 注册公钥 |
| **NHP-RAK** | 14 | Server → Agent | 确认密钥注册成功 |

同时支持两种密钥分发体系：
- **PKI 模式**：传统公钥基础设施，Agent 通过 NHP-REG 直接向 NHP-Server 注册公钥
- **IBC/CL-PKC 模式**：基于标识的密码体系 / 无证书密钥体系，通过 KGC 进行密钥分发

### 1.1 设计决策（已明确）

| 决策项 | 决定 | 说明 |
| --- | --- | --- |
| OTP 分发渠道 | **Email** | ASP 插件通过邮件发送一次性验证码 |
| 密钥生成方 | **Agent 优先，Relay 兜底** | Agent 自行生成密钥对（如 js-agent 在浏览器端生成）；若环境受限无法生成，通过 NHP-Relay 提供的 API 代为生成 |
| 公钥存储 | **SQLite** | NHP-Server 使用 SQLite 持久化已注册 Agent 的公钥 |

## 2. 当前实现状态

### 2.1 已实现（基础框架）

#### 数据包类型定义
- `nhp/core/packet.go:28-30` — `NHP_OTP`(12)、`NHP_REG`(13)、`NHP_RAK`(14) 均已定义
- `nhp/core/packet.go:85-104` — `HeaderTypeToDeviceType` 映射完整：REG/OTP→AGENT, RAK→SERVER
- `nhp/core/transaction.go:34,39,87,93` — REG/RAK 已注册为事务请求/响应对

#### 消息结构
- `nhp/common/nhpmsg.go:31-53` — `AgentOTPMsg`、`AgentRegisterMsg`、`ServerRegisterAckMsg` 结构体完整
- `nhp/common/types.go:47-57` — `NhpOTPRequest`、`NhpRegisterRequest` 包装类型完整

#### 服务端处理
- `endpoints/server/udpserver.go:938-942` — `recvMessageRoutine` 正确分派 NHP_OTP、NHP_REG
- `endpoints/server/msghandler.go:169-204` — `HandleOTPRequest`：解析消息 → 查插件 → 调用 `RequestOTP()`，不发回响应
- `endpoints/server/msghandler.go:208-276` — `HandleRegisterRequest`：解析消息 → 提取公钥 → 查插件 → 调用 `RegisterAgent()` → 发送 NHP_RAK 响应

#### 客户端请求
- `endpoints/agent/request.go:14-75` — `RequestOtp()`：构造 `AgentOTPMsg` → 发送 NHP_OTP → 发后即忘
- `endpoints/agent/request.go:77-163` — `RegisterPublicKey()`：构造 `AgentRegisterMsg` → 发送 NHP_REG → 阻塞等待 NHP_RAK → 返回结果

#### 插件接口
- `nhp/plugins/serverpluginhandler.go:23-24` — `RequestOTP`、`RegisterAgent` 在 `PluginHandler` 接口中已定义
- `nhp/plugins/serverpluginhandler.go:36-41` — 符号表包含 `sRequestOTP`、`sRegisterAgent` 用于反射加载

#### KGC 基础设施
- `endpoints/kgc/` — 完整的 CL-PKC 密钥生成逻辑（SM2 曲线 + SM3 哈希）
- 支持 `setup`（生成主密钥）、`keygen`（生成用户密钥）、`sign`、`verify` 四个 CLI 命令

#### 协议包头 IBC 支持
- `nhp/core/scheme/curve/header.go:17-24` — Curve25519 包头预留 64+16 字节 IBC Identity 密文槽位
- `nhp/core/scheme/gmsm/header.go:17-24` — GMSM 包头同样预留
- `nhp/common/packet.go:7` — `NHP_FLAG_CL_PKC` 包头标志位已定义（bit 2）
- `nhp/core/initiator.go:32,387-389` — `MsgData.ClPkc` 标志可传递至包头 `NHP_FLAG_CL_PKC`

### 2.2 缺失 / 待实现

#### A. 插件侧实现为空
示例插件（`examples/server_plugin/basic/main.go`）**未导出** `RequestOTP`、`RegisterAgent`、`ListService` 方法。当前 basic 插件只实现了 `AuthWithNHP` 和 `AuthWithHttp`，导致 REG/OTP 流程会因 `handler == nil`（找不到 ASP 对应的插件）而直接返回 `ErrAuthHandlerNotFound`。

其他插件（authenticator、oidc）同样未实现这三个方法。

#### B. KGC 仅为离线 CLI 工具
当前 `endpoints/kgc/` 是纯命令行工具，没有网络服务能力。协议规范中描述的「NHP 代理向 KGC 服务器发起 NHP-OTP / NHP-REG」场景无法实现。

#### C. 包头 IBC Identity 字段未填充
- `nhp/core/scheme/curve/curve.go:75` — `Identity()` 返回 `nil`
- `nhp/core/scheme/gmsm/gmsm.go:88` — `Identity()` 返回 `nil`
- `MidPublicKey()` 在两个方案中都返回 `nil`

这意味着即使设置了 `NHP_FLAG_CL_PKC` 标志，包头中 80 字节的 IBC Identity 密文槽位始终为零值，IBC/CL-PKC 模式实际不可用。

#### D. REG/OTP 流程未与上层调用链集成
Agent 端的 `RequestOtp()` 和 `RegisterPublicKey()` 方法已实现，但有两点缺失：
1. **无上层调用者**：`UdpAgent` 没有公开的 API 方法供外部（如 HTTP API、CLI、js-agent）触发 OTP 申请和注册流程
2. **无自动触发逻辑**：首次连接时若服务器返回「未注册」错误，Agent 不会自动发起注册流程

#### E. 服务器端缺少公钥存储机制
`HandleRegisterRequest` 将公钥传给插件的 `RegisterAgent` 方法后，插件需要持久化该公钥。当前没有定义：
- 公钥存储接口（数据库 / 文件 / etcd）
- 已注册 Agent 的 peer 表管理

#### F. 协议规范中的 IBC 注册字段未实现
协议规范表 D.4 定义了 IBC 体系标识公钥注册的组成字段（NHP 类型标志位、用户可读标识、设备唯一标识、随机数、有效时长），表 D.5 定义了 KGC 分发的用户数字证书字段。当前实现中这些均不存在。

## 3. 实现方案

### 3.1 总体架构

```
┌─────────┐  NHP-OTP(12)  ┌──────────┐  HTTP/短信/邮件  ┌─────┐
│  Agent  │ ─────────────→│  Server  │ ───────────────→│ ASP │
│         │               │          │                  │     │
│         │  NHP-REG(13)  │          │                  └─────┘
│         │ ─────────────→│          │
│         │               │          │
│         │  NHP-RAK(14)  │          │
│         │ ←─────────────│          │
└─────────┘               └──────────┘

PKI 模式：Agent ←→ Server（直接注册公钥，ASP 验证 OTP）
IBC 模式：Agent ←→ KGC（通过 KGC 进行标识密钥注册与分发）
```text

### 3.2 分阶段实施

#### 阶段一：PKI 模式 — 完善 REG/OTP/RAK 核心链路（本次）

**目标**：使 PKI 模式下的 Agent 注册流程端到端可用。

**设计前提**：
- OTP 由 ASP 插件通过 **Email** 发送
- Agent 密钥对 **优先由 Agent 自行生成**（如 js-agent 可在浏览器端使用 Web Crypto API 生成）；若 Agent 环境受限无法生成，可通过 HTTP API 由 NHP-Server 代为生成
- 已注册 Agent 公钥使用 **SQLite** 持久化存储

**工作项**：

1. **实现 ASP 插件中的 `RequestOTP` 方法 — Email 发送**
   - 文件：`examples/server_plugin/basic/main.go`
   - 功能：接收 OTP 请求 → 生成随机验证码（6位数字）→ 通过 SMTP 发送邮件给用户
   - 配置：SMTP 服务器地址、端口、认证信息、发件人地址、邮件模板
   - 存储：OTP 与 (userId, deviceId) 关联，设置有效期（建议 5 分钟），存储于 SQLite

2. **实现 ASP 插件中的 `RegisterAgent` 方法 — 验证 + 公钥登记**
   - 文件：同上
   - 功能：
     - 验证 OTP 有效性（匹配 userId/deviceId，未过期）
     - 将 Agent 公钥（由 Agent 在 NHP-REG 包头中携带）与 (userId, deviceId) 关联写入 SQLite
     - 返回 `ServerRegisterAckMsg`

3. **NHP-Server 端 SQLite 公钥存储**
   - 文件：`endpoints/server/` 新增 `keystore.go`
   - 功能：基于 SQLite 的 Agent 公钥表 CRUD + OTP 管理
   - Schema 见 3.5 节

4. **NHP-Relay 端密钥生成 API（兜底方案）**
   - 文件：`endpoints/relay/` 新增 `keygen_api.go`
   - 功能：提供 HTTP API 供无法自行生成密钥的 Agent 调用，生成 Curve25519 密钥对并返回
   - 注意：这是辅助功能，不影响核心注册流程

5. **js-agent 注册页面**
   - 文件：`endpoints/js-agent/examples/reg.html`（**新文件**）
   - 入口：`https://reg.opennhp.org`
   - 功能：
     - 用户填写 userId、deviceId、Email 地址
     - 点击「获取验证码」→ js-agent 本地生成 Curve25519 密钥对 → 发送 NHP-OTP 请求 → Server/ASP 发送 Email OTP
     - 用户输入收到的 OTP → 点击「注册」→ 发送 NHP-REG（包头携带公钥）
     - 注册成功后将私钥保存到本地（localStorage 或下载为文件）
   - 技术栈：纯 HTML + ES Module，与现有 `relay-test.html` 风格一致，复用 `@opennhp/agent` SDK
   - SDK 需扩展：暴露 `requestOtp()` 和 `registerPublicKey()` 方法（当前仅在内部实现）

6. **错误码完善**
   - 文件：`nhp/common/errors.go`
   - 新增：
     - `ErrOTPInvalid` — OTP 验证失败
     - `ErrOTPExpired` — OTP 已过期
     - `ErrPublicKeyAlreadyRegistered` — 公钥已被其他用户注册
     - `ErrAgentAlreadyRegistered` — 该设备已注册（保留，用于上层语义区分）
     - `ErrAgentNotRegistered` — Agent 未注册即尝试敲门

7. **js-agent SDK 方法暴露**
   - 文件：`endpoints/js-agent/src/index.ts`、`NHPAgent.ts`
   - 新增公开方法：
     - `requestOtp(target)` → 发送 NHP-OTP，返回 void
     - `registerPublicKey(target, otp)` → 发送 NHP-REG，返回注册结果
   - 当前这些方法在原生 Agent 中已实现（`request.go`），js-agent 侧需补齐对应 TypeScript 实现

#### 3.2.1 Demo 部署变更

新增域名 `reg.opennhp.org` 承载 js-agent 注册页面，需联动以下基础设施：

| 组件 | 文件 | 变更 |
| --- | --- | --- |
| DNS | `terraform/demo/dns.tf` | 新增 `reg.opennhp.org` CNAME → relay 公网 IP |
| Nginx vhost | `deploy/nginx/reg.conf.template`（**新文件**） | 新增 vhost，root 指向 `/var/www/jsagent/reg/`，与现有 `agent.opennhp.org` 共享 relay EC2 实例 |
| CI/CD | `.github/workflows/deploy-demo-v2.yml` | 在 `deploy-jsagent` job 中增加 `reg.html` 的部署步骤 |
| TLS | 通过 certbot/ACME 自动申请，SAN 扩展 `reg.opennhp.org` | 与 relay 主证书共享或独立证书 |

部署后目录结构（relay EC2）：

```text
/var/www/jsagent/
  ├── index.html          # 现有 relay-test.html（agent.opennhp.org）
  ├── config.json
  ├── reg/
  │   └── index.html      # 注册页面（reg.opennhp.org）
  └── nhp-js/
      └── dist/           # SDK 构建产物
```

#### 阶段二：IBC/CL-PKC 模式（后续）

1. **KGC 网络化改造**
   - 将 KGC 从 CLI 工具改造为 UDP 网络服务（类似 NHP-Server）
   - 支持接收 NHP-OTP、NHP-REG 消息并响应 NHP-RAK

2. **包头 IBC Identity 字段填充**
   - 实现 `Identity()` 和 `MidPublicKey()` 方法
   - 在 CL-PKC 模式下将用户标识公钥加密写入包头 Identity 字段

3. **IBC 注册流程**
   - 实现表 D.4 的标识公钥注册字段
   - 实现表 D.5 的用户数字证书生成与分发
   - 支持 SM9/CPK 算法（国内 IBC 标准）

### 3.3 关键代码改动清单

| 文件 | 改动类型 | 说明 |
| --- | --- | --- |
| `examples/server_plugin/basic/main.go` | 新增方法 | 导出 `RequestOTP`（Email SMTP）、`RegisterAgent`（验证OTP+登记公钥） |
| `endpoints/server/keystore.go` | **新文件** | SQLite Agent 公钥存储：建表、CRUD、OTP 管理 |
| `endpoints/relay/keygen_api.go` | **新文件** | NHP-Relay HTTP API：无法自行生成密钥的 Agent 调用，返回 Curve25519 密钥对（兜底） |
| `nhp/common/errors.go` | 新增 | 注册相关错误码（`ErrOTPInvalid`、`ErrPublicKeyAlreadyRegistered` 等） |
| `endpoints/agent/udpagent.go` | 扩展 | 暴露 `Register()` 公开方法；本地密钥生成（`nhp/core` ECDH） |
| `endpoints/agent/request.go` | 扩展 | `RegisterPublicKey` 使用 Agent 本地生成的公钥（与当前实现一致） |
| `endpoints/server/msghandler.go` | 扩展 | `HandleRegisterRequest` 协调公钥登记+SQLite 存储+RAK 响应 |
| `endpoints/server/config.go` | 扩展 | 新增 `DatabasePath` 配置项 |
| `endpoints/js-agent/src/NHPAgent.ts` | 扩展 | 暴露 `requestOtp()`、`registerPublicKey()` 公开方法 |
| `endpoints/js-agent/examples/reg.html` | **新文件** | Agent 注册页面（`reg.opennhp.org`），含 OTP 申请 + 注册流程 UI |
| `terraform/demo/dns.tf` | 扩展 | 新增 `reg.opennhp.org` DNS 记录 |
| `deploy/nginx/reg.conf.template` | **新文件** | `reg.opennhp.org` Nginx vhost 配置 |
| `.github/workflows/deploy-demo-v2.yml` | 扩展 | 部署 js-agent 注册页面到 relay EC2 |
| `docs/features/agent-register.md` | 本文件 | 现状分析与实现方案 |

### 3.4 PKI 模式注册流程

![PKI Mode Registration Flow](./images/PKI_Mode_Registration_Flow.png)

**兜底流程：Agent 无法自行生成密钥时，通过 NHP-Relay HTTP API 获取**

```text
Agent                              Relay(HTTP API)
  │                                  │
  │  POST /api/keygen                │
  │  {usrId, devId}                  │
  │ ───────────────────────────────→ │
  │                                  │ 生成 Curve25519 密钥对
  │  {publicKey, privateKey}         │
  │ ←─────────────────────────────── │
  │                                  │
  │ （之后继续主流程步骤 1-9）       │
```

### 3.5 SQLite 存储设计

#### 3.5.1 OTP 表

```sql
CREATE TABLE otp_records (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    usr_id     TEXT NOT NULL,
    dev_id     TEXT NOT NULL,
    otp_code   TEXT NOT NULL,
    created_at INTEGER NOT NULL,  -- Unix timestamp (seconds)
    expires_at INTEGER NOT NULL,  -- Unix timestamp (seconds)
    used       INTEGER DEFAULT 0  -- 0=未使用, 1=已使用
);
CREATE INDEX idx_otp_usr_dev ON otp_records(usr_id, dev_id);
```

#### 3.5.2 Agent 公钥表

```sql
CREATE TABLE agent_keys (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    usr_id     TEXT NOT NULL,
    dev_id     TEXT NOT NULL,
    public_key TEXT NOT NULL UNIQUE,  -- 公钥全局唯一，防止密钥复用
    cipher     INTEGER DEFAULT 0,     -- 0=Curve25519, 1=GMSM
    created_at INTEGER NOT NULL,      -- Unix timestamp (seconds)
    expires_at INTEGER,               -- NULL 表示永不过期
    active     INTEGER DEFAULT 1,     -- 0=已吊销, 1=有效
    UNIQUE(usr_id, dev_id)            -- 同一用户+设备唯一，支持一用户多设备
);
CREATE INDEX idx_agent_usr ON agent_keys(usr_id);
CREATE INDEX idx_agent_pubkey ON agent_keys(public_key);
```

#### 3.5.3 注册冲突处理策略

| 场景 | 条件 | 策略 | 错误码 |
| --- | --- | --- | --- |
| 公钥冲突 | 相同 `public_key`，不同 `usr_id` | **拒绝注册** | `ErrPublicKeyAlreadyRegistered` |
| 设备重复注册 | 相同 `usr_id` + `dev_id` | **覆盖更新**（密钥轮换） | 无错误，更新 `public_key` + `created_at` |
| 多设备注册 | 相同 `usr_id`，不同 `dev_id` | **允许** | 无错误，插入新记录 |
| 相同用户+设备+公钥 | 全部相同 | **幂等**，视为已注册 | 返回成功，不更新 |

#### 3.5.4 数据库位置

- 默认路径：`<exe_dir>/data/nhp_server.db`
- 通过 `config.toml` 中的 `DatabasePath` 配置项指定

### 3.6 数据流关键点

- **OTP 流程**：Agent → Server（NHP-OTP）→ Plugin（RequestOTP），Plugin 生成随机验证码、通过 SMTP 发送 Email、将 OTP 写入 SQLite，Server 不做 NHP 响应
- **REG 流程**：Agent → Server（NHP-REG，包头携带 Agent 自行生成的公钥 + 消息体携带 OTP）→ Plugin（RegisterAgent），Plugin 验证 OTP → 公钥写入 SQLite → Server 回复 NHP-RAK 确认
- **密钥生成（主路径）**：Agent 本地生成 Curve25519 密钥对（js-agent 使用 Web Crypto API，原生 Agent 使用 `nhp/core` ECDH）
- **密钥生成（兜底路径）**：Agent 环境受限时，通过 NHP-Relay HTTP API `POST /api/keygen` 生成并返回密钥对
- **公钥传递方向**：Agent → Server（通过 Noise 包头静态公钥），与当前实现一致，无需修改
- **OTP 校验**：Plugin 查询 SQLite 验证 OTP 有效性（匹配 userId/deviceId，未过期，未使用）

### 3.7 协议规范对照

| 规范章节 | 规范内容 | 当前实现状态 | 本次实现 |
| --- | --- | --- | --- |
| D.3.1 NHP-REG | Agent 注册公钥，消息类型13 | 消息结构完整，插件逻辑空 | 实现插件 `RegisterAgent` |
| D.3.2 NHP-OTP | 申请一次性验证码，消息类型12 | 消息结构完整，插件逻辑空 | 实现插件 `RequestOTP` |
| D.3.3 NHP-RAK | 注册确认，消息类型14，消息体可为空 | 消息结构完整，发送逻辑完整 | 无需改动（已完成） |
| D.3.4 IBC 密钥分发 | SM9/CPK 标识密钥注册 | KGC 仅 CLI，包头 IBC 字段空 | 阶段二 |
| D.3.5 CL-PKC | 无证书密钥体系 | KGC 算法完整，网络层缺失 | 阶段二 |

## 4. 风险与注意事项

1. **向后兼容**：`PluginHandlerSymbol` 对所有可选方法已做 nil 检查（`nhp/plugins/serverpluginhandler.go:90-106`），当旧插件未导出 `RequestOTP` / `RegisterAgent` 时，返回 `errPluginNotImplemented` 而非 panic。因此旧插件无需修改即可加载，只是 REG/OTP 功能不可用。无需额外改动

2. **公钥存储安全**：插件中存储的 Agent 公钥表需要防止篡改，生产环境应使用数据库或 etcd 持久化，而非内存存储

3. **OTP 安全**：TOTP 种子需安全存储，OTP 有效期应限制在合理范围内（建议 5 分钟）

4. **重放攻击防护**：包头中的 Counter 字段已提供防重放能力，REG 消息复用此机制

5. **签名提交**：所有 commit 必须使用 GPG/SSH 签名，遵循项目 CLAUDE.md 要求

## 5. 待澄清信息与决策项

以下列出基于当前设计前提（Email OTP、Server 生成密钥、SQLite 存储），在进入实现阶段前**尚缺少的信息**和**需要决策的项**。

### 5.1 待澄清信息

#### I1. Agent 密钥生成路径

主路径：Agent 自行生成密钥对（公钥通过 Noise 包头携带），与当前实现一致。

兜底路径：Agent 环境受限时调用 Server HTTP API 获取密钥对。

```
主路径：  Agent（本地生成密钥）──NHP-REG（包头携带公钥）──→ Server ──登记公钥──→ SQLite
兜底：    Agent ──POST /api/keygen──→ Relay ──生成密钥对──→ 返回公私钥给 Agent
         Agent（使用返回的密钥）──NHP-REG──→ Server ──登记公钥──→ SQLite
```text

**需澄清**：

- js-agent 在浏览器端使用 Web Crypto API 生成 Curve25519 密钥对是否可行？（SubtleCrypto.generateKey 支持 Ed25519 但不直接支持 Curve25519/X25519，可能需要额外库如 `@noble/curves`）
- 兜底 API `POST /api/keygen`（由 NHP-Relay 提供）是否需要鉴权？还是仅依赖 TLS？
- 通过 Relay HTTP API 传输私钥的安全性考量

#### I2. Email 发送的配置与运行环境

- SMTP 服务器的地址、端口、认证方式是固定的还是可配置的？
- 发件人地址和邮件内容模板由谁定义？放在插件配置文件（TOML）中？
- 是否需要支持多种邮件服务（如 AWS SES、阿里云邮件推送）？当前阶段是否只需一个 SMTP 实现即可？

#### I3. Agent 首次注册的触发方式

- Agent 是主动调用 `register` CLI 命令，还是在首次 knock 被拒（未注册）后自动发起注册流程？
- 如果是自动触发，Server 侧如何区分「未注册」和「认证失败」？需要新增错误码如 `ErrAgentNotRegistered`。

#### I4. 注册后的密钥存储（Agent 侧）

- Agent 收到 Server 返回的私钥后，存储在哪里？
  - 内存（重启丢失，需重新注册）
  - 本地文件（如 `agent_key.json`，类似 Server 的 `config.toml` 存储方式）
  - 系统密钥链（macOS Keychain / Linux Secret Service）

#### I5. 已注册 Agent 与现有 Peer 表的关系

当前 Server 的 peer 表通过 `server.toml` 静态配置。已注册 Agent 是动态加入的：

- 注册后的 Agent 公钥是否应写入 `server.toml` 的 peer 列表？还是仅在 SQLite 中维护？
- Knock 认证时，是先查 SQLite 还是先查静态 peer 表？优先级如何？

### 5.2 需决策项

#### D1. NHP-RAK 消息体是否需要扩展

主路径下 Agent 自行生成密钥，NHP-RAK 只需确认注册结果，当前 `ServerRegisterAckMsg`（errCode + errMsg + aspId）已满足需求，**无需扩展**，与协议规范 D.3.3（消息体可为空）一致。

兜底路径的密钥分发通过独立的 HTTP API 完成，不走 NHP 协议通道。

**结论**：NHP-RAK 消息体保持现状，无需改动。

#### D2. SQLite 驱动选择

| 方案 | 驱动 | 特点 |
| --- | --- | --- |
| A | `github.com/mattn/go-sqlite3` | Go 社区标准，cgo 依赖 |
| B | `modernc.org/sqlite` | 纯 Go 实现，无 cgo，交叉编译友好 |

**建议**：方案 B（`modernc.org/sqlite`），项目模块较多且涉及交叉编译（eBPF、ARM），纯 Go 方案构建更简单。

#### D3. OTP 格式

| 方案 | 格式 | 特点 |
| --- | --- | --- |
| A | 6 位随机数字 | 简单，Email 友好 |
| B | TOTP（基于时间） | 无需存储，但需要 shared secret |
| C | 随机字符串（字母+数字） | 安全性更高 |

**建议**：方案 A，6 位数字验证码，与短信验证码体验一致，用户友好。

#### D4. SQLite 数据库文件路径

| 方案 | 路径 | 适用场景 |
| --- | --- | --- |
| A | `<exe_dir>/data/nhp_server.db` | 单实例部署 |
| B | 通过 `config.toml` 配置 `DatabasePath` | 灵活部署 |
| C | 固定系统路径如 `/var/lib/nhp-server/keys.db` | 生产部署 |

**建议**：方案 B，通过配置项指定，默认值用方案 A。符合项目现有配置管理模式（TOML 文件）。

#### D5. Plugin 是否需要直接访问 SQLite

| 方案 | 描述 |
| --- | --- |
| A | Plugin 通过 `NhpServerPluginHelper` 获得 SQLite 操作接口 |
| B | Plugin 只做 OTP 生成/验证逻辑，SQLite 操作由 Server 侧统一完成 |

**建议**：方案 B。Plugin 的职责是 ASP 业务逻辑（生成 OTP、发送 Email、验证 OTP），数据持久化由 Server 统一管理，避免 Plugin 直接接触数据库，保持关注点分离。

### 5.3 总结：进入实现前必须确认的阻塞项

| 序号 | 项 | 类型 | 优先级 | 阻塞 |
| --- | --- | --- | --- | --- |
| 1 | SMTP 配置方式（I2） | 澄清 | P0 | 🔴 阻塞 Plugin 配置设计 |
| 2 | Agent 端密钥存储位置（I4） | 澄清 | P0 | 🔴 阻塞 Agent 端存储实现 |
| 3 | js-agent 浏览器端 Curve25519 生成方案（I1） | 澄清 | P0 | 🔴 影响 js-agent 密钥生成实现 |
| 4 | 注册触发方式（I3） | 澄清 | P1 | 🟡 影响 CLI/API 设计 |
| 5 | SQLite 驱动选择（D2） | 决策 | P1 | 🟡 影响 go.mod 依赖 |
| 6 | OTP 格式（D3） | 决策 | P1 | 🟡 影响 Plugin OTP 逻辑 |
| 7 | SQLite 与 Peer 表关系（I5） | 澄清 | P1 | 🟡 影响 Knock 认证流程 |
| 8 | Relay 兜底 API 鉴权方式（I1） | 澄清 | P1 | 🟡 影响 Relay keygen API 设计 |
| 9 | Plugin 与 SQLite 边界（D5） | 决策 | P2 | 🟢 架构优化，不影响核心功能 |
| 10 | 数据库文件路径（D4） | 决策 | P2 | 🟢 有默认值，可后续调整 |

## 6. 参考文件路径

| 关注点 | 路径 |
| --- | --- |
| 数据包类型定义 | `nhp/core/packet.go:15-45` |
| 消息结构体 | `nhp/common/nhpmsg.go:31-53` |
| 请求包装类型 | `nhp/common/types.go:47-57` |
| 服务端 REG/OTP 处理 | `endpoints/server/msghandler.go:167-276` |
| 客户端 REG/OTP 请求 | `endpoints/agent/request.go:14-163` |
| 插件接口定义 | `nhp/plugins/serverpluginhandler.go:17-28` |
| 协议包头 IBC 字段 | `nhp/core/scheme/curve/header.go:17-24` |
| KGC 实现 | `endpoints/kgc/kgc.go` |
| 示例插件 | `examples/server_plugin/basic/main.go` |
| 协议文档 | `docs/protocol/header.md` |
