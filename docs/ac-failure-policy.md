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

Omitted, empty, or `"fail-open"` selects the availability policy: when all server
discoveries fail, the AC opens IPv4 INPUT and FORWARD traffic through its reserved fallback chain.
`"fail-closed"` prevents that fallback. Existing grants keep their normal lifetime;
established connections may survive grant expiry. Discovery/reporting intervals
and grant lifetimes are unchanged. IPv6 and XDP are unchanged;
selecting fail-closed with XDP is rejected.

The setting requires a restart. Invalid values fail startup. During reload, a
valid policy change is logged and deferred until restart; the running policy and
filter mode remain active while live settings such as `LogLevel` and `DefaultIp`
are applied. An invalid policy rejects the base-config update. Independent HTTP
and server-peer sections of an etcd reload are still applied. Failure observations
for unchanged endpoints survive reload, so an unrelated edit cannot close an
active fail-open fallback during an outage; a successful discovery clears that
endpoint's failed state.

Local startup now reports a missing or malformed `etc/config.toml` immediately,
instead of continuing with an empty configuration and failing later.

This feature requires iptables 1.4.20 or newer for the bare `-w` lock option.
Earlier binaries are unsupported: they reject the option, so fallback and cleanup
cannot work. Check `iptables --version` before upgrading.

The chain name `NHP-AC-FALLBACK` is reserved for the AC. When fallback is open,
it contains an unconditional ACCEPT rule, with unconditional jumps from INPUT
and FORWARD. This uses built-in iptables features and does not require the
`comment` match. Do not add operator rules or references to this chain.
The AC inspects the chain and all its references before modifying it. Unfamiliar
contents or references cause an error; the AC never flushes the chain or deletes
chains by prefix.

On startup, both modes remove owned jumps and the fallback chain, including
partial updates and duplicates left after a crash. Exact `opennhp-ac-fail-open`
comment-tagged rules from earlier drafts are also removed. Switching from
fail-open to fail-closed therefore needs no manual cleanup of these owned rules.
After successful cleanup, fail-open starts with the normal firewall baseline and
opens again only after the usual discovery failure threshold.

Failed firewall commands are reported and retried on subsequent discovery ticks.
Fail-closed refuses startup if required cleanup or inspection fails. Fail-open
logs the failure and continues startup, so leftover owned rules may remain until
cleanup succeeds. If fallback cannot open during an outage, the AC logs a critical
error when the failure begins; retries do not repeat that alert on every tick.

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
Fail-open warns with the exact untagged rule text and leaves it unchanged, so
traffic it permits remains allowed even after discovery recovers. Neither mode
audits arbitrary rules
or installs the baseline; operators must supply the normal iptables/ipset baseline.
Only one AC may own fallback rules in a network namespace. External firewall rewrites
are outside this feature's scope.

If the running peer map becomes empty, fail-open retains the existing behavior of
opening access on the next discovery check; fail-closed does not insert fallback
rules. This does not bypass validation of the server configuration at startup.

The existing Go methods `AcceptAllInput()` and `ResetAllInput()` retain their
no-return signatures for interface and callback compatibility and log failures.
The AC uses `SetFailurePolicyOpen(bool) error` to handle failures and retry.
The IPv6 methods retain their existing behavior; this setting only controls IPv4.

Reload and shutdown wait for an in-flight firewall update. Each command has a
five-second timeout; this is not a limit on the whole update. A normal change
uses up to five commands, and stale duplicates require additional deletions.
Shutdown also attempts a final cleanup. Allow for these operations in service
stop timeouts. For systemd deployments, check the AC unit's `TimeoutStopSec`
with `systemctl cat <ac-unit>` and size it for cleanup under lock contention.
An interrupted update followed by final cleanup can take tens of seconds;
duplicates and DNS resolution can add more. There is no universal safe fixed
value. If systemd kills the process before cleanup finishes, fail-open rules may
remain until the next successful AC startup. Do not disable the stop deadline
as a substitute for investigating repeated cleanup failures.

An initial hostname-resolution failure marks that endpoint failed and is retried
without requiring a configuration reload. A successful lookup does not clear the
failure until discovery receives a successful response. Hostname resolution uses
the existing system resolver. An in-progress DNS lookup
can also delay discovery replacement or shutdown.
