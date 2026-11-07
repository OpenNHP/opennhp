---
layout: page
title: AC discovery failure policy
nav_exclude: true
permalink: /ac-failure-policy/
---

# AC discovery failure policy

For the IPv4 iptables AC (`FilterMode = 0`), set in `etc/config.toml`:

```toml
FailurePolicy = "fail-closed"
```

For remote etcd TOML configuration, place this setting under `[BaseConfig]`.

Omitted, empty, or `"fail-open"` preserves existing behavior: when all server
discoveries fail, the AC inserts broad INPUT and FORWARD ACCEPT rules.
`"fail-closed"` prevents that fallback. Existing grants keep their normal lifetime;
established connections may survive grant expiry. Discovery/reporting intervals
and grant lifetimes are unchanged. IPv6 and XDP are unchanged;
selecting fail-closed with XDP is rejected.

The setting requires a restart. Invalid values fail startup. During reload, a
valid policy change is logged and deferred until restart; the running policy and
filter mode remain active while live settings such as `LogLevel` and `DefaultIp`
are applied. An invalid policy rejects the base-config update. Independent HTTP
and server-peer sections of an etcd reload are still applied.

Local startup now reports a missing or malformed `etc/config.toml` immediately,
instead of continuing with an empty configuration and failing later.

The AC requires the iptables `comment` match (`xt_comment` on legacy iptables)
to tag its broad fallback rules with `opennhp-ac-fail-open`. Ensure this support
is available in the host kernel and iptables installation, including for container
deployments. Fail-open startup verifies support by inserting a comment-bearing
RETURN rule in a temporary, unreferenced chain, then removing that chain. Failure
of the probe or its cleanup is logged in fail-open mode, and startup continues.
Discovery retries failed rule changes; fallback cannot open until comment support
is available. The probe does not open traffic. Fail-closed skips the insertion
probe and still refuses startup if required cleanup or inspection fails.

On startup, both modes attempt to remove those owned rules, including duplicates
left after a crash.
Switching from fail-open to fail-closed therefore needs no manual rule cleanup
when the rules were created by this version. After successful cleanup, fail-open
starts with the normal firewall baseline and opens again only after the usual
discovery failure threshold.
Failed firewall commands are reported and retried on subsequent discovery ticks;
fail-closed refuses startup if owned-rule cleanup fails. Fail-open logs the failure
and continues startup, so leftover owned rules may remain until cleanup succeeds.

Older versions created untagged unconditional ACCEPT rules. They cannot safely be
distinguished from operator rules. Before upgrading, stop the old AC gracefully
and inspect the firewall. Fail-closed startup refuses recognized untagged fallback
rules anywhere in INPUT or FORWARD, including intentional operator rules, not just
rules at the first position. Inspect and remove confirmed AC leftovers before
restarting. If an unconditional ACCEPT is intentional, give it a distinct operator
comment (for example `-m comment --comment operator-baseline -j ACCEPT`), preserving
its position and other conditions. Never use the reserved `opennhp-ac-fail-open`
comment for operator rules. Fail-closed then leaves the labeled rule in place;
traffic allowed by that rule remains allowed. This is not a baseline firewall audit.
**Refusal leaves the existing firewall in effect, which may still allow traffic.**
Fail-open does not remove these untagged rules. Neither mode audits arbitrary rules
or installs the baseline; operators must supply the normal iptables/ipset baseline.
Only one AC may own fallback rules in a network namespace. External firewall rewrites
are outside this feature's scope.

With no configured servers, fail-open retains the existing behavior of opening
access on the next discovery check; fail-closed does not insert fallback rules.

The existing Go methods `AcceptAllInput()` and `ResetAllInput()` retain their
no-return signatures for interface and callback compatibility and log failures.
The AC uses `SetFailurePolicyOpen(bool) error` to handle failures and retry.
The IPv6 methods retain their existing behavior; this setting only controls IPv4.
