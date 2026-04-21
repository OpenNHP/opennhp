---
layout: page
title: 协议参考
nav_order: 13
has_children: true
permalink: /zh-cn/protocol/
description: OpenNHP 线上格式参考。将 CSA NHP 规范映射到 Go 实现。
---

# 协议参考
{: .fs-9 }

OpenNHP 所实现的网络基础设施隐身协议（NHP）的线上格式参考文档。
如果你正在编写新的客户端、将 NHP 移植到其他语言，或开发需要解码 NHP
数据包的工具，这里是权威的起点。
{: .fs-6 .fw-300 }

*对应 CSA《Stealth Mode SDP for Zero Trust Network Infrastructure》白皮书——§NHP Message Header、§NHP Message Types 及 Appendix 2。*

## 本节内容

- **[消息头]({{ '/zh-cn/protocol/header/' | relative_url }})** —— 每个 NHP 或 DHP 数据包前缀的 240/304 字节固定头。字段布局、混淆方案、各字段在密码学协议中的作用。
- **[消息类型]({{ '/zh-cn/protocol/messages/' | relative_url }})** —— 全部 17 种 NHP 消息类型加 11 种 DHP 类型，逐一列出发送方/接收方角色、载荷字段与源码入口。

## 规范 ↔ 实现

| CSA 规范章节 | OpenNHP 代码 |
|---|---|
| NHP Message Header (Table 3) | [`nhp/core/scheme/curve/header.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/scheme/curve/header.go)、[`nhp/core/scheme/gmsm/header.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/scheme/gmsm/header.go)、[`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go) |
| NHP Message Types (Table 4, Appendix 2) | [`nhp/common/nhpmsg.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/nhpmsg.go) |
| 密码算法与框架 | [`nhp/core/crypto.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/crypto.go)、[`nhp/core/device.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/device.go) |
| NHP 工作流程（Figure 3） | [`endpoints/agent/`](https://github.com/OpenNHP/opennhp/tree/main/endpoints/agent)、[`endpoints/server/`](https://github.com/OpenNHP/opennhp/tree/main/endpoints/server)、[`endpoints/ac/`](https://github.com/OpenNHP/opennhp/tree/main/endpoints/ac) |
| 与 SDP 集成 | 参见 [部署]({{ '/zh-cn/deploy/' | relative_url }}) |
| 日志 | [`nhp/log/`](https://github.com/OpenNHP/opennhp/tree/main/nhp/log) |

## 规范版本

OpenNHP 遵循 CSA 零信任工作组发布的 NHP 白皮书。该白皮书将 OpenNHP
**v0.6.0** 作为参考实现进行引用。当规范与代码不一致时，本仓库以代码为
准，并将差异反馈给工作组。

## 范围

本节所覆盖的是线上格式：

- 消息头字节布局
- 加密与认证信封（Noise + AEAD）
- 公开头字段的混淆机制
- 消息类型 ID 及其载荷 schema
- Session-ID、计数器（Counter）与重放保护语义

**不在本节范围：**

- 配置文件格式 → 参见 [部署 OpenNHP]({{ '/zh-cn/deploy/' | relative_url }})
- 组件内部实现细节 → 参见 [源代码解读]({{ '/zh-cn/code/' | relative_url }})
- 插件接口 → 参见 [服务器插件开发]({{ '/zh-cn/server_plugin/' | relative_url }}) 与 [客户端 SDK]({{ '/zh-cn/agent_sdk/' | relative_url }})

## 术语

本节所使用的规范术语定义于 [术语表]({{ '/zh-cn/glossary/' | relative_url }})。
白皮书在 SDP 集成语境下偶尔使用同义词（例如用"网关"指代 NHP-AC），为
了与 Go 代码保持一致，本文档始终使用 NHP 官方命名。
