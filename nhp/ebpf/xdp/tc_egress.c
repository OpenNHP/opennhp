#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>
#include "nhp_maps.h"

#define TC_ACT_UNSPEC         (-1)
#define TC_ACT_OK               0
#define TC_ACT_SHOT             2
#define TC_ACT_STOLEN           4

#define ETH_P_IP    0x0800
#define ETH_P_IPV6  0x86DD
#define IPPROTO_TCP 6
#define IPPROTO_UDP 17
#define IPPROTO_ICMP 1
#define MAX_ENTRIES 1000000

/* Egress conn-track entries are kept for this long. This is the upper
 * bound on how long an AC-initiated outbound flow's return packets
 * can keep flowing without re-knocking; the XDP reverse-tuple lookup
 * will pick up entries created here.
 *
 * Important: this is NOT related to the knock TTL (which is bounded
 * by resource.toml's OpenTime + ACOpenCompensationTime). The two-gate
 * logic in tc_egress_prog() makes sure knock flows never get an
 * egress entry written for them. */
#define EGRESS_CONN_TTL_NS (180ULL * 1000000000ULL)

/* Local copy of reverseTuple; the one in nhp_ebpf_xdp.c is
 * static __always_inline and not visible across translation units. */
static __always_inline void reverseTuple(struct ipv4_ct_tuple *key)
{
    __u32 tmp_ip = key->daddr;
    __u16 tmp_port = key->dport;
    key->flags = !key->flags;
    key->daddr = key->saddr;
    key->saddr = tmp_ip;
    key->dport = key->sport;
    key->sport = tmp_port;
}

/* Insert an egress-direction conn_track entry for the given five-tuple.
 * Uses BPF_NOEXIST so an existing entry's absolute expire time is never
 * extended — that property is what guarantees we cannot re-introduce
 * the "knock TTL never expires" bug from this program. */
static __always_inline void insert_egress_entry(struct ipv4_ct_tuple *key)
{
    __u64 now = bpf_ktime_get_ns();
    struct conn_value cv = {
        .timestamp      = now,
        .last_timestamp = now,
        .ttl_ns         = EGRESS_CONN_TTL_NS,
        .state          = CT_NEW,
        .flags          = CT_FLAG_NONE,
        .rx_packets     = 0,
        .tx_packets     = 1,
    };
    bpf_map_update_elem(&conn_track, key, &cv, BPF_NOEXIST);
}

/* Egress program: track AC-initiated outbound flows so their return
 * packets can be allowed back in by the XDP reverse-tuple lookup.
 *
 * Two gates prevent re-introducing the historical "knock TTL never
 * expires" bug:
 *
 *   Gate 1 (every protocol): if a reverse-tuple lookup (i.e. ingress
 *   direction) hits a conn_track entry, this packet IS the return
 *   path of an XDP-allowed ingress flow. We must NOT add a fresh
 *   egress entry — doing so would either extend the life of a knock
 *   flow past its TTL or consume a slot reserved for genuine AC
 *   outbound flows.
 *
 *   Gate 2 (TCP only): only a pure SYN from AC is treated as AC
 *   initiating a new outbound flow. Anything else (ACK / FIN / RST)
 *   is either already tracked in conn_track or should not be
 *   tracked at all.
 *
 * We never write spp / sdwhitelist / any whitelist map — that was
 * the historical bug and we are not repeating it. */
SEC("tc/egress")
int tc_egress_prog(struct __sk_buff *ctx)
{
    struct ethhdr eth;
    if (bpf_skb_load_bytes(ctx, 0, &eth, sizeof(eth)) < 0)
        return TC_ACT_OK;
    if (eth.h_proto != bpf_htons(ETH_P_IP))
        return TC_ACT_OK; /* IPv6 etc. are not relevant here. */

    struct iphdr iph;
    if (bpf_skb_load_bytes(ctx, sizeof(eth), &iph, sizeof(iph)) < 0)
        return TC_ACT_OK;
    if (iph.ihl < 5)
        return TC_ACT_OK;

    __u32 l4_off = sizeof(eth) + (__u32)iph.ihl * 4;

    struct ipv4_ct_tuple egress_key = {
        .saddr   = iph.saddr,
        .daddr   = iph.daddr,
        .nexthdr = iph.protocol,
        .flags   = CT_DIR_EGRESS,
    };

    switch (iph.protocol) {
    case IPPROTO_TCP: {
        struct tcphdr tcp;
        if (bpf_skb_load_bytes(ctx, l4_off, &tcp, sizeof(tcp)) < 0)
            return TC_ACT_OK;
        egress_key.sport = tcp.source;
        egress_key.dport = tcp.dest;

        /* Gate 1: ingress direction lookup. */
        struct ipv4_ct_tuple ing_key = egress_key;
        reverseTuple(&ing_key);
        if (bpf_map_lookup_elem(&conn_track, &ing_key))
            return TC_ACT_OK;

        /* Gate 2: TCP — only pure SYN is AC-initiating a new flow.
         * Anything with ACK / FIN / RST is already (or should remain)
         * tracked through the XDP path; anything without SYN is not
         * a fresh outbound flow. */
        if (tcp.ack || tcp.fin || tcp.rst)
            return TC_ACT_OK;
        if (!tcp.syn)
            return TC_ACT_OK;

        insert_egress_entry(&egress_key);
        return TC_ACT_OK;
    }
    case IPPROTO_UDP: {
        struct udp_port_pair {
            __be16 sport;
            __be16 dport;
        } ports;
        if (bpf_skb_load_bytes(ctx, l4_off, &ports, sizeof(ports)) < 0)
            return TC_ACT_OK;
        egress_key.sport = ports.sport;
        egress_key.dport = ports.dport;

        /* Gate 1: ingress direction lookup. */
        struct ipv4_ct_tuple ing_key = egress_key;
        reverseTuple(&ing_key);
        if (bpf_map_lookup_elem(&conn_track, &ing_key))
            return TC_ACT_OK;

        insert_egress_entry(&egress_key);
        return TC_ACT_OK;
    }
    case IPPROTO_ICMP: {
        /* sport = dport = 0 (matches XDP convention). No L4 ports to
         * read; we still must run Gate 1 to avoid double-tracking
         * ICMP echo flows. */
        struct ipv4_ct_tuple ing_key = egress_key;
        reverseTuple(&ing_key);
        if (bpf_map_lookup_elem(&conn_track, &ing_key))
            return TC_ACT_OK;

        insert_egress_entry(&egress_key);
        return TC_ACT_OK;
    }
    default:
        return TC_ACT_OK;
    }
}

char _license[] SEC("license") = "Dual BSD/GPL";