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

/* ICMP destination-unreachable / fragmentation-needed, i.e. the one ICMP
 * message a host that opens outbound TCP connections cannot do without. */
#define ICMP_DEST_UNREACH  3
#define ICMP_FRAG_NEEDED   4

#define SSH_PORT 22

/* Client protocols whose answers arrive on a socket the kernel cannot tell
 * apart from a listener, so has_local_flow() cannot admit them. See the DHCP
 * and NTP branches below for why each one is here; nothing may be added to
 * this list without the same kind of argument. */
#define DHCP_PORT_SERVER 67
#define DHCP_PORT_CLIENT 68
#define NTP_PORT         123

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
 * address, or a knock that passed). Per-CPU so the counter needs no atomics;
 * the effective ceiling is NHP_EVENT_BURST * nr_cpus per second per class,
 * which is ample for diagnosis and bounded regardless of offered load. */
#define NHP_EVENT_ACTIONS 16
#define NHP_EVENT_WINDOW_NS 1000000000ULL
#define NHP_EVENT_BURST 16

struct nhp_rate_state {
    __u64 window_start;
    __u32 count;
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, NHP_EVENT_ACTIONS);
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

static __always_inline bool event_budget_available(__u8 action, __u64 now) {
    __u32 slot = action;
    if (slot >= NHP_EVENT_ACTIONS)
        return true;

    struct nhp_rate_state *st = bpf_map_lookup_elem(&nhp_event_rate, &slot);
    if (!st)
        return true;

    if (now - st->window_start >= NHP_EVENT_WINDOW_NS) {
        st->window_start = now;
        st->count = 1;
        return true;
    }
    if (st->count >= NHP_EVENT_BURST)
        return false;

    st->count++;
    return true;
}

/* Count the packet against its action, and put an event on the perf ring if
 * the class is worth a line and has budget left.
 *
 * `report` is what keeps the log readable. A verdict is worth a line when it
 * says something the previous lines did not: a knock admitted, an SSH attempt
 * from an address that is not the relay, a scan. It is worth nothing when it is
 * the four-thousandth data segment of a connection whose one real decision --
 * "the host opened this flow" -- was already logged, or was never a decision at
 * all because the kernel's socket table answered it. Those bulk classes
 * (TCP_ESTABLISHED, UDP_ESTABLISHED, and every SSH_RELAY packet after the SYN)
 * are passed report=false: they scale with *bytes transferred*, not with
 * events, so one `dnf update` or one CI scp of the release tarball would
 * otherwise write more log than a month of knocks. They are still counted, and
 * the per-window summary reports them.
 *
 * The verdict never depends on this argument -- only the telemetry does. */
static __always_inline void record_packet(void *ctx, __u8 action, struct iphdr *iph,
                                          __be16 src_port, __be16 dst_port,
                                          __u8 relay_hit, bool report) {
    __u32 slot = action;
    struct nhp_action_stat *stat = NULL;

    if (slot < NHP_EVENT_ACTIONS)
        stat = bpf_map_lookup_elem(&nhp_action_stats, &slot);
    if (stat)
        stat->packets++;

    if (!report)
        return;

    __u64 now = bpf_ktime_get_ns();
    if (!event_budget_available(action, now))
        return;

    struct nhp_event_t ev = {};
    ev.timestamp = now;
    ev.action = action;
    ev.src_ip = iph->saddr;
    ev.dst_ip = iph->daddr;
    ev.src_port = src_port;
    ev.dst_port = dst_port;
    ev.protocol = iph->protocol;
    ev.pkt_len = iph->tot_len;
    ev.relay_hit = relay_hit;

    if (bpf_perf_event_output(ctx, &nhp_events, BPF_F_CURRENT_CPU, &ev, sizeof(ev)) == 0 && stat)
        stat->logged++;
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

SEC("xdp")
int xdp_server_prog(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_DROP;

    switch (bpf_ntohs(eth->h_proto)) {
        /* ARP and IPv6 pass unreported: ARP keeps the host on its subnet, and
         * the demo hosts have no IPv6 service to filter -- reporting either
         * would drown the perf ring in neighbour traffic. */
        case ETH_P_ARP:  return XDP_PASS;
        case ETH_P_IPV6: return XDP_PASS;
        case ETH_P_IP:   break;
        default:         return XDP_DROP;
    }

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return XDP_DROP;
    if (iph->ihl < 5)
        return XDP_DROP;

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
             * has_local_flow() cannot let it.
             *
             * This is the outage that made the filter look like it was
             * "closing the whole host an hour after every deploy". EC2 hands
             * out the primary private IPv4 by DHCP on a finite lease;
             * systemd-networkd renews it at half the lease and drops the
             * address (and the default route with it) when a renewal never
             * completes. The renewal answer is a datagram from udp/67 to
             * udp/68, and the client receives it either on a raw AF_PACKET
             * socket or on a UDP socket that is *bound* to port 68 and never
             * connected -- so bpf_sk_lookup_udp() either finds nothing at all
             * or finds something indistinguishable from a listener, and
             * has_local_flow() correctly refuses both. XDP_DROP here is
             * therefore terminal for the lease: nothing later in the stack
             * ever sees the packet, and once the lease expires the host is
             * dark on every port, SSH and the knock port alike, with no way
             * back in short of a reboot.
             *
             * So DHCP replies are admitted by port pair, the same way the AC's
             * program has always done it (`DHCP_PORT_R || DHCP_PORT_O` in
             * nhp/ebpf/xdp/nhp_ebpf_xdp.c). The pair is narrow on purpose: a
             * neighbour's broadcast DHCPDISCOVER is addressed to udp/67 and
             * still dropped, and the only thing this lets an attacker reach is
             * the host's DHCP client with a forged lease -- which is bounded
             * by that client's own xid/server checks, and is the same exposure
             * every unfiltered host on the subnet has.
             *
             * Reported per packet: a lease renewal is a handful of datagrams
             * an hour, and after the outage above it is precisely the line an
             * operator wants to see in the log. */
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
