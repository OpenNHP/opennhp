---
layout: page
title: eBPF 原理
nav_order: 15
permalink: /zh-cn/ebpf/
description: OpenNHP 的 XDP/eBPF 层如何决定一个包的命运，以及 etc/xdp.toml 里的每个开关到底控制什么。
---

# OpenNHP 的 eBPF 层是如何工作的
{: .fs-9 }

OpenNHP 有两层过滤：一层在内核中，用 XDP 程序决定一个包是否应该存在；另一层在用户态，用 Noise 握手决定它属于谁。本文讲的是第一层——eBPF 程序、加载它们的加载器，以及驱动它们的配置。
{: .fs-6 .fw-300 }


---

如果你要写一个 NHP 集成，下面三个问题一定会先出现：

1. **我的敲门在哪一层被决定？** 答：分两步。内核里的 XDP 程序决定这个数据报是否**能到达守护进程**——它挂在主机默认路由网卡上，早于网络协议栈看到它之前。而敲门**是否有效**，是由用户态的 Noise 握手决定的。
2. **我改了 `etc/xdp.toml` 里的某个字段，它生效吗？** 答：只有 `RelayIPs` 生效，其他字段都在启动时读一次。第 7 章会逐一钉死哪个是哪个。
3. **过滤器没挂上，我该看哪里？** 答：守护进程的日志，以及二进制旁边的 `etc/nhp_server_xdp.o`。第 8 章讲这些日志行以及它们的含义。

本文解释上面三个答案背后的机制。文中每一条陈述都能对应到仓库里的具体文件；每个组件第一次出现时都会给出源码路径。代码是权威：本文引用某个文件时，引用是逐字的；两者若有出入，以源码为准。

## 目录

1. [为什么要在内核层过滤](#1-为什么要在内核层过滤)
2. [一页的 eBPF](#2-一页的-ebpf)
3. [拓扑：三个程序、两个守护进程、一个加载器](#3-拓扑三个程序两个守护进程一个加载器)
4. [nhp-server 入站过滤器：逐包看一遍](#4-nhp-server-入站过滤器逐包看一遍)
5. [nhp-ac 的按次放行过滤器](#5-nhp-ac-的按次放行过滤器)
6. [加载器：从 `.o` 到一个真正在跑的过滤器](#6-加载器从-o-到一个真正在跑的过滤器)
7. [配置：`etc/xdp.toml` 到底控制什么](#7-配置etcxdptoml-到底控制什么)
8. [可观测性与排障](#8-可观测性与排障)

## 1. 为什么要在内核层过滤

NHP 是一个"先认证、后连接"的协议。受保护的服务在敲门被认证之前不应该是可达的，所以真正的问题不是"这个包是谁发来的"，而是"这个包该不该被发到这台主机上来"。

在 OpenNHP 里这个问题由两个不同的地方回答，而这个划分很重要：

```
   ┌──────────────────────────────────────────────────────────────────┐
   │  the wire                                                        │
   └───────────────────────────┬──────────────────────────────────────┘
                               │  frame arrives at the NIC
                               ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │  KERNEL — XDP hook        nhp_server_xdp.o                       │
   │                                                                  │
   │    "should this packet exist at all?"                            │
   │                                                                  │
   │    XDP_DROP ──────────────────►  the frame dies here.            │
   │                                  No netfilter, no routing,        │
   │                                  no socket, no daemon.            │
   │    XDP_PASS ──────────────────►  ▼                               │
   └──────────────────────────────────────────────────────────────────┘
                               ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │  KERNEL — netfilter, routing, TCP/UDP stack, socket lookup      │
   └───────────────────────────┬──────────────────────────────────────┘
                               ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │  USER SPACE — nhp-serverd                                       │
   │    "who is this, and are they allowed?"                          │
   │    Noise handshake, identity, ACL, session state                 │
   └──────────────────────────────────────────────────────────────────┘
```

XDP 程序回答第一个问题。它**不解析 NHP 协议**——身份仍然由用户态的 Noise 握手决定（`nhp/core/packet.go` 里的 `RecvPrecheck`）。内核唯一与 NHP 有关的检查是一个最小数据报长度，这足以让敲门端口对普通扫描器变成一个"无应答地址"，而且代价为零。

那为什么要放到内核里？三个理由，全都能在代码里看到：

- **代价。** 在 XDP 被丢弃的包不会走 netfilter、不会走路由、不会到达 socket 查找、也不会唤醒任何守护进程。在一台被持续扫描的主机上，这是"什么都不做"和"替别人干活"的区别。
- **时机。** XDP 跑在 netfilter、路由和 socket 查找之前。在那里丢弃是终局的：主机不会有一个 socket 可以把这个包排队，因此没有 slowloris 可利用，也没有半开状态需要清理。
- **可达性。** 守护进程自己的监听端口只有一个 UDP 端口。其他一切——SSH、HTTP 敲门监听器、metrics 端点、以后有人绑定的任何服务——在驱动层都受这个过滤器管辖，不管守护进程知不知道它们。

最后一条有一个值得记住的推论：nhp-server 的过滤器是**主机**过滤器，不是**守护进程**过滤器。它会丢弃到主机上任何服务的入站 TCP，不只是到 `nhp-serverd` 的。第 8 章会讲这与某个监听端口冲突时会发生什么。

## 2. 一页的 eBPF

要改这一层不需要成为内核程序员，但有五个概念承载了大部分分量。

**验证器（verifier）。** 当一个 BPF 程序被加载时，内核会对它的指令做一遍静态分析。如果这个程序可能死循环、可能读它无权读取的内存，或者可能做任何其他未定义的行为，加载会被拒绝——没有部分加载，也没有"先跑着以后再说"。这就是 BPF 程序都很短的原因：验证器要推理每一条路径。

**挂载点（hook）。** 程序挂在一个**钩子点**上，而不是挂在某个文件或某个守护进程上。OpenNHP 用了两个：

| 挂载点 | 什么时候运行 | 谁在用 |
|--------|--------------|--------|
| `SEC("xdp")` | 入站，网络协议栈之前 | server 过滤器、AC 的入站过滤器 |
| `SEC("tc/egress")` | 出站，路由决策之后 | AC 的出站过滤器 |

**Map。** BPF 程序不能把状态放在普通内存里。它记得的一切都放在 *map* 里——一种内核数据结构，用户态也能读和写。OpenNHP 用到 `LPM_TRIE`（最长前缀匹配，用于 SSH 白名单）、`PERF_EVENT_ARRAY`（用户态抽取的环形缓冲）、`PERCPU_ARRAY`（每 CPU 计数器，因此自增不需要原子操作）、`LRU_PERCPU_HASH`（用于扫描聚合的有界表）和 `LRU_HASH`（AC 的白名单）。

**CO-RE。** 这些程序是针对 `nhp/ebpf/xdp/vmlinux.h` 编译的——那是从某个内核的 BTF 生成的头文件，带着 clang 的 `preserve_access_index` 属性。编译器因此发出重定位信息而不是写死字段偏移，加载器在加载时针对*目标*内核的 BTF 解析它们。于是同一个目标文件可以跨内核工作，不必重新编译。

**能力（capabilities）。** 三个，各覆盖一步：

| 能力 | 用于 |
|------|------|
| `CAP_BPF` | 加载程序、创建它的 map——以及之后写入 map，但仅限于 `bpf(2)` 本身就是特权调用的内核（见第 6 章） |
| `CAP_NET_ADMIN` | 把程序挂到网卡上 |
| `CAP_PERFMON` | 打开事件日志要读的 perf ring |

OpenNHP 在启动时把这三个能力授予 `nhp-serverd`，并在加载决策一做出就收回来——在 `bpf(2)` 本身就是特权调用的内核上，`CAP_BPF` 除外。细节见第 6 章。

Go 这一侧，加载器使用 [`github.com/cilium/ebpf`](https://github.com/cilium/ebpf) v0.22.0（`nhp/go.mod`）——一个纯 Go 库，直接与 `bpf(2)` 打交道：它自己解析 ELF 目标文件、自己应用 CO-RE 重定位，并提供 `link` 和 `perf` 两个辅助包。它并不链接 libbpf——但 C 源码用的是 libbpf 的头文件约定（`SEC()`、`LIBBPF_PIN_BY_NAME`），所以本文里这些名字来自那里。

## 3. 拓扑：三个程序、两个守护进程、一个加载器

树里有三个 eBPF 程序，分属两个不同的守护进程：

| 程序 | 源码 | 挂载点 | 由谁挂载 | 作用 |
|------|------|--------|----------|------|
| `nhp_server_xdp.o` | `nhp/ebpf/xdp/nhp_server_xdp.c` | `SEC("xdp")` | `nhp-serverd` | 整台主机固定的入站策略：敲门端口 + 白名单来源的 SSH + 主机自己发起流的应答。 |
| `nhp_ebpf_xdp.o` | `nhp/ebpf/xdp/nhp_ebpf_xdp.c` | `SEC("xdp")` | `nhp-acd` | **按次敲门**打开受保护端口，并在敲门过期后把它们关上。 |
| `tc_egress.o` | `nhp/ebpf/xdp/tc_egress.c` | `SEC("tc/egress")` | `nhp-acd` | AC 过滤器的另一半：让它**自己**发出的流能把应答放进来，并记录它发过什么。 |

server 和 AC 完全没有共用的 map。server 的策略是"只有这两样东西可以到达我"，AC 的策略是"这个来源，这个端口，直到这次敲门过期"。两者之间无话可说，所以也不共用一套 schema。

它们真正共用的是 map 周边的一切——memlock 处理、目标文件加载、在 `/sys/fs/bpf` 下做 pin、解析默认路由网卡、挂载、把 perf ring 抽成日志、再把 pin 拆掉。这些都在一个文件里：`nhp/utils/ebpf/engine_linux.go`，由调用方选择变体：

```go
// EngineVariant selects which object file, map set and attachment points a load
// uses. VariantAC is iota so that a zero-value EngineLoadParams keeps the AC
// behavior it had before this package existed.
type EngineVariant int

const (
	VariantAC EngineVariant = iota
	VariantServer
)
```

server 自己的封装是 `endpoints/server/ebpf/serverengine.go`，AC 的是 `endpoints/ac/ebpf/ebpfegine.go`。两者都很薄——只负责填路径和参数，然后交给共享加载器。

### 编译目标文件

```bash
make ebpf        # objects, then the full build
make ebpf-objs   # objects only
```

eBPF 目标**刻意不是** `acd` / `serverd` 的前置依赖：

```makefile
# Deliberately NOT a prerequisite of `acd` / `serverd`: those must keep
# building on a host with no clang, which is why the daemon targets and the
# object targets are separate goals the pipeline invokes in sequence. Both
# daemons treat a missing object as a runtime condition, not a build error
# (nhp-acd refuses to start in FilterMode=1; nhp-serverd falls open).
```

目标文件会落在守护进程二进制旁边：

```
release/nhp-ac/etc/nhp_ebpf_xdp.o
release/nhp-ac/etc/tc_egress.o
release/nhp-server/etc/nhp_server_xdp.o
```

因此"缺少目标文件"是一个**运行时**状态，有两种不同的答案——AC 拒绝启动，server 不带过滤运行。第 7 章会回到这个分歧，因为它是这一层里最出人意料的事情。

注意 `docker/Dockerfile.server` 用的是 `make init serverd plugins`，并不编译 eBPF 目标文件，所以容器镜像里根本没有 `nhp_server_xdp.o`。容器化的 `nhp-serverd` 因此不带入站过滤器运行——对于一台暴露面由别处管理的主机来说，这正是正确的 fail-open 答案。

## 4. nhp-server 入站过滤器：逐包看一遍

程序是 `nhp/ebpf/xdp/nhp_server_xdp.c`，入口是：

```c
SEC("xdp")
int xdp_server_prog(struct xdp_md *ctx) {
```

许可证是 `Dual BSD/GPL`。下面的判定树就是全部策略——没有第二阶段。

### 判定树

```
frame
 │
 ├─ ethertype?
 │    ├─ ARP                    → PASS   (not reported; keeps the host on its subnet)
 │    ├─ IPv6                   → handle_ipv6()          [see below]
 │    ├─ IPv4                   → continue
 │    └─ anything else          → DROP   (not recorded)
 │
 ├─ iph->ihl < 5                → DROP   (not recorded)
 │
 ├─ non-first fragment?         → PASS   (IPV4_FRAGMENT — see "fragments" below)
 │
 ├─ TCP
 │    ├─ dport 22?
 │    │    ├─ source in nhp_relay_ips → PASS   (SSH_RELAY; reported on the SYN only)
 │    │    └─ otherwise              → DROP   (TCP_SSH_OTHER)
 │    ├─ dport == nhp_listen_port   → DROP   (TCP_NHP_PORT — NHP is UDP-only)
 │    ├─ has_local_flow()?          → PASS   (TCP_ESTABLISHED; counted, not reported)
 │    ├─ SYN without ACK            → DROP   (TCP_SYN_SCAN)
 │    ├─ no SYN, ACK or RST         → DROP   (TCP_STEALTH_SCAN)
 │    └─ otherwise                  → DROP   (TCP_OTHER)
 │
 ├─ UDP
 │    ├─ dport != nhp_listen_port
 │    │    ├─ 67 → 68 (DHCP)        → PASS   (DHCP_CLIENT; reported)
 │    │    ├─ 123 → 123 (NTP)       → PASS   (NTP_CLIENT; counted)
 │    │    ├─ has_local_flow()?      → PASS   (UDP_ESTABLISHED; counted)
 │    │    └─ otherwise              → DROP   (UDP_OTHER)
 │    └─ dport == nhp_listen_port
 │         ├─ source in nhp_relay_ips → PASS   (NHP_RELAY — exempt from the length floor)
 │         ├─ udp->len < nhp_min_udp_len → DROP  (UDP_SHORT)
 │         └─ otherwise               → PASS   (NHP_DEFAULT)
 │
 ├─ ICMP
 │    ├─ type 3 code 4 (frag needed) → PASS   (ICMP_FRAG_NEEDED)
 │    └─ otherwise                    → DROP   (NON_TCP_UDP)
 │
 └─ anything else               → DROP   (NON_TCP_UDP)
```

其中两个分支值得单独一节，因为它们正是读者最容易搞错的地方。

注意那些"(not recorded)"标记。一个畸形的包——头被截断、`ihl` 小于 5、未知以太类型——会在没有 action code 的情况下被丢掉，因此既不计数也不上报。每一个*属于策略决定*的丢弃都带着一个 action code；没有 code 的那些与一个损坏的帧无从区分。`[NHP-STAT]` 的总数对每一个被分类的类别都是精确的。

程序是以**通用模式**（`link.XDPGenericMode`）挂载的，不是驱动模式：程序看到帧的时候，它已经被转成 skb 了，所以一次丢弃省下的是协议栈的活，而不是那次分配。驱动模式跑得更早、连分配也一起省了，但它需要驱动支持、也不可移植——这一点很重要，因为过滤器是挂在主机默认路由所走的那块网卡上的。

### socket 表：主机自己发出去的请求怎么把应答收回来

server 既是服务端也是客户端。它解析域名、它的认证插件向 SES 提交 OTP 邮件、主机还访问软件包镜像。一个只放行敲门和 SSH 的过滤器会把这些的**应答**全部丢掉，而最初的症状出在解析器上：`read udp 10.0.1.78:58655->10.0.0.2:53: i/o timeout`——DNS 的**应答**被丢了，连带后面的 OTP 邮件也没了。

AC 用一个 TC 出站程序加一张 `conn_track` map 来解决这个问题，因为它必须推理"被敲门授权的流"。server 没有这种歧义，于是它直接问内核自己的 socket 表：

```c
    owned = is_tcp ? sk->state != BPF_TCP_LISTEN : sk->dst_port != 0;
```

这就是全部规则，等价于 netfilter 的 `--ctstate ESTABLISHED,RELATED -j ACCEPT`：

- **TCP** —— 除 `BPF_TCP_LISTEN` 以外的任何状态。监听 socket 会匹配发往主机上任意服务端口的每一个 SYN，所以放行它等于在某天有人绑了 tcp/8080 之后悄悄废掉整个过滤器。
- **UDP** —— 只认已连接（`dst_port != 0`）的 socket。解析器、NTP 客户端或 `net/smtp` 的拨号器都会 `connect(2)` 自己的 socket，因此内核的查找只会匹配来自它正在对话的那个对端的报文。一个只是绑定在某个端口上的未连接 socket，换个名字也是监听 socket。

查找走的是 `bpf_sk_lookup_tcp()` / `bpf_sk_lookup_udp()`，带 `BPF_F_CURRENT_NETNS`，它解析到程序所挂网卡的命名空间——也就是 `nhp-serverd` 的 socket 所在之处。代价是一次哈希查找，而且只发生在本来就要被丢弃的包上；这正是如果没有 XDP、网络栈也要做的那次查找。

有两个客户端协议没法这样回答，只能按端口对放行：

- **DHCP** —— 客户端是在裸 `AF_PACKET` socket 上、或者仅仅绑定在 udp/68 的 socket 上收到租约的，所以 socket 查找什么也找不到。丢掉续租报文是致命的：租约过期、地址消失，主机在**所有**端口上都不见了。67 → 68 这一对足够窄，邻居广播的 DHCPDISCOVER（目的端口 udp/67）仍然会被丢掉。
- **NTP** —— chronyd 是从它自己*绑定*在 udp/123 的 socket 发出轮询的，所以应答（123 → 123）看上去和一次主动查询一模一样。丢掉它损失的是时钟而不是地址，而 NHP 自己会察觉到：敲门报文带时间戳，server 会拿它跟一个时间窗比对。

这两个都不是"凡是发给已绑定端口的一律放行"这种通用例外，因为那会把"没有主动报文能进来"这个性质整个丢掉。

### IPv6，以及为什么 SSH 必须走 IPv4

`handle_ipv6()` 跑的是同一棵判定树，只是去掉了白名单。白名单是一棵 **IPv4 前缀**的 trie，所以程序在 v6 上认不出 relay——任何东西都到不了 tcp/22 或敲门端口。

例外是那些让 v6 主机还能工作的协议：ICMPv6 的 "packet too big"，以及连续的一段 MLD/ND（类型 130–137，其中包含邻居发现——v6 的 ARP）、DHCPv6（547 → 546）、NTP，以及通过 `has_local_flow6()` 放行的、主机自己 v6 流的应答。其余都是 `V6_OTHER`。

加载器里的 `warnOnGlobalIPv6()` 会在被过滤网卡拥有全局 v6 地址时打一条 `Warning` 并点名该地址，因为在一台唯一 SSH 通路是 v6 的主机上，这会是一次自锁。

### 分片

IPv4 数据报只有第一个分片带有 L4 头，因此只有它能被基于端口的策略判定。其余分片被**放行**，并使用自己的 action code（`IPV4_FRAGMENT`）：

- 它们自己无法送达任何 socket——重组需要第一个分片，而那个分片已经走完了整棵树。
- 相反把它们丢掉会让每一个大于路径 MTU 的数据报失败，敲门也不例外。
- 从分片的载荷里"读出"端口（这个程序以前就是这么干的）比前两者都糟：一个载荷碰巧读成 67 → 68 的构造分片能骗过 DHCP 例外。

IPv6 分片头中偏移非零的也一样处理（`IPV6_FRAGMENT`）。

### 各个 map

| Map | 类型 | 容量 | 装的是什么 |
|-----|------|------|-----------|
| `nhp_relay_ips` | `LPM_TRIE` | 4096，`BPF_F_NO_PREALLOC`，已 pin | SSH 白名单。对 IPv4 前缀做最长前缀匹配。 |
| `nhp_events` | `PERF_EVENT_ARRAY` | 1024 | 用户态抽取的 perf ring。 |
| `nhp_event_rate` | `PERCPU_ARRAY` | 33 | 令牌桶：32 个 action 分类 + 1 个全局。 |
| `nhp_action_stats` | `PERCPU_ARRAY` | 32 | 每个 action 分类的 `packets` 与 `logged` 计数。 |
| `nhp_port_hits` | `LRU_PERCPU_HASH` | 65536 | 按目的端口统计的被丢弃报文数。 |
| `nhp_src_hits` | `LRU_PERCPU_HASH` | 65536 | 按 IPv4 源地址统计的被丢弃报文数。 |

用 trie 而不是一张主机地址的哈希表是刻意的：白名单必须能扛住 relay 被**替换**。relay 的私网地址没有被 Terraform 固定，所以任何一次重建都会给它一个新地址——而一张只写旧地址的白名单会把新 relay 的 SSH 挡在门外，而那条路正是 CI 推送修正后白名单的唯一途径。把 relay 所在网段作为前缀写进去就堵住了这个坑。一个裸地址其实就是 /32。

那两张 `LRU_PERCPU_HASH` 表存在的原因是：一次 65535 端口的扫描不能把表撑爆；LRU 给表封了上界，用户态在怀疑发生淘汰时会报 `LRU=1`。精确的分类总数在 `nhp_action_stats` 里，不受影响。

### action code

每个**被分类**的报文都会得到一个——畸形帧根本不会有 action code（见上面判定树后的说明）。C 侧在 `nhp_server_xdp.c` 里定义这些 code，Go 侧在 `engine_linux.go` 的 `serverActionName` 里镜像它们，日志行把它们打印成 `REASON=`：

| Code | Reason | 判定 | 是否逐包记录 |
|------|--------|------|--------------|
| 1 | `SSH_RELAY` | PASS | 仅 SYN |
| 2 | `NHP_RELAY` | PASS | 是 |
| 3 | `NHP_DEFAULT` | PASS | 是 |
| 4 | `TCP_ESTABLISHED` | PASS | 只计数 |
| 5 | `UDP_ESTABLISHED` | PASS | 只计数 |
| 6 | `ICMP_FRAG_NEEDED` | PASS | 是 |
| 7 | `DHCP_CLIENT` | PASS | 是 |
| 8 | `NTP_CLIENT` | PASS | 只计数 |
| 16 | `IPV4_FRAGMENT` | PASS | 是 |
| 17 | `ICMPV6_CONTROL` | PASS | 只计数 |
| 18 | `V6_ESTABLISHED` | PASS | 只计数 |
| 19 | `IPV6_FRAGMENT` | PASS | 是 |
| 0 | `OTHER` | DROP | 已定义，但今天没有任何分支会发出它 |
| 10 | `TCP_SSH_OTHER` | DROP | 是 |
| 11 | `TCP_NHP_PORT` | DROP | 是 |
| 12 | `TCP_OTHER` | DROP | 是 |
| 13 | `UDP_OTHER` | DROP | 是 |
| 14 | `UDP_SHORT` | DROP | 是 |
| 15 | `NON_TCP_UDP` | DROP | 是 |
| 20 | `V6_OTHER` | DROP | 是 |
| 21 | `TCP_SYN_SCAN` | DROP | 是 |
| 22 | `TCP_STEALTH_SCAN` | DROP | 是 |

那些"只计数"的分类是那些随*传输字节数*而不是随*事件数*增长的类——一条连接的第 4 个数据段、一次 DNS 查询的应答。它们照样被计数，只是每个都写一行日志不值当。一次 `dnf update` 或者一次 CI 用 `scp` 传发布包，写出来的日志比一个月的敲门还多。

## 5. nhp-ac 的按次放行过滤器

AC 的职责不同：它才是那个在敲门被授权时真正*打开*受保护端口、并在敲门过期后重新关上的组件。所以它的过滤器是一张带生命周期的、按次敲门生效的白名单，而且需要两个程序才能工作。

### 入站程序（`nhp_ebpf_xdp.c`，`SEC("xdp")`）

```
frame
 │
 ├─ ethertype
 │    ├─ ARP / IPv6              → PASS   (the AC does not filter IPv6 at all)
 │    ├─ IPv4                   → continue
 │    └─ otherwise              → DROP
 │
 ├─ iph->ihl < 5                → DROP   (not recorded)
 │
 ├─ TCP to dport 22             → PASS   (SSH to the AC is never knock-gated)
 ├─ UDP 67 → 68 (DHCP)          → PASS   (the lease reply, by port pair, before any whitelist or conn_track lookup — the only UDP exception)
 │
 ├─ ICMP
 │    ├─ echo reply (type 0 code 0)   → PASS
 │    ├─ echo request (type 8 code 0)
 │    │    ├─ no icmpwhitelist entry → DROP
 │    │    ├─ entry expired           → DROP   (the entry is deleted)
 │    │    ├─ allowed == 1            → PASS
 │    │    └─ allowed != 1            → falls through to the lookups below
 │    └─ any other ICMP                → falls through to the lookups below
 │
 ├─ conn_track has an INGRESS entry for this tuple and it is not a pure SYN?
 │    └─ live window             → PASS   (a flow a whitelist admitted earlier)
 │
 ├─ conn_track has an EGRESS entry (the reply side of a flow the AC opened)?
 │    └─ live window             → PASS   (see "the reverse branch" below)
 │
 ├─ any of the five whitelists admits it?
 │    ├─ spp          (src + dst + port + protocol)
 │    ├─ sdwhitelist  (src + dst, for an "any"-protocol resource)
 │    ├─ src_port     (src + port)
 │    ├─ port_list    (src + a wildcard port range)
 │    └─ protocol_port(port + protocol)
 │         └─ yes, and not expired → PASS   (and the flow is recorded)
 │
 └─ otherwise                   → DROP
```

DHCP 是这里唯一的 UDP 例外，并且是按 **端口对**（67 → 68）匹配的。源端口由发送方自己决定，证明不了任何事；这个分支以前还会放行任意来源的 `dport == 67 || dport == 68`，以及——更糟的——任何 `sport == 53` 的报文，而且都发生在 conn_track 和全部白名单**之前**。后者意味着任何从 udp/53 发出的报文都能打到 AC 保护的任意 UDP 端口（DNS、QUIC/HTTP3、WireGuard、游戏服务等），完全不需要敲门。

DNS 不需要专门的例外：AC 的查询从 `tc_egress.c` 所挂的同一个网卡发出，应答由下面的反向分支凭它留下的 EGRESS 条目放行，和 NTP 一样。这对**源端口**被 gate 3 接受的查询成立——也就是 ephemeral 端口，stub 解析器（glibc、systemd-resolved）用的就是它。**跑在 AC 上**、并且把出站端口随机化到 1024–65535 的递归解析器（BIND 的 `use-v4-udp-ports` 默认值）则落在和 AC 上其它固定低端口 UDP 客户端同一条限制里：见 `terraform/demo/RUNBOOK.md` 的 *Known limitation: fixed low UDP source ports*。

每个白名单条目都带一个过期时间，而过期的条目会在某个报文发现它时**被删除**——于是一个已经跑完的条目立刻不再放行流量。

这五次查找按固定顺序进行，并在第一次命中时停下，这带来一个看得见的后果：前面某张表里的一个**过期**条目会把自己删掉，就地把这个包丢掉，后面的表根本不会被查。因此对同一个包来说，一条过期的 `spp` 条目会盖过一条仍然有效的 `protocol_port` 条目。

反向分支是最微妙的一处，所以上面的判定树只是它的概括，而不是规则本身。有三个条件限定了它：

- 对于**持有有效敲门**的对端，这个分支会被**跳过**，除非该条目带有 `CT_FLAG_AC_CLIENT`——也就是说，除非这个流是 AC 自己发起的。被敲门授权的对端必须由下面那些白名单来放行，绝不能从反向查找里得到服务。
- 被 `knock_peers` 记录挡住的条目会被**删除**，而不是被服务。这样的记录永不失效，所以它挡住的条目不能留在原地：出站程序不检查过期，AC 自己的发送会不断复活那个条目，而任何入站报文都无权触达删除它的那段代码。
- AC 已经认领（claim）的、**已过期**的条目会被**保留**而不是删除。这个认领是 AC 对该元组的长期持有权；把它丢掉，会让下一个被放行的报文写入那条 `knock_peers` 记录，而那条记录随后会把出站闸门关在 AC 自己的流上。

同样的限定也适用于"一个被放行的报文会记录什么"：对于 AC 自己拥有的流，什么都不记录——既不写 `knock_peers`，也不写 `conn_track`。

这五种白名单形态由用户态写入（`nhp/utils/ebpf/ebpf.go`），每一种对应协议支持的一种敲门形态。`spp` 的键是来源、服务、端口和协议：

```go
func (r *whitelistKey) ToWlKey() []byte {
	keyBytes := make([]byte, 11)
	binary.LittleEndian.PutUint32(keyBytes[0:4], r.SrcIP)
	binary.LittleEndian.PutUint32(keyBytes[4:8], r.DstIP)
	binary.BigEndian.PutUint16(keyBytes[8:10], r.DstPort)
	keyBytes[10] = r.Protocol
	return keyBytes
}
```

值里带着同一个过期时间：

```go
func (r *whitelistKey) ToWlValue(ttlSec uint64) whitelistValue {
	now, _ := getBootTimeNanos()
	return whitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}
}
```

### 反向分支与 `knock_peers`

AC 同样是客户端——它要跟 nhp-server 说话、要访问解析器和软件包镜像。这些流的应答必须能回来，而把它们和"被敲门授权的流"区分开来的唯一办法，是知道哪些流是**AC 自己**发起的。

这就是 `conn_track` 的用途，而 TC 出站程序负责写入 AC 那一半。但光有 `conn_track` 还不够，因为 AC 自己的流于是也能被对端刷新：对端自己挑源端口，所以一个只按裸 4 元组刷新的窗口，会让一个只敲过一次门的对端把该元组一直吊住、并在其上开新的连接。

所以入站程序会为每个被白名单放行的**非 TCP** 报文，在 `knock_peers` 里记下一条持久的、不过期的记录：

```c
/*
 * Peers a knock has admitted, and to which port on the AC. Written by
 * xdp_white_prog() when a whitelist admits a non-TCP packet, read by
 * tc_egress.c (gate 4) and by xdp_white_prog()'s reverse branch.
 *
 * Unlike the whitelists this is *not* a permission: nothing is ever admitted
 * because of an entry here. No expiry is enforced on it, and it is deliberately
 * never deleted -- an entry that aged out would let the hole it closes reopen,
 * since the AC answering a peer long after the knock ran out is exactly the
 * case gate 4 guards against. The map is an LRU so it stays bounded regardless,
 * and nhp-acd drops the pin at startup.
 */
```

`knock_peers` 这条记录，正是让出站闸门拒绝把"受敲门保护的流"认作自己流量所需要的东西。没有它，敲门过期后的第一个数据报就会把一个出站条目和一个可以无限刷新的回头路交给对端。

### 出站程序（`tc_egress.c`，`SEC("tc/egress")`）

出站程序回答的是另一个问题：*这个报文是 AC 在说话，还是 AC 在应答？* 它有四道闸门：

- **闸门 1 —— 命中入站项。** 反向 4 元组已经以 INGRESS 方向存在于 `conn_track` 中，说明是这个对端发起的流。什么都不创建也不刷新；这条流的寿命只属于敲门 TTL。唯一被记录的是 AC 自己发出的 FIN，这样一个伪造的入站 FIN 就不能把一条已被准入的会话永久切断。
- **闸门 2 —— TCP 只有纯 SYN 才创建条目。** 监听 socket 永远不会发出纯 SYN，所以纯 SYN 本身就证明这里是 AC 作为客户端。该流被打上 `CT_FLAG_AC_CLIENT` 标记。
- **闸门 3 —— UDP：源端口必须看起来像客户端。** AC 作为客户端打开的 socket 会绑定一个临时端口，而它监听的服务不会。这个下界取自主机自己的 `net.ipv4.ip_local_port_range`，在加载时读出并写进 `nhp_config` map——写死 Linux 默认值会让任何调低了该范围的主机丢掉 AC 自己的应答。有三个固定端口是例外，因为它们的应答会回到同一个固定端口：NTP（123）、IKE（500）和 IKE NAT-T（4500）。
- **闸门 4 —— UDP：排除那些持有或曾经持有过敲门的对端。** `knock_peers` 加上按地址索引的白名单一起回答这个问题。这道闸门保证受敲门保护的服务不会被 AC 自己的流认领。

那个认领标记就是把两者分开的钥匙。一条在敲门出现**之前**就由 AC 打开的流永远不会被压制，而且它的端口永远不会被写进 `knock_peers`——否则一次来自"AC 也在通信的那个地址"的敲门（nhp-server 自己的地址、某个 NTP 服务器的地址）就会在守护进程的整个生命周期里关掉 AC 自己控制通道上的闸门 4。

### AC 的各个 map

AC 的所有 map 都用 `LIBBPF_PIN_BY_NAME` 做 pin，入站程序和出站程序正是靠这个按名字共享同一张内核 map。

| Map | 类型 | 键 |
|-----|------|-----|
| `spp` | `LRU_HASH` | 源 IP + 目的 IP + 目的端口 + 协议 |
| `sdwhitelist` | `LRU_HASH` | 源 IP + 目的 IP |
| `src_port` | `LRU_HASH` | 源 IP + 目的端口 |
| `port_list` | `LRU_HASH` | 源 IP + 端口范围 |
| `protocol_port` | `LRU_HASH` | 目的端口 + 协议 |
| `icmpwhitelist` | `HASH` | 源 IP + 目的 IP |
| `conn_track` | `LRU_HASH` | 4 元组 + 方向 + 标志 |
| `knock_peers` | `LRU_HASH` | 对端 IP + 服务端口 + 协议 |
| `nhp_config` | `ARRAY` | 一个槽位：临时端口下界 |
| `events` | `PERF_EVENT_ARRAY` | 用户态抽取的 perf ring（未 pin） |

这两条规则管辖着这些 map，都很重要：

- **两个 eBPF 程序都不许写白名单。** 白名单由用户态在敲门被授权时写入，过期时间取自资源的 `OpenTime`；程序只读。

  > Neither eBPF program may write them -- tc_egress.c used to write `spp` for
  > every outgoing packet, which kept the door open long past OpenTime.

- **`knock_peers` 永远不会被删除**，两个程序都不会删。它是 LRU，所以有界；但一个老化的条目会让它堵住的那个洞重新打开。

AC 的 perf 事件带一个二值 action——0 = `DENY`，1 = `ACCEPT`——并写入两个独立的日志：`nhp_deny` 和 `nhp_accept`。server 只写一个文件 `nhp_server_xdp`，因为它的 action 说的是报文**为什么**被放行或丢弃，按判定结果拆开会把这些信息丢掉。

### 出站程序从不做的一件事

`tc_egress.c` 里每一个 `return` 都是 `TC_ACT_OK`。出站程序**从不丢包**——它只做记录。AC 的入站程序才是决定访问权的那一个；出站程序存在的意义，是让 AC 自己发出的流能把应答收回来，并且记录它发过什么。

这也是闸门 2 到闸门 4 讲的是**认领**一个流、而不是拒绝一个流的原因：认领错了只是丢掉一次认领，拒绝错了要赔上整个协议。

## 6. 加载器：从 `.o` 到一个真正在跑的过滤器

`nhp/utils/ebpf/engine_linux.go` 是这一层里行为最出人意料的部分，所以值得按顺序走一遍。这个顺序不是偶然的——它就是契约。

```
EngineLoad(params)
 │
 ├─ record the pin directory, sweep this variant's leftover pins
 ├─ rlimit.RemoveMemlock()
 │
 └─ loadServerEngine(params, pinDir)
      │
      ├─ os.Stat(.o)                     missing → error, nothing attached
      ├─ ebpf.LoadCollectionSpec(path)    bad object → error, nothing attached
      ├─ setServerConstants(spec, params) ← .rodata must be rewritten HERE
      ├─ spec.LoadAndAssign(&objs, ...)   creates the maps, pins them
      │      └─ failure → error, nothing attached
      ├─ check nhp_events / nhp_relay_ips are non-nil
      ├─ pin xdp_server_prog
      ├─ ReplaceRelayIPs(...)             ← the whitelist goes in HERE
      ├─ resolveInterface() + warnOnGlobalIPv6()
      ├─ perf.NewReader(nhp_events)       failure is logged, not fatal
      ├─ link.AttachXDP(..., XDPGenericMode)
      │      └─ failure → error; deferred CleanupBPFFiles detaches and unpins
      └─ start the reader / stats / scan / prune goroutines
             │
             └─ return the EngineHandle
```

这个顺序里有四处是承重的。

### 常量在目标文件被加载之前改写

敲门端口和长度下限是 `.rodata` 常量而不是 `#define`，因为两者都是用户态配置：

```c
volatile const __u16 nhp_listen_port = 62206;
volatile const __u16 nhp_min_udp_len = 240;
```

`volatile const` 的含义是"对验证器来说是常量，对编译器来说是变量"——改写之后验证器仍然会把死分支折叠掉，但编译器无法把初始值烤进指令流。加载器在验证器看到它们之前改写它们：

```go
	// Rewrite the .rodata constants before LoadAndAssign: after the collection
	// is loaded the verifier has already folded them, so this is the only
	// moment the knock port and the length floor can still be set.
	if constErr := setServerConstants(spec, params); constErr != nil {
		return nil, constErr
	}
```

一个早于这些常量存在的目标文件会被**拒绝**，而不是带着自己的默认值挂上去：

```go
		v, ok := spec.Variables[c.name]
		if !ok {
			log.Error("server eBPF object has no '%s' constant — it is older than the loader; rebuild it with `make ebpf-objs`", c.name)
			return fmt.Errorf("server eBPF object has no %q constant", c.name)
		}
```

把一个敲门端口写死成 62206 的目标文件挂到一台监听在别的端口的 server 上，会让每一个敲门都在驱动层被丢掉。这看上去和 server 宕机一模一样，而且在用户态完全看不出来——比干脆不挂要糟糕得多，而不挂本来就是守护进程已经能处理的状态。

### 白名单在程序挂上去之前写进去

```go
	// The whitelist goes in before the program goes on. Everything the filter
	// admits that is not a knock arrives on tcp/22 from one of these prefixes,
	// so a program attached with an empty trie is a host with no way in — and
	// if this write is what failed, user space has just proved it cannot fix
	// the map it would have to fix.
```

白名单是**加载参数**，而不是事后写进去的，正是为了保证这个顺序。一个挂着空 trie 的程序会把所有来源的 tcp/22 都丢掉：在写入落地前的几微秒里无所谓，如果那次写入随后失败就无可挽回——调用方会面对一个已经挂上的过滤器、没有 SSH，以及一张它刚才就写不进去、无从修正的 map。作为参数传入，失败发生在挂载**之前**，会像其他所有挂载前的错误一样回退到 fail-open。

### 所有可能失败的事情，都在挂载之前失败

那个延迟清理的存在，是因为加载之后的失败已经在主机上留下了东西：

```go
	// From here on a failure has left state on the host: LoadAndAssign has
	// already created the map pins under bpffs, and past the attach below there
	// is a live filter. Unwind both, or the daemon runs "fail-open" while the
	// driver is still dropping everything — and with no handle, nothing in user
	// space could fix the whitelist it is enforcing.
	defer func() {
		if retErr != nil {
			CleanupBPFFiles(VariantServer)
		}
	}()
```

调用方只会从 `EngineLoad` 拿到一个错误，而且永远拿不到 `EngineHandle`。一个在返回错误之后仍然挂着的过滤器，会用一份用户态已经没有 handle 可修的空白名单在执行规则。在一台唯一入口就是那份白名单的主机上，这不可挽回。

那个 defer 是在 `LoadAndAssign` **成功之后**才注册的，所以在这一步失败的加载没有清理可做——它留下的 pin 由下一次 `EngineLoad` 扫掉，后者在做任何别的事情之前先移除本变体遗留的 pin。

### 丢掉事件日志不算致命

```go
	// Opened here rather than inside the reader goroutine because
	// perf_event_open(2) is the last thing that needs CAP_PERFMON, and the
	// capabilities are given back when this function returns. Not fatal if it
	// fails: the event log is diagnostics, and losing it is no reason to leave
	// the host unfiltered. The filter runs on without it, saying so.
```

这个区分值得记住：**判定**才是过滤器，**事件日志**只是诊断。没有事件日志的主机依然被过滤；没有过滤器的主机则根本没有被过滤。

### 能力被还回去

`nhp-serverd` 被授予 `CAP_BPF`、`CAP_NET_ADMIN` 和 `CAP_PERFMON` 三个**环境（ambient）**能力。启动时大约需要它们一秒钟，之后一个都不需要：

- 写 map 只需要进程已经持有的 map fd；
- 读 perf ring 是对 fd 的 `read(2)`；
- 卸载是对 link fd 的 `close(2)`，而删除一个 bpffs pin 是对服务用户本来就拥有的目录做 `unlink(2)`。

把它们留到进程生命周期结束，是一个比看上去大得多的授权。`CAP_BPF` 配合 `CAP_PERFMON` 能加载 kprobe 和 tracing 程序，而它们可以调用 `bpf_probe_read_kernel()`——任意内核内存，包括别的进程的凭证以及这个守护进程自己的 Noise 私钥。`CAP_NET_ADMIN` 能卸载这个过滤器本身、改写路由、改写 netfilter 规则。而且这是一个解析来自整个互联网的不可信 UDP、并把第三方认证插件 `dlopen` 进自己地址空间的进程。

所以加载器把它们的归还作为最后一件事，在函数的**每一条**出口路径上：

```go
	defer func() { DropLoaderPrivileges(retErr == nil) }()
```

注意这个参数。`retErr == nil` 意味着加载产出了一个活着的过滤器，只有这时 `CAP_BPF` 才值得保留——因为 `xdp.toml` 的重载路径要通过它写 relay 白名单 map。失败的加载一个能力都不留：

> That is the case the daemon runs in for the longest — a load that fails is fail-open for the life of the process — so it is the one that least deserves a capability.

有一个例外，而且它是内核属性而不是策略选择。在 `kernel.unprivileged_bpf_disabled != 0` 的内核上——Amazon Linux 2023、Ubuntu 和 Debian 的默认值——`bpf(2)` 是**整体**特权化的，包括对进程已持有的 fd 的 map 更新。在那里丢掉 `CAP_BPF` 会让之后每一次 `ReplaceRelayIPs` 都变成 `EPERM`，而重载路径是在一台唯一入口就是那份白名单的主机上，远程修正白名单的唯一途径。所以在这样的内核上 `CAP_BPF` 被保留，并且日志会说明原因：

```go
	if keptBpf {
		log.Info("XDP loader privileges dropped: CAP_NET_ADMIN and CAP_PERFMON released, CAP_BPF kept because %s is non-zero (bpf(2) is privileged on this kernel, and the xdp.toml reload path writes the relay whitelist map). The daemon now holds %q",
			unprivilegedBpfSysctl, want)
```

### 这次"扣除"是一次减法，不是一次赋值

```go
// loaderCaps are the three capabilities the unit grants for the load, and the
// only three this file ever takes away. Anything else the process happens to
// hold was granted for some other reason and is not the loader's to give back.
var loaderCaps = []cap.Value{cap.BPF, cap.NET_ADMIN, cap.PERFMON}
```

加载器从环境、可许、可生效、可继承这四个集合里减去这三个能力。它不赋值一个固定集合，也不重置成空集。在演示主机上这两者是同一件事——授予的正好就是这三个——但 `nhp-serverd` 还会以 root 身份作为容器发布（`docker/Dockerfile.server`），在那里进程持有完整的 root 集合，理由与这个加载器毫无关系：卷挂载的配置文件属于别的 uid，需要 `CAP_DAC_OVERRIDE`；`http.toml` 重载要重新绑定 80/443，需要 `CAP_NET_BIND_SERVICE`。在那里重置成空集——在每一条路径上，包括一台根本没要求过这些的主机——是一次不可逆的、静默的、而且完全无关的权限变更。

同样的道理适用于那些根本没走到加载器的路径。一台没有 `xdp.toml`、有一个解析不了的 `xdp.toml`、`Enabled = false`、白名单撑不进 trie、或者 `NhpMinFrameBytes` 越界的主机，都不会调用 `EngineLoad`——但 unit 无论如何都以**环境能力**的形式授予这三个能力，所以这些状态过去都意味着一个守护进程在它整个生命周期里一直持有 `CAP_BPF`、`CAP_NET_ADMIN` 和 `CAP_PERFMON`，却一次也没用过。因此这次扣除被延迟到配置路径的每一个 return 上：

```go
	attached := false
	defer func() { dropXdpLoaderPrivileges(attached) }()
```

### 替换白名单

`ReplaceRelayIPs` 是重载路径上唯一一次 map 写入，它的写法是针对 **map** 而不是仅针对文件做到全有或全无：

- **解析是全有或全无的。** 一条不可用的条目会让整次调用失败。跳过坏条目看起来更宽容，但对一个拼写错误来说那才是危险的读法：一张只有一个地址拼错的列表会解析出零条前缀，而一张只有"SSH 真正来自的那一条"拼错的列表，会解析出一张把运维人员排除在外的白名单。两者在 map 层面与一份刻意写成的列表无法区分，而在一台唯一入口就是这份白名单的主机上，这个错误无法远程修正。
- **空列表被拒绝。** 把 map 清空会关掉所有来源的 tcp/22——在一台唯一入口就是这个 map 的主机上。
- **新条目先进，旧条目后出。** 一张新旧列表都有的地址永远不会短暂缺席——先删后插会开出一个窗口，让 relay 自己的 SSH 会话和转发的敲门被这次本该让它们继续工作的重载挡掉。
- **一次失败的写入会回滚。** 如果某个 `Update` 失败，这次调用加进去的键会被取回，清理陈旧条目的扫描也不会执行，于是 map 保持的正是它之前持有的那份列表。
- **一张读不出来的 map 干脆一个字节都不写。** 这个方向才能让一台可达的主机继续可达。

白名单的解析放在 `nhp/utils/ebpf/relayips.go` 而不是加载器里，这样 server 的配置路径才能在决定挂载**之前**校验一份渲染好的 `xdp.toml`——而且这份校验可以在所有平台上编译和测试，而不只是加载器所在的那个平台。

## 7. 配置：`etc/xdp.toml` 到底控制什么

这个文件是 `etc/xdp.toml`，放在守护进程二进制旁边（或者 `XdpConfigPath` 指向的任意位置）。**是它的存在本身让一台主机选择被过滤。** 完整模板（每个字段都带注释）在 `deploy/config-templates/server/xdp.toml`，大致长这样：

```toml
Enabled = true
NhpMinFrameBytes = 240
RelayIPs = ["10.0.1.0/24", "203.0.113.7"]
ScanReportIntervalSeconds = 300
ScanReportTopN = 20
EventLogEnabled = true
```

这里真正重要的是：它的哪些字段是**被应用**的，哪些只是**被记录**的，哪些可以热重载。

| 字段 | 何时生效 | 说明 |
|------|----------|------|
| `Enabled` | 仅启动时 | `false`、文件缺失、或缺少 `RelayIPs` 键 ⇒ 什么都不挂。 |
| `RelayIPs` | **热重载** | 唯一可以热重载的字段。文件一变，活的 map 就跟着变。 |
| `NhpMinFrameBytes` | 仅启动时 | 在加载时写进目标文件的 `.rodata`。是被应用的，不是被记录的。 |
| `ScanReportIntervalSeconds` | 仅启动时 | 只影响 `[NHP-SCAN]` 摘要。 |
| `ScanReportTopN` | 仅启动时 | 只影响 `[NHP-SCAN]` 摘要。 |
| `EventLogEnabled` | 仅启动时 | `false` 只停掉逐包日志；摘要和过滤器都不受影响。重载时改这一项会被**静默**忽略——这一项没有警告。 |

敲门端口**不在**这个文件里。守护进程从自己的 `ListenPort` 取它，加载器把它和长度下限一起写进 `nhp_listen_port`。这是刻意的：过滤器和监听 socket 不可能各说各话。

重载路径是一个文件监听器，而且刻意做得很窄：

```go
	xdpConfigWatch = utils.WatchFile(fileName, func() {
		log.Info("xdp config: %s has been updated", fileName)
		content, err := s.loadConfigFile(fileName)
		if err != nil {
			log.Error("failed to reread xdp config, keeping the active whitelist: %v", err)
			return
		}
		var xdpConf XdpTomlConfig
		if err := toml.Unmarshal(content, &xdpConf); err != nil {
			log.Error("failed to unmarshal xdp config, keeping the active whitelist: %v", err)
			return
		}
		s.applyXdpConfig(&xdpConf)
	})
```

如果文件在**重载**时缺失或解析不了，什么都不会被应用，内核保持它已有的白名单。用一份被截断或写了一半的文件内容覆盖一份能用的白名单，正是把运维人员锁在 SSH 门外的那个故障，所以一次重载只有在 TOML 解析成功之后才会生效。

而 `applyXdpConfig` 对它无法应用的字段是明说的：

```go
	if conf.NhpMinFrameBytes != 0 && conf.NhpMinFrameBytes != s.xdpMinFrameBytes() {
		// The floor is written into the object at load time, so an edit after
		// that is not enforced until the next restart.
		log.Warning("xdp config: NhpMinFrameBytes=%d is not applied — the attached program is enforcing %d bytes; restart nhp-serverd to change it",
			conf.NhpMinFrameBytes, s.xdpMinFrameBytes())
	}
```

对 `Enabled = false`（一个活着的过滤器不能靠改文件卸载）和两个 `[NHP-SCAN]` 开关，日志里会给出同样的警告。

### 遥测参数越界不会让过滤器停摆

那两个 `[NHP-SCAN]` 设置只影响遥测，所以超出范围的值绝不能决定过滤器是否挂载。在安全组对所有端口开放的情况下，过滤器是唯一的入站控制，在这里拒绝就等于把主机完全暴露：

```go
	if seconds < minScanReportIntervalSeconds || seconds > maxScanReportIntervalSeconds {
		log.Error("xdp config: ScanReportIntervalSeconds=%d is out of range [%d, %d]; the [NHP-SCAN] summary is disabled, the XDP ingress filter still attaches",
			seconds, minScanReportIntervalSeconds, maxScanReportIntervalSeconds)
		return 0, DefaultScanReportTopN
	}
```

白名单的校验则不同：它是第 8 章列出的那些拒绝之一——一张 trie 装不下的白名单在挂载**之前**就被拒绝，而不是挂上之后。

{: .warning }
在一台唯一入口就是白名单的主机上，一份写错的 `RelayIPs` 不是一个可以挽回的错误。列表里没有可用前缀时守护进程会拒绝挂载，重载时遇到坏的列表会保留上一份——但一份能解析、却指向了错误前缀的列表会被应用，而做这次修改的 SSH 会话本身也会被它挡在门外。这是刻意不设 break-glass 的。恢复流程写在演示运维手册里（`terraform/demo/RUNBOOK.md`，"nhp-server ingress filter (XDP)"）；在一般情况下，恢复意味着从带外触达这台主机。

### fail-open 与 fail-closed

两个守护进程对"加载失败"的回答正好相反，而这个差异是刻意的：

| | nhp-server（入站过滤器） | nhp-ac（按次放行过滤器） |
|---|---|---|
| 加载失败 | **Fail-open** —— 主机保持过滤器出现之前的暴露面 | **Fail-closed** —— 守护进程拒绝启动 |
| 原因 | 访问控制由协议负责，过滤器只是纵深防御。一个兜底必须用手工方式重新实现白名单、敲门端口和 DHCP 例外，而一个写错的兜底会把主机锁死——包括那个白名单自己的入口，也就是 SSH | 过滤器*就是*访问控制：只要 AC 没有被证明正在执行，受保护端口就必须保持关闭 |
| 兜底 | 无 | `deploy/scripts/nhp-ac-backstop.sh` —— 一条 netfilter 链，在守护进程没有执行时把受保护端口 DROP 掉，并且只在确认 XDP 程序已经挂上之后才撤掉 |
| 重启/掉电窗口 | 过滤器没挂上的那段时间里主机是暴露的 | 受保护端口保持关闭 |

AC 需要这个兜底，是因为 XDP 的 link 只存在于守护进程的内存里——`link.AttachXDP` 把返回的 link 存在一个包级变量里，所以 link 的生命周期等于进程的生命周期。被 pin 住的只有*程序*，那让目标文件活着，却什么都没挂。于是执行在守护进程消失的瞬间就没了：一次 `systemctl stop`、一次崩溃加上 `Restart=on-failure` 退避、或者一次失败的部署，都会把受保护端口暴露给全世界。这条 netfilter 链从守护进程之外恢复了这个性质。

server 没有这样的兜底，而这种不对称正是两种策略的代价。server 的过滤器并不是"让它的受保护服务隐身"的东西——协议才是；过滤器只是针对一台反正会被扫描的主机的纵深防御。AC 的过滤器**就是**访问控制。

### 看门狗

server 还做了一件值得知道的事，因为它是"一次失误"和"一次重建"之间的区别：

```go
// watchHostAddressing is the dead man's switch for the ingress filter.
//
// The filter is written entirely in terms of the host's IPv4 configuration,
// and there is no break-glass path into the demo host: if something the host
// needs is dropped and the effect is that it loses that configuration, nobody
// can get in to fix it.
```

如果被过滤的网卡连续两分钟（4 次 30 秒）没有可路由的 IPv4 地址，看门狗会打一条 `Critical` 日志并卸载过滤器，好让下一次 DHCP 交换能把这台主机救回来。`169.254.0.0/16` 这个自动配置地址不算数——那正是一台丢了 DHCP 租约的主机最后会拿到的地址。过滤器会一直保持关闭，直到守护进程被重启；自动恢复只会把同一个陷阱重新武装一遍。

## 8. 可观测性与排障

过滤器会写出三类日志行，全部写进 `logs/nhp_server_xdp-<date>.log`。

### 逐包日志行

```
15:04:05 srv1 [NHP-PASS] REASON=NHP_DEFAULT SRC=203.0.113.9 DST=10.0.1.5 LEN=240 PROTO=UDP SPT=41337 DPT=62206 RELAY=0
15:04:05 srv1 [NHP-DROP] REASON=TCP_SYN_SCAN SRC=198.51.100.4 DST=10.0.1.5 LEN=60 PROTO=TCP SPT=51234 DPT=22 RELAY=0
```

`REASON=` 就是第 4 章那张表里的 action code。`RELAY=1` 表示来源命中了白名单。

### 每分钟总数：`[NHP-STAT]`

```
15:05:00 srv1 [NHP-STAT] VERDICT=DROP REASON=TCP_SYN_SCAN WINDOW=60s PKTS=1432 LOGGED=16 TOTAL=1432
```

`PKTS` 是这个窗口内的增量，`LOGGED` 是其中真正进入事件流的数量，`TOTAL` 是过滤器挂上以来的累计。计数器对**每一个**报文都会自增，不管有没有预算，所以不管丢了多少记录，总数都始终精确——这正是它们存在的理由。

### 扫描聚合：`[NHP-SCAN]`

```
15:05:00 srv1 [NHP-SCAN] WINDOW=300s KIND=PORT PROTO=TCP DPT=22 PKTS=981
15:05:00 srv1 [NHP-SCAN] WINDOW=300s KIND=SRC SRC=198.51.100.4 PKTS=704
15:05:00 srv1 [NHP-SCAN] WINDOW=300s KIND=TOTAL UNIQ_SRC=31 UNIQ_PORT=4 PKTS=1902 LRU=0
```

它以比判定分类更细的粒度回答"是什么在打这台主机、从哪里来"。`LRU=1` 表示表已经超过 95% 满、很可能发生了淘汰，因此 Top-N 列表是近似的——精确的分类总数仍然来自 `[NHP-STAT]`。

### 日志为什么是有界的

程序看到的每一个报文都是发包的人选的，而读取端会把每一个事件追加到一个按日期轮转、却从不封顶的日志文件里。没有限制的话，任何能到达主机的人都可以靠扫描把它的磁盘写满——而那会打垮过滤器本该保护的那个守护进程。有三道界限同时生效，而且没有一道会改变判定：

- **内核里的每分类令牌桶。** 每个 action 分类每 CPU 每秒 16 个突发额度，另有一个跨所有分类的全局桶，每 CPU 每秒 8 个。每分类的桶防止某一类洪水把另一类的记录饿死；全局的那个则给总量封上界，无论它是由什么分类混合而成的。没有它，一个 IPv4 发送方能触及的那十来个可上报分类，每个十六个突发额度，再乘以 CPU 数，就是每秒几百行——也就是每天几个 GB。
- **写入侧每天的字节预算。** 每个日历日 64 MiB 的逐包文本。花完之后写入端不再增加行并说明原因——但 `[NHP-STAT]` 摘要仍然继续报出精确总数。
- **保留期清理。** 14 天或 256 MiB，哪个先到算哪个；先删最老的文件，当天正在写的文件永不删除。

那些批量分类干脆完全不记录（`TCP_ESTABLISHED`、`UDP_ESTABLISHED`、`V6_ESTABLISHED`、`ICMPV6_CONTROL`，以及 SYN 之后的每一个 `SSH_RELAY` 报文）。它们随传输字节数增长，而不是随事件数增长。

### 过滤器没挂上的时候

每一条让主机保持未过滤的路径都会打一行日志。`Warning` 意味着加载器失败；其余的都是配置路径**有意决定**不挂载——文件缺失或被禁用是 `Info`，数值越界是 `Error`，而一份会关掉 tcp/22 的白名单是 `Critical`：

```
INFO   no etc/xdp.toml: the XDP ingress filter is not attached (...)
INFO   xdp config: etc/xdp.toml has Enabled = false; the XDP ingress filter is not attached
ERROR  failed to unmarshal xdp config, the XDP ingress filter is not attached: ...
ERROR  xdp config: NhpMinFrameBytes=70000 is out of range; refusing to attach the XDP ingress filter
ERROR  xdp config: cannot filter for UDP listen port 62206; refusing to attach the XDP ingress filter
CRIT   xdp config: etc/xdp.toml has an unusable RelayIPs entry (...) — refusing to attach the XDP ingress filter rather than closing tcp/22 for every source
CRIT   xdp config: etc/xdp.toml lists no RelayIPs — refusing to attach the XDP ingress filter rather than closing tcp/22 for every source. Fix RELAY_IPS in the deploy pipeline
WARN   server eBPF engine load failed, fail-open (no XDP ingress filter): ...
```

那条 `Warning` 是加载器的失败：缺少 `nhp_server_xdp.o`、目标文件老到不带 `nhp_listen_port`、内核老到不支持 XDP、bpffs 没挂载，或者缺少能力。目标文件应该在二进制旁边，路径是 `<exe dir>/etc/nhp_server_xdp.o`，由 `make ebpf-objs` 产出。

还有一条拒绝值得知道，因为它关于的是**这台主机**而不只是过滤器：

```
CRIT   xdp config: refusing to attach the XDP ingress filter because the HTTP knock listener is enabled on 0.0.0.0:443 — the filter drops inbound TCP other than SSH from RelayIPs, so that service would be unreachable. Disable it, bind it to 127.0.0.1, or set Enabled = false in etc/xdp.toml
```

同样的拒绝也覆盖绑定在非环回地址上的 metrics 端点。过滤器会丢弃到主机上任何服务的入站 TCP，不只是到 `nhp-serverd` 的，所以当守护进程自己的某个监听端口会被黑洞掉时，它会拒绝挂载。绑定在环回地址上的不算冲突——程序挂在默认路由网卡上，发往 127.0.0.1 的流量根本到不了它。

### 离线分析日志

`endpoints/xdplog-report` 能把这些文件解析成一份报告：

```bash
xdplog-report -in logs/nhp_server_xdp-2026-10-01.log -out report.md
```

它的 `site` 子命令会发布一个静态摘要站点，而精确的报文总数取自 `[NHP-STAT]` 行，而不是逐包行——后者按设计就是一个被限速的采样。

## 延伸阅读

- [源代码解读](/zh-cn/code/) —— 仓库导览，包括 eBPF 源码在其中的位置。
- [加密算法](/zh-cn/cryptography/) —— 第 4 章那个长度下限所依据的密码套件。
- [协议参考](/zh-cn/protocol/) —— XDP 过滤器刻意不去解析的消息头。
- [功能列表](/zh-cn/features/) —— 这一层实现的那个默认拒绝性质。
- `terraform/demo/RUNBOOK.md` —— 运维手册："nhp-server ingress filter (XDP)" 与 "Why each rule is there"，是第 4 章和第 7 章的、由事故驱动的版本。
