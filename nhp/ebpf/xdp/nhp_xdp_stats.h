/* nhp_xdp_stats.h
 *
 * The XDP telemetry mechanism, shared by the two OpenNHP ingress filters.
 *
 * It started out inside nhp_server_xdp.c and was moved here verbatim so the
 * nhp-ac can report what it dropped in the same shape, with the same bounds:
 * a per-class token bucket over a perf ring, exact per-class counters, and
 * the per-port / per-source aggregation behind the [NHP-SCAN] summary.
 *
 * What is deliberately NOT here: the action codes (ACT_* / ACT_AC_*) and the
 * drop decision (action_is_drop()). Those belong to each program, because
 * each one has its own code space and its own idea of what a drop is. The
 * caller passes the verdict in as the `drop` flag.
 *
 * Include this AFTER vmlinux.h: it uses struct iphdr and struct ipv6hdr.
 */

#ifndef NHP_XDP_STATS_H
#define NHP_XDP_STATS_H

/* scan_port() below classifies a destination port by protocol, so it needs
 * these two. Both programs define them too; the guards keep a second
 * definition from being an error and let this header be included at any
 * point after vmlinux.h. */
#ifndef IPPROTO_TCP
#define IPPROTO_TCP 6
#endif
#ifndef IPPROTO_UDP
#define IPPROTO_UDP 17
#endif

/* NHP_FAMILY_* is the `family` field of nhp_event_t, not AF_INET: it says which
 * of the two address pairs below carries this event's addresses. */
#define NHP_FAMILY_V4 4
#define NHP_FAMILY_V6 6

struct nhp_event_t {
    __u64  timestamp;
    __u8   action;
    __be32 src_ip;
    __be32 dst_ip;
    __be16 src_port;
    __be16 dst_port;
    __u8   protocol;
    __be16 pkt_len;
    __u8   relay_hit;
    /* Which pair holds the addresses: NHP_FAMILY_V4 means src_ip/dst_ip and
     * the v6 fields are zero, NHP_FAMILY_V6 the reverse. The v6 fields were
     * added when IPv6 stopped being an unconditional pass -- a drop class whose
     * log lines have no source address is a blind spot, and a __be32 cannot
     * hold a v6 address. Parsed by readServerEvents in
     * nhp/utils/ebpf/engine_linux.go; serverEventByteSize there is sizeof this
     * struct, so a field added here has to be added there too. */
    __u8   family;
    __u8   src_ip6[16];
    __u8   dst_ip6[16];
} __attribute__((packed));

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(max_entries, 1024);
} nhp_events SEC(".maps");

/* Per-action, per-CPU token bucket over the perf ring.
 *
 * Every packet this program sees is chosen by whoever is sending it, and the
 * user-space reader appends each event to a log file that nhp/log rotates by
 * date but never caps or prunes. Without a limit, anyone who can reach the
 * host can fill its disk by scanning it -- which would take down the very
 * daemon this filter protects. So the reporting is bounded and the verdict is
 * not: the decision tree below runs unchanged, only the telemetry is dropped
 * once a class is over budget.
 *
 * The budget is per action code, so a flood of one class (a port scan, say)
 * cannot starve out the records of another (an SSH attempt from a non-relay
 * address, or a knock that passed). Per-CPU so the counter needs no atomics.
 *
 * A per-class budget alone is not a bound on the log, though, and that was the
 * first version's mistake. A remote sender picks the class: over IPv4 it can
 * reach about eleven reportable ones (TCP_SSH_OTHER, TCP_NHP_PORT, TCP_OTHER,
 * UDP_OTHER, UDP_SHORT, NON_TCP_UDP, IPV4_FRAGMENT, NHP_DEFAULT,
 * ICMP_FRAG_NEEDED, a spoofed 67->68 DHCP_CLIENT, NHP_RELAY from the subnet),
 * and spreading the flows across RSS queues fills every CPU's bucket. Eleven
 * classes x 16 x nr_cpus is hundreds of lines a second, i.e. gigabytes a day,
 * on a host whose whole event-log budget is 256 MiB. So a second bucket sits
 * across all of them: NHP_EVENT_GLOBAL_BURST caps the total, whatever mix of
 * classes it is made of, while the per-class bucket keeps that total from
 * being spent entirely on one. The per-CPU multiplier still applies to both --
 * what makes the *size* bound hard is serverLogDailyEventBytes on the writing
 * side (nhp/utils/ebpf/engine_linux.go); this is what keeps the ring and the
 * CPU cost of reporting bounded.
 *
 * NHP_EVENT_ACTIONS sizes the counter array and the per-class part of this
 * one, so it must stay above the highest action code (a code at or past it is
 * still judged, just neither counted nor rate limited). It carries a little
 * headroom over the codes in use; serverEventActions in
 * nhp/utils/ebpf/engine_linux.go mirrors it, and a test fails if the two drift
 * apart. The global bucket lives in one extra slot past the end. */
#define NHP_EVENT_ACTIONS 32
#define NHP_EVENT_WINDOW_NS 1000000000ULL
#define NHP_EVENT_BURST 16
#define NHP_EVENT_GLOBAL_BURST 8
#define NHP_EVENT_GLOBAL_SLOT NHP_EVENT_ACTIONS
#define NHP_EVENT_RATE_SLOTS (NHP_EVENT_ACTIONS + 1)

struct nhp_rate_state {
    __u64 window_start;
    __u32 count;
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, NHP_EVENT_RATE_SLOTS);
    __type(key, __u32);
    __type(value, struct nhp_rate_state);
} nhp_event_rate SEC(".maps");

/* What the budget above throws away, kept as numbers.
 *
 * A token bucket alone makes the log lie by omission: over budget, a flood and
 * a trickle look the same, and "how much of this is there" -- the first
 * question anyone asks of an ingress filter -- is unanswerable. These counters
 * are incremented for *every* packet, budget or no, so user space can print one
 * summary line per window per class (see reportServerStats in
 * nhp/utils/ebpf/engine_linux.go) and the total is never wrong no matter how
 * much reporting was dropped. `logged` is the subset that reached the ring, so
 * the difference is the suppressed count, stated rather than guessed at.
 *
 * Per-CPU, so ++ needs no atomics; user space sums the slices. Not pinned: only
 * this process reads them, same as nhp_events. */
struct nhp_action_stat {
    __u64 packets;
    __u64 logged;
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, NHP_EVENT_ACTIONS);
    __type(key, __u32);
    __type(value, struct nhp_action_stat);
} nhp_action_stats SEC(".maps");

/* Per-destination-port and per-source aggregation of dropped packets, so the
 * question "what is hitting this host, from where" has an answer finer than a
 * verdict class. Updated for drop classes only, for every dropped packet
 * whatever the reporting budget says (same "every packet" semantics as
 * nhp_action_stats). LRU so a 65k-port sweep cannot exhaust the table; user
 * space reports LRU=1 when it suspects eviction, and the exact per-class totals
 * stay in nhp_action_stats. Per-CPU, not pinned. Telemetry only: a failed
 * lookup or update is ignored and never influences a verdict. */
struct nhp_scan_port_key {
    __u8 proto;
    __u8 _pad;
    __u16 port; /* host byte order */
};

struct {
    __uint(type, BPF_MAP_TYPE_LRU_PERCPU_HASH);
    __uint(max_entries, 65536);
    __type(key, struct nhp_scan_port_key);
    __type(value, __u64);
} nhp_port_hits SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_PERCPU_HASH);
    __uint(max_entries, 65536);
    __type(key, __be32); /* IPv4 source only */
    __type(value, __u64);
} nhp_src_hits SEC(".maps");

static __always_inline void bump_hit(void *map, const void *key) {
    __u64 *v = bpf_map_lookup_elem(map, key);
    if (v) {
        (*v)++;
        return;
    }
    __u64 one = 1;
    bpf_map_update_elem(map, key, &one, BPF_NOEXIST);
}

/* saddr == 0 skips the source table (IPv6 packets). */
static __always_inline void scan_account(__u8 proto, __u16 dport_host, __be32 saddr) {
    struct nhp_scan_port_key pk = {};
    pk.proto = proto;
    pk.port = dport_host;
    bump_hit(&nhp_port_hits, &pk);
    if (saddr)
        bump_hit(&nhp_src_hits, &saddr);
}

/* Host-order destination port for TCP/UDP, 0 for anything else. */
static __always_inline __u16 scan_port(__u8 proto, __be16 dst_port) {
    if (proto == IPPROTO_TCP || proto == IPPROTO_UDP)
        return bpf_ntohs(dst_port);
    return 0;
}

/* Take one token from the bucket in `slot`, or report that it is empty for the
 * rest of this window. A slot the map cannot hand back is treated as having
 * budget: the rate limit is telemetry policy, never a verdict, so a map failure
 * must not start silently hiding packets. */
static __always_inline bool spend_token(__u32 slot, __u32 burst, __u64 now) {
    struct nhp_rate_state *st = bpf_map_lookup_elem(&nhp_event_rate, &slot);
    if (!st)
        return true;

    if (now - st->window_start >= NHP_EVENT_WINDOW_NS) {
        st->window_start = now;
        st->count = 1;
        return true;
    }
    if (st->count >= burst)
        return false;

    st->count++;
    return true;
}

/* Per class first, then across all classes. An event that clears its class but
 * not the global cap has spent a class token it did not use; that only makes
 * the class slightly stingier in a window the global cap has already closed,
 * and it keeps the common path to a single lookup. */
static __always_inline bool event_budget_available(__u8 action, __u64 now) {
    __u32 slot = action;
    if (slot >= NHP_EVENT_ACTIONS)
        return true;

    if (!spend_token(slot, NHP_EVENT_BURST, now))
        return false;

    return spend_token(NHP_EVENT_GLOBAL_SLOT, NHP_EVENT_GLOBAL_BURST, now);
}

/* Tick this action's total, which happens for every packet whatever the
 * reporting decision below is, and hand the caller the slot so it can tick
 * `logged` too if the event reaches the ring. */
static __always_inline struct nhp_action_stat *count_packet(__u8 action) {
    __u32 slot = action;
    struct nhp_action_stat *stat = NULL;

    if (slot < NHP_EVENT_ACTIONS)
        stat = bpf_map_lookup_elem(&nhp_action_stats, &slot);
    if (stat)
        stat->packets++;
    return stat;
}

/* Shared tail of the two recorders: stamp the event, spend a token, emit. */
static __always_inline void emit_event(void *ctx, struct nhp_event_t *ev,
                                       struct nhp_action_stat *stat) {
    __u64 now = bpf_ktime_get_ns();
    if (!event_budget_available(ev->action, now))
        return;
    ev->timestamp = now;

    if (bpf_perf_event_output(ctx, &nhp_events, BPF_F_CURRENT_CPU, ev, sizeof(*ev)) == 0 && stat)
        stat->logged++;
}

/* Count the packet against its action, and put an event on the perf ring if
 * the class is worth a line and has budget left.
 *
 * `report` is what keeps the log readable. A verdict is worth a line when it
 * says something the previous lines did not: a knock admitted, an SSH attempt
 * from an address that is not the relay, a scan. It is worth nothing when it
 * is the four-thousandth data segment of a connection whose one real decision
 * -- "the host opened this flow" -- was already logged, or was never a decision
 * at all because the kernel's socket table answered it. Those bulk classes
 * (TCP_ESTABLISHED, UDP_ESTABLISHED, V6_ESTABLISHED, ICMPV6_CONTROL and every
 * SSH_RELAY packet after the SYN) are passed report=false: they scale with
 * *bytes transferred* or with a subnet's background chatter, not with events,
 * so one `dnf update` or one CI scp of the release tarball would otherwise
 * write more log than a month of knocks. They are still counted, and the
 * per-window summary reports them.
 *
 * `drop` says whether the packet was dropped, which is what feeds the
 * per-port and per-source aggregation behind the [NHP-SCAN] summary. The
 * caller decides it, because what counts as a drop is the program's own
 * verdict and not this file's business.
 *
 * The verdict never depends on either argument -- only the telemetry does. */
static __always_inline void record_event(void *ctx, __u8 action, struct iphdr *iph,
                                         __be16 src_port, __be16 dst_port,
                                         __u8 relay_hit, bool report, bool drop) {
    struct nhp_action_stat *stat = count_packet(action);

    if (drop)
        scan_account(iph->protocol, scan_port(iph->protocol, dst_port), iph->saddr);

    if (!report)
        return;

    struct nhp_event_t ev = {};
    ev.action = action;
    ev.family = NHP_FAMILY_V4;
    ev.src_ip = iph->saddr;
    ev.dst_ip = iph->daddr;
    ev.src_port = src_port;
    ev.dst_port = dst_port;
    ev.protocol = iph->protocol;
    ev.pkt_len = iph->tot_len;
    ev.relay_hit = relay_hit;

    emit_event(ctx, &ev, stat);
}

/* The IPv6 recorder. `proto` is the next-header value the extension-header walk
 * arrived at, i.e. the actual L4 protocol, rather than ip6h->nexthdr, which is
 * the first extension header whenever there is one. relay_hit is not a
 * parameter: the whitelist is IPv4-only, so no v6 packet can ever hit it. */
static __always_inline void record_v6_event(void *ctx, __u8 action, struct ipv6hdr *ip6h,
                                            __u8 proto, __be16 src_port, __be16 dst_port,
                                            bool report, bool drop) {
    struct nhp_action_stat *stat = count_packet(action);

    if (drop)
        scan_account(proto, scan_port(proto, dst_port), 0);

    if (!report)
        return;

    struct nhp_event_t ev = {};
    ev.action = action;
    ev.family = NHP_FAMILY_V6;
    __builtin_memcpy(ev.src_ip6, ip6h->saddr.in6_u.u6_addr8, sizeof(ev.src_ip6));
    __builtin_memcpy(ev.dst_ip6, ip6h->daddr.in6_u.u6_addr8, sizeof(ev.dst_ip6));
    ev.src_port = src_port;
    ev.dst_port = dst_port;
    ev.protocol = proto;
    ev.pkt_len = ip6h->payload_len;

    emit_event(ctx, &ev, stat);
}

#endif /* NHP_XDP_STATS_H */
