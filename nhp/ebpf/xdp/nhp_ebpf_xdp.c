#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>

#include "nhp_maps.h"
#include "nhp_xdp_stats.h"

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
/* The client protocols this program admits by port pair, named by role and
 * not by number, because the checks below are direction-sensitive: 67 is the
 * DHCP server's port and 68 the client's, and a rule that looked at either one
 * alone would be the hole described at the check. */
#define DHCP_PORT_SERVER 67
#define DHCP_PORT_CLIENT 68
#define NTP_PORT 123

/* The v6 spelling of the DHCP pair, for the same reason as the v4 one below:
 * a host that gets its v6 address from DHCPv6 rather than from SLAAC loses
 * it at lease expiry if the reply (547 -> 546) is dropped. */
#define DHCP6_PORT_SERVER 547
#define DHCP6_PORT_CLIENT 546

/* IPv6 next-header values: the three extension headers that may sit in front
 * of the L4 header on ordinary traffic (an MLD report carries a hop-by-hop
 * router alert), and ICMPv6. */
#define IPPROTO_HOPOPTS  0
#define IPPROTO_ROUTING  43
#define IPPROTO_DSTOPTS  60
#define IPPROTO_ICMPV6   58
#define IPPROTO_FRAGMENT 44

/* ICMPv6 types that keep a host working on an IPv6 subnet: "packet too big"
 * (the frag-needed of v6) and the contiguous MLD/ND block 130..137 --
 * multicast listener query/report/done, router and neighbour
 * solicitation/advertisement, redirect. Neighbour discovery *is* v6's ARP, so
 * dropping it would take the host off its own subnet. */
#define ICMPV6_PKT_TOOBIG 2
#define ICMPV6_MLD_FIRST  130
#define ICMPV6_ND_LAST    137

/* iph->frag_off, network order. OFFSET is non-zero on every fragment but the
 * first, which is the one that has no L4 header behind the IP header.
 * IP6_FRAG_OFFSET is the same field in an IPv6 fragment header (13 bits, the
 * low 3 being reserved + M). */
#define IP_OFFSET        0x1FFF
#define IP6_FRAG_OFFSET  0xFFF8

/* The AC's own action codes.
 *
 * These are a separate code space from the nhp-server's ACT_* defines: the
 * two programs never share a log file and each log is decoded with its own
 * name table (acActionName in nhp/utils/ebpf/engine_linux.go), so the same
 * number can mean different things in the two.
 *
 * Every value must stay below NHP_EVENT_ACTIONS (32) in nhp_xdp_stats.h: a
 * code at or past it is still judged, but it is neither counted in
 * nhp_action_stats nor rate limited by the per-class token bucket, so its
 * packets would be invisible to the [NHP-STAT] summary and unbounded on the
 * perf ring. */
#define ACT_AC_DROP_KNOCK_EXPIRED   0
#define ACT_AC_DROP_TCP_SYN         1
#define ACT_AC_DROP_TCP_STEALTH     2
#define ACT_AC_DROP_TCP_OTHER       3
#define ACT_AC_DROP_UDP_OTHER       4
#define ACT_AC_DROP_NON_TCP_UDP     5
#define ACT_AC_PASS_KNOCK_ADMITTED  6
#define ACT_AC_PASS_FLOW_TRACKED    7
#define ACT_AC_PASS_HOST_FLOW       8
#define ACT_AC_PASS_DHCP_CLIENT     9
#define ACT_AC_PASS_NTP_CLIENT      10
#define ACT_AC_PASS_IPV4_FRAGMENT   11
#define ACT_AC_PASS_V6_CONTROL      12
#define ACT_AC_PASS_V6_FRAGMENT     13
#define ACT_AC_PASS_V6_ESTABLISHED  14
#define ACT_AC_DROP_V6_OTHER        15
#define ACT_AC_PASS_HOST_ECHO       16
/* The v6 spellings of the two client protocols the socket table cannot answer
 * for. They get their own codes rather than reusing the v4 ones because the
 * IPv6 tree is a separate decision tree with its own classes -- V6_CONTROL,
 * V6_FRAGMENT, V6_ESTABLISHED -- and an operator reading the AC's counters
 * needs to be able to tell a v4 lease from a v6 one. */
#define ACT_AC_PASS_V6_DHCP_CLIENT  17
#define ACT_AC_PASS_V6_NTP_CLIENT   18

/* The whitelist maps, conn_track and their key/value structs live in
 * nhp_maps.h -- they are shared with tc_egress.c through their pins. */

/* The accept / deny audit channel.
 *
 * This ring and the one in nhp_xdp_stats.h are two independent paths: this
 * one carries the accepted and denied flows to the nhp_deny-*.log and
 * nhp_accept-*.log files; the nhp_events ring carries the counted,
 * rate-limited telemetry that the stats site reads.
 *
 * The audit ring used to be one line per packet with nothing on top of it.
 * That was tolerable while the security group let in only tcp/443 and the
 * knock port, because almost nothing reached the filter to be logged. It is
 * not tolerable with the group open to the whole internet: every packet that
 * arrives is then chosen by whoever sent it, and one sender is enough to fill
 * the disk the daemon runs on. So the ring is fed behind a budget of its own
 * -- the same per-class bucket and the same global cap the telemetry ring
 * uses, in the telemetry map's second bank of slots (legacy_event_budget_
 * available in nhp_xdp_stats.h) -- and the files behind it carry a daily byte
 * budget and are swept like the telemetry log's (readAcEvents and
 * acEventLogPrefixes in nhp/utils/ebpf/engine_linux.go).
 *
 * The verdict never depends on any of it: a packet that finds the budget
 * spent is still judged the same way, and only the line is not written. The
 * per-class totals in [NHP-STAT] count every packet, so what was not written
 * is still counted. */
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
//
// `verdict` is what lands in the log line (0 = DENY, 1 = ACCEPT); `class` is the
// verdict class the packet belongs to and is what the audit ring's own budget
// is keyed on. The two are different things: a knock admitted and a knock
// expired are both ACCEPT/DENY but different classes, and a budget keyed on
// the verdict alone would let a flood of one kind starve the other.
//
// The budget is spent before the event is emitted, so an over-budget packet
// costs a map lookup and nothing else. The verdict does not depend on it: the
// packet is judged exactly as before, only its line is not written. The
// per-class totals in [NHP-STAT] count every packet whatever the budget says.
static __always_inline int submit_event(void *ctx, __u8 verdict, __u8 class, __be32 src_ip, __be32 dst_ip, __be16 src_port, __be16 dst_port, __u8 protocol, __be16 len) {
    struct event_t ev = {};
    __u64 now = bpf_ktime_get_ns();

    if (!legacy_event_budget_available(class, now))
        return 0;

    ev.timestamp = now;
    ev.action = verdict; // 0 = DENY, 1 = ACCEPT
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

/* Build the conn_track entry for a flow the whitelists just admitted, with
 * `expire_time` taken from the whitelist entry that admitted it.
 *
 * For TCP `ttl_ns` is an idle window rather than the knock's remaining
 * lifetime: once a connection exists it is governed by activity, the way
 * netfilter's `--state ESTABLISHED` accept governs it in FilterMode 0, so an
 * established session is not cut mid-stream when the knock expires.
 *
 * `track` is what keeps that from becoming an open door. The peer picks its own
 * source port, so a refreshed window on a bare 4-tuple would let a peer that
 * knocked once hold the tuple open forever and open fresh connections on it.
 * For TCP the caller therefore passes `track` only for a pure SYN, i.e. only
 * the packet that opens the connection the knock authorised creates an entry;
 * every later SYN on the same tuple bypasses conn_track and has to satisfy the
 * whitelists again. A mid-stream TCP packet admitted by a whitelist (its entry
 * was evicted from the LRU, say) is passed without re-creating one --
 * fail-closed: it keeps needing a live knock.
 *
 * Non-TCP has no such packet to gate on, so it is bounded the other way
 * instead: nhp_ct_ingress_ttl_ns() caps its window at the time left on the
 * knock, and the forward branch below never refreshes it. A datagram flow
 * therefore cannot outlive the knock that admitted it, and one that is still
 * running when the entry ages out is simply re-admitted by the whitelists. */
static __always_inline void new_ingress_conn(struct ipv4_ct_tuple *ct_key, __u8 protocol, __u64 now, bool track, __u64 expire_time) {
    if (!track)
        return;

    struct conn_value new_val = {
        .timestamp = now,
        .last_timestamp = now,
        .ttl_ns = nhp_ct_ingress_ttl_ns(protocol, now, expire_time),
        .state = CT_ESTABLISHED,
        .flags = CT_FLAG_NONE,
        .rx_packets = 1,
        .tx_packets = 0,
    };
    bpf_map_update_elem(&conn_track, ct_key, &new_val, BPF_ANY);
}

/* Everything a whitelist branch does once it has decided to admit the packet.
 *
 * Besides the event and the conn_track entry it records the peer in
 * `knock_peers` -- peer address, the port on us it was admitted to, protocol --
 * which is the lasting record tc_egress.c's gate 4 needs. The whitelist entry
 * that admitted this packet is deleted by the cascade below the moment a later
 * packet finds it expired, so without this there would be nothing left to say
 * this peer was ever knock-gated by the time the AC answers it with the knock
 * gone; the AC's reply would then get an egress conn_track entry and the peer
 * an indefinitely refreshable way back in. See nhp_record_knocked_peer(), which
 * ignores TCP (its egress gate is exact and needs no record).
 *
 * Every branch that admits a packet must go through here: one that does not is
 * one whose knock shape reopens that hole.
 *
 * `ac_owns_flow` is the one case where nothing is recorded at all: the reverse
 * lookup below found this 4-tuple already claimed by the AC as a client
 * (CT_FLAG_AC_CLIENT), so the packet is the reply side of a flow the AC opened
 * and it only reached the whitelists because the peer's address also happens to
 * hold a knock -- an agent knocking from the nhp-server's own address, an NTP
 * server behind the same NAT. Recording either thing would break that flow:
 *
 *   - a `knock_peers` record never lapses, so it would shut gate 4 in
 *     tc_egress.c on the AC's own port for the life of the daemon, and have
 *     the reverse branch delete the entry the AC depends on;
 *   - an INGRESS entry would make tc_egress.c's gate 1 match the AC's own
 *     outgoing packets, which stops the EGRESS entry from being refreshed and
 *     lets it expire under a live flow.
 *
 * Nothing is lost by staying quiet: the peer gets no access it did not already
 * have from the AC's own flow, whose window only the AC's sending refreshes. */
static __always_inline int admit_ingress(void *ctx, struct iphdr *iph,
                                         struct ipv4_ct_tuple *ct_key, __u64 now,
                                         bool track, bool ac_owns_flow,
                                         __u64 expire_time) {
    /* A new flow the AC admitted because a knock was live. Worth a line: it
     * is a decision, and it is rare. drop=false -- an admitted flow is not
     * an attack, so it stays out of the per-port / per-source aggregation
     * behind the [NHP-SCAN] summary. */
    record_event(ctx, ACT_AC_PASS_KNOCK_ADMITTED, iph, ct_key->sport, ct_key->dport, 0, true, false);
    submit_event(ctx, 1, ACT_AC_PASS_KNOCK_ADMITTED, iph->saddr, iph->daddr, ct_key->sport, ct_key->dport, iph->protocol, iph->tot_len);
    if (ac_owns_flow)
        return XDP_PASS;
    nhp_record_knocked_peer(iph->saddr, ct_key->dport, iph->protocol, expire_time);
    new_ingress_conn(ct_key, iph->protocol, now, track, expire_time);
    return XDP_PASS;
}

/* Apply the TCP close semantics to an entry an inbound packet is refreshing: a
 * FIN or a RST cuts it down to the close window so it dies with the connection
 * rather than lingering for the full idle TTL, and the next packet of a flow
 * that is still running puts the full window (`idle_ttl_ns`) back.
 *
 * Neither flag deletes the entry, and neither cut is permanent, because XDP can
 * validate neither of them -- no sequence check here, and the kernel's RFC 5961
 * check runs after us. Deleting, or cutting for good, would hand any off-path
 * host that guesses the 4-tuple a way to cut an admitted session whose knock has
 * since expired, or to blackhole the replies of a connection the AC itself
 * opened. A FIN or RST the AC *sends* is a different matter: tc_egress.c
 * deletes on the RST, and its FIN is what makes the cut stick. See
 * nhp_ct_close_ttl_refresh(). */
static __always_inline void apply_close_ttl(struct conn_value *val, bool is_fin, bool is_rst, __u64 idle_ttl_ns) {
    nhp_ct_close_ttl_refresh(val, is_fin, is_rst, false /* from the peer */, idle_ttl_ns);
}

/* The reason a packet that no whitelist admitted is dropped.
 *
 * The rule is the nhp-server's, unchanged (nhp_server_xdp.c): a SYN without
 * an ACK is a peer asking to open a connection -- a connect or a SYN scan;
 * a packet with neither SYN, ACK nor RST is a stealth scan (FIN / NULL /
 * XMAS); anything else on TCP is TCP_OTHER. UDP is UDP_OTHER and every other
 * protocol is NON_TCP_UDP.
 *
 * Telemetry only. The caller drops the packet either way; this decides which
 * class it is counted under, never whether it is dropped. */
static __always_inline __u8 ac_classify_drop(__u8 proto, bool is_syn, bool is_ack, bool is_rst) {
    if (proto == IPPROTO_TCP) {
        if (is_syn)
            return ACT_AC_DROP_TCP_SYN;
        if (!is_ack && !is_rst)
            return ACT_AC_DROP_TCP_STEALTH;
        return ACT_AC_DROP_TCP_OTHER;
    }
    if (proto == IPPROTO_UDP)
        return ACT_AC_DROP_UDP_OTHER;
    return ACT_AC_DROP_NON_TCP_UDP;
}

/* Does this packet belong to a UDP flow the host itself has a socket for?
 *
 * The AC's own answer to that question is its conn_track: tc_egress.c
 * records every flow the host opens -- not only nhp-acd's -- and the reverse
 * branch in xdp_white_prog() serves the answer out of that entry. That covers
 * the host's TCP traffic (an SSH session the operator opened, the SSM agent's
 * HTTPS connection) and its UDP traffic whose source port is ephemeral (a
 * resolver's DNS query).
 *
 * The kernel's socket table answers the same question for the one case the
 * conn_track cannot: a socket bound to a fixed port below the ephemeral range
 * gets no egress entry (gate 3 in tc_egress.c), so the answer to it would be
 * dropped and the daemon behind it would hang until it timed out. The lookup
 * is the server's own (nhp_server_xdp.c) and asks the kernel which socket
 * would receive this packet; a connected UDP socket is a client (a service
 * socket is unconnected), so a match means the host opened this flow.
 *
 * Only UDP is asked. A TCP socket that is not a listener matches an *inbound*
 * connection to a port the AC serves too -- the very flow the knock gates --
 * and admitting it here would let a peer keep a connection open after its
 * knock closed. The conn_track already covers the host's own TCP traffic, and
 * it does so without that hole. */
static __always_inline bool has_local_udp_flow(struct xdp_md *ctx, struct iphdr *iph,
                                                __be16 sport, __be16 dport) {
    struct bpf_sock_tuple tuple = {};
    struct bpf_sock *sk;

    tuple.ipv4.saddr = iph->saddr;
    tuple.ipv4.daddr = iph->daddr;
    tuple.ipv4.sport = sport;
    tuple.ipv4.dport = dport;

    sk = bpf_sk_lookup_udp(ctx, &tuple, sizeof(tuple.ipv4), BPF_F_CURRENT_NETNS, 0);
    if (!sk)
        return false;

    /* A connected socket has a peer; an unconnected one is a listener by
     * another name. */
    bool owned = sk->dst_port != 0;

    bpf_sk_release(sk);
    return owned;
}

/* has_local_udp_flow() for IPv6. Same helper, same verdict rule, only the
 * tuple differs -- and it has to exist, because the v6 branch below drops
 * everything it cannot account for: without this the host's own v6 traffic
 * (a resolver reachable over v6, a package mirror with an AAAA record) would
 * hang exactly the way the v4 answers did before the lookup was added. */
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
 * Deliberately *not* symmetrical with IPv4 in one respect: there is no
 * service exception. The AC's whitelists are IPv4 tables, so this program
 * cannot tell one peer's v6 address from another's, and "the peer that
 * knocked may reach this port" is a policy it can only express over v4.
 * Nothing therefore reaches a knocked port over v6.
 *
 * What does get in is neighbour discovery, the two address/clock-keeping
 * client protocols, and the replies to the host's own v6 flows -- the same
 * shape as nhp_server_xdp.c, so a host on a v6 subnet still works and a
 * filter that says it drops everything but X does not have a hole the size
 * of an address family.
 *
 * The two client protocols are admitted by their port pair, not by the
 * socket table: a DHCPv6 lease and a chronyd answer both arrive on a socket
 * the kernel cannot tell apart from a listener, so the lookup below finds
 * nothing and the answer would be dropped. Their IPv4 counterparts are
 * admitted the same way, for the same reason. */
static __always_inline int handle_ipv6(struct xdp_md *ctx, void *l3, void *data_end) {
    struct ipv6hdr *ip6h = l3;
    if ((void *)(ip6h + 1) > data_end)
        return XDP_DROP;

    /* Walk the extension-header chain to the L4 header. Two hops is enough
     * for what a host on a v6 subnet actually receives (an MLD report is a
     * hop-by-hop router alert followed by ICMPv6, and a fragment header may
     * follow one of those); a longer chain falls through to the drop at the
     * bottom rather than being guessed at. */
    __u8 nexthdr = ip6h->nexthdr;
    void *cur = (void *)(ip6h + 1);

#pragma unroll
    for (int i = 0; i < 2; i++) {
        /* A fragment header is where v6 keeps the offset that makes IPv4's
         * non-first fragments unreadable, so it gets the same answer as the
         * IPv4 offset branch below: a fragment with a non-zero offset has no
         * L4 header to judge and is passed as inert (the first fragment
         * carries the ports and takes the full tree), while the first
         * fragment simply continues down the chain. */
        if (nexthdr == IPPROTO_FRAGMENT) {
            struct frag_hdr *frag = cur;
            if ((void *)(frag + 1) > data_end)
                return XDP_DROP;

            if (bpf_ntohs(frag->frag_off) & IP6_FRAG_OFFSET) {
                record_v6_event(ctx, ACT_AC_PASS_V6_FRAGMENT, ip6h, nexthdr, 0, 0, false, false);
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
         * any v6 subnet, which is exactly what an unconditional pass would
         * have been keeping out of the log. The per-window summary has the
         * numbers. */
        __u8 type = icmp6->icmp6_type;
        if (type == ICMPV6_PKT_TOOBIG ||
            (type >= ICMPV6_MLD_FIRST && type <= ICMPV6_ND_LAST)) {
            record_v6_event(ctx, ACT_AC_PASS_V6_CONTROL, ip6h, nexthdr, 0, 0, false, false);
            return XDP_PASS;
        }
        record_v6_event(ctx, ACT_AC_DROP_V6_OTHER, ip6h, nexthdr, 0, 0, true, true);
        return XDP_DROP;
    }

    if (nexthdr == IPPROTO_TCP) {
        struct tcphdr *tcp = cur;
        if ((void *)(tcp + 1) > data_end)
            return XDP_DROP;

        /* The reply to a connection the host itself opened. */
        if (has_local_flow6(ctx, ip6h, tcp->source, tcp->dest, true)) {
            record_v6_event(ctx, ACT_AC_PASS_V6_ESTABLISHED, ip6h, nexthdr, tcp->source, tcp->dest, false, false);
            return XDP_PASS;
        }
        record_v6_event(ctx, ACT_AC_DROP_V6_OTHER, ip6h, nexthdr, tcp->source, tcp->dest, true, true);
        return XDP_DROP;
    }

    if (nexthdr == IPPROTO_UDP) {
        struct udphdr *udp = cur;
        if ((void *)(udp + 1) > data_end)
            return XDP_DROP;

        /* The v6 spellings of the two exceptions the socket table cannot
         * answer for, reported under their own action codes so a v6 lease is
         * not confused with a v4 one. Same reason, same shape, and the same
         * order as the IPv4 branch below and as nhp_server_xdp.c: the port
         * pair is the exception, the socket table is the fallback.
         *
         * DHCPv6: the host's own address depends on this getting in, and
         * has_local_flow6() cannot let it -- the client receives the lease on
         * a socket that is merely bound to udp/546, which the lookup refuses
         * because it cannot tell it apart from a listener. Dropping it is
         * terminal for the lease, and a host that loses its address loses
         * every port with it. Reported per packet: a renewal is a few
         * datagrams an hour, and it is the line an operator wants.
         *
         * NTPv6: chrony polls its servers from the socket it has bound to
         * udp/123 rather than from a connected ephemeral one, and the answer
         * (123 -> 123) then looks exactly like an unsolicited query to the
         * lookup. Counted, not reported: chrony polls every 32-64s, so this
         * would be a steady trickle of lines saying the same thing. */
        if (udp->source == bpf_htons(DHCP6_PORT_SERVER) &&
            udp->dest == bpf_htons(DHCP6_PORT_CLIENT)) {
            record_v6_event(ctx, ACT_AC_PASS_V6_DHCP_CLIENT, ip6h, nexthdr,
                            udp->source, udp->dest, true, false);
            return XDP_PASS;
        }
        if (udp->source == bpf_htons(NTP_PORT) && udp->dest == bpf_htons(NTP_PORT)) {
            record_v6_event(ctx, ACT_AC_PASS_V6_NTP_CLIENT, ip6h, nexthdr,
                            udp->source, udp->dest, false, false);
            return XDP_PASS;
        }

        if (has_local_flow6(ctx, ip6h, udp->source, udp->dest, false)) {
            record_v6_event(ctx, ACT_AC_PASS_V6_ESTABLISHED, ip6h, nexthdr, udp->source, udp->dest, false, false);
            return XDP_PASS;
        }
        record_v6_event(ctx, ACT_AC_DROP_V6_OTHER, ip6h, nexthdr, udp->source, udp->dest, true, true);
        return XDP_DROP;
    }

    record_v6_event(ctx, ACT_AC_DROP_V6_OTHER, ip6h, nexthdr, 0, 0, true, true);
    return XDP_DROP;
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
        /* IPv6 used to be an unconditional pass, on the grounds that the demo
         * hosts have no v6 address. That left every v6 listener -- sshd's
         * [::]:22 above all -- reachable unfiltered the day a v6 CIDR is added
         * to the VPC, with the deploy's "filter attached" check still green.
         * With the security group open to the whole internet that is not an
         * academic gap: it is a standing bypass into the host. So v6 is
         * filtered with the same tree as v4, minus the whitelists, which are
         * IPv4 tables. */
        case ETH_P_IPV6: return handle_ipv6(ctx, (void *)(eth + 1), data_end);
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
     * fragment as UDP_OTHER -- or, worse, passed it, because the garbage
     * happened to read 67 -> 68 or 53 -> anything.
     *
     * So they are separated out and passed, under their own action code. A
     * non-first fragment is inert on its own: the kernel cannot deliver
     * anything to a socket until the *first* fragment arrives, and that one
     * carries the ports and goes through the full tree below. Dropping them
     * instead would make this filter, and not the network, the reason a large
     * datagram fails. */
    if (bpf_ntohs(iph->frag_off) & IP_OFFSET) {
        /* Counted, not reported: a fragment flood would otherwise write one
         * line per packet. */
        record_event(ctx, ACT_AC_PASS_IPV4_FRAGMENT, iph, 0, 0, 0, false, false);
        return XDP_PASS;
    }

    /* The L4 header starts after the IP options, i.e. at ihl * 4 -- not at
     * iph + 1, which is only the same thing when there are no options. The
     * conn_track keys built here have to line up byte for byte with the ones
     * tc_egress.c builds (it reads ihl * 4 too), or the gates there miss. */
    ct_key.nexthdr = iph->protocol;
    bool is_tcp = false, is_syn = false, is_fin = false, is_rst = false;
    /* Telemetry only: ac_classify_drop() needs to tell a pure SYN from a
     * stealth probe, and that means knowing whether the ACK bit is set. */
    bool is_ack = false;
    if (iph->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)iph + (iph->ihl * 4);
        if ((void *)(tcp + 1) > data_end) {
            return XDP_DROP;
        }
        ct_key.sport = tcp->source;
        ct_key.dport = tcp->dest;
        is_tcp = true;
        /* A pure SYN (no ACK) is a peer asking to open a *new* connection, so
         * it must always be decided by the whitelists -- never by a conn_track
         * entry left over from an earlier one on the same 4-tuple. */
        is_syn = tcp->syn && !tcp->ack;
        is_fin = tcp->fin;
        is_rst = tcp->rst;
        is_ack = tcp->ack;

        if (tcp->dest == bpf_htons(22)) {
            /* SSH is not filtered here, because the security group is what
             * restricts it: tcp/22 is carved out of the all-ports range and
             * stays open to the relay SG only. The nhp-server's own filter
             * pins the same port to its relay whitelist in the program
             * instead, because that host has no such group rule to lean on.
             * Widening the AC's group does not widen tcp/22. */
            return XDP_PASS;
        }
    } else if (iph->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)iph + (iph->ihl * 4);
        if ((void *)(udp + 1) > data_end)
            return XDP_DROP;
        ct_key.sport = udp->source;
        ct_key.dport = udp->dest;
    }

    /* The two client protocols whose answers arrive on a socket the kernel
     * cannot tell apart from a listener, so neither the conn_track nor the
     * socket table can let them in on their own. Both are admitted by port
     * pair, exactly as nhp_server_xdp.c admits them, and both are the
     * complete list: "some daemon's replies are being dropped" is not on its
     * own a reason to extend it, because the alternative -- admit anything
     * addressed to a bound port -- throws the whole policy away.
     *
     * Both halves of each pair are tested, and that is the whole point: a
     * source port is chosen by whoever sends the packet, so it proves nothing
     * about who sent it, and a rule that looked at either port on its own is
     * what made this block a bypass.
     *
     *   - `dport == 67 || dport == 68` let anyone reach udp/67 and udp/68 on
     *     the AC and on every host behind it, including a neighbour's
     *     broadcast DHCPDISCOVER;
     *   - `sport == 53` was far worse. It proved nothing about the sender and
     *     this test ran *before* conn_track and before every whitelist, so
     *     anything sent from udp/53 reached any UDP port the AC protects --
     *     DNS, QUIC / HTTP3, WireGuard, game servers -- with no knock at all,
     *     which is the "nothing is visible until you knock" guarantee given
     *     away for a single `--sport 53`.
     *
     * DHCP: the host's own address depends on the reply getting in, and its
     * loss is terminal for the lease. The client sends from 0.0.0.0:68 to
     * 255.255.255.255:67, so the entry tc_egress.c would write is keyed on a
     * broadcast address the reply does not come back from -- and the reply is
     * received on a raw AF_PACKET socket, or on one merely bound to udp/68,
     * either way nothing the reverse lookup can match. The pair is narrow
     * enough that a neighbour's broadcast DHCPDISCOVER to udp/67 is still
     * dropped.
     *
     * NTP: chronyd polls its servers from the socket it has bound to udp/123
     * rather than from a connected ephemeral one, so the answer (123 -> 123)
     * looks exactly like an unsolicited query. Dropping it does not take the
     * host off the air, it lets its clock drift -- which NHP notices on its
     * own, because knock packets carry a timestamp the server checks against
     * a window.
     *
     * DNS needs no exception of its own: it is admitted the same way the
     * AC's other client traffic is. The query goes out the interface
     * tc_egress.c is attached to -- the loader attaches both programs to the
     * same one, as a pair, and tears both down if either attach fails -- and
     * the EGRESS conn_track entry it leaves behind is what the reverse branch
     * below answers the reply from. The same holds for the answer to a query
     * whose SOURCE port gate 3 in tc_egress.c accepts: at or above the host's
     * ip_local_port_range floor, or one of the fixed client ports in
     * is_wellknown_client_port(). A stub resolver (glibc, systemd-resolved)
     * takes a kernel-assigned ephemeral port and is covered. A recursive
     * resolver that randomises its own outgoing port over 1024-65535 --
     * BIND's default use-v4-udp-ports, Unbound configured the same way --
     * running ON the AC and over this interface is not: the answers to its
     * low-port queries have no entry to be admitted from, and the symptom is
     * intermittent rather than clean. That is the same limitation every other
     * fixed-low-port UDP client on the AC already has; see "Known limitation:
     * fixed low UDP source ports" in terraform/demo/RUNBOOK.md for the
     * symptom and the ways out. Bringing `sport == 53` back is not one of
     * them -- it does not identify a resolver, only a sender that typed 53.
     *
     * The IPv6 tree above has its own pair for the same two protocols
     * (DHCPv6 547 -> 546, NTPv6 123 -> 123), because there is no egress
     * conn_track for v6 to admit the answers from. */
    if (iph->protocol == IPPROTO_UDP) {
        if (ct_key.sport == bpf_htons(DHCP_PORT_SERVER) && ct_key.dport == bpf_htons(DHCP_PORT_CLIENT)) {
            record_event(ctx, ACT_AC_PASS_DHCP_CLIENT, iph, ct_key.sport, ct_key.dport, 0, true, false);
            return XDP_PASS;
        }
        if (ct_key.sport == bpf_htons(NTP_PORT) && ct_key.dport == bpf_htons(NTP_PORT)) {
            /* Counted, not reported: chrony polls every 32-64s, so this would
             * be a steady trickle of lines saying the same thing. */
            record_event(ctx, ACT_AC_PASS_NTP_CLIENT, iph, ct_key.sport, ct_key.dport, 0, false, false);
            return XDP_PASS;
        }
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

            /* The reply to a ping the AC sent itself.
             *
             * An echo reply used to be passed unconditionally, from anywhere:
             * with the security group open that let any host on the internet
             * reach whatever answers its pings on this host. Narrowing it to
             * the whitelist entry the echo request needs closed that, but it
             * also stopped the AC from getting answers to its own pings --
             * a diagnostic the operator runs on the host, and the way the
             * AC's own reachability is checked.
             *
             * So the reply is admitted on one of two grounds: the peer
             * knocked for ICMP access (the whitelist entry below), or the AC
             * pinged that peer itself and this is the answer (the `icmp_echo`
             * entry tc_egress.c wrote when the request went out). The
             * identifier is part of that key, so only the reply to the
             * request we actually sent is admitted, and the entry expires in
             * NHP_ICMP_ECHO_TTL_NS. */
            if (icmp->type == ICMP_ECHOREPLY) {
                struct icmp_echo_key echokey = {
                    .peer_ip = iph->saddr,
                    .id = icmp->un.echo.id,
                };
                struct icmp_echo_value *echo_val = bpf_map_lookup_elem(&icmp_echo, &echokey);
                if (echo_val && echo_val->expire_time >= now) {
                    record_event(ctx, ACT_AC_PASS_HOST_ECHO, iph, 0, 0, 0, false, false);
                    return XDP_PASS;
                }
            }

            /* Both directions are decided by the same whitelist entry. The
             * reply is the answer to a ping of ours, so it is admitted on the
             * same entry the request needs -- the peer that knocked for ICMP,
             * and only while its window is open. */
            //Lookup icmpwhitelist entry
            struct icmpwhitelist_value *iw_val = bpf_map_lookup_elem(&icmpwhitelist, &icmpkey);
            if (!iw_val) {
                /* No ICMP whitelist for this pair at all: nobody knocked
                 * for it, so it counts as an unsolicited probe. */
                record_event(ctx, ACT_AC_DROP_NON_TCP_UDP, iph, ct_key.sport, ct_key.dport, 0, true, true);
                return XDP_DROP;
            }
            // Check if whitelist entry has expired
            if (iw_val->expire_time < now) {
                bpf_map_delete_elem(&icmpwhitelist, &icmpkey);
                /* The entry existed but its window closed: a peer that did
                 * knock, unlike one that never did.
                 *
                 * drop=false: an expired knock is not a scan. It is a
                 * legitimate peer that came back too late, and counting it
                 * would put an authorized user's address into the per-source
                 * table behind the [NHP-SCAN] summary -- and, through the
                 * site's own totals, into the published attacker counts. See
                 * CatKnockExpired in endpoints/xdplog-report/classify.go. */
                record_event(ctx, ACT_AC_DROP_KNOCK_EXPIRED, iph, ct_key.sport, ct_key.dport, 0, true, false);
                return XDP_DROP;
            }
            // Check if source IP is icmpwhitelisted and allowed
            if (iw_val->allowed == 1) {
                return XDP_PASS;
            }
            /* The entry exists but does not allow this peer: fall through to
             * the whitelist lookups below, which drop it. */
        }
    }

    ct_key.saddr = iph->saddr;
    ct_key.daddr = iph->daddr;
    ct_key.flags = CT_DIR_INGRESS;
    struct conn_value *existing_val;

    /* The (peer, AC) pair a knock for an "any"-protocol resource writes. Looked
     * up here rather than with the other whitelists below because the reverse
     * branch needs it too -- see there. */
    struct sdwhitelist_key sdkey = {
        .src_ip = iph->saddr,
        .dst_ip = iph->daddr
    };
    struct sdwhitelist_value *sd_val = bpf_map_lookup_elem(&sdwhitelist, &sdkey);
    /* `allowed` is part of the test so that skipping the reverse branch below
     * can never lose a packet the sdwhitelist lookup would not then admit. */
    bool peer_has_live_knock = sd_val && sd_val->allowed == 1 &&
                               sd_val->expire_time >= now;

    /*
     * Forward (CT_DIR_INGRESS) hit: a flow the peer opened and that we admitted
     * via one of the whitelists while its knock was valid.
     *
     * For TCP it lives on an idle window -- `timestamp` is pushed forward by
     * every packet -- so an established session is not cut mid-stream when the
     * knock expires. That is what FilterMode 0 does too: the AC's iptables
     * baseline accepts `-m state --state ESTABLISHED` ahead of the per-knock
     * ipset, so expiry there blocks new connections only.
     *
     * A refreshed window on the 4-tuple alone would not be safe, because the
     * peer chooses its own source port: it could bind a fixed one, keep the
     * entry alive with one packet per window, and open new connections on it
     * long after the knock closed. So for TCP the entry tracks the connection,
     * not just the tuple, exactly where netfilter draws the line:
     *
     *   - a pure SYN skips this branch entirely and goes to the whitelists
     *     below (netfilter turns a SYN on a closed/TIME_WAIT entry back into
     *     NEW, so `--state ESTABLISHED` does not match it either). Re-using a
     *     source port therefore buys the peer nothing;
     *   - a FIN or a RST cuts the entry down to NHP_CT_CLOSE_TTL_NS, enough for
     *     the rest of the close handshake and no more, and the next packet of a
     *     flow that is still running restores the full window. See
     *     apply_close_ttl() for why an inbound RST neither deletes the entry
     *     outright nor cuts it for good.
     *
     * Non-TCP has no handshake to bind an entry to one exchange, so it is not
     * refreshed here at all: its window was capped at the knock's remaining
     * lifetime when it was created (nhp_ct_ingress_ttl_ns) and it is allowed to
     * run out, which is what keeps a datagram flow from outliving its knock.
     *
     * Once past its window the entry is dropped and the packet falls through to
     * the whitelist lookups, so a flow that ages out -- or goes quiet and comes
     * back -- needs a knock that is still valid.
     */
    existing_val = bpf_map_lookup_elem(&conn_track, &ct_key);
    if (existing_val && !is_syn) {
        if (check_conn_expiry(existing_val)) {
            /* Idle out. Fall through to the whitelist lookups below: if the
             * knock is still valid the flow is simply admitted again. */
            bpf_map_delete_elem(&conn_track, &ct_key);
        } else {
            struct conn_value new_val = *existing_val;
            new_val.tx_packets++;
            new_val.last_timestamp = now;
            if (is_tcp) {
                new_val.timestamp = now;
                /* The window a TCP ingress entry was created with
                 * (nhp_ct_ingress_ttl_ns), i.e. what a packet that is not a
                 * close restores. */
                apply_close_ttl(&new_val, is_fin, is_rst, NHP_CT_TCP_IDLE_TTL_NS);
            }
            bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_EXIST);
            /* A flow the AC already admitted. report=false: this is the
             * bulk of the traffic the AC carries, and a line per packet would
             * write more log than every other class together. It is still
             * counted, and the per-window summary reports it. */
            record_event(ctx, ACT_AC_PASS_FLOW_TRACKED, iph, ct_key.sport, ct_key.dport, 0, false, false);
            return XDP_PASS;
        }
    }

    /* Has this peer ever been admitted to this port of ours by a knock? Unlike
     * `peer_has_live_knock` the answer survives the knock, which is what keeps
     * the reverse branch below from serving a peer whose knock has expired. It
     * is never asked about a port the AC has claimed as its own client socket,
     * because no record is written for one -- see admit_ingress().
     *
     * Read here rather than with the sdwhitelist lookup above so that the
     * forward branch -- the path every packet of an already-admitted flow takes
     * -- does not pay for a map lookup only the reverse branch uses, and before
     * reverseTuple() while ct_key.dport is still the port on us. Non-TCP only:
     * the TCP egress gate is exact, so an EGRESS entry for a TCP flow really is
     * a connection the AC opened, and nothing is recorded for TCP anyway. */
    bool peer_knocked_port = iph->protocol != IPPROTO_TCP &&
                             nhp_peer_knocked(iph->saddr, ct_key.dport, iph->protocol);

    /*
     * Reverse (CT_DIR_EGRESS) hit: the reply to a connection the AC itself
     * opened, whose entry was created by tc_egress.c. The gates there are what
     * keep a knock-authorised flow from ever having an EGRESS entry, so this
     * branch is not supposed to see one -- and how much an inbound packet may
     * do to it is decided by how far the gates can be trusted for that
     * protocol:
     *
     *   TCP  -- the pure-SYN gate is exact (a listening socket never sends
     *           one), so `timestamp` is refreshed here and the TTL is a sliding
     *           idle window, which a long download needs when the AC is only
     *           receiving.
     *   else -- the UDP gates are heuristics (client-shaped source port, peer
     *           holds no knock entry). Refreshing would let a peer slide the
     *           window forward for as long as it keeps sending, so it is
     *           deliberately not done. Such an entry lives on the AC's own
     *           sending alone:
     *           tc_egress.c pushes it forward on each outgoing packet, which is
     *           what keeps nhp-acd's long-lived UDP channel to the nhp-server
     *           (one socket, NHP_KPL every 20s) answerable, and it dies
     *           NHP_CT_EGRESS_IDLE_TTL_NS after the AC last spoke on it.
     *
     * As in the forward branch a pure SYN is never answered from here, and a
     * FIN or RST cuts the entry down to the close window -- non-permanently and
     * without deleting it -- instead of removing it (apply_close_ttl); the
     * window it restores is the egress one the entry was created with, since
     * tc_egress.c wrote it. An entry that has idled out falls through to the
     * whitelist lookups rather than dropping the packet outright -- the same
     * treatment the forward branch gives, and it lets a live knock admit the
     * packet on its own merits.
     *
     * A peer that holds a live "any"-protocol knock is not served from here
     * either, nor -- for non-TCP -- is one that `knock_peers` says has been
     * admitted to this port of ours at any point. Gate 4 in tc_egress.c already
     * refuses to create an EGRESS entry for such a peer, and this is the other
     * half of that: an entry that pre-dates the knock, or that the gates
     * missed, must not become a second way in. Nothing is lost by skipping the
     * branch -- the whitelist lookups below admit the packet if a knock really
     * is live, and they are the only thing that may.
     *
     * The two differ in what happens to the entry. A live knock is temporary,
     * so the entry is only skipped: once the knock lapses it may serve the AC's
     * own flow again. A `knock_peers` record never lapses, so an entry it
     * blocks can never serve anything again -- and leaving it would strand it,
     * because tc_egress.c's refresh path does not test expiry, so the AC's own
     * sending would keep resurrecting an entry no inbound packet is allowed to
     * reach the delete for. It would sit in the LRU for the life of the daemon
     * and show up in the conn_track dump operators are told to check after a
     * knock expires (see the UDP case in terraform/demo/RUNBOOK.md). Delete it
     * instead; gate 4 makes sure it is not created again.
     *
     * Both of those give way to an entry the AC has claimed as its own client
     * flow (CT_FLAG_AC_CLIENT, see nhp_ct_is_ac_client()). Such an entry was
     * created while nothing said the peer was knock-gated on this port, so it
     * is not "an entry the gates missed" -- it is the AC talking to a host that
     * has since acquired a knock, which happens as soon as an agent knocks from
     * an address the AC also talks to: the nhp-server's own address (shared by
     * NAT, or a small office running both), an NTP server, a resolver. Skipping
     * the branch there would be harmless on its own, but what follows is not:
     * the whitelists would admit the packet, record the AC's own port in
     * `knock_peers` and give the flow an INGRESS entry, and the pair of those
     * takes the flow down for good once the knock expires -- gate 1 stops the
     * EGRESS entry from being refreshed, this branch then deletes it on the
     * `knock_peers` record, and gate 4 refuses to create it again. For the AC's
     * channel to the nhp-server that means every NHP_AOP dropped until nhp-acd
     * restarts. So a claimed entry is served here while its window is open and
     * is exempt from the `knock_peers` delete, and admit_ingress() records
     * nothing for it.
     *
     * An expired claimed entry still admits nothing -- the packet falls through
     * to the whitelists like any other -- but it is kept rather than deleted,
     * because it is the AC's standing claim on the tuple and dropping it would
     * let the very next admitted packet write that `knock_peers` record. The
     * AC's next outgoing packet revives it.
     */
    reverseTuple(&ct_key);
    existing_val = bpf_map_lookup_elem(&conn_track, &ct_key);
    bool ac_owns_flow = nhp_ct_is_ac_client(existing_val);
    int reverse_verdict = -1;
    if (existing_val && !is_syn && (ac_owns_flow || !peer_has_live_knock)) {
        if (check_conn_expiry(existing_val)) {
            /* Idled out. A claimed entry is kept as the AC's standing claim on
             * the tuple; either way the packet goes on to the whitelists. */
            if (!ac_owns_flow)
                bpf_map_delete_elem(&conn_track, &ct_key);
        } else if (peer_knocked_port && !ac_owns_flow) {
            bpf_map_delete_elem(&conn_track, &ct_key);
        } else {
            struct conn_value new_val = *existing_val;
            new_val.rx_packets++;
            if (is_tcp) {
                new_val.timestamp = now;
                apply_close_ttl(&new_val, is_fin, is_rst, NHP_CT_EGRESS_IDLE_TTL_NS);
            }
            new_val.last_timestamp = now;
            bpf_map_update_elem(&conn_track, &ct_key, &new_val, BPF_EXIST);
            reverse_verdict = XDP_PASS;
        }
    }
    /* Back to the ingress orientation: the whitelist lookups below and
     * new_ingress_conn() both read ct_key. */
    reverseTuple(&ct_key);
    if (reverse_verdict >= 0)
        return reverse_verdict;

    /* A TCP flow is tracked from the packet that opens it and from no other.
     * See new_ingress_conn(). A tuple the AC owns as a client is not tracked at
     * all -- admit_ingress() short-circuits on `ac_owns_flow`. */
    bool track_new_conn = !is_tcp || is_syn;

    struct whitelist_key key = {
        .src_ip = iph->saddr,
        .dst_ip = iph->daddr,
        .dst_port = ct_key.dport,
        .protocol = iph->protocol
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
    //sdwhitelist was looked up above (sd_val), the reverse branch needs it too
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
            /* The peer knocked and its window has closed. A legitimate peer
             * that came back too late, as opposed to one that never knocked.
             *
             * drop=false: an expired knock is not a scan, so it stays out of
             * the per-port and per-source tables behind the [NHP-SCAN]
             * summary. The published site excludes the class from its
             * attacker totals for the same reason -- see CatKnockExpired in
             * endpoints/xdplog-report/classify.go. */
            record_event(ctx, ACT_AC_DROP_KNOCK_EXPIRED, iph, ct_key.sport, ct_key.dport, 0, true, false);
            return XDP_DROP;
        }
        if (w_val->allowed == 1) {
            return admit_ingress(ctx, iph, &ct_key, now, track_new_conn, ac_owns_flow, expire_time);
        }
    }
    if (sd_val) {
        __u64 expire_time = sd_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&sdwhitelist, &sdkey);
            record_event(ctx, ACT_AC_DROP_KNOCK_EXPIRED, iph, ct_key.sport, ct_key.dport, 0, true, false);
            return XDP_DROP;
        }
        if (sd_val->allowed == 1) {
            return admit_ingress(ctx, iph, &ct_key, now, track_new_conn, ac_owns_flow, expire_time);
        }
    }

    if (sp_val) {
        __u64 expire_time = sp_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&src_port, &spkey);
            record_event(ctx, ACT_AC_DROP_KNOCK_EXPIRED, iph, ct_key.sport, ct_key.dport, 0, true, false);
            return XDP_DROP;
        }
        if (sp_val->allowed == 1) {
            return admit_ingress(ctx, iph, &ct_key, now, track_new_conn, ac_owns_flow, expire_time);
        }
    }

    if (pl_val) {
        __u64 expire_time = pl_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&port_list, &pl_key);
            record_event(ctx, ACT_AC_DROP_KNOCK_EXPIRED, iph, ct_key.sport, ct_key.dport, 0, true, false);
            return XDP_DROP;
        }
        if (pl_val->allowed == 1) {
            return admit_ingress(ctx, iph, &ct_key, now, track_new_conn, ac_owns_flow, expire_time);
        }
    }
    if (pp_val) {
        __u64 expire_time = pp_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&protocol_port, &pp_key);
            record_event(ctx, ACT_AC_DROP_KNOCK_EXPIRED, iph, ct_key.sport, ct_key.dport, 0, true, false);
            return XDP_DROP;
        }
        if (pp_val->allowed == 1) {
            return admit_ingress(ctx, iph, &ct_key, now, track_new_conn, ac_owns_flow, expire_time);
        }
    }
    /* The host's own UDP traffic, asked last.
     *
     * Everything above is the AC's protocol: a knock opens a port, the
     * whitelists admit the peer that asked for it, and the conn_track carries
     * the flows that follow. What is left is the answer to a datagram the
     * *host* sent -- a DNS reply above all, without which nhp-acd cannot
     * resolve the nhp-server and every agent's knock fails.
     *
     * This is asked last on purpose. The conn_track above already serves the
     * host's own flows, with the bookkeeping that goes with them (the
     * `knock_peers` record, the egress entry the reply side is served from);
     * the socket table is the backstop for the one case the conn_track cannot
     * cover, a socket bound to a fixed port below the ephemeral range. See
     * has_local_udp_flow(). */
    if (iph->protocol == IPPROTO_UDP && has_local_udp_flow(ctx, iph, ct_key.sport, ct_key.dport)) {
        record_event(ctx, ACT_AC_PASS_HOST_FLOW, iph, ct_key.sport, ct_key.dport, 0, false, false);
        return XDP_PASS;
    }
    /* No whitelist admitted the packet. Which class it is depends on what the
     * packet was trying to do; either way it is dropped. */
    __u8 drop_class = ac_classify_drop(iph->protocol, is_syn, is_ack, is_rst);
    record_event(ctx, drop_class, iph, ct_key.sport, ct_key.dport, 0, true, true);
    submit_event(ctx, 0, drop_class, iph->saddr, iph->daddr, ct_key.sport, ct_key.dport, iph->protocol, iph->tot_len);
    return XDP_DROP;
}

char _license[] SEC("license") = "Dual BSD/GPL";