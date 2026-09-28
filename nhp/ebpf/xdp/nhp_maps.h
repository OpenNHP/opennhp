/*
 * Shared eBPF map definitions for the NHP access controller.
 *
 * Both nhp_ebpf_xdp.c (XDP ingress) and tc_egress.c (TC egress) pin
 * `conn_track` with LIBBPF_PIN_BY_NAME, i.e. they share one kernel map. The
 * loader (cilium/ebpf, see endpoints/ac/ebpf/ebpfegine.go) validates map type,
 * key size, value size and max_entries when the second object reuses the pin,
 * so any drift between the two definitions makes nhp-acd fail to start.
 * Keeping the definitions here rather than duplicating them in both .c files
 * makes that drift impossible.
 *
 * Include after "vmlinux.h" and <bpf/bpf_helpers.h>.
 */

#ifndef __NHP_MAPS_H__
#define __NHP_MAPS_H__

#ifndef MAX_ENTRIES
#define MAX_ENTRIES 1000000
#endif

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

/*
 * Connection-tracking key.
 *
 * NOTE: this struct is deliberately NOT packed. Its natural alignment makes it
 * 16 bytes with 2 bytes of tail padding after `flags`. The layout must stay
 * exactly as-is (field order and types) because it is the key of a pinned map
 * shared by two programs.
 *
 * Because of that tail padding, every ipv4_ct_tuple used as a map key MUST be
 * zeroed in full before its fields are assigned (`= {}` or
 * __builtin_memset(&k, 0, sizeof(k))). Leaving stack garbage in the padding
 * makes lookups and updates silently miss their counterpart in the other
 * program.
 */
struct ipv4_ct_tuple {
    __be32 daddr;
    __be32 saddr;
    __be16 dport;
    __be16 sport;
    __u8 nexthdr;
    __u8 flags;
};

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

#endif /* __NHP_MAPS_H__ */
