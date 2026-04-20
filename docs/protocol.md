---
layout: page
title: Protocol Reference
nav_order: 11
has_children: true
permalink: /protocol/
description: OpenNHP wire-format reference. Maps the CSA NHP specification to the Go implementation.
---

# Protocol Reference
{: .fs-9 }

Wire-format reference for the Network-infrastructure Hiding Protocol as
implemented by OpenNHP. If you are building a new client, porting NHP to
another language, or writing tooling that needs to decode NHP packets, this
is the authoritative starting point.
{: .fs-6 .fw-300 }

*Implements the CSA "Stealth Mode SDP for Zero Trust Network Infrastructure" whitepaper — §NHP Message Header, §NHP Message Types, and Appendix 2.*

## What's in this section

- **[Message Header]({{ '/protocol/header/' | relative_url }})** — the 240/304-byte fixed header that prefixes every NHP or DHP packet. Field layout, obfuscation scheme, cryptographic role of each field.
- **[Message Types]({{ '/protocol/messages/' | relative_url }})** — all 17 NHP message types plus the 11 DHP types, each with the role of its sender/receiver, payload fields, and pointers into the code.

## Spec ↔ implementation

| CSA spec section | OpenNHP code |
|---|---|
| NHP Message Header (Table 3) | [`nhp/core/scheme/curve/header.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/scheme/curve/header.go), [`nhp/core/scheme/gmsm/header.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/scheme/gmsm/header.go), [`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go) |
| NHP Message Types (Table 4, Appendix 2) | [`nhp/common/nhpmsg.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/nhpmsg.go) |
| Cryptographic Algorithms and Frameworks | [`nhp/core/crypto.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/crypto.go), [`nhp/core/device.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/device.go) |
| NHP Workflow (Figure 3) | [`endpoints/agent/`](https://github.com/OpenNHP/opennhp/tree/main/endpoints/agent), [`endpoints/server/`](https://github.com/OpenNHP/opennhp/tree/main/endpoints/server), [`endpoints/ac/`](https://github.com/OpenNHP/opennhp/tree/main/endpoints/ac) |
| Integration with SDP | See [Deploy]({{ '/deploy/' | relative_url }}) |
| Logging | [`nhp/log/`](https://github.com/OpenNHP/opennhp/tree/main/nhp/log) |

## Spec version

OpenNHP follows the CSA NHP whitepaper published by the CSA Zero Trust Working
Group. The whitepaper cites OpenNHP **v0.6.0** as the reference implementation.
When the spec and code disagree, the code wins in this repository and we raise
the discrepancy with the Working Group.

## Scope

The Protocol Reference covers the wire format:

- Header byte layout
- Encryption and authentication envelope (Noise + AEAD)
- Obfuscation of public header fields
- Message type IDs and their payload schemas
- Session-ID, counter, and replay-protection semantics

It **does not** cover:

- Config-file formats → see [How to Deploy]({{ '/deploy/' | relative_url }})
- Component-level internals → see [Understand the Code]({{ '/code/' | relative_url }})
- Plugin interfaces → see [Server Plugins]({{ '/server_plugin/' | relative_url }}) and [Client SDKs]({{ '/agent_sdk/' | relative_url }})

## Terminology

This section uses the canonical terms defined in the [Glossary]({{ '/glossary/' | relative_url }}). The spec occasionally uses SDP-flavored synonyms (e.g., "gateway" for NHP-AC); we stick to the NHP names here to match the Go code.
