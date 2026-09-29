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

决策树与 action 编码（`nhp_server_xdp.c`，日志里的 `REASON=`）：

| 条件 | 结果 | 日志 REASON |
| --- | --- | --- |
| TCP/22，src ∈ 白名单 | PASS | `SSH_RELAY` |
| TCP/22，src ∉ 白名单 | DROP | `TCP_SSH_OTHER` |
| TCP/62206 | DROP | `TCP_NHP_PORT` |
| TCP 其它端口 | DROP | `TCP_OTHER` |
| UDP/62206，src ∈ 白名单 | PASS（跳过长度校验） | `NHP_RELAY` |
| UDP/62206，长度 ≥ 240 | PASS | `NHP_DEFAULT` |
| UDP/62206，长度 < 240 | DROP | `UDP_SHORT` |
| UDP 其它端口 | DROP | `UDP_OTHER` |
| 非 TCP/UDP（含 ICMP） | DROP | `NON_TCP_UDP` |
| ARP / IPv6 | PASS | 不上报（避免邻居流量刷屏） |

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
> **默认路由网卡**并把过滤器挂上去，结果是本机除 22（白名单内）与 UDP/62206 外
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
udpserver.go:347:  [Info] server XDP engine loaded
engine_linux.go:   [Info] Start listening for server eBPF events (PERF BUFFER)
config.go:         [Info] xdp relay whitelist applied: [10.99.0.9]
file.go:           [Info] start watching file /tmp/nhpsrv/etc/xdp.toml
```

另开一个终端，确认已 attach：

```bash
sudo ip netns exec nhpxdp ip -details link show veth-n | tail -1
# ... prog/xdp id 305 name xdp_server_prog tag 6fbfd255ac4fee2c jited ...
```

**流量矩阵**（在 host 侧，源地址 10.99.0.1，未进白名单）：

```bash
ping -c1 -W2 10.99.0.2                                        # 期望：100% 丢包
timeout 3 bash -c 'exec 3<>/dev/tcp/10.99.0.2/22'   ; echo $? # 期望：124（超时=被丢）
timeout 3 bash -c 'exec 3<>/dev/tcp/10.99.0.2/62206'; echo $? # 期望：124
timeout 3 bash -c 'exec 3<>/dev/tcp/10.99.0.2/443'  ; echo $? # 期望：124
python3 -c "import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*100,('10.99.0.2',62206))"  # <240，丢
python3 -c "import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*300,('10.99.0.2',62206))"  # ≥240，放行
python3 -c "import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*300,('10.99.0.2',53))"     # 非 62206，丢
```

判决日志 `/tmp/nhpsrv/logs/nhp_server_xdp-<date>.log`（实测输出）：

```
20:24:31 test-server [NHP-DROP] REASON=NON_TCP_UDP   SRC=10.99.0.1 DST=10.99.0.2 LEN=84  PROTO=ICMP SPT=0     DPT=0     RELAY=0
20:24:33 test-server [NHP-DROP] REASON=TCP_SSH_OTHER SRC=10.99.0.1 DST=10.99.0.2 LEN=60  PROTO=TCP  SPT=58350 DPT=22    RELAY=0
20:24:36 test-server [NHP-DROP] REASON=TCP_NHP_PORT  SRC=10.99.0.1 DST=10.99.0.2 LEN=60  PROTO=TCP  SPT=46220 DPT=62206 RELAY=0
20:25:26 test-server [NHP-DROP] REASON=TCP_OTHER     SRC=10.99.0.1 DST=10.99.0.2 LEN=60  PROTO=TCP  SPT=53458 DPT=443   RELAY=0
20:24:39 test-server [NHP-DROP] REASON=UDP_SHORT     SRC=10.99.0.1 DST=10.99.0.2 LEN=128 PROTO=UDP  SPT=49355 DPT=62206 RELAY=0
20:24:39 test-server [NHP-PASS] REASON=NHP_DEFAULT   SRC=10.99.0.1 DST=10.99.0.2 LEN=328 PROTO=UDP  SPT=38438 DPT=62206 RELAY=0
20:24:39 test-server [NHP-DROP] REASON=UDP_OTHER     SRC=10.99.0.1 DST=10.99.0.2 LEN=328 PROTO=UDP  SPT=57077 DPT=53    RELAY=0
```

**白名单两条分支**（把 10.99.0.1 加进白名单后重测）：

```bash
sed -i 's/10.99.0.9/10.99.0.1/' /tmp/nhpsrv/etc/xdp.toml     # 热更新，无需重启
sleep 1
timeout 3 bash -c 'exec 3<>/dev/tcp/10.99.0.2/22'            # 期望：Connection refused（瞬时）——包被放行，内核回 RST
python3 -c "import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*40,('10.99.0.2',62206))"  # 40 字节也放行
```

```
20:25:13 test-server [NHP-PASS] REASON=SSH_RELAY SRC=10.99.0.1 ... DPT=22    RELAY=1
20:25:13 test-server [NHP-PASS] REASON=NHP_RELAY SRC=10.99.0.1 ... DPT=62206 RELAY=1
```

> 「超时（rc=124）」= 被 XDP 丢；「Connection refused（毫秒级）」= 被放行、内核回了
> RST。这一对差异就是后面线上 before/after 对比要用的信号。

**热更新与 map 内容**（`bpftool` 要进守护进程的 mount namespace）：

```bash
PID=$(pgrep -x nhp-serverd)
sudo nsenter -t $PID -m -n bpftool map dump pinned /sys/fs/bpf/nhp_relay_ips
# [{"key": 16802570,"value": 1}]     ← key 是 __be32，按小端解读才是 10.99.0.1

# 按 IP 精确查一条（key 的 hex 就是点分四段的十六进制）
IP=10.99.0.1
sudo nsenter -t $PID -m -n bpftool map lookup pinned /sys/fs/bpf/nhp_relay_ips \
  key hex $(echo $IP | tr '.' ' ' | xargs printf '%02x %02x %02x %02x')
```

改 `xdp.toml` 后守护进程日志应出现（实测）：

```
config.go:625:      [Info] xdp config: /tmp/nhpsrv/etc/xdp.toml has been updated
engine_linux.go:656:[Info] relay whitelist applied: 2 address(es) active, 0 removed
config.go:668:      [Info] xdp relay whitelist applied: [10.99.0.1 192.0.2.7]
```

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

流水线写进 `xdp.toml` 的是 `dig relay.opennhp.org` 的解析结果，也就是 relay 的
**公网 EIP**；而 CI 是 `ProxyJump` 经 relay 连 **server 私网地址**的（
`.github/workflows/deploy-demo-v2.yml` 里 `Host nhp-server → HostName <server_private_ip>`），
同一 VPC 内走私网路径，server 看到的源地址是 relay 的**私网地址**。两者必须对上：

```bash
# a) 流水线将写入白名单的地址
dig +short relay.opennhp.org @1.1.1.1               # = $RELAY_PUB

# b) server 实际看到的 SSH 源地址（在 server 上执行，保持另一个 SSH 会话在线）
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV \
  "ss -tn state established '( sport = :22 )' | awk 'NR>1{print \$3, \$4}'"
#    左列是本机 22，右列 Peer Address = 真正需要进白名单的地址

# c) 抓一次 SYN 交叉印证（另开会话时触发）
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV \
  "sudo timeout 20 tcpdump -ni any 'tcp port 22 and tcp[tcpflags] & tcp-syn != 0' -c 3"

# d) relay 转发过来的 NHP 流量源地址
ssh -J ec2-user@$RELAY_PUB ec2-user@$SERVER_PRIV \
  "sudo timeout 20 tcpdump -ni any 'udp port 62206' -c 5"
```

**判定**：若 (b)/(c) 显示的地址不在 (a) 的结果里（例如是 `10.x.x.x` 而 (a) 是公网
EIP），**先不要部署**，先让 `RELAY_IPS` 同时包含私网地址：

```bash
# 期望的 xdp.toml 内容（RelayIPs 是数组，可以多元素）
RelayIPs = ["<relay 公网 EIP>", "<relay 私网 IP>"]
```

（`terraform output -raw relay_private_ip` 即可拿到，见 §7 风险 1。）

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
echo "== listening ==";        sudo ss -lunp | grep 62206
echo "== ebpf log lines ==";   grep -icE 'ebpf|xdp' /home/ec2-user/nhp-server/logs/server-$(date +%F).log 2>/dev/null
EOS
```

部署前期望：`xdp attached` 为 `(none)`，`/sys/fs/bpf/` 里没有 `xdp_server_prog` /
`nhp_relay_ips`，unit 没有 `AmbientCapabilities`，`etc/` 下没有 `xdp.toml` 和
`nhp_server_xdp.o`，`nhp-serverd` 正在监听 UDP/62206。

### 2.3 从 relay（VPC 内）看的基线

XDP 的行为变化主要体现在 VPC 内侧，所以对比要从 relay 上打：

```bash
ssh ec2-user@$RELAY_PUB "bash -s $SERVER_PRIV" <<'EOS' | tee before-from-relay.txt
set +e
SRV=$1
echo "== ping ==";        ping -c2 -W2 $SRV | tail -2
echo "== tcp/22 ==";      timeout 3 bash -c "exec 3<>/dev/tcp/$SRV/22";      echo "rc=$?"
echo "== tcp/443 ==";     timeout 3 bash -c "exec 3<>/dev/tcp/$SRV/443";     echo "rc=$?"
echo "== tcp/62206 ==";   timeout 3 bash -c "exec 3<>/dev/tcp/$SRV/62206";   echo "rc=$?"
echo "== udp/62206 短包 ==" ; python3 -c "
import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*64,('$SRV',62206));print('sent')"
EOS
```

部署前期望：ping 通；tcp/22 `rc=0`（连上）；tcp/443、tcp/62206 是**立即** refused
（`rc=1`，安全组内部互通、端口没人听）；短 UDP 包直达 user space。

### 2.4 从公网看的基线

```bash
nmap -Pn -n -p 22,80,443,62206 $SERVER_PUB      | tee before-from-internet.txt
sudo nmap -Pn -n -sU -p 62206 $SERVER_PUB       | tee -a before-from-internet.txt
ping -c3 $SERVER_PUB                            | tee -a before-from-internet.txt
```

> 公网视角在部署前后**基本不会变**，这是预期的：安全组早就只放行 UDP/62206（22 只对
> relay 安全组开）。XDP 的增量在于 VPC 内侧、以及公网上那些能穿过安全组的
> UDP/62206 噪声包（`nmap -sU` 的空包会从「进 user space」变成「内核丢弃」，外部看
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
| `configure` → `Resolve relay IPs via DNS` | `dig +short relay.opennhp.org @1.1.1.1` 必须**恰好一条** A 记录，否则 `exit 1` | 整个流水线红，**不渲染 / 不 scp / 不重启**，线上继续用旧白名单 |
| `build` → `test -s release/nhp-server/etc/nhp_server_xdp.o` | 目标文件必须编出来 | 红；避免 `.o` 缺失导致线上静默 fail-open |
| `deploy-server` → `Deploy nhp-serverd and plugins` | scp `.o` + `xdp.toml`，安装 `/etc/systemd/system/nhp-serverd.service.d/10-ebpf.conf`（`CAP_BPF CAP_NET_ADMIN CAP_PERFMON` + 准备 `/sys/fs/bpf`），重启 | — |
| `deploy-server` → `Verify the XDP ingress filter attached` | `ip -details link show \| grep -q xdpgeneric`，没挂上就红 | 这是把「fail-open 静默失效」变成「部署失败」的唯一关卡 |

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
grep 'xdp relay whitelist applied' /home/ec2-user/nhp-server/logs/server-$(date +%F).log | tail -1
```

### 4.4 事件日志（判决可观测）

```bash
LOG=/home/ec2-user/nhp-server/logs/nhp_server_xdp-$(date +%F).log
tail -f $LOG
# 按 REASON 汇总，看扫描噪声被挡在哪里
grep -o 'REASON=[A-Z_]*' $LOG | sort | uniq -c | sort -rn
# 确认 relay 的 SSH/转发走的是 PASS 分支
grep -E 'SSH_RELAY|NHP_RELAY' $LOG | tail -5
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
tcp/62206   →  rc=124 超时              （之前：rc=1 立即 refused）
udp/62206 64B → 事件日志 UDP_SHORT DROP（若 relay 在白名单则是 NHP_RELAY PASS）
```

从公网侧另找一台机器（不在白名单内）打一发合法长度的 UDP，确认 NHP 仍然可达：

```bash
python3 -c "
import socket;s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM);s.sendto(b'A'*300,('$SERVER_PUB',62206))"
# server 侧事件日志：REASON=NHP_DEFAULT ... [NHP-PASS]
```

### 4.6 业务回归（最重要的一条）

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
| server 日志 | 无 eBPF 相关行 | `server XDP engine loaded`、`xdp relay whitelist applied: [...]` |
| `logs/nhp_server_xdp-*.log` | 不存在 | 持续写入 PASS/DROP 判决 |
| relay → server ICMP | 通 | 全丢（`NON_TCP_UDP`） |
| relay → server tcp/443 | 立即 refused（rc=1） | 超时（rc=124，`TCP_OTHER`） |
| relay → server tcp/22 | 连上 | 连上（`SSH_RELAY`，前提：白名单正确） |
| 公网 → UDP/62206 <240B | 进 user space 由 `RecvPrecheck` 丢 | 内核丢（`UDP_SHORT`），不消耗用户态 |
| 公网 → UDP/62206 ≥240B | 通 | 通（`NHP_DEFAULT`）——NHP 语义不变 |
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

### 6.2 白名单写错 → 失联后的恢复

**没有破玻璃通道。** 按代价从低到高：

1. **摘盘改文件**（保状态）：停实例 → detach 根卷 → 挂到另一台实例 →
   改 `/home/ec2-user/nhp-server/etc/xdp.toml`（或 `systemctl disable nhp-serverd`）→
   挂回 → 启动。EIP 由 `aws_eip` 关联保留。
2. **重建实例**：`terraform apply -replace=aws_instance.server`。新实例上
   `nhp-serverd` 尚未部署、XDP 未挂载，SSH 暂时开放，CI 重新部署后才上锁。

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

1. **白名单是公网 IP、实际源地址是私网 IP** —— 流水线用 `dig relay.opennhp.org`
   取的是 relay 的公网 EIP，而 CI/relay 经 VPC 内网连 server 私网地址，server 看到的
   是 relay 的**私网** IP。若两者不一致，第一次 `systemctl restart nhp-serverd` 之后
   SSH（含正在跑的那条会话）立即被丢。**首次部署前必须做 §2.1，必要时把
   `RELAY_IPS` 写成「公网 EIP + 私网 IP」两条。**
2. **fail-open 是静默的** —— 内核旧、能力缺失、`.o` 丢失都只打一行 warning。日常巡检
   请把 `ip -details link show | grep -q xdpgeneric` 纳入告警，而不是只看
   `systemctl is-active`。
3. **XDP 无连接跟踪** —— 白名单变更对已建立连接同样生效，改 `xdp.toml` 等同于改防火墙。
4. **`Enabled` / `NhpMinFrameBytes` 只是记录值** —— 前者不会阻止 attach，后者编译进
   `.o`（`NHP_MIN_UDP_LEN`）。改长度下界要重新 `make ebpf-objs` 并重发 `.o`。
5. **事件日志有速率限制** —— 每个 action 每 CPU 每秒 64 条（`NHP_EVENT_BURST`），大流量
   扫描下日志条数少于真实丢包数，计数请用 `bpf_stats` 的 `run_cnt`。
