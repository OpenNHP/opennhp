#ifndef NHP_MAPS_H
#define NHP_MAPS_H

/* Shared definitions for nhp_ebpf_xdp.c (XDP) and tc_egress.c (TC egress).
 *
 * Both programs share the conn_track LRU map by LIBBPF_PIN_BY_NAME, so the
 * key/value size, max_entries and map type MUST stay identical. Keeping
 * these definitions in a single header eliminates drift between the two
 * translation units, which would otherwise cause cilium/ebpf LoadAndAssign
 * to fail at startup.
 *
 * This header intentionally does NOT include vmlinux.h or any libbpf
 * header; the includer is expected to provide those so the same types
 * (__be32, __be16, __u8, __u32, __u64, __uint, __type, __packed,
 *  BPF_MAP_TYPE_LRU_HASH, LIBBPF_PIN_BY_NAME, MAX_ENTRIES, ...) are
 * already in scope.
 */

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
    __uint(max_entries, MAX_ENTRIES);
    __type(key, struct ipv4_ct_tuple);
    __type(value, struct conn_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} conn_track SEC(".maps");

#endif /* NHP_MAPS_H */