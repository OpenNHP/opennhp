#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>

#include "nhp_maps.h"

#define ETH_P_ARP 0x0806
#define ETH_P_IP    0x0800
#define IPPROTO_ICMP 1
#define IPPROTO_TCP 6
#define ICMP_ECHO 8
#define ICMP_ECHOREPLY 0
#define ETH_P_IPV6   0x86DD
#define IPPROTO_UDP 17
#define MIN_PORT 0
#define MAX_PORT 65535
#define DNS_PORT 53
#define DHCP_PORT_R 67
#define DHCP_PORT_O 68

/* The whitelist maps, conn_track and their key/value structs live in
 * nhp_maps.h -- they are shared with tc_egress.c through their pins. */

struct event_t {
    __u64 timestamp;    
    __u8 action;        
    __be32 src_ip;
    __be32 dst_ip;
    __be16 src_port;    
    __be16 dst_port;    
    __u8 protocol;      
    __be16 len;
} __attribute__((packed));

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(max_entries, 1024);
} events SEC(".maps");

// Submit the event details including source and destination IP addresses, source and destination ports, protocol, and total packet length to the user space.
static __always_inline int submit_event(void *ctx, __u8 action, __be32 src_ip, __be32 dst_ip, __be16 src_port, __be16 dst_port, __u8 protocol, __be16 len) {
    struct event_t ev = {};
    ev.timestamp = bpf_ktime_get_ns();
    ev.action = action; // 0 = DENY, 1 = ACCEPT
    ev.src_ip = src_ip;
    ev.dst_ip = dst_ip;
    ev.src_port = src_port;
    ev.dst_port = dst_port;
    ev.protocol = protocol;
    ev.len = len;

    // Submit the event details to the user space Perf Buffer.
    return bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, &ev, sizeof(ev));
}

static __always_inline void reverseTuple(struct ipv4_ct_tuple *key) {
    __u32 tmp_ip = key->daddr;
    __u16 tmp_port = key->dport;
    key->flags = !key->flags;
    key->daddr = key->saddr;
    key->saddr = tmp_ip;
    key->dport = key->sport;
    key->sport = tmp_port;
}

static __always_inline bool check_conn_expiry(struct conn_value *val) {
    __u64 now = bpf_ktime_get_ns();
    return (now > val->timestamp + val->ttl_ns);
}

/* Build the conn_track entry for a flow the whitelists just admitted. `ttl_ns`
 * is an idle window, not the knock's remaining lifetime: once a flow exists it
 * is governed by activity, the way netfilter's `--state ESTABLISHED` accept
 * governs it in FilterMode 0. The knock TTL still decides which *new* flows may
 * be created, and this entry's key carries the peer's source port, so keeping
 * it alive can only ever keep this one connection alive. */
static __always_inline void new_ingress_conn(struct ipv4_ct_tuple *ct_key, __u8 protocol, __u64 now) {
    struct conn_value new_val = {
        .timestamp = now,
        .last_timestamp = now,
        .ttl_ns = nhp_ct_idle_ttl_ns(protocol),
        .state = CT_ESTABLISHED,
        .flags = CT_FLAG_NONE,
        .rx_packets = 1,
        .tx_packets = 0,
    };
    bpf_map_update_elem(&conn_track, ct_key, &new_val, BPF_ANY);
}

SEC("xdp")
static __always_inline int xdp_white_prog(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    struct ipv4_ct_tuple ct_key = {};
    struct ethhdr *eth = data;
    
    if (data + sizeof(*eth) > data_end) {
        return XDP_DROP;
    }   
    
    if ((void *)(eth + 1) > data_end)
        return XDP_DROP;

    switch (bpf_ntohs(eth->h_proto)) {
        case ETH_P_ARP:  return XDP_PASS;
        case ETH_P_IP:   break;
        case ETH_P_IPV6: return XDP_PASS;
        default:         return XDP_DROP;
    }

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return XDP_DROP;

    if (iph->ihl < 5)
        return XDP_DROP;

    /* The L4 header starts after the IP options, i.e. at ihl * 4 -- not at
     * iph + 1, which is only the same thing when there are no options. The
     * conn_track keys built here have to line up byte for byte with the ones
     * tc_egress.c builds (it reads ihl * 4 too), or the gates there miss. */
    ct_key.nexthdr = iph->protocol;
    if (iph->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)iph + (iph->ihl * 4);
        if ((void *)(tcp + 1) > data_end) {
            return XDP_DROP;
        }
        ct_key.sport = tcp->source;
        ct_key.dport = tcp->dest;

        if (tcp->dest == bpf_htons(22)) {
            return XDP_PASS;
        }
    } else if (iph->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)iph + (iph->ihl * 4);
        if ((void *)(udp + 1) > data_end)
            return XDP_DROP;
        ct_key.sport = udp->source;
        ct_key.dport = udp->dest;
    }

    if (iph->protocol == IPPROTO_UDP &&
        (ct_key.dport == bpf_htons(DHCP_PORT_R) || ct_key.dport == bpf_htons(DHCP_PORT_O) || ct_key.sport == bpf_htons(DNS_PORT))) {
        return XDP_PASS;
    }
    __u64 now = bpf_ktime_get_ns();

    // ICMP
    if (iph->protocol == IPPROTO_ICMP) {
        struct icmphdr *icmp = (void *)iph + (iph->ihl * 4);
        if ((void *)(icmp + 1) > data_end)
            return XDP_DROP;
        struct icmpwhitelist_key icmpkey = {
            .src_ip = iph->saddr,
            .dst_ip = iph->daddr,
        };
        //only processes ICMP Echo Requests (type 8, code 0) and ICMP Echo Replies (type 0, code 0)
        if ((icmp->type == ICMP_ECHO && icmp->code == 0) || 
            (icmp->type == ICMP_ECHOREPLY && icmp->code == 0)) {
            
            if (icmp->type == ICMP_ECHO) {
                //Lookup icmpwhitelist entry
                struct icmpwhitelist_value *iw_val = bpf_map_lookup_elem(&icmpwhitelist, &icmpkey);
                if (!iw_val) {
                    return XDP_DROP;
                }   
                __u64 now = bpf_ktime_get_ns();
                // Check if whitelist entry has expired
                if (iw_val->expire_time < now) {
                    bpf_map_delete_elem(&icmpwhitelist, &icmpkey);
                    return XDP_DROP;
                }
                // Check if source IP is icmpwhitelisted and allowed
                if (iw_val->allowed == 1) {
                    return XDP_PASS;
                }
            }  else {
                return XDP_PASS;
            }
        }
    }

    ct_key.saddr = iph->saddr;
    ct_key.daddr = iph->daddr;
    ct_key.flags = CT_DIR_INGRESS;
    struct conn_value *existing_val;

    /*
     * Forward (CT_DIR_INGRESS) hit: a flow the peer opened and that we admitted
     * via one of the whitelists while its knock was valid. It now lives on its
     * idle window -- `timestamp` is pushed forward by every packet -- so an
     * established session is not cut mid-stream when the knock expires. That is
     * what FilterMode 0 does too: the AC's iptables baseline accepts
     * `-m state --state ESTABLISHED` ahead of the per-knock ipset, so expiry
     * there blocks new connections only.
     *
     * This does not hold the door open, because the key includes the peer's
     * source port: refreshing keeps this one connection alive and nothing else.
     * A new connection gets a new source port, misses here, and has to satisfy
     * the whitelist lookups below -- which only user space writes.
     *
     * Once idle past its window the entry is dropped and the packet falls
     * through to those lookups, so a flow that goes quiet and comes back needs
     * a knock that is still valid.
     */
    existing_val = bpf_map_lookup_elem(&conn_track, &ct_key);
    if (existing_val) {
        if (check_conn_expiry(existing_val)) {
            /* Idle out. Fall through to the whitelist lookups below: if the
             * knock is still valid the flow is simply admitted again. */
            bpf_map_delete_elem(&conn_track, &ct_key);
        } else {
            struct conn_value new_val = *existing_val;
            new_val.tx_packets++;
            new_val.timestamp = now;
            new_val.last_timestamp = now;
            bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_EXIST);
            return XDP_PASS;
        }
    }
    reverseTuple(&ct_key);
    /*
     * Reverse (CT_DIR_EGRESS) hit: the reply to a connection the AC itself
     * opened, whose entry was created by tc_egress.c. The gates there are what
     * keep a knock-authorised flow from ever having an EGRESS entry, so this
     * branch is not supposed to see one -- and how long it may live is decided
     * by how much the gates can be trusted for that protocol:
     *
     *   TCP  -- the pure-SYN gate is exact (a listening socket never sends
     *           one), so `timestamp` is refreshed and the TTL becomes a sliding
     *           idle window. tc_egress.c writes the entry once, on the SYN, so
     *           without this a download longer than the TTL would be cut off.
     *   else -- the UDP gates are heuristics (ephemeral source port, peer holds
     *           no knock entry). Refreshing would let a peer slide the window
     *           forward for as long as it keeps sending, so it is deliberately
     *           not done: an entry that slips through the gates still dies
     *           NHP_CT_EGRESS_IDLE_TTL_NS after it was created. The AC's own
     *           outbound datagram flows (NTP, its UDP channel to the
     *           nhp-server) are short enough not to notice, and tc_egress.c
     *           writes them with BPF_NOEXIST so an outgoing packet cannot push
     *           the cap out either.
     */
    existing_val = bpf_map_lookup_elem(&conn_track, &ct_key);
    if (existing_val) {
        if (check_conn_expiry(existing_val)) {
            bpf_map_delete_elem(&conn_track, &ct_key);
            reverseTuple(&ct_key);
            return XDP_DROP;
        }
        struct conn_value new_val = *existing_val;
        new_val.rx_packets++;
        if (ct_key.nexthdr == IPPROTO_TCP)
            new_val.timestamp = now;
        new_val.last_timestamp = now;
        bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_EXIST);
        reverseTuple(&ct_key);
        return XDP_PASS;
    }
    reverseTuple(&ct_key);

    struct whitelist_key key = {
        .src_ip = iph->saddr,
        .dst_ip = iph->daddr,
        .dst_port = ct_key.dport,
        .protocol = iph->protocol
    };

    struct sdwhitelist_key sdkey = {
        .src_ip = iph->saddr,
        .dst_ip = iph->daddr
    };

    struct src_port_list_key spkey = {
        .src_ip = iph->saddr,
        .dst_port = ct_key.dport
    };

    struct port_list_key pl_key = {
        .src_ip = iph->saddr,
        .min_port = MIN_PORT,
        .max_port = MAX_PORT
    };
    struct protocol_port_key pp_key = {
        .dst_port = ct_key.dport,
        .protocol = iph->protocol
    };

    //Lookup whitelist entry
    struct whitelist_value *w_val = bpf_map_lookup_elem(&spp, &key);
    //Lookup sdwhitelist entry
    struct sdwhitelist_value *sd_val = bpf_map_lookup_elem(&sdwhitelist, &sdkey);
    //Lookup src_port_list entry
    struct src_port_list_value *sp_val = bpf_map_lookup_elem(&src_port, &spkey);
    //Lookup port_list entry
    struct port_list_value *pl_val= bpf_map_lookup_elem(&port_list, &pl_key);
    //Lookup protocol_port entry
    struct protocol_port_value *pp_val= bpf_map_lookup_elem(&protocol_port, &pp_key);

    if (w_val) {
        __u64 expire_time = w_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&spp, &key);
            return XDP_DROP;
        }
        if (w_val->allowed == 1) {
            submit_event(ctx, 1, iph->saddr, iph->daddr, ct_key.sport, ct_key.dport, iph->protocol, iph->tot_len);
            new_ingress_conn(&ct_key, iph->protocol, now);
            return XDP_PASS;
        }
    }
    if (sd_val) {
        __u64 expire_time = sd_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&sdwhitelist, &sdkey);
            return XDP_DROP;
        }
        if (sd_val->allowed == 1) {
            submit_event(ctx, 1, iph->saddr, iph->daddr, ct_key.sport, ct_key.dport, iph->protocol, iph->tot_len);
            new_ingress_conn(&ct_key, iph->protocol, now);
            return XDP_PASS;
        }       
    }

    if (sp_val) {
        __u64 expire_time = sp_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&src_port, &spkey);
            return XDP_DROP;
        }
        if (sp_val->allowed == 1) {
            submit_event(ctx, 1, iph->saddr, iph->daddr, ct_key.sport, ct_key.dport, iph->protocol, iph->tot_len);
            new_ingress_conn(&ct_key, iph->protocol, now);
            return XDP_PASS;
        }
    }

    if (pl_val) {
        __u64 expire_time = pl_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&port_list, &pl_key);
            return XDP_DROP;
        }
        if (pl_val->allowed == 1) {
            submit_event(ctx, 1, iph->saddr, iph->daddr, ct_key.sport, ct_key.dport, iph->protocol, iph->tot_len);
            new_ingress_conn(&ct_key, iph->protocol, now);
            return XDP_PASS;
        }
    }
    if (pp_val) {
        __u64 expire_time = pp_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&protocol_port, &pp_key);
            return XDP_DROP;
        }
        if (pp_val->allowed == 1) {
            submit_event(ctx, 1, iph->saddr, iph->daddr, ct_key.sport, ct_key.dport, iph->protocol, iph->tot_len);
            new_ingress_conn(&ct_key, iph->protocol, now);
            return XDP_PASS;
        }
    }
    submit_event(ctx, 0, iph->saddr, iph->daddr, ct_key.sport, ct_key.dport, iph->protocol, iph->tot_len);
    return XDP_DROP;
}

char _license[] SEC("license") = "Dual BSD/GPL";