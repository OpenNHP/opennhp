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

### `site` subcommand: a public daily stats site

`xdplog-report site` builds a multi-day static site for `stats.opennhp.org`
(or any nginx-served directory). It scans a directory of pulled logs,
analyses each day, and writes a per-day HTML page plus an index. Days the
collector could not fetch are skipped silently.

```bash
go run ./xdplog-report site \
  -logs  /home/ec2-user/xdpstats/raw \
  -out   /var/www/stats \
  -state /home/ec2-user/xdpstats/state.json \
  -top   20
```

Layout under `-logs`:

```
<logs>/
  server/
    nhp_server_xdp-2026-10-15.log
    nhp_server_xdp-2026-10-16.log.gz
    ...
```

Layout under `-out`:

```
<out>/
  index.html
  index.json
  countup.js                # self-hosted homepage counter (see below)
  totalnumber.js            # the total as one JS constant (see below)
  d/server/2026-10-16.html  # one per day
  d/server/2026-10-16.json
```

The homepage's `<h1>` is **Invisible Server, Visible Attacks** and it opens with a
hero line: the all-time **total attack packets blocked**, summed over every
row of `index.json`, with a `Since YYYY-MM-DD` line naming the earliest day
still published. Both are derived from the index, not stored: an index reader
can reproduce the number by summing `drop_pkts`. A day whose raw log has
already aged out of the collector's cache still counts as long as its
`d/<source>/<day>.json` is on disk.

The `<title>` is a different string — **Zero Trust in Action: Live Attacks on
OpenNHP Servers** — because a browser tab, a search result and a chat preview
have room to say what the site is *about*, while the heading above the fold
does not. The same `<head>` carries a meta description, a canonical URL and the
Open Graph / Twitter card tags (`og:title` repeats the `<h1>`). Those tag values
are the only absolute URLs in the output, and they are the one documented
exception to the site's self-containment: a crawler reads them, the browser
never fetches them, so they point at `https://stats.opennhp.org/` — our own
origin — rather than at a third party. The two scripts the page loads are
`countup.js` and `totalnumber.js`, both same-origin and both written here, so
the homepage can quote its own total as `NHP_TOTAL_ATTACK_BLOCKED`. Nothing may
load off-host; `assertSelfContained` in `site_test.go` fails the build if it
does. `og:image` points at `/og.png`, which has to exist in the web root for
social previews to render anything.

Days before `-since` (**default `2026-10-16`**) are not published at all: no
detail page, no index row, and no contribution to the total. The XDP filter
was attached to the demo host after those logs were written, so they are a
partial record; counting them would understate the totals by an amount nobody
can audit. Pass `-since ""` to publish everything (local debugging), or any
other `YYYY-MM-DD` to move the floor. A `-since` that would exclude every
discovered day is refused rather than published, so a typo can never blank a
populated index.

"Not published" is enforced, not just skipped. A site that already served a
day below the floor — as stats.opennhp.org did before the floor existed —
has that day's `d/<source>/<day>.{html,json}` **deleted** on the next run, and
its `state.json` entry dropped. Nothing links to those pages any more, but a
file left in the web root is still reachable at its old URL and in search
caches, which is exactly what the floor is meant to prevent. Deleting them
also marks the index dirty, so the row leaves `index.json` and the hero total
in the same run. This is the only thing the site subcommand deletes; the raw
logs and the collector cache are never touched. A run that cannot remove a
page still writes the index and state, then fails, so the leftover page is
loud rather than silent.

Those deletions happen only once every publishable day has been processed
successfully, and immediately before the index that stops linking them is
written — which is written before `state.json`, so a run that gives up on any
other day (a log it cannot analyze, an index it refuses to blank) leaves the
site it was serving exactly as it found it, links and hero total unchanged,
rather than removing pages the live index still points at. The next successful
run then does the deletions.

### Template changes reach published pages

`state.json` records, per day, the `siteRenderVersion` that produced its page.
Bumping that constant in `site.go` re-renders every already-published day on
the next run, so editing `render_html.go` (a new title, a new subtitle) does
not need an operator to pass `-force` or delete `state.json` — without it, only
the days whose raw log happens to change that week would ever be rewritten.

Days whose raw log the collector has already pruned cannot be re-analyzed,
but their HTML is a pure function of the `d/<source>/<day>.json` published
beside them, so a version bump re-renders those too, from that JSON. Their
numbers are never recomputed.

A pruned day whose published JSON has gone missing or stopped parsing drops
out of the index — there is nothing left to publish — but one whose JSON is
fine and whose *page* cannot be written keeps its row and its share of the
total, and the run fails instead. Losing a day to a transient `EACCES` or
`ENOSPC` on the web root would make `NHP_TOTAL_ATTACK_BLOCKED` fall and then
recover on the next good run, which reads as an attack surge nobody had.

`index.html`'s Source column renders `server` as **nhp-server**, so the column
names the product instead of repeating a generic noun. `index.json` keeps the
raw directory name in `source` — that field is part of the published contract
and anyone already parsing it depends on the value.

The hero number animates (2.5s count-up, thousands separators) using
`countup.js`, written next to `index.html` and loaded with a plain
`<script src="countup.js" defer>`. It is progressive enhancement: the span
already contains the real formatted total, so with JavaScript off, blocked by
a stricter older CSP, or under `prefers-reduced-motion`, the number reads as
static text. The site loads no third-party resource — the nginx vhost's CSP is
`script-src 'self'`, and countup.js is the only script it allows.

### `totalnumber.js`: embedding the total on another site

`totalnumber.js` publishes the same all-time total as a single JavaScript
constant, for a page elsewhere that wants the headline figure and nothing
else:

```html
<script src="https://stats.opennhp.org/totalnumber.js"></script>
<script>console.log(NHP_TOTAL_ATTACK_BLOCKED);</script>
```

It is a sibling of `index.json`, not a replacement: `index.json` stays
authoritative (one row per day, total reproducible by summing `drop_pkts`),
while this file is the one-number answer. The value is a pure function of the
summaries, so a rebuild with no new data rewrites a byte-identical file. On an
empty site it is still written, exporting `0` with no counting-start comment,
so an embedding page never gets a 404 and a dangling reference. As a classic
script it needs no CORS header — the *importing* page's CSP governs it.

`state.json` (the incremental build state) goes wherever `-state` says and
**not** into `-out` by default on the demo deployment. It records the
absolute path of every input file, which is host layout an anonymous visitor
to a public site has no business reading; the deploy passes a private path
and the nginx vhost refuses `/state.json` as a second line of defence.
Delete it to force a full rebuild (or pass `-force`).

`index.json` is a JSON array of summaries, sorted by day descending:

```json
[
  {
    "source": "server",
    "day": "2026-10-16",
    "partial": false,
    "drop_pkts": 1204,
    "pass_pkts": 18,
    "drop_share": 0.9853,
    "unique_attack_ips": 6,
    "peak_hour": "2026-10-16 18:04:00",
    "restarts_detected": 0,
    "budget_exhausted": false,
    "unknown_reasons": false,
    "top_category": "TCP scan / probe"
  }
]
```

The per-day JSON files reuse the single-file `Report` (`schema: 1`), so the
existing `xdplog-report -in <one day>.json -format html` path can regenerate a
page from any archived day. Each page is titled **Daily Attack Report:**
followed by the day (or the day range, for a multi-day CLI report), and opens
with a sentence built from that day's own exact counters — packets dropped,
unique attack sources, and the DROP share of all inbound traffic.

Dates are UTC (server log rotation uses `time.Now()` in UTC by default). Day
files written today are marked `partial: true` in the index; they are still
published and re-analysed on every run.

**Masking**: `-mask-ip` defaults to **`true`** in `site` mode (different from
the single-file CLI). Public-facing pages truncate source IPv4 to `a.b.c.x`
and IPv6 to the first 48 bits. The `-mask-ip=false` knob is there for local
debugging, not for production.

**Idempotency**: each (source, day) pair is fingerprinted by `size + mtime`
plus the render version. Unchanged days are not re-analysed and not
re-written. Index/state files are only rewritten when at least one day was
rebuilt or a page was unpublished.

**`:-force`**: pass `-force` to rebuild every day regardless of the
fingerprint, e.g. after updating `classify.go`. A change to `render_html.go`
does not need it — bump `siteRenderVersion` instead.

### Site subcommand flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-logs` | required | Directory of `<source>/nhp_server_xdp-YYYY-MM-DD.log[.gz]`. The collector script stages logs into this layout. |
| `-out` | required | Web root directory. Created if missing. Existing files are kept unless rewritten, except the detail pages of days the `-since` floor excludes, which are deleted (see above). |
| `-state` | `<out>/state.json` | Incremental-build state. Point it OUT of the web root on a public site: it records the absolute path of every input file. |
| `-top` | `20` | Rows in each TOP table on the per-day pages. |
| `-geo-csv` | empty | Optional `cidr,country[,city]` CSV. Mutually exclusive with `-geo-mmdb`. |
| `-geo-mmdb` | empty | Optional MaxMind GeoLite2-City `.mmdb`. |
| `-mask-ip` | **`true`** | Truncate source addresses for the public site. Different default than the single-file CLI. |
| `-since` | `2026-10-16` | Earliest `YYYY-MM-DD` day to publish. Older days are neither rendered nor counted, and any detail page already published for one is deleted. Empty publishes everything. |
| `-force` | `false` | Rebuild every day even if inputs are unchanged. Not needed to push a template change out: bumping `siteRenderVersion` does that. |

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
   checked-in engine does not: a deployed build can lead the branch.

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

### `site` 子命令：每日攻击统计静态站

`xdplog-report site` 把多天的日志一次性渲染成可由 nginx 直接托管的静态站点
（`stats.opennhp.org` 或任意目录）。它扫描已拉取的日志目录，逐天分析，
写出每天一个 HTML 详情页加一个总首页。无法取到日期的文件会被静默忽略。

```bash
go run ./xdplog-report site \
  -logs  /home/ec2-user/xdpstats/raw \
  -out   /var/www/stats \
  -state /home/ec2-user/xdpstats/state.json \
  -top   20
```

`-logs` 目录布局：

```
<logs>/
  server/
    nhp_server_xdp-2026-10-15.log
    nhp_server_xdp-2026-10-16.log.gz
    ...
```

`-out` 目录布局：

```
<out>/
  index.html
  index.json
  countup.js
  totalnumber.js          # 累计总数，导出为一个 JS 常量
  d/server/2026-10-16.html  # 每天一个
  d/server/2026-10-16.json
```

首页的 `<h1>` 是 **Invisible Server, Visible Attacks**，标题下方先给出一个 hero 大数字：
**累计被拦截的攻击数据包数**，等于 `index.json` 所有行的 `drop_pkts` 之和；再下一行小字
`Since YYYY-MM-DD` 标出统计起点（最早发布的那一天）。两者都是从索引现算的，不是另外存的：
读者对 `drop_pkts` 求和就能复现这个数字。某天的原始日志即使已经过期滚出采集缓存，只要
`d/<source>/<day>.json` 还在磁盘上就仍然计入。

`<title>` 用的则是另一句 —— **Zero Trust in Action: Live Attacks on OpenNHP
Servers**。浏览器标签页、搜索结果和聊天预览有空间说清这个站"是关于什么的"，而首屏大标题
没有；两处各司其职。同一个 `<head>` 里还有 description、canonical，以及 Open Graph /
Twitter 卡片标签（`og:title` 重复 `<h1>` 的那句）。这些标签里的绝对 URL 是全站唯一一处
"不自包含"，而且是写进文档的例外项：爬虫会读它们，浏览器不会去取，所以它们指向的是
`https://stats.opennhp.org/` 这个**自己的**源站，而不是任何第三方。页面加载的两个脚本
`countup.js` 和 `totalnumber.js` 都是同源、都在这里生成，所以首页自己也能用
`NHP_TOTAL_ATTACK_BLOCKED` 引用同一个总数。任何东西都不允许从站外加载 —— 一旦有，
`site_test.go` 里的 `assertSelfContained` 会让构建失败。`og:image` 指向 `/og.png`，
这个文件必须真实存在于 web 根目录，社交预览才有图可显示。

`-since`（**默认 `2026-10-16`**）之前的数据完全不发布：不生成详情页、不进索引、也不计入
总数。XDP 过滤器是这些日志写完之后才挂到 demo 主机上的，所以那几天的记录并不完整；把它们
算进去会让总数偏小，而且偏了多少没人能复核。传 `-since ""` 可以全部发布（本地调试用），
也可以传其它 `YYYY-MM-DD` 调整起点。如果某个 `-since` 会把所有天都排除掉，命令会直接报错
而不是发布一个空索引 —— 打错一个日期不该把已经有数据的首页清空。

"不发布"是强制的，不只是跳过。对已经发布过 floor 之前那几天的站点（stats.opennhp.org 在
加入 floor 之前就是这样），下一次运行会把该天的 `d/<source>/<day>.{html,json}` **删掉**，
并清掉 `state.json` 里的对应条目。这些页面虽然已经没有任何链接指向，但留在 web 根目录里仍然
可以通过旧 URL（以及搜索引擎缓存）访问，而这正是 floor 要避免的。删除同样会把索引标记为
dirty，因此该行会在同一次运行里从 `index.json` 和首页总数中消失。这也是 `site` 子命令唯一
的删除行为：原始日志和采集缓存一律不动。某个页面删不掉时，索引和状态照常写完，但整次运行
仍然返回失败 —— 残留页面必须是有声音的，而不是悄无声息。

### 模板改动会覆盖已发布的页面

`state.json` 会为每一天记录生成该页面的 `siteRenderVersion`。在 `site.go` 里把这个常量 +1，
下一次运行就会重渲染所有已发布的日子，所以改 `render_html.go`（换标题、加副标题）不需要
运维去加 `-force` 或删 `state.json`：不加的话，只有当周原始日志恰好发生变化的那几天才会被
重写，其余页面会永远停留在旧标题上。

原始日志已被采集脚本清掉的日子无法重新分析，但它们的 HTML 是同目录下
`d/<source>/<day>.json` 的纯函数，所以版本号 +1 时同样会基于该 JSON 重渲染，数值则一律不
重新计算。

首页表格的 Source 列把 `server` 显示成 **nhp-server**，让这一列说明产品名而不是重复一个
泛称。`index.json` 的 `source` 字段保持目录原名 —— 它属于已发布的契约，已有解析代码依赖
这个值。

`totalnumber.js` 把同一个累计总数导出成一个 JavaScript 常量，供别的站点直接引用：

```html
<script src="https://stats.opennhp.org/totalnumber.js"></script>
<script>console.log(NHP_TOTAL_ATTACK_BLOCKED);</script>
```

它是 `index.json` 的补充而不是替代：`index.json` 依然是权威数据（每天一行，总数可通过对
`drop_pkts` 求和复现），这个文件只回答"总数是多少"这一个问题。内容是 summaries 的纯函数，
所以没有新数据时重跑会写出逐字节相同的文件。站点为空时它照样生成，导出 `0` 且不带统计起点
注释，引用方不会拿到 404 然后留下一个悬空引用。它是普通 classic script，跨站引用不需要
CORS 头 —— 受引用方自己的 CSP 约束。

`state.json`（内部增量状态）写到 `-state` 指定的位置。demo 部署把它放在
`-out` 之外：它记录了每个输入文件的绝对路径，属于主机布局，不该出现在公开站点上；
部署脚本传入私有路径，nginx vhost 另外再用 `location = /state.json { return 404; }`
挡一道。删除该文件（或加 `-force`）即可触发全量重算。

`index.json` 是一个按日期倒序排列的摘要数组：

```json
[
  {
    "source": "server",
    "day": "2026-10-16",
    "partial": false,
    "drop_pkts": 1204,
    "pass_pkts": 18,
    "drop_share": 0.9853,
    "unique_attack_ips": 6,
    "peak_hour": "2026-10-16 18:04:00",
    "restarts_detected": 0,
    "budget_exhausted": false,
    "unknown_reasons": false,
    "top_category": "TCP scan / probe"
  }
]
```

每天的 JSON 文件复用了单文件 `Report`（`schema: 1`），因此已有的
`xdplog-report -in <某天>.json -format html` 路径可以从任何归档的日期重新生成页面。

日期是 UTC（server 日志轮转默认用 `time.Now()` UTC）。当天还在写的日志在首页标记为
`partial: true`，仍然会被发布并在每次运行时重新分析。

**脱敏**：`-mask-ip` 在 `site` 模式下**默认 `true`**（与单文件 CLI 不同）。
源 IPv4 截断为 `a.b.c.x`，IPv6 保留前 48 位。`-mask-ip=false` 仅供本地排查，生产
不要关。

**幂等性**：每个 (source, day) 用 `size + mtime` + 渲染版本号做指纹。未变化的天不会重新
分析、不会重新写。首页和 state.json 只在至少有一天被重建、或某个页面被取消发布时才重写。

**强制重算**：分类器（`classify.go`）升级后，可用 `-force` 重算所有日期。改
`render_html.go` 则不需要它 —— 把 `siteRenderVersion` +1 即可。

### site 子命令参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-logs` | 必填 | 形如 `<source>/nhp_server_xdp-YYYY-MM-DD.log[.gz]` 的目录，采集脚本负责布局。 |
| `-out` | 必填 | 站点根目录，不存在则创建；未被重写的文件保留不动，只有被 `-since` 排除的那些天的详情页会被删除（见上文）。 |
| `-state` | `<out>/state.json` | 增量构建状态。公开站点请把它指到 web root 之外：它记录了每个输入文件的绝对路径。 |
| `-top` | `20` | 每天详情页 TOP 表的行数。 |
| `-geo-csv` | 空 | 可选的 `cidr,country[,city]` CSV。与 `-geo-mmdb` 互斥。 |
| `-geo-mmdb` | 空 | 可选的 MaxMind GeoLite2-City `.mmdb`。 |
| `-mask-ip` | **`true`** | 对外站点做 IP 截断；默认值与单文件 CLI 不同。 |
| `-since` | `2026-10-16` | 最早发布的 `YYYY-MM-DD` 日期。更早的天既不渲染也不计入总数，已经为这些天发布过的详情页会被删除；传空则全部发布。 |
| `-force` | `false` | 即便输入未变化也强制重算每一天。推送模板改动不需要它：`siteRenderVersion` +1 即可。 |

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
   覆盖完整。分类器可以包含引擎尚未定义的 reason：线上部署的版本可能领先于分支。
