---
layout: page
title: Attack Stats
nav_order: 14
permalink: /attack-stats/
description: How the nhp-server XDP ingress filter's dropped packets become the daily summaries published at stats.opennhp.org, and what each published number means.
---

# How the XDP Filter and the Attack Stats Site Fit Together
{: .fs-9 }

OpenNHP's demo server runs with almost every port exposed to the internet on purpose, so that real scans and attacks reach its network card and can be measured. A kernel-level XDP program drops everything that is not a legitimate NHP knock, and the drops are written to a daily log. A collector on a second host pulls that log back, analyzes it offline, and publishes the result as a static site. This article is about the seam between the two halves: what the filter promises the site, what the site does with it, and what each published figure actually means.
{: .fs-6 .fw-300 }


---

Everything here maps to a specific file in the repository, and the source path is given at the end of every chapter. The code is the authority: where this guide quotes a file, the quote is verbatim, and if the two ever disagree, the source wins.

The published site is <https://stats.opennhp.org/>. This article never quotes a number from it: those change every day, and the machine-readable `index.json` is the only authoritative version of them.

## Contents

1. [Why the demo runs with every port open](#1-why-the-demo-runs-with-every-port-open)
2. [The contract between the filter and the site](#2-the-contract-between-the-filter-and-the-site)
3. [The log the filter writes](#3-the-log-the-filter-writes)
4. [From a dropped packet to a published row](#4-from-a-dropped-packet-to-a-published-row)
5. [What the published numbers mean](#5-what-the-published-numbers-mean)
6. [The published artifacts and their contracts](#6-the-published-artifacts-and-their-contracts)
7. [What is published, and what is not](#7-what-is-published-and-what-is-not)
8. [Running and operating it](#8-running-and-operating-it)
9. [Where to look next](#9-where-to-look-next)

## 1. Why the demo runs with every port open

A security group is the cheapest way to keep a host quiet, and it is also the cheapest way to make a host invisible to measurement. A packet the security group drops never reaches the network interface, so a filter attached to that interface never sees it, never counts it, and never records a reason for it. Under a narrow security group, the overwhelming majority of the internet's background noise disappears before the filter is even invoked, and the statistics site has nothing to show.

The demo therefore has two postures, and both are opt-in:

| Terraform variable | Default | What it opens |
| --- | --- | --- |
| `server_open_all_ports` | `false` | Every tcp port except 22, every udp port, and icmp, to `0.0.0.0/0` |
| `ac_open_all_ports` | `false` | The same, for the nhp-ac's security group |

With the wide posture in effect, the XDP filter attached to the ingress interface is the only inbound control on everything except tcp/22. That is deliberate: tcp/22 stays restricted to the relay's security group at the network layer, so SSH never depends on the filter, whose link is not pinned and which is fail-open by design.

Two consequences follow, and both matter for reading the published numbers:

- **The filter's blind spots are the site's blind spots.** Whatever the filter cannot judge — a filter that is not attached, a reason the program has no class for — is either absent from the log or reported as an unknown reason. The site publishes what the log contains; it does not see past it.
- **The site's numbers describe one demo host.** They are a sample of how the filter behaves in the noise of the real internet. They are not a throughput measurement, a benchmark, or a claim about OpenNHP's performance.

The wide posture is also not free. The filter is detached on every deploy, every restart, every crash, every reboot, and every failed load, and while it is detached the host is open on every port it was opened on. The published figures cannot tell the difference between "the filter saw nothing" and "the filter was not there", which is why a missing hour in the published data is never read as a quiet hour.

### Source locations

- `terraform/demo/variables.tf` — `server_open_all_ports`, `ac_open_all_ports`.
- `terraform/demo/RUNBOOK.md` — sections *All-ports exposure mode* and *All-ports exposure mode on the AC*; the fail-open windows are enumerated there.
- `nhp/ebpf/xdp/nhp_server_xdp.c` — the program attached to the nhp-server's ingress interface.
- `nhp/utils/ebpf/engine_linux.go` — `dropLoaderPrivileges`, the fail-open load path and the attach path.

## 2. The contract between the filter and the site

The XDP program parses no NHP protocol. It does not know what a knock is. It decides on packet shape — address family, protocol, ports, frame length, and whether the source appears in the SSH whitelist — and it answers with three things:

1. a **verdict**, `PASS` or `DROP`;
2. a **reason code**, a stable uppercase name;
3. a **count**, per class, per window.

Identity is decided elsewhere, in user space, by the Noise handshake. The filter's job is narrower: to keep packets that cannot belong to the protocol from existing at all.

The reason vocabulary is the contract. The naming authority for the nhp-server is `serverActionName` in `nhp/utils/ebpf/engine_linux.go`, and it is the only place those names are defined. The site consumes them through `dropCategories` in `endpoints/xdplog-report/classify.go`, and a cross-check test (`TestClassifierMatchesEngine` in `endpoints/xdplog-report/report_test.go`) reads the engine's source and fails if the site does not know a reason the filter can emit. The two lists cannot drift apart silently.

The nhp-server's PASS reasons:

| Reason | What it means |
| --- | --- |
| `SSH_RELAY` | An SSH connection from a whitelisted relay address |
| `NHP_RELAY` | Traffic relayed for an established NHP session |
| `NHP_DEFAULT` | Traffic on the daemon's own default path |
| `TCP_ESTABLISHED` | A reply to a connection the host itself opened |
| `UDP_ESTABLISHED` | A reply to a connection the host itself opened |
| `ICMP_FRAG_NEEDED` | An ICMP fragment-needed reply to a host-opened flow |
| `DHCP_CLIENT` | The DHCP client's own exchange (67 → 68) |
| `NTP_CLIENT` | The NTP client's own exchange (123 → 123) |
| `IPV4_FRAGMENT` | A non-first IPv4 fragment of an allowed flow |
| `ICMPV6_CONTROL` | An IPv6 control message |
| `V6_ESTABLISHED` | A reply to a v6 connection the host itself opened |
| `IPV6_FRAGMENT` | A non-first IPv6 fragment of an allowed flow |

The nhp-server's DROP reasons, and the category the site files each under:

| Reason | Category | What it means |
| --- | --- | --- |
| `TCP_SSH_OTHER` | SSH probe (non-whitelisted source) | An SSH connection from a source that is not in the whitelist |
| `TCP_NHP_PORT` | TCP to NHP knock port | TCP aimed at the UDP knock port, or the daemon's own port |
| `TCP_OTHER` | TCP scan / probe | Any other inbound TCP |
| `TCP_SYN_SCAN` | TCP SYN scan | A bare SYN, with no established flow behind it |
| `TCP_STEALTH_SCAN` | TCP scan / probe | A scan technique the program recognizes separately from a bare SYN |
| `UDP_OTHER` | UDP scan / non-NHP UDP | UDP that is not a knock and not a reply to the host's own query |
| `UDP_SHORT` | Short UDP (below NHP minimum frame) | A datagram shorter than the NHP minimum frame length |
| `NON_TCP_UDP` | Other protocol (ICMP etc.) | A protocol that is neither TCP nor UDP |
| `V6_OTHER` | IPv6 | IPv6 traffic that no rule admits |
| `OTHER` | Other | Everything else the program drops |
| `UNKNOWN-<n>` | Unknown reason | An action code the naming table does not know |

The nhp-ac publishes into the same site, and its DROP reasons reuse the server's vocabulary on purpose — a category the site does not know about cannot be counted together with the rest. It adds exactly one of its own:

| Reason | Category | What it means |
| --- | --- | --- |
| `KNOCK_EXPIRED` | Knock window expired | A peer that knocked, but whose knock window closed before it spoke again |

That last one is a *legitimate peer arriving late* — a retransmit, a session that outlived its window, an agent whose clock ran ahead. It is not a scan, and the site does not count it as an attack. The packets are dropped either way; the exclusion is about the accounting, not about what the filter admits.

The site's classifier treats an **unknown** reason differently, and in the opposite direction: an unknown reason *does* count as an attack. The classifier is a closed list, so a reason this build has not heard of is more likely a new attack class than a new non-attack one, and under-reporting attacks is the worse mistake.

### Source locations

- `nhp/utils/ebpf/engine_linux.go` — `serverActionName`, `acActionName`; the two naming tables.
- `endpoints/xdplog-report/classify.go` — `dropCategories`, `classify`, `nonAttackCategories`, `isAttackReason`.
- `endpoints/xdplog-report/report_test.go` — `TestClassifierMatchesEngine`, the cross-check that keeps the two vocabularies in step.
- `nhp/ebpf/xdp/nhp_server_xdp.c` — the program behind the server's verdict tree.

## 3. The log the filter writes

The filter's verdict is turned into text by the daemon, and the text is the site's only input. Three line shapes exist, plus one marker. These are real lines from the repository's test fixture, `endpoints/xdplog-report/testdata/nhp_server_xdp-2026-10-01.log`:

```
10:00:00 srv1 [NHP-STAT] VERDICT=DROP REASON=UDP_SHORT WINDOW=60s PKTS=1000 LOGGED=0 TOTAL=1000
10:03:00 srv1 [NHP-DROP] REASON=TCP_OTHER SRC=203.0.113.5 DST=10.0.0.1 LEN=60 PROTO=TCP SPT=1000 DPT=22 RELAY=0
10:04:00 srv1 [NHP-SCAN] WINDOW=300s KIND=TOTAL UNIQ_SRC=2 UNIQ_PORT=2 PKTS=2 LRU=0
```

and the budget marker:

```
10:02:00 srv1 [NHP-STAT] EVENT_LOG_BUDGET_EXHAUSTED
```

What each one means:

| Line | Cadence | Exactness | What it carries |
| --- | --- | --- | --- |
| `[NHP-STAT]` | One per active class per window (`WINDOW=60s`) | **Exact** | `PKTS` (packets in this window), `LOGGED` (how many were also written per-packet), `TOTAL` (the class's running total since the daemon started) |
| `[NHP-DROP]` | Per sampled packet | **A sample** | `REASON`, `SRC`, `DST`, `LEN`, `PROTO`, `SPT`, `DPT`, `RELAY` |
| `[NHP-SCAN]` | Periodic summary (`WINDOW=300s`) | Exact per window, unless `LRU=1` | `UNIQ_SRC`, `UNIQ_PORT`, `PKTS` — the window's dropped packets, counted per source and per destination port — and `LRU`, the flag that says the counting tables were at least 95% full and may have evicted entries |
| `EVENT_LOG_BUDGET_EXHAUSTED` | Once per day | — | The day's per-packet byte budget is spent; the `[NHP-STAT]` totals continue |

The distinction between the first two lines is the single most important thing to understand about the published numbers:

- **`[NHP-STAT]` is the count of record.** It is emitted per class per window, it is not rate-limited, and it is where the site's exact packet totals come from.
- **`[NHP-DROP]` is a sample.** It is written under a token bucket in the kernel, so it shows *what* was dropped and *from where*, but never *how many*. Any ranking derived from it — the top sources, the top destination ports — is a ranking of a sample.

The third line is a third channel again, and the site does not publish from it. The kernel counts *every* dropped packet into two bounded tables — one per (protocol, destination port), one per source — whatever the reporting budget says, so the `PKTS` on that line is an exact count of the window's drops. The tables are LRU, so a sweep wider than the table evicts the coldest keys and the ranking becomes approximate; `LRU=1` is what announces that. The site's own top sources and top destination ports come from the sampled per-packet lines instead.

The `TOTAL` field is cumulative since the daemon started. That is what makes a restart visible: a `TOTAL` that goes backwards between two windows means the daemon restarted, and the site reports that as `restarts_detected`.

Three bounds keep the log from becoming a denial-of-service tool against the host it is protecting, and none of them changes a verdict:

- **A per-class token bucket in the kernel.** Each action class gets a burst of 16 per second per CPU, plus a global bucket of 8 per second per CPU across all classes. The per-class bucket stops a flood of one class from starving the records of another; the global one bounds the total whatever mix of classes it is made of.
- **A daily byte budget in the writer.** 64 MiB of per-packet text per calendar day. When it is spent, the writer stops adding per-packet lines and says so — but the `[NHP-STAT]` summaries keep reporting exact totals.
- **A retention sweep.** 14 days or 256 MiB, whichever bites first; oldest files go first and today's file is never removed.

The log files themselves are created with mode 0600 by `nhp/log/logger.go`, and they hold full, unmasked source addresses. On the demo host that is the only place those addresses exist.

### Source locations

- `endpoints/xdplog-report/testdata/nhp_server_xdp-2026-10-01.log` — the fixture the samples above are taken from, byte for byte.
- `nhp/utils/ebpf/engine_linux.go` — `serverLogDailyEventBytes`, `serverLogRetentionDays`, `serverLogRetentionBytes`, `serverLogSweepInterval`, and the `[NHP-STAT]` writer.
- `nhp/ebpf/xdp/nhp_xdp_stats.h` — `NHP_EVENT_BURST`, `NHP_EVENT_GLOBAL_BURST`: the kernel-side rate limits.
- `nhp/log/logger.go` — the writer that creates the file with mode 0600.
- `endpoints/xdplog-report/analyze.go` — the restart detection that turns a decreasing `TOTAL` into `restarts_detected`.

## 4. From a dropped packet to a published row

The pipeline has five hops. Each one is narrow on purpose.

```
   ┌──────────────────────────────────────────────────────────────────┐
   │  the internet                                                    │
   └───────────────────────────┬──────────────────────────────────────┘
                               │  frame arrives at the NIC
                               ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │  nhp-server host                                                │
   │                                                                  │
   │  XDP program ──DROP──► perf ring ──► Go reader ──► log line      │
   │  (kernel)                        (engine_linux.go)                │
   │                                                                  │
   │  logs/nhp_server_xdp-YYYY-MM-DD.log   mode 0600, 14 days         │
   └───────────────────────────┬──────────────────────────────────────┘
                               │  SSH, forced command: list / get
                               ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │  relay host                                                      │
   │                                                                  │
   │  xdpstats-collect.sh                                             │
   │    raw/<source>/nhp_<source>_xdp-<date>.log.gz  (+ .meta)        │
   │    0700, 90 days, flock-guarded                                  │
   │                                                                  │
   │  xdplog-report site ──► /var/www/stats                           │
   └───────────────────────────┬──────────────────────────────────────┘
                               │  nginx, read-only
                               ▼
   ┌──────────────────────────────────────────────────────────────────┐
   │  https://stats.opennhp.org/                                      │
   │    index.html  index.json  d/<source>/<day>.{html,json}          │
   │    totalnumber.js  countup.js                                    │
   └──────────────────────────────────────────────────────────────────┘
```

**Hop 1 — the filter writes the log.** Covered in chapter 3.

**Hop 2 — a read-only SSH channel.** The collector does not get a shell. A dedicated ed25519 public key is installed in the server's `authorized_keys` behind a forced command, and that command accepts exactly two verbs:

- `list` — one TSV line per log file: date, size, mtime.
- `get <YYYY-MM-DD>` — stream that one log to stdout.

Everything else is refused at the script: shell metacharacters, path separators, dates that do not match `^[0-9]{4}-[0-9]{2}-[0-9]{2}$`, days that were never written, and missing files. The `get` path refuses to follow a symlink, so a hostile owner of the log directory cannot point the collector at an arbitrary file.

The key is pinned to the relay's private address and carries `restrict`, which turns on every restriction OpenSSH knows about — no port forwarding, no X11 forwarding, no agent forwarding, no pty, no user rc — so a leaked key is useless from anywhere but the relay. The entry reads, in full:

```
from="<relay private IP>",restrict,command="<path>/nhp-xdplog-export.sh" <collector public key>
```

**Hop 3 — the collector caches.** The collector runs on the relay host, once a day. For every (source, day) it pulls a day when the cache or its sidecar is missing, when the day is today — its log is still being written — or when the size or mtime the server reports for it has changed. For each such day it runs `get`, rejects an empty download, trims a trailing partial line, gzips the result with `gzip -n`, and writes it atomically. The sidecar is what makes the third case work: it records the size and mtime the server reported at the moment of the pull, so the next run can tell "unchanged" from "changed" without comparing against the compressed local copy. The cache directory is mode 0700 and is the only place on the relay where unmasked source addresses exist.

Two operational details of this hop are worth knowing:

- The run is guarded by a non-blocking `flock`. A second invocation — a timer tick landing on top of a manual run — exits successfully without doing any work.
- A source that could not be listed makes the collector exit non-zero. The server keeps only about two weeks of logs, so a collector that has quietly stopped reaching the server is a failure nobody notices until the data is gone.

**Hop 4 — the analyzer builds the site.** `xdplog-report site` reads the cache and writes the web root. It is read-only on its inputs; the only files it removes are the pre-floor detail pages it published itself. It is incremental: a day whose inputs have not moved is skipped, and the index is rewritten only when something moved.

The collector passes four optional flags, each only when the thing it names is actually available:

| Flag | Passed when | What it changes |
| --- | --- | --- |
| `-geo-mmdb <path>` | The GeoLite2-City database is readable on the relay | A source address resolves to an ISO 3166-1 alpha-2 region; without it the site publishes no region figures at all |
| `-history <path>` | The history snapshot is readable | The days whose per-day records are gone are published from the snapshot instead |
| `-force` | A deploy asked for a full rebuild | Every cached day is re-analyzed and every page is rewritten |
| `-state <path>` | Always | Where the incremental state lives — deliberately *outside* the web root |

A `-history` path that cannot be read stops the run. The snapshot is the only carrier of the totals for the days before the publishing floor, so silently falling back to counting only the site's own days would shrink the published all-time numbers by an amount nobody could audit.

**Hop 5 — nginx serves it.** The site is a flat static tree under `/var/www/stats`, with a read-only vhost. The details of what that vhost refuses are in chapter 7.

### Source locations

- `deploy/scripts/nhp-xdplog-export.sh` — the forced command: the `list` / `get` verbs, the date regex, the symlink refusal.
- `deploy/scripts/nhp-ac-xdplog-export.sh` — the same script for the nhp-ac.
- `deploy/scripts/xdpstats-collect.sh` — the collector: the cache layout, the flock, the retention sweep, the flags it passes.
- `deploy/scripts/xdpstats.timer` — the daily schedule.
- `deploy/config-templates/xdpstats/sources.conf` — the source list, and the note on what a new source needs.
- `deploy/nginx/stats.conf.template` — the vhost.
- `endpoints/xdplog-report/site.go` — `runSite`, the incremental build and the publishing floor.

## 5. What the published numbers mean

This is the chapter a reader of the site actually needs, so it is worth being precise about each field.

### The per-day row

`index.json` is an array of one row per (source, day). The fields and their exact meaning:

| Field | Meaning |
| --- | --- |
| `source` | Which host wrote the log: `server` or `ac`. The homepage renders these as `nhp-server` and `nhp-ac`; the JSON keeps the raw name. |
| `day` | The UTC calendar day the row covers. |
| `partial` | The row covers a day that is still being written — normally the current UTC day. Its numbers will grow. |
| `drop_pkts` | The exact number of attack packets dropped that day, taken from the `[NHP-STAT]` lines. |
| `pass_pkts` | The number of packets admitted. On a row that came from a history snapshot it means "not recorded", not zero. |
| `drop_share` | The share of inbound traffic that was dropped: attack drops over the day's total of attack drops, non-attack drops, and passes. |
| `unique_attack_ips` | The number of *distinct* attack sources that day. It is a per-day count, not a per-day sum, and it is derived from the sampled per-packet lines — see *Which figures are exact* below. |
| `peak_hour` | The hour of the day with the most dropped attack packets. |
| `restarts_detected` | How many times the filter restarted that day, detected from a `[NHP-STAT]` `TOTAL` that went backwards. |
| `budget_exhausted` | The day's per-packet byte budget was spent, so the per-packet lines stop part-way through. The packet totals remain exact. |
| `unknown_reasons` | A reason appeared that the site's classifier does not know. Those packets are counted as attacks. |
| `top_category` | The category with the largest share of the day's dropped attack packets. |
| `history` | The row came from a history snapshot rather than from a log. Such a row has no detail page. |

Two of these need more than a line of explanation.

**`unique_attack_ips` cannot be added up.** An address that scans the host every day is one attacker. Summing the per-day counts counts it once per day. The site's all-time attacker figure is therefore not the sum of the rows — it is a *union* over the days, computed from per-day sets the site keeps privately (chapter 6).

**A missing hour is not a quiet hour.** The published hourly series only contains hours that have a `[NHP-STAT]` line. A missing hour means the filter produced no summary for it — the filter was detached, the daemon was restarting, or the day's per-packet budget had been spent. It does not mean zero attacks happened in that hour.

### Which figures are exact, and which come from the sample

Not every published figure comes from the same channel, and the distinction matters: a reader who assumes they are all exact will over-read the two that are not.

Exact — they come from the `[NHP-STAT]` lines, which are not rate-limited:

| Figure | Field |
| --- | --- |
| Dropped attack packets | `drop_pkts` |
| Admitted packets | `pass_pkts` |
| Share of inbound traffic dropped | `drop_share` |
| Busiest hour | `peak_hour` |
| Largest category | `top_category` |
| Filter restarts | `restarts_detected` |
| Byte budget spent | `budget_exhausted` |
| Unclassified reasons seen | `unknown_reasons` |

Sampled — they come from the per-packet `[NHP-DROP]` lines, which the filter rate-limits:

| Figure | Field |
| --- | --- |
| Distinct attack sources that day | `unique_attack_ips` |
| Distinct attack sources, all time | `NHP_TOTAL_ATTACK_SOURCE_IPS` |
| Distinct regions, all time | `NHP_TOTAL_ATTACK_REGIONS` |

A source whose packets were all rate-limited away before they were recorded does not appear in the sampled figures, so the distinct-source and region counts are **lower bounds** on the sources that actually attacked: they count the attackers the filter managed to write down. The packet totals are not affected — they come from the exact channel — and neither are the restarts, the budget marker, or the unknown reasons. The history baseline, when it is applied, is the one thing that can push the source figure the other way, to an upper bound; see *The history baseline* below.

### The all-time figures

The homepage's hero number, and the first constant in `totalnumber.js`, is the **exact** total of attack packets dropped over every published day. It is reproducible: sum `drop_pkts` over every row of `index.json` and you get the same figure.

The other two all-time figures are **unions**, not sums:

- **Distinct source IPs** — the number of addresses seen on at least one published day, counted once each.
- **Distinct regions** — the number of distinct ISO 3166-1 alpha-2 codes those addresses resolved to. A source the GeoIP database cannot place belongs to no region and raises the source count only.

Neither is reproducible from `index.json`, because the per-day sets behind them cannot be published (chapter 7).

### The publishing floor

The site publishes nothing before `2026-10-16`. That date is `defaultSiteSince` in `endpoints/xdplog-report/site.go`, and it is a *publishing* decision, not a data-availability one: the XDP filter was attached to the demo host after the logs that were already sitting in its log directory, so everything earlier is a partial record. Counting it would understate the totals by an amount nobody can audit.

Days below the floor are not analyzed, not rendered, not in the index, and not in the totals. A `-since` that excludes every discovered day is refused *before* anything is deleted — a typo cannot blank a populated index or leave a site of dead links.

### The history baseline

A history snapshot is a saved copy of the site's own numbers for the days the collector's cache no longer reaches. It does two things: it publishes the days it lists whose per-day records are gone, and it seeds the all-time counters.

What it contributes depends on what it can prove:

| State | What it means | What the figures become |
| --- | --- | --- |
| Applied | Some of the snapshot's days have lost their per-day records; the snapshot's own source count stands in for them | The source-IP figure is **an upper bound** — an address seen in both periods can be counted twice. The region figure is unaffected. |
| Missing | A published day's per-day records are gone and nothing stands in for them | Both figures are **lower bounds** |
| Below the floor | The snapshot lists days the site does not publish | The baseline is **dropped entirely** — a snapshot-level total cannot be split by day — and both figures are lower bounds |

The snapshot's own *region* count is never applied, under any state. Regions are a closed set of roughly 250 codes that both periods share almost entirely, so any arithmetic combining the two can only be wrong: adding them would count most regions twice, while taking the larger would miss a region seen in one period alone. The published region total is the union over the published days and nothing else.

The homepage and `totalnumber.js` both print, figure by figure, which kind of number each one is. When they are exact, they say nothing.

### Source locations

- `endpoints/xdplog-report/site.go` — `SiteDaySummary` (the published row contract), `defaultSiteSince`, `siteRenderVersion`.
- `endpoints/xdplog-report/analyze.go` — `Report`, `maskIP`, `isAttackReason`'s effect on the counters.
- `endpoints/xdplog-report/classify.go` — `nonAttackCategories`, `isAttackReason`.
- `endpoints/xdplog-report/history.go` — `historyBaseline`, `sourcesBound`, `regionsBound`, `boundNote`, `BelowFloor`.
- `endpoints/xdplog-report/facts.go` — `unionDayFacts`, the per-day sets the unions are computed from.
- `endpoints/xdplog-report/geo.go` — `isReserved`, the ranges that never get a geo lookup.

## 6. The published artifacts and their contracts

The site is a flat directory of static files. There is no application server and no database behind it.

| Artifact | What it is | Who reads it |
| --- | --- | --- |
| `index.html` | The homepage: the all-time hero total, the counting start date, and one row per day | A browser |
| `index.json` | The machine-readable index: one `SiteDaySummary` object per (source, day) | Anything that wants to chart the data |
| `d/<source>/<day>.html` | One day's detail page | A browser |
| `d/<source>/<day>.json` | One day's `Report`, schema 2 | Anything that wants the detail |
| `totalnumber.js` | Three JavaScript constants for the all-time figures | Another site, via `<script src>` |
| `countup.js` | The homepage counter animation | The browser |

**`index.json` is authoritative.** `totalnumber.js` is its sibling, not its replacement: the packet total in it is reproducible by summing `drop_pkts` over `index.json`, and the other two figures are not — they are unions, and the per-day sets behind them are private. A consumer that wants the numbers and nothing else should use `totalnumber.js`; a consumer that wants to verify them should use `index.json`.

**`totalnumber.js` exports three constants:**

```javascript
const NHP_TOTAL_ATTACK_BLOCKED = <packets>;
const NHP_TOTAL_ATTACK_SOURCE_IPS = <addresses>;
const NHP_TOTAL_ATTACK_REGIONS = <regions>;
```

All three are written even on an empty site, as `0`, so an embedding page never gets a 404 and a dangling reference. The third constant counts **regions**: the two-letter ISO 3166-1 alpha-2 codes the international standard assigns to every country and area it lists, which includes dependencies, disputed areas and special political areas as well as states. It was once called `NHP_TOTAL_ATTACK_COUNTRIES`, and that name is gone — a published figure called a country count is a claim about those codes that the standard does not support. The old name is still *read* when the `history` subcommand parses an older file, and is never written.

**The detail page's JSON is a `Report`, schema 2.** It carries the day's totals (`total_pkts`, `drop_pkts`, `pass_pkts`, `drop_share`, `non_attack_drop_pkts`), the per-reason and per-category breakdowns, the top sources and top destination ports, the hourly series, the protocol mix, the restart list, and the unknown reasons. The hourly series is present only for hours that have a `[NHP-STAT]` line, which is why a missing hour is not a quiet hour. It also says how representative the sampled figures are: `drop_sampled_events` is how many dropped packets were written out per-packet, and `drop_sampling_rate` is that number over the day's exact drop total — the share of the drops the filter managed to write down.

**The homepage renders a source's display name, the JSON does not.** `server` reads as a generic noun in a table that is otherwise all dates and numbers, so the homepage shows `nhp-server` and `nhp-ac`. `source` is a published field of `index.json` and renaming it would break anyone already parsing it, so the JSON keeps the raw directory name.

**The site loads no third-party resources.** The two scripts it references — `countup.js` and `totalnumber.js` — are same-origin files the build writes, and the content security policy allows scripts only from the site's own origin (`script-src 'self'`). There is exactly one documented exception to self-containment: the homepage's `<head>` carries a canonical URL and the Open Graph / Twitter card tags, whose values are absolute URLs, because a crawler reads them and the browser never fetches them.

The counter is progressive enhancement. The hero number is already in the HTML for readers without JavaScript, without that content security policy, or with reduced motion; the script only rewinds it and counts up.

**Template changes reach published pages.** Each day's state records the `siteRenderVersion` that produced its page, so bumping that constant re-renders every published day on the next run. Days whose raw log has already been pruned are re-rendered from their published JSON — the HTML is a pure function of it — but their numbers are never recomputed.

### Source locations

- `endpoints/xdplog-report/site.go` — `SiteDaySummary`, `SiteDayState`, `SiteState`, `siteRenderVersion`, `siteStateSchema`.
- `endpoints/xdplog-report/render_index.go` — `siteURL`, `siteOGImage`, `sourceLabels`, `sourceLabel`, `writeTotalNumberJS`, `totalNumberJS`, `writeCountUpJS`.
- `endpoints/xdplog-report/analyze.go` — `Report`, `reportSchema`.
- `endpoints/xdplog-report/site_test.go` — `assertSelfContained`, the test that keeps the site's own references in bounds.

## 7. What is published, and what is not

The site is a deliberately public, read-only summary of attack telemetry. There is no login, no rate limit, and no geo-restriction.

**Source addresses are masked.** An IPv4 address is truncated to `a.b.c.x`; an IPv6 address is truncated to its first 48 bits. The `site` subcommand defaults `-mask-ip` to true, which is different from the single-file command line tool, whose default is false. Masking is applied at the site, not by the filter: the raw log on the server always holds the full address.

**Private and reserved ranges are excluded** from the attack source count and never get a geo lookup. A source the GeoIP database cannot place counts towards the source figure but belongs to no region.

**The published artifacts hold no full addresses, and the private ones do.** Three directories, three permission stories:

| Directory | Mode | What it holds |
| --- | --- | --- |
| The web root | 0755 under `/var/www` | The published site: masked addresses only |
| The raw cache | 0700 under `$HOME` | The gzipped logs, with **full** source addresses — the only unmasked copy on the relay |
| The state directory | 0700 under `$HOME` | `state.json`, and the per-day *facts* files |

The state file records the absolute path of every raw cache file, which is host layout an anonymous visitor has no business reading. The facts files hold the per-day sets of hashed source addresses and region codes that the all-time unions are computed from. Both live outside the web root, and the nginx vhost refuses `/state.json` and `/facts/` a second time in case a future change puts them back inside.

The facts files store a 64-bit hash of an address, never the address. But a 64-bit hash of an IPv4 address is reversible by brute force over the whole 2^32 space — publishing the hashes would hand back exactly what the masking hides. That is why the facts files are private, and why they live beside the state rather than under the web root.

The web root is under `/var/www` rather than under `$HOME` for a mundane reason: on Amazon Linux 2023 `/home/ec2-user` is mode 0700 and nginx workers run as `nginx`, so a web root under `$HOME` would answer every request with a 403.

### Source locations

- `endpoints/xdplog-report/analyze.go` — `maskIP`.
- `endpoints/xdplog-report/geo.go` — `isReserved`.
- `endpoints/xdplog-report/facts.go` — the facts file format, its magic and its hash.
- `deploy/scripts/xdpstats-collect.sh` — the three directory roots and their modes.
- `deploy/nginx/stats.conf.template` — the refusals of `/state.json` and `/facts/`.

## 8. Running and operating it

The operational truth for this pipeline is the runbook, not this article. This chapter says what runs and when, and where the details live; it deliberately does not duplicate the runbook's procedures, which would produce a second copy that rots.

**The daily run.** A systemd timer on the relay host fires the collector once a day at 00:20 UTC, one full hour after the UTC day rollover so a slow writer cannot leave the previous day's log half-written. The timer is `Persistent=true`, so a host that was off at the window runs the service once on the next boot to catch up.

**A forced rebuild.** `xdplog-report site` only rewrites the index when an input moved. A deploy that changes what the site says — the published figures, the hero, the markup — would therefore leave the previous build published until some log happened to change. The deploy asks for a forced rebuild through the systemd manager's environment, so the run still happens inside the unit's sandbox, and the collector passes `-force` on. The daily timer run does not: there, the incremental path is the point.

**The deploy order.** The `stats.opennhp.org` A record is a Terraform resource, and the deploy workflow never applies Terraform — it only reads the state. So the release that introduces a change to the site has to be `infra-demo` (apply) first, and `deploy-demo-v2` (deploy) second. The deploy's first step is a preflight that requires `stats.opennhp.org` to resolve to exactly one A record equal to the relay's public address; a forgotten apply is a red run before the relay is touched, rather than a red curl afterwards.

**The deploy's own gates.** The deploy checks three things about the run it started, because the collector's exit code cannot answer any of them: that the forced-rebuild line appears in the collect log, that the published `index.html` is newer than a marker created before the run, and that the published region figure is a plausible number — there are only about 250 ISO 3166-1 alpha-2 codes, so a larger figure means the arithmetic is wrong. A source that could not be listed is reported as a warning, not a failure, because the site is still perfectly well formed; it just may be stale.

**Adding a source.** A new source needs four things: a line in `sources.conf` (the name must be `server` or `ac`, because the log prefix `nhp_<name>_xdp` is derived from it), an export script on that host, a collector public key installed in that host's `authorized_keys` behind that script, and a parser that understands the log's shape. The nhp-ac's export script is a near-copy of the nhp-server's; the two hosts' data ends up in the same published site.

**Where the procedures are.** `terraform/demo/RUNBOOK.md` carries the deployment and rollback steps, the recovery paths, the manual backfill procedure when a timer window is missed, and the key rotation procedure. The hands-on procedure — the netns rehearsal for a filter change, the pre-deploy baseline, and the same commands to run after a deploy — is in `terraform/demo/VERIFY-server-xdp.zh-cn.md`, which is written in Chinese. An English reader who wants the filter's behavior explained rather than rehearsed is served by [eBPF Internals](/ebpf/) and by the runbook's own *nhp-server ingress filter (XDP)* section.

### Source locations

- `deploy/scripts/xdpstats.timer` — the schedule.
- `deploy/scripts/xdpstats-collect.sh` — the `-f` / `XDPSTATS_SITE_FORCE` path.
- `.github/workflows/deploy-demo-v2.yml` — the `deploy-xdpstats` job: its preflight, its forced rebuild, and its gates.
- `terraform/demo/RUNBOOK.md` — sections *XDP attack stats site (stats.opennhp.org)*, *Analysing the XDP event log (attack report)*, and *nhp-server ingress filter (XDP)*.
- `terraform/demo/VERIFY-server-xdp.zh-cn.md` — the verification procedure (Chinese).

## 9. Where to look next

| You want to know | Go to |
| --- | --- |
| How the filter decides a packet's fate, and what each field of `etc/xdp.toml` controls | [eBPF Internals](/ebpf/) |
| Where things live in the repository | [Understand the Code](/code/) |
| The CLI of the analyzer, every flag, and the report's fields | `endpoints/xdplog-report/README.md` |
| Deploying, rolling back, and recovering the stats site | `terraform/demo/RUNBOOK.md` |
| Verifying a filter change before and after it ships | `terraform/demo/VERIFY-server-xdp.zh-cn.md` (Chinese) |
| The protocol the filter deliberately does not parse | [Protocol Reference](/protocol/) |
| The default-deny property this layer implements | [Features](/features/) |

The site itself is at <https://stats.opennhp.org/>. Its `index.json` is the authoritative, machine-readable answer to "what did the filter drop", and `totalnumber.js` is the short form of the same answer for a page that wants to embed it.
