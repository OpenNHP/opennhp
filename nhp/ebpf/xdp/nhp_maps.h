/*
 * Shared eBPF map definitions for the NHP access controller.
 *
 * Both nhp_ebpf_xdp.c (XDP ingress) and tc_egress.c (TC egress) pin their maps
 * with LIBBPF_PIN_BY_NAME, i.e. they share one kernel map per name. The loader
 * (cilium/ebpf, see endpoints/ac/ebpf/ebpfegine.go) validates map type, key
 * size, value size and max_entries when the second object reuses the pin, so
 * any drift between the two definitions makes nhp-acd fail to start. Keeping
 * the definitions here rather than duplicating them in both .c files makes that
 * drift impossible.
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
 * Lifetimes for conn_track entries. None of them is the knock TTL -- that lives
 * in the whitelist maps below and governs which *new* flows may be created.
 *
 * These two are the windows an admitted inbound flow gets, and both are idle
 * windows: the entry's `timestamp` is pushed forward by every packet of that
 * flow, so it lives for as long as it keeps moving and dies once it stops.
 *
 * The values follow netfilter's conntrack defaults closely enough that
 * FilterMode 1 (eBPF) and FilterMode 0 (iptables + ipset, whose baseline
 * accepts `-m state --state ESTABLISHED`) behave the same way for an
 * established session. nf_conntrack's tcp_timeout_established is five days; an
 * hour is the deliberate difference, since these entries sit in a pinned LRU
 * map with no netfilter to garbage-collect them.
 *
 * For TCP the idle window is not the only thing that bounds the entry, and it
 * must not be -- the peer picks its own source port, so an idle window alone
 * would let a peer that knocked once pin a 4-tuple open forever by sending one
 * packet an hour. xdp_white_prog() therefore tracks the connection's state the
 * way netfilter does:
 *
 *   - only a pure SYN that a whitelist admits creates an entry, so the entry
 *     exists exactly for the connection that the knock authorised;
 *   - a pure SYN never matches an existing entry -- every new connection goes
 *     through the whitelist lookups, so re-using a source port buys nothing;
 *   - a RST deletes the entry, and a FIN cuts it down to NHP_CT_CLOSE_TTL_NS,
 *     so the entry dies with the connection instead of lingering for an hour.
 *
 * Non-TCP has no handshake to key off, so a UDP entry stays a plain idle
 * window, which is exactly what nf_conntrack gives FilterMode 0.
 */
#define NHP_CT_TCP_IDLE_TTL_NS   (3600ULL * 1000000000ULL)  /* nf: 5 days      */
#define NHP_CT_OTHER_IDLE_TTL_NS (120ULL * 1000000000ULL)   /* nf: udp_stream  */

/*
 * Window a TCP entry keeps after the first FIN, long enough for the rest of the
 * close handshake and any straggling retransmit. nf_conntrack uses 60s for
 * CLOSE_WAIT and 120s for TIME_WAIT.
 */
#define NHP_CT_CLOSE_TTL_NS (60ULL * 1000000000ULL)

/*
 * Reply window for connections the AC itself initiates (dnf, certbot, NTP, its
 * UDP channel to the nhp-server, ...), written by tc_egress.c. The XDP program
 * drops everything that is not whitelisted, so outbound flows need their
 * replies let back in -- the equivalent of iptables'
 * `--ctstate ESTABLISHED,RELATED -j ACCEPT`.
 *
 * It is an idle window driven by the AC's *own* sending: tc_egress.c pushes
 * `timestamp` forward on every outgoing packet of the flow. An inbound packet
 * refreshes it only for TCP, where the gate that creates the entry (pure SYN
 * from a non-listening socket) is exact; for UDP, where the gates are
 * heuristics, xdp_white_prog() deliberately does not refresh, so a peer can
 * never slide the window forward on its own. That is strictly tighter than
 * netfilter, which refreshes on traffic in either direction.
 */
#define NHP_CT_EGRESS_IDLE_TTL_NS (180ULL * 1000000000ULL)

/*
 * Fallback lower bound of the local ephemeral port range (Linux default
 * net.ipv4.ip_local_port_range = "32768 60999"). A socket the AC opens as a
 * *client* binds a port from this range; a socket it *listens* on does not.
 * tc_egress.c uses that to tell its own outbound connections apart from the
 * reply side of a service it is serving -- see gate 3 there.
 *
 * Distros and hardened images do change ip_local_port_range, and a host whose
 * range starts below this would leave the AC's own UDP flows from the lower
 * ports (a resolver, chrony, the nhp-server channel) untracked and therefore
 * unanswerable. So the real bound is read from the host at load time and put in
 * `nhp_config` by endpoints/ac/ebpf/ebpfegine.go; this constant is only what
 * the program falls back to if that slot is unset.
 */
#define NHP_EPHEMERAL_PORT_MIN 32768

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

/*
 * The knock whitelists. User space (nhp/utils/ebpf/ebpf.go) writes them when a
 * knock is authorised, with `expire_time` derived from the resource's OpenTime;
 * xdp_white_prog() reads them to decide whether a *new* flow may be created.
 * Neither eBPF program may write them -- tc_egress.c used to write `spp` for
 * every outgoing packet, which kept the door open long past OpenTime.
 *
 * tc_egress.c reads `spp` and `src_port` (never writes any of them) to
 * recognise a peer that is talking to a knock-protected service, see gate 4
 * there for why only those two.
 */
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
    __uint(max_entries, MAX_ENTRIES);
    __type(key, struct ipv4_ct_tuple);
    __type(value, struct conn_value);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} conn_track SEC(".maps");

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
    __type(value, struct icmpwhitelist_value);
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
    __type(value, struct port_list_value);
    __uint(max_entries, MAX_ENTRIES);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} port_list SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct protocol_port_key);
    __type(value, struct protocol_port_value);
    __uint(max_entries, MAX_ENTRIES);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} protocol_port SEC(".maps");

/*
 * Runtime knobs the programs cannot read for themselves, one __u32 per slot.
 * Written once by user space when the objects are loaded
 * (endpoints/ac/ebpf/ebpfegine.go); the eBPF side only reads it, and treats an
 * unset slot as "use the compiled-in default", so a stale or empty map degrades
 * to the previous behaviour rather than to an open door.
 */
enum {
    NHP_CFG_EPHEMERAL_PORT_MIN = 0,
    NHP_CFG_SLOTS,
};

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __type(key, __u32);
    __type(value, __u32);
    __uint(max_entries, NHP_CFG_SLOTS);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} nhp_config SEC(".maps");

/* Host's net.ipv4.ip_local_port_range low bound, or the default if user space
 * has not filled it in. See NHP_EPHEMERAL_PORT_MIN. */
static __always_inline __u32 nhp_ephemeral_port_min(void)
{
    __u32 idx = NHP_CFG_EPHEMERAL_PORT_MIN;
    __u32 *val = bpf_map_lookup_elem(&nhp_config, &idx);

    if (!val || *val == 0 || *val > 65535)
        return NHP_EPHEMERAL_PORT_MIN;
    return *val;
}

/*
 * Idle window an admitted flow gets, by L4 protocol. TCP sessions are
 * long-lived by nature (SSH, WebSockets, long-polling, a slow download);
 * everything else gets the shorter datagram window.
 */
static __always_inline __u64 nhp_ct_idle_ttl_ns(__u8 protocol)
{
    return protocol == 6 /* IPPROTO_TCP */ ? NHP_CT_TCP_IDLE_TTL_NS
                                           : NHP_CT_OTHER_IDLE_TTL_NS;
}

#endif /* __NHP_MAPS_H__ */
