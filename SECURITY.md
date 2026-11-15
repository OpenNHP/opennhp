# Security Policy

## Reporting Security Issues

The OpenNHP team and community take security bugs in OpenNHP seriously. We appreciate your efforts to responsibly disclose your findings, and will make every effort to acknowledge your contributions.

To report a security issue, please use the GitHub Security Advisory ["Report a Vulnerability"](https://github.com/opennhp/opennhp/security/advisories/new) tab (preferred, as it keeps the report private), or email the security contact at [support@opennhp.org](mailto:support@opennhp.org).

The OpenNHP team will send a response indicating the next steps in handling your report. After the initial reply to your report, the security team will keep you informed of the progress towards a fix and full announcement, and may ask for additional information or guidance.

Report security bugs in third-party modules to the person or team maintaining the module. 

## Known limitations

### The firewall grant is bound to a source address, not to an identity

The server hands the access controller a firewall rule derived from the source
address of the authenticated knock, and the access controller matches on that
address alone. The identity the knock authenticated (`user`, `device`,
`organization`) is not part of the rule.

Two consequences follow, and neither is closed by the packet-level replay
protection described below:

- A host that shares a NAT or CGNAT egress address with a legitimate agent is
  indistinguishable from that agent to the access controller. It needs no
  key material and no captured traffic.
- The server's source-IP association map can widen a grant to other addresses
  associated with the same source address.

Binding the grant to the authenticated identity — carrying a verifiable
credential from the server to the access controller and checking it in the data
plane — is a separate change with a different blast radius.

### Packet-level replay protection is per server process

The server refuses a byte-identical repeat of a knock it has already accepted,
and refuses a knock older than the configured freshness window (60 seconds by
default). That protects the packet.

The cache is process-local: it clears on restart, and in a multi-replica
deployment it only refuses a replay that lands on the same instance that
accepted the original. An agent's knocks must be sticky to one instance for
the protection to hold. See [docs/deploy.md](docs/deploy.md) ("KNK replay
protection") for the configuration and the metrics to alert on.

