/* Ingress filter for the nhp-server host.
 *
 * Unlike the AC's nhp_ebpf_xdp.c, which opens ports per knock, the server's
 * *listening* exposure is fixed and tiny: the NHP UDP knock port, and SSH from
 * the relay jump host. Everything else that would start a new flow is dropped
 * at the driver, before the kernel's network stack sees it. There are therefore
 * no whitelist maps here beyond the relay list, and no AC-style conn_track /
 * TC egress pair, which is why this program can stay short and share nothing
 * with the AC's schema.
 *
 * What it does need is the other half of any ingress filter: the replies to
 * connections the *host itself* opened. nhp-serverd is a client as well as a
 * server -- it resolves names, its auth plugin submits OTP mail to SES, the
 * host talks to package mirrors -- and without a way back in, every one
 * of those hangs until it times out (the first symptom was the resolver: "read
 * udp 10.0.1.78:58655->10.0.0.2:53: i/o timeout", i.e. the DNS *answer* dropped
 * here, and with it the OTP mail behind it). The AC pays for that with a TC
 * egress program and a conn_track map because it must reason about flows a
 * knock authorised; the server has no such ambiguity, so it asks the kernel's
 * own socket table instead -- see has_local_flow(). That keeps the "no
 * unsolicited packet gets in" property exactly: a listening socket is never an
 * answer, only an established or connected one is.
 *
 * The socket table cannot answer for every client, though, and the two it
 * cannot answer for are named explicitly in the UDP branch below: DHCP, whose
 * reply lands on a raw or merely-bound socket and whose loss costs the host
 * its IP address one lease later, and NTP, whose reply comes back to chronyd's
 * own udp/123. Both are admitted by port pair, and that is the entire list --
 * "some daemon's replies are being dropped" is not on its own a reason to
 * extend it, because the alternative (admit anything addressed to a bound
 * port) is the property above thrown away.
 *
 * It deliberately does NOT parse the NHP protocol. Identity is still decided in
 * user space by the Noise handshake (nhp/core/packet.go::RecvPrecheck); the
 * only NHP-shaped test here is a minimum datagram length, which costs nothing
 * and turns the knock port into a non-answer for ordinary scanners.
 *
 * Two shapes that are not a plain unfragmented IPv4 datagram get an explicit
 * answer here rather than being waved through, because a filter that says it
 * drops everything but X must not have a hole the size of an address family:
 *
 *   IPv4 fragments -- only the first fragment carries the L4 header, so the
 *   port-based tree above can only judge that one. Later fragments are PASSed
 *   (see the IP_OFFSET branch): they cannot reach a socket on their own, since
 *   reassembly needs the first fragment, and that one went through the full
 *   policy. Dropping them instead would break every datagram larger than the
 *   path MTU, knocks included -- and reading "ports" out of a fragment's
 *   payload, which is what this program used to do, is worse than either. An
 *   IPv6 fragment header gets the same answer, in handle_ipv6().
 *
 *   IPv6 -- filtered with the same tree as IPv4, minus the whitelist: the relay
 *   list is IPv4 prefixes (see nhp_relay_ips), so nothing gets to SSH over v6.
 *   It used to be an unconditional XDP_PASS on the grounds that the demo hosts
 *   have no v6 address, which left sshd's [::]:22 -- and every other listener --
 *   reachable unfiltered the day a v6 CIDR is added to the VPC, with the
 *   deploy's "filter attached" check still green. ICMPv6 neighbour discovery
 *   and MLD are passed so a host on a v6 subnet still works, and replies to the
 *   host's own v6 flows are admitted by the same socket-table lookup as v4.
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
#define IPPROTO_ICMP 1
#define IPPROTO_TCP  6
#define IPPROTO_UDP  17

/* IPv6 next-header values: the three extension headers that may sit in front of
 * the L4 header on ordinary traffic (an MLD report carries a hop-by-hop router
 * alert), and ICMPv6. */
#define IPPROTO_HOPOPTS  0
#define IPPROTO_ROUTING  43
#define IPPROTO_DSTOPTS  60
#define IPPROTO_ICMPV6   58
#define IPPROTO_FRAGMENT 44

/* ICMP destination-unreachable / fragmentation-needed, i.e. the one ICMP
 * message a host that opens outbound TCP connections cannot do without. */
#define ICMP_DEST_UNREACH  3
#define ICMP_FRAG_NEEDED   4

/* iph->frag_off, network order. MF says "more fragments follow"; OFFSET is
 * non-zero on every fragment but the first, which is the one that has no L4
 * header behind the IP header. IP6_FRAG_OFFSET is the same field in an IPv6
 * fragment header (13 bits, the low 3 being reserved + M). */
#define IP_MF            0x2000
#define IP_OFFSET        0x1FFF
#define IP6_FRAG_OFFSET  0xFFF8

/* ICMPv6 types that keep a host working on an IPv6 subnet: "packet too big"
 * (the frag-needed of v6, see the ICMP branch at the bottom of the program) and
 * the contiguous MLD/ND block 130..137 -- multicast listener query/report/done,
 * router and neighbour solicitation/advertisement, redirect. Neighbour
 * discovery *is* v6's ARP, so dropping it would take the host off its own
 * subnet; everything else, echo request included, stays dropped. */
#define ICMPV6_PKT_TOOBIG  2
#define ICMPV6_MLD_FIRST   130
#define ICMPV6_ND_LAST     137

#define SSH_PORT 22

/* Client protocols whose answers arrive on a socket the kernel cannot tell
 * apart from a listener, so has_local_flow() cannot admit them. See the DHCP
 * and NTP branches below for why each one is here; nothing may be added to
 * this list without the same kind of argument. */
#define DHCP_PORT_SERVER 67
#define DHCP_PORT_CLIENT 68
#define NTP_PORT         123

/* The v6 spelling of the DHCP pair, for the same reason: a host that gets its
 * v6 address from DHCPv6 rather than from SLAAC loses it at lease expiry if the
 * reply (547 -> 546) is dropped. */
#define DHCP6_PORT_SERVER 547
#define DHCP6_PORT_CLIENT 546

/* The knock port and the length floor are .rodata constants, not #defines,
 * because both are user-space configuration: ListenPort in etc/config.toml and
 * NhpMinFrameBytes in etc/xdp.toml. The loader rewrites them from those two
 * files before the program is verified (loadServerEngine in
 * nhp/utils/ebpf/engine_linux.go), so a server that listens anywhere other
 * than 62206 filters for the port it actually listens on instead of dropping
 * every knock it receives. The values below are only the fallback a bare
 * `bpftool prog load` of this object would get.
 *
 * Read as `volatile const` so the verifier still treats them as constants
 * after the rewrite (dead branches get folded away) while the compiler cannot
 * bake the initialiser into the instruction stream.
 *
 * nhp_min_udp_len defaults to the 240-byte NHP_KPL header of the curve25519
 * cipher suite (nhp/core/packet.go); the gmsm suite's 304 bytes clears it too,
 * so one threshold covers both. Anything shorter cannot be a knock no matter
 * what it decrypts to, so it is dropped without touching the payload. The
 * length compared is the UDP header's own length field, i.e. header+payload. */
volatile const __u16 nhp_listen_port = 62206;
volatile const __u16 nhp_min_udp_len = 240;

/* action codes reported to user space; see serverActionName() in
 * nhp/utils/ebpf/engine_linux.go, which formats them, and the decision table
 * in terraform/demo/VERIFY-server-xdp.zh-cn.md. */
#define ACT_DROP_OTHER          0
#define ACT_SSH_RELAY           1
#define ACT_NHP_RELAY           2
#define ACT_NHP_DEFAULT         3
#define ACT_TCP_ESTABLISHED     4
#define ACT_UDP_ESTABLISHED     5
#define ACT_ICMP_FRAG_NEEDED    6
#define ACT_DHCP_CLIENT         7
#define ACT_NTP_CLIENT          8
#define ACT_DROP_TCP_SSH_OTHER  10
#define ACT_DROP_TCP_NHP        11
#define ACT_DROP_TCP_OTHER      12
#define ACT_DROP_UDP_OTHER      13
#define ACT_DROP_UDP_SHORT      14
#define ACT_DROP_NONUDP         15
#define ACT_IPV4_FRAGMENT       16
#define ACT_V6_ICMP_CONTROL     17
#define ACT_V6_ESTABLISHED      18
#define ACT_V6_FRAGMENT         19
#define ACT_DROP_V6_OTHER       20

/* Source prefixes allowed to reach SSH, and allowed to reach the knock port
 * without meeting the length floor. Driven from user space by
 * endpoints/server/ebpf/serverengine.go::UpdateRelayIPs, which mirrors
 * etc/xdp.toml into it on every reload.
 *
 * An LPM trie rather than a hash of host addresses, because the whitelist has
 * to survive the relay being *replaced*. The relay's private address is not
 * pinned by Terraform (aws_instance.relay takes whatever the subnet hands it),
 * so any rebuild gives it a new one -- and a whitelist naming only the old
 * address drops SSH from the new relay, which is the one path CI has for
 * pushing a corrected whitelist. Listing the relay's subnet as a prefix closes
 * that trap: a replacement lands in the same subnet and is still allowed in.
 * A bare address is simply a /32, so the host-address form still works.
 *
 * The address half of the key is __be32, i.e. wire order, which is also the
 * order an LPM trie matches prefixes in, so iph->saddr is looked up with no
 * byte swapping in the hot path. */
struct relay_prefix_key {
    __u32  prefixlen;
    __be32 addr;
};

struct {
    __uint(type, BPF_MAP_TYPE_LPM_TRIE);
    __uint(max_entries, 4096);
    /* LPM tries are only creatable with BPF_F_NO_PREALLOC. */
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __uint(pinning, LIBBPF_PIN_BY_NAME);
    __type(key, struct relay_prefix_key);
    __type(value, __u8);
} nhp_relay_ips SEC(".maps");

/* NHP_FAMILY_* is the `family` field of nhp_event_t, not AF_INET: it says which
 * of the two address pairs below carries this event's addresses. */
#define NHP_FAMILY_V4 4
#define NHP_FAMILY_V6 6

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
    /* Which pair holds the addresses: NHP_FAMILY_V4 means src_ip/dst_ip and
     * the v6 fields are zero, NHP_FAMILY_V6 the reverse. The v6 fields were
     * added when IPv6 stopped being an unconditional pass -- a drop class whose
     * log lines have no source address is a blind spot, and a __be32 cannot
     * hold a v6 address. Parsed by readServerEvents in
     * nhp/utils/ebpf/engine_linux.go; serverEventByteSize there is sizeof this
     * struct, so a field added here has to be added there too. */
    __u8   family;
    __u8   src_ip6[16];
    __u8   dst_ip6[16];
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
 * address, or a knock that passed). Per-CPU so the counter needs no atomics.
 *
 * A per-class budget alone is not a bound on the log, though, and that was the
 * first version's mistake. A remote sender picks the class: over IPv4 it can
 * reach about eleven reportable ones (TCP_SSH_OTHER, TCP_NHP_PORT, TCP_OTHER,
 * UDP_OTHER, UDP_SHORT, NON_TCP_UDP, IPV4_FRAGMENT, NHP_DEFAULT,
 * ICMP_FRAG_NEEDED, a spoofed 67->68 DHCP_CLIENT, NHP_RELAY from the subnet),
 * and spreading the flows across RSS queues fills every CPU's bucket. Eleven
 * classes x 16 x nr_cpus is hundreds of lines a second, i.e. gigabytes a day,
 * on a host whose whole event-log budget is 256 MiB. So a second bucket sits
 * across all of them: NHP_EVENT_GLOBAL_BURST caps the total, whatever mix of
 * classes it is made of, while the per-class bucket keeps that total from
 * being spent entirely on one. The per-CPU multiplier still applies to both --
 * what makes the *size* bound hard is serverLogDailyEventBytes on the writing
 * side (nhp/utils/ebpf/engine_linux.go); this is what keeps the ring and the
 * CPU cost of reporting bounded.
 *
 * NHP_EVENT_ACTIONS sizes the counter array and the per-class part of this
 * one, so it must stay above the highest ACT_* code (a code at or past it is
 * still judged, just neither counted nor rate limited). It carries a little
 * headroom over the codes in use; serverEventActions in
 * nhp/utils/ebpf/engine_linux.go mirrors it, and a test fails if the two drift
 * apart. The global bucket lives in one extra slot past the end. */
#define NHP_EVENT_ACTIONS 24
#define NHP_EVENT_WINDOW_NS 1000000000ULL
#define NHP_EVENT_BURST 16
#define NHP_EVENT_GLOBAL_BURST 8
#define NHP_EVENT_GLOBAL_SLOT NHP_EVENT_ACTIONS
#define NHP_EVENT_RATE_SLOTS (NHP_EVENT_ACTIONS + 1)

struct nhp_rate_state {
    __u64 window_start;
    __u32 count;
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, NHP_EVENT_RATE_SLOTS);
    __type(key, __u32);
    __type(value, struct nhp_rate_state);
} nhp_event_rate SEC(".maps");

/* What the budget above throws away, kept as numbers.
 *
 * A token bucket alone makes the log lie by omission: over budget, a flood and
 * a trickle look the same, and "how much of this is there" -- the first
 * question anyone asks of an ingress filter -- is unanswerable. These counters
 * are incremented for *every* packet, budget or no, so user space can print one
 * summary line per window per class (see reportServerStats in
 * nhp/utils/ebpf/engine_linux.go) and the total is never wrong no matter how
 * much reporting was dropped. `logged` is the subset that reached the ring, so
 * the difference is the suppressed count, stated rather than guessed at.
 *
 * Per-CPU, so ++ needs no atomics; user space sums the slices. Not pinned: only
 * this process reads them, same as nhp_events. */
struct nhp_action_stat {
    __u64 packets;
    __u64 logged;
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, NHP_EVENT_ACTIONS);
    __type(key, __u32);
    __type(value, struct nhp_action_stat);
} nhp_action_stats SEC(".maps");

/* Take one token from the bucket in `slot`, or report that it is empty for the
 * rest of this window. A slot the map cannot hand back is treated as having
 * budget: the rate limit is telemetry policy, never a verdict, so a map failure
 * must not start silently hiding packets. */
static __always_inline bool spend_token(__u32 slot, __u32 burst, __u64 now) {
    struct nhp_rate_state *st = bpf_map_lookup_elem(&nhp_event_rate, &slot);
    if (!st)
        return true;

    if (now - st->window_start >= NHP_EVENT_WINDOW_NS) {
        st->window_start = now;
        st->count = 1;
        return true;
    }
    if (st->count >= burst)
        return false;

    st->count++;
    return true;
}

/* Per class first, then across all classes. An event that clears its class but
 * not the global cap has spent a class token it did not use; that only makes
 * the class slightly stingier in a window the global cap has already closed,
 * and it keeps the common path to a single lookup. */
static __always_inline bool event_budget_available(__u8 action, __u64 now) {
    __u32 slot = action;
    if (slot >= NHP_EVENT_ACTIONS)
        return true;

    if (!spend_token(slot, NHP_EVENT_BURST, now))
        return false;

    return spend_token(NHP_EVENT_GLOBAL_SLOT, NHP_EVENT_GLOBAL_BURST, now);
}

/* Tick this action's total, which happens for every packet whatever the
 * reporting decision below is, and hand the caller the slot so it can tick
 * `logged` too if the event reaches the ring. */
static __always_inline struct nhp_action_stat *count_packet(__u8 action) {
    __u32 slot = action;
    struct nhp_action_stat *stat = NULL;

    if (slot < NHP_EVENT_ACTIONS)
        stat = bpf_map_lookup_elem(&nhp_action_stats, &slot);
    if (stat)
        stat->packets++;
    return stat;
}

/* Shared tail of the two recorders: stamp the event, spend a token, emit. */
static __always_inline void emit_event(void *ctx, struct nhp_event_t *ev,
                                       struct nhp_action_stat *stat) {
    __u64 now = bpf_ktime_get_ns();
    if (!event_budget_available(ev->action, now))
        return;
    ev->timestamp = now;

    if (bpf_perf_event_output(ctx, &nhp_events, BPF_F_CURRENT_CPU, ev, sizeof(*ev)) == 0 && stat)
        stat->logged++;
}

/* Count the packet against its action, and put an event on the perf ring if
 * the class is worth a line and has budget left. record_packet is the IPv4
 * entry point and record_v6_packet below the IPv6 one; both count through
 * count_packet and emit through emit_event, so neither family can end up
 * outside the accounting or outside the budget.
 *
 * `report` is what keeps the log readable. A verdict is worth a line when it
 * says something the previous lines did not: a knock admitted, an SSH attempt
 * from an address that is not the relay, a scan. It is worth nothing when it is
 * the four-thousandth data segment of a connection whose one real decision --
 * "the host opened this flow" -- was already logged, or was never a decision at
 * all because the kernel's socket table answered it. Those bulk classes
 * (TCP_ESTABLISHED, UDP_ESTABLISHED, V6_ESTABLISHED, ICMPV6_CONTROL and every
 * SSH_RELAY packet after the SYN) are passed report=false: they scale with
 * *bytes transferred* or with a subnet's background chatter, not with events,
 * so one `dnf update` or one CI scp of the release tarball would otherwise
 * write more log than a month of knocks. They are still counted, and the
 * per-window summary reports them.
 *
 * The verdict never depends on this argument -- only the telemetry does. */
static __always_inline void record_packet(void *ctx, __u8 action, struct iphdr *iph,
                                          __be16 src_port, __be16 dst_port,
                                          __u8 relay_hit, bool report) {
    struct nhp_action_stat *stat = count_packet(action);

    if (!report)
        return;

    struct nhp_event_t ev = {};
    ev.action = action;
    ev.family = NHP_FAMILY_V4;
    ev.src_ip = iph->saddr;
    ev.dst_ip = iph->daddr;
    ev.src_port = src_port;
    ev.dst_port = dst_port;
    ev.protocol = iph->protocol;
    ev.pkt_len = iph->tot_len;
    ev.relay_hit = relay_hit;

    emit_event(ctx, &ev, stat);
}

/* The IPv6 recorder. `proto` is the next-header value the extension-header walk
 * arrived at, i.e. the actual L4 protocol, rather than ip6h->nexthdr, which is
 * the first extension header whenever there is one. relay_hit is not a
 * parameter: the whitelist is IPv4-only, so no v6 packet can ever hit it. */
static __always_inline void record_v6_packet(void *ctx, __u8 action, struct ipv6hdr *ip6h,
                                             __u8 proto, __be16 src_port, __be16 dst_port,
                                             bool report) {
    struct nhp_action_stat *stat = count_packet(action);

    if (!report)
        return;

    struct nhp_event_t ev = {};
    ev.action = action;
    ev.family = NHP_FAMILY_V6;
    __builtin_memcpy(ev.src_ip6, ip6h->saddr.in6_u.u6_addr8, sizeof(ev.src_ip6));
    __builtin_memcpy(ev.dst_ip6, ip6h->daddr.in6_u.u6_addr8, sizeof(ev.dst_ip6));
    ev.src_port = src_port;
    ev.dst_port = dst_port;
    ev.protocol = proto;
    ev.pkt_len = ip6h->payload_len;

    emit_event(ctx, &ev, stat);
}

static __always_inline bool is_relay_src(__be32 saddr) {
    /* A full-length key: the trie matches it against the longest stored
     * prefix, so this hits a /32 host entry and a /24 subnet entry alike. */
    struct relay_prefix_key key = {
        .prefixlen = 32,
        .addr      = saddr,
    };
    return bpf_map_lookup_elem(&nhp_relay_ips, &key) != NULL;
}

/* Does this packet belong to a flow the host itself already has a socket for?
 *
 * This is the equivalent of netfilter's `--ctstate ESTABLISHED,RELATED
 * -j ACCEPT`, and the AC solves it with tc_egress.c writing a conn_track entry
 * for every outbound flow. The server does not need any of that machinery: the
 * kernel is already keeping the only table that matters -- its socket table --
 * and bpf_sk_lookup_{tcp,udp}() can be asked about it from XDP. No map, no TTL,
 * no heuristic about which source ports are "client" ports, and nothing to get
 * out of step with reality; the entry appears when the socket is created and is
 * gone the moment it closes.
 *
 * The tuple is filled straight from the packet (source = the remote end), which
 * is how the helper is defined for an ingress hook: it returns the socket that
 * *would receive* this packet. So the answer is exactly "is there a local
 * socket expecting this", and the caller turns that into a verdict:
 *
 *   TCP  -- any state but BPF_TCP_LISTEN. A listener matches every SYN sent to
 *           a port the host serves, so admitting it would silently undo this
 *           whole filter the day something binds tcp/8080 on the box. Every
 *           other state (SYN_SENT waiting for a SYN-ACK, ESTABLISHED, the
 *           closing states) belongs to a connection this host opened or already
 *           accepted through the branches above.
 *   UDP  -- connected sockets only (`dst_port != 0`). A resolver, an NTP client
 *           or net/smtp's dialer connect(2) their socket, so the kernel's
 *           lookup only matches them for datagrams from the peer they are
 *           talking to; an unconnected socket bound to some port is a listener
 *           by another name and gets nothing from here.
 *
 * Cost: one hash lookup for packets that are about to be dropped anyway, i.e.
 * scan traffic. That is the same lookup the network stack would do had XDP
 * passed the packet, so a flood is no more expensive than it was before this
 * program existed -- and still cheaper, because the verdict stops here.
 *
 * Note this is per-netns: BPF_F_CURRENT_NETNS resolves to the namespace of the
 * device the program is attached to, which is where nhp-serverd's sockets live.
 */
static __always_inline bool has_local_flow(struct xdp_md *ctx, struct iphdr *iph,
                                           __be16 sport, __be16 dport, bool is_tcp) {
    struct bpf_sock_tuple tuple = {};
    struct bpf_sock *sk;
    bool owned;

    tuple.ipv4.saddr = iph->saddr;
    tuple.ipv4.daddr = iph->daddr;
    tuple.ipv4.sport = sport;
    tuple.ipv4.dport = dport;

    if (is_tcp)
        sk = bpf_sk_lookup_tcp(ctx, &tuple, sizeof(tuple.ipv4),
                               BPF_F_CURRENT_NETNS, 0);
    else
        sk = bpf_sk_lookup_udp(ctx, &tuple, sizeof(tuple.ipv4),
                               BPF_F_CURRENT_NETNS, 0);
    if (!sk)
        return false;

    owned = is_tcp ? sk->state != BPF_TCP_LISTEN : sk->dst_port != 0;

    bpf_sk_release(sk);
    return owned;
}

/* has_local_flow() for IPv6. Same helper, same verdict rule, only the tuple
 * differs -- and it has to exist, because the v6 branch below drops everything
 * it cannot account for: without this, the host's own v6 client traffic (a
 * resolver reachable over v6, a package mirror with an AAAA record) would hang
 * exactly the way the v4 DNS answers did before has_local_flow() was added. */
static __always_inline bool has_local_flow6(struct xdp_md *ctx, struct ipv6hdr *ip6h,
                                            __be16 sport, __be16 dport, bool is_tcp) {
    struct bpf_sock_tuple tuple = {};
    struct bpf_sock *sk;
    bool owned;

    __builtin_memcpy(tuple.ipv6.saddr, ip6h->saddr.in6_u.u6_addr32, sizeof(tuple.ipv6.saddr));
    __builtin_memcpy(tuple.ipv6.daddr, ip6h->daddr.in6_u.u6_addr32, sizeof(tuple.ipv6.daddr));
    tuple.ipv6.sport = sport;
    tuple.ipv6.dport = dport;

    if (is_tcp)
        sk = bpf_sk_lookup_tcp(ctx, &tuple, sizeof(tuple.ipv6),
                               BPF_F_CURRENT_NETNS, 0);
    else
        sk = bpf_sk_lookup_udp(ctx, &tuple, sizeof(tuple.ipv6),
                               BPF_F_CURRENT_NETNS, 0);
    if (!sk)
        return false;

    owned = is_tcp ? sk->state != BPF_TCP_LISTEN : sk->dst_port != 0;

    bpf_sk_release(sk);
    return owned;
}

/* The IPv6 half of the decision tree.
 *
 * Deliberately *not* symmetrical with IPv4 in one respect: there is no service
 * exception. The relay whitelist is a trie of IPv4 prefixes, so this program
 * cannot tell the relay's v6 address from anyone else's, and "SSH from the
 * relay" is a policy it can only express over v4. Nothing therefore reaches
 * tcp/22 or the knock port over v6; what gets in is neighbour discovery, the
 * two address/clock-keeping client protocols, and replies to flows the host
 * itself opened. Operators must keep reaching this host over IPv4 -- the
 * loader logs a Warning when the filtered interface has a global v6 address,
 * and deploy-server's "$SSH_CONNECTION is in the whitelist" gate fails outright
 * for a v6 peer address, since no entry can contain it.
 */
static __always_inline int handle_ipv6(struct xdp_md *ctx, void *l3, void *data_end) {
    struct ipv6hdr *ip6h = l3;
    if ((void *)(ip6h + 1) > data_end)
        return XDP_DROP;

    /* Walk the extension-header chain to the L4 header. Two hops is enough for
     * what a host on a v6 subnet actually receives (an MLD report is a
     * hop-by-hop router alert followed by ICMPv6, and a fragment header may
     * follow one of those); a longer chain falls through to the drop at the
     * bottom rather than being guessed at. */
    __u8 nexthdr = ip6h->nexthdr;
    void *cur = (void *)(ip6h + 1);

#pragma unroll
    for (int i = 0; i < 2; i++) {
        /* A fragment header is where v6 keeps the offset that makes IPv4's
         * non-first fragments unreadable, so it gets the same answer as the
         * IP_OFFSET branch in the main program: a fragment with a non-zero
         * offset has no L4 header to judge and is passed as inert (the first
         * fragment carries the ports and takes the full tree), while the first
         * fragment simply continues down the chain. */
        if (nexthdr == IPPROTO_FRAGMENT) {
            struct frag_hdr *frag = cur;
            if ((void *)(frag + 1) > data_end)
                return XDP_DROP;

            if (bpf_ntohs(frag->frag_off) & IP6_FRAG_OFFSET) {
                record_v6_packet(ctx, ACT_V6_FRAGMENT, ip6h, nexthdr, 0, 0, true);
                return XDP_PASS;
            }
            nexthdr = frag->nexthdr;
            cur = (void *)(frag + 1);
            continue;
        }

        if (nexthdr != IPPROTO_HOPOPTS && nexthdr != IPPROTO_ROUTING &&
            nexthdr != IPPROTO_DSTOPTS)
            break;

        struct ipv6_opt_hdr *ext = cur;
        if ((void *)(ext + 1) > data_end)
            return XDP_DROP;
        nexthdr = ext->nexthdr;
        cur += ((__u32)ext->hdrlen + 1) * 8;
        if (cur > data_end)
            return XDP_DROP;
    }

    if (nexthdr == IPPROTO_ICMPV6) {
        struct icmp6hdr *icmp6 = cur;
        if ((void *)(icmp6 + 1) > data_end)
            return XDP_DROP;

        /* Counted, not reported: ND and MLD are steady background chatter on
         * any v6 subnet, which is exactly what the old unconditional pass was
         * trying to keep out of the log. The per-window summary has the
         * numbers. */
        __u8 type = icmp6->icmp6_type;
        if (type == ICMPV6_PKT_TOOBIG ||
            (type >= ICMPV6_MLD_FIRST && type <= ICMPV6_ND_LAST)) {
            record_v6_packet(ctx, ACT_V6_ICMP_CONTROL, ip6h, nexthdr, 0, 0, false);
            return XDP_PASS;
        }
        record_v6_packet(ctx, ACT_DROP_V6_OTHER, ip6h, nexthdr, 0, 0, true);
        return XDP_DROP;
    }

    if (nexthdr == IPPROTO_TCP) {
        struct tcphdr *tcp = cur;
        if ((void *)(tcp + 1) > data_end)
            return XDP_DROP;

        if (has_local_flow6(ctx, ip6h, tcp->source, tcp->dest, true)) {
            record_v6_packet(ctx, ACT_V6_ESTABLISHED, ip6h, nexthdr, tcp->source, tcp->dest, false);
            return XDP_PASS;
        }
        record_v6_packet(ctx, ACT_DROP_V6_OTHER, ip6h, nexthdr, tcp->source, tcp->dest, true);
        return XDP_DROP;
    }

    if (nexthdr == IPPROTO_UDP) {
        struct udphdr *udp = cur;
        if ((void *)(udp + 1) > data_end)
            return XDP_DROP;

        /* The v6 spellings of the two exceptions the socket table cannot
         * answer for, reported under the same action codes as their v4
         * counterparts: same client, same reason, same thing an operator
         * greps for. */
        if (udp->source == bpf_htons(DHCP6_PORT_SERVER) &&
            udp->dest == bpf_htons(DHCP6_PORT_CLIENT)) {
            record_v6_packet(ctx, ACT_DHCP_CLIENT, ip6h, nexthdr, udp->source, udp->dest, true);
            return XDP_PASS;
        }
        if (udp->source == bpf_htons(NTP_PORT) && udp->dest == bpf_htons(NTP_PORT)) {
            record_v6_packet(ctx, ACT_NTP_CLIENT, ip6h, nexthdr, udp->source, udp->dest, false);
            return XDP_PASS;
        }
        if (has_local_flow6(ctx, ip6h, udp->source, udp->dest, false)) {
            record_v6_packet(ctx, ACT_V6_ESTABLISHED, ip6h, nexthdr, udp->source, udp->dest, false);
            return XDP_PASS;
        }
        record_v6_packet(ctx, ACT_DROP_V6_OTHER, ip6h, nexthdr, udp->source, udp->dest, true);
        return XDP_DROP;
    }

    record_v6_packet(ctx, ACT_DROP_V6_OTHER, ip6h, nexthdr, 0, 0, true);
    return XDP_DROP;
}

SEC("xdp")
int xdp_server_prog(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_DROP;

    switch (bpf_ntohs(eth->h_proto)) {
        /* ARP passes unreported: it keeps the host on its subnet, and
         * reporting it would drown the perf ring in neighbour traffic. */
        case ETH_P_ARP:  return XDP_PASS;
        case ETH_P_IPV6: return handle_ipv6(ctx, (void *)(eth + 1), data_end);
        case ETH_P_IP:   break;
        default:         return XDP_DROP;
    }

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return XDP_DROP;
    if (iph->ihl < 5)
        return XDP_DROP;

    /* Fragments other than the first have no L4 header: the bytes at ihl * 4
     * are payload, and reading them as one is how this program used to hand
     * itself a garbage source and destination port. It then dropped the
     * fragment as UDP_OTHER, so every datagram over the path MTU -- a knock
     * forwarded from the internet gateway's 1500 while this interface is at
     * 9001, a large EDNS0 answer from outside the VPC -- lost its tail and
     * timed out in reassembly, with nothing in user space to say why. Worse,
     * a crafted fragment whose payload happened to read 67 -> 68 or 123 -> 123
     * was passed by the exception branches below.
     *
     * So they are separated out and passed, under their own action code. A
     * non-first fragment is inert on its own: the kernel cannot deliver
     * anything to a socket until the *first* fragment arrives, and that one
     * carries the ports and goes through the full tree below. The cost is the
     * reassembly buffer an attacker can occupy with fragments whose first
     * never comes, which the kernel already bounds
     * (net.ipv4.ipfrag_high_thresh) and which is the same exposure any host
     * without a fragment-dropping firewall has. Dropping them instead would
     * make this filter, and not the network, the reason large knocks fail.
     *
     * First fragments (MF set, offset 0) fall through unchanged. udp->len is
     * the length of the whole datagram, not of the fragment, so the length
     * floor below still judges the knock and not the piece of it that
     * arrived. */
    if (bpf_ntohs(iph->frag_off) & IP_OFFSET) {
        record_packet(ctx, ACT_IPV4_FRAGMENT, iph, 0, 0, 0, true);
        return XDP_PASS;
    }

    /* L4 starts after the IP options, i.e. at ihl * 4 -- not at iph + 1, which
     * is only the same when there are none. */
    if (iph->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)iph + (iph->ihl * 4);
        if ((void *)(tcp + 1) > data_end)
            return XDP_DROP;

        if (tcp->dest == bpf_htons(SSH_PORT)) {
            if (is_relay_src(iph->saddr)) {
                /* One line per session, on the SYN: an operator wants to know
                 * that the relay opened an SSH connection and when, not to
                 * read every packet of the scp that follows. */
                bool is_syn = tcp->syn && !tcp->ack;
                record_packet(ctx, ACT_SSH_RELAY, iph, tcp->source, tcp->dest, 1, is_syn);
                return XDP_PASS;
            }
            record_packet(ctx, ACT_DROP_TCP_SSH_OTHER, iph, tcp->source, tcp->dest, 0, true);
            return XDP_DROP;
        }
        /* NHP is UDP-only, so a TCP connect to the knock port is a scanner
         * fingerprinting the host. Reported under its own action so that
         * traffic is countable separately from ordinary port scans. */
        if (bpf_ntohs(tcp->dest) == nhp_listen_port) {
            record_packet(ctx, ACT_DROP_TCP_NHP, iph, tcp->source, tcp->dest, 0, true);
            return XDP_DROP;
        }
        /* Anything left is either the reply side of a connection this host
         * opened (the plugin's SMTP submission to SES, certbot, dnf) or an
         * unsolicited packet. Only the socket table can tell them apart, and
         * it is asked after the two service ports above so that the policy on
         * those stays a pure function of the address -- a socket can never
         * re-open tcp/22 to a non-relay source. */
        if (has_local_flow(ctx, iph, tcp->source, tcp->dest, true)) {
            record_packet(ctx, ACT_TCP_ESTABLISHED, iph, tcp->source, tcp->dest, 0, false);
            return XDP_PASS;
        }
        record_packet(ctx, ACT_DROP_TCP_OTHER, iph, tcp->source, tcp->dest, 0, true);
        return XDP_DROP;
    }

    if (iph->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)iph + (iph->ihl * 4);
        if ((void *)(udp + 1) > data_end)
            return XDP_DROP;

        if (bpf_ntohs(udp->dest) != nhp_listen_port) {
            /* DHCP: the host's own address depends on this getting in, and
             * has_local_flow() cannot let it -- the client receives the lease
             * on a raw AF_PACKET socket or on one merely *bound* to udp/68, so
             * the socket lookup finds nothing, or finds something
             * indistinguishable from a listener. Dropping it is terminal for
             * the lease and takes the host off every port once the lease
             * expires. Admitted by port pair instead, as the AC's program has
             * always done it (`DHCP_PORT_R || DHCP_PORT_O` in
             * nhp/ebpf/xdp/nhp_ebpf_xdp.c); the pair is narrow enough that a
             * neighbour's broadcast DHCPDISCOVER to udp/67 is still dropped.
             * Reported per packet -- a renewal is a few datagrams an hour, and
             * it is the line an operator wants. See "Why each rule is there"
             * in terraform/demo/RUNBOOK.md for the outage behind this. */
            if (udp->source == bpf_htons(DHCP_PORT_SERVER) &&
                udp->dest == bpf_htons(DHCP_PORT_CLIENT)) {
                record_packet(ctx, ACT_DHCP_CLIENT, iph, udp->source, udp->dest, 0, true);
                return XDP_PASS;
            }
            /* NTP, for the same reason one layer up: unless it is configured
             * client-only (`port 0`), chronyd polls its servers from the
             * socket it has bound to udp/123 rather than from a connected
             * ephemeral one, and the answer (123 -> 123) then looks exactly
             * like an unsolicited query to has_local_flow(). Dropping it does
             * not take the host off the air, it lets its clock drift -- which
             * NHP notices on its own, because knock packets carry a timestamp
             * the server checks against a window. (A connected client socket
             * would be admitted by has_local_flow anyway, so this branch costs
             * nothing on a host configured that way.)
             *
             * Counted, not reported: chrony polls every 32-64s, so this would
             * be a steady trickle of lines saying the same thing. The
             * per-window [NHP-STAT] summary is where it belongs, and a clock
             * that stops being served shows up there as PKTS=0. */
            if (udp->source == bpf_htons(NTP_PORT) && udp->dest == bpf_htons(NTP_PORT)) {
                record_packet(ctx, ACT_NTP_CLIENT, iph, udp->source, udp->dest, 0, false);
                return XDP_PASS;
            }
            /* The datagram answer to something the host asked for -- a DNS
             * reply above all, which is what nhp-serverd needs before it can
             * reach SES at all. Connected sockets only, see has_local_flow(). */
            if (has_local_flow(ctx, iph, udp->source, udp->dest, false)) {
                record_packet(ctx, ACT_UDP_ESTABLISHED, iph, udp->source, udp->dest, 0, false);
                return XDP_PASS;
            }
            record_packet(ctx, ACT_DROP_UDP_OTHER, iph, udp->source, udp->dest, 0, true);
            return XDP_DROP;
        }
        /* The relay forwards agent knocks, including the short control
         * packets of the handshake, so it is exempt from the length floor. */
        if (is_relay_src(iph->saddr)) {
            record_packet(ctx, ACT_NHP_RELAY, iph, udp->source, udp->dest, 1, true);
            return XDP_PASS;
        }
        if (bpf_ntohs(udp->len) < nhp_min_udp_len) {
            record_packet(ctx, ACT_DROP_UDP_SHORT, iph, udp->source, udp->dest, 0, true);
            return XDP_DROP;
        }
        /* Reported: a knock is a handful of datagrams and the one packet class
         * the whole daemon exists for. A flood of knock-shaped datagrams is
         * capped by the budget like any other class, and shows up in full in
         * the per-window summary. */
        record_packet(ctx, ACT_NHP_DEFAULT, iph, udp->source, udp->dest, 0, true);
        return XDP_PASS;
    }

    /* ICMP is dropped -- the server answers no pings from the internet, and an
     * operator who needs one reaches the host over the relay's SSH path --
     * with exactly one exception: "fragmentation needed" (type 3 code 4).
     *
     * That is the signalling half of path-MTU discovery, and the host now has
     * outbound TCP connections worth protecting from a black hole: its
     * interface MTU is the VPC's 9001 while anything reached through the
     * internet gateway is 1500, so a large send on a path that has not
     * advertised a small enough MSS stalls silently, retransmitting full-size
     * segments forever. It is the failure mode that looks exactly like the DNS
     * one this program already caused, one layer up.
     *
     * The rest of type 3 stays dropped: those only turn a timeout into a
     * faster error, which is not worth handing an off-path attacker a way to
     * tear down flows. A spoofed frag-needed can lower a flow's PMTU (the
     * kernel floors it at 552 and matches it to a socket by the quoted header)
     * and nothing else. */
    if (iph->protocol == IPPROTO_ICMP) {
        struct icmphdr *icmp = (void *)iph + (iph->ihl * 4);
        if ((void *)(icmp + 1) > data_end)
            return XDP_DROP;

        if (icmp->type == ICMP_DEST_UNREACH && icmp->code == ICMP_FRAG_NEEDED) {
            record_packet(ctx, ACT_ICMP_FRAG_NEEDED, iph, 0, 0, 0, true);
            return XDP_PASS;
        }
    }

    record_packet(ctx, ACT_DROP_NONUDP, iph, 0, 0, 0, true);
    return XDP_DROP;
}

char _license[] SEC("license") = "Dual BSD/GPL";
