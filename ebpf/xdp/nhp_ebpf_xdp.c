#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>

#define ETH_P_ARP 0x0806
#define ETH_P_IP    0x0800
#define IPPROTO_ICMP 1
#define IPPROTO_TCP 6
#define ICMP_ECHO 8
#define ICMP_ECHOREPLY 0
#define ETH_P_IPV6   0x86DD
#define IPPROTO_UDP 17

enum {
    CT_NEW,
    CT_ESTABLISHED,
};

enum {
    CT_FLAG_NONE = 0,
    CT_FLAG_SYN = 1 << 0,
    CT_FLAG_FIN = 1 << 1,
    CT_FLAG_RST = 1 << 2,
    CT_FLAG_ACK = 1 << 3,
};

enum {
    CT_DIR_INGRESS = 0,
    CT_DIR_EGRESS = 1,
};

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
    __uint(max_entries, 1000000);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} whitelist SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct src_port_list_key);
    __type(value, struct src_port_list_value);
    __uint(max_entries, 1000000);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} src_port_list SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, struct icmpwhitelist_key);
    __type(value,  struct icmpwhitelist_value);
    __uint(max_entries, 1000000);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} icmpwhitelist SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct sdwhitelist_key);
    __type(value, struct sdwhitelist_value);
    __uint(max_entries, 1000000);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} sdwhitelist SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct port_list_key);
    __type(value,struct port_list_value);
    __uint(max_entries, 10000);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} port_list SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct protocol_port_key);
    __type(value,struct protocol_port_value);
    __uint(max_entries, 10000);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} protocol_port SEC(".maps");

struct ipv4_ct_tuple {
    __be32 daddr;
    __be32 saddr;
    __be16 dport;
    __be16 sport;
    __u8 nexthdr;
    __u8 flags;
} __packed;

struct conn_value {
    __u64 timestamp;
    __u64 last_timestamp;
    __u64 ttl_ns;   
    __u8 state;
    __u8 flags;
    __u32 rx_packets;
    __u32 tx_packets;
};

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 1000000);
    __type(key, struct ipv4_ct_tuple);
    __type(value, struct conn_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} conn_track SEC(".maps");

static __always_inline void reverseTuple(struct ipv4_ct_tuple *key) {
    __u32 tmp_ip = key->daddr;
    __u16 tmp_port = key->dport;
    key->flags = !key->flags;
    key->daddr = key->saddr;
    key->saddr = tmp_ip;
    key->dport = key->sport;
    key->sport = tmp_port;
}

static __always_inline void print_ip(__u32 ip) {
    bpf_printk("%d.%d.%d.%d", 
        (ip >> 24) & 0xFF,
        (ip >> 16) & 0xFF,
        (ip >> 8) & 0xFF,
        ip & 0xFF);
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
    } else if (iph->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)(iph + 1);
        if ((void *)(udp + 1) > data_end)
            return XDP_DROP;
        ct_key.nexthdr = IPPROTO_UDP;
        ct_key.sport = udp->source;
        ct_key.dport = udp->dest;
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
        (ct_key.dport == bpf_htons(67) || ct_key.dport == bpf_htons(68))) {
        return XDP_PASS;
    }

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
                //Lookup in the icmp whitelist
                struct icmpwhitelist_value *iw_val = bpf_map_lookup_elem(&icmpwhitelist, &icmpkey);
                if (!iw_val) {
                    bpf_printk("icmpnotWhitelisted src IP: 0x%08x, dst IP: 0x%08x", icmpkey.src_ip, icmpkey.dst_ip);
                    print_ip(bpf_ntohl(icmpkey.src_ip));
                    print_ip(bpf_ntohl(icmpkey.dst_ip));
                    return XDP_DROP;
                }   
                __u64 now = bpf_ktime_get_ns();
                if (iw_val->expire_time < now) {
                    bpf_printk("Time check: now=%llu expire=%llu delta=%lld", 
                        bpf_ktime_get_ns(), 
                        iw_val->expire_time,
                        iw_val->expire_time - bpf_ktime_get_ns());

                    bpf_map_delete_elem(&icmpwhitelist, &icmpkey);
                    return XDP_DROP;
                }
                if (iw_val->allowed == 1) {
                    // bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_ANY);
                    bpf_printk("icmpWhitelisted src IP: 0x%08x, dst IP: 0x%08x", icmpkey.src_ip, icmpkey.dst_ip);
                    print_ip(bpf_ntohl(icmpkey.src_ip));
                    print_ip(bpf_ntohl(icmpkey.dst_ip));
                    // bpf_printk("Whitelisted IP: 0x%08x", key.src_ip);
                    return XDP_PASS;
                }
            }  else {
                return XDP_PASS;
            }
        }
    }

    ct_key.saddr = iph->saddr;
    ct_key.daddr = iph->daddr;
    ct_key.flags = CT_DIR_EGRESS;
    struct conn_value *existing_val;

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

        bpf_printk("direct--- src : 0x%08x, dst: 0x%08x, port: %d, protocol: %d", 
            iph->saddr, iph->daddr, bpf_htons(ct_key.dport), (int)iph->protocol);
        print_ip(bpf_ntohl(iph->saddr));
        print_ip(bpf_ntohl(iph->daddr));
        return XDP_PASS;
    }

    reverseTuple(&ct_key);
    existing_val = bpf_map_lookup_elem(&conn_track, &ct_key);
    if (existing_val) {
        if (check_conn_expiry(existing_val)) {
            bpf_map_delete_elem(&conn_track, &ct_key);
            reverseTuple(&ct_key);
            return XDP_DROP;
        }
        struct conn_value new_val = *existing_val;
        new_val.rx_packets++;
        new_val.last_timestamp = bpf_ktime_get_ns();
        bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_EXIST);
        reverseTuple(&ct_key);

        bpf_printk("opposite----src : 0x%08x, dst: 0x%08x, port: %d, protocol: %d", 
            iph->saddr, iph->daddr, bpf_htons(ct_key.dport), (int)iph->protocol);
        print_ip(bpf_ntohl(iph->saddr));
        print_ip(bpf_ntohl(iph->daddr));
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
        .min_port = 0,
        .max_port = 65535
    };
    __u16 dst_port = bpf_ntohs(ct_key.dport);

    struct protocol_port_key pp_key = {
        .dst_port = ct_key.dport,
        .protocol = iph->protocol
    };

    //Lookup in the whitelist
    struct whitelist_value *w_val = bpf_map_lookup_elem(&whitelist, &key);
    struct sdwhitelist_value *sd_val = bpf_map_lookup_elem(&sdwhitelist, &sdkey);
    struct src_port_list_value *sp_val = bpf_map_lookup_elem(&src_port_list, &spkey);
    struct port_list_value *pl_val= bpf_map_lookup_elem(&port_list, &pl_key);
    struct protocol_port_value *pp_val= bpf_map_lookup_elem(&protocol_port, &pp_key);


    __u64 now = bpf_ktime_get_ns();
    if (w_val && (w_val->expire_time < now)) {
        bpf_printk("Time check: now=%llu expire=%llu delta=%lld", 
            bpf_ktime_get_ns(), 
            w_val->expire_time,
            w_val->expire_time - bpf_ktime_get_ns());

        bpf_map_delete_elem(&whitelist, &key);
        return XDP_DROP;
    }

    if (sd_val && (sd_val->expire_time < now)) {
        bpf_printk("Time check: now=%llu expire=%llu delta=%lld", 
            bpf_ktime_get_ns(), 
            sd_val->expire_time,
            sd_val->expire_time - bpf_ktime_get_ns());

        bpf_map_delete_elem(&sdwhitelist, &sdkey);
        return XDP_DROP;
    }

    if (sp_val && (sp_val->expire_time < now)) {
        bpf_printk("Time check: now=%llu expire=%llu delta=%lld", 
            bpf_ktime_get_ns(), 
            sp_val->expire_time,
            sp_val->expire_time - bpf_ktime_get_ns());

        bpf_map_delete_elem(&src_port_list, &spkey);
        return XDP_DROP;
    }

    if (pl_val && (pl_val->expire_time < now)) {
        bpf_printk("Time check: now=%llu expire=%llu delta=%lld", 
            bpf_ktime_get_ns(), 
            pl_val->expire_time,
            pl_val->expire_time - bpf_ktime_get_ns());

        bpf_map_delete_elem(&port_list, &pl_key);
        return XDP_DROP;
    }

    if (pp_val && (pp_val->expire_time < now)) {
        bpf_printk("Time check: now=%llu expire=%llu delta=%lld", 
            bpf_ktime_get_ns(), 
            pp_val->expire_time,
            pp_val->expire_time - bpf_ktime_get_ns());

        bpf_map_delete_elem(&protocol_port, &pp_key);
        return XDP_DROP;
    }

    if (w_val && w_val->allowed == 1) {
        struct conn_value new_val = {
            .timestamp = bpf_ktime_get_ns(),
            .last_timestamp = bpf_ktime_get_ns(),
            .ttl_ns = w_val->expire_time - now,
            .state = CT_ESTABLISHED,
            .flags = CT_FLAG_NONE,
            .rx_packets = 1,
            .tx_packets = 0,
        };
        bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_ANY);
        bpf_printk("Whitelisted src : 0x%08x, dst: 0x%08x, port: %d, protocol: %d", 
            key.src_ip, key.dst_ip, bpf_htons(key.dst_port), (int)key.protocol);
        print_ip(bpf_ntohl(iph->saddr));
        print_ip(bpf_ntohl(iph->daddr));
        return XDP_PASS;
    }
    
    if (sd_val && sd_val->allowed == 1) {
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
        bpf_printk("Whitelisted src : 0x%08x, dst: 0x%08x, port: %d, protocol: %d", 
            iph->saddr, iph->daddr, bpf_htons(ct_key.dport), (int)iph->protocol);
        print_ip(bpf_ntohl(iph->saddr));
        print_ip(bpf_ntohl(iph->daddr));
        return XDP_PASS;
    }
       
    if (sp_val && sp_val->allowed == 1) {
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
        bpf_printk("Whitelisted src : 0x%08x, dst: 0x%08x, port: %d, protocol: %d", 
            iph->saddr, iph->daddr, bpf_htons(ct_key.dport), (int)iph->protocol);
        print_ip(bpf_ntohl(iph->saddr));
        print_ip(bpf_ntohl(iph->daddr));
        return XDP_PASS;
    }

    if ((pl_val && pl_val->allowed == 1) && (dst_port >= pl_key.min_port && dst_port <= pl_key.max_port)) {
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
        bpf_printk("Whitelisted src : 0x%08x, dst: 0x%08x, port: %d, protocol: %d", 
            iph->saddr, iph->daddr, bpf_htons(ct_key.dport), (int)iph->protocol);
        print_ip(bpf_ntohl(iph->saddr));
        print_ip(bpf_ntohl(iph->daddr));
        return XDP_PASS;
    }

    if (pp_val && pp_val->allowed == 1){
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
        bpf_printk("Whitelisted src : 0x%08x, dst: 0x%08x, port: %d, protocol: %d", 
            iph->saddr, iph->daddr, bpf_htons(ct_key.dport), (int)iph->protocol);
        print_ip(bpf_ntohl(iph->saddr));
        print_ip(bpf_ntohl(iph->daddr));
        return XDP_PASS;
    }

    bpf_printk("Deny: src : 0x%08x, dst: 0x%08x, port: %d, protocol: %d", 
        iph->saddr, iph->daddr, bpf_htons(ct_key.dport), (int)iph->protocol);
    print_ip(bpf_ntohl(iph->saddr));
    print_ip(bpf_ntohl(iph->daddr));
    return XDP_DROP;
}

char _license[] SEC("license") = "GPL";