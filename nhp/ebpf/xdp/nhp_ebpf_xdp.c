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

struct whitelist_key {
    __be32 src_ip;
    __be32 dst_ip;
    __be16 dst_port;
    __u8 protocol;
} __attribute__((packed));

struct src_port_list_key {
    __be32 src_ip;
    __be16 dst_port;
} __attribute__((packed));

struct port_list_key {
    __be32 src_ip;
    __be16 min_port;
    __be16 max_port;
} __attribute__((packed));

struct protocol_port_key {
    __be16 dst_port;
    __u8 protocol;
} __attribute__((packed));

struct icmpwhitelist_key {
    __be32 src_ip;
    __be32 dst_ip;
} __attribute__((packed));

struct sdwhitelist_key {
    __be32 src_ip;
    __be32 dst_ip;
} __attribute__((packed));

struct whitelist_value {
    __u8 allowed;
    __u64 expire_time;
};

struct icmpwhitelist_value {
    __u8 allowed;
    __u64 expire_time;
};

struct sdwhitelist_value {
    __u8 allowed;
    __u64 expire_time;
};

struct src_port_list_value {
    __u8 allowed;
    __u64 expire_time;
};

struct port_list_value {
    __u8 allowed;
    __u64 expire_time;
};

struct protocol_port_value {
    __u8 allowed;
    __u64 expire_time;
} __attribute__((packed));

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct whitelist_key);
    __type(value, struct whitelist_value);
    __uint(max_entries, MAX_ENTRIES);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} spp SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct src_port_list_key);
    __type(value, struct src_port_list_value);
    __uint(max_entries, MAX_ENTRIES);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} src_port SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct icmpwhitelist_key);
    __type(value,  struct icmpwhitelist_value);
    __uint(max_entries, MAX_ENTRIES);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} icmpwhitelist SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct sdwhitelist_key);
    __type(value, struct sdwhitelist_value);
    __uint(max_entries, MAX_ENTRIES);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} sdwhitelist SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct port_list_key);
    __type(value,struct port_list_value);
    __uint(max_entries, MAX_ENTRIES);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} port_list SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct protocol_port_key);
    __type(value,struct protocol_port_value);
    __uint(max_entries, MAX_ENTRIES);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} protocol_port SEC(".maps");

/* struct ipv4_ct_tuple, struct conn_value and the conn_track map live in
 * nhp_maps.h -- they are shared with tc_egress.c through the pinned map. */

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

#ifndef __constant_htons
#define __constant_htons(x) ((__u16)((((x) & 0xFF00) >> 8) | (((x) & 0x00FF) << 8)))
#endif

static __always_inline bool check_conn_expiry(struct conn_value *val) {
    __u64 now = bpf_ktime_get_ns();
    return (now > val->timestamp + val->ttl_ns);
}

SEC("xdp")
static __always_inline int xdp_white_prog(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    struct tcphdr *tcp;
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

    if (iph->protocol == IPPROTO_TCP) {
        tcp = (void *)(iph + 1);
        if ((void *)(tcp + 1) > data_end) {
            return XDP_DROP;
        }
        ct_key.nexthdr = IPPROTO_TCP;
        ct_key.sport = tcp->source;
        ct_key.dport = tcp->dest;
        // ct_key.dport = bpf_htons(tcp->dest);
    } else if (iph->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)(iph + 1);
        if ((void *)(udp + 1) > data_end)
            return XDP_DROP;
        ct_key.nexthdr = IPPROTO_UDP;
        ct_key.sport = udp->source;
        ct_key.dport = udp->dest;
        // ct_key.dport = bpf_htons(udp->dest);
    } 

    
    if (iph->protocol == IPPROTO_TCP) {
        void *tcp_start = (void *)iph + (iph->ihl * 4);
        if ((void *)(tcp_start + sizeof(struct tcphdr)) > data_end)
            return XDP_DROP;

        struct tcphdr *tcp = tcp_start;
        if (__constant_htons(tcp->dest) == 22) {
            return XDP_PASS;
        }
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
     * Forward (CT_DIR_INGRESS) hit: a flow the peer opened and that we
     * admitted via one of the whitelists. `timestamp` -- the base of the
     * expiry check -- is deliberately NOT refreshed here: the entry's ttl_ns
     * was derived from the knock's expiry, so the flow must die exactly when
     * the knock does. Refreshing it would keep the door open for as long as
     * traffic keeps arriving, which is the bug this asymmetry exists to
     * prevent. Only last_timestamp (bookkeeping) is updated.
     */
    existing_val = bpf_map_lookup_elem(&conn_track, &ct_key);
    if (existing_val) {
        if (check_conn_expiry(existing_val)) {
            bpf_map_delete_elem(&conn_track, &ct_key);
            return XDP_DROP;
        }
        struct conn_value new_val = *existing_val;
        new_val.tx_packets++;
        new_val.last_timestamp = bpf_ktime_get_ns();
        bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_EXIST);
        return XDP_PASS;
    }
    reverseTuple(&ct_key);
    /*
     * Reverse (CT_DIR_EGRESS) hit: the reply to a connection the AC itself
     * opened, whose entry was created by tc_egress.c. Here `timestamp` IS
     * refreshed, turning EGRESS_CONN_TTL_NS into a sliding idle window rather
     * than a hard cap -- tc_egress.c only writes the entry once (on the SYN),
     * so without this a download longer than the TTL would be cut off.
     *
     * This cannot extend a knock: reaching this branch means the forward
     * lookup above missed, and a knock-authorised flow never has an EGRESS
     * entry at all -- the two gates in tc_egress_prog() keep it from being
     * created.
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
        new_val.timestamp = bpf_ktime_get_ns();
        new_val.last_timestamp = new_val.timestamp;
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
    __u16 dst_port = bpf_ntohs(ct_key.dport);

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
            struct conn_value new_val = {
                .timestamp = bpf_ktime_get_ns(),
                .last_timestamp = bpf_ktime_get_ns(),
                .ttl_ns = expire_time - now,
                .state = CT_ESTABLISHED,
                .flags = CT_FLAG_NONE,
                .rx_packets = 1,
                .tx_packets = 0,
            };
            bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_ANY);
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
            struct conn_value new_val = {
                .timestamp = bpf_ktime_get_ns(),
                .last_timestamp = bpf_ktime_get_ns(),
                .ttl_ns = sd_val->expire_time - now,
                .state = CT_ESTABLISHED,
                .flags = CT_FLAG_NONE,
                .rx_packets = 1,
                .tx_packets = 0,
            };
            bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_ANY);
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
            struct conn_value new_val = {
                .timestamp = bpf_ktime_get_ns(),
                .last_timestamp = bpf_ktime_get_ns(), 
                .ttl_ns = sp_val->expire_time - now,
                .state = CT_ESTABLISHED,
                .flags = CT_FLAG_NONE,
                .rx_packets = 1,
                .tx_packets = 0,
            };
            bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_ANY);
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
            struct conn_value new_val = {
                .timestamp = bpf_ktime_get_ns(),
                .last_timestamp = bpf_ktime_get_ns(), 
                .ttl_ns = pl_val->expire_time - now,
                .state = CT_ESTABLISHED,
                .flags = CT_FLAG_NONE,
                .rx_packets = 1,
                .tx_packets = 0,
            };
            bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_ANY);
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
            struct conn_value new_val = {
                .timestamp = bpf_ktime_get_ns(),
                .last_timestamp = bpf_ktime_get_ns(), 
                .ttl_ns = pp_val->expire_time - now,
                .state = CT_ESTABLISHED,
                .flags = CT_FLAG_NONE,
                .rx_packets = 1,
                .tx_packets = 0,
            };
            bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_ANY);
            return XDP_PASS;
        }
    }
    submit_event(ctx, 0, iph->saddr, iph->daddr, ct_key.sport, ct_key.dport, iph->protocol, iph->tot_len);
    return XDP_DROP;
}

char _license[] SEC("license") = "Dual BSD/GPL";