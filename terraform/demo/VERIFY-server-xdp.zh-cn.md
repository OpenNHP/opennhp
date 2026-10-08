# nhp-server eBPF/XDP 入口过滤 —— 验证手册（部署前 / 部署后对比）

本文给出手工验证 `nhp-server` 入口 XDP 过滤的完整命令清单，分三段：

1. **本机验证**（§1）—— 代码上线前在一台 Linux 开发机上跑完，用 netns 隔离，不碰 demo；
2. **部署前基线**（§2）—— 在 demo 现网主机上采集「还没有 XDP」时的状态，存档待比对；
3. **部署后验证**（§3、§4）—— 用与 §2 **完全相同**的命令再跑一遍，逐项 diff。

涉及的实现：

| 组件 | 路径 |
| --- | --- |
| XDP 程序 | `nhp/ebpf/xdp/nhp_server_xdp.c` → `etc/nhp_server_xdp.o` |
| 装载/attach/perf reader | `nhp/utils/ebpf/engine_linux.go`（`Variant = VariantServer`） |
| server 侧封装 | `endpoints/server/ebpf/serverengine.go` |
| 配置与热更新 | `endpoints/server/config.go`（`etc/xdp.toml`，`WatchFile`） |
| 配置模板 | `deploy/config-templates/server/xdp.toml` |
| 流水线 | `.github/workflows/deploy-demo-v2.yml`（`configure` / `build` / `deploy-server`） |

> **过滤器是 opt-in 的。** 只有当 `etc/xdp.toml` 存在、能解析、`Enabled = true`
> 且 `RelayIPs` 的每一项都能解析成 IPv4 地址或 CIDR 前缀时，`nhp-serverd` 才会
> attach。注意判断的是「解析得出的前缀」而不是「字符串条数」：
> `["relay.opennhp.org"]`、`[""]`、纯 IPv6 列表、或者 `["10.0.1.300"]` 这种笔误
> 都是非空列表，却一个前缀也进不了 LPM trie——按条数放行就等于 attach 一张空
> 白名单，tcp/22 对所有来源关闭。只要有一项解析失败就整张列表作废（不会只跳过
> 坏的那项：解析失败的那条可能正是 SSH 实际来源）；热更新同理，此时保留正在
> 生效的白名单。此外，HTTP knock 监听或对外
> 暴露的 metrics 端点开着时它会拒绝 attach（过滤器会把这些监听端口的入站 TCP
> 全丢掉）。没有 `xdp.toml` 的主机行为与过滤器上线前完全一致。
>
> 本文示例中的端口全部以 `$NHP_PORT` 表示，请按本机实际端口替换（demo 当前
> `$NHP_PORT=443`，由 `terraform/demo/variables.tf` 的 `var.nhp_listen_port`
> 控制）。knock 端口来自守护进程自己的 `ListenPort`，长度下限来自
> `xdp.toml` 的 `NhpMinFrameBytes`，两者在 attach 前被写进对象的
> `.rodata`（`nhp_listen_port` / `nhp_min_udp_len`），不是编译期常量。
>
> ```sh
> NHP_PORT=443
> ```

决策树与 action 编码（`nhp_server_xdp.c`，日志里的 `REASON=`）：

| 条件 | 结果 | 日志 REASON | 逐包写日志？ |
| --- | --- | --- | --- |
| TCP/22，src ∈ 白名单 | PASS | `SSH_RELAY` | 仅 SYN（每会话一行） |
| TCP/22，src ∉ 白名单 | DROP | `TCP_SSH_OTHER` | 是（限速） |
| TCP/${NHP_PORT} | DROP | `TCP_NHP_PORT` | 是（限速） |
| TCP 其它端口，本机有**非 LISTEN** 的 socket | PASS | `TCP_ESTABLISHED` | **否，只计数** |
| TCP 其它端口 | DROP | `TCP_OTHER` | 是（限速） |
| UDP/${NHP_PORT}，src ∈ 白名单 | PASS（跳过长度校验） | `NHP_RELAY` | 是（限速） |
| UDP/${NHP_PORT}，长度 ≥ 240 | PASS | `NHP_DEFAULT` | 是（限速） |
| UDP/${NHP_PORT}，长度 < 240 | DROP | `UDP_SHORT` | 是（限速） |
| UDP 67 → 68（DHCP 续租应答） | PASS | `DHCP_CLIENT` | 是（限速） |
| UDP 123 → 123（NTP 应答） | PASS | `NTP_CLIENT` | **否，只计数** |
| UDP 其它端口，本机有 **connected** socket | PASS | `UDP_ESTABLISHED` | **否，只计数** |
| UDP 其它端口 | DROP | `UDP_OTHER` | 是（限速） |
| ICMP type 3 code 4（需要分片） | PASS | `ICMP_FRAG_NEEDED` | 是（限速） |
| 非 TCP/UDP（含其余 ICMP） | DROP | `NON_TCP_UDP` | 是（限速） |
| IPv4 非首片（`frag_off & IP_OFFSET`） | PASS | `IPV4_FRAGMENT` | 是（限速） |
| IPv6 ICMPv6 type 2 / 130–137（PTB、MLD、ND） | PASS | `ICMPV6_CONTROL` | **否，只计数** |
| IPv6 UDP 547 → 546（DHCPv6） | PASS | `DHCP_CLIENT` | 是（限速） |
| IPv6 UDP 123 → 123（NTP） | PASS | `NTP_CLIENT` | **否，只计数** |
| IPv6 TCP/UDP，本机有非 LISTEN / connected socket | PASS | `V6_ESTABLISHED` | **否，只计数** |
| IPv6 分片头 offset≠0（非首片） | PASS | `IPV6_FRAGMENT` | 是（限速） |
| 其余 IPv6（含 tcp/22、knock 端口、>2 层扩展头） | DROP | `V6_OTHER` | 是（限速） |
| ARP | PASS | 不上报（避免邻居流量刷屏） | 否，也不计数 |

**日志体量是有上界的**，这一列就是上界的来源（`record_packet()` 的 `report`
参数）。事件日志里每一行都对应一个「别人决定发过来」的包，所以无节制地写就等于
把磁盘写满的开关交给扫描者 —— 而磁盘写满会打死它本来要保护的那个守护进程。三
道闸：

1. **逐包但限速**：每个 action 一个 per-CPU 令牌桶，`NHP_EVENT_BURST`（16）
   条/秒/CPU/类。按类分桶，所以一轮端口扫描刷不掉同期 `SSH_RELAY` 的记录。
2. **批量类只计数**：`TCP_ESTABLISHED` / `UDP_ESTABLISHED` / SYN 之后的
   `SSH_RELAY` 是**本机流量的字节数**，不是事件数 —— 一次 `dnf update`、一次
   CI scp 就是几万个包，而它们说的是同一件事（这条流早就判过了）。这些包只进
   计数器。
3. **每分钟一行汇总**：`nhp_action_stats`（per-CPU array）对**每一个**包计数，
   与限速无关，用户态每 60s 把有变化的 action 各打一行
   `[NHP-STAT]`（`reportServerStats`）。所以「被限速丢掉多少」「静默类有多少」
   都是**准确数字**而不是估算。

再加上保留策略（`pruneServerEventLogs`，每小时扫一次）：`nhp_server_xdp-*.log`
超过 14 天或总量超过 256 MiB 就从最旧的删起，当天的文件永不删。`nhp/log` 只按
日期切分、从不清理，没有这一步的话主机活多久日志就攒多久。

`TCP_ESTABLISHED` / `UDP_ESTABLISHED` 两行是**本机自己发起的连接的回包**
（`has_local_flow()`，用 `bpf_sk_lookup_{tcp,udp}` 查内核 socket 表）。AC 用
TC egress + `conn_track` 做这件事，server 不需要：内核的 socket 表就是权威的
连接状态，socket 在就放行、socket 没了就不放行，没有 TTL、没有 map 会写错。
**LISTEN 状态的 socket 永远不算数**，未连接的 UDP socket 同理，所以「只有敲门
端口和 relay 的 SSH 对外可见」这个性质没有被放宽。

> 这一段是补 bug：第一版没有回包通路，`nhp-serverd` 自己发起的连接全部收不到
> 响应 —— 线上表现是 OTP 邮件发不出，日志里是 `lookup
> email-smtp.us-east-2.amazonaws.com on 10.0.0.2:53: i/o timeout`（DNS **应答**
> 被丢）。验证时 §1.4 的「回包矩阵」和 §4.5 的 DNS/SMTP 探测必须都过。

`DHCP_CLIENT` / `NTP_CLIENT` 两行是 socket 表**答不了**的那两类客户端，只能按
端口对放行：

> **这一段也是补 bug，而且是本特性造成过的最严重的一次故障。** DHCP 客户端
> （AL2023 = systemd-networkd）收租约用的是 raw `AF_PACKET` socket 或只
> `bind()` 到 udp/68 的 socket，两者在内核 socket 表里都不是「已连接」，
> `has_local_flow()` 必然拒绝 —— 于是 **67 → 68 的续租应答被 XDP 丢掉**。EC2 的
> 主私网 IPv4 是有租期的，networkd 在半个租期时续租，续不上就在租期到点时把地址
> （连同默认路由）从网卡上摘掉：**部署当时一切正常、约一个租期之后整台主机在所有
> 端口上同时失联** —— SSH、敲门端口、以及除重启之外的所有恢复通道。注意故障现象
> 不像「过滤器写错了」，因为敲门端口在地址消失前一直是好的。chronyd 的形状一样
> （应答回到它自己的 udp/123），代价是时钟而不是地址，所以一并放行。
>
> **这张放行名单就到此为止。** 「某个守护进程的回包被丢了」本身不构成加一条的
> 理由：把「凡是发给已 bind 端口的包都放进来」写进去，等于把整个过滤器的前提
> 扔掉。AC 端的程序一直有 DHCP 短路（`DHCP_PORT_R || DHCP_PORT_O`），是 server
> 这份 fork 把它删掉了。
>
> 回归防线有三层：`nhp/utils/ebpf/engine_linux_test.go` 的
> `TestServerFilterPassesClientRepliesAndDropsUnsolicited`（在 lo 上真发包，未
> connected 的 socket 收到 = 分支生效，普通端口收不到 = 没有放宽）、
> `deploy-server` 的 `Verify the host can still renew its DHCP lease`
> （`networkctl renew` 后必须出现新的 `REASON=DHCP_CLIENT`），以及
> `endpoints/server/ebpf/serverengine.go` 里的看门狗（网卡上连续两分钟没有可路由
> 的 IPv4 地址就打 `Critical` 并摘掉 XDP，让 DHCP 能自己恢复）。

`IPV4_FRAGMENT` 与 `V6_OTHER` 两类是「不是一个普通的未分片 IPv4 报文」时的**显式**
答案，两者都曾是漏洞：

> **IPv4 非首片**：只有首片带 L4 头，而 TCP/UDP 分支过去无条件把 `ihl * 4` 处当
> 成 L4 头读 —— 非首片读出来的「端口」是载荷字节。后果一是**任何超过路径 MTU 的
> 报文只有首片能过**（IGW 侧 1500，本机网卡 9001），其余片按 `UDP_OTHER` 丢，内核
> 重组超时，一次 knock 就这么消失且用户态什么都看不到；后果二是构造一个载荷恰好
> 读成 `67 → 68` 或 `123 → 123` 的非首片可以走 DHCP/NTP 短路。现在在协议分派**之
> 前**单独判 `frag_off & IP_OFFSET` 并 PASS：非首片单独到达是惰性的，内核没有首片
> 就交付不到任何 socket，而首片走的是完整决策树。代价是重组缓冲区可被占用
> （`net.ipv4.ipfrag_high_thresh` 已有上界，和任何没做分片过滤的主机一样）；反过来
> 全丢分片就等于「过滤器本身」成了大 knock 失败的原因。首片（MF=1、offset=0）行为
> 不变，`udp->len` 是整个数据报的长度，所以长度下限判的仍是 knock 而不是碎片。v6 的
> 分片头（`IPPROTO_FRAGMENT`）同理：offset≠0 的片按 `IPV6_FRAGMENT` 放行，首片继续
> 沿扩展头链往下判。
>
> **IPv6**：第一版是一行 `case ETH_P_IPV6: return XDP_PASS`，理由是 demo 主机没有
> v6 服务。但 sshd 默认监听 `[::]:22` —— 哪天 VPC 加上 v6 CIDR（或这个对象被别的
> 主机复用），tcp/22 和其它所有端口就在 v6 上完全无白名单地暴露，而 `deploy-server`
> 的「XDP 已 attach」检查照样是绿的。现在 `handle_ipv6()` 跑同一棵树，**只少了白名
> 单**：LPM trie 里放的是 IPv4 前缀，任何 v6 来源都无法被识别成 relay，所以 v6 上
> 没有 SSH、也没有 knock 端口。放行的只有 ICMPv6 的 ND/MLD 和 PTB（否则主机在 v6
> 子网上活不下去）、DHCPv6/NTP（与 v4 同理）、以及本机自己发起的流的回包
> （`has_local_flow6()`）。**因此 SSH 必须走 IPv4**：网卡上有全局 v6 地址时
> `warnOnGlobalIPv6` 会打一条带地址的 `Warning`，而 `deploy-server` 的
> `$SSH_CONNECTION` 包含性检查对 v6 对端地址会直接失败（没有 IPv4 条目能包含它），
> 所以这不会变成一次静默锁死。事件记录里为此加了地址族字段和一对 v6 地址，v6 的
> DROP 行打的是真实源地址；`serverEventByteSize` 就是这个结构体的 `sizeof`，改结构
> 体必须同时改它（回归：`TestServerFilterFiltersIPv6` 会把日志行读回来比对）。

---

## 0. 前置变量

```bash
cd terraform/demo
terraform init -backend-config="bucket=$TF_STATE_BUCKET"   # 若尚未 init

export RELAY_PUB=$(terraform output -raw relay_public_ip)
export RELAY_PRIV=$(terraform output -raw relay_private_ip)
export SERVER_PUB=$(terraform output -raw server_public_ip)
export SERVER_PRIV=$(terraform output -raw server_private_ip)
echo "relay  pub=$RELAY_PUB  priv=$RELAY_PRIV"
echo "server pub=$SERVER_PUB priv=$SERVER_PRIV"
```

登录 server（只能经 relay 跳板，server 安全组只放行来自 relay 安全组的 22）：

```bash
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV     # == terraform output ssh_jump_command
```

server 上的固定路径：

```bash
REMOTE_DIR=/home/ec2-user/nhp-server
$REMOTE_DIR/etc/xdp.toml                 # 白名单（热更新）
$REMOTE_DIR/etc/nhp_server_xdp.o         # XDP 目标文件
$REMOTE_DIR/logs/server-$(date +%F).log  # 守护进程日志（不是 journald）
$REMOTE_DIR/logs/nhp_server_xdp-$(date +%F).log   # XDP 逐包判决日志
```

---

## 1. 本机验证（不碰 demo）

### 1.1 编译

```bash
make ebpf-objs          # 需要 clang + libbpf-dev
ls -l release/nhp-server/etc/nhp_server_xdp.o
make serverd
```

### 1.2 目标文件静态检查

```bash
# 必须有 xdp section 与 .maps
llvm-readelf -S release/nhp-server/etc/nhp_server_xdp.o | grep -E ' xdp | \.maps '
#   [ 3] xdp   PROGBITS ...
#   [ 6] .maps PROGBITS ...

# 程序名与 map 名（Go 侧按名字取，改名即失配）
bpftool btf dump file release/nhp-server/etc/nhp_server_xdp.o \
  | grep -E "VAR 'nhp_relay_ips'|VAR 'nhp_events'|FUNC 'xdp_server_prog'"
```

> **不要用 `bpftool prog loadall` 加载这个 `.o` 做验证。** 它会在
> `map 'nhp_events': failed to create: Invalid argument(-22)` 处失败：perf event
> array 在 C 里没写 `key_size`/`value_size`（与 AC 的 `events` 写法一致），libbpf
> 不补全而 `cilium/ebpf` 会补全。用 §1.4 的真实守护进程路径验证。

### 1.3 单元测试

```bash
cd nhp       && go test ./utils/ebpf/... && cd ..
cd endpoints && go test ./server/...     && cd ..
```

### 1.4 netns 内的全链路实测（推荐）

> ⚠️ **不要在开发机上直接 `./nhp-serverd run`。** `resolveInterface("")` 会取
> **默认路由网卡**并把过滤器挂上去，结果是本机除 22（白名单内）与 UDP/${NHP_PORT} 外
> 全部被丢弃——包括你当前的 SSH 会话。务必在 netns 或一次性虚拟机里做。

建环境（veth 对，10.99.0.1 = 外部，10.99.0.2 = 被保护主机）：

```bash
sudo ip netns add nhpxdp
sudo ip link add veth-h type veth peer name veth-n
sudo ip link set veth-n netns nhpxdp
sudo ip addr add 10.99.0.1/24 dev veth-h && sudo ip link set veth-h up
sudo ip netns exec nhpxdp sh -c '
  ip addr add 10.99.0.2/24 dev veth-n
  ip link set veth-n up; ip link set lo up
  ip route add default via 10.99.0.1 dev veth-n'      # 必须有默认路由，否则取不到网卡
```

起守护进程（`ip netns exec` 会新建 mount namespace，bpffs 要在里面挂）：

```bash
cp -r release/nhp-server /tmp/nhpsrv
printf 'Enabled = true\nNhpMinFrameBytes = 240\nRelayIPs = ["10.99.0.9"]\n' > /tmp/nhpsrv/etc/xdp.toml

sudo ip netns exec nhpxdp sh -c '
  mount -t bpf bpf /sys/fs/bpf
  cd /tmp/nhpsrv && ./nhp-serverd run'
```

期望日志（`/tmp/nhpsrv/logs/server-<date>.log`）：

```
engine_linux.go:   [Info] server XDP filter configured for udp/${NHP_PORT} with a 240-byte datagram floor (0 = the object's own default)
engine_linux.go:   [Info] relay whitelist applied: 1 prefix(es) active, 0 removed
config.go:         [Info] server XDP engine loaded: filtering udp/${NHP_PORT} with a 240-byte floor, SSH from [10.99.0.9]
engine_linux.go:   [Info] Start listening for server eBPF events (PERF BUFFER)
file.go:           [Info] start watching file /tmp/nhpsrv/etc/xdp.toml
```

> 注意顺序：`relay whitelist applied` 在 `server XDP engine loaded` **之前**。
> 白名单是随 load 一起写进 map 的，写不进去就整个 load 失败（fail-open，什么都不挂）。
> 所以不存在「程序已挂上、白名单还是空的」这个窗口——那个状态下 tcp/22 对所有来源
> 都是关的，而且恰恰是在没法再去修 map 的时候发生。

> 没有这几行、而是 `no .../etc/xdp.toml: the XDP ingress filter is not attached`
> 或 `Enabled = false ...`，说明这台机器根本没有 opt-in——这是预期行为，不是故障。

另开一个终端，确认已 attach：

```bash
sudo ip netns exec nhpxdp ip -details link show veth-n | tail -1
# ... prog/xdp id 305 name xdp_server_prog tag 6fbfd255ac4fee2c jited ...
```

**流量矩阵**（在 host 侧，源地址 10.99.0.1，未进白名单）：

```bash
ping -c1 -W2 10.99.0.2                                        # 期望：100% 丢包
timeout 3 bash -c 'exec 3<>/dev/tcp/10.99.0.2/22'   ; echo $? # 期望：124（超时=被丢）
timeout 3 bash -c 'exec 3<>/dev/tcp/10.99.0.2/${NHP_PORT}'; echo $? # 期望：124
timeout 3 bash -c 'exec 3<>/dev/tcp/10.99.0.2/443'  ; echo $? # 期望：124
python3 -c "import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*100,('10.99.0.2',${NHP_PORT}))"  # <240，丢
python3 -c "import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*300,('10.99.0.2',${NHP_PORT}))"  # ≥240，放行
python3 -c "import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*300,('10.99.0.2',53))"     # 非 ${NHP_PORT}，丢
```

判决日志 `/tmp/nhpsrv/logs/nhp_server_xdp-<date>.log`（实测输出）：

```
20:24:31 test-server [NHP-DROP] REASON=NON_TCP_UDP   SRC=10.99.0.1 DST=10.99.0.2 LEN=84  PROTO=ICMP SPT=0     DPT=0     RELAY=0
20:24:33 test-server [NHP-DROP] REASON=TCP_SSH_OTHER SRC=10.99.0.1 DST=10.99.0.2 LEN=60  PROTO=TCP  SPT=58350 DPT=22    RELAY=0
20:24:36 test-server [NHP-DROP] REASON=TCP_NHP_PORT  SRC=10.99.0.1 DST=10.99.0.2 LEN=60  PROTO=TCP  SPT=46220 DPT=${NHP_PORT} RELAY=0
20:25:26 test-server [NHP-DROP] REASON=TCP_OTHER     SRC=10.99.0.1 DST=10.99.0.2 LEN=60  PROTO=TCP  SPT=53458 DPT=443   RELAY=0
20:24:39 test-server [NHP-DROP] REASON=UDP_SHORT     SRC=10.99.0.1 DST=10.99.0.2 LEN=128 PROTO=UDP  SPT=49355 DPT=${NHP_PORT} RELAY=0
20:24:39 test-server [NHP-PASS] REASON=NHP_DEFAULT   SRC=10.99.0.1 DST=10.99.0.2 LEN=328 PROTO=UDP  SPT=38438 DPT=${NHP_PORT} RELAY=0
20:24:39 test-server [NHP-DROP] REASON=UDP_OTHER     SRC=10.99.0.1 DST=10.99.0.2 LEN=328 PROTO=UDP  SPT=57077 DPT=53    RELAY=0
```

**回包矩阵**（被保护主机**自己发起**的连接，必须能收到响应 —— 这就是 OTP 邮件
发不出去那个 bug 的最小复现）。先在 host 侧起两个「外部服务」：

```bash
# UDP 回显（冒充 DNS），以及一个 TCP 服务（冒充 SES 的 tcp/587）
python3 - <<'PY' & 
import socket
s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM); s.bind(('10.99.0.1',5353))
while True:
    d,a=s.recvfrom(2048); s.sendto(b'PONG:'+d,a)
PY
python3 - <<'PY' &
import socket
s=socket.socket(); s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1)
s.bind(('10.99.0.1',8080)); s.listen(5)
while True:
    c,_=s.accept(); c.sendall(b'HELLO\n'); c.close()
PY
```

再从 netns 里（= 被保护主机）当客户端打出去：

```bash
sudo ip netns exec nhpxdp python3 - <<'PY'
import socket
s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM); s.settimeout(3)
s.connect(('10.99.0.1',5353)); s.send(b'PING')
print("A udp-reply:", s.recv(100))          # 期望 b'PONG:PING'；修复前：timed out
t=socket.socket(); t.settimeout(3)
t.connect(('10.99.0.1',8080))
print("B tcp-connect:", t.recv(100))        # 期望 b'HELLO\n'；修复前：timed out
PY
```

```
A udp-reply: b'PONG:PING'
B tcp-connect: b'HELLO\n'
```

这两类回包**不逐包写日志**（见上表第 4 列：它们的量等于本机的吞吐量），验证
要看每分钟一次的汇总行 —— 等最多 60 秒：

```
<时间> test-server [NHP-STAT] VERDICT=PASS REASON=UDP_ESTABLISHED WINDOW=60s PKTS=1 LOGGED=0 TOTAL=1
<时间> test-server [NHP-STAT] VERDICT=PASS REASON=TCP_ESTABLISHED WINDOW=60s PKTS=3 LOGGED=0 TOTAL=3
```

`PKTS` 是本窗口的增量，`LOGGED` 是其中进了事件流的条数（这两类恒为 0），
`TOTAL` 是 attach 以来的累计。判断「回包通了」看 `PKTS > 0`；判断「回包被丢
了」看同期有没有 `[NHP-DROP] REASON=UDP_OTHER ... SPT=5353`。

**反向确认（最重要的一条）**：回包通路不等于开放端口。在 netns 里 `bind` 一个
tcp/8080 的 **listener**，从 host 侧连它，必须仍然被丢 —— 放行只认
established/connected，不认 LISTEN：

```bash
sudo ip netns exec nhpxdp python3 -c "
import socket; s=socket.socket(); s.bind(('10.99.0.2',8080)); s.listen(5); s.accept()" &
timeout 3 bash -c 'exec 3<>/dev/tcp/10.99.0.2/8080'; echo $?   # 期望：124（超时=被丢）
```

**未 connected 的客户端 socket（DHCP / NTP）**——上面那条「只认 connected」的
性质有两个必须开口子的例外，而 DHCP 那个曾经把整台主机弄下线（见开头的说明）。
在 netns 里用**只 bind、不 connect** 的 socket 收，才是真实形状：

```bash
# 67 → 68：DHCP 续租应答。收端只 bind
sudo ip netns exec nhpxdp python3 - <<'PY' &
import socket
s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM); s.settimeout(5)
s.bind(('10.99.0.2',68)); print("dhcp:", s.recvfrom(64))    # 期望收到；修复前：timeout
PY
sleep 1
sudo python3 - <<'PY'
import socket
s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM); s.bind(('10.99.0.1',67))
s.sendto(b'DHCPACK',('10.99.0.2',68))
PY
wait
# 123 → 123：NTP 应答（chronyd 的收发是同一个 bind 在 123 的 socket）
# 同上，把两个端口都换成 123 即可
```

事件日志里对应 `[NHP-PASS] REASON=DHCP_CLIENT ... SPT=67 DPT=68`；`NTP_CLIENT`
只计数，看 `[NHP-STAT]`。同一组断言（含「普通端口的未 connected socket 仍然收
不到」这条反向确认）已经写成单元测试，有 CAP_BPF 时直接跑：

```bash
sudo -E env "PATH=$PATH" go test -run TestServerFilterPassesClientReplies -v ./utils/ebpf/   # 在 nhp/ 下
```

**白名单两条分支**（把 10.99.0.1 加进白名单后重测）：

```bash
sed -i 's/10.99.0.9/10.99.0.1/' /tmp/nhpsrv/etc/xdp.toml     # 热更新，无需重启
sleep 1
timeout 3 bash -c 'exec 3<>/dev/tcp/10.99.0.2/22'            # 期望：Connection refused（瞬时）——包被放行，内核回 RST
python3 -c "import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*40,('10.99.0.2',${NHP_PORT}))"  # 40 字节也放行
```

```
20:25:13 test-server [NHP-PASS] REASON=SSH_RELAY SRC=10.99.0.1 ... DPT=22    RELAY=1
20:25:13 test-server [NHP-PASS] REASON=NHP_RELAY SRC=10.99.0.1 ... DPT=${NHP_PORT} RELAY=1
```

`SSH_RELAY` 只在 SYN 上出现一行；这条连接后续的包（scp 的数据段等）只进计数
器，一分钟后体现在 `[NHP-STAT] ... REASON=SSH_RELAY PKTS=<很大> LOGGED=1` 里。

> 「超时（rc=124）」= 被 XDP 丢；「Connection refused（毫秒级）」= 被放行、内核回了
> RST。这一对差异就是后面线上 before/after 对比要用的信号。

**热更新与 map 内容**（`bpftool` 要进守护进程的 mount namespace）：

```bash
PID=$(pgrep -x nhp-serverd)
sudo nsenter -t $PID -m -n bpftool map dump pinned /sys/fs/bpf/nhp_relay_ips
```

> `nhp_relay_ips` 是 **LPM trie**（不再是 hash）：key 为 8 字节
> `struct relay_prefix_key { __u32 prefixlen; __be32 addr; }` —— 前 4 字节是
> 主机序的前缀长度，后 4 字节是网络序地址。白名单条目既可以是主机地址
> （即 `/32`），也可以是 CIDR 前缀；线上会下发 VPC 子网前缀，这样 relay 被
> 重建、换了私网地址后 SSH 仍然进得来（见 CLAUDE.md）。

```bash
# 按 IP 精确查一条：prefixlen=32（小端 20 00 00 00）+ 点分四段的十六进制
IP=10.99.0.1
sudo nsenter -t $PID -m -n bpftool map lookup pinned /sys/fs/bpf/nhp_relay_ips \
  key hex 20 00 00 00 $(echo $IP | tr '.' ' ' | xargs printf '%02x %02x %02x %02x')
```

改 `xdp.toml` 后守护进程日志应出现（实测）：

```
config.go:625:      [Info] xdp config: /tmp/nhpsrv/etc/xdp.toml has been updated
engine_linux.go:656:[Info] relay whitelist applied: 2 prefix(es) active, 0 removed
config.go:668:      [Info] xdp relay whitelist applied: [10.99.0.1 192.0.2.7]
```

> 只有 `RelayIPs` 支持热更新。`Enabled` 与 `NhpMinFrameBytes` 只在启动时读一次
> （长度下限会被写进对象的 `.rodata`），改完要重启才生效；改了但没重启时日志里
> 会有一条 `Warning` 说明这一点。

**退出与清理**（顺带验证 `Stop()` 会摘掉 XDP）：

```bash
sudo pkill -x nhp-serverd
# 日志：Successfully removed BPF file: /sys/fs/bpf/xdp_server_prog
#       Successfully removed BPF file: /sys/fs/bpf/nhp_relay_ips
#       server XDP link detached and closed / server eBPF engine unloaded
sudo ip netns exec nhpxdp ip -details link show veth-n | grep -c xdpgeneric   # 0
sudo ip netns del nhpxdp; sudo ip link del veth-h; sudo rm -rf /tmp/nhpsrv
```

---

## 2. 部署前基线（在 demo 现网上采集并存档）

### 2.1 ★ 上线前强制预检：白名单里的 IP 必须等于真正到达 server 的源 IP

这是**唯一一条做错就会永久失联**的检查，必须在第一次部署前手工确认，因为：

* 没有破玻璃入口，SSH 走的就是被过滤的 22；
* XDP 不做连接跟踪，白名单写错时**已经建立的 SSH 会话也会当场断**（`deploy-server`
  自己的会话同样会断）；
* 恢复手段只有「重建实例」或「摘盘改文件」（§6.2）。

CI 是 `ProxyJump` 经 relay 连 **server 私网地址**的（
`.github/workflows/deploy-demo-v2.yml` 里 `Host nhp-server → HostName <server_private_ip>`），
同一 VPC 内走私网路径，server 看到的源地址是 relay 的**私网地址**。第一版流水线
只把 `dig relay.opennhp.org` 的结果（relay 的**公网 EIP**）写进白名单，于是
`deploy-server` 在 `Starting nhp-serverd...` 之后的每一条 ssh 都被驱动层丢掉
（`channel 0: open failed: connect failed: Connection timed out`）。现在 `configure`
把 `relay_private_ip` 和公网 EIP **两条**都写进 `RELAY_IPS`，白名单与真实源地址仍
必须对上：

```bash
# a) 流水线将写入白名单的两个地址
cd terraform/demo && terraform output -raw relay_private_ip   # = $RELAY_PRIV，SSH 真正的源地址
dig +short relay.opennhp.org @1.1.1.1                         # = $RELAY_PUB，公网路径

# b) server 实际看到的 SSH 源地址（在 server 上执行，保持另一个 SSH 会话在线）
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV \
  "ss -tn state established '( sport = :22 )' | awk 'NR>1{print \$3, \$4}'"
#    左列是本机 22，右列 Peer Address = 真正需要进白名单的地址

# c) 抓一次 SYN 交叉印证（另开会话时触发）
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV \
  "sudo timeout 20 tcpdump -ni any 'tcp port 22 and tcp[tcpflags] & tcp-syn != 0' -c 3"

# d) relay 转发过来的 NHP 流量源地址
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV \
  "sudo timeout 20 tcpdump -ni any 'udp port ${NHP_PORT}' -c 5"
```

**判定**：(b)/(c) 显示的地址必须出现在 (a) 的结果里。期望的 `xdp.toml` 内容：

```bash
# RelayIPs 是数组，可以多元素；RELAY_IPS 里用逗号分隔，
# scripts/generate-nhp-keys.sh 负责加引号、去重、校验是否 IPv4
RelayIPs = ["<relay 私网 IP>", "<relay 公网 EIP>"]
```

若 (b)/(c) 的地址两条都不是（例如 relay 重建换了私网地址、或 `ProxyJump` 拓扑改了），
**先不要部署**，先把该地址加进 `configure` 的 `RELAY_IPS`。`deploy-server` 的
`Check this job's own SSH source is in the XDP whitelist` 这一步会用
`$SSH_CONNECTION` 向主机反问「你看到的对端地址是多少」，对不上就在第一次 scp 之前
失败，所以忘了这条也是红流水线而不是失联——但预检仍值得做，它能在跑流水线之前
就发现问题（见 §7 风险 1）。

### 2.2 主机内基线（在 server 上执行，输出存档）

```bash
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV 'bash -s' <<'EOS' | tee before-host.txt
set +e
echo "== uname ==";            uname -r
echo "== xdp attached ==";     ip -details link show | grep -A3 xdpgeneric || echo "(none)"
echo "== bpf pins ==";         sudo ls -l /sys/fs/bpf/ 2>&1
echo "== bpf progs ==";        sudo bpftool prog show 2>&1 | grep -i xdp || echo "(none)"
echo "== unit caps ==";        systemctl cat nhp-serverd | grep -iE 'capab|ExecStartPre' || echo "(none)"
echo "== bpffs mounted ==";    mountpoint /sys/fs/bpf 2>&1
echo "== etc ==";              ls -l /home/ec2-user/nhp-server/etc/ | grep -E 'xdp' || echo "(no xdp.toml / .o)"
echo "== service ==";          systemctl is-active nhp-serverd
echo "== listening ==";        sudo ss -lunp | grep ${NHP_PORT}
echo "== ebpf log lines ==";   grep -icE 'ebpf|xdp' /home/ec2-user/nhp-server/logs/server-$(date +%F).log 2>/dev/null
EOS
```

部署前期望：`xdp attached` 为 `(none)`，`/sys/fs/bpf/` 里没有 `xdp_server_prog` /
`nhp_relay_ips`，unit 没有 `AmbientCapabilities`，`etc/` 下没有 `xdp.toml` 和
`nhp_server_xdp.o`，`nhp-serverd` 正在监听 UDP/${NHP_PORT}。

### 2.3 从 relay（VPC 内）看的基线

XDP 的行为变化主要体现在 VPC 内侧，所以对比要从 relay 上打：

```bash
ssh ec2-user@$RELAY_PUB "bash -s $SERVER_PRIV" <<'EOS' | tee before-from-relay.txt
set +e
SRV=$1
echo "== ping ==";        ping -c2 -W2 $SRV | tail -2
echo "== tcp/22 ==";      timeout 3 bash -c "exec 3<>/dev/tcp/$SRV/22";      echo "rc=$?"
echo "== tcp/443 ==";     timeout 3 bash -c "exec 3<>/dev/tcp/$SRV/443";     echo "rc=$?"
echo "== tcp/${NHP_PORT} ==";   timeout 3 bash -c "exec 3<>/dev/tcp/$SRV/${NHP_PORT}";   echo "rc=$?"
echo "== udp/${NHP_PORT} 短包 ==" ; python3 -c "
import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*64,('$SRV',${NHP_PORT}));print('sent')"
EOS
```

部署前期望：ping 通；tcp/22 `rc=0`（连上）；tcp/443、tcp/${NHP_PORT} 是**立即** refused
（`rc=1`，安全组内部互通、端口没人听）；短 UDP 包直达 user space。

### 2.4 从公网看的基线

```bash
nmap -Pn -n -p 22,80,443,${NHP_PORT} $SERVER_PUB      | tee before-from-internet.txt
sudo nmap -Pn -n -sU -p ${NHP_PORT} $SERVER_PUB       | tee -a before-from-internet.txt
ping -c3 $SERVER_PUB                            | tee -a before-from-internet.txt
```

> 公网视角在部署前后**基本不会变**，这是预期的：安全组早就只放行 UDP/${NHP_PORT}（22 只对
> relay 安全组开）。XDP 的增量在于 VPC 内侧、以及公网上那些能穿过安全组的
> UDP/${NHP_PORT} 噪声包（`nmap -sU` 的空包会从「进 user space」变成「内核丢弃」，外部看
> 都是 `open|filtered`，只能在 §4.3 的事件日志里看出差别）。

### 2.5 业务基线（确认 NHP 本身可用）

```bash
# 浏览器 demo（经 relay → server）
curl -sS -o /dev/null -w '%{http_code}\n' https://demo.opennhp.org/
# 或用原生 agent 敲一次门，记录成功日志
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV \
  "tail -20 /home/ec2-user/nhp-server/logs/server-$(date +%F).log"
```

---

## 3. 部署过程中的 CI 关卡

按顺序盯这几个 step（`deploy-demo-v2`）：

| Job / Step | 作用 | 失败表现 |
| --- | --- | --- |
| `configure` → `Resolve the relay addresses for the nhp-server XDP whitelist` | `terraform output -raw relay_private_ip` 必须是 IPv4，且 `dig +short relay.opennhp.org @1.1.1.1` 必须**恰好一条** A 记录，否则 `exit 1` | 整个流水线红，**不渲染 / 不 scp / 不重启**，线上继续用旧白名单 |
| `deploy-server` → `Check this job's own SSH source is in the XDP whitelist` | `ssh nhp-server 'printf "%s" "${SSH_CONNECTION%% *}"'` 拿到主机看到的对端地址，必须出现在待部署的 `xdp.toml` 的 `RelayIPs` 里 | 红；在**第一次 scp 之前**失败，线上白名单未被触碰。这是把「白名单写错 → 永久失联」变成「部署失败」的关卡 |
| `build` → `test -s release/nhp-server/etc/nhp_server_xdp.o` | 目标文件必须编出来 | 红；避免 `.o` 缺失导致线上静默 fail-open |
| `deploy-server` → `Deploy nhp-serverd and plugins` | scp `.o` + `xdp.toml`，安装 `/etc/systemd/system/nhp-serverd.service.d/10-ebpf.conf`（`CAP_BPF CAP_NET_ADMIN CAP_PERFMON` + 准备 `/sys/fs/bpf`），重启 | — |
| `deploy-server` → `Verify the XDP ingress filter attached` | `ip -details link show \| grep -q xdpgeneric`，没挂上就红 | 这是把「fail-open 静默失效」变成「部署失败」的唯一关卡 |
| `deploy-server` → `Verify the XDP loader gave its capabilities back` | 读 `/proc/<pid>/task/*/status`，每个线程的 `CapAmb` 必须为 0，`CapPrm`/`CapEff` 不得含 `CAP_NET_ADMIN`(12) / `CAP_PERFMON`(38)；只允许剩 `CAP_BPF`(39) | 红；挂载完成后 `dropLoaderPrivileges` 会交还能力，这一步保证「交还」没有悄悄失效。`kernel.unprivileged_bpf_disabled` 非 0 时保留 `CAP_BPF`（否则 `xdp.toml` 热加载写不了 map）。注意 `dropLoaderPrivileges` 是**只减这三个能力**，不是把能力集清空；本 unit 授予的恰好就是这三个，所以结果为空集，这一步才能这么断言 |
| `deploy-server` → `Verify the server's own outbound flows still get their replies` | 在 server 上 `getent ahostsv4 <SMTP_HOST>`（udp/53 应答）+ `/dev/tcp/<SMTP_HOST>/587`（SYN-ACK），各重试 3 次 | 红；这是把「回包被丢 → OTP 邮件发不出」变成「部署失败」的关卡。DNS 失败看 `REASON=UDP_OTHER`，SMTP 失败看 `REASON=TCP_OTHER` |
| `deploy-server` → `Verify the host can still renew its DHCP lease` | `sudo networkctl renew <iface>`，然后事件日志里 `REASON=DHCP_CLIENT` 的行数必须增加，且网卡地址不变 | 红；这是把「续租应答被丢 → 一个租期后整机失联」变成「部署失败」的关卡。失败时看 `REASON=UDP_OTHER ... SPT=67`。**其它任何探测都看不见这个故障**，因为租期没到之前一切正常 |

本地复现 DNS 关卡：

```bash
dig +short relay.opennhp.org @1.1.1.1 | grep -E '^[0-9.]+$' | wc -l        # 必须是 1
```

---

## 4. 部署后验证

### 4.1 重跑 §2 的三段脚本并 diff

```bash
# 与 2.2 / 2.3 / 2.4 完全相同的命令，输出到 after-*.txt
diff before-host.txt        after-host.txt
diff before-from-relay.txt  after-from-relay.txt
diff before-from-internet.txt after-from-internet.txt
```

### 4.2 内核侧（在 server 上）

```bash
ip -details link show | grep -A3 xdpgeneric
#  ens5: ... xdpgeneric ...
#      prog/xdp id <N> name xdp_server_prog tag <...> jited

sudo bpftool prog show name xdp_server_prog
sudo ls -l /sys/fs/bpf/            # 期望出现 xdp_server_prog、nhp_relay_ips
systemctl cat nhp-serverd | grep -iE 'AmbientCapabilities|CapabilityBoundingSet|ExecStartPre'
#  ExecStartPre=+/bin/sh -c 'mountpoint -q /sys/fs/bpf || mount ...'
#  AmbientCapabilities=CAP_BPF CAP_NET_ADMIN CAP_PERFMON
```

可选的量化计数（用完记得关，`bpf_stats` 有额外开销）：

```bash
sudo sysctl -w kernel.bpf_stats_enabled=1
sudo bpftool prog show name xdp_server_prog     # run_cnt / run_time_ns 递增即在处理包
sudo sysctl -w kernel.bpf_stats_enabled=0
```

### 4.3 白名单 map 与配置

```bash
sudo bpftool map dump pinned /sys/fs/bpf/nhp_relay_ips
IP=$RELAY_PUB     # 以及 relay 私网地址（见 §2.1）
sudo bpftool map lookup pinned /sys/fs/bpf/nhp_relay_ips \
  key hex $(echo $IP | tr '.' ' ' | xargs printf '%02x %02x %02x %02x')
# 命中：{"key": ..., "value": 1}；未命中：Error: ... No such file or directory

cat /home/ec2-user/nhp-server/etc/xdp.toml
# 启动时的白名单在 `server XDP engine loaded ... SSH from [...]` 这一行里；
# `xdp relay whitelist applied` 只有热更新（改 xdp.toml）才会打
grep -E 'xdp relay whitelist applied|server XDP engine loaded' \
  /home/ec2-user/nhp-server/logs/server-$(date +%F).log | tail -1
# 另外两条要确认「没出现」的日志：写了个没有 RelayIPs 的 xdp.toml 会打第一行，
# 写了个每项都解析不了（主机名 / IPv6 / 笔误 / 带行内注释）的会打第二行。
# 两种情况都保留原有白名单，不会把 tcp/22 全关掉
grep 'lists no RelayIPs' /home/ec2-user/nhp-server/logs/server-$(date +%F).log | tail -1
grep 'unusable RelayIPs entry' /home/ec2-user/nhp-server/logs/server-$(date +%F).log | tail -1
```

### 4.4 事件日志（判决可观测）

```bash
LOG=/home/ec2-user/nhp-server/logs/nhp_server_xdp-$(date +%F).log
tail -f $LOG
# 按 REASON 汇总，看扫描噪声被挡在哪里（注意：这是被限速后的样本数，不是真实包数）
grep -o 'REASON=[A-Z_]*' $LOG | grep -v NHP-STAT | sort | uniq -c | sort -rn
# 真实包数看每分钟的汇总行：TOTAL 是 attach 以来的累计，PKTS 是本窗口增量
grep NHP-STAT $LOG | tail -20
# 确认 relay 的 SSH/转发走的是 PASS 分支（SSH_RELAY 每会话一行，在 SYN 上）
grep -E 'SSH_RELAY|NHP_RELAY' $LOG | tail -5
# 日志占用（保留策略：>14 天或总量 >256 MiB 从最旧删起，当天的不删）
du -sh $(dirname $LOG) && ls -lt $(dirname $LOG)/nhp_server_xdp-*.log | head
```

> **抓不到包是正常的**：generic XDP 在 AF_PACKET tap 之前执行，被丢弃的包
> `tcpdump` 看不到（实测 `tcpdump -ni <iface> 'tcp port 443 or icmp'` 为 0 包，
> 而同期 `run_cnt` 在涨）。判决只能从上面这份事件日志看。

### 4.5 通路验证（从 relay 打）

重跑 §2.3，期望变成：

```
ping        →  100% packet loss        （之前：通）
tcp/22      →  rc=0                    （不变；relay 在白名单内，否则就是 §7 风险 1）
tcp/443     →  rc=124 超时              （之前：rc=1 立即 refused）
tcp/${NHP_PORT}   →  rc=124 超时              （之前：rc=1 立即 refused）
udp/${NHP_PORT} 64B → 事件日志 UDP_SHORT DROP（若 relay 在白名单则是 NHP_RELAY PASS）
```

从公网侧另找一台机器（不在白名单内）打一发合法长度的 UDP，确认 NHP 仍然可达：

```bash
python3 -c "
import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*300,('$SERVER_PUB',${NHP_PORT}))"
# server 侧事件日志：REASON=NHP_DEFAULT ... [NHP-PASS]
```

### 4.6 出站回包（OTP 邮件依赖这条）

在 server 上跑，等价于 CI 的 `Verify the server's own outbound flows still get
their replies`；两条都必须过，否则 OTP 邮件一定发不出去：

```bash
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV '
  getent ahostsv4 email-smtp.us-east-2.amazonaws.com | head -1   # udp/53 应答回得来
  timeout 8 bash -c "exec 3<>/dev/tcp/email-smtp.us-east-2.amazonaws.com/587" && echo "smtp ok"
  # 回包只计数不逐包记日志，所以看汇总行的 TOTAL（最多等 60s 刷新一次）
  grep "REASON=UDP_ESTABLISHED" ~/nhp-server/logs/nhp_server_xdp-$(date +%F).log | tail -1'
```

DNS 挂了会在事件日志里表现为 `REASON=UDP_OTHER ... SPT=53`（应答被丢），
SMTP 挂了则是 `REASON=TCP_OTHER ... SPT=587`。两者都说明 `has_local_flow()`
没有生效 —— 通常是 `.o` 是旧的（未重新 `make ebpf-objs`），或内核太老不支持
`bpf_sk_lookup_*`（那种情况下整个程序装载失败，是 fail-open，见 §6.1）。

### 4.6b DHCP 续租（★ 这条决定主机还能不能被找到）

等价于 CI 的 `Verify the host can still renew its DHCP lease`。**这一条不过就
不要走开**：主机会在当前租期到点时丢掉自己的 IPv4 地址，届时所有端口一起失联，
只能重启实例。

```bash
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV '
  IFACE=$(ip -4 -o route show default | awk "{print \$5; exit}")
  LOG=~/nhp-server/logs/nhp_server_xdp-$(date +%F).log
  ip -4 -o addr show dev $IFACE scope global          # 记住地址
  BEFORE=$(grep -c REASON=DHCP_CLIENT $LOG 2>/dev/null || true)
  sudo networkctl renew $IFACE
  sleep 5
  echo "DHCP_CLIENT 行数: ${BEFORE:-0} -> $(grep -c REASON=DHCP_CLIENT $LOG 2>/dev/null || true)"
  grep -E "SPT=(67|68)" $LOG | tail -3
  ip -4 -o addr show dev $IFACE scope global          # 地址必须没变'
```

期望：`DHCP_CLIENT` 行数增加，地址不变。若看到的是
`REASON=UDP_OTHER ... SPT=67 DPT=68`，说明部署的 `.o` 是不带 DHCP 放行分支的
旧版本，立刻回到 §6.2 的恢复流程，不要等租期到点。

顺便看一眼时钟（`NTP_CLIENT` 被丢的表现是时钟慢慢漂，NHP 的时间戳窗口最终会
拒掉敲门）：

```bash
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV 'chronyc -n tracking | head -4'
```

### 4.7 业务回归（最重要的一条）

```bash
# 浏览器 demo 全链路敲门一次，确认 relay → server → ac 正常
curl -sS -o /dev/null -w '%{http_code}\n' https://demo.opennhp.org/
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV \
  "grep -iE 'knock|ack' /home/ec2-user/nhp-server/logs/server-$(date +%F).log | tail -10"
```

以及 **AC 端零变更**的回归（engine 被抽取到 `nhp/utils/ebpf` 后要确认 AC 行为未漂移）：

```bash
ssh -J ec2-user@$RELAY_PUB ec2-user@$(cd terraform/demo && terraform output -raw ac_private_ip) '
  ip -details link show | grep -A3 xdpgeneric | grep xdp_white_prog
  sudo ls /sys/fs/bpf/ | sort
  ls -lt /home/ec2-user/nhp-ac/logs/ | head -5'      # nhp_accept / nhp_deny 日志名不变
```

---

## 5. before / after 速查表

| 观测点 | 部署前 | 部署后 |
| --- | --- | --- |
| `ip -details link show \| grep xdpgeneric` | 无 | `prog/xdp ... name xdp_server_prog` |
| `/sys/fs/bpf/` | 无 server 相关 pin | `xdp_server_prog`、`nhp_relay_ips` |
| `etc/` | 无 `xdp.toml` / `nhp_server_xdp.o` | 两者都在 |
| unit 能力集 | 无 `AmbientCapabilities` | `CAP_BPF CAP_NET_ADMIN CAP_PERFMON` + `ExecStartPre` 挂 bpffs |
| server 日志 | 无 eBPF 相关行 | `relay whitelist applied: N prefix(es) active`、紧接着 `server XDP engine loaded: ... SSH from [...]` |
| `logs/nhp_server_xdp-*.log` | 不存在 | 持续写入 PASS/DROP 判决（限速）+ 每分钟 `[NHP-STAT]` 汇总；>14 天或 >256 MiB 自动清理 |
| relay → server ICMP | 通 | 全丢（`NON_TCP_UDP`） |
| relay → server tcp/443 | 立即 refused（rc=1） | 超时（rc=124，`TCP_OTHER`） |
| relay → server tcp/22 | 连上 | 连上（`SSH_RELAY`，前提：白名单正确） |
| 公网 → UDP/${NHP_PORT} <240B | 进 user space 由 `RecvPrecheck` 丢 | 内核丢（`UDP_SHORT`），不消耗用户态 |
| 公网 → UDP/${NHP_PORT} ≥240B | 通 | 通（`NHP_DEFAULT`）——NHP 语义不变 |
| server 自己的 DNS 查询 | 通 | 通（应答 `UDP_ESTABLISHED`，见 `[NHP-STAT]`）——**这一条曾经是坏的** |
| server 自己的 DHCP 续租 | 通 | 通（`DHCP_CLIENT`）——**这一条曾经是坏的，代价是整台主机失联** |
| server 自己的 NTP 对时 | 通 | 通（`NTP_CLIENT`，见 `[NHP-STAT]`）——**这一条曾经是坏的** |
| server 自己的 SMTP/HTTPS 出站 | 通 | 通（`TCP_ESTABLISHED`，见 `[NHP-STAT]`） |
| 公网 → server 上任意 LISTEN 端口 | 由安全组决定 | 一律丢（socket 是 LISTEN，不构成回包） |
| 任何来源 → server 的 v6 端口（主机有 v6 地址时） | 由安全组决定 | 一律丢（`V6_OTHER`）——**SSH 必须走 IPv4** |
| 超过路径 MTU 的 knock（分片） | 通 | 通（首片走决策树，其余片 `IPV4_FRAGMENT`）——**这一条曾经是坏的** |
| `nmap` 公网扫描结果 | — | 基本不变（安全组本就挡住 TCP） |

---

## 6. 异常演练与回滚

### 6.1 fail-open 路径（安全方向，可在线上演练）

```bash
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV '
  mv ~/nhp-server/etc/nhp_server_xdp.o ~/nhp-server/etc/nhp_server_xdp.o.bak
  sudo systemctl restart nhp-serverd; sleep 3
  grep -i "fail-open" ~/nhp-server/logs/server-$(date +%F).log | tail -2
  ip -details link show | grep -c xdpgeneric'     # 0 → 无过滤但服务照常
# 恢复
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV '
  mv ~/nhp-server/etc/nhp_server_xdp.o.bak ~/nhp-server/etc/nhp_server_xdp.o
  sudo systemctl restart nhp-serverd'
```

期望日志：`server eBPF engine load failed, fail-open (no XDP ingress filter): ...`。
注意这时主机重新「裸奔」，只在需要验证 CI 关卡（§3 最后一行）时做，做完立刻恢复。

### 6.2 失联后的恢复

**没有破玻璃通道。** 先分清是哪一类失联，两类的代价差一个数量级：

**(a) 主机丢了自己的 IPv4 地址**（DHCP 续租被丢，或任何同类问题）——现象是
**所有**端口一起超时，敲门端口也不通。**直接重启实例**即可，不用碰磁盘：

```bash
aws ec2 reboot-instances --instance-ids <server-instance-id> --region us-east-2
```

DHCP 在 `nhp-serverd` 启动之前跑完，所以地址会回来，并且在守护进程重新挂上过滤
器之前有一段 tcp/22 敞开的窗口。抓住这个窗口做二选一：

```bash
# 要么把过滤器停掉（链接句柄在进程里，进程停 = 过滤器摘掉），慢慢排查
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV 'sudo systemctl stop nhp-serverd'
# 要么直接重跑带修复的流水线
```

从此之后这一类会自愈：`serverengine.go` 的看门狗发现网卡连续两分钟没有可路由的
IPv4 地址就会打 `Critical` 并摘掉 XDP（日志里 `xdp watchdog:`），下一次 DHCP
交互就能把地址拿回来。它只在「主机反正什么都答不了」的状态下才会触发。

**(b) 白名单写错**（过滤器挂着、地址也在，只是 tcp/22 的源地址不在名单里）——
重启没用，重启后同一份 `xdp.toml` 会再次生效。按代价从低到高：

1. **摘盘改文件**（保状态）：停实例 → detach 根卷 → 挂到另一台实例 →
   改 `/home/ec2-user/nhp-server/etc/xdp.toml`（或 `systemctl disable nhp-serverd`）→
   挂回 → 启动。EIP 由 `aws_eip` 关联保留。
2. **重建实例**：`terraform apply -replace=aws_instance.server`。新实例上
   `nhp-serverd` 尚未部署、XDP 未挂载，SSH 暂时开放，CI 重新部署后才上锁。

分辨 (a) / (b) 的办法：(a) 连敲门都不通（浏览器 demo 整条链路挂），(b) 敲门
正常、只有 SSH 不通。

**演练时的自保**：改白名单前先在 relay 上开一个持续 ping 与一个 `while` 循环连
tcp/22 的观测窗口，一旦 `SSH_RELAY` 消失立刻把 `xdp.toml` 改回（热更新，无需重启）。

### 6.3 整体回滚（下线该特性）

```bash
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV '
  sudo rm -f /etc/systemd/system/nhp-serverd.service.d/10-ebpf.conf
  sudo systemctl daemon-reload
  rm -f ~/nhp-server/etc/xdp.toml ~/nhp-server/etc/nhp_server_xdp.o
  sudo systemctl restart nhp-serverd
  ip -details link show | grep -c xdpgeneric'      # 0
```

停掉 `nhp-serverd` 也会摘掉过滤器（链接句柄在进程内存里），与 AC 不同的是 server
侧**没有** fail-closed 兜底——这是刻意的：server 挂了就没有 NHP 服务可保护。

---

## 7. 已知风险清单（验证时重点看）

1. **白名单是公网 IP、实际源地址是私网 IP** —— 已发生过一次：第一版流水线只写
   `dig relay.opennhp.org` 的公网 EIP，而 CI/relay 经 VPC 内网连 server 私网地址，
   server 看到的是 relay 的**私网** IP，`deploy-server` 在重启 `nhp-serverd` 之后
   所有 ssh 全部超时。现在 `RELAY_IPS` = 「私网 IP + 公网 EIP」两条，且
   `deploy-server` 增加了用 `$SSH_CONNECTION` 反查源地址的预检关卡（§3）。
   **仍要注意**：relay 重建、`ProxyJump` 拓扑变化、多加一台跳板机都会引入新的源
   地址，届时必须同步 `RELAY_IPS`；改白名单前先做 §2.1。
2. **fail-open 是静默的** —— 内核旧、能力缺失、`.o` 丢失都只打一行 warning。日常巡检
   请把 `ip -details link show | grep -q xdpgeneric` 纳入告警，而不是只看
   `systemctl is-active`。
3. **入站方向没有连接跟踪** —— 白名单变更对已建立的入站连接同样生效（每个包都要
   重新过一遍白名单），改 `xdp.toml` 等同于改防火墙，SSH 会话会当场断。
   出站方向相反：本机发起的连接由内核 socket 表决定（`has_local_flow()`），
   socket 还在就一直放行，与 `xdp.toml` 无关。
4. **`Enabled` / `NhpMinFrameBytes` 只是记录值** —— 前者不会阻止 attach，后者编译进
   `.o`（`NHP_MIN_UDP_LEN`）。改长度下界要重新 `make ebpf-objs` 并重发 `.o`。
5. **事件日志有速率限制** —— 每个 action 每 CPU 每秒 16 条（`NHP_EVENT_BURST`），大流量
   扫描下日志条数少于真实丢包数；准确数字看每分钟的 `[NHP-STAT]` 行（`nhp_action_stats`
   对每个包计数，与限速无关），或 `bpf_stats` 的 `run_cnt`。
6. **回包放行依赖 `bpf_sk_lookup_{tcp,udp}`** —— 内核 ≥4.20 才有。太老的内核上整个
   程序装载失败（fail-open，见风险 2），而不是「只丢回包」；升级 `.o` 时如果只换了
   `.o` 没换二进制也没关系，两者之间没有新增接口。IPv6 用的是同一个 helper 的 v6
   元组（`has_local_flow6()`），不是放行 —— v6 现在也在过滤范围内，见决策表。
7. **出站回包不覆盖 unconnected UDP socket** —— `has_local_flow()` 的放行条件是
   socket 已 `connect()`（glibc/Go 的 resolver、`net/smtp` 都是）。**已经踩过两次**：
   DHCP 客户端（raw / 只 bind 在 68）导致主机丢地址整机失联，chronyd（收发都在
   bind 的 123 上）导致时钟漂移；这两类现在按端口对显式放行，各有 `DHCP_CLIENT` /
   `NTP_CLIENT` 一个 action。若将来又有组件用未连接的 UDP socket 收外部响应
   （SNMP trap、某些服务发现），表现同样是 `REASON=UDP_OTHER`，而且**可能延迟到
   下一个租期/超时才暴露**。加放行分支时要写清楚「为什么 socket 表答不了」，
   不要靠把整段端口区间放开来修，更不要把「凡发给已 bind 端口的包都放行」写进去
   —— 那等于取消这个过滤器。
8. **这类故障会延迟暴露，部署当时的绿灯不算数** —— DHCP 那次是部署后约一个租期
   才失联。凡是动 `nhp_server_xdp.c` 的改动，除了 §4.6 / §4.6b 的探测，还要
   在改完的**一小时后**再回来看一眼主机是否还在（`ip -4 addr`、敲一次门）。
