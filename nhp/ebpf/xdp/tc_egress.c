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

/*
 * Does the peer we are sending to hold a knock entry for this exact port and
 * protocol? If so, this packet is the reply side of a service the AC is
 * serving, not a connection the AC opened, and its lifetime belongs to the
 * knock TTL alone.
 *
 * `src_ip`/`dst_ip` are this egress packet's, so the whitelist keys -- which
 * are written for the peer -> AC direction -- are built the other way round,
 * with our source port as the peer's destination port. Expiry is not checked on
 * purpose: an entry that exists at all, live or stale, marks the peer.
 *
 * Only the two maps whose key carries *both* the peer address and the port are
 * consulted. `sdwhitelist` and `port_list` are keyed on the address alone and
 * `protocol_port` on the port alone, so a single stale entry in any of them
 * would suppress tracking for unrelated flows -- including nhp-acd's own UDP
 * channel to the nhp-server, whose replies only get back in because of the
 * entry written below. Losing that would take the AC off the air, which is a
 * far worse failure than the narrow case those lookups would cover, and gate 3
 * plus the hard cap on datagram entries already bound it.
 */
static __always_inline bool peer_is_knocking(__be32 src_ip, __be32 dst_ip,
                                             __be16 sport, __u8 protocol)
{
    struct whitelist_key wl_key = {
        .src_ip = dst_ip,
        .dst_ip = src_ip,
        .dst_port = sport,
        .protocol = protocol,
    };
    if (bpf_map_lookup_elem(&spp, &wl_key))
        return true;

    struct src_port_list_key sp_key = {
        .src_ip = dst_ip,
        .dst_port = sport,
    };
    if (bpf_map_lookup_elem(&src_port, &sp_key))
        return true;

    return false;
}

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

    /* Only TCP and UDP are tracked. ICMP needs nothing from here: XDP passes
     * echo replies unconditionally, and it never puts IPPROTO_ICMP in a
     * conn_track key, so entries written for it could never be matched. */
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
     * Gates 2-4 catch what gate 1 cannot: its ingress entry may already have
     * been expired and deleted by XDP, or evicted from the LRU map, and the AC
     * may well answer afterwards (a retransmit, a keepalive, a reply to a
     * datagram that was already queued). Without them that answer would create
     * an EGRESS entry for a knock-protected flow, and the peer's datagrams
     * would then be admitted by the reverse lookup in XDP for as long as that
     * entry lives -- long after the knock closed.
     */
    if (is_tcp) {
        /*
         * Gate 2, TCP: only a pure SYN creates an entry. The AC acting as a
         * server never sends a pure SYN from a listening port, so this alone
         * is enough for TCP.
         */
        if (!is_syn)
            return TC_ACT_OK;
    } else {
        /*
         * Gate 3, UDP: there is no SYN to key off, so use the source port. A
         * socket the AC opens as a client (dnf, chrony, a resolver) binds an
         * ephemeral port; a service it listens on does not. This is the UDP
         * equivalent of "servers don't originate".
         */
        if (bpf_ntohs(sport) < NHP_EPHEMERAL_PORT_MIN)
            return TC_ACT_OK;

        /*
         * Gate 4, UDP: and in case a knock-protected service does listen on a
         * port inside the ephemeral range (QUIC/HTTP3, WireGuard, a game
         * server), skip any peer that holds a knock entry for it.
         *
         * Gates 3 and 4 are heuristics where gate 2 is exact, so the XDP side
         * backs them up: it never refreshes a non-TCP EGRESS entry, and one
         * that does slip through therefore dies NHP_CT_EGRESS_IDLE_TTL_NS
         * after it was created instead of sliding forward indefinitely.
         */
        if (peer_is_knocking(src_ip, dst_ip, sport, protocol))
            return TC_ACT_OK;
    }

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
        .ttl_ns = NHP_CT_EGRESS_IDLE_TTL_NS,
        .state = CT_NEW,
        .flags = CT_FLAG_NONE,
        .rx_packets = 0,
        .tx_packets = 1,
    };

    /*
     * TCP: BPF_ANY, a retransmitted SYN legitimately restarts the window.
     * UDP: BPF_NOEXIST, so an outgoing packet cannot push an existing entry's
     * expiry out. Together with XDP not refreshing non-TCP EGRESS entries this
     * makes NHP_CT_EGRESS_IDLE_TTL_NS a hard cap for datagram flows.
     */
    bpf_map_update_elem(&conn_track, &egress_key, &egress_val,
                        is_tcp ? BPF_ANY : BPF_NOEXIST);

    return TC_ACT_OK;
}


char _license[] SEC("license") = "Dual BSD/GPL";
