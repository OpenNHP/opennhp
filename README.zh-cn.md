[![en](https://img.shields.io/badge/lang-en-red.svg)](https://github.com/OpenNHP/opennhp/blob/master/README.md)

# 零信任网络隐身协议

## 当今挑战：AI 将互联网变成一个“黑暗森林”
人工智能技术，尤其是大型语言模型（LLM）的快速发展正在深刻改变网络安全领域的格局。自主漏洞利用（AVE）的出现标志着人工智能时代的重大升级，它简化了漏洞利用的过程，正如本研究论文中所讨论的那样。这一发展使任何暴露的网络服务都面临着更高的风险，与互联网的 “黑暗森林假说 ”产生了共鸣。人工智能工具不断扫描数字环境，快速识别并利用弱点。因此，互联网演变成了 “黑暗森林”，可见性等同于脆弱性。

![OpenNHP Logo](docs/Vul_Risks.png)

这种模式的转变要求我们重新评估传统的网络安全方法，强调主动防御、快速反应机制，并在可能的情况下采用网络隐藏技术来保护关键基础设施。

## 解决方案：零信任网络隐身协议
NHP，或者 **网络隐身协议** 是一种零信任通信协议，在 OSI网络模型的会话层运行，是网络可见性和连接管理的最佳场所。它的主要目标是使受保护的资源不被未经授权的实体发现，同时通过持续验证只允许经过验证的授权主体访问。NHP 从云安全联盟（CSA）发布的软件定义边界（SDP）规范中的单包授权（SPA）协议中汲取灵感。除了 SPA 的功能外，NHP 还增强了安全性、可靠性、可扩展性和可扩展性。这里列出了 NHP 和 SPA 的详细比较。

![OpenNHP as the OSI 5th layer](docs/OSI_OpenNHP.png)

**OpenNHP** 是用 *Golang* 开发的 NHP 协议开源实现。它的设计遵循安全第一的原则，在 OSI 网络模型的会话层（第 5 层）协议上实现了真正的零信任架构。由于会话层负责连接建立和对话控制，因此在会话层实施零信任具有显著的优势：
- **降低漏洞风险：** TCP/IP 协议的开放性导致了 “默认信任 ”连接模式，允许任何人与提供服务的服务器端口建立连接。攻击者利用这种开放性来攻击服务器漏洞。NHP 协议通过在服务器端默认执行 “拒绝所有 ”规则，只允许授权主机建立连接，从而实现了 “绝不信任，始终验证 ”的零信任原则。这能有效减少漏洞利用，尤其是零日漏洞利用。 
- **减少网络钓鱼攻击：** DNS 劫持是对互联网安全的严重威胁，被用于网络钓鱼、窃取敏感信息或传播恶意软件等恶意目的。NHP 协议可作为加密 DNS 解析服务来缓解这一问题。当客户端的 NHP-Agent 向控制器组件 NHP-Server 发送带有受保护资源标识符（如域名）的敲击请求时，如果 NHP-Agent 成功通过验证，NHP-Server 将返回受保护资源的 IP 地址和端口号。由于 NHP 通信是经过加密和相互验证的，因此可以有效降低 DNS 劫持的风险。 
- **减轻DDoS攻击：** 如上所述，客户端无法在未经验证的情况下获取受保护资源的 IP 地址和端口号。如果受保护的资源分布在多个地点，NHP 服务器可能会向不同的客户端返回不同的 IP 地址，从而大大增加了 DDoS 攻击的难度和实施成本。 攻击归属： TCP/IP 协议的连接模式是基于 IP 的。有了 NHP，连接模式就变成了基于身份（ID）。在建立连接之前，必须对连接发起者的身份进行验证，从而大大提高了攻击的可识别性和可追溯性。 

## 安全优势 

- 通过隐藏基础设施减少攻击面 
- 防止未经授权的网络侦察 
- 减少漏洞利用 
- 通过加密DNS阻止网络钓鱼
- 防止DDoS攻击 
- 实现细粒度访问控制 
- 提供基于身份的连接跟踪 

## 主要功能 

- 默认执行 “全部拒绝 ”规则
- 减少漏洞利用 
- 通过加密DNS解析防止网络钓鱼攻击 
- 通过隐藏基础设施防止DDoS攻击 
- 通过基于身份的连接实现攻击归因 
- 默认拒绝所有受保护资源的访问控制 
- 网络访问前基于身份和设备的身份验证 
- 加密 DNS 解析
- 防止 DNS 劫持 
- 可缓解 DDoS 攻击的分布式基础设施 
- 具有解耦组件的可扩展架构 与现有身份和访问管理系统集成 
- 支持各种部署模式（客户端到网关、客户端到服务器等） 
- 使用现代算法（ECC、噪声协议、IBC）确保加密安全

## 架构和工作流程 
OpenNHP 架构受 NIST 零信任架构标准的启发。它采用模块化设计，核心组件如下：

![OpenNHP architecture](docs/OpenNHP_Arch.png)

### OpenNHP核心组件：
#### NHP-Agent 
NHP-Agent是一个客户端组件，用于启动通信和请求访问受保护的资源。它可以通过以下方式实现：
- 独立的客户端应用程序 
- 集成到现有应用程序中的SDK 
- 浏览器插件 
- 移动应用程序 

NHP-Agent负责:
- 生成并向NHP服务器发送敲击请求 
- 维护安全通信渠道 
- 处理身份验证流程 

#### NHP服务器 
NHP服务器是中央控制器，功能:
 - 处理和验证来自代理的验证请求 
 - 与授权服务提供商互动，做出政策决定 
 - 管理NHP-AC组件，允许/拒绝访问 
 - 处理密钥管理和加密操作 

 NHP服务器可以部署在分布式或集群配置中，以实现高可用性和可扩展性。 

#### NHP-AC

NHP-AC （访问控制）组件执行受保护资源的访问策略，主要功能 ：
- 执行默认的全部拒绝规则 
- 根据NHP服务器指令打开/关闭访问权限 
- 确保受保护资源的网络隐蔽性 
- 记录访问尝试

### 与OpenNHP交互的组件： 
- **受保护资源：** 资源提供者负责保护这些资源，如API接口、应用服务器、网关、路由器、网络设备等。在SDP 方案中，受保护资源是 SDP 网关和控制器。 
- **授权服务提供商（ASP）：** 该提供商负责验证访问策略，并提供受保护资源的实际访问地址。在 SDP 方案中，ASP 可以是SDP控制器。 

### 工作流程
1.'NHP-Agent' 向 'NHP-Server' 发送敲击请求 
2.'NHP-Server'验证请求并检索代理信息 
3.'NHP-Server'查询授权服务提供商 
4.如果获得授权，'NHP-Server'指示'NHP-AC'允许访问 
5.'NHP-AC' 打开连接并通知 'NHP-Server'
6.'NHP-Server'向'NHP-Agent'提供资源访问详情 
7.'NHP-Agent'现在可以访问受保护的资源 
8. 访问记录作为日记存储，用于审计

## 快速开始 
在几分钟内启动并运行 OpenNHP：
```bash
git clone https://github.com/opennhp/opennhp.git
cd opennhp
make
./nhp-server run
```

## 安装步骤：
1.去github clone到本地:
```bash
git clone https://github.com/opennhp/nhp.git
```
2.进入文件夹
```
cd nhp
```
3.build工程:
```bash
make
```
4.安装(optional):
```bash
sudo make install
```
注意：运行 `sudo make install` 需要root权限。运行此命令前，请确保您信任源代码。

## 部署模式 

OpenNHP 支持多种部署模式，以适应不同的使用情况： 

- 客户端到网关：确保访问网关后面的多个服务器 
- 客户端到服务器：直接保护单个服务器/应用程序的安全 
- 服务器到服务器：确保后端服务之间的通信安全 
- 网关到网关：确保站点到站点连接的安全 

## 加密基础 
OpenNHP 采用最先进的加密算法： 
- 椭圆曲线加密算法（ECC）：用于高效的公钥操作 
- 噪声协议框架：用于安全密钥交换和身份验证 
- 基于身份的加密算法（IBC）：大规模简化密钥管理

## SPA与NHP的比较(todo)
NHP 利用现代加密算法和编程语言确保安全性和高性能，有效解决了 SPA 的局限性。


## 贡献 
我们欢迎为 OpenNHP 投稿！有关如何参与的详细信息，请参阅我们的贡献指南。 

## 许可证 
OpenNHP 根据 Apache 2.0 许可发布。 

## 联系我们 
- 项目网站: https://github.com/OpenNHP/opennhp 
- Slack频道：加入我们的Slack

如需更详细的文档，请访问我们的官方文档。(https://docs.opennhp.org)


## 引用

- Software-Defined Perimeter (SDP) Specification v2.0. Jason Garbis, Juanita Koilpillai, Junaid lslam, Bob Flores, Daniel Bailey, Benfeng Chen, Eitan Bremler, Michael Roza, Ahmed Refaey Hussein. Cloud Security Alliance(CSA). Mar 2022.
- AHAC: Advanced Network-Hiding Access Control Framework. Mudi Xu, Benfeng Chen, Zhizhong Tan, Shan Chen, Lei Wang, Yan Liu, Tai Io San, Sou Wang Fong, Wenyong Wang, and Jing Feng. Applied Sciences Journal. June 2024.
- Vulnerability Management Framework project. https://phoenix.security/web-vuln-management/