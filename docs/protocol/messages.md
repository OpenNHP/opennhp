---
layout: page
title: Message Types
parent: Protocol Reference
nav_order: 2
permalink: /protocol/messages/
description: Complete reference of the NHP and DHP message types, their direction, and their payload shape.
---

# NHP Message Types
{: .fs-9 }

Every packet carries a 2-byte Message Type ID in the header. That ID
routes the packet to a logical handler on the receiver. Twenty-eight IDs
are defined today: seventeen **NHP** types (IDs 0–16) covering the
knock/auth/access flow, AC lifecycle, relay, registration, OTP, and
explicit session exit; plus eleven **DHP** types (IDs 17–27) used by the
Data-object Hiding Protocol.
{: .fs-6 .fw-300 }

*Implements CSA Stealth Mode SDP §NHP Message Types (Table 4) and Appendix 2. IDs and string names defined in [`nhp/core/packet.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/packet.go); payload structs in [`nhp/common/nhpmsg.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/nhpmsg.go).*

## Message envelope

Every message type uses the same header (see [Message Header]({{ '/protocol/header/' | relative_url }})) followed by an encrypted payload. Payloads are JSON, optionally zlib-compressed (see the `NHP_FLAG_COMPRESS` flag), then AEAD-encrypted using the session key derived during header parsing — unless the type description says otherwise (e.g., `NHP-KPL` has no body; `NHP-RLY` carries a raw inner packet).

## Index

NHP — Network-infrastructure Hiding Protocol

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
| 12 | [NHP-OTP](#nhp-otp--one-time-password) | Agent → Server | Request an out-of-band OTP for registration. |
| 13 | [NHP-REG](#nhp-reg--register) | Agent → Server | Agent registers its static public key, authenticated by the OTP. |
| 14 | [NHP-RAK](#nhp-rak--register-acknowledge) | Server → Agent | Confirmation (or failure code) for `NHP-REG`. |
| 15 | [NHP-ACC](#nhp-acc--access) | Agent → AC | Agent presents its temporary access token to the AC's listener. |
| 16 | [NHP-EXT](#nhp-ext--exit) | Agent → Server | Request early closure of an active session. Empty body. |

DHP — Data-object Hiding Protocol *(documented here for completeness; detailed DHP semantics live with the DHP docs)*

| ID | Constant | Direction | Purpose |
|---:|---|---|---|
| 17 | `NHP_DRG` | DB → Server | Register a data object with the NHP-Server. |
| 18 | `NHP_DAK` | Server → DB | Acknowledgement / error code for `NHP_DRG`. |
| 19 | `NHP_DAR` | DHP-Agent → Server | Request access to a registered data object. |
| 20 | `NHP_DAG` | Server → DHP-Agent | Authorization decision for `NHP_DAR`. |
| 21 | `NHP_DSA` | Server → DHP-Agent | Request a self-attestation (TEE / evidence). |
| 22 | `NHP_DAV` | DHP-Agent → Server | Attestation proof returned to the server. |
| 23 | `NHP_DWR` | Server → DB | Ask the DB for the wrapped data-object private key. |
| 24 | `NHP_DWA` | DB → Server | DB returns the wrapped key (Key Access Object). |
| 25 | `NHP_DOL` | DB → Server | DB online / status announcement. |
| 26 | `NHP_DBA` | Server → DB | Acknowledgement of `NHP_DOL`. |
| 27 | `DHP_KNK` | DHP-Agent → Server | DHP-flavoured knock carrying attestation evidence. |

---

## NHP-KPL — Keepalive {#nhp-kpl--keepalive}

**ID:** `0` · **Direction:** bidirectional (Agent ↔ Server, AC ↔ Server) · **Body:** empty (header only)

Keeps the NAT binding / TCP connection alive between Agent/AC and Server. Receivers do nothing beyond accepting the packet. Relay nodes do **not** forward keepalives.

## NHP-KNK — Knock {#nhp-knk--knock}

**ID:** `1` · **Direction:** Agent → Server · **Payload struct:** `common.AgentKnockMsg`

| JSON key | Field | Description |
|---|---|---|
| `headerType` | Header Type | Echo of the on-wire message type. |
| `usrId` | User ID | Per-user identifier tying the request to a human or service principal. |
| `devId` | Device ID | Per-device identifier for multi-device / posture policies. |
| `orgId` | Organization ID | Optional tenant scope. |
| `aspId` | ASP ID | Which Authorization Service Provider the Server should consult. |
| `resId` | Resource ID | The Protected Resource being requested (domain, service name, or opaque ID). |
| `results` | Check Results | Optional `{checkID: result}` map reflecting client-side posture checks. |
| `usrData` | User Data | Optional free-form data passed through to the ASP plugin. |

The knock is the only message that initiates an NHP exchange from a cold start.

## NHP-ACK — Acknowledge {#nhp-ack--acknowledge}

**ID:** `2` · **Direction:** Server → Agent · **Payload struct:** `common.ServerKnockAckMsg`

| JSON key | Field | Description |
|---|---|---|
| `errCode` / `errMsg` | Error code / message | `"0"` (or empty) on success; any other value indicates the failure reason. |
| `resHost` | Resource Hosts | Map of resource name → host address. |
| `opnTime` | Open Time | Seconds the open-door window remains valid. |
| `aspToken` | ASP Token | Optional token for AC-side backend validation. |
| `agentAddr` | Agent Address | The agent's public tuple as the server saw it. |
| `acTokens` | AC Tokens | Map of resource name → short-lived token used with `NHP-ACC`. |
| `preActions` | Pre-access Actions | Optional per-resource `PreAccessInfo` (AC IP, port, pubkey, token, cipher scheme) used when the deployment requires an `NHP-ACC` step. |
| `redirectUrl` | Redirect URL | Optional HTTPS redirect (e.g., login flow) instead of direct connect. |

## NHP-AOP — AC Operation {#nhp-aop--ac-operation}

**ID:** `3` · **Direction:** Server → AC · **Payload struct:** `common.ServerACOpsMsg`

| JSON key | Field | Description |
|---|---|---|
| `usrId` / `devId` / `orgId` | Agent identity | Attribution fields. |
| `aspId` / `resId` | ASP + Resource | Correlates to the original knock. |
| `srcAddrs` | Source addresses | List of `NetAddress` the AC should allow. |
| `dstAddrs` | Destination addresses | Protected-resource tuples. |
| `opnTime` | Open Time | Seconds; `0` means **deny** / close. |

NAT note: behind CGNAT or shared egress the Source IP is not unique per agent. Deployments should layer an application-layer token/cookie between AC and the Protected Resource (the `NHP-ACC` path does exactly this), or prefer IPv6 end-to-end.

## NHP-ART — AC Result {#nhp-art--ac-result}

**ID:** `4` · **Direction:** AC → Server · **Payload struct:** `common.ACOpsResultMsg`

Carries `errCode` / `errMsg`, the granted `opnTime` (0 = denied), the AC-issued `token`, and an optional `preAct` (`PreAccessInfo`). Only after ART reaches the Server does the Server send NHP-ACK to the Agent.

## NHP-LST — List {#nhp-lst--list}

**ID:** `5` · **Direction:** Agent → Server · **Payload struct:** `common.AgentListMsg`

Carries `usrId`, `devId`, optional `orgId`, `aspId`, and free-form `usrData`.

## NHP-LRT — List Result {#nhp-lrt--list-result}

**ID:** `6` · **Direction:** Server → Agent · **Payload struct:** `common.ServerListResultMsg`

Carries `errCode` / `errMsg` and a `list` map whose shape is defined by the ASP plugin.

## NHP-COK — Cookie {#nhp-cok--cookie}

**ID:** `7` · **Direction:** Server → Agent · **Payload struct:** `common.ServerCookieMsg`

Carries the server-echoed `trxId` and a server-generated `cookie`. Issued when the Server is under load. The Agent must re-knock using NHP-RKN and include this cookie in the HMAC calculation to prove a round-trip and survive early-drop.

## NHP-RKN — Re-Knock {#nhp-rkn--re-knock}

**ID:** `8` · **Direction:** Agent → Server · **Payload:** identical to NHP-KNK.

Difference from NHP-KNK: the header HMAC is keyed with the NHP-COK cookie in addition to the normal chaining key (see `addHMAC(sumCookie: true)` in [`nhp/core/initiator.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/core/initiator.go)).

## NHP-RLY — Relay {#nhp-rly--relay}

**ID:** `9` · **Direction:** Relay → Server · **Payload:** the raw inner NHP packet from the origin.

Preserves the origin source address through the relay, which most other message types do not need (they are forwarded transparently).

## NHP-AOL — AC Online {#nhp-aol--ac-online}

**ID:** `10` · **Direction:** AC → Server · **Payload struct:** `common.ACOnlineMsg`

Carries `aspId`, the list of `resIds` the AC protects, and an optional `acId`. Announces an AC joining the control plane.

## NHP-AAK — AC Acknowledge {#nhp-aak--ac-acknowledge}

**ID:** `11` · **Direction:** Server → AC · **Payload struct:** `common.ServerACAckMsg`

Carries `errCode` / `errMsg` and `acAddr`. Confirms AC registration and echoes back the AC's public address (useful when the AC is behind NAT and learns its external tuple from the Server).

## NHP-OTP — One-Time Password {#nhp-otp--one-time-password}

**ID:** `12` · **Direction:** Agent → Server · **Payload struct:** `common.AgentOTPMsg`

Carries `usrId`, `devId`, optional `orgId`, `aspId`, an optional `pass` (pre-shared passcode), and free-form `usrData`. Triggers the ASP plugin to issue an OTP out-of-band (SMS, email, QR). The server does not reply with a dedicated message type; success is signalled by the ASP delivering the OTP through its own channel.

## NHP-REG — Register {#nhp-reg--register}

**ID:** `13` · **Direction:** Agent → Server · **Payload struct:** `common.AgentRegisterMsg`

Carries `usrId`, `devId`, optional `orgId`, `aspId`, the `otp` received out-of-band, and free-form `usrData`. Registers the Agent's static public key (carried in the encrypted header) against its identity.

## NHP-RAK — Register Acknowledge {#nhp-rak--register-acknowledge}

**ID:** `14` · **Direction:** Server → Agent · **Payload struct:** `common.ServerRegisterAckMsg`

Carries `errCode` (`"0"` on success), `errMsg`, and `aspId`. Unlike some NHP flows, failures are **reported explicitly** with a non-zero `errCode` rather than by silent drop — see [`HandleRegisterRequest`](https://github.com/OpenNHP/opennhp/blob/main/endpoints/server/msghandler.go).

## NHP-ACC — Access {#nhp-acc--access}

**ID:** `15` · **Direction:** Agent → AC · **Payload struct:** `common.AgentAccessMsg`

Carries `usrId`, `devId`, optional `orgId`, `acToken` (from NHP-ACK's `acTokens`), and `usrData`. Presented directly to the AC's short-lived listener when the deployment uses per-session temporary endpoints (`PreAccessInfo`) rather than long-lived allow-lists. The AC replies with a `common.ACAccessAckMsg` (`errCode`, `errMsg`, `agentAddr`) carried as the plain response to this exchange — it is not a separate NHP header type.

## NHP-EXT — Exit {#nhp-ext--exit}

**ID:** `16` · **Direction:** Agent → Server · **Body:** empty.

Agent explicitly requests early teardown of an active session. The Server then sends an NHP-AOP with `opnTime = 0` to close the AC rule.

---

## DHP message types {#dhp}

The Data-object Hiding Protocol reuses the NHP wire format for a separate
set of flows around data-object registration, access, attestation, and key
wrapping. They are listed here so the ID table is complete; their payload
fields (`DRGMsg`, `DAKMsg`, `DARMsg`, `DAGMsg`, `DSAMsg`, `DAVMsg`, `DWRMsg`,
`DWAMsg`, `DBOnlineMsg`, `ServerDBAckMsg`, `DHPKnockMsg`) are defined in
[`nhp/common/nhpmsg.go`](https://github.com/OpenNHP/opennhp/blob/main/nhp/common/nhpmsg.go).
A dedicated DHP reference page will expand these; for now, see the
[DHP Quick Start]({{ '/dhp_quick_start/' | relative_url }}).

---

## See also

- [Message Header]({{ '/protocol/header/' | relative_url }}) — the envelope every type shares
- [Cryptography]({{ '/cryptography/' | relative_url }}) — how the payload is encrypted
- [Glossary]({{ '/glossary/' | relative_url }}) — canonical names for every role that appears here
