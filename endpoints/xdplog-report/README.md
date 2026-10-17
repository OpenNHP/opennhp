# xdplog-report

[English](#english) | [中文](#中文)

---

## English

`xdplog-report` turns the XDP event logs written by `nhp-serverd` into an attack
report. It is fully offline: it reads log files and an optional local GeoIP
database, and never resolves or looks up anything over the network.

### Where the logs come from

`nhp-serverd` writes `logs/nhp_server_xdp-YYYY-MM-DD.log` next to its executable
(the XDP filter's per-packet and per-minute lines). Copy the files off the host
over SSH, then run this tool on your workstation.

Each line has this shape:

```
HH:MM:SS <server-id> [NHP-PASS|NHP-DROP] REASON=... SRC=... DST=... LEN=... PROTO=... SPT=... DPT=... RELAY=0
HH:MM:SS <server-id> [NHP-STAT] VERDICT=... REASON=... WINDOW=60s PKTS=... LOGGED=... TOTAL=...
HH:MM:SS <server-id> [NHP-STAT] EVENT_LOG_BUDGET_EXHAUSTED BYTES=...
HH:MM:SS <server-id> [NHP-SCAN] WINDOW=300s KIND=TOTAL|SRC|PORT ...
```

`<server-id>` is the `Hostname` from the daemon's `config.toml`, not the OS
hostname. It identifies the host when logs from several servers are collected
together.

### Quick start

```bash
cd endpoints

# Markdown to stdout
go run ./xdplog-report -in ~/xdp-logs/nhp_server_xdp-2026-10-16.log

# Standalone HTML with GeoIP, 20 rows per table
go run ./xdplog-report \
  -in ~/xdp-logs/ \
  -geo-mmdb ~/GeoLite2-City.mmdb \
  -top 20 -format html \
  -out attack-report.html
```

Build a binary once instead of using `go run`:

```bash
cd endpoints
go build -o bin/xdplog-report ./xdplog-report
./bin/xdplog-report -in ~/xdp-logs/ -format md
```

> Use `-o` when building. `go build ./xdplog-report` from the `endpoints`
> directory refuses to overwrite the `xdplog-report/` source directory with a
> binary of the same name.

### Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-in` | required | Log file, glob, or directory. Repeatable. `.gz` is supported. A directory input picks up `nhp_server_xdp*.log` / `.log.gz` only. |
| `-since` | empty | First day to include, `YYYY-MM-DD` (inclusive). |
| `-until` | empty | Last day to include, `YYYY-MM-DD` (inclusive). |
| `-date` | empty | Day to assume for a file whose name has no `YYYY-MM-DD`. |
| `-top` | `10` | Rows in each TOP table. |
| `-format` | `md` | `md`, `json`, `csv`, or `html`. |
| `-out` | stdout | Output file. The directory must already exist. |
| `-geo-csv` | empty | Optional `cidr,country[,city]` CSV for the region section. |
| `-geo-mmdb` | empty | Optional MaxMind GeoLite2-City `.mmdb` for the region section. Mutually exclusive with `-geo-csv`. |
| `-mask-ip` | `false` | Truncate source addresses for sharing: IPv4 `a.b.c.x`, IPv6 first 48 bits. |

Dates come from file names. The log itself has no date per line, so files must
be given (or discovered) in chronological order for restart detection to work.

### Reading the report

- **Attack** means a `[NHP-DROP]` line. `[NHP-PASS]` is allowed traffic and is
  shown only as a baseline.
- **Exact packet totals** come from the per-minute `[NHP-STAT]` lines. `PKTS`
  deltas are summed and duplicate STAT lines are ignored. This is the number to
  trust.
- **Sampled events** come from per-packet lines. The filter rate-limits them per
  class, so bulk classes may have few or none. IP, region, port and protocol
  rankings are samples and are not comparable in volume across classes.
- `[NHP-SCAN]` lines are scan-detector summaries emitted by newer builds. They
  are parsed and counted under "scan-detector summary lines", but they are not
  DROP events and never enter the attack totals.
- A **budget-exhausted** warning means per-packet lines were suppressed for the
  rest of that day. Exact totals stay correct; rankings after that point are
  incomplete.
- **Restarts** are detected when a STAT `TOTAL` counter goes backwards. One real
  restart resets every counter, so the report shows a single entry rather than
  one per reason.
- **Unknown reasons** mean the classifier is older than the deployed XDP build.
  Add the reason to `dropCategories` in `classify.go`, then run the tests.

### Output formats

- `md`: Markdown report, suitable for a ticket or terminal.
- `json`: Machine-readable report, `schema: 1`. Includes `bad_lines`,
  `scan_summary_lines`, `restarts_detected`, and `unknown_reasons`.
- `csv`: One flat `section,key,pkts_exact,sampled_events,share,extra` table.
- `html`: Standalone page with inline CSS, no scripts and no external assets.
  Safe to attach to a ticket or open offline.

### GeoIP

Both geo sources are local files. `-geo-mmdb` accepts a MaxMind GeoLite2-City
database; note that GeoLite2 is subject to MaxMind's license. Private, loopback,
link-local, multicast and unspecified addresses are never looked up. Without a
geo source the region section says "Not enabled" and no lookups happen.

### Tests

```bash
cd endpoints
go test ./xdplog-report/
```

The fixtures under `testdata/` are hand-written synthetic logs, not captured
from a live host. They are exempted from the repository-wide `*.log` ignore rule
so the tests work on a fresh checkout.

### Adding a new DROP reason

1. Add the reason to `dropCategories` in `classify.go` with a category.
2. If the checked-in `nhp/utils/ebpf/engine_linux.go` also defines it, the
   cross-check test will confirm coverage. The classifier may know reasons the
   checked-in engine does not: a deployed build can lead the branch (for example
   `TCP_SYN_SCAN`).

---

## 中文

`xdplog-report` 把 `nhp-serverd` 写出的 XDP 事件日志分析成攻击报告。工具完全
离线运行：只读取日志文件和可选的本地 GeoIP 数据库，不做任何网络解析或查询。

### 日志来源

`nhp-serverd` 会在可执行文件所在目录下写 `logs/nhp_server_xdp-YYYY-MM-DD.log`，
内容是 XDP 过滤器的逐包事件和每分钟统计。通过 SSH 把文件拷回工作站后运行本工具。

每行的格式如下：

```
HH:MM:SS <server-id> [NHP-PASS|NHP-DROP] REASON=... SRC=... DST=... LEN=... PROTO=... SPT=... DPT=... RELAY=0
HH:MM:SS <server-id> [NHP-STAT] VERDICT=... REASON=... WINDOW=60s PKTS=... LOGGED=... TOTAL=...
HH:MM:SS <server-id> [NHP-STAT] EVENT_LOG_BUDGET_EXHAUSTED BYTES=...
HH:MM:SS <server-id> [NHP-SCAN] WINDOW=300s KIND=TOTAL|SRC|PORT ...
```

`<server-id>` 是 daemon `config.toml` 里的 `Hostname`，不是操作系统的 hostname。
多台服务器的日志汇总到一起时，用它区分来源。

### 快速开始

```bash
cd endpoints

# 输出 Markdown 到终端
go run ./xdplog-report -in ~/xdp-logs/nhp_server_xdp-2026-10-16.log

# 带 GeoIP 的独立 HTML，每张表 20 行
go run ./xdplog-report \
  -in ~/xdp-logs/ \
  -geo-mmdb ~/GeoLite2-City.mmdb \
  -top 20 -format html \
  -out attack-report.html
```

也可以先编译成二进制：

```bash
cd endpoints
go build -o bin/xdplog-report ./xdplog-report
./bin/xdplog-report -in ~/xdp-logs/ -format md
```

> 编译时请使用 `-o`。在 `endpoints` 目录下直接执行 `go build ./xdplog-report`
> 会尝试生成同名二进制，与源码目录 `xdplog-report/` 冲突。

### 参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-in` | 必填 | 日志文件、glob 或目录，可重复。支持 `.gz`。目录输入只匹配 `nhp_server_xdp*.log` / `.log.gz`。 |
| `-since` | 空 | 起始日期，`YYYY-MM-DD`（含当天）。 |
| `-until` | 空 | 结束日期，`YYYY-MM-DD`（含当天）。 |
| `-date` | 空 | 文件名里没有 `YYYY-MM-DD` 时假定的日期。 |
| `-top` | `10` | 每张 TOP 表的行数。 |
| `-format` | `md` | `md`、`json`、`csv` 或 `html`。 |
| `-out` | stdout | 输出文件。所在目录必须已存在。 |
| `-geo-csv` | 空 | 可选的 `cidr,country[,city]` CSV，用于地区统计。 |
| `-geo-mmdb` | 空 | 可选的 MaxMind GeoLite2-City `.mmdb`，用于地区统计。与 `-geo-csv` 互斥。 |
| `-mask-ip` | `false` | 对外分享时截断源地址：IPv4 为 `a.b.c.x`，IPv6 保留前 48 位。 |

日期来自文件名。日志每行本身没有日期，所以文件必须按时间顺序输入（或被发现），
否则重启检测会不准。

### 如何理解报告

- **攻击**指 `[NHP-DROP]` 行。`[NHP-PASS]` 是允许的流量，只作为基线展示。
- **精确包数**来自每分钟的 `[NHP-STAT]` 行：累加 `PKTS` 增量，重复的 STAT 行会
  被忽略。这是最可信的数字。
- **采样事件数**来自逐包行。过滤器按类别限速，因此大流量类别可能只有很少的逐包
  行。IP、地区、端口、协议排行都是采样结果，不同类别之间不能直接比较数量。
- `[NHP-SCAN]` 是较新版本输出的扫描检测摘要。工具会解析并计入
  “scan-detector summary lines”，但它们不是 DROP 事件，不计入攻击总量。
- **budget exhausted** 警告表示当天剩余时间的逐包行被抑制。精确总量仍然正确，
  但该时间点之后的排行不完整。
- 某个 STAT `TOTAL` 计数器变小时判定为**重启**。一次真实重启会重置所有计数器，
  因此报告只显示一条记录，而不是每个 reason 一条。
- **Unknown reasons** 表示分类器落后于线上部署的 XDP 版本。在 `classify.go` 的
  `dropCategories` 中补上该 reason，然后运行测试。

### 输出格式

- `md`：Markdown 报告，适合贴到工单或在终端查看。
- `json`：机器可读，`schema: 1`，包含 `bad_lines`、`scan_summary_lines`、
  `restarts_detected`、`unknown_reasons`。
- `csv`：扁平表，列为 `section,key,pkts_exact,sampled_events,share,extra`。
- `html`：独立页面，内联 CSS，无脚本、无外部资源，可离线打开或作为附件分享。

### GeoIP

两种地理数据源都是本地文件。`-geo-mmdb` 接受 MaxMind GeoLite2-City 数据库；
GeoLite2 的使用受 MaxMind 许可证约束。私有、回环、链路本地、组播和未指定地址不会
做归属查询。不提供地理数据源时，地区部分显示 “Not enabled”，且不会做任何查询。

### 测试

```bash
cd endpoints
go test ./xdplog-report/
```

`testdata/` 下的 fixture 是手写的合成日志，不是从线上主机抓取的。它们从仓库根目录
的 `*.log` 忽略规则中豁免，保证全新 checkout 也能跑测试。

### 新增 DROP reason

1. 在 `classify.go` 的 `dropCategories` 中加入该 reason 和对应类别。
2. 如果当前分支的 `nhp/utils/ebpf/engine_linux.go` 也定义了它，交叉校验测试会确认
   覆盖完整。分类器可以包含引擎尚未定义的 reason：线上部署的版本可能领先于分支
   （例如 `TCP_SYN_SCAN`）。
