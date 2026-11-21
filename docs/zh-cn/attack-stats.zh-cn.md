---
layout: page
title: 攻击统计
nav_order: 16
permalink: /zh-cn/attack-stats/
description: nhp-server 的 XDP 入站过滤器丢弃的包如何变成 stats.opennhp.org 上每天发布的汇总，以及每个数字的确切含义。
---

# XDP 过滤器与攻击统计站是如何协同的
{: .fs-9 }

OpenNHP 的 demo 服务器**故意**把几乎所有端口暴露在公网上，好让真实的扫描与攻击能够到达网卡，从而被测量。挂在网卡上的内核级 XDP 程序丢掉一切不是合法 NHP 敲门的包，这些丢弃被写进每天的日志。第二台主机上的采集器把日志拉回去，离线分析，再发布成一个静态站点。本文讲的是这两半之间的接缝：过滤器向站点承诺了什么，站点拿它做了什么，以及每个发布的数字究竟意味着什么。
{: .fs-6 .fw-300 }


---

文中每一处都能对应到仓库里的具体文件，每一章末尾都会给出源码路径。代码是权威：本文引用某个文件时，引用是逐字的；两者若有出入，以源码为准。

站点在 <https://stats.opennhp.org/>。本文**不引用**站点上的任何具体数字：那些数字每天都在变，而机器可读的 `index.json` 是它们唯一权威的版本。

## 目录

1. [为什么 demo 要把端口全开](#1-为什么-demo-要把端口全开)
2. [过滤器与站点之间的契约](#2-过滤器与站点之间的契约)
3. [过滤器写出的日志](#3-过滤器写出的日志)
4. [从一个被丢弃的包到一行已发布的记录](#4-从一个被丢弃的包到一行已发布的记录)
5. [已发布的数字意味着什么](#5-已发布的数字意味着什么)
6. [发布产物及其契约](#6-发布产物及其契约)
7. [发布了什么，没发布什么](#7-发布了什么没发布什么)
8. [运行与运维](#8-运行与运维)
9. [接下来读什么](#9-接下来读什么)

## 1. 为什么 demo 要把端口全开

安全组是让一台主机保持安静的最低成本手段，同时也是让一台主机**对观测不可见**的最低成本手段。被安全组丢掉的包根本到不了网卡，于是挂在那个网卡上的过滤器永远看不到它、永远不会统计它、也永远不会为它记下一个理由。在窄安全组姿态下，互联网上绝大部分的背景噪声在过滤器被调用之前就已经消失了，统计站也就没有什么可展示的。

因此 demo 有两种姿态，都是显式选择才生效的：

| Terraform 变量 | 默认值 | 它放开什么 |
| --- | --- | --- |
| `server_open_all_ports` | `false` | 除 22 以外的每个 tcp 端口、每个 udp 端口，以及 icmp，对 `0.0.0.0/0` |
| `ac_open_all_ports` | `false` | 同上，作用于 nhp-ac 的安全组 |

在宽姿态下，挂在入站网卡上的 XDP 过滤器是**除 tcp/22 以外一切入站流量的唯一控制点**。这是有意为之：tcp/22 在网络层始终只对 relay 的安全组开放，所以 SSH 永远不依赖过滤器——那个链接没有被 pin 住，而且按设计是 fail-open 的。

由此得到两个结论，二者对解读已发布的数字都很重要：

- **过滤器的盲区就是站点的盲区。** 过滤器判不了的东西——没挂上的过滤器、程序里没有类别的理由——要么根本不在日志里，要么被记成一个未知原因。站点发布的是日志里有的东西；它看不到日志之外。
- **站点上的数字描述的是一台 demo 主机。** 它们是「过滤器在真实互联网噪声中的行为」的样本，不是吞吐测量、不是基准测试，也不是关于 OpenNHP 性能的任何主张。

宽姿态也不是没有代价。过滤器会在每次部署、每次重启、每次崩溃、每次开机、以及每次加载失败时被摘下来；在它被摘下来期间，主机在所有被放开的端口上都是敞开的。已发布的数字无法区分「过滤器什么都没看到」与「过滤器根本不在」，这正是**已发布数据里缺失的一个小时永远不能被读成平静的一个小时**的原因。

### 源码位置

- `terraform/demo/variables.tf` —— `server_open_all_ports`、`ac_open_all_ports`。
- `terraform/demo/RUNBOOK.md` —— *All-ports exposure mode* 与 *All-ports exposure mode on the AC* 两节；fail-open 的各个窗口在那里逐条列出。
- `nhp/ebpf/xdp/nhp_server_xdp.c` —— 挂在 nhp-server 入站网卡上的程序。
- `nhp/utils/ebpf/engine_linux.go` —— `dropLoaderPrivileges`，fail-open 的加载路径与挂载路径。

## 2. 过滤器与站点之间的契约

XDP 程序不解析任何 NHP 协议。它不知道敲门是什么。它只看包的形状——地址族、协议、端口、帧长，以及源地址是否在 SSH 白名单里——然后给出三样东西：

1. 一个**判决**，`PASS` 或 `DROP`；
2. 一个**原因码**，一个稳定的大写名字；
3. 一个**计数**，按类别、按窗口。

身份是在别处决定的，在用户态，由 Noise 握手决定。过滤器的职责更窄：把不可能属于本协议的包挡在「根本不存在」这一步。

原因词表就是契约本身。nhp-server 侧的命名权威是 `nhp/utils/ebpf/engine_linux.go` 里的 `serverActionName`，那些名字只在这一处被定义。站点通过 `endpoints/xdplog-report/classify.go` 里的 `dropCategories` 消费它们，而一个交叉校验测试（`endpoints/xdplog-report/report_test.go` 里的 `TestClassifierMatchesEngine`）会读取引擎源码，一旦站点不认识过滤器可能发出的某个原因就失败。两张表不可能悄悄漂移。

nhp-server 的 PASS 原因：

| 原因 | 含义 |
| --- | --- |
| `SSH_RELAY` | 来自白名单 relay 地址的 SSH 连接 |
| `NHP_RELAY` | 来自白名单 relay 地址、发往敲门端口的 UDP，豁免长度下限 |
| `NHP_DEFAULT` | 发往敲门端口、且达到 NHP 最小帧长的 UDP——也就是「敲门」这一类本身 |
| `TCP_ESTABLISHED` | 主机自己发起的连接的回应 |
| `UDP_ESTABLISHED` | 主机自己发起的连接的回应 |
| `ICMP_FRAG_NEEDED` | 任何 ICMP fragmentation-needed 报文（无条件放行；内核再把它匹配到一个 socket） |
| `DHCP_CLIENT` | DHCP 客户端自身的交互（67 → 68） |
| `NTP_CLIENT` | NTP 客户端自身的交互（123 → 123） |
| `IPV4_FRAGMENT` | 任何非首片 IPv4 分片（没有 L4 头可供判断；无条件放行） |
| `ICMPV6_CONTROL` | IPv6 控制报文 |
| `V6_ESTABLISHED` | 主机自己发起的 v6 连接的回应 |
| `IPV6_FRAGMENT` | 任何非首片 IPv6 分片（没有 L4 头可供判断；无条件放行） |

nhp-server 的 DROP 原因，以及站点给每个原因归入的类别：

| 原因 | 类别 | 含义 |
| --- | --- | --- |
| `TCP_SSH_OTHER` | SSH probe (non-whitelisted source) | 来自白名单之外源的 SSH 连接 |
| `TCP_NHP_PORT` | TCP to NHP knock port | 指向 UDP 敲门端口的 TCP |
| `TCP_OTHER` | TCP scan / probe | 其它任何入站 TCP |
| `TCP_SYN_SCAN` | TCP SYN scan | 裸 SYN，后面没有已建立的流 |
| `TCP_STEALTH_SCAN` | TCP scan / probe | 不带 SYN、ACK 或 RST 的 TCP 报文——FIN、NULL 或 XMAS 扫描 |
| `UDP_OTHER` | UDP scan / non-NHP UDP | 既不是敲门、也不是主机自身查询回应的 UDP |
| `UDP_SHORT` | Short UDP (below NHP minimum frame) | 短于 NHP 最小帧长度的数据报 |
| `NON_TCP_UDP` | Other protocol (ICMP etc.) | 既非 TCP 也非 UDP 的协议 |
| `V6_OTHER` | IPv6 | 没有任何规则放行的 IPv6 流量 |
| `OTHER` | Other | 程序丢弃的其余一切 |
| `UNKNOWN-<n>` | Unknown reason | 命名表不认识的 action code |

nhp-ac 的数据发布到同一个站点，它的 DROP 原因**刻意**复用服务器的词表——站点不认识的类别没法和其余类别一起统计。它只多出一个自己的：

| 原因 | 类别 | 含义 |
| --- | --- | --- |
| `KNOCK_EXPIRED` | Knock window expired | 敲过门、但在再次说话之前敲门窗口就关闭的对端 |

最后这一条是**一个来晚了的合法对端**——一次重传、一个活得比窗口更久的会话、一个时钟走快了的 agent。它不是扫描，站点不把它算作攻击。包照样被丢掉；这条排除说的是**记账**，不是过滤器放了什么。

站点对**未知**原因的处理方向相反：未知原因**要**算作攻击。分类器是一张封闭清单，所以这个版本没听说过的原因，更可能是一个新的攻击类别而不是一个新的非攻击类别，而**少报攻击是更糟的错误**。

### 源码位置

- `nhp/utils/ebpf/engine_linux.go` —— `serverActionName`、`acActionName`；两张命名表。
- `endpoints/xdplog-report/classify.go` —— `dropCategories`、`classify`、`nonAttackCategories`、`isAttackReason`。
- `endpoints/xdplog-report/report_test.go` —— `TestClassifierMatchesEngine`，让两张词表保持同步的交叉校验。
- `nhp/ebpf/xdp/nhp_server_xdp.c` —— 服务器判决树背后的程序。

## 3. 过滤器写出的日志

过滤器的判决由守护进程转成文本，而这段文本是站点唯一的输入。一共存在三种行型，外加一个标记行。下面是仓库测试样本 `endpoints/xdplog-report/testdata/nhp_server_xdp-2026-10-01.log` 里的真实行：

```
10:00:00 srv1 [NHP-STAT] VERDICT=DROP REASON=UDP_SHORT WINDOW=60s PKTS=1000 LOGGED=0 TOTAL=1000
10:03:00 srv1 [NHP-DROP] REASON=TCP_OTHER SRC=203.0.113.5 DST=10.0.0.1 LEN=60 PROTO=TCP SPT=1000 DPT=22 RELAY=0
10:04:00 srv1 [NHP-SCAN] WINDOW=300s KIND=TOTAL UNIQ_SRC=2 UNIQ_PORT=2 PKTS=2 LRU=0
```

以及那个预算标记行：

```
10:02:00 srv1 [NHP-STAT] EVENT_LOG_BUDGET_EXHAUSTED
```

每一种的含义：

| 行 | 频率 | 精确性 | 携带什么 |
| --- | --- | --- | --- |
| `[NHP-STAT]` | 每个活跃类别每个窗口一行（`WINDOW=60s`） | **精确** | `PKTS`（本窗口内的包数）、`LOGGED`（其中有多少条同时写了逐包行）、`TOTAL`（该类别自守护进程启动以来的累计） |
| `[NHP-DROP]` | 每个被采样的包 | **一个采样** | `REASON`、`SRC`、`DST`、`LEN`、`PROTO`、`SPT`、`DPT`、`RELAY` |
| `[NHP-SCAN]` | 周期性汇总（`WINDOW=300s`） | 逐窗口精确，除非 `LRU=1` | `UNIQ_SRC`、`UNIQ_PORT`、`PKTS`——该窗口被丢弃的包，按源地址和目的端口分别计数——以及 `LRU`，它表示计数表至少已满 95%，可能已经发生淘汰 |
| `EVENT_LOG_BUDGET_EXHAUSTED` | 每天一次 | —— | 当天的逐包字节预算用尽；`[NHP-STAT]` 的总数继续 |

前两行之间的区别，是理解已发布数字时最重要的一件事：

- **`[NHP-STAT]` 是记账用的计数。** 它按类别按窗口发出，不受限速，站点精确的报文总数就来自它。
- **`[NHP-DROP]` 是一个采样。** 它在内核的令牌桶下写出，所以它展示的是**丢了什么**、**来自哪里**，但永远不是**丢了多少**。任何由它导出的排名——top 来源、top 目的端口——都是对采样的排名。

第三行又是第三条通道，而站点不从中发布任何数字。内核把**每一个**被丢弃的包都计入两张有界的表——一张按 (协议, 目的端口)，一张按源地址——无论上报预算是多少，所以那一行上的 `PKTS` 是该窗口丢弃包的精确计数。这两张表是 LRU 的，因此比表更大的扫描会把最冷的键淘汰掉，排名随之变成近似值；`LRU=1` 就是在宣布这件事。站点自己的 top 来源与 top 目的端口来自被采样的逐包行。

`TOTAL` 字段是自守护进程启动以来的累计值。这正是重启可见的原因：两个窗口之间 `TOTAL` 变小，说明守护进程重启过，站点把它报告为 `restarts_detected`。

有三道界限防止日志变成攻击它所保护的那台主机的拒绝服务工具，而且没有一道会改变判决：

- **内核里按类别的令牌桶。** 每个 action 类别每 CPU 每秒 16 的突发额度，另有一个跨所有类别的全局桶，每 CPU 每秒 8。按类别的桶防止某一类洪水饿死其它类别的记录；全局桶则无论类别怎样混合都把总量限制住。
- **写端每天的字节预算。** 每个日历日 64 MiB 的逐包文本。预算用尽后，写端停止追加逐包行并明确说出来——而 `[NHP-STAT]` 汇总继续报告精确总数。
- **保留期清扫。** 14 天或 256 MiB，先到者为准；先删最老的文件，当天的文件永不删除。

日志文件本身由 `nhp/log/logger.go` 以 0600 权限创建，里面是**完整未掩码**的源地址。在 demo 主机上，那是这些地址唯一存在的地方。

### 源码位置

- `endpoints/xdplog-report/testdata/nhp_server_xdp-2026-10-01.log` —— 上面的样本逐字节取自这个样本文件。
- `nhp/utils/ebpf/engine_linux.go` —— `serverLogDailyEventBytes`、`serverLogRetentionDays`、`serverLogRetentionBytes`、`serverLogSweepInterval`，以及 `[NHP-STAT]` 的写端。
- `nhp/ebpf/xdp/nhp_xdp_stats.h` —— `NHP_EVENT_BURST`、`NHP_EVENT_GLOBAL_BURST`：内核侧的限速。
- `nhp/log/logger.go` —— 以 0600 权限创建文件的写端。
- `endpoints/xdplog-report/analyze.go` —— 把 `TOTAL` 变小转成 `restarts_detected` 的重启检测。

## 4. 从一个被丢弃的包到一行已发布的记录

整条流水线有五跳，每一跳都刻意做得很窄。

```
   ┌──────────────────────────────────────────────────────────────────┐
   │  互联网                                                            │
   └───────────────────────────┬──────────────────────────────────────┘
                               │  帧到达网卡
                               ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │  nhp-server 主机                                                  │
   │                                                                  │
   │  XDP 程序 ──DROP──► perf ring ──► Go 读取端 ──► 日志行           │
   │  （内核）                       （engine_linux.go）               │
   │                                                                  │
   │  logs/nhp_server_xdp-YYYY-MM-DD.log   0600，14 天                │
   └───────────────────────────┬──────────────────────────────────────┘
                               │  SSH，forced command：list / get
                               ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │  relay 主机                                                       │
   │                                                                  │
   │  xdpstats-collect.sh                                              │
   │    raw/<source>/nhp_<source>_xdp-<date>.log.gz  （+ .meta）      │
   │    0700，90 天，flock 保护                                        │
   │                                                                  │
   │  xdplog-report site ──► /var/www/stats                           │
   └───────────────────────────┬──────────────────────────────────────┘
                               │  nginx，只读
                               ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │  https://stats.opennhp.org/                                      │
   │    index.html  index.json  d/<source>/<day>.{html,json}          │
   │    totalnumber.js  countup.js                                    │
   └──────────────────────────────────────────────────────────────────┘
```

**第 1 跳——过滤器写日志。** 见第 3 章。

**第 2 跳——一条只读的 SSH 通道。** 采集器拿不到 shell。一把专用的 ed25519 公钥被装进服务器的 `authorized_keys`，背后挂着一个 forced command，而这个命令只接受两个动词：

- `list` —— 每个日志文件一行 TSV：日期、大小、mtime。
- `get <YYYY-MM-DD>` —— 把那一个日志流到标准输出。

其余一切都在脚本处被拒绝：shell 元字符、路径分隔符、不匹配 `^[0-9]{4}-[0-9]{2}-[0-9]{2}$` 的日期、守护进程从未写过的日子，以及不存在的文件。`get` 路径拒绝跟随符号链接，所以一个控制了日志目录的恶意属主也无法把采集器指向任意文件。

这把钥匙还被钉在 relay 的私网地址上，并带有 `restrict`——它开启了 OpenSSH 知道的**全部**限制：不做端口转发、不做 X11 转发、不做 agent 转发、不要 pty、不要 user rc——所以这把钥匙即便泄露，在 relay 之外也毫无用处。那一行完整的条目长这样：

```
from="<relay private IP>",restrict,command="<path>/nhp-xdplog-export.sh" <collector public key>
```

**第 3 跳——采集器做缓存。** 采集器跑在 relay 主机上，每天一次。对每个 (source, day)，缓存或它的 sidecar 缺失、这一天就是今天（它的日志还在写）、或者服务器报告的大小与 mtime 发生了变化时，它就重新拉取。对每一个这样的日子，它执行 `get`，拒绝空下载，裁掉末尾的半行，用 `gzip -n` 压缩，并原子写入。sidecar 正是第三种情况得以成立的原因：它记录拉取当时服务器报告的大小与 mtime，于是下一轮就能区分「没变」与「变了」，而不必拿本地压缩副本去比。缓存目录是 0700，是 relay 上唯一存放未掩码源地址的地方。

这一跳有两个运维细节值得知道：

- 整个运行由一个非阻塞 `flock` 保护。第二次调用——一次定时器 tick 撞上一次手工运行——会成功退出而不做任何事。
- 某个源列不出来时，采集器以非零码退出。服务器只保留大约两周的日志，所以一个悄悄连不上服务器的采集器，是那种「等数据没了才会有人发现」的故障。

**第 4 跳——分析器建站。** `xdplog-report site` 读缓存、写 web root。它对输入是只读的；它唯一会删除的文件，是它自己发布过的、低于下限的详情页。它是增量的：输入没有变动的日子会被跳过，索引也只在有东西变动时才重写。

采集器会传四个可选参数，每个只在它所指的东西确实存在时才传：

| 参数 | 何时传入 | 它改变什么 |
| --- | --- | --- |
| `-geo-mmdb <path>` | relay 上 GeoLite2-City 数据库可读时 | 源地址能解析出 ISO 3166-1 alpha-2 地区；没有它，站点根本不发布地区数字 |
| `-history <path>` | 历史快照可读时 | 逐日记录已丢失的日子改由快照发布 |
| `-force` | 部署要求完整重建时 | 每个缓存的日子都被重新分析，每一页都被重写 |
| `-state <path>` | 总是 | 增量状态存放的位置——刻意放在 web root **之外** |

一个读不了的 `-history` 路径会让运行失败。快照是下限之前那些日子总数的唯一载体，所以悄悄退化成「只数站点自己的日子」会让已发布的全时总数缩小到一个没人能审计的量。

**第 5 跳——nginx 提供服务。** 站点是 `/var/www/stats` 下一个扁平的静态目录，配一个只读的 vhost。这个 vhost 拒绝什么，见第 7 章。

### 源码位置

- `deploy/scripts/nhp-xdplog-export.sh` —— forced command：`list` / `get` 两个动词、日期正则、符号链接拒绝。
- `deploy/scripts/nhp-ac-xdplog-export.sh` —— nhp-ac 用的同一份脚本。
- `deploy/scripts/xdpstats-collect.sh` —— 采集器：缓存布局、flock、保留期清扫，以及它传的参数。
- `deploy/scripts/xdpstats.timer` —— 每日排程。
- `deploy/config-templates/xdpstats/sources.conf` —— 源清单，以及新增一个源需要什么的说明。
- `deploy/nginx/stats.conf.template` —— vhost。
- `endpoints/xdplog-report/site.go` —— `runSite`，增量构建与发布下限。

## 5. 已发布的数字意味着什么

这一章才是站点的读者真正需要的，所以每个字段都值得说精确。

### 逐日的一行

`index.json` 是一个数组，每个 (source, day) 一行。各字段的确切含义：

| 字段 | 含义 |
| --- | --- |
| `source` | 是哪台主机写的日志：`server` 或 `ac`。首页把它们渲染成 `nhp-server` 与 `nhp-ac`；JSON 保留原始名字。 |
| `day` | 这一行覆盖的 UTC 日历日。 |
| `partial` | 这一行覆盖的是一个还在写的日子——通常是当前的 UTC 日。它的数字还会增长。 |
| `drop_pkts` | 当天被丢弃的**攻击包**精确数量，来自 `[NHP-STAT]` 行。 |
| `pass_pkts` | 被放行的包数。在来自历史快照的行上，它的含义是「未记录」，不是零。 |
| `drop_share` | 入站流量中被丢弃的比例：攻击丢弃数除以当天的攻击丢弃、非攻击丢弃与放行之和。 |
| `unique_attack_ips` | 当天**去重**后的攻击源数量。它是逐日的计数，不是逐日的求和，并且它取自被采样的逐包行——见下面*哪些数字是精确的*一节。 |
| `peak_hour` | 被丢弃攻击包最多的那个小时。 |
| `restarts_detected` | 过滤器当天重启的次数，由 `[NHP-STAT]` 里倒退的 `TOTAL` 检出。 |
| `budget_exhausted` | 当天的逐包字节预算用尽，于是逐包行在一天中途停住。报文总数仍然是精确的。 |
| `unknown_reasons` | 出现了站点分类器不认识的原因。这些包被算作攻击。 |
| `top_category` | 当天被丢弃攻击包中占比最高的类别。 |
| `history` | 这一行来自历史快照而非日志。这样的行没有详情页。 |

其中两个需要多说一句。

**`unique_attack_ips` 不能相加。** 一个每天扫描主机的地址是**一个**攻击者。把逐日计数加起来，它就被数了每天一次。因此站点的全时攻击者数字不是各行之和——它是各天之间的**并集**，由站点私下保存的逐日集合算出（见第 6 章）。

**缺失的一个小时不是平静的一个小时。** 已发布的逐小时序列只包含有**攻击类** `[NHP-STAT]` 行的小时——汇总是精确的，而且在当天的逐包字节预算用尽之后依然继续发出，但只有针对攻击原因的 DROP 行会进入逐小时序列。某个小时缺失，意味着过滤器没有为它产出这样的汇总：过滤器被摘下了、守护进程正在重启，或者那个小时里根本没有攻击被汇总过。它**不**意味着那个小时没有发生过攻击。

### 哪些数字是精确的，哪些来自采样

并不是每个已发布数字都来自同一个通道，这个区别很重要：一个假定它们全都精确的读者，会过度解读其中那两个并不精确的。

精确——它们来自 `[NHP-STAT]` 行，这一路不受限速：

| 数字 | 字段 |
| --- | --- |
| 被丢弃的攻击包 | `drop_pkts` |
| 被放行的包 | `pass_pkts` |
| 入站流量被丢弃的占比 | `drop_share` |
| 最忙的小时 | `peak_hour` |
| 占比最大的类别 | `top_category` |
| 过滤器重启次数 | `restarts_detected` |
| 字节预算是否用尽 | `budget_exhausted` |
| 是否出现未分类原因 | `unknown_reasons` |

采样——它们来自逐包的 `[NHP-DROP]` 行，而这一路被过滤器限速：

| 数字 | 字段 |
| --- | --- |
| 当天去重的攻击源数 | `unique_attack_ips` |
| 全时去重的攻击源数 | `NHP_TOTAL_ATTACK_SOURCE_IPS` |
| 全时去重的地区数 | `NHP_TOTAL_ATTACK_REGIONS` |

一个源如果它的包在记录下来之前就全被限速丢掉了，它就不会出现在采样出来的数字里，所以去重源数与地区数是真正攻击过的主机的源的**下界**：它们数的是过滤器来得及写下来的那些攻击者。报文总数不受影响——它来自精确的那一路——重启次数、预算标记和未知原因同样不受影响。而历史基线在生效时，是唯一能把源数字往另一个方向推成上界的东西；见下面*历史基线*一节。

### 全时数字

首页那个醒目的数字，也是 `totalnumber.js` 里的第一个常量，是**每个已发布日**被丢弃的攻击包的**精确**总数。它是可复现的：把 `index.json` 每一行的 `drop_pkts` 加起来，得到的就是同一个数字。

另外两个全时数字是**并集**，不是求和：

- **去重源 IP 数** —— 至少在某一个已发布日出现过的地址数，每个只数一次。
- **去重地区数** —— 这些地址解析出的不同 ISO 3166-1 alpha-2 代码数。GeoIP 数据库放不下的源不属于任何地区，只抬高源计数。

两者都无法从 `index.json` 复现，因为它们背后的逐日集合不能被发布（见第 7 章）。

### 发布下限

站点不发布 `2026-10-16` 之前的内容。这个日期是 `endpoints/xdplog-report/site.go` 里的 `defaultSiteSince`，它是一个**发布**决定，不是数据可用性决定：XDP 过滤器挂到 demo 主机上的时间，晚于它日志目录里早就躺着的那些日志，所以更早的一切都是残缺记录。数进去会让总数少一个没人能审计的量。

低于下限的日子不被分析、不被渲染、不进索引、也不计入总数。一个把所有已发现日子都排除掉的 `-since` 会在**删除任何东西之前**被拒绝——一次笔误既不能清空一个已填充的索引，也不能留下一个全是死链的站点。

### 历史基线

历史快照是站点自己数字的一份存档，覆盖采集器缓存已经够不到的日子。它做两件事：发布它列出的、逐日记录已消失的那些日子；以及给全时计数器打底。

它贡献什么，取决于它能证明什么：

| 状态 | 含义 | 数字变成什么 |
| --- | --- | --- |
| Applied | 快照中部分日子的逐日记录已丢失，快照自身的源计数替它们顶上 | 源 IP 数字是**上界**——两个时期都出现过的地址可能被数两次。地区数字不受影响。 |
| Missing | 某个已发布日子的逐日记录丢失，且没有东西替它顶上 | 两个数字都是**下界** |
| Below the floor | 快照列出了站点不发布的日子 | 基线**整体被丢弃**——快照级的总数无法按天拆分——两个数字都是下界 |

快照自身的**地区**计数在任何状态下都**不会**被采用。地区是一组大约 250 个代码的封闭集合，两个时期几乎完全共享，因此任何把两者做算术的做法都只会是错的：相加会把绝大多数地区数两遍，取较大者又会漏掉只在其中一个时期出现过的地区。已发布的地区总数就是**各已发布日的并集**，别无其它。

首页与 `totalnumber.js` 都会逐个地打印出每个数字是哪一种：精确、上界还是下界。两者都精确时，它们什么都不说。

### 源码位置

- `endpoints/xdplog-report/site.go` —— `SiteDaySummary`（已发布行的契约）、`defaultSiteSince`、`siteRenderVersion`。
- `endpoints/xdplog-report/analyze.go` —— `Report`、`maskIP`，以及 `isAttackReason` 对各计数的影响。
- `endpoints/xdplog-report/classify.go` —— `nonAttackCategories`、`isAttackReason`。
- `endpoints/xdplog-report/history.go` —— `historyBaseline`、`sourcesBound`、`regionsBound`、`boundNote`、`BelowFloor`。
- `endpoints/xdplog-report/facts.go` —— `unionDayFacts`，并集所依据的逐日集合。
- `endpoints/xdplog-report/geo.go` —— `isReserved`，永远不做地理查询的地址段。

## 6. 发布产物及其契约

站点是一个静态文件的扁平目录。背后没有应用服务器，也没有数据库。

| 产物 | 它是什么 | 谁读它 |
| --- | --- | --- |
| `index.html` | 首页：全时醒目的总数、计数起始日期，以及每天一行 | 浏览器 |
| `index.json` | 机器可读索引：每个 (source, day) 一个 `SiteDaySummary` 对象 | 任何想把数据画成图的东西 |
| `d/<source>/<day>.html` | 某一天的详情页 | 浏览器 |
| `d/<source>/<day>.json` | 某一天的 `Report`，schema 2 | 任何想要明细的东西 |
| `totalnumber.js` | 全时数字的三个 JavaScript 常量 | 另一个站点，通过 `<script src>` |
| `countup.js` | 首页计数动画 | 浏览器 |

**`index.json` 是权威的。** `totalnumber.js` 是它的兄弟，不是它的替代：里面的报文总数可以通过对 `index.json` 的 `drop_pkts` 求和复现，而另外两个数字不能——它们是并集，背后的逐日集合是私有的。只想要数字的消费者应该用 `totalnumber.js`；想核对数字的消费者应该用 `index.json`。

**`totalnumber.js` 导出三个常量：**

```javascript
const NHP_TOTAL_ATTACK_BLOCKED = <packets>;
const NHP_TOTAL_ATTACK_SOURCE_IPS = <addresses>;
const NHP_TOTAL_ATTACK_REGIONS = <regions>;
```

即使在空站点上这三个常量也会被写出，值为 `0`，这样引用它的页面不会拿到 404 和一个悬空引用。第三个常量数的是**地区**：国际标准分配给它列出的每一个国家和区域的两字母 ISO 3166-1 alpha-2 代码，其中也包括属地、争议地区和特别政治区域，而不仅仅是主权国家。它曾经叫 `NHP_TOTAL_ATTACK_COUNTRIES`，那个名字已经没有了——一个叫做国家数的已发布数字，是对那些代码提出了标准并不支持的主张。旧名字在 `history` 子命令解析旧文件时**仍会被读取**，但永远不会被写出。

**详情页的 JSON 是一个 `Report`，schema 2。** 它带着当天的总数（`total_pkts`、`drop_pkts`、`pass_pkts`、`drop_share`、`non_attack_drop_pkts`）、按原因与按类别的分解、top 来源与 top 目的端口、逐小时序列、协议分布、重启列表和未知原因。逐小时序列只对有**攻击类** `[NHP-STAT]` 行的小时存在，这正是「缺失的一个小时不是平静的一个小时」的原因。它还会说明采样出来的数字有多有代表性：`drop_sampled_events` 是被逐包写出来的被丢弃包数，`drop_sampling_rate` 是这个数除以当天精确的丢弃总数——也就是过滤器真正写下来的那部分丢弃所占的份额。

**首页会渲染源的显示名，JSON 不会。** `server` 在一张其余全是日期和数字的表里读起来像个泛指名词，所以首页显示 `nhp-server` 与 `nhp-ac`。`source` 是 `index.json` 的已发布字段，改名会破坏任何已经在解析它的人，所以 JSON 保留原始目录名。

**站点不加载任何第三方资源。** 它引用的两个脚本——`countup.js` 与 `totalnumber.js`——都是构建写出的同源文件，内容安全策略只允许脚本来自站点自己的源（`script-src 'self'`）。自包含只有一处被记录的例外：首页的 `<head>` 里带着 canonical URL 和 Open Graph / Twitter card 标签，它们的值是绝对 URL，因为爬虫会读它们而浏览器从不取它们。

计数器是渐进增强。首页那个数字已经写在 HTML 里了，所以没有 JavaScript、没有那条内容安全策略、或者开了减少动效的读者都能直接读到；脚本只是把它倒回去再数上来。

**模板改动会到达已发布的页面。** 每天的状态里都记着产生该页的 `siteRenderVersion`，所以把这个常量加一，下一次运行就会重渲每一个已发布的日子。那些原始日志已经被清掉的日子，会由它们已发布的 JSON 重新渲染——HTML 是它的纯函数——但它们的数字绝不重算。

### 源码位置

- `endpoints/xdplog-report/site.go` —— `SiteDaySummary`、`SiteDayState`、`SiteState`、`siteRenderVersion`、`siteStateSchema`。
- `endpoints/xdplog-report/render_index.go` —— `siteURL`、`siteOGImage`、`sourceLabels`、`sourceLabel`、`writeTotalNumberJS`、`totalNumberJS`、`writeCountUpJS`。
- `endpoints/xdplog-report/analyze.go` —— `Report`、`reportSchema`。
- `endpoints/xdplog-report/site_test.go` —— `assertSelfContained`，把站点自身的引用范围钉住的测试。

## 7. 发布了什么，没发布什么

站点是一个刻意公开、只读的攻击遥测摘要。没有登录，没有限流，没有地域限制。

**源地址被掩码。** IPv4 地址截断成 `a.b.c.x`；IPv6 地址截断到前 48 位。`site` 子命令的 `-mask-ip` 默认是 true，这与单文件命令行工具的默认值（false）不同。掩码是在站点这一侧做的，不是过滤器做的：服务器上的原始日志永远是完整地址。

**私有与保留地址段被排除**在攻击源计数之外，也永远不做地理查询。GeoIP 数据库放不下的源计入源数字，但不属于任何地区。

**已发布的产物里没有完整地址，私有的那些有。** 三个目录，三套权限：

| 目录 | 权限 | 装的是什么 |
| --- | --- | --- |
| web root | `/var/www` 下 0755 | 已发布的站点：只有掩码后的地址 |
| 原始缓存 | `$HOME` 下 0700 | gzip 后的日志，带着**完整**源地址——relay 上唯一的未掩码副本 |
| 状态目录 | `$HOME` 下 0700 | `state.json`，以及逐日的 *facts* 文件 |

状态文件记录着原始缓存里每个文件的绝对路径，那是匿名访客没有资格读到的宿主布局。facts 文件装着全时并集所依据的逐日「已哈希源地址集合」与地区码集合。两者都住在 web root 之外，而 nginx vhost 还第二次拒绝 `/state.json` 与 `/facts/`，以防将来某次改动把它们挪回里面。

facts 文件存的是地址的 64 位哈希，从不存地址本身。但一个 IPv4 地址的 64 位哈希在整个 2^32 空间上是可被暴力枚举反解的——把哈希发布出去，等于把掩码藏起来的东西原样交还。这正是 facts 文件保持私有、并且与状态文件并排而不是放在 web root 下的原因。

web root 之所以在 `/var/www` 而不是 `$HOME` 下，理由很朴素：在 Amazon Linux 2023 上 `/home/ec2-user` 是 0700，而 nginx worker 以 `nginx` 身份运行，所以放在 `$HOME` 下的 web root 会对每个请求都回 403。

### 源码位置

- `endpoints/xdplog-report/analyze.go` —— `maskIP`。
- `endpoints/xdplog-report/geo.go` —— `isReserved`。
- `endpoints/xdplog-report/facts.go` —— facts 文件格式、它的 magic 与它的哈希。
- `deploy/scripts/xdpstats-collect.sh` —— 三个目录根与它们的权限。
- `deploy/nginx/stats.conf.template` —— 对 `/state.json` 与 `/facts/` 的拒绝。

## 8. 运行与运维

这条流水线的运维真相在 runbook 里，不在本文里。本章说清楚什么在跑、什么时候跑，以及细节在哪里；它刻意不复制 runbook 的操作步骤，因为那只会产生一份会腐烂的副本。

**每日运行。** relay 主机上的一个 systemd 定时器在每天 00:20 UTC 触发采集器——比 UTC 日切整整晚一个小时，好让一个慢的写端不会留下写了一半的前一天日志。定时器是 `Persistent=true` 的，所以在窗口期处于关机的宿主会在下次开机时把服务跑一次补上。

**强制重建。** `xdplog-report site` 只在有输入变动时才重写索引。因此一次改变了站点所发布内容的部署（发布出来的数字、醒目的数字、标记），会让上一个构建继续对外服务，直到某个日志恰好发生变化。部署通过 systemd manager 的环境请求一次强制重建，于是这次运行仍然发生在 unit 的沙箱之内，采集器据此传上 `-force`。每日定时器那次运行不这么做：在那里，增量路径正是重点。

**部署顺序。** `stats.opennhp.org` 的 A 记录是一个 Terraform 资源，而部署工作流从不 apply Terraform——它只读状态。因此引入站点改动的发布，必须先 `infra-demo`（apply），再 `deploy-demo-v2`（deploy）。部署的第一步是一次预检，要求 `stats.opennhp.org` 恰好解析出一条 A 记录，且等于 relay 的公网地址；忘记 apply 会在碰到 relay 之前就红，而不是事后一个红的 curl。

**部署自己的门禁。** 部署会检查三件关于它自己发起的那次运行的事，因为采集器的退出码回答不了其中任何一件：collect 日志里出现了强制重建那一行；已发布的 `index.html` 比运行前创建的标记更新；以及已发布的地区数字是一个合理的数——ISO 3166-1 alpha-2 代码只有大约 250 个，更大的数字意味着算错了。某个源列不出来只报为警告而不是失败，因为站点本身依然完好；它只是可能陈旧。

**新增一个源。** 一个新源需要四样东西：`sources.conf` 里的一行（名字必须是 `server` 或 `ac`，因为日志前缀 `nhp_<name>_xdp` 是由它推导出来的）、该主机上的一个导出脚本、装进该主机 `authorized_keys` 并挂在该脚本之后的采集器公钥，以及一个能理解该日志形状的解析器。nhp-ac 的导出脚本与 nhp-server 的几乎一样；两台主机的数据进的是同一个已发布站点。

**操作步骤在哪里。** `terraform/demo/RUNBOOK.md` 承载部署与回滚步骤、恢复路径、定时器窗口错过之后的手工补录流程，以及密钥轮换流程。`terraform/demo/VERIFY-server-xdp.zh-cn.md` 承载过滤器改动的 netns 演练、上线前基线，以及部署之后要再跑一遍的同一批命令。

### 源码位置

- `deploy/scripts/xdpstats.timer` —— 排程。
- `deploy/scripts/xdpstats-collect.sh` —— `-f` / `XDPSTATS_SITE_FORCE` 这条路径。
- `.github/workflows/deploy-demo-v2.yml` —— `deploy-xdpstats` job：它的预检、它的强制重建，以及它的门禁。
- `terraform/demo/RUNBOOK.md` —— *XDP attack stats site (stats.opennhp.org)*、*Analysing the XDP event log (attack report)*、*nhp-server ingress filter (XDP)* 三节。
- `terraform/demo/VERIFY-server-xdp.zh-cn.md` —— 核验流程。

## 9. 接下来读什么

| 你想知道 | 去哪里 |
| --- | --- |
| 过滤器如何决定一个包的命运，以及 `etc/xdp.toml` 里每个字段控制什么 | [eBPF 原理](/zh-cn/ebpf/) |
| 仓库里各部分的位置 | [源代码解读](/zh-cn/code/) |
| 分析器的命令行、每个参数，以及报告的字段 | `endpoints/xdplog-report/README.md` |
| 部署、回滚与恢复统计站 | `terraform/demo/RUNBOOK.md` |
| 在过滤器改动前后做核验 | `terraform/demo/VERIFY-server-xdp.zh-cn.md` |
| 过滤器刻意不去解析的协议 | [协议参考](/zh-cn/protocol/) |
| 这一层实现的默认拒绝性质 | [功能列表](/zh-cn/features/) |

站点本身在 <https://stats.opennhp.org/>。它的 `index.json` 是「过滤器丢了什么」这个问题的权威、机器可读的答案，而 `totalnumber.js` 是同一个答案的短形式，供想要嵌入它的页面使用。
