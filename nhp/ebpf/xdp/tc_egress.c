#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_endian.h>

#define TC_ACT_UNSPEC         (-1)
#define TC_ACT_OK               0
#define TC_ACT_SHOT             2
#define TC_ACT_STOLEN           4

#define ETH_P_IP    0x0800
#define ETH_P_IPV6  0x86DD
#define IPPROTO_TCP 6
#define IPPROTO_UDP 17
#define IPPROTO_ICMP 1
#define MAX_ENTRIES 1000000

// tc_egress only handles egress traffic (server -> client responses).
// It does NOT add entries to any whitelist map to avoid refreshing
// the whitelist TTL and causing access to never expire.
// Established connections are tracked by the conn_track map in the
// XDP program, so egress traffic for established connections is
// automatically allowed through.

SEC("tc/egress")
int tc_egress_prog(struct __sk_buff *ctx)
{
    // Egress traffic (server responses) for established connections
    // is automatically allowed by the conn_track map in the XDP program.
    // This tc_egress program simply passes all egress traffic.
    // No whitelist updates here - doing so would refresh TTL and
    // cause access to never expire (the 180s extension was a bug).
    return TC_ACT_OK;
}


char _license[] SEC("license") = "Dual BSD/GPL";