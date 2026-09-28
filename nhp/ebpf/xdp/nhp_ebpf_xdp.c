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
 * one whose knock shape reopens that hole. */
static __always_inline int admit_ingress(void *ctx, struct iphdr *iph,
                                         struct ipv4_ct_tuple *ct_key, __u64 now,
                                         bool track, __u64 expire_time) {
    submit_event(ctx, 1, iph->saddr, iph->daddr, ct_key->sport, ct_key->dport, iph->protocol, iph->tot_len);
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
    bool is_tcp = false, is_syn = false, is_fin = false, is_rst = false;
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

    /* Has this peer ever been admitted to this port of ours by a knock? Unlike
     * `peer_has_live_knock` this survives the knock, and it is what keeps the
     * reverse branch below from serving a peer whose knock has expired. Only
     * non-TCP: the TCP egress gate is exact, so an EGRESS entry for a TCP flow
     * really is a connection the AC opened, and `knock_peers` holds nothing for
     * TCP anyway. Read here, before reverseTuple(), while ct_key.dport is still
     * the port on us. */
    bool peer_knocked_port = iph->protocol != IPPROTO_TCP &&
                             nhp_peer_knocked(iph->saddr, ct_key.dport, iph->protocol);

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
            return XDP_PASS;
        }
    }

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
     * A peer that holds a live "any"-protocol knock is never served from here
     * either, nor -- for non-TCP -- is one that `knock_peers` says has been
     * admitted to this port of ours at any point. Gate 4 in tc_egress.c already
     * refuses to create an EGRESS entry for such a peer, and this is the other
     * half of that: an entry that pre-dates the knock, or that the gates
     * missed, must not become a second way in. Nothing is lost by skipping the
     * branch -- the whitelist lookups below admit the packet if a knock really
     * is live, and they are the only thing that may.
     */
    reverseTuple(&ct_key);
    existing_val = bpf_map_lookup_elem(&conn_track, &ct_key);
    int reverse_verdict = -1;
    if (existing_val && !is_syn && !peer_has_live_knock && !peer_knocked_port) {
        if (check_conn_expiry(existing_val)) {
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
     * See new_ingress_conn(). */
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
            return XDP_DROP;
        }
        if (w_val->allowed == 1) {
            return admit_ingress(ctx, iph, &ct_key, now, track_new_conn, expire_time);
        }
    }
    if (sd_val) {
        __u64 expire_time = sd_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&sdwhitelist, &sdkey);
            return XDP_DROP;
        }
        if (sd_val->allowed == 1) {
            return admit_ingress(ctx, iph, &ct_key, now, track_new_conn, expire_time);
        }
    }

    if (sp_val) {
        __u64 expire_time = sp_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&src_port, &spkey);
            return XDP_DROP;
        }
        if (sp_val->allowed == 1) {
            return admit_ingress(ctx, iph, &ct_key, now, track_new_conn, expire_time);
        }
    }

    if (pl_val) {
        __u64 expire_time = pl_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&port_list, &pl_key);
            return XDP_DROP;
        }
        if (pl_val->allowed == 1) {
            return admit_ingress(ctx, iph, &ct_key, now, track_new_conn, expire_time);
        }
    }
    if (pp_val) {
        __u64 expire_time = pp_val->expire_time;
        if (expire_time < now) {
            bpf_map_delete_elem(&protocol_port, &pp_key);
            return XDP_DROP;
        }
        if (pp_val->allowed == 1) {
            return admit_ingress(ctx, iph, &ct_key, now, track_new_conn, expire_time);
        }
    }
    submit_event(ctx, 0, iph->saddr, iph->daddr, ct_key.sport, ct_key.dport, iph->protocol, iph->tot_len);
    return XDP_DROP;
}

char _license[] SEC("license") = "Dual BSD/GPL";