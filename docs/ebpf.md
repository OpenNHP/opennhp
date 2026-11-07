---
layout: page
title: eBPF Internals
nav_order: 13
permalink: /ebpf/
description: How OpenNHP's XDP/eBPF layer decides the fate of a packet, and what each knob in etc/xdp.toml actually controls.
---

# How OpenNHP's eBPF Layer Works
{: .fs-9 }

OpenNHP filters in two places: a kernel-level XDP program that decides whether a packet may exist at all, and the user-space Noise handshake that decides who it belongs to. This article covers the first one — the eBPF programs, the loader that attaches them, and the configuration that drives them.
{: .fs-6 .fw-300 }


---

If you write an NHP integration, three questions come up before anything else:

1. **At which layer is my knock decided?** Answer: in two steps. The XDP program in the kernel decides whether the datagram reaches the daemon at all — it is attached to the host's default-route interface and runs before the network stack ever sees the packet. Whether the knock is *valid* is decided in user space, by the Noise handshake.
2. **I changed a field in `etc/xdp.toml`. Does it take effect?** Answer: only `RelayIPs` does. Every other field is read once, at startup. Chapter 7 pins down which is which.
3. **The filter is not attached. Where do I look?** Answer: the daemon's log, and `etc/nhp_server_xdp.o` next to the binary. Chapter 8 covers the log lines and what they mean.

This article explains the mechanisms behind those answers. Every statement here maps to a specific file in the repository; the source path is given the first time each component appears. The code is the authority: where this guide quotes a file, the quote is verbatim, and if the two ever disagree, the source wins.

## Contents

1. [Why filter in the kernel](#1-why-filter-in-the-kernel)
2. [eBPF in one page](#2-ebpf-in-one-page)
3. [The topology: three programs, two daemons, one loader](#3-the-topology-three-programs-two-daemons-one-loader)
4. [The nhp-server ingress filter, one packet at a time](#4-the-nhp-server-ingress-filter-one-packet-at-a-time)
5. [The nhp-ac per-knock filter](#5-the-nhp-ac-per-knock-filter)
6. [The loader: from `.o` to a running filter](#6-the-loader-from-o-to-a-running-filter)
7. [Configuration: what `etc/xdp.toml` actually controls](#7-configuration-what-etcxdptoml-actually-controls)
8. [Observability and troubleshooting](#8-observability-and-troubleshooting)

## 1. Why filter in the kernel

NHP is an *authenticate-before-connect* protocol. A protected service is not supposed to be reachable until a knock has been authenticated, so the interesting question is not "who is this packet from" but "should this packet have been sent to this host at all".

In OpenNHP that question is answered in two different places, and the split matters:

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

The XDP program answers the first question. It **does not parse the NHP protocol** — identity is still decided in user space by the Noise handshake (`RecvPrecheck` in `nhp/core/packet.go`). The only NHP-shaped test the kernel makes is a minimum datagram length, which is enough to turn the knock port into a non-answer for ordinary scanners and costs nothing.

Why bother with the kernel at all? Three reasons, all of them visible in the code:

- **Cost.** A packet dropped at XDP never runs netfilter, never runs routing, never reaches the socket lookup and never wakes a daemon. On a host being scanned continuously, that is the difference between a machine doing nothing and a machine doing work on behalf of someone else.
- **Timing.** XDP runs before netfilter, before routing and before the socket lookup. A drop there is final: the host never has a socket to queue the packet on, so there is nothing to slowloris and no half-open state to clean up.
- **Reachability.** The daemon's own listener is one UDP port. Everything else — SSH, the HTTP knock listener, the metrics endpoint, any service someone bound later — is subject to the filter at the driver, whether or not the daemon knows about it.

That last point has a consequence worth internalizing: the nhp-server's filter is a *host* filter, not a *daemon* filter. It drops inbound TCP to any service on the box, not just to `nhp-serverd`. Chapter 8 covers what happens when that conflicts with a listener.

## 2. eBPF in one page

You do not need to be a kernel programmer to work on this layer, but five concepts carry most of the weight.

**The verifier.** When a BPF program is loaded, the kernel runs a static analysis pass over its instructions. If the program could loop forever, read memory it has no right to read, or do anything else undefined, the load is rejected — there is no partial load and no "it works until it does not". This is why BPF programs are short: the verifier reasons about every path.

**Hooks.** A program is attached to a *hook point*, not to a file or a daemon. OpenNHP uses two:

| Hook | When it runs | Used by |
|------|--------------|---------|
| `SEC("xdp")` | On ingress, before the network stack | the server filter, the AC's ingress filter |
| `SEC("tc/egress")` | On egress, after the routing decision | the AC's egress filter |

**Maps.** A BPF program cannot hold state in ordinary memory. Everything it remembers lives in a *map* — a kernel data structure that user space can also read and write. OpenNHP uses `LPM_TRIE` (longest-prefix match, for the SSH whitelist), `PERF_EVENT_ARRAY` (a ring buffer user space drains), `PERCPU_ARRAY` (per-CPU counters, so increments need no atomics), `LRU_PERCPU_HASH` (bounded tables for scan aggregation) and `LRU_HASH` (the AC's whitelists).

**CO-RE.** The programs are compiled against `nhp/ebpf/xdp/vmlinux.h`, a header generated from a kernel's BTF, carrying clang's `preserve_access_index` attribute. The compiler emits relocations instead of hard-coded field offsets, and the loader resolves them against the *target* kernel's BTF at load time. One object therefore works across kernels without recompiling.

**Capabilities.** Three, and each covers one step:

| Capability | Needed for |
|------------|-----------|
| `CAP_BPF` | loading the program and creating its maps — and writing to a map afterwards, but only on kernels where `bpf(2)` is itself privileged (chapter 6) |
| `CAP_NET_ADMIN` | attaching the program to a network interface |
| `CAP_PERFMON` | opening the perf ring the event log reads |

OpenNHP grants all three to `nhp-serverd` at startup and takes them back as soon as the load decision is made — `CAP_BPF` excepted on kernels where `bpf(2)` is itself privileged. Chapter 6 has the details.

On the Go side, the loader uses [`github.com/cilium/ebpf`](https://github.com/cilium/ebpf) v0.22.0 (`nhp/go.mod`), a pure-Go library that talks to `bpf(2)` directly: it parses the ELF object, applies the CO-RE relocations itself, and provides the `link` and `perf` helpers. It does not link libbpf — but the C sources use libbpf's header conventions (`SEC()`, `LIBBPF_PIN_BY_NAME`), so those names in this article come from there.

## 3. The topology: three programs, two daemons, one loader

There are three eBPF programs in the tree, and they belong to two different daemons:

| Program | Source | Hook | Attached by | Purpose |
|---------|--------|------|-------------|---------|
| `nhp_server_xdp.o` | `nhp/ebpf/xdp/nhp_server_xdp.c` | `SEC("xdp")` | `nhp-serverd` | One fixed ingress policy for the whole host: knock port + SSH from the whitelist + replies to the host's own flows. |
| `nhp_ebpf_xdp.o` | `nhp/ebpf/xdp/nhp_ebpf_xdp.c` | `SEC("xdp")` | `nhp-acd` | Opens protected ports **per knock**, and closes them when the knock expires. |
| `tc_egress.o` | `nhp/ebpf/xdp/tc_egress.c` | `SEC("tc/egress")` | `nhp-acd` | The other half of the AC's filter: lets the replies to the AC's *own* outbound flows back in, and records what it sent. |

The server and the AC share no maps at all. The server's policy is "these two things may reach me"; the AC's policy is "this peer, for this port, until this knock expires". They have nothing to say to each other, so they do not share a schema.

What they *do* share is everything around the maps — memlock handling, object loading, pinning under `/sys/fs/bpf`, resolving the default-route interface, attaching, draining the perf ring into a log, and tearing the pins down again. That lives in one file, `nhp/utils/ebpf/engine_linux.go`, and the caller selects the variant:

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

The server's own wrapper is `endpoints/server/ebpf/serverengine.go`; the AC's is `endpoints/ac/ebpf/ebpfegine.go`. Both are thin — they fill in paths and parameters and delegate.

### Building the objects

```bash
make ebpf        # objects, then the full build
make ebpf-objs   # objects only
```

The eBPF targets are deliberately **not** a prerequisite of `acd` / `serverd`:

```makefile
# Deliberately NOT a prerequisite of `acd` / `serverd`: those must keep
# building on a host with no clang, which is why the daemon targets and the
# object targets are separate goals the pipeline invokes in sequence. Both
# daemons treat a missing object as a runtime condition, not a build error
# (nhp-acd refuses to start in FilterMode=1; nhp-serverd falls open).
```

The objects land next to the daemon binaries:

```
release/nhp-ac/etc/nhp_ebpf_xdp.o
release/nhp-ac/etc/tc_egress.o
release/nhp-server/etc/nhp_server_xdp.o
```

"Missing object" is therefore a runtime condition with two different answers — the AC refuses to start, the server runs unfiltered. Chapter 7 returns to that split, because it is the single most surprising thing in this layer.

Note that `docker/Dockerfile.server` builds with `make init serverd plugins` and does not build the eBPF objects, so the container image ships no `nhp_server_xdp.o` at all. A containerized `nhp-serverd` therefore runs without the ingress filter, which is the correct fail-open answer for a host whose exposure is managed elsewhere.

## 4. The nhp-server ingress filter, one packet at a time

The program is `nhp/ebpf/xdp/nhp_server_xdp.c`, and its entry point is:

```c
SEC("xdp")
int xdp_server_prog(struct xdp_md *ctx) {
```

Its license is `Dual BSD/GPL`. The decision tree below is the whole policy — there is no second stage.

### The tree

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

Two of these branches deserve their own paragraph, because they are the ones a reader is most likely to get wrong.

Note the "(not recorded)" markers. A malformed packet — a truncated header, an `ihl` below 5, an unknown ethertype — is dropped without an action code, so it is neither counted nor reported. Every drop that *is* a policy decision carries one; the unrecorded ones are indistinguishable from a corrupted frame. The `[NHP-STAT]` totals are exact for every classified class.

The program is attached in **generic mode** (`link.XDPGenericMode`), not driver mode: the frame has already been turned into a skb by the time the program sees it, so what a drop saves is the stack work rather than the allocation. Driver mode runs earlier and skips the allocation too, but it needs driver support and is not portable — which matters, because the filter is attached to whichever interface the host's default route exits through.

### The socket table: how the host's own replies get back in

The server is a client as well as a server. It resolves names, its auth plugin submits OTP mail to SES, the host talks to package mirrors. A filter that admitted nothing but knocks and SSH would drop the *answers* to all of that, and the first symptom was the resolver: `read udp 10.0.1.78:58655->10.0.0.2:53: i/o timeout` — the DNS *answer* dropped, and with it the OTP mail behind it.

The AC solves this with a TC egress program and a `conn_track` map, because it has to reason about flows a knock authorised. The server has no such ambiguity, so it asks the kernel's own socket table:

```c
    owned = is_tcp ? sk->state != BPF_TCP_LISTEN : sk->dst_port != 0;
```

That is the whole rule, and it is the equivalent of netfilter's `--ctstate ESTABLISHED,RELATED -j ACCEPT`:

- **TCP** — any state except `BPF_TCP_LISTEN`. A listener matches every SYN sent to a port the host serves, so admitting it would silently undo the entire filter the day something binds tcp/8080 on the box.
- **UDP** — connected sockets only (`dst_port != 0`). A resolver, an NTP client or `net/smtp`'s dialer `connect(2)` their socket, so the kernel's lookup only matches datagrams from the peer they are talking to. An unconnected socket bound to some port is a listener by another name.

The lookup is `bpf_sk_lookup_tcp()` / `bpf_sk_lookup_udp()` with `BPF_F_CURRENT_NETNS`, which resolves to the namespace of the device the program is attached to — where `nhp-serverd`'s sockets live. The cost is one hash lookup, for packets that are about to be dropped anyway, and the same lookup the network stack would have done had XDP passed the packet.

Two client protocols cannot be answered this way and are admitted by port pair instead:

- **DHCP** — the client receives its lease on a raw `AF_PACKET` socket or on one merely *bound* to udp/68, so the socket lookup finds nothing. Dropping the renewal is terminal: the lease expires, the address goes away, and the host is dark on **every** port. The pair 67 → 68 is narrow enough that a neighbour's broadcast DHCPDISCOVER to udp/67 is still dropped.
- **NTP** — chronyd polls from the socket it has *bound* to udp/123, so the answer (123 → 123) looks exactly like an unsolicited query. Dropping it costs the clock rather than the address, and NHP notices on its own, because knock packets carry a timestamp the server checks against a window.

Neither is a general "admit anything addressed to a bound port" exception, because that would throw away the "no unsolicited packet gets in" property entirely.

### IPv6, and why SSH must arrive over IPv4

`handle_ipv6()` runs the same tree minus the whitelist. The whitelist is a trie of **IPv4 prefixes**, so the program cannot recognise the relay over v6 — nothing reaches tcp/22 or the knock port over v6 at all.

The exceptions are the ones that keep a v6 host working: ICMPv6 "packet too big" and the contiguous MLD/ND block (types 130–137, which include neighbour discovery — v6's ARP), DHCPv6 (547 → 546), NTP, and replies to the host's own v6 flows via `has_local_flow6()`. Everything else is `V6_OTHER`.

`warnOnGlobalIPv6()` in the loader logs a `Warning` naming the address when the filtered interface has a global v6 address, because on a host whose only SSH path is v6 this would be a lockout.

### Fragments

Only the first fragment of an IPv4 datagram carries the L4 header, so only that one can be judged by a port-based policy. Non-first fragments are **passed** under their own action code (`IPV4_FRAGMENT`):

- They cannot reach a socket on their own — reassembly needs the first fragment, and that one went through the full tree.
- Dropping them instead would make every datagram larger than the path MTU fail, knocks included.
- Reading "ports" out of a fragment's payload, which the program used to do, is worse than either: a crafted fragment whose payload happened to read 67 → 68 took the DHCP exception.

An IPv6 fragment header with a non-zero offset gets the same answer (`IPV6_FRAGMENT`).

### The maps

| Map | Type | Size | What it holds |
|-----|------|------|---------------|
| `nhp_relay_ips` | `LPM_TRIE` | 4096, `BPF_F_NO_PREALLOC`, pinned | The SSH whitelist. Longest-prefix match on IPv4 prefixes. |
| `nhp_events` | `PERF_EVENT_ARRAY` | 1024 | The perf ring user space drains. |
| `nhp_event_rate` | `PERCPU_ARRAY` | 33 | Token buckets: 32 action classes + 1 global. |
| `nhp_action_stats` | `PERCPU_ARRAY` | 32 | `packets` and `logged` counters per action class. |
| `nhp_port_hits` | `LRU_PERCPU_HASH` | 65536 | Dropped packets per destination port. |
| `nhp_src_hits` | `LRU_PERCPU_HASH` | 65536 | Dropped packets per IPv4 source address. |

The trie rather than a hash of host addresses is deliberate: the whitelist has to survive the relay being *replaced*. The relay's private address is not pinned by Terraform, so any rebuild gives it a new one — and a whitelist naming only the old address drops SSH from the new relay, which is the one path CI has for pushing a corrected whitelist. Listing the relay's subnet as a prefix closes that trap. A bare address is simply a /32.

The two `LRU_PERCPU_HASH` tables exist because a 65k-port sweep must not be able to exhaust them; LRU bounds the table, and user space reports `LRU=1` when it suspects eviction. The exact per-class totals live in `nhp_action_stats` and are never affected.

### The action codes

Every *classified* packet gets one — a malformed frame is dropped without an action code at all (see the note above the tree). The C side defines the codes in `nhp_server_xdp.c`, the Go side mirrors them in `engine_linux.go` (`serverActionName`), and the log line prints them as `REASON=`:

| Code | Reason | Verdict | Reported per packet? |
|------|--------|---------|----------------------|
| 1 | `SSH_RELAY` | PASS | only on the SYN |
| 2 | `NHP_RELAY` | PASS | yes |
| 3 | `NHP_DEFAULT` | PASS | yes |
| 4 | `TCP_ESTABLISHED` | PASS | counted only |
| 5 | `UDP_ESTABLISHED` | PASS | counted only |
| 6 | `ICMP_FRAG_NEEDED` | PASS | yes |
| 7 | `DHCP_CLIENT` | PASS | yes |
| 8 | `NTP_CLIENT` | PASS | counted only |
| 16 | `IPV4_FRAGMENT` | PASS | yes |
| 17 | `ICMPV6_CONTROL` | PASS | counted only |
| 18 | `V6_ESTABLISHED` | PASS | counted only |
| 19 | `IPV6_FRAGMENT` | PASS | yes |
| 0 | `OTHER` | DROP | defined, but not emitted by any branch today |
| 10 | `TCP_SSH_OTHER` | DROP | yes |
| 11 | `TCP_NHP_PORT` | DROP | yes |
| 12 | `TCP_OTHER` | DROP | yes |
| 13 | `UDP_OTHER` | DROP | yes |
| 14 | `UDP_SHORT` | DROP | yes |
| 15 | `NON_TCP_UDP` | DROP | yes |
| 20 | `V6_OTHER` | DROP | yes |
| 21 | `TCP_SYN_SCAN` | DROP | yes |
| 22 | `TCP_STEALTH_SCAN` | DROP | yes |

The "counted only" classes are the ones that scale with *bytes transferred* rather than with events — the fourth data segment of a connection, the reply to a DNS query. They are still counted; they are just not worth a log line each. One `dnf update` or one CI `scp` of a release tarball would otherwise write more log than a month of knocks.

## 5. The nhp-ac per-knock filter

The AC's job is different: it is the component that actually *opens* a protected port when a knock is authorized, and closes it again when the knock expires. Its filter is therefore a per-knock whitelist with a lifetime, and it needs two programs to work.

### The ingress program (`nhp_ebpf_xdp.c`, `SEC("xdp")`)

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
 ├─ UDP to dport 67 or 68, or with source port 53  → PASS   (DHCP and DNS replies — unconditionally, before any whitelist or conn_track lookup)
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

Every whitelist entry carries an expiry, and an expired entry is **deleted** when a packet finds it — so an entry that ran out stops admitting traffic immediately.

The five lookups run in a fixed order and stop at the first match, which has one visible consequence: an *expired* entry in an earlier map deletes itself and drops the packet on the spot, without the later maps being consulted. An expired `spp` entry therefore beats a live `protocol_port` entry for the same packet.

The reverse branch is the subtle one, so the tree above summarises it rather than stating the rule. Three conditions qualify it:

- It is **skipped** for a peer that holds a live knock, unless the entry carries `CT_FLAG_AC_CLIENT` — that is, unless the AC itself opened the flow. A knock-authorised peer must be served by the whitelists below, never from the reverse lookup.
- An entry blocked by a `knock_peers` record is **deleted**, not served. Such a record never lapses, so the entry it blocks must not linger: the egress program does not test expiry, so the AC's own sending would keep resurrecting an entry that no inbound packet is allowed to reach the delete for.
- An *expired* entry the AC has claimed is **kept**, not deleted. The claim is the AC's standing hold on the tuple, and dropping it would let the very next admitted packet write the `knock_peers` record that would then shut the egress gate on the AC's own flow.

The same qualification applies to what an admitted packet records: nothing at all is recorded for a flow the AC owns — no `knock_peers` row and no `conn_track` entry.

The five whitelist shapes are written by user space (`nhp/utils/ebpf/ebpf.go`), one per knock shape the protocol supports. The `spp` key is the peer, the service, the port and the protocol:

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

The value carries the same expiry:

```go
func (r *whitelistKey) ToWlValue(ttlSec uint64) whitelistValue {
	now, _ := getBootTimeNanos()
	return whitelistValue{
		Allowed:    1,
		ExpireTime: now + ttlSec*1e9,
	}
}
```

### The reverse branch, and `knock_peers`

The AC is a client too — it talks to the nhp-server, to resolvers, to package mirrors. The replies to those flows have to come back in, and the only way to tell them apart from a knock-authorised flow is to know which flows the *AC* opened.

That is what `conn_track` is for, and the TC egress program is what writes the AC's half of it. But a `conn_track` entry alone is not enough, because the AC's own flows would then be refreshable by a peer: the peer chooses its own source port, so a refreshed window on a bare 4-tuple would let a peer that knocked once hold the tuple open forever and open fresh connections on it.

So the ingress program records a lasting, non-expiring note in `knock_peers` for every non-TCP packet a whitelist admits:

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

The `knock_peers` record is what lets the egress gate refuse to claim a knock-protected flow as one of its own. Without it, the first datagram after a knock expires would hand the peer an egress entry and an indefinitely refreshable way back in.

### The egress program (`tc_egress.c`, `SEC("tc/egress")`)

The egress program answers a different question: *is this packet the AC talking, or the AC answering?* It has four gates:

- **Gate 1 — ingress hit.** The reverse 4-tuple is already in `conn_track` as an INGRESS entry, so the peer opened this flow. Nothing is created or refreshed; the flow's lifetime belongs to the knock TTL. The one thing recorded is a FIN the AC itself sends, so that a spoofed inbound FIN cannot cut an admitted session for good.
- **Gate 2 — TCP only a pure SYN creates an entry.** A listening socket never sends a bare SYN, so a pure SYN is proof that the AC is the client here. The flow is claimed with `CT_FLAG_AC_CLIENT`.
- **Gate 3 — UDP: the source port must look like a client's.** A socket the AC opens as a client binds an ephemeral port; a service it listens on does not. The bound is the host's own `net.ipv4.ip_local_port_range`, read at load time and written into the `nhp_config` map — hard-coding the Linux default would drop the AC's own replies on any host that lowers the range. Three fixed ports are the exceptions, because their replies come back to that same fixed port: NTP (123), IKE (500) and IKE NAT-T (4500).
- **Gate 4 — UDP: a peer that holds, or has ever held, a knock is excluded.** `knock_peers` plus the address-keyed whitelists answer it. This is the gate that keeps a knock-protected service from being claimed by the AC's own flow.

The claim flag is what keeps the two apart. A flow the AC opened *before* any knock appeared is never suppressed, and no `knock_peers` record is ever written for its port — otherwise one knock from an address the AC also talks to (the nhp-server's own address, an NTP server's) would shut gate 4 on the AC's own control channel for the life of the daemon.

### The AC's maps

All of the AC's maps are pinned with `LIBBPF_PIN_BY_NAME`, which is how the ingress program and the egress program share one kernel map per name.

| Map | Type | Key |
|-----|------|-----|
| `spp` | `LRU_HASH` | src IP + dst IP + dst port + protocol |
| `sdwhitelist` | `LRU_HASH` | src IP + dst IP |
| `src_port` | `LRU_HASH` | src IP + dst port |
| `port_list` | `LRU_HASH` | src IP + port range |
| `protocol_port` | `LRU_HASH` | dst port + protocol |
| `icmpwhitelist` | `HASH` | src IP + dst IP |
| `conn_track` | `LRU_HASH` | the 4-tuple + direction + flags |
| `knock_peers` | `LRU_HASH` | peer IP + service port + protocol |
| `nhp_config` | `ARRAY` | one slot: the ephemeral-port floor |
| `events` | `PERF_EVENT_ARRAY` | the perf ring user space drains (not pinned) |

Two rules govern these maps, and both matter:

- **Neither eBPF program may write a whitelist.** User space writes them when a knock is authorised, with the expiry derived from the resource's `OpenTime`; the programs only read them.

  > Neither eBPF program may write them -- tc_egress.c used to write `spp` for
  > every outgoing packet, which kept the door open long past OpenTime.

- **`knock_peers` is never deleted**, by either program. It is an LRU so it stays bounded, but an entry that aged out would let the hole it closes reopen.

The AC's perf events carry a two-valued action — 0 = `DENY`, 1 = `ACCEPT` — and are written to two separate logs, `nhp_deny` and `nhp_accept`. The server writes one file, `nhp_server_xdp`, because its actions say *why* a packet was passed or dropped, and splitting them by verdict would throw that away.

### One thing the egress program never does

Every `return` in `tc_egress.c` is `TC_ACT_OK`. The egress program **never drops a packet** — it only records. The AC's ingress program is the one that decides access; the egress program exists so that the AC's own outbound flows get their replies back, and to keep track of what it sent.

That is also why gates 2 to 4 are about *claiming* a flow rather than about denying one: a claim is what tells the ingress program "this flow is mine, serve it from the reverse lookup". Getting a claim wrong costs the claim; getting a denial wrong costs the protocol.


## 6. The loader: from `.o` to a running filter

`nhp/utils/ebpf/engine_linux.go` is the part of this layer with the most surprising behaviour, so it is worth walking through in order. The order is not incidental — it is the contract.

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

Four things in that sequence are load-bearing.

### The constants are written before the object is loaded

The knock port and the length floor are `.rodata` constants, not `#define`s, because both are user-space configuration:

```c
volatile const __u16 nhp_listen_port = 62206;
volatile const __u16 nhp_min_udp_len = 240;
```

`volatile const` is read as "a constant to the verifier, a variable to the compiler" — the verifier still folds the dead branches away after the rewrite, but the compiler cannot bake the initialiser into the instruction stream. The loader rewrites them before the verifier sees them:

```go
	// Rewrite the .rodata constants before LoadAndAssign: after the collection
	// is loaded the verifier has already folded them, so this is the only
	// moment the knock port and the length floor can still be set.
	if constErr := setServerConstants(spec, params); constErr != nil {
		return nil, constErr
	}
```

An object that predates the constants is **refused**, not attached with its own defaults:

```go
		v, ok := spec.Variables[c.name]
		if !ok {
			log.Error("server eBPF object has no '%s' constant — it is older than the loader; rebuild it with `make ebpf-objs`", c.name)
			return fmt.Errorf("server eBPF object has no %q constant", c.name)
		}
```

Attaching an object whose knock port is a compile-time 62206 to a server listening elsewhere would drop every knock at the driver. That looks exactly like the server being down and is invisible in user space — much worse than not attaching at all, which is a state the daemon already handles.

### The whitelist goes in before the program goes on

```go
	// The whitelist goes in before the program goes on. Everything the filter
	// admits that is not a knock arrives on tcp/22 from one of these prefixes,
	// so a program attached with an empty trie is a host with no way in — and
	// if this write is what failed, user space has just proved it cannot fix
	// the map it would have to fix.
```

The whitelist is a **load parameter** rather than something written afterwards, precisely so that this ordering holds. A program attached with an empty trie drops tcp/22 from every source: harmless for the microseconds until the write lands, unrecoverable if that write then fails — the caller would be left with an attached filter, no SSH, and nothing to correct but a map it could not write a moment ago. Passed in as a parameter, a failure happens *before* the attach and unwinds to fail-open like every other pre-attach error.

### Everything that can fail, fails before the attach

The deferred cleanup exists because a failure after the load has left state on the host:

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

The caller only ever sees an error from `EngineLoad`, and it never gets an `EngineHandle` back. A filter that stayed attached behind a returned error would be enforcing with an empty SSH whitelist that user space no longer has a handle to fix. On a host whose only way in is that whitelist, that is unrecoverable.

The defer is registered *after* `LoadAndAssign` succeeds, so a load that fails at that step has no cleanup to run — the pins it left behind are swept by the next `EngineLoad`, which removes this variant's leftovers before it does anything else.

### Losing the event log is not fatal

```go
	// Opened here rather than inside the reader goroutine because
	// perf_event_open(2) is the last thing that needs CAP_PERFMON, and the
	// capabilities are given back when this function returns. Not fatal if it
	// fails: the event log is diagnostics, and losing it is no reason to leave
	// the host unfiltered. The filter runs on without it, saying so.
```

The distinction is worth holding on to: the *verdict* is the filter, the *event log* is the diagnostic. A host with no event log is still filtered; a host with no filter is not filtered at all.

### The capabilities go back

`nhp-serverd` is granted `CAP_BPF`, `CAP_NET_ADMIN` and `CAP_PERFMON` as *ambient* capabilities. It needs all three for about a second at startup and none of them afterwards:

- writing a map needs a map fd the process already holds;
- reading the perf ring is `read(2)` on an fd;
- detaching is `close(2)` on the link fd, and removing a bpffs pin is `unlink(2)`.

Keeping them for the life of the process is a much bigger grant than it looks. `CAP_BPF` together with `CAP_PERFMON` loads kprobe and tracing programs, which may call `bpf_probe_read_kernel()` — arbitrary kernel memory, including other processes' credentials and this daemon's own Noise private key. `CAP_NET_ADMIN` detaches this very filter, rewrites routes and rewrites netfilter rules. And this is a process that parses untrusted UDP from the whole internet and `dlopen`s third-party auth plugins into its own address space.

So the loader drops them as its last act, on **every** path out of the function:

```go
	defer func() { DropLoaderPrivileges(retErr == nil) }()
```

Note the argument. `retErr == nil` means the load produced a live filter, and only then is `CAP_BPF` worth keeping — because the `xdp.toml` reload path writes the relay whitelist map through it. A failed load keeps nothing at all:

> That is the case the daemon runs in for the longest — a load that fails is fail-open for the life of the process — so it is the one that least deserves a capability.

There is one exception, and it is a kernel property rather than a policy choice. On a kernel with `kernel.unprivileged_bpf_disabled != 0` — which is the default on Amazon Linux 2023, Ubuntu and Debian — `bpf(2)` is privileged *wholesale*, including map updates on an fd the process already owns. Dropping `CAP_BPF` there would turn every later `ReplaceRelayIPs` into `EPERM`, and the reload path is the only remote way to correct a whitelist on a host whose only way in is that whitelist. So on such a kernel `CAP_BPF` is kept, and the log says why:

```go
	if keptBpf {
		log.Info("XDP loader privileges dropped: CAP_NET_ADMIN and CAP_PERFMON released, CAP_BPF kept because %s is non-zero (bpf(2) is privileged on this kernel, and the xdp.toml reload path writes the relay whitelist map). The daemon now holds %q",
			unprivilegedBpfSysctl, want)
```

### The drop is a subtraction, never an assignment

```go
// loaderCaps are the three capabilities the unit grants for the load, and the
// only three this file ever takes away. Anything else the process happens to
// hold was granted for some other reason and is not the loader's to give back.
var loaderCaps = []cap.Value{cap.BPF, cap.NET_ADMIN, cap.PERFMON}
```

The loader subtracts those three from the ambient, permitted, effective and inheritable sets. It does not assign a fixed set, and it does not reset to the empty set. On the demo unit the two are the same thing — the grant is exactly these three — but `nhp-serverd` also ships as a container that runs as root (`docker/Dockerfile.server`), and there the process holds the full root set for reasons that have nothing to do with this loader: `CAP_DAC_OVERRIDE` for volume-mounted config owned by another uid, `CAP_NET_BIND_SERVICE` for an `http.toml` reload that rebinds 80/443. Resetting to the empty set there — on every path, including a host that never asked for any of this — would be an irreversible, silent and entirely unrelated privilege change.

The same reasoning applies to the paths that never reach the loader at all. A host with no `xdp.toml`, an unparsable one, `Enabled = false`, a whitelist the trie cannot hold, or an out-of-range `NhpMinFrameBytes` never calls `EngineLoad` — but the unit grants the three capabilities as *ambient* regardless, so each of those states used to mean a daemon holding `CAP_BPF`, `CAP_NET_ADMIN` and `CAP_PERFMON` for its whole life having never once used them. The drop is therefore deferred across every return of the config path too:

```go
	attached := false
	defer func() { dropXdpLoaderPrivileges(attached) }()
```

### Replacing the whitelist

`ReplaceRelayIPs` is the reload path's only map write, and it is written to be all-or-nothing against the *map*, not only against the file:

- **Parsing is all-or-nothing.** One unusable entry fails the whole call. Skipping bad entries looks forgiving but is the dangerous reading of a typo: a list of one misspelled address parses to no prefixes at all, and a list where only the entry SSH really arrives from is misspelled parses to a whitelist that excludes the operator.
- **An empty list is refused.** Emptying the map closes tcp/22 for every source on a host whose only way in is that map.
- **New entries go in before stale ones come out.** An address on both lists is never momentarily absent — a delete-then-insert would open a window in which the relay's own SSH session and forwarded knocks are dropped by the very reload that was meant to keep them working.
- **A failed write rolls back.** If an `Update` fails, the keys this call added are taken back out and the stale sweep never runs, so the map is left holding exactly the list it held before.
- **A map that cannot be read is not written to at all.** The direction that keeps a reachable host reachable.

The whitelist parsing lives in `nhp/utils/ebpf/relayips.go` rather than in the loader, so that the server's config path can validate a rendered `xdp.toml` *before* it decides to attach anything — and so that the validation compiles and is testable on every platform, not only where the loader does.

## 7. Configuration: what `etc/xdp.toml` actually controls

The file is `etc/xdp.toml`, next to the daemon binary (or wherever `XdpConfigPath` points). **Its presence is what opts a host into being filtered at all.** The full template, with a comment per field, is `deploy/config-templates/server/xdp.toml`; the shape of it is:

```toml
Enabled = true
NhpMinFrameBytes = 240
RelayIPs = ["10.0.1.0/24", "203.0.113.7"]
ScanReportIntervalSeconds = 300
ScanReportTopN = 20
EventLogEnabled = true
```

What matters here is which of its fields are applied, which are recorded, and which are hot-reloadable.

| Field | When it takes effect | Notes |
|-------|----------------------|-------|
| `Enabled` | Startup only | `false`, or a missing file, or a missing `RelayIPs` key ⇒ nothing is attached. |
| `RelayIPs` | **Hot reload** | The only hot-reloadable field. A change to the file is applied to the live map. |
| `NhpMinFrameBytes` | Startup only | Written into the object's `.rodata` at load time. Applied, not recorded. |
| `ScanReportIntervalSeconds` | Startup only | Shapes the `[NHP-SCAN]` summary only. |
| `ScanReportTopN` | Startup only | Shapes the `[NHP-SCAN]` summary only. |
| `EventLogEnabled` | Startup only | `false` stops the per-packet log only; the summaries and the filter are unaffected. A reload that changes it is ignored *silently* — there is no warning for this one. |

The knock port is **not** in this file. The daemon takes it from its own `ListenPort`, and the loader writes it into `nhp_listen_port` alongside the length floor. That is deliberate: the filter and the listener cannot disagree.

The reload path is a file watcher, and it is deliberately narrow:

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

If the file is missing or unparsable on a *reload*, nothing is applied and the kernel keeps the whitelist it already had. Overwriting a working whitelist with the contents of a truncated or half-written file is the failure that locks the operator out of SSH, so a reload only ever takes effect once the TOML has parsed.

And `applyXdpConfig` is explicit about the fields it cannot apply:

```go
	if conf.NhpMinFrameBytes != 0 && conf.NhpMinFrameBytes != s.xdpMinFrameBytes() {
		// The floor is written into the object at load time, so an edit after
		// that is not enforced until the next restart.
		log.Warning("xdp config: NhpMinFrameBytes=%d is not applied — the attached program is enforcing %d bytes; restart nhp-serverd to change it",
			conf.NhpMinFrameBytes, s.xdpMinFrameBytes())
	}
```

The same warning is emitted for `Enabled = false` (a live filter cannot be detached by editing a file) and for both `[NHP-SCAN]` knobs.

### Out-of-range telemetry does not stop the filter

The two `[NHP-SCAN]` settings only shape telemetry, so a value outside its range must never decide whether the filter attaches. With a security group open to all ports the filter is the only inbound control, and a refusal here would leave the host fully exposed:

```go
	if seconds < minScanReportIntervalSeconds || seconds > maxScanReportIntervalSeconds {
		log.Error("xdp config: ScanReportIntervalSeconds=%d is out of range [%d, %d]; the [NHP-SCAN] summary is disabled, the XDP ingress filter still attaches",
			seconds, minScanReportIntervalSeconds, maxScanReportIntervalSeconds)
		return 0, DefaultScanReportTopN
	}
```

Whitelist validation is different: it is one of the refusals listed in chapter 8, and a whitelist the trie cannot hold is refused **before** the attach rather than after.

{: .warning }
On a host whose only way in is the whitelist, a wrong `RelayIPs` list is not a recoverable mistake. The daemon refuses to attach when the list names no usable prefix, and it keeps the previous list on a bad reload — but a list that parses and names the wrong prefix will be applied, and the SSH session doing the applying is dropped by it too. There is no break-glass path by design. The recovery procedures are in the demo runbook (`terraform/demo/RUNBOOK.md`, "nhp-server ingress filter (XDP)"); in the general case, recovering means reaching the host out of band.

### fail-open and fail-closed

The two daemons answer a failed load in opposite ways, and the difference is deliberate:

| | nhp-server (ingress filter) | nhp-ac (per-knock filter) |
|---|---|---|
| Load failure | **Fail-open** — the host keeps the exposure it had before the filter existed | **Fail-closed** — the daemon refuses to start |
| Why | The protocol is the access control; the filter is defence in depth. A backstop would have to re-implement the whitelist, the knock port and the DHCP exception by hand, and a backstop that is wrong locks the host out — including the SSH path that is the whitelist's own entry point | The filter *is* the access control: the protected port must be closed whenever the AC is not demonstrably enforcing |
| Backstop | None | `deploy/scripts/nhp-ac-backstop.sh` — a netfilter chain that DROPs the protected ports, installed when the daemon is not enforcing and lifted only once the XDP program is confirmed attached |
| Reboot / restart window | The host is exposed for as long as the filter is not attached | The protected port stays closed |

The AC's backstop exists because the XDP links live only in the daemon's memory — `link.AttachXDP` keeps the returned link in a package variable, so link lifetime equals process lifetime. Only the *programs* are pinned under `/sys/fs/bpf`, which keeps the objects alive but attaches nothing. Enforcement therefore disappears the moment the daemon goes away: a `systemctl stop`, a crash with `Restart=on-failure` backoff, or a failed deploy would all leave the protected port open to the world. The netfilter chain restores the property from outside the daemon.

The server has no such backstop, and the asymmetry is the price of the two policies. The server's filter is not what keeps its protected service hidden — the protocol is; the filter is defence in depth against a host that is scanned anyway. The AC's filter *is* the access control.

### The watchdog

One more thing the server does that is worth knowing about, because it is the difference between "a mistake" and "a rebuild":

```go
// watchHostAddressing is the dead man's switch for the ingress filter.
//
// The filter is written entirely in terms of the host's IPv4 configuration,
// and there is no break-glass path into the demo host: if something the host
// needs is dropped and the effect is that it loses that configuration, nobody
// can get in to fix it.
```

If the filtered interface has no routable IPv4 address for two minutes (four 30-second strikes), the watchdog logs a `Critical` line and detaches the filter, so the next DHCP exchange can restore the host. A `169.254.0.0/16` autoconfiguration address does not count — that is precisely what a host that has lost its DHCP lease ends up with. The filter stays off until the daemon is restarted; coming back automatically would just re-arm the same trap.

## 8. Observability and troubleshooting

The filter writes three kinds of line, all into `logs/nhp_server_xdp-<date>.log`.

### Per-packet lines

```
15:04:05 srv1 [NHP-PASS] REASON=NHP_DEFAULT SRC=203.0.113.9 DST=10.0.1.5 LEN=240 PROTO=UDP SPT=41337 DPT=62206 RELAY=0
15:04:05 srv1 [NHP-DROP] REASON=TCP_SYN_SCAN SRC=198.51.100.4 DST=10.0.1.5 LEN=60 PROTO=TCP SPT=51234 DPT=22 RELAY=0
```

`REASON=` is the action code from the table in chapter 4. `RELAY=1` means the source matched the whitelist.

### Per-minute totals: `[NHP-STAT]`

```
15:05:00 srv1 [NHP-STAT] VERDICT=DROP REASON=TCP_SYN_SCAN WINDOW=60s PKTS=1432 LOGGED=16 TOTAL=1432
```

`PKTS` is the delta over the window, `LOGGED` how many of those reached the event stream, `TOTAL` the count since the filter was attached. The counters are incremented for **every** packet, budget or no, so the totals are exact no matter how much reporting was dropped — which is the whole point of having them.

### Scan aggregation: `[NHP-SCAN]`

```
15:05:00 srv1 [NHP-SCAN] WINDOW=300s KIND=PORT PROTO=TCP DPT=22 PKTS=981
15:05:00 srv1 [NHP-SCAN] WINDOW=300s KIND=SRC SRC=198.51.100.4 PKTS=704
15:05:00 srv1 [NHP-SCAN] WINDOW=300s KIND=TOTAL UNIQ_SRC=31 UNIQ_PORT=4 PKTS=1902 LRU=0
```

This answers "what is hitting this host, from where" at a finer grain than a verdict class. `LRU=1` means the tables were at least 95% full and eviction is likely, so the Top-N list is approximate — the exact per-class totals still come from `[NHP-STAT]`.

### Why the log is bounded

Every packet the program sees is chosen by whoever is sending it, and the reader appends each event to a log file that is rotated by date but never capped. Without a limit, anyone who can reach the host can fill its disk by scanning it — which would take down the very daemon the filter protects. Three bounds apply, and none of them changes a verdict:

- **A per-class token bucket in the kernel.** Each action class gets a burst of 16 per second, per CPU, plus a global bucket of 8 per second per CPU across all classes. The per-class bucket stops a flood of one class from starving the records of another; the global one bounds the total, whatever mix of classes it is made of. Without it, the dozen or so reportable classes an IPv4 sender can reach, at a burst of sixteen each, times the number of CPUs, is hundreds of lines a second — i.e. gigabytes a day.
- **A daily byte budget in the writer.** 64 MiB of per-packet text per calendar day. When it is spent, the writer stops adding lines and says so — but the `[NHP-STAT]` summaries keep reporting exact totals.
- **A retention sweep.** 14 days or 256 MiB, whichever bites first; oldest files go first and today's file is never removed.

The bulk classes are also counted but never reported at all (`TCP_ESTABLISHED`, `UDP_ESTABLISHED`, `V6_ESTABLISHED`, `ICMPV6_CONTROL`, and every `SSH_RELAY` packet after the SYN). They scale with bytes transferred, not with events.

### When the filter is not attached

Every path that leaves the host unfiltered logs a line. A `Warning` is the loader failing; everything else is the config path deciding, on purpose, not to attach — at `Info` for a missing or disabled file, `Error` for an out-of-range value, and `Critical` for a whitelist that would close tcp/22:

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

The `Warning` is the loader's failure: a missing `nhp_server_xdp.o`, an object too old to carry `nhp_listen_port`, a kernel too old for XDP, an unmounted bpffs, or a missing capability. The object file is expected next to the binary, at `<exe dir>/etc/nhp_server_xdp.o`, and is produced by `make ebpf-objs`.

One more refusal is worth knowing, because it is about *this* host rather than about the filter:

```
CRIT   xdp config: refusing to attach the XDP ingress filter because the HTTP knock listener is enabled on 0.0.0.0:443 — the filter drops inbound TCP other than SSH from RelayIPs, so that service would be unreachable. Disable it, bind it to 127.0.0.1, or set Enabled = false in etc/xdp.toml
```

The same refusal covers a metrics endpoint bound off-loopback. The filter drops inbound TCP to any service on the host, not just to `nhp-serverd`, so the daemon refuses to attach while a listener of its own would be black-holed. Loopback binds are not a conflict — the program is attached to the default-route interface, so traffic to 127.0.0.1 never reaches it.


### Analysing the log offline

`endpoints/xdplog-report` parses these files into a report:

```bash
xdplog-report -in logs/nhp_server_xdp-2026-10-01.log -out report.md
```

Its `site` subcommand publishes a static summary site, and the exact packet totals come from the `[NHP-STAT]` lines rather than from the per-packet ones, which are a rate-limited sample by design.

## Further reading

- [Understand the Code](/code/) — the repository tour, including where the eBPF sources sit.
- [Cryptography](/cryptography/) — the cipher schemes the length floor in chapter 4 is derived from.
- [Protocol Reference](/protocol/) — the message header the XDP filter deliberately does not parse.
- [Features](/features/) — the default-deny property this layer implements.
- `terraform/demo/RUNBOOK.md` — the operational runbook: "nhp-server ingress filter (XDP)" and "Why each rule is there", which is the incident-driven version of chapters 4 and 7.
