---
layout: page
title: Message Types
parent: Protocol Reference
nav_order: 2
permalink: /protocol/messages/
description: Complete reference of the 20 NHP message types, their direction, and their payload shape.
---

# NHP Message Types
{: .fs-9 }

Every NHP packet carries a 2-byte Message Type ID in the header. That ID
routes the packet to a logical handler on the receiver. Twenty types are
defined today, covering the knock/auth/access flow, AC lifecycle, relay,
logging, registration, and OTP.
{: .fs-6 .fw-300 }

*Implements CSA Stealth Mode SDP §NHP Message Types (Table 4) and Appendix 2. IDs defined in [`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go); message structs in [`nhp/common/message.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/message.go).*

## Message envelope

Every message type uses the same header (see [Message Header]({{ '/protocol/header/' | relative_url }})) followed by an encrypted payload. Payloads are JSON, zlib-compressed, then AEAD-encrypted using the session key derived during header parsing — unless the type description says otherwise (e.g., `NHP-KPL` has no body; `NHP-RLY` carries a raw inner packet).

## Index

| ID | Name | Direction | Purpose |
|---:|---|---|---|
| 0 | [NHP-KPL](#nhp-kpl--keepalive) | Agent ↔ Server, AC ↔ Server | Keepalive. Empty body. |
| 1 | [NHP-KNK](#nhp-knk--knock) | Agent → Server | Knock: request access to a Protected Resource. |
| 2 | [NHP-ACK](#nhp-ack--acknowledge) | Server → Agent | Response to a knock. Carries resource address and session parameters on success. |
| 3 | [NHP-AOP](#nhp-aop--ac-operation) | Server → AC | Instruct the AC to open (or deny) access for a given agent → resource flow. |
| 4 | [NHP-ART](#nhp-art--ac-result) | AC → Server | AC response to `NHP-AOP`. |
| 5 | [NHP-LST](#nhp-lst--list) | Agent → Server | Request the list of services an agent is entitled to. |
| 6 | [NHP-LRT](#nhp-lrt--list-result) | Server → Agent | Response to `NHP-LST`. |
| 7 | [NHP-COK](#nhp-cok--cookie) | Server → Agent | Rate-limit / DDoS-defence cookie issued when the server is overloaded. |
| 8 | [NHP-RKN](#nhp-rkn--re-knock) | Agent → Server | Second knock, HMAC'd with the cookie from `NHP-COK`. |
| 9 | [NHP-RLY](#nhp-rly--relay) | Relay → Server | Forward a raw NHP packet from an NHP-Relay, preserving source address. |
| 10 | [NHP-AOL](#nhp-aol--ac-online) | AC → Server | AC announces its online status and the resources it protects. |
| 11 | [NHP-AAK](#nhp-aak--ac-acknowledge) | Server → AC | Server confirms the AC's registration. Carries the AC's public IP/port. |
| 12 | [NHP-OTP](#nhp-otp--one-time-password) | Agent ↔ Server / KGC | Request / verify an out-of-band OTP used during registration. |
| 13 | [NHP-REG](#nhp-reg--register) | Agent → Server | Agent registers its static public key. |
| 14 | [NHP-RAK](#nhp-rak--register-acknowledge) | Server / KGC → Agent | Confirmation of successful registration. Empty body. |
| 15 | [NHP-ACC](#nhp-acc--access) | Agent → AC | Agent presents its temporary access token to the AC's listener. |
| 16 | [NHP-EXT](#nhp-ext--exit) | Agent → Server | Request early closure of an active session. Empty body. |
| 17 | [NHP-LOG](#nhp-log--log) | AC → Server | AC submits an access-log record. |
| 18 | [NHP-LAK](#nhp-lak--log-acknowledge) | Server → AC | Ack for a received log record. |
| 19 | [NHP-ARD](#nhp-ard--ac-redispatch) | Server → AC | Redirect an AC to a different NHP-Server endpoint for load-balanced reassignment. |

---

## NHP-KPL — Keepalive {#nhp-kpl--keepalive}

**ID:** `0` · **Direction:** bidirectional (Agent ↔ Server, AC ↔ Server) · **Body:** empty (header only, all fields after Length set to zero)

Keeps the NAT binding / TCP connection alive between Agent/AC and Server. Receivers do nothing beyond acknowledging receipt. Relay nodes do **not** forward keepalives.

## NHP-KNK — Knock {#nhp-knk--knock}

**ID:** `1` · **Direction:** Agent → Server · **Payload fields:**

| Field | Description |
|---|---|
| User ID | Per-user identifier tying the request to a human or service principal. |
| Device ID | Per-device identifier for multi-device / posture policies. |
| Authorized Service Provider ID | Which ASP (IAM / SDP Controller / custom) the Server should consult. |
| Resource ID | The Protected Resource being requested (domain, service name, or opaque ID). |
| Source IP / Port | Optional. Exit address of the agent, when the agent can discover it. |
| Terminal Environment Parameters | Optional `{checkID: result}` map reflecting client-side posture checks (OS version, disk encryption, etc.). |

The knock is the only message that initiates an NHP exchange from a cold start.

## NHP-ACK — Acknowledge {#nhp-ack--acknowledge}

**ID:** `2` · **Direction:** Server → Agent · **Payload fields:**

| Field | Description |
|---|---|
| User ID / Device ID / Service Provider ID / Resource ID | Echo of the matching NHP-KNK for session correlation. |
| Session ID | 64-bit session tracker. |
| Error Code | `0` on success; nonzero indicates the authoritative failure reason. |
| Resource Address | IP / hostname of the Protected Resource (or the fronting AC). |
| Access Duration | Seconds the open-door window remains valid. |
| Temporary Access Endpoint / Key | Optional — present when NHP-ACC is used to gate a secondary listener. |

## NHP-AOP — AC Operation {#nhp-aop--ac-operation}

**ID:** `3` · **Direction:** Server → AC · **Payload fields:**

| Field | Description |
|---|---|
| Session ID | Matches the agent's session (from NHP-ACK). |
| Device ID | Agent identity, for session attribution. |
| Public Key | Agent's static public key. |
| Source IP / Port | Tuple the AC should allow. |
| Destination IP / Port | Protected-resource tuple. |
| Access Duration | Seconds; `0` means **deny** / close. |

NAT note: behind CGNAT or shared egress the Source IP is not unique per agent. Deployments should layer an application-layer token/cookie between AC and the Protected Resource, or prefer IPv6 end-to-end.

## NHP-ART — AC Result {#nhp-art--ac-result}

**ID:** `4` · **Direction:** AC → Server · **Payload fields:** Session ID, granted Access Duration (0 = denied), optional Temporary Access Endpoint and Key.

Forms the second half of the AOP/ART pair. Only after ART reaches the Server does the Server send NHP-ACK to the Agent.

## NHP-LST — List {#nhp-lst--list}

**ID:** `5` · **Direction:** Agent → Server · **Payload fields:** User ID, Device ID, Request ID, Request Port Information flag, Requested Application Information (UUID or custom string).

## NHP-LRT — List Result {#nhp-lrt--list-result}

**ID:** `6` · **Direction:** Server → Agent · **Payload fields:** List Request ID, Server Name, Server Type, Service and Application Info.

## NHP-COK — Cookie {#nhp-cok--cookie}

**ID:** `7` · **Direction:** Server → Agent · **Payload fields:** 32-byte server-generated cookie.

Issued when the Server is under load. The Agent must re-knock using NHP-RKN and include this cookie in the HMAC to prove a round-trip and survive early-drop.

## NHP-RKN — Re-Knock {#nhp-rkn--re-knock}

**ID:** `8` · **Direction:** Agent → Server · **Payload fields:** identical to NHP-KNK.

Difference from NHP-KNK: the header HMAC is keyed with the NHP-COK cookie in addition to the normal chaining key.

## NHP-RLY — Relay {#nhp-rly--relay}

**ID:** `9` · **Direction:** Relay → Server · **Payload:** the raw inner NHP packet from the origin, uncompressed, un-re-encrypted. The protocol flag's compression bit is `0`.

Preserves the origin source address through the relay, which most other message types do not need (they're forwarded transparently).

## NHP-AOL — AC Online {#nhp-aol--ac-online}

**ID:** `10` · **Direction:** AC → Server · **Payload fields:** AC ID, Service and Application Info (resources the AC protects, with inbound/outbound traffic info).

Announces an AC joining the control plane.

## NHP-AAK — AC Acknowledge {#nhp-aak--ac-acknowledge}

**ID:** `11` · **Direction:** Server → AC · **Payload fields:** NHP-AC public IP and port.

Confirms AC registration and echoes back the AC's public address (useful when the AC is behind NAT and learns its external tuple from the Server).

## NHP-OTP — One-Time Password {#nhp-otp--one-time-password}

**ID:** `12` · **Direction:** Agent ↔ Server ↔ KGC · **Payload fields:** User ID, Device ID.

Pre-registration primitive: triggers the ASP to issue an OTP out-of-band (SMS, email, QR).

## NHP-REG — Register {#nhp-reg--register}

**ID:** `13` · **Direction:** Agent → Server · **Payload fields:** User ID, Device ID, OTP.

Registers the Agent's static public key against its identity, authenticated by the OTP.

## NHP-RAK — Register Acknowledge {#nhp-rak--register-acknowledge}

**ID:** `14` · **Direction:** Server / KGC → Agent · **Body:** empty.

Success marker for NHP-REG. Failure is signalled by silence (and an eventual client timeout).

## NHP-ACC — Access {#nhp-acc--access}

**ID:** `15` · **Direction:** Agent → AC · **Payload fields:** User ID, Device ID, Temporary Access Token (from NHP-ACK).

Presented directly to the AC's short-lived listener when the deployment uses per-session temporary endpoints rather than long-lived allow-lists.

## NHP-EXT — Exit {#nhp-ext--exit}

**ID:** `16` · **Direction:** Agent → Server · **Body:** empty.

Agent explicitly requests early teardown of an active session. The Server then sends an NHP-AOP with `Access Duration = 0`.

## NHP-LOG — Log {#nhp-log--log}

**ID:** `17` · **Direction:** AC → Server · **Payload fields:** AC ID, Log ID (usually a content hash), Log Content.

## NHP-LAK — Log Acknowledge {#nhp-lak--log-acknowledge}

**ID:** `18` · **Direction:** Server → AC · **Payload fields:** Log ID.

Mirrors the AOP/ART pattern for logging reliability.

## NHP-ARD — AC Redispatch {#nhp-ard--ac-redispatch}

**ID:** `19` · **Direction:** Server → AC · **Payload fields:** new NHP-Server address(es).

Sent during `NHP-AOL` handling to steer an AC to a different Server for load-balancing or failover, without exposing topology to outside observers.

---

## See also

- [Message Header]({{ '/protocol/header/' | relative_url }}) — the envelope every type shares
- [Cryptography]({{ '/cryptography/' | relative_url }}) — how the payload is encrypted
- [Glossary]({{ '/glossary/' | relative_url }}) — canonical names for every role that appears here
