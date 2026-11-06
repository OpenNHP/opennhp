# AC discovery failure policy

For the IPv4 iptables AC (`FilterMode = 0`), set in `etc/config.toml`
(or `BaseConfig.failurePolicy` in remote JSON configuration):

```toml
FailurePolicy = "fail-closed"
```

Omitted, empty, or `"fail-open"` preserves existing behavior: when all server
discoveries fail, the AC inserts broad INPUT and FORWARD ACCEPT rules.
`"fail-closed"` prevents that fallback. Existing grants keep their normal lifetime;
established connections may survive grant expiry. Discovery/reporting intervals
and grant lifetimes are unchanged. IPv6 and XDP are unchanged;
selecting fail-closed with XDP is rejected.

The setting requires a restart. Invalid values fail startup. A live policy change
rejects the base-config update and logs that a restart is required.

The AC tags its broad fallback rules with `opennhp-ac-fail-open`. On startup,
both modes remove those owned rules, including duplicates left after a crash.
Switching from fail-open to fail-closed therefore needs no manual rule cleanup
when the rules were created by this version. Fail-open starts with the normal
firewall baseline and opens again only after the usual discovery failure threshold.
Failed firewall commands are reported and retried on subsequent discovery ticks;
startup fails if owned-rule cleanup fails.

Older versions created untagged unconditional ACCEPT rules. They cannot safely be
distinguished from operator rules. Before upgrading, stop the old AC gracefully
and inspect the firewall. Fail-closed startup refuses recognized untagged fallback
rules in INPUT or FORWARD; inspect and remove confirmed leftovers before restarting.
**Refusal leaves the existing firewall in effect, which may still allow traffic.**
Fail-open does not remove these untagged rules. Neither mode audits arbitrary rules
or installs the baseline; operators must supply the normal iptables/ipset baseline.
Only one AC may own fallback rules in a network namespace. External firewall rewrites
are outside this feature's scope.
