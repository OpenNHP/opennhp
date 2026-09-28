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

/*
 * Reply window for connections the AC itself initiates (dnf, certbot, NTP,
 * ...). The XDP ingress program drops everything that is not whitelisted, so
 * outbound flows need their replies to be let back in -- the equivalent of
 * iptables' `--ctstate ESTABLISHED,RELATED -j ACCEPT`.
 *
 * This is NOT the knock TTL. Knock-authorised flows are governed solely by the
 * expiry that the user space knock path writes into the whitelist maps; the two
 * gates in tc_egress_prog() below make sure this program never creates or
 * refreshes an entry for such a flow.
 */
#define EGRESS_CONN_TTL_NS (180 * 1000000000ULL)

SEC("tc/egress")
int tc_egress_prog(struct __sk_buff *ctx)
{

    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    struct ethhdr *eth = data;

    if (data + sizeof(*eth) > data_end)
        return TC_ACT_OK;

    if (bpf_ntohs(eth->h_proto) != ETH_P_IP)
        return TC_ACT_OK;

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return TC_ACT_OK;

    if (iph->ihl < 5 || iph->version != 4)
        return TC_ACT_OK;

    __be32 src_ip = iph->saddr;
    __be32 dst_ip = iph->daddr;
    __u8 protocol = iph->protocol;

    __be16 sport = 0, dport = 0;
    bool is_tcp = false, is_syn = false;

    if (protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)iph + (iph->ihl * 4);
        if ((void *)(tcp + 1) > data_end)
            return TC_ACT_OK;
        sport = tcp->source;
        dport = tcp->dest;
        is_tcp = true;
        /* A pure SYN (no ACK) is the only packet that opens a connection the
         * AC itself initiates. */
        is_syn = tcp->syn && !tcp->ack;
    } else if (protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)iph + (iph->ihl * 4);
        if ((void *)(udp + 1) > data_end)
            return TC_ACT_OK;
        sport = udp->source;
        dport = udp->dest;
    } else if (protocol == IPPROTO_ICMP) {
        sport = 0;
        dport = 0;
    } else {
        return TC_ACT_OK;
    }

    /*
     * Gate 1: is this the reply side of an inbound flow that XDP already
     * admitted? Rebuild the ingress key exactly as xdp_white_prog() does for
     * the peer -> AC direction and look it up. A hit means the peer opened this
     * flow (knock, SSH, ...), so its lifetime belongs to the knock TTL alone --
     * do not create or refresh anything here.
     */
    struct ipv4_ct_tuple ingress_key;
    __builtin_memset(&ingress_key, 0, sizeof(ingress_key));
    ingress_key.saddr = dst_ip;   /* peer  -> the ingress packet's source */
    ingress_key.daddr = src_ip;   /* AC    -> the ingress packet's destination */
    ingress_key.sport = dport;
    ingress_key.dport = sport;
    ingress_key.nexthdr = protocol;
    ingress_key.flags = CT_DIR_INGRESS;

    if (bpf_map_lookup_elem(&conn_track, &ingress_key))
        return TC_ACT_OK;

    /*
     * Gate 2: for TCP, only a pure SYN creates an entry. The AC acting as a
     * server never sends a pure SYN from its listening port, so even if gate 1
     * misses (the ingress entry may already have been expired and deleted by
     * XDP) a knock-authorised flow can never be kept alive from here.
     */
    if (is_tcp && !is_syn)
        return TC_ACT_OK;

    struct ipv4_ct_tuple egress_key;
    __builtin_memset(&egress_key, 0, sizeof(egress_key));
    egress_key.saddr = src_ip;
    egress_key.daddr = dst_ip;
    egress_key.sport = sport;
    egress_key.dport = dport;
    egress_key.nexthdr = protocol;
    egress_key.flags = CT_DIR_EGRESS;

    __u64 now = bpf_ktime_get_ns();
    struct conn_value egress_val = {
        .timestamp = now,
        .last_timestamp = now,
        .ttl_ns = EGRESS_CONN_TTL_NS,
        .state = CT_NEW,
        .flags = CT_FLAG_NONE,
        .rx_packets = 0,
        .tx_packets = 1,
    };

    /*
     * TCP: BPF_ANY, a retransmitted SYN legitimately restarts the window.
     * UDP/ICMP: BPF_NOEXIST, otherwise every outgoing packet of a long-lived
     * flow would push the absolute expiry further out indefinitely.
     */
    bpf_map_update_elem(&conn_track, &egress_key, &egress_val,
                        is_tcp ? BPF_ANY : BPF_NOEXIST);

    return TC_ACT_OK;
}


char _license[] SEC("license") = "Dual BSD/GPL";
