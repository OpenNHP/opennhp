[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)
[![zh-tw](https://img.shields.io/badge/lang-zh--tw-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-tw.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.es.md)

![OpenNHP Logo](docs/images/logo11.png)

# OpenNHP：开源零信任安全工具套件

[![Build](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml/badge.svg)](https://github.com/OpenNHP/opennhp/actions/workflows/ubuntu-build.yml)
[![Release](https://img.shields.io/github/v/tag/OpenNHP/opennhp?label=release)](https://github.com/OpenNHP/opennhp/tags)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
[![codecov](https://codecov.io/gh/OpenNHP/opennhp/branch/main/graph/badge.svg)](https://codecov.io/gh/OpenNHP/opennhp)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/OpenNHP/opennhp)

**OpenNHP** 是一个轻量级、基于加密技术的开源工具套件，为基础设施、应用和数据提供零信任安全保障。它是[**云安全联盟（CSA）**](https://cloudsecurityalliance.org/) *[网络基础设施隐藏协议（NHP）规范](https://cloudsecurityalliance.org/artifacts/stealth-mode-sdp-for-zero-trust-network-infrastructure)*的参考实现，包含两项核心协议：

- **网络基础设施隐藏协议（NHP）：** 隐藏服务器端口、IP 地址和域名，防止应用和基础设施被未授权访问。
- **数据内容隐藏协议（DHP）：** 通过加密和机密计算保障数据安全与隐私，实现数据*"可用而不可见"*。

**[官方网站](https://opennhp.org) · [愿景](https://opennhp.org/vision/) · [在线演示](https://opennhp.org/demo/) · [文档](https://docs.opennhp.org) · [Discord](https://discord.gg/CpyVmspx5x)**

---

## 为什么选择 OpenNHP

当今的互联网是一片[黑暗森林](https://en.wikipedia.org/wiki/Dark_forest_hypothesis)。攻击者——越来越多地借助大语言模型（LLM）通过[自主漏洞利用（AVE）](https://arxiv.org/abs/2404.08144)以机器速度进行扫描、指纹识别和漏洞利用——将所有可达的服务都视为攻击目标。[Gartner 预测](https://www.gartner.com/en/newsroom/press-releases/2024-08-28-gartner-forecasts-global-information-security-spending-to-grow-15-percent-in-2025) AI 驱动的网络攻击将快速增长。传统防御仅在*网络放行之后*才对用户进行身份验证，使得暴露的端口、IP 和域名成为永久的攻击面。

> **在 AI 时代，可见性 = 脆弱性。**

OpenNHP 颠覆了这一模式：**验证之前不可见。** 所有端口、IP 和主机名均置于默认拒绝的门之后。只有经过加密签名的"敲门"请求通过带外认证与授权之后，才会开放访问。攻击者无法利用他们发现不了的东西。

### 第三代网络隐藏协议

NHP 是"先隐藏服务"这一设计思路的下一代演进：

| 代际 | 协议 | 局限 |
|---|---|---|
| 1 | 端口敲门（Port Knocking） | 明文传输，易受重放攻击 |
| 2 | 单包授权（SPA） | 共享密钥、单向通信、通常仅隐藏端口、通常使用 C/C++ 实现 |
| **3** | **NHP** | 现代加密、带状态的双向通信、同时隐藏域名 + IP + 端口、无状态可水平扩展、内存安全的 Go 实现 |

NHP 与现有的 IAM、DNS、FIDO 和零信任策略引擎协同工作，而不是取代它们——它是对现有技术栈的扩展，而非分叉。

---

## 架构

OpenNHP 采用模块化设计，包含三个核心组件，灵感来自 [NIST 零信任架构](https://www.nist.gov/publications/zero-trust-architecture)：

![OpenNHP architecture](docs/images/OpenNHP_Arch.gif)

| 核心组件 | 职责 |
|-----------|------|
| **NHP-Agent** | 客户端，发送加密的"敲门"请求以获得访问权限 |
| **NHP-Server** | 认证并授权请求；独立部署，在架构上与受保护主机解耦 |
| **NHP-AC** | 访问控制器，管理受保护服务器上的防火墙规则 |

| 附加组件 | 职责 |
|-----------|------|
| **NHP-Relay** | HTTP 到 UDP 桥接，使浏览器代理能够通过 HTTPS 发送 NHP 敲门请求 |
| **NHP-KGC** | 基于身份加密（IBC）的密钥生成中心 |

### 协议流程

1. Agent 向 Server 发送加密的敲门请求（`NHP_KNK`）。
2. Server 校验敲门请求，并向 AC 发送操作请求（`NHP_AOP`）。
3. AC 打开防火墙，并回复（`NHP_ART`）给 Server。
4. Server 向 Agent 返回带访问信息的确认（`NHP_ACK`）。
5. Agent 通过 AC 访问受保护资源。

### 加密算法

OpenNHP 提供两种可互换的加密套件：

- **`CIPHER_SCHEME_CURVE`** —— Curve25519 + AES-256-GCM + BLAKE2s
- **`CIPHER_SCHEME_GMSM`** —— SM2 + SM4-GCM + SM3

两者均基于 [Noise 协议框架](https://noiseprotocol.org/)。基于身份的加密（IBC）模式通过密钥生成中心（KGC）提供。

> 更多协议细节、部署模型和加密设计，请参阅[官方文档](https://docs.opennhp.org)。

---

## 仓库结构

```
opennhp/
├── nhp/              # 核心协议库（Go 模块）
│   ├── core/         # 数据包处理、加密、Noise 协议、设备管理
│   ├── common/       # 共享类型与消息定义
│   ├── utils/        # 工具函数
│   ├── plugins/      # 插件处理接口
│   ├── log/          # 日志基础设施
│   └── etcd/         # 分布式配置支持
└── endpoints/        # 守护进程实现（Go 模块，依赖 nhp）
    ├── agent/        # NHP-Agent 守护进程
    ├── server/       # NHP-Server 守护进程
    ├── ac/           # NHP-AC（访问控制器）守护进程
    ├── db/           # NHP-DB（DHP 的数据经纪人）
    ├── kgc/          # NHP-KGC（密钥生成中心）
    └── relay/        # NHP-Relay 守护进程
```

---

## 快速开始

### 先决条件

- Go 1.25.6+
- `make`
- Docker 与 Docker Compose（用于完整的演示环境）

### 构建

```bash
# 构建所有组件
make

# 构建单个守护进程
make agentd    # NHP-Agent
make serverd   # NHP-Server
make acd       # NHP-AC
make db        # NHP-DB
make relayd    # NHP-Relay
make kgc       # NHP-KGC
```

### 测试

```bash
cd nhp && go test ./...
cd endpoints && go test ./...
```

### 使用 Docker 运行

```bash
cd docker && docker-compose up --build
```

请参考[快速入门教程](https://docs.opennhp.org/nhp_quick_start/)，在 Docker 环境中模拟完整的认证工作流。

---

## 贡献

欢迎贡献代码！提交 Pull Request 前请先阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

**注意：** 所有提交必须使用已验证的 GPG 或 SSH 密钥签名。

```bash
git commit -S -m "your message"
```

---

## 安全

发现漏洞了吗？请遵循 [SECURITY.md](SECURITY.md) 中的负责任披露流程，而不是提交公开 issue。

---

## 赞助商

<a href="https://layerv.ai">
  <img src="docs/images/layerv_logo.png" height="40" alt="LayerV.ai logo">
</a>

---

## 许可协议

基于 [Apache 2.0 许可协议](LICENSE)发布。

## 联系方式

- 邮箱：[support@opennhp.org](mailto:support@opennhp.org)
- Discord：[加入我们的 Discord](https://discord.gg/CpyVmspx5x)
- 官网：[https://opennhp.org](https://opennhp.org)
