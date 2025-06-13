[![en](https://img.shields.io/badge/lang-en-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh--cn-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.zh-cn.md)
[![de](https://img.shields.io/badge/lang-de-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.de.md)
[![ja](https://img.shields.io/badge/lang-ja-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.ja.md)
[![fr](https://img.shields.io/badge/lang-fr-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.fr.md)
[![es](https://img.shields.io/badge/lang-es-green.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.es.md)

![OpenNHP Logo](docs/images/logo11.png)
# OpenNHP: 零信任网络隐身协议
OpenNHP是一个轻量级、基于加密算法的零信任网络协议，其工作在OSI网络模型第五层，用于隐藏您的服务器和数据，避免被攻击者发现和访问

![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Version](https://img.shields.io/badge/version-1.0.0-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)

---

## 挑战：AI 将互联网变为“黑暗森林”

**AI** 技术的快速发展，尤其是大语言模型（LLM），正在显著改变网络安全格局。**自主漏洞利用（AVE）** 的兴起是 AI 时代的一个重大飞跃，大大简化了漏洞的利用，这一点在[这篇研究论文](https://arxiv.org/abs/2404.08144)中有详细说明。这一发展显著增加了任何暴露网络服务的风险，与互联网的[黑暗森林假说](https://en.wikipedia.org/wiki/Dark_forest_hypothesis)不谋而合。AI 驱动的工具不断扫描数字环境，迅速识别和利用弱点。因此，互联网正逐渐成为一个**“黑暗森林”**，**可见性意味着脆弱性**。

![Vulnerability Risks](docs/images/Vul_Risks.png)

Gartner 研究预测，[AI 驱动的网络攻击将迅速增加](https://www.gartner.com/en/newsroom/press-releases/2024-08-28-gartner-forecasts-global-information-security-spending-to-grow-15-percent-in-2025)。这一变化要求重新评估传统的网络安全策略，重点放在主动防御、快速响应机制和网络隐藏技术的采用，以保护关键基础设施。

---

## 快速演示：查看 OpenNHP 的工作原理

在深入了解 OpenNHP 的详细信息之前，让我们先来看一个 OpenNHP 如何保护服务器免受未经授权访问的演示。您可以通过访问 https://acdemo.opennhp.org 查看其实际效果。

### 1) 受保护的服务器对未经身份验证的用户“不可见”

默认情况下，任何试图连接受保护服务器的操作都会导致超时错误，因为所有端口都是关闭的，使服务器看起来像是*“离线”*且实际上是“不可见”的。

![OpenNHP Demo](docs/images/OpenNHP_ACDemo0.png)

对服务器进行端口扫描也会返回超时错误。

![OpenNHP Demo](docs/images/OpenNHP_ScanDemo.png)

### 2) 身份验证后，受保护的服务器变得可访问

OpenNHP 支持多种身份验证方法，如 OAuth、SAML、二维码等。为了演示方便，本次演示使用 https://demologin.opennhp.org 上的基本用户名/密码身份验证服务来展示该过程。

![OpenNHP Demo](docs/images/OpenNHP_DemoLogin.png)

点击“登录”按钮后，身份验证成功完成，您会被重定向到受保护的服务器。此时，服务器在您的设备上变得*“可见”*并且可以访问。

![OpenNHP Demo](docs/images/OpenNHP_ACDemo1.png)

---

## 愿景：让互联网变得值得信赖

TCP/IP 协议的开放性推动了互联网应用的爆炸式增长，但也暴露了漏洞，使得恶意攻击者可以获得未经授权的访问并利用任何暴露的 IP 地址。尽管 [OSI 网络模型](https://en.wikipedia.org/wiki/OSI_model) 在*第五层（会话层）*定义了连接管理，但在实际中很少有有效的解决方案能够应对这一挑战。

**NHP**，即**“网络基础设施隐藏协议”**，是一种轻量级、基于加密的零信任网络协议，旨在工作于*OSI 会话层*，该层在管理网络可见性和连接方面是最佳选择。NHP 的主要目标是将受保护的资源隐藏于未授权的实体，只允许经过验证的用户通过持续认证访问，从而为更值得信赖的互联网作出贡献。

![Trustworthy Internet](docs/images/TrustworthyCyberspace.png)

---

## 解决方案：OpenNHP 解决网络可见性控制问题

**OpenNHP** 是 NHP 协议的开源实现。它基于加密技术，采用安全优先的原则，在*OSI 会话层*实现了真正的零信任架构。

![OpenNHP as the OSI 5th layer](docs/images/OSI_OpenNHP.png)

OpenNHP 构建在早期的网络隐藏技术研究基础之上，利用现代加密框架和架构确保安全性和高性能，从而克服了前代技术的局限性。

| 网络隐藏协议 | 第一代 | 第二代 | 第三代 |
|:---|:---|:---|:---|
| **核心技术** | [端口敲门](https://en.wikipedia.org/wiki/Port_knocking) | [单包认证（SPA）](https://cloudsecurityalliance.org/artifacts/software-defined-perimeter-zero-trust-specification-v2) | 网络基础设施隐藏协议（NHP） |
| **身份认证** | 端口序列 | 共享密钥 | 现代加密框架 |
| **架构** | 无控制平面 | 无控制平面 | 可扩展控制平面 |
| **功能** | 隐藏端口 | 隐藏端口 | 隐藏端口、IP 和域名 |
| **访问控制** | IP 层级 | 端口层级 | 应用层级 |
| **开源项目** | [knock](https://github.com/jvinet/knock) *(C)* | [fwknop](https://github.com/mrash/fwknop) *(C++)* | [OpenNHP](https://github.com/OpenNHP/opennhp) *(Go)* |

> 开发 OpenNHP 选择使用**内存安全**的语言如 *Go*，这一点在[美国政府技术报告](https://www.whitehouse.gov/wp-content/uploads/2024/02/Final-ONCD-Technical-Report.pdf)中得到了强调。有关 **SPA 和 NHP** 之间详细的比较，请参见[下文](#comparison-between-spa-and-nhp)。

## 安全性优势

由于 OpenNHP 在 *OSI 会话层*实现了零信任原则，因此具有显著的优势：

- 通过隐藏基础设施减少攻击面
- 防止未经授权的网络侦察
- 减少漏洞利用的可能性
- 通过加密的 DNS 保护防止钓鱼
- 抵御 DDoS 攻击
- 提供细粒度的访问控制
- 实现基于身份的连接追踪
- 支持攻击溯源

## 架构

OpenNHP 的架构受 [NIST 零信任架构标准](https://www.nist.gov/publications/zero-trust-architecture) 启发，采用模块化设计，包含三个核心组件：**NHP-Server**、**NHP-AC** 和 **NHP-Agent**，如下图所示。

![OpenNHP architecture](docs/images/OpenNHP_Arch.png)

> 有关架构和工作流程的详细信息，请参阅 [OpenNHP 文档](https://opennhp.org/)。

## 核心：加密算法

加密是 OpenNHP 的核心，提供强大的安全性、出色的性能和可扩展性，使用了先进的加密算法。以下是 OpenNHP 采用的关键加密算法和框架：

- **[椭圆曲线密码学（ECC）](https://en.wikipedia.org/wiki/Elliptic-curve_cryptography)**：用于高效的公钥密码学。

> 与 RSA 相比，ECC 具有更高的效率，以较短的密钥长度提供更强的加密能力，从而提高网络传输和计算性能。下表显示了 RSA 和 ECC 在安全强度、密钥长度和密钥长度比率上的差异，以及其有效期。

| 安全强度（位） | DSA/RSA 密钥长度（位） | ECC 密钥长度（位） | 比率：ECC 与 DSA/RSA | 有效期 |
|:---------------:|:----------------------:|:-----------------:|:------------------:|:------:|
| 80              | 1024                   | 160-223           | 1:6                | 到 2010 年 |
| 112             | 2048                   | 224-255           | 1:9                | 到 2030 年 |
| 128             | 3072                   | 256-383           | 1:12               | 2031 年后 |
| 192             | 7680                   | 384-511           | 1:20               | |
| 256             | 15360                  | 512+              | 1:30               | |

- **[Noise 协议框架](https://noiseprotocol.org/)**：用于安全的密钥交换、消息加密/解密和相互身份认证。

> Noise 协议基于[Diffie-Hellman 密钥交换](https://en.wikipedia.org/wiki/Diffie%E2%80%93Hellman_key_exchange)，提供了现代加密解决方案，如相互和可选认证、身份隐藏、前向安全性和零轮次加密。它已被 WhatsApp、Slack 和 WireGuard 等应用广泛验证并使用，证明其安全性和性能。

- **[基于身份的加密（IBC）](https://en.wikipedia.org/wiki/Identity-based_cryptography)**：简化了大规模的密钥分发。

> 高效的密钥分发是实现零信任的关键。OpenNHP 支持 PKI 和 IBC。虽然 PKI 已经被广泛使用，但它依赖于集中式的证书颁发机构（CA）进行身份验证和密钥管理，这在时间和成本上较为昂贵。相比之下，IBC 允许在身份验证和密钥管理方面采用去中心化和自我管理的方法，使其在 OpenNHP 的零信任环境中更具成本效益，尤其是在需要实时保护和管理数十亿设备或服务器的情况下。

- **[无证书公钥加密（CL-PKC）](https://en.wikipedia.org/wiki/Certificateless_cryptography)**：推荐的 IBC 算法。

> CL-PKC 是一种通过避免密钥托管和解决基于身份的加密（IBC）局限性来增强安全性的方案。在大多数 IBC 系统中，用户的私钥由密钥生成中心（KGC）生成，这带来了显著的风险。如果 KGC 被攻破，所有用户的私钥都可能被泄露，这要求对 KGC 完全信任。CL-PKC 通过将密钥生成过程分离，使 KGC 仅了解部分私钥，从而避免这一问题。结果，CL-PKC 结合了 PKI 和 IBC 的优点，在不牺牲安全性的情况下提供更强的保护。

更多阅读：

> 有关 OpenNHP 中使用的加密算法的详细说明，请参阅 [OpenNHP 文档](https://opennhp.org/cryptography/)。

## 主要特性

- 通过强制默认“全部拒绝”规则减少漏洞利用
- 通过加密的 DNS 解决防止钓鱼攻击
- 通过隐藏基础设施保护免受 DDoS 攻击
- 通过身份追踪连接实现攻击溯源
- 对所有受保护资源的默认拒绝访问控制
- 在网络访问前进行基于身份和设备的身份认证
- 加密的 DNS 解决防止 DNS 劫持
- 分布式基础设施抵御 DDoS 攻击
- 解耦组件实现可扩展架构
- 与现有身份和访问管理系统集成
- 支持多种部署模型（客户端到网关、客户端到服务器等）
- 使用现代算法（ECC、Noise 协议、IBC）进行加密确保安全性

<details>
<summary>点击展开特性详情</summary>

- **默认拒绝访问控制**：所有资源默认隐藏，只有通过身份验证和授权后才会变得可访问。
- **基于身份和设备的身份验证**：确保只有已知用户在授权设备上可以访问。
- **加密的 DNS 解决**：防止 DNS 劫持和相关的钓鱼攻击。
- **DDoS 缓解**：分布式基础设施设计有助于抵御分布式拒绝服务攻击。
- **可扩展架构**：解耦组件允许灵活部署和扩展。
- **IAM 集成**：可以与现有身份和访问管理系统配合使用。
- **灵活部署**：支持包括客户端到网关、客户端到服务器等多种模型。
- **强大加密**：使用现代算法如 ECC、Noise 协议和 IBC 确保安全性。
</details>

## 部署

OpenNHP 支持多种部署模型，以适应不同的使用场景：

- 客户端到网关：保护网关后面的多个服务器的访问
- 客户端到服务器：直接保护单个服务器/应用
- 服务器到服务器：保护后端服务之间的通信
- 网关到网关：保护站点到站点的连接

> 有关详细部署说明，请参阅 [OpenNHP 文档](https://opennhp.org/deploy/)。

## SPA 和 NHP 的比较
单包认证（SPA）协议被包含在 [软件定义边界（SDP）规范](https://cloudsecurityalliance.org/artifacts/software-defined-perimeter-zero-trust-specification-v2) 中，由 [云安全联盟（CSA）](https://cloudsecurityalliance.org/) 发布。NHP 通过现代加密框架和架构在安全性、可靠性、可扩展性和可扩展性方面进行了改进，这一点在 [AHAC 研究论文](https://www.mdpi.com/2076-3417/14/13/5593) 中得到了验证。

| - | SPA | NHP | NHP 优势 |
|:---|:---|:---|:---|
| **架构** | SPA 服务器中的 SPA 数据包解密和用户/设备身份验证组件与网络访问控制组件是耦合的。 | NHP-Server（数据包解密和用户/设备身份验证组件）和 NHP-AC（访问控制组件）是解耦的。NHP-Server 可以部署在独立的主机上，并支持水平扩展。 | <ul><li>性能：资源消耗大的组件 NHP-Server 从受保护服务器分离。</li><li>可扩展性：NHP-Server 可以以分布式或集群模式部署。</li><li>安全性：受保护服务器的 IP 地址在身份验证成功之前对客户端是不可见的。</li></ul> |
| **通信** | 单向 | 双向 | 更好的可靠性，访问控制状态通知 |
| **加密框架** | 共享密钥 | PKI 或 IBC，Noise 框架 | <ul><li>安全性：经过验证的安全密钥交换机制，减轻中间人攻击威胁</li><li>低成本：适合零信任模型的高效密钥分发</li><li>性能：高性能加密/解密</li></ul> |
| **隐藏网络基础设施的能力** | 仅服务器端口 | 域名、IP 和端口 | 更强大，针对各种攻击（如漏洞利用、DNS 劫持和 DDoS 攻击） |
| **可扩展性** | 无，仅适用于 SDP | 通用 | 支持任何需要服务暗化的场景 |
| **互操作性** | 不支持 | 可定制 | NHP 可以无缝集成现有协议（如 DNS、FIDO 等） |

## 贡献

我们欢迎对 OpenNHP 的贡献！有关如何参与的更多信息，请参阅我们的[贡献指南](CONTRIBUTING.md)。

## 许可协议

OpenNHP 遵循 [Apache 2.0 许可协议](LICENSE)。

## 联系方式

- 项目网站：[https://github.com/OpenNHP/opennhp](https://github.com/OpenNHP/opennhp)
- 电子邮件：[opennhp@gmail.com](mailto:opennhp@gmail.com)
- Slack 频道：[加入我们的 Slack](https://slack.opennhp.org)

有关更详细的文档，请访问我们的[官方网站](https://opennhp.org)。

## 参考文献

- [软件定义边界（SDP）规范 v2.0](https://cloudsecurityalliance.org/artifacts/software-defined-perimeter-zero-trust-specification-v2)。Jason Garbis、Juanita Koilpillai、Junaid lslam、Bob Flores、Daniel Bailey、Benfeng Chen、Eitan Bremler、Michael Roza、Ahmed Refaey Hussein。[*云安全联盟（CSA）*](https://cloudsecurityalliance.org/)。2022 年 3 月。
- [AHAC：高级网络隐藏访问控制框架](https://www.mdpi.com/2076-3417/14/13/5593)。Mudi Xu、Benfeng Chen、Zhizhong Tan、Shan Chen、Lei Wang、Yan Liu、Tai Io San、Sou Wang Fong、Wenyong Wang 和 Jing Feng。*应用科学杂志*。2024 年 6 月。
- [STALE：利用电子邮件和ECDH密钥交换的可扩展、安全的跨境认证方案](https://www.mdpi.com/2079-9292/14/12/2399).Jiexin Zheng, Mudi Xu, Jianqing Li, Benfeng Chen, Zhizhong Tan, Anyu Wang, Shuo Zhang, Yan Liu, Kevin Qi Zhang, Lirong Zheng, 和 Wenyong Wang.*电子学报*。2025年6月。
- Noise 协议框架。https://noiseprotocol.org/
- 漏洞管理框架项目。https://phoenix.security/web-vuln-management/

---

🌟 感谢您对 OpenNHP 的关注！我们期待您的贡献和反馈。

