---
layout: page
title: 部署OpenNHP
nav_order: 6
permalink: /zh-cn/deploy/
---

# 部署OpenNHP
{: .fs-9 }


---

## 1. OpenNHP组件说明

根据上一章中的构建步骤，构建结果将输出到 *release* 目录下，该目录下的三个子目录分别包含OpenNHP的三大核心组件：*nhp-agent*、*nhp-server*和*nhp-ac*。

- **nhp-agent代理：** 发起敲门请求的模块，敲门请求中带有数据访问者的身份和设备信息，通常安装在用户终端的设备上。
- **nhp-server服务器:** 处理并验证敲门请求的模块，通常是服务器程序。其功能包括验证敲门请求并与外部授权服务提供商进行交互实现鉴权操作，以及控制NHP门禁进行开门动作。
- **nhp-ac门禁：** 访问控制的执行模块，通常是服务器程序。该模块执行默认“拒绝一切”（deny all）的安全策略并确保被保护资源的网络隐身状态，通常位于被保护资源所在的同一主机上。负责向已授权的NHP代理开放访问权限或向已失去授权的NHP代理关闭访问权限，并根据NHP服务器返回参数执行针对NHP代理的放行动作。

## 2. OpenNHP开发测试环境搭建

### 2.1 开发测试环境：Windows/MacOS开发主机 + Linux虚拟机

假设开发主机为Windows或者macOS，可通过安装虚拟机环境（如VirualBox）并创建两台Linux虚拟机来搭建简单的OpenNHP测试环境。在创建虚拟机时，请将网卡选项设置为`"Host-only Adapter"`（如下图），可使虚拟机的IP与开发主机IP在同一个网段。

 ![VirualBox Network](/images/vbnetwork.png)

 **提示：** 如需该虚拟机同时具备访问互联网能力，可以另外增加一个`"NAT"`网卡：
 ![VirualBox Network](/images/vbnetwork2.png)

至此，NHP三大组件的环境搭建如下：

- 【nhp-server】 运行在Linux虚拟主机上，IP地址为*192.168.56.101*
- 【nhp-ac】 运行在Linux虚拟主机上，IP地址为*192.168.56.102*
- 【nhp-agent】 运行在Windows/macOS开发主机上，IP地址为*192.168.56.1*

### 2.2 开发测试环境的网络拓扑与基础信息

 ![OpenNHP-Dev-WSL](/images/dev_wsl.png)

| 服务器名称 | IP地址 | 基础配置信息  |
|:--:|:--:|:--:|
| NHP-Server | 192.168.56.101  | **公钥:** WqJxe+Z4+wLen3VRgZx6YnbjvJFmptz99zkONCt/7gc=<br/>**私钥:** eHdyRHKJy/YZJsResCt5XTAZgtcwvLpSXAiZ8DBc0V4= <br/> **Hostname:** localhost <br/> **ListenPort:** 62206 <br/> **aspId:** example |
| NHP-AC | 192.168.56.102  | **公钥:** Fr5jzZDVpNh5m9AcBDMtHGmbCAczHyPegT8IxQ3XAzE=<br/>**私钥:** +B0RLGbe+nknJBZ0Fjt7kCBWfSTUttbUqkGteLfIp30=<br/>**ACId:** testAC-1 <br/> 被保护资源的 **resId:** test |
| NHP-Agent | 192.168.56.1  | **公钥:** WnJAolo88/q0x2VdLQYdmZNtKjwG2ocBd1Ozj41AKlo=<br/>**私钥:** +Jnee2lP6Kn47qzSaqwSmWxORsBkkCV6YHsRqXCegVo= <br/> **UserId:** agent-0 |

**【注意】** 每个组件都有对应的配置文件，需要配置正确才能成功启动。关于配置文件的格式，请见下文中各个组件的“配置文件”相关信息。

**【提示】** 从0.3.3版本起，各组件配置文件中的大多数字段都支持动态更新，详见各配置文件注释说明。

### 2.3 NHP-Server的配置与运行

#### 2.3.1 NHP-Server系统要求

- Linux服务器或者Windows

#### 2.3.2 NHP-Server运行

将*release*目录下*nhp-server*目录复制到目标机器上。配置好*etc*目录下 `toml`文件(详细参数见下一节)，运行`nhp-serverd run`。

- Linux环境：

   ```bash
   nohup ./nhp-serverd run 2>&1 &
   ```

- Windows环境：

   ```bat
   nhp-serverd.exe run
   ```

*【可选项】* 禁止UDP端口暴露：运行`iptables_default.sh`

#### 2.3.3 NHP-Server服务器的配置文件

- 基础配置：[config.toml](https://github.com/OpenNHP/opennhp/tree/main/server/main/etc/config.toml)  
- 门禁peer列表配置：[ac.toml](https://github.com/OpenNHP/opennhp/tree/main/server/main/etc/ac.toml)  
- 客户端peer列表配置：[agent.toml](https://github.com/OpenNHP/opennhp/tree/main/server/main/etc/agent.toml)  
- http服务配置：[http.toml](https://github.com/OpenNHP/opennhp/tree/main/server/main/etc/http.toml)  
- 服务器插件读取配置：[resource.toml](https://github.com/OpenNHP/opennhp/tree/main/server/main/etc/resource.toml)  
- 源地址关联列表：[srcip.toml](https://github.com/OpenNHP/opennhp/tree/main/server/main/etc/srcip.toml)  
- 服务器插件资源配置：[resource.toml](https://github.com/OpenNHP/opennhp/tree/main/server/main/etc/resource.toml)  

### 2.4 NHP-AC的配置与运行

#### 2.4.1 NHP-AC系统要求

NHP-AC 支持两种数据包过滤后端，由 `config.toml` 中的 `FilterMode` 决定：

- **`FilterMode = 0`（iptables/ipset）**——遗留模式，要求：
  - Linux 内核支持 `ip_set`：`lsmod | grep ip_set`
  - 主机上已安装 `iptables` 和 `ipset` 二进制（Amazon Linux 2023 下执行 `dnf install iptables ipset`）。

- **`FilterMode = 1`（eBPF / XDP + TC egress）**——demo 流水线默认模式，要求：
  - Linux 内核 **>= 5.6**（XDP generic 模式最低要求）。推荐 **>= 5.8**，否则无法识别 `CAP_BPF` capability，需要授予 `CAP_SYS_ADMIN`。
  - 主机已安装 `bpftool`（`dnf install bpftool kernel-tools`）用于排障。`nhp-acd` 本身不调用 `bpftool`，但工具缺失通常说明 kernel-tools 也缺失。
  - 已挂载 `bpf` 文件系统到 `/sys/fs/bpf`（幂等命令：`mount -t bpf bpf /sys/fs/bpf`）。
  - `nhp-acd` 的 systemd unit 已授予 `CAP_NET_ADMIN` + `CAP_BPF`（参见内置 unit 中的 `AmbientCapabilities=` 行）。

> **说明：** `deploy-demo-v2` GitHub Actions 流水线只使用 eBPF 后端（`FilterMode = 1`）。若您在 demo 流水线之外自行部署 AC，两种模式均可。

#### 2.4.2 NHP-AC运行

将 *release* 目录下 *nhp-ac* 目录复制到目标机器上。配置好 *etc* 目录下 `toml` 文件（详细参数见下一章），再启动 `nhp-acd`。后续步骤取决于 `FilterMode`：

- **`FilterMode = 0`：** 先运行 `iptables_default.sh` 添加防火墙规则——此时外部连接将被全部拒绝。再启动守护进程。**【注意】** `nhp-acd` 与 `iptables_default.sh` 均需在 **root** 权限下运行。

- **`FilterMode = 1`：** 确认 `/sys/fs/bpf` 已挂载（参见上文前置条件），直接启动守护进程即可。`nhp-acd` 会将 XDP/TC 程序与 maps pin 到 `/sys/fs/bpf`，并在默认路由出口网卡上以 XDP generic 模式挂载。

- Linux（iptables 模式）：

   ```bash
   su
   ./iptables_default.sh
   nohup ./nhp-acd run 2>&1 &
   ```

- Linux（eBPF 模式）：

   ```bash
   mount -t bpf bpf /sys/fs/bpf      # 若尚未挂载
   nohup ./nhp-acd run 2>&1 &
   ```

如果想清除旧 iptables 模式下残留的规则，可以运行 `deploy/scripts/cleanup-ac-iptables.sh`（幂等），或者：

   ```bash
   iptables -F
   ```

#### 2.4.3 NHP-AC门禁配置文件

- 基础配置：[config.toml](https://github.com/OpenNHP/opennhp/tree/main/ac/main/etc/config.toml)  
- 服务器peer列表：[server.toml](https://github.com/OpenNHP/opennhp/tree/main/ac/main/etc/server.toml)  

### 2.5 NHP-Agent的配置与运行

#### 2.5.1 NHP-Agent系统要求

- 所有平台：Windows、Linux、macOS、Android、iOS

#### 2.5.2 NHP-Agent运行

将*release*目录下*nhp-agent*目录复制到目标机器上。配置好*etc*目录下 `toml`文件(详细参数见下一章)，运行`nhp-agentd run`。

- Linux环境：

   ```bash
   nohup ./nhp-agentd run 2>&1 &
   ```

- Windows环境：

   ```bat
   nhp-agentd.exe run
   ```

#### 2.5.3 NHP-Agent的配置文件

- 基础配置：[config.toml](https://github.com/OpenNHP/opennhp/tree/main/agent/main/etc/config.toml)  
- 敲门目标配置：[resource.toml](https://github.com/OpenNHP/opennhp/tree/main/agent/main/etc/resource.toml)  
- 服务器peer列表：[server.toml](https://github.com/OpenNHP/opennhp/tree/main/agent/main/etc/server.toml)  

### 2.6 测试NHP网络隐身效果

验证NHP网络隐身效果，可以通过nhp-agent主机 *（IP：192.168.56.1）*进行`nmap扫描（以80端口为例）` nhp-ac主机 *（IP：192.168.56.102）*来测试。此外，可以在另外一台虚拟机（模拟黑客扫描攻击），扫描nhp-ac 主机查看效果。

| 测试用例 | 测试命令 | 测试目的  | 预期结果  |
|:--:|:--:|:--:|:--:|
| nhp-agent未运行 |`nmap -sS -p 80 192.168.56.102` | 测试AC对Agent隐身 | 80/tcp filtered  |
| nhp-agent已运行 |`nmap -sS -p 80 192.168.56.102` | 测试AC对Agent开放 | 80/tcp open  |
| nhp-agent已运行 |`nmap -sS -p 80 192.168.56.102` | 测试AC对黑客隐身 | 80/tcp filtered  |

## 3. 日志说明

### 3.1 日志文件位置

日志文件生成于每个组件各自的*logs*目录下，以日期作为文件名，可通过`tail`命令查看。

- 查看nhp-server的日志

   ```bash
   tail -f release/nhp-server/logs/server-2024-03-10.log
   ```

- 查看nhp-ac的日志

   ```bash
   tail -f release/nhp-ac/logs/ac-2024-03-10.log
   ```

- 查看nhp-agent的日志

   ```bash
   tail -f release/nhp-agent/logs/agent-2024-03-10.log
   ```

### 3.2 日志文件格式

日志的格式如下：

   ```text
   时间戳 代码位置 NHP组件名 [日志权重] 日志消息
   ```

日志权重分成以下几个级别：

- Error
- Critical
- Warning
- Info
- Debug

## 4. 附录A：常见问题FAQ

- **Q：** Windows平台上编译错误：`running gcc failed: exec: "gcc": executable file not found in %PATH%` 
  **A：** 原因是没有安装`gcc`编译工具。请按照上文中3.1.3中步骤安装GCC。

- 日志中显示错误：`NHP-AC [Critical] received stale packet from 192.168.56.101:62206, drop packet` 。 
   【原因】接收方对包的接收时间有要求，数据包发送时间不能早于接收方10分钟以上。
   【修复】两台机器的时间同步

- 怎么调整一次认证后开通的时间？  怎么限制只开放指定的端口？
   【方法】在nhp-server/plugins/下对应的插件模块中，找到etc/resource.toml文件，里面能配置资源的端口、时长、id等信息；如果用的是nhp-agent敲门，则默认是example插件。如果是微信扫码敲门，用的是wxweb插件。

- 如何检查配置是否生效？
   【方法】server和ac的日志里有记录。也可以在ac的系统里输入ipset -L命令查看授权的源ip目的端口和时长。

- nhp-agent 敲门成功，但访问不通可能的原因。

   问题原因：***可能原因是 nhp-server 下发到 nhp-ac 进行 ipset，所添加记录中的 resource 目标没和请求中对应的 resource 目标的 IP 对上***，这种情况可能出现在 resource 与 nhp-ac 在同一台服务器上的情况。可以先手动配置 ipset 规则：
   ``` shell
   sudo ipset add defaultset [SourceIP],tcp:80,[ResourceIP]
   ```
   ***SourceIP 来源IP，即 Agent 的公有 IP，可通过 tcpdump 或 ipset list 确认***
   ***ResourceIP 请求资源的 IP*** 该IP 对应 nhp-server 中 ./plugins/example/etc/resource.toml，<span style="color:red">如果与请求对应不上，则会出现敲门成功但无法请求问题。</span>

   抓包调试：在 ```nhp-ac``` 侧 ```tcpdump -i any port 80```(根据具体情况调整)

   解决方案：将可用的 IP 配置在 nhp-server 的 ```./plugins/example/etc/resource.toml``` 配置中的 ```Addr.Ip = ""```
