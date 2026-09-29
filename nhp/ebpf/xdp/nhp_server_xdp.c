/* Ingress filter for the nhp-server host.
 *
 * Unlike the AC's nhp_ebpf_xdp.c, which opens ports per knock, the server's
 * exposure is fixed and tiny: the NHP UDP knock port, and SSH from the relay
 * jump host. Everything else is dropped at the driver, before the kernel's
 * network stack sees it. There are therefore no whitelist maps, no conn_track
 * and no TC egress program here -- the decision is a pure function of
 * (src ip, l4 proto, dst port, udp length), which is why this program can stay
 * a few dozen lines and share nothing with the AC's schema.
 *
 * It deliberately does NOT parse the NHP protocol. Identity is still decided in
 * user space by the Noise handshake (nhp/core/packet.go::RecvPrecheck); the
 * only NHP-shaped test here is a minimum datagram length, which costs nothing
 * and turns the knock port into a non-answer for ordinary scanners.
 *
 * Map names carry an `nhp_` prefix so a host that somehow ran both this and the
 * AC object would not collide on LIBBPF_PIN_BY_NAME pins in /sys/fs/bpf.
 */

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>

#define ETH_P_ARP    0x0806
#define ETH_P_IP     0x0800
#define ETH_P_IPV6   0x86DD
#define IPPROTO_TCP  6
#define IPPROTO_UDP  17

#define SSH_PORT 22
#define NHP_PORT 62206

/* Shortest NHP datagram on the wire: the 240-byte NHP_KPL header of the
 * curve25519 cipher suite (nhp/core/packet.go). The gmsm suite's 304 bytes
 * clears it too, so one threshold covers both. Anything shorter cannot be a
 * knock no matter what it decrypts to, so it is dropped without touching the
 * payload. The length compared is the UDP header's own length field, i.e.
 * header + payload. */
#define NHP_MIN_UDP_LEN 240

/* action codes reported to user space; see serverengine.go, which formats
 * them, and the table in the design doc. */
#define ACT_DROP_OTHER          0
#define ACT_SSH_RELAY           1
#define ACT_NHP_RELAY           2
#define ACT_NHP_DEFAULT         3
#define ACT_DROP_TCP_SSH_OTHER  10
#define ACT_DROP_TCP_NHP        11
#define ACT_DROP_TCP_OTHER      12
#define ACT_DROP_UDP_OTHER      13
#define ACT_DROP_UDP_SHORT      14
#define ACT_DROP_NONUDP         15

/* Source addresses allowed to reach SSH, and allowed to reach the knock port
 * without meeting the length floor. Driven from user space by
 * endpoints/server/ebpf/serverengine.go::UpdateRelayIPs, which mirrors
 * etc/xdp.toml into it on every reload.
 *
 * Keys are __be32, i.e. wire order, so they compare directly against
 * iph->saddr with no byte swapping in the hot path. LRU rather than plain hash
 * only so that a wedged user space can never make an insert fail; the map is
 * always far below max_entries in practice. */
struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 4096);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
    __type(key, __be32);
    __type(value, __u8);
} nhp_relay_ips SEC(".maps");

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
 * address, or a knock that passed). Per-CPU so the counter needs no atomics;
 * the effective ceiling is NHP_EVENT_BURST * nr_cpus per second per class,
 * which is ample for diagnosis and bounded regardless of offered load. */
#define NHP_EVENT_ACTIONS 16
#define NHP_EVENT_WINDOW_NS 1000000000ULL
#define NHP_EVENT_BURST 64

struct nhp_rate_state {
    __u64 window_start;
    __u32 count;
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, NHP_EVENT_ACTIONS);
    __type(key, __u32);
    __type(value, struct nhp_rate_state);
} nhp_event_rate SEC(".maps");

static __always_inline bool event_budget_available(__u8 action, __u64 now) {
    __u32 slot = action;
    if (slot >= NHP_EVENT_ACTIONS)
        return true;

    struct nhp_rate_state *st = bpf_map_lookup_elem(&nhp_event_rate, &slot);
    if (!st)
        return true;

    if (now - st->window_start >= NHP_EVENT_WINDOW_NS) {
        st->window_start = now;
        st->count = 1;
        return true;
    }
    if (st->count >= NHP_EVENT_BURST)
        return false;

    st->count++;
    return true;
}

static __always_inline void submit_event(void *ctx, __u8 action, struct iphdr *iph,
                                         __be16 src_port, __be16 dst_port, __u8 relay_hit) {
    __u64 now = bpf_ktime_get_ns();
    if (!event_budget_available(action, now))
        return;

    struct nhp_event_t ev = {};
    ev.timestamp = bpf_ktime_get_ns();
    ev.action = action;
    ev.src_ip = iph->saddr;
    ev.dst_ip = iph->daddr;
    ev.src_port = src_port;
    ev.dst_port = dst_port;
    ev.protocol = iph->protocol;
    ev.pkt_len = iph->tot_len;
    ev.relay_hit = relay_hit;

    bpf_perf_event_output(ctx, &nhp_events, BPF_F_CURRENT_CPU, &ev, sizeof(ev));
}

static __always_inline bool is_relay_src(__be32 saddr) {
    return bpf_map_lookup_elem(&nhp_relay_ips, &saddr) != NULL;
}

SEC("xdp")
int xdp_server_prog(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_DROP;

    switch (bpf_ntohs(eth->h_proto)) {
        /* ARP and IPv6 pass unreported: ARP keeps the host on its subnet, and
         * the demo hosts have no IPv6 service to filter -- reporting either
         * would drown the perf ring in neighbour traffic. */
        case ETH_P_ARP:  return XDP_PASS;
        case ETH_P_IPV6: return XDP_PASS;
        case ETH_P_IP:   break;
        default:         return XDP_DROP;
    }

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return XDP_DROP;
    if (iph->ihl < 5)
        return XDP_DROP;

    /* L4 starts after the IP options, i.e. at ihl * 4 -- not at iph + 1, which
     * is only the same when there are none. */
    if (iph->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)iph + (iph->ihl * 4);
        if ((void *)(tcp + 1) > data_end)
            return XDP_DROP;

        if (tcp->dest == bpf_htons(SSH_PORT)) {
            if (is_relay_src(iph->saddr)) {
                submit_event(ctx, ACT_SSH_RELAY, iph, tcp->source, tcp->dest, 1);
                return XDP_PASS;
            }
            submit_event(ctx, ACT_DROP_TCP_SSH_OTHER, iph, tcp->source, tcp->dest, 0);
            return XDP_DROP;
        }
        /* NHP is UDP-only, so a TCP connect to the knock port is a scanner
         * fingerprinting the host. Reported under its own action so that
         * traffic is countable separately from ordinary port scans. */
        if (tcp->dest == bpf_htons(NHP_PORT)) {
            submit_event(ctx, ACT_DROP_TCP_NHP, iph, tcp->source, tcp->dest, 0);
            return XDP_DROP;
        }
        submit_event(ctx, ACT_DROP_TCP_OTHER, iph, tcp->source, tcp->dest, 0);
        return XDP_DROP;
    }

    if (iph->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)iph + (iph->ihl * 4);
        if ((void *)(udp + 1) > data_end)
            return XDP_DROP;

        if (udp->dest != bpf_htons(NHP_PORT)) {
            submit_event(ctx, ACT_DROP_UDP_OTHER, iph, udp->source, udp->dest, 0);
            return XDP_DROP;
        }
        /* The relay forwards agent knocks, including the short control
         * packets of the handshake, so it is exempt from the length floor. */
        if (is_relay_src(iph->saddr)) {
            submit_event(ctx, ACT_NHP_RELAY, iph, udp->source, udp->dest, 1);
            return XDP_PASS;
        }
        if (bpf_ntohs(udp->len) < NHP_MIN_UDP_LEN) {
            submit_event(ctx, ACT_DROP_UDP_SHORT, iph, udp->source, udp->dest, 0);
            return XDP_DROP;
        }
        submit_event(ctx, ACT_NHP_DEFAULT, iph, udp->source, udp->dest, 0);
        return XDP_PASS;
    }

    /* ICMP included: the server answers no pings from the internet, and an
     * operator who needs one reaches the host over the relay's SSH path. */
    submit_event(ctx, ACT_DROP_NONUDP, iph, 0, 0, 0);
    return XDP_DROP;
}

char _license[] SEC("license") = "Dual BSD/GPL";
