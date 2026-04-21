---
layout: page
title: 术语表
nav_order: 12
permalink: /zh-cn/glossary/
description: OpenNHP 与 CSA NHP 规范通用的规范术语。
---

# 术语表
{: .fs-9 }

OpenNHP 与 CSA NHP 规范通用的规范术语。在 issue、PR、生成的代码
与工具中优先使用这里的写法。白皮书在 SDP 集成语境下偶尔使用别
名；这种情况下本表会把别名列在"亦称"一列，但实现始终使用规范
术语。
{: .fs-6 .fw-300 }

*对应 CSA《Stealth Mode SDP》白皮书 §NHP Core Components 与 §Components That Interact with NHP。*

## 协议角色

| 规范术语 | 亦称 | 含义 |
|---|---|---|
| **NHP-Agent** | agent、客户端、发起方 | 客户端组件，可内嵌于主机、SDK、浏览器或应用中。发送加密敲门请求，授权通过后连接到被保护资源。 |
| **NHP-Server** | server、控制器（SDP 语境） | 控制面组件：解密敲门包、鉴权 NHP-Agent、通过 ASP 校验策略，并指挥 NHP-AC。对应 NIST SP 800-207 中的 *Policy Engine/Administrator*。 |
| **NHP-AC** | 访问控制器、网关（SDP 语境） | 数据面组件：执行默认拒绝，并在 NHP-Server 指令下动态放行防火墙。对应 NIST SP 800-207 中的 *Policy Enforcement Point*。 |
| **NHP-DB** | — | 数据对象隐身协议（DHP）的数据对象后端，保存加密的零信任数据对象。 |
| **NHP-KGC** | KGC | 用于基于标识加密（IBC）的密钥生成中心，负责为 agent / server 签发密钥。 |
| **NHP-Relay** | relay、中继 | 可选的 NHP 报文转发器，跨网段中继；通过 NHP-RLY 消息类型保留源地址。 |

## 交互方

| 术语 | 含义 |
|---|---|
| **资源访问主体（Resource Requestor）** | 承载 NHP-Agent 并请求访问被保护资源的实体（用户、设备、应用或服务器）。 |
| **授权服务提供方（ASP）** | 外部 IAM / 策略系统，NHP-Server 在鉴权时会向其查询（例如企业 IAM；当 NHP 前置在 SDP 部署前时，也可以是 SDP Controller）。 |
| **被保护资源（Protected Resource）** | 被 NHP 隐藏的服务、主机、端口或数据对象。 |
| **日志服务器（Log Server）** | 可选的外部 SIEM 或日志存储，接收 NHP-Server / NHP-AC 的访问与审计日志。 |

## 协议概念

| 术语 | 含义 |
|---|---|
| **敲门（Knock）** | NHP-Agent 发往 NHP-Server 的初始加密包（消息类型 `NHP-KNK`），包含 agent 身份、设备信息与目标资源标识。 |
| **先鉴权后连接（Authenticate-before-connect）** | 在任何到被保护资源的 TCP/UDP 连接尚未建立前就完成身份验证。NHP 的核心贡献。 |
| **默认拒绝（Default-deny）** | NHP-AC 的默认姿态：入站流量一律拒绝，除非 NHP-Server 明确指令放行。 |
| **会话（Session）** | 经过认证的、有时效的访问窗口。由 `Session ID` 在 `NHP-KNK` / `NHP-ACK` / `NHP-AOP` / `NHP-ART` 之间关联。 |
| **放行（Open-door）** | 在 NHP-AC 中动态允许特定「源→目的」流量的动作，由 `NHP-AOP` 触发。 |
| **临时密钥（Ephemeral key）** | 每个消息新鲜生成的一次性 ECC 密钥对，用于驱动 Noise 握手，提供每会话的前向保密性。 |
| **静态密钥（Static key）** | NHP-Agent / NHP-Server / NHP-AC 的长期身份密钥，通过带外方式或 NHP-REG / NHP-RAK 流程注册。 |

## 密码学

| 术语 | 含义 |
|---|---|
| **ECC** | 椭圆曲线加密算法。NHP 国际套件默认使用 Curve25519（256 位）；扩展 / 国密套件使用 SM2。 |
| **Noise Protocol Framework** | NHP 用于双向认证与每会话密钥派生的握手与对称加密框架。OpenNHP 的握手结构遵循 Noise `K` 模式（`e`、`es`、`ss`），并在静态密钥密文之前额外插入一段经 AEAD 包裹的 IBC 身份块（`MaximumIdentitySize`；PKI 模式下该块全零）。 |
| **IBC** | 基于标识的加密算法。用户的 ID 字符串即为公钥，简化了密钥分发，但引入 NHP-KGC 的密钥托管问题。 |
| **CL-PKC** | 无证书公钥密码。一种 IBC 变体，通过将密钥生成职责拆分给 KGC 与用户双方来缓解密钥托管问题。 |
| **PKI** | 传统的 X.509 证书与 CA 体系。与 IBC 并行支持。 |
| **密码套件（Cipher scheme）** | 一组密码原语的命名组合。`CIPHER_SCHEME_CURVE` = Curve25519 + AES-256-GCM + BLAKE2s（国际 / 标准）；`CIPHER_SCHEME_GMSM` = SM2 + SM4-GCM + SM3（国密）。 |

## 架构框架

| 术语 | 含义 |
|---|---|
| **零信任架构（ZTA）** | NHP 所践行的安全模型 ——"永不信任，持续验证" —— 由 NIST SP 800-207 正式化。 |
| **软件定义边界（SDP）** | CSA 框架；NHP 最初旨在对其进行增强。在 SDP 术语中，NHP-AC 对应 *Gateway*，NHP-Server 对应 *Controller*。 |
| **单包授权（SPA）** | NHP 所取代的上一代网络隐身技术。SPA 采用单向加密敲门；NHP 增加了双向认证、显式反馈与可扩展性。 |
| **默认开放 vs 默认拒绝** | TCP/IP 默认 *开放*（任何对端都可发起连接）。NHP 在鉴权成功前保持 *拒绝*。 |

## 消息类型命名前缀

所有 NHP 协议消息名均以 `NHP-` 开头，后跟三个字母的助记符。完整列表见 [消息类型参考]({{ '/zh-cn/protocol/messages/' | relative_url }})。
